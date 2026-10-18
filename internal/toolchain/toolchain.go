// Package toolchain 负责「怎么用」工具链：编 C、链接、冒烟测试、doctor（docs/specs/0003 §2/§4）。
package toolchain

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"aila/internal/home"
	"aila/internal/resolve"
)

// Check 是 doctor 的一行。
type Check struct {
	Name   string
	Level  string // ok | warn | fail
	Detail string
	Fix    string
}

// IRToObject 把 IR 文本降为目标文件（llc 或 clang）。
func IRToObject(irText, outPath string, triple string, tool resolve.Tool) error {
	if !tool.OK() {
		return fmt.Errorf("没有可用的 LLVM 后端：%s", tool.Note)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	var args []string
	if strings.HasPrefix(tool.Name, "llc") {
		args = []string{"-filetype=obj", "-O0", "-o", outPath, "-"}
	} else {
		args = []string{"-x", "ir", "-c", "-O0", "-o", outPath, "-"}
	}
	cmd := exec.Command(tool.Path, args...)
	cmd.Stdin = strings.NewReader(irText)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s 编译 IR 失败：%s", tool.Name, firstLine(stderr.String()))
	}
	return nil
}

func firstLine(s string) string {
	line := strings.TrimSpace(strings.SplitN(strings.TrimSpace(s), "\n", 2)[0])
	if line == "" {
		return "无输出"
	}
	if len(line) > 400 {
		return line[:400]
	}
	return line
}

// CompileC 编译若干 C 源为目标文件。
func CompileC(sources []string, outDir string, cc resolve.Tool, extra []string) ([]string, error) {
	if !cc.OK() {
		return nil, fmt.Errorf("没有可用的 C 编译器：%s", cc.Note)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	var objs []string
	for _, src := range sources {
		obj := filepath.Join(outDir, filepath.Base(src)+".o")
		args := append([]string{"-c", "-O0", "-o", obj, src}, extra...)
		cmd := exec.Command(cc.Path, args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("编译 C 源失败：%s\n%s", filepath.Base(src), firstLine(stderr.String()))
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

// Link 链接所有目标文件。
func Link(objs []string, outPath string, cc resolve.Tool, extra []string) error {
	if !cc.OK() {
		return fmt.Errorf("没有可用的链接器：%s", cc.Note)
	}
	if len(objs) == 0 {
		return fmt.Errorf("没有任何目标文件可链接")
	}
	args := append([]string{"-o", outPath}, objs...)
	args = append(args, extra...)
	cmd := exec.Command(cc.Path, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("链接失败：%s", firstLine(stderr.String()))
	}
	return nil
}

// SmokeTest 把最小 IR → .o → 链接 → 运行，验证整条后端链路。
func SmokeTest(cc, llvm resolve.Tool) (bool, string, string) {
	if !llvm.OK() {
		return false, "LLVM 后端不可用：" + llvm.Describe(), "装随包工具链：`aila toolchain install <release.zip>`"
	}
	if !cc.OK() {
		return false, "C 编译器/链接器不可用：" + cc.Describe(), "装随包工具链，或装 clang/gcc"
	}
	dir, err := os.MkdirTemp("", "aila_smoke_")
	if err != nil {
		return false, "创建临时目录失败：" + err.Error(), ""
	}
	defer os.RemoveAll(dir)
	obj := filepath.Join(dir, "smoke.o")
	exe := filepath.Join(dir, "smoke"+home.ExeSuffix())
	triple := resolve.HostTriple()
	ir := fmt.Sprintf("target triple = \"%s\"\n\ndefine i32 @main() {\nentry:\n  ret i32 7\n}\n", triple)
	if err := IRToObject(ir, obj, triple, llvm); err != nil {
		return false, "IR → .o 失败：" + err.Error(), "见 docs/specs/0003 §2"
	}
	if err := Link([]string{obj}, exe, cc, nil); err != nil {
		return false, "链接失败：" + err.Error(), "两端版本可能不兼容"
	}
	cmd := exec.Command(exe)
	_ = cmd.Run()
	if cmd.ProcessState == nil {
		return false, "冒烟程序无法运行", "检查能否执行临时目录下的程序"
	}
	if code := cmd.ProcessState.ExitCode(); code != 7 {
		return false, fmt.Sprintf("冒烟程序退出码 %d，期望 7", code), "见 docs/specs/0003 §2"
	}
	return true, "IR → .o → 链接 → 运行 全部通过", ""
}

// Doctor 返回自检表；只有 fail 会让 `aila doctor` 以退出码 3 结束。
func Doctor(ailaVersion, cliCC, cliLLVM string) []Check {
	checks := []Check{
		{Name: "AILA_HOME", Level: "ok", Detail: home.Root()},
		{Name: "aila（Go 版）", Level: "ok", Detail: ailaVersion},
	}
	tools := resolve.ResolveAll(cliCC, cliLLVM)
	irTool, ccTool, clangTool := tools[resolve.RoleIR2Obj], tools[resolve.RoleCC], tools[resolve.RoleClang]

	triple := resolve.HostTripleGuess()
	bundle := home.BundleDir(triple)
	if st, err := os.Stat(bundle); err == nil && st.IsDir() {
		level := "ok"
		detail := bundle
		if v := home.BundleLLVMVersion(triple); v != "" {
			detail += "（LLVM " + v + "）"
		}
		if irTool.Source != resolve.SourceBundled {
			level = "warn"
		}
		checks = append(checks, Check{Name: "随包工具链", Level: level, Detail: detail,
			Fix: "随包目录在，但解析没用上它——检查 toolchain/<triple>/bin/ 是否完整"})
	} else {
		checks = append(checks, Check{Name: "随包工具链", Level: "warn", Detail: "未安装（" + bundle + "）",
			Fix: "下载 release 后 `aila toolchain install <zip>`"})
	}

	irLevel := "ok"
	irFix := "装随包工具链：`aila toolchain install <release.zip>`"
	if !irTool.OK() {
		irLevel = "fail"
	}
	checks = append(checks, Check{Name: "LLVM 后端（IR → .o）", Level: irLevel,
		Detail: irTool.Describe(), Fix: irFix})

	ccLevel := "ok"
	if !ccTool.OK() {
		ccLevel = "fail"
	}
	checks = append(checks, Check{Name: "C 编译器 / 链接器", Level: ccLevel,
		Detail: ccTool.Describe(),
		Fix:    "装随包工具链，或清单写 [build.toolchain] cc = \"system:gcc\""})

	clangLevel := "ok"
	if !clangTool.OK() {
		clangLevel = "warn"
	}
	checks = append(checks, Check{Name: "Clang（bindgen 首选档）", Level: clangLevel,
		Detail: clangTool.Describe(),
		Fix:    "装随包工具链可获完整 C 声明映射（非必装，缺失时用 builtin 兜底档）"})

	runtimeSrc := home.RuntimeSrc()
	runtimeLevel := "ok"
	if _, err := os.Stat(runtimeSrc); err != nil {
		runtimeLevel = "fail"
	}
	checks = append(checks, Check{Name: "运行时源码", Level: runtimeLevel, Detail: runtimeSrc,
		Fix: "重新解压 AILA 发布包"})

	stdlib := home.StdlibDir()
	stdlibLevel := "ok"
	if st, err := os.Stat(stdlib); err != nil || !st.IsDir() {
		stdlibLevel = "fail"
	}
	checks = append(checks, Check{Name: "标准库", Level: stdlibLevel, Detail: stdlib,
		Fix: "重新解压 AILA 发布包"})

	if irTool.OK() && ccTool.OK() {
		ok, detail, fix := SmokeTest(ccTool, irTool)
		level := "ok"
		if !ok {
			level = "fail"
		}
		checks = append(checks, Check{Name: "链接冒烟测试", Level: level, Detail: detail, Fix: fix})
	} else {
		checks = append(checks, Check{Name: "链接冒烟测试", Level: "fail", Detail: "前置项未通过，跳过",
			Fix: "先修好上面两项"})
	}
	return checks
}

// Failed 统计 fail 数。
func Failed(checks []Check) int {
	n := 0
	for _, c := range checks {
		if c.Level == "fail" {
			n++
		}
	}
	return n
}
