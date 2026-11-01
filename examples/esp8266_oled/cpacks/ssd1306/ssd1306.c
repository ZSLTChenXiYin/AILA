// cpacks/ssd1306/ssd1306.c —— 4 针 I2C OLED（SSD1306 128x64）的 ESP8266 薄边界层
//
// 分工：AILA 负责「画什么」（字库、排版、帧缓冲），这里只负责「怎么发到线上去」
// ——软件 I2C 的时序与 SSD1306 的帧写协议都是微秒级硬实时，按 designs/0002 §1
// 的纯度三线放在 C 侧。
//
// 为什么不用硬件 I2C：ESP8266 的硬件 I2C（SDK i2c_master_*）要 SDK 头与库，
// 而 cpack 走的是「手抄寄存器、零 include」的同一条路（见 runtime/esp8266/）。
// 软件 I2C 只要两根 GPIO + ets_delay_us，接线自由、四个引脚就是 VCC/GND/SCL/SDA。
//
// 寄存器地址来自 ESP8266 技术参考手册的 GPIO 块（0x60000300）：
//   +0x00 OUT      +0x04 OUT_W1TS   +0x08 OUT_W1TC
//   +0x0C ENABLE   +0x10 ENABLE_W1TS +0x14 ENABLE_W1TC   +0x18 IN
//
// ets_delay_us 是 ROM 里的函数（与 runtime/esp8266/nosdk.c 同一用法）。

#define GPIO0_BASE 0x60000300u

#define GPIO_OUT_W1TS (*(volatile unsigned int *)(GPIO0_BASE + 0x04))
#define GPIO_OUT_W1TC (*(volatile unsigned int *)(GPIO0_BASE + 0x08))
#define GPIO_ENABLE_W1TS (*(volatile unsigned int *)(GPIO0_BASE + 0x10))
#define GPIO_ENABLE_W1TC (*(volatile unsigned int *)(GPIO0_BASE + 0x14))
#define GPIO_IN (*(volatile unsigned int *)(GPIO0_BASE + 0x18))

extern void ets_delay_us(unsigned int us);

#define OLED_ADDR 0x3C   /* 常见 4 针模块的 I2C 地址（有的板子是 0x3D） */
#define OLED_ADDR_ALT 0x3D
/* 列地址偏移：SSD1306（128 列）为 0；SH1106（132 列，可视窗口从第 2 列起）为 2。
   先按 0 跑；若整幅图左移/右侧少 2 列，就改成 2。 */
#define OLED_COL_OFFSET 0u
#define I2C_DELAY_US 2   /* 半周期 ≈ 2µs → SCL ≈ 250kHz，SSD1306（≤400kHz）稳 */

static unsigned int g_scl = 5; /* 默认 D1 = GPIO5 */
static unsigned int g_sda = 4; /* 默认 D2 = GPIO4 */
static unsigned int g_addr = OLED_ADDR;
static unsigned int g_nack = 0; /* 累计失败事务数（诊断：>0 说明屏幕没收到东西） */

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

/* oled_cmd(b)：单条命令 */
/* oled_nack()：累计失败事务数——AILA 侧每秒打印一次，>0 即「写进单片机了但没写进屏幕」 */
unsigned int oled_nack(void) { return g_nack; }

void oled_cmd(unsigned char b) { oled_write(0x00, &b, 1); }

/* 帧缓冲指针只在开机时绑一次，后续刷新只传标量。
   理由：`[U8]` 视图跨 C 边界时究竟是 (ptr) 还是 (ptr,len) 两个机器字，取决于代码生成，
   一旦 C 侧形参形状猜错，后面的标量实参就会错位（page 收到 len）。只取**第一个**参数
   （两套 ABI 里都是指针）就彻底绕开这个坑。 */
static const unsigned char *g_fb = 0;

void oled_bind(const unsigned char *fb) { g_fb = fb; }

/* ABI 探针：AILA 侧 `oled_abi(fb, 7)`。返回 7 = 视图只传指针；返回 1024 = 视图传了 (ptr,len) */
unsigned int oled_abi(const unsigned char *fb, unsigned int a) {
    (void)fb;
    return a;
}

/* oled_flush_page：刷一页（128 字节），老派页寻址（0xB0+page + 列地址低位/高位）。
   面板若是 132 列（SH1106 一类），必须把列地址偏移 2；SSD1306 则是 0。 */
void oled_flush_page(unsigned char page) {
    unsigned char win[3];
    if (g_fb == 0) {
        return;
    }
    win[0] = (unsigned char)(0xB0 | (page & 0x07));
    win[1] = (unsigned char)(OLED_COL_OFFSET & 0x0F);        /* 列地址低 4 位 */
    win[2] = (unsigned char)(0x10 | ((OLED_COL_OFFSET >> 4) & 0x07)); /* 列地址高 4 位 */
    oled_write(0x00, win, 3);
    oled_write(0x40, g_fb + (unsigned long)page * 128u, 128);
}

/* oled_flush_all：整屏 8 页顺序刷（页寻址，不依赖 0x21/0x22 水平寻址指令） */
void oled_flush_all(void) {
    unsigned char q;
    for (q = 0; q < 8u; q++) {
        oled_flush_page(q);
    }
}