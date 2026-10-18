// 语言服务器的端到端冒烟：真起 `aila lsp` 子进程，按 LSP 协议握手并发请求。
//
// 为什么要有它：编辑器里调不通时（例如 Content-Length 分帧、range 换算、模块映射），
// 只有真跑一遍协议才能发现；纯单元测试容易漏掉传输层。
//
// 用法：node scripts/lsp-smoke.mjs [工程目录]
// 默认对仓库 examples/result 工程跑：initialize → didOpen → documentSymbol → definition → hover → shutdown。
import { spawn } from "node:child_process";
import { readFileSync, existsSync } from "node:fs";
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
