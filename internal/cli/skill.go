package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"aila/internal/skill"
)

// newSkillCmd 实现 `aila skill`：把「给 AI 编程工具的技能包」交到用户手上。
//
// 设计取向：AI 需要的是**文本**，所以默认动作是把入口文档打到 stdout，
// 让"把 aila 技能喂给你的 AI"变成一条管道；其余动作（列出/读单篇/安装）都是可选项。
func newSkillCmd() *cobra.Command {
	var (
		showPath bool
		listMode bool
		installTo string
	)
	cmd := &cobra.Command{
		Use:   "skill [文件]",
		Short: "AI 技能包：打印入口文档 / 列出文件 / 装进你的 AI 工具",
		Long: `AILA 自带一份写给 AI 编程工具的技能包（SKILL.md + reference/*.md），
讲的是「怎么用这个打包好的 aila」而不是编译器源码。

  aila skill                          把入口文档 SKILL.md 打到 stdout
  aila skill reference/language.md    打印其中一篇
  aila skill --list                   列出全部文件
  aila skill --path                   打印技能目录
  aila skill --install <目录>         复制到 <目录>/aila（例如 ~/.claude/skills）`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := skill.Dir()
			if _, err := os.Stat(dir); err != nil {
				return fmt.Errorf("技能目录不存在：%s（发布包应带 share/aila/skills/%s）", dir, skill.Name)
			}
			if installTo != "" {
				dest, err := skill.Install(installTo)
				if err != nil {
					return err
				}
				if flags.json {
					return writeJSON(map[string]any{"installed": dest})
				}
				fmt.Fprintf(stdout, "已安装技能到 %s\n", dest)
				fmt.Fprintf(stdout, "把它的 %s 交给你的 AI（或在 AI 工具里把该目录登记为 skill 源）。\n", skill.Entry)
				return nil
			}
			if showPath {
				if flags.json {
					return writeJSON(map[string]any{"path": dir})
				}
				fmt.Fprintln(stdout, dir)
				return nil
			}
			files, err := skill.Files()
			if err != nil {
				return err
			}
			if listMode {
				if flags.json {
					return writeJSON(map[string]any{"path": dir, "files": files})
				}
				fmt.Fprintf(stdout, "技能目录：%s\n", dir)
				for _, f := range files {
					fmt.Fprintf(stdout, "  %-32s %6d B\n", f.Rel, f.Bytes)
				}
				return nil
			}
			rel := skill.Entry
			if len(args) == 1 {
				rel = args[0]
			}
			text, err := skill.Read(rel)
			if err != nil {
				return fmt.Errorf("读不到技能文件 %q：%w（用 aila skill --list 看有哪些）", rel, err)
			}
			fmt.Fprint(stdout, text)
			return nil
		},
	}
	cmd.Flags().BoolVar(&showPath, "path", false, "只打印技能目录")
	cmd.Flags().BoolVar(&listMode, "list", false, "列出技能里的全部文件")
	cmd.Flags().StringVar(&installTo, "install", "", "把技能复制到指定目录下的 "+skill.Name+"/（如 ~/.claude/skills）")
	return cmd
}

// writeJSON 是 --json 的统一出口。
func writeJSON(v any) error {
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
