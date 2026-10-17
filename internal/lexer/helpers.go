package lexer

import (
	"strconv"
	"strings"
)

func itoa(n int) string { return strconv.Itoa(n) }

func parseIntBase(body string, base int) int64 {
	v, err := strconv.ParseInt(body, base, 64)
	if err != nil {
		// 超长字面量：退回浮点再截断，保证不 panic（sema 会做范围检查）
		if f, ferr := strconv.ParseFloat(body, 64); ferr == nil {
			return int64(f)
		}
		return 0
	}
	return v
}

func parseFloat(clean string) float64 {
	v, err := strconv.ParseFloat(strings.ReplaceAll(clean, "_", ""), 64)
	if err != nil {
		return 0
	}
	return v
}

func parseUnicode(hex []rune) (rune, bool) {
	s := string(hex)
	v, err := strconv.ParseInt(s, 16, 32)
	if err != nil || v < 0 || v > 0x10FFFF {
		return 0, false
	}
	return rune(v), true
}
