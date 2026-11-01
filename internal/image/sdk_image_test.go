package image

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// fakeSDKELF 拼一个多段 ELF32 LSB：ELF 头 + 若干 PT_LOAD（顺序即 program header 顺序）。
// SDK 镜像要同时有 irom 段（虚拟地址在 0x40200000 以上）与 RAM 段，单段的 fakeELF 不够用。
func fakeSDKELF(t *testing.T, dir string, entry uint32, segs []seg) string {
	t.Helper()
	const ehSize, phSize = 52, 32
	phoff := uint32(ehSize)
	dataOff := uint32(ehSize + phSize*len(segs))

	var eh [ehSize]byte
	copy(eh[0:], "\x7fELF")
	eh[4] = 1                                     // ELFCLASS32
	eh[5] = 1                                     // ELFDATA2LSB
	eh[6] = 1                                     // EV_CURRENT
	binary.LittleEndian.PutUint16(eh[16:], 2)     // e_type = EXEC
	binary.LittleEndian.PutUint16(eh[18:], 94)    // e_machine = EM_XTENSA
	binary.LittleEndian.PutUint32(eh[24:], entry) // e_entry
	binary.LittleEndian.PutUint32(eh[28:], phoff) // e_phoff
	binary.LittleEndian.PutUint32(eh[42:], phSize)
	binary.LittleEndian.PutUint16(eh[44:], uint16(len(segs)))

	raw := append([]byte{}, eh[:]...)
	off := dataOff
	for _, s := range segs {
		var ph [phSize]byte
		binary.LittleEndian.PutUint32(ph[0:], 1) // PT_LOAD
		binary.LittleEndian.PutUint32(ph[4:], off)
		binary.LittleEndian.PutUint32(ph[8:], s.addr)  // p_vaddr
		binary.LittleEndian.PutUint32(ph[12:], s.addr) // p_paddr
		binary.LittleEndian.PutUint32(ph[16:], uint32(len(s.data)))
		binary.LittleEndian.PutUint32(ph[20:], uint32(len(s.data)))
		raw = append(raw, ph[:]...)
		off += uint32(len(s.data))
	}
	for _, s := range segs {
		raw = append(raw, s.data...)
	}

	path := filepath.Join(dir, "sdk.elf")
	if err := os.WriteFile(path, raw, 0o666); err != nil {
		t.Fatal(err)
	}
	return path
}

// Test8266SDKFotaLayout 守着 FOTA v2 布局：irom 紧跟 16B 头、槽位对齐、CRC32 尾。
// 这套字节序是上板能否启动的硬约束（designs/0005 §6.C）。
func Test8266SDKFotaLayout(t *testing.T) {
	const (
		entry  = uint32(0x40101170)
		origin = uint32(0x40201010) // user1 槽 flash 0x1000 + 0x10
		slot   = uint32(0x1000)
	)
	irom := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66} // 6B → 补到 16B
	iram := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	dram := []byte{0x01, 0x02, 0x03, 0x04}

	dir := t.TempDir()
	elf := fakeSDKELF(t, dir, entry, []seg{
		{addr: 0x3ffe8000, data: dram},
		{addr: 0x40100000, data: iram},
		{addr: origin, data: irom},
	})
	out := filepath.Join(dir, "app.bin")
	if err := Pack8266SDK(elf, out, 3, 4<<4, origin, slot); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	// 主头：0xEA + 固定 0x04 + mode + size/freq + 入口
	if raw[0] != 0xEA || raw[1] != imageV2Segment || raw[2] != 3 || raw[3] != 0x40 {
		t.Fatalf("主头错误：% X", raw[:4])
	}
	if got := binary.LittleEndian.Uint32(raw[4:8]); got != entry {
		t.Fatalf("入口错误：0x%X", got)
	}

	// irom 段：加载地址固定写 0，长度含 16B 对齐填充，数据在 0x10 起
	if got := binary.LittleEndian.Uint32(raw[8:12]); got != 0 {
		t.Fatalf("irom 段头加载地址必须写 0，got 0x%X", got)
	}
	iromLen := binary.LittleEndian.Uint32(raw[12:16])
	if iromLen != 16 {
		t.Fatalf("irom 段长应补齐到 16，got %d", iromLen)
	}
	if string(raw[16:16+len(irom)]) != string(irom) {
		t.Fatalf("irom 数据错误：% X", raw[16:16+len(irom)])
	}

	// 次头：0xE9 + RAM 段数（irom 不算）
	at := 16 + int(iromLen)
	if raw[at] != 0xE9 || raw[at+1] != 2 {
		t.Fatalf("次头错误：% X", raw[at:at+4])
	}
	if got := binary.LittleEndian.Uint32(raw[at+4:]); got != entry {
		t.Fatalf("次头入口错误：0x%X", got)
	}

	// RAM 段按 ELF program header 顺序（DRAM 先、IRAM 后），每段 {addr,len,data}
	checksum := byte(0xEF)
	off := at + 8
	for _, want := range []seg{{addr: 0x3ffe8000, data: dram}, {addr: 0x40100000, data: iram}} {
		if got := binary.LittleEndian.Uint32(raw[off:]); got != want.addr {
			t.Fatalf("段地址错误：0x%X want 0x%X", got, want.addr)
		}
		n := binary.LittleEndian.Uint32(raw[off+4:])
		if int(n) != len(want.data) {
			t.Fatalf("段长错误：%d want %d", n, len(want.data))
		}
		for _, b := range raw[off+8 : off+8+int(n)] {
			checksum ^= b
		}
		off += 8 + int(n)
	}

	// 校验和占 16B 块的最后一字节；其后 4B 是 SDK bootloader 校验的 CRC32
	for off%16 != 15 {
		if raw[off] != 0x00 {
			t.Fatalf("校验和前的对齐填充应为 0x00，got 0x%02X", raw[off])
		}
		off++
	}
	if raw[off] != checksum {
		t.Fatalf("校验和错误：got 0x%02X want 0x%02X", raw[off], checksum)
	}
	if len(raw) != off+1+4 {
		t.Fatalf("镜像长度应为 校验和位置+1+4，got %d（校验和在 %d）", len(raw), off)
	}
	if got, want := binary.LittleEndian.Uint32(raw[off+1:]), esp8266CRC32(raw[:off+1]); got != want {
		t.Fatalf("CRC32 尾错误：got 0x%08X want 0x%08X", got, want)
	}
}

// Test8266SDKOriginMustMatchSlot 守住「链接地址与槽位对齐」这条硬约束：
// cache 是 flash 0x0 ↔ 0x40200000 的固定线性映射，irom 落在槽位 + 0x10 处。
// 非 FOTA 老布局（irom@0x40210000 = flash 0x10000）用 v2 镜像根本表达不出来——
// 这正是之前上板必然 Fatal exception(0) 的根因。
func Test8266SDKOriginMustMatchSlot(t *testing.T) {
	dir := t.TempDir()
	oldLayout := fakeSDKELF(t, dir, 0x40100000, []seg{
		{addr: 0x40100000, data: []byte{0, 0, 0, 0}},
		{addr: 0x40210000, data: []byte{1, 2, 3, 4}}, // 老布局（flash 0x10000）
	})
	if err := Pack8266SDK(oldLayout, filepath.Join(dir, "x.bin"), 3, 0x40, 0x40210000, 0x1000); err == nil {
		t.Fatal("irom 链接地址与槽位不符时应当报错")
	}

	// 与槽位对齐的布局（烧 flash 0x0，irom 落在 0x10）应当通过
	aligned := fakeSDKELF(t, dir, 0x40100000, []seg{
		{addr: 0x40100000, data: []byte{0, 0, 0, 0}},
		{addr: 0x40200010, data: []byte{1, 2, 3, 4}},
	})
	if err := Pack8266SDK(aligned, filepath.Join(dir, "y.bin"), 3, 0x40, 0x40200010, 0); err != nil {
		t.Fatalf("槽位对齐时不应报错：%v", err)
	}
}

// Test8266SDKNeedsIromSegment 没有 flash 段的镜像（纯裸机布局）不能当 SDK 镜像打包。
func Test8266SDKNeedsIromSegment(t *testing.T) {
	dir := t.TempDir()
	elf := fakeSDKELF(t, dir, 0x40100000, []seg{{addr: 0x40100000, data: []byte{0, 0, 0, 0}}})
	if err := Pack8266SDK(elf, filepath.Join(dir, "x.bin"), 3, 0x40, 0x40201010, 0x1000); err == nil {
		t.Fatal("缺 irom 段时应当报错")
	}
}
