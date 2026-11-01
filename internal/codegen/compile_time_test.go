package codegen

import (
	"strings"
	"testing"

	"aila/internal/diag"
	"aila/internal/parser"
	"aila/internal/sema"
	"aila/internal/types"
)

// compileInline 走与 buildIR 相同的流水线，但源码内联（designs/0006 §1 的剪枝用例）。
func compileInline(t *testing.T, src string) (string, *diag.Bag) {
	t.Helper()
	mod, bag := parser.Parse("inline.aila", src)
	if bag.HasErrors() {
		return "", bag
	}
	table := sema.BuildTable(mod, bag)
	sigs := sema.BuildSigs(mod, table, bag)
	check := sema.Check(mod, table, sigs, bag)
	if bag.HasErrors() {
		return "", bag
	}
	layout := types.NewLayout(8)
	lw := New(mod, "_.main", table, sigs, check, layout, hostTriple, "_.main", false)
	return lw.Generate(), bag
}

const prunePrelude = `
const MODE: U8 = 2;

func path_low() {}
func path_high() {}
`

// count 数 IR 里某名字出现的次数：被剪的分支只剩「函数定义」一次，未剪则是
// 「定义 + 调用」两次。
func count(ir, name string) int { return strings.Count(ir, name) }

// TestCompileIfPrunes 守护 `#compile.if`：条件编译期已知时，未选中分支不产生任何指令
//（连它引用的字符串常量都不应出现在 IR 里）；这正是不依赖跨模块内联也能剪枝的关键。
func TestCompileIfPrunes(t *testing.T) {
	ir, bag := compileInline(t, prunePrelude+`
func main() {
	#compile.if MODE == 2 { path_high(); } else { path_low(); }
}
`)
	if bag.HasErrors() {
		t.Fatalf("不该报错：\n%s", bag.Render(false))
	}
	if count(ir, "path_high") < 2 {
		t.Fatalf("选中分支应当有调用：\n%s", ir)
	}
	if count(ir, "path_low") != 1 {
		t.Fatalf("未选中分支不该产生调用（path_low 出现 %d 次）：\n%s", count(ir, "path_low"), ir)
	}
}

// TestCompileIfLetConstPropagation 守护「函数内 let 绑到 const-expr 也算编译期已知」。
func TestCompileIfLetConstPropagation(t *testing.T) {
	ir, bag := compileInline(t, prunePrelude+`
func main() {
	let fast: Bool = (MODE == 2);
	#compile.if fast { path_high(); } else { path_low(); }
}
`)
	if bag.HasErrors() {
		t.Fatalf("不该报错：\n%s", bag.Render(false))
	}
	if count(ir, "path_high") < 2 || count(ir, "path_low") != 1 {
		t.Fatalf("let 常量传播没生效：\n%s", ir)
	}
}

// TestCompileIfRejectsRuntimeCond 守护口径：条件求不出编译期常量必须报错，绝不静默降级成
// 运行时 if（designs/0006 §1.3 规则 1）。
func TestCompileIfRejectsRuntimeCond(t *testing.T) {
	_, bag := compileInline(t, prunePrelude+`
func main() {
	#compile.if MODE == 2 { path_high(); } else { path_low(); }
	pick(MODE);
}

func pick(n: U8) {
	#compile.if n == 2 { path_high(); } else { path_low(); }
}
`)
	if !bag.HasErrors() {
		t.Fatal("形参作为条件应当报错")
	}
	if !strings.Contains(bag.Render(false), "E3210") {
		t.Fatalf("应报 E3210，实得：\n%s", bag.Render(false))
	}
}

// TestCompileMatchPrunes 守护 `#compile.match`：命中之外的分支不生成；且编译期已知时
// 不要求穷尽（这里故意省略 `0`）。
func TestCompileMatchPrunes(t *testing.T) {
	ir, bag := compileInline(t, prunePrelude+`
func main() {
	#compile.match MODE {
		1 => { path_low(); },
		2 => { path_high(); },
		_ => {},
	}
}
`)
	if bag.HasErrors() {
		t.Fatalf("不该报错：\n%s", bag.Render(false))
	}
	if count(ir, "path_high") < 2 || count(ir, "path_low") != 1 {
		t.Fatalf("match 剪枝没生效：\n%s", ir)
	}
}

// TestCompileMatchNoArmHits 守护：判别值已知却没有分支命中是编译错误（而不是静默跳过）。
func TestCompileMatchNoArmHits(t *testing.T) {
	_, bag := compileInline(t, prunePrelude+`
func main() {
	#compile.match MODE {
		1 => { path_low(); },
		3 => { path_high(); },
	}
}
`)
	if !bag.HasErrors() {
		t.Fatal("判别值无分支命中应当报错")
	}
	if !strings.Contains(bag.Render(false), "E3212") {
		t.Fatalf("应报 E3212，实得：\n%s", bag.Render(false))
	}
}

// TestCompileIfDeadBranchStillTypeChecked 守护安全性：死分支仍完整类型检查
//（C 的 `#if` 做不到这一点）。
func TestCompileIfDeadBranchStillTypeChecked(t *testing.T) {
	_, bag := compileInline(t, prunePrelude+`
func main() {
	#compile.if MODE == 2 { path_high(); } else { no_such_function(); }
}
`)
	if !bag.HasErrors() {
		t.Fatal("死分支里的未定义名字必须报错")
	}
}

// TestCompileMatchConstHeadNoNullLoad 守护一个真踩过的坑：`#compile.match` 的判别值是
// **常量**时，常量不是主名字，lowerPlace 给不出地址——旧 codegen 因此生成
// `load i8, ptr null`，宿主上段错误、裸机上 `Fatal exception (28) LoadProhibited` 死循环
//（designs/0006 §4.1 第 7 条）。断言：判别值按值求值，IR 里不许出现空指针读。
func TestCompileMatchConstHeadNoNullLoad(t *testing.T) {
	ir, bag := compileInline(t, prunePrelude+`
func main() {
	#compile.match MODE {
		1 => { path_low(); },
		_ => { path_high(); },
	}
}
`)
	if bag.HasErrors() {
		t.Fatalf("不该报错：\n%s", bag.Render(false))
	}
	if strings.Contains(ir, "ptr null") {
		t.Fatalf("判别值是常量时不该读空指针（旧 bug 的指纹）：\n%s", ir)
	}
	if !strings.Contains(ir, "store i8 2,") {
		t.Fatalf("判别值应当按值落槽（MODE = 2）：\n%s", ir)
	}
}

// TestPruneKeepsABI 是 designs/0006 §3 断言 6 的落地：**剪枝不改模块接口 / ABI**。
//
// 判据：同一份源码只改 const 取值（判定的那一支换了），函数签名集合必须逐字相同，
// 而函数体的 IR 必须不同（说明确实剪了不同的分支）。签名取自 IR 的 `define` / `declare`
// 头部——模块对外的调用约定就是这些行，其余部分变了才叫「ABI 变了」。
func TestPruneKeepsABI(t *testing.T) {
	program := func(sel string) string {
		return `
const SEL: U8 = ` + sel + `;

func drv_a() -> U16 { return 11; }
func drv_b() -> U16 { return 22; }

func main() {
	#compile.match SEL {
		1 => { _ = drv_a(); },
		_ => { _ = drv_b(); },
	}
}
`
	}
	sigs := func(src string) (map[string]string, string) {
		t.Helper()
		ir, bag := compileInline(t, src)
		if bag.HasErrors() {
			t.Fatalf("不该报错：\n%s", bag.Render(false))
		}
		out := map[string]string{}
		for _, line := range strings.Split(ir, "\n") {
			if strings.HasPrefix(line, "define ") || strings.HasPrefix(line, "declare ") {
				head := strings.TrimSuffix(strings.TrimSpace(line), " {")
				name := ""
				if i := strings.Index(head, "@"); i >= 0 {
					name = head[i:]
					if j := strings.Index(name, "("); j >= 0 {
						name = name[:j]
					}
				}
				if name != "" {
					out[name] = head
				}
			}
		}
		return out, ir
	}
	a, irA := sigs(program("1"))
	b, irB := sigs(program("2"))
	if len(a) == 0 || len(b) == 0 {
		t.Fatal("没抽到任何签名")
	}
	for name, headA := range a {
		headB, ok := b[name]
		if !ok {
			t.Fatalf("剪枝换了取值后少了符号 %s（ABI 变了）", name)
		}
		if headA != headB {
			t.Fatalf("剪枝换了取值后 %s 的签名变了：\n  %s\n  %s", name, headA, headB)
		}
	}
	for name := range b {
		if _, ok := a[name]; !ok {
			t.Fatalf("剪枝换了取值后多了符号 %s（ABI 变了）", name)
		}
	}
	if irA == irB {
		t.Fatal("换了 const 取值后函数体应当不同（剪枝没生效）")
	}
}
