# `aila.toml` 清单参考

一个工程 = 一个**包**，包根必须有 `aila.toml`。字段少而硬，写错会被点名（`E7002`/`E7006`/`E7007`/`W7001`）。

## 最小可用

```toml
[package]
name = "demo"          # 必填；[A-Za-z_][A-Za-z0-9_-]*
version = "0.1.0"      # 必填
entry = "_.main"       # 入口模块名（默认就是 "_.main"）

[target]
board = "host"         # 默认 host
```

## 全字段

```toml
[package]
name = "demo"
version = "0.1.0"
entry = "_.main"                 # 入口模块：模块名 = 路径（sub/app.aila → "_.sub.app"）

[target]
board = "esp32s3"                # 内置 board：host / esp32s3 / riscv32 / aarch64
# arch = "xtensa"                # board 不在表里时，必须给 arch 或 triple
# triple = "xtensa-esp32s3-elf"

[build]
opt = 0                          # 0..3（默认 0）
warnings = "default"             # default | strict | allow
debug = false

[build.toolchain]
cc = "bundled"                   # auto（默认）| bundled | system:clang | system:gcc
llvm = "bundled"                 # auto（默认）| bundled | system:clang

[domains]                        # 物理域（嵌入式分域内存）；名 → 种类
psram = { kind = "external-ram", size = "8M", align = 16 }
rtc = { kind = "rtc-slow", size = "16K" }

[dependencies]
mylib = { path = "../mylib" }                       # 本地依赖
# mylib = { git = "https://...", tag = "v1.0.0" }   # 远端：tag 与 rev 二选一，必须锁死
```

## 字段说明

| 段 | 键 | 说明 |
|---|---|---|
| `[package]` | `name` | 必填；包名（依赖别人时它就是导入根） |
| | `version` | 必填；`aila.lock` 会记录 |
| | `entry` | 入口模块名，默认 `"_.main"`；该模块里必须有 `func main()`（否则 E7005） |
| `[target]` | `board` | 默认 `host`；`aila target list` 看候选 |
| | `arch` / `triple` | board 不在内置表时才需要 |
| `[build]` | `opt` | `0..3`，越界 E7007 |
| | `warnings` | `default`/`strict`/`allow` |
| | `debug` | 是否带调试信息 |
| `[build.toolchain]` | `cc` | `auto`/`bundled`/`system:名字`；默认只用随包 clang |
| | `llvm` | 同上 |
| `[domains]` | `<名>` | `kind` 必须是 `internal-ram`/`external-ram`/`rtc-slow`/`default`；嵌入式 target 的 `size` 必填（形如 `64K`/`8M`） |
| `[dependencies]` | `<名>` | 名字即导入根；`path` 或 `git`+`tag`/`rev` 二选一（E7021） |

## 与命令行的关系

- 清单里的值可以被命令行**临时覆盖**（不改文件）：`--board`、`--opt`、`--cc`、`--llvm`。
- `aila deps --update` 会按依赖重新生成 `aila.lock`（锁文件记录来源/版本/内容哈希，
  不会自动跟着漂移，不一致就报 E7023）。
- 未知段/键不会静默忽略，会给你 `W7001`。

## 目录约定（与清单配套）

```
demo/
  aila.toml
  main.aila                 # entry 指向的模块
  <名字>.aila               # 同包其它模块 → import _.名字;
  <子包>/<名字>.aila        # 子包模块     → import _.子包.名字;
  cpacks/<名字>/{include,*.c}   # C 包 → import ch _.cpacks.<名字>;
  build/                    # 产物（可删：aila clean）
  aila.lock / vendor/       # 依赖锁与离线副本
```

- 子包名与模块名都必须是标识符（`my-lib` 会报 E7003）。
- `cpacks/` 只在**根包**里解析；依赖包里的 C 导入当前不支持（E7010）。
