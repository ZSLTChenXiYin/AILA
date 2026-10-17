package types

import "testing"

func TestPredicates(t *testing.T) {
	view := View{Elem: U8}
	unsafeView := UnsafeView{}
	enum := Named{Info: &EnumInfo{Name: "E", Base: U32, TagType: U32}}
	moveStruct := Named{Info: &StructInfo{Name: "M", Fields: []Field{{Name: "v", Type: view}}}}
	plainStruct := Named{Info: &StructInfo{Name: "S", Fields: []Field{{Name: "a", Type: I32}}}}

	if !IsViewFamily(view) || !IsViewFamily(Text) || !IsViewFamily(unsafeView) {
		t.Fatal("视图家族判定有误")
	}
	if IsViewFamily(I32) || IsViewFamily(plainStruct) {
		t.Fatal("标量 / 纯数据 struct 不属于视图家族")
	}
	if !IsHandle(view) || !IsHandle(Text) || !IsHandle(enum) || !IsHandle(FnPtr{}) || !IsHandle(Chan{}) {
		t.Fatal("句柄类清单判定有误")
	}
	if IsHandle(I32) {
		t.Fatal("标量不可为空")
	}
	if !IsMoveType(view) || !IsMoveType(moveStruct) || !IsMoveType(Result{Ok: I32, Err: enum}) {
		t.Fatal("move 族判定有误")
	}
	if IsMoveType(plainStruct) || IsMoveType(I32) {
		t.Fatal("纯数据不该是 move 族")
	}
	if !HasFnPtr(FnPtr{}) || !HasFnPtr(Named{Info: &StructInfo{Name: "F", Fields: []Field{{Name: "f", Type: FnPtr{}}}}}) {
		t.Fatal("函数指针判定有误")
	}
}

func TestIntWidth(t *testing.T) {
	cases := []struct {
		t    Type
		bits int
	}{{I8, 8}, {U16, 16}, {I32, 32}, {C32, 32}, {F64, 64}, {Bool, 8}, {USize, 64}, {ISize, 64}, {Text, 0}, {View{Elem: U8}, 0}}
	for _, c := range cases {
		if got := IntWidth(c.t, 8); got != c.bits {
			t.Fatalf("%s 位宽：got=%d want=%d", TypeName(c.t), got, c.bits)
		}
	}
	if got := IntWidth(USize, 4); got != 32 {
		t.Fatalf("32 位目标的 USize 位宽：got=%d", got)
	}
}

func TestScalarSizes(t *testing.T) {
	l := NewLayout(8)
	cases := []struct {
		t Type
		s int
		a int
	}{
		{I8, 1, 1}, {U16, 2, 2}, {I32, 4, 4}, {F64, 8, 8}, {Bool, 1, 1},
		{Text, 16, 8}, {View{Elem: U8}, 16, 8}, {UnsafeView{}, 16, 8},
		{FnPtr{}, 8, 8}, {Chan{}, 8, 8}, {Fixed{Elem: F32, N: 4}, 16, 4},
	}
	for _, c := range cases {
		if s, a := l.SizeAlign(c.t); s != c.s || a != c.a {
			t.Fatalf("%s：got=(%d,%d) want=(%d,%d)", TypeName(c.t), s, a, c.s, c.a)
		}
	}
	if l32 := NewLayout(4); l32.Size(Text) != 8 {
		t.Fatalf("32 位目标的 Text 应为 8 字节，实得 %d", l32.Size(Text))
	}
}

func TestTypeNames(t *testing.T) {
	cases := []struct {
		t    Type
		want string
	}{
		{I32, "I32"}, {Text, "Text"}, {USize, "USize"},
		{View{Elem: U8}, "[U8]"},
		{Fixed{Elem: F32, N: 4}, "[F32; 4]"},
		{UnsafeView{}, "[Unsafe]"},
		{Optional{Inner: View{Elem: U8}}, "[U8]?"},
		{Chan{Elem: I32, Dir: "send"}, "send chan I32"},
		{Coro{}, "coro"},
		{Group{Items: []Type{U8, U16}}, "(U8, U16)"},
		{Void{}, "（无返回值）"},
		{Never{}, "!"},
	}
	for _, c := range cases {
		if got := TypeName(c.t); got != c.want {
			t.Fatalf("拼写：got=%s want=%s", got, c.want)
		}
	}
}
