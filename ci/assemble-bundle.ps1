# 组装发布目录（docs/specs/0003 §5）
#
# 产物目录（解压即用）：
#   aila-<version>-<platform>/
#     aila[.exe]                     ← 管理器 + 编译器前端（Go 单文件，无 cgo）
#     share/aila/runtime/aila_rt.c
#     share/aila/stdlib/*.aila
#     share/aila/data/*.json
#     examples/                     ← 可直接 `aila run examples/hello` 的示例工程
#     toolchain/<triple>/{bin,lib,VERSION,LICENSE.TXT}
#     README.md  LICENSE  NOTICE  VERSION  NOTICES.md
#
# 用法：assemble-bundle.ps1 -Platform windows-x64 -Version 0.1.0
param(
    [string]$Platform = "windows-x64",
    [string]$Version = "0.1.0",
    [string]$Manager = "go",
    [string]$OutDir = ""
)
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
if (-not $OutDir) { $OutDir = Join-Path $Root "dist" }
$Bundle = Join-Path $OutDir "aila-$Version-$Platform"
if (Test-Path $Bundle) { Remove-Item $Bundle -Recurse -Force }
New-Item -ItemType Directory -Force -Path $Bundle | Out-Null

if ($Manager -ne "go") {
    throw "管理器只有 Go 单文件形态（Python 前端已在 G4 删除）；`-Manager $Manager 不再支持"
}
Write-Host "== 组装 $Bundle（manager=go）=="

# 1) 管理器 + 编译器前端（同一个二进制）
Push-Location $Root
$prevGOOS = $env:GOOS
$prevGOARCH = $env:GOARCH
$env:CGO_ENABLED = "0"
$env:GOOS = if ($Platform -like "windows*") { "windows" } else { "linux" }
$env:GOARCH = if ($Platform -like "*arm64") { "arm64" } else { "amd64" }
go build -trimpath -ldflags "-s -w -X github.com/aila-lang/aila/internal/cli.Version=$Version" `
    -o (Join-Path $Bundle "aila$(if ($env:GOOS -eq 'windows') { '.exe' } else { '' })") ./cmd/aila
$env:GOOS = $prevGOOS
$env:GOARCH = $prevGOARCH
Pop-Location

# 2) 运行时 + 标准库 + 数据表（相对目录，AILA_HOME 解析靠它）
New-Item -ItemType Directory -Force -Path (Join-Path $Bundle "share\aila") | Out-Null
Copy-Item (Join-Path $Root "runtime") (Join-Path $Bundle "share\aila\runtime") -Recurse
Copy-Item (Join-Path $Root "stdlib") (Join-Path $Bundle "share\aila\stdlib") -Recurse
Copy-Item (Join-Path $Root "share\aila\data") (Join-Path $Bundle "share\aila\data") -Recurse

# 3) 随包工具链
$tc = Join-Path $Root "toolchain"
if (Test-Path $tc) {
    Copy-Item $tc (Join-Path $Bundle "toolchain") -Recurse
} else {
    Write-Warning "没有 toolchain/ 目录：先跑 ci/fetch-llvm.ps1"
}

# 4) 示例工程（`aila doctor` 的提示与「解压即用」体验都靠它）
Copy-Item (Join-Path $Root "examples") $Bundle -Recurse

# 5) 文档与法律
Copy-Item (Join-Path $Root "README.md") $Bundle
Copy-Item (Join-Path $Root "LICENSE") $Bundle
Copy-Item (Join-Path $Root "NOTICE") $Bundle
Set-Content -Path (Join-Path $Bundle "VERSION") -Value $Version -Encoding utf8
$notice = @"
AILA $Version ($Platform)

本发布包内含 llvm-mingw（clang + lld + MinGW-w64 运行时），
LLVM/Clang 部分按 Apache-2.0 WITH LLVM-exception 再分发，
许可证原文见 toolchain/<triple>/LICENSE.TXT。AILA 自身同样以 Apache-2.0 发布。

工具链解析规则（docs/specs/0003 §2）：默认只使用本包 toolchain/ 下的 clang/llc；
只有当项目清单显式写 [build.toolchain] cc = "system:..." 时才会去系统里找。

自检：./aila doctor     编译并运行：./aila run examples/hello
"@
[System.IO.File]::WriteAllText((Join-Path $Bundle "NOTICES.md"), $notice, [System.Text.Encoding]::UTF8)

Write-Host "== 打包 =="
$zip = Join-Path $OutDir "aila-$Version-$Platform.zip"
if (Test-Path $zip) { Remove-Item $zip -Force }
Compress-Archive -Path $Bundle -DestinationPath $zip
Write-Host "产物：$zip"
