package resolve

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"aila/internal/home"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller 失败")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

func TestVersionBand(t *testing.T) {
	lo, hi := VersionBand()
	if lo != "18" || hi != "23" {
		t.Fatalf("版本带异常：%s..%s", lo, hi)
	}
	if !InBand("19.1.8") || !InBand("22.1.0") {
		t.Fatal("19.1.8 / 22.1.0 应在带内")
	}
	if InBand("17.0.6") || InBand("") {
		t.Fatal("17.0.6 与空版本不应在带内")
	}
}

func TestBundledOnlyNeverFallsBack(t *testing.T) {
	home.Override(t.TempDir())
	defer home.Override("")
	tool := ResolveCC("", "bundled")
	if tool.OK() {
		t.Fatalf("没有工具链时 `bundled` 不应可用：%s", tool.Describe())
	}
	if tool.Source != SourceMissing {
		t.Fatalf("source=%s", tool.Source)
	}
}

func TestSystemVersionFloorRejects(t *testing.T) {
	home.Override(repoRoot(t))
	defer home.Override("")
	for _, name := range []string{"clang", "gcc", "cc"} {
		tool := ResolveCC("", "system:"+name+">=999")
		if tool.OK() {
			continue // 该名字不存在时本来就 OK()==false
		}
		if tool.Note == "" {
			t.Fatalf("版本不符时应有说明（%s）", name)
		}
		return
	}
	t.Skip("本机没有 clang/gcc/cc 可供验证")
}

// fakeGCC 把系统 gcc/clang 拷进 gcc/<triple>/<ver>/bin（冒充目标 GCC）。
// gcc 目录解析只看目录名与驱动名，不需要探测，任何可执行文件都能验证解析路径。
func fakeGCC(t *testing.T, triple, version string) {
	t.Helper()
	src := ""
	for _, n := range []string{"clang", "gcc", "cc"} {
		if p, err := execLookPath(n); err == nil {
			src = p
			break
		}
	}
	if src == "" {
		t.Skip("本机没有 C 编译器，无法伪造 GCC 目录")
	}
	data, err := os.ReadFile(src)
	if err != nil {
		t.Skip("读取编译器失败")
	}
	bin := filepath.Join(home.ToolchainsRoot(), "gcc", triple, version, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, triple+"-gcc"+home.ExeSuffix()), data, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestGCCDirWinsForCC(t *testing.T) {
	home.Override(t.TempDir())
	defer home.Override("")
	triple := "xtensa-lx106-elf"
	fakeGCC(t, triple, "8.4.0")
	tool := ResolveCC(triple, "auto")
	if tool.Source != SourceBundled || tool.Family != "gcc" {
		t.Fatalf("GCC 目录应优先承担 cc：%s", tool.Describe())
	}
	if !tool.OK() {
		t.Fatalf("GCC 工具应可用（无版本带约束）：%s（note=%s）", tool.Describe(), tool.Note)
	}
}

func TestParseRegisteredTargets(t *testing.T) {
	out := "  x86    -            32-bit x86\n" +
		"  x86-64 -            64-bit x86\n" +
		"  xtensa -            Xtensa\n\nSome trailer\n"
	got := parseRegisteredTargets(out)
	if len(got) != 3 || got[0] != "x86" || got[2] != "xtensa" {
		t.Fatalf("解析异常：%v", got)
	}
	if !supportsArchText(out, "Xtensa") || !supportsArchText(out, "x86_64") {
		t.Fatal("归一化比较失败（xtensa / x86_64）")
	}
	if supportsArchText(out, "riscv32") {
		t.Fatal("riscv32 不应命中")
	}
}

// supportsArchText 直接对文本做归一化比较（不经进程探测）。
func supportsArchText(out, arch string) bool {
	want := normArch(arch)
	for _, got := range parseRegisteredTargets(out) {
		if normArch(got) == want {
			return true
		}
	}
	return false
}

// TestBundledClangPreferred 守护「解析走到随包」：需要真实 clang
// （探测 --print-targets / -print-target-triple），伪造二进制过不了。
func TestBundledClangPreferred(t *testing.T) {
	home.Override(t.TempDir())
	defer home.Override("")
	clangPath := ""
	if p, err := execLookPath("clang"); err == nil {
		clangPath = p
	}
	if clangPath == "" {
		t.Skip("本机没有 clang，无法伪造 clang 家族安装")
	}
	data, err := os.ReadFile(clangPath)
	if err != nil {
		t.Skip("读取 clang 失败")
	}
	bin := filepath.Join(home.ToolchainsRoot(), "clang", "19.1.8", "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "clang"+home.ExeSuffix()), data, 0o755); err != nil {
		t.Fatal(err)
	}
	// cc：默认三元组探测命中即随包优先
	triple := HostTriple()
	tool := ResolveCC(triple, "auto")
	if tool.Source != SourceBundled {
		t.Fatalf("auto 应优先随包：%s", tool.Describe())
	}
	if !tool.OK() {
		t.Fatalf("随包工具应可用：%s（note=%s）", tool.Describe(), tool.Note)
	}
	// ir2obj：同一家族应被选中（通版 clang 自报支持宿主架构）
	llvm := ResolveLLVMFor(triple, "auto")
	if llvm.Source != SourceBundled || !llvm.OK() {
		t.Fatalf("ir2obj 应落在随包：%s", llvm.Describe())
	}
}

func TestOutOfBandClangRejected(t *testing.T) {
	home.Override(t.TempDir())
	defer home.Override("")
	clangPath := ""
	if p, err := execLookPath("clang"); err == nil {
		clangPath = p
	}
	if clangPath == "" {
		t.Skip("本机没有 clang，无法伪造 clang 家族安装")
	}
	data, err := os.ReadFile(clangPath)
	if err != nil {
		t.Skip("读取 clang 失败")
	}
	bin := filepath.Join(home.ToolchainsRoot(), "clang", "11.0.0", "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "clang"+home.ExeSuffix()), data, 0o755); err != nil {
		t.Fatal(err)
	}
	tool := ResolveLLVMFor("", "bundled")
	if tool.OK() {
		t.Fatalf("带外版本不应可用：%s", tool.Describe())
	}
	if tool.Note == "" || !contains(tool.Note, "版本带") {
		t.Fatalf("应说明版本带越界：%q", tool.Note)
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
