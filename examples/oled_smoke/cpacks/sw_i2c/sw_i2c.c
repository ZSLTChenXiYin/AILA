// cpacks/sw_i2c/sw_i2c.c —— 软件 I2C 传输层（与具体器件无关）
//
// 边界（designs/0006 §2.2 的 L6 / §4.1）：这里**只做「把字节发到两根线上」**——
// GPIO 位时序、START/STOP、字节 + ACK、总线恢复。SSD1306 的命令序列、页寻址窗口、
// 行主序→页格式的搬运全部在 AILA 侧（`_.oled.panel` + `std.paint.transpose_page`）。
//
// 为什么位时序留 C：半周期 2µs、裸机 esp8266 主频 52MHz → 一个半周期只有 ≈104 个周期。
// AILA 里一次 GPIO 写要跨模块调用（铁律 7 不允许跨模块内联）+ 没有关中断原语，
// 每 bit 的调用开销和 ISR 抖动都会吃掉这个预算（designs/0006 §4.1 第 13 条）。
//
// 为什么不用硬件 I2C：ESP8266 的硬件 I2C 要 SDK 头与库，而 cpack 走「手抄寄存器、
// 零 include」的同一条路（runtime/<家族>/ 同款）。软件 I2C 只要两根 GPIO。
//
// 寄存器地址来自 ESP8266 TRM 的 GPIO 块（0x60000300）：
//   +0x00 OUT  +0x04 OUT_W1TS  +0x08 OUT_W1TC
//   +0x0C ENABLE  +0x10 ENABLE_W1TS  +0x14 ENABLE_W1TC  +0x18 IN
//
// ABI 约定（designs/0006 §4.1 第 2 条）：`[U8]` 跨 C 边界**只传指针**，长度必须另开
// 标量形参（这里是 `unsigned int n`）——AILA 侧的手写声明见 `oled/panel.aila`。

#define GPIO0_BASE 0x60000300u

#define GPIO_OUT_W1TS (*(volatile unsigned int *)(GPIO0_BASE + 0x04))
#define GPIO_OUT_W1TC (*(volatile unsigned int *)(GPIO0_BASE + 0x08))
#define GPIO_ENABLE_W1TS (*(volatile unsigned int *)(GPIO0_BASE + 0x10))
#define GPIO_ENABLE_W1TC (*(volatile unsigned int *)(GPIO0_BASE + 0x14))
#define GPIO_IN (*(volatile unsigned int *)(GPIO0_BASE + 0x18))

extern void ets_delay_us(unsigned int us);

/* 半周期 ≈ 2µs → SCL ≈ 250kHz（SSD1306 / 多数 4 针模块上限 400kHz，留余量） */
#define I2C_DELAY_US 2

static unsigned int g_scl = 5; /* 默认 D1 = GPIO5 */
static unsigned int g_sda = 4; /* 默认 D2 = GPIO4 */
static unsigned int g_addr = 0x3C;
static unsigned int g_nack = 0;  /* 累计失败事务数（诊断：>0 说明器件没收到东西） */
static unsigned int g_sent = 0;  /* 累计发送的数据字节数 */

static void scl_hi(void) { GPIO_OUT_W1TS = 1u << g_scl; }
static void scl_lo(void) { GPIO_OUT_W1TC = 1u << g_scl; }
static void sda_hi(void) { GPIO_OUT_W1TS = 1u << g_sda; }
static void sda_lo(void) { GPIO_OUT_W1TC = 1u << g_sda; }

static void sda_out(void) { GPIO_ENABLE_W1TS = 1u << g_sda; }
static void sda_in(void) { GPIO_ENABLE_W1TC = 1u << g_sda; }
static unsigned int sda_get(void) { return (GPIO_IN >> g_sda) & 1u; }

static void i2c_start(void) {
    sda_out();
    sda_hi();
    scl_hi();
    ets_delay_us(I2C_DELAY_US);
    sda_lo(); /* SCL 高时 SDA 下降 = START */
    ets_delay_us(I2C_DELAY_US);
    scl_lo();
    ets_delay_us(I2C_DELAY_US);
}

static void i2c_stop(void) {
    sda_out();
    sda_lo();
    ets_delay_us(I2C_DELAY_US);
    scl_hi();
    ets_delay_us(I2C_DELAY_US);
    sda_hi(); /* SCL 高时 SDA 上升 = STOP */
    ets_delay_us(I2C_DELAY_US);
}

/* 发一个字节并读回 ACK（1 = 从机应答）。ACK 位把 SDA 交还从机，所以要切输入。 */
static unsigned int i2c_write_byte(unsigned char b) {
    unsigned int i;
    unsigned int ack;
    sda_out();
    for (i = 0; i < 8; i++) {
        if (b & 0x80u) {
            sda_hi();
        } else {
            sda_lo();
        }
        b = (unsigned char)(b << 1);
        ets_delay_us(I2C_DELAY_US);
        scl_hi();
        ets_delay_us(I2C_DELAY_US);
        scl_lo();
        ets_delay_us(I2C_DELAY_US);
    }
    sda_in(); /* 让从机拉低 ACK */
    ets_delay_us(I2C_DELAY_US);
    scl_hi();
    ets_delay_us(I2C_DELAY_US);
    ack = sda_get() ? 0u : 1u;
    scl_lo();
    sda_out();
    ets_delay_us(I2C_DELAY_US);
    return ack;
}

/* 总线恢复：从机在事务中途掉电/丢位时会把 SDA 一直拉低，整条总线就死了
   （之后每次地址都没有 ACK、屏永远停在最后一帧）。标准急救法：主控放开 SDA，
   手动打 9 个 SCL 把从机移位寄存器里的半个字节顶出去，再补一个 STOP。 */
void sw_i2c_recover(void) {
    unsigned int k;
    sda_in(); /* 放开 SDA，交给上拉 */
    for (k = 0; k < 9; k++) {
        scl_hi();
        ets_delay_us(I2C_DELAY_US * 3);
        scl_lo();
        ets_delay_us(I2C_DELAY_US * 3);
    }
    i2c_stop();
    ets_delay_us(100);
}

/* sw_i2c_begin(scl, sda, addr7)：配两线为输出、上电稳定、救总线、探地址。
   返回 1 = 器件应答。addr7 是 7 位地址（SSD1306 常见 0x3C / 少数 0x3D）。 */
unsigned int sw_i2c_begin(unsigned int scl, unsigned int sda, unsigned int addr7) {
    unsigned int ok;
    g_scl = scl;
    g_sda = sda;
    g_addr = addr7;
    GPIO_ENABLE_W1TS = (1u << g_scl) | (1u << g_sda);
    scl_lo();
    sda_lo();
    ets_delay_us(10000); /* 上电稳定 */

    sw_i2c_recover();

    i2c_start();
    ok = i2c_write_byte((unsigned char)(g_addr << 1)); /* 地址 + 写位 */
    i2c_stop();
    if (!ok) {
        g_nack++;
    }
    return ok;
}

/* sw_i2c_write(ctrl, bytes, n)：一次完整事务 = 控制字节 + n 字节数据。
   控制字节是器件语义（SSD1306：0x00 = 命令流、0x40 = 数据流），本层当不透明数据。 */
int sw_i2c_write(unsigned int ctrl, const unsigned char *bytes, unsigned int n) {
    unsigned int i;
    unsigned int ok;
    i2c_start();
    ok = i2c_write_byte((unsigned char)(g_addr << 1));
    if (ok) {
        ok = i2c_write_byte((unsigned char)ctrl);
    }
    for (i = 0; ok && i < n; i++) {
        ok = i2c_write_byte(bytes[i]);
    }
    i2c_stop();
    if (!ok) {
        g_nack++;
        return -1;
    }
    g_sent += n;
    return (int)n;
}

/* 诊断计数：失败事务数 / 累计发送字节数（上层每秒打印，用来判断"写进单片机了
   但没写进器件"和"局部刷新到底省了多少流量"）。 */
unsigned int sw_i2c_nack(void) { return g_nack; }
unsigned int sw_i2c_sent(void) { return g_sent; }

/* 调试打点：直接写 UART0 FIFO，完全不经过 AILA 的 io.printf。
   用途：固件静默停住时靠这些字母判断停在哪一步；FIFO 满就丢弃，绝不阻塞。 */
#define DBG_UART_FIFO (*(volatile unsigned int *)(0x60000000u))
#define DBG_UART_STATUS (*(volatile unsigned int *)(0x6000001Cu))

void board_mark(unsigned int v) {
    unsigned int cnt = (DBG_UART_STATUS >> 16) & 0xFFu;
    if (cnt < 120u) {
        DBG_UART_FIFO = (unsigned int)('A' + (v % 26u));
    }
}
