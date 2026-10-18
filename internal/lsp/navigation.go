package lsp

import (
	"strings"

	"aila/internal/ast"
	"aila/internal/lexer"
)

// 本文件实现「跳转类」补充能力：
//   - typeDefinition：跳到「变量/参数/字段的类型声明」
//   - implementation：跳到「壳内前向声明 ↔ 块内实现」的另一半（AILA 特有，`E7031` 报的就是这件事）
//   - documentLink：`import` 路径变成可点击的链接

// TypeDefinition 返回光标处符号的**类型**声明位置。
func (w *Workspace) TypeDefinition(uri string, pos Position) []Location {
	path := URI2Path(uri)
	a := w.analysisFor(path)
	info := a.ByPath[path]
	if info == nil {
		return nil
	}
	word, _, ok := identifierAt(info.Text, pos)
	if !ok {
		return nil
	}
	// ① 变量 / 参数 / 常量 → 它们的类型名
	if tn := localTypeName(info, word); tn != "" {
		if loc := typeDeclLocation(a, info, tn); loc != nil {
			return []Location{*loc}
		}
	}
	// ② 本身就是类型名（或 `Type.member` 的 `Type`）
	if loc := typeDeclLocation(a, info, word); loc != nil {
		return []Location{*loc}
	}
	if qual := qualifierBefore(info.Text, pos); qual != "" {
		if loc := typeDeclLocation(a, info, qual); loc != nil {
			return []Location{*loc}
		}
	}
	return nil
}

// typeDeclLocation 找某个短类型名的声明位置。
func typeDeclLocation(a *analysis, info *moduleInfo, short string) *Location {
	if short == "" {
		return nil
	}
	if lexer.TypeWords[short] {
		return nil // 内置类型没有声明处
	}
	for _, owner := range typeOwners(a, info, short) {
		for _, s := range owner.mod.Symbols {
			if s.Local || s.Name != owner.name {
				continue
			}
			if s.Kind == KindStruct || s.Kind == KindEnum || s.Kind == KindTypeParam {
				return &Location{URI: owner.mod.URI, Range: s.Range}
			}
		}
	}
	return nil
}

// Implementation 返回「壳内前向声明 ↔ 块内实现」的另一半。
func (w *Workspace) Implementation(uri string, pos Position) []Location {
	path := URI2Path(uri)
	a := w.analysisFor(path)
	info := a.ByPath[path]
	if info == nil {
		return nil
	}
	word, _, ok := identifierAt(info.Text, pos)
	if !ok {
		return nil
	}
	target := info
	qual := qualifierBefore(info.Text, pos)
	if qual != "" {
		if dep, ok := info.Imports[qual]; ok {
			target = a.Modules[dep]
		}
		if target == nil {
			return nil
		}
	}
	sig := sigByName(target, word)
	if sig == nil && qual != "" {
		sig = sigByName(target, qual+"."+word)
	}
	if sig == nil {
		return nil
	}
	var out []Location
	add := func(fn *ast.FuncDecl) {
		if fn == nil {
			return
		}
		out = append(out, Location{URI: target.URI, Range: spanRange(fn.Span)})
	}
	// 前向声明与实现都返回：编辑器会弹出二选一，也保证「在声明上按 Ctrl+F12」总能跳到实现
	if sig.Decl != nil && sig.Impl != nil && sig.Decl.Span != sig.Impl.Span {
		add(sig.Decl)
		add(sig.Impl)
		return out
	}
	if sig.Impl != nil {
		add(sig.Impl)
		return out
	}
	add(sig.Decl)
	return out
}

// DocumentLinks 把 `import` 路径变成指向目标模块文件的可点击链接。
func (w *Workspace) DocumentLinks(uri string) []DocumentLink {
	path := URI2Path(uri)
	a := w.analysisFor(path)
	info := a.ByPath[path]
	if info == nil || info.Mod == nil {
		return []DocumentLink{}
	}
	lines := lineContents(info.Text)
	out := []DocumentLink{}
	for _, imp := range info.Mod.Imports {
		target := importTargetModule(imp)
		if target == "" {
			continue
		}
		dep := a.Modules[target]
		if dep == nil {
			continue
		}
		line := imp.Span.Line - 1
		if line < 0 || line >= len(lines) {
			continue
		}
		r, ok := importPathRange(lines[line], line, imp.Span.Col-1)
		if !ok {
			continue
		}
		tip := "打开模块 " + target
		if dep.Generated {
			tip = "C 包 " + dep.Cpack + " 的绑定（bindgen 生成）"
		}
		out = append(out, DocumentLink{Range: r, Target: dep.URI, Tooltip: tip})
	}
	return out
}

// importPathRange 在 import 那一行里圈出路径部分（`import std.io as io;` → `std.io`）。
func importPathRange(line string, lineNo, col int) (Range, bool) {
	if col < 0 || col > len(line) {
		return Range{}, false
	}
	rest := line[col:]
	idx := strings.Index(rest, "import")
	if idx < 0 {
		return Range{}, false
	}
	start := col + idx + len("import")
	for start < len(line) && (line[start] == ' ' || line[start] == '\t') {
		start++
	}
	end := start
	for end < len(line) {
		c := line[end]
		if c == '_' || c == '.' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' {
			end++
			continue
		}
		break
	}
	if end <= start {
		return Range{}, false
	}
	return Range{
		Start: Position{Line: lineNo, Character: start},
		End:   Position{Line: lineNo, Character: end},
	}, true
}
