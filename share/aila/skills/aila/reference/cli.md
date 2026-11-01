<!-- 生成物：由 `go run ./tools/skillgen` 从命令树生成，勿手改；CI 用 -check 校验。 -->

# `aila` 命令参考

> 有疑问的命令随时自查：`aila <子命令> --help`。退出码：`0` 成功、`1` 有诊断、`2` 用法错误、`3` 环境不满足、`4` 管理器内部错误。

## `aila`

AILA 是面向嵌入式、以区域内存模型与模块级热替换为核心的编程语言。
这是用 Go 重写的管理器与编译器前端（rfcs/0003；G1–G3 已完成，G4 进行中）。

用法：`aila [command]`

参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila bindgen`

把 cpack 的 C 头文件翻成 `.ail` 绑定文件（机器生成，`.aila` 才是人写的源码）。
不给目标 = 生成项目 `cpacks/` 下的全部 cpack；给目标 = cpack 名（`math`）、相对路径（`cpacks/math`）或含头文件的目录。

用法：`aila bindgen [目标]`

参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--backend` | string | `auto` | bindgen 档位：auto \| clang \| builtin |
| `--out` | string | `—` | 输出文件（单 cpack 时可用） |
| `--triple` | string | `—` | 目标三元组（临时覆盖清单，影响 long/size_t 的映射） |

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila build`

全流程构建（IR → 目标文件 → 链接）

用法：`aila build [路径]`

参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--cc` | string | `—` | C 编译器 / 链接器选择（auto / bundled / system:...） |
| `--deps-update` | bool | `—` | 重新解析跨项目依赖并重写 aila.lock |
| `--emit-bindings` | bool | `—` | 把生成的 C 绑定 `.ail` 复制到产物目录 |
| `--emit-ir` | bool | `—` | 额外写出 `.ll` 文本 |
| `--emit-json` | bool | `—` | 写出构建报告 `aila-build.json` |
| `--emit-size-delta` | bool | `—` | 构建后报告产物与每个模块目标文件的大小（量化换驱动 / 关功能的体积差） |
| `--gc-sections` | bool | `—` | 函数/数据分段 + 链接期 --gc-sections：未被引用的模块代码不进产物（配合 #compile.match 选驱动） |
| `--llvm` | string | `—` | LLVM 后端选择（auto / bundled / system:...） |
| `--opt` | int | `0` | 优化级别（0-3，临时覆盖清单） |
| `--out` | string | `—` | 输出文件名（默认 `app` + 平台后缀） |
| `--triple` | string | `—` | 目标三元组（临时覆盖清单，如 xtensa-esp8266-elf） |
| `--vendor` | bool | `—` | 把依赖复制到 vendor/（离线构建） |

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila check`

只做检查（解析 + 语义），不产产物

用法：`aila check [路径]`

参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--cc` | string | `—` | C 编译器 / 链接器选择（auto / bundled / system:...） |
| `--deps-update` | bool | `—` | 重新解析跨项目依赖并重写 aila.lock |
| `--emit-bindings` | bool | `—` | 把生成的 C 绑定 `.ail` 复制到产物目录 |
| `--emit-ir` | bool | `—` | 额外写出 `.ll` 文本 |
| `--emit-json` | bool | `—` | 写出构建报告 `aila-build.json` |
| `--emit-size-delta` | bool | `—` | 构建后报告产物与每个模块目标文件的大小（量化换驱动 / 关功能的体积差） |
| `--gc-sections` | bool | `—` | 函数/数据分段 + 链接期 --gc-sections：未被引用的模块代码不进产物（配合 #compile.match 选驱动） |
| `--llvm` | string | `—` | LLVM 后端选择（auto / bundled / system:...） |
| `--opt` | int | `0` | 优化级别（0-3，临时覆盖清单） |
| `--out` | string | `—` | 输出文件名（默认 `app` + 平台后缀） |
| `--triple` | string | `—` | 目标三元组（临时覆盖清单，如 xtensa-esp8266-elf） |
| `--vendor` | bool | `—` | 把依赖复制到 vendor/（离线构建） |

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila clean`

清理产物

用法：`aila clean [路径]`

参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--cache-only` | bool | `—` | 只删构建缓存 |

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila deps`

查看/更新跨项目依赖（打印依赖图、写锁文件、生成 vendor）

用法：`aila deps [路径]`

参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--json` | bool | `—` | 机器可读输出 |
| `--update` | bool | `—` | 重新解析依赖并重写 aila.lock |
| `--vendor` | bool | `—` | 把依赖复制到 vendor/（离线构建） |

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila doctor`

环境自检（含链接冒烟测试）

用法：`aila doctor [路径]`

参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--cc` | string | `—` | 临时覆盖清单的 C 编译器选择 |
| `--llvm` | string | `—` | 临时覆盖清单的 LLVM 后端选择 |

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila explain`

解释错误码并给出修法

用法：`aila explain <CODE>`

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila ir`

把一份 AILA 源文件降成 LLVM IR 文本并打印

用法：`aila ir <文件.aila>`

参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--triple` | string | `—` | 目标三元组（默认取宿主三元组） |

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila lsp`

启动语言服务器（编辑器插件用，JSON-RPC over stdio）

用法：`aila lsp [command]`

参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--log` | string | `—` | 把服务器日志写到文件（默认 stderr） |

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila new`

新建工程骨架

用法：`aila new <目录>`

参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--name` | string | `—` | 包名（默认取目录名） |
| `--triple` | string | `—` | 目标三元组（写入 [target]；不给 = 宿主） |

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila parse`

解析一份 AILA 源文件并打印 token 流或语法树（默认 S-表达式）

用法：`aila parse <文件.aila>`

参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--tokens` | bool | `—` | 只打印 token 流 |

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila run`

构建并运行

用法：`aila run [路径] [-- 参数...]`

参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--arg` | stringArray | `[]` | 传给程序的参数（可重复） |
| `--cc` | string | `—` | C 编译器 / 链接器选择（auto / bundled / system:...） |
| `--deps-update` | bool | `—` | 重新解析跨项目依赖并重写 aila.lock |
| `--emit-bindings` | bool | `—` | 把生成的 C 绑定 `.ail` 复制到产物目录 |
| `--emit-ir` | bool | `—` | 额外写出 `.ll` 文本 |
| `--emit-json` | bool | `—` | 写出构建报告 `aila-build.json` |
| `--emit-size-delta` | bool | `—` | 构建后报告产物与每个模块目标文件的大小（量化换驱动 / 关功能的体积差） |
| `--gc-sections` | bool | `—` | 函数/数据分段 + 链接期 --gc-sections：未被引用的模块代码不进产物（配合 #compile.match 选驱动） |
| `--llvm` | string | `—` | LLVM 后端选择（auto / bundled / system:...） |
| `--opt` | int | `0` | 优化级别（0-3，临时覆盖清单） |
| `--out` | string | `—` | 输出文件名（默认 `app` + 平台后缀） |
| `--triple` | string | `—` | 目标三元组（临时覆盖清单，如 xtensa-esp8266-elf） |
| `--vendor` | bool | `—` | 把依赖复制到 vendor/（离线构建） |

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila target`

目标三元组与已装工具覆盖

用法：`aila target [command]`

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila target list`

概览已装工具链覆盖的架构与宿主三元组

用法：`aila target list [command]`

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila target show`

拆解一个三元组并探测哪些已装工具能服务它

用法：`aila target show <triple>`

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila toolchain`

随包工具链：查看 / 安装

用法：`aila toolchain [command]`

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila toolchain info`

打印工具链解析结果

用法：`aila toolchain info [command]`

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila toolchain install`

安装工具链：按 available 清单联网下载，或本地 zip 离线安装

用法：`aila toolchain install <family[@version] | gcc/<triple>[@version] | sdk/<名称>[@版本] | 本地zip>`

参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--family` | string | `—` | 本地 zip 安装时指定家族（clang / esp-clang / gcc） |
| `--force` | bool | `—` | 覆盖已有目录 |
| `--mirror` | string | `—` | 下载镜像前缀（镜像会拼在原 URL 前，可逗号分隔多个） |
| `--triple` | string | `—` | gcc 家族本地 zip 安装时指定目标三元组 |
| `--version` | string | `—` | 本地 zip 安装时指定版本（默认 zip 内 VERSION 文件或文件名） |

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila toolchain list`

列出已安装的工具链家族与可下载清单

用法：`aila toolchain list [command]`

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

## `aila version`

显示版本、AILA_HOME 与工具链来源

用法：`aila version [command]`

继承参数：

| 参数 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--color` | string | `auto` | auto\|always\|never |
| `--json` | bool | `—` | 机器可读输出 |
| `--quiet` | bool | `—` | 少输出 |
| `--verbose` | bool | `—` | 多输出 |

