package sema

import (
	"strings"

	"github.com/aila-lang/aila/internal/ast"
	"github.com/aila-lang/aila/internal/diag"
)

// Unit 是一个已装载模块的语义产物。
type Unit struct {
	Name string
	// Package 是该模块所属的包名（`_` / 依赖名 / `std`）。
	Package string
	Mod   *ast.Module
	Table *Table
	Sigs  *Sigs
	Check *Checker

	// Imports 是「导入绑定名 -> 被导入模块」。
	Imports map[string]*Unit
}

// Program 是多模块语义图（docs/specs/0002 §2 模块图）。
type Program struct {
	Entry string
	Units map[string]*Unit
	Order []string
	// Deps 是「包名 → 该包声明的依赖包名」（docs/rfcs/0004）：导入根必须是其中之一。
	Deps map[string][]string
}

// Unit 返回模块单元。
func (p *Program) Unit(name string) *Unit {
	if p == nil {
		return nil
	}
	return p.Units[name]
}

// PublicType 按「公开可见性」取类型表里的类型（跨模块引用只认 `public`）。
func (u *Unit) PublicType(name string) (*Table, string, bool) {
	if u == nil {
		return nil, name, false
	}
	if i := strings.LastIndex(name, "."); i >= 0 {
		binding, rest := name[:i], name[i+1:]
		sub, ok := u.Imports[binding]
		if !ok {
			return nil, name, false
		}
		return sub.PublicType(rest)
	}
	return u.Table, name, true
}

// BuildProgram 是 Go 侧的多模块语义入口：
// 声明类型 → 链接导入 → 解析类型体 → 构建签名 → 逐模块类型检查。
func BuildProgram(mods map[string]*ast.Module, order []string, entry string, bag *diag.Bag) *Program {
	return BuildProgramWithDeps(mods, order, entry, nil, bag)
}

// BuildProgramWithDeps 与 BuildProgram 相同，但带上跨项目依赖图（包名 → 依赖名）。
func BuildProgramWithDeps(mods map[string]*ast.Module, order []string, entry string,
	deps map[string][]string, bag *diag.Bag) *Program {
	p := &Program{Entry: entry, Units: map[string]*Unit{}, Order: order, Deps: deps}
	for _, name := range order {
		mod := mods[name]
		if mod == nil {
			continue
		}
		p.Units[name] = &Unit{
			Name:    name,
			Package: PackageOf(name),
			Mod:     mod,
			Table:   DeclareTypes(mod, bag),
			Imports: map[string]*Unit{},
		}
	}
	// 导入链接：先把绑定名解析成模块单元，再让类型表看见对方（类型体解析需要）
	for _, name := range order {
		u := p.Units[name]
		if u == nil {
			continue
		}
		for _, imp := range u.Mod.Imports {
			target, ok := p.resolveImport(u.Package, imp)
			if !ok {
				bag.Error("E7010", imp.Span,
					"跨项目依赖 `"+imp.Root+"` 未在 `[dependencies]` 里声明", "依赖根",
					"在 `aila.toml` 写 `[dependencies] "+imp.Root+" = { path = \"...\" }`（或 git 形式）")
				continue
			}
			binding := importBinding(imp)
			dep, ok := p.Units[target]
			if !ok {
				if imp.KindV == "ch" || imp.KindV == "cc" {
					bag.Error("E7011", imp.Span, "C 绑定导入需要 bindgen（Go 侧待做）",
						"`import "+imp.KindV+"` 要先由 bindgen 生成 `.ail`",
						"本批次先跑纯 AILA 模块；C 互操作在构建流水线批次补齐")
					continue
				}
				bag.Error("E3001", imp.Span, "找不到被导入的模块 `"+target+"`",
					"导入路径", "检查目录/文件名是否与导入路径一致（路径即身份）")
				continue
			}
			if _, dup := u.Imports[binding]; dup {
				bag.Error("E7004", imp.Span, "导入绑定名 `"+binding+"` 重复",
					"同一模块内导入名必须唯一", "用 `as` 起别名")
				continue
			}
			u.Imports[binding] = dep
			u.Table.Imports[binding] = dep.Table
		}
	}
	for _, name := range order {
		if u := p.Units[name]; u != nil {
			u.Table.ResolveBodies(u.Mod, bag)
		}
	}
	for _, name := range order {
		u := p.Units[name]
		if u == nil {
			continue
		}
		u.Sigs = BuildSigs(u.Mod, u.Table, bag)
	}
	for _, name := range order {
		u := p.Units[name]
		if u == nil {
			continue
		}
		u.Check = NewChecker(u.Mod, u.Table, u.Sigs, bag)
		u.Check.SetModuleContext(u.Name, u.Imports, p.Units)
		u.Check.pkg = u.Package
		u.Check.Run()
	}
	return p
}

// resolveImport 把导入语句解析成模块全名。
//
// 规则（designs/0001 §2.3 规则 1 + docs/rfcs/0004）：`_` = 当前包自身的根
// （依赖包里的 `_` 指该依赖自己）、`std` = 标准库、其余根必须是该包声明的依赖。
func (p *Program) resolveImport(pkg string, imp *ast.Import) (string, bool) {
	join := func(root string) string {
		if len(imp.Path) == 0 {
			return root
		}
		return root + "." + strings.Join(imp.Path, ".")
	}
	switch imp.Root {
	case "_":
		return join(pkg), true
	case "std":
		return join("std"), true
	}
	for _, dep := range p.Deps[pkg] {
		if dep == imp.Root {
			return join(dep), true
		}
	}
	return "", false
}

// importBinding 计算导入后的绑定名（默认取路径末段）。
func importBinding(imp *ast.Import) string {
	if imp.Alias != "" {
		return imp.Alias
	}
	if len(imp.Path) > 0 {
		return imp.Path[len(imp.Path)-1]
	}
	return string(imp.Root)
}
