# AILA 语言速查（写给 AI 的精确版）

只写**当前批次真实可用**的写法。凡是这里没出现的能力，先写最小片段跑 `aila check` 确认，
不要假设它有。

---

## 1. 文件与模块

- 一个 `.aila` 文件 = 一个模块；**模块名 = 文件路径**（`sub/format.aila` → `_.sub.format`）。
- 一个模块 = **壳**（`module { }` 外面）+ **块**（`module { }` 里面）。
  - 壳只放纯声明：`import`、类型、`const`、`#[c]` 无体函数、有体函数的**前向声明**。
  - 块放实现：带函数体的 `func`、`var`、`state`、`view`、`region` 等运行期对象。
  - 顺序固定：`import` → 类型/`const`/`#[c]` 声明 → `module { }` 块（写反会报 E2010 / E2001）。
- 轻量形态：整个文件**没有** `module { }` 块时，顶层声明按「块」语义处理——
  顶层 `func` **可以带函数体**，也允许顶层 `var`/`state`。
- ⚠️ 两种形态不能混：只要文件里出现了 `module { }` 块，块外任何**带函数体**的 `func`
  都会报 `E2010`。所以要么「全程轻量」（不写 `module`），要么「壳只留声明、实现全进块」。
- `module hot { }` 是热替换模块：**禁止模块级 `var` 与 `@static` 分配**（E4006），
  可迁移状态只能放 `state`。

## 2. 可见性

三级、单向：文件内 ⊂ 子包内 ⊂ 包外；`public` 恒为「提一级」。

| 写在哪 | 默认 | `public` 后 |
|---|---|---|
| 块内（`module { }`） | 文件内 | 子包内 |
| 壳内 | 子包内 | 包外 |

跨子包/跨项目 = 只认「包外」那一级，所以函数要写「壳内前向声明 + 块内实现」：

```aila
public func twice(v: I32) -> I32;   // 壳：接口

module {
    func twice(v: I32) -> I32 {     // 块：实现（签名必须一致）
        return v + v;
    }
}
```

`E7031` = 声明没配到实现；`E7020` = 引用的声明在此不可见（诊断会指明缺哪一级）。

## 3. 声明

```aila
const LIMIT: I32 = 100;              // 常量（必须初始化）
var counter: I32 = 0;                // 模块级可变（仅壳里/顶层）
state fuel: U32 = 0;                 // 热替换模块的可迁移状态
let n: I32 = 1;                      // 局部只读（深不可变）
var m: I32 = 1;                      // 局部可变

struct Point { x: I32, y: I32 };                    // 注意结尾分号
enum Color { Red, Green, Blue };                     // 无载荷
enum Err { Bad, Truncated { got: USize } };          // 带载荷变体（花括号）

type Id = U32;                                       // 类型别名
view bytes: [U8] = text;                             // 视图重标（必须写目标类型）
```

- `let` 没有类型标注时按初始化器推断；没有初始化器就必须写类型。
- 名字不能用关键字与内置类型名（`I32`、`Text`、`chan` …）→ E3009。

## 4. 函数与方法

```aila
func f(a: I32, b: I32) -> I32 { return a + b; }   // 有返回值的每条路径都要 return
func g() { }                                      // 无返回值：不写 `-> T`
func P.m(self, v: I32) -> I32 { return self.x + v; }  // 方法 = 类型前缀
func f(v: I32) -> I32 ! Err { ... }               // 可恢复错误通道
func fill(n: USize) @region[r] -> [U8] { return [0; n]@r; }  // 区域参数
#[c] func c_add(a: I32, b: I32) -> I32;           // C 绑定声明（无函数体，只能进壳）
```

- `self` 免标注，类型就是前缀类型。
- 效果子句顺序固定：`@region[...]` → `@mut[...]` → `@move[...]`；同类多参写 `@mut[a, b]`。
- 变参：`#[c] func printf(fmt: Text, ...) -> I32;`（只用于 C 边界）。
- `-> (T1, T2)` 产多值组，`return (1, 2);` 产出，解构赋值要先声明变量：

```aila
func two() -> (I32, I32) { return (1, 2); }

func main() {
    var a: I32 = 0;
    var b: I32 = 0;
    (a, b) = two();          // 组解构（注意：不是 `let (a, b) = ...`）
    _ = a + b;
}
```

- `-> T ! E` 拆出错误通道（见 §9）。

## 5. 类型

| 类别 | 拼写 | 备注 |
|---|---|---|
| 有符号整数 | `I8 I16 I32 I64` | 溢出**默认 trap** |
| 无符号整数 | `U8 U16 U32 U64` | 同上 |
| 浮点 | `F32 F64` | 无隐式提升，`F32`/`F64` 混用要转 |
| 布尔 | `Bool` | `true` / `false` |
| 码点 | `C32` | 单引号字面量 `'x'` 恰好一个码点 |
| 文本 | `Text` | 不可变 UTF-8 视图；与 `[U8]` 可等宽重标 |
| 尺寸 | `USize` / `ISize` | 与平台指针同宽 |
| 视图（切片） | `[T]` | 只有三条来路：分配、借用、`view` 重标 |
| 内联数组 | `[T; N]` | 长度是类型的一部分，可直接做字段 |
| 可空 | `T?` | 仅句柄类（视图/Text/enum/函数指针/协程句柄…） |
| 错误通道 | `T ! E` | `E` 一般是 enum |
| 通道 / 协程 | `chan T` / `coro T` | 并发：见 §9 |
| 不安全 | `Unsafe` | `[Unsafe]` 才可越过安全区 |

- `#size(T)` / `#align(T)` 查布局；视图、opaque、函数指针没有静态布局（E3602）。
- 算术：`+ - * / %`（`%` 仅整数），比较 `< <= > >= == !=`，
  逻辑 `&& || !`，位运算 `& | ^ ~ << >>`，复合赋值 `+= -= *= /= %= &= |= ^= <<= >>=`。
- **比较链 `a < b < c` 非法**（E2004），写 `a < b && b < c`。
- 视图 / `Text` / struct / 带载荷 enum 没有 `==`（E3205），内容比较自己写函数。

## 6. 转换：`as` 与 `trunc`

- `as`：**只做无损**——同宽互释（`I32 as U32`）或拓宽（`U8 as I64`）。
- `trunc`：**有损**方向（变窄、整数↔浮点）。名字就提醒你「这里会丢信息」。
- 没有隐式转换：`let x: I64 = i32_value;` 直接报 E3201。
- 无符号目标不能承接负字面量（E3204）：写 `0xFF` 而不是 `-1 as U8`。

```aila
let a: I32 = 300;
let b: U8 = a trunc U8;        // 有损
let c: I64 = a as I64;         // 无损拓宽
let d: F64 = a trunc F64;      // 整数→浮点属有损
```

## 7. 内存：区域、分配、视图、move

AILA 没有 `malloc` / `free` / 析构。内存只有三条来路：

```aila
region work {                        // ① 体内区域块：出块整块回收
    var buf: [U8] = [0; 4]@work;     //    分配点写裸区域名
    buf[0] = 7;
    let n: USize = len buf;
    _ = n;
}

func fill(n: USize, v: U8) @region[r] -> [U8] {   // ② 区域参数：寿命由调用方决定
    return [v; n]@r;
}

var table: [U8] = [0; 16]@static;    // ③ 静态区：寿命无限（hot 模块禁止）
```

```aila
func main() {
    region work {
        var buf: [U8] = fill(8, 1)@[work];   // 区域实参通道：分配落进 work
        _ = buf;
    }
}
```

- `copy x @[work]` 深拷贝到指定区域；`move x` 把主名字搬走（之后原名字不能再用，E4301）。
- `view b: [U8] = t;` 等宽零拷贝重标；改宽度必须进 `unsafe { }`。
- **逃逸**：返回体内 `region` 块里分配的视图 → E4101（块退出即回收）。改用区域参数或 `@static`。
- `for x in view` 的 `x` 是**元素拷贝**，改 `x` 不影响源；要写就用 `@mut[...]` 参数或下标。

## 8. 控制流与 match

```aila
if cond { } else if cond2 { } else { }      // if 恒为语句，不产出值（E2008）

for var i: I32 = 0; i < n; i = i + 1 { }    // 三段式（首槽只能是一条 var）
for !done { }                                // 条件循环
for x in xs { }                              // 元素
for i, x in xs { }                           // 下标 + 元素

match v {                                    // 取值位置：所有分支同型
    1 => "one",
    _ => "other",
};

match c {                                    // 表达式分支
    Color.Red => io.print_text("red"),
    _ => io.print_text("?"),
}

match e {                                    // 块分支必须在每条非发散路径 out 一个值
    ParseErr.Truncated { got, need } => {
        out got;                             // 或 return / trap
    }
    _ => { }
}
```

- 封闭值域（enum / `T?` / `T ! E`）必须穷尽（E3401）；整数族必须 `_` 且放最后（E3403）。
- 判别表达式若以「类型名 + 构造字面量」收尾要做括号：`match (Shape.Circle { r: 1.0 }) { }`（E2003）。
- `break` / `continue` / `return` / `defer { }` / `trap`（不可恢复终止）。

## 9. 错误与可空

```aila
enum ParseErr { BadMagic, Truncated { got: USize, need: USize } };

func parse(buf: [U8]) -> I32 ! ParseErr {
    if len buf < 8 {
        return _ ! ParseErr.Truncated { got: len buf, need: 8 };  // 造错
    }
    return 0x2A;
}

func caller(buf: [U8]) -> I32 ! ParseErr {
    let v: I32 = parse(buf)?;    // `?`：错误原样上抛（语言不引入栈展开）
    return v;
}

match parse(buf) {               // 接住
    Ok  { v } => io.print_i32_nl(v),
    Err { e } => { _ = e; }
}
```

- **可选 `T?`**：类型存在，但当前批次没有 `Some`/`None` 构造器可用（`E3001`）。
  要表达「可能没有」就用 `T ! Missing` 这种枚举错误通道，或用 `0`/空视图当哨兵。
- `_ = f();` 是「显式丢错」的**唯一**豁免口；`let _ = f();` 非法（E2006）。

## 10. 并发（当前批次：仅语法，代码生成未落地）

- 能过 `aila check` 的只有「起线程跑一次调用」：`thread f(args);`（要求本文件是轻量形态，
  或该 `thread` 语句写在 `module { }` 块内）。
- `coroutine` / `chan` / `<-` / `yield` 的完整形态还没定型，且**代码生成未落地**（E9002）。
  真需要并发时：先用 C 边界（`cpacks/`）或等后续批次，不要硬写。
- 线程里禁止写模块级 `var`（E5003），入口参数白名单见 E5002。

## 11. C 互操作

1. 把 C 件放进项目的 `cpacks/<名字>/`：`include/*.h` + 实现 `.c`（同目录即关联）。
2. 导入：`import ch _.cpacks.<名字>;`（`import cc` 只编译不产生声明）。
3. 直接用 C 原名：`let r: I32 = mathkit.mathkit_add(1, 2);`
4. 需要 `.ail` 产物时：`aila bindgen`（默认写 `build/bindgen/<cpack>.ail`）。

注意：
- 返回值是 C 错误码时，AILA 侧自己判码（语言不抛异常）。
- 变参只能收标量/无载荷 enum/视图/函数指针（E6001）；`Text` 要转 `[U8]`。
- union 永不直译（E6002 / W3002）：用 `[U8; N]` + `unsafe` 或写 C shim。
- 没有 clang 时 bindgen 走 builtin 兜底档，覆盖窄（W3001 会逐条列出没映射的声明）。

## 12. 标准库现状（别假设更多）

| 模块 | 提供 |
|---|---|
| `std.io` | `print_text` `print_bytes` `print_i32` `print_u32` `print_usize` `print_f32` `print_f64` `print_bool` `print_char` `print_nl` `print_i32_nl` `print_u32_nl` |
| `std.math` | `wrapping_add_u32` `wrapping_add_i32` `abs_i32` `min_i32` `max_i32` `fmod` |

没有 `std.str` / `std.mem` / `std.vec`。需要就把实现写进用户项目里（它们同样是 AILA 源码）。

## 13. 十三个「别的语言不是这样」的点

1. `if` 不产出值 → 用 `match` 或「占位 + if 赋值」。
2. 没有 `while` → `for 条件 { }`。
3. `len`/`copy`/`move`/`resume` 是**前缀**关键字，没有方法形态（E2005）。
4. 裸块 `{ }` 非法 → `region 名字 { }`（E2002）。
5. `::` 不存在 → 点号（E1007）。
6. 没有隐式数值转换（E3201）。
7. 算术默认 trap：不写 `wrapping_*` 就会在溢出时终止。
8. 浮点取余没有 `%` → `std.math.fmod`（E3209）。
9. 比较链非法（E2004）。
10. 有返回值函数不能当语句裸调用（E3206）。
11. 单引号只能是**一个码点**，多字符是字符串（E1003）。
12. 数字不能带类型后缀（`1u8` 非法，E1006）；八进制写 `0o777`。
13. `let` 是深不可变；`var` 才可写；跨函数改写要 `@mut[参数]`（E4201）。
