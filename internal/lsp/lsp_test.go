package lsp

import (
	"runtime"
	"strings"
	"testing"
)

func TestURIRoundTrip(t *testing.T) {
	// 用各平台自己的合法路径测（路径形态是平台语义，URI 函数本身两边都对）
	path := "/work/demo/main.aila"
	if runtime.GOOS == "windows" {
		path = `C:\work\demo\main.aila`
	}
	if got := URI2Path(Path2URI(path)); got != path {
		t.Fatalf("往返失败：%q", got)
	}
	if !strings.HasPrefix(Path2URI(path), "file:///") {
		t.Fatalf("URI 应带三个斜杠：%q", Path2URI(path))
	}
}

func TestIdentifierAt(t *testing.T) {
	text := "let value: I32 = other.value;\n"
	w, r, ok := identifierAt(text, Position{Line: 0, Character: 8})
	if !ok || w != "value" {
		t.Fatalf("取标识符失败：%q %v", w, ok)
	}
	if r.Start.Character != 4 || r.End.Character != 9 {
		t.Fatalf("区间不对：%+v", r)
	}
	if q := qualifierBefore(text, Position{Line: 0, Character: 26}); q != "other" {
		t.Fatalf("限定名解析失败：%q", q)
	}
}

func TestDiagnosticsAndSymbols(t *testing.T) {
	dir := t.TempDir()
	mustDir(t, dir+"/sub")
	mustWrite(t, dir+"/aila.toml", "[package]\nname = \"demo\"\nversion = \"0.1.0\"\nentry = \"_.main\"\n")
	mustWrite(t, dir+"/main.aila", "import _.sub.util;\n\nfunc main() {\n    let v: I32 = util.double(21);\n    _ = v;\n}\n")
	mustWrite(t, dir+"/sub/util.aila", "public func double(v: I32) -> I32 {\n    return v + v;\n}\n")

	w := NewWorkspace()
	w.SetRoot(dir)
	uri := Path2URI(dir + "/main.aila")

	diags := w.Diagnostics(uri)
	if len(diags) != 0 {
		t.Fatalf("干净工程不该有诊断：%+v", diags)
	}

	syms := w.DocumentSymbols(uri)
	flat := flatten(syms)
	if !strings.Contains(flat, "main") {
		t.Fatalf("大纲里应有 main：%s", flat)
	}

	// 跳转：`u.double` → sub/util.aila 的 double
	pos := Position{Line: 3, Character: 22} // double 的起点附近
	defs := w.Definition(uri, pos)
	if len(defs) == 0 {
		t.Fatalf("应能跳到跨模块的 double")
	}
	if !strings.HasSuffix(defs[0].URI, "util.aila") {
		t.Fatalf("跳转目标应是 util.aila，实得 %s", defs[0].URI)
	}

	// 悬停
	h := w.Hover(uri, pos)
	if h == nil || !strings.Contains(h.Contents.Value, "double") {
		t.Fatalf("悬停应显示 double：%+v", h)
	}

	// 引用：跨模块至少两处（声明 + 调用）
	refs := w.References(uri, pos)
	if len(refs) < 2 {
		t.Fatalf("引用数量不足：%d", len(refs))
	}

	// 工程符号
	if got := w.WorkspaceSymbols("doub"); len(got) == 0 {
		t.Fatalf("工程符号搜索应有结果")
	}
}

func TestDiagnosticsOnSyntaxError(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir+"/aila.toml", "[package]\nname = \"demo\"\nversion = \"0.1.0\"\nentry = \"_.main\"\n")
	mustWrite(t, dir+"/main.aila", "func main( {\n")
	w := NewWorkspace()
	w.SetRoot(dir)
	diags := w.Diagnostics(Path2URI(dir + "/main.aila"))
	if len(diags) == 0 {
		t.Fatal("语法错误应产生诊断")
	}
	if diags[0].Range.Start.Line < 0 {
		t.Fatalf("诊断区间应为 0 基：%+v", diags[0])
	}
	if diags[0].Severity != 1 {
		t.Fatalf("语法错误应是 error 级：%d", diags[0].Severity)
	}
}

func flatten(syms []DocumentSymbol) string {
	var b strings.Builder
	var walk func(list []DocumentSymbol)
	walk = func(list []DocumentSymbol) {
		for _, s := range list {
			b.WriteString(s.Name + " ")
			walk(s.Children)
		}
	}
	walk(syms)
	return b.String()
}
