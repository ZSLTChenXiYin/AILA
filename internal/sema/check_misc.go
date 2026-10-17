package sema

import (
	"github.com/aila-lang/aila/internal/ast"
	"github.com/aila-lang/aila/internal/diag"
	"github.com/aila-lang/aila/internal/types"
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
