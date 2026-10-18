package sema

import (
	"aila/internal/ast"
	"aila/internal/diag"
	"aila/internal/types"
)

// Symbol 是一个绑定。
type Symbol struct {
	Name     string
	Type     types.Type
	Mutable  bool
	Kind     string // local | param | global | state
	Moved    bool
	Used     bool
	IsSelf   bool
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
	Scope       *Scope
	Regions     map[string]bool // @region 参数
	Lexical     []string        // 体内 region 块
	InCoroutine bool
	InThread    bool
	UnsafeDepth int
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
	mod      *ast.Module
	table    *Table
	sigs     *Sigs
	bag      *diag.Bag
	ctx      *FuncCtx
	loopDepth int

	// 类型注解（sema 产物，不进 AST 序列化）
	Types       map[ast.Node]types.Type
	Bindings    map[*ast.LetDecl]types.Type
	Syms        map[*ast.LetDecl]*Symbol
	Origins     map[ast.Node]string
	PathSyms    map[*ast.Path]*Symbol
	PathGlobals map[*ast.Path]*GlobalInfo
	PathConsts  map[*ast.Path]*ConstInfo
	PathFuncs   map[*ast.Path]*FuncSig
	PathVariants map[*ast.Path]variantRef
	PathBuiltin map[*ast.Path]string
	CallSigs    map[ast.Node]*FuncSig
	CallKinds   map[ast.Node]string
	CallRecv    map[ast.Node]ast.Node
	BinOperands map[ast.Node]types.Type
	CastTypes   map[ast.Node][2]types.Type
	TryModes    map[ast.Node]string
	StructLits  map[*ast.StructLit]*types.StructInfo
	VariantLits map[*ast.StructLit]variantRef
	BuiltinLits  map[*ast.StructLit]string
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
	LenRefs      map[*ast.Len]types.Type
	DestructRefs map[*ast.DestructAssign][]*Symbol
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
		Types:       map[ast.Node]types.Type{},
		Bindings:    map[*ast.LetDecl]types.Type{},
		Syms:        map[*ast.LetDecl]*Symbol{},
		Origins:     map[ast.Node]string{},
		PathSyms:    map[*ast.Path]*Symbol{},
		PathGlobals: map[*ast.Path]*GlobalInfo{},
		PathConsts:  map[*ast.Path]*ConstInfo{},
		PathFuncs:   map[*ast.Path]*FuncSig{},
		PathVariants: map[*ast.Path]variantRef{},
		PathBuiltin: map[*ast.Path]string{},
		CallSigs:    map[ast.Node]*FuncSig{},
		CallKinds:   map[ast.Node]string{},
		CallRecv:    map[ast.Node]ast.Node{},
		BinOperands: map[ast.Node]types.Type{},
		CastTypes:   map[ast.Node][2]types.Type{},
		TryModes:    map[ast.Node]string{},
		StructLits:  map[*ast.StructLit]*types.StructInfo{},
		VariantLits: map[*ast.StructLit]variantRef{},
		BuiltinLits:  map[*ast.StructLit]string{},
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
		LenRefs:      map[*ast.Len]types.Type{},
		DestructRefs: map[*ast.DestructAssign][]*Symbol{},
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
				c.err("E3204", span, "无符号目标不能承接负数",
					"`-N` 落在有符号域", "要最大值请写十六进制字面量，如 `0xFF`")
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
