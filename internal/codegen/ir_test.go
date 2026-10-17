package codegen

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ailang/aila/internal/parser"
	"github.com/ailang/aila/internal/sema"
	"github.com/ailang/aila/internal/types"
)

// irSubset 与 tools/goldengen 的 irCorpus 保持一致：只比对 Go 已实现的语料。
var irSubset = []string{"ir_basic.aila", "ir_wide.aila"}

const hostTriple = "x86_64-w64-windows-gnu"

// repoRoot 从本文件定位仓库根（internal/codegen → ../..）。
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("无法定位测试文件")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// buildIR 走与 Python 参考实现相同的流水线，产出单模块 IR 文本。
func buildIR(t *testing.T, name string) string {
	t.Helper()
	root := repoRoot(t)
	path := filepath.Join(root, "tests", "corpus", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读不到语料 %s：%v", name, err)
	}
	mod, bag := parser.Parse(name, string(data))
	if bag.HasErrors() {
		t.Fatalf("%s 解析失败：\n%s", name, bag.Render(false))
	}
	table := sema.BuildTable(mod, bag)
	sigs := sema.BuildSigs(mod, table, bag)
	check := sema.Check(mod, table, sigs, bag)
	if bag.HasErrors() {
		t.Fatalf("%s 语义检查失败：\n%s", name, bag.Render(false))
	}
	layout := types.NewLayout(8)
	lw := New(mod, "_.main", table, sigs, check, layout, hostTriple, "_.main")
	return lw.Generate()
}

// TestIRParity 逐字节比对 Go 生成的 IR 与 Python 参考快照。
func TestIRParity(t *testing.T) {
	root := repoRoot(t)
	for _, name := range irSubset {
		t.Run(name, func(t *testing.T) {
			got := buildIR(t, name)
			snapPath := filepath.Join(root, "tests", "corpus",
				strings.TrimSuffix(name, ".aila")+".ir")
			want, err := os.ReadFile(snapPath)
			if err != nil {
				t.Fatalf("读不到快照 %s：%v", snapPath, err)
			}
			wantText := strings.TrimRight(string(want), " \t\r\n")
			gotText := strings.TrimRight(got, " \t\r\n")
			if gotText == wantText {
				return
			}
			// 逐行定位首个差异，方便排查
			gl := strings.Split(gotText, "\n")
			wl := strings.Split(wantText, "\n")
			for i := 0; i < len(gl) || i < len(wl); i++ {
				var g, w string
				if i < len(gl) {
					g = gl[i]
				}
				if i < len(wl) {
					w = wl[i]
				}
				if g != w {
					t.Fatalf("IR 与快照第 %d 行不一致：\n  期望: %q\n  实得: %q\n(共 %d 行 / 快照 %d 行)",
						i+1, w, g, len(gl), len(wl))
				}
			}
			t.Fatalf("IR 与快照长度不同：实得 %d 行，快照 %d 行", len(gl), len(wl))
		})
	}
}
