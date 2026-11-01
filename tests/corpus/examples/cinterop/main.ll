; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [26 x i8] c"mathkit_add(20, 22) = %d\0A\00", align 1
@.str.1 = private unnamed_addr constant [33 x i8] c"mathkit_clamp(999, 0, 100) = %d\0A\00", align 1
@.str.2 = private unnamed_addr constant [25 x i8] c"mathkit_hash(1, 2) = %u\0A\00", align 1
@.str.3 = private unnamed_addr constant [18 x i8] c"MATHKIT_MAX = %d\0A\00", align 1

declare i32 @mathkit_add(i32, i32)
declare void @aila_std_io__printf({ ptr, i64 }, { ptr, i64 })
declare i32 @mathkit_clamp(i32, i32, i32)
declare i32 @mathkit_hash(i32, i32)

define void @aila_main__main() {
  %slot_s_1 = alloca i32, align 8
  %slot_c_2 = alloca i32, align 8
  %slot_h_3 = alloca i32, align 8
  %slot___lit_4 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_5 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_6 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_7 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_8 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_9 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_10 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_11 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %t1 = call i32 @mathkit_add(i32 20, i32 22)
  store i32 %t1, ptr %slot_s_1
  %t2 = insertvalue { ptr, i64 } undef, ptr @.str.0, 0
  %t3 = insertvalue { ptr, i64 } %t2, i64 25, 1
  %t4 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t4, ptr %slot___lit_4
  %t5 = load i32, ptr %slot_s_1
  %t6 = sext i32 %t5 to i64
  %t7 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_4, i32 0, i32 1, i32 0
  store i64 %t6, ptr %t7
  %t8 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_4
  %t9 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_5, i64 0
  store { i8, { { ptr, i64 } } } %t8, ptr %t9
  %t10 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_5, 0
  %t11 = insertvalue { ptr, i64 } %t10, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t3, { ptr, i64 } %t11)
  %t12 = call i32 @mathkit_clamp(i32 999, i32 0, i32 100)
  store i32 %t12, ptr %slot_c_2
  %t13 = insertvalue { ptr, i64 } undef, ptr @.str.1, 0
  %t14 = insertvalue { ptr, i64 } %t13, i64 32, 1
  %t15 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t15, ptr %slot___lit_6
  %t16 = load i32, ptr %slot_c_2
  %t17 = sext i32 %t16 to i64
  %t18 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_6, i32 0, i32 1, i32 0
  store i64 %t17, ptr %t18
  %t19 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_6
  %t20 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_7, i64 0
  store { i8, { { ptr, i64 } } } %t19, ptr %t20
  %t21 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_7, 0
  %t22 = insertvalue { ptr, i64 } %t21, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t14, { ptr, i64 } %t22)
  %t23 = call i32 @mathkit_hash(i32 1, i32 2)
  store i32 %t23, ptr %slot_h_3
  %t24 = insertvalue { ptr, i64 } undef, ptr @.str.2, 0
  %t25 = insertvalue { ptr, i64 } %t24, i64 24, 1
  %t26 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t26, ptr %slot___lit_8
  %t27 = load i32, ptr %slot_h_3
  %t28 = zext i32 %t27 to i64
  %t29 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_8, i32 0, i32 1, i32 0
  store i64 %t28, ptr %t29
  %t30 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_8
  %t31 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_9, i64 0
  store { i8, { { ptr, i64 } } } %t30, ptr %t31
  %t32 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_9, 0
  %t33 = insertvalue { ptr, i64 } %t32, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t25, { ptr, i64 } %t33)
  %t34 = insertvalue { ptr, i64 } undef, ptr @.str.3, 0
  %t35 = insertvalue { ptr, i64 } %t34, i64 17, 1
  %t36 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t36, ptr %slot___lit_10
  %t37 = sext i32 1000 to i64
  %t38 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_10, i32 0, i32 1, i32 0
  store i64 %t37, ptr %t38
  %t39 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_10
  %t40 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_11, i64 0
  store { i8, { { ptr, i64 } } } %t39, ptr %t40
  %t41 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_11, 0
  %t42 = insertvalue { ptr, i64 } %t41, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t35, { ptr, i64 } %t42)
  ret void
}

define i32 @main(i32 %argc, ptr %argv) {
entry:
  call void @aila_main__main()
  ret i32 0
}
