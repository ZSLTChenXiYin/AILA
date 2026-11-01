// skillgen 生成 AI 技能包里两份「必须与实现同步」的参考：
//
//	share/aila/skills/aila/reference/cli.md     命令 / 参数 / 退出码（从 cobra 命令树）
//	share/aila/skills/aila/reference/errors.md  错误码 → 含义 → 修法（从 share/aila/data/diag_codes.json）
//
// 为什么要生成：这两份是 AI 最常查的内容，手抄必然漂移。CI 用 `-check` 卡住漂移。
//
// 用法：
//
//	go run ./tools/skillgen            # 重新生成
//	go run ./tools/skillgen -check     # 只校验（CI 用）
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"aila/internal/cli"
)

const (
	skillRoot  = "share/aila/skills/aila"
	bannerCLI  = "<!-- 生成物：由 `go run ./tools/skillgen` 从命令树生成，勿手改；CI 用 -check 校验。 -->"
	bannerErrs = "<!-- 生成物：由 `go run ./tools/skillgen` 从 share/aila/data/diag_codes.json 生成，勿手改；CI 用 -check 校验。 -->"
)

func main() {
	check := flag.Bool("check", false, "只校验生成物是否与实现一致")
	flag.Parse()
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "取不到工作目录：", err)
		os.Exit(2)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		fmt.Fprintln(os.Stderr, "请在仓库根目录运行（找不到 go.mod）")
		os.Exit(2)
	}

	docs := []struct {
		rel  string
		body func(root string) (string, error)
	}{
		{"reference/cli.md", renderCLI},
		{"reference/errors.md", renderErrors},
	}

	stale := 0
	for _, d := range docs {
		want, err := d.body(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "生成 %s 失败：%v\n", d.rel, err)
			os.Exit(1)
		}
		path := filepath.Join(root, filepath.FromSlash(skillRoot), filepath.FromSlash(d.rel))
		if *check {
			old, err := os.ReadFile(path)
			if err != nil || string(old) != want {
				fmt.Fprintf(os.Stderr, "  - %s 与实现不一致\n", d.rel)
				stale++
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
			fmt.Fprintf(os.Stderr, "建目录失败：%v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(path, []byte(want), 0o666); err != nil {
			fmt.Fprintf(os.Stderr, "写 %s 失败：%v\n", d.rel, err)
			os.Exit(1)
		}
		fmt.Printf("已生成 %s（%d 字节）\n", d.rel, len(want))
	}
	if *check {
		if stale > 0 {
			fmt.Fprintln(os.Stderr, "技能参考与当前实现不一致，重新生成：go run ./tools/skillgen")
			os.Exit(1)
		}
		fmt.Println("技能参考与当前实现一致")
	}
}

// --------------------------------------------------------------------------- //
// CLI 参考
// --------------------------------------------------------------------------- //

func renderCLI(root string) (string, error) {
	var b strings.Builder
	b.WriteString(bannerCLI + "\n\n")
	b.WriteString("# `aila` 命令参考\n\n")
	b.WriteString("> 有疑问的命令随时自查：`aila <子命令> --help`。退出码：`0` 成功、")
	b.WriteString("`1` 有诊断、`2` 用法错误、`3` 环境不满足、`4` 管理器内部错误。\n\n")

	var walk func(cmd *cobra.Command, path string)
	walk = func(cmd *cobra.Command, path string) {
		full := strings.TrimSpace(path + " " + cmd.Name())
		b.WriteString("## `" + full + "`\n\n")
		if long := strings.TrimSpace(cmd.Long); long != "" {
			b.WriteString(long + "\n\n")
		} else if short := strings.TrimSpace(cmd.Short); short != "" {
			b.WriteString(short + "\n\n")
		}
		usage := strings.TrimSpace(strings.TrimPrefix(cmd.Use, cmd.Name()))
		if usage == "" {
			usage = "[command]"
		}
		b.WriteString("用法：`" + full + " " + usage + "`\n\n")
		writeFlags(&b, "参数", cmd.NonInheritedFlags())
		writeFlags(&b, "继承参数", cmd.InheritedFlags())

		subs := cmd.Commands()
		sort.Slice(subs, func(i, j int) bool { return subs[i].Name() < subs[j].Name() })
		for _, sub := range subs {
			if !sub.IsAvailableCommand() && sub.Name() != "help" {
				continue
			}
			if sub.Name() == "help" || sub.Name() == "completion" {
				continue // 框架自带的，不写进技能
			}
			walk(sub, full)
		}
	}
	walk(cli.RootCommand(), "")
	return b.String(), nil
}

func writeFlags(b *strings.Builder, title string, set *pflag.FlagSet) {
	var flags []*pflag.Flag
	seen := map[string]bool{}
	set.VisitAll(func(f *pflag.Flag) {
		if f.Hidden || seen[f.Name] {
			return
		}
		seen[f.Name] = true
		flags = append(flags, f)
	})
	if len(flags) == 0 {
		return
	}
	b.WriteString(title + "：\n\n| 参数 | 类型 | 默认 | 说明 |\n|---|---|---|---|\n")
	for _, f := range flags {
		def := f.DefValue
		if def == "" || def == "false" {
			def = "—"
		}
		fmt.Fprintf(b, "| `--%s` | %s | `%s` | %s |\n", f.Name, f.Value.Type(), mdEscape(def), mdEscape(f.Usage))
	}
	b.WriteString("\n")
}

// --------------------------------------------------------------------------- //
// 错误码参考
// --------------------------------------------------------------------------- //

type codeEntry struct {
	Title   string   `json:"title"`
	Explain string   `json:"explain"`
	Fix     []string `json:"fix"`
	Doc     string   `json:"doc"`
}

func renderErrors(root string) (string, error) {
	path := filepath.Join(root, "share", "aila", "data", "diag_codes.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var payload struct {
		Codes map[string]codeEntry `json:"codes"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", err
	}
	codes := make([]string, 0, len(payload.Codes))
	for c := range payload.Codes {
		codes = append(codes, c)
	}
	sort.Strings(codes)

	var b strings.Builder
	b.WriteString(bannerErrs + "\n\n")
	b.WriteString("# AILA 诊断码参考\n\n")
	fmt.Fprintf(&b, "共 %d 条。权威输出永远是 `aila explain <CODE>`（本表与它同源）。\n", len(codes))
	b.WriteString("`E` = 错误（退出码 1），`W` = 警告（不阻断构建）。\n\n")
	b.WriteString("| 码 | 标题 | 含义 | 怎么修 | 设计锚点 |\n|---|---|---|---|---|\n")
	for _, code := range codes {
		e := payload.Codes[code]
		fix := strings.Join(e.Fix, "<br>")
		if fix == "" {
			fix = "—"
		}
		doc := e.Doc
		if doc == "" {
			doc = "—"
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s | %s | %s |\n",
			code, mdEscape(e.Title), mdEscape(e.Explain), mdEscape(fix), mdEscape(doc))
	}
	b.WriteString("\n## 按段速查\n\n")
	b.WriteString("| 段 | 主题 |\n|---|---|\n")
	for _, seg := range []struct{ prefix, what string }{
		{"E1", "词法：字符、字符串/字符字面量、数字拼法、转义、注释、`::`"},
		{"E2", "语法：缺分号、裸块、match 歧义、比较链、前缀关键字、`if` 取值"},
		{"E3", "语义：类型不匹配、名字未定义/不可见、`as`/`trunc` 判据、丢弃与返回"},
		{"E34", "match：不完备、分支产出不统一、缺 `_`、块分支缺 `out`"},
		{"E36", "类型形状：不可为空的类型、没有静态布局的类型"},
		{"E4", "区域与借用：分配点、视图来路、区域名冲突、`view` 约束、`@` 子句、逃逸、只读、move/copy"},
		{"E5", "并发：`<-` 位置、thread 白名单、线程写全局、`yield` 位置"},
		{"E6", "C 边界：变参类型、union、bindgen 重名"},
		{"E7", "工程：清单、路径标识符、导入别名、入口、目标三元组、依赖、工具链"},
		{"E8", "构建：LLVM/cc 缺失、IR 校验、子进程失败、交叉链接"},
		{"E9", "待批次：本批次未实现的检查/代码生成"},
		{"W", "警告：转换建议、清单未知键、未映射的 C 声明、未使用绑定、版本带"},
	} {
		fmt.Fprintf(&b, "| `%s…` | %s |\n", seg.prefix, seg.what)
	}
	return b.String(), nil
}

// mdEscape 让文本能安全塞进 Markdown 表格的一格。
func mdEscape(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}
