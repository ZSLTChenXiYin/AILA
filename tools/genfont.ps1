# 生成点阵字库（AILA 源）：ASCII 半角 8x16 + 汉字全角 16x16
#
# 128x64 的 SSD1306 只有一个字节寻址的 GDDRAM，没有字体引擎，汉字必须自带点阵。
# 这里用 Windows 自带字体取位图，输出四张常量表（升序码点 + 定长点阵 + 每字宽度）：
#   CPS    : [U32; N]      升序码点
#   BITS   : [U8; N*32]    每字 32 字节 = 16 行 x 2 字节（行内 MSB 在左）
#   WIDTHS : [U8; N]       1 = 半角（8px 宽，只用每行第 0 字节）/ 2 = 全角（16px 宽）
#
# 为什么定长 32 字节而不是按宽度压缩：绘制侧（std.paint.blit_glyph）因此零查表、
# 零偏移累加，一个字形的记录地址 = index*32。半角字的右 8 列留 0，
# 代价是 ~95 个 ASCII 字共 3KB 而非 1.5KB（flash 里，不占运行期 RAM）。
#
# 生成物放子包目录 font/（包根只允许 main.aila，specs/0001 §1）；
# 壳层只有 const/type/enum/struct/无体 func，且 const 初始化器必须是编译期标量，
# 所以两张表留在 module 块里（private），对外开**访问器**：
#   - 标量访问器（font_cp / font_byte / font_width）：老代码直接用；
#   - 视图访问器（font_cps / font_bits / font_widths）：返回 @static 视图，
#     用来装配 std.paint.GlyphSet（designs/0006 §4.1 第 1 条的合法形态）。
#
# 用法： pwsh -File tools/genfont.ps1 [-Extra "想加的字"] [-Out 目标.aila]
param(
    [string]$Extra = "",
    [string]$Out = "",
    [string]$FontName = "SimSun",
    [string]$AsciiFont = "Consolas",
    [int]$Size = 16,
    [int]$AsciiSize = 13,
    [string]$Cjk = "",
    [switch]$NoHalf
)
$ErrorActionPreference = "Stop"
Add-Type -AssemblyName System.Drawing

$Root = Split-Path -Parent $PSScriptRoot
if (-not $Out) { $Out = Join-Path $Root "examples\display_demo\font\font16.aila" }

$ascii = -join (0x20..0x7E | ForEach-Object { [char]$_ })
if (-not $Cjk) {
    $Cjk = "你好世界温度摄氏度电压电流时间秒分小时网络已连接断开等待收到显示清屏服务端中文测试成功失败错误重试正在初始化信号强度名字密码输入输出第一二三四五六七八九十零年月日上下左右前后大小开关启动停止运行空闲状态内存剩余重启发送接收数据包端口地址路由器配置保存读取写入删除修改确认取消是的不在有和人中为个到说国地也子时道出而要于就下得可年生自会那后能对着事其里所去行过家十用发天如然作方成者多日都三小军二无同么经法当起与好看学进种将还分此心前面又定见只主没公从"
}
$chars = [System.Collections.Generic.SortedSet[int]]::new()
foreach ($c in ($ascii + $Cjk + $Extra).ToCharArray()) {
    if (-not $chars.Contains([int]$c)) { $null = $chars.Add([int]$c) }
}

# 一个 16x16 的取字面：半角字画在左 8 列，全角字画满 16 列
$bmp = New-Object System.Drawing.Bitmap 16, 16
$g = [System.Drawing.Graphics]::FromImage($bmp)
$g.TextRenderingHint = [System.Drawing.Text.TextRenderingHint]::SingleBitPerPixelGridFit
$fontWide = New-Object System.Drawing.Font($FontName, $Size, [System.Drawing.FontStyle]::Regular, [System.Drawing.GraphicsUnit]::Pixel)
$fontHalf = New-Object System.Drawing.Font($AsciiFont, $AsciiSize, [System.Drawing.FontStyle]::Regular, [System.Drawing.GraphicsUnit]::Pixel)
$brush = [System.Drawing.Brushes]::White

$cps = @()
$widths = @()
$bits = New-Object System.Collections.Generic.List[byte]
foreach ($cp in $chars) {
    $half = ($cp -lt 0x2E80 -and -not $NoHalf)
    $g.Clear([System.Drawing.Color]::Black)
    $fmt = New-Object System.Drawing.StringFormat
    $rectW = 16
    if ($half) { $rectW = 8 }
    $fmt.Alignment = [System.Drawing.StringAlignment]::Center
    $fmt.LineAlignment = [System.Drawing.StringAlignment]::Center
    $font = $fontWide
    if ($half) { $font = $fontHalf }
    $g.DrawString([string][char]$cp, $font, $brush, (New-Object System.Drawing.RectangleF 0, 0, $rectW, 16), $fmt)
    $cps += $cp
    if ($half) { $widths += 1 } else { $widths += 2 }
    for ($y = 0; $y -lt 16; $y++) {
        $hi = 0; $lo = 0
        for ($x = 0; $x -lt 8; $x++) {
            if ($bmp.GetPixel($x, $y).R -gt 128) { $hi = $hi -bor (0x80 -shr $x) }
        }
        if (-not $half) {
            for ($x = 8; $x -lt 16; $x++) {
                if ($bmp.GetPixel($x, $y).R -gt 128) { $lo = $lo -bor (0x80 -shr ($x - 8)) }
            }
        }
        $bits.Add([byte]$hi); $bits.Add([byte]$lo)
    }
}
$g.Dispose(); $bmp.Dispose()

$nHalf = 0
foreach ($w in $widths) { if ($w -eq 1) { $nHalf++ } }

$sb = New-Object System.Text.StringBuilder
[void]$sb.AppendLine("// 16x16 / 8x16 点阵字库（自动生成，勿手改）：pwsh -File tools/genfont.ps1")
[void]$sb.AppendLine("// 字源：$FontName ${Size}px（全角）+ $AsciiFont ${AsciiSize}px（半角）；$($cps.Count) 个字（半角 $nHalf / 全角 $($cps.Count - $nHalf)）")
[void]$sb.AppendLine("// 结构：CPS 升序码点 -> 同下标处的 BITS 定长记录（i*32 .. i*32+31，16 行 x 2 字节，行内 MSB 在左）")
[void]$sb.AppendLine("//       WIDTHS[i] = 1（半角，只用每行第 0 字节）/ 2（全角）；缺字由 std.paint 画实心方块")
[void]$sb.AppendLine("//")
[void]$sb.AppendLine("// 壳层只有无体 func：两张表在 module 块里（private），视图访问器返回 @static 视图，")
[void]$sb.AppendLine("// 供装配 std.paint.GlyphSet（designs/0006 §2.4）。")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("public func font_count() -> USize;")
[void]$sb.AppendLine("public func font_cp(i: USize) -> U32;")
[void]$sb.AppendLine("public func font_byte(i: USize) -> U8;")
[void]$sb.AppendLine("public func font_width(i: USize) -> U8;")
[void]$sb.AppendLine("public func font_cps() -> [U32];")
[void]$sb.AppendLine("public func font_bits() -> [U8];")
[void]$sb.AppendLine("public func font_widths() -> [U8];")
[void]$sb.AppendLine("public func font_px_h() -> U16;")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("module {")
[void]$sb.AppendLine("let CPS: [U32; $($cps.Count)] = [")
$line = "    "
foreach ($v in $cps) {
    $item = ("0x{0:X}" -f $v) + ","
    if ($line.Length + $item.Length -gt 100) { [void]$sb.AppendLine($line); $line = "    " }
    $line += $item
}
if ($line.Trim()) { [void]$sb.AppendLine($line) }
[void]$sb.AppendLine("];")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("let BITS: [U8; $($cps.Count * 32)] = [")
$line = "    "
foreach ($v in $bits) {
    $item = ("0x{0:X2}" -f $v) + ","
    if ($line.Length + $item.Length -gt 100) { [void]$sb.AppendLine($line); $line = "    " }
    $line += $item
}
if ($line.Trim()) { [void]$sb.AppendLine($line) }
[void]$sb.AppendLine("];")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("let WIDTHS: [U8; $($cps.Count)] = [")
$line = "    "
foreach ($v in $widths) {
    $item = "$v,"
    if ($line.Length + $item.Length -gt 100) { [void]$sb.AppendLine($line); $line = "    " }
    $line += $item
}
if ($line.Trim()) { [void]$sb.AppendLine($line) }
[void]$sb.AppendLine("];")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("    public func font_count() -> USize { return $($cps.Count); }")
[void]$sb.AppendLine("    public func font_cp(i: USize) -> U32 { return CPS[i]; }")
[void]$sb.AppendLine("    public func font_byte(i: USize) -> U8 { return BITS[i]; }")
[void]$sb.AppendLine("    public func font_width(i: USize) -> U8 { return WIDTHS[i]; }")
[void]$sb.AppendLine("    public func font_cps() -> [U32] {")
[void]$sb.AppendLine("        view v: [U32] = CPS;")
[void]$sb.AppendLine("        return v;")
[void]$sb.AppendLine("    }")
[void]$sb.AppendLine("    public func font_bits() -> [U8] {")
[void]$sb.AppendLine("        view v: [U8] = BITS;")
[void]$sb.AppendLine("        return v;")
[void]$sb.AppendLine("    }")
[void]$sb.AppendLine("    public func font_widths() -> [U8] {")
[void]$sb.AppendLine("        view v: [U8] = WIDTHS;")
[void]$sb.AppendLine("        return v;")
[void]$sb.AppendLine("    }")
[void]$sb.AppendLine("    public func font_px_h() -> U16 { return 16; }")
[void]$sb.AppendLine("}")

New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Out) | Out-Null
[System.IO.File]::WriteAllText($Out, ($sb.ToString().TrimEnd() + "`n"), [System.Text.UTF8Encoding]::new($false))
Write-Host "已生成 $Out（$($cps.Count) 字：半角 $nHalf / 全角 $($cps.Count - $nHalf)；$($bits.Count) 字节点阵）"
