// AILA VSCode 扩展入口。
//
// 职责（M1）：
//   - 注册 AILA 语言（语法高亮 / 括号 / 片段都由 JSON 贡献点提供）
//   - 提供四个命令：build / run / check / doctor，输出走一个输出通道
//   - 提供 `aila` 任务类型，方便放进 tasks.json 或绑定快捷键
//
// 刻意不做：LSP / 语义高亮 / 跳转。这些要等前端把「解析结果导出」做成稳定接口，
// 直接在这里猜语法树只会制造第二事实源。
import * as vscode from "vscode";
import * as cp from "child_process";
import * as path from "path";

let output: vscode.OutputChannel | undefined;

function channel(): vscode.OutputChannel {
  if (!output) output = vscode.window.createOutputChannel("AILA");
  return output;
}

function ailaPath(): string {
  const cfg = vscode.workspace.getConfiguration("aila");
  return cfg.get<string>("executable") || "aila";
}

// workspaceRoot 取当前文件所在工程的根（向上找 aila.toml），找不到就用第一个工作区目录。
async function workspaceRoot(): Promise<string | undefined> {
  const active = vscode.window.activeTextEditor?.document.uri;
  if (active) {
    let dir = path.dirname(active.fsPath);
    for (;;) {
      const man = vscode.Uri.joinPath(vscode.Uri.file(dir), "aila.toml");
      try {
        await vscode.workspace.fs.stat(man);
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

function run(sub: string, args: string[] = [], cwd?: string): Promise<number> {
  const ch = channel();
  const exe = ailaPath();
  const cmd = [exe, sub, ...args].join(" ");
  ch.show(true);
  ch.appendLine(`$ ${cmd}${cwd ? `    (cwd: ${cwd})` : ""}`);
  return new Promise<number>((resolve) => {
    const child = cp.spawn(exe, [sub, ...args], { cwd, shell: process.platform === "win32" });
    child.stdout?.on("data", (d: Buffer) => ch.append(d.toString()));
    child.stderr?.on("data", (d: Buffer) => ch.append(d.toString()));
    child.on("error", (err: Error) => {
      ch.appendLine(`[失败] 启动不了 \`${exe}\`：${err.message}`);
      ch.appendLine("请把 aila 放进 PATH，或用设置 `aila.executable` 指定完整路径。");
      resolve(-1);
    });
    child.on("close", (code: number | null) => {
      ch.appendLine(`[退出码 ${code ?? 0}]`);
      resolve(code ?? 0);
    });
  });
}

export function activate(ctx: vscode.ExtensionContext): void {
  const withRoot = async (sub: string) => {
    const root = await workspaceRoot();
    if (!root) {
      void vscode.window.showWarningMessage("AILA：找不到工程根（需要 aila.toml）");
      return;
    }
    await run(sub, [], root);
  };

  ctx.subscriptions.push(
    vscode.commands.registerCommand("aila.build", () => withRoot("build")),
    vscode.commands.registerCommand("aila.run", () => withRoot("run")),
    vscode.commands.registerCommand("aila.check", () => withRoot("check")),
    vscode.commands.registerCommand("aila.doctor", () => withRoot("doctor")),
  );

  ctx.subscriptions.push(
    vscode.tasks.registerTaskProvider("aila", {
      // resolveTask：把 tasks.json 里 "type": "aila" 的定义补全成可执行任务
      resolveTask: (task: vscode.Task) => {
        const def = task.definition as { command?: string };
        const sub = def.command ?? "build";
        return new vscode.Task(
          task.definition,
          task.scope ?? vscode.TaskScope.Workspace,
          task.name || `aila ${sub}`,
          "aila",
          new vscode.ShellExecution(ailaPath(), [sub]),
          ["$aila"],
        );
      },
      provideTasks: async () => {
        const root = await workspaceRoot();
        const defs = ["build", "run", "check"] as const;
        return defs.map((sub) => {
          const def: vscode.TaskDefinition = { type: "aila", command: sub };
          const task = new vscode.Task(
            def,
            vscode.TaskScope.Workspace,
            `aila ${sub}`,
            "aila",
            new vscode.ShellExecution(ailaPath(), [sub], { cwd: root ?? "${workspaceFolder}" }),
            ["$aila"],
          );
          task.group = sub === "build" ? vscode.TaskGroup.Build : undefined;
          return task;
        });
      },
    }),
  );
}

export function deactivate(): void {
  output?.dispose();
  output = undefined;
}
