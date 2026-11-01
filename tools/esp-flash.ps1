# ESP8266 烧写脚本 —— 三种已验证布局（详见 docs/designs/0005-esp8266-boot-and-flash.md）
#
# 用法：
#   pwsh -File tools/esp-flash.ps1 -Mode oled          # AILA freestanding OLED 演示（烧 0x0）
#   pwsh -File tools/esp-flash.ps1 -Mode at            # 官方 AT 固件（Wi-Fi 可用，烧 boot+user1+PHY）
#   pwsh -File tools/esp-flash.ps1 -Mode sdk           # AILA SDK 应用（v2 镜像，烧 boot+user1+user2）
#
# 关键规则（实测）：flash map 由 bootloader 镜像头的 size nibble 决定，应用编译的 map 必须与之一致：
#   AT(map2/1MB) → boot byte3 = 0x20 ；AILA SDK(map4/4MB) → boot byte3 = 0x40
param(
    [ValidateSet("oled", "at", "sdk")]
    [string]$Mode = "oled",
    [string]$Port = "COM3",
    [int]$Baud = 115200
)
$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$sdkBin = Join-Path $root ".local/toolchains/sdk/esp8266-nosdk/3.0.5/bin"
$esptool = "esptool"
$tmp = [System.IO.Path]::GetTempPath()

function Invoke-Esptool([string[]]$esptoolArgs) {
    & $esptool --chip esp8266 --port $Port --baud $Baud @esptoolArgs
    if ($LASTEXITCODE -ne 0) { throw "esptool 失败：$($esptoolArgs -join ' ')" }
}

function New-PatchedBoot([string]$source, [byte]$sizeNibble, [string]$outPath) {
    $b = [System.IO.File]::ReadAllBytes($source)
    $b[3] = [byte](($sizeNibble -shl 4) -bor ($b[3] -band 0x0F))  # 只改 size 高 4 位，保留 freq
    [System.IO.File]::WriteAllBytes($outPath, $b)
    return $outPath
}

switch ($Mode) {
    "oled" {
        $app = Join-Path $root ".local/build/examples/oled_smoke/xtensa_esp8266_elf/app.bin"
        if (-not (Test-Path $app)) { throw "先构建：aila build examples/oled_smoke" }
        Write-Host "[oled] freestanding 镜像烧 0x0（ROM 直启，image v1）"
        Invoke-Esptool @("--before", "default-reset", "--after", "hard-reset",
                         "write-flash", "-fm", "keep", "-fs", "keep", "-ff", "keep", "0x0", $app)
    }
    "at" {
        $boot = New-PatchedBoot (Join-Path $sdkBin "boot_v1.7.bin") 2 (Join-Path $tmp "boot_1m.bin")
        Write-Host "[at] boot 头 → map 2(1MB, byte3=0x20)；AT 固件烧 0x1000；PHY 烧 0x7C000"
        # RF_CAL 必须是擦除态，写全 0 会 rf_cal[0] !=0x05 反复重启
        Invoke-Esptool @("erase-region", "0x7B000", "0x5000")
        Invoke-Esptool @("--before", "default-reset", "--after", "hard-reset",
                         "write-flash", "-fm", "keep", "-fs", "keep", "-ff", "keep",
                         "0x0", $boot,
                         "0x1000", (Join-Path $sdkBin "at/512+512/user1.1024.new.2.bin"),
                         "0x7C000", (Join-Path $sdkBin "esp_init_data_default_v08.bin"))
        Write-Host "[at] 完成。AT 控制台 115200：AT / AT+GMR / AT+CWJAP=..."
    }
    "sdk" {
        $app = Join-Path $root ".local/build/examples/esp8266_oled/xtensa_esp8266_elf/app.bin"
        if (-not (Test-Path $app)) { throw "先构建：aila build examples/esp8266_oled" }
        $boot = New-PatchedBoot (Join-Path $sdkBin "boot_v1.7.bin") 4 (Join-Path $tmp "boot_4m.bin")
        # 构建产物本身就是 FOTA v2（0xEA）镜像：irom 紧跟 16B 镜像头，链接地址
        # 0x40201010 与 user1 槽 flash 0x1000 对齐 —— 不再需要外挂 esptool elf2image。
        Write-Host "[sdk] boot 头 → map 4(4MB, byte3=0x40)；v2 镜像烧 0x1000 与 user2 0x81000"
        Invoke-Esptool @("erase-region", "0xFB000", "0x5000")
        Invoke-Esptool @("--before", "default-reset", "--after", "hard-reset",
                         "write-flash", "-fm", "keep", "-fs", "keep", "-ff", "keep",
                         "0x0", $boot, "0x1000", $app, "0x81000", $app,
                         "0xFC000", (Join-Path $sdkBin "esp_init_data_default_v08.bin"))
        Write-Host "[sdk] 完成（串口 115200；ROM/bootloader 自身日志在 74880）"
    }
}
