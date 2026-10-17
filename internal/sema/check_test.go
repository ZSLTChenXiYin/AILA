package sema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/aila-lang/aila/internal/ast"
	"github.com/aila-lang/aila/internal/diag"
	"github.com/aila-lang/aila/internal/parser"
	"github.com/aila-lang/aila/internal/types"
)

type bindingsSnapshot struct {
	Files   map[string]map[string][][2]string `json:"files"`
	Skipped map[string][]string               `json:"skipped"`
}

// collectBindings 按源码顺序收集函数体里的 let/var 绑定类型。
func collectBindings(b *ast.Block, c *Checker, out *[][2]string) {
	if b == nil {
		return
	}
	for _, st := range b.Stmts {
		switch x := st.(type) {
		case *ast.LetDecl:
			t := c.Bindings[x]
			name := "?"
			if t != nil {
				name = types.TypeName(t)
			}
			*out = append(*out, [2]string{x.Name, name})
		case *ast.Block:
			collectBindings(x, c, out)
		case *ast.If:
			collectBindings(x.Then, c, out)
			for _, pair := range x.Elifs {
				if blk, ok := pair.B.(*ast.Block); ok {
					collectBindings(blk, c, out)
				}
			}
			collectBindings(x.Els, c, out)
		case *ast.ForC:
			collectBindings(x.Body, c, out)
		case *ast.ForIn:
			collectBindings(x.Body, c, out)
		case *ast.Region:
			collectBindings(x.Body, c, out)
		case *ast.Unsafe:
			collectBindings(x.Body, c, out)
		case *ast.Defer:
			collectBindings(x.Body, c, out)
		case *ast.Match:
			for _, arm := range x.Arms {
				if blk, ok := arm.Body.(*ast.Block); ok {
					collectBindings(blk, c, out)
				}
			}
		}
	}
}

// TestBindingsMatchReference 是 G3 的第三条判据：
// 同一份语料，Go 推断出的每个 let/var 绑定类型必须与 Python 参考实现一致。
func TestBindingsMatchReference(t *testing.T) {
	root := repoRoot(t)
	corpus := filepath.Join(root, "tests", "corpus")
	wantRaw, err := os.ReadFile(filepath.Join(corpus, "bindings.json"))
	if err != nil {
		t.Fatalf("读不到 bindings.json（先跑 python tools/gen_exprtypes.py）：%v", err)
	}
	var want bindingsSnapshot
	if err := json.Unmarshal(wantRaw, &want); err != nil {
		t.Fatalf("快照不是合法 JSON：%v", err)
	}
	if len(want.Files) == 0 {
		t.Fatal("快照为空")
	}
	compared := 0
	for file, funcs := range want.Files {
		src, err := os.ReadFile(filepath.Join(corpus, file))
		if err != nil {
			t.Fatalf("读不到语料 %s：%v", file, err)
		}
		mod, bag := parser.Parse(file, string(src))
		if bag.HasErrors() {
			t.Fatalf("%s 解析失败：%s", file, bag.Render(false))
		}
		table := BuildTable(mod, bag)
		sigs := BuildSigs(mod, table, bag)
		checker := Check(mod, table, sigs, bag)
		if bag.HasErrors() {
			t.Fatalf("%s 语义检查失败（Go 侧不应与参考实现分歧）：%s", file, bag.Render(false))
		}
		for fnName, wantBinds := range funcs {
			var fn *ast.FuncDecl
			for _, d := range mod.AllDecls() {
				decl, ok := d.(*ast.FuncDecl)
				if !ok || decl.Body == nil {
					continue
				}
				key := decl.Name
				if decl.OwnerType != "" {
					key = decl.OwnerType + "." + decl.Name
				}
				if key == fnName {
					fn = decl
					break
				}
			}
			if fn == nil {
				t.Fatalf("%s: 找不到函数 %s", file, fnName)
			}
			var got [][2]string
			collectBindings(fn.Body, checker, &got)
			if len(got) != len(wantBinds) {
				t.Fatalf("%s %s 绑定条数：got=%d want=%d\ngot=%v\nwant=%v",
					file, fnName, len(got), len(wantBinds), got, wantBinds)
			}
			for i := range wantBinds {
				if got[i][0] != wantBinds[i][0] || got[i][1] != wantBinds[i][1] {
					t.Fatalf("%s %s 第 %d 条绑定：got=%v want=%v", file, fnName, i, got[i], wantBinds[i])
				}
			}
			compared++
		}
	}
	if compared < 5 {
		t.Fatalf("只比对了 %d 个函数，太少", compared)
	}
}

// TestCheckReportsBasicErrors 覆盖几条基础诊断（不依赖快照）。
func TestCheckReportsBasicErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
		code string
	}{
		{"未定义名", "func main() { let x = nope; }", "E3001"},
		{"零隐式转换", "func main() { let x: I32 = 1; let y: F32 = x; }", "E3201"},
		{"只读绑定写入", "func main() { let x: I32 = 1; x = 2; }", "E4201"},
		{"match不完备", "enum E { A, B };\nfunc f(e: E) -> I32 { return match e { E.A => 1, }; }\nfunc main() { }", "E3401"},
		{"整数match缺兜底", "func f(x: I32) -> I32 { return match x { 0 => 1, }; }\nfunc main() { }", "E3403"},
		{"丢弃有值表达式", "func f() -> I32 { return 1; }\nfunc main() { f(); }", "E3206"},
		{"缺区域实参", "func f() @region[r] -> [U8] { return [0; 4]@nowhere; }\nfunc main() { }", "E4001"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bag := &diag.Bag{}
			mod, pbag := parser.Parse("t.aila", c.src)
			if pbag.HasErrors() {
				t.Fatalf("解析失败：%s", pbag.Render(false))
			}
			for _, d := range pbag.Items {
				bag.Items = append(bag.Items, d)
			}
			table := BuildTable(mod, bag)
			sigs := BuildSigs(mod, table, bag)
			Check(mod, table, sigs, bag)
			codes := map[string]bool{}
			for _, d := range bag.Items {
				codes[d.Code] = true
			}
			if !codes[c.code] {
				got := make([]string, 0, len(codes))
				for k := range codes {
					got = append(got, k)
				}
				t.Fatalf("期望 %s，实得 %v", c.code, got)
			}
		})
	}
}

var _ = strings.TrimSpace
var _ = runtime.Caller

// TestExternalPackageRootRejected 守护「非 `_` / `std` 根 = 外部包」这条边界：
// 依赖解析尚未实现时，必须显式报 E7010，不能静默当成「没有导入」（那只会变成
// 「未定义的名字」这种误导性诊断）。
func TestExternalPackageRootRejected(t *testing.T) {
	src := "import mylib.utils;\n\nfunc main() {\n    utils.greet();\n}\n"
	bag := &diag.Bag{}
	mod, pbag := parser.Parse("main.aila", src)
	bag.Items = append(bag.Items, pbag.Items...)
	mod.Name = "_.main"
	BuildProgram(map[string]*ast.Module{"_.main": mod}, []string{"_.main"}, "_.main", bag)
	codes := map[string]bool{}
	for _, d := range bag.Items {
		codes[d.Code] = true
	}
	if !codes["E7010"] {
		t.Fatalf("外部包根应报 E7010，实得 %v", codes)
	}
}