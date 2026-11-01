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

	// currentRegion 是词法上最内层 region 名（#runtime.chan/coro 的区域推断源）。
	currentRegion string

	outValue   *valueRef
	n          int
	labelN     int
	slotN      int
	terminated bool

	// 协程（designs/0001 §10）：isCoro 时 IR 形态为 presplitcoroutine，
	// 返回句柄控制块指针；coroHdl 是 entry 前奏产出的句柄 SSA。
	isCoro  bool
	coroHdl string
	coroCB  string
	// coroRetLbl 是共享返回尾块标签（coro.end fallthrough 唯一性要求）
	coroRetLbl string
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
	if sig.IsCoroutine {
		g.isCoro = true
		g.coroCB = "%cb"
		l.hasCoroutines = true
	}
	body := g.run()

	ret, _ := l.funcLLType(sig)
	names := make([]string, 0, len(sig.Regions)+len(sig.Params)+2)
	for i := range sig.Regions {
		names = append(names, "ptr %r"+strconv.Itoa(i))
	}
	for i, p := range sig.Params {
		plt := l.llty(p.Type)
		if i == 0 && mutsSelf(sig) {
			plt = "ptr"
		} else if l.paramByRef(sig, i, p.Type) {
			plt = "ptr"
		}
		names = append(names, plt+" %p_"+escName(p.Name))
	}
	if g.isCoro {
		// 两个隐藏尾参：%cb（句柄控制块，调用方分配）与 %frameReg（协程帧的区域）。
		names = append(names, "ptr %cb", "ptr %frameReg")
	}
	// M1：全部函数给外部链接。理由：热替换要求跨模块调用不可内联（designs/0001 铁律 7），
	// 而 M1 尚未产出分派表，先用外部链接保证跨模块调用可解析（M3 换分派表）。
	define := "define " + ret + " @" + symbolFor(l.modName, sig.Name, sig.Owner) +
		"(" + strings.Join(names, ", ") + ")"
	if g.isCoro {
		// 协程函数：IR 返回类型恒为句柄（ptr），多两个隐藏尾参
		// %cb（句柄控制块，调用方分配）与 %frameReg（协程帧的区域）。
		define = "define ptr @" + symbolFor(l.modName, sig.Name, sig.Owner) +
			"(" + strings.Join(names, ", ") + ") #0"
	}
	return define + " {\n" + body + "}\n"
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
		if i == 0 && mutsSelf(sig) {
			// @mut[self]：self 的槽就是传入指针——成员读写直接落到调用方内存。
			s := &slot{lt: g.l.llty(p.Type), ptr: "%p_" + escName(p.Name)}
			if i < len(paramSyms) && paramSyms[i] != nil {
				g.slots[paramSyms[i]] = s
			}
			g.namedSlots[p.Name] = s
			continue
		}
		if g.l.paramByRef(sig, i, p.Type) {
			// 大聚合体按只读指针传入（AX-9）：槽直接指向调用方内存，只读 load。
			s := &slot{lt: g.l.llty(p.Type), ptr: "%p_" + escName(p.Name)}
			if i < len(paramSyms) && paramSyms[i] != nil {
				g.slots[paramSyms[i]] = s
			}
			g.namedSlots[p.Name] = s
			continue
		}
		var s *slot
		if i < len(paramSyms) {
			s = g.slotFor(paramSyms[i])
		}
		if s == nil {
			s = &slot{lt: g.l.llty(p.Type), ptr: g.alloca(p.Name, g.l.llty(p.Type))}
		}
		g.emitRaw("  store " + g.l.llty(p.Type) + " %p_" + escName(p.Name) + ", ptr " + s.ptr)
	}
	if g.isCoro {
		g.emitCoroPrelude()
	}
	g.lowerBlock(g.decl.Body, false)

	lines := make([]string, 0, len(g.entryLines)+len(g.lines)+4)
	lines = append(lines, g.entryLines...)
	lines = append(lines, g.lines...)
	if g.isCoro {
		// 函数体自然落完 = 完成协程（有值协程缺 return 在 sema 已拦）。
		// 共享返回尾块无条件落出（return 路径 br 到它）。
		mark := len(g.lines)
		if !g.terminated {
			g.coroFinish(1, nil, nil)
		}
		g.emitCoroRetTail()
		lines = append(lines, g.lines[mark:]...)
		return strings.Join(lines, "\n") + "\n"
	}
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
//
// 宿主 C main 收 (argc, argv)；程序 import std.os 时先转存给 runtime 的
// aila_rt_args_save（std.os.args_count/args_get 从那里读），否则不转发——
// 参数无人使用。AILA 的 main 签名不动（designs/0003 §4.3）；嵌入式
// startup.S 以 0 参调 main，多余实参被 C ABI 忽略，签名加宽无影响。
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
	var b strings.Builder
	b.WriteString("define i32 @main(i32 %argc, ptr %argv) {\nentry:\n")
	if l.progUsesOS {
		l.ensureDeclare("aila_rt_args_save", "void @aila_rt_args_save(i32, ptr)")
		b.WriteString("  call void @aila_rt_args_save(i32 %argc, ptr %argv)\n")
	}
	b.WriteString("  call void @" + sym + "(" + strings.Join(args, ", ") + ")\n  ret i32 0\n}\n")
	l.funcs = append(l.funcs, b.String())
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
	case *ast.LetGroup:
		g.lowerLetGroup(x)
	case *ast.ChanSend:
		g.lowerChanSend(x)
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
	case *ast.Select:
		g.lowerSelect(x)
	case *ast.Region:
		g.lowerRegion(x)
	case *ast.Unsafe:
		g.lowerBlock(x.Body, true)
	case *ast.Defer:
		if len(g.defers) > 0 {
			g.defers[len(g.defers)-1] = append(g.defers[len(g.defers)-1], x)
		}
	case *ast.Yield:
		g.lowerYield()
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
	var s *slot
	if sym := g.viewSymbol(d); sym != nil {
		s = g.slotFor(sym)
		if s == nil {
			return
		}
	} else {
		t := g.ty(d.Init)
		lt := g.l.llty(t)
		s = &slot{lt: lt, ptr: g.alloca(d.Name, lt)}
		g.namedSlots[d.Name] = s
	}
	// 定长数组 → 视图重标（§3.8）：数组退化为视图，内存不动——
	// 视图 = (首元素地址, 字节数)（[Unsafe] 的 len 按规范是 byte_len）。
	if f, ok := g.ty(d.Init).(types.Fixed); ok {
		_, ptr := g.lowerPlace(d.Init)
		byteN := int64(f.N) * int64(g.l.layout.Size(f.Elem))
		v := g.tmp()
		g.emit(v + " = insertvalue " + g.l.viewLT() + " poison, ptr " + ptr + ", 0")
		v2 := g.tmp()
		g.emit(v2 + " = insertvalue " + g.l.viewLT() + " " + v + ", " + g.l.lenLT() + " " + strconv.FormatInt(byteN, 10) + ", 1")
		g.emit("store " + g.l.viewLT() + " " + v2 + ", ptr " + s.ptr)
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
	prev := g.currentRegion
	g.currentRegion = stmt.Name
	g.lowerBlock(stmt.Body, true)
	g.currentRegion = prev
	if !g.terminated {
		g.l.ensureDeclare("aila_rt_region_leave", "void @aila_rt_region_leave(ptr)")
		g.emit("call void @aila_rt_region_leave(ptr " + reg + ")")
	}
	delete(g.regionVals, stmt.Name)
}

// lowerIf 生成 if / else if / else 链。
// `#compile.if` 已被 sema 定下剪枝（x.Prune >= 0）：只生成命中的那一支，
// 其余分支不产生任何指令——但它们此前已经过完整的解析与类型检查（designs/0006 §1.3）。
func (g *funcGen) lowerIf(x *ast.If) {
	if x.Prune >= 0 {
		switch {
		case x.Prune == 0:
			g.lowerBlock(x.Then, true)
		case x.Prune <= len(x.Elifs):
			if blk, ok := x.Elifs[x.Prune-1].B.(*ast.Block); ok {
				g.lowerBlock(blk, true)
			}
		default:
			if x.Els != nil {
				g.lowerBlock(x.Els, true)
			}
		}
		return
	}
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

// lowerReturn 生成 return（含错误通道与多值组）。协程函数走 lowerCoroReturn。
func (g *funcGen) lowerReturn(x *ast.Return) {
	if g.isCoro {
		g.lowerCoroReturn(x)
		return
	}
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
	// v2：`return .成员;` —— 错误成员构造自动装载错误态（designs/0004 §3.4）
	if g.l.check.ErrReturns[x] {
		v := g.lowerExpr(x.Value)
		rval := g.buildResult(lt, baseT, g.sig.Err, false, v.s)
		g.flushDefers()
		g.emit("ret " + lt + " " + rval)
		g.terminated = true
		return
	}
	v := g.lowerExpr(x.Value)
	s := v.s
	if g.sig.Err != nil {
		s = g.buildResult(lt, baseT, g.sig.Err, true, s)
	} else if opt, ok := g.sig.Ret.(types.Optional); ok {
		// v2：`return 值;` 对 `T?` 自动装载非空态（designs/0004 §5）——
		// sema 已把值按 Inner 检查，这里补包装；niche 表示不变（零包装）。
		if _, isOpt := g.ty(x.Value).(types.Optional); !isOpt {
			if g.l.isNicheOptional(opt) {
				s = v.s
			} else {
				cur := g.tmp()
				g.emit(cur + " = insertvalue " + lt + " undef, i8 1, 0")
				nxt := g.tmp()
				g.emit(nxt + " = insertvalue " + lt + " " + cur + ", " + g.l.llty(opt.Inner) + " " + v.s + ", 1")
				s = nxt
			}
		}
	}
	g.flushDefers()
	g.emit("ret " + lt + " " + s)
	g.terminated = true
}

// --------------------------------------------------------------------------- //
// 协程（designs/0001 §10.2-10.5：无栈协程，llvm.coro intrinsics）
// --------------------------------------------------------------------------- //

// emitCoroPrelude 在 entry 块（参数落槽后）发协程前奏：
// coro.id → coro.size → 帧分配（落 @函数区域）→ coro.begin → 控制块初始化。
// 控制块 {i8 state; ptr frame; T out; E err} 由调用方分配并经 %cb 传入。
func (g *funcGen) emitCoroPrelude() {
	g.l.ensureDeclare("llvm.coro.id", "token @llvm.coro.id(i32, ptr, ptr, ptr)")
	g.l.ensureDeclare("llvm.coro.size.i64", "i64 @llvm.coro.size.i64()")
	g.l.ensureDeclare("llvm.coro.begin", "ptr @llvm.coro.begin(token, ptr)")
	g.l.ensureDeclare("llvm.coro.save", "token @llvm.coro.save(ptr)")
	g.l.ensureDeclare("llvm.coro.suspend", "i8 @llvm.coro.suspend(token, i1)")
	g.l.ensureDeclare("llvm.coro.end", "void @llvm.coro.end(ptr, i1, token)")
	g.l.ensureDeclare("llvm.coro.resume", "void @llvm.coro.resume(ptr)")
	g.l.ensureDeclare("aila_rt_region_alloc", "ptr @aila_rt_region_alloc(ptr, i64, i64)")
	id := g.tmp()
	g.emit(id + " = call token @llvm.coro.id(i32 0, ptr null, ptr null, ptr null)")
	sz := g.tmp()
	g.emit(sz + " = call i64 @llvm.coro.size.i64()")
	fr := g.tmp()
	g.emit(fr + " = call ptr @aila_rt_region_alloc(ptr %frameReg, i64 " + sz + ", i64 16)")
	hdl := g.tmp()
	g.emit(hdl + " = call ptr @llvm.coro.begin(token " + id + ", ptr " + fr + ")")
	g.coroHdl = hdl
	fp := g.coroFieldPtr(8)
	g.emit("store ptr " + hdl + ", ptr " + fp)
	sp := g.coroFieldPtr(0)
	g.emit("store i8 0, ptr " + sp) // 0 = Suspended（启动即运行至首个挂起点）
}

// coroLayout 返回控制块布局 (offOut, offErr, cbsize)。
// state@0 (i8)、frame@8 (ptr)、out@16、err 对齐排在其后；无 Ok 时 offOut=-1，
// 无 Err 时 offErr=-1。整体按 8 字节对齐收尾。
func (g *funcGen) coroLayout() (int, int, int) {
	return g.l.coroLayoutOf(g.sig)
}

// coroLayoutOf 按签名算协程控制块布局（协程体内与启动点共用同一份算法）。
func (l *Lowerer) coroLayoutOf(sig *sema.FuncSig) (int, int, int) {
	var okT types.Type
	if sig.RetGroup != nil {
		okT = types.Group{Items: sig.RetGroup}
	} else if sig.Ret != nil {
		okT = sig.Ret
	}
	if okT != nil && l.llty(okT) == "void" {
		okT = nil
	}
	return l.coroStateLayout(okT, sig.Err)
}

// coroStateLayout 按 Ok/Err 载荷类型算控制块布局；resume 构造判别和时用它
// （Ok/Err 与被调签名 Ret/Err 同源，两处布局必然一致）。
func (l *Lowerer) coroStateLayout(okT, errT types.Type) (int, int, int) {
	offOut, offErr := -1, -1
	size := 16
	if okT != nil {
		offOut = 16
		off := 16 + l.layout.Size(okT)
		if errT != nil {
			al := l.layout.Align(errT)
			if al < 1 {
				al = 1
			}
			offErr = ((off + al - 1) / al) * al
			size = offErr + l.layout.Size(errT)
		} else {
			size = off
		}
	} else if errT != nil {
		offErr = 16
		size = 16 + l.layout.Size(errT)
	}
	if r := size % 8; r != 0 {
		size += 8 - r
	}
	return offOut, offErr, size
}

// coroFieldPtr 生成控制块内 byte 偏移的字段地址。
func (g *funcGen) coroFieldPtr(off int) string {
	p := g.tmp()
	g.emit(p + " = getelementptr inbounds i8, ptr " + g.coroCB + ", i64 " + strconv.Itoa(off))
	return p
}

// coroFinish 完成协程：写 out/err 槽（可选）→ state → defer → coro.end(final) →
// ret 句柄。state：1=Ok（正常 return / 自然落完 / 可空 `?` 无值）、2=Err（`return _ ! E`）。
func (g *funcGen) coroFinish(state int, okVal, errVal *string) {
	offOut, offErr, _ := g.coroLayout()
	if state == 1 && okVal != nil {
		// out 槽放的是基础返回类型（I32 / Group），不是整包 Result
		var okT types.Type
		if g.sig.RetGroup != nil {
			okT = types.Group{Items: g.sig.RetGroup}
		} else if g.sig.Ret != nil {
			okT = g.sig.Ret
		}
		g.emit("store " + g.l.llty(okT) + " " + *okVal + ", ptr " + g.coroFieldPtr(offOut))
	}
	if state == 2 && errVal != nil {
		g.emit("store " + g.l.llty(g.sig.Err) + " " + *errVal + ", ptr " + g.coroFieldPtr(offErr))
	}
	sp := g.coroFieldPtr(0)
	g.emit("store i8 " + strconv.Itoa(state) + ", ptr " + sp)
	g.flushDefers()
	// 收敛到共享返回尾块：coro.end（fallthrough，全函数唯一）+ ret 句柄。
	// 多条 return 路径只能各自 br 到这里，否则 coro-split 报
	// 「Only one coro.end can be marked as fallthrough」。
	g.emit("br label %" + g.coroRetLabel())
	g.terminated = true
}

// coroRetLabel 返回（并惰性登记）本函数唯一的协程返回尾块标签。
func (g *funcGen) coroRetLabel() string {
	if g.coroRetLbl == "" {
		g.coroRetLbl = g.newLabel("coro.ret")
	}
	return g.coroRetLbl
}

// emitCoroRetTail 落出协程返回尾块（coro.end 是普通返回处的唯一 fallthrough 标记；
// resume/destroy 克隆里 split 会改写 coro.end 并把 ret 转成 ret void；
// 完成态通过控制块 state（1/2）+ resume 侧状态机传给调用方）。
func (g *funcGen) emitCoroRetTail() {
	g.emitRaw(g.coroRetLabel() + ":")
	g.terminated = false // emit 在 terminated 后会丢弃指令；尾块本身是活代码
	g.emit("call void @llvm.coro.end(ptr " + g.coroHdl + ", i1 false, token none)")
	g.emit("ret ptr %cb")
	g.terminated = true
}

// lowerCoroReturn 协程函数的 return：完成语义（不走普通结果聚合）。
func (g *funcGen) lowerCoroReturn(x *ast.Return) {
	if x.Err != nil {
		v := g.lowerExpr(x.Err)
		g.coroFinish(2, nil, &v.s)
		return
	}
	// v2：`return .成员;` —— 错误成员构造自动装载错误态（designs/0004 §3.4）
	if g.l.check.ErrReturns[x] {
		v := g.lowerExpr(x.Value)
		g.coroFinish(2, nil, &v.s)
		return
	}
	if x.Value == nil {
		g.coroFinish(1, nil, nil)
		return
	}
	v := g.lowerExpr(x.Value)
	g.coroFinish(1, &v.s, nil)
}

// coroSuspendRetry 落出一次挂起点（yield / chan 收发共用）：
// save → suspend → switch（0=恢复后进 resume 续体，非 0=挂起）。
// 挂起路径与完成路径一样汇入共享返回尾块（coro.end + ret 句柄）。
// 实证（clang 参照）：所有「把控制权还给调用方」的路径必须收敛到
// 唯一的 coro.end 标记块——裸 ret 会让 CoroSplit 把 resume 克隆的
// 内部挂起路径置 unreachable，re-suspension 全部失效。
// resume 续体目标由调用方给出（yield 的下一语句 / 收发重试循环头）。
func (g *funcGen) coroSuspendRetry(resumeLbl string) {
	sv := g.tmp()
	g.emit(sv + " = call token @llvm.coro.save(ptr null)")
	su := g.tmp()
	g.emit(su + " = call i8 @llvm.coro.suspend(token " + sv + ", i1 false)")
	susp := g.newLabel("coro.susp")
	// coro.suspend 返回 0=恢复继续、非 0（1）=挂起。coro-split 后 ramp 以
	// 常量 -1 走 default=挂起路径，resume 克隆以常量 0 走 case-0=续体，
	// 故 case-0 必须是续体（实证见 coro-mini build/split.ll）。
	// switch 多 case 之间用空白分隔（LLVM 文法：case 项不带项间逗号）。
	g.emit("switch i8 " + su + ", label %" + susp + " [ i8 0, label %" + resumeLbl + " ]")
	g.emitRaw(susp + ":")
	g.terminated = false
	g.emit("br label %" + g.coroRetLabel())
	g.terminated = true
}

// lowerYield 生成 `yield;`：挂起，恢复后从下一语句继续。
func (g *funcGen) lowerYield() {
	after := g.newLabel("coro.after")
	g.coroSuspendRetry(after)
	g.emitRaw(after + ":")
	g.terminated = false
}

// threadStartInto 是线程启动的共享主体（`#runtime.spawn`，designs/0004 §2.4）：
// 实参先求值（与普通调用一致的求值顺序），再按声明布局
// 逐参对齐打包 → aila_rt_thread_args 分配私有堆块 → ensureThreadThunk 私有解包
// → aila_rt_thread_spawn 交给 OS。返回码非零 = trap。
func (g *funcGen) threadStartInto(sig *sema.FuncSig, args []ast.Node, stack string, span diagSpan) string {
	vals := make([]valueRef, len(args))
	for i, a := range args {
		vals[i] = g.lowerExpr(a)
	}
	type packed struct {
		off int
		plt string
		val string
	}
	var items []packed
	off := 0
	for i, p := range sig.Params {
		if i >= len(vals) {
			break // 实参数量不符 sema 已报，这里防御
		}
		plt := g.l.paramLLType(sig, p.Type)
		av := vals[i]
		if plt == "ptr" && av.lt != "ptr" {
			av.s = g.viewPtr(args[i], av.s, av.lt)
		} else if plt != av.lt && isIntLL(plt) && (isIntLL(av.lt) || av.lt == "i1") {
			av.s = g.coerceInt(av.lt, plt, av.s)
		}
		align := g.l.layout.Align(p.Type)
		if align < 1 {
			align = 1
		}
		off = (off + align - 1) &^ (align - 1)
		items = append(items, packed{off, plt, av.s})
		off += g.l.layout.Size(p.Type)
	}
	argp := "null"
	if off > 0 {
		g.l.ensureDeclare("aila_rt_thread_args", "ptr @aila_rt_thread_args(i64)")
		argp = g.tmp()
		g.emit(argp + " = call ptr @aila_rt_thread_args(i64 " + strconv.Itoa(off) + ")")
		for _, it := range items {
			gp := argp
			if it.off > 0 {
				gp = g.tmp()
				g.emit(gp + " = getelementptr i8, ptr " + argp + ", i64 " + strconv.Itoa(it.off))
			}
			g.emit("store " + it.plt + " " + it.val + ", ptr " + gp)
		}
	}
	thunk := g.l.ensureThreadThunk(sig)
	g.l.ensureDeclare("aila_rt_thread_spawn", "i32 @aila_rt_thread_spawn(i64, ptr, ptr)")
	rc := g.tmp()
	g.emit(rc + " = call i32 @aila_rt_thread_spawn(i64 " + stack + ", ptr @" + thunk + ", ptr " + argp + ")")
	g.emitCheck("icmp ne i32 "+rc+", 0", "线程创建失败", span)
	return rc
}
