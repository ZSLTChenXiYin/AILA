// Package lexer 是 AILA 的词法分析器（docs/specs/0002 §2）。
//
// 与 Python 参考实现逐 token 对齐：tests/corpus/*.tokens.json 是契约快照，
// 任何差异都会让语料测试失败——这是重写期唯一可靠的判据。
package lexer

import (
	"strings"

	"github.com/aila-lang/aila/internal/diag"
)

// Kind 是 token 类别；关键字 token 的 Kind 就是关键字本身（与 Python 版一致）。
type Kind string

// 词法类别。
const (
	KindIdent Kind = "ident"
	KindInt   Kind = "int"
	KindFloat Kind = "float"
	KindStr   Kind = "str"
	KindChar  Kind = "char"
	KindPunct Kind = "punct"
	KindWild  Kind = "wild"
	KindEOF   Kind = "eof"
)

// Keywords 是保留字（designs/0001 §1.4）。
var Keywords = map[string]bool{
	"func": true, "let": true, "var": true, "const": true, "type": true, "struct": true,
	"enum": true, "state": true, "view": true, "module": true, "hot": true, "import": true,
	"public": true, "return": true, "if": true, "else": true, "for": true, "in": true,
	"match": true, "region": true, "defer": true, "unsafe": true, "coroutine": true,
	"thread": true, "yield": true, "out": true, "trap": true, "move": true, "len": true,
	"copy": true, "resume": true, "as": true, "trunc": true, "break": true, "continue": true,
	"true": true, "false": true,
}

// TypeWords 是内置类型族名（在类型位由 parser 识别）。
var TypeWords = map[string]bool{
	"I8": true, "I16": true, "I32": true, "I64": true,
	"U8": true, "U16": true, "U32": true, "U64": true,
	"F32": true, "F64": true, "Bool": true, "C32": true, "Text": true,
	"USize": true, "ISize": true, "Unsafe": true, "chan": true, "send": true,
	"recv": true, "coro": true,
}

var builtinVariants = map[string]bool{
	"Some": true, "None": true, "Ok": true, "Err": true, "Suspended": true, "Done": true,
}

var punct3 = []string{"<<=", ">>=", "..."}
var punct2 = []string{"->", "=>", "==", "!=", "<=", ">=", "+=", "-=", "*=", "/=", "%=",
	"&=", "|=", "^=", "<<", ">>", "&&", "||", "<-", "::"}
var punct1 = "(){}[],;:.@#=+-*/%&|^~!?<>"

var simpleEscapes = map[rune]rune{'n': '\n', 't': '\t', 'r': '\r', '\\': '\\', '"': '"', '\'': '\'', '0': 0}

// Token 是一个词法单元。
type Token struct {
	Kind  Kind
	Text  string
	Value any
	Span  diag.Span
}

// IsPunct 判断是否是某个标点。
func (t Token) IsPunct(texts ...string) bool {
	if t.Kind != KindPunct {
		return false
	}
	for _, x := range texts {
		if t.Text == x {
			return true
		}
	}
	return false
}

// IsKw 判断是否是某个关键字。
func (t Token) IsKw(texts ...string) bool {
	for _, x := range texts {
		if string(t.Kind) == x {
			return true
		}
	}
	return false
}

// IsWord 判断是否是某个内置类型族名（chan / send / recv / coro / Unsafe 不是关键字）。
func (t Token) IsWord(texts ...string) bool {
	if t.Kind != KindIdent {
		return false
	}
	for _, x := range texts {
		if t.Text == x {
			return true
		}
	}
	return false
}

type lexer struct {
	runes []rune
	path  string
	lines []string // 源码按行切分（诊断渲染用）
	pos   int
	ln    int
	col   int
	bag   *diag.Bag
	toks  []Token
}

// Lex 对整份源码做词法分析；诊断进 bag。
func Lex(path, text string) ([]Token, *diag.Bag) {
	bag := &diag.Bag{}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	clean := strings.TrimPrefix(strings.TrimPrefix(text, "\ufeff"), "")
	runes := []rune(clean)
	l := &lexer{runes: runes, path: path, lines: strings.Split(clean, "\n"),
		pos: 0, ln: 1, col: 1, bag: bag}
	l.run()
	return l.toks, bag
}

func (l *lexer) eof() bool { return l.pos >= len(l.runes) }

func (l *lexer) peekRune(off int) rune {
	if l.pos+off >= len(l.runes) {
		return 0
	}
	return l.runes[l.pos+off]
}

func (l *lexer) span(startLine, startCol, length int) diag.Span {
	text := ""
	if startLine-1 >= 0 && startLine-1 < len(l.lines) {
		text = l.lines[startLine-1]
	}
	return diag.Span{File: l.path, Line: startLine, Col: startCol, Length: length, Text: text}
}

func (l *lexer) advance(n int) {
	for i := 0; i < n; i++ {
		if l.pos >= len(l.runes) {
			return
		}
		if l.runes[l.pos] == '\n' {
			l.ln++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}
}

func (l *lexer) startswith(s string) bool {
	r := []rune(s)
	if l.pos+len(r) > len(l.runes) {
		return false
	}
	for i, c := range r {
		if l.runes[l.pos+i] != c {
			return false
		}
	}
	return true
}

func (l *lexer) errAt(code string, line, col, length int, msg, label, help string) {
	l.bag.Error(code, l.span(line, col, length), msg, label, help)
}

func (l *lexer) add(kind Kind, text string, value any, line, col, length int) {
	l.toks = append(l.toks, Token{Kind: kind, Text: text, Value: value, Span: l.span(line, col, length)})
}

func (l *lexer) run() {
	for {
		l.skipTrivia()
		if l.eof() {
			l.add(KindEOF, "", nil, l.ln, l.col, 1)
			return
		}
		startLine, startCol := l.ln, l.col
		c := l.runes[l.pos]
		switch {
		case isIdentStart(c):
			l.identOrRaw(startLine, startCol)
		case c >= '0' && c <= '9':
			l.number(startLine, startCol)
		case c == '"':
			l.strLit(startLine, startCol, false)
		case c == '\'':
			l.charLit(startLine, startCol)
		case c == '.' && l.peekRune(1) >= '0' && l.peekRune(1) <= '9':
			l.errAt("E1006", startLine, startCol, 1, "浮点字面量缺少前导数字",
				"这里以 `.` 开头", "写成 `0.5`，而不是 `.5`")
			l.number(startLine, startCol)
		default:
			l.punct(startLine, startCol)
		}
	}
}

func isIdentStart(c rune) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentPart(c rune) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}

func isDigit(c rune) bool { return c >= '0' && c <= '9' }

func (l *lexer) skipTrivia() {
	for !l.eof() {
		c := l.runes[l.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			l.advance(1)
			continue
		}
		if l.startswith("//") {
			for !l.eof() && l.runes[l.pos] != '\n' {
				l.advance(1)
			}
			continue
		}
		if l.startswith("/*") {
			sl, sc := l.ln, l.col
			l.advance(2)
			depth := 1
			for !l.eof() {
				if l.startswith("*/") {
					l.advance(2)
					depth--
					if depth == 0 {
						break
					}
				} else if l.startswith("/*") {
					l.errAt("E1004", l.ln, l.col, 2, "块注释不嵌套",
						"AILA 的块注释不嵌套（与 Rust 不同）", "把内层 `/*` 改成 `//`")
					l.advance(2)
				} else {
					l.advance(1)
				}
			}
			if depth != 0 {
				l.errAt("E1004", sl, sc, 2, "块注释未闭合", "从这里开始", "补上配对的 `*/`")
			}
			continue
		}
		return
	}
}

func (l *lexer) identOrRaw(startLine, startCol int) {
	begin := l.pos
	for !l.eof() && isIdentPart(l.runes[l.pos]) {
		l.advance(1)
	}
	word := string(l.runes[begin:l.pos])
	if word == "r" && !l.eof() && l.runes[l.pos] == '"' {
		l.strLit(startLine, startCol, true)
		return
	}
	n := l.pos - begin
	switch {
	case Keywords[word]:
		l.add(Kind(word), word, word, startLine, startCol, n)
	case word == "_":
		l.add(KindWild, "_", nil, startLine, startCol, 1)
	case builtinVariants[word]:
		l.add(KindIdent, word, word, startLine, startCol, n)
	default:
		l.add(KindIdent, word, word, startLine, startCol, n)
	}
}

func (l *lexer) number(startLine, startCol int) {
	begin := l.pos
	isFloat := false
	if l.startswith("0x") || l.startswith("0X") || l.startswith("0b") || l.startswith("0B") ||
		l.startswith("0o") || l.startswith("0O") {
		baseCh := l.runes[l.pos+1]
		l.advance(2)
		var valid func(rune) bool
		base := 10
		switch baseCh {
		case 'x', 'X':
			valid = func(c rune) bool {
				return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || c == '_'
			}
			base = 16
		case 'b', 'B':
			valid = func(c rune) bool { return c == '0' || c == '1' || c == '_' }
			base = 2
		default:
			valid = func(c rune) bool { return (c >= '0' && c <= '7') || c == '_' }
			base = 8
		}
		digitsBegin := l.pos
		for !l.eof() && valid(l.runes[l.pos]) {
			l.advance(1)
		}
		body := strings.ReplaceAll(string(l.runes[digitsBegin:l.pos]), "_", "")
		if body == "" {
			l.errAt("E1006", startLine, startCol, l.pos-begin, "进制前缀后没有数字",
				"", "例如 `0xFF` / `0b1010` / `0o755`")
			body = "0"
		}
		l.add(KindInt, string(l.runes[begin:l.pos]), parseIntBase(body, base), startLine, startCol, l.pos-begin)
		return
	}
	for !l.eof() && (isDigit(l.runes[l.pos]) || l.runes[l.pos] == '_') {
		l.advance(1)
	}
	if !l.eof() && l.runes[l.pos] == '.' {
		if l.peekRune(1) >= '0' && l.peekRune(1) <= '9' {
			isFloat = true
			l.advance(1)
			for !l.eof() && (isDigit(l.runes[l.pos]) || l.runes[l.pos] == '_') {
				l.advance(1)
			}
		} else {
			l.errAt("E1006", startLine, startCol, l.pos-begin+1, "浮点字面量缺少小数位",
				"小数点后必须有数字", "写成 `1.0`，而不是 `1.`")
			l.advance(1)
			isFloat = true
		}
	}
	if !l.eof() && (l.runes[l.pos] == 'e' || l.runes[l.pos] == 'E') {
		save := l.pos
		l.advance(1)
		if !l.eof() && (l.runes[l.pos] == '+' || l.runes[l.pos] == '-') {
			l.advance(1)
		}
		if !l.eof() && isDigit(l.runes[l.pos]) {
			isFloat = true
			for !l.eof() && (isDigit(l.runes[l.pos]) || l.runes[l.pos] == '_') {
				l.advance(1)
			}
		} else {
			l.pos = save
			l.errAt("E1006", startLine, startCol, 1, "指数部分缺少数字", "", "例如 `1e9`、`1.5e-3`")
		}
	}
	text := string(l.runes[begin:l.pos])
	// 数字后缀（1u8 这类）在 AILA 不存在
	if !l.eof() && isIdentStart(l.runes[l.pos]) {
		badBegin := l.pos
		for !l.eof() && isIdentPart(l.runes[l.pos]) {
			l.advance(1)
		}
		suffix := string(l.runes[badBegin:l.pos])
		l.errAt("E1006", startLine, startCol, len([]rune(text))+len([]rune(suffix)),
			"数字字面量不接受类型后缀 `"+suffix+"`",
			"多态字面量由目标类型定型，不需要后缀",
			"去掉后缀，靠上下文/标注定型：`let x: U8 = 200;`")
		text = string(l.runes[begin:badBegin])
	}
	clean := strings.ReplaceAll(text, "_", "")
	if isFloat {
		l.add(KindFloat, text, parseFloat(clean), startLine, startCol, len([]rune(text)))
		return
	}
	if len(clean) > 1 && clean[0] == '0' && isDigit(rune(clean[1])) {
		l.errAt("E1006", startLine, startCol, len([]rune(text)), "C 式 0 前导八进制被禁用",
			"AILA 的八进制必须写 `0o`", "把 `"+text+"` 改成 `0o"+clean[1:]+"`，或去掉前导 0")
	}
	l.add(KindInt, text, parseIntBase(clean, 10), startLine, startCol, len([]rune(text)))
}

// readEscaped 读取到结束引号为止，返回 (值, 是否闭合)。
func (l *lexer) readEscaped(raw bool, quote rune) (string, bool) {
	var out []rune
	for !l.eof() {
		c := l.runes[l.pos]
		if c == '\n' {
			break
		}
		if c == quote {
			l.advance(1)
			return string(out), true
		}
		if c == '\\' && !raw {
			escLine, escCol := l.ln, l.col
			l.advance(1)
			if l.eof() {
				break
			}
			e := l.runes[l.pos]
			if v, ok := simpleEscapes[e]; ok {
				out = append(out, v)
				l.advance(1)
				continue
			}
			if e == 'u' {
				l.advance(1)
				if !l.eof() && l.runes[l.pos] == '{' {
					l.advance(1)
					var hex []rune
					for !l.eof() && l.runes[l.pos] != '}' {
						hex = append(hex, l.runes[l.pos])
						l.advance(1)
					}
					if !l.eof() {
						l.advance(1)
					}
					if code, ok := parseUnicode(hex); ok {
						out = append(out, code)
					} else {
						l.errAt("E1005", escLine, escCol, 4, "非法 Unicode 码点转义", "",
							"形如 `\\u{4e2d}`，码点 ≤ 0x10FFFF")
					}
				} else {
					l.errAt("E1005", escLine, escCol, 2, "`\\u` 后必须跟 `{...}`", "", "写成 `\\u{4e2d}`")
				}
				continue
			}
			l.errAt("E1005", escLine, escCol, 2, "非法转义 `\\"+string(e)+"`",
				"AILA 转义集：\\n \\t \\r \\\\ \\\" \\' \\0 \\u{...}",
				"字节需求请用数组字面量 `[0x41, 0x42]`")
			l.advance(1)
			continue
		}
		out = append(out, c)
		l.advance(1)
	}
	return string(out), false
}

func (l *lexer) strLit(startLine, startCol int, raw bool) {
	begin := l.pos
	// `r` 已由 identOrRaw 消费；这里只消费开引号（与 Python 版一致）
	l.advance(1) // "
	value, closed := l.readEscaped(raw, '"')
	length := l.pos - begin
	if !closed {
		l.errAt("E1002", startLine, startCol, length, "字符串未闭合", "从这里开始",
			"在行内补上收尾的 `\"`；长文本可用 `r\"...\"`")
	}
	l.add(KindStr, value, value, startLine, startCol, length)
}

func (l *lexer) charLit(startLine, startCol int) {
	begin := l.pos
	l.advance(1) // '
	value, closed := l.readEscaped(false, '\'')
	length := l.pos - begin
	if !closed {
		l.errAt("E1002", startLine, startCol, length, "字符字面量未闭合", "", "单引号必须成对；形如 `'A'`")
	}
	if runes := []rune(value); len(runes) != 1 {
		l.errAt("E1003", startLine, startCol, length,
			"单引号必须是恰好一个码点，这里有 "+itoa(len(runes))+" 个",
			"C32 字面量只能是一个码点", "多字符请用双引号字符串 `\"...\"`")
	}
	out := "\x00"
	if rs := []rune(value); len(rs) > 0 {
		out = string(rs[0])
	}
	l.add(KindChar, value, out, startLine, startCol, length)
}


func (l *lexer) punct(startLine, startCol int) {
	for _, p := range punct3 {
		if l.startswith(p) {
			n := len([]rune(p))
			l.advance(n)
			l.add(KindPunct, p, p, startLine, startCol, n)
			return
		}
	}
	for _, p := range punct2 {
		if l.startswith(p) {
			n := len([]rune(p))
			l.advance(n)
			if p == "::" {
				l.errAt("E1007", startLine, startCol, 2, "`::` 不是 AILA 符号",
					"类型名空间统一用点号", "`State::Idle` → `State.Idle`")
			}
			l.add(KindPunct, p, p, startLine, startCol, n)
			return
		}
	}
	c := l.runes[l.pos]
	if strings.ContainsRune(punct1, c) {
		l.advance(1)
		l.add(KindPunct, string(c), string(c), startLine, startCol, 1)
		return
	}
	l.errAt("E1001", startLine, startCol, 1, "非法字符 `"+string(c)+"`",
		"AILA 词法不接受的字符", "删掉它或改写成合法写法")
	l.advance(1)
}
