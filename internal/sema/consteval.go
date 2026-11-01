package sema

import (
	"strconv"

	"aila/internal/ast"
)

// 编译期常量求值（designs/0006 §1.2）。
//
// `#compile.if` / `#compile.match` 的条件与判别式必须是**编译期常量表达式**：
//   - 字面量（整数 / 字符 / 布尔）
//   - 模块级 `const`（含 import 进来的模块 const）
//   - 函数内一次性 `let` 绑到 const-expr（`let` 不可重新赋值，天然 SSA）
//   - 仅由上述构成的纯运算（+ - * / % 位运算 比较 逻辑 取负 取反）
//
// 求不出 → 由调用方报错，绝不静默降级成运行时分支（见 check_stmt.go / check_match.go）。

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// constValueInt64 把 const 初始化器的标量值折成整数（布尔按 0/1，字符按码点）。
func constValueInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case int64:
		return t, true
	case int:
		return int64(t), true
	case int32:
		return int64(t), true
	case uint32:
		return int64(t), true
	case uint64:
		return int64(t), true
	case bool:
		return boolToInt64(t), true
	case string:
		rs := []rune(t)
		if len(rs) == 1 {
			return int64(rs[0]), true
		}
	}
	return 0, false
}

// constExprInt 求编译期整数常量；第二个返回值表示「是否编译期已知」。
func (c *Checker) constExprInt(n ast.Node) (int64, bool) {
	switch x := n.(type) {
	case nil:
		return 0, false
	case *ast.IntLit:
		return x.Raw, true
	case *ast.CharLit:
		rs := []rune(x.Value)
		if len(rs) == 1 {
			return int64(rs[0]), true
		}
		return 0, false
	case *ast.BoolLit:
		return boolToInt64(x.Value), true
	case *ast.Path:
		if ci, ok := c.PathConsts[x]; ok && ci != nil {
			return constValueInt64(ci.Value)
		}
		if sym, ok := c.PathSyms[x]; ok && sym != nil {
			if v, has := c.ConstVals[sym]; has {
				return v, true
			}
		}
		return 0, false
	case *ast.Unary:
		v, ok := c.constExprInt(x.Operand)
		if !ok {
			return 0, false
		}
		switch x.Op {
		case "-":
			return -v, true
		case "+":
			return v, true
		case "~":
			return ^v, true
		case "!":
			return boolToInt64(v == 0), true
		}
		return 0, false
	case *ast.Binary:
		a, ok1 := c.constExprInt(x.Left)
		b, ok2 := c.constExprInt(x.Right)
		if !ok1 || !ok2 {
			return 0, false
		}
		switch x.Op {
		case "+":
			return a + b, true
		case "-":
			return a - b, true
		case "*":
			return a * b, true
		case "/":
			if b == 0 {
				return 0, false
			}
			return a / b, true
		case "%":
			if b == 0 {
				return 0, false
			}
			return a % b, true
		case "&":
			return a & b, true
		case "|":
			return a | b, true
		case "^":
			return a ^ b, true
		case "<<":
			if b < 0 || b > 63 {
				return 0, false
			}
			return a << uint(b), true
		case ">>":
			if b < 0 || b > 63 {
				return 0, false
			}
			return a >> uint(b), true
		case "==":
			return boolToInt64(a == b), true
		case "!=":
			return boolToInt64(a != b), true
		case "<":
			return boolToInt64(a < b), true
		case "<=":
			return boolToInt64(a <= b), true
		case ">":
			return boolToInt64(a > b), true
		case ">=":
			return boolToInt64(a >= b), true
		case "&&":
			return boolToInt64(a != 0 && b != 0), true
		case "||":
			return boolToInt64(a != 0 || b != 0), true
		}
		return 0, false
	}
	return 0, false
}

// patternConstInt 求 `#compile.match` 分支模式的常量值（编译期匹配用）。
// 支持：整数/字符/布尔字面量、指向 const 的路径、以及已解析出的枚举变体判别值（经 ArmRef.Tag）。
func (c *Checker) patternConstInt(arm *ast.Arm) (int64, bool) {
	if arm == nil {
		return 0, false
	}
	switch p := arm.Pattern.(type) {
	case *ast.LitPattern:
		v, ok := constValueInt64(p.Value)
		if !ok {
			return 0, false
		}
		if p.Negative {
			v = -v
		}
		return v, true
	case *ast.Path, *ast.PathPattern, *ast.VariantPattern:
		// 枚举变体 / const 名：sema 已把可用信息挂在 ArmRef.Tag 上（enum 判别值）。
		if ref := c.ArmRefs[arm]; ref != nil && !ref.Wild {
			return ref.Tag, true
		}
	}
	return 0, false
}

// pruneCompileIf 给 `#compile.if` 定剪枝（designs/0006 §1.3 规则 1：求不出常量就报错，
// 绝不静默降级成运行时 if）。
func (c *Checker) pruneCompileIf(x *ast.If) {
	v, ok := c.constExprInt(x.Cond)
	if !ok {
		c.err("E3210", ast.SpanOf(x.Cond), "`#compile.if` 的条件不是编译期常量",
			"条件只能是 const、字面量、或绑到常量的 let（以及它们的纯运算）",
			"改成编译期常量，或改用普通 `if`（后者由优化器决定是否折叠）")
		return
	}
	if v != 0 {
		x.Prune = 0
		return
	}
	for i, pair := range x.Elifs {
		e, ok := pair.A.(ast.Node)
		if !ok {
			continue
		}
		ev, ok := c.constExprInt(e)
		if !ok {
			c.err("E3210", ast.SpanOf(e), "`#compile.if` 的 else-if 条件不是编译期常量",
				"同一链上的条件都必须是编译期常量",
				"改成编译期常量，或改用普通 `if`")
			return
		}
		if ev != 0 {
			x.Prune = i + 1
			return
		}
	}
	// 全为假：有 else 就选 else；没有 else 时这个下标越界，codegen 会什么都不生成。
	x.Prune = len(x.Elifs) + 1
}

// pruneCompileMatch 给 `#compile.match` 定剪枝（designs/0006 §1.3 规则 4）。
func (c *Checker) pruneCompileMatch(m *ast.Match) {
	v, ok := c.constExprInt(m.Head)
	if !ok {
		c.err("E3211", ast.SpanOf(m.Head), "`#compile.match` 的判别式不是编译期常量",
			"判别式只能是 const、字面量、或绑到常量的 let",
			"改成编译期常量，或改用普通 `match`")
		return
	}
	unsupported := false
	for i, arm := range m.Arms {
		if _, isWild := arm.Pattern.(*ast.Wildcard); isWild {
			m.PruneArm = i
			return
		}
		pv, ok := c.patternConstInt(arm)
		if !ok {
			unsupported = true
			continue
		}
		if pv == v {
			m.PruneArm = i
			return
		}
	}
	if unsupported {
		c.err("E3213", m.Span, "`#compile.match` 的分支模式不是编译期常量",
			"编译期匹配只支持字面量 / const / 枚举变体判别值",
			"改用 `#compile.if` 直接比较常量")
		return
	}
	c.err("E3212", m.Span, "`#compile.match` 的判别值 "+strconv.FormatInt(v, 10)+" 没有命中任何分支",
		"判别值在编译期已知，因此每个可能的取值都应有分支",
		"补上该值的分支，或在最后加 `_ => ...`")
}
