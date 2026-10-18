// 语言服务器的端到端冒烟：真起 `aila lsp` 子进程，按 LSP 协议握手并发请求。
//
// 为什么要有它：编辑器里调不通时（例如 Content-Length 分帧、range 换算、模块映射），
// 只有真跑一遍协议才能发现；纯单元测试容易漏掉传输层。
//
// 用法：node scripts/lsp-smoke.mjs [工程目录]
// 默认对仓库 examples/result 工程跑：initialize → didOpen → documentSymbol → definition → hover → shutdown。
import { spawn } from "node:child_process";
import { readFileSync, existsSync, mkdtempSync, writeFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const repo = resolve(here, "..", "..", "..");

// 1) 找 aila 可执行文件：优先 bin/aila(.exe)，其次 PATH 里的 aila
function findAila() {
  const cands = process.platform === "win32"
    ? [join(repo, "bin", "aila.exe"), join(repo, "aila.exe")]
    : [join(repo, "bin", "aila"), join(repo, "aila")];
  for (const c of cands) if (existsSync(c)) return c;
  return "aila";
}

const projectDir = process.argv[2] || join(repo, "examples", "result");
const srcFile = join(projectDir, "main.aila");
const srcText = readFileSync(srcFile, "utf8");
const srcURI = pathToFileURL(srcFile).href;

const child = spawn(findAila(), ["lsp"], {
  cwd: projectDir,
  stdio: ["pipe", "pipe", "pipe"],
});

let stderr = "";
child.stderr.on("data", (d) => (stderr += d.toString()));

// 2) 极简 LSP 客户端：按 Content-Length 分帧收发
const pending = new Map();
let nextId = 1;
let buffer = Buffer.alloc(0);

child.stdout.on("data", (chunk) => {
  buffer = Buffer.concat([buffer, chunk]);
  for (;;) {
    const sep = buffer.indexOf("\r\n\r\n");
    if (sep < 0) return;
    const header = buffer.slice(0, sep).toString("ascii");
    const m = /Content-Length:\s*(\d+)/i.exec(header);
    if (!m) {
      buffer = buffer.slice(sep + 4);
      continue;
    }
    const len = Number(m[1]);
    if (buffer.length < sep + 4 + len) return;
    const body = buffer.slice(sep + 4, sep + 4 + len).toString("utf8");
    buffer = buffer.slice(sep + 4 + len);
    const msg = JSON.parse(body);
    if (msg.id !== undefined && pending.has(msg.id)) {
      const { resolve: res } = pending.get(msg.id);
      pending.delete(msg.id);
      res(msg);
    } else if (msg.method === "textDocument/publishDiagnostics") {
      diagnostics.push(msg.params);
    }
  }
});

const diagnostics = [];

function send(obj) {
  const data = Buffer.from(JSON.stringify(obj), "utf8");
  child.stdin.write(`Content-Length: ${data.length}\r\n\r\n`);
  child.stdin.write(data);
}

function request(method, params) {
  const id = nextId++;
  return new Promise((res, rej) => {
    pending.set(id, { resolve: res, reject: rej });
    send({ jsonrpc: "2.0", id, method, params });
    setTimeout(() => {
      if (pending.has(id)) {
        pending.delete(id);
        rej(new Error(`${method} 超时（stderr: ${stderr.slice(-400)}）`));
      }
    }, 10000);
  });
}

function notify(method, params) {
  send({ jsonrpc: "2.0", method, params });
}

// 3) 断言工具
let failures = 0;
function check(cond, label, extra = "") {
  if (cond) {
    console.log(`✔ ${label}`);
  } else {
    failures++;
    console.error(`✘ ${label}${extra ? "\n   " + extra : ""}`);
  }
}

function lineOf(text, needle) {
  const lines = text.split("\n");
  for (let i = 0; i < lines.length; i++) {
    const col = lines[i].indexOf(needle);
    if (col >= 0) return { line: i, character: col };
  }
  return null;
}

// applyEdits 把一组 LSP 文本编辑应用到文本上（从后往前，避免偏移错位）。
function applyEdits(text, edits) {
  const lines = text.split("\n");
  const offsetAt = (p) => {
    let off = 0;
    for (let i = 0; i < p.line; i++) off += lines[i].length + 1;
    return off + p.character;
  };
  const sorted = [...edits].sort((a, b) => offsetAt(b.range.start) - offsetAt(a.range.start));
  let out = text;
  for (const e of sorted) {
    out = out.slice(0, offsetAt(e.range.start)) + e.newText + out.slice(offsetAt(e.range.end));
  }
  return out;
}

try {
  const init = await request("initialize", {
    processId: process.pid,
    rootUri: pathToFileURL(projectDir).href,
    capabilities: {},
  });
  const caps = init.result?.capabilities ?? {};
  check(caps.documentSymbolProvider === true, "initialize：声明大纲能力");
  check(caps.definitionProvider === true, "initialize：声明跳转定义能力");
  check(caps.hoverProvider === true, "initialize：声明悬停能力");
  check(caps.referencesProvider === true, "initialize：声明查找引用能力");
  check(caps.documentHighlightProvider === true, "initialize：声明同文件高亮能力");
  check(caps.workspaceSymbolProvider === true, "initialize：声明工程符号能力");
  check(caps.completionProvider !== undefined, "initialize：声明补全能力");
  check(caps.signatureHelpProvider !== undefined, "initialize：声明签名帮助能力");
  check(caps.renameProvider === true, "initialize：声明重命名能力");
  check(caps.foldingRangeProvider === true, "initialize：声明折叠能力");
  check(caps.selectionRangeProvider === true, "initialize：声明选区能力");
  check(caps.documentFormattingProvider === true, "initialize：声明格式化能力");
  check(caps.inlayHintProvider === true, "initialize：声明内联提示能力");
  notify("initialized", {});

  notify("textDocument/didOpen", {
    textDocument: { uri: srcURI, languageId: "aila", version: 1, text: srcText },
  });
  await new Promise((r) => setTimeout(r, 800)); // 等诊断推送

  const diags = diagnostics.find((d) => d.uri === srcURI)?.diagnostics ?? [];
  check(Array.isArray(diags), "didOpen：收到诊断推送", JSON.stringify(diags).slice(0, 200));
  check(diags.every((d) => typeof d.message === "string" && d.range), "诊断结构合法");

  // 大纲：应包含 parse_header / read_or_default / main
  const symbols = await request("textDocument/documentSymbol", { textDocument: { uri: srcURI } });
  const names = JSON.stringify(symbols.result ?? []);
  check(names.includes("parse_header"), "documentSymbol：找到 parse_header");
  check(names.includes("main"), "documentSymbol：找到 main");

  // 跳转：`main` 里对 `parse_header` 的调用 → 定义
  const at = lineOf(srcText, "parse_header(short)");
  if (!at) throw new Error("样例里找不到 parse_header 调用点");
  const def = await request("textDocument/definition", {
    textDocument: { uri: srcURI },
    position: { line: at.line, character: at.character + 2 },
  });
  const defs = def.result ?? [];
  check(defs.length > 0, "definition：能找到 parse_header 的声明");
  if (defs.length > 0) {
    check(defs[0].uri.endsWith("main.aila"), "definition：目标在同一模块", defs[0].uri);
    check(defs[0].range.start.line < at.line, "definition：声明在调用点之前");
  }

  // 悬停
  const hov = await request("textDocument/hover", {
    textDocument: { uri: srcURI },
    position: { line: at.line, character: at.character + 2 },
  });
  const hoverText = JSON.stringify(hov.result ?? {});
  check(hoverText.includes("parse_header"), "hover：显示名字", hoverText.slice(0, 160));

  // 引用
  const refs = await request("textDocument/references", {
    textDocument: { uri: srcURI },
    position: { line: at.line, character: at.character + 2 },
    context: { includeDeclaration: true },
  });
  check((refs.result ?? []).length >= 2, "references：至少声明 + 调用两处");

  // 工程符号
  const ws = await request("workspace/symbol", { query: "parse" });
  check((ws.result ?? []).length >= 1, "workspace/symbol：按前缀搜到符号");

  // 跳转到类型定义：`ParseErr.Truncated` 里的 `ParseErr` → enum 声明
  const enumAt = lineOf(srcText, "enum ParseErr {");
  const enumUse = lineOf(srcText, "ParseErr.Truncated { got, need }");
  if (!enumAt || !enumUse) throw new Error("样例里找不到 ParseErr 的用法");
  const td = await request("textDocument/typeDefinition", {
    textDocument: { uri: srcURI },
    position: { line: enumUse.line, character: enumUse.character + 2 },
  });
  check((td.result ?? []).length === 1, "typeDefinition：`ParseErr` → 枚举声明");
  if ((td.result ?? []).length === 1) {
    check(td.result[0].range.start.line === enumAt.line, "typeDefinition：落在 enum 声明那一行",
      JSON.stringify(td.result[0].range));
  }

  // 跳到实现：`read_or_default` 只有块内实现，也应能定位
  const implAt = lineOf(srcText, "func read_or_default");
  if (!implAt) throw new Error("样例里找不到 read_or_default 声明");
  const impl = await request("textDocument/implementation", {
    textDocument: { uri: srcURI },
    position: { line: implAt.line, character: implAt.character + 10 },
  });
  check((impl.result ?? []).length >= 1, "implementation：函数能定位到实现体",
    JSON.stringify(impl.result ?? {}).slice(0, 200));

  // import 路径可点击
  const links = await request("textDocument/documentLink", { textDocument: { uri: srcURI } });
  const linkList = links.result ?? [];
  check(linkList.length >= 1, "documentLink：`import std.io` 变成链接", JSON.stringify(linkList).slice(0, 200));
  check(linkList.every((l) => typeof l.target === "string" && l.target.startsWith("file://")),
    "documentLink：目标是文件 URI");

  // 补全：`io.` 之后应列出 std.io 的公开函数
  const ioAt = lineOf(srcText, "io.print_text(");
  if (!ioAt) throw new Error("样例里找不到 io.print_text 调用点");
  const comp = await request("textDocument/completion", {
    textDocument: { uri: srcURI },
    position: { line: ioAt.line, character: ioAt.character + 3 },
  });
  const labels = (comp.result?.items ?? []).map((i) => i.label);
  check(labels.includes("print_text"), "completion：`io.` 之后列出 print_text", labels.slice(0, 12).join(","));
  check((comp.result?.items ?? []).length > 5, "completion：候选不止一条");

  // 补全：普通位置的局部变量与关键字
  const localAt = lineOf(srcText, "let v = parse_header(buf)?;");
  if (localAt) {
    const comp2 = await request("textDocument/completion", {
      textDocument: { uri: srcURI },
      position: { line: localAt.line, character: localAt.character + 4 },
    });
    const l2 = (comp2.result?.items ?? []).map((i) => i.label);
    check(l2.includes("func") && l2.includes("I32"), "completion：裸位置给出关键字与内置类型", l2.slice(0, 12).join(","));
  }

  // 签名帮助：第二个实参处 activeParameter = 1
  const callAt = lineOf(srcText, "maybe_scale(14,");
  if (!callAt) throw new Error("样例里找不到 maybe_scale 调用点");
  const sig = await request("textDocument/signatureHelp", {
    textDocument: { uri: srcURI },
    position: { line: callAt.line, character: callAt.character + "maybe_scale(14, ".length },
  });
  check((sig.result?.signatures ?? []).length === 1, "signatureHelp：给出唯一候选签名");
  check(sig.result?.activeParameter === 1, "signatureHelp：逗号之后定位到第 2 个形参",
    JSON.stringify(sig.result ?? {}).slice(0, 200));
  check(JSON.stringify(sig.result ?? {}).includes("maybe_scale"), "signatureHelp：签名里带函数名");

  // 重命名：函数名跨引用
  const ready = await request("textDocument/prepareRename", {
    textDocument: { uri: srcURI },
    position: { line: at.line, character: at.character + 2 },
  });
  check(ready.result?.placeholder === "parse_header", "prepareRename：可改名并给出占位文本",
    JSON.stringify(ready.result ?? {}));
  const ren = await request("textDocument/rename", {
    textDocument: { uri: srcURI },
    position: { line: at.line, character: at.character + 2 },
    newName: "parse_hdr",
  });
  const edits = Object.values(ren.result?.changes ?? {}).flat();
  check(edits.length >= 2, "rename：声明与调用一起改", JSON.stringify(edits).slice(0, 200));
  check(edits.every((e) => e.newText === "parse_hdr"), "rename：新名字正确");

  // 折叠：代码块与注释
  const folds = await request("textDocument/foldingRange", { textDocument: { uri: srcURI } });
  const foldList = folds.result ?? [];
  check(foldList.length >= 2, "foldingRange：给出多个可折叠区间");
  check(foldList.some((f) => f.kind === "region"), "foldingRange：包含代码块");
  check(foldList.every((f) => f.endLine > f.startLine), "foldingRange：区间必须跨行");

  // 选区：逐级扩大
  const sel = await request("textDocument/selectionRange", {
    textDocument: { uri: srcURI },
    positions: [{ line: at.line, character: at.character + 2 }],
  });
  const first = (sel.result ?? [])[0];
  check(first !== undefined && first.parent !== undefined, "selectionRange：返回带父级的选区链");

  // 格式化：幂等（已格式化的样例不应改动）
  const fmt = await request("textDocument/formatting", {
    textDocument: { uri: srcURI },
    options: { tabSize: 4, insertSpaces: true },
  });
  check(Array.isArray(fmt.result), "formatting：返回编辑数组");
  if ((fmt.result ?? []).length > 0) {
    const applied = applyEdits(srcText, fmt.result);
    const again = await request("textDocument/formatting", {
      textDocument: { uri: srcURI },
      options: { tabSize: 4, insertSpaces: true },
    });
    check(applyEdits(applied, again.result ?? []) === applied, "formatting：格式化结果幂等");
  }

  // 内联提示：`let` 推断出的类型
  const hints = await request("textDocument/inlayHint", {
    textDocument: { uri: srcURI },
    range: { start: { line: 0, character: 0 }, end: { line: 100000, character: 0 } },
  });
  const hintList = hints.result ?? [];
  check(hintList.length >= 1, "inlayHint：`let x = ...` 给出推断类型", JSON.stringify(hintList).slice(0, 200));
  check(hintList.every((h) => typeof h.label === "string" && h.label.startsWith(":")),
    "inlayHint：标签形如 `: T`");

  // 格式化：临时工程里放一份缩进凌乱的文件，验证真的会改且幂等
  const fmtDir = mkdtempSync(join(tmpdir(), "aila-fmt-"));
  try {
    writeFileSync(
      join(fmtDir, "aila.toml"),
      '[package]\nname = "demo"\nversion = "0.1.0"\nentry = "_.main"\n\n[target]\nboard = "host"\n',
    );
    const messy = "module {\n   func main() {\n         let x: I32 = 1;   \n  }\n}\n";
    writeFileSync(join(fmtDir, "main.aila"), messy);
    const messyURI = pathToFileURL(join(fmtDir, "main.aila")).href;
    notify("textDocument/didOpen", {
      textDocument: { uri: messyURI, languageId: "aila", version: 1, text: messy },
    });
    const f1 = await request("textDocument/formatting", {
      textDocument: { uri: messyURI },
      options: { tabSize: 4, insertSpaces: true },
    });
    check((f1.result ?? []).length === 1, "formatting：凌乱文件给出一处整文档替换");
    const formatted = applyEdits(messy, f1.result ?? []);
    check(
      formatted === "module {\n    func main() {\n        let x: I32 = 1;\n    }\n}\n",
      "formatting：缩进改成 4 空格、去掉行尾空白",
      JSON.stringify(formatted),
    );
    notify("textDocument/didChange", {
      textDocument: { uri: messyURI, version: 2 },
      contentChanges: [{ text: formatted }],
    });
    const f2 = await request("textDocument/formatting", {
      textDocument: { uri: messyURI },
      options: { tabSize: 4, insertSpaces: true },
    });
    check((f2.result ?? []).length === 0, "formatting：已格式化的文件不再改动（幂等）");
  } finally {
    try {
      rmSync(fmtDir, { recursive: true, force: true });
    } catch {
      /* 清理失败不影响结论 */
    }
  }

  // ---- C 导入（`import ch _.cpacks.mathkit`）：编辑器里也要能补全 / 跳转 ----
  const cMainFile = join(repo, "examples", "cinterop", "main.aila");
  if (existsSync(cMainFile)) {
    const cText = readFileSync(cMainFile, "utf8");
    const cURI = pathToFileURL(cMainFile).href;
    notify("textDocument/didOpen", {
      textDocument: { uri: cURI, languageId: "aila", version: 1, text: cText },
    });
    await new Promise((r) => setTimeout(r, 900)); // 等 bindgen 生成 + 诊断推送

    const cDiags = diagnostics.filter((d) => d.uri === cURI).flatMap((d) => d.diagnostics ?? []);
    check(cDiags.length === 0, "C 导入：编辑器里解析成功、没有诊断", JSON.stringify(cDiags).slice(0, 240));

    const cAt = lineOf(cText, "mathkit.mathkit_add");
    if (!cAt) throw new Error("cinterop 样例里找不到 mathkit_add 调用");
    const cComp = await request("textDocument/completion", {
      textDocument: { uri: cURI },
      position: { line: cAt.line, character: cAt.character + "mathkit".length + 1 },
    });
    const cLabels = (cComp.result?.items ?? []).map((i) => i.label);
    check(
      cLabels.includes("mathkit_add") && cLabels.includes("MATHKIT_MAX"),
      "C 导入：`mathkit.` 之后补全出 C 函数与常量",
      cLabels.slice(0, 12).join(","),
    );

    const cDef = await request("textDocument/definition", {
      textDocument: { uri: cURI },
      position: { line: cAt.line, character: cAt.character + "mathkit.".length + 2 },
    });
    const cDefs = cDef.result ?? [];
    check(cDefs.length > 0, "C 导入：C 函数能跳转");
    check((cDefs[0]?.uri ?? "").endsWith(".ail"), "C 导入：跳转目标是 bindgen 生成的 .ail", cDefs[0]?.uri);

    const cHover = await request("textDocument/hover", {
      textDocument: { uri: cURI },
      position: { line: cAt.line, character: cAt.character + "mathkit.".length + 2 },
    });
    check(
      JSON.stringify(cHover.result ?? {}).includes("C 包"),
      "C 导入：悬停说明来自哪个 C 包",
      JSON.stringify(cHover.result ?? {}).slice(0, 200),
    );

    const cLinks = await request("textDocument/documentLink", { textDocument: { uri: cURI } });
    const cLinkList = cLinks.result ?? [];
    check(
      cLinkList.some((l) => (l.target ?? "").endsWith(".ail")),
      "C 导入：import 链接指向绑定文件",
      JSON.stringify(cLinkList).slice(0, 200),
    );
    notify("textDocument/didClose", { textDocument: { uri: cURI } });
  }

  await request("shutdown", null);
  notify("exit", {});
} catch (err) {
  failures++;
  console.error("✘ 运行失败：" + err.message);
} finally {
  setTimeout(() => child.kill(), 200);
}

console.log("");
if (failures > 0) {
  console.error(`语言服务器冒烟失败：${failures} 项`);
  process.exit(1);
}
console.log("语言服务器冒烟通过");
