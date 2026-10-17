package codegen

import (
	"strconv"
	"strings"

	"github.com/aila-lang/aila/internal/ast"
	"github.com/aila-lang/aila/internal/diag"
	"github.com/aila-lang/aila/internal/sema"
	"github.com/aila-lang/aila/internal/types"
)

// --------------------------------------------------------------------------- //
// 循环
// --------------------------------------------------------------------------- //

// lowerForC 生成 C 风格 for。
func (g *funcGen) lowerForC(stmt *ast.ForC) {
	exitLbl := g.newLabel("for.end")
	head := g.newLabel("for.cond")
	stepLbl := g.newLabel("for.step")
	if stmt.Init != nil {
		g.lowerBinding(stmt.Init)
	}
	g.emit("br label %" + head)
	g.emitRaw(head + ":")
	g.terminated = false
	if stmt.Cond != nil {
		v := g.lowerExpr(stmt.Cond)
		cb := g.toBool(stmt.Cond, v.s)
		bodyLbl := g.newLabel("for.body")
		g.emit("br i1 " + cb + ", label %" + bodyLbl + ", label %" + exitLbl)
		g.emitRaw(bodyLbl + ":")
	}
	g.terminated = false
	g.loopExit = append(g.loopExit, exitLbl)
	g.loopCont = append(g.loopCont, stepLbl)
	g.lowerBlock(stmt.Body, true)
	g.loopExit = g.loopExit[:len(g.loopExit)-1]
	g.loopCont = g.loopCont[:len(g.loopCont)-1]
	if !g.terminated {
		g.emit("br label %" + stepLbl)
	}
	g.emitRaw(stepLbl + ":")
	g.terminated = false
	if stmt.Step != nil {
		g.lowerStmt(stmt.Step)
	}
	if !g.terminated {
		g.emit("br label %" + head)
	}
	g.emitRaw(exitLbl + ":")
	g.terminated = false
}

// lowerForIn 生成 for-in：`[T]` / Text 是胖视图，`[T; N]` 是内联定长数组。
func (g *funcGen) lowerForIn(stmt *ast.ForIn) {
	ref := g.l.check.ForRefs[stmt]
	exitLbl := g.newLabel("for.end")
	isFixed := false
	if ref != nil {
		_, isFixed = ref.Iter.(types.Fixed)
	}
	var elemT types.Type = types.I32
	if ref != nil && ref.Elem != nil {
		elemT = ref.Elem
	}
	elemLt := g.l.llty(elemT)

	idxSlot := g.alloca("__idx", "i64")
	g.emit("store i64 0, ptr " + idxSlot)

	lenConst := ""
	fixedLt := ""
	var ptrForElem string
	if isFixed {
		ft := ref.Iter.(types.Fixed)
		lenConst = strconv.FormatInt(ft.N, 10)
		fixedLt = g.l.llty(ft)
		// 迭代对象需要有地址：优先取左值，否则落进临时槽
		arrLt, basePtr := g.lowerPlace(stmt.Iterable)
		if arrLt != fixedLt {
			arr := g.lowerExpr(stmt.Iterable)
			tmp := g.alloca("__arr", arr.lt)
			g.emit("store " + arr.lt + " " + arr.s + ", ptr " + tmp)
			basePtr = tmp
		}
		ptrForElem = basePtr
	} else {
		it := g.lowerExpr(stmt.Iterable)
		lt := g.l.llty(ref.Iter)
		s := g.alloca("__iter", lt)
		g.emit("store " + lt + " " + it.s + ", ptr " + s)
		lv0 := g.tmp()
		gv := g.tmp()
		g.emit(gv + " = getelementptr inbounds " + lt + ", ptr " + s + ", i32 0, i32 1")
		g.emit(lv0 + " = load i64, ptr " + gv)
		lenSlot := g.alloca("__len", "i64")
		g.emit("store i64 " + lv0 + ", ptr " + lenSlot)
		p := g.tmp()
		g.emit(p + " = load ptr, ptr " + s)
		ptrForElem = p
		lenConst = lenSlot
	}

	head := g.newLabel("fi.cond")
	stepLbl := g.newLabel("fi.step")
	bodyLbl := g.newLabel("fi.body")
	g.emit("br label %" + head)
	g.emitRaw(head + ":")
	g.terminated = false
	iV := g.tmp()
	g.emit(iV + " = load i64, ptr " + idxSlot)
	var c string
	if isFixed {
		c = g.tmp()
		g.emit(c + " = icmp ult i64 " + iV + ", " + lenConst)
	} else {
		lv := g.tmp()
		g.emit(lv + " = load i64, ptr " + lenConst)
		c = g.tmp()
		g.emit(c + " = icmp ult i64 " + iV + ", " + lv)
	}
	g.emit("br i1 " + c + ", label %" + bodyLbl + ", label %" + exitLbl)
	g.emitRaw(bodyLbl + ":")
	g.terminated = false
	ep := g.tmp()
	if isFixed {
		g.emit(ep + " = getelementptr inbounds " + fixedLt + ", ptr " + ptrForElem + ", i64 0, i64 " + iV)
	} else {
		g.emit(ep + " = getelementptr inbounds " + elemLt + ", ptr " + ptrForElem + ", i64 " + iV)
	}
	ev := g.tmp()
	g.emit(ev + " = load " + elemLt + ", ptr " + ep)

	var varSym, idxSym = symOf(ref, true), symOf(ref, false)
	vslot := g.slotFor(varSym)
	if vslot == nil {
		vslot = &slot{lt: elemLt, ptr: g.alloca(stmt.Var, elemLt)}
		g.namedSlots[stmt.Var] = vslot
	}
	g.emit("store " + elemLt + " " + ev + ", ptr " + vslot.ptr)
	if stmt.Index != "" {
		islot := g.slotFor(idxSym)
		if islot == nil {
			islot = &slot{lt: "i64", ptr: g.alloca(stmt.Index, "i64")}
			g.namedSlots[stmt.Index] = islot
		}
		g.emit("store i64 " + iV + ", ptr " + islot.ptr)
	}

	g.loopExit = append(g.loopExit, exitLbl)
	g.loopCont = append(g.loopCont, stepLbl)
	g.lowerBlock(stmt.Body, true)
	g.loopExit = g.loopExit[:len(g.loopExit)-1]
	g.loopCont = g.loopCont[:len(g.loopCont)-1]
	if !g.terminated {
		g.emit("br label %" + stepLbl)
	}
	g.emitRaw(stepLbl + ":")
	g.terminated = false
	ni := g.tmp()
	g.emit(ni + " = add i64 " + iV + ", 1")
	g.emit("store i64 " + ni + ", ptr " + idxSlot)
	g.emit("br label %" + head)
	g.emitRaw(exitLbl + ":")
	g.terminated = false
}

func symOf(ref *sema.ForRef, isVar bool) *sema.Symbol {
	if ref == nil {
		return nil
	}
	if isVar {
		return ref.Var
	}
	return ref.Index
}

// --------------------------------------------------------------------------- //
// match
// --------------------------------------------------------------------------- //

// lowerMatch 生成 match：判别值落槽，逐分支比较，最后汇合。
func (g *funcGen) lowerMatch(m *ast.Match, valueSlot string) {
	ref := g.l.check.MatchRefs[m]
	kind := "int"
	var headT types.Type = types.I32
	if ref != nil {
		kind = ref.Kind
		if ref.Head != nil {
			headT = ref.Head
		}
	}
	lt := g.l.llty(headT)

	slotName := g.alloca("__head", lt)
	switch h := m.Head.(type) {
	case *ast.Path, *ast.FieldAccess:
		plt, pptr := g.lowerPlace(m.Head)
		v := g.tmp()
		g.emit(v + " = load " + plt + ", ptr " + pptr)
		g.emit("store " + lt + " " + v + ", ptr " + slotName)
	default:
		_ = h
		v := g.lowerExpr(m.Head)
		g.emit("store " + lt + " " + v.s + ", ptr " + slotName)
	}

	endLbl := g.newLabel("m.end")
	armLabels := make([]string, len(m.Arms))
	for i := range m.Arms {
		armLabels[i] = g.newLabel("m.arm" + strconv.Itoa(i))
	}

	switch kind {
	case "enum":
		headVal := g.tmp()
		g.emit(headVal + " = load " + lt + ", ptr " + slotName)
		info := enumOf(headT)
		var tagv string
		if info != nil && info.Payload {
			tagv = g.tmp()
			g.emit(tagv + " = extractvalue " + lt + " " + headVal + ", 0")
		} else {
			tagv = headVal
		}
		tagLt := "i32"
		if info != nil {
			tagLt = g.l.llty(info.TagType)
		}
		for i, arm := range m.Arms {
			aref := g.l.check.ArmRefs[arm]
			if aref != nil && aref.Wild {
				continue
			}
			tag := int64(0)
			if aref != nil {
				tag = aref.Tag
			}
			c := g.tmp()
			g.emit(c + " = icmp eq " + tagLt + " " + tagv + ", " + strconv.FormatInt(tag, 10))
			nxt := g.newLabel("m.next")
			g.emit("br i1 " + c + ", label %" + armLabels[i] + ", label %" + nxt)
			g.emitRaw(nxt + ":")
			g.terminated = false
		}
	case "optional":
		headVal := g.tmp()
		g.emit(headVal + " = load " + lt + ", ptr " + slotName)
		var c string
		if opt, ok := headT.(types.Optional); ok && g.l.isNicheOptional(opt) {
			pv := g.tmp()
			g.emit(pv + " = extractvalue " + lt + " " + headVal + ", 0")
			c = g.tmp()
			g.emit(c + " = icmp ne ptr " + pv + ", null")
		} else {
			pv := g.tmp()
			g.emit(pv + " = extractvalue " + lt + " " + headVal + ", 0")
			c = g.tmp()
			g.emit(c + " = icmp ne i8 " + pv + ", 0")
		}
		for i, arm := range m.Arms {
			aref := g.l.check.ArmRefs[arm]
			if aref != nil && aref.Wild {
				continue
			}
			wantSome := aref != nil && aref.Builtin == "Some"
			use := c
			if !wantSome {
				use = g.tmp()
				g.emit(use + " = xor i1 " + c + ", true")
			}
			nxt := g.newLabel("m.next")
			g.emit("br i1 " + use + ", label %" + armLabels[i] + ", label %" + nxt)
			g.emitRaw(nxt + ":")
			g.terminated = false
		}
	case "result":
		headVal := g.tmp()
		g.emit(headVal + " = load " + lt + ", ptr " + slotName)
		flag := g.tmp()
		g.emit(flag + " = extractvalue " + lt + " " + headVal + ", 0")
		for i, arm := range m.Arms {
			aref := g.l.check.ArmRefs[arm]
			if aref != nil && aref.Wild {
				continue
			}
			wantOk := aref != nil && aref.Builtin == "Ok"
			use := flag
			if !wantOk {
				use = g.tmp()
				g.emit(use + " = xor i1 " + flag + ", true")
			}
			nxt := g.newLabel("m.next")
			g.emit("br i1 " + use + ", label %" + armLabels[i] + ", label %" + nxt)
			g.emitRaw(nxt + ":")
			g.terminated = false
		}
	default:
		headVal := g.tmp()
		g.emit(headVal + " = load " + lt + ", ptr " + slotName)
		for i, arm := range m.Arms {
			aref := g.l.check.ArmRefs[arm]
			if aref != nil && aref.Wild {
				continue
			}
			lit := int64(0)
			if aref != nil {
				lit = aref.Tag
			}
			c := g.tmp()
			g.emit(c + " = icmp eq " + lt + " " + headVal + ", " + strconv.FormatInt(lit, 10))
			nxt := g.newLabel("m.next")
			g.emit("br i1 " + c + ", label %" + armLabels[i] + ", label %" + nxt)
			g.emitRaw(nxt + ":")
			g.terminated = false
		}
	}

	wildIdx := -1
	for i, arm := range m.Arms {
		if aref := g.l.check.ArmRefs[arm]; aref != nil && aref.Wild {
			wildIdx = i
			break
		}
	}
	if wildIdx >= 0 {
		g.emit("br label %" + armLabels[wildIdx])
	} else {
		g.trapNow("match 未覆盖的判别值", m.Span)
		g.terminated = true
	}

	for i, arm := range m.Arms {
		g.emitRaw(armLabels[i] + ":")
		g.terminated = false
		g.bindArm(arm, lt, slotName, headT, kind)
		body := arm.Body
		g.outValue = nil
		if blk, ok := body.(*ast.Block); ok {
			g.lowerBlock(blk, true)
			if !g.terminated {
				if valueSlot != "" && g.outValue != nil {
					g.emit("store " + g.outValue.lt + " " + g.outValue.s + ", ptr " + valueSlot)
				}
				g.emit("br label %" + endLbl)
			}
		} else {
			bv := g.lowerExpr(body)
			if valueSlot != "" {
				g.emit("store " + bv.lt + " " + bv.s + ", ptr " + valueSlot)
			}
			g.emit("br label %" + endLbl)
		}
	}
	g.emitRaw(endLbl + ":")
	g.terminated = false
}

// bindArm 把分支模式绑定写进各自的槽。
func (g *funcGen) bindArm(arm *ast.Arm, headLt, slotName string, headT types.Type, kind string) {
	aref := g.l.check.ArmRefs[arm]
	if aref == nil || len(aref.Bindings) == 0 {
		return
	}
	for _, b := range aref.Bindings {
		blt := g.l.llty(b.Type)
		var v string
		if kind == "enum" {
			fieldNames := []string{}
			if aref.HasVar {
				for _, f := range aref.Variant.Payload {
					fieldNames = append(fieldNames, f.Name)
				}
			}
			idx := -1
			for i, n := range fieldNames {
				if n == b.Name {
					idx = i
				}
			}
			if idx >= 0 {
				gp := g.tmp()
				g.emit(gp + " = getelementptr inbounds " + headLt + ", ptr " + slotName +
					", i32 0, i32 1, i32 " + strconv.Itoa(idx))
				v = g.tmp()
				g.emit(v + " = load " + blt + ", ptr " + gp)
			} else {
				v = "undef"
			}
		} else {
			headVal := g.tmp()
			g.emit(headVal + " = load " + headLt + ", ptr " + slotName)
			switch kind {
			case "result":
				idx := 2
				if aref.Builtin == "Ok" {
					idx = 1
				}
				v = g.tmp()
				g.emit(v + " = extractvalue " + headLt + " " + headVal + ", " + strconv.Itoa(idx))
			case "optional":
				if opt, ok := headT.(types.Optional); ok && g.l.isNicheOptional(opt) {
					v = headVal
				} else {
					v = g.tmp()
					g.emit(v + " = extractvalue " + headLt + " " + headVal + ", 1")
				}
			default:
				v = "undef"
			}
		}
		s := g.slotFor(b.Sym)
		if s == nil {
			s = &slot{lt: blt, ptr: g.alloca(b.Name, blt)}
			g.namedSlots[b.Name] = s
		}
		g.emit("store " + blt + " " + v + ", ptr " + s.ptr)
	}
}

// --------------------------------------------------------------------------- //
// 赋值
// --------------------------------------------------------------------------- //

// lowerAssign 生成赋值（含复合赋值）。
func (g *funcGen) lowerAssign(stmt *ast.Assign) {
	lt, ptr := g.lowerPlace(stmt.Target)
	if stmt.Op == "=" {
		v := g.lowerExpr(stmt.Value)
		g.emit("store " + lt + " " + v.s + ", ptr " + ptr)
		return
	}
	op := strings.TrimSuffix(stmt.Op, "=")
	cur := g.tmp()
	g.emit(cur + " = load " + lt + ", ptr " + ptr)
	opT := g.ty(stmt.Target)
	v := g.lowerExpr(stmt.Value)
	res := g.arith(op, lt, cur, v.s, opT, signedType(opT), false, stmt.Span)
	g.emit("store " + lt + " " + res + ", ptr " + ptr)
}

// lowerDestructAssign 生成解构赋值 `(a, b) = 值;`。
func (g *funcGen) lowerDestructAssign(stmt *ast.DestructAssign) {
	t := g.ty(stmt.Value)
	lt := g.l.llty(t)
	gp, _ := t.(types.Group)
	v := g.lowerExpr(stmt.Value)
	s := g.alloca("__destr", lt)
	g.emit("store " + lt + " " + v.s + ", ptr " + s)
	syms := g.l.check.DestructRefs[stmt]
	for i, name := range stmt.Names {
		if name == "_" {
			continue
		}
		itemT := types.Type(types.I32)
		if i < len(gp.Items) {
			itemT = gp.Items[i]
		}
		itemLt := g.l.llty(itemT)
		p := g.tmp()
		g.emit(p + " = getelementptr inbounds " + lt + ", ptr " + s + ", i32 0, i32 " + strconv.Itoa(i))
		iv := g.tmp()
		g.emit(iv + " = load " + itemLt + ", ptr " + p)
		var sym *sema.Symbol
		if i < len(syms) {
			sym = syms[i]
		}
		dst := g.slotFor(sym)
		if dst == nil {
			dst = &slot{lt: itemLt, ptr: g.alloca(name, itemLt)}
			g.namedSlots[name] = dst
		}
		g.emit("store " + itemLt + " " + iv + ", ptr " + dst.ptr)
	}
}

// --------------------------------------------------------------------------- //
// 检查与陷阱
// --------------------------------------------------------------------------- //

// emitCheck 插入「条件为真 = 出错」的检查；cond 可以是 i1 值或 `icmp ...` 指令体。
func (g *funcGen) emitCheck(cond, message string, span diag.Span) {
	okLbl := g.newLabel("chk.ok")
	badLbl := g.newLabel("chk.bad")
	c := cond
	if strings.HasPrefix(cond, "icmp ") || strings.HasPrefix(cond, "fcmp ") {
		c = g.tmp()
		g.emit(c + " = " + cond)
	}
	g.emit("br i1 " + c + ", label %" + badLbl + ", label %" + okLbl)
	g.emitRaw(badLbl + ":")
	g.terminated = false
	g.trapNow(message, span)
	g.emitRaw(okLbl + ":")
	g.terminated = false
}

// trapNow 生成对 `aila_rt_trap` 的调用（编译期插入的陷阱统一出口）。
func (g *funcGen) trapNow(message string, span diag.Span) {
	msg := g.l.stringConst(message)
	file := g.l.fileConst()
	g.l.ensureDeclare("aila_rt_trap", "void @aila_rt_trap(ptr, i64, ptr, i64)")
	g.emit("call void @aila_rt_trap(ptr " + msg + ", i64 " + strconv.Itoa(len(message)) +
		", ptr " + file + ", i64 " + strconv.Itoa(span.Line) + ")")
	g.emit("unreachable")
	g.terminated = true
}

// buildResult 构造 `{ i1, ok, err }` 聚合值。
func (g *funcGen) buildResult(lt string, okT, errT types.Type, isOk bool, payload string) string {
	flag := "0"
	if isOk {
		flag = "1"
	}
	cur := g.tmp()
	g.emit(cur + " = insertvalue " + lt + " undef, i1 " + flag + ", 0")
	nxt := g.tmp()
	if isOk {
		g.emit(nxt + " = insertvalue " + lt + " " + cur + ", " + g.l.llty(okT) + " " + payload + ", 1")
	} else {
		g.emit(nxt + " = insertvalue " + lt + " " + cur + ", " + g.l.llty(errT) + " " + payload + ", 2")
	}
	return nxt
}
