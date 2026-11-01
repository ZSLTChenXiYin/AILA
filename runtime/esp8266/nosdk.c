/* AILA ESP8266 NonOS SDK 入口层（nosdk 模式）。
 *
 * SDK 启动链：ROM → call_user_start（libmain.a 提供，ld ENTRY）→ SDK 初始化 →
 *   user_rf_cal_sector_set → user_pre_init → user_init。
 * 三个回调都是 libmain 的未定义引用，v3.0 起 user_pre_init 必须注册分区表
 * （SDK 文档《Partition Table 说明》）。类型与常量按 include/user_interface.h
 * 手抄声明——不 include SDK 头，保持 freestanding 无 newlib 依赖。
 *
 * std.wifi 胶水（designs/0003 §5.2）：wifi_*（libnet80211）+ lwip raw TCP
 * （liblwip，tcp.h/pbuf.h 布局手抄，见 TCP 段注释）。
 *
 * 数据驱动：flash map 与分区表来自 chip.json 的 sdks 段，由 aila build
 * 生成 nosdk_cfg.h 注入；bump 堆窗口 [_bss_end, _heap_end) 由 SDK 变体
 * 链接脚本定义（_heap_end = _aila_heap_end），aila_rt_esp8266.c 零改动复用。
 */

#include <stdint.h>

#include "chip_cfg.h"
#include "nosdk_cfg.h"

/* ---- SDK API（user_interface.h 手抄子集；bool 返回按 32 位读，int 兼容） ---- */

typedef struct {
    uint32_t type;  /* partition_type_t：RF_CAL=4 / PHY_DATA=5 / SYSTEM_PARAMETER=6，
                     * CUSTOM_BEGIN=100 之后用户自定义 */
    uint32_t addr;
    uint32_t size;
} partition_item_t;

int system_partition_table_regist(const partition_item_t* table, uint32_t n, uint32_t map);

/* aila_rt_esp8266.c 提供：uart_init + bump 堆复位（_bss_end 起）+ 调 main（不返回） */
void aila_rt_main(void);

/* 分区表数据来自 chip.json（aila build 生成） */
static const partition_item_t sdk_partitions[] = SDK_PARTITION_TABLE_DATA;

/* ---- 启动链打点：直写 UART0 FIFO，不依赖 SDK、也不依赖我们的 runtime ----
   用途：SDK 镜像一路静默时判断卡在哪一环
     R = 进 user_rf_cal_sector_set　　r = 它返回了
     P = 进 user_pre_init　　　　　　 p = 分区表注册成功　! = 注册失败
     U = 进 user_init（之后 aila_rt_main 会把分频改成 115200）
   ROM 默认分频对应 74880（26MHz 晶振），所以 R/r/P/p/U 要在 74880 上抓。 */
#define BOOT_UART_FIFO   (*(volatile uint32_t*)(0x60000000u))
#define BOOT_UART_STATUS (*(volatile uint32_t*)(0x6000001Cu))
#define BOOT_UART_CLKDIV (*(volatile uint32_t*)(0x60000014u))

static void boot_mark(char c) {
    uint32_t guard = 0;
    /* 每次打点先把分频按「SDK 已把 CPU 切到 80MHz」重设成 115200（80e6/115200 = 694），
       这样不管 SDK 之前把 UART 改成了什么，打点都稳定落在 115200 上 */
    BOOT_UART_CLKDIV = 694u;
    while (((BOOT_UART_STATUS >> 16) & 0xFFu) > 120u && guard < 200000u) {
        guard++; /* 有上限：FIFO 状态异常也不能把启动卡在打点上 */
    }
    BOOT_UART_FIFO = (uint32_t)(unsigned char)c;
}

/* ---- 启动路径：官方 bootloader 在 0x0，应用在 0x1000 ----
 *
 * ROM 冷启动只加载「flash 0x0 处镜像头声明的段」并跳到 entry，**不会**打开 flash cache。
 * 而 SDK 的代码绝大部分在 IROM（flash 映射 0x40200000+），必须在 cache 打开后才能取指。
 * 这件事由 SDK 自带的 bootloader 负责：boot_v1.x.bin（0x0）→ 读 0x1000 处镜像头 →
 * 配 SPI 读模式 + Cache_Read_Enable → 跳 call_user_start（libmain，此时 cache 已就绪）。
 *
 * 所以正确烧法是三片：
 *   0x00000  boot_v1.7.bin                    （SDK 自带 bin/）
 *   0x01000  app.bin                          （本工程构建产物，镜像内 irom 偏移 = 0xF000）
 *   0x3FC000 esp_init_data_default_v08.bin    （SDK 自带 bin/，RF 校准默认值）
 * 曾经走过的弯路（保留作记录）：把应用单片烧在 0x0，靠自己的入口调
 * SPIReadModeCnfig + Cache_Read_Enable 再跳 IROM —— 实测 cache 仍不可读，
 * 读 0x402475E0 直接 LoadStoreError，说明这套 cache 开启不能由应用侧代劳。
 */
/* ---- 当前进度（2026-11-01，SDK 路径）----
 * 已打通：镜像格式（ESP8266 image v2 / 0xEA，用 esptool elf2image --version 2 产出）、
 *   烧写布局（4MB 头 bootloader@0x0 + 应用@0x1000/user2@0x81000）、IRAM 段路由
 *   （app_main.o 的 call_user_start_local / Cache_Read_Enable_New / user_start 必须留 IRAM，
 *   见 internal/chip/chip.go 的 .iram0.boot）。
 * 实测：bootloader 正常 jump 后，应用入口确实执行（此前用入口打点验证过打印 'E'）。
 * 未打通：应用侧「打开 flash cache」这一步——SDK 的 ets_run 初始化链需要 cache 就绪才能执行
 *   flash 里的代码，而应用自己调 Cache_Read_Enable_New(0,0,1) 会挂死，ROM 的
 *   Cache_Read_Enable(0,0,1)/(0,4,1) + SPIReadModeCnfig(3) 组合也不成立。
 *   官方 AT 镜像在同一条 bootloader 下能跑，说明差异在它自己的 crt（入口 call0 到一个
 *   IRAM 函数完成 cache 初始化）。下一步就是对着它把这段再对齐。
 * ---- 以下为启动打点（保留复用）---- */
/* user_rf_pre_init：libmain 的弱符号，SDK 初始化最早期调用（比 user_pre_init 还早）。
   我们覆写它只为了打一个 'F'——用来判断到底是「SDK 初始化还没走到回调」还是
   「走到了回调但后面炸了」。返回 0 = 不指定，SDK 用默认值。 */
uint32_t user_rf_pre_init(void) {
    boot_mark('F');
    return 0;
}

/* RF 校准扇区：SDK 初始化早期调用，返回 RF_CAL 分区起始扇区号
 *（SDK 头里的返回类型写作 uint32，freestanding 下用等宽的 uint32_t） */uint32_t user_rf_cal_sector_set(void) {
    boot_mark('R');
    for (uint32_t i = 0; i < sizeof(sdk_partitions) / sizeof(sdk_partitions[0]); i++) {
        if (sdk_partitions[i].type == 4u) { /* SYSTEM_PARTITION_RF_CAL */
            boot_mark('r');
            return sdk_partitions[i].addr / 4096u;
        }
    }
    boot_mark('r');
    return 0;
}

void user_pre_init(void) {
    boot_mark('P');
    if (!system_partition_table_regist(sdk_partitions,
            sizeof(sdk_partitions) / sizeof(sdk_partitions[0]), SDK_FLASH_MAP)) {
        /* 分区表错误启动无意义：停机等看门狗 */
        boot_mark('!');
        for (;;) {
        }
    }
    boot_mark('p');
}

void user_init(void) {
    boot_mark('U');
    aila_rt_main(); /* 进入 AILA：UART + 堆 + main()，不返回 */
}

/* ---------------------------------------------------------------------- */
/* 自备入口：替代 libmain 的 user_start（SDK 启动链，designs/0005 §6.C）    */
/* ---------------------------------------------------------------------- */
/* 为什么需要自己的入口：libmain.a 里 GLOBAL 的 call_user_start_local → 本地
 * user_start 那条官方自举在本板上静默挂死（实测只剩入口打点 'E'，之后无输出、
 * 无异常）。挂点在 user_start 内部：它先 phy_get_bb_evm / ets_install_putc1，
 * 再用 ROM SPIRead 读 flash 头定 map，然后 Cache_Read_Enable_New → 回调
 * user_rf_pre_init / user_pre_init / user_init。
 *
 * 自己写入口时必须自己补上「打开 flash cache」这一步，因为：
 *   1) boot_v1.x 跳到应用时 irom 窗口（0x40200000+）还是空的——我们的 C 运行时
 *      与用户在 IRAM（0x4010xxxx），所以入口阶段能跑；但 AILA 的代码在 irom，
 *      第一条 flash 取指就是 Fatal exception(0) 非法指令（实测
 *      epc1=0x40220d8c = system_partition_table_regist，断在 addi a1,a1,-48 之前）；
 *   2) 不能借 libmain 的 Cache_Read_Enable_New()：它先 call 一份**被链接进 irom**
 *      的 user_iram_memory_is_enabled()（cache 未开就取 flash，必炸），再读
 *      app_main.o 自己的 `.data` 配置字节（SDK 初值 0xFF，只有 0/1 才转发到 ROM，
 *      否则整段跳过＝空转）。
 * 所以直接调 ROM 原语 Cache_Read_Enable(0, 0, 1)（.a 里的 PROVIDE 给到
 * 0x40004678，驻 ROM/IRAM，不需要 cache）。实测这一句之后 irom 可正常取指：
 * 依次打出 E → C → U → `oled: i2c=1 fonts=292`，AILA 的显示栈在 SDK 镜像上跑通。
 *
 * 仍然挡着的两件事（designs/0005 §6.C，都属于「SDK 闭源初始化链」）：
 *   - user_pre_init() 里的 system_partition_table_regist() 会挂在 spi_flash_read
 *     （SPI 读路径在旁路自举下不通），所以分区表注册先跳过；
 *   - 同理 SDK 自己的堆/事件循环没初始化，os_malloc 报 `E:M 1208`，
 *     **Wi-Fi / lwIP 在这条路径上不可用**（显示屏纯 GPIO + I2C，不受影响）。
 * 另外：SDK 变体的 cpu_hz=80MHz 前提是官方自举把 CPU 切到 80MHz；旁路路径下
 * CPU 仍是 52MHz，所以串口实际落在 74880（不是 115200）。
 */
#define AILA_SDK_ENTRY_BYPASS 1

#if AILA_SDK_ENTRY_BYPASS
/* ROM 符号（eagle.rom.addr.v6.ld: PROVIDE(Cache_Read_Enable = 0x40004678)） */
void Cache_Read_Enable(uint32_t odd_even, uint32_t mb_count, uint32_t no_aa);

void call_user_start(void) {
    boot_mark('E');              /* 入口（IRAM，cache 未开也能跑） */
    Cache_Read_Enable(0, 0, 1);  /* flash 0x0.. 线性映射到 0x40200000 */
    boot_mark('C');              /* cache 就绪：此后才允许执行 irom 里的代码 */
    user_rf_pre_init();
    /* user_pre_init() 跳过：它会挂在 spi_flash_read（见文件头注释） */
    user_init();                 /* → aila_rt_main → AILA main()，不返回 */
    for (;;) {
    }
}
#endif

/* ---------------------------------------------------------------------- */
/* std.wifi 胶水（designs/0003 §5.2）：extern 手抄声明，不 include SDK 头。 */
/* Xtensa 窄返回 ABI：SDK 的 bool/uint8 返回经寄存器全宽传递，int 接收安全。 */
/* ---------------------------------------------------------------------- */

/* lwip 等宽类型（arch.h @xtensa）：u16_t/u8_t 必须等宽——pbuf 的
 * len/tot_len 是 u16_t，宽了会错位读字段；err_t = s8_t（err.h），0 = ERR_OK */
typedef unsigned char  u8_t;
typedef signed char    s8_t;
typedef unsigned short u16_t;
typedef unsigned int   u32_t;
typedef int            err_t;

/* _source: NonOS SDK v3.0.5 user_interface.h（wifi_* 在 libnet80211，
 * system_soft_wdt_feed 在 libmain；ets_delay_us 在 ROM 符号表） */
int  wifi_set_opmode_current(unsigned char opmode);            /* STATION_MODE=0x01 */
int  wifi_station_set_config_current(void* config);
int  wifi_station_connect(void);
int  wifi_station_get_connect_status(void);                    /* 0..5，5 = GOT_IP */
int  wifi_get_ip_info(unsigned char if_index, void* info);     /* STATION_IF=0x00 */
void system_soft_wdt_feed(void);
void ets_delay_us(u32_t us);

/* _source: user_interface.h station_config（v3.0.5，108 字节）——零结构 +
 * ssid/password 是 SDK 示例的惯用配法（authmode 0 = 不限认证方式，
 * 由扫描结果自动识别；open_and_wep_mode_disable 0 = 允许 open/WEP） */
struct station_config {
    unsigned char ssid[32];      /* @0（31 字符 + NUL 截断） */
    unsigned char password[64];  /* @32（63 字符 + NUL 截断） */
    unsigned char channel;       /* @96 */
    unsigned char bssid_set;     /* @97 */
    unsigned char bssid[6];      /* @98 */
    struct {
        signed char   rssi;      /* @104 */
        unsigned char authmode;  /* @105：AUTH_OPEN=0 … AUTH_WPA_WPA2_PSK=4 */
    } threshold;
    unsigned char open_and_wep_mode_disable; /* @106 */
    unsigned char all_channel_scan;          /* @107 */
};

/* _source: user_interface.h / lwip ip_addr.h——ip_info = 3 × ip_addr(12B)，
 * addr 为网络序 u32 */
struct aila_ip_addr {
    unsigned int addr;
};
struct aila_ip_info {
    struct aila_ip_addr ip, netmask, gw;
};

void aila_rt_wifi_mode_station(void) {
    (void)wifi_set_opmode_current(1u); /* STATION_MODE */
}

/* 连 SSID 并阻塞轮询：GOT_IP / 确定性失败（2/3/4）即返回，~15s 兜底超时。
 * 10ms 步进喂软看门狗（阻塞用户任务期间不喂会被 soft WDT 复位）。 */
uint32_t aila_rt_wifi_connect(const uint8_t* ssid, uint32_t ssid_len,
                              const uint8_t* pass, uint32_t pass_len) {
    static struct station_config cfg; /* 108B 不上栈；单任务上下文安全 */
    uint8_t* p = (uint8_t*)&cfg;
    for (uint32_t i = 0; i < sizeof(cfg); i++) {
        p[i] = 0;
    }
    uint32_t sl = (ssid_len < 31u) ? ssid_len : 31u;
    uint32_t pl = (pass_len < 63u) ? pass_len : 63u;
    for (uint32_t i = 0; i < sl; i++) {
        cfg.ssid[i] = ssid[i];
    }
    for (uint32_t i = 0; i < pl; i++) {
        cfg.password[i] = pass[i];
    }
    (void)wifi_set_opmode_current(1u); /* connect 前置：确保 station 已启用 */
    (void)wifi_station_set_config_current(&cfg);
    (void)wifi_station_connect();
    for (int i = 0; i < 1500; i++) { /* 1500 × 10ms = 15s */
        system_soft_wdt_feed();
        ets_delay_us(10000);
        uint32_t st = (uint32_t)wifi_station_get_connect_status();
        if (st == 5u || st == 2u || st == 3u || st == 4u) {
            return st; /* GOT_IP / WRONG_PASSWORD / NO_AP_FOUND / CONNECT_FAIL */
        }
    }
    return (uint32_t)wifi_station_get_connect_status();
}

uint32_t aila_rt_wifi_status(void) {
    return (uint32_t)wifi_station_get_connect_status();
}

uint32_t aila_rt_wifi_ip(void) {
    struct aila_ip_info info;
    info.ip.addr = 0;
    if (!wifi_get_ip_info(0u, &info)) { /* STATION_IF */
        return 0;
    }
    return info.ip.addr;
}

/* ---- TCP（lwip raw API）：官方发布版把 netconn/sockets 层剥离了
 * （liblwip.a 的 api_lib.o 是空壳，只有 espconn 路径的代码编进二进制），
 * espconn 又是纯异步回调面——M1 用 raw API 收敛成同步原语：回调（SDK 网络任务
 * 上下文）只灌静态环形缓冲 / 置状态位，用户任务轮询取数。M1 约束：单连接
 * （g_tcp 静态态），环满丢弃新数据。_source: lwip tcp.h / pbuf.h（1.4 布局） */

#define TCP_WRITE_FLAG_COPY 0x01

struct tcp_pcb; /* 不透明句柄（raw API 全程只传指针） */

/* lwip 1.4 pbuf.h 原布局（xtensa 32 位 = 16 字节）——字段序不可动 */
struct aila_pbuf {
    struct aila_pbuf* next;
    void*             payload;
    u16_t             tot_len;
    u16_t             len;
    unsigned char     type;
    unsigned char     flags;
    u16_t             ref;
};

typedef int  (*aila_tcp_recv_fn)(void*, struct tcp_pcb*, struct aila_pbuf*, int);
typedef int  (*aila_tcp_conn_fn)(void*, struct tcp_pcb*, int);
typedef void (*aila_tcp_err_fn)(void*, int);

/* 回调返回 int 而 lwip 内部按 err_t(s8_t) 读：负值低字节一致，ERR_OK=0 等价，
 * Xtensa 窄返回 ABI 兼容（aila 手抄声明的既有约定） */
struct tcp_pcb* tcp_new(void);
void            tcp_arg(struct tcp_pcb*, void*);
void            tcp_recv(struct tcp_pcb*, aila_tcp_recv_fn);
void            tcp_err(struct tcp_pcb*, aila_tcp_err_fn);
err_t           tcp_connect(struct tcp_pcb*, struct aila_ip_addr*, u16_t, aila_tcp_conn_fn);
err_t           tcp_write(struct tcp_pcb*, const void*, u16_t, unsigned char);
err_t           tcp_output(struct tcp_pcb*);
void            tcp_recved(struct tcp_pcb*, u16_t);
err_t           tcp_close(struct tcp_pcb*);
void            tcp_abort(struct tcp_pcb*);
u16_t           pbuf_free(struct aila_pbuf*);

static struct {
    struct tcp_pcb*   pcb;    /* 非 0 = 本连接句柄（err 回调 / close 置 0） */
    volatile uint32_t est;    /* connected 回调置 1 */
    volatile uint32_t dead;   /* err 回调 / 对端关闭 / 显式关闭置 1 */
    volatile int      errcode;
    uint8_t           ring[2048]; /* 收包环形缓冲（回调生产，轮询消费） */
    volatile uint32_t rhead, rtail;
} g_tcp;

static int on_tcp_connected(void* arg, struct tcp_pcb* pcb, int err) {
    (void)arg;
    (void)pcb;
    if (err == 0) {
        g_tcp.est = 1;
    } else {
        g_tcp.dead = 1;
        g_tcp.errcode = err;
    }
    return 0; /* ERR_OK */
}

static int on_tcp_recv(void* arg, struct tcp_pcb* pcb, struct aila_pbuf* p, int err) {
    (void)arg;
    (void)err;
    if (p == 0) { /* p=NULL = 对端优雅关闭 */
        g_tcp.dead = 1;
        return 0;
    }
    for (struct aila_pbuf* q = p; q; q = q->next) { /* 走分片链 */
        const uint8_t* d = (const uint8_t*)q->payload;
        for (u16_t i = 0; i < q->len; i++) {
            uint32_t next = (g_tcp.rhead + 1) & (sizeof(g_tcp.ring) - 1);
            if (next == g_tcp.rtail) {
                break; /* 环满丢弃（M1） */
            }
            g_tcp.ring[g_tcp.rhead] = d[i];
            g_tcp.rhead = next;
        }
    }
    tcp_recved(pcb, p->tot_len); /* 滑动窗口回收 */
    pbuf_free(p);
    return 0;
}

static void on_tcp_err(void* arg, int err) {
    (void)arg;
    g_tcp.dead = 1;
    g_tcp.errcode = err;
    g_tcp.pcb = 0;
}

int64_t aila_rt_tcp_connect(uint32_t ip_be, uint32_t port) {
    g_tcp.pcb     = 0;
    g_tcp.est     = 0;
    g_tcp.dead    = 0;
    g_tcp.errcode = 0;
    g_tcp.rhead   = 0;
    g_tcp.rtail   = 0;
    struct tcp_pcb* pcb = tcp_new();
    if (!pcb) {
        return -1; /* pcb 资源耗尽（ERR_MEM 同义） */
    }
    g_tcp.pcb = pcb;
    tcp_arg(pcb, &g_tcp);
    tcp_err(pcb, on_tcp_err);
    tcp_recv(pcb, on_tcp_recv);
    struct aila_ip_addr a;
    a.addr = ip_be;
    err_t e = tcp_connect(pcb, &a, (u16_t)port, on_tcp_connected);
    if (e != 0) {
        tcp_abort(pcb);
        g_tcp.pcb = 0;
        return (int64_t)e;
    }
    for (int i = 0; i < 1500; i++) { /* 阻塞轮询握手，10ms 步进 + 喂软看门狗 */
        system_soft_wdt_feed();
        ets_delay_us(10000);
        if (g_tcp.est) {
            return (int64_t)(intptr_t)pcb; /* 句柄 = tcp_pcb 指针（恒正） */
        }
        if (g_tcp.dead) {
            return (g_tcp.errcode != 0) ? (int64_t)g_tcp.errcode : -1;
        }
    }
    tcp_abort(pcb); /* 超时强拆 */
    g_tcp.pcb = 0;
    return -1;
}

int64_t aila_rt_tcp_send(uint64_t h, const uint8_t* p, uint32_t n) {
    struct tcp_pcb* pcb = (struct tcp_pcb*)(uintptr_t)h;
    if (!pcb || !p || pcb != g_tcp.pcb) {
        return -1;
    }
    uint32_t off = 0;
    while (off < n) { /* tcp_write 单次限 u16 */
        uint32_t chunk = n - off;
        if (chunk > 0xF000u) {
            chunk = 0xF000u;
        }
        err_t e = tcp_write(pcb, p + off, (u16_t)chunk, TCP_WRITE_FLAG_COPY);
        if (e != 0) {
            return (int64_t)e; /* lwip 发送缓冲满 = ERR_MEM，M1 不重试 */
        }
        off += chunk;
    }
    return (int64_t)tcp_output(pcb);
}

int64_t aila_rt_tcp_recv(uint64_t h, uint8_t* p, uint32_t cap) {
    (void)h; /* 单连接 M1：只读环形缓冲，不解引用句柄（关闭后仍可排空） */
    if (!p || cap == 0) {
        return -1;
    }
    uint32_t got = 0;
    while (got < cap && g_tcp.rtail != g_tcp.rhead) {
        p[got++] = g_tcp.ring[g_tcp.rtail];
        g_tcp.rtail = (g_tcp.rtail + 1) & (sizeof(g_tcp.ring) - 1);
    }
    if (got > 0) {
        return (int64_t)got;
    }
    if (g_tcp.dead) { /* 无数据且已死：对端关闭报 -1，异常报 lwip err */
        return (g_tcp.errcode != 0) ? (int64_t)g_tcp.errcode : -1;
    }
    return 0;
}

void aila_rt_tcp_close(uint64_t h) {
    struct tcp_pcb* pcb = (struct tcp_pcb*)(uintptr_t)h;
    if (!pcb) {
        return;
    }
    if (g_tcp.pcb == pcb) {
        g_tcp.pcb = 0;
        g_tcp.dead = 1;
    }
    if (tcp_close(pcb) != 0) { /* 发送队列未清会 ERR_MEM 且 pcb 仍被持有 */
        tcp_abort(pcb);        /* 强拆立即释放 */
    }
}
