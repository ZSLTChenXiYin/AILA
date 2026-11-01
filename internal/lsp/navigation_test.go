package lsp

import (
	"strings"
	"testing"
)

// navProject 建一个「壳内前向声明 + 块内实现」的工程（跨包导出必须这么写）。
func navProject(t *testing.T) (*Workspace, string) {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, dir+"/aila.toml",
		"[package]\nname = \"demo\"\nversion = \"0.1.0\"\nentry = \"_.main\"\n")
	mustWrite(t, dir+"/main.aila", strings.Join([]string{
		"import _.sub.util;",
		"",
		"public struct Point {",
		"    x: I32,",
		"};",
		"",
		"public func twice(v: I32) -> I32;",
		"",
		"module {",
		"    func twice(v: I32) -> I32 {",
		"        return v + v;",
		"    }",
		"",
		"    func main() {",
		"        let p: Point = Point { x: 1 };",
		"        let n = twice(p.x);",
		"        let m = util.double(n);",
		"        _ = m;",
		"    }",
		"}",
		"",
	}, "\n"))
	mustWrite(t, dir+"/sub/util.aila", strings.Join([]string{
		"public func double(v: I32) -> I32 {",
		"    return v + v;",
		"}",
		"",
	}, "\n"))
	w := NewWorkspace()
	w.SetRoot(dir)
	return w, Path2URI(dir + "/main.aila")
}

func TestTypeDefinition(t *testing.T) {
	w, uri := navProject(t)
	// 变量 `p: Point` 的 `p` → struct Point 的声明（第 3 行，0 基 2）
	pos := posAtWord(t, w, uri, "let p: Point", "p")
	defs := w.TypeDefinition(uri, pos)
	if len(defs) == 0 {
		t.Fatal("`p` 应能跳到它的类型声明")
	}
	if defs[0].Range.Start.Line != 2 {
		t.Fatalf("类型声明行不对：%+v", defs[0].Range)
	}
	// 内置类型没有声明处
	var i32Pos Position
	for i, line := range lineContents(mustText(t, w, uri)) {
		if idx := strings.Index(line, "func twice(v: I32)"); idx >= 0 {
			i32Pos = Position{Line: i, Character: idx + strings.Index(line[idx:], "I32") + 1}
			break
		}
	}
	if got := w.TypeDefinition(uri, i32Pos); len(got) != 0 {
		t.Fatalf("内置类型不该有类型定义：%+v", got)
	}
}

func TestImplementationPairsDeclaration(t *testing.T) {
	w, uri := navProject(t)
	// 在前向声明上 → 能跳到块内实现
	declPos := posAtWord(t, w, uri, "public func twice", "twice")
	impls := w.Implementation(uri, declPos)
	if len(impls) == 0 {
		t.Fatal("前向声明应能跳到实现")
	}
	foundBlock := false
	for _, l := range impls {
		if l.Range.Start.Line == 9 { // 块内实现那一行
			foundBlock = true
		}
	}
	if !foundBlock {
		t.Fatalf("应包含块内实现的位置：%+v", impls)
	}
	// 在调用点上也能找到同一对
	callPos := posAtWord(t, w, uri, "let n = twice", "twice")
	if got := w.Implementation(uri, callPos); len(got) == 0 {
		t.Fatal("调用点也应能跳实现")
	}
}

func TestDocumentLinksOnImports(t *testing.T) {
	w, uri := navProject(t)
	links := w.DocumentLinks(uri)
	if len(links) != 1 {
		t.Fatalf("应只有一个 import 链接：%+v", links)
	}
	link := links[0]
	if !strings.HasSuffix(link.Target, "util.aila") {
		t.Fatalf("链接目标应是 util.aila：%s", link.Target)
	}
	// 区间应正好圈住 `_.sub.util`（import 在第 1 行，0 基 0）
	if link.Range.Start.Line != 0 || link.Range.End.Line != 0 {
		t.Fatalf("链接应在 import 那一行：%+v", link.Range)
	}
	text := "import _.sub.util;"
	if got := text[link.Range.Start.Character:link.Range.End.Character]; got != "_.sub.util" {
		t.Fatalf("链接区间应对准路径，实得 %q", got)
	}
}

func mustText(t *testing.T, w *Workspace, uri string) string {
	t.Helper()
	info := w.analysisFor(URI2Path(uri)).ByPath[URI2Path(uri)]
	if info == nil {
		t.Fatal("分析结果里没有该文件")
	}
	return info.Text
}
