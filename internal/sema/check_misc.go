package sema

import (
	"strings"

	"aila/internal/ast"
	"aila/internal/diag"
	"aila/internal/types"
)

func isVoidOrNever(t types.Type) bool {
	if t == nil {
		return true
	}
	switch t.(type) {
	case types.Void, types.Never:
		return true
	}
	return false
}

func (c *Checker) requireInteger(t types.Type, span diag.Span, what string) {
	if _, ok := t.(types.UntypedInt); ok {
		return
	}
	if _, ok := t.(types.UntypedFloat); ok {
		return
	}
	if !types.IsInteger(t) {
		c.err("E3201", span, what+"必须是整数，实得 `"+typeName(t)+"`", "", "")
	}
}

// --------------------------------------------------------------------------- //
// 绑定 / 视图 / 返回 / 循环 / region
// --------------------------------------------------------------------------- //

func (c *Checker) checkBinding(d *ast.LetDecl) {
	ctx := c.ctx
	var expected types.Type
	if d.Type != nil {
		expected = c.table.Resolve(d.Type, c.bag)
	}
	var t types.Type
	if d.Init != nil {
		t = c.checkExpr(d.Init, expected)
	} else {
		t = expected
		if t == nil {
			t = types.I32
		}
	}
	if expected != nil {
		c.unify(expected, t, d.Span, "绑定 `"+d.Name+"`")
		t = expected
	} else if isVoid(t) {
		c.err("E3207", d.Span, "`"+d.Name+"` 的初始化器没有值", "无值不可绑定", "")
		t = types.I32
	}
	sym := &Symbol{Name: d.Name, Type: t, Mutable: d.KindV == "var", Kind: "local", OriginRegion: c.originOf(d.Init)}
	if !ctx.Scope.Declare(sym) {
		c.err("E3001", d.Span, "`"+d.Name+"` 在同一作用域重复声明", "这里", "")
	}
	c.Bindings[d] = t
	c.Syms[d] = sym
	c.locals = append(c.locals, sym)
	// `let 名 = <const-expr>` 也记成编译期已知（designs/0006 §1.2）：
	// let 不可重新赋值，所以它的值在整段作用域内恒定。
	if d.Init != nil && !sym.Mutable {
		if v, ok := c.constExprInt(d.Init); ok {
			c.ConstVals[sym] = v
		}
	}
}

func (c *Checker) checkView(d *ast.ViewDecl) {
	var target types.Type
	if d.Type != nil {
		target = c.table.Resolve(d.Type, c.bag)
	}
	src := c.checkExpr(d.Init, nil)
	if target != nil && sameType(target, src) {
		c.err("E4004", d.Span, "`view` 的目标类型与源同型", "同型就该用 `let`",
			"写成 `let "+d.Name+": "+typeName(target)+" = 源;`")
	}
	sym := &Symbol{Name: d.Name, Type: target, Kind: "local", OriginRegion: c.originOf(d.Init)}
	c.ctx.Scope.Declare(sym)
	c.locals = append(c.locals, sym)
}

func (c *Checker) checkReturn(r *ast.Return) {
	sig := c.ctx.Sig
	hasValue := sig.RetGroup != nil || !isVoid(sig.Ret)
	// v2：`return .成员;` —— 错误成员构造自动装载错误态（designs/0004 §3.4）
	if ld, ok := r.Value.(*ast.StructLit); ok && ld.LeadingDot {
		if sig.Err == nil {
			c.err("E3201", r.Span, "该函数不返回 `T ! E`，不能返回错误成员", "这里",
				"想要错误通道就把签名写成 `-> T ! 错误类型`")
			c.checkExpr(ld, nil)
			return
		}
		et := c.checkExpr(ld, sig.Err)
		c.unify(sig.Err, et, ast.SpanOf(ld), "错误值")
		c.ErrReturns[r] = true
		return
	}
	if r.Err != nil {
		if sig.Err == nil {
			c.err("E3201", r.Span, "该函数不返回 `T ! E`，不能写 `return _ ! 错误;`", "这里",
				"想要错误通道就把签名写成 `-> T ! 错误类型`")
			return
		}
		et := c.checkExpr(r.Err, sig.Err)
		c.unify(sig.Err, et, ast.SpanOf(r.Err), "错误值")
		return
	}
	if r.Value == nil {
		if hasValue {
			c.err("E3208", r.Span, "有值函数必须 `return 值;`", "这里", "")
		}
		return
	}
	if !hasValue {
		c.err("E3201", r.Span, "无返回值函数不能 `return 值;`", "return 只带控制流", "")
		c.checkExpr(r.Value, nil)
		return
	}
	if sig.RetGroup != nil {
		t := c.checkExpr(r.Value, nil)
		want := types.Group{Items: sig.RetGroup}
		if _, isGroup := r.Value.(*ast.GroupExpr); !isGroup {
			c.err("E3201", ast.SpanOf(r.Value), "多值组返回值必须写 `return (a, b);`", "这里", "")
		}
		c.unify(want, t, ast.SpanOf(r.Value), "返回值")
		return
	}
	// v2：`return 值;` 对 `T?` 签名自动装载非空态（designs/0004 §5）——
	// niche 可空视图（ptr==0 即 None）表示不变、零包装
	if opt, ok := sig.Ret.(types.Optional); ok {
		if p, isNone := r.Value.(*ast.Path); !isNone || p.Dotted() != "None" {
			ot := c.checkExpr(r.Value, opt.Inner)
			c.unify(opt.Inner, ot, ast.SpanOf(r.Value), "返回值")
			c.checkEscape(r.Value, "返回值")
			return
		}
		// `return None;` 走常规路径按整体类型检查
	}
	t := c.checkExpr(r.Value, sig.Ret)
	c.unify(sig.Ret, t, ast.SpanOf(r.Value), "返回值")
	c.checkEscape(r.Value, "返回值")
}

func (c *Checker) checkForIn(f *ast.ForIn) {
	ctx := c.ctx
	it := c.checkExpr(f.Iterable, nil)
	var elem types.Type
	switch x := it.(type) {
	case types.View:
		elem = x.Elem
	case types.Fixed:
		elem = x.Elem
	case types.Prim:
		if x.K == "text" {
			elem = types.C32
		}
	case types.Optional:
		c.err("E3201", ast.SpanOf(f.Iterable), "可空视图不能直接迭代", "先 match 解包", "")
		elem = types.U8
	case types.UnsafeView:
		c.err("E3201", ast.SpanOf(f.Iterable), "`[Unsafe]` 没有元素类型，不能迭代", "先 `copy` 定型", "")
		elem = types.U8
	}
	if elem == nil {
		c.err("E3201", ast.SpanOf(f.Iterable), "该类型不可迭代", "for-in 白名单：`[T]` 与 `Text`", "")
		elem = types.U8
	}
	if f.VarType != nil {
		declared := c.table.Resolve(f.VarType, c.bag)
		if !sameType(declared, elem) {
			c.err("E3201", f.VarType.Span, "元素类型标注不匹配", "这里", "")
		}
		elem = declared
	}
	prev := ctx.Scope
	ctx.Scope = NewScope(prev)
	var isym *Symbol
	if f.Index != "" {
		isym = &Symbol{Name: f.Index, Type: types.USize, Kind: "local"}
		ctx.Scope.Declare(isym)
		c.locals = append(c.locals, isym)
	}
	vsym := &Symbol{Name: f.Var, Type: elem, Kind: "local"}
	ctx.Scope.Declare(vsym)
	c.locals = append(c.locals, vsym)
	c.noteFor(f, it, elem, vsym, isym)
	c.loopDepth++
	c.checkBlock(f.Body, false)
	c.loopDepth--
	ctx.Scope = prev
	c.Types[f] = it
}

func (c *Checker) checkRegion(r *ast.Region) {
	ctx := c.ctx
	if ctx.Regions[r.Name] {
		c.err("E4003", r.Span, "区域名 `"+r.Name+"` 与 `@region` 参数重名", "同一函数内区域名是一个空间", "")
	}
	for _, existing := range ctx.Lexical {
		if existing == r.Name {
			c.err("E4003", r.Span, "内层 `region` 与外层同名 `"+r.Name+"`", "寿命跟区域名走", "给内层区域换个名字")
		}
	}
	if r.Size != nil {
		st := c.checkExpr(r.Size, types.USize)
		c.requireInteger(st, ast.SpanOf(r.Size), "区域配额")
	}
	ctx.Lexical = append(ctx.Lexical, r.Name)
	c.checkBlock(r.Body, true)
	ctx.Lexical = ctx.Lexical[:len(ctx.Lexical)-1]
}

// --------------------------------------------------------------------------- //
// 赋值 / 解构 / 丢弃
// --------------------------------------------------------------------------- //

func (c *Checker) checkAssign(a *ast.Assign) {
	if c.targetTouchesGlobal(a.Target) {
		c.ctx.WroteGlobal = true // E5003 判据：本函数体写过模块级可变状态
	}
	target := c.checkPlace(a.Target)
	if a.Op == "=" {
		src := c.checkExpr(a.Value, target)
		c.unify(target, src, ast.SpanOf(a.Value), "赋值")
		c.checkEscape(a.Value, "存储")
		return
	}
	src := c.checkExpr(a.Value, target)
	if types.IsFloat(target) {
		if a.Op == "%=" || a.Op == "&=" || a.Op == "|=" || a.Op == "^=" || a.Op == "<<=" || a.Op == ">>=" {
			c.err("E3209", a.Span, "浮点不支持 `"+a.Op+"`", "`%=` 只对整数", "")
		}
		c.unify(target, src, ast.SpanOf(a.Value), "复合赋值")
		return
	}
	if !types.IsInteger(target) {
		c.err("E3209", a.Span, "`"+a.Op+"` 只对整数标量或浮点生效", "这里", "")
		return
	}
	c.unify(target, src, ast.SpanOf(a.Value), "复合赋值")
}

// checkPlace 检查左值并返回其类型（写权限在这里校验）。
func (c *Checker) checkPlace(e ast.Node) types.Type {
	t := c.checkExpr(e, nil)
	switch x := e.(type) {
	case *ast.Path:
		sym := c.placeSymbol(x)
		if sym != nil && !sym.Mutable {
			c.err("E4201", x.Span, "`"+sym.Name+"` 是只读绑定（`let`）",
				"需要写权请声明为 `var`；跨函数写要走 `@mut[...]`", "")
		}
	case *ast.Index, *ast.FieldAccess:
		var inner ast.Node = x.(ast.Node)
		if idx, ok := inner.(*ast.Index); ok {
			c.checkPlace(idx.Base)
		} else if fa, ok := inner.(*ast.FieldAccess); ok {
			c.checkPlace(fa.Base)
		}
	default:
		c.err("E2001", ast.SpanOf(e), "该表达式不能作为赋值目标", "赋值目标只能是绑定 / 字段 / 下标", "")
	}
	return t
}

// checkLetGroup 检查 `let (a, b) = expr;` 多值解构绑定（designs/0001 §10.6）。
func (c *Checker) checkLetGroup(d *ast.LetGroup) {
	t := c.checkExpr(d.Init, nil)
	g, ok := t.(types.Group)
	if !ok || len(g.Items) != len(d.Names) {
		c.err("E3201", d.Span, "解构绑定需要与名字数量相等的多值组", "这里", "")
		return
	}
	syms := make([]*Symbol, len(d.Names))
	for i, name := range d.Names {
		if name == "_" {
			continue
		}
		sym := &Symbol{Name: name, Type: g.Items[i], Mutable: false, Kind: "local",
			OriginRegion: c.originOf(d.Init)}
		if !c.ctx.Scope.Declare(sym) {
			c.err("E3001", d.Span, "`"+name+"` 在同一作用域重复声明", "这里", "")
		}
		syms[i] = sym
		c.locals = append(c.locals, sym)
	}
	c.DestructRefs[d] = syms // 复用解构符号通道，codegen 按位取槽
}

// checkChanSend 检查 `ch <- v;` 发送语句（designs/0001 §10.6）：send 端 +
// 值可装载 + 仅协程内合法（满则挂起）。
func (c *Checker) checkChanSend(x *ast.ChanSend) {
	t := c.checkExpr(x.Chan, nil)
	ch, ok := t.(types.Chan)
	if !ok || ch.Dir != "send" {
		c.err("E3201", x.Span, "`<-` 发送需要 `send chan` 端", "这里", "")
		return
	}
	vt := c.checkExpr(x.Value, ch.Elem)
	c.unify(ch.Elem, vt, ast.SpanOf(x), "发送的消息值")
	if !c.ctx.InCoroutine {
		c.err("E5001", x.Span, "`<-` 只在 `coroutine func` 内合法", "发送可能挂起", "")
	}
}

// checkSelect 检查 select 语句（designs/0004 §6）：事件源合法性 + 绑定 +
// 互斥约束。事件白名单 = chan recv 端（`<-ch`）与 coro 句柄（`resume h`）。
//
// `resume h` 是**等完成**事件（§6 v2）：协程完成时才就绪，产出两态——有错误
// 通道是 `T ! E`、只有返回值是 `T`、无返回值则不绑定名字（旧的三态判别和与
// `Suspended` / `Done` 内建模式已随本语义退役）。
//
// select 在非协程（如 main）里也合法：那里「全不就绪」无从挂起，按声明序重试
// 整条链（等价用户手写的轮转驱动循环）。
func (c *Checker) checkSelect(x *ast.Select) {
	srcSeen := map[*Symbol]bool{}
	dupSrc := func(sym *Symbol, span diag.Span) {
		if sym == nil {
			return
		}
		if srcSeen[sym] {
			c.err("E5008", span, "同一事件源 `"+sym.Name+"` 在 select 中重复", "就绪判定先到先得，重复 case 永远选不中", "")
		}
		srcSeen[sym] = true
	}
	for _, sc := range x.Cases {
		var bindT types.Type
		implicit := false
		switch ev := sc.Event.(type) {
		case *ast.Recv:
			t := c.checkExpr(ev.Chan, nil)
			ch, ok := t.(types.Chan)
			if !ok || ch.Dir != "recv" {
				c.err("E3201", ast.SpanOf(ev), "select 的 `<-` 事件需要 `recv chan` 端", "这里", "")
				ch = types.Chan{Elem: types.U8, Dir: "recv"}
			}
			bindT = ch.Elem
			implicit = true // chan 事件省绑定名时隐式绑定 v（designs/0004 §6 用例）
			c.Types[ev] = ch.Elem
			dupSrc(c.placeSymbol(ev.Chan), ast.SpanOf(ev))
		case *ast.Resume:
			t := c.checkExpr(ev.Operand, nil)
			co, ok := t.(types.Coro)
			if !ok {
				c.err("E3201", ast.SpanOf(ev), "select 的 `resume` 事件需要协程句柄", "这里", "")
				co = types.Coro{}
			}
			bindT = coroDoneType(co)
			c.Types[ev] = types.CoroState{Ok: co.Ok, Err: co.Err} // 内部三态表示（codegen 布局用）
			dupSrc(c.placeSymbol(ev.Operand), ast.SpanOf(ev))
		default:
			c.err("E2001", ast.SpanOf(sc.Event), "select 事件只能是 `<-ch` 或 `resume h`", "这里", "")
		}
		prev := c.ctx.Scope
		c.ctx.Scope = NewScope(prev)
		name := sc.Bind
		if name == "" && implicit {
			name = "v"
		}
		if name != "" && name != "_" {
			if bindT == nil {
				c.err("E3201", ast.SpanOf(sc.Event), "无返回值协程的 `resume` 事件没有产出，不能绑定名字", "这里",
					"写成 `case resume h => { ... }`；要看错误用 `-> T ! E` 的协程")
			} else {
				sym := &Symbol{Name: name, Type: bindT, Kind: "local"}
				c.ctx.Scope.Declare(sym)
				c.locals = append(c.locals, sym)
				c.SelectBinds[sc] = sym
			}
		}
		c.checkBlock(sc.Body, false)
		c.ctx.Scope = prev
	}
}

// coroDoneType 是 `case 名 = resume h` 的绑定类型（等完成语义，designs/0004 §6）：
// 有错误通道 → `T ! E`；只有返回值 → `T`；无返回值 → nil（不产出）。
func coroDoneType(co types.Coro) types.Type {
	switch {
	case co.Ok != nil && co.Err != nil:
		return types.Result{Ok: co.Ok, Err: co.Err}
	case co.Ok != nil:
		return co.Ok
	}
	return nil
}

// checkLayerCall 检查层 API 调用（designs/0004 §2.4 v0 清单）。
// v0 层校验只做 E9204（未知 API）；E9201/E9202 的编译期上下文判定随常量
// 求值器扩展落地——属性参数等常量位由「需要常量表达式」天然拦截运行时 API。
// chan/coro 的区域注解由词法区域上下文推断（codegen 取最内层 region）。
func (c *Checker) checkLayerCall(x *ast.LayerCall) types.Type {
	ref := &LayerRef{}
	arity := func(n int, what string) bool {
		if len(x.Args) != n {
			c.err("E9204", x.Span, "`#"+x.Layer+"."+x.Name+"` 需要 "+what, "这里", "")
			return false
		}
		return true
	}
	typeArg := func(i int) types.Type {
		tr, ok := x.Args[i].(*ast.TypeRef)
		if !ok {
			c.err("E9204", ast.SpanOf(x.Args[i]), "`#"+x.Layer+"."+x.Name+"` 的第 "+itoa(i+1)+" 个参数必须是类型", "这里", "")
			return types.U8
		}
		return c.table.Resolve(tr, c.bag)
	}
	switch x.Layer + "." + x.Name {
	case "compile.sizeof":
		if arity(1, "1 个类型参数") {
			ref.Elem = typeArg(0) // codegen 布局需要目标类型
			ref.Kind = "sizeof"
			c.noteLayer(x, ref)
		}
		return types.USize
	case "compile.alignof":
		if arity(1, "1 个类型参数") {
			ref.Elem = typeArg(0) // codegen 布局需要目标类型
			ref.Kind = "alignof"
			c.noteLayer(x, ref)
		}
		return types.USize
	case "compile.as", "runtime.as":
		if arity(2, "类型参数 + 源值") {
			t := typeArg(0)
			st := c.checkExpr(x.Args[1], nil)
			ref.Kind = "as"
			ref.Elem = t
			ref.Src = st
			c.noteLayer(x, ref)
			return t
		}
		return types.U8
	case "comprun.len":
		if !arity(1, "1 个值参数") {
			return types.USize
		}
		vt := c.checkExpr(x.Args[0], nil)
		switch v := vt.(type) {
		case types.Fixed, types.View:
		case types.Prim:
			if v.K != "text" {
				c.err("E3201", ast.SpanOf(x), "`#comprun.len` 不接受该类型（`"+typeName(vt)+"`）", "这里", "")
			}
		default:
			c.err("E3201", ast.SpanOf(x), "`#comprun.len` 不接受该类型（`"+typeName(vt)+"`）", "这里", "")
		}
		ref.Kind = "len"
		ref.At = vt
		c.noteLayer(x, ref)
		return types.USize
	case "runtime.trap":
		if arity(1, "1 个消息参数") {
			mt := c.checkExpr(x.Args[0], types.Text)
			c.unify(types.Text, mt, ast.SpanOf(x.Args[0]), "trap 消息")
			ref.Kind = "trap"
			c.noteLayer(x, ref)
		}
		return types.Never{}
	case "runtime.checked_add", "runtime.checked_sub", "runtime.checked_mul":
		if !arity(2, "两个同型整数") {
			return types.Group{Items: []types.Type{types.I32, types.Bool}}
		}
		at := c.checkExpr(x.Args[0], nil)
		c.requireInteger(at, ast.SpanOf(x.Args[0]), "checked 运算操作数")
		if _, isUT := at.(types.UntypedInt); isUT {
			at = types.I32 // 纯字面量操作数落型，codegen 需要具体宽度
		}
		bt := c.checkExpr(x.Args[1], at)
		c.unify(at, bt, ast.SpanOf(x.Args[1]), "checked 运算操作数")
		ref.Kind = "checked"
		ref.At = at
		ref.CheckedOp = x.Name[len("checked_"):] // "add" / "sub" / "mul"
		c.noteLayer(x, ref)
		return types.Group{Items: []types.Type{at, types.Bool}}
	case "runtime.atomic_load8", "runtime.atomic_load16", "runtime.atomic_load32", "runtime.atomic_load64",
		"runtime.atomic_store8", "runtime.atomic_store16", "runtime.atomic_store32", "runtime.atomic_store64",
		"runtime.atomic_fetch_add8", "runtime.atomic_fetch_add16", "runtime.atomic_fetch_add32", "runtime.atomic_fetch_add64",
		"runtime.atomic_fetch_sub8", "runtime.atomic_fetch_sub16", "runtime.atomic_fetch_sub32", "runtime.atomic_fetch_sub64",
		"runtime.atomic_cas8", "runtime.atomic_cas16", "runtime.atomic_cas32", "runtime.atomic_cas64":
		rest := x.Name[len("atomic_"):] // "load8" / "store16" / "fetch_add32" / "cas64"…
		at := types.U8
		switch rest[len(rest)-2:] {
		case "16":
			at = types.U16
		case "32":
			at = types.U32
		case "64":
			at = types.U64
		}
		n, what := 2, "load 形态实参"
		switch {
		case strings.HasPrefix(rest, "store"):
			n, what = 3, "store 形态实参"
		case strings.HasPrefix(rest, "fetch_add"), strings.HasPrefix(rest, "fetch_sub"):
			n, what = 3, "fetch 形态实参"
		case strings.HasPrefix(rest, "cas"):
			n, what = 4, "cas 形态实参"
		}
		if !arity(n, what) {
			return types.U8
		}
		bt := c.checkExpr(x.Args[0], nil)
		if _, ok := bt.(types.UnsafeView); !ok {
			c.err("E3201", ast.SpanOf(x.Args[0]), "原子操作只作用于 `[Unsafe]` 裸块",
				"原子读写是与 C 侧共享内存的唯一通道（designs/0001 §10.10）", "传 `[Unsafe]` 位")
		}
		ot := c.checkExpr(x.Args[1], types.USize)
		c.requireInteger(ot, ast.SpanOf(x.Args[1]), "原子偏移")
		for i := 2; i < n; i++ {
			vt := c.checkExpr(x.Args[i], at)
			c.unify(at, vt, ast.SpanOf(x.Args[i]), "原子操作数")
		}
		ref.Kind = "atomic"
		ref.At = at
		ref.AtomicFn = "aila_rt_atomic_" + rest
		c.noteLayer(x, ref)
		if strings.HasPrefix(rest, "store") {
			return types.Void{}
		}
		return at
	case "runtime.chan":
		if !arity(2, "元素类型 + 容量") {
			return types.Group{Items: []types.Type{
				types.Chan{Elem: types.U8, Dir: "send"}, types.Chan{Elem: types.U8, Dir: "recv"}}}
		}
		elem := typeArg(0)
		if !isValuePayload(elem) {
			c.err("E5007", ast.SpanOf(x), "通道消息类型必须是值语义类型（类型 `"+typeName(elem)+"`）",
				"消息所有权随值转移，SPSC 下数据竞争免疫", "改传纯数据 struct / enum，或把数据放进通道外的区域")
		}
		capT := c.checkExpr(x.Args[1], types.USize)
		c.requireInteger(capT, ast.SpanOf(x.Args[1]), "通道容量")
		ref.Kind = "chan"
		ref.Elem = elem
		c.noteLayer(x, ref)
		return types.Group{Items: []types.Type{
			types.Chan{Elem: elem, Dir: "send"}, types.Chan{Elem: elem, Dir: "recv"}}}
	case "runtime.coro":
		return c.checkLayerStart(x, ref, true)
	case "runtime.spawn":
		return c.checkLayerStart(x, ref, false)
	default:
		c.err("E9204", x.Span, "未知层 API `#"+x.Layer+"."+x.Name+"`",
			"层内不存在该能力（拼写错误或 v0 未收录）", "")
		return types.I32
	}
}

// checkLayerStart 检查 `#runtime.coro(f, args...)` / `#runtime.spawn(f, args...)`：
// f 必须是函数名 Path；其余实参按目标签名逐参检查。
func (c *Checker) checkLayerStart(x *ast.LayerCall, ref *LayerRef, wantCoro bool) types.Type {
	if len(x.Args) < 1 {
		c.err("E9204", x.Span, "`#"+x.Layer+"."+x.Name+"` 需要目标函数", "这里", "")
		return types.I32
	}
	fp, ok := x.Args[0].(*ast.Path)
	if !ok {
		c.err("E9204", ast.SpanOf(x.Args[0]), "`#"+x.Layer+"."+x.Name+"` 的首个参数必须是函数名", "这里", "")
		return types.I32
	}
	c.checkExpr(fp, nil)
	sig := c.PathFuncs[fp]
	if sig == nil {
		c.err("E9204", ast.SpanOf(fp), "`"+fp.Dotted()+"` 不是函数", "这里", "")
		return types.I32
	}
	if wantCoro != sig.IsCoroutine {
		what := "非协程函数"
		if wantCoro {
			what = "协程函数"
		}
		c.err("E5006", ast.SpanOf(fp),
			"`#"+x.Layer+"."+x.Name+"` 启动的必须是 "+what, "这里", "")
		return types.I32
	}
	if sig.Variadic || sig.VariadicElem != nil {
		c.err("E5006", ast.SpanOf(fp), "变参函数不能作为启动目标", "这里", "先包一层固定参数函数")
		return types.I32
	}
	if wantCoro {
		// 层 API 无区域实参通道：实参直传被调签名，区域参数会错位
		if len(sig.Regions) > 0 {
			c.err("E5006", ast.SpanOf(fp), "启动目标不能带区域参数（`#runtime.coro` 无 `@区域` 通道）",
				"这里", "包一层无区域参数的 `coroutine func`")
			return types.I32
		}
	} else if !sig.IsThread {
		c.err("E5006", ast.SpanOf(fp), "`#runtime.spawn` 启动的必须是 `thread func`",
			"线程入口白名单（值语义 / chan 端 / [Unsafe]）在声明处检查", "把目标函数标注为 `thread func`")
		return types.I32
	}
	got := len(x.Args) - 1
	if got != len(sig.Params) {
		c.err("E2002", ast.SpanOf(x), "实参数量不符：需要 "+itoa(len(sig.Params))+"，实得 "+itoa(got), "这里", "")
	}
	for i := 1; i < len(x.Args) && i-1 < len(sig.Params); i++ {
		pt := sig.Params[i-1].Type
		at := c.checkExpr(x.Args[i], pt)
		c.unify(pt, at, ast.SpanOf(x.Args[i]), "启动实参")
	}
	ref.Kind = map[bool]string{true: "coro", false: "spawn"}[wantCoro]
	ref.Sig = sig
	c.noteLayer(x, ref)
	if wantCoro {
		// 句柄产出类型对齐 CoroCreate：`-> T ! E` → `coro T ! E`；`-> T` → `coro T`；无返回 → 裸 `coro`
		var base types.Type = types.Void{}
		if sig.RetGroup != nil {
			base = types.Group{Items: sig.RetGroup}
		} else if sig.Ret != nil {
			base = sig.Ret
		}
		co := types.Coro{}
		if _, isVoid := base.(types.Void); !isVoid {
			co.Ok = base
		}
		co.Err = sig.Err
		return co
	}
	return types.Void{}
}

// isValuePayload 判定消息 / 跨执行流白名单的「值语义类型」（designs/0001
// §10.6/§10.7）：标量（除 Text）/ `[T; N]` 内联数组 / 纯数据 struct /
// 载荷全为值语义的 enum。
func isValuePayload(t types.Type) bool {
	switch x := t.(type) {
	case types.Prim:
		return x.K != "text"
	case types.Fixed:
		return isValuePayload(x.Elem)
	case types.Named:
		switch info := x.Info.(type) {
		case *types.StructInfo:
			for _, f := range info.Fields {
				if !isValuePayload(f.Type) {
					return false
				}
			}
			return true
		case *types.EnumInfo:
			for _, v := range info.Variants {
				for _, f := range v.Payload {
					if !isValuePayload(f.Type) {
						return false
					}
				}
			}
			return true
		}
	}
	return false
}

// threadParamOK 判定线程入口参数白名单（designs/0001 §10.7）：
// 值语义类型 / chan 任一端 / [Unsafe] 豁免位（含 `?` 包裹形态）。
func threadParamOK(t types.Type) bool {
	switch x := t.(type) {
	case types.Chan, types.UnsafeView:
		return true
	case types.Optional:
		return threadParamOK(x.Inner)
	}
	return isValuePayload(t)
}

// recordCallEdge 记录「当前函数 → 被调函数」边（E5003 线程调用链分析）。
func (c *Checker) recordCallEdge(sig *FuncSig) {
	if c.ctx == nil || c.ctx.Decl == nil || sig == nil {
		return
	}
	c.CallEdges[c.ctx.Decl] = append(c.CallEdges[c.ctx.Decl], sig)
}

// targetTouchesGlobal 判定赋值目标是否落到模块级可变状态（E5003 判据）：
// 沿字段 / 下标链下钻到底部绑定，本模块全局 var / state 或跨模块 `dep.全局`。
func (c *Checker) targetTouchesGlobal(e ast.Node) bool {
	for {
		switch x := e.(type) {
		case *ast.Path:
			return c.pathIsGlobal(x)
		case *ast.FieldAccess:
			e = x.Base
		case *ast.Index:
			e = x.Base
		default:
			return false
		}
	}
}

// pathIsGlobal 判定路径是否指向模块级可变状态。
func (c *Checker) pathIsGlobal(p *ast.Path) bool {
	if len(p.Names) == 0 {
		return false
	}
	head := p.Names[0]
	if c.ctx != nil && c.ctx.Scope != nil {
		if sym := c.ctx.Scope.Lookup(head); sym != nil {
			// 局部 / 参数遮蔽：写到的是局部，不算全局（符号本身的 Kind 已区分）
			return sym.Kind == "global" || sym.Kind == "state"
		}
	}
	if g := c.PathGlobals[p]; g != nil {
		return g.Mutable
	}
	if len(p.Names) >= 2 {
		if g, ok := c.sigs.Globals[head]; ok && g.Mutable {
			return true // 本模块全局的字段链
		}
		if _, ok := c.imports[head]; ok {
			return true // dep.全局 跨模块写（导入符号必为全局）
		}
		return false
	}
	if g, ok := c.sigs.Globals[head]; ok && g.Mutable {
		return true
	}
	return false
}

func (c *Checker) checkDestructAssign(d *ast.DestructAssign) {
	t := c.checkExpr(d.Value, nil)
	g, ok := t.(types.Group)
	if !ok || len(g.Items) != len(d.Names) {
		c.err("E3201", d.Span, "解构赋值需要与左值数量相等的多值组", "这里", "")
		return
	}
	syms := make([]*Symbol, len(d.Names))
	for i, name := range d.Names {
		if name == "_" {
			continue
		}
		sym := c.ctx.Scope.Lookup(name)
		syms[i] = sym
		if sym == nil {
			c.err("E3001", d.Span, "解构目标 `"+name+"` 未声明", "赋值解构只能写给已存在的绑定", "")
			continue
		}
		if !sym.Mutable {
			c.err("E4201", d.Span, "`"+name+"` 是只读绑定", "需要写权请声明为 `var`", "")
		}
		c.unify(sym.Type, g.Items[i], d.Span, "解构 `"+name+"`")
	}
	c.noteDestruct(d, syms)
}

// --------------------------------------------------------------------------- //
// 逃逸与占位判定
// --------------------------------------------------------------------------- //

func (c *Checker) originOf(e ast.Node) string {
	if e == nil {
		return ""
	}
	if r, ok := c.Origins[e]; ok {
		return r
	}
	return ""
}

func (c *Checker) checkEscape(e ast.Node, where string) {
	r := c.originOf(e)
	if r == "" || r == "static" {
		return
	}
	for _, lex := range c.ctx.Lexical {
		if lex == r {
			c.err("E4101", ast.SpanOf(e), where+"指向函数体内 `region "+r+"` 的分配，块退出即回收",
				"逃逸检查：返回 / 存储 / 跨并发捕获", "把分配挪到 `@region` 区域参数或 `@static`")
			return
		}
	}
}

func (c *Checker) placeSymbol(e ast.Node) *Symbol {
	switch x := e.(type) {
	case *ast.Path:
		if sym, ok := c.PathSyms[x]; ok {
			return sym
		}
	case *ast.Index:
		return c.placeSymbol(x.Base)
	case *ast.FieldAccess:
		return c.placeSymbol(x.Base)
	}
	return nil
}

func isVoid(t types.Type) bool {
	if t == nil {
		return false
	}
	_, ok := t.(types.Void)
	return ok
}

var _ = diag.Empty
var _ = ast.Sexp
