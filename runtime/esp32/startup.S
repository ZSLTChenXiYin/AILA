/* AILA ESP32 启动代码（windowed ABI——官方 GCC 8.4 工具链构建期固定 windowed，ESP-IDF 同款）。
 *
 * 纪律：字面量池必须在代码之前——l32r 只能引用更低地址的字面量。
 * 栈顶（_stack_top）由生成的链接脚本提供：DRAM 顶，向下生长。
 * window overflow 向量沿用 ROM（VECBASE 保持 ROM 默认，ROM 自带溢出处理）。
 */
    .section .text
    .global call_start

.Lstack_top:
    .word _stack_top

call_start:
    entry   a1, 16              /* windowed 帧分配（谁 call 进来都安全） */
    l32r    a1, .Lstack_top     /* a1 = 栈顶 */
    call8   aila_rt_main        /* 进平台 runtime：初始化 UART，跑 AILA main */

.Lhang:
    j       .Lhang              /* main 返回即停机（裸机无处可退） */
