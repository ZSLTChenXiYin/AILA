# 拉取并裁剪「免安装解压即用」的工具链（docs/specs/0003 §1.1 路线 A：llvm-mingw）
#
# 为什么是 llvm-mingw 而不是官方 LLVM 安装包：
#   官方 win64 发行版只有 clang + lld，**不带** MinGW-w64 的 C 运行时 / 头文件 / 导入库；
#   而 AILA 在 Windows 上的三元组是 `x86_64-w64-windows-gnu`，链接期需要 crt2.o、libmingw32、ucrt。
#   llvm-mingw 把这些打成一个自包含包（clang + lld + MinGW 运行时），解压即用才真正成立。
#
# 纪律（designs/0001 §12.5）：
#   * 预拉发生在 CI / 发布机，用户永远从稳定 URL 拿裁剪好的包；
#   * 裁剪规则固定（见下），其余一律不进包；
#   * 版本与 AILA 发布绑定：VERSION 首段是 LLVM 版本，落在 VERSION_BAND 之外时 doctor 报 W8012。
#
# 用法：
#   ./ci/fetch-llvm.ps1                                  # 默认 UCRT、自动取最新 tag
#   ./ci/fetch-llvm.ps1 -Tag 20261006 -Runtime ucrt
#   ./ci/fetch-llvm.ps1 -Mirror https://ghproxy.net/      # GitHub 直连慢时用镜像前缀
#   $env:LLVM_MINGW_SHA256="<sha256>"; ./ci/fetch-llvm.ps1  # 校验（发行前必须填）
param(
    [string]$Tag = "",
    [ValidateSet("ucrt", "msvcrt")] [string]$Runtime = "ucrt",
    [string]$Mirror = "",
    [string]$Sha256 = "",
    [switch]$KeepWork
)
$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$Work = Join-Path $Root "ci-work"
# rfcs/0006 注册表布局：toolchains/clang/<LLVM 版本>/（版本在解包后确定）
$Toolchains = Join-Path $Root "toolchains"

function Get-LatestTag {
    $api = "https://api.github.com/repos/mstorsjo/llvm-mingw/releases/latest"
    if ($Mirror) { $api = $Mirror.TrimEnd("/") + "/" + $api }
    $json = (Invoke-WebRequest -Uri $api -UseBasicParsing -TimeoutSec 60).Content | ConvertFrom-Json
    return $json.tag_name
}

if (-not $Tag) { $Tag = Get-LatestTag }
$Asset = "llvm-mingw-$Tag-$Runtime-x86_64.zip"
$Url = "https://github.com/mstorsjo/llvm-mingw/releases/download/$Tag/$Asset"
if ($Mirror) { $Url = $Mirror.TrimEnd("/") + "/" + $Url }

New-Item -ItemType Directory -Force -Path $Work | Out-Null
$Zip = Join-Path $Work $Asset
Write-Host "== 拉取 llvm-mingw $Tag ($Runtime, windows-x64) =="
if (Test-Path $Zip) {
    Write-Host "   已有缓存：$Zip"
} else {
    Write-Host "   $Url"
    Invoke-WebRequest -Uri $Url -OutFile $Zip -TimeoutSec 3600 -UseBasicParsing
}

$want = if ($Sha256) { $Sha256 } elseif ($env:LLVM_MINGW_SHA256) { $env:LLVM_MINGW_SHA256 } else { "" }
if ($want) {
    $actual = (Get-FileHash $Zip -Algorithm SHA256).Hash.ToLower()
    if ($actual -ne $want.ToLower()) { throw "sha256 不匹配：得到 $actual，期望 $want" }
    Write-Host "   sha256 校验通过"
} else {
    Write-Warning "未提供 sha256（发行前请用 -Sha256 或 LLVM_MINGW_SHA256 固定下来）"
}

Write-Host "== 解包 =="
$Extract = Join-Path $Work "llvm-mingw"
if (Test-Path $Extract) { Remove-Item $Extract -Recurse -Force }
Expand-Archive -Path $Zip -DestinationPath $Extract -Force
$Src = Get-ChildItem $Extract -Directory | Select-Object -First 1
if (-not $Src) { throw "解包后找不到 llvm-mingw 根目录" }
$S = $Src.FullName
Write-Host "   源：$S"

# LLVM 版本决定安装目录名（toolchains/clang/<ver>/）
$llvmVer = (& (Join-Path $S "bin\clang.exe") --version | Select-String -Pattern "clang version" | Select-Object -First 1)
if (-not $llvmVer) { throw "取不到 clang 版本号（$S\bin\clang.exe --version 失败）" }
$ver = ($llvmVer.ToString() -replace '.*clang version ([0-9][0-9.]*).*', '$1')
$Dest = Join-Path $Toolchains (Join-Path "clang" $ver)

Write-Host "== 裁剪到 $Dest =="
if (Test-Path $Dest) { Remove-Item $Dest -Recurse -Force }
New-Item -ItemType Directory -Force -Path (Join-Path $Dest "bin") | Out-Null

# ---- 1) bin：clang 系 + lld + llvm 小工具（clang 同时充当「C 编译器」与「IR → .o」） ----
# 不进包：调试器(llldb)、Python 全家桶、clangd/clang-tidy/clang-format/scan-build、
#         sanitizer 动态库、C++ 运行时 dll（AILA 只编 C，不需要）
$drop = @("lldb*", "python*", "idle3*", "pydoc*", "clangd*", "clang-tidy*", "run-clang-tidy*",
    "scan-build*", "analyze-build*", "intercept-build*", "clang-format*", "clang-scan-deps*",
    "liblldb.dll", "libpython*.dll", "libffi-8.dll", "libclang_rt.asan*",
    "*.mustache", "hwasan_symbolize", "*wrapper.sh", "*.md")
# 注意：clang.exe 自身是用 libc++ 构建并动态链接的，libc++.dll / libunwind.dll /
# libwinpthread-1.dll 必须随包，否则 STATUS_DLL_NOT_FOUND（0xC0000135）。
Get-ChildItem (Join-Path $S "bin") -File | ForEach-Object {
    $skip = $false
    foreach ($pat in $drop) { if ($_.Name -like $pat) { $skip = $true; break } }
    if (-not $skip) { Copy-Item $_.FullName (Join-Path $Dest "bin") }
}

# ---- 2) clang 资源目录：内建头 + compiler-rt 的 x86_64 内建函数 ----
$resource = Join-Path $S "lib\clang"
$resVer = (Get-ChildItem $resource -Directory | Select-Object -First 1).Name
$resDest = Join-Path $Dest "lib\clang\$resVer"
New-Item -ItemType Directory -Force -Path $resDest | Out-Null
Copy-Item (Join-Path $resource "$resVer\include") (Join-Path $resDest "include") -Recurse
New-Item -ItemType Directory -Force -Path (Join-Path $resDest "lib\windows") | Out-Null
Get-ChildItem (Join-Path $resource "$resVer\lib\windows") -File -Filter "*-x86_64.a" |
    Where-Object { $_.Name -like "*builtins*" -or $_.Name -like "*profile*" } |
    ForEach-Object { Copy-Item $_.FullName (Join-Path $resDest "lib\windows") }

# ---- 3) C 头（sysroot include）：只留 C 侧，丢掉 C++ 与图形/驱动相关头 ----
$incKeep = @()
Get-ChildItem (Join-Path $S "include") | ForEach-Object {
    if ($_.PSIsContainer -and $_.Name -in @("c++", "ddk", "gdiplus", "GL", "KHR", "mach-o", "wrl")) { return }
    Copy-Item $_.FullName (Join-Path $Dest "include") -Recurse -Force
}

# ---- 4) MinGW-w64 sysroot 库（CRT / 导入库 / 运行时）：只留 x86_64 的 lib ----
$sys = Join-Path $Dest "x86_64-w64-mingw32"
New-Item -ItemType Directory -Force -Path $sys | Out-Null
Copy-Item (Join-Path $S "x86_64-w64-mingw32\lib") (Join-Path $sys "lib") -Recurse

# ---- 5) 许可证与版本（VERSION 首段必须是 LLVM 版本：aila 按它判版本带；无 BOM） ----
if (Test-Path (Join-Path $S "LICENSE.TXT")) { Copy-Item (Join-Path $S "LICENSE.TXT") $Dest }
[System.IO.File]::WriteAllText((Join-Path $Dest "VERSION"), "$ver (llvm-mingw $Tag, $Runtime, trimmed)",
    (New-Object System.Text.UTF8Encoding($false)))

# ---- 6) 把本平台条目合并进 toolchains/toolchains.json 的 available 段 ----
$registry = Join-Path $Toolchains "toolchains.json"
$url = "https://github.com/mstorsjo/llvm-mingw/releases/download/$Tag/llvm-mingw-$Tag-$Runtime-x86_64.zip"
if ($Mirror) { $url = $Mirror.TrimEnd("/") + "/" + $url }
$sizeMB = [math]::Round((Get-Item $Zip).Length / 1MB)
if (Test-Path $registry) {
    $doc = Get-Content $registry -Raw -Encoding UTF8 | ConvertFrom-Json
    $keep = @($doc.available | Where-Object { $_.family -ne "clang" -or $_.url -ne $url })
    $entry = [pscustomobject]@{ family = "clang"; version = $ver; url = $url; sha256 = $Sha256; size = "$sizeMB MB" }
    $doc.available = @($keep + @($entry))
    $doc | ConvertTo-Json -Depth 10 | Out-Null
    [System.IO.File]::WriteAllText($registry, ($doc | ConvertTo-Json -Depth 10), (New-Object System.Text.UTF8Encoding($false)))
    Write-Host "   available 条目已合并：toolchains/toolchains.json"
} else {
    Write-Warning "没有 $registry，跳过 available 合并"
}

if (-not $KeepWork) { Remove-Item $Extract -Recurse -Force -ErrorAction SilentlyContinue }

$mb = [math]::Round((Get-ChildItem $Dest -Recurse -File | Measure-Object -Property Length -Sum).Sum / 1MB)
Write-Host "== 完成：$Dest（$mb MB，LLVM $ver）=="
Write-Host ("   bin: " + (Get-ChildItem (Join-Path $Dest "bin") -File | Measure-Object).Count + " 个文件")
& (Join-Path $Dest "bin\clang.exe") --version | Select-Object -First 2
