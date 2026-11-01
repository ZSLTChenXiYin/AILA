package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGCSectionsDropsPrunedDriver 是 designs/0006 §3 断言 4 的落地：
// `#compile.match` 选中的驱动进产物，**未选中的驱动连符号都不在**（编译期剪枝
// 不留调用 + 链接期 --gc-sections 丢掉没人引用的函数）。
//
// 判据用符号名扫描而不是 nm：可执行文件的符号字符串表里有 `aila_<模块>__<函数>`，
// 扫字节就够，且不依赖宿主是否装了 nm。
func TestGCSectionsDropsPrunedDriver(t *testing.T) {
	root := t.TempDir()
	write := func(path, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o666); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(root, "aila.toml"),
		"[package]\nname = \"prune\"\nversion = \"0.1.0\"\nentry = \"_.main\"\n\n[build]\nopt = 0\n")

	// 两个驱动各占一个函数；SEL 决定谁被选中（另一个分支整段不生成代码）
	program := func(sel string) string {
		return "import std.io;\n\nconst SEL: U8 = " + sel + ";\n\nmodule {\n" +
			"    func drv_a() -> U16 { return 11; }\n" +
			"    func drv_b() -> U16 { return 22; }\n\n" +
			"    func main() {\n" +
			"        var v: U16 = 0;\n" +
			"        #compile.match SEL {\n" +
			"            1 => { v = drv_a(); },\n" +
			"            _ => { v = drv_b(); },\n" +
			"        }\n" +
			"        io.printf(\"v=%u\\n\", io.FmtVal.U { v: #compile.as(U64, v) });\n" +
			"    }\n}\n"
	}

	cases := []struct {
		sel    string
		keep   string
		drop   string
		output string
	}{
		{"1", "drv_a", "drv_b", "v=11\n"},
		{"2", "drv_b", "drv_a", "v=22\n"},
	}
	src := filepath.Join(root, "main.aila")
	for _, tc := range cases {
		write(src, program(tc.sel))
		res := Build(root, Options{GCSections: true, SizeDelta: true})
		if !res.OK {
			if reason, missing := toolchainMissing(res.Diags); missing {
				t.Skip(reason)
			}
			t.Fatalf("SEL=%s 构建失败：\n%s", tc.sel, res.Diags.Render(false))
		}
		exe, err := os.ReadFile(res.Exe)
		if err != nil {
			t.Fatalf("读不到产物：%v", err)
		}
		blob := string(exe)
		if !strings.Contains(blob, "aila_main__"+tc.keep) {
			t.Fatalf("SEL=%s：选中驱动 %s 的符号应当留在产物里", tc.sel, tc.keep)
		}
		if strings.Contains(blob, "aila_main__"+tc.drop) {
			t.Fatalf("SEL=%s：未选中驱动 %s 的符号不该出现在产物里（--gc-sections 没生效？）",
				tc.sel, tc.drop)
		}
		if size, ok := res.Artifacts["size"]; !ok || size == nil {
			t.Fatalf("SEL=%s：--emit-size-delta 应产出体积报告", tc.sel)
		}
	}
}
