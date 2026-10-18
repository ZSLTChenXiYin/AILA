package sema

import (
	"aila/internal/ast"
	"aila/internal/types"
)

// checkMatch 检查 match：分支产出统一性 + 完备性 + 模式绑定。
//
// valuePos=true 表示处在取值位置（`let x = match …`），此时所有分支必须产出同型值；
// valuePos=false 是纯分派（语句位置），分支不得产出值。
func (c *Checker) checkMatch(m *ast.Match, valuePos bool, expected types.Type) types.Type {
	headT := c.checkExpr(m.Head, nil)
	ctx := c.ctx

	kind := "int"
	var enum *types.EnumInfo
	switch t := headT.(type) {
	case types.Named:
		if en, ok := t.Info.(*types.EnumInfo); ok {
			kind = "enum"
			enum = en
		}
	case types.Optional:
		kind = "optional"
	case types.Result:
		kind = "result"
	case types.Prim:
		if t.K == "bool" || t.K == "text" || t.K == "f32" || t.K == "f64" {
			c.err("E3201", ast.SpanOf(m.Head), "该类型不能作为 match 判别对象",
				"判别对象限：enum / `T?` / `T ! E` / 整数族 / `C32`", "")
		}
	}
	if kind == "int" && isFloatHead(headT) {
		c.err("E3201", ast.SpanOf(m.Head), "浮点不能作为 match 判别对象", "浮点相等比较是经典坑", "")
	}

	seen := map[string]bool{}
	wildcardAt := -1
	var armTypes []types.Type
	var litSeen []int64

	for idx, arm := range m.Arms {
		prevScope := ctx.Scope
		ctx.Scope = NewScope(prevScope)

		pat := arm.Pattern
		switch p := pat.(type) {
		case *ast.Wildcard:			if wildcardAt >= 0 {
				c.err("E3401", ast.SpanOf(p), "`_` 分支重复", "这里", "")
			}
			wildcardAt = idx
			c.noteArm(arm, &ArmRef{Wild: true})
			if idx != len(m.Arms)-1 {
				c.err("E3401", ast.SpanOf(p), "`_` 必须是最后一个分支", "其后不得再有分支（死分支禁止）", "")
			}
		case *ast.VariantPattern:
			name := p.Path[len(p.Path)-1]
			if kind == "enum" && enum != nil {
				v, ok := enum.Variant(name)
				if !ok {
					c.err("E3001", ast.SpanOf(p), "enum `"+enum.Name+"` 没有变体 `"+name+"`", "这里", "")
					break
				}
				if seen[v.Name] {
					c.err("E3401", ast.SpanOf(p), "变体 `"+name+"` 的分支重复", "同一模式出现两次", "")
				}
				seen[v.Name] = true
				c.ArmVariants[arm] = variantRef{enum: enum, variant: v}
				c.noteArm(arm, &ArmRef{Tag: v.Tag, Enum: enum, Variant: v, HasVar: true})
				for _, fp := range p.Fields_ {
					var ftype types.Type
					for _, f := range v.Payload {
						if f.Name == fp.Name {
							ftype = f.Type
						}
					}
					if ftype == nil {
						c.err("E3001", fp.Span, "变体 `"+name+"` 没有字段 `"+fp.Name+"`", "这里", "")
						continue
					}
					if fp.Binding != "" {
						bsym := &Symbol{Name: fp.Binding, Type: ftype, Kind: "local"}
						ctx.Scope.Declare(bsym)
						c.locals = append(c.locals, bsym)
						c.noteArmBinding(arm, fp.Name, ftype, bsym)
					}
				}
			} else if kind == "optional" || kind == "result" {
				want := map[string][]string{
					"optional": {"Some", "None"},
					"result":   {"Ok", "Err"},
				}[kind]
				okName := false
				for _, w := range want {
					if w == name {
						okName = true
					}
				}
				if !okName {
					c.err("E3001", ast.SpanOf(p), "该判别对象没有分支 `"+name+"`", "这里", "")
					break
				}
				if seen[name] {
					c.err("E3401", ast.SpanOf(p), "分支 `"+name+"` 重复", "这里", "")
				}
				seen[name] = true
				var payloadT types.Type
				if kind == "optional" {
					payloadT = headT.(types.Optional).Inner
					if name == "None" {
						payloadT = nil
					}
				} else {
					r := headT.(types.Result)
					if name == "Ok" {
						payloadT = r.Ok
					} else {
						payloadT = r.Err
					}
				}
				c.noteArm(arm, &ArmRef{Builtin: name})
				if payloadT != nil {
					for _, fp := range p.Fields_ {
						if fp.Binding != "" {
							bsym := &Symbol{Name: fp.Binding, Type: payloadT, Kind: "local"}
							ctx.Scope.Declare(bsym)
							c.locals = append(c.locals, bsym)
						c.locals = append(c.locals, bsym)
							c.noteArmBinding(arm, "v", payloadT, bsym)
						}
					}
				}
			}
		case *ast.PathPattern:
			name := p.Path[len(p.Path)-1]
			if kind == "optional" || kind == "result" {
				okName := name == "Some" || name == "None" || name == "Ok" || name == "Err"
				if !okName {
					c.err("E3001", ast.SpanOf(p), "该判别对象没有分支 `"+name+"`", "这里", "")
					break
				}
				if seen[name] {
					c.err("E3401", ast.SpanOf(p), "分支 `"+name+"` 重复", "这里", "")
				}
				seen[name] = true
				c.noteArm(arm, &ArmRef{Builtin: name})
			} else if kind == "enum" {
				if enum != nil {
					if v, ok := enum.Variant(name); ok {
						if seen[v.Name] {
							c.err("E3401", ast.SpanOf(p), "变体 `"+name+"` 的分支重复", "这里", "")
						}
						seen[v.Name] = true
						c.ArmVariants[arm] = variantRef{enum: enum, variant: v}
						c.noteArm(arm, &ArmRef{Tag: v.Tag, Enum: enum, Variant: v, HasVar: true})
					} else {
						c.err("E3001", ast.SpanOf(p), "enum `"+enum.Name+"` 没有变体 `"+name+"`", "这里", "")
					}
				}
			} else {
				// 整数族：const 名作模式
				if ci, ok := c.sigs.Consts[name]; ok {
					if v, isInt := ci.Value.(int64); isInt {
						for _, prev := range litSeen {
							if prev == v {
								c.err("E3401", ast.SpanOf(p), "模式重复", "同一模式出现两次", "")
							}
						}
						litSeen = append(litSeen, v)
						c.noteArm(arm, &ArmRef{Tag: v})
					}
				} else {
					c.err("E3401", ast.SpanOf(p), "模式必须是字面量 / `const` 名 / `_`", "这里", "")
				}
			}
		case *ast.LitPattern:
			if p.KindV == "float" || p.KindV == "str" || p.KindV == "bool" {
				c.err("E3401", ast.SpanOf(p), "该模式形态不合法", "这里", "")
			} else if v, ok := p.Value.(int64); ok {
				val := v
				if p.Negative {
					val = -val
				}
				for _, prev := range litSeen {
					if prev == val {
						c.err("E3401", ast.SpanOf(p), "模式重复", "这里", "")
					}
				}
				litSeen = append(litSeen, val)
				c.noteArm(arm, &ArmRef{Tag: val})
			}
		}

		// 分支体
		var armT types.Type
		switch body := arm.Body.(type) {
		case *ast.Block:
			c.checkBlock(body, false)
			var outs []types.Type
			for _, st := range body.Stmts {
				if out, ok := st.(*ast.Out); ok {
					outs = append(outs, c.Types[out.Value])
				}
			}
			if valuePos {
				if len(outs) == 0 && !c.blockDiverges(body) {
					c.err("E3404", ast.SpanOf(body), "取值位置的块分支必须以 `out 值;` 产出",
						"块值只走显式 `out`，没有隐式尾值", "在每条非发散路径上写 `out 值;`")
				}
				if len(outs) > 0 {
					armT = outs[len(outs)-1]
				}
			} else if len(outs) > 0 {
				c.err("E3402", ast.SpanOf(body), "语句位置的 match 分支不允许产出值", "这里", "")
			}
		default:
			bt := c.checkExpr(arm.Body, expectedOrNil(expected, valuePos))
			if valuePos {
				if isVoidOrNever(bt) {
					c.err("E3402", ast.SpanOf(arm.Body), "取值位置的分支必须产出同型值，这个分支没有值",
						"这里", "要「一半有值一半没值」请用 if 语句 + 占位初始化")
				} else {
					armT = bt
				}
			} else if !isVoidOrNever(bt) {
				c.err("E3402", ast.SpanOf(arm.Body), "语句位置的 match 分支不能产出值",
					"静默丢值是编译错误", "")
			}
		}
		armTypes = append(armTypes, armT)
		ctx.Scope = prevScope
	}

	// 产出统一性
	var result types.Type = types.Void{}
	if valuePos {
		var concrete types.Type
		for _, t := range armTypes {
			if t == nil {
				continue
			}
			if concrete == nil {
				concrete = t
				continue
			}
			if !sameType(concrete, t) {
				c.err("E3402", m.Span, "match 分支产出类型不一致", "所有分支必须同型", "")
				break
			}
		}
		if concrete != nil {
			if isUntypedInt(concrete) {
				if expected != nil && types.IsInteger(expected) {
					concrete = expected
				} else {
					concrete = types.I32
				}
			} else if isUntypedFloat(concrete) {
				if expected != nil && types.IsFloat(expected) {
					concrete = expected
				} else {
					concrete = types.F64
				}
			}
			result = concrete
			// 取值位置的字面量分支要回写定型结果：否则 codegen 会按未定型类型的
			// IR 表示（i64）产出值，与槽类型不符（defect #18）。
			for _, arm := range m.Arms {
				if _, isBlock := arm.Body.(*ast.Block); isBlock {
					continue
				}
				if bt, ok := c.Types[arm.Body]; ok {
					if _, untypedInt := bt.(types.UntypedInt); untypedInt {
						c.Types[arm.Body] = concrete
					} else if _, untypedFloat := bt.(types.UntypedFloat); untypedFloat {
						c.Types[arm.Body] = concrete
					}
				}
			}
		}
	}

	// 完备性
	switch kind {
	case "enum":
		if wildcardAt < 0 && enum != nil {
			var missing []string
			for _, v := range enum.Variants {
				if !seen[v.Name] {
					missing = append(missing, v.Name)
				}
			}
			if len(missing) > 0 {
				c.err("E3401", m.Span, "match 不完备：`"+enum.Name+"` 有变体没有分支",
					"封闭值域由编译器背书查漏", "补齐缺失变体，或在最后加 `_ => ...`")
			}
		}
	case "optional", "result":
		if wildcardAt < 0 {
			need := []string{"Some", "None"}
			if kind == "result" {
				need = []string{"Ok", "Err"}
			}
			for _, n := range need {
				if !seen[n] {
					c.err("E3401", m.Span, "match 不完备：缺少 `"+n+"` 分支", "封闭值域必须覆盖全部分支", "")
				}
			}
		}
	case "int":
		if wildcardAt < 0 {
			c.err("E3403", m.Span, "整数族 match 必须带 `_` 兜底", "整数域不可穷举",
				"在最后一个分支加 `_ => ...`")
		}
	}
	c.MatchResults[m] = result
	c.noteMatch(m, kind, headT, result)
	return result
}

func expectedOrNil(expected types.Type, valuePos bool) types.Type {
	if !valuePos {
		return nil
	}
	return expected
}

func isFloatHead(t types.Type) bool {
	p, ok := t.(types.Prim)
	return ok && (p.K == "f32" || p.K == "f64")
}
