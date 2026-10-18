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
  aila[.exe]                    ← 管理器（Go 单文件；移植完成前为 Python 启动器 + src/）
  share/aila/runtime/aila_rt.c  ← 语言运行时（C 薄层）
  share/aila/stdlib/*.aila      ← 标准库（AILA 源码）
  share/aila/data/*.json        ← 错误码目录与 board 表
  share/aila/skills/aila/       ← 给 AI 编程工具的技能包（`aila skill` 读它）
  toolchain/<triple>/
    bin/{clang, llc, ld.lld, llvm-ar, llvm-objcopy}   ← 随包 LLVM/Clang（裁剪）
    lib/clang/<ver>/include/                          ← clang 内建头（bindgen / C 编译需要）
    lib/libclang* | libLLVM*                          ← 必要动态库（平台相关）
    VERSION                                            ← 该 bundle 的 LLVM 版本
    LICENSE.TXT                                        ← Apache-2.0 WITH LLVM-exception
  VERSION  NOTICES.md  README.md  LICENSE
```

另有一份 **`aila-skill-<version>-<platform>.zip`**：只含 `aila/`（`SKILL.md` + `reference/*.md`），
不含任何二进制。用途是「不装 AILA 也能把这份技能喂给 AI 工具」——技能讲的是怎么用 `aila` 命令，
所以它必须能独立分发（见 specs/0001 §6.7）。

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
> 下载 → 解包 → 按固定规则裁剪到 `toolchain/<triple>/`。裁剪规则（脚本即规格）：
>
> | 保留 | 丢掉的 |
> |---|---|
> | `bin/` 里的 clang 系（`clang`/`clang++`）、`lld`/`ld.lld`、`llvm-{ar,nm,objcopy,objdump,ranlib,readobj,strip,windres}` 与小包装器，以及 **clang 自己依赖的 `libLLVM-*.dll`、`libclang-cpp.dll`、`libc++.dll`、`libunwind.dll`、`libwinpthread-1.dll`** | `lldb*`、`python*`、`clangd`/`clang-tidy`/`clang-format`/`scan-build`、sanitizer 动态库、`libomp.dll` |
> | `lib/clang/<ver>/include`（内建头）+ `lib/clang/<ver>/lib/windows/` 里的 `*builtins*`/`*profile*`（仅宿主 x86_64） | `lib/clang/<ver>/lib/linux` 与其它架构、fuzzer/ubsan/asan 等运行时 |
> | `include/`（C 头，sysroot）+ `x86_64-w64-mingw32/lib/`（CRT、导入库、运行时） | `include/c++`、`ddk`、`gdiplus`、`GL`、`KHR`、`wrl`、`x86_64-w64-mingw32/{bin,share}` |
> | `LICENSE.TXT`、`VERSION`（首段是 LLVM 版本，doctor 按它判版本带） | `share/`、`busybox/`、`python/` |
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
| `auto`（默认） | ①随包 `toolchain/<triple>/bin` → ②开发期回退（`llvmlite` / 系统 clang·gcc），并在诊断里标注「开发回退」 | `E8010` / `E8011`，附装包命令 |
| `bundled` | **只**认随包目录，绝不外找 | `E8010`/`E8011`，提示 `aila toolchain install` |
| `system:<名>[>=版本]` | 只在 PATH 找；版本低于要求 = 不认（不降级） | 报错并说明当前版本与要求 |
| `llvmlite` | 仅 IR 后端：进程内 llvmlite（开发用） | `E8010` |

**为什么默认不"先查 PATH"**：同一工程在不同机器上行为不同 = 分发事故源
（[designs/0001 §5.9](../designs/0001-language-spec.md) 同一条纪律）。发布包的 `auto` 一定命中随包工具，
所以**用户永远不感知工具链存在**；只有高级用户主动写 `system:...` 才会离开随包目录。

**版本绑定**（§12.5）：AILA 发布版本绑定一段 LLVM 版本带 `VERSION_BAND`（当前 `18..23`）。
随包 `VERSION` 落在带外 → 该工具**不被采用**（`ok=false`），并给出 `W8012` 警告，
避免「IR 语法/行为随 LLVM 版本漂移」这类静默事故。

解析结果对所有命令可见：

```text
aila version           # 三行：版本 + AILA_HOME + 后端/编译器来源
aila toolchain info    # 逐项列出 role / 来源 / 版本 / 路径 / 版本带
aila doctor            # 自检表（含来源与第 6 项链接冒烟测试）
```

## 3. `aila toolchain` 命令

| 命令 | 作用 | 退出码 |
|---|---|---|
| `aila toolchain info` | 打印解析结果（AILA_HOME、三元组、版本带、每个 role 的最终选择） | 0 |
| `aila toolchain list` | 列出已安装的 bundle 目录、版本、可执行文件数 | 0 |
| `aila toolchain install <zip> [--triple T] [--force]` | 把官方 release zip 解到 `toolchain/<T>/`；自动剥掉 zip 单层根目录；POSIX 上补 `+x`；校验 `bin/clang` 存在 | 0 / 1（E7012/E7013）/ 2（用法） |

`install` **不联网**：包从哪来由用户/安装器决定（发布页、离线介质、企业内网），
管理器只负责解包与校验布局——把网络策略留给调用方，是「不加隐性行为」的同一纪律。

## 4. CI 流水线（GitHub Actions）

```text
ci/fetch-llvm.{ps1,sh}     拉取官方 LLVM → 裁剪 → toolchain/<triple>/
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
| 1 | 仓库内（无 bundle）解析为「开发回退」且 `doctor` 只有 warn | `aila doctor` |
| 2 | 造一个 bundle 后解析切到「随包」 | `aila toolchain info`（`tests/test_toolchain.py` 已锁定） |
| 3 | `bundled` 缺包时报错且不回退 | 清单写 `llvm = "bundled"` → `E8010` |
| 4 | `system:x>=版本` 版本不足 = 不认 | `tests/test_toolchain.py::test_system_version_floor_rejects` |
| 5 | 装包后布局正确、`VERSION` 可读 | `aila toolchain install <zip>` + `aila toolchain list` |
| 6 | 发布目录里 AILA_HOME 指向解压根、std 与 runtime 均可见 | `aila doctor` 首行 |
