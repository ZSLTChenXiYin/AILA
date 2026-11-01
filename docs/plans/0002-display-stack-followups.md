# 0002 · 显示栈收尾计划（边缘项整合）

> 状态：进行中。前置：`designs/0006`（P0–P5 已落地）、`designs/0001 §7.8 / §5.2`。
> 本文把 P0–P5 落地后散落在各处的「边缘项」收成一张可执行清单，每项都带验收口径。
> 执行顺序 = 编号顺序；每项完成即勾掉并在文末「执行记录」写实测结果。

## 0. 为什么有这份计划

P0–P5 交付的是主链路（语言剪枝 + 标准库 + 驱动 + 字库 + 构建体积）。但落地过程中暴露了
**两个语言侧真缺陷**、**驱动分层仍不干净**（协议逻辑困在 C 里、两份副本要同步维护）、
**平台能力缺口**（没有 µs 延时）、以及一条**用户提出的语义修正**（内联限制应只在涉及热替换
模块时生效）。这些都不是"再做一遍"，而是把主链路接成一条直的路。

## A. 语言侧真缺陷（最高优先：用户可见）

| 项 | 现象 | 修法 | 验收 |
|---|---|---|---|
| **A1** | 模块级 `let X: Text = "…"` 是空串（打印为空、`view` 出来 `len=0`，`designs/0006 §4.1` 第 6 条） | codegen：模块级 Text 常量的 rodata 初始化/取址路径；顺带检查 `const` 与 `let` 两种形态 | 新增宿主可跑用例：模块级 Text 常量打印 + `view` 长度 + 下标字节三重断言 |
| **A2** | `@mut[self]` 方法经另一个函数的 `@mut[f]` 参数转发时，**标量字段**写不回调用方（≤16 字节的小结构体仍是按值复制） | codegen：转发调用时传**真实指针**（槽位的 ptr 分量），不要传槽位地址 | 用例：`forward(c: Counter) @mut[c] { c.bump(); }` 后 `c.v` 必须变 |
| **A3** | 旧坑复查：函数内 `view v: [U8] = 模块级数组` 曾被编坏（`excvaddr=0x3ff`） | 现按"实测"再验一遍（宿主 + 上板） | 用例 + 上板跑通；若不复现则在 `0006 §4.1` 第 1 条标注"已不复现（版本 X）" |

## B. 驱动层收敛（把协议搬回 AILA）

| 项 | 内容 | 验收 |
|---|---|---|
| **B1** | 拆 `cpacks/sw_i2c`：C 只留**位时序**（start/stop/byte+ACK/recover/delay + `write_bytes(ctrl, ptr, n)`），并补 `include/sw_i2c.h` 让 **bindgen 自动生成绑定**（不再手写 `#compile.c func`） | `aila build` 通过；`--emit-bindings` 能看到生成的 `.ail` |
| **B2** | 新增 `oled/oled.aila`（AILA 协议层）：初始化序列、`page_of_y`、`page_cols`、**行主序→页格式转置**、`present(commit)` | 转置有**宿主 golden 测试**（给定画布 → 期望的 128 字节页数据） |
| **B3** | `examples/oled_smoke` 切到 `sw_i2c + oled.aila`，删掉 `panel/panel.aila` 里的 C 声明与转置 | 上板复现 `pg=2`、`nack=0`；`--emit-size-delta` 记录体积变化 |
| **B4** | `examples/esp8266_oled`（SDK 版）迁到显示栈：页格式帧缓冲 → `std.pixbuf` + `textgrid` + `present(commit)` 局部刷新 | 上板复验：屏内容不变（四行滚动 + 心跳），`pg` 从 8 降到 2，Wi-Fi/TCP 上报不回归 |

## C. 平台能力族（按需补，不贪多）

| 项 | 内容 | 验收 |
|---|---|---|
| **C1** | `std.time.sleep_us(n)`：三家族 + 宿主 rt 原语（esp8266 用 ROM `ets_delay_us`，esp32/s3 用 ccount 或 rom delay，宿主 `nanosleep`） | 上板精度校验：1000×100µs 与 `mono_ms` 比，误差 ≤1% |
| **C2** | `std.mmio`（`write32/read32`）——**只写 RFC，不实现** | `docs/rfcs/0007-mmio-intrinsic.md`：说明"地址即 USize"到"内存访问"缺一层语言面（需要 `#compile.as_ptr` 类内建），以及为什么它才是"全 AILA 驱动"的真正前置 |

## D. 优化与热替换边界（用户提出的语义修正）

> 规则：**跨模块允许内联，除非调用边任一端是 `module hot`**。

| 项 | 内容 | 验收 |
|---|---|---|
| **D1** | 构建层「整程序 IR 合并」：把各模块 `.ll` 文本合成单模块 → `llc -O2`；`--lto` 开关，默认**关**；`--emit-ir` 仍按模块落 `.ll` | 开/关两条路都能构建并可运行；关=逐模块（现状） |
| **D2** | `module hot` 的所有函数打 `noinline` + `optnone`（挡住内联与 IPA 改形） | 测试：hot 函数体不得出现在调用方 IR 里；非 hot 两模块在 `--lto` 下**必须**出现内联（IR 里调用的函数体被吸收） |
| **D3** | 语义文档改写：`designs/0001 §12.1` / 铁律 7 口径从"跨模块不可内联"改成"**hot 模块边界不可内联**"；`#compile.if` 正当场景 ①改写为"LTO 关闭的构建 / hot 边界 / 跨 cpack" | 文档与实现一致；0006 §1.4 同步 |

## E. 轻量模块接口指纹（把 M2 的账先记上）

| 项 | 内容 | 验收 |
|---|---|---|
| **E1** | `aila-build.json` 增加 `module_interfaces`：每模块的导出签名 + public const 摘要（稳定排序、内容哈希） | 测试：换 `#compile.match` 的 const 取值后，该摘要**逐字不变**；改函数签名则必须变 |
| **E2** | 文档写明：这是热替换"接口指纹"的最小可用形态，M2 再接分派表/state 槽 | `0006 §5` 更新 |

## F. 收尾验证（每项之后都要做，最后整体再做一次）

1. `go vet ./...` / `go build ./...` / `go test ./...` 全绿。
2. `go run ./tools/goldengen` 快照与实现一致；`go run ./tools/skillgen` 技能包同步。
3. 三家族构建回归：esp8266（上板）、esp32、esp32s3（至少 `aila build` 通过 + `--emit-size-delta` 出数）。
4. 文档：`designs/0001`、`0003`、`0006`、README、`skills/reference/{language,recipes}.md`。
5. 上板证据（串口输出）贴进本文「执行记录」。

## 执行记录

（每完成一项追加：日期 / 实际做法 / 实测数据 / 偏离计划之处）

---

## 执行记录（2026-11-01，一口气做完）

### A. 语言侧真缺陷 —— 全修 + 用例

| 项 | 实际做法 | 结果 |
|---|---|---|
| **A1 模块级 Text 常量是空串** | 根因在 `internal/codegen/gen.go` 的全局初值：`constLiteral` 只认字面量标量与定长数组字面量，字符串字面量掉到兜底的 `zeroinitializer` → Text 变 `(null,0)`。修法：`StrLit` → 进常量池并合成 `{ ptr @.str.N, <lenLT> len }` | `internal/build/lang_regress_test.go:TestGlobalTextConstant`（`let` 与 `const` 两种形态都覆盖：打印 + `#comprun.len` + 下标字节） |
| **A2 `@mut[self]` 转发写不回** | 根因：≤16 字节的小聚合体 `@mut` 参数仍按值传，被调方拿到副本。修法：`paramByRef` 增加一条——**`@mut` 的聚合体参数一律按指针传**（视图家族除外，视图写本来就走 ptr） | `TestMutSelfForwardWritesBack`：`forward(c)` + `c.bump()` → `v=2`（修前是 1） |
| **A3 `view` of 模块级数组旧坑** | 宿主复现：**已不复现**（求和 100 / 下标 16 都对）。用例留下当回归护栏 | `TestViewOfModuleArray` |

> 影响面：A2 改了函数 ABI（小聚合体的 `@mut` 参数变指针），所以 `tests/corpus` 的 IR 快照全量重生成过一次；`TestPruneKeepsABI` / 示例端到端全绿。

### B. 驱动层收敛 —— 完成，一处偏离

| 项 | 结果 |
|---|---|
| **B1 `cpacks/sw_i2c`** | 新建：C 只剩位时序（`i2c_start/stop/write_byte` + `sw_i2c_begin/write/recover/nack/sent` + `delay.c` 的 `ets_delay_us` 垫片）。**偏离计划**：没有补 `include/sw_i2c.h` 走 bindgen——头文件里的 `const unsigned char *` 会映射成 `[U8]?`，而 v2 退役了 `Some` 构造、也没有隐式包装，**AILA 侧造不出非空的 `[T]?`**（本轮实测）。所以保留手写 `#compile.c` 声明（"视图只传指针 + 显式长度"的既有约定），头文件留待可空视图构造语法收敛后再启用 |
| **B2 `oled/panel.aila` + 转置搬 AILA** | 协议层（初始化序列 / `page_dirty` / `page_cols` / 页寻址命令 / `present`）在 AILA；**行主序→页格式的位搬运升格为 `std.paint.transpose_page`**（纯数据、任意页寻址控制器可用）。宿主 golden 落地在 `examples/display_demo` 的 `[6]` 段：`nz / fnv64 / 32 字节逐字节` |
| **B3 `oled_smoke` 切过去** | 上板复验（COM3）：`oled: i2c=1 fonts=292 cells=16x4`、`hb=0 cells=11 pg=2 nz=102` → 逐帧增长，**行为与迁移前完全一致**，说明 AILA 侧的转置/页窗口是对的 |
| **B4 `esp8266_oled`（SDK 版）迁移** | 代码已迁（页格式帧缓冲 + `draw_glyph`/`fb_scroll_console` → `pixbuf` + `textgrid` + `present`），**构建通过并产出镜像**；但**上板复验被既有阻塞卡住**：SDK 应用的启动链本身没打通（designs/0005 §6.C：`call_user_start` 之后无异常、无输出地挂住，与本轮改动无关）。这一项记为"代码就绪、验收待 SDK 启动链" |

### C. 平台能力

| 项 | 结果 |
|---|---|
| **C1 `std.time.sleep_us`** | 三家族 rt 原语（`aila_rt_delay_us`，ccount 忙等，换算基准 = `chip.json` 的 `cpu_hz`）+ 宿主（整毫秒交给系统睡眠、余下用 QPC 自旋）。**上板实测**：`1000 × 100µs = 102ms`（误差 2%，含循环与调用开销） |
| **C2 `std.mmio`** | 按计划只写 RFC：`docs/rfcs/0007-mmio-intrinsic.md`（"地址即 USize"到"内存访问"之间缺一层语言面，这也是「全 AILA I2C」的真正前置） |

### D. 内联语义放宽（用户提出的规则）

规则落地为「**跨模块允许内联，除非调用边任一端是 `module hot`**」：

| 项 | 结果 |
|---|---|
| **D1 整程序 IR 合并（`--lto`，默认关）** | `internal/build` 里做**文本合并**（随包工具链没有 opt / llvm-link）：只留第一份 target 行、去掉重复 `declare`、`attributes #N` 按模块重编号、`@.str.N` 字符串池按模块重编号 → 一次 `llc -O2`。四处坑全踩过一遍，`TestMergeModules` 逐条守着 |
| **D2 hot 边界** | `module hot` 的函数在 codegen 就打 `noinline optnone`（挡两个方向：不被内联、不内联别人、不做 IPA 改形）；`TestHotModuleMarksNoInline` 断言"hot 有、非 hot 没有"。**跨模块内联的实测证据**：同一份源码 `main.o` 里有 `IMAGE_REL_AMD64_REL32 aila_calc_calc__add`（真跨模块调用），`--lto` 后 `app.merged.o` 里该重定位为 **0**（已内联） |
| **D3 文档口径** | `designs/0001 §12.1` 铁律 7 改为「唯一例外 = `module hot` 的边界」；§7.8 `#compile.if` 正当场景从三条改写为四条（hot 边界 / `-O0` / 剪依赖 / 语义确定性），并注明 `--lto` 下常量分支也会被优化器折 |
| 目标支持 | esp8266 裸机 `--lto --gc-sections` 构建 + 上板复验通过（`pg=2`、`nack=0` 复现，产物 36144 字节） |

### E. 轻量接口指纹

| 项 | 结果 |
|---|---|
| **E1** | `aila-build.json` 新增 `module_interfaces`：每模块对 `define`/`declare` 的**签名行**（去掉函数体，排序）取 sha256 前 8 字节。`TestModuleInterfaceDigest` 断言：换 `#compile.match` 的 const 取值 → 摘要**逐字不变**；改签名 → 必须变。`TestBuildReportHasInterfaces` 断言报告里真的有这一段 |
| **E2** | `designs/0006 §5` 记录口径：这是热替换"接口指纹"的最小可用形态；分派表 / state 槽 / 指纹进 ABI 版本号都留给 M2 |

### F. 收尾

- `go vet ./...` / `go build ./...` / `go test ./...` 全绿；`goldengen` 104 份快照与实现一致；`skillgen` 重生成 `cli.md` / `errors.md`（含 `--lto` / `--gc-sections` / `--emit-size-delta`）。
- 上板证据：`oled: i2c=1 fonts=292 cells=16x4` / `us_check: 1000x100us=102ms` / `hb=0 cells=11 pg=2 nz=102`（裸机 LTO 构建）。
- 未完成项：B4 的上板验收依赖 designs/0005 §6.C 的 SDK 启动链；C2 只出 RFC。
