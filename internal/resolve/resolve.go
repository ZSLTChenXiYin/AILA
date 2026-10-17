// Package resolve 解析「用哪个工具链」（docs/specs/0003 §2）。
//
// 傻瓜路（默认）= 只用随包相对目录；清单显式写 `system:...` 才去 PATH 找；
// 写 `bundled` 则只认随包（缺就报错，绝不悄悄换）。
package resolve

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/ailang/aila/internal/home"
	"github.com/ailang/aila/internal/targets"
)

// 工具来源。
const (
	SourceBundled = "bundled"
	SourceSystem  = "system"
	SourceMissing = "missing"
)

// 角色。
const (
	RoleIR2Obj = "ir2obj"
	RoleCC     = "cc"
	RoleClang  = "clang"
)

// Tool 是一次解析结果。
type Tool struct {
	Role    string
	Name    string
	Path    string
	Version string
	Source  string
	InBand  bool
	Note    string
}

// OK 表示这个工具可用；随包但版本带外 = 不可用（designs/0001 §12.5 版本绑定）。
func (t Tool) OK() bool {
	if t.Source == SourceMissing || t.Path == "" {
		return false
	}
	if t.Source == SourceBundled && !t.InBand {
		return false
	}
	return true
}

// Describe 是人类可读的来源描述。
func (t Tool) Describe() string {
	if !t.OK() {
		note := t.Note
		if note == "" {
			note = "无"
		}
		return "未找到（" + note + "）"
	}
	tag := map[string]string{SourceBundled: "随包", SourceSystem: "系统"}[t.Source]
	if tag == "" {
		tag = t.Source
	}
	ver := ""
	if t.Version != "" {
		ver = " " + t.Version
	}
	return tag + " " + t.Name + ver + " @ " + t.Path
}

// VersionBand 返回绑定的 LLVM 版本带。
func VersionBand() (string, string) { return targets.VersionBand() }

// InBand 判断版本是否落在带内。
func InBand(version string) bool {
	if version == "" {
		return false
	}
	lo, hi := VersionBand()
	major, err := strconv.Atoi(strings.SplitN(version, ".", 2)[0])
	if err != nil {
		return false
	}
	l, err1 := strconv.Atoi(lo)
	h, err2 := strconv.Atoi(hi)
	if err1 != nil || err2 != nil {
		return true
	}
	return major >= l && major <= h
}

var versionRe = regexp.MustCompile(`(?:LLVM|GCC|clang)\s+version\s+([0-9][0-9.]*)`)
var numRe = regexp.MustCompile(`([0-9]+(\.[0-9]+)+)`)

func runVersion(path string) string {
	out, err := exec.Command(path, "--version").CombinedOutput()
	if err != nil && len(out) == 0 {
		return ""
	}
	first := ""
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) != "" {
			first = strings.TrimSpace(line)
			break
		}
	}
	if m := versionRe.FindStringSubmatch(first); m != nil {
		return m[1]
	}
	if m := numRe.FindStringSubmatch(first); m != nil {
		return m[1]
	}
	if len(first) > 40 {
		return first[:40]
	}
	return first
}

// HostTripleGuess 返回已安装 bundle 的三元组目录名（没有 bundle 时回落到平台默认）。
func HostTripleGuess() string {
	for _, t := range home.InstalledTriples() {
		return t
	}
	return PlatformTriple()
}

// PlatformTriple 是按平台推断的默认三元组。
func PlatformTriple() string {
	switch runtime.GOOS {
	case "windows":
		return "x86_64-w64-windows-gnu"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "arm64-apple-darwin"
		}
		return "x86_64-apple-darwin"
	default:
		if runtime.GOARCH == "arm64" {
			return "aarch64-unknown-linux-gnu"
		}
		return "x86_64-unknown-linux-gnu"
	}
}

// HostTriple 优先问随包 clang，其次按平台推断。
func HostTriple() string {
	if clang := bundleTool("clang"); clang != "" {
		out, err := exec.Command(clang, "-print-target-triple").Output()
		if err == nil {
			if t := strings.TrimSpace(string(out)); t != "" {
				return t
			}
		}
	}
	return PlatformTriple()
}

func bundleTool(name string) string {
	for _, triple := range home.InstalledTriples() {
		bins := home.BundleBins(triple)
		if p, ok := bins[name]; ok {
			return p
		}
	}
	return ""
}

func fromBundle(role string, names []string) (Tool, bool) {
	for _, name := range names {
		p := bundleTool(name)
		if p == "" {
			continue
		}
		triple := filepath.Base(filepath.Dir(filepath.Dir(p)))
		bandVersion := home.BundleLLVMVersion(triple)
		version := runVersion(p)
		check := bandVersion
		if check == "" {
			check = version
		}
		tool := Tool{
			Role: role, Name: name, Path: p, Version: version,
			Source: SourceBundled, InBand: InBand(check),
		}
		if version == "" {
			tool.Version = bandVersion
		}
		if !tool.InBand {
			lo, hi := VersionBand()
			shown := check
			if shown == "" {
				shown = "未知"
			}
			tool.Note = "bundle 版本 " + shown + " 不在绑定版本带 " + lo + ".." + hi + " 内"
		}
		return tool, true
	}
	return Tool{}, false
}

func fromPath(role, name, minVersion string) (Tool, bool) {
	p, err := exec.LookPath(name)
	if err != nil {
		return Tool{}, false
	}
	version := runVersion(p)
	tool := Tool{Role: role, Name: name, Path: p, Version: version, Source: SourceSystem, InBand: true}
	if minVersion != "" && !ge(version, minVersion) {
		tool.Note = "版本 " + orUnknown(version) + " 低于声明要求 " + minVersion
		tool.Path = ""
		tool.Source = SourceMissing
		tool.InBand = false
		return tool, true
	}
	return tool, true
}

func orUnknown(v string) string {
	if v == "" {
		return "未知"
	}
	return v
}

func ge(version, want string) bool {
	num := func(s string) []int {
		var out []int
		for _, m := range numRe.FindAllStringSubmatch(s, -1) {
			for _, part := range strings.Split(m[1], ".") {
				if n, err := strconv.Atoi(part); err == nil {
					out = append(out, n)
				}
			}
			break
		}
		return out
	}
	a, b := num(version), num(want)
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return len(a) >= len(b)
}

// ResolveCC 解析 C 编译器 / 链接器。
func ResolveCC(preference string) Tool {
	pref := strings.TrimSpace(preference)
	if pref == "" {
		pref = "auto"
	}
	switch {
	case pref == "bundled":
		if t, ok := fromBundle(RoleCC, []string{"clang", "clang-cl", "cc"}); ok {
			return t
		}
		return Tool{Role: RoleCC, Source: SourceMissing, Note: "清单写了 `bundled`，但随包目录里没有 clang"}
	case strings.HasPrefix(pref, "system:"):
		spec := strings.TrimPrefix(pref, "system:")
		name, minv := splitSpec(spec)
		if name == "" {
			name = "clang"
		}
		if t, ok := fromPath(RoleCC, name, minv); ok {
			return t
		}
		return Tool{Role: RoleCC, Source: SourceMissing, Note: "清单要求 `" + pref + "`，系统 PATH 里找不到合格版本"}
	}
	if t, ok := fromBundle(RoleCC, []string{"clang", "clang-cl", "cc"}); ok {
		return t
	}
	for _, name := range []string{"clang", "gcc", "cc"} {
		if t, ok := fromPath(RoleCC, name, ""); ok {
			if t.Note == "" {
				t.Note = "未找到随包工具链，已回退到系统编译器（开发期行为）"
			}
			return t
		}
	}
	return Tool{Role: RoleCC, Source: SourceMissing, Note: "没有随包工具链，系统 PATH 里也没有 clang/gcc"}
}

// ResolveLLVM 解析 IR → 目标文件的 LLVM 后端（llc / clang）。
func ResolveLLVM(preference string) Tool {
	pref := strings.TrimSpace(preference)
	if pref == "" {
		pref = "auto"
	}
	switch {
	case pref == "bundled":
		if t, ok := fromBundle(RoleIR2Obj, []string{"llc", "clang"}); ok {
			return t
		}
		return Tool{Role: RoleIR2Obj, Source: SourceMissing, Note: "清单写了 `bundled`，但随包目录里没有 llc/clang"}
	case strings.HasPrefix(pref, "system:"):
		spec := strings.TrimPrefix(pref, "system:")
		name, minv := splitSpec(spec)
		cands := []string{}
		if name != "" {
			cands = append(cands, name)
		}
		cands = append(cands, "llc", "clang")
		for _, c := range cands {
			if t, ok := fromPath(RoleIR2Obj, c, minv); ok && t.OK() {
				return t
			}
		}
		return Tool{Role: RoleIR2Obj, Source: SourceMissing, Note: "清单要求 `" + pref + "`，系统 PATH 里找不到"}
	case pref == "llvmlite":
		// Go 版链接不到 llvmlite（不引 cgo）：明确拒绝并指路
		return Tool{Role: RoleIR2Obj, Source: SourceMissing,
			Note: "Go 版不支持 `llvmlite` 后端（不引 cgo）；请用随包 llc/clang"}
	}
	if t, ok := fromBundle(RoleIR2Obj, []string{"llc", "clang"}); ok {
		return t
	}
	for _, name := range []string{"llc", "clang"} {
		if t, ok := fromPath(RoleIR2Obj, name, ""); ok {
			t.Note = "未找到随包工具链，已回退到系统 LLVM（开发期行为）"
			return t
		}
	}
	return Tool{Role: RoleIR2Obj, Source: SourceMissing, Note: "没有随包 LLVM，系统里也没有 llc/clang"}
}

// ResolveClang 解析 bindgen 首选档。
func ResolveClang() Tool {
	if t, ok := fromBundle(RoleClang, []string{"clang"}); ok {
		return t
	}
	if t, ok := fromPath(RoleClang, "clang", ""); ok {
		t.Note = "未找到随包 clang，使用系统 clang"
		return t
	}
	return Tool{Role: RoleClang, Source: SourceMissing, Note: "无 clang → bindgen 走 builtin 兜底档"}
}

func splitSpec(spec string) (name, minVersion string) {
	if idx := strings.Index(spec, ">="); idx >= 0 {
		return strings.TrimSpace(spec[:idx]), strings.TrimSpace(spec[idx+2:])
	}
	return strings.TrimSpace(spec), ""
}

// ResolveAll 一次解析三件工具。
func ResolveAll(ccPref, llvmPref string) map[string]Tool {
	return map[string]Tool{
		RoleCC:     ResolveCC(ccPref),
		RoleIR2Obj: ResolveLLVM(llvmPref),
		RoleClang:  ResolveClang(),
	}
}

// Exists 判断路径是否存在（供 profile 用）。
func Exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
