package manifest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aila-lang/aila/internal/diag"
)

func write(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, `[package]
name = "demo"
version = "0.1.0"
entry = "_.main"

[target]
board = "esp32s3"

[build]
opt = 2
warnings = "strict"

[build.toolchain]
cc = "system:clang>=18"
llvm = "bundled"

[domains.fast]
kind = "internal-ram"
size = "64K"
align = 16
`)
	bag := &diag.Bag{}
	m := Load(dir, bag)
	if m == nil {
		t.Fatalf("清单应加载成功：%v", bag.Render(false))
	}
	if bag.HasErrors() {
		t.Fatalf("不应有错误：%s", bag.Render(false))
	}
	if m.Name != "demo" || m.Opt != 2 || m.Warnings != "strict" {
		t.Fatalf("字段异常：%+v", m)
	}
	if m.CC != "system:clang>=18" || m.LLVM != "bundled" {
		t.Fatalf("toolchain 字段异常：%s / %s", m.CC, m.LLVM)
	}
	if got := m.EffectiveTriple("x86_64-unknown-linux-gnu"); got != "xtensa-esp32s3-elf" {
		t.Fatalf("EffectiveTriple=%s", got)
	}
	if m.PtrSize() != 4 {
		t.Fatalf("PtrSize=%d", m.PtrSize())
	}
	if d, ok := m.Domains["fast"]; !ok || d.Kind != "internal-ram" || d.Align != 16 {
		t.Fatalf("域解析异常：%+v", m.Domains)
	}
}

func TestMissingFields(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "[package]\nname = \"x\"\n")
	bag := &diag.Bag{}
	Load(dir, bag)
	if !bag.HasErrors() {
		t.Fatal("缺 version 应报错")
	}
	found := false
	for _, d := range bag.Items {
		if d.Code == "E7002" {
			found = true
		}
	}
	if !found {
		t.Fatalf("应报 E7002：%v", bag.Items)
	}
}

func TestUnknownBoard(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "[package]\nname = \"x\"\nversion = \"0.1.0\"\n[target]\nboard = \"nope\"\n")
	bag := &diag.Bag{}
	Load(dir, bag)
	codes := map[string]bool{}
	for _, d := range bag.Items {
		codes[d.Code] = true
	}
	if !codes["E7006"] {
		t.Fatalf("应报 E7006：%v", bag.Items)
	}
}

func TestBadOptAndUnknownKey(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "[package]\nname = \"x\"\nversion = \"0.1.0\"\n[build]\nopt = 9\n[tool]\nx = 1\n")
	bag := &diag.Bag{}
	Load(dir, bag)
	codes := map[string]bool{}
	for _, d := range bag.Items {
		codes[d.Code] = true
	}
	if !codes["E7007"] {
		t.Fatalf("应报 E7007：%v", bag.Items)
	}
	if !codes["W7001"] {
		t.Fatalf("应报 W7001：%v", bag.Items)
	}
}

func TestMissingManifest(t *testing.T) {
	dir := t.TempDir()
	bag := &diag.Bag{}
	if m := Load(dir, bag); m != nil {
		t.Fatal("没有清单应返回 nil")
	}
	if !bag.HasErrors() {
		t.Fatal("没有清单应报错")
	}
}
