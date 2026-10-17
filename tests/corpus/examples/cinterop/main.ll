; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [23 x i8] c"mathkit_add(20, 22) = \00", align 1
@.str.1 = private unnamed_addr constant [30 x i8] c"mathkit_clamp(999, 0, 100) = \00", align 1
@.str.2 = private unnamed_addr constant [22 x i8] c"mathkit_hash(1, 2) = \00", align 1
@.str.3 = private unnamed_addr constant [15 x i8] c"MATHKIT_MAX = \00", align 1

declare i32 @mathkit_add(i32, i32)
declare void @aila_std_io__print_text({ ptr, i64 })
declare void @aila_std_io__print_i32_nl(i32)
declare i32 @mathkit_clamp(i32, i32, i32)
declare i32 @mathkit_hash(i32, i32)
declare void @aila_std_io__print_u32(i32)
declare void @aila_std_io__print_nl()

define void @aila_main__main() {
  %slot_s_1 = alloca i32, align 8
  %slot_c_2 = alloca i32, align 8
  %slot_h_3 = alloca i32, align 8
  %t1 = call i32 @mathkit_add(i32 20, i32 22)
  store i32 %t1, ptr %slot_s_1
  %t2 = insertvalue { ptr, i64 } undef, ptr @.str.0, 0
  %t3 = insertvalue { ptr, i64 } %t2, i64 22, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t3)
  %t4 = load i32, ptr %slot_s_1
  call void @aila_std_io__print_i32_nl(i32 %t4)
  %t5 = call i32 @mathkit_clamp(i32 999, i32 0, i32 100)
  store i32 %t5, ptr %slot_c_2
  %t6 = insertvalue { ptr, i64 } undef, ptr @.str.1, 0
  %t7 = insertvalue { ptr, i64 } %t6, i64 29, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t7)
  %t8 = load i32, ptr %slot_c_2
  call void @aila_std_io__print_i32_nl(i32 %t8)
  %t9 = call i32 @mathkit_hash(i32 1, i32 2)
  store i32 %t9, ptr %slot_h_3
  %t10 = insertvalue { ptr, i64 } undef, ptr @.str.2, 0
  %t11 = insertvalue { ptr, i64 } %t10, i64 21, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t11)
  %t12 = load i32, ptr %slot_h_3
  call void @aila_std_io__print_u32(i32 %t12)
  call void @aila_std_io__print_nl()
  %t13 = insertvalue { ptr, i64 } undef, ptr @.str.3, 0
  %t14 = insertvalue { ptr, i64 } %t13, i64 14, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t14)
  call void @aila_std_io__print_i32_nl(i32 1000)
  ret void
}

define i32 @main() {
entry:
  call void @aila_main__main()
  ret i32 0
}
