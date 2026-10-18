package cli

import (
	"os"

	"github.com/spf13/cobra"

	"aila/internal/lsp"
)

// newLSPCmd 是 `aila lsp`：给编辑器用的语言服务器（JSON-RPC over stdio）。
//
// 编辑器插件（editors/vscode）把它当子进程拉起来，拿到语义级能力：
// 大纲、跳转定义、悬停、查找引用、诊断标红、工程内符号搜索。
// 语义只来自编译器本体，插件不做第二份语法树。
func newLSPCmd() *cobra.Command {
	var logPath string
	cmd := &cobra.Command{
		Use:   "lsp",
		Short: "启动语言服务器（编辑器插件用，JSON-RPC over stdio）",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// stdout 归协议独占：日志只能走 stderr 或文件
			var logw *os.File
			if logPath != "" {
				f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
				if err != nil {
					return &exitError{ExitUsage}
				}
				logw = f
				defer f.Close()
			} else {
				logw = os.Stderr
			}
			srv := lsp.NewServer(os.Stdin, os.Stdout, logw)
			if err := srv.Serve(); err != nil {
				return &exitError{ExitInternal}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&logPath, "log", "", "把服务器日志写到文件（默认 stderr）")
	return cmd
}
