# RFC 0004 · 跨项目依赖（包管理）

> **编号**：0004 · **类型**：RFC · **状态**：已采纳并落地（M1 范围：路径依赖 + git 依赖 + 锁文件 + vendor + 跨包可见性）
> **上游**：[designs/0001](../designs/0001-language-spec.md) §2.1–§2.3（模块 / 可见性 / 导入）、
> [specs/0001](../specs/0001-manager-cli.md) §1.2 / §2（工程模型与清单）
> **实现**：`internal/deps`（解析 / 锁 / vendor）、`internal/build`（多包装载与导入解析）、`internal/sema`（跨包可见性 `E7020`）、`internal/cli`（`aila deps`）

## 1. 术语与不变量

| 词 | 含义 | 不变量 |
|---|---|---|
| **包（package）** | 一个工程 = 一个包 = 一个 `aila.toml` 所在的目录 | 包名 = 清单 `package.name` |
| **子包（subpackage）** | 包内的目录 | 子包名 = 目录名，可嵌套，路径用 `.` 连接 |
| **模块（module）** | 包内的 `.aila` 文件 | 模块名 = 文件名；模块是热替换的最小单位 |
| **依赖（dependency）** | **另一个包**，通过 `[dependencies]` 声明 | 依赖名 = 导入根 = 依赖包的包名（不强制与对方 `package.name` 相同，但强烈建议相同） |

一句话：**包管理管的是「包与包」；`import` 指向的永远是「模块」**——`import mylib.math.ops;`
依次是〔依赖名 = 包〕·〔子包 math〕·〔模块 ops〕。

## 2. 清单写法

```toml
[dependencies]
mylib = { path = "../mylib" }                       # 本地路径（相对本包根）
util  = { git = "https://…/util.git", tag = "v1.2.0" }  # git，tag 锁定
mat   = { git = "https://…/mat.git", rev = "3f9a1c2" }  # git，提交锁定
math  = { path = "../math", version = "0.3" }        # 可选版本要求（前缀匹配）
```

规则：

| 规则 | 违反 |
|---|---|
| 恰好一种来源：`path` 或 `git` | `E7021` |
| git 必须锁 `tag` 或 `rev`（**不允许分支**，否则构建不可复现），二者不能同时写 | `E7021` |
| 依赖名必须是合法标识符（它就是导入根） | `E7021` |
| `path` 指向依赖的**包根**（含 `aila.toml` 的那一层） | `E7022` |
| `version` 写前缀（`"0.3"` 匹配 `0.3.7`，不匹配 `0.30.0`） | `E7023` |
| 同一个依赖名只能解析到一个目录 | `E7024` |
| 依赖图必须有向无环 | `E7025` |

传递依赖照常：依赖包自己的 `[dependencies]` 继续生效，形成依赖图（拓扑序编译，依赖在前）。

## 3. 解析顺序

对每条依赖声明：

1. **`vendor/<名字>/` 存在** → 优先用它（离线构建；嵌入式与 CI 友好）；
2. 否则按声明取目录：`path` 直接拼；`git` 取缓存 `<AILA_DEPS_CACHE>` → `$AILA_HOME/cache/deps` → 用户缓存目录，未命中才 `git clone --depth 1 --branch <tag>`；
3. 读它的 `aila.toml` 校验 `package.version`（存在且满足要求的版本前缀）。

目录**永远由清单决定**，锁文件不参与定位——这样 `aila.lock` 可以随仓库走、不绑定本机绝对路径。

## 4. 锁文件 `aila.lock`

```toml
[dep.mylib]
source = "path:../mylib"
version = "0.1.0"
hash = "ec023f1b088bab2e"
```

- `source` 记**清单怎么写就怎么记**的可移植形态；`hash` = 包内 `aila.toml` + 全部 `.aila` 按路径排序后的 sha256 前 16 位；
- 校验时机：每次 `check/build/run`。来源、版本、内容哈希**任一**不同 → `E7023`，并提示跑 `aila deps --update`；
- 第一次解析（无锁）自动写入；
- 设计取向：**默认拒绝漂移**。构建可复现优先于「悄悄跟上最新依赖」。

## 5. 跨包可见性

`public` 是全语言唯一一档可见性提升（designs §2.2）。跨包引用**只认带 `public` 的声明**：

| 引用对象 | 不 public |
|---|---|
| 函数 / 方法 | `E7020` |
| 类型（struct / enum，含 `绑定.类型 { … }` 构造） | `E7020` |
| 常量 / 模块级绑定 | `E7020` |
| 枚举变体 | 随其 enum：enum 不 public 即 `E7020` |

同包内跨模块照旧（不需要 `public`），包边界是唯一需要 `public` 的地方。

> **设计稿缺陷 14（本文一并裁决）**：按 designs §2.2 的字面表格，`module {}` 块内的 `public` 只到「包内」，
> 而「壳」禁止有体函数——于是**任何 AILA 实现的函数都无法跨包导出**，与 §2.3「非 `_` 根 = 外部包」直接矛盾。
> 裁决：**`public` 在块内也构成跨包导出**；块内 `public` 的「分派表面」含义不变（两者是不同维度的语义）。
> 设计稿修订建议见 [rfcs/0002 缺陷 14](0002-design-defects.md)。

## 6. 命令行

```text
aila deps [路径] [--update] [--vendor] [--json]   # 打印依赖图 / 重写 aila.lock / 生成 vendor/
aila build|run|check [路径] [--deps-update] [--vendor]
```

- `--update`：忽略锁文件里的校验，重新解析并重写；
- `--vendor`：把解析好的依赖复制到 `vendor/<名字>/`（只复制 `aila.toml` 与 `.aila`，跳过 `build/`、`.git`）。

## 7. 与其它机制的边界

| 机制 | 关系 |
|---|---|
| **热替换** | 依赖包的 hot 模块与新版本边界同样走「跨模块调用不可内联、走分派表」的既有铁律；依赖是**构建期**概念，运行期不感知包 |
| **C 互操作** | `import ch|cc` 仍只在**根包**的 `cpacks/` 里解析；依赖包里的 C 导入暂不支持（报 `E7010`），把 C 包留在根项目 |
| **标准库** | `std` 是内建包（不是依赖），不需要声明，不需要锁；它的对外函数带 `public` |
| **动态加载** | 「运行期才确定」的依赖走 `std.dynlib`，不进包管理（designs §2.3 依赖通道判定） |

## 8. 刻意不做（M1 非目标）

| 不做 | 理由 |
|---|---|
| 中心化注册表 / 版本求解 | 嵌入式与「傻瓜式」优先：路径 + git + vendor 已覆盖真实场景；版本冲突留给显式声明与同名冲突（`E7024`） |
| 依赖名与包名的一致性强制校验 | 允许改名（如两个同名库并存），代价是使用者要写清楚；`aila deps` 会把人读图打出来 |
| 增量缓存与网络优化 | 已有 `build/.cache`（bindgen）与 git 缓存，模块级增量缓存见 specs/0001 §5.3 批次 |

## 9. 验收

| 判据 | 位置 |
|---|---|
| 路径依赖解析 / 锁文件可移植 / 漂移检测 | `internal/deps` `TestPathDependency` |
| vendor 离线构建 | `TestVendorOffline` |
| 版本前缀匹配 | `TestVersionMismatch` / `TestVersionMatches` |
| 成环 / 声明不合法 | `TestCycleRejected` / `TestBadDependencyDeclaration` |
| 跨项目真实编译运行 + 非 public 拒绝 | `internal/build` `TestCrossProjectDependency` |
