// Package codegen 把已定型的 AST 降成 LLVM IR 文本（docs/specs/0002 §7）。
//
// 纪律（与 Python 参考实现逐字节对齐）：
//   - 只用**不透明指针**（LLVM 15+ 硬要求），不写 `target datalayout`；
//   - 只发保守、跨版本稳定的指令与 intrinsic（`*.with.overflow` / `sdiv` / `ashr` …）；
//   - 陷阱（溢出 / 除零 / 移位 / 越界）由编译器插入检查，全部汇到 `aila_rt_trap`；
//   - 聚合值一律走内存槽（alloca + load/store）。
package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"aila/internal/ast"
	"aila/internal/diag"
	"aila/internal/sema"
	"aila/internal/types"
)

// Lowerer 生成单个模块的 IR。
type Lowerer struct {
	mod        *ast.Module
	modName    string
	srcPath    string
	table      *sema.Table
	sigs       *sema.Sigs
	check      *sema.Checker
	layout     *types.Layout
	triple     string
	entry      string
	ptrSize    int
	progUsesOS bool // 程序（全模块图）是否 import std.os —— 决定 C 入口是否转存 argc/argv

	globals        []string
	consts         map[string]constVal
	constNames     []string
	strSyms        map[string]string
	strLines       []string
	declares       map[string]string
	declareOrder   []string
	defined        map[string]bool
	pendingGlobals []string
	globalsSeen    map[string]bool
	// globalInits 是模块级 let/var/state 的初始化表达式（名字 → 语法节点），
	// 全局变量的 LLVM 初值从它来（字面量才编得出来，其余退化成 0）。
	globalInits map[string]ast.Node
	funcs          []string
	hasCoroutines  bool            // 模块含协程函数 → 需要发 presplitcoroutine 属性组
	threadThunks   map[string]bool // 线程入口解包 thunk（每签名一份）

	// unsupported 记录本批次尚未实现的构造（Go 侧不静默丢代码）。
	unsupported []string
}

type constVal struct {
	lt  string
	lit string
}

// New 构造 Lowerer（entry = 入口模块名，如 "_.main"；progUsesOS = 程序是否
// import 了 std.os，见 genEntryWrapper）。
//
// 陷阱信息里的源文件路径不取调用方传进来的路径，而是由**模块名派生**
// （`_.main` → `main.aila`，`std.io` → `std/io.aila`）：产物跨机器一致（缺陷 13）。
func New(mod *ast.Module, modName string, table *sema.Table, sigs *sema.Sigs,
	check *sema.Checker, layout *types.Layout, triple, entry string, progUsesOS bool) *Lowerer {
	return &Lowerer{
		mod: mod, modName: modName, srcPath: DisplayPath(modName), table: table, sigs: sigs, check: check,
		layout: layout, triple: triple, entry: entry, progUsesOS: progUsesOS,
		// 指针宽度必须跟目标走：xtensa/32 位目标上 USize 是 4 字节。
		// 这里曾经写死 8，所有嵌入式目标的 USize/指针都按 64 位出 IR（ABI 直接错）。
		ptrSize:      layout.PtrSize,
		consts:       map[string]constVal{},
		strSyms:      map[string]string{},
		declares:     map[string]string{},
		defined:      map[string]bool{},
		globalsSeen:  map[string]bool{},
		threadThunks: map[string]bool{},
	}
}

// Generate 产出整份 IR 文本。
func (l *Lowerer) Generate() string {
	l.collectConsts()
	l.collectGlobalInits()
	for _, d := range l.mod.AllDecls() {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Body != nil && !fn.IsExtern {
			l.defined[symbolFor(l.modName, fn.Name, fn.OwnerType)] = true
		}
	}
	for _, d := range l.mod.AllDecls() {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		l.funcs = append(l.funcs, l.genFunction(fn))
	}
	l.genEntryWrapper()

	head := []string{
		"; AILA module: " + l.modName,
		"; target: " + l.triple,
		`target triple = "` + l.triple + `"`,
		"",
	}
	head = append(head, l.globals...)
	head = append(head, l.strLines...)
	if len(l.consts) > 0 {
		head = append(head, "")
	}
	head = append(head, l.pendingGlobals...)
	head = append(head, "")
	for _, sym := range l.declareOrder {
		head = append(head, "declare "+l.declares[sym])
	}
	if len(l.declareOrder) > 0 {
		head = append(head, "")
	}
	out := strings.Join(append(head, l.funcs...), "\n")
	if l.hasCoroutines {
		// 协程函数统一引用属性组 #0（designs/0001 §10.2：无栈协程走 llvm.coro）。
		// 拼写必须是 unquoted 的 `presplitcoroutine`——引号串属性 coro pass 不识别。
		out += "\nattributes #0 = { presplitcoroutine }\n"
	}
	return strings.TrimRight(out, " \t\r\n") + "\n"
}

// lenLT 是「长度 / 下标」的 LLVM 整数类型：32 位目标 → i32，64 位 → i64。
// 视图的 len 字段、#comprun.len 的产出、for-in 的索引都用它——必须和 USize 同宽，
// 否则 32 位目标上 icmp/gep/实参处处宽度对不上（历史 bug：这里曾写死 i64）。
func (l *Lowerer) lenLT() string { return "i" + strconv.Itoa(l.ptrSize*8) }

// viewLT 是视图 / Text 的 LLVM 类型（胖指针：数据指针 + 长度）。
func (l *Lowerer) viewLT() string { return "{ ptr, " + l.lenLT() + " }" }

// Unsupported 返回本批次尚未实现的构造清单（空表示全量覆盖）。
func (l *Lowerer) Unsupported() []string { return l.unsupported }

// collectConsts 收集模块级 const，并生成其 IR 表示（字符串进常量池）。
func (l *Lowerer) collectConsts() {
	order := l.sigs.ConstOrder
	if len(order) == 0 {
		order = sortedKeys(l.sigs.Consts)
	}
	for _, name := range order {
		ci := l.sigs.Consts[name]
		t := ci.Type
		switch {
		case isTextType(t):
			text := fmt.Sprint(ci.Value)
			sym := l.stringConst(text)
			l.consts[name] = constVal{l.viewLT(),
				"{ ptr " + sym + ", " + l.lenLT() + " " + strconv.Itoa(len(text)) + " }"}
		case types.IsFloat(t):
			l.consts[name] = constVal{l.llty(t),
				l.llty(t) + " " + floatLit(toFloat(ci.Value), llvmFloatBits(l.llty(t)))}
		case isBoolType(t):
			l.consts[name] = constVal{"i1", "i1 " + boolLit(toBool(ci.Value))}
		default:
			bits := types.IntWidth(t, l.ptrSize)
			if bits == 0 {
				bits = 32
			}
			var v int64
			if iv, ok := asInt(ci.Value); ok {
				if bits >= 64 {
					v = iv
				} else {
					v = iv & ((int64(1) << uint(bits)) - 1)
				}
			}
			l.consts[name] = constVal{l.llty(t), l.llty(t) + " " + strconv.FormatInt(v, 10)}
		}
	}
}

// --------------------------------------------------------------------------- //
// 名字与常量
// --------------------------------------------------------------------------- //

// escName 把非法标识符字符转义成 `_<hex>_`（与参考实现的 `_esc` 一致）。
func escName(name string) string {
	var b strings.Builder
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '$' || ch == '.' {
			b.WriteRune(ch)
			continue
		}
		b.WriteString("_" + strconv.FormatInt(int64(ch), 16) + "_")
	}
	return b.String()
}

// symbolFor 生成函数符号名：`aila_<模块扁平名>__[Owner__]名字`。
func symbolFor(module, name, owner string) string {
	flat := strings.ReplaceAll(module, "_.", "")
	flat = strings.ReplaceAll(flat, ".", "_")
	if flat == "" {
		flat = "root"
	}
	if owner != "" {
		return escName("aila_" + flat + "__" + owner + "__" + name)
	}
	return escName("aila_" + flat + "__" + name)
}

// stringConst 把字节串放进常量池，返回符号名（去重）。
func (l *Lowerer) stringConst(data string) string {
	if sym, ok := l.strSyms[data]; ok {
		return sym
	}
	sym := "@.str." + strconv.Itoa(len(l.strSyms))
	l.strSyms[data] = sym
	l.strLines = append(l.strLines, fmt.Sprintf("%s = private unnamed_addr constant [%d x i8] c\"%s\\00\", align 1",
		sym, len(data)+1, cEscape(data)))
	return sym
}

// cEscape 按 LLVM 常量串规则转义。
func cEscape(data string) string {
	var b strings.Builder
	for i := 0; i < len(data); i++ {
		c := data[i]
		switch {
		case c == '"':
			b.WriteString(`\22`)
		case c == '\\':
			b.WriteString(`\5C`)
		case c >= 32 && c < 127:
			b.WriteByte(c)
		default:
			fmt.Fprintf(&b, `\%02X`, c)
		}
	}
	return b.String()
}

// DisplayPath 把模块名映射成确定性的展示路径（进 IR 的 trap 位置）。
func DisplayPath(modName string) string {
	if modName == "std" {
		return "std.aila"
	}
	if strings.HasPrefix(modName, "std.") {
		return "std/" + strings.ReplaceAll(strings.TrimPrefix(modName, "std."), ".", "/") + ".aila"
	}
	if modName == "_" || modName == "_.main" || modName == "" {
		return "main.aila"
	}
	if strings.HasPrefix(modName, "_.") {
		return strings.ReplaceAll(strings.TrimPrefix(modName, "_."), ".", "/")
	}
	return modName
}

// fileConst 返回源文件名字符串常量（陷阱信息用）。
func (l *Lowerer) fileConst() string { return l.stringConst(l.srcPath) }

// symbolOf 返回函数符号名：extern 用 C 名，其余用 AILA 修饰名。
//
// 跨模块调用要用**声明模块**的名字做修饰（`io.println` → `aila_std_io__println`）。
func (l *Lowerer) symbolOf(sig *sema.FuncSig) string {
	if sig.IsExtern {
		return sig.Name
	}
	mod := sig.Module
	if mod == "" {
		mod = l.modName
	}
	return symbolFor(mod, sig.Name, sig.Owner)
}

// ensureDeclare 登记外部声明（已定义的符号不重复声明）。
func (l *Lowerer) ensureDeclare(sym, sigText string) {
	if l.defined[sym] {
		return
	}
	if _, ok := l.declares[sym]; ok {
		return
	}
	l.declares[sym] = sigText
	l.declareOrder = append(l.declareOrder, sym)
}

// ensureThreadThunk 产出线程入口的私有解包 thunk（每签名一份，designs/0001 §10.8）：
// `define internal ptr @<sym>__thunk(ptr %arg)` —— 按编译期布局从打包块解包实参
// → call 真函数（返回值丢弃，fire-and-forget）→ 释放打包块 → `ret ptr null`。
// 签名对齐 C 侧 `void* (*)(void*)` 蹦床（runtime aila_rt.c）。
func (l *Lowerer) ensureThreadThunk(sig *sema.FuncSig) string {
	sym := l.symbolOf(sig) + "__thunk"
	if l.threadThunks[sym] {
		return sym
	}
	l.threadThunks[sym] = true
	var b []string
	add := func(s string) { b = append(b, "  "+s) }
	add("define internal ptr @" + sym + "(ptr %arg) {")
	callArgs := make([]string, 0, len(sig.Regions)+len(sig.Params))
	for range sig.Regions {
		callArgs = append(callArgs, "ptr null") // 区域参数禁止进线程（sema E5002），死代码兜底
	}
	off := 0
	for i, p := range sig.Params {
		plt := l.paramLLType(sig, p.Type)
		align := l.layout.Align(p.Type)
		if align < 1 {
			align = 1
		}
		off = (off + align - 1) &^ (align - 1)
		gp := "%arg"
		if off > 0 {
			gp = "%p" + strconv.Itoa(i)
			add(gp + " = getelementptr i8, ptr %arg, i64 " + strconv.Itoa(off))
		}
		lv := "%v" + strconv.Itoa(i)
		add(lv + " = load " + plt + ", ptr " + gp)
		callArgs = append(callArgs, plt+" "+lv)
		off += l.layout.Size(p.Type)
	}
	total := off
	retLt, params := l.funcLLType(sig)
	callee := l.symbolOf(sig)
	l.ensureDeclare(callee, retLt+" @"+callee+"("+strings.Join(params, ", ")+")")
	// C 变参 extern 入口：文本 IR 要求 call 带完整函数类型注解（同 lowerCall）。
	callTy := ""
	if sig.Variadic {
		callTy = retLt + " (" + strings.Join(params, ", ") + ")"
	}
	if retLt == "void" {
		add("call " + callTySep(retLt, callTy, "@"+callee) + "(" + strings.Join(callArgs, ", ") + ")")
	} else {
		add("%rv = call " + callTySep(retLt, callTy, "@"+callee) + "(" + strings.Join(callArgs, ", ") + ")")
	}
	if total > 0 {
		l.ensureDeclare("aila_rt_thread_args_free", "void @aila_rt_thread_args_free(ptr)")
		add("call void @aila_rt_thread_args_free(ptr %arg)")
	}
	add("ret ptr null")
	add("}")
	l.funcs = append(l.funcs, strings.Join(b, "\n")+"\n")
	return sym
}

// ensureGlobal 登记模块级全局变量（每个符号只发一次）。
//
// 初值取自该全局声明的初始化表达式（`globalInits`，见 Generate 里的收集）：
// 只有字面量能编成 LLVM 聚合初值，其它情况退化成 zeroinitializer（与旧行为一致）。
func (l *Lowerer) ensureGlobal(name, sym, lt string) {
	if l.globalsSeen[sym] {
		return
	}
	l.globalsSeen[sym] = true
	l.pendingGlobals = append(l.pendingGlobals,
		sym+" = internal global "+lt+" "+l.constLiteral(lt, l.globalInits[name])+", align 8")
}

// collectGlobalInits 记下模块级 let/var/state 的初始化表达式，供全局初值使用。
//
// 之前全局变量一律零初始化、也不生成启动期 store——模块级数据表（字库、命令表）
// 会在运行期整片为 0。现在把这些初始化表达式带进 IR 的 global 初值。
func (l *Lowerer) collectGlobalInits() {
	for _, d := range l.mod.AllDecls() {
		ld, ok := d.(*ast.LetDecl)
		if !ok || ld.Init == nil {
			continue
		}
		if l.globalInits == nil {
			l.globalInits = map[string]ast.Node{}
		}
		l.globalInits[ld.Name] = ld.Init
	}
}

func (l *Lowerer) constLiteral(lt string, expr ast.Node) string {
	if expr == nil {
		return "zeroinitializer"
	}
	if e, ok := expr.(*ast.ArrayLit); ok {
		return l.arrayConstLiteral(lt, e)
	}
	// 字符串字面量的全局初值：Text 是 (ptr, len) 胖视图，字面量进常量池，初值填 { ptr @.str, len }。
	// 之前这里没有分支 → 掉到兜底的 zeroinitializer，模块级 `let S: Text = "…"` 会变成空串
	//（打印为空、view 出来 len=0、下标必 trap；designs/0006 §4.1 第 6 条）。
	if e, ok := expr.(*ast.StrLit); ok && lt == l.viewLT() {
		sym := l.stringConst(e.Value)
		return "{ ptr " + sym + ", " + l.lenLT() + " " + strconv.Itoa(len(e.Value)) + " }"
	}
	if v, ok := l.scalarConstLiteral(lt, expr); ok {
		return v
	}
	if strings.HasPrefix(lt, "i") || lt == "float" || lt == "double" {
		return "0"
	}
	return "zeroinitializer"
}

// arrayConstLiteral 把定长数组字面量编成 LLVM 聚合初值（`[i8 1, i8 2, …]`）。
//
// 之前这里只有 repeat 形态（`[0; N]` → zeroinitializer），显式元素列表会掉到
// 兜底的 zeroinitializer——字库、命令表这类数据表会被静默清零。现在逐元素编，
// 编不出来的元素走 unsupported 报告（宁可报错也不静默丢数据）。
func (l *Lowerer) arrayConstLiteral(lt string, e *ast.ArrayLit) string {
	if e.Repeat != nil {
		return "zeroinitializer"
	}
	elemLt := arrayElemLLType(lt)
	if elemLt == "" || len(e.Elems) == 0 {
		return "zeroinitializer"
	}
	parts := make([]string, 0, len(e.Elems))
	for _, el := range e.Elems {
		v, ok := l.scalarConstLiteral(elemLt, el)
		if !ok {
			l.unsupported = append(l.unsupported,
				"全局定长数组的初始值含非字面量元素（已退化成 0，请改用字面量或运行期赋值）")
			return "zeroinitializer"
		}
		parts = append(parts, elemLt+" "+v)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// arrayElemLLType 从 `[N x T]` 里取元素类型 T；不是定长数组返回空串。
func arrayElemLLType(lt string) string {
	if !strings.HasPrefix(lt, "[") || !strings.HasSuffix(lt, "]") {
		return ""
	}
	body := strings.TrimSuffix(strings.TrimPrefix(lt, "["), "]")
	i := strings.Index(body, " x ")
	if i < 0 {
		return ""
	}
	return strings.TrimSpace(body[i+3:])
}

// scalarConstLiteral 编标量常量；不可编返回 false。
func (l *Lowerer) scalarConstLiteral(lt string, expr ast.Node) (string, bool) {
	switch e := expr.(type) {
	case *ast.IntLit:
		return strconv.FormatInt(e.Raw, 10), true
	case *ast.BoolLit:
		return boolLit(e.Value), true
	case *ast.FloatLit:
		return floatLit(e.Value, llvmFloatBits(lt)), true
	case *ast.Unary:
		if e.Op == "-" {
			if il, ok := e.Operand.(*ast.IntLit); ok {
				return strconv.FormatInt(-il.Raw, 10), true
			}
		}
	}
	return "", false
}

// --------------------------------------------------------------------------- //
// 类型映射
// --------------------------------------------------------------------------- //

// llty 把 AILA 类型映射成 LLVM 类型文本。
func (l *Lowerer) llty(t types.Type) string {
	switch x := t.(type) {
	case nil:
		return "i32"
	case types.Prim:
		if x.K == "text" {
			return l.viewLT()
		}
		if x.K == "usize" || x.K == "isize" {
			return "i" + strconv.Itoa(l.ptrSize*8)
		}
		if s, ok := primLL[x.K]; ok {
			return s
		}
		return "i32"
	case types.View:
		return l.viewLT()
	case types.UnsafeView:
		return l.viewLT()
	case types.Fixed:
		return "[" + strconv.FormatInt(x.N, 10) + " x " + l.llty(x.Elem) + "]"
	case types.Optional:
		if l.isNicheOptional(x) {
			return l.llty(x.Inner)
		}
		return "{ i8, " + l.llty(x.Inner) + " }"
	case types.Result:
		return "{ i1, " + l.llty(x.Ok) + ", " + l.llty(x.Err) + " }"
	case types.CoroState:
		// resume 匿名判别和：{ i8 state, OkT, ErrT }；state 0=Suspended/1=Ok|Done/2=Err，
		// 与协程控制块状态码一致（§10.5）。二态（Ok==nil）退化成 { i8 }。
		if x.Ok == nil {
			return "{ i8 }"
		}
		s := "{ i8, " + l.llty(x.Ok)
		if x.Err != nil {
			s += ", " + l.llty(x.Err)
		}
		return s + " }"
	case types.Chan, types.Coro, types.FnPtr:
		return "ptr"
	case types.Group:
		parts := make([]string, 0, len(x.Items))
		for _, it := range x.Items {
			parts = append(parts, l.llty(it))
		}
		return "{ " + strings.Join(parts, ", ") + " }"
	case types.Named:
		switch info := x.Info.(type) {
		case *types.StructInfo:
			parts := make([]string, 0, len(info.Fields))
			for _, f := range info.Fields {
				parts = append(parts, l.llty(f.Type))
			}
			body := "{ }"
			if len(parts) > 0 {
				body = "{ " + strings.Join(parts, ", ") + " }"
			}
			if info.Packed {
				return "<" + body + ">"
			}
			return body
		case *types.EnumInfo:
			return l.enumLL(info)
		}
		return "i64"
	case types.ErrorSpace:
		// 保留错误空间超集（designs/0004 §3.1）：布局 = 全局聚合错误集 enum
		return l.enumLL(x.Info)
	case types.Void, types.Never:
		return "void"
	}
	return "i64"
}

// enumLL 生成 enum 的 LLVM 聚合布局（tag + 载荷联合）。
func (l *Lowerer) enumLL(info *types.EnumInfo) string {
	tag := l.llty(info.TagType)
	if !info.Payload {
		return tag
	}
	// 载荷联合取字节宽度最大的变体（Text 这类 FatPtr 单字段 16B，
	// 不能按字段数比——否则宽载荷会被塞进窄槽，store 越界踩栈）
	best, bestW := "{ }", -1
	for _, v := range info.Variants {
		if len(v.Payload) == 0 {
			continue
		}
		parts := make([]string, 0, len(v.Payload))
		w := 0
		for _, f := range v.Payload {
			parts = append(parts, l.llty(f.Type))
			w += l.layout.Size(f.Type)
		}
		if w > bestW {
			best, bestW = "{ "+strings.Join(parts, ", ")+" }", w
		}
	}
	return "{ " + tag + ", " + best + " }"
}

var primLL = map[string]string{
	"i8": "i8", "i16": "i16", "i32": "i32", "i64": "i64",
	"u8": "i8", "u16": "i16", "u32": "i32", "u64": "i64",
	"f32": "float", "f64": "double", "bool": "i1", "c32": "i32",
}

// isNicheOptional 判定 `T?` 是否用空指针做 niche 表示。
func (l *Lowerer) isNicheOptional(t types.Optional) bool {
	switch t.Inner.(type) {
	case types.FnPtr, types.Chan, types.Coro:
		return true
	}
	return types.IsViewFamily(t.Inner)
}

// isAgg 判定是否为聚合值（需要走内存槽）。
func (l *Lowerer) isAgg(t types.Type) bool {
	switch l.llty(t) {
	case "i1", "i8", "i16", "i32", "i64", "float", "double", "ptr", "void":
		return false
	}
	return true
}

// zero 生成类型的零值字面量。
func (l *Lowerer) zero(t types.Type) string {
	lt := l.llty(t)
	switch {
	case lt == "void":
		return ""
	case lt == "ptr":
		return "null"
	case lt == "float" || lt == "double":
		return "0.0"
	case strings.HasPrefix(lt, "i"):
		return "0"
	}
	return "zeroinitializer"
}

// mutsSelf 判断方法的 self 是否带 @mut（就地写语义）。
func mutsSelf(sig *sema.FuncSig) bool {
	if len(sig.Params) == 0 || sig.Params[0].Name != "self" {
		return false
	}
	for _, m := range sig.Muts {
		if m == "self" {
			return true
		}
	}
	return false
}

// paramLLType 是 C 边界上的参数类型：视图家族只传 ptr，len 丢弃（designs/0001 §11.3）。
// 注意 @mut[self] 的 self 不走这里——它按参数下标在 funcLLType / genFunction /
// 调用点单独判（mutsSelf），避免误伤其余参数。
func (l *Lowerer) paramLLType(sig *sema.FuncSig, t types.Type) string {
	if sig.IsExtern {
		if p, ok := t.(types.Prim); ok && p.K == "bool" {
			return "i8"
		}
		if types.IsViewFamily(t) {
			return "ptr"
		}
		if o, ok := t.(types.Optional); ok && types.IsViewFamily(o.Inner) {
			return "ptr"
		}
		if n, ok := t.(types.Named); ok {
			if info, isEnum := n.Info.(*types.EnumInfo); isEnum {
				if !info.Payload {
					return l.llty(info.Base)
				}
				return l.llty(info.TagType)
			}
		}
	}
	return l.llty(t)
}

// sigRetType 是签名的 IR 返回类型：有错误通道时整体是 `T ! E` 的聚合。
func (l *Lowerer) sigRetType(sig *sema.FuncSig) types.Type {
	var base types.Type
	if sig.RetGroup != nil {
		base = types.Group{Items: sig.RetGroup}
	} else {
		base = sig.Ret
	}
	if sig.Err != nil {
		return types.Result{Ok: base, Err: sig.Err}
	}
	return base
}

// paramByRef 判断某参数是否按「只读指针」传。（AX-9）
//
// 背景：按值传大聚合体在 16 寄存器目标（xtensa call0）上会把 backend 直接顶崩——
// esp-llvm 的 RA 报 "Cannot scavenge register without an emergency spill slot"，
// 且 -O0..-O3 都一样（designs/0006 §4.1 第 8 条实测：≤32 字节的聚合体没事，
// 6 字以上的 `{ ... }` 在循环里按值传就崩）。C ABI 对大结构体也是按指针传，
// 这里与之对齐：**非 extern 函数**的大聚合体参数一律退化成只读指针，
// 调用点传主名字地址（rvalue 则先落一个临时槽）。
//
// 只读性由 sema 保证（非 @mut 参数不可写），所以被调方拿到 ptr 后只做 load 是安全的。
// extern（`#compile.c`）函数保持 C ABI 原样，不做这条转换。
func (l *Lowerer) paramByRef(sig *sema.FuncSig, i int, t types.Type) bool {
	if sig.IsExtern {
		return false
	}
	if i == 0 && mutsSelf(sig) {
		return false // @mut[self] 已按 ptr 传，走自己那条路
	}
	if types.IsViewFamily(t) {
		return false // 视图按值传 { ptr, len }：写已经经 ptr 落地，不必再套一层
	}
	lt := l.llty(t)
	if !strings.HasPrefix(lt, "{") && !strings.HasPrefix(lt, "[") {
		return false // 标量 / 函数指针照旧按值
	}
	if l.layout.Size(t) > 16 {
		return true // 大聚合体：16 寄存器目标上按值传会把后端顶崩
	}
	// 小聚合体的 @mut 参数也必须按指针传：否则被调方拿到的是副本，
	// 「@mut[self] 方法经 @mut 参数转发」写不回调用方（designs/0006 §4.1 第 5 条）。
	for _, m := range sig.Muts {
		if m == sig.Params[i].Name {
			return true
		}
	}
	return false
}

// funcLLType 返回（返回类型, 参数类型列表）；区域参数一律 ptr。
// extern 无名 `...`（C 变参）在参数列表尾部补 `...`——declare 与调用点函数类型注解共用。
func (l *Lowerer) funcLLType(sig *sema.FuncSig) (string, []string) {
	params := make([]string, 0, len(sig.Regions)+len(sig.Params)+1)
	for range sig.Regions {
		params = append(params, "ptr")
	}
	for i, p := range sig.Params {
		if i == 0 && mutsSelf(sig) {
			params = append(params, "ptr")
			continue
		}
		if l.paramByRef(sig, i, p.Type) {
			params = append(params, "ptr")
			continue
		}
		params = append(params, l.paramLLType(sig, p.Type))
	}
	if sig.Variadic {
		params = append(params, "...")
	}
	return l.llty(l.sigRetType(sig)), params
}

// --------------------------------------------------------------------------- //
// 小工具
// --------------------------------------------------------------------------- //

// enumOf 取 Named 类型背后的 enum 信息。
func enumOf(t types.Type) *types.EnumInfo {
	if es, ok := t.(types.ErrorSpace); ok {
		// v2：保留错误空间超集按聚合错误集 enum 分派（designs/0004 §4.1）
		return es.Info
	}
	n, ok := t.(types.Named)
	if !ok {
		return nil
	}
	info, ok := n.Info.(*types.EnumInfo)
	if !ok {
		return nil
	}
	return info
}

// emptyBag 给类型解析提供一个丢弃用途的诊断袋（sema 已校验过）。
func emptyBag() *diag.Bag { return &diag.Bag{} }

// diagSpan 是陷阱位置类型（与 diag.Span 同一类型）。
type diagSpan = diag.Span

func isTextType(t types.Type) bool {
	p, ok := t.(types.Prim)
	return ok && p.K == "text"
}

func isBoolType(t types.Type) bool {
	p, ok := t.(types.Prim)
	return ok && p.K == "bool"
}

func llvmFloatBits(lt string) int {
	if lt == "float" {
		return 32
	}
	return 64
}

func boolLit(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func toFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int64:
		return float64(x)
	case int:
		return float64(x)
	}
	return 0
}

func toBool(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case int64:
		return x != 0
	}
	return false
}

func asInt(v any) (int64, bool) {
	switch x := v.(type) {
	case int64:
		return x, true
	case int:
		return int64(x), true
	case int32:
		return int64(x), true
	case float64:
		return int64(x), true
	}
	return 0, false
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// floatLit 生成 LLVM 浮点字面量（十进制科学计数法，跨版本稳定）。
func floatLit(v float64, bits int) string {
	if bits == 32 {
		v = float64(float32(v))
		return ensureFloatPoint(strconv.FormatFloat(v, 'g', 9, 32))
	}
	return ensureFloatPoint(strconv.FormatFloat(v, 'g', 17, 64))
}

func ensureFloatPoint(s string) string {
	if !strings.ContainsAny(s, "eE.") {
		return s + ".0"
	}
	return s
}
