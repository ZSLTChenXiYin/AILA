package parser

import (
	"github.com/ailang/aila/internal/ast"
	"github.com/ailang/aila/internal/lexer"
)

// isEffectClause 判断 `@` 之后的子句名：region / mut / move。
// 注意 `mut` 不是词法保留字（lexer.Keywords 只有 region/move），必须按文本识别。
func isEffectClause(t lexer.Token) bool {
	if t.IsKw("region", "mut", "move") {
		return true
	}
	return t.IsWord("region", "mut", "move")
}

func (p *Parser) parseEffects() []*ast.Effect {
	var effects []*ast.Effect
	for p.atPunct("@") {
		nxt := p.peek(1)
		if !isEffectClause(nxt) {
			break
		}
		at := p.adv()
		kind := p.adv().Text
		p.expectPunct("[", "效果子句")
		var names []string
		for !p.atPunct("]") && !p.atKind(lexer.KindEOF) {
			names = append(names, p.expectKind(lexer.KindIdent, "区域/参数名").Text)
			if _, ok := p.eatPunct(","); !ok {
				break
			}
		}
		p.expectPunct("]", "效果子句结束")
		effects = append(effects, &ast.Effect{Span: p.spanOf(at), KindV: string(kind), Names: names})
	}
	return effects
}

func (p *Parser) parseFunc(attrs []*ast.Attr, allowBodyless bool) ast.Node {
	start := p.spanOf(p.cur())
	_, isCoro := p.eatKw("coroutine")
	_, isThread := p.eatKw("thread")
	p.expectKind(lexer.Kind("func"), "`func`")
	nameTok := p.expectKind(lexer.KindIdent, "函数名")
	owner := ""
	if _, ok := p.eatPunct("."); ok {
		owner = nameTok.Text
		nameTok = p.expectKind(lexer.KindIdent, "方法名")
	}
	p.expectPunct("(", "参数列表")
	var params []*ast.Param
	variadic := false
	for !p.atPunct(")") && !p.atKind(lexer.KindEOF) {
		if p.atPunct("...") {
			p.adv()
			variadic = true
			break
		}
		if p.atKind(lexer.KindIdent) && p.cur().Text == "self" {
			ptok := p.adv()
			params = append(params, &ast.Param{Span: p.spanOf(ptok), Name: "self", IsSelf: true})
		} else {
			pname := p.expectKind(lexer.KindIdent, "参数名")
			p.expectPunct(":", "参数类型")
			params = append(params, &ast.Param{Span: p.spanOf(pname), Name: pname.Text, Type: p.parseType()})
		}
		if _, ok := p.eatPunct(","); !ok {
			break
		}
	}
	p.expectPunct(")", "参数列表结束")
	effects := p.parseEffects()
	var ret *ast.TypeRef
	if _, ok := p.eatPunct("->"); ok {
		ret = p.parseType()
	}
	// 设计稿 §13.2 的示例把效果写在返回类型之后；两种顺序都接受（rfcs/0002 缺陷 2）
	effects = append(effects, p.parseEffects()...)
	var body *ast.Block
	isExtern := hasCAttr(attrs)
	isIntrinsic := hasIntrinsicAttr(attrs)
	if p.atPunct(";") {
		p.adv()
		if !isExtern && !isIntrinsic && !allowBodyless {
			p.errSpan(p.spanTo(start, p.spanOf(p.cur())),
				"E2001", "函数 `"+nameTok.Text+"` 没有函数体",
				"只有 `#[c]` 外部声明或壳里的前向声明可以省略函数体",
				"补上 `{ ... }`；要在包里对外导出，就在 `module` 块外写 `public func ...;` 并在块内实现")
		}
	} else {
		body = p.parseBlock()
	}
	return &ast.FuncDecl{Span: p.spanTo(start, p.spanOf(p.cur())), Name: nameTok.Text,
		OwnerType: owner, IsCoroutine: isCoro, IsThread: isThread, IsExtern: isExtern, IsIntrinsic: isIntrinsic,
		Params: params, Effects: effects, Ret: ret, Body: body, Variadic: variadic}
}

func (p *Parser) parseBinding(semi bool, allowNoInit bool) ast.Node {
	kw := p.adv()
	if kw.Kind == lexer.Kind("const") || kw.Kind == lexer.Kind("state") {
		name := p.expectKind(lexer.KindIdent, "常量名")
		var typ *ast.TypeRef
		if _, ok := p.eatPunct(":"); ok {
			typ = p.parseType()
		}
		p.expectPunct("=", "初始化器")
		init := p.parseExpr()
		if semi {
			p.expectPunct(";", "声明结束")
		}
		return &ast.LetDecl{Span: p.spanTo(p.spanOf(kw), p.spanOf(p.cur())), KindV: string(kw.Kind),
			Name: name.Text, Type: typ, Init: init, HasInit: true}
	}
	name := p.expectKind(lexer.KindIdent, "绑定名")
	var typ *ast.TypeRef
	if _, ok := p.eatPunct(":"); ok {
		typ = p.parseType()
	}
	var init ast.Node
	hasInit := true
	if _, ok := p.eatPunct("="); ok {
		init = p.parseExpr()
	} else {
		hasInit = false
		if !allowNoInit {
			p.errSpan(p.spanTo(p.spanOf(kw), p.spanOf(p.cur())),
				"E2001", "绑定 `"+name.Text+"` 没有初始化器",
				"声明即初始化（铁律 3）：不允许先声明后赋值",
				"写成 `"+string(kw.Kind)+" "+name.Text+"[冒号类型] = 初始值;`，或加 #[c] 表示外部符号")
		}
	}
	if semi {
		p.expectPunct(";", "绑定声明结束")
	}
	return &ast.LetDecl{Span: p.spanTo(p.spanOf(kw), p.spanOf(p.cur())), KindV: string(kw.Kind),
		Name: name.Text, Type: typ, Init: init, HasInit: hasInit}
}

func (p *Parser) parseView() ast.Node {
	kw := p.adv() // view
	name := p.expectKind(lexer.KindIdent, "视图名")
	var typ *ast.TypeRef
	if _, ok := p.eatPunct(":"); ok {
		typ = p.parseType()
	} else {
		p.errSpan(p.spanTo(p.spanOf(kw), p.spanOf(name)),
			"E4004", "`view` 声明必须写类型标注",
			"零拷贝重标的意义就在类型不同",
			"同类型请用 `let 名字: 类型 = 源;`")
	}
	p.expectPunct("=", "视图源表达式")
	init := p.parseExpr()
	if p.atPunct("@") {
		p.errAt(p.cur(), "E4004", "`view` 声明不允许 `@`",
			"`view` 永不分配，只给已有内存起第二个名字",
			"需要新内存请写 `copy 源 @[区域名]`")
	}
	p.expectPunct(";", "视图声明结束")
	return &ast.ViewDecl{Span: p.spanTo(p.spanOf(kw), p.spanOf(p.cur())), Name: name.Text, Type: typ, Init: init}
}

// --------------------------------------------------------------------------- //
// 类型
// --------------------------------------------------------------------------- //

func (p *Parser) parseType() *ast.TypeRef {
	base := p.parseTypeAtom()
	base = p.parseTypePostfix(base)
	if p.atPunct("!") {
		p.adv()
		errT := p.parseTypePostfix(p.parseTypeAtom())
		return &ast.TypeRef{Span: base.Span, KindV: "result", Inner: base, Err: errT}
	}
	return base
}

func (p *Parser) parseTypePostfix(base *ast.TypeRef) *ast.TypeRef {
	for p.atPunct("?") {
		q := p.adv()
		if base.KindV == "result" {
			p.errAt(q, "E2001", "`(T ! E)?` 需要加括号",
				"`?` 与中缀 `!` 的结合没有第二种解释", "写成 `(T ! E)?`")
		}
		base = &ast.TypeRef{Span: base.Span, KindV: "optional", Inner: base}
	}
	return base
}

func (p *Parser) parseTypeAtom() *ast.TypeRef {
	tok := p.cur()
	switch {
	case tok.IsPunct("("):
		p.adv()
		var items []*ast.TypeRef
		for !p.atPunct(")") && !p.atKind(lexer.KindEOF) {
			items = append(items, p.parseType())
			if _, ok := p.eatPunct(","); !ok {
				break
			}
		}
		p.expectPunct(")", "类型列表结束")
		if _, ok := p.eatPunct("->"); ok {
			ret := p.parseType()
			return &ast.TypeRef{Span: p.spanOf(tok), KindV: "fnptr", Params: items, Ret: ret}
		}
		if len(items) == 1 {
			return items[0]
		}
		if len(items) == 0 {
			p.errAt(tok, "E2001", "空括号不是类型", "", "无返回值就别写 `->`")
			return &ast.TypeRef{Span: p.spanOf(tok), KindV: "group"}
		}
		return &ast.TypeRef{Span: p.spanOf(tok), KindV: "group", Items: items}
	case tok.IsPunct("["):
		p.adv()
		if p.atWord("Unsafe") && p.peek(1).IsPunct("]") {
			p.adv()
			p.adv()
			return &ast.TypeRef{Span: p.spanOf(tok), KindV: "unsafe"}
		}
		elem := p.parseType()
		if _, ok := p.eatPunct(";"); ok {
			size := p.parseExpr()
			p.expectPunct("]", "定长数组类型结束")
			return &ast.TypeRef{Span: p.spanOf(tok), KindV: "fixed", Elem: elem, Size: size}
		}
		p.expectPunct("]", "视图类型结束")
		return &ast.TypeRef{Span: p.spanOf(tok), KindV: "view", Elem: elem}
	case tok.Kind == lexer.KindIdent && primTypes[tok.Text] != "":
		p.adv()
		return &ast.TypeRef{Span: p.spanOf(tok), KindV: "prim", Name: primTypes[tok.Text]}
	case tok.IsWord("chan"):
		p.adv()
		elem := p.parseTypePostfix(p.parseTypeAtom())
		return &ast.TypeRef{Span: p.spanOf(tok), KindV: "chan", Elem: elem}
	case tok.IsWord("send", "recv"):
		d := p.adv().Text
		if !p.atWord("chan") {
			p.errAt(p.cur(), "E2001", "`"+d+"` 之后必须是 `chan`", "这里", "")
		} else {
			p.adv()
		}
		elem := p.parseTypePostfix(p.parseTypeAtom())
		return &ast.TypeRef{Span: p.spanOf(tok), KindV: "chan", Elem: elem, Dir: d}
	case tok.IsWord("coro"):
		p.adv()
		var inner, errT *ast.TypeRef
		nxt := p.cur()
		if nxt.Kind == lexer.KindIdent || nxt.IsPunct("[") || nxt.IsPunct("(") {
			inner = p.parseTypePostfix(p.parseTypeAtom())
			if p.atPunct("!") {
				p.adv()
				errT = p.parseTypePostfix(p.parseTypeAtom())
			}
		}
		return &ast.TypeRef{Span: p.spanOf(tok), KindV: "coro", Inner: inner, Err: errT}
	case tok.Kind == lexer.KindIdent:
		return p.parseTypeName()
	}
	got := tok.Text
	if got == "" {
		got = "文件结束"
	}
	p.errAt(tok, "E2001", "此处期望类型，读到 `"+got+"`", "这里", "")
	if !p.atKind(lexer.KindEOF) {
		p.adv()
	}
	return &ast.TypeRef{Span: p.spanOf(tok), KindV: "prim", Name: "i32"}
}

func (p *Parser) parseTypeName() *ast.TypeRef {
	first := p.expectKind(lexer.KindIdent, "类型名")
	name := first.Text
	for p.atPunct(".") && p.peek(1).Kind == lexer.KindIdent {
		p.adv()
		t := p.adv()
		name += "." + t.Text
	}
	return &ast.TypeRef{Span: p.spanOf(first), KindV: "name", Name: name}
}

// --------------------------------------------------------------------------- //
// 语句
// --------------------------------------------------------------------------- //

func (p *Parser) parseBlock() *ast.Block {
	start := p.expectPunct("{", "块开始")
	var stmts []ast.Node
	for !p.atPunct("}") && !p.atKind(lexer.KindEOF) {
		if s := p.parseStmt(); s != nil {
			stmts = append(stmts, s)
		}
	}
	end := p.expectPunct("}", "块结束")
	return &ast.Block{Span: p.spanTo(p.spanOf(start), p.spanOf(end)), Stmts: stmts}
}

func (p *Parser) parseStmt() ast.Node {
	tok := p.cur()
	switch {
	case tok.IsPunct("{"):
		p.errAt(tok, "E2002", "独立裸块语句被禁用",
			"AILA 用 `region 名字 { ... }` 表达局部作用域与回收点",
			"把 `{ ... }` 改成 `region tmp { ... }`")
		return p.parseBlock()
	case tok.Kind == lexer.Kind("let"), tok.Kind == lexer.Kind("var"),
		tok.Kind == lexer.Kind("const"), tok.Kind == lexer.Kind("state"):
		return p.parseBinding(true, false)
	case tok.Kind == lexer.Kind("view"):
		return p.parseView()
	case tok.Kind == lexer.Kind("return"):
		return p.parseReturn()
	case tok.Kind == lexer.Kind("if"):
		return p.parseIf()
	case tok.Kind == lexer.Kind("for"):
		return p.parseFor()
	case tok.Kind == lexer.Kind("match"):
		return p.parseMatch()
	case tok.Kind == lexer.Kind("region"):
		return p.parseRegion()
	case tok.Kind == lexer.Kind("defer"):
		start := p.adv()
		if p.atPunct("{") {
			return &ast.Defer{Span: p.spanOf(start), Body: p.parseBlock()}
		}
		e := p.parseExpr()
		p.expectPunct(";", "defer 语句结束")
		return &ast.Defer{Span: p.spanOf(start), Expr: e}
	case tok.Kind == lexer.Kind("unsafe"):
		start := p.adv()
		return &ast.Unsafe{Span: p.spanOf(start), Body: p.parseBlock()}
	case tok.Kind == lexer.Kind("trap"):
		start := p.adv()
		msg := p.parseExpr()
		p.expectPunct(";", "trap 语句结束")
		return &ast.Trap{Span: p.spanOf(start), Msg: msg}
	case tok.Kind == lexer.Kind("yield"):
		start := p.adv()
		p.expectPunct(";", "yield 语句结束")
		return &ast.Yield{Span: p.spanOf(start)}
	case tok.Kind == lexer.Kind("out"):
		start := p.adv()
		v := p.parseExpr()
		p.expectPunct(";", "out 语句结束")
		return &ast.Out{Span: p.spanOf(start), Value: v}
	case tok.Kind == lexer.Kind("thread"):
		start := p.adv()
		var size ast.Node
		if _, ok := p.eatPunct(":"); ok {
			size = p.parseExpr()
		}
		call := p.parseExpr()
		p.expectPunct(";", "thread 语句结束")
		return &ast.ThreadStart{Span: p.spanOf(start), StackSize: size, Call: call}
	case tok.Kind == lexer.Kind("break"):
		start := p.adv()
		p.expectPunct(";", "")
		return &ast.Break{Span: p.spanOf(start)}
	case tok.Kind == lexer.Kind("continue"):
		start := p.adv()
		p.expectPunct(";", "")
		return &ast.Continue{Span: p.spanOf(start)}
	case tok.Kind == lexer.KindWild && p.peek(1).IsPunct("=", "+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "<<=", ">>="):
		return p.parseAssign()
	case tok.IsPunct("(") && p.looksLikeDestructAssign():
		return p.parseDestructAssign()
	case tok.IsPunct(";"):
		p.adv()
		return nil
	}
	return p.parseAssign()
}

func (p *Parser) looksLikeDestructAssign() bool {
	depth := 0
	expectIdent := true
	for j := p.i; j < len(p.toks); j++ {
		t := p.toks[j]
		if t.Kind == lexer.KindPunct {
			switch t.Text {
			case "(":
				depth++
				continue
			case ")":
				depth--
				if depth == 0 {
					if j+1 < len(p.toks) && p.toks[j+1].IsPunct("=") {
						return true
					}
					return false
				}
				continue
			case ",":
				expectIdent = true
				continue
			default:
				return false
			}
		}
		switch t.Kind {
		case lexer.KindIdent:
			if !expectIdent {
				return false
			}
			expectIdent = false
		case lexer.KindWild:
			expectIdent = false
		default:
			return false
		}
	}
	return false
}

func (p *Parser) parseDestructAssign() ast.Node {
	start := p.expectPunct("(", "")
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
	p.expectPunct("=", "解构赋值")
	value := p.parseExpr()
	end := p.expectPunct(";", "解构赋值结束")
	return &ast.DestructAssign{Span: p.spanTo(p.spanOf(start), p.spanOf(end)), Names: names, Value: value}
}

func (p *Parser) parseReturn() ast.Node {
	start := p.adv() // return
	if _, ok := p.eatPunct(";"); ok {
		return &ast.Return{Span: p.spanOf(start)}
	}
	if p.atKind(lexer.KindWild) && p.peek(1).IsPunct("!") {
		p.adv()
		p.adv()
		e := p.parseExpr()
		end := p.expectPunct(";", "return 语句结束")
		return &ast.Return{Span: p.spanTo(p.spanOf(start), p.spanOf(end)), Err: e}
	}
	v := p.parseExpr()
	end := p.expectPunct(";", "return 语句结束")
	return &ast.Return{Span: p.spanTo(p.spanOf(start), p.spanOf(end)), Value: v}
}

func (p *Parser) parseIf() ast.Node {
	start := p.adv() // if
	cond := p.parseExprNoStructLit()
	then := p.parseBlock()
	var elifs []ast.Pair
	var els *ast.Block
	for p.atKw("else") {
		p.adv()
		if _, ok := p.eatKw("if"); ok {
			c := p.parseExprNoStructLit()
			b := p.parseBlock()
			elifs = append(elifs, ast.P(condPair(c), b))
			continue
		}
		els = p.parseBlock()
		break
	}
	return &ast.If{Span: p.spanTo(p.spanOf(start), p.spanOf(p.cur())), Cond: cond, Then: then,
		Elifs: elifs, Els: els}
}

// condPair 只是为了让 Pair.A 是「表达式节点」而不是接口 nil。
func condPair(n ast.Node) any { return n }

func (p *Parser) parseFor() ast.Node {
	start := p.adv() // for
	if p.atPunct("{") {
		return &ast.ForC{Span: p.spanOf(start), Body: p.parseBlock()}
	}
	if p.atKw("var", "let") {
		if p.atKw("let") {
			p.errAt(p.cur(), "E2011", "三段式 `for` 的声明槽只能是 `var`",
				"计数器必须可变", "写成 `for var i: USize = 0; ...`")
		}
		init := p.parseBinding(false, false).(*ast.LetDecl)
		p.expectPunct(";", "循环条件前")
		cond := p.parseExprNoStructLit()
		p.expectPunct(";", "循环步进前")
		step := p.parseSimpleStmtNoSemi()
		body := p.parseBlock()
		return &ast.ForC{Span: p.spanOf(start), Init: init, Cond: cond, Step: step, Body: body}
	}
	if p.looksLikeForIn() {
		return p.parseForIn(start)
	}
	cond := p.parseExprNoStructLit()
	body := p.parseBlock()
	return &ast.ForC{Span: p.spanOf(start), Cond: cond, Body: body}
}

func (p *Parser) looksLikeForIn() bool {
	if p.cur().Kind != lexer.KindIdent {
		return false
	}
	nxt := p.peek(1)
	return nxt.IsKw("in") || nxt.IsPunct(",", ":")
}

func (p *Parser) parseForIn(start lexer.Token) ast.Node {
	first := p.expectKind(lexer.KindIdent, "迭代变量")
	index := ""
	v := first.Text
	var vtype *ast.TypeRef
	if _, ok := p.eatPunct(":"); ok {
		vtype = p.parseType()
	}
	if _, ok := p.eatPunct(","); ok {
		index = v
		v = p.expectKind(lexer.KindIdent, "元素变量").Text
	}
	if _, ok := p.eatKw("in"); !ok {
		p.errAt(p.cur(), "E2001", "`for` 头部期望 `in`", "这里", "")
	}
	iterable := p.parseExprNoStructLit()
	body := p.parseBlock()
	return &ast.ForIn{Span: p.spanTo(p.spanOf(start), p.spanOf(p.cur())), Var: v, VarType: vtype,
		Index: index, Iterable: iterable, Body: body}
}

func (p *Parser) parseRegion() ast.Node {
	start := p.adv() // region
	name := p.expectKind(lexer.KindIdent, "区域名")
	domain := ""
	var size ast.Node
	if _, ok := p.eatPunct(":"); ok {
		if p.cur().Kind == lexer.KindInt {
			size = p.parseExpr()
		} else if p.cur().Kind == lexer.KindIdent {
			domain = p.adv().Text
			if _, ok := p.eatPunct(","); ok {
				size = p.parseExpr()
			}
		} else {
			size = p.parseExpr()
		}
	}
	body := p.parseBlock()
	return &ast.Region{Span: p.spanTo(p.spanOf(start), p.spanOf(p.cur())), Name: name.Text,
		Domain: domain, Size: size, Body: body}
}

func (p *Parser) parseMatch() ast.Node {
	start := p.adv() // match
	head := p.parseExprNoStructLit()
	if p.atPunct("{") && isPath(head) && p.braceLooksLikeStructLit() {
		p.errSpan(p.spanTo(p.spanOf(p.cur()), p.spanOf(p.cur())),
			"E2003", "match 头有歧义：判别表达式以构造字面量收尾",
			"`{` 既可理解为构造体，也可理解为 match 体",
			"给判别表达式加括号：`match (Shape.Circle { r: 1.0 }) { ... }`")
	}
	p.expectPunct("{", "match 体")
	var arms []*ast.Arm
	for !p.atPunct("}") && !p.atKind(lexer.KindEOF) {
		pat := p.parsePattern()
		p.expectPunct("=>", "match 分支箭头")
		var body ast.Node
		if p.atPunct("{") {
			body = p.parseBlock()
		} else {
			body = p.parseExpr()
		}
		arms = append(arms, &ast.Arm{Span: p.spanOf(p.cur()), Pattern: pat, Body: body})
		if _, isBlock := body.(*ast.Block); isBlock {
			p.eatPunct(",")
			continue
		}
		if _, ok := p.eatPunct(","); !ok {
			if p.atPunct("}") {
				break
			}
			p.errAt(p.cur(), "E2001", "单表达式分支之间需要 `,`", "这里", "")
			if !p.atKind(lexer.KindEOF) {
				p.adv()
			}
		}
	}
	end := p.expectPunct("}", "match 体结束")
	return &ast.Match{Span: p.spanTo(p.spanOf(start), p.spanOf(end)), Head: head, Arms: arms}
}

func isPath(n ast.Node) bool {
	_, ok := n.(*ast.Path)
	return ok
}

func (p *Parser) braceLooksLikeStructLit() bool {
	return p.peek(1).Kind == lexer.KindIdent && p.peek(2).IsPunct(":")
}
