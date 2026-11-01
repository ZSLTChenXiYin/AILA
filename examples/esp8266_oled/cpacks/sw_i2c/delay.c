// 裸机延时的垫片：软件 I2C 需要微秒级半周期，而 freestanding 路径**不链 SDK**
// （ets_delay_us 在 SDK 构建里由 ROM 符号表提供，裸机没有）。
//
// 这里用忙等自给自足，并沿用上一轮上板验证过的循环量：内层 240 次约 4µs（52MHz、-O1），
// 于是实际 SCL 约 125kHz —— 比 SSD1306 的 400kHz 上限保守，但实测稳定（designs/0005）。
// 想要更快就把 I2C_DELAY_US 与这里的循环量一起调小，并重新上板确认 NACK 计数为 0。
//
// 注意：这份垫片只服务「C 侧位时序」。AILA 侧要微秒延时请用 std.time.sleep_us
// （三家族走 ccount，换算基准 = chip.json 的 cpu_hz）。
void ets_delay_us(unsigned int us) {
    while (us--) {
        volatile unsigned int i = 240;
        while (i--) {
        }
    }
}
