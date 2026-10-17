package parser

import (
	"github.com/ailang/aila/internal/ast"
	"github.com/ailang/aila/internal/diag"
	"github.com/ailang/aila/internal/lexer"
)

// --------------------------------------------------------------------------- //
// 模式
// --------------------------------------------------------------------------- //

func (p *Parser) parsePattern() ast.Node {
	tok := p.cur()
	switch {
	case tok.Kind == lexer.KindWild:
		p.adv()
		return &ast.Wildcard{Span: p.spanOf(tok)}
	case tok.Kind == lexer.KindInt:
		p.adv()
		return &ast.LitPattern{Span: p.spanOf(tok), Value: tok.Value, KindV: "int"}
	case tok.Kind == lexer.KindFloat:
		p.adv()
		return &ast.LitPattern{Span: p.spanOf(tok), Value: tok.Value, KindV: "float"}
	case tok.Kind == lexer.KindStr:
		p.adv()
		return &ast.LitPattern{Span: p.spanOf(tok), Value: tok.Value, KindV: "str"}
	case tok.Kind == lexer.KindChar:
		p.adv()
		return &ast.LitPattern{Span: p.spanOf(tok), Value: tok.Value, KindV: "char"}
	case tok.IsKw("true", "false"):
		p.adv()
		return &ast.LitPattern{Span: p.spanOf(tok), Value: tok.Kind == lexer.Kind("true"), KindV: "bool"}
	case tok.IsPunct("-"):
		p.adv()
		num := p.cur()
		if num.Kind == lexer.KindInt || num.Kind == lexer.KindFloat {
			p.adv()
			kind := "int"
			if num.Kind == lexer.KindFloat {
				kind = "float"
			}
			return &ast.LitPattern{Span: p.spanOf(tok), Value: num.Value, Negative: true, KindV: kind}
		}
		p.errAt(num, "E2001", "`-` 之后必须是数字字面量", "这里", "")
		return &ast.Wildcard{Span: p.spanOf(tok)}
	case tok.IsPunct("("):
		p.adv()
		var names []string
		for !p.atPunct(")") && !p.atKind(lexer.KindEOF) {
			if p.atKind(lexer.KindWild) {
				p.adv()
				names = append(names, "_")
			} else {
				names = append(names, p.expectKind(lexer.KindIdent, "绑定名").Text)
			}
			if _, ok := p.eatPunct(","); !ok {
				break
			}
		}
		p.expectPunct(")", "")
		return &ast.PathPattern{Span: p.spanOf(tok), Path: names}
	case tok.Kind == lexer.KindIdent:
		names := []string{p.adv().Text}
		for p.atPunct(".") && p.peek(1).Kind == lexer.KindIdent {
			p.adv()
			names = append(names, p.adv().Text)
		}
		if p.atPunct("{") {
			p.adv()
			var fields []*ast.FieldPat
			for !p.atPunct("}") && !p.atKind(lexer.KindEOF) {
				fname := p.expectKind(lexer.KindIdent, "字段名")
				fp := &ast.FieldPat{Span: p.spanOf(fname), Name: fname.Text}
				if _, ok := p.eatPunct(":"); ok {
					if p.atPunct("(") {
						g := p.adv()
						var names2 []string
						for !p.atPunct(")") && !p.atKind(lexer.KindEOF) {
							if p.atKind(lexer.KindWild) {
								p.adv()
								names2 = append(names2, "_")
							} else {
								names2 = append(names2, p.expectKind(lexer.KindIdent, "绑定名").Text)
							}
							if _, ok := p.eatPunct(","); !ok {
								break
							}
						}
						p.expectPunct(")", "")
						fp.Group = &ast.GroupPattern{Span: p.spanOf(g), Names: names2}
					} else if isLiteralStart(p.cur()) {
						if lit, ok := p.parsePattern().(*ast.LitPattern); ok {
							fp.Value = lit
						}
					} else {
						fp.Binding = p.expectKind(lexer.KindIdent, "绑定名").Text
					}
				} else {
					fp.Binding = fname.Text
				}
				fields = append(fields, fp)
				if _, ok := p.eatPunct(","); !ok {
					break
				}
			}
			p.expectPunct("}", "")
			return &ast.VariantPattern{Span: p.spanOf(tok), Path: names, Fields_: fields}
		}
		return &ast.PathPattern{Span: p.spanOf(tok), Path: names}
	}
	got := tok.Text
	if got == "" {
		got = "文件结束"
	}
	p.errAt(tok, "E2001", "此处期望模式，读到 `"+got+"`", "这里", "")
	p.adv()
	return &ast.Wildcard{Span: p.spanOf(tok)}
}

func isLiteralStart(t lexer.Token) bool {
	return t.Kind == lexer.KindInt || t.Kind == lexer.KindFloat || t.Kind == lexer.KindStr ||
		t.Kind == lexer.KindChar || t.IsKw("true", "false")
}

// --------------------------------------------------------------------------- //
// 表达式
// --------------------------------------------------------------------------- //

func (p *Parser) parseExpr() ast.Node { return p.parseBinary(0) }

func (p *Parser) parseExprNoStructLit() ast.Node {
	p.noStructLit++
	defer func() { p.noStructLit-- }()
	return p.parseExpr()
}

func (p *Parser) parseBinary(level int) ast.Node {
	if level >= len(binLevels) {
		return p.parseConversion()
	}
	left := p.parseBinary(level + 1)
	for p.atPunct(binLevels[level]...) {
		opTok := p.adv()
		right := p.parseBinary(level + 1)
		if contains(cmpOps, opTok.Text) {
			if b, ok := left.(*ast.Binary); ok && contains(cmpOps, b.Op) {
				p.errSpan(p.spanTo(b.Span, p.spanOf(opTok)),
					"E2004", "比较链被禁用",
					"`a < b < c` 的语义漂移是暗坑",
					"写成合取：`a < b && b < c`")
			}
		}
		left = &ast.Binary{Span: p.spanOf(opTok), Op: opTok.Text, Left: left, Right: right}
	}
	return left
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func (p *Parser) parseConversion() ast.Node {
	expr := p.parseUnary()
	for p.atKw("as", "trunc") {
		op := p.adv()
		target := p.parseTypePostfix(p.parseTypeAtom())
		expr = &ast.Cast{Span: p.spanOf(op), KindV: string(op.Kind), Operand: expr, Target: target}
	}
	return expr
}

func (p *Parser) parseUnary() ast.Node {
	tok := p.cur()
	if tok.IsPunct("-", "!", "~") {
		p.adv()
		operand := p.parseUnary()
		return &ast.Unary{Span: p.spanOf(tok), Op: tok.Text, Operand: operand}
	}
	if tok.IsPunct("?") && p.peek(1).IsPunct("(") {
		p.adv()
		p.adv()
		inner := p.parseExpr()
		end := p.expectPunct(")", "`?(...)` 检测表达式结束")
		return &ast.TryDetect{Span: p.spanTo(p.spanOf(tok), p.spanOf(end)), Operand: inner}
	}
	if tok.IsPunct("<-") {
		p.adv()
		operand := p.parseUnary()
		return &ast.Recv{Span: p.spanOf(tok), Chan: operand}
	}
	if formingUnary[string(tok.Kind)] {
		kwTok := p.adv()
		nxt := p.cur()
		if nxt.IsPunct("(") && nxt.Span.Line == kwTok.Span.Line &&
			kwTok.Span.Col+kwTok.Span.Length == nxt.Span.Col {
			p.errSpan(p.spanTo(p.spanOf(kwTok), p.spanOf(nxt)),
				"E2005", "`"+string(kwTok.Kind)+"` 是前缀定形关键字，没有函数/方法形态",
				"定形关键字只能写在操作数左边",
				"写成 `"+string(kwTok.Kind)+" 操作数`；`copy` 的区域实参写成 `copy x @[r]`")
		}
		if formingUnary[string(p.cur().Kind)] {
			p.errAt(p.cur(), "E2005", "定形一元操作不可嵌套前缀",
				"`copy move x` / `len move x` 都不合法", "拆成两条语句，或先绑定中间值")
		}
		var operand ast.Node
		if kwTok.Kind == lexer.Kind("copy") {
			p.noRegionArgs++
			operand = p.parseUnary()
			p.noRegionArgs--
		} else {
			operand = p.parseUnary()
		}
		switch kwTok.Kind {
		case lexer.Kind("move"):
			return &ast.Move{Span: p.spanOf(kwTok), Operand: operand}
		case lexer.Kind("len"):
			return &ast.Len{Span: p.spanOf(kwTok), Operand: operand}
		case lexer.Kind("resume"):
			return &ast.Resume{Span: p.spanOf(kwTok), Operand: operand}
		default: // copy
			regionArgs := p.parseRegionArgs()
			end := p.spanOf(p.cur())
			return &ast.Copy{Span: p.spanTo(p.spanOf(kwTok), end), Operand: operand, RegionArgs: regionArgs}
		}
	}
	return p.parsePostfix()
}

func (p *Parser) parseRegionArgs() []string {
	if !(p.atPunct("@") && p.peek(1).IsPunct("[")) {
		return nil
	}
	p.adv()
	p.adv()
	var names []string
	for !p.atPunct("]") && !p.atKind(lexer.KindEOF) {
		names = append(names, p.expectKind(lexer.KindIdent, "区域名").Text)
		if _, ok := p.eatPunct(","); !ok {
			break
		}
	}
	p.expectPunct("]", "")
	return names
}

func (p *Parser) parsePostfix() ast.Node {
	expr := p.parseAtom()
	for {
		tok := p.cur()
		switch {
		case tok.IsPunct("("):
			p.adv()
			var args []ast.Node
			for !p.atPunct(")") && !p.atKind(lexer.KindEOF) {
				args = append(args, p.parseExpr())
				if _, ok := p.eatPunct(","); !ok {
					break
				}
			}
			p.expectPunct(")", "实参列表结束")
			regionArgs := p.parseRegionArgs()
			expr = &ast.Call{Span: p.spanOf(tok), Callee: expr, Args: args, RegionArgs: regionArgs}
			continue
		case tok.IsPunct("["):
			p.adv()
			idx := p.parseExpr()
			p.expectPunct("]", "下标结束")
			expr = &ast.Index{Span: p.spanOf(tok), Base: expr, Index_: idx}
			continue
		case tok.IsPunct("."):
			p.adv()
			name := p.expectKind(lexer.KindIdent, "字段名")
			if p.atPunct("(") {
				p.adv()
				var args []ast.Node
				for !p.atPunct(")") && !p.atKind(lexer.KindEOF) {
					args = append(args, p.parseExpr())
					if _, ok := p.eatPunct(","); !ok {
						break
					}
				}
				p.expectPunct(")", "")
				regionArgs := p.parseRegionArgs()
				expr = &ast.MethodCall{Span: p.spanOf(tok), Recv: expr, Name: name.Text,
					Args: args, RegionArgs: regionArgs}
			} else {
				expr = &ast.FieldAccess{Span: p.spanOf(name), Base: expr, Name: name.Text}
			}
			continue
		case tok.IsPunct("?"):
			p.adv()
			expr = &ast.TryPropagate{Span: p.spanOf(tok), Operand: expr}
			continue
		case tok.IsPunct("@"):
			if p.peek(1).IsPunct("[") {
				if p.noRegionArgs > 0 {
					return expr // 区域实参属于更外层的 `copy`
				}
				p.adv()
				regionArgs := p.parseRegionArgs()
				switch e := expr.(type) {
				case *ast.Call:
					e.RegionArgs = regionArgs
				case *ast.MethodCall:
					e.RegionArgs = regionArgs
				case *ast.Copy:
					e.RegionArgs = regionArgs
				default:
					p.errAt(tok, "E2007", "`@[...]` 只能跟在调用或 `copy` 之后",
						"区域实参通道只服务「需要供给区域的操作」",
						"分配标注请写裸名形态 `[...]@区域名`")
				}
				continue
			}
			p.adv()
			region := p.expectKind(lexer.KindIdent, "区域名")
			if p.atPunct("[") {
				p.errAt(tok, "E2007", "两种 `@` 拼法混用了",
					"分配标注是裸名 `@区域名`，区域实参是列表 `@[a, b]`",
					"分配写成 `[...]@work`，调用写成 `f(...)@[work]`")
			}
			expr = &ast.Alloc{Span: p.spanOf(tok), Expr: expr, Region: region.Text}
			continue
		}
		break
	}
	return expr
}

func (p *Parser) parseAtom() ast.Node {
	tok := p.cur()
	switch {
	case tok.Kind == lexer.KindInt:
		p.adv()
		return &ast.IntLit{Span: p.spanOf(tok), Raw: toInt64(tok.Value)}
	case tok.Kind == lexer.KindFloat:
		p.adv()
		return &ast.FloatLit{Span: p.spanOf(tok), Value: toFloat64(tok.Value)}
	case tok.Kind == lexer.KindStr:
		p.adv()
		return &ast.StrLit{Span: p.spanOf(tok), Value: asString(tok.Value)}
	case tok.Kind == lexer.KindChar:
		p.adv()
		return &ast.CharLit{Span: p.spanOf(tok), Value: asString(tok.Value)}
	case tok.IsKw("true", "false"):
		p.adv()
		return &ast.BoolLit{Span: p.spanOf(tok), Value: tok.Kind == lexer.Kind("true")}
	case tok.Kind == lexer.KindWild:
		p.adv()
		return &ast.Placeholder{Span: p.spanOf(tok)}
	case tok.IsPunct("("):
		p.adv()
		saved := p.noStructLit
		p.noStructLit = 0
		var exprs []ast.Node
		exprs = append(exprs, p.parseExpr())
		for {
			if _, ok := p.eatPunct(","); !ok {
				break
			}
			exprs = append(exprs, p.parseExpr())
		}
		p.noStructLit = saved
		p.expectPunct(")", "括号结束")
		if len(exprs) == 1 {
			return exprs[0]
		}
		return &ast.GroupExpr{Span: p.spanOf(tok), Elems: exprs}
	case tok.IsPunct("["):
		return p.parseArrayLit()
	case tok.IsPunct("#"):
		return p.parseCompileQuery()
	case tok.IsWord("chan"):
		return p.parseChanCreate()
	case tok.IsKw("coroutine"):
		return p.parseCoroCreate()
	case tok.IsKw("match"):
		return p.parseMatch()
	case tok.Kind == lexer.KindIdent:
		names := []string{p.adv().Text}
		for p.atPunct(".") && p.peek(1).Kind == lexer.KindIdent {
			p.adv()
			names = append(names, p.adv().Text)
		}
		if p.atPunct("{") && p.noStructLit == 0 {
			p.adv()
			var pairs []ast.Pair
			for !p.atPunct("}") && !p.atKind(lexer.KindEOF) {
				fname := p.expectKind(lexer.KindIdent, "字段名")
				if _, ok := p.eatPunct(":"); !ok {
					p.errAt(fname, "E2001", "构造字面量必须写全字段名",
						"构造 = `类型 { 字段: 值, ... }`",
						"写成 `"+fname.Text+": 值`")
				}
				value := p.parseExpr()
				pairs = append(pairs, ast.P(fname.Text, value))
				if _, ok := p.eatPunct(","); !ok {
					break
				}
			}
			p.expectPunct("}", "构造字面量结束")
			if p.atPunct("@") {
				p.errAt(p.cur(), "E2007", "构造字面量不能再挂 `@`",
					"分配标注只写在数组字面量上", "")
			}
			return &ast.StructLit{Span: p.spanOf(tok), TypePath: names, FieldPairs: pairs}
		}
		return &ast.Path{Span: p.spanOf(tok), Names: names}
	}
	got := tok.Text
	if got == "" {
		got = "文件结束"
	}
	p.errAt(tok, "E2001", "此处期望表达式，读到 `"+got+"`", "这里", "")
	if !p.atKind(lexer.KindEOF) {
		p.adv()
	}
	return &ast.IntLit{Span: p.spanOf(tok), Raw: 0}
}

func (p *Parser) parseArrayLit() ast.Node {
	start := p.adv() // [
	if _, ok := p.eatPunct("]"); ok {
		return &ast.ArrayLit{Span: p.spanOf(start)}
	}
	first := p.parseExpr()
	if _, ok := p.eatPunct(";"); ok {
		count := p.parseExpr()
		end := p.expectPunct("]", "数组字面量结束")
		return &ast.ArrayLit{Span: p.spanTo(p.spanOf(start), p.spanOf(end)), Repeat: first, Count: count}
	}
	elems := []ast.Node{first}
	for {
		if _, ok := p.eatPunct(","); !ok {
			break
		}
		if p.atPunct("]") {
			break
		}
		elems = append(elems, p.parseExpr())
	}
	end := p.expectPunct("]", "数组字面量结束")
	return &ast.ArrayLit{Span: p.spanTo(p.spanOf(start), p.spanOf(end)), Elems: elems}
}

func (p *Parser) parseCompileQuery() ast.Node {
	start := p.adv() // #
	name := p.adv()
	p.expectPunct("(", "编译期查询")
	target := p.parseType()
	end := p.expectPunct(")", "编译期查询结束")
	sp := p.spanTo(p.spanOf(start), p.spanOf(end))
	switch name.Text {
	case "size":
		return &ast.SizeOf{Span: sp, Target: target}
	case "align":
		return &ast.AlignOf{Span: sp, Target: target}
	}
	p.errAt(name, "E3001", "未知的编译期查询 `#"+name.Text+"`",
		"编译期域只有 `#size(T)` / `#align(T)`", "")
	return &ast.SizeOf{Span: sp, Target: target}
}

func (p *Parser) parseChanCreate() ast.Node {
	start := p.adv() // chan
	region := ""
	if p.atPunct("@") && p.peek(1).Kind == lexer.KindIdent {
		p.adv()
		region = p.adv().Text
	}
	elem := p.parseTypeAtom()
	p.expectPunct("(", "通道容量")
	capacity := p.parseExpr()
	end := p.expectPunct(")", "通道容量结束")
	return &ast.ChanCreate{Span: p.spanTo(p.spanOf(start), p.spanOf(end)), Elem: elem,
		Cap: capacity, Region: region}
}

func (p *Parser) parseCoroCreate() ast.Node {
	start := p.adv() // coroutine
	region := ""
	if p.atPunct("@") && p.peek(1).Kind == lexer.KindIdent {
		p.adv()
		region = p.adv().Text
	}
	call := p.parseUnary()
	return &ast.CoroCreate{Span: p.spanOf(start), Call: call, FrameRegion: region}
}

// --------------------------------------------------------------------------- //
// 简单语句
// --------------------------------------------------------------------------- //

func (p *Parser) parseSimpleStmtNoSemi() ast.Node {
	tok := p.cur()
	if tok.Kind == lexer.KindWild && p.peek(1).IsPunct("=") {
		p.adv()
		p.adv()
		value := p.parseExpr()
		return &ast.Discard{Span: p.spanOf(tok), Value: value}
	}
	expr := p.parseExpr()
	if p.atPunct("=") || p.atPunct(compoundOps...) {
		op := p.adv().Text
		value := p.parseExpr()
		return &ast.Assign{Span: p.spanOf(tok), Target: expr, Op: op, Value: value}
	}
	return &ast.ExprStmt{Span: p.spanOf(tok), Expr: expr}
}

func (p *Parser) parseAssign() ast.Node {
	tok := p.cur()
	if tok.Kind == lexer.KindWild {
		p.adv()
		p.expectPunct("=", "丢弃")
		value := p.parseExpr()
		end := p.expectPunct(";", "丢弃语句结束")
		return &ast.Discard{Span: p.spanTo(p.spanOf(tok), p.spanOf(end)), Value: value}
	}
	stmt := p.parseSimpleStmtNoSemi()
	p.expectPunct(";", "语句结束")
	return stmt
}

// --------------------------------------------------------------------------- //
// 值转换
// --------------------------------------------------------------------------- //

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case float64:
		return int64(x)
	}
	return 0
}

func toFloat64(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int64:
		return float64(x)
	case int:
		return float64(x)
	}
	return 0
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

var _ = diag.Empty
