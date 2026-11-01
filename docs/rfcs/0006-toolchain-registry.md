# RFC 0006：toolchains 注册表与目标三元组

> **编号**：0006 · **类型**：RFC · **状态**：已落地（2026-10-19，见 §8）
> **上游**：[specs/0003](../specs/0003-distribution-and-toolchain.md)（现行 per-triple bundle 模式）、
> [specs/0001](../specs/0001-manager-cli.md) §2.1（board 表）、[designs/0001](../designs/0001-language-spec.md) §12.4/§12.5、
> 2026-10-19 设计讨论（ESP8266 全链路实战后）
> **下游**：落地时回改 specs/0003 §1–4、specs/0001 §2/§2.1/§4

---

## 1. 背景与动机

现行模式（specs/0003）把工具链按**目标三元组**组织成 bundle——`toolchain/<triple>/bin/`，每个 bundle 视为「这个 triple 的一整套工具」。ESP8266 全链路实战（2026-10-19）暴露了它的结构性问题：

| 问题 | 实例 |
|---|---|
| 目录布局即语义，工具能力靠猜 | esp-clang 的 `llc` 只注册 Xtensa（无 x86），host IR 一旦被它接走就编不了；resolve 层为此打了「bundle 内按工具名顺序、不跨 bundle 抢同名工具」的补丁 |
| 版本不可共存 | 一个 triple 一个 bundle，换版本 = 整目录替换；无法 per-project 锁版本 |
| 厂商 fork 定位窄化 | esp-clang 实际支持 esp32 / esp32s3 / esp8266 / esp32c3 全系列，塞进 `toolchain/xtensa-esp8266-elf/` 等于把它当成单一 triple 的专用工具 |
| 全量分发体积失控 | esp-clang 仅 bin 目录就 911 MB，按 triple 全量预打包不合理 |

## 2. 目标与非目标

**目标**：

- 工具按「家族 + 版本」组织，多版本共存，可 per-project 锁定
- 能力（支持哪些 triple、承担哪些角色）**运行时探测**，不人工维护映射表
- 注册表只承担**下载清单**职责（family / 版本 / URL / SHA256），年维护量 = 厂商发版时加几行
- 清单 `[target]` 以 **triple 为主键**，board 概念彻底移除

**非目标**：

- **不做 GCC 代码生成**。GCC 没有 LLVM IR 输入通道（能吃的输入只有 C 源码 / 汇编源码 / 目标文件），「让 GCC 生成 .o」不存在轻量开关；给 AILA 加 C 后端是独立的编译器架构批次，与本 RFC 无关，且不解决任何现存痛点（ESP8266 链路已验证可出镜像）。相关路线对比：AILA→C 后端（成熟但工程大，Nim 路线）、自生成 Xtensa 汇编喂 `as`（最重）、llvm-cbe（实验性，不作承诺）。
- 不改变「默认只用随包，外找必须显式」的纪律（specs/0003 §2）。

## 3. 目录布局

```text
AILA_HOME/
  aila[.exe]
  share/aila/...
  toolchains/
    toolchains.json            ← 注册表（§4）
    clang/<version>/           ← 通版 LLVM/Clang（llvm-mingw 裁剪包）：ir2obj + cc + lld + bindgen
    esp-clang/<version>/       ← 厂商 fork：ir2obj（可支持的 targets 运行时探测）
    gcc/<triple>/<version>/    ← GCC 系：cc + GNU ld + libgcc；目录名 = triple（不带 -gcc 后缀：
                                 一个 GCC 目录里是 gcc+ld+objdump 一整套，语义是「这个 target 的 GCC」）
```

版本号与家族名来源：安装时由 `install` 命令写入注册表，目录名即事实。

## 4. toolchains.json：只记下载清单，不记能力

两个职责分两段：

```jsonc
{
  "installed": {                     // aila 本地写：已装什么、在哪（toolchain list 读它）
    "esp-clang": { "version": "22.1.4", "path": "toolchains/esp-clang/22.1.4" }
  },
  "available": [                     // CI 随包内置（来源见 §8）：能下载什么（install 读它）
    { "family": "clang",     "version": "23.1.3", "url": "...", "sha256": "...", "size": "85MB" },
    { "family": "esp-clang", "version": "22.1.4", "url": "...", "sha256": "...", "size": "911MB" },
    { "family": "gcc", "triple": "xtensa-lx106-elf", "version": "8.4.0", "url": "...", "sha256": "..." }
  ]
}
```

**能力信息一律探测，人工维护量为零**：

| 信息 | 获取方式 |
|---|---|
| clang 系支持哪些架构 | `clang --print-targets` 自报 |
| 某架构支持哪些 cpu | `--print-supported-cpus` 自报 |
| gcc 支持哪些 triple | 目录名即 triple，存在即支持 |
| 链接器归谁 | 跟 cc 家族（§5），零配置 |

## 5. 角色推导规则（内置，不是配置项）

| 角色 | 推导规则 | ESP8266 实际结果 |
|---|---|---|
| ir2obj | 在自报支持目标 triple 的 clang 系家族里选；**特化家族（esp-clang）优先于通版** | esp-clang 22.1.4，`-target xtensa-esp8266-elf -mcpu esp8266`（不显式传 target 会被覆盖成默认 riscv32-esp） |
| cc | 目标 triple 的 GCC 目录存在 → 用它；否则 clang 系 | `gcc/xtensa-lx106-elf/8.4.0` |
| ld | 跟 cc 家族：cc 是 GCC → GCC 驱动（GNU ld + libgcc）；cc 是 clang → lld | GNU ld（**lld 没有 Xtensa 后端**，已验证的硬约束） |
| bindgen | clang 系（目标 triple 可用者优先） | esp-clang |

混搭是常态不是例外：ESP8266 = esp-clang 编 IR + GCC 8.4 链接；host = llvm-mingw 一把抓。两个已验证场景都被同一套规则覆盖，新增任何厂商 fork 都不改规则。

**版本带**（designs/0001 §12.5）：per-family 校验。各家族自报 `llvm_version`（esp-clang 22.1.4 与 llvm-mingw 23.1.3 版本号不同源，不能共用一条带），落带外不采用（`W8012`），纪律不变。

## 6. 清单 `[target]`：triple 主键，board 移除

```toml
[target]
triple = "xtensa-esp8266-elf"   # 缺省 = 宿主（什么都不写 = host）
cpu    = ""                     # 缺省 = triple 第二段候选，探测验证后采用
```

triple 不携带的东西，每样给出路：

| 丢了什么 | 出路 |
|---|---|
| arch | triple 第一段拆出，不设独立字段 |
| cpu（`-mcpu=esp8266`） | 默认取 triple 第二段（esp 系芯片名即 cpu），`--print-supported-cpus` 验证通过即采用；验证不过且未显式写 = 报错提示补写 |
| domains 地形（§5.9 的 psram / rtc-slow） | 清单显式配置，默认 internal-ram（freestanding）；M1 的 domains 未落地，零损失 |
| embedded 判定（可执行 vs 到 .o 为止） | triple 后缀：`-elf/-none/-eabi` → freestanding；`gnu/linux/darwin` → host 语义 |

连带改动：

- `share/aila/data/targets.json` 的 board 表删除（版本带等全局配置并入 toolchains.json 或保留最小配置文件）
- `aila target show <triple>` 语义 = 拆 triple + 探测已装工具链报告能力（比查内置表更真实）
- triple 探测失败（拼写错 / 未装对应工具链）→ 报错列出已装工具链实际支持的 triple 候选，不甩「找不到」

## 7. 分发包矩阵

| 包 | 内容 | 适用 |
|---|---|---|
| 轻量（默认推荐） | host clang 全家桶（llvm-mingw 裁剪） | host 开发；交叉场景按需 `aila toolchain install <family>` 联网补齐 |
| 全量 | host 全家桶 + 主流嵌入式（esp 系 clang + 常用 triple 的 gcc） | 离线 / 内网环境 |

长尾平台一律走按需下载：available 清单给 URL / SHA256，`install` 下载并校验后写入新布局。specs/0003 §3「install 不联网」的纪律相应放宽为：**联网只发生在用户显式触发的 install**，离线安装本地 zip 的路径保留。

**镜像（GitHub 直连不通时）**：available 里的 URL 都指向 GitHub Releases，国内直连常年不通。`install` 因此支持镜像前缀，与 `ci/fetch-llvm.*` 的 `-Mirror` 同一约定（镜像拼在原 URL 前）：

```
aila toolchain install clang@23.1.3 --mirror https://gh-proxy.com
AILA_TOOLCHAIN_MIRROR=https://gh-proxy.com,https://ghfast.top aila toolchain install clang
```

- 可逗号分隔多个候选，**按序尝试，原始直连永远排最后兜底**；每个候选失败重试一轮。
- 只对 `https://github.com/` 开头的 URL 套镜像（esp 系 URL 是厂商 CDN，不需要）。
- 镜像反代常常先回源、首字节很慢，所以只限制「等响应头」90s 不限制整体时长，并每 32MB 报一次进度。
- 校验不放松：解出来的 zip 仍按清单里的 sha256 逐字节比对（实测 `gh-proxy.com` 取的 llvm-mingw 23.1.3 与清单 sha 一致）。
- 完全离线的路只有一条且永远有效：`aila toolchain install <本地 zip>`（sha 命中清单即自动取 family/version/triple）。

## 8. 落地记录（2026-10-19）

代码与文档已全部落地（specs/0001 §2/§2.1/§4、specs/0003 §1–4 已回改）。实现中的决策记录：

1. **available 清单随包内置**（开放点 2 的答案）：`toolchains/toolchains.json` 由 CI 随分发包给出；
   仓库侧同路径文件是它的源。esp-clang 原始分发是 .tar.xz，待 CI 打成 zip 后补 available 条目，
   当前用本地 zip 安装。
2. **cpu 推导 = triple 第二段 + 豁免表**（`unknown/none/generic/pc/w64/apple` 等），
   不做 `--print-supported-cpus` 探测：esp 系芯片名即 cpu 的约定已经覆盖现实场景，探测反而引入不确定性。
3. **宿主 triple 时通版 clang 优先于特化家族**：host 的 cc / ir2obj / 链接器须同家族同版本
   （MinGW CRT 配套），特化家族（如 esp-clang）只服务交叉目标。
4. **带外候选不采用，继续找**：`clangForIR2Obj`/`clangForCC` 逐候选校验版本带，
   全部落空时返回最后一个带外原因供诊断（`llc --version` 首行是 `LLVM (http://llvm.org/)`，
   版本号逐行找）。
5. **嵌入式 build 在 E8020 处提前收束**：IR→.o 之后即停，cc 解析与 runtime C 编译只在宿主路径发生
   （esp8266 的 runtime C 与链接由项目层 rt/build.ps1 承担，M2 批次收进 aila）。
6. **GCC 目录名 = 目标三元组是硬约定**：官方 GCC 的 triple（如 `xtensa-lx106-elf`）与 AILA 项目
   triple（如 `xtensa-esp8266-elf`）不同名时，交叉 cc 不参与（嵌入式路径不需要）；
   M2 收链时再决定别名或改名。
7. **CI 脚本（`ci/fetch-llvm.*`、assemble-bundle）按本 RFC 布局的更新待 CI 批次**，
   当前脚本仍写旧 `toolchain/<triple>/` 布局（specs/0003 §4 已注明）。

## 9. 落地时回改（已完成）

| 文件 | 章节 |
|---|---|
| specs/0003 | §1（目录布局）、§2（解析）、§3（toolchain 命令）、§4（CI 矩阵） |
| specs/0001 | §2（清单 `[target]`）、§2.1（board 表删除）、§4（探测顺序） |
| 代码 | `internal/registry`（新增）、`internal/resolve`（探测 + 推导重写）、`internal/targets`（board 表退场）、`internal/home`（ToolchainsRoot + 家族发现）、`internal/manifest`（triple/cpu）、`internal/build`（E8020 嵌入式收束）、`internal/toolchain`（doctor 冒烟）、`internal/cli`（target / toolchain / install 命令） |
