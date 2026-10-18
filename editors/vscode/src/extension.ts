// AILA VSCode 扩展入口。
//
// 架构：**插件只做渲染，语义全部来自编译器**。
//   - 语法高亮 / 括号 / 片段：纯声明式贡献点（package.json + tmLanguage + snippets）
//   - 大纲 / 跳转定义 / 悬停 / 查找引用 / 同文件高亮 / 工程符号 / 诊断标红 /
//     补全 / 签名帮助 / 重命名 / 折叠 / 选区 / 格式化 / 内联提示：一律问 `aila lsp`
//
// 为什么自己写一个极简 LSP 客户端而不是引 vscode-languageclient：
// 我们只用只读 + 少量编辑类能力（无动态注册、无代码动作），
// 一个三百行的分帧客户端足够，既不引运行时依赖，也不绑定编辑器 API 版本。
import * as vscode from "vscode";
import * as cp from "node:child_process";
import * as fs from "node:fs";
import * as path from "node:path";

type Json = any;

let output: vscode.OutputChannel | undefined;

function channel(): vscode.OutputChannel {
  if (!output) output = vscode.window.createOutputChannel("AILA");
  return output;
}

function cfg<T>(key: string, fallback: T): T {
  return vscode.workspace.getConfiguration("aila").get<T>(key) ?? fallback;
}

// workspaceRoot 取当前文件所在工程的根（向上找 aila.toml），找不到用第一个工作区目录。
async function workspaceRoot(): Promise<string | undefined> {
  const active = vscode.window.activeTextEditor?.document.uri;
  if (active) {
    let dir = path.dirname(active.fsPath);
    for (;;) {
      try {
        await vscode.workspace.fs.stat(vscode.Uri.joinPath(vscode.Uri.file(dir), "aila.toml"));
        return dir;
      } catch {
        const parent = path.dirname(dir);
        if (parent === dir) break;
        dir = parent;
      }
    }
  }
  const folders = vscode.workspace.workspaceFolders;
  return folders && folders.length > 0 ? folders[0].uri.fsPath : undefined;
}

// --------------------------------------------------------------------------- //
// 找 `aila` 可执行文件（傻瓜式：不配也能用）
//
// 顺序：显式配置 → $AILA_HOME → 从工程根逐级向上找 bin/aila(.exe) / aila(.exe) → PATH 里的 aila。
// 「从工程根向上找」是为了源码仓库里直接打开 examples/xxx 就能用（仓库自带 bin/aila.exe）。
// --------------------------------------------------------------------------- //

let resolvedExe: string | undefined;

function exeName(): string {
  return process.platform === "win32" ? "aila.exe" : "aila";
}

function existsFile(p: string): boolean {
  try {
    return fs.existsSync(p) && fs.statSync(p).isFile();
  } catch {
    return false;
  }
}

async function resolveExecutable(): Promise<string> {
  if (resolvedExe) return resolvedExe;
  const configured = cfg<string>("executable", "").trim();
  const name = exeName();
  if (configured && configured !== "aila") {
    resolvedExe = configured;
    return resolvedExe;
  }
  const cands: string[] = [];
  const home = process.env.AILA_HOME;
  if (home) cands.push(path.join(home, "bin", name), path.join(home, name));
  const root = await workspaceRoot();
  if (root) {
    let dir = root;
    for (let i = 0; i < 6; i++) {
      cands.push(path.join(dir, "bin", name), path.join(dir, name));
      const parent = path.dirname(dir);
      if (parent === dir) break;
      dir = parent;
    }
  }
  for (const c of cands) {
    if (existsFile(c)) {
      resolvedExe = c;
      channel().appendLine(`[aila] 使用 ${c}`);
      return resolvedExe;
    }
  }
  resolvedExe = "aila"; // 兜底交给 PATH
  channel().appendLine("[aila] PATH 里找 aila（未在工作区里发现 bin/aila）");
  return resolvedExe;
}

// forgetExecutable 在用户改了设置或重启语言服务器时清缓存。
function forgetExecutable(): void {
  resolvedExe = undefined;
}

// --------------------------------------------------------------------------- //
// 极简 LSP 客户端（分帧 + 请求/通知 + 诊断推送）
// --------------------------------------------------------------------------- //

class LSPClient implements vscode.Disposable {
  private proc?: cp.ChildProcessWithoutNullStreams;
  private buf = Buffer.alloc(0);
  private id = 1;
  private pending = new Map<number, { resolve: (v: Json) => void }>();
  private ready?: Promise<void>;

  constructor(
    private exe: string,
    private cwd: string,
    private collection: vscode.DiagnosticCollection,
  ) {}

  start(): Promise<void> {
    if (this.ready) return this.ready;
    this.ready = new Promise<void>((resolve) => {
      const proc = cp.spawn(this.exe, ["lsp"], { cwd: this.cwd, shell: process.platform === "win32" });
      this.proc = proc as cp.ChildProcessWithoutNullStreams;
      proc.stdout.on("data", (d: Buffer) => this.onData(d));
      proc.stderr?.on("data", (d: Buffer) => channel().append(d.toString()));
      proc.on("error", (err: Error) => {
        channel().appendLine(`[lsp] 启动失败：${err.message}（用的是 ${this.exe}）`);
        void vscode.window
          .showWarningMessage(
            `AILA：启动语言服务器失败。找不到 \`${this.exe}\`——请把设置 aila.executable 指向 aila 可执行文件（例如 <仓库>/bin/aila.exe）。`,
            "打开设置",
            "查看日志",
          )
          .then((pick) => {
            if (pick === "打开设置") {
              void vscode.commands.executeCommand("workbench.action.openSettings", "aila.executable");
            } else if (pick === "查看日志") {
              channel().show(true);
            }
          });
        resolve();
      });
      proc.on("exit", (code: number | null) => {
        channel().appendLine(`[lsp] 退出（code=${code}）`);
        this.proc = undefined;
        this.ready = undefined;
      });
      void this.request("initialize", {
        processId: process.pid,
        rootUri: vscode.Uri.file(this.cwd).toString(),
        capabilities: {
          workspace: { symbol: {} },
          textDocument: {
            documentSymbol: { hierarchicalDocumentSymbolSupport: true },
            hover: { contentFormat: ["markdown", "plaintext"] },
            definition: {},
            references: {},
            documentHighlight: {},
            typeDefinition: {},
            implementation: {},
            documentLink: {},
            completion: {
              completionItem: { snippetSupport: true, documentationFormat: ["markdown", "plaintext"] },
              contextSupport: true,
            },
            signatureHelp: {},
            rename: { prepareSupport: true },
            foldingRange: { lineFoldingOnly: true },
            selectionRange: {},
            formatting: {},
            inlayHint: {},
          },
        },
      }).then(() => {
        this.notify("initialized", {});
        resolve();
      });
    });
    return this.ready;
  }

  private onData(chunk: Buffer) {
    this.buf = Buffer.concat([this.buf, chunk]);
    for (;;) {
      const sep = this.buf.indexOf("\r\n\r\n");
      if (sep < 0) return;
      const header = this.buf.subarray(0, sep).toString("ascii");
      const m = /Content-Length:\s*(\d+)/i.exec(header);
      if (!m) {
        this.buf = this.buf.subarray(sep + 4);
        continue;
      }
      const len = Number(m[1]);
      if (this.buf.length < sep + 4 + len) return;
      const body = this.buf.subarray(sep + 4, sep + 4 + len).toString("utf8");
      this.buf = this.buf.subarray(sep + 4 + len);
      let msg: Json;
      try {
        msg = JSON.parse(body);
      } catch {
        continue;
      }
      this.dispatch(msg);
    }
  }

  private dispatch(msg: Json) {
    if (msg.id !== undefined && this.pending.has(msg.id)) {
      const p = this.pending.get(msg.id)!;
      this.pending.delete(msg.id);
      p.resolve(msg.result);
      return;
    }
    if (msg.method === "textDocument/publishDiagnostics") {
      this.setDiagnostics(msg.params);
    }
  }

  private setDiagnostics(params: Json) {
    const uri = vscode.Uri.parse(params.uri);
    const diags: vscode.Diagnostic[] = (params.diagnostics ?? []).map((d: Json) => {
      const range = new vscode.Range(
        d.range.start.line, d.range.start.character,
        d.range.end.line, d.range.end.character,
      );
      const diag = new vscode.Diagnostic(
        range, d.message,
        d.severity === 1 ? vscode.DiagnosticSeverity.Error : vscode.DiagnosticSeverity.Warning,
      );
      diag.source = "aila";
      if (d.code) diag.code = d.code;
      return diag;
    });
    this.collection.set(uri, diags);
  }

  private send(obj: Json) {
    const proc = this.proc;
    if (!proc || !proc.stdin.writable) return;
    const data = Buffer.from(JSON.stringify(obj), "utf8");
    proc.stdin.write(`Content-Length: ${data.length}\r\n\r\n`);
    proc.stdin.write(data);
  }

  notify(method: string, params: Json) {
    this.send({ jsonrpc: "2.0", method, params });
  }

  request(method: string, params: Json, timeoutMs = 8000): Promise<Json> {
    const id = this.id++;
    return new Promise<Json>((resolve, reject) => {
      this.pending.set(id, { resolve });
      this.send({ jsonrpc: "2.0", id, method, params });
      setTimeout(() => {
        if (this.pending.has(id)) {
          this.pending.delete(id);
          reject(new Error(`${method} 超时`));
        }
      }, timeoutMs);
    });
  }

  dispose() {
    this.notify("shutdown", null);
    this.notify("exit", {});
    this.proc?.kill();
    this.proc = undefined;
    this.ready = undefined;
  }
}

// --------------------------------------------------------------------------- //
// LSP ⇄ VSCode 类型转换
// --------------------------------------------------------------------------- //

function toRange(r: Json): vscode.Range {
  return new vscode.Range(r.start.line, r.start.character, r.end.line, r.end.character);
}

const symbolKinds: Record<number, vscode.SymbolKind> = {
  2: vscode.SymbolKind.Module,
  6: vscode.SymbolKind.Method,
  8: vscode.SymbolKind.Field,
  10: vscode.SymbolKind.Enum,
  12: vscode.SymbolKind.Function,
  13: vscode.SymbolKind.Variable,
  14: vscode.SymbolKind.Constant,
  22: vscode.SymbolKind.EnumMember,
  23: vscode.SymbolKind.Struct,
  26: vscode.SymbolKind.TypeParameter,
};

function toSymbol(s: Json): vscode.DocumentSymbol {
  const sym = new vscode.DocumentSymbol(
    s.name,
    s.detail ?? "",
    symbolKinds[s.kind] ?? vscode.SymbolKind.Object,
    toRange(s.range),
    toRange(s.selectionRange ?? s.range),
  );
  for (const child of s.children ?? []) sym.children.push(toSymbol(child));
  return sym;
}

// 补全项种类（LSP 的 CompletionItemKind 与 vscode.CompletionItemKind 数值一致）。
const completionKinds: Record<number, vscode.CompletionItemKind> = {
  2: vscode.CompletionItemKind.Method,
  3: vscode.CompletionItemKind.Function,
  5: vscode.CompletionItemKind.Field,
  6: vscode.CompletionItemKind.Variable,
  9: vscode.CompletionItemKind.Module,
  13: vscode.CompletionItemKind.Enum,
  14: vscode.CompletionItemKind.Keyword,
  15: vscode.CompletionItemKind.Snippet,
  20: vscode.CompletionItemKind.EnumMember,
  21: vscode.CompletionItemKind.Constant,
  22: vscode.CompletionItemKind.Struct,
  25: vscode.CompletionItemKind.TypeParameter,
};

function toSelectionRange(r: Json): vscode.SelectionRange {
  const parent = r.parent ? toSelectionRange(r.parent) : undefined;
  return new vscode.SelectionRange(toRange(r.range), parent);
}

// --------------------------------------------------------------------------- //
// 激活
// --------------------------------------------------------------------------- //

export async function activate(ctx: vscode.ExtensionContext): Promise<void> {
  const collection = vscode.languages.createDiagnosticCollection("aila");
  ctx.subscriptions.push(collection);

  let client: LSPClient | undefined;

  const ensureClient = async (): Promise<LSPClient | undefined> => {
    if (!cfg<boolean>("lsp.enable", true)) return undefined;
    if (client) return client;
    const root = await workspaceRoot();
    if (!root) return undefined;
    client = new LSPClient(await resolveExecutable(), root, collection);
    await client.start();
    return client;
  };

  if (await workspaceRoot()) void ensureClient();

  // ---- 文档同步（全量） ----
  const syncOpen = (doc: vscode.TextDocument) => {
    if (doc.languageId !== "aila") return;
    void ensureClient().then((c) =>
      c?.notify("textDocument/didOpen", {
        textDocument: {
          uri: doc.uri.toString(), languageId: "aila", version: doc.version, text: doc.getText(),
        },
      }),
    );
  };
  vscode.workspace.textDocuments.forEach(syncOpen);

  ctx.subscriptions.push(
    vscode.workspace.onDidOpenTextDocument(syncOpen),
    vscode.workspace.onDidChangeTextDocument((e) => {
      if (e.document.languageId !== "aila") return;
      void ensureClient().then((c) =>
        c?.notify("textDocument/didChange", {
          textDocument: { uri: e.document.uri.toString(), version: e.document.version },
          contentChanges: [{ text: e.document.getText() }],
        }),
      );
    }),
    vscode.workspace.onDidCloseTextDocument((doc) => {
      void client?.notify("textDocument/didClose", { textDocument: { uri: doc.uri.toString() } });
      collection.delete(doc.uri);
    }),
  );

  // ---- 语言能力（全部转给 aila lsp） ----
  const selector: vscode.DocumentSelector = [{ language: "aila" }, { language: "aila-binding" }];

  ctx.subscriptions.push(
    vscode.languages.registerDocumentSymbolProvider(selector, {
      async provideDocumentSymbols(doc) {
        const c = await ensureClient();
        if (!c) return [];
        const res: Json = await c.request("textDocument/documentSymbol", {
          textDocument: { uri: doc.uri.toString() },
        });
        return (res ?? []).map(toSymbol);
      },
    }),

    vscode.languages.registerDefinitionProvider(selector, {
      async provideDefinition(doc, pos) {
        const c = await ensureClient();
        if (!c) return undefined;
        const res: Json = await c.request("textDocument/definition", {
          textDocument: { uri: doc.uri.toString() },
          position: { line: pos.line, character: pos.character },
        });
        return (res ?? []).map(
          (l: Json) => new vscode.Location(vscode.Uri.parse(l.uri), toRange(l.range)),
        );
      },
    }),

    vscode.languages.registerHoverProvider(selector, {
      async provideHover(doc, pos) {
        const c = await ensureClient();
        if (!c) return undefined;
        const res: Json = await c.request("textDocument/hover", {
          textDocument: { uri: doc.uri.toString() },
          position: { line: pos.line, character: pos.character },
        });
        if (!res) return undefined;
        const text = res.contents?.value ?? "";
        if (!text) return undefined;
        return new vscode.Hover(
          new vscode.MarkdownString(text),
          res.range ? toRange(res.range) : undefined,
        );
      },
    }),

    vscode.languages.registerReferenceProvider(selector, {
      async provideReferences(doc, pos, context) {
        const c = await ensureClient();
        if (!c) return undefined;
        const res: Json = await c.request("textDocument/references", {
          textDocument: { uri: doc.uri.toString() },
          position: { line: pos.line, character: pos.character },
          context: { includeDeclaration: context.includeDeclaration },
        });
        return (res ?? []).map(
          (l: Json) => new vscode.Location(vscode.Uri.parse(l.uri), toRange(l.range)),
        );
      },
    }),

    vscode.languages.registerDocumentHighlightProvider(selector, {
      async provideDocumentHighlights(doc, pos) {
        const c = await ensureClient();
        if (!c) return undefined;
        const res: Json = await c.request("textDocument/documentHighlight", {
          textDocument: { uri: doc.uri.toString() },
          position: { line: pos.line, character: pos.character },
        });
        return (res ?? []).map((h: Json) => new vscode.DocumentHighlight(toRange(h.range)));
      },
    }),

    // ---- 补全：关键字 / 类型 / 模块符号 / 局部 / `qual.` 成员 ----
    vscode.languages.registerCompletionItemProvider(
      selector,
      {
        async provideCompletionItems(doc, pos) {
          const c = await ensureClient();
          if (!c) return undefined;
          const res: Json = await c.request("textDocument/completion", {
            textDocument: { uri: doc.uri.toString() },
            position: { line: pos.line, character: pos.character },
          });
          if (!res) return undefined;
          const word = doc.getWordRangeAtPosition(pos);
          const items: vscode.CompletionItem[] = (res.items ?? []).map((it: Json) => {
            const item = new vscode.CompletionItem(
              it.label,
              completionKinds[it.kind] ?? vscode.CompletionItemKind.Text,
            );
            if (it.detail) item.detail = it.detail;
            if (it.sortText) item.sortText = it.sortText;
            if (it.insertText) {
              item.insertText =
                it.insertTextFormat === 2 ? new vscode.SnippetString(it.insertText) : it.insertText;
            }
            if (word) item.range = word;
            return item;
          });
          return new vscode.CompletionList(items, res.isIncomplete ?? false);
        },
      },
      ".",
      "@",
      "#",
      ":",
      "(",
    ),

    // ---- 签名帮助：光标落在实参列表里 ----
    vscode.languages.registerSignatureHelpProvider(
      selector,
      {
        async provideSignatureHelp(doc, pos) {
          const c = await ensureClient();
          if (!c) return undefined;
          const res: Json = await c.request("textDocument/signatureHelp", {
            textDocument: { uri: doc.uri.toString() },
            position: { line: pos.line, character: pos.character },
          });
          if (!res || !res.signatures?.length) return undefined;
          const help = new vscode.SignatureHelp();
          help.signatures = res.signatures.map((s: Json) => {
            const sig = new vscode.SignatureInformation(s.label);
            sig.parameters = (s.parameters ?? []).map((p: Json) => new vscode.ParameterInformation(p.label));
            if (typeof s.activeParameter === "number") sig.activeParameter = s.activeParameter;
            return sig;
          });
          help.activeSignature = res.activeSignature ?? 0;
          help.activeParameter = res.activeParameter ?? 0;
          return help;
        },
      },
      "(",
      ",",
    ),

    // ---- 重命名（含跨模块限定名） ----
    vscode.languages.registerRenameProvider(selector, {
      async prepareRename(doc, pos) {
        const c = await ensureClient();
        if (!c) return undefined;
        const res: Json = await c.request("textDocument/prepareRename", {
          textDocument: { uri: doc.uri.toString() },
          position: { line: pos.line, character: pos.character },
        });
        if (!res) throw new Error("这里不能重命名（关键字 / 内置类型 / 无法解析的符号）");
        return { range: toRange(res.range), placeholder: res.placeholder };
      },
      async provideRenameEdits(doc, pos, newName) {
        const c = await ensureClient();
        if (!c) return undefined;
        const res: Json = await c.request("textDocument/rename", {
          textDocument: { uri: doc.uri.toString() },
          position: { line: pos.line, character: pos.character },
          newName,
        });
        if (!res?.changes) return undefined;
        const edit = new vscode.WorkspaceEdit();
        for (const [uri, edits] of Object.entries(res.changes as Record<string, Json[]>)) {
          const u = vscode.Uri.parse(uri);
          for (const e of edits) edit.replace(u, toRange(e.range), e.newText);
        }
        return edit;
      },
    }),

    // ---- 折叠：代码块 / 注释 / import 段 ----
    vscode.languages.registerFoldingRangeProvider(selector, {
      async provideFoldingRanges(doc) {
        const c = await ensureClient();
        if (!c) return undefined;
        const res: Json = await c.request("textDocument/foldingRange", {
          textDocument: { uri: doc.uri.toString() },
        });
        const kinds: Record<string, vscode.FoldingRangeKind | undefined> = {
          comment: vscode.FoldingRangeKind.Comment,
          region: vscode.FoldingRangeKind.Region,
          imports: vscode.FoldingRangeKind.Imports,
        };
        return (res ?? []).map((f: Json) => new vscode.FoldingRange(f.startLine, f.endLine, kinds[f.kind]));
      },
    }),

    // ---- 逐级扩大选区 ----
    vscode.languages.registerSelectionRangeProvider(selector, {
      async provideSelectionRanges(doc, positions) {
        const c = await ensureClient();
        if (!c) return undefined;
        const res: Json = await c.request("textDocument/selectionRange", {
          textDocument: { uri: doc.uri.toString() },
          positions: positions.map((p) => ({ line: p.line, character: p.character })),
        });
        return (res ?? []).map(toSelectionRange);
      },
    }),

    // ---- 格式化：缩进 / 行尾空白 / 末行换行 ----
    vscode.languages.registerDocumentFormattingEditProvider(selector, {
      async provideDocumentFormattingEdits(doc, options) {
        const c = await ensureClient();
        if (!c) return undefined;
        const res: Json = await c.request("textDocument/formatting", {
          textDocument: { uri: doc.uri.toString() },
          options: { tabSize: options.tabSize, insertSpaces: options.insertSpaces },
        });
        return (res ?? []).map((e: Json) => vscode.TextEdit.replace(toRange(e.range), e.newText));
      },
    }),

    // ---- 内联提示：let 推断出的类型 ----
    vscode.languages.registerInlayHintsProvider(selector, {
      async provideInlayHints(doc, range) {
        const c = await ensureClient();
        if (!c) return undefined;
        const res: Json = await c.request("textDocument/inlayHint", {
          textDocument: { uri: doc.uri.toString() },
          range: {
            start: { line: range.start.line, character: range.start.character },
            end: { line: range.end.line, character: range.end.character },
          },
        });
        return (res ?? []).map((h: Json) => {
          const hint = new vscode.InlayHint(
            new vscode.Position(h.position.line, h.position.character),
            h.label,
            h.kind === 1 ? vscode.InlayHintKind.Type : undefined,
          );
          hint.paddingLeft = !!h.paddingLeft;
          hint.paddingRight = !!h.paddingRight;
          return hint;
        });
      },
    }),

    // ---- 跳转到类型定义（变量 → 结构体/枚举声明） ----
    vscode.languages.registerTypeDefinitionProvider(selector, {
      async provideTypeDefinition(doc, pos) {
        const c = await ensureClient();
        if (!c) return undefined;
        const res: Json = await c.request("textDocument/typeDefinition", {
          textDocument: { uri: doc.uri.toString() },
          position: { line: pos.line, character: pos.character },
        });
        return (res ?? []).map(
          (l: Json) => new vscode.Location(vscode.Uri.parse(l.uri), toRange(l.range)),
        );
      },
    }),

    // ---- 跳转到实现（壳内前向声明 ↔ 块内实现，`Ctrl+F12`） ----
    vscode.languages.registerImplementationProvider(selector, {
      async provideImplementation(doc, pos) {
        const c = await ensureClient();
        if (!c) return undefined;
        const res: Json = await c.request("textDocument/implementation", {
          textDocument: { uri: doc.uri.toString() },
          position: { line: pos.line, character: pos.character },
        });
        return (res ?? []).map(
          (l: Json) => new vscode.Location(vscode.Uri.parse(l.uri), toRange(l.range)),
        );
      },
    }),

    // ---- `import` 路径变成可点击链接 ----
    vscode.languages.registerDocumentLinkProvider(selector, {
      async provideDocumentLinks(doc) {
        const c = await ensureClient();
        if (!c) return undefined;
        const res: Json = await c.request("textDocument/documentLink", {
          textDocument: { uri: doc.uri.toString() },
        });
        return (res ?? [])
          .filter((l: Json) => l.target)
          .map((l: Json) => {
            const link = new vscode.DocumentLink(toRange(l.range), vscode.Uri.parse(l.target));
            if (l.tooltip) link.tooltip = l.tooltip;
            return link;
          });
      },
    }),

    vscode.languages.registerWorkspaceSymbolProvider({
      async provideWorkspaceSymbols(query) {
        const c = await ensureClient();
        if (!c) return undefined;
        const res: Json = await c.request("workspace/symbol", { query });
        return (res ?? []).map(
          (s: Json) =>
            new vscode.SymbolInformation(
              s.name,
              symbolKinds[s.kind] ?? vscode.SymbolKind.Object,
              "",
              new vscode.Location(vscode.Uri.parse(s.location.uri), toRange(s.location.range)),
            ),
        );
      },
    }),

    vscode.commands.registerCommand("aila.restartLSP", async () => {
      client?.dispose();
      client = undefined;
      forgetExecutable();
      await ensureClient();
      void vscode.window.showInformationMessage("AILA：语言服务器已重启");
    }),

    // 改了 `aila.executable` 就重开，别让缓存的可执行文件路径挡住
    vscode.workspace.onDidChangeConfiguration((e) => {
      if (!e.affectsConfiguration("aila.executable")) return;
      forgetExecutable();
      client?.dispose();
      client = undefined;
      void ensureClient();
    }),
  );

  // ---- 构建 / 运行命令（不依赖语言服务器） ----
  const runAila = (sub: string) =>
    new Promise<void>((resolve) => {
      void (async () => {
        const dir = await workspaceRoot();
        const ch = channel();
        ch.show(true);
        const exe = await resolveExecutable();
        ch.appendLine(`$ ${exe} ${sub}${dir ? `    (cwd: ${dir})` : ""}`);
        const proc = cp.spawn(exe, [sub], { cwd: dir, shell: process.platform === "win32" });
        proc.stdout?.on("data", (d: Buffer) => ch.append(d.toString()));
        proc.stderr?.on("data", (d: Buffer) => ch.append(d.toString()));
        proc.on("close", (code: number | null) => {
          ch.appendLine(`[退出码 ${code ?? 0}]`);
          resolve();
        });
        proc.on("error", (err: Error) => {
          ch.appendLine(`[失败] 启动不了 \`${exe}\`：${err.message}`);
          resolve();
        });
      })();
    });

  ctx.subscriptions.push(
    vscode.commands.registerCommand("aila.build", () => runAila("build")),
    vscode.commands.registerCommand("aila.run", () => runAila("run")),
    vscode.commands.registerCommand("aila.check", () => runAila("check")),
    vscode.commands.registerCommand("aila.doctor", () => runAila("doctor")),
    vscode.commands.registerCommand("aila.bindgen", () => runAila("bindgen")),
  );

  // ---- 任务类型 ----
  ctx.subscriptions.push(
    vscode.tasks.registerTaskProvider("aila", {
      provideTasks: async () => {
        const dir = await workspaceRoot();
        const exe = await resolveExecutable();
        return (["build", "run", "check"] as const).map((sub) => {
          const task = new vscode.Task(
            { type: "aila", command: sub },
            vscode.TaskScope.Workspace,
            `aila ${sub}`,
            "aila",
            new vscode.ShellExecution(exe, [sub], {
              cwd: dir ?? "${workspaceFolder}",
            }),
            ["$aila"],
          );
          if (sub === "build") task.group = vscode.TaskGroup.Build;
          return task;
        });
      },
      resolveTask: async (task) => {
        const def = task.definition as { command?: string };
        const sub = def.command ?? "build";
        return new vscode.Task(
          task.definition,
          task.scope ?? vscode.TaskScope.Workspace,
          task.name || `aila ${sub}`,
          "aila",
          new vscode.ShellExecution(await resolveExecutable(), [sub]),
          ["$aila"],
        );
      },
    }),
  );
}

export function deactivate(): void {
  output?.dispose();
  output = undefined;
}
