package codegen

import (
	"strconv"
	"strings"

	"github.com/aila-lang/aila/internal/ast"
	"github.com/aila-lang/aila/internal/sema"
	"github.com/aila-lang/aila/internal/types"
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
		g.emit(r + " = insertvalue { ptr, i64 } undef, ptr " + sym + ", 0")
		r2 := g.tmp()
		g.emit(r2 + " = insertvalue { ptr, i64 } " + r + ", i64 " + strconv.Itoa(len(x.Value)) + ", 1")
		return valueRef{"{ ptr, i64 }", r2}
	case *ast.CharLit:
		return valueRef{lt, strconv.FormatInt(runeValue(x.Value), 10)}
	case *ast.Path:
		return g.lowerPath(x)
	case *ast.GroupExpr:
		parts := make([]valueRef, 0, len(x.Elems))
		for _, item := range x.Elems {
			parts = append(parts, g.lowerExpr(item))
		}
		if len(parts) == 0 {
			return valueRef{lt, "undef"}
		}
		r := g.tmp()
		g.emit(r + " = insertvalue " + lt + " undef, " + parts[0].lt + " " + parts[0].s + ", 0")
		for i := 1; i < len(parts); i++ {
			r2 := g.tmp()
			g.emit(r2 + " = insertvalue " + lt + " " + r + ", " + parts[i].lt + " " + parts[i].s +
				", " + strconv.Itoa(i))
			r = r2
		}
		return valueRef{lt, r}
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
	case *ast.Cast:
		return g.lowerCast(x, lt)
	case *ast.Len:
		return g.lowerLen(x)
	case *ast.Copy:
		return g.lowerExpr(x.Operand)
	case *ast.Move:
		return g.lowerExpr(x.Operand)
	case *ast.TryPropagate:
		return g.lowerTryPropagate(x, lt)
	case *ast.TryDetect:
		g.unsupported("`?(expr)` 的代码生成待批次（M2）")
		return valueRef{lt, g.zero(t)}
	case *ast.Alloc:
		return g.lowerAlloc(x)
	case *ast.SizeOf:
		return valueRef{lt, strconv.Itoa(g.layoutOf(x.Target, false))}
	case *ast.AlignOf:
		return valueRef{lt, strconv.Itoa(g.layoutOf(x.Target, true))}
	case *ast.Match:
		s := g.alloca("__match", lt)
		g.lowerMatch(x, s)
		v := g.tmp()
		g.emit(v + " = load " + lt + ", ptr " + s)
		return valueRef{lt, v}
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

// layoutOf 计算 `#size(T)` / `#align(T)`。
func (g *funcGen) layoutOf(ref *ast.TypeRef, align bool) int {
	t := g.l.table.Resolve(ref, emptyBag())
	if t == nil {
		return 0
	}
	if align {
		return g.l.layout.Align(t)
	}
	return g.l.layout.Size(t)
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
		l.ensureGlobal(sym, lt, nil)
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
		g.emit(r + " = insertvalue { ptr, i64 } undef, ptr " + sym + ", 0")
		r2 := g.tmp()
		g.emit(r2 + " = insertvalue { ptr, i64 } " + r + ", i64 " + strconv.Itoa(len(data)) + ", 1")
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
				g.l.ensureGlobal(sym, baseLt, nil)
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
			g.l.ensureGlobal(sym, lt, nil)
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
			baseLt := "{ ptr, i64 }"
			if ref != nil && ref.Base != nil {
				baseLt = g.l.llty(ref.Base)
			}
			base := g.lowerExpr(x.Base)
			p := g.tmp()
			g.emit(p + " = extractvalue " + baseLt + " " + base.s + ", 0")
			n := g.tmp()
			g.emit(n + " = extractvalue " + baseLt + " " + base.s + ", 1")
			bad := g.tmp()
			g.emit(bad + " = icmp uge i64 " + iv.s + ", " + n)
			g.emitCheck(bad, "index out of bounds", x.Span)
			gp := g.tmp()
			g.emit(gp + " = getelementptr inbounds " + elemLt + ", ptr " + p + ", i64 " + iv.s)
			return elemLt, gp
		}
		blt, bptr := g.lowerPlace(x.Base)
		gp := g.tmp()
		g.emit(gp + " = getelementptr inbounds " + blt + ", ptr " + bptr + ", i64 0, i64 " + iv.s)
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
	if ref != nil && ref.Builtin != "" {
		switch ref.Builtin {
		case "Some":
			field := fields["v"]
			var v valueRef
			if field != nil {
				v = g.lowerExpr(field)
			} else {
				v = valueRef{"i64", "0"}
			}
			if opt, ok := g.ty(e).(types.Optional); ok && g.l.isNicheOptional(opt) {
				return valueRef{lt, v.s}
			}
			cur := g.tmp()
			g.emit(cur + " = insertvalue " + lt + " undef, i8 1, 0")
			nxt := g.tmp()
			g.emit(nxt + " = insertvalue " + lt + " " + cur + ", " + g.l.llty(ref.Payload) + " " + v.s + ", 1")
			return valueRef{lt, nxt}
		case "Ok", "Err":
			key := "e"
			if ref.Builtin == "Ok" {
				key = "v"
			}
			field := fields[key]
			var v valueRef
			if field != nil {
				v = g.lowerExpr(field)
			} else {
				v = valueRef{g.l.llty(ref.Payload), "undef"}
			}
			okT, errT := types.Type(types.I32), types.Type(types.I32)
			if r, ok := g.ty(e).(types.Result); ok {
				okT, errT = r.Ok, r.Err
			}
			return valueRef{lt, g.buildResult(lt, okT, errT, ref.Builtin == "Ok", v.s)}
		}
	}
	if ref != nil && ref.Enum != nil {
		tagLt := g.l.llty(ref.Enum.TagType)
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
			countVal = cv.s
		}
		bytesV = g.tmp()
		g.emit(bytesV + " = mul i64 " + countVal + ", " + strconv.Itoa(elemSize))
		countStr = countVal
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
				g.emit(gp + " = getelementptr inbounds " + elemLt + ", ptr " + p + ", i64 " + strconv.Itoa(i))
				g.emit("store " + iv.lt + " " + iv.s + ", ptr " + gp)
			}
		}
	}
	r := g.tmp()
	g.emit(r + " = insertvalue " + lt + " undef, ptr " + p + ", 0")
	r2 := g.tmp()
	g.emit(r2 + " = insertvalue " + lt + " " + r + ", i64 " + countStr + ", 1")
	return valueRef{lt, r2}
}

// lowerLen 生成 `len`。
func (g *funcGen) lowerLen(e *ast.Len) valueRef {
	t := g.l.check.LenRefs[e]
	if f, ok := t.(types.Fixed); ok {
		return valueRef{"i64", strconv.FormatInt(f.N, 10)}
	}
	lt := "{ ptr, i64 }"
	if t != nil {
		lt = g.l.llty(t)
	}
	v := g.lowerExpr(e.Operand)
	r := g.tmp()
	g.emit(r + " = extractvalue " + lt + " " + v.s + ", 1")
	return valueRef{"i64", r}
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
			"<": map[bool]string{true: "slt", false: "ult"}[signed],
			"<=": map[bool]string{true: "sle", false: "ule"}[signed],
			">": map[bool]string{true: "sgt", false: "ugt"}[signed],
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
		kind := ""
		switch {
		case op == "+" && signed:
			kind = "sadd"
		case op == "+":
			kind = "uadd"
		case op == "-" && signed:
			kind = "ssub"
		case op == "-":
			kind = "usub"
		case op == "*" && signed:
			kind = "smul"
		default:
			kind = "umul"
		}
		pair := g.tmp()
		g.emit(pair + " = call { " + lt + ", i1 } @llvm." + kind + ".with.overflow." + lt +
			"(" + lt + " " + a + ", " + lt + " " + b + ")")
		val := g.tmp()
		g.emit(val + " = extractvalue { " + lt + ", i1 } " + pair + ", 0")
		ovf := g.tmp()
		g.emit(ovf + " = extractvalue { " + lt + ", i1 } " + pair + ", 1")
		g.emitCheck(ovf, "integer overflow", span)
		return val
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

// lowerCast 生成显式转换。
func (g *funcGen) lowerCast(e *ast.Cast, lt string) valueRef {
	types2, ok := g.l.check.CastTypes[e]
	src, dst := types.Type(types.I32), types.Type(types.I32)
	if ok {
		src, dst = types2[0], types2[1]
	}
	v := g.lowerExpr(e.Operand)
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
			return valueRef{dlt, v.s}
		}
		if db > sb {
			g.emit(r + " = fpext " + slt + " " + v.s + " to " + dlt)
		} else {
			g.emit(r + " = fptrunc " + slt + " " + v.s + " to " + dlt)
		}
		return valueRef{dlt, r}
	case isF(slt) && isI(dlt):
		g.emit(r + " = fptosi " + slt + " " + v.s + " to " + dlt)
		return valueRef{dlt, r}
	case isI(slt) && isF(dlt):
		if signedType(src) {
			g.emit(r + " = sitofp " + slt + " " + v.s + " to " + dlt)
		} else {
			g.emit(r + " = uitofp " + slt + " " + v.s + " to " + dlt)
		}
		return valueRef{dlt, r}
	case slt == dlt:
		return valueRef{dlt, v.s}
	case db > sb:
		if signedType(src) {
			g.emit(r + " = sext " + slt + " " + v.s + " to " + dlt)
		} else {
			g.emit(r + " = zext " + slt + " " + v.s + " to " + dlt)
		}
		return valueRef{dlt, r}
	}
	g.emit(r + " = trunc " + slt + " " + v.s + " to " + dlt)
	return valueRef{dlt, r}
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
	rest := sig.Params
	if len(rest) > 0 && rest[0].Name == "self" {
		rest = rest[1:]
	}
	for i, arg := range args {
		av := g.lowerExpr(arg)
		var plt string
		if i < len(rest) {
			plt = g.l.paramLLType(sig, rest[i].Type)
		} else {
			plt = av.lt
		}
		if plt == "ptr" && av.lt != "ptr" {
			av.s = g.viewPtr(arg, av.s, av.lt)
		} else if plt != av.lt && isIntLL(plt) && (isIntLL(av.lt) || av.lt == "i1") {
			av.s = g.coerceInt(av.lt, plt, av.s)
		}
		out = append(out, plt+" "+av.s)
	}
	sym := g.l.symbolOf(sig)
	retLt, params := g.l.funcLLType(sig)
	calleeText := "@" + sym
	if kind == "fnptr" {
		if callee == nil {
			if x, ok := expr.(*ast.Call); ok {
				callee = x.Callee
			}
		}
		if callee != nil {
			cv := g.lowerExpr(callee)
			calleeText = cv.s
		}
	}
	g.l.ensureDeclare(sym, retLt+" @"+sym+"("+strings.Join(params, ", ")+")")
	if retLt == "void" {
		g.emit("call void " + calleeText + "(" + strings.Join(out, ", ") + ")")
		return valueRef{"void", "undef"}
	}
	r := g.tmp()
	g.emit(r + " = call " + retLt + " " + calleeText + "(" + strings.Join(out, ", ") + ")")
	return valueRef{retLt, r}
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
	if mode == "result" {
		ev := g.tmp()
		g.emit(ev + " = extractvalue " + ltIn + " " + v.s + ", 2")
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
