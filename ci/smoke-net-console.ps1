# 端到端验证：net_console 服务端 ← TCP ← 假 ESP8266 客户端
# 用法：pwsh -File ci/smoke-net-console.ps1
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
$env:AILA_HOME = $Root
$aila = Join-Path $Root ".local\bin\aila.exe"
$proj = Join-Path $Root "examples\net_console"
$out = Join-Path $env:TEMP "net_console_smoke"

# 1) 构建
Push-Location $proj
& $aila build
if ($LASTEXITCODE -ne 0) { Pop-Location; throw "构建失败" }
Pop-Location
$exe = Join-Path $Root ".local\build\examples\net_console\x86_64_w64_windows_gnu\app.exe"
if (-not (Test-Path $exe)) { throw "没有产物：$exe" }

# 2) 准备服务端控制台输入（UTF-8 中文 + 空行退出）
New-Item -ItemType Directory -Force -Path $out | Out-Null
$stdin = Join-Path $out "stdin.txt"
$stdout = Join-Path $out "stdout.txt"
$lines = @("你好，世界", "ESP8266 收到中文了", "CLEAR", "第二屏：温度 25.6C")
[System.IO.File]::WriteAllText($stdin, (($lines -join "`n") + "`n`n"), [System.Text.UTF8Encoding]::new($false))

# 3) 起服务端：控制台输入重定向到文件
$p = Start-Process -FilePath $exe -RedirectStandardInput $stdin -RedirectStandardOutput $stdout -PassThru -WindowStyle Hidden

# 4) 当假 ESP8266：连上去收字节
Start-Sleep -Milliseconds 600
$client = [System.Net.Sockets.TcpClient]::new()
$connected = $false
for ($i = 0; $i -lt 40; $i++) {
    try { $client.Connect("127.0.0.1", 8266); $connected = $true; break } catch { Start-Sleep -Milliseconds 150 }
}
if (-not $connected) { $p.Kill(); throw "连不上服务端（8266）" }

$stream = $client.GetStream()
$stream.ReadTimeout = 4000
$buf = New-Object byte[] 4096
$got = New-Object System.Collections.Generic.List[byte]
$deadline = (Get-Date).AddSeconds(6)
try {
    while ((Get-Date) -lt $deadline) {
        $n = $stream.Read($buf, 0, $buf.Length)
        if ($n -le 0) { break }
        for ($k = 0; $k -lt $n; $k++) { $got.Add($buf[$k]) }
        if ($got.Count -gt 0 -and $got[$got.Count - 1] -eq 10 -and $got.Count -ge 60) { break }
    }
} catch { }
$client.Close()

$text = [System.Text.Encoding]::UTF8.GetString($got.ToArray())
$p.WaitForExit(3000) | Out-Null
if (-not $p.HasExited) { $p.Kill() }

Write-Host "===== 客户端（假 ESP8266）收到 ====="
Write-Host $text
Write-Host "===== 服务端控制台输出 ====="
Get-Content $stdout -Encoding utf8

# 5) 断言
$expect = @("你好，世界", "ESP8266 收到中文了", "CLEAR", "第二屏：温度 25.6C")
foreach ($e in $expect) {
    if (-not $text.Contains($e)) { throw "客户端没收到「$e」" }
}
if ([regex]::Matches($text, "`n").Count -lt 4) { throw "行数不对：$text" }
Write-Host "✔ 端到端通过：4 行中文按 '一行一消息' 送达"
