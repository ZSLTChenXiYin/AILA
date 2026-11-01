; AILA module: std.path
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.1 = private unnamed_addr constant [14 x i8] c"std/path.aila\00", align 1


declare void @aila_rt_trap(ptr, i64, ptr, i64)

define { i64, i64 } @aila_std_path__base({ ptr, i64 } %p_p, i8 %p_sep) {
  %slot_p_1 = alloca { ptr, i64 }, align 8
  %slot_sep_2 = alloca i8, align 8
  %slot_b_3 = alloca { ptr, i64 }, align 8
  %slot_n_4 = alloca i64, align 8
  %slot_i_5 = alloca i64, align 8
  store { ptr, i64 } %p_p, ptr %slot_p_1
  store i8 %p_sep, ptr %slot_sep_2
  %t1 = load { ptr, i64 }, ptr %slot_p_1
  store { ptr, i64 } %t1, ptr %slot_b_3
  %t2 = load { ptr, i64 }, ptr %slot_b_3
  %t3 = extractvalue { ptr, i64 } %t2, 1
  store i64 %t3, ptr %slot_n_4
  %t4 = load i64, ptr %slot_n_4
  store i64 %t4, ptr %slot_i_5
  br label %for.cond2
for.cond2:
  %t5 = load i64, ptr %slot_i_5
  %t6 = icmp ugt i64 %t5, 0
  br i1 %t6, label %for.body4, label %for.end1
for.body4:
  %t7 = load i64, ptr %slot_i_5
  %t8 = sub i64 %t7, 1
  store i64 %t8, ptr %slot_i_5
  %t9 = load i64, ptr %slot_i_5
  %t10 = load { ptr, i64 }, ptr %slot_b_3
  %t11 = extractvalue { ptr, i64 } %t10, 0
  %t12 = extractvalue { ptr, i64 } %t10, 1
  %t13 = icmp uge i64 %t9, %t12
  br i1 %t13, label %chk.bad9, label %chk.ok8
chk.bad9:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 19)
  unreachable
chk.ok8:
  %t14 = getelementptr inbounds i8, ptr %t11, i64 %t9
  %t15 = load i8, ptr %t14
  %t16 = load i8, ptr %slot_sep_2
  %t17 = icmp eq i8 %t15, %t16
  br i1 %t17, label %if.then6, label %if.next7
if.then6:
  %t18 = load i64, ptr %slot_i_5
  %t19 = add i64 %t18, 1
  %t20 = load i64, ptr %slot_n_4
  %t21 = insertvalue { i64, i64 } undef, i64 %t19, 0
  %t22 = insertvalue { i64, i64 } %t21, i64 %t20, 1
  ret { i64, i64 } %t22
if.next7:
  br label %if.end5
if.end5:
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t23 = load i64, ptr %slot_n_4
  %t24 = insertvalue { i64, i64 } undef, i64 0, 0
  %t25 = insertvalue { i64, i64 } %t24, i64 %t23, 1
  ret { i64, i64 } %t25
}

define { i64, i64 } @aila_std_path__dir({ ptr, i64 } %p_p, i8 %p_sep) {
  %slot_p_1 = alloca { ptr, i64 }, align 8
  %slot_sep_2 = alloca i8, align 8
  %slot_b_3 = alloca { ptr, i64 }, align 8
  %slot_n_4 = alloca i64, align 8
  %slot_i_5 = alloca i64, align 8
  store { ptr, i64 } %p_p, ptr %slot_p_1
  store i8 %p_sep, ptr %slot_sep_2
  %t1 = load { ptr, i64 }, ptr %slot_p_1
  store { ptr, i64 } %t1, ptr %slot_b_3
  %t2 = load { ptr, i64 }, ptr %slot_b_3
  %t3 = extractvalue { ptr, i64 } %t2, 1
  store i64 %t3, ptr %slot_n_4
  %t4 = load i64, ptr %slot_n_4
  store i64 %t4, ptr %slot_i_5
  br label %for.cond2
for.cond2:
  %t5 = load i64, ptr %slot_i_5
  %t6 = icmp ugt i64 %t5, 0
  br i1 %t6, label %for.body4, label %for.end1
for.body4:
  %t7 = load i64, ptr %slot_i_5
  %t8 = sub i64 %t7, 1
  store i64 %t8, ptr %slot_i_5
  %t9 = load i64, ptr %slot_i_5
  %t10 = load { ptr, i64 }, ptr %slot_b_3
  %t11 = extractvalue { ptr, i64 } %t10, 0
  %t12 = extractvalue { ptr, i64 } %t10, 1
  %t13 = icmp uge i64 %t9, %t12
  br i1 %t13, label %chk.bad9, label %chk.ok8
chk.bad9:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 33)
  unreachable
chk.ok8:
  %t14 = getelementptr inbounds i8, ptr %t11, i64 %t9
  %t15 = load i8, ptr %t14
  %t16 = load i8, ptr %slot_sep_2
  %t17 = icmp eq i8 %t15, %t16
  br i1 %t17, label %if.then6, label %if.next7
if.then6:
  %t18 = load i64, ptr %slot_i_5
  %t19 = add i64 %t18, 1
  %t20 = insertvalue { i64, i64 } undef, i64 0, 0
  %t21 = insertvalue { i64, i64 } %t20, i64 %t19, 1
  ret { i64, i64 } %t21
if.next7:
  br label %if.end5
if.end5:
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t22 = insertvalue { i64, i64 } undef, i64 0, 0
  %t23 = insertvalue { i64, i64 } %t22, i64 0, 1
  ret { i64, i64 } %t23
}

define { i64, i64 } @aila_std_path__ext({ ptr, i64 } %p_p) {
  %slot_p_1 = alloca { ptr, i64 }, align 8
  %slot_bs_2 = alloca i64, align 8
  %slot_be_3 = alloca i64, align 8
  %slot_b_4 = alloca { ptr, i64 }, align 8
  %slot_i_5 = alloca i64, align 8
  %slot_c_6 = alloca i8, align 8
  %slot___destr_7 = alloca { i64, i64 }, align 8
  store { ptr, i64 } %p_p, ptr %slot_p_1
  store i64 0, ptr %slot_bs_2
  store i64 0, ptr %slot_be_3
  %t1 = load { ptr, i64 }, ptr %slot_p_1
  %t2 = call { i64, i64 } @aila_std_path__base({ ptr, i64 } %t1, i8 47)
  store { i64, i64 } %t2, ptr %slot___destr_7
  %t3 = getelementptr inbounds { i64, i64 }, ptr %slot___destr_7, i32 0, i32 0
  %t4 = load i64, ptr %t3
  store i64 %t4, ptr %slot_bs_2
  %t5 = getelementptr inbounds { i64, i64 }, ptr %slot___destr_7, i32 0, i32 1
  %t6 = load i64, ptr %t5
  store i64 %t6, ptr %slot_be_3
  %t7 = load { ptr, i64 }, ptr %slot_p_1
  store { ptr, i64 } %t7, ptr %slot_b_4
  %t8 = load i64, ptr %slot_be_3
  store i64 %t8, ptr %slot_i_5
  br label %for.cond2
for.cond2:
  %t9 = load i64, ptr %slot_i_5
  %t10 = load i64, ptr %slot_bs_2
  %t11 = icmp ugt i64 %t9, %t10
  br i1 %t11, label %for.body4, label %for.end1
for.body4:
  %t12 = load i64, ptr %slot_i_5
  %t13 = sub i64 %t12, 1
  store i64 %t13, ptr %slot_i_5
  %t14 = load i64, ptr %slot_i_5
  %t15 = load { ptr, i64 }, ptr %slot_b_4
  %t16 = extractvalue { ptr, i64 } %t15, 0
  %t17 = extractvalue { ptr, i64 } %t15, 1
  %t18 = icmp uge i64 %t14, %t17
  br i1 %t18, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 50)
  unreachable
chk.ok5:
  %t19 = getelementptr inbounds i8, ptr %t16, i64 %t14
  %t20 = load i8, ptr %t19
  store i8 %t20, ptr %slot_c_6
  %t21 = load i8, ptr %slot_c_6
  %t22 = icmp eq i8 %t21, 46
  br i1 %t22, label %if.then8, label %if.next9
if.then8:
  %t23 = load i64, ptr %slot_be_3
  %t24 = load i64, ptr %slot_i_5
  %t25 = add i64 %t24, 1
  %t26 = sub i64 %t23, %t25
  %t27 = icmp eq i64 %t26, 0
  br i1 %t27, label %if.then11, label %if.next12
if.then11:
  %t28 = insertvalue { i64, i64 } undef, i64 0, 0
  %t29 = insertvalue { i64, i64 } %t28, i64 0, 1
  ret { i64, i64 } %t29
if.next12:
  br label %if.end10
if.end10:
  %t30 = load i64, ptr %slot_i_5
  %t31 = add i64 %t30, 1
  %t32 = load i64, ptr %slot_be_3
  %t33 = insertvalue { i64, i64 } undef, i64 %t31, 0
  %t34 = insertvalue { i64, i64 } %t33, i64 %t32, 1
  ret { i64, i64 } %t34
if.next9:
  br label %if.end7
if.end7:
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t35 = insertvalue { i64, i64 } undef, i64 0, 0
  %t36 = insertvalue { i64, i64 } %t35, i64 0, 1
  ret { i64, i64 } %t36
}

define i64 @aila_std_path__join({ ptr, i64 } %p_a, { ptr, i64 } %p_b, i8 %p_sep, { ptr, i64 } %p_dst) {
  %slot_a_1 = alloca { ptr, i64 }, align 8
  %slot_b_2 = alloca { ptr, i64 }, align 8
  %slot_sep_3 = alloca i8, align 8
  %slot_dst_4 = alloca { ptr, i64 }, align 8
  %slot_need_5 = alloca i64, align 8
  %slot_cap_6 = alloca i64, align 8
  %slot_w_7 = alloca i64, align 8
  %slot_take_a_8 = alloca i64, align 8
  %slot_room_9 = alloca i64, align 8
  %slot_take_b_10 = alloca i64, align 8
  %slot___sc_11 = alloca i1, align 8
  %slot___sc_12 = alloca i1, align 8
  %slot___sc_13 = alloca i1, align 8
  store { ptr, i64 } %p_a, ptr %slot_a_1
  store { ptr, i64 } %p_b, ptr %slot_b_2
  store i8 %p_sep, ptr %slot_sep_3
  store { ptr, i64 } %p_dst, ptr %slot_dst_4
  %t1 = load { ptr, i64 }, ptr %slot_a_1
  %t2 = load { ptr, i64 }, ptr %slot_b_2
  %t3 = load i8, ptr %slot_sep_3
  %t4 = call i64 @aila_std_path__join_len({ ptr, i64 } %t1, { ptr, i64 } %t2, i8 %t3)
  store i64 %t4, ptr %slot_need_5
  %t5 = load { ptr, i64 }, ptr %slot_dst_4
  %t6 = extractvalue { ptr, i64 } %t5, 1
  store i64 %t6, ptr %slot_cap_6
  %t7 = load i64, ptr %slot_cap_6
  %t8 = icmp eq i64 %t7, 0
  br i1 %t8, label %sc.l4, label %sc.r5
sc.l4:
  store i1 1, ptr %slot___sc_11
  br label %sc.end6
sc.r5:
  %t9 = load i64, ptr %slot_need_5
  %t10 = icmp eq i64 %t9, 0
  store i1 %t10, ptr %slot___sc_11
  br label %sc.end6
sc.end6:
  %t11 = load i1, ptr %slot___sc_11
  br i1 %t11, label %if.then2, label %if.next3
if.then2:
  ret i64 0
if.next3:
  br label %if.end1
if.end1:
  store i64 0, ptr %slot_w_7
  %t12 = load { ptr, i64 }, ptr %slot_b_2
  %t13 = load i8, ptr %slot_sep_3
  %t14 = call i1 @aila_std_path__has_abs({ ptr, i64 } %t12, i8 %t13)
  %t15 = xor i1 %t14, true
  br i1 %t15, label %sc.r11, label %sc.l10
sc.l10:
  store i1 0, ptr %slot___sc_12
  br label %sc.end12
sc.r11:
  %t16 = load { ptr, i64 }, ptr %slot_a_1
  %t17 = extractvalue { ptr, i64 } %t16, 1
  %t18 = icmp ugt i64 %t17, 0
  store i1 %t18, ptr %slot___sc_12
  br label %sc.end12
sc.end12:
  %t19 = load i1, ptr %slot___sc_12
  br i1 %t19, label %if.then8, label %if.next9
if.then8:
  %t20 = load i64, ptr %slot_need_5
  %t21 = sub i64 %t20, 1
  %t22 = load i64, ptr %slot_cap_6
  %t23 = load { ptr, i64 }, ptr %slot_a_1
  %t24 = extractvalue { ptr, i64 } %t23, 1
  %t25 = call i64 @aila_std_path__min3(i64 %t21, i64 %t22, i64 %t24)
  store i64 %t25, ptr %slot_take_a_8
  %t26 = load { ptr, i64 }, ptr %slot_dst_4
  %t27 = load { ptr, i64 }, ptr %slot_a_1
  %t28 = load i64, ptr %slot_take_a_8
  call void @aila_std_path__blit_at({ ptr, i64 } %t26, i64 0, { ptr, i64 } %t27, i64 %t28)
  %t29 = load i64, ptr %slot_take_a_8
  store i64 %t29, ptr %slot_w_7
  %t30 = load i64, ptr %slot_w_7
  %t31 = load i64, ptr %slot_cap_6
  %t32 = icmp ult i64 %t30, %t31
  br i1 %t32, label %sc.r17, label %sc.l16
sc.l16:
  store i1 0, ptr %slot___sc_13
  br label %sc.end18
sc.r17:
  %t33 = load i64, ptr %slot_w_7
  %t34 = load i64, ptr %slot_need_5
  %t35 = icmp ult i64 %t33, %t34
  store i1 %t35, ptr %slot___sc_13
  br label %sc.end18
sc.end18:
  %t36 = load i1, ptr %slot___sc_13
  br i1 %t36, label %if.then14, label %if.next15
if.then14:
  %t37 = load i64, ptr %slot_w_7
  %t38 = load { ptr, i64 }, ptr %slot_dst_4
  %t39 = extractvalue { ptr, i64 } %t38, 0
  %t40 = extractvalue { ptr, i64 } %t38, 1
  %t41 = icmp uge i64 %t37, %t40
  br i1 %t41, label %chk.bad20, label %chk.ok19
chk.bad20:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 76)
  unreachable
chk.ok19:
  %t42 = getelementptr inbounds i8, ptr %t39, i64 %t37
  %t43 = load i8, ptr %slot_sep_3
  store i8 %t43, ptr %t42
  %t44 = load i64, ptr %slot_w_7
  %t45 = add i64 %t44, 1
  store i64 %t45, ptr %slot_w_7
  br label %if.end13
if.next15:
  br label %if.end13
if.end13:
  br label %if.end7
if.next9:
  br label %if.end7
if.end7:
  %t46 = load i64, ptr %slot_w_7
  %t47 = load i64, ptr %slot_cap_6
  %t48 = icmp ult i64 %t46, %t47
  br i1 %t48, label %if.then22, label %if.next23
if.then22:
  %t49 = load i64, ptr %slot_cap_6
  %t50 = load i64, ptr %slot_w_7
  %t51 = sub i64 %t49, %t50
  store i64 %t51, ptr %slot_room_9
  %t52 = load { ptr, i64 }, ptr %slot_b_2
  %t53 = extractvalue { ptr, i64 } %t52, 1
  %t54 = load i64, ptr %slot_room_9
  %t55 = load i64, ptr %slot_need_5
  %t56 = load i64, ptr %slot_w_7
  %t57 = sub i64 %t55, %t56
  %t58 = call i64 @aila_std_path__min3(i64 %t53, i64 %t54, i64 %t57)
  store i64 %t58, ptr %slot_take_b_10
  %t59 = load i64, ptr %slot_take_b_10
  %t60 = icmp ugt i64 %t59, 0
  br i1 %t60, label %if.then25, label %if.next26
if.then25:
  %t61 = load { ptr, i64 }, ptr %slot_dst_4
  %t62 = load i64, ptr %slot_w_7
  %t63 = load { ptr, i64 }, ptr %slot_b_2
  %t64 = load i64, ptr %slot_take_b_10
  call void @aila_std_path__blit_at({ ptr, i64 } %t61, i64 %t62, { ptr, i64 } %t63, i64 %t64)
  %t65 = load i64, ptr %slot_w_7
  %t66 = load i64, ptr %slot_take_b_10
  %t67 = add i64 %t65, %t66
  store i64 %t67, ptr %slot_w_7
  br label %if.end24
if.next26:
  br label %if.end24
if.end24:
  br label %if.end21
if.next23:
  br label %if.end21
if.end21:
  %t68 = load i64, ptr %slot_w_7
  ret i64 %t68
}

define i64 @aila_std_path__join_len({ ptr, i64 } %p_a, { ptr, i64 } %p_b, i8 %p_sep) {
  %slot_a_1 = alloca { ptr, i64 }, align 8
  %slot_b_2 = alloca { ptr, i64 }, align 8
  %slot_sep_3 = alloca i8, align 8
  %slot___sc_4 = alloca i1, align 8
  store { ptr, i64 } %p_a, ptr %slot_a_1
  store { ptr, i64 } %p_b, ptr %slot_b_2
  store i8 %p_sep, ptr %slot_sep_3
  %t1 = load { ptr, i64 }, ptr %slot_b_2
  %t2 = load i8, ptr %slot_sep_3
  %t3 = call i1 @aila_std_path__has_abs({ ptr, i64 } %t1, i8 %t2)
  br i1 %t3, label %sc.l4, label %sc.r5
sc.l4:
  store i1 1, ptr %slot___sc_4
  br label %sc.end6
sc.r5:
  %t4 = load { ptr, i64 }, ptr %slot_a_1
  %t5 = extractvalue { ptr, i64 } %t4, 1
  %t6 = icmp eq i64 %t5, 0
  store i1 %t6, ptr %slot___sc_4
  br label %sc.end6
sc.end6:
  %t7 = load i1, ptr %slot___sc_4
  br i1 %t7, label %if.then2, label %if.next3
if.then2:
  %t8 = load { ptr, i64 }, ptr %slot_b_2
  %t9 = extractvalue { ptr, i64 } %t8, 1
  ret i64 %t9
if.next3:
  br label %if.end1
if.end1:
  %t10 = load { ptr, i64 }, ptr %slot_b_2
  %t11 = extractvalue { ptr, i64 } %t10, 1
  %t12 = icmp eq i64 %t11, 0
  br i1 %t12, label %if.then8, label %if.next9
if.then8:
  %t13 = load { ptr, i64 }, ptr %slot_a_1
  %t14 = extractvalue { ptr, i64 } %t13, 1
  ret i64 %t14
if.next9:
  br label %if.end7
if.end7:
  %t15 = load { ptr, i64 }, ptr %slot_a_1
  %t16 = extractvalue { ptr, i64 } %t15, 1
  %t17 = add i64 %t16, 1
  %t18 = load { ptr, i64 }, ptr %slot_b_2
  %t19 = extractvalue { ptr, i64 } %t18, 1
  %t20 = add i64 %t17, %t19
  ret i64 %t20
}

define i1 @aila_std_path__has_abs({ ptr, i64 } %p_b, i8 %p_sep) {
  %slot_b_1 = alloca { ptr, i64 }, align 8
  %slot_sep_2 = alloca i8, align 8
  %slot_bb_3 = alloca { ptr, i64 }, align 8
  %slot___sc_4 = alloca i1, align 8
  store { ptr, i64 } %p_b, ptr %slot_b_1
  store i8 %p_sep, ptr %slot_sep_2
  %t1 = load { ptr, i64 }, ptr %slot_b_1
  store { ptr, i64 } %t1, ptr %slot_bb_3
  %t2 = load { ptr, i64 }, ptr %slot_bb_3
  %t3 = extractvalue { ptr, i64 } %t2, 1
  %t4 = icmp ugt i64 %t3, 0
  br i1 %t4, label %sc.r2, label %sc.l1
sc.l1:
  store i1 0, ptr %slot___sc_4
  br label %sc.end3
sc.r2:
  %t5 = load { ptr, i64 }, ptr %slot_bb_3
  %t6 = extractvalue { ptr, i64 } %t5, 0
  %t7 = extractvalue { ptr, i64 } %t5, 1
  %t8 = icmp uge i64 0, %t7
  br i1 %t8, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 107)
  unreachable
chk.ok4:
  %t9 = getelementptr inbounds i8, ptr %t6, i64 0
  %t10 = load i8, ptr %t9
  %t11 = load i8, ptr %slot_sep_2
  %t12 = icmp eq i8 %t10, %t11
  store i1 %t12, ptr %slot___sc_4
  br label %sc.end3
sc.end3:
  %t13 = load i1, ptr %slot___sc_4
  ret i1 %t13
}

define void @aila_std_path__blit_at({ ptr, i64 } %p_dst, i64 %p_w, { ptr, i64 } %p_t, i64 %p_n) {
  %slot_dst_1 = alloca { ptr, i64 }, align 8
  %slot_w_2 = alloca i64, align 8
  %slot_t_3 = alloca { ptr, i64 }, align 8
  %slot_n_4 = alloca i64, align 8
  %slot_tb_5 = alloca { ptr, i64 }, align 8
  %slot_i_6 = alloca i64, align 8
  store { ptr, i64 } %p_dst, ptr %slot_dst_1
  store i64 %p_w, ptr %slot_w_2
  store { ptr, i64 } %p_t, ptr %slot_t_3
  store i64 %p_n, ptr %slot_n_4
  %t1 = load { ptr, i64 }, ptr %slot_t_3
  store { ptr, i64 } %t1, ptr %slot_tb_5
  store i64 0, ptr %slot_i_6
  br label %for.cond2
for.cond2:
  %t2 = load i64, ptr %slot_i_6
  %t3 = load i64, ptr %slot_n_4
  %t4 = icmp ult i64 %t2, %t3
  br i1 %t4, label %for.body4, label %for.end1
for.body4:
  %t5 = load i64, ptr %slot_w_2
  %t6 = load i64, ptr %slot_i_6
  %t7 = add i64 %t5, %t6
  %t8 = load { ptr, i64 }, ptr %slot_dst_1
  %t9 = extractvalue { ptr, i64 } %t8, 0
  %t10 = extractvalue { ptr, i64 } %t8, 1
  %t11 = icmp uge i64 %t7, %t10
  br i1 %t11, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 115)
  unreachable
chk.ok5:
  %t12 = getelementptr inbounds i8, ptr %t9, i64 %t7
  %t13 = load i64, ptr %slot_i_6
  %t14 = load { ptr, i64 }, ptr %slot_tb_5
  %t15 = extractvalue { ptr, i64 } %t14, 0
  %t16 = extractvalue { ptr, i64 } %t14, 1
  %t17 = icmp uge i64 %t13, %t16
  br i1 %t17, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 115)
  unreachable
chk.ok7:
  %t18 = getelementptr inbounds i8, ptr %t15, i64 %t13
  %t19 = load i8, ptr %t18
  store i8 %t19, ptr %t12
  %t20 = load i64, ptr %slot_i_6
  %t21 = add i64 %t20, 1
  store i64 %t21, ptr %slot_i_6
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  ret void
}

define i64 @aila_std_path__min3(i64 %p_a, i64 %p_b, i64 %p_c) {
  %slot_a_1 = alloca i64, align 8
  %slot_b_2 = alloca i64, align 8
  %slot_c_3 = alloca i64, align 8
  %slot_m_4 = alloca i64, align 8
  store i64 %p_a, ptr %slot_a_1
  store i64 %p_b, ptr %slot_b_2
  store i64 %p_c, ptr %slot_c_3
  %t1 = load i64, ptr %slot_a_1
  store i64 %t1, ptr %slot_m_4
  %t2 = load i64, ptr %slot_b_2
  %t3 = load i64, ptr %slot_m_4
  %t4 = icmp ult i64 %t2, %t3
  br i1 %t4, label %if.then2, label %if.next3
if.then2:
  %t5 = load i64, ptr %slot_b_2
  store i64 %t5, ptr %slot_m_4
  br label %if.end1
if.next3:
  br label %if.end1
if.end1:
  %t6 = load i64, ptr %slot_c_3
  %t7 = load i64, ptr %slot_m_4
  %t8 = icmp ult i64 %t6, %t7
  br i1 %t8, label %if.then5, label %if.next6
if.then5:
  %t9 = load i64, ptr %slot_c_3
  store i64 %t9, ptr %slot_m_4
  br label %if.end4
if.next6:
  br label %if.end4
if.end4:
  %t10 = load i64, ptr %slot_m_4
  ret i64 %t10
}
