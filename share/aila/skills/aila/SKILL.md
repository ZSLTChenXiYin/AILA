---
name: aila
description: 用打包好的 AILA 编译器（`aila` 命令）写、构建、运行、排错程序时使用。AILA 是面向嵌入式的静态编译语言，核心是区域内存模型（region）、错误即值（`T ! E`）与模块级热替换。当用户让你写或改 `.aila` 程序、构建/运行/交叉编译、调用 C 库、或解释 AILA 的报错时，先读本技能，再动代码。
---

# AILA：给 AI 编程工具的使用手册

你要做的是**用 `aila` 这个工具把程序做出来**。你不需要、也不应该去读 AILA 编译器自己的源码——
你手上能依赖的全部信息面就是：这份手册 + `aila` 命令的输出（诊断、`--json`、`aila explain`）。

## 0. 铁律（照做能省掉大半天）

1. **只用 `aila` 这一个入口**。不要找编译器源码，不要尝试自己模拟 AILA 语义。
2. **改完就 `aila check`**。它只做解析 + 语义，秒级返回、不产产物，是最快的验证手段。
3. **任何错误码先去 `aila explain <CODE>`**。它的「怎么修」通常就是答案，比猜快十倍。
4. **不确定某个写法是否被支持，就写最小片段跑 `aila check`**，别按别的语言经验硬猜。
5. AILA **没有**：泛型、trait/接口、union、异常、隐式类型转换、GC、`while`、`do`、
   裸块 `{ }`、`::`、`if`/`for` 取值。
6. 算术**回绕**（C 语义零检查）。确定性溢出检查显式走 `#runtime.checked_add/sub/mul`。
7. **一次只改一件事**，每步都跑 `check`。AILA 的诊断很细（带行号、插入符、错误码、修法），
   一次改太多会让报错互相掩盖。

## 1. 三十秒上手（解压即用，不需要装任何东西）

```bash
aila doctor          # 环境自检（含链接冒烟）：先让这一项全绿
aila new demo        # 生成骨架：aila.toml + main.aila
cd demo && aila run  # 构建并运行 → hello, AILA!
```

`aila` 是**单文件**可执行程序，运行时/标准库/带 clang 的工具链都在它旁边的相对目录里
（`aila version` 会打印 `AILA_HOME` 与两条工具链的实际路径）。换机器只需拷这个目录。

## 2. 标准工作循环

```
写/改 .aila  →  aila check（快，看诊断）  →  修  →  aila run [-- 参数]
```

| 想知道什么 | 用哪条命令 |
|---|---|
| 现在有没有错 | `aila check` | 
| 跑起来什么结果 | `aila run` |
| 诊断要机器可读（你要解析它） | 任意命令加 `--json` |
| 错误码是什么意思 | `aila explain E3201` |
| 词法层怪问题 | `aila parse main.aila --tokens` |
| 语法树（写工具时） | `aila parse main.aila`（S-表达式） |
| 生成的 LLVM IR | `aila build --emit-ir` 或 `aila ir main.aila` |
| 环境/工具链问题 | `aila doctor`、`aila version`、`aila toolchain info` |
| 目标平台 | `aila target list`、`aila target show xtensa-esp32s3-elf` |

## 3. 工程形状

```
demo/
  aila.toml            # 清单：name / version / entry / [target] triple
  main.aila            # 入口模块（entry = "_.main"，模块名 = 文件名 = 路径即身份）
  util.aila            # 同包模块   → import _.util;
  sub/format.aila      # 子包       → import _.sub.format;
  cpacks/mathkit/      # C 包       → import ch _.cpacks.mathkit;
```

- `_` = 当前项目根；`std` = 标准库（`import std.io;`）；`ch` / `cc` = C 包。
- 模块名与子包名必须是标识符 `[A-Za-z_][A-Za-z0-9_]*`（写 `my-lib` 会报 E7003）。
- import 绑定名 = 路径最后一段（`std.io` → 用 `io.println`）；没有别名语法（v2 退役）。

## 4. 可见性阶梯（**最高频的坑**）

三级、单向：**文件内 ⊂ 子包内 ⊂ 包外**；`public` 的语义恒为「提一级」。

| 写在哪 | 默认可见 | 加 `public` 后 |
|---|---|---|
| `module { }` 块内 | 文件内 | 子包内 |
| `module` 外面（壳） | 子包内 | 包外 |
| 没有 `module` 块的文件（轻量形态） | 按壳算 | 包外 |

两条推论，记牢：

1. **同子包另一个文件**要用 → 声明处加 `public`（块内）或直接放壳里。
2. **跨子包 / 跨项目**要用 → 必须是「壳里 + `public`」。函数的标准写法是
   **壳里前向声明 + 块里实现**：

```aila
public func twice(v: I32) -> I32;        // 壳：接口（对外可见）

module {
    func twice(v: I32) -> I32 {          // 块：实现（签名必须一致）
        return v + v;
    }
}
```

只写前向声明不写实现 → `E7031`；签名对不上 → `E3201`。引用不到 → `E7020`（诊断里会告诉你加在哪一级）。

## 5. 语言最小可用子集（够你写出真程序）

```aila
import std.io;

const MAX: I32 = 100;                    // 常量（壳里 public 才跨包）

struct Point { x: I32, y: I32 };         // 结构体
enum Color { Red, Green, Blue };         // 枚举（可带载荷：Bad { code: I32 }）

func Point.norm2(self) -> I32 {          // 方法 = 类型前缀式函数
    return self.x * self.x + self.y * self.y;
}

func clamp01(v: F32) -> F32 {            // 普通函数
    var m: F32 = v;
    if m < 0.0 { m = 0.0; }              // if 恒为语句，不产出值
    if m > 1.0 { m = 1.0; }
    return m;
}

func main() {                            // 入口：func main()，可以没有参数
    var p: Point = Point { x: 3, y: 4 }; // 构造字面量
    let n: I32 = p.norm2();              // let 深不可变；var 可写
    io.printf("norm2 = %d\n", io.FmtVal.I { v: #compile.as(I64, n) });

    // 循环三形态
    for var i: I32 = 0; i < 3; i = i + 1 { _ = i; }   // 三段式（首槽只能是一条 var）
    var done: Bool = false;
    for !done { done = true; }                         // 条件循环（没有 while）
    let xs: [I32] = [10, 20, 7]@static;                // 定长数组 + 分配标注
    for i, x in xs { _ = i + x; }                      // for-in（i 下标、x 元素拷贝）

    // 取值只能靠 match（没有 if 表达式）；变体模式一律 leading-dot
    let name: Text = match Color.Green {
        .Red   => "red",
        .Green => "green",
        .Blue  => "blue",
    };
    _ = name;

    // 整数族是开放值域，match 必须带 `_`
    let k: I32 = match n {
        0 => 100,
        _ => 0,
    };
    _ = k + MAX;
}
```

要点速查（细节见 `reference/language.md`）：

| 主题 | 一句话 |
|---|---|
| 类型 | `I8..I64`/`U8..U64`/`F32`/`F64`/`Bool`/`C32`/`Text`/`USize`/`ISize`；`[T]` 视图、`[T; N]` 内联数组、`chan`、`coro`、`Unsafe` |
| 转换 | 显式转换只有 `#compile.as(T, x)`（纯重解释）/ `#runtime.as(T, x)`（带检查）。**零隐式转换**，`U8`+`I32` 必须显式转 |
| 层系统 | `#compile`（编译期）/ `#comprun`（双态）/ `#runtime`（运行时）三层；层内不存在的 API = E9204 |
| 内存 | 只有三条来路：`region 名字 { }` 块、`@region[r]` 区域参数、`@static`。没有 `malloc/free` |
| 分配 | `[0; 4]@work`（裸名落到区域）、`f(...)@[work]`（区域实参通道）。两种 `@` 不能混用 |
| 视图 | `let t: Text = "AILA"; view b: [U8] = t;` 等宽重标；`#comprun.len(b)` 取长度 |
| 变参 | 末位参数 `rest: ...T`，体内 `rest` 即 `[T]` 只读视图（for-in / `#comprun.len` / 下标直接用）；C 边界另有无名 `...`（仅 `#compile.c` 声明） |
| 长度 | `#comprun.len(x)`，不是 `x.len()` |
| 错误 | `-> T ! E`（E 须 `#runtime.error`/`#comprun.error` 标注 enum）：`return .成员;` 造错、`f()?` 上抛、`match f() { let v: T => .., let e: E => .. }` 接住 |
| 丢弃 | `_ = f();`（有返回值的函数不能裸调用，`E3206`） |
| 布局 | `#compile.sizeof(T)`、`#compile.alignof(T)` |
| 陷阱 | `unsafe { }` 才是放宽区（改宽度重标等）；`defer { }` 收尾；`#runtime.trap("…")` 不可恢复终止 |
| C 互操作 | 放 `cpacks/名字/{include/*.h, *.c}` → `import ch _.cpacks.名字;` → 直接 `名字.C函数(...)`，**不用手写绑定** |
| 标准库 | 三层清单见 `reference/language.md` §12。常用：`std.io`（打印三函数 `print/println/printf`，实参 `io.FmtVal.I/U/F/B/C/T { v: ... }` 装箱）、`std.math`、`std.strconv/strings/utf8/binary/sort`（通用层）、`std.time/os/rand/path`（仅 OS target）、`std.gpio/uart`（三家族嵌入式）、`std.pixfmt/pixbuf/damage/paint/textgrid/display`（显示栈：点阵画布 + 脏区 + 文本控制台 + commit 提交；`std.display_host` 是宿主调试出口，不进固件）、`std.wifi`（仅 esp8266 `[sdk]="esp8266-nosdk"`）；别假设有 `std.str`/`std.mem`/`std.atomics`（原子 = `#runtime.atomic_*`） |
| 并发 | 已落地：`coroutine func` + `yield`（启动 `#runtime.coro(f, args)`，eager 跑到首个挂起点）、`#runtime.chan(T, 容量)` SPSC 双端、`#runtime.spawn(f, args)`（宿主）、`#runtime.atomic_*` 原子原语、`select` 统一等待（`case v = <-rx` / `case r = resume h`）。协程事件是「等完成」：绑定值只有 `T ! E` / `T`，**没有** `Suspended` / `Done` 分支，`resume h` 也不能单独当表达式（E5010）；范例 `examples/concurrency_demo` |

## 6. 高频报错速查（完整表见 `reference/errors.md`）

| 现象 | 修法 | 码 |
|---|---|---|
| 类型不匹配 | 别指望自动转换：`#compile.as(I64, x)` | E3201 |
| `T ! E` 的 E 不是错误集 | 给 enum 加 `#runtime.error` / `#comprun.error` 标注 | E3201 |
| 名字未定义 | 拼写；跨模块先 `import` 再 `模块.名字` | E3001 |
| 不可见 | 按「可见性阶梯」加 `public` 或挪到壳里 | E3002 / E7020 |
| 壳里声明没有实现 | 块里补同名同签名 `func`，或删掉声明 | E7031 |
| match 不完备 | 补齐变体；整数族加 `_` 且放最后 | E3401 / E3403 |
| 有值表达式被当语句 | 接住返回值或写 `_ = f();` | E3206 |
| 只读绑定被写入 | 声明处 `var`；跨函数写要 `@mut[参数]` | E4201 |
| 分配点没有可见区域 | 加 `@region[r]` 参数，或体内 `region tmp { }` | E4001 |
| 视图寿命不足（逃逸） | 别返回体内 region 块的视图；挪到区域参数或 `@static` | E4101 |
| 语句缺 `;` | 声明/语句（除块语句）都要分号 | E2009 |
| 这个位置不接受声明 | 带函数体/运行期对象的声明只能进 `module { }` | E2010 |
| 比较链 `a<b<c` | 写 `a < b && b < c` | E2004 |
| 未知层 API | 查 §5 层系统表：`#compile/#comprun/#runtime` 各自的 API 清单 | E9204 |

## 7. 排错剧本

1. 读诊断：`文件:行:列` + 错误码 + 「这里」插入符 + 「怎么修」。
2. 拿码查详解：`aila explain <CODE>`（`--json` 给机器）。
3. 还是不懂：把**最小复现**写成小文件，`aila check` 它；必要时 `aila parse --tokens` 看词法。
4. 怀疑环境：`aila doctor`（它会把 clang/llvm/链接各跑一遍），`aila version` 看 `AILA_HOME`。
5. 只在链接/运行期炸：`aila build --emit-ir` 留产物，`aila build --verbose` 看子进程命令与 stderr。

## 8. 需要更多细节时

| 文件 | 什么时候读 |
|---|---|
| `reference/language.md` | 要确认某个语法/类型/内存规则的确切写法 |
| `reference/recipes.md` | 要按任务抄配方：新建工程、读写、C 库、跨包、交叉编译、热替换 |
| `reference/manifest.md` | 要改 `aila.toml`（目标三元组、依赖、域、工具链选择） |
| `reference/cli.md` | 要确认某条命令的参数与退出码（生成物，与实现同步） |
| `reference/errors.md` | 要查某个 `E****`/`W****` 的含义与修法（生成物，与实现同步） |

命令随时可自查：`aila --help`、`aila <子命令> --help`。
