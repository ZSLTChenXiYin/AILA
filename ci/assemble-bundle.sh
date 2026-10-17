#!/usr/bin/env bash
# 组装发布目录（POSIX 版，与 assemble-bundle.ps1 等价）
# 用法：assemble-bundle.sh <platform> <version>
set -euo pipefail

PLATFORM="${1:-linux-x64}"
VERSION="${2:-0.1.0}"
MANAGER="${3:-go}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="$ROOT/dist"
BUNDLE="$OUT/aila-$VERSION-$PLATFORM"

if [ "$MANAGER" != "go" ]; then
  echo "管理器只有 Go 单文件形态（Python 前端已在 G4 删除）；$MANAGER 不再支持" >&2
  exit 2
fi

rm -rf "$BUNDLE"
mkdir -p "$BUNDLE"
echo "== 组装 $BUNDLE（manager=go）=="

case "$PLATFORM" in
  windows*) GOOS=windows ;;
  macos*)   GOOS=darwin ;;
  *)        GOOS=linux ;;
esac
case "$PLATFORM" in
  *arm64) GOARCH=arm64 ;;
  *)      GOARCH=amd64 ;;
esac

( cd "$ROOT" && CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
    go build -trimpath -ldflags "-s -w -X github.com/ailang/aila/internal/cli.Version=$VERSION" \
    -o "$BUNDLE/aila" ./cmd/aila )
if [ "$GOOS" = "windows" ]; then
  mv "$BUNDLE/aila" "$BUNDLE/aila.exe"
fi

mkdir -p "$BUNDLE/share/aila"
cp -r "$ROOT/runtime" "$BUNDLE/share/aila/runtime"
cp -r "$ROOT/stdlib" "$BUNDLE/share/aila/stdlib"
cp -r "$ROOT/share/aila/data" "$BUNDLE/share/aila/data"

if [ -d "$ROOT/toolchain" ]; then
  cp -r "$ROOT/toolchain" "$BUNDLE/toolchain"
else
  echo "警告：没有 toolchain/ 目录（先跑 ci/fetch-llvm.sh）" >&2
fi

cp -r "$ROOT/examples" "$BUNDLE/examples"
cp "$ROOT/README.md" "$BUNDLE/"
cp "$ROOT/LICENSE" "$BUNDLE/"
cp "$ROOT/NOTICE" "$BUNDLE/"
echo "$VERSION" > "$BUNDLE/VERSION"
cat > "$BUNDLE/NOTICES.md" <<EOF
AILA $VERSION ($PLATFORM)

本发布包内含 llvm-mingw（clang + lld + MinGW-w64 运行时），
LLVM/Clang 部分按 Apache-2.0 WITH LLVM-exception 再分发，
许可证原文见 toolchain/<triple>/LICENSE.TXT。AILA 自身同样以 Apache-2.0 发布。

工具链解析规则（docs/specs/0003 §2）：默认只使用本包 toolchain/ 下的 clang/llc；
只有当项目清单显式写 [build.toolchain] cc = "system:..." 时才会去系统里找。

自检：./aila doctor     编译并运行：./aila run examples/hello
EOF

echo "== 打包 =="
( cd "$OUT" && zip -qr "aila-$VERSION-$PLATFORM.zip" "aila-$VERSION-$PLATFORM" )
echo "产物：$OUT/aila-$VERSION-$PLATFORM.zip"
