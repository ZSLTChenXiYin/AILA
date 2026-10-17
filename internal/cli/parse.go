package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aila-lang/aila/internal/ast"
	"github.com/aila-lang/aila/internal/diag"
	"github.com/aila-lang/aila/internal/lexer"
	"github.com/aila-lang/aila/internal/parser"
)

// newParseCmd 是 G2 落地的可见入口：把一份 `.aila` 的词法 / 语法结果打出来。
// 用途：①排查 Go 与 Python 的解析差异；②给编辑器 / CI 一个不依赖 Python 的解析检查。
func newParseCmd() *cobra.Command {
	var tokensOnly bool
	cmd := &cobra.Command{
		Use:   "parse <文件.aila>",
		Short: "解析一份 AILA 源文件并打印 token 流或语法树（默认 S-表达式）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			raw, err := os.ReadFile(path)
			if err != nil {
				fmt.Fprintf(stderr, "读不到文件：%s\n", path)
				return &exitError{ExitUsage}
			}
			if tokensOnly {
				toks, bag := lexer.Lex(path, string(raw))
				for _, t := range toks {
					if t.Kind == lexer.KindEOF {
						break
					}
					fmt.Fprintf(stdout, "%4d:%-4d %-8s %s\n", t.Span.Line, t.Span.Col, t.Kind, t.Text)
				}
				if bag.HasErrors() {
					printDiags(bag)
					return &exitError{ExitDiag}
				}
				return nil
			}
			mod, bag := parser.Parse(path, string(raw))
			if bag.HasErrors() {
				printDiags(bag)
				return &exitError{ExitDiag}
			}
			fmt.Fprintln(stdout, ast.Sexp(mod))
			return nil
		},
	}
	cmd.Flags().BoolVar(&tokensOnly, "tokens", false, "只打印 token 流")
	return cmd
}

// ensureImports 防止 diag 的渲染入口被优化掉（parse 命令里用到了）。
var _ = diag.Empty
var _ = strings.TrimSpace
