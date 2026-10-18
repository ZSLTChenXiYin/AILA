// Package cli 是 `aila` 的命令行外壳（docs/specs/0001 §3）。
//
// G1 范围：与前端无关的命令全部可用（version / doctor / toolchain / target / explain / new / clean）；
// 需要编译能力的命令（check / build / run / bindgen）在 G2/G3 移植完成后接入。
package cli

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ailang/aila/internal/diag"
	"github.com/ailang/aila/internal/home"
	"github.com/ailang/aila/internal/manifest"
	"github.com/ailang/aila/internal/resolve"
	"github.com/ailang/aila/internal/targets"
	"github.com/ailang/aila/internal/toolchain"
)

// 退出码（与 Python 版一致，docs/specs/0001 §3.1）。
const (
	ExitOK       = 0
	ExitDiag     = 1
	ExitUsage    = 2
	ExitEnv      = 3
	ExitInternal = 4
)

// Version 由构建时注入（-ldflags "-X ..."）。
var Version = "0.1.0-dev"

type rootFlags struct {
	json  bool
	color string
}

var flags rootFlags

// stdout / stderr 可被测试替换（CLI 内部一律写这两个变量，不直接碰 os.Stdout）。
var (
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "aila",
		Short:         "AILA 管理器（Go 版）",
		Long:          "AILA 是面向嵌入式、以区域内存模型与模块级热替换为核心的编程语言。\n这是用 Go 重写的管理器与编译器前端（rfcs/0003；G1–G3 已完成，G4 进行中）。",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().BoolVar(&flags.json, "json", false, "机器可读输出")
	root.PersistentFlags().StringVar(&flags.color, "color", "auto", "auto|always|never")
	root.PersistentFlags().Bool("quiet", false, "少输出")
	root.PersistentFlags().Bool("verbose", false, "多输出")
	root.SetVersionTemplate("aila {{.Version}}\n")
	root.Version = Version
	root.SetOut(stdout)
	root.SetErr(stderr)

		checkCmd, buildCmd, runCmd := newCheckCmd(), newBuildCmd(), newRunCmd()
	for _, c := range []*cobra.Command{checkCmd, buildCmd, runCmd} {
		addBuildFlags(c)
	}
	root.AddCommand(
		newVersionCmd(), newDoctorCmd(), newToolchainCmd(), newTargetCmd(),
		newExplainCmd(), newNewCmd(), newCleanCmd(), newIRCmd(),
		checkCmd, buildCmd, runCmd, newDepsCmd(),
		newBindgenCmd(),
		newParseCmd(),
	)
	return root
}

// Execute 是入口；返回进程退出码。
func Execute() int {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		if ee, ok := err.(*exitError); ok {
			return ee.code
		}
		fmt.Fprintln(stderr, "aila 内部错误（这是管理器的缺陷）：", err)
		return ExitInternal
	}
	return ExitOK
}

type exitError struct{ code int }

func (e *exitError) Error() string { return fmt.Sprintf("exit %d", e.code) }

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	switch flags.color {
	case "always":
		return true
	case "never":
		return false
	}
	st, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice != 0
}

func printDiags(bag *diag.Bag) {
	color := colorEnabled()
	if s := bag.Render(color); s != "" {
		fmt.Fprintln(stderr, s)
	}
	if s := bag.Summary(color); s != "" {
		fmt.Fprintln(stderr, s)
	}
}

// --------------------------------------------------------------------------- //
// version
// --------------------------------------------------------------------------- //

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示版本、AILA_HOME 与工具链来源",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(stdout, "aila %s (go)\n", Version)
			fmt.Fprintf(stdout, "  AILA_HOME：%s\n", home.Root())
			tools := resolve.ResolveAll("", "")
			fmt.Fprintf(stdout, "  LLVM 后端：%s\n", tools[resolve.RoleIR2Obj].Describe())
			fmt.Fprintf(stdout, "  C 编译器：%s\n", tools[resolve.RoleCC].Describe())
			return nil
		},
	}
}

// --------------------------------------------------------------------------- //
// doctor
// --------------------------------------------------------------------------- //

func newDoctorCmd() *cobra.Command {
	var cc, llvm string
	cmd := &cobra.Command{
		Use:   "doctor [路径]",
		Short: "环境自检（含链接冒烟测试）",
		RunE: func(cmd *cobra.Command, args []string) error {
			checks := toolchain.Doctor(Version, cc, llvm)
			failed := toolchain.Failed(checks)
			if flags.json {
				type row struct {
					Name   string `json:"name"`
					Level  string `json:"level"`
					Detail string `json:"detail"`
					Fix    string `json:"fix"`
				}
				rows := make([]row, 0, len(checks))
				for _, c := range checks {
					rows = append(rows, row{c.Name, c.Level, c.Detail, c.Fix})
				}
				payload := map[string]any{"checks": rows, "ok": failed == 0}
				out, _ := json.MarshalIndent(payload, "", "  ")
				fmt.Fprintln(stdout, string(out))
				if failed > 0 {
					return &exitError{ExitEnv}
				}
				return nil
			}
			fmt.Fprintln(stdout, "AILA 环境自检")
			fmt.Fprintln(stdout, strings.Repeat("─", 60))
			for _, c := range checks {
				mark := map[string]string{"ok": "✔", "warn": "⚠", "fail": "✘"}[c.Level]
				fmt.Fprintf(stdout, " %s %-22s %s\n", mark, c.Name, c.Detail)
				if c.Level != "ok" && c.Fix != "" {
					kind := "修复"
					if c.Level == "warn" {
						kind = "建议"
					}
					fmt.Fprintf(stdout, "   → %s：%s\n", kind, c.Fix)
				}
			}
			fmt.Fprintln(stdout, strings.Repeat("─", 60))
			if failed > 0 {
				fmt.Fprintf(stdout, "%d 项未通过。按上面的「修复」逐条处理后再跑 `aila doctor`。\n", failed)
				return &exitError{ExitEnv}
			}
			fmt.Fprintln(stdout, "可以 `aila doctor` 之外的编译命令了。")
			return nil
		},
	}
	cmd.Flags().StringVar(&cc, "cc", "", "临时覆盖清单的 C 编译器选择")
	cmd.Flags().StringVar(&llvm, "llvm", "", "临时覆盖清单的 LLVM 后端选择")
	return cmd
}

// --------------------------------------------------------------------------- //
// toolchain
// --------------------------------------------------------------------------- //

func newToolchainCmd() *cobra.Command {
	parent := &cobra.Command{
		Use:   "toolchain",
		Short: "随包工具链：查看 / 安装",
	}
	parent.AddCommand(
		&cobra.Command{
			Use:   "info",
			Short: "打印工具链解析结果",
			RunE: func(cmd *cobra.Command, args []string) error {
				lo, hi := resolve.VersionBand()
				fmt.Fprintln(stdout, "AILA 工具链解析")
				fmt.Fprintln(stdout, strings.Repeat("─", 60))
				fmt.Fprintf(stdout, "AILA_HOME          : %s\n", home.Root())
				fmt.Fprintf(stdout, "随包工具链根       : %s\n", home.ToolchainRoot())
				fmt.Fprintf(stdout, "宿主三元组         : %s\n", resolve.HostTripleGuess())
				fmt.Fprintf(stdout, "绑定 LLVM 版本带   : %s..%s\n", lo, hi)
				fmt.Fprintln(stdout, strings.Repeat("─", 60))
				tools := resolve.ResolveAll("", "")
				roles := []string{resolve.RoleCC, resolve.RoleIR2Obj, resolve.RoleClang}
				for _, r := range roles {
					fmt.Fprintf(stdout, "%-10s %s\n", r, tools[r].Describe())
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "列出已安装的随包工具链",
			RunE: func(cmd *cobra.Command, args []string) error {
				root := home.ToolchainRoot()
				triples := home.InstalledTriples()
				if len(triples) == 0 {
					fmt.Fprintf(stdout, "没有随包工具链目录：%s\n", root)
					fmt.Fprintln(stdout, "下载 AILA release 后运行 `aila toolchain install <zip>`")
					return nil
				}
				fmt.Fprintf(stdout, "随包工具链目录：%s\n", root)
				sort.Strings(triples)
				for _, t := range triples {
					bins := home.BundleBins(t)
					version := home.BundleLLVMVersion(t)
					label := "(无 VERSION)"
					if version != "" {
						label = "LLVM " + version
					}
					fmt.Fprintf(stdout, "  %-32s %s  %d 个可执行文件\n", t, label, len(bins))
				}
				return nil
			},
		},
	)
	var triple string
	var force bool
	install := &cobra.Command{
		Use:   "install <release.zip>",
		Short: "把官方 release zip 解到 toolchain/<triple>/",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return installBundle(args[0], triple, force)
		},
	}
	install.Flags().StringVar(&triple, "triple", "", "安装到指定三元组目录")
	install.Flags().BoolVar(&force, "force", false, "覆盖已有目录")
	parent.AddCommand(install)
	return parent
}

func installBundle(archive, triple string, force bool) error {
	if _, err := os.Stat(archive); err != nil {
		fmt.Fprintf(stderr, "找不到压缩包：%s\n", archive)
		return &exitError{ExitUsage}
	}
	if triple == "" {
		triple = resolve.HostTripleGuess()
	}
	dest := home.BundleDir(triple)
	if entries, err := os.ReadDir(dest); err == nil && len(entries) > 0 && !force {
		fmt.Fprintf(stderr, "目标已存在：%s（用 --force 覆盖）\n", dest)
		return &exitError{ExitUsage}
	}
	zr, err := zip.OpenReader(archive)
	if err != nil {
		bag := &diag.Bag{}
		bag.Error("E7013", diag.Empty, "解压失败："+err.Error(), archive, "确认拿到的是 AILA 官方 release zip")
		printDiags(bag)
		return &exitError{ExitDiag}
	}
	defer zr.Close()
	strip := singleRoot(zr)
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "/") {
			continue
		}
		rel := f.Name
		if strip != "" {
			rel = strings.TrimPrefix(rel, strip+"/")
		}
		if rel == "" || strings.HasPrefix(rel, "..") {
			continue
		}
		out := filepath.Join(dest, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		if err := extractOne(f, out); err != nil {
			bag := &diag.Bag{}
			bag.Error("E7013", diag.Empty, "解压失败："+err.Error(), f.Name, "")
			printDiags(bag)
			return &exitError{ExitDiag}
		}
	}
	bins := home.BundleBins(triple)
	names := make([]string, 0, len(bins))
	for n := range bins {
		names = append(names, n)
	}
	sort.Strings(names)
	if _, ok := bins["clang"]; !ok {
		bag := &diag.Bag{}
		bag.Error("E7012", diag.Empty, "随包工具链不完整，缺少：clang", dest, "重新下载完整 release 包")
		printDiags(bag)
		return &exitError{ExitDiag}
	}
	fmt.Fprintf(stdout, "已安装随包工具链：%s\n", dest)
	fmt.Fprintf(stdout, "  可执行文件：%s\n", strings.Join(names, ", "))
	if v := home.BundleLLVMVersion(triple); v != "" {
		lo, hi := resolve.VersionBand()
		fmt.Fprintf(stdout, "  LLVM 版本：%s（绑定版本带 %s..%s）\n", v, lo, hi)
	}
	fmt.Fprintln(stdout, "接着跑 `aila doctor` 复核。")
	return nil
}

func singleRoot(zr *zip.ReadCloser) string {
	roots := map[string]bool{}
	total := 0
	for _, f := range zr.File {
		name := strings.TrimSpace(f.Name)
		if name == "" {
			continue
		}
		total++
		roots[strings.SplitN(name, "/", 2)[0]] = true
	}
	if len(roots) != 1 || total == 0 {
		return ""
	}
	for r := range roots {
		// 只有当所有成员都在同一个根目录下时才剥层
		for _, f := range zr.File {
			if !strings.HasPrefix(f.Name, r+"/") {
				return ""
			}
		}
		return r
	}
	return ""
}

func extractOne(f *zip.File, out string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	mode := f.Mode()
	if mode&0o111 != 0 {
		mode = 0o755
	} else {
		mode = 0o644
	}
	dst, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, rc)
	return err
}

// --------------------------------------------------------------------------- //
// target
// --------------------------------------------------------------------------- //

func newTargetCmd() *cobra.Command {
	parent := &cobra.Command{Use: "target", Short: "目标与地形表"}
	parent.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "列出内置 board 表",
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := targets.Names()
			if err != nil {
				return err
			}
			fmt.Fprintln(stdout, "内置 board 表：")
			for _, n := range names {
				b, _ := targets.Find(n)
				triple := "(宿主默认)"
				if b.Triple != nil {
					triple = *b.Triple
				}
				domains := make([]string, 0, len(b.Domains))
				for d := range b.Domains {
					domains = append(domains, d)
				}
				sort.Strings(domains)
				fmt.Fprintf(stdout, "  %-10s arch=%-8s triple=%-24s 域=%s\n", n, b.Arch, triple, strings.Join(domains, ", "))
			}
			return nil
		},
	})
	parent.AddCommand(&cobra.Command{
		Use:   "show <board>",
		Short: "显示某个 board 的细节",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, ok := targets.Find(args[0])
			if !ok {
				names, _ := targets.Names()
				fmt.Fprintf(stderr, "未知 board：%s\n候选：%s\n", args[0], strings.Join(names, ", "))
				return &exitError{ExitUsage}
			}
			fmt.Fprintf(stdout, "board      : %s\n", args[0])
			fmt.Fprintf(stdout, "arch       : %s\n", b.Arch)
			triple := "(宿主默认三元组)"
			if b.Triple != nil {
				triple = *b.Triple
			}
			fmt.Fprintf(stdout, "triple     : %s\n", triple)
			embedded := "否"
			if b.Embedded {
				embedded = "是"
			}
			fmt.Fprintf(stdout, "嵌入式     : %s\n", embedded)
			fmt.Fprintln(stdout, "预置物理域 :")
			names := make([]string, 0, len(b.Domains))
			for d := range b.Domains {
				names = append(names, d)
			}
			sort.Strings(names)
			for _, d := range names {
				fmt.Fprintf(stdout, "  %-10s kind=%s\n", d, b.Domains[d])
			}
			return nil
		},
	})
	return parent
}

// --------------------------------------------------------------------------- //
// explain
// --------------------------------------------------------------------------- //

func newExplainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "explain <CODE>",
		Short: "解释错误码并给出修法",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			code := args[0]
			entry, ok := diag.Find(code)
			if flags.json {
				payload := map[string]any{"code": code, "found": ok}
				if ok {
					payload["title"] = entry.Title
					payload["explain"] = entry.Explain
					payload["fix"] = entry.Fix
					payload["doc"] = entry.Doc
				}
				out, _ := json.MarshalIndent(payload, "", "  ")
				fmt.Fprintln(stdout, string(out))
				if !ok {
					return &exitError{ExitUsage}
				}
				return nil
			}
			fmt.Fprintln(stdout, diag.FormatExplain(code))
			if !ok {
				return &exitError{ExitUsage}
			}
			return nil
		},
	}
}

// --------------------------------------------------------------------------- //
// new / clean / pending
// --------------------------------------------------------------------------- //

const manifestTemplate = `[package]
name = "%s"
version = "0.1.0"
entry = "_.main"

[target]
board = "%s"

[build]
opt = 0
warnings = "default"
`

const mainTemplate = `// %s — AILA 入口模块（模块名 = 文件名 = main）
//
// 三条命令跑起来：aila doctor → aila run
import std.io;

func main() {
    io.print_text("hello, AILA!");
    io.print_nl();
}
`

const readmeTemplate = `# %s

AILA 工程（项目 = 包、子包 = 目录、模块 = 文件）。

- 包根只放入口模块 ` + "`main.aila`" + `；
- 其余模块放子包目录，导入写 ` + "`import _.子包.模块;`" + `；
- C 生态放 ` + "`cpacks/<名字>/`" + `，导入写 ` + "`import ch _.cpacks.<名字>;`" + `。
`

func newNewCmd() *cobra.Command {
	var name, board string
	cmd := &cobra.Command{
		Use:   "new <目录>",
		Short: "新建工程骨架",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			if entries, err := os.ReadDir(path); err == nil && len(entries) > 0 {
				bag := &diag.Bag{}
				bag.Error("E7011", diag.Empty, "目标目录非空："+path, "", "换一个路径，或先清空该目录")
				printDiags(bag)
				return &exitError{ExitDiag}
			}
			if name == "" {
				name = filepath.Base(path)
			}
			if _, ok := targets.Find(board); !ok {
				fmt.Fprintf(stderr, "未知 board：%s（候选：aila target list）\n", board)
				return &exitError{ExitUsage}
			}
			if err := os.MkdirAll(path, 0o755); err != nil {
				return err
			}
			writeIf := func(rel, content string) error {
				return os.WriteFile(filepath.Join(path, rel), []byte(content), 0o644)
			}
			if err := writeIf(manifest.FileName, fmt.Sprintf(manifestTemplate, name, board)); err != nil {
				return err
			}
			if err := writeIf("main.aila", fmt.Sprintf(mainTemplate, name)); err != nil {
				return err
			}
			if err := writeIf("README.md", fmt.Sprintf(readmeTemplate, name)); err != nil {
				return err
			}
			fmt.Fprintf(stdout, "已创建 AILA 工程：%s\n", path)
			fmt.Fprintln(stdout, "接下来：")
			fmt.Fprintf(stdout, "  cd %s\n", args[0])
			fmt.Fprintln(stdout, "  aila doctor      # 环境自检")
			fmt.Fprintln(stdout, "  aila run         # 编译并运行")
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "包名（默认取目录名）")
	cmd.Flags().StringVar(&board, "board", "host", "目标 board")
	return cmd
}

func newCleanCmd() *cobra.Command {
	var cacheOnly bool
	cmd := &cobra.Command{
		Use:   "clean [路径]",
		Short: "清理产物",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot(argOrDot(args))
			target := filepath.Join(root, "build")
			if cacheOnly {
				target = filepath.Join(target, ".cache")
			}
			if err := os.RemoveAll(target); err != nil {
				return err
			}
			if cacheOnly {
				fmt.Fprintln(stdout, "已清理 构建缓存")
			} else {
				fmt.Fprintln(stdout, "已清理 build/ 目录")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&cacheOnly, "cache-only", false, "只删构建缓存")
	return cmd
}

func newPendingCmd(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short + "（尚未移植到 Go 批次）",
		RunE: func(cmd *cobra.Command, args []string) error {
			bag := &diag.Bag{}
			bag.Error("E9003", diag.Empty,
				"命令 `"+use+"` 需要编译前端，Go 移植尚未到该批次",
				"当前批次 G1 只含管理器命令",
				"过渡期请用仓库里的 Python 版：`./aila`（Windows: `aila.cmd`）")
			printDiags(bag)
			return &exitError{ExitDiag}
		},
	}
}

// --------------------------------------------------------------------------- //
// helpers
// --------------------------------------------------------------------------- //

func argOrDot(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return "."
}

// findRoot 从路径向上找含 aila.toml 的目录。
func findRoot(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	if st, err := os.Stat(abs); err == nil && !st.IsDir() {
		abs = filepath.Dir(abs)
	}
	cur := abs
	for {
		if _, err := os.Stat(filepath.Join(cur, manifest.FileName)); err == nil {
			return cur
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return abs
		}
		cur = parent
	}
}
