package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMergeModules 守护 `--lto` 的文本合并：整程序合成一个 LLVM 模块时有四件事必须做对，
// 少一件后端就报重复定义（这几条全是本轮踩出来的）。
func TestMergeModules(t *testing.T) {
	modA := `; target: x86_64-w64-windows-gnu
target datalayout = "e-m:w-p270:32:32"
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [3 x i8] c"ab\00"
declare void @aila_rt_trap(ptr, i64, ptr, i64)

define i32 @aila_a__get() noinline optnone {
entry:
  call void @aila_rt_trap(ptr @.str.0, i64 2, ptr @.str.0, i64 2)
  ret i32 1
}
attributes #0 = { presplitcoroutine }
`
	modB := `; target: x86_64-w64-windows-gnu
target datalayout = "e-m:w-p270:32:32"
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [3 x i8] c"cd\00"
declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare i32 @aila_a__get()

define i32 @aila_b__main() {
entry:
  %v = call i32 @aila_a__get()
  ret i32 %v
}
attributes #0 = { "frame-pointer"="none" }
`
	out := mergeModules([][2]string{{"_.a", modA}, {"_.b", modB}})
	if n := strings.Count(out, "target triple ="); n != 1 {
		t.Fatalf("target triple 只该留一份，实得 %d：\n%s", n, out)
	}
	if n := strings.Count(out, "target datalayout ="); n != 1 {
		t.Fatalf("target datalayout 只该留一份，实得 %d", n)
	}
	// declare 去重：同一个 aila_rt_trap 声明两次 + define-then-declare 的 aila_a__get
	if n := strings.Count(out, "declare void @aila_rt_trap"); n != 1 {
		t.Fatalf("重复的 declare 必须去掉，实得 %d 份：\n%s", n, out)
	}
	if strings.Contains(out, "declare i32 @aila_a__get") {
		t.Fatalf("已有定义的函数不该再留 declare：\n%s", out)
	}
	// 字符串常量池按模块重编号
	if !strings.Contains(out, "@.str.0.0") || !strings.Contains(out, "@.str.1.0") {
		t.Fatalf("字符串常量符号没按模块重编号：\n%s", out)
	}
	// attribute group 重编号：两个模块各有一个 #0，合并后必须是 #0 + #1
	if n := strings.Count(out, "attributes #0 ="); n != 1 {
		t.Fatalf("attribute group 撞号了（attributes #0 出现 %d 次）：\n%s", n, out)
	}
	if !strings.Contains(out, "attributes #1 =") {
		t.Fatalf("第二个模块的 attribute group 没被重编号：\n%s", out)
	}
	// 热替换属性必须原样保住
	if !strings.Contains(out, "noinline optnone") {
		t.Fatalf("hot 模块的属性在合并后丢了：\n%s", out)
	}
}

// TestHotModuleMarksNoInline 守护「铁律 7 的边界只落在 module hot 上」：
// hot 模块的每个函数都必须带 noinline + optnone（替换点可寻址 + 调用点形状不变），
// 非 hot 模块不带——否则整程序视图下的跨模块内联就没了（designs/0001 §12.1）。
func TestHotModuleMarksNoInline(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "hotmod"), 0o777); err != nil {
		t.Fatal(err)
	}
	write := func(rel, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, rel), []byte(body), 0o666); err != nil {
			t.Fatal(err)
		}
	}
	write("aila.toml", "[package]\nname = \"hotdemo\"\nversion = \"0.1.0\"\nentry = \"_.main\"\n\n[build]\nopt = 0\n")
	write(filepath.Join("hotmod", "hotmod.aila"), `public func ping() -> U16;

module hot {
    public func ping() -> U16 {
        return 7;
    }
}
`)
	write("main.aila", `import std.io;
import _.hotmod.hotmod;

module {
    func plain() -> U16 {
        return 1;
    }

    func main() {
        io.printf("%u %u\n",
            io.FmtVal.U { v: #compile.as(U64, hotmod.ping()) },
            io.FmtVal.U { v: #compile.as(U64, plain()) });
    }
}
`)
	res := Load(root, Options{})
	if res.Diags.HasErrors() {
		t.Fatalf("装载失败：\n%s", res.Diags.Render(false))
	}
	irs := res.GenerateIR()
	if res.Diags.HasErrors() {
		t.Fatalf("代码生成失败：\n%s", res.Diags.Render(false))
	}
	hotIR, ok := irs["_.hotmod.hotmod"]
	if !ok {
		t.Fatalf("没有 hot 模块的 IR：%v", keysOf(irs))
	}
	if !strings.Contains(hotIR, "noinline optnone") {
		t.Fatalf("hot 模块的函数必须打 noinline + optnone：\n%s", hotIR)
	}
	mainIR := irs["_.main"]
	for _, line := range strings.Split(mainIR, "\n") {
		if strings.HasPrefix(line, "define ") && strings.Contains(line, "plain") {
			if strings.Contains(line, "noinline") || strings.Contains(line, "optnone") {
				t.Fatalf("非 hot 模块的函数不该被打上 noinline/optnone（否则跨模块内联就没了）：\n%s", line)
			}
		}
	}
}

func keysOf(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// TestLTOBuildRuns 是 `--lto` 的端到端判据：整程序合并后程序行为必须与非 LTO 一致。
func TestLTOBuildRuns(t *testing.T) {
	root := writeProject(t, `import std.io;

module {
    func main() {
        var t: U16 = 0;
        var i: U16 = 0;
        for i < 5 {
            t = t + i;
            i = i + 1;
        }
        io.printf("t=%u\n", io.FmtVal.U { v: #compile.as(U64, t) });
    }
}
`)
	res := Build(root, Options{LTO: true})
	if !res.OK {
		if reason, missing := toolchainMissing(res.Diags); missing {
			t.Skip(reason)
		}
		t.Fatalf("--lto 构建失败：\n%s", res.Diags.Render(false))
	}
	got := runBuilt(t, res.Exe)
	if got != "t=10\n" {
		t.Fatalf("--lto 产物输出不对：%q", got)
	}
}
