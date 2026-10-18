// 语法高亮的冒烟测试：用 vscode-textmate + vscode-oniguruma 真正跑一遍文法。
//
// 用法：node scripts/grammar-smoke.mjs [文件...]
// 默认对 examples/hello/main.aila 与 snippets 里出现过的主要语法点做断言。
//
// 这个脚本存在的理由：`.tmLanguage.json` 写错了 VSCode 只会「悄悄不高亮」，
// 用 CI 跑一遍断言才能保证关键字/类型/字符串/属性/效果子句真的被识别。
import { readFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import vsctmDefault from "vscode-textmate";
const vsctm = vsctmDefault.default ?? vsctmDefault;
import onigurumaDefault from "vscode-oniguruma";

const here = dirname(fileURLToPath(import.meta.url));
const extRoot = resolve(here, "..");

const grammarPath = join(extRoot, "syntaxes", "aila.tmLanguage.json");
const grammarJSON = JSON.parse(readFileSync(grammarPath, "utf8"));

const wasmPath = join(extRoot, "node_modules", "vscode-oniguruma", "release", "onig.wasm");
const wasmBin = readFileSync(wasmPath).buffer;
const oniguruma = onigurumaDefault;
const onigLib = oniguruma.loadWASM(wasmBin).then(() => ({
  createOnigScanner: (sources) => new oniguruma.OnigScanner(sources),
  createOnigString: (s) => new oniguruma.OnigString(s),
}));

const registry = new vsctm.Registry({
  onigLib,
  loadGrammar: async (scopeName) => {
    if (scopeName === grammarJSON.scopeName) return vsctm.parseRawGrammar(
      readFileSync(grammarPath, "utf8"), grammarPath);
    return null;
  },
});

// tokenize 返回 [[text, scopes], ...]，只保留最内层的两个 scope 方便断言。
async function tokenize(source) {
  const grammar = await registry.loadGrammar(grammarJSON.scopeName);
  let ruleStack = vsctm.INITIAL;
  const out = [];
  for (const line of source.split("\n")) {
    const result = grammar.tokenizeLine(line, ruleStack);
    for (const t of result.tokens) {
      const text = line.slice(t.startIndex, t.endIndex);
      if (!text.trim()) continue;
      out.push([text, t.scopes.slice(1)]); // scopes[0] 恒为 source.aila
    }
    ruleStack = result.ruleStack;
  }
  return out;
}

// 断言：某个片段至少有一个 token 的 scope 里包含 want
// 断言：某个前缀（如整行注释）命中某个 scope
function expectPrefix(tokens, prefix, want, label) {
  const hit = tokens.find(([text, scopes]) =>
    text.trim().startsWith(prefix) && scopes.some((s) => s.includes(want)));
  if (!hit) {
    console.error(`✘ ${label}：期望以 \`${prefix}\` 开头的 token 命中 ${want}`);
    return false;
  }
  console.log(`✔ ${label}：${hit[0].trim().slice(0, 18)} → ${hit[1].join(" ")}`);
  return true;
}

function expectScope(tokens, snippet, want, label) {
  const hit = tokens.find(([text, scopes]) =>
    text.trim() === snippet && scopes.some((s) => s.includes(want)));
  if (!hit) {
    const nearby = tokens.filter(([t]) => t.trim() === snippet)
      .map(([, scopes]) => scopes.join(" "));
    console.error(`✘ ${label}：期望 \`${snippet}\` 命中 ${want}`);
    if (nearby.length) console.error(`   实际 scopes：${nearby.join(" | ")}`);
    return false;
  }
  console.log(`✔ ${label}：${snippet} → ${hit[1].join(" ")}`);
  return true;
}

const sample = `// 注释行
import std.io;
import mylib.math.ops as ops;

public func twice(v: I32) -> I32;

#[intrinsic] func write(bytes: [U8], n: USize);
#[c] public func c_add(a: I32, b: I32) -> I32;

public struct Point {
    x: I32,
    y: F32,
};

public enum Color { Red, Green, Blue };

module {
    func demo(buf: [U8]) @region[buf] @mut[buf] {
        var nums: [U8] = [0; 4]@static;
        let n: USize = #size(Point);
        let m: USize = #align(Color);
        var i: I32 = 0;
        for!, i < 3; i = i + 1 {
            if i == 0 { continue; }
        }
        return;
    }
}
`;

const tokens = await tokenize(sample);
let ok = true;
ok = expectPrefix(tokens, "//", "comment.line", "行注释") && ok;
ok = expectScope(tokens, "import", "keyword.control.import", "import 关键字") && ok;
ok = expectScope(tokens, "std", "entity.name.namespace.root", "导入根 std") && ok;
ok = expectScope(tokens, "ops", "variable.other.alias", "as 别名") && ok;
ok = expectScope(tokens, "public", "keyword.declaration", "public 关键字") && ok;
ok = expectScope(tokens, "func", "keyword.declaration", "func 关键字") && ok;
ok = expectScope(tokens, "twice", "entity.name.function", "函数名") && ok;
ok = expectScope(tokens, "I32", "support.type.primitive", "内建类型 I32") && ok;
ok = expectScope(tokens, "#[intrinsic]", "entity.name.function.decorator", "#[intrinsic] 属性") && ok;
ok = expectScope(tokens, "#[c]", "entity.name.function.decorator", "#[c] 属性") && ok;
ok = expectScope(tokens, "#size", "support.function.builtin", "#size 内建查询") && ok;
ok = expectScope(tokens, "@region", "keyword.operator.effect", "@region 效果子句") && ok;
ok = expectScope(tokens, "@mut", "keyword.operator.effect", "@mut 效果子句") && ok;
ok = expectScope(tokens, "@static", "keyword.operator.allocate", "@static 分配") && ok;
ok = expectScope(tokens, "struct", "keyword.declaration", "struct 关键字") && ok;
ok = expectScope(tokens, "Point", "entity.name.type", "类型名 Point") && ok;
ok = expectScope(tokens, "module", "keyword.declaration", "module 关键字") && ok;
ok = expectScope(tokens, "for", "keyword.control", "for 关键字") && ok;

if (!ok) {
  console.error("\n语法高亮冒烟失败");
  process.exit(1);
}
console.log(`\n语法高亮冒烟通过（${tokens.length} 个 token）`);
