// Package ast 是 AILA 的语法树（docs/specs/0002 §4）。
//
// 设计要点：
//  1. 每个节点都实现 Node（Kind + Fields），Fields 的顺序与字段名**严格对齐 Python 参考实现**，
//     因为 tests/corpus/*.ast.txt 的 S-表达式是逐字符比对的契约；
//  2. 只暴露「解析产物」字段：sema/codegen 才填的字段（ty / path_kind / sig / locals …）不进 Fields；
//  3. 位置信息在 Span 里，不参与序列化。
package ast

import "github.com/ailang/aila/internal/diag"

// Node 是语法树节点。
type Node interface {
	Kind() string
	Fields() []F
}

// F 是一个「字段名 = 值」对。
type F struct {
	Name  string
	Value any
}

// Pair 表示二元组（如 If.elifs、StructLit.fields），序列化成 [a b]。
type Pair struct {
	A any
	B any
}

// P 构造 Pair。
func P(a, b any) Pair { return Pair{A: a, B: b} }

// N 构造字段。
func N(name string, v any) F { return F{Name: name, Value: v} }

// --------------------------------------------------------------------------- //
// 顶层
// --------------------------------------------------------------------------- //

// Attr 是声明属性 `#[...]`。
type Attr struct {
	Span  diag.Span
	Name  string
	Args  []Node
}

func (n *Attr) Kind() string  { return "Attr" }
func (n *Attr) Fields() []F {
	return []F{N("name", n.Name), N("args", n.Args)}
}

// Import 是 import 语句。
type Import struct {
	Span  diag.Span
	Root  string
	Path  []string
	KindV string
	Alias string
}

func (n *Import) Kind() string { return "Import" }
func (n *Import) Fields() []F {
	return []F{N("root", n.Root), N("path", n.Path), N("kind", n.KindV), N("alias", n.Alias)}
}

// TypeRef 是语法层的类型引用。
type TypeRef struct {
	Span  diag.Span
	KindV string // prim / name / view / fixed / unsafe / optional / result / chan / coro / fnptr / group
	Name  string
	Inner *TypeRef
	Elem  *TypeRef
	Size  Node
	Items []*TypeRef
	Dir   string
	Params []*TypeRef
	Ret   *TypeRef
	Err   *TypeRef
}

func (n *TypeRef) Kind() string { return "TypeRef" }
func (n *TypeRef) Fields() []F {
	return []F{
		N("kind", n.KindV), N("name", n.Name), N("inner", optNode(n.Inner)),
		N("elem", optNode(n.Elem)), N("size", optNode(n.Size)), N("items", refs(n.Items)),
		N("dir", n.Dir), N("params", refs(n.Params)), N("ret", optNode(n.Ret)), N("err", optNode(n.Err)),
	}
}

// Field 是结构体字段 / enum 载荷字段。
type Field struct {
	Span diag.Span
	Name string
	Type *TypeRef
}

func (n *Field) Kind() string  { return "Field" }
func (n *Field) Fields() []F   { return []F{N("name", n.Name), N("type", optNode(n.Type))} }

// Variant 是 enum 变体。
type Variant struct {
	Span         diag.Span
	Name         string
	Discriminant Node
	Payload      []*Field
}

func (n *Variant) Kind() string { return "Variant" }
func (n *Variant) Fields() []F {
	return []F{N("name", n.Name), N("discriminant", optNode(n.Discriminant)), N("payload", fields(n.Payload))}
}

// Param 是函数参数。
type Param struct {
	Span   diag.Span
	Name   string
	Type   *TypeRef
	IsSelf bool
}

func (n *Param) Kind() string { return "Param" }
func (n *Param) Fields() []F {
	return []F{N("name", n.Name), N("type", optNode(n.Type)), N("is_self", n.IsSelf)}
}

// Effect 是 @ 子句。
type Effect struct {
	Span  diag.Span
	KindV string
	Names []string
}

func (n *Effect) Kind() string { return "Effect" }
func (n *Effect) Fields() []F  { return []F{N("kind", n.KindV), N("names", n.Names)} }

// Attrs 是所有声明的公共字段。
type Attrs struct {
	Attrs    []*Attr
	IsPublic bool
	// InShell 表示该声明写在 `module { }` 块**外面**（壳）。
	//
	// 可见性阶梯（designs/0001 §2.2）：块内默认文件内、`public` → 子包内；
	// 壳内默认子包内、`public` → 包外。轻量形态（文件无块）顶层声明按壳语义，
	// 因为它们是字面意义上的「写在 module 外面」。
	InShell bool
}

// StructDecl 是 struct 声明。
type StructDecl struct {
	Span diag.Span
	Attrs
	Name   string
	Fields_ []*Field
}

func (n *StructDecl) Kind() string { return "StructDecl" }
func (n *StructDecl) Fields() []F {
	return append(attrsFields(n.Attrs), N("name", n.Name), N("fields", fields(n.Fields_)))
}

// EnumDecl 是 enum 声明。
type EnumDecl struct {
	Span diag.Span
	Attrs
	Name     string
	Base     *TypeRef
	Variants []*Variant
}

func (n *EnumDecl) Kind() string { return "EnumDecl" }
func (n *EnumDecl) Fields() []F {
	return append(attrsFields(n.Attrs), N("name", n.Name), N("base", optNode(n.Base)),
		N("variants", variants(n.Variants)))
}

// TypeAlias 是 type 别名。
type TypeAlias struct {
	Span diag.Span
	Attrs
	Name   string
	Target *TypeRef
}

func (n *TypeAlias) Kind() string { return "TypeAlias" }
func (n *TypeAlias) Fields() []F {
	return append(attrsFields(n.Attrs), N("name", n.Name), N("target", optNode(n.Target)))
}

// FuncDecl 是函数声明。
type FuncDecl struct {
	Span diag.Span
	Attrs
	Name        string
	OwnerType   string
	IsCoroutine bool
	IsThread    bool
	IsExtern    bool
	// IsIntrinsic 表示这是语言原语（#[intrinsic]）：由运行时实现，标准库不出现 C 
	IsIntrinsic bool
	Params      []*Param
	Effects     []*Effect
	Ret         *TypeRef
	Body        *Block
	Variadic    bool
}

func (n *FuncDecl) Kind() string { return "FuncDecl" }
func (n *FuncDecl) Fields() []F {
	return append(attrsFields(n.Attrs),
		N("name", n.Name), N("owner_type", n.OwnerType),
		N("is_coroutine", n.IsCoroutine), N("is_thread", n.IsThread), N("is_extern", n.IsExtern),
		N("params", params(n.Params)), N("effects", effects(n.Effects)),
		N("ret", optNode(n.Ret)), N("body", optNode(n.Body)), N("variadic", n.Variadic))
}

// LetDecl 是 let / var / const / state 绑定。
type LetDecl struct {
	Span diag.Span
	Attrs
	KindV   string
	Name    string
	Type    *TypeRef
	Init    Node
	HasInit bool
}

func (n *LetDecl) Kind() string { return "LetDecl" }
func (n *LetDecl) Fields() []F {
	return append(attrsFields(n.Attrs), N("kind", n.KindV), N("name", n.Name),
		N("type", optNode(n.Type)), N("init", optNode(n.Init)), N("has_init", n.HasInit))
}

// ViewDecl 是 `view` 声明。
type ViewDecl struct {
	Span diag.Span
	Attrs
	Name string
	Type *TypeRef
	Init Node
}

func (n *ViewDecl) Kind() string { return "ViewDecl" }
func (n *ViewDecl) Fields() []F {
	return append(attrsFields(n.Attrs), N("name", n.Name), N("type", optNode(n.Type)), N("init", optNode(n.Init)))
}

// Module 是一个源文件。
type Module struct {
	Span          diag.Span
	Name          string
	Imports       []*Import
	Shell         []Node
	Block         []Node
	IsHot         bool
	IsLightweight bool
}

func (n *Module) Kind() string { return "Module" }
func (n *Module) Fields() []F {
	return []F{N("name", n.Name), N("imports", imports(n.Imports)), N("shell", n.Shell),
		N("block", n.Block), N("is_hot", n.IsHot), N("is_lightweight", n.IsLightweight)}
}

// AllDecls 返回壳 + 块的声明。
func (n *Module) AllDecls() []Node {
	out := make([]Node, 0, len(n.Shell)+len(n.Block))
	out = append(out, n.Shell...)
	out = append(out, n.Block...)
	return out
}
