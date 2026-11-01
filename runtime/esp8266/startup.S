/* AILA ESP8266 启动代码（call0 ABI）。
 *
 * 纪律：字面量池必须在代码之前——l32r 只能引用更低地址的字面量。
 * 栈顶（_stack_top）由链接脚本 esp8266.ld 提供：DRAM 顶，向下生长。
 */
    .section .text
    .global call_start

.Lstack_top:
    .word _stack_top

call_start:
    l32r    a1, .Lstack_top     /* a1 = 栈顶 */
    call0   aila_rt_main        /* 进平台 runtime：初始化 UART，跑 AILA main */

.Lhang:
    j       .Lhang              /* main 返回即停机（裸机无处可退） */
