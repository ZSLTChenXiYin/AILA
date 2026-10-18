package ast

import (
	"bytes"
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"

	"aila/internal/diag"
)

// 与参考实现完全一致的序列化规则（现由 tools/goldengen 生成快照）：
//   - 空值（nil / "" / 空列表）不输出；
//   - 节点写成 (Kind field=value …)，字段顺序 = 声明顺序；
//   - 列表写成 [a b c]；二元组写成 [a b]；
//   - 字符串能裸写就裸写，否则 JSON 转义。
var bareRe = regexp.MustCompile(`^[A-Za-z0-9_.:@\[\]/+-]+$`)

// Sexp 把语法树渲染成契约快照用的 S-表达式。
func Sexp(v any) string {
	var b bytes.Buffer
	writeSexp(&b, v)
	return b.String()
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	switch x := v.(type) {
	case string:
		return x == ""
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Func, reflect.Chan:
		return rv.IsNil()
	case reflect.Slice, reflect.Array:
		return rv.Len() == 0
	}
	return false
}

// optNode 把「可能为 nil 的指针/接口」变成序列化器能识别的 nil。
func optNode(v any) any {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Func, reflect.Chan:
		if rv.IsNil() {
			return nil
		}
	case reflect.Slice:
		if rv.Len() == 0 {
			return nil
		}
	}
	return v
}

func writeSexp(b *bytes.Buffer, v any) {
	if v == nil {
		b.WriteString("nil")
		return
	}
	switch x := v.(type) {
	case Node:
		b.WriteString("(")
		b.WriteString(x.Kind())
		for _, f := range x.Fields() {
			if isEmpty(f.Value) {
				continue
			}
			b.WriteString(" ")
			b.WriteString(f.Name)
			b.WriteString("=")
			writeSexp(b, f.Value)
		}
		b.WriteString(")")
		return
	case Pair:
		b.WriteString("[")
		writeSexp(b, x.A)
		b.WriteString(" ")
		writeSexp(b, x.B)
		b.WriteString("]")
		return
	case bool:
		if x {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
		return
	case int64:
		b.WriteString(strconv.FormatInt(x, 10))
		return
	case int:
		b.WriteString(strconv.Itoa(x))
		return
	case float64:
		b.WriteString(formatFloat(x))
		return
	case string:
		b.WriteString(atomString(x))
		return
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		b.WriteString("[")
		for i := 0; i < rv.Len(); i++ {
			if i > 0 {
				b.WriteString(" ")
			}
			writeSexp(b, rv.Index(i).Interface())
		}
		b.WriteString("]")
		return
	}
	b.WriteString(atomString(reflect.ValueOf(v).String()))
}

// formatFloat 贴近 Python 的 repr(float)：整数浮点也带 .0。
func formatFloat(v float64) string {
	s := strconv.FormatFloat(v, 'g', -1, 64)
	for _, c := range s {
		if c == '.' || c == 'e' || c == 'E' || c == 'n' || c == 'i' {
			return s
		}
	}
	return s + ".0"
}

func atomString(s string) string {
	if s != "" && bareRe.MatchString(s) {
		return s
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	out := buf.String()
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return out
}

// --------------------------------------------------------------------------- //
// 列表辅助：空列表返回 nil，序列化时整段省略（与 Python 的 skip 规则一致）
// --------------------------------------------------------------------------- //

func fields(v []*Field) any {
	if len(v) == 0 {
		return nil
	}
	out := make([]any, 0, len(v))
	for _, x := range v {
		out = append(out, x)
	}
	return out
}

func variants(v []*Variant) any {
	if len(v) == 0 {
		return nil
	}
	out := make([]any, 0, len(v))
	for _, x := range v {
		out = append(out, x)
	}
	return out
}

func params(v []*Param) any {
	if len(v) == 0 {
		return nil
	}
	out := make([]any, 0, len(v))
	for _, x := range v {
		out = append(out, x)
	}
	return out
}

func effects(v []*Effect) any {
	if len(v) == 0 {
		return nil
	}
	out := make([]any, 0, len(v))
	for _, x := range v {
		out = append(out, x)
	}
	return out
}

func imports(v []*Import) any {
	if len(v) == 0 {
		return nil
	}
	out := make([]any, 0, len(v))
	for _, x := range v {
		out = append(out, x)
	}
	return out
}

func refs(v []*TypeRef) any {
	if len(v) == 0 {
		return nil
	}
	out := make([]any, 0, len(v))
	for _, x := range v {
		out = append(out, x)
	}
	return out
}

func fieldPats(v []*FieldPat) any {
	if len(v) == 0 {
		return nil
	}
	out := make([]any, 0, len(v))
	for _, x := range v {
		out = append(out, x)
	}
	return out
}

func arms(v []*Arm) any {
	if len(v) == 0 {
		return nil
	}
	out := make([]any, 0, len(v))
	for _, x := range v {
		out = append(out, x)
	}
	return out
}

func pairsAny(v []Pair) any {
	if len(v) == 0 {
		return nil
	}
	out := make([]any, 0, len(v))
	for _, x := range v {
		out = append(out, x)
	}
	return out
}

func attrsFields(a Attrs) []F {
	var attrs any
	if len(a.Attrs) > 0 {
		list := make([]any, 0, len(a.Attrs))
		for _, x := range a.Attrs {
			list = append(list, x)
		}
		attrs = list
	}
	return []F{N("attrs", attrs), N("is_public", a.IsPublic)}
}

// SpanOf 取节点的源码位置（所有节点都有 Span 字段，接口里只暴露 Kind/Fields）。
func SpanOf(n Node) diag.Span {
	if n == nil {
		return diag.Empty
	}
	rv := reflect.ValueOf(n)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return diag.Empty
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return diag.Empty
	}
	f := rv.FieldByName("Span")
	if !f.IsValid() || !f.CanInterface() {
		return diag.Empty
	}
	if sp, ok := f.Interface().(diag.Span); ok {
		return sp
	}
	return diag.Empty
}