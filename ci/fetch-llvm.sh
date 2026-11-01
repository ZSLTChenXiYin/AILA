#!/usr/bin/env bash
# 拉取并裁剪「免安装解压即用」的工具链（docs/specs/0003 §1.1 路线 A：llvm-mingw）
#
# 与 ci/fetch-llvm.ps1 等价，只是宿主换成 Linux / macOS：
#   Linux  → llvm-mingw-<tag>-<rt>-ubuntu-22.04-x86_64.tar.xz
#   macOS  → llvm-mingw-<tag>-<rt>-macos-universal.tar.xz
# 裁剪规则与 ps1 完全一致：clang 系 + clang 资源目录（只留宿主架构的 builtins）+ C 头，
# 丢掉 lldb / Python / clangd / sanitizer / C++ 头。
#
# 产出（rfcs/0006 注册表布局）：toolchains/clang/<LLVM 版本>/（单一根目录），
# 并把本平台条目合并进 toolchains/toolchains.json 的 available 段（有 python3 时）。
#
# 用法：
#   ./ci/fetch-llvm.sh                                        # 默认 UCRT、最新 tag
#   TAG=20261006 RUNTIME=ucrt ./ci/fetch-llvm.sh
#   MIRROR=https://dl.espressif.cn/github_assets/ ./ci/fetch-llvm.sh   # 镜像前缀
#   LLVM_MINGW_SHA256=<sha256> ./ci/fetch-llvm.sh             # 发行前必须校验
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK="$ROOT/ci-work"
TOOLCHAINS="$ROOT/toolchains"
RUNTIME="${RUNTIME:-ucrt}"
MIRROR="${MIRROR:-}"

os="$(uname -s)"
arch="$(uname -m)"
case "$os" in
  Linux)
    host="ubuntu-22.04-x86_64"
    [ "$arch" = "aarch64" ] && host="ubuntu-22.04-aarch64"
    LLVM_ARCH="x86_64"
    [ "$arch" = "aarch64" ] && LLVM_ARCH="aarch64"
    ;;
  Darwin)
    host="macos-universal"
    LLVM_ARCH="$(uname -m)"
    ;;
  *)
    echo "不支持的系统：$os（Windows 用 ci/fetch-llvm.ps1）" >&2
    exit 1
    ;;
esac

if [ -z "${TAG:-}" ]; then
  api="https://api.github.com/repos/mstorsjo/llvm-mingw/releases/latest"
  [ -n "$MIRROR" ] && api="${MIRROR%/}/$api"
  TAG="$(curl -fsSL "$api" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -1)"
  [ -n "$TAG" ] || { echo "取不到最新 tag，请显式设 TAG=..." >&2; exit 1; }
fi

ASSET="llvm-mingw-$TAG-$RUNTIME-$host.tar.xz"
URL="https://github.com/mstorsjo/llvm-mingw/releases/download/$TAG/$ASSET"
[ -n "$MIRROR" ] && URL="${MIRROR%/}/$URL"

mkdir -p "$WORK"
echo "== 拉取 llvm-mingw $TAG ($RUNTIME, $host) =="
ARCHIVE="$WORK/$ASSET"
if [ -f "$ARCHIVE" ]; then
  echo "   已有缓存：$ARCHIVE"
else
  echo "   $URL"
  curl -fL --retry 3 -o "$ARCHIVE" "$URL"
fi

if [ -n "${LLVM_MINGW_SHA256:-}" ]; then
  actual="$(shasum -a 256 "$ARCHIVE" 2>/dev/null | cut -d' ' -f1 || sha256sum "$ARCHIVE" | cut -d' ' -f1)"
  if [ "$actual" != "$LLVM_MINGW_SHA256" ]; then
    echo "sha256 不匹配：得到 $actual，期望 $LLVM_MINGW_SHA256" >&2
    exit 1
  fi
  echo "   sha256 校验通过"
else
  echo "   警告：未提供 LLVM_MINGW_SHA256（发行前请固定下来）" >&2
fi

echo "== 解包 =="
EXTRACT="$WORK/llvm-mingw"
if [ -d "$EXTRACT" ] && [ -n "$(ls -A "$EXTRACT" 2>/dev/null)" ]; then
  echo "   已有解包缓存：$EXTRACT"
else
  rm -rf "$EXTRACT" && mkdir -p "$EXTRACT"
  tar -xf "$ARCHIVE" -C "$EXTRACT"
fi
SRC="$(find "$EXTRACT" -maxdepth 1 -mindepth 1 -type d | head -1)"
[ -n "$SRC" ] || { echo "解包后找不到 llvm-mingw 根目录" >&2; exit 1; }
echo "   源：$SRC"

# LLVM 版本决定安装目录名（toolchains/clang/<ver>/）
LLVM_VER="$("$SRC/bin/clang" --version 2>/dev/null | sed -n 's/^.*clang version \([0-9][0-9.]*\).*/\1/p' | head -1)"
[ -n "$LLVM_VER" ] || { echo "取不到 clang 版本号（$SRC/bin/clang --version 失败）" >&2; exit 1; }
DEST="$TOOLCHAINS/clang/$LLVM_VER"

echo "== 裁剪到 $DEST =="
rm -rf "$DEST" && mkdir -p "$DEST/bin"

# 1) bin：丢掉调试器 / Python / 编辑器工具 / sanitizer dll；保留 clang 自身依赖的
#    libc++ / libunwind / libwinpthread（AILA 只编 C，但 clang 是用 libc++ 构建的）
for f in "$SRC"/bin/*; do
  name="$(basename "$f")"
  case "$name" in
    lldb*|python*|idle3*|pydoc*|clangd*|clang-tidy*|run-clang-tidy*|scan-build*|analyze-build*|intercept-build*|clang-format*|clang-scan-deps*|liblldb*|libpython*|libffi*|*asan*.dll|*.mustache) continue ;;
  esac
  cp "$f" "$DEST/bin/"
done

# 1b) clang 自身的共享库（Linux: lib/*.so，macOS: lib/*.dylib）：
#     clang 的 rpath 是 $ORIGIN/../lib，运行期按 SONAME 加载 libclang-cpp / libLLVM；
#     Windows 版这些是 bin/ 里的 dll（上面的 bin 拷贝已带上），POSIX 版必须从这里补。
#     只拷版本化真身（symlink 只在链接期有用，运行不需要，还规避跨平台拷贝问题）
mkdir -p "$DEST/lib"
for f in "$SRC"/lib/libclang-cpp.so.* "$SRC"/lib/libLLVM.so.* \
         "$SRC"/lib/libclang-cpp.*.dylib "$SRC"/lib/libLLVM.*.dylib; do
  [ -f "$f" ] || continue
  [ -L "$f" ] && continue
  cp "$f" "$DEST/lib/"
done

# 2) clang 资源目录：内建头 + 宿主架构的 builtins/profile 运行时
#    （llvm-mingw 的 compiler-rt 目标库目录名恒为 windows/，与宿主平台无关）
RES_ROOT="$SRC/lib/clang"
RES_VER="$(ls "$RES_ROOT" | head -1)"
RES_SRC="$RES_ROOT/$RES_VER"
RES_DEST="$DEST/lib/clang/$RES_VER"
mkdir -p "$RES_DEST"
cp -r "$RES_SRC/include" "$RES_DEST/include"
for sub in windows linux; do
  [ -d "$RES_SRC/lib/$sub" ] || continue
  for f in "$RES_SRC/lib/$sub"/*-"$LLVM_ARCH".a; do
    [ -e "$f" ] || continue
    case "$(basename "$f")" in
      *builtins*|*profile*)
        mkdir -p "$RES_DEST/lib/$sub"
        cp "$f" "$RES_DEST/lib/$sub/"
        ;;
    esac
  done
done

# 3) C 头（sysroot include）：Windows zip 在顶层 include/，Linux/macOS tar 的顶层
#    include 是指向 generic-w64-mingw32/include 的符号链接（空目录视同没有）；
#    丢 C++ 与图形/驱动相关
INC_SRC="$SRC/generic-w64-mingw32/include"
if [ -d "$SRC/include" ] && [ -n "$(ls -A "$SRC/include" 2>/dev/null)" ]; then
  INC_SRC="$SRC/include"
fi
mkdir -p "$DEST/include"
for entry in "$INC_SRC"/*; do
  case "$(basename "$entry")" in
    c++|ddk|gdiplus|GL|KHR|mach-o|wrl) continue ;;
  esac
  cp -r "$entry" "$DEST/include/"
done

# 4) sysroot 链接库（CRT / 导入库 / 运行时）：只留宿主架构的 lib（头已并入上面 include）
SYS_SRC="$SRC/$LLVM_ARCH-w64-mingw32"
if [ -d "$SYS_SRC/lib" ]; then
  mkdir -p "$DEST/$LLVM_ARCH-w64-mingw32"
  cp -r "$SYS_SRC/lib" "$DEST/$LLVM_ARCH-w64-mingw32/lib"
fi

# 5) 许可证与版本（VERSION 首段必须是 LLVM 版本：aila 按它判版本带；无 BOM）
[ -f "$SRC/LICENSE.TXT" ] && cp "$SRC/LICENSE.TXT" "$DEST/LICENSE.TXT"
printf '%s (llvm-mingw %s, %s, trimmed)\n' "$LLVM_VER" "$TAG" "$RUNTIME" > "$DEST/VERSION"

# 6) 把本平台条目合并进 toolchains/toolchains.json 的 available 段（长尾 install 的下载清单）
if command -v python3 >/dev/null 2>&1; then
  LLVM_MINGW_SHA256="${LLVM_MINGW_SHA256:-}" SIZE_KB="$(du -k "$ARCHIVE" | cut -f1)" \
  python3 - "$TOOLCHAINS/toolchains.json" "$LLVM_VER" "$URL" <<'PYEOF'
import json, os, sys
path, version, url = sys.argv[1], sys.argv[2], sys.argv[3]
sha, size_kb = os.environ.get("LLVM_MINGW_SHA256", ""), int(os.environ.get("SIZE_KB", "0"))
size = f"{size_kb // 1024}MB" if size_kb >= 1024 else f"{size_kb}KB"
with open(path, encoding="utf-8") as f:
    doc = json.load(f)
doc["available"] = [e for e in doc.get("available", []) if e.get("family") != "clang" or e.get("url") == url]
doc["available"].append({"family": "clang", "version": version, "url": url,
                         "sha256": sha, "size": size if sha else f"（待固定 sha256；约 {size}）"})
with open(path, "w", encoding="utf-8", newline="\n") as f:
    json.dump(doc, f, ensure_ascii=False, indent=2)
    f.write("\n")
PYEOF
  echo "   available 条目已合并：toolchains/toolchains.json"
else
  echo "   警告：无 python3，跳过 toolchains.json available 合并" >&2
fi

echo "== 完成：$DEST（$(du -sh "$DEST" 2>/dev/null | cut -f1)，LLVM $LLVM_VER）=="
ls "$DEST/bin" | head -20
