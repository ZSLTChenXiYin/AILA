# 0003 · 标准库三层规划（2026-10-26 共识）

状态：批次 1 / 2 / 3 已全部落地；批次 4（显示栈：pixfmt / pixbuf / damage / paint / textgrid / display / display_host，见 designs/0006）已落地（批次 1：strconv / strings / utf8 / binary / sort 五模块 + stdlib_demo 验收；批次 2：time / os / rand / path 四模块 + 宿主与三家族 rt 原语 + os_demo 验收；批次 3：gpio / uart 三家族寄存器实现 + wifi SDK 胶水 + esp8266/gpio_demo、wifi_std 验收）。

## 1. 分层与组织（共识）

标准库分三层，全部保持在 `std` 前缀下、独立模块（方案 A）：

| 层 | 内容 | runtime 关系 | target 校验 |
|---|---|---|---|
| 1 通用库 | strconv / strings / utf8 / binary / sort / io / math / time / rand / path | 纯 AILA + 少量「两边都实现」的 rt 原语 | 任何 target 可 import |
| 2 os 库 | os（fs/env/args/exit） | 仅宿主 aila_rt.c 实现 | 嵌入式上误用 → 链接期 undefined symbol（优化报错信息指认 std.os） |
| 3 嵌入式封装 | gpio / uart / wifi | 各家族 aila_rt_<家族>.c 实现 | 宿主上误用 → 同上 |

- **M1 不加 `#[target]` 模块属性**：属性校验是错误体验优化（编译期 import 报错），非实现前提。链接期 undefined symbol 报错信息做映射优化（缺 `aila_rt_file_open` → 提示「std.os 需要 OS target」）即可，属性后补。
- 物理布局：stdlib/ 扁平 `.aila`，模块名 = import 名（`std.<名>` → `stdlib/<名>.aila`，home.StdlibDir 解析，无需改构建代码）。
- 第 3 层 API 蓝本：embedded-hal + Arduino API，AILA 化裁剪（fnptr 回调、T!E、region 缓冲、显式 `@mut`）。

## 2. 现状基线

- `std.core`：IoErr { Closed, NoSpace{need}, Timeout }、fill_u64/fill_i64/fill_f64、utf8_encode。
- `std.io`：vtable Writer + stdout 后端 + print/println/printf（FmtVal 封闭联合）；OS 边界仅 `#[c] aila_rt_write(bytes, n) / aila_rt_flush()`。
- `std.math`：回绕/最值/fmod 窄函数族。
- 语言风格要点（写库必须遵守）：普通函数直接给实现体；Writer 方法先声明签名后 module 内实现；分配走 `region <名> { var buf: [T] = [0; n]@<名>; }`；显式转换 `(x trunc U32 + 48) trunc U8`；写权 `@mut[buf]`；Text↔[U8] 用 `view bytes: [U8] = t`；变参 rest 是只读视图（for-in/len/下标）。

## 3. 批次 1：纯 AILA 库（零 runtime）

### 3.1 std.strconv（已落地）
```aila
public enum Err { Empty, Invalid { at: USize }, Range };

// 解析：只跳前导空白（\t \n \v \f \r 空格），尾部空白非法；失败返回 Err
public func parse_i64(t: Text) -> I64 ! Err;   // 溢出判式 acc > (max - d) / 10，先判后乘
public func parse_u64(t: Text) -> U64 ! Err;   // '-' 非法；'+/' 合法
public func parse_f64(t: Text) -> F64 ! Err;   // [+-]digits[.digits][(e|E)[+-]digits]；指数上限 10000；乘爆 Inf = Range；下溢 = 0
public func format_i64(v: I64, buf: [U8]) @mut[buf] -> USize;   // core.fill_* 直通
public func format_u64(v: U64, buf: [U8]) @mut[buf] -> USize;
public func format_f64(v: F64, buf: [U8]) @mut[buf] -> USize;
```

实现要点：I64.MIN 特判（acc = 2^63 直接返回 `-9223372036854775808`，I64 域取负在 U64 域做会回绕 trap）；≥ 2^63 的字面量必须写十六进制（十进制超 I64.MAX 落有符号域报 E3204）。

**实现期语言修复（批次 1 引入）**：≥ 2^63 的整数字面量原先被 lexer 以「ParseInt 溢出 → ParseFloat 十进制误读」吞掉（0x8000000000000000 变 8e15、0xFFFFFFFFFFFFFFFF 变 0）。修复：lexer 全程 `ParseUint` 位型记录（`UntypedInt{Value: 位型, BitPattern: true}`），sema 限定位型只可直落 64 位无符号（U64 / USize），有符号 / 窄目标 E3201；一元负号按无符号幅值取负（`-2^63` 合法且恰为 I64.MIN）；超 64 位字面量 lexer E1006 报错。

**语言查证结论（2026-10-26）**：Text 只有字面量路径（StrLit → rodata），无运行时构造——「返回动态 Text」的 API（to_text/split/concat/env/args）M1 不可行。**全部改为调用方缓冲 + 返回写入数**的零分配形态（下方签名已是缓冲版）；`text_from_bytes` 内建记为 M2 语言项。

### 3.2 std.strings（已落地，签名有两处实现期修正）

```aila
// 查询（零分配）
public func len_runes(t: Text) -> USize;          // UTF-8 码点数（非法序列按 1 字节 1 码点跳过）
public func contains(t: Text, sub: Text) -> Bool;
public func index(t: Text, sub: Text) -> USize;    // 字节下标；无则 len t（哨兵）；空子串恒 0
public func has_prefix(t: Text, p: Text) -> Bool;
public func has_suffix(t: Text, s: Text) -> Bool;
public func bounds(t: Text) -> (USize, USize);     // 去首尾 ASCII 空白后的 [start, end)；空 = (0, 0)

// 构造（写进调用方缓冲，返回实际写入字节数 = min(需要量, len buf)；容量不足截断不报错）
public func concat(a: Text, b: Text, buf: [U8]) @mut[buf] -> USize;
public func split(t: Text, sep: U8, marks: [USize]) @mut[marks] -> USize;  // 填每段起始偏移（按容量截断），返回完整段数；"a," = 2 段
public func join(buf: [U8], sep: U8, parts: ...Text) @mut[buf] -> USize;   // 变参必须末位 → buf/sep 前置
public func replace(t: Text, from: Text, to: Text, buf: [U8]) @mut[buf] -> USize;  // from 空 = 原样拷贝（M1）
public func upper_ascii(t: Text, buf: [U8]) @mut[buf] -> USize;
public func lower_ascii(t: Text, buf: [U8]) @mut[buf] -> USize;
```

实现期修正（相对初稿）：① `out` 是语言保留字 → 缓冲参数一律改名 `buf`；② 变参必须末位（language.md §4）→ `join` 签名重排为 `(buf, sep, ...parts)`。

### 3.3 std.utf8（已落地：encode 自 core 迁出，std.io 改调本模块，core 不再持有）
```aila
public func encode(cp: U32, buf: [U8]) @mut[buf] -> USize;   // buf 需 ≥ 4
public func rune_len(first: U8) -> USize;          // 首字节 → 序列长；0 = 非法
public func cp_at(t: Text, i: USize) -> U32;       // 0xFFFFFFFF = 非法（哨兵）
public func valid(t: Text) -> Bool;                // 结构性校验（不查超长/代理区，M2 收紧）
```

### 3.4 std.binary（已落地：M1 整数族 U16..U64 / I16..I64 × LE/BE）

```aila
public func le_read_u16(buf: [U8], off: USize) -> U16;   // ……全整数族；有符号族 = 无符号拼装 + 同宽重释（U16 as I16）
// 写：le_write_u16(buf, off, v) @mut[buf] —— ……全整数族
// BE：be_read_*/be_write_* 同型
```

**f32/f64 族 M1 不做**：位重解释（U32 as F32）在当前语言不可表达——`as` 只做数值无损转换（sema `castLossless` 整浮互斥）、`trunc` 只做数值有损转换。位重解释记为 **M2 语言项**，落地后补六件 f 族 API。纯 AILA 手工位分解（1.m × 2^e）精度边界风险大，不做。

### 3.5 std.sort
```aila
public func sort_i64(buf: [I64]) @mut[buf];        // 升序（简单快排/希尔）
public func sort_u64(buf: [U64]) @mut[buf];
public func sort_f64(buf: [F64]) @mut[buf];
public func sort_i64_by(buf: [I64], cmp: (I64, I64) -> Bool) @mut[buf];
public func sort_u64_by(buf: [U64], cmp: (U64, U64) -> Bool) @mut[buf];
public func sort_f64_by(buf: [F64], cmp: (F64, F64) -> Bool) @mut[buf];
```
通用元素版（esize + 字节交换）M2 泛型后做。

## 4. 批次 2：rt 原语 + time/os/rand/path

### 4.1 rt 原语清单（全部 `#[c]` 声明进对应 AILA 模块）

| 原语 | 宿主 aila_rt.c | 嵌入式三家族 |
|---|---|---|
| `aila_rt_time_ms() -> U64` | clock_gettime(CLOCK_MONOTONIC) / QueryPerformanceCounter | ccount 累积法（静态 last + 差值模 2^32 累积；注释注明调用间隔须 < 半周期） |
| `aila_rt_sleep_ms(n: U64)` | MinGW Sleep / POSIX nanosleep | ccount 忙等 |
| `aila_rt_exit(code: U32)` | exit(code) | 不提供（嵌入式无退出语义） |
| `aila_rt_env_get(name, n) -> I64`（写宿主缓冲返回字节数，-1 无） | getenv | 不提供 |
| `aila_rt_env_set(name, n, val, vn) -> I64` | setenv/_putenv | 不提供 |
| `aila_rt_args_count() -> I64` | 见 4.3 | 不提供 |
| `aila_rt_args_get(i, buf, cap) -> I64` | 同上 | 不提供 |
| `aila_rt_file_open(path, n, mode: U32) -> U64`（0=失败；mode 0 只读/1 写创建/2 追加） | fopen 系（"rb"/"wb"/"ab"） | 不提供 |
| `aila_rt_file_read(h: U64, buf, n) -> I64`（-1 错） | fread | 不提供 |
| `aila_rt_file_write(h: U64, bytes, n) -> I64` | fwrite | 不提供 |
| `aila_rt_file_close(h: U64)` | fclose | 不提供 |
| `aila_rt_file_flush(h: U64)` | fflush | 不提供 |
| `aila_rt_seed() -> U64` | urandom（fread /dev/urandom）| ccount 快照 |

### 4.2 std.time / std.os / std.rand / std.path

```aila
// std.time（层 1）
public func mono_ms() -> U64;
public func sleep_ms(n: U64);

// std.os（层 2，仅宿主链接通过）
public struct File { h: U64 };
public func open(path: Text, mode: U32) -> File ! IoErr;   // IoErr 扩 NotFound / Denied
public func read(f: File, buf: [U8]) @mut[buf] -> USize ! IoErr;
public func write(f: File, bytes: [U8]) -> USize ! IoErr;
public func close(f: File);
public func exit(code: U32);
public func env(name: Text, out: [U8]) @mut[out] -> USize ! IoErr;   // 未设置 = NotFound
public func env_set(name: Text, val: Text) -> IoErr?;                // None = 成功
public func args_count() -> USize;
public func args_get(i: USize, out: [U8]) @mut[out] -> USize ! IoErr; // 越界 = Range? 用 NotFound

// std.rand（层 1；xorshift64*）
public struct Rand { s: U64 };
public func Rand.seed(self, s: U64) @mut[self];
public func Rand.next_u64(self) @mut[self] -> U64;
public func Rand.below(self, bound: U64) @mut[self] -> U64; // 拒绝采样

// std.path（层 1；sep 参数化，宿主/嵌入式惯用 '/' = 47）
public func join(a: Text, b: Text, sep: U8, out: [U8]) @mut[out] -> USize;
public func base(p: Text, sep: U8) -> (USize, USize);   // 末段 [start, end)（零拷贝）
public func dir(p: Text, sep: U8) -> (USize, USize);    // 末段之前 [start, end)；无分隔符 = (0, 0)
public func ext(p: Text) -> (USize, USize);             // 末个 '.' 之后；无 = (0, 0)
```

实现以 stdlib/*.aila 为准：签名里的 `out` 参数落地为 `dst`（`out` 是保留字）、rand 换三元 xorshift、os 的 rt 原语带显式 `*_len/*_cap`——差异清单见 §4.4。

### 4.3 宿主 args 的入口问题（已落地）

实现形态：codegen 的 `genEntryWrapper` 生成的 C 入口 `@main` 改为收 `(i32 %argc, ptr %argv)`；**仅当程序模块图 import 了 std.os**（`res.Program.Order` 含 `"std.os"`）时，在调用 AILA main 前插入 `call void @aila_rt_args_save(i32 %argc, ptr %argv)` 转存，`args_count/args_get` 从 runtime 静态状态读。AILA 的 main 签名不动。嵌入式工程不 import std.os → 不生成转发调用；嵌入式 startup.S 以 0 参调 main，多余实参被 C ABI 忽略，签名加宽无影响（corpus IR 快照已随之重生成）。

### 4.4 批次 2 实现期修正（与规划稿的差异）

1. **`out` 是保留字**（批次 1 已踩过，os/path 再次踩到）：缓冲参数统一改名 `dst`；os.aila / path.aila 的公开签名与规划稿不一致处以实现为准。
2. **`#[c]` 视图参数只传指针不展开长度**（既有约定，io.aila 的 `aila_rt_write(bytes, n)` 即先例）：os 的 rt 声明一律显式跟 `*_len / *_cap` 参数；调用点传 `len 视图`。
3. **import 不做非限定展开**：模块名前缀必须写全（os.aila 里用 `core.IoErr`，不能裸写 `IoErr`）。
4. **`IoErr?` 合法**（enum 属句柄类，`T?` 可用）；`Some { v: ... }` 是显式包装形态，`None` 是空臂值。
5. **rand 不用 xorshift64\*（乘法变体）**：AILA 整数运算是溢出陷阱语义，64 位乘法必然溢出——改用三元 xorshift（13/7/17），纯移位 + 异或。`below` 的拒绝采样阈值 `2^64 mod bound` 由 `(0xFFFFFFFFFFFFFFFF % bound) + 1` 等价算出，U64 域无溢出。
6. **`@mut[self]` 方法的 codegen 缺口（本批修复）**：此前 self 按值传入、方法内写不落调用方（「此前只是没有语料踩到」）。现 self 以 `ptr` 传入（`mutsSelf(sig)` 按参数下标 0 判），调用点传 var 主名字槽位地址（`lowerPlace`），方法内读写直接作用调用方内存——designs/0001 §5.4「就地缩放」语义落地。IR 快照随之重生成。
7. **随包 MinGW 头不全**（缺 `psdk_inc/intrin-impl.h`，`windows.h` 不可用）：宿主 aila_rt.c 不 include windows.h，直接声明用到的三个 WinAPI（QPC / QPF / Sleep）+ `LARGE_INTEGER` 自定义形态；`_getpid` 走 process.h。
8. **`chip.json` 新增 `cpu_hz`**（esp8266 80M 固定；esp32/esp32s3 复位默认 240M，AILA runtime 不改频），机械渲染进 `chip_cfg.h` 的 `AILA_CPU_HZ`——三家族 time/sleep/seed 的 ccount 换算基准。
9. **IoErr 的读写失败归型**：M1 变体有限，read 非 0 错 / write 短写 / env-args 放不下统一 `NoSpace { need }`；`NotFound`（未设置 / 越界 / 文件不存在）已用上，`Denied` 留给 M2 接口化细分。
10. **链接期 undefined symbol 映射**（§6 收尾项提前落地）：宿主链接失败时 `explainLinkError` 识别 `aila_rt_args_*/env_*/exit/file_*`，追加「std.os 需要 OS target」提示。

## 5. 批次 3：嵌入式封装（第 3 层）

### 5.1 std.gpio / std.uart（三家族寄存器实现，加数据不加语义；已落地）

```aila
// std.gpio
public const MODE_IN: U32 = 0;
public const MODE_OUT: U32 = 1;
public func mode(pin: U32, m: U32);        // m: MODE_IN / MODE_OUT
public func write(pin: U32, v: Bool);
public func read(pin: U32) -> Bool;

// std.uart
public func init(id: U32, baud: U32);
public func write(id: U32, bytes: [U8]) -> USize;   // 阻塞写满
public func read(id: U32, buf: [U8]) @mut[buf] -> USize;  // 非阻塞，返回当前可用
```

- rt 原语：`aila_rt_gpio_mode/read/write`、`aila_rt_uart_init/write/read`（全部落地）。
- 实现源：三家族 `aila_rt_<家族>.c`；寄存器常量经 chip.json 的 gpio 段（单一事实源）生成 chip_cfg.h 注入——esp8266 GPIO@0x60000300（OUT 0x00 / W1TS 0x04 / W1TC 0x08 / ENABLE_W1TS 0x10 / ENABLE_W1TC 0x14 / IN 0x18，eagle_soc.h），esp32 @0x3FF44000 与 esp32s3 @0x60004000（OUT +0x04 / W1TS +0x08 / W1TC +0x0C / ENABLE_W1TS +0x24 / ENABLE_W1TC +0x28 / IN +0x3C，esp-idf gpio_reg.h）。UART 增补 rxcnt_shift/rxcnt_bits（esp8266 RX=[7:0]、esp32 [7:0]、esp32s3 [9:0]）。
- esp8266 现有 stdout UART 驱动重构为该原语的实现之一（aila_rt_write 内部调 uart 写路径，行为不回归；三家族同构）。

### 5.2 std.wifi（esp8266 NonOS SDK build 限定，M1 station + TCP 最小集；已落地）

```aila
// 仅 [sdk] = esp8266-nosdk 的工程可链接（aila_rt_wifi_* 只在 nosdk.c 提供）
public const ST_IDLE..ST_GOT_IP: U32;             // NonOS esconn_stat 枚举（见 §5.3 ⑤）
public func mode_station();                       // wifi_set_opmode_current(STATION_MODE)
public func connect(ssid: Text, pass: Text) -> Bool;   // 阻塞轮询（GOT_IP / 确定性失败 / ~15s）
public func status() -> U32;                      // NonOS 原生状态码
public func ip() -> U32;                          // ip_info.ip（网络序 u32；0 = 无）
// TCP（lwip raw API——见 §5.3 ⑥）
public func tcp_connect(ipa: U32, port: U32) -> U64 ! IoErr;  // 句柄 = tcp_pcb 指针
public func tcp_send(h: U64, bytes: [U8]) -> I64;   // 0 = 成功 / 负值 = lwip err
public func tcp_recv(h: U64, buf: [U8]) @mut[buf] -> I64;  // 字节数 / 0 = 无数据 / 负值
public func tcp_close(h: U64);
```

- 实现放 `runtime/esp8266/nosdk.c`（SDK build 才编它）：SDK API（wifi_set_opmode_current / wifi_station_set_config_current / wifi_station_connect / wifi_station_get_connect_status / wifi_get_ip_info）+ lwip **raw TCP**（tcp_new/connect/write/recv/close/abort + pbuf_free）。extern 手声明，不 include SDK 头；SDK 自定义类型按官方头布局手抄（station_config 108B / pbuf 16B / ip_addr）。
- TCP 同步化：raw 回调（SDK 网络任务上下文）灌 2KB 静态环形缓冲 + 置状态位，用户任务轮询——M1 单连接、环满丢弃。
- 示例工程 `esp8266/wifi_std`（[sdk] 段 manifest）已构建通过 app.bin——std.wifi 验收完成。

### 5.3 批次 3 实现期修正（与规划稿的差异）

1. **esp8266 UART TX/RX 字段勘误（chip.json 修正）**：早前记录的 TXFIFO_CNT=[31:26]/6 位是错的（eagle_soc.h 根本无 UART 定义，TRM 记法亦有误）。三源仲裁（SDK v3.0.5 uart_register.h + esp8266/Arduino esp8266_peri.h 的 USTXC=16/USRXC=0 实机印证 + FIFO 深度 128B 需 8 位字段）：**TX=[23:16] 8 位、RX=[7:0] 8 位**。chip.json 已改并更新 _source。
2. **gpio M1 范围**：只做 GPIO 块主寄存器；IO_MUX / pad 复用与输入使能管理记 M2（esp32 系 read 的 pad 前置条件见 rt C 注释；esp8266 pad 输入恒使能）。位域 0-31：esp8266 引脚 0-15（GPIO16 属 RTC 域），esp32/esp32s3 引脚 32+ 走 OUT1/ENABLE1 组寄存器（M2）。越界运行期 trap。
3. **tcp_connect 的 port：U16 → U32**：U16 实参的 ABI 提升行为未实证，为清零 ABI 风险改 U32（C 侧收敛到 u16_t）。
4. **tcp_send 的 bytes 走 view**：Text 实参不隐式转 [U8]，示例中 `view reqb: [U8] = req;` 显式借用。
5. **connect status 枚举勘误**：NonOS esconn_stat 实为 0=IDLE / 1=CONNECTING / 2=WRONG_PASSWORD / 3=NO_AP_FOUND / 4=CONNECT_FAIL / 5=GOT_IP（初稿「2 连接中」有误）；connect() 对 2/3/4 确定性失败早退。
6. **TCP 底座：netconn → raw API**：官方发布版 liblwip.a 把 netconn/sockets 层剥离（api_lib.o 空壳），只有 espconn 路径编进二进制。改用 raw API（tcp_pcb 不透明指针），回调收敛成环形缓冲 + 轮询（§5.2）；tcp_write 单次限 u16，发送缓冲满 = ERR_MEM 不重试。
7. **构建链两处 chip 级数据（esp8266 sdks 段）**：① `ir_text_section: ".irom0.text"`——IRAM 32K 被 libpp+libmain 固占约 23K，用户/std 代码必须驻留 flash（ICACHE_FLASH_ATTR 同语义），配 `-mtext-section-literals` 防 l32r 跨 IRAM↔IROM 越窗；② `ir_flags` 增补 `-O1`——call0 ABI 寄存器窗口窄，llc -O0 对高寄存器压力函数报 Cannot scavenge register（esp-llvm 已知 RA 缺陷）。
8. **esp8266/esp32/esp32s3 hello 示例更新**：批次 2 改造 io（print_text/print_nl → println）后示例未同步，本次修正为当前 API。

## 6. 验证与收尾

- 每批：host 示例实跑 + goldengen 快照重生成 + 全量 go test + `.local` 同步（stdlib/*.aila、bin/aila.exe、runtime 不分发宿主——确认 share/aila/runtime 布局同步三家族）。
- 嵌入式回归：esp8266/esp32/esp32s3 hello `--emit-ir` + app.bin 字节基线变化记录（stdlib 新增会改变 IR，属预期）。
- **示例清单两处硬编码**：新增 examples/<名> 要同步 build_test.go 与 tools/goldengen/main.go 的 exampleNames。gpio/uart/wifi 示例是嵌入式工程（放 esp8266/ 目录，不进 examples/，不进快照）。
- 文档：README std 表、skills 三份手写源（SKILL.md 标准库行 / language.md / recipes.md 新章节）、designs/0003 状态翻转、project_memory 记录。

## 7. 实现顺序

1. 查证语言能力：Text 构造路径（决定 to_text/split/args 形态）、`@region[r]` 返回逃逸检查现状、struct 方法 `@mut[self]`、`Text?` 语法。
2. 批次 1（五个纯库模块 + examples/stdlib_demo + 快照 + 测试）。
3. 批次 2（runtime C 六平台段 + 四模块 + 示例 + 链接错误信息优化）。
4. 批次 3（gpio/uart 三家族 + wifi SDK 胶水 + 嵌入式示例）。
5. 收尾（文档 + .local + memory）。
