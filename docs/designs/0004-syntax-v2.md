# 设计 0004：语法简化 v2（层系统 · 类型模式 match · 可空视图 · select）

> 状态：**定案，实施中**（阶段状态见 §9）
> 关系：修订 designs/0001 的错误处理（§7）、可空类型（§3.7）、属性机制（§4.3）、并发（§10）中与本设计冲突的条目；0001 回写已在 P7 完成（头部 v2 修订注记 + 四处冲突节标记）。
> 来源：2026-10-27/28 设计共识（对话裁决），用户裁定「按此方案全部落地」。

## 0. 一句话

把语言收敛为二分语汇：**看到 `#` = 向编译器/运行时请求服务（层系统，开放集合）；看不到 `#` = 程序结构（封闭集合）**。错误、可空、并发等待三个内置场景从匿名判别和撤出，改由「Error 保留错误空间 + match 类型模式 + None 可空视图 + select」承载；六个内置变体保留名（Ok/Err/Some/None/Suspended/Done）全部退役。

## 1. 动机与设计法则

现有语法的机制数量失控：错误处理五套元素（`T!E` + `?` + match 解构 + Ok/Err 保留变体 + coro 三态）、类型聚合四套、编译期四个入口、变体命名三套规则。本设计以四条法则收敛：

1. **语汇二分**：`#层.API` = 服务调用（可枚举、可校验、可持续扩展）；裸形态 = 程序结构（数量封闭、永不再增长）。学语言 = 一张封闭结构清单 + 一张开放 API 表（与「学 Linux = shell 语法 + syscall 表」同构）。
2. **每个概念只允许一种拼法**：不设别名、不设省略形态、不设属性/调用双形态。
3. **层语义划界**：`#` 化的判据 = 是否对编译器/运行时的服务请求。控制流与类型骨架（match/region/yield/defer/if/for）无层语义，永不 `#` 化。
4. **显性最大化**：共管 API 的执行层由调用点显式选择，无默认省略形态（消灭 Zig 式「编译期能算就算」隐式规则）。

## 2. 层系统：`#compile` / `#comprun` / `#runtime`

### 2.1 单一形态

```aila
let n = #compile.sizeof(Dev);        // 纯编译期求值
let l = #comprun.len(buf);           // 编译期能折叠则折叠，否则运行时执行
let v = #runtime.as(I64, x);         // 强制运行时执行
let d = #runtime.checked_add(a, b);  // 运行时服务：溢出确定性 trap
```

- `#` 后只能是三个层名之一：`compile` / `comprun` / `runtime`（语言级保留命名空间，用户模块不得占用）。
- 层名后是 API 名：`#层.API(...)`。API 本体住在层命名空间里，是普通函数形态（延续「标准库函数必须是普通 AILA 函数」原则——编译器认识它们，但签名/调用皆函数形态）。
- **无默认裸形态**：共管能力的调用必须三选一，无第四种省略写法。

### 2.2 层校验规则

| 规则 | 说明 |
|---|---|
| `#compile.API` 只能在编译期上下文求值 | 出现在运行时上下文 = 编译错误 E9201 |
| `#runtime.API` 强制运行时执行 | 出现在常量/编译期求值上下文 = 编译错误 E9202 |
| `#comprun.API` 双态 | 编译期上下文强制折叠（折叠不了 = E9203）；运行时上下文自然执行 |
| 层内不存在的能力 = 拼写错误 | `#compile.alloc(...)` → E9204（该 API 不在 compile 层） |

### 2.3 声明位置指令（属性系统退役）

声明层标注与调用层调用共用同一形态——**指令也是 API**，`#[...]` 方括号属性语法退役：

```aila
#compile.c func strlen(s: Text) -> USize;       // 外部 C 符号绑定（原 #[c]）
#compile.inline func hot() {}                    // 内联指令（原 #[inline]）
#comprun.error enum Error { Nack, Timeout }      // 错误空间分片声明（§3）
```

### 2.4 v0 层 API 清单（开放表，持续扩展）

| 层 | API | 替代的旧语法 |
|---|---|---|
| compile | `sizeof(T)` | `sizeof` 内建 |
| compile | `as(T, x)` | `as` 关键字（纯重解释语义） |
| compile | `c` / `inline` / `packed` / `align(N)` | `#[c]` / `#[inline]` 等属性 |
| comprun | `len(v)` | 视图 `.len` / 内建 len |
| comprun | `error`（声明指令） | 无（新增） |
| runtime | `as(T, x)` | `as` 关键字（带检查/trap 语义） |
| runtime | `trap(msg)` | 隐式 trap 路径 |
| runtime | `alloc / free` | region 原语直接暴露形态待 P3 定 |
| runtime | `atomic_load8..cas64` | `std.atomics` `#[c]` 声明 + 转发样板（成片消失） |
| runtime | `chan(T, cap)` / `spawn(f, args)` / `coro(f, args)` | `chan@r` 创建 / `thread f();` / `coroutine@r f()` 启动表达式 |
| runtime | `checked_add / checked_sub / checked_mul` | 运算符内建溢出 trap（§7） |

> 算术语义拆分（重大裁决）：运算符 `+ - *` 回归回绕语义（C 语义，零检查、性能回 C 水平）；确定性溢出检查改为显式 `#runtime.checked_add/sub/mul`。「默认快 + 显式安全」。

## 3. 错误模型：Error 保留错误空间 + match 类型模式

### 3.1 错误类型声明（Zig error set 模型）

```aila
// Error 是保留 enum 名；任何模块可分片声明；成员名可跨分片重复出现
#comprun.error enum Error {
    Nack, Timeout,
}

// 模块自有错误集：#error 标注的普通 enum 名
#runtime.error enum I2cErr {
    BusBusy, ArbLost,
}
```

- `#error` 标注（层前缀区分构造层）告诉编译器：该 enum 参与 `!` 错误通道。
- Error 是保留类型名 = 全部错误集的**超集兜底**（跨模块统一错误空间）。
- `#comprun.error`：错误构造编译期可折叠则折叠（常量错误），否则运行时；`#runtime.error`：错误构造强制运行时。

### 3.2 可失败签名与传播

```aila
func read_reg(dev: Dev, addr: U8) -> U8 ! I2cErr { ... }   // E 仍可指定具体错误集（静态文档）
func read_temp(dev: Dev) -> I32 ! Error { ... }            // 或直接超集

let v = read_reg(dev, 0x3B)?;   // ? 上抛照旧：无变体名参与，抛给签名错误通道
```

- `T ! E` 拼法保留；E 必须是 `#error` 标注的 enum 类型；`Error` 是合法 E。
- `Error ! Error`（载荷也是错误类型）= 编译错误（保证类型模式可分派）。
- niche 布局照旧：`T ! E` 运行时表示 = 单 tag 判别和 `{ T 载荷 | E tag }`，零额外成本。

### 3.3 match 类型模式

```aila
match read_reg(dev, 0x3B) {
    let v: U8        => { use(v); },      // 成功分支：绑定载荷，类型静态已知
    let e: I2cErr    => { recover(e); },  // 错误分支：绑定具体错误集
    let e: Error     => { log(e); },      // 超集兜底分支
}
```

- 类型模式 `let 名: 类型 =>`：按类型分派，绑定新变量。
- 分派顺序即声明顺序，先匹配先得（`U8` 与 `Error` 天然互斥，无需顺序规则；同族重叠类型按声明序，重叠分支其后必死代码 = 编译警告）。
- 错误值的具体成员判别**复用变体模式**（不新增机制）：

```aila
match e {
    .Nack     => { retry(); },
    .BusBusy  => { backoff(); },
    _         => { fatal(); },
}
```

### 3.4 退役对照

| 旧 | 新 |
|---|---|
| `Ok { v: x }` 构造 | `return x;`（返回位置自动装载成功态） |
| `Err { e: ... }` 构造 | `return .Nack;`（返回位置自动装载错误态，错误成员直接构造） |
| `match r { Ok { v } => .., Err { e } => .. }` | `match r { let v: T => .., let e: E => .. }` |
| Ok/Err/Suspended/Done/Some/None 保留变体名 | 全部退役（None 降为字面量，见 §5） |

## 4. match v2：类型感知的分派器

### 4.1 模式清单（封闭，四族）

| 模式族 | 示例 | 适用判别对象 |
|---|---|---|
| 变体模式 | `.Data { v }` / `.Eof` / `.Stop` / `.Move { x, _ }` | 用户 enum、`#error` enum（成员判别） |
| 类型模式 | `let v: U8` / `let e: Error` | `!` 结果、可空视图（非 None 分支） |
| 字面量模式 | `None` | 可空视图 |
| 通配 | `_` | 一切（仅穷尽性兜底，不绑定不解构） |

- 变体模式一律 `.名`（leading-dot，在判别对象类型的变体集合里静态解析）；无载荷变体裸写，带载荷 `{ }` 解构，`_` 占位忽略字段。
- 用户 enum 与错误 enum 的变体模式完全同权，零特例。

### 4.2 穷尽性规则（三条）

1. **enum 对象**：全部变体覆盖，或以 `_` 收尾（编译期检查）。
2. **`!` 结果**：类型模式须使成功态与错误态皆有归属（至少一个 `let _: E`/`let e: E` 或 `_` 收尾）。
3. **可空视图**：`None` 分支与非空分支二分覆盖。

### 4.3 match 语句形态（v0 边界）

match 是**语句**，不做产出值的表达式形态；不引入 if-let/while-let（match 是唯一消费语法）；整数/字面量值分派留给将来的独立 `switch` 关键字；守卫、嵌套解构、裸绑定糖均不做（等真实痛点）。

## 5. 可空视图：`T?` 收窄 + None 字面量

```aila
let e = table.find(k);          // e: view Entry?（可空视图）
match e {
    let v: view Entry => { use(v); },
    None              => { miss(); },
}
```

- `T?` 收窄为**可空视图**专用：视图 `{ptr, byte_len}` 的 `ptr == 0` 即 None——零 tag 成本，「无」对视图是自然概念。
- **值类型可空（`I32?`）退役**：值可选场景用用户 enum 表达。
- 解引用安全：match 二分穷尽编译期保证（None 分支已处理则非空分支静态可知）；match 之外对可能为 None 的视图解引用 = 确定性 trap（运行时兜底，"null view dereference"）。
- Some 保留名退役；None 降为视图字面量（非变体名）。

## 6. select：并发等待的统一结构

```aila
select {
    case <-recv          => { sum = sum + v; },        // recv 就绪（有数据）
    case r = resume hs   => { total = total + r; },    // hs 可 resume（挂起过）
    case w = #runtime.spawn(f, x) => { ... },          // 线程完成（join 事件）
}
```

- select 是结构语法（C 档，不 `#` 化）；case 左侧是**就绪事件**（chan 可收 / coro 可 resume / 线程已结束），右侧是处理块。
- **展开语义（v0，无调度器）**：静态展开为声明序 try 链——依次推进各事件（chan try_recv / coro 推进一次 / 线程 join 探测），首个就绪者执行其块；全不就绪时协程内 = yield 挂起、resume 后从头重试，非协程（main）= 重走 try 链（轮转驱动）。
- 事件源互斥约束：同一 case 集内不可出现同一事件源两次（编译错误）。
- **coro 事件是「等完成」**（P10 定案）：探测即推进一次协程，**完成**（state 1/2）才算就绪，推进后仍挂起（state 0）按不就绪处理。于是绑定值只有完成态两态——`-> T ! E` 绑 `T ! E`、`-> T` 绑 `T`、无返回值不绑定；`Suspended`/`Done` 两个内建模式与 `resume h` 的独立表达式形态一并退役（E5010），单路等待 = 单 case select。

## 7. 迁移对照总表（全量）

| 旧语法 | 新语法 | 涉及层 |
|---|---|---|
| `#[c] func f(...);` | `#compile.c func f(...);` | compile |
| `#[inline]` | `#compile.inline` | compile |
| `as T` / `T(expr)` | `#compile.as(T, x)` / `#runtime.as(T, x)` | compile/runtime |
| `sizeof(T)` | `#compile.sizeof(T)` | compile |
| `视图.len` / len 内建 | `#comprun.len(v)` | comprun |
| `T?`（值类型可空） | 退役（用户 enum 替代） | — |
| `T?`（视图可空） | 可空视图 + `None` 字面量 | 结构 |
| `Some { v }` / `None` 变体 | 类型模式 / `None` 字面量 | 结构 |
| `Ok { v: x }` / `Err { e }` | 返回位置自动装载 / 错误成员构造 | 结构 |
| `match { Ok { v } => .., Err { e } => .. }` | `match { let v: T => .., let e: E => .. }` | 结构 |
| `Suspended` / `Done` 变体 | select 就绪事件 | 结构 |
| `thread f(args);` | `#runtime.spawn(f, args)` | runtime |
| `coroutine@r f(x)@[r2]` 启动 | `#runtime.coro(f, x)` | runtime |
| `chan@r T(n)` 创建 | `#runtime.chan(T, n)` | runtime |
| `std.atomics` `#[c]` + 转发样板 | `#runtime.atomic_load8..cas64` | runtime |
| 运算符溢出 trap | 回绕语义 + `#runtime.checked_add/sub/mul` | runtime |
| `view sh: [Unsafe] = block` 重标 | 层 API 形态（P3 定名） | runtime |
| `U8 ! E` 匿名判别和 Ok/Err 变体 | `U8 ! E` + 类型模式（`!` 拼法保留） | 结构 |

## 8. 明确不做（v0 边界）

- match 表达式化、if-let、守卫、嵌套解构、裸绑定糖、整数 switch（独立关键字留扩展点）
- 效果系统 / 非值效果通道（远期议题，前置 = 泛型）
- prelude 机制（层系统 + leading-dot 后不再需要）
- `#` 层用户自定义命名空间（compile/comprun/runtime 三层封闭，扩展 = 加 API 项）

## 9. 实施阶段与状态

| 阶段 | 内容 | 状态 |
|---|---|---|
| P0 | 本设计文档定稿 | ✅ 完成 |
| P1 | lexer/parser：`#层` token、None 字面量、变体保留名退役（Ok/Err/Some/None/Suspended/Done 出保留表）、`#[...]` 属性语法下线 | ✅ 完成 |
| P2 | 错误模型：`#error` enum 声明、`T ! E` 重定义（E 限错误集）、返回位置自动装载、`?` 重接、niche 布局保持 | ✅ 完成 |
| P3 | match v2：类型模式 + 新穷尽性 + 变体模式 leading-dot 全面切换 | ✅ 完成 |
| P4 | 可空视图：`T?` 收窄、None 检查 lowering、trap 兜底 | ✅ 完成 |
| P5 | select：事件探测 + try 链展开 + 挂起重试 | ✅ 完成 |
| P6 | 层系统 codegen：`#compile/#comprun/#runtime` 全 API、运算符回绕 + checked 族、thread/coro/chan 启动 API 化、原子样板消除 | ✅ 完成 |
| P7 | 全量迁移：std、examples、快照、三 demo、diag_codes、language.md / SKILL.md / designs 0001 回写 | ✅ 完成 |
| P8 | 验证收口：全量 go test + 四 demo + esp8266 冒烟 + 快照重生成 | ✅ 完成 |

### P1 实施备注（2026-10-28）

- **并行引入策略**（偏离原文「保留名退役/属性下线」的时间点）：v2 语法形态全部进入 lexer/parser 并落 AST，旧机制（`Ok/Err/...` 保留名、`#[...]` 属性、`#size/#align` 旧查询）**暂不删除**——它们仍服务现有 std/examples，P7 全量迁移时统一下线。理由：保留名一动下游全断，中间态无法保持全绿。
- 声明位层指令 `#compile.c` 等落为 `Attr{Name, ViaLayer:true}`，**复用现有属性语义管道**（下游 sema/codegen 零改动即生效）；`#comprun.error` → `EnumDecl.ErrLayer`。
- 表达式位 `#层.API(...)` → 新 `ast.LayerCall` 节点（类型参数以 `*TypeRef` 进 Args：`sizeof/alignof` 单类型参、`as` 类型+值两参）；API 名按 token 文本读取（`len/as/trap` 等 P1 阶段还是关键字）。
- match 新模式：`.名` leading-dot 变体模式（复用 `VariantPattern/PathPattern`）、`let 名: 类型` 类型模式（新 `ast.TypePattern`）；`None` 字面量模式节点（`ast.NonePattern`）已定义，sema 接入延至 P4。
- 快照重生成 93 份（`Attr` 新增 `via_layer` 字段导致 AST 契约漂移，预期内）；新增 `TestSyntaxV2Parsing` / `TestSyntaxV2BadLayer` 锁定 v2 解析形态。全量 go test 18 包全绿。

### P2 实施备注（2026-10-28）

- **错误值全局统一编号（Zig 语义，实施中确立）**：`#error` enum（含模块自有错误集）在声明时成员 tag 直接改写为**全局错误空间编号**（`types.RegisterErrorShard`），同名成员跨分片 = 同一成员同一编号。收益：任意具体错误集上抛进 `Error` 超集**零 tag 重映射**，codegen 只做宽度对齐（trunc/zext）与聚合重装；带载荷成员按 tag 线性分发拷贝载荷（`coerceErr`）。
- **ErrorSpace 超集类型**：`types.ErrorSpace{Info}` 进程级单例，`Info` 指向全局聚合 `*EnumInfo`（成员集合 = 各分片并集，tag 宽度按需重算）。类型位 `Error` 解析规则：表中存在 `#error enum Error` 分片或不存在 → 全部解析为 `ErrorSpace`；**旧式未标注 `enum Error` 保持原 Named 语义**（P7 下线时一并收紧）。
- **P2 宽容期**（并行引入策略延续）：`T ! E` 的 E 暂允许未标注普通 enum；leading-dot 构造对普通 enum 亦放行。P7 全量迁移后收紧为仅 `#error` 标注。
- **`return .成员;`**：sema `checkReturn` 特判（LeadingDot + `sig.Err != nil`）→ 按 `sig.Err` 解析成员 + `ErrReturns[r]` 标记；codegen `lowerReturn` 查标记走 `buildResult(ok=false)`。无载荷 enum 的 LLVM 标量表示（llty = tag 类型）直接以 tag 常量为值——顺带修复了旧管道对无载荷 enum 变体构造生成非法 `insertvalue` 的隐患。
- **`?` 上抛超集规则**：`sig.Err` 是 `ErrorSpace` 时接收任意错误集上抛（超集通道）；反向（超集 → 具体集）仍拒绝。`TryRef` 新增 `Err` 字段记录操作数原始错误类型，codegen 据此决定是否生成 `coerceErr`。
- 冒烟示例 `examples/p2smoke`：分片声明（含载荷成员）、`return .Nack;`、`return .NoSpace { need: 8 };`、带标注绑定 `let e: Error = .Nack;`、跨集上抛（无载荷 + 带载荷）全链路运行通过；全量 go test 18 包全绿。

### P3 实施备注（2026-10-28）

- **类型模式分派键复用旧形态**：`let v: U8` / `let e: Error` 在 sema 侧解析为 `ArmRef{Builtin: "Ok"/"Err"}`（optional 的非空类型模式 → `"Some"`），codegen 的 result/optional 分派管道**零改动**。绑定类型取**通道载荷类型**（槽类型）：`let e: Error` 在 `U8 ! I2cErr` 上绑定的 e 是 `I2cErr` 值（模式标注 `Error` 只作「覆盖整个错误通道」的归属约束，不做超集装箱——真实痛点出现再加 coerceErr 级联）。
- **错误模式整态覆盖规则**（v0 收窄）：错误侧类型模式须为 `Error` 超集或与通道类型一致；**子集模式**（`U8 ! Error` 上 `let e: I2cErr`）需要运行时成员判定 + 不命中回退，v0 不做（编译报 E3201）。
- **ErrorSpace 作判别对象**：head 类型分派新增 `types.ErrorSpace → kind="enum"`（enum = 全局聚合 Info），成员判别 `.Nack =>` / `.Timeout =>` / `_` 直接走既有 enum 管道（codegen `enumOf` 同步认识 ErrorSpace）；带载荷成员的解构绑定（`.NoSpace { need }`）照常工作。
- **穷尽性**：复用 `seen` 集合（类型模式与旧变体模式都写 `Ok/Err/Some` 键），`!` 结果的「成功态与错误态皆有归属」与可空视图的「None / 非空二分」检查逻辑不变；corostate（resume 三态）暂保留旧模式，P5 select 化时迁移。
- 同族重叠类型模式的「其后必死代码 = 编译警告」v0 未做（diag 无警告通道），记录为 P8 后 polish。
- 冒烟扩展 `examples/v2smoke`（原 p2smoke 更名）：类型模式两分支、错误空间成员判别（`.Nack/.Timeout/_`）、跨集成员 `.BusBusy` 判别全链路运行通过；全量 go test 18 包全绿。

### P4 实施备注（2026-10-28）

- **None 模式接入**：parser `parsePattern` 新增 `None =>` → `ast.NonePattern`（早于 KindIdent 分派）；sema `checkMatch` 的 NonePattern case 仅在 optional 通道合法（否则 E3201「None 模式只用于可空视图 T?」），穷尽性写 `seen["None"]`，codegen 复用 `Builtin: "None"` 分派键零改动。AST 快照随 NonePattern 节点漂移已重生成。
- **`return 值;` 自动装载（`T?` 签名）**：sema `checkReturn` 特判——值不是「整个可空值字面量」时按 `opt.Inner` 检查（`checkEscape` 照走）；codegen `lowerReturn` 补包装：niche 可空视图表示不变零包装，tag 形态按 `Some` 模板 `insertvalue i8 1, 0` + Inner 槽 1。判定辅助 `optionalWholeLit`（Path `None` / 旧式 `Some {...}` StructLit）走常规整体检查。
- **宽容期兼容**：旧式 `return Some { v: e };`（std/os.aila 的 env_set 等存量写法）经 `optionalWholeLit` 排除自动装载、按整体类型检查照常编译；P7 全量迁移统一为 `return e;` 后移除该豁免。
- **`return None;` 判定时序**：不能依赖 `PathBuiltin`（检查前未填），改按语法判断 `p.Dotted() == "None"`——fallthrough 常规路径按整体类型检查，codegen 由 `PathBuiltin == "None"` → zero 值既有管道承接。
- **v2smoke 可空用例**：`name_of -> Text?` 的 `return None;` / `return "ok";` 自动装载 + match（`let s: Text =>` / `None =>`）两分支判别全链路运行通过（输出 none hit / ok）；全量 go test 18 包全绿。

### P5 实施备注（2026-10-28）

- **事件白名单 v0**：chan recv 端（`<-ch`）与 coro 句柄（`resume h`）两类；设计例第三类「线程已结束」（`case w = #runtime.spawn(f, x)`）依赖 P6 的 spawn API 返回句柄 + join 探测，随 P6 落地后接入 select。
- **coro 事件就绪条件**：`state != 3`（0=挂起可恢复执行；1/2=完成未观察，可观察产出；3=已消费）。设计注释「挂起过」放宽为「可安全 resume」——完成未观察态若不算就绪，「等结果」用例会死等（已完成协程无人 resume，state 停在 1/2）。就绪后动作 = 完整 resume 语义（`lowerResume` 重构出共享主体 `coroResume`，select 与独立 resume 表达式同一管道）。
- **case 绑定**：`case 名 = 事件 => 块`；chan 事件省绑定名时隐式绑定 `v`（对齐设计用例 `case <-recv => { sum = sum + v; }`）；`_ =` 形态丢弃产出。coro 事件绑定 r = resume 完整产出（三态判别和），与独立 resume 表达式同型——用户 match 消费。
- **展开语义**：声明序 try 链——逐 case 非阻塞探测（chan = `aila_rt_chan_try_recv`；coro = `load i8 state` + `icmp ne 3`），首个就绪者 br 入其块；全不就绪 `coroSuspendRetry` 挂起，恢复后重走整条链。仅协程内合法（E5001）；同一事件源（句柄符号身份）重复 = 新码 **E5008**（实施中由 E5006 改号，E5006 已被「协程启动目标不合法」占用）。新关键字 `select`/`case` 入 lexer Keywords。
- **corostate 类型模式迁移**（P3 遗留承诺）：TypePattern 加 `corostate` 分支——完成态按 §4.2 规则 2 的 `!` 同族处理（`let v: T` / `let e: E` 各至多一个，错误侧 Error 超集恒覆盖）；Suspended/Done 保留内建模式拼法（地位同 None，从变体退役但模式可用）。穷尽性按 `seen[Suspended/Ok|Done/Err]` 天然接上，concurrency_demo 旧形态不受影响（P7 统一迁移）。
- **v2smoke select 用例**：`driver` 协程内 select（chan 事件 + coro 事件）由 main `match resume hd` 驱动，嵌套 `feeder` 协程 yield 两轮后发送 7、完成值 5；输出 `select: done 5` / `select: recv 7` / `select driver ok`，同时覆盖 corostate 类型模式（Suspended / `let r: I32` / `let e: Error`）。「全不就绪挂起重试」路径未在冒烟覆盖（单驱动场景会死锁），IR 结构与 `lowerRecv` 既有挂起重试管道等价。
- 语法事实澄清：`yield;` 无值、coro 启动 = `let h = coroutine@r f(args);`（eager 跑到首个挂起点）——冒烟按既有形态；AST 快照随 Select/SelectCase 节点漂移已重生成。

### P6 实施备注（2026-10-29）

- **层校验 v0 = E9204 单码**：未知 API / 参数个数不符 / 类型参数形态错全走 E9204（登记进 diag_codes.json 唯一事实源——`Bag.emit` 对未登记码直接 panic，契约生效）。E9201（编译期不可用）/E9202（运行时 API 进常量位）依赖常量求值器扩展，随 v0 后落地。顺手补登了 P2-P5 已用未登的 E5006/E5007/E5008（panic 第一现场由 E9204 暴露）。
- **sema 静态解释**：`LayerRef{Kind, Sig, Elem, Src, At, AtomicFn, CheckedOp}` 挂 `Checker.LayerRefs[x]`，codegen 按 Kind 分派零猜测。API 名切片坑：`x.Name` 只含 API 段（`atomic_fetch_add32`），后缀解析用 `len("atomic_")` / `len("checked_")` 切（`len("runtime.atomic_")` 会越界——实施中两处踩过）。
- **运算符回绕**：`+ - *` 直接 `add/sub/mul`（无符号有符号同指令），溢出检查移除；`/ %`（除零 + INT_MIN/-1）与移位检查不变。快照随回绕漂移重生成（"integer overflow" 字符串消失，预期内）。
- **checked 族**：`#runtime.checked_{add,sub,mul}(a, b)` 产出 `(T, Bool)` 多值组（`T?` 对值类型已退役，不再用 niche 装载）；LLVM `llvm.{s|u}{add|sub|mul}.with.overflow` 探测后 `insertvalue` 组装，溢出标志不 trap、调用侧自决。纯字面量操作数在 sema 落型 I32（codegen 需要具体宽度）。
- **启动 API 化**：`threadStartInto`（ThreadStart 与 `#runtime.spawn` 共用）、`coroCreateInto`（CoroCreate 与 `#runtime.coro` 共用）抽共享主体；`#runtime.spawn` 要求目标 `thread func`（入口白名单在声明处已查），`#runtime.coro` 拒绝带区域参数目标（层 API 无 `@区域` 实参通道）。`#runtime.spawn` 产出 Void（fire-and-forget 语句位）——**设计例的 select「线程已结束」事件（spawn 返回句柄 + join 探测）v0 未接**，需要 spawn 句柄化 + join 原语，记 P8 后设计细化。
- **区域推断**：`#runtime.chan/coro` 无 `@r` 注解，codegen 取词法最内层 `region`（`funcGen.currentRegion`，lowerRegion 进出 push/pop）；无区域上下文兜底 null（与显式形态一致）。parser 为 `#runtime.chan` 补类型参数解析 case（默认 parseExpr 会把 `I32` 解析成 Path 而非 TypeRef）。
- **v2smoke P6 用例**（输出 4 / 4 / true / 1 / 16 / 3 / 200 / 5 / 100000）：U8 回绕加法不 trap、checked_add 溢出探测双产出、sizeof（struct / `[I32; 4]`）、`#comprun.len`、`#compile.as`（U8→I32 zext）、`#runtime.chan+coro` 区域推断协程驱动、`#runtime.spawn+atomic_load/fetch_add/store32` 计数自旋收口。
- **并行引入状态**：`#[...]` 属性与 `#compile.c` 声明指令并存（P7 收口）、旧关键字 `as/len/trap` 与 `#size/#align` 仍在（std/examples 依赖）、`view sh: [Unsafe]` 重标的层 API 形态（§7 迁移表「P3 定名」）待定名。全量 go test 18 包全绿；快照 93 份已重生成。

### P7 实施备注（2026-10-30）

- **错误空间编号基线（实施中确立）**：全局错误空间编号随注册顺序累积，逐模块批处理（goldengen / go test 顺序 Load 多工程）与单工程编译（CLI）编号漂移 → `types.ResetErrorSpace()` 挂 `build.Load` 入口（每次装载 = 一次编译基线），goldengen 复用同一入口语义。快照与 CLI 输出从此同基线。
- **协程错误返回修复**：`lowerCoroReturn` 只有旧 `return _ ! E;` 分支，`return .成员;`（ErrReturns 标记）落到成功态 → 协程完成值 = tag 当载荷（`flaky ok v=0`）。补 ErrReturns 分支走 `coroFinish(2, …)`。
- **std.atomics 成片消失（§7 迁移表兑现）**：`stdlib/atomics.aila` 删除，绑定样板与调用方（concurrency_demo / atomics-demo / v2smoke）全部直调 `#runtime.atomic_*`。
- **`thread: 栈` 语义退役**：`#runtime.spawn` 层 API 无栈深参数（§2.4 定案），`thread f();` / `thread: N f();` / E5005（嵌入式栈必填）一并退役；嵌入式栈深控制留待 spawn 增补可选参数时再议。
- **`import … as` 别名退役**：全仓零使用 + `as` 出保留字表，绑定名恒为路径末段（sema / lsp 同步）。
- **env_set 改造（值类型可空退役的落点）**：`std.os.env_set` 由 `-> core.IoErr?`（Some/None）改为 `-> USize ! core.IoErr`；os_demo / os-tour 同步类型模式。
- **match 表达式保留**：corpus 中 `return match …` / `let x: I32 = match …` 不在 §7 迁移表，保留现状（§4.3 v0 边界的实现超前形态，收口议题记 P8 后）。
- **编译器下线清单**：lexer 保留字 as/trunc/len/trap、builtinVariants 的 Some/None/Ok/Err（Suspended/Done 按 P5 定案保留为内建模式）；parser `#[...]` 属性分支、parseConversion、`len` 前缀、trap/thread 语句、`#size/#align` 查询、`chan@`/`coroutine@` 启动、import 别名；AST 节点 Cast/Len/SizeOf/AlignOf/Trap/ThreadStart/ChanCreate/CoroCreate/TryDetect；sema 对应 case + LenRefs/CastTypes/ThreadStarts 注记 + `optionalWholeLit`（Some 豁免）+ E5005；codegen 对应 lower 分支。`#` 表达式位现在是纯层调用（parseLayerCall 内校验层名）。
- **sema 收紧（P2 宽容期结束）**：`T ! E` 的 E 必须是 `#error` 标注错误集（E3201）；类型位 `Error` 恒解析全局超集（旧式未标注 `enum Error` 不再特判）。
- **周边同步**：bindgen 模板 `#[c]` → `#compile.c`（快照 mathkit.ail 重生成）；corpus 的 layout/types Color 补 `#runtime.error`；测试内联源与 lsp 工程样例全部 v2 化；`#compile.alignof` 补齐（sema/codegen）。
- **demo 迁移**：工作区 8 个 demo（atomics/thread/chan/coro-demo/coro-mini/coro-smoke/os-tour/smoke）全部 v2 化并运行验证；`std.atomics` 导入者改层 API。
- 快照重生成 92 份（std_atomics.ll 随 atomics.aila 删除）；全量 go test 18 包全绿；v2smoke 17 行输出不变。

### P8 实施备注（2026-10-30）

- 验证面全绿：go test 18 包（含 IR parity / 端到端逐字节比对 / 跨项目依赖 / bindgen / lsp）；10 个示例工程全部重建成功；concurrency_demo（协程三态 / chan 双挂起路径 / 双线程原子竞争 / cas 抢占 / u8 宽度族）、os_demo（args/env/file/time/rand/path 错误通道）、v2smoke 17 行输出逐行正确；工作区 8 个 demo 全部构建运行通过（coro-demo 错误协程、atomics-demo 200000 精确计数、thread-demo、chan-demo、coro-mini、os-tour、smoke）。
- esp8266 冒烟：coro-smoke 交叉编译 `xtensa-esp8266-elf` 出 `app.bin`（llvm.coro 后端路由正常）。
- 文档同步：designs/0001 头部 v2 修订注记 + §3.7/§4.3/第 8 章/第 10 章四处冲突节标记；skill 文档（SKILL.md / language.md / recipes.md / errors.md / cli.md）v2 化并同步 `.local` 与 `.trae` 副本；diag_codes.json 的 E2010/E3201/E3202/E5005/E9204/W3203 文案随退役更新（E5005/W3203 保留登记但标注已退役）。
- **设计状态**：0004 全部阶段（P0-P8）完成，v2 成为唯一语法事实。挂账项：match 表达式化 / switch 关键字（§4.3 v0 边界）、spawn 句柄化 + join 事件（§6 select 线程事件）、`#runtime.as` 检查语义、嵌入式 spawn 栈深参数。

### P9 实施备注（残留下线收口）

P7 的下线清单只兑现了 lexer 与 `optionalWholeLit` 一侧：`Some` / `Ok` / `Err` 的**构造与变体模式**在 sema / codegen 里仍是活路（`let a: Text? = Some { v: "x" };`、`match r { Ok { v } => .. }` 都能编译并出 IR）。P9 把这条尾巴切干净：

- **sema**：删 `checkBuiltinLit`（`Some`/`Ok`/`Err` 构造管道）与其分派；删 `checkMatch` 里 optional/result 的旧变体模式两支（`Some { v }` / `Ok { v }` / `Err { e }` 及裸名 `Some`/`Ok`/`Err`），改为 `E3001` + 迁移提示。`None` 字面量、`Suspended`/`Done` 内建模式不动（§5 / §6 明文保留）。
- **codegen**：删 `lowerStructLit` 的内置变体支（`LitRef.Builtin`/`Payload` 随之删除）；`ArmRef.Builtin` 保留——它是类型模式 / `None` 模式 / resume 三态的**内部派发键**，不是用户可写的名字。
- **lexer**：删 `builtinVariants` 与它那个 case——该 case 产出的 token 与 `default` 分支逐字段相同，纯空转死码。
- **名字空间**：三个名字回到普通标识符，用户可自用（`enum Mine { Ok, Some }`、`func Err()` 均实测通过，无撞保留名）。
- **测试**：新增 `internal/sema/legacy_variants_test.go`——五个旧拼法必须报 `E3001`，v2 拼法必须原样通过（原来没有任何测试覆盖过这些旧路径，所以 P7 漏删没被发现）。
- **顺带修的**：`TestLayoutMatchesReference` 未重置全局错误空间，`-shuffle=on` 下 tag 漂移（layout.aila 的 Color 是 `#runtime.error`）。
- **文档**：specs/0002 §2 token 表（内置变体名 → 已退役）、designs/0001 §1.4 / §8.2 / §10.5 三处表格、VS Code 片段（`Some { v }` → 类型模式、`#[c]` → `#compile.c`）。

**Suspended / Done 当时为何没跟着删**：`select` 那时只替代「等待」——就绪条件 `state != 3` 允许进入分支后 resume 出 `Suspended`（协程又让出），所以 case 体仍要靠 `Suspended` 分支收口；而 `resume h` 独立表达式形态（§6 原文保留）产出的就是三态判别和。彻底删掉这两个名字需要两步**语义**改动（见 P10）。

### P10 实施备注（select 等完成：Suspended / Done 彻底退役）

P9 把内置变体名收到只剩 `Suspended` / `Done`（§10.5 的三态观测）。P10 按 §6 的原意把这两个也拿掉——办法是把「等待」这件事整体搬进 select：

- **select 的 coro 事件改成「等完成」**（`internal/codegen/stmt.go` 的 `lowerSelect` + 新 `storeCoroDone`）：探测即推进一次协程（`coroResume`），**完成**（state 1|2）才进 arm，推进后仍挂起（state 0）就按「不就绪」处理、继续试下一个事件。绑定值从三态收窄成完成态两态：`-> T ! E` → `Result{Ok,Err}`（用户看到 `T ! E`）、`-> T` → `T`、无返回值 → 不绑定（绑名字报 E3201）。三态聚合没消失，但只作为 `types.CoroState` 在 codegen 内部描述控制块布局，不再出现在用户类型里。
- **`resume h` 的表达式形态退役**：sema `checkExpr` 的 `*ast.Resume` 分支改报 **E5010**（新登记进 `diag_codes.json`），codegen 删 `lowerResume`；`resume` 关键字只剩 select case 语法位。
- **select 在非协程里也合法**：E5001 守卫撤销（`<-ch` / `yield` 的那条保留）。全不就绪时协程内仍 `coro.suspend` 挂起重试，main 里改为「重走 try 链」——这正是过去 main 手写 `for !d { match resume h { ... } }` 轮转循环的等价物，语义与输出都保持。没有这条，main 就再没有驱动协程的手段（协程句柄只能靠 select 观测与推进）。
- **match 侧清空**：`checkMatch` 的 corostate 分支（类型模式 / 变体模式 / 裸名 / 穷尽性）与 codegen 的两处 corostate 派发全删；`Suspended` / `Done` 回到普通标识符（用户可自用）。对三态判别和直接 match 现在报 E3201。
- **迁移**：`examples/concurrency_demo`（协程段 + chan 双端轮转）、`examples/v2smoke`（driver 内 select、main 驱动、P6 chan/coro 段）全部改成 select；`go run ./tools/goldengen` 重生 92 份快照（只有 concurrency_demo 的 main.ll 结构漂移，预期内）。
- **为什么必须动 codegen 而不是只删名字**：`Suspended` 不是死码——它是「本次 resume 没能跑完」的唯一可观测信号，用户正是靠它写轮转调度。把它收进 select 内部（就当不就绪处理）之后，用户侧不再需要这个信号，名字才能真删。


> 每完成一个阶段更新本表状态，并在该阶段条目下追加「实施备注」小节记录偏离决策。
