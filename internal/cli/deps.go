package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ailang/aila/internal/deps"
	"github.com/ailang/aila/internal/diag"
	"github.com/ailang/aila/internal/manifest"
)

// newDepsCmd 是 `aila deps`：打印依赖图 / 更新锁 / 生成 vendor。
//
// 跨项目依赖的规则见 docs/rfcs/0004：
//   - 清单 `[dependencies]`：`mylib = { path = "../mylib" }` 或 `{ git = "...", tag = "v1.0.0" }`
//   - 解析结果写进 `aila.lock`（来源 + 版本 + 内容哈希），依赖变了不 `--update` 就报错
//   - `--vendor` 把依赖复制到 `vendor/`，离线构建优先用它
func newDepsCmd() *cobra.Command {
	var update, vendor, asJSON bool
	cmd := &cobra.Command{
		Use:   "deps [路径]",
		Short: "查看/更新跨项目依赖（打印依赖图、写锁文件、生成 vendor）",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot(argOrDot(args))
			bag := &diag.Bag{}
			man := manifest.Load(root, bag)
			if man == nil {
				printDiags(bag)
				return &exitError{ExitDiag}
			}
			graph := deps.Resolve(root, update, bag)
			if vendor {
				if err := graph.Vendor(); err != nil {
					bag.Error("E7022", diag.Empty, "vendor 复制失败", err.Error(), "")
				}
			}
			if bag.HasErrors() {
				printDiags(bag)
				return &exitError{ExitDiag}
			}
			if update || !graph.Locked() {
				if err := graph.Write(); err != nil {
					bag.Warning("W7021", diag.Empty, "写不了 `aila.lock`", err.Error(), "")
				}
				fmt.Fprintf(stdout, "已更新 %s\n", filepath.Join(root, deps.LockFileName))
			}
			if asJSON {
				type row struct {
					Name     string `json:"name"`
					Version  string `json:"version"`
					Source   string `json:"source"`
					Hash     string `json:"hash"`
					Dir      string `json:"dir"`
					Vendored bool   `json:"vendored"`
				}
				rows := []row{}
				for _, pkg := range graph.Order {
					if pkg == deps.RootPkg {
						continue
					}
					r := graph.ByPkg[pkg]
					rows = append(rows, row{pkg, r.Version, r.Source, r.Hash, r.Dir, r.Vendored})
				}
				printJSON(rows)
				return nil
			}
			if len(graph.Order) <= 1 {
				fmt.Fprintln(stdout, "没有依赖（清单里没有 `[dependencies]`）")
				return nil
			}
			fmt.Fprint(stdout, graph.Describe())
			return nil
		},
	}
	cmd.Flags().BoolVar(&update, "update", false, "重新解析依赖并重写 aila.lock")
	cmd.Flags().BoolVar(&vendor, "vendor", false, "把依赖复制到 vendor/（离线构建）")
	cmd.Flags().BoolVar(&asJSON, "json", false, "机器可读输出")
	return cmd
}

// printJSON 是 dbg 之外的统一 JSON 出口（`deps --json` 用）。
func printJSON(v any) {
	data, err := jsonMarshalIndent(v)
	if err != nil {
		fmt.Fprintf(stderr, "JSON 编码失败：%v\n", err)
		return
	}
	fmt.Fprintln(stdout, string(data))
}
