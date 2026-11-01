package targets

import "testing"

func TestParseEmbeddedTriples(t *testing.T) {
	cases := []struct {
		triple   string
		arch     string
		cpu      string
		embedded bool
	}{
		{"xtensa-esp8266-elf", "xtensa", "esp8266", true},
		{"xtensa-esp32s3-elf", "xtensa", "esp32s3", true},
		{"riscv32-unknown-elf", "riscv32", "", true},
		{"aarch64-unknown-none", "aarch64", "", true},
		{"arm-none-eabi", "arm", "", true},
		{"x86_64-w64-windows-gnu", "x86_64", "", false},
		{"x86_64-unknown-linux-gnu", "x86_64", "", false},
		{"aarch64-apple-darwin", "aarch64", "", false},
	}
	for _, c := range cases {
		got := Parse(c.triple)
		if got.Arch != c.arch || got.CPUDerived() != c.cpu || got.Embedded != c.embedded {
			t.Fatalf("%s → arch=%s cpu=%s embedded=%v，期望 %s/%s/%v",
				c.triple, got.Arch, got.CPUDerived(), got.Embedded, c.arch, c.cpu, c.embedded)
		}
	}
}

func TestValidTriple(t *testing.T) {
	for _, ok := range []string{"xtensa-esp8266-elf", "x86_64-w64-windows-gnu", "arm-none-eabi"} {
		if !Valid(ok) {
			t.Fatalf("%s 应合法", ok)
		}
	}
	for _, bad := range []string{"", "host", "esp8266", "xtensa--elf", "xtensa-esp8266-"} {
		if Valid(bad) {
			t.Fatalf("%q 应不合法", bad)
		}
	}
}

func TestVersionBandAndKinds(t *testing.T) {
	lo, hi := VersionBand()
	if lo == "" || hi == "" {
		t.Fatal("版本带为空")
	}
	if len(DomainKinds()) == 0 {
		t.Fatal("domain kind 为空")
	}
}
