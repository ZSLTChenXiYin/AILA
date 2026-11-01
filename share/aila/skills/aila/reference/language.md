# AILA 语言速查（AI 编程工具用）

> 权威来源是 `docs/designs/0001-language-spec.md`（语义）与各 `docs/designs/000N`（机制）。
> 本文件是**当前实现**的浓缩版：写法以它为先行，遇到冲突以编译器的行为为准。
>
> 复用纪律：标准库已经有的，不要重写（`stdlib/*.aila` 全是 AILA 源码，直接读最快）。

## 1. 文件与模块

- 一个文件一个模块，模块名 = `import` 名：根包 `_.<子目录>.<文件>`；标准库 `std.<模块>` → `stdlib/<模块>.aila`。
- 包根只放 `main.aila`；其它文件必须进子目录（`specs/0001 §1`）。
- 文件是**整块形态**（壳 + `module { }`）**或**整块放在顶层（无 `module` 块时 `func` 可以直接带体）——**两者不能混**（E2010）。
- 壳（`module` 块外）只放：`const` / `type` / `struct` / `enum` / 无体 `func` 声明 / `#compile.c func` 外部声明。

```aila
import std.io;
import _.font.font16;        // 跨子包
import cc _.cpacks.ssd1306;  // C 包（ch|cc：按扩展名分派）

public func ping() -> U8;    // 壳：对外可见的实现体声明

module {                     // 实现必须落在 module 块内
    public func ping() -> U8 { return 1; }
}
```

## 2. 可见性

- 默认**包内可见**；`public` = 包外（含根子包）可见——跨包可见只有这一档，没有逐字段可见性。
- `import` 不做非限定展开：跨模块引用必须带模块名前缀（`io.printf` / `font16.font_cps` / `core.IoErr`）。

## 3. 声明

| 形态 | 说明 |
|---|---|
| `let x: T = v;` | 只读绑定（**深不可变**） |
| `var x: T = v;` | 读写绑定 |
| `const NAME: T = 字面量;` | 编译期常量；初始化器**只能是纯字面量表达式**（`const B: U8 = A * 2` 不合法） |
| `view b: [U8] = text;` | 变型重标视图（**目标类型必须与源不同**；同型用 `let`，E4004） |
| `struct X { 字段: 类型, };` | 纯数据；字段可内联定长数组 `[T; N]` |
| `func X.m(self, ...)` | 方法写在 struct 外，类型前缀式 |

## 4. 函数与方法

签名三段式：**参数 = 数据流、`@` 段 = 效果、`->` = 产出**。

```aila
func fill(dst: [U8], v: U8) @mut[dst];              // 写权
func make(n: USize) @region[r] -> [U8];             // 分配许可（体内可 @r 分配）
func consume(m: M) @move[m];                        // 转移
func mix(a: [U8], b: [U8], n: USize) @region[r] @mut[a] @mut[b] @region[r];  // 顺序固定
```

- `@` 顺序固定 `@region → @mut → @move`；同一参数不得既 `@mut` 又 `@move`。
- **效果不传染**：调用不要求调用方具备被调方的效果声明（纯函数天然调不了需要供给的函数）。
- `@mut[self]` 方法按指针传 self，**写回调用方内存**；但**不要在方法里通过 self 再调本模块的其它方法**（转发写权是已知缺口 → 用自由函数 + 传值兜住）。
- 大聚合体（>16 字节的纯数据 struct）跨函数**按只读指针**传（实现约定，语义仍是拷贝）；`@mut[x]` 的同型参数同样按指针传，因此大聚合体的写会落回调用方。**≤16 字节的小结构体仍是按值复制**：它的标量字段写不回，可变接口必须用 `@mut[self]` 方法。
- 大状态建议做成**句柄 + 区域内存**（`struct Damage { raw: [U32] }`），句柄两个机器字，写经 `raw` 切片落地（`std.damage` / `std.textgrid` 就是这么做的）。
- 变参：`f(args: ...FmtVal)`，末位；`for v in args` 只读迭代。

## 5. 类型

- 基元：`I8..I64` `U8..U64` `F32` `F64` `Bool` `C32` `Text` `USize` `ISize` `Unsafe`。
- 视图家族（同为 `(ptr, len)`）：`[T]` / `Text` / `[Unsafe]`；`[T; N]` 是**内联定长数组**（值语义，不参与区域绑定）。
- 没有泛型、trait、接口、union、`List<T>`；没有隐式转换。
- `T?` 仅句柄类（enum / struct 句柄）；`struct` 是唯一自定义复合类型，enum 可带载荷（tagged union）。

## 6. 层系统：`#compile` / `#comprun` / `#runtime`

`#` = 服务请求，三层封闭（层名保留，用户不得占用）；层内不存在的 API = E9204。

| 层 | 语义 | v0 API |
|---|---|---|
| `#compile` | 纯编译期求值（运行时上下文 = E9201） | `sizeof(T)` `alignof(T)` `as(T, x)`（纯重解释）；声明指令 `c` / `inline` / `packed` / `align(N)`；剪枝指令 `if` / `match` |
| `#comprun` | 编译期可折叠则折叠，否则运行时 | `len(v)`；声明指令 `error` |
| `#runtime` | 强制运行时执行（常量位 = E9202） | `as(T, x)`（带检查/trap）、`trap(msg)`、`checked_add/sub/mul`、`atomic_*`、`chan(T, n)`、`spawn(f, args)`、`coro(f, args)` |

```aila
let n = #compile.sizeof(Dev);
let l = #comprun.len(buf);
let (cv, ov) = #runtime.checked_add(a, b);   // (T, Bool)：溢出标志不 trap，调用侧自决
```

- **显式转换只有 `#compile.as(T, x)` / `#runtime.as(T, x)`**；`as` / `trunc` / `len` / `trap` 这些关键字已退役。
- 长度只有 `#comprun.len(v)`：`[T]` = 元素数，Text = 字节数，`[Unsafe]` = 字节长。

### 编译期剪枝：`#compile.if` / `#compile.match`（语句级）

条件是**编译期常量表达式**：字面量、`const`（含 import 进来的）、`#size(T)`/`#align(T)`、函数内一次性
`let` 绑 const-expr、以及仅由这些构成的纯运算。命中分支正常生成代码，其余分支**只在代码生成阶段被省略**
（解析 / 类型检查 / 区域与能力分析 / 告警照做，死代码不会腐烂）；求不出常量就**报错不降级**
（`E3210`–`E3213`）。

```aila
const DISPLAY: U8 = 1;

func main() {
    #compile.match DISPLAY {
        1 => ssd1306_run(),          // 只有这一支进产物
        2 => st7735_run(),
        _ => {},
    }
}
```

- 剪枝**不改 ABI / 模块接口指纹**；`#compile.match` 不必穷尽，被剪臂仍完整类型检查；
- 只依赖源码常量，没有 `-D`、没有 manifest `[cfg]`；
- 报告：`aila build --emit-json` 的 `compile_time_prunes`；配 `--gc-sections` 连未选中模块的符号都不进产物。

## 7. 内存：区域、分配、视图、move

没有 `malloc` / `free` / 析构。内存三条来路：

```aila
region work {                              // ① 体内区域块：出块整块回收
    var buf: [U8] = [0; 16]@work;          //    定长
    var dyn: [U8] = [0; n]@work;           //    运行期长度
    let one: [U8] = [0]@work;              //    单值内存（C *U8 对应物）
    buf[0] = 7;
}

func make(n: USize) @region[r] -> [U8] {   // ② 区域参数：寿命由调用方给
    return [0; n]@r;
}

var table: [U8] = [0; 16]@static;          // ③ 静态区：寿命无限
```

- 引用的三条来路：分配（`[...]@r`）、同型借用（`let v: [U8] = a;`）、变型重标（`view`）。栈值**绑不成** `[T]`。
- 逃逸检查位：返回 / 存入更长寿命的位置 / 跨并发捕获。返回 `@static` 或参数视图合法；返回体内 region 块的视图 = 编译错误。
- `copy x @[r]` 显式深拷贝；`move x` 显式转移主名字（内存不动）。

## 8. 控制流与 match

- **没有 `while`**：`for 条件 { }`；无限循环 `for { }`；`for i in 0..n`（白名单迭代）、`for v in 数组`。
- `if` 不产出值；需要值用 `match`。裸块 `{ }` 非法（E2002）。
- `break` / `continue` 是 C 语义；`defer` 块级、LIFO，trap 不触发。
- `match` 是唯一多路分派与解构手段：`match v { .I { v } => ..., _ => ... }`；判别值域覆盖不全要 `_`。

## 9. 错误与可空

```aila
public enum Err { Empty, Invalid { at: USize }, Range };   // 错误集（#runtime.error 定空间）

public func parse(t: Text) -> I64 ! Err {                 // T ! E
    if n == 0 { return .Empty; }                          // 错误臂：只写变体
    return v;                                             // 成功臂：直接返回值
}

let x: I64 = parse(s)?;                                   // ? 上抛（= 提前 return）
```

- `T?` = `None` / `Some { v: ... }`（句柄类才有）。
- `#runtime.trap("…")` = 不可恢复终止；算术溢出 / 除零 / 越界 / 移位超宽默认就是 trap（`#runtime.checked_*` 才拿得到标志）。

## 10. 并发

`#runtime.spawn(f, args)`（thread，fire-and-forget）/ `#runtime.coro(f, args)`（协程句柄 + `resume`）/
`#runtime.chan(T, n)`（双端 SPSC，`<-` 收发）/ `select` + `case` / `#runtime.atomic_*`（`[Unsafe]` + 偏移，全 seq_cst）。

## 11. C 互操作

- `#compile.c func name(参数) -> 返回;` 声明外部符号；C 源码/头放 `cpacks/<名字>/`，`import cc _.cpacks.<名字>` 一起编译链接。
- `[U8]` 跨 C 边界**只传指针**（`(ptr, len)` 里的 len 不展开）：要么把指针绑一次（bind 式），要么形参照抄 `(ptr, len, ...)` 形状。
- `T*` → `[T]?`，`char*` → `[U8]?`；C 结构体直译成 struct（自然对齐），`union` 永不直译。

## 12. 标准库现状（三层规划 designs/0003，全部已落地）

| 层 | 模块 | 提供 |
|---|---|---|
| 通用 | `std.core` | fill 引擎（`fill_u64` `fill_i64` `fill_f64`）、`IoErr`（`Closed` `NoSpace{need}` `Timeout` `NotFound` `Denied`） |
| 通用 | `std.io` | `print/println/printf`（`io.FmtVal.I/U/F/B/C/T { v: ... }` 装箱，`%d %u %f %b %c %s %%`）；vtable `Writer`（`stdout()` + `put_text/put/put_u64/put_i64/put_f64/put_bool/put_char/put_nl`） |
| 通用 | `std.math` | `wrapping_add_u32` `wrapping_add_i32` `abs_i32` `min_i32` `max_i32` `fmod` |
| 通用 | `std.strconv` | `format_i64/u64/f64`、`parse_i64/u64/f64`（`! Err`） |
| 通用 | `std.strings` | `len_runes/contains/index/has_prefix/has_suffix/bounds` + `concat/split/join/replace/upper_ascii/lower_ascii`（写进调用方缓冲） |
| 通用 | `std.utf8` | `encode` / `rune_len` / `cp_at` / `valid` |
| 通用 | `std.binary` | U16..U64 / I16..I64 × LE/BE 的 `le_/be_read_*`、`le_/be_write_*` |
| 通用 | `std.sort` | `sort_i64/u64/f64` + `_by` 版（`cmp(a,b) -> Bool` 函数指针） |
| 宿主 | `std.time` | `mono_ms()` / `sleep_ms(ms)` / `sleep_us(us)`（I2C 位时序、单总线协议：三家族走 ccount，宿主先睡整毫秒再自旋） |
| 宿主 | `std.os` | 文件 `open/read/write/close`（`! IoErr`）+ `args_count/args_get` + `env/env_set`；**只有 OS target 可链接** |
| 宿主 | `std.rand` | `Rand{s}`（`next_u64` / `below(n)`，xorshift） |
| 宿主 | `std.path` | `base/dir/ext` + `join` |
| 嵌入式 | `std.gpio` | `MODE_IN/MODE_OUT` + `mode/write/read`（三家族寄存器直写） |
| 嵌入式 | `std.uart` | `init(id, baud)` / `write` / `read`（stdout 同路 UART0） |
| 显示栈 | `std.pixfmt` | `mono1()` / `of(code)` / `row_bytes(fmt,w)` / `stride_for(fmt,w)` / `byte_of` / `bit_of`；1bpp、行 4 字节对齐、MSB 在左 |
| 显示栈 | `std.pixbuf` | `Pixbuf{data,w,h,stride,fmt}` + `new/new_of/cap_bytes/row_bytes/clear/fill/set_px/get_px/cols/rows`；cell = 8×16 像素 |
| 显示栈 | `std.damage` | 脏区句柄 `Damage{raw}`（`[0; damage.WORDS]@reg`）+ `new/clear/add_rect/add_cell/has_cell/cell_count/clip_count/clip/union_into/bounds`；`Rect{x0,y0,x1,y1}` 半开 |
| 显示栈 | `std.paint` | `GlyphSet{cps,bits,widths,count,px_h}` + `glyph_index/glyph_width/fill/fill_rect/hline/vline/blit_glyph/text_cells/cell_rect/decode` |
| 显示栈 | `std.textgrid` | 控制台句柄 `TextGrid{raw}`（`[0; textgrid.words(cols,rows)]@reg`）+ `new/clear/get/set/scroll_by/newline/put_cp/put_text/render/render_all`；滚动 = 内容模型 + cell 级 diff |
| 显示栈 | `std.display` | `Commit{fb,dmg,enable}` + `new_commit(fb,dmg)` / `set_damage` / `damage_of` / `set_buf` / `enabled`：**唯一推送通道**（驱动侧实现 `present(commit)`） |
| 显示栈 | `std.display_host` | `sink_host_pbm(commit)` / `sink_host_ascii(commit)`——宿主导出专用，**不进固件** |
| esp8266 | `std.wifi` | **仅 `[sdk] = "esp8266-nosdk"`**：`mode_station/connect/status/ip` + TCP `tcp_connect/tcp_send/tcp_recv/tcp_close` |

没有 `std.str` / `std.mem` / `std.vec` / `std.atomics`（原子原语 = `#runtime.atomic_*`）。需要就把实现写进用户项目里。

## 13. 「别的语言不是这样」的点

1. `if` 不产出值 → 用 `match` 或「占位 + if 赋值」。
2. 没有 `while` → `for 条件 { }`；裸块 `{ }` 非法 → `region 名字 { }`（E2002）。
3. `let` 深不可变；`var` 才可写；跨函数改写要 `@mut[参数]`（E4201）。
4. 零隐式转换：`#compile.as(T, x)` 显式；窄化 / 浮点↔整数走 `#runtime.as`（带检查）。
5. 算术默认 trap（溢出 / 除零 / 越界 / 移位超宽），要回绕就写 `#runtime.checked_*` 或 `std.math.wrapping_*`。
6. 保留字：`out` / `copy` / `move` / `view` / `region` / `import` / `module` 都不能当变量名。
7. 函数实现必须进 `module { }` 块（只要文件里出现了该块）；壳里只放声明。
8. 模块级 `let X: Text = "…"` 当前是空串 → 文本常量放函数内（`let s: Text = "…"; view b: [U8] = s;`）。
9. `const` 初始化器必须是纯字面量（不引用别的 const、不做运算）。模块级 `let/const X: Text = "…"` 现在**已经可用**（此前是空串的 bug 已修）。
10. `[T; N]` 与 `[T]` 不是一种类型：传参前先 `view` 重标（同型用 `let`）。
11. trap / 越界 / 溢出都会终止进程（嵌入式上 = 重启），所以**库内一律夹住越界再画/写**。
12. **跨模块内联的边界只在 `module hot`**：非 hot 模块之间允许被优化器内联（`aila build --lto` 把整程序合成一个模块再 -O2）；热替换模块的函数带 `noinline optnone`。所以想让"某段代码一定不存在"，仍用 `#compile.if/match` 剪枝（它不依赖优化器）。
13. 没有方法重载、没有默认参数、没有具名实参；同名函数不能重载 → 用 `类型.名` 或不同前缀区分。
