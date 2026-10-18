package deps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aila/internal/diag"
)

// writePkg 造一个最小包（清单 + 一个模块）。
func writePkg(t *testing.T, dir, name, version, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o777); err != nil {
		t.Fatal(err)
	}
	man := "[package]\nname = \"" + name + "\"\nversion = \"" + version + "\"\nentry = \"_.main\"\n\n[target]\nboard = \"host\"\n"
	if err := os.WriteFile(filepath.Join(dir, "aila.toml"), []byte(man), 0o666); err != nil {
		t.Fatal(err)
	}
	if body == "" {
		body = "func main() {}\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "main.aila"), []byte(body), 0o666); err != nil {
		t.Fatal(err)
	}
}

// writeApp 写一个只有清单的根包（带依赖声明）。
func writeApp(t *testing.T, dir, depsTOML string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o777); err != nil {
		t.Fatal(err)
	}
	man := "[package]\nname = \"app\"\nversion = \"0.1.0\"\nentry = \"_.main\"\n\n[target]\nboard = \"host\"\n\n" + depsTOML
	if err := os.WriteFile(filepath.Join(dir, "aila.toml"), []byte(man), 0o666); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(dir, "main.aila"), []byte("func main() {}\n"), 0o666)
}

// TestPathDependency 覆盖本地路径依赖的解析、锁文件与漂移检测。
func TestPathDependency(t *testing.T) {
	root := t.TempDir()
	writePkg(t, filepath.Join(root, "mylib"), "mylib", "0.1.0", "public func f() -> I32 { return 1; }\n")
	writeApp(t, filepath.Join(root, "app"), "[dependencies]\nmylib = { path = \"../mylib\" }\n")

	bag := &diag.Bag{}
	g := Resolve(filepath.Join(root, "app"), true, bag)
	if bag.HasErrors() {
		t.Fatalf("解析失败：\n%s", bag.Render(false))
	}
	res := g.ByPkg["mylib"]
	if res == nil {
		t.Fatalf("没有解析到 mylib：%v", g.ByPkg)
	}
	if res.Version != "0.1.0" || res.Hash == "" || res.Source != "path:"+filepath.Join("..", "mylib") {
		t.Fatalf("解析结果不对：%+v", res)
	}
	if err := g.Write(); err != nil {
		t.Fatal(err)
	}

	// 锁文件应是可移植形态（不含本机绝对路径）
	lock, err := os.ReadFile(filepath.Join(root, "app", LockFileName))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(lock), root) {
		t.Fatalf("锁文件里不该出现本机绝对路径：\n%s", lock)
	}

	// 再解析一次：锁命中，不报错
	bag2 := &diag.Bag{}
	Resolve(filepath.Join(root, "app"), false, bag2)
	if bag2.HasErrors() {
		t.Fatalf("锁命中时不该报错：\n%s", bag2.Render(false))
	}

	// 改依赖内容 → E7023（内容变了）
	if err := os.WriteFile(filepath.Join(root, "mylib", "main.aila"),
		[]byte("public func f() -> I32 { return 2; }\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	bag3 := &diag.Bag{}
	Resolve(filepath.Join(root, "app"), false, bag3)
	if !hasCode(bag3, "E7023") {
		t.Fatalf("内容变了应报 E7023，实得：\n%s", bag3.Render(false))
	}

	// --update 之后恢复正常
	bag4 := &diag.Bag{}
	Resolve(filepath.Join(root, "app"), true, bag4)
	if bag4.HasErrors() {
		t.Fatalf("--update 后不该报错：\n%s", bag4.Render(false))
	}
}

// TestVendorOffline 覆盖 vendor：路径断掉后仍能构建。
func TestVendorOffline(t *testing.T) {
	root := t.TempDir()
	writePkg(t, filepath.Join(root, "mylib"), "mylib", "1.0.0", "public func f() -> I32 { return 3; }\n")
	appDir := filepath.Join(root, "app")
	writeApp(t, appDir, "[dependencies]\nmylib = { path = \"../mylib\" }\n")

	bag := &diag.Bag{}
	g := Resolve(appDir, true, bag)
	if bag.HasErrors() {
		t.Fatalf("解析失败：\n%s", bag.Render(false))
	}
	if err := g.Vendor(); err != nil {
		t.Fatal(err)
	}
	if err := g.Write(); err != nil {
		t.Fatal(err)
	}
	// 断掉原路径，只留 vendor
	if err := os.RemoveAll(filepath.Join(root, "mylib")); err != nil {
		t.Fatal(err)
	}
	bag2 := &diag.Bag{}
	g2 := Resolve(appDir, false, bag2)
	if bag2.HasErrors() {
		t.Fatalf("有 vendor 时不该报错：\n%s", bag2.Render(false))
	}
	if !g2.ByPkg["mylib"].Vendored {
		t.Fatal("应走 vendor")
	}
}

// TestVersionMismatch 覆盖清单版本要求不满足。
func TestVersionMismatch(t *testing.T) {
	root := t.TempDir()
	writePkg(t, filepath.Join(root, "mylib"), "mylib", "2.5.0", "")
	writeApp(t, filepath.Join(root, "app"),
		"[dependencies]\nmylib = { path = \"../mylib\", version = \"1.0\" }\n")
	bag := &diag.Bag{}
	Resolve(filepath.Join(root, "app"), true, bag)
	if !hasCode(bag, "E7023") {
		t.Fatalf("版本不匹配应报 E7023：\n%s", bag.Render(false))
	}
}

// TestCycleRejected 覆盖依赖成环。
func TestCycleRejected(t *testing.T) {
	root := t.TempDir()
	writePkg(t, filepath.Join(root, "a"), "a", "0.1.0", "")
	writeApp(t, filepath.Join(root, "a"), "[dependencies]\nb = { path = \"../b\" }\n")
	writePkg(t, filepath.Join(root, "b"), "b", "0.1.0", "")
	writePkg(t, filepath.Join(root, "app"), "app", "0.1.0", "")
	writeApp(t, filepath.Join(root, "app"), "[dependencies]\na = { path = \"../a\" }\n")
	// b 依赖 a → a → b 成环
	manB := "[package]\nname = \"b\"\nversion = \"0.1.0\"\nentry = \"_.main\"\n\n[target]\nboard = \"host\"\n\n[dependencies]\na = { path = \"../a\" }\n"
	if err := os.WriteFile(filepath.Join(root, "b", "aila.toml"), []byte(manB), 0o666); err != nil {
		t.Fatal(err)
	}
	bag := &diag.Bag{}
	Resolve(filepath.Join(root, "app"), true, bag)
	if !hasCode(bag, "E7025") {
		t.Fatalf("成环应报 E7025：\n%s", bag.Render(false))
	}
}

// TestBadDependencyDeclaration 覆盖声明本身不合法。
func TestBadDependencyDeclaration(t *testing.T) {
	root := t.TempDir()
	writeApp(t, filepath.Join(root, "app"), "[dependencies]\nbad = { }\n")
	bag := &diag.Bag{}
	Resolve(filepath.Join(root, "app"), true, bag)
	if !hasCode(bag, "E7021") {
		t.Fatalf("空声明应报 E7021：\n%s", bag.Render(false))
	}
	// git 依赖必须锁 tag/rev
	root2 := t.TempDir()
	writeApp(t, filepath.Join(root2, "app"),
		"[dependencies]\nbad = { git = \"https://example.com/x\" }\n")
	bag2 := &diag.Bag{}
	Resolve(filepath.Join(root2, "app"), true, bag2)
	if !hasCode(bag2, "E7021") {
		t.Fatalf("git 依赖没锁 tag/rev 应报 E7021：\n%s", bag2.Render(false))
	}
}

// TestVersionMatches 单测版本前缀匹配。
func TestVersionMatches(t *testing.T) {
	cases := []struct {
		have, want string
		ok         bool
	}{
		{"0.1.3", "0.1", true},
		{"0.1.0", "0.1.0", true},
		{"0.2.0", "0.1", false},
		{"1.0.0", "", true},
		{"0.10.0", "0.1", false}, // 前缀匹配要求「.」边界
	}
	for _, c := range cases {
		if got := versionMatches(c.have, c.want); got != c.ok {
			t.Errorf("versionMatches(%q,%q)=%v 期望 %v", c.have, c.want, got, c.ok)
		}
	}
}

func hasCode(bag *diag.Bag, code string) bool {
	for _, d := range bag.Items {
		if d.Code == code {
			return true
		}
	}
	return false
}
