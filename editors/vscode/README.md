# AILA for VSCode

AILA 语言的 VSCode 扩展：**语法高亮 + 语言配置 + 代码片段 + 构建/运行命令**。

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
npm run check               # = tsc 编译 + 语法高亮冒烟（19 条断言，119 个 token）
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
| **语法高亮** | 关键字（`func`/`module`/`region`/`match`/`defer`…）、内建类型（`I32`/`U8`/`Text`/`USize`…）、`#[c]`/`#[intrinsic]`/`#size` 属性、`@region`/`@mut`/`@move` 效果子句、`@[区域名]`、字符串（含 `r"..."` 原始串）、数字（十六进制/浮点/下划线分隔）、`import` 路径与别名 |
| **语言配置** | `//` 与 `/* */` 注释、括号自动配对、`module`/`func`/`struct`/`region`/`match` 等块自动缩进 |
| **代码片段** | `pubfunc`（壳内 public + 块内实现）、`func`、`method`、`struct`、`enum`、`result`（`T ! E`）、`opt`（`T?`）、`region`、`match`、`forc`、`forin`、`module`、`hotmodule`、`cop`（`#[c]` 绑定） |
| **命令** | 命令面板里搜 `AILA:` —— 构建 / 构建并运行 / 只做检查 / 环境自检 |
| **任务** | `Terminal → Run Task` 里出现 `aila build` / `aila run` / `aila check`，可绑快捷键或放进 `tasks.json` |

命令都调用**外部 `aila` 可执行文件**（不内嵌编译器）：

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

- **LSP / 语义跳转 / 重命名**：这些必须由 `aila` 导出一份稳定的「解析结果接口」再做，
  否则插件里另写一遍语法树就是第二事实源（和 `designs/0001 §1.1` 的纪律冲突）。
  目前排查语法问题用 `aila parse <文件> [--tokens]`。
- **诊断标红**：`aila check` 的输出现阶段走输出通道；等诊断有了稳定 JSON 结构（`--json`）再接入
  `DiagnosticCollection`，避免解析人读文本。

## 目录

```text
editors/vscode/
  package.json               扩展清单（语言/语法/片段/命令/任务/设置）
  language-configuration.json 注释、括号、缩进
  syntaxes/aila.tmLanguage.json TextMate 语法
  snippets/aila.json          代码片段
  src/extension.ts            命令与任务入口（TypeScript）
  tsconfig.json
```
