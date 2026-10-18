package sema

import (
	"aila/internal/ast"
	"aila/internal/diag"
	"aila/internal/types"
)

// --------------------------------------------------------------------------- //
// 调用
// --------------------------------------------------------------------------- //

func sigReturn(sig *FuncSig) types.Type {
	var base types.Type = types.Void{}
	if sig.RetGroup != nil {
		base = types.Group{Items: sig.RetGroup}
	} else {
		base = sig.Ret
	}
	if sig.Err != nil {
		return types.Result{Ok: base, Err: sig.Err}
	}
	return base
}

func (c *Checker) findMethod(owner types.Named, name string) *FuncSig {
	key := owner.Info.TypeInfoName() + "." + name
	if sig, ok := c.sigs.Methods[key]; ok {
		return sig
	}
	// 跨模块：类型定义在别的模块里，方法也在那张表里
	ownerMod := ""
	switch info := owner.Info.(type) {
	case *types.StructInfo:
		ownerMod = info.Module
	case *types.EnumInfo:
		ownerMod = info.Module
	}
	if ownerMod != "" && ownerMod != c.modName {
		if u, ok := c.units[ownerMod]; ok && u.Sigs != nil {
			sig := u.Sigs.Methods[key]
			c.sigVisible(nil, sig)
			return sig
		}
	}
	return nil
}

func (c *Checker) checkCall(x *ast.Call, expected types.Type) types.Type {
	var sig *FuncSig
	var recv ast.Node
	calleePath, isPath := x.Callee.(*ast.Path)
	if isPath {
		kind, payload := c.resolvePath(calleePath)
		switch kind {
		case pkFunc:
			sig = payload.(*FuncSig)
		case pkLocal:
			sym := payload.(*Symbol)
			if len(calleePath.Names) >= 2 {
				// 方法调用：接收者是局部绑定
				if named, ok := sym.Type.(types.Named); ok {
					sig = c.findMethod(named, calleePath.Names[len(calleePath.Names)-1])
					if sig == nil {
						c.err("E3001", calleePath.Span,
							"类型 `"+named.Info.TypeInfoName()+"` 没有方法 `"+calleePath.Names[len(calleePath.Names)-1]+"`",
							"方法用 `func 类型.方法名(self, ...)` 声明", "")
					}
					recv = &ast.Path{Span: calleePath.Span, Names: calleePath.Names[:len(calleePath.Names)-1]}
					c.PathSyms[recv.(*ast.Path)] = sym
				} else {
					c.err("E3001", calleePath.Span, "该类型没有方法", "这里", "")
				}
			} else {
				// 单名：函数指针值间接调用
				t := c.checkExpr(calleePath, nil)
				if fp, ok := t.(types.FnPtr); ok {
					params := make([]Param, 0, len(fp.Params))
					for i, pt := range fp.Params {
						params = append(params, Param{Name: "a" + itoa(i), Type: pt})
					}
					ret := fp.Ret
					if ret == nil {
						ret = types.Void{}
					}
					sig = &FuncSig{Name: "<fnptr>", Params: params, Ret: ret, Err: fp.Err}
					c.CallKinds[x] = "fnptr"
				} else {
					c.err("E3201", calleePath.Span, "该值不可调用", "这里", "")
					return types.I32
				}
			}
		case pkField:
			fr := payload.(fieldRef)
			var cur types.Type
			if fr.sym != nil {
				cur = fr.sym.Type
				for _, fname := range fr.fields {
					if named, ok := cur.(types.Named); ok {
						if info, ok2 := named.Info.(*types.StructInfo); ok2 {
							for _, f := range info.Fields {
								if f.Name == fname {
									cur = f.Type
								}
							}
						}
					}
				}
				if named, ok := cur.(types.Named); ok {
					sig = c.findMethod(named, calleePath.Names[len(calleePath.Names)-1])
					if sig == nil {
						c.err("E3001", calleePath.Span, "方法未找到", "这里", "")
					}
					recv = &ast.Path{Span: calleePath.Span, Names: calleePath.Names[:len(calleePath.Names)-1]}
					c.PathSyms[recv.(*ast.Path)] = fr.sym
					c.CallRecv[x] = recv
				}
			}
		default:
			c.checkExpr(calleePath, nil)
			return types.I32
		}
	} else {
		t := c.checkExpr(x.Callee, nil)
		if fp, ok := t.(types.FnPtr); ok {
			params := make([]Param, 0, len(fp.Params))
			for i, pt := range fp.Params {
				params = append(params, Param{Name: "a" + itoa(i), Type: pt})
			}
			ret := fp.Ret
			if ret == nil {
				ret = types.Void{}
			}
			sig = &FuncSig{Name: "<fnptr>", Params: params, Ret: ret, Err: fp.Err}
			c.CallKinds[x] = "fnptr"
		} else {
			c.err("E3201", ast.SpanOf(x), "`"+typeName(t)+"` 不可调用", "这里", "")
			return types.I32
		}
	}
	if sig == nil {
		return types.I32
	}
	c.CallSigs[x] = sig
	if _, isFn := c.CallKinds[x]; !isFn {
		if sig.IsExtern {
			c.CallKinds[x] = "extern"
		} else if recv != nil {
			c.CallKinds[x] = "method"
		} else {
			c.CallKinds[x] = "func"
		}
	}

	params := sig.Params
	if len(params) > 0 && params[0].Name == "self" {
		params = params[1:]
	}
	if len(x.Args) != len(params) {
		if !(sig.Variadic && len(x.Args) >= len(params)) {
			c.err("E3201", x.Span, "`"+sig.FullName()+"` 需要与声明数量相等的实参", "AILA 无默认参数、无具名实参", "")
		}
	}
	for i, arg := range x.Args {
		if i >= len(params) {
			break
		}
		prm := params[i]
		at := c.checkExpr(arg, prm.Type)
		c.unify(prm.Type, at, ast.SpanOf(arg), "实参 `"+prm.Name+"`")
		if containsStr(sig.Muts, prm.Name) {
			if sym := c.placeSymbol(arg); sym != nil && !sym.Mutable {
				c.err("E4201", ast.SpanOf(arg), "`"+prm.Name+"` 需要写权，但该绑定是只读",
					"供给需持有：只读来源给不出写权", "")
			}
		}
		if containsStr(sig.Moves, prm.Name) {
			if _, isMove := arg.(*ast.Move); !isMove && types.IsMoveType(prm.Type) {
				c.err("E4301", ast.SpanOf(arg), "消耗参数必须显式 `move`", "双显式：签名 + 调用点各写一次", "")
			}
		}
	}
	if len(x.RegionArgs) != len(sig.Regions) {
		c.err("E4302", x.Span, "区域实参数量与签名 `@region[...]` 不符",
			"按序一一对应、数量相等，少给多给都不行", "")
	}
	for _, r := range x.RegionArgs {
		c.checkRegionArg(r, x.Span)
	}
	// self 的写权（@mut[self]）
	if recv != nil && containsStr(sig.Muts, "self") {
		if sym := c.placeSymbol(recv); sym != nil && !sym.Mutable {
			c.err("E4201", ast.SpanOf(recv), "该调用需要 `@mut[self]`，但接收者是只读绑定", "这里", "")
		}
	}
	return sigReturn(sig)
}

func (c *Checker) checkMethodCall(x *ast.MethodCall, expected types.Type) types.Type {
	recvT := c.checkExpr(x.Recv, nil)
	named, ok := recvT.(types.Named)
	if !ok {
		c.err("E3201", ast.SpanOf(x), "该类型没有方法 `"+x.Name+"`", "这里", "")
		return types.I32
	}
	sig := c.findMethod(named, x.Name)
	if sig == nil {
		c.err("E3001", x.Span, "类型 `"+named.Info.TypeInfoName()+"` 没有方法 `"+x.Name+"`", "这里",
			"方法用 `func 类型.方法名(self, ...)` 在 `module` 块内声明")
		return types.I32
	}
	c.CallSigs[x] = sig
	c.CallKinds[x] = "method"
	c.CallRecv[x] = x.Recv
	params := sig.Params
	if len(params) > 0 && params[0].Name == "self" {
		params = params[1:]
	}
	if len(x.Args) != len(params) {
		c.err("E3201", x.Span, "`"+sig.FullName()+"` 需要与声明数量相等的实参", "这里", "")
	}
	for i, arg := range x.Args {
		if i >= len(params) {
			break
		}
		prm := params[i]
		at := c.checkExpr(arg, prm.Type)
		c.unify(prm.Type, at, ast.SpanOf(arg), "实参 `"+prm.Name+"`")
	}
	if len(x.RegionArgs) != len(sig.Regions) {
		c.err("E4302", x.Span, "区域实参数量与签名不符", "这里", "")
	}
	for _, r := range x.RegionArgs {
		c.checkRegionArg(r, x.Span)
	}
	if containsStr(sig.Muts, "self") {
		if sym := c.placeSymbol(x.Recv); sym != nil && !sym.Mutable {
			c.err("E4201", ast.SpanOf(x), "该调用需要 `@mut[self]`，但接收者是只读绑定", "这里", "")
		}
	}
	return sigReturn(sig)
}

// --------------------------------------------------------------------------- //
// 下标 / 字段 / 运算符
// --------------------------------------------------------------------------- //

func (c *Checker) checkIndex(x *ast.Index) types.Type {
	base := c.checkExpr(x.Base, nil)
	idx := c.checkExpr(x.Index_, types.USize)
	if _, ok := idx.(types.UntypedInt); ok {
		c.unify(types.USize, idx, ast.SpanOf(x), "下标")
	} else if !types.IsInteger(idx) {
		c.err("E3201", ast.SpanOf(x), "下标必须是整数（推荐 USize）", "这里", "")
	}
	switch b := base.(type) {
	case types.View:
		c.noteIndex(x, base, "view", b.Elem)
		return b.Elem
	case types.Fixed:
		c.noteIndex(x, base, "fixed", b.Elem)
		return b.Elem
	case types.Prim:
		if b.K == "text" {
			c.noteIndex(x, base, "text", types.U8)
			return types.U8
		}
	case types.UnsafeView:
		c.err("E3201", ast.SpanOf(x), "`[Unsafe]` 不能解引用", "先 `copy raw @[区域名]` 定型", "")
		return types.U8
	case types.Optional:
		c.err("E3201", ast.SpanOf(x), "可空视图必须先 match 解包再下标", "这里", "")
		return types.U8
	}
	c.err("E3201", ast.SpanOf(x), "`"+typeName(base)+"` 不支持下标", "这里", "")
	return types.U8
}

func (c *Checker) checkField(x *ast.FieldAccess) types.Type {
	base := c.checkExpr(x.Base, nil)
	named, ok := base.(types.Named)
	if !ok {
		c.err("E3201", ast.SpanOf(x), "该类型不支持字段访问", "这里", "")
		return types.I32
	}
	info, ok := named.Info.(*types.StructInfo)
	if !ok {
		c.err("E3201", ast.SpanOf(x), "只有 struct 支持字段访问", "这里", "")
		return types.I32
	}
	for i, f := range info.Fields {
		if f.Name == x.Name {
			c.noteField(x, base, info, i, f.Type)
			return f.Type
		}
	}
	c.err("E3001", x.Span, "结构体 `"+info.Name+"` 没有字段 `"+x.Name+"`", "这里", "")
	return types.I32
}

func (c *Checker) checkBinary(x *ast.Binary, expected types.Type) types.Type {
	if x.Op == "&&" || x.Op == "||" {
		lt := c.checkExpr(x.Left, types.Bool)
		rt := c.checkExpr(x.Right, types.Bool)
		c.unify(types.Bool, lt, ast.SpanOf(x), "逻辑运算左操作数")
		c.unify(types.Bool, rt, ast.SpanOf(x), "逻辑运算右操作数")
		return types.Bool
	}
	var lt, rt types.Type
	if expected != nil && (types.IsInteger(expected) || types.IsFloat(expected)) {
		lt = c.checkExpr(x.Left, expected)
		rt = c.checkExpr(x.Right, lt)
	} else {
		lt = c.checkExpr(x.Left, nil)
		rt = c.checkExpr(x.Right, nil)
	}
	var opT types.Type
	switch {
	case isUntypedInt(lt) && isUntypedInt(rt):
		if expected != nil && (types.IsInteger(expected) || types.IsFloat(expected)) {
			opT = expected
		} else {
			opT = types.I32
		}
	case isUntypedInt(lt):
		opT = rt
		c.unify(rt, lt, ast.SpanOf(x), "字面量定型")
	case isUntypedInt(rt):
		opT = lt
		c.unify(lt, rt, ast.SpanOf(x), "字面量定型")
	case isUntypedFloat(lt) && types.IsFloat(rt):
		opT = rt
	case isUntypedFloat(rt) && types.IsFloat(lt):
		opT = lt
	case isUntypedFloat(lt) && isUntypedFloat(rt):
		opT = types.F64
	default:
		opT = lt
		c.unify(lt, rt, ast.SpanOf(x), "`"+x.Op+"` 的操作数")
	}
	if isCmp(x.Op) {
		if !isComparable(opT) {
			c.err("E3205", x.Span, "`"+typeName(opT)+"` 不支持 `"+x.Op+"` 比较",
				"视图 / Text / struct / 带载荷 enum / 函数指针都没有比较", "")
		}
		c.BinOperands[x] = opT
		return types.Bool
	}
	if types.IsFloat(opT) {
		if x.Op == "%" || x.Op == "&" || x.Op == "|" || x.Op == "^" || x.Op == "<<" || x.Op == ">>" {
			c.err("E3201", x.Span, "浮点不支持 `"+x.Op+"`", "浮点余数走 `std.math.fmod`", "")
		}
	} else if !types.IsInteger(opT) {
		c.err("E3201", x.Span, "`"+typeName(opT)+"` 不支持 `"+x.Op+"`", "算术 / 位运算只对整数与浮点", "")
	}
	if isUntypedInt(opT) || isUntypedFloat(opT) {
		opT = types.I32
	}
	c.BinOperands[x] = opT
	return opT
}

func isUntypedInt(t types.Type) bool { _, ok := t.(types.UntypedInt); return ok }
func isUntypedFloat(t types.Type) bool { _, ok := t.(types.UntypedFloat); return ok }

func isCmp(op string) bool {
	switch op {
	case "==", "!=", "<", "<=", ">", ">=":
		return true
	}
	return false
}

func isComparable(t types.Type) bool {
	switch x := t.(type) {
	case types.Prim:
		return x.K != "text"
	case types.Named:
		if en, ok := x.Info.(*types.EnumInfo); ok {
			return !en.Payload
		}
	}
	return false
}

func (c *Checker) checkUnary(x *ast.Unary, expected types.Type) types.Type {
	if x.Op == "!" {
		t := c.checkExpr(x.Operand, types.Bool)
		c.unify(types.Bool, t, ast.SpanOf(x), "`!` 操作数")
		return types.Bool
	}
	var exp types.Type
	switch expected.(type) {
	case types.Result, types.Optional:
	default:
		exp = expected
	}
	t := c.checkExpr(x.Operand, exp)
	if x.Op == "-" {
		if ui, ok := t.(types.UntypedInt); ok {
			return types.UntypedInt{Value: -ui.Value}
		}
		if !types.IsFloat(t) && !types.IsInteger(t) {
			c.err("E3201", x.Span, "`-` 只作用于数值", "这里", "")
			return types.I32
		}
		return t
	}
	if x.Op == "~" {
		if !types.IsInteger(t) {
			c.err("E3201", x.Span, "`~` 只作用于整数", "这里", "")
			return types.I32
		}
		return t
	}
	return t
}

func (c *Checker) checkCast(x *ast.Cast) types.Type {
	src := c.checkExpr(x.Operand, nil)
	if _, ok := src.(types.UntypedInt); ok {
		src = types.I32
	}
	dst := c.table.Resolve(x.Target, c.bag)
	if _, ok := dst.(types.Prim); !ok || isText(dst) {
		c.err("E3201", x.Span, "转换目标必须是标量类型", "视图 / Text 的类型变换走 `view` 声明", "")
		return dst
	}
	lossless := castLossless(src, dst)
	if x.KindV == "as" {
		if !lossless {
			c.err("E3202", x.Span, "`as` 不能做有损 / 跨域转换",
				"`as` 只做位型无损解释（同宽互释 / 拓宽）",
				"有损方向用 `trunc`；钳制 / 失败返回走 `std.math.saturate_*` / `try_*`")
		}
	}
	c.CastTypes[x] = [2]types.Type{src, dst}
	return dst
}

func isText(t types.Type) bool {
	p, ok := t.(types.Prim)
	return ok && p.K == "text"
}

func castLossless(src, dst types.Type) bool {
	sb := types.IntWidth(src, 8)
	db := types.IntWidth(dst, 8)
	if sb == 0 || db == 0 {
		return false
	}
	sf, df := types.IsFloat(src), types.IsFloat(dst)
	if sf && df {
		return db >= sb
	}
	if sf != df {
		return false
	}
	if sb == db {
		return true
	}
	if db < sb {
		return false
	}
	if types.IsSigned(src) && !types.IsSigned(dst) {
		return false
	}
	return true
}

func (c *Checker) checkTryPropagate(x *ast.TryPropagate) types.Type {
	inner := c.checkExpr(x.Operand, nil)
	sig := c.ctx.Sig
	switch t := inner.(type) {
	case types.Result:
		if sig.Err == nil {
			c.err("E3201", x.Span, "`?` 需要函数返回 `T ! E` 才能上抛", "这里", "")
		} else if !sameType(sig.Err, t.Err) {
			c.err("E3201", x.Span, "`?` 上抛的错误类型不匹配", "零隐式转换", "")
		}
		c.TryModes[x] = "result"
		c.noteTry(x, "result", t.Ok)
		return t.Ok
	case types.Optional:
		if _, ok := sig.Ret.(types.Optional); !ok {
			c.err("E3201", x.Span, "`?` 需要函数返回 `T?` 才能上抛 None", "这里", "")
		}
		c.TryModes[x] = "optional"
		c.noteTry(x, "optional", t.Inner)
		return t.Inner
	}
	c.err("E3201", x.Span, "`?` 只能作用于 `T ! E` 或 `T?`", "这里", "")
	return types.I32
}

// --------------------------------------------------------------------------- //
// 字面量构造 / 分配
// --------------------------------------------------------------------------- //

// checkArrayLit 记录元素类型与常量长度后转交真正的检查。
func (c *Checker) checkArrayLit(x *ast.ArrayLit, expected types.Type) types.Type {
	t := c.checkArrayLitInner(x, expected)
	var elem types.Type
	if v, ok := t.(types.View); ok {
		elem = v.Elem
	}
	if f, ok := t.(types.Fixed); ok {
		elem = f.Elem
	}
	count, hasCount := int64(0), false
	if f, ok := t.(types.Fixed); ok {
		count, hasCount = f.N, true
	} else if x.Repeat != nil {
		if n, ok := constInt(x.Count); ok {
			count, hasCount = n, true
		}
	}
	c.noteArrayElem(x, elem, count, hasCount)
	return t
}

func (c *Checker) checkArrayLitInner(x *ast.ArrayLit, expected types.Type) types.Type {
	var elemExpected types.Type
	if expected != nil {
		switch e := expected.(type) {
		case types.View:
			elemExpected = e.Elem
		case types.Fixed:
			elemExpected = e.Elem
		case types.Optional:
			if v, ok := e.Inner.(types.View); ok {
				elemExpected = v.Elem
			}
		}
	}
	if x.Repeat != nil {
		elem := c.checkExpr(x.Repeat, elemExpected)
		count, isConst := constInt(x.Count)
		if !isConst {
			if x.Count != nil {
				ct := c.checkExpr(x.Count, types.USize)
				if _, ok := ct.(types.UntypedInt); ok {
					c.unify(types.USize, ct, ast.SpanOf(x.Count), "数组长度")
				} else {
					c.requireInteger(ct, ast.SpanOf(x.Count), "数组长度")
				}
			}
			if _, isView := expected.(types.View); !isView {
				c.err("E3201", x.Span, "定长数组的长度必须是编译期整数常量", "这里", "")
			}
			return types.View{Elem: elem}
		}
		if f, ok := expected.(types.Fixed); ok {
			if f.N != count {
				c.err("E3201", x.Span, "字面量长度与标注不符", "长度必须一致", "")
			}
			c.unify(f.Elem, elem, x.Span, "数组元素")
			return f
		}
		if v, ok := expected.(types.View); ok {
			c.unify(v.Elem, elem, x.Span, "数组元素")
			return types.View{Elem: v.Elem}
		}
		return types.Fixed{Elem: elem, N: count}
	}
	if len(x.Elems) == 0 {
		if f, ok := expected.(types.Fixed); ok {
			return f
		}
		c.err("E3201", x.Span, "空数组字面量没有上下文无法定型", "这里", "")
		return types.Fixed{Elem: types.I32}
	}
	var elem types.Type
	for _, el := range x.Elems {
		t := c.checkExpr(el, elemExpected)
		if elem == nil {
			elem = t
		} else if !sameType(elem, t) {
			if isUntypedInt(elem) && types.IsInteger(t) {
				elem = t
			} else if !isUntypedInt(t) {
				c.err("E3201", x.Span, "数组元素类型不一致", "这里", "")
			}
		}
	}
	if isUntypedInt(elem) {
		elem = types.I32
	}
	if f, ok := expected.(types.Fixed); ok {
		if int(f.N) != len(x.Elems) {
			c.err("E3201", x.Span, "字面量长度与标注不符", "长度必须一致", "")
		}
		c.unify(f.Elem, elem, x.Span, "数组元素")
		return f
	}
	return types.View{Elem: elem}
}

func (c *Checker) checkStructLit(x *ast.StructLit, expected types.Type) types.Type {
	names := x.TypePath
	if len(names) == 1 {
		if t, ok := c.table.Lookup(names[0]); ok {
			if named, isNamed := t.(types.Named); isNamed {
				info, isStruct := named.Info.(*types.StructInfo)
				if !isStruct {
					c.err("E3201", x.Span, "enum 构造必须写 `类型.变体 { ... }`", "这里", "")
					return t
				}
				c.StructLits[x] = info
				c.noteLit(x, &LitRef{Struct: info})
				seen := map[string]bool{}
				for _, pair := range x.FieldPairs {
					fname, _ := pair.A.(string)
					seen[fname] = true
					var ftype types.Type
					for _, f := range info.Fields {
						if f.Name == fname {
							ftype = f.Type
						}
					}
					if ftype == nil {
						c.err("E3001", x.Span, "结构体 `"+info.Name+"` 没有字段 `"+fname+"`", "这里", "")
						continue
					}
					fe, ok := pair.B.(ast.Node)
					if !ok {
						continue
					}
					ft := c.checkExpr(fe, ftype)
					c.unify(ftype, ft, ast.SpanOf(fe), "字段 `"+fname+"`")
				}
				for _, f := range info.Fields {
					if !seen[f.Name] {
						c.err("E3201", x.Span, "构造 `"+info.Name+"` 缺少字段 `"+f.Name+"`",
							"构造一律写全（铁律 10）", "")
					}
				}
				return named
			}
		}
		if names[0] == "Some" || names[0] == "Ok" || names[0] == "Err" {
			return c.checkBuiltinLit(x, names[0], expected)
		}
		c.err("E3001", x.Span, "未知类型 `"+names[0]+"`", "构造必须类型前缀", "")
		return types.I32
	}
	// 跨包构造：`绑定.类型 { ... }` / `绑定.类型.变体 { ... }`（docs/rfcs/0004）
	if len(names) >= 2 {
		if dep, isDep := c.imports[names[0]]; isDep && dep != nil {
			if info, ok := dep.Table.Types[names[1]]; ok {
				switch it := info.(type) {
				case *types.StructInfo:
					if len(names) > 2 {
						c.err("E3001", x.Span, "struct 没有变体 `"+names[2]+"`", "这里", "")
						return types.Named{Info: it}
					}
					c.requireVisible(x, names[0]+"."+names[1], "类型", dep.Name, it.IsPublic, it.InShell)
					return c.checkStructFields(x, types.Named{Info: it}, it, nil)
				case *types.EnumInfo:
					if len(names) != 3 {
						c.err("E3201", x.Span, "enum 构造必须写 `类型.变体 { ... }`", "这里", "")
						return types.Named{Info: it}
					}
					v, ok2 := it.Variant(names[2])
					if !ok2 {
						c.err("E3001", x.Span, "enum `"+it.Name+"` 没有变体 `"+names[2]+"`", "这里", "")
						return types.Named{Info: it}
					}
					c.requireVisible(x, names[0]+"."+names[1], "枚举", dep.Name, it.IsPublic, it.InShell)
					return c.checkEnumVariant(x, types.Named{Info: it}, it, v)
				}
			}
		}
	}
	if len(names) == 2 {
		t, ok := c.table.Lookup(names[0])
		if !ok {
			c.err("E3001", x.Span, "`"+names[0]+"` 不是可构造的 enum 类型", "这里", "")
			return types.I32
		}
		named := t.(types.Named)
		en, ok := named.Info.(*types.EnumInfo)
		if !ok {
			c.err("E3001", x.Span, "`"+names[0]+"` 不是 enum", "这里", "")
			return types.I32
		}
		v, ok := en.Variant(names[1])
		if !ok {
			c.err("E3001", x.Span, "enum `"+en.Name+"` 没有变体 `"+names[1]+"`", "这里", "")
			return named
		}
		c.VariantLits[x] = variantRef{enum: en, variant: v}
		c.noteLit(x, &LitRef{Enum: en, Variant: v, HasVar: true})
		if v.Payload == nil {
			if len(x.FieldPairs) > 0 {
				c.err("E3201", x.Span, "该变体没有载荷", "这里", "")
			}
			return named
		}
		seen := map[string]bool{}
		for _, pair := range x.FieldPairs {
			fname, _ := pair.A.(string)
			seen[fname] = true
			var ftype types.Type
			for _, f := range v.Payload {
				if f.Name == fname {
					ftype = f.Type
				}
			}
			if ftype == nil {
				c.err("E3001", x.Span, "变体没有字段 `"+fname+"`", "这里", "")
				continue
			}
			if fe, ok := pair.B.(ast.Node); ok {
				ft := c.checkExpr(fe, ftype)
				c.unify(ftype, ft, ast.SpanOf(fe), "字段 `"+fname+"`")
			}
		}
		for _, f := range v.Payload {
			if !seen[f.Name] {
				c.err("E3201", x.Span, "构造缺少字段 `"+f.Name+"`", "构造一律写全（铁律 10）", "")
			}
		}
		return named
	}
	c.err("E3001", x.Span, "无法解析构造类型", "这里", "")
	return types.I32
}

func (c *Checker) checkBuiltinLit(x *ast.StructLit, which string, expected types.Type) types.Type {
	get := func(name string) ast.Node {
		for _, pair := range x.FieldPairs {
			if pair.A.(string) == name {
				if n, ok := pair.B.(ast.Node); ok {
					return n
				}
			}
		}
		return nil
	}
	switch which {
	case "Some":
		var inner types.Type
		if opt, ok := expected.(types.Optional); ok {
			inner = opt.Inner
		}
		v := get("v")
		if v == nil {
			c.err("E3201", x.Span, "`Some` 必须写载荷 `{ v: 值 }`", "构造显式（铁律 10）", "")
			inner = types.I32
		} else {
			t := c.checkExpr(v, inner)
			if inner != nil {
				c.unify(inner, t, ast.SpanOf(v), "`Some` 载荷")
			} else {
				inner = t
			}
		}
		c.BuiltinLits[x] = "Some"
		c.noteLit(x, &LitRef{Builtin: "Some", Payload: inner})
		return types.Optional{Inner: inner}
	case "Ok":
		var ok types.Type
		var errT types.Type = types.Named{Info: &types.EnumInfo{Name: "Err"}}
		if r, isRes := expected.(types.Result); isRes {
			ok, errT = r.Ok, r.Err
		}
		v := get("v")
		if v != nil {
			t := c.checkExpr(v, ok)
			if ok != nil {
				c.unify(ok, t, ast.SpanOf(v), "`Ok` 载荷")
			} else {
				ok = t
			}
		}
		c.BuiltinLits[x] = "Ok"
		c.noteLit(x, &LitRef{Builtin: "Ok", Payload: ok})
		return types.Result{Ok: ok, Err: errT}
	default: // Err
		var okT types.Type = types.I32
		var errT types.Type
		if r, isRes := expected.(types.Result); isRes {
			okT, errT = r.Ok, r.Err
		}
		v := get("e")
		if v != nil {
			t := c.checkExpr(v, errT)
			if errT != nil {
				c.unify(errT, t, ast.SpanOf(v), "`Err` 载荷")
			} else {
				errT = t
			}
		}
		c.BuiltinLits[x] = "Err"
		c.noteLit(x, &LitRef{Builtin: "Err", Payload: errT})
		return types.Result{Ok: okT, Err: errT}
	}
}

// checkAlloc 记录元素类型后转交真正的检查。
func (c *Checker) checkAlloc(x *ast.Alloc, expected types.Type) types.Type {
	t := c.checkAllocInner(x, expected)
	ref := &AllocRef{}
	if v, ok := t.(types.View); ok {
		ref.Elem = v.Elem
	} else if o, ok := t.(types.Optional); ok {
		if v, ok := o.Inner.(types.View); ok {
			ref.Elem = v.Elem
		}
	}
	if inner, ok := x.Expr.(*ast.ArrayLit); ok {
		if n, isConst := constInt(inner.Count); isConst {
			ref.Count, ref.HasCount = n, true
		} else if inner.Repeat == nil {
			ref.Count, ref.HasCount = int64(len(inner.Elems)), true
		}
	}
	c.noteAlloc(x, ref)
	return t
}

func (c *Checker) checkAllocInner(x *ast.Alloc, expected types.Type) types.Type {
	c.checkRegionArg(x.Region, x.Span)
	inner, ok := x.Expr.(*ast.ArrayLit)
	if !ok {
		c.err("E2007", x.Span, "分配标注 `@区域名` 只能贴在数组字面量上", "这里",
			"调用点请写 `f(...)@[区域名]`")
		t := c.checkExpr(x.Expr, expected)
		return t
	}
	var elemExpected types.Type
	if v, isView := expected.(types.View); isView {
		elemExpected = v.Elem
	}
	if e, isOpt := expected.(types.Optional); isOpt {
		if v, isView := e.Inner.(types.View); isView {
			elemExpected = v.Elem
		}
	}
	t := c.checkArrayLit(inner, types.View{Elem: elemExpectedOr(elemExpected)})
	c.Origins[x] = x.Region
	if v, isView := t.(types.View); isView {
		return v
	}
	return types.View{Elem: types.I32}
}

func elemExpectedOr(t types.Type) types.Type {
	if t == nil {
		return nil
	}
	return t
}

func (c *Checker) checkRegionArg(name string, span diag.Span) {
	if name == "" {
		c.err("E4001", span, "缺少区域实参", "分配点必须显式（铁律 1）", "")
		return
	}
	if c.ctx == nil || !c.ctx.RegionVisible(name) {
		c.err("E4001", span, "区域 `"+name+"` 在此不可见",
			"函数内可见区域 = `@region` 参数 / 体内 `region` 块 / `@static`",
			"给函数签名加 `@region[r]` 并在调用点传 `@[work]`")
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
// checkStructFields 校验 `类型 { 字段: 值, ... }` 并对字段表达式做类型检查。
//
// 供「本包 struct」与「跨包 struct」两条路径共用（字段一律写全，铁律 10）。
func (c *Checker) checkStructFields(x *ast.StructLit, named types.Named, info *types.StructInfo,
	_ []types.Field) types.Type {
	c.StructLits[x] = info
	c.noteLit(x, &LitRef{Struct: info})
	seen := map[string]bool{}
	for _, pair := range x.FieldPairs {
		fname, _ := pair.A.(string)
		seen[fname] = true
		var ftype types.Type
		for _, f := range info.Fields {
			if f.Name == fname {
				ftype = f.Type
			}
		}
		if ftype == nil {
			c.err("E3001", x.Span, "结构体 `"+info.Name+"` 没有字段 `"+fname+"`", "这里", "")
			continue
		}
		fe, ok := pair.B.(ast.Node)
		if !ok {
			continue
		}
		ft := c.checkExpr(fe, ftype)
		c.unify(ftype, ft, ast.SpanOf(fe), "字段 `"+fname+"`")
	}
	for _, f := range info.Fields {
		if !seen[f.Name] {
			c.err("E3201", x.Span, "构造 `"+info.Name+"` 缺少字段 `"+f.Name+"`",
				"构造一律写全（铁律 10）", "")
		}
	}
	return named
}

// checkEnumVariant 校验 `枚举.变体 { 字段: 值, ... }`。
func (c *Checker) checkEnumVariant(x *ast.StructLit, named types.Named, en *types.EnumInfo,
	v types.VariantInfo) types.Type {
	c.VariantLits[x] = variantRef{enum: en, variant: v}
	c.noteLit(x, &LitRef{Enum: en, Variant: v, HasVar: true})
	if v.Payload == nil {
		if len(x.FieldPairs) > 0 {
			c.err("E3201", x.Span, "该变体没有载荷", "这里", "")
		}
		return named
	}
	seen := map[string]bool{}
	for _, pair := range x.FieldPairs {
		fname, _ := pair.A.(string)
		seen[fname] = true
		var ftype types.Type
		for _, f := range v.Payload {
			if f.Name == fname {
				ftype = f.Type
			}
		}
		if ftype == nil {
			c.err("E3001", x.Span, "变体没有字段 `"+fname+"`", "这里", "")
			continue
		}
		if fe, ok := pair.B.(ast.Node); ok {
			ft := c.checkExpr(fe, ftype)
			c.unify(ftype, ft, ast.SpanOf(fe), "字段 `"+fname+"`")
		}
	}
	for _, f := range v.Payload {
		if !seen[f.Name] {
			c.err("E3201", x.Span, "构造缺少字段 `"+f.Name+"`", "构造一律写全（铁律 10）", "")
		}
	}
	return named
}
