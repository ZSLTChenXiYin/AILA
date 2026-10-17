package sema

import (
	"strings"
	"testing"

	"github.com/ailang/aila/internal/ast"
	"github.com/ailang/aila/internal/diag"
	"github.com/ailang/aila/internal/parser"
)

// runModules 解析并语义检查一组模块（模块名 → 源码），返回错误码集合。
func runModules(t *testing.T, mods map[string]string) (map[string]bool, *Program) {
	t.Helper()
	bag := &diag.Bag{}
	parsed := map[string]*ast.Module{}
	order := []string{}
	names := make([]string, 0, len(mods))
	for name := range mods {
		names = append(names, name)
	}
	// 稳定的处理顺序
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[j] < names[i] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	for _, name := range names {
		mod, mbag := parser.Parse(name+".aila", mods[name])
		bag.Items = append(bag.Items, mbag.Items...)
		mod.Name = name
		parsed[name] = mod
		order = append(order, name)
	}
	prog := BuildProgram(parsed, order, "_.main", bag)
	codes := map[string]bool{}
	for _, d := range bag.Items {
		codes[d.Code] = true
	}
	return codes, prog
}

// TestVisibilityLadder 守护可见性阶梯（designs/0001 §2.2）：
// 块内默认文件内、块内 `public` → 子包内、壳内 `public` → 包外。
func TestVisibilityLadder(t *testing.T) {
	cases := []struct {
		name    string
		decl    string // 声明所在的 a.aila
		use     string // 使用处的 main.aila（同为 `_` 子包或跨子包）
		wantErr bool
	}{
		{
			name:    "块内默认：跨文件不可见",
			decl:    "module {\n    func hidden() -> I32 { return 1; }\n}\n",
			use:     "import _.sub.a as a;\n\nmodule {\n    func probe() -> I32 { return a.hidden(); }\n}\n",
			wantErr: true,
		},
		{
			name:    "块内 public：同子包可见",
			decl:    "module {\n    public func shared() -> I32 { return 1; }\n}\n",
			use:     "import _.sub.a as a;\n\nmodule {\n    func probe() -> I32 { return a.shared(); }\n}\n",
			wantErr: false,
		},
		{
			name:    "块内 public：跨子包不可见",
			decl:    "module {\n    public func shared() -> I32 { return 1; }\n}\n",
			use:     "import _.sub.a as a;\n\nfunc main() {\n    _ = a.shared();\n}\n",
			wantErr: true,
		},
		{
			name:    "壳内 public：跨子包可见",
			decl:    "public func ext() -> I32;\n\nmodule {\n    func ext() -> I32 { return 2; }\n}\n",
			use:     "import _.sub.a as a;\n\nfunc main() {\n    _ = a.ext();\n}\n",
			wantErr: false,
		},
		{
			name:    "壳内默认（子包内）：跨子包不可见",
			decl:    "struct Frame {\n    w: I32,\n};\n\nmodule {\n    func keep() -> I32 { return 0; }\n}\n",
			use:     "import _.sub.a as a;\n\nfunc main() {\n    let f: a.Frame = a.Frame { w: 1 };\n    _ = f;\n}\n",
			wantErr: true,
		},
		{
			name:    "壳内 public 类型：跨子包可见",
			decl:    "public struct Frame {\n    w: I32,\n};\n\nmodule {\n    func keep() -> I32 { return 0; }\n}\n",
			use:     "import _.sub.a as a;\n\nfunc main() {\n    let f: a.Frame = a.Frame { w: 1 };\n    _ = f;\n}\n",
			wantErr: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// 使用处放在 `_`（根子包），声明放在 `_.sub`（子包）：
			// 用「同子包」用例时把使用处也放进 `_.sub`
			useName := "_.main"
			if strings.Contains(c.name, "同子包") {
				useName = "_.sub.b"
			}
			mods := map[string]string{
				"_.sub.a":  c.decl,
				useName:    c.use,
			}
			if useName != "_.main" {
				mods["_.main"] = "func main() {}\n"
			}
			codes, _ := runModules(t, mods)
			if c.wantErr && !codes["E7020"] {
				t.Fatalf("期望 E7020，实得 %v", keys(codes))
			}
			if !c.wantErr && codes["E7020"] {
				t.Fatalf("不该报 E7020，实得 %v", keys(codes))
			}
		})
	}
}

// TestForwardDeclaration 守护「壳内声明 + 块内实现」的配对。
func TestForwardDeclaration(t *testing.T) {
	// 正常配对
	codes, _ := runModules(t, map[string]string{"_.main": "public func f(a: I32) -> I32;\n\nmodule {\n    func f(a: I32) -> I32 { return a; }\n}\n\nfunc main() {\n    _ = f(1);\n}\n"})
	if codes["E7031"] || codes["E3201"] {
		t.Fatalf("配对正常时不该报错，实得 %v", keys(codes))
	}
	// 声明了没实现
	codes, _ = runModules(t, map[string]string{"_.main": "public func ghost() -> I32;\n\nfunc main() {}\n"})
	if !codes["E7031"] {
		t.Fatalf("声明没实现应报 E7031，实得 %v", keys(codes))
	}
	// 签名不一致
	codes, _ = runModules(t, map[string]string{"_.main": "public func f(a: I32) -> I32;\n\nmodule {\n    func f(a: U8) -> I32 { return 0; }\n}\n\nfunc main() {}\n"})
	if !codes["E3201"] {
		t.Fatalf("签名不一致应报 E3201，实得 %v", keys(codes))
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
