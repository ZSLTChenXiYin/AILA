package sema

import (
	"github.com/ailang/aila/internal/ast"
	"github.com/ailang/aila/internal/types"
)

// --------------------------------------------------------------------------- //
// 函数体
// --------------------------------------------------------------------------- //

func (c *Checker) checkFunc(fn *ast.FuncDecl) {
	key := fn.Name
	if fn.OwnerType != "" {
		key = fn.OwnerType + "." + fn.Name
	}
	var sig *FuncSig
	if fn.OwnerType == "" {
		sig = c.sigs.Funcs[fn.Name]
	} else {
		sig = c.sigs.Methods[key]
	}
	if sig == nil {
		return
	}
	ctx := &FuncCtx{
		Sig: sig, Scope: NewScope(nil), Regions: map[string]bool{},
		InCoroutine: sig.IsCoroutine, InThread: sig.IsThread,
		Ret: sig.Ret, Err: sig.Err, RetGroup: sig.RetGroup,
	}
	for _, r := range sig.Regions {
		ctx.Regions[r] = true
	}
	prevLocals := c.locals
	c.locals = nil
	for _, p := range sig.Params {
		mutable := containsStr(sig.Muts, p.Name) || containsStr(sig.Moves, p.Name)
		sym := &Symbol{Name: p.Name, Type: p.Type, Mutable: mutable, Kind: "param", IsSelf: p.Name == "self"}
		ctx.Scope.Declare(sym)
		c.ParamSyms[fn] = append(c.ParamSyms[fn], sym)
		c.locals = append(c.locals, sym)
	}
	prev := c.ctx
	c.ctx = ctx
	c.checkBlock(fn.Body, false)
	c.ctx = prev
	c.Locals[fn] = c.locals
	c.locals = prevLocals

	// 有值函数：所有路径必须 return（发散路径豁免）
	if _, isVoid := sig.Ret.(types.Void); !isVoid && sig.Err == nil && sig.RetGroup == nil {
		if !c.blockDiverges(fn.Body) {
			c.err("E3208", fn.Span, "函数 `"+fn.Name+"` 存在未 return 的路径",
				"有值函数必须让每条非发散路径都显式 `return 值;`",
				"补齐缺失路径；发散路径可以 `trap` 或 `return _ ! 错误;`")
		}
	}
}

func (c *Checker) blockDiverges(b *ast.Block) bool {
	if b == nil || len(b.Stmts) == 0 {
		return false
	}
	return c.stmtDiverges(b.Stmts[len(b.Stmts)-1])
}

func (c *Checker) stmtDiverges(s ast.Node) bool {
	switch x := s.(type) {
	case *ast.Return, *ast.Trap:
		return true
	case *ast.Block:
		return c.blockDiverges(x)
	case *ast.Region:
		return c.blockDiverges(x.Body)
	case *ast.Unsafe:
		return c.blockDiverges(x.Body)
	case *ast.If:
		if x.Els == nil {
			return false
		}
		if !c.blockDiverges(x.Then) || !c.blockDiverges(x.Els) {
			return false
		}
		for _, pair := range x.Elifs {
			if b, ok := pair.B.(*ast.Block); ok && !c.blockDiverges(b) {
				return false
			}
		}
		return true
	case *ast.Match:
		if len(x.Arms) == 0 {
			return false
		}
		for _, arm := range x.Arms {
			b, ok := arm.Body.(*ast.Block)
			if !ok || !c.blockDiverges(b) {
				return false
			}
		}
		return true
	}
	return false
}

func (c *Checker) checkBlock(b *ast.Block, newScope bool) {
	if b == nil || c.ctx == nil {
		return
	}
	prev := c.ctx.Scope
	if newScope {
		c.ctx.Scope = NewScope(prev)
	}
	for _, s := range b.Stmts {
		c.checkStmt(s)
	}
	if newScope {
		c.ctx.Scope = prev
	}
}

// --------------------------------------------------------------------------- //
// 语句
// --------------------------------------------------------------------------- //

func (c *Checker) checkStmt(s ast.Node) {
	ctx := c.ctx
	if ctx == nil {
		return
	}
	switch x := s.(type) {
	case *ast.Block:
		c.checkBlock(x, true)
	case *ast.LetDecl:
		c.checkBinding(x)
	case *ast.ViewDecl:
		c.checkView(x)
	case *ast.Return:
		c.checkReturn(x)
	case *ast.If:
		ct := c.checkExpr(x.Cond, types.Bool)
		c.unify(types.Bool, ct, ast.SpanOf(x), "条件")
		c.checkBlock(x.Then, true)
		for _, pair := range x.Elifs {
			if e, ok := pair.A.(ast.Node); ok {
				ct := c.checkExpr(e, types.Bool)
				c.unify(types.Bool, ct, ast.SpanOf(e), "条件")
			}
			if b, ok := pair.B.(*ast.Block); ok {
				c.checkBlock(b, true)
			}
		}
		if x.Els != nil {
			c.checkBlock(x.Els, true)
		}
	case *ast.ForC:
		prev := ctx.Scope
		ctx.Scope = NewScope(prev)
		if x.Init != nil {
			c.checkBinding(x.Init)
		}
		if x.Cond != nil {
			ct := c.checkExpr(x.Cond, types.Bool)
			c.unify(types.Bool, ct, ast.SpanOf(x), "循环条件")
		}
		if x.Step != nil {
			c.checkStmt(x.Step)
		}
		c.loopDepth++
		c.checkBlock(x.Body, true)
		c.loopDepth--
		ctx.Scope = prev
	case *ast.ForIn:
		c.checkForIn(x)
	case *ast.Match:
		c.checkMatch(x, false, nil)
	case *ast.Region:
		c.checkRegion(x)
	case *ast.Defer:
		if x.Body != nil {
			c.checkBlock(x.Body, true)
		} else if x.Expr != nil {
			c.checkExpr(x.Expr, nil)
		}
	case *ast.Unsafe:
		ctx.UnsafeDepth++
		c.checkBlock(x.Body, true)
		ctx.UnsafeDepth--
	case *ast.Trap:
		t := c.checkExpr(x.Msg, types.Text)
		c.unify(types.Text, t, ast.SpanOf(x), "trap 消息")
	case *ast.Yield:
		if !ctx.InCoroutine {
			c.err("E5001", x.Span, "`yield` 只能出现在 `coroutine func` 内", "挂起是协程专属语义", "")
		}
		if len(ctx.Lexical) > 0 {
			c.err("E5004", x.Span, "`region` 块内禁止挂起", "挂起会无限期推迟区域退出回收", "")
		}
	case *ast.ThreadStart:
		if x.StackSize != nil {
			st := c.checkExpr(x.StackSize, types.USize)
			c.requireInteger(st, ast.SpanOf(x), "线程栈大小")
		}
		c.checkExpr(x.Call, nil)
	case *ast.Break:
		if c.loopDepth == 0 {
			c.err("E2001", x.Span, "`break` 只能出现在循环体内", "这里", "")
		}
	case *ast.Continue:
		if c.loopDepth == 0 {
			c.err("E2001", x.Span, "`continue` 只能出现在循环体内", "这里", "")
		}
	case *ast.Assign:
		c.checkAssign(x)
	case *ast.Discard:
		c.checkExpr(x.Value, nil)
	case *ast.DestructAssign:
		c.checkDestructAssign(x)
	case *ast.Out:
		c.checkExpr(x.Value, nil)
	case *ast.ExprStmt:
		t := c.checkExpr(x.Expr, nil)
		if !isVoidOrNever(t) {
			c.err("E3206", ast.SpanOf(x), "有值的表达式被当语句丢弃（类型 `"+typeName(t)+"`）",
				"静默丢值是编译错误", "接收返回值，或显式丢弃：`_ = 表达式;`")
		}
	}
}
