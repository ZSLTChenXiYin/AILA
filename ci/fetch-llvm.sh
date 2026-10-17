#!/usr/bin/env bash
# 拉取并裁剪「免安装解压即用」的 Clang/LLVM（docs/specs/0003 §4）
#
# Linux：用 apt.llvm.org 的官方 deb 解包（不用系统 apt 安装，保持可搬运）
# macOS：用 Homebrew 的 llvm@<ver>（runner 自带 brew）
#
# 首次 release 前必须做（脚本里标了 TODO）：填 sha256 / 在目标平台跑 aila doctor。
set -euo pipefail

LLVM_VERSION="${LLVM_VERSION:-18.1.8}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK="$ROOT/ci-work"
TOOLCHAIN="$ROOT/toolchain"
mkdir -p "$WORK" "$TOOLCHAIN"

case "$(uname -s)" in
  Linux)  PLATFORM="linux-x64";;
  Darwin) PLATFORM="macos-$(uname -m)";;
  *) echo "不支持的系统：$(uname -s)" >&2; exit 1;;
esac
echo "== 拉取 LLVM $LLVM_VERSION ($PLATFORM) =="

if [ "$(uname -s)" = "Linux" ]; then
  CODENAME="$(. /etc/os-release && echo "$VERSION_CODENAME")"
  BASE="https://apt.llvm.org/${CODENAME}"
  TRIPLE="$(cc -dumpmachine 2>/dev/null || echo x86_64-unknown-linux-gnu)"
  DEST="$TOOLCHAIN/$TRIPLE"
  mkdir -p "$DEST/bin"
  # TODO(release): 用 apt-get download --print-uris 固定 sha256 并校验
  curl -fsSL "$BASE/llvm-toolchain-${CODENAME}-${LLVM_VERSION%.*}/dists/${CODENAME}/main/binary-amd64/Packages.gz" \
    -o "$WORK/Packages.gz" || echo "（Packages 索引拉取失败，退化为 deb 直链拼接）"
  for pkg in "clang-${LLVM_VERSION}" "libclang-cpp-${LLVM_VERSION}" "llvm-${LLVM_VERSION}" "lld-${LLVM_VERSION}"; do
    echo "-- $pkg"
    ( cd "$WORK" && apt-get download "$pkg" 2>/dev/null ) || \
      ( cd "$WORK" && curl -fsSLO "$BASE/pool/main/l/llvm-toolchain-${LLVM_VERSION%.*}/$(echo "$pkg" | tr -d - )_${LLVM_VERSION}_amd64.deb" 2>/dev/null ) || true
  done
  for deb in "$WORK"/*.deb; do
    [ -e "$deb" ] || continue
    dpkg-deb -x "$deb" "$WORK/llvm"
  done
  for name in clang llc lld ld.lld llvm-ar llvm-objcopy llvm-objdump llvm-nm; do
    for cand in "$WORK/llvm/usr/bin/$name" "$WORK/llvm/usr/lib/llvm-${LLVM_VERSION%.*}/bin/$name"; do
      if [ -x "$cand" ]; then cp "$cand" "$DEST/bin/" && break; fi
    done
  done
  for lib in "$WORK/llvm/usr/lib/llvm-${LLVM_VERSION%.*}/lib"/libclang*.so* "$WORK/llvm/usr/lib/x86_64-linux-gnu"/libclang*.so*; do
    [ -e "$lib" ] || continue
    mkdir -p "$DEST/lib" && cp -P "$lib" "$DEST/lib/"
  done
  RESOURCE="$WORK/llvm/usr/lib/llvm-${LLVM_VERSION%.*}/lib/clang"
  [ -d "$RESOURCE" ] && mkdir -p "$DEST/lib" && cp -r "$RESOURCE" "$DEST/lib/clang"
  [ -f "$WORK/llvm/usr/share/doc/llvm-${LLVM_VERSION}/copyright" ] && \
    cp "$WORK/llvm/usr/share/doc/llvm-${LLVM_VERSION}/copyright" "$DEST/LICENSE.TXT"
else
  TRIPLE="$(clang -print-target-triple 2>/dev/null || echo "$(uname -m)-apple-darwin")"
  DEST="$TOOLCHAIN/$TRIPLE"
  mkdir -p "$DEST/bin"
  brew update >/dev/null 2>&1 || true
  brew install "llvm@${LLVM_VERSION%.*}" >/dev/null 2>&1 || brew install llvm >/dev/null
  LLVM_PREFIX="$(brew --prefix "llvm@${LLVM_VERSION%.*}" 2>/dev/null || brew --prefix llvm)"
  for name in clang llc lld ld.lld llvm-ar llvm-objcopy llvm-objdump llvm-nm; do
    [ -x "$LLVM_PREFIX/bin/$name" ] && cp "$LLVM_PREFIX/bin/$name" "$DEST/bin/"
  done
  [ -d "$LLVM_PREFIX/lib/clang" ] && mkdir -p "$DEST/lib" && cp -r "$LLVM_PREFIX/lib/clang" "$DEST/lib/clang"
  [ -f "$LLVM_PREFIX/LICENSE.TXT" ] && cp "$LLVM_PREFIX/LICENSE.TXT" "$DEST/LICENSE.TXT"
  # dylib 依赖一并带上，保持「解压即用」
  for lib in "$LLVM_PREFIX/lib"/libLLVM*.dylib "$LLVM_PREFIX/lib"/libclang*.dylib; do
    [ -e "$lib" ] || continue
    mkdir -p "$DEST/lib" && cp -P "$lib" "$DEST/lib/"
  done
fi

echo "$LLVM_VERSION ($PLATFORM, trimmed)" > "$DEST/VERSION"
echo "== 完成：$DEST =="
ls -la "$DEST/bin" || true
