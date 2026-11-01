package codegen

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"aila/internal/parser"
	"aila/internal/sema"
	"aila/internal/types"
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
	lw := New(mod, "_.main", table, sigs, check, layout, hostTriple, "_.main", false)
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

// TestGlobalArrayInit 守护模块级数据表的全局初值：
// 定长数组字面量必须编进 `global` 的聚合初值，而不是退化成 zeroinitializer
// （字库/命令表这类表被清零时，程序跑起来是「不报错但全空」——最难查的那种）。
func TestGlobalArrayInit(t *testing.T) {
	src := `let TABLE: [U8; 4] = [1, 2, 3, 4];

func pick(i: USize) -> U8 {
    return TABLE[i];
}

func main() {
    _ = pick(0);
}
`
	mod, bag := parser.Parse("t.aila", src)
	if bag.HasErrors() {
		t.Fatalf("解析失败：%s", bag.Render(false))
	}
	table := sema.BuildTable(mod, bag)
	sigs := sema.BuildSigs(mod, table, bag)
	check := sema.Check(mod, table, sigs, bag)
	if bag.HasErrors() {
		t.Fatalf("语义检查失败：%s", bag.Render(false))
	}
	ir := New(mod, "_.main", table, sigs, check, types.NewLayout(8), hostTriple, "_.main", false).Generate()
	if !strings.Contains(ir, "[i8 1, i8 2, i8 3, i8 4]") {
		t.Fatalf("定长数组的初值没进 global：\n%s", ir)
	}
	if strings.Contains(ir, "zeroinitializer") && strings.Contains(ir, "TABLE") {
		t.Fatalf("全局数组退化成 zeroinitializer：\n%s", ir)
	}
}

// TestPtrSizeFollowsTarget 守护目标指针宽度：xtensa（32 位）上 USize 必须是 i32。
// New() 曾把 ptrSize 写死 8，嵌入式目标的 USize/指针全按 64 位出 IR（ABI 直接错）。
func TestPtrSizeFollowsTarget(t *testing.T) {
	src := `func take(n: USize) -> USize {
    return n;
}

func main() {
    _ = take(1);
}
`
	mod, bag := parser.Parse("t.aila", src)
	if bag.HasErrors() {
		t.Fatalf("解析失败：%s", bag.Render(false))
	}
	table := sema.BuildTable(mod, bag)
	sigs := sema.BuildSigs(mod, table, bag)
	check := sema.Check(mod, table, sigs, bag)
	if bag.HasErrors() {
		t.Fatalf("语义检查失败：%s", bag.Render(false))
	}
	// 32 位目标：布局 ptrSize=4 → USize = i32
	ir32 := New(mod, "_.main", table, sigs, check, types.NewLayout(4),
		"xtensa-esp8266-elf", "_.main", false).Generate()
	if !strings.Contains(ir32, "define i32 @aila_main__take(i32 %p_n)") {
		t.Fatalf("32 位目标的 USize 应该是 i32：\n%s", ir32)
	}
	// 64 位宿主：仍是 i64（不能为了嵌入式把宿主一起改坏）
	ir64 := New(mod, "_.main", table, sigs, check, types.NewLayout(8),
		hostTriple, "_.main", false).Generate()
	if !strings.Contains(ir64, "define i64 @aila_main__take(i64 %p_n)") {
		t.Fatalf("64 位宿主的 USize 应该是 i64：\n%s", ir64)
	}
}
