// goldengen 是**黄金语料生成器**（rfcs/0003 §4.1）。
//
// 把 `tests/corpus` 与 `examples` 的契约快照全部由 **Go 实现**重新生成：
//
//	tests/corpus/<名字>.tokens.json    token 序列 [[kind, text, value], …]
//	tests/corpus/<名字>.lexdiags.json  词法层诊断码
//	tests/corpus/<名字>.ast.txt        AST 的 S-表达式
//	tests/corpus/<名字>.diags.json     有错误时的诊断码（此时不出 AST 快照）
//	tests/corpus/layout.json           类型布局
//	tests/corpus/sigs.json             函数签名与模块级值
//	tests/corpus/bindings.json         let/var 绑定推断类型
//	tests/corpus/<名字>.ir             单模块 IR（Go 已实现子集）
//	tests/corpus/bindings/<cpack>.ail  C 绑定
//	tests/corpus/examples/<名字>/*.ll  示例工程每模块 IR
//
// 用法：
//
//	go run ./tools/goldengen            # 重新生成
//	go run ./tools/goldengen -check     # 只校验（CI 用）
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ailang/aila/internal/ast"
	"github.com/ailang/aila/internal/bindgen"
	"github.com/ailang/aila/internal/build"
	"github.com/ailang/aila/internal/codegen"
	"github.com/ailang/aila/internal/diag"
	"github.com/ailang/aila/internal/lexer"
	"github.com/ailang/aila/internal/parser"
	"github.com/ailang/aila/internal/sema"
	"github.com/ailang/aila/internal/types"
)

const hostTriple = "x86_64-w64-windows-gnu"

// irCorpus 是「Go codegen 已实现子集」的语料（与 internal/codegen/ir_test.go 保持一致）。
var irCorpus = []string{"ir_basic", "ir_wide"}

// exampleNames 是全部示例工程。
var exampleNames = []string{"hello", "control", "region", "struct", "result", "cinterop"}

type generator struct {
	root    string
	check   bool
	stale   []string
	written []string
}

func main() {
	check := flag.Bool("check", false, "只校验快照是否与当前实现一致")
	flag.Parse()
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "取不到工作目录：", err)
		os.Exit(2)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		fmt.Fprintln(os.Stderr, "请在仓库根目录运行（找不到 go.mod）")
		os.Exit(2)
	}
	g := &generator{root: root, check: *check}
	g.corpus()
	g.layout()
	g.sigs()
	g.bindingsTypes()
	g.ir()
	g.bindingAil()
	g.exampleIR()
	if g.check {
		if len(g.stale) > 0 {
			fmt.Fprintln(os.Stderr, "快照与当前实现不一致：")
			for _, name := range g.stale {
				fmt.Fprintln(os.Stderr, "  -", name)
			}
			fmt.Fprintln(os.Stderr, "重新生成：go run ./tools/goldengen")
			os.Exit(1)
		}
		fmt.Println("全部快照与当前实现一致")
		return
	}
	fmt.Printf("已生成 %d 份快照\n", len(g.written))
}

// emit 写文件（或在校验模式下比对）。
func (g *generator) emit(rel string, content string) {
	path := filepath.Join(g.root, rel)
	if g.check {
		old, err := os.ReadFile(path)
		if err != nil || string(old) != content {
			g.stale = append(g.stale, rel)
		}
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		fmt.Fprintln(os.Stderr, "建不了目录：", err)
		os.Exit(2)
	}
	if err := os.WriteFile(path, []byte(content), 0666); err != nil {
		fmt.Fprintln(os.Stderr, "写不了文件：", err)
		os.Exit(2)
	}
	g.written = append(g.written, rel)
	if !g.check {
		fmt.Println("  已写出", rel)
	}
}

// remove 在校验模式下删除多余文件；生成模式下也删（与参考生成器行为一致）。
func (g *generator) remove(rel string) {
	path := filepath.Join(g.root, rel)
	if g.check {
		if _, err := os.Stat(path); err == nil {
			g.stale = append(g.stale, rel+"（应删除）")
		}
		return
	}
	_ = os.Remove(path)
}

// --------------------------------------------------------------------------- //
// 各阶段
// --------------------------------------------------------------------------- //

// corpus 生成 token / lexdiags / AST 或诊断码快照。
func (g *generator) corpus() {
	dir := filepath.Join(g.root, "tests", "corpus")
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "读不到语料目录：", err)
		os.Exit(2)
	}
	var files []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".aila") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "语料目录为空")
		os.Exit(2)
	}
	for _, name := range files {
		rel := filepath.Join("tests", "corpus", name)
		raw, err := os.ReadFile(filepath.Join(g.root, rel))
		if err != nil {
			continue
		}
		src := string(raw)
		toks, lexBag := lexer.Lex(name, src)
		lexCodes := map[string]bool{}
		for _, d := range lexBag.Items {
			if strings.HasPrefix(d.Code, "E10") {
				lexCodes[d.Code] = true
			}
		}
		mod, bag := parser.Parse(name, src)
		base := strings.TrimSuffix(rel, ".aila")

		// tokens：[[kind, text, value], …]（EOF 不进快照）
		rows := make([]any, 0, len(toks))
		for _, t := range toks {
			rows = append(rows, []any{string(t.Kind), t.Text, jsonValue(t.Value)})
		}
		g.emit(base+".tokens.json", pyJSON(rows, 0))

		codes := sortedKeys(lexCodes)
		g.emit(base+".lexdiags.json", pyCompact(codes))

		allErr := map[string]bool{}
		for _, d := range bag.Items {
			if d.Code != "" {
				allErr[d.Code] = true
			}
		}
		if bag.HasErrors() {
			g.emit(base+".diags.json", pyCompact(sortedKeys(allErr)))
			g.remove(base + ".ast.txt")
		} else {
			g.emit(base+".ast.txt", ast.Sexp(mod)+"\n")
			g.remove(base + ".diags.json")
		}
	}
}

// check 装载单模块并跑语义检查（layout / sigs / bindings / ir 共用）。
func (g *generator) check1(name string, wantNoErrors bool) (*sema.Program, *sema.Table, *sema.Sigs, *sema.Checker, *types.Layout) {
	rel := filepath.Join("tests", "corpus", name)
	raw, err := os.ReadFile(filepath.Join(g.root, rel))
	if err != nil {
		fmt.Fprintln(os.Stderr, "读不到语料：", err)
		os.Exit(2)
	}
	bag := &diag.Bag{}
	mod, pbag := parser.Parse(name, string(raw))
	bag.Items = append(bag.Items, pbag.Items...)
	mod.Name = "_.main"
	mods := map[string]*ast.Module{"_.main": mod}
	prog := sema.BuildProgram(mods, []string{"_.main"}, "_.main", bag)
	if wantNoErrors && bag.HasErrors() {
		fmt.Fprintf(os.Stderr, "%s 语义检查失败：\n%s", name, bag.Render(false))
		os.Exit(2)
	}
	u := prog.Unit("_.main")
	layout := types.NewLayout(8)
	return prog, u.Table, u.Sigs, u.Check, layout
}

// layout 生成类型布局快照。
func (g *generator) layout() {
	_, table, _, check, layout := g.check1("layout.aila", true)
	_ = check
	out := map[string]any{"types": map[string]any{}, "aliases": map[string]any{}}
	typesMap := map[string]any{}
	for name, info := range table.Types {
		t := types.Named{Info: info}
		entry := map[string]any{
			"type":  types.TypeName(t),
			"size":  layout.Size(t),
			"align": layout.Align(t),
		}
		if si, ok := info.(*types.StructInfo); ok {
			entry["kind"] = "struct"
			entry["packed"] = si.Packed
			entry["align_attr"] = si.Align
			offsets := layout.FieldOffsets(si)
			fields := []any{}
			for i, f := range si.Fields {
				off := 0
				if i < len(offsets) {
					off = offsets[i]
				}
				fields = append(fields, map[string]any{
					"name":   f.Name,
					"offset": off,
					"size":   layout.Size(f.Type),
					"align":  layout.Align(f.Type),
				})
			}
			entry["fields"] = fields
		} else if ei, ok := info.(*types.EnumInfo); ok {
			entry["kind"] = "enum"
			entry["payload"] = ei.Payload
			entry["tag"] = types.TypeName(ei.TagType)
			entry["tag_size"] = layout.Size(ei.TagType)
			variants := []any{}
			for _, v := range ei.Variants {
				variants = append(variants, map[string]any{"name": v.Name, "tag": v.Tag})
			}
			entry["variants"] = variants
		}
		typesMap[name] = entry
	}
	out["types"] = typesMap
	aliases := map[string]any{}
	for name, t := range table.Aliases {
		if t == nil {
			continue
		}
		aliases[name] = map[string]any{
			"type":  types.TypeName(t),
			"size":  layout.Size(t),
			"align": layout.Align(t),
		}
	}
	out["aliases"] = aliases
	g.emit(filepath.Join("tests", "corpus", "layout.json"), pyJSON(out, 2))
}

// sigs 生成签名快照。
func (g *generator) sigs() {
	_, _, sigs, _, _ := g.check1("sigs.aila", true)
	funcs := map[string]any{}
	for name, sig := range sigs.Funcs {
		funcs[name] = sigEntryOf(sig)
	}
	methods := map[string]any{}
	for key, sig := range sigs.Methods {
		methods[key] = sigEntryOf(sig)
	}
	consts := map[string]any{}
	for name, ci := range sigs.Consts {
		consts[name] = map[string]any{
			"type":  types.TypeName(ci.Type),
			"value": ci.Value,
		}
	}
	globals := map[string]any{}
	for name, gi := range sigs.Globals {
		globals[name] = map[string]any{
			"type":    types.TypeName(gi.Type),
			"kind":    gi.Kind,
			"mutable": gi.Mutable,
		}
	}
	out := map[string]any{"funcs": funcs, "methods": methods, "consts": consts, "globals": globals}
	g.emit(filepath.Join("tests", "corpus", "sigs.json"), pyJSON(out, 2))
}

func sigEntryOf(sig *sema.FuncSig) map[string]any {
	params := []any{}
	for _, p := range sig.Params {
		params = append(params, []any{p.Name, types.TypeName(p.Type)})
	}
	entry := map[string]any{
		"owner":     sig.Owner,
		"params":    params,
		"ret":       types.TypeName(sig.Ret),
		"regions":   orEmpty(sig.Regions),
		"muts":      orEmpty(sig.Muts),
		"moves":     orEmpty(sig.Moves),
		"coroutine": sig.IsCoroutine,
		"thread":    sig.IsThread,
		"extern":    sig.IsExtern,
		"public":    sig.IsPublic,
		"variadic":  sig.Variadic,
	}
	if sig.Err != nil {
		entry["err"] = types.TypeName(sig.Err)
	} else {
		entry["err"] = nil
	}
	if sig.RetGroup != nil {
		group := []any{}
		for _, t := range sig.RetGroup {
			group = append(group, types.TypeName(t))
		}
		entry["group"] = group
	} else {
		entry["group"] = nil
	}
	return entry
}

// bindingsTypes 生成 let/var 绑定推断类型快照。
func (g *generator) bindingsTypes() {
	dir := filepath.Join(g.root, "tests", "corpus")
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "读不到语料目录：", err)
		os.Exit(2)
	}
	filesMap := map[string]any{}
	skipped := map[string]any{}
	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".aila") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		bag := &diag.Bag{}
		mod, pbag := parser.Parse(name, string(raw))
		bag.Items = append(bag.Items, pbag.Items...)
		mod.Name = "_.main"
		prog := sema.BuildProgram(map[string]*ast.Module{"_.main": mod}, []string{"_.main"}, "_.main", bag)
		codes := map[string]bool{}
		for _, d := range bag.Items {
			if d.Code != "" && d.Severity == diag.SeverityError {
				codes[d.Code] = true
			}
		}
		if bag.HasErrors() {
			skipped[name] = sortedKeys(codes)
			continue
		}
		u := prog.Unit("_.main")
		perFunc := map[string]any{}
		for _, d := range mod.AllDecls() {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue // 无函数体（extern / 声明）没有绑定
			}
			key := fn.Name
			if fn.OwnerType != "" {
				key = fn.OwnerType + "." + fn.Name
			}
			var out [][]string
			collectBindings(fn.Body, u.Check, &out)
			rows := make([]any, 0, len(out))
			for _, pair := range out {
				rows = append(rows, []any{pair[0], pair[1]})
			}
			perFunc[key] = rows
		}
		filesMap[name] = perFunc
	}
	out := map[string]any{"files": filesMap, "skipped": skipped}
	g.emit(filepath.Join("tests", "corpus", "bindings.json"), pyJSON(out, 2))
}

// collectBindings 按源码顺序收集函数体里的 let/var 绑定类型（含嵌套块）。
func collectBindings(b *ast.Block, c *sema.Checker, out *[][]string) {
	if b == nil {
		return
	}
	for _, st := range b.Stmts {
		switch x := st.(type) {
		case *ast.LetDecl:
			name := "?"
			if t := c.Bindings[x]; t != nil {
				name = types.TypeName(t)
			}
			*out = append(*out, []string{x.Name, name})
		case *ast.Block:
			collectBindings(x, c, out)
		case *ast.If:
			collectBindings(x.Then, c, out)
			for _, pair := range x.Elifs {
				if blk, ok := pair.B.(*ast.Block); ok {
					collectBindings(blk, c, out)
				}
			}
			collectBindings(x.Els, c, out)
		case *ast.ForC:
			collectBindings(x.Body, c, out)
		case *ast.ForIn:
			collectBindings(x.Body, c, out)
		case *ast.Region:
			collectBindings(x.Body, c, out)
		case *ast.Unsafe:
			collectBindings(x.Body, c, out)
		case *ast.Defer:
			collectBindings(x.Body, c, out)
		case *ast.Match:
			for _, arm := range x.Arms {
				if blk, ok := arm.Body.(*ast.Block); ok {
					collectBindings(blk, c, out)
				}
			}
		}
	}
}

// ir 生成单模块 IR 快照（Go 已实现子集）。
func (g *generator) ir() {
	for _, name := range irCorpus {
		prog, table, sigs, check, layout := g.check1(name+".aila", true)
		u := prog.Unit("_.main")
		_ = u
		lw := codegen.New(prog.Unit("_.main").Mod, "_.main", table, sigs, check, layout, hostTriple, "_.main")
		ir := lw.Generate()
		if bad := lw.Unsupported(); len(bad) > 0 {
			fmt.Fprintf(os.Stderr, "%s 有未实现构造：%v\n", name, bad)
			os.Exit(2)
		}
		g.emit(filepath.Join("tests", "corpus", name+".ir"), ir)
	}
}

// bindingAil 生成 cpack 的 `.ail` 绑定快照。
func (g *generator) bindingAil() {
	cpackDir := filepath.Join(g.root, "examples", "cinterop", "cpacks", "mathkit")
	target := bindgen.Target{Board: "host", Triple: hostTriple, PtrSize: 8}
	bag := &diag.Bag{}
	res := bindgen.Run("mathkit", cpackDir, target, bag, "builtin")
	if bag.HasErrors() {
		fmt.Fprintf(os.Stderr, "绑定生成失败：\n%s", bag.Render(false))
		os.Exit(2)
	}
	text := res.Header("mathkit", target) + "\n\n" + res.Text
	g.emit(filepath.Join("tests", "corpus", "bindings", "mathkit.ail"), text)
}

// exampleIR 生成示例工程每个模块的 IR 快照。
func (g *generator) exampleIR() {
	for _, name := range exampleNames {
		res := build.Load(filepath.Join(g.root, "examples", name), build.Options{})
		if res.Diags.HasErrors() {
			fmt.Fprintf(os.Stderr, "%s 装载失败：\n%s", name, res.Diags.Render(false))
			os.Exit(2)
		}
		irs := res.GenerateIR()
		if res.Diags.HasErrors() {
			fmt.Fprintf(os.Stderr, "%s 代码生成失败：\n%s", name, res.Diags.Render(false))
			os.Exit(2)
		}
		modNames := make([]string, 0, len(irs))
		for mod := range irs {
			modNames = append(modNames, mod)
		}
		sort.Strings(modNames)
		for _, mod := range modNames {
			safe := strings.NewReplacer("_.", "", ".", "_", "/", "_").Replace(mod)
			g.emit(filepath.Join("tests", "corpus", "examples", name, safe+".ll"), irs[mod])
		}
	}
}

// --------------------------------------------------------------------------- //
// JSON 与杂项
// --------------------------------------------------------------------------- //

// pyJSON 复刻 Python `json.dumps(v, ensure_ascii=False, indent=n)` 的形态：
//   n>0：每层换行 + n 空格缩进，键值分隔 `": "`、元素分隔 `","`；
//   n=0：每层换行但不缩进（参考生成器对 tokens 快照用的就是这个形态）。
func pyJSON(v any, indent int) string {
	var b strings.Builder
	writePyJSON(&b, v, indent, 0)
	b.WriteString("\n")
	return b.String()
}

func writePyJSON(b *strings.Builder, v any, indent, level int) {
	pad := func(n int) string { return strings.Repeat(" ", n) }
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		if x {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case string:
		b.WriteString(jsonScalar(x))
	case float64:
		b.WriteString(formatNumber(x))
	case float32:
		b.WriteString(formatNumber(float64(x)))
	case int, int64, int32:
		b.WriteString(fmt.Sprint(x))
	case []string:
		items := make([]any, 0, len(x))
		for _, s := range x {
			items = append(items, s)
		}
		writePyJSON(b, items, indent, level)
	case []any:
		if len(x) == 0 {
			b.WriteString("[]")
			return
		}
		inner := level + 1
		b.WriteString("[\n")
		for i, item := range x {
			b.WriteString(pad(indent * inner))
			writePyJSON(b, item, indent, inner)
			if i != len(x)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(pad(indent * level))
		b.WriteString("]")
	case map[string]any:
		if len(x) == 0 {
			b.WriteString("{}")
			return
		}
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		inner := level + 1
		b.WriteString("{\n")
		for i, k := range keys {
			b.WriteString(pad(indent * inner))
			b.WriteString(jsonScalar(k))
			b.WriteString(": ")
			writePyJSON(b, x[k], indent, inner)
			if i != len(keys)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(pad(indent * level))
		b.WriteString("}")
	default:
		b.WriteString(jsonScalar(fmt.Sprint(x)))
	}
}

// pyCompact 复刻 Python `json.dumps(v, ensure_ascii=False)`（默认分隔符 `", "` / `": "`）。
func pyCompact(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintln(os.Stderr, "JSON 编码失败：", err)
		os.Exit(2)
	}
	// Go 默认紧凑分隔符是无空格的；Python 默认带空格
	s := string(raw)
	return strings.ReplaceAll(s, ",", ", ") + "\n"
}

func jsonScalar(s string) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return `""`
	}
	return strings.TrimRight(buf.String(), "\n")
}

// formatNumber 复刻 Python 的浮点渲染（整数样子的浮点不加 `.0`，指数形式写 `e-05`）。
func formatNumber(f float64) string {
	s := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.ContainsAny(s, ".eE") {
		return s + ".0"
	}
	if strings.Contains(s, "e") {
		// Python 用 `e-05`（两位指数）
		parts := strings.SplitN(s, "e", 2)
		mant, exp := parts[0], parts[1]
		if !strings.Contains(mant, ".") {
			mant += ".0"
		}
		sign := "+"
		digits := exp
		if strings.HasPrefix(exp, "-") {
			sign = "-"
			digits = exp[1:]
		} else if strings.HasPrefix(exp, "+") {
			digits = exp[1:]
		}
		for len(digits) < 2 {
			digits = "0" + digits
		}
		return mant + "e" + sign + digits
	}
	return s
}
func jsonValue(v any) any {
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case float32:
		return x
	}
	return v
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func orEmpty(list []string) []any {
	out := make([]any, 0, len(list))
	for _, s := range list {
		out = append(out, s)
	}
	return out
}
