// Package build 是构建流水线（docs/specs/0001 §5）。
//
// discover → manifest → parse → sema（多模块）→ lower(IR) → object → csrc → link → report
package build

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"aila/internal/ast"
	"aila/internal/bindgen"
	"aila/internal/chip"
	"aila/internal/codegen"
	"aila/internal/deps"
	"aila/internal/diag"
	"aila/internal/home"
	"aila/internal/manifest"
	"aila/internal/parser"
	"aila/internal/project"
	"aila/internal/registry"
	"aila/internal/resolve"
	"aila/internal/sema"
	"aila/internal/targets"
	"aila/internal/toolchain"
	"aila/internal/types"
)

// Options 是一次构建的可选参数（与 `aila build` 的命令行开关一一对应）。
type Options struct {
	Opt int
	// UpdateDeps 强制重新解析依赖并重写 `aila.lock`。
	UpdateDeps bool
	// Vendor 把解析好的依赖复制到 `vendor/`（离线构建）。
	Vendor   bool
	Triple   string
	LLVM     string
	CC       string
	EmitIR   bool
	EmitJSON bool
	// EmitBindings 把 bindgen 生成的 `.ail` 复制到产物目录（便于审计）
	EmitBindings bool
	OutName      string
	// GCSections 打开函数/数据分段 + 链接期 --gc-sections：**没被引用的模块代码
	// 不进产物**——配合 `#compile.match` 选驱动，未选中的驱动整块消失
	//（designs/0006 §2.6 / §4 的 P5）。
	GCSections bool
	// SizeDelta 在构建后报告体积：产物字节数 + 每个模块目标文件的大小，
	// 用来量化「换驱动 / 关功能」到底省了多少（designs/0006 §1.3 规则 8）。
	SizeDelta bool
	// LTO 把全部模块合成一个 LLVM 模块再 -O2：非 hot 模块之间允许内联
	//（designs/0001 §12.1 的铁律 7 边界只在 module hot 上），顺带让链接器
	// 更容易删掉死代码。默认关：体积/行为都会变，嵌入式目标需要逐块复验。
	LTO bool
}

// Result 是一次构建的结果。
type Result struct {
	OK       bool
	Diags    *diag.Bag
	Manifest *manifest.Manifest
	Project  *project.Project
	Program  *sema.Program
	Triple   string
	Exe      string
	// Deps 是跨项目依赖图；Packages 是本次参与构建的包名（不含根包的 `_` 在前）。
	Deps     *deps.Graph
	Packages []string
	Objects  []string
	IR       map[string]string
	Cpacks   []string
	// bindText 是本次构建生成的 C 绑定文本（cpack → .ail 内容）
	bindText  map[string]string
	OutDir    string
	Artifacts map[string]any
}

// Load 是「解析 + 语义」阶段：模块发现、清单、逐模块解析、多模块类型检查。
func Load(root string, opts Options) *Result {
	types.ResetErrorSpace() // 每次装载 = 一次编译：错误全局编号只由本工程的模块加载序决定
	bag := &diag.Bag{}
	res := &Result{Diags: bag, IR: map[string]string{}, bindText: map[string]string{}, Artifacts: map[string]any{}}
	man := manifest.Load(root, bag)
	if man == nil {
		return res
	}
	res.Manifest = man
	if opts.Triple != "" {
		man.Triple = opts.Triple
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
	man.Triple = res.Triple // 归一化：PtrSize 等后续按生效 triple 算

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

	bindTarget := bindgen.Target{Triple: res.Triple, PtrSize: man.PtrSize()}
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
				if !ok {
					// `import cc`（只编译 C 源、不要绑定）：登记一个空模块，
					// 让 `cc` 导入在 sema 侧能解析到一个「没有公开符号」的模块。
					if imp.KindV == "cc" {
						name := "_.cpacks." + imp.Path[1]
						if !loaded[name] {
							loadOne("_", name, name, "")
							loaded[name] = true
						}
					}
					continue
				}
				if loaded[bindName] {
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
	res.bindText[name] = text
	// 缓存：同一份头（哈希）+ 同一档位 + 同一映射规则版本才复用同一份绑定
	// （映射规则改了就换 schema，避免拿旧缓存装作成功）
	const bindgenSchema = "v2"
	header := fmt.Sprintf("%s-%08x.ail", name, fnv32(bindgenSchema+br.HashInput+br.Backend+target.Triple))
	cacheDir := filepath.Join(home.BuildDirFor(prj.Root), ".cache", "bindgen")
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
	outDir := filepath.Join(home.BuildDirFor(prj.Root), strings.ReplaceAll(res.Triple, "-", "_"))
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
	llvm := resolve.ResolveLLVMFor(res.Triple, man.LLVM)
	// 模块级 coro 路由：llc 不跑 coro 中端 pass（intrinsic 直落后端崩），
	// 含 llvm.coro 的 IR 必须走 clang `-x ir -c`（已实测双后端可行）。
	// clang 按「自报支持该目标架构」挑（esp8266 命中 esp-clang，宿主命中通版）。
	coroLLVM := llvm
	if strings.HasPrefix(llvm.Name, "llc") {
		if c, ok := resolve.ResolveClangForIR(res.Triple); ok && c.OK() {
			coroLLVM = c
			coroLLVM.Role = llvm.Role
		}
	}
	// C 绑定产物（`--emit-bindings`）：`.ail` 是机器生成的声明文件，`.aila` 才是源码
	if opts.EmitBindings {
		bindDir := filepath.Join(outDir, "bindgen")
		if err := os.MkdirAll(bindDir, 0o777); err == nil {
			names := make([]string, 0, len(res.bindText))
			for name := range res.bindText {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				_ = os.WriteFile(filepath.Join(bindDir, name+".ail"), []byte(res.bindText[name]), 0o666)
			}
		}
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
	// -mcpu：清单显式 cpu 优先，缺省从 triple 第二段推导（esp 系芯片名即 cpu）
	cpu := man.CPU
	if cpu == "" {
		cpu = targets.Parse(res.Triple).CPUDerived()
	}
	// SDK 模式：用户代码跨区调用（IRAM → IROM）超出直调范围，后端标志
	// 由 chip.json 的 sdks 数据提供（此时 linkEmbedded 还没跑，提前读一份）；
	// irTextSection 非空时用户 IR 的 define 全部落该段（esp8266 = .irom0.text，
	// IRAM 被 libpp/libmain 固占 23K，用户代码必须驻留 flash）
	var irFlags []string
	irTextSection := ""
	// 芯片级后端事实（与 SDK 无关）：esp8266 的 call0 ABI 在 llc -O0 下会对
	// 高寄存器压力函数报 "Cannot scavenge register"，必须带 chip.json 的 ir_flags
	//（含 -O1）。裸机与 SDK 两条路都要，否则显示栈这种大函数直接编不出来。
	if fl, ok := chipBackendFlags(res.Triple); ok {
		irFlags = append(irFlags, fl...)
	}
	if man.Sdk != "" {
		if fl, sec, ferr := sdkBackendFlags(res.Triple, man.Sdk); ferr == nil {
			irFlags = append(irFlags, fl...)
			irTextSection = sec
		}
	}
	// ---- 对象文件：逐模块（默认）或整程序合并（--lto）----
	//
	// 铁律 7 的边界只在 `module hot`（designs/0001 §12.1 / §7.8）：非 hot 模块之间
	// **允许**内联，而 AILA 的模块降的是各自的文本 IR，所以「允许」要靠把整程序合成
	// 一个 LLVM 模块再 -O2 才能兑现（工具链里没有 opt / llvm-link，文本合并就够）。
	// hot 模块的函数在 codegen 已经打了 noinline optnone，合并后优化器不会动它们。
	if opts.LTO {
		var parts [][2]string // [模块名, IR]
		for _, name := range res.Program.Order {
			ir, ok := irs[name]
			if !ok || !strings.Contains(ir, "define ") {
				continue
			}
			if irTextSection != "" {
				ir = placeIRInSection(ir, irTextSection)
			}
			parts = append(parts, [2]string{name, ir})
		}
		if len(parts) > 0 {
			merged := mergeModules(parts)
			if opts.EmitIR {
				if err := os.WriteFile(filepath.Join(outDir, "merged.ll"), []byte(merged), 0666); err != nil {
					bag.Error("E7001", diag.Empty, "写不了 IR 文件："+err.Error(), "目录", "")
					return res
				}
			}
			obj := filepath.Join(outDir, "app.merged.o")
			backend := llvm
			if strings.Contains(merged, "llvm.coro") {
				backend = coroLLVM
			}
			flags := append([]string{}, irFlags...)
			flags = append(flags, "-O2") // 整程序视图下才谈得上跨模块内联
			if opts.GCSections {
				flags = append(flags, sectionFlags(backend)...)
			}
			if err := toolchain.IRToObject(merged, obj, res.Triple, cpu, backend, flags...); err != nil {
				bag.Error("E8010", diag.Empty, "LLVM 后端失败（--lto）："+err.Error(), "这里",
					"先 `aila build --emit-ir --lto` 看 merged.ll；或退回逐模块构建（去掉 --lto）")
				return res
			}
			objects = append(objects, obj)
		}
	} else {
		for _, name := range res.Program.Order {
			ir, ok := irs[name]
			if !ok || !strings.Contains(ir, "define ") {
				continue // 纯声明模块不产目标文件
			}
			if irTextSection != "" {
				ir = placeIRInSection(ir, irTextSection)
			}
			safe := safeName(name)
			obj := filepath.Join(outDir, safe+".o")
			backend := llvm
			if strings.Contains(ir, "llvm.coro") {
				backend = coroLLVM
			}
			// --gc-sections 要求每个函数/全局各占一段，否则一个 .o 的 .text 是一整块、
			// 链接器删不掉函数（designs/0006 §2.6：设备选择靠「编译期剪枝 + 链接期 GC」）。
			flags := irFlags
			if opts.GCSections {
				flags = append(append([]string{}, irFlags...), sectionFlags(backend)...)
			}
			if err := toolchain.IRToObject(ir, obj, res.Triple, cpu, backend, flags...); err != nil {
				bag.Error("E8010", diag.Empty, "LLVM 后端失败："+err.Error(), "这里",
					"确认 llc/clang 版本在受支持区间；或先 `aila build --emit-ir` 看 IR")
				return res
			}
			objects = append(objects, obj)
		}
	}

	// ---- 嵌入式：平台 runtime + 交叉链接 + flash 镜像，全部在 aila 内部完成 ----
	if tgt := targets.Parse(res.Triple); tgt.Embedded {
		return res.linkEmbedded(objects, opts)
	}

	// ---- C 侧：运行时 + `import ch|cc` 的 cpack 源（宿主语义）----
	cc := resolve.ResolveCC(res.Triple, man.CC)
	if !cc.OK() {
		bag.Error("E8011", diag.Empty, "没有可用的 C 编译器 / 链接器", "工具链第二件",
			"安装 clang 或 MinGW-w64，并把 `clang`/`gcc` 放进 PATH")
		return res
	}
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
	exeName := opts.OutName
	if exeName == "" {
		exeName = "app" + home.ExeSuffix()
	}
	extraLibs := []string{}
	for _, name := range res.Cpacks {
		extraLibs = append(extraLibs, bindgen.FindLibs(prj.Cpacks[name])...)
	}
	// `[build] libs = [...]` → 系统库（-l 形态），如 Windows 走 Winsock 的 ws2_32
	var libFlags []string
	for _, l := range res.Manifest.Libs {
		libFlags = append(libFlags, "-l"+l)
	}
	exe := filepath.Join(outDir, exeName)
	linkFlags := libFlags
	if opts.GCSections {
		linkFlags = append(append([]string{}, libFlags...), "-Wl,--gc-sections")
	}
	if err := toolchain.Link(append(objects, extraLibs...), exe, cc, linkFlags); err != nil {
		bag.Error("E8010", diag.Empty, "链接失败："+explainLinkError(err.Error()), "这里", "")
		return res
	}
	res.Exe = exe
	res.Objects = objects

	// ---- 体积报告（--emit-size-delta）----
	if opts.SizeDelta {
		res.Artifacts["size"] = sizeReport(res, objects)
	}

	// ---- 报告 ----
	if opts.EmitJSON {
		writeReport(res, outDir)
	}
	res.OK = !bag.HasErrors()
	return res
}

// moduleInterfaces 给每个模块算一份「接口摘要」：只对签名取哈希，不看函数体。
//
// 这是 designs/0006 §5 记的「接口指纹最小可用形态」：热替换要判的是「只换了实现、
// 没动接口」，所以摘要必须对函数体（含剪枝结果）免疫。实现上用 IR 的 define/declare
// 行（去掉 `{` 之前的部分 = 调用约定与参数类型），排序后 sha256。
func moduleInterfaces(irs map[string]string) map[string]any {
	out := map[string]any{}
	for name, ir := range irs {
		var sigs []string
		for _, line := range strings.Split(ir, "\n") {
			if strings.HasPrefix(line, "define ") || strings.HasPrefix(line, "declare ") {
				head := strings.TrimSpace(line)
				if k := strings.IndexByte(head, '{'); k >= 0 {
					head = strings.TrimSpace(head[:k])
				}
				sigs = append(sigs, head)
			}
		}
		sort.Strings(sigs)
		sum := sha256.Sum256([]byte(strings.Join(sigs, "\n")))
		out[name] = map[string]any{
			"signatures": len(sigs),
			"digest":     hex.EncodeToString(sum[:8]),
		}
	}
	return out
}

// mergeModules 把各模块的文本 IR 合成**一个** LLVM 模块（`--lto` 的实现）。
//
// 为什么是文本合并而不是 llvm-link：随包工具链只有 llc / clang / llvm-nm，没有 opt 与
// llvm-link；而 AILA 的模块降出来就是文本 IR，拼起来再交给 llc -O2 效果相同。
// 三件必须做的事：
//  1. 只保留第一份 `target triple` / `target datalayout`（重复会让解析器报错）；
//  2. `; target:` 注释与模块级标记只留一份；
//  3. `attributes #N` 组按模块重编号（每个模块都从 #0 开始，直接拼会撞号）。
// 符号本身不需要改名：函数是 `aila_<模块>__<名>`、全局是 `g_<模块>_<名>`，天然不撞。
//
// 铁律 7 的边界只落在 `module hot` 上：那些函数在 codegen 已带 noinline optnone，
// 合并 + -O2 不会跨过它们（designs/0001 §12.1 / §7.8）。
func mergeModules(parts [][2]string) string {
	// 先扫一遍：哪些符号在整程序里**有定义**。合并后同一符号只剩一个模块，
	// 别处的 `declare` 必须删掉——LLVM 文本 IR 里 define 之后再出现 declare
	// 会报 "invalid redefinition of function"。
	defined := map[string]bool{}
	for _, part := range parts {
		for _, line := range strings.Split(part[1], "\n") {
			if strings.HasPrefix(line, "define ") {
				if sym := irSymbolOf(line); sym != "" {
					defined[sym] = true
				}
			}
		}
	}
	declared := map[string]bool{} // 已经输出过的 declare（同一符号只留第一份）
	var b strings.Builder
	offset := 0 // 已用掉的 attribute group 数
	for i, part := range parts {
		ir := part[1]
		// 属性组重编号：先把本模块的 `attributes #N = {...}` 掏出来，再从正文里
		// 把 `#N` 引用一起挪到新号段。
		renum := map[string]string{}
		var attrLines []string
		var body []string
		for _, line := range strings.Split(ir, "\n") {
			t := strings.TrimRight(line, " \t\r")
			switch {
			case i == 0 && (strings.HasPrefix(t, "target triple =") ||
				strings.HasPrefix(t, "target datalayout =") || strings.HasPrefix(t, "; target: ")):
				// 第一份的模块级事实保留
			case strings.HasPrefix(t, "target triple ="), strings.HasPrefix(t, "target datalayout ="),
				strings.HasPrefix(t, "; target: "):
				continue // 后续模块的重复行必须丢掉
			case strings.HasPrefix(t, "declare "):
				// 别处已有定义、或本符号已经声明过 → 声明必须消失（LLVM 文本 IR 不接受
				// 重复的 declare，也不接受 define 之后再 declare）。
				sym := irSymbolOf(t)
				if defined[sym] || declared[sym] {
					continue
				}
				declared[sym] = true
			case strings.HasPrefix(t, "attributes #"):
				rest := strings.TrimPrefix(t, "attributes #")
				num := rest
				if k := strings.IndexByte(rest, ' '); k >= 0 {
					num = rest[:k]
				}
				newNum := strconv.Itoa(offset + len(renum))
				renum[num] = newNum
				attrLines = append(attrLines, strings.Replace(t, "attributes #"+num, "attributes #"+newNum, 1))
				continue
			}
			body = append(body, line)
		}
		// 正文里的 `#N` 引用统一换号（只认纯数字组号，别的 `#` 不碰）
		text := strings.Join(body, "\n")
		// 字符串常量池的符号名是**模块内**编号（`@.str.0`），跨模块会撞名：
		// 按模块序重写一遍（定义与所有引用一起改，纯文本替换即可）。
		text = renameStrPool(text, i)
		for num, newNum := range renum {
			text = replaceAttrRef(text, num, newNum)
		}
		if i == 0 {
			b.WriteString(text)
		} else {
			b.WriteString("\n")
			b.WriteString(text)
		}
		b.WriteString("\n")
		for _, l := range attrLines {
			b.WriteString(l)
			b.WriteString("\n")
		}
		offset += len(renum)
	}
	return b.String()
}

// renameStrPool 把模块私有的字符串常量符号 `@.str.<n>` 改成 `@.str.<mod>.<n>`。
// codegen 在每个模块里都从 0 开始编号（模块内唯一即可），整程序合并时必须去重。
func renameStrPool(text string, mod int) string {
	const needle = "@.str."
	var b strings.Builder
	for {
		i := strings.Index(text, needle)
		if i < 0 {
			b.WriteString(text)
			return b.String()
		}
		b.WriteString(text[:i])
		rest := text[i+len(needle):]
		k := 0
		for k < len(rest) && rest[k] >= '0' && rest[k] <= '9' {
			k++
		}
		if k == 0 {
			b.WriteString(needle)
			text = rest
			continue
		}
		b.WriteString("@.str." + strconv.Itoa(mod) + "." + rest[:k])
		text = rest[k:]
	}
}

// irSymbolOf 从 `define`/`declare` 行里取被声明的全局符号名（"@aila_x__y"）。
func irSymbolOf(line string) string {
	at := strings.IndexByte(line, '@')
	if at < 0 {
		return ""
	}
	rest := line[at:]
	if k := strings.IndexAny(rest, "( ,"); k >= 0 {
		return rest[:k]
	}
	return rest
}

// replaceAttrRef 把文本里的 `#<num>` 换号，只匹配「# 后面直接是数字且后面不是数字/字母」，
// 避免误伤 `#compile` 之类的注释或 `#1`→`#10` 的前缀包含。
func replaceAttrRef(s, num, newNum string) string {
	if num == newNum {
		return s
	}
	var b strings.Builder
	needle := "#" + num
	for {
		i := strings.Index(s, needle)
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		end := i + len(needle)
		okEnd := end >= len(s) || !(s[end] >= '0' && s[end] <= '9') && !(s[end] >= 'a' && s[end] <= 'z')
		b.WriteString(s[:i])
		if okEnd {
			b.WriteString("#" + newNum)
		} else {
			b.WriteString(needle)
		}
		s = s[end:]
	}
}

// sectionFlags 返回「每个函数/全局各占一段」的后端标志（--gc-sections 的前提）。
// llc 用 -function-sections/-data-sections，clang 用 -f 前缀的两件套。
func sectionFlags(tool resolve.Tool) []string {
	if strings.HasPrefix(tool.Name, "llc") {
		return []string{"-function-sections", "-data-sections"}
	}
	return []string{"-ffunction-sections", "-fdata-sections"}
}

// sizeReport 统计产物与每个目标文件的大小（键 = 模块名，值 = 字节数）。
// 用来量化「换驱动 / 关功能」的体积差（designs/0006 §1.3 规则 8）。
func sizeReport(res *Result, objects []string) map[string]any {
	out := map[string]any{}
	total := int64(0)
	total += fileSize(res.Exe)
	out["exe"] = res.Exe
	out["exe_bytes"] = fileSize(res.Exe)
	per := map[string]int64{}
	for _, obj := range objects {
		n := fileSize(obj)
		per[strings.TrimSuffix(filepath.Base(obj), ".o")] = n
		total += n
	}
	out["objects_bytes"] = per
	out["objects_total"] = total - fileSize(res.Exe)
	return out
}

// fileSize 返回文件字节数（读不到 = 0）。
func fileSize(path string) int64 {
	if path == "" {
		return 0
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// linkEmbedded 完成嵌入式目标的收尾：读家族 chip.json（物理事实单一来源），
// 生成链接脚本与 chip_cfg.h，编平台 runtime（启动汇编 + 平台 C），随包 GCC
// 交叉链接（GCC 驱动 GNU ld——lld 没有 Xtensa 后端），打出 flash 镜像；
// esp32 系还编自研最小 bootloader 并一起产出（esp8266 由 ROM 直载，无需）。
// 项目里零平台文件：runtime 文件随 aila 分发在 share/aila/runtime/<家族>/
// （chip.json + startup.S + aila_rt_<家族>.c [+ boot.c]，docs/specs/0003 §1）。
func (res *Result) linkEmbedded(objects []string, opts Options) *Result {
	bag := res.Diags
	tgt := targets.Parse(res.Triple)
	family := tgt.RuntimeFamily()
	gccTriple := tgt.GCCTriple()

	rtDir := home.PlatformRuntime(family)
	if rtDir == "" {
		bag.Error("E8020", diag.Empty,
			"平台 `"+family+"` 尚未随包适配（找不到 runtime/"+family+"/）",
			"平台运行时",
			"适配一个平台 = 在 share/aila/runtime/ 下提供 chip.json、startup.S、aila_rt_<家族>.c（esp32 系再加 boot.c）")
		return res
	}
	chipDef, err := chip.Load(rtDir)
	if err != nil {
		bag.Error("E8020", diag.Empty, "平台 `"+family+"` 的 chip.json 不可用："+err.Error(),
			"平台运行时", "chip.json 是内存布局 / 镜像格式 / UART 常量的单一事实源")
		return res
	}

	cc := resolve.ResolveCrossCC(gccTriple, res.Manifest.CC)
	if !cc.OK() {
		bag.Error("E8011", diag.Empty, "没有可用的交叉 C 编译器 / 链接器", "工具链第二件", cc.Note)
		return res
	}

	// SDK 模式（aila.toml 的 [sdk] name）：入口层换 SDK 回调 + 预编译 .a 链接 +
	// SDK 变体链接脚本，irom 代码内嵌镜像（flash cache 线性映射直读）
	if res.Manifest.Sdk != "" {
		return res.linkSDK(chipDef, cc, rtDir, objects, opts)
	}

	rtOut := filepath.Join(res.OutDir, "rt")
	// 生成件：链接脚本 + UART 常量头（同一份 chip.json 渲染，物理事实同源）
	ldScript := filepath.Join(rtOut, family+".ld")
	if err := os.MkdirAll(rtOut, 0o755); err == nil {
		_ = os.WriteFile(ldScript, []byte(chipDef.LinkerScript()), 0o666)
		if h, herr := chipDef.ChipCfgH(); herr == nil {
			_ = os.WriteFile(filepath.Join(rtOut, "chip_cfg.h"), []byte(h), 0o666)
		}
	}

	// 平台 runtime：启动汇编 + 平台 C（GCC 一把编，.S 自动走预处理汇编）
	rtSrc := filepath.Join(rtDir, "aila_rt_"+family+".c")
	cObjs, err := toolchain.CompileC([]string{
		filepath.Join(rtDir, "startup.S"),
		rtSrc,
	}, rtOut, cc, append([]string{"-I", rtOut}, chipDef.GCCFlags...))
	if err != nil {
		bag.Error("E8010", diag.Empty, "编译平台运行时失败："+err.Error(), "这里", "")
		return res
	}

	// C 互操作（cpack）：`import ch|cc _.cpacks.<名字>` 的 C 源与头用同一套
	// 交叉 GCC 编（aila → C 的桥 bindgen 是全架构通用的），预编译库（.a）透传
	// 给链接——ld 从左到右解析符号，库必须排在用户 .o 之后。
	extraLibs := []string{}
	for _, name := range res.Cpacks {
		dir := res.Project.Cpacks[name]
		if cpSrcs := bindgen.FindCSources(dir); len(cpSrcs) > 0 {
			incFlags := append([]string{
				"-I", rtOut, "-I", filepath.Join(dir, "include"), "-I", dir,
			}, chipDef.GCCFlags...)
			cpObjs, cerr := toolchain.CompileC(cpSrcs, filepath.Join(res.OutDir, "cpack"), cc, incFlags)
			if cerr != nil {
				bag.Error("E8010", diag.Empty, "编译 cpack `"+name+"` 失败："+cerr.Error(), "这里", "")
				return res
			}
			objects = append(objects, cpObjs...)
		}
		extraLibs = append(extraLibs, bindgen.FindLibs(dir)...)
	}

	// 交叉链接：startup 最先，用户 .o 次之，cpack 库再次，libgcc 收尾（64 位除法 / 浮点助手）
	elfPath := filepath.Join(res.OutDir, "app.elf")
	all := append([]string{cObjs[0], cObjs[1]}, objects...)
	all = append(all, extraLibs...)
	embFlags := []string{"-nostdlib", "-T", ldScript, "-lgcc"}
	if opts.GCSections {
		embFlags = append(embFlags, "-Wl,--gc-sections")
	}
	if err := toolchain.Link(all, elfPath, cc, embFlags); err != nil {
		bag.Error("E8010", diag.Empty, "交叉链接失败："+err.Error(), "这里", "")
		return res
	}

	// flash 镜像（esptool elf2image 的原生实现，格式由 chip.json 的 image.kind 决定）
	binPath := filepath.Join(res.OutDir, "app.bin")
	if err := chipDef.Pack(elfPath, binPath); err != nil {
		bag.Error("E8010", diag.Empty, "打包 flash 镜像失败："+err.Error(), "这里", "")
		return res
	}
	res.Exe = binPath

	// esp32 系：ROM 只认二级 bootloader——用同一套 GCC 编自研最小 bootloader，
	// 与 app 一起产出（烧录偏移来自 chip.json 的 boot 段）
	if chipDef.Boot != nil {
		bootDir := filepath.Join(res.OutDir, "boot")
		if err := res.buildBootloader(chipDef, cc, rtDir, bootDir); err != nil {
			bag.Error("E8010", diag.Empty, "构建 bootloader 失败："+err.Error(), "这里", "")
			return res
		}
	}

	res.Objects = append(objects, cObjs...)
	if opts.SizeDelta {
		res.Artifacts["size"] = sizeReport(res, res.Objects)
	}
	if opts.EmitJSON {
		writeReport(res, res.OutDir)
	}
	res.OK = !bag.HasErrors()
	return res
}

// chipBackendFlags 读家族 chip.json 的**芯片级** IR → .o 标志（与 SDK 无关）。
// 读不到（宿主目标 / 平台未适配）返回 ok=false，调用方静默跳过。
func chipBackendFlags(triple string) ([]string, bool) {
	family := targets.Parse(triple).RuntimeFamily()
	if family == "" {
		return nil, false
	}
	rtDir := home.PlatformRuntime(family)
	if rtDir == "" {
		return nil, false
	}
	def, err := chip.Load(rtDir)
	if err != nil {
		return nil, false
	}
	return def.IRFlags, len(def.IRFlags) > 0
}

// sdkBackendFlags 从家族 chip.json 读 SDK 的后端标志（IR → .o 用），
// 在 Build 主流程里提前于 linkEmbedded 调用；读不到（平台/SDK 未适配）返回错误，
// 由调用方静默跳过——后续 linkSDK 会给出正式诊断。
func sdkBackendFlags(triple, sdkName string) ([]string, string, error) {
	family := targets.Parse(triple).RuntimeFamily()
	rtDir := home.PlatformRuntime(family)
	if rtDir == "" {
		return nil, "", fmt.Errorf("平台 %q 运行时不存在", family)
	}
	def, err := chip.Load(rtDir)
	if err != nil {
		return nil, "", err
	}
	s, err := def.SDK(sdkName)
	if err != nil {
		return nil, "", err
	}
	return s.IRFlags, s.IRTextSection, nil
}

// placeIRInSection 把模块里所有 define 落进指定段（esp8266 SDK 模式的
// .irom0.text：用户代码驻留 flash，IRAM 留给 libpp/libmain 热代码与平台 C）。
// 只改写行首 define、以 " {" 收尾的行——IR 里唯一有段属性需求的就是函数体。
func placeIRInSection(ir, section string) string {
	suffix := ` section "` + section + `" {`
	var out []string
	for _, line := range strings.Split(ir, "\n") {
		if strings.HasPrefix(line, "define ") && strings.HasSuffix(line, " {") {
			line = strings.TrimSuffix(line, " {") + suffix
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// linkSDK 是 SDK 模式的收尾：平台入口层（aila_rt + SDK 回调层，无 startup.S——
// 入口符号 call_user_start 由 libmain.a 提供）用随包 GCC 交叉编译，SDK 变体
// 链接脚本（官方 eagle.app.v6.ld 语义 + AILA 堆窗口）由 chip.json 生成，
// 预编译 .a 序列在 --start-group 内循环解析符号（libc/libgcc 也取 SDK 自带的，
// 与 SDK 同源），镜像把 irom 段内嵌到其 flash 映射偏移（官方 gen_appbin 手法）。
func (res *Result) linkSDK(chipDef chip.Chip, cc resolve.Tool, rtDir string, objects []string, opts Options) *Result {
	bag := res.Diags
	family := chipDef.Family
	sdkName := res.Manifest.Sdk

	sdkDef, err := chipDef.SDK(sdkName)
	if err != nil {
		bag.Error("E8020", diag.Empty, err.Error(), `aila.toml 的 [sdk] name`,
			"可用 SDK 见 share/aila/runtime/"+family+"/chip.json 的 sdks 段")
		return res
	}

	// SDK 安装目录：注册表 sdk/<名字> 键。专有许可的 SDK 无下载源，
	// 只能本地 zip 安装（aila toolchain install sdk/<名字> <zip>）
	reg := registry.Load()
	inst, ok := reg.Installed[registry.Key("sdk", sdkName)]
	if !ok {
		bag.Error("E8011", diag.Empty, "SDK `"+sdkName+"` 未安装", "工具链",
			"本地 zip 安装：`aila toolchain install sdk/"+sdkName+" <SDK zip 路径>`")
		return res
	}
	sdkRoot := filepath.Join(home.Root(), filepath.FromSlash(inst.Path))
	romLd := filepath.Join(sdkRoot, "ld", sdkDef.RomLD)
	if _, err := os.Stat(romLd); err != nil {
		bag.Error("E8020", diag.Empty, "SDK `"+sdkName+"` 目录不完整（缺 "+sdkDef.RomLD+"）",
			"SDK 安装", "重装：`aila toolchain install sdk/"+sdkName+" <zip>`")
		return res
	}

	rtOut := filepath.Join(res.OutDir, "rt")
	if err := os.MkdirAll(rtOut, 0o755); err != nil {
		bag.Error("E7001", diag.Empty, "建不了构建目录："+err.Error(), "目录", "")
		return res
	}
	// 生成件：SDK 变体链接脚本 + UART 常量头 + 分区表头（同一份 chip.json 渲染）
	ldScript := filepath.Join(rtOut, family+"-sdk.ld")
	_ = os.WriteFile(ldScript, []byte(chipDef.SDKLinkerScript(sdkDef, romLd)), 0o666)
	if h, herr := chipDef.ChipCfgH(); herr == nil {
		_ = os.WriteFile(filepath.Join(rtOut, "chip_cfg.h"), []byte(h), 0o666)
	}
	if h, herr := chipDef.SDKCfgH(sdkDef); herr == nil {
		_ = os.WriteFile(filepath.Join(rtOut, "nosdk_cfg.h"), []byte(h), 0o666)
	}

	// 平台入口层：aila_rt + SDK 回调层（runtime 名来自 chip.json 数据）
	rtSrc := filepath.Join(rtDir, "aila_rt_"+family+".c")
	// 平台 runtime 必须留在 IRAM：SDK 启动链（user_pre_init / user_init）在 flash
	// cache 就绪前就被调用，丢进 .irom0.text 会静默重启。用户侧 cpack 才进 flash。
	flags := append(append([]string{"-I", rtOut}, chipDef.GCCFlags...), sdkDef.GCCFlags...)
	// SDK 会在 user_init 之前把 CPU 从复位默认的 52MHz 切到 80MHz，所以 UART 分频
	// 与 ccount 换算必须按 SDK 的主频重算——chip_cfg.h 里的同名宏有 #ifndef 守卫，
	// 这里以 -D 覆盖（chip.json 顶层记录的是裸机真值 52MHz）。
	if sdkDef.CPUHz != 0 {
		flags = append(flags,
			fmt.Sprintf("-DAILA_UART_CLK_HZ=%du", sdkDef.CPUHz),
			fmt.Sprintf("-DAILA_CPU_HZ=%du", sdkDef.CPUHz))
	}
	cObjs, err := toolchain.CompileC([]string{
		rtSrc,
		filepath.Join(rtDir, sdkDef.Runtime+".c"),
	}, rtOut, cc, flags)
	if err != nil {
		bag.Error("E8010", diag.Empty, "编译平台运行时失败："+err.Error(), "这里", "")
		return res
	}

	// 交叉链接：runtime/cpack/用户 .o 在前，cpack 透传 .a 次之，
	// SDK .a 序列（--start-group 循环解析）在后；libc/libgcc 取 SDK 自带
	//（-L 优先命中，与 SDK 的 GCC 同源）
	elfPath := filepath.Join(res.OutDir, "app.elf")
	// Windows 上把路径换成斜杠：ld 不认反斜杠，`-T C:\\...\\x.ld` 会报 Invalid argument
	linkFlags := []string{"-nostdlib", "-T", filepath.ToSlash(ldScript), "-u", "call_user_start"}
	extraLibs := []string{}
	// C 互操作（cpack）：C 源用同一套交叉 GCC 编（SDK 头路径追加，桥接层
	// include SDK 头），预编译库（.a）透传给链接
	for _, name := range res.Cpacks {
		dir := res.Project.Cpacks[name]
		if cpSrcs := bindgen.FindCSources(dir); len(cpSrcs) > 0 {
			incFlags := append([]string{
				"-I", rtOut,
				"-I", filepath.Join(dir, "include"), "-I", dir,
				"-I", filepath.ToSlash(filepath.Join(sdkRoot, "include")),
			}, chipDef.GCCFlags...)
			incFlags = append(incFlags, sdkDef.GCCFlags...)
			incFlags = append(incFlags, "-ffunction-sections", "-fdata-sections")
			cpObjs, cerr := toolchain.CompileC(cpSrcs, filepath.Join(res.OutDir, "cpack"), cc, incFlags)
			if cerr != nil {
				bag.Error("E8010", diag.Empty, "编译 cpack `"+name+"` 失败："+cerr.Error(), "这里", "")
				return res
			}
			cObjs = append(cObjs, cpObjs...)
		}
		extraLibs = append(extraLibs, bindgen.FindLibs(dir)...)
	}
	linkFlags = append(linkFlags, extraLibs...)
	linkFlags = append(linkFlags,
		"-L", filepath.ToSlash(filepath.Join(sdkRoot, "lib")),
		"-Wl,--start-group",
	)
	for _, lib := range sdkDef.Libs {
		linkFlags = append(linkFlags, "-l"+lib)
	}
	linkFlags = append(linkFlags, "-lc", "-lgcc", "-Wl,--end-group",
		"-Wl,-Map="+filepath.ToSlash(filepath.Join(res.OutDir, "app.map")))
	if err := toolchain.Link(cObjs, elfPath, cc, append(append([]string{}, objects...), linkFlags...)); err != nil {
		bag.Error("E8010", diag.Empty, "交叉链接失败："+err.Error(), "这里",
			"检查 [sdk] name 与已安装 SDK 是否匹配；缺符号时看该符号属于哪个 SDK 库")
		return res
	}

	// flash 镜像：RAM 段进 0xE9 头，irom 段内嵌到其 flash 映射偏移
	binPath := filepath.Join(res.OutDir, "app.bin")
	if err := chipDef.PackSDK(elfPath, binPath, sdkDef); err != nil {
		bag.Error("E8010", diag.Empty, "打包 flash 镜像失败："+err.Error(), "这里", "")
		return res
	}
	res.Exe = binPath
	res.Objects = append(objects, cObjs...)
	if opts.EmitJSON {
		writeReport(res, res.OutDir)
	}
	res.OK = !bag.HasErrors()
	return res
}

// buildBootloader 编链接自研最小 bootloader 并打包（esp32 系专用）。
func (res *Result) buildBootloader(chipDef chip.Chip, cc resolve.Tool, rtDir, bootDir string) error {
	bootLd := filepath.Join(bootDir, "boot.ld")
	ldText, err := chipDef.BootLinkerScript()
	if err != nil {
		return err
	}
	cfgText, err := chipDef.BootCfgH()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(bootDir, 0o755); err != nil {
		return err
	}
	_ = os.WriteFile(bootLd, []byte(ldText), 0o666)
	_ = os.WriteFile(filepath.Join(bootDir, "boot_cfg.h"), []byte(cfgText), 0o666)

	extra := append([]string{"-I", bootDir, "-I", filepath.Join(res.OutDir, "rt")}, chipDef.GCCFlags...)
	bootObjs, err := toolchain.CompileC([]string{filepath.Join(rtDir, "boot.c")}, bootDir, cc, extra)
	if err != nil {
		return err
	}
	bootElf := filepath.Join(bootDir, "bootloader.elf")
	if err := toolchain.Link(bootObjs, bootElf, cc,
		[]string{"-nostdlib", "-T", bootLd, "-lgcc"}); err != nil {
		return err
	}
	bootBin := filepath.Join(bootDir, "bootloader.bin")
	if err := chipDef.Pack(bootElf, bootBin); err != nil {
		return err
	}
	res.Artifacts["bootloader"] = bootBin
	res.Artifacts["boot_offset"] = chipDef.Boot.BootOffset
	res.Artifacts["app_offset"] = chipDef.Boot.AppOffset
	return nil
}

// programHasModule 判断程序模块图里是否有某模块（std.os 被任意模块 import 时
// Load 会把它加载进来，Order 里即出现；用于入口转存决策，见 codegen.genEntryWrapper）。
func programHasModule(prog *sema.Program, name string) bool {
	for _, m := range prog.Order {
		if m == name {
			return true
		}
	}
	return false
}

// explainLinkError 把链接器的 undefined symbol 报错翻译成用户能行动的提示：
// std.os 的原语只在宿主 runtime（runtime/aila_rt.c）里实现——嵌入式 target
// import 了 std.os（或其依赖链）就会在这里爆出来（designs/0003 §4.1）。
func explainLinkError(msg string) string {
	for _, sym := range []string{
		"aila_rt_args_save", "aila_rt_args_count", "aila_rt_args_get",
		"aila_rt_env_get", "aila_rt_env_set", "aila_rt_exit",
		"aila_rt_file_open", "aila_rt_file_read", "aila_rt_file_write",
		"aila_rt_file_close", "aila_rt_file_flush",
	} {
		if strings.Contains(msg, sym) {
			return msg + "（std.os 需要 OS target：`" + sym +
				"` 只在宿主 runtime 实现，嵌入式芯片上没有进程 / 文件 / 环境变量——去掉对 std.os 的 import）"
		}
	}
	return msg
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
		lw := codegen.New(u.Mod, name, u.Table, u.Sigs, u.Check, layout, res.Triple,
			res.Manifest.Entry, programHasModule(res.Program, "std.os"))
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
		"entry":        res.Manifest.Entry,
		"package":      res.Manifest.Name,
		"opt":          res.Manifest.Opt,
		"exe":          res.Exe,
		"objects":      len(res.Objects),
	}
	// 编译期剪枝清单（designs/0006 §1.3 规则 7）：每条都是「条件编译期已知 → 该分支被剪掉」，
	// 用来回答"我这个 #compile.if 到底生效没有、省了哪一支"。
	if res.Program != nil {
		var prunes []map[string]any
		for _, name := range res.Program.Order {
			u := res.Program.Unit(name)
			for _, rec := range u.Prunes() {
				prunes = append(prunes, map[string]any{
					"module": rec.Module,
					"kind":   rec.Kind,
					"line":   rec.Line,
					"value":  rec.Value,
					"branch": rec.Branch,
				})
			}
		}
		report["compile_time_prunes"] = prunes
	}
	// 模块接口摘要（designs/0006 §5 的「接口指纹」最小可用形态）：
	// 只对**签名**取摘要——`define` / `declare` 行去掉函数体，排序后哈希。
	// 于是：换 `#compile.match` 的 const 取值（剪掉另一支）摘要不变（剪枝不动 ABI），
	// 而函数签名/增删导出一定会变。热替换模块将来用它做「只换实现」的合法性判定；
	// 非 hot 模块用它做重编译的可观测性。真正的分派表/state 槽在 M2。
	if res.IR != nil {
		report["module_interfaces"] = moduleInterfaces(res.IR)
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
