// Package chip 是芯片家族的物理事实数据层：每家族一份 chip.json，aila 据此
// 生成链接脚本与 chip_cfg.h / boot_cfg.h——芯片适配「加数据不加代码」。
//
// 数值来源（ESP-IDF v5.1.4 对应文件 + 乐鑫 TRM，逐项核对）：
//   - 内存布局：components/esp_system/ld/<chip>/memory.ld.in（iram0_0_seg / dram0_0_seg）
//   - UART：components/soc/<chip>/include/soc/uart_reg.h + hal/uart_ll.h
//   - WDT：components/soc/<chip>/include/soc/rtc_cntl_reg.h + esp_hw_support/rtc_wdt.c
//   - ROM SPI flash：components/esp_rom/<chip>/ld/<chip>.rom.ld（esp_rom_spiflash_read）
//   - 镜像头：esptool bin_image.py（esp32 系主头 8B + 扩展头 16B + SHA256）
package chip

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"aila/internal/image"
)

// Chip 描述一个芯片家族的物理事实。
type Chip struct {
	Family string `json:"family"`
	// ABI 是 startup/链接层的调用约定：esp8266 = call0（LX106 硬件只有 call0），
	// esp32 系 = windowed（官方 GCC 8.4 工具链构建期固定，ESP-IDF 同款）。
	// AILA 代码生成不感知：IR → .o 由 -mcpu 决定 feature，与宿主一致。
	ABI string `json:"abi"`
	// Memory 是 RAM 布局（不把代码放 flash 执行，无 flash 映射段）。
	Memory struct {
		IRAMOrigin string `json:"iram_origin"`
		IRAMLen    string `json:"iram_len"`
		DRAMOrigin string `json:"dram_origin"`
		DRAMLen    string `json:"dram_len"`
		StackSize  string `json:"stack_size"` // 栈在 DRAM 顶向下
		Entry      string `json:"entry"`
	} `json:"memory"`
	// Image 是 flash 镜像头参数。size/freq 直接存编码值——esp8266 与 esp32 系
	// 的 flash 大小编码表不同（4MB：esp8266=4，esp32=2），不能共用字符串表。
	Image struct {
		Kind     string `json:"kind"` // esp8266 | esp32
		ChipID   uint16 `json:"chip_id"`
		ModeCode byte   `json:"mode_code"`
		SizeCode byte   `json:"size_code"`
		FreqCode byte   `json:"freq_code"`
	} `json:"image"`
	// CPUHz 是 CPU 主频（ccount 每周期 +1，time/sleep 原语的换算基准）：
	// esp8266 remap 后停在晶振二倍频 52MHz（本板 26MHz XTAL，80MHz 要 SDK 软件切，
	// 裸机运行时不切——见 chip.json _source.cpu_hz）；esp32/esp32s3 复位默认 240MHz。
	CPUHz uint64 `json:"cpu_hz"`
	// UART 物理常量（渲染进 chip_cfg.h，平台 C 只 include 不再硬编码）。
	UART struct {
		Base           string `json:"base"`
		ClkHz          uint64 `json:"clk_hz"` // UART 源时钟 = 当前 CPU/APB：esp8266 裸机 52M（SDK 80M）；esp32 系 80M；s3 XTAL 40M
		Baud           uint64 `json:"baud"`
		ClkdivFracBits uint   `json:"clkdiv_frac_bits"` // 0 = esp8266（寄存器无小数域）；4 = esp32 系（FRAG[23:20]）
		TXCntShift     uint   `json:"txcnt_shift"`      // STATUS 里 TXFIFO_CNT 的位偏移：esp8266/esp32[23:16] s3[25:16]
		TXCntBits      uint   `json:"txcnt_bits"`
		RXCntShift     uint   `json:"rxcnt_shift"` // STATUS 里 RXFIFO_CNT 的位偏移：esp8266/esp32[7:0] s3[9:0]
		RXCntBits      uint   `json:"rxcnt_bits"`
	} `json:"uart"`
	// GPIO 块寄存器偏移（相对 GPIO 基址，designs/0003 §5.1 的单一事实源）。
	// M1 只做主块位域 0-31：esp32 引脚 32-39 / esp32s3 32-48 走 OUT1/ENABLE1
	// 组寄存器（+0x10/+0x2C），M2 再扩；esp8266 物理引脚 0-15（GPIO16 属 RTC 域）。
	GPIO struct {
		Base       string `json:"base"`
		Out        string `json:"out"`
		OutW1TS    string `json:"out_w1ts"`
		OutW1TC    string `json:"out_w1tc"`
		Enable     string `json:"enable"`
		EnableW1TS string `json:"enable_w1ts"`
		EnableW1TC string `json:"enable_w1tc"`
		In         string `json:"in"`
		PinCount   uint   `json:"pin_count"` // M1 可用引脚数（≤32，位域上限）
	} `json:"gpio"`
	// Boot 是 esp32 系的自研 bootloader 约定；esp8266 ROM 直接加载 app，为 nil。
	Boot *Boot `json:"boot,omitempty"`
	// GCCFlags 是编平台 C / bootloader 时传给交叉 GCC 的附加标志。
	GCCFlags []string `json:"gcc_flags"`
	// IRFlags 是 IR → .o 时传给 llc/clang 的**芯片级**附加标志（与是否用 SDK 无关的
	// 后端事实：esp8266 是 call0 ABI，llc -O0 对高寄存器压力函数会报
	// "Cannot scavenge register"（esp-llvm 已知 RA 缺陷），必须 -O1）。
	// SDK 模式另有 SDK 级 ir_flags，两者叠加。
	IRFlags []string `json:"ir_flags,omitempty"`
	// SDKs 是该家族支持的闭源 SDK（预编译 .a + 头 + ld），键 = 注册名
	//（toolchains 里 sdk/<键>，aila.toml 的 [sdk] name = <键>）。
	SDKs map[string]*SDK `json:"sdks,omitempty"`
}

// SDK 描述一个闭源 SDK 的构建事实：链接布局、库序列与分区表。
// NonOS SDK v3.0.5 的语义参考：gen_appbin.py（镜像分段与 irom 填充）、
// eagle.app.v6.ld（段收集）、Partition Table 说明.md（user_pre_init 必须注册）。
type SDK struct {
	// Runtime 是平台入口层源文件名（runtime/<家族>/<Runtime>.c）：
	// 提供 SDK 回调 user_init / user_pre_init / user_rf_cal_sector_set，
	// 并在 user_init 里接管 AILA 的 aila_rt_main（堆初始化 + 调 main）。
	Runtime string `json:"runtime"`
	// FlashMap 是 flash 布局编号：镜像头 byte3 高 4 位（SDK call_user_start 按它
	// 配 cache 窗口），也是 system_partition_table_regist 的第三参。4 = 4MB(512+512)。
	FlashMap int `json:"flash_map"`
	// DRAMLen 是 SDK 布局的 DRAM 段长度（覆盖 Chip.Memory.DRAMLen：SDK 顶部留系统区）。
	DRAMLen string `json:"dram_len"`
	// IROM 是 flash 代码映射段：cache 把 flash 0x0 线性映射到 0x40200000，
	// 所以 irom 的虚拟地址必须等于「槽位 flash 偏移 + 镜像头长度」。
	// FOTA user1 槽（flash 0x1000）+ 16B 镜像头 ⇒ 0x40201010（官方 app1.ld 同值）。
	IROM struct {
		Origin string `json:"origin"` // irom 段虚拟地址（= 0x40200000 + Slot + 0x10）
		Len    string `json:"len"`    // 段长度上限
		Slot   string `json:"slot"`   // 应用镜像的槽位 flash 偏移（user1 = 0x1000）
	} `json:"irom"`
	// AilaHeap 是 AILA bump 堆窗口大小：埋在 _bss_end 与 SDK _heap_start 之间，
	// 互不侵犯（SDK 堆从 _aila_heap_end 起步）。
	AilaHeap string `json:"aila_heap"`
	// RomLD 是 SDK 自带的 ROM 符号脚本文件名（<sdk>/ld/<RomLD>，INCLUDE 进生成脚本）。
	RomLD string `json:"rom_ld"`
	// Libs 是链接库序列（官方 Makefile 的 --start-group 内顺序，组内循环解析）。
	Libs []string `json:"libs"`
	// GCCFlags 是编平台 C（aila_rt / SDK 入口层）追加的 GCC 标志。
	// SDK 布局下代码跨 IRAM→IROM 调用超 256KB 直调范围，官方全程 -mlongcalls。
	GCCFlags []string `json:"gcc_flags"`
	// IRFlags 是用户 IR → .o（llc）追加的标志（+long-calls 对应 GCC -mlongcalls；
	// esp8266 的 -O1 后置覆盖默认 -O0，绕开其 RA scavenge 缺陷，见该 chip.json _source）。
	IRFlags []string `json:"ir_flags"`
	// IRTextSection 非空时，用户 IR 的所有 define 落进该自定义段（esp8266 =
	// ".irom0.text"：IRAM 32K 被 libpp/libmain 占 23K，用户代码必须驻留 flash——
	// 即 SDK 的 ICACHE_FLASH_ATTR 语义；配 -mtext-section-literals 让 l32r 字面量
	// 随函数同段，避免跨 IRAM↔IROM 的 l32r 256KB 窗口越界）。
	IRTextSection string `json:"ir_text_section"`
	// CPUHz 非 0 时覆盖 Chip.CPUHz / UART.ClkHz（同值）：SDK 会在 user_init 前把
	// CPU 从复位默认的 52MHz 切到 80MHz，裸机的 chip_cfg.h 写的是 52MHz，
	// SDK 构建必须用这个值重算 UART 分频与 ccount 换算（build 的 linkSDK 以
	// -DAILA_UART_CLK_HZ / -DAILA_CPU_HZ 传入，chip_cfg.h 的同名宏有 #ifndef 守卫）。
	CPUHz uint64 `json:"cpu_hz"`
	// Partitions 是 user_pre_init 注册的分区表（partition_item_t = {type, addr, size}）。
	Partitions []SDKPartition `json:"partitions"`
}

// SDKPartition 是分区表一行（type 为 partition_type_t 枚举值：RF_CAL=4、
// PHY_DATA=5、SYSTEM_PARAMETER=6，CUSTOM_BEGIN=100 之后用户自定义）。
type SDKPartition struct {
	Type int    `json:"type"`
	Addr string `json:"addr"`
	Size string `json:"size"`
}

// Boot 是 bootloader 的落位与 ROM 服务地址。
type Boot struct {
	AppOffset  string `json:"app_offset"`  // app 镜像烧录偏移（固定地址读，无分区表）
	BootOffset string `json:"boot_offset"` // bootloader 自身烧录偏移（ROM 决定：esp32=0x1000，s3=0x0）
	IRAMOrigin string `json:"iram_origin"` // boot 代码区（不与 app 的 IRAM/DRAM 重叠）
	IRAMLen    string `json:"iram_len"`
	DRAMOrigin string `json:"dram_origin"`
	DRAMLen    string `json:"dram_len"`
	ROMSPIRead string `json:"rom_spiflash_read"` // ROM 导出的 esp_rom_spiflash_read 地址（boot.ld PROVIDE 定址）
	WDT        struct {
		RTCBase    string `json:"rtc_base"`
		WDTKeyOff  string `json:"wdtkey_off"`
		FeedOff    string `json:"feed_off"`
		Config0Off string `json:"config0_off"`
		Key        string `json:"key"`
	} `json:"wdt"`
}

// Load 读 <dir>/chip.json（dir = home.PlatformRuntime(family)）。
func Load(dir string) (Chip, error) {
	var c Chip
	raw, err := os.ReadFile(filepath.Join(dir, "chip.json"))
	if err != nil {
		return c, fmt.Errorf("读不了 chip.json：%w", err)
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return c, fmt.Errorf("解析 chip.json：%w", err)
	}
	if c.Family == "" || c.Memory.Entry == "" || c.Image.Kind == "" {
		return c, fmt.Errorf("chip.json 缺 family/memory.entry/image.kind")
	}
	return c, nil
}

func hex(s string) (uint64, error) {
	v, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("chip.json 里的十六进制值 %q 不合法", s)
	}
	return v, nil
}

// ClkDiv 算 UART 分频寄存器值：div16 = (clk<<4)/baud，整数段右移 4，
// esp32 系把 4 位小数放进 FRAG[23:20]；esp8266 寄存器无小数域，只取整数。
func (c Chip) ClkDiv() (uint32, error) {
	u := &c.UART
	if u.Baud == 0 {
		return 0, fmt.Errorf("chip.json 的 uart.baud 为 0")
	}
	div16 := (u.ClkHz << 4) / u.Baud
	val := div16 >> 4
	if u.ClkdivFracBits > 0 {
		val |= (div16 & 0xF) << 20
	}
	return uint32(val), nil
}

// TXCntMask 是 TXFIFO_CNT 的位域掩码。
func (c Chip) TXCntMask() uint64 {
	return (uint64(1) << c.UART.TXCntBits) - 1
}

// RXCntMask 是 RXFIFO_CNT 的位域掩码。
func (c Chip) RXCntMask() uint64 {
	return (uint64(1) << c.UART.RXCntBits) - 1
}

// LinkerScript 渲染 app 链接脚本（各 ESP 家族段落布局一致，只有数字不同）。
func (c Chip) LinkerScript() string {
	return fmt.Sprintf(`/* AILA %s 链接脚本——由 runtime/%s/chip.json 生成（内存布局的单一事实源） */
OUTPUT_FORMAT("elf32-xtensa-le")
ENTRY(%s)

MEMORY
{
    iram (rx)  : ORIGIN = %s, LENGTH = %s
    dram (rwx) : ORIGIN = %s, LENGTH = %s
}

SECTIONS
{
    .text :
    {
        *(.literal .literal.*)   /* Xtensa 字面量池（l32r 只能引用更低地址） */
        *(.text .text.*)
    } > iram

    .rodata :
    {
        *(.rodata .rodata.*)
    } > dram

    .data :
    {
        *(.data .data.*)
    } > dram

    .bss (NOLOAD) : ALIGN(8)
    {
        _bss_start = .;
        *(.bss .bss.*)
        *(COMMON)
        . = ALIGN(8);
        _bss_end = .;
    } > dram

    /* ROM 只加载已初始化的段（.rodata/.data），.bss 是 NOLOAD——上电后里面是 DRAM 残留，
       必须由平台 runtime 在调 main 前按 [_bss_start, _bss_end) 逐字清零，
       否则「零初始化」的静态量全是垃圾（实测：mono_ms 基线 2e14、固件随机挂在任意打印处）。 */

    /* 顶部留 %s 栈（向下生长）；bump 堆从 _bss_end 向上，与栈背向 */
    _heap_end  = ORIGIN(dram) + LENGTH(dram) - %s;
    _stack_top = ORIGIN(dram) + LENGTH(dram);

    /DISCARD/ :
    {
        *(.eh_frame .eh_frame.*)  /* freestanding 无异常：不进镜像 */
        *(.comment .note .note.*)
    }
}
`, c.Family, c.Family, c.Memory.Entry,
		c.Memory.IRAMOrigin, c.Memory.IRAMLen,
		c.Memory.DRAMOrigin, c.Memory.DRAMLen,
		c.Memory.StackSize, c.Memory.StackSize)
}

// BootLinkerScript 渲染 bootloader 链接脚本（ROM 服务地址 PROVIDE 定址，
// boot.c 里 extern 调用即可，不需要额外头文件）。
func (c Chip) BootLinkerScript() (string, error) {
	b := c.Boot
	if b == nil {
		return "", fmt.Errorf("家族 %s 没有 boot 段（esp8266 无需 bootloader）", c.Family)
	}
	return fmt.Sprintf(`/* AILA %s bootloader 链接脚本——由 runtime/%s/chip.json 生成 */
OUTPUT_FORMAT("elf32-xtensa-le")
ENTRY(boot_start)

MEMORY
{
    biram (rx)  : ORIGIN = %s, LENGTH = %s
    bdram (rwx) : ORIGIN = %s, LENGTH = %s
}

SECTIONS
{
    .text :
    {
        *(.literal .literal.*)
        *(.text .text.*)
    } > biram

    .data :
    {
        *(.data .data.*)
    } > bdram

    .bss (NOLOAD) :
    {
        __boot_bss_start = .;
        *(.bss .bss.*)
        *(COMMON)
        . = ALIGN(8);
        __boot_bss_end = .;
    } > bdram

    _boot_stack_top = ORIGIN(bdram) + LENGTH(bdram);

    /* ROM 导出的 SPI flash 读函数（esp32.rom.ld / esp32s3.rom.ld 同名符号） */
    PROVIDE(esp_rom_spiflash_read = %s);

    /DISCARD/ :
    {
        *(.eh_frame .eh_frame.*)
        *(.comment .note .note.*)
    }
}
`, c.Family, c.Family,
		b.IRAMOrigin, b.IRAMLen,
		b.DRAMOrigin, b.DRAMLen,
		b.ROMSPIRead), nil
}

// ChipCfgH 渲染平台 runtime C 用的 UART / GPIO 物理常量头。
// GPIO 段可选：chip.json 无 gpio 时省略（平台 runtime 依赖三家族均已有）。
func (c Chip) ChipCfgH() (string, error) {
	div, err := c.ClkDiv()
	if err != nil {
		return "", err
	}
	gpioLines := ""
	if c.GPIO.Base != "" {
		gpioOut, err := hex(c.GPIO.Out)
		if err != nil {
			return "", err
		}
		gpioW1TS, err := hex(c.GPIO.OutW1TS)
		if err != nil {
			return "", err
		}
		gpioW1TC, err := hex(c.GPIO.OutW1TC)
		if err != nil {
			return "", err
		}
		gpioEn, err := hex(c.GPIO.Enable)
		if err != nil {
			return "", err
		}
		gpioEnW1TS, err := hex(c.GPIO.EnableW1TS)
		if err != nil {
			return "", err
		}
		gpioEnW1TC, err := hex(c.GPIO.EnableW1TC)
		if err != nil {
			return "", err
		}
		gpioIn, err := hex(c.GPIO.In)
		if err != nil {
			return "", err
		}
		gpioLines = fmt.Sprintf(`#define GPIO0_BASE          %su
#define GPIO_OUT_OFF        0x%Xu
#define GPIO_OUT_W1TS_OFF   0x%Xu
#define GPIO_OUT_W1TC_OFF   0x%Xu
#define GPIO_ENABLE_OFF     0x%Xu
#define GPIO_ENABLE_W1TS_OFF 0x%Xu
#define GPIO_ENABLE_W1TC_OFF 0x%Xu
#define GPIO_IN_OFF         0x%Xu
#define GPIO_PIN_COUNT      %du
`, c.GPIO.Base, gpioOut, gpioW1TS, gpioW1TC, gpioEn, gpioEnW1TS, gpioEnW1TC, gpioIn,
			c.GPIO.PinCount)
	}
	return fmt.Sprintf(`/* 由 aila build 从 runtime/%s/chip.json 生成——UART / GPIO 物理事实的单一来源 */
#define UART0_BASE          %su
#ifndef AILA_UART_CLK_HZ /* SDK 构建以 -DAILA_UART_CLK_HZ 覆盖：SDK 会把 CPU 切到 80MHz */
#define UART_CLK_HZ         %du
#endif
#define UART_BAUD_HZ        %du
#define UART_CLKDIV_VAL     0x%Xu
#define UART_CLKDIV_FRAC_BITS %du
#define UART_TXCNT_SHIFT    %du
#define UART_TXCNT_MASK     0x%Xu
#define UART_RXCNT_SHIFT    %du
#define UART_RXCNT_MASK     0x%Xu
%s#ifndef AILA_CPU_HZ /* 同上：ccount → 毫秒的换算基准必须与实际主频一致 */
#define AILA_CPU_HZ         %du
#endif
`, c.Family, c.UART.Base, c.UART.ClkHz, c.UART.Baud, div, c.UART.ClkdivFracBits,
		c.UART.TXCntShift, c.TXCntMask(), c.UART.RXCntShift, c.RXCntMask(),
		gpioLines, c.CPUHz), nil
}

// BootCfgH 渲染 bootloader 用的偏移与看门狗寄存器常量头。
func (c Chip) BootCfgH() (string, error) {
	b := c.Boot
	if b == nil {
		return "", fmt.Errorf("家族 %s 没有 boot 段（esp8266 无需 bootloader）", c.Family)
	}
	app, err := hex(b.AppOffset)
	if err != nil {
		return "", err
	}
	rtc, err := hex(b.WDT.RTCBase)
	if err != nil {
		return "", err
	}
	keyOff, err := hex(b.WDT.WDTKeyOff)
	if err != nil {
		return "", err
	}
	feedOff, err := hex(b.WDT.FeedOff)
	if err != nil {
		return "", err
	}
	cfgOff, err := hex(b.WDT.Config0Off)
	if err != nil {
		return "", err
	}
	key, err := hex(b.WDT.Key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`/* 由 aila build 从 runtime/%s/chip.json 生成——bootloader 的烧录偏移与 WDT 寄存器 */
#define BOOT_APP_OFFSET  0x%Xu
#define RTC_CNTL_BASE    0x%Xu
#define WDT_WKEY_OFF     0x%Xu
#define WDT_FEED_OFF     0x%Xu
#define WDT_CONFIG0_OFF  0x%Xu
#define WDT_WKEY_VALUE   0x%Xu
`, c.Family, app, rtc, keyOff, feedOff, cfgOff, key), nil
}

// SizeFreq 是镜像头 byte3（size<<4 | freq）。
func (c Chip) SizeFreq() byte { return c.Image.SizeCode<<4 | c.Image.FreqCode }

// SDK 按注册名取 SDK 定义。
func (c Chip) SDK(name string) (*SDK, error) {
	s, ok := c.SDKs[name]
	if !ok {
		return nil, fmt.Errorf("chip.json 里没有 SDK %q（可用：%v）", name, mapKeys(c.SDKs))
	}
	return s, nil
}

func mapKeys(m map[string]*SDK) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// SDKLinkerScript 渲染 SDK 模式链接脚本。段落语义照抄 NonOS SDK 的
// eagle.app.v6.ld（老布局非 FOTA）：irom0.text 按库名收集大库进 flash，
// libmain/phy/pp 的热代码留在 IRAM；dport/dram/iram/irom 四区。
// 与官方的差异只有两处：_aila_heap_end 埋 AILA bump 窗口、INCLUDE 用绝对路径。
func (c Chip) SDKLinkerScript(s *SDK, romLdPath string) string {
	// DRAM 长度被 SDK 布局覆盖（SDK 顶部留系统区，比裸机短）
	dramLen := s.DRAMLen
	if dramLen == "" {
		dramLen = c.Memory.DRAMLen
	}
	return fmt.Sprintf(`/* AILA %s SDK 模式链接脚本——按 NonOS SDK eagle.app.v6.ld 语义生成 */
OUTPUT_FORMAT("elf32-xtensa-le")
ENTRY(call_user_start)
EXTERN(_DebugExceptionVector)
EXTERN(_DoubleExceptionVector)
EXTERN(_KernelExceptionVector)
EXTERN(_NMIExceptionVector)
EXTERN(_UserExceptionVector)
PROVIDE(_memmap_vecbase_reset = 0x40000000);
/* cache 属性常量（官方脚本原样，SDK 启动代码引用） */
_memmap_cacheattr_wb_base = 0x00000110;
_memmap_cacheattr_wt_base = 0x00000110;
_memmap_cacheattr_bp_base = 0x00000220;
_memmap_cacheattr_unused_mask = 0xFFFFF00F;
_memmap_cacheattr_wb_trapnull = 0x2222211F;
_memmap_cacheattr_wba_trapnull = 0x2222211F;
_memmap_cacheattr_wbna_trapnull = 0x2222211F;
_memmap_cacheattr_wt_trapnull = 0x2222211F;
_memmap_cacheattr_bp_trapnull = 0x2222222F;
_memmap_cacheattr_wb_strict = 0xFFFFF11F;
_memmap_cacheattr_wt_strict = 0xFFFFF11F;
_memmap_cacheattr_bp_strict = 0xFFFFF22F;
_memmap_cacheattr_wb_allvalid = 0x22222112;
_memmap_cacheattr_wt_allvalid = 0x22222112;
_memmap_cacheattr_bp_allvalid = 0x22222222;
PROVIDE(_memmap_cacheattr_reset = _memmap_cacheattr_wb_trapnull);

MEMORY
{
    dport0_0_seg : org = 0x3FF00000, len = 0x10
    dram0_0_seg  : org = %s, len = %s
    iram1_0_seg  : org = %s, len = %s
    irom0_0_seg  : org = %s, len = %s
}

PHDRS
{
    dport0_0_phdr PT_LOAD;
    dram0_0_phdr PT_LOAD;
    dram0_0_bss_phdr PT_LOAD;
    iram1_0_phdr PT_LOAD;
    irom0_0_phdr PT_LOAD;
}

SECTIONS
{
    .dport0.rodata : ALIGN(4)
    {
        _dport0_rodata_start = ABSOLUTE(.);
        *(.dport0.rodata)
        *(.dport.rodata)
        _dport0_rodata_end = ABSOLUTE(.);
    } >dport0_0_seg :dport0_0_phdr

    .dport0.literal : ALIGN(4)
    {
        _dport0_literal_start = ABSOLUTE(.);
        *(.dport0.literal)
        *(.dport.literal)
        _dport0_literal_end = ABSOLUTE(.);
    } >dport0_0_seg :dport0_0_phdr

    .dport0.data : ALIGN(4)
    {
        _dport0_data_start = ABSOLUTE(.);
        *(.dport0.data)
        *(.dport.data)
        _dport0_data_end = ABSOLUTE(.);
    } >dport0_0_seg :dport0_0_phdr

    .data : ALIGN(4)
    {
        _data_start = ABSOLUTE(.);
        *(.data)
        *(.data.*)
        *(.gnu.linkonce.d.*)
        *(.data1)
        *(.sdata)
        *(.sdata.*)
        *(.gnu.linkonce.s.*)
        *(.sdata2)
        *(.sdata2.*)
        *(.gnu.linkonce.s2.*)
        *(.jcr)
        _data_end = ABSOLUTE(.);
    } >dram0_0_seg :dram0_0_phdr

    .rodata : ALIGN(4)
    {
        _rodata_start = ABSOLUTE(.);
        *(.sdk.version)
        *(.rodata)
        *(.rodata.*)
        *(.gnu.linkonce.r.*)
        *(.rodata1)
        __XT_EXCEPTION_TABLE__ = ABSOLUTE(.);
        *(.xt_except_table)
        *(.gcc_except_table)
        *(.gnu.linkonce.e.*)
        *(.gnu.version_r)
        *(.eh_frame)
        . = ALIGN(4);
        /* SDK 启动代码按此表清 bss（_bss_start.._bss_end 逐字） */
        _bss_table_start = ABSOLUTE(.);
        LONG(_bss_start)
        LONG(_bss_end)
        _bss_table_end = ABSOLUTE(.);
        _rodata_end = ABSOLUTE(.);
    } >dram0_0_seg :dram0_0_phdr

    .bss ALIGN(8) (NOLOAD) : ALIGN(4)
    {
        . = ALIGN(8);
        _bss_start = ABSOLUTE(.);
        *(.dynsbss)
        *(.sbss)
        *(.sbss.*)
        *(.gnu.linkonce.sb.*)
        *(.scommon)
        *(.sbss2)
        *(.sbss2.*)
        *(.gnu.linkonce.sb2.*)
        *(.dynbss)
        *(.bss)
        *(.bss.*)
        *(.gnu.linkonce.b.*)
        *(COMMON)
        . = ALIGN(8);
        _bss_end = ABSOLUTE(.);
        /* AILA bump 堆窗口 [_bss_end, _aila_heap_end)；SDK 堆从 _aila_heap_end 起步。
         * _heap_end 是 aila_rt 的 bump 上界别名（与裸机链接脚本同名同义） */
        _aila_heap_end = _bss_end + %s;
        _heap_end = _aila_heap_end;
        _heap_start = _aila_heap_end;
    } >dram0_0_seg :dram0_0_bss_phdr

    /* SDK 启动对象必须驻留 IRAM —— 这里必须排在 flash 兜底之前：
       随包 SDK 的 .a 是 -ffunction-sections 编的，若不在这里先捕获，
       下面的 .irom0.text 兜底会把它们扫进 flash。而 app_main.o 里的
       call_user_start_local / Cache_Read_Enable_New 负责「打开 flash cache」，
       它们自己在 flash 里时，入口 callx0 过去就静默挂死
       （实测：bootloader 正常 jump，之后无任何输出）。 */
    .iram0.boot : ALIGN(4)
    {
        /* 只拉启动必需的两个函数（-ffunction-sections 让它们各自独立成段）：
           其余 app_main.o 函数留在 flash，避免 IRAM 溢出 */
        *libmain.a:app_main.o(.text.call_user_start_local .literal.call_user_start_local)
        *libmain.a:app_main.o(.text.Cache_Read_Enable_New .literal.Cache_Read_Enable_New)
        *libmain.a:app_main.o(.text.user_start .literal.user_start)
    } >iram1_0_seg :iram1_0_phdr

    /* flash 代码段：官方 eagle.app.v6.ld 原样——AT/ssl 等库的代码在普通
     * .text 段，靠库名规则进 flash；ICACHE_FLASH_ATTR 代码在 .irom0.text
     * 自定义段，靠全局收集落位（libmain/phy/pp 热代码留在 IRAM） */
    .irom0.text : ALIGN(4)
    {
        _irom0_text_start = ABSOLUTE(.);
        *libat.a:(.literal.* .text.*)
        *libcrypto.a:(.literal.* .text.*)
        *libespnow.a:(.literal.* .text.*)
        *libjson.a:(.literal.* .text.*)
        *liblwip.a:(.literal.* .text.*)
        *libnet80211.a:(.literal.* .text.*)
        *libsmartconfig.a:(.literal.* .text.*)
        *libssl.a:(.literal.* .text.*)
        *libupgrade.a:(.literal.* .text.*)
        *libwpa.a:(.literal.* .text.*)
        *libwpa2.a:(.literal.* .text.*)
        *libwps.a:(.literal.* .text.*)
        *libmbedtls.a:(.literal.* .text.*)
        *libm.a:(.literal .text .literal.* .text.*)
        *(.irom0.literal .irom.literal .irom.text.literal .irom0.text .irom.text)
        /* 兜底：其余所有 .text.* / .literal.* 一律进 flash。
           启动必需的对象（app_main.o 的 call_user_start_local / Cache_Read_Enable_New）
           已在上面 .iram0.boot 里先行捕获，所以不会被这个兜底扫走。
           随包 SDK 的 .a 是 -ffunction-sections 编的（函数各自成 .text.xxx 段），
           因此这个兜底不能省——省掉它们会变成孤儿段落进 IRAM，直接 32KB 溢出；
           但也不能把它单独用于启动对象，否则开 cache 的代码待在 flash 里，入口必死。
           AILA 自己的 IR 由 placeIRInSection 显式落 .irom0.text（上面那行已收）。 */
        *(.text.* .literal.*)
        _irom0_text_end = ABSOLUTE(.);
    } >irom0_0_seg :irom0_0_phdr

    .text : ALIGN(4)
    {
        _stext = .;
        _text_start = ABSOLUTE(.);
        *(.UserEnter.text)
        . = ALIGN(16);
        *(.DebugExceptionVector.text)
        . = ALIGN(16);
        *(.NMIExceptionVector.text)
        . = ALIGN(16);
        *(.KernelExceptionVector.text)
        LONG(0)
        LONG(0)
        LONG(0)
        LONG(0)
        . = ALIGN(16);
        *(.UserExceptionVector.text)
        LONG(0)
        LONG(0)
        LONG(0)
        LONG(0)
        . = ALIGN(16);
        *(.DoubleExceptionVector.text)
        LONG(0)
        LONG(0)
        LONG(0)
        LONG(0)
        . = ALIGN(16);
        *(.entry.text)
        *(.init.literal)
        *(.init)
        *(.literal .text .stub .gnu.warning .gnu.linkonce.literal.* .gnu.linkonce.t.*.literal .gnu.linkonce.t.*)
        *(.fini.literal)
        *(.fini)
        *(.gnu.version)
        _text_end = ABSOLUTE(.);
        _etext = .;
    } >iram1_0_seg :iram1_0_phdr

    .lit4 : ALIGN(4)
    {
        _lit4_start = ABSOLUTE(.);
        *(*.lit4)
        *(.lit4.*)
        *(.gnu.linkonce.lit4.*)
        _lit4_end = ABSOLUTE(.);
    } >iram1_0_seg :iram1_0_phdr

    /DISCARD/ :
    {
        *(.comment .note .note.*)
    }
}

/* ROM 导出符号地址表（SDK 自带，路径由 aila 生成时注入） */
INCLUDE "%s"
`, c.Family,
		c.Memory.DRAMOrigin, dramLen,
		c.Memory.IRAMOrigin, c.Memory.IRAMLen,
		s.IROM.Origin, s.IROM.Len,
		s.AilaHeap,
		filepath.ToSlash(romLdPath))
}

// SDKCfgH 渲染 SDK 模式入口层（nosdk.c）用的配置头：flash map + 分区表数据。
// 表数据用初始化列表宏——partition_item_t 类型在 nosdk.c 里 extern 声明，
// 不依赖 SDK 头文件。
func (c Chip) SDKCfgH(s *SDK) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "/* 由 aila build 从 runtime/%s/chip.json 生成——SDK 模式的分区表与 flash 布局 */\n", c.Family)
	fmt.Fprintf(&b, "#define SDK_FLASH_MAP %du\n", s.FlashMap)
	fmt.Fprintf(&b, "#define SDK_PARTITION_TABLE_DATA { \\") // 行尾续行：宏内容紧随，不能自带换行
	for _, p := range s.Partitions {
		addr, err := hex(p.Addr)
		if err != nil {
			return "", err
		}
		size, err := hex(p.Size)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "\n    {%du, 0x%Xu, 0x%Xu}, \\", p.Type, addr, size)
	}
	b.WriteString("\n}\n")
	return b.String(), nil
}

// Pack 把链接好的 ELF 打成该家族的 flash 镜像（image 包按 kind 分支）。
func (c Chip) Pack(elfPath, outPath string) error {
	switch c.Image.Kind {
	case "esp8266":
		return image.Pack8266(elfPath, outPath, c.Image.ModeCode, c.SizeFreq())
	case "esp32":
		return image.PackESP32(elfPath, outPath, c.Image.ChipID, c.Image.ModeCode, c.SizeFreq())
	default:
		return fmt.Errorf("chip.json 的 image.kind %q 不认识（esp8266 | esp32）", c.Image.Kind)
	}
}

// PackSDK 打包 SDK 模式的 FOTA v2 镜像（烧 user1 槽 flash 0x1000 起）：
// irom 段紧跟 16B 镜像头内嵌，其余 RAM 段随后；irom 的虚拟地址必须与
// 槽位对齐（0x40200000 ↔ flash 0x0 固定线性），否则上板立刻非法指令。
func (c Chip) PackSDK(elfPath, outPath string, s *SDK) error {
	origin, err := hex(s.IROM.Origin)
	if err != nil {
		return err
	}
	slot, err := hex(s.IROM.Slot)
	if err != nil {
		return err
	}
	return image.Pack8266SDK(elfPath, outPath, c.Image.ModeCode, c.SizeFreq(),
		uint32(origin), uint32(slot))
}
