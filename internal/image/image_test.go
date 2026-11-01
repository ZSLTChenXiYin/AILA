package image

import (
	"crypto/sha256"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// fakeELF 手工拼一个最小 ELF32 LSB：ELF 头 + 一个 PT_LOAD 程序段。
// 只用于镜像格式测试，不做真实架构校验（打包只看 program header）。
func fakeELF(t *testing.T, dir string, entry uint32, segAddr uint32, segData []byte) string {
	t.Helper()
	phoff := uint32(52) // ELF32 头大小
	var ph [32]byte
	binary.LittleEndian.PutUint32(ph[0:], 1)                     // p_type = PT_LOAD
	binary.LittleEndian.PutUint32(ph[4:], phoff+32)              // p_offset：段数据紧跟在 header+phdr 之后
	binary.LittleEndian.PutUint32(ph[8:], segAddr)               // p_vaddr
	binary.LittleEndian.PutUint32(ph[12:], segAddr)              // p_paddr
	binary.LittleEndian.PutUint32(ph[16:], uint32(len(segData))) // p_filesz
	binary.LittleEndian.PutUint32(ph[20:], uint32(len(segData))) // p_memsz

	var eh [52]byte
	copy(eh[0:], "\x7fELF")
	eh[4] = 1                                     // ELFCLASS32
	eh[5] = 1                                     // ELFDATA2LSB
	eh[6] = 1                                     // EV_CURRENT
	binary.LittleEndian.PutUint16(eh[16:], 2)     // e_type = EXEC
	binary.LittleEndian.PutUint16(eh[18:], 94)    // e_machine = EM_XTENSA
	binary.LittleEndian.PutUint32(eh[24:], entry) // e_entry
	binary.LittleEndian.PutUint32(eh[28:], phoff) // e_phoff
	binary.LittleEndian.PutUint32(eh[42:], 32)    // e_phentsize
	binary.LittleEndian.PutUint16(eh[44:], 1)     // e_phnum

	raw := append(eh[:], ph[:]...)
	raw = append(raw, segData...)
	path := filepath.Join(dir, "mini.elf")
	if err := os.WriteFile(path, raw, 0o666); err != nil {
		t.Fatal(err)
	}
	return path
}

// pad8 是镜像打包的段长对齐（见 appendSegment：ROM loader 要求 8 字节对齐）
func pad8(n int) int { return (n + 7) / 8 * 8 }

func TestESP8266Layout(t *testing.T) {
	dir := t.TempDir()
	seg := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02}
	elfPath := fakeELF(t, dir, 0x4010000C, 0x40100000, seg)
	out := filepath.Join(dir, "app.bin")

	if err := ESP8266(elfPath, out, DefaultFlash()); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	// 头：magic / 段数 / mode(dio=2) / size(4M=4)<<4|freq(40m=0) / 入口
	if raw[0] != 0xE9 || raw[1] != 1 || raw[2] != 2 || raw[3] != 0x40 {
		t.Fatalf("镜像头错误：% X", raw[:4])
	}
	if got := binary.LittleEndian.Uint32(raw[4:8]); got != 0x4010000C {
		t.Fatalf("入口错误：0x%X", got)
	}
	// 段头：地址 + 长度（esptool bin_image.py / SDK gen_appbin.py 均为此序，与直觉相反）
	if got := binary.LittleEndian.Uint32(raw[8:12]); got != 0x40100000 {
		t.Fatalf("段地址错误：0x%X", got)
	}
	// 段长补到 8 字节对齐：ROM loader 按字搬运，未对齐时会把下一个段头也划进校验
	if got := binary.LittleEndian.Uint32(raw[12:16]); got != uint32(pad8(len(seg))) {
		t.Fatalf("段长错误：%d", got)
	}
	// 段数据
	if string(raw[16:16+len(seg)]) != string(seg) {
		t.Fatalf("段数据错误：% X", raw[16:16+len(seg)])
	}
	// 校验和占最后一字节且镜像长度 16 对齐
	if len(raw)%16 != 0 {
		t.Fatalf("镜像长度 %d 不是 16 的倍数", len(raw))
	}
	sum := byte(0xEF)
	for _, b := range seg {
		sum ^= b
	}
	if raw[len(raw)-1] != sum {
		t.Fatalf("校验和错误：got 0x%02X want 0x%02X", raw[len(raw)-1], sum)
	}
}

func TestESP8266RejectsBadFlash(t *testing.T) {
	dir := t.TempDir()
	elfPath := fakeELF(t, dir, 0x40100000, 0x40100000, []byte{1})
	if err := ESP8266(elfPath, filepath.Join(dir, "x.bin"), Flash{Mode: "warp", Size: "4M", Freq: "40m"}); err == nil {
		t.Fatal("未知 flash 模式应当报错")
	}
}

func TestESP32Layout(t *testing.T) {
	dir := t.TempDir()
	seg := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02}
	elfPath := fakeELF(t, dir, 0x400800F0, 0x40080000, seg)
	out := filepath.Join(dir, "app.bin")

	if err := PackESP32(elfPath, out, 0, 2, 2<<4|0); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	// 主头 8B：magic / 段数 / dio=2 / 4M(esp32 size 表=2)<<4|40m / 入口
	if raw[0] != 0xE9 || raw[1] != 1 || raw[2] != 2 || raw[3] != 0x20 {
		t.Fatalf("主头错误：% X", raw[:4])
	}
	if got := binary.LittleEndian.Uint32(raw[4:8]); got != 0x400800F0 {
		t.Fatalf("入口错误：0x%X", got)
	}
	// 扩展头 16B：wp_pin=0xEE、spi_pin_drv×3=0、chip_id、hash_appended=1
	if raw[8] != 0xEE {
		t.Fatalf("wp_pin 错误：0x%02X", raw[8])
	}
	if got := binary.LittleEndian.Uint16(raw[12:14]); got != 0 {
		t.Fatalf("chip_id 错误：%d", got)
	}
	if raw[23] != 1 {
		t.Fatalf("hash_appended 错误：0x%02X", raw[23])
	}
	// 段表与数据（头共 24B；段头是 地址+长度，同上）
	if got := binary.LittleEndian.Uint32(raw[24:28]); got != 0x40080000 {
		t.Fatalf("段地址错误：0x%X", got)
	}
	if got := binary.LittleEndian.Uint32(raw[28:32]); got != uint32(pad8(len(seg))) {
		t.Fatalf("段长错误：%d", got)
	}
	if string(raw[32:32+len(seg)]) != string(seg) {
		t.Fatalf("段数据错误：% X", raw[32:32+len(seg)])
	}
	// 校验和占对齐末字节，其后跟 32B SHA256
	if len(raw)%16 != 0 {
		t.Fatalf("镜像长度 %d 不是 16 的倍数", len(raw))
	}
	sum := byte(0xEF)
	for _, b := range seg {
		sum ^= b
	}
	if raw[len(raw)-33] != sum {
		t.Fatalf("校验和错误：got 0x%02X want 0x%02X", raw[len(raw)-33], sum)
	}
	want := sha256.Sum256(raw[:len(raw)-32])
	if string(raw[len(raw)-32:]) != string(want[:]) {
		t.Fatal("SHA256 与镜像内容不符")
	}
}

func TestESP32S3ChipID(t *testing.T) {
	dir := t.TempDir()
	elfPath := fakeELF(t, dir, 0x40374000, 0x40374000, []byte{0x42})
	out := filepath.Join(dir, "s3.bin")
	if err := PackESP32(elfPath, out, 9, 2, 2<<4|0); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if got := binary.LittleEndian.Uint16(raw[12:14]); got != 9 {
		t.Fatalf("esp32s3 的 chip_id 应为 9，got %d", got)
	}
}
