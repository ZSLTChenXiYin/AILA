package lexer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// corpusDir 返回 tests/corpus 的绝对路径。
func corpusDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller 失败")
	}
	return filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(file))), "tests", "corpus")
}

// normalize 把 token 序列转成与 JSON 快照同构的形式（数字统一成 float64 便于比对）。
func normalize(toks []Token) []any {
	out := make([]any, 0, len(toks))
	for _, tk := range toks {
		var v any
		switch x := tk.Value.(type) {
		case int64:
			v = float64(x)
		case int:
			v = float64(x)
		case float64:
			v = x
		case string:
			v = x
		case nil:
			v = nil
		default:
			v = x
		}
		out = append(out, []any{string(tk.Kind), tk.Text, v})
	}
	return out
}

func TestCorpusTokens(t *testing.T) {
	dir := corpusDir(t)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("读不到语料目录 %s：%v", dir, err)
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
	for _, name := range files {
		name := name
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			wantRaw, err := os.ReadFile(filepath.Join(dir, strings.TrimSuffix(name, ".aila")+".tokens.json"))
			if err != nil {
				t.Fatalf("缺 token 快照（先跑 python tools/gen_corpus.py）：%v", err)
			}
			var want []any
			if err := json.Unmarshal(wantRaw, &want); err != nil {
				t.Fatalf("快照不是合法 JSON：%v", err)
			}
			toks, _ := Lex(name, string(src))
			got := normalize(toks)
			if !reflect.DeepEqual(got, want) {
				diff := firstDiff(got, want)
				t.Fatalf("%s 的 token 序列与快照不一致：\n%s", name, diff)
			}
		})
	}
}

func TestCorpusLexDiags(t *testing.T) {
	dir := corpusDir(t)
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".aila" {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ".aila")
		wantRaw, err := os.ReadFile(filepath.Join(dir, base+".lexdiags.json"))
		if err != nil {
			t.Fatalf("缺词法诊断快照：%v", err)
		}
		var want []string
		if err := json.Unmarshal(wantRaw, &want); err != nil {
			t.Fatal(err)
		}
		src, _ := os.ReadFile(filepath.Join(dir, e.Name()))
		_, bag := Lex(e.Name(), string(src))
		got := map[string]bool{}
		for _, d := range bag.Items {
			got[d.Code] = true
		}
		gotList := make([]string, 0, len(got))
		for code := range got {
			gotList = append(gotList, code)
		}
		sort.Strings(gotList)
		if len(want) == 0 && len(gotList) == 0 {
			continue
		}
		if !reflect.DeepEqual(gotList, want) {
			t.Fatalf("%s 的词法诊断码不一致：got=%v want=%v", e.Name(), gotList, want)
		}
	}
}

func firstDiff(got, want []any) string {
	n := len(got)
	if len(want) < n {
		n = len(want)
	}
	for i := 0; i < n; i++ {
		g, _ := json.Marshal(got[i])
		w, _ := json.Marshal(want[i])
		if string(g) != string(w) {
			return "第 " + itoa(i) + " 个 token：got=" + string(g) + " want=" + string(w)
		}
	}
	return "长度不同：got=" + itoa(len(got)) + " want=" + itoa(len(want))
}

func TestKeywordsAndWords(t *testing.T) {
	for _, kw := range []string{"func", "let", "return", "match", "out", "copy", "trunc"} {
		if !Keywords[kw] {
			t.Fatalf("%s 应是关键字", kw)
		}
	}
	for _, w := range []string{"I32", "Text", "chan", "send", "recv", "coro", "Unsafe"} {
		if Keywords[w] {
			t.Fatalf("%s 不应是关键字（是内置类型族名）", w)
		}
		if !TypeWords[w] {
			t.Fatalf("%s 应在 TypeWords 里", w)
		}
	}
}

func TestTokenHelpers(t *testing.T) {
	toks, bag := Lex("t.aila", "func f(a: I32) { let x = 1; }")
	if bag.HasErrors() {
		t.Fatalf("不应有错误：%s", bag.Render(false))
	}
	if !toks[0].IsKw("func") {
		t.Fatalf("首 token 应是 func：%v", toks[0])
	}
	if !toks[1].IsWord("f") && toks[1].Kind != KindIdent {
		t.Fatalf("第二个 token 应是 identifier：%v", toks[1])
	}
	if toks[len(toks)-1].Kind != KindEOF {
		t.Fatalf("末 token 应是 eof：%v", toks[len(toks)-1])
	}
}
