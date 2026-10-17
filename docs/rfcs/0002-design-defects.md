> **实现落点说明（G4 收尾后）**：本文记录缺陷与裁决；「实现落点」里的路径已随 Go 重写改到
> `internal/*`（Python 参考实现已在 G4 删除）。历史条目里若仍出现 `src/aila/*.py` 或 Python 测试名，
> 一律按下表对应到 Go 包：
>
> | 原 Python | 现 Go |
> |---|---|
> | `src/aila/lexer.py` / `parser.py` / `ast.py` | `internal/lexer` / `internal/parser` / `internal/ast` |
> | `src/aila/types.py` / `sema.py` | `internal/types` / `internal/sema` |
> | `src/aila/codegen.py` / `backend.py` | `internal/codegen` / `internal/toolchain` |
> | `src/aila/bindgen.py` / `build.py` / `project.py` | `internal/bindgen` / `internal/build` / `internal/project` |
> | `src/aila/{home,resolve,toolchain,manifest,cli,diag}.py` | `internal/{home,resolve,toolchain,manifest,cli,diag}` |
> | `tests/test_*.py` | 对应包的 `*_test.go` |

> **编号**：0002 · **类型**：RFC（缺陷登记 + 裁决） · **状态**：已裁决，待设计稿修订
> **上游**：[designs/0001](../designs/0001-language-spec.md)
> **原则**：`designs/0001` 是唯一权威。本文件只登记「两处规则互相矛盾」的地方，并给出**按权威优先级**的裁决与实现落点；**不擅自改写设计稿**——修订设计稿是设计批次的事，本文件是它的输入。

## 缺陷 1：`let` 绑定的视图能否写入（§3.3 例 vs §4.2 规则 3）

| 位置 | 原文 | 读出来的规则 |
|---|---|---|
| §3.3「下标纪律」示例 | `let a: [U8] = [0; 16]@work;` 后写 `a[i] = 2;` 并标注 **✓** | 只读绑定也允许写元素 |
| §4.2 规则 3 | 「`let` = **深不可变**（绑定与值整体，字段不可赋值），要改用 `var`」 | 只读绑定禁止写 |
| §5.3「权限」 | 「`let` = 读权视图；`var` / `@mut[x]` = 读写权；**可变性属于绑定**」 | 只读绑定禁止写 |

**裁决**：以 §4.2 规则 3 + §5.3 为准——`let` 绑定的视图**不可写**，写元素必须 `var`。

**理由**：①两票对一票，且规则条（规范性条款）优先于示例；②若放开，则「`let` = 读权视图」这条权限模型直接失效，`@mut` 的「供给需持有」也失去锚点；③§3.3 该行示例旁边的注释是「下标必须 USize」，它的**考点是下标类型**，写权只是顺带，属笔误。

**实现落点**：`src/aila/sema.py` 的 `check_place(..., writing=True)` → 只读绑定写入报 `E4201`；`tests/test_sema.py::test_let_not_writable` 锁定。

**对设计稿的建议改法**：把 §3.3 示例改成 `var a: [U8] = [0; 16]@work;`（一行修改，考点不变）。

## 缺陷 2：`@` 子句族的位置（§5.2 vs §13.2）

| 位置 | 原文 | 读出来的规则 |
|---|---|---|
| §5.2「签名三段式」 | 「@ 段位置：**参数列表闭括号之后、`->` 之前**」 | `func f(...) @region[r] -> T` |
| §13.2 API 面表 | `func pool.Pool.take(self) -> [U8]? @mut[self]` | 效果写在返回类型**之后** |

**裁决**：§5.2 为规范位置（写在 `->` 之前）；但实现**两种顺序都接受**，以便 `§13.2` 这类写法不改也能编。

**理由**：语法宽容不产生语义歧义（`@` 在类型后出现时，唯一解释只能是效果子句）；硬拒会让设计稿自己的 std 示例编不过，属于「文档自我打脸」而不是「纪律」。

**实现落点**：`src/aila/parser.py::parse_func` 在 `->` 前后各解析一次效果子句；诊断仍按 §5.2 的顺序规则（`@region → @mut → @move`）校验。

**对设计稿的建议改法**：统一 §13.2 表格为 `func pool.Pool.take(self) @mut[self] -> [U8]?`。

## 缺陷 3：`out` 既是关键字又当参数名（§10.6 示例 vs §1.4 词汇表）

| 位置 | 原文 | 读出来的规则 |
|---|---|---|
| §1.4 词汇归类 | `out` 属「流程」关键字（`return if else for in match region defer unsafe coroutine yield out trap`） | `out` 是保留字，不能作标识符 |
| §10.6 示例 | `coroutine func producer(out: send chan Cmd) { ... }` | 用 `out` 当参数名 |

**裁决**：以 §1.4 为准——`out` 是保留字，不能作参数名/绑定名；示例里的 `out` 是笔误（作者想表达「这是一个出口端」，属于示意性命名）。

**实现落点**：解析器按关键字处理，`producer(out: ...)` 报 `E2001`；黄金语料因此改用 `tx`（`tests/corpus/control.aila`）。

**对设计稿的建议改法**：§10.6 的 `producer(out:) / consumer(in:)` 改成 `producer(tx:) / consumer(rx:)`；同时补一句「`out`/`in` 都是保留字，通道端建议命名 `tx`/`rx`」。

## 非缺陷：口径确认（实现时按此固化，不必改设计稿）

| 议题 | 设计稿依据 | 实现口径 |
|---|---|---|
| 分组括号 `(T)` 与 `T` | §9.1 arity=1 括号消除 | parser 在 `(` 单元素时返回内层类型，不产生 group |
| `Text` 的 `len` | §3.3「Text len = 字节数」 | `len text` 产 `USize`，与 `[U8]` 字节数一致 |
| `T?` 的表示 | §3.7 规则 6 niche | 句柄类用 ptr=null 表示 None；enum/struct 暂用「present 标志位」布局（M2 优化为 niche） |
| `T ! E` 的表示 | §8.2「≡ Ok{v}/Err{e}」 | `{ i1 is_ok, T, E }` 直白布局；不跨 C ABI（§11.6 已定） |
| 跨模块调用的链接 | 铁律 7（不可内联） | M1 全部函数给外部链接以保证可解析（分派表在 M3）；`manifest.json` 与 `W9002` 明示 |

## 缺陷 4：无载荷 enum 的宽度没有跟随底层类型（§3.5 规则 5 vs 实现）

| 位置 | 原文 | 读出来的规则 |
|---|---|---|
| §3.5 规则 3 | 「底层类型默认 `U32`（对齐 C int）；`: U8` 标注收窄」 | 显式底层类型决定宽度 |
| §3.5 规则 5 | 「enum 宽度 = 底层类型宽度（`Reg` 占 1 字节）——第 3 条推论」 | 同上，且给了字节数 |
| 参考实现（Python） | `EnumInfo.tag_type` 默认 `U32`，只有**带载荷**的分支才会重算宽度 | `enum Reg: U8 { … }` 实际占 4 字节 |

**裁决**：以设计稿为准——**无载荷 enum 的宽度 = 底层类型宽度**；带载荷 enum 若显式写了底层类型，也以它作 tag 宽度（但必须容纳全部判别值，越界报 `E3201`）；未写则仍按「最小宽度」推定。

**实现落点**：Python 的 `Sema.resolve_type_bodies` 与 Go 的 `sema.buildEnum` 同步修正；黄金语料 `tests/corpus/layout.aila` 里的 `enum Narrow: U8`（1 字节）与 `enum Wide: U16 { A = 300 }`（2 字节）把这条口径钉住。

**影响面**：`import ch` 映射来的 `enum X: U8` 寄存器映射不再白占 3 字节；纯 AILA 侧此前按 `U32` 排布的 struct 会随之变小（属修正，不是破坏）。

**对设计稿的建议**：§3.5 规则 3 补一句「带载荷 enum 显式写底层类型时，宽度取该类型且必须容纳全部判别值」。

## 缺陷 5：`@mut[...]` 子句从来没被解析过（实现缺陷，非设计冲突）

| 事实 | 说明 |
|---|---|
| 现象 | `func f(dst: [U8]) @mut[dst] { }` 解析失败；只有 `@region` / `@move` 能用 |
| 根因① | `mut` 不在词法保留字里（`KEYWORDS` 只有 `region` / `move`），于是「下一个 token 是子句名」的判断对 `@mut` 恒为假，整个子句被静默跳过 |
| 根因② | 即便进了分支，子句类别取的是 token 的 `kind`（`mut` 是普通标识符 → `ident`），会把 `@mut` 记成 `ident` |

**裁决**：子句名按**文本**识别（`region`/`move` 是关键字，`mut` 是普通标识符都接受），子句类别取 `text`。不改词法保留字表——设计稿 §1.4 的关键字清单里没有 `mut`，把它升格成关键字会破坏已有标识符。

**影响面**：这是**能力级缺陷**：`@mut` 是「写权供给」的签名半边，缺了它，`@mut[self]` 方法、`@mut[dst]` 的调用点检查全都无从谈起（此前只是没有语料踩到）。

**实现落点**：Python `parser._is_effect_clause` + `kind = text`；Go `parser.isEffectClause` + `p.adv().Text`；语料 `tests/corpus/sigs.aila` 的 `Matrix.scale(self, k) @mut[self]` 与 `with_effects(dst, src) @region[..] @mut[dst]` 锁定。

## 缺陷 6：`same_type(None, None)` 返回 False（实现缺陷）

`FnPtr` / `Coro` 的 `err` 字段常态为空；`same_type` 一开始就写 `if a is None or b is None: return False`，于是「两侧都没有错误通道」被判成不同类型：
`func callback() -> (U8) -> U16 { return noret_call; }` 报出「期望 `(U8) -> U16`，实得 `(U8) -> U16`」——最典型的自相矛盾诊断。

**修法**：两侧皆空 = 同型。Go 侧移植时同样按此实现。

## 缺陷 7：单名局部函数指针调用返回类型恒为 I32（实现缺陷）

`func g(f: (U8) -> U16) -> U16 { return f(1); }` —— `f(1)` 走到 `check_call` 的兜底分支（只处理了 `kind == "local" and len(names) >= 2` 的方法调用），`sig` 为空后直接返回 `I32`，于是返回值检查假报类型不匹配。

**修法**：`local` 单名且类型为 `FnPtr` 时走间接调用通道（designs/0001 §3.9「函数指针值直接调用」）。语料 `fnptr_param` 锁定。

## 缺陷 8：`region` / `unsafe` 块内的 `return` 不被认作发散（实现缺陷）

`func detect(img: [U8]) @region[work] -> I32 { region tmp { return 1; } }` 报 `E3208`（路径不完整）。块里的 `return` 同样终结函数，`stmt_diverges` 却没有把 `Region` / `Unsafe` 当作可发散语句。

**修法**：`Region` / `Unsafe` 的发散性 = 其块的发散性。

## 缺陷 9：取值位置 `match` 的字面量分支产出与槽类型不符的 IR（实现缺陷）

`func classify(n: I32) -> I32 { return match n { 0 => 100, _ => 300 }; }` 的 IR 里出现

```
  %slot___match_2 = alloca i32, align 8
  store i64 100, ptr %slot___match_2
  %t4 = load i32, ptr %slot___match_2
```

`store` 的类型（`i64`）与 `alloca` 的类型（`i32`）不一致——按 LLVM 校验规则这是非法 IR，`llc` 会直接拒绝。根因：分支体的字面量保持未定型类型（`UntypedInt`），`llty()` 对未定型类型走兜底分支返回 `i64`；`match` 汇合时算出的定型结果（`I32`）没有回写到分支体。

**裁决**：`match` 在取值位置确定产出类型后，把定型结果**回写**给所有非块状分支体（仅限未定型字面量），让分支体的 IR 表示与汇合槽一致。

**实现落点**：Python `sema.check_match` 的 `m.result_type = first` 之后；Go `sema.checkMatch` 的 `concrete` 定型之后。语料 `tests/corpus/ir_wide.aila` 的 `classify` / `color_tag` 锁定（`ir_wide.ir` 逐字节判据）。

**影响面**：只影响此前从未被判据覆盖的 `match` 取值路径；若不受此修，任何「整数 match 产出字面量」的程序都无法真正编译成目标文件。

## 缺陷 10：一元运算不把上下文类型下推给操作数（实现缺陷）

`var acc: I32 = 0; ... acc = -1;` 生成的 IR 是 `%t = sub i64 0, 1`（操作数 `1` 与结果都按未定型类型的兜底表示走了 `i64`），随后 `store i32 %t` 再次类型不符。

Python 的 `check_unary` 本来就把 `expected` 下推给操作数（`self.check_expr(e.operand, expected ...)`），Go 移植时漏成了 `nil`：于是 `-1` 的操作数停在 `UntypedInt`，整个一元式也就停在未定型状态。

**裁决**：一元运算与二元运算一样，**上下文类型必须下推**（`Result` / `Optional` 上下文除外，与参考实现一致）。

**实现落点**：Go `sema.checkUnary`（`exp` 取值逻辑）。语料 `tests/corpus/ir_wide.aila` 的 `acc = -1` 分支锁定。

## 缺陷 11：Windows 上的宿主三元组取 llvmlite 的 `msvc`，与 MinGW 链接器不匹配（实现缺陷）

`aila build` 在 Windows 上把三元组推成 `x86_64-pc-windows-msvc`（llvmlite 的默认值），而实际链接用的是 MinGW gcc / 分发的 llvm-mingw（gnu ABI）。后果：
①目标文件里留下 `_GLOBAL_OFFSET_TABLE_` 引用，`gcc` 链接阶段直接报未定义符号；
②Go 版按平台推断出的 `x86_64-w64-windows-gnu` 与 Python 版不一致，同一份工程两边产出的 IR 头两行都不同。

**裁决**：Windows 上一律以 **MinGW 三元组**为准（`x86_64-w64-windows-gnu` / `i686-w64-windows-gnu`），随包 clang 存在时仍以 `clang -print-target-triple` 的回答为准。

**实现落点**：Python `resolve.host_triple` / `llvm_default_or_guess`（先按平台判 gnu，再去问 llvmlite）；Go `resolve.PlatformTriple` 本来就是 gnu，无需改。

## 缺陷 12：构建产物 `.ll` 的行尾随平台漂移（实现缺陷）

Python 侧写 `.ll` 用的是文本模式默认换行，Windows 上变成 CRLF；Go 侧固定写 `\n`。同一份 IR 的字节比对因此永远不等，构建产物也不可复现。

**裁决**：**所有**文本产物（`.ll` / `.json` / `.ail`）一律写 `\n`（`newline="\n"`）。

**实现落点**：Python `build.build` 的 `.ll` 写出；Go 侧本来就写 `\n`。语料判据（`tools/gen_ir.py` 早已用 `newline="\n"`）不受影响。

## 缺陷 13：IR 里内嵌绝对源文件路径，产物不可复现（实现缺陷）

`aila_rt_trap(ptr @.str.N, i64 L, ptr @.str.M, i64 line)` 的第四个参数是行号，第三个是**源文件名**；这个文件名此前取的是 `SourceFile.path`，也就是解析时的**绝对路径**。后果：

| 现象 | 说明 |
|---|---|
| 快照不可跨机器 | 同一份源码在两台机器上编出的 IR 第 6 行不同，示例 IR 快照没法当判据（只有仓库相对布局下偶然一致） |
| 构建不可复现 | 换个路径克隆仓库，`.o` 的字节就不同；增量缓存与二进制比对全失效 |
| 泄露本机路径 | 分发出去的二进制里带开发机用户名与目录结构 |

**裁决**：陷阱信息里的源文件路径**由模块名派生**（确定性）：`_.main` → `main.aila`；`_.sensors.camera` → `sensors/camera`；`std.io` → `std/io.aila`。调用方不再传路径。

**实现落点**：Python `codegen.Lowerer.display_path()`（`file_const` 用它）；Go `codegen.DisplayPath()`（`New` 内部调用）。`tests/corpus/*.ir` 与 `tests/corpus/examples/*.ll` 的 `@.str` 行随之固定为 `main.aila` / `std/io.aila`。

## 缺陷 14：跨包导出没有落地通道（壳禁止有体函数 vs 块内 `public` 只到子包内）

**现象**：按 §2.2 的阶梯，块内 `public` 只到「子包内」，而 §2.1 规定壳里禁止**有体** `func`——
于是「对外导出 AILA 实现」在语法上无路可走：`stdlib/io.aila` 的 `print_i32` 这种实现函数没法给用户代码用。
跨项目依赖一接通就撞上了这件事。

**裁决（按设计者澄清）**：阶梯就是三级——**文件内 ⊂ 子包内 ⊂ 包外**，`public` 恒为「提一级」：

| 声明位置 | 默认 | `public` |
|---|---|---|
| 块内（`module { }` 里） | 文件内 | 子包内（同一目录） |
| 壳内（`module { }` 外面） | 子包内 | **包外**（跨子包 / 跨项目） |

轻量形态（文件里没有 `module` 块）的顶层声明字面上「写在 module 外面」，按壳语义取等级。
**跨包导出的写法**：壳里写前向声明（无体、非 `#[c]`），实现在块内，签名必须一致
（声明没实现 = `E7031`，签名不一致 = `E3201`）——接口与可见性在壳、可热替换的实现在块，两不冲突。

**实现落点**：`internal/parser`（壳内允许无体 `func` + `Attrs.InShell`）、`internal/sema/vis.go`
（三级判定 + `E7020`）、`internal/sema/link.go`（前向声明配对 + `E7031`）、`bindgen` 生成的绑定一律
`#[c] public`（cpack 的接口天然包外可见）、`stdlib` 对外函数 `public`。
判据：`internal/sema/vis_test.go`（六个阶梯用例 + 前向声明三例）、`internal/build/TestCrossProjectDependency`。