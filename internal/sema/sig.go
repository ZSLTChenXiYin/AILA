package sema

import (
	"sort"

	"aila/internal/ast"
	"aila/internal/diag"
	"aila/internal/types"
)

// FuncSig 是函数签名（docs/specs/0002 §4 / §6 的承载结构）。
type FuncSig struct {
	Name string
	// Module 是声明该函数的模块名（跨模块调用的符号修饰要用它）。
	Module      string
	Owner       string
	Params      []Param
	Ret         types.Type
	Err         types.Type
	RetGroup    []types.Type
	Regions     []string
	Muts        []string
	Moves       []string
	IsCoroutine bool
	IsThread    bool
	IsExtern    bool
	IsPublic    bool
	InShell     bool
	// Impl 是壳里前向声明对应的块内实现（没有前向声明时为空）。
	// Decl 是签名来自的那条声明（壳内前向声明时就是它）。
	Decl     *ast.FuncDecl
	Impl     *ast.FuncDecl
	Variadic bool
	// VariadicElem 是 AILA 原生变参 `rest: ...T` 的元素类型 T（nil = 无原生变参）。
	// rest 本身作为普通参数留在 Params 末位，类型是 [T] 视图（designs/0002 §6）。
	VariadicElem types.Type
	Span         diag.Span
}

// Param 是一个形参。
type Param struct {
	Name string
	Type types.Type
}

// FullName 返回 `Owner.name` 或 `name`。
func (f *FuncSig) FullName() string {
	if f.Owner != "" {
		return f.Owner + "." + f.Name
	}
	return f.Name
}

// ParamIndex 按名字找形参下标（找不到返回 -1）。
func (f *FuncSig) ParamIndex(name string) int {
	for i, p := range f.Params {
		if p.Name == name {
			return i
		}
	}
	return -1
}

// ConstInfo 是模块级常量。
type ConstInfo struct {
	Name     string
	Type     types.Type
	Value    any
	IsPublic bool
	InShell  bool
}

// GlobalInfo 是模块级绑定（var / let / state）。
type GlobalInfo struct {
	Name     string
	Type     types.Type
	Mutable  bool
	Kind     string // var | let | state
	IsPublic bool
	InShell  bool
}

// Sigs 是一个模块的签名与模块级值表。
type Sigs struct {
	Funcs   map[string]*FuncSig
	Methods map[string]*FuncSig // key = "Owner.name"
	Consts  map[string]*ConstInfo
	// ConstOrder 是 const 的声明顺序（IR 常量池顺序按它走，保证与参考实现一致）。
	ConstOrder []string
	Globals    map[string]*GlobalInfo
}

// BuildSigs 从 AST + 类型表构建签名表。
//
// 与 Python 参考实现一致：参数 = 数据流、`@` 段 = 效果、返回 = 产出；
// `self` 免标注，类型恒为前缀类型；`-> (T1, T2)` 产多值组；`-> T ! E` 拆出错误通道。
func BuildSigs(mod *ast.Module, table *Table, bag *diag.Bag) *Sigs {
	s := &Sigs{
		Funcs:   map[string]*FuncSig{},
		Methods: map[string]*FuncSig{},
		Consts:  map[string]*ConstInfo{},
		Globals: map[string]*GlobalInfo{},
	}
	for _, d := range mod.AllDecls() {
		switch x := d.(type) {
		case *ast.FuncDecl:
			// 壳里的前向声明（`public func f(...) -> T;`）与块内实现配对：
			// 签名只登记一次，可见性取声明、函数体取实现（designs/0001 §2.1 壳/块分工）。
			if x.OwnerType == "" {
				if prev, ok := s.Funcs[x.Name]; ok {
					if prev.Decl == nil || x.Body == nil || prev.Impl != nil {
						bag.Error("E3001", x.Span, "函数 `"+x.Name+"` 重复定义", "这里", "")
						continue
					}
					prev.Impl = x
					continue
				}
			} else {
				key := x.OwnerType + "." + x.Name
				if prev, ok := s.Methods[key]; ok {
					if prev.Decl == nil || x.Body == nil || prev.Impl != nil {
						bag.Error("E3001", x.Span, "方法 `"+key+"` 重复定义", "这里", "")
						continue
					}
					prev.Impl = x
					continue
				}
			}
			sig := buildFuncSig(x, table, bag, mod.Name)
			sig.Decl = x
			if x.OwnerType == "" {
				s.Funcs[x.Name] = sig
			} else {
				if _, ok := table.Types[x.OwnerType]; !ok {
					if _, ok2 := table.Aliases[x.OwnerType]; !ok2 {
						bag.Error("E3001", x.Span, "方法前缀的类型 `"+x.OwnerType+"` 未定义", "这里", "")
					}
				}
				key := x.OwnerType + "." + x.Name
				if _, dup := s.Methods[key]; dup {
					bag.Error("E3001", x.Span, "方法 `"+key+"` 重复定义", "这里", "")
					continue
				}
				s.Methods[key] = sig
			}
		case *ast.LetDecl:
			if x.KindV == "const" {
				var t types.Type
				if x.Type != nil {
					t = table.Resolve(x.Type, bag)
				}
				v, hasVal := constValue(x.Init)
				if t == nil {
					t = inferConstType(v, hasVal)
				}
				if !hasVal {
					bag.Error("E3201", x.Span, "`const` 初始化器必须是编译期常量", "这里", "")
				}
				s.Consts[x.Name] = &ConstInfo{Name: x.Name, Type: t, Value: v, IsPublic: x.IsPublic, InShell: x.InShell}
				s.ConstOrder = append(s.ConstOrder, x.Name)
				continue
			}
			var t types.Type
			if x.Type != nil {
				t = table.Resolve(x.Type, bag)
			} else {
				t = types.I32
			}
			s.Globals[x.Name] = &GlobalInfo{Name: x.Name, Type: t,
				// state 与 var 逐条同语法（designs/0001 §4.2 规则 11），因此同样可写
				Mutable: x.KindV != "const", Kind: x.KindV, IsPublic: x.IsPublic, InShell: x.InShell}
		}
	}
	return s
}

func buildFuncSig(x *ast.FuncDecl, table *Table, bag *diag.Bag, module string) *FuncSig {
	// 无名 `...` 是 C 调用约定变参，只允许 #[c] extern 声明；
	// AILA 原生变参必须有名字与元素类型（`rest: ...T`，designs/0002 §6）。
	if x.Variadic && !x.IsExtern {
		bag.Error("E3201", x.Span, "函数 `"+x.Name+"` 的无名 `...` 只允许出现在 `#[c]` 外部声明上",
			"AILA 原生变参必须带名字与元素类型",
			"写成末位参数 `rest: ...T`，体内 `rest` 即 `[T]` 只读视图")
	}
	sig := &FuncSig{
		Module: module,
		Name:   x.Name, Owner: x.OwnerType,
		IsCoroutine: x.IsCoroutine, IsThread: x.IsThread, IsExtern: x.IsExtern,
		IsPublic: x.IsPublic, InShell: x.InShell, Variadic: x.Variadic, Span: x.Span,
	}
	for _, p := range x.Params {
		if p.IsSelf {
			if x.OwnerType == "" {
				bag.Error("E3001", p.Span, "`self` 参数只能用于类型前缀式方法", "这里",
					"写成 `func 类型.方法名(self, ...)`")
				sig.Params = append(sig.Params, Param{Name: "self", Type: types.I32})
				continue
			}
			t, ok := table.Lookup(x.OwnerType)
			if !ok {
				t = types.I32
			}
			sig.Params = append(sig.Params, Param{Name: "self", Type: t})
			continue
		}
		sig.Params = append(sig.Params, Param{Name: p.Name, Type: table.Resolve(p.Type, bag)})
	}
	// AILA 原生变参：末位参数是 rest 视图，记下元素类型供调用点实参检查。
	if n := len(x.Params); n > 0 && x.Params[n-1].IsVariadic {
		if v, ok := sig.Params[n-1].Type.(types.View); ok {
			sig.VariadicElem = v.Elem
		}
	}
	retRaw := types.Type(types.Void{})
	if x.Ret != nil {
		retRaw = table.Resolve(x.Ret, bag)
	}
	sig.Ret = retRaw
	if r, ok := retRaw.(types.Result); ok {
		sig.Ret = r.Ok
		sig.Err = r.Err
	}
	if g, ok := sig.Ret.(types.Group); ok {
		sig.RetGroup = g.Items
	}

	// @ 子句族：顺序固定 region → mut → move；同类多参数写单子句列表
	order := map[string]int{"region": 0, "mut": 1, "move": 2}
	prev := -1
	seen := []string{}
	for _, ef := range x.Effects {
		if len(seen) > 0 && seen[len(seen)-1] == ef.KindV {
			bag.Error("E4005", ef.Span, "`@"+ef.KindV+"` 子句重复",
				"同类多参数写成单子句列表：`@mut[a, b]`", "")
		} else if containsStr(seen, ef.KindV) {
			bag.Error("E4005", ef.Span, "`@` 子句顺序必须是 `@region → @mut → @move`",
				"能力 → 权限 → 转移，层次固定", "")
		}
		if cur := order[ef.KindV]; cur < prev {
			bag.Error("E4005", ef.Span, "`@` 子句顺序必须是 `@region → @mut → @move`",
				"能力 → 权限 → 转移，层次固定", "")
		}
		prev = order[ef.KindV]
		seen = append(seen, ef.KindV)
		switch ef.KindV {
		case "region":
			sig.Regions = append(sig.Regions, ef.Names...)
		case "mut":
			sig.Muts = append(sig.Muts, ef.Names...)
		case "move":
			sig.Moves = append(sig.Moves, ef.Names...)
		}
	}
	for _, n := range sig.Muts {
		for _, m := range sig.Moves {
			if n == m {
				bag.Error("E4005", x.Span, "参数 `"+n+"` 不能同时在 `@mut` 与 `@move` 中",
					"move ⊃ mut，重复声明冗余", "")
			}
		}
	}
	for _, n := range append(append([]string{}, sig.Muts...), sig.Moves...) {
		if sig.ParamIndex(n) < 0 {
			bag.Error("E4005", x.Span, "`@` 子句引用了不存在的参数 `"+n+"`", "这里", "")
		}
	}
	// 变参 rest 是调用期借用的只读视图，进 @mut/@move 无意义（designs/0002 §6）。
	if n := len(x.Params); n > 0 && x.Params[n-1].IsVariadic {
		restName := x.Params[n-1].Name
		if containsStr(sig.Muts, restName) || containsStr(sig.Moves, restName) {
			bag.Error("E4005", x.Span, "变参参数 `"+restName+"` 是只读视图，不能进入 `@mut`/`@move`",
				"变参视图是调用期借用，只读遍历", "")
		}
	}
	return sig
}

func inferConstType(v any, ok bool) types.Type {
	if !ok {
		return types.I32
	}
	switch v.(type) {
	case bool:
		return types.Bool
	case int64:
		return types.I32
	case float64:
		return types.F64
	case string:
		return types.Text
	}
	return types.I32
}

// constValue 求 const 初始化器的值（最小集合：标量字面量）。
func constValue(n ast.Node) (any, bool) {
	switch x := n.(type) {
	case *ast.IntLit:
		return x.Raw, true
	case *ast.FloatLit:
		return x.Value, true
	case *ast.BoolLit:
		return x.Value, true
	case *ast.StrLit:
		return x.Value, true
	case *ast.CharLit:
		return x.Value, true
	case *ast.Path:
		return nil, false
	case *ast.Unary:
		if v, ok := constValue(x.Operand); ok {
			switch x.Op {
			case "-":
				switch n := v.(type) {
				case int64:
					return -n, true
				case float64:
					return -n, true
				}
			case "+":
				return v, true
			}
		}
	}
	if v, ok := constInt(n); ok {
		return v, true
	}
	return nil, false
}

// SortedConstNames 便于快照对比（确定性输出）。
func SortedConstNames(s *Sigs) []string {
	names := make([]string, 0, len(s.Consts))
	for n := range s.Consts {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func containsStr(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
