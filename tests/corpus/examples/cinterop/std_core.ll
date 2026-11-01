; AILA module: std.core
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [17 x i8] c"division by zero\00", align 1
@.str.1 = private unnamed_addr constant [14 x i8] c"std/core.aila\00", align 1
@.str.2 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1

declare void @aila_rt_trap(ptr, i64, ptr, i64)

define i64 @aila_std_core__fill_u64(i64 %p_v, { ptr, i64 } %p_buf) {
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
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 20)
  unreachable
chk.ok5:
  %t7 = udiv i64 %t5, 10
  store i64 %t7, ptr %slot_x_3
  %t8 = load i64, ptr %slot_n_4
  %t9 = add i64 %t8, 1
  store i64 %t9, ptr %slot_n_4
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t10 = load i64, ptr %slot_v_1
  store i64 %t10, ptr %slot_y_5
  %t11 = load i64, ptr %slot_n_4
  store i64 %t11, ptr %slot_i_6
  br label %for.cond8
for.cond8:
  %t12 = load i64, ptr %slot_i_6
  %t13 = icmp eq i64 %t12, 0
  %t14 = xor i1 %t13, true
  br i1 %t14, label %for.body10, label %for.end7
for.body10:
  %t15 = load i64, ptr %slot_i_6
  %t16 = sub i64 %t15, 1
  store i64 %t16, ptr %slot_i_6
  %t17 = load i64, ptr %slot_y_5
  %t18 = icmp eq i64 10, 0
  br i1 %t18, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 28)
  unreachable
chk.ok11:
  %t19 = urem i64 %t17, 10
  store i64 %t19, ptr %slot_d_7
  %t20 = load i64, ptr %slot_i_6
  %t21 = load { ptr, i64 }, ptr %slot_buf_2
  %t22 = extractvalue { ptr, i64 } %t21, 0
  %t23 = extractvalue { ptr, i64 } %t21, 1
  %t24 = icmp uge i64 %t20, %t23
  br i1 %t24, label %chk.bad14, label %chk.ok13
chk.bad14:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 29)
  unreachable
chk.ok13:
  %t25 = getelementptr inbounds i8, ptr %t22, i64 %t20
  %t26 = load i64, ptr %slot_d_7
  %t27 = trunc i64 %t26 to i32
  %t28 = add i32 %t27, 48
  %t29 = trunc i32 %t28 to i8
  store i8 %t29, ptr %t25
  %t30 = load i64, ptr %slot_y_5
  %t31 = icmp eq i64 10, 0
  br i1 %t31, label %chk.bad16, label %chk.ok15
chk.bad16:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 30)
  unreachable
chk.ok15:
  %t32 = udiv i64 %t30, 10
  store i64 %t32, ptr %slot_y_5
  br label %for.step9
for.step9:
  br label %for.cond8
for.end7:
  %t33 = load i64, ptr %slot_n_4
  ret i64 %t33
}

define i64 @aila_std_core__fill_i64(i64 %p_v, { ptr, i64 } %p_buf) {
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
  %t4 = sub i64 0, %t3
  store i64 %t4, ptr %slot_mag_3
  %t6 = load i64, ptr %slot_mag_3
  %t7 = load { ptr, i64 }, ptr %slot_buf_2
  %t8 = call i64 @aila_std_core__fill_u64(i64 %t6, { ptr, i64 } %t7)
  %t9 = add i64 %t8, 1
  store i64 %t9, ptr %slot_n_4
  %t10 = load i64, ptr %slot_n_4
  store i64 %t10, ptr %slot_k_5
  br label %for.cond5
for.cond5:
  %t11 = load i64, ptr %slot_k_5
  %t12 = icmp eq i64 %t11, 1
  %t13 = xor i1 %t12, true
  br i1 %t13, label %for.body7, label %for.end4
for.body7:
  %t14 = load i64, ptr %slot_k_5
  %t15 = sub i64 %t14, 1
  store i64 %t15, ptr %slot_k_5
  %t16 = load i64, ptr %slot_k_5
  %t17 = load { ptr, i64 }, ptr %slot_buf_2
  %t18 = extractvalue { ptr, i64 } %t17, 0
  %t19 = extractvalue { ptr, i64 } %t17, 1
  %t20 = icmp uge i64 %t16, %t19
  br i1 %t20, label %chk.bad9, label %chk.ok8
chk.bad9:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 43)
  unreachable
chk.ok8:
  %t21 = getelementptr inbounds i8, ptr %t18, i64 %t16
  %t22 = load i64, ptr %slot_k_5
  %t23 = sub i64 %t22, 1
  %t24 = load { ptr, i64 }, ptr %slot_buf_2
  %t25 = extractvalue { ptr, i64 } %t24, 0
  %t26 = extractvalue { ptr, i64 } %t24, 1
  %t27 = icmp uge i64 %t23, %t26
  br i1 %t27, label %chk.bad11, label %chk.ok10
chk.bad11:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 43)
  unreachable
chk.ok10:
  %t28 = getelementptr inbounds i8, ptr %t25, i64 %t23
  %t29 = load i8, ptr %t28
  store i8 %t29, ptr %t21
  br label %for.step6
for.step6:
  br label %for.cond5
for.end4:
  %t30 = load { ptr, i64 }, ptr %slot_buf_2
  %t31 = extractvalue { ptr, i64 } %t30, 0
  %t32 = extractvalue { ptr, i64 } %t30, 1
  %t33 = icmp uge i64 0, %t32
  br i1 %t33, label %chk.bad13, label %chk.ok12
chk.bad13:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 45)
  unreachable
chk.ok12:
  %t34 = getelementptr inbounds i8, ptr %t31, i64 0
  store i8 45, ptr %t34
  %t35 = load i64, ptr %slot_n_4
  ret i64 %t35
if.next3:
  br label %if.end1
if.end1:
  %t36 = load i64, ptr %slot_v_1
  %t38 = load { ptr, i64 }, ptr %slot_buf_2
  %t39 = call i64 @aila_std_core__fill_u64(i64 %t36, { ptr, i64 } %t38)
  ret i64 %t39
}

define i64 @aila_std_core__fill_f64(double %p_v, { ptr, i64 } %p_buf) {
  %slot_v_1 = alloca double, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_w_3 = alloca double, align 8
  %slot_pos_4 = alloca i64, align 8
  %slot_ipi_5 = alloca i64, align 8
  %slot_ip_6 = alloca i64, align 8
  %slot_nd_7 = alloca i64, align 8
  %slot_y_8 = alloca i64, align 8
  %slot_i_9 = alloca i64, align 8
  %slot_d_10 = alloca i64, align 8
  %slot_frac_11 = alloca double, align 8
  %slot_digits_12 = alloca i64, align 8
  %slot_f_13 = alloca double, align 8
  %slot_d_14 = alloca i32, align 8
  %slot_last_15 = alloca i64, align 8
  store double %p_v, ptr %slot_v_1
  store { ptr, i64 } %p_buf, ptr %slot_buf_2
  %t1 = load double, ptr %slot_v_1
  store double %t1, ptr %slot_w_3
  store i64 0, ptr %slot_pos_4
  %t2 = load double, ptr %slot_w_3
  %t3 = fcmp olt double %t2, 0.0
  br i1 %t3, label %if.then2, label %if.next3
if.then2:
  %t4 = load { ptr, i64 }, ptr %slot_buf_2
  %t5 = extractvalue { ptr, i64 } %t4, 0
  %t6 = extractvalue { ptr, i64 } %t4, 1
  %t7 = icmp uge i64 0, %t6
  br i1 %t7, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 57)
  unreachable
chk.ok4:
  %t8 = getelementptr inbounds i8, ptr %t5, i64 0
  store i8 45, ptr %t8
  store i64 1, ptr %slot_pos_4
  %t9 = load double, ptr %slot_w_3
  %t10 = fsub double 0.0, %t9
  store double %t10, ptr %slot_w_3
  br label %if.end1
if.next3:
  br label %if.end1
if.end1:
  %t11 = load double, ptr %slot_w_3
  %t12 = fptosi double %t11 to i64
  store i64 %t12, ptr %slot_ipi_5
  %t13 = load i64, ptr %slot_ipi_5
  store i64 %t13, ptr %slot_ip_6
  %t15 = load i64, ptr %slot_ip_6
  %t16 = call i64 @aila_std_core__digits_u64(i64 %t15)
  store i64 %t16, ptr %slot_nd_7
  %t17 = load i64, ptr %slot_ip_6
  store i64 %t17, ptr %slot_y_8
  %t18 = load i64, ptr %slot_pos_4
  %t19 = load i64, ptr %slot_nd_7
  %t20 = add i64 %t18, %t19
  store i64 %t20, ptr %slot_i_9
  br label %for.cond7
for.cond7:
  %t21 = load i64, ptr %slot_i_9
  %t22 = load i64, ptr %slot_pos_4
  %t23 = icmp eq i64 %t21, %t22
  %t24 = xor i1 %t23, true
  br i1 %t24, label %for.body9, label %for.end6
for.body9:
  %t25 = load i64, ptr %slot_i_9
  %t26 = sub i64 %t25, 1
  store i64 %t26, ptr %slot_i_9
  %t27 = load i64, ptr %slot_y_8
  %t28 = icmp eq i64 10, 0
  br i1 %t28, label %chk.bad11, label %chk.ok10
chk.bad11:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 69)
  unreachable
chk.ok10:
  %t29 = urem i64 %t27, 10
  store i64 %t29, ptr %slot_d_10
  %t30 = load i64, ptr %slot_i_9
  %t31 = load { ptr, i64 }, ptr %slot_buf_2
  %t32 = extractvalue { ptr, i64 } %t31, 0
  %t33 = extractvalue { ptr, i64 } %t31, 1
  %t34 = icmp uge i64 %t30, %t33
  br i1 %t34, label %chk.bad13, label %chk.ok12
chk.bad13:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 70)
  unreachable
chk.ok12:
  %t35 = getelementptr inbounds i8, ptr %t32, i64 %t30
  %t36 = load i64, ptr %slot_d_10
  %t37 = trunc i64 %t36 to i32
  %t38 = add i32 %t37, 48
  %t39 = trunc i32 %t38 to i8
  store i8 %t39, ptr %t35
  %t40 = load i64, ptr %slot_y_8
  %t41 = icmp eq i64 10, 0
  br i1 %t41, label %chk.bad15, label %chk.ok14
chk.bad15:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 71)
  unreachable
chk.ok14:
  %t42 = udiv i64 %t40, 10
  store i64 %t42, ptr %slot_y_8
  br label %for.step8
for.step8:
  br label %for.cond7
for.end6:
  %t43 = load double, ptr %slot_w_3
  %t44 = load i64, ptr %slot_ip_6
  %t45 = uitofp i64 %t44 to double
  %t46 = fsub double %t43, %t45
  store double %t46, ptr %slot_frac_11
  %t47 = load double, ptr %slot_frac_11
  %t48 = fcmp ogt double %t47, 0.0
  br i1 %t48, label %if.then17, label %if.next18
if.then17:
  store i64 0, ptr %slot_digits_12
  %t49 = load double, ptr %slot_frac_11
  store double %t49, ptr %slot_f_13
  br label %for.cond20
for.cond20:
  %t50 = load i64, ptr %slot_digits_12
  %t51 = icmp ult i64 %t50, 6
  br i1 %t51, label %for.body22, label %for.end19
for.body22:
  %t52 = load double, ptr %slot_f_13
  %t53 = fmul double %t52, 10.0
  store double %t53, ptr %slot_f_13
  %t54 = load double, ptr %slot_f_13
  %t55 = fptosi double %t54 to i32
  store i32 %t55, ptr %slot_d_14
  %t56 = load i64, ptr %slot_pos_4
  %t57 = load i64, ptr %slot_nd_7
  %t58 = add i64 %t56, %t57
  %t59 = add i64 %t58, 1
  %t60 = load i64, ptr %slot_digits_12
  %t61 = add i64 %t59, %t60
  %t62 = load { ptr, i64 }, ptr %slot_buf_2
  %t63 = extractvalue { ptr, i64 } %t62, 0
  %t64 = extractvalue { ptr, i64 } %t62, 1
  %t65 = icmp uge i64 %t61, %t64
  br i1 %t65, label %chk.bad24, label %chk.ok23
chk.bad24:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 81)
  unreachable
chk.ok23:
  %t66 = getelementptr inbounds i8, ptr %t63, i64 %t61
  %t67 = load i32, ptr %slot_d_14
  %t68 = add i32 %t67, 48
  %t69 = trunc i32 %t68 to i8
  store i8 %t69, ptr %t66
  %t70 = load i64, ptr %slot_digits_12
  %t71 = add i64 %t70, 1
  store i64 %t71, ptr %slot_digits_12
  %t72 = load double, ptr %slot_f_13
  %t73 = load i32, ptr %slot_d_14
  %t74 = uitofp i32 %t73 to double
  %t75 = fsub double %t72, %t74
  store double %t75, ptr %slot_f_13
  br label %for.step21
for.step21:
  br label %for.cond20
for.end19:
  %t76 = load i64, ptr %slot_digits_12
  store i64 %t76, ptr %slot_last_15
  br label %for.cond26
for.cond26:
  %t77 = load i64, ptr %slot_last_15
  %t78 = icmp ugt i64 %t77, 0
  br i1 %t78, label %for.body28, label %for.end25
for.body28:
  %t79 = load i64, ptr %slot_pos_4
  %t80 = load i64, ptr %slot_nd_7
  %t81 = add i64 %t79, %t80
  %t82 = load i64, ptr %slot_last_15
  %t83 = add i64 %t81, %t82
  %t84 = load { ptr, i64 }, ptr %slot_buf_2
  %t85 = extractvalue { ptr, i64 } %t84, 0
  %t86 = extractvalue { ptr, i64 } %t84, 1
  %t87 = icmp uge i64 %t83, %t86
  br i1 %t87, label %chk.bad33, label %chk.ok32
chk.bad33:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 87)
  unreachable
chk.ok32:
  %t88 = getelementptr inbounds i8, ptr %t85, i64 %t83
  %t89 = load i8, ptr %t88
  %t90 = icmp ne i8 %t89, 48
  br i1 %t90, label %if.then30, label %if.next31
if.then30:
  br label %for.end25
if.next31:
  br label %if.end29
if.end29:
  %t91 = load i64, ptr %slot_last_15
  %t92 = sub i64 %t91, 1
  store i64 %t92, ptr %slot_last_15
  br label %for.step27
for.step27:
  br label %for.cond26
for.end25:
  %t93 = load i64, ptr %slot_last_15
  %t94 = icmp ugt i64 %t93, 0
  br i1 %t94, label %if.then35, label %if.next36
if.then35:
  %t95 = load i64, ptr %slot_pos_4
  %t96 = load i64, ptr %slot_nd_7
  %t97 = add i64 %t95, %t96
  %t98 = load { ptr, i64 }, ptr %slot_buf_2
  %t99 = extractvalue { ptr, i64 } %t98, 0
  %t100 = extractvalue { ptr, i64 } %t98, 1
  %t101 = icmp uge i64 %t97, %t100
  br i1 %t101, label %chk.bad38, label %chk.ok37
chk.bad38:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 93)
  unreachable
chk.ok37:
  %t102 = getelementptr inbounds i8, ptr %t99, i64 %t97
  store i8 46, ptr %t102
  %t103 = load i64, ptr %slot_pos_4
  %t104 = load i64, ptr %slot_nd_7
  %t105 = add i64 %t103, %t104
  %t106 = add i64 %t105, 1
  %t107 = load i64, ptr %slot_last_15
  %t108 = add i64 %t106, %t107
  ret i64 %t108
if.next36:
  br label %if.end34
if.end34:
  br label %if.end16
if.next18:
  br label %if.end16
if.end16:
  %t109 = load i64, ptr %slot_pos_4
  %t110 = load i64, ptr %slot_nd_7
  %t111 = add i64 %t109, %t110
  ret i64 %t111
}

define i64 @aila_std_core__digits_u64(i64 %p_v) {
  %slot_v_1 = alloca i64, align 8
  %slot_x_2 = alloca i64, align 8
  %slot_n_3 = alloca i64, align 8
  store i64 %p_v, ptr %slot_v_1
  %t1 = load i64, ptr %slot_v_1
  store i64 %t1, ptr %slot_x_2
  store i64 1, ptr %slot_n_3
  br label %for.cond2
for.cond2:
  %t2 = load i64, ptr %slot_x_2
  %t3 = icmp ult i64 %t2, 10
  %t4 = xor i1 %t3, true
  br i1 %t4, label %for.body4, label %for.end1
for.body4:
  %t5 = load i64, ptr %slot_x_2
  %t6 = icmp eq i64 10, 0
  br i1 %t6, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 107)
  unreachable
chk.ok5:
  %t7 = udiv i64 %t5, 10
  store i64 %t7, ptr %slot_x_2
  %t8 = load i64, ptr %slot_n_3
  %t9 = add i64 %t8, 1
  store i64 %t9, ptr %slot_n_3
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t10 = load i64, ptr %slot_n_3
  ret i64 %t10
}
