package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/aila-lang/aila/internal/codegen"
	"github.com/aila-lang/aila/internal/parser"
	"github.com/aila-lang/aila/internal/resolve"
	"github.com/aila-lang/aila/internal/sema"
	"github.com/aila-lang/aila/internal/types"
)

// newIRCmd 是 G3 落地的可见入口：单文件走完「解析 → 语义 → LLVM IR」，把 IR 打到 stdout。
//
// 用途：①与参考实现的 `*.ir` 快照逐字节对照（tests/corpus）；②没有 clang/llc 的环境里
// 也能看到前端产物。多模块工程请用 `aila build --emit-ir`（模块图批次）。
func newIRCmd() *cobra.Command {
	var triple string
	cmd := &cobra.Command{
		Use:   "ir <文件.aila>",
		Short: "把一份 AILA 源文件降成 LLVM IR 文本并打印",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			raw, err := os.ReadFile(path)
			if err != nil {
				fmt.Fprintf(stderr, "读不到文件：%s\n", path)
				return &exitError{ExitUsage}
			}
			if triple == "" {
				triple = resolve.HostTriple()
			}
			mod, bag := parser.Parse(path, string(raw))
			if bag.HasErrors() {
				printDiags(bag)
				return &exitError{ExitDiag}
			}
			table := sema.BuildTable(mod, bag)
			sigs := sema.BuildSigs(mod, table, bag)
			check := sema.Check(mod, table, sigs, bag)
			if bag.HasErrors() {
				printDiags(bag)
				return &exitError{ExitDiag}
			}
			layout := types.NewLayout(8)
			lw := codegen.New(mod, "_.main", table, sigs, check, layout, triple, "_.main")
			ir := lw.Generate()
			if blockers := lw.Unsupported(); len(blockers) > 0 {
				for _, what := range blockers {
					fmt.Fprintf(stderr, "提示：%s\n", what)
				}
			}
			fmt.Fprint(stdout, ir)
			return nil
		},
	}
	cmd.Flags().StringVar(&triple, "triple", "", "目标三元组（默认取宿主三元组）")
	return cmd
}
