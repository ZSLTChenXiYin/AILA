package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"aila/internal/ast"
)

func corpusDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller 失败")
	}
	return filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(file))), "tests", "corpus")
}

func corpusFiles(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(corpusDir(t))
	if err != nil {
		t.Fatalf("读不到语料目录：%v", err)
	}
	var files []string
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".aila" {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	if len(files) < 5 {
		t.Fatalf("语料太少：%v", files)
	}
	return files
}

// TestCorpusAST 是 G2 的核心判据：解析结果必须与 Python 参考实现的 AST 快照逐字符一致。
func TestCorpusAST(t *testing.T) {
	dir := corpusDir(t)
	checked := 0
	for _, name := range corpusFiles(t) {
		base := strings.TrimSuffix(name, ".aila")
		wantPath := filepath.Join(dir, base+".ast.txt")
		wantRaw, err := os.ReadFile(wantPath)
		if err != nil {
			continue // 该语料是「错误用例」，没有 AST 快照
		}
		src, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		mod, bag := Parse(name, string(src))
		if bag.HasErrors() {
			t.Fatalf("%s 不应有诊断：%s", name, bag.Render(false))
		}
		got := ast.Sexp(mod)
		want := strings.TrimRight(string(wantRaw), " \t\r\n")
		if got != want {
			t.Fatalf("%s 的 AST 与快照不一致：\n第一个差异点：%s", name, firstDiffAt(got, want))
		}
		checked++
	}
	if checked < 4 {
		t.Fatalf("只比对了 %d 份 AST 快照，太少了", checked)
	}
}

// TestCorpusParseDiags 锁定解析期诊断码（错误用例）。
func TestCorpusParseDiags(t *testing.T) {
	dir := corpusDir(t)
	for _, name := range corpusFiles(t) {
		base := strings.TrimSuffix(name, ".aila")
		wantRaw, err := os.ReadFile(filepath.Join(dir, base+".diags.json"))
		if err != nil {
			continue
		}
		var want []string
		if err := json.Unmarshal(wantRaw, &want); err != nil {
			t.Fatal(err)
		}
		src, _ := os.ReadFile(filepath.Join(dir, name))
		_, bag := Parse(name, string(src))
		seen := map[string]bool{}
		for _, d := range bag.Items {
			seen[d.Code] = true
		}
		got := make([]string, 0, len(seen))
		for code := range seen {
			got = append(got, code)
		}
		sort.Strings(got)
		sort.Strings(want)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("%s 的解析诊断码不一致：got=%v want=%v", name, got, want)
		}
	}
}

// TestSyntaxV2Parsing 锁定 designs/0004 §2/§3/§4 的 v2 语法解析形态：
// 层指令 / 层 API 调用 / 错误 enum / leading-dot 与类型模式。
func TestSyntaxV2Parsing(t *testing.T) {
	src := `#compile.c func ext_len(s: Text) -> USize;

#comprun.error enum Error {
    Nack, Timeout,
};

struct Dev { id: U8, };

enum Msg {
    Data { v: I32 },
    Eof,
};

module {
    func probe(d: Dev) -> U8 ! Error {
        return 0;
    }

    func main() {
        let n = #compile.sizeof(Dev);
        let buf: [U8; 4] = [0; 4];
        let l = #comprun.len(buf);
        let x = #runtime.as(I64, n);
        match probe(d0) {
            let v: U8 => { _ = v; },
            let e: Error => { _ = e; },
        }
        match m {
            .Data { v } => { _ = v; },
            .Eof => {},
            _ => {},
        }
    }
}
`
	mod, bag := Parse("v2.aila", src)
	if bag.HasErrors() {
		t.Fatalf("v2 语法不应有解析诊断：\n%s", bag.Render(false))
	}
	var extFn *ast.FuncDecl
	var errEnum *ast.EnumDecl
	var mainFn *ast.FuncDecl
	for _, d := range mod.Shell {
		if f, ok := d.(*ast.FuncDecl); ok {
			extFn = f
		}
		if e, ok := d.(*ast.EnumDecl); ok && e.Name == "Error" {
			errEnum = e
		}
	}
	for _, d := range mod.Block {
		if f, ok := d.(*ast.FuncDecl); ok && f.Name == "main" {
			mainFn = f
		}
	}
	// ① 壳内层指令 `#compile.c` → Attr{Name:"c", ViaLayer:true}
	if extFn == nil {
		t.Fatal("壳内没有解析出 ext_len 函数")
	}
	if len(extFn.Attrs.Attrs) != 1 || extFn.Attrs.Attrs[0].Name != "c" || !extFn.Attrs.Attrs[0].ViaLayer {
		t.Fatalf("层指令应落为 Attr{Name:c, ViaLayer:true}，实得 %+v", extFn.Attrs.Attrs)
	}
	// ② `#comprun.error enum Error` → EnumDecl.ErrLayer
	if errEnum == nil || errEnum.Name != "Error" || errEnum.ErrLayer != "comprun" {
		t.Fatalf("错误 enum 解析不符：%+v", errEnum)
	}
	if mainFn == nil {
		t.Fatal("没有解析出 main")
	}
	// ③ 层 API 调用节点（定向提取：main 体里的 let 初始化与 match 臂）
	var layerCalls []*ast.LayerCall
	var typePats []*ast.TypePattern
	var dotVariants, dotPaths int
	if mainFn.Body == nil {
		t.Fatal("main 没有函数体")
	}
	for _, st := range mainFn.Body.Stmts {
		switch x := st.(type) {
		case *ast.LetDecl:
			if lc, ok := x.Init.(*ast.LayerCall); ok {
				layerCalls = append(layerCalls, lc)
			}
		case *ast.Match:
			for _, arm := range x.Arms {
				switch pat := arm.Pattern.(type) {
				case *ast.TypePattern:
					typePats = append(typePats, pat)
				case *ast.VariantPattern:
					if len(pat.Path) > 0 && pat.Path[0] == "Data" {
						dotVariants++
					}
				case *ast.PathPattern:
					if len(pat.Path) == 1 && pat.Path[0] == "Eof" {
						dotPaths++
					}
				}
			}
		}
	}
	if len(layerCalls) != 3 {
		t.Fatalf("应有 3 个 LayerCall，实得 %d", len(layerCalls))
	}
	want := [][3]string{{"compile", "sizeof", ""}, {"comprun", "len", ""}, {"runtime", "as", ""}}
	for i, w := range want {
		if layerCalls[i].Layer != w[0] || layerCalls[i].Name != w[1] {
			t.Fatalf("LayerCall[%d] = %s.%s，期望 %s.%s", i, layerCalls[i].Layer, layerCalls[i].Name, w[0], w[1])
		}
	}
	if len(layerCalls[0].Args) != 1 || layerCalls[0].Args[0].Kind() != "TypeRef" {
		t.Fatalf("sizeof 参数应是类型，实得 %+v", layerCalls[0].Args)
	}
	if len(layerCalls[2].Args) != 2 || layerCalls[2].Args[0].Kind() != "TypeRef" {
		t.Fatalf("as 应是 (类型, 值) 两参，实得 %+v", layerCalls[2].Args)
	}
	// ④ 模式
	if len(typePats) != 2 || typePats[0].Name != "v" || typePats[1].Name != "e" {
		t.Fatalf("类型模式解析不符：%+v", typePats)
	}
	if dotVariants != 1 || dotPaths != 1 {
		t.Fatalf("leading-dot 模式解析不符：VariantPattern=%d PathPattern=%d", dotVariants, dotPaths)
	}
}

// TestSyntaxV2BadLayer 锁定 `#` 后非法层名的解析诊断。
func TestSyntaxV2BadLayer(t *testing.T) {
	src := "module {\n    func main() {\n        let n = #foo.sizeof(I32);\n    }\n}\n"
	_, bag := Parse("bad.aila", src)
	if !bag.HasErrors() {
		t.Fatal("非法层名应报诊断")
	}
}

// firstDiffAt 返回两个字符串的第一个差异位置及其上下文。
func firstDiffAt(got, want string) string {
	gr, wr := []rune(got), []rune(want)
	n := len(gr)
	if len(wr) < n {
		n = len(wr)
	}
	for i := 0; i < n; i++ {
		if gr[i] != wr[i] {
			return "位置 " + itoa(i) + "：\n  got  …" + sliceRunes(gr, i-40, i+60) +
				"\n  want …" + sliceRunes(wr, i-40, i+60)
		}
	}
	return "长度不同：got=" + itoa(len(gr)) + " want=" + itoa(len(wr)) +
		"\n  got 尾部：" + sliceRunes(gr, n-40, n) + "\n  want 尾部：" + sliceRunes(wr, n-40, n)
}

func sliceRunes(r []rune, from, to int) string {
	if from < 0 {
		from = 0
	}
	if to > len(r) {
		to = len(r)
	}
	if from > to {
		return ""
	}
	return string(r[from:to])
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf []byte
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		return "-" + string(buf)
	}
	return string(buf)
}
