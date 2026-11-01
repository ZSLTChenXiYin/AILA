// Package targets 提供目标三元组解析与共享配置（docs/rfcs/0006 §6）。
//
// board 表已退场：清单 [target] 以 triple 为主键，arch 从第一段拆出、
// cpu 从第二段推导、embedded 看后缀。数据源 share/aila/data/compiler.json
// 只留全局配置（版本带、域 kind 枚举）。
package targets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"aila/internal/home"
)

// Target 是解析后的目标三元组。
type Target struct {
	Triple string
	Arch   string // 第一段（xtensa / x86_64 / riscv32 / ...）
	CPU    string // 清单显式写的 -mcpu 值（可空；空时用 CPUDerived 推导）
	// Embedded 是 freestanding 判定：后缀含 -elf/-none/-eabi。
	// 嵌入式目标由 build 走平台 runtime + 交叉链接 + 镜像打包，宿主目标链接出可执行文件。
	Embedded bool
}

// cpuExempt 是不能当 -mcpu 用的 triple 第二段（平台词，不是芯片名）。
var cpuExempt = map[string]bool{
	"unknown": true, "none": true, "generic": true,
	"w64": true, "pc": true, "apple": true, "linux": true,
}

// Parse 拆解一个三元组。
func Parse(triple string) Target {
	t := Target{Triple: triple}
	if triple == "" {
		return t
	}
	parts := strings.Split(triple, "-")
	t.Arch = parts[0]
	for _, p := range parts[1:] {
		if p == "elf" || p == "none" || p == "eabi" {
			t.Embedded = true
		}
	}
	return t
}

// Valid 校验 triple 形态：至少两段、段非空。
func Valid(triple string) bool {
	if triple == "" {
		return false
	}
	parts := strings.Split(triple, "-")
	if len(parts) < 2 {
		return false
	}
	for _, p := range parts {
		if strings.TrimSpace(p) == "" {
			return false
		}
	}
	return true
}

// CPUDerived 返回 triple 第二段作为 -mcpu 候选（esp 系芯片名即 cpu：
// xtensa-esp8266-elf → esp8266）；第二段缺失或属平台豁免词时返回空串。
func (t Target) CPUDerived() string {
	parts := strings.Split(t.Triple, "-")
	if len(parts) < 2 || cpuExempt[parts[1]] {
		return ""
	}
	return parts[1]
}

// chipGCC 是芯片名 → 官方 GCC 工具链 triple（toolchains/gcc/ 目录名）。
// 厂商 GCC 的 triple 与 AILA 项目 triple 不总是一致（esp8266 ↔ lx106）；
// 这是芯片家族知识，每家族一行数据：新家族加一行即可，不改框架代码。
var chipGCC = map[string]string{
	"esp8266": "xtensa-lx106-elf",
	"esp32":   "xtensa-esp32-elf",
	"esp32s3": "xtensa-esp32s3-elf",
}

// GCCTriple 返回能服务本目标的 GCC triple：芯片别名优先，否则 triple 原样。
func (t Target) GCCTriple() string {
	if cpu := t.CPUDerived(); cpu != "" {
		if g, ok := chipGCC[cpu]; ok {
			return g
		}
	}
	return t.Triple
}

// RuntimeFamily 返回平台 runtime 家族名（share/aila/runtime/<家族>/），
// 即 triple 第二段的芯片名；宿主目标（无芯片名）返回空串。
func (t Target) RuntimeFamily() string { return t.CPUDerived() }

type payload struct {
	VersionBand []string `json:"version_band"`
	DomainKinds []string `json:"domain_kinds"`
}

var cached *payload

func load() (*payload, error) {
	if cached != nil {
		return cached, nil
	}
	path := filepath.Join(home.DataDir(), "compiler.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	cached = &p
	return cached, nil
}

// VersionBand 返回 AILA 绑定的 LLVM 版本带（含端点）。
func VersionBand() (string, string) {
	p, err := load()
	if err != nil || len(p.VersionBand) < 2 {
		return "18", "23"
	}
	return p.VersionBand[0], p.VersionBand[1]
}

// DomainKinds 返回合法物理域 kind。
func DomainKinds() []string {
	p, err := load()
	if err != nil {
		return []string{"internal-ram", "external-ram", "rtc-slow", "default"}
	}
	return p.DomainKinds
}
