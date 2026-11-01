// Package home 定位 AILA_HOME 与随包相对目录（docs/specs/0003 §1）。
//
// 一条铁律：管理器只认相对目录。可执行文件所在目录 = AILA_HOME，
// 运行时源码、标准库、随包 clang/llvm 全部按相对路径找——解压即用。
package home

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// HomeEnv 是显式覆盖 AILA_HOME 的环境变量（CI / 测试 / 便携部署用）。
const HomeEnv = "AILA_HOME"

var override string

// Override 强制指定 AILA_HOME（测试用）；传空串恢复默认行为。
func Override(root string) { override = root }

// Root 解析顺序：Override / AILA_HOME → 可执行文件同级 → 上溯找仓库布局 → 当前目录。
func Root() string {
	if override != "" {
		return abs(override)
	}
	if env := os.Getenv(HomeEnv); env != "" {
		return abs(env)
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		if hasLayout(dir) {
			return dir
		}
		// go run / go test：可执行文件在临时目录，向上找仓库根
		if root := walkUp(dir, 6); root != "" {
			return root
		}
	}
	if wd, err := os.Getwd(); err == nil {
		if root := walkUp(wd, 6); root != "" {
			return root
		}
		return wd
	}
	return "."
}

func abs(p string) string {
	if a, err := filepath.Abs(p); err == nil {
		return a
	}
	return p
}

func walkUp(dir string, levels int) string {
	cur := dir
	for i := 0; i < levels; i++ {
		if hasLayout(cur) {
			return cur
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return ""
}

func hasLayout(root string) bool {
	probes := []string{
		filepath.Join(root, "share", "aila", "data"),
		filepath.Join(root, "share", "aila", "stdlib"),
	}
	for _, p := range probes {
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return true
		}
	}
	// 仓库布局：有 src/aila 或 stdlib/ 也算
	for _, p := range []string{filepath.Join(root, "src", "aila"), filepath.Join(root, "stdlib")} {
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return true
		}
	}
	return false
}

// ShareRoot 返回 share/aila（打包布局）或仓库根（开发布局）。
func ShareRoot() string {
	cand := filepath.Join(Root(), "share", "aila")
	if st, err := os.Stat(cand); err == nil && st.IsDir() {
		return cand
	}
	return Root()
}

// ProductsRoot 是**工作产物**根目录：编译输出、随包工具链、各种缓存都落在这里，
// 源码树因此保持干净（`.local/` 整体进 .gitignore）。布局：
//
//	.local/bin/            管理器二进制（go build -o .local/bin/aila.exe）
//	.local/build/<工程>/   该工程的 .ll/.o/可执行体与缓存
//	.local/toolchains/     `aila toolchain install` 装的东西 + 安装状态
//
// 分发出去的 bundle 仍然用 `toolchains/`（随包只读），两者由 ToolchainRoots 一并扫描。
func ProductsRoot() string {
	return filepath.Join(Root(), ".local")
}

// BinDir 是管理器二进制建议落点。
func BinDir() string {
	return filepath.Join(ProductsRoot(), "bin")
}

// ProductsRootFor 返回某个工程该用的产物根：
//   - 工程在 AILA_HOME 内（仓库里的 examples/*）→ `<AILA_HOME>/.local`，全仓产物集中一处；
//   - 工程在 AILA_HOME 外（独立项目）→ 工程自己的 `.local`，不往别人家目录写东西。
func ProductsRootFor(projectRoot string) string {
	if rel, err := filepath.Rel(Root(), projectRoot); err == nil && insideRel(rel) {
		return ProductsRoot()
	}
	return filepath.Join(projectRoot, ".local")
}

// BuildDirFor 返回某个工程（以工程根目录标识）的产物目录：
// `<工程产物根>/build/<在 AILA_HOME 内的相对路径>`；工程不在 AILA_HOME 内时相对部分为空。
func BuildDirFor(projectRoot string) string {
	rel := ""
	if r, err := filepath.Rel(Root(), projectRoot); err == nil && insideRel(r) && r != "." {
		rel = r
	}
	return filepath.Join(ProductsRootFor(projectRoot), "build", rel)
}

// insideRel 判断相对路径是否落在基准目录内（不是 `..` 开头的外部路径）。
func insideRel(rel string) bool {
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// DataDir 是共享数据表（错误码 / 编译器全局配置）所在目录。
func DataDir() string {
	for _, cand := range []string{
		filepath.Join(ShareRoot(), "data"),
		filepath.Join(Root(), "share", "aila", "data"),
	} {
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand
		}
	}
	return filepath.Join(ShareRoot(), "data")
}

// RuntimeSrc 返回语言运行时 C 源码路径。
func RuntimeSrc() string {
	for _, cand := range []string{
		filepath.Join(ShareRoot(), "runtime", "aila_rt.c"),
		filepath.Join(Root(), "runtime", "aila_rt.c"),
	} {
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
	}
	return filepath.Join(Root(), "runtime", "aila_rt.c")
}

// PlatformRuntime 返回平台 runtime 目录（startup 汇编 / 链接脚本 / 平台 C）。
// 布局：share/aila/runtime/<家族>/（打包）或 runtime/<家族>/（仓库布局）。
// 返回空串 = 该平台家族未随包适配。
func PlatformRuntime(family string) string {
	if family == "" {
		return ""
	}
	for _, cand := range []string{
		filepath.Join(ShareRoot(), "runtime", family),
		filepath.Join(Root(), "runtime", family),
	} {
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand
		}
	}
	return ""
}

// StdlibDir 返回标准库目录。
func StdlibDir() string {
	for _, cand := range []string{
		filepath.Join(ShareRoot(), "stdlib"),
		filepath.Join(Root(), "stdlib"),
	} {
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand
		}
	}
	return filepath.Join(Root(), "stdlib")
}

// ToolchainsRoot 是**安装**根目录（docs/rfcs/0006 §3）：`.local/toolchains/`。
//
// 布局：<纯数据目录>/{clang/<ver>/, esp-clang/<ver>/, gcc/<triple>/<ver>/}。
// 工具按「家族 + 版本」组织，目录名即事实，能力运行时探测。
func ToolchainsRoot() string { return filepath.Join(ProductsRoot(), "toolchains") }

// BundledToolchainsRoot 是随包只读工具链根（分发布局 `toolchains/`，AILA_HOME 下）。
func BundledToolchainsRoot() string { return filepath.Join(Root(), "toolchains") }

// ToolchainRoots 返回要扫描的工具链根目录：先本地产物（可写），再随包只读。
func ToolchainRoots() []string {
	roots := []string{ToolchainsRoot()}
	bundled := BundledToolchainsRoot()
	if a, errA := filepath.Abs(bundled); errA == nil {
		if b, errB := filepath.Abs(roots[0]); errB != nil || a != b {
			roots = append(roots, bundled)
		}
	} else {
		roots = append(roots, bundled)
	}
	return roots
}

// ExeSuffix 是平台可执行文件后缀。
func ExeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// ClangInstall 是一个 clang 系家族的某版本安装。
type ClangInstall struct {
	Family  string // "clang"（通版）| "esp-clang"（厂商 fork）| ...
	Version string // 目录名即版本
	Dir     string // 绝对路径
}

// GCCInstall 是一个目标三元组的 GCC 安装（gcc/<triple>/<version>/）。
type GCCInstall struct {
	Triple  string // 目录名即 triple（不带 -gcc 后缀）
	Version string
	Dir     string
}

// InstalledClangs 扫各工具链根下除 gcc 外的所有家族目录；
// 版本子目录必须含 bin/（子目录名 = 版本）。同名家族+版本只取第一个。
func InstalledClangs() []ClangInstall {
	var out []ClangInstall
	seen := map[string]bool{}
	for _, root := range ToolchainRoots() {
		fams, err := readDirs(root)
		if err != nil {
			continue
		}
		for _, fam := range fams {
			if fam == "gcc" {
				continue
			}
			vers, err := readDirs(filepath.Join(root, fam))
			if err != nil {
				continue
			}
			for _, v := range vers {
				if seen[fam+"/"+v] {
					continue
				}
				dir := filepath.Join(root, fam, v)
				if st, err := os.Stat(filepath.Join(dir, "bin")); err == nil && st.IsDir() {
					seen[fam+"/"+v] = true
					out = append(out, ClangInstall{Family: fam, Version: v, Dir: dir})
				}
			}
		}
	}
	return out
}

// InstalledGCCs 扫各工具链根下的 gcc/<triple>/<version>/。
func InstalledGCCs() []GCCInstall {
	var out []GCCInstall
	for _, root := range ToolchainRoots() {
		out = append(out, installedGCCsIn(filepath.Join(root, "gcc"))...)
	}
	return out
}

func installedGCCsIn(root string) []GCCInstall {
	var out []GCCInstall
	triples, err := readDirs(root)
	if err != nil {
		return nil
	}
	for _, triple := range triples {
		vers, err := readDirs(filepath.Join(ToolchainsRoot(), "gcc", triple))
		if err != nil {
			continue
		}
		for _, v := range vers {
			dir := filepath.Join(root, triple, v)
			if st, err := os.Stat(filepath.Join(dir, "bin")); err == nil && st.IsDir() {
				out = append(out, GCCInstall{Triple: triple, Version: v, Dir: dir})
			}
		}
	}
	return out
}

// Bins 列出一个安装目录 bin/ 下的全部可执行文件（name → 绝对路径）。
// 不设白名单：能力靠探测，目录里有什么工具就是什么工具。
func Bins(dir string) map[string]string {
	out := map[string]string{}
	binDir := filepath.Join(dir, "bin")
	entries, err := os.ReadDir(binDir)
	if err != nil {
		return out
	}
	suffix := ExeSuffix()
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if suffix != "" && strings.HasSuffix(name, suffix) {
			name = strings.TrimSuffix(name, suffix)
		}
		if _, exists := out[name]; !exists {
			out[name] = filepath.Join(binDir, e.Name())
		}
	}
	return out
}

func readDirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}
