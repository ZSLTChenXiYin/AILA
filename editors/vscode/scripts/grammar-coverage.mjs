// 语法高亮的「无遗漏」检查：
// 把仓库里所有 .aila 源文件 + 一份由 lexer 关键字表生成的构造样例喂给真正的
// TextMate 引擎，任何一段非空白 token 只要拿不到 scope（只剩 source.aila），
// 就算「退化成纯文本」→ 直接失败。
//
// 另外单独断言：internal/lexer 里的每个关键字/内置类型名都拿到了对应 scope。
// 关键字表是从 Go 源码里抽出来的，所以加了新关键字而忘了改文法，这里会红。
//
// 用法：node scripts/grammar-coverage.mjs [--list]
import { readFileSync, readdirSync, statSync } from "node:fs";
import { dirname, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import vsctmDefault from "vscode-textmate";
const vsctm = vsctmDefault.default ?? vsctmDefault;
import onigurumaDefault from "vscode-oniguruma";

const here = dirname(fileURLToPath(import.meta.url));
const extRoot = resolve(here, "..");
const repoRoot = resolve(extRoot, "..", "..");

const grammarPath = join(extRoot, "syntaxes", "aila.tmLanguage.json");
const grammarSource = readFileSync(grammarPath, "utf8");
const grammarJSON = JSON.parse(grammarSource);
const bindingPath = join(extRoot, "syntaxes", "aila-binding.tmLanguage.json");
const bindingSource = readFileSync(bindingPath, "utf8");
const bindingJSON = JSON.parse(bindingSource);

const wasmBin = readFileSync(join(extRoot, "node_modules", "vscode-oniguruma", "release", "onig.wasm")).buffer;
const oniguruma = onigurumaDefault;
const onigLib = oniguruma.loadWASM(wasmBin).then(() => ({
  createOnigScanner: (sources) => new oniguruma.OnigScanner(sources),
  createOnigString: (s) => new oniguruma.OnigString(s),
}));

const registry = new vsctm.Registry({
  onigLib,
  loadGrammar: async (scopeName) => {
    if (scopeName === grammarJSON.scopeName) return vsctm.parseRawGrammar(grammarSource, grammarPath);
    if (scopeName === bindingJSON.scopeName) return vsctm.parseRawGrammar(bindingSource, bindingPath);
    return null;
  },
});
const grammar = await registry.loadGrammar(grammarJSON.scopeName);
const bindingGrammar = await registry.loadGrammar(bindingJSON.scopeName);

// tokenize：逐行推进 ruleStack，返回 [{ line, text, scopes }]，scopes 已去掉根 scope。
function tokenizeWith(gram, rootScope, source) {
  let ruleStack = vsctm.INITIAL;
  const out = [];
  source.split("\n").forEach((line, i) => {
    const result = gram.tokenizeLine(line, ruleStack);
    for (const t of result.tokens) {
      const text = line.slice(t.startIndex, t.endIndex);
      if (!text.trim()) continue;
      out.push({ line: i + 1, text, scopes: t.scopes.filter((s) => s !== rootScope) });
    }
    ruleStack = result.ruleStack;
  });
  return out;
}

const tokenize = (source) => tokenizeWith(grammar, grammarJSON.scopeName, source);

// ---- 1. 从 Go lexer 里抽关键字 / 内置类型 / 标点表 -------------------------------
function goStringSet(name) {
  const lexerPath = join(repoRoot, "internal", "lexer", "lexer.go");
  const src = readFileSync(lexerPath, "utf8");
  const start = src.indexOf(`var ${name} = map[string]bool{`);
  if (start < 0) throw new Error(`在 internal/lexer/lexer.go 里找不到 ${name}`);
  const close = src.indexOf("\n}", start);
  const body = src.slice(start, close);
  return [...body.matchAll(/"([^"]+)"/g)].map((m) => m[1]);
}

function goPunctLists() {
  const src = readFileSync(join(repoRoot, "internal", "lexer", "lexer.go"), "utf8");
  const grab = (re) => [...(src.match(re) ?? [])].flatMap((m) => [...m.matchAll(/"([^"]+)"/g)].map((x) => x[1]));
  const p3 = grab(/var punct3 = \[\]string\{[^}]*\}/);
  const p2 = grab(/var punct2 = \[\]string\{[^}]*\}/);
  const p1m = src.match(/var punct1 = "([^"]*)"/);
  const p1 = p1m ? [...p1m[1]] : [];
  return { p3, p2, p1 };
}

const keywords = goStringSet("Keywords");
const typeWords = goStringSet("TypeWords");
const { p3: punct3, p2: punct2, p1: punct1 } = goPunctLists();
if (!keywords.length || !typeWords.length) throw new Error("关键字表抽取失败");

// ---- 2. 组装样例：每个关键字/类型/标点都放进一个真实上下文 ------------------------
const sink = [];
sink.push("// 覆盖样例：关键字、类型、标点、字面量、属性、效果子句、成员访问");
sink.push("/* 块注释 */");
sink.push('import std.io; import ch.cstdio as c; import mylib.math.ops as ops;');
for (const kw of keywords) {
  // 每个关键字单独一行；declarations 类关键字需要名字才会被「声明」规则吃掉，
  // 其余关键字裸出现也应该拿到 keyword scope。
  sink.push(`${kw} kv_${kw};`);
}
for (const tw of typeWords) sink.push(`let tw_${tw}: ${tw};`);
// 标点：三个一组塞进表达式上下文，保证每个标点都被匹配到。
sink.push(`let a: I32 = b ${punct1.join(" ")} c;`);
sink.push(`let d: I32 = b ${punct2.join(" ")} c;`);
sink.push(`let e: I32 = b ${punct3.join(" ")} c;`);
sink.push("func sink_main() -> I32 {");
sink.push("    let p: Point = Point { x: 1, y: 2 };");
sink.push('    let s: Text = "hi \\n";');
sink.push("    let ch: C32 = 'x';");
sink.push("    let h: U32 = 0xFF_00;");
sink.push("    let f: F64 = 1.5e3;");
sink.push("    let n: USize = #size(Point) + #align(Point);");
sink.push("    let z: ?I32 = None;");
sink.push("    let buf: [U8] = [0; 4]@static;");
  sink.push("    let r: [U8]@region[buf]@mut[buf] = buf;");
sink.push("    if n > 0 { i += 1; } else { i -= 1; }");
sink.push("    match z { Some(v) => v, None => 0 }");
sink.push("    for i in 0..4 { out.write(i); }");
sink.push("    coroutine c1 { yield 1; }");
sink.push("    unsafe { trap; }");
sink.push("    return x.y.z(a, b);");
sink.push("}");

// ---- 3. 收集仓库里的 .aila 源文件 ----------------------------------------------
function walk(dir, acc, ext) {
  let entries;
  try {
    entries = readdirSync(dir, { withFileTypes: true });
  } catch {
    return acc;
  }
  for (const ent of entries) {
    if (ent.name === "node_modules" || ent.name === "build" || ent.name.startsWith(".")) continue;
    const p = join(dir, ent.name);
    if (ent.isDirectory()) walk(p, acc, ext);
    else if (ent.name.endsWith(ext)) acc.push(p);
  }
  return acc;
}
const roots = ["tests", "examples", "stdlib"].map((d) => join(repoRoot, d));
const files = roots.flatMap((r) => walk(r, [], ".aila"));
const bindFiles = roots.flatMap((r) => walk(r, [], ".ail"));
if (!files.length) throw new Error("没找到任何 .aila 源文件，覆盖检查没有意义");
if (!bindFiles.length) throw new Error("没找到任何 .ail 绑定文件，覆盖检查没有意义");

// ---- 4. 检查：非空白 token 必须有 scope ---------------------------------------
const failures = [];
let tokenCount = 0;
function check(label, source, gram, rootScope) {
  const tokens = gram ? tokenizeWith(gram, rootScope, source) : tokenize(source);
  for (const tok of tokens) {
    tokenCount++;
    if (tok.scopes.length === 0) {
      failures.push(`${label}:${tok.line} 无高亮 → ${JSON.stringify(tok.text)}`);
    }
  }
  return tokens;
}

const sinkTokens = check("syntaxes/coverage-sink", sink.join("\n"));
for (const f of files) check(relative(repoRoot, f), readFileSync(f, "utf8"));
for (const f of bindFiles) {
  const toks = tokenizeWith(bindingGrammar, bindingJSON.scopeName, readFileSync(f, "utf8"));
  for (const tok of toks) {
    tokenCount++;
    if (tok.scopes.length === 0) failures.push(`${relative(repoRoot, f)}:${tok.line} 无高亮 → ${JSON.stringify(tok.text)}`);
  }
}

// ---- 5. 检查：关键字/类型必须有语义 scope ------------------------------------
function scopeOfWord(word) {
  const hits = sinkTokens.filter((t) => t.text.trim() === word);
  return hits.map((t) => t.scopes.join(" ")).filter(Boolean);
}
const badKeywords = keywords.filter((kw) => {
  const scopes = scopeOfWord(kw);
  return !scopes.length || !scopes.some((s) => /keyword|constant\.language/.test(s));
});
const badTypes = typeWords.filter((tw) => {
  const scopes = scopeOfWord(tw);
  return !scopes.length || !scopes.some((s) => /support\.type/.test(s));
});

// ---- 6. 报告 ------------------------------------------------------------------
if (process.argv.includes("--list")) {
  console.log(`扫描 ${files.length} 个 .aila + ${bindFiles.length} 个 .ail + 1 份构造样例，共 ${tokenCount} 个 token`);
}

let ok = true;
if (failures.length) {
  ok = false;
  console.error(`✘ 有 ${failures.length} 个 token 退化成纯文本：`);
  for (const f of failures.slice(0, 80)) console.error(`   ${f}`);
  if (failures.length > 80) console.error(`   …… 还有 ${failures.length - 80} 条`);
}
if (badKeywords.length) {
  ok = false;
  console.error(`✘ 这些关键字没有拿到 keyword scope：${badKeywords.join(" ")}`);
}
if (badTypes.length) {
  ok = false;
  console.error(`✘ 这些内置类型没有拿到 support.type scope：${badTypes.join(" ")}`);
}
if (!ok) {
  console.error("\n语法高亮覆盖检查失败");
  process.exit(1);
}
console.log(
  `语法高亮覆盖检查通过：${files.length} 个源文件 + 构造样例，${tokenCount} 个 token 全部有 scope；` +
    `${keywords.length} 个关键字、${typeWords.length} 个内置类型、${punct1.length + punct2.length + punct3.length} 个标点全部可识别`,
);
