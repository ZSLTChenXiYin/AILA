// Package diag 是诊断内核：错误码目录（读共享 JSON）、定位渲染、`aila explain` 数据源。
//
// 与 Python 版逐字对齐（docs/specs/0001 §3.3）：错误码唯一事实源是
// share/aila/data/diag_codes.json —— 两套实现读同一份文件（docs/rfcs/0003 §4.2）。
package diag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ailang/aila/internal/home"
)

// 严重级别。
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
	SeverityNote    = "note"
)

// Span 是源码位置；Text 是所在行原文（前端填，缺失时渲染器省略上下文行）。
type Span struct {
	File   string
	Line   int
	Col    int
	Length int
	Text   string
}

// Empty 表示无位置。
var Empty = Span{}

func (s Span) Valid() bool { return s.File != "" }

func (s Span) String() string {
	if !s.Valid() {
		return "<unknown>"
	}
	return fmt.Sprintf("%s:%d:%d", s.File, s.Line, s.Col)
}

// Label 是位置上的一句话。
type Label struct {
	Span    Span
	Message string
}

// Diagnostic 是一条诊断。
type Diagnostic struct {
	Code     string
	Severity string
	Message  string
	Labels   []Label
	Help     string
	Span     Span
}

// CodeEntry 是错误码的解释条目。
type CodeEntry struct {
	Title   string   `json:"title"`
	Explain string   `json:"explain"`
	Fix     []string `json:"fix"`
	Doc     string   `json:"doc"`
}

type codesFile struct {
	Codes map[string]CodeEntry `json:"codes"`
}

var catalog map[string]CodeEntry

// Catalog 返回错误码目录（首次调用时从共享 JSON 加载）。
func Catalog() (map[string]CodeEntry, error) {
	if catalog != nil {
		return catalog, nil
	}
	path := filepath.Join(home.DataDir(), "diag_codes.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读不到错误码表 %s：%w（AILA_HOME 指向解压根目录？）", path, err)
	}
	var f codesFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("错误码表不是合法 JSON：%w", err)
	}
	catalog = f.Codes
	return catalog, nil
}

// Find 查一个错误码。
func Find(code string) (CodeEntry, bool) {
	c, err := Catalog()
	if err != nil {
		return CodeEntry{}, false
	}
	e, ok := c[code]
	return e, ok
}

// FormatExplain 渲染 `aila explain <CODE>` 的人类可读输出（与 Python 版同文）。
func FormatExplain(code string) string {
	entry, ok := Find(code)
	if !ok {
		return "未登记的错误码：" + code + "\n" +
			"用 `aila explain <CODE>` 查询；错误码分段见 docs/specs/0001 §3.4"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s · %s\n\n%s\n\n怎么修：\n", code, entry.Title, entry.Explain))
	for i, step := range entry.Fix {
		b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, step))
	}
	if entry.Doc != "" {
		b.WriteString(fmt.Sprintf("\n设计依据：docs/%s\n", entry.Doc))
	}
	return strings.TrimRight(b.String(), "\n")
}

// Bag 收集一次编译的诊断。
type Bag struct {
	Items []Diagnostic
}

// Error 记一条错误；未登记的错误码视为实现缺陷（契约：错误码唯一事实源）。
func (b *Bag) Error(code string, span Span, message, label, help string) *Diagnostic {
	return b.emit(SeverityError, code, span, message, label, help)
}

// Warning 记一条警告。
func (b *Bag) Warning(code string, span Span, message, label, help string) *Diagnostic {
	return b.emit(SeverityWarning, code, span, message, label, help)
}

func (b *Bag) emit(severity, code string, span Span, message, label, help string) *Diagnostic {
	if _, ok := Find(code); !ok {
		panic("未登记的错误码 " + code + "：请先写入 share/aila/data/diag_codes.json")
	}
	// 去重：同一个错误码 + 同一行 + 同一句消息只登记一次
	// （同一个构造可能被「类型标注」与「字面量」两条路径各查一遍，不该报两遍）
	for i := range b.Items {
		old := b.Items[i]
		if old.Code == code && old.Message == message &&
			old.Span.File == span.File && old.Span.Line == span.Line {
			return &b.Items[i]
		}
	}
	d := Diagnostic{Code: code, Severity: severity, Message: message, Help: help, Span: span}
	if span.Valid() {
		d.Labels = []Label{{Span: span, Message: label}}
	}
	b.Items = append(b.Items, d)
	return &d
}

// HasErrors 是否有 error 级诊断。
func (b *Bag) HasErrors() bool {
	for _, d := range b.Items {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}

// Counts 返回 (errors, warnings)。
func (b *Bag) Counts() (int, int) {
	var e, w int
	for _, d := range b.Items {
		switch d.Severity {
		case SeverityError:
			e++
		case SeverityWarning:
			w++
		}
	}
	return e, w
}

type palette struct {
	red, yellow, blue, green, cyan, bold, reset string
}

func newPalette(color bool) palette {
	if !color {
		return palette{}
	}
	return palette{
		red: "\033[31m", yellow: "\033[33m", blue: "\033[34m",
		green: "\033[32m", cyan: "\033[36m", bold: "\033[1m", reset: "\033[0m",
	}
}

// Render 按文件 / 行号排序后渲染全部诊断（与 Python 版同格式）。
func (b *Bag) Render(color bool) string {
	if len(b.Items) == 0 {
		return ""
	}
	items := make([]Diagnostic, len(b.Items))
	copy(items, b.Items)
	sort.SliceStable(items, func(i, j int) bool {
		pi, pj := 0, 0
		if items[i].Severity != SeverityError {
			pi = 1
		}
		if items[j].Severity != SeverityError {
			pj = 1
		}
		if pi != pj {
			return pi < pj
		}
		si, sj := items[i].Span, items[j].Span
		if si.File != sj.File {
			return si.File < sj.File
		}
		if si.Line != sj.Line {
			return si.Line < sj.Line
		}
		return items[i].Code < items[j].Code
	})
	var blocks []string
	for _, d := range items {
		blocks = append(blocks, renderOne(d, color))
	}
	return strings.Join(blocks, "\n\n")
}

func renderOne(d Diagnostic, color bool) string {
	c := newPalette(color)
	var b strings.Builder
	sev := d.Severity
	if d.Severity == SeverityError {
		sev = c.red + d.Severity + c.reset
	} else {
		sev = c.yellow + d.Severity + c.reset
	}
	b.WriteString(fmt.Sprintf("%s%s[%s]%s: %s", sev, c.bold, d.Code, c.reset, d.Message))

	span := d.Span
	if !span.Valid() && len(d.Labels) > 0 {
		span = d.Labels[0].Span
	}
	if span.Valid() {
		b.WriteString(fmt.Sprintf("\n  %s┌─%s %s", c.blue, c.reset, span.String()))
		if span.Text != "" {
			b.WriteString(fmt.Sprintf("\n  %s│%s", c.blue, c.reset))
			b.WriteString(fmt.Sprintf("\n%s%4d │%s %s", c.blue, span.Line, c.reset, span.Text))
			markers := make([]rune, max(len([]rune(span.Text)), 1))
			for i := range markers {
				markers[i] = ' '
			}
			for _, lb := range d.Labels {
				if lb.Span.File != span.File || lb.Span.Line != span.Line {
					continue
				}
				start := lb.Span.Col - 1
				if start < 0 {
					start = 0
				}
				length := lb.Span.Length
				if length < 1 {
					length = 1
				}
				for i := start; i < start+length && i < len(markers); i++ {
					markers[i] = '^'
				}
			}
			if line := strings.TrimRight(string(markers), " "); line != "" {
				b.WriteString(fmt.Sprintf("\n   %s│%s %s%s%s", c.blue, c.reset, c.green, line, c.reset))
			}
			for _, lb := range d.Labels {
				if lb.Message != "" {
					b.WriteString(fmt.Sprintf("\n   %s│%s %s", c.blue, c.reset, lb.Message))
				}
			}
		} else {
			for _, lb := range d.Labels {
				if lb.Message != "" {
					b.WriteString(fmt.Sprintf("\n   %s│%s %s", c.blue, c.reset, lb.Message))
				}
			}
		}
		b.WriteString(fmt.Sprintf("\n  %s│%s", c.blue, c.reset))
	}
	if d.Help != "" {
		b.WriteString(fmt.Sprintf("\n  %s=%s %s帮助%s: %s", c.cyan, c.reset, c.bold, c.reset, d.Help))
	}
	if entry, ok := Find(d.Code); ok && entry.Doc != "" {
		b.WriteString(fmt.Sprintf("\n  %s=%s %s参见%s: %s", c.cyan, c.reset, c.bold, c.reset, entry.Doc))
	}
	return b.String()
}

// Summary 返回末尾汇总行（与 Python 版同文）。
func (b *Bag) Summary(color bool) string {
	e, w := b.Counts()
	if e == 0 && w == 0 {
		return ""
	}
	var parts []string
	if e > 0 {
		parts = append(parts, fmt.Sprintf("%d 个错误", e))
	}
	if w > 0 {
		parts = append(parts, fmt.Sprintf("%d 个警告", w))
	}
	text := "共 " + strings.Join(parts, "，")
	c := newPalette(color)
	if e > 0 {
		return c.red + text + c.reset
	}
	return c.yellow + text + c.reset
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
