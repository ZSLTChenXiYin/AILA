package lexer

import (
	"strconv"
	"strings"
)

func itoa(n int) string { return strconv.Itoa(n) }

func parseIntBase(body string, base int) (int64, bool) {
	// 全程 ParseUint：0..2^64-1 都能装下；≥ 2^63 的值以**位型**存 int64
	// （0xFFFFFFFFFFFFFFFF → -1），落点检查由 sema 按目标类型把关。
	if u, err := strconv.ParseUint(body, base, 64); err == nil {
		return int64(u), true
	}
	// 超出 64 位：AILA 没有 >64 位整数，报错不静默
	return 0, false
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
