/* AILA ESP32 平台运行时（freestanding，无 libc）。
 *
 * 对应 runtime/aila_rt.c 的宿主实现，七个入口符号与语义不变：
 *   - 区域账本：O(1) bump + mark 复位——region_enter 记 bump 快照，
 *     region_leave 把 bump 指针退回快照，整个区域一次回收（designs/0001 §12.3）
 *   - trap / write / flush：走 UART0
 *
 * ABI 注意（必须与代码生成一致）：
 *   - `aila_rt_trap` / `aila_rt_region_*` 的标量参数在代码生成里固定为 i64
 *     （internal/codegen：ensureDeclare 的硬编码签名），C 侧用 int64_t；
 *   - `aila_rt_write` 的 n 是 USize——esp32 指针宽 4 字节，按 32 位传参。
 *
 * @static 的假设：静态分配发生在模块初始化（进入 main 之前、任何 region 之外），
 * 因此普通 region 的 mark 恒高于静态分配水位，leave 不会误回收静态内存。
 */

#include <stdint.h>
#include <stddef.h>

#define AILA_NORETURN __attribute__((noreturn))

/* trap 的消息长度用 sizeof 编译期算，不手数 */
#define RT_TRAP(m) aila_rt_trap(m, (int64_t)(sizeof(m) - 1), "aila_rt", 0)

AILA_NORETURN void aila_rt_trap(const char* msg, int64_t len, const char* file, int64_t line);

/* ---------------------------------------------------------------------- */
/* UART0 / GPIO0：物理常量来自 chip_cfg.h（由 chip.json 生成，单一事实源） */
/* UART：FIFO@0x00、CLKDIV@0x14、STATUS@0x1C（TX/RX [23:16]/[7:0]，        */
/*   esp-idf uart_reg.h）；GPIO 块偏移见 esp32 chip.json _source           */
/* ---------------------------------------------------------------------- */

#include "chip_cfg.h"

#define UART_FIFO   (*(volatile uint32_t*)(UART0_BASE + 0x00))
#define UART_CLKDIV (*(volatile uint32_t*)(UART0_BASE + 0x14))
#define UART_STATUS (*(volatile uint32_t*)(UART0_BASE + 0x1C))
#define UART_TXCNT  (((UART_STATUS) >> UART_TXCNT_SHIFT) & UART_TXCNT_MASK)
#define UART_RXCNT  (((UART_STATUS) >> UART_RXCNT_SHIFT) & UART_RXCNT_MASK)

#define GPIO_OUT         (*(volatile uint32_t*)(GPIO0_BASE + GPIO_OUT_OFF))
#define GPIO_OUT_W1TS    (*(volatile uint32_t*)(GPIO0_BASE + GPIO_OUT_W1TS_OFF))
#define GPIO_OUT_W1TC    (*(volatile uint32_t*)(GPIO0_BASE + GPIO_OUT_W1TC_OFF))
#define GPIO_ENABLE_W1TS (*(volatile uint32_t*)(GPIO0_BASE + GPIO_ENABLE_W1TS_OFF))
#define GPIO_ENABLE_W1TC (*(volatile uint32_t*)(GPIO0_BASE + GPIO_ENABLE_W1TC_OFF))
#define GPIO_IN          (*(volatile uint32_t*)(GPIO0_BASE + GPIO_IN_OFF))

static void uart_init(void) {
    UART_CLKDIV = UART_CLKDIV_VAL;
}

static void uart_putc(uint8_t c) {
    /* STATUS 的 TXFIFO_CNT：非 0 即等，留足写单字节的空间 */
    while (UART_TXCNT != 0) {
    }
    UART_FIFO = c;
}

static void uart_puts(const char* s) {
    while (*s) {
        uart_putc((uint8_t)*s++);
    }
}

static void uart_drain(void) {
    while (UART_TXCNT != 0) {
    } /* 等 TX FIFO 排空 */
}

/* ---- std.uart 原语（designs/0003 §5.1）：M1 只有 UART0（stdout 同路）---- */

void aila_rt_uart_init(uint32_t id, uint32_t baud) {
    if (id != 0) {
        RT_TRAP("uart: only UART0 exists");
    }
    if (baud == 0) {
        RT_TRAP("uart: baud is 0");
    }
    uint32_t div16 = (uint32_t)(((uint64_t)UART_CLK_HZ << 4) / baud);
    uint32_t val   = div16 >> 4;
    if (UART_CLKDIV_FRAC_BITS > 0) { /* esp32：FRAG[23:20] 4 位小数 */
        val |= (div16 & 0xF) << 20;
    }
    UART_CLKDIV = val;
}

uint32_t aila_rt_uart_write(uint32_t id, const uint8_t* p, uint32_t n) {
    if (id != 0) {
        RT_TRAP("uart: only UART0 exists");
    }
    for (uint32_t i = 0; i < n; i++) {
        uart_putc(p[i]);
    }
    return n;
}

uint32_t aila_rt_uart_read(uint32_t id, uint8_t* p, uint32_t cap) {
    if (id != 0) {
        RT_TRAP("uart: only UART0 exists");
    }
    uint32_t avail = UART_RXCNT;
    if (avail > cap) {
        avail = cap;
    }
    for (uint32_t i = 0; i < avail; i++) {
        p[i] = (uint8_t)UART_FIFO;
    }
    return avail;
}

/* ---- std.gpio 原语（designs/0003 §5.1）：M1 只做 GPIO 块主寄存器 ----
 * 位域 0-31（引脚 32-39 走 OUT1/ENABLE1 组寄存器，M2）；越界 trap
 * （防 1u<<pin 的 C 移位 UB）。read() 前置：pad 已配 GPIO 功能且输入使能
 * （esp32 的 IO_MUX/FUN_IE 属 pad 管理，M2）。 */

void aila_rt_gpio_mode(uint32_t pin, uint32_t m) {
    if (pin >= GPIO_PIN_COUNT) {
        RT_TRAP("gpio: pin out of range");
    }
    uint32_t bit = 1u << pin;
    if (m != 0) {
        GPIO_ENABLE_W1TS = bit; /* 输出使能 */
    } else {
        GPIO_ENABLE_W1TC = bit; /* 输入方向 */
    }
}

void aila_rt_gpio_write(uint32_t pin, uint32_t v) {
    if (pin >= GPIO_PIN_COUNT) {
        RT_TRAP("gpio: pin out of range");
    }
    uint32_t bit = 1u << pin;
    if (v != 0) {
        GPIO_OUT_W1TS = bit; /* 原子置位，不读改写 */
    } else {
        GPIO_OUT_W1TC = bit;
    }
}

uint32_t aila_rt_gpio_read(uint32_t pin) {
    if (pin >= GPIO_PIN_COUNT) {
        RT_TRAP("gpio: pin out of range");
    }
    return (GPIO_IN >> pin) & 1u;
}

/* ---------------------------------------------------------------------- */
/* 堆与区域账本：DRAM 里 bump，栈从 DRAM 顶向下（边界由链接脚本定）       */
/* ---------------------------------------------------------------------- */

extern uint8_t _bss_end[];   /* 链接脚本提供：bump 起点 */
extern uint8_t _heap_end[];  /* 链接脚本提供：bump 上界（栈之下） */

static uint8_t* heap_ptr;

typedef struct AilaRegion {
    struct AilaRegion* prev;
    uint8_t* mark;   /* enter 时刻的 bump 快照；leave 复位到此 */
    uint64_t used;   /* 归属配额计量（designs/0001 §5.6） */
    uint64_t cap;    /* 0 = 无上限 */
} AilaRegion;

static AilaRegion* g_current_region;
static AilaRegion* g_static_region;

static void* bump_alloc(uint64_t size, uint64_t align) {
    if (align == 0) {
        align = 8;
    }
    uintptr_t base    = (uintptr_t)heap_ptr;
    uintptr_t aligned = (base + (uintptr_t)(align - 1)) & ~(uintptr_t)(align - 1);
    uint8_t*  next    = (uint8_t*)(uintptr_t)(aligned + (uintptr_t)size);
    if (next > _heap_end) {
        RT_TRAP("region: physical memory exhausted");
    }
    heap_ptr = next;
    uint8_t* out = (uint8_t*)aligned;
    for (uint64_t i = 0; i < size; i++) {
        out[i] = 0;
    }
    return out;
}

void* aila_rt_region_enter(int64_t cap_hint) {
    uint8_t* mark = heap_ptr;  /* 先取快照：region 头自身也由 bump 出，leave 一起回收 */
    AilaRegion* r = (AilaRegion*)bump_alloc(sizeof(AilaRegion), 8);
    r->prev = g_current_region;
    r->mark = mark;
    r->used = 0;
    r->cap  = (cap_hint > 0) ? (uint64_t)cap_hint : 0;
    g_current_region = r;
    return (void*)r;
}

void* aila_rt_region_alloc(void* region, int64_t size, int64_t align) {
    AilaRegion* r = (AilaRegion*)(region ? region : (void*)g_current_region);
    if (!r) {
        RT_TRAP("allocation without an active region");
    }
    if (align <= 0) {
        align = 8;
    }
    uint64_t need = (size > 0) ? (uint64_t)size : 1;
    if (r->cap && r->used + need > r->cap) {
        RT_TRAP("region quota exceeded");
    }
    r->used += need;
    return bump_alloc(need, (uint64_t)align);
}

void* aila_rt_static_region(void) {
    if (!g_static_region) {
        g_static_region = (AilaRegion*)bump_alloc(sizeof(AilaRegion), 8);
        g_static_region->prev = NULL;
        g_static_region->mark = heap_ptr;  /* 永不 leave */
        g_static_region->used = 0;
        g_static_region->cap  = 0;
    }
    return (void*)g_static_region;
}

void aila_rt_region_leave(void* region) {
    AilaRegion* r = (AilaRegion*)region;
    if (!r) {
        return;
    }
    heap_ptr = r->mark;              /* O(1)：整个区域连头一起回收 */
    g_current_region = r->prev;
}

/* ---------------------------------------------------------------------- */
/* trap 与标准输出                                                        */
/* ---------------------------------------------------------------------- */

AILA_NORETURN void aila_rt_trap(const char* msg, int64_t len, const char* file, int64_t line) {
    uart_puts("\nAILA trap: ");
    if (msg && len > 0) {
        for (int64_t i = 0; i < len; i++) {
            uart_putc((uint8_t)msg[i]);
        }
    } else {
        uart_puts("unknown");
    }
    uart_puts("\n  at ");
    if (file) {
        uart_puts(file);
    }
    uart_putc(':');
    uint64_t v = (line < 0) ? 0 : (uint64_t)line;
    char digits[20];
    int  n = 0;
    do {
        digits[n++] = (char)('0' + (v % 10));
        v /= 10;
    } while (v);
    while (n > 0) {
        uart_putc((uint8_t)digits[--n]);
    }
    uart_putc('\n');
    uart_drain();
    for (;;) {
    } /* 嵌入式没有 exit：停在这里，等看门狗或手动复位 */
}

void aila_rt_write(const uint8_t* p, int32_t n) {  /* n = USize @esp32（32 位） */
    if (p && n > 0) {
        (void)aila_rt_uart_write(0, p, (uint32_t)n); /* 委托 std.uart 写路径 */
    }
}

void aila_rt_flush(void) { uart_drain(); }

/* ---------------------------------------------------------------------- */
/* 时间与随机种子（std.time / std.rand 的原语，designs/0003 §4.1）         */
/*                                                                        */
/* ccount 累积法：Xtensa ccount 每周期 +1（复位默认 240MHz），32 位约      */
/* 17.9s 回绕。用「无符号差值累加」自动处理回绕——前提是相邻调用间隔       */
/* < 半周期（约 8.9s），超过就会丢一段计数。                               */
/* ---------------------------------------------------------------------- */

static inline uint32_t rt_ccount(void) {
    uint32_t v;
    __asm__ volatile("rsr.ccount %0" : "=r"(v));
    return v;
}

uint64_t aila_rt_time_ms(void) {
    static uint32_t last = 0;
    static uint64_t cycles = 0; /* 自启动累积的周期数 */
    uint32_t now = rt_ccount();
    cycles += (uint64_t)(uint32_t)(now - last);
    last = now;
    return cycles / (AILA_CPU_HZ / 1000u);
}

void aila_rt_sleep_ms(uint64_t ms) {
    uint64_t need = ms * (AILA_CPU_HZ / 1000u);
    uint32_t last = rt_ccount();
    uint64_t waited = 0;
    while (waited < need) { /* 忙等：M1 无调度器，内层差值天然远小于半周期 */
        uint32_t now = rt_ccount();
        waited += (uint64_t)(uint32_t)(now - last);
        last = now;
    }
}
/* aila_rt_delay_us：微秒级忙等（I2C 位时序 / 单总线协议用）。
 * ccount 每周期 +1，换算基准 AILA_CPU_HZ 来自 chip.json（esp32 系复位默认 240MHz）。
 * 忙等不关中断：中断会把单次延时拉长，位时序敏感的整段事务应放进一次 C 调用
 *（designs/0006 §4.1）。 */
void aila_rt_delay_us(uint32_t us) {
    uint32_t need = us * (AILA_CPU_HZ / 1000000u);
    if (need == 0) {
        return;
    }
    uint32_t last = rt_ccount();
    uint32_t waited = 0;
    while (waited < need) {
        uint32_t now = rt_ccount();
        waited += (uint32_t)(now - last);
        last = now;
    }
}

uint64_t aila_rt_seed(void) {
    uint32_t a = rt_ccount();
    uint32_t b = rt_ccount();
    return 0x9E3779B97F4A7C15ull ^ ((uint64_t)a << 32) ^ (uint64_t)b;
}

/* ---------------------------------------------------------------------- */
/* 入口：startup.S → aila_rt_main → AILA 生成的 main                      */
/* ---------------------------------------------------------------------- */

extern int main(void);  /* AILA 的 func main 直接映射为目标入口符号 */

void aila_rt_main(void) {
    uart_init();
    heap_ptr = _bss_end;  /* bump_alloc 自己做对齐 */
    main();
    for (;;) {
    }
}
