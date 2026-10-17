# 拉取并裁剪「免安装解压即用」的 Clang/LLVM（docs/specs/0003 §4）
#
# 纪律（designs/0001 §12.5）：
#   * 预拉发生在 CI，不在用户机器上——用户永远从稳定 URL 拿裁剪好的包；
#   * 裁剪规则固定：只留编译器组件 + clang 资源目录 + 链接期 binutils 子集；
#   * 版本与 AILA 发布绑定：LLVM_VERSION 必须落在 src/aila/resolve.py 的 VERSION_BAND 内。
#
# 首次 release 前必须做的两件事（脚本里已标 TODO）：
#   1. 填 sha256；
#   2. 在目标平台跑一次 `aila doctor`（第 6 项冒烟测试）确认 clang/lld 真能用。
$ErrorActionPreference = "Stop"

$LlvmVersion = if ($env:LLVM_VERSION) { $env:LLVM_VERSION } else { "18.1.8" }
$Root = Split-Path -Parent $PSScriptRoot
$Work = Join-Path $Root "ci-work"
$Toolchain = Join-Path $Root "toolchain"
$Triple = "x86_64-w64-windows-gnu"
$Dest = Join-Path $Toolchain $Triple

Write-Host "== 拉取 LLVM $LlvmVersion (windows-x64) =="
New-Item -ItemType Directory -Force -Path $Work | Out-Null
$Asset = "LLVM-$LlvmVersion-win64.exe"
$Url = "https://github.com/llvm/llvm-project/releases/download/llvmorg-$LlvmVersion/$Asset"
$Installer = Join-Path $Work $Asset
if (-not (Test-Path $Installer)) {
    Invoke-WebRequest -Uri $Url -OutFile $Installer
}
# TODO(release): 填官方 SHA256 并启用校验
# $expected = "<sha256>"
# $actual = (Get-FileHash $Installer -Algorithm SHA256).Hash.ToLower()
# if ($actual -ne $expected) { throw "sha256 mismatch: $actual" }

Write-Host "== 解包（7z 可直接展开 NSIS 安装器）=="
$Extract = Join-Path $Work "llvm"
if (Test-Path $Extract) { Remove-Item $Extract -Recurse -Force }
New-Item -ItemType Directory -Force -Path $Extract | Out-Null
& 7z x $Installer "-o$Extract" -y | Out-Null

$Src = Get-ChildItem $Extract -Directory | Select-Object -First 1
if (-not $Src) { throw "解包后找不到 LLVM 根目录" }

Write-Host "== 裁剪 =="
if (Test-Path $Dest) { Remove-Item $Dest -Recurse -Force }
New-Item -ItemType Directory -Force -Path (Join-Path $Dest "bin") | Out-Null

$keep = @("clang.exe", "clang-cl.exe", "llc.exe", "lld-link.exe", "ld.lld.exe",
          "llvm-ar.exe", "llvm-objcopy.exe", "llvm-objdump.exe", "llvm-nm.exe", "llvm-strip.exe")
foreach ($name in $keep) {
    $p = Join-Path $Src.FullName "bin\$name"
    if (Test-Path $p) { Copy-Item $p (Join-Path $Dest "bin\$name") }
}

# clang 资源目录（内建头文件）：bindgen 与 C 编译都需要
$resource = Join-Path $Src.FullName "lib\clang"
if (Test-Path $resource) {
    Copy-Item $resource (Join-Path $Dest "lib\clang") -Recurse
}
$lic = Join-Path $Src.FullName "LICENSE.TXT"
if (Test-Path $lic) { Copy-Item $lic (Join-Path $Dest "LICENSE.TXT") }

Set-Content -Path (Join-Path $Dest "VERSION") -Value "$LlvmVersion (windows-x64, trimmed)" -Encoding utf8
Write-Host "== 完成：$Dest =="
Get-ChildItem (Join-Path $Dest "bin") | Select-Object Name
