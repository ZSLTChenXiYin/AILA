package lsp

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"aila/internal/ast"
	"aila/internal/bindgen"
	"aila/internal/diag"
	"aila/internal/manifest"
	"aila/internal/project"
	"aila/internal/resolve"
)

// 本文件让编辑器也认得 **C 导入**：`import ch _.cpacks.<名字>`。
//
// 编译器的做法是「导入时现算 bindgen → 生成 `.ail` → 当普通模块」；
// 编辑器要显示补全 / 跳转 / 悬停，就得拿到同一份 `.ail`，所以这里复用同一条流水线：
//
//	读 cpack 目录 → bindgen.Run → `.ail` 文本 → 当模块索引
//
// 为了不在每次按键都跑一遍 bindgen，缓存分两层：
//  1. 进程内：`Workspace.cbind`，命中条件 = 规则版本 + triple + cpack 内容指纹；
//  2. 磁盘：`build/.cache/bindgen-lsp/<cpack>.ail` + `.key`，跨会话复用。
//
// 指纹只用「路径 + 大小 + 修改时间」，不做内容哈希——够快，也够准（改头文件必然变）。

// cBindingSchema 是映射规则的版本号：规则改了就换它，避免拿旧缓存装作成功
// （`internal/build` 里也有一个同义常量，两处都为「缓存不许撒谎」服务）。
const cBindingSchema = "lsp-v2"

// cBinding 是一条 C 导入解析出来的绑定模块。
type cBinding struct {
	Key   string // 规则版本 + triple + cpack 指纹
	Cpack string
	Name  string // 模块名：`_.cpacks.<cpack>`
	// Text 为空表示生成失败，原因在 Problems 里（负缓存，避免每次按键重跑 bindgen）。
	Text     string
	Path     string // 绑定的落盘位置（同时用作跳转 URI）
	Problems []diag.Diagnostic
}

// importTargetModule 返回一条 import 指向的 AILA 模块名（C 导入返回绑定模块名）。
//
// 与 `internal/build` 的命名保持一致：
//   - `_` 根 → `_.<path>`（示例：`_.sub.util`）
//   - `std` 根 → `std.<path>`
//   - `import ch _.cpacks.X` → `_.cpacks.X`（bindgen 生成的绑定模块）
//   - `import cc _.cpacks.X` → 空（只登记 C 源一起编译，不产生绑定）
func importTargetModule(imp *ast.Import) string {
	if imp.KindV == "ch" || imp.KindV == "cc" {
		if imp.Root != "_" || len(imp.Path) < 2 || imp.Path[0] != "cpacks" {
			return ""
		}
		if imp.KindV == "cc" {
			return ""
		}
		return "_.cpacks." + imp.Path[1]
	}
	switch imp.Root {
	case "_":
		if len(imp.Path) == 0 {
			return "_"
		}
		return "_." + strings.Join(imp.Path, ".")
	case "std":
		if len(imp.Path) == 0 {
			return "std"
		}
		return "std." + strings.Join(imp.Path, ".")
	}
	return "" // 依赖包根（如 `mylib.foo`）目前不索引，留到依赖批次
}

// importBindingName 返回 import 在当前文件里绑定的名字（`as` 优先，否则最后一段）。
func importBindingName(imp *ast.Import) string {
	if imp.Alias != "" {
		return imp.Alias
	}
	if len(imp.Path) > 0 {
		return imp.Path[len(imp.Path)-1]
	}
	return ""
}

// cBindingFor 解析（必要时生成）一条 C 导入的绑定模块。
// 参数 man 为 nil 时返回 nil（清单都没读出来，谈不上生成）。
func (w *Workspace) cBindingFor(root string, prj *project.Project, man *manifest.Manifest, imp *ast.Import) *cBinding {
	if man == nil || prj == nil {
		return nil
	}
	if imp.Root != "_" {
		return &cBinding{Problems: []diag.Diagnostic{{
			Code: "E7010", Span: imp.Span, Severity: diag.SeverityError,
			Message: "C 导入必须以 `_` 为根（`_.cpacks.<名字>`）",
		}}}
	}
	if len(imp.Path) < 2 || imp.Path[0] != "cpacks" {
		return &cBinding{Problems: []diag.Diagnostic{{
			Code: "E7010", Span: imp.Span, Severity: diag.SeverityError,
			Message: "C 导入路径必须形如 `_.cpacks.<名字>`（把 C 包放到 `cpacks/<名字>/` 下）",
		}}}
	}
	name := imp.Path[1]
	dir, ok := prj.Cpacks[name]
	if !ok {
		return &cBinding{Problems: []diag.Diagnostic{{
			Code: "E3001", Span: imp.Span, Severity: diag.SeverityError,
			Message: "找不到 cpack 目录 `cpacks/" + name + "`（目录名即 cpack 名）",
		}}}
	}

	triple := man.EffectiveTriple(resolve.HostTriple())
	key := cBindingSchema + "|" + triple + "|" + cpackStamp(dir)
	cacheKey := root + "|" + name
	if w.cbind == nil {
		w.cbind = map[string]*cBinding{}
	}
	if b, ok := w.cbind[cacheKey]; ok && b.Key == key {
		return b
	}

	cacheDir := filepath.Join(root, "build", ".cache", "bindgen-lsp")
	ailPath := filepath.Join(cacheDir, name+".ail")
	keyPath := filepath.Join(cacheDir, name+".key")
	if text, ok := readCachedBinding(keyPath, ailPath, key); ok {
		b := &cBinding{Key: key, Cpack: name, Name: "_.cpacks." + name, Text: text, Path: ailPath}
		w.cbind[cacheKey] = b
		return b
	}

	bag := &diag.Bag{}
	target := bindgen.Target{Board: man.Board, Triple: triple, PtrSize: man.PtrSize()}
	br := bindgen.Run(name, dir, target, bag, "auto")
	if br == nil || bag.HasErrors() {
		b := &cBinding{Key: key, Cpack: name, Name: "_.cpacks." + name, Problems: bag.Items}
		w.cbind[cacheKey] = b // 负缓存：别每次按键都重跑 bindgen
		return b
	}
	text := br.Header(name, target) + "\n\n" + br.Text
	if err := os.MkdirAll(cacheDir, 0o777); err == nil {
		_ = os.WriteFile(ailPath, []byte(text), 0o666)
		_ = os.WriteFile(keyPath, []byte(key), 0o666)
	}
	b := &cBinding{Key: key, Cpack: name, Name: "_.cpacks." + name, Text: text, Path: ailPath}
	w.cbind[cacheKey] = b
	return b
}

// readCachedBinding 读磁盘缓存（key 必须完全一致）。
func readCachedBinding(keyPath, ailPath, key string) (string, bool) {
	kb, err := os.ReadFile(keyPath)
	if err != nil || strings.TrimSpace(string(kb)) != key {
		return "", false
	}
	text, err := os.ReadFile(ailPath)
	if err != nil {
		return "", false
	}
	return string(text), true
}

// cpackStamp 用「相对路径 + 大小 + 修改时间」给 cpack 目录拍指纹。
func cpackStamp(dir string) string {
	var paths []string
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == "build" {
				return fs.SkipDir
			}
			return nil
		}
		paths = append(paths, p)
		return nil
	})
	sort.Strings(paths)
	h := sha256.New()
	for _, p := range paths {
		st, err := os.Stat(p)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			rel = p
		}
		fmt.Fprintf(h, "%s|%d|%d\n", filepath.ToSlash(rel), st.Size(), st.ModTime().UnixNano())
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}
