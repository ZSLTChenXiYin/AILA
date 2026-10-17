// Package sema 是语义分析的入口。G3 先落地「类型表」（声明收集 + TypeRef → Type 解析），
// 后续再补名字解析、类型检查、区域效果与 IR 生成。
package sema

import (
	"fmt"
	"strings"

	"github.com/aila-lang/aila/internal/ast"
	"github.com/aila-lang/aila/internal/diag"
	"github.com/aila-lang/aila/internal/types"
)

// Table 是一个模块的类型表。
type Table struct {
	Types   map[string]types.Info
	Aliases map[string]types.Type
	// Imports 是「导入绑定名 -> 被导入模块的类型表」：跨模块类型名的解析入口。
	Imports map[string]*Table
	// Pkg 是本表所属包名（`_` / 依赖名 / `std`），跨包可见性检查要用。
	Pkg string
	order   []string // 声明顺序（别名解析要用）
}

// BuildTable 收集类型声明并解析其体（结构体字段 / 枚举变体 / 别名目标）。
//
// 与 Python 参考实现的两段式一致：先全部登记名字，再逐个解析体内引用。
func BuildTable(mod *ast.Module, bag *diag.Bag) *Table {
	t := DeclareTypes(mod, bag)
	t.ResolveBodies(mod, bag)
	return t
}

// DeclareTypes 只登记类型名字（跨模块链接需要「先声明、后解析」两段式）。
func DeclareTypes(mod *ast.Module, bag *diag.Bag) *Table {
	t := &Table{Types: map[string]types.Info{}, Aliases: map[string]types.Type{},
		Imports: map[string]*Table{}, Pkg: PackageOf(mod.Name)}
	for _, d := range mod.AllDecls() {
		switch x := d.(type) {
		case *ast.StructDecl:
			if _, dup := t.Types[x.Name]; dup {
				bag.Error("E3001", x.Span, "类型 `"+x.Name+"` 重复定义", "这里", "")
				continue
			}
			info := &types.StructInfo{Name: x.Name, Module: mod.Name, IsPublic: x.Attrs.IsPublic,
				Packed: hasAttr(x.Attrs.Attrs, "packed")}
			if n, ok := attrInt(x.Attrs.Attrs, "align"); ok {
				info.Align = n
			}
			t.Types[x.Name] = info
			t.order = append(t.order, x.Name)
		case *ast.EnumDecl:
			if _, dup := t.Types[x.Name]; dup {
				bag.Error("E3001", x.Span, "类型 `"+x.Name+"` 重复定义", "这里", "")
				continue
			}
			info := &types.EnumInfo{Name: x.Name, Module: mod.Name, IsPublic: x.Attrs.IsPublic,
				Base: types.U32, TagType: types.U32}
			t.Types[x.Name] = info
			t.order = append(t.order, x.Name)
		case *ast.TypeAlias:
			if _, dup := t.Aliases[x.Name]; dup {
				bag.Error("E3001", x.Span, "类型别名 `"+x.Name+"` 重复定义", "这里", "")
				continue
			}
			t.Aliases[x.Name] = nil // 占位，稍后解析
			t.order = append(t.order, x.Name)
		}
	}
	return t
}

// ResolveBodies 解析类型体（结构体字段 / 枚举变体 / 别名目标）。
func (t *Table) ResolveBodies(mod *ast.Module, bag *diag.Bag) {
	for _, d := range mod.AllDecls() {
		switch x := d.(type) {
		case *ast.StructDecl:
			info, _ := t.Types[x.Name].(*types.StructInfo)
			if info == nil {
				continue
			}
			seen := map[string]bool{}
			fields := make([]types.Field, 0, len(x.Fields_))
			for _, f := range x.Fields_ {
				if seen[f.Name] {
					bag.Error("E3001", f.Span, "字段 `"+f.Name+"` 重复", "这里", "")
				}
				seen[f.Name] = true
				fields = append(fields, types.Field{Name: f.Name, Type: t.Resolve(f.Type, bag)})
			}
			info.Fields = fields
		case *ast.EnumDecl:
			t.buildEnum(x, bag)
		case *ast.TypeAlias:
			t.Aliases[x.Name] = t.Resolve(x.Target, bag)
		}
	}
}

func (t *Table) buildEnum(x *ast.EnumDecl, bag *diag.Bag) {
	info, _ := t.Types[x.Name].(*types.EnumInfo)
	if info == nil {
		return
	}
	if x.Base != nil {
		bt := t.Resolve(x.Base, bag)
		if !types.IsInteger(bt) {
			bag.Error("E3201", x.Base.Span, "enum 底层类型必须是无符号整数", "这里", "")
			bt = types.U32
		}
		info.Base = bt
	}
	next := int64(0)
	seen := map[string]bool{}
	var payloads int
	for _, v := range x.Variants {
		if seen[v.Name] {
			bag.Error("E3001", v.Span, "变体 `"+v.Name+"` 重复", "这里", "")
		}
		seen[v.Name] = true
		if v.Discriminant != nil {
			if d, ok := constInt(v.Discriminant); ok {
				next = d
			} else {
				bag.Error("E3201", v.Span, "判别值必须是编译期整数常量", "这里", "")
			}
		}
		vi := types.VariantInfo{Name: v.Name, Tag: next}
		next++
		if len(v.Payload) > 0 {
			fields := make([]types.Field, 0, len(v.Payload))
			for _, f := range v.Payload {
				fields = append(fields, types.Field{Name: f.Name, Type: t.Resolve(f.Type, bag)})
			}
			vi.Payload = fields
			payloads++
		}
		info.Variants = append(info.Variants, vi)
	}
	info.Payload = payloads > 0
	var maxTag int64
	for _, v := range info.Variants {
		if v.Tag > maxTag {
			maxTag = v.Tag
		}
	}
	switch {
	case !info.Payload:
		// 无载荷 enum：宽度 = 底层类型宽度（designs/0001 §3.5 规则 5）
		info.TagType = info.Base
	case x.Base != nil:
		// 带载荷 + 显式底层类型：作者指定优先，但必须容纳全部判别值
		info.TagType = info.Base
		if bits := types.IntWidth(info.Base, 8); bits > 0 && maxTag >= (int64(1) << uint(bits)) {
			bag.Error("E3201", x.Span,
				fmt.Sprintf("判别值 %d 超出底层类型 `%s` 的宽度", maxTag, types.TypeName(info.Base)),
				"tag 宽度必须容纳全部判别值", "")
		}
	default:
		width := 1
		for (int64(1) << (width * 8)) <= maxTag {
			width *= 2
		}
		kinds := map[int]string{1: "u8", 2: "u16", 4: "u32", 8: "u64"}
		if k, ok := kinds[width]; ok {
			if p, ok2 := types.PrimByKind(k); ok2 {
				info.TagType = p
			}
		}
	}
	// 判别值重复（v0 禁止别名判别值）
	tags := map[int64]string{}
	for _, v := range info.Variants {
		if prev, dup := tags[v.Tag]; dup {
			bag.Error("E3401", x.Span,
				fmt.Sprintf("判别值 %d 与变体 `%s` 重复", v.Tag, prev),
				"v0 禁止别名判别值", "改用不同判别值")
		}
		tags[v.Tag] = v.Name
	}
}

// Resolve 把语法层类型引用解析成类型项。
func (t *Table) Resolve(ref *ast.TypeRef, bag *diag.Bag) types.Type {
	if ref == nil {
		return types.Void{}
	}
	switch ref.KindV {
	case "prim":
		if p, ok := types.PrimByKind(ref.Name); ok {
			return p
		}
		return types.I32
	case "view":
		return types.View{Elem: t.Resolve(ref.Elem, bag)}
	case "fixed":
		n, ok := constInt(ref.Size)
		if !ok || n < 0 {
			bag.Error("E3201", ref.Span, "定长数组长度必须是编译期非负整数", "这里", "")
			n = 0
		}
		return types.Fixed{Elem: t.Resolve(ref.Elem, bag), N: n}
	case "unsafe":
		return types.UnsafeView{}
	case "optional":
		inner := t.Resolve(ref.Inner, bag)
		if !types.IsHandle(inner) {
			bag.Error("E3601", ref.Span, "`"+types.TypeName(inner)+"` 不可为空",
				"`T?` 只给句柄类（视图 / Text / move struct / enum / 函数指针 / 协程句柄）",
				"标量没有空值需求；要表达「可能没有」请用枚举或 `T ! E`")
		}
		return types.Optional{Inner: inner}
	case "result":
		ok := t.Resolve(ref.Inner, bag)
		errT := t.Resolve(ref.Err, bag)
		if _, isEnum := errT.(types.Named); !isEnum {
			bag.Error("E3201", ref.Span, "`T ! E` 的 E 必须是单一 enum 类型", "这里",
				"错误上下文写在 enum 载荷字段里")
		}
		return types.Result{Ok: ok, Err: errT}
	case "chan":
		return types.Chan{Elem: t.Resolve(ref.Elem, bag), Dir: ref.Dir}
	case "coro":
		c := types.Coro{}
		if ref.Inner != nil {
			c.Ok = t.Resolve(ref.Inner, bag)
		}
		if ref.Err != nil {
			c.Err = t.Resolve(ref.Err, bag)
		}
		return c
	case "fnptr":
		params := make([]types.Type, 0, len(ref.Params))
		for _, p := range ref.Params {
			params = append(params, t.Resolve(p, bag))
		}
		f := types.FnPtr{Params: params}
		if ref.Ret != nil {
			ret := t.Resolve(ref.Ret, bag)
			if r, ok := ret.(types.Result); ok {
				f.Ret = r.Ok
				f.Err = r.Err
			} else {
				f.Ret = ret
			}
		}
		return f
	case "group":
		items := make([]types.Type, 0, len(ref.Items))
		for _, it := range ref.Items {
			items = append(items, t.Resolve(it, bag))
		}
		return types.Group{Items: items}
	case "name":
		if t != nil {
			if info, ok := t.Types[ref.Name]; ok {
				return types.Named{Info: info}
			}
			if alias, ok := t.Aliases[ref.Name]; ok && alias != nil {
				return alias
			}
			if binding, rest, qualified := splitQualified(ref.Name); qualified {
				if dep, ok := t.Imports[binding]; ok && dep != nil {
					if info, ok2 := dep.Types[rest]; ok2 {
						if !typeVisible(t.Pkg, info) {
							bag.Error("E7020", ref.Span, "类型 `"+ref.Name+"` 未导出（缺 `public`）",
								"跨包引用只认 `public` 声明",
								"在依赖包里给该类型加 `public`，或不要跨包用它")
						}
						return types.Named{Info: info}
					}
					if alias, ok2 := dep.Aliases[rest]; ok2 && alias != nil {
						return alias
					}
				}
			}
		}
		bag.Error("E3001", ref.Span, "未知类型 `"+ref.Name+"`", "类型名", "")
		return types.I32
	}
	return types.I32
}

// Lookup 按名字取类型（结构体 / 枚举 / 别名，含跨模块 `绑定.类型` 形式）。
func (t *Table) Lookup(name string) (types.Type, bool) {
	if t == nil {
		return nil, false
	}
	if info, ok := t.Types[name]; ok {
		return types.Named{Info: info}, true
	}
	if alias, ok := t.Aliases[name]; ok && alias != nil {
		return alias, true
	}
	if binding, rest, qualified := splitQualified(name); qualified {
		if dep, ok := t.Imports[binding]; ok && dep != nil {
			return dep.Lookup(rest)
		}
	}
	return nil, false
}

// splitQualified 拆开 `绑定.名字` 形式（取第一段作绑定名）。
func splitQualified(name string) (string, string, bool) {
	i := strings.Index(name, ".")
	if i <= 0 || i == len(name)-1 {
		return "", "", false
	}
	return name[:i], name[i+1:], true
}

// --------------------------------------------------------------------------- //
// 属性与常量求值
// --------------------------------------------------------------------------- //

func hasAttr(attrs []*ast.Attr, name string) bool {
	for _, a := range attrs {
		if a.Name == name {
			return true
		}
	}
	return false
}

func attrInt(attrs []*ast.Attr, name string) (int, bool) {
	for _, a := range attrs {
		if a.Name != name {
			continue
		}
		if len(a.Args) == 1 {
			if v, ok := constInt(a.Args[0]); ok {
				return int(v), true
			}
		}
	}
	return 0, false
}

// constInt 是最小编译期整数求值（字面量 / 一元负号 / 四则与位运算 / 括号）。
func constInt(n ast.Node) (int64, bool) {
	switch x := n.(type) {
	case *ast.IntLit:
		return x.Raw, true
	case *ast.CharLit:
		rs := []rune(x.Value)
		if len(rs) == 1 {
			return int64(rs[0]), true
		}
		return 0, false
	case *ast.Unary:
		v, ok := constInt(x.Operand)
		if !ok {
			return 0, false
		}
		switch x.Op {
		case "-":
			return -v, true
		case "+":
			return v, true
		case "~":
			return ^v, true
		}
		return 0, false
	case *ast.Cast:
		v, ok := constInt(x.Operand)
		if !ok {
			return 0, false
		}
		return v, true
	case *ast.Binary:
		a, ok1 := constInt(x.Left)
		b, ok2 := constInt(x.Right)
		if !ok1 || !ok2 {
			return 0, false
		}
		switch x.Op {
		case "+":
			return a + b, true
		case "-":
			return a - b, true
		case "*":
			return a * b, true
		case "/":
			if b == 0 {
				return 0, false
			}
			return a / b, true
		case "%":
			if b == 0 {
				return 0, false
			}
			return a % b, true
		case "<<":
			return a << uint64(b), true
		case ">>":
			return a >> uint64(b), true
		case "&":
			return a & b, true
		case "|":
			return a | b, true
		case "^":
			return a ^ b, true
		}
	}
	return 0, false
}

// PackageOf 从模块全名取包名：`_.main` → `_`，`mylib.utils` → `mylib`，`std.io` → `std`。
func PackageOf(module string) string {
	if module == "" {
		return "_"
	}
	if module == "_" {
		return "_"
	}
	if i := strings.Index(module, "."); i > 0 {
		return module[:i]
	}
	return module
}

// typeVisible 判断「从 pkg 里看 info 这个类型」是否可见：
// 同包恒可见；跨包要求类型声明带 `public`（designs/0001 §2.2 + docs/rfcs/0002 缺陷 14）。
func typeVisible(pkg string, info types.Info) bool {
	if info == nil {
		return true
	}
	mod := ""
	switch x := info.(type) {
	case *types.StructInfo:
		mod = x.Module
	case *types.EnumInfo:
		mod = x.Module
	}
	if mod == "" {
		return true
	}
	if PackageOf(mod) == pkg {
		return true
	}
	switch x := info.(type) {
	case *types.StructInfo:
		return x.IsPublic
	case *types.EnumInfo:
		return x.IsPublic
	}
	return true
}