; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.1 = private unnamed_addr constant [10 x i8] c"main.aila\00", align 1
@.str.2 = private unnamed_addr constant [2 x i8] c"-\00", align 1
@.str.3 = private unnamed_addr constant [2 x i8] c"a\00", align 1
@.str.4 = private unnamed_addr constant [2 x i8] c"b\00", align 1
@.str.5 = private unnamed_addr constant [2 x i8] c"c\00", align 1

declare void @aila_std_io__print({ ptr, i64 })
declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare void @aila_std_io__println({ ptr, i64 })

define i32 @aila_main__sum(i32 %p_base, { ptr, i64 } %p_rest) {
  %slot_base_1 = alloca i32, align 8
  %slot_rest_2 = alloca { ptr, i64 }, align 8
  %slot_s_3 = alloca i32, align 8
  %slot_v_4 = alloca i32, align 8
  %slot___idx_5 = alloca i64, align 8
  %slot___iter_6 = alloca { ptr, i64 }, align 8
  %slot___len_7 = alloca i64, align 8
  store i32 %p_base, ptr %slot_base_1
  store { ptr, i64 } %p_rest, ptr %slot_rest_2
  %t1 = load i32, ptr %slot_base_1
  store i32 %t1, ptr %slot_s_3
  store i64 0, ptr %slot___idx_5
  %t2 = load { ptr, i64 }, ptr %slot_rest_2
  store { ptr, i64 } %t2, ptr %slot___iter_6
  %t4 = getelementptr inbounds { ptr, i64 }, ptr %slot___iter_6, i32 0, i32 1
  %t3 = load i64, ptr %t4
  store i64 %t3, ptr %slot___len_7
  %t5 = load ptr, ptr %slot___iter_6
  br label %fi.cond2
fi.cond2:
  %t6 = load i64, ptr %slot___idx_5
  %t7 = load i64, ptr %slot___len_7
  %t8 = icmp ult i64 %t6, %t7
  br i1 %t8, label %fi.body4, label %for.end1
fi.body4:
  %t9 = getelementptr inbounds i32, ptr %t5, i64 %t6
  %t10 = load i32, ptr %t9
  store i32 %t10, ptr %slot_v_4
  %t11 = load i32, ptr %slot_s_3
  %t12 = load i32, ptr %slot_v_4
  %t13 = add i32 %t11, %t12
  store i32 %t13, ptr %slot_s_3
  br label %fi.step3
fi.step3:
  %t14 = add i64 %t6, 1
  store i64 %t14, ptr %slot___idx_5
  br label %fi.cond2
for.end1:
  %t15 = load i32, ptr %slot_s_3
  ret i32 %t15
}

define void @aila_main__join({ ptr, i64 } %p_sep, { ptr, i64 } %p_rest) {
  %slot_sep_1 = alloca { ptr, i64 }, align 8
  %slot_rest_2 = alloca { ptr, i64 }, align 8
  %slot_first_3 = alloca i1, align 8
  %slot_t_4 = alloca { ptr, i64 }, align 8
  %slot___idx_5 = alloca i64, align 8
  %slot___iter_6 = alloca { ptr, i64 }, align 8
  %slot___len_7 = alloca i64, align 8
  %slot___lit_8 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_9 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_10 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_11 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  store { ptr, i64 } %p_sep, ptr %slot_sep_1
  store { ptr, i64 } %p_rest, ptr %slot_rest_2
  store i1 1, ptr %slot_first_3
  store i64 0, ptr %slot___idx_5
  %t1 = load { ptr, i64 }, ptr %slot_rest_2
  store { ptr, i64 } %t1, ptr %slot___iter_6
  %t3 = getelementptr inbounds { ptr, i64 }, ptr %slot___iter_6, i32 0, i32 1
  %t2 = load i64, ptr %t3
  store i64 %t2, ptr %slot___len_7
  %t4 = load ptr, ptr %slot___iter_6
  br label %fi.cond2
fi.cond2:
  %t5 = load i64, ptr %slot___idx_5
  %t6 = load i64, ptr %slot___len_7
  %t7 = icmp ult i64 %t5, %t6
  br i1 %t7, label %fi.body4, label %for.end1
fi.body4:
  %t8 = getelementptr inbounds { ptr, i64 }, ptr %t4, i64 %t5
  %t9 = load { ptr, i64 }, ptr %t8
  store { ptr, i64 } %t9, ptr %slot_t_4
  %t10 = load i1, ptr %slot_first_3
  %t11 = xor i1 %t10, true
  br i1 %t11, label %if.then6, label %if.next7
if.then6:
  %t12 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t12, ptr %slot___lit_8
  %t13 = load { ptr, i64 }, ptr %slot_sep_1
  %t14 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_8, i32 0, i32 1, i32 0
  store { ptr, i64 } %t13, ptr %t14
  %t15 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_8
  %t16 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_9, i64 0
  store { i8, { { ptr, i64 } } } %t15, ptr %t16
  %t17 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_9, 0
  %t18 = insertvalue { ptr, i64 } %t17, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t18)
  br label %if.end5
if.next7:
  br label %if.end5
if.end5:
  %t19 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t19, ptr %slot___lit_10
  %t20 = load { ptr, i64 }, ptr %slot_t_4
  %t21 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_10, i32 0, i32 1, i32 0
  store { ptr, i64 } %t20, ptr %t21
  %t22 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_10
  %t23 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_11, i64 0
  store { i8, { { ptr, i64 } } } %t22, ptr %t23
  %t24 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_11, 0
  %t25 = insertvalue { ptr, i64 } %t24, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t25)
  store i1 0, ptr %slot_first_3
  br label %fi.step3
fi.step3:
  %t26 = add i64 %t5, 1
  store i64 %t26, ptr %slot___idx_5
  br label %fi.cond2
for.end1:
  ret void
}

define i64 @aila_main__pick({ ptr, i64 } %p_rest) {
  %slot_rest_1 = alloca { ptr, i64 }, align 8
  store { ptr, i64 } %p_rest, ptr %slot_rest_1
  %t1 = load { ptr, i64 }, ptr %slot_rest_1
  %t2 = extractvalue { ptr, i64 } %t1, 0
  %t3 = extractvalue { ptr, i64 } %t1, 1
  %t4 = icmp uge i64 0, %t3
  br i1 %t4, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 33)
  unreachable
chk.ok1:
  %t5 = getelementptr inbounds i64, ptr %t2, i64 0
  %t6 = load i64, ptr %t5
  %t7 = load { ptr, i64 }, ptr %slot_rest_1
  %t8 = extractvalue { ptr, i64 } %t7, 1
  %t9 = add i64 %t6, %t8
  ret i64 %t9
}

define void @aila_main__main() {
  %slot___lit_1 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_2 = alloca [3 x i32], align 8
  %slot_vargs_3 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_4 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_5 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot_vargs_6 = alloca [3 x { ptr, i64 }], align 8
  %slot___lit_7 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_8 = alloca [3 x i64], align 8
  %slot_vargs_9 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %t1 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t1, ptr %slot___lit_1
  %t2 = getelementptr inbounds i32, ptr %slot_vargs_2, i64 0
  store i32 2, ptr %t2
  %t3 = getelementptr inbounds i32, ptr %slot_vargs_2, i64 1
  store i32 3, ptr %t3
  %t4 = getelementptr inbounds i32, ptr %slot_vargs_2, i64 2
  store i32 4, ptr %t4
  %t5 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_2, 0
  %t6 = insertvalue { ptr, i64 } %t5, i64 3, 1
  %t7 = call i32 @aila_main__sum(i32 1, { ptr, i64 } %t6)
  %t8 = sext i32 %t7 to i64
  %t9 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_1, i32 0, i32 1, i32 0
  store i64 %t8, ptr %t9
  %t10 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_1
  %t11 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_3, i64 0
  store { i8, { { ptr, i64 } } } %t10, ptr %t11
  %t12 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_3, 0
  %t13 = insertvalue { ptr, i64 } %t12, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t13)
  %t14 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t14, ptr %slot___lit_4
  %t15 = call i32 @aila_main__sum(i32 7, { ptr, i64 } zeroinitializer)
  %t16 = sext i32 %t15 to i64
  %t17 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_4, i32 0, i32 1, i32 0
  store i64 %t16, ptr %t17
  %t18 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_4
  %t19 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_5, i64 0
  store { i8, { { ptr, i64 } } } %t18, ptr %t19
  %t20 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_5, 0
  %t21 = insertvalue { ptr, i64 } %t20, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t21)
  %t22 = insertvalue { ptr, i64 } undef, ptr @.str.2, 0
  %t23 = insertvalue { ptr, i64 } %t22, i64 1, 1
  %t24 = insertvalue { ptr, i64 } undef, ptr @.str.3, 0
  %t25 = insertvalue { ptr, i64 } %t24, i64 1, 1
  %t26 = insertvalue { ptr, i64 } undef, ptr @.str.4, 0
  %t27 = insertvalue { ptr, i64 } %t26, i64 1, 1
  %t28 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t29 = insertvalue { ptr, i64 } %t28, i64 1, 1
  %t30 = getelementptr inbounds { ptr, i64 }, ptr %slot_vargs_6, i64 0
  store { ptr, i64 } %t25, ptr %t30
  %t31 = getelementptr inbounds { ptr, i64 }, ptr %slot_vargs_6, i64 1
  store { ptr, i64 } %t27, ptr %t31
  %t32 = getelementptr inbounds { ptr, i64 }, ptr %slot_vargs_6, i64 2
  store { ptr, i64 } %t29, ptr %t32
  %t33 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_6, 0
  %t34 = insertvalue { ptr, i64 } %t33, i64 3, 1
  call void @aila_main__join({ ptr, i64 } %t23, { ptr, i64 } %t34)
  call void @aila_std_io__println({ ptr, i64 } zeroinitializer)
  %t35 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t35, ptr %slot___lit_7
  %t36 = getelementptr inbounds i64, ptr %slot_vargs_8, i64 0
  store i64 10, ptr %t36
  %t37 = getelementptr inbounds i64, ptr %slot_vargs_8, i64 1
  store i64 20, ptr %t37
  %t38 = getelementptr inbounds i64, ptr %slot_vargs_8, i64 2
  store i64 30, ptr %t38
  %t39 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_8, 0
  %t40 = insertvalue { ptr, i64 } %t39, i64 3, 1
  %t41 = call i64 @aila_main__pick({ ptr, i64 } %t40)
  %t43 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_7, i32 0, i32 1, i32 0
  store i64 %t41, ptr %t43
  %t44 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_7
  %t45 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_9, i64 0
  store { i8, { { ptr, i64 } } } %t44, ptr %t45
  %t46 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_9, 0
  %t47 = insertvalue { ptr, i64 } %t46, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t47)
  ret void
}

define i32 @main(i32 %argc, ptr %argv) {
entry:
  call void @aila_main__main()
  ret i32 0
}
