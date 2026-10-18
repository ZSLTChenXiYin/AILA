// Package deps 解析跨项目依赖（docs/rfcs/0004：包 = 项目）。
//
// 清单写法（`aila.toml`）：
//
//	[dependencies]
//	mylib = { path = "../mylib" }                      # 本地路径
//	util  = { git = "https://...", tag = "v1.2.0" }     # git（tag 或 rev 锁定）
//
// 解析产物：每个依赖解析成一个目录，加上版本与内容哈希；结果写进 `aila.lock`。
// 判据（离线可复现）：
//   - 锁文件里记了哈希，依赖内容变了就报错，除非显式 `--deps-update`；
//   - `vendor/<名字>/` 存在时优先用它（嵌入式要离线构建）；
//   - git 依赖缓存在 `<AILA_HOME>/cache/deps/`，无网络时用缓存。
package deps

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"aila/internal/diag"
	"aila/internal/manifest"
)

// Source 是依赖的取值方式。
type Source string

const (
	SourcePath Source = "path" // 本地相对目录
	SourceGit  Source = "git"  // git 仓库（tag / rev 锁定）
)

// Dep 是清单里的一条依赖声明。
type Dep struct {
	Name    string
	Source  Source
	Path    string // SourcePath：相对包根的路径
	Git     string // SourceGit：仓库地址（https / ssh / 本地路径）
	Tag     string
	Rev     string
	Version string // 可选：期望的版本前缀（"0.1" 匹配 "0.1.3"）
}

// Resolved 是一个已定位的依赖。
type Resolved struct {
	Name    string
	Dir     string // 包根（含 aila.toml）
	Version string
	Source  string // "path:../mylib" / "git:https://...@v1.2.0" / "vendor"
	Hash    string // 内容哈希（sha256，前 16 位十六进制）
	Vendored bool
}

// Graph 是整个依赖图（含传递依赖）。
type Graph struct {
	Root   string // 根包目录
	RootPkg string // 根包的包名键（源码里写 `_`）
	// ByPkg 是「包名 → 解析结果」。根包的名字用清单里的 package.name，
	// 但模块名根始终是 `_`（源码里的 `import _` 只指根包）。
	ByPkg map[string]*Resolved
	// Edges 是「包名 → 该包直接依赖的包名列表」（含根包）。
	Edges map[string][]string
	// Order 是按依赖拓扑排好的包名（根包在最后，保证依赖先编译）。
	Order []string
	// Lock 是本次实际使用的锁内容。
	Lock *LockFile
}

// DepsOf 返回某个包的直接依赖名。
func (g *Graph) DepsOf(pkg string) []string {
	if g == nil {
		return nil
	}
	return g.Edges[pkg]
}

// PackageDir 返回包根目录（根包用 g.Root）。
func (g *Graph) PackageDir(pkg string) string {
	if pkg == g.RootPkg {
		return g.Root
	}
	if r, ok := g.ByPkg[pkg]; ok {
		return r.Dir
	}
	return ""
}

// 根包在 Edges/ByPkg 里的键（源码里写成 `_`）。
const RootPkg = "_"

// --------------------------------------------------------------------------- //
// 解析依赖声明
// --------------------------------------------------------------------------- //

// FromManifest 把清单里的 `[dependencies]` 转成 Dep 列表（校验语义）。
func FromManifest(root string, m *manifest.Manifest, bag *diag.Bag) []Dep {
	out := make([]Dep, 0, len(m.Dependencies))
	names := make([]string, 0, len(m.Dependencies))
	for name := range m.Dependencies {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		raw := m.Dependencies[name]
		d := Dep{Name: name, Version: strings.TrimSpace(raw.Version)}
		switch {
		case strings.TrimSpace(raw.Path) != "" && strings.TrimSpace(raw.Git) != "":
			bag.Error("E7021", diag.Empty, "依赖 `"+name+"` 同时写了 `path` 与 `git`", "只能二选一", "")
			continue
		case strings.TrimSpace(raw.Path) != "":
			d.Source = SourcePath
			d.Path = filepath.Clean(strings.TrimSpace(raw.Path))
		case strings.TrimSpace(raw.Git) != "":
			d.Source = SourceGit
			d.Git = strings.TrimSpace(raw.Git)
			d.Tag = strings.TrimSpace(raw.Tag)
			d.Rev = strings.TrimSpace(raw.Rev)
			if d.Tag == "" && d.Rev == "" {
				bag.Error("E7021", diag.Empty, "git 依赖 `"+name+"` 必须锁 `tag` 或 `rev`",
					"可复现构建", "分支会漂移；写 `tag = \"v1.0.0\"` 或 `rev = \"<sha>\"`")
				continue
			}
			if d.Tag != "" && d.Rev != "" {
				bag.Error("E7021", diag.Empty, "git 依赖 `"+name+"` 的 `tag` 与 `rev` 只能写一个", "这里", "")
				continue
			}
		default:
			bag.Error("E7021", diag.Empty, "依赖 `"+name+"` 既没写 `path` 也没写 `git`", "这里",
				"写 `{ path = \"../x\" }` 或 `{ git = \"...\", tag = \"v1.0.0\" }`")
			continue
		}
		out = append(out, d)
	}
	return out
}

// --------------------------------------------------------------------------- //
// 解析图
// --------------------------------------------------------------------------- //

// Resolve 从根包出发解析整张依赖图（含传递依赖、环检测、版本与哈希校验）。
func Resolve(root string, update bool, bag *diag.Bag) *Graph {
	g := &Graph{Root: root, RootPkg: RootPkg, ByPkg: map[string]*Resolved{}, Edges: map[string][]string{}}
	lock, lockErr := LoadLock(root)
	if lockErr != nil && !update {
		bag.Warning("W7021", diag.Empty, "锁文件读不了，按清单重新解析", lockErr.Error(), "")
	}
	if update || lockErr != nil {
		lock = &LockFile{Deps: map[string]LockEntry{}}
	}
	g.Lock = lock

	rootManifest := manifest.Load(root, bag)
	if rootManifest == nil {
		return g
	}
	g.RootPkg = RootPkg

	// 广度优先：根包 → 直接依赖 → 传递依赖
	type queueItem struct {
		pkg string // 源码里使用的包名根（`pkg`）
		dir string // 包根目录
	}
	seen := map[string]bool{}
	queue := []queueItem{{pkg: RootPkg, dir: root}}
	byName := map[string]string{} // 包名 → 包根（防止同一包两个来源）
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		man := rootManifest
		if item.pkg != RootPkg {
			man = manifest.Load(item.dir, bag)
			if man == nil {
				continue
			}
		}
		declared := FromManifest(item.dir, man, bag)
		edges := make([]string, 0, len(declared))
		for _, d := range declared {
			res, ok := g.resolveOne(root, item.dir, d, update, bag)
			if !ok {
				continue
			}
			edges = append(edges, d.Name)
			if prev, dup := byName[d.Name]; dup && prev != res.Dir {
				bag.Error("E7024", diag.Empty, "依赖 `"+d.Name+"` 出现在两个不同目录",
					"同名依赖只能有一个来源", prev+" / "+res.Dir)
				continue
			}
			byName[d.Name] = res.Dir
			if seen[d.Name] {
				continue
			}
			seen[d.Name] = true
			g.ByPkg[d.Name] = res
			queue = append(queue, queueItem{pkg: d.Name, dir: res.Dir})
		}
		g.Edges[item.pkg] = edges
	}

	// 拓扑序（依赖在前）；环检测
	visited := map[string]bool{}
	onStack := map[string]bool{}
	var order []string
	var visit func(pkg string)
	visit = func(pkg string) {
		if visited[pkg] {
			return
		}
		if onStack[pkg] {
			bag.Error("E7025", diag.Empty, "依赖成环：`"+pkg+"` 又绕回自己", "依赖图必须是有向无环图",
				"拆掉环，或把公共部分抽成第三个包")
			return
		}
		onStack[pkg] = true
		for _, dep := range g.Edges[pkg] {
			visit(dep)
		}
		onStack[pkg] = false
		visited[pkg] = true
		order = append(order, pkg)
	}
	visit(RootPkg)
	g.Order = order
	return g
}

// resolveOne 把一条依赖声明定位成目录，并与锁文件比对。
//
// 语义（docs/rfcs/0004 §4）：
//   - **目录永远由清单决定**（vendor 优先，其次 path/git 缓存）——锁文件不参与定位，
//     这样 `aila.lock` 可以随仓库走、不绑定机器绝对路径；
//   - 锁文件只用来**校验**：来源串、版本、内容哈希三者任一不同即 E7023，
//     要求显式 `aila deps --update`（构建可复现的前提）。
func (g *Graph) resolveOne(root, fromDir string, d Dep, update bool, bag *diag.Bag) (*Resolved, bool) {
	// 1) vendor 目录优先（离线构建）
	if vdir := filepath.Join(root, "vendor", d.Name); dirExists(vdir) {
		if !fileExists(filepath.Join(vdir, manifest.FileName)) {
			bag.Error("E7022", diag.Empty, "`vendor/"+d.Name+"` 里没有 `aila.toml`", vdir, "")
			return nil, false
		}
		res := &Resolved{Name: d.Name, Dir: vdir, Source: "vendor", Vendored: true}
		fillVersionHash(res, bag)
		if !update {
			if ent, ok := g.Lock.Deps[d.Name]; ok {
				if ent.Hash != "" && ent.Hash != res.Hash {
					bag.Error("E7023", diag.Empty, "`vendor/"+d.Name+"` 的内容与 `aila.lock` 不一致",
						"锁：" + ent.Hash + "；vendor：" + res.Hash,
						"确认 vendor 无误后跑 `aila deps --update` 重新锁定")
					return nil, false
				}
			}
		}
		return res, true
	}
	// 2) 按声明定位目录
	var dir string
	switch d.Source {
	case SourcePath:
		dir = filepath.Join(fromDir, d.Path)
		if !fileExists(filepath.Join(dir, manifest.FileName)) {
			bag.Error("E7022", diag.Empty, "依赖 `"+d.Name+"` 的路径里没有 `aila.toml`",
				dir, "path 指向包根（不是 src/ 之类子目录）")
			return nil, false
		}
	case SourceGit:
		var ok bool
		dir, ok = gitFetch(d, bag)
		if !ok {
			return nil, false
		}
	}
	res := &Resolved{Name: d.Name, Dir: dir, Source: describeDep(d)}
	fillVersionHash(res, bag)
	if res.Version == "" {
		bag.Error("E7022", diag.Empty, "依赖 `"+d.Name+"` 的清单缺少 `package.version`", dir, "")
		return nil, false
	}
	if d.Version != "" && !versionMatches(res.Version, d.Version) {
		bag.Error("E7023", diag.Empty, "依赖 `"+d.Name+"` 的版本是 `"+res.Version+
			"`，不满足清单要求 `"+d.Version+"`", "版本不匹配", "改清单，或升级依赖")
		return nil, false
	}
	// 3) 与锁文件比对（来源 / 版本 / 内容哈希）
	if !update {
		if ent, ok := g.Lock.Deps[d.Name]; ok {
			switch {
			case ent.Source != "" && ent.Source != res.Source:
				bag.Error("E7023", diag.Empty, "依赖 `"+d.Name+"` 的来源变了",
					"清单："+res.Source+"；锁："+ent.Source,
					"确认无误后跑 `aila deps --update` 重新锁定")
				return nil, false
			case ent.Version != "" && ent.Version != res.Version:
				bag.Error("E7023", diag.Empty, "依赖 `"+d.Name+"` 的版本变了",
					"清单解析："+res.Version+"；锁："+ent.Version,
					"确认无误后跑 `aila deps --update`")
				return nil, false
			case ent.Hash != "" && ent.Hash != res.Hash:
				bag.Error("E7023", diag.Empty, "依赖 `"+d.Name+"` 的内容变了（哈希不一致）",
					"锁："+ent.Hash+"；现在："+res.Hash,
					"确认改动无误后跑 `aila deps --update`")
				return nil, false
			}
		}
	}
	return res, true
}
// fillVersionHash 读依赖清单的版本并算内容哈希。
func fillVersionHash(res *Resolved, bag *diag.Bag) {
	sub := &diag.Bag{}
	m := manifest.Load(res.Dir, sub)
	if m != nil {
		res.Version = m.Version
	}
	res.Hash = dirHash(res.Dir)
}

// sameSource 判断锁里的来源描述是否与当前声明一致。
func sameSource(locked string, d Dep) bool { return locked == describeDep(d) }

// describeDep 生成可读且稳定的来源描述。
func describeDep(d Dep) string {
	switch d.Source {
	case SourcePath:
		return "path:" + d.Path
	case SourceGit:
		if d.Tag != "" {
			return "git:" + d.Git + "@" + d.Tag
		}
		return "git:" + d.Git + "@" + d.Rev
	}
	return "unknown"
}

// lockDir 还原锁里记录的目录（path 依赖记绝对路径，git 依赖记缓存目录）。
func lockDir(root string, ent LockEntry) string {
	if strings.HasPrefix(ent.Source, "path:") {
		p := strings.TrimPrefix(ent.Source, "path:")
		if filepath.IsAbs(p) {
			if fileExists(filepath.Join(p, manifest.FileName)) {
				return p
			}
			return ""
		}
		p2 := filepath.Join(root, p)
		if fileExists(filepath.Join(p2, manifest.FileName)) {
			return p2
		}
		return ""
	}
	if ent.Dir != "" && fileExists(filepath.Join(ent.Dir, manifest.FileName)) {
		return ent.Dir
	}
	return ""
}

// versionMatches 判断版本是否满足前缀要求（"0.1" 匹配 "0.1.3"；空要求恒成立）。
func versionMatches(have, want string) bool {
	want = strings.TrimSpace(want)
	if want == "" {
		return true
	}
	if have == want {
		return true
	}
	return strings.HasPrefix(have, want+".")
}

// --------------------------------------------------------------------------- //
// git 依赖
// --------------------------------------------------------------------------- //

// gitFetch 把 git 依赖取到缓存目录并返回该目录。
func gitFetch(d Dep, bag *diag.Bag) (string, bool) {
	cacheRoot := gitCacheRoot()
	key := shortHash(d.Git + "@" + d.Tag + d.Rev)
	dir := filepath.Join(cacheRoot, d.Name+"-"+key)
	// 命中缓存就不需要 git：离线机器、CI 复用、vendor 之前都能用
	if fileExists(filepath.Join(dir, manifest.FileName)) {
		return dir, true
	}
	if _, err := exec.LookPath("git"); err != nil {
		bag.Error("E7022", diag.Empty, "依赖 `"+d.Name+"` 是 git 来源，但本机没有 `git`",
			"PATH 里找不到 git",
			"装 git，或用 `aila deps --vendor` 先把依赖取下来（vendor/ 可随仓库分发）")
		return "", false
	}
	if err := os.MkdirAll(cacheRoot, 0o777); err != nil {
		bag.Error("E7022", diag.Empty, "建不了依赖缓存目录", err.Error(), "")
		return "", false
	}
	args := []string{"clone", "--depth", "1"}
	if d.Tag != "" {
		args = append(args, "--branch", d.Tag)
	}
	args = append(args, d.Git, dir)
	if d.Rev != "" {
		// 指定 rev 时先 clone 再 checkout（浅克隆可能取不到该提交，失败则全量）
		args = []string{"clone", d.Git, dir}
	}
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		bag.Error("E7022", diag.Empty, "拉取 git 依赖 `"+d.Name+"` 失败", firstLine(msg),
			"检查网络与地址；离线环境请先 `aila deps --vendor` 再把 vendor/ 一起提交")
		return "", false
	}
	if d.Rev != "" {
		co := exec.Command("git", "-C", dir, "checkout", "--detach", d.Rev)
		if out2, err2 := co.CombinedOutput(); err2 != nil {
			bag.Error("E7022", diag.Empty, "检出 rev `"+d.Rev+"` 失败",
				firstLine(strings.TrimSpace(string(out2))), "")
			return "", false
		}
	}
	if !fileExists(filepath.Join(dir, manifest.FileName)) {
		bag.Error("E7022", diag.Empty, "git 依赖 `"+d.Name+"` 仓库根没有 `aila.toml`",
			d.Git, "依赖仓库的根目录必须是 AILA 包根")
		return "", false
	}
	return dir, true
}

// gitCacheRoot 是 git 依赖的缓存根目录。
//
// 优先级：`AILA_DEPS_CACHE`（测试/CI 隔离）→ `AILA_HOME/cache/deps` → 用户缓存目录。
func gitCacheRoot() string {
	if p := strings.TrimSpace(os.Getenv("AILA_DEPS_CACHE")); p != "" {
		return p
	}
	if home := strings.TrimSpace(os.Getenv("AILA_HOME")); home != "" {
		return filepath.Join(home, "cache", "deps")
	}
	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, "aila", "deps")
	}
	return filepath.Join(os.TempDir(), "aila-deps")
}

// --------------------------------------------------------------------------- //
// 内容哈希
// --------------------------------------------------------------------------- //

// dirHash 计算包目录的内容哈希（aila.toml + 全部 .aila 源文件，按路径排序）。
func dirHash(dir string) string {
	var files []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		name := info.Name()
		if name == manifest.FileName || strings.HasSuffix(name, ".aila") {
			rel, relErr := filepath.Rel(dir, path)
			if relErr == nil {
				files = append(files, rel)
			}
		}
		return nil
	})
	sort.Strings(files)
	h := sha256.New()
	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			continue
		}
		fmt.Fprintf(h, "%s\x00%d\x00", filepath.ToSlash(rel), len(data))
		h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:12]
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}

var errNotImplemented = errors.New("not implemented")
