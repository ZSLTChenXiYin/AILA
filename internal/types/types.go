// Package types 是 AILA 的类型项与 C ABI 布局（docs/specs/0002 §5）。
//
// 与 Python 参考实现同一套口径：判据是 `tests/corpus/layout.json` 快照
// （各类型的 size / align / 字段偏移，由 tools/gen_layout.py 从参考实现生成）。
package types

import "fmt"

// Type 是类型项。
type Type interface{ typeName() string }

// --------------------------------------------------------------------------- //
// 基元与复合
// --------------------------------------------------------------------------- //

// Prim 是基元类型（kind：i8…u64/f32/f64/bool/c32/text/usize/isize）。
type Prim struct{ K string }

// 拼写与 Python 的 describe() 一致（诊断里要显示 `I32` / `Text` / `USize`）。
var primSpelling = map[string]string{
	"i8": "I8", "i16": "I16", "i32": "I32", "i64": "I64",
	"u8": "U8", "u16": "U16", "u32": "U32", "u64": "U64",
	"f32": "F32", "f64": "F64", "bool": "Bool", "c32": "C32",
	"text": "Text", "usize": "USize", "isize": "ISize",
}

func (t Prim) typeName() string {
	if s, ok := primSpelling[t.K]; ok {
		return s
	}
	return t.K
}

// 常用单例。
var (
	I8    = Prim{"i8"}
	I16   = Prim{"i16"}
	I32   = Prim{"i32"}
	I64   = Prim{"i64"}
	U8    = Prim{"u8"}
	U16   = Prim{"u16"}
	U32   = Prim{"u32"}
	U64   = Prim{"u64"}
	F32   = Prim{"f32"}
	F64   = Prim{"f64"}
	Bool  = Prim{"bool"}
	C32   = Prim{"c32"}
	Text  = Prim{"text"}
	USize = Prim{"usize"}
	ISize = Prim{"isize"}
)

// PrimByKind 把 kind 映射成单例。
func PrimByKind(k string) (Type, bool) {
	switch k {
	case "i8":
		return I8, true
	case "i16":
		return I16, true
	case "i32":
		return I32, true
	case "i64":
		return I64, true
	case "u8":
		return U8, true
	case "u16":
		return U16, true
	case "u32":
		return U32, true
	case "u64":
		return U64, true
	case "f32":
		return F32, true
	case "f64":
		return F64, true
	case "bool":
		return Bool, true
	case "c32":
		return C32, true
	case "text":
		return Text, true
	case "usize":
		return USize, true
	case "isize":
		return ISize, true
	}
	return nil, false
}

// View 是 `[T]` 胖视图。
type View struct{ Elem Type }

func (t View) typeName() string { return "[" + TypeName(t.Elem) + "]" }

// Fixed 是 `[T; N]` 内联数组。
type Fixed struct {
	Elem Type
	N    int64
}

func (t Fixed) typeName() string { return fmt.Sprintf("[%s; %d]", TypeName(t.Elem), t.N) }

// UnsafeView 是 `[Unsafe]`。
type UnsafeView struct{}

func (UnsafeView) typeName() string { return "[Unsafe]" }

// Optional 是 `T?`。
type Optional struct{ Inner Type }

func (t Optional) typeName() string { return TypeName(t.Inner) + "?" }

// Result 是 `T ! E`。
type Result struct{ Ok, Err Type }

func (t Result) typeName() string { return TypeName(t.Ok) + " ! " + TypeName(t.Err) }

// Chan 是 `chan T` / `send chan T` / `recv chan T`。
type Chan struct {
	Elem Type
	Dir  string
}

func (t Chan) typeName() string {
	if t.Dir == "" {
		return "chan " + TypeName(t.Elem)
	}
	return t.Dir + " chan " + TypeName(t.Elem)
}

// Coro 是 `coro [T ! E]`；Ok/Err 可为 nil。
type Coro struct{ Ok, Err Type }

func (t Coro) typeName() string {
	if t.Ok == nil {
		return "coro"
	}
	if t.Err != nil {
		return "coro " + TypeName(t.Ok) + " ! " + TypeName(t.Err)
	}
	return "coro " + TypeName(t.Ok)
}

// FnPtr 是裸函数指针。
type FnPtr struct {
	Params []Type
	Ret    Type
	Err    Type
}

func (t FnPtr) typeName() string {
	s := "("
	for i, p := range t.Params {
		if i > 0 {
			s += ", "
		}
		s += TypeName(p)
	}
	s += ")"
	if t.Ret != nil {
		s += " -> " + TypeName(t.Ret)
	}
	if t.Err != nil {
		s += " ! " + TypeName(t.Err)
	}
	return s
}

// Group 是多值组。
type Group struct{ Items []Type }

func (t Group) typeName() string {
	s := "("
	for i, x := range t.Items {
		if i > 0 {
			s += ", "
		}
		s += TypeName(x)
	}
	return s + ")"
}

// Never 是发散（trap / 永不返回）。
type Never struct{}

func (Never) typeName() string { return "!" }

// Void 是无返回值。
type Void struct{}

func (Void) typeName() string { return "（无返回值）" }

// UntypedInt 是多态整数字面量。
type UntypedInt struct{ Value int64 }

func (t UntypedInt) typeName() string { return "整数常量" }

// UntypedFloat 是多态浮点字面量。
type UntypedFloat struct{ Value float64 }

func (t UntypedFloat) typeName() string { return "浮点常量" }

// --------------------------------------------------------------------------- //
// 用户定义类型
// --------------------------------------------------------------------------- //

// Info 是 Named 背后的类型定义。
type Info interface {
	TypeInfoName() string
	moduleName() string
}

// Field 是一个字段。
type Field struct {
	Name string
	Type Type
}

// StructInfo 是 struct 定义。
type StructInfo struct {
	Name   string
	Module string
	Fields []Field
	Packed bool
	Align  int
	Opaque bool
}

// TypeInfoName 实现 Info。
func (s *StructInfo) TypeInfoName() string { return s.Name }

func (s *StructInfo) moduleName() string { return s.Module }

// VariantInfo 是 enum 变体。
type VariantInfo struct {
	Name    string
	Tag     int64
	Payload []Field // nil = 无载荷
}

// EnumInfo 是 enum 定义。
type EnumInfo struct {
	Name    string
	Module  string
	Base    Type
	Payload bool
	Variants []VariantInfo
	TagType Type
}

// TypeInfoName 实现 Info。
func (e *EnumInfo) TypeInfoName() string { return e.Name }

func (e *EnumInfo) moduleName() string { return e.Module }

// Variant 按名字查变体。
func (e *EnumInfo) Variant(name string) (VariantInfo, bool) {
	for _, v := range e.Variants {
		if v.Name == name {
			return v, true
		}
	}
	return VariantInfo{}, false
}

// Named 是用户定义类型引用（同一性 = 同一个 Info 指针）。
type Named struct{ Info Info }

func (t Named) typeName() string { return t.Info.TypeInfoName() }

// TypeName 返回类型的 AILA 拼写。
func TypeName(t Type) string {
	if t == nil {
		return "?"
	}
	return t.typeName()
}

// --------------------------------------------------------------------------- //
// 判定谓词（specs/0002 §5.3）
// --------------------------------------------------------------------------- //

// IsViewFamily 判断是否是视图家族（含 Text）。
func IsViewFamily(t Type) bool {
	switch x := t.(type) {
	case View, UnsafeView:
		return true
	case Prim:
		return x.K == "text"
	}
	return false
}

// IsHandle 判断 `T?` 是否合法目标（designs/0001 §3.7 规则 2）。
func IsHandle(t Type) bool {
	if IsViewFamily(t) {
		return true
	}
	switch x := t.(type) {
	case FnPtr, Chan, Coro, Result:
		return true
	case Named:
		switch info := x.Info.(type) {
		case *StructInfo:
			return IsMoveType(t)
		case *EnumInfo:
			_ = info
			return true
		}
	}
	return false
}

// IsMoveType 是 §3.6 两族判定：含句柄字段者属 move 族。
func IsMoveType(t Type) bool {
	if IsViewFamily(t) {
		return true
	}
	switch x := t.(type) {
	case Chan, Coro, Result, Group:
		return true
	case Named:
		switch info := x.Info.(type) {
		case *StructInfo:
			for _, f := range info.Fields {
				if IsMoveType(f.Type) {
					return true
				}
			}
		case *EnumInfo:
			for _, v := range info.Variants {
				for _, f := range v.Payload {
					if IsMoveType(f.Type) {
						return true
					}
				}
			}
		}
	}
	return false
}

// HasFnPtr 判断类型里是否含函数指针字段。
func HasFnPtr(t Type) bool {
	switch x := t.(type) {
	case FnPtr:
		return true
	case Named:
		switch info := x.Info.(type) {
		case *StructInfo:
			for _, f := range info.Fields {
				if HasFnPtr(f.Type) {
					return true
				}
			}
		case *EnumInfo:
			for _, v := range info.Variants {
				for _, f := range v.Payload {
					if HasFnPtr(f.Type) {
						return true
					}
				}
			}
		}
	}
	return false
}

// IsInteger 判断是否整数标量（含 usize/isize；位宽随目标）。
func IsInteger(t Type) bool {
	x, ok := t.(Prim)
	if !ok {
		return false
	}
	switch x.K {
	case "i8", "i16", "i32", "i64", "u8", "u16", "u32", "u64", "usize", "isize":
		return true
	}
	return false
}

// IsFloat 判断是否浮点标量。
func IsFloat(t Type) bool {
	x, ok := t.(Prim)
	return ok && (x.K == "f32" || x.K == "f64")
}

// IsSigned 判断是否有符号整数。
func IsSigned(t Type) bool {
	x, ok := t.(Prim)
	if !ok {
		return false
	}
	switch x.K {
	case "i8", "i16", "i32", "i64", "isize":
		return true
	}
	return false
}

// IntWidth 返回位宽（位）；非标量返回 0。
func IntWidth(t Type, ptrSize int) int {
	x, ok := t.(Prim)
	if !ok {
		return 0
	}
	if x.K == "usize" || x.K == "isize" {
		return ptrSize * 8
	}
	switch x.K {
	case "i8", "u8", "bool":
		return 8
	case "i16", "u16":
		return 16
	case "i32", "u32", "c32", "f32":
		return 32
	case "i64", "u64", "f64":
		return 64
	}
	return 0
}
