// Package build 是构建流水线（docs/specs/0001 §5）。
//
// discover → manifest → parse → sema（多模块）→ lower(IR) → object → csrc → link → report
package build

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ailang/aila/internal/ast"
	"github.com/ailang/aila/internal/bindgen"
	"github.com/ailang/aila/internal/codegen"
	"github.com/ailang/aila/internal/deps"
	"github.com/ailang/aila/internal/diag"
	"github.com/ailang/aila/internal/home"
	"github.com/ailang/aila/internal/manifest"
	"github.com/ailang/aila/internal/parser"
	"github.com/ailang/aila/internal/project"
	"github.com/ailang/aila/internal/resolve"
	"github.com/ailang/aila/internal/sema"
	"github.com/ailang/aila/internal/toolchain"
	"github.com/ailang/aila/internal/types"
)

// Options 是一次构建的可选参数（与 `aila build` 的命令行开关一一对应）。
type Options struct {
	Opt      int
	// UpdateDeps 强制重新解析依赖并重写 `aila.lock`。
	UpdateDeps bool
	// Vendor 把解析好的依赖复制到 `vendor/`（离线构建）。
	Vendor bool
	Board    string
	LLVM     string
	CC       string
	EmitIR   bool
	EmitJSON bool
	OutName  string
}

// Result 是一次构建的结果。
type Result struct {
	OK        bool
	Diags     *diag.Bag
	Manifest  *manifest.Manifest
	Project   *project.Project
	Program   *sema.Program
	Triple    string
	Exe       string
	// Deps 是跨项目依赖图；Packages 是本次参与构建的包名（不含根包的 `_` 在前）。
	Deps     *deps.Graph
	Packages []string
	Objects   []string
	IR        map[string]string
	Cpacks    []string
	OutDir    string
	Artifacts map[string]any
}

// Load 是「解析 + 语义」阶段：模块发现、清单、逐模块解析、多模块类型检查。
func Load(root string, opts Options) *Result {
	bag := &diag.Bag{}
	res := &Result{Diags: bag, IR: map[string]string{}, Artifacts: map[string]any{}}
	man := manifest.Load(root, bag)
	if man == nil {
		return res
	}
	res.Manifest = man
	if opts.Board != "" {
		man.Board = opts.Board
	}
	if opts.Opt > 0 {
		man.Opt = opts.Opt
	}
	if opts.LLVM != "" {
		man.LLVM = opts.LLVM
	}
	if opts.CC != "" {
		man.CC = opts.CC
	}
	res.Triple = man.EffectiveTriple(resolve.HostTriple())

	// ---- 依赖图（跨项目依赖：path / git，docs/rfcs/0004）----
	graph := deps.Resolve(root, opts.UpdateDeps, bag)
	res.Deps = graph
	if opts.Vendor {
		if err := graph.Vendor(); err != nil {
			bag.Error("E7022", diag.Empty, "vendor 复制失败", err.Error(), "")
		}
	}

	// ---- 逐包发现模块：根包 `_` + 每个依赖 + 标准库（按需）----
	rootPrj := project.DiscoverPackage(root, "_", bag)
	res.Project = rootPrj
	pkgs := map[string]*project.Project{"_": rootPrj}
	for _, pkg := range graph.Order {
		if pkg == deps.RootPkg {
			continue
		}
		dir := graph.PackageDir(pkg)
		if dir == "" {
			continue
		}
		pkgs[pkg] = project.DiscoverPackage(dir, pkg, bag)
	}
	res.Packages = append([]string{"_"}, nonRootPackages(graph)...)
	if bag.HasErrors() {
		return res
	}
	if _, ok := rootPrj.Modules[man.Entry]; !ok {
		bag.Error("E7005", diag.Empty, "清单 `package.entry` 指向的模块 `"+man.Entry+"` 不存在",
			"入口模块", "检查 `entry` 的值，或补上入口模块文件")
		return res
	}

	mods := map[string]*ast.Module{}
	order := []string{}
	pkgOf := map[string]string{} // 模块名 → 包名
	loadOne := func(pkg, name, path, text string) {
		mod, mbag := parser.Parse(path, text)
		mergeBag(bag, mbag)
		mod.Name = name
		mods[name] = mod
		order = append(order, name)
		pkgOf[name] = pkg
	}
	for pkg, prj := range pkgs {
		for _, name := range prj.ModuleList {
			text, err := project.ReadSource(prj.Modules[name])
			if err != nil {
				bag.Error("E7001", diag.Empty, "读不到模块 `"+name+"`："+err.Error(), "文件", "")
				continue
			}
			loadOne(pkg, name, prj.Modules[name], text)
		}
	}
	// 排序，保证输出与模块图稳定（不影响语义）
	sort.Strings(order)

	bindTarget := bindgen.Target{Board: man.Board, Triple: res.Triple, PtrSize: man.PtrSize()}
	loaded := map[string]bool{}
	for _, name := range order {
		loaded[name] = true
	}
	for i := 0; i < len(order); i++ {
		name := order[i]
		mod := mods[name]
		if mod == nil {
			continue
		}
		pkg := pkgOf[name]
		for _, imp := range mod.Imports {
			// ---- C 绑定：`import ch|cc _.cpacks.<名字>` → bindgen 生成 `.ail` 模块 ----
			if imp.KindV == "ch" || imp.KindV == "cc" {
				if pkg != "_" {
					bag.Error("E7010", imp.Span, "依赖包里的 C 导入尚未支持",
						"cpack 只在根包里解析", "把 C 包留在根项目，或等依赖批次补齐")
					continue
				}
				bindName, bindText, ok := bindCpack(rootPrj, imp, res, bindTarget)
				if !ok || loaded[bindName] {
					continue
				}
				loadOne("_", bindName, bindName, bindText)
				loaded[bindName] = true
				continue
			}
			target, ok := resolveImportTarget(pkg, imp, graph)
			if !ok {
				bag.Error("E7010", imp.Span,
					"跨项目依赖 `"+imp.Root+"` 未在 `[dependencies]` 里声明", "依赖根",
					"在 `aila.toml` 里写 `[dependencies] "+imp.Root+" = { path = \"...\" }`（或 git 形式）")
				continue
			}
			if loaded[target] {
				continue
			}
			// 定位文件：优先在本包内找，`std.*` 走标准库目录
			path := ""
			if strings.HasPrefix(target, "std") {
				path = rootPrj.ModulePath(target)
			} else if targetPkg := packageOfModule(target); pkgs[targetPkg] != nil {
				path = pkgs[targetPkg].Modules[target]
			}
			if path == "" {
				bag.Error("E3001", imp.Span, "找不到被导入的模块 `"+target+"`", "导入路径",
					"检查目录/文件名是否与导入路径一致（路径即身份）")
				continue
			}
			text, err := project.ReadSource(path)
			if err != nil {
				bag.Error("E7001", diag.Empty, "读不到模块 `"+target+"`："+err.Error(), "文件", "")
				continue
			}
			loadOne(packageOfModule(target), target, path, text)
			loaded[target] = true
		}
	}
	// 锁文件：第一次解析或有变化时写回（依赖没变就不动它）
	if !bag.HasErrors() && (opts.UpdateDeps || !graph.Locked()) {
		if err := graph.Write(); err != nil {
			bag.Warning("W7021", diag.Empty, "写不了 `aila.lock`", err.Error(), "")
		}
	}
	res.Artifacts["modules"] = mods
	if bag.HasErrors() {
		return res
	}
	prog := sema.BuildProgramWithDeps(mods, order, man.Entry, graph.Edges, bag)
	res.Program = prog
	res.OK = !bag.HasErrors()
	return res
}

// nonRootPackages 列出依赖包名（不含根包）。
func nonRootPackages(g *deps.Graph) []string {
	out := make([]string, 0, len(g.Order))
	for _, pkg := range g.Order {
		if pkg != deps.RootPkg {
			out = append(out, pkg)
		}
	}
	return out
}

// resolveImportTarget 把导入语句解析成模块全名。
//
// 规则（designs/0001 §2.3 规则 1 + docs/rfcs/0004）：
//   - `_`      → 当前包自身的根（依赖包里的 `_` 指该依赖自己）
//   - `std`    → 标准库
//   - 依赖名   → 该包（必须是 pkg 在清单里声明的依赖）
func resolveImportTarget(pkg string, imp *ast.Import, g *deps.Graph) (string, bool) {
	join := func(root string) string {
		if len(imp.Path) == 0 {
			return root
		}
		return root + "." + strings.Join(imp.Path, ".")
	}
	switch imp.Root {
	case "_":
		return join(pkg), true
	case "std":
		return join("std"), true
	}
	for _, dep := range g.DepsOf(pkg) {
		if dep == imp.Root {
			return join(dep), true
		}
	}
	return "", false
}

// packageOfModule 从模块全名取包名（`_` / 依赖名 / `std`）。
func packageOfModule(module string) string {
	if module == "_" {
		return "_"
	}
	if i := strings.Index(module, "."); i > 0 {
		head := module[:i]
		if head == "_" {
			return "_"
		}
		return head
	}
	return module
}

// bindCpack 处理一条 C 导入：生成绑定模块文本（带缓存），并登记要一起编译的 cpack。
func bindCpack(prj *project.Project, imp *ast.Import, res *Result, target bindgen.Target) (string, string, bool) {
	bag := res.Diags
	if imp.Root != "_" {
		bag.Error("E7010", imp.Span, "C 导入必须以 `_` 为根（`_.cpacks.<名字>`）", "这里", "")
		return "", "", false
	}
	if len(imp.Path) < 2 || imp.Path[0] != "cpacks" {
		bag.Error("E7010", imp.Span, "C 导入路径必须形如 `_.cpacks.<名字>`", "cpack 约定",
			"把 C 包放到 `cpacks/<名字>/` 下")
		return "", "", false
	}
	name := imp.Path[1]
	dir, ok := prj.Cpacks[name]
	if !ok {
		bag.Error("E3001", imp.Span, "找不到 cpack 目录 `cpacks/"+name+"`", "这里",
			"目录名即 cpack 名（共置即关联）")
		return "", "", false
	}
	if !containsStr(res.Cpacks, name) {
		res.Cpacks = append(res.Cpacks, name)
	}
	if imp.KindV == "cc" {
		return "", "", false // `import cc` 只登记 C 源一起编译，不生成绑定
	}
	modName := "_.cpacks." + name
	br := bindgen.Run(name, dir, target, bag, "auto")
	if bag.HasErrors() {
		return "", "", false
	}
	text := br.Header(name, target) + "\n\n" + br.Text
	// 缓存：同一份头（哈希）+ 同一档位 + 同一映射规则版本才复用同一份绑定
	// （映射规则改了就换 schema，避免拿旧缓存装作成功）
	const bindgenSchema = "v2"
	header := fmt.Sprintf("%s-%08x.ail", name, fnv32(bindgenSchema+br.HashInput+br.Backend+target.Triple))
	cacheDir := filepath.Join(prj.Root, "build", ".cache", "bindgen")
	if err := os.MkdirAll(cacheDir, 0777); err == nil {
		_ = os.WriteFile(filepath.Join(cacheDir, header), []byte(text), 0666)
	}
	return modName, text, true
}

func fnv32(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

func containsStr(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

// mergeBag 把子诊断并入主诊断袋（保留诊断结构，不重排）。
func mergeBag(dst, src *diag.Bag) {
	if src == nil {
		return
	}
	dst.Items = append(dst.Items, src.Items...)
}

func importTarget(root string, path []string) string {
	switch root {
	case "_":
		if len(path) == 0 {
			return "_"
		}
		return "_." + strings.Join(path, ".")
	case "std":
		if len(path) == 0 {
			return "std"
		}
		return "std." + strings.Join(path, ".")
	}
	return ""
}

// Check 只做检查（不产产物）。
func Check(root string, opts Options) *Result {
	res := Load(root, opts)
	if res.OK {
		for _, name := range []string{"借用区间（NLL）", "区域配额静态证明", "热替换迁移完备性"} {
			res.Diags.Warning("W9001", diag.Empty, "`"+name+"` 检查待批次，本项未校验",
				"", "见 docs/plans/0001 里程碑表")
		}
	}
	return res
}

// Build 跑完整流水线，产出可执行文件。
func Build(root string, opts Options) *Result {
	res := Load(root, opts)
	if !res.OK {
		return res
	}
	res.OK = false // 只有走到最后一步（链接成功）才算构建通过
	bag := res.Diags
	man := res.Manifest
	prj := res.Project
	outDir := filepath.Join(prj.Root, "build", strings.ReplaceAll(res.Triple, "-", "_"))
	if man.Board != "" {
		outDir = filepath.Join(prj.Root, "build", man.Board)
	}
	if err := os.MkdirAll(outDir, 0777); err != nil {
		bag.Error("E7001", diag.Empty, "建不了构建目录："+err.Error(), "目录", "")
		return res
	}
	res.OutDir = outDir

	// ---- IR：逐模块 ----
	irs := res.GenerateIR()
	if bag.HasErrors() {
		return res
	}

	// ---- 目标文件 ----
	llvm := resolve.ResolveLLVM(man.LLVM)
	cc := resolve.ResolveCC(man.CC)
	if !cc.OK() {
		bag.Error("E8011", diag.Empty, "没有可用的 C 编译器 / 链接器", "工具链第二件",
			"安装 clang 或 MinGW-w64，并把 `clang`/`gcc` 放进 PATH")
		return res
	}
	// 先把 IR 全部落盘（`--emit-ir`），再逐个降成目标文件
	for _, name := range res.Program.Order {
		ir, ok := irs[name]
		if !ok || !strings.Contains(ir, "define ") {
			continue // 纯声明模块不产目标文件
		}
		if opts.EmitIR {
			if err := os.WriteFile(filepath.Join(outDir, safeName(name)+".ll"), []byte(ir), 0666); err != nil {
				bag.Error("E7001", diag.Empty, "写不了 IR 文件："+err.Error(), "目录", "")
				return res
			}
		}
	}
	var objects []string
	for _, name := range res.Program.Order {
		ir, ok := irs[name]
		if !ok || !strings.Contains(ir, "define ") {
			continue // 纯声明模块不产目标文件
		}
		safe := safeName(name)
		obj := filepath.Join(outDir, safe+".o")
		if err := toolchain.IRToObject(ir, obj, res.Triple, llvm); err != nil {
			bag.Error("E8010", diag.Empty, "LLVM 后端失败："+err.Error(), "这里",
				"确认 llc/clang 版本在受支持区间；或先 `aila build --emit-ir` 看 IR")
			return res
		}
		objects = append(objects, obj)
	}

	// ---- C 侧：运行时 + `import ch|cc` 的 cpack 源 ----
	cSources := []string{}
	if src := home.RuntimeSrc(); src != "" {
		cSources = append(cSources, src)
	} else {
		bag.Error("E8011", diag.Empty, "找不到运行时源码 `runtime/aila_rt.c`", "这里", "")
		return res
	}
	incFlags := []string{}
	for _, name := range res.Cpacks {
		dir := prj.Cpacks[name]
		cSources = append(cSources, bindgen.FindCSources(dir)...)
		incFlags = append(incFlags, "-I", filepath.Join(dir, "include"), "-I", dir)
	}
	cObjs, err := toolchain.CompileC(cSources, filepath.Join(outDir, "c"), cc, incFlags)
	if err != nil {
		bag.Error("E8010", diag.Empty, "编译 C 运行时失败："+err.Error(), "这里", "")
		return res
	}
	objects = append(objects, cObjs...)

	// ---- 链接 ----
	if man.Board != "" && man.Board != "host" {
		bag.Error("E8020", diag.Empty, "目标 `"+man.Board+"` 的交叉链接待批次（M2）",
			"M1 只有 host 能产出可执行文件", "目标文件已生成在 "+res.OutDir)
		res.Objects = objects
		return res
	}
	exeName := opts.OutName
	if exeName == "" {
		exeName = "app" + home.ExeSuffix()
	}
	extraLibs := []string{}
	for _, name := range res.Cpacks {
		extraLibs = append(extraLibs, bindgen.FindLibs(prj.Cpacks[name])...)
	}
	exe := filepath.Join(outDir, exeName)
	if err := toolchain.Link(append(objects, extraLibs...), exe, cc, nil); err != nil {
		bag.Error("E8010", diag.Empty, "链接失败："+err.Error(), "这里", "")
		return res
	}
	res.Exe = exe
	res.Objects = objects

	// ---- 报告 ----
	if opts.EmitJSON {
		writeReport(res, outDir)
	}
	res.OK = !bag.HasErrors()
	return res
}

// GenerateIR 为每个模块产出 LLVM IR 文本（不碰后端工具，判据与构建共用）。
func (res *Result) GenerateIR() map[string]string {
	bag := res.Diags
	layout := types.NewLayout(res.Manifest.PtrSize())
	irs := map[string]string{}
	for _, name := range res.Program.Order {
		u := res.Program.Unit(name)
		if u == nil || u.Check == nil || u.Mod == nil {
			continue
		}
		// 陷阱信息里的源文件路径由模块名派生（确定性，不随机器漂移）：
		// `_.main` → `main.aila`，`std.io` → `std/io.aila`。
		lw := codegen.New(u.Mod, name, u.Table, u.Sigs, u.Check, layout, res.Triple, res.Manifest.Entry)
		ir := lw.Generate()
		for _, what := range lw.Unsupported() {
			bag.Error("E9002", diag.Empty, what, "代码生成", "见 docs/plans/0001 里程碑表")
		}
		irs[name] = ir
	}
	res.IR = irs
	return irs
}

// writeReport 写出构建报告（`build/<目标>/aila-build.json`）。
func writeReport(res *Result, outDir string) {
	report := map[string]any{
		"aila_version": version(),
		"target":       res.Triple,
		"board":        res.Manifest.Board,
		"entry":        res.Manifest.Entry,
		"package":      res.Manifest.Name,
		"opt":          res.Manifest.Opt,
		"exe":          res.Exe,
		"objects":      len(res.Objects),
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(outDir, "aila-build.json"), append(data, '\n'), 0666)
}

func version() string {
	if data, err := os.ReadFile(filepath.Join(home.ShareRoot(), "aila", "VERSION")); err == nil {
		return strings.TrimSpace(string(data))
	}
	return "0.1.0-m1"
}

// moduleDisplayPath 把模块名映射成确定性的展示路径。
func moduleDisplayPath(name string) string {
	flat := strings.TrimPrefix(name, "_.")
	if name == "_" || name == "_.main" || flat == "" {
		if name == "std" || strings.HasPrefix(name, "std.") {
			flat = strings.ReplaceAll(strings.TrimPrefix(name, "std."), ".", "/")
			if flat == "" {
				return "std.aila"
			}
			return "std/" + flat + ".aila"
		}
		return "main.aila"
	}
	if strings.HasPrefix(name, "std.") {
		return "std/" + strings.ReplaceAll(flat, ".", "/") + ".aila"
	}
	return strings.ReplaceAll(flat, ".", "/")
}

func safeName(name string) string {
	s := strings.ReplaceAll(name, "_.", "")
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, "/", "_")
	return s
}

var _ = fmt.Sprintf
