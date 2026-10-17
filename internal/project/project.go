// Package project 是工程模型：项目 = 包、子包 = 目录、模块 = 文件（docs/specs/0001 §1）。
package project

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ailang/aila/internal/diag"
	"github.com/ailang/aila/internal/home"
)

var identRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// reservedDirs 是不参与模块发现的目录（构建产物与工具目录）。
var reservedDirs = map[string]bool{
	"build": true, "cpacks": true, ".git": true, ".aila": true,
	"target": true, "node_modules": true,
}

// Project 是一个已发现的工程（一个「包」）。
type Project struct {
	Root       string
	// Prefix 是模块名根：根包是 `_`，依赖包是它的依赖名（`mylib`），标准库是 `std`。
	Prefix     string
	Modules    map[string]string // 模块名 -> 源文件绝对路径
	ModuleList []string          // 发现顺序（确定性输出用）
	Cpacks     map[string]string // cpack 名 -> 目录
	StdlibRoot string
}

// Discover 扫描根包（模块名根 = `_`）。
func Discover(root string, bag *diag.Bag) *Project {
	return DiscoverPackage(root, "_", bag)
}

// DiscoverPackage 扫描一个包（模块名根用 prefix），收集模块与 cpack。
func DiscoverPackage(root, prefix string, bag *diag.Bag) *Project {
	abs, err := filepath.Abs(root)
	if err == nil {
		root = abs
	}
	p := &Project{
		Root:       root,
		Prefix:     prefix,
		Modules:    map[string]string{},
		Cpacks:     map[string]string{},
		StdlibRoot: home.StdlibDir(),
	}
	cpacksDir := filepath.Join(root, "cpacks")
	if entries, err := os.ReadDir(cpacksDir); err == nil {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				names = append(names, e.Name())
			}
		}
		sort.Strings(names)
		for _, name := range names {
			p.Cpacks[name] = filepath.Join(cpacksDir, name)
		}
	}
	// 递归扫描（跳过保留目录）
	var walk func(dir string, parts []string)
	walk = func(dir string, parts []string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		dirNames := make([]string, 0, len(entries))
		fileNames := make([]string, 0, len(entries))
		for _, e := range entries {
			name := e.Name()
			if strings.HasPrefix(name, ".") {
				continue
			}
			if e.IsDir() {
				if reservedDirs[name] {
					continue
				}
				dirNames = append(dirNames, name)
				continue
			}
			fileNames = append(fileNames, name)
		}
		sort.Strings(dirNames)
		sort.Strings(fileNames)
		for _, fname := range fileNames {
			if !strings.HasSuffix(fname, ".aila") {
				continue
			}
			full := filepath.Join(dir, fname)
			stem := strings.TrimSuffix(fname, ".aila")
			span := spanAt(full)
			if !identRe.MatchString(stem) {
				bag.Error("E7003", span, "模块文件名 `"+stem+"` 不是合法标识符",
					"路径即身份，模块名 = 文件名",
					"改成 `[A-Za-z_][A-Za-z0-9_]*`（如 `my-lib` → `mylib`）")
				continue
			}
			bad := ""
			for _, part := range parts {
				if !identRe.MatchString(part) {
					bad = part
					break
				}
			}
			if bad != "" {
				bag.Error("E7003", span, "子包目录名 `"+bad+"` 不是合法标识符",
					"子包名 = 目录名，路径即身份", "")
				continue
			}
			if len(parts) == 0 && stem != "main" {
				bag.Error("E7003", span, "包根只允许入口模块 `main.aila`，发现 `"+fname+"`",
					"其余模块一律放子包目录",
					"把它移到 `"+stem+"/` 目录下，导入名变成 `_."+stem+"."+stem+"` 或改名")
				continue
			}
			// 模块全名 = <包名根>.<子包...>.<模块>；包根入口模块是 `<根>.main`
			name := prefix + ".main"
			if len(parts) > 0 {
				name = prefix + "." + strings.Join(append(append([]string{}, parts...), stem), ".")
			}
			if _, dup := p.Modules[name]; dup {
				bag.Error("E7003", span, "模块名 `"+name+"` 重复", "这里", "")
				continue
			}
			p.Modules[name] = full
			p.ModuleList = append(p.ModuleList, name)
		}
		for _, d := range dirNames {
			walk(filepath.Join(dir, d), append(append([]string{}, parts...), d))
		}
	}
	walk(root, nil)
	return p
}

// ModulePath 返回模块名对应的源文件路径（含 `std.*` 标准库回退）。
func (p *Project) ModulePath(name string) string {
	if path, ok := p.Modules[name]; ok {
		return path
	}
	if name == "std" {
		candidate := filepath.Join(p.StdlibRoot, "std.aila")
		if fileExists(candidate) {
			return candidate
		}
		return ""
	}
	if strings.HasPrefix(name, "std.") {
		rel := strings.TrimPrefix(name, "std.")
		candidate := filepath.Join(p.StdlibRoot, filepath.Join(strings.Split(rel, ".")...)+".aila")
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

// ReadSource 读入模块源码。
func ReadSource(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func spanAt(path string) diag.Span {
	return diag.Span{File: path, Line: 1, Col: 1, Length: 1}
}
