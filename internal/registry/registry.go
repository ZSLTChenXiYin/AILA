// Package registry 管理 toolchains/toolchains.json（docs/rfcs/0006 §4）。
//
// 注册表只承担「下载清单」职责，不记能力——能力一律运行时探测
// （clang --print-targets、gcc 目录名即 triple）。两段各司其职：
//
//	installed：aila 本地写：已装什么、在哪（toolchain list 读它）
//	available：CI 随包内置：能下载什么（install 读它）
package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"aila/internal/home"
)

// Entry 是一条可下载的工具链清单。
type Entry struct {
	Family  string `json:"family"`           // clang / esp-clang / gcc / sdk / ...
	Triple  string `json:"triple,omitempty"` // gcc / sdk 系才有：目录名即 triple / SDK 注册名
	Version string `json:"version"`          // 安装后的目录名
	URL     string `json:"url,omitempty"`    // 为空 = 只能本地 zip 安装（专有许可）
	SHA256  string `json:"sha256,omitempty"` // 为空 = 下载时不校验（CI 待填）
	Size    string `json:"size,omitempty"`   // 人读的体积提示
	// Keep 非空时解包只保留这些顶层条目（SDK 裁剪 examples/documents 等）
	Keep []string `json:"keep,omitempty"`
}

// InstalledEntry 是一条已安装记录。
type InstalledEntry struct {
	Version string `json:"version"`
	Path    string `json:"path"` // 相对 AILA_HOME 的斜杠路径
}

// File 是注册表的内存形态。
type File struct {
	Installed map[string]InstalledEntry `json:"installed"`
	Available []Entry                   `json:"available,omitempty"`
	// VersionBands 是 per-family 的 LLVM 版本带（rfcs/0006 §5）；缺省回落全局带。
	VersionBands map[string][]string `json:"version_bands,omitempty"`
}

// ManifestPath 是「能下载什么」清单的位置：随包只读（分发布局的 toolchains/，
// 仓库里就是源码树那份），不随安装动作变化。
func ManifestPath() string {
	cands := []string{
		filepath.Join(home.BundledToolchainsRoot(), "toolchains.json"),
		filepath.Join(home.ToolchainsRoot(), "toolchains.json"),
	}
	for _, p := range cands {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return cands[0]
}

// StatePath 是「装了什么」的位置：本地产物目录 `.local/toolchains/installed.json`。
// 与清单分开 = 代码（清单）与产物（安装状态 + 载荷）互不污染。
func StatePath() string { return filepath.Join(home.ToolchainsRoot(), "installed.json") }

// Path 是旧版单文件位置（兼容读取用）。
func Path() string { return filepath.Join(home.ToolchainsRoot(), "toolchains.json") }

// Key 家族键：gcc / sdk 系 = "<family>/<triple>"（目录即 triple / SDK 注册名），
// 其余 = 家族名。
func Key(family, triple string) string {
	if family == "gcc" || family == "sdk" {
		return family + "/" + triple
	}
	return family
}

// Load 读清单与安装状态；两者都缺或损坏 = 空注册表（不报错，能力全靠探测）。
func Load() *File {
	f := empty()
	// 清单：available（随包只读）
	if raw, err := os.ReadFile(ManifestPath()); err == nil {
		var m File
		if json.Unmarshal(raw, &m) == nil {
			f.Available = m.Available
			f.VersionBands = m.VersionBands
			// 旧布局把 installed 写在清单里，仍认
			for k, v := range m.Installed {
				f.Installed[k] = v
			}
		}
	}
	// 状态：installed（本地产物）
	if raw, err := os.ReadFile(StatePath()); err == nil {
		var s File
		if json.Unmarshal(raw, &s) == nil {
			for k, v := range s.Installed {
				f.Installed[k] = v
			}
		}
	}
	return f
}

func empty() *File {
	return &File{Installed: map[string]InstalledEntry{}}
}

// Save 只写安装状态（清单随包只读，安装动作不碰源码树）。
func (f *File) Save() error {
	if err := os.MkdirAll(home.ToolchainsRoot(), 0o755); err != nil {
		return err
	}
	state := File{Installed: f.Installed}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(StatePath(), append(data, '\n'), 0o644)
}

// FindAvailable 按家族键（或 "family@version"）查 available 清单。
func (f *File) FindAvailable(spec string) (Entry, bool) {
	for _, e := range f.Available {
		if Key(e.Family, e.Triple) == spec {
			return e, true
		}
		if spec == e.Family+"@"+e.Version || spec == Key(e.Family, e.Triple)+"@"+e.Version {
			return e, true
		}
	}
	// 宽松匹配：不带版本时，同名家族取最新（按版本号）
	if !strings.Contains(spec, "@") {
		var best *Entry
		for i := range f.Available {
			e := &f.Available[i]
			if Key(e.Family, e.Triple) != spec {
				continue
			}
			if best == nil || versionGe(e.Version, best.Version) {
				best = e
			}
		}
		if best != nil {
			return *best, true
		}
	}
	return Entry{}, false
}

// RelPath 把绝对目录转成相对 AILA_HOME 的斜杠路径（记录进 installed）。
func RelPath(dir string) string {
	rel, err := filepath.Rel(home.Root(), dir)
	if err != nil {
		return dir
	}
	return filepath.ToSlash(rel)
}

// versionGe 粗粒度版本比较（目录名即版本，形如 23.1.3）。
func versionGe(a, b string) bool {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		na, ea := atoi(as[i])
		nb, eb := atoi(bs[i])
		if ea && eb && na != nb {
			return na > nb
		}
		if as[i] != bs[i] {
			return as[i] > bs[i]
		}
	}
	return len(as) >= len(bs)
}

func atoi(s string) (int, bool) {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
	}
	return n, len(s) > 0
}
