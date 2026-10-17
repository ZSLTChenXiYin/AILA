package ast

import "github.com/aila-lang/aila/internal/diag"

// --------------------------------------------------------------------------- //
// 语句
// --------------------------------------------------------------------------- //

// Block 是块。
type Block struct {
	Span  diag.Span
	Stmts []Node
}

func (n *Block) Kind() string { return "Block" }
func (n *Block) Fields() []F  { return []F{N("stmts", n.Stmts)} }

// Return 是 return 语句。
type Return struct {
	Span  diag.Span
	Value Node
	Err   Node
}

func (n *Return) Kind() string { return "Return" }
func (n *Return) Fields() []F {
	return []F{N("value", optNode(n.Value)), N("error", optNode(n.Err))}
}

// If 是 if 语句（恒为语句）。
type If struct {
	Span  diag.Span
	Cond  Node
	Then  *Block
	Elifs []Pair
	Els   *Block
}

func (n *If) Kind() string { return "If" }
func (n *If) Fields() []F {
	return []F{N("cond", optNode(n.Cond)), N("then", optNode(n.Then)),
		N("elifs", n.Elifs), N("els", optNode(n.Els))}
}

// ForC 是三段式 / 条件 / 无限循环。
type ForC struct {
	Span  diag.Span
	Init  *LetDecl
	Cond  Node
	Step  Node
	Body  *Block
}

func (n *ForC) Kind() string { return "ForC" }
func (n *ForC) Fields() []F {
	return []F{N("init", optNode(n.Init)), N("cond", optNode(n.Cond)),
		N("step", optNode(n.Step)), N("body", optNode(n.Body))}
}

// ForIn 是 `for x in v` / `for i, x in v`。
type ForIn struct {
	Span     diag.Span
	Var      string
	VarType  *TypeRef
	Index    string
	Iterable Node
	Body     *Block
}

func (n *ForIn) Kind() string { return "ForIn" }
func (n *ForIn) Fields() []F {
	return []F{N("var", n.Var), N("var_type", optNode(n.VarType)), N("index", n.Index),
		N("iterable", optNode(n.Iterable)), N("body", optNode(n.Body))}
}

// Arm 是 match 的一个分支。
type Arm struct {
	Span    diag.Span
	Pattern Node
	Body    Node
}

func (n *Arm) Kind() string { return "Arm" }
func (n *Arm) Fields() []F {
	return []F{N("pattern", optNode(n.Pattern)), N("body", optNode(n.Body))}
}

// Region 是 region 块。
type Region struct {
	Span   diag.Span
	Name   string
	Domain string
	Size   Node
	Body   *Block
}

func (n *Region) Kind() string { return "Region" }
func (n *Region) Fields() []F {
	return []F{N("name", n.Name), N("domain", n.Domain), N("size", optNode(n.Size)), N("body", optNode(n.Body))}
}

// Defer 是 defer。
type Defer struct {
	Span diag.Span
	Body *Block
	Expr Node
}

func (n *Defer) Kind() string { return "Defer" }
func (n *Defer) Fields() []F {
	return []F{N("body", optNode(n.Body)), N("expr", optNode(n.Expr))}
}

// Unsafe 是 unsafe 块。
type Unsafe struct {
	Span diag.Span
	Body *Block
}

func (n *Unsafe) Kind() string { return "Unsafe" }
func (n *Unsafe) Fields() []F  { return []F{N("body", optNode(n.Body))} }

// Trap 是 trap 语句。
type Trap struct {
	Span diag.Span
	Msg  Node
}

func (n *Trap) Kind() string { return "Trap" }
func (n *Trap) Fields() []F  { return []F{N("message", optNode(n.Msg))} }

// Yield 是挂起点。
type Yield struct {
	Span diag.Span
}

func (n *Yield) Kind() string { return "Yield" }
func (n *Yield) Fields() []F  { return nil }

// ThreadStart 是 `thread: 栈大小 调用();`。
type ThreadStart struct {
	Span      diag.Span
	StackSize Node
	Call      Node
}

func (n *ThreadStart) Kind() string { return "ThreadStart" }
func (n *ThreadStart) Fields() []F {
	return []F{N("stack_size", optNode(n.StackSize)), N("call", optNode(n.Call))}
}

// Break 是 break。
type Break struct {
	Span diag.Span
}

func (n *Break) Kind() string { return "Break" }
func (n *Break) Fields() []F  { return nil }

// Continue 是 continue。
type Continue struct {
	Span diag.Span
}

func (n *Continue) Kind() string { return "Continue" }
func (n *Continue) Fields() []F  { return nil }

// Assign 是赋值 / 复合赋值。
type Assign struct {
	Span   diag.Span
	Target Node
	Op     string
	Value  Node
}

func (n *Assign) Kind() string { return "Assign" }
func (n *Assign) Fields() []F {
	return []F{N("target", optNode(n.Target)), N("op", n.Op), N("value", optNode(n.Value))}
}

// Discard 是 `_ = expr;`。
type Discard struct {
	Span  diag.Span
	Value Node
}

func (n *Discard) Kind() string { return "Discard" }
func (n *Discard) Fields() []F  { return []F{N("value", optNode(n.Value))} }

// DestructAssign 是 `(a, b) = f();`。
type DestructAssign struct {
	Span  diag.Span
	Names []string
	Value Node
}

func (n *DestructAssign) Kind() string { return "DestructAssign" }
func (n *DestructAssign) Fields() []F {
	return []F{N("names", n.Names), N("value", optNode(n.Value))}
}

// ExprStmt 是表达式语句。
type ExprStmt struct {
	Span diag.Span
	Expr Node
}

func (n *ExprStmt) Kind() string { return "ExprStmt" }
func (n *ExprStmt) Fields() []F  { return []F{N("expr", optNode(n.Expr))} }

// Out 是块值产出。
type Out struct {
	Span  diag.Span
	Value Node
}

func (n *Out) Kind() string { return "Out" }
func (n *Out) Fields() []F  { return []F{N("value", optNode(n.Value))} }

// --------------------------------------------------------------------------- //
// 模式
// --------------------------------------------------------------------------- //

// Wildcard 是 `_`。
type Wildcard struct {
	Span diag.Span
}

func (n *Wildcard) Kind() string { return "Wildcard" }
func (n *Wildcard) Fields() []F  { return nil }

// LitPattern 是字面量模式。
type LitPattern struct {
	Span     diag.Span
	Value    any
	Negative bool
	KindV    string
}

func (n *LitPattern) Kind() string { return "LitPattern" }
func (n *LitPattern) Fields() []F {
	return []F{N("value", n.Value), N("negative", n.Negative), N("kind", n.KindV)}
}

// PathPattern 是路径模式（const 名 / 变体裸名 / 多值组叶）。
type PathPattern struct {
	Span diag.Span
	Path []string
}

func (n *PathPattern) Kind() string { return "PathPattern" }
func (n *PathPattern) Fields() []F  { return []F{N("path", n.Path)} }

// FieldPat 是变体模式里的一个字段。
type FieldPat struct {
	Span    diag.Span
	Name    string
	Value   *LitPattern
	Binding string
	Group   *GroupPattern
}

func (n *FieldPat) Kind() string { return "FieldPat" }
func (n *FieldPat) Fields() []F {
	return []F{N("name", n.Name), N("value", optNode(n.Value)), N("binding", n.Binding),
		N("group", optNode(n.Group))}
}

// GroupPattern 是多值组叶 `(a, b)`。
type GroupPattern struct {
	Span  diag.Span
	Names []string
}

func (n *GroupPattern) Kind() string { return "GroupPattern" }
func (n *GroupPattern) Fields() []F  { return []F{N("names", n.Names)} }

// VariantPattern 是变体模式。
type VariantPattern struct {
	Span   diag.Span
	Path   []string
	Fields_ []*FieldPat
}

func (n *VariantPattern) Kind() string { return "VariantPattern" }
func (n *VariantPattern) Fields() []F {
	return []F{N("path", n.Path), N("fields", fieldPats(n.Fields_))}
}
