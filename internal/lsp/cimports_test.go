package lsp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// cimportProject 建一个带 cpack 的工程（`cpacks/mathkit`），返回工作区与入口 URI。
func cimportProject(t *testing.T, mainSrc string) (*Workspace, string) {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, dir+"/aila.toml",
		"[package]\nname = \"demo\"\nversion = \"0.1.0\"\nentry = \"_.main\"\n")
	mustWrite(t, dir+"/main.aila", mainSrc)
	mustWrite(t, dir+"/cpacks/mathkit/include/mathkit.h", strings.Join([]string{
		"#ifndef MATHKIT_H",
		"#define MATHKIT_H",
		"#define MATHKIT_MAX 1000",
		"int mathkit_add(int a, int b);",
		"int mathkit_clamp(int v, int lo, int hi);",
		"#endif",
		"",
	}, "\n"))
	mustWrite(t, dir+"/cpacks/mathkit/mathkit.c", strings.Join([]string{
		"#include \"mathkit.h\"",
		"int mathkit_add(int a, int b) { return a + b; }",
		"int mathkit_clamp(int v, int lo, int hi) { return v < lo ? lo : (v > hi ? hi : v); }",
		"",
	}, "\n"))
	w := NewWorkspace()
	w.SetRoot(dir)
	return w, Path2URI(dir + "/main.aila")
}

const cimportMain = `import ch _.cpacks.mathkit;

func main() {
    let s: I32 = mathkit.mathkit_add(20, 22);
    let c: I32 = mathkit.mathkit_clamp(99, 0, 100);
    let m: I32 = mathkit.MATHKIT_MAX;
    _ = s + c + m;
}
`

func TestCImportResolvesBindingModule(t *testing.T) {
	w, uri := cimportProject(t, cimportMain)
	info := w.analysisFor(URI2Path(uri)).ByPath[URI2Path(uri)]
	if info == nil {
		t.Fatal("入口文件应被分析")
	}
	diags := w.Diagnostics(uri)
	if len(diags) != 0 {
		t.Fatalf("C 导入工程不该有诊断：%+v", diags)
	}
	dep := w.analysisFor(URI2Path(uri)).Modules["_.cpacks.mathkit"]
	if dep == nil {
		t.Fatal("应把 cpack 绑定当模块索引（`_.cpacks.mathkit`）")
	}
	if !dep.Generated || dep.Cpack != "mathkit" {
		t.Fatalf("绑定模块应标记来源：%+v（%s）", dep, dep.Cpack)
	}
	if !strings.HasSuffix(dep.Path, ".ail") {
		t.Fatalf("绑定模块的落盘路径应是 .ail：%s", dep.Path)
	}
	if _, err := os.Stat(dep.Path); err != nil {
		t.Fatalf("生成的 .ail 应该真的写出来了：%v", err)
	}
	if !strings.Contains(dep.Text, "mathkit_add") {
		t.Fatalf("绑定文本应含 C 函数：%s", dep.Text[:min(120, len(dep.Text))])
	}
}

func TestCImportCompletionAndHover(t *testing.T) {
	w, uri := cimportProject(t, cimportMain)
	// `mathkit.` → 列出 C 函数与常量
	pos := posAtWord(t, w, uri, "mathkit.mathkit_clamp", "mathkit_clamp")
	line := lineContents(mustText(t, w, uri))[pos.Line]
	cursor := Position{Line: pos.Line, Character: strings.Index(line, "mathkit.") + len("mathkit") + 1}
	list := w.Completion(uri, cursor)
	labels := map[string]bool{}
	for _, it := range list.Items {
		labels[it.Label] = true
	}
	if !labels["mathkit_add"] || !labels["mathkit_clamp"] || !labels["MATHKIT_MAX"] {
		t.Fatalf("`mathkit.` 之后应补全出 C 符号，实得 %v", labels)
	}

	// 悬停：说清来自哪个 C 包
	h := w.Hover(uri, pos)
	if h == nil {
		t.Fatal("C 函数应可悬停")
	}
	if !strings.Contains(h.Contents.Value, "mathkit") {
		t.Fatalf("悬停应显示符号名：%s", h.Contents.Value)
	}
	if !strings.Contains(h.Contents.Value, "C 包") {
		t.Fatalf("悬停应说明来自 C 包：%s", h.Contents.Value)
	}
}

func TestCImportDefinitionAndLink(t *testing.T) {
	w, uri := cimportProject(t, cimportMain)
	pos := posAtWord(t, w, uri, "mathkit.mathkit_add", "mathkit_add")
	defs := w.Definition(uri, pos)
	if len(defs) == 0 {
		t.Fatal("C 函数应能跳到绑定的 `.ail`")
	}
	if !strings.HasSuffix(defs[0].URI, ".ail") {
		t.Fatalf("跳转目标应是生成的 .ail：%s", defs[0].URI)
	}

	// 签名帮助也能拿到 C 函数的签名
	line := lineContents(mustText(t, w, uri))[pos.Line]
	callPos := Position{Line: pos.Line, Character: strings.Index(line, "(") + 1}
	help := w.SignatureHelp(uri, callPos)
	if help == nil || len(help.Signatures) == 0 {
		t.Fatal("C 函数应有签名帮助")
	}
	if len(help.Signatures[0].Parameters) != 2 {
		t.Fatalf("mathkit_add 应有两个形参：%+v", help.Signatures[0].Parameters)
	}

	// import 路径可点击，指向绑定文件，并说明是 C 包
	links := w.DocumentLinks(uri)
	if len(links) != 1 {
		t.Fatalf("应有一个 import 链接：%+v", links)
	}
	if !strings.HasSuffix(links[0].Target, ".ail") {
		t.Fatalf("链接应指向绑定文件：%s", links[0].Target)
	}
	if !strings.Contains(links[0].Tooltip, "C 包") {
		t.Fatalf("链接提示应说明是 C 包：%s", links[0].Tooltip)
	}
}

func TestCImportMissingCpackReportsDiagnostic(t *testing.T) {
	w, uri := cimportProject(t, "import ch _.cpacks.nope;\n\nfunc main() {\n}\n")
	diags := w.Diagnostics(uri)
	if len(diags) == 0 {
		t.Fatal("cpack 目录不存在时应报诊断")
	}
	found := false
	for _, d := range diags {
		if d.Code == "E3001" && strings.Contains(d.Message, "nope") {
			found = true
		}
	}
	if !found {
		t.Fatalf("应报 E3001 并指出缺哪个 cpack：%+v", diags)
	}
	// 目标不存在不该凭空造出模块
	if dep := w.analysisFor(URI2Path(uri)).Modules["_.cpacks.nope"]; dep != nil {
		t.Fatalf("不存在的 cpack 不该有模块：%+v", dep)
	}
}

func TestCImportBindingIsGeneratedAndCached(t *testing.T) {
	w, uri := cimportProject(t, cimportMain)
	a := w.analysisFor(URI2Path(uri))
	dep := a.Modules["_.cpacks.mathkit"]
	if dep == nil {
		t.Fatal("应先解析出绑定模块")
	}
	// 进程内缓存：同一份头文件不该再生成第二份（键含内容指纹）
	key := filepath.Dir(URI2Path(uri)) + "|mathkit"
	cached, ok := w.cbind[key]
	if !ok || cached.Key == "" {
		t.Fatalf("应有进程内绑定缓存（键 %s）：%+v", key, w.cbind)
	}
	again := w.analysisFor(URI2Path(uri)).Modules["_.cpacks.mathkit"]
	if again == nil || again.Path != dep.Path {
		t.Fatalf("重复分析应命中同一份绑定：%+v", again)
	}
	// 磁盘缓存：key 文件与 .ail 都在
	dir := filepath.Dir(dep.Path)
	if _, err := os.Stat(filepath.Join(dir, "mathkit.key")); err != nil {
		t.Fatalf("应写 key 缓存文件：%v", err)
	}
	// 生成的绑定文件里，改名/格式化一律拒绝
	if got := w.PrepareRename(dep.URI, Position{Line: 3, Character: 20}); got != nil {
		t.Fatalf("生成的 .ail 不该支持改名：%+v", got)
	}
	if got := w.Format(dep.URI, FormattingOptions{TabSize: 4, InsertSpaces: true}); len(got) != 0 {
		t.Fatalf("生成的 .ail 不该被格式化：%+v", got)
	}
}
