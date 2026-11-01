package bindgen

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"aila/internal/diag"
)

// --------------------------------------------------------------------------- //
// C 词法
// --------------------------------------------------------------------------- //

// CTok 是一个 C token。
type CTok struct {
	Kind string // ws / comment / str / num / id / punct（注释与空白已丢弃）
	Text string
	Line int
}

// TokenizeC 把 C 源文本切成 token 流（注释与空白丢弃，保留行号）。
func TokenizeC(text string) []CTok {
	toks := []CTok{}
	line := 1
	i := 0
	n := len(text)
	isIdentStart := func(c byte) bool {
		return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
	}
	isIdent := func(c byte) bool { return isIdentStart(c) || (c >= '0' && c <= '9') }
	isDigit := func(c byte) bool { return c >= '0' && c <= '9' }
	for i < n {
		c := text[i]
		// 空白
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			if c == '\n' {
				line++
			}
			i++
			continue
		}
		// 注释
		if c == '/' && i+1 < n && text[i+1] == '/' {
			for i < n && text[i] != '\n' {
				i++
			}
			continue
		}
		if c == '/' && i+1 < n && text[i+1] == '*' {
			i += 2
			for i+1 < n && !(text[i] == '*' && text[i+1] == '/') {
				if text[i] == '\n' {
					line++
				}
				i++
			}
			i += 2
			continue
		}
		// 字符串 / 字符字面量
		if c == '"' || c == '\'' {
			quote := c
			start := i
			i++
			for i < n {
				if text[i] == '\\' && i+1 < n {
					i += 2
					continue
				}
				if text[i] == quote {
					i++
					break
				}
				if text[i] == '\n' {
					line++
				}
				i++
			}
			toks = append(toks, CTok{Kind: "str", Text: text[start:i], Line: line})
			continue
		}
		// 数字
		if isDigit(c) || (c == '.' && i+1 < n && isDigit(text[i+1])) {
			start := i
			if c == '0' && i+1 < n && (text[i+1] == 'x' || text[i+1] == 'X') {
				i += 2
				for i < n && isHex(text[i]) {
					i++
				}
			} else {
				for i < n && isDigit(text[i]) {
					i++
				}
				if i < n && text[i] == '.' {
					i++
					for i < n && isDigit(text[i]) {
						i++
					}
				}
			}
			toks = append(toks, CTok{Kind: "num", Text: text[start:i], Line: line})
			continue
		}
		// 标识符
		if isIdentStart(c) {
			start := i
			for i < n && isIdent(text[i]) {
				i++
			}
			toks = append(toks, CTok{Kind: "id", Text: text[start:i], Line: line})
			continue
		}
		// 标点
		toks = append(toks, CTok{Kind: "punct", Text: string(c), Line: line})
		i++
	}
	return toks
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// --------------------------------------------------------------------------- //
// 轻量预处理
// --------------------------------------------------------------------------- //

// StripPreprocessor 收集对象式宏、按已定义的宏裁决 `#ifdef` / `#ifndef`。
//
//   - `#ifndef X`：X 未定义 → 走「未定义分支」（保留 body，这是 include guard 的常态）；
//   - `#ifdef X`：X 已定义 → 保留 body；
//   - `#if` / `#elif` / `#else`：不做真求值，整体按「不可判定」处理并记录提示（绝不静默降级）。
func StripPreprocessor(text string, macros map[string]string, order *[]string, notes *[]string) string {
	out := []string{}
	type frame struct {
		parent    bool
		thisOn    bool
		sawElse   bool
	}
	stack := []*frame{}
	active := true
	for _, raw := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		s := strings.TrimSpace(raw)
		if strings.HasPrefix(s, "#") {
			if name, params, body, ok := parseDefine(s); ok {
				if active && params == "" {
					if _, seen := macros[name]; !seen && order != nil {
						*order = append(*order, name)
					}
					macros[name] = strings.TrimSpace(body)
				} else if active {
					*notes = append(*notes, "函数式宏未映射："+name+"(…)（建议写 C shim，见 designs/0001 §11.7）")
				}
				continue
			}
			if kind, name, ok := parseIfdef(s); ok {
				cond := false
				if kind == "ifndef" {
					_, defined := macros[name]
					cond = !defined
				} else {
					_, defined := macros[name]
					cond = defined
				}
				f := &frame{parent: active, thisOn: active && cond}
				stack = append(stack, f)
				active = active && cond
				continue
			}
			if strings.HasPrefix(s, "#if") && !strings.HasPrefix(s, "#ifdef") && !strings.HasPrefix(s, "#ifndef") ||
				strings.HasPrefix(s, "#elif") {
				if active {
					*notes = append(*notes, "未解析条件编译："+s+"（builtin 档不做表达式求值）")
				}
				stack = append(stack, &frame{parent: active, thisOn: false})
				active = false
				continue
			}
			if strings.HasPrefix(s, "#else") {
				if len(stack) > 0 {
					f := stack[len(stack)-1]
					if f.sawElse {
						*notes = append(*notes, "未解析条件编译：重复 `#else`")
					}
					f.sawElse = true
					active = f.parent && !f.thisOn
				}
				continue
			}
			if strings.HasPrefix(s, "#endif") {
				if len(stack) > 0 {
					f := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					active = f.parent
				}
				continue
			}
			if strings.HasPrefix(s, "#undef") {
				fields := strings.Fields(strings.TrimPrefix(s, "#undef"))
				if len(fields) > 0 {
					delete(macros, fields[0])
				}
				continue
			}
			if strings.HasPrefix(s, "#include") {
				rest := strings.TrimPrefix(s, "#include")
				if i := strings.Index(rest, `"`); i >= 0 {
					if j := strings.Index(rest[i+1:], `"`); j >= 0 && active {
						*notes = append(*notes, "未展开的头文件包含："+rest[i+1:i+1+j]+
							"（把该头一起放进 cpack/include/）")
					}
				}
				continue
			}
			continue
		}
		if active {
			out = append(out, raw)
		}
	}
	return strings.Join(out, "\n")
}

// parseDefine 解析 `#define NAME[(params)] body`。
func parseDefine(s string) (name, params, body string, ok bool) {
	rest := strings.TrimSpace(strings.TrimPrefix(s, "#"))
	if !strings.HasPrefix(rest, "define") {
		return "", "", "", false
	}
	rest = strings.TrimSpace(strings.TrimPrefix(rest, "define"))
	i := 0
	for i < len(rest) && (isIdentByte(rest[i])) {
		i++
	}
	if i == 0 {
		return "", "", "", false
	}
	name = rest[:i]
	rest = rest[i:]
	if strings.HasPrefix(rest, "(") {
		if j := strings.Index(rest, ")"); j >= 0 {
			params = rest[:j+1]
			rest = rest[j+1:]
		}
	}
	return name, params, strings.TrimSpace(rest), true
}

// parseIfdef 解析 `#ifdef X` / `#ifndef X`。
func parseIfdef(s string) (kind, name string, ok bool) {
	rest := strings.TrimSpace(strings.TrimPrefix(s, "#"))
	if strings.HasPrefix(rest, "ifndef") {
		fields := strings.Fields(strings.TrimSpace(strings.TrimPrefix(rest, "ifndef")))
		if len(fields) > 0 {
			return "ifndef", fields[0], true
		}
	}
	if strings.HasPrefix(rest, "ifdef") {
		fields := strings.Fields(strings.TrimSpace(strings.TrimPrefix(rest, "ifdef")))
		if len(fields) > 0 {
			return "ifdef", fields[0], true
		}
	}
	return "", "", false
}

func isIdentByte(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// --------------------------------------------------------------------------- //
// builtin 档
// --------------------------------------------------------------------------- //

// runBuiltin 是内置 C 声明解析器（无 clang 时的兜底）。
func runBuiltin(cpack, cpackDir string, target Target, bag *diag.Bag) *Result {
	res := &Result{Backend: "builtin"}
	headers := FindHeaders(cpackDir)
	if len(headers) == 0 {
		return res
	}
	macros := map[string]string{}
	macroOrder := []string{}
	chunks := []string{}
	var headerNotes []string
	for _, header := range headers {
		data, err := os.ReadFile(header)
		if err != nil {
			continue
		}
		notes := []string{}
		chunks = append(chunks, StripPreprocessor(string(data), macros, &macroOrder, &notes))
		for _, note := range notes {
			if !contains(headerNotes, note) {
				headerNotes = append(headerNotes, note)
			}
		}
		headerNotes = append(headerNotes, "头文件："+baseName(header))
	}
	res.Notes = append(res.Notes, headerNotes...)
	// 宏常量映射（designs/0001 §11.7：能在 C 侧求值成常量 → 映射为 `const`）
	constLines := []string{}
	for _, k := range macroOrder {
		v := macros[k]
		if strings.Contains(k, "(") {
			continue
		}
		lit := macroLiteral(v)
		if lit == nil {
			if strings.TrimSpace(v) != "" {
				res.Notes = append(res.Notes, "宏无法求值，已跳过："+k+" = "+v)
			}
			continue
		}
		if strings.HasSuffix(k, "_H") {
			res.Notes = append(res.Notes, "包含守卫宏跳过："+k)
			continue
		}
		constLines = append(constLines, "public const "+ValueName(k)+": "+lit[0]+" = "+lit[1]+";")
		res.Notes = append(res.Notes, "宏定型："+k+" = "+lit[1]+" ("+lit[0]+")")
	}
	combined := strings.Join(chunks, "\n") + "\n"
	toks := TokenizeC(combined)
	var items []*item
	i, n := 0, len(toks)
	for i < n {
		tok := toks[i]
		if tok.Text == "typedef" || tok.Text == "struct" || tok.Text == "enum" || tok.Text == "union" ||
			tok.Kind == "id" {
			it, next := parseTopItem(toks, i)
			if it != nil {
				items = append(items, it)
			}
			i = next
			continue
		}
		i++
	}
	declared := map[string]string{}
	for _, it := range items {
		if it.kind == "struct" || it.kind == "enum" {
			declared[it.name] = TypeName(it.name)
		}
	}
	lines := []string{}
	for _, it := range items {
		lines = append(lines, renderItem(it, declared, target, bag, cpack)...)
	}
	res.Text = strings.Join(append(constLines, lines...), "\n")
	return res
}

// item 是一条顶层声明。
type item struct {
	kind      string // struct / union / enum / decl
	name      string
	body      []CTok
	alias     string
	isTypedef bool
	parts     []CTok
}

// parseTopItem 粗粒度顶层声明切分。
func parseTopItem(toks []CTok, i int) (*item, int) {
	n := len(toks)
	isTypedef := toks[i].Text == "typedef"
	j := i
	if isTypedef {
		j = i + 1
	}
	for j < n && toks[j].Kind == "id" && isStorageSpecifier(toks[j].Text) {
		j++
	}
	if j >= n {
		return nil, n
	}
	kw := toks[j].Text
	if kw == "struct" || kw == "union" || kw == "enum" {
		it := &item{kind: kw, isTypedef: isTypedef}
		j++
		if j < n && toks[j].Kind == "id" {
			it.name = toks[j].Text
			j++
		}
		if j < n && toks[j].Text == "{" {
			it.body, j = collectBraces(toks, j)
		}
		tail := []CTok{}
		for j < n && toks[j].Text != ";" {
			tail = append(tail, toks[j])
			j++
		}
		j++
		if len(tail) > 0 && tail[len(tail)-1].Kind == "id" {
			it.alias = tail[len(tail)-1].Text
		}
		if isTypedef && it.alias != "" {
			it.name = it.alias
		}
		return it, j
	}
	it := &item{kind: "decl", isTypedef: isTypedef}
	for j < n && toks[j].Text != ";" {
		if toks[j].Text == "{" {
			_, j = collectBraces(toks, j)
			continue
		}
		it.parts = append(it.parts, toks[j])
		j++
	}
	j++
	return it, j
}

func isStorageSpecifier(s string) bool {
	switch s {
	case "const", "static", "extern", "inline", "volatile":
		return true
	}
	return false
}

func collectBraces(toks []CTok, i int) ([]CTok, int) {
	depth := 0
	out := []CTok{}
	n := len(toks)
	for i < n {
		t := toks[i]
		if t.Text == "{" {
			depth++
			if depth == 1 {
				i++
				continue
			}
		} else if t.Text == "}" {
			depth--
			if depth == 0 {
				return out, i + 1
			}
		}
		out = append(out, t)
		i++
	}
	return out, n
}

// renderItem 把一条顶层声明渲染成 AILA 文本。
func renderItem(it *item, declared map[string]string, target Target, bag *diag.Bag, cpack string) []string {
	switch it.kind {
	case "struct", "union":
		if it.kind == "union" {
			bag.Warning("W3002", diag.Empty, "union `"+it.name+"` 未直译（AILA 类型系统永不直译 union）",
				"改用 `[U8; N]` + unsafe 偏移，或写 C shim", "")
			return nil
		}
		if len(it.body) == 0 {
			if it.name != "" {
				declared[it.name] = TypeName(it.name)
				return []string{"// opaque: struct " + it.name + "（无布局句柄型，仅经 `[" +
					TypeName(it.name) + "]?` 指针通道传递）"}
			}
			return nil
		}
		aname := TypeName(firstNonEmpty(it.name, it.alias))
		if _, ok := declared[firstNonEmpty(it.name, it.alias)]; !ok {
			declared[firstNonEmpty(it.name, it.alias)] = aname
		}
		fields := []string{}
		k := 0
		for k < len(it.body) {
			seg := []CTok{}
			for k < len(it.body) && it.body[k].Text != ";" {
				seg = append(seg, it.body[k])
				k++
			}
			k++
			if len(seg) == 0 {
				continue
			}
			var arr *int
			var fname, typeText string
			if seg[len(seg)-1].Text == "]" {
				open := -1
				for idx, t := range seg {
					if t.Text == "[" {
						open = idx
					}
				}
				inside := seg[open+1 : len(seg)-1]
				if len(inside) > 0 && inside[0].Kind == "num" {
					if v, err := strconv.ParseInt(strings.TrimRight(inside[0].Text, "uUlL"), 0, 64); err == nil {
						n := int(v)
						arr = &n
					}
				}
				if open > 0 {
					fname = seg[open-1].Text
				} else {
					fname = fmt.Sprintf("anon_%d", len(fields))
				}
				typeText = textsOf(seg[:open-1])
			} else {
				fname = seg[len(seg)-1].Text
				typeText = textsOf(seg[:len(seg)-1])
			}
			mapped := MapCType(typeText, target, declared, bag, diag.Empty, cpack)
			if arr != nil {
				mapped = "[" + strings.TrimSuffix(mapped, "?") + "; " + strconv.Itoa(*arr) + "]"
			}
			fields = append(fields, "    "+ValueName(fname)+": "+mapped+",")
		}
		return []string{"struct " + aname + " {\n" + strings.Join(fields, "\n") + "\n};"}
	case "enum":
		aname := TypeName(firstNonEmpty(it.name, it.alias))
		key := firstNonEmpty(it.name, it.alias)
		if _, ok := declared[key]; !ok {
			declared[key] = aname
		}
		entries := [][]CTok{}
		cur := []CTok{}
		depth := 0
		for _, t := range it.body {
			if t.Text == "(" {
				depth++
			} else if t.Text == ")" {
				depth--
			}
			if t.Text == "," && depth == 0 {
				entries = append(entries, cur)
				cur = []CTok{}
				continue
			}
			cur = append(cur, t)
		}
		if len(cur) > 0 {
			entries = append(entries, cur)
		}
		body := []string{}
		for _, entry := range entries {
			if len(entry) == 0 || entry[0].Kind != "id" {
				continue
			}
			vname := ValueName(entry[0].Text)
			var val *int64
			for idx, t := range entry {
				if t.Text == "=" {
					if v, ok := evalInt(textsSlice(entry[idx+1:])); ok {
						val = &v
					}
					break
				}
			}
			if val != nil {
				body = append(body, "    "+vname+" = "+strconv.FormatInt(*val, 10))
			} else {
				body = append(body, "    "+vname)
			}
		}
		return []string{"enum " + aname + " {\n" + strings.Join(body, ",\n") + ",\n};"}
	case "decl":
		if it.isTypedef {
			if len(it.parts) > 0 && it.parts[len(it.parts)-1].Kind == "id" {
				mapped := MapCType(textsOf(it.parts[:len(it.parts)-1]), target, declared, bag, diag.Empty, cpack)
				if !strings.HasPrefix(mapped, "（") {
					declared[it.parts[len(it.parts)-1].Text] = mapped
				}
			}
			return nil
		}
		texts := textsSlice(it.parts)
		if open := indexOf(texts, "("); open >= 0 {
			fname := ""
			if open > 0 {
				fname = texts[open-1]
			}
			ret := ""
			if open > 1 {
				ret = strings.Join(texts[:open-1], " ")
			}
			argsText := strings.Join(texts[open:], " ")
			retMapped := MapCType(ret, target, declared, bag, diag.Empty, cpack)
			args := []string{}
			inner := ""
			if l := strings.Index(argsText, "("); l >= 0 {
				if r := strings.LastIndex(argsText, ")"); r > l {
					inner = argsText[l+1 : r]
				}
			}
			idx := 0
			for _, raw := range strings.Split(inner, ",") {
				raw = strings.TrimSpace(raw)
				if raw == "" || raw == "void" {
					continue
				}
				if strings.Contains(raw, "...") {
					args = append(args, "...")
					continue
				}
				words := strings.Fields(raw)
				ptype := raw
				if len(words) >= 2 {
					ptype = strings.Join(words[:len(words)-1], " ")
				}
				args = append(args, fmt.Sprintf("a%d: %s", idx, MapCType(ptype, target, declared, bag, diag.Empty, cpack)))
				idx++
			}
			retPart := ""
			if !strings.HasPrefix(retMapped, "（") {
				retPart = " -> " + retMapped
			}
			// 生成的绑定就是 cpack 的对外接口：一律 public（包外可见）
			return []string{"#compile.c public func " + ValueName(fname) + "(" + strings.Join(args, ", ") + ")" + retPart + ";"}
		}
		if len(texts) >= 2 && it.parts[len(it.parts)-1].Kind == "id" {
			vname := it.parts[len(it.parts)-1].Text
			mapped := MapCType(strings.Join(texts[:len(texts)-1], " "), target, declared, bag, diag.Empty, cpack)
			if strings.HasPrefix(mapped, "（") {
				return nil
			}
			kw := "var"
			for _, w := range texts[:len(texts)-1] {
				if w == "const" {
					kw = "let"
				}
			}
			return []string{"#compile.c public " + kw + " " + ValueName(vname) + ": " + mapped + ";"}
		}
	}
	return nil
}

// macroLiteral 把宏体识别成 AILA 字面量（返回「类型, 字面量」）。
func macroLiteral(value string) []string {
	v := strings.TrimSpace(value)
	trimmed := strings.TrimRight(v, "uUlLfF")
	if isHexLiteral(trimmed) {
		if n, err := strconv.ParseInt(strings.TrimPrefix(strings.TrimPrefix(trimmed, "0x"), "0X"), 16, 64); err == nil {
			return []string{"I32", strconv.FormatInt(n, 10)}
		}
	}
	if isDecLiteral(trimmed) {
		if n, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return []string{"I32", strconv.FormatInt(n, 10)}
		}
	}
	if isFloatLiteral(trimmed) {
		return []string{"F64", trimmed}
	}
	if len(v) >= 2 && strings.HasPrefix(v, `"`) && strings.HasSuffix(v, `"`) {
		return []string{"Text", v}
	}
	return nil
}

func isHexLiteral(s string) bool {
	t := s
	if strings.HasPrefix(t, "-") {
		t = t[1:]
	}
	if !strings.HasPrefix(t, "0x") && !strings.HasPrefix(t, "0X") {
		return false
	}
	t = t[2:]
	if t == "" {
		return false
	}
	for i := 0; i < len(t); i++ {
		if !isHex(t[i]) {
			return false
		}
	}
	return true
}

func isDecLiteral(s string) bool {
	t := s
	if strings.HasPrefix(t, "-") {
		t = t[1:]
	}
	if t == "" {
		return false
	}
	for i := 0; i < len(t); i++ {
		if t[i] < '0' || t[i] > '9' {
			return false
		}
	}
	return true
}

func isFloatLiteral(s string) bool {
	t := s
	if strings.HasPrefix(t, "-") {
		t = t[1:]
	}
	if !strings.Contains(t, ".") {
		return false
	}
	dot := false
	for i := 0; i < len(t); i++ {
		if t[i] == '.' {
			if dot {
				return false
			}
			dot = true
			continue
		}
		if t[i] < '0' || t[i] > '9' {
			return false
		}
	}
	return true
}

// evalInt 只求值「字面量 + 四则 / 位运算 / 括号」这种枚举里常见的形式。
func evalInt(texts []string) (int64, bool) {
	src := strings.Join(texts, "")
	if src == "" {
		return 0, false
	}
	p := &cExprParser{src: src}
	v, ok := p.parseExpr()
	if !ok || p.pos != len(p.src) {
		return 0, false
	}
	return v, true
}

type cExprParser struct {
	src string
	pos int
}

func (p *cExprParser) parseExpr() (int64, bool) {
	left, ok := p.parseTerm()
	if !ok {
		return 0, false
	}
	for p.pos < len(p.src) {
		op := p.src[p.pos]
		if op != '+' && op != '-' {
			break
		}
		p.pos++
		right, ok := p.parseTerm()
		if !ok {
			return 0, false
		}
		if op == '+' {
			left += right
		} else {
			left -= right
		}
	}
	return left, true
}

func (p *cExprParser) parseTerm() (int64, bool) {
	left, ok := p.parseAtom()
	if !ok {
		return 0, false
	}
	for p.pos < len(p.src) {
		op := p.src[p.pos]
		if op != '*' && op != '/' && op != '%' && op != '<' && op != '>' && op != '&' && op != '|' && op != '^' {
			break
		}
		p.pos++
		// 跳过移位 / 位运算的双字符形态
		double := op
		if (op == '<' || op == '>') && p.pos < len(p.src) && p.src[p.pos] == op {
			double = op
			p.pos++
		}
		right, ok := p.parseAtom()
		if !ok {
			return 0, false
		}
		switch double {
		case '*':
			left *= right
		case '/':
			if right == 0 {
				return 0, false
			}
			left /= right
		case '%':
			if right == 0 {
				return 0, false
			}
			left %= right
		case '&':
			left &= right
		case '|':
			left |= right
		case '^':
			left ^= right
		default:
			return 0, false
		}
	}
	return left, true
}

func (p *cExprParser) parseAtom() (int64, bool) {
	for p.pos < len(p.src) && p.src[p.pos] == ' ' {
		p.pos++
	}
	if p.pos >= len(p.src) {
		return 0, false
	}
	if p.src[p.pos] == '(' {
		p.pos++
		v, ok := p.parseExpr()
		if !ok {
			return 0, false
		}
		if p.pos >= len(p.src) || p.src[p.pos] != ')' {
			return 0, false
		}
		p.pos++
		return v, true
	}
	sign := int64(1)
	if p.src[p.pos] == '-' {
		sign = -1
		p.pos++
	} else if p.src[p.pos] == '+' {
		p.pos++
	}
	start := p.pos
	for p.pos < len(p.src) && (isIdentByte(p.src[p.pos])) {
		p.pos++
	}
	if start == p.pos {
		return 0, false
	}
	token := strings.TrimRight(p.src[start:p.pos], "uUlL")
	var v int64
	var err error
	if isHexLiteral(token) {
		v, err = strconv.ParseInt(strings.TrimPrefix(strings.TrimPrefix(token, "0x"), "0X"), 16, 64)
	} else {
		v, err = strconv.ParseInt(token, 10, 64)
	}
	if err != nil {
		return 0, false
	}
	return sign * v, true
}

// --------------------------------------------------------------------------- //
// 小工具
// --------------------------------------------------------------------------- //

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

func baseName(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func textsOf(toks []CTok) string {
	out := make([]string, 0, len(toks))
	for _, t := range toks {
		out = append(out, t.Text)
	}
	return strings.Join(out, " ")
}

func textsSlice(toks []CTok) []string {
	out := make([]string, 0, len(toks))
	for _, t := range toks {
		out = append(out, t.Text)
	}
	return out
}

func indexOf(list []string, want string) int {
	for i, v := range list {
		if v == want {
			return i
		}
	}
	return -1
}

func itoa(n int) string { return strconv.Itoa(n) }

func formatFloat(f float64) string { return strconv.FormatFloat(f, 'f', 8, 64) }
// collectMacroConsts 收集头文件里的对象式宏并渲染成 `const` 行。
//
// 返回：const 行、头文件级提示（预处理警告 + 「头文件：x」）、宏定型提示。
// clang 档与 builtin 档共用这一段——`#define` 只存在于预处理阶段，AST 里看不到。
func collectMacroConsts(headers []string) (constLines, headerNotes, macroNotes []string) {
	macros := map[string]string{}
	macroOrder := []string{}
	for _, header := range headers {
		data, err := os.ReadFile(header)
		if err != nil {
			continue
		}
		notes := []string{}
		StripPreprocessor(string(data), macros, &macroOrder, &notes)
		for _, note := range notes {
			if !contains(headerNotes, note) {
				headerNotes = append(headerNotes, note)
			}
		}
		headerNotes = append(headerNotes, "头文件："+baseName(header))
	}
	for _, k := range macroOrder {
		v := macros[k]
		if strings.Contains(k, "(") {
			continue
		}
		lit := macroLiteral(v)
		if lit == nil {
			if strings.TrimSpace(v) != "" {
				macroNotes = append(macroNotes, "宏无法求值，已跳过："+k+" = "+v)
			}
			continue
		}
		if strings.HasSuffix(k, "_H") {
			macroNotes = append(macroNotes, "包含守卫宏跳过："+k)
			continue
		}
		constLines = append(constLines, "public const "+ValueName(k)+": "+lit[0]+" = "+lit[1]+";")
		macroNotes = append(macroNotes, "宏定型："+k+" = "+lit[1]+" ("+lit[0]+")")
	}
	return constLines, headerNotes, macroNotes
}
