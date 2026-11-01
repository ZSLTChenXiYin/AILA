package build

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// writeProject 在 t.TempDir 里铺一个最小工程（根包只有一个 main.aila）。
func writeProject(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "aila.toml"),
		[]byte("[package]\nname = \"regress\"\nversion = \"0.1.0\"\nentry = \"_.main\"\n\n[build]\nopt = 0\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.aila"), []byte(body), 0o666); err != nil {
		t.Fatal(err)
	}
	return root
}

// runProject 构建并运行，返回归一化后的 stdout。
func runProject(t *testing.T, root string) string {
	t.Helper()
	res := Build(root, Options{})
	if !res.OK {
		if reason, missing := toolchainMissing(res.Diags); missing {
			t.Skip(reason)
		}
		t.Fatalf("构建失败：\n%s", res.Diags.Render(false))
	}
	cmd := exec.Command(res.Exe)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("运行失败：%v", err)
	}
	return strings.ReplaceAll(out.String(), "\r\n", "\n")
}

// TestGlobalTextConstant 守护模块级 Text 常量：`module { let S: Text = "…" }` 与
// `const S: Text = "…"` 都必须真的带上内容。
//
// 旧 bug：全局初值只认字面量标量 / 定长数组字面量，字符串字面量掉到兜底的
// zeroinitializer → 模块级 Text 变成 (null, 0)：打印为空、view 出来 len=0、
// 下标直接 trap（designs/0006 §4.1 第 6 条）。
func TestGlobalTextConstant(t *testing.T) {
	root := writeProject(t, `import std.io;

module {
    const G_CONST: Text = "const-text";      // 10 字节
    let G_LET: Text = "let-text";            // 8 字节

    func main() {
        io.println(io.FmtVal.T { v: G_CONST });
        io.println(io.FmtVal.T { v: G_LET });
        view a: [U8] = G_CONST;
        view b: [U8] = G_LET;
        io.printf("a=%u b=%u first=%u\n",
            io.FmtVal.U { v: #compile.as(U64, #comprun.len(a)) },
            io.FmtVal.U { v: #compile.as(U64, #comprun.len(b)) },
            io.FmtVal.U { v: #compile.as(U64, b[0]) });
    }
}
`)
	got := runProject(t, root)
	want := "const-text\nlet-text\na=10 b=8 first=108\n"
	if got != want {
		t.Fatalf("模块级 Text 常量不对：\n--- 实得 ---\n%s--- 期望 ---\n%s", got, want)
	}
}

// TestMutSelfForwardWritesBack 守护「@mut 参数的转发写权」：
//
//	func forward(c: Counter) @mut[c] { c.bump(); }   // bump 是 @mut[self] 方法
//
// 旧 bug：小聚合体（≤16 字节）的 `@mut` 参数按值传，被调方拿到的是副本，
// 转发调用的 self 是「副本的地址」——标量字段写不回调用方（designs/0006 §4.1 第 5 条）。
// 现在 @mut 的聚合体参数一律按指针传，转发写权才真正成立。
func TestMutSelfForwardWritesBack(t *testing.T) {
	root := writeProject(t, `import std.io;

struct Counter {
    v: U8,
};

module {
    func Counter.bump(self) @mut[self] {
        self.v = self.v + 1;
    }

    func make() -> Counter {
        return Counter { v: 0 };
    }

    func forward(c: Counter) @mut[c] {
        c.bump();
    }

    func main() {
        var c: Counter = make();
        forward(c);          // 经 @mut 参数转发 → 必须写回
        c.bump();            // 直接调用
        io.printf("v=%u\n", io.FmtVal.U { v: #compile.as(U64, c.v) });
    }
}
`)
	got := runProject(t, root)
	if got != "v=2\n" {
		t.Fatalf("@mut 转发没写回调用方：实得 %q，期望 \"v=2\\n\"", got)
	}
}

// TestViewOfModuleArray 复查一个历史坑：函数内 `view v: [U8] = 模块级数组;` 曾被编坏
//（裸机上 `LoadProhibited, excvaddr=0x3ff`，designs/0006 §4.1 第 1 条）。
// 这里在宿主上把「求和 / 下标 / 传参」三条路径都走一遍，值不对就红。
func TestViewOfModuleArray(t *testing.T) {
	root := writeProject(t, `import std.io;

module {
    let TAB: [U8; 4] = [10, 20, 30, 40];
    let CFG: [U8; 3] = [7, 8, 9];

    func sum(v: [U8]) -> USize {
        var s: USize = 0;
        var i: USize = 0;
        for i < #comprun.len(v) {
            s = s + #compile.as(USize, v[i]);
            i = i + 1;
        }
        return s;
    }

    func via_view() -> USize {
        view v: [U8] = TAB;
        return sum(v);
    }

    func via_index() -> USize {
        view v: [U8] = CFG;
        return #compile.as(USize, v[0]) + #compile.as(USize, v[2]);
    }

    func main() {
        io.printf("sum=%u idx=%u\n",
            io.FmtVal.U { v: #compile.as(U64, via_view()) },
            io.FmtVal.U { v: #compile.as(U64, via_index()) });
    }
}
`)
	got := runProject(t, root)
	if got != "sum=100 idx=16\n" {
		t.Fatalf("模块级数组的视图读错：实得 %q，期望 \"sum=100 idx=16\\n\"", got)
	}
}

// runBuilt 跑一个已经构建好的产物，返回归一化后的 stdout。
func runBuilt(t *testing.T, exe string) string {
	t.Helper()
	cmd := exec.Command(exe)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("运行失败：%v", err)
	}
	return strings.ReplaceAll(out.String(), "\r\n", "\n")
}