# RFC 0005 · 标准库：能用 AILA 写的就用 AILA 写

> **编号**：0005 · **类型**：RFC · **状态**：已采纳并落地
> **上游**：[designs/0001](../designs/0001-language-spec.md) §12.3（运行时薄层四件套）、§13.1（标准库边界）
> **实现**：`stdlib/io.aila`（AILA 算法 + 两条 `#[c]` 边界）、`runtime/aila_rt.c`（`aila_rt_write` / `aila_rt_flush`）

## 1. 动机

M1 最初的 `stdlib/io.aila` 是「C 打印函数的薄包装」：每个 `print_*` 里只有一句
`aila_print_*` 的 `#[c]` 调用，格式化全部落在 C 的 `printf`。这带来三个问题：

| 问题 | 说明 |
|---|---|
| **维护成本** | 标准库的语义被劈成两半：一半 AILA 一半 C。改一个格式化行为要同时动两处，且 C 侧没有 AILA 的类型/区域检查 |
| **跨目标不可控** | `printf("%g")` 之类的行为在 xtensa / riscv32 / newlib / msvcrt 上细节不同；标准库应当自己定义输出，而不是继承 libc 的口味 |
| **边界模糊** | 「哪些东西能进标准库」没有硬规则，下一次很容易又把某个 C 库包一层就说是 std |

## 2. 裁决

1. **能用 AILA 实现的，就用 AILA 实现**：数字格式化、符号处理、UTF-8 编码、浮点小数位、换行与组合函数
   全部写在 `stdlib/*.aila` 里，不转手给 libc 的 `printf`。
2. **只有跨平台的 OS 边界交给 C**：写标准输出在各平台的差异（`fwrite` / `write(2)` / ESP-IDF）由
   `runtime/aila_rt.c` 消化，标准库只声明两条最小绑定：

```aila
#[c] func aila_rt_write(bytes: [U8], n: USize);
#[c] func aila_rt_flush();
```

3. **不新增语言机制**。曾试用过 `#[intrinsic]`（「语言原语」）来把 C 名字从标准库里抹掉，
   按设计决定**已删除**：有 `#[c]` 就够了，多一个声明形态只会让语言表面变复杂。
   如果将来某个平台需要别的系统调用，就再加一条 `#[c]` 绑定 + 运行时实现，而不是加新语法。
## 3. AILA 侧的实现要点

| 关注点 | 做法 |
|---|---|
| 缓冲区 | `region` 块里就地分配 `[U8; N]`：打印一次、回收一次，**不留模块级可变状态** |
| 十进制 | 先算位数，再从末尾往前填（AILA 没有指针偏移，不能倒序写） |
| 负数 | 在 I64 里取反，避开 `I32` 的 `MIN` 取反溢出陷阱 |
| 浮点 | 整数部分 + 最多 6 位小数，逐位乘 10 取整后去尾零（`0.75` → `0.75`）；嵌入式够用且完全可移植 |
| UTF-8 | 按码点分四档编码（1/2/3/4 字节），不依赖 libc |
| 换行 | 写 `\n` 后调 `flush()`（交互式输出要求） |

## 4. 运行时的边界

`runtime/aila_rt.c` 现在只剩三类东西：

| 类别 | 内容 |
|---|---|
| **输出原语** | `aila_rt_write(buf, n)` / `aila_rt_flush()` |
| 区域账本 | `aila_rt_region_enter/alloc/leave`、`aila_rt_static_region`（designs §5 的分配器） |
| 陷阱出口 | `aila_rt_trap(msg, len, file, line)` |

运行时**不再**提供任何 `printf`-系语义；换平台只需实现这两个原语与区域账本。

## 5. 验收

| 判据 | 位置 |
|---|---|
| stdlib 里没有 C | `grep -c '#\[c\]' stdlib/*.aila` 为 0（CI 可加） |
| 六个示例输出不变 | `tests/corpus/examples/*/expected.out` 逐字节比对（`TestExamplesEndToEnd`） |
| IR 快照更新 | `tests/corpus/examples/*/std_io.ll`（`std_io` 现在是纯 AILA 实现） |
| 边界可替换 | 换平台只动 `runtime/aila_rt.c` 的两个函数；标准库源码不动 |

## 6. 刻意不做

| 不做 | 理由 |
|---|---|
| 为了省事把整块格式化交给 C | 边界只允许是「平台相关的系统调用」，不允许把数字/字符串处理搬进 C |
| 浮点的最短往返表示（Ryu/Grisu） | M1 用 6 位小数即可；对嵌入式更可预测，且不需要大整数 |
| 把 `printf` 语义照搬 | 标准库定义自己的输出格式，不继承 libc 的平台差异 |
