; AILA module: std.io
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.1 = private unnamed_addr constant [12 x i8] c"std/io.aila\00", align 1
@.str.2 = private unnamed_addr constant [17 x i8] c"division by zero\00", align 1
@.str.3 = private unnamed_addr constant [17 x i8] c"integer overflow\00", align 1
@.str.4 = private unnamed_addr constant [5 x i8] c"true\00", align 1
@.str.5 = private unnamed_addr constant [6 x i8] c"false\00", align 1
@.str.6 = private unnamed_addr constant [25 x i8] c"shift count out of range\00", align 1

declare ptr @aila_rt_region_enter(i64)
declare ptr @aila_rt_region_alloc(ptr, i64, i64)
declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare void @aila_rt_write(ptr, i64)
declare void @aila_rt_region_leave(ptr)
declare void @aila_rt_flush()

define void @aila_std_io__put_byte(i8 %p_b) {
  %slot_b_1 = alloca i8, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  store i8 %p_b, ptr %slot_b_1
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 1, i64 1)
  %t3 = insertvalue { ptr, i64 } undef, ptr %t2, 0
  %t4 = insertvalue { ptr, i64 } %t3, i64 1, 1
  store { ptr, i64 } %t4, ptr %slot_buf_2
  %t5 = load { ptr, i64 }, ptr %slot_buf_2
  %t6 = extractvalue { ptr, i64 } %t5, 0
  %t7 = extractvalue { ptr, i64 } %t5, 1
  %t8 = icmp uge i64 0, %t7
  br i1 %t8, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 47)
  unreachable
chk.ok1:
  %t9 = getelementptr inbounds i8, ptr %t6, i64 0
  %t10 = load i8, ptr %slot_b_1
  store i8 %t10, ptr %t9
  %t11 = load { ptr, i64 }, ptr %slot_buf_2
  %t12 = extractvalue { ptr, i64 } %t11, 0
  call void @aila_rt_write(ptr %t12, i64 1)
  call void @aila_rt_region_leave(ptr %t1)
  ret void
}

define i64 @aila_std_io__fill_u64(i64 %p_v, { ptr, i64 } %p_buf) {
  %slot_v_1 = alloca i64, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_x_3 = alloca i64, align 8
  %slot_n_4 = alloca i64, align 8
  %slot_y_5 = alloca i64, align 8
  %slot_i_6 = alloca i64, align 8
  %slot_d_7 = alloca i64, align 8
  store i64 %p_v, ptr %slot_v_1
  store { ptr, i64 } %p_buf, ptr %slot_buf_2
  %t1 = load i64, ptr %slot_v_1
  store i64 %t1, ptr %slot_x_3
  store i64 1, ptr %slot_n_4
  br label %for.cond2
for.cond2:
  %t2 = load i64, ptr %slot_x_3
  %t3 = icmp ult i64 %t2, 10
  %t4 = xor i1 %t3, true
  br i1 %t4, label %for.body4, label %for.end1
for.body4:
  %t5 = load i64, ptr %slot_x_3
  %t6 = icmp eq i64 10, 0
  br i1 %t6, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.2, i64 16, ptr @.str.1, i64 57)
  unreachable
chk.ok5:
  %t7 = udiv i64 %t5, 10
  store i64 %t7, ptr %slot_x_3
  %t8 = load i64, ptr %slot_n_4
  %t9 = call { i64, i1 } @llvm.uadd.with.overflow.i64(i64 %t8, i64 1)
  %t10 = extractvalue { i64, i1 } %t9, 0
  %t11 = extractvalue { i64, i1 } %t9, 1
  br i1 %t11, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 58)
  unreachable
chk.ok7:
  store i64 %t10, ptr %slot_n_4
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t12 = load i64, ptr %slot_v_1
  store i64 %t12, ptr %slot_y_5
  %t13 = load i64, ptr %slot_n_4
  store i64 %t13, ptr %slot_i_6
  br label %for.cond10
for.cond10:
  %t14 = load i64, ptr %slot_i_6
  %t15 = icmp eq i64 %t14, 0
  %t16 = xor i1 %t15, true
  br i1 %t16, label %for.body12, label %for.end9
for.body12:
  %t17 = load i64, ptr %slot_i_6
  %t18 = call { i64, i1 } @llvm.usub.with.overflow.i64(i64 %t17, i64 1)
  %t19 = extractvalue { i64, i1 } %t18, 0
  %t20 = extractvalue { i64, i1 } %t18, 1
  br i1 %t20, label %chk.bad14, label %chk.ok13
chk.bad14:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 64)
  unreachable
chk.ok13:
  store i64 %t19, ptr %slot_i_6
  %t21 = load i64, ptr %slot_y_5
  %t22 = icmp eq i64 10, 0
  br i1 %t22, label %chk.bad16, label %chk.ok15
chk.bad16:
  call void @aila_rt_trap(ptr @.str.2, i64 16, ptr @.str.1, i64 65)
  unreachable
chk.ok15:
  %t23 = urem i64 %t21, 10
  store i64 %t23, ptr %slot_d_7
  %t24 = load i64, ptr %slot_i_6
  %t25 = load { ptr, i64 }, ptr %slot_buf_2
  %t26 = extractvalue { ptr, i64 } %t25, 0
  %t27 = extractvalue { ptr, i64 } %t25, 1
  %t28 = icmp uge i64 %t24, %t27
  br i1 %t28, label %chk.bad18, label %chk.ok17
chk.bad18:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 66)
  unreachable
chk.ok17:
  %t29 = getelementptr inbounds i8, ptr %t26, i64 %t24
  %t30 = load i64, ptr %slot_d_7
  %t31 = trunc i64 %t30 to i32
  %t32 = call { i32, i1 } @llvm.uadd.with.overflow.i32(i32 %t31, i32 48)
  %t33 = extractvalue { i32, i1 } %t32, 0
  %t34 = extractvalue { i32, i1 } %t32, 1
  br i1 %t34, label %chk.bad20, label %chk.ok19
chk.bad20:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 66)
  unreachable
chk.ok19:
  %t35 = trunc i32 %t33 to i8
  store i8 %t35, ptr %t29
  %t36 = load i64, ptr %slot_y_5
  %t37 = icmp eq i64 10, 0
  br i1 %t37, label %chk.bad22, label %chk.ok21
chk.bad22:
  call void @aila_rt_trap(ptr @.str.2, i64 16, ptr @.str.1, i64 67)
  unreachable
chk.ok21:
  %t38 = udiv i64 %t36, 10
  store i64 %t38, ptr %slot_y_5
  br label %for.step11
for.step11:
  br label %for.cond10
for.end9:
  %t39 = load i64, ptr %slot_n_4
  ret i64 %t39
}

define i64 @aila_std_io__fill_i64(i64 %p_v, { ptr, i64 } %p_buf) {
  %slot_v_1 = alloca i64, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_mag_3 = alloca i64, align 8
  %slot_n_4 = alloca i64, align 8
  %slot_k_5 = alloca i64, align 8
  store i64 %p_v, ptr %slot_v_1
  store { ptr, i64 } %p_buf, ptr %slot_buf_2
  %t1 = load i64, ptr %slot_v_1
  %t2 = icmp slt i64 %t1, 0
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  %t3 = load i64, ptr %slot_v_1
  %t4 = call { i64, i1 } @llvm.ssub.with.overflow.i64(i64 0, i64 %t3)
  %t5 = extractvalue { i64, i1 } %t4, 0
  %t6 = extractvalue { i64, i1 } %t4, 1
  br i1 %t6, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 75)
  unreachable
chk.ok4:
  store i64 %t5, ptr %slot_mag_3
  %t8 = load i64, ptr %slot_mag_3
  %t9 = load { ptr, i64 }, ptr %slot_buf_2
  %t10 = call i64 @aila_std_io__fill_u64(i64 %t8, { ptr, i64 } %t9)
  %t11 = call { i64, i1 } @llvm.uadd.with.overflow.i64(i64 %t10, i64 1)
  %t12 = extractvalue { i64, i1 } %t11, 0
  %t13 = extractvalue { i64, i1 } %t11, 1
  br i1 %t13, label %chk.bad7, label %chk.ok6
chk.bad7:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 76)
  unreachable
chk.ok6:
  store i64 %t12, ptr %slot_n_4
  %t14 = load i64, ptr %slot_n_4
  store i64 %t14, ptr %slot_k_5
  br label %for.cond9
for.cond9:
  %t15 = load i64, ptr %slot_k_5
  %t16 = icmp eq i64 %t15, 1
  %t17 = xor i1 %t16, true
  br i1 %t17, label %for.body11, label %for.end8
for.body11:
  %t18 = load i64, ptr %slot_k_5
  %t19 = call { i64, i1 } @llvm.usub.with.overflow.i64(i64 %t18, i64 1)
  %t20 = extractvalue { i64, i1 } %t19, 0
  %t21 = extractvalue { i64, i1 } %t19, 1
  br i1 %t21, label %chk.bad13, label %chk.ok12
chk.bad13:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 79)
  unreachable
chk.ok12:
  store i64 %t20, ptr %slot_k_5
  %t22 = load i64, ptr %slot_k_5
  %t23 = load { ptr, i64 }, ptr %slot_buf_2
  %t24 = extractvalue { ptr, i64 } %t23, 0
  %t25 = extractvalue { ptr, i64 } %t23, 1
  %t26 = icmp uge i64 %t22, %t25
  br i1 %t26, label %chk.bad15, label %chk.ok14
chk.bad15:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 80)
  unreachable
chk.ok14:
  %t27 = getelementptr inbounds i8, ptr %t24, i64 %t22
  %t28 = load i64, ptr %slot_k_5
  %t29 = call { i64, i1 } @llvm.usub.with.overflow.i64(i64 %t28, i64 1)
  %t30 = extractvalue { i64, i1 } %t29, 0
  %t31 = extractvalue { i64, i1 } %t29, 1
  br i1 %t31, label %chk.bad17, label %chk.ok16
chk.bad17:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 80)
  unreachable
chk.ok16:
  %t32 = load { ptr, i64 }, ptr %slot_buf_2
  %t33 = extractvalue { ptr, i64 } %t32, 0
  %t34 = extractvalue { ptr, i64 } %t32, 1
  %t35 = icmp uge i64 %t30, %t34
  br i1 %t35, label %chk.bad19, label %chk.ok18
chk.bad19:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 80)
  unreachable
chk.ok18:
  %t36 = getelementptr inbounds i8, ptr %t33, i64 %t30
  %t37 = load i8, ptr %t36
  store i8 %t37, ptr %t27
  br label %for.step10
for.step10:
  br label %for.cond9
for.end8:
  %t38 = load { ptr, i64 }, ptr %slot_buf_2
  %t39 = extractvalue { ptr, i64 } %t38, 0
  %t40 = extractvalue { ptr, i64 } %t38, 1
  %t41 = icmp uge i64 0, %t40
  br i1 %t41, label %chk.bad21, label %chk.ok20
chk.bad21:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 82)
  unreachable
chk.ok20:
  %t42 = getelementptr inbounds i8, ptr %t39, i64 0
  store i8 45, ptr %t42
  %t43 = load i64, ptr %slot_n_4
  ret i64 %t43
if.next3:
  br label %if.end1
if.end1:
  %t44 = load i64, ptr %slot_v_1
  %t46 = load { ptr, i64 }, ptr %slot_buf_2
  %t47 = call i64 @aila_std_io__fill_u64(i64 %t44, { ptr, i64 } %t46)
  ret i64 %t47
}

define void @aila_std_io__print_u32(i32 %p_v) {
  %slot_v_1 = alloca i32, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_n_3 = alloca i64, align 8
  store i32 %p_v, ptr %slot_v_1
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 24, i64 1)
  %t3 = insertvalue { ptr, i64 } undef, ptr %t2, 0
  %t4 = insertvalue { ptr, i64 } %t3, i64 24, 1
  store { ptr, i64 } %t4, ptr %slot_buf_2
  %t5 = load i32, ptr %slot_v_1
  %t6 = zext i32 %t5 to i64
  %t7 = load { ptr, i64 }, ptr %slot_buf_2
  %t8 = call i64 @aila_std_io__fill_u64(i64 %t6, { ptr, i64 } %t7)
  store i64 %t8, ptr %slot_n_3
  %t9 = load { ptr, i64 }, ptr %slot_buf_2
  %t10 = extractvalue { ptr, i64 } %t9, 0
  %t11 = load i64, ptr %slot_n_3
  call void @aila_rt_write(ptr %t10, i64 %t11)
  call void @aila_rt_region_leave(ptr %t1)
  ret void
}

define void @aila_std_io__print_i32(i32 %p_v) {
  %slot_v_1 = alloca i32, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_n_3 = alloca i64, align 8
  store i32 %p_v, ptr %slot_v_1
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 24, i64 1)
  %t3 = insertvalue { ptr, i64 } undef, ptr %t2, 0
  %t4 = insertvalue { ptr, i64 } %t3, i64 24, 1
  store { ptr, i64 } %t4, ptr %slot_buf_2
  %t5 = load i32, ptr %slot_v_1
  %t6 = sext i32 %t5 to i64
  %t7 = load { ptr, i64 }, ptr %slot_buf_2
  %t8 = call i64 @aila_std_io__fill_i64(i64 %t6, { ptr, i64 } %t7)
  store i64 %t8, ptr %slot_n_3
  %t9 = load { ptr, i64 }, ptr %slot_buf_2
  %t10 = extractvalue { ptr, i64 } %t9, 0
  %t11 = load i64, ptr %slot_n_3
  call void @aila_rt_write(ptr %t10, i64 %t11)
  call void @aila_rt_region_leave(ptr %t1)
  ret void
}

define void @aila_std_io__print_usize(i64 %p_v) {
  %slot_v_1 = alloca i64, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_n_3 = alloca i64, align 8
  store i64 %p_v, ptr %slot_v_1
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 24, i64 1)
  %t3 = insertvalue { ptr, i64 } undef, ptr %t2, 0
  %t4 = insertvalue { ptr, i64 } %t3, i64 24, 1
  store { ptr, i64 } %t4, ptr %slot_buf_2
  %t5 = load i64, ptr %slot_v_1
  %t7 = load { ptr, i64 }, ptr %slot_buf_2
  %t8 = call i64 @aila_std_io__fill_u64(i64 %t5, { ptr, i64 } %t7)
  store i64 %t8, ptr %slot_n_3
  %t9 = load { ptr, i64 }, ptr %slot_buf_2
  %t10 = extractvalue { ptr, i64 } %t9, 0
  %t11 = load i64, ptr %slot_n_3
  call void @aila_rt_write(ptr %t10, i64 %t11)
  call void @aila_rt_region_leave(ptr %t1)
  ret void
}

define void @aila_std_io__print_f64(double %p_v) {
  %slot_v_1 = alloca double, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_w_3 = alloca double, align 8
  %slot_ip_4 = alloca i64, align 8
  %slot_pos_5 = alloca i64, align 8
  %slot_frac_6 = alloca double, align 8
  %slot_digits_7 = alloca i64, align 8
  %slot_f_8 = alloca double, align 8
  %slot_d_9 = alloca i32, align 8
  %slot_last_10 = alloca i64, align 8
  store double %p_v, ptr %slot_v_1
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 40, i64 1)
  %t3 = insertvalue { ptr, i64 } undef, ptr %t2, 0
  %t4 = insertvalue { ptr, i64 } %t3, i64 40, 1
  store { ptr, i64 } %t4, ptr %slot_buf_2
  %t5 = load double, ptr %slot_v_1
  store double %t5, ptr %slot_w_3
  %t6 = load double, ptr %slot_w_3
  %t7 = fcmp olt double %t6, 0.0
  br i1 %t7, label %if.then2, label %if.next3
if.then2:
  call void @aila_std_io__put_byte(i8 45)
  %t8 = load double, ptr %slot_w_3
  %t9 = fsub double 0.0, %t8
  store double %t9, ptr %slot_w_3
  br label %if.end1
if.next3:
  br label %if.end1
if.end1:
  %t10 = load double, ptr %slot_w_3
  %t11 = fptosi double %t10 to i64
  store i64 %t11, ptr %slot_ip_4
  %t12 = load i64, ptr %slot_ip_4
  %t14 = load { ptr, i64 }, ptr %slot_buf_2
  %t15 = call i64 @aila_std_io__fill_u64(i64 %t12, { ptr, i64 } %t14)
  store i64 %t15, ptr %slot_pos_5
  %t16 = load double, ptr %slot_w_3
  %t17 = load i64, ptr %slot_ip_4
  %t18 = sitofp i64 %t17 to double
  %t19 = fsub double %t16, %t18
  store double %t19, ptr %slot_frac_6
  %t20 = load double, ptr %slot_frac_6
  %t21 = fcmp ogt double %t20, 0.0
  br i1 %t21, label %if.then5, label %if.next6
if.then5:
  store i64 0, ptr %slot_digits_7
  %t22 = load double, ptr %slot_frac_6
  store double %t22, ptr %slot_f_8
  br label %for.cond8
for.cond8:
  %t23 = load i64, ptr %slot_digits_7
  %t24 = icmp ult i64 %t23, 6
  br i1 %t24, label %for.body10, label %for.end7
for.body10:
  %t25 = load double, ptr %slot_f_8
  %t26 = fmul double %t25, 10.0
  store double %t26, ptr %slot_f_8
  %t27 = load double, ptr %slot_f_8
  %t28 = fptosi double %t27 to i32
  store i32 %t28, ptr %slot_d_9
  %t29 = load i64, ptr %slot_pos_5
  %t30 = call { i64, i1 } @llvm.uadd.with.overflow.i64(i64 %t29, i64 1)
  %t31 = extractvalue { i64, i1 } %t30, 0
  %t32 = extractvalue { i64, i1 } %t30, 1
  br i1 %t32, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 130)
  unreachable
chk.ok11:
  %t33 = load i64, ptr %slot_digits_7
  %t34 = call { i64, i1 } @llvm.uadd.with.overflow.i64(i64 %t31, i64 %t33)
  %t35 = extractvalue { i64, i1 } %t34, 0
  %t36 = extractvalue { i64, i1 } %t34, 1
  br i1 %t36, label %chk.bad14, label %chk.ok13
chk.bad14:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 130)
  unreachable
chk.ok13:
  %t37 = load { ptr, i64 }, ptr %slot_buf_2
  %t38 = extractvalue { ptr, i64 } %t37, 0
  %t39 = extractvalue { ptr, i64 } %t37, 1
  %t40 = icmp uge i64 %t35, %t39
  br i1 %t40, label %chk.bad16, label %chk.ok15
chk.bad16:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 130)
  unreachable
chk.ok15:
  %t41 = getelementptr inbounds i8, ptr %t38, i64 %t35
  %t42 = load i32, ptr %slot_d_9
  %t43 = call { i32, i1 } @llvm.uadd.with.overflow.i32(i32 %t42, i32 48)
  %t44 = extractvalue { i32, i1 } %t43, 0
  %t45 = extractvalue { i32, i1 } %t43, 1
  br i1 %t45, label %chk.bad18, label %chk.ok17
chk.bad18:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 130)
  unreachable
chk.ok17:
  %t46 = trunc i32 %t44 to i8
  store i8 %t46, ptr %t41
  %t47 = load i64, ptr %slot_digits_7
  %t48 = call { i64, i1 } @llvm.uadd.with.overflow.i64(i64 %t47, i64 1)
  %t49 = extractvalue { i64, i1 } %t48, 0
  %t50 = extractvalue { i64, i1 } %t48, 1
  br i1 %t50, label %chk.bad20, label %chk.ok19
chk.bad20:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 131)
  unreachable
chk.ok19:
  store i64 %t49, ptr %slot_digits_7
  %t51 = load double, ptr %slot_f_8
  %t52 = load i32, ptr %slot_d_9
  %t53 = uitofp i32 %t52 to double
  %t54 = fsub double %t51, %t53
  store double %t54, ptr %slot_f_8
  br label %for.step9
for.step9:
  br label %for.cond8
for.end7:
  %t55 = load i64, ptr %slot_digits_7
  store i64 %t55, ptr %slot_last_10
  br label %for.cond22
for.cond22:
  %t56 = load i64, ptr %slot_last_10
  %t57 = icmp ugt i64 %t56, 0
  br i1 %t57, label %for.body24, label %for.end21
for.body24:
  %t58 = load i64, ptr %slot_pos_5
  %t59 = load i64, ptr %slot_last_10
  %t60 = call { i64, i1 } @llvm.uadd.with.overflow.i64(i64 %t58, i64 %t59)
  %t61 = extractvalue { i64, i1 } %t60, 0
  %t62 = extractvalue { i64, i1 } %t60, 1
  br i1 %t62, label %chk.bad29, label %chk.ok28
chk.bad29:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 136)
  unreachable
chk.ok28:
  %t63 = load { ptr, i64 }, ptr %slot_buf_2
  %t64 = extractvalue { ptr, i64 } %t63, 0
  %t65 = extractvalue { ptr, i64 } %t63, 1
  %t66 = icmp uge i64 %t61, %t65
  br i1 %t66, label %chk.bad31, label %chk.ok30
chk.bad31:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 136)
  unreachable
chk.ok30:
  %t67 = getelementptr inbounds i8, ptr %t64, i64 %t61
  %t68 = load i8, ptr %t67
  %t69 = icmp ne i8 %t68, 48
  br i1 %t69, label %if.then26, label %if.next27
if.then26:
  br label %for.end21
if.next27:
  br label %if.end25
if.end25:
  %t70 = load i64, ptr %slot_last_10
  %t71 = call { i64, i1 } @llvm.usub.with.overflow.i64(i64 %t70, i64 1)
  %t72 = extractvalue { i64, i1 } %t71, 0
  %t73 = extractvalue { i64, i1 } %t71, 1
  br i1 %t73, label %chk.bad33, label %chk.ok32
chk.bad33:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 139)
  unreachable
chk.ok32:
  store i64 %t72, ptr %slot_last_10
  br label %for.step23
for.step23:
  br label %for.cond22
for.end21:
  %t74 = load i64, ptr %slot_last_10
  %t75 = icmp ugt i64 %t74, 0
  br i1 %t75, label %if.then35, label %if.next36
if.then35:
  %t76 = load i64, ptr %slot_pos_5
  %t77 = load { ptr, i64 }, ptr %slot_buf_2
  %t78 = extractvalue { ptr, i64 } %t77, 0
  %t79 = extractvalue { ptr, i64 } %t77, 1
  %t80 = icmp uge i64 %t76, %t79
  br i1 %t80, label %chk.bad38, label %chk.ok37
chk.bad38:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 142)
  unreachable
chk.ok37:
  %t81 = getelementptr inbounds i8, ptr %t78, i64 %t76
  store i8 46, ptr %t81
  %t82 = load { ptr, i64 }, ptr %slot_buf_2
  %t83 = extractvalue { ptr, i64 } %t82, 0
  %t84 = load i64, ptr %slot_pos_5
  %t85 = call { i64, i1 } @llvm.uadd.with.overflow.i64(i64 %t84, i64 1)
  %t86 = extractvalue { i64, i1 } %t85, 0
  %t87 = extractvalue { i64, i1 } %t85, 1
  br i1 %t87, label %chk.bad40, label %chk.ok39
chk.bad40:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 143)
  unreachable
chk.ok39:
  %t88 = load i64, ptr %slot_last_10
  %t89 = call { i64, i1 } @llvm.uadd.with.overflow.i64(i64 %t86, i64 %t88)
  %t90 = extractvalue { i64, i1 } %t89, 0
  %t91 = extractvalue { i64, i1 } %t89, 1
  br i1 %t91, label %chk.bad42, label %chk.ok41
chk.bad42:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 143)
  unreachable
chk.ok41:
  call void @aila_rt_write(ptr %t83, i64 %t90)
  ret void
if.next36:
  br label %if.end34
if.end34:
  br label %if.end4
if.next6:
  br label %if.end4
if.end4:
  %t92 = load { ptr, i64 }, ptr %slot_buf_2
  %t93 = extractvalue { ptr, i64 } %t92, 0
  %t94 = load i64, ptr %slot_pos_5
  call void @aila_rt_write(ptr %t93, i64 %t94)
  call void @aila_rt_region_leave(ptr %t1)
  ret void
}

define void @aila_std_io__print_f32(float %p_v) {
  %slot_v_1 = alloca float, align 8
  store float %p_v, ptr %slot_v_1
  %t1 = load float, ptr %slot_v_1
  %t2 = fpext float %t1 to double
  call void @aila_std_io__print_f64(double %t2)
  ret void
}

define void @aila_std_io__print_bool(i1 %p_v) {
  %slot_v_1 = alloca i1, align 8
  store i1 %p_v, ptr %slot_v_1
  %t1 = load i1, ptr %slot_v_1
  br i1 %t1, label %if.then2, label %if.next3
if.then2:
  %t2 = insertvalue { ptr, i64 } undef, ptr @.str.4, 0
  %t3 = insertvalue { ptr, i64 } %t2, i64 4, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t3)
  br label %if.end1
if.next3:
  %t4 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t5 = insertvalue { ptr, i64 } %t4, i64 5, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t5)
  br label %if.end1
if.end1:
  ret void
}

define void @aila_std_io__print_char(i32 %p_c) {
  %slot_c_1 = alloca i32, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_cp_3 = alloca i32, align 8
  store i32 %p_c, ptr %slot_c_1
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 4, i64 1)
  %t3 = insertvalue { ptr, i64 } undef, ptr %t2, 0
  %t4 = insertvalue { ptr, i64 } %t3, i64 4, 1
  store { ptr, i64 } %t4, ptr %slot_buf_2
  %t5 = load i32, ptr %slot_c_1
  store i32 %t5, ptr %slot_cp_3
  %t7 = load i32, ptr %slot_cp_3
  %t8 = icmp ult i32 %t7, 128
  br i1 %t8, label %if.then2, label %if.next3
if.then2:
  %t9 = load { ptr, i64 }, ptr %slot_buf_2
  %t10 = extractvalue { ptr, i64 } %t9, 0
  %t11 = extractvalue { ptr, i64 } %t9, 1
  %t12 = icmp uge i64 0, %t11
  br i1 %t12, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 169)
  unreachable
chk.ok4:
  %t13 = getelementptr inbounds i8, ptr %t10, i64 0
  %t14 = load i32, ptr %slot_cp_3
  %t15 = trunc i32 %t14 to i8
  store i8 %t15, ptr %t13
  %t16 = load { ptr, i64 }, ptr %slot_buf_2
  %t17 = extractvalue { ptr, i64 } %t16, 0
  call void @aila_rt_write(ptr %t17, i64 1)
  ret void
if.next3:
  br label %if.end1
if.end1:
  %t18 = load i32, ptr %slot_cp_3
  %t19 = icmp ult i32 %t18, 2048
  br i1 %t19, label %if.then7, label %if.next8
if.then7:
  %t20 = load { ptr, i64 }, ptr %slot_buf_2
  %t21 = extractvalue { ptr, i64 } %t20, 0
  %t22 = extractvalue { ptr, i64 } %t20, 1
  %t23 = icmp uge i64 0, %t22
  br i1 %t23, label %chk.bad10, label %chk.ok9
chk.bad10:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 174)
  unreachable
chk.ok9:
  %t24 = getelementptr inbounds i8, ptr %t21, i64 0
  %t25 = load i32, ptr %slot_cp_3
  %t26 = icmp uge i32 6, 32
  br i1 %t26, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.6, i64 24, ptr @.str.1, i64 174)
  unreachable
chk.ok11:
  %t27 = lshr i32 %t25, 6
  %t28 = call { i32, i1 } @llvm.uadd.with.overflow.i32(i32 192, i32 %t27)
  %t29 = extractvalue { i32, i1 } %t28, 0
  %t30 = extractvalue { i32, i1 } %t28, 1
  br i1 %t30, label %chk.bad14, label %chk.ok13
chk.bad14:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 174)
  unreachable
chk.ok13:
  %t31 = trunc i32 %t29 to i8
  store i8 %t31, ptr %t24
  %t32 = load { ptr, i64 }, ptr %slot_buf_2
  %t33 = extractvalue { ptr, i64 } %t32, 0
  %t34 = extractvalue { ptr, i64 } %t32, 1
  %t35 = icmp uge i64 1, %t34
  br i1 %t35, label %chk.bad16, label %chk.ok15
chk.bad16:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 175)
  unreachable
chk.ok15:
  %t36 = getelementptr inbounds i8, ptr %t33, i64 1
  %t37 = load i32, ptr %slot_cp_3
  %t38 = and i32 %t37, 63
  %t39 = call { i32, i1 } @llvm.uadd.with.overflow.i32(i32 128, i32 %t38)
  %t40 = extractvalue { i32, i1 } %t39, 0
  %t41 = extractvalue { i32, i1 } %t39, 1
  br i1 %t41, label %chk.bad18, label %chk.ok17
chk.bad18:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 175)
  unreachable
chk.ok17:
  %t42 = trunc i32 %t40 to i8
  store i8 %t42, ptr %t36
  %t43 = load { ptr, i64 }, ptr %slot_buf_2
  %t44 = extractvalue { ptr, i64 } %t43, 0
  call void @aila_rt_write(ptr %t44, i64 2)
  ret void
if.next8:
  br label %if.end6
if.end6:
  %t45 = load i32, ptr %slot_cp_3
  %t46 = icmp ult i32 %t45, 65536
  br i1 %t46, label %if.then20, label %if.next21
if.then20:
  %t47 = load { ptr, i64 }, ptr %slot_buf_2
  %t48 = extractvalue { ptr, i64 } %t47, 0
  %t49 = extractvalue { ptr, i64 } %t47, 1
  %t50 = icmp uge i64 0, %t49
  br i1 %t50, label %chk.bad23, label %chk.ok22
chk.bad23:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 180)
  unreachable
chk.ok22:
  %t51 = getelementptr inbounds i8, ptr %t48, i64 0
  %t52 = load i32, ptr %slot_cp_3
  %t53 = icmp uge i32 12, 32
  br i1 %t53, label %chk.bad25, label %chk.ok24
chk.bad25:
  call void @aila_rt_trap(ptr @.str.6, i64 24, ptr @.str.1, i64 180)
  unreachable
chk.ok24:
  %t54 = lshr i32 %t52, 12
  %t55 = call { i32, i1 } @llvm.uadd.with.overflow.i32(i32 224, i32 %t54)
  %t56 = extractvalue { i32, i1 } %t55, 0
  %t57 = extractvalue { i32, i1 } %t55, 1
  br i1 %t57, label %chk.bad27, label %chk.ok26
chk.bad27:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 180)
  unreachable
chk.ok26:
  %t58 = trunc i32 %t56 to i8
  store i8 %t58, ptr %t51
  %t59 = load { ptr, i64 }, ptr %slot_buf_2
  %t60 = extractvalue { ptr, i64 } %t59, 0
  %t61 = extractvalue { ptr, i64 } %t59, 1
  %t62 = icmp uge i64 1, %t61
  br i1 %t62, label %chk.bad29, label %chk.ok28
chk.bad29:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 181)
  unreachable
chk.ok28:
  %t63 = getelementptr inbounds i8, ptr %t60, i64 1
  %t64 = load i32, ptr %slot_cp_3
  %t65 = icmp uge i32 6, 32
  br i1 %t65, label %chk.bad31, label %chk.ok30
chk.bad31:
  call void @aila_rt_trap(ptr @.str.6, i64 24, ptr @.str.1, i64 181)
  unreachable
chk.ok30:
  %t66 = lshr i32 %t64, 6
  %t67 = and i32 %t66, 63
  %t68 = call { i32, i1 } @llvm.uadd.with.overflow.i32(i32 128, i32 %t67)
  %t69 = extractvalue { i32, i1 } %t68, 0
  %t70 = extractvalue { i32, i1 } %t68, 1
  br i1 %t70, label %chk.bad33, label %chk.ok32
chk.bad33:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 181)
  unreachable
chk.ok32:
  %t71 = trunc i32 %t69 to i8
  store i8 %t71, ptr %t63
  %t72 = load { ptr, i64 }, ptr %slot_buf_2
  %t73 = extractvalue { ptr, i64 } %t72, 0
  %t74 = extractvalue { ptr, i64 } %t72, 1
  %t75 = icmp uge i64 2, %t74
  br i1 %t75, label %chk.bad35, label %chk.ok34
chk.bad35:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 182)
  unreachable
chk.ok34:
  %t76 = getelementptr inbounds i8, ptr %t73, i64 2
  %t77 = load i32, ptr %slot_cp_3
  %t78 = and i32 %t77, 63
  %t79 = call { i32, i1 } @llvm.uadd.with.overflow.i32(i32 128, i32 %t78)
  %t80 = extractvalue { i32, i1 } %t79, 0
  %t81 = extractvalue { i32, i1 } %t79, 1
  br i1 %t81, label %chk.bad37, label %chk.ok36
chk.bad37:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 182)
  unreachable
chk.ok36:
  %t82 = trunc i32 %t80 to i8
  store i8 %t82, ptr %t76
  %t83 = load { ptr, i64 }, ptr %slot_buf_2
  %t84 = extractvalue { ptr, i64 } %t83, 0
  call void @aila_rt_write(ptr %t84, i64 3)
  ret void
if.next21:
  br label %if.end19
if.end19:
  %t85 = load { ptr, i64 }, ptr %slot_buf_2
  %t86 = extractvalue { ptr, i64 } %t85, 0
  %t87 = extractvalue { ptr, i64 } %t85, 1
  %t88 = icmp uge i64 0, %t87
  br i1 %t88, label %chk.bad39, label %chk.ok38
chk.bad39:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 186)
  unreachable
chk.ok38:
  %t89 = getelementptr inbounds i8, ptr %t86, i64 0
  %t90 = load i32, ptr %slot_cp_3
  %t91 = icmp uge i32 18, 32
  br i1 %t91, label %chk.bad41, label %chk.ok40
chk.bad41:
  call void @aila_rt_trap(ptr @.str.6, i64 24, ptr @.str.1, i64 186)
  unreachable
chk.ok40:
  %t92 = lshr i32 %t90, 18
  %t93 = call { i32, i1 } @llvm.uadd.with.overflow.i32(i32 240, i32 %t92)
  %t94 = extractvalue { i32, i1 } %t93, 0
  %t95 = extractvalue { i32, i1 } %t93, 1
  br i1 %t95, label %chk.bad43, label %chk.ok42
chk.bad43:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 186)
  unreachable
chk.ok42:
  %t96 = trunc i32 %t94 to i8
  store i8 %t96, ptr %t89
  %t97 = load { ptr, i64 }, ptr %slot_buf_2
  %t98 = extractvalue { ptr, i64 } %t97, 0
  %t99 = extractvalue { ptr, i64 } %t97, 1
  %t100 = icmp uge i64 1, %t99
  br i1 %t100, label %chk.bad45, label %chk.ok44
chk.bad45:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 187)
  unreachable
chk.ok44:
  %t101 = getelementptr inbounds i8, ptr %t98, i64 1
  %t102 = load i32, ptr %slot_cp_3
  %t103 = icmp uge i32 12, 32
  br i1 %t103, label %chk.bad47, label %chk.ok46
chk.bad47:
  call void @aila_rt_trap(ptr @.str.6, i64 24, ptr @.str.1, i64 187)
  unreachable
chk.ok46:
  %t104 = lshr i32 %t102, 12
  %t105 = and i32 %t104, 63
  %t106 = call { i32, i1 } @llvm.uadd.with.overflow.i32(i32 128, i32 %t105)
  %t107 = extractvalue { i32, i1 } %t106, 0
  %t108 = extractvalue { i32, i1 } %t106, 1
  br i1 %t108, label %chk.bad49, label %chk.ok48
chk.bad49:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 187)
  unreachable
chk.ok48:
  %t109 = trunc i32 %t107 to i8
  store i8 %t109, ptr %t101
  %t110 = load { ptr, i64 }, ptr %slot_buf_2
  %t111 = extractvalue { ptr, i64 } %t110, 0
  %t112 = extractvalue { ptr, i64 } %t110, 1
  %t113 = icmp uge i64 2, %t112
  br i1 %t113, label %chk.bad51, label %chk.ok50
chk.bad51:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 188)
  unreachable
chk.ok50:
  %t114 = getelementptr inbounds i8, ptr %t111, i64 2
  %t115 = load i32, ptr %slot_cp_3
  %t116 = icmp uge i32 6, 32
  br i1 %t116, label %chk.bad53, label %chk.ok52
chk.bad53:
  call void @aila_rt_trap(ptr @.str.6, i64 24, ptr @.str.1, i64 188)
  unreachable
chk.ok52:
  %t117 = lshr i32 %t115, 6
  %t118 = and i32 %t117, 63
  %t119 = call { i32, i1 } @llvm.uadd.with.overflow.i32(i32 128, i32 %t118)
  %t120 = extractvalue { i32, i1 } %t119, 0
  %t121 = extractvalue { i32, i1 } %t119, 1
  br i1 %t121, label %chk.bad55, label %chk.ok54
chk.bad55:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 188)
  unreachable
chk.ok54:
  %t122 = trunc i32 %t120 to i8
  store i8 %t122, ptr %t114
  %t123 = load { ptr, i64 }, ptr %slot_buf_2
  %t124 = extractvalue { ptr, i64 } %t123, 0
  %t125 = extractvalue { ptr, i64 } %t123, 1
  %t126 = icmp uge i64 3, %t125
  br i1 %t126, label %chk.bad57, label %chk.ok56
chk.bad57:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 189)
  unreachable
chk.ok56:
  %t127 = getelementptr inbounds i8, ptr %t124, i64 3
  %t128 = load i32, ptr %slot_cp_3
  %t129 = and i32 %t128, 63
  %t130 = call { i32, i1 } @llvm.uadd.with.overflow.i32(i32 128, i32 %t129)
  %t131 = extractvalue { i32, i1 } %t130, 0
  %t132 = extractvalue { i32, i1 } %t130, 1
  br i1 %t132, label %chk.bad59, label %chk.ok58
chk.bad59:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 189)
  unreachable
chk.ok58:
  %t133 = trunc i32 %t131 to i8
  store i8 %t133, ptr %t127
  %t134 = load { ptr, i64 }, ptr %slot_buf_2
  %t135 = extractvalue { ptr, i64 } %t134, 0
  call void @aila_rt_write(ptr %t135, i64 4)
  call void @aila_rt_region_leave(ptr %t1)
  ret void
}

define void @aila_std_io__print_nl() {
  call void @aila_std_io__put_byte(i8 10)
  call void @aila_rt_flush()
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
  call void @aila_rt_write(ptr %t3, i64 %t5)
  ret void
}

define void @aila_std_io__print_bytes({ ptr, i64 } %p_b) {
  %slot_b_1 = alloca { ptr, i64 }, align 8
  store { ptr, i64 } %p_b, ptr %slot_b_1
  %t1 = load { ptr, i64 }, ptr %slot_b_1
  %t2 = extractvalue { ptr, i64 } %t1, 0
  %t3 = load { ptr, i64 }, ptr %slot_b_1
  %t4 = extractvalue { ptr, i64 } %t3, 1
  call void @aila_rt_write(ptr %t2, i64 %t4)
  ret void
}

define void @aila_std_io__print_i32_nl(i32 %p_v) {
  %slot_v_1 = alloca i32, align 8
  store i32 %p_v, ptr %slot_v_1
  %t1 = load i32, ptr %slot_v_1
  call void @aila_std_io__print_i32(i32 %t1)
  call void @aila_std_io__print_nl()
  ret void
}

define void @aila_std_io__print_u32_nl(i32 %p_v) {
  %slot_v_1 = alloca i32, align 8
  store i32 %p_v, ptr %slot_v_1
  %t1 = load i32, ptr %slot_v_1
  call void @aila_std_io__print_u32(i32 %t1)
  call void @aila_std_io__print_nl()
  ret void
}
