// Package image 把链接好的 ELF 打包成乐鑫 flash 镜像——esptool elf2image 的
// Go 原生实现，`aila build` 出可烧录产物不依赖任何外部工具。
//
// ELF 解析是手写的 ELF32 小端读取（program header 表），不引 debug/elf——
// 只需要 e_entry 与 PT_LOAD 的 (paddr, offset, filesz)，手解 40 行更直接。
//
// 两个镜像格式（esptool 的 esp8266_v2 / esp32 系 V2）：
//
//	主头 8B：magic 0xE9 · 段数 · flash 模式 · flash 大小/频率 · 入口地址（4B LE）
//	每段：4B 长度 + 4B 加载地址 + 数据（取自 ELF 的 PT_LOAD 段）
//	末尾补 0xFF 到 16B 对齐（校验和占最后一字节）；校验和 = 0xEF XOR 全部段数据
//	esp32 系差异：主头后带 16B 扩展头（wp_pin / chip_id / hash_appended），
//	校验和之后追加 32B SHA256（覆盖此前整镜像）
//
// flash 大小编码两族不同（4MB：esp8266=4、esp32=2），编码值由 chip.json 提供。
package image

import (
	"sort"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
)

// Flash 是镜像头里的 flash 参数（默认 DIO / 4MB / 40MHz，兼容绝大多数 ESP8266 模块）。
type Flash struct {
	Mode string // qio | qout | dio | dout
	Size string // 256K | 512K | 1M | 2M | 4M | 8M | 16M
	Freq string // 40m | 26m | 20m | 80m
}

// DefaultFlash 返回默认 flash 参数。
func DefaultFlash() Flash { return Flash{Mode: "dio", Size: "4M", Freq: "40m"} }

var flashModes = map[string]byte{"qio": 0, "qout": 1, "dio": 2, "dout": 3}
var flashSizes = map[string]byte{"256K": 1, "512K": 0, "1M": 2, "2M": 3, "4M": 4, "8M": 5, "16M": 6}
var flashFreqs = map[string]byte{"40m": 0, "26m": 1, "20m": 2, "80m": 0xF}

// ELF32 常量：program header 类型与布局（字节数为小端 32 位）。
const (
	phTypeLoad = 1  // PT_LOAD
	phSize     = 32 // 一个 Phdr32 的大小
	ehEntryOff = 24 // e_entry 在 ELF 头里的偏移
	ehPhoffOff = 28 // e_phoff
	ehPhnumOff = 44 // e_phnum（e_phentsize 在 42）
)

// seg 是一个待打包的可加载段。
type seg struct {
	addr uint32
	data []byte
}

// loadSegments 手解 ELF32，取 e_entry 与全部 PT_LOAD（filesz>0，地址 paddr 优先 vaddr 兜底）。
func loadSegments(elfPath string) (entry uint32, segs []seg, err error) {
	raw, err := os.ReadFile(elfPath)
	if err != nil {
		return 0, nil, fmt.Errorf("读不了 ELF：%w", err)
	}
	if len(raw) < 52 || string(raw[:4]) != "\x7fELF" || raw[4] != 1 || raw[5] != 1 {
		return 0, nil, fmt.Errorf("不是 ELF32 小端文件：%s", elfPath)
	}
	phoff := binary.LittleEndian.Uint32(raw[ehPhoffOff:])
	phnum := binary.LittleEndian.Uint16(raw[ehPhnumOff:])
	for i := 0; i < int(phnum); i++ {
		at := int(phoff) + i*phSize
		if at+phSize > len(raw) {
			return 0, nil, fmt.Errorf("program header 越界")
		}
		ph := raw[at : at+phSize]
		if binary.LittleEndian.Uint32(ph[0:]) != phTypeLoad {
			continue
		}
		filesz := binary.LittleEndian.Uint32(ph[16:])
		if filesz == 0 {
			continue
		}
		off := binary.LittleEndian.Uint32(ph[4:])
		if int(off+filesz) > len(raw) {
			return 0, nil, fmt.Errorf("段数据越界")
		}
		addr := binary.LittleEndian.Uint32(ph[12:]) // p_paddr
		if addr == 0 {
			addr = binary.LittleEndian.Uint32(ph[8:]) // p_vaddr 兜底
		}
		segs = append(segs, seg{addr: addr, data: raw[off : off+filesz]})
	}
	if len(segs) == 0 {
		return 0, nil, fmt.Errorf("ELF 里没有可加载段")
	}
	if len(segs) > 255 {
		return 0, nil, fmt.Errorf("段数 %d 超出镜像格式上限", len(segs))
	}
	return binary.LittleEndian.Uint32(raw[ehEntryOff:]), segs, nil
}

// appendSegment 把一段写进镜像并累计校验和。
// 段头字节序是 (addr, len)——esptool bin_image.py save_segment 与 SDK
// gen_appbin.py 一致，与直觉相反（len 在后），照抄官方不得改动。
//
// 段数据必须补 0 到 8 字节对齐：ROM loader 按字（4B）搬运，长度不是整字时会把
// 后面那段「段头」的前几字节也划进本段一起校验，checksum 就对不上，表现为
// 「打完 tail 就静默停住、连 chksum 都不打印」（实测 SDK 镜像 len 8375 → tail 7）。
// 官方 gen_appbin/bin_image 同样补 0；补的是 0，XOR 校验和不变。
func appendSegment(out []byte, s seg, checksum *byte) []byte {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], s.addr)
	out = append(out, buf[:]...)
	if pad := (8 - len(s.data)%8) % 8; pad != 0 {
		s.data = append(s.data[:len(s.data):len(s.data)], make([]byte, pad)...) // 三索引切片：强制拷贝，别动 ELF 缓冲
	}
	binary.LittleEndian.PutUint32(buf[:], uint32(len(s.data)))
	out = append(out, buf[:]...)
	out = append(out, s.data...)
	for _, b := range s.data {
		*checksum ^= b
	}
	return out
}

// Pack8266 把 ELF 打成 ESP8266 V2 镜像（编码值直传，chip 数据层调用）。
func Pack8266(elfPath, outPath string, mode, sizeFreq byte) error {
	entry, segs, err := loadSegments(elfPath)
	if err != nil {
		return err
	}

	out := make([]byte, 0, 8+len(segs)*(8+64))
	out = append(out, 0xE9, byte(len(segs)), mode, sizeFreq)
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], entry)
	out = append(out, buf[:]...)

	// 段顺序：IRAM（0x40100000..）在 DRAM 之前——ROM loader 依赖这个顺序，
	// 反过来时加载完 IRAM 就停住（不跳转、不打 checksum），表现成「烧进去没反应」。
	sort.SliceStable(segs, func(i, j int) bool {
		ii := segs[i].addr >= 0x40100000
		ij := segs[j].addr >= 0x40100000
		if ii != ij {
			return ii
		}
		return segs[i].addr < segs[j].addr
	})

	checksum := byte(0xEF)
	for _, s := range segs {
		out = appendSegment(out, s, &checksum)
	}
	for (len(out)+1)%16 != 0 {
		out = append(out, 0xFF)
	}
	out = append(out, checksum)

	return os.WriteFile(outPath, out, 0o666)
}

// Pack8266SDK 打包 SDK 模式镜像（NonOS SDK 老布局，参考 gen_appbin.py）：
// RAM 段（paddr < iromOrigin）进 0xE9 头——(addr,len) 段头 + checksum；
// irom 段（paddr ≥ iromOrigin）不进头：checksum 字节之后 0xFF 填充到
// iromFlashOff 再放裸数据。ROM loader 只按头部加载 RAM 段并校验 checksum，
// irom 代码由 flash cache 线性映射（0x40200000 ↔ flash 0x0）直读——所以
// irom 数据必须落在 flash 绝对偏移 (iromOrigin - 0x40200000) 处。
// 等价于官方 none 模式的 flash.bin + irom0text.bin 两文件，这里合并成
// 单文件（仿官方 FOTA 模式的 0xFF 填充手法），烧 0x0 一片搞定。
func Pack8266SDK(elfPath, outPath string, mode, sizeFreq byte, iromOrigin, iromFlashOff uint32) error {
	entry, segs, err := loadSegments(elfPath)
	if err != nil {
		return err
	}
	var ram []seg
	var irom *seg
	for i := range segs {
		if segs[i].addr >= iromOrigin {
			if irom != nil {
				return fmt.Errorf("SDK 镜像只允许一个 irom 段")
			}
			irom = &segs[i]
		} else {
			ram = append(ram, segs[i])
		}
	}
	if irom == nil {
		return fmt.Errorf("SDK 镜像缺 irom 段（paddr ≥ %#x）", iromOrigin)
	}
	if len(irom.data)%4 != 0 {
		return fmt.Errorf("irom 段长度 %#x 不是 4 的倍数", len(irom.data))
	}

	// RAM 段顺序：IRAM（0x40100000..）必须在最前，DRAM 段按地址升序随后。
	// 官方镜像就是这个顺序（IRAM → DRAM）；倒过来时 ROM loader 加载完两段后
	// 不跳转、也不打最后一行 checksum，表现为「烧进去毫无反应」。
	sort.SliceStable(ram, func(i, j int) bool {
		ii := ram[i].addr >= 0x40100000
		ij := ram[j].addr >= 0x40100000
		if ii != ij {
			return ii
		}
		return ram[i].addr < ram[j].addr
	})

	out := make([]byte, 0, 8+len(ram)*(8+64)+int(iromFlashOff))
	out = append(out, 0xE9, byte(len(ram)), mode, sizeFreq)
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], entry)
	out = append(out, buf[:]...)

	checksum := byte(0xEF)
	for _, s := range ram {
		out = appendSegment(out, s, &checksum)
	}
	for (len(out)+1)%16 != 0 {
		out = append(out, 0xFF)
	}
	out = append(out, checksum)

	if uint32(len(out)) > iromFlashOff {
		return fmt.Errorf("RAM 段总长 %#x 已越过 irom 的 flash 偏移 %#x", len(out), iromFlashOff)
	}
	for uint32(len(out)) < iromFlashOff {
		out = append(out, 0xFF)
	}
	out = append(out, irom.data...)

	return os.WriteFile(outPath, out, 0o666)
}

// ESP8266 把 ELF 的 PT_LOAD 段打成 ESP8266 V2 镜像（烧录到 flash 0x0）。
func ESP8266(elfPath, outPath string, f Flash) error {
	mode, ok := flashModes[f.Mode]
	if !ok {
		return fmt.Errorf("未知 flash 模式 %q（qio/qout/dio/dout）", f.Mode)
	}
	size, ok := flashSizes[f.Size]
	if !ok {
		return fmt.Errorf("未知 flash 大小 %q", f.Size)
	}
	freq, ok := flashFreqs[f.Freq]
	if !ok {
		return fmt.Errorf("未知 flash 频率 %q", f.Freq)
	}
	return Pack8266(elfPath, outPath, mode, size<<4|freq)
}

// PackESP32 把 ELF 打成 ESP32 系 V2 镜像：主头 8B + 扩展头 16B + 段表 +
// 0xFF 补齐 16 对齐 + 校验和（0xEF XOR 段数据）+ 32B SHA256（覆盖此前整镜像）。
func PackESP32(elfPath, outPath string, chipID uint16, mode, sizeFreq byte) error {
	entry, segs, err := loadSegments(elfPath)
	if err != nil {
		return err
	}

	out := make([]byte, 0, 24+len(segs)*(8+64))
	out = append(out, 0xE9, byte(len(segs)), mode, sizeFreq)
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], entry)
	out = append(out, buf[:]...)
	// 扩展头 16B：wp_pin=0xEE（禁用）· spi_pin_drv×3=0 · chip_id(2B LE) ·
	// min_chip_rev=0 · min/max_chip_rev_full=0 · 保留×4=0 · hash_appended=1
	var buf2 [2]byte
	binary.LittleEndian.PutUint16(buf2[:], chipID)
	out = append(out, 0xEE, 0, 0, 0)
	out = append(out, buf2[:]...)
	out = append(out, make([]byte, 10)...) // 0x0E..0x16：min_chip_rev 等 + 保留×4
	out[len(out)-1] = 1                    // hash_appended（扩展头末字节，偏移 0x17）

	// 段顺序：IRAM（0x40100000..）在 DRAM 之前——ROM loader 依赖这个顺序，
	// 反过来时加载完 IRAM 就停住（不跳转、不打 checksum），表现成「烧进去没反应」。
	sort.SliceStable(segs, func(i, j int) bool {
		ii := segs[i].addr >= 0x40100000
		ij := segs[j].addr >= 0x40100000
		if ii != ij {
			return ii
		}
		return segs[i].addr < segs[j].addr
	})

	checksum := byte(0xEF)
	for _, s := range segs {
		out = appendSegment(out, s, &checksum)
	}
	for (len(out)+1)%16 != 0 {
		out = append(out, 0xFF)
	}
	out = append(out, checksum)

	digest := sha256.Sum256(out)
	out = append(out, digest[:]...)

	return os.WriteFile(outPath, out, 0o666)
}
