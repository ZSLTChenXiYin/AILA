// Package resolve 解析「用哪个工具链」（docs/rfcs/0006 §5）。
//
// 工具按「家族 + 版本」装在 toolchains/ 下，能力运行时探测：
// clang 系用 --print-targets 自报架构覆盖，gcc 目录名即 triple。
// 角色推导内置而非配置：
//
//	ir2obj：自报支持目标架构的 clang 系家族（交叉时特化家族优先，宿主时通版优先）
//	cc    ：目标 triple 的 GCC 目录优先（交叉 C 必须走默认目标 = triple 的 GCC），
//	        否则默认目标恰为 triple 的 clang 系（宿主语义）
//	clang ：bindgen 首选档，通版 clang 优先
//
// 傻瓜路（默认）= 只用 toolchains/；清单显式写 `system:...` 才去 PATH 找；
// 写 `bundled` 则只认 toolchains/（缺就报错，绝不悄悄换）。
package resolve

import (
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"aila/internal/home"
	"aila/internal/registry"
	"aila/internal/targets"
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
	Family  string // 家族（clang / esp-clang / gcc）；system 来源为空
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
	fam := ""
	if t.Family != "" {
		fam = t.Family + " "
	}
	ver := ""
	if t.Version != "" {
		ver = " " + t.Version
	}
	return tag + " " + fam + t.Name + ver + " @ " + t.Path
}

// VersionBand 返回全局绑定的 LLVM 版本带。
func VersionBand() (string, string) { return targets.VersionBand() }

// InBand 判断版本是否落在全局带内。
func InBand(version string) bool {
	lo, hi := VersionBand()
	return inRange(version, lo, hi)
}

func inRange(version, lo, hi string) bool {
	if version == "" {
		return false
	}
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

// inBandFor 按 family 校验版本带：registry 有 per-family 带则用之，
// 否则全局带；gcc 无 LLVM 版本带（永远带内）。
func inBandFor(family, version string) bool {
	if family == "gcc" {
		return true
	}
	if f := registry.Load(); f != nil {
		if b, ok := f.VersionBands[family]; ok && len(b) >= 2 {
			return inRange(version, b[0], b[1])
		}
	}
	return InBand(version)
}

func bandNote(family, version string) string {
	lo, hi := VersionBand()
	label := "全局绑定版本带"
	if f := registry.Load(); f != nil {
		if b, ok := f.VersionBands[family]; ok && len(b) >= 2 {
			lo, hi = b[0], b[1]
			label = "家族绑定版本带"
		}
	}
	shown := version
	if shown == "" {
		shown = "未知"
	}
	return "版本 " + shown + " 不在" + label + " " + lo + ".." + hi + " 内"
}

var versionRe = regexp.MustCompile(`(?:LLVM|GCC|clang)\s+version\s+([0-9][0-9.]*)`)
var numRe = regexp.MustCompile(`([0-9]+(\.[0-9]+)+)`)

func runVersion(path string) string {
	out, err := exec.Command(path, "--version").CombinedOutput()
	if err != nil && len(out) == 0 {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	// llc --version 的版本号不在首行（首行是 "LLVM (http://llvm.org/)"），逐行找
	for _, line := range lines {
		if m := versionRe.FindStringSubmatch(line); m != nil {
			return m[1]
		}
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if m := numRe.FindStringSubmatch(line); m != nil {
			return m[1]
		}
		break
	}
	return ""
}

// --------------------------------------------------------------------------- //
// 探测：clang 系自报架构覆盖（--print-targets / llc --version）、默认三元组
// --------------------------------------------------------------------------- //

var probeCache sync.Map // "path|mode" → []string（架构列表）或 triple 字符串

// RegisteredTargets 探测一个工具自报支持的架构列表。
// clang 系用 `--print-targets`，llc 用 `--version`（都输出 Registered Targets 段）。
func RegisteredTargets(toolPath string, isLLC bool) []string {
	args := []string{"--print-targets"}
	if isLLC {
		args = []string{"--version"}
	}
	key := toolPath + "|" + strings.Join(args, " ")
	if v, ok := probeCache.Load(key); ok {
		if list, ok := v.([]string); ok {
			return list
		}
	}
	out, err := exec.Command(toolPath, args...).CombinedOutput()
	if err != nil && len(out) == 0 {
		probeCache.Store(key, []string{})
		return nil
	}
	list := parseRegisteredTargets(string(out))
	probeCache.Store(key, list)
	return list
}

// parseRegisteredTargets 从 "Registered Targets" 段提取首列架构名；
// 没有该头部时按整段目标行解析（行格式 `name  -  description`）。
func parseRegisteredTargets(out string) []string {
	lines := strings.Split(out, "\n")
	start := -1
	for i, line := range lines {
		if strings.Contains(line, "Registered Targets") {
			start = i + 1
			break
		}
	}
	if start < 0 {
		start = 0
	}
	var list []string
	for _, line := range lines[start:] {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == "-" {
			list = append(list, fields[0])
			continue
		}
		if len(list) > 0 {
			break // 段结束
		}
	}
	return list
}

// supportsArch 探测一个工具是否自报支持某架构（归一化比较：x86_64 ↔ x86-64）。
func supportsArch(toolPath string, isLLC bool, arch string) bool {
	want := normArch(arch)
	for _, got := range RegisteredTargets(toolPath, isLLC) {
		if normArch(got) == want {
			return true
		}
	}
	return false
}

func normArch(s string) string {
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return strings.ToLower(s)
}

// defaultTriple 探测一个 clang 的默认目标三元组（-print-target-triple）。
func defaultTriple(clangPath string) string {
	if v, ok := probeCache.Load("dt|" + clangPath); ok {
		if s, _ := v.(string); s != "" {
			return s
		}
		return ""
	}
	out, err := exec.Command(clangPath, "-print-target-triple").Output()
	t := ""
	if err == nil {
		t = strings.TrimSpace(string(out))
	}
	probeCache.Store("dt|"+clangPath, t)
	return t
}

// --------------------------------------------------------------------------- //
// 家族发现与排序
// --------------------------------------------------------------------------- //

// clangInstalls 返回排序后的 clang 系安装：
//   - 目标为宿主三元组时通版 clang（family=clang）在前——它与 cc/链接器同家族同版本；
//   - 交叉目标时特化家族（esp-clang 等）在前——厂商验证过的后端优先；
//   - 同族内版本高的在前。
func clangInstalls(targetTriple string) []home.ClangInstall {
	installs := home.InstalledClangs()
	genericFirst := targetTriple == "" || targetTriple == HostTriple()
	sort.SliceStable(installs, func(i, j int) bool {
		a, b := installs[i], installs[j]
		ag, bg := a.Family == "clang", b.Family == "clang"
		if ag != bg {
			if genericFirst {
				return ag
			}
			return bg
		}
		if a.Family != b.Family {
			return a.Family < b.Family
		}
		return versionNewer(b.Version, a.Version)
	})
	return installs
}

// versionNewer 粗粒度版本比较（目录名即版本，形如 23.1.3）。
func versionNewer(a, b string) bool {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		na, ea := strconv.Atoi(strings.TrimSpace(as[i]))
		nb, eb := strconv.Atoi(strings.TrimSpace(bs[i]))
		if ea == nil && eb == nil && na != nb {
			return na > nb
		}
	}
	return false
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

// HostTriple 问通版 clang 的默认目标三元组（特化家族的默认 target 不代表宿主）；
// 没有通版 clang 时按平台推断。
func HostTriple() string {
	for _, inst := range home.InstalledClangs() {
		if inst.Family != "clang" {
			continue
		}
		if p, ok := home.Bins(inst.Dir)["clang"]; ok {
			if t := defaultTriple(p); t != "" {
				return t
			}
			break
		}
	}
	return PlatformTriple()
}

// --------------------------------------------------------------------------- //
// 角色解析
// --------------------------------------------------------------------------- //

// clangForIR2Obj 在自报支持目标架构的 clang 系家族里选 IR 后端。
// 带外候选不采用（继续找）；全部落空时返回最后一个带外原因供诊断。
func clangForIR2Obj(triple string) (Tool, bool) {
	if !targets.Valid(triple) {
		return Tool{}, false
	}
	arch := targets.Parse(triple).Arch
	var last Tool
	for _, inst := range clangInstalls(triple) {
		bins := home.Bins(inst.Dir)
		for _, name := range []string{"llc", "clang"} {
			p, ok := bins[name]
			if !ok {
				continue
			}
			if !supportsArch(p, name == "llc", arch) {
				continue
			}
			ver := runVersion(p)
			if ver == "" {
				ver = inst.Version
			}
			tool := Tool{
				Role: RoleIR2Obj, Name: name, Path: p, Version: ver,
				Source: SourceBundled, Family: inst.Family,
				InBand: inBandFor(inst.Family, ver),
			}
			if !tool.InBand {
				tool.Note = bandNote(inst.Family, ver)
				last = tool
				continue
			}
			return tool, true
		}
	}
	return last, last.Note != ""
}

// gccForCC 目标 triple 的 GCC 目录（交叉 C 编译 + GNU ld + libgcc）。
func gccForCC(triple string) (Tool, bool) {
	for _, g := range home.InstalledGCCs() {
		if g.Triple != triple {
			continue
		}
		bins := home.Bins(g.Dir)
		for _, name := range []string{triple + "-gcc", "gcc"} {
			p, ok := bins[name]
			if !ok {
				continue
			}
			ver := runVersion(p)
			if ver == "" {
				ver = g.Version
			}
			// GCC 不吃 LLVM IR，无 LLVM 版本带约束
			return Tool{Role: RoleCC, Name: name, Path: p, Version: ver,
				Source: SourceBundled, Family: "gcc", InBand: true}, true
		}
	}
	return Tool{}, false
}

// clangForCC 默认目标恰为 triple 的 clang 系（宿主语义）。
// CompileC 不传 -target：交叉 C 必须走默认目标 = triple 的 GCC（gccForCC），
// 让 clang 兼任交叉 cc 会静默编错架构。
func clangForCC(triple string) (Tool, bool) {
	var last Tool
	for _, inst := range clangInstalls(triple) {
		p, ok := home.Bins(inst.Dir)["clang"]
		if !ok {
			continue
		}
		if defaultTriple(p) != triple {
			continue
		}
		ver := runVersion(p)
		if ver == "" {
			ver = inst.Version
		}
		tool := Tool{Role: RoleCC, Name: "clang", Path: p, Version: ver,
			Source: SourceBundled, Family: inst.Family,
			InBand: inBandFor(inst.Family, ver)}
		if !tool.InBand {
			tool.Note = bandNote(inst.Family, ver)
			last = tool
			continue
		}
		return tool, true
	}
	return last, last.Note != ""
}

// ResolveCrossCC 交叉目标的 C 编译器 / 链接器：auto / bundled 只认随包 GCC 目录
// （目录名 = 官方工具链 triple，如 xtensa-lx106-elf），**绝不回退宿主 PATH**——
// 宿主 gcc 编不出目标架构代码，静默编错比明确报错更糟。
// system:<名> 仍尊重用户显式指定（高级用户自担后果）。
func ResolveCrossCC(gccTriple, preference string) Tool {
	pref := strings.TrimSpace(preference)
	if pref == "" {
		pref = "auto"
	}
	if strings.HasPrefix(pref, "system:") {
		return ResolveCC(gccTriple, pref)
	}
	if t, ok := gccForCC(gccTriple); ok {
		return t
	}
	return Tool{Role: RoleCC, Source: SourceMissing,
		Note: "交叉链接需要随包 GCC：toolchains/gcc/" + gccTriple +
			" 未安装（执行 `aila toolchain install gcc/" + gccTriple + "`）"}
}

// ResolveCC 解析 C 编译器 / 链接器。
func ResolveCC(triple, preference string) Tool {
	if triple == "" {
		triple = HostTriple()
	}
	pref := strings.TrimSpace(preference)
	if pref == "" {
		pref = "auto"
	}
	switch {
	case pref == "bundled":
		if t, ok := gccForCC(triple); ok {
			return t
		}
		if t, ok := clangForCC(triple); ok {
			return t
		}
		return Tool{Role: RoleCC, Source: SourceMissing,
			Note: "清单写了 `bundled`，但 toolchains/ 里没有默认目标为 " + triple + " 的 GCC 或 clang"}
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
	if t, ok := gccForCC(triple); ok {
		return t
	}
	if t, ok := clangForCC(triple); ok {
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
	return Tool{Role: RoleCC, Source: SourceMissing, Note: "toolchains/ 没有可服务 " + triple + " 的工具，系统 PATH 里也没有 clang/gcc"}
}

// ResolveLLVM 解析 IR → 目标文件的 LLVM 后端，host 语义。
func ResolveLLVM(preference string) Tool {
	return ResolveLLVMFor("", preference)
}

// ResolveLLVMFor 按目标 triple 解析 LLVM 后端。
func ResolveLLVMFor(triple, preference string) Tool {
	if triple == "" {
		triple = HostTriple()
	}
	pref := strings.TrimSpace(preference)
	if pref == "" {
		pref = "auto"
	}
	switch {
	case pref == "bundled":
		if t, ok := clangForIR2Obj(triple); ok {
			return t
		}
		return Tool{Role: RoleIR2Obj, Source: SourceMissing,
			Note: "清单写了 `bundled`，但 toolchains/ 里没有自报支持 " + triple + " 的 clang 系工具"}
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
	if t, ok := clangForIR2Obj(triple); ok {
		return t
	}
	for _, name := range []string{"llc", "clang"} {
		if t, ok := fromPath(RoleIR2Obj, name, ""); ok {
			t.Note = "未找到随包工具链，已回退到系统 LLVM（开发期行为）"
			return t
		}
	}
	return Tool{Role: RoleIR2Obj, Source: SourceMissing,
		Note: "toolchains/ 里没有自报支持 " + triple + " 的 clang 系工具，系统里也没有 llc/clang"}
}

// ResolveClangForIR 选「能以 clang `-x ir -c` 编译该目标 IR」的 clang（coro 路由用）。
// 与 clangForIR2Obj 同一筛选（自报支持目标架构 + 带内优先），但只要 clang——
// llc 不跑 coro 中端 pass，含 llvm.coro 的 IR 必须交给 clang；
// 特化家族（esp-clang 等）在 clangInstalls 里已按交叉目标排前。
func ResolveClangForIR(triple string) (Tool, bool) {
	if !targets.Valid(triple) {
		return Tool{}, false
	}
	arch := targets.Parse(triple).Arch
	var last Tool
	for _, inst := range clangInstalls(triple) {
		p, ok := home.Bins(inst.Dir)["clang"]
		if !ok {
			continue
		}
		if !supportsArch(p, false, arch) {
			continue
		}
		ver := runVersion(p)
		if ver == "" {
			ver = inst.Version
		}
		tool := Tool{
			Role: RoleClang, Name: "clang", Path: p, Version: ver,
			Source: SourceBundled, Family: inst.Family,
			InBand: inBandFor(inst.Family, ver),
		}
		if !tool.InBand {
			tool.Note = bandNote(inst.Family, ver)
			last = tool
			continue
		}
		return tool, true
	}
	return last, last.Note != ""
}

// ResolveClang 解析 bindgen 首选档（通版 clang 优先——AST dump 覆盖面最大）。
func ResolveClang(triple string) Tool {
	installs := clangInstalls(triple)
	// 通版优先
	for _, inst := range installs {
		if inst.Family != "clang" {
			continue
		}
		if p, ok := home.Bins(inst.Dir)["clang"]; ok {
			ver := runVersion(p)
			if ver == "" {
				ver = inst.Version
			}
			return Tool{Role: RoleClang, Name: "clang", Path: p, Version: ver,
				Source: SourceBundled, Family: inst.Family, InBand: inBandFor(inst.Family, ver)}
		}
	}
	// 其次特化家族
	for _, inst := range installs {
		if inst.Family == "clang" {
			continue
		}
		if p, ok := home.Bins(inst.Dir)["clang"]; ok {
			ver := runVersion(p)
			if ver == "" {
				ver = inst.Version
			}
			return Tool{Role: RoleClang, Name: "clang", Path: p, Version: ver,
				Source: SourceBundled, Family: inst.Family, InBand: inBandFor(inst.Family, ver)}
		}
	}
	if t, ok := fromPath(RoleClang, "clang", ""); ok {
		t.Note = "未找到随包 clang，使用系统 clang"
		return t
	}
	return Tool{Role: RoleClang, Source: SourceMissing, Note: "无 clang → bindgen 走 builtin 兜底档"}
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

func splitSpec(spec string) (name, minVersion string) {
	if idx := strings.Index(spec, ">="); idx >= 0 {
		return strings.TrimSpace(spec[:idx]), strings.TrimSpace(spec[idx+2:])
	}
	return strings.TrimSpace(spec), ""
}

// ResolveAll 一次解析三件工具（triple 为空 = 宿主）。
func ResolveAll(triple, ccPref, llvmPref string) map[string]Tool {
	return map[string]Tool{
		RoleCC:     ResolveCC(triple, ccPref),
		RoleIR2Obj: ResolveLLVMFor(triple, llvmPref),
		RoleClang:  ResolveClang(triple),
	}
}

// Exists 判断路径是否存在（供 profile 用）。
func Exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
