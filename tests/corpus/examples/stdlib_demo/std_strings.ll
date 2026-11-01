; AILA module: std.strings
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.1 = private unnamed_addr constant [17 x i8] c"std/strings.aila\00", align 1

declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare i64 @aila_std_utf8__rune_len(i8)

define i64 @aila_std_strings__len_runes({ ptr, i64 } %p_t) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_b_2 = alloca { ptr, i64 }, align 8
  %slot_n_3 = alloca i64, align 8
  %slot_i_4 = alloca i64, align 8
  %slot_count_5 = alloca i64, align 8
  %slot_step_6 = alloca i64, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  %t1 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t1, ptr %slot_b_2
  %t2 = load { ptr, i64 }, ptr %slot_b_2
  %t3 = extractvalue { ptr, i64 } %t2, 1
  store i64 %t3, ptr %slot_n_3
  store i64 0, ptr %slot_i_4
  store i64 0, ptr %slot_count_5
  br label %for.cond2
for.cond2:
  %t4 = load i64, ptr %slot_i_4
  %t5 = load i64, ptr %slot_n_3
  %t6 = icmp ult i64 %t4, %t5
  br i1 %t6, label %for.body4, label %for.end1
for.body4:
  %t7 = load i64, ptr %slot_i_4
  %t8 = load { ptr, i64 }, ptr %slot_b_2
  %t9 = extractvalue { ptr, i64 } %t8, 0
  %t10 = extractvalue { ptr, i64 } %t8, 1
  %t11 = icmp uge i64 %t7, %t10
  br i1 %t11, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 20)
  unreachable
chk.ok5:
  %t12 = getelementptr inbounds i8, ptr %t9, i64 %t7
  %t13 = load i8, ptr %t12
  %t14 = call i64 @aila_std_utf8__rune_len(i8 %t13)
  store i64 %t14, ptr %slot_step_6
  %t15 = load i64, ptr %slot_step_6
  %t16 = icmp eq i64 %t15, 0
  br i1 %t16, label %if.then8, label %if.next9
if.then8:
  store i64 1, ptr %slot_step_6
  br label %if.end7
if.next9:
  br label %if.end7
if.end7:
  %t17 = load i64, ptr %slot_i_4
  %t18 = load i64, ptr %slot_step_6
  %t19 = add i64 %t17, %t18
  store i64 %t19, ptr %slot_i_4
  %t20 = load i64, ptr %slot_count_5
  %t21 = add i64 %t20, 1
  store i64 %t21, ptr %slot_count_5
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t22 = load i64, ptr %slot_count_5
  ret i64 %t22
}

define i64 @aila_std_strings__index({ ptr, i64 } %p_t, { ptr, i64 } %p_sub) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_sub_2 = alloca { ptr, i64 }, align 8
  %slot_b_3 = alloca { ptr, i64 }, align 8
  %slot_s_4 = alloca { ptr, i64 }, align 8
  %slot_n_5 = alloca i64, align 8
  %slot_m_6 = alloca i64, align 8
  %slot_i_7 = alloca i64, align 8
  %slot_j_8 = alloca i64, align 8
  %slot_ok_9 = alloca i1, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  store { ptr, i64 } %p_sub, ptr %slot_sub_2
  %t1 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t1, ptr %slot_b_3
  %t2 = load { ptr, i64 }, ptr %slot_sub_2
  store { ptr, i64 } %t2, ptr %slot_s_4
  %t3 = load { ptr, i64 }, ptr %slot_b_3
  %t4 = extractvalue { ptr, i64 } %t3, 1
  store i64 %t4, ptr %slot_n_5
  %t5 = load { ptr, i64 }, ptr %slot_s_4
  %t6 = extractvalue { ptr, i64 } %t5, 1
  store i64 %t6, ptr %slot_m_6
  %t7 = load i64, ptr %slot_m_6
  %t8 = icmp eq i64 %t7, 0
  br i1 %t8, label %if.then2, label %if.next3
if.then2:
  ret i64 0
if.next3:
  br label %if.end1
if.end1:
  %t9 = load i64, ptr %slot_m_6
  %t10 = load i64, ptr %slot_n_5
  %t11 = icmp ugt i64 %t9, %t10
  br i1 %t11, label %if.then5, label %if.next6
if.then5:
  %t12 = load i64, ptr %slot_n_5
  ret i64 %t12
if.next6:
  br label %if.end4
if.end4:
  store i64 0, ptr %slot_i_7
  br label %for.cond8
for.cond8:
  %t13 = load i64, ptr %slot_i_7
  %t14 = load i64, ptr %slot_n_5
  %t15 = load i64, ptr %slot_m_6
  %t16 = sub i64 %t14, %t15
  %t17 = icmp ule i64 %t13, %t16
  br i1 %t17, label %for.body10, label %for.end7
for.body10:
  store i64 0, ptr %slot_j_8
  store i1 1, ptr %slot_ok_9
  br label %for.cond12
for.cond12:
  %t18 = load i64, ptr %slot_j_8
  %t19 = load i64, ptr %slot_m_6
  %t20 = icmp ult i64 %t18, %t19
  br i1 %t20, label %for.body14, label %for.end11
for.body14:
  %t21 = load i64, ptr %slot_i_7
  %t22 = load i64, ptr %slot_j_8
  %t23 = add i64 %t21, %t22
  %t24 = load { ptr, i64 }, ptr %slot_b_3
  %t25 = extractvalue { ptr, i64 } %t24, 0
  %t26 = extractvalue { ptr, i64 } %t24, 1
  %t27 = icmp uge i64 %t23, %t26
  br i1 %t27, label %chk.bad19, label %chk.ok18
chk.bad19:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 48)
  unreachable
chk.ok18:
  %t28 = getelementptr inbounds i8, ptr %t25, i64 %t23
  %t29 = load i8, ptr %t28
  %t30 = load i64, ptr %slot_j_8
  %t31 = load { ptr, i64 }, ptr %slot_s_4
  %t32 = extractvalue { ptr, i64 } %t31, 0
  %t33 = extractvalue { ptr, i64 } %t31, 1
  %t34 = icmp uge i64 %t30, %t33
  br i1 %t34, label %chk.bad21, label %chk.ok20
chk.bad21:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 48)
  unreachable
chk.ok20:
  %t35 = getelementptr inbounds i8, ptr %t32, i64 %t30
  %t36 = load i8, ptr %t35
  %t37 = icmp ne i8 %t29, %t36
  br i1 %t37, label %if.then16, label %if.next17
if.then16:
  store i1 0, ptr %slot_ok_9
  br label %for.end11
if.next17:
  br label %if.end15
if.end15:
  %t38 = load i64, ptr %slot_j_8
  %t39 = add i64 %t38, 1
  store i64 %t39, ptr %slot_j_8
  br label %for.step13
for.step13:
  br label %for.cond12
for.end11:
  %t40 = load i1, ptr %slot_ok_9
  br i1 %t40, label %if.then23, label %if.next24
if.then23:
  %t41 = load i64, ptr %slot_i_7
  ret i64 %t41
if.next24:
  br label %if.end22
if.end22:
  %t42 = load i64, ptr %slot_i_7
  %t43 = add i64 %t42, 1
  store i64 %t43, ptr %slot_i_7
  br label %for.step9
for.step9:
  br label %for.cond8
for.end7:
  %t44 = load i64, ptr %slot_n_5
  ret i64 %t44
}

define i1 @aila_std_strings__contains({ ptr, i64 } %p_t, { ptr, i64 } %p_sub) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_sub_2 = alloca { ptr, i64 }, align 8
  %slot_b_3 = alloca { ptr, i64 }, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  store { ptr, i64 } %p_sub, ptr %slot_sub_2
  %t1 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t1, ptr %slot_b_3
  %t2 = load { ptr, i64 }, ptr %slot_t_1
  %t3 = load { ptr, i64 }, ptr %slot_sub_2
  %t4 = call i64 @aila_std_strings__index({ ptr, i64 } %t2, { ptr, i64 } %t3)
  %t5 = load { ptr, i64 }, ptr %slot_b_3
  %t6 = extractvalue { ptr, i64 } %t5, 1
  %t7 = icmp ne i64 %t4, %t6
  ret i1 %t7
}

define i1 @aila_std_strings__has_prefix({ ptr, i64 } %p_t, { ptr, i64 } %p_p) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_p_2 = alloca { ptr, i64 }, align 8
  %slot_b_3 = alloca { ptr, i64 }, align 8
  %slot_s_4 = alloca { ptr, i64 }, align 8
  %slot_m_5 = alloca i64, align 8
  %slot_i_6 = alloca i64, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  store { ptr, i64 } %p_p, ptr %slot_p_2
  %t1 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t1, ptr %slot_b_3
  %t2 = load { ptr, i64 }, ptr %slot_p_2
  store { ptr, i64 } %t2, ptr %slot_s_4
  %t3 = load { ptr, i64 }, ptr %slot_s_4
  %t4 = extractvalue { ptr, i64 } %t3, 1
  store i64 %t4, ptr %slot_m_5
  %t5 = load i64, ptr %slot_m_5
  %t6 = load { ptr, i64 }, ptr %slot_b_3
  %t7 = extractvalue { ptr, i64 } %t6, 1
  %t8 = icmp ugt i64 %t5, %t7
  br i1 %t8, label %if.then2, label %if.next3
if.then2:
  ret i1 0
if.next3:
  br label %if.end1
if.end1:
  store i64 0, ptr %slot_i_6
  br label %for.cond5
for.cond5:
  %t9 = load i64, ptr %slot_i_6
  %t10 = load i64, ptr %slot_m_5
  %t11 = icmp ult i64 %t9, %t10
  br i1 %t11, label %for.body7, label %for.end4
for.body7:
  %t12 = load i64, ptr %slot_i_6
  %t13 = load { ptr, i64 }, ptr %slot_b_3
  %t14 = extractvalue { ptr, i64 } %t13, 0
  %t15 = extractvalue { ptr, i64 } %t13, 1
  %t16 = icmp uge i64 %t12, %t15
  br i1 %t16, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 78)
  unreachable
chk.ok11:
  %t17 = getelementptr inbounds i8, ptr %t14, i64 %t12
  %t18 = load i8, ptr %t17
  %t19 = load i64, ptr %slot_i_6
  %t20 = load { ptr, i64 }, ptr %slot_s_4
  %t21 = extractvalue { ptr, i64 } %t20, 0
  %t22 = extractvalue { ptr, i64 } %t20, 1
  %t23 = icmp uge i64 %t19, %t22
  br i1 %t23, label %chk.bad14, label %chk.ok13
chk.bad14:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 78)
  unreachable
chk.ok13:
  %t24 = getelementptr inbounds i8, ptr %t21, i64 %t19
  %t25 = load i8, ptr %t24
  %t26 = icmp ne i8 %t18, %t25
  br i1 %t26, label %if.then9, label %if.next10
if.then9:
  ret i1 0
if.next10:
  br label %if.end8
if.end8:
  %t27 = load i64, ptr %slot_i_6
  %t28 = add i64 %t27, 1
  store i64 %t28, ptr %slot_i_6
  br label %for.step6
for.step6:
  br label %for.cond5
for.end4:
  ret i1 1
}

define i1 @aila_std_strings__has_suffix({ ptr, i64 } %p_t, { ptr, i64 } %p_s) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_s_2 = alloca { ptr, i64 }, align 8
  %slot_b_3 = alloca { ptr, i64 }, align 8
  %slot_x_4 = alloca { ptr, i64 }, align 8
  %slot_n_5 = alloca i64, align 8
  %slot_m_6 = alloca i64, align 8
  %slot_i_7 = alloca i64, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  store { ptr, i64 } %p_s, ptr %slot_s_2
  %t1 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t1, ptr %slot_b_3
  %t2 = load { ptr, i64 }, ptr %slot_s_2
  store { ptr, i64 } %t2, ptr %slot_x_4
  %t3 = load { ptr, i64 }, ptr %slot_b_3
  %t4 = extractvalue { ptr, i64 } %t3, 1
  store i64 %t4, ptr %slot_n_5
  %t5 = load { ptr, i64 }, ptr %slot_x_4
  %t6 = extractvalue { ptr, i64 } %t5, 1
  store i64 %t6, ptr %slot_m_6
  %t7 = load i64, ptr %slot_m_6
  %t8 = load i64, ptr %slot_n_5
  %t9 = icmp ugt i64 %t7, %t8
  br i1 %t9, label %if.then2, label %if.next3
if.then2:
  ret i1 0
if.next3:
  br label %if.end1
if.end1:
  store i64 0, ptr %slot_i_7
  br label %for.cond5
for.cond5:
  %t10 = load i64, ptr %slot_i_7
  %t11 = load i64, ptr %slot_m_6
  %t12 = icmp ult i64 %t10, %t11
  br i1 %t12, label %for.body7, label %for.end4
for.body7:
  %t13 = load i64, ptr %slot_n_5
  %t14 = load i64, ptr %slot_m_6
  %t15 = sub i64 %t13, %t14
  %t16 = load i64, ptr %slot_i_7
  %t17 = add i64 %t15, %t16
  %t18 = load { ptr, i64 }, ptr %slot_b_3
  %t19 = extractvalue { ptr, i64 } %t18, 0
  %t20 = extractvalue { ptr, i64 } %t18, 1
  %t21 = icmp uge i64 %t17, %t20
  br i1 %t21, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 97)
  unreachable
chk.ok11:
  %t22 = getelementptr inbounds i8, ptr %t19, i64 %t17
  %t23 = load i8, ptr %t22
  %t24 = load i64, ptr %slot_i_7
  %t25 = load { ptr, i64 }, ptr %slot_x_4
  %t26 = extractvalue { ptr, i64 } %t25, 0
  %t27 = extractvalue { ptr, i64 } %t25, 1
  %t28 = icmp uge i64 %t24, %t27
  br i1 %t28, label %chk.bad14, label %chk.ok13
chk.bad14:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 97)
  unreachable
chk.ok13:
  %t29 = getelementptr inbounds i8, ptr %t26, i64 %t24
  %t30 = load i8, ptr %t29
  %t31 = icmp ne i8 %t23, %t30
  br i1 %t31, label %if.then9, label %if.next10
if.then9:
  ret i1 0
if.next10:
  br label %if.end8
if.end8:
  %t32 = load i64, ptr %slot_i_7
  %t33 = add i64 %t32, 1
  store i64 %t33, ptr %slot_i_7
  br label %for.step6
for.step6:
  br label %for.cond5
for.end4:
  ret i1 1
}

define { i64, i64 } @aila_std_strings__bounds({ ptr, i64 } %p_t) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_b_2 = alloca { ptr, i64 }, align 8
  %slot_n_3 = alloca i64, align 8
  %slot_start_4 = alloca i64, align 8
  %slot_c_5 = alloca i8, align 8
  %slot_end_6 = alloca i64, align 8
  %slot_c_7 = alloca i8, align 8
  %slot___sc_8 = alloca i1, align 8
  %slot___sc_9 = alloca i1, align 8
  %slot___sc_10 = alloca i1, align 8
  %slot___sc_11 = alloca i1, align 8
  %slot___sc_12 = alloca i1, align 8
  %slot___sc_13 = alloca i1, align 8
  %slot___sc_14 = alloca i1, align 8
  %slot___sc_15 = alloca i1, align 8
  %slot___sc_16 = alloca i1, align 8
  %slot___sc_17 = alloca i1, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  %t1 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t1, ptr %slot_b_2
  %t2 = load { ptr, i64 }, ptr %slot_b_2
  %t3 = extractvalue { ptr, i64 } %t2, 1
  store i64 %t3, ptr %slot_n_3
  store i64 0, ptr %slot_start_4
  br label %for.cond2
for.cond2:
  %t4 = load i64, ptr %slot_start_4
  %t5 = load i64, ptr %slot_n_3
  %t6 = icmp ult i64 %t4, %t5
  br i1 %t6, label %for.body4, label %for.end1
for.body4:
  %t7 = load i64, ptr %slot_start_4
  %t8 = load { ptr, i64 }, ptr %slot_b_2
  %t9 = extractvalue { ptr, i64 } %t8, 0
  %t10 = extractvalue { ptr, i64 } %t8, 1
  %t11 = icmp uge i64 %t7, %t10
  br i1 %t11, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 112)
  unreachable
chk.ok5:
  %t12 = getelementptr inbounds i8, ptr %t9, i64 %t7
  %t13 = load i8, ptr %t12
  store i8 %t13, ptr %slot_c_5
  %t14 = load i8, ptr %slot_c_5
  %t15 = icmp eq i8 %t14, 9
  br i1 %t15, label %sc.l22, label %sc.r23
sc.l22:
  store i1 1, ptr %slot___sc_12
  br label %sc.end24
sc.r23:
  %t16 = load i8, ptr %slot_c_5
  %t17 = icmp eq i8 %t16, 10
  store i1 %t17, ptr %slot___sc_12
  br label %sc.end24
sc.end24:
  %t18 = load i1, ptr %slot___sc_12
  br i1 %t18, label %sc.l19, label %sc.r20
sc.l19:
  store i1 1, ptr %slot___sc_11
  br label %sc.end21
sc.r20:
  %t19 = load i8, ptr %slot_c_5
  %t20 = icmp eq i8 %t19, 11
  store i1 %t20, ptr %slot___sc_11
  br label %sc.end21
sc.end21:
  %t21 = load i1, ptr %slot___sc_11
  br i1 %t21, label %sc.l16, label %sc.r17
sc.l16:
  store i1 1, ptr %slot___sc_10
  br label %sc.end18
sc.r17:
  %t22 = load i8, ptr %slot_c_5
  %t23 = icmp eq i8 %t22, 12
  store i1 %t23, ptr %slot___sc_10
  br label %sc.end18
sc.end18:
  %t24 = load i1, ptr %slot___sc_10
  br i1 %t24, label %sc.l13, label %sc.r14
sc.l13:
  store i1 1, ptr %slot___sc_9
  br label %sc.end15
sc.r14:
  %t25 = load i8, ptr %slot_c_5
  %t26 = icmp eq i8 %t25, 13
  store i1 %t26, ptr %slot___sc_9
  br label %sc.end15
sc.end15:
  %t27 = load i1, ptr %slot___sc_9
  br i1 %t27, label %sc.l10, label %sc.r11
sc.l10:
  store i1 1, ptr %slot___sc_8
  br label %sc.end12
sc.r11:
  %t28 = load i8, ptr %slot_c_5
  %t29 = icmp eq i8 %t28, 32
  store i1 %t29, ptr %slot___sc_8
  br label %sc.end12
sc.end12:
  %t30 = load i1, ptr %slot___sc_8
  br i1 %t30, label %if.then8, label %if.next9
if.then8:
  %t31 = load i64, ptr %slot_start_4
  %t32 = add i64 %t31, 1
  store i64 %t32, ptr %slot_start_4
  br label %if.end7
if.next9:
  br label %for.end1
if.end7:
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t33 = load i64, ptr %slot_n_3
  store i64 %t33, ptr %slot_end_6
  br label %for.cond26
for.cond26:
  %t34 = load i64, ptr %slot_end_6
  %t35 = load i64, ptr %slot_start_4
  %t36 = icmp ugt i64 %t34, %t35
  br i1 %t36, label %for.body28, label %for.end25
for.body28:
  %t37 = load i64, ptr %slot_end_6
  %t38 = sub i64 %t37, 1
  %t39 = load { ptr, i64 }, ptr %slot_b_2
  %t40 = extractvalue { ptr, i64 } %t39, 0
  %t41 = extractvalue { ptr, i64 } %t39, 1
  %t42 = icmp uge i64 %t38, %t41
  br i1 %t42, label %chk.bad30, label %chk.ok29
chk.bad30:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 121)
  unreachable
chk.ok29:
  %t43 = getelementptr inbounds i8, ptr %t40, i64 %t38
  %t44 = load i8, ptr %t43
  store i8 %t44, ptr %slot_c_7
  %t45 = load i8, ptr %slot_c_7
  %t46 = icmp eq i8 %t45, 9
  br i1 %t46, label %sc.l46, label %sc.r47
sc.l46:
  store i1 1, ptr %slot___sc_17
  br label %sc.end48
sc.r47:
  %t47 = load i8, ptr %slot_c_7
  %t48 = icmp eq i8 %t47, 10
  store i1 %t48, ptr %slot___sc_17
  br label %sc.end48
sc.end48:
  %t49 = load i1, ptr %slot___sc_17
  br i1 %t49, label %sc.l43, label %sc.r44
sc.l43:
  store i1 1, ptr %slot___sc_16
  br label %sc.end45
sc.r44:
  %t50 = load i8, ptr %slot_c_7
  %t51 = icmp eq i8 %t50, 11
  store i1 %t51, ptr %slot___sc_16
  br label %sc.end45
sc.end45:
  %t52 = load i1, ptr %slot___sc_16
  br i1 %t52, label %sc.l40, label %sc.r41
sc.l40:
  store i1 1, ptr %slot___sc_15
  br label %sc.end42
sc.r41:
  %t53 = load i8, ptr %slot_c_7
  %t54 = icmp eq i8 %t53, 12
  store i1 %t54, ptr %slot___sc_15
  br label %sc.end42
sc.end42:
  %t55 = load i1, ptr %slot___sc_15
  br i1 %t55, label %sc.l37, label %sc.r38
sc.l37:
  store i1 1, ptr %slot___sc_14
  br label %sc.end39
sc.r38:
  %t56 = load i8, ptr %slot_c_7
  %t57 = icmp eq i8 %t56, 13
  store i1 %t57, ptr %slot___sc_14
  br label %sc.end39
sc.end39:
  %t58 = load i1, ptr %slot___sc_14
  br i1 %t58, label %sc.l34, label %sc.r35
sc.l34:
  store i1 1, ptr %slot___sc_13
  br label %sc.end36
sc.r35:
  %t59 = load i8, ptr %slot_c_7
  %t60 = icmp eq i8 %t59, 32
  store i1 %t60, ptr %slot___sc_13
  br label %sc.end36
sc.end36:
  %t61 = load i1, ptr %slot___sc_13
  br i1 %t61, label %if.then32, label %if.next33
if.then32:
  %t62 = load i64, ptr %slot_end_6
  %t63 = sub i64 %t62, 1
  store i64 %t63, ptr %slot_end_6
  br label %if.end31
if.next33:
  br label %for.end25
if.end31:
  br label %for.step27
for.step27:
  br label %for.cond26
for.end25:
  %t64 = load i64, ptr %slot_start_4
  %t65 = load i64, ptr %slot_end_6
  %t66 = icmp eq i64 %t64, %t65
  br i1 %t66, label %if.then50, label %if.next51
if.then50:
  %t67 = insertvalue { i64, i64 } undef, i64 0, 0
  %t68 = insertvalue { i64, i64 } %t67, i64 0, 1
  ret { i64, i64 } %t68
if.next51:
  br label %if.end49
if.end49:
  %t69 = load i64, ptr %slot_start_4
  %t70 = load i64, ptr %slot_end_6
  %t71 = insertvalue { i64, i64 } undef, i64 %t69, 0
  %t72 = insertvalue { i64, i64 } %t71, i64 %t70, 1
  ret { i64, i64 } %t72
}

define i64 @aila_std_strings__concat({ ptr, i64 } %p_a, { ptr, i64 } %p_b, { ptr, i64 } %p_buf) {
  %slot_a_1 = alloca { ptr, i64 }, align 8
  %slot_b_2 = alloca { ptr, i64 }, align 8
  %slot_buf_3 = alloca { ptr, i64 }, align 8
  %slot_cap_4 = alloca i64, align 8
  %slot_w_5 = alloca i64, align 8
  %slot_ab_6 = alloca { ptr, i64 }, align 8
  %slot_bb_7 = alloca { ptr, i64 }, align 8
  store { ptr, i64 } %p_a, ptr %slot_a_1
  store { ptr, i64 } %p_b, ptr %slot_b_2
  store { ptr, i64 } %p_buf, ptr %slot_buf_3
  %t1 = load { ptr, i64 }, ptr %slot_buf_3
  %t2 = extractvalue { ptr, i64 } %t1, 1
  store i64 %t2, ptr %slot_cap_4
  store i64 0, ptr %slot_w_5
  %t3 = load { ptr, i64 }, ptr %slot_a_1
  store { ptr, i64 } %t3, ptr %slot_ab_6
  %t4 = load { ptr, i64 }, ptr %slot_buf_3
  %t5 = load i64, ptr %slot_w_5
  %t6 = load i64, ptr %slot_cap_4
  %t7 = load { ptr, i64 }, ptr %slot_ab_6
  %t8 = call i64 @aila_std_strings__blit({ ptr, i64 } %t4, i64 %t5, i64 %t6, { ptr, i64 } %t7)
  store i64 %t8, ptr %slot_w_5
  %t9 = load { ptr, i64 }, ptr %slot_b_2
  store { ptr, i64 } %t9, ptr %slot_bb_7
  %t10 = load { ptr, i64 }, ptr %slot_buf_3
  %t11 = load i64, ptr %slot_w_5
  %t12 = load i64, ptr %slot_cap_4
  %t13 = load { ptr, i64 }, ptr %slot_bb_7
  %t14 = call i64 @aila_std_strings__blit({ ptr, i64 } %t10, i64 %t11, i64 %t12, { ptr, i64 } %t13)
  store i64 %t14, ptr %slot_w_5
  %t15 = load i64, ptr %slot_w_5
  ret i64 %t15
}

define i64 @aila_std_strings__split({ ptr, i64 } %p_t, i8 %p_sep, { ptr, i64 } %p_marks) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_sep_2 = alloca i8, align 8
  %slot_marks_3 = alloca { ptr, i64 }, align 8
  %slot_b_4 = alloca { ptr, i64 }, align 8
  %slot_n_5 = alloca i64, align 8
  %slot_cap_6 = alloca i64, align 8
  %slot_count_7 = alloca i64, align 8
  %slot_i_8 = alloca i64, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  store i8 %p_sep, ptr %slot_sep_2
  store { ptr, i64 } %p_marks, ptr %slot_marks_3
  %t1 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t1, ptr %slot_b_4
  %t2 = load { ptr, i64 }, ptr %slot_b_4
  %t3 = extractvalue { ptr, i64 } %t2, 1
  store i64 %t3, ptr %slot_n_5
  %t4 = load { ptr, i64 }, ptr %slot_marks_3
  %t5 = extractvalue { ptr, i64 } %t4, 1
  store i64 %t5, ptr %slot_cap_6
  %t6 = load i64, ptr %slot_cap_6
  %t7 = icmp ugt i64 %t6, 0
  br i1 %t7, label %if.then2, label %if.next3
if.then2:
  %t8 = load { ptr, i64 }, ptr %slot_marks_3
  %t9 = extractvalue { ptr, i64 } %t8, 0
  %t10 = extractvalue { ptr, i64 } %t8, 1
  %t11 = icmp uge i64 0, %t10
  br i1 %t11, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 155)
  unreachable
chk.ok4:
  %t12 = getelementptr inbounds i64, ptr %t9, i64 0
  store i64 0, ptr %t12
  br label %if.end1
if.next3:
  br label %if.end1
if.end1:
  store i64 1, ptr %slot_count_7
  store i64 0, ptr %slot_i_8
  br label %for.cond7
for.cond7:
  %t13 = load i64, ptr %slot_i_8
  %t14 = load i64, ptr %slot_n_5
  %t15 = icmp ult i64 %t13, %t14
  br i1 %t15, label %for.body9, label %for.end6
for.body9:
  %t16 = load i64, ptr %slot_i_8
  %t17 = load { ptr, i64 }, ptr %slot_b_4
  %t18 = extractvalue { ptr, i64 } %t17, 0
  %t19 = extractvalue { ptr, i64 } %t17, 1
  %t20 = icmp uge i64 %t16, %t19
  br i1 %t20, label %chk.bad14, label %chk.ok13
chk.bad14:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 160)
  unreachable
chk.ok13:
  %t21 = getelementptr inbounds i8, ptr %t18, i64 %t16
  %t22 = load i8, ptr %t21
  %t23 = load i8, ptr %slot_sep_2
  %t24 = icmp eq i8 %t22, %t23
  br i1 %t24, label %if.then11, label %if.next12
if.then11:
  %t25 = load i64, ptr %slot_count_7
  %t26 = load i64, ptr %slot_cap_6
  %t27 = icmp ult i64 %t25, %t26
  br i1 %t27, label %if.then16, label %if.next17
if.then16:
  %t28 = load i64, ptr %slot_count_7
  %t29 = load { ptr, i64 }, ptr %slot_marks_3
  %t30 = extractvalue { ptr, i64 } %t29, 0
  %t31 = extractvalue { ptr, i64 } %t29, 1
  %t32 = icmp uge i64 %t28, %t31
  br i1 %t32, label %chk.bad19, label %chk.ok18
chk.bad19:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 162)
  unreachable
chk.ok18:
  %t33 = getelementptr inbounds i64, ptr %t30, i64 %t28
  %t34 = load i64, ptr %slot_i_8
  %t35 = add i64 %t34, 1
  store i64 %t35, ptr %t33
  br label %if.end15
if.next17:
  br label %if.end15
if.end15:
  %t36 = load i64, ptr %slot_count_7
  %t37 = add i64 %t36, 1
  store i64 %t37, ptr %slot_count_7
  br label %if.end10
if.next12:
  br label %if.end10
if.end10:
  %t38 = load i64, ptr %slot_i_8
  %t39 = add i64 %t38, 1
  store i64 %t39, ptr %slot_i_8
  br label %for.step8
for.step8:
  br label %for.cond7
for.end6:
  %t40 = load i64, ptr %slot_count_7
  ret i64 %t40
}

define i64 @aila_std_strings__join({ ptr, i64 } %p_buf, i8 %p_sep, { ptr, i64 } %p_parts) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_sep_2 = alloca i8, align 8
  %slot_parts_3 = alloca { ptr, i64 }, align 8
  %slot_cap_4 = alloca i64, align 8
  %slot_w_5 = alloca i64, align 8
  %slot_first_6 = alloca i1, align 8
  %slot_p_7 = alloca { ptr, i64 }, align 8
  %slot_pb_8 = alloca { ptr, i64 }, align 8
  %slot___idx_9 = alloca i64, align 8
  %slot___iter_10 = alloca { ptr, i64 }, align 8
  %slot___len_11 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i8 %p_sep, ptr %slot_sep_2
  store { ptr, i64 } %p_parts, ptr %slot_parts_3
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = extractvalue { ptr, i64 } %t1, 1
  store i64 %t2, ptr %slot_cap_4
  store i64 0, ptr %slot_w_5
  store i1 1, ptr %slot_first_6
  store i64 0, ptr %slot___idx_9
  %t3 = load { ptr, i64 }, ptr %slot_parts_3
  store { ptr, i64 } %t3, ptr %slot___iter_10
  %t5 = getelementptr inbounds { ptr, i64 }, ptr %slot___iter_10, i32 0, i32 1
  %t4 = load i64, ptr %t5
  store i64 %t4, ptr %slot___len_11
  %t6 = load ptr, ptr %slot___iter_10
  br label %fi.cond2
fi.cond2:
  %t7 = load i64, ptr %slot___idx_9
  %t8 = load i64, ptr %slot___len_11
  %t9 = icmp ult i64 %t7, %t8
  br i1 %t9, label %fi.body4, label %for.end1
fi.body4:
  %t10 = getelementptr inbounds { ptr, i64 }, ptr %t6, i64 %t7
  %t11 = load { ptr, i64 }, ptr %t10
  store { ptr, i64 } %t11, ptr %slot_p_7
  %t12 = load i1, ptr %slot_first_6
  %t13 = xor i1 %t12, true
  br i1 %t13, label %if.then6, label %if.next7
if.then6:
  %t14 = load i64, ptr %slot_w_5
  %t15 = load i64, ptr %slot_cap_4
  %t16 = icmp ult i64 %t14, %t15
  br i1 %t16, label %if.then9, label %if.next10
if.then9:
  %t17 = load i64, ptr %slot_w_5
  %t18 = load { ptr, i64 }, ptr %slot_buf_1
  %t19 = extractvalue { ptr, i64 } %t18, 0
  %t20 = extractvalue { ptr, i64 } %t18, 1
  %t21 = icmp uge i64 %t17, %t20
  br i1 %t21, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 180)
  unreachable
chk.ok11:
  %t22 = getelementptr inbounds i8, ptr %t19, i64 %t17
  %t23 = load i8, ptr %slot_sep_2
  store i8 %t23, ptr %t22
  %t24 = load i64, ptr %slot_w_5
  %t25 = add i64 %t24, 1
  store i64 %t25, ptr %slot_w_5
  br label %if.end8
if.next10:
  br label %if.end8
if.end8:
  br label %if.end5
if.next7:
  br label %if.end5
if.end5:
  store i1 0, ptr %slot_first_6
  %t26 = load { ptr, i64 }, ptr %slot_p_7
  store { ptr, i64 } %t26, ptr %slot_pb_8
  %t27 = load { ptr, i64 }, ptr %slot_buf_1
  %t28 = load i64, ptr %slot_w_5
  %t29 = load i64, ptr %slot_cap_4
  %t30 = load { ptr, i64 }, ptr %slot_pb_8
  %t31 = call i64 @aila_std_strings__blit({ ptr, i64 } %t27, i64 %t28, i64 %t29, { ptr, i64 } %t30)
  store i64 %t31, ptr %slot_w_5
  br label %fi.step3
fi.step3:
  %t32 = add i64 %t7, 1
  store i64 %t32, ptr %slot___idx_9
  br label %fi.cond2
for.end1:
  %t33 = load i64, ptr %slot_w_5
  ret i64 %t33
}

define i64 @aila_std_strings__replace({ ptr, i64 } %p_t, { ptr, i64 } %p_from, { ptr, i64 } %p_to, { ptr, i64 } %p_buf) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_from_2 = alloca { ptr, i64 }, align 8
  %slot_to_3 = alloca { ptr, i64 }, align 8
  %slot_buf_4 = alloca { ptr, i64 }, align 8
  %slot_cap_5 = alloca i64, align 8
  %slot_b_6 = alloca { ptr, i64 }, align 8
  %slot_f_7 = alloca { ptr, i64 }, align 8
  %slot_n_8 = alloca i64, align 8
  %slot_m_9 = alloca i64, align 8
  %slot_w_10 = alloca i64, align 8
  %slot_i_11 = alloca i64, align 8
  %slot_ok_12 = alloca i1, align 8
  %slot_j_13 = alloca i64, align 8
  %slot_tb_14 = alloca { ptr, i64 }, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  store { ptr, i64 } %p_from, ptr %slot_from_2
  store { ptr, i64 } %p_to, ptr %slot_to_3
  store { ptr, i64 } %p_buf, ptr %slot_buf_4
  %t1 = load { ptr, i64 }, ptr %slot_buf_4
  %t2 = extractvalue { ptr, i64 } %t1, 1
  store i64 %t2, ptr %slot_cap_5
  %t3 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t3, ptr %slot_b_6
  %t4 = load { ptr, i64 }, ptr %slot_from_2
  store { ptr, i64 } %t4, ptr %slot_f_7
  %t5 = load { ptr, i64 }, ptr %slot_b_6
  %t6 = extractvalue { ptr, i64 } %t5, 1
  store i64 %t6, ptr %slot_n_8
  %t7 = load { ptr, i64 }, ptr %slot_f_7
  %t8 = extractvalue { ptr, i64 } %t7, 1
  store i64 %t8, ptr %slot_m_9
  store i64 0, ptr %slot_w_10
  %t9 = load i64, ptr %slot_m_9
  %t10 = icmp eq i64 %t9, 0
  br i1 %t10, label %if.then2, label %if.next3
if.then2:
  %t11 = load { ptr, i64 }, ptr %slot_buf_4
  %t12 = load i64, ptr %slot_w_10
  %t13 = load i64, ptr %slot_cap_5
  %t14 = load { ptr, i64 }, ptr %slot_b_6
  %t15 = call i64 @aila_std_strings__blit({ ptr, i64 } %t11, i64 %t12, i64 %t13, { ptr, i64 } %t14)
  store i64 %t15, ptr %slot_w_10
  %t16 = load i64, ptr %slot_w_10
  ret i64 %t16
if.next3:
  br label %if.end1
if.end1:
  store i64 0, ptr %slot_i_11
  br label %for.cond5
for.cond5:
  %t17 = load i64, ptr %slot_i_11
  %t18 = load i64, ptr %slot_n_8
  %t19 = icmp ult i64 %t17, %t18
  br i1 %t19, label %for.body7, label %for.end4
for.body7:
  store i1 0, ptr %slot_ok_12
  %t20 = load i64, ptr %slot_i_11
  %t21 = load i64, ptr %slot_m_9
  %t22 = add i64 %t20, %t21
  %t23 = load i64, ptr %slot_n_8
  %t24 = icmp ule i64 %t22, %t23
  br i1 %t24, label %if.then9, label %if.next10
if.then9:
  store i1 1, ptr %slot_ok_12
  store i64 0, ptr %slot_j_13
  br label %for.cond12
for.cond12:
  %t25 = load i64, ptr %slot_j_13
  %t26 = load i64, ptr %slot_m_9
  %t27 = icmp ult i64 %t25, %t26
  br i1 %t27, label %for.body14, label %for.end11
for.body14:
  %t28 = load i64, ptr %slot_i_11
  %t29 = load i64, ptr %slot_j_13
  %t30 = add i64 %t28, %t29
  %t31 = load { ptr, i64 }, ptr %slot_b_6
  %t32 = extractvalue { ptr, i64 } %t31, 0
  %t33 = extractvalue { ptr, i64 } %t31, 1
  %t34 = icmp uge i64 %t30, %t33
  br i1 %t34, label %chk.bad19, label %chk.ok18
chk.bad19:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 211)
  unreachable
chk.ok18:
  %t35 = getelementptr inbounds i8, ptr %t32, i64 %t30
  %t36 = load i8, ptr %t35
  %t37 = load i64, ptr %slot_j_13
  %t38 = load { ptr, i64 }, ptr %slot_f_7
  %t39 = extractvalue { ptr, i64 } %t38, 0
  %t40 = extractvalue { ptr, i64 } %t38, 1
  %t41 = icmp uge i64 %t37, %t40
  br i1 %t41, label %chk.bad21, label %chk.ok20
chk.bad21:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 211)
  unreachable
chk.ok20:
  %t42 = getelementptr inbounds i8, ptr %t39, i64 %t37
  %t43 = load i8, ptr %t42
  %t44 = icmp ne i8 %t36, %t43
  br i1 %t44, label %if.then16, label %if.next17
if.then16:
  store i1 0, ptr %slot_ok_12
  br label %for.end11
if.next17:
  br label %if.end15
if.end15:
  %t45 = load i64, ptr %slot_j_13
  %t46 = add i64 %t45, 1
  store i64 %t46, ptr %slot_j_13
  br label %for.step13
for.step13:
  br label %for.cond12
for.end11:
  br label %if.end8
if.next10:
  br label %if.end8
if.end8:
  %t47 = load i1, ptr %slot_ok_12
  br i1 %t47, label %if.then23, label %if.next24
if.then23:
  %t48 = load { ptr, i64 }, ptr %slot_to_3
  store { ptr, i64 } %t48, ptr %slot_tb_14
  %t49 = load { ptr, i64 }, ptr %slot_buf_4
  %t50 = load i64, ptr %slot_w_10
  %t51 = load i64, ptr %slot_cap_5
  %t52 = load { ptr, i64 }, ptr %slot_tb_14
  %t53 = call i64 @aila_std_strings__blit({ ptr, i64 } %t49, i64 %t50, i64 %t51, { ptr, i64 } %t52)
  store i64 %t53, ptr %slot_w_10
  %t54 = load i64, ptr %slot_i_11
  %t55 = load i64, ptr %slot_m_9
  %t56 = add i64 %t54, %t55
  store i64 %t56, ptr %slot_i_11
  br label %if.end22
if.next24:
  %t57 = load i64, ptr %slot_w_10
  %t58 = load i64, ptr %slot_cap_5
  %t59 = icmp ult i64 %t57, %t58
  br i1 %t59, label %if.then26, label %if.next27
if.then26:
  %t60 = load i64, ptr %slot_w_10
  %t61 = load { ptr, i64 }, ptr %slot_buf_4
  %t62 = extractvalue { ptr, i64 } %t61, 0
  %t63 = extractvalue { ptr, i64 } %t61, 1
  %t64 = icmp uge i64 %t60, %t63
  br i1 %t64, label %chk.bad29, label %chk.ok28
chk.bad29:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 224)
  unreachable
chk.ok28:
  %t65 = getelementptr inbounds i8, ptr %t62, i64 %t60
  %t66 = load i64, ptr %slot_i_11
  %t67 = load { ptr, i64 }, ptr %slot_b_6
  %t68 = extractvalue { ptr, i64 } %t67, 0
  %t69 = extractvalue { ptr, i64 } %t67, 1
  %t70 = icmp uge i64 %t66, %t69
  br i1 %t70, label %chk.bad31, label %chk.ok30
chk.bad31:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 224)
  unreachable
chk.ok30:
  %t71 = getelementptr inbounds i8, ptr %t68, i64 %t66
  %t72 = load i8, ptr %t71
  store i8 %t72, ptr %t65
  %t73 = load i64, ptr %slot_w_10
  %t74 = add i64 %t73, 1
  store i64 %t74, ptr %slot_w_10
  br label %if.end25
if.next27:
  br label %if.end25
if.end25:
  %t75 = load i64, ptr %slot_i_11
  %t76 = add i64 %t75, 1
  store i64 %t76, ptr %slot_i_11
  br label %if.end22
if.end22:
  br label %for.step6
for.step6:
  br label %for.cond5
for.end4:
  %t77 = load i64, ptr %slot_w_10
  ret i64 %t77
}

define i64 @aila_std_strings__upper_ascii({ ptr, i64 } %p_t, { ptr, i64 } %p_buf) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_cap_3 = alloca i64, align 8
  %slot_b_4 = alloca { ptr, i64 }, align 8
  %slot_n_5 = alloca i64, align 8
  %slot_w_6 = alloca i64, align 8
  %slot_i_7 = alloca i64, align 8
  %slot_c_8 = alloca i8, align 8
  %slot___sc_9 = alloca i1, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  store { ptr, i64 } %p_buf, ptr %slot_buf_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_2
  %t2 = extractvalue { ptr, i64 } %t1, 1
  store i64 %t2, ptr %slot_cap_3
  %t3 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t3, ptr %slot_b_4
  %t4 = load { ptr, i64 }, ptr %slot_b_4
  %t5 = extractvalue { ptr, i64 } %t4, 1
  store i64 %t5, ptr %slot_n_5
  store i64 0, ptr %slot_w_6
  store i64 0, ptr %slot_i_7
  br label %for.cond2
for.cond2:
  %t6 = load i64, ptr %slot_i_7
  %t7 = load i64, ptr %slot_n_5
  %t8 = icmp ult i64 %t6, %t7
  br i1 %t8, label %for.body4, label %for.end1
for.body4:
  %t9 = load i64, ptr %slot_w_6
  %t10 = load i64, ptr %slot_cap_3
  %t11 = icmp eq i64 %t9, %t10
  br i1 %t11, label %if.then6, label %if.next7
if.then6:
  %t12 = load i64, ptr %slot_w_6
  ret i64 %t12
if.next7:
  br label %if.end5
if.end5:
  %t13 = load i64, ptr %slot_i_7
  %t14 = load { ptr, i64 }, ptr %slot_b_4
  %t15 = extractvalue { ptr, i64 } %t14, 0
  %t16 = extractvalue { ptr, i64 } %t14, 1
  %t17 = icmp uge i64 %t13, %t16
  br i1 %t17, label %chk.bad9, label %chk.ok8
chk.bad9:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 244)
  unreachable
chk.ok8:
  %t18 = getelementptr inbounds i8, ptr %t15, i64 %t13
  %t19 = load i8, ptr %t18
  store i8 %t19, ptr %slot_c_8
  %t20 = load i8, ptr %slot_c_8
  %t21 = icmp uge i8 %t20, 97
  br i1 %t21, label %sc.r14, label %sc.l13
sc.l13:
  store i1 0, ptr %slot___sc_9
  br label %sc.end15
sc.r14:
  %t22 = load i8, ptr %slot_c_8
  %t23 = icmp ule i8 %t22, 122
  store i1 %t23, ptr %slot___sc_9
  br label %sc.end15
sc.end15:
  %t24 = load i1, ptr %slot___sc_9
  br i1 %t24, label %if.then11, label %if.next12
if.then11:
  %t25 = load i8, ptr %slot_c_8
  %t26 = sub i8 %t25, 32
  store i8 %t26, ptr %slot_c_8
  br label %if.end10
if.next12:
  br label %if.end10
if.end10:
  %t27 = load i64, ptr %slot_w_6
  %t28 = load { ptr, i64 }, ptr %slot_buf_2
  %t29 = extractvalue { ptr, i64 } %t28, 0
  %t30 = extractvalue { ptr, i64 } %t28, 1
  %t31 = icmp uge i64 %t27, %t30
  br i1 %t31, label %chk.bad17, label %chk.ok16
chk.bad17:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 248)
  unreachable
chk.ok16:
  %t32 = getelementptr inbounds i8, ptr %t29, i64 %t27
  %t33 = load i8, ptr %slot_c_8
  store i8 %t33, ptr %t32
  %t34 = load i64, ptr %slot_w_6
  %t35 = add i64 %t34, 1
  store i64 %t35, ptr %slot_w_6
  %t36 = load i64, ptr %slot_i_7
  %t37 = add i64 %t36, 1
  store i64 %t37, ptr %slot_i_7
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t38 = load i64, ptr %slot_w_6
  ret i64 %t38
}

define i64 @aila_std_strings__lower_ascii({ ptr, i64 } %p_t, { ptr, i64 } %p_buf) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_cap_3 = alloca i64, align 8
  %slot_b_4 = alloca { ptr, i64 }, align 8
  %slot_n_5 = alloca i64, align 8
  %slot_w_6 = alloca i64, align 8
  %slot_i_7 = alloca i64, align 8
  %slot_c_8 = alloca i8, align 8
  %slot___sc_9 = alloca i1, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  store { ptr, i64 } %p_buf, ptr %slot_buf_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_2
  %t2 = extractvalue { ptr, i64 } %t1, 1
  store i64 %t2, ptr %slot_cap_3
  %t3 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t3, ptr %slot_b_4
  %t4 = load { ptr, i64 }, ptr %slot_b_4
  %t5 = extractvalue { ptr, i64 } %t4, 1
  store i64 %t5, ptr %slot_n_5
  store i64 0, ptr %slot_w_6
  store i64 0, ptr %slot_i_7
  br label %for.cond2
for.cond2:
  %t6 = load i64, ptr %slot_i_7
  %t7 = load i64, ptr %slot_n_5
  %t8 = icmp ult i64 %t6, %t7
  br i1 %t8, label %for.body4, label %for.end1
for.body4:
  %t9 = load i64, ptr %slot_w_6
  %t10 = load i64, ptr %slot_cap_3
  %t11 = icmp eq i64 %t9, %t10
  br i1 %t11, label %if.then6, label %if.next7
if.then6:
  %t12 = load i64, ptr %slot_w_6
  ret i64 %t12
if.next7:
  br label %if.end5
if.end5:
  %t13 = load i64, ptr %slot_i_7
  %t14 = load { ptr, i64 }, ptr %slot_b_4
  %t15 = extractvalue { ptr, i64 } %t14, 0
  %t16 = extractvalue { ptr, i64 } %t14, 1
  %t17 = icmp uge i64 %t13, %t16
  br i1 %t17, label %chk.bad9, label %chk.ok8
chk.bad9:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 266)
  unreachable
chk.ok8:
  %t18 = getelementptr inbounds i8, ptr %t15, i64 %t13
  %t19 = load i8, ptr %t18
  store i8 %t19, ptr %slot_c_8
  %t20 = load i8, ptr %slot_c_8
  %t21 = icmp uge i8 %t20, 65
  br i1 %t21, label %sc.r14, label %sc.l13
sc.l13:
  store i1 0, ptr %slot___sc_9
  br label %sc.end15
sc.r14:
  %t22 = load i8, ptr %slot_c_8
  %t23 = icmp ule i8 %t22, 90
  store i1 %t23, ptr %slot___sc_9
  br label %sc.end15
sc.end15:
  %t24 = load i1, ptr %slot___sc_9
  br i1 %t24, label %if.then11, label %if.next12
if.then11:
  %t25 = load i8, ptr %slot_c_8
  %t26 = add i8 %t25, 32
  store i8 %t26, ptr %slot_c_8
  br label %if.end10
if.next12:
  br label %if.end10
if.end10:
  %t27 = load i64, ptr %slot_w_6
  %t28 = load { ptr, i64 }, ptr %slot_buf_2
  %t29 = extractvalue { ptr, i64 } %t28, 0
  %t30 = extractvalue { ptr, i64 } %t28, 1
  %t31 = icmp uge i64 %t27, %t30
  br i1 %t31, label %chk.bad17, label %chk.ok16
chk.bad17:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 270)
  unreachable
chk.ok16:
  %t32 = getelementptr inbounds i8, ptr %t29, i64 %t27
  %t33 = load i8, ptr %slot_c_8
  store i8 %t33, ptr %t32
  %t34 = load i64, ptr %slot_w_6
  %t35 = add i64 %t34, 1
  store i64 %t35, ptr %slot_w_6
  %t36 = load i64, ptr %slot_i_7
  %t37 = add i64 %t36, 1
  store i64 %t37, ptr %slot_i_7
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t38 = load i64, ptr %slot_w_6
  ret i64 %t38
}

define i64 @aila_std_strings__blit({ ptr, i64 } %p_dst, i64 %p_w, i64 %p_cap, { ptr, i64 } %p_src) {
  %slot_dst_1 = alloca { ptr, i64 }, align 8
  %slot_w_2 = alloca i64, align 8
  %slot_cap_3 = alloca i64, align 8
  %slot_src_4 = alloca { ptr, i64 }, align 8
  %slot_k_5 = alloca i64, align 8
  %slot_at_6 = alloca i64, align 8
  %slot_n_7 = alloca i64, align 8
  store { ptr, i64 } %p_dst, ptr %slot_dst_1
  store i64 %p_w, ptr %slot_w_2
  store i64 %p_cap, ptr %slot_cap_3
  store { ptr, i64 } %p_src, ptr %slot_src_4
  store i64 0, ptr %slot_k_5
  %t1 = load i64, ptr %slot_w_2
  store i64 %t1, ptr %slot_at_6
  %t2 = load { ptr, i64 }, ptr %slot_src_4
  %t3 = extractvalue { ptr, i64 } %t2, 1
  store i64 %t3, ptr %slot_n_7
  br label %for.cond2
for.cond2:
  %t4 = load i64, ptr %slot_k_5
  %t5 = load i64, ptr %slot_n_7
  %t6 = icmp ult i64 %t4, %t5
  br i1 %t6, label %for.body4, label %for.end1
for.body4:
  %t7 = load i64, ptr %slot_at_6
  %t8 = load i64, ptr %slot_cap_3
  %t9 = icmp eq i64 %t7, %t8
  br i1 %t9, label %if.then6, label %if.next7
if.then6:
  %t10 = load i64, ptr %slot_at_6
  ret i64 %t10
if.next7:
  br label %if.end5
if.end5:
  %t11 = load i64, ptr %slot_at_6
  %t12 = load { ptr, i64 }, ptr %slot_dst_1
  %t13 = extractvalue { ptr, i64 } %t12, 0
  %t14 = extractvalue { ptr, i64 } %t12, 1
  %t15 = icmp uge i64 %t11, %t14
  br i1 %t15, label %chk.bad9, label %chk.ok8
chk.bad9:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 288)
  unreachable
chk.ok8:
  %t16 = getelementptr inbounds i8, ptr %t13, i64 %t11
  %t17 = load i64, ptr %slot_k_5
  %t18 = load { ptr, i64 }, ptr %slot_src_4
  %t19 = extractvalue { ptr, i64 } %t18, 0
  %t20 = extractvalue { ptr, i64 } %t18, 1
  %t21 = icmp uge i64 %t17, %t20
  br i1 %t21, label %chk.bad11, label %chk.ok10
chk.bad11:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 288)
  unreachable
chk.ok10:
  %t22 = getelementptr inbounds i8, ptr %t19, i64 %t17
  %t23 = load i8, ptr %t22
  store i8 %t23, ptr %t16
  %t24 = load i64, ptr %slot_at_6
  %t25 = add i64 %t24, 1
  store i64 %t25, ptr %slot_at_6
  %t26 = load i64, ptr %slot_k_5
  %t27 = add i64 %t26, 1
  store i64 %t27, ptr %slot_k_5
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t28 = load i64, ptr %slot_at_6
  ret i64 %t28
}
