package sema

import "aila/internal/ast"

// PruneRecord 是一条生效的编译期剪枝记录（供 `aila build --emit-json` 报告，
// designs/0006 §1.3 规则 7 的可观测性）。
type PruneRecord struct {
	Module string // 模块名
	Kind   string // "if" | "match"
	Line   int    // 源码行号
	Value  int64  // 判定值
	Branch int    // 选中的分支：if 里 0=Then、i+1=Elifs[i]、len(Elifs)+1=Else；match 里为臂下标
}

// Prunes 收集本模块里所有生效的 `#compile.if` / `#compile.match` 剪枝点。
func (u *Unit) Prunes() []PruneRecord {
	if u == nil || u.Mod == nil || u.Check == nil {
		return nil
	}
	var out []PruneRecord
	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		switch x := n.(type) {
		case nil:
			return
		case *ast.Block:
			if x == nil {
				return
			}
			for _, s := range x.Stmts {
				walk(s)
			}
		case *ast.FuncDecl:
			if x == nil {
				return
			}
			if x.Body != nil {
				walk(x.Body)
			}
		case *ast.If:
			if x == nil {
				return
			}
			if x.CompileTime && x.Prune >= 0 {
				v, _ := u.Check.constExprInt(x.Cond)
				out = append(out, PruneRecord{Module: u.Name, Kind: "if", Line: x.Span.Line, Value: v, Branch: x.Prune})
			}
			walk(x.Cond)
			walk(x.Then)
			for _, p := range x.Elifs {
				if e, ok := p.A.(ast.Node); ok {
					walk(e)
				}
				if b, ok := p.B.(*ast.Block); ok {
					walk(b)
				}
			}
			walk(x.Els)
		case *ast.Match:
			if x == nil {
				return
			}
			if x.CompileTime && x.PruneArm >= 0 {
				v, _ := u.Check.constExprInt(x.Head)
				out = append(out, PruneRecord{Module: u.Name, Kind: "match", Line: x.Span.Line, Value: v, Branch: x.PruneArm})
			}
			walk(x.Head)
			for _, arm := range x.Arms {
				walk(arm.Body)
			}
		case *ast.ForC:
			if x == nil {
				return
			}
			walk(x.Body)
		case *ast.ForIn:
			if x == nil {
				return
			}
			walk(x.Body)
		case *ast.Region:
			if x == nil {
				return
			}
			walk(x.Body)
		case *ast.Unsafe:
			if x == nil {
				return
			}
			walk(x.Body)
		case *ast.Defer:
			if x == nil {
				return
			}
			walk(x.Body)
		}
	}
	for _, d := range u.Mod.Shell {
		walk(d)
	}
	for _, d := range u.Mod.Block {
		walk(d)
	}
	return out
}
