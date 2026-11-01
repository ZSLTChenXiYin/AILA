/* AILA ESP8266 平台运行时（freestanding，无 libc）。
 *
 * 对应 runtime/aila_rt.c 的宿主实现，七个入口符号与语义不变：
 *   - 区域账本：O(1) bump + mark 复位——region_enter 记 bump 快照，
 *     region_leave 把 bump 指针退回快照，整个区域一次回收（designs/0001 §12.3）
 *   - trap / write / flush：走 UART0
 *
 * ABI 注意（必须与代码生成一致）：
 *   - `aila_rt_trap` / `aila_rt_region_*` 的标量参数在代码生成里固定为 i64
 *     （internal/codegen：ensureDeclare 的硬编码签名），C 侧用 int64_t；
 *   - `aila_rt_write` 的 n 是 USize——esp8266 指针宽 4 字节，按 32 位传参。
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
/* UART：FIFO@0x00、CLKDIV@0x14、STATUS@0x1C（TX [23:16] / RX [7:0]，      */
/*   NonOS SDK uart_register.h）；GPIO 块偏移见 esp8266 chip.json _source  */
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

/* UART 分频必须按「当前的 CPU 时钟」算：ESP8266 上电停在晶振 52MHz，
   80MHz 是软件切上去的（SDK 会切，裸机运行时不切）。chip_cfg.h 的 clk_hz
   是裸机实际值，SDK 构建由 nosdk_cfg.h 覆盖成 80MHz。 */
#ifndef AILA_UART_CLK_HZ
#define AILA_UART_CLK_HZ UART_CLK_HZ
#endif

static void uart_init(void) {
    UART_CLKDIV = (uint32_t)(((uint64_t)AILA_UART_CLK_HZ + UART_BAUD_HZ / 2) / UART_BAUD_HZ);
}

static void uart_putc(uint8_t c) {
    /* 只在 FIFO 将满时等：原写法「TXFIFO_CNT != 0 就等」等于每个字节都等整条队列排空，
       慢且在实机上会把整机挂死在打印里。而且等待必须有上限——UART 状态若读出异常值，
       无上限自旋就是不可恢复的死机（丢了字符还能看见后面的日志，挂死就什么都看不见）。 */
    uint32_t guard = 0;
    while (UART_TXCNT > 120u && guard < 200000u) {
        guard++;
    }
    UART_FIFO = c;
}

static void uart_puts(const char* s) {
    while (*s) {
        uart_putc((uint8_t)*s++);
    }
}

static void uart_drain(void) {
    uint32_t guard = 0;
    while (UART_TXCNT != 0 && guard < 2000000u) {
        guard++;
    } /* 等 TX FIFO 排空（有上限，绝不挂死） */
}

/* ---- std.uart 原语（designs/0003 §5.1）：M1 只有 UART0（stdout 同路）---- */

void aila_rt_uart_init(uint32_t id, uint32_t baud) {
    if (id != 0) {
        RT_TRAP("uart: only UART0 exists");
    }
    if (baud == 0) {
        RT_TRAP("uart: baud is 0");
    }
    uint32_t div16 = (uint32_t)(((uint64_t)AILA_UART_CLK_HZ << 4) / baud);
    uint32_t val   = div16 >> 4;
    if (UART_CLKDIV_FRAC_BITS > 0) { /* esp8266 恒为 0：寄存器无小数域 */
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
 * 引脚 0-15（GPIO16 属 RTC 域）；越界 trap（防 1u<<pin 的 C 移位 UB）。 */

void aila_rt_gpio_mode(uint32_t pin, uint32_t m) {
    if (pin >= GPIO_PIN_COUNT) {
        RT_TRAP("gpio: pin out of range");
    }
    uint32_t bit = 1u << pin;
    if (m != 0) {
        GPIO_ENABLE_W1TS = bit; /* 输出使能 */
    } else {
        GPIO_ENABLE_W1TC = bit; /* 输入（esp8266 pad 输入恒使能，无需 IO_MUX） */
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

void aila_rt_write(const uint8_t* p, int32_t n) {  /* n = USize @esp8266（32 位） */
    if (p && n > 0) {
        (void)aila_rt_uart_write(0, p, (uint32_t)n); /* 委托 std.uart 写路径 */
    }
}

void aila_rt_flush(void) { uart_drain(); }

/* ---------------------------------------------------------------------- */
/* 时间与随机种子（std.time / std.rand 的原语，designs/0003 §4.1）         */
/*                                                                        */
/* ccount 累积法：Xtensa ccount 每周期 +1（80MHz），32 位约 53.7s 回绕。   */
/* 用「无符号差值累加」自动处理回绕——前提是相邻调用间隔 < 半周期（约 26s）*/
/* 超过就会丢一段计数（长睡眠请用 aila_rt_sleep_ms 的循环路径）。          */
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

uint64_t aila_rt_seed(void) {
    uint32_t a = rt_ccount();
    uint32_t b = rt_ccount();
    return 0x9E3779B97F4A7C15ull ^ ((uint64_t)a << 32) ^ (uint64_t)b;
}

/* ---------------------------------------------------------------------- */
/* 入口：startup.S → aila_rt_main → AILA 生成的 main                      */
/* ---------------------------------------------------------------------- */

extern int main(void);  /* AILA 的 func main 直接映射为目标入口符号 */

/* .bss 是 NOLOAD 段：ROM 只把已初始化的段搬进 DRAM，零初始化区里是上电残留。
   freestanding 没有 C 运行时的 crt 来做这件事，必须自己按链接脚本给的边界清零。
   `_bss_end` 已在上方以 uint8_t[] 声明（bump 堆起点用），这里只补 `_bss_start`。 */
extern uint8_t _bss_start[];

static void rt_clear_bss(void) {
    uintptr_t p = (uintptr_t)_bss_start;
    uintptr_t end = (uintptr_t)_bss_end;
    while (p < end) {
        *(volatile uint32_t*)p = 0u;
        p += 4u;
    }
}

/* freestanding 镜像没有装任何 ISR：ROM 留下的定时器/Wi-Fi 中断一旦触发，
   就会跳到未初始化的中断向量 → CPU 跑飞、表现为「随机静默挂死」（实测：
   同一份固件有时跑 25s，有时 3s 就停，且停在不同语句）。所以进 main 之前
   先把 PS.INTLEVEL 抬到 15，屏蔽全部可屏蔽中断。rsil 会返回旧 PS，将来要开中断照抄回去即可。 */
static unsigned int rt_lock_interrupts(void) {
    unsigned int old_ps;
    __asm__ volatile("rsil %0, 15" : "=r"(old_ps));
    return old_ps;
}

void aila_rt_main(void) {
    rt_lock_interrupts(); /* 先关中断：没有 ISR 表的裸机镜像必须这样做 */
    rt_clear_bss();       /* 再清 .bss：之后所有零初始化的静态量才可信 */
    uart_init();
    heap_ptr = _bss_end;  /* bump_alloc 自己做对齐 */
    main();
    for (;;) {
    }
}
