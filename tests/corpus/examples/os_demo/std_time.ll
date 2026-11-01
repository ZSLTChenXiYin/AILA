; AILA module: std.time
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"


declare i64 @aila_rt_time_ms()
declare void @aila_rt_sleep_ms(i64)

define i64 @aila_std_time__mono_ms() {
  %t1 = call i64 @aila_rt_time_ms()
  ret i64 %t1
}

define void @aila_std_time__sleep_ms(i64 %p_n) {
  %slot_n_1 = alloca i64, align 8
  store i64 %p_n, ptr %slot_n_1
  %t1 = load i64, ptr %slot_n_1
  call void @aila_rt_sleep_ms(i64 %t1)
  ret void
}
