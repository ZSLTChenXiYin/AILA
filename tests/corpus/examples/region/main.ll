; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.1 = private unnamed_addr constant [10 x i8] c"main.aila\00", align 1
@.str.2 = private unnamed_addr constant [10 x i8] c"sum = %d\0A\00", align 1
@.str.3 = private unnamed_addr constant [10 x i8] c"buf[0] = \00", align 1
@.str.4 = private unnamed_addr constant [9 x i8] c"  len = \00", align 1
@.str.5 = private unnamed_addr constant [13 x i8] c" (bytes %u)\0A\00", align 1
@.str.6 = private unnamed_addr constant [5 x i8] c"AILA\00", align 1

declare ptr @aila_rt_region_alloc(ptr, i64, i64)
declare ptr @aila_rt_region_enter(i64)
declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare void @aila_std_io__printf({ ptr, i64 }, { ptr, i64 })
declare void @aila_std_io__print({ ptr, i64 })
declare void @aila_std_io__println({ ptr, i64 })
declare void @aila_rt_region_leave(ptr)

define { ptr, i64 } @aila_main__fill(ptr %r0, i64 %p_n, i8 %p_v) {
  %slot_n_1 = alloca i64, align 8
  %slot_v_2 = alloca i8, align 8
  %slot___fi_3 = alloca i64, align 8
  store i64 %p_n, ptr %slot_n_1
  store i8 %p_v, ptr %slot_v_2
  %t1 = load i64, ptr %slot_n_1
  %t2 = mul i64 %t1, 1
  %t3 = call ptr @aila_rt_region_alloc(ptr %r0, i64 %t2, i64 1)
  %t4 = load i8, ptr %slot_v_2
  store i64 0, ptr %slot___fi_3
  br label %ai.cond1
ai.cond1:
  %t5 = load i64, ptr %slot___fi_3
  %t6 = icmp ult i64 %t5, %t1
  br i1 %t6, label %ai.body2, label %ai.end3
ai.body2:
  %t7 = getelementptr inbounds i8, ptr %t3, i64 %t5
  store i8 %t4, ptr %t7
  %t8 = add i64 %t5, 1
  store i64 %t8, ptr %slot___fi_3
  br label %ai.cond1
ai.end3:
  %t9 = insertvalue { ptr, i64 } undef, ptr %t3, 0
  %t10 = insertvalue { ptr, i64 } %t9, i64 %t1, 1
  ret { ptr, i64 } %t10
}

define i32 @aila_main__sum_of({ ptr, i64 } %p_buf) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_s_2 = alloca i32, align 8
  %slot_x_3 = alloca i8, align 8
  %slot___idx_4 = alloca i64, align 8
  %slot___iter_5 = alloca { ptr, i64 }, align 8
  %slot___len_6 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i32 0, ptr %slot_s_2
  store i64 0, ptr %slot___idx_4
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  store { ptr, i64 } %t1, ptr %slot___iter_5
  %t3 = getelementptr inbounds { ptr, i64 }, ptr %slot___iter_5, i32 0, i32 1
  %t2 = load i64, ptr %t3
  store i64 %t2, ptr %slot___len_6
  %t4 = load ptr, ptr %slot___iter_5
  br label %fi.cond2
fi.cond2:
  %t5 = load i64, ptr %slot___idx_4
  %t6 = load i64, ptr %slot___len_6
  %t7 = icmp ult i64 %t5, %t6
  br i1 %t7, label %fi.body4, label %for.end1
fi.body4:
  %t8 = getelementptr inbounds i8, ptr %t4, i64 %t5
  %t9 = load i8, ptr %t8
  store i8 %t9, ptr %slot_x_3
  %t10 = load i32, ptr %slot_s_2
  %t11 = load i8, ptr %slot_x_3
  %t12 = zext i8 %t11 to i32
  %t13 = add i32 %t10, %t12
  store i32 %t13, ptr %slot_s_2
  br label %fi.step3
fi.step3:
  %t14 = add i64 %t5, 1
  store i64 %t14, ptr %slot___idx_4
  br label %fi.cond2
for.end1:
  %t15 = load i32, ptr %slot_s_2
  ret i32 %t15
}

define void @aila_main__main() {
  %slot_nums_1 = alloca { ptr, i64 }, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_text_3 = alloca { ptr, i64 }, align 8
  %slot_bytes_4 = alloca { ptr, i64 }, align 8
  %slot___lit_5 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_6 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_7 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_8 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_9 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_10 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_11 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_12 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_13 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_14 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_15 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_16 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_17 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_18 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 4, i64 1)
  %t3 = getelementptr inbounds i8, ptr %t2, i64 0
  store i8 5, ptr %t3
  %t4 = getelementptr inbounds i8, ptr %t2, i64 1
  store i8 5, ptr %t4
  %t5 = getelementptr inbounds i8, ptr %t2, i64 2
  store i8 5, ptr %t5
  %t6 = getelementptr inbounds i8, ptr %t2, i64 3
  store i8 5, ptr %t6
  %t7 = insertvalue { ptr, i64 } undef, ptr %t2, 0
  %t8 = insertvalue { ptr, i64 } %t7, i64 4, 1
  store { ptr, i64 } %t8, ptr %slot_nums_1
  %t9 = load { ptr, i64 }, ptr %slot_nums_1
  %t10 = extractvalue { ptr, i64 } %t9, 0
  %t11 = extractvalue { ptr, i64 } %t9, 1
  %t12 = icmp uge i64 0, %t11
  br i1 %t12, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 26)
  unreachable
chk.ok1:
  %t13 = getelementptr inbounds i8, ptr %t10, i64 0
  store i8 5, ptr %t13
  %t14 = insertvalue { ptr, i64 } undef, ptr @.str.2, 0
  %t15 = insertvalue { ptr, i64 } %t14, i64 9, 1
  %t16 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t16, ptr %slot___lit_5
  %t17 = load { ptr, i64 }, ptr %slot_nums_1
  %t18 = call i32 @aila_main__sum_of({ ptr, i64 } %t17)
  %t19 = sext i32 %t18 to i64
  %t20 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_5, i32 0, i32 1, i32 0
  store i64 %t19, ptr %t20
  %t21 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_5
  %t22 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_6, i64 0
  store { i8, { { ptr, i64 } } } %t21, ptr %t22
  %t23 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_6, 0
  %t24 = insertvalue { ptr, i64 } %t23, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t15, { ptr, i64 } %t24)
  %t25 = call { ptr, i64 } @aila_main__fill(ptr %t1, i64 4, i8 7)
  store { ptr, i64 } %t25, ptr %slot_buf_2
  %t26 = load { ptr, i64 }, ptr %slot_buf_2
  %t27 = extractvalue { ptr, i64 } %t26, 0
  %t28 = extractvalue { ptr, i64 } %t26, 1
  %t29 = icmp uge i64 0, %t28
  br i1 %t29, label %chk.bad4, label %chk.ok3
chk.bad4:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 31)
  unreachable
chk.ok3:
  %t30 = getelementptr inbounds i8, ptr %t27, i64 0
  store i8 7, ptr %t30
  %t31 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t31, ptr %slot___lit_7
  %t32 = insertvalue { ptr, i64 } undef, ptr @.str.3, 0
  %t33 = insertvalue { ptr, i64 } %t32, i64 9, 1
  %t34 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_7, i32 0, i32 1, i32 0
  store { ptr, i64 } %t33, ptr %t34
  %t35 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_7
  %t36 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_8, i64 0
  store { i8, { { ptr, i64 } } } %t35, ptr %t36
  %t37 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_8, 0
  %t38 = insertvalue { ptr, i64 } %t37, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t38)
  %t39 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t39, ptr %slot___lit_9
  %t40 = load { ptr, i64 }, ptr %slot_buf_2
  %t41 = extractvalue { ptr, i64 } %t40, 0
  %t42 = extractvalue { ptr, i64 } %t40, 1
  %t43 = icmp uge i64 0, %t42
  br i1 %t43, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 34)
  unreachable
chk.ok5:
  %t44 = getelementptr inbounds i8, ptr %t41, i64 0
  %t45 = load i8, ptr %t44
  %t46 = zext i8 %t45 to i64
  %t47 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_9, i32 0, i32 1, i32 0
  store i64 %t46, ptr %t47
  %t48 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_9
  %t49 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_10, i64 0
  store { i8, { { ptr, i64 } } } %t48, ptr %t49
  %t50 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_10, 0
  %t51 = insertvalue { ptr, i64 } %t50, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t51)
  %t52 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t52, ptr %slot___lit_11
  %t53 = insertvalue { ptr, i64 } undef, ptr @.str.4, 0
  %t54 = insertvalue { ptr, i64 } %t53, i64 8, 1
  %t55 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_11, i32 0, i32 1, i32 0
  store { ptr, i64 } %t54, ptr %t55
  %t56 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_11
  %t57 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_12, i64 0
  store { i8, { { ptr, i64 } } } %t56, ptr %t57
  %t58 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_12, 0
  %t59 = insertvalue { ptr, i64 } %t58, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t59)
  %t60 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t60, ptr %slot___lit_13
  %t61 = load { ptr, i64 }, ptr %slot_buf_2
  %t62 = extractvalue { ptr, i64 } %t61, 1
  %t64 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_13, i32 0, i32 1, i32 0
  store i64 %t62, ptr %t64
  %t65 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_13
  %t66 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_14, i64 0
  store { i8, { { ptr, i64 } } } %t65, ptr %t66
  %t67 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_14, 0
  %t68 = insertvalue { ptr, i64 } %t67, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t68)
  %t69 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t70 = insertvalue { ptr, i64 } %t69, i64 12, 1
  %t71 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t71, ptr %slot___lit_15
  %t72 = mul i64 4, 4
  %t74 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_15, i32 0, i32 1, i32 0
  store i64 %t72, ptr %t74
  %t75 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_15
  %t76 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_16, i64 0
  store { i8, { { ptr, i64 } } } %t75, ptr %t76
  %t77 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_16, 0
  %t78 = insertvalue { ptr, i64 } %t77, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t70, { ptr, i64 } %t78)
  %t79 = insertvalue { ptr, i64 } undef, ptr @.str.6, 0
  %t80 = insertvalue { ptr, i64 } %t79, i64 4, 1
  store { ptr, i64 } %t80, ptr %slot_text_3
  %t81 = load { ptr, i64 }, ptr %slot_text_3
  store { ptr, i64 } %t81, ptr %slot_bytes_4
  %t82 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t82, ptr %slot___lit_17
  %t83 = load { ptr, i64 }, ptr %slot_bytes_4
  %t84 = extractvalue { ptr, i64 } %t83, 1
  %t86 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_17, i32 0, i32 1, i32 0
  store i64 %t84, ptr %t86
  %t87 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_17
  %t88 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_18, i64 0
  store { i8, { { ptr, i64 } } } %t87, ptr %t88
  %t89 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_18, 0
  %t90 = insertvalue { ptr, i64 } %t89, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t90)
  call void @aila_rt_region_leave(ptr %t1)
  ret void
}

define i32 @main(i32 %argc, ptr %argv) {
entry:
  call void @aila_main__main()
  ret i32 0
}
