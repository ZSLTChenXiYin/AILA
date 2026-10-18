package skill

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDirFollowsHome 验证技能目录按 AILA_HOME 的相对路径解析（解压即用的前提）。
func TestDirFollowsHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AILA_HOME", dir)
	want := filepath.Join(dir, "share", "aila", "skills", "aila")
	if got := Dir(); got != want {
		t.Fatalf("技能目录应为 %s，实得 %s", want, got)
	}
}

// TestFilesAndRead 验证入口在最前、路径用 /、能读回内容。
func TestFilesAndRead(t *testing.T) {
	home := t.TempDir()
	t.Setenv("AILA_HOME", home)
	root := Dir()
	if err := os.MkdirAll(filepath.Join(root, "reference"), 0o777); err != nil {
		t.Fatal(err)
	}
	write := func(rel, body string) {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(rel)), []byte(body), 0o666); err != nil {
			t.Fatal(err)
		}
	}
	write("reference/b.md", "b")
	write("SKILL.md", "# 技能")
	write("reference/a.md", "a")

	files, err := Files()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Fatalf("应有 3 个文件：%+v", files)
	}
	if files[0].Rel != Entry {
		t.Fatalf("SKILL.md 应排最前：%+v", files)
	}
	if files[1].Rel != "reference/a.md" || files[2].Rel != "reference/b.md" {
		t.Fatalf("其余按路径升序：%+v", files)
	}
	if files[0].Bytes != len("# 技能") {
		t.Fatalf("字节数不对：%+v", files[0])
	}
	if got, err := Read("reference/a.md"); err != nil || got != "a" {
		t.Fatalf("Read 失败：%q %v", got, err)
	}
	if got, err := Read(`reference\b.md`); err != nil || got != "b" {
		t.Fatalf("反斜杠路径也要能读：%q %v", got, err)
	}
}

// TestInstall 验证安装会把整棵树复制到 <dst>/aila/，且内容一致。
func TestInstall(t *testing.T) {
	src := t.TempDir()
	t.Setenv("AILA_HOME", src)
	root := Dir()
	if err := os.MkdirAll(filepath.Join(root, "reference"), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, Entry), []byte("# 技能"), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "reference", "x.md"), []byte("x"), 0o666); err != nil {
		t.Fatal(err)
	}

	dst := t.TempDir()
	installed, err := Install(dst)
	if err != nil {
		t.Fatal(err)
	}
	if installed != filepath.Join(dst, Name) {
		t.Fatalf("安装目标应为 %s，实得 %s", filepath.Join(dst, Name), installed)
	}
	want := map[string]string{Entry: "# 技能", "reference/x.md": "x"}
	for rel, body := range want {
		data, err := os.ReadFile(filepath.Join(installed, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("安装后读不到 %s：%v", rel, err)
		}
		if string(data) != body {
			t.Fatalf("%s 内容不对：%q", rel, string(data))
		}
	}
}
