package build

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// moduleInterfacesOf 读 aila-build.json 里的 module_interfaces 段。
func moduleInterfacesOf(t *testing.T, root string) map[string]map[string]any {
	t.Helper()
	res := Load(root, Options{})
	res = Load(root, Options{})
	if res.Diags.HasErrors() {
		t.Fatalf("装载失败：\n%s", res.Diags.Render(false))
	}
	if res.Program == nil {
		t.Fatalf("没有产出模块图")
	}
	irs := res.GenerateIR()
	if res.Diags.HasErrors() {
		t.Fatalf("代码生成失败：\n%s", res.Diags.Render(false))
	}
	out := map[string]map[string]any{}
	for name, v := range moduleInterfaces(irs) {
		if m, ok := v.(map[string]any); ok {
			out[name] = m
		}
	}
	return out
}

// TestModuleInterfaceDigest 是 designs/0006 §5「接口指纹最小可用形态」的判据：
//
//   - 换 `#compile.match` 的 const 取值（判定另一支被剪）→ 摘要必须**逐字不变**（剪枝不动 ABI）；
//   - 改函数签名 → 摘要必须变（否则热替换会把不兼容的实现当兼容）。
func TestModuleInterfaceDigest(t *testing.T) {
	writeProg := func(sel, extraParam string) string {
		callExtra := ""
		if extraParam != "" {
			callExtra = ", 0"
		}
		return `import std.io;

const SEL: U8 = ` + sel + `;

module {
    func drv_a() -> U16 { return 11; }
    func drv_b() -> U16 { return 22; }

    public func api(x: U16` + extraParam + `) -> U16 {
        var v: U16 = 0;
        #compile.match SEL {
            1 => { v = drv_a(); },
            _ => { v = drv_b(); },
        }
        return v + x;
    }

    func main() {
        io.printf("%u\n", io.FmtVal.U { v: #compile.as(U64, api(0` + callExtra + `)) });
    }
}
`
	}
	rootA := writeProject(t, writeProg("1", ""))
	a := moduleInterfacesOf(t, rootA)
	rootB := writeProject(t, writeProg("2", ""))
	b := moduleInterfacesOf(t, rootB)

	digestOf := func(m map[string]map[string]any, mod string) string {
		t.Helper()
		entry, ok := m[mod]
		if !ok {
			t.Fatalf("报告里没有模块 %s：%v", mod, m)
		}
		s, _ := entry["digest"].(string)
		return s
	}
	if digestOf(a, "_.main") == "" {
		t.Fatal("摘要不该为空")
	}
	if digestOf(a, "_.main") != digestOf(b, "_.main") {
		t.Fatalf("剪枝换了分支后接口摘要不该变（剪枝不动 ABI）：%s vs %s",
			digestOf(a, "_.main"), digestOf(b, "_.main"))
	}

	// 改签名（多一个参数）→ 摘要必须变
	rootC := writeProject(t, writeProg("1", ", y: U16"))
	c := moduleInterfacesOf(t, rootC)
	if digestOf(a, "_.main") == digestOf(c, "_.main") {
		t.Fatalf("改了函数签名后接口摘要必须变，实得同一个 %s", digestOf(c, "_.main"))
	}
}

// TestBuildReportHasInterfaces 确认 `--emit-json` 真的把摘要写进了 aila-build.json。
func TestBuildReportHasInterfaces(t *testing.T) {
	root := writeProject(t, `import std.io;

module {
    func main() {
        io.printf("ok\n");
    }
}
`)
	res := Build(root, Options{EmitJSON: true})
	if !res.OK {
		if reason, missing := toolchainMissing(res.Diags); missing {
			t.Skip(reason)
		}
		t.Fatalf("构建失败：\n%s", res.Diags.Render(false))
	}
	raw, err := os.ReadFile(filepath.Join(res.OutDir, "aila-build.json"))
	if err != nil {
		t.Fatalf("读不到构建报告：%v", err)
	}
	var report map[string]any
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatalf("构建报告不是合法 JSON：%v", err)
	}
	ifaces, ok := report["module_interfaces"].(map[string]any)
	if !ok || len(ifaces) == 0 {
		t.Fatalf("aila-build.json 缺 module_interfaces：%s", string(raw))
	}
	if _, ok := ifaces["_.main"]; !ok {
		t.Fatalf("module_interfaces 里没有根包：%v", ifaces)
	}
}
