// 裸机延时：SSD1306 软件 I2C 需要微秒级半周期。bare-metal 路径不链 SDK，
// 所以自己实现 ets_delay_us（52MHz 晶振下忙等，1µs ≈ 52 周期）。
// 精度不要求高：SSD1306 上限 400kHz，我们目标 ~250kHz。
void ets_delay_us(unsigned int us) {
    while (us--) {
        volatile unsigned int i = 240; /* 约 4µs（52MHz，-O1）→ SCL ≈ 125kHz，SSD1306 稳 */
        while (i--) {
        }
    }
}
