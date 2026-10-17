package resolve

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/aila-lang/aila/internal/home"
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
	tool := ResolveCC("bundled")
	if tool.OK() {
		t.Fatalf("没有 bundle 时 `bundled` 不应可用：%s", tool.Describe())
	}
	if tool.Source != SourceMissing {
		t.Fatalf("source=%s", tool.Source)
	}
}

func TestSystemVersionFloorRejects(t *testing.T) {
	home.Override(repoRoot(t))
	defer home.Override("")
	for _, name := range []string{"clang", "gcc", "cc"} {
		tool := ResolveCC("system:" + name + ">=999")
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

func TestBundledPreferred(t *testing.T) {
	tmp := t.TempDir()
	home.Override(tmp)
	defer home.Override("")
	triple := "x86_64-unknown-linux-gnu"
	bin := home.BundleBin(triple)
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	// 用一个真实可执行的 shell 顶替 clang 只为验证「解析走到随包」
	cc := ""
	for _, n := range []string{"clang", "gcc", "cc"} {
		if p, err := execLookPath(n); err == nil {
			cc = p
			break
		}
	}
	if cc == "" {
		t.Skip("本机没有 C 编译器，无法伪造 bundle")
	}
	data, err := os.ReadFile(cc)
	if err != nil {
		t.Skip("读取编译器失败")
	}
	if err := os.WriteFile(filepath.Join(bin, "clang"+home.ExeSuffix()), data, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home.BundleDir(triple), "VERSION"), []byte("19.1.8\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tool := ResolveCC("auto")
	if tool.Source != SourceBundled {
		t.Fatalf("auto 应优先随包：%s", tool.Describe())
	}
	if !tool.OK() {
		t.Fatalf("随包工具应可用：%s（note=%s）", tool.Describe(), tool.Note)
	}
}

func TestOutOfBandBundleRejected(t *testing.T) {
	tmp := t.TempDir()
	home.Override(tmp)
	defer home.Override("")
	triple := "x86_64-unknown-linux-gnu"
	bin := home.BundleBin(triple)
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"clang", "gcc", "cc"} {
		if p, err := execLookPath(n); err == nil {
			data, _ := os.ReadFile(p)
			_ = os.WriteFile(filepath.Join(bin, "clang"+home.ExeSuffix()), data, 0o755)
			break
		}
	}
	if _, err := os.Stat(filepath.Join(bin, "clang"+home.ExeSuffix())); err != nil {
		t.Skip("本机没有 C 编译器，无法伪造 bundle")
	}
	if err := os.WriteFile(filepath.Join(home.BundleDir(triple), "VERSION"), []byte("11.0.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tool := ResolveLLVM("bundled")
	if tool.OK() {
		t.Fatalf("带外版本不应可用：%s", tool.Describe())
	}
	if tool.Note == "" || !contains(tool.Note, "版本带") {
		t.Fatalf("应说明版本带越界：%q", tool.Note)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
