package sema

import (
	"aila/internal/ast"
	"aila/internal/types"
)

type pathKind int

const (
	pkUnknown pathKind = iota
	pkLocal
	pkGlobal
	pkConst
	pkFunc
	pkField
	pkVariant
	pkBuiltin
)

// checkExpr 推断表达式类型；expected 非空时按目标定型（字面量多态）。
func (c *Checker) checkExpr(e ast.Node, expected types.Type) types.Type {
	if e == nil {
		return types.Void{}
	}
	t := c.checkExprInner(e, expected)
	c.Types[e] = t
	return t
}

func (c *Checker) checkExprInner(e ast.Node, expected types.Type) types.Type {
	switch x := e.(type) {
	case *ast.IntLit:
		// ≥ 2^63 的字面量由 lexer 以位型记录（Raw < 0），随类型一路携带
		// BitPattern 标志，落点检查在 unify。
		ui := types.UntypedInt{Value: x.Raw, BitPattern: x.Raw < 0}
		if expected != nil && (types.IsInteger(expected) || isC32(expected)) {
			c.unify(expected, ui, x.Span, "字面量定型")
			return expected
		}
		return ui
	case *ast.FloatLit:
		if expected != nil && types.IsFloat(expected) {
			return expected
		}
		return types.UntypedFloat{Value: x.Value}
	case *ast.StrLit:
		if expected != nil {
			c.unify(expected, types.Text, x.Span, "字符串字面量")
		}
		return types.Text
	case *ast.CharLit:
		if expected != nil {
			c.unify(expected, types.C32, x.Span, "字符字面量")
			if types.IsInteger(expected) {
				return expected
			}
		}
		return types.C32
	case *ast.BoolLit:
		if expected != nil {
			c.unify(expected, types.Bool, x.Span, "布尔字面量")
		}
		return types.Bool
	case *ast.Placeholder:
		c.err("E2001", x.Span, "`_` 只能出现在解构 / 丢弃 / `return _ ! e` 位置", "这里", "")
		return types.I32
	case *ast.Path:
		return c.checkPath(x, expected)
	case *ast.GroupExpr:
		items := make([]types.Type, 0, len(x.Elems))
		for _, it := range x.Elems {
			items = append(items, c.checkExpr(it, nil))
		}
		return types.Group{Items: items}
	case *ast.ArrayLit:
		return c.checkArrayLit(x, expected)
	case *ast.StructLit:
		return c.checkStructLit(x, expected)
	case *ast.Call:
		return c.checkCall(x, expected)
	case *ast.MethodCall:
		return c.checkMethodCall(x, expected)
	case *ast.Index:
		return c.checkIndex(x)
	case *ast.FieldAccess:
		return c.checkField(x)
	case *ast.Binary:
		return c.checkBinary(x, expected)
	case *ast.Unary:
		return c.checkUnary(x, expected)
	case *ast.TryPropagate:
		return c.checkTryPropagate(x)
	case *ast.Move:
		t := c.checkExpr(x.Operand, nil)
		if p, ok := x.Operand.(*ast.Path); ok {
			if sym := c.PathSyms[p]; sym != nil {
				sym.Moved = true
			}
		}
		return t
	case *ast.Copy:
		t := c.checkExpr(x.Operand, nil)
		if len(x.RegionArgs) != 1 {
			c.err("E4302", x.Span, "`copy` 需要恰好一个区域实参", "分配点必须显式（铁律 1）", "写成 `copy x @[区域名]`")
		} else {
			c.checkRegionArg(x.RegionArgs[0], x.Span)
		}
		if len(x.RegionArgs) > 0 {
			c.Origins[x] = x.RegionArgs[0]
		}
		return t
	case *ast.Resume:
		// `resume h` 的值形态已退役（designs/0004 §6 v2）：协程等待一律走 select 的
		// `case 名 = resume h`——「还要继续等」不再作为用户可 match 的第三态出现。
		t := c.checkExpr(x.Operand, nil)
		if _, ok := t.(types.Coro); !ok {
			c.err("E3201", x.Span, "`resume` 需要协程句柄", "这里", "")
		}
		c.err("E5010", x.Span, "`resume` 只能作为 select 事件", "协程等待一律走 select（单 case select = 单路等待）",
			"写成 `select { case r = resume h => { ... } }`；绑定值就是完成态（`let v: T` / `let e: E`）")
		return types.I32
	case *ast.Recv:
		t := c.checkExpr(x.Chan, nil)
		ch, ok := t.(types.Chan)
		if !ok {
			c.err("E3201", x.Span, "`<-` 需要通道端", "这里", "")
			return types.I32
		}
		if !c.ctx.InCoroutine {
			c.err("E5001", x.Span, "`<-` 只在 `coroutine func` 内合法", "接收可能挂起", "")
		}
		return ch.Elem
	case *ast.Alloc:
		return c.checkAlloc(x, expected)
	case *ast.Match:
		return c.checkMatch(x, true, expected)
	case *ast.LayerCall:
		return c.checkLayerCall(x)
	}
	return types.I32
}

// --------------------------------------------------------------------------- //
// 名字解析
// --------------------------------------------------------------------------- //

func (c *Checker) resolvePath(p *ast.Path) (pathKind, any) {
	if len(p.Names) == 1 {
		n := p.Names[0]
		if c.ctx != nil {
			if sym := c.ctx.Scope.Lookup(n); sym != nil {
				return pkLocal, sym
			}
		}
		if gi, ok := c.sigs.Globals[n]; ok {
			return pkGlobal, gi
		}
		if ci, ok := c.sigs.Consts[n]; ok {
			return pkConst, ci
		}
		if sig, ok := c.sigs.Funcs[n]; ok {
			return pkFunc, sig
		}
		if n == "None" {
			return pkBuiltin, "None"
		}
		if _, ok := c.table.Types[n]; ok {
			return pkVariant, n
		}
		return pkUnknown, n
	}
	head := p.Names[0]
	// 枚举变体 Color.Red
	if t, ok := c.table.Lookup(head); ok {
		if named, ok2 := t.(types.Named); ok2 && len(p.Names) == 2 {
			if en, ok3 := named.Info.(*types.EnumInfo); ok3 {
				if v, ok4 := en.Variant(p.Names[1]); ok4 {
					return pkVariant, variantRef{enum: en, variant: v}
				}
			}
		}
	}
	if c.ctx != nil {
		if sym := c.ctx.Scope.Lookup(head); sym != nil {
			return pkField, fieldRef{sym: sym, fields: p.Names[1:]}
		}
	}
	if gi, ok := c.sigs.Globals[head]; ok {
		return pkField, fieldRef{global: gi, fields: p.Names[1:]}
	}
	// 跨模块：`绑定.名字` / `绑定.类型.变体`
	if dep, ok := c.imports[head]; ok && dep != nil {
		rest := p.Names[1:]
		if len(rest) == 1 {
			name := rest[0]
			if sig, ok := dep.Sigs.Funcs[name]; ok {
				c.requireVisible(p, sig.Name, "函数", dep.Name, sig.IsPublic, sig.InShell)
				return pkFunc, sig
			}
			if ci, ok := dep.Sigs.Consts[name]; ok {
				c.requireVisible(p, ci.Name, "常量", dep.Name, ci.IsPublic, ci.InShell)
				return pkConst, ci
			}
			if gi, ok := dep.Sigs.Globals[name]; ok {
				c.requireVisible(p, gi.Name, "模块级绑定", dep.Name, gi.IsPublic, gi.InShell)
				return pkGlobal, gi
			}
			if _, ok := dep.Table.Types[name]; ok {
				return pkVariant, head + "." + name
			}
		}
		if len(rest) == 2 {
			if t, ok := dep.Table.Lookup(rest[0]); ok {
				if named, ok2 := t.(types.Named); ok2 {
					if en, ok3 := named.Info.(*types.EnumInfo); ok3 {
						if v, ok4 := en.Variant(rest[1]); ok4 {
							c.requireVisible(p, en.Name, "枚举", dep.Name, en.IsPublic, en.InShell)
							return pkVariant, variantRef{enum: en, variant: v}
						}
					}
					if sig, ok4 := dep.Sigs.Methods[rest[0]+"."+rest[1]]; ok4 {
						c.requireVisible(p, rest[1], "方法", dep.Name, sig.IsPublic, sig.InShell)
						return pkFunc, sig
					}
				}
			}
		}
	}
	return pkUnknown, head
}

type variantRef struct {
	enum    *types.EnumInfo
	variant types.VariantInfo
}

type fieldRef struct {
	sym    *Symbol
	global *GlobalInfo
	fields []string
}

func (c *Checker) checkPath(p *ast.Path, expected types.Type) types.Type {
	kind, payload := c.resolvePath(p)
	switch kind {
	case pkLocal:
		sym := payload.(*Symbol)
		if sym.Moved {
			c.err("E4301", ast.SpanOf(p), "`"+sym.Name+"` 已被 `move` 转移，主名字已失效", "这里",
				"在 `move` 之前用完；要两份请 `copy x @[区域名]`")
		}
		sym.Used = true
		c.PathSyms[p] = sym
		if sym.OriginRegion != "" {
			c.Origins[p] = sym.OriginRegion
		}
		return sym.Type
	case pkGlobal:
		gi := payload.(*GlobalInfo)
		c.PathGlobals[p] = gi
		return gi.Type
	case pkConst:
		ci := payload.(*ConstInfo)
		c.PathConsts[p] = ci
		return ci.Type
	case pkFunc:
		sig := payload.(*FuncSig)
		c.PathFuncs[p] = sig
		params := make([]types.Type, 0, len(sig.Params))
		for _, prm := range sig.Params {
			params = append(params, prm.Type)
		}
		var ret types.Type = types.Void{}
		if sig.RetGroup != nil {
			ret = types.Group{Items: sig.RetGroup}
		} else {
			ret = sig.Ret
		}
		return types.FnPtr{Params: params, Ret: ret, Err: sig.Err}
	case pkField:
		fr := payload.(fieldRef)
		var cur types.Type
		if fr.sym != nil {
			cur = fr.sym.Type
			if sym := fr.sym; sym != nil {
				c.PathSyms[p] = sym
				if sym.OriginRegion != "" {
					c.Origins[p] = sym.OriginRegion
				}
			}
		} else {
			cur = fr.global.Type
		}
		steps := make([]FieldStep, 0, len(fr.fields))
		for _, fname := range fr.fields {
			named, ok := cur.(types.Named)
			if !ok {
				c.err("E3201", ast.SpanOf(p), "该类型没有字段 `"+fname+"`", "只有 struct 支持字段访问", "")
				return types.I32
			}
			info, ok := named.Info.(*types.StructInfo)
			if !ok {
				c.err("E3201", ast.SpanOf(p), "该类型没有字段 `"+fname+"`", "只有 struct 支持字段访问", "")
				return types.I32
			}
			found := false
			for i, f := range info.Fields {
				if f.Name == fname {
					steps = append(steps, FieldStep{Info: info, Index: i, Type: f.Type})
					cur = f.Type
					found = true
					break
				}
			}
			if !found {
				c.err("E3001", ast.SpanOf(p), "结构体 `"+info.Name+"` 没有字段 `"+fname+"`", "这里", "")
				return types.I32
			}
		}
		c.noteFieldChain(p, fr.sym, fr.global, steps)
		return cur
	case pkVariant:
		switch vr := payload.(type) {
		case variantRef:
			c.PathVariants[p] = vr
			return types.Named{Info: vr.enum}
		case string:
			c.err("E3201", ast.SpanOf(p), "类型 `"+vr+"` 不能当值使用", "这里", "")
			return types.I32
		}
	case pkBuiltin:
		if payload == "None" && expected != nil {
			if _, ok := expected.(types.Optional); ok {
				c.PathBuiltin[p] = "None"
				return expected
			}
		}
		c.err("E3201", ast.SpanOf(p), "`None` 需要可空类型的上下文", "例如 `let p: [U8]? = None;`", "")
		return types.Optional{Inner: types.View{Elem: types.U8}}
	}
	c.err("E3001", ast.SpanOf(p), "未定义的名字 `"+p.Dotted()+"`",
		"当前作用域 / 已定义类型与函数里都找不到", "")
	return types.I32
}
