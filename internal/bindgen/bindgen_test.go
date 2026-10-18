package bindgen

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"aila/internal/diag"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// TestMathkitBindings 是 bindgen 的第一条判据：`examples/cinterop` 的 mathkit cpack
// 用 builtin 档生成的 `.ail`，必须与参考实现（Python）逐字节一致。
func TestMathkitBindings(t *testing.T) {
	root := repoRoot(t)
	cpackDir := filepath.Join(root, "examples", "cinterop", "cpacks", "mathkit")
	target := Target{Board: "host", Triple: "x86_64-w64-windows-gnu", PtrSize: 8}
	bag := &diag.Bag{}
	res := Run("mathkit", cpackDir, target, bag, "builtin")
	if bag.HasErrors() {
		t.Fatalf("bindgen 报错：\n%s", bag.Render(false))
	}
	if res.Backend != "builtin" {
		t.Fatalf("期望 builtin 档，实得 %s", res.Backend)
	}
	full := res.Header("mathkit", target) + "\n\n" + res.Text
	snap := filepath.Join(root, "tests", "corpus", "bindings", "mathkit.ail")
	want, err := os.ReadFile(snap)
	if err != nil {
		t.Fatalf("读不到快照（先跑 go run ./tools/goldengen）：%v", err)
	}
	if strings.TrimRight(full, " \t\r\n") != strings.TrimRight(string(want), " \t\r\n") {
		t.Fatalf("bindgen 输出与快照不一致：\n--- 实得 ---\n%s\n--- 期望 ---\n%s", full, string(want))
	}
}

// TestMapCType 守护 C → AILA 的类型映射表（docs/specs/0001 §6.2）。
func TestMapCType(t *testing.T) {
	target := Target{Board: "host", Triple: "x86_64-w64-windows-gnu", PtrSize: 8}
	named := map[string]string{"mathkit_err_t": "MathkitErr"}
	cases := map[string]string{
		"int":                    "I32",
		"unsigned int":           "U32",
		"unsigned long":          "U64",
		"long":                   "I64",
		"float":                  "F32",
		"double":                 "F64",
		"_Bool":                  "Bool",
		"char *":                 "[U8]?",
		"const char *":           "[U8]?",
		"void *":                 "[Unsafe]?",
		"int *":                  "[I32]?",
		"void":                   "（void）",
		"size_t":                 "USize",
		"mathkit_err_t":          "MathkitErr",
		"struct Frame":           "Frame",
		"int **":                 "[Unsafe]?",
	}
	for in, want := range cases {
		bag := &diag.Bag{}
		got := MapCType(in, target, named, bag, diag.Empty, "test")
		if got != want {
			t.Errorf("MapCType(%q) = %q，期望 %q", in, got, want)
		}
	}
}

// TestTypeAndValueNames 守护命名规范（docs/specs/0001 §6.3）。
func TestTypeAndValueNames(t *testing.T) {
	typeCases := map[string]string{
		"frame_t":   "Frame",
		"frame":     "Frame",
		"my_frame":  "MyFrame",
		"_frame":    "Frame",
		"struct":    "Struct",
		"":          "Anon",
		"io_buffer": "IoBuffer",
	}
	for in, want := range typeCases {
		if got := TypeName(in); got != want {
			t.Errorf("TypeName(%q) = %q，期望 %q", in, got, want)
		}
	}
	if got := ValueName("func"); got != "func_c" {
		t.Errorf("ValueName(func) = %q", got)
	}
	if got := ValueName("mathkit_add"); got != "mathkit_add" {
		t.Errorf("ValueName(mathkit_add) = %q", got)
	}
}

// TestTokenizeAndPreprocess 覆盖 include guard 与宏收集。
func TestTokenizeAndPreprocess(t *testing.T) {
	src := "#ifndef X_H\n#define X_H\n#define MAX 10\nint f(int a);\n#endif\n"
	macros := map[string]string{}
	order := []string{}
	notes := []string{}
	out := StripPreprocessor(src, macros, &order, &notes)
	if !strings.Contains(out, "int f(int a);") {
		t.Fatalf("include guard 处理错误，保留内容：%q", out)
	}
	if macros["MAX"] != "10" {
		t.Fatalf("宏收集失败：%v", macros)
	}
	toks := TokenizeC("int f(int a);")
	if len(toks) == 0 || toks[0].Text != "int" {
		t.Fatalf("tokenize 失败：%v", toks)
	}
}
