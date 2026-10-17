// Package bindgen 把 C 头文件翻成 AILA `.ail` 中间绑定模块（docs/specs/0001 §6）。
//
// 两档后端（docs/rfcs/0001 §3.3）：
//   - `clang`：`clang -Xclang -ast-dump=json -fsyntax-only` 取 AST（首选，覆盖全）；
//   - `builtin`：内置 C 声明子集解析器（无 clang 时的兜底，边界逐条诊断）。
//
// 两档共用同一套类型映射与命名规范，所以 AILA 侧代码不感知档位。
package bindgen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ailang/aila/internal/diag"
)

// Target 是绑定生成需要的目标信息。
type Target struct {
	Board   string
	Triple  string
	PtrSize int
}

// Result 是一次绑定生成的结果。
type Result struct {
	Text      string
	Backend   string   // "clang" | "builtin"
	Notes     []string // 人读提示（宏定型 / 跳过项）
	HashInput string
}

// Header 生成 `.ail` 文件头的注释块。
func (r *Result) Header(cpack string, t Target) string {
	lines := []string{
		"// 由 aila bindgen 自动生成，请勿手改",
		"// cpack: " + cpack + "   backend: " + r.Backend + "   target: " + t.Board + " (" + t.Triple + ")",
		"// C 实现本体由 aila build 用 cc 编译并自动链接（共置即关联）",
	}
	for _, note := range r.Notes {
		lines = append(lines, "// "+note)
	}
	return strings.Join(lines, "\n")
}

// --------------------------------------------------------------------------- //
// 命名规范（docs/specs/0001 §6.3）
// --------------------------------------------------------------------------- //

// ailaReserved 是与 AILA 关键字 / 内建类型同名的 C 标识符（映射时改名）。
var ailaReserved = map[string]bool{
	"func": true, "let": true, "var": true, "const": true, "type": true, "struct": true,
	"enum": true, "state": true, "view": true, "module": true, "hot": true, "import": true,
	"public": true, "return": true, "if": true, "else": true, "for": true, "in": true,
	"match": true, "region": true, "defer": true, "unsafe": true, "coroutine": true,
	"thread": true, "yield": true, "out": true, "trap": true, "move": true, "len": true,
	"copy": true, "resume": true, "as": true, "trunc": true, "break": true, "continue": true,
	"true": true, "false": true,
	"I8": true, "I16": true, "I32": true, "I64": true, "U8": true, "U16": true, "U32": true,
	"U64": true, "F32": true, "F64": true, "Bool": true, "C32": true, "Text": true,
	"USize": true, "ISize": true, "Unsafe": true, "chan": true, "send": true, "recv": true,
	"coro": true, "self": true,
}

// TypeName 把 C 类型名映射成 AILA 类型名（帕斯卡命名，去掉 `_t` 后缀）。
func TypeName(cName string) string {
	base := strings.TrimSpace(cName)
	if strings.HasSuffix(base, "_t") && len(base) > 2 {
		base = base[:len(base)-2]
	}
	parts := make([]string, 0, 4)
	for _, p := range strings.Split(base, "_") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	pascal := b.String()
	if pascal == "" {
		pascal = "Anon"
	}
	if ailaReserved[pascal] {
		pascal += "_c"
	}
	return pascal
}

// ValueName 把 C 值名映射成 AILA 名字（撞保留字时加 `_c`）。
func ValueName(cName string) string {
	if ailaReserved[cName] {
		return cName + "_c"
	}
	return cName
}

// --------------------------------------------------------------------------- //
// C 类型 → AILA 类型
// --------------------------------------------------------------------------- //

var basePrims = map[string]string{
	"bool": "Bool", "_Bool": "Bool",
	"char":       "U8",
	"signed char": "I8", "unsigned char": "U8",
	"short": "I16", "unsigned short": "U16", "short int": "I16", "unsigned short int": "U16",
	"int": "I32", "unsigned": "U32", "unsigned int": "U32", "signed": "I32", "signed int": "I32",
	"long long": "I64", "unsigned long long": "U64",
	"long long int": "I64", "unsigned long long int": "U64",
	"float": "F32", "double": "F64",
	"size_t": "USize", "ssize_t": "ISize", "ptrdiff_t": "ISize",
	"uintptr_t": "USize", "intptr_t": "ISize",
	"int8_t": "I8", "int16_t": "I16", "int32_t": "I32", "int64_t": "I64",
	"uint8_t": "U8", "uint16_t": "U16", "uint32_t": "U32", "uint64_t": "U64",
}

// BaseType 把裸类型文本映射成 AILA 类型名（nil 表示 void）。
func BaseType(text string, target Target, named map[string]string) *string {
	t := strings.Join(strings.Fields(strings.ReplaceAll(text, "*", " ")), " ")
	if p, ok := basePrims[t]; ok {
		return &p
	}
	if t == "long" || t == "long int" {
		v := "I32"
		if target.PtrSize == 8 {
			v = "I64"
		}
		return &v
	}
	if t == "unsigned long" || t == "unsigned long int" {
		v := "U32"
		if target.PtrSize == 8 {
			v = "U64"
		}
		return &v
	}
	if v, ok := named[t]; ok && v != "" {
		return &v
	}
	for _, prefix := range []string{"struct ", "enum ", "union "} {
		if strings.HasPrefix(t, prefix) {
			key := strings.TrimPrefix(t, prefix)
			if v, ok := named[key]; ok && v != "" {
				return &v
			}
			v := TypeName(key)
			return &v
		}
	}
	if t == "void" || t == "" {
		return nil
	}
	if v, ok := named[t]; ok && v != "" {
		return &v
	}
	v := TypeName(t)
	return &v
}

// MapCType 把 C 声明片段映射成 AILA 类型文本（`（void）` 表示无返回）。
func MapCType(text string, target Target, named map[string]string, bag *diag.Bag, span diag.Span, cpack string) string {
	raw := strings.Join(strings.Fields(text), " ")
	isFuncPtr := strings.Contains(raw, "(*") || (strings.Contains(raw, "(") && strings.Contains(raw, "*"))
	clean := stripQualifiers(raw)
	if isFuncPtr {
		bag.Warning("W3001", span, "函数指针类型 `"+raw+"` 未映射",
			"bindgen builtin 档不支持函数指针字段", "写 C shim 包装函数，或装 clang 用 clang 档")
		return "[Unsafe]?"
	}
	stars := strings.Count(clean, "*")
	base := strings.TrimSpace(strings.ReplaceAll(clean, "*", " "))
	if i := strings.Index(base, "["); i >= 0 {
		base = strings.TrimSpace(base[:i])
	}
	bt := BaseType(base, target, named)
	if stars == 0 {
		if bt == nil {
			return "（void）"
		}
		return *bt
	}
	if bt == nil {
		return "[Unsafe]?"
	}
	if stars > 1 {
		return "[Unsafe]?"
	}
	switch strings.TrimSpace(base) {
	case "char", "signed char", "unsigned char":
		return "[U8]?"
	}
	return "[" + *bt + "]?"
}

// stripQualifiers 去掉 C 的存储类 / 限定符。
func stripQualifiers(raw string) string {
	var kept []string
	for _, w := range strings.Fields(raw) {
		switch w {
		case "const", "volatile", "restrict", "static", "extern", "inline", "register",
			"__restrict", "__restrict__", "_Nonnull":
			continue
		}
		kept = append(kept, w)
	}
	return strings.Join(kept, " ")
}

// --------------------------------------------------------------------------- //
// 文件发现
// --------------------------------------------------------------------------- //

// FindHeaders 收集 cpack 里的头文件（`include/` 优先，余下递归）。
func FindHeaders(cpackDir string) []string {
	out := []string{}
	seen := map[string]bool{}
	var walk func(root string)
	walk = func(root string) {
		entries, err := os.ReadDir(root)
		if err != nil {
			return
		}
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), ".") {
				continue
			}
			names = append(names, e.Name())
		}
		sort.Strings(names)
		for _, name := range names {
			full := filepath.Join(root, name)
			st, err := os.Stat(full)
			if err != nil {
				continue
			}
			if st.IsDir() {
				walk(full)
				continue
			}
			if strings.HasSuffix(name, ".h") || strings.HasSuffix(name, ".hpp") {
				abs, err := filepath.Abs(full)
				if err != nil {
					abs = full
				}
				if !seen[abs] {
					seen[abs] = true
					out = append(out, abs)
				}
			}
		}
	}
	inc := filepath.Join(cpackDir, "include")
	if st, err := os.Stat(inc); err == nil && st.IsDir() {
		walk(inc)
	}
	walk(cpackDir)
	return out
}

// FindCSources 收集 cpack 里的 C 源文件。
func FindCSources(cpackDir string) []string {
	return findFiles(cpackDir, func(name string) bool { return strings.HasSuffix(name, ".c") })
}

// FindLibs 收集 cpack 里的预编译库。
func FindLibs(cpackDir string) []string {
	return findFiles(cpackDir, func(name string) bool {
		for _, suffix := range []string{".a", ".lib", ".so", ".dll", ".dylib"} {
			if strings.HasSuffix(name, suffix) {
				return true
			}
		}
		return false
	})
}

func findFiles(cpackDir string, match func(string) bool) []string {
	out := []string{}
	var walk func(root string)
	walk = func(root string) {
		entries, err := os.ReadDir(root)
		if err != nil {
			return
		}
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			if !strings.HasPrefix(e.Name(), ".") {
				names = append(names, e.Name())
			}
		}
		sort.Strings(names)
		for _, name := range names {
			full := filepath.Join(root, name)
			st, err := os.Stat(full)
			if err != nil {
				continue
			}
			if st.IsDir() {
				walk(full)
				continue
			}
			if match(name) {
				out = append(out, full)
			}
		}
	}
	walk(cpackDir)
	return out
}

// Run 是绑定生成入口：优先 clang 档，失败或不可用则退到 builtin 档。
func Run(cpack, cpackDir string, target Target, bag *diag.Bag, preference string) *Result {
	hashInput := buildHashInput(cpackDir)
	if preference == "" {
		preference = "auto"
	}
	if preference == "auto" || preference == "clang" {
		if clangPath := findClang(); clangPath != "" {
			if res := runClang(cpack, cpackDir, target, bag, clangPath); res != nil {
				res.HashInput = hashInput
				return res
			}
		} else if preference == "clang" {
			bag.Error("E8011", diag.Empty, "指定了 `--backend clang` 但找不到 clang",
				"这里", "安装 clang，或改用 `--backend builtin`")
		}
	}
	res := runBuiltin(cpack, cpackDir, target, bag)
	if res.Backend == "builtin" {
		res.Notes = append(res.Notes, "builtin 兜底档：覆盖范围窄于 clang 档（见 docs/specs/0001 §6.5）")
	}
	res.HashInput = hashInput
	return res
}

// buildHashInput 生成缓存键的输入（头文件路径 + 内容摘要）。
func buildHashInput(cpackDir string) string {
	var b strings.Builder
	for _, header := range FindHeaders(cpackDir) {
		data, err := os.ReadFile(header)
		if err != nil {
			continue
		}
		fmt.Fprintf(&b, "%s:%d:%d;", header, len(data), checksum(data))
	}
	return b.String()
}

func checksum(data []byte) uint32 {
	var h uint32 = 2166136261
	for _, c := range data {
		h ^= uint32(c)
		h *= 16777619
	}
	return h
}
