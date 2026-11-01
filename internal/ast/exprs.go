package ast

import "aila/internal/diag"

// --------------------------------------------------------------------------- //
// 表达式
// --------------------------------------------------------------------------- //

// IntLit 是整数字面量。
type IntLit struct {
	Span diag.Span
	Raw  int64
}

func (n *IntLit) Kind() string { return "IntLit" }
func (n *IntLit) Fields() []F  { return []F{N("raw", n.Raw)} }

// FloatLit 是浮点字面量。
type FloatLit struct {
	Span  diag.Span
	Value float64
}

func (n *FloatLit) Kind() string { return "FloatLit" }
func (n *FloatLit) Fields() []F  { return []F{N("value", n.Value)} }

// StrLit 是字符串字面量。
type StrLit struct {
	Span  diag.Span
	Value string
}

func (n *StrLit) Kind() string { return "StrLit" }
func (n *StrLit) Fields() []F  { return []F{N("value", n.Value)} }

// CharLit 是字符字面量（C32）。
type CharLit struct {
	Span  diag.Span
	Value string
}

func (n *CharLit) Kind() string { return "CharLit" }
func (n *CharLit) Fields() []F  { return []F{N("value", n.Value)} }

// BoolLit 是布尔字面量。
type BoolLit struct {
	Span  diag.Span
	Value bool
}

func (n *BoolLit) Kind() string { return "BoolLit" }
func (n *BoolLit) Fields() []F  { return []F{N("value", n.Value)} }

// Path 是名字路径。
type Path struct {
	Span  diag.Span
	Names []string
}

// Dotted 返回点号连接的路径。
func (n *Path) Dotted() string {
	out := ""
	for i, s := range n.Names {
		if i > 0 {
			out += "."
		}
		out += s
	}
	return out
}

func (n *Path) Kind() string { return "Path" }
func (n *Path) Fields() []F  { return []F{N("names", n.Names)} }

// Placeholder 是 `_`。
type Placeholder struct {
	Span diag.Span
}

func (n *Placeholder) Kind() string { return "Placeholder" }
func (n *Placeholder) Fields() []F  { return nil }

// ArrayLit 是数组字面量（重复式 / 列表式）。
type ArrayLit struct {
	Span   diag.Span
	Elems  []Node
	Repeat Node
	Count  Node
}

func (n *ArrayLit) Kind() string { return "ArrayLit" }
func (n *ArrayLit) Fields() []F {
	return []F{N("elems", n.Elems), N("repeat", optNode(n.Repeat)), N("count", optNode(n.Count))}
}

// StructLit 是构造字面量（含 enum 变体构造）。
// LeadingDot 表示 leading-dot 形态 `.名 { 字段 }`（designs/0004 §3.4）：
// 错误成员直接构造，成员名由期待类型的变体集合静态解析。
type StructLit struct {
	Span       diag.Span
	TypePath   []string
	LeadingDot bool
	FieldPairs []Pair
}

func (n *StructLit) Kind() string { return "StructLit" }
func (n *StructLit) Fields() []F {
	return []F{N("type_path", n.TypePath), N("leading_dot", n.LeadingDot), N("fields", pairsAny(n.FieldPairs))}
}

// GroupExpr 是多值组。
type GroupExpr struct {
	Span  diag.Span
	Elems []Node
}

func (n *GroupExpr) Kind() string { return "GroupExpr" }
func (n *GroupExpr) Fields() []F  { return []F{N("elems", n.Elems)} }

// Call 是调用。
type Call struct {
	Span       diag.Span
	Callee     Node
	Args       []Node
	RegionArgs []string
}

func (n *Call) Kind() string { return "Call" }
func (n *Call) Fields() []F {
	return []F{N("callee", optNode(n.Callee)), N("args", n.Args), N("region_args", n.RegionArgs)}
}

// MethodCall 是 `recv.name(args)`。
type MethodCall struct {
	Span       diag.Span
	Recv       Node
	Name       string
	Args       []Node
	RegionArgs []string
}

func (n *MethodCall) Kind() string { return "MethodCall" }
func (n *MethodCall) Fields() []F {
	return []F{N("recv", optNode(n.Recv)), N("name", n.Name), N("args", n.Args),
		N("region_args", n.RegionArgs)}
}

// Index 是下标。
type Index struct {
	Span   diag.Span
	Base   Node
	Index_ Node
}

func (n *Index) Kind() string { return "Index" }
func (n *Index) Fields() []F {
	return []F{N("base", optNode(n.Base)), N("index", optNode(n.Index_))}
}

// FieldAccess 是字段访问。
type FieldAccess struct {
	Span diag.Span
	Base Node
	Name string
}

func (n *FieldAccess) Kind() string { return "FieldAccess" }
func (n *FieldAccess) Fields() []F {
	return []F{N("base", optNode(n.Base)), N("name", n.Name)}
}

// Binary 是二元运算。
type Binary struct {
	Span  diag.Span
	Op    string
	Left  Node
	Right Node
}

func (n *Binary) Kind() string { return "Binary" }
func (n *Binary) Fields() []F {
	return []F{N("op", n.Op), N("left", optNode(n.Left)), N("right", optNode(n.Right))}
}

// Unary 是一元运算。
type Unary struct {
	Span    diag.Span
	Op      string
	Operand Node
}

func (n *Unary) Kind() string { return "Unary" }
func (n *Unary) Fields() []F {
	return []F{N("op", n.Op), N("operand", optNode(n.Operand))}
}

// TryPropagate 是后缀 `?`。
type TryPropagate struct {
	Span    diag.Span
	Operand Node
}

func (n *TryPropagate) Kind() string { return "TryPropagate" }
func (n *TryPropagate) Fields() []F  { return []F{N("operand", optNode(n.Operand))} }

// Move 是 `move x`。
type Move struct {
	Span    diag.Span
	Operand Node
}

func (n *Move) Kind() string { return "Move" }
func (n *Move) Fields() []F  { return []F{N("operand", optNode(n.Operand))} }

// Copy 是 `copy x @[r]`。
type Copy struct {
	Span       diag.Span
	Operand    Node
	RegionArgs []string
}

func (n *Copy) Kind() string { return "Copy" }
func (n *Copy) Fields() []F {
	return []F{N("operand", optNode(n.Operand)), N("region_args", n.RegionArgs)}
}

// Resume 是 `resume h`。
type Resume struct {
	Span    diag.Span
	Operand Node
}

func (n *Resume) Kind() string { return "Resume" }
func (n *Resume) Fields() []F  { return []F{N("operand", optNode(n.Operand))} }

// Recv 是 `<-ch`。
type Recv struct {
	Span diag.Span
	Chan Node
}

func (n *Recv) Kind() string { return "Recv" }
func (n *Recv) Fields() []F  { return []F{N("chan", optNode(n.Chan))} }

// Alloc 是分配点 `[...]@r`。
type Alloc struct {
	Span   diag.Span
	Expr   Node
	Region string
}

func (n *Alloc) Kind() string { return "Alloc" }
func (n *Alloc) Fields() []F {
	return []F{N("expr", optNode(n.Expr)), N("region", n.Region)}
}

// Match 是 match 表达式。
type Match struct {
	Span diag.Span
	Head Node
	Arms []*Arm

	// CompileTime 表示这是 `#compile.match`（designs/0006 §1）：判别式必须是编译期常量，
	// 命中分支之外的臂只在代码生成阶段省略（类型检查照常覆盖全部臂）。
	CompileTime bool
	// PruneArm 是 sema 的剪枝结论（仅 CompileTime 为真时有效）：
	//   -1 = 未剪；否则为命中的臂下标。
	PruneArm int
}

func (n *Match) Kind() string { return "Match" }
func (n *Match) Fields() []F {
	fields := []F{N("head", optNode(n.Head)), N("arms", arms(n.Arms))}
	if n.CompileTime {
		fields = append(fields, N("compile_time", true), N("prune_arm", n.PruneArm))
	}
	return fields
}
