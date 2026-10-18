# AILA

AILA 是一门面向嵌入式（ESP32-S3 优先）、以**区域内存模型**与**模块级原生热替换**为核心的编程语言。
本仓库是它的**设计文档 + 管理器 + 编译器前端**——管理器与前端是**一份 Go 单文件**（`aila`），
无 cgo、无脚本运行时，解压即用（`rfcs/0003`）。

- 语言设计（唯一权威）：[`docs/designs/0001-language-spec.md`](docs/designs/0001-language-spec.md)
- 文档索引与「设计 → RFC → SPEC → 计划」流程：[`docs/README.md`](docs/README.md)
- Go 重写过程与判据：[`docs/rfcs/0003-go-rewrite.md`](docs/rfcs/0003-go-rewrite.md)

## 三条命令跑起来

```text
aila new demo      # 建工程（生成 aila.toml + main.aila）
aila doctor        # 环境自检（缺什么，直接把要执行的命令打印出来）
aila run           # 编译 + 链接 + 运行
```

从源码构建管理器（无 cgo，可交叉编译）：

```bash
go build -o bin/aila ./cmd/aila
GOOS=linux GOARCH=amd64 go build -o bin/aila-linux ./cmd/aila
```

## 工具链从哪来（默认：只看随包目录）

发布包里自带 **llvm-mingw**（clang + lld + MinGW-w64 运行时一体，约 300 MB，压缩后 ~85 MB），
所以 Windows 上**不需要另装 MinGW / Clang / LLVM**：`aila build` 用的就是包内那一份。
（角色分工与「为什么是 MinGW 而不是官方 LLVM」见 [specs/0003 §1.1](docs/specs/0003-distribution-and-toolchain.md)。）

发布形态是 zip 解压即用，工具链**随包分发**，解析规则固定：

```text
<AILA_HOME>/toolchain/<triple>/bin/{clang, llc, ...}   ← 默认只用这个
        ▲
        └─ 只有清单显式写 [build.toolchain] cc = "system:clang>=18" 时才去系统 PATH 找
```

| 清单取值 | 行为 |
|---|---|
| 不写 / `"auto"` | 随包优先；本机没有随包时回退到系统 `clang`/`llc`，并在诊断里标注「开发回退」 |
| `"bundled"` | 只认随包，缺了就报 `E8010`/`E8011`（**不偷偷换**） |
| `"system:<名字>[>=版本]"` | 只在 PATH 找；版本不足 = 不认 |

装包与管理：

```text
aila toolchain install aila-0.1.0-windows-x64.zip   # 解到 toolchain/<triple>/
aila toolchain info                                 # 看每个 role 实际用了谁
aila toolchain list                                 # 看已装了哪些 bundle
aila version                                        # 三行：版本 + AILA_HOME + 后端/编译器来源
```

版本绑定：AILA 发布绑定一段 LLVM 版本带（当前 `18..23`），随包 `VERSION` 落在带外**不被采用**并给 `W8012`。

## 需要什么环境

| 件 | 用途 | 最低要求 |
|---|---|---|
| Go 1.22+ | 只用来**构建** `aila` 本身（产物是单文件，终端用户不需要 Go） | 必需（开发） |
| LLVM 后端 | IR → 目标 `.o` | 随包 `clang`（llvm-mingw）；开发期可装 clang/llvm 或跑 `ci/fetch-llvm.*` |
| C 编译器/链接器 | 编运行时与 `import cc`，并做最终链接 | 随包 `clang`（带 lld 与 MinGW 运行时）；开发期 `clang`/`gcc` |
| Clang | `import ch` 的 bindgen 首选档 | 随包自带；缺失时降级 builtin 兜底档 |

`aila doctor` 逐项测并用「链接冒烟测试」验证 IR → `.o` → 链接 → 运行整条链路。

## 常用命令

```text
aila new <目录> [--board host|esp32s3|riscv32|aarch64]
aila doctor [--json] [--cc ...] [--llvm ...]
aila check [路径]
aila build [路径] [--emit-ir] [--emit-json] [--out 名字] [--board <board>] [--opt 0-3]
aila run   [路径] [--arg 参数]
aila parse <文件.aila> [--tokens]      # 语法树 / token 流（排查用）
aila ir    <文件.aila> [--triple ...]  # 单文件降到 LLVM IR 并打印
aila toolchain info | list | install <zip>
aila target list | show <board>
aila explain <错误码>
aila clean [--cache-only]
aila version
```

### `build` 与 `run` 的分工（编译型语言的两条正路）

| 命令 | 做什么 | 产物 | 什么时候用 |
|---|---|---|---|
| `aila build` | 全流程到产物：IR → `.o` → 链接 | `build/<board>/app[.exe]`（host） | CI、烧录、交叉编译、要给别人的二进制 |
| `aila run` | `build` 之后**再执行**一次 | 无（复用 build 产物） | 本地改完就想看结果 |

`build` 不执行程序；`run` = `build` + 执行，并把程序的退出码原样透传。
`--emit-ir` 落 `.ll` 便于审计，`--emit-json` 写构建报告，`--out` 改名，`--board` 换目标。

退出码：`0` 成功 / `1` 有诊断 / `2` 用法错误 / `3` 环境不满足 / `4` 内部错误。

## 目录

```text
docs/       设计稿（designs）、提案（rfcs）、规格（specs）、计划（plans）——每类目录独立编号，从 0001 起
cmd/aila/   管理器 + 编译器前端入口（Go 单文件产物）
internal/   cli / home / diag / targets / resolve / toolchain / manifest /
            lexer / ast / parser / types / sema / codegen / bindgen / build / project
share/aila/data/  共享数据表（错误码 `diag_codes.json`、target 表 `targets.json`）
stdlib/     标准库（**纯 AILA 实现**，当前含 io / math；见 rfcs/0005）
runtime/    语言运行时（C 薄层：区域账本 + trap 通道 + 打印出口）
examples/   六个可运行示例（hello / control / region / struct / result / cinterop）
tests/corpus/  黄金语料与契约快照（token / AST / 布局 / 签名 / 绑定 / IR / 示例输出）
tools/goldengen/  快照生成器（`go run ./tools/goldengen [-check]`）
ci/         打包脚本（拉 llvm-mingw 并裁剪到 toolchain/、拼发布目录）
editors/vscode/  VSCode 扩展（语法高亮 / 片段 / 语言服务器客户端 / 构建运行命令）
toolchain/  随包工具链（ci/fetch-llvm.* 生成，不进版本库；发布时随 zip 分发）
```


## 跨项目复用（包管理）

**包 = 项目**（一个 `aila.toml` 的目录）、**子包 = 目录**、**模块 = 文件**。
要用另一个项目的代码，就在清单里声明依赖，然后按「依赖名.子包.模块」导入：

```toml
# 我的工程 aila.toml
[dependencies]
mylib = { path = "../mylib" }                        # 本地路径
util  = { git = "https://…/util.git", tag = "v1.2.0" }  # git（tag 或 rev 锁定，不许分支）
```

```aila
import mylib.math.ops as ops;    // 依赖名 mylib · 子包 math · 模块 ops

func main() {
    io.print_i32_nl(ops.twice(21));
}
```

要**导出**函数给别的子包/项目用，写「壳内前向声明 + 块内实现」（壳内 `public` = 包外可见）：

```aila
public func twice(v: I32) -> I32;        // 壳：接口 + 可见性

module {
    func twice(v: I32) -> I32 { return v + v; }   // 块：实现
}
```

规则速查：

| 项 | 说明 |
|---|---|
| 可见性 | 三级阶梯：文件内 ⊂ 子包内（块内 `public` / 壳内默认）⊂ 包外（壳内 `public`）；违反 = `E7020` |
| 锁文件 | `aila.lock` 记来源 / 版本 / 内容哈希；依赖变了不 `aila deps --update` 就报 `E7023`（构建可复现优先） |
| 离线构建 | `aila deps --vendor` 把依赖复制进 `vendor/`，此后无网络、原路径断开也能编 |
| 查看依赖 | `aila deps`（人读）/ `aila deps --json`（机器可读） |
| 传递依赖 | 依赖包自己的 `[dependencies]` 继续生效，按拓扑序编译 |
| C 包 | `import ch`/`import cc` 只在根包的 `cpacks/` 里解析；依赖包里的 C 导入暂不支持 |

完整规格见 [`docs/rfcs/0004`](docs/rfcs/0004-package-dependencies.md)。


## 标准库与运行时边界

标准库遵循「**能用 AILA 写的就用 AILA 写**」（[`rfcs/0005`](docs/rfcs/0005-pure-aila-stdlib.md)）：
十进制转换、符号、UTF-8 编码、浮点小数位、换行与组合函数全在 `stdlib/io.aila` 里，不走 libc 的 `printf`。
只有真正跨平台的 OS 边界（写标准输出）交给 C，以两条最小 `#[c]` 绑定出现：

```aila
#[c] func aila_rt_write(bytes: [U8], n: USize);
#[c] func aila_rt_flush();
```

各平台差异（`fwrite` / `write(2)` / ESP-IDF）由 `runtime/aila_rt.c` 消化；运行时里只剩这三类东西：
输出边界、区域账本、`aila_rt_trap` 陷阱出口。换平台只改运行时，标准库不动。
**不为它新增语言机制**——`#[c]` 已经够用（曾试过 `#[intrinsic]`，已删除）。

## 编辑器支持

`editors/vscode/` 是 AILA 的 VSCode 扩展。**语义能力来自编译器本体**（`aila lsp` 语言服务器）：
跳转定义、类型定义、跳到实现（壳内前向声明 ↔ 块内实现）、`import` 链接、查找引用、悬停签名、
大纲、工程符号搜索、诊断标红、补全、签名帮助、重命名、折叠、选区、格式化、内联类型提示；
语法高亮 / 括号 / 片段是纯声明式贡献点。
高亮覆盖 `internal/lexer` 的全部关键字 / 内置类型 / 标点，并由 `scripts/grammar-coverage.mjs`
保证「仓库里每个 token 都有 scope」。
另有 `AILA: 构建 / 构建并运行 / 只做检查 / 环境自检 / 生成 C 绑定 / 重启语言服务器` 命令与 `aila` 任务类型。
装法见 [`editors/vscode/README.md`](editors/vscode/README.md)（软链进 `~/.vscode/extensions` 或 `vsce package`）。


## `.aila` 与 `.ail` 不是一种文件

| 扩展名 | 是什么 | 谁生成 | 进版本库 |
|---|---|---|---|
| `.aila` | **AILA 源码**（模块） | 人写 | 是 |
| `.ail` | **C 绑定声明**（bindgen 产物，形如只有壳声明的 AILA 模块） | `aila build` 自动生成，也可 `aila bindgen` 单独产出 | 否（`build/` 下，随构建重生成） |

调 C 库**不需要手动绑定**：在 `cpacks/<名字>/` 放好 `include/*.h` 与 `.c`，然后

```aila
import ch _.cpacks.mathkit;   // 声明导入：自动对该 cpack 的头跑 bindgen
import cc _.cpacks.legacy;    // 实现导入：只编译链接，不产生声明
```

`aila build/run/check` 会自动跑 bindgen（缓存键 = 头文件哈希 + 档位 + 三元组），
想单独看产物就 `aila bindgen`（默认写到 `build/bindgen/<cpack>.ail`），
或 `aila build --emit-bindings` 把 `.ail` 一并放进产物目录。

## 当前实现范围（M1）

已落地：**全部 v0 语法解析** + 名字/类型/布局/区域效果/match 完备性/主名字检查 + LLVM IR 生成与目标文件产出，
覆盖标量、struct、enum、函数与方法、`if`/`for`/`match`、`T?`、`T ! E` 与 `?`、多返回值、`defer`、`view`、
`unsafe`、`region` 与 `@` 子句族、`#[c]`、`import ch`/`import cc`（bindgen 两档后端）。

未落地（按计划批次，见 [`docs/plans/0001`](docs/plans/0001-manager-frontend-dev-plan.md)）：
协程/通道/线程的代码生成、热替换 MLO 与分派表、借用区间（NLL）精确检查、完整 std。
**这些未落地项不会静默通过**——语义检查仍在，代码生成会明确报 `E9002`；未落地的检查项以 `W9001` 显式列出。

## 开发

```bash
go vet ./... && go build ./... && go test ./...
go run ./tools/goldengen            # 重新生成黄金快照
go run ./tools/goldengen -check     # 只校验快照是否与实现一致（CI 用）
```

测试面覆盖：词法/语法（token 与 AST 逐字节）、语义错误码与绑定类型、布局与 C `sizeof` 实测对齐、
C 绑定映射、生成 IR 的逐字节快照（含六个示例的每模块 IR）、六个示例的端到端运行输出。

## 许可证

AILA 以 [Apache License 2.0](LICENSE) 发布，版权声明与第三方组件说明见 [`NOTICE`](NOTICE)。

发布包内的工具链组件按各自许可证再分发：LLVM/Clang 为
`Apache-2.0 WITH LLVM-exception`（原文随包放在 `toolchain/<triple>/LICENSE.TXT`）。