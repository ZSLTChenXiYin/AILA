package home

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller 失败")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

func TestOverride(t *testing.T) {
	t.Setenv(HomeEnv, "")
	tmp := t.TempDir()
	Override(tmp)
	defer Override("")
	if got := Root(); got != mustAbs(t, tmp) {
		t.Fatalf("Root=%s, want %s", got, tmp)
	}
}

func TestShareRootInRepo(t *testing.T) {
	Override(repoRoot(t))
	defer Override("")
	if got := ShareRoot(); got != filepath.Join(mustAbs(t, repoRoot(t)), "share", "aila") {
		t.Fatalf("ShareRoot=%s", got)
	}
	if st, err := os.Stat(DataDir()); err != nil || !st.IsDir() {
		t.Fatalf("DataDir 不可用：%s", DataDir())
	}
	if _, err := os.Stat(RuntimeSrc()); err != nil {
		t.Fatalf("RuntimeSrc 不可用：%s", RuntimeSrc())
	}
	if st, err := os.Stat(StdlibDir()); err != nil || !st.IsDir() {
		t.Fatalf("StdlibDir 不可用：%s", StdlibDir())
	}
}

func TestBundleHelpers(t *testing.T) {
	tmp := t.TempDir()
	Override(tmp)
	defer Override("")
	triple := "x86_64-unknown-linux-gnu"
	bin := BundleBin(triple)
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	name := "clang" + ExeSuffix()
	if err := os.WriteFile(filepath.Join(bin, name), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(BundleDir(triple), "VERSION"), []byte("19.1.8\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bins := BundleBins(triple)
	if _, ok := bins["clang"]; !ok {
		t.Fatalf("BundleBins 没找到 clang：%v", bins)
	}
	if v := BundleLLVMVersion(triple); v != "19.1.8" {
		t.Fatalf("VERSION=%q", v)
	}
	if got := InstalledTriples(); len(got) != 1 || got[0] != triple {
		t.Fatalf("InstalledTriples=%v", got)
	}
}

func mustAbs(t *testing.T, p string) string {
	t.Helper()
	a, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return a
}
