package sema

import "testing"

// TestLegacyBuiltinVariantsRetired 守护 designs/0004 §7 的下线清单。
//
// v2 之前 `Some { v }` / `Ok { v }` / `Err { e }` 是内置变体名，构造与 match
// 模式两用。它们已随 v2 全部退役：可空写 `let v: 类型` / `None`，错误通道写
// `let v: 类型` / `let e: 错误集`，构造走「返回位置自动装载 + leading-dot 错误
// 成员构造」。本用例锁住「谁把它们请回来就红」。
func TestLegacyBuiltinVariantsRetired(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{
			name: "Some 构造",
			src: "func opt() -> Text? { return \"x\"; }\n\nfunc main() {\n" +
				"    let a: Text? = Some { v: \"y\" };\n    _ = a;\n}\n",
		},
		{
			name: "Ok 构造",
			src: "#comprun.error enum Error { Nack };\n\nfunc main() {\n" +
				"    let b: U8 ! Error = Ok { v: 1 };\n    _ = b;\n}\n",
		},
		{
			name: "Err 构造",
			src: "#comprun.error enum Error { Nack };\n\nfunc main() {\n" +
				"    let c: U8 ! Error = Err { e: .Nack };\n    _ = c;\n}\n",
		},
		{
			name: "Some / None 变体模式",
			src: "func opt(t: Text) -> Text? { return t; }\n\nfunc main() {\n" +
				"    let a = opt(\"x\");\n    match a {\n" +
				"        Some { v } => { _ = v; },\n        None       => {},\n    }\n}\n",
		},
		{
			name: "Ok / Err 变体模式",
			src: "#comprun.error enum Error { Nack };\n\nfunc f() -> U8 ! Error { return 7; }\n\n" +
				"func main() {\n    let r = f();\n    match r {\n" +
				"        Ok { v }  => { _ = v; },\n        Err { e } => { _ = e; },\n    }\n}\n",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			codes, _ := runModules(t, map[string]string{"_.main": c.src})
			if !codes["E3001"] {
				t.Fatalf("旧内置变体名应报 E3001，实际错误码：%v", codes)
			}
		})
	}
}

// TestResumeValueFormRetired 守护 P10：`resume h` 只能作为 select 事件。
// 三态判别和（Suspended / Done）不再逃到用户代码，等完成语义由 select 承担。
func TestResumeValueFormRetired(t *testing.T) {
	src := `coroutine func g() -> I32 { return 1; }

func main() {
    region r {
        let h = #runtime.coro(g);
        let v = resume h;
        _ = v;
    }
}
`
	codes, _ := runModules(t, map[string]string{"_.main": src})
	if !codes["E5010"] {
		t.Fatalf("`resume` 表达式形态应报 E5010，实际错误码：%v", codes)
	}
}

// TestSelectAwaitCoro 是对照组：select 的等完成事件在 main 里也合法，
// 绑定值就是完成态（这里是 `I32`），没有 Suspended / Done 分支要写。
func TestSelectAwaitCoro(t *testing.T) {
	src := `coroutine func g() -> I32 { return 1; }

func main() {
    region r {
        let h = #runtime.coro(g);
        select {
            case v = resume h => { _ = v; },
        }
    }
}
`
	codes, _ := runModules(t, map[string]string{"_.main": src})
	if len(codes) != 0 {
		t.Fatalf("select 等完成事件不应报错，实际错误码：%v", codes)
	}
}

// TestV2FormsStillAccepted 是对照组：退役只针对旧拼法，v2 拼法必须原样可用。
func TestV2FormsStillAccepted(t *testing.T) {
	src := `#comprun.error enum Error { Nack };

func f() -> U8 ! Error { return .Nack; }

func opt(t: Text) -> Text? { return t; }

func main() {
    let r = f();
    match r {
        let v: U8    => { _ = v; },
        let e: Error => { _ = e; },
    }
    let o = opt("x");
    match o {
        let s: Text => { _ = s; },
        None        => {},
    }
}
`
	codes, _ := runModules(t, map[string]string{"_.main": src})
	if len(codes) != 0 {
		t.Fatalf("v2 拼法不应报错，实际错误码：%v", codes)
	}
}
