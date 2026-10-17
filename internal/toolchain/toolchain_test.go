package toolchain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aila-lang/aila/internal/resolve"
)

func TestIRToObjectRejectsMissingTool(t *testing.T) {
	tool := resolve.Tool{Role: resolve.RoleIR2Obj, Source: resolve.SourceMissing, Note: "没有随包 LLVM"}
	err := IRToObject("target triple = \"x86_64-unknown-linux-gnu\"\n", filepath.Join(t.TempDir(), "a.o"), "x86_64-unknown-linux-gnu", tool)
	if err == nil {
		t.Fatal("缺后端时应报错")
	}
	if !strings.Contains(err.Error(), "没有随包 LLVM") {
		t.Fatalf("错误信息应带上原因：%v", err)
	}
}

func TestSmokeTestExplainsMissingToolchain(t *testing.T) {
	missing := resolve.Tool{Role: resolve.RoleIR2Obj, Source: resolve.SourceMissing, Note: "没有随包 LLVM"}
	ok, detail, fix := SmokeTest(missing, missing)
	if ok {
		t.Fatal("缺工具时冒烟测试不应通过")
	}
	if !strings.Contains(detail, "LLVM 后端不可用") {
		t.Fatalf("说明异常：%s", detail)
	}
	if !strings.Contains(fix, "toolchain install") {
		t.Fatalf("应给出装包命令：%s", fix)
	}
}

func TestDoctorRows(t *testing.T) {
	checks := Doctor("test-version", "", "")
	names := map[string]bool{}
	for _, c := range checks {
		names[c.Name] = true
		if c.Level != "ok" && c.Level != "warn" && c.Level != "fail" {
			t.Fatalf("非法等级：%s=%s", c.Name, c.Level)
		}
	}
	for _, want := range []string{"AILA_HOME", "随包工具链", "LLVM 后端（IR → .o）", "C 编译器 / 链接器", "运行时源码", "标准库", "链接冒烟测试"} {
		if !names[want] {
			t.Fatalf("doctor 缺少检查项 %q：%v", want, names)
		}
	}
	if Failed(checks) < 0 {
		t.Fatal("Failed 不应为负")
	}
}

func TestCompileCLinksNothingWhenEmpty(t *testing.T) {
	missing := resolve.Tool{Role: resolve.RoleCC, Source: resolve.SourceMissing, Note: "没有 cc"}
	if _, err := CompileC(nil, t.TempDir(), missing, nil); err == nil {
		t.Fatal("缺编译器时应报错")
	}
	if err := Link(nil, filepath.Join(t.TempDir(), "x"), missing, nil); err == nil {
		t.Fatal("缺链接器时应报错")
	}
}

func TestFirstLine(t *testing.T) {
	if got := firstLine(""); got != "无输出" {
		t.Fatalf("firstLine(\"\")=%q", got)
	}
	if got := firstLine("a\nb"); got != "a" {
		t.Fatalf("firstLine=%q", got)
	}
	long := strings.Repeat("x", 1000)
	if len(firstLine(long)) > 400 {
		t.Fatal("长输出应被截断")
	}
	_ = os.TempDir()
}
