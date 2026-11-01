; AILA module: std.strconv
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.1 = private unnamed_addr constant [17 x i8] c"std/strconv.aila\00", align 1
@.str.2 = private unnamed_addr constant [17 x i8] c"division by zero\00", align 1

declare i64 @aila_std_core__fill_i64(i64, { ptr, i64 })
declare i64 @aila_std_core__fill_u64(i64, { ptr, i64 })
declare i64 @aila_std_core__fill_f64(double, { ptr, i64 })
declare void @aila_rt_trap(ptr, i64, ptr, i64)

define i64 @aila_std_strconv__format_i64(i64 %p_v, { ptr, i64 } %p_buf) {
  %slot_v_1 = alloca i64, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  store i64 %p_v, ptr %slot_v_1
  store { ptr, i64 } %p_buf, ptr %slot_buf_2
  %t1 = load i64, ptr %slot_v_1
  %t2 = load { ptr, i64 }, ptr %slot_buf_2
  %t3 = call i64 @aila_std_core__fill_i64(i64 %t1, { ptr, i64 } %t2)
  ret i64 %t3
}

define i64 @aila_std_strconv__format_u64(i64 %p_v, { ptr, i64 } %p_buf) {
  %slot_v_1 = alloca i64, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  store i64 %p_v, ptr %slot_v_1
  store { ptr, i64 } %p_buf, ptr %slot_buf_2
  %t1 = load i64, ptr %slot_v_1
  %t2 = load { ptr, i64 }, ptr %slot_buf_2
  %t3 = call i64 @aila_std_core__fill_u64(i64 %t1, { ptr, i64 } %t2)
  ret i64 %t3
}

define i64 @aila_std_strconv__format_f64(double %p_v, { ptr, i64 } %p_buf) {
  %slot_v_1 = alloca double, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  store double %p_v, ptr %slot_v_1
  store { ptr, i64 } %p_buf, ptr %slot_buf_2
  %t1 = load double, ptr %slot_v_1
  %t2 = load { ptr, i64 }, ptr %slot_buf_2
  %t3 = call i64 @aila_std_core__fill_f64(double %t1, { ptr, i64 } %t2)
  ret i64 %t3
}

define { i1, i64, { i8, { i64 } } } @aila_std_strconv__parse_i64({ ptr, i64 } %p_t) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_b_2 = alloca { ptr, i64 }, align 8
  %slot_n_3 = alloca i64, align 8
  %slot_i_4 = alloca i64, align 8
  %slot_neg_5 = alloca i1, align 8
  %slot_max_6 = alloca i64, align 8
  %slot_acc_7 = alloca i64, align 8
  %slot_d_8 = alloca i64, align 8
  %slot_dv_9 = alloca i64, align 8
  %slot___lit_10 = alloca { i8, { i64 } }, align 8
  %slot___lit_11 = alloca { i8, { i64 } }, align 8
  %slot___lit_12 = alloca { i8, { i64 } }, align 8
  %slot___lit_13 = alloca { i8, { i64 } }, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  %t1 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t1, ptr %slot_b_2
  %t2 = load { ptr, i64 }, ptr %slot_b_2
  %t3 = extractvalue { ptr, i64 } %t2, 1
  store i64 %t3, ptr %slot_n_3
  %t4 = load { ptr, i64 }, ptr %slot_b_2
  %t5 = call i64 @aila_std_strconv__skip_ws({ ptr, i64 } %t4, i64 0)
  store i64 %t5, ptr %slot_i_4
  %t6 = load i64, ptr %slot_i_4
  %t7 = load i64, ptr %slot_n_3
  %t8 = icmp eq i64 %t6, %t7
  br i1 %t8, label %if.then2, label %if.next3
if.then2:
  %t9 = insertvalue { i8, { i64 } } undef, i8 0, 0
  %t10 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t11 = insertvalue { i1, i64, { i8, { i64 } } } %t10, { i8, { i64 } } %t9, 2
  ret { i1, i64, { i8, { i64 } } } %t11
if.next3:
  br label %if.end1
if.end1:
  store i1 0, ptr %slot_neg_5
  %t12 = load i64, ptr %slot_i_4
  %t13 = load { ptr, i64 }, ptr %slot_b_2
  %t14 = extractvalue { ptr, i64 } %t13, 0
  %t15 = extractvalue { ptr, i64 } %t13, 1
  %t16 = icmp uge i64 %t12, %t15
  br i1 %t16, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 43)
  unreachable
chk.ok7:
  %t17 = getelementptr inbounds i8, ptr %t14, i64 %t12
  %t18 = load i8, ptr %t17
  %t19 = icmp eq i8 %t18, 43
  br i1 %t19, label %if.then5, label %if.next6
if.then5:
  %t20 = load i64, ptr %slot_i_4
  %t21 = add i64 %t20, 1
  store i64 %t21, ptr %slot_i_4
  br label %if.end4
if.next6:
  %t22 = load i64, ptr %slot_i_4
  %t23 = load { ptr, i64 }, ptr %slot_b_2
  %t24 = extractvalue { ptr, i64 } %t23, 0
  %t25 = extractvalue { ptr, i64 } %t23, 1
  %t26 = icmp uge i64 %t22, %t25
  br i1 %t26, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 45)
  unreachable
chk.ok11:
  %t27 = getelementptr inbounds i8, ptr %t24, i64 %t22
  %t28 = load i8, ptr %t27
  %t29 = icmp eq i8 %t28, 45
  br i1 %t29, label %if.then9, label %if.next10
if.then9:
  store i1 1, ptr %slot_neg_5
  %t30 = load i64, ptr %slot_i_4
  %t31 = add i64 %t30, 1
  store i64 %t31, ptr %slot_i_4
  br label %if.end4
if.next10:
  br label %if.end4
if.end4:
  %t32 = load i64, ptr %slot_i_4
  %t33 = load i64, ptr %slot_n_3
  %t34 = icmp eq i64 %t32, %t33
  br i1 %t34, label %if.then14, label %if.next15
if.then14:
  %t35 = insertvalue { i8, { i64 } } undef, i8 1, 0
  store { i8, { i64 } } %t35, ptr %slot___lit_11
  %t36 = load i64, ptr %slot_i_4
  %t37 = getelementptr inbounds { i8, { i64 } }, ptr %slot___lit_11, i32 0, i32 1, i32 0
  store i64 %t36, ptr %t37
  %t38 = load { i8, { i64 } }, ptr %slot___lit_11
  %t39 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t40 = insertvalue { i1, i64, { i8, { i64 } } } %t39, { i8, { i64 } } %t38, 2
  ret { i1, i64, { i8, { i64 } } } %t40
if.next15:
  br label %if.end13
if.end13:
  store i64 9223372036854775807, ptr %slot_max_6
  %t41 = load i1, ptr %slot_neg_5
  br i1 %t41, label %if.then17, label %if.next18
if.then17:
  store i64 -9223372036854775808, ptr %slot_max_6
  br label %if.end16
if.next18:
  br label %if.end16
if.end16:
  store i64 0, ptr %slot_acc_7
  br label %for.cond20
for.cond20:
  %t42 = load i64, ptr %slot_i_4
  %t43 = load i64, ptr %slot_n_3
  %t44 = icmp ult i64 %t42, %t43
  br i1 %t44, label %for.body22, label %for.end19
for.body22:
  %t45 = load i64, ptr %slot_i_4
  %t46 = load { ptr, i64 }, ptr %slot_b_2
  %t47 = extractvalue { ptr, i64 } %t46, 0
  %t48 = extractvalue { ptr, i64 } %t46, 1
  %t49 = icmp uge i64 %t45, %t48
  br i1 %t49, label %chk.bad24, label %chk.ok23
chk.bad24:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 60)
  unreachable
chk.ok23:
  %t50 = getelementptr inbounds i8, ptr %t47, i64 %t45
  %t51 = load i8, ptr %t50
  %t52 = call i64 @aila_std_strconv__digit_val(i8 %t51)
  store i64 %t52, ptr %slot_d_8
  %t53 = load i64, ptr %slot_d_8
  %t54 = icmp slt i64 %t53, 0
  br i1 %t54, label %if.then26, label %if.next27
if.then26:
  %t55 = insertvalue { i8, { i64 } } undef, i8 1, 0
  store { i8, { i64 } } %t55, ptr %slot___lit_12
  %t56 = load i64, ptr %slot_i_4
  %t57 = getelementptr inbounds { i8, { i64 } }, ptr %slot___lit_12, i32 0, i32 1, i32 0
  store i64 %t56, ptr %t57
  %t58 = load { i8, { i64 } }, ptr %slot___lit_12
  %t59 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t60 = insertvalue { i1, i64, { i8, { i64 } } } %t59, { i8, { i64 } } %t58, 2
  ret { i1, i64, { i8, { i64 } } } %t60
if.next27:
  br label %if.end25
if.end25:
  %t61 = load i64, ptr %slot_d_8
  store i64 %t61, ptr %slot_dv_9
  %t63 = load i64, ptr %slot_acc_7
  %t64 = load i64, ptr %slot_max_6
  %t65 = load i64, ptr %slot_dv_9
  %t66 = sub i64 %t64, %t65
  %t67 = icmp eq i64 10, 0
  br i1 %t67, label %chk.bad32, label %chk.ok31
chk.bad32:
  call void @aila_rt_trap(ptr @.str.2, i64 16, ptr @.str.1, i64 65)
  unreachable
chk.ok31:
  %t68 = udiv i64 %t66, 10
  %t69 = icmp ugt i64 %t63, %t68
  br i1 %t69, label %if.then29, label %if.next30
if.then29:
  %t70 = insertvalue { i8, { i64 } } undef, i8 2, 0
  %t71 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t72 = insertvalue { i1, i64, { i8, { i64 } } } %t71, { i8, { i64 } } %t70, 2
  ret { i1, i64, { i8, { i64 } } } %t72
if.next30:
  br label %if.end28
if.end28:
  %t73 = load i64, ptr %slot_acc_7
  %t74 = mul i64 %t73, 10
  %t75 = load i64, ptr %slot_dv_9
  %t76 = add i64 %t74, %t75
  store i64 %t76, ptr %slot_acc_7
  %t77 = load i64, ptr %slot_i_4
  %t78 = add i64 %t77, 1
  store i64 %t78, ptr %slot_i_4
  br label %for.step21
for.step21:
  br label %for.cond20
for.end19:
  %t79 = load i1, ptr %slot_neg_5
  br i1 %t79, label %if.then34, label %if.next35
if.then34:
  %t80 = load i64, ptr %slot_acc_7
  %t81 = icmp eq i64 %t80, -9223372036854775808
  br i1 %t81, label %if.then37, label %if.next38
if.then37:
  %t82 = sub i64 0, -9223372036854775808
  %t83 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 1, 0
  %t84 = insertvalue { i1, i64, { i8, { i64 } } } %t83, i64 %t82, 1
  ret { i1, i64, { i8, { i64 } } } %t84
if.next38:
  br label %if.end36
if.end36:
  %t85 = load i64, ptr %slot_acc_7
  %t87 = sub i64 0, %t85
  %t88 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 1, 0
  %t89 = insertvalue { i1, i64, { i8, { i64 } } } %t88, i64 %t87, 1
  ret { i1, i64, { i8, { i64 } } } %t89
if.next35:
  br label %if.end33
if.end33:
  %t90 = load i64, ptr %slot_acc_7
  %t92 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 1, 0
  %t93 = insertvalue { i1, i64, { i8, { i64 } } } %t92, i64 %t90, 1
  ret { i1, i64, { i8, { i64 } } } %t93
}

define { i1, i64, { i8, { i64 } } } @aila_std_strconv__parse_u64({ ptr, i64 } %p_t) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_b_2 = alloca { ptr, i64 }, align 8
  %slot_n_3 = alloca i64, align 8
  %slot_i_4 = alloca i64, align 8
  %slot_acc_5 = alloca i64, align 8
  %slot_d_6 = alloca i64, align 8
  %slot_dv_7 = alloca i64, align 8
  %slot___lit_8 = alloca { i8, { i64 } }, align 8
  %slot___lit_9 = alloca { i8, { i64 } }, align 8
  %slot___lit_10 = alloca { i8, { i64 } }, align 8
  %slot___lit_11 = alloca { i8, { i64 } }, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  %t1 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t1, ptr %slot_b_2
  %t2 = load { ptr, i64 }, ptr %slot_b_2
  %t3 = extractvalue { ptr, i64 } %t2, 1
  store i64 %t3, ptr %slot_n_3
  %t4 = load { ptr, i64 }, ptr %slot_b_2
  %t5 = call i64 @aila_std_strconv__skip_ws({ ptr, i64 } %t4, i64 0)
  store i64 %t5, ptr %slot_i_4
  %t6 = load i64, ptr %slot_i_4
  %t7 = load i64, ptr %slot_n_3
  %t8 = icmp eq i64 %t6, %t7
  br i1 %t8, label %if.then2, label %if.next3
if.then2:
  %t9 = insertvalue { i8, { i64 } } undef, i8 0, 0
  %t10 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t11 = insertvalue { i1, i64, { i8, { i64 } } } %t10, { i8, { i64 } } %t9, 2
  ret { i1, i64, { i8, { i64 } } } %t11
if.next3:
  br label %if.end1
if.end1:
  %t12 = load i64, ptr %slot_i_4
  %t13 = load { ptr, i64 }, ptr %slot_b_2
  %t14 = extractvalue { ptr, i64 } %t13, 0
  %t15 = extractvalue { ptr, i64 } %t13, 1
  %t16 = icmp uge i64 %t12, %t15
  br i1 %t16, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 90)
  unreachable
chk.ok7:
  %t17 = getelementptr inbounds i8, ptr %t14, i64 %t12
  %t18 = load i8, ptr %t17
  %t19 = icmp eq i8 %t18, 43
  br i1 %t19, label %if.then5, label %if.next6
if.then5:
  %t20 = load i64, ptr %slot_i_4
  %t21 = add i64 %t20, 1
  store i64 %t21, ptr %slot_i_4
  br label %if.end4
if.next6:
  br label %if.end4
if.end4:
  %t22 = load i64, ptr %slot_i_4
  %t23 = load i64, ptr %slot_n_3
  %t24 = icmp eq i64 %t22, %t23
  br i1 %t24, label %if.then10, label %if.next11
if.then10:
  %t25 = insertvalue { i8, { i64 } } undef, i8 1, 0
  store { i8, { i64 } } %t25, ptr %slot___lit_9
  %t26 = load i64, ptr %slot_i_4
  %t27 = getelementptr inbounds { i8, { i64 } }, ptr %slot___lit_9, i32 0, i32 1, i32 0
  store i64 %t26, ptr %t27
  %t28 = load { i8, { i64 } }, ptr %slot___lit_9
  %t29 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t30 = insertvalue { i1, i64, { i8, { i64 } } } %t29, { i8, { i64 } } %t28, 2
  ret { i1, i64, { i8, { i64 } } } %t30
if.next11:
  br label %if.end9
if.end9:
  store i64 0, ptr %slot_acc_5
  br label %for.cond13
for.cond13:
  %t31 = load i64, ptr %slot_i_4
  %t32 = load i64, ptr %slot_n_3
  %t33 = icmp ult i64 %t31, %t32
  br i1 %t33, label %for.body15, label %for.end12
for.body15:
  %t34 = load i64, ptr %slot_i_4
  %t35 = load { ptr, i64 }, ptr %slot_b_2
  %t36 = extractvalue { ptr, i64 } %t35, 0
  %t37 = extractvalue { ptr, i64 } %t35, 1
  %t38 = icmp uge i64 %t34, %t37
  br i1 %t38, label %chk.bad17, label %chk.ok16
chk.bad17:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 98)
  unreachable
chk.ok16:
  %t39 = getelementptr inbounds i8, ptr %t36, i64 %t34
  %t40 = load i8, ptr %t39
  %t41 = call i64 @aila_std_strconv__digit_val(i8 %t40)
  store i64 %t41, ptr %slot_d_6
  %t42 = load i64, ptr %slot_d_6
  %t43 = icmp slt i64 %t42, 0
  br i1 %t43, label %if.then19, label %if.next20
if.then19:
  %t44 = insertvalue { i8, { i64 } } undef, i8 1, 0
  store { i8, { i64 } } %t44, ptr %slot___lit_10
  %t45 = load i64, ptr %slot_i_4
  %t46 = getelementptr inbounds { i8, { i64 } }, ptr %slot___lit_10, i32 0, i32 1, i32 0
  store i64 %t45, ptr %t46
  %t47 = load { i8, { i64 } }, ptr %slot___lit_10
  %t48 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t49 = insertvalue { i1, i64, { i8, { i64 } } } %t48, { i8, { i64 } } %t47, 2
  ret { i1, i64, { i8, { i64 } } } %t49
if.next20:
  br label %if.end18
if.end18:
  %t50 = load i64, ptr %slot_d_6
  store i64 %t50, ptr %slot_dv_7
  %t52 = load i64, ptr %slot_acc_5
  %t53 = load i64, ptr %slot_dv_7
  %t54 = sub i64 -1, %t53
  %t55 = icmp eq i64 10, 0
  br i1 %t55, label %chk.bad25, label %chk.ok24
chk.bad25:
  call void @aila_rt_trap(ptr @.str.2, i64 16, ptr @.str.1, i64 103)
  unreachable
chk.ok24:
  %t56 = udiv i64 %t54, 10
  %t57 = icmp ugt i64 %t52, %t56
  br i1 %t57, label %if.then22, label %if.next23
if.then22:
  %t58 = insertvalue { i8, { i64 } } undef, i8 2, 0
  %t59 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t60 = insertvalue { i1, i64, { i8, { i64 } } } %t59, { i8, { i64 } } %t58, 2
  ret { i1, i64, { i8, { i64 } } } %t60
if.next23:
  br label %if.end21
if.end21:
  %t61 = load i64, ptr %slot_acc_5
  %t62 = mul i64 %t61, 10
  %t63 = load i64, ptr %slot_dv_7
  %t64 = add i64 %t62, %t63
  store i64 %t64, ptr %slot_acc_5
  %t65 = load i64, ptr %slot_i_4
  %t66 = add i64 %t65, 1
  store i64 %t66, ptr %slot_i_4
  br label %for.step14
for.step14:
  br label %for.cond13
for.end12:
  %t67 = load i64, ptr %slot_acc_5
  %t68 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 1, 0
  %t69 = insertvalue { i1, i64, { i8, { i64 } } } %t68, i64 %t67, 1
  ret { i1, i64, { i8, { i64 } } } %t69
}

define { i1, double, { i8, { i64 } } } @aila_std_strconv__parse_f64({ ptr, i64 } %p_t) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_b_2 = alloca { ptr, i64 }, align 8
  %slot_n_3 = alloca i64, align 8
  %slot_i_4 = alloca i64, align 8
  %slot_neg_5 = alloca i1, align 8
  %slot_mant_6 = alloca double, align 8
  %slot_any_7 = alloca i1, align 8
  %slot_d_8 = alloca i64, align 8
  %slot_d_9 = alloca i64, align 8
  %slot_enorm_10 = alloca i1, align 8
  %slot_e_11 = alloca i64, align 8
  %slot_d_12 = alloca i64, align 8
  %slot___lit_13 = alloca { i8, { i64 } }, align 8
  %slot___lit_14 = alloca { i8, { i64 } }, align 8
  %slot___sc_15 = alloca i1, align 8
  %slot___lit_16 = alloca { i8, { i64 } }, align 8
  %slot___sc_17 = alloca i1, align 8
  %slot___sc_18 = alloca i1, align 8
  %slot___sc_19 = alloca i1, align 8
  %slot___sc_20 = alloca i1, align 8
  %slot___lit_21 = alloca { i8, { i64 } }, align 8
  %slot___lit_22 = alloca { i8, { i64 } }, align 8
  %slot___lit_23 = alloca { i8, { i64 } }, align 8
  %slot___lit_24 = alloca { i8, { i64 } }, align 8
  %slot___lit_25 = alloca { i8, { i64 } }, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  %t1 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t1, ptr %slot_b_2
  %t2 = load { ptr, i64 }, ptr %slot_b_2
  %t3 = extractvalue { ptr, i64 } %t2, 1
  store i64 %t3, ptr %slot_n_3
  %t4 = load { ptr, i64 }, ptr %slot_b_2
  %t5 = call i64 @aila_std_strconv__skip_ws({ ptr, i64 } %t4, i64 0)
  store i64 %t5, ptr %slot_i_4
  %t6 = load i64, ptr %slot_i_4
  %t7 = load i64, ptr %slot_n_3
  %t8 = icmp eq i64 %t6, %t7
  br i1 %t8, label %if.then2, label %if.next3
if.then2:
  %t9 = insertvalue { i8, { i64 } } undef, i8 0, 0
  %t10 = insertvalue { i1, double, { i8, { i64 } } } undef, i1 0, 0
  %t11 = insertvalue { i1, double, { i8, { i64 } } } %t10, { i8, { i64 } } %t9, 2
  ret { i1, double, { i8, { i64 } } } %t11
if.next3:
  br label %if.end1
if.end1:
  store i1 0, ptr %slot_neg_5
  %t12 = load i64, ptr %slot_i_4
  %t13 = load { ptr, i64 }, ptr %slot_b_2
  %t14 = extractvalue { ptr, i64 } %t13, 0
  %t15 = extractvalue { ptr, i64 } %t13, 1
  %t16 = icmp uge i64 %t12, %t15
  br i1 %t16, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 123)
  unreachable
chk.ok7:
  %t17 = getelementptr inbounds i8, ptr %t14, i64 %t12
  %t18 = load i8, ptr %t17
  %t19 = icmp eq i8 %t18, 43
  br i1 %t19, label %if.then5, label %if.next6
if.then5:
  %t20 = load i64, ptr %slot_i_4
  %t21 = add i64 %t20, 1
  store i64 %t21, ptr %slot_i_4
  br label %if.end4
if.next6:
  %t22 = load i64, ptr %slot_i_4
  %t23 = load { ptr, i64 }, ptr %slot_b_2
  %t24 = extractvalue { ptr, i64 } %t23, 0
  %t25 = extractvalue { ptr, i64 } %t23, 1
  %t26 = icmp uge i64 %t22, %t25
  br i1 %t26, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 125)
  unreachable
chk.ok11:
  %t27 = getelementptr inbounds i8, ptr %t24, i64 %t22
  %t28 = load i8, ptr %t27
  %t29 = icmp eq i8 %t28, 45
  br i1 %t29, label %if.then9, label %if.next10
if.then9:
  store i1 1, ptr %slot_neg_5
  %t30 = load i64, ptr %slot_i_4
  %t31 = add i64 %t30, 1
  store i64 %t31, ptr %slot_i_4
  br label %if.end4
if.next10:
  br label %if.end4
if.end4:
  %t32 = load i64, ptr %slot_i_4
  %t33 = load i64, ptr %slot_n_3
  %t34 = icmp eq i64 %t32, %t33
  br i1 %t34, label %if.then14, label %if.next15
if.then14:
  %t35 = insertvalue { i8, { i64 } } undef, i8 1, 0
  store { i8, { i64 } } %t35, ptr %slot___lit_14
  %t36 = load i64, ptr %slot_i_4
  %t37 = getelementptr inbounds { i8, { i64 } }, ptr %slot___lit_14, i32 0, i32 1, i32 0
  store i64 %t36, ptr %t37
  %t38 = load { i8, { i64 } }, ptr %slot___lit_14
  %t39 = insertvalue { i1, double, { i8, { i64 } } } undef, i1 0, 0
  %t40 = insertvalue { i1, double, { i8, { i64 } } } %t39, { i8, { i64 } } %t38, 2
  ret { i1, double, { i8, { i64 } } } %t40
if.next15:
  br label %if.end13
if.end13:
  store double 0.0, ptr %slot_mant_6
  store i1 0, ptr %slot_any_7
  br label %for.cond17
for.cond17:
  %t41 = load i64, ptr %slot_i_4
  %t42 = load i64, ptr %slot_n_3
  %t43 = icmp ult i64 %t41, %t42
  br i1 %t43, label %for.body19, label %for.end16
for.body19:
  %t44 = load i64, ptr %slot_i_4
  %t45 = load { ptr, i64 }, ptr %slot_b_2
  %t46 = extractvalue { ptr, i64 } %t45, 0
  %t47 = extractvalue { ptr, i64 } %t45, 1
  %t48 = icmp uge i64 %t44, %t47
  br i1 %t48, label %chk.bad21, label %chk.ok20
chk.bad21:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 135)
  unreachable
chk.ok20:
  %t49 = getelementptr inbounds i8, ptr %t46, i64 %t44
  %t50 = load i8, ptr %t49
  %t51 = call i64 @aila_std_strconv__digit_val(i8 %t50)
  store i64 %t51, ptr %slot_d_8
  %t52 = load i64, ptr %slot_d_8
  %t53 = icmp slt i64 %t52, 0
  br i1 %t53, label %if.then23, label %if.next24
if.then23:
  br label %for.end16
if.next24:
  br label %if.end22
if.end22:
  %t54 = load double, ptr %slot_mant_6
  %t55 = fmul double %t54, 10.0
  %t56 = load i64, ptr %slot_d_8
  %t57 = sitofp i64 %t56 to double
  %t58 = fadd double %t55, %t57
  store double %t58, ptr %slot_mant_6
  store i1 1, ptr %slot_any_7
  %t59 = load i64, ptr %slot_i_4
  %t60 = add i64 %t59, 1
  store i64 %t60, ptr %slot_i_4
  br label %for.step18
for.step18:
  br label %for.cond17
for.end16:
  %t61 = load i64, ptr %slot_i_4
  %t62 = load i64, ptr %slot_n_3
  %t63 = icmp ult i64 %t61, %t62
  br i1 %t63, label %sc.r29, label %sc.l28
sc.l28:
  store i1 0, ptr %slot___sc_15
  br label %sc.end30
sc.r29:
  %t64 = load i64, ptr %slot_i_4
  %t65 = load { ptr, i64 }, ptr %slot_b_2
  %t66 = extractvalue { ptr, i64 } %t65, 0
  %t67 = extractvalue { ptr, i64 } %t65, 1
  %t68 = icmp uge i64 %t64, %t67
  br i1 %t68, label %chk.bad32, label %chk.ok31
chk.bad32:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 143)
  unreachable
chk.ok31:
  %t69 = getelementptr inbounds i8, ptr %t66, i64 %t64
  %t70 = load i8, ptr %t69
  %t71 = icmp eq i8 %t70, 46
  store i1 %t71, ptr %slot___sc_15
  br label %sc.end30
sc.end30:
  %t72 = load i1, ptr %slot___sc_15
  br i1 %t72, label %if.then26, label %if.next27
if.then26:
  %t73 = load i64, ptr %slot_i_4
  %t74 = add i64 %t73, 1
  store i64 %t74, ptr %slot_i_4
  br label %for.cond34
for.cond34:
  %t75 = load i64, ptr %slot_i_4
  %t76 = load i64, ptr %slot_n_3
  %t77 = icmp ult i64 %t75, %t76
  br i1 %t77, label %for.body36, label %for.end33
for.body36:
  %t78 = load i64, ptr %slot_i_4
  %t79 = load { ptr, i64 }, ptr %slot_b_2
  %t80 = extractvalue { ptr, i64 } %t79, 0
  %t81 = extractvalue { ptr, i64 } %t79, 1
  %t82 = icmp uge i64 %t78, %t81
  br i1 %t82, label %chk.bad38, label %chk.ok37
chk.bad38:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 146)
  unreachable
chk.ok37:
  %t83 = getelementptr inbounds i8, ptr %t80, i64 %t78
  %t84 = load i8, ptr %t83
  %t85 = call i64 @aila_std_strconv__digit_val(i8 %t84)
  store i64 %t85, ptr %slot_d_9
  %t86 = load i64, ptr %slot_d_9
  %t87 = icmp slt i64 %t86, 0
  br i1 %t87, label %if.then40, label %if.next41
if.then40:
  br label %for.end33
if.next41:
  br label %if.end39
if.end39:
  %t88 = load double, ptr %slot_mant_6
  %t89 = fmul double %t88, 10.0
  %t90 = load i64, ptr %slot_d_9
  %t91 = sitofp i64 %t90 to double
  %t92 = fadd double %t89, %t91
  store double %t92, ptr %slot_mant_6
  store i1 1, ptr %slot_any_7
  %t93 = load i64, ptr %slot_i_4
  %t94 = add i64 %t93, 1
  store i64 %t94, ptr %slot_i_4
  %t95 = load double, ptr %slot_mant_6
  %t96 = fdiv double %t95, 10.0
  store double %t96, ptr %slot_mant_6
  br label %for.step35
for.step35:
  br label %for.cond34
for.end33:
  br label %if.end25
if.next27:
  br label %if.end25
if.end25:
  %t97 = load i1, ptr %slot_any_7
  %t98 = xor i1 %t97, true
  br i1 %t98, label %if.then43, label %if.next44
if.then43:
  %t99 = insertvalue { i8, { i64 } } undef, i8 1, 0
  store { i8, { i64 } } %t99, ptr %slot___lit_16
  %t100 = load i64, ptr %slot_i_4
  %t101 = getelementptr inbounds { i8, { i64 } }, ptr %slot___lit_16, i32 0, i32 1, i32 0
  store i64 %t100, ptr %t101
  %t102 = load { i8, { i64 } }, ptr %slot___lit_16
  %t103 = insertvalue { i1, double, { i8, { i64 } } } undef, i1 0, 0
  %t104 = insertvalue { i1, double, { i8, { i64 } } } %t103, { i8, { i64 } } %t102, 2
  ret { i1, double, { i8, { i64 } } } %t104
if.next44:
  br label %if.end42
if.end42:
  %t105 = load i64, ptr %slot_i_4
  %t106 = load i64, ptr %slot_n_3
  %t107 = icmp ult i64 %t105, %t106
  br i1 %t107, label %sc.r49, label %sc.l48
sc.l48:
  store i1 0, ptr %slot___sc_17
  br label %sc.end50
sc.r49:
  %t108 = load i64, ptr %slot_i_4
  %t109 = load { ptr, i64 }, ptr %slot_b_2
  %t110 = extractvalue { ptr, i64 } %t109, 0
  %t111 = extractvalue { ptr, i64 } %t109, 1
  %t112 = icmp uge i64 %t108, %t111
  br i1 %t112, label %chk.bad55, label %chk.ok54
chk.bad55:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 159)
  unreachable
chk.ok54:
  %t113 = getelementptr inbounds i8, ptr %t110, i64 %t108
  %t114 = load i8, ptr %t113
  %t115 = icmp eq i8 %t114, 101
  br i1 %t115, label %sc.l51, label %sc.r52
sc.l51:
  store i1 1, ptr %slot___sc_18
  br label %sc.end53
sc.r52:
  %t116 = load i64, ptr %slot_i_4
  %t117 = load { ptr, i64 }, ptr %slot_b_2
  %t118 = extractvalue { ptr, i64 } %t117, 0
  %t119 = extractvalue { ptr, i64 } %t117, 1
  %t120 = icmp uge i64 %t116, %t119
  br i1 %t120, label %chk.bad57, label %chk.ok56
chk.bad57:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 159)
  unreachable
chk.ok56:
  %t121 = getelementptr inbounds i8, ptr %t118, i64 %t116
  %t122 = load i8, ptr %t121
  %t123 = icmp eq i8 %t122, 69
  store i1 %t123, ptr %slot___sc_18
  br label %sc.end53
sc.end53:
  %t124 = load i1, ptr %slot___sc_18
  store i1 %t124, ptr %slot___sc_17
  br label %sc.end50
sc.end50:
  %t125 = load i1, ptr %slot___sc_17
  br i1 %t125, label %if.then46, label %if.next47
if.then46:
  %t126 = load i64, ptr %slot_i_4
  %t127 = add i64 %t126, 1
  store i64 %t127, ptr %slot_i_4
  store i1 0, ptr %slot_enorm_10
  %t128 = load i64, ptr %slot_i_4
  %t129 = load i64, ptr %slot_n_3
  %t130 = icmp ult i64 %t128, %t129
  br i1 %t130, label %sc.r62, label %sc.l61
sc.l61:
  store i1 0, ptr %slot___sc_19
  br label %sc.end63
sc.r62:
  %t131 = load i64, ptr %slot_i_4
  %t132 = load { ptr, i64 }, ptr %slot_b_2
  %t133 = extractvalue { ptr, i64 } %t132, 0
  %t134 = extractvalue { ptr, i64 } %t132, 1
  %t135 = icmp uge i64 %t131, %t134
  br i1 %t135, label %chk.bad65, label %chk.ok64
chk.bad65:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 162)
  unreachable
chk.ok64:
  %t136 = getelementptr inbounds i8, ptr %t133, i64 %t131
  %t137 = load i8, ptr %t136
  %t138 = icmp eq i8 %t137, 43
  store i1 %t138, ptr %slot___sc_19
  br label %sc.end63
sc.end63:
  %t139 = load i1, ptr %slot___sc_19
  br i1 %t139, label %if.then59, label %if.next60
if.then59:
  %t140 = load i64, ptr %slot_i_4
  %t141 = add i64 %t140, 1
  store i64 %t141, ptr %slot_i_4
  br label %if.end58
if.next60:
  %t142 = load i64, ptr %slot_i_4
  %t143 = load i64, ptr %slot_n_3
  %t144 = icmp ult i64 %t142, %t143
  br i1 %t144, label %sc.r69, label %sc.l68
sc.l68:
  store i1 0, ptr %slot___sc_20
  br label %sc.end70
sc.r69:
  %t145 = load i64, ptr %slot_i_4
  %t146 = load { ptr, i64 }, ptr %slot_b_2
  %t147 = extractvalue { ptr, i64 } %t146, 0
  %t148 = extractvalue { ptr, i64 } %t146, 1
  %t149 = icmp uge i64 %t145, %t148
  br i1 %t149, label %chk.bad72, label %chk.ok71
chk.bad72:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 164)
  unreachable
chk.ok71:
  %t150 = getelementptr inbounds i8, ptr %t147, i64 %t145
  %t151 = load i8, ptr %t150
  %t152 = icmp eq i8 %t151, 45
  store i1 %t152, ptr %slot___sc_20
  br label %sc.end70
sc.end70:
  %t153 = load i1, ptr %slot___sc_20
  br i1 %t153, label %if.then66, label %if.next67
if.then66:
  store i1 1, ptr %slot_enorm_10
  %t154 = load i64, ptr %slot_i_4
  %t155 = add i64 %t154, 1
  store i64 %t155, ptr %slot_i_4
  br label %if.end58
if.next67:
  br label %if.end58
if.end58:
  %t156 = load i64, ptr %slot_i_4
  %t157 = load i64, ptr %slot_n_3
  %t158 = icmp eq i64 %t156, %t157
  br i1 %t158, label %if.then74, label %if.next75
if.then74:
  %t159 = insertvalue { i8, { i64 } } undef, i8 1, 0
  store { i8, { i64 } } %t159, ptr %slot___lit_21
  %t160 = load i64, ptr %slot_i_4
  %t161 = getelementptr inbounds { i8, { i64 } }, ptr %slot___lit_21, i32 0, i32 1, i32 0
  store i64 %t160, ptr %t161
  %t162 = load { i8, { i64 } }, ptr %slot___lit_21
  %t163 = insertvalue { i1, double, { i8, { i64 } } } undef, i1 0, 0
  %t164 = insertvalue { i1, double, { i8, { i64 } } } %t163, { i8, { i64 } } %t162, 2
  ret { i1, double, { i8, { i64 } } } %t164
if.next75:
  br label %if.end73
if.end73:
  store i64 0, ptr %slot_e_11
  br label %for.cond77
for.cond77:
  %t165 = load i64, ptr %slot_i_4
  %t166 = load i64, ptr %slot_n_3
  %t167 = icmp ult i64 %t165, %t166
  br i1 %t167, label %for.body79, label %for.end76
for.body79:
  %t168 = load i64, ptr %slot_i_4
  %t169 = load { ptr, i64 }, ptr %slot_b_2
  %t170 = extractvalue { ptr, i64 } %t169, 0
  %t171 = extractvalue { ptr, i64 } %t169, 1
  %t172 = icmp uge i64 %t168, %t171
  br i1 %t172, label %chk.bad81, label %chk.ok80
chk.bad81:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 173)
  unreachable
chk.ok80:
  %t173 = getelementptr inbounds i8, ptr %t170, i64 %t168
  %t174 = load i8, ptr %t173
  %t175 = call i64 @aila_std_strconv__digit_val(i8 %t174)
  store i64 %t175, ptr %slot_d_12
  %t176 = load i64, ptr %slot_d_12
  %t177 = icmp slt i64 %t176, 0
  br i1 %t177, label %if.then83, label %if.next84
if.then83:
  %t178 = insertvalue { i8, { i64 } } undef, i8 1, 0
  store { i8, { i64 } } %t178, ptr %slot___lit_22
  %t179 = load i64, ptr %slot_i_4
  %t180 = getelementptr inbounds { i8, { i64 } }, ptr %slot___lit_22, i32 0, i32 1, i32 0
  store i64 %t179, ptr %t180
  %t181 = load { i8, { i64 } }, ptr %slot___lit_22
  %t182 = insertvalue { i1, double, { i8, { i64 } } } undef, i1 0, 0
  %t183 = insertvalue { i1, double, { i8, { i64 } } } %t182, { i8, { i64 } } %t181, 2
  ret { i1, double, { i8, { i64 } } } %t183
if.next84:
  br label %if.end82
if.end82:
  %t184 = load i64, ptr %slot_e_11
  %t185 = mul i64 %t184, 10
  %t186 = load i64, ptr %slot_d_12
  %t188 = add i64 %t185, %t186
  store i64 %t188, ptr %slot_e_11
  %t189 = load i64, ptr %slot_e_11
  %t190 = icmp ugt i64 %t189, 10000
  br i1 %t190, label %if.then86, label %if.next87
if.then86:
  %t191 = insertvalue { i8, { i64 } } undef, i8 2, 0
  %t192 = insertvalue { i1, double, { i8, { i64 } } } undef, i1 0, 0
  %t193 = insertvalue { i1, double, { i8, { i64 } } } %t192, { i8, { i64 } } %t191, 2
  ret { i1, double, { i8, { i64 } } } %t193
if.next87:
  br label %if.end85
if.end85:
  %t194 = load i64, ptr %slot_i_4
  %t195 = add i64 %t194, 1
  store i64 %t195, ptr %slot_i_4
  br label %for.step78
for.step78:
  br label %for.cond77
for.end76:
  %t196 = load i1, ptr %slot_enorm_10
  br i1 %t196, label %if.then89, label %if.next90
if.then89:
  br label %for.cond92
for.cond92:
  %t197 = load i64, ptr %slot_e_11
  %t198 = icmp ugt i64 %t197, 0
  br i1 %t198, label %for.body94, label %for.end91
for.body94:
  %t199 = load double, ptr %slot_mant_6
  %t200 = fdiv double %t199, 10.0
  store double %t200, ptr %slot_mant_6
  %t201 = load i64, ptr %slot_e_11
  %t202 = sub i64 %t201, 1
  store i64 %t202, ptr %slot_e_11
  br label %for.step93
for.step93:
  br label %for.cond92
for.end91:
  br label %if.end88
if.next90:
  br label %for.cond96
for.cond96:
  %t203 = load i64, ptr %slot_e_11
  %t204 = icmp ugt i64 %t203, 0
  br i1 %t204, label %for.body98, label %for.end95
for.body98:
  %t205 = load double, ptr %slot_mant_6
  %t206 = fmul double %t205, 10.0
  store double %t206, ptr %slot_mant_6
  %t207 = load i64, ptr %slot_e_11
  %t208 = sub i64 %t207, 1
  store i64 %t208, ptr %slot_e_11
  br label %for.step97
for.step97:
  br label %for.cond96
for.end95:
  br label %if.end88
if.end88:
  br label %if.end45
if.next47:
  br label %if.end45
if.end45:
  %t209 = load i64, ptr %slot_i_4
  %t210 = load i64, ptr %slot_n_3
  %t211 = icmp ult i64 %t209, %t210
  br i1 %t211, label %if.then100, label %if.next101
if.then100:
  %t212 = insertvalue { i8, { i64 } } undef, i8 1, 0
  store { i8, { i64 } } %t212, ptr %slot___lit_24
  %t213 = load i64, ptr %slot_i_4
  %t214 = getelementptr inbounds { i8, { i64 } }, ptr %slot___lit_24, i32 0, i32 1, i32 0
  store i64 %t213, ptr %t214
  %t215 = load { i8, { i64 } }, ptr %slot___lit_24
  %t216 = insertvalue { i1, double, { i8, { i64 } } } undef, i1 0, 0
  %t217 = insertvalue { i1, double, { i8, { i64 } } } %t216, { i8, { i64 } } %t215, 2
  ret { i1, double, { i8, { i64 } } } %t217
if.next101:
  br label %if.end99
if.end99:
  %t218 = load double, ptr %slot_mant_6
  %t219 = fcmp ogt double %t218, 1.7976931348623157e+308
  br i1 %t219, label %if.then103, label %if.next104
if.then103:
  %t220 = insertvalue { i8, { i64 } } undef, i8 2, 0
  %t221 = insertvalue { i1, double, { i8, { i64 } } } undef, i1 0, 0
  %t222 = insertvalue { i1, double, { i8, { i64 } } } %t221, { i8, { i64 } } %t220, 2
  ret { i1, double, { i8, { i64 } } } %t222
if.next104:
  br label %if.end102
if.end102:
  %t223 = load i1, ptr %slot_neg_5
  br i1 %t223, label %if.then106, label %if.next107
if.then106:
  %t224 = load double, ptr %slot_mant_6
  %t225 = fsub double 0.0, %t224
  %t226 = insertvalue { i1, double, { i8, { i64 } } } undef, i1 1, 0
  %t227 = insertvalue { i1, double, { i8, { i64 } } } %t226, double %t225, 1
  ret { i1, double, { i8, { i64 } } } %t227
if.next107:
  br label %if.end105
if.end105:
  %t228 = load double, ptr %slot_mant_6
  %t229 = insertvalue { i1, double, { i8, { i64 } } } undef, i1 1, 0
  %t230 = insertvalue { i1, double, { i8, { i64 } } } %t229, double %t228, 1
  ret { i1, double, { i8, { i64 } } } %t230
}

define i64 @aila_std_strconv__skip_ws({ ptr, i64 } %p_b, i64 %p_from) {
  %slot_b_1 = alloca { ptr, i64 }, align 8
  %slot_from_2 = alloca i64, align 8
  %slot_i_3 = alloca i64, align 8
  %slot_n_4 = alloca i64, align 8
  %slot_c_5 = alloca i8, align 8
  %slot___sc_6 = alloca i1, align 8
  %slot___sc_7 = alloca i1, align 8
  %slot___sc_8 = alloca i1, align 8
  %slot___sc_9 = alloca i1, align 8
  %slot___sc_10 = alloca i1, align 8
  store { ptr, i64 } %p_b, ptr %slot_b_1
  store i64 %p_from, ptr %slot_from_2
  %t1 = load i64, ptr %slot_from_2
  store i64 %t1, ptr %slot_i_3
  %t2 = load { ptr, i64 }, ptr %slot_b_1
  %t3 = extractvalue { ptr, i64 } %t2, 1
  store i64 %t3, ptr %slot_n_4
  br label %for.cond2
for.cond2:
  %t4 = load i64, ptr %slot_i_3
  %t5 = load i64, ptr %slot_n_4
  %t6 = icmp ult i64 %t4, %t5
  br i1 %t6, label %for.body4, label %for.end1
for.body4:
  %t7 = load i64, ptr %slot_i_3
  %t8 = load { ptr, i64 }, ptr %slot_b_1
  %t9 = extractvalue { ptr, i64 } %t8, 0
  %t10 = extractvalue { ptr, i64 } %t8, 1
  %t11 = icmp uge i64 %t7, %t10
  br i1 %t11, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 215)
  unreachable
chk.ok5:
  %t12 = getelementptr inbounds i8, ptr %t9, i64 %t7
  %t13 = load i8, ptr %t12
  store i8 %t13, ptr %slot_c_5
  %t14 = load i8, ptr %slot_c_5
  %t15 = icmp eq i8 %t14, 9
  br i1 %t15, label %sc.l22, label %sc.r23
sc.l22:
  store i1 1, ptr %slot___sc_10
  br label %sc.end24
sc.r23:
  %t16 = load i8, ptr %slot_c_5
  %t17 = icmp eq i8 %t16, 10
  store i1 %t17, ptr %slot___sc_10
  br label %sc.end24
sc.end24:
  %t18 = load i1, ptr %slot___sc_10
  br i1 %t18, label %sc.l19, label %sc.r20
sc.l19:
  store i1 1, ptr %slot___sc_9
  br label %sc.end21
sc.r20:
  %t19 = load i8, ptr %slot_c_5
  %t20 = icmp eq i8 %t19, 11
  store i1 %t20, ptr %slot___sc_9
  br label %sc.end21
sc.end21:
  %t21 = load i1, ptr %slot___sc_9
  br i1 %t21, label %sc.l16, label %sc.r17
sc.l16:
  store i1 1, ptr %slot___sc_8
  br label %sc.end18
sc.r17:
  %t22 = load i8, ptr %slot_c_5
  %t23 = icmp eq i8 %t22, 12
  store i1 %t23, ptr %slot___sc_8
  br label %sc.end18
sc.end18:
  %t24 = load i1, ptr %slot___sc_8
  br i1 %t24, label %sc.l13, label %sc.r14
sc.l13:
  store i1 1, ptr %slot___sc_7
  br label %sc.end15
sc.r14:
  %t25 = load i8, ptr %slot_c_5
  %t26 = icmp eq i8 %t25, 13
  store i1 %t26, ptr %slot___sc_7
  br label %sc.end15
sc.end15:
  %t27 = load i1, ptr %slot___sc_7
  br i1 %t27, label %sc.l10, label %sc.r11
sc.l10:
  store i1 1, ptr %slot___sc_6
  br label %sc.end12
sc.r11:
  %t28 = load i8, ptr %slot_c_5
  %t29 = icmp eq i8 %t28, 32
  store i1 %t29, ptr %slot___sc_6
  br label %sc.end12
sc.end12:
  %t30 = load i1, ptr %slot___sc_6
  br i1 %t30, label %if.then8, label %if.next9
if.then8:
  %t31 = load i64, ptr %slot_i_3
  %t32 = add i64 %t31, 1
  store i64 %t32, ptr %slot_i_3
  br label %if.end7
if.next9:
  br label %for.end1
if.end7:
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t33 = load i64, ptr %slot_i_3
  ret i64 %t33
}

define i64 @aila_std_strconv__digit_val(i8 %p_c) {
  %slot_c_1 = alloca i8, align 8
  %slot___sc_2 = alloca i1, align 8
  store i8 %p_c, ptr %slot_c_1
  %t1 = load i8, ptr %slot_c_1
  %t2 = icmp ult i8 %t1, 48
  br i1 %t2, label %sc.l4, label %sc.r5
sc.l4:
  store i1 1, ptr %slot___sc_2
  br label %sc.end6
sc.r5:
  %t3 = load i8, ptr %slot_c_1
  %t4 = icmp ugt i8 %t3, 57
  store i1 %t4, ptr %slot___sc_2
  br label %sc.end6
sc.end6:
  %t5 = load i1, ptr %slot___sc_2
  br i1 %t5, label %if.then2, label %if.next3
if.then2:
  %t6 = sub i64 0, 1
  ret i64 %t6
if.next3:
  br label %if.end1
if.end1:
  %t7 = load i8, ptr %slot_c_1
  %t8 = zext i8 %t7 to i64
  %t9 = sub i64 %t8, 48
  ret i64 %t9
}
