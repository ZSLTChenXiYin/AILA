# AILA for VSCode

AILA 语言的 VSCode 扩展。分两层：

| 层 | 内容 | 依赖 |
|---|---|---|
| **声明式**（不跑代码） | 语法高亮、括号/注释、代码片段 | 无 |
| **语言服务器**（`aila lsp`） | 大纲、跳转定义、悬停、查找引用、同文件高亮、工程符号、诊断标红、补全、签名帮助、重命名、折叠、选区、格式化、内联类型提示 | 需要 `aila` 可执行文件 |

再加构建/运行/检查/绑定生成等命令与 `aila` 任务类型。

## 装法（二选一）

### A. 直接从源码目录装（推荐，不需要打包）

```bash
cd editors/vscode
npm install          # 只为编译 TypeScript（语法高亮/片段不需要运行期依赖）
npm run compile      # 产出 out/extension.js
```

然后：

```bash
# Windows
mklink /D "%USERPROFILE%\.vscode\extensions\ailang.aila-0.1.0" "<repo>\editors\vscode"
# macOS / Linux
ln -s "$(pwd)" ~/.vscode/extensions/ailang.aila-0.1.0
```

重启 VSCode，打开任意 `.aila` 文件即可（右下角语言模式显示 **AILA**）。

### 本机已验证过的完整命令（Windows + Node 24）
```bash
npm install                 # 4 个包：typescript / @types/*
npm run check               # = tsc 编译 + 文法冒烟 + 全量高亮覆盖 + LSP 端到端
npx @vscode/vsce package --allow-missing-repository   # → aila-0.1.0.vsix（16 KB）
code --install-extension aila-0.1.0.vsix --force       # 装进本机 VSCode
code --list-extensions --show-versions | findstr aila  # → ailang.aila@0.1.0
```

### B. 打包成 `.vsix` 分发

```bash
cd editors/vscode
npm install -g @vscode/vsce
npm run compile
vsce package            # 产出 aila-0.1.0.vsix
code --install-extension aila-0.1.0.vsix
```

> 只想要语法高亮、不想装 Node？把 `package.json` 里的 `main` 删掉也能用——
> 高亮 / 括号 / 片段都是纯声明式贡献点，不需要 `out/extension.js`。

## 功能

| 功能 | 说明 |
|---|---|
| **语法高亮** | 覆盖 `internal/lexer` 的**全部** 37 个关键字、20 个内置类型名与 49 个标点：声明（`func`/`struct`/`enum`/`type`/`let`/`var`/`const`/`state`/`view`/`module hot`/`region`）、控制（`if`/`else`/`for`/`in`/`match`/`return`/`break`/`continue`/`defer`/`trap`/`yield`/`out`/`resume`/`copy`/`move`/`len`/`as`/`trunc`）、`true`/`false`、内建类型、`#[c]` 等属性、`#size`/`#align` 查询、`@region[...]`/`@mut[...]`/`@move[...]` 效果子句、`@[区域名]`、`@static`/`@heap`、字符串（含 `r"..."` 原始串与转义）、字符字面量、数字（十六进制/浮点/指数/下划线分隔）、`import` 路径与 `as` 别名、调用与成员访问、全部运算符（含 `<<=`/`>>=` 等复合赋值）。<br>**完整性由 `scripts/grammar-coverage.mjs` 断言**：仓库里所有 `.aila` + `.ail` 文件的每一个非空白 token 都必须拿到 scope，退化成纯文本即判失败 |
| **语言配置** | `//` 与 `/* */` 注释、括号自动配对、`module`/`func`/`struct`/`region`/`match` 等块自动缩进 |
| **绑定文件高亮** | `.ail`（`aila bindgen` 生成物）单独一个语言 `aila-binding` / `source.aila.binding`，复用主文法，避免和手写源码混在一起 |
| **代码片段** | `pubfunc`（壳内 public + 块内实现）、`func`、`method`、`struct`、`enum`、`result`（`T ! E`）、`opt`（`T?`）、`region`、`match`、`forc`、`forin`、`module`、`hotmodule`、`cop`（`#[c]` 绑定） |
| **语言智能**（LSP） | `F12` 跳转定义（含 `import` 别名、方法、枚举变体、字段）、`Ctrl+Shift+F12` 类型定义（变量 → struct/enum 声明）、`Ctrl+F12` 跳实现（壳内前向声明 ↔ 块内实现）、`Shift+F12` 查找引用、悬停看签名、`Ctrl+Shift+O` 大纲、`Ctrl+T` 工程符号、同名高亮、诊断标红（输入即推）、`import` 路径可点击跳转 |
| **补全** | 局部变量 / 模块级符号 / 导入别名 / 关键字 / 内置类型；`qual.` 之后按语义给成员（`alias.` → 目标模块公开符号，`变量.`/`类型.` → 字段 + 方法 + 变体）；`#[` 给属性名、`@` 给效果名、`import ` 之后给模块路径；函数项插入 `name(${1:参数})` 片段 |
| **签名帮助** | 光标在实参列表里时显示签名，并按已敲的逗号定位当前形参（`(`、`,` 触发） |
| **重命名**（`F2`） | 声明 + 同文件引用 + 跨模块的 `别名.名字` 限定引用一起改；局部符号只在本文件内改；关键字与内置类型拒绝改名 |
| **折叠 / 选区** | 代码块（`{}`）、连续 `//` 注释、多行 `/* */`、连续 `import` 段可折叠；`Shift+Alt+→/←` 按「标识符 → 括号 → 行 → 文件」逐级扩大/收缩 |
| **格式化**（`Shift+Alt+F`） | 按花括号层级重排缩进（4 空格）、去行尾空白、末行恰好一个换行；多行字符串与块注释内部原样保留（改缩进会改语义）；幂等 |
| **内联提示** | `let`/`var` 没写类型标注时，在名字后显示编译器推断出的类型（如 `: I32`） |
| **C 包**（`import ch _.cpacks.<名字>`） | 语言服务器就地跑 bindgen（缓存到 `build/.cache/bindgen-lsp/`），于是 `mathkit.` 能补全 C 函数与常量、`F12` 跳进生成的 `.ail`、悬停说明「来自 C 包 `mathkit` 的绑定」、`import` 行可点击、签名帮助拿得到 C 形参；cpack 目录缺失直接在 `import` 那行标红（`E3001`）。生成的 `.ail` 是只读产物，改名/格式化被拒绝 |
| **命令** | 命令面板里搜 `AILA:` —— 构建 / 构建并运行 / 只做检查 / 环境自检 / 生成 C 绑定 / 重启语言服务器 |
| **任务** | `Terminal → Run Task` 里出现 `aila build` / `aila run` / `aila check`，可绑快捷键或放进 `tasks.json` |

命令与语言服务器都调用**外部 `aila` 可执行文件**（不内嵌编译器）。**一般不用配**：
`aila.executable` 保持默认时按这个顺序自动找

1. 环境变量 `AILA_HOME` 下的 `bin/aila(.exe)` / `aila(.exe)`
2. 从**工程根**（含 `aila.toml` 的目录）逐级向上 6 层里的 `bin/aila(.exe)` / `aila(.exe)`
   —— 所以在源码仓库里直接打开 `examples/xxx` 就能用仓库自带的 `bin/aila.exe`
3. `PATH` 里的 `aila`

要指定别的编译器就写绝对路径：

```jsonc
// settings.json（可选）
{
  "aila.executable": "C:/path/to/aila/bin/aila.exe"
}
```

## 建议的 `tasks.json`

```json
{
  "version": "2.0.0",
  "tasks": [
    { "type": "aila", "command": "build", "problemMatcher": [], "group": { "kind": "build", "isDefault": true } },
    { "type": "aila", "command": "run", "problemMatcher": [] }
  ]
}
```

## 刻意不做

- **插件里另写一套分析**：语义一律问 `aila lsp`。插件只做「协议 → 编辑器 API」的翻译，
  这样编译器改语义时插件不用跟着改（和 `designs/0001 §1.1` 的纪律一致）。
- **语义着色（Semantic Tokens）**：TextMate 文法已经覆盖全部 token 且有覆盖检查兜底，
  再叠一层语义着色只会和主题打架；等真需要「同名变量同色」时再说。
- **代码动作 / 快速修复**：诊断现在只报错不自动改。要加也是先让 `aila check --json` 输出稳定结构。
- **索引依赖包**：语言服务器索引「当前工程 + 标准库」；`deps` 里的外部包只在跳转时按需读取。

## 目录

```text
editors/vscode/
  package.json                 扩展清单（语言/语法/片段/命令/任务/设置）
  language-configuration.json  注释、括号、缩进
  syntaxes/aila.tmLanguage.json        TextMate 语法（.aila，全量覆盖）
  syntaxes/aila-binding.tmLanguage.json 生成物 .ail 的语法（复用主文法）
  snippets/aila.json           代码片段
  src/extension.ts             语言服务器客户端 + 命令与任务入口（TypeScript）
  scripts/grammar-smoke.mjs    文法冒烟（21 条断言）
  scripts/grammar-coverage.mjs 高亮完整性检查（所有 token 必须有 scope）
  scripts/lsp-smoke.mjs        语言服务器端到端（53 条协议断言，含 C 导入）
  tsconfig.json
```
