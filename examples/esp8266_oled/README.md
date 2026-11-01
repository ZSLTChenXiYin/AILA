# esp8266_oled —— 连 Wi-Fi → TCP 收中文 → 4 针 I2C OLED 显示

配套服务端：`examples/net_console`（PC 控制台敲中文，回车发过来）。

## 接线（4 针模块 = VCC / GND / SCL / SDA）

| OLED | ESP8266（NodeMCU 丝印） | 说明 |
|---|---|---|
| VCC | 3V3 | 别接 5V，SSD1306 模块多数只有 3.3V 稳压 |
| GND | GND | |
| SCL | GPIO5 = D1 | `main.aila` 里 `PIN_SCL` |
| SDA | GPIO4 = D2 | `main.aila` 里 `PIN_SDA` |

I2C 地址按 `0x3C`（绝大多数模块）；少数模块是 `0x3D`，改
`cpacks/ssd1306/ssd1306.c` 里的 `OLED_ADDR`。

## 改三个地方就能用

```aila
const WIFI_SSID: Text = "ChenXiYin";   // 已填你的 Wi-Fi
const WIFI_PASS: Text = "84170919";
const SERVER_IP: U32  = 0xC0A8010A;    // ← 跑 net_console 那台 PC 的局域网 IP
const SERVER_PORT: U32 = 8266;         // ← 与服务端一致
```

`SERVER_IP` 的形状是 `a.b.c.d` → `0xAABBCCDD`：192.168.1.10 就是 `0xC0A8010A`。

## 屏幕与协议

- 128x64 屏，每个字 16x16 → **8 列 × 4 行**；满 4 行后整屏上滚一行。
- 协议：一行一条消息（UTF-8 + `\n`）；收到 `CLEAR` 清屏。
- 收到但字库里没有的字画成实心方块（一眼看出缺字，不会静默变空白）。
- 所有收到的行同时从 UART0（115200）打出来，接串口能看到同一份内容。

## 点阵字库

`font/font16.aila` 是自动生成的（ASCII + 284 个常用汉字，约 9 KB 点阵）：

```bash
pwsh -File tools/genfont.ps1                  # 重新生成
pwsh -File tools/genfont.ps1 -Extra "空调除湿"  # 追加几个字
```

字源是 Windows 自带的宋体（SimSun 16px 有内嵌点阵，切出来就是干净的 16x16）。
要换字号/字体：`pwsh -File tools/genfont.ps1 -FontName "Microsoft YaHei"`（16px 时宋体最锐）。

## 构建：还差什么

`aila check` 现在就能过（含 xtensa 目标的语义检查与 IR 生成），出 `app.bin` 还缺三件：

| 缺的东西 | 为什么 | 怎么补 |
|---|---|---|
| **xtensa 后端** | AILA 出的是 LLVM IR，把 IR 编成 Xtensa 机器码需要带 Xtensa 的 clang（本机装的 llvm-mingw 只有 x86/x86-64 后端） | 装 esp-clang（注册表里那条 `esp-clang` 条目要等 CI 打包，或拿乐鑫官方工具链的 clang 做本地 zip：`aila toolchain install <zip>`） |
| **xtensa 链接器** | 最终 `app.elf`/`app.bin` 由随包 GCC 链接 | `aila toolchain install gcc/xtensa-lx106-elf` |
| **NonOS SDK** | `std.wifi` 的 TCP/Wi-Fi 胶水在 `runtime/esp8266/nosdk.c`，需要 SDK 的头与库 | `aila toolchain install sdk/esp8266-nosdk` |

缺后端时的表现是 `E8010`（IR 已经落在
`.local/build/examples/esp8266_oled/xtensa_esp8266_elf/*.ll`，可以先用
`llc -march=xtensa` 之类工具单独试）。

## 代码结构

| 文件 | 干什么 |
|---|---|
| `main.aila` | Wi-Fi 连接、TCP 收行、UTF-8 解码、排版、写帧缓冲、刷新屏 |
| `font/font16.aila` | 16x16 点阵字库（生成物，别手改） |
| `cpacks/ssd1306/ssd1306.c` | 软件 I2C + SSD1306 命令/帧写（微秒级时序放 C，符合 designs/0002 的纯度三线） |

时序参数：`I2C_DELAY_US = 2`（半周期）→ SCL ≈ 250kHz，SSD1306 上限 400kHz，稳。
接线长了、屏花屏就把这个值调到 4~5。
