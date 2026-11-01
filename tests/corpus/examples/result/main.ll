; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.1 = private unnamed_addr constant [10 x i8] c"main.aila\00", align 1
@.str.2 = private unnamed_addr constant [28 x i8] c"match \E6\9C\AA\E8\A6\86\E7\9B\96\E7\9A\84\E5\88\A4\E5\88\AB\E5\80\BC\00", align 1
@.str.3 = private unnamed_addr constant [46 x i8] c"parse(4 bytes) -> Truncated{got=%u, need=%u}\0A\00", align 1
@.str.4 = private unnamed_addr constant [13 x i8] c"parse failed\00", align 1
@.str.5 = private unnamed_addr constant [32 x i8] c"safe parse failed, fallback = 0\00", align 1
@.str.6 = private unnamed_addr constant [17 x i8] c"recovered by ?: \00", align 1

declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare ptr @aila_rt_region_enter(i64)
declare ptr @aila_rt_region_alloc(ptr, i64, i64)
declare void @aila_std_io__println({ ptr, i64 })
declare void @aila_std_io__printf({ ptr, i64 }, { ptr, i64 })
declare void @aila_std_io__print({ ptr, i64 })
declare void @aila_rt_region_leave(ptr)

define { i1, i32, { i8, { i64, i64 } } } @aila_main__parse_header({ ptr, i64 } %p_buf) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot___lit_2 = alloca { i8, { i64, i64 } }, align 8
  %slot___lit_3 = alloca { i8, { i64, i64 } }, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = extractvalue { ptr, i64 } %t1, 1
  %t3 = icmp ult i64 %t2, 8
  br i1 %t3, label %if.then2, label %if.next3
if.then2:
  %t4 = insertvalue { i8, { i64, i64 } } undef, i8 1, 0
  store { i8, { i64, i64 } } %t4, ptr %slot___lit_2
  %t5 = load { ptr, i64 }, ptr %slot_buf_1
  %t6 = extractvalue { ptr, i64 } %t5, 1
  %t7 = getelementptr inbounds { i8, { i64, i64 } }, ptr %slot___lit_2, i32 0, i32 1, i32 0
  store i64 %t6, ptr %t7
  %t8 = getelementptr inbounds { i8, { i64, i64 } }, ptr %slot___lit_2, i32 0, i32 1, i32 1
  store i64 8, ptr %t8
  %t9 = load { i8, { i64, i64 } }, ptr %slot___lit_2
  %t10 = insertvalue { i1, i32, { i8, { i64, i64 } } } undef, i1 0, 0
  %t11 = insertvalue { i1, i32, { i8, { i64, i64 } } } %t10, { i8, { i64, i64 } } %t9, 2
  ret { i1, i32, { i8, { i64, i64 } } } %t11
if.next3:
  br label %if.end1
if.end1:
  %t12 = load { ptr, i64 }, ptr %slot_buf_1
  %t13 = extractvalue { ptr, i64 } %t12, 0
  %t14 = extractvalue { ptr, i64 } %t12, 1
  %t15 = icmp uge i64 0, %t14
  br i1 %t15, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 19)
  unreachable
chk.ok7:
  %t16 = getelementptr inbounds i8, ptr %t13, i64 0
  %t17 = load i8, ptr %t16
  %t18 = icmp ne i8 %t17, 65
  br i1 %t18, label %if.then5, label %if.next6
if.then5:
  %t19 = insertvalue { i8, { i64, i64 } } undef, i8 0, 0
  %t20 = insertvalue { i1, i32, { i8, { i64, i64 } } } undef, i1 0, 0
  %t21 = insertvalue { i1, i32, { i8, { i64, i64 } } } %t20, { i8, { i64, i64 } } %t19, 2
  ret { i1, i32, { i8, { i64, i64 } } } %t21
if.next6:
  br label %if.end4
if.end4:
  %t22 = insertvalue { i1, i32, { i8, { i64, i64 } } } undef, i1 1, 0
  %t23 = insertvalue { i1, i32, { i8, { i64, i64 } } } %t22, i32 42, 1
  ret { i1, i32, { i8, { i64, i64 } } } %t23
}

define i32 @aila_main__maybe_scale(i32 %p_v, i32 %p_factor) {
  %slot_v_1 = alloca i32, align 8
  %slot_factor_2 = alloca i32, align 8
  %slot_scaled_3 = alloca i32, align 8
  store i32 %p_v, ptr %slot_v_1
  store i32 %p_factor, ptr %slot_factor_2
  store i32 0, ptr %slot_scaled_3
  %t1 = load i32, ptr %slot_factor_2
  %t2 = icmp ne i32 %t1, 0
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  %t3 = load i32, ptr %slot_v_1
  %t4 = load i32, ptr %slot_factor_2
  %t5 = mul i32 %t3, %t4
  store i32 %t5, ptr %slot_scaled_3
  br label %if.end1
if.next3:
  br label %if.end1
if.end1:
  %t6 = load i32, ptr %slot_scaled_3
  ret i32 %t6
}

define { i1, i32, { i8, { i64, i64 } } } @aila_main__read_or_default({ ptr, i64 } %p_buf) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_v_2 = alloca i32, align 8
  %slot___try_3 = alloca { i1, i32, { i8, { i64, i64 } } }, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = call { i1, i32, { i8, { i64, i64 } } } @aila_main__parse_header({ ptr, i64 } %t1)
  store { i1, i32, { i8, { i64, i64 } } } %t2, ptr %slot___try_3
  %t3 = extractvalue { i1, i32, { i8, { i64, i64 } } } %t2, 0
  br i1 %t3, label %try.ok1, label %try.err2
try.err2:
  %t4 = extractvalue { i1, i32, { i8, { i64, i64 } } } %t2, 2
  %t5 = insertvalue { i1, i32, { i8, { i64, i64 } } } undef, i1 0, 0
  %t6 = insertvalue { i1, i32, { i8, { i64, i64 } } } %t5, { i8, { i64, i64 } } %t4, 2
  ret { i1, i32, { i8, { i64, i64 } } } %t6
try.ok1:
  %t7 = extractvalue { i1, i32, { i8, { i64, i64 } } } %t2, 1
  store i32 %t7, ptr %slot_v_2
  %t8 = load i32, ptr %slot_v_2
  %t9 = insertvalue { i1, i32, { i8, { i64, i64 } } } undef, i1 1, 0
  %t10 = insertvalue { i1, i32, { i8, { i64, i64 } } } %t9, i32 %t8, 1
  ret { i1, i32, { i8, { i64, i64 } } } %t10
}

define void @aila_main__main() {
  %slot_short_1 = alloca { ptr, i64 }, align 8
  %slot_r_2 = alloca { i1, i32, { i8, { i64, i64 } } }, align 8
  %slot_v_3 = alloca i32, align 8
  %slot_e_4 = alloca { i8, { i64, i64 } }, align 8
  %slot_got_5 = alloca i64, align 8
  %slot_need_6 = alloca i64, align 8
  %slot_v_7 = alloca i32, align 8
  %slot_e_8 = alloca { i8, { i64, i64 } }, align 8
  %slot___head_9 = alloca { i1, i32, { i8, { i64, i64 } } }, align 8
  %slot___lit_10 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_11 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_12 = alloca { i8, { i64, i64 } }, align 8
  %slot___lit_13 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___lit_14 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_15 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_16 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_17 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_18 = alloca { i1, i32, { i8, { i64, i64 } } }, align 8
  %slot___lit_19 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_20 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_21 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_22 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_23 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_24 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_25 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_26 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 4, i64 1)
  %t3 = insertvalue { ptr, i64 } undef, ptr %t2, 0
  %t4 = insertvalue { ptr, i64 } %t3, i64 4, 1
  store { ptr, i64 } %t4, ptr %slot_short_1
  %t5 = load { ptr, i64 }, ptr %slot_short_1
  %t6 = call { i1, i32, { i8, { i64, i64 } } } @aila_main__parse_header({ ptr, i64 } %t5)
  store { i1, i32, { i8, { i64, i64 } } } %t6, ptr %slot_r_2
  %t7 = load { i1, i32, { i8, { i64, i64 } } }, ptr %slot_r_2
  store { i1, i32, { i8, { i64, i64 } } } %t7, ptr %slot___head_9
  %t8 = load { i1, i32, { i8, { i64, i64 } } }, ptr %slot___head_9
  %t9 = extractvalue { i1, i32, { i8, { i64, i64 } } } %t8, 0
  br i1 %t9, label %m.arm02, label %m.next4
m.next4:
  %t10 = xor i1 %t9, true
  br i1 %t10, label %m.arm13, label %m.next5
m.next5:
  call void @aila_rt_trap(ptr @.str.2, i64 27, ptr @.str.1, i64 44)
  unreachable
m.arm02:
  %t11 = load { i1, i32, { i8, { i64, i64 } } }, ptr %slot___head_9
  %t12 = extractvalue { i1, i32, { i8, { i64, i64 } } } %t11, 1
  store i32 %t12, ptr %slot_v_3
  %t13 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t13, ptr %slot___lit_10
  %t14 = load i32, ptr %slot_v_3
  %t15 = sext i32 %t14 to i64
  %t16 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_10, i32 0, i32 1, i32 0
  store i64 %t15, ptr %t16
  %t17 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_10
  %t18 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_11, i64 0
  store { i8, { { ptr, i64 } } } %t17, ptr %t18
  %t19 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_11, 0
  %t20 = insertvalue { ptr, i64 } %t19, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t20)
  br label %m.end1
m.arm13:
  %t21 = load { i1, i32, { i8, { i64, i64 } } }, ptr %slot___head_9
  %t22 = extractvalue { i1, i32, { i8, { i64, i64 } } } %t21, 2
  store { i8, { i64, i64 } } %t22, ptr %slot_e_4
  %t23 = load { i8, { i64, i64 } }, ptr %slot_e_4
  store { i8, { i64, i64 } } %t23, ptr %slot___head_12
  %t24 = load { i8, { i64, i64 } }, ptr %slot___head_12
  %t25 = extractvalue { i8, { i64, i64 } } %t24, 0
  %t26 = icmp eq i8 %t25, 1
  br i1 %t26, label %m.arm07, label %m.next9
m.next9:
  br label %m.arm18
m.arm07:
  %t27 = getelementptr inbounds { i8, { i64, i64 } }, ptr %slot___head_12, i32 0, i32 1, i32 0
  %t28 = load i64, ptr %t27
  store i64 %t28, ptr %slot_got_5
  %t29 = getelementptr inbounds { i8, { i64, i64 } }, ptr %slot___head_12, i32 0, i32 1, i32 1
  %t30 = load i64, ptr %t29
  store i64 %t30, ptr %slot_need_6
  %t31 = insertvalue { ptr, i64 } undef, ptr @.str.3, 0
  %t32 = insertvalue { ptr, i64 } %t31, i64 45, 1
  %t33 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t33, ptr %slot___lit_13
  %t34 = load i64, ptr %slot_got_5
  %t36 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_13, i32 0, i32 1, i32 0
  store i64 %t34, ptr %t36
  %t37 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_13
  %t38 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t38, ptr %slot___lit_14
  %t39 = load i64, ptr %slot_need_6
  %t41 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_14, i32 0, i32 1, i32 0
  store i64 %t39, ptr %t41
  %t42 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_14
  %t43 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_15, i64 0
  store { i8, { { ptr, i64 } } } %t37, ptr %t43
  %t44 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_15, i64 1
  store { i8, { { ptr, i64 } } } %t42, ptr %t44
  %t45 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_15, 0
  %t46 = insertvalue { ptr, i64 } %t45, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t32, { ptr, i64 } %t46)
  br label %m.end6
m.arm18:
  %t47 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t47, ptr %slot___lit_16
  %t48 = insertvalue { ptr, i64 } undef, ptr @.str.4, 0
  %t49 = insertvalue { ptr, i64 } %t48, i64 12, 1
  %t50 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_16, i32 0, i32 1, i32 0
  store { ptr, i64 } %t49, ptr %t50
  %t51 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_16
  %t52 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_17, i64 0
  store { i8, { { ptr, i64 } } } %t51, ptr %t52
  %t53 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_17, 0
  %t54 = insertvalue { ptr, i64 } %t53, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t54)
  br label %m.end6
m.end6:
  br label %m.end1
m.end1:
  %t55 = load { ptr, i64 }, ptr %slot_short_1
  %t56 = call { i1, i32, { i8, { i64, i64 } } } @aila_main__read_or_default({ ptr, i64 } %t55)
  store { i1, i32, { i8, { i64, i64 } } } %t56, ptr %slot___head_18
  %t57 = load { i1, i32, { i8, { i64, i64 } } }, ptr %slot___head_18
  %t58 = extractvalue { i1, i32, { i8, { i64, i64 } } } %t57, 0
  br i1 %t58, label %m.arm011, label %m.next13
m.next13:
  %t59 = xor i1 %t58, true
  br i1 %t59, label %m.arm112, label %m.next14
m.next14:
  call void @aila_rt_trap(ptr @.str.2, i64 27, ptr @.str.1, i64 58)
  unreachable
m.arm011:
  %t60 = load { i1, i32, { i8, { i64, i64 } } }, ptr %slot___head_18
  %t61 = extractvalue { i1, i32, { i8, { i64, i64 } } } %t60, 1
  store i32 %t61, ptr %slot_v_7
  %t62 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t62, ptr %slot___lit_19
  %t63 = load i32, ptr %slot_v_7
  %t64 = sext i32 %t63 to i64
  %t65 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_19, i32 0, i32 1, i32 0
  store i64 %t64, ptr %t65
  %t66 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_19
  %t67 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_20, i64 0
  store { i8, { { ptr, i64 } } } %t66, ptr %t67
  %t68 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_20, 0
  %t69 = insertvalue { ptr, i64 } %t68, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t69)
  br label %m.end10
m.arm112:
  %t70 = load { i1, i32, { i8, { i64, i64 } } }, ptr %slot___head_18
  %t71 = extractvalue { i1, i32, { i8, { i64, i64 } } } %t70, 2
  store { i8, { i64, i64 } } %t71, ptr %slot_e_8
  %t72 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t72, ptr %slot___lit_21
  %t73 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t74 = insertvalue { ptr, i64 } %t73, i64 31, 1
  %t75 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_21, i32 0, i32 1, i32 0
  store { ptr, i64 } %t74, ptr %t75
  %t76 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_21
  %t77 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_22, i64 0
  store { i8, { { ptr, i64 } } } %t76, ptr %t77
  %t78 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_22, 0
  %t79 = insertvalue { ptr, i64 } %t78, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t79)
  %t80 = load { i8, { i64, i64 } }, ptr %slot_e_8
  br label %m.end10
m.end10:
  %t81 = load { ptr, i64 }, ptr %slot_short_1
  %t82 = call { i1, i32, { i8, { i64, i64 } } } @aila_main__parse_header({ ptr, i64 } %t81)
  %t83 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t83, ptr %slot___lit_23
  %t84 = insertvalue { ptr, i64 } undef, ptr @.str.6, 0
  %t85 = insertvalue { ptr, i64 } %t84, i64 16, 1
  %t86 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_23, i32 0, i32 1, i32 0
  store { ptr, i64 } %t85, ptr %t86
  %t87 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_23
  %t88 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_24, i64 0
  store { i8, { { ptr, i64 } } } %t87, ptr %t88
  %t89 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_24, 0
  %t90 = insertvalue { ptr, i64 } %t89, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t90)
  %t91 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t91, ptr %slot___lit_25
  %t92 = call i32 @aila_main__maybe_scale(i32 14, i32 3)
  %t93 = sext i32 %t92 to i64
  %t94 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_25, i32 0, i32 1, i32 0
  store i64 %t93, ptr %t94
  %t95 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_25
  %t96 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_26, i64 0
  store { i8, { { ptr, i64 } } } %t95, ptr %t96
  %t97 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_26, 0
  %t98 = insertvalue { ptr, i64 } %t97, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t98)
  call void @aila_rt_region_leave(ptr %t1)
  ret void
}

define i32 @main(i32 %argc, ptr %argv) {
entry:
  call void @aila_main__main()
  ret i32 0
}
