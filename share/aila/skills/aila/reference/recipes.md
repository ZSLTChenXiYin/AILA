# AILA 任务配方（照抄即可，每步都能用 `aila check` 验证）

所有配方的前提：工作目录在该工程的包根（含 `aila.toml`），`aila` 已在 PATH 或已在用它。

---

## 1. 从零建一个能跑的程序

```bash
aila doctor                 # 先保证环境绿
aila new hello              # 生成 aila.toml + main.aila + README.md
cd hello
aila run                    # → hello, AILA!
```

`aila new` 生成的 `main.aila`：

```aila
import std.io;

func main() {
    io.println(io.FmtVal.T { v: "hello, AILA!" });
}
```

## 2. 打印各种值

打印只有三个函数（designs/0002 §4）：`print` 逐实参直接连接、`println` 再加换行冲洗、
`printf` 格式符驱动（`%d %u %f %b %c %s %%`，最小集无宽/精度）。所有实参用
`io.FmtVal` 包装箱传入——AILA 的带载荷 enum 本身就是 tagged union，窄类型先 `as`
拓宽到规范宽（I64/U64/F64/Bool/C32/Text）；格式符与箱型错配是编程错误，直接 `trap`。
遇写入错误静默吞；要拿写入字节数或换输出后端，用 `stdout()` 取 `Writer` 实例调
`put_*` 方法：

```aila
import std.io;

func main() {
    io.println(io.FmtVal.I { v: -42 });       // 带换行
    io.printf("f = %f ok = %b\n",
        io.FmtVal.F { v: 1.5 },
        io.FmtVal.B { v: true });
    io.print(io.FmtVal.T { v: "no newline" });
    var w: io.Writer = io.stdout();
    _ = w.put_text("via Writer\n");           // 返回写入字节数
}
```

要自己排版就把字节填进缓冲，再 `_ = w.put(buf);`（`Writer` 槽位直收 `[U8]`）。

## 3. 读命令行参数

`aila run -- a b c` 会把 `a b c` 传给程序；程序内用 `std` 之外的入口签名接收（当前批次
未提供 argv 绑定）。稳妥做法：把参数写进 `aila.toml` 的常量或环境变量，或用 C 的
`main(argc, argv)` 通过 `#compile.c` 绑定自取。

```bash
aila run -- 3 4        # 传递参数（分隔符 `--` 之后全部透传）
aila run --arg 3 --arg 4   # 等价写法
```

## 4. 加法/数值工具（含 trap 的边界处理）

```aila
import std.io;
import std.math;

func main() {
    // 算术回绕（C 语义）；确定性溢出检查显式走 #runtime.checked_add
    let w: U32 = math.wrapping_add_u32(4294967295, 1);
    io.println(io.FmtVal.U { v: #compile.as(U64, w) });            // 0

    let m: I32 = math.min_i32(3, 9);
    io.println(io.FmtVal.I { v: #compile.as(I64, m) });            // 3
}
```

## 5. 结构体 + 方法 + 按字段打印

```aila
import std.io;

struct Point { x: I32, y: I32 };

func Point.norm2(self) -> I32 {
    return self.x * self.x + self.y * self.y;
}

func main() {
    let p: Point = Point { x: 3, y: 4 };
    io.printf("norm2 = %d\n", io.FmtVal.I { v: #compile.as(I64, p.norm2()) });
}
```

## 6. 用枚举 + match 做分支（替代 if 表达式）

```aila
import std.io;

enum Level { Low, Mid, High };

func score_of(l: Level) -> I32 {
    return match l {
        .Low  => 1,
        .Mid  => 5,
        .High => 9,
    };
}

func main() {
    io.println(io.FmtVal.I { v: #compile.as(I64, score_of(Level.Mid)) });
}
```

## 7. 错误即值（`T ! E`）：造错 / 上抛 / 接住

```aila
import std.io;

#runtime.error enum ReadErr { Truncated { need: USize }, BadMagic };

func parse(buf: [U8]) -> I32 ! ReadErr {
    if #comprun.len(buf) < 4 {
        return .Truncated { need: 4 };
    }
    return 0x2A;
}

func relay(buf: [U8]) -> I32 ! ReadErr {
    let v: I32 = parse(buf)?;      // 错误原样上抛
    return v;
}

func main() {
    region work {
        let buf: [U8] = [0; 4]@work;
        match relay(buf) {
            let v: I32 => { io.printf("ok %d\n", io.FmtVal.I { v: #compile.as(I64, v) }); }
            let e: ReadErr => {
                match e {
                    .Truncated { need } => { io.println(io.FmtVal.U { v: #compile.as(U64, need) }); }
                    _ => io.println(io.FmtVal.T { v: "bad" }),
                }
            }
        }
    }
}
```

## 8. 用区域管理内存（没有 malloc/free）

```aila
import std.io;

func fill(n: USize, v: U8) @region[r] -> [U8] {
    return [v; n]@r;               // 分配落进调用方给的区域
}

func main() {
    region work {
        var buf: [U8] = fill(8, 7)@[work];
        buf[0] = 1;
        io.println(io.FmtVal.U { v: len buf as U64 });
    }                              // 出块整块回收
}
```

反例（会 `E4101` 逃逸）：`func bad() -> [U8] { region t { return [0; 4]@t; } }`。

## 9. 调用一个 C 库（不用手写绑定）

1. 建目录与文件：

```
demo/
  aila.toml
  main.aila
  cpacks/mathkit/include/mathkit.h
  cpacks/mathkit/mathkit.c
```

```c
/* cpacks/mathkit/include/mathkit.h */
#ifndef MATHKIT_H
#define MATHKIT_H
#define MATHKIT_MAX 1000
int mathkit_add(int a, int b);
#endif
```

```c
/* cpacks/mathkit/mathkit.c */
#include "mathkit.h"
int mathkit_add(int a, int b) { return a + b; }
```

2. 导入并使用（`bindgen` 由 `aila build/run/check` 自动跑，不用手动）：

```aila
import std.io;
import ch _.cpacks.mathkit;

func main() {
    let s: I32 = mathkit.mathkit_add(20, 22);
    io.println(io.FmtVal.I { v: #compile.as(I64, s) });                     // 42
    io.println(io.FmtVal.I { v: #compile.as(I64, mathkit.MATHKIT_MAX) });   // 1000
}
```

3. 想看生成的绑定：`aila bindgen`（→ `build/bindgen/mathkit.ail`）。
   有 C 声明没被映射时会给出 `W3001` 并逐条列出。

## 10. 跨模块/跨子包调用（可见性阶梯）

```
demo/
  aila.toml
  main.aila
  sub/util.aila
```

```aila
// sub/util.aila —— 轻量形态：顶层 public = 包外可见
public func double(v: I32) -> I32 {
    return v + v;
}
```

```aila
// main.aila
import std.io;
import _.sub.util;          // 不起别名 → 用最后一段 `util.`

func main() {
    io.println(io.FmtVal.I { v: #compile.as(I64, util.double(21)) });   // 42
}
```

import 绑定名恒为路径最后一段（`util.`）；没有别名语法（v2 退役）。
若 util 里改用 `module { }` 块，则函数必须写成「壳里 `public func double(...) -> I32;` + 块里实现」。

## 11. 定长数组 / 视图 / 长度

```aila
import std.io;

func main() {
    // 定长数组做 struct 字段或局部变量
    var xs: [I32; 3] = [1, 2, 3];
    xs[0] = 10;

    // 视图：从分配来
    region work {
        let buf: [U8] = [0; 4]@work;
        io.println(io.FmtVal.U { v: #compile.as(U64, #comprun.len(buf)) });  // 4（元素数）
    }

    // Text → 字节视图（等宽重标）
    let t: Text = "AILA";
    view bytes: [U8] = t;
    io.println(io.FmtVal.U { v: #compile.as(U64, #comprun.len(bytes)) });    // 4

    // 布局查询
    io.println(io.FmtVal.U { v: #compile.as(U64, #compile.sizeof(I32)) });   // 4
}
```

## 12. 交叉编译 / 换目标

```bash
aila target list                          # 已装工具链覆盖的架构 + 宿主三元组
aila target show xtensa-esp32s3-elf       # arch / cpu 候选 / 嵌入式判定 / 工具服务情况
aila check --triple xtensa-esp32s3-elf    # 只检查（按目标布局）
aila build --triple xtensa-esp32s3-elf    # 嵌入式三元组（-elf 后缀）直接出可烧录 app.bin（平台已适配）
# esp8266 单 app.bin；esp32/esp32s3 另出 bootloader.bin（自研二级 bootloader）：
#   esptool --chip esp32   write_flash 0x1000 bootloader.bin 0x10000 app.bin
#   esptool --chip esp32s3 write_flash 0x0    bootloader.bin 0x10000 app.bin
```

目标差异（指针宽度、`long`/`size_t` 映射）由 triple 决定，所以**先定 triple 再写代码**，
别先按宿主写完再切。交叉 C 编译需要 `toolchains/gcc/<triple>/`（`aila toolchain install gcc/<triple>`）。

### 12.1 嵌入式用 C（cpack 透传）

嵌入式 build 会把 cpack 的 C 源交给目标 GCC 交叉编译、`.a` 静态库原样透传进链接，
C 头文件放 `cpacks/<名>/include/`。固件里操作寄存器直接写 C：

```aila
import ch _.cpacks.blink;      // cpacks/blink/blink.c：寄存器级 GPIO 操作
blink.blink_init();
```

### 12.2 厂商 SDK 模式（以 ESP8266 NonOS SDK 为例）

项目 `aila.toml` 声明 `[sdk] name = "esp8266-nosdk"`，然后装 SDK（一次）：

```bash
aila toolchain install sdk/esp8266-nosdk@3.0.5   # 联网下载并按白名单裁剪
aila build --triple xtensa-esp8266-elf           # 出单文件 app.bin（直接烧 0x0）
```

SDK 模式下固件入口由 SDK 接管（ROM → `call_user_start` → `user_init`），AILA 代码经
runtime 的 `user_init` 启动；WiFi/网络能力用 **`std.wifi`**（仅 `[sdk]="esp8266-nosdk"`
可链接，其他 target import 会在链接期报未定义符号；底层为 lwip raw TCP、M1 单连接）：

```aila
import std.core;
import std.io;
import std.wifi;

func main() {
    wifi.mode_station();
    if wifi.connect("SSID", "PASS") {          // 阻塞至拿到 IP；false = 失败/超时
        let ip: U32 = wifi.ip();                // 网络序 u32，0 = 未拿到
        let req: Text = "GET / HTTP/1.0\r\n\r\n";
        view reqb: [U8] = req;                  // Text 传边界要显式借用
        match wifi.tcp_connect(0x01010101, 80) {   // 1.1.1.1（M1 无 DNS，网络序）
            let h: U64 => { _ = wifi.tcp_send(h, reqb); wifi.tcp_close(h); }
            let e: core.IoErr => { _ = e; }
        }
    }
}
```

- `tcp_send` / `tcp_recv` 返回 `I64`：0 = 成功/无数据，负值 = lwip err（recv 的 -1 =
  对端关闭）；`tcp_recv` 需 `@mut` 缓冲。
- 状态码细分用 `wifi.status()`（`ST_*` 常量，如 `ST_WRONG_PASSWORD`）。

完整可编译示例：esp8266/wifi_std（connect → 打印点分 IP → 发 HTTP GET → 收 128 字节
→ close）。SDK 库的 IRAM/IROM 布局、分区表、`-mlongcalls` 等全部来自随包 chip.json
数据，项目里零 SDK 配置文件。

## 13. 热替换模块（`module hot`）

```aila
module hot {
    state fuel: U32 = 0;             // 可迁移状态只能放 state

    func step(n: U32) {
        fuel = fuel + n;
    }
}
```

hot 模块禁止模块级 `var` 与 `@static`（E4006）。热替换的装载/迁移产物在当前批次
还没落地（`W9002` / E9002），写代码可以，跑起来先按普通模块验证逻辑。

## 14. 多文件工程 + 依赖别的项目

```toml
# aila.toml
[package]
name = "app"
version = "0.1.0"
entry = "_.main"

[dependencies]
mylib = { path = "../mylib" }                     # 本地
# mylib = { git = "https://...", tag = "v1.0.0" } # 远端（tag/rev 二选一锁死）
```

```aila
import mylib.format;        // 依赖名 = 导入根 = 包名
```

```bash
aila deps                   # 打印依赖图
aila deps --update          # 重写 aila.lock
aila deps --vendor          # 复制到 vendor/（离线构建）
```

## 15. 「只在链接/运行期炸了」怎么查

```bash
aila doctor                       # 环境：clang / llvm / 链接冒烟
aila version                      # AILA_HOME 与两条工具链的真实路径
aila build --verbose              # 看它到底调了哪条子进程、stderr 是什么
aila build --emit-ir              # 留 .ll 现场
aila ir main.aila                 # 单文件直接看 IR
aila doctor --json                # 机器可读的自检结果
```

## 16. 驱动一块单色屏（显示栈：画布 + 脏区 + 文本控制台）

128×64 OLED 这类屏没有字体引擎，**别在手写帧缓冲和整屏重发**——显示栈已经把这些拆好了。

```aila
import std.pixbuf;
import std.damage;
import std.paint;
import std.textgrid;
import std.display;
import _.font.font16;

func main() {
    let gs: paint.GlyphSet = paint.new_glyph_set(
        font16.font_cps(), font16.font_bits(), font16.font_widths(),
        font16.font_count(), font16.font_px_h());

    region reg {
        // 三块区域内存：画布 / 文本网格 / 脏区。句柄只是视图，不持有内存。
        var fb: [U8] = [0; 1024]@reg;                                  // 128×64 mono1
        var grid: [U32] = [0; textgrid.words(16, 4)]@reg;
        var dmg: [U32] = [0; damage.WORDS]@reg;

        var pb: pixbuf.Pixbuf = pixbuf.new(fb, 128, 64);
        var g: textgrid.TextGrid = textgrid.new(grid, 16, 4);
        var d: damage.Damage = damage.new(dmg);
        var c: display.Commit = display.new_commit(pb, d);

        // 画一行文本（全/半角由字库 widths 决定，绘制侧不做 CJK 判断）
        let s: Text = "温度 25C";
        view b: [U8] = s;
        textgrid.put_text(g, gs, b, #comprun.len(b));

        // 渲染：只把**真变了的 cell** 记进 d（滚动同理，绝不整屏重发）
        damage.clear(d);
        textgrid.render(g, gs, pb, d);

        // 图形图元自己记脏（网格之外才允许画图形）
        paint.hline(pb, 0, 127, 62, true);
        damage.add_rect(d, damage.rect(0, 62, 128, 63));

        // 提交：驱动自己实现 present(commit)，按 damage 算脏页
        c.set_damage(d);
        // my_driver.present(c);
    }
}
```

**要点**

- `Damage` / `TextGrid` 是**句柄**（`{ raw: [U32] }`）：先在 region 里 `[0; WORDS]@reg` 分配，再 `new(...)` 包一层；
  句柄按值传很便宜，大聚合体按值传在 xtensa 上会把后端顶崩（designs/0006 §4.1 第 8 条）。
- 刷新量 = 脏 cell → 脏页。128×64 一帧一行文字通常 **2 页 ≈5ms**（整屏 8 页 ≈37ms）。
- 宿主上直接 `aila run`，用 `std.display_host.sink_host_pbm/ascii` 把画布导出成 golden；
  **别把 `std.display_host` 带进固件**（它要 std.io，白占体积）。
- 设备选择用 `#compile.match DISPLAY { ... }`，配 `aila build --gc-sections`：
  未选中的驱动连符号都不进产物。
