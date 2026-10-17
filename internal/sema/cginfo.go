package sema

import (
	"github.com/ailang/aila/internal/ast"
	"github.com/ailang/aila/internal/types"
)

// 本文件存放「代码生成注解」：语义检查阶段顺带记下 codegen 需要的静态事实。
// 这些信息不进 AST 序列化（快照判据不受影响），只在 Go 进程内使用。

// FieldRef 是一次字段访问的静态信息。
type FieldRef struct {
	Base  types.Type
	Info  *types.StructInfo
	Index int
	Type  types.Type
}

// IndexRef 是一次下标的静态信息。
//
// Kind: "view"（`[T]`）、"text"（Text）、"fixed"（`[T; N]`）。
type IndexRef struct {
	Base types.Type
	Kind string
	Elem types.Type
}

// ForRef 是 for-in 的静态信息。
type ForRef struct {
	Iter  types.Type
	Elem  types.Type
	Var   *Symbol
	Index *Symbol
}

// ArmBinding 是一个模式绑定的静态信息。
type ArmBinding struct {
	Name string
	Type types.Type
	Sym  *Symbol
}

// ArmRef 是一个 match 分支的静态信息。
type ArmRef struct {
	Wild     bool
	Builtin  string // "Some" / "None" / "Ok" / "Err"
	Tag      int64
	Enum     *types.EnumInfo
	Variant  types.VariantInfo
	HasVar   bool
	Bindings []ArmBinding
}

// MatchRef 是一次 match 的静态信息。
type MatchRef struct {
	Kind  string // "enum" / "optional" / "result" / "int"
	Head  types.Type
	Value types.Type // 取值位置的产出类型（语句位置为 Void）
}

// LitRef 是构造字面量的静态信息。
type LitRef struct {
	Struct  *types.StructInfo
	Enum    *types.EnumInfo
	Variant types.VariantInfo
	HasVar  bool
	Builtin string // "Some" / "Ok" / "Err"
	Payload types.Type
}

// AllocRef 是一次区域分配的静态信息。
type AllocRef struct {
	Elem     types.Type
	Count    int64
	HasCount bool
	Size     int64
	Align    int64
}

// TryRef 是 `?` 的静态信息。
type TryRef struct {
	Mode string // "result" / "optional"
	Ok   types.Type
}

// --------------------------------------------------------------------------- //

// FieldStep 是字段链上的一步（`a.b.c` 的每一段）。
type FieldStep struct {
	Info  *types.StructInfo
	Index int
	Type  types.Type
}

// FieldChain 是点号路径的字段访问链。
type FieldChain struct {
	Base   *Symbol
	Global *GlobalInfo
	Steps  []FieldStep
}

func (c *Checker) noteFieldChain(p *ast.Path, base *Symbol, global *GlobalInfo, steps []FieldStep) {
	c.FieldChains[p] = &FieldChain{Base: base, Global: global, Steps: steps}
}

// FieldChainOf 返回路径的字段链（没有则 nil）。
func (c *Checker) FieldChainOf(p *ast.Path) *FieldChain { return c.FieldChains[p] }

func (c *Checker) noteField(fa *ast.FieldAccess, base types.Type, info *types.StructInfo, idx int, t types.Type) {
	c.FieldRefs[fa] = &FieldRef{Base: base, Info: info, Index: idx, Type: t}
}

func (c *Checker) noteIndex(ix *ast.Index, base types.Type, kind string, elem types.Type) {
	c.IndexRefs[ix] = &IndexRef{Base: base, Kind: kind, Elem: elem}
}

func (c *Checker) noteFor(f *ast.ForIn, it, elem types.Type, vsym, isym *Symbol) {
	c.ForRefs[f] = &ForRef{Iter: it, Elem: elem, Var: vsym, Index: isym}
}

func (c *Checker) noteArm(arm *ast.Arm, ref *ArmRef) {
	cur := c.ArmRefs[arm]
	if cur == nil {
		c.ArmRefs[arm] = ref
		return
	}
	if ref.Wild {
		cur.Wild = true
	}
	if ref.Builtin != "" {
		cur.Builtin = ref.Builtin
	}
	if ref.HasVar {
		cur.Tag, cur.Enum, cur.Variant, cur.HasVar = ref.Tag, ref.Enum, ref.Variant, true
	}
}

func (c *Checker) noteArmBinding(arm *ast.Arm, name string, t types.Type, sym *Symbol) {
	ref := c.ArmRefs[arm]
	if ref == nil {
		ref = &ArmRef{}
		c.ArmRefs[arm] = ref
	}
	ref.Bindings = append(ref.Bindings, ArmBinding{Name: name, Type: t, Sym: sym})
}

func (c *Checker) noteMatch(m *ast.Match, kind string, head, value types.Type) {
	c.MatchRefs[m] = &MatchRef{Kind: kind, Head: head, Value: value}
}

func (c *Checker) noteLit(x *ast.StructLit, ref *LitRef) { c.LitRefs[x] = ref }

func (c *Checker) noteArrayElem(x *ast.ArrayLit, elem types.Type, count int64, hasCount bool) {
	c.ArrRefs[x] = &ArrayRef{Elem: elem, Count: count, HasCount: hasCount}
}

func (c *Checker) noteAlloc(x *ast.Alloc, ref *AllocRef) { c.AllocRefs[x] = ref }

func (c *Checker) noteTry(x *ast.TryPropagate, mode string, ok types.Type) {
	c.TryRefs[x] = &TryRef{Mode: mode, Ok: ok}
}

func (c *Checker) noteLen(x *ast.Len, t types.Type) { c.LenRefs[x] = t }

func (c *Checker) noteDestruct(d *ast.DestructAssign, syms []*Symbol) { c.DestructRefs[d] = syms }

// ArrayRef 是数组字面量的静态信息。
type ArrayRef struct {
	Elem     types.Type
	Count    int64
	HasCount bool
}

// --------------------------------------------------------------------------- //
// 变体信息访问（codegen 需要 enum / 变体的标签与载荷）
// --------------------------------------------------------------------------- //

// PathVariant 返回路径指向的 enum 变体。
func (c *Checker) PathVariant(p *ast.Path) (*types.EnumInfo, types.VariantInfo, bool) {
	if vr, ok := c.PathVariants[p]; ok {
		return vr.enum, vr.variant, true
	}
	return nil, types.VariantInfo{}, false
}

// LitVariant 返回构造字面量指向的 enum 变体。
func (c *Checker) LitVariant(x *ast.StructLit) (*types.EnumInfo, types.VariantInfo, bool) {
	if vr, ok := c.VariantLits[x]; ok {
		return vr.enum, vr.variant, true
	}
	return nil, types.VariantInfo{}, false
}

// ArmVariant 返回 match 分支指向的 enum 变体。
func (c *Checker) ArmVariant(a *ast.Arm) (*types.EnumInfo, types.VariantInfo, bool) {
	if vr, ok := c.ArmVariants[a]; ok {
		return vr.enum, vr.variant, true
	}
	return nil, types.VariantInfo{}, false
}