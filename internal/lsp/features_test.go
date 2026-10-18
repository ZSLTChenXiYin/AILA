package lsp

import (
	"strings"
	"testing"
)

// featureProject 建一个带跨模块调用的小工程，返回工作区与入口 URI。
func featureProject(t *testing.T) (*Workspace, string) {
	t.Helper()
	dir := t.TempDir()
	mustDir(t, dir+"/sub")
	mustWrite(t, dir+"/aila.toml",
		"[package]\nname = \"demo\"\nversion = \"0.1.0\"\nentry = \"_.main\"\n\n[target]\nboard = \"host\"\n")
	main := strings.Join([]string{
		"import _.sub.util as u;",
		"",
		"public struct Point {",
		"    x: I32,",
		"    y: I32,",
		"};",
		"",
		"public enum Color { Red, Green };",
		"",
		"// 注释一",
		"// 注释二",
		"module {",
		"    func main() {",
		"        let p: Point = Point { x: 1, y: 2 };",
		"        let n = u.double(p.x);",
		"        let m = u.add(n, 1);",
		"        _ = m;",
		"    }",
		"}",
		"",
	}, "\n")
	mustWrite(t, dir+"/main.aila", main)
	mustWrite(t, dir+"/sub/util.aila", strings.Join([]string{
		"public func double(v: I32) -> I32 {",
		"    return v + v;",
		"}",
		"",
		"public func add(a: I32, b: I32) -> I32 {",
		"    return a + b;",
		"}",
		"",
	}, "\n"))

	w := NewWorkspace()
	w.SetRoot(dir)
	return w, Path2URI(dir + "/main.aila")
}

// posOfLine 找到包含 needle 的行，并把光标放在 needle 之后第 after 个字符处。
func posOfLine(t *testing.T, w *Workspace, uri, needle string, after int) Position {
	t.Helper()
	text := w.analysisFor(URI2Path(uri)).ByPath[URI2Path(uri)].Text
	for i, line := range lineContents(text) {
		if idx := strings.Index(line, needle); idx >= 0 {
			return Position{Line: i, Character: idx + len(needle) + after}
		}
	}
	t.Fatalf("找不到 %q", needle)
	return Position{}
}

// posAtWord 找到包含 needle 的行，并把光标放在该行 word 的中间（更直观的写法）。
func posAtWord(t *testing.T, w *Workspace, uri, needle, word string) Position {
	t.Helper()
	text := w.analysisFor(URI2Path(uri)).ByPath[URI2Path(uri)].Text
	for i, line := range lineContents(text) {
		if strings.Contains(line, needle) {
			idx := strings.Index(line, word)
			if idx < 0 {
				t.Fatalf("行 %q 里找不到 %q", line, word)
			}
			return Position{Line: i, Character: idx + len(word)/2}
		}
	}
	t.Fatalf("找不到包含 %q 的行", needle)
	return Position{}
}

func TestCompletionOnLocalField(t *testing.T) {
	w, uri := featureProject(t)
	// `p.x` 里把光标放在点之后
	pos := posOfLine(t, w, uri, "u.double(p.", 0)
	list := w.Completion(uri, pos)
	labels := map[string]bool{}
	for _, it := range list.Items {
		labels[it.Label] = true
	}
	if !labels["x"] || !labels["y"] {
		t.Fatalf("`p.` 之后应补全出字段 x/y，实得 %v", labels)
	}
}

func TestCompletionOnImportAlias(t *testing.T) {
	w, uri := featureProject(t)
	pos := posOfLine(t, w, uri, "u.", 0)
	list := w.Completion(uri, pos)
	labels := map[string]bool{}
	for _, it := range list.Items {
		labels[it.Label] = true
	}
	if !labels["double"] || !labels["add"] {
		t.Fatalf("`u.` 之后应补全出 double/add，实得 %v", labels)
	}
}

func TestCompletionKeywordAndLocals(t *testing.T) {
	w, uri := featureProject(t)
	pos := posOfLine(t, w, uri, "let m = ", 0)
	list := w.Completion(uri, pos)
	labels := map[string]bool{}
	for _, it := range list.Items {
		labels[it.Label] = true
	}
	if !labels["n"] || !labels["p"] {
		t.Fatalf("应补全出可见局部变量，实得 %v", labels)
	}
	if !labels["if"] || !labels["func"] {
		t.Fatalf("应补全出关键字，实得 %v", labels)
	}
	if !labels["I32"] {
		t.Fatalf("应补全出内置类型，实得 %v", labels)
	}
}

func TestSignatureHelp(t *testing.T) {
	w, uri := featureProject(t)
	pos := posOfLine(t, w, uri, "u.double(", 0)
	help := w.SignatureHelp(uri, pos)
	if help == nil {
		t.Fatal("应给出签名帮助")
	}
	if !strings.Contains(help.Signatures[0].Label, "double") || !strings.Contains(help.Signatures[0].Label, "I32") {
		t.Fatalf("签名不对：%s", help.Signatures[0].Label)
	}
	if help.ActiveParameter != 0 {
		t.Fatalf("第一个实参，activeParameter 应为 0，实得 %d", help.ActiveParameter)
	}
	// 第二个实参
	pos2 := posOfLine(t, w, uri, "u.add(n, ", 0)
	help2 := w.SignatureHelp(uri, pos2)
	if help2 == nil || help2.ActiveParameter != 1 {
		t.Fatalf("逗号之后 activeParameter 应为 1：%+v", help2)
	}
	if len(help2.Signatures[0].Parameters) != 2 {
		t.Fatalf("add 应有两个形参：%+v", help2.Signatures[0].Parameters)
	}
}

func TestPrepareRenameAndRename(t *testing.T) {
	w, uri := featureProject(t)
	// 函数名跨模块改名（光标落在 `double` 中间）
	pos := posOfLine(t, w, uri, "u.double", -3)
	if p := w.PrepareRename(uri, pos); p == nil || p.Placeholder != "double" {
		t.Fatalf("函数名应可改名：%+v", p)
	}
	edit := w.Rename(uri, pos, "twice")
	if edit == nil {
		t.Fatal("应产生重命名编辑")
	}
	total := 0
	var utilEdits int
	for u, edits := range edit.Changes {
		total += len(edits)
		if strings.HasSuffix(u, "util.aila") {
			utilEdits = len(edits)
		}
	}
	if utilEdits != 1 {
		t.Fatalf("util.aila 里应改 1 处（声明），实得 %d（总 %d）", utilEdits, total)
	}
	if total < 2 {
		t.Fatalf("应至少改 2 处（声明 + 调用），实得 %d", total)
	}

	// 关键字不许改名
	kw := posOfLine(t, w, uri, "let p", -2) // 落在 `let` 上
	if p := w.PrepareRename(uri, kw); p != nil {
		t.Fatalf("关键字不该可改名：%+v", p)
	}

	// 局部变量只在本文件改名
	lpos := posOfLine(t, w, uri, "let n", 4)
	ledit := w.Rename(uri, lpos, "count")
	if ledit == nil {
		t.Fatal("局部变量应可改名")
	}
	for u := range ledit.Changes {
		if strings.HasSuffix(u, "util.aila") {
			t.Fatal("局部变量不该改到别的模块")
		}
	}
}

func TestFoldingRanges(t *testing.T) {
	w, uri := featureProject(t)
	folds := w.FoldingRanges(uri)
	hasKind := func(kind string) bool {
		for _, f := range folds {
			if f.Kind == kind {
				return true
			}
		}
		return false
	}
	if !hasKind("region") {
		t.Fatalf("应有代码块折叠：%+v", folds)
	}
	if !hasKind("comment") {
		t.Fatalf("应有注释折叠：%+v", folds)
	}
	for _, f := range folds {
		if f.EndLine <= f.StartLine {
			t.Fatalf("折叠区间必须跨行：%+v", f)
		}
	}
}

func TestSelectionRanges(t *testing.T) {
	w, uri := featureProject(t)
	pos := posOfLine(t, w, uri, "u.double(", 0)
	rs := w.SelectionRanges(uri, []Position{pos})
	if len(rs) != 1 {
		t.Fatalf("应返回一个选区链：%+v", rs)
	}
	depth := 0
	for n := &rs[0]; n != nil; n = n.Parent {
		depth++
	}
	if depth < 3 {
		t.Fatalf("选区链至少三层（标识符 → 括号 → 行 → 文件），实得 %d", depth)
	}
	if rs[0].Range.Start.Character == 0 && rs[0].Range.End.Line == rs[0].Range.Start.Line && rs[0].Range.End.Character == 0 {
		t.Fatal("最内层选区不该是空区间")
	}
}

func TestFormatDocument(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir+"/aila.toml",
		"[package]\nname = \"demo\"\nversion = \"0.1.0\"\nentry = \"_.main\"\n\n[target]\nboard = \"host\"\n")
	messy := "module {\n   func main() {\n         let x: I32 = 1;   \n  }\n}\n\n\n"
	mustWrite(t, dir+"/main.aila", messy)
	w := NewWorkspace()
	w.SetRoot(dir)
	uri := Path2URI(dir + "/main.aila")
	edits := w.Format(uri, FormattingOptions{TabSize: 4, InsertSpaces: true})
	if len(edits) != 1 {
		t.Fatalf("应返回一个整文档替换：%+v", edits)
	}
	got := edits[0].NewText
	want := "module {\n    func main() {\n        let x: I32 = 1;\n    }\n}\n"
	if got != want {
		t.Fatalf("格式化结果不对：\n%q\n期望\n%q", got, want)
	}
	if again := formatText(got, 4); again != got {
		t.Fatalf("格式化必须幂等：\n%q\nvs\n%q", again, got)
	}
}

func TestInlayHints(t *testing.T) {
	w, uri := featureProject(t)
	hints := w.InlayHints(uri, Range{Start: Position{}, End: Position{Line: 1 << 20, Character: 0}})
	if len(hints) == 0 {
		t.Fatal("`let n = ...` 应有推断类型提示")
	}
	found := false
	for _, h := range hints {
		if strings.Contains(h.Label, "I32") {
			found = true
		}
	}
	if !found {
		t.Fatalf("提示应是 I32：%+v", hints)
	}
	// 显式标注的绑定不出提示
	for _, h := range hints {
		if h.Position.Line == 13 { // `let p: Point = ...` 那一行
			t.Fatalf("显式标注不该出内联提示：%+v", h)
		}
	}
}

func TestScanCodeMasksLiterals(t *testing.T) {
	code, inside := scanCode("let s: Text = \"a{b\"; // }\nlet c: C32 = '{';\n")
	if strings.Count(code, "{") != 0 || strings.Count(code, "}") != 0 {
		t.Fatalf("字面量与注释里的花括号应被抹掉：%q", code)
	}
	if len(inside) != 3 { // 末尾换行会产生一个空行
		t.Fatalf("行数不对：%d", len(inside))
	}
	for i, in := range inside {
		if in {
			t.Fatalf("第 %d 行不该被标记为「行首在字面量里」", i)
		}
	}
	// 多行原始字符串里的行要被标记
	_, insideRaw := scanCode("let s: Text = r\"line1\nline2\";\n")
	if len(insideRaw) < 2 || !insideRaw[1] {
		t.Fatalf("原始字符串的续行应被标记：%+v", insideRaw)
	}
}
