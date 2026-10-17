package sema

import (
	"strings"

	"github.com/aila-lang/aila/internal/ast"
	"github.com/aila-lang/aila/internal/types"
)

// 可见性阶梯（designs/0001 §2.2）：`public` 的含义是「提一级」。
//
//	块内（`module { }` 里）   默认 文件内   `public` → 子包内
//	壳内（`module { }` 外面） 默认 子包内   `public` → 包外（跨子包 / 跨项目）
//
// 轻量形态（文件里没有 `module` 块）的顶层声明字面上「写在 module 外面」，
// 因此按壳语义取等级。
const (
	visFile       = iota + 1 // 只有本文件可见
	visSubpackage            // 同一个子包（目录）内可见
	visPackage               // 包外可见（跨子包、跨项目）
)

// levelOf 由「是否 public」「是否写在壳里」推出可见性等级。
func levelOf(isPublic, inShell bool) int {
	if inShell {
		if isPublic {
			return visPackage
		}
		return visSubpackage
	}
	if isPublic {
		return visSubpackage
	}
	return visFile
}

// SubpackageOf 从模块全名取子包名：`_.math.ops` → `_.math`；`mylib.main` → `mylib`。
//
// 包名 = 第一段（`_` / 依赖名 / `std`），子包名 = 去掉模块名之后的路径。
func SubpackageOf(module string) string {
	parts := strings.Split(module, ".")
	if len(parts) <= 2 {
		return parts[0]
	}
	return strings.Join(parts[:len(parts)-1], ".")
}

// visOK 判断「从 fromModule 里引用 declModule 的声明」是否允许。
func visOK(fromModule, declModule string, isPublic, inShell bool) bool {
	if fromModule == declModule {
		return true // 同一个文件
	}
	switch levelOf(isPublic, inShell) {
	case visFile:
		return false
	case visSubpackage:
		return PackageOf(fromModule) == PackageOf(declModule) &&
			SubpackageOf(fromModule) == SubpackageOf(declModule)
	}
	return true
}

// requireVisible 校验一次跨模块引用；不允许时登记 E7020。
func (c *Checker) requireVisible(p ast.Node, name, kind, declModule string, isPublic, inShell bool) bool {
	if visOK(c.modName, declModule, isPublic, inShell) {
		return true
	}
	level := levelOf(isPublic, inShell)
	hint := "在声明处加 `public`（块内 → 子包内；壳内 → 包外）"
	switch level {
	case visFile:
		hint = "它只在本文件可见：把声明移到壳里（默认子包内），或加 `public`（子包内）"
	case visSubpackage:
		hint = "它只在子包内可见：把声明写到 `module` 块外并加 `public`（包外可见）"
	}
	c.err("E7020", ast.SpanOf(p), "`"+name+"` 是"+kind+"但在此不可见", "可见性阶梯", hint)
	return false
}

// sigVisible 判定一次函数/方法引用是否可见。
func (c *Checker) sigVisible(p ast.Node, sig *FuncSig) bool {
	if sig == nil {
		return true
	}
	mod := sig.Module
	if mod == "" {
		mod = c.modName
	}
	return c.requireVisible(p, sig.Name, kindOfSig(sig), mod, sig.IsPublic, sig.InShell)
}

func kindOfSig(sig *FuncSig) string {
	if sig.Owner != "" {
		return "方法"
	}
	return "函数"
}

// visOfType 取类型声明的可见性两位（module / public / inShell）。
func visOfType(info types.Info) (module string, isPublic, inShell bool) {
	switch x := info.(type) {
	case *types.StructInfo:
		return x.Module, x.IsPublic, x.InShell
	case *types.EnumInfo:
		return x.Module, x.IsPublic, x.InShell
	}
	return "", true, true
}

// typeVisibleFrom 判定「从 fromModule 引用 info 这个类型」是否允许。
func typeVisibleFrom(fromModule string, info types.Info) bool {
	mod, pub, shell := visOfType(info)
	if mod == "" {
		return true
	}
	return visOK(fromModule, mod, pub, shell)
}
