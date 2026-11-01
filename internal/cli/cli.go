// Package cli 是 `aila` 的命令行外壳（docs/specs/0001 §3）。
//
// G1 范围：与前端无关的命令全部可用（version / doctor / toolchain / target / explain / new / clean）；
// 需要编译能力的命令（check / build / run / bindgen）在 G2/G3 移植完成后接入。
package cli

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"aila/internal/diag"
	"aila/internal/home"
	"aila/internal/manifest"
	"aila/internal/registry"
	"aila/internal/resolve"
	"aila/internal/targets"
	"aila/internal/toolchain"
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
		checkCmd, buildCmd, runCmd, newDepsCmd(), newLSPCmd(),
		newBindgenCmd(), newParseCmd(),
	)
	return root
}

// RootCommand 暴露命令树（`tools/skillgen` 用它生成 CLI 参考文档，保证文档与实现同步）。
func RootCommand() *cobra.Command { return newRootCmd() }

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
			tools := resolve.ResolveAll("", "", "")
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
				fmt.Fprintf(stdout, "工具链根           : %s\n", home.ToolchainsRoot())
				fmt.Fprintf(stdout, "宿主三元组         : %s\n", resolve.HostTriple())
				fmt.Fprintf(stdout, "绑定 LLVM 版本带   : %s..%s\n", lo, hi)
				fmt.Fprintln(stdout, strings.Repeat("─", 60))
				tools := resolve.ResolveAll("", "", "")
				roles := []string{resolve.RoleCC, resolve.RoleIR2Obj, resolve.RoleClang}
				for _, r := range roles {
					fmt.Fprintf(stdout, "%-10s %s\n", r, tools[r].Describe())
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "列出已安装的工具链家族与可下载清单",
			RunE: func(cmd *cobra.Command, args []string) error {
				reg := registry.Load()
				clangs := home.InstalledClangs()
				gccs := home.InstalledGCCs()
				if len(clangs) == 0 && len(gccs) == 0 {
					fmt.Fprintf(stdout, "没有已安装的工具链：%s\n", home.ToolchainsRoot())
					fmt.Fprintln(stdout, "运行 `aila toolchain install <family>`（联网）或 `aila toolchain install <zip>`（离线）")
					return nil
				}
				fmt.Fprintf(stdout, "工具链目录：%s\n", home.ToolchainsRoot())
				for _, c := range clangs {
					fmt.Fprintf(stdout, "  clang 系  %-12s %-10s %d 个可执行文件\n",
						c.Family, c.Version, len(home.Bins(c.Dir)))
				}
				for _, g := range gccs {
					fmt.Fprintf(stdout, "  gcc       %-12s %-10s %d 个可执行文件（目录名 = 目标三元组）\n",
						g.Triple, g.Version, len(home.Bins(g.Dir)))
				}
				if len(reg.Available) > 0 {
					fmt.Fprintln(stdout, "可下载（available）：")
					for _, e := range reg.Available {
						key := registry.Key(e.Family, e.Triple)
						size := ""
						if e.Size != "" {
							size = "  " + e.Size
						}
						fmt.Fprintf(stdout, "  %-34s %-10s → aila toolchain install %s@%s%s\n",
							key, e.Version, key, e.Version, size)
					}
				}
				return nil
			},
		},
	)
	var family, version, triple, mirror string
	var force bool
	install := &cobra.Command{
		Use:   "install <family[@version] | gcc/<triple>[@version] | sdk/<名称>[@版本] | 本地zip>",
		Short: "安装工具链：按 available 清单联网下载，或本地 zip 离线安装",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return installToolchain(args[0], family, version, triple, mirror, force)
		},
	}
	install.Flags().StringVar(&family, "family", "", "本地 zip 安装时指定家族（clang / esp-clang / gcc）")
	install.Flags().StringVar(&version, "version", "", "本地 zip 安装时指定版本（默认 zip 内 VERSION 文件或文件名）")
	install.Flags().StringVar(&triple, "triple", "", "gcc 家族本地 zip 安装时指定目标三元组")
	// GitHub 直连不通时用镜像前缀（与 ci/fetch-llvm.* 的 -Mirror 同一用法）：
	// 可逗号分隔多个，按序尝试；也可用环境变量 AILA_TOOLCHAIN_MIRROR 兜底。
	install.Flags().StringVar(&mirror, "mirror", "", "下载镜像前缀（镜像会拼在原 URL 前，可逗号分隔多个）")
	install.Flags().BoolVar(&force, "force", false, "覆盖已有目录")
	parent.AddCommand(install)
	return parent
}

// installToolchain 统一入口：本地 zip 存在 → 离线安装；否则按 available 清单联网下载。
func installToolchain(spec, family, version, triple, mirror string, force bool) error {
	if st, err := os.Stat(spec); err == nil && !st.IsDir() {
		return installLocalZip(spec, family, version, triple, force)
	}
	reg := registry.Load()
	entry, ok := reg.FindAvailable(spec)
	if !ok {
		fmt.Fprintf(stderr, "available 清单里没有 `%s`。已收录的：\n", spec)
		for _, e := range reg.Available {
			fmt.Fprintf(stderr, "  %s@%s\n", registry.Key(e.Family, e.Triple), e.Version)
		}
		fmt.Fprintln(stderr, "也可以直接给本地 zip：`aila toolchain install <zip>`")
		return &exitError{ExitUsage}
	}
	if entry.Version != "" && strings.Contains(spec, "@") {
		// family@version 显式指定了版本
	} else if v := versionOf(spec); v != "" && v != entry.Version {
		entry.Version = v
	}
	return downloadAndInstall(entry, mirror, force)
}

func versionOf(spec string) string {
	if idx := strings.LastIndex(spec, "@"); idx >= 0 {
		return spec[idx+1:]
	}
	return ""
}

// downloadAndInstall 联网下载 available 条目 → sha256 校验 → 解压 → 写注册表。
//
// GitHub 直连经常不通（尤其国内），镜像前缀会拼在原 URL 前：
//
//	aila toolchain install clang@23.1.3 --mirror https://gh-proxy.com
//	AILA_TOOLCHAIN_MIRROR=https://gh-proxy.com aila toolchain install clang
//
// 多个镜像用逗号分隔，按序尝试；直连排最后兜底。每个候选各试两轮，镜像的反向
// 代理前面常有一段时间在拉上游（首字节慢），所以只限制「等响应头」的时长，
// 不限制整体下载时长，并且边下边报进度。
func downloadAndInstall(entry registry.Entry, mirror string, force bool) error {
	if entry.URL == "" {
		fmt.Fprintf(stderr, "`%s` 在清单里没有下载 URL，请用本地 zip 安装：`aila toolchain install <zip>`\n",
			registry.Key(entry.Family, entry.Triple))
		return &exitError{ExitUsage}
	}
	cands := downloadCandidates(entry.URL, mirror)
	tmp, err := os.CreateTemp("", "aila-toolchain-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	var n int64
	var lastErr error
	for i, url := range cands {
		for attempt := 1; attempt <= 2; attempt++ {
			if len(cands) > 1 {
				fmt.Fprintf(stdout, "下载 %s（%s，候选 %d/%d，第 %d 次尝试）\n",
					url, sizeHint(entry.Size), i+1, len(cands), attempt)
			} else {
				fmt.Fprintf(stdout, "下载 %s（%s）\n", url, sizeHint(entry.Size))
			}
			n, lastErr = fetchTo(tmp, url)
			if lastErr == nil {
				goto done
			}
			fmt.Fprintf(stderr, "下载失败：%v\n", lastErr)
		}
	}
	return &exitError{ExitEnv}
done:
	tmp.Seek(0, io.SeekStart)
	fmt.Fprintf(stdout, "已下载 %d 字节，校验 sha256…\n", n)
	sum, err := sha256Hex(tmp)
	if err != nil {
		return err
	}
	if entry.SHA256 != "" && !strings.EqualFold(sum, entry.SHA256) {
		fmt.Fprintf(stderr, "sha256 不符：期望 %s，实得 %s\n", entry.SHA256, sum)
		return &exitError{ExitEnv}
	}
	fmt.Fprintln(stdout, "校验通过，解压安装…")
	return installFromReader(tmp, entry, force)
}

// downloadCandidates 拼出候选 URL 列表：镜像（按序）→ 原始直连。
func downloadCandidates(orig, mirror string) []string {
	if mirror == "" {
		mirror = os.Getenv("AILA_TOOLCHAIN_MIRROR")
	}
	var out []string
	if strings.HasPrefix(orig, "https://github.com/") {
		for _, m := range strings.Split(mirror, ",") {
			m = strings.TrimSpace(m)
			if m == "" {
				continue
			}
			out = append(out, strings.TrimRight(m, "/")+"/"+orig)
		}
	}
	return append(out, orig)
}

// fetchTo 把 url 下载进 f（覆盖写）。返回写入字节数。
func fetchTo(f *os.File, url string) (int64, error) {
	if err := f.Truncate(0); err != nil {
		return 0, err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	cli := &http.Client{
		// 镜像反代要先回源，首字节可能很慢：只卡响应头，不卡总量。
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ResponseHeaderTimeout: 90 * time.Second,
			TLSHandshakeTimeout:   30 * time.Second,
			DisableCompression:    true, // zip 本身已压缩
		},
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "aila-toolchain/1 (https://github.com/ailang/aila)")
	resp, err := cli.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP %d（%s）", resp.StatusCode, url)
	}
	cw := &countingWriter{w: f, every: 32 << 20}
	n, err := io.Copy(cw, resp.Body)
	cw.tick(true)
	return n, err
}

// countingWriter 每写满 every 字节打一行进度（大包下载时让人知道没卡死）。
type countingWriter struct {
	w     io.Writer
	n     int64
	every int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	c.tick(false)
	return n, err
}

func (c *countingWriter) tick(force bool) {
	if c.every <= 0 {
		return
	}
	if force || c.n%c.every < int64(len(""))+c.every/64 {
		fmt.Fprintf(stdout, "  …已下载 %d MB\n", c.n>>20)
	}
}

func sizeHint(s string) string {
	if s == "" {
		return "体积见清单"
	}
	return s
}

// installLocalZip 离线安装：sha256 命中 available → 自动取 family/version/triple；
// 否则用 flags 兜底。
func installLocalZip(archive, family, version, triple string, force bool) error {
	f, err := os.Open(archive)
	if err != nil {
		fmt.Fprintf(stderr, "打不开压缩包：%s\n", archive)
		return &exitError{ExitUsage}
	}
	defer f.Close()
	sum, err := sha256Hex(f)
	if err != nil {
		return err
	}
	f.Seek(0, io.SeekStart)
	entry := registry.Entry{Family: family, Triple: triple, Version: version}
	if entry.Family == "" || ((entry.Family == "gcc" || entry.Family == "sdk") && entry.Triple == "") || entry.Version == "" {
		// 用 sha256 反查 available 清单
		for _, e := range registry.Load().Available {
			if e.SHA256 != "" && strings.EqualFold(e.SHA256, sum) {
				if entry.Family == "" {
					entry.Family = e.Family
				}
				if entry.Triple == "" {
					entry.Triple = e.Triple
				}
				if entry.Version == "" {
					entry.Version = e.Version
				}
				entry.Keep = e.Keep // 裁剪白名单跟着 available 条目走
				break
			}
		}
	}
	if entry.Version == "" {
		// zip 内 VERSION 文件兜底
		if v := zipVersionFile(f.Name()); v != "" {
			entry.Version = v
		}
	}
	if entry.Family == "" || entry.Version == "" || ((entry.Family == "gcc" || entry.Family == "sdk") && entry.Triple == "") {
		fmt.Fprintln(stderr, "无法确定家族/版本/三元组：sha256 不在 available 清单里，请补 --family / --version / --triple")
		return &exitError{ExitUsage}
	}
	fmt.Fprintf(stdout, "sha256: %s → 安装 %s %s\n", sum, registry.Key(entry.Family, entry.Triple), entry.Version)
	return installFromReader(f, entry, force)
}

// zipVersionFile 从 zip 里读 VERSION 文件首行（不落盘）。
func zipVersionFile(path string) string {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return ""
	}
	defer zr.Close()
	for _, f := range zr.File {
		if strings.EqualFold(filepath.Base(f.Name), "VERSION") {
			rc, err := f.Open()
			if err != nil {
				return ""
			}
			defer rc.Close()
			b, err := io.ReadAll(io.LimitReader(rc, 64))
			if err != nil {
				return ""
			}
			return strings.TrimSpace(string(b))
		}
	}
	return ""
}

// installFromReader 把 zip 流解压到 toolchains/<家族目录> 并写注册表。
func installFromReader(r io.ReaderAt, entry registry.Entry, force bool) error {
	dest := ""
	switch entry.Family {
	case "gcc", "sdk":
		dest = filepath.Join(home.ToolchainsRoot(), entry.Family, entry.Triple, entry.Version)
	default:
		dest = filepath.Join(home.ToolchainsRoot(), entry.Family, entry.Version)
	}
	if entries, err := os.ReadDir(dest); err == nil && len(entries) > 0 && !force {
		fmt.Fprintf(stderr, "目标已存在：%s（用 --force 覆盖）\n", dest)
		return &exitError{ExitUsage}
	}
	size, err := fileSize(r)
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(r, size)
	if err != nil {
		bag := &diag.Bag{}
		bag.Error("E7013", diag.Empty, "解压失败："+err.Error(), "确认拿到的是 zip 格式工具链包", "")
		printDiags(bag)
		return &exitError{ExitDiag}
	}
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
		// Keep 白名单裁剪（SDK 只留 include/lib/ld/bin，跳过 examples 等）
		if len(entry.Keep) > 0 && !topIn(rel, entry.Keep) {
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
	// 最小完整性检查：clang 系要有 clang，gcc 系要有目标驱动，sdk 系要有库目录
	if entry.Family == "sdk" {
		if entries, err := os.ReadDir(filepath.Join(dest, "lib")); err != nil || len(entries) == 0 {
			bag := &diag.Bag{}
			bag.Error("E7012", diag.Empty, "SDK 包不完整：没有 lib/ 目录", dest, "确认包内容与家族一致")
			printDiags(bag)
			return &exitError{ExitDiag}
		}
	} else {
		bins := home.Bins(dest)
		need := "clang"
		if entry.Family == "gcc" {
			need = entry.Triple + "-gcc"
			if _, ok := bins[need]; !ok {
				need = "gcc"
			}
		}
		if _, ok := bins[need]; !ok {
			bag := &diag.Bag{}
			bag.Error("E7012", diag.Empty, "工具链不完整，缺少："+need, dest, "确认包内容与家族一致")
			printDiags(bag)
			return &exitError{ExitDiag}
		}
	}
	// 写注册表 installed 段
	reg := registry.Load()
	key := registry.Key(entry.Family, entry.Triple)
	reg.Installed[key] = registry.InstalledEntry{Version: entry.Version, Path: registry.RelPath(dest)}
	if err := reg.Save(); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "已安装：%s\n", dest)
	fmt.Fprintln(stdout, "接着跑 `aila doctor` 复核。")
	return nil
}

func topIn(rel string, keep []string) bool {
	head := strings.SplitN(rel, "/", 2)[0]
	for _, k := range keep {
		if head == k {
			return true
		}
	}
	return false
}

func sha256Hex(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// fileSize 兼容 *os.File 与一般 ReaderAt。
func fileSize(r io.ReaderAt) (int64, error) {
	switch v := r.(type) {
	case *os.File:
		st, err := v.Stat()
		if err != nil {
			return 0, err
		}
		return st.Size(), nil
	}
	return 0, fmt.Errorf("无法确定流大小")
}

func singleRoot(zr *zip.Reader) string {
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
	parent := &cobra.Command{Use: "target", Short: "目标三元组与已装工具覆盖"}
	parent.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "概览已装工具链覆盖的架构与宿主三元组",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(stdout, "宿主三元组：%s\n", resolve.HostTriple())
			clangs := home.InstalledClangs()
			if len(clangs) > 0 {
				fmt.Fprintln(stdout, "已装 clang 系覆盖的架构（--print-targets 自报）：")
				for _, c := range clangs {
					archs := []string{}
					if p, ok := home.Bins(c.Dir)["clang"]; ok {
						for _, a := range resolve.RegisteredTargets(p, false) {
							archs = append(archs, a)
						}
					}
					fmt.Fprintf(stdout, "  %-12s %-8s %s\n", c.Family, c.Version, strings.Join(archs, " "))
				}
			}
			gccs := home.InstalledGCCs()
			if len(gccs) > 0 {
				fmt.Fprintln(stdout, "已装 gcc（目录名即目标三元组）：")
				for _, g := range gccs {
					fmt.Fprintf(stdout, "  %-12s %-8s gcc/%s\n", g.Triple, g.Version, g.Triple)
				}
			}
			if len(clangs) == 0 && len(gccs) == 0 {
				fmt.Fprintf(stdout, "没有已装工具链：%s\n", home.ToolchainsRoot())
			}
			return nil
		},
	})
	parent.AddCommand(&cobra.Command{
		Use:   "show <triple>",
		Short: "拆解一个三元组并探测哪些已装工具能服务它",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			triple := args[0]
			if !targets.Valid(triple) {
				fmt.Fprintf(stderr, "不是合法三元组：%s（至少两段、段非空，如 xtensa-esp8266-elf）\n", triple)
				fmt.Fprintln(stderr, "已装工具覆盖的架构见 `aila target list`")
				return &exitError{ExitUsage}
			}
			t := targets.Parse(triple)
			fmt.Fprintf(stdout, "triple     : %s\n", t.Triple)
			fmt.Fprintf(stdout, "arch       : %s\n", t.Arch)
			if cpu := t.CPUDerived(); cpu != "" {
				fmt.Fprintf(stdout, "cpu 候选   : %s（triple 第二段；可用 [target] cpu 显式覆盖）\n", cpu)
			} else {
				fmt.Fprintln(stdout, "cpu 候选   : 无（triple 第二段是平台词）")
			}
			embedded := "否（宿主语义：链接出可执行文件）"
			if t.Embedded {
				embedded = "是（freestanding：构建到 .o，交叉链接待 M2）"
			}
			fmt.Fprintf(stdout, "嵌入式     : %s\n", embedded)
			fmt.Fprintln(stdout, strings.Repeat("─", 60))
			llvm := resolve.ResolveLLVMFor(triple, "")
			cc := resolve.ResolveCC(triple, "")
			fmt.Fprintf(stdout, "ir2obj     : %s\n", llvm.Describe())
			fmt.Fprintf(stdout, "cc         : %s\n", cc.Describe())
			if !llvm.OK() && !t.Embedded {
				fmt.Fprintln(stdout, "提示：宿主目标建议用默认（不写 [target]）")
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

[build]
opt = 0
warnings = "default"
`

const mainTemplate = `// %s — AILA 入口模块（模块名 = 文件名 = main）
//
// 三条命令跑起来：aila doctor → aila run
import std.io;

func main() {
    io.println(io.FmtVal.T { v: "hello, AILA!" });
}
`

const readmeTemplate = `# %s

AILA 工程（项目 = 包、子包 = 目录、模块 = 文件）。

- 包根只放入口模块 ` + "`main.aila`" + `；
- 其余模块放子包目录，导入写 ` + "`import _.子包.模块;`" + `；
- C 生态放 ` + "`cpacks/<名字>/`" + `，导入写 ` + "`import ch _.cpacks.<名字>;`" + `。
`

func newNewCmd() *cobra.Command {
	var name, triple string
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
			targetSection := ""
			if triple != "" {
				if !targets.Valid(triple) {
					fmt.Fprintf(stderr, "不是合法三元组：%s（至少两段、段非空，如 xtensa-esp8266-elf）\n", triple)
					return &exitError{ExitUsage}
				}
				targetSection = "\n[target]\ntriple = \"" + triple + "\"\n"
			}
			if err := os.MkdirAll(path, 0o755); err != nil {
				return err
			}
			writeIf := func(rel, content string) error {
				return os.WriteFile(filepath.Join(path, rel), []byte(content), 0o644)
			}
			if err := writeIf(manifest.FileName, fmt.Sprintf(manifestTemplate, name)+targetSection); err != nil {
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
	cmd.Flags().StringVar(&triple, "triple", "", "目标三元组（写入 [target]；不给 = 宿主）")
	return cmd
}

func newCleanCmd() *cobra.Command {
	var cacheOnly bool
	cmd := &cobra.Command{
		Use:   "clean [路径]",
		Short: "清理产物",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot(argOrDot(args))
			target := home.BuildDirFor(root)
			if cacheOnly {
				target = filepath.Join(target, ".cache")
			}
			if err := os.RemoveAll(target); err != nil {
				return err
			}
			if cacheOnly {
				fmt.Fprintln(stdout, "已清理 构建缓存")
			} else {
				fmt.Fprintf(stdout, "已清理产物目录 %s\n", target)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&cacheOnly, "cache-only", false, "只删构建缓存")
	return cmd
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
