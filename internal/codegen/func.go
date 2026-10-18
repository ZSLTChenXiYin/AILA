package codegen

import (
	"strconv"
	"strings"

	"aila/internal/ast"
	"aila/internal/sema"
	"aila/internal/types"
)

// slot 是一个栈槽：LLVM 类型 + 槽名（`%slot_名字_N`）。
type slot struct {
	lt  string
	ptr string
}

// valueRef 是一个已求值的表达式（IR 类型 + SSA 值）。
type valueRef struct {
	lt string
	s  string
}

// funcGen 是单函数的 IR 生成状态机。
type funcGen struct {
	l    *Lowerer
	sig  *sema.FuncSig
	decl *ast.FuncDecl

	slots      map[*sema.Symbol]*slot // 按符号身份取槽（同名遮蔽互不干扰）
	namedSlots map[string]*slot       // 名字回退表
	regionVals map[string]string      // 区域名 -> SSA（ptr）

	lines      []string
	entryLines []string
	loopExit   []string
	loopCont   []string
	defers     [][]ast.Node

	outValue   *valueRef
	n          int
	labelN     int
	slotN      int
	terminated bool
}

// --------------------------------------------------------------------------- //
// 基础
// --------------------------------------------------------------------------- //

func (g *funcGen) tmp() string {
	g.n++
	return "%t" + strconv.Itoa(g.n)
}

func (g *funcGen) newLabel(hint string) string {
	g.labelN++
	return hint + strconv.Itoa(g.labelN)
}

func (g *funcGen) emit(text string) {
	if g.terminated {
		return
	}
	g.lines = append(g.lines, "  "+text)
}

func (g *funcGen) emitRaw(text string) { g.lines = append(g.lines, text) }

func (g *funcGen) alloca(name, lt string) string {
	g.slotN++
	s := "%slot_" + escName(name) + "_" + strconv.Itoa(g.slotN)
	g.entryLines = append(g.entryLines, "  "+s+" = alloca "+lt+", align 8")
	return s
}

// slotFor 按符号身份取槽；首次访问时分配。
func (g *funcGen) slotFor(sym *sema.Symbol) *slot {
	if sym == nil {
		return nil
	}
	if s, ok := g.slots[sym]; ok {
		return s
	}
	lt := g.l.llty(sym.Type)
	if lt == "void" {
		lt = "i64"
	}
	s := &slot{lt: lt, ptr: g.alloca(sym.Name, lt)}
	g.slots[sym] = s
	if _, ok := g.namedSlots[sym.Name]; !ok {
		g.namedSlots[sym.Name] = s
	}
	return s
}

// ty 取表达式的静态类型（缺失按 I32 处理，与参考实现一致）。
func (g *funcGen) ty(e ast.Node) types.Type {
	if e == nil {
		return types.I32
	}
	if t, ok := g.l.check.Types[e]; ok && t != nil {
		return t
	}
	return types.I32
}

// zero 生成该类型的零值。
func (g *funcGen) zero(t types.Type) string { return g.l.zero(t) }

// --------------------------------------------------------------------------- //
// 函数
// --------------------------------------------------------------------------- //

// sigOfDecl 返回函数声明的签名。
func (l *Lowerer) sigOfDecl(fn *ast.FuncDecl) *sema.FuncSig {
	if fn.OwnerType != "" {
		if sig, ok := l.sigs.Methods[fn.OwnerType+"."+fn.Name]; ok {
			return sig
		}
	}
	if sig, ok := l.sigs.Funcs[fn.Name]; ok {
		return sig
	}
	return nil
}

// genFunction 生成一个函数的 IR（含 `define`/`}` 与结尾换行）。
func (l *Lowerer) genFunction(fn *ast.FuncDecl) string {
	sig := l.sigOfDecl(fn)
	if sig == nil {
		return ""
	}
	g := &funcGen{
		l: l, sig: sig, decl: fn,
		slots:      map[*sema.Symbol]*slot{},
		namedSlots: map[string]*slot{},
		regionVals: map[string]string{},
	}
	body := g.run()

	ret, _ := l.funcLLType(sig)
	names := make([]string, 0, len(sig.Regions)+len(sig.Params))
	for i := range sig.Regions {
		names = append(names, "ptr %r"+strconv.Itoa(i))
	}
	for _, p := range sig.Params {
		names = append(names, l.llty(p.Type)+" %p_"+escName(p.Name))
	}
	// M1：全部函数给外部链接。理由：热替换要求跨模块调用不可内联（designs/0001 铁律 7），
	// 而 M1 尚未产出分派表，先用外部链接保证跨模块调用可解析（M3 换分派表）。
	return "define " + ret + " @" + symbolFor(l.modName, sig.Name, sig.Owner) +
		"(" + strings.Join(names, ", ") + ") {\n" + body + "}\n"
}

// run 生成函数体：入口槽位、参数落槽、语句、收尾 ret。
func (g *funcGen) run() string {
	sig := g.sig
	for i, rn := range sig.Regions {
		g.regionVals[rn] = "%r" + strconv.Itoa(i)
	}
	for _, sym := range g.l.check.Locals[g.decl] {
		g.slotFor(sym)
	}
	paramSyms := g.l.check.ParamSyms[g.decl]
	for i, p := range sig.Params {
		var s *slot
		if i < len(paramSyms) {
			s = g.slotFor(paramSyms[i])
		}
		if s == nil {
			s = &slot{lt: g.l.llty(p.Type), ptr: g.alloca(p.Name, g.l.llty(p.Type))}
		}
		g.emitRaw("  store " + g.l.llty(p.Type) + " %p_" + escName(p.Name) + ", ptr " + s.ptr)
	}
	g.lowerBlock(g.decl.Body, false)

	lines := make([]string, 0, len(g.entryLines)+len(g.lines)+1)
	lines = append(lines, g.entryLines...)
	lines = append(lines, g.lines...)
	if !g.terminated {
		retT := g.l.sigRetType(sig)
		if g.l.llty(retT) == "void" {
			lines = append(lines, "  ret void")
		} else {
			lines = append(lines, "  ret "+g.l.llty(retT)+" "+g.zero(retT))
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

// genEntryWrapper 生成 C 入口 @main：调用模块入口函数后返回 0。
func (l *Lowerer) genEntryWrapper() {
	if l.modName != l.entry {
		return
	}
	sig, ok := l.sigs.Funcs["main"]
	if !ok {
		return
	}
	sym := symbolFor(l.modName, "main", "")
	args := make([]string, 0, len(sig.Params))
	for _, p := range sig.Params {
		args = append(args, l.llty(p.Type)+" "+l.zero(p.Type))
	}
	l.funcs = append(l.funcs, "define i32 @main() {\nentry:\n  call void @"+sym+
		"("+strings.Join(args, ", ")+")\n  ret i32 0\n}\n")
}

// --------------------------------------------------------------------------- //
// 语句
// --------------------------------------------------------------------------- //

func (g *funcGen) lowerBlock(b *ast.Block, newScope bool) {
	if b == nil {
		return
	}
	g.defers = append(g.defers, nil)
	for _, st := range b.Stmts {
		if g.terminated {
			break
		}
		g.lowerStmt(st)
	}
	if !g.terminated {
		group := g.defers[len(g.defers)-1]
		for i := len(group) - 1; i >= 0; i-- {
			g.runDefer(group[i])
		}
	}
	g.defers = g.defers[:len(g.defers)-1]
}

func (g *funcGen) runDefer(d ast.Node) {
	saved := g.terminated
	g.terminated = false
	if df, ok := d.(*ast.Defer); ok {
		if df.Expr != nil {
			g.lowerExpr(df.Expr)
		} else if df.Body != nil {
			for _, s := range df.Body.Stmts {
				g.lowerStmt(s)
			}
		}
	}
	g.terminated = saved
}

func (g *funcGen) flushDefers() {
	for i := len(g.defers) - 1; i >= 0; i-- {
		group := g.defers[i]
		for j := len(group) - 1; j >= 0; j-- {
			g.runDefer(group[j])
		}
	}
}

func (g *funcGen) lowerStmt(st ast.Node) {
	switch x := st.(type) {
	case *ast.Block:
		g.lowerBlock(x, true)
	case *ast.LetDecl:
		g.lowerBinding(x)
	case *ast.ViewDecl:
		g.lowerView(x)
	case *ast.Return:
		g.lowerReturn(x)
	case *ast.If:
		g.lowerIf(x)
	case *ast.ForC:
		g.lowerForC(x)
	case *ast.ForIn:
		g.lowerForIn(x)
	case *ast.Match:
		g.lowerMatch(x, "")
	case *ast.Region:
		g.lowerRegion(x)
	case *ast.Unsafe:
		g.lowerBlock(x.Body, true)
	case *ast.Defer:
		if len(g.defers) > 0 {
			g.defers[len(g.defers)-1] = append(g.defers[len(g.defers)-1], x)
		}
	case *ast.Trap:
		if s, ok := x.Msg.(*ast.StrLit); ok {
			g.trapNow(s.Value, x.Span)
		} else {
			g.trapNow("trap", x.Span)
		}
	case *ast.Yield:
		g.unsupported("协程挂起的代码生成待批次（M4）")
		g.terminated = true
	case *ast.ThreadStart:
		g.unsupported("线程创建的代码生成待批次（M4）")
		g.terminated = true
	case *ast.Break:
		if n := len(g.loopExit); n > 0 {
			g.emit("br label %" + g.loopExit[n-1])
		}
		g.terminated = true
	case *ast.Continue:
		if n := len(g.loopCont); n > 0 {
			g.emit("br label %" + g.loopCont[n-1])
		}
		g.terminated = true
	case *ast.Assign:
		g.lowerAssign(x)
	case *ast.Discard:
		g.lowerExpr(x.Value)
	case *ast.DestructAssign:
		g.lowerDestructAssign(x)
	case *ast.Out:
		v := g.lowerExpr(x.Value)
		g.outValue = &v
	case *ast.ExprStmt:
		g.lowerExpr(x.Expr)
	default:
		g.unsupported("语句 " + kindOf(st) + " 的代码生成待批次")
	}
}

// unsupported 记录「本批次未实现」的构造（Go 侧不静默丢代码）。
func (g *funcGen) unsupported(what string) {
	g.l.unsupported = append(g.l.unsupported, what)
}

func kindOf(n ast.Node) string {
	if n == nil {
		return "nil"
	}
	return n.Kind()
}

// lowerBinding 生成 `let/var` 初始化。
func (g *funcGen) lowerBinding(d *ast.LetDecl) {
	sym := g.l.check.Syms[d]
	if sym == nil {
		return
	}
	lt := g.l.llty(sym.Type)
	if lt == "void" {
		return
	}
	s := g.slotFor(sym)
	if s == nil {
		s = &slot{lt: lt, ptr: g.alloca(d.Name, lt)}
		g.namedSlots[d.Name] = s
	}
	v := g.lowerExpr(d.Init)
	g.emit("store " + lt + " " + v.s + ", ptr " + s.ptr)
}

// lowerView 生成 `view` 声明。
func (g *funcGen) lowerView(d *ast.ViewDecl) {
	sym := g.viewSymbol(d)
	if sym == nil {
		t := g.ty(d.Init)
		lt := g.l.llty(t)
		s := &slot{lt: lt, ptr: g.alloca(d.Name, lt)}
		g.namedSlots[d.Name] = s
		v := g.lowerExpr(d.Init)
		g.emit("store " + lt + " " + v.s + ", ptr " + s.ptr)
		return
	}
	s := g.slotFor(sym)
	if s == nil {
		return
	}
	v := g.lowerExpr(d.Init)
	g.emit("store " + s.lt + " " + v.s + ", ptr " + s.ptr)
}

// viewSymbol 回查 `view` 绑定对应的符号（sema 已把符号放进函数局部表）。
func (g *funcGen) viewSymbol(d *ast.ViewDecl) *sema.Symbol {
	for _, sym := range g.l.check.Locals[g.decl] {
		if sym.Name == d.Name && sym.Kind == "local" {
			if _, ok := g.slots[sym]; ok {
				return sym
			}
		}
	}
	return nil
}

// lowerRegion 生成 `region` 块：进入 / 分配 / 离开。
func (g *funcGen) lowerRegion(stmt *ast.Region) {
	capv := "0"
	if stmt.Size != nil {
		v := g.lowerExpr(stmt.Size)
		capv = v.s
	}
	reg := g.tmp()
	g.l.ensureDeclare("aila_rt_region_enter", "ptr @aila_rt_region_enter(i64)")
	g.emit(reg + " = call ptr @aila_rt_region_enter(i64 " + capv + ")")
	g.regionVals[stmt.Name] = reg
	g.lowerBlock(stmt.Body, true)
	if !g.terminated {
		g.l.ensureDeclare("aila_rt_region_leave", "void @aila_rt_region_leave(ptr)")
		g.emit("call void @aila_rt_region_leave(ptr " + reg + ")")
	}
	delete(g.regionVals, stmt.Name)
}

// lowerIf 生成 if / else if / else 链。
func (g *funcGen) lowerIf(x *ast.If) {
	end := g.newLabel("if.end")
	type branch struct {
		cond ast.Node
		body *ast.Block
	}
	branches := []branch{{x.Cond, x.Then}}
	for _, pair := range x.Elifs {
		blk, _ := pair.B.(*ast.Block)
		cond, _ := pair.A.(ast.Node)
		branches = append(branches, branch{cond, blk})
	}
	for _, br := range branches {
		thenLbl := g.newLabel("if.then")
		nxt := g.newLabel("if.next")
		v := g.lowerExpr(br.cond)
		cb := g.toBool(br.cond, v.s)
		g.emit("br i1 " + cb + ", label %" + thenLbl + ", label %" + nxt)
		g.emitRaw(thenLbl + ":")
		g.terminated = false
		g.lowerBlock(br.body, true)
		if !g.terminated {
			g.emit("br label %" + end)
		}
		g.emitRaw(nxt + ":")
		g.terminated = false
	}
	if x.Els != nil {
		g.lowerBlock(x.Els, true)
		if !g.terminated {
			g.emit("br label %" + end)
		}
	} else {
		g.emit("br label %" + end)
	}
	g.emitRaw(end + ":")
	g.terminated = false
}

// toBool 把值规约成 i1。
func (g *funcGen) toBool(e ast.Node, val string) string {
	lt := g.l.llty(g.ty(e))
	if lt == "i1" {
		return val
	}
	r := g.tmp()
	if lt == "float" || lt == "double" {
		g.emit(r + " = fcmp une " + lt + " " + val + ", 0.0")
	} else {
		g.emit(r + " = icmp ne " + lt + " " + val + ", 0")
	}
	return r
}

// lowerReturn 生成 return（含错误通道与多值组）。
func (g *funcGen) lowerReturn(x *ast.Return) {
	retT := g.l.sigRetType(g.sig)
	lt := g.l.llty(retT)
	var baseT types.Type
	if g.sig.RetGroup != nil {
		baseT = types.Group{Items: g.sig.RetGroup}
	} else {
		baseT = g.sig.Ret
	}
	if x.Err != nil {
		v := g.lowerExpr(x.Err)
		rval := g.buildResult(lt, baseT, g.sig.Err, false, v.s)
		g.flushDefers()
		g.emit("ret " + lt + " " + rval)
		g.terminated = true
		return
	}
	if x.Value == nil {
		g.flushDefers()
		if lt == "void" {
			g.emit("ret void")
		} else {
			g.emit("ret " + lt + " " + g.zero(retT))
		}
		g.terminated = true
		return
	}
	v := g.lowerExpr(x.Value)
	s := v.s
	if g.sig.Err != nil {
		s = g.buildResult(lt, baseT, g.sig.Err, true, s)
	}
	g.flushDefers()
	g.emit("ret " + lt + " " + s)
	g.terminated = true
}
