// Package manifest 解析并校验包清单 aila.toml（docs/specs/0001 §2）。
package manifest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/ailang/aila/internal/diag"
	"github.com/ailang/aila/internal/targets"
)

// FileName 是清单文件名。
const FileName = "aila.toml"

// Manifest 是校验过的清单。
type Manifest struct {
	Path     string
	Name     string
	Version  string
	Entry    string
	Board    string
	Arch     string
	Triple   string
	Opt      int
	Warnings string
	Debug    bool
	CC       string
	LLVM     string
	Domains  map[string]Domain
	// Dependencies 是跨项目依赖（docs/rfcs/0004）：名字 → 声明。
	Dependencies map[string]Dependency
}

// Dependency 是一条 `[dependencies]` 声明。
type Dependency struct {
	Path    string `toml:"path"`
	Git     string `toml:"git"`
	Tag     string `toml:"tag"`
	Rev     string `toml:"rev"`
	Version string `toml:"version"`
}

// Domain 是一个物理域声明。
type Domain struct {
	Kind  string
	Size  string
	Align int
}

type rawManifest struct {
	Package struct {
		Name    string `toml:"name"`
		Version string `toml:"version"`
		Entry   string `toml:"entry"`
	} `toml:"package"`
	Target struct {
		Board  string `toml:"board"`
		Arch   string `toml:"arch"`
		Triple string `toml:"triple"`
	} `toml:"target"`
	Build struct {
		Opt       *int   `toml:"opt"`
		Warnings  string `toml:"warnings"`
		Debug     bool   `toml:"debug"`
		Toolchain struct {
			CC   string `toml:"cc"`
			LLVM string `toml:"llvm"`
		} `toml:"toolchain"`
	} `toml:"build"`
	Domains map[string]struct {
		Kind  string `toml:"kind"`
		Size  string `toml:"size"`
		Align int    `toml:"align"`
	} `toml:"domains"`
	Dependencies map[string]Dependency `toml:"dependencies"`
}

var nameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)
var sizeRe = regexp.MustCompile(`^[0-9]+[KMGkmg]?$`)

// Load 读取并校验清单；失败返回 nil 与诊断。
func Load(root string, bag *diag.Bag) *Manifest {
	path := filepath.Join(root, FileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		bag.Error("E7002", diag.Empty, "目录里没有清单 `"+FileName+"`",
			root, "新建工程用 `aila new <目录>`；已有工程请在包根放 aila.toml")
		return nil
	}
	var r rawManifest
	md, err := toml.Decode(string(raw), &r)
	if err != nil {
		var perr toml.ParseError
		if errors.As(err, &perr) {
			bag.Error("E7002", diag.Span{File: path, Line: perr.Position.Line, Col: perr.Position.Start + 1},
				"清单不是合法 TOML", firstLineOf(perr.Error()), "")
			return nil
		}
		bag.Error("E7002", diag.Empty, "清单不是合法 TOML", err.Error(), "")
		return nil
	}
	m := &Manifest{Path: path, Domains: map[string]Domain{}, Dependencies: map[string]Dependency{}}
	m.Name = strings.TrimSpace(r.Package.Name)
	m.Version = strings.TrimSpace(r.Package.Version)
	m.Entry = strings.TrimSpace(r.Package.Entry)
	if m.Entry == "" {
		m.Entry = "_.main"
	}
	if m.Name == "" {
		bag.Error("E7002", diag.Empty, "清单缺少 `[package] name`", "必填字段", `补 name = "demo"`)
	}
	if m.Version == "" {
		bag.Error("E7002", diag.Empty, "清单缺少 `[package] version`", "必填字段", `补 version = "0.1.0"`)
	} else if !nameRe.MatchString(m.Name) && m.Name != "" {
		bag.Error("E7002", diag.Empty, "包名 `"+m.Name+"` 不合法", "应为 [A-Za-z_][A-Za-z0-9_-]*", "")
	}

	m.Board = strings.TrimSpace(r.Target.Board)
	if m.Board == "" {
		m.Board = "host"
	}
	m.Arch = strings.TrimSpace(r.Target.Arch)
	m.Triple = strings.TrimSpace(r.Target.Triple)
	if _, ok := targets.Find(m.Board); !ok && m.Arch == "" && m.Triple == "" {
		names, _ := targets.Names()
		bag.Error("E7006", diag.Empty, "未知 board `"+m.Board+"`，且没有给 `arch` / `triple`",
			"未知 board", "候选："+strings.Join(names, " / "))
	}

	m.Opt = 0
	if r.Build.Opt != nil {
		if *r.Build.Opt < 0 || *r.Build.Opt > 3 {
			bag.Error("E7007", diag.Empty, fmt.Sprintf("`build.opt` 只接受 0..3，实得 %d", *r.Build.Opt), "", "")
		} else {
			m.Opt = *r.Build.Opt
		}
	}
	m.Warnings = r.Build.Warnings
	if m.Warnings == "" {
		m.Warnings = "default"
	}
	if m.Warnings != "default" && m.Warnings != "strict" && m.Warnings != "allow" {
		bag.Error("E7007", diag.Empty, "`build.warnings` 只接受 default/strict/allow，实得 `"+m.Warnings+"`", "", "")
	}
	m.Debug = r.Build.Debug
	m.CC = defaultIfEmpty(r.Build.Toolchain.CC, "auto")
	m.LLVM = defaultIfEmpty(r.Build.Toolchain.LLVM, "auto")

	kinds := map[string]bool{}
	for _, k := range targets.DomainKinds() {
		kinds[k] = true
	}
	for name, spec := range r.Domains {
		kind := defaultIfEmpty(spec.Kind, "default")
		if !kinds[kind] {
			bag.Error("E7008", diag.Empty, "域 `"+name+"` 的 kind `"+kind+"` 不在内置枚举内",
				"候选："+strings.Join(targets.DomainKinds(), ", "), "")
		}
		if spec.Size != "" && !sizeRe.MatchString(spec.Size) {
			bag.Error("E7008", diag.Empty, "域 `"+name+"` 的 size `"+spec.Size+"` 不合法", "形如 64K / 1M", "")
		}
		if spec.Align != 0 && (spec.Align < 0 || spec.Align&(spec.Align-1) != 0) {
			bag.Error("E7008", diag.Empty, "域 `"+name+"` 的 align 必须是 2 的幂", "", "")
		}
		m.Domains[name] = Domain{Kind: kind, Size: spec.Size, Align: spec.Align}
	}

	for name, d := range r.Dependencies {
		if !nameRe.MatchString(name) {
			bag.Error("E7021", diag.Empty, "依赖名 `"+name+"` 不是合法标识符", "这里",
				"依赖名会当成导入根：`import "+name+".子包.模块;`")
			continue
		}
		m.Dependencies[name] = Dependency{Path: d.Path, Git: d.Git, Tag: d.Tag, Rev: d.Rev, Version: d.Version}
	}

	for _, key := range md.Undecoded() {
		bag.Warning("W7001", diag.Empty, "清单含未知键 `"+key.String()+"`", "未知键不会生效",
			"对照 docs/specs/0001 §2 修正拼写")
	}
	return m
}

func firstLineOf(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}

func defaultIfEmpty(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return strings.TrimSpace(v)
}

// EffectiveTriple 返回该清单生效的目标三元组。
func (m *Manifest) EffectiveTriple(hostTriple string) string {
	if m.Triple != "" {
		return m.Triple
	}
	if b, ok := targets.Find(m.Board); ok && b.Triple != nil {
		return *b.Triple
	}
	return hostTriple
}

// PtrSize 返回目标指针字节数（32 位与 xtensa 为 4）。
func (m *Manifest) PtrSize() int {
	t := m.Triple
	if t == "" {
		if b, ok := targets.Find(m.Board); ok && b.Triple != nil {
			t = *b.Triple
		}
	}
	if strings.Contains(t, "32") || strings.HasPrefix(t, "xtensa") {
		return 4
	}
	return 8
}
