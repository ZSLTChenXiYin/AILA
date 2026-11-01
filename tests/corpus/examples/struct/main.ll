; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [35 x i8] c"Detection{label=%d score=%f bbox=[\00", align 1
@.str.1 = private unnamed_addr constant [3 x i8] c", \00", align 1
@.str.2 = private unnamed_addr constant [3 x i8] c"]}\00", align 1
@.str.3 = private unnamed_addr constant [18 x i8] c"norm2(3, 4) = %d\0A\00", align 1
@.str.4 = private unnamed_addr constant [21 x i8] c"sizeof(Detection) = \00", align 1
@.str.5 = private unnamed_addr constant [11 x i8] c"  align = \00", align 1

declare void @aila_std_io__printf({ ptr, i64 }, { ptr, i64 })
declare void @aila_std_io__print({ ptr, i64 })
declare void @aila_std_io__println({ ptr, i64 })

define i32 @aila_main__Detection__norm2(ptr %p_self) {
  %slot_self_1 = alloca { i32, float, [4 x float] }, align 8
  %slot_x_2 = alloca i32, align 8
  %t1 = getelementptr inbounds { i32, float, [4 x float] }, ptr %p_self, i32 0, i32 2
  %t2 = getelementptr inbounds [4 x float], ptr %t1, i64 0, i64 0
  %t3 = load float, ptr %t2
  %t4 = fptosi float %t3 to i32
  store i32 %t4, ptr %slot_x_2
  %t5 = getelementptr inbounds { i32, float, [4 x float] }, ptr %p_self, i32 0, i32 0
  %t6 = load i32, ptr %t5
  %t7 = getelementptr inbounds { i32, float, [4 x float] }, ptr %p_self, i32 0, i32 0
  %t8 = load i32, ptr %t7
  %t9 = mul i32 %t6, %t8
  %t10 = load i32, ptr %slot_x_2
  %t11 = load i32, ptr %slot_x_2
  %t12 = mul i32 %t10, %t11
  %t13 = add i32 %t9, %t12
  ret i32 %t13
}

define void @aila_main__Detection__describe(ptr %p_self) {
  %slot_self_1 = alloca { i32, float, [4 x float] }, align 8
  %slot_i_2 = alloca i64, align 8
  %slot_v_3 = alloca float, align 8
  %slot___lit_4 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___lit_5 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_6 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot___idx_7 = alloca i64, align 8
  %slot___lit_8 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_9 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_10 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_11 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_12 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_13 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %t1 = insertvalue { ptr, i64 } undef, ptr @.str.0, 0
  %t2 = insertvalue { ptr, i64 } %t1, i64 34, 1
  %t3 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t3, ptr %slot___lit_4
  %t4 = getelementptr inbounds { i32, float, [4 x float] }, ptr %p_self, i32 0, i32 0
  %t5 = load i32, ptr %t4
  %t6 = sext i32 %t5 to i64
  %t7 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_4, i32 0, i32 1, i32 0
  store i64 %t6, ptr %t7
  %t8 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_4
  %t9 = insertvalue { i8, { { ptr, i64 } } } undef, i8 2, 0
  store { i8, { { ptr, i64 } } } %t9, ptr %slot___lit_5
  %t10 = getelementptr inbounds { i32, float, [4 x float] }, ptr %p_self, i32 0, i32 1
  %t11 = load float, ptr %t10
  %t12 = fpext float %t11 to double
  %t13 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_5, i32 0, i32 1, i32 0
  store double %t12, ptr %t13
  %t14 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_5
  %t15 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_6, i64 0
  store { i8, { { ptr, i64 } } } %t8, ptr %t15
  %t16 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_6, i64 1
  store { i8, { { ptr, i64 } } } %t14, ptr %t16
  %t17 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_6, 0
  %t18 = insertvalue { ptr, i64 } %t17, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t2, { ptr, i64 } %t18)
  store i64 0, ptr %slot___idx_7
  %t19 = getelementptr inbounds { i32, float, [4 x float] }, ptr %p_self, i32 0, i32 2
  br label %fi.cond2
fi.cond2:
  %t20 = load i64, ptr %slot___idx_7
  %t21 = icmp ult i64 %t20, 4
  br i1 %t21, label %fi.body4, label %for.end1
fi.body4:
  %t22 = getelementptr inbounds [4 x float], ptr %t19, i64 0, i64 %t20
  %t23 = load float, ptr %t22
  store float %t23, ptr %slot_v_3
  store i64 %t20, ptr %slot_i_2
  %t24 = load i64, ptr %slot_i_2
  %t25 = icmp ugt i64 %t24, 0
  br i1 %t25, label %if.then6, label %if.next7
if.then6:
  %t26 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t26, ptr %slot___lit_8
  %t27 = insertvalue { ptr, i64 } undef, ptr @.str.1, 0
  %t28 = insertvalue { ptr, i64 } %t27, i64 2, 1
  %t29 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_8, i32 0, i32 1, i32 0
  store { ptr, i64 } %t28, ptr %t29
  %t30 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_8
  %t31 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_9, i64 0
  store { i8, { { ptr, i64 } } } %t30, ptr %t31
  %t32 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_9, 0
  %t33 = insertvalue { ptr, i64 } %t32, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t33)
  br label %if.end5
if.next7:
  br label %if.end5
if.end5:
  %t34 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t34, ptr %slot___lit_10
  %t35 = load float, ptr %slot_v_3
  %t36 = fptosi float %t35 to i32
  %t37 = sext i32 %t36 to i64
  %t38 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_10, i32 0, i32 1, i32 0
  store i64 %t37, ptr %t38
  %t39 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_10
  %t40 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_11, i64 0
  store { i8, { { ptr, i64 } } } %t39, ptr %t40
  %t41 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_11, 0
  %t42 = insertvalue { ptr, i64 } %t41, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t42)
  br label %fi.step3
fi.step3:
  %t43 = add i64 %t20, 1
  store i64 %t43, ptr %slot___idx_7
  br label %fi.cond2
for.end1:
  %t44 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t44, ptr %slot___lit_12
  %t45 = insertvalue { ptr, i64 } undef, ptr @.str.2, 0
  %t46 = insertvalue { ptr, i64 } %t45, i64 2, 1
  %t47 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_12, i32 0, i32 1, i32 0
  store { ptr, i64 } %t46, ptr %t47
  %t48 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_12
  %t49 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_13, i64 0
  store { i8, { { ptr, i64 } } } %t48, ptr %t49
  %t50 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_13, 0
  %t51 = insertvalue { ptr, i64 } %t50, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t51)
  ret void
}

define void @aila_main__main() {
  %slot_d_1 = alloca { i32, float, [4 x float] }, align 8
  %slot___lit_2 = alloca { i32, float, [4 x float] }, align 8
  %slot___arr_3 = alloca [4 x float], align 8
  %slot___lit_4 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_5 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_6 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_7 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_8 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_9 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_10 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_11 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_12 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_13 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %t1 = getelementptr inbounds { i32, float, [4 x float] }, ptr %slot___lit_2, i32 0, i32 0
  store i32 3, ptr %t1
  %t2 = getelementptr inbounds { i32, float, [4 x float] }, ptr %slot___lit_2, i32 0, i32 1
  store float 0.75, ptr %t2
  %t3 = getelementptr inbounds [4 x float], ptr %slot___arr_3, i64 0, i64 0
  store float 1.0, ptr %t3
  %t4 = getelementptr inbounds [4 x float], ptr %slot___arr_3, i64 0, i64 1
  store float 2.0, ptr %t4
  %t5 = getelementptr inbounds [4 x float], ptr %slot___arr_3, i64 0, i64 2
  store float 3.0, ptr %t5
  %t6 = getelementptr inbounds [4 x float], ptr %slot___arr_3, i64 0, i64 3
  store float 4.0, ptr %t6
  %t7 = load [4 x float], ptr %slot___arr_3
  %t8 = getelementptr inbounds { i32, float, [4 x float] }, ptr %slot___lit_2, i32 0, i32 2
  store [4 x float] %t7, ptr %t8
  %t9 = load { i32, float, [4 x float] }, ptr %slot___lit_2
  store { i32, float, [4 x float] } %t9, ptr %slot_d_1
  %t10 = insertvalue { ptr, i64 } undef, ptr @.str.3, 0
  %t11 = insertvalue { ptr, i64 } %t10, i64 17, 1
  %t12 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t12, ptr %slot___lit_4
  %t13 = call i32 @aila_main__Detection__norm2(ptr %slot_d_1)
  %t14 = sext i32 %t13 to i64
  %t15 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_4, i32 0, i32 1, i32 0
  store i64 %t14, ptr %t15
  %t16 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_4
  %t17 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_5, i64 0
  store { i8, { { ptr, i64 } } } %t16, ptr %t17
  %t18 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_5, 0
  %t19 = insertvalue { ptr, i64 } %t18, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t11, { ptr, i64 } %t19)
  call void @aila_main__Detection__describe(ptr %slot_d_1)
  %t20 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t20, ptr %slot___lit_6
  %t21 = insertvalue { ptr, i64 } undef, ptr @.str.4, 0
  %t22 = insertvalue { ptr, i64 } %t21, i64 20, 1
  %t23 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_6, i32 0, i32 1, i32 0
  store { ptr, i64 } %t22, ptr %t23
  %t24 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_6
  %t25 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_7, i64 0
  store { i8, { { ptr, i64 } } } %t24, ptr %t25
  %t26 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_7, 0
  %t27 = insertvalue { ptr, i64 } %t26, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t27)
  %t28 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t28, ptr %slot___lit_8
  %t30 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_8, i32 0, i32 1, i32 0
  store i64 24, ptr %t30
  %t31 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_8
  %t32 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_9, i64 0
  store { i8, { { ptr, i64 } } } %t31, ptr %t32
  %t33 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_9, 0
  %t34 = insertvalue { ptr, i64 } %t33, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t34)
  %t35 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t35, ptr %slot___lit_10
  %t36 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t37 = insertvalue { ptr, i64 } %t36, i64 10, 1
  %t38 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_10, i32 0, i32 1, i32 0
  store { ptr, i64 } %t37, ptr %t38
  %t39 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_10
  %t40 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_11, i64 0
  store { i8, { { ptr, i64 } } } %t39, ptr %t40
  %t41 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_11, 0
  %t42 = insertvalue { ptr, i64 } %t41, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t42)
  %t43 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t43, ptr %slot___lit_12
  %t45 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_12, i32 0, i32 1, i32 0
  store i64 4, ptr %t45
  %t46 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_12
  %t47 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_13, i64 0
  store { i8, { { ptr, i64 } } } %t46, ptr %t47
  %t48 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_13, 0
  %t49 = insertvalue { ptr, i64 } %t48, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t49)
  ret void
}

define i32 @main(i32 %argc, ptr %argv) {
entry:
  call void @aila_main__main()
  ret i32 0
}
