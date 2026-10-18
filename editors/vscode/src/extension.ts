// AILA VSCode 扩展入口。
//
// 架构：**插件只做渲染，语义全部来自编译器**。
//   - 语法高亮 / 括号 / 片段：纯声明式贡献点（package.json + tmLanguage + snippets）
//   - 大纲 / 跳转定义 / 悬停 / 查找引用 / 同文件高亮 / 工程符号搜索 / 诊断标红：一律问 `aila lsp`
//
// 为什么自己写一个极简 LSP 客户端而不是引 vscode-languageclient：
// 我们只用到 6 个只读能力（无动态注册、无代码动作、无重命名），
// 一个两百行的分帧客户端足够，既不引运行时依赖，也不绑定编辑器 API 版本。
import * as vscode from "vscode";
import * as cp from "node:child_process";
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
        channel().appendLine(`[lsp] 启动失败：${err.message}`);
        void vscode.window.showWarningMessage("AILA：启动语言服务器失败（检查 aila.executable 设置）");
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
          textDocument: {
            documentSymbol: { hierarchicalDocumentSymbolSupport: true },
            hover: { contentFormat: ["markdown", "plaintext"] },
            definition: {},
            references: {},
            documentHighlight: {},
          },
          workspace: { symbol: {} },
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
    client = new LSPClient(cfg<string>("executable", "aila"), root, collection);
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
      await ensureClient();
      void vscode.window.showInformationMessage("AILA：语言服务器已重启");
    }),
  );

  // ---- 构建 / 运行命令（不依赖语言服务器） ----
  const runAila = (sub: string) =>
    new Promise<void>((resolve) => {
      void (async () => {
        const dir = await workspaceRoot();
        const ch = channel();
        ch.show(true);
        const exe = cfg<string>("executable", "aila");
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
        return (["build", "run", "check"] as const).map((sub) => {
          const task = new vscode.Task(
            { type: "aila", command: sub },
            vscode.TaskScope.Workspace,
            `aila ${sub}`,
            "aila",
            new vscode.ShellExecution(cfg<string>("executable", "aila"), [sub], {
              cwd: dir ?? "${workspaceFolder}",
            }),
            ["$aila"],
          );
          if (sub === "build") task.group = vscode.TaskGroup.Build;
          return task;
        });
      },
      resolveTask: (task) => {
        const def = task.definition as { command?: string };
        const sub = def.command ?? "build";
        return new vscode.Task(
          task.definition,
          task.scope ?? vscode.TaskScope.Workspace,
          task.name || `aila ${sub}`,
          "aila",
          new vscode.ShellExecution(cfg<string>("executable", "aila"), [sub]),
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
