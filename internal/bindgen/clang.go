package bindgen

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/aila-lang/aila/internal/diag"
	"github.com/aila-lang/aila/internal/resolve"
)

// findClang 返回可用的 clang 路径（随包优先，其次系统 PATH）。
func findClang() string {
	t := resolve.ResolveClang()
	if !t.OK() {
		return ""
	}
	return t.Path
}

// clangNode 是 clang `-ast-dump=json` 的节点（只取用到的字段）。
type clangNode struct {
	Kind              string         `json:"kind"`
	Name              string         `json:"name"`
	CompleteDef       bool           `json:"completeDefinition"`
	Type              clangType      `json:"type"`
	Inner             []*clangNode   `json:"inner"`
	Value             any            `json:"value"`
	Loc               clangLoc       `json:"loc"`
}

type clangType struct {
	QualType string `json:"qualType"`
}

type clangLoc struct {
	Line int `json:"line"`
	Col  int `json:"col"`
}

// runClang 用 clang 的 JSON AST 生成绑定；失败返回 nil（调用方回退 builtin 档）。
func runClang(cpack, cpackDir string, target Target, bag *diag.Bag, clangPath string) *Result {
	headers := FindHeaders(cpackDir)
	if len(headers) == 0 {
		return nil
	}
	res := &Result{Backend: "clang"}
	declared := map[string]string{}
	out := []string{}
	for _, header := range headers {
		ast, ok := clangAST(header, cpackDir, bag, clangPath)
		if !ok {
			return nil
		}
		for _, node := range ast.Inner {
			walkClang(node, &out, declared, target, bag, cpack)
		}
	}
	res.Text = strings.Join(out, "\n")
	return res
}

// clangAST 取一个头文件的 JSON AST。
func clangAST(header, cpackDir string, bag *diag.Bag, clangPath string) (*clangNode, bool) {
	args := []string{"-Xclang", "-ast-dump=json", "-fsyntax-only", header,
		"-I", cpackDir + "/include", "-I", cpackDir}
	cmd := exec.Command(clangPath, args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	raw, err := cmd.Output()
	if err != nil || len(raw) == 0 {
		first := ""
		if lines := strings.Split(strings.TrimSpace(stderr.String()), "\n"); len(lines) > 0 {
			first = strings.TrimSpace(lines[0])
		}
		if first == "" {
			first = "无输出"
		}
		bag.Warning("W3001", diag.Empty, "clang 解析头文件失败，回退 builtin 档", first, "")
		return nil, false
	}
	var node clangNode
	if err := json.Unmarshal(raw, &node); err != nil {
		return nil, false
	}
	return &node, true
}

// walkClang 递归遍历 clang AST，挑出要映射的声明。
func walkClang(node *clangNode, out *[]string, declared map[string]string, target Target,
	bag *diag.Bag, cpack string) {
	if node == nil {
		return
	}
	switch node.Kind {
	case "TypedefDecl":
		underlying := node.Type.QualType
		if strings.HasPrefix(underlying, "struct ") || strings.HasPrefix(underlying, "enum ") {
			for _, child := range node.Inner {
				if child.Kind == "RecordDecl" && child.CompleteDef {
					emitRecord(child, out, declared, target, bag, cpack, node.Name)
					return
				}
				if child.Kind == "EnumDecl" {
					emitEnum(child, out, declared, bag, cpack, node.Name)
					return
				}
			}
		}
		if strings.Contains(underlying, "(*") {
			bag.Warning("W3001", diag.Empty, "函数指针 typedef `"+node.Name+"` 未映射", "写 C shim", "")
			return
		}
		mapped := MapCType(underlying, target, declared, bag, diag.Empty, cpack)
		if strings.HasPrefix(mapped, "（") {
			return
		}
		aname := TypeName(node.Name)
		declared[node.Name] = aname
		if _, ok := declared[strings.TrimPrefix(underlying, "struct ")]; !ok {
			*out = append(*out, "type "+aname+" = "+mapped+";")
		}
		return
	case "RecordDecl":
		if node.CompleteDef {
			emitRecord(node, out, declared, target, bag, cpack, "")
		}
		return
	case "EnumDecl":
		emitEnum(node, out, declared, bag, cpack, "")
		return
	case "FunctionDecl":
		if len(node.Inner) == 0 && node.Type.QualType == "" {
			return
		}
		emitFunction(node, out, declared, target, bag, cpack)
		return
	case "VarDecl":
		emitVar(node, out, declared, target, bag, cpack)
		return
	}
	for _, child := range node.Inner {
		walkClang(child, out, declared, target, bag, cpack)
	}
}

// qualToDecl 把 `int (const char *, int)` 拆成 (返回类型, [参数类型])。
func qualToDecl(qualt string) (string, []string) {
	qualt = strings.TrimSpace(qualt)
	open := strings.Index(qualt, "(")
	if open < 0 {
		return qualt, nil
	}
	close := strings.LastIndex(qualt, ")")
	if close < open {
		return qualt, nil
	}
	ret := strings.TrimSpace(qualt[:open])
	argsText := qualt[open+1 : close]
	args := []string{}
	depth := 0
	cur := strings.Builder{}
	for i := 0; i < len(argsText); i++ {
		ch := argsText[i]
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
		}
		if ch == ',' && depth == 0 {
			if s := strings.TrimSpace(cur.String()); s != "" {
				args = append(args, s)
			}
			cur.Reset()
			continue
		}
		cur.WriteByte(ch)
	}
	if s := strings.TrimSpace(cur.String()); s != "" {
		args = append(args, s)
	}
	return ret, args
}

func emitRecord(node *clangNode, out *[]string, declared map[string]string, target Target,
	bag *diag.Bag, cpack, forcedName string) {
	name := forcedName
	if name == "" {
		name = node.Name
	}
	if name == "" {
		bag.Warning("W3003", diag.Empty, "无名 struct 未映射（clang 档要求具名）", "请给 C 类型起名", "")
		return
	}
	aname := TypeName(name)
	if prev, ok := declared[name]; ok && prev == aname {
		for _, line := range *out {
			if strings.HasPrefix(line, "struct "+aname+" ") {
				return
			}
		}
	}
	declared[name] = aname
	fields := []string{}
	for _, child := range node.Inner {
		if child.Kind != "FieldDecl" {
			continue
		}
		fname := child.Name
		if fname == "" {
			bag.Warning("W3003", diag.Empty, "`"+aname+"` 有匿名字段，已合成命名", "请检查 C 侧匿名成员", "")
			fname = "anon_" + itoa(len(fields))
		}
		ftype := MapCType(child.Type.QualType, target, declared, bag, diag.Empty, cpack)
		fields = append(fields, "    "+fname+": "+ftype+",")
	}
	*out = append(*out, "struct "+aname+" {\n"+strings.Join(fields, "\n")+"\n};")
	declared[name] = aname
}

func emitEnum(node *clangNode, out *[]string, declared map[string]string, bag *diag.Bag, cpack, forcedName string) {
	name := forcedName
	if name == "" {
		name = node.Name
	}
	if name == "" {
		bag.Warning("W3003", diag.Empty, "无名 enum 未映射（clang 档要求具名）", "请给 C 类型起名", "")
		return
	}
	aname := TypeName(name)
	for _, v := range declared {
		if v == aname {
			return
		}
	}
	declared[name] = aname
	body := []string{}
	for _, child := range node.Inner {
		if child.Kind != "EnumConstantDecl" {
			continue
		}
		vname := ValueName(child.Name)
		line := "    " + vname
		for _, sub := range child.Inner {
			if sub.Kind == "ConstantExpr" && sub.Value != nil {
				line += " = " + valueText(sub.Value)
			}
		}
		body = append(body, line)
	}
	*out = append(*out, "enum "+aname+" {\n"+strings.Join(body, ",\n")+",\n};")
}

func emitFunction(node *clangNode, out *[]string, declared map[string]string, target Target,
	bag *diag.Bag, cpack string) {
	fname := node.Name
	ret, args := qualToDecl(node.Type.QualType)
	var retMapped string
	if strings.HasPrefix(ret, "_Bool") || ret == "bool" {
		retMapped = "Bool"
	} else {
		retMapped = MapCType(ret, target, declared, bag, diag.Empty, cpack)
	}
	parts := []string{}
	anon := 0
	for _, arg := range args {
		if arg == "void" || arg == "" {
			continue
		}
		if strings.Contains(arg, "...") {
			parts = append(parts, "...")
			continue
		}
		t := MapCType(arg, target, declared, bag, diag.Empty, cpack)
		parts = append(parts, "a"+itoa(anon)+": "+t)
		anon++
	}
	retPart := ""
	if !strings.HasPrefix(retMapped, "（") {
		retPart = " -> " + retMapped
	}
	*out = append(*out, "#[c] func "+ValueName(fname)+"("+strings.Join(parts, ", ")+")"+retPart+";")
}

func emitVar(node *clangNode, out *[]string, declared map[string]string, target Target,
	bag *diag.Bag, cpack string) {
	vname := node.Name
	mapped := MapCType(node.Type.QualType, target, declared, bag, diag.Empty, cpack)
	if strings.HasPrefix(mapped, "（") {
		return
	}
	kw := "var"
	for _, w := range strings.Fields(node.Type.QualType) {
		if w == "const" {
			kw = "let"
		}
	}
	*out = append(*out, "#[c] "+kw+" "+ValueName(vname)+": "+mapped+";")
}

// valueText 把 clang 的 ConstantExpr 值渲染成文本。
func valueText(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		if x == float64(int64(x)) {
			return itoa(int(x))
		}
		return strings.TrimRight(strings.TrimRight(formatFloat(x), "0"), ".")
	case bool:
		if x {
			return "1"
		}
		return "0"
	}
	return ""
}
