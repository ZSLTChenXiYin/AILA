package build

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ailang/aila/internal/diag"
)

// repoRoot 从本文件定位仓库根（internal/build → ../..）。
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("无法定位测试文件")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// exampleNames 是全部示例（cinterop 覆盖 bindgen + C 侧编译链接）。
var exampleNames = []string{"hello", "control", "region", "struct", "result", "cinterop"}

// TestCheckExamples 验证 Go 版能独立完成六示例中五个的「解析 + 多模块语义」。
func TestCheckExamples(t *testing.T) {
	root := repoRoot(t)
	for _, name := range exampleNames {
		t.Run(name, func(t *testing.T) {
			res := Check(filepath.Join(root, "examples", name), Options{})
			if res.Diags.HasErrors() {
				t.Fatalf("%s 检查失败：\n%s", name, res.Diags.Render(false))
			}
			if res.Program == nil {
				t.Fatalf("%s 没有产出模块图", name)
			}
			if res.Program.Unit(res.Manifest.Entry) == nil {
				t.Fatalf("%s 入口模块 `%s` 不在模块图里", name, res.Manifest.Entry)
			}
		})
	}
}

// TestExampleIRParity 是 G4 的多模块判据：Go 版对每个示例模块产出的 IR，必须与
// Python 参考实现生成的快照（tests/corpus/examples/）逐字节一致。
func TestExampleIRParity(t *testing.T) {
	root := repoRoot(t)
	for _, name := range exampleNames {
		t.Run(name, func(t *testing.T) {
			res := Load(filepath.Join(root, "examples", name), Options{})
			if res.Diags.HasErrors() {
				t.Fatalf("%s 装载失败：\n%s", name, res.Diags.Render(false))
			}
			irs := res.GenerateIR()
			if res.Diags.HasErrors() {
				t.Fatalf("%s 代码生成失败：\n%s", name, res.Diags.Render(false))
			}
			snapDir := filepath.Join(root, "tests", "corpus", "examples", name)
			entries, err := os.ReadDir(snapDir)
			if err != nil {
				t.Fatalf("%s 没有 IR 快照（先跑 go run ./tools/goldengen）：%v", name, err)
			}
			checked := 0
			for _, e := range entries {
				if !strings.HasSuffix(e.Name(), ".ll") {
					continue
				}
				mod := strings.TrimSuffix(e.Name(), ".ll")
				// 快照文件名是模块名扁平化后的样子
				var got string
				for modName, ir := range irs {
					if safeName(modName) == mod {
						got = ir
						break
					}
				}
				want, err := os.ReadFile(filepath.Join(snapDir, e.Name()))
				if err != nil {
					t.Fatalf("读不到快照 %s：%v", e.Name(), err)
				}
				if strings.TrimRight(got, " \t\r\n") != strings.TrimRight(string(want), " \t\r\n") {
					t.Fatalf("%s/%s 与参考实现的 IR 不一致", name, e.Name())
				}
				checked++
			}
			if checked == 0 {
				t.Fatalf("%s 没有可比对的快照", name)
			}
		})
	}
}
// TestExamplesEndToEnd 是 G4 的端到端判据：Go 版从源码一路编到可执行文件，
// 运行后的 stdout 必须与 `tests/corpus/examples/<名字>/expected.out` 逐字节一致。
//
// 没有 LLVM 后端（无随包 llc/clang、PATH 里也没有）或没有 C 编译器时跳过——
// 这两种环境限制与实现无关，CI 上必须真跑。
func TestExamplesEndToEnd(t *testing.T) {
	root := repoRoot(t)
	for _, name := range exampleNames {
		t.Run(name, func(t *testing.T) {
			res := Build(filepath.Join(root, "examples", name), Options{})
			if !res.OK {
				if reason, missing := toolchainMissing(res.Diags); missing {
					t.Skip(reason)
				}
				t.Fatalf("%s 构建失败：\n%s", name, res.Diags.Render(false))
			}
			want, err := os.ReadFile(filepath.Join(root, "tests", "corpus", "examples", name, "expected.out"))
			if err != nil {
				t.Fatalf("读不到期望输出：%v", err)
			}
			cmd := exec.Command(res.Exe)
			var out strings.Builder
			cmd.Stdout = &out
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("%s 运行失败：%v", name, err)
			}
			got := strings.ReplaceAll(out.String(), "\r\n", "\n")
			wantText := strings.ReplaceAll(string(want), "\r\n", "\n")
			if got != wantText {
				t.Fatalf("%s 输出不一致：\n--- 实得 ---\n%s\n--- 期望 ---\n%s", name, got, wantText)
			}
		})
	}
}

// toolchainMissing 判断构建失败是否只是环境缺工具链。
func toolchainMissing(bag *diag.Bag) (string, bool) {
	for _, item := range bag.Items {
		if item.Code == "E8010" || item.Code == "E8011" || item.Code == "E8020" {
			return "跳过：本机没有可用的 LLVM 后端 / C 编译器（" + item.Code + "）", true
		}
	}
	return "", false
}
// TestCrossProjectDependency 是跨项目依赖的端到端判据（docs/rfcs/0004）：
// 两个真实项目，根包 import 依赖包的 public 符号，编译并运行出预期输出；
// 非 public 符号必须被 E7020 拒绝。
func TestCrossProjectDependency(t *testing.T) {
	root := t.TempDir()
	libDir := filepath.Join(root, "mylib")
	appDir := filepath.Join(root, "app")
	mustWrite := func(path, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o666); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite(filepath.Join(libDir, "aila.toml"),
		"[package]\nname = \"mylib\"\nversion = \"0.1.0\"\nentry = \"_.main\"\n\n[target]\nboard = \"host\"\n")
	mustWrite(filepath.Join(libDir, "main.aila"), "func main() {}\n")
	mustWrite(filepath.Join(libDir, "math", "ops.aila"),
		"public func twice(v: I32) -> I32 {\n    return add(v, v);\n}\n\n"+
			"func add(a: I32, b: I32) -> I32 {\n    return a + b;\n}\n")
	mustWrite(filepath.Join(appDir, "aila.toml"),
		"[package]\nname = \"app\"\nversion = \"0.1.0\"\nentry = \"_.main\"\n\n[target]\nboard = \"host\"\n\n"+
			"[dependencies]\nmylib = { path = \"../mylib\" }\n")
	mustWrite(filepath.Join(appDir, "main.aila"),
		"import std.io;\nimport mylib.math.ops as ops;\n\n"+
			"func main() {\n    io.print_i32_nl(ops.twice(21));\n}\n")

	// ① 正常路径：构建成功、运行输出 42
	res := Build(appDir, Options{})
	if !res.OK {
		if r, missing := toolchainMissing(res.Diags); missing {
			t.Skip(r)
		}
		t.Fatalf("跨项目构建失败：\n%s", res.Diags.Render(false))
	}
	out, err := exec.Command(res.Exe).Output()
	if err != nil {
		t.Fatalf("运行失败：%v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "42" {
		t.Fatalf("输出 %q，期望 42", got)
	}

	// ② 非 public：必须报 E7020
	mustWrite(filepath.Join(appDir, "main.aila"),
		"import mylib.math.ops as ops;\n\nfunc main() {\n    _ = ops.add(1, 2);\n}\n")
	res2 := Check(appDir, Options{})
	found := false
	for _, d := range res2.Diags.Items {
		if d.Code == "E7020" {
			found = true
		}
	}
	if !found {
		t.Fatalf("非 public 跨包引用应报 E7020：\n%s", res2.Diags.Render(false))
	}

	// ③ 未声明的依赖根：E7010
	mustWrite(filepath.Join(appDir, "main.aila"),
		"import nomath.calc as calc;\n\nfunc main() {\n    _ = calc.f();\n}\n")
	res3 := Check(appDir, Options{})
	found = false
	for _, d := range res3.Diags.Items {
		if d.Code == "E7010" {
			found = true
		}
	}
	if !found {
		t.Fatalf("未声明的依赖根应报 E7010：\n%s", res3.Diags.Render(false))
	}
}
