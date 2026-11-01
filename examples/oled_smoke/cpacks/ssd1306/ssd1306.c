// cpacks/ssd1306/ssd1306.c —— 4 针 I2C OLED（SSD1306 128x64）的 ESP8266 薄边界层
//
// 分工（designs/0006 §2.2 的 L6）：AILA 负责「画什么」（std.pixbuf 点阵画布 +
// std.paint 图元 + std.textgrid 排版 + std.damage 脏区），这里只负责
// **「怎么发到线上去」**：软件 I2C 时序、SSD1306 页寻址、行主序→页格式的位搬运。
//
// 三条实测约束（designs/0006 §4.1）：
//   1. 画布是**行主序**（stride 对齐、MSB 在左），SSD1306 是**页格式**（一字节 = 8 竖像素）
//      → 本文件逐列按位现搬，不做整块转置缓冲（省 1KB RAM；1024 次位运算是几十 µs，
//      远小于整屏 I2C 的 37ms）。
//   2. SSD1306 忽略 0x21/0x22 水平寻址（designs/0005 §6）→ 一律页寻址：
//      每页先发 0xB0+page + 列地址低/高 4 位，再发该范围的列数据。
//   3. `[U8]` 视图跨 C 边界只传指针（实测 oled_abi(fb,7)==7）→ 帧缓冲指针开机只绑一次
//      （oled_bind），之后刷新只传标量。
//
// 为什么不用硬件 I2C：ESP8266 的 SDK I2C 要 SDK 头与库，而 cpack 走「手抄寄存器、零 include」
// 的同一条路（runtime/esp8266/）。软件 I2C 只要两根 GPIO + ets_delay_us。
//
// 寄存器地址来自 ESP8266 TRM 的 GPIO 块（0x60000300）：
//   +0x00 OUT  +0x04 OUT_W1TS  +0x08 OUT_W1TC
//   +0x0C ENABLE  +0x10 ENABLE_W1TS  +0x14 ENABLE_W1TC  +0x18 IN

#define GPIO0_BASE 0x60000300u

#define GPIO_OUT_W1TS (*(volatile unsigned int *)(GPIO0_BASE + 0x04))
#define GPIO_OUT_W1TC (*(volatile unsigned int *)(GPIO0_BASE + 0x08))
#define GPIO_ENABLE_W1TS (*(volatile unsigned int *)(GPIO0_BASE + 0x10))
#define GPIO_ENABLE_W1TC (*(volatile unsigned int *)(GPIO0_BASE + 0x14))
#define GPIO_IN (*(volatile unsigned int *)(GPIO0_BASE + 0x18))

extern void ets_delay_us(unsigned int us);

#define OLED_ADDR 0x3C   /* 常见 4 针模块的 I2C 地址（有的板子是 0x3D） */
#define OLED_ADDR_ALT 0x3D
/* 列地址偏移：SSD1306（128 列）为 0；SH1106（132 列，可视窗口从第 2 列起）为 2 */
#define OLED_COL_OFFSET 0u
#define I2C_DELAY_US 2   /* 半周期 ≈ 2µs → SCL ≈ 250kHz，SSD1306（≤400kHz）稳 */
#define PANEL_W 128u
#define PANEL_PAGES 8u

static unsigned int g_scl = 5; /* 默认 D1 = GPIO5 */
static unsigned int g_sda = 4; /* 默认 D2 = GPIO4 */
static unsigned int g_addr = OLED_ADDR;
static unsigned int g_nack = 0;   /* 累计失败事务数（诊断：>0 说明屏幕没收到东西） */
static unsigned int g_bytes = 0;  /* 累计发送的数据字节数（诊断：看局部刷新省了多少） */

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

/* 发一个字节并读回 ACK（1 = 器件应答）。ACK 位把 SDA 交还从机，所以要切输入。 */
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

/* 一次事务：控制字节（0x00 = 命令流 / 0x40 = 数据流）+ n 字节 */
static unsigned int oled_write(unsigned char ctrl, const unsigned char *p, unsigned long n) {
    unsigned long i;
    unsigned int ok;
    i2c_start();
    ok = i2c_write_byte((unsigned char)(g_addr << 1)); /* 地址 + 写位 */
    if (ok) {
        ok = i2c_write_byte(ctrl);
    }
    for (i = 0; ok && i < n; i++) {
        ok = i2c_write_byte(p[i]);
    }
    i2c_stop();
    if (!ok) {
        g_nack++; /* 事务失败：屏幕上不会出现这次数据（诊断计数） */
    } else {
        g_bytes += (unsigned int)n;
    }
    return ok;
}

/* 总线恢复：从机在事务中途掉电/丢位时会把 SDA 一直拉低，整个总线就死了，
   之后每次地址都没有 ACK、屏幕永远停在最后一帧。标准急救法是：主控放开 SDA，
   手动打 9 个 SCL 时钟把从机移位寄存器里的半个字节顶出去，再补一个 STOP。 */
static void i2c_recover(void) {
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

/* oled_init(scl, sda)：配两线为输出、拉低，返回 1 = 屏在线（地址有应答） */
unsigned int oled_init(unsigned int scl, unsigned int sda) {
    unsigned int ok;
    g_scl = scl;
    g_sda = sda;
    GPIO_ENABLE_W1TS = (1u << g_scl) | (1u << g_sda);
    scl_lo();
    sda_lo();
    ets_delay_us(10000); /* 上电稳定 */

    i2c_recover(); /* 先救总线：上次运行/复位可能把 SDA 留在低电平 */

    /* 探地址：先 0x3C，没应答应就试 0x3D（两种常见模块） */
    g_addr = OLED_ADDR;
    i2c_start();
    ok = i2c_write_byte((unsigned char)(g_addr << 1));
    i2c_stop();
    if (!ok) {
        g_addr = OLED_ADDR_ALT;
        i2c_start();
        ok = i2c_write_byte((unsigned char)(g_addr << 1));
        i2c_stop();
    }
    if (!ok) {
        g_nack++;
        return 0;
    }
    return 1;
}

void oled_cmd(unsigned char b) { oled_write(0x00, &b, 1); }

/* oled_nack()：累计失败事务数——AILA 侧每秒打印一次，>0 即「写进单片机了但没写进屏幕」 */
unsigned int oled_nack(void) { return g_nack; }
/* oled_bytes()：累计发送的数据字节数（局部刷新省了多少流量，看这个） */
unsigned int oled_bytes(void) { return g_bytes; }

/* ---- 画布绑定：行主序 canvas，开机绑一次 ---- */

static const unsigned char *g_fb = 0;
static unsigned int g_stride = PANEL_W / 8u; /* 128 宽、行对齐 4 → 16 字节/行 */

void oled_bind(const unsigned char *fb) { g_fb = fb; }
void oled_stride(unsigned int bytes) { g_stride = bytes; }

/* ABI 探针：AILA 侧 `oled_abi(fb, 7)`。返回 7 = 视图只传指针；返回 1024 = 视图传了 (ptr,len) */
unsigned int oled_abi(const unsigned char *fb, unsigned int a) {
    (void)fb;
    return a;
}

/* page_col_byte 把行主序画布里 (col, page) 那一列的 8 个竖像素压成一个字节。
   行主序：像素 (x, y) 在 fb[y*stride + x/8] 的 bit (0x80 >> (x%8))；
   页格式：bit k = 该页第 k 行（y = page*8 + k），bit0 在最上。 */
static unsigned char page_col_byte(unsigned int col, unsigned int page) {
    unsigned int k;
    unsigned char out = 0;
    unsigned int byte_off = col >> 3;
    unsigned char mask = (unsigned char)(0x80u >> (col & 7u));
    const unsigned char *row = g_fb + (unsigned long)page * 8u * g_stride;
    for (k = 0; k < 8u; k++) {
        if (row[(unsigned long)k * g_stride + byte_off] & mask) {
            out |= (unsigned char)(1u << k);
        }
    }
    return out;
}

static unsigned char g_line[PANEL_W];

/* oled_present(p0, p1, c0, c1)：只刷 [p0,p1) 页 × [c0,c1) 列（半开区间）。
   AILA 侧的 commit.damage 换算出来的就是这四个数——整屏刷新就是 (0,8,0,128)。
   每页一次事务：命令（页起始 + 列窗口）+ 数据（该窗口的列字节）。 */
void oled_present(unsigned int p0, unsigned int p1, unsigned int c0, unsigned int c1) {
    unsigned int page;
    unsigned int col;
    unsigned int n;
    unsigned char win[3];
    if (g_fb == 0) {
        return;
    }
    if (p1 > PANEL_PAGES) {
        p1 = PANEL_PAGES;
    }
    if (c1 > PANEL_W) {
        c1 = PANEL_W;
    }
    if (p0 >= p1 || c0 >= c1) {
        return;
    }
    n = c1 - c0;
    for (page = p0; page < p1; page++) {
        win[0] = (unsigned char)(0xB0u | (page & 0x07u));
        win[1] = (unsigned char)((c0 + OLED_COL_OFFSET) & 0x0Fu);
        win[2] = (unsigned char)(0x10u | (((c0 + OLED_COL_OFFSET) >> 4) & 0x07u));
        oled_write(0x00, win, 3);
        for (col = 0; col < n; col++) {
            g_line[col] = page_col_byte(c0 + col, page);
        }
        oled_write(0x40, g_line, n);
    }
}

/* ---- 调试打点（固件静默停住时判断停在哪一步）---- */

#define DBG_UART_FIFO   (*(volatile unsigned int *)(0x60000000u))
#define DBG_UART_STATUS (*(volatile unsigned int *)(0x6000001Cu))

/* 心跳灯：GPIO2（ESP-12 板上蓝灯，低电平点亮）。与 UART 完全独立的存活信号 */
void dbg_led_toggle(void) {
    static unsigned int inited;
    static unsigned int on;
    if (!inited) {
        GPIO_ENABLE_W1TS = 1u << 2;
        inited = 1;
    }
    if (on) {
        GPIO_OUT_W1TC = 1u << 2; /* 高电平 = 灭 */
        on = 0;
    } else {
        GPIO_OUT_W1TS = 1u << 2; /* 低电平 = 亮 */
        on = 1;
    }
}

void dbg_mark(unsigned int v) {
    unsigned int cnt = (DBG_UART_STATUS >> 16) & 0xFFu;
    dbg_led_toggle();
    if (cnt < 120u) {
        DBG_UART_FIFO = (unsigned int)('A' + (v % 26u));
    }
}
