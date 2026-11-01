package lsp

import (
	"sort"
	"strings"

	"aila/internal/ast"
	"aila/internal/lexer"
	"aila/internal/sema"
	"aila/internal/types"
)

// 本文件实现「编辑器基础能力」：补全 / 签名帮助 / 重命名 / 折叠 / 选区 /
// 格式化 / 内联提示。语义一律取自编译器的解析结果（AST + 签名表），
// 只有「光标处到底在哪」这类纯文本判断才做字符扫描。

// --------------------------------------------------------------------------- //
// 通用文本工具
// --------------------------------------------------------------------------- //

// scanCode 把注释与字符串/字符字面量原地抹成空格（长度与换行保持不变），
// 这样扫描括号时不会被字面量里的括号带跑偏。
// 第二个返回值标记「该行开头处在多行字符串 / 块注释内部」——格式化要跳过这些行。
func scanCode(text string) (string, []bool) {
	nLines := strings.Count(text, "\n") + 1
	startInside := make([]bool, nLines)
	buf := []byte(text)
	line := 0
	const (
		stCode = iota
		stLineComment
		stBlockComment
		stString
		stRawString
		stChar
	)
	state := stCode
	for i := 0; i < len(text); i++ {
		c := text[i]
		if c == '\n' {
			if state == stLineComment {
				state = stCode
			}
			line++
			if line < nLines {
				startInside[line] = state != stCode
			}
			continue
		}
		switch state {
		case stCode:
			switch {
			case c == '/' && i+1 < len(text) && text[i+1] == '/':
				buf[i], buf[i+1] = ' ', ' '
				i++
				state = stLineComment
			case c == '/' && i+1 < len(text) && text[i+1] == '*':
				buf[i], buf[i+1] = ' ', ' '
				i++
				state = stBlockComment
			case c == 'r' && i+1 < len(text) && text[i+1] == '"':
				buf[i], buf[i+1] = ' ', ' '
				i++
				state = stRawString
			case c == '"':
				buf[i] = ' '
				state = stString
			case c == '\'':
				buf[i] = ' '
				state = stChar
			}
		case stLineComment:
			buf[i] = ' '
		case stBlockComment:
			if c == '*' && i+1 < len(text) && text[i+1] == '/' {
				buf[i], buf[i+1] = ' ', ' '
				i++
				state = stCode
			} else {
				buf[i] = ' '
			}
		case stString:
			if c == '\\' && i+1 < len(text) {
				buf[i], buf[i+1] = ' ', ' '
				i++
			} else if c == '"' {
				buf[i] = ' '
				state = stCode
			} else {
				buf[i] = ' '
			}
		case stRawString:
			if c == '"' {
				buf[i] = ' '
				state = stCode
			} else {
				buf[i] = ' '
			}
		case stChar:
			if c == '\\' && i+1 < len(text) {
				buf[i], buf[i+1] = ' ', ' '
				i++
			} else if c == '\'' {
				buf[i] = ' '
				state = stCode
			} else {
				buf[i] = ' '
			}
		}
	}
	return string(buf), startInside
}

// offsetOf 把零基行列换算成字节偏移（越界时夹到合法范围）。
func offsetOf(text string, pos Position) int {
	if pos.Line <= 0 && pos.Character <= 0 {
		return 0
	}
	off := 0
	line := 0
	for line < pos.Line {
		idx := strings.IndexByte(text[off:], '\n')
		if idx < 0 {
			return len(text)
		}
		off += idx + 1
		line++
	}
	off += pos.Character
	if off > len(text) {
		return len(text)
	}
	if off < 0 {
		return 0
	}
	return off
}

// lineContents 返回文档的每一行（去掉 \r）。
func lineContents(text string) []string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, "\r")
	}
	return lines
}

// isIdentWord 判断字符串是不是一个合法标识符。
func isIdentWord(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		letter := c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
		if i == 0 && !letter {
			return false
		}
		if !letter && !(c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

// isReservedWord 判断是不是关键字或内置类型名（这些不能被重命名/当变量补全）。
func isReservedWord(s string) bool {
	return lexer.Keywords[s] || lexer.TypeWords[s]
}

// completionKindOf 把 SymbolKind 映射成 CompletionItemKind。
func completionKindOf(k int) int {
	switch k {
	case KindFunction:
		return CIKindFunction
	case KindMethod:
		return CIKindMethod
	case KindStruct:
		return CIKindStruct
	case KindEnum:
		return CIKindEnum
	case KindEnumMember:
		return CIKindEnumMember
	case KindField:
		return CIKindField
	case KindConstant:
		return CIKindConstant
	case KindVariable:
		return CIKindVariable
	case KindModule:
		return CIKindModule
	case KindTypeParam:
		return CIKindTypeParameter
	}
	return CIKindText
}

// --------------------------------------------------------------------------- //
// 补全
// --------------------------------------------------------------------------- //

// Completion 返回光标处的补全候选。
func (w *Workspace) Completion(uri string, pos Position) CompletionList {
	out := CompletionList{}
	path := URI2Path(uri)
	a := w.analysisFor(path)
	info := a.ByPath[path]
	if info == nil || info.Mod == nil {
		return out
	}
	lines := lineContents(info.Text)
	if pos.Line < 0 || pos.Line >= len(lines) {
		return out
	}
	line := lines[pos.Line]
	col := pos.Character
	if col < 0 {
		col = 0
	}
	if col > len(line) {
		col = len(line)
	}
	before := line[:col]

	// ① `qual.` / `qual.part` → 成员
	if qual, ok := memberQualifier(before); ok {
		out.Items = w.memberItems(a, info, qual)
		return out
	}
	// ② `#[` → 属性名
	if i := strings.LastIndex(before, "#["); i >= 0 && !strings.Contains(before[i:], "]") {
		out.Items = []CompletionItem{{
			Label:  "c",
			Kind:   CIKindKeyword,
			Detail: "#[c]：声明这是 C 侧符号（函数 / 常量），由 cc 链接",
		}}
		return out
	}
	// ③ `@` → 效果 / 区域
	if i := strings.LastIndex(before, "@"); i >= 0 && i == len(before)-1 {
		out.Items = append(out.Items,
			CompletionItem{Label: "static", Kind: CIKindKeyword, Detail: "@static：分配在静态区"},
			CompletionItem{Label: "heap", Kind: CIKindKeyword, Detail: "@heap：分配在堆区"},
			CompletionItem{Label: "region", Kind: CIKindKeyword, Detail: "@region[x]：把参数登记为区域效果"},
			CompletionItem{Label: "mut", Kind: CIKindKeyword, Detail: "@mut[x]：可变借用"},
			CompletionItem{Label: "move", Kind: CIKindKeyword, Detail: "@move[x]：移入区域"},
		)
		return out
	}
	// ④ `import ` 之后 → 模块路径
	if ref := strings.TrimSpace(before); strings.HasPrefix(ref, "import ") {
		out.Items = modulePathItems(a, strings.TrimSpace(strings.TrimPrefix(ref, "import ")))
		return out
	}

	// ⑤ 普通位置：局部 → 模块符号 → 导入别名 → 关键字 → 类型
	for _, s := range visibleLocals(info, pos) {
		out.Items = append(out.Items, symbolItem(info, s, "0"))
	}
	if inMethod(info, pos) {
		out.Items = append(out.Items, CompletionItem{
			Label: "self", Kind: CIKindVariable, Detail: "当前类型的 self", SortText: "0",
		})
	}
	for _, s := range moduleLevelSymbols(info) {
		out.Items = append(out.Items, symbolItem(info, s, "1"))
	}
	for alias := range info.Imports {
		out.Items = append(out.Items, CompletionItem{
			Label: alias, Kind: CIKindModule, Detail: "导入的模块 " + info.Imports[alias], SortText: "2",
		})
	}
	for _, kw := range sortedKeys(lexer.Keywords) {
		out.Items = append(out.Items, CompletionItem{Label: kw, Kind: CIKindKeyword, SortText: "8"})
	}
	for _, tw := range sortedKeys(lexer.TypeWords) {
		out.Items = append(out.Items, CompletionItem{
			Label: tw, Kind: CIKindKeyword, Detail: "内置类型", SortText: "7",
		})
	}
	return out
}

// memberQualifier 判断光标是否处于 `qual.partial` 的成员位，返回 qual。
func memberQualifier(before string) (string, bool) {
	i := len(before)
	for i > 0 && isIdentByte(before[i-1]) {
		i--
	}
	if i == 0 || before[i-1] != '.' {
		return "", false
	}
	j := i - 1
	for j > 0 && isIdentByte(before[j-1]) {
		j--
	}
	if j == i-1 {
		return "", false
	}
	qual := before[j : i-1]
	if !isIdentWord(qual) {
		return "", false // `1.5` 之类不是成员访问
	}
	return qual, true
}

// memberItems 返回 `qual.` 后面的候选。
func (w *Workspace) memberItems(a *analysis, info *moduleInfo, qual string) []CompletionItem {
	var out []CompletionItem
	// 导入别名 → 目标模块的公开符号
	if target, ok := info.Imports[qual]; ok {
		if dep := a.Modules[target]; dep != nil {
			for _, s := range moduleLevelSymbols(dep) {
				if !s.Public && dep != info {
					continue
				}
				out = append(out, symbolItem(dep, s, "1"))
			}
		}
		return out
	}
	// 本模块类型名 / 导入模块的类型名 → 字段 + 方法 + 变体
	for _, owner := range typeOwners(a, info, qual) {
		out = append(out, typeMemberItems(owner.mod, owner.name)...)
	}
	if len(out) > 0 {
		return out
	}
	// 局部变量 / 全局变量的类型成员
	if tn := localTypeName(info, qual); tn != "" {
		for _, owner := range typeOwners(a, info, tn) {
			out = append(out, typeMemberItems(owner.mod, owner.name)...)
		}
	}
	return out
}

// typeOwner 是一个类型的归属模块。
type typeOwner struct {
	mod  *moduleInfo
	name string
}

// typeOwners 找某个短类型名在工程里对应的声明（本模块优先，其次导入模块）。
func typeOwners(a *analysis, info *moduleInfo, short string) []typeOwner {
	if short == "" {
		return nil
	}
	var out []typeOwner
	has := func() bool {
		for _, s := range info.Symbols {
			if !s.Local && s.Name == short && (s.Kind == KindStruct || s.Kind == KindEnum || s.Kind == KindTypeParam) {
				return true
			}
		}
		return false
	}
	if has() {
		return append(out, typeOwner{mod: info, name: short})
	}
	// 导入模块里的同名类型
	for _, target := range info.Imports {
		dep := a.Modules[target]
		if dep == nil {
			continue
		}
		for _, s := range dep.Symbols {
			if !s.Local && s.Name == short && (s.Kind == KindStruct || s.Kind == KindEnum || s.Kind == KindTypeParam) {
				out = append(out, typeOwner{mod: dep, name: short})
				break
			}
		}
	}
	return out
}

// typeMemberItems 返回某个类型的成员候选（字段 / 方法 / 枚举变体）。
func typeMemberItems(mod *moduleInfo, typeName string) []CompletionItem {
	var out []CompletionItem
	for _, s := range mod.Symbols {
		if s.Local {
			continue
		}
		switch {
		case s.Kind == KindField && s.Container == typeName:
			out = append(out, symbolItem(mod, s, "0"))
		case s.Kind == KindMethod && s.Container == typeName:
			out = append(out, symbolItem(mod, s, "1"))
		case s.Kind == KindEnumMember && s.Container == typeName:
			short := strings.TrimPrefix(s.Name, typeName+".")
			out = append(out, CompletionItem{
				Label: short, Kind: CIKindEnumMember, Detail: "变体 " + s.Name, SortText: "0",
			})
		}
	}
	return out
}

// moduleLevelSymbols 返回模块级（非局部）符号，跳过方法/字段/变体这类成员。
func moduleLevelSymbols(info *moduleInfo) []*symbol {
	var out []*symbol
	for _, s := range info.Symbols {
		if s.Local || s.Container != "" {
			continue
		}
		switch s.Kind {
		case KindFunction, KindStruct, KindEnum, KindConstant, KindVariable, KindTypeParam:
			out = append(out, s)
		}
	}
	return out
}

// symbolItem 把索引符号转成补全项（函数带参数占位）。
func symbolItem(info *moduleInfo, s *symbol, sortPrefix string) CompletionItem {
	item := CompletionItem{
		Label: s.Name, Kind: completionKindOf(s.Kind), Detail: s.Detail, SortText: sortPrefix + s.Name,
	}
	// 文档浮层与 hover 同款：种类 + 签名 + 来源，补全列表里就能看到，不用悬停
	value := "**" + kindNameOf(s.Kind) + "** `" + s.Name + "`"
	if s.Detail != "" {
		value += "\n\n```aila\n" + s.Detail + "\n```"
	}
	if info.Generated {
		value += "\n\n来自 C 包 `" + info.Cpack + "` 的绑定"
	} else {
		value += "\n\n来自模块 `" + info.Name + "`"
	}
	item.Documentation = &MarkupContent{Kind: "markdown", Value: value}
	if s.Kind == KindMethod {
		if i := strings.LastIndex(s.Name, "."); i >= 0 {
			item.Label = s.Name[i+1:]
			item.SortText = sortPrefix + item.Label
			item.FilterText = item.Label
		}
	}
	if s.Kind == KindFunction || s.Kind == KindMethod {
		if snippet, ok := callSnippet(info, s.Name); ok {
			item.InsertText = snippet
			item.InsertTextFormat = 2
		}
	}
	return item
}

// callSnippet 生成 `name(${1:a}, ${2:b})` 形式的片段。
func callSnippet(info *moduleInfo, fullName string) (string, bool) {
	sig := sigByName(info, fullName)
	if sig == nil {
		return "", false
	}
	short := fullName
	if i := strings.LastIndex(short, "."); i >= 0 && sig.Owner != "" {
		short = short[i+1:]
	}
	var b strings.Builder
	b.WriteString(short)
	b.WriteString("(")
	for i, p := range sig.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		if isIdentWord(p.Name) && p.Name != "self" {
			b.WriteString("${" + itoa(i+1) + ":" + p.Name + "}")
		}
	}
	b.WriteString(")")
	return b.String(), true
}

// itoa 是 strconv.Itoa 的小包装（避免为两处调用引入 import 噪音）。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// visibleLocals 返回声明在 pos 之前的局部符号。
func visibleLocals(info *moduleInfo, pos Position) []*symbol {
	var out []*symbol
	seen := map[string]bool{}
	for _, s := range info.Symbols {
		if !s.Local || s.Name == "" || seen[s.Name] {
			continue
		}
		if rangeStartsAfter(s.Range, pos_as_range(pos)) {
			continue
		}
		seen[s.Name] = true
		out = append(out, s)
	}
	return out
}

// inMethod 判断光标是否在带 self 的方法里。
func inMethod(info *moduleInfo, pos Position) bool {
	for _, d := range info.Mod.AllDecls() {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.OwnerType == "" || fn.Body == nil {
			continue
		}
		if within(spanRange(fn.Body.Span), pos) {
			return true
		}
	}
	return false
}

// modulePathItems 给出 `import ` 之后可选的模块名。
func modulePathItems(a *analysis, prefix string) []CompletionItem {
	var out []CompletionItem
	seen := map[string]bool{}
	add := func(name, detail string) {
		if name == "" || seen[name] {
			return
		}
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			return
		}
		seen[name] = true
		out = append(out, CompletionItem{Label: name, Kind: CIKindModule, Detail: detail})
	}
	add("std", "标准库")
	add("ch", "C 侧")
	add("std.io", "标准库：输入输出")
	add("std.math", "标准库：数学")
	add("ch.cstdio", "C 侧：stdio")
	for _, name := range a.Order {
		add(name, "本工程模块")
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}

// sortedKeys 返回 map 的键（升序）。
func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// --------------------------------------------------------------------------- //
// 签名帮助
// --------------------------------------------------------------------------- //

// SignatureHelp 返回光标所在调用处的签名信息。
func (w *Workspace) SignatureHelp(uri string, pos Position) *SignatureHelp {
	path := URI2Path(uri)
	a := w.analysisFor(path)
	info := a.ByPath[path]
	if info == nil || info.Mod == nil {
		return nil
	}
	code, _ := scanCode(info.Text)
	off := offsetOf(info.Text, pos)
	open, active, ok := enclosingCall(code, off)
	if !ok {
		return nil
	}
	qual, name := calleeBefore(code, open)
	if name == "" {
		return nil
	}
	sig := w.sigAt(a, info, qual, name)
	if sig == nil {
		return nil
	}
	params := make([]ParameterInformation, 0, len(sig.Params))
	for _, p := range sig.Params {
		params = append(params, ParameterInformation{Label: p.Name + ": " + types.TypeName(p.Type)})
	}
	ap := active
	if ap > len(params)-1 {
		ap = len(params) - 1
	}
	if ap < 0 {
		ap = 0
	}
	return &SignatureHelp{
		Signatures: []SignatureInformation{{
			Label:      sigText(sig),
			Parameters: params,
		}},
		ActiveParameter: ap,
	}
}

// enclosingCall 从 off 往前找包住光标的 `(`，并数出当前是第几个实参。
func enclosingCall(code string, off int) (int, int, bool) {
	if off > len(code) {
		off = len(code)
	}
	depth := 0
	for i := off - 1; i >= 0; i-- {
		switch code[i] {
		case ')', ']', '}':
			depth++
		case '(', '[', '{':
			if depth > 0 {
				depth--
				continue
			}
			if code[i] != '(' {
				return 0, 0, false
			}
			return i, countArgs(code, i+1, off), true
		}
	}
	return 0, 0, false
}

// countArgs 数出 [start,end) 里深度为 0 的逗号个数。
func countArgs(code string, start, end int) int {
	depth := 0
	n := 0
	for i := start; i < end && i < len(code); i++ {
		switch code[i] {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				n++
			}
		}
	}
	return n
}

// calleeBefore 取 `(` 前面的限定名（`a.b(` → ("a", "b")）。
func calleeBefore(code string, open int) (string, string) {
	i := open
	for i > 0 && (code[i-1] == ' ' || code[i-1] == '\t') {
		i--
	}
	end := i
	for i > 0 && isIdentByte(code[i-1]) {
		i--
	}
	name := code[i:end]
	if !isIdentWord(name) {
		return "", ""
	}
	if i > 0 && code[i-1] == '.' {
		j := i - 1
		for j > 0 && isIdentByte(code[j-1]) {
			j--
		}
		return code[j : i-1], name
	}
	return "", name
}

// sigAt 按「限定名 + 名字」解析签名。
func (w *Workspace) sigAt(a *analysis, info *moduleInfo, qual, name string) *sema.FuncSig {
	if qual == "" {
		return sigByName(info, name)
	}
	if target, ok := info.Imports[qual]; ok {
		if dep := a.Modules[target]; dep != nil {
			return sigByName(dep, name)
		}
		return nil
	}
	// `Type.method` 或 `recv.method`
	if sig := sigByName(info, qual+"."+name); sig != nil {
		return sig
	}
	if tn := localTypeName(info, qual); tn != "" {
		if sig := sigByName(info, tn+"."+name); sig != nil {
			return sig
		}
	}
	for _, owner := range typeOwners(a, info, qual) {
		if sig := sigByName(owner.mod, owner.name+"."+name); sig != nil {
			return sig
		}
	}
	return nil
}

// sigByName 在模块里按名字（或 `Owner.name`）找签名。
func sigByName(info *moduleInfo, name string) *sema.FuncSig {
	if info == nil || info.Sigs == nil {
		return nil
	}
	if s, ok := info.Sigs.Funcs[name]; ok {
		return s
	}
	if s, ok := info.Sigs.Methods[name]; ok {
		return s
	}
	return nil
}

// localTypeName 猜测变量的类型短名（局部符号 → 全局 → 常量）。
func localTypeName(info *moduleInfo, name string) string {
	if t := typeOfLocal(info, name); t != "" {
		return t
	}
	if info.Sigs != nil {
		if g, ok := info.Sigs.Globals[name]; ok && g.Type != nil {
			return typeShortName(g.Type)
		}
		if c, ok := info.Sigs.Consts[name]; ok && c.Type != nil {
			return typeShortName(c.Type)
		}
	}
	return ""
}

// --------------------------------------------------------------------------- //
// 重命名
// --------------------------------------------------------------------------- //

// PrepareRename 判断光标处能不能改名，能则返回原名的区间。
func (w *Workspace) PrepareRename(uri string, pos Position) *PrepareRenameResult {
	path := URI2Path(uri)
	a := w.analysisFor(path)
	info := a.ByPath[path]
	if info == nil {
		return nil
	}
	word, r, ok := identifierAt(info.Text, pos)
	if !ok || isReservedWord(word) {
		return nil
	}
	if info.Generated {
		return nil // 生成的 C 绑定不许改名
	}
	if kind, _, _ := w.describe(info, a, pos, word); kind == "" {
		return nil // 未解析的符号（比如语法错误处）不给改名
	}
	return &PrepareRenameResult{Range: r, Placeholder: word}
}

// Rename 返回重命名产生的跨文件编辑。
func (w *Workspace) Rename(uri string, pos Position, newName string) *WorkspaceEdit {
	if !isIdentWord(newName) || isReservedWord(newName) {
		return nil
	}
	path := URI2Path(uri)
	a := w.analysisFor(path)
	info := a.ByPath[path]
	if info == nil {
		return nil
	}
	word, _, ok := identifierAt(info.Text, pos)
	if !ok || isReservedWord(word) || word == newName {
		return nil
	}
	if info.Generated {
		return nil // 生成的 C 绑定不许改名
	}
	// 目标符号：先按跳转解析，再取声明处的符号种类
	target := info
	local := false
	member := false
	if loc := w.resolve(uri, pos); loc != nil {
		if declMod := a.ByPath[URI2Path(loc.URI)]; declMod != nil {
			target = declMod
			if s := symbolAt(target, loc.Range.Start); s != nil {
				local = s.Local
				member = s.Kind == KindField || s.Kind == KindEnumMember || s.Kind == KindMethod
			}
		}
	}
	changes := map[string][]TextEdit{}
	for _, name := range a.Order {
		mod := a.Modules[name]
		if mod == nil {
			continue
		}
		if local && mod != target {
			continue // 局部符号只在声明模块内改名
		}
		lines := lineContents(mod.Text)
		for _, occ := range occurrences(mod, word) {
			if qual := qualifierAtLine(lines, occ.Range.Start); qual != "" {
				if local || !member {
					if !qualifierRefersTo(qual, mod, target) {
						continue
					}
				}
			}
			changes[mod.URI] = append(changes[mod.URI], TextEdit{Range: occ.Range, NewText: newName})
		}
	}
	if len(changes) == 0 {
		return nil
	}
	return &WorkspaceEdit{Changes: changes}
}

// qualifierAtLine 取某位置前面的 `qual.`（没有则空）。
func qualifierAtLine(lines []string, pos Position) string {
	if pos.Line < 0 || pos.Line >= len(lines) {
		return ""
	}
	line := lines[pos.Line]
	if pos.Character <= 0 || pos.Character > len(line) {
		return ""
	}
	if line[pos.Character-1] != '.' {
		return ""
	}
	i := pos.Character - 1
	j := i
	for j > 0 && isIdentByte(line[j-1]) {
		j--
	}
	if j == i {
		return ""
	}
	return line[j:i]
}

// qualifierRefersTo 判断限定名是否指向目标模块（别名或最后一段路径）。
func qualifierRefersTo(qual string, mod, target *moduleInfo) bool {
	if target == nil {
		return true
	}
	if viaAlias, ok := mod.Imports[qual]; ok {
		return viaAlias == target.Name
	}
	segs := strings.Split(target.Name, ".")
	if segs[len(segs)-1] == qual {
		return true
	}
	return false
}

// --------------------------------------------------------------------------- //
// 折叠
// --------------------------------------------------------------------------- //

// FoldingRanges 返回可折叠区间（代码块 / 注释 / import 段）。
func (w *Workspace) FoldingRanges(uri string) []FoldingRange {
	path := URI2Path(uri)
	a := w.analysisFor(path)
	info := a.ByPath[path]
	if info == nil {
		return nil
	}
	text := info.Text
	code, _ := scanCode(text)
	codeLines := strings.Split(code, "\n")
	origLines := lineContents(text)
	var out []FoldingRange

	// ① 花括号块
	type open struct{ line, col int }
	var stack []open
	for i, ln := range codeLines {
		for col := 0; col < len(ln); col++ {
			switch ln[col] {
			case '{':
				stack = append(stack, open{i, col})
			case '}':
				if len(stack) == 0 {
					continue
				}
				op := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				start := op.line
				// `{` 独占一行时，折叠从上一行（声明行）开始更像 IDE
				openLine := origLines[op.line]
				if op.col <= len(openLine) && strings.TrimSpace(openLine[:op.col]) == "" && op.line > 0 {
					start = op.line - 1
				}
				if i-start >= 2 {
					out = append(out, FoldingRange{StartLine: start, EndLine: i - 1, Kind: "region"})
				}
			}
		}
	}

	// ② 连续行注释
	runStart := -1
	flush := func(end int) {
		if runStart >= 0 && end > runStart {
			out = append(out, FoldingRange{StartLine: runStart, EndLine: end, Kind: "comment"})
		}
		runStart = -1
	}
	for i, ln := range origLines {
		if strings.HasPrefix(strings.TrimSpace(ln), "//") {
			if runStart < 0 {
				runStart = i
			}
		} else {
			if runStart >= 0 {
				flush(i - 1)
			}
		}
	}
	flush(len(origLines) - 1)

	// ③ 多行块注释
	if idx := strings.Index(text, "/*"); idx >= 0 {
		off := idx
		for off < len(text) {
			end := strings.Index(text[off+2:], "*/")
			if end < 0 {
				break
			}
			closeAt := off + 2 + end
			startLine := strings.Count(text[:off], "\n")
			endLine := strings.Count(text[:closeAt], "\n")
			if endLine-startLine >= 1 {
				out = append(out, FoldingRange{StartLine: startLine, EndLine: endLine, Kind: "comment"})
			}
			next := closeAt + 2
			rel := strings.Index(text[next:], "/*")
			if rel < 0 {
				break
			}
			off = next + rel
		}
	}

	// ④ 连续 import 段
	impStart := -1
	flushImp := func(end int) {
		if impStart >= 0 && end-impStart >= 1 {
			out = append(out, FoldingRange{StartLine: impStart, EndLine: end, Kind: "imports"})
		}
		impStart = -1
	}
	for i, ln := range origLines {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "import ") || t == "import" {
			if impStart < 0 {
				impStart = i
			}
		} else {
			if impStart >= 0 {
				flushImp(i - 1)
			}
		}
	}
	flushImp(len(origLines) - 1)

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].StartLine != out[j].StartLine {
			return out[i].StartLine < out[j].StartLine
		}
		return out[i].EndLine < out[j].EndLine
	})
	return out
}

// --------------------------------------------------------------------------- //
// 选区
// --------------------------------------------------------------------------- //

// bracketPair 是一对匹配的括号。
type bracketPair struct {
	openLine, openCol   int
	closeLine, closeCol int
}

// bracketPairs 扫描出所有匹配的括号对（注释与字面量已抹掉）。
func bracketPairs(code string) []bracketPair {
	lines := strings.Split(code, "\n")
	type entry struct {
		line, col int
		ch        byte
	}
	var stack []entry
	var out []bracketPair
	match := func(open, close byte) bool {
		return open == '(' && close == ')' || open == '[' && close == ']' || open == '{' && close == '}'
	}
	for i, ln := range lines {
		for col := 0; col < len(ln); col++ {
			c := ln[col]
			switch c {
			case '(', '[', '{':
				stack = append(stack, entry{i, col, c})
			case ')', ']', '}':
				for k := len(stack) - 1; k >= 0; k-- {
					if match(stack[k].ch, c) {
						e := stack[k]
						stack = append(stack[:k], stack[k+1:]...)
						out = append(out, bracketPair{e.line, e.col, i, col})
						break
					}
				}
			}
		}
	}
	return out
}

// SelectionRanges 返回每个位置「由小到大」的选区链。
func (w *Workspace) SelectionRanges(uri string, positions []Position) []SelectionRange {
	path := URI2Path(uri)
	a := w.analysisFor(path)
	info := a.ByPath[path]
	if info == nil {
		return nil
	}
	code, _ := scanCode(info.Text)
	pairs := bracketPairs(code)
	lines := lineContents(info.Text)
	docEnd := Position{Line: len(lines) - 1, Character: len(lines[len(lines)-1])}
	out := make([]SelectionRange, 0, len(positions))
	for _, pos := range positions {
		var chain []Range
		if _, r, ok := identifierAt(info.Text, pos); ok {
			chain = append(chain, r)
		}
		// 由内到外的括号对
		var inside []bracketPair
		for _, p := range pairs {
			open := Position{Line: p.openLine, Character: p.openCol}
			close := Position{Line: p.closeLine, Character: p.closeCol}
			if posBefore(open, pos) && posBefore(pos, close) {
				inside = append(inside, p)
			}
		}
		sort.SliceStable(inside, func(i, j int) bool {
			return posBefore(Position{Line: inside[j].openLine, Character: inside[j].openCol},
				Position{Line: inside[i].openLine, Character: inside[i].openCol})
		})
		for _, p := range inside {
			chain = append(chain, Range{
				Start: Position{Line: p.openLine, Character: p.openCol},
				End:   Position{Line: p.closeLine, Character: p.closeCol + 1},
			})
		}
		if pos.Line >= 0 && pos.Line < len(lines) {
			chain = append(chain, Range{
				Start: Position{Line: pos.Line, Character: 0},
				End:   Position{Line: pos.Line, Character: len(lines[pos.Line])},
			})
		}
		chain = append(chain, Range{Start: Position{}, End: docEnd})
		var node *SelectionRange
		for i := len(chain) - 1; i >= 0; i-- {
			node = &SelectionRange{Range: chain[i], Parent: node}
		}
		out = append(out, *node)
	}
	return out
}

// posBefore 判断 a 是否严格早于 b。
func posBefore(a, b Position) bool {
	if a.Line != b.Line {
		return a.Line < b.Line
	}
	return a.Character < b.Character
}

// --------------------------------------------------------------------------- //
// 格式化
// --------------------------------------------------------------------------- //

// Format 返回格式化编辑（缩进按 4 空格、去行尾空白、末尾恰好一个换行）。
func (w *Workspace) Format(uri string, opts FormattingOptions) []TextEdit {
	path := URI2Path(uri)
	a := w.analysisFor(path)
	info := a.ByPath[path]
	if info == nil {
		return nil
	}
	width := opts.TabSize
	if width <= 0 || width > 8 {
		width = 4
	}
	if info.Generated {
		return []TextEdit{} // 生成的 C 绑定不格式化（改了也会被下次生成覆盖）
	}
	formatted := formatText(info.Text, width)
	if formatted == info.Text {
		return []TextEdit{} // 已经格式化了：返回空数组（不是 null）
	}
	lines := lineContents(info.Text)
	end := Position{Line: len(lines) - 1, Character: len(lines[len(lines)-1])}
	return []TextEdit{{Range: Range{Start: Position{}, End: end}, NewText: formatted}}
}

// formatText 是格式化的纯函数（幂等）。
func formatText(text string, width int) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	eol := "\n"
	if strings.Contains(text, "\r\n") {
		eol = "\r\n"
	}
	code, startInside := scanCode(text)
	codeLines := strings.Split(code, "\n")
	origLines := strings.Split(text, "\n")
	depth := 0
	out := make([]string, 0, len(origLines))
	for i, raw := range origLines {
		line := strings.TrimRight(raw, "\r")
		body := strings.TrimSpace(line)
		if body == "" {
			out = append(out, "")
			if i < len(codeLines) {
				depth = adjustDepth(depth, codeLines[i])
			}
			continue
		}
		if startInside[i] {
			// 多行字符串 / 块注释内部：原样保留（改缩进会改语义）
			out = append(out, line)
			if i < len(codeLines) {
				depth = adjustDepth(depth, codeLines[i])
			}
			continue
		}
		myDepth := depth
		if codeLines[i] != "" {
			if c := firstCodeChar(codeLines[i]); c == '}' || c == ')' || c == ']' {
				if myDepth > 0 {
					myDepth--
				}
			}
		}
		out = append(out, strings.Repeat(" ", myDepth*width)+body)
		if i < len(codeLines) {
			depth = adjustDepth(depth, codeLines[i])
		}
	}
	// 去掉末尾空行，只留一个换行
	for len(out) > 0 && strings.TrimSpace(out[len(out)-1]) == "" {
		out = out[:len(out)-1]
	}
	return strings.Join(out, eol) + eol
}

// adjustDepth 按一行的花括号增减深度。
func adjustDepth(depth int, codeLine string) int {
	for i := 0; i < len(codeLine); i++ {
		switch codeLine[i] {
		case '{':
			depth++
		case '}':
			if depth > 0 {
				depth--
			}
		}
	}
	return depth
}

// firstCodeChar 返回一行里第一个非空白字符。
func firstCodeChar(codeLine string) byte {
	for i := 0; i < len(codeLine); i++ {
		if codeLine[i] != ' ' && codeLine[i] != '\t' {
			return codeLine[i]
		}
	}
	return 0
}

// --------------------------------------------------------------------------- //
// 内联提示
// --------------------------------------------------------------------------- //

// InlayHints 返回区间内 `let` / `var` 推断出的类型提示。
func (w *Workspace) InlayHints(uri string, r Range) []InlayHint {
	path := URI2Path(uri)
	a := w.analysisFor(path)
	info := a.ByPath[path]
	if info == nil || info.Check == nil {
		return nil
	}
	lines := lineContents(info.Text)
	var out []InlayHint
	for d, sym := range info.Check.Syms {
		if d == nil || sym == nil || d.Type != nil || sym.Type == nil {
			continue // 显式标注了类型就不用提示
		}
		if !d.HasInit {
			continue
		}
		line := d.Span.Line - 1
		col := d.Span.Col - 1
		if line < 0 || line >= len(lines) {
			continue
		}
		text := lines[line]
		if col < 0 || col > len(text) {
			col = 0
		}
		idx := strings.Index(text[col:], d.Name)
		if idx < 0 {
			continue
		}
		at := Position{Line: line, Character: col + idx + len(d.Name)}
		if !within(r, at) {
			continue
		}
		out = append(out, InlayHint{
			Position: at,
			Label:    ": " + types.TypeName(sym.Type),
			Kind:     1,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return posBefore(out[i].Position, out[j].Position)
	})
	return out
}
