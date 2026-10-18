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

	"github.com/ailang/aila/internal/ast"
	"github.com/ailang/aila/internal/diag"
	"github.com/ailang/aila/internal/sema"
	"github.com/ailang/aila/internal/types"
)

// Lowerer 生成单个模块的 IR。
type Lowerer struct {
	mod     *ast.Module
	modName string
	srcPath string
	table   *sema.Table
	sigs    *sema.Sigs
	check   *sema.Checker
	layout  *types.Layout
	triple  string
	entry   string
	ptrSize int

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
	funcs          []string

	// unsupported 记录本批次尚未实现的构造（Go 侧不静默丢代码）。
	unsupported []string
}

type constVal struct {
	lt  string
	lit string
}

// New 构造 Lowerer（entry = 入口模块名，如 "_.main"）。
//
// 陷阱信息里的源文件路径不取调用方传进来的路径，而是由**模块名派生**
// （`_.main` → `main.aila`，`std.io` → `std/io.aila`）：产物跨机器一致（缺陷 13）。
func New(mod *ast.Module, modName string, table *sema.Table, sigs *sema.Sigs,
	check *sema.Checker, layout *types.Layout, triple, entry string) *Lowerer {
	return &Lowerer{
		mod: mod, modName: modName, srcPath: DisplayPath(modName), table: table, sigs: sigs, check: check,
		layout: layout, triple: triple, entry: entry, ptrSize: 8,
		consts:      map[string]constVal{},
		strSyms:     map[string]string{},
		declares:    map[string]string{},
		defined:     map[string]bool{},
		globalsSeen: map[string]bool{},
	}
}

// Generate 产出整份 IR 文本。
func (l *Lowerer) Generate() string {
	l.collectConsts()
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
	return strings.TrimRight(out, " \t\r\n") + "\n"
}

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
			l.consts[name] = constVal{"{ ptr, i64 }",
				"{ ptr " + sym + ", i64 " + strconv.Itoa(len(text)) + " }"}
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
// 跨模块调用要用**声明模块**的名字做修饰（`io.print_text` → `aila_std_io__print_text`）。
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

// ensureGlobal 登记模块级全局变量（每个符号只发一次）。
func (l *Lowerer) ensureGlobal(sym, lt string, init ast.Node) {
	if l.globalsSeen[sym] {
		return
	}
	l.globalsSeen[sym] = true
	l.pendingGlobals = append(l.pendingGlobals,
		sym+" = internal global "+lt+" "+l.constLiteral(lt, init)+", align 8")
}

func (l *Lowerer) constLiteral(lt string, expr ast.Node) string {
	if expr == nil {
		return "zeroinitializer"
	}
	switch e := expr.(type) {
	case *ast.IntLit:
		return strconv.FormatInt(e.Raw, 10)
	case *ast.BoolLit:
		return boolLit(e.Value)
	case *ast.FloatLit:
		return floatLit(e.Value, llvmFloatBits(lt))
	case *ast.Unary:
		if e.Op == "-" {
			if il, ok := e.Operand.(*ast.IntLit); ok {
				return strconv.FormatInt(-il.Raw, 10)
			}
		}
	case *ast.ArrayLit:
		if e.Repeat != nil {
			return "zeroinitializer"
		}
	}
	if strings.HasPrefix(lt, "i") || lt == "float" || lt == "double" {
		return "0"
	}
	return "zeroinitializer"
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
			return "{ ptr, i64 }"
		}
		if x.K == "usize" || x.K == "isize" {
			return "i" + strconv.Itoa(l.ptrSize*8)
		}
		if s, ok := primLL[x.K]; ok {
			return s
		}
		return "i32"
	case types.View:
		return "{ ptr, i64 }"
	case types.UnsafeView:
		return "{ ptr, i64 }"
	case types.Fixed:
		return "[" + strconv.FormatInt(x.N, 10) + " x " + l.llty(x.Elem) + "]"
	case types.Optional:
		if l.isNicheOptional(x) {
			return l.llty(x.Inner)
		}
		return "{ i8, " + l.llty(x.Inner) + " }"
	case types.Result:
		return "{ i1, " + l.llty(x.Ok) + ", " + l.llty(x.Err) + " }"
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
			tag := l.llty(info.TagType)
			if !info.Payload {
				return tag
			}
			best, bestN := "{ }", -1
			for _, v := range info.Variants {
				if len(v.Payload) == 0 {
					continue
				}
				parts := make([]string, 0, len(v.Payload))
				for _, f := range v.Payload {
					parts = append(parts, l.llty(f.Type))
				}
				if len(v.Payload) > bestN {
					best, bestN = "{ "+strings.Join(parts, ", ")+" }", len(v.Payload)
				}
			}
			return "{ " + tag + ", " + best + " }"
		}
		return "i64"
	case types.Void, types.Never:
		return "void"
	}
	return "i64"
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

// paramLLType 是 C 边界上的参数类型：视图家族只传 ptr，len 丢弃（designs/0001 §11.3）。
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

// funcLLType 返回（返回类型, 参数类型列表）；区域参数一律 ptr。
func (l *Lowerer) funcLLType(sig *sema.FuncSig) (string, []string) {
	params := make([]string, 0, len(sig.Regions)+len(sig.Params))
	for range sig.Regions {
		params = append(params, "ptr")
	}
	for _, p := range sig.Params {
		params = append(params, l.paramLLType(sig, p.Type))
	}
	return l.llty(l.sigRetType(sig)), params
}

// --------------------------------------------------------------------------- //
// 小工具
// --------------------------------------------------------------------------- //

// enumOf 取 Named 类型背后的 enum 信息。
func enumOf(t types.Type) *types.EnumInfo {
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
