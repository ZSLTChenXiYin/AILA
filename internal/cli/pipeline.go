package cli

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/ailang/aila/internal/build"
)

// loadOpts 把命令行开关转成构建选项。
func loadOpts() build.Options {
	return build.Options{
		Board:    buildBoard,
		Opt:      buildOpt,
		LLVM:     buildLLVM,
		CC:       buildCC,
		EmitIR:   buildEmitIR,
		EmitJSON: buildEmitJSON,
		OutName:  buildOutName,
		UpdateDeps: buildDepsUpd,
		Vendor:     buildVendor,
	}
}

// reportDiags 打印诊断并返回退出码（有错误 → 1）。
func reportDiags(res *build.Result) error {
	if res.Diags != nil && len(res.Diags.Items) > 0 {
		printDiags(res.Diags)
	}
	if res.Diags != nil && res.Diags.HasErrors() {
		return &exitError{ExitDiag}
	}
	return nil
}

// newCheckCmd 是 `aila check`：只做检查，不产产物。
func newCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check [路径]",
		Short: "只做检查（解析 + 语义），不产产物",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot(argOrDot(args))
			res := build.Check(root, loadOpts())
			if err := reportDiags(res); err != nil {
				return err
			}
			mods := len(res.Program.Order)
			fmt.Fprintf(stdout, "检查通过：%d 个模块（entry = %s，目标 %s）\n",
				mods, res.Manifest.Entry, res.Triple)
			return nil
		},
	}
}

// newBuildCmd 是 `aila build`：全流程构建。AILA 是编译型语言，产物是可执行文件。
func newBuildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build [路径]",
		Short: "全流程构建（IR → 目标文件 → 链接）",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot(argOrDot(args))
			res := build.Build(root, loadOpts())
			if err := reportDiags(res); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "构建完成：%s\n", res.Exe)
			return nil
		},
	}
}

// newRunCmd 是 `aila run`：构建后立即运行。
func newRunCmd() *cobra.Command {
	var passArgs []string
	cmd := &cobra.Command{
		Use:   "run [路径] [-- 参数...]",
		Short: "构建并运行",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot(argOrDot(args))
			res := build.Build(root, loadOpts())
			if err := reportDiags(res); err != nil {
				return err
			}
			proc := exec.Command(res.Exe, passArgs...)
			proc.Stdout = stdout
			proc.Stderr = stderr
			proc.Stdin = stdin
			if err := proc.Run(); err != nil {
				if ee, ok := err.(*exec.ExitError); ok {
					return &exitError{ee.ExitCode()}
				}
				fmt.Fprintf(stderr, "运行失败：%v\n", err)
				return &exitError{ExitInternal}
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&passArgs, "arg", nil, "传给程序的参数（可重复）")
	return cmd
}
