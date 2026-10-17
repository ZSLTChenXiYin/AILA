; AILA module: std.io
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"


declare void @aila_print_i32(i32)
declare void @aila_print_u32(i32)
declare void @aila_print_f64(double)
declare void @aila_print_bool(i8)
declare void @aila_print_char(i32)
declare void @aila_print_nl()
declare void @aila_print_bytes(ptr, i64)

define void @aila_std_io__print_i32(i32 %p_v) {
  %slot_v_1 = alloca i32, align 8
  store i32 %p_v, ptr %slot_v_1
  %t1 = load i32, ptr %slot_v_1
  call void @aila_print_i32(i32 %t1)
  ret void
}

define void @aila_std_io__print_u32(i32 %p_v) {
  %slot_v_1 = alloca i32, align 8
  store i32 %p_v, ptr %slot_v_1
  %t1 = load i32, ptr %slot_v_1
  call void @aila_print_u32(i32 %t1)
  ret void
}

define void @aila_std_io__print_usize(i64 %p_v) {
  %slot_v_1 = alloca i64, align 8
  store i64 %p_v, ptr %slot_v_1
  %t1 = load i64, ptr %slot_v_1
  %t2 = trunc i64 %t1 to i32
  call void @aila_print_u32(i32 %t2)
  ret void
}

define void @aila_std_io__print_f64(double %p_v) {
  %slot_v_1 = alloca double, align 8
  store double %p_v, ptr %slot_v_1
  %t1 = load double, ptr %slot_v_1
  call void @aila_print_f64(double %t1)
  ret void
}

define void @aila_std_io__print_f32(float %p_v) {
  %slot_v_1 = alloca float, align 8
  store float %p_v, ptr %slot_v_1
  %t1 = load float, ptr %slot_v_1
  %t2 = fpext float %t1 to double
  call void @aila_print_f64(double %t2)
  ret void
}

define void @aila_std_io__print_bool(i1 %p_v) {
  %slot_v_1 = alloca i1, align 8
  store i1 %p_v, ptr %slot_v_1
  %t1 = load i1, ptr %slot_v_1
  %t2 = zext i1 %t1 to i8
  call void @aila_print_bool(i8 %t2)
  ret void
}

define void @aila_std_io__print_char(i32 %p_c) {
  %slot_c_1 = alloca i32, align 8
  store i32 %p_c, ptr %slot_c_1
  %t1 = load i32, ptr %slot_c_1
  call void @aila_print_char(i32 %t1)
  ret void
}

define void @aila_std_io__print_nl() {
  call void @aila_print_nl()
  ret void
}

define void @aila_std_io__print_text({ ptr, i64 } %p_t) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_bytes_2 = alloca { ptr, i64 }, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  %t1 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t1, ptr %slot_bytes_2
  %t2 = load { ptr, i64 }, ptr %slot_bytes_2
  %t3 = extractvalue { ptr, i64 } %t2, 0
  %t4 = load { ptr, i64 }, ptr %slot_bytes_2
  %t5 = extractvalue { ptr, i64 } %t4, 1
  call void @aila_print_bytes(ptr %t3, i64 %t5)
  ret void
}

define void @aila_std_io__print_bytes({ ptr, i64 } %p_b) {
  %slot_b_1 = alloca { ptr, i64 }, align 8
  store { ptr, i64 } %p_b, ptr %slot_b_1
  %t1 = load { ptr, i64 }, ptr %slot_b_1
  %t2 = extractvalue { ptr, i64 } %t1, 0
  %t3 = load { ptr, i64 }, ptr %slot_b_1
  %t4 = extractvalue { ptr, i64 } %t3, 1
  call void @aila_print_bytes(ptr %t2, i64 %t4)
  ret void
}

define void @aila_std_io__print_i32_nl(i32 %p_v) {
  %slot_v_1 = alloca i32, align 8
  store i32 %p_v, ptr %slot_v_1
  %t1 = load i32, ptr %slot_v_1
  call void @aila_print_i32(i32 %t1)
  call void @aila_print_nl()
  ret void
}
