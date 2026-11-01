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

func TestToolchainsHelpers(t *testing.T) {
	tmp := t.TempDir()
	Override(tmp)
	defer Override("")
	// 安装根是 .local/toolchains（本地产物）；随包 toolchains/ 也要一起扫。
	installRoot := filepath.Join(tmp, ".local", "toolchains")
	bundleRoot := filepath.Join(tmp, "toolchains")
	// 通版 clang 家族：放在随包只读根，验证「也认随包」
	clangBin := filepath.Join(bundleRoot, "clang", "19.1.8", "bin")
	if err := os.MkdirAll(clangBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(clangBin, "clang"+ExeSuffix()), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	// 厂商 fork 家族：放在本地安装根
	espBin := filepath.Join(installRoot, "esp-clang", "22.1.4", "bin")
	if err := os.MkdirAll(espBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(espBin, "llc"+ExeSuffix()), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	// gcc 系
	gccBin := filepath.Join(installRoot, "gcc", "xtensa-lx106-elf", "8.4.0", "bin")
	if err := os.MkdirAll(gccBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gccBin, "xtensa-lx106-elf-gcc"+ExeSuffix()), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	// toolchains.json 是文件不是目录，家族扫描必须跳过它
	if err := os.WriteFile(filepath.Join(bundleRoot, "toolchains.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	clangs := InstalledClangs()
	if len(clangs) != 2 {
		t.Fatalf("InstalledClangs 应有 2 个安装：%+v", clangs)
	}
	fams := map[string]string{}
	for _, c := range clangs {
		fams[c.Family] = c.Version
	}
	if fams["clang"] != "19.1.8" || fams["esp-clang"] != "22.1.4" {
		t.Fatalf("家族版本异常：%v", fams)
	}
	gccs := InstalledGCCs()
	if len(gccs) != 1 || gccs[0].Triple != "xtensa-lx106-elf" || gccs[0].Version != "8.4.0" {
		t.Fatalf("InstalledGCCs=%+v", gccs)
	}
	// 顺序是「先本地安装根，再随包」：本地安装的就是排在前面那个
	var esp *ClangInstall
	for i := range clangs {
		if clangs[i].Family == "esp-clang" {
			esp = &clangs[i]
		}
	}
	if esp == nil {
		t.Fatalf("没扫到 esp-clang：%+v", clangs)
	}
	if clangs[0].Family != "esp-clang" {
		t.Fatalf("本地安装根应排在随包前面：%+v", clangs)
	}
	bins := Bins(esp.Dir)
	if _, ok := bins["llc"]; !ok {
		t.Fatalf("Bins 没找到 llc：%v", bins)
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
