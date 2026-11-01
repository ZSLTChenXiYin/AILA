/* AILA ESP32 系最小二级 bootloader（freestanding，windowed ABI）。
 *
 * ESP32/S3 的 ROM 上电只做一件事：从 flash 固定偏移加载二级 bootloader 并跳转，
 * 应用必须由它加载（esp8266 是 ROM 直载，无此环节）。本 bootloader 只做四件事，
 * 不做 OTA / 分区表 / 安全启动——app 偏移与 WDT 寄存器来自 boot_cfg.h（chip.json 生成）：
 *   1. 关 RTC 看门狗（ROM 上电默认开启，不关会在数秒后复位芯片）
 *   2. UART 打横幅（复用 chip_cfg.h 的物理常量）
 *   3. 用 ROM 导出的 esp_rom_spiflash_read 把固定偏移（0x10000）的 AILA app
 *      镜像逐段拷进 RAM（段加载地址即 ELF paddr，全 RAM 段、无 flash 映射段）
 *   4. 跳 app 入口（windowed 普通函数调用，call4/8 由 GCC 决定）
 *
 * 镜像格式：esp32 系 V2——主头 8B（magic/段数/flash 参数/入口）+ 扩展头 16B +
 * 段表 + 校验和 + SHA256。bootloader 只读头与段表；校验和与 SHA 由 aila 打包时
 * 保证正确，ROM 加载本 bootloader 时已做过同样的格式校验。
 *
 * windowed ABI 注意：entry 指令不依赖调用方的 call4/8/12（旋转即可），ROM 无论
 * 以何种方式进入都安全；window overflow 向量沿用 ROM（VECBASE 保持默认）。
 */

#include <stdint.h>
#include <stddef.h>
#include "chip_cfg.h"
#include "boot_cfg.h"

#define AILA_NORETURN __attribute__((noreturn))

/* boot.ld 提供 */
extern uint8_t __boot_bss_start[];
extern uint8_t __boot_bss_end[];
extern uint8_t _boot_stack_top[];

/* ROM 导出（boot.ld PROVIDE 定址）：返回 0 = 成功 */
extern int esp_rom_spiflash_read(uint32_t src, uint32_t* dest, int32_t len);

#define UART_FIFO   (*(volatile uint32_t*)(UART0_BASE + 0x00))
#define UART_CLKDIV (*(volatile uint32_t*)(UART0_BASE + 0x14))
#define UART_STATUS (*(volatile uint32_t*)(UART0_BASE + 0x1C))

static void uart_init(void) {
    UART_CLKDIV = UART_CLKDIV_VAL;
}

static void uart_putc(uint8_t c) {
    while (((UART_STATUS >> UART_TXCNT_SHIFT) & UART_TXCNT_MASK) != 0) {
    }
    UART_FIFO = c;
}

static void uart_puts(const char* s) {
    while (*s) {
        uart_putc((uint8_t)*s++);
    }
}

static void uart_drain(void) {
    while (((UART_STATUS >> UART_TXCNT_SHIFT) & UART_TXCNT_MASK) != 0) {
    }
}

/* 禁 RTC 看门狗：解锁 → 喂狗 → CONFIG0 写 0（全阶段 OFF + 清 EN，rtc_wdt.c 的最小等效） */
static void wdt_disable(void) {
    *(volatile uint32_t*)(RTC_CNTL_BASE + WDT_WKEY_OFF) = WDT_WKEY_VALUE;
    *(volatile uint32_t*)(RTC_CNTL_BASE + WDT_FEED_OFF) = 1u << 31;
    *(volatile uint32_t*)(RTC_CNTL_BASE + WDT_CONFIG0_OFF) = 0;
}

AILA_NORETURN static void boot_fail(const char* why) {
    uart_puts(why);
    uart_drain();
    for (;;) {
    }
}

/* 分块读 flash：ROM 读按字进行，长度补到 4 的倍数，超出段尾的字节只落在栈缓冲 */
static int flash_read(uint32_t off, void* dst, uint32_t len) {
    static uint32_t buf[64]; /* 256B 一块（bss 里，不占栈） */
    uint8_t* out = (uint8_t*)dst;
    while (len > 0) {
        uint32_t chunk  = (len > sizeof(buf)) ? (uint32_t)sizeof(buf) : len;
        uint32_t padded = (chunk + 3u) & ~3u;
        if (esp_rom_spiflash_read(off, buf, (int32_t)padded) != 0) {
            return -1;
        }
        for (uint32_t i = 0; i < chunk; i++) {
            out[i] = ((uint8_t*)buf)[i];
        }
        off += padded;
        out += chunk;
        len -= chunk;
    }
    return 0;
}

void boot_main(void);

/* 入口桩：ROM → boot_start → 设栈 → boot_main。
 * p2align 必须：本桩由内联汇编跟在其他函数后（地址任意），entry 3 字节会
 * 使 l32r 的 PC+3 参考偏出字对齐——不对齐直接汇编报错（unaligned entry /
 * l32r out of range）。字面量在代码前，l32r 只引用更低地址。 */
__asm__(
    ".section .text\n"
    ".p2align 2\n"
    ".global boot_start\n"
    ".Lbs:\n"
    "   .word _boot_stack_top\n"
    ".p2align 2\n"
    "boot_start:\n"
    "   entry a1, 16\n"
    "   l32r  a1, .Lbs\n"
    "   call8 boot_main\n"
    ".Lbhang:\n"
    "   j .Lbhang\n");

void boot_main(void) {
    /* bss 自清：ROM loader 只拷 filesz>0 的段 */
    for (uint8_t* p = __boot_bss_start; p < __boot_bss_end; p++) {
        *p = 0;
    }
    wdt_disable();
    uart_init();
    uart_puts("AILA boot\r\n");

    /* 主头 8B + 扩展头 16B：hdr[0]=magic/段数/flash 参数，hdr[1]=入口地址 */
    uint32_t hdr[6];
    if (flash_read(BOOT_APP_OFFSET, hdr, sizeof(hdr)) != 0) {
        boot_fail("AILA boot: read header fail\r\n");
    }
    if ((hdr[0] & 0xFFu) != 0xE9u) {
        boot_fail("AILA boot: bad magic\r\n");
    }
    uint8_t  nseg  = (uint8_t)((hdr[0] >> 8) & 0xFFu);
    uint32_t entry = hdr[1];
    uint32_t off   = BOOT_APP_OFFSET + sizeof(hdr);
    for (uint8_t i = 0; i < nseg; i++) {
        uint32_t sh[2]; /* 4B 长度 + 4B 加载地址 */
        if (flash_read(off, sh, sizeof(sh)) != 0) {
            boot_fail("AILA boot: read segment header fail\r\n");
        }
        off += sizeof(sh);
        if (flash_read(off, (void*)(uintptr_t)sh[1], sh[0]) != 0) {
            boot_fail("AILA boot: read segment fail\r\n");
        }
        off += sh[0];
    }

    uart_drain();
    ((void (*)(void))entry)(); /* 进 app 的 call_start（windowed） */
    for (;;) {
    }
}
