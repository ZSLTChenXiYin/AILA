# 0006 · 显示栈与编译期剪枝（完整方案）

> 状态：**P0–P5 全部落地**（2026-11-01），后续收尾项见 `docs/plans/0002`（语言缺陷修复、驱动层拆分
> `sw_i2c` + `oled/panel`、`std.time.sleep_us`、`--lto` 与 hot 边界、接口摘要，均已完成）。
> P0 语言机制、P1–P2 标准库、P3 驱动、P4 字库与示例、P5 构建体积都已在宿主 golden 与 ESP8266 真机上验收；
> 落地过程中暴露的语言 / 后端缺陷见 §4.1。
> 前置：`designs/0001`（语言约束：无泛型/trait/接口/union；铁律 7 跨模块不可内联；§3.9 函数指针仅限模块内）、
> `designs/0003`（stdlib 分层与"宿主误用 = stub + trap"惯例）、`designs/0005`（ESP8266 上板实测）。
>
> 本文两部分：**§1 语言前置（编译期剪枝）** 与 **§2 显示栈分层**。§1 可独立先行落地，§2 依赖 §1 做设备剪裁。
>
> **语义定稿位置**：`#compile.if` / `#compile.match` 的语言语义已回写 [designs/0001 §7.8](0001-language-spec.md)；
> 大聚合体按只读指针传的实现约定写在 [designs/0001 §5.2](0001-language-spec.md)。本文保留方案与验收记录。

---

## 0. 结论摘要

1. **不引入任何新语言机制就能得到"条件编译"**：`const`、`match` 的 const 模式、`#size`/`#align`、字面量已经是编译期值；
   只需补一条规则（函数内 `let` 绑 const 表达式也算编译期已知）+ 一个语句级指令 `#compile.if` / `#compile.match`。
2. **不加 `@const` 形参**（`@` 族是内存与效果语义域，不是编译期值域；且常量形参 = monomorphization-lite，
   与"模块 = 热替换最小单位"冲突）。跨模块的设备选择改为**应用侧 `#compile.match` + 链接器 GC**。
3. **不做文本级 `#if`**：死分支仍完整解析并类型检查；可移植性由 stdlib 既有的 stub+trap 惯例承担。
4. 显示栈按 Linux DRM/KMS 的**对象/状态/helper 三分 + damage 提交 + 数据表驱动**重做，
   但**不搬**运行期注册驱动、ioctl 多态、跨模块 ops 表——这三样在 AILA 里会直接撞 B6。
5. 点阵缓冲区成为跨平台模块：**行主序 + stride + 1bpp**、**cell 网格（8×16）**、**全/半角由字库数据决定**、
   **滚动归文本控制台（客户端策略）**、**唯一推送通道 = 带 damage 的 commit**。

---

## 1. 语言前置：#compile.if / #compile.match

### 1.1 语义

语句级指令，条件必须是**编译期常量表达式**（下称 const-expr）。判定为真/命中的分支照常生成代码，
其余分支**只在代码生成阶段被省略**，此前所有阶段（解析、解析名、类型检查、区域/能力分析、告警）都完整进行。

```aila
const MODE: U8 = 2;

func demo(n: U32) {
    let fast: Bool = (MODE == 2);           // const-expr → 编译期已知
    #compile.if fast { use_fast() } else { use_generic() }

    #compile.match MODE {
        1 => low_power(),
        2 => full_speed(),                  // 只有这一支生成代码
        _ => {},
    }
}
```

### 1.2 const-expr 的封闭定义

允许：字面量、`const` 名（本模块壳/块内、或 `import` 进来的模块 const）、`#size(T)`、`#align(T)`、
`#comprun.len(字面量数组)`、函数内**一次性** `let` 绑定到 const-expr（`let` 不可重新赋值，天然 SSA），
以及仅由上述构成的纯运算符表达式（算术/比较/位/布尔/标量 `as`·`trunc`）。

不允许（求不出 → 报错，不降级）：普通形参、模块级 `var`、任何含函数调用的表达式、`chan`/句柄读值。

### 1.3 规则表

| # | 规则 | 理由 |
|---|---|---|
| 1 | 条件非 const-expr → **编译错误** `E2xxx`，诊断给出"为什么不是常量"的来源链（形参/`var`/含调用） | 静默降级成运行时 `if` 是最危险的失真（以为 0 字节，实际留分支+依赖） |
| 2 | 死分支**完整解析 + 类型检查**；不允许靠隐藏代码来表达"该目标没有此 API" | 安全性：死代码永不腐烂；可移植性用 stdlib 的 stub+trap（`designs/0003 §1` 已有先例） |
| 3 | 剪枝**不改变模块 ABI / 接口指纹**；区域、`@mut` 能力、逃逸、未使用告警全部跑**未剪枝 AST** | 否则同一模块在不同剪枝结果下 ABI 不同，热替换要按配置重新验证（撞 B6） |
| 4 | `#compile.match` **不要求穷尽**（判别值已定）；被剪臂仍类型检查；模式仍须是编译期已知值（沿用 `0001 §7`） | 值已定，强制穷尽只会逼人写 `_ =>` |
| 5 | 剪枝结果只依赖源码常量，**不依赖环境**（无 manifest cfg、无 `-D`）→ 构建可复现、指纹不变 | 用户要求：机制属于开发者，不属于构建配置 |
| 6 | 跨模块引用 const 时，该 const 纳入**模块接口指纹**；指纹变化触发依赖方重编，并在报告里列出 | 否则"改了 const 却没重编"会产出错误固件 |
| 7 | 必须报告剪枝结果：`aila build --emit-report`（每处剪枝点：位置 / 判定值 / 被剪臂 / 判定来源链） | 可观测性 |
| 8 | 可选 `--emit-size-delta`：对比不同 const 取值的 code size 差 | 让"用它省体积"可量化 |

### 1.4 正当使用场景（写进文档，防滥用）

同模块内的 `if` + 常量本来就可能被 LLVM 折叠（同模块可内联）。`#compile.if` 的**正当场景只有三种**：

1. **跨模块常量**：铁律 7 禁止跨模块内联 → LLVM 折不了，这是唯一的语言级手段；
2. **`-O0` / debug 构建也要剪**（优化器不参与时）；
3. **剪掉后连依赖都不需要**（cpack 符号、字库表、整段驱动代码）。

### 1.5 明确不做

| 不做 | 理由 |
|---|---|
| 文本级 `#if` / 宏 | 先删文本再解析 → 死代码里的语法/类型错误永远发现不了，宏状态泄漏；与本方案"安全性"目标相反 |
| `@const` 形参 | `@` 族 = 内存与效果语义域（`@mut`/`@move`/`@static`/region），不是编译期值域；且按值特化与"模块 = 热替换最小单位"冲突 |
| manifest `[cfg]` 段 | 机制属于源码，不属于构建配置（用户明确要求）；设备选择用应用侧 `#compile.match` |
| 运行期设备注册表 | 需要跨模块持函数指针 → 撞 `0001 §3.9` 与 B6 |

> 若未来出现"库内剪枝、常量由使用方决定"的硬需求，应作为**独立 RFC**：命名进 `#compile.*` 族
> （如 `n: U32 #compile.const`），并写明三条代价（按值复制、跨模块不特化、升级需重验所有特化副本）+ 份数上限。

### 1.6 落地位置

实现顺序：`internal/parser`（语句级指令）→ `internal/sema`（const-expr 求值 + 判定 + 诊断）→
`internal/codegen`（跳过死分支）→ 报告。语义定义最终回写 `designs/0001`（新章节，紧邻 §6.9 优先级与 §1.4 词汇归类）。

### 1.7 落地状态（P0 已完成）

| 项 | 位置 | 说明 |
|---|---|---|
| 语法 | `internal/parser/decls.go` | `parseCompileStmt` / `atCompileStmt`；`parseIfTail`/`parseMatchTail` 复用普通 `if`/`match` 解析 |
| AST 标记 | `internal/ast/{stmts,exprs}.go` | `If.CompileTime/Prune`、`Match.CompileTime/PruneArm`（`Prune = -1` 表示未剪） |
| 常量求值 | `internal/sema/consteval.go` | `constExprInt`：字面量 / `const`（含导入）/ 函数内 `let` of const-expr / 一元二元纯运算 |
| 判定与诊断 | `check_stmt.go` / `check_match.go` | `E3210` 条件非常量、`E3211` 判别式非常量、`E3212` 无分支命中、`E3213` 模式非常量 |
| 剪枝生成 | `internal/codegen/func.go` / `stmt.go` | `lowerIf` 只生成命中分支；`lowerMatch` 抽出 `emitArmBody` 后只生成命中臂 |
| 可观测性 | `internal/sema/prunes.go` + `writeReport` | `aila-build.json` 新增 `compile_time_prunes`（模块 / if\|match / 行号 / 判定值 / 命中分支） |
| 错误码文档 | `share/aila/data/diag_codes.json` | E3210–E3213 已登记（title/explain/fix/doc） |
| 测试 | `internal/codegen/compile_time_test.go` | 6 条：剪枝生效、`let` 常量传播、形参条件报错、match 剪枝、判别值无命中报错、死分支仍类型检查 |
| 目标验证 | `examples/oled_smoke` | `const VERBOSE: Bool` + `#compile.if` 包住每帧 UART 日志；`false` 时板上只剩开机行（固件照跑），`true` 时恢复逐帧日志 |

> 已知限制：`#compile.match` 目前只接受字面量 / const / 枚举变体判别值模式（解构/绑定/类型模式报 `E3213`）；
> 常量形参（跨模块特化）不在 P0 范围（见 §1.5）。


---

## 2. 显示栈分层

### 2.1 Linux 抽象 → AILA 映射

| Linux | 它解决的问题 | AILA 落法 | 差异/理由 |
|---|---|---|---|
| `drm_gem_object` + `drm_framebuffer` | 内存与"如何解释像素"分离 | `std.pixbuf.Pixbuf { data, w, h, stride, fmt }` | 引入 **stride**（我们现在硬编码 128 列） |
| `drm_format_info`（数据表） | 格式差异查表，不进代码分支 | `std.pixfmt.PixelFormat` **const 表** | 完全照搬（数据驱动） |
| damage clips + atomic commit | 唯一推送通道；局部刷新是一等公民 | `std.damage.Damage` + `Commit { buf, damage }` | 定长 clip 数组（无分配），cell 位图替代任意多边形 |
| plane/crtc/encoder/connector | 合成/时序/链路/端点分离 | M1 只保留"面板 = 时序 + 电源序列"一层 | 单面板裸机不需要 plane 合成 |
| `drm_panel` + deferred IO | 器件物理特性独立成层 | `cpacks/<panel>` 的 `PanelOps` | e-paper 的延迟刷新留给 M2 |
| `drm_format_helper`（被驱动调用，IoC） | 通用算法与驱动骨架分离 | `std.paint` 纯函数，**驱动调用它** | 关键：依赖方向下→上，避免跨模块回调 |
| fbcon（`console_struct_ops`） | 文本控制台是 fb 的**客户端** | `std.textgrid`（滚动/换行/软光标都在这里） | 完全照搬 |
| `*_funcs` 运行期 vtable + `register_framebuffer()` | 运行期可插拔驱动 | **不做**：应用侧 `#compile.match` + 链接器 GC | 跨模块持函数指针违 B6 |

### 2.2 分层与模块

| 层 | 模块 | 内容（public API 草案） | 归属 |
|---|---|---|---|
| L0 | `std.pixfmt` | `PixelFormat { bpp, bytes_pp, row_align, bit_order }`；`mono1()` | 批次 1（零 runtime） |
| L1 | `std.pixbuf` | `Pixbuf { data:[U8], w, h, stride, fmt }`；`new/clear/fill/set/get/row_bytes`；`cell_w/cell_h/cols/rows` 元信息 | 批次 1 |
| L2 | `std.damage` | `Rect { x0,y0,x1,y1 }`；`Damage { clips:[Rect;8], n, cells:[U8;16] }`；`clear/add_cell/add_rect/union/is_empty/cells_count` | 批次 1 |
| L3 | `std.paint` | `fill_rect / hline / vline / blit_glyph / text_cells / blit_line`（全部纯函数，参数 = `Pixbuf` + `GlyphSet` + 坐标） | 批次 1 |
| L4 | `std.textgrid` | `TextGrid { lines, rows, cols, cursor, scroll }`；`put_line/scroll_by/clear/render(grid, buf) -> Damage` | 批次 1 |
| L5 | `std.display` | `Commit { buf, damage, enable }`；`present(commit)` 的**驱动侧**实现契约；`sink_host_pbm(commit)`（调试/测试 sink） | 批次 1（host sink）+ 3（设备） |
| L6 | `cpacks/ssd1306` 等 | `const PANEL_OPS: PanelOps = ...`（模块内）；`init/flush(commit)/sleep`；只做传输与控制器时序 | 批次 3 |

**Inversion of Control**：`std.display`/`std.paint` 不知道设备存在；设备模块 `import` 它们并调用。
`PanelOps` 这类 ops 表只在**驱动模块内部**做 const（`designs/0001 §3.9` 批准的形态），不进任何公开签名。

### 2.3 canonical 像素格式

**行主序、bit7 = 最左像素、每行按 `row_align`（4 字节）对齐、1bpp**。

- `stride = ceil(w/8)` 再向上对齐到 4 字节；128 宽 → 16B/行，总 1024B。
- SSD1306 的"页格式"（一字节 = 8 竖像素）**属驱动私事**：`cpacks/ssd1306.flush()` 在读 canvas 时按位搬运
  （不做整块转置缓冲，省 1KB RAM；实测整屏 I2C 37ms 远大于搬运算术开销）。
- 收益：宿主可直接导出 PBM、任意面板可接、diff = 普通字节比较。

### 2.4 cell 网格与全/半角

- `cell_w = 8`、`cell_h = 16` → 128×64 = **16×4 cells**；`Damage.cells` 位图 = 16 字节（16×4 bit，留余量）。
- **全/半角是字库数据，不是绘制逻辑**：`GlyphSet` 带 `widths: [U8]`（每字占 1 或 2 cell）+ `advance = widths[i]`；
  `std.paint.text_cells` 只按 advance 推进，不做任何"这是不是 CJK"的判断。
- 文本 API **以 cell 坐标为参数**（像素坐标只给图形图元）：帧对比、局部刷新、全半角排版共用同一个最小单位。

```aila
type GlyphSet = struct {
    cps:   [U32],    // 升序码点
    bits:  [U8],     // 每字 32B（16 行 × 2B，MSB 在左）
    widths:[U8],     // 1 = 半角(8px 宽)，2 = 全角(16px 宽)
    count: USize,
    px_h:  U16,      // 16
};
```

> 现状 `examples/oled_smoke/font/font16.aila` 是"284 字全 16×16"，需要 `tools/genfont.ps1` 扩展：
> 生成半角 8×16 的 ASCII 子集 + `widths` 表。迁移见 §6。

### 2.5 滚动归客户端

- `std.pixbuf` **不支持滚动**（同意用户方案）。
- `std.textgrid` 用**内容模型**实现滚动：持"行 → 码点"，`scroll_by(n)` 只改索引，
  随后 `render()` 把 16×4 cell 与上一帧比较，产出**精确到 cell 的 Damage**。
- 这与"画布内 memmove 上移"的区别：后者滚动后几乎整屏都脏，OLED 上等于整屏重发（37ms）；
  前者只发真正变化的页（实测单页 ~5ms）。

### 2.6 设备选择（不引入运行期分派）

```aila
// 应用模块内，5 行搞定；常量是唯一"配置"
const DISPLAY: U8 = 1;      // 1 = ssd1306_i2c，2 = st7735_spi（M2）

func main() {
    #compile.match DISPLAY {
        1 => ssd1306.run(),
        2 => st7735.run(),
        _ => {},
    }
}
```

- `#compile.match` 在**同模块内**剪枝 → 只有被选中的驱动分支生成代码；
- 未调用的驱动模块无人引用 → 由链接器 GC（构建已有 `-ffunction-sections`，补 `--gc-sections`）；
- 所以 `std.display` 里**不需要**"按 kind 运行期分派"的函数——`DisplayKind` enum 也随之下线。

### 2.7 内存与性能预算（ESP8266 · 128×64）

| 项 | 大小 | 说明 |
|---|---|---|
| canvas（行主序） | 1024 B | region 一次分配 |
| damage（clips+cells） | ~40 B | 定长，无分配 |
| textgrid 行模型（4×16 码点） | 256 B | 码点数组 |
| 读取缓冲（可选，驱动侧） | 0 B | 不转置，直接按位搬 |
| 字库（284 字） | 9088 B bits + 1.1 KB cps + 284 B widths | 在 `.data`/flash，不占运行期堆 |
| **运行期 RAM 合计** | **≈1.3 KB** | 相比现状（1024 B fb）增加可控 |

| 操作 | 预算 |
|---|---|
| 整屏 I2C 写 | 37 ms（实测，125 kHz 软件 I2C） |
| 单页写（damage 局部） | ~5 ms |
| cell diff（16×4） | 数十 µs |
| 位搬运（1024 B） | 数十 µs |

---

## 3. 可测性（本方案最大增值点）

- L0–L4 全是**纯 AILA、零 runtime** → 可在宿主 `aila run` 直接跑，无硬件。
- `sink_host_pbm(commit)` 把 canvas 输出为 PBM(P4) 文本 → golden 测试。
- 断言清单（建议进 `tests/`）：
  1. 半角 ASCII 占 1 cell、全角占 2 cell（边界对齐）；
  2. 换行与行内截断；
  3. `scroll_by(1)` 后 `Damage.cells` 只包含变化行 + 新行（**不是整屏**）；
  4. `#compile.match` 剪枝后，未选中驱动目标符号不出现在产物里（`nm` 断言）；
  5. `#compile.if` 条件非 const-expr 时报 `E2xxx`（负向用例）；
  6. 剪枝前后**模块接口指纹一致**（ABI 不变）；
  7. 死分支中的类型错误仍被报出（负向用例）。

---

## 4. 迁移路径（每步独立验收）

| 步 | 内容 | 验收 |
|---|---|---|
| **P0** | `#compile.if` / `#compile.match`（语句级、const-expr、报错口径、`--emit-report`） | 6 条测试中的 5、6、7 通过；`designs/0001` 补语义章节 |
| **P1** | `std.pixfmt` + `std.pixbuf` + `std.paint` + `sink_host_pbm` | 宿主 golden：图元 + 全/半角排版（用现有 `font16` 先全角跑通） |
| **P2** | `std.damage`（cell 位图 + clip 合并）+ `std.textgrid`（滚动 = 内容模型） | 断言 1–3 通过 |
| **P3** | `cpacks/ssd1306` 改造：`init/flush(commit)`，按 damage→页范围逐页写；行主序按位搬 | 板上：4 行滚动终端，串口报 `nack=0` 与每帧 dirty 页数 |
| **P4** | `examples/oled_smoke` → `examples/display_demo` 瘦身（只留 `#compile.match` 选屏 + 喂数据）；`tools/genfont.ps1` 出半角 8×16 + `widths` | 板上混合排版（中英混排） |
| **P5** | 构建加 `--gc-sections`；`--emit-size-delta` 度量"换驱动/关功能"的体积差 | `nm` 断言 + 报告 |

### 4.1 已知坑（实测，API 设计必须规避）

1. **函数内 `view v: [U8] = 模块级数组;` 会被编坏**（`LoadProhibited`, `excvaddr=0x3ff`）。
   → 标准库 API 一律"`[U8]` 视图只作参数"，scratch 由调用方在 region 内开好传入。
2. **cpacks 的 `[U8]` ABI**：跨 C 边界只传一个指针（实测 `oled_abi(fb,7) == 7`）。
   → 适配器契约二选一并写死：bind 式（绑一次指针、后续只传标量）或严格照抄 `(ptr, len, ...)` 形参形状。
   本轮的 `panel.bind(fb, stride)` + 之后全标量，即 bind 式。
3. **SSD1306 忽略 0x21/0x22 水平寻址**（`designs/0005 §6`）→ 驱动必须走页寻址；
   `Damage` 到页范围的换算因此是驱动核心逻辑（`panel.page_dirty` / `panel.page_cols`）。
4. 驱动必须在 `oled_init` 里做 I2C 总线恢复（9 个 SCL + STOP），并暴露 NACK 计数供上层观测。
5. **`@mut[self]` 方法经另一个函数的 `@mut[f]` 参数转发时写不回调用方**。大聚合体改按指针传后
   （第 8 条）这条自动成立；但 **≤16 字节的小结构体仍是按值复制**，其标量字段写不回，
   所以小结构体的可变接口仍要用 `@mut[self]` 方法、且方法内避免「通过 self 再调本模块方法」。
   `std.damage` / `std.textgrid` 最终用句柄形态彻底绕开这条。
6. **模块级 `let X: Text = "…"` 是空串**（`view` 出来 `len=0`，打印也空）→ 文本常量一律放函数内
   （`let s: Text = "…"; view b: [U8] = s;`）；字库这类数据模块用 `[U8; N]` 数组 + 视图访问器。
7. **`#compile.match` 的判别值是常量时，旧 codegen 生成 `load i8, ptr null`**（常量不是主名字，
   `lowerPlace` 给不出地址）→ 裸机上就是 `Fatal exception (28) LoadProhibited`、`epc1` 落在入口的
   `l8ui` 上。已在 codegen 修（退回按值求值）。
8. **xtensa call0（16 寄存器）上按值传 / 构造大聚合体必崩**：
   `LLVM ERROR: Cannot scavenge register without an emergency spill slot`，`-O0..-O3` 与
   `-regalloc=basic|fast` 全都一样（esp-llvm 后端硬限制）。只能两头解决：
   - **参数**：大聚合体（>16 字节）改按**只读指针**传（已落地，见 designs/0001 §5.2）；
   - **构造 / 返回 / 内联字段**：把承载大状态的结构做成**句柄 + 区域内存**
     （`struct Damage { raw: [U32] }`，状态在调用方 region 里），句柄两个机器字。
   本轮 `std.damage` / `std.textgrid` 即句柄形态；`std.pixbuf.Pixbuf`（≈20B）与
   `std.display.Commit`（≈32B）保持值形态。
9. **调试出口不能进固件**：`sink_host_pbm/ascii` 要 `std.io`（格式化 + Writer），设备侧连它的循环
   都会把寄存器分配顶崩 → 拆出 `std.display_host`，设备只 import `std.display`（顺带固件变小）。
10. **`region_alloc` 的尺寸参数恒为 i64**，32 位目标上长度是 i32：旧 codegen 生成 `mul i64 %i32, N`
    → 非法 IR。已修（`toI64` 收口）。这类错在宿主上完全测不出来（宿主 `USize` = 64 位），
    **改 ABI 相关代码后必须真机验证**。
11. **`chip.json` 的 `ir_flags`（`-O1`）原本只对 SDK 模式生效**：裸机 esp8266 走 `-O0` 必崩 →
    现在 `ir_flags` 是**芯片级**事实，裸机与 SDK 都叠加（`internal/build.chipBackendFlags`）。
12. `const` 初始化器必须是纯字面量表达式（`const B: U8 = A * 2` 不合法），`std.damage` 的 `WORDS`
    因此写字面量 + 注释标明来源。

13. **模块级 `let X: Text = "…"` 是空串的根因**：全局初值只认字面量标量 / 定长数组字面量，
    字符串字面量掉到兜底的 `zeroinitializer` → Text 变 `(null,0)`。已修（`StrLit` → 常量池 +
    `{ ptr, len }`），`internal/build` 有回归用例。
14. **`@mut[self]` 转发写不回的根因**：≤16 字节的小聚合体 `@mut` 参数按值传，被调方拿到副本。
    已修：**`@mut` 的聚合体参数一律按指针传**（视图家族除外）。这条修完，`std.damage` /
    `std.textgrid` 的句柄化不再只是"绕开"，而是"两条路都通"。
15. **`[T]?` 现在造不出非空值**：v2 退役了 `Some` 构造、也没有隐式包装，所以 bindgen 生成的
    `[U8]?` 形参在 AILA 侧无法调用。→ cpack 一律"手写 `#compile.c` 声明 + 视图只传指针 + 显式长度"，
    头文件留待可空视图构造语法收敛（`docs/plans/0002 §B1`）。
16. **铁律 7 的边界只在 `module hot`**：非 hot 模块之间允许跨模块内联（`aila build --lto`：
    整程序文本 IR 合并 + `llc -O2`）；hot 模块的函数打 `noinline optnone`。
    合并时必须做四件事（少一件后端就报重复定义）：只留第一份 target 行、去掉重复 `declare`、
    `attributes #N` 按模块重编号、`@.str.N` 字符串池按模块重编号。
17. **页格式搬运应当留在 AILA**：`std.paint.transpose_page` 是纯位运算，宿主上能逐字节 golden，
    开销（1024 次位测试/页 ≈ 几十 µs）远小于该页 I2C 传输（≈4.6ms）。C 侧只留位时序——
    这也让两个工程（裸机 / SDK 版）共用同一份驱动协议。

### 4.2 落地状态（P0–P5 全绿）

| 步 | 落地物 | 验收证据 |
|---|---|---|
| **P0** | `#compile.if` / `#compile.match`（parser / sema / codegen / 报告）+ `E3210–E3213`；语义回写 designs/0001 §7.8 | `internal/codegen/compile_time_test.go` 6 条；`#compile.if VERBOSE` 真机 A/B |
| **P1** | `std.pixfmt` / `std.pixbuf` / `std.paint`（`GlyphSet` + UTF-8 解码） | `examples/display_demo` 的 `[1]` `[2]` 段 + ASCII 预览 golden |
| **P2** | `std.damage`（句柄 + clip 合并 + cell 位图）/ `std.textgrid`（内容模型 + cell diff） | `examples/display_demo` 的 `[3a]–[3g]`：不改 0 脏、单 cell 编辑 1 脏、滚动只脏真变行 |
| **P3** | `cpacks/ssd1306` 页寻址改造（行主序按位搬 + 列窗口 + NACK/字节计数）+ `examples/oled_smoke/panel` | 真机：`oled: i2c=1 fonts=292 cells=16x4`、每帧 `cells=11..13 pg=2`——**只刷 2 页 ≈5ms，而非整屏 8 页 ≈37ms** |
| **P4** | `tools/genfont.ps1` 半角 8×16 + `widths` + 视图访问器；`examples/display_demo`（宿主）/ `examples/oled_smoke`（设备） | 中英混排上屏（真机 `nz` 稳定增长、四行滚动） |
| **P5** | `aila build --gc-sections` / `--emit-size-delta`；`internal/build/gc_sections_test.go` | 未选中驱动的**符号不在产物里**（两条用例互相验证）+ 逐模块体积表 |

---

## 5. 风险与未决项

| 项 | 风险 | 处置 |
|---|---|---|
| `#compile.if` 滥用 | 开发者用它做本可由优化器完成的微优化 | §1.4 三条正当场景写进文档；`--emit-report` 可见 |
| 跨模块 const 指纹 | 漏重编会产生错误固件 | 纳入接口指纹 + 报告列出受影响模块 |
| 字库体积 | 半角 8×16 + 全角 16×16 双套字库 | 半角只收 ASCII 区；**定长 32 字节记录**（半角只用左半字节）换来绘制零查表，代价 ≈3KB flash（不占 RAM） |
| 1bpp 是否够 | 未来彩屏/灰度 | M1 **明确写死 1bpp**；`PixelFormat` 表为多格式留位，但只实现 mono1 |
| 局部刷新收益 | 页是最小寻址单位（8 行），cell 高 16 行 → 最小刷 2 页 | 实测即 2 页/帧；文档标明"最小刷新粒度 = 2 页" |
| `const Text` 是否存在 | 设备名想用字符串常量 | 结论：模块级 `Text` 常量当前不可用（§4.1 第 6 条），设备码仍用 `U8` |
| 大聚合体 ABI | 「按只读指针传」是**实现约定**，尚未进接口指纹口径 | 已落地最小可用形态：`aila-build.json` 的 `module_interfaces` 只对**签名行**取摘要（`--lto` 下内联也不影响摘要）；正式口径（ABI 版本号）留 M2 |

### 已拍板（P0 实施时确认）

1. `#compile.match` **允许**不穷尽（判别值已定）。
2. 跨模块 const 纳入接口指纹；剪枝**不改** ABI / 指纹（同一模块在任何剪枝结果下指纹相同）。
3. cell 尺寸 = **8×16**（半角 1 格 / 全角 2 格）。
4. M1 **写死 1bpp**。
5. P0 先于显示栈落地（显示栈的设备选择依赖它）。
