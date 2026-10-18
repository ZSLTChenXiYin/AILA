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
