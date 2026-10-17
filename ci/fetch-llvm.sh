#!/usr/bin/env bash
# 拉取并裁剪「免安装解压即用」的工具链（docs/specs/0003 §1.1 路线 A：llvm-mingw）
#
# 与 ci/fetch-llvm.ps1 等价，只是宿主换成 Linux / macOS：
#   Linux  → llvm-mingw-<tag>-<rt>-ubuntu-22.04-x86_64.tar.xz
#   macOS  → llvm-mingw-<tag>-<rt>-macos-universal.tar.xz
# 裁剪规则与 ps1 完全一致：clang 系 + clang 资源目录（只留宿主架构的 builtins）+ C 头，
# 丢掉 lldb / Python / clangd / sanitizer / C++ 头。
#
# 用法：
#   ./ci/fetch-llvm.sh                                        # 默认 UCRT、最新 tag
#   TAG=20261006 RUNTIME=ucrt ./ci/fetch-llvm.sh
#   MIRROR=https://ghproxy.net/ ./ci/fetch-llvm.sh            # 镜像前缀
#   LLVM_MINGW_SHA256=<sha256> ./ci/fetch-llvm.sh             # 发行前必须校验
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK="$ROOT/ci-work"
TOOLCHAIN="$ROOT/toolchain"
RUNTIME="${RUNTIME:-ucrt}"
MIRROR="${MIRROR:-}"
TRIPLE="${TRIPLE:-}"

os="$(uname -s)"
arch="$(uname -m)"
case "$os" in
  Linux)
    host="ubuntu-22.04-x86_64"
    [ "$arch" = "aarch64" ] && host="ubuntu-22.04-aarch64"
    ;;
  Darwin)
    host="macos-universal"
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

# 目标三元组：默认用宿主 clang 的 triple；拿不到就按平台兜底
if [ -z "$TRIPLE" ]; then
  TRIPLE="$(clang -print-target-triple 2>/dev/null || true)"
  if [ -z "$TRIPLE" ]; then
    case "$os/$arch" in
      Darwin/arm64) TRIPLE="arm64-apple-darwin" ;;
      Darwin/*)     TRIPLE="x86_64-apple-darwin" ;;
      */aarch64)    TRIPLE="aarch64-unknown-linux-gnu" ;;
      *)            TRIPLE="x86_64-unknown-linux-gnu" ;;
    esac
  fi
fi
DEST="$TOOLCHAIN/$TRIPLE"
LLVM_ARCH="$(echo "$arch" | sed 's/^x86_64$/x86_64/; s/^aarch64$/aarch64/')"

mkdir -p "$WORK"
echo "== 拉取 llvm-mingw $TAG ($RUNTIME, $host) → $TRIPLE =="
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
rm -rf "$EXTRACT" && mkdir -p "$EXTRACT"
tar -xf "$ARCHIVE" -C "$EXTRACT"
SRC="$(find "$EXTRACT" -maxdepth 1 -mindepth 1 -type d | head -1)"
[ -n "$SRC" ] || { echo "解包后找不到 llvm-mingw 根目录" >&2; exit 1; }
echo "   源：$SRC"

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

# 2) clang 资源目录：内建头 + 宿主架构的 builtins/profile 运行时
RES_ROOT="$SRC/lib/clang"
RES_VER="$(ls "$RES_ROOT" | head -1)"
mkdir -p "$DEST/lib/clang/$RES_VER/lib/windows" "$DEST/lib/clang/$RES_VER/lib/linux"
cp -r "$RES_ROOT/$RES_VER/include" "$DEST/lib/clang/$RES_VER/include"
for f in "$RES_ROOT/$RES_VER/lib/windows"/*"-$LLVM_ARCH.a" "$RES_ROOT/$RES_VER/lib/linux"/*"-$LLVM_ARCH.a"; do
  [ -e "$f" ] || continue
  case "$(basename "$f")" in
    *builtins*|*profile*)
      cp "$f" "$DEST/lib/clang/$RES_VER/lib/$(basename "$(dirname "$f")")/"
      ;;
  esac
done

# 3) C 头（sysroot include）：丢 C++ 与图形/驱动相关
mkdir -p "$DEST/include"
for entry in "$SRC"/include/*; do
  case "$(basename "$entry")" in
    c++|ddk|gdiplus|GL|KHR|mach-o|wrl) continue ;;
  esac
  cp -r "$entry" "$DEST/include/"
done

# 4) 宿主平台的 sysroot 库（CRT / 导入库 / 运行时）
case "$os" in
  Linux)
    for d in "$SRC"/*-w64-mingw32 "$SRC"/lib/gcc/*/* ; do
      [ -d "$d" ] || continue
      mkdir -p "$DEST/$(basename "$(dirname "$d")")"
      cp -r "$d" "$DEST/$(basename "$(dirname "$d")")/"
    done
    ;;
  Darwin)
    for d in "$SRC"/*-w64-mingw32; do
      [ -d "$d" ] || continue
      cp -r "$d" "$DEST/"
    done
    ;;
esac

# 5) 许可证与版本（VERSION 首段必须是 LLVM 版本：doctor 按它判版本带）
[ -f "$SRC/LICENSE.TXT" ] && cp "$SRC/LICENSE.TXT" "$DEST/LICENSE.TXT"
LLVM_VER="$("$DEST/bin/clang" --version 2>/dev/null | sed -n 's/^clang version \([0-9][0-9.]*\).*/\1/p' | head -1)"
echo "${LLVM_VER:-unknown} (llvm-mingw $TAG, $RUNTIME, trimmed)" > "$DEST/VERSION"

echo "== 完成：$DEST（$(du -sh "$DEST" 2>/dev/null | cut -f1)，LLVM ${LLVM_VER:-unknown}）=="
ls "$DEST/bin" | head -20
