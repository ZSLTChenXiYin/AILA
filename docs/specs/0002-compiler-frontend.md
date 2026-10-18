# SPEC 0002：编译器前端（词法 / 语法 / AST / 语义 / LLVM IR）

> **编号**：0002 · **类型**：SPEC · **状态**：定稿（M1 实现依据）
> **上游**：[designs/0001](../designs/0001-language-spec.md)（语法与语义唯一权威）、[rfcs/0001](../rfcs/0001-manager-and-frontend.md)
> **同批**：[specs/0001](0001-manager-cli.md)（管理器）

前端职责：**AILA 源码 → LLVM IR 文本**（[designs/0001 §12.1](../designs/0001-language-spec.md)，不转译 C）。本文只写「实现契约」：阶段划分、token、文法、AST 节点、内部类型表示、检查项与错误码、IR 约定、M1 落地边界。语义本身一律以 designs/0001 为准，本文不复述其理由。

---

## 1. 流水线

```text
源文件 ──lexer──► Token[] ──parser──► AST ──sema──► 已定型 AST + ModuleFacts ──lower──► IR 文本
                                          │
                                          └─ 依赖：名字表（本项目模块 + bindgen .ail + std）、布局表
```

| 阶段 | 输入 | 输出 | 纯函数？ |
|---|---|---|---|
| `lexer` | 源码文本 + 文件名 | `Token[]` | 是 |
| `parser` | `Token[]` | `ast.Module` | 是 |
| `sema` | 本项目全部 `ast.Module` + 依赖模块的 `ModuleFacts` | 带类型标注的 AST、每模块 `ModuleFacts`（导出签名、布局、`hot`/`public` 事实）、诊断 | 否（跨模块） |
| `lower` | 单模块已定型 AST + 布局表 | IR 文本（字符串） | 否（需链接符号表） |

顺序固定：`discover → manifest → bindgen → parse（全部） → sema（全部） → lower（逐模块） → object`。任一步有 `error` 级诊断即中止，不产半成品产物。

### 1.1 源码约定

- 源文件恒 UTF-8（§3.2）；BOM 允许并跳过；行尾 `\n` 与 `\r\n` 等价；
- 注释：`//` 行注释、`/* */` 块注释（可嵌套？**不嵌套**，与 Rust 之外的主流一致；未闭合 = `E1004`）；
- 行号从 1、列号从 1，列按 **Unicode 码点**计（诊断里 `^` 数量 = 码点数）。

---

## 2. Token

| 类别 | 内容 |
|---|---|
| 关键字（保留字） | `func let var const type struct enum state view module hot import public return if else for in match region defer unsafe coroutine thread yield out trap move len copy resume as trunc break continue true false` |
| 内置类型名（非保留字，可作普通标识符？**不可**——`I32` 等由名字解析在类型位识别，普通位撞名报 `E3009` 提示改名） | `I8 I16 I32 I64 U8 U16 U32 U64 F32 F64 Bool C32 Text USize ISize` + `Unsafe chan send recv coro` |
| 内置变体名（非关键字） | `None Some Ok Err Suspended Done` |
| 标识符 | `[A-Za-z_][A-Za-z0-9_]*`；`_` 单独出现 = 通配/丢弃 token |
| 整数 | 十进制、`0x`、`0b`、`0o`；下划线分隔；**无后缀**（§6.1） |
| 浮点 | `1.0`、`1e9`、`1.5e-3`、`1_000.5`；**`1.` 与 `.5` 均不合法**（`E1006`，要求前导/尾随数字） |
| 字符串 | `"..."`（Text）、`r"..."`（Raw，反斜杠零转义） |
| 字符 | `'X'`（C32，恰好一个码点，转义后亦然；`'ab'`/`''` = `E1003`） |
| 转义 | `\n \t \r \\ \" \' \0 \u{...}`；其他转义 = `E1005` |
| 标点 | `( ) { } [ ] , ; : . @ # -> => = == != < <= > >= + - * / % += -= *= /= %= & &= \| \|= ^ ^= << <<= >> >>= && \|\| ! ~ ? <- ::`（`::` 不合法，`E1007` 提示用 `.`） |

`?` 的三重身份（§6.9）由**语法位置**消解：紧跟 `(` = 陷阱检测前缀；表达式后缀 = 传播；类型后缀 = 可空。词法层不区分，parser 按上下文解释。

---

## 3. 文法（EBNF，`{ }` = 重复零次以上，`[ ]` = 可选）

```ebnf
file        = { import } { shell_decl } [ module_block ] { closing_decl }?  ;
import      = "import" [ "ch" | "cc" ] path [ "as" ident ] ";" ;
path        = root { "." ident } ;                       root = "_" | ident ;

module_block= "module" [ "hot" ] "{" { block_decl } "}" ;
shell_decl  = { attr } ( type_decl | const_decl | enum_decl | struct_decl
                       | extern_func_decl | extern_var_decl ) ;
block_decl  = { attr } ( func_decl | struct_decl | enum_decl | type_decl
                       | const_decl | let_decl | var_decl | state_decl ) ;

struct_decl = "struct" ident "{" [ field { "," field } [ "," ] ] "}" ";" ;
field       = ident ":" type ;
enum_decl   = "enum" ident [ ":" type ] "{" [ variant { "," variant } [ "," ] ] "}" ";" ;
variant     = ident [ "=" const_expr ] [ "{" [ field { "," field } [ "," ] ] "}" ] ;
type_decl   = "type" ident "=" type ";" ;
const_decl  = "const" ident [ ":" type ] "=" expr ";" ;
let_decl    = "let"  ident [ ":" type ] "=" expr ";" ;
var_decl    = "var"  ident [ ":" type ] "=" expr ";" ;
state_decl  = "state" ident [ ":" type ] "=" expr ";" ;        // 仅 module hot 块内合法（§4.2-11）
view_decl   = "view" ident ":" type "=" expr ";" ;             // 语句位

func_decl   = [ "public" ] [ "coroutine" | "thread" ] "func" cname
              "(" [ params ] ")" { effect } [ "->" type ] body ;
cname       = ident [ "." ident ] ;                            // 类型前缀式方法/关联函数（§4.4）
params      = param { "," param } ;
param       = "self" | ident ":" type ;
effect      = "@" ( "region" | "mut" | "move" ) "[" ident { "," ident } "]" ;
body        = block | ";" ;                                    // ";" 仅 #[c] 声明合法
extern_*    = 同 func_decl/var_decl 但必须带 #[c] 属性且无 body/无初始化器

type        = type_atom { "?" } [ "!" type_atom_error ] ;
type_atom   = prim_type | "Text" | "[" type "]" | "[" type ";" const_expr "]"
            | "[Unsafe]" | "chan" type | "send" "chan" type | "recv" "chan" type
            | "coro" [ type [ "!" type ] ] | "(" type { "," type } ")"
            | "(" [ types ] ")" "->" type                    // 函数指针签名
            | path ;

stmt        = block_stmt | let_decl | var_decl | const_decl | view_decl | state_decl
            | return_stmt | if_stmt | for_stmt | match_stmt | region_stmt
            | defer_stmt | unsafe_stmt | trap_stmt | yield_stmt | thread_stmt
            | break_stmt | continue_stmt | assign_stmt | expr_stmt ;

return_stmt = "return" [ expr | "_" "!" expr ] ";" ;
if_stmt     = "if" expr block { "else" "if" expr block } [ "else" block ] ;
for_stmt    = "for" block                                   // 无限循环
            | "for" expr block                               // 条件循环（吸收 while）
            | "for" var_decl_nosemi ";" expr ";" assign_stmt_nosemi block
            | "for" [ ident ":" type ] [ ident "," ] ident "in" expr block ;
match_stmt  = "match" expr "{" { arm } "}" ;
arm         = pattern "=>" ( block | expr ) [ "," ] ;
pattern     = "_" | literal | path | path "{" [ fpat { "," fpat } [ "," ] ] "}"
            | "(" ident { "," ident } ")" ;
fpat        = ident [ ":" ( literal | ident | "(" ident { "," ident } ")" ) ] ;
region_stmt = "region" ident [ ":" ( dom_or_size ) ] block ;
dom_or_size = ident [ "," const_expr ] | const_expr ;
defer_stmt  = "defer" ( block | expr ";" ) ;
unsafe_stmt = "unsafe" block ;
trap_stmt   = "trap" expr ";" ;
yield_stmt  = "yield" ";" ;
thread_stmt = "thread" [ ":" expr ] call ";" ;
assign_stmt = lvalue ( "=" | "+=" | "-=" | "*=" | "/=" | "%=" | "&=" | "|=" | "^="
                     | "<<=" | ">>=" ) expr ";"
            | "_" "=" expr ";" | "(" ident { "," ident } ")" "=" expr ";" ;
expr_stmt   = expr ";" ;
```

### 3.1 表达式优先级（§6.9，11 档）

| 档 | 形式 | 结合 |
|---|---|---|
| 0 | 原子、`(expr)`、多值组 `(e1, e2)`、数组字面量、构造、`#size(T)`/`#align(T)`、`chan@r T(n)`、`coroutine@r f(...)` | — |
| 1 | 后缀链：`f(args)`、`a[i]`、`a.b`、`e?`、`e@[r, ...]`、`lit@r` | 左 |
| 2 | 前缀：`- ! ~ ?(expr) <- move len copy resume` | 右 |
| 3 | `as` `trunc` | 左 |
| 4 | `* / %` | 左 |
| 5 | `+ -` | 左 |
| 6 | `<< >>` | 左 |
| 7 | `&` | 左 |
| 8 | `^` | 左 |
| 9 | `\|` | 左 |
| 10 | `== != < <= > >=` | 左 |
| 11 | `&&` | 左 |
| 12 | `\|\|` | 左 |

解析禁区（照 designs/0001 直接出诊断，不静默兜底）：

| 写法 | 诊断 |
|---|---|
| `a < b < c` | `E2004` 比较链禁用，改用 `a < b && b < c` |
| 裸块语句 `{ ... }` | `E2002` 独立裸块禁用；局部作用域用 `region 名字 { }` |
| `len(x)` / `x.len` / `x.copy()` | `E2005` 定形关键字必须前缀式（`len x`） |
| `let _ = f();` | `E2006` 显式丢弃的正确写法是 `_ = f();` |
| `[...]@[r]` / `f()@r` | `E2007` 两种 `@` 拼法不得混用 |
| `if` 作表达式（`let x = if ...`） | `E2008` if 恒为语句，取值用 match |
| match 头尾随构造字面量且未加括号 | `E2003` 请写 `match (Shape.Circle { r: 1.0 }) { ... }` |
| `1.` / `.5` / `1u8` / `0o777` 以外的 C 式八进制 | `E1006` 数字字面量拼法 |

---

## 4. AST 节点（`internal/ast`，全部带 `span=(file,line,col,len)`）

| 分组 | 节点 |
|---|---|
| 顶层 | `Module(file, imports, shell_decls, block_decls, is_hot, is_lightweight)`、`Import(path, kind, alias)` |
| 类型声明 | `StructDecl(name, fields, attrs)`、`EnumDecl(name, base, variants, attrs)`、`Variant(name, discriminant, payload_fields)`、`TypeAlias(name, target)` |
| 函数 | `FuncDecl(name, owner_type, kind(normal/coroutine/thread), is_public, params, effects, ret, body, attrs, is_extern)`、`Param(name, type, is_self)`、`Effect(kind, names)` |
| 绑定 | `LetDecl/VarDecl/ConstDecl/StateDecl(name, type, init)`、`ViewDecl(name, type, init)` |
| 语句 | `Block(stmts)`、`Return(value_or_err)`、`If(cond, then, elifs, els)`、`ForC(init, cond, step, body)`、`ForIn(idx, var, var_type, iterable, body)`、`Match(head, arms)`、`Arm(pattern, body)`、`Region(name, domain, size, body)`、`Defer(body)`、`Unsafe(body)`、`Trap(msg)`、`Yield`、`ThreadStart(stack_size, call)`、`Break/Continue`、`Assign(target, op, value)`、`DestructAssign(names, value)`、`Discard(value)`、`ExprStmt(expr)` |
| 模式 | `Wildcard`、`LitPattern(value, neg)`、`PathPattern(path)`、`VariantPattern(path, fields)`、`FieldPat(name, value_or_binding, group)`、`GroupPattern(names)` |
| 表达式 | `IntLit/FloatLit/StrLit/CharLit/BoolLit`、`Path(names)`、`ArrayLit(elems or (value,count))`、`StructLit(type_path, fields)`、`VariantLit(path, fields)`、`Call(callee, args, region_args)`、`MethodCall(recv, name, args, region_args)`、`Index(base, idx)`、`Field(base, name)`、`Binary(op, l, r)`、`Unary(op, e)`、`Cast(op(as/trunc), e, type)`、`TryPropagate(e)`、`TryDetect(e)`（`?(expr)`）、`Move(e)`、`Len(e)`、`Copy(e, region_args)`、`Resume(e)`、`Recv(e)`（`<-ch`）、`Alloc(expr, region)`、`GroupExpr(elems)`、`SizeOf(type)/AlignOf(type)`、`ChanCreate(elem_type, cap, region)`、`CoroCreate(call, frame_region)`、`Placeholder`（`_`） |
| 属性 | `Attr(name, args)` |

---

## 5. 内部类型表示与布局

### 5.1 类型项（`internal/types`）

```text
Prim(kind)                       kind ∈ {i8,i16,i32,i64,u8,u16,u32,u64,f32,f64,bool,c32,text,usize,isize}
View(elem)                       [T]   —— (ptr, len) 胖视图
FixedArray(elem, n)              [T; N] —— 内联
UnsafeView                        [Unsafe]
Optional(inner)                  T?    —— 仅句柄类（§3.7 规则 2），标量可空 = E3601
Result(ok, err)                  T ! E
Chan(elem, dir)                  chan T / send chan T / recv chan T
Coro(ok)                         coro / coro T / coro T ! E
FnPtr(params, ret, err)          (T1, T2) -> R ! E
Group(types)                     多值组（仅四个固定位置 + match 值叶，§9.1）
Named(decl_ref)                  struct / enum / type 别名（别名在解析期展开）
UntypedInt / UntypedFloat        多态字面量（§6.1）
Never                             trap / 发散
```

### 5.2 布局（C ABI，§11.4）

| 类型 | size | align | 备注 |
|---|---|---|---|
| `iN`/`uN`/`fN` | N/8 | N/8 | |
| `Bool` | 1 | 1 | |
| `C32` | 4 | 4 | |
| `Text` / `View` / `UnsafeView` | `2*ptr` | `ptr` | (ptr, len) |
| `FixedArray(T,N)` | N*size(T) | align(T) | |
| struct | 按声明序自然对齐，尾部补齐 | max(field align) | `#[packed]` 去 padding 且 align=1；`#[align(N)]` 覆盖 max |
| enum（无载荷） | 底层类型宽度（默认 U32） | 同底层 | 判别值 = 显式或递增 |
| enum（带载荷） | tag + payload（最小 tag 宽度：2ⁿ≥变体数） | max | payload = 最大变体字段序列 |
| `Optional(handle)` | = size(handle) | 同 | niche 布局：None = ptr 全 0 |
| `Result` | tag + ok + err（M1 直白布局，不优化） | max | 与 C ABI 无关（不跨 C 边界） |
| `FnPtr` | `ptr` | `ptr` | |
| `Coro` / `Chan` | `ptr` | `ptr` | |

`#size(T)` / `#align(T)` 直接返回上表结果（编译期常量，收 `USize`）；视图 / opaque / 函数指针 = `E3602`。

### 5.3 两族判定（§3.6）

`is_move_type(T)`：含 `View`/`Text`/`UnsafeView`/`Chan`/`Coro`/`Result` 字段者（递归）。`Result` 为 move 类型。值语义族 = 其余。

---

## 6. 语义检查项与错误码（M1 落地 / 待批次逐条标注）

| 组 | 检查 | 码 | 批次 |
|---|---|---|---|
| 名字 | 未定义名 / 路径不可见（可见性梯子 §2.2） | `E3001` / `E3002` | M1 |
| 名字 | 导入别名重复、模块重名 | `E7004` | M1 |
| 名字 | 与内置类型名/关键字撞名的普通声明 | `E3009` | M1 |
| 类型 | 零隐式转换：赋值 / 实参 / 返回 / 运算符操作数 | `E3201` | M1 |
| 类型 | `as` 不满足无损判据 / `trunc` 用于无损场景（警告） | `E3202` / `W3203` | M1 |
| 类型 | 标量可空 `U8?`、非法 `T?` 目标 | `E3601` | M1 |
| 类型 | `#size`/`#align` 收非法类型 | `E3602` | M1 |
| 类型 | 一元负号作用于无符号目标 | `E3204` | M1 |
| 类型 | 比较链 / 视图与 struct 比较 | `E2004` / `E3205` | M1 |
| 语句 | 有值函数裸调用、有值表达式作语句 | `E3206` | M1 |
| 语句 | 无值调用被绑定 | `E3207` | M1 |
| 语句 | 有值函数路径完备性（所有路径显式 return） | `E3208` | M1 |
| 语句 | `let` 绑定被写入 / `@mut` 实参不是 var | `E4201` | M1 |
| 语句 | 复合赋值操作数类型（浮点 `%=` 等） | `E3209` | M1 |
| match | 分支产出统一性（全有 / 全无） | `E3402` | M1 |
| match | enum 完备性、`_` 位置、重复模式 | `E3401` | M1 |
| match | 整数族缺 `_` | `E3403` | M1 |
| match | 块分支缺 `out` / 多余 `out` | `E3404` | M1 |
| 区域 | 分配点落在不可见区域（效果闭合 §5.2） | `E4001` | M1 |
| 区域 | 逃逸：返回/存储体内 region 块视图 | `E4101` | M1 |
| 区域 | 非区域来源绑成 `[T]`（引用无来处） | `E4002` | M1 |
| 区域 | 区域嵌套同名、`region` 参数名与块名冲突 | `E4003` | M1 |
| 区域 | `view` 硬约束（无类型标注 / 同型 / 带 `@` / 宽度改变未进 unsafe） | `E4004` | M1 |
| 区域 | `@` 子句顺序、重复子句、`@mut`∩`@move` | `E4005` | M1 |
| 区域 | hot 模块内 `@static` 与模块级 `var`（§2.4 规则 3、§5.8） | `E4006` | M1 |
| 所有权 | `move` 后使用、二次 move（主名字唯一） | `E4301` | M1（保守：同函数内线性追踪） |
| 所有权 | `copy` 缺区域实参 / `@[...]` 数量不匹配 | `E4302` | M1 |
| 并发 | `<-` 不在 `coroutine func` 内 | `E5001` | M1 |
| 并发 | thread 入口参数不在白名单 | `E5002` | M1 |
| 并发 | thread 入口写模块级 `var` | `E5003` | M1 |
| 并发 | `yield` 在 region / defer 块内 | `E5004` | M1 |
| 互操作 | 变参调用传 struct / 带载荷 enum / Text | `E6001` | M1 |
| 批次 | 已解析但语义未实现 | `E9001` | — |
| 批次 | 已定型但代码生成未实现（协程/chan/thread/热替换产物） | `E9002` | — |

「待批次」项必须**显式报警而不是静默通过**：未落地的检查项在 `check`/`build` 中以 `W9001`「`<检查名>` 检查待批次，本项未校验」列出，汇总行提示条数。

---

## 7. LLVM IR 生成约定（`internal/codegen`）

### 7.0 标准库的 C 边界（用 `#[c]`，不新增语法）

标准库**能用 AILA 实现的就用 AILA 实现**（数字格式化、UTF-8、浮点小数位都在 `stdlib/io.aila`）；
只有平台相关的系统调用交给 C，以最小 `#[c]` 绑定出现：

```aila
#[c] func aila_rt_write(bytes: [U8], n: USize);
#[c] func aila_rt_flush();
```

符号直接就是运行时导出的名字（`runtime/aila_rt.c`），参数映射与其它 `#[c]` 一致（视图家族传 `ptr`）。
**不要为此引入新声明形态**——曾试验过 `#[intrinsic]`，已按设计决定删除。

### 7.1 模块骨架

```llvm
; AILA module: _.sensors.camera   target: x86_64-w64-windows-gnu   backend: llvmlite/LLVM 22.1
target triple = "x86_64-w64-windows-gnu"

%aila.view = type { ptr, i64 }        ; [T] / Text / [Unsafe] 统一胖视图
```

| 约定 | 内容 |
|---|---|
| 指针 | **不透明指针 `ptr`**（LLVM 15+ 硬要求） |
| triple | 由 target 决定；`host` = LLVM 默认三元组；可用清单 `triple` 覆盖。**不写 `target datalayout`**（交给目标默认值，避免手抄错） |
| 名字修饰 | 普通函数 `@<模块扁平名>__<函数名>`（模块路径 `_`/`.` → `__`）；方法 `@<模块扁平名>__<类型>__<方法名>`；`#[c]` 声明用**原 C 名**且只 `declare`；入口 `main` 额外产 `@main` wrapper（`define i32 @main()`） |
| 全局 | `const` 标量 → `@g.<module>.<name>` 带 `constant`；模块级 `var` → `global`；两者均带显式对齐 |
| `hot` 模块 | 本批：函数照常生成 + `hot.json` 记录 `{module, version_hash, publics:[{name, signature}]}`；跨模块调用**不走**分派表（M3 加，`W9002` 提示） |

### 7.2 类型到 IR 的映射

| AILA | IR |
|---|---|
| `iN/uN` | `iN` |
| `Bool` | `i8`（真值 `icmp ne i8 %v, 0`） |
| `C32` | `i32` |
| `F32/F64` | `float`/`double` |
| `USize/ISize` | 目标指针宽度整数（host x64 = `i64`） |
| `Text`/`[T]`/`[Unsafe]` | `{ ptr, i64 }` |
| `[T; N]` | `[N x <T>]` |
| struct | 具名 `%<module>.<Type>` = `{ ... }`（字段序 + padding 由 LLVM 按 C ABI 自然补齐）；`#[packed]` → `<{ ... }>` |
| enum 无载荷 | 底层整数 |
| enum 带载荷 | `{ iN tag, payload }` |
| `T?`（句柄） | 同 `T`，None = `ptr null`（+ len 0） |
| `T ! E` | `{ i1 is_ok, T ok, E err }`（M1 直白布局） |
| 多值组 | 匿名 `{ ... }` 结构体（仅函数返回 / 解构位） |

### 7.3 运行时 ABI（`runtime/aila_rt.c` 导出，AILA 侧 `#[c]` 声明）

```c
typedef struct { unsigned char* ptr; size_t len; } aila_view;   /* 与 {ptr,i64} 同布局 */
void*      aila_rt_region_enter(size_t cap_hint);   /* 返回区域实例指针；cap=0 表示无上限 */
void*      aila_rt_region_alloc(void* region, size_t size, size_t align);
void       aila_rt_region_leave(void* region);
void       aila_rt_trap(const char* msg, size_t len, const char* file, size_t line) /* noreturn */;
void       aila_rt_print_i64(long long);   void aila_rt_print_u64(unsigned long long);
void       aila_rt_print_f64(double);      void aila_rt_print_str(const char* p, size_t len);
void       aila_rt_print_char(unsigned);   void aila_rt_print_bool(int);
void       aila_rt_print_nl(void);         void aila_rt_print_cstr(const char*);
```

- 分配点 `[v; N]@r` → `%p = call ptr @aila_rt_region_alloc(ptr %r, i64 size, i64 align)` + `%v = insertvalue {ptr,i64} {ptr null, i64 0}, ptr %p, 0` + `..., i64 N, 1`；
- 区域块 `region r { }` → `%r = call ptr @aila_rt_region_enter(i64 cap)` / `call void @aila_rt_region_leave(ptr %r)`。**注意**：`region` 变量与「区域参数（@region）」同为 `ptr` 类型句柄，命名空间在 sema 内统一。
- 跨函数传区域：调用点 `@[work]` → 把 `%work` 作为附加隐藏实参**按签名顺序前置**（IR 级细节，源码不可见；`@region` 参数在 IR 参数表里排在最前）。

### 7.4 语句与表达式降级

| AILA | IR |
|---|---|
| `let x = e;` | 栈槽 `%x = alloca`（入口块统一分配）+ `store`；值语义类型直接 `store`，视图存 `{ptr,i64}` |
| 赋值 | `store`（需 `@mut` / var 检查通过） |
| `if` | `br i1` + 两分支 + merge 块 |
| `for`（三段/条件/无限） | `cond` / `body` / `step` / `exit` 四块 |
| `for x in view` | 头固化 `len`，索引循环；元素 = `load`（拷贝） |
| `match`（整数族） | `switch`（有 `_` 时 default → 兜底块） |
| `match`（enum） | 链式 `icmp eq` 于 tag（变体数 > 4 用 `switch`） |
| `?` 传播 | 判 `is_ok` → 否则构造错误返回值并 `ret` |
| `?(expr)` | 每个陷阱点改用 `*.with.overflow` / 显式检查，任一失败 → `None`；否则 `Some` |
| 整数 `+ - *` | `llvm.sadd.with.overflow.*` / `uadd` / `ssub` / `usub` / `smul` / `umul`，溢出 → trap 块 |
| `/ %` | 除零检查 + `sdiv`/`srem`（有符号另查 `INT_MIN / -1`） |
| `<< >>` | 计数 `>= 位宽` trap；`>>` 有符号用 `ashr`、无符号 `lshr` |
| 下标 `v[i]` | `i >= len` trap + `getelementptr` |
| `trap "msg"` | `call void @aila_rt_trap(ptr @.str, i64 len, ptr @.file, i64 line)` + `unreachable` |
| `defer` | 每条退出边（`return` / 块尾 / `?` 上抛）前插入延迟体调用；LIFO |
| `yield` / `resume` / `chan` / `thread` | **M1 不生成**：`E9002`（已定型，待批次） |
| `std.io.print_*` | `#[c]` 声明直接调用 `aila_rt_print_*` |

### 7.5 陷阱块形态

```llvm
ovf:                                              ; 每个函数一个共享 trap 块即可
  %m = getelementptr inbounds [<n> x i8], ptr @.trap.ovf, i64 0, i64 0
  call void @aila_rt_trap(ptr %m, i64 <n>, ptr %file, i64 <line>)
  unreachable
```

行号信息：`trap` 消息与位置用函数内的字面量常量，保证「哪一行溢出」可诊断（§12.3 一条通道）。

---

## 8. M1 落地边界（一份清单，不许含糊）

| 能力 | M1 状态 |
|---|---|
| 全部 v0 语法解析（含 coroutine / chan / thread / unsafe / view / state / 属性 / `@` 子句） | ✅ 解析 + 结构性诊断 |
| 标量、struct、enum、类型别名、函数、方法、`let/var/const`、`if/for/match`、`defer`、`trap`、`view`、`region`、`[T]` / `[T;N]` / `Text` / `[Unsafe]` | ✅ 语义 + 代码生成 |
| `T?`（句柄）、`T ! E`、`?` 传播、多返回值与解构、`as`/`trunc`、陷阱检查、`?(expr)` | ✅ 语义 + 代码生成 |
| `#[c]` 外部声明与 C 调用、变参调用 | ✅ 代码生成（变参类型检查做静态判定） |
| `unsafe { }` 块（类型重标检查放宽） | ✅ 语义（无额外代码生成） |
| 所有权主名字追踪（`move` 后使用） | ✅ 同函数内线性追踪（跨分支保守） |
| `coroutine` / `chan` / `thread` / `std.rt` 热替换 | ❌ 语义检查在，代码生成 `E9002` |
| 热替换 MLO / 双槽 / 分派表代码 | ❌ M3（仅产 `hot.json`） |
| 借用区间（NLL）精确检查 | ❌ M2（M1 只做「只读绑定不可写」） |
| 优化 pass / LTO | ❌ 只透传 `-O`（M1 实际固定 `-O0` 等价：无优化标记） |
| 完整 std.* | ❌ 只 `std.io`（打印）/`std.math`（回绕）/`std.mem`（等值比较） |

---

## 9. 验收（本 SPEC 的测试面）

1. `tests/test_lexer.py` / `test_parser.py`：全部 token 与文法分支各至少一例，负例命中对应错误码；
2. `tests/test_sema.py`：§6 表中每个 M1 错误码至少一正一反例；
3. `tests/test_layout.py`：§5.2 每种类型 size/align 与 C 侧 `sizeof`/`alignof` 实测一致；
4. `tests/test_codegen.py`：示例程序的 IR 经 `llvmlite` 校验（`mod.verify()`）通过；
5. `examples/` 六个示例 `aila run` 输出与期望完全一致；
6. 生成 IR 不含「不该出现的形态」：无 typed pointer、无实验 intrinsic、无未声明符号。

---

## 10. 实现口径补充（M1 落地时固化，与 §7 同等效力）

写在这里的每一条都是实现时做出、且需要被后续批次看清的决定；与 `designs/0001` 不冲突，属实现自由度内的取舍。

| 项 | 口径 | 理由 |
|---|---|---|
| `Bool` 的 IR 类型 | `i1`（内存布局仍是 1 字节） | 比较指令天然产出 `i1`；若映射成 `i8` 则每次比较都要 `zext`。`#[c]` 边界的 `Bool` 仍按 C ABI 传 `i8` |
| `@` 效果子句位置 | `->` 之前为规范位置，`->` 之后也接受 | 见 [rfcs/0002](../rfcs/0002-design-defects.md)：设计稿 §13.2 的 std 示例用的是后置写法 |
| `@static` 分配 | 运行时提供 `aila_rt_static_region()`，进程级常驻区域 | 嵌入式目标由链接期段落位实现（M2），接口不变 |
| 动态长度分配 | `[v; n]@r` 的 `n` 可为运行期值（仅分配点） | §5.1「`[0; n]` 运行期长度」；定长 `[T; N]` 仍要求编译期常量 |
| 函数链接性 | M1 全部函数外部链接 | 跨模块调用要能解析；分派表在 M3（`W9002` 明示） |
| `Result` 布局 | `{ i1, ok, err }`（直白，不优化） | 不跨 C ABI（§11.6），M2 可换 niche |
| `#[c]` 形参映射 | 视图家族只传 `ptr`（len 丢弃）；enum 按底层整数 | §11.3「视图作 C 指针形参时只传 ptr」 |
| 数组字面量 | 未标分配点的 `[...]` 产视图 = `E4001` | §5.8「分配点必须显式」；例外是标注为 `[T; N]` 的内联数组 |
| 模式绑定作用域 | 每个 match 分支单独一个作用域 | 分支间不串名（与 `if`/`for` 块一致） |
| 栈槽分配 | 按 `Symbol` 身份分配（同名遮蔽不串槽） | 保证 `for var i` 与 `for i, x` 共存时语义正确 |
