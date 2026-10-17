package sema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/aila-lang/aila/internal/parser"
	"github.com/aila-lang/aila/internal/types"
)

// sigSnapshot 是 tools/gen_sigs.py 生成的期望值。
type sigSnapshot struct {
	Funcs   map[string]sigEntry `json:"funcs"`
	Methods map[string]sigEntry `json:"methods"`
	Consts  map[string]struct {
		Type  string `json:"type"`
		Value any    `json:"value"`
	} `json:"consts"`
	Globals map[string]struct {
		Type    string `json:"type"`
		Kind    string `json:"kind"`
		Mutable bool   `json:"mutable"`
	} `json:"globals"`
}

type sigEntry struct {
	Owner     string     `json:"owner"`
	Params    [][]string `json:"params"` // [[名字, 类型], …]
	Ret       string     `json:"ret"`
	Err       *string    `json:"err"`
	Group     []string   `json:"group"`
	Regions   []string   `json:"regions"`
	Muts      []string   `json:"muts"`
	Moves     []string   `json:"moves"`
	Coroutine bool       `json:"coroutine"`
	Thread    bool       `json:"thread"`
	Extern    bool       `json:"extern"`
	Public    bool       `json:"public"`
	Variadic  bool       `json:"variadic"`
}

// TestSigsMatchReference 是 G3 的第二条判据：
// 同一份 sigs.aila，Go 抽出的函数签名（形参/返回/错误通道/多值组/效果子句/染色）
// 必须与 Python 参考实现一致。
func TestSigsMatchReference(t *testing.T) {
	root := repoRoot(t)
	srcPath := filepath.Join(root, "tests", "corpus", "sigs.aila")
	wantPath := filepath.Join(root, "tests", "corpus", "sigs.json")
	src, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("读不到语料：%v", err)
	}
	wantRaw, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("读不到签名快照（先跑 python tools/gen_sigs.py）：%v", err)
	}
	var want sigSnapshot
	if err := json.Unmarshal(wantRaw, &want); err != nil {
		t.Fatalf("快照不是合法 JSON：%v", err)
	}

	mod, bag := parser.Parse("sigs.aila", string(src))
	if bag.HasErrors() {
		t.Fatalf("解析失败：%s", bag.Render(false))
	}
	table := BuildTable(mod, bag)
	sigs := BuildSigs(mod, table, bag)
	if bag.HasErrors() {
		t.Fatalf("签名构建失败：%s", bag.Render(false))
	}

	compare := func(kind, name string, got *FuncSig, exp sigEntry) {
		if len(got.Params) != len(exp.Params) {
			t.Fatalf("%s %s 形参数：got=%d want=%d", kind, name, len(got.Params), len(exp.Params))
		}
		for i, p := range exp.Params {
			if got.Params[i].Name != p[0] {
				t.Fatalf("%s %s 形参 %d 名字：got=%s want=%s", kind, name, i, got.Params[i].Name, p[0])
			}
			if g := types.TypeName(got.Params[i].Type); g != p[1] {
				t.Fatalf("%s %s 形参 %s 类型：got=%s want=%s", kind, name, p[0], g, p[1])
			}
		}
		if g := types.TypeName(got.Ret); g != exp.Ret {
			t.Fatalf("%s %s 返回类型：got=%s want=%s", kind, name, g, exp.Ret)
		}
		gotErr := ""
		if got.Err != nil {
			gotErr = types.TypeName(got.Err)
		}
		wantErr := ""
		if exp.Err != nil {
			wantErr = *exp.Err
		}
		if gotErr != wantErr {
			t.Fatalf("%s %s 错误通道：got=%q want=%q", kind, name, gotErr, wantErr)
		}
		var gotGroup []string
		for _, g := range got.RetGroup {
			gotGroup = append(gotGroup, types.TypeName(g))
		}
		if strings.Join(gotGroup, ",") != strings.Join(exp.Group, ",") {
			t.Fatalf("%s %s 多值组：got=%v want=%v", kind, name, gotGroup, exp.Group)
		}
		if strings.Join(got.Regions, ",") != strings.Join(exp.Regions, ",") {
			t.Fatalf("%s %s @region：got=%v want=%v", kind, name, got.Regions, exp.Regions)
		}
		if strings.Join(got.Muts, ",") != strings.Join(exp.Muts, ",") {
			t.Fatalf("%s %s @mut：got=%v want=%v", kind, name, got.Muts, exp.Muts)
		}
		if strings.Join(got.Moves, ",") != strings.Join(exp.Moves, ",") {
			t.Fatalf("%s %s @move：got=%v want=%v", kind, name, got.Moves, exp.Moves)
		}
		if got.IsCoroutine != exp.Coroutine || got.IsThread != exp.Thread ||
			got.IsExtern != exp.Extern || got.IsPublic != exp.Public || got.Variadic != exp.Variadic {
			t.Fatalf("%s %s 标志位不一致：got=(coro=%v thread=%v extern=%v public=%v variadic=%v)",
				kind, name, got.IsCoroutine, got.IsThread, got.IsExtern, got.IsPublic, got.Variadic)
		}
	}

	if len(sigs.Funcs) != len(want.Funcs) {
		got := make([]string, 0, len(sigs.Funcs))
		for n := range sigs.Funcs {
			got = append(got, n)
		}
		sort.Strings(got)
		t.Fatalf("函数数量不符：got=%v want=%d 个", got, len(want.Funcs))
	}
	for name, exp := range want.Funcs {
		got, ok := sigs.Funcs[name]
		if !ok {
			t.Fatalf("缺少函数 %s", name)
		}
		compare("函数", name, got, exp)
	}
	for name, exp := range want.Methods {
		got, ok := sigs.Methods[name]
		if !ok {
			t.Fatalf("缺少方法 %s", name)
		}
		compare("方法", name, got, exp)
	}

	// 常量
	if len(sigs.Consts) != len(want.Consts) {
		t.Fatalf("常量数量不符：got=%d want=%d", len(sigs.Consts), len(want.Consts))
	}
	for name, exp := range want.Consts {
		got, ok := sigs.Consts[name]
		if !ok {
			t.Fatalf("缺少常量 %s", name)
		}
		if g := types.TypeName(got.Type); g != exp.Type {
			t.Fatalf("常量 %s 类型：got=%s want=%s", name, g, exp.Type)
		}
		if !sameScalar(got.Value, exp.Value) {
			t.Fatalf("常量 %s 值：got=%v want=%v", name, got.Value, exp.Value)
		}
	}

	// 模块级绑定（state）
	for name, exp := range want.Globals {
		got, ok := sigs.Globals[name]
		if !ok {
			t.Fatalf("缺少模块级绑定 %s", name)
		}
		if g := types.TypeName(got.Type); g != exp.Type {
			t.Fatalf("绑定 %s 类型：got=%s want=%s", name, g, exp.Type)
		}
		if got.Kind != exp.Kind || got.Mutable != exp.Mutable {
			t.Fatalf("绑定 %s：got=(%s,%v) want=(%s,%v)", name, got.Kind, got.Mutable, exp.Kind, exp.Mutable)
		}
	}
}

func sameScalar(got, want any) bool {
	toF := func(v any) (float64, bool) {
		switch x := v.(type) {
		case int64:
			return float64(x), true
		case int:
			return float64(x), true
		case float64:
			return x, true
		}
		return 0, false
	}
	if g, ok1 := toF(got); ok1 {
		w, ok2 := toF(want)
		return ok2 && g == w
	}
	switch g := got.(type) {
	case string:
		w, ok := want.(string)
		return ok && g == w
	case bool:
		w, ok := want.(bool)
		return ok && g == w
	}
	return got == want
}
