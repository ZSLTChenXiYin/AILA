package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// run 执行一次 CLI，捕获 stdout / stderr 与退出码。
func run(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	oldOut, oldErr := stdout, stderr
	stdout, stderr = &out, &errb
	defer func() { stdout, stderr = oldOut, oldErr }()

	root := newRootCmd()
	root.SetArgs(args)
	code := ExitOK
	if err := root.Execute(); err != nil {
		if ee, ok := err.(*exitError); ok {
			code = ee.code
		} else {
			code = ExitInternal
			errb.WriteString(err.Error())
		}
	}
	return code, out.String(), errb.String()
}

func TestVersionCommand(t *testing.T) {
	code, out, _ := run(t, "version")
	if code != ExitOK {
		t.Fatalf("exit=%d", code)
	}
	for _, want := range []string{"aila", "AILA_HOME", "LLVM 后端", "C 编译器"} {
		if !strings.Contains(out, want) {
			t.Fatalf("version 输出缺少 %q：%s", want, out)
		}
	}
}

func TestRootVersionFlag(t *testing.T) {
	code, out, _ := run(t, "--version")
	if code != ExitOK {
		t.Fatalf("exit=%d", code)
	}
	if !strings.Contains(out, "aila") {
		t.Fatalf("--version 输出异常：%s", out)
	}
}

func TestExplainCommand(t *testing.T) {
	code, out, _ := run(t, "explain", "E4101")
	if code != ExitOK {
		t.Fatalf("exit=%d", code)
	}
	if !strings.Contains(out, "逃逸") || !strings.Contains(out, "怎么修") {
		t.Fatalf("explain 输出异常：%s", out)
	}
	code, _, _ = run(t, "explain", "E0000")
	if code != ExitUsage {
		t.Fatalf("未知码应返回 2，实得 %d", code)
	}
}

func TestTargetCommands(t *testing.T) {
	code, out, _ := run(t, "target", "list")
	if code != ExitOK {
		t.Fatalf("exit=%d", code)
	}
	if !strings.Contains(out, "esp32s3") {
		t.Fatalf("target list 输出异常：%s", out)
	}
	code, out, _ = run(t, "target", "show", "esp32s3")
	if code != ExitOK || !strings.Contains(out, "internal-ram") {
		t.Fatalf("target show 输出异常：%s", out)
	}
	code, _, _ = run(t, "target", "show", "nope")
	if code != ExitUsage {
		t.Fatalf("未知 board 应返回 2，实得 %d", code)
	}
}

func TestToolchainInfo(t *testing.T) {
	code, out, _ := run(t, "toolchain", "info")
	if code != ExitOK {
		t.Fatalf("exit=%d", code)
	}
	for _, want := range []string{"AILA_HOME", "绑定 LLVM 版本带", "ir2obj", "cc"} {
		if !strings.Contains(out, want) {
			t.Fatalf("toolchain info 输出缺少 %q：%s", want, out)
		}
	}
}

func TestNewProject(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo")
	code, out, errs := run(t, "new", path)
	if code != ExitOK {
		t.Fatalf("exit=%d out=%s err=%s", code, out, errs)
	}
	for _, name := range []string{"aila.toml", "main.aila", "README.md"} {
		if _, err := os.Stat(filepath.Join(path, name)); err != nil {
			t.Fatalf("缺文件 %s", name)
		}
	}
	code, _, _ = run(t, "new", path)
	if code != ExitDiag {
		t.Fatalf("非空目录应返回 1，实得 %d", code)
	}
}

// TestBindgenCommandGeneratesAil 守护 `aila bindgen`：它产出的是 `.ail` 绑定文件
// （不是 `.aila` 源码），且默认落在 `build/bindgen/`。
func TestBindgenCommandGeneratesAil(t *testing.T) {
	root := t.TempDir()
	write := func(name, body string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o666); err != nil {
			t.Fatal(err)
		}
	}
	write("aila.toml", "[package]\nname = \"bg\"\nversion = \"0.1.0\"\nentry = \"_.main\"\n\n[target]\nboard = \"host\"\n")
	write("main.aila", "import ch _.cpacks.kit;\n\nfunc main() {\n    _ = kit.kit_add(1, 2);\n}\n")
	write("cpacks/kit/include/kit.h", "int kit_add(int a, int b);\n")
	write("cpacks/kit/kit.c", "int kit_add(int a, int b) { return a + b; }\n")

	// 切到临时工程目录前先把 AILA_HOME 指回仓库：错误码表与 stdlib 从这里读
	wd, _ := os.Getwd()
	t.Setenv("AILA_HOME", filepath.Clean(filepath.Join(wd, "..", "..")))
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()

	code, out, errs := run(t, "bindgen")
	if code != ExitOK {
		t.Fatalf("bindgen 应成功（exit=0），实得 %d：%s", code, errs)
	}
	if !strings.Contains(out, "kit.ail") {
		t.Fatalf("输出应提到 kit.ail：%s", out)
	}
	data, err := os.ReadFile(filepath.Join(root, "build", "bindgen", "kit.ail"))
	if err != nil {
		t.Fatalf("没有生成 build/bindgen/kit.ail：%v", err)
	}
	if !strings.Contains(string(data), "kit_add") {
		t.Fatalf("绑定文件里应有 kit_add：%s", string(data))
	}
	if strings.Contains(string(data), "func kit_add(int") {
		t.Fatalf("绑定文件应是 AILA 声明，不该是 C 源码：%s", string(data))
	}
}

func TestDoctorRunsAndReportsLevels(t *testing.T) {
	code, out, _ := run(t, "doctor")
	if code != ExitOK && code != ExitEnv {
		t.Fatalf("doctor 退出码异常：%d", code)
	}
	for _, want := range []string{"AILA_HOME", "随包工具链", "LLVM 后端", "链接冒烟测试"} {
		if !strings.Contains(out, want) {
			t.Fatalf("doctor 输出缺少 %q：%s", want, out)
		}
	}
}

func TestInstallBadZip(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "bad.zip")
	if err := os.WriteFile(archive, []byte("not a zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errs := run(t, "toolchain", "install", archive, "--triple", "x86_64-unknown-linux-gnu")
	if code != ExitDiag {
		t.Fatalf("坏压缩包应返回 1，实得 %d", code)
	}
	if !strings.Contains(errs, "E7013") {
		t.Fatalf("应报 E7013：%s", errs)
	}
}

// TestIRCommandMatchesSnapshot 是 G3 的可见入口判据：`aila ir` 的输出必须与参考实现
// 的 `*.ir` 快照逐字节一致（快照由 go run ./tools/goldengen 生成）。
func TestIRCommandMatchesSnapshot(t *testing.T) {
	root := ".."
	for _, name := range []string{"ir_basic", "ir_wide"} {
		src := filepath.Join(root, "..", "tests", "corpus", name+".aila")
		want, err := os.ReadFile(filepath.Join(root, "..", "tests", "corpus", name+".ir"))
		if err != nil {
			t.Fatalf("读不到快照：%v", err)
		}
		code, out, errs := run(t, "ir", src, "--triple", "x86_64-w64-windows-gnu")
		if code != ExitOK {
			t.Fatalf("%s：exit=%d stderr=%s", name, code, errs)
		}
		if strings.TrimRight(out, " \t\r\n") != strings.TrimRight(string(want), " \t\r\n") {
			t.Fatalf("%s：IR 与快照不一致", name)
		}
	}
}

