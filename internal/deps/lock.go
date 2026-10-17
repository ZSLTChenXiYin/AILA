package deps

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/ailang/aila/internal/diag"
)

// LockFileName 是锁文件名（放在包根，与 aila.toml 同级）。
const LockFileName = "aila.lock"

// LockFile 是 `aila.lock` 的内容：每条依赖的来源、版本、内容哈希与目录。
type LockFile struct {
	Deps map[string]LockEntry `toml:"dep"`
}

// LockEntry 是一条锁记录。
type LockEntry struct {
	Source  string `toml:"source"`
	Version string `toml:"version"`
	Hash    string `toml:"hash"`
	Dir     string `toml:"dir,omitempty"`
}

// LoadLock 读锁文件；不存在返回空锁（不是错误）。
func LoadLock(root string) (*LockFile, error) {
	path := filepath.Join(root, LockFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &LockFile{Deps: map[string]LockEntry{}}, nil
		}
		return nil, err
	}
	var lf LockFile
	if _, err := toml.Decode(string(data), &lf); err != nil {
		return nil, fmt.Errorf("aila.lock 不是合法 TOML：%v", err)
	}
	if lf.Deps == nil {
		lf.Deps = map[string]LockEntry{}
	}
	return &lf, nil
}

// Write 把当前解析结果写回锁文件（调用方决定何时写：`--deps-update` 或首次解析）。
func (g *Graph) Write() error {
	lf := &LockFile{Deps: map[string]LockEntry{}}
	names := make([]string, 0, len(g.ByPkg))
	for name := range g.ByPkg {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		res := g.ByPkg[name]
		// 锁里的来源用「清单怎么写就怎么记」的形态（可移植，不含本机绝对路径）
		lf.Deps[name] = LockEntry{Source: res.Source, Version: res.Version, Hash: res.Hash}
	}
	var b strings.Builder
	b.WriteString("# 由 aila 自动生成，请勿手改（改依赖请编辑 aila.toml 后跑 `aila deps --update`）\n")
	if err := toml.NewEncoder(&b).Encode(lf); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(g.Root, LockFileName), []byte(b.String()), 0o666)
}

// Locked 判断锁文件是否已与当前图一致（用于决定要不要提示 `--update`）。
func (g *Graph) Locked() bool {
	if g.Lock == nil || len(g.Lock.Deps) != len(g.ByPkg) {
		return false
	}
	for name, res := range g.ByPkg {
		ent, ok := g.Lock.Deps[name]
		if !ok || ent.Hash != res.Hash {
			return false
		}
	}
	return true
}

// Vendor 把解析好的依赖复制到 `<root>/vendor/<名字>/`（离线构建用）。
func (g *Graph) Vendor() error {
	for name, res := range g.ByPkg {
		dst := filepath.Join(g.Root, "vendor", name)
		if err := os.RemoveAll(dst); err != nil {
			return err
		}
		if err := copyTree(res.Dir, dst); err != nil {
			return err
		}
	}
	return nil
}

// copyTree 复制目录（只复制 AILA 源码与清单，跳过 build/ 等产物目录）。
func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := info.Name()
			if rel != "." && (base == "build" || base == ".git" || strings.HasPrefix(base, ".")) {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, rel), 0o777)
		}
		if !strings.HasSuffix(info.Name(), ".aila") && info.Name() != "aila.toml" && info.Name() != "aila.lock" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dst, rel)), 0o777); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dst, rel), data, 0o666)
	})
}

// Describe 生成人读的依赖图（`aila deps` 用）。
func (g *Graph) Describe() string {
	var b strings.Builder
	for i, pkg := range g.Order {
		if pkg == RootPkg {
			fmt.Fprintf(&b, "根包  %s\n", g.Root)
			continue
		}
		res := g.ByPkg[pkg]
		tag := ""
		if res.Vendored {
			tag = "  (vendor)"
		}
		fmt.Fprintf(&b, "  %d. %-16s %-10s %-34s %s%s\n", i, pkg, res.Version, res.Source, res.Hash, tag)
	}
	return b.String()
}

var _ = diag.Empty
