package chip

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeChip 落一份最小 chip.json 供 Load 测试。
func writeChip(t *testing.T, text string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "chip.json"), []byte(text), 0o666); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestClkDiv(t *testing.T) {
	cases := []struct {
		name string
		chip Chip
		want uint32
	}{
		// esp8266：无小数域，80M/115200 = 694（0x2B6），与原手写 C 常量一致
		{"esp8266", Chip{UART: struct {
			Base           string `json:"base"`
			ClkHz          uint64 `json:"clk_hz"`
			Baud           uint64 `json:"baud"`
			ClkdivFracBits uint   `json:"clkdiv_frac_bits"`
			TXCntShift     uint   `json:"txcnt_shift"`
			TXCntBits      uint   `json:"txcnt_bits"`
			RXCntShift     uint   `json:"rxcnt_shift"`
			RXCntBits      uint   `json:"rxcnt_bits"`
		}{ClkHz: 80000000, Baud: 115200}}, 0x2B6},
		// esp32：div16=11111 → 整数 694 + 小数 7 进 FRAG[23:20]
		{"esp32", Chip{UART: struct {
			Base           string `json:"base"`
			ClkHz          uint64 `json:"clk_hz"`
			Baud           uint64 `json:"baud"`
			ClkdivFracBits uint   `json:"clkdiv_frac_bits"`
			TXCntShift     uint   `json:"txcnt_shift"`
			TXCntBits      uint   `json:"txcnt_bits"`
			RXCntShift     uint   `json:"rxcnt_shift"`
			RXCntBits      uint   `json:"rxcnt_bits"`
		}{ClkHz: 80000000, Baud: 115200, ClkdivFracBits: 4}}, 0x7002B6},
		// s3：复位默认 XTAL 40MHz，div16=5555 → 347 + 3
		{"esp32s3", Chip{UART: struct {
			Base           string `json:"base"`
			ClkHz          uint64 `json:"clk_hz"`
			Baud           uint64 `json:"baud"`
			ClkdivFracBits uint   `json:"clkdiv_frac_bits"`
			TXCntShift     uint   `json:"txcnt_shift"`
			TXCntBits      uint   `json:"txcnt_bits"`
			RXCntShift     uint   `json:"rxcnt_shift"`
			RXCntBits      uint   `json:"rxcnt_bits"`
		}{ClkHz: 40000000, Baud: 115200, ClkdivFracBits: 4}}, 0x30015B},
	}
	for _, tc := range cases {
		got, err := tc.chip.ClkDiv()
		if err != nil {
			t.Fatalf("%s：%v", tc.name, err)
		}
		if got != tc.want {
			t.Fatalf("%s：CLKDIV got 0x%X want 0x%X", tc.name, got, tc.want)
		}
	}
}

func TestTXCntMask(t *testing.T) {
	c := Chip{}
	c.UART.TXCntBits = 8
	if m := c.TXCntMask(); m != 0xFF {
		t.Fatalf("8 位掩码错误：0x%X", m)
	}
	c.UART.TXCntBits = 10
	if m := c.TXCntMask(); m != 0x3FF {
		t.Fatalf("10 位掩码错误：0x%X", m)
	}
	c.UART.RXCntBits = 10
	if m := c.RXCntMask(); m != 0x3FF {
		t.Fatalf("RX 掩码错误：0x%X", m)
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(t.TempDir()); err == nil {
		t.Fatal("缺 chip.json 应当报错")
	}
}

func TestLoadAndRender(t *testing.T) {
	dir := writeChip(t, `{
		"family": "testchip",
		"memory": {"iram_origin": "0x40000000", "iram_len": "0x8000", "dram_origin": "0x3FF00000", "dram_len": "0x8000", "stack_size": "0x1000", "entry": "call_start"},
		"image": {"kind": "esp8266", "mode_code": 2, "size_code": 4, "freq_code": 0},
		"uart": {"base": "0x60000000", "clk_hz": 80000000, "baud": 115200, "clkdiv_frac_bits": 0, "txcnt_shift": 26, "txcnt_bits": 6}
	}`)
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	ld := c.LinkerScript()
	for _, want := range []string{
		`ENTRY(call_start)`,
		`ORIGIN = 0x40000000, LENGTH = 0x8000`,
		`_stack_top = ORIGIN(dram) + LENGTH(dram);`,
		`*(.literal .literal.*)`,
	} {
		if !strings.Contains(ld, want) {
			t.Fatalf("链接脚本缺 %q\n%s", want, ld)
		}
	}
	cfg, err := c.ChipCfgH()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cfg, "UART_CLKDIV_VAL     0x2B6u") {
		t.Fatalf("chip_cfg.h 分频值错误：\n%s", cfg)
	}
	if !strings.Contains(cfg, "UART_RXCNT_MASK     0x0u") {
		t.Fatalf("chip_cfg.h RXCNT 掩码默认错误：\n%s", cfg)
	}
	if _, err := c.BootCfgH(); err == nil {
		t.Fatal("无 boot 段时 BootCfgH 应当报错")
	}
	if _, err := c.BootLinkerScript(); err == nil {
		t.Fatal("无 boot 段时 BootLinkerScript 应当报错")
	}
}

func TestBootCfgH(t *testing.T) {
	dir := writeChip(t, `{
		"family": "testchip",
		"memory": {"entry": "call_start", "stack_size": "0x1000"},
		"image": {"kind": "esp32"},
		"boot": {"app_offset": "0x10000", "boot_offset": "0x1000",
			"rom_spiflash_read": "0x40062ED8",
			"wdt": {"rtc_base": "0x3FF48000", "wdtkey_off": "0xA4", "feed_off": "0xA0", "config0_off": "0x8C", "key": "0x50D83AA1"}}
	}`)
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	h, err := c.BootCfgH()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"BOOT_APP_OFFSET  0x10000", "WDT_WKEY_VALUE   0x50D83AA1", "RTC_CNTL_BASE    0x3FF48000"} {
		if !strings.Contains(h, want) {
			t.Fatalf("boot_cfg.h 缺 %q：\n%s", want, h)
		}
	}
	ld, err := c.BootLinkerScript()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ld, "PROVIDE(esp_rom_spiflash_read = 0x40062ED8);") {
		t.Fatalf("boot.ld 缺 ROM 符号：\n%s", ld)
	}
}
