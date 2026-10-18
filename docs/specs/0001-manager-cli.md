# SPEC 0001：管理器（`aila` CLI、工程模型、清单、bindgen、工具链供给）

> **编号**：0001 · **类型**：SPEC · **状态**：定稿（M1 实现依据）
> **上游**：[designs/0001](../designs/0001-language-spec.md)（§1.2 B1–B3、§2.1–2.3、§5.9、§11 章、§12.2–12.4）、[rfcs/0001](../rfcs/0001-manager-and-frontend.md)
> **下游**：[specs/0002](0002-compiler-frontend.md)（前端）、[plans/0001](../plans/0001-manager-frontend-dev-plan.md)

---

## 1. 工程模型（不可协商的三条）

| # | 规则 | 依据 |
|---|---|---|
| 1 | **项目 = 包**：工程根目录（含 `aila.toml`）就是一个包；包名 = 清单 `package.name` | §1.2 B1 |
| 2 | **子包 = 目录**：包内每个子目录（除保留目录）是一个子包，子包名 = 目录名；子包可嵌套，路径用 `.` 连接 | §1.2 B2 |
| 3 | **模块 = 文件**：`X.aila` 在其所在子包内名为 `X`；模块是热替换的最小单位 | §1.2 B3/B4 |

Python 标识符化：模块名 / 子包名必须匹配 `[A-Za-z_][A-Za-z0-9_]*`，否则 `E7003`（源码路径与符号一一对应，不允许「目录名叫 `my-lib` 但导入写 `my_lib`」，不做隐式转写）。

### 1.1 目录布局

```text
<包根>/
  aila.toml                 # 清单，唯一
  main.aila                 # 入口模块（包根唯一允许的 .aila 文件），模块名 = _.main
  <子包>/<模块>.aila         # 项目内模块，导入名 = _.子包.模块
  cpacks/<cpack>/           # C 包：include/ 与 .c/.a/.so 共置即关联（§11.8）
  build/                    # 产物与缓存，管理器独占，可整体删除
```

保留目录（不参与子包扫描）：`build`、`cpacks`、`.git`、`.aila`，以及任何以 `.` 开头的目录。

### 1.2 导入解析（§2.3）

| 写法 | 解析 |
|---|---|
| `import _.sensors.camera;` | 文件 `<包根>/sensors/camera.aila`；绑定名 = `camera` |
| `import _.sensors.camera as cam;` | 同上，绑定名 = `cam` |
| `import _;` | 包根模块（即 `_.main`） |
| `import ch _.cpacks.math;` | C 声明导入：对该 cpack 的头文件跑 bindgen，产物 `.ail` 作为模块参与前端 |
| `import cc _.cpacks.legacy;` | C 实现导入：只编译 + 链接该 cpack 内 `.c`，不产 AILA 声明 |
| `import <依赖名>.子包.模块;` | 外部包：`<依赖名>` 必须是清单 `[dependencies]` 里声明过的依赖（见 §2.2）；未声明 = `E7010` |

同一模块内 `as` 别名重复、或别名与模块名撞名 = `E7004`。

### 1.3 入口

清单 `package.entry` 指定入口模块（默认 `_.main`）。入口函数 = 该模块内名为 `main` 的 `func`（`main` 不是关键字，§2.2）；后端把 `main` 直接映射为目标入口符号。找不到 = `E7005`。

---

## 2. 清单 `aila.toml`

TOML 子集（字符串 / 整数 / 布尔 / 表格）。**未知段落/键 = `W7001` 警告并在 `--strict` 下升级为错误**（拼错不静默生效）。

```toml
[package]
name = "demo"          # 必填；[A-Za-z_][A-Za-z0-9_-]*
version = "0.1.0"      # 必填；语义化版本
entry = "_.main"       # 可选，默认 "_.main"

[target]
board = "host"         # 傻瓜路：一个词决定 arch / triple / 地形
arch = ""              # 可选；board 未给时必填（riscv32 / xtensa / x86_64 / aarch64）
triple = ""            # 可选；都给时以 triple 为准（内置 board 表已带 triple）

[build]
opt = 0                # 0..3，默认 0（嵌入式默认关闭优化，WCET 可读）
warnings = "default"   # default | strict | allow，默认 default
debug = false          # 是否产出位置的详细 IR 调试信息（M1：仅影响 IR 注释）

[build.toolchain]      # 傻瓜路整段可省
cc = "auto"            # auto | system:<name>[>=ver] | bundled
llvm = "auto"          # auto | bundled | system:llc | system:clang

[domains.fast]         # 可选：自定义物理域（§5.9）
kind = "internal-ram"
size = "64K"
align = 16

[dependencies]         # 可选：跨项目依赖（rfcs/0004）
mylib = { path = "../mylib" }                        # 本地路径（相对本包根）
util  = { git = "https://…/util.git", tag = "v1.2.0" }  # git：tag 或 rev 锁定，不许分支
math  = { path = "../math", version = "0.3" }         # 可选版本前缀要求
```

校验规则：

| 键 | 规则 | 违反 |
|---|---|---|
| `package.name` | 缺失 / 空 | `E7002` |
| `package.version` | 缺失 | `E7002` |
| `target.board` | 不在内置 board 表且未给 `arch` | `E7006`，诊断附候选 board 列表 |
| `build.opt` | 非 0..3 | `E7007` |
| `domains.*.kind` | 不在 kind 枚举（`internal-ram`/`external-ram`/`rtc-slow`/`default`） | `E7008` |
| `domains.*.size` | 形如 `<num>[KMG]`；嵌入式 target 必填 | `E7008` |
| `dependencies.<名>` | 名是合法标识符；恰好 `path` 或 `git`；git 必须锁 `tag`/`rev` | `E7021` |
| 依赖定位 | `path` 指向包根（含 `aila.toml`）；git 可拉取；`package.version` 存在且满足要求 | `E7022` / `E7023` |
| 依赖图 | 同名依赖同源；有向无环 | `E7024` / `E7025` |
| 跨模块引用 | 三级阶梯：文件内 ⊂ 子包内（块内 `public`/壳内默认）⊂ 包外（壳内 `public`） | `E7020` |

### 2.1 内置 board 表（M1）

| board | arch | triple | 预置域 |
|---|---|---|---|
| `host` | 宿主 | 由 LLVM 默认三元组决定 | `default` |
| `esp32s3` | xtensa | `xtensa-esp32s3-elf` | `default`(internal-ram) / `psram`(external-ram) / `rtc`(rtc-slow) |
| `riscv32` | riscv32 | `riscv32-unknown-elf` | `default` |
| `aarch64` | aarch64 | `aarch64-unknown-none` | `default` |

M1 只有 `host` 可产出可执行产物；其余 board `aila build` 到 `.o` 为止，链接阶段报 `E8020`「交叉链接器待批次」并在 `aila doctor` 中标明。

---

### 2.2 跨项目依赖与 `aila.lock`（rfcs/0004）

依赖 = **另一个包**；`import <依赖名>.子包.模块;` 里的 `<依赖名>` 就是清单里的键。

| 行为 | 规则 |
|---|---|
| 解析顺序 | `vendor/<名>/` 优先（离线）→ `path` 直接定位 → `git` 缓存（未命中才 clone） |
| 编译模型 | 依赖源码一起进模块图（AILA 是源码分发，没有预编译库）；模块名 = `<依赖名>.<子包>.<模块>`，符号天然带包前缀互不冲突 |
| 锁文件 | `aila.lock` 记每条依赖的 `source` / `version` / `hash`（内容哈希）；来源、版本、哈希任一变化 → `E7023`，需 `aila deps --update` |
| 首次构建 | 没有锁文件时自动写入 |
| vendor | `aila deps --vendor` 复制到 `vendor/`；此后即使原路径断掉、无网络也能构建 |
| 传递依赖 | 依赖包自己的 `[dependencies]` 继续生效；按拓扑序编译 |
| 可见性 | 三级阶梯（see §1.2）：块内默认文件内、块内 `public` 子包内、壳内 `public` 包外；违反 = `E7020`。跨包导出函数 = 壳内前向声明 + 块内实现（`E7031`/`E3201`） |

`aila deps` 输出示例：

```text
  0. mylib            0.1.0      path:..\mylib                      ec023f1b088bab2e
根包  <包根目录>
```
## 3. CLI 契约

### 3.1 全局

```text
aila [--version] [--help] <command> ...
```

| 退出码 | 含义 |
|---|---|
| `0` | 成功 |
| `1` | 编译 / 检查 / 构建失败（有诊断） |
| `2` | 用法错误（参数不合法） |
| `3` | 环境不满足（`aila doctor` 项失败） |
| `4` | 管理器内部错误（栈已带 `--verbose`） |

公共开关：`--json`（机器可读输出，供 CI/AI 用）、`--verbose`、`--quiet`、`--color auto|always|never`（默认 auto，`NO_COLOR` 环境变量生效）。

### 3.2 命令清单

| 命令 | 作用 | 关键参数 | 退出码行为 |
|---|---|---|---|
| `aila new <path>` | 建工程骨架（`aila.toml` + `main.aila` + README） | `--board <board>` `--no-README` | 目标非空目录 = `E7011` |
| `aila doctor` | 环境自检 + 链接冒烟测试 | `--json` `--fix`（打印可复制命令，不自动装） | 任一项失败 = 3 |
| `aila check [路径]` | 只到语义检查（不产产物） | `--json` `--deny-warnings` `--target` | 有 error = 1 |
| `aila build [路径]` | 全流程到可执行 / `.o` | `-o` `--target` `--opt` `--emit ir\|json\|bindings\|all` `--no-cache` | 失败 = 1 |
| `aila run [路径]` | `build` 后执行，透传 `-- ARGS` | 同 build + `--` | 被执行程序退出码透传 |
| `aila clean [路径]` | 删 `build/` | `--cache-only` | 0 |
| `aila bindgen [目标]` | 生成 C 绑定 `.ail`（不给目标 = 项目内全部 cpack；目标可为 cpack 名 / `cpacks/<名>` / 含头文件的目录） | `--backend auto\|clang\|builtin` `--target` `--out` | 有未映射项但产物成立 = 0（带 `W3001`）；完全失败 = 1 |
| `aila target list\|show\|add` | 目标与地形表可发现性三件套（§5.9） | `show <board>` | 未知 board = 2 |
| `aila explain <CODE>` | 错误码解释 + 傻瓜式修复步骤 | `--json` | 未知码 = 2 |
| `aila parse <文件>` | 解析一份 `.aila`，打印 token 流（`--tokens`）或语法树 S-表达式 | `--tokens` | 有诊断 = 1 |
| `aila ir <文件>` | 单文件走完「解析 → 语义 → LLVM IR」并打印 IR 文本 | `--triple` | 有诊断 = 1 |
| `aila deps [路径]` | 打印依赖图 / 重写 `aila.lock` / 生成 `vendor/` | `--update` `--vendor` `--json` | 0；有诊断 = 1 |
| `aila lsp` | 启动语言服务器（JSON-RPC over stdio）：大纲 / 跳转定义 / 类型定义 / 跳到实现 / import 链接 / 悬停 / 引用 / 同文件高亮 / 工程符号 / 诊断推送 / 补全 / 签名帮助 / 重命名 / 折叠 / 选区 / 格式化 / 内联提示（§6.6） | `--log <文件>` | 0；内部错误 = 4 |
| `aila skill [文件]` | 给 AI 编程工具的技能包：默认把 `SKILL.md` 打到 stdout，也可打印单篇 / 列出 / 装进 AI 工具目录（§6.7） | `--path` `--list` `--install <目录>` | 0；找不到技能目录 = 1 |

### 3.3 诊断输出格式（人读）

```text
error[E3201]: 类型不匹配
  ┌─ sensors/camera.aila:12:18
  │
12 │     let x: F32 = count;
  │                  ^^^^^ 期望 `F32`，实得 `I32`
  │
  = 帮助：AILA 零隐式转换。整数转浮点请写 `count trunc F32`
  = 参见：docs/designs/0001 §6.3
```

- 多文件汇总时按「先 error 后 warning、组内按文件行号」排序；
- 末尾固定汇总行：`error: 共 3 个错误，1 个警告`（`--quiet` 时省略）；
- `--json` 输出 `{"diagnostics":[{"code","severity","message","labels":[{"file","line","col","span","text","label"}],"help","doc"}],"summary":{...}}`。

### 3.4 错误码分段（D6）

| 段 | 域 | 例 |
|---|---|---|
| `E1xxx` | 词法 | `E1001` 非法字符 / `E1002` 未闭合字符串 / `E1003` 单引号多码点 |
| `E2xxx` | 语法 | `E2001` 期望 token / `E2002` 裸块语句 / `E2003` 歧义需加括号 |
| `E3xxx` | 类型与名字 | `E3001` 未定义名 / `E3201` 类型不匹配 / `E3401` match 不完备 |
| `E4xxx` | 内存与效果 | `E4001` 缺分配许可 / `E4101` 逃逸 / `E4201` 只读绑定被写 |
| `E5xxx` | 并发 | `E5001` `<-` 不在协程内 / `E5002` thread 参数白名单 |
| `E6xxx` | C 互操作 | `E6001` 变参禁传类型 / `E6002` union 不支持 |
| `E7xxx` | 工程、清单与依赖 | `E7002` 清单缺字段 / `E7003` 路径非法标识符 / `E7010` 导入根未声明 / `E7020` 跨包引用未导出 / `E7021` 依赖声明不合法 / `E7022` 依赖定位失败 / `E7023` 依赖与锁不一致 / `E7024` 同名依赖冲突 / `E7025` 依赖成环 |
| `E8xxx` | 工具链与链接 | `E8010` 无可用 LLVM 后端 / `E8020` 交叉链接待批次 |
| `E9xxx` | 批次未实现 | `E9001` 语法已解析但语义待批次 / `E9002` 代码生成待批次 |

`aila explain <CODE>` 数据源 = 包内 `share/aila/data/diag_codes.json`（唯一事实源，禁止两处手写；Go 侧 `internal/diag` 直接读它）。

---

## 4. 工具链供给（§12.4）

解析优先级：**CLI > 清单 > 内置默认**。三件工具，各自独立探测：

| 件 | 用途 | 探测顺序 | 缺失后果 |
|---|---|---|---|
| `llvm-backend` | IR 文本 → `.o` | ①清单 `llvm` 指定 ②随包 `llc`/`clang`（`toolchain/<triple>/bin/`）③系统 PATH 里的 `llc`/`clang` | `E8010`：错误信息直接给出安装 LLVM 或 `aila toolchain install` 的命令 |
| `cc` | 编 C（`import cc`、runtime）与链接 | ①清单 `cc` ②`clang` ③`gcc` ④`cc` | `E8011` |
| `clang` | bindgen 首选档 | `clang` / `clang.exe` | 降级 `builtin` 档 + `W3001` |

`aila doctor` 检查项（全通过才退出 0）：

| # | 检查 | 失败提示（可复制执行） |
|---|---|---|
| 1 | Python ≥ 3.11 | 升级 Python |
| 2 | LLVM 后端可用（三级任一） | `pip install llvmlite` / 安装 LLVM |
| 3 | `cc` 可用且能链接 | 安装 clang 或 MinGW |
| 4 | runtime 源码存在 | 重新拉取仓库 |
| 5 | **链接冒烟测试**：把一段最小 IR 编成 `.o` → 用 `cc` 链接 → 运行 → 校验输出 | 给出两端版本信息与不兼容分析 |
| 6 | 清单可解析（若在工程内） | 指向清单错误码 |

---

## 5. 构建流水线与产物

### 5.1 流水线

```text
1 discover   扫描包 → 模块表（子包/模块/入口），校验命名
2 manifest   解析并校验 aila.toml，合并 CLI 覆盖，定型 target
3 bindgen    对每个 import ch 的 cpack 跑 bindgen（缓存命中即跳过）→ <cpack>.ail
4 parse      每模块：lexer → parser → AST（含 bindgen 产物，当普通 AILA 模块）
5 sema       名字解析 → 类型/布局 → 区域与效果 → match 完备性 → 丢弃/逃逸
6 lower      AST → LLVM IR 文本（每模块一个 .ll）
7 object     .ll → .o（LLVM 后端）
8 csrc       编 `import cc` 的 .c 与 runtime .c → .o（cc）
9 link       全部 .o + cpack 内 .a/.so → 可执行 / 静态库
10 report    build/manifest.json + build/hot.json（分派表元数据）+ 诊断汇总
```

任一步失败即停，前面的产物保留在缓存（供增量重跑）。

### 5.2 产物布局

```text
build/
  <target>/                 # target 名 = board 名或 triple 短名
    <module>.ll             # --emit ir 时保留
    <module>.o
    <cpack>.ail             # --emit bindings 时复制（真身始终在缓存）
    aila_rt.o
    app[.exe]               # 可执行（host）或 app.o（交叉）
    manifest.json           # 构建事实（版本、target、工具链、哈希、时间）
    hot.json                # hot 模块 + public 函数的分派表元数据（M3 消费）
  .cache/
    bindgen/<key>.ail       # key = sha256(头文件内容 + include 链 + defines + target + backend)
    mod/<key>.json          # key = sha256(模块源码 + 依赖模块接口哈希 + 清单哈希 + target)
    smoke/                  # doctor 冒烟测试产物
```

`aila clean` 删 `build/`；`aila clean --cache-only` 只删 `build/.cache/`。

### 5.3 增量规则

模块级：接口哈希 = 该模块导出声明（类型签名 + 布局 + 可见性）的稳定序列化哈希。依赖模块接口哈希不变且自身源码不变 ⇒ 复用 `mod/<key>.json` 与 `.o`。首版允许保守（任何变化全量重编），但**缓存键必须正确**，错误增量比慢更糟。

---

## 6. bindgen（`import ch` → `.ail`）

### 6.0 两种文件别混：`.aila` 是源码，`.ail` 是产物

| 扩展名 | 内容 | 谁生成 | 进版本库 | 参与编译 |
|---|---|---|---|---|
| `.aila` | AILA **源码**模块（可有 `module {}` 块、函数体、`state`…） | 人写 | 是 | 是 |
| `.ail` | **C 绑定声明**：只有壳内声明（`#[c] public func …`、`public const …`、`type X = …`），没有函数体 | `bindgen` 生成 | **否**（`build/` 下，可随时重生成） | 是（当普通 AILA 模块参与前端） |

**用户不需要手动绑定**：把 `cpacks/<名字>/{include/*.h, *.c}` 放好，源码里写 `import ch _.cpacks.<名字>;`
即可——`aila build/run/check` 在解析阶段自动跑 bindgen（缓存键 = 头文件内容 + 档位 + 三元组），
产物落在 `build/.cache/bindgen/`。`aila bindgen` 只是把同一份产物显式写出来（给人看 / 预生成）：

```text
aila bindgen                       # 项目内全部 cpack → build/bindgen/<cpack>.ail
aila bindgen cpacks/mathkit        # 单个 cpack
aila bindgen cpacks/mathkit --backend builtin --out /tmp/mathkit.ail
aila build --emit-bindings         # 把 .ail 一并复制进 build/<board>/bindgen/
```

### 6.1 产物形态（§11.1）

对一个 cpack 生成**一个** `.ail` 文件，形如普通 AILA 模块的**壳声明**（无 `module {}` 块，全为壳内声明）：

```aila
// 由 aila bindgen 自动生成，请勿手改
// cpack: _.cpacks.math   backend: clang   target: host   hash: 3f2a...
// 宏定型：MATH_MAX=2147483647 (I32)；跳过：未定义分支 #ifdef __linux__

import ch _.cpacks.math;              // 供 AILA 侧继续解析 C 实现
type MathErr = I32;
#[c] func math_add(a: I32, b: I32) -> I32;
#[c] func math_scale(v: F32, k: F32) -> F32;
#[c] let math_epsilon: F32;
```

C 侧实现本体（`.c`）不进前端，只进 `cc` 编译与链接（§11.1、§12.2）。

### 6.2 类型映射（§11.2 / §11.3，逐条落地）

| C | AILA |
|---|---|
| `bool`/`_Bool` | `Bool` |
| `char` | `U8` |
| `signed char` / `unsigned char` | `I8` / `U8` |
| `short`/`unsigned short` | `I16`/`U16` |
| `int`/`unsigned int` | `I32`/`U32` |
| `long`/`unsigned long` | 按 target 实宽（ILP32→I32/U32，LP64→I64/U64） |
| `long long`/`unsigned long long` | `I64`/`U64` |
| `size_t`/`uintptr_t` | `USize` |
| `ptrdiff_t`/`intptr_t` | `ISize` |
| `float`/`double` | `F32`/`F64` |
| `long double` | 不支持，`W3001` |
| `T*`/`const T*` | `[T]?` |
| `char*`/`const char*` | `[U8]?` |
| `void*` | `[Unsafe]?` |
| `T**` 及以上 | `[Unsafe]?` |
| opaque struct（仅 `typedef struct X X;`） | 无布局句柄类型，指针处 → `[X]?` |
| 数组成员 `T a[N]` | `[T; N]` |
| `union` | **不直译**，`W3002` + C shim 建议 |
| 位域 | 不支持，`W3001` |
| 无名 struct/enum | 合成名 `Anon_<行>_<列>`，`W3003` |
| 函数指针 | 内联签名 `(T1, T2) -> R`（`_Nullable` 时加 `?`） |
| 变参函数 | 固定参数映射 + 尾随 `...`（产物写 `/* variadic */`） |

### 6.3 命名规范（§11.5）

| C | AILA 名 |
|---|---|
| 函数 / 变量 / 宏常量 | **保留原名** |
| `struct`/`enum`/`typedef` 类型名 | snake → Pascal，剥离 `_t` 后缀（`esp_err_t` → `EspErr`） |
| 与 AILA 保留字（含 `as`/`trunc`/`len`/`copy`/`move`/`state`/`view`/`out`/`trap`/`resume` 等）撞名 | 加 `_c` 后缀（`trunc` → `trunc_c`），并出 `W3004` 诊断「原名 → 新名」 |
| 无名类型 | `Anon_<line>_<col>` |

同一 cpack 内映射后重名 = `E6100`（不做静默加数字后缀，否则调用点不可预测）。

### 6.4 宏（§11.7）

| C 宏形态 | 处理 |
|---|---|
| 对象式常量宏（可在 C 侧求值为常量） | 映射为 `const`，按 C 标准类型规则定型（不发明推断） |
| 函数式宏 / `static inline` | **不自动翻译**；`W3005` 列出名 + 建议写 C shim；`aila bindgen --shim` 生成 shim 骨架（M2） |
| 无法求值的对象宏 | `W3001` 列出，绝不静默丢弃 |

### 6.5 两档后端的行为差异

| 维度 | `clang` 档 | `builtin` 档 |
|---|---|---|
| 触发条件 | `clang` 可用或 `--backend clang` | 显式 `--backend builtin` 或 clang 不可用 |
| `#include` 解析 | 完整 include 链 | 只在 cpack 的 `include/` 与同目录内递归解析 `"..."`；`<...>` 系统头只按内置最小声明表识别 |
| `#define` 常量 | 聚类求值 | 只认字面量 / 字面量四则表达式 |
| `#ifdef` | 按 defines 定型 | 只走「未定义」分支 + 产物头注释 `/* WARNING: conditional compilation not resolved */` |
| packed/aligned | 从 AST 属性透传 | 识别 `__attribute__((packed))` / `__attribute__((aligned(N)))` / `#pragma pack(push,N)` 的文本形态 |

两档产出同一套 §6.2/§6.3 规则，因此 AILA 侧代码无需感知档位。

### 6.6 语言服务器能力矩阵（`aila lsp`）

**纪律**：语义一律由编译器给出，插件只负责「协议 → 编辑器 API」的翻译；插件里不另写语法树。

| LSP 方法 | 能力 | 说明 |
|---|---|---|
| `initialize` | 能力声明 | `textDocumentSync=1`（全量同步）；补全触发字符 `.` `@` `#` `:` `(`；签名帮助触发 `(` `,` |
| `textDocument/publishDiagnostics` | 诊断 | 打开/输入即推；`E**` → Error，`W**` → Warning；同工程其它模块的诊断也一起推 |
| `textDocument/documentSymbol` | 大纲 | 模块 → struct/enum/函数/方法 → 字段/变体 |
| `textDocument/definition` | 跳转定义 | 局部 → 模块级 → `import` 别名 → `Type.member` → 字段 |
| `textDocument/typeDefinition` | 跳到类型定义 | 变量/参数/常量 → 它的 struct/enum/`type` 声明；内置类型没有声明处，返回空 |
| `textDocument/implementation` | 跳到实现 | **壳内前向声明 ↔ 块内实现**互相跳（`E7031` 报的就是这两半没配上）；两边都在时返回两条让编辑器选 |
| `textDocument/documentLink` | 可点击链接 | `import <路径>` 圈成链接指向目标模块文件（`ch` 之类走 bindgen 产物的导入不给链接） |
| `textDocument/hover` | 悬停 | 种类 + 签名/类型 + 来源模块 |
| `textDocument/references` | 查找引用 | 工程内按名字匹配 |
| `textDocument/documentHighlight` | 同名高亮 | 同文件 |
| `workspace/symbol` | 工程符号 | 按子串模糊搜索 |
| `textDocument/completion` | 补全 | 局部/模块符号/导入别名/关键字/内置类型；`alias.` → 目标模块公开符号；`变量.`/`类型.` → 字段+方法+变体；`#[` → 属性名；`@` → 效果名；`import ` → 模块路径；函数项为 `name(${1:参数})` 片段 |
| `textDocument/signatureHelp` | 签名帮助 | 从光标回溯找包住它的 `(`，按深度 0 的逗号定位当前形参 |
| `textDocument/prepareRename` / `rename` | 重命名 | 声明 + 同文件引用 + 跨模块 `别名.名字` 一起改；局部符号只在本模块；关键字与内置类型拒绝；新名字必须是合法标识符 |
| `textDocument/foldingRange` | 折叠 | `{}` 块（`{` 独占行时从声明行起折）、连续 `//`、多行 `/* */`、连续 `import` 段 |
| `textDocument/selectionRange` | 选区 | 标识符 → 括号对（由内到外）→ 行 → 整文件 |
| `textDocument/formatting` | 格式化 | 花括号层级缩进（`options.tabSize`，默认 4）+ 去行尾空白 + 末行恰好一个换行；多行字符串与块注释内部原样保留；**幂等**，无需改动时返回空数组 |
| `textDocument/inlayHint` | 内联提示 | `let`/`var` 未标注类型时显示编译器推断的类型（`kind=Type`） |

#### C 导入（`import ch _.cpacks.<名字>`）在编辑器里的行为

C 包在编译器侧是「导入时现算 bindgen → 生成 `.ail` → 当普通模块」；编辑器要显示补全/跳转/悬停，
**必须拿到同一份 `.ail`**，所以语言服务器复用同一条流水线，而不是另写一套猜测：

| 环节 | 做法 |
|---|---|
| 解析时机 | 分析工程时遇到 `imp.KindV == "ch"` 就调 `bindgen.Run`（`--backend auto` 同款），生成 `.ail` 文本后当模块索引，模块名 `_.cpacks.<名字>`（与 `internal/build` 完全一致） |
| 落盘 | `build/.cache/bindgen-lsp/<cpack>.ail` + `<cpack>.key`（`build/` 本就 gitignore）；跳转 URI 就是这个文件 |
| 缓存 | 进程内按「规则版本 + triple + cpack 指纹」缓存；指纹用「相对路径 + 大小 + 修改时间」，改头文件必然失效；生成失败也做负缓存，避免每次按键重跑 bindgen |
| 可见能力 | `alias.` 补全（`ch` 导入不带 `as` 时绑定名 = cpack 名）、跳转定义 → 生成的 `.ail`、悬停（注明「来自 C 包 `X` 的绑定」）、签名帮助、`import` 链接 → 绑定文件、诊断 |
| 诊断 | cpack 目录不存在 → `E3001`；路径写错 → `E7010`（与编译器同码同话术，直接标在 `import` 那一行） |
| `import cc` | 只让 C 源参与编译、不产生绑定模块，因此编辑器里不索引（符合编译器的语义） |
| 只读保护 | 生成的 `.ail` 上拒绝改名与格式化（改了也会被下次生成覆盖） |

### 6.7 给 AI 编程工具的技能包（`aila skill`）

**定位**：用户的 AI 面向的是**这个工具**（`aila` 命令），不是编译器源码。所以技能讲的全是
「怎么用打包好的 aila 把程序做出来」，一份二进制也不用带。

| 项 | 约定 |
|---|---|
| 位置 | `<AILA_HOME>/share/aila/skills/aila/`（仓库里同一相对路径；既随发布包分发，也可单独 zip） |
| 入口 | `SKILL.md`，带 YAML frontmatter（`name: aila` + 一句 description），供 AI 工具识别与触发 |
| 结构 | `SKILL.md`（铁律 / 上手 / 工作循环 / 可见性阶梯 / 最小子集 / 报错速查 / 排错剧本）+ `reference/{language,recipes,manifest,cli,errors}.md` |
| 生成物 | `reference/cli.md` 与 `reference/errors.md` 由 `go run ./tools/skillgen` 从命令树与 `share/aila/data/diag_codes.json` 生成；CI 用 `-check` 卡住漂移 |
| 命令 | `aila skill`（打印入口）/ `aila skill <文件>`（打印单篇）/ `--list` / `--path` / `--install <目录>`（复制成 `<目录>/aila/`） |
| 分发 | 发布包内 `share/aila/skills/`；另出 `aila-skill-<version>-<platform>.zip`（只含 `aila/`，见 specs/0003 §1） |
| 判据 | `internal/skill` 单测（目录随 AILA_HOME / 列文件 / 读 / 安装）；`tools/skillgen -check` 在 CI 全绿；发布包解开后 `./aila skill --list` 列出 6 个文件 |

---

## 7. 验收命令（本 SPEC 的测试面）

```text
aila new /tmp/demo --board host          # 生成骨架，aila.toml + main.aila 齐全
aila check /tmp/demo                     # 0 退出；故意破坏语法 → 1 退出且带 E2xxx
aila build /tmp/demo --emit ir           # build/host/main.ll 存在且可被 llc 验证
aila run /tmp/demo                       # 打印预期输出，退出码透传
aila bindgen cpacks/math --backend builtin   # 产出 .ail + W3001 诊断
aila target show esp32s3                 # 打印 arch/triple/预置域
aila explain E4101                       # 打印含义 + 修复步骤 + 设计锚点
aila clean                               # build/ 消失
```
