package codegen

import (
	"strconv"
	"strings"

	"aila/internal/ast"
	"aila/internal/diag"
	"aila/internal/sema"
	"aila/internal/types"
)

// lowerExpr 把表达式降成 (IR 类型, SSA 值)。
func (g *funcGen) lowerExpr(e ast.Node) valueRef {
	t := g.ty(e)
	lt := g.l.llty(t)
	switch x := e.(type) {
	case *ast.IntLit:
		return valueRef{lt, intLiteral(x.Raw, lt)}
	case *ast.FloatLit:
		return valueRef{lt, floatLit(x.Value, llvmFloatBits(lt))}
	case *ast.BoolLit:
		return valueRef{lt, boolLit(x.Value)}
	case *ast.StrLit:
		sym := g.l.stringConst(x.Value)
		r := g.tmp()
		g.emit(r + " = insertvalue " + g.l.viewLT() + " undef, ptr " + sym + ", 0")
		r2 := g.tmp()
		g.emit(r2 + " = insertvalue " + g.l.viewLT() + " " + r + ", " + g.l.lenLT() + " " + strconv.Itoa(len(x.Value)) + ", 1")
		return valueRef{g.l.viewLT(), r2}
	case *ast.CharLit:
		return valueRef{lt, strconv.FormatInt(runeValue(x.Value), 10)}
	case *ast.Path:
		return g.lowerPath(x)
	case *ast.GroupExpr:
		// 多值组：每个元素都要落到组类型的对应槽位宽度上。
		// 元素表达式的推断类型可能更宽（无标注字面量默认 i64），直接 insertvalue 会
		// 生成 `insertvalue {i32,i32} … , i64 …` 这种非法 IR（32 位目标上必撞）。
		// 签名优先：返回位置的多值组以**声明返回类型**为准（字面量元素可能被推断成
		// 更宽的类型，例如无标注整数字面量默认 I64，而签名写的是 U32）。
		var items []types.Type
		if g.sig != nil && len(g.sig.RetGroup) == len(x.Elems) {
			items = g.sig.RetGroup
		} else if grp, ok := g.ty(x).(types.Group); ok {
			items = grp.Items
		}
		itemLL := make([]string, 0, len(items))
		for _, it := range items {
			itemLL = append(itemLL, g.l.llty(it))
		}
		aggLt := lt
		if len(items) > 0 {
			aggLt = g.l.llty(types.Group{Items: items})
		}
		parts := make([]valueRef, 0, len(x.Elems))
		for i, item := range x.Elems {
			v := g.lowerExpr(item)
			if i < len(itemLL) && v.lt != itemLL[i] && isIntLL(v.lt) && isIntLL(itemLL[i]) {
				v = valueRef{itemLL[i], g.coerceInt(v.lt, itemLL[i], v.s)}
			}
			parts = append(parts, v)
		}
		if len(parts) == 0 {
			return valueRef{aggLt, "undef"}
		}
		r := g.tmp()
		g.emit(r + " = insertvalue " + aggLt + " undef, " + parts[0].lt + " " + parts[0].s + ", 0")
		for i := 1; i < len(parts); i++ {
			r2 := g.tmp()
			g.emit(r2 + " = insertvalue " + aggLt + " " + r + ", " + parts[i].lt + " " + parts[i].s +
				", " + strconv.Itoa(i))
			r = r2
		}
		return valueRef{aggLt, r}
	case *ast.ArrayLit:
		return g.lowerArrayLit(x, lt)
	case *ast.StructLit:
		return g.lowerStructLit(x, lt)
	case *ast.Call:
		return g.lowerCall(x, x.Callee, x.Args, x.RegionArgs, lt)
	case *ast.MethodCall:
		return g.lowerCall(x, nil, x.Args, x.RegionArgs, lt)
	case *ast.Index:
		plt, ptr := g.lowerPlace(x)
		v := g.tmp()
		g.emit(v + " = load " + plt + ", ptr " + ptr)
		return valueRef{plt, v}
	case *ast.FieldAccess:
		plt, ptr := g.lowerPlace(x)
		v := g.tmp()
		g.emit(v + " = load " + plt + ", ptr " + ptr)
		return valueRef{plt, v}
	case *ast.Binary:
		return g.lowerBinary(x)
	case *ast.Unary:
		return g.lowerUnary(x, lt)
	case *ast.Copy:
		return g.lowerExpr(x.Operand)
	case *ast.Move:
		return g.lowerExpr(x.Operand)
	case *ast.TryPropagate:
		return g.lowerTryPropagate(x, lt)
	case *ast.Alloc:
		return g.lowerAlloc(x)
	case *ast.Match:
		s := g.alloca("__match", lt)
		g.lowerMatch(x, s)
		v := g.tmp()
		g.emit(v + " = load " + lt + ", ptr " + s)
		return valueRef{lt, v}
	case *ast.Recv:
		return g.lowerRecv(x)
	case *ast.LayerCall:
		return g.lowerLayerCall(x, lt)
	}
	g.unsupported("表达式 " + kindOf(e) + " 的代码生成待批次")
	return valueRef{lt, g.zero(t)}
}

func runeValue(s string) int64 {
	r := []rune(s)
	if len(r) == 0 {
		return 0
	}
	return int64(r[0])
}

// --------------------------------------------------------------------------- //
// 路径与左值
// --------------------------------------------------------------------------- //

// lowerPath 生成路径取值（局部 / 常量 / 模块级 / 函数 / 变体）。
func (g *funcGen) lowerPath(p *ast.Path) valueRef {
	t := g.ty(p)
	lt := g.l.llty(t)
	if chain := g.l.check.FieldChainOf(p); chain != nil {
		plt, ptr := g.lowerPlace(p)
		v := g.tmp()
		g.emit(v + " = load " + plt + ", ptr " + ptr)
		return valueRef{plt, v}
	}
	if sym, ok := g.l.check.PathSyms[p]; ok && sym != nil {
		s := g.slotFor(sym)
		if s == nil {
			s = g.namedSlots[firstName(p)]
		}
		if s == nil {
			return valueRef{lt, g.zero(t)}
		}
		v := g.tmp()
		g.emit(v + " = load " + s.lt + ", ptr " + s.ptr)
		return valueRef{s.lt, v}
	}
	if ci, ok := g.l.check.PathConsts[p]; ok && ci != nil {
		return g.constValue(ci, lt, t)
	}
	if gi, ok := g.l.check.PathGlobals[p]; ok && gi != nil {
		sym := g.globalSymbol(gi)
		l := g.l
		l.ensureGlobal(gi.Name, sym, lt)
		v := g.tmp()
		g.emit(v + " = load " + lt + ", ptr " + sym)
		return valueRef{lt, v}
	}
	if sig, ok := g.l.check.PathFuncs[p]; ok && sig != nil {
		sym := g.l.symbolOf(sig)
		ret, params := g.l.funcLLType(sig)
		g.l.ensureDeclare(sym, ret+" @"+sym+"("+strings.Join(params, ", ")+")")
		return valueRef{"ptr", "@" + sym}
	}
	if enum, v, ok := g.l.check.PathVariant(p); ok && enum != nil {
		if enum.Payload {
			tagLt := g.l.llty(enum.TagType)
			r := g.tmp()
			g.emit(r + " = insertvalue " + lt + " undef, " + tagLt + " " + strconv.FormatInt(v.Tag, 10) + ", 0")
			return valueRef{lt, r}
		}
		return valueRef{lt, strconv.FormatInt(v.Tag, 10)}
	}
	if b, ok := g.l.check.PathBuiltin[p]; ok && b == "None" {
		return valueRef{lt, g.zero(t)}
	}
	return valueRef{lt, g.zero(t)}
}

func firstName(p *ast.Path) string {
	if len(p.Names) > 0 {
		return p.Names[0]
	}
	return ""
}

// constValue 内联常量（跨模块常量也不依赖本模块 consts 表）。
func (g *funcGen) constValue(ci *sema.ConstInfo, lt string, t types.Type) valueRef {
	switch {
	case isTextType(ci.Type):
		text := strconv.Quote("")
		_ = text
		data := strValue(ci.Value)
		sym := g.l.stringConst(data)
		r := g.tmp()
		g.emit(r + " = insertvalue " + g.l.viewLT() + " undef, ptr " + sym + ", 0")
		r2 := g.tmp()
		g.emit(r2 + " = insertvalue " + g.l.viewLT() + " " + r + ", " + g.l.lenLT() + " " + strconv.Itoa(len(data)) + ", 1")
		return valueRef{lt, r2}
	case isBoolType(ci.Type):
		return valueRef{lt, boolLit(toBool(ci.Value))}
	case types.IsFloat(ci.Type):
		return valueRef{lt, floatLit(toFloat(ci.Value), llvmFloatBits(lt))}
	}
	if v, ok := asInt(ci.Value); ok {
		return valueRef{lt, intLiteral(v, lt)}
	}
	return valueRef{lt, g.zero(t)}
}

func strValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// globalSymbol 生成模块级变量符号名。
func (g *funcGen) globalSymbol(gi *sema.GlobalInfo) string {
	flat := strings.ReplaceAll(g.l.modName, "_.", "")
	flat = strings.ReplaceAll(flat, ".", "_")
	return "@g_" + escName(flat) + "_" + escName(gi.Name)
}

// lowerPlace 左值位置：返回 (llty, 指向内存的指针)。
func (g *funcGen) lowerPlace(e ast.Node) (string, string) {
	switch x := e.(type) {
	case *ast.Path:
		t := g.ty(x)
		lt := g.l.llty(t)
		if chain := g.l.check.FieldChainOf(x); chain != nil {
			var baseLt, basePtr string
			if chain.Base != nil {
				s := g.slotFor(chain.Base)
				if s == nil {
					s = g.namedSlots[chain.Base.Name]
				}
				if s == nil {
					baseLt = g.l.llty(chain.Base.Type)
					basePtr = g.alloca(chain.Base.Name, baseLt)
				} else {
					baseLt, basePtr = s.lt, s.ptr
				}
			} else if chain.Global != nil {
				sym := g.globalSymbol(chain.Global)
				baseLt = g.l.llty(chain.Global.Type)
				g.l.ensureGlobal(chain.Global.Name, sym, baseLt)
				basePtr = sym
			} else {
				return lt, "null"
			}
			ptr := basePtr
			for _, step := range chain.Steps {
				cllt := g.l.llty(types.Named{Info: step.Info})
				gp := g.tmp()
				g.emit(gp + " = getelementptr inbounds " + cllt + ", ptr " + ptr + ", i32 0, i32 " +
					strconv.Itoa(step.Index))
				ptr = gp
			}
			return lt, ptr
		}
		if sym, ok := g.l.check.PathSyms[x]; ok && sym != nil {
			s := g.slotFor(sym)
			if s == nil {
				s = g.namedSlots[firstName(x)]
			}
			if s == nil {
				s = &slot{lt: lt, ptr: g.alloca(firstName(x), lt)}
				g.namedSlots[firstName(x)] = s
			}
			return s.lt, s.ptr
		}
		if gi, ok := g.l.check.PathGlobals[x]; ok && gi != nil {
			sym := g.globalSymbol(gi)
			g.l.ensureGlobal(gi.Name, sym, lt)
			return lt, sym
		}
		return lt, "null"
	case *ast.Index:
		ref := g.l.check.IndexRefs[x]
		elemLt := "i32"
		kind := "fixed"
		if ref != nil {
			if ref.Elem != nil {
				elemLt = g.l.llty(ref.Elem)
			}
			kind = ref.Kind
		}
		iv := g.lowerExpr(x.Index_)
		if kind == "view" || kind == "text" {
			baseLt := g.l.viewLT()
			if ref != nil && ref.Base != nil {
				baseLt = g.l.llty(ref.Base)
			}
			base := g.lowerExpr(x.Base)
			p := g.tmp()
			g.emit(p + " = extractvalue " + baseLt + " " + base.s + ", 0")
			n := g.tmp()
			g.emit(n + " = extractvalue " + baseLt + " " + base.s + ", 1")
			// 视图长度恒为 i64：32 位目标上下标是 i32（USize），必须先补宽，
			// 否则 icmp / gep 的操作数宽度对不上（xtensa 上每一处下标都会撞）。
			idx := g.toLen(iv)
			bad := g.tmp()
			g.emit(bad + " = icmp uge " + g.l.lenLT() + " " + idx + ", " + n)
			g.emitCheck(bad, "index out of bounds", x.Span)
			gp := g.tmp()
			g.emit(gp + " = getelementptr inbounds " + elemLt + ", ptr " + p + ", " + g.l.lenLT() + " " + idx)
			return elemLt, gp
		}
		blt, bptr := g.lowerPlace(x.Base)
		gp := g.tmp()
		g.emit(gp + " = getelementptr inbounds " + blt + ", ptr " + bptr + ", " + g.l.lenLT() + " 0, " + g.l.lenLT() + " " + g.toLen(iv))
		return elemLt, gp
	case *ast.FieldAccess:
		blt, bptr := g.lowerPlace(x.Base)
		_ = blt
		ref := g.l.check.FieldRefs[x]
		if ref == nil || ref.Info == nil {
			return g.l.llty(g.ty(x)), bptr
		}
		cllt := g.l.llty(types.Named{Info: ref.Info})
		gp := g.tmp()
		g.emit(gp + " = getelementptr inbounds " + cllt + ", ptr " + bptr + ", i32 0, i32 " +
			strconv.Itoa(ref.Index))
		return g.l.llty(ref.Type), gp
	}
	v := g.lowerExpr(e)
	s := g.alloca("__tmpv", v.lt)
	g.emit("store " + v.lt + " " + v.s + ", ptr " + s)
	return v.lt, s
}

// --------------------------------------------------------------------------- //
// 字面量构造
// --------------------------------------------------------------------------- //

func (g *funcGen) lowerArrayLit(e *ast.ArrayLit, lt string) valueRef {
	ref := g.l.check.ArrRefs[e]
	elemLt := "i32"
	if ref != nil && ref.Elem != nil {
		elemLt = g.l.llty(ref.Elem)
	}
	s := g.alloca("__arr", lt)
	if e.Repeat != nil {
		n := int64(0)
		if ref != nil {
			n = ref.Count
		}
		v := g.lowerExpr(e.Repeat)
		for i := int64(0); i < n; i++ {
			gp := g.tmp()
			g.emit(gp + " = getelementptr inbounds " + lt + ", ptr " + s + ", i64 0, i64 " + strconv.FormatInt(i, 10))
			g.emit("store " + elemLt + " " + v.s + ", ptr " + gp)
		}
	} else {
		for i, item := range e.Elems {
			v := g.lowerExpr(item)
			gp := g.tmp()
			g.emit(gp + " = getelementptr inbounds " + lt + ", ptr " + s + ", i64 0, i64 " + strconv.Itoa(i))
			g.emit("store " + elemLt + " " + v.s + ", ptr " + gp)
		}
	}
	v := g.tmp()
	g.emit(v + " = load " + lt + ", ptr " + s)
	return valueRef{lt, v}
}

func (g *funcGen) lowerStructLit(e *ast.StructLit, lt string) valueRef {
	s := g.alloca("__lit", lt)
	ref := g.l.check.LitRefs[e]
	fields := map[string]ast.Node{}
	for _, pair := range e.FieldPairs {
		name, _ := pair.A.(string)
		if n, ok := pair.B.(ast.Node); ok {
			fields[name] = n
		}
	}
	if ref != nil && ref.Enum != nil {
		tagLt := g.l.llty(ref.Enum.TagType)
		// 无载荷 enum 的标量表示（LLVM 类型 = tag 类型）：值就是 tag 常量本身，
		// 不能走 insertvalue 聚合构造（designs/0004 §3.4 错误成员构造大量命中）
		if lt == tagLt {
			return valueRef{lt, strconv.FormatInt(ref.Variant.Tag, 10)}
		}
		cur := g.tmp()
		g.emit(cur + " = insertvalue " + lt + " undef, " + tagLt + " " + strconv.FormatInt(ref.Variant.Tag, 10) + ", 0")
		if len(ref.Variant.Payload) == 0 {
			return valueRef{lt, cur}
		}
		// 先写 tag 落槽，再逐字段写入载荷（顺序不能反：store 整个聚合会覆盖已写字段）
		g.emit("store " + lt + " " + cur + ", ptr " + s)
		for i, f := range ref.Variant.Payload {
			fexpr := fields[f.Name]
			if fexpr == nil {
				continue
			}
			fv := g.lowerExpr(fexpr)
			gp := g.tmp()
			g.emit(gp + " = getelementptr inbounds " + lt + ", ptr " + s + ", i32 0, i32 1, i32 " +
				strconv.Itoa(i))
			g.emit("store " + fv.lt + " " + fv.s + ", ptr " + gp)
		}
		r := g.tmp()
		g.emit(r + " = load " + lt + ", ptr " + s)
		return valueRef{lt, r}
	}
	if ref != nil && ref.Struct != nil {
		for i, f := range ref.Struct.Fields {
			fexpr := fields[f.Name]
			if fexpr == nil {
				continue
			}
			fv := g.lowerExpr(fexpr)
			gp := g.tmp()
			g.emit(gp + " = getelementptr inbounds " + lt + ", ptr " + s + ", i32 0, i32 " + strconv.Itoa(i))
			g.emit("store " + fv.lt + " " + fv.s + ", ptr " + gp)
		}
	}
	r := g.tmp()
	g.emit(r + " = load " + lt + ", ptr " + s)
	return valueRef{lt, r}
}

// lowerAlloc 生成 `[v; n]@区域名` 的区域分配。
func (g *funcGen) lowerAlloc(e *ast.Alloc) valueRef {
	t := g.ty(e)
	lt := g.l.llty(t)
	ref := g.l.check.AllocRefs[e]
	elemT := types.Type(types.U8)
	if ref != nil && ref.Elem != nil {
		elemT = ref.Elem
	}
	elemLt := g.l.llty(elemT)
	elemSize := g.l.layout.Size(elemT)
	elemAlign := g.l.layout.Align(elemT)

	var reg string
	if e.Region == "static" {
		g.l.ensureDeclare("aila_rt_static_region", "ptr @aila_rt_static_region()")
		reg = g.tmp()
		g.emit(reg + " = call ptr @aila_rt_static_region()")
	} else {
		reg = g.regionVals[e.Region]
		if reg == "" {
			reg = "null"
		}
	}

	countStr := "0"
	var bytesV string
	inner, isArray := e.Expr.(*ast.ArrayLit)
	hasCount := ref != nil && ref.HasCount
	if hasCount {
		countStr = strconv.FormatInt(ref.Count, 10)
		bytesV = strconv.Itoa(elemSize * int(ref.Count))
	} else {
		countVal := "0"
		if isArray && inner.Count != nil {
			cv := g.lowerExpr(inner.Count)
			// 分配尺寸参数恒为 i64：32 位目标上长度是 i32，必须先补齐，
			// 否则会生成 `mul i64 %i32, N` 这种非法 IR（designs/0006 §4.1 第 7 条实测）。
			countVal = g.toI64(cv)
			countStr = cv.s // 视图的 len 字段是「长度宽度」（32 位目标 i32）
		}
		bytesV = g.tmp()
		g.emit(bytesV + " = mul i64 " + countVal + ", " + strconv.Itoa(elemSize))
	}
	g.l.ensureDeclare("aila_rt_region_alloc", "ptr @aila_rt_region_alloc(ptr, i64, i64)")
	p := g.tmp()
	g.emit(p + " = call ptr @aila_rt_region_alloc(ptr " + reg + ", i64 " + bytesV + ", i64 " +
		strconv.Itoa(elemAlign) + ")")

	if isArray {
		if inner.Repeat != nil {
			zlit := false
			if il, ok := inner.Repeat.(*ast.IntLit); ok && il.Raw == 0 {
				zlit = true
			}
			if !zlit {
				rv := g.lowerExpr(inner.Repeat)
				if hasCount {
					for i := int64(0); i < ref.Count; i++ {
						gp := g.tmp()
						g.emit(gp + " = getelementptr inbounds " + elemLt + ", ptr " + p + ", i64 " +
							strconv.FormatInt(i, 10))
						g.emit("store " + elemLt + " " + rv.s + ", ptr " + gp)
					}
				} else {
					iSlot := g.alloca("__fi", "i64")
					g.emit("store i64 0, ptr " + iSlot)
					head := g.newLabel("ai.cond")
					bodyLbl := g.newLabel("ai.body")
					done := g.newLabel("ai.end")
					g.emit("br label %" + head)
					g.emitRaw(head + ":")
					g.terminated = false
					iv := g.tmp()
					g.emit(iv + " = load i64, ptr " + iSlot)
					c := g.tmp()
					g.emit(c + " = icmp ult i64 " + iv + ", " + countStr)
					g.emit("br i1 " + c + ", label %" + bodyLbl + ", label %" + done)
					g.emitRaw(bodyLbl + ":")
					g.terminated = false
					gp := g.tmp()
					g.emit(gp + " = getelementptr inbounds " + elemLt + ", ptr " + p + ", i64 " + iv)
					g.emit("store " + elemLt + " " + rv.s + ", ptr " + gp)
					ni := g.tmp()
					g.emit(ni + " = add i64 " + iv + ", 1")
					g.emit("store i64 " + ni + ", ptr " + iSlot)
					g.emit("br label %" + head)
					g.emitRaw(done + ":")
					g.terminated = false
				}
			}
		} else {
			for i, item := range inner.Elems {
				iv := g.lowerExpr(item)
				gp := g.tmp()
				g.emit(gp + " = getelementptr inbounds " + elemLt + ", ptr " + p + ", " + g.l.lenLT() + " " + strconv.Itoa(i))
				g.emit("store " + iv.lt + " " + iv.s + ", ptr " + gp)
			}
		}
	}
	r := g.tmp()
	g.emit(r + " = insertvalue " + lt + " undef, ptr " + p + ", 0")
	r2 := g.tmp()
	g.emit(r2 + " = insertvalue " + lt + " " + r + ", " + g.l.lenLT() + " " + countStr + ", 1")
	return valueRef{lt, r2}
}

// --------------------------------------------------------------------------- //
// 运算
// --------------------------------------------------------------------------- //

func isCmpOp(op string) bool {
	switch op {
	case "==", "!=", "<", "<=", ">", ">=":
		return true
	}
	return false
}

func signedType(t types.Type) bool { return types.IsSigned(t) }

// lowerBinary 生成二元运算（含短路逻辑与比较）。
func (g *funcGen) lowerBinary(e *ast.Binary) valueRef {
	if e.Op == "&&" || e.Op == "||" {
		return g.lowerShortCircuit(e)
	}
	a := g.lowerExpr(e.Left)
	b := g.lowerExpr(e.Right)
	opT := g.l.check.BinOperands[e]
	if opT == nil {
		opT = g.ty(e.Left)
	}
	olt := g.l.llty(opT)
	signed := signedType(opT)
	if isCmpOp(e.Op) {
		if olt == "float" || olt == "double" {
			instr := map[string]string{"==": "oeq", "!=": "one", "<": "olt", "<=": "ole", ">": "ogt", ">=": "oge"}[e.Op]
			r := g.tmp()
			g.emit(r + " = fcmp " + instr + " " + olt + " " + a.s + ", " + b.s)
			return valueRef{"i1", r}
		}
		instr := map[string]string{
			"==": "eq", "!=": "ne",
			"<":  map[bool]string{true: "slt", false: "ult"}[signed],
			"<=": map[bool]string{true: "sle", false: "ule"}[signed],
			">":  map[bool]string{true: "sgt", false: "ugt"}[signed],
			">=": map[bool]string{true: "sge", false: "uge"}[signed],
		}[e.Op]
		r := g.tmp()
		g.emit(r + " = icmp " + instr + " " + olt + " " + a.s + ", " + b.s)
		return valueRef{"i1", r}
	}
	return valueRef{olt, g.arith(e.Op, olt, a.s, b.s, opT, signed, false, e.Span)}
}

// arith 生成算术 / 位运算；陷阱由编译器插入。
//
// 形态参数：
//   - lt：IR 类型；
//   - opT：AILA 操作数类型（决定有无符号与位宽）；
//   - span：溢出 / 除零陷阱的位置。
func (g *funcGen) arith(op, lt, a, b string, opT types.Type, signed, isAssign bool, span diagSpan) string {
	if lt == "float" || lt == "double" {
		instr := map[string]string{"+": "fadd", "-": "fsub", "*": "fmul", "/": "fdiv"}[op]
		if instr == "" {
			instr = "fadd"
		}
		r := g.tmp()
		g.emit(r + " = " + instr + " " + lt + " " + a + ", " + b)
		return r
	}
	switch op {
	case "+", "-", "*":
		// C 语义回绕（designs/0004 §2.4）：无符号有符号同指令，溢出检查
		// 改由显式 `#runtime.checked_*` 提供
		instr := map[string]string{"+": "add", "-": "sub", "*": "mul"}[op]
		r := g.tmp()
		g.emit(r + " = " + instr + " " + lt + " " + a + ", " + b)
		return r
	case "/", "%":
		g.emitCheck("icmp eq "+lt+" "+b+", 0", "division by zero", span)
		if signed {
			bits := types.IntWidth(opT, g.l.ptrSize)
			if bits > 0 {
				minv := int64(-1) << uint(bits-1)
				g.emitCheck("icmp eq "+lt+" "+a+", "+strconv.FormatInt(minv, 10),
					"integer overflow in division", span)
			}
			instr := "sdiv"
			if op == "%" {
				instr = "srem"
			}
			r := g.tmp()
			g.emit(r + " = " + instr + " " + lt + " " + a + ", " + b)
			return r
		}
		instr := "udiv"
		if op == "%" {
			instr = "urem"
		}
		r := g.tmp()
		g.emit(r + " = " + instr + " " + lt + " " + a + ", " + b)
		return r
	case "<<", ">>":
		width := types.IntWidth(opT, g.l.ptrSize)
		g.emitCheck("icmp uge "+lt+" "+b+", "+strconv.Itoa(width), "shift count out of range", span)
		r := g.tmp()
		if op == "<<" {
			g.emit(r + " = shl " + lt + " " + a + ", " + b)
		} else if signed {
			g.emit(r + " = ashr " + lt + " " + a + ", " + b)
		} else {
			g.emit(r + " = lshr " + lt + " " + a + ", " + b)
		}
		return r
	}
	instr := map[string]string{"&": "and", "|": "or", "^": "xor"}[op]
	if instr == "" {
		instr = "and"
	}
	r := g.tmp()
	g.emit(r + " = " + instr + " " + lt + " " + a + ", " + b)
	return r
}

// lowerShortCircuit 生成 `&&` / `||` 的短路求值。
func (g *funcGen) lowerShortCircuit(e *ast.Binary) valueRef {
	s := g.alloca("__sc", "i1")
	leftLbl := g.newLabel("sc.l")
	rhsLbl := g.newLabel("sc.r")
	endLbl := g.newLabel("sc.end")
	a := g.lowerExpr(e.Left)
	ab := g.toBool(e.Left, a.s)
	if e.Op == "&&" {
		g.emit("br i1 " + ab + ", label %" + rhsLbl + ", label %" + leftLbl)
	} else {
		g.emit("br i1 " + ab + ", label %" + leftLbl + ", label %" + rhsLbl)
	}
	g.emitRaw(leftLbl + ":")
	g.terminated = false
	if e.Op == "&&" {
		g.emit("store i1 0, ptr " + s)
	} else {
		g.emit("store i1 1, ptr " + s)
	}
	g.emit("br label %" + endLbl)
	g.emitRaw(rhsLbl + ":")
	g.terminated = false
	b := g.lowerExpr(e.Right)
	bb := g.toBool(e.Right, b.s)
	g.emit("store i1 " + bb + ", ptr " + s)
	g.emit("br label %" + endLbl)
	g.emitRaw(endLbl + ":")
	g.terminated = false
	r := g.tmp()
	g.emit(r + " = load i1, ptr " + s)
	return valueRef{"i1", r}
}

// lowerUnary 生成一元运算。
func (g *funcGen) lowerUnary(e *ast.Unary, lt string) valueRef {
	v := g.lowerExpr(e.Operand)
	switch e.Op {
	case "!":
		b := g.toBool(e.Operand, v.s)
		r := g.tmp()
		g.emit(r + " = xor i1 " + b + ", true")
		return valueRef{"i1", r}
	case "-":
		if lt == "float" || lt == "double" {
			r := g.tmp()
			g.emit(r + " = fneg " + lt + " " + v.s)
			return valueRef{lt, r}
		}
		r := g.tmp()
		g.emit(r + " = sub " + lt + " 0, " + v.s)
		return valueRef{lt, r}
	case "~":
		r := g.tmp()
		g.emit(r + " = xor " + lt + " " + v.s + ", -1")
		return valueRef{lt, r}
	}
	return valueRef{lt, v.s}
}

// castValue 是显式转换的共享主体（`#compile.as` / `#runtime.as`，
// designs/0004 §2.4）。
func (g *funcGen) castValue(v valueRef, src, dst types.Type) string {
	slt := g.l.llty(src)
	dlt := g.l.llty(dst)
	sb := types.IntWidth(src, g.l.ptrSize)
	db := types.IntWidth(dst, g.l.ptrSize)
	isF := func(s string) bool { return s == "float" || s == "double" }
	isI := func(s string) bool { return strings.HasPrefix(s, "i") }
	r := g.tmp()
	switch {
	case isF(slt) && isF(dlt):
		if db == sb {
			return v.s
		}
		if db > sb {
			g.emit(r + " = fpext " + slt + " " + v.s + " to " + dlt)
		} else {
			g.emit(r + " = fptrunc " + slt + " " + v.s + " to " + dlt)
		}
		return r
	case isF(slt) && isI(dlt):
		g.emit(r + " = fptosi " + slt + " " + v.s + " to " + dlt)
		return r
	case isI(slt) && isF(dlt):
		if signedType(src) {
			g.emit(r + " = sitofp " + slt + " " + v.s + " to " + dlt)
		} else {
			g.emit(r + " = uitofp " + slt + " " + v.s + " to " + dlt)
		}
		return r
	case slt == dlt:
		return v.s
	case isI(slt) && isI(dlt):
		// 整数↔整数一律按 **LLVM 类型宽度** 决定指令（不能只看 AILA 类型：
		// 字面量/未定型表达式可能带不出 IntWidth，那时按 AILA 宽度会挑错指令，
		// 生成 `zext i64 … to i32` 这种非法 IR）。
		if intWidthOf(dlt) > intWidthOf(slt) {
			if signedType(src) {
				g.emit(r + " = sext " + slt + " " + v.s + " to " + dlt)
			} else {
				g.emit(r + " = zext " + slt + " " + v.s + " to " + dlt)
			}
		} else {
			g.emit(r + " = trunc " + slt + " " + v.s + " to " + dlt)
		}
		return r
	case db > sb:
		if signedType(src) {
			g.emit(r + " = sext " + slt + " " + v.s + " to " + dlt)
		} else {
			g.emit(r + " = zext " + slt + " " + v.s + " to " + dlt)
		}
		return r
	}
	g.emit(r + " = trunc " + slt + " " + v.s + " to " + dlt)
	return r
}

// lowerLayerCall 生成 `#层.API(...)`（designs/0004 §2.4 v0 清单）。
// 静态解释在 sema 的 LayerRef；这里按 Kind 分派 lowering。
func (g *funcGen) lowerLayerCall(x *ast.LayerCall, lt string) valueRef {
	ref := g.l.check.LayerRefs[x]
	if ref == nil {
		return valueRef{lt, g.zero(g.ty(x))}
	}
	switch ref.Kind {
	case "sizeof":
		return valueRef{g.l.lenLT(), strconv.Itoa(g.l.layout.Size(ref.Elem))}
	case "alignof":
		return valueRef{g.l.lenLT(), strconv.Itoa(g.l.layout.Align(ref.Elem))}
	case "as":
		v := g.lowerExpr(x.Args[1])
		return valueRef{g.l.llty(ref.Elem), g.castValue(v, ref.Src, ref.Elem)}
	case "len":
		if f, ok := ref.At.(types.Fixed); ok {
			return valueRef{g.l.lenLT(), strconv.FormatInt(f.N, 10)}
		}
		vlt := g.l.llty(ref.At)
		v := g.lowerExpr(x.Args[0])
		r := g.tmp()
		g.emit(r + " = extractvalue " + vlt + " " + v.s + ", 1")
		return valueRef{g.l.lenLT(), r}
	case "trap":
		if s, ok := x.Args[0].(*ast.StrLit); ok {
			g.trapNow(s.Value, x.Span)
		} else {
			g.trapNow("trap", x.Span)
		}
		return valueRef{lt, "undef"}
	case "atomic":
		return g.lowerAtomic(ref, x)
	case "checked":
		return g.lowerChecked(ref, x)
	case "chan":
		return g.lowerLayerChan(ref, x)
	case "coro":
		// 帧区域取词法最内层 region（区域推断，designs/0004 §2.4）
		freg := g.regionVals[g.currentRegion]
		if freg == "" {
			freg = "null"
		}
		return valueRef{lt, g.coroCreateInto(ref.Sig, x.Args[1:], nil, freg)}
	case "spawn":
		g.threadStartInto(ref.Sig, x.Args[1:], "0", x.Span)
		return valueRef{lt, "undef"}
	}
	g.unsupported("层 API `" + ref.Kind + "` 的代码生成待批次")
	return valueRef{lt, g.zero(g.ty(x))}
}

// lowerAtomic 生成 `#runtime.atomic_*`（designs/0001 §10.10 / 0004 §2.4）：
// (block ptr, off i64, 其余按宽度类型) → aila_rt_atomic_*（C 侧 seq_cst + 对齐检查）。
func (g *funcGen) lowerAtomic(ref *sema.LayerRef, x *ast.LayerCall) valueRef {
	at := ref.At
	alt := g.l.llty(at)
	out := make([]string, 0, len(x.Args))
	bv := g.lowerExpr(x.Args[0])
	out = append(out, "ptr "+g.viewPtr(x.Args[0], bv.s, bv.lt))
	ov := g.lowerExpr(x.Args[1])
	if ov.lt != "i64" && isIntLL(ov.lt) {
		ov.s = g.coerceInt(ov.lt, "i64", ov.s)
	}
	out = append(out, "i64 "+ov.s)
	for _, a := range x.Args[2:] {
		av := g.lowerExpr(a)
		if av.lt != alt && isIntLL(alt) && (isIntLL(av.lt) || av.lt == "i1") {
			av.s = g.coerceInt(av.lt, alt, av.s)
		}
		out = append(out, alt+" "+av.s)
	}
	isStore := strings.Contains(ref.AtomicFn, "_store")
	retLt := alt
	if isStore {
		retLt = "void"
	}
	// declare 形态与 AILA#[c] 签名一致：(block ptr, off i64, 宽度参数…)
	params := "ptr, i64"
	for i := 2; i < len(x.Args); i++ {
		params += ", " + alt
	}
	g.l.ensureDeclare(ref.AtomicFn, retLt+" @"+ref.AtomicFn+"("+params+")")
	if isStore {
		g.emit("call void @" + ref.AtomicFn + "(" + strings.Join(out, ", ") + ")")
		return valueRef{"void", "undef"}
	}
	r := g.tmp()
	g.emit(r + " = call " + alt + " @" + ref.AtomicFn + "(" + strings.Join(out, ", ") + ")")
	return valueRef{alt, r}
}

// lowerChecked 生成 `#runtime.checked_{add,sub,mul}`：llvm.*.with.overflow
// 探测溢出，产出 (T, Bool) 多值组——溢出标志不 trap，由调用侧决定处置
// （designs/0004 §2.4）。
func (g *funcGen) lowerChecked(ref *sema.LayerRef, x *ast.LayerCall) valueRef {
	lt := g.l.llty(ref.At)
	s := "u"
	if signedType(ref.At) {
		s = "s"
	}
	a := g.lowerExpr(x.Args[0])
	b := g.lowerExpr(x.Args[1])
	pair := g.tmp()
	g.emit(pair + " = call { " + lt + ", i1 } @llvm." + s + ref.CheckedOp + ".with.overflow." + lt +
		"(" + lt + " " + a.s + ", " + lt + " " + b.s + ")")
	v := g.tmp()
	g.emit(v + " = extractvalue { " + lt + ", i1 } " + pair + ", 0")
	ovf := g.tmp()
	g.emit(ovf + " = extractvalue { " + lt + ", i1 } " + pair + ", 1")
	r := g.tmp()
	g.emit(r + " = insertvalue { " + lt + ", i1 } undef, " + lt + " " + v + ", 0")
	r2 := g.tmp()
	g.emit(r2 + " = insertvalue { " + lt + ", i1 } " + r + ", i1 " + ovf + ", 1")
	return valueRef{"{ " + lt + ", i1 }", r2}
}

// lowerLayerChan 生成 `#runtime.chan(T, cap)`：与 `chan@r T(cap)` 同一 lowering，
// 区域取词法最内层 region（区域推断）。
func (g *funcGen) lowerLayerChan(ref *sema.LayerRef, x *ast.LayerCall) valueRef {
	esize := g.l.layout.Size(ref.Elem)
	reg := g.regionVals[g.currentRegion]
	if reg == "" {
		reg = "null"
	}
	capV := g.lowerExpr(x.Args[1])
	if capV.lt != "i64" && isIntLL(capV.lt) {
		capV.s = g.coerceInt(capV.lt, "i64", capV.s)
	}
	g.l.ensureDeclare("aila_rt_chan_create", "ptr @aila_rt_chan_create(ptr, i64, i64)")
	p := g.tmp()
	g.emit(p + " = call ptr @aila_rt_chan_create(ptr " + reg + ", i64 " + capV.s +
		", i64 " + strconv.Itoa(esize) + ")")
	agg := g.tmp()
	g.emit(agg + " = insertvalue { ptr, ptr } undef, ptr " + p + ", 0")
	agg2 := g.tmp()
	g.emit(agg2 + " = insertvalue { ptr, ptr } " + agg + ", ptr " + p + ", 1")
	return valueRef{"{ ptr, ptr }", agg2}
}

// --------------------------------------------------------------------------- //
// 调用
// --------------------------------------------------------------------------- //

// lowerCall 生成调用（普通 / 方法 / 外部 / 函数指针）。
func (g *funcGen) lowerCall(expr ast.Node, callee ast.Node, args []ast.Node, regionArgs []string, lt string) valueRef {
	sig := g.l.check.CallSigs[expr]
	if sig == nil {
		return valueRef{lt, g.zero(g.ty(expr))}
	}
	kind := g.l.check.CallKinds[expr]
	out := make([]string, 0, len(args)+1+len(regionArgs))
	for _, rn := range regionArgs {
		reg := g.regionVals[rn]
		if reg == "" {
			reg = "null"
		}
		out = append(out, "ptr "+reg)
	}
	recv := g.l.check.CallRecv[expr]
	if recv != nil {
		if mutsSelf(sig) {
			// @mut[self]：主名字须是 var 槽（sema 已查），传槽位地址实现就地写。
			_, ptr := g.lowerPlace(recv)
			out = append(out, "ptr "+ptr)
		} else if len(sig.Params) > 0 && g.l.paramByRef(sig, 0, sig.Params[0].Type) {
			// 大聚合体接收者按只读指针传（AX-9），与普通实参同款。
			if _, ptr := g.tryPlacePtr(recv); ptr != "" {
				out = append(out, "ptr "+ptr)
			} else {
				rv := g.lowerExpr(recv)
				sp := g.alloca("recv", rv.lt)
				g.emit("store " + rv.lt + " " + rv.s + ", ptr " + sp)
				out = append(out, "ptr "+sp)
			}
		} else {
			rv := g.lowerExpr(recv)
			rplt := rv.lt
			if len(sig.Params) > 0 {
				rplt = g.l.paramLLType(sig, sig.Params[0].Type)
			}
			if rplt == "ptr" && rv.lt != "ptr" {
				rv.s = g.viewPtr(recv, rv.s, rv.lt)
			}
			out = append(out, rplt+" "+rv.s)
		}
	}
	rest := sig.Params
	if len(rest) > 0 && rest[0].Name == "self" {
		rest = rest[1:]
	}
	// AILA 原生变参（designs/0002 §6）：末位 rest 是视图参数，多出的实参
	// 在调用点栈上打包成 [N x T] 再传 {ptr, N} 视图——零堆分配、零 region。
	nativeVariadic := sig.VariadicElem != nil && len(rest) > 0
	fixedN := len(rest)
	elemLt := ""
	if nativeVariadic {
		fixedN--
		elemLt = g.l.llty(sig.VariadicElem)
	}
	var pack []string
	for i, arg := range args {
		byRef := i < fixedN && g.l.paramByRef(sig, paramIndexFor(sig, i), rest[i].Type)
		if byRef {
			// 大聚合体按只读指针传（AX-9）：实参是主名字就传槽位地址，
			// 否则先落到临时槽再传地址。被调方只读，不涉及写回。
			if _, ptr := g.tryPlacePtr(arg); ptr != "" {
				out = append(out, "ptr "+ptr)
			} else {
				av := g.lowerExpr(arg)
				sp := g.alloca("byref", av.lt)
				g.emit("store " + av.lt + " " + av.s + ", ptr " + sp)
				out = append(out, "ptr "+sp)
			}
			continue
		}
		av := g.lowerExpr(arg)
		var plt string
		if i < fixedN {
			plt = g.l.paramLLType(sig, rest[i].Type)
		} else if nativeVariadic {
			plt = elemLt
		} else {
			plt = av.lt
		}
		if plt == "ptr" && av.lt != "ptr" {
			av.s = g.viewPtr(arg, av.s, av.lt)
		} else if plt != av.lt && isIntLL(plt) && (isIntLL(av.lt) || av.lt == "i1") {
			av.s = g.coerceInt(av.lt, plt, av.s)
		}
		if nativeVariadic && i >= fixedN {
			pack = append(pack, plt+" "+av.s)
			continue
		}
		out = append(out, plt+" "+av.s)
	}
	if nativeVariadic {
		restLt := g.l.paramLLType(sig, rest[len(rest)-1].Type)
		if len(pack) == 0 {
			out = append(out, restLt+" zeroinitializer")
		} else {
			arr := g.alloca("vargs", "["+strconv.Itoa(len(pack))+" x "+elemLt+"]")
			for i, pv := range pack {
				gp := g.tmp()
				g.emit(gp + " = getelementptr inbounds " + elemLt + ", ptr " + arr + ", i64 " + strconv.Itoa(i))
				g.emit("store " + pv + ", ptr " + gp)
			}
			vr := g.tmp()
			g.emit(vr + " = insertvalue " + restLt + " undef, ptr " + arr + ", 0")
			vr2 := g.tmp()
			g.emit(vr2 + " = insertvalue " + restLt + " " + vr + ", " + g.l.lenLT() + " " + strconv.Itoa(len(pack)) + ", 1")
			out = append(out, restLt+" "+vr2)
		}
	}
	sym := g.l.symbolOf(sig)
	retLt, params := g.l.funcLLType(sig)
	calleeText := "@" + sym
	if kind == "fnptr" {
		// 间接调用：被调者来自「函数指针值」表达式——字段形态由 sema 合成
		// （CallCallee），一般表达式形态取 Call 的 Callee。不产生 declare。
		callee = g.l.check.CallCallee[expr]
		if callee == nil {
			if x, ok := expr.(*ast.Call); ok {
				callee = x.Callee
			}
		}
		if callee != nil {
			cv := g.lowerExpr(callee)
			calleeText = cv.s
		}
	} else {
		g.l.ensureDeclare(sym, retLt+" @"+sym+"("+strings.Join(params, ", ")+")")
	}
	// C 变参调用：LLVM 文本 IR 要求带完整函数类型注解（`i32 (ptr, ...)`），
	// 注解本身已含返回类型，直接取代 `retLt @sym` 片段。
	callTy := ""
	if sig.Variadic && kind != "fnptr" {
		callTy = retLt + " (" + strings.Join(params, ", ") + ")"
	}
	if retLt == "void" {
		g.emit("call " + callTySep(retLt, callTy, calleeText) + "(" + strings.Join(out, ", ") + ")")
		return valueRef{"void", "undef"}
	}
	r := g.tmp()
	g.emit(r + " = call " + callTySep(retLt, callTy, calleeText) + "(" + strings.Join(out, ", ") + ")")
	return valueRef{retLt, r}
}

// tryPlacePtr 试着把实参当「主名字地址」取（AX-9 的按引用传参用）：
// 只有 Path（变量 / 字段链）够格；不是主名字或取不到地址时返回空串，
// 调用方自行落一个临时槽。
func (g *funcGen) tryPlacePtr(e ast.Node) (string, string) {
	if _, ok := e.(*ast.Path); !ok {
		return "", ""
	}
	lt, ptr := g.lowerPlace(e)
	if ptr == "" || ptr == "null" {
		return "", ""
	}
	return lt, ptr
}

// paramIndexFor 把「去掉 self 之后的实参下标」换算成签名里的参数下标。
func paramIndexFor(sig *sema.FuncSig, i int) int {
	if len(sig.Params) > 0 && sig.Params[0].Name == "self" {
		return i + 1
	}
	return i
}

// callTySep 组装 call 指令的被调者片段：有函数类型注解时 `ret (params) @sym`，
// 否则 `ret @sym`。
func callTySep(retLt, callTy, calleeText string) string {
	if callTy == "" {
		return retLt + " " + calleeText
	}
	return callTy + " " + calleeText
}

func isIntLL(s string) bool {
	switch s {
	case "i1", "i8", "i16", "i32", "i64":
		return true
	}
	return false
}

// viewPtr 取视图的指针分量（C 边界只要 ptr）。
func (g *funcGen) viewPtr(expr ast.Node, val, lt string) string {
	t := g.ty(expr)
	if o, ok := t.(types.Optional); ok {
		t = o.Inner
	}
	if types.IsViewFamily(t) {
		p := g.tmp()
		g.emit(p + " = extractvalue " + lt + " " + val + ", 0")
		return p
	}
	return "null"
}

// coerceInt 做整数位宽适配。
func (g *funcGen) coerceInt(src, dst, val string) string {
	sb := intWidthOf(src)
	db := intWidthOf(dst)
	if sb == db || sb == 0 || db == 0 {
		return val
	}
	r := g.tmp()
	if db > sb {
		g.emit(r + " = zext " + src + " " + val + " to " + dst)
	} else {
		g.emit(r + " = trunc " + src + " " + val + " to " + dst)
	}
	return r
}

func intWidthOf(lt string) int {
	if !strings.HasPrefix(lt, "i") {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimPrefix(lt, "i"))
	if err != nil {
		return 0
	}
	return n
}

// intLiteral 生成整数字面量（按目标位宽截断）。
func intLiteral(v int64, lt string) string {
	bits := intWidthOf(lt)
	if bits <= 0 || bits >= 64 {
		return strconv.FormatInt(v, 10)
	}
	mask := int64(1)<<uint(bits) - 1
	return strconv.FormatInt(v&mask, 10)
}

// lowerTryPropagate 生成 `?` 上抛。
func (g *funcGen) lowerTryPropagate(e *ast.TryPropagate, lt string) valueRef {
	ref := g.l.check.TryRefs[e]
	mode := "result"
	if ref != nil && ref.Mode != "" {
		mode = ref.Mode
	}
	innerT := g.ty(e.Operand)
	ltIn := g.l.llty(innerT)
	v := g.lowerExpr(e.Operand)
	s := g.alloca("__try", ltIn)
	g.emit("store " + ltIn + " " + v.s + ", ptr " + s)
	okLbl := g.newLabel("try.ok")
	errLbl := g.newLabel("try.err")
	if mode == "result" {
		flag := g.tmp()
		g.emit(flag + " = extractvalue " + ltIn + " " + v.s + ", 0")
		g.emit("br i1 " + flag + ", label %" + okLbl + ", label %" + errLbl)
	} else {
		opt, _ := innerT.(types.Optional)
		if g.l.isNicheOptional(opt) {
			plt := g.l.llty(opt.Inner)
			pv := g.tmp()
			g.emit(pv + " = extractvalue " + plt + " " + v.s + ", 0")
			isnull := g.tmp()
			g.emit(isnull + " = icmp eq ptr " + pv + ", null")
			g.emit("br i1 " + isnull + ", label %" + errLbl + ", label %" + okLbl)
		} else {
			flag := g.tmp()
			g.emit(flag + " = extractvalue " + ltIn + " " + v.s + ", 0")
			g.emit("br i1 " + flag + ", label %" + okLbl + ", label %" + errLbl)
		}
	}
	// 错误路径：原样返回
	g.emitRaw(errLbl + ":")
	g.terminated = false
	retT := g.l.sigRetType(g.sig)
	retLt := g.l.llty(retT)
	var baseT types.Type
	if g.sig.RetGroup != nil {
		baseT = types.Group{Items: g.sig.RetGroup}
	} else {
		baseT = g.sig.Ret
	}
	if g.isCoro {
		// 协程内 `?` 上抛：Err 写控制块 err 槽（state=2）；可空通道无值完成
		// （裸 coro 的 Done，state=1）。成功路径照旧 extractvalue。
		if mode == "result" {
			ev := g.tmp()
			g.emit(ev + " = extractvalue " + ltIn + " " + v.s + ", 2")
			g.coroFinish(2, nil, &ev)
		} else {
			g.coroFinish(1, nil, nil)
		}
		g.emitRaw(okLbl + ":")
		g.terminated = false
		outT := types.Type(types.I32)
		if ref != nil && ref.Ok != nil {
			outT = ref.Ok
		}
		olt := g.l.llty(outT)
		if mode == "result" {
			r := g.tmp()
			g.emit(r + " = extractvalue " + ltIn + " " + v.s + ", 1")
			return valueRef{olt, r}
		}
		opt, _ := innerT.(types.Optional)
		if g.l.isNicheOptional(opt) {
			return valueRef{olt, v.s}
		}
		r := g.tmp()
		g.emit(r + " = extractvalue " + ltIn + " " + v.s + ", 1")
		return valueRef{olt, r}
	}
	if mode == "result" {
		ev := g.tmp()
		g.emit(ev + " = extractvalue " + ltIn + " " + v.s + ", 2")
		if ref != nil && ref.Err != nil && ref.Err != g.sig.Err {
			ev = g.coerceErr(ev, ref.Err, g.sig.Err)
		}
		rval := g.buildResult(retLt, baseT, g.sig.Err, false, ev)
		g.flushDefers()
		g.emit("ret " + retLt + " " + rval)
	} else {
		g.flushDefers()
		g.emit("ret " + retLt + " " + g.zero(retT))
	}
	g.terminated = true
	// 成功路径
	g.emitRaw(okLbl + ":")
	g.terminated = false
	outT := types.Type(types.I32)
	if ref != nil && ref.Ok != nil {
		outT = ref.Ok
	}
	olt := g.l.llty(outT)
	if mode == "result" {
		r := g.tmp()
		g.emit(r + " = extractvalue " + ltIn + " " + v.s + ", 1")
		return valueRef{olt, r}
	}
	opt, _ := innerT.(types.Optional)
	if g.l.isNicheOptional(opt) {
		return valueRef{olt, v.s}
	}
	r := g.tmp()
	g.emit(r + " = extractvalue " + ltIn + " " + v.s + ", 1")
	return valueRef{olt, r}
}

// innerResultErr 取操作数的错误通道类型（result 模式专用）。
func innerResultErr(t types.Type) types.Type {
	if r, ok := t.(types.Result); ok {
		return r.Err
	}
	return nil
}

// coerceErr 把上抛的错误值从操作数错误通道转成函数签名的错误通道
// （designs/0004 §3.2）。错误成员 tag 已全局统一编号
// （types.RegisterErrorShard），转换只需宽度对齐 + 聚合重装；
// 带载荷成员按 tag 线性分发拷贝载荷。
func (g *funcGen) coerceErr(v string, src, dst types.Type) string {
	if src == nil || dst == nil {
		return v
	}
	srcLt, dstLt := g.l.llty(src), g.l.llty(dst)
	if srcLt == dstLt {
		return v
	}
	es, ok := dst.(types.ErrorSpace)
	if !ok {
		return v // v0 唯一跨错误集上抛方向：具体错误集 → Error 超集
	}
	srcEn := types.EnumInfoOf(src)
	dstTagLt := g.l.llty(es.Info.TagType)
	// 1) tag 宽度对齐（tag 无符号：截断 / 零扩展）
	tag := v
	if srcLt != dstTagLt {
		r := g.tmp()
		if ltBits(srcLt) > ltBits(dstTagLt) {
			g.emit(r + " = trunc " + srcLt + " " + v + " to " + dstTagLt)
		} else {
			g.emit(r + " = zext " + srcLt + " " + v + " to " + dstTagLt)
		}
		tag = r
	}
	// 2) 目标标量（全局空间无载荷成员时）：tag 就是值
	if dstLt == dstTagLt {
		return tag
	}
	// 3) 目标聚合 { tag, union }：tag 装 0 号槽
	cur := g.tmp()
	g.emit(cur + " = insertvalue " + dstLt + " undef, " + dstTagLt + " " + tag + ", 0")
	if srcEn == nil || !srcEn.Payload {
		return cur // 源无载荷成员：载荷槽保持 undef
	}
	// 4) 源有载荷成员：按 tag 逐变体分发拷贝（tag 全局唯一，至多命中一个）
	slot := g.alloca("__errcoerce", dstLt)
	g.emit("store " + dstLt + " " + cur + ", ptr " + slot)
	for _, sv := range srcEn.Variants {
		if len(sv.Payload) == 0 {
			continue
		}
		hit := g.newLabel("coerce.hit")
		next := g.newLabel("coerce.next")
		cmp := g.tmp()
		g.emit(cmp + " = icmp eq " + dstTagLt + " " + tag + ", " + strconv.FormatInt(sv.Tag, 10))
		g.emit("br i1 " + cmp + ", label %" + hit + ", label %" + next)
		g.emitRaw(hit + ":")
		g.terminated = false
		cv := g.tmp()
		g.emit(cv + " = load " + dstLt + ", ptr " + slot)
		for i := range sv.Payload {
			fv := g.tmp()
			g.emit(fv + " = extractvalue " + srcLt + " " + v + ", 1, " + strconv.Itoa(i))
			cv2 := g.tmp()
			g.emit(cv2 + " = insertvalue " + dstLt + " " + cv + ", " + g.l.llty(sv.Payload[i].Type) + " " + fv + ", 1, " + strconv.Itoa(i))
			cv = cv2
		}
		g.emit("store " + dstLt + " " + cv + ", ptr " + slot)
		g.emit("br label %" + next)
		g.emitRaw(next + ":")
		g.terminated = false
	}
	out := g.tmp()
	g.emit(out + " = load " + dstLt + ", ptr " + slot)
	return out
}

// ltBits 从 LLVM 整型拼写取位宽（"i8" → 8；非整型返回 0）。
func ltBits(lt string) int {
	if len(lt) < 2 || lt[0] != 'i' {
		return 0
	}
	n, err := strconv.Atoi(lt[1:])
	if err != nil {
		return 0
	}
	return n
}

// --------------------------------------------------------------------------- //
// 协程启动与恢复（designs/0001 §10.2 / §10.5）
// --------------------------------------------------------------------------- //

// coroCreateInto 是协程启动的共享主体（`#runtime.coro`，designs/0004 §2.4）：
// 分配控制块 → 打包实参 → 调协程函数（eager 到首挂起点）。
func (g *funcGen) coroCreateInto(sig *sema.FuncSig, args []ast.Node, regionArgs []string, freg string) string {
	_, _, cbsize := g.l.coroLayoutOf(sig)
	g.l.ensureDeclare("aila_rt_region_alloc", "ptr @aila_rt_region_alloc(ptr, i64, i64)")
	cb := g.tmp()
	g.emit(cb + " = call ptr @aila_rt_region_alloc(ptr " + freg + ", i64 " +
		strconv.Itoa(cbsize) + ", i64 8)")

	// 实参：区域实参在前（与普通调用同一通道），随后普通实参、隐藏尾参 (cb, 帧区域)
	out := make([]string, 0, len(regionArgs)+len(args)+2)
	for _, rn := range regionArgs {
		reg := g.regionVals[rn]
		if reg == "" {
			reg = "null"
		}
		out = append(out, "ptr "+reg)
	}
	params := sig.Params
	for i, arg := range args {
		av := g.lowerExpr(arg)
		plt := av.lt
		if i < len(params) {
			plt = g.l.paramLLType(sig, params[i].Type)
		}
		if plt == "ptr" && av.lt != "ptr" {
			av.s = g.viewPtr(arg, av.s, av.lt)
		} else if plt != av.lt && isIntLL(plt) && (isIntLL(av.lt) || av.lt == "i1") {
			av.s = g.coerceInt(av.lt, plt, av.s)
		}
		out = append(out, plt+" "+av.s)
	}
	out = append(out, "ptr "+cb, "ptr "+freg)

	sym := g.l.symbolOf(sig)
	_, pl := g.l.funcLLType(sig)
	// 协程 IR 形态：返回恒为句柄 ptr，多两个隐藏尾参
	g.l.ensureDeclare(sym, "ptr @"+sym+"("+strings.Join(append(append([]string{}, pl...), "ptr", "ptr"), ", ")+")")
	hdl := g.tmp()
	g.emit(hdl + " = call ptr @" + sym + "(" + strings.Join(out, ", ") + ")")
	return hdl
}

// coroResume 是 resume 的共享主体（select 的 coro 等完成事件用，designs/0004 §6）：
// state=0 → 恢复至下一挂起点或完成；state=1/2 → 首次观察完成态（报告后标记
// 3=已消费）；state=3 → 重复 resume 已观察的完成句柄 = trap。
// 判别和分派必须用恢复后的新状态（eager 完成发生在 resume 期间）。
//
// 返回值是编译器内部三态判别和（`types.CoroState`）：state 0 = 仍然挂起，
// 1 = 完成（Ok|Done），2 = 错误。它只在本文件与 lowerSelect 内部流动，
// 不出现在用户类型里（`resume` 的值形态已退役，sema E5010）。
func (g *funcGen) coroResume(co types.CoroState, hv valueRef, lt string, span diag.Span) string {
	sum := g.alloca("__rst", lt)
	st := g.tmp()
	g.emit(st + " = load i8, ptr " + hv.s)
	alive := g.tmp()
	g.emit(alive + " = icmp eq i8 " + st + ", 0")
	runLbl := g.newLabel("resume.run")
	doneChk := g.newLabel("resume.donechk")
	g.emit("br i1 " + alive + ", label %" + runLbl + ", label %" + doneChk)

	// ---- run 路径：state==0，正常恢复 ----
	g.emitRaw(runLbl + ":")
	g.terminated = false
	fp := g.tmp()
	g.emit(fp + " = getelementptr inbounds i8, ptr " + hv.s + ", i64 8")
	frame := g.tmp()
	g.emit(frame + " = load ptr, ptr " + fp)
	g.l.ensureDeclare("llvm.coro.resume", "void @llvm.coro.resume(ptr)")
	g.emit("call void @llvm.coro.resume(ptr " + frame + ")")
	st2 := g.tmp()
	g.emit(st2 + " = load i8, ptr " + hv.s) // 恢复后的新状态（挂起=0 / 完成=1|2）
	g.lowerResumeSum(co, lt, sum, hv.s, st2)
	needMark := g.tmp()
	g.emit(needMark + " = icmp ne i8 " + st2 + ", 0")
	markLbl := g.newLabel("resume.mark")
	joinLbl := g.newLabel("resume.join")
	g.emit("br i1 " + needMark + ", label %" + markLbl + ", label %" + joinLbl)
	g.emitRaw(markLbl + ":")
	g.terminated = false
	g.emit("store i8 3, ptr " + hv.s) // 完成态被观察过一次 → 已消费
	g.emit("br label %" + joinLbl)

	// ---- done 路径：state==1|2，首次观察完成态 ----
	g.emitRaw(doneChk + ":")
	g.terminated = false
	dead := g.tmp()
	g.emit(dead + " = icmp eq i8 " + st + ", 3")
	trapLbl := g.newLabel("resume.trap")
	doneLbl := g.newLabel("resume.done")
	g.emit("br i1 " + dead + ", label %" + trapLbl + ", label %" + doneLbl)
	g.emitRaw(trapLbl + ":")
	g.terminated = false
	g.trapNow("resume 已观察过完成态的协程（状态机违规）", span)
	g.emitRaw(doneLbl + ":")
	g.terminated = false
	g.lowerResumeSum(co, lt, sum, hv.s, st)
	g.emit("store i8 3, ptr " + hv.s)
	g.emit("br label %" + joinLbl)

	g.emitRaw(joinLbl + ":")
	g.terminated = false
	v := g.tmp()
	g.emit(v + " = load " + lt + ", ptr " + sum)
	return v
}

// lowerResumeSum 把 { state, Ok 载荷, Err 载荷 } 构造进判别和槽。
// Ok/Err 载荷在控制块 out/err 槽（布局与被调侧同一份算法，Ok/Err 同源）。
func (g *funcGen) lowerResumeSum(co types.CoroState, lt, sum, cb, st string) {
	if co.Ok == nil {
		// 二态 { Suspended, Done }
		r := g.tmp()
		g.emit(r + " = insertvalue " + lt + " undef, i8 " + st + ", 0")
		g.emit("store " + lt + " " + r + ", ptr " + sum)
		return
	}
	r0 := g.tmp()
	g.emit(r0 + " = insertvalue " + lt + " undef, i8 " + st + ", 0")
	offOut, offErr, _ := g.l.coroStateLayout(co.Ok, co.Err)
	op := g.tmp()
	g.emit(op + " = getelementptr inbounds i8, ptr " + cb + ", i64 " + strconv.Itoa(offOut))
	ov := g.tmp()
	g.emit(ov + " = load " + g.l.llty(co.Ok) + ", ptr " + op)
	r1 := g.tmp()
	g.emit(r1 + " = insertvalue " + lt + " " + r0 + ", " + g.l.llty(co.Ok) + " " + ov + ", 1")
	if co.Err != nil {
		ep := g.tmp()
		g.emit(ep + " = getelementptr inbounds i8, ptr " + cb + ", i64 " + strconv.Itoa(offErr))
		ev := g.tmp()
		g.emit(ev + " = load " + g.l.llty(co.Err) + ", ptr " + ep)
		r2 := g.tmp()
		g.emit(r2 + " = insertvalue " + lt + " " + r1 + ", " + g.l.llty(co.Err) + " " + ev + ", 2")
		g.emit("store " + lt + " " + r2 + ", ptr " + sum)
		return
	}
	g.emit("store " + lt + " " + r1 + ", ptr " + sum)
}

// --------------------------------------------------------------------------- //
// 通道（designs/0001 §10.6）
// --------------------------------------------------------------------------- //

// lowerRecv 生成 `<-ch` 接收表达式：try 成功取值；失败挂起（coro.suspend），
// 恢复后重试。仅协程内合法（sema E5001 已保证 isCoro）。
func (g *funcGen) lowerRecv(x *ast.Recv) valueRef {
	ch := g.lowerExpr(x.Chan)
	elemT := g.ty(x)
	elemLt := g.l.llty(elemT)
	out := g.alloca("__recv", elemLt)
	g.l.ensureDeclare("aila_rt_chan_try_recv", "i1 @aila_rt_chan_try_recv(ptr, ptr)")
	retry := g.newLabel("chan.recv")
	g.emit("br label %" + retry)
	g.emitRaw(retry + ":")
	ok := g.tmp()
	g.emit(ok + " = call i1 @aila_rt_chan_try_recv(ptr " + ch.s + ", ptr " + out + ")")
	have := g.newLabel("chan.have")
	wait := g.newLabel("chan.wait")
	g.emit("br i1 " + ok + ", label %" + have + ", label %" + wait)
	g.emitRaw(wait + ":")
	g.terminated = false
	g.coroSuspendRetry(retry)
	g.emitRaw(have + ":")
	g.terminated = false
	v := g.tmp()
	g.emit(v + " = load " + elemLt + ", ptr " + out)
	return valueRef{elemLt, v}
}

// toLen 把整数操作数补到「长度宽度」（lenLT：32 位目标 i32 / 64 位 i64）。
// 视图的 len 字段就是这个宽度，下标参与 icmp / gep 前必须先对齐，
// 否则 32 位目标上必定出现 `icmp i64 %i32, %i32` 这种非法 IR。
func (g *funcGen) toLen(v valueRef) string {
	dst := g.l.lenLT()
	if v.lt == dst {
		return v.s
	}
	if !isIntLL(v.lt) || !isIntLL(dst) {
		return v.s
	}
	t := g.tmp()
	if intWidthOf(dst) > intWidthOf(v.lt) {
		g.emit(t + " = zext " + v.lt + " " + v.s + " to " + dst)
	} else {
		g.emit(t + " = trunc " + v.lt + " " + v.s + " to " + dst)
	}
	return t
}

// toI64 把整数操作数补到 i64——`aila_rt_region_alloc` 的两条尺寸参数恒为 i64，
// 32 位目标上长度是 i32，不补齐就出非法 IR（`mul i64 %i32, N`）。
func (g *funcGen) toI64(v valueRef) string {
	const dst = "i64"
	if v.lt == dst || !isIntLL(v.lt) {
		return v.s
	}
	t := g.tmp()
	if intWidthOf(dst) > intWidthOf(v.lt) {
		g.emit(t + " = zext " + v.lt + " " + v.s + " to " + dst)
	} else {
		g.emit(t + " = trunc " + v.lt + " " + v.s + " to " + dst)
	}
	return t
}

