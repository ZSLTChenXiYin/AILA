# 开发计划 0001：管理器 + 编译器前端（M1–M4）

> **编号**：0001 · **类型**：开发计划 · **状态**：M1 已完成，M2 待启动
> **依据**：[rfcs/0001](../rfcs/0001-manager-and-frontend.md)、[specs/0001](../specs/0001-manager-cli.md)、[specs/0002](../specs/0002-compiler-frontend.md)

## 0. 傻瓜式开发流程（本计划的第一条验收）

开发者（含新手）只记三条命令，其余全部由诊断引导：

```text
aila new demo        # 1. 建工程（生成 aila.toml + main.aila，不需要懂目录约定）
aila doctor          # 2. 环境自检（缺什么，直接把要执行的命令打印出来）
aila run             # 3. 编译 + 链接 + 运行（出错时给错误码 + 修法 + 设计锚点）
```

### 0.1 里程碑节奏（每个里程碑都必须是「能跑的一条命令」）

```text
M1 能跑一整个示例           → aila run examples/hello
M2 能写真实嵌入式控制逻辑     → 区域 + 定长数组 + C 调用 + 错误传播
M3 能热替换                 → hot 模块 + 分派表 + 双槽装载器
M4 能并发                   → 协程 / chan / thread
```

每个里程碑结束时的动作固定：跑 `python -m unittest discover -s tests`、跑 `aila doctor`、跑 `examples/` 全量、更新本文「状态」列。

## 1. 里程碑与批次

### M1 — 骨架打通（本批重点）

| # | 任务 | 验收 | 状态 |
|---|---|---|---|
| M1-1 | 仓库骨架、`aila` 启动器、`docs/` 编号规约 | `aila --version` 可执行 | 进行中 |
| M1-2 | 诊断内核：错误码表、源码定位渲染、`explain` | `aila explain E3201` 输出修法 | 待开始 |
| M1-3 | 管理器：工程发现、清单解析、target 表、工具链探测 | `aila new` + `aila doctor` | 待开始 |
| M1-4 | lexer + parser：全 v0 语法 + 解析禁区诊断 | `tests/test_parser.py` 全绿 | 待开始 |
| M1-5 | types + 布局（C ABI、packed/align、tag/niche） | `tests/test_layout.py` 与 C `sizeof` 对齐 | 待开始 |
| M1-6 | sema：名字 / 类型 / 区域 / 效果 / match / 丢弃 / 主名字 | `tests/test_sema.py` 全绿 | 待开始 |
| M1-7 | codegen：IR 文本（函数、控制流、陷阱检查、区域、视图、Result） | `aila build --emit ir` 且 `mod.verify()` 通过 | 待开始 |
| M1-8 | 后端落地：`llvmlite` → `.o`；`cc` 链接；runtime C | `aila run` 打印预期输出 | 待开始 |
| M1-9 | bindgen 双档（clang AST / builtin 兜底） | `aila bindgen cpacks/demo` 产 `.ail` | 待开始 |
| M1-10 | 示例 6 个 + 文档收口 | `examples/` 全跑通 | 待开始 |

M1 冻结条件（**四条全绿才算完**）：
1. `python -m unittest discover -s tests` 全绿；
2. `aila run examples/<每个>` 输出符合期望；
3. `aila doctor` 全 ✔；
4. `aila build --emit ir` 的 `.ll` 能被 `llvmlite` 与（若装了）`llc` 双双验证。

### M2 — 能写真实逻辑（下一批）

| # | 任务 | 说明 |
|---|---|---|
| M2-1 | 借用区间（NLL 式）检查 | §5.3 读写互斥精确化，替换 M1 的保守版 |
| M2-2 | 区域配额运行期账本 | `region r: 4096` 的 `used+size>cap` trap |
| M2-3 | `std.text` / `std.str` / `std.math` / `std.mem` 全量签名落地 | 纯 AILA 实现 |
| M2-4 | bindgen：C shim 骨架生成、`#ifdef` 定型、packed 透传；分发裁剪版 Clang | 覆盖 `clang` 档 |
| M2-5 | 类型化物理域落位（链接脚本段生成）+ `esp32s3` 交叉产物 | `aila build --target esp32s3` 出 `.o` |
| M2-6 | 诊断质量批次：`W9xxx` 待批次项逐个消账 | 把「未校验」条目清到 0 |

### M2-D — 分发：随包工具链 + 发布流水线（本批已完成）

| # | 任务 | 验收 | 状态 |
|---|---|---|---|
| M2-D1 | `AILA_HOME` 相对目录解析（`share/aila/{runtime,stdlib}` 与仓库布局双兼容） | `aila doctor` 首行显示正确根目录 | ✅ |
| M2-D2 | 工具链解析改为**随包优先**，`system:` 才外找，`bundled` 缺则不降级 | `tests/test_toolchain.py` 全绿 | ✅ |
| M2-D3 | 版本带绑定 + 带外不采用（`W8012`） | 带外 bundle 解析 `ok=false` | ✅ |
| M2-D4 | `aila toolchain info|l list|install` | 解 zip 后解析切到「随包」 | ✅ |
| M2-D5 | CI：`ci/fetch-llvm.*` + `ci/assemble-bundle.*` + `release.yml` | 首次 release 三平台 `aila doctor` 全 ✔ | ⏳ 需 CI 实跑 |

> 已知边界：`install` 不联网（包来源交给发布页/内网）；Linux/macOS 的 LLVM 拉取路径属**推断实现**，
> 首次 release 会暴露差异——已在脚本头标明「首次 release 需人工验证」。

### M2-G — Go 重写（见 [rfcs/0003](../rfcs/0003-go-rewrite.md)）

| # | 批次 | 范围 | 状态 |
|---|---|---|---|
| G1 | 骨架 + 无前端命令 | `cmd/aila` + `internal/{cli,home,resolve,toolchain,targets,manifest,diag}` | ✅ 完成（go test 全绿） |
| G2 | 词法/语法/AST | 与 Python 版逐 token / 逐 AST 对齐（黄金语料） | ✅ 完成（7 份语料，token 与 AST 双对齐） |
| G3 | 语义 + IR | 六示例 IR 字节级对齐 | ✅ 完成（`ir_basic` + `ir_wide` 逐字节一致） |
| G4 | 切换与清理 | CI 用 `-Manager go`；删 Python 前端 | ✅ 完成（多模块语义 / 构建流水线 / bindgen / CI 与发布切 Go / Python 前端已删） |

### M3 — 热替换

| # | 任务 | 说明 |
|---|---|---|
| M3-1 | 分派表：跨模块调用改间接调用（铁律 7） | `hot.json` 驱动，跨模块不内联 |
| M3-2 | MLO 产物（重定位表 + 段属性）、`verify` 哈希与 ABI 版本核对 | §2.4 六步的前两步 |
| M3-3 | 双逻辑槽装载器 + 排水协议 + state 迁移 + 世代 trap | 运行时（C 侧）实现 |
| M3-4 | `state` 槽跨版本迁移完备性检查（改字段未写 `on_load` = 警告） | 前端检查项 |

### M4 — 并发

| # | 任务 | 说明 |
|---|---|---|
| M4-1 | 无栈协程：`llvm.coro.*` 降级（`coroutine` / `yield` / `resume` 三态） | §10.2–10.5 |
| M4-2 | Channel：SPSC 双端 + `<-` 收发 + 满/空挂起 | §10.6 |
| M4-3 | `thread`：FreeRTOS / pthread 两种封装 + 栈大小字节→words 换算 | §10.8 |
| M4-4 | `std.atomics` / `std.mmio` | §10.10、§13.6 |

## 2. 依赖顺序（不可跳）

```text
诊断内核 ─► 管理器发现/清单 ─► lexer ─► parser ─► types/布局 ─► sema ─► codegen ─► 后端+链接 ─► 示例
                                                   └────────► bindgen（独立，可与 sema 并行）
```

`bindgen` 只依赖「types 的拼写规则」与诊断，不依赖 sema/codegen，故可与 M1-5..M1-8 并行推进。

## 3. 风险与对策（对应 rfcs/0001 §5）

| 风险 | 触发信号 | 对策 |
|---|---|---|
| 无 LLVM 后端 | `aila doctor` 第 2 项 ✘ | 三级供给 + 直接打印 `pip install llvmlite`；IR 文本仍可产出供审计 |
| 语义检查做过头，M1 发不出去 | M1 超过「示例能跑」所需 | 检查项按 §6 表批次落地，未落地项打 `W9001` 而不是假装通过 |
| bindgen 兜底档覆盖面不足 | 示例里 `import ch` 需 clang | 示例只用兜底档能覆盖的 C 声明子集；M2 补 clang 档 |
| Python 性能 | 单次全量 > 5 s | 模块级缓存（specs/0001 §5.3）+ 单线程直跑 |

## 4. 当前批次（M1）执行清单 — 勾选式

- [x] docs 编号规约（各目录从 0001 起）+ 索引
- [x] designs/0001（重编号）
- [x] rfcs/0001
- [x] specs/0001、specs/0002
- [x] plans/0001（本文）
- [x] `aila` 启动器 + `--version` / `--help`
- [x] 诊断内核 + 错误码表 + `explain`
- [x] 管理器（new/doctor/check/build/run/clean/bindgen/target）
- [x] lexer / parser / ast / types / sema / codegen
- [x] runtime（区域账本 + trap 通道 + 打印出口）
- [x] bindgen（clang / builtin 双档；本机环境 clang 缺失 → builtin 档实测通过）
- [x] 最小 std（io / math；`std.mem` 随内容比较落到 M2）
- [x] 示例 6 个 + 测试套件（53 条）
- [x] README（三条命令起步）

> 偏差登记：`std.mem` 未在本批创建（`==` 对视图已按 §6.4 给出 `E3205` 诊断，内容比较函数随 M2 的 std 批次一起进）。

## 5. 提交纪律

1. 每个任务完成即勾选，**不批量勾**；
2. 任何与 `designs/0001` 冲突的实现，先改文档或先报缺陷，不在代码里发明语义；
3. 错误码只增不改语义；废弃码标注 `(deprecated)` 并保留解释条目；
4. 每个里程碑结束必须在本文追加「实际耗时 / 偏差 / 下一批调整」三行小结。

## 6. M1 小结（实测记录）

| 项 | 结论 |
|---|---|
| 实际产出 | 管理器（9 命令）+ 前端（lexer/parser/ast/types/sema/codegen）+ LLVM 后端三级供给 + bindgen 双档 + runtime + stdlib(io/math) + 6 示例 + 53 条测试 |
| 偏差 1 | 本机无 clang、无 Rust/LLVM：前端落 Python，后端落 `llvmlite`（已在 `aila doctor` 明示） |
| 偏差 2 | 后端必须用 `reloc="static"`：Windows/MinGW 下默认重定位会引用 `_GLOBAL_OFFSET_TABLE_`，ld 无此符号；已写进代码注释防「优化掉」 |
| 偏差 3 | 模块级增量缓存（specs/0001 §5.3）只落地了 bindgen 缓存，`.o` 全量重编；接口哈希已在 `ModuleFacts.interface_hash()` 备好，M2 接上 |
| 偏差 4 | 实现期发现设计稿两处内部不一致，已另立 [rfcs/0002](../rfcs/0002-design-defects.md) 裁决，未擅改设计稿 |
| 下一批优先级 | M2-1（借用区间）> M2-3（std 全量签名）> M2-5（esp32s3 交叉落位）> M2-6（消 `W9001` 账） |
## 7. M2 小结（分发批次）

| 项 | 结论 |
|---|---|
| 已落地 | 相对目录 `AILA_HOME`、随包优先解析、版本带绑定、`aila toolchain` 三命令、CI 打包脚本与 release 工作流 |
| 行为变化 | 默认**不再"先查 PATH"**；开发期无 bundle 时回退到 llvmlite/系统 cc，并在 `doctor`/`version`/`toolchain info` 三处标注「开发回退」 |
| 未验证 | 真机 bundled 编译（本机无官方 clang/llvm 二进制）与 CI 三平台打包——U 标在 M2-D5 |
| 与 Go 的关系 | 真正「免安装」要等 G4 换成 Go 单文件；当前 release 包若用 python 形态，仍要求用户机有 Python 3.11 |
## 8. G1 小结（Go 管理器第一批）

| 项 | 结论 |
|---|---|
| 已落地 | cobra CLI 外壳 + 7 个命令；随包工具链解析（与 Python 同规则）；错误码/target 表**共享 JSON**；doctor 含链接冒烟测试 |
| 行为一致性 | `version` / `toolchain info` / `explain` / `target show` 输出与 Python 版同义；`doctor` 表项一致 |
| 差异（有意） | Go 版不支持 `llvm = "llvmlite"`（不引 cgo），显式报错并指路随包 llc/clang |
| 待办 | G2：lexer/parser/ast；同时把 `tests/corpus` 黄金语料生成器接上 |
## 9. G2 小结（进行中：语料 + 词法）

| 项 | 结论 |
|---|---|
| 黄金语料 | `tests/corpus/*.aila` 7 份（hello / decls / types / exprs / control / variadic / errors），覆盖声明、类型拼法、表达式优先级、三种 for、match 载荷模式、region/defer/unsafe/thread/coroutine、extern 变参、解析期错误 |
| 生成器 | `tools/gen_corpus.py` 产出 `*.tokens.json` / `*.ast.txt`（S-表达式）/ `*.lexdiags.json` / `*.diags.json`；`--check` 供 CI 校验快照未过期 |
| Go lexer | `internal/lexer` 与 Python 版**逐 token 一致**（7 份语料全绿），词法诊断码（E1002/E1003/E1006）一致 |
| 语料当场抓出的 7 个真缺陷 | ①`send chan`/`recv chan` 类型解析用错 `at_kw`（内置类型名不是关键字）②`copy x @[r]` 的区域实参被操作数的后缀链吞掉 → 假 `E2007` ③`#[c] let x: T;` 无初始化器被误报 ④`else if` 条件里的 `x {` 被当成构造字面量 ⑤`...` 变参在词法层没有 token ⑥原始字符串 `r"abc"` 把 `r` 吃两次（Go 版照抄同处，已双边修复，直接后果是 `r"no\escape"` 读成 `o\escape`）⑦`send chan T` 的通道方向丢失：解析器取 token 的 `kind`（恒为 `ident`）而不是 `text`，`send`/`recv` 端类型因此无法静态区分 |
| 下一步 | G2 剩余：`internal/ast`（节点 + `Fields()` 序列化）+ `internal/parser`，用 `*.ast.txt` 做逐节点比对 |

## 10. G2 收口（词法 + 语法 + AST 全部对齐）

| 项 | 结论 |
|---|---|
| `internal/ast` | 全部节点（声明 / 语句 / 模式 / 表达式）实现 `Kind()` + `Fields()`，字段名与顺序严格对齐 Python dataclass；`Sexp()` 负责契约快照序列化（空值省略、浮点带 `.0`、字符串按需 JSON 转义） |
| `internal/parser` | 完整 v0 语法移植：模块壳/块划分、`@` 效果子句（前置与后置都接受）、全部类型拼法、三种 `for`、match 载荷模式、`out`、`defer`、`unsafe`、`thread`、`coroutine`、变参、构造字面量、`?(...)`/`?`/`as`/`trunc`、`copy @[r]` 等在列 |
| 判据 | `TestCorpusAST`：6 份语料 AST 与 Python **逐字符一致**；`TestCorpusParseDiags`：错误语料的诊断码集合一致 |
| CLI | 新增 `aila parse <文件>`（默认打印 S-表达式，`--tokens` 打印 token 流），退出码与诊断风格与其他命令一致 |
| 规模 | Go 5956 行（其中测试 952 行），10 个包 `go vet` / `go test` 双绿 |

## 11. G3 进度（类型与布局）

| 项 | 结论 |
|---|---|
| `internal/types` | 类型项（Prim/View/Fixed/Unsafe/Optional/Result/Chan/Coro/FnPtr/Group/Named + StructInfo/EnumInfo）与 **C ABI 布局**（size / align / 字段偏移 / 枚举 tag 宽度）；拼写与 Python `describe()` 逐字一致 |
| `internal/sema`（第一片） | `BuildTable`：两段式类型表（先登记名字、再解析体内引用）、TypeRef → Type、最小编译期常量求值（枚举判别值）、`#[packed]`/`#[align(N)]` |
| 新判据 | `tests/corpus/layout.json`（13 个类型、12 个别名）由 `tools/gen_layout.py` 从 Python 生成；Go 侧 `TestLayoutMatchesReference` 逐项比对 size/align/字段偏移/tag 宽度/变体 tag —— 全绿 |
| 语料 | 新增 `tests/corpus/layout.aila`（含 `#[packed]`、`#[align(16)]`、嵌套 struct、视图/定长数组字段、四种 enum、12 类型别名） |
| 又抓到的缺陷 | 缺陷 4：无载荷 enum 的宽度没有跟随底层类型（`enum Reg: U8` 实占 4 字节），已双边修正（见 rfcs/0002） |
| 下一步 | G3 剩余：函数签名与名字解析 → 表达式类型检查 / 区域效果 / match 完备性 → `internal/codegen`（判据：`*.ir` 快照与六示例端到端） |

## 12. G3 进度（函数签名）

| 项 | 结论 |
|---|---|
| `internal/sema/sig.go` | `FuncSig`（形参/返回/错误通道/多值组/`@` 三族/染色/公开/变参）+ `BuildSigs`；`self` 免标注、类型恒为前缀类型；`@` 子句的顺序/重复/互斥/参数存在性校验 |
| 新判据 | `tests/corpus/sigs.json`（16 函数、4 方法、4 常量、1 state）由 `tools/gen_sigs.py` 生成；Go 侧 `TestSigsMatchReference` 逐字段比对 —— 全绿 |
| 语料 | 新增 `tests/corpus/sigs.aila`：普通函数、无返回值、`@region+@mut` 组合、`@move` 消耗、方法/关联函数、`@mut[self]`、协程与线程染色、extern 变参、`T ! E`、多值组、函数指针返回与形参、`T?` 返回、hot 模块 state |
| 又抓到 4 个缺陷 | 缺陷 5（`@mut` 从未被解析，能力级）／6（`same_type(None,None)`）／7（单名函数指针调用返回 I32）／8（region·unsafe 块内 return 不算发散）——均已双边修正并登记 rfcs/0002 |
| 下一步 | 作用域与表达式类型检查（`let/var` 绑定、字面量多态、运算符、调用/方法、match、`?`/`as`/`trunc`）→ 区域效果与逃逸 → codegen |

## 13. G3 进度（作用域与类型检查）

| 项 | 结论 |
|---|---|
| `internal/sema/scope.go` | `Symbol` / `Scope` 作用域链、`FuncCtx`（`@region` 参数 / 体内 region / 协程·线程 / unsafe 深度 / 返回通道）、`Checker` 的全套类型注解表（表达式类型、绑定类型、区域来源、调用签名、cast 端点、match 结果…） |
| `internal/sema/check_*.go` | 语句层：绑定/`view`/`return`/`if`/三种 `for`/`match`/`region`/`defer`/`unsafe`/`trap`/`yield`/`thread`/赋值/复合赋值/解构/丢弃/表达式语句；表达式层：字面量多态、路径（局部/全局/常量/函数/枚举变体/字段链）、调用（函数/方法/关联函数/函数指针间接调用/`self`）、下标、字段、二元（含比较链禁令）、一元、`as`/`trunc`（无损判定）、`?` 传播、`?(...)` 检测、`len`/`copy`/`move`/`resume`/`<-`、数组与构造字面量（含 `Some`/`Ok`/`Err`）、分配点、`#size`/`#align`、`match`（完备性 + 分支产出统一 + 模式绑定） |
| 新判据 | `tests/corpus/bindings.json`（3 文件 / 33 条绑定）由 `tools/gen_exprtypes.py` 生成；Go 侧 `TestBindingsMatchReference` 逐条比对绑定名与推断类型 —— 全绿；另有 `TestCheckReportsBasicErrors` 覆盖 7 条基础诊断 |
| 语料 | 新增 `tests/corpus/sema.aila`（10 个函数：算术/位运算/比较/逻辑、类型转换、视图与分配、字面量与构造、三种 match、`?` 传播） |
| 修掉的实现缺陷 | Go 侧字面量范围检查在 64 位目标上移位溢出（`1<<64` → 0），把 `buf[0]`、`arr[0]` 这类下标误报成「字面量超出值域」 |
| 下一步 | `internal/codegen`：IR 文本生成（函数/控制流/陷阱检查/区域/视图/Result），判据是六示例 `*.ir` 与 Python 版逐行比对 + 端到端 `run` |

## 14. G3 收口（代码生成与参考实现逐字节一致）

| 项 | 结论 |
|---|---|
| `internal/codegen` | `gen.go`（`Lowerer`：常量池 / 字符串池 / 全局 / `declare` / 类型映射 / 签名返回 / 零值 / 浮点与整数常量）、`func.go`（`funcGen`：槽位分配、参数落槽、`if`、`return`、`defer`、`region` 块、入口 wrapper）、`stmt.go`（三种 `for`、`match` 判别与绑定、赋值/复合赋值/解构、检查插入与 `aila_rt_trap`、`Result` 构造）、`expr.go`（字面量、路径与字段链、左值、数组/构造/分配字面量、`len`、算术与陷阱、短路、一元、转换、调用、`?` 传播） |
| 代码生成注解层 | 新增 `internal/sema/cginfo.go`：`FieldRefs` / `FieldChains` / `IndexRefs` / `ForRefs` / `ArmRefs` / `MatchRefs` / `LitRefs` / `ArrRefs` / `AllocRefs` / `TryRefs` / `LenRefs` / `DestructRefs` / `Locals` / `ParamSyms` / `CallRecv` —— `codegen` 只读这些注解，不回头改 AST（快照判据不受影响） |
| 判据 | `tests/corpus/ir_basic.ir`（136 行）与新增 `tests/corpus/ir_wide.ir`（758 行）**逐字节一致**；Go 侧 `TestIRParity` 与 CLI 侧 `TestIRCommandMatchesSnapshot` 双路锁定 |
| 新增语料 | `tests/corpus/ir_wide.aila`：struct + 方法、enum（无载荷/带载荷）、`match` 四种判别（enum / 整数 / `T?` / `T ! E`）、`?` 传播、定长数组与 `for-in`、视图分配与 `region` 块、三段式 `for` + `break`/`continue`、显式转换（`as` / `trunc`）、短路、解构赋值、`#size`/`#align` |
| 新命令 | `aila ir <文件>`：单文件「解析 → 语义 → IR」，无 clang/llc 也能看到前端产物，IR 与快照逐字节一致 |
| 修掉的实现缺陷 | 缺陷 9（取值位置 `match` 的字面量分支产出 i64 却存进 i32 槽 → 非法 IR）、缺陷 10（一元运算不下推上下文类型，`-1` 生成 `sub i64`） |
| 下一步 | G4 前置：多模块语义（`Program` / 模块图 / `std` 与 `cpacks` 绑定）+ 构建流水线（IR → 目标文件 → 链接），判据是六示例端到端输出一致 |

## 15. G4 进度（多模块语义 + 构建流水线）

| 项 | 结论 |
|---|---|
| `internal/project` | 工程模型：递归发现模块（跳过 `build`/`cpacks`/`.git` 等保留目录）、`E7003` 校验（模块名 = 文件名、包根只允许 `main.aila`、子包名 = 目录名）、`std.*` 标准库路径回退、cpack 目录收集 |
| `internal/sema/program.go` | 模块图 `Program`/`Unit`：两段式类型表（`DeclareTypes` → 链接导入 → `ResolveBodies`）、导入绑定名解析（`as` 别名 / 末段默认）、`Checker.SetModuleContext`（跨模块名字解析）、`findMethod` 按类型定义模块找方法 |
| 跨模块能力 | 类型名 `cam.Frame`（`Table.Imports` + `splitQualified`）、值路径 `io.print_i32`、枚举变体 `cam.Color.Red`、跨模块常量/全局、跨模块方法；符号修饰用**声明模块**（`aila_std_io__print_text`） |
| `internal/build` | 流水线：`Load`（清单 → 模块发现 → 解析 → 多模块语义）、`Check`（+ `W9001` 待批次警告）、`Build`（IR → `.o` → C 运行时 → 链接 → 报告）、`GenerateIR`（与判据共用，不碰后端工具） |
| CLI | `check` / `build` / `run` 从 `E9003` 占位变成真实命令；新增开关 `--board/--opt/--llvm/--cc/--emit-ir/--emit-json/--out`；`build`/`run` 走随包工具链优先 |
| 判据 1 | `internal/build/TestCheckExamples`：Go 版独立完成 hello/control/region/struct/result 五个示例的「解析 + 多模块语义」（cinterop 待 bindgen） |
| 判据 2 | `internal/build/TestExampleIRParity`：五个示例、十个模块（含 `std.io`）的 IR 与 `tests/corpus/examples/**/*.ll` **逐字节一致**（快照由新工具 `tools/gen_examples_ir.py` 生成） |
| 判据 3 | 端到端执行：Go 前端 IR 经 `tools/link_ir.py`（llvmlite 降目标文件 + gcc 链接）跑出的 stdout 与 Python 版**逐字节一致**（本机无 clang/llc 时的验证通道） |
| 修掉的实现缺陷 | 缺陷 11（Windows 三元组取 llvmlite 的 `msvc`，与 MinGW 链接器冲突）、缺陷 12（`.ll` 行尾随平台漂移）、缺陷 13（IR 内嵌绝对路径，产物不可复现） |
| 下一步 | G4 剩余：`internal/bindgen`（`import ch`/`import cc` → `.ail`）+ CI/发布流水线切到 Go 单文件 + 删除 Python 前端 |

## 16. G4 进度（bindgen + CI 切 Go）

| 项 | 结论 |
|---|---|
| `internal/bindgen` | C 头 → `.ail` 两档后端：`clang`（`-ast-dump=json` 走 AST：TypedefDecl / RecordDecl / EnumDecl / FunctionDecl / VarDecl）与 `builtin`（自写 C 词法器 + 轻量预处理 + 顶层声明切分）。命名规范（`type_name`/`value_name`）、类型映射（含 `char *` → `[U8]?`、`void *`/多级指针 → `[Unsafe]?`、`long` 按目标字宽）、宏定型（`#define` 对象式宏 → `const`）与参考实现逐字节一致 |
| 绑定判据 | `tests/corpus/bindings/mathkit.ail`（`tools/gen_bindings.py` 生成）；Go 侧 `TestMathkitBindings` 逐字节比对，另有 `TestMapCType` / `TestTypeAndValueNames` / `TestTokenizeAndPreprocess` 守住映射表与命名规范 |
| 构建接线 | `import ch _.cpacks.<名字>` → 生成绑定模块挂进模块图（带 `build/.cache/bindgen` 缓存，键 = 头文件哈希 + 档位 + 三元组）；`import cc` 只登记 C 源一起编译；cpack 的 `*.c` 与 `-I include` 进 C 侧编译，`*.a/*.lib/*.so` 进链接 |
| cinterop 打通 | `examples/cinterop` 走 Go 版全链路（bindgen → sema → IR → `.o` → C 链接）跑出的输出与 Python 版逐字节一致；IR 也进了示例快照 |
| 端到端判据 | `internal/build/TestExamplesEndToEnd`：六示例从源码编到可执行文件、运行 stdout 与 `tests/corpus/examples/<名字>/expected.out` 逐字节一致；没有 LLVM 后端 / C 编译器的机器上按 E8010/E8011 **显式跳过**（不是假装通过） |
| 示例清单 | `exampleNames` 扩到六个（含 cinterop）；`tools/gen_examples_ir.py` 同步覆盖，快照里多了 `cpacks_mathkit.ll` |
| CI | 三条 Job：`go`（三角色矩阵 Linux/macOS/Windows，装 llvm/clang/mingw → `go vet`/`go build`/`go test -v` → 单文件产物冒烟 `aila version/target/check/build/run`）、`go-cross`（三平台交叉编译上传 artifact）、`reference`（过渡期 Python 参考实现，可用仓库变量 `AILA_KEEP_PYTHON_REF=0` 关掉） |
| 发布 | `release.yml` 与 `ci/assemble-bundle.{ps1,sh}` 的默认管理器切到 **go**（`vars.AILA_MANAGER` 可覆盖） |
| 修掉的实现缺陷 | 构建失败路径上 `Result.OK` 没落回 false（早退时仍报「成功」），导致端到端判据会误判 |
| 下一步 | G4 收尾：删除 Python 前端（`src/aila`、启动器、Python 测试与生成器）并把 `reference` Job 一并删掉 |

## 17. G4 收口（Python 前端删除，AILA 只剩 Go 一份实现）

| 项 | 结论 |
|---|---|
| 快照生成器换 Go | 新增 `tools/goldengen`（`go run ./tools/goldengen [-check]`）：token / lexdiags / AST 或诊断码 / 布局 / 签名 / 绑定类型 / 单模块 IR / cpack 绑定 / 示例每模块 IR 全部由 Go 生成；输出格式复刻参考实现的 `json.dumps` 细节（`indent=0` 的换行不缩进、默认分隔符带空格、浮点恒带小数点），因此**移动前已生成的快照逐字节不变** |
| 迁移验收 | 用 Go 版生成器重跑一遍后与旧快照逐字节比对：除 `bindings.json` 外全部一致；`bindings.json` 的唯一差异是 Go 的单模块语义检查不做入口检查（`E7005`），故 `variadic.aila` 不再落在 `skipped` 里——这条差异已确认无害（`TestBindingsMatchReference` 只比对 `files` 里的条目） |
| 删除清单 | `src/aila/`（38 个文件的前端 + 管理器）、`src/aila_launcher.py`、`aila`（POSIX 启动器）、`aila.cmd`、`tests/*.py`（9 个 unittest 套件）、`tools/*.py`（8 个快照生成/接线脚本）、`bin/aila.exe` 构建残留 |
| 保留清单 | `internal/*`（16 个包）、`cmd/aila`、`share/aila/data/*.json`、`stdlib/`、`runtime/`、`examples/`、`tests/corpus/**`（契约快照，现在是**冻结的事实**，由 Go 测试守护）、`tools/goldengen`、`ci/`、`docs/` |
| CI | 去掉 `reference` Job；`go` Job 增加一步 `go run ./tools/goldengen -check`（快照一致性），其余不变 |
| 文档 | `README.md` 改成 Go 单文件的说明；`rfcs/0001` 标注为历史决策记录（并说明被误删后已据可见片段重建）；`rfcs/0002` 顶部加「实现落点 → Go 包」映射表；`specs/0001–0003` 里的 `src/aila/*.py` 全部指向 `internal/*` |
| 判据 | `go vet`/`go build`/`go test ./...` 全绿（16 个包）；`go run ./tools/goldengen -check` 报「全部快照与当前实现一致」 |
| 遗留 | 端到端运行判据（`TestExamplesEndToEnd`）在本机因无 LLVM 后端跳过，由 CI 的 llvm/clang 环境真跑；协程/通道/线程代码生成、热替换 MLO、NLL 精确检查仍在 M2+ 计划内 |

## 18. 遗留项（下一批要么做、要么明确不做）

| 项 | 现状 | 影响 | 建议 |
|---|---|---|---|
| ~~**Windows 发布包的 ABI 供给**~~ | ✅ 已解决（§19）：脚本换成 llvm-mingw，随包自带 MinGW 运行时与 lld | — | 实测「干净 PATH + 与仓库无关目录」下六个示例全部可编译运行 |
| 协程 / 通道 / 线程代码生成 | 语义检查到位，codegen 报 `E9002` | 这三个特性还编不出目标文件 | M2+ 批次 |
| 热替换 MLO 与分派表 | 未开始 | `hot module` 仍走普通链接 | M3 |
| 借用区间（NLL）精确检查 / 区域配额静态证明 | 未开始，`aila check` 以 `W9001` 明示未校验 | 不静默通过，但也没有保护 | M2+ |
| `aila bindgen` 单独命令 | 报「尚未实现」；绑定目前由 `import ch`/`import cc` 在构建流水线里自动生成 | 想单独产出绑定要进 `build/.cache/bindgen/` 取 | 需要时补一个命令（`internal/bindgen` 已就绪） |
| `docs/rfcs/0001` 原稿 | 上一轮被误清空后**据可见片段重建**（文首已标注） | 措辞可能与原版有差异 | 有原稿副本时覆盖回去 |
| 设计稿回写 | `rfcs/0002` 登记了 13 条缺陷裁决，`designs/0001` 尚未回改 | 权威文档与实现口径仍有文字层面的差异 | 设计批次统一修订 |

## 19. 路线 A 落地（llvm-mingw 随包 → 真正的「解压即用」）

| 项 | 结论 |
|---|---|
| 为什么换 | 官方 LLVM win64 发行版只有 clang + lld，**不带** MinGW 运行时；AILA 的 Windows 三元组是 `x86_64-w64-windows-gnu`，链接必须有 CRT/导入库 → 之前打出的包在没有 MinGW 的机器上链接失败 |
| 做法 | `ci/fetch-llvm.{ps1,sh}` 改为拉 **llvm-mingw**（20261006 / LLVM 23.1.3 / ucrt），按固定规则裁剪到 `toolchain/<triple>/`：clang 系 + lld + clang 资源目录（只留宿主架构 builtins）+ C 头 + `x86_64-w64-mingw32/lib`；丢掉 lldb/python/clangd/sanitizer/C++ 头 |
| 坑 1 | `clang.exe` 用 libc++ 构建并动态链接：丢掉 `libc++.dll`/`libunwind.dll`/`libwinpthread-1.dll` → `STATUS_DLL_NOT_FOUND`（0xC0000135）。这三件必须随包 |
| 坑 2 | `VERSION` 首段必须是 LLVM 版本（doctor 按它判 `18..23` 版本带），否则误报 `W8012` |
| 坑 3 | 有了随包 clang 后 bindgen 会走 **clang 档**：clang 的 JSON AST 会把系统头与内建 typedef（`__int128_t` 等）一起吐出来。修法：按「声明所在文件是否在 cpack 目录内」过滤（`declInCpack`），并把宏常量映射抽成两档共用的 `collectMacroConsts`（`#define` 只在预处理阶段可见） |
| 坑 4 | bindgen 缓存键没带「映射规则版本」，改了映射仍读旧缓存 → 现在键里带 `bindgenSchema` |
| 实测 | 裁剪后 **313 MB**、压缩包 **85 MB**；`PATH` 只留 `C:\Windows\System32;C:\Windows`（无 clang/gcc/Go/Python/MinGW）时，把 zip 解压到与仓库无关的临时目录后：`aila version` / `doctor`（6 项含链接冒烟全绿）/ `run examples/hello` / `new demo` + `run demo` / 六个示例 `build`+运行输出与 `expected.out` 逐字节一致 —— **零安装** |
| 新判据 | `internal/toolchain/TestBundledToolchainSmoke`：仓库里有 `toolchain/<triple>/` 时，`bundled` 偏好下 doctor 不得有 fail（无随包工具链的机器上跳过） |
| 发布包内容 | 新增 `examples/`（`aila doctor` 的提示与「解压即用」体验要用），外加 `LICENSE`/`NOTICE`/`NOTICES.md`/`README.md`/`VERSION`/`share/aila/**`/`toolchain/**` |
| 遗留 | `fetch-llvm.sh` 的 Linux/macOS 分支只做了静态审阅与逐条对齐 ps1 规则，**没在真机上跑过**（本机是 Windows）；首次 release 前需要在 CI 上跑一次并修差异。sha256 需要在发行前用 `LLVM_MINGW_SHA256` 固定 |

## 20. 跨项目依赖落地（包管理）

| 项 | 结论 |
|---|---|
| 概念澄清 | **包 = 项目**（一个 `aila.toml`）、**子包 = 目录**、**模块 = 文件**；`import <依赖名>.<子包>.<模块>` 的第一个名字是**依赖名**（= 清单 `[dependencies]` 的键），不是包内的子包 |
| 依赖形态 | 路径依赖 `{ path = "../mylib" }` 与 git 依赖 `{ git = "…", tag/rev = "…" }`（**分支不许用**）；可选 `version = "0.3"` 前缀要求 |
| 定位顺序 | `vendor/<名>/` 优先（离线）→ `path` → git 缓存（`AILA_DEPS_CACHE` / `$AILA_HOME/cache/deps` / 用户缓存），未命中才 clone |
| 锁文件 | `aila.lock` 记 `source` / `version` / `hash`（`aila.toml` + 全部 `.aila` 的内容哈希）；三者任一漂移 → `E7023`，须 `aila deps --update`；首次构建自动写；锁文件**不含本机绝对路径** |
| 编译模型 | 依赖源码一起进模块图（源码分发、无预编译库），模块名带包前缀 → 符号天然隔离；按拓扑序编译，依赖的 `.o` 一起链接 |
| 可见性 | 跨包只认 `public`（`E7020`）；`stdlib` 的对外函数补了 `public`；同包内跨模块不受影响 |
| 新错误码 | `E7020` 未导出 / `E7021` 声明不合法 / `E7022` 定位失败 / `E7023` 锁不一致 / `E7024` 同名冲突 / `E7025` 成环（全部写进 `share/aila/data/diag_codes.json`） |
| 新命令 | `aila deps [--update] [--vendor] [--json]`；`check/build/run` 加 `--deps-update` / `--vendor` |
| 判据 | `internal/deps` 六个单测（路径 / vendor 离线 / 版本 / 成环 / 声明 / 版本匹配）+ `internal/build/TestCrossProjectDependency`（真实两项目编译运行 + `E7020` + `E7010`）|
| 顺手修的缺陷 | ① 非 `_`/`std` 导入根原先被**静默忽略**（只剩「未定义的名字」误导）→ 现在 `E7010`；② 诊断按「码 + 行 + 消息」去重（同一处被类型标注与字面量各查一遍会重复报）；③ 设计稿可见性阶梯与跨包导出矛盾 → 登记缺陷 14 并裁决 |
| 非目标 | 中心化注册表 / 版本求解 / 依赖改名强校验（见 rfcs/0004 §8） |

## 21. 可见性阶梯收口（按设计者澄清重做）

| 项 | 结论 |
|---|---|
| 澄清内容 | `public` = 可见性提一级；**块内**（`module {}` 里）默认**文件内**、`public` → **子包内**；**壳内**（`module` 外面）默认**子包内**、`public` → **包外**。轻量形态（文件无 `module` 块）的顶层声明字面上写在 module 外面 → 按壳语义 |
| 与上一版的差别 | 上一轮把「`public` 在任意位置都算跨包可见」当作缺陷 14 的裁决；本轮按澄清改成三级阶梯，并把缺陷 14 重写为「跨包导出没有落地通道」，裁决改为「壳内前向声明 + 块内实现」 |
| 语法新增 | 壳里允许**无体非 `#[c]` 的 `func` 声明**（前向声明）；轻量文件里仍报 `E2001`（没有实现处） |
| 语义检查 | `E7020` 按三级判定（`internal/sema/vis.go`）；`E7031` 声明未实现、`E3201` 声明与实现签名不一致（`internal/sema/link.go`）；签名只登记一次：可见性取声明、函数体取实现 |
| 落地改动 | `ast.Attrs.InShell`；parser 标记壳/轻量顶层；`types.StructInfo/EnumInfo` 与 `FuncSig/ConstInfo/GlobalInfo` 记录 `InShell`；`Table.Module` 供类型可见性判定；bindgen 生成 `#[c] public`（cinterop 靠它）；`stdlib` 对外函数 `public` |
| 判据 | `internal/sema/vis_test.go`：六个阶梯用例（块内默认跨文件不可见 / 块内 public 同子包可见 / 跨子包不可见 / 壳内 public 跨子包可见 / 壳内默认类型跨子包不可见 / 壳内 public 类型可见）+ 前向声明三例（正常 / `E7031` / `E3201`）；`TestCrossProjectDependency` 仍全绿 |
| 文档 | designs §2.1（壳允许前向声明）/§2.2（阶梯表 + 三级说明表 + 导出写法）；rfcs/0002 缺陷 14；rfcs/0004 §5；specs/0001 §1.2/§2.2；README |
| 实测 | 干净 PATH + 随包工具链下：六示例全绿；同子包 `public` 调用通过、跨子包报 `E7020`、壳内 `public` 跨子包真编译运行输出正确 |
