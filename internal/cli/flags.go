package cli

import (
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// stdin 可被测试替换（与 stdout / stderr 同一套路）。
var stdin io.Reader = os.Stdin

// 构建相关开关（check / build / run 共用）。
var (
	buildBoard    string
	buildOpt      int
	buildLLVM     string
	buildCC       string
	buildEmitIR   bool
	buildEmitJSON bool
	buildOutName  string
	buildDepsUpd  bool
	buildVendor   bool
)

// jsonMarshalIndent 统一缩进 JSON（2 空格）。
func jsonMarshalIndent(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// addBuildFlags 给命令挂上构建开关。
func addBuildFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&buildBoard, "board", "", "目标 board（临时覆盖清单）")
	cmd.Flags().IntVar(&buildOpt, "opt", 0, "优化级别（0-3，临时覆盖清单）")
	cmd.Flags().StringVar(&buildLLVM, "llvm", "", "LLVM 后端选择（auto / bundled / system:...）")
	cmd.Flags().StringVar(&buildCC, "cc", "", "C 编译器 / 链接器选择（auto / bundled / system:...）")
	cmd.Flags().BoolVar(&buildEmitIR, "emit-ir", false, "额外写出 `.ll` 文本")
	cmd.Flags().BoolVar(&buildEmitJSON, "emit-json", false, "写出构建报告 `aila-build.json`")
	cmd.Flags().StringVar(&buildOutName, "out", "", "输出文件名（默认 `app` + 平台后缀）")
	cmd.Flags().BoolVar(&buildDepsUpd, "deps-update", false, "重新解析跨项目依赖并重写 aila.lock")
	cmd.Flags().BoolVar(&buildVendor, "vendor", false, "把依赖复制到 vendor/（离线构建）")
}
