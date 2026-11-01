package codegen

import (
	"strconv"
	"strings"

	"aila/internal/ast"
	"aila/internal/diag"
	"aila/internal/sema"
	"aila/internal/types"
)

// lowerSelect 生成 `select { case ... => {...}, ... }`（designs/0004 §6）：
// 声明序 try 链——依次探测各 case 事件，首个就绪者执行其块；全不就绪时协程内
// 挂起（coro.suspend）并在恢复后从头重试，非协程上下文（main）重走 try 链
// （等价用户手写的轮转驱动循环）。
//
// `resume h` 是**等完成**事件：探测即推进一次协程，完成（state 1|2）才算就绪，
// 推进后仍挂起（state 0）按不就绪处理——「还要继续等」不再作为可观测的第三态
// 逃到用户代码里（旧三态判别和与 `Suspended` / `Done` 内建模式已退役）。
func (g *funcGen) lowerSelect(x *ast.Select) {
	type evInfo struct {
		kind string // "chan" / "coro"
		out  string // chan 接收槽
		co   types.CoroState
		lt   string // coro 判别和 IR 类型（编译器内部三态表示）
		sum  string // coro 判别和落槽（arm 里取）
		bind *sema.Symbol
		ch   ast.Node
		op   ast.Node
		span diag.Span
	}
	infos := make([]evInfo, len(x.Cases))
	for i, sc := range x.Cases {
		inf := evInfo{bind: g.l.check.SelectBinds[sc], span: sc.Span}
		switch ev := sc.Event.(type) {
		case *ast.Recv:
			inf.kind = "chan"
			inf.ch = ev.Chan
			inf.out = g.alloca("__sel_recv", g.l.llty(g.ty(ev)))
		case *ast.Resume:
			inf.kind = "coro"
			inf.op = ev.Operand
			t := g.ty(ev)
			inf.lt = g.l.llty(t)
			inf.co, _ = t.(types.CoroState)
			inf.sum = g.alloca("__sel_sum", inf.lt)
		}
		infos[i] = inf
	}
	armLabels := make([]string, len(x.Cases))
	for i := range x.Cases {
		armLabels[i] = g.newLabel("sel.arm")
	}
	retry := g.newLabel("sel.retry")
	g.emit("br label %" + retry)
	g.emitRaw(retry + ":")
	g.terminated = false
	nxt := ""
	for i, inf := range infos {
		if nxt != "" {
			g.emitRaw(nxt + ":")
			g.terminated = false
		}
		if inf.kind == "chan" {
			ch := g.lowerExpr(inf.ch)
			g.l.ensureDeclare("aila_rt_chan_try_recv", "i1 @aila_rt_chan_try_recv(ptr, ptr)")
			ok := g.tmp()
			g.emit(ok + " = call i1 @aila_rt_chan_try_recv(ptr " + ch.s + ", ptr " + inf.out + ")")
			nxt = g.newLabel("sel.next")
			g.emit("br i1 " + ok + ", label %" + armLabels[i] + ", label %" + nxt)
			continue
		}
		// coro：state==3（已消费）→ 永远不就绪；否则推进一次 → 完成才进 arm。
		hv := g.lowerExpr(inf.op) // 句柄 = 控制块指针（load 槽，无副作用）
		st0 := g.tmp()
		g.emit(st0 + " = load i8, ptr " + hv.s)
		dead := g.tmp()
		g.emit(dead + " = icmp eq i8 " + st0 + ", 3")
		probe := g.newLabel("sel.coro")
		nxt = g.newLabel("sel.next")
		g.emit("br i1 " + dead + ", label %" + nxt + ", label %" + probe)
		g.emitRaw(probe + ":")
		g.terminated = false
		sum := g.coroResume(inf.co, hv, inf.lt, inf.span)
		g.emit("store " + inf.lt + " " + sum + ", ptr " + inf.sum)
		st1 := g.tmp()
		g.emit(st1 + " = extractvalue " + inf.lt + " " + sum + ", 0")
		done := g.tmp()
		g.emit(done + " = icmp ne i8 " + st1 + ", 0")
		g.emit("br i1 " + done + ", label %" + armLabels[i] + ", label %" + nxt)
	}
	g.emitRaw(nxt + ":")
	g.terminated = false
	if g.isCoro {
		g.coroSuspendRetry(retry) // 协程内：全不就绪挂起，恢复后重走 try 链
	} else {
		g.emit("br label %" + retry) // 非协程：无从挂起，重走 try 链（轮转驱动）
		g.terminated = true
	}
	endLbl := g.newLabel("sel.end")
	for i, sc := range x.Cases {
		g.emitRaw(armLabels[i] + ":")
		g.terminated = false
		inf := infos[i]
		if inf.kind == "coro" {
			g.storeCoroDone(inf.bind, inf.lt, inf.sum)
		} else if inf.bind != nil {
			s := g.slotFor(inf.bind)
			v := g.tmp()
			g.emit(v + " = load " + s.lt + ", ptr " + inf.out)
			g.emit("store " + s.lt + " " + v + ", ptr " + s.ptr)
		}
		g.lowerBlock(sc.Body, false)
		if !g.terminated {
			g.emit("br label %" + endLbl)
		}
	}
	g.emitRaw(endLbl + ":")
	g.terminated = false
}

// storeCoroDone 把 select coro 事件的完成态写进绑定槽（designs/0004 §6 v2）：
// 绑定类型是 `T ! E` 时把内部三态重装成 Result（flag 取 state==1），只有 `T`
// 时直接取 ok 载荷；无返回值（bind==nil）不绑定。
func (g *funcGen) storeCoroDone(bind *sema.Symbol, sumLt, sumSlot string) {
	if bind == nil {
		return
	}
	s := g.slotFor(bind)
	sum := g.tmp()
	g.emit(sum + " = load " + sumLt + ", ptr " + sumSlot)
	st := g.tmp()
	g.emit(st + " = extractvalue " + sumLt + " " + sum + ", 0")
	if _, isRes := bind.Type.(types.Result); !isRes {
		v := g.tmp()
		g.emit(v + " = extractvalue " + sumLt + " " + sum + ", 1")
		g.emit("store " + s.lt + " " + v + ", ptr " + s.ptr)
		return
	}
	r := bind.Type.(types.Result)
	okV := g.tmp()
	g.emit(okV + " = extractvalue " + sumLt + " " + sum + ", 1")
	errV := g.tmp()
	g.emit(errV + " = extractvalue " + sumLt + " " + sum + ", 2")
	flag := g.tmp()
	g.emit(flag + " = icmp eq i8 " + st + ", 1")
	cur := g.tmp()
	g.emit(cur + " = insertvalue " + s.lt + " undef, i1 " + flag + ", 0")
	n1 := g.tmp()
	g.emit(n1 + " = insertvalue " + s.lt + " " + cur + ", " + g.l.llty(r.Ok) + " " + okV + ", 1")
	n2 := g.tmp()
	g.emit(n2 + " = insertvalue " + s.lt + " " + n1 + ", " + g.l.llty(r.Err) + " " + errV + ", 2")
	g.emit("store " + s.lt + " " + n2 + ", ptr " + s.ptr)
}

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

	idxSlot := g.alloca("__idx", g.l.lenLT())
	g.emit("store " + g.l.lenLT() + " 0, ptr " + idxSlot)

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
		g.emit(lv0 + " = load " + g.l.lenLT() + ", ptr " + gv)
		lenSlot := g.alloca("__len", g.l.lenLT())
		g.emit("store " + g.l.lenLT() + " " + lv0 + ", ptr " + lenSlot)
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
	g.emit(iV + " = load " + g.l.lenLT() + ", ptr " + idxSlot)
	var c string
	if isFixed {
		c = g.tmp()
		g.emit(c + " = icmp ult " + g.l.lenLT() + " " + iV + ", " + lenConst)
	} else {
		lv := g.tmp()
		g.emit(lv + " = load " + g.l.lenLT() + ", ptr " + lenConst)
		c = g.tmp()
		g.emit(c + " = icmp ult " + g.l.lenLT() + " " + iV + ", " + lv)
	}
	g.emit("br i1 " + c + ", label %" + bodyLbl + ", label %" + exitLbl)
	g.emitRaw(bodyLbl + ":")
	g.terminated = false
	ep := g.tmp()
	if isFixed {
		g.emit(ep + " = getelementptr inbounds " + fixedLt + ", ptr " + ptrForElem + ", " + g.l.lenLT() + " 0, " + g.l.lenLT() + " " + iV)
	} else {
		g.emit(ep + " = getelementptr inbounds " + elemLt + ", ptr " + ptrForElem + ", " + g.l.lenLT() + " " + iV)
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
			islot = &slot{lt: g.l.lenLT(), ptr: g.alloca(stmt.Index, g.l.lenLT())}
			g.namedSlots[stmt.Index] = islot
		}
		g.emit("store " + g.l.lenLT() + " " + iV + ", ptr " + islot.ptr)
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
	g.emit(ni + " = add " + g.l.lenLT() + " " + iV + ", 1")
	g.emit("store " + g.l.lenLT() + " " + ni + ", ptr " + idxSlot)
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
		if pptr == "" || pptr == "null" {
			// 判别值是**常量**（`#compile.match MODE` 的常态）：常量不是主名字，
			// lowerPlace 给不出地址。退回按值求值——否则生成 `load i8, ptr null`，
			// 真机上就是 LoadProhibited 死循环（裸机 epc1 落在该 l8ui 上）。
			v := g.lowerExpr(m.Head)
			g.emit("store " + lt + " " + v.s + ", ptr " + slotName)
			break
		}
		v := g.tmp()
		g.emit(v + " = load " + plt + ", ptr " + pptr)
		g.emit("store " + lt + " " + v + ", ptr " + slotName)
	default:
		_ = h
		v := g.lowerExpr(m.Head)
		g.emit("store " + lt + " " + v.s + ", ptr " + slotName)
	}

	endLbl := g.newLabel("m.end")
	// `#compile.match` 已由 sema 定下命中臂：只生成那一支，其余臂不产生指令
	//（它们此前已完整类型检查；见 designs/0006 §1.3）。
	if m.PruneArm >= 0 {
		arm := m.Arms[m.PruneArm]
		g.bindArm(arm, lt, slotName, headT, kind)
		g.emitArmBody(arm, valueSlot, endLbl)
		g.emitRaw(endLbl + ":")
		g.terminated = false
		return
	}
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
		g.emitArmBody(arm, valueSlot, endLbl)
	}
	g.emitRaw(endLbl + ":")
	g.terminated = false
}

// emitArmBody 生成一个 match 分支体：块体走 lowerBlock，表达式臂写进 valueSlot。
func (g *funcGen) emitArmBody(arm *ast.Arm, valueSlot, endLbl string) {
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
		return
	}
	bv := g.lowerExpr(body)
	if valueSlot != "" {
		g.emit("store " + bv.lt + " " + bv.s + ", ptr " + valueSlot)
	}
	g.emit("br label %" + endLbl)
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
	g.lowerDestructInto(stmt, stmt.Names, stmt.Value)
}

// lowerLetGroup 生成多值解构绑定 `let (a, b) = 值;`（通道双端等）。
func (g *funcGen) lowerLetGroup(x *ast.LetGroup) {
	g.lowerDestructInto(x, x.Names, x.Init)
}

// lowerDestructInto 是 DestructAssign / LetGroup 共用的按位拆解落槽；
// keyNode 是 sema 侧 DestructRefs 的符号表键。
func (g *funcGen) lowerDestructInto(keyNode ast.Node, names []string, value ast.Node) {
	t := g.ty(value)
	lt := g.l.llty(t)
	gp, _ := t.(types.Group)
	v := g.lowerExpr(value)
	s := g.alloca("__destr", lt)
	g.emit("store " + lt + " " + v.s + ", ptr " + s)
	syms := g.l.check.DestructRefs[keyNode]
	for i, name := range names {
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

// lowerChanSend 生成 `ch <- v;` 发送语句：值落临时槽，try 成功即返回；
// 失败挂起（coro.suspend），恢复后重试。仅协程内合法（sema E5001 已保证）。
func (g *funcGen) lowerChanSend(x *ast.ChanSend) {
	ch := g.lowerExpr(x.Chan)
	v := g.lowerExpr(x.Value)
	elemLt := v.lt
	if ch2, ok := g.ty(x.Chan).(types.Chan); ok {
		elemLt = g.l.llty(ch2.Elem)
	}
	tmp := g.alloca("__send", elemLt)
	g.emit("store " + elemLt + " " + v.s + ", ptr " + tmp)
	g.l.ensureDeclare("aila_rt_chan_try_send", "i1 @aila_rt_chan_try_send(ptr, ptr)")
	retry := g.newLabel("chan.send")
	g.emit("br label %" + retry)
	g.emitRaw(retry + ":")
	ok := g.tmp()
	g.emit(ok + " = call i1 @aila_rt_chan_try_send(ptr " + ch.s + ", ptr " + tmp + ")")
	done := g.newLabel("chan.sent")
	wait := g.newLabel("chan.wait")
	g.emit("br i1 " + ok + ", label %" + done + ", label %" + wait)
	g.emitRaw(wait + ":")
	g.terminated = false
	g.coroSuspendRetry(retry)
	g.emitRaw(done + ":")
	g.terminated = false
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
