package lsp

import (
	"strings"

	"github.com/ailang/aila/internal/types"
)

// 本文件实现查询类能力：大纲 / 跳转定义 / 悬停 / 查找引用 / 同文件高亮 / 工程符号搜索。

// DocumentSymbols 返回一个模块的大纲（层级：模块 → 类型与函数 → 成员与局部）。
func (w *Workspace) DocumentSymbols(uri string) []DocumentSymbol {
	path := URI2Path(uri)
	a := w.analysisFor(path)
	info := a.ByPath[path]
	if info == nil {
		return nil
	}
	root := DocumentSymbol{
		Name:           info.Name,
		Kind:           KindModule,
		Detail:         "模块",
		Range:          Range{Start: Position{}, End: Position{Line: 1 << 20, Character: 1000}},
		SelectionRange: Range{Start: Position{}, End: Position{Line: 0, Character: len(info.Name)}},
	}
	byContainer := map[string]*DocumentSymbol{}
	var top []*DocumentSymbol
	for _, s := range info.Symbols {
		if s.Local {
			continue // 局部绑定不放大纲（hover 仍然可用）
		}
		ds := DocumentSymbol{
			Name: s.Name, Kind: s.Kind, Detail: s.Detail,
			Range: s.Range, SelectionRange: s.Range,
		}
		switch {
		case s.Kind == KindField || (s.Kind == KindEnumMember && s.Container != ""):
			parent := byContainer[s.Container]
			if parent == nil {
				parent = &DocumentSymbol{Name: s.Container, Kind: KindStruct, Range: s.Range, SelectionRange: s.Range}
				byContainer[s.Container] = parent
				top = append(top, parent)
			}
			parent.Children = append(parent.Children, ds)
		case s.Kind == KindMethod && s.Container != "":
			parent := byContainer[s.Container]
			if parent == nil {
				parent = &DocumentSymbol{Name: s.Container, Kind: KindStruct, Range: s.Range, SelectionRange: s.Range}
				byContainer[s.Container] = parent
				top = append(top, parent)
			}
			parent.Children = append(parent.Children, ds)
		default:
			top = append(top, &ds)
		}
	}
	for _, ds := range top {
		root.Children = append(root.Children, *ds)
	}
	if len(root.Children) == 1 && root.Children[0].Name == info.Name {
		return root.Children
	}
	return []DocumentSymbol{root}
}

// Definition 返回跳转目标（找不到返回 nil）。
func (w *Workspace) Definition(uri string, pos Position) []Location {
	loc := w.resolve(uri, pos)
	if loc == nil {
		return nil
	}
	return []Location{*loc}
}

// resolve 解析光标处的标识符，返回其声明位置。
func (w *Workspace) resolve(uri string, pos Position) *Location {
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
	qual := qualifierBefore(info.Text, pos)

	// ① 限定名：`binding.symbol`（import 别名）或 `Type.symbol`（方法 / 变体 / 字段）
	if qual != "" {
		if target, ok := info.Imports[qual]; ok {
			if dep := a.Modules[target]; dep != nil {
				if s := findSymbol(dep, word, false); s != nil {
					return &Location{URI: dep.URI, Range: s.Range}
				}
				if s := findSymbol(dep, qual+"."+word, false); s != nil {
					return &Location{URI: dep.URI, Range: s.Range}
				}
			}
			return nil
		}
		if s := findSymbol(info, qual+"."+word, false); s != nil {
			return &Location{URI: info.URI, Range: s.Range}
		}
		// 字段：`x.field` → 找类型中的同名字段
		if owner := typeOfLocal(info, qual); owner != "" {
			if s := findSymbol(info, word, false); s != nil && s.Kind == KindField {
				return &Location{URI: info.URI, Range: s.Range}
			}
			_ = owner
		}
		return nil
	}

	// ② 局部：在包含光标的函数里找最近的同名绑定（声明必须早于使用点）
	var localHit *symbol
	var localModule *moduleInfo
	for _, s := range info.Symbols {
		if !s.Local || s.Name != word {
			continue
		}
		if spanLess(s.Range, pos_as_range(pos)) {
			continue // 声明在光标之后，不算
		}
		if localHit == nil || rangeStartsAfter(s.Range, localHit.Range) {
			localHit = s
			localModule = info
		}
	}
	if localHit != nil && localModule != nil {
		return &Location{URI: localModule.URI, Range: localHit.Range}
	}

	// ③ 模块级
	if s := findSymbol(info, word, false); s != nil {
		return &Location{URI: info.URI, Range: s.Range}
	}

	// ④ 导入模块里的模块级符号（`import x as io` 后裸写 `io.print` 已在 ① 处理；
	//    这里兜底：同名符号出现在某个导入模块里）
	for _, target := range info.Imports {
		if dep := a.Modules[target]; dep != nil {
			if s := findSymbol(dep, word, false); s != nil {
				return &Location{URI: dep.URI, Range: s.Range}
			}
		}
	}
	return nil
}

// pos_as_range 把一个位置包成零宽区间（用于比较顺序）。
func pos_as_range(p Position) Range { return Range{Start: p, End: p} }

// rangeStartsAfter 判断 a 的起点是否在 b 之后。
func rangeStartsAfter(a, b Range) bool {
	if a.Start.Line != b.Start.Line {
		return a.Start.Line > b.Start.Line
	}
	return a.Start.Character > b.Start.Character
}

// typeOfLocal 猜测某个局部变量/参数的类型名（用于字段跳转）。
func typeOfLocal(info *moduleInfo, name string) string {
	if info.Check == nil {
		return ""
	}
	for d, sym := range info.Check.Syms {
		if d != nil && sym != nil && d.Name == name {
			return typeShortName(sym.Type)
		}
	}
	return ""
}

// typeShortName 取类型名的最后一段（`_.sub.Type` → `Type`）。
func typeShortName(t any) string {
	s := "?"
	if tt, ok := t.(types.Type); ok {
		s = types.TypeName(tt)
	}
	if i := strings.LastIndex(s, "."); i >= 0 {
		s = s[i+1:]
	}
	return s
}

// Hover 返回悬停信息（种类 + 名字 + 类型）。
func (w *Workspace) Hover(uri string, pos Position) *Hover {
	path := URI2Path(uri)
	a := w.analysisFor(path)
	info := a.ByPath[path]
	if info == nil {
		return nil
	}
	word, wordRange, ok := identifierAt(info.Text, pos)
	if !ok {
		return nil
	}
	kind, detail, mod := w.describe(info, a, pos, word)
	if kind == "" {
		return nil
	}
	value := "**" + kind + "** `" + word + "`"
	if detail != "" {
		value += "\n\n```aila\n" + detail + "\n```"
	}
	if mod != "" && mod != info.Name {
		value += "\n\n来自模块 `" + mod + "`"
	}
	return &Hover{Contents: MarkupContent{Kind: "markdown", Value: value}, Range: &wordRange}
}

// describe 生成 hover 的（种类，明细，来源模块）。
func (w *Workspace) describe(info *moduleInfo, a *analysis, pos Position, word string) (string, string, string) {
	kindName := func(k int) string {
		switch k {
		case KindFunction:
			return "函数"
		case KindMethod:
			return "方法"
		case KindStruct:
			return "struct"
		case KindEnum:
			return "enum"
		case KindEnumMember:
			return "枚举变体"
		case KindField:
			return "字段"
		case KindConstant:
			return "常量"
		case KindVariable:
			return "变量"
		case KindTypeParam:
			return "类型别名"
		}
		return "符号"
	}
	loc := w.resolve(uriOf(info), pos)
	if loc != nil {
		if target := a.ByPath[URI2Path(loc.URI)]; target != nil {
			if s := symbolAt(target, loc.Range.Start); s != nil {
				detail := s.Detail
				if detail == "" {
					detail = s.Name
				}
				return kindName(s.Kind), detail, target.Name
			}
		}
	}
	// 兜底：光标处本身就是声明
	if s := symbolAt(info, pos); s != nil {
		return kindName(s.Kind), s.Detail, info.Name
	}
	return "", "", ""
}

func uriOf(info *moduleInfo) string { return info.URI }

// References 返回工程内所有引用（按名字匹配；声明也算）。
func (w *Workspace) References(uri string, pos Position) []Location {
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
	var out []Location
	for _, name := range a.Order {
		mod := a.Modules[name]
		if mod == nil {
			continue
		}
		out = append(out, occurrences(mod, word)...)
	}
	return out
}

// Highlights 返回同文件内的所有出现位置。
func (w *Workspace) Highlights(uri string, pos Position) []map[string]any {
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
	var out []map[string]any
	for _, loc := range occurrences(info, word) {
		out = append(out, map[string]any{"range": loc.Range, "kind": 1})
	}
	return out
}

// WorkspaceSymbols 在整工程里按名字模糊搜索。
func (w *Workspace) WorkspaceSymbols(query string) []SymbolInformation {
	var out []SymbolInformation
	roots := []string{w.root}
	if w.root == "" {
		return nil
	}
	a := w.analysisFor(w.root)
	if a == nil {
		return nil
	}
	q := strings.ToLower(strings.TrimSpace(query))
	for _, name := range a.Order {
		mod := a.Modules[name]
		if mod == nil {
			continue
		}
		for _, s := range mod.Symbols {
			if s.Local {
				continue
			}
			if q != "" && !strings.Contains(strings.ToLower(s.Name), q) {
				continue
			}
			out = append(out, SymbolInformation{
				Name: s.Name, Kind: s.Kind,
				Location: Location{URI: mod.URI, Range: s.Range},
			})
		}
	}
	_ = roots
	return out
}

// occurrences 在一份文本里找某个标识符的全部出现（按行扫描）。
func occurrences(info *moduleInfo, word string) []Location {
	var out []Location
	if word == "" {
		return nil
	}
	lines := strings.Split(info.Text, "\n")
	for i, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		start := 0
		for {
			idx := strings.Index(line[start:], word)
			if idx < 0 {
				break
			}
			at := start + idx
			// 词边界检查
			leftOK := at == 0 || !isIdentByte(line[at-1])
			end := at + len(word)
			rightOK := end >= len(line) || !isIdentByte(line[end])
			if leftOK && rightOK {
				out = append(out, Location{URI: info.URI, Range: Range{
					Start: Position{Line: i, Character: at},
					End:   Position{Line: i, Character: end},
				}})
			}
			start = at + len(word)
			if start >= len(line) {
				break
			}
		}
	}
	return out
}

func isIdentByte(b byte) bool {
	return b == '_' || b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

// Diagnostics 返回某个文件当前的诊断。
func (w *Workspace) Diagnostics(uri string) []Diagnostic {
	w.invalidate(URI2Path(uri))
	path := URI2Path(uri)
	a := w.analysisFor(path)
	if info := a.ByPath[path]; info != nil {
		out := make([]Diagnostic, 0, len(info.Diags))
		out = append(out, info.Diags...)
		return out
	}
	return []Diagnostic{}
}

// OtherDiagnostics 返回同工程里其它模块的诊断（编辑器里能一起标红）。
func (w *Workspace) OtherDiagnostics(uri string) map[string][]Diagnostic {
	path := URI2Path(uri)
	a := w.analysisFor(path)
	out := map[string][]Diagnostic{}
	for _, name := range a.Order {
		info := a.Modules[name]
		if info == nil || info.URI == uri || len(info.Diags) == 0 {
			continue
		}
		out[info.URI] = info.Diags
	}
	return out
}
