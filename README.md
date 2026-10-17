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
| LLVM 后端 | IR → 目标 `.o` | 随包 `llc`/`clang`；开发期装 `clang`/`llvm` 并放进 PATH |
| C 编译器/链接器 | 编运行时与 `import cc`，并做最终链接 | 随包 `clang`；开发期 `clang`/`gcc` |
| Clang | `import ch` 的 bindgen 首选档 | 可选（缺失时降级 builtin 兜底档） |

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
stdlib/     标准库（AILA 源码，当前含 io / math）
runtime/    语言运行时（C 薄层：区域账本 + trap 通道 + 打印出口）
examples/   六个可运行示例（hello / control / region / struct / result / cinterop）
tests/corpus/  黄金语料与契约快照（token / AST / 布局 / 签名 / 绑定 / IR / 示例输出）
tools/goldengen/  快照生成器（`go run ./tools/goldengen [-check]`）
ci/         打包脚本（拉取裁剪 LLVM、拼发布目录）
```

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