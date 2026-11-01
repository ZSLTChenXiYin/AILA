# 硬件回环：AILA 服务端（PC）→ Wi-Fi → ESP8266（板上 AT 固件）收中文
#
# 为什么要有这个：自己的固件还编不出来（缺 xtensa 后端），但「控制台中文 → TCP →
# 板子收到原文」这条链路可以用板上自带的 AT 固件先验通。OLED 的显示部分验不了
# （那要真固件 + 接屏），这里只验到「板子手上拿到了正确的字节」。
#
# 用法：
#   pwsh -File tools/esp-at-roundtrip.ps1                       # COM3，默认端口 8266
#   pwsh -File tools/esp-at-roundtrip.ps1 -Port COM4 -WaitSec 10
param(
    [string]$Port = "COM3",
    [int]$Baud = 115200,
    [int]$ServerPort = 8266,
    [string]$ServerExe = "",
    [int]$WaitSec = 8
)
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
$env:AILA_HOME = $Root

if (-not $ServerExe) {
    $ServerExe = Join-Path $Root ".local\build\examples\net_console\x86_64_w64_windows_gnu\app.exe"
}
if (-not (Test-Path $ServerExe)) {
    throw "找不到服务端产物：$ServerExe（先 aila build examples/net_console）"
}

# ---- 1) 串口先开：问 ESP 自己在哪个网段，再挑同网段的 PC 地址 ----
$sp = New-Object System.IO.Ports.SerialPort $Port, $Baud, 'None', 8, 'One'
$sp.ReadTimeout = 500
$sp.DtrEnable = $false
$sp.RtsEnable = $false
$sp.Open()
function Read-Chunk($s, [int]$ms) {
    $sb = New-Object System.Text.StringBuilder
    $end = (Get-Date).AddMilliseconds($ms)
    while ((Get-Date) -lt $end) {
        if ($s.BytesToRead -gt 0) {
            $buf = New-Object byte[] $s.BytesToRead
            $n = $s.Read($buf, 0, $buf.Length)
            for ($i = 0; $i -lt $n; $i++) { [void]$sb.Append([char]$buf[$i]) }   # Latin-1 一字符一字节
        } else { Start-Sleep -Milliseconds 15 }
    }
    return $sb.ToString()
}
[void](Read-Chunk $sp 300)
$sp.Write("AT+CIPSTA?`r`n")
$sta = Read-Chunk $sp 1200
$espIp = ([regex]::Match($sta, 'ip:"([0-9.]+)"')).Groups[1].Value
$espGw = ([regex]::Match($sta, 'gateway:"([0-9.]+)"')).Groups[1].Value
if (-not $espIp) { throw "ESP 没拿到 IP（先 `tools/esp-at.ps1 -Join -SSID ... -Pass ...` 连上 Wi-Fi）" }
$prefix = ($espGw -split '\.')[0..2] -join '.'
Write-Host "ESP IP：$espIp（网关 $espGw，网段 $prefix.x）"

# PC 上找同网段的地址（避开 WSL/Hyper-V 那些 172.* 虚拟网卡）
$ip = (Get-NetIPAddress -AddressFamily IPv4 |
    Where-Object { $_.IPAddress -like "$prefix.*" } |
    Select-Object -First 1).IPAddress
if (-not $ip) { throw "PC 上没有 $prefix.x 网段的地址（确认 PC 连的是同一个 Wi-Fi）" }
Write-Host "PC 局域网 IP：$ip；服务端端口：$ServerPort"

# ---- 2) 起服务端：控制台输入用文件喂（中文三行 + 空行退出）----
$work = Join-Path $env:TEMP "esp_roundtrip"
New-Item -ItemType Directory -Force -Path $work | Out-Null
$stdin = Join-Path $work "stdin.txt"
$stdout = Join-Path $work "stdout.txt"
$lines = @("你好，世界", "ESP 收到中文了", "温度 25.6C")
[System.IO.File]::WriteAllText($stdin, (($lines -join "`n") + "`n`n"), [System.Text.UTF8Encoding]::new($false))
$srv = Start-Process -FilePath $ServerExe -RedirectStandardInput $stdin -RedirectStandardOutput $stdout -PassThru -WindowStyle Hidden
Start-Sleep -Milliseconds 800
Write-Host "服务端已启动（pid=$($srv.Id)），等 ESP 连上来…"

# ---- 3) ESP 侧：AT 指令连 TCP，抓 +IPD 里的字节（串口已在上一步打开）----
[void](Read-Chunk $sp 300)
$sp.Write("AT+CIPMUX=0`r`n"); [void](Read-Chunk $sp 800)
$sp.Write("AT+CIPSTART=`"TCP`",`"$ip`",$ServerPort`r`n")
Write-Host "已发 CIPSTART，等数据…"
$raw = Read-Chunk $sp ($WaitSec * 1000)
$sp.Close()

# ---- 4) 收尾 ----
$srv.WaitForExit(4000) | Out-Null
if (-not $srv.HasExited) { $srv.Kill() }

Write-Host "===== 服务端控制台 ====="
Get-Content $stdout -Encoding utf8 | ForEach-Object { Write-Host $_ }
Write-Host "===== ESP 串口原始回显（不可见字节点成 .）====="
Write-Host ($raw -replace "[^\x20-\x7E\r\n]", '.')

# ---- 5) 从 +IPD,<len>:<payload> 里把 Payload 抠出来按 UTF-8 解 ----
$bytes = [System.Text.Encoding]::GetEncoding(28591).GetBytes($raw)
$got = New-Object System.Collections.Generic.List[string]
$i = 0
while ($i -lt $bytes.Length) {
    # 找 "+IPD,"
    $hit = -1
    for ($k = $i; $k + 4 -lt $bytes.Length; $k++) {
        if ($bytes[$k] -eq 43 -and $bytes[$k+1] -eq 73 -and $bytes[$k+2] -eq 80 -and $bytes[$k+3] -eq 68 -and $bytes[$k+4] -eq 44) { $hit = $k; break }
    }
    if ($hit -lt 0) { break }
    $p = $hit + 5
    $lenStr = ""
    while ($p -lt $bytes.Length -and $bytes[$p] -ge 48 -and $bytes[$p] -le 57) { $lenStr += [char]$bytes[$p]; $p++ }
    if ($p -lt $bytes.Length -and $bytes[$p] -eq 58) { $p++ }   # ':'
    $len = [int]$lenStr
    if ($len -le 0 -or $p + $len -gt $bytes.Length) { break }
    $payload = $bytes[$p..($p + $len - 1)]
    $got.Add([System.Text.Encoding]::UTF8.GetString($payload))
    $i = $p + $len
}

Write-Host "===== ESP 收到的消息（UTF-8 解码）====="
foreach ($g in $got) { Write-Host ("  «" + $g.TrimEnd("`n") + "»") }

if ($got.Count -eq 0) { throw "ESP 一个 +IPD 都没收到（检查：是否连的同一个 Wi-Fi、PC 防火墙是否放行 $ServerPort）" }
$missing = @()
foreach ($l in $lines) {
    if (-not ($got | Where-Object { $_.Contains($l) })) { $missing += $l }
}
if ($missing.Count -gt 0) { throw ("ESP 没收到这些行：" + ($missing -join "、")) }
Write-Host "✔ 回环通过：$($lines.Count) 行中文经 Wi-Fi + TCP 原样到达 ESP8266"
