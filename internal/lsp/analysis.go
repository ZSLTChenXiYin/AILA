package lsp

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"unicode"

	"aila/internal/ast"
	"aila/internal/diag"
	"aila/internal/manifest"
	"aila/internal/parser"
	"aila/internal/project"
	"aila/internal/sema"
	"aila/internal/types"
)

// symbol 是索引里的一条符号。
type symbol struct {
	Name      string
	Kind      int
	Range     Range
	Module    string
	Container string // 所属函数 / 类型（局部符号与成员才有）
	Detail    string
	Local     bool
	// Public 表示该符号能跨包可见（壳内 + public，designs/0001 §2.2）。
	Public bool
}

// moduleInfo 是一个已分析的模块。
type moduleInfo struct {
	Name    string
	Path    string
	URI     string
	Text    string
	Mod     *ast.Module
	Sigs    *sema.Sigs
	Check   *sema.Checker
	Symbols []*symbol
	Imports map[string]string // 绑定名 → 模块名
	Diags   []Diagnostic
	// Generated 表示这是 bindgen 生成的 C 绑定（`.ail`），不是人手写的源码：
	// 改名 / 格式化这类会写文件的动作一律拒绝。
	Generated bool
	Cpack     string // 绑定来自哪个 cpack
	Key       string // 绑定的缓存键（指纹变了就得重算符号表）
}

// analysis 是一次工程分析的产物。
type analysis struct {
	Modules map[string]*moduleInfo // 模块名 → 信息
	ByPath  map[string]*moduleInfo // 文件路径 → 信息
	Order   []string
}

// Workspace 持有打开的文档与最近一次分析结果。
type Workspace struct {
	root    string
	buffers map[string]string // uri → 文本
	last    map[string]*analysis
	// cbind 是 C 导入解析结果的进程内缓存（键：工程根 + cpack，值带指纹）。
	cbind map[string]*cBinding
}

// NewWorkspace 构造工作区。
func NewWorkspace() *Workspace {
	return &Workspace{buffers: map[string]string{}, last: map[string]*analysis{}}
}

// SetRoot 设置工程根（initialize 给的 rootUri）。
func (w *Workspace) SetRoot(path string) { w.root = path }

// Close 释放（目前无资源）。
func (w *Workspace) Close() {}

// Open 记录一份打开/修改过的文档。
func (w *Workspace) Open(uri, text string) { w.buffers[uri] = text }

// CloseDoc 丢弃文档缓冲（回到磁盘内容）。
func (w *Workspace) CloseDoc(uri string) {
	delete(w.buffers, uri)
}

// --------------------------------------------------------------------------- //
// URI 与路径
// --------------------------------------------------------------------------- //

// URI2Path 把 file:// URI 转成本机路径。
func URI2Path(uri string) string {
	if uri == "" {
		return ""
	}
	s := strings.TrimPrefix(uri, "file://")
	if runtime.GOOS == "windows" {
		s = strings.TrimPrefix(s, "/")
		s = strings.ReplaceAll(s, "/", "\\")
		// 还原 %20 之类的转义（只处理常见几种，够用）
		s = strings.ReplaceAll(s, "%20", " ")
	}
	return s
}

// Path2URI 把本机路径转成 file:// URI。
func Path2URI(path string) string {
	if path == "" {
		return ""
	}
	p := filepath.ToSlash(path)
	if runtime.GOOS == "windows" && !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return "file://" + p
}

// --------------------------------------------------------------------------- //
// 分析
// --------------------------------------------------------------------------- //

// projectRoot 从文件路径向上找含 aila.toml 的目录；找不到就用工作区根。
func (w *Workspace) projectRoot(path string) string {
	dir := path
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		dir = filepath.Dir(dir)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, manifest.FileName)); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return w.root
}

// analysisFor 返回（必要时重算）某个文件所属工程的分析结果。
func (w *Workspace) analysisFor(path string) *analysis {
	root := w.projectRoot(path)
	if a, ok := w.last[root]; ok {
		return a
	}
	a := w.analyze(root)
	w.last[root] = a
	return a
}

// invalidate 让下次查询重算（文档变化后调用）。
func (w *Workspace) invalidate(path string) {
	root := w.projectRoot(path)
	delete(w.last, root)
}

// analyze 解析并检查整个工程（打开中的文档用缓冲内容覆盖磁盘内容）。
func (w *Workspace) analyze(root string) *analysis {
	bag := &diag.Bag{}
	prj := project.Discover(root, bag)
	out := &analysis{Modules: map[string]*moduleInfo{}, ByPath: map[string]*moduleInfo{}}
	man := manifest.Load(root, bag)
	entry := "_.main"
	if man != nil && man.Entry != "" {
		entry = man.Entry
	}

	read := func(path string) (string, bool) {
		if text, ok := w.buffers[Path2URI(path)]; ok {
			return text, true
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", false
		}
		return string(data), true
	}

	mods := map[string]*ast.Module{}
	order := []string{}
	for _, name := range prj.ModuleList {
		path := prj.Modules[name]
		text, ok := read(path)
		if !ok {
			continue
		}
		mod, pbag := parser.Parse(path, text)
		mod.Name = name
		info := &moduleInfo{Name: name, Path: path, URI: Path2URI(path), Text: text, Mod: mod,
			Imports: map[string]string{}}
		for _, d := range pbag.Items {
			info.Diags = append(info.Diags, toDiagnostic(d))
		}
		mods[name] = mod
		order = append(order, name)
		out.Modules[name] = info
		out.ByPath[path] = info
	}
	// 标准库按需装载（与 build 的做法一致：`std.*` 走 AILA_HOME/stdlib）
	loaded := map[string]bool{}
	for _, name := range order {
		loaded[name] = true
	}
	for i := 0; i < len(order); i++ {
		info := out.Modules[order[i]]
		if info == nil || info.Mod == nil {
			continue
		}
		for _, imp := range info.Mod.Imports {
			// C 导入：`import ch _.cpacks.<名字>` → 复用 bindgen 生成 `.ail` 当模块索引；
			// `import cc` 只让 C 源参与编译，不产生绑定模块。
			if imp.KindV == "ch" || imp.KindV == "cc" {
				if imp.KindV == "cc" {
					continue
				}
				b := w.cBindingFor(root, prj, man, imp)
				if b != nil && len(b.Problems) > 0 {
					for _, d := range b.Problems {
						info.Diags = append(info.Diags, toDiagnostic(d))
					}
				}
				if b == nil || b.Text == "" || loaded[b.Name] {
					continue
				}
				mod, cbag := parser.Parse(b.Path, b.Text)
				mod.Name = b.Name
				sub := &moduleInfo{Name: b.Name, Path: b.Path, URI: Path2URI(b.Path), Text: b.Text,
					Mod: mod, Imports: map[string]string{}, Generated: true, Cpack: b.Cpack}
				for _, d := range cbag.Items {
					sub.Diags = append(sub.Diags, toDiagnostic(d))
				}
				mods[b.Name] = mod
				order = append(order, b.Name)
				out.Modules[b.Name] = sub
				out.ByPath[sub.Path] = sub
				loaded[b.Name] = true
				continue
			}
			target := importTargetModule(imp)
			if target == "" || loaded[target] {
				continue
			}
			path := prj.ModulePath(target)
			if path == "" {
				continue
			}
			text, ok := read(path)
			if !ok {
				continue
			}
			mod, pbag := parser.Parse(path, text)
			mod.Name = target
			sub := &moduleInfo{Name: target, Path: path, URI: Path2URI(path), Text: text, Mod: mod,
				Imports: map[string]string{}}
			for _, d := range pbag.Items {
				sub.Diags = append(sub.Diags, toDiagnostic(d))
			}
			_ = pbag
			mods[target] = mod
			order = append(order, target)
			out.Modules[target] = sub
			out.ByPath[path] = sub
			loaded[target] = true
		}
	}

	semaBag := &diag.Bag{}
	prog := sema.BuildProgram(mods, order, entry, semaBag)
	for _, d := range semaBag.Items {
		path := d.Span.File
		for _, info := range out.Modules {
			if info.Path == path {
				info.Diags = append(info.Diags, toDiagnostic(d))
				break
			}
		}
	}
	_ = prog
	for _, name := range order {
		info := out.Modules[name]
		if info == nil || info.Mod == nil {
			continue
		}
		u := prog.Unit(name)
		if u != nil {
			info.Sigs = u.Sigs
			info.Check = u.Check
		}
		info.Symbols = buildSymbols(info)
		for _, imp := range info.Mod.Imports {
			binding := importBindingName(imp)
			if binding == "" {
				continue
			}
			if target := importTargetModule(imp); target != "" {
				info.Imports[binding] = target
			}
		}
	}
	out.Order = order
	return out
}

// toDiagnostic 把编译器诊断转成 LSP 诊断。
func toDiagnostic(d diag.Diagnostic) Diagnostic {
	line := d.Span.Line
	col := d.Span.Col
	if line <= 0 {
		line = 1
	}
	if col <= 0 {
		col = 1
	}
	length := d.Span.Length
	if length <= 0 {
		length = 1
	}
	sev := 1
	if d.Severity == diag.SeverityWarning {
		sev = 2
	}
	return Diagnostic{
		Range: Range{
			Start: Position{Line: line - 1, Character: col - 1},
			End:   Position{Line: line - 1, Character: col - 1 + length},
		},
		Severity: sev,
		Code:     d.Code,
		Source:   "aila",
		Message:  d.Message,
	}
}

// buildSymbols 走一遍 AST，收集模块级声明与函数内的局部绑定。
func buildSymbols(info *moduleInfo) []*symbol {
	var out []*symbol
	add := func(name string, kind int, span diag.Span, container, detail string, local, public bool) {
		if name == "" || name == "_" {
			return
		}
		out = append(out, &symbol{Name: name, Kind: kind, Range: spanRange(span), Module: info.Name,
			Container: container, Detail: detail, Local: local, Public: public})
	}
	// crossPublic 判断一条声明是否能跨包可见（壳内 + public）。
	crossPublic := func(a ast.Attrs) bool { return a.IsPublic && a.InShell }
	for _, d := range info.Mod.AllDecls() {
		switch x := d.(type) {
		case *ast.FuncDecl:
			kind := KindFunction
			if x.OwnerType != "" {
				kind = KindMethod
			}
			display := x.Name
			if x.OwnerType != "" {
				display = x.OwnerType + "." + x.Name
			}
			detail := funcDetail(info, x)
			add(display, kind, x.Span, x.OwnerType, detail, false, crossPublic(x.Attrs))
			collectLocals(x, func(name string, kind int, span diag.Span, detail string) {
				add(name, kind, span, display, detail, true, false)
			})
		case *ast.StructDecl:
			pub := crossPublic(x.Attrs)
			add(x.Name, KindStruct, x.Span, "", "struct", false, pub)
			for _, f := range x.Fields_ {
				add(f.Name, KindField, f.Span, x.Name, "字段", false, pub)
			}
		case *ast.EnumDecl:
			pub := crossPublic(x.Attrs)
			add(x.Name, KindEnum, x.Span, "", "enum", false, pub)
			for _, v := range x.Variants {
				add(x.Name+"."+v.Name, KindEnumMember, v.Span, x.Name, "变体", false, pub)
			}
		case *ast.TypeAlias:
			add(x.Name, KindTypeParam, x.Span, "", "type 别名", false, crossPublic(x.Attrs))
		case *ast.LetDecl:
			kind := KindConstant
			if x.KindV != "const" {
				kind = KindVariable
			}
			add(x.Name, kind, x.Span, "", x.KindV, false, crossPublic(x.Attrs))
		}
	}
	return out
}

// collectLocals 收集函数内的参数与局部绑定。
func collectLocals(fn *ast.FuncDecl, emit func(name string, kind int, span diag.Span, detail string)) {
	for _, p := range fn.Params {
		emit(p.Name, KindVariable, p.Span, "参数")
	}
	var walkBlock func(b *ast.Block)
	walkStmt := func(st ast.Node) {}
	walkBlock = func(b *ast.Block) {
		if b == nil {
			return
		}
		for _, st := range b.Stmts {
			switch x := st.(type) {
			case *ast.LetDecl:
				kind := KindVariable
				if x.KindV == "const" {
					kind = KindConstant
				}
				emit(x.Name, kind, x.Span, x.KindV)
			case *ast.Block:
				walkBlock(x)
			case *ast.If:
				walkBlock(x.Then)
				for _, pair := range x.Elifs {
					if blk, ok := pair.B.(*ast.Block); ok {
						walkBlock(blk)
					}
				}
				walkBlock(x.Els)
			case *ast.ForC:
				if x.Init != nil {
					emit(x.Init.Name, KindVariable, x.Init.Span, "循环变量")
				}
				walkBlock(x.Body)
			case *ast.ForIn:
				if x.Index != "" {
					emit(x.Index, KindVariable, x.Span, "下标")
				}
				emit(x.Var, KindVariable, x.Span, "元素")
				walkBlock(x.Body)
			case *ast.Region:
				walkBlock(x.Body)
			case *ast.Unsafe:
				walkBlock(x.Body)
			case *ast.Defer:
				walkBlock(x.Body)
			case *ast.Match:
				for _, arm := range x.Arms {
					if blk, ok := arm.Body.(*ast.Block); ok {
						walkBlock(blk)
					}
				}
			}
		}
	}
	_ = walkStmt
	walkBlock(fn.Body)
}

// funcDetail 生成函数签名摘要（hover 与大纲共用）。
func funcDetail(info *moduleInfo, fn *ast.FuncDecl) string {
	if info.Sigs != nil {
		key := fn.Name
		sig := info.Sigs.Funcs[key]
		if fn.OwnerType != "" {
			sig = info.Sigs.Methods[fn.OwnerType+"."+fn.Name]
		}
		if sig != nil {
			return sigText(sig)
		}
	}
	return "func"
}

// sigText 把签名渲染成「名字(参数) -> 返回」。
func sigText(sig *sema.FuncSig) string {
	var b strings.Builder
	b.WriteString(sig.Name)
	b.WriteString("(")
	for i, p := range sig.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(p.Name)
		b.WriteString(": ")
		b.WriteString(types.TypeName(p.Type))
	}
	b.WriteString(")")
	if sig.RetGroup != nil {
		parts := make([]string, 0, len(sig.RetGroup))
		for _, t := range sig.RetGroup {
			parts = append(parts, types.TypeName(t))
		}
		b.WriteString(" -> (" + strings.Join(parts, ", ") + ")")
	} else if !isVoidType(sig.Ret) {
		b.WriteString(" -> " + types.TypeName(sig.Ret))
	}
	if sig.Err != nil {
		b.WriteString(" ! " + types.TypeName(sig.Err))
	}
	return b.String()
}

func isVoidType(t types.Type) bool {
	_, ok := t.(types.Void)
	return ok
}

// spanRange 把编译器位置转成 LSP 区间（1 基 → 0 基）。
func spanRange(s diag.Span) Range {
	line := s.Line
	col := s.Col
	if line <= 0 {
		line = 1
	}
	if col <= 0 {
		col = 1
	}
	length := s.Length
	if length <= 0 {
		length = 1
	}
	return Range{
		Start: Position{Line: line - 1, Character: col - 1},
		End:   Position{Line: line - 1, Character: col - 1 + length},
	}
}

// identifierAt 取指定位置上的标识符（含前后缀）。
func identifierAt(text string, pos Position) (string, Range, bool) {
	lines := strings.Split(text, "\n")
	if pos.Line < 0 || pos.Line >= len(lines) {
		return "", Range{}, false
	}
	line := strings.TrimRight(lines[pos.Line], "\r")
	if pos.Character < 0 || pos.Character > len(line) {
		return "", Range{}, false
	}
	isIdent := func(b byte) bool {
		return b == '_' || b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
	}
	start := pos.Character
	for start > 0 && isIdent(line[start-1]) {
		start--
	}
	end := pos.Character
	for end < len(line) && isIdent(line[end]) {
		end++
	}
	if start == end {
		return "", Range{}, false
	}
	word := line[start:end]
	if !unicode.IsLetter(rune(word[0])) && word[0] != '_' {
		return "", Range{}, false
	}
	return word, Range{
		Start: Position{Line: pos.Line, Character: start},
		End:   Position{Line: pos.Line, Character: end},
	}, true
}

// qualifierBefore 取 `qualifier.` 里的 qualifier（若紧邻标识符前是点）。
func qualifierBefore(text string, pos Position) string {
	lines := strings.Split(text, "\n")
	if pos.Line < 0 || pos.Line >= len(lines) {
		return ""
	}
	line := strings.TrimRight(lines[pos.Line], "\r")
	isIdent := func(b byte) bool {
		return b == '_' || b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
	}
	i := pos.Character
	for i > 0 && isIdent(line[i-1]) {
		i--
	}
	if i == 0 || line[i-1] != '.' {
		return ""
	}
	j := i - 1
	for j > 0 && isIdent(line[j-1]) {
		j--
	}
	return line[j : i-1]
}

// findSymbol 在模块里找符号（Local 标记区分作用域）。
func findSymbol(info *moduleInfo, name string, local bool) *symbol {
	var hit *symbol
	for _, s := range info.Symbols {
		if s.Name != name || s.Local != local {
			continue
		}
		hit = s
	}
	return hit
}

// symbolAt 找「包含该位置」的符号（用于 hover/definition 定位声明）。
func symbolAt(info *moduleInfo, pos Position) *symbol {
	var best *symbol
	for _, s := range info.Symbols {
		if within(s.Range, pos) {
			if best == nil || spanLess(best.Range, s.Range) {
				best = s
			}
		}
	}
	return best
}

func within(r Range, p Position) bool {
	if p.Line < r.Start.Line || p.Line > r.End.Line {
		return false
	}
	if p.Line == r.Start.Line && p.Character < r.Start.Character {
		return false
	}
	if p.Line == r.End.Line && p.Character > r.End.Character {
		return false
	}
	return true
}

// spanLess 判断 a 是否比 b 更靠前（用于取最小范围）。
func spanLess(a, b Range) bool {
	if a.Start.Line != b.Start.Line {
		return a.Start.Line > b.Start.Line
	}
	return a.Start.Character > b.Start.Character
}

// sortSymbols 让结果稳定（按位置）。
func sortSymbols(list []*symbol) {
	sort.SliceStable(list, func(i, j int) bool {
		return spanLess(list[j].Range, list[i].Range)
	})
}
