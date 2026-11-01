# SPEC 0003：分发形态与随包工具链

> **编号**：0003 · **类型**：SPEC · **状态**：定稿（M2-D 实现依据）
> **上游**：[designs/0001](../designs/0001-language-spec.md) §12.4/§12.5（工具链供给与分发）、
> [rfcs/0001](../rfcs/0001-manager-and-frontend.md)（三级供给）、[rfcs/0003](../rfcs/0003-go-rewrite.md)（Go 重写）
> **实现**：`internal/home`（相对目录）、`internal/resolve`（解析）、
> `internal/toolchain`（动作 + doctor）、`internal/cli`（`aila toolchain`）、`ci/`（打包）

---

## 1. 发布物形态：解压即用的相对目录

发布物是一个 zip，解开后**全部相对定位**，不写注册表、不改 PATH、不需要管理员权限：

```text
aila-<version>-<platform>/
  aila[.exe]                    ← 管理器（Go 单文件）
  share/aila/runtime/aila_rt.c  ← 语言运行时（C 薄层，宿主）
  share/aila/runtime/<家族>/    ← 平台 runtime：startup.S / aila_rt_<家族>.c / chip.json
                                   （esp32 系另有 boot.c；esp8266 另有 nosdk.c——SDK 模式入口层；
                                     链接脚本与 *_cfg.h 由 chip.json 在构建时生成——嵌入式 build
                                     自动取用，项目里零平台文件）
  share/aila/stdlib/*.aila      ← 标准库（AILA 源码）
  share/aila/data/*.json        ← 错误码目录与全局配置（版本带等；board 表已退场，rfcs/0006）
  share/aila/skills/aila/       ← 给 AI 编程工具的技能包（目录直接登记进 AI 工具）
  toolchains/
    toolchains.json             ← 注册表：installed（aila 写）+ available（随包内置，rfcs/0006 §4）
    clang/<ver>/                ← 通版 clang（llvm-mingw）：bin/ lib/ …
    esp-clang/<ver>/            ← 特化 clang 家族（乐鑫 esp-clang 等）
    gcc/<triple>/<ver>/         ← 目标 GCC（目录名即目标三元组）
    sdk/<name>/<ver>/           ← 厂商 SDK（如 esp8266-nosdk = NonOS SDK；zip 按 Keep 白名单裁剪，
                                   如 include/ lib/ ld/ bin/）
  VERSION  NOTICES.md  README.md  LICENSE
```

注册表只记「下载清单」，不记能力——能力一律运行时探测（`clang --print-targets`、
`-print-target-triple`、GCC 目录名即 triple）。分发包内置 `available` 清单；
`install` 联网下载（用户显式触发）或装本地 zip。

技能包是纯文本目录（`SKILL.md` + `reference/*.md`，见 specs/0001 §6.7）：
不装 AILA 也能整目录拷给 AI 工具，所以不再单独打 zip。

### 1.0 开发仓库里的代码 / 产物分界（`.local/`）

分发布局的 `toolchains/` 是**随包只读**的。开发仓库不能这么放东西，否则源码树会被
几百 MB 的产物糊住，所以开发侧一律写进 `.local/`（整体进 `.gitignore`）：

```text
<AILA_HOME>/.local/
  bin/aila[.exe]              ← 管理器二进制（go build -o .local/bin/aila ./cmd/aila）
  build/<工程路径>/<triple>/   ← 该工程的 IR / .o / 可执行体 / 构建报告
  build/<工程路径>/.cache/     ← bindgen 缓存（含 LSP 侧）
  toolchains/                 ← `aila toolchain install` 的落点
    installed.json            ← 装了什么（版本 + 相对路径）
    clang/<ver>/  gcc/<triple>/<ver>/  sdk/<name>/<ver>/
```

- **清单**（能下载什么）仍在 `toolchains/toolchains.json`（源码树，进版本库）；
  **安装状态 + 载荷**在 `.local/toolchains/`。两者分离，`install` 绝不写源码树。
- 工具链发现同时扫 `.local/toolchains`（先）与随包 `toolchains/`（后）——分发包里
  `clang/` 与 `toolchains.json` 同目录，开发仓库里则是「清单在树上、载荷在 .local」。
- 工程在 `AILA_HOME` 外（独立项目）时，产物根是**该工程自己的** `.local/`。
- `aila clean` 清的就是该工程的产物目录。

### 1.1 三个角色，别混为一谈（常见误解）

「AILA 用 LLVM + Clang」说的是**后端**，不是全部。一次 `aila build` 需要三件东西：

| 角色（role） | 干什么 | 典型提供者 | 缺了会怎样 |
|---|---|---|---|
| **LLVM 后端**（`llc` / `clang -c`） | 把前端产出的 **LLVM IR 文本**降成目标文件 `.o` | 随包 `llc`/`clang`；系统 llc/clang | `E8010`（IR 已落在 `build/` 里） |
| **C 编译器 + 链接器**（`clang` / `gcc`） | 编 `runtime/aila_rt.c` 与 `import cc` 的 cpack 源；把全部 `.o` **链接**成可执行文件 | 随包 clang；系统 clang/gcc | `E8011` |
| **C 运行时 / 系统导入库（决定 ABI）** | 提供 `crt2.o`、`libmingw32`、`msvcrt`（gnu ABI）或 MSVC CRT（msvc ABI） | **MinGW-w64**（w64devkit 的 gcc，或 llvm-mingw 自带的 MinGW 运行时）；MSVC 侧为 Windows SDK | 链接期报未定义符号（如 `_GLOBAL_OFFSET_TABLE_`） |

**所以 MinGW 不是 LLVM 的替代品，而是 Windows 上 `*-w64-windows-gnu` 这个 ABI 的「C 运行时 + 链接器 + 系统导入库」提供者。**
Clang 只是**编译器前端**：它能把 C 编成 `.o`，但**不带** Windows CRT/import 库，也不自己实现链接
（`lld` 是链接器，仍需要那些库才能产出 `.exe`）。这正是缺陷 11 的由来：三元组写成 `x86_64-pc-windows-msvc`、
却用 MinGW 系链接器，会缺 `_GLOBAL_OFFSET_TABLE_`；统一到 `x86_64-w64-windows-gnu` 后两边才对得上
（见 [rfcs/0002 缺陷 11](../rfcs/0002-design-defects.md)）。

两条自洽的路线（二选一，**不要混**）：

| 路线 | 随包内容 | 三元组 | 说明 |
|---|---|---|---|
| **A. llvm-mingw（推荐）** | clang + lld + MinGW-w64 运行时与头文件（llvm-mingw 发行版自带） | `x86_64-w64-windows-gnu` | 单包自包含、解压即用；AILA 现在的默认三元组就是按它定的 |
| B. 官方 LLVM + 系统 MinGW | clang + lld（官方 win64 发行版**不含** MinGW 运行时） | 同上 | 用户机必须另有 MinGW-w64（如 w64devkit）；缺了 `aila doctor` 报 `E8011` |

> **现状（已落地）**：`ci/fetch-llvm.{ps1,sh}` 走的就是**路线 A（llvm-mingw）**：
> 下载 → 解包 → 按固定规则裁剪到 `toolchains/clang/<ver>/`（CI 脚本输出布局按 rfcs/0006 的更新待批次，当前脚本仍写旧 `toolchain/<triple>/`）。裁剪规则（脚本即规格）：
>
> | 保留 | 丢掉的 |
> |---|---|
> | `bin/` 里的 clang 系（`clang`/`clang++`）、`lld`/`ld.lld`、`llvm-{ar,nm,objcopy,objdump,ranlib,readobj,strip,windres}` 与小包装器，以及 **clang 自己依赖的 `libLLVM-*.dll`、`libclang-cpp.dll`、`libc++.dll`、`libunwind.dll`、`libwinpthread-1.dll`** | `lldb*`、`python*`、`clangd`/`clang-tidy`/`clang-format`/`scan-build`、sanitizer 动态库、`libomp.dll` |
> | `lib/clang/<ver>/include`（内建头）+ `lib/clang/<ver>/lib/windows/` 里的 `*builtins*`/`*profile*`（仅宿主 x86_64） | `lib/clang/<ver>/lib/linux` 与其它架构、fuzzer/ubsan/asan 等运行时 |
> | `include/`（C 头，sysroot）+ `x86_64-w64-mingw32/lib/`（CRT、导入库、运行时） | `include/c++`、`ddk`、`gdiplus`、`GL`、`KHR`、`wrl`、`x86_64-w64-mingw32/{bin,share}` |
> | `LICENSE.TXT`（版本即安装目录名，doctor 按 `--version` 自报判版本带） | `share/`、`busybox/`、`python/` |
>
> 实测（llvm-mingw 20261006 / LLVM 23.1.3 / ucrt）：裁剪后 **313 MB**，压缩包 **85 MB**；
> 解压到与仓库无关的目录、`PATH` 只留 `C:\Windows\System32;C:\Windows` 时，
> `aila version/doctor/run/new/build` 全部可用，六个示例全部编译并跑出与快照一致的输出。
>
> **踩过的坑（写进脚本注释）**：clang.exe 是用 libc++ 构建并动态链接的，丢掉 `libc++.dll`/`libunwind.dll`/
> `libwinpthread-1.dll` 会让它直接 `STATUS_DLL_NOT_FOUND`（0xC0000135）；`lib/clang/<ver>/lib/windows/` 里
> **只留 builtins/profile 的宿主架构**，别的一律不要。

`AILA_HOME` 解析顺序（`internal/home`）：

| # | 来源 | 说明 |
|---|---|---|
| 1 | 环境变量 `AILA_HOME` | CI / 测试 / 便携部署显式指定 |
| 2 | 仓库布局 | `<repo>/go.mod` 所在目录上溯，存在 `runtime/` 或 `stdlib/` |
| 3 | 可执行文件同级 | `sys.executable`（Go 版为 `os.Executable()`）所在目录存在 `share/aila/` |

布局探测同时接受 `share/aila/{runtime,stdlib}`（打包布局）与 `{runtime,stdlib}`（仓库布局），
因此同一份代码既能跑仓库、也能跑发布包。

## 2. 工具链解析：**默认只用随包，外找必须显式**

清单 `[build.toolchain]` 两个键（`aila.toml`）：

```toml
[build.toolchain]
cc   = "auto"            # auto（默认）| bundled | system:<名字>[>=版本]
llvm = "auto"            # auto（默认）| bundled | system:<名字>[>=版本] | llvmlite
```

| 取值 | 行为 | 缺失时 |
|---|---|---|
| `auto`（默认） | ①随包 `toolchains/`（家族发现 + 运行时探测，rfcs/0006 §5）→ ②开发期回退（系统 clang·gcc），并在诊断里标注「开发回退」 | `E8010` / `E8011`，附装包命令 |
| `bundled` | **只**认随包目录，绝不外找 | `E8010`/`E8011`，提示 `aila toolchain install` |
| `system:<名>[>=版本]` | 只在 PATH 找；版本低于要求 = 不认（不降级） | 报错并说明当前版本与要求 |
| `llvmlite` | 仅 IR 后端：进程内 llvmlite（开发用） | `E8010` |

**为什么默认不"先查 PATH"**：同一工程在不同机器上行为不同 = 分发事故源
（[designs/0001 §5.9](../designs/0001-language-spec.md) 同一条纪律）。发布包的 `auto` 一定命中随包工具，
所以**用户永远不感知工具链存在**；只有高级用户主动写 `system:...` 才会离开随包目录。

**版本绑定**（§12.5）：AILA 发布版本绑定一段 LLVM 版本带（当前 `18..23`；registry 可按家族覆盖）。
clang 系安装的版本带判定按其工具 `--version` 自报（版本即安装目录名）；带外 → 该工具**不被采用**
（`ok=false`），诊断给出版本带区间，避免「IR 语法/行为随 LLVM 版本漂移」这类静默事故。
GCC 没有 LLVM 版本带（永远带内）。

解析结果对所有命令可见：

```text
aila version           # 三行：版本 + AILA_HOME + 后端/编译器来源
aila toolchain info    # 逐项列出 role / 来源 / 版本 / 路径 / 版本带
aila doctor            # 自检表（含来源与第 6 项链接冒烟测试）
```

## 3. `aila toolchain` 命令

| 命令 | 作用 | 退出码 |
|---|---|---|
| `aila toolchain info` | 打印解析结果（AILA_HOME、宿主三元组、版本带、每个 role 的最终选择） | 0 |
| `aila toolchain list` | 列出已装家族/版本（扫描 `toolchains/`）+ available 清单（提示 `install <键>@<版本>`） | 0 |
| `aila toolchain install <键>[@版本]` | 按 available 清单联网下载安装（`AILA_TOOLCHAIN_MIRROR` 可换镜像前缀，仅对 github URL 生效） | 0 / 1（E7012/E7013）/ 2（用法） |
| `aila toolchain install <本地.zip> [--family F] [--triple T] [--force]` | 装本地 zip：sha256 反查 available 得家族/版本/triple（flags 兜底）；自动剥单层根目录；POSIX 补 `+x`；校验布局 | 同上 |

`install` 只在用户显式给出 `<键>` 时才联网；本地 zip 安装不联网。

## 4. CI 流水线（GitHub Actions）

```text
ci/fetch-llvm.{ps1,sh}     拉取官方 LLVM → 裁剪 → toolchains/clang/<ver>/（脚本按 rfcs/0006 布局的更新待 CI 批次，现仍写旧 toolchain/<triple>/）
ci/assemble-bundle.{ps1,sh} 拼发布目录 → dist/aila-<ver>-<platform>.zip
.github/workflows/release.yml  矩阵：windows-x64 / linux-x64 / macos-arm64 → 挂到 GitHub Release
```

裁剪规则（固定，不随心情变）：

| 保留 | 丢弃 |
|---|---|
| `bin/{clang, clang-cl, llc, lld-link, ld.lld, llvm-ar, llvm-objcopy, llvm-objdump, llvm-nm, llvm-strip}` | `bin/` 里其余工具（clangd、clang-tidy、llvm-mc、opt、llvm-as…） |
| `lib/clang/<ver>/include`（clang 资源目录） | `docs/`、`share/`、`examples/`、`tools/`、`include/`（开发头） |
| 必要的 `libclang*` / `libLLVM*` 动态库 | 静态库与开发期 `.a` |
| `LICENSE.TXT` | — |

各平台来源（无官方通用二进制，按平台取最稳的一手）：Windows = llvm-project release 的 `LLVM-*-win64.exe`（7z 展开）；
Linux = apt.llvm.org 的 deb 解包；macOS = Homebrew `llvm@<ver>` 后拷目录。

**首次 release 前必须做**：①填 sha256 校验（脚本里留了 TODO）；②在三个平台各跑一次
`aila doctor`，第 6 项「链接冒烟测试」通过才算包可用。

## 5. 验收

| # | 检查 | 命令 |
|---|---|---|
| 1 | 仓库内（无安装）解析为「开发回退」且 `doctor` 只有 warn | `aila doctor` |
| 2 | 装一份工具链后解析切到「随包」 | `aila toolchain info`（`internal/toolchain` 测试已锁定） |
| 3 | `bundled` 缺包时报错且不回退 | 清单写 `llvm = "bundled"` → `E8010` |
| 4 | `system:x>=版本` 版本不足 = 不认 | `internal/resolve::TestSystemVersionFloorRejects` |
| 5 | 装包后布局正确（家族/版本目录 + 注册表 installed 记录） | `aila toolchain install …` + `aila toolchain list` |
| 6 | 发布目录里 AILA_HOME 指向解压根、std 与 runtime 均可见 | `aila doctor` 首行 |
