; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.1 = private unnamed_addr constant [10 x i8] c"main.aila\00", align 1
@.str.2 = private unnamed_addr constant [17 x i8] c"integer overflow\00", align 1
@.str.3 = private unnamed_addr constant [28 x i8] c"match \E6\9C\AA\E8\A6\86\E7\9B\96\E7\9A\84\E5\88\A4\E5\88\AB\E5\80\BC\00", align 1
@.str.4 = private unnamed_addr constant [33 x i8] c"parse(4 bytes) -> Truncated{got=\00", align 1
@.str.5 = private unnamed_addr constant [8 x i8] c", need=\00", align 1
@.str.6 = private unnamed_addr constant [2 x i8] c"}\00", align 1
@.str.7 = private unnamed_addr constant [13 x i8] c"parse failed\00", align 1
@.str.8 = private unnamed_addr constant [31 x i8] c"safe parse failed, fallback = \00", align 1
@.str.9 = private unnamed_addr constant [17 x i8] c"recovered by ?: \00", align 1

declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare ptr @aila_rt_region_enter(i64)
declare ptr @aila_rt_region_alloc(ptr, i64, i64)
declare void @aila_std_io__print_i32_nl(i32)
declare void @aila_std_io__print_text({ ptr, i64 })
declare void @aila_std_io__print_usize(i64)
declare void @aila_std_io__print_nl()
declare void @aila_rt_region_leave(ptr)

define { i1, i32, { i8, { i64, i64 } } } @aila_main__parse_header({ ptr, i64 } %p_buf) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot___lit_2 = alloca { i8, { i64, i64 } }, align 8
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
  %t5 = call { i32, i1 } @llvm.smul.with.overflow.i32(i32 %t3, i32 %t4)
  %t6 = extractvalue { i32, i1 } %t5, 0
  %t7 = extractvalue { i32, i1 } %t5, 1
  br i1 %t7, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.2, i64 16, ptr @.str.1, i64 29)
  unreachable
chk.ok4:
  store i32 %t6, ptr %slot_scaled_3
  br label %if.end1
if.next3:
  br label %if.end1
if.end1:
  %t8 = load i32, ptr %slot_scaled_3
  ret i32 %t8
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
  %slot___head_10 = alloca { i8, { i64, i64 } }, align 8
  %slot___head_11 = alloca { i1, i32, { i8, { i64, i64 } } }, align 8
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
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 44)
  unreachable
m.arm02:
  %t11 = load { i1, i32, { i8, { i64, i64 } } }, ptr %slot___head_9
  %t12 = extractvalue { i1, i32, { i8, { i64, i64 } } } %t11, 1
  store i32 %t12, ptr %slot_v_3
  %t13 = load i32, ptr %slot_v_3
  call void @aila_std_io__print_i32_nl(i32 %t13)
  br label %m.end1
m.arm13:
  %t14 = load { i1, i32, { i8, { i64, i64 } } }, ptr %slot___head_9
  %t15 = extractvalue { i1, i32, { i8, { i64, i64 } } } %t14, 2
  store { i8, { i64, i64 } } %t15, ptr %slot_e_4
  %t16 = load { i8, { i64, i64 } }, ptr %slot_e_4
  store { i8, { i64, i64 } } %t16, ptr %slot___head_10
  %t17 = load { i8, { i64, i64 } }, ptr %slot___head_10
  %t18 = extractvalue { i8, { i64, i64 } } %t17, 0
  %t19 = icmp eq i8 %t18, 1
  br i1 %t19, label %m.arm07, label %m.next9
m.next9:
  br label %m.arm18
m.arm07:
  %t20 = getelementptr inbounds { i8, { i64, i64 } }, ptr %slot___head_10, i32 0, i32 1, i32 0
  %t21 = load i64, ptr %t20
  store i64 %t21, ptr %slot_got_5
  %t22 = getelementptr inbounds { i8, { i64, i64 } }, ptr %slot___head_10, i32 0, i32 1, i32 1
  %t23 = load i64, ptr %t22
  store i64 %t23, ptr %slot_need_6
  %t24 = insertvalue { ptr, i64 } undef, ptr @.str.4, 0
  %t25 = insertvalue { ptr, i64 } %t24, i64 32, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t25)
  %t26 = load i64, ptr %slot_got_5
  call void @aila_std_io__print_usize(i64 %t26)
  %t27 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t28 = insertvalue { ptr, i64 } %t27, i64 7, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t28)
  %t29 = load i64, ptr %slot_need_6
  call void @aila_std_io__print_usize(i64 %t29)
  %t30 = insertvalue { ptr, i64 } undef, ptr @.str.6, 0
  %t31 = insertvalue { ptr, i64 } %t30, i64 1, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t31)
  call void @aila_std_io__print_nl()
  br label %m.end6
m.arm18:
  %t32 = insertvalue { ptr, i64 } undef, ptr @.str.7, 0
  %t33 = insertvalue { ptr, i64 } %t32, i64 12, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t33)
  br label %m.end6
m.end6:
  br label %m.end1
m.end1:
  %t34 = load { ptr, i64 }, ptr %slot_short_1
  %t35 = call { i1, i32, { i8, { i64, i64 } } } @aila_main__read_or_default({ ptr, i64 } %t34)
  store { i1, i32, { i8, { i64, i64 } } } %t35, ptr %slot___head_11
  %t36 = load { i1, i32, { i8, { i64, i64 } } }, ptr %slot___head_11
  %t37 = extractvalue { i1, i32, { i8, { i64, i64 } } } %t36, 0
  br i1 %t37, label %m.arm011, label %m.next13
m.next13:
  %t38 = xor i1 %t37, true
  br i1 %t38, label %m.arm112, label %m.next14
m.next14:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 61)
  unreachable
m.arm011:
  %t39 = load { i1, i32, { i8, { i64, i64 } } }, ptr %slot___head_11
  %t40 = extractvalue { i1, i32, { i8, { i64, i64 } } } %t39, 1
  store i32 %t40, ptr %slot_v_7
  %t41 = load i32, ptr %slot_v_7
  call void @aila_std_io__print_i32_nl(i32 %t41)
  br label %m.end10
m.arm112:
  %t42 = load { i1, i32, { i8, { i64, i64 } } }, ptr %slot___head_11
  %t43 = extractvalue { i1, i32, { i8, { i64, i64 } } } %t42, 2
  store { i8, { i64, i64 } } %t43, ptr %slot_e_8
  %t44 = insertvalue { ptr, i64 } undef, ptr @.str.8, 0
  %t45 = insertvalue { ptr, i64 } %t44, i64 30, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t45)
  call void @aila_std_io__print_i32_nl(i32 0)
  br label %m.end10
m.end10:
  %t46 = load { ptr, i64 }, ptr %slot_short_1
  %t47 = call { i1, i32, { i8, { i64, i64 } } } @aila_main__parse_header({ ptr, i64 } %t46)
  %t48 = insertvalue { ptr, i64 } undef, ptr @.str.9, 0
  %t49 = insertvalue { ptr, i64 } %t48, i64 16, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t49)
  %t50 = call i32 @aila_main__maybe_scale(i32 14, i32 3)
  call void @aila_std_io__print_i32_nl(i32 %t50)
  call void @aila_rt_region_leave(ptr %t1)
  ret void
}

define i32 @main() {
entry:
  call void @aila_main__main()
  ret i32 0
}
