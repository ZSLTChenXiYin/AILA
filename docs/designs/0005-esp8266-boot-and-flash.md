# 0005 · ESP8266 启动链与 flash 布局（实测结论）

> 状态：bootloader / flash map / 镜像格式 / 参数扇区四件事已实测确认（2026-11-01），
> 板子：ESP8266EX + 26MHz 晶振 + **4MB** flash（CH340，COM3）。
> 本文只记录"上板验证过"的事实，推断处明确标注。

## 1. 两条启动路径

| 路径 | flash 0x0 | 应用位置 | 镜像格式 | AILA 用途 |
|---|---|---|---|---|
| ① ROM 直启 | 应用镜像（`0xE9`，image v1） | 0x0 | v1 | **freestanding**（`examples/oled_smoke`） |
| ② bootloader | `boot_v1.x.bin`（`0xE9`，image v1） | 0x1000（user1）/ 0x81000（map 4 的 user2） | **v2（`0xEA`）** | SDK 模式（`[sdk] name = "esp8266-nosdk"`） |

- ROM 只把镜像头声明的段搬进 RAM，然后跳 entry；**不会替应用打开 flash cache**。
- 路径 ② 的 bootloader 负责打开 cache，并把 flash map 通过 RTC 传给应用。
- `esptool image-info <bin>` 可直接看出镜像版本/段表/entry，是排查第一站。

## 2. flash map 的唯一来源：bootloader 镜像头的 size nibble

`boot_v1.x.bin` 的镜像头 8 字节是 `E9 nseg mode size_freq entry`，其中
**`size_freq` 的高 4 位 = flash size code**，bootloader 据此决定 map 并通过 RTC 告诉应用：

| boot 头 byte3 | bootloader 打印 | SDK 侧 map |
|---|---|---|
| `0x00` | `SPI Flash Size & Map: 4Mbit(256KB+256KB)` | 0（512KB） |
| `0x20` | `8Mbit(512KB+512KB)` | 2（1MB） |
| `0x40` | `32Mbit(512KB+512KB)` | 4（4MB） |

**应用编译时带的 map（`system_partition_table_regist` 的 map 参数）必须与它相等**，否则
应用启动最早期就报 `mismatch map X,spi_size_map Y` → `system param partition error`
→ 重启循环。实测：map-2 的官方 AT 固件必须配 `byte3 = 0x20` 的 bootloader 才能跑。

> 这也是"用 4MB 板子跑官方 AT 固件"的关键：SDK 只带 512+512(map 2) 与 1024+1024(map 5)
> 两种 AT 构建，把它们当成更小容量的板子来喂 boot 头即可（flash 物理容量大不影响使用）。

## 3. 镜像格式

- **v1（`0xE9`）**：头 8B + 每段 `{len, addr}` + 数据 + 每段 1 字节 XOR 校验，末段后 16 字节对齐 + 总校验字节。
  段数据必须**补 0 到 8 字节对齐**，否则 ROM 把下一段头也划进校验，`tail` 后静默停机
  （`internal/image` 已修）。
- **v2（`0xEA`）**：头 8B + `{0, irom_len}`（8B）+ irom 数据 + 每段 `{len, addr}` + 数据。
  `esptool image-info` 显示的 seg0（load=0）就是这段 irom。**FOTA 槽位（0x1000）只认 v2**，
  喂 v1 时 bootloader 会"跳了但什么都不做"（无任何报错，最容易误判）。
  当前产出方式：`esptool --chip esp8266 elf2image --version 2 --flash-mode dout --flash-freq 40m --flash-size 4MB -o app.bin app.elf`。

## 4. IRAM 驻留纪律（SDK 路径）

`libmain.a:app_main.o` 的 `call_user_start_local` / `Cache_Read_Enable_New` / `user_start`
是**打开 flash cache 的启动链**，必须驻留 IRAM（0x4010xxxx）。随包 SDK 的 `.a` 用
`-ffunction-sections` 编译，若链接脚本让 flash 段的全局兜底 `*(.text.* .literal.*)` 把它们
扫进 flash，入口就会 `callx0` 到还没开 cache 的 flash 上 → **静默挂死**。
`internal/chip/chip.go` 用 `.iram0.boot` 段在兜底之前定点捕获这三个函数。

## 5. 参数扇区

| 扇区 | map 2（1MB 布局） | map 4（4MB 布局） | 内容 |
|---|---|---|---|
| RF_CAL | 0x7B000 | 0x3FB000 | 必须是**擦除态 0xFF**，SDK 首次启动自行校准；写成全 0 会 `rf_cal[0] !=0x05` 反复重启 |
| PHY_DATA | 0x7C000 | 0x3FC000 | `esp_init_data_default_v08.bin` |
| SYSTEM_PARAM | 0x7D000 | 0x3FD000 | 擦除即可，SDK 会自建 |

## 6. 已验证的烧写配方

**A. freestanding（OLED 演示，可用）**
```powershell
aila build examples/oled_smoke
esptool --chip esp8266 --port COM3 write-flash -fm keep -fs keep -ff keep 0x0 <app.bin>
```

**B. 官方 AT 固件（Wi-Fi 可用，已端到端验证）**
```powershell
# 1) 把 boot_v1.7 头的 size 位改成 map 2（1MB）
$b = [IO.File]::ReadAllBytes("$sdk/bin/boot_v1.7.bin"); $b[3] = 0x20; [IO.File]::WriteAllBytes("$TEMP\boot_1m.bin", $b)
# 2) 擦干净参数区（RF_CAL 必须 0xFF）
esptool --chip esp8266 --port COM3 erase-region 0x7B000 0x5000
# 3) 三片
esptool --chip esp8266 --port COM3 write-flash -fm keep -fs keep -ff keep `
  0x0      $TEMP\boot_1m.bin `
  0x1000   $sdk/bin/at/512+512/user1.1024.new.2.bin `
  0x7C000  $sdk/bin/esp_init_data_default_v08.bin
```
验证（2026-11-01 实测）：`AT` → `OK`；`AT+GMR` → AT 1.7.5.0 / SDK 3.0.5；
`AT+CWJAP` 连上 AP 拿到 `192.168.1.100`；`AT+CIPSTART` + `AT+CIPSEND` 向 PC 服务端
（`examples/net_console` 的 8266 端口）发送 `你好，OLED`，服务端收到的字节为
`E4 BD A0 E5 A5 BD EF BC 8C 4F 4C 45 44 0D 0A`（UTF-8 完整）。

**C. AILA 的 SDK 应用（Wi-Fi + OLED，显示已打通、Wi-Fi 未打通）**
boot 头 `0x40`（4MB/map 4）+ `aila build` 自产的 FOTA v2 镜像 @0x1000（user2 0x81000 同镜像）。

**已确认（2026-11-02，本轮全部修好）**：
1. **irom 链接地址必须与槽位对齐**——这一条是此前「静默挂死 / Fatal exception」的**真根因**：
   cache 是 flash 0x0 ↔ 0x40200000 的固定线性映射，v2 镜像的 irom 段紧跟 16B 镜像头
   （`esptool image-info` 的 seg0 file offs=0x08），所以烧在 user1 槽（flash 0x1000）的
   应用，其 irom 必须链接到 **0x40201010**（与官方 `eagle.app.v6.new.512.app1.ld` 同值）。
   此前按非 FOTA 老布局链到 0x40210000（= flash 0x10000）而实际烧在 0x1010，
   所有 flash 取指差 0xF000 → `Fatal exception (0)` 非法指令 + 重启循环。
   `chip.json` 现在由 `sdks.*.irom.{origin,slot}` 表达这条不变量，`Pack8266SDK` 会校验。
2. **构建产物必须就是能烧的那个镜像**：`Pack8266SDK` 从「v1 老布局 + 0xFF 填充」
   改成 **FOTA v2 + SDK CRC32 尾**（主头 0xEA / irom `{0,len}` / 次头 0xE9 / RAM 段 /
   校验和占 16B 块末字节 / `esp8266_crc32`），`tools/esp-flash.ps1 -Mode sdk` 直接烧
   `app.bin`，不再外挂 `esptool elf2image`（少了「构建产物 ≠ 烧写产物」这条暗坑）。
3. **flash cache 必须由应用自己打开**：bootloader 跳到应用时 irom 窗口是空的。
   - 不能借 libmain 的 `Cache_Read_Enable_New()`：它先 `call` 一份**被链接进 irom** 的
     `user_iram_memory_is_enabled()`（cache 未开就取 flash → 必炸），再读 `app_main.o`
     自己的 `.data` 配置字节（`R_XTENSA_32 .data`，SDK 初值 **0xFF**；只有 0/1 才转发到
     ROM，否则整段跳过＝空转）。这个字节正是官方 `user_start` 在探测 flash 头后写的。
   - 现在直接调 ROM 原语 `Cache_Read_Enable(0, 0, 1)`（`PROVIDE 0x40004678`，驻 ROM）。
     实测之后 irom 可正常取指：串口依次给出 `E C F U` →
     **`oled: i2c=1 fonts=292`**（OLED 初始化 + 字库就位），无异常、无重启。
   - 官方入口 `call_user_start_local()`（GLOBAL）在本板上**仍然静默挂死**：只剩入口打点，
     无输出无异常；`.bss` 先清零也不改变结论。挂点在 `user_start` 内部的 SDK 初始化。
4. 剩下挡着的（都属于 SDK 闭源初始化链，与显示无关）：
   - `user_pre_init()` → `system_partition_table_regist()` 挂在 `spi_flash_read`
     （SPI 读路径在旁路自举下不通），所以分区表注册暂时跳过；
   - SDK 自己的堆 / 事件循环没初始化 → `os_malloc` 打 `E:M 1208`，**Wi-Fi / lwIP 不可用**。

**串口波特率**：SDK 变体的 `cpu_hz=80MHz` 前提是官方自举把 CPU 切到 80MHz；旁路路径下
CPU 仍是 52MHz，所以应用的输出落在 **74880**（bootloader 的 `load/chksum` 与
`Fatal exception` 也是 74880）。

**下一步**（要在真机跑 Wi-Fi，必须对齐 SDK 自举而不是绕开它）：对着官方 AT 镜像的 crt
把 `user_start` 里挂住的那段（SPIRead 读 flash 头 → 定 map → 装 putc → 事件循环）
逐条复刻；或在 SDK 里找到等价的可调用初始化序列（堆初始化 + 事件循环）后自建入口。

## 7. 其它实测硬件事实

- 本板 **QIO 可用**（镜像头 mode=0 亦能启动）；此前"只认 DOUT"的结论只测到 DIO 失败。
- 裸机 CPU 时钟是 **52MHz**（26MHz 晶振二倍频），80MHz 要软件切换；UART 分频必须按实际主频算，
  否则 115200 输出会变成 74880。见 `runtime/esp8266/chip.json` 的 `_source.cpu_hz`。
- `.bss` 是 NOLOAD 段，ROM 不会清零，freestanding 运行时必须自己清（`aila_rt_main` 已做）。
- 裸机镜像没装 ISR，进 `main` 前必须 `rsil 15` 屏蔽可中断，否则随机跑飞。
