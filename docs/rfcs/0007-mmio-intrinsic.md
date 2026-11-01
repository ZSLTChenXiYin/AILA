# 0007 · `std.mmio` 内建（RFC · 未实施）

> 状态：**RFC，仅记录**（`docs/plans/0002 §C2` 的产出）。实施前不要引用本文件里的 API。

## 0. 为什么要这个

AILA 现在有两类"寄存器访问"：

1. **运行时内部**（`runtime/<家族>/aila_rt_<家族>.c`）：直接 `*(volatile u32*)ADDR = v`，寄存器地址来自
   `chip.json`（单一事实源），AILA 侧只看到 `std.gpio.write(pin, v)` 这类窄接口。
2. **cpack（C）**：同样是手抄寄存器，见 `examples/*/cpacks/sw_i2c/sw_i2c.c`。

两条路都要求"最后 10 米"落在 C。对 I2C 这类**微秒位时序**这是对的（designs/0006 §4.1 第 13 条）；
但对"我偶尔要读一个寄存器"的场景（用户自己写传感器驱动、写 SPI、写 DMA 描述符），
被迫引入一个 cpack + 一个 C 文件，代价明显高于收益——尤其是这几件事：

- 用户在 AILA 里画好了协议，却要把最后一步搬到 C；
- C 那一半无法被 `--lto`、无法被 AILA 的类型检查覆盖；
- 每个外设一个 cpack = 每个外设一份构建参与。

## 1. 缺的不是"能不能写内存"，是**地址到指针**的语言面

designs/0001 §11.9 定的是「RawPtr 类型不做，**地址即 USize**」。于是 AILA 侧：

```aila
let addr: USize = 0x60000304;     // 合法：就是个整数
// 但没有任何方式把它当内存读写 —— 没有 & / *，也没有 as 到指针
```

这不是疏漏，而是安全模型的一部分（B6：行为编译期可推导）。所以要做 `std.mmio`，必须**显式**开一个
最小的、命名的、可审计的逃生口，而不是恢复指针运算。

## 2. 候选方案

### 方案 A（建议）：层内建 `#compile.mmio.read32/write32`

```aila
#compile.mmio.write32(addr: USize, v: U32);
#compile.mmio.read32(addr: USize) -> U32;
```

- 编成**一条** `store volatile i32` / `load volatile i32`（不是 C 调用）——这正是 I2C 位时序需要的那条路；
- 只接受**字对齐**地址，运行期不平（`addr % 4 != 0` → 编译期能查就查，查不到就 trap）；
- 只给 32/16/8 三个宽度后缀式入口，不给指针算术、不给 memcpy 形态；
- 寄存器地址建议仍来自 `chip.json`（渲染成 `std.mmio` 可引用的常量），避免用户硬编码魔数；
- 审计面：`aila build --emit-json` 里列出模块用了多少个 mmio 点（沿用剪枝报告的形态）。

### 方案 B：`[Unsafe]` + 偏移

`std.mmio.write32(base: [Unsafe], off: USize, v: U32)`，与 `#runtime.atomic_*` 的形态一致
（designs/0001 §10.10 就是这么定的：原子操作挂在 `[Unsafe]` + 偏移上）。

- 优点：与 atomics 同构，不新增"整数当指针"的形态；
- 缺点：**`[Unsafe]` 从哪来**仍然无解——需要"把 USize 变成 `[Unsafe]`"的那一步，也就是方案 A 的核心。

### 方案 C：把 `[Unsafe]?` 当作 C 边界返回值

用 cpack 的一个函数 `mmio_base(addr: USize) -> [Unsafe]?` 造视图，再配 `#runtime.atomic_*` 风格的
库函数。可行，但把"寄存器访问"重新推回 C，等于没解决问题。

**结论**：做方案 A 的窄内建（三宽度、字对齐、volatile），方案 B 作为**上层形态**建立在 A 之上。
两者都不允许指针算术——这是与 B6 的边界。

## 3. 与 I2C 的关系（为什么这不是"把驱动搬回 AILA"的充分条件）

即使有了 `std.mmio`，AILA 版软件 I2C 还差两样：

1. **µs 级延时**：`std.time.sleep_us` 已经落地（三家族 ccount），够用；
2. **关中断**：位时序被 ISR 拉长 → 从机判总线错误。设计上要么给 `#runtime.irq_off/on`（新能力面），
   要么规定"AILA 版位时序只在裸机/单任务构建里用"。目前**没有**这个原语，所以 `sw_i2c` 的位时序
   继续留 C。

所以 `std.mmio` 的价值排序是：用户态寄存器访问 > SPI/DMA 描述符 > （将来）全 AILA I2C。

## 4. 开放问题

1. `volatile` 语义是否够？DMA 一致性还要求 cache 操作（designs/0001 §13.6 的 `std.cache` 族）。
2. 是否允许"地址常量"参与 `#compile.if` 判定（例如按 `chip.json` 的寄存器表剪枝）？倾向允许
   ——它本来就是 const-expr。
3. 越界/未映射地址的处罚：trap 还是"就是读到垃圾"？倾向 trap（B6），但这要求能拿到内存映射表。
4. `std.mmio` 是否需要 `#[unsafe]` 式的显式块，而不是普通函数调用？倾向需要——让"越过安全线"在源码里
   一眼可见（`unsafe { std.mmio.write32(...) }`）。
