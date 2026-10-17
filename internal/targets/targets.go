// Package targets 提供内置 board 表与 domain kind 枚举。
//
// 数据源与 Python 版共用：share/aila/data/targets.json（docs/rfcs/0003 §4.2）。
package targets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/aila-lang/aila/internal/home"
)

// Board 是一个内置目标。
type Board struct {
	Arch     string            `json:"arch"`
	Triple   *string           `json:"triple"`
	Embedded bool              `json:"embedded"`
	Domains  map[string]string `json:"domains"`
}

type payload struct {
	VersionBand []string         `json:"version_band"`
	DomainKinds []string         `json:"domain_kinds"`
	Boards      map[string]Board `json:"boards"`
}

var cached *payload

func load() (*payload, error) {
	if cached != nil {
		return cached, nil
	}
	path := filepath.Join(home.DataDir(), "targets.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读不到 target 表 %s：%w", path, err)
	}
	var p payload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("target 表不是合法 JSON：%w", err)
	}
	cached = &p
	return cached, nil
}

// Boards 返回全部内置 board。
func Boards() (map[string]Board, error) {
	p, err := load()
	if err != nil {
		return nil, err
	}
	return p.Boards, nil
}

// Names 返回按名字排序的 board 列表。
func Names() ([]string, error) {
	boards, err := Boards()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(boards))
	for name := range boards {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// Find 查一个 board。
func Find(name string) (Board, bool) {
	boards, err := Boards()
	if err != nil {
		return Board{}, false
	}
	b, ok := boards[name]
	return b, ok
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
