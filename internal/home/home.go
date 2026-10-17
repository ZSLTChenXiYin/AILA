// Package home 定位 AILA_HOME 与随包相对目录（docs/specs/0003 §1）。
//
// 一条铁律：管理器只认相对目录。可执行文件所在目录 = AILA_HOME，
// 运行时源码、标准库、随包 clang/llvm 全部按相对路径找——解压即用。
package home

import (
	"os"
	"path/filepath"
	"runtime"
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

// DataDir 是共享数据表（错误码 / target 表）所在目录。
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

// ToolchainRoot 是随包工具链根目录。
func ToolchainRoot() string { return filepath.Join(Root(), "toolchain") }

// BundleDir 是某个三元组的 bundle 目录。
func BundleDir(triple string) string { return filepath.Join(ToolchainRoot(), triple) }

// BundleBin 是 bundle 的 bin 目录。
func BundleBin(triple string) string { return filepath.Join(BundleDir(triple), "bin") }

// ExeSuffix 是平台可执行文件后缀。
func ExeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// BundleBinNames 是随包工具链里被认可的工具名。
var BundleBinNames = []string{
	"clang", "clang-cl", "llc", "ld.lld", "lld", "lld-link",
	"llvm-ar", "llvm-objcopy", "llvm-objdump", "llvm-nm", "llvm-strip",
}

// BundleBins 列出某个三元组 bundle 里真实存在的工具（name → 绝对路径）。
func BundleBins(triple string) map[string]string {
	out := map[string]string{}
	dir := BundleBin(triple)
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return out
	}
	for _, name := range BundleBinNames {
		for _, cand := range []string{name + ExeSuffix(), name} {
			p := filepath.Join(dir, cand)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				out[name] = p
				break
			}
		}
	}
	return out
}

// BundleLLVMVersion 读随包 VERSION 文件（CI 写入），没有就返回空串。
func BundleLLVMVersion(triple string) string {
	b, err := os.ReadFile(filepath.Join(BundleDir(triple), "VERSION"))
	if err != nil {
		return ""
	}
	line := ""
	for _, r := range string(b) {
		if r == '\n' || r == '\r' {
			break
		}
		line += string(r)
	}
	return trimSpace(line)
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// InstalledTriples 列出已安装的 bundle 目录。
func InstalledTriples() []string {
	entries, err := os.ReadDir(ToolchainRoot())
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if st, err := os.Stat(filepath.Join(ToolchainRoot(), e.Name(), "bin")); err == nil && st.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out
}
