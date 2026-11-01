package cli

import (
	"fmt"
	"os/exec"
	"sort"

	"github.com/spf13/cobra"

	"aila/internal/build"
	"aila/internal/targets"
)

// loadOpts 把命令行开关转成构建选项。
func loadOpts() build.Options {
	return build.Options{
		Triple:       buildTriple,
		Opt:          buildOpt,
		LLVM:         buildLLVM,
		CC:           buildCC,
		EmitIR:       buildEmitIR,
		EmitJSON:     buildEmitJSON,
		EmitBindings: buildEmitBnd,
		OutName:      buildOutName,
		UpdateDeps:   buildDepsUpd,
		Vendor:       buildVendor,
		GCSections:   buildGC,
		SizeDelta:    buildSizeDlt,
		LTO:          buildLTO,
	}
}

// printSize 在开 `--emit-size-delta` 时打印体积表：产物字节数 + 每个模块目标文件。
func printSize(res *build.Result) {
	raw, ok := res.Artifacts["size"]
	if !ok {
		return
	}
	size, ok := raw.(map[string]any)
	if !ok {
		return
	}
	if n, ok := size["exe_bytes"].(int64); ok {
		fmt.Fprintf(stdout, "体积：产物 %d 字节\n", n)
	}
	objs, ok := size["objects_bytes"].(map[string]int64)
	if !ok {
		return
	}
	names := make([]string, 0, len(objs))
	for name := range objs {
		names = append(names, name)
	}
	sort.Strings(names)
	var total int64
	for _, name := range names {
		fmt.Fprintf(stdout, "  %-28s %8d\n", name, objs[name])
		total += objs[name]
	}
	fmt.Fprintf(stdout, "  %-28s %8d\n", "（目标文件合计）", total)
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
			printSize(res)
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
			// `run [路径] -- 参数...`：`--` 之前是工程路径，之后全部透传给程序
			var runArgs []string
			var rootArg string
			if n := cmd.ArgsLenAtDash(); n >= 0 {
				rootArg = argOrDot(args[:n])
				runArgs = args[n:]
			} else {
				rootArg = argOrDot(args)
			}
			root := findRoot(rootArg)
			res := build.Build(root, loadOpts())
			if err := reportDiags(res); err != nil {
				return err
			}
			// 嵌入式产物是芯片镜像，宿主没法执行——给出烧录提示而不是报运行失败
			if targets.Parse(res.Triple).Embedded {
				esptoolChip := targets.Parse(res.Triple).CPUDerived()
				if bl, _ := res.Artifacts["bootloader"].(string); bl != "" {
					bootOff, _ := res.Artifacts["boot_offset"].(string)
					appOff, _ := res.Artifacts["app_offset"].(string)
					fmt.Fprintf(stderr, "`%s` 是芯片 flash 镜像，无法在宿主执行。\n烧录：esptool --chip %s write_flash %s %s %s %s\n",
						res.Triple, esptoolChip, bootOff, bl, appOff, res.Exe)
				} else {
					fmt.Fprintf(stderr, "`%s` 是芯片 flash 镜像，无法在宿主执行。\n烧录：esptool --chip %s write_flash 0x0 %s\n", res.Triple, esptoolChip, res.Exe)
				}
				return &exitError{ExitUsage}
			}
			proc := exec.Command(res.Exe, append(passArgs, runArgs...)...)
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
