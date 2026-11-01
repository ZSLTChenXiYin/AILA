# ESP8266 串口小工具：给板子发 AT 指令、看启动日志、做一次 TCP 回环
#
# 用途：在还没有能力烧自己的固件前，先用板上的 AT 固件验证
#   「串口通 → Wi-Fi 能连 → TCP 能收发中文」这条链路是不是好的。
#
# 用法：
#   pwsh -File tools/esp-at.ps1 -Boot                       # 复位并抓 74880 启动日志
#   pwsh -File tools/esp-at.ps1 -Cmd "AT","AT+GMR"          # 发任意 AT 指令（默认 COM3@115200）
#   pwsh -File tools/esp-at.ps1 -Scan                       # 扫附近 AP（找 SSID）
#   pwsh -File tools/esp-at.ps1 -Join -SSID "ChenXiYin" -Pass "84170919"   # 连 Wi-Fi 并打印 IP
#   pwsh -File tools/esp-at.ps1 -Port COM4 -Baud 74880 -Boot
#
# 说明：AT 固件的启动日志固定在 74880（晶振分频），指令通道是 115200。
param(
    [string]$Port = "COM3",
    [int]$Baud = 115200,
    [string[]]$Cmd = @(),
    [int]$WaitMs = 1500,
    [switch]$Boot,
    [switch]$Scan,
    [switch]$Join,
    [string]$SSID = "",
    [string]$Pass = ""
)
$ErrorActionPreference = "Stop"

function Open-Port([string]$name, [int]$baud) {
    $sp = New-Object System.IO.Ports.SerialPort $name, $baud, 'None', 8, 'One'
    $sp.ReadTimeout = 500
    $sp.DtrEnable = $false
    $sp.RtsEnable = $false
    $sp.Open()
    return $sp
}

function Read-Chunk($sp, [int]$ms) {
    $sb = New-Object System.Text.StringBuilder
    $end = (Get-Date).AddMilliseconds($ms)
    while ((Get-Date) -lt $end) {
        try {
            if ($sp.BytesToRead -gt 0) {
                $buf = New-Object byte[] $sp.BytesToRead
                $n = $sp.Read($buf, 0, $buf.Length)
                for ($i = 0; $i -lt $n; $i++) { [void]$sb.Append([char]$buf[$i]) }
            } else { Start-Sleep -Milliseconds 15 }
        } catch { break }
    }
    return $sb.ToString()
}

# 不可见字节点成 '.'，但保留中文（UTF-8 三字节要按字节拼回去再解）
function Show-Text([string]$s) {
    $bytes = [System.Text.Encoding]::GetEncoding(28591).GetBytes($s)   # Latin-1：一字符一字节还原
    $utf8 = [System.Text.Encoding]::UTF8.GetString($bytes)
    if ($utf8 -match '[一-龥]') { return $utf8 }
    return ($s -replace "[^\x20-\x7E\r\n]", '.')
}

function Send-AT($sp, [string]$cmd, [int]$wait) {
    $sp.Write("$cmd`r`n")
    $r = Read-Chunk $sp $wait
    Write-Host "--- $cmd ---"
    if ($r.Length -eq 0) { Write-Host "（无回应）" } else { Write-Host (Show-Text $r) }
    return $r
}

if ($Boot) {
    $spb = Open-Port $Port 74880
    Start-Sleep -Milliseconds 100
    $spb.RtsEnable = $true; Start-Sleep -Milliseconds 120; $spb.RtsEnable = $false
    $log = Read-Chunk $spb 1800
    $spb.Close()
    Write-Host "=== 起机日志（74880）==="
    Write-Host (Show-Text $log)
    if (-not ($Scan -or $Join -or $Cmd.Count)) { return }
    Start-Sleep -Milliseconds 400
}

if ($Scan -or $Join -or $Cmd.Count) {
    $sp = Open-Port $Port $Baud
    Write-Host "已打开 $Port @$Baud"
    [void](Read-Chunk $sp 300)
    if ($Scan) { [void](Send-AT $sp "AT+CWLAP" 12000) }
    if ($Join) {
        if (-not $SSID) { throw "连 Wi-Fi 需要 -SSID" }
        [void](Send-AT $sp "AT+CWMODE=1" 1500)
        [void](Send-AT $sp ("AT+CWJAP=`"$SSID`",`"$Pass`"") 18000)
        [void](Send-AT $sp "AT+CIPSTA?" 2500)
    }
    foreach ($c in $Cmd) { [void](Send-AT $sp $c $WaitMs) }
    $sp.Close()
    Write-Host "端口已关闭"
}
