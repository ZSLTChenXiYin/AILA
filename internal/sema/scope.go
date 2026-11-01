package sema

import (
	"aila/internal/ast"
	"aila/internal/diag"
	"aila/internal/types"
)

// Symbol 是一个绑定。
type Symbol struct {
	Name         string
	Type         types.Type
	Mutable      bool
	Kind         string // local | param | global | state
	Moved        bool
	Used         bool
	IsSelf       bool
	OriginRegion string // 分配来源区域（逃逸检查）
}

// Scope 是词法作用域链。
type Scope struct {
	parent  *Scope
	symbols map[string]*Symbol
}

// NewScope 新建作用域（parent 可为 nil）。
func NewScope(parent *Scope) *Scope {
	return &Scope{parent: parent, symbols: map[string]*Symbol{}}
}

// Declare 声明一个符号；重名返回 false。
func (s *Scope) Declare(sym *Symbol) bool {
	if _, dup := s.symbols[sym.Name]; dup {
		return false
	}
	s.symbols[sym.Name] = sym
	return true
}

// Lookup 沿作用域链查找。
func (s *Scope) Lookup(name string) *Symbol {
	for cur := s; cur != nil; cur = cur.parent {
		if sym, ok := cur.symbols[name]; ok {
			return sym
		}
	}
	return nil
}

// FuncCtx 是一个函数体的检查上下文。
type FuncCtx struct {
	Sig         *FuncSig
	Decl        *ast.FuncDecl // 当前函数声明（调用链记录键）
	Scope       *Scope
	Regions     map[string]bool // @region 参数
	Lexical     []string        // 体内 region 块
	InCoroutine bool
	InThread    bool
	UnsafeDepth int
	WroteGlobal bool // 本函数体写过模块级可变状态（E5003 判据）
	Ret         types.Type
	Err         types.Type
	RetGroup    []types.Type
}

// RegionVisible 判断区域名在当前位置是否可见（@region 参数 / 体内 region 块 / @static）。
func (c *FuncCtx) RegionVisible(name string) bool {
	if name == "static" || name == "" {
		return name == "static"
	}
	if c.Regions != nil && c.Regions[name] {
		return true
	}
	for _, r := range c.Lexical {
		if r == name {
			return true
		}
	}
	return false
}

// Checker 是语义检查器。
type Checker struct {
	mod       *ast.Module
	table     *Table
	sigs      *Sigs
	bag       *diag.Bag
	ctx       *FuncCtx
	loopDepth int

	// 类型注解（sema 产物，不进 AST 序列化）
	Types        map[ast.Node]types.Type
	Bindings     map[*ast.LetDecl]types.Type
	Syms         map[*ast.LetDecl]*Symbol
	Origins      map[ast.Node]string
	PathSyms     map[*ast.Path]*Symbol
	PathGlobals  map[*ast.Path]*GlobalInfo
	PathConsts   map[*ast.Path]*ConstInfo
	PathFuncs    map[*ast.Path]*FuncSig
	PathVariants map[*ast.Path]variantRef
	PathBuiltin  map[*ast.Path]string
	CallSigs     map[ast.Node]*FuncSig
	CallKinds    map[ast.Node]string
	CallRecv     map[ast.Node]ast.Node
	// CoroCalls 标记「coroutine@ 的启动点 call」：豁免「直接调用协程函数」检查。
	CoroCalls map[ast.Node]bool
	// ErrReturns 标记 v2 `return .成员;`（designs/0004 §3.4）：
	// 返回位置自动装载错误态，codegen 走 buildResult(ok=false)。
	ErrReturns map[*ast.Return]bool
	// SelectBinds 记录 select case 的绑定符号（designs/0004 §6）：
	// chan 事件省绑定时隐式绑定 v；nil = 无绑定（事件产出丢弃）。
	SelectBinds map[*ast.SelectCase]*Symbol
	// LayerRefs 记录层 API 调用的静态解释（designs/0004 §2.4）。
	LayerRefs map[*ast.LayerCall]*LayerRef
	// CallCallee 记录 fnptr 间接调用的「函数指针值」表达式（字段形态由 sema 合成）。
	CallCallee   map[ast.Node]ast.Node
	BinOperands  map[ast.Node]types.Type
	TryModes     map[ast.Node]string
	StructLits   map[*ast.StructLit]*types.StructInfo
	VariantLits  map[*ast.StructLit]variantRef
	ArmVariants  map[*ast.Arm]variantRef
	MatchResults map[*ast.Match]types.Type
	ParamSyms    map[*ast.FuncDecl][]*Symbol
	Locals       map[*ast.FuncDecl][]*Symbol

	locals []*Symbol

	// 模块图上下文（多模块语义）
	modName string
	pkg     string // 当前模块所属包（跨包可见性检查用）
	imports map[string]*Unit
	units   map[string]*Unit

	// 代码生成注解（见 cginfo.go）
	FieldRefs    map[*ast.FieldAccess]*FieldRef
	FieldChains  map[*ast.Path]*FieldChain
	IndexRefs    map[*ast.Index]*IndexRef
	ForRefs      map[*ast.ForIn]*ForRef
	ArmRefs      map[*ast.Arm]*ArmRef
	MatchRefs    map[*ast.Match]*MatchRef
	LitRefs      map[*ast.StructLit]*LitRef
	ArrRefs      map[*ast.ArrayLit]*ArrayRef
	AllocRefs    map[*ast.Alloc]*AllocRef
	TryRefs      map[*ast.TryPropagate]*TryRef
	DestructRefs map[ast.Node][]*Symbol
	// 线程调用链检查（E5003，designs/0001 §10.7 全局可变状态禁令）
	FuncDecls    map[*FuncSig]*ast.FuncDecl
	CallEdges    map[*ast.FuncDecl][]*FuncSig
	WroteGlobals map[*ast.FuncDecl]bool

	// ConstVals 记录「函数内 let 绑到 const-expr」的整数值（designs/0006 §1.2）：
	// `#compile.if` / `#compile.match` 的条件可以是这类 let 绑定，它天然不可重新赋值。
	ConstVals map[*Symbol]int64
}

// Check 检查模块级值与全部函数体，返回可查询的类型注解。
func Check(mod *ast.Module, table *Table, sigs *Sigs, bag *diag.Bag) *Checker {
	c := NewChecker(mod, table, sigs, bag)
	c.Run()
	return c
}

// NewChecker 构造检查器（多模块场景要先挂上导入上下文，再 Run）。
func NewChecker(mod *ast.Module, table *Table, sigs *Sigs, bag *diag.Bag) *Checker {
	return &Checker{
		mod: mod, table: table, sigs: sigs, bag: bag,
		Types:        map[ast.Node]types.Type{},
		Bindings:     map[*ast.LetDecl]types.Type{},
		Syms:         map[*ast.LetDecl]*Symbol{},
		Origins:      map[ast.Node]string{},
		PathSyms:     map[*ast.Path]*Symbol{},
		PathGlobals:  map[*ast.Path]*GlobalInfo{},
		PathConsts:   map[*ast.Path]*ConstInfo{},
		PathFuncs:    map[*ast.Path]*FuncSig{},
		PathVariants: map[*ast.Path]variantRef{},
		PathBuiltin:  map[*ast.Path]string{},
		CallSigs:     map[ast.Node]*FuncSig{},
		CallKinds:    map[ast.Node]string{},
		CallRecv:     map[ast.Node]ast.Node{},
		CallCallee:   map[ast.Node]ast.Node{},
		CoroCalls:    map[ast.Node]bool{},
		BinOperands:  map[ast.Node]types.Type{},
		TryModes:     map[ast.Node]string{},
		ErrReturns:   map[*ast.Return]bool{},
		SelectBinds:  map[*ast.SelectCase]*Symbol{},
		LayerRefs:    map[*ast.LayerCall]*LayerRef{},
		StructLits:   map[*ast.StructLit]*types.StructInfo{},
		VariantLits:  map[*ast.StructLit]variantRef{},
		ArmVariants:  map[*ast.Arm]variantRef{},
		MatchResults: map[*ast.Match]types.Type{},
		ParamSyms:    map[*ast.FuncDecl][]*Symbol{},
		Locals:       map[*ast.FuncDecl][]*Symbol{},
		FieldRefs:    map[*ast.FieldAccess]*FieldRef{},
		FieldChains:  map[*ast.Path]*FieldChain{},
		IndexRefs:    map[*ast.Index]*IndexRef{},
		ForRefs:      map[*ast.ForIn]*ForRef{},
		ArmRefs:      map[*ast.Arm]*ArmRef{},
		MatchRefs:    map[*ast.Match]*MatchRef{},
		LitRefs:      map[*ast.StructLit]*LitRef{},
		ArrRefs:      map[*ast.ArrayLit]*ArrayRef{},
		AllocRefs:    map[*ast.Alloc]*AllocRef{},
		TryRefs:      map[*ast.TryPropagate]*TryRef{},
		DestructRefs: map[ast.Node][]*Symbol{},
		FuncDecls:    map[*FuncSig]*ast.FuncDecl{},
		CallEdges:    map[*ast.FuncDecl][]*FuncSig{},
		WroteGlobals: map[*ast.FuncDecl]bool{},
		ConstVals:    map[*Symbol]int64{},
	}
}

// Run 跑模块级值与全部函数体。
func (c *Checker) Run() {
	c.checkForwardDeclarations()
	c.checkModuleValues()
	for _, d := range c.mod.AllDecls() {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Body != nil {
			c.checkFunc(fn)
		}
	}
	c.finishThreads()
}

// finishThreads 跑线程创建的调用链检查（designs/0001 §10.7）：
// 线程入口调用链内写模块级可变状态 = E5003。
// v2：启动点经 `#runtime.spawn(f, args)`（designs/0004 §2.4）。
func (c *Checker) finishThreads() {
	for x, ref := range c.LayerRefs {
		if ref == nil || ref.Kind != "spawn" || ref.Sig == nil {
			continue
		}
		if c.chainTouchesGlobal(ref.Sig, map[*ast.FuncDecl]bool{}) {
			c.err("E5003", x.Span, "线程入口调用链内写了模块级可变状态",
				"数据竞争在语言层构造性不可能（B6 无 UB）",
				"线程共享状态走通道 / atomics；模块级 var 只读放行")
		}
	}
}

// chainTouchesGlobal 沿调用边搜线程入口可达函数是否写模块级可变状态。
func (c *Checker) chainTouchesGlobal(sig *FuncSig, seen map[*ast.FuncDecl]bool) bool {
	decl, ok := c.FuncDecls[sig]
	if !ok {
		return false // #[c] extern / 跨单元：本单元检查器不掌握，v0 放行
	}
	if seen[decl] {
		return false
	}
	seen[decl] = true
	if c.WroteGlobals[decl] {
		return true
	}
	for _, cs := range c.CallEdges[decl] {
		if c.chainTouchesGlobal(cs, seen) {
			return true
		}
	}
	return false
}

// SetModuleContext 挂上模块图上下文（跨模块名字解析用）。
func (c *Checker) SetModuleContext(modName string, imports map[string]*Unit, units map[string]*Unit) {
	c.modName = modName
	c.imports = imports
	c.units = units
}

// Type 返回表达式的推断类型（未标注返回 nil）。
func (c *Checker) Type(n ast.Node) types.Type { return c.Types[n] }

// --------------------------------------------------------------------------- //
// 诊断与类型工具
// --------------------------------------------------------------------------- //

func (c *Checker) err(code string, span diag.Span, msg, label, help string) {
	c.bag.Error(code, span, msg, label, help)
}

// typeName 是 describe 的别名（拼写与 Python 一致）。
func typeName(t types.Type) string { return types.TypeName(t) }

// sameType 判等（结构化 + 字面量多态）。
func sameType(a, b types.Type) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if _, ok := a.(types.UntypedInt); ok {
		if _, ok2 := b.(types.UntypedInt); ok2 {
			return true
		}
		if types.IsInteger(b) {
			return true
		}
		if p, ok3 := b.(types.Prim); ok3 && p.K == "c32" {
			return true
		}
		return false
	}
	if _, ok := b.(types.UntypedInt); ok {
		return sameType(b, a)
	}
	if _, ok := a.(types.UntypedFloat); ok {
		if _, ok2 := b.(types.UntypedFloat); ok2 {
			return true
		}
		return types.IsFloat(b)
	}
	if _, ok := b.(types.UntypedFloat); ok {
		return sameType(b, a)
	}
	switch x := a.(type) {
	case types.Prim:
		y, ok := b.(types.Prim)
		return ok && x.K == y.K
	case types.View:
		y, ok := b.(types.View)
		return ok && sameType(x.Elem, y.Elem)
	case types.Fixed:
		y, ok := b.(types.Fixed)
		return ok && x.N == y.N && sameType(x.Elem, y.Elem)
	case types.UnsafeView:
		_, ok := b.(types.UnsafeView)
		return ok
	case types.Optional:
		y, ok := b.(types.Optional)
		return ok && sameType(x.Inner, y.Inner)
	case types.Result:
		y, ok := b.(types.Result)
		return ok && sameType(x.Ok, y.Ok) && sameType(x.Err, y.Err)
	case types.Chan:
		y, ok := b.(types.Chan)
		return ok && x.Dir == y.Dir && sameType(x.Elem, y.Elem)
	case types.Coro:
		y, ok := b.(types.Coro)
		return ok && sameType(x.Ok, y.Ok) && sameType(x.Err, y.Err)
	case types.FnPtr:
		y, ok := b.(types.FnPtr)
		if !ok || len(x.Params) != len(y.Params) {
			return false
		}
		for i := range x.Params {
			if !sameType(x.Params[i], y.Params[i]) {
				return false
			}
		}
		return sameType(x.Ret, y.Ret) && sameType(x.Err, y.Err)
	case types.Group:
		y, ok := b.(types.Group)
		if !ok || len(x.Items) != len(y.Items) {
			return false
		}
		for i := range x.Items {
			if !sameType(x.Items[i], y.Items[i]) {
				return false
			}
		}
		return true
	case types.Named:
		y, ok := b.(types.Named)
		return ok && x.Info == y.Info
	case types.ErrorSpace:
		y, ok := b.(types.ErrorSpace)
		return ok && x.Info == y.Info
	case types.Void:
		_, ok := b.(types.Void)
		return ok
	case types.Never:
		_, ok := b.(types.Never)
		return ok
	}
	return false
}

// unify 校验期望类型与实际类型（零隐式转换，字面量按上下文定型）。
func (c *Checker) unify(expected, actual types.Type, span diag.Span, what string) {
	if expected == nil || actual == nil {
		return
	}
	if _, isVoid := expected.(types.Void); isVoid {
		return
	}
	if _, isNever := actual.(types.Never); isNever {
		return
	}
	if ui, ok := actual.(types.UntypedInt); ok {
		if types.IsInteger(expected) || isC32(expected) {
			bits := types.IntWidth(expected, 8)
			if bits == 0 {
				bits = 64
			}
			signed := types.IsSigned(expected)
			if !signed && ui.Value < 0 && !isC32(expected) {
				if !ui.BitPattern {
					c.err("E3204", span, "无符号目标不能承接负数",
						"`-N` 落在有符号域", "要最大值请写十六进制字面量，如 `0xFF`")
					return
				}
				// 位型字面量（≥ 2^63）：只有 64 位无符号能直落
				if bits < 64 {
					c.err("E3201", span, "字面量超出目标类型值域", what,
						"≥ 2^63 的字面量只能直接落 U64 / USize")
					return
				}
				return
			}
			if ui.BitPattern {
				// 有符号 / C32 目标：位型值 ≥ 2^63 落不下
				c.err("E3201", span, "字面量超出目标类型值域", what,
					"≥ 2^63 的字面量只能直接落 U64 / USize")
				return
			}
			// 64 位目标：int64 字面量恒可容纳（移位 64 在 Go 里会溢出成 0，必须特判）
			if bits >= 64 {
				return
			}
			limit := int64(1) << uint(bits)
			if signed {
				limit = int64(1) << uint(bits-1)
			}
			if ui.Value < 0 {
				if -ui.Value >= limit {
					c.err("E3201", span, "字面量超出目标类型值域", what, "")
				}
				return
			}
			if ui.Value >= limit {
				c.err("E3201", span, "字面量超出目标类型值域", what, "")
			}
			return
		}
		if types.IsFloat(expected) {
			c.err("E3201", span, "整数常量不能直接当作 `"+typeName(expected)+"`",
				"整数↔浮点一律走显式转换", "写成 `1.0`（浮点字面量）")
			return
		}
		c.err("E3201", span, "整数常量不能当作 `"+typeName(expected)+"`", what, "")
		return
	}
	if _, ok := actual.(types.UntypedFloat); ok {
		if types.IsFloat(expected) {
			return
		}
		c.err("E3201", span, "浮点常量不能当作 `"+typeName(expected)+"`", what, "")
		return
	}
	if sameType(expected, actual) {
		return
	}
	c.err("E3201", span, "类型不匹配：期望 `"+typeName(expected)+"`，实得 `"+typeName(actual)+"`",
		what, "AILA 零隐式转换：无损方向用 `as`，有损方向用 `trunc`")
}

func isC32(t types.Type) bool {
	p, ok := t.(types.Prim)
	return ok && p.K == "c32"
}

// --------------------------------------------------------------------------- //
// 模块级值
// --------------------------------------------------------------------------- //

func (c *Checker) checkModuleValues() {
	for _, d := range c.mod.AllDecls() {
		decl, ok := d.(*ast.LetDecl)
		if !ok || decl.Init == nil {
			continue
		}
		gi, isGlobal := c.sigs.Globals[decl.Name]
		var expected types.Type
		if isGlobal {
			expected = gi.Type
		} else if ci, isConst := c.sigs.Consts[decl.Name]; isConst {
			expected = ci.Type
		}
		c.ctx = &FuncCtx{Scope: NewScope(nil), Regions: map[string]bool{}}
		t := c.checkExpr(decl.Init, expected)
		if expected != nil {
			c.unify(expected, t, decl.Span, "模块级初始化器")
		}
		c.Bindings[decl] = t
		c.ctx = nil
	}
}
