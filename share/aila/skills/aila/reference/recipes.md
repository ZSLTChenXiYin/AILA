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
    io.print_text("hello, AILA!");
    io.print_nl();
}
```

## 2. 打印各种值

标准库目前只有 `std.io` 的打印函数（无格式化字符串）：

```aila
import std.io;

func main() {
    io.print_text("n = ");
    io.print_i32_nl(-42);            // 带换行
    io.print_text("f = ");
    io.print_f64(1.5);
    io.print_nl();
    io.print_text("ok = ");
    io.print_bool(true);
    io.print_nl();
}
```

要自己排版就先用 `region` 分配一个字节缓冲，填好后 `io.print_bytes(buf)`。

## 3. 读命令行参数

`aila run -- a b c` 会把 `a b c` 传给程序；程序内用 `std` 之外的入口签名接收（当前批次
未提供 argv 绑定）。稳妥做法：把参数写进 `aila.toml` 的常量或环境变量，或用 C 的
`main(argc, argv)` 通过 `#[c]` 绑定自取。

```bash
aila run -- 3 4        # 传递参数（分隔符 `--` 之后全部透传）
aila run --arg 3 --arg 4   # 等价写法
```

## 4. 加法/数值工具（含 trap 的边界处理）

```aila
import std.io;
import std.math;

func main() {
    // 默认算术在溢出时 trap；要回绕就显式用 std.math
    let w: U32 = math.wrapping_add_u32(4294967295, 1);
    io.print_u32_nl(w);                 // 0

    let m: I32 = math.min_i32(3, 9);
    io.print_i32_nl(m);                 // 3
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
    io.print_text("norm2 = ");
    io.print_i32_nl(p.norm2());
}
```

## 6. 用枚举 + match 做分支（替代 if 表达式）

```aila
import std.io;

enum Level { Low, Mid, High };

func score_of(l: Level) -> I32 {
    return match l {
        Level.Low  => 1,
        Level.Mid  => 5,
        Level.High => 9,
    };
}

func main() {
    io.print_i32_nl(score_of(Level.Mid));
}
```

## 7. 错误即值（`T ! E`）：造错 / 上抛 / 接住

```aila
import std.io;

enum ReadErr { Truncated { need: USize }, BadMagic };

func parse(buf: [U8]) -> I32 ! ReadErr {
    if len buf < 4 {
        return _ ! ReadErr.Truncated { need: 4 };
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
            Ok  { v } => { io.print_text("ok "); io.print_i32_nl(v); }
            Err { e } => {
                match e {
                    ReadErr.Truncated { need } => { io.print_usize(need); io.print_nl(); }
                    _ => io.print_text("bad"),
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
        io.print_usize(len buf);
        io.print_nl();
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
    io.print_i32_nl(s);                     // 42
    io.print_i32_nl(mathkit.MATHKIT_MAX);   // 1000
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
    io.print_i32_nl(util.double(21));   // 42
}
```

要别名：`import _.sub.util as u;` → `u.double(21)`。
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
        io.print_usize(len buf);        // 4（元素数）
        io.print_nl();
    }

    // Text → 字节视图（等宽重标）
    let t: Text = "AILA";
    view bytes: [U8] = t;
    io.print_usize(len bytes);          // 4
    io.print_nl();

    // 布局查询
    io.print_usize(#size(I32));         // 4
    io.print_nl();
}
```

## 12. 交叉编译 / 换板子

```bash
aila target list                 # 看内置 board
aila target show esp32s3         # arch / triple / 预置物理域
aila check --board esp32s3       # 只检查（按目标布局）
aila build --board esp32s3       # 交叉目标当前停在 .o（E8020 说明链接待批次）
```

板级差异（指针宽度、`long`/`size_t` 映射）由 board 决定，所以**先切 board 再写代码**，
别先按 host 写完再切。

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
