# 串口监听：把 COM 口的原始字节按行打出来（ESP 固件 115200 8N1）
#
# 用途：烧完固件后看 AILA 的 UART 诊断。超时或凑够行数就退出。
#
# 用法： pwsh -File tools/esp-monitor.ps1 [-Port COM3] [-Baud 115200] [-Seconds 8] [-Lines 20]
param(
    [string]$Port = "COM3",
    [int]$Baud = 115200,
    [int]$Seconds = 8,
    [int]$Lines = 0
)
$ErrorActionPreference = "Stop"

$sp = New-Object System.IO.Ports.SerialPort $Port, $Baud, 'None', 8, 'One'
$sp.ReadTimeout = 500
$sp.Open()
$deadline = (Get-Date).AddSeconds($Seconds)
$buf = New-Object System.Text.StringBuilder
$count = 0
try {
    while ((Get-Date) -lt $deadline) {
        try {
            $chunk = $sp.ReadExisting()
        } catch {
            $chunk = ""
        }
        if ($chunk) {
            [void]$buf.Append($chunk)
            $text = $buf.ToString()
            while ($text.Contains("`n")) {
                $i = $text.IndexOf("`n")
                $line = $text.Substring(0, $i).TrimEnd("`r")
                $text = $text.Substring($i + 1)
                Write-Output $line
                $count++
            }
            [void]$buf.Clear()
            [void]$buf.Append($text)
            if ($Lines -gt 0 -and $count -ge $Lines) { break }
        } else {
            Start-Sleep -Milliseconds 50
        }
    }
} finally {
    $sp.Close()
    $sp.Dispose()
}
if ($buf.Length -gt 0) { Write-Output ("[tail] " + $buf.ToString()) }
