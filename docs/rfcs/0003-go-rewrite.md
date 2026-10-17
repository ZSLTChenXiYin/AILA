# RFC 0003：用 Go 重写管理器与前端的决定与分批计划

> **编号**：0003 · **类型**：RFC · **状态**：已完成（G1 ✅、G2 ✅、G3 ✅、G4 ✅：AILA 只剩 Go 一份实现，Python 前端与测试已删除）
> **上游**：[designs/0001](../designs/0001-language-spec.md) §12.4/§12.5、[rfcs/0001](0001-manager-and-frontend.md)、[specs/0003](../specs/0003-distribution-and-toolchain.md)
> **结论先行**：**是，用 Go 重写**；但**分批**（G1–G4），Python 版在过渡期作为**参考实现**保留，直到黄金测试全绿才删。

## 1. 判据：分发这件事只有三个硬指标

| 硬指标 | 含义 | 对语言的约束 |
|---|---|---|
| **单文件** | 用户下载一个 zip，解压即用；管理器本身不该依赖用户装运行时 | 要么自带运行时，要么静态链接 |
| **零系统依赖** | 不要求用户机上有 Python/LLVM/MSVC 运行库 | 不能有 cgo 之外的动态依赖；cgo 会引入 libLLVM |
| **可交叉编译** | CI 上一条命令出 windows/linux/macos 三平台产物 | 工具链必须自带跨平台目标 |

Python 用自己的解释器（`python.exe`）满足不了前两条；PyInstaller/Nuitka 能"打成 exe"，但代价是
单文件 20–40MB、启动 100–300ms、macOS 需签名公证、Windows 杀软误报高——**这是应急手段，不是长期形态**。

## 2. 方案对比

| 方案 | 单文件 | 零依赖 | 交叉编译 | ADT/模式匹配 | 判定 |
|---|---|---|---|---|---|
| **Go（采纳）** | ✅ 静态单文件 | ✅ 无 cgo（见 §3） | ✅ `GOOS/GOARCH` 一行 | ❌ interface + type switch | 采纳 |
| Rust | ✅ | ✅ | ✅（需 `rustup target add`，CI 稍重） | ✅ 极适合编译器 | 备选：语义表达更好，但团队/CI 门槛高，与「傻瓜式」冲突 |
| C++ + libLLVM | ✅ | ❌ 依赖系统运行库 | ❌ 三平台构建复杂 | ❌ | 弃 |
| Python 保持 + PyInstaller | ⚠️ 20–40MB | ⚠️ 自带解释器但仍庞大 | ⚠️ 需 per-OS 构建 | — | 仅作 P0 应急 |
| Python 保持原样 | ❌ 要用户装 Python | ❌ | — | — | 与发布目标冲突 |

**为什么 Go 足够**：前端只做「源码 → LLVM IR 文本」，**不链接 LLVM**；把 IR 变成 `.o` 交给随包 `llc`/`clang`
（`specs/0003 §2`）。因此 Go 版**完全不需要 cgo**，也就不需要 `libLLVM`、不需要 `llvm-config`——
这正好把"内嵌 LLVM"从"链接库"降级成"随包可执行文件"，而后者本来就要为 C 互操作随包分发。

代价（登记，不回避）：Go 没有 sum type / 模式匹配，AST、类型项、错误码表会比 Python 版更啰嗦；
用 `interface{} + type switch` 集中在一个 `internal/ast` 包里可把影响面控住。

## 3. 分批计划（每批都可独立验收）

| 批 | 范围 | 验收（必须可自动跑） |
|---|---|---|
| **G1 骨架 + 无前端命令** | `cmd/aila`、`internal/{cli,home,targets,resolve,toolchain,manifest,diag}`：`version / doctor / toolchain info·install·list / target list·show / explain / new / clean` | ✅ `go test ./...` 全绿；`aila-go doctor / toolchain info / explain` 与 Python 版同义 |
| **G2 前端下层** | `internal/{lexer,parser,ast}`：全 v0 语法 + 解析禁区诊断 | ✅ 7 份语料的 token 与 AST 双双逐字符对齐 |
| **G3 语义 + 代码生成** | `internal/{types,sema,codegen}`：布局、区域效果、match 完备性、IR 生成 | ✅ 布局（`layout.json`）、✅ 签名（`sigs.json`）、✅ 绑定类型（`bindings.json`）；IR 判据待 codegen |
| **G4 切换与清理** | CI 与 `ci/assemble-bundle.*` 改用 `-Manager go`；删 Python 前端；文档收口 | release 三平台产物 `aila doctor` 全 ✔，`examples/` 全跑通 |

顺序不可跳：G1 让"分发形态"先成立（管理器能跑、能自检、能装工具链），G2/G3 才动编译能力；
G4 之前**两条实现并存**，任何行为差异都以**黄金测试**判胜负，而不是"哪个看着对"。

## 4. 对齐策略：黄金测试（关键设计）

没有它，重写必然变成"边写边猜语义"。三层锁定：

1. **共享语料**：`tests/corpus/` 下每个 `.aila` 文件配二份期望输出——
   `*.tokens.json`（token 序列）、`*.ast.json`（去位置信息的 AST 树）、`*.ir`（IR 文本）。
   由 Python 版生成并**进版本库**（它是契约快照，不是产物）。
2. **共享数据表**：错误码目录、target 表从 Python 里**外置成 JSON**（`share/aila/diag_codes.json`、
   `share/aila/targets.json`），Go 用 `go:embed` 读同一份文件——从根上消灭"两套码表漂移"。
3. **双向跑**：CI 的 `ci.yml` 同时跑 Python 与 Go 的测试；差异即失败，不做"已知差异"白名单。

## 5. 影响面

| 面 | 变化 |
|---|---|
| 文档 | 本 RFC + [specs/0003](../specs/0003-distribution-and-toolchain.md)；`specs/0001/0002` 作为规格不变（Python 版是实现，不是规格） |
| CI | `ci.yml` 增加 `go build/test`（已有占位 job）；`release.yml` 的 `-Manager go` 在 G4 启用 |
| 目录 | 仓库根新增 `go.mod`、`cmd/aila`、`internal/`；Python 版留在 `src/`（G4 删除） |
| 风险 | ①Go 无模式匹配 → 用类型断言集中封装；②`.tokens/.ast` 快照会随语言演进变动 → 快照更新必须走 PR 评审；③G2/G3 期间双向维护成本 → 用批次数硬约束（G2/G3 各不超过两轮） |

## 6. 不做的事（本 RFC 划界）

- **不换后端模型**：Go 版同样只产 IR 文本，绝不 cgo 链 libLLVM（这是"单文件"的前提）；
- **不引入 Go 侧泛型抽象**：AST/类型用具体结构体，避免 `any` 满天飞；
- **不动语言语义**：本 RFC 只换实现语言，`designs/0001` 一个字不改；
- **不并行改 CLI 语义**：命令名、退出码、诊断格式以 [specs/0001](../specs/0001-manager-cli.md) 为准，Go 版逐字对齐。

## 7. G1 落地记录（实测）

| 项 | 结果 |
|---|---|
| 代码 | `cmd/aila`（入口）+ `internal/{cli,home,diag,targets,resolve,toolchain,manifest}`，依赖 cobra + BurntSushi/toml（无 cgo） |
| 已可用命令 | `version` / `doctor`（含链接冒烟测试）/ `toolchain info·list·install` / `target list·show` / `explain` / `new` / `clean` |
| 未接入命令 | `check` / `build` / `run` / `bindgen` → 明确报 `E9003` 并指回过渡期的 Python 版（G2/G3 后替换） |
| 共享数据表 | `share/aila/data/diag_codes.json`（80 个码）与 `targets.json` —— **Python 与 Go 读同一份**；Python 侧新增守卫测试「源码里出现的每个错误码都必须已登记」，本轮据此补回了 bindgen 漏登记的 `W3001/W3002/W3003`（原本是潜在崩溃点） |
| 测试 | `go test ./...`：diag / home / targets / resolve / manifest / cli 六个包全绿；Python 侧 71 条全绿 |
| 构建 | `go build -o bin/aila ./cmd/aila`；交叉编译 `GOOS=windows\|linux\|darwin go build ./cmd/aila` 均无 cgo 依赖 |
| 版本注入 | `-ldflags "-X github.com/aila-lang/aila/internal/cli.Version=<ver>"`（`ci/assemble-bundle.*` 已接好） |