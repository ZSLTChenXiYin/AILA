package sema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"aila/internal/diag"
	"aila/internal/parser"
	"aila/internal/types"
)

// layoutSnapshot 是 tools/gen_layout.py 生成的期望值。
type layoutSnapshot struct {
	Types   map[string]typeEntry `json:"types"`
	Aliases map[string]typeEntry `json:"aliases"`
}

type typeEntry struct {
	Kind      string       `json:"kind"`
	Type      string       `json:"type"`
	Size      int          `json:"size"`
	Align     int          `json:"align"`
	Packed    bool         `json:"packed"`
	AlignAttr int          `json:"align_attr"`
	Payload   bool         `json:"payload"`
	Tag       string       `json:"tag"`
	TagSize   int          `json:"tag_size"`
	Fields    []fieldEntry `json:"fields"`
	Variants  []varEntry   `json:"variants"`
}

type fieldEntry struct {
	Name   string `json:"name"`
	Offset int    `json:"offset"`
	Size   int    `json:"size"`
	Align  int    `json:"align"`
}

type varEntry struct {
	Name string `json:"name"`
	Tag  int64  `json:"tag"`
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller 失败")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// TestLayoutMatchesReference 是 G3 的第一条判据：
// 同一份 layout.aila，Go 算出的 size / align / 字段偏移 / enum tag 宽度必须与 Python 参考实现一致。
func TestLayoutMatchesReference(t *testing.T) {
	root := repoRoot(t)
	srcPath := filepath.Join(root, "tests", "corpus", "layout.aila")
	wantPath := filepath.Join(root, "tests", "corpus", "layout.json")

	src, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("读不到语料：%v", err)
	}
	wantRaw, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("读不到布局快照（先跑 python tools/gen_layout.py）：%v", err)
	}
	var want layoutSnapshot
	if err := json.Unmarshal(wantRaw, &want); err != nil {
		t.Fatalf("快照不是合法 JSON：%v", err)
	}

	// 错误空间编号是进程级全局状态：本测试不经过 build.Load，必须自己重置
	// 基线，否则 enum tag 会随本包其它测试的注册顺序漂移（-shuffle 下复现）。
	types.ResetErrorSpace()

	mod, bag := parser.Parse("layout.aila", string(src))
	if bag.HasErrors() {
		t.Fatalf("解析失败：%s", bag.Render(false))
	}
	table := BuildTable(mod, bag)
	if bag.HasErrors() {
		t.Fatalf("类型表构建失败：%s", bag.Render(false))
	}
	layout := types.NewLayout(8)

	// 结构体 / 枚举
	for name, exp := range want.Types {
		info, ok := table.Types[name]
		if !ok {
			t.Fatalf("缺少类型 %s", name)
		}
		named := types.Named{Info: info}
		if got := types.TypeName(named); got != exp.Type {
			t.Fatalf("%s 类型拼写：got=%s want=%s", name, got, exp.Type)
		}
		if got := layout.Size(named); got != exp.Size {
			t.Fatalf("%s size：got=%d want=%d", name, got, exp.Size)
		}
		if got := layout.Align(named); got != exp.Align {
			t.Fatalf("%s align：got=%d want=%d", name, got, exp.Align)
		}
		switch x := info.(type) {
		case *types.StructInfo:
			if exp.Kind != "struct" {
				t.Fatalf("%s 种类：got=struct want=%s", name, exp.Kind)
			}
			if x.Packed != exp.Packed {
				t.Fatalf("%s packed：got=%v want=%v", name, x.Packed, exp.Packed)
			}
			offsets := layout.FieldOffsets(x)
			if len(offsets) != len(exp.Fields) {
				t.Fatalf("%s 字段数：got=%d want=%d", name, len(offsets), len(exp.Fields))
			}
			for i, f := range exp.Fields {
				if x.Fields[i].Name != f.Name {
					t.Fatalf("%s 字段 %d 名字：got=%s want=%s", name, i, x.Fields[i].Name, f.Name)
				}
				if offsets[i] != f.Offset {
					t.Fatalf("%s.%s offset：got=%d want=%d", name, f.Name, offsets[i], f.Offset)
				}
				if got := layout.Size(x.Fields[i].Type); got != f.Size {
					t.Fatalf("%s.%s size：got=%d want=%d", name, f.Name, got, f.Size)
				}
				if got := layout.Align(x.Fields[i].Type); got != f.Align {
					t.Fatalf("%s.%s align：got=%d want=%d", name, f.Name, got, f.Align)
				}
			}
		case *types.EnumInfo:
			if exp.Kind != "enum" {
				t.Fatalf("%s 种类：got=enum want=%s", name, exp.Kind)
			}
			if x.Payload != exp.Payload {
				t.Fatalf("%s payload：got=%v want=%v", name, x.Payload, exp.Payload)
			}
			if got := types.TypeName(x.TagType); got != exp.Tag {
				t.Fatalf("%s tag 类型：got=%s want=%s", name, got, exp.Tag)
			}
			if got := layout.Size(x.TagType); got != exp.TagSize {
				t.Fatalf("%s tag 宽度：got=%d want=%d", name, got, exp.TagSize)
			}
			if len(x.Variants) != len(exp.Variants) {
				t.Fatalf("%s 变体数：got=%d want=%d", name, len(x.Variants), len(exp.Variants))
			}
			for i, v := range exp.Variants {
				if x.Variants[i].Name != v.Name || x.Variants[i].Tag != v.Tag {
					t.Fatalf("%s 变体 %d：got=(%s,%d) want=(%s,%d)", name, i,
						x.Variants[i].Name, x.Variants[i].Tag, v.Name, v.Tag)
				}
			}
		}
	}

	// 别名
	for name, exp := range want.Aliases {
		got, ok := table.Lookup(name)
		if !ok {
			t.Fatalf("缺少别名 %s", name)
		}
		if s := types.TypeName(got); s != exp.Type {
			t.Fatalf("%s 类型拼写：got=%s want=%s", name, s, exp.Type)
		}
		if s := layout.Size(got); s != exp.Size {
			t.Fatalf("%s size：got=%d want=%d", name, s, exp.Size)
		}
		if a := layout.Align(got); a != exp.Align {
			t.Fatalf("%s align：got=%d want=%d", name, a, exp.Align)
		}
	}
}

func TestUndefinedTypeIsReported(t *testing.T) {
	src := "struct S { a: Nope };\nfunc main() { }\n"
	mod, bag := parser.Parse("t.aila", src)
	if bag.HasErrors() {
		t.Fatalf("解析失败：%s", bag.Render(false))
	}
	BuildTable(mod, bag)
	if !bag.HasErrors() {
		t.Fatal("未知类型应当报错")
	}
	found := false
	for _, d := range bag.Items {
		if d.Code == "E3001" {
			found = true
		}
	}
	if !found {
		t.Fatalf("应报 E3001：%v", bag.Items)
	}
}

func TestDuplicateTypeIsReported(t *testing.T) {
	src := "struct S { a: I8 };\nstruct S { b: I8 };\nfunc main() { }\n"
	mod, bag := parser.Parse("t.aila", src)
	BuildTable(mod, bag)
	if !bag.HasErrors() {
		t.Fatal("重复类型应当报错")
	}
}

func TestScalarOptionalRejected(t *testing.T) {
	src := "type T = I32?;\nfunc main() { }\n"
	mod, bag := parser.Parse("t.aila", src)
	BuildTable(mod, bag)
	codes := map[string]bool{}
	for _, d := range bag.Items {
		codes[d.Code] = true
	}
	if !codes["E3601"] {
		t.Fatalf("标量可空应报 E3601：%v", bag.Items)
	}
}

func TestConstIntEval(t *testing.T) {
	src := "enum E: U8 { A = 2 + 3, B, C = 0x10 };\nfunc main() { }\n"
	mod, bag := parser.Parse("t.aila", src)
	table := BuildTable(mod, bag)
	if bag.HasErrors() {
		t.Fatalf("不应报错：%s", bag.Render(false))
	}
	info, _ := table.Types["E"].(*types.EnumInfo)
	want := []int64{5, 6, 16}
	for i, w := range want {
		if info.Variants[i].Tag != w {
			t.Fatalf("变体 %s tag：got=%d want=%d", info.Variants[i].Name, info.Variants[i].Tag, w)
		}
	}
}

var _ = diag.Empty
