package diag

import (
	"strings"
	"testing"
)

func TestCatalogLoads(t *testing.T) {
	c := Catalog
	_ = c
	catalog, err := Catalog()
	if err != nil {
		t.Fatalf("加载错误码表失败：%v", err)
	}
	if len(catalog) < 70 {
		t.Fatalf("错误码数量异常：%d", len(catalog))
	}
	entry, ok := Find("E3201")
	if !ok {
		t.Fatal("E3201 未登记")
	}
	if !strings.Contains(entry.Title, "类型不匹配") {
		t.Fatalf("E3201 标题异常：%q", entry.Title)
	}
	if len(entry.Fix) == 0 {
		t.Fatal("E3201 没有修复步骤")
	}
}

func TestFormatExplain(t *testing.T) {
	out := FormatExplain("E4101")
	if !strings.Contains(out, "E4101") || !strings.Contains(out, "怎么修") {
		t.Fatalf("explain 输出异常：%q", out)
	}
	unknown := FormatExplain("E0000")
	if !strings.Contains(unknown, "未登记") {
		t.Fatalf("未知码提示异常：%q", unknown)
	}
}

func TestRenderMatchesPythonShape(t *testing.T) {
	bag := &Bag{}
	span := Span{File: "main.aila", Line: 12, Col: 18, Length: 5, Text: "    let x: F32 = count;"}
	bag.Error("E3201", span, "类型不匹配", "期望 `F32`，实得 `I32`", "AILA 零隐式转换")
	out := bag.Render(false)
	for _, want := range []string{"error[E3201]: 类型不匹配", "main.aila:12:18", "let x: F32 = count;", "^^^^^", "帮助"} {
		if !strings.Contains(out, want) {
			t.Fatalf("渲染缺少 %q：\n%s", want, out)
		}
	}
	if !strings.Contains(bag.Summary(false), "1 个错误") {
		t.Fatalf("汇总异常：%q", bag.Summary(false))
	}
}

func TestUnknownCodePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("未登记的错误码应当立刻 panic（错误码唯一事实源纪律）")
		}
	}()
	bag := &Bag{}
	bag.Error("E9999", Empty, "x", "", "")
}

func TestCounts(t *testing.T) {
	bag := &Bag{}
	bag.Warning("W7001", Empty, "w", "", "")
	bag.Error("E7002", Empty, "e", "", "")
	e, w := bag.Counts()
	if e != 1 || w != 1 {
		t.Fatalf("计数异常：%d/%d", e, w)
	}
	if !bag.HasErrors() {
		t.Fatal("HasErrors 应为 true")
	}
}
