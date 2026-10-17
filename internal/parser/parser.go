// Package parser 是 AILA 的语法分析器（docs/specs/0002 §3）。
//
// 与 Python 参考实现逐节点对齐：tests/corpus/*.ast.txt 是契约快照。
// 解析器对**设计明文禁用的写法**直接出诊断（不静默兜底），见 specs/0002 §3.1。
package parser

import (
	"strings"

	"github.com/aila-lang/aila/internal/ast"
	"github.com/aila-lang/aila/internal/diag"
	"github.com/aila-lang/aila/internal/lexer"
)

// primTypes 是内置标量类型字（C32/Text 等族名见 lexer.TypeWords）。
var primTypes = map[string]string{
	"I8": "i8", "I16": "i16", "I32": "i32", "I64": "i64",
	"U8": "u8", "U16": "u16", "U32": "u32", "U64": "u64",
	"F32": "f32", "F64": "f64", "Bool": "bool", "C32": "c32",
	"Text": "text", "USize": "usize", "ISize": "isize",
}

var cmpOps = []string{"==", "!=", "<", "<=", ">", ">="}
var compoundOps = []string{"+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "<<=", ">>="}
var formingUnary = map[string]bool{"move": true, "len": true, "copy": true, "resume": true}

// 二元运算优先级（档 4..12；档 0-3 是原子/后缀/前缀/转换）。
var binLevels = [][]string{
	{"||"}, {"&&"}, cmpOps, {"|"}, {"^"}, {"&"}, {"<<", ">>"}, {"+", "-"}, {"*", "/", "%"},
}

// Parser 是语法分析器状态。
type Parser struct {
	toks  []lexer.Token
	i     int
	bag   *diag.Bag
	path  string
	lines []string

	noStructLit  int
	noRegionArgs int
}

// Parse 解析一个源文件。
func Parse(path, text string) (*ast.Module, *diag.Bag) {
	toks, bag := lexer.Lex(path, text)
	p := &Parser{toks: toks, bag: bag, path: path,
		lines: strings.Split(strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n"), "\n")}
	return p.parseFile(), p.bag
}

// --------------------------------------------------------------------------- //
// token 工具
// --------------------------------------------------------------------------- //

func (p *Parser) peek(k int) lexer.Token {
	j := p.i + k
	if j >= len(p.toks) {
		return p.toks[len(p.toks)-1]
	}
	return p.toks[j]
}

func (p *Parser) cur() lexer.Token { return p.peek(0) }

func (p *Parser) atKind(k lexer.Kind) bool { return p.cur().Kind == k }

func (p *Parser) atPunct(texts ...string) bool { return p.cur().IsPunct(texts...) }

func (p *Parser) atKw(texts ...string) bool { return p.cur().IsKw(texts...) }

func (p *Parser) atWord(texts ...string) bool { return p.cur().IsWord(texts...) }

func (p *Parser) adv() lexer.Token {
	t := p.cur()
	if p.i < len(p.toks)-1 {
		p.i++
	}
	return t
}

func (p *Parser) eatPunct(texts ...string) (lexer.Token, bool) {
	if p.atPunct(texts...) {
		return p.adv(), true
	}
	return lexer.Token{}, false
}

func (p *Parser) eatKw(texts ...string) (lexer.Token, bool) {
	if p.atKw(texts...) {
		return p.adv(), true
	}
	return lexer.Token{}, false
}

func (p *Parser) spanOf(t lexer.Token) diag.Span {
	text := ""
	if t.Span.Line-1 >= 0 && t.Span.Line-1 < len(p.lines) {
		text = p.lines[t.Span.Line-1]
	}
	return diag.Span{File: p.path, Line: t.Span.Line, Col: t.Span.Col, Length: t.Span.Length, Text: text}
}

func (p *Parser) spanTo(a, b diag.Span) diag.Span {
	out := a
	if b.Line == a.Line {
		end := b.Col + b.Length
		if end > a.Col+a.Length {
			out.Length = end - a.Col
		}
	}
	if out.Length < 1 {
		out.Length = 1
	}
	return out
}

func (p *Parser) errAt(t lexer.Token, code, msg, label, help string) {
	p.bag.Error(code, p.spanOf(t), msg, label, help)
}

func (p *Parser) errSpan(sp diag.Span, code, msg, label, help string) {
	p.bag.Error(code, sp, msg, label, help)
}

func (p *Parser) expectPunct(text, what string) lexer.Token {
	if t, ok := p.eatPunct(text); ok {
		return t
	}
	got := p.cur().Text
	if got == "" {
		got = "文件结束"
	}
	help := "补上 `" + text + "`"
	if what != "" {
		p.errAt(p.cur(), "E2001", "期望 `"+text+"`（"+what+"），读到 `"+got+"`", "这里", help)
	} else {
		p.errAt(p.cur(), "E2001", "期望 `"+text+"`，读到 `"+got+"`", "这里", help)
	}
	return lexer.Token{Kind: lexer.KindPunct, Text: text, Span: p.cur().Span}
}

func (p *Parser) expectKind(k lexer.Kind, what string) lexer.Token {
	if p.atKind(k) {
		return p.adv()
	}
	got := p.cur().Text
	if got == "" {
		got = "文件结束"
	}
	if what == "" {
		what = "标识符"
	}
	p.errAt(p.cur(), "E2001", "期望 "+what+"，读到 `"+got+"`", "这里", "")
	return lexer.Token{Kind: k, Span: p.cur().Span}
}

// --------------------------------------------------------------------------- //
// 模块
// --------------------------------------------------------------------------- //

func (p *Parser) parseFile() *ast.Module {
	mod := &ast.Module{}
	for p.atKw("import") {
		mod.Imports = append(mod.Imports, p.parseImport())
	}
	seenBlock := false
	for !p.atKind(lexer.KindEOF) {
		attrs := p.parseAttrs()
		if p.atKw("module") {
			if seenBlock {
				p.errAt(p.cur(), "E2010", "一个文件只能有一个 `module` 块", "这里", "")
			}
			if mod.IsLightweight && len(mod.Shell) > 0 {
				mod.Block = append(mod.Block, mod.Shell...)
				mod.Shell = nil
			}
			seenBlock = true
			p.adv()
			if _, ok := p.eatKw("hot"); ok {
				mod.IsHot = true
			}
			mod.IsLightweight = false
			p.expectPunct("{", "模块块开始")
			for !p.atPunct("}") && !p.atKind(lexer.KindEOF) {
				if d := p.parseDecl(false, p.parseAttrs()); d != nil {
					mod.Block = append(mod.Block, d)
				}
			}
			p.expectPunct("}", "模块块结束")
			continue
		}
		if seenBlock {
			p.errAt(p.cur(), "E2010", "`module` 块之后不得再有其它声明",
				"顺序要求：壳（import/类型/const/#[c] 声明）在前，`module` 块在后",
				"把这条声明移进 `module { ... }` 块内")
		}
		hasModuleLater := p.hasModuleBlockLater()
		decl := p.parseDecl(hasModuleLater, attrs)
		if decl == nil {
			continue
		}
		switch {
		case seenBlock:
			mod.Block = append(mod.Block, decl)
		case !mod.IsLightweight && p.hasModuleBlockLater():
			mod.Shell = append(mod.Shell, decl)
		default:
			mod.Block = append(mod.Block, decl)
		}
	}
	if !seenBlock {
		mod.IsLightweight = true
		mod.Block = append(mod.Shell, mod.Block...)
		mod.Shell = nil
	}
	return mod
}

// hasModuleBlockLater 扫描后续 token，判断是否还有 `module` 块（决定当前声明属壳还是块）。
func (p *Parser) hasModuleBlockLater() bool {
	depth := 0
	for j := p.i; j < len(p.toks); j++ {
		t := p.toks[j]
		if t.Kind == lexer.KindPunct {
			switch t.Text {
			case "{", "[", "(":
				depth++
			case "}", "]", ")":
				depth--
			}
		}
		if depth < 0 {
			return false
		}
		if depth == 0 && t.Kind == lexer.Kind("module") {
			return true
		}
	}
	return false
}

func (p *Parser) parseImport() *ast.Import {
	start := p.adv() // import
	kind := "aila"
	if p.atKind(lexer.KindIdent) && (p.cur().Text == "ch" || p.cur().Text == "cc") {
		kind = p.adv().Text
	}
	root := "_"
	switch {
	case p.atKind(lexer.KindWild):
		p.adv()
	case p.atKind(lexer.KindIdent):
		root = p.adv().Text
	default:
		p.errAt(p.cur(), "E2001", "`import` 之后期望导入路径（`_` 或包名）", "这里", "")
	}
	var path []string
	for {
		if _, ok := p.eatPunct("."); !ok {
			break
		}
		seg := p.expectKind(lexer.KindIdent, "模块名")
		path = append(path, seg.Text)
	}
	alias := ""
	if _, ok := p.eatKw("as"); ok {
		alias = p.expectKind(lexer.KindIdent, "别名").Text
	}
	p.expectPunct(";", "import 语句结束")
	return &ast.Import{Span: p.spanOf(start), Root: root, Path: path, KindV: kind, Alias: alias}
}

// --------------------------------------------------------------------------- //
// 属性
// --------------------------------------------------------------------------- //

func (p *Parser) parseAttrs() []*ast.Attr {
	var attrs []*ast.Attr
	for p.atPunct("#") {
		start := p.spanOf(p.adv())
		p.expectPunct("[", "属性列表")
		for !p.atPunct("]") && !p.atKind(lexer.KindEOF) {
			nameTok := p.cur()
			if nameTok.Kind != lexer.KindIdent {
				p.errAt(nameTok, "E2001", "属性名必须是标识符", "这里", "")
				p.adv()
				break
			}
			p.adv()
			var args []ast.Node
			if _, ok := p.eatPunct("("); ok {
				for !p.atPunct(")") && !p.atKind(lexer.KindEOF) {
					args = append(args, p.parseExpr())
					if _, ok := p.eatPunct(","); !ok {
						break
					}
				}
				p.expectPunct(")", "")
			}
			attrs = append(attrs, &ast.Attr{Span: p.spanTo(start, p.spanOf(nameTok)), Name: nameTok.Text, Args: args})
			if _, ok := p.eatPunct(","); !ok {
				break
			}
		}
		end := p.expectPunct("]", "属性列表结束")
		if len(attrs) > 0 {
			attrs[len(attrs)-1].Span = p.spanTo(start, p.spanOf(end))
		}
	}
	return attrs
}

// --------------------------------------------------------------------------- //
// 声明
// --------------------------------------------------------------------------- //

func hasCAttr(attrs []*ast.Attr) bool {
	for _, a := range attrs {
		if a.Name == "c" {
			return true
		}
	}
	return false
}

func (p *Parser) parseDecl(shell bool, attrs []*ast.Attr) ast.Node {
	_, isPublic := p.eatKw("public")
	tok := p.cur()
	var decl ast.Node
	switch {
	case tok.Kind == lexer.Kind("struct"):
		decl = p.parseStruct()
	case tok.Kind == lexer.Kind("enum"):
		decl = p.parseEnum()
	case tok.Kind == lexer.Kind("type"):
		decl = p.parseTypeAlias()
	case tok.Kind == lexer.Kind("func"), tok.Kind == lexer.Kind("coroutine"), tok.Kind == lexer.Kind("thread"):
		decl = p.parseFunc(attrs)
	case tok.Kind == lexer.Kind("let"), tok.Kind == lexer.Kind("var"),
		tok.Kind == lexer.Kind("const"), tok.Kind == lexer.Kind("state"):
		decl = p.parseBinding(true, hasCAttr(attrs))
	case tok.Kind == lexer.Kind("view"):
		decl = p.parseView()
	default:
		got := tok.Text
		if got == "" {
			got = "文件结束"
		}
		p.errAt(tok, "E2001",
			"此处期望声明（`func`/`struct`/`enum`/`type`/`let`/`var`/`const`/`state`/`view`），读到 `"+got+"`", "这里", "")
		if !p.atKind(lexer.KindEOF) {
			p.adv()
		}
		return nil
	}
	p.applyAttrs(decl, attrs, isPublic)
	if shell {
		switch d := decl.(type) {
		case *ast.LetDecl:
			if d.KindV != "const" {
				p.errSpan(d.Span, "E2010", "壳（`module` 块外）不允许 `"+d.KindV+"` 声明",
					"壳只允许纯声明：类型 / `const` / `#[c]` 无体函数",
					"把它移进 `module { ... }` 块内")
			}
		case *ast.ViewDecl:
			p.errSpan(d.Span, "E2010", "壳（`module` 块外）不允许 `view` 声明",
				"壳只允许纯声明", "把它移进 `module { ... }` 块内")
		case *ast.FuncDecl:
			if d.Body != nil && !d.IsExtern {
				p.errSpan(d.Span, "E2010", "壳（`module` 块外）不允许带函数体的 `func`",
					"方法 / 函数实现必须落在 `module` 块内",
					"把这条 `func` 移进 `module { ... }` 块；壳里只留 `#[c]` 无体声明")
			}
		}
	}
	return decl
}

func (p *Parser) applyAttrs(decl ast.Node, attrs []*ast.Attr, isPublic bool) {
	set := func(a *ast.Attrs) {
		a.Attrs = attrs
		a.IsPublic = isPublic
	}
	switch d := decl.(type) {
	case *ast.StructDecl:
		set(&d.Attrs)
	case *ast.EnumDecl:
		set(&d.Attrs)
	case *ast.TypeAlias:
		set(&d.Attrs)
	case *ast.FuncDecl:
		set(&d.Attrs)
	case *ast.LetDecl:
		set(&d.Attrs)
	case *ast.ViewDecl:
		set(&d.Attrs)
	}
}

func (p *Parser) parseStruct() ast.Node {
	start := p.spanOf(p.adv()) // struct
	name := p.expectKind(lexer.KindIdent, "结构体名")
	p.expectPunct("{", "结构体体")
	var fields []*ast.Field
	for !p.atPunct("}") && !p.atKind(lexer.KindEOF) {
		fname := p.expectKind(lexer.KindIdent, "字段名")
		p.expectPunct(":", "字段类型")
		ftype := p.parseType()
		fields = append(fields, &ast.Field{Span: p.spanOf(fname), Name: fname.Text, Type: ftype})
		if _, ok := p.eatPunct(","); !ok {
			break
		}
	}
	p.expectPunct("}", "结构体体结束")
	end := p.expectPunct(";", "结构体声明结束")
	return &ast.StructDecl{Span: p.spanTo(start, p.spanOf(end)), Name: name.Text, Fields_: fields}
}

func (p *Parser) parseEnum() ast.Node {
	start := p.spanOf(p.adv()) // enum
	name := p.expectKind(lexer.KindIdent, "枚举名")
	var base *ast.TypeRef
	if _, ok := p.eatPunct(":"); ok {
		base = p.parseTypeAtom()
	}
	p.expectPunct("{", "枚举体")
	var variants []*ast.Variant
	for !p.atPunct("}") && !p.atKind(lexer.KindEOF) {
		vname := p.expectKind(lexer.KindIdent, "变体名")
		var disc ast.Node
		if _, ok := p.eatPunct("="); ok {
			disc = p.parseExpr()
		}
		var payload []*ast.Field
		if p.atPunct("{") {
			p.adv()
			for !p.atPunct("}") && !p.atKind(lexer.KindEOF) {
				fname := p.expectKind(lexer.KindIdent, "载荷字段名")
				p.expectPunct(":", "字段类型")
				payload = append(payload, &ast.Field{Span: p.spanOf(fname), Name: fname.Text, Type: p.parseType()})
				if _, ok := p.eatPunct(","); !ok {
					break
				}
			}
			p.expectPunct("}", "变体载荷结束")
		}
		variants = append(variants, &ast.Variant{Span: p.spanOf(vname), Name: vname.Text,
			Discriminant: disc, Payload: payload})
		if _, ok := p.eatPunct(","); !ok {
			break
		}
	}
	p.expectPunct("}", "枚举体结束")
	end := p.expectPunct(";", "枚举声明结束")
	return &ast.EnumDecl{Span: p.spanTo(start, p.spanOf(end)), Name: name.Text, Base: base, Variants: variants}
}

func (p *Parser) parseTypeAlias() ast.Node {
	start := p.spanOf(p.adv()) // type
	name := p.expectKind(lexer.KindIdent, "别名")
	p.expectPunct("=", "别名目标")
	target := p.parseType()
	end := p.expectPunct(";", "类型别名结束")
	return &ast.TypeAlias{Span: p.spanTo(start, p.spanOf(end)), Name: name.Text, Target: target}
}
