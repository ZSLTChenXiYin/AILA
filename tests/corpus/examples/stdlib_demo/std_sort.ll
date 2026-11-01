; AILA module: std.sort
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [17 x i8] c"division by zero\00", align 1
@.str.1 = private unnamed_addr constant [14 x i8] c"std/sort.aila\00", align 1
@.str.2 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1

declare void @aila_rt_trap(ptr, i64, ptr, i64)

define void @aila_std_sort__sort_i64({ ptr, i64 } %p_buf) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  call void @aila_std_sort__sort_i64_by({ ptr, i64 } %t1, ptr @aila_std_sort__lt_i64)
  ret void
}

define void @aila_std_sort__sort_u64({ ptr, i64 } %p_buf) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  call void @aila_std_sort__sort_u64_by({ ptr, i64 } %t1, ptr @aila_std_sort__lt_u64)
  ret void
}

define void @aila_std_sort__sort_f64({ ptr, i64 } %p_buf) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  call void @aila_std_sort__sort_f64_by({ ptr, i64 } %t1, ptr @aila_std_sort__lt_f64)
  ret void
}

define void @aila_std_sort__sort_i64_by({ ptr, i64 } %p_buf, ptr %p_cmp) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_cmp_2 = alloca ptr, align 8
  %slot_n_3 = alloca i64, align 8
  %slot_gap_4 = alloca i64, align 8
  %slot_i_5 = alloca i64, align 8
  %slot_tmp_6 = alloca i64, align 8
  %slot_j_7 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store ptr %p_cmp, ptr %slot_cmp_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = extractvalue { ptr, i64 } %t1, 1
  store i64 %t2, ptr %slot_n_3
  %t3 = load i64, ptr %slot_n_3
  %t4 = icmp eq i64 2, 0
  br i1 %t4, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 30)
  unreachable
chk.ok1:
  %t5 = udiv i64 %t3, 2
  store i64 %t5, ptr %slot_gap_4
  br label %for.cond4
for.cond4:
  %t6 = load i64, ptr %slot_gap_4
  %t7 = icmp ugt i64 %t6, 0
  br i1 %t7, label %for.body6, label %for.end3
for.body6:
  %t8 = load i64, ptr %slot_gap_4
  store i64 %t8, ptr %slot_i_5
  br label %for.cond8
for.cond8:
  %t9 = load i64, ptr %slot_i_5
  %t10 = load i64, ptr %slot_n_3
  %t11 = icmp ult i64 %t9, %t10
  br i1 %t11, label %for.body10, label %for.end7
for.body10:
  %t12 = load i64, ptr %slot_i_5
  %t13 = load { ptr, i64 }, ptr %slot_buf_1
  %t14 = extractvalue { ptr, i64 } %t13, 0
  %t15 = extractvalue { ptr, i64 } %t13, 1
  %t16 = icmp uge i64 %t12, %t15
  br i1 %t16, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 34)
  unreachable
chk.ok11:
  %t17 = getelementptr inbounds i64, ptr %t14, i64 %t12
  %t18 = load i64, ptr %t17
  store i64 %t18, ptr %slot_tmp_6
  %t19 = load i64, ptr %slot_i_5
  store i64 %t19, ptr %slot_j_7
  br label %for.cond14
for.cond14:
  %t20 = load i64, ptr %slot_j_7
  %t21 = load i64, ptr %slot_gap_4
  %t22 = icmp uge i64 %t20, %t21
  br i1 %t22, label %for.body16, label %for.end13
for.body16:
  %t23 = load i64, ptr %slot_tmp_6
  %t24 = load i64, ptr %slot_j_7
  %t25 = load i64, ptr %slot_gap_4
  %t26 = sub i64 %t24, %t25
  %t27 = load { ptr, i64 }, ptr %slot_buf_1
  %t28 = extractvalue { ptr, i64 } %t27, 0
  %t29 = extractvalue { ptr, i64 } %t27, 1
  %t30 = icmp uge i64 %t26, %t29
  br i1 %t30, label %chk.bad21, label %chk.ok20
chk.bad21:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 37)
  unreachable
chk.ok20:
  %t31 = getelementptr inbounds i64, ptr %t28, i64 %t26
  %t32 = load i64, ptr %t31
  %t33 = load ptr, ptr %slot_cmp_2
  %t34 = call i1 %t33(i64 %t23, i64 %t32)
  br i1 %t34, label %if.then18, label %if.next19
if.then18:
  %t35 = load i64, ptr %slot_j_7
  %t36 = load { ptr, i64 }, ptr %slot_buf_1
  %t37 = extractvalue { ptr, i64 } %t36, 0
  %t38 = extractvalue { ptr, i64 } %t36, 1
  %t39 = icmp uge i64 %t35, %t38
  br i1 %t39, label %chk.bad23, label %chk.ok22
chk.bad23:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 38)
  unreachable
chk.ok22:
  %t40 = getelementptr inbounds i64, ptr %t37, i64 %t35
  %t41 = load i64, ptr %slot_j_7
  %t42 = load i64, ptr %slot_gap_4
  %t43 = sub i64 %t41, %t42
  %t44 = load { ptr, i64 }, ptr %slot_buf_1
  %t45 = extractvalue { ptr, i64 } %t44, 0
  %t46 = extractvalue { ptr, i64 } %t44, 1
  %t47 = icmp uge i64 %t43, %t46
  br i1 %t47, label %chk.bad25, label %chk.ok24
chk.bad25:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 38)
  unreachable
chk.ok24:
  %t48 = getelementptr inbounds i64, ptr %t45, i64 %t43
  %t49 = load i64, ptr %t48
  store i64 %t49, ptr %t40
  %t50 = load i64, ptr %slot_j_7
  %t51 = load i64, ptr %slot_gap_4
  %t52 = sub i64 %t50, %t51
  store i64 %t52, ptr %slot_j_7
  br label %if.end17
if.next19:
  br label %for.end13
if.end17:
  br label %for.step15
for.step15:
  br label %for.cond14
for.end13:
  %t53 = load i64, ptr %slot_j_7
  %t54 = load { ptr, i64 }, ptr %slot_buf_1
  %t55 = extractvalue { ptr, i64 } %t54, 0
  %t56 = extractvalue { ptr, i64 } %t54, 1
  %t57 = icmp uge i64 %t53, %t56
  br i1 %t57, label %chk.bad27, label %chk.ok26
chk.bad27:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 44)
  unreachable
chk.ok26:
  %t58 = getelementptr inbounds i64, ptr %t55, i64 %t53
  %t59 = load i64, ptr %slot_tmp_6
  store i64 %t59, ptr %t58
  %t60 = load i64, ptr %slot_i_5
  %t61 = add i64 %t60, 1
  store i64 %t61, ptr %slot_i_5
  br label %for.step9
for.step9:
  br label %for.cond8
for.end7:
  %t62 = load i64, ptr %slot_gap_4
  %t63 = icmp eq i64 2, 0
  br i1 %t63, label %chk.bad29, label %chk.ok28
chk.bad29:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 47)
  unreachable
chk.ok28:
  %t64 = udiv i64 %t62, 2
  store i64 %t64, ptr %slot_gap_4
  br label %for.step5
for.step5:
  br label %for.cond4
for.end3:
  ret void
}

define void @aila_std_sort__sort_u64_by({ ptr, i64 } %p_buf, ptr %p_cmp) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_cmp_2 = alloca ptr, align 8
  %slot_n_3 = alloca i64, align 8
  %slot_gap_4 = alloca i64, align 8
  %slot_i_5 = alloca i64, align 8
  %slot_tmp_6 = alloca i64, align 8
  %slot_j_7 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store ptr %p_cmp, ptr %slot_cmp_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = extractvalue { ptr, i64 } %t1, 1
  store i64 %t2, ptr %slot_n_3
  %t3 = load i64, ptr %slot_n_3
  %t4 = icmp eq i64 2, 0
  br i1 %t4, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 54)
  unreachable
chk.ok1:
  %t5 = udiv i64 %t3, 2
  store i64 %t5, ptr %slot_gap_4
  br label %for.cond4
for.cond4:
  %t6 = load i64, ptr %slot_gap_4
  %t7 = icmp ugt i64 %t6, 0
  br i1 %t7, label %for.body6, label %for.end3
for.body6:
  %t8 = load i64, ptr %slot_gap_4
  store i64 %t8, ptr %slot_i_5
  br label %for.cond8
for.cond8:
  %t9 = load i64, ptr %slot_i_5
  %t10 = load i64, ptr %slot_n_3
  %t11 = icmp ult i64 %t9, %t10
  br i1 %t11, label %for.body10, label %for.end7
for.body10:
  %t12 = load i64, ptr %slot_i_5
  %t13 = load { ptr, i64 }, ptr %slot_buf_1
  %t14 = extractvalue { ptr, i64 } %t13, 0
  %t15 = extractvalue { ptr, i64 } %t13, 1
  %t16 = icmp uge i64 %t12, %t15
  br i1 %t16, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 58)
  unreachable
chk.ok11:
  %t17 = getelementptr inbounds i64, ptr %t14, i64 %t12
  %t18 = load i64, ptr %t17
  store i64 %t18, ptr %slot_tmp_6
  %t19 = load i64, ptr %slot_i_5
  store i64 %t19, ptr %slot_j_7
  br label %for.cond14
for.cond14:
  %t20 = load i64, ptr %slot_j_7
  %t21 = load i64, ptr %slot_gap_4
  %t22 = icmp uge i64 %t20, %t21
  br i1 %t22, label %for.body16, label %for.end13
for.body16:
  %t23 = load i64, ptr %slot_tmp_6
  %t24 = load i64, ptr %slot_j_7
  %t25 = load i64, ptr %slot_gap_4
  %t26 = sub i64 %t24, %t25
  %t27 = load { ptr, i64 }, ptr %slot_buf_1
  %t28 = extractvalue { ptr, i64 } %t27, 0
  %t29 = extractvalue { ptr, i64 } %t27, 1
  %t30 = icmp uge i64 %t26, %t29
  br i1 %t30, label %chk.bad21, label %chk.ok20
chk.bad21:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 61)
  unreachable
chk.ok20:
  %t31 = getelementptr inbounds i64, ptr %t28, i64 %t26
  %t32 = load i64, ptr %t31
  %t33 = load ptr, ptr %slot_cmp_2
  %t34 = call i1 %t33(i64 %t23, i64 %t32)
  br i1 %t34, label %if.then18, label %if.next19
if.then18:
  %t35 = load i64, ptr %slot_j_7
  %t36 = load { ptr, i64 }, ptr %slot_buf_1
  %t37 = extractvalue { ptr, i64 } %t36, 0
  %t38 = extractvalue { ptr, i64 } %t36, 1
  %t39 = icmp uge i64 %t35, %t38
  br i1 %t39, label %chk.bad23, label %chk.ok22
chk.bad23:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 62)
  unreachable
chk.ok22:
  %t40 = getelementptr inbounds i64, ptr %t37, i64 %t35
  %t41 = load i64, ptr %slot_j_7
  %t42 = load i64, ptr %slot_gap_4
  %t43 = sub i64 %t41, %t42
  %t44 = load { ptr, i64 }, ptr %slot_buf_1
  %t45 = extractvalue { ptr, i64 } %t44, 0
  %t46 = extractvalue { ptr, i64 } %t44, 1
  %t47 = icmp uge i64 %t43, %t46
  br i1 %t47, label %chk.bad25, label %chk.ok24
chk.bad25:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 62)
  unreachable
chk.ok24:
  %t48 = getelementptr inbounds i64, ptr %t45, i64 %t43
  %t49 = load i64, ptr %t48
  store i64 %t49, ptr %t40
  %t50 = load i64, ptr %slot_j_7
  %t51 = load i64, ptr %slot_gap_4
  %t52 = sub i64 %t50, %t51
  store i64 %t52, ptr %slot_j_7
  br label %if.end17
if.next19:
  br label %for.end13
if.end17:
  br label %for.step15
for.step15:
  br label %for.cond14
for.end13:
  %t53 = load i64, ptr %slot_j_7
  %t54 = load { ptr, i64 }, ptr %slot_buf_1
  %t55 = extractvalue { ptr, i64 } %t54, 0
  %t56 = extractvalue { ptr, i64 } %t54, 1
  %t57 = icmp uge i64 %t53, %t56
  br i1 %t57, label %chk.bad27, label %chk.ok26
chk.bad27:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 68)
  unreachable
chk.ok26:
  %t58 = getelementptr inbounds i64, ptr %t55, i64 %t53
  %t59 = load i64, ptr %slot_tmp_6
  store i64 %t59, ptr %t58
  %t60 = load i64, ptr %slot_i_5
  %t61 = add i64 %t60, 1
  store i64 %t61, ptr %slot_i_5
  br label %for.step9
for.step9:
  br label %for.cond8
for.end7:
  %t62 = load i64, ptr %slot_gap_4
  %t63 = icmp eq i64 2, 0
  br i1 %t63, label %chk.bad29, label %chk.ok28
chk.bad29:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 71)
  unreachable
chk.ok28:
  %t64 = udiv i64 %t62, 2
  store i64 %t64, ptr %slot_gap_4
  br label %for.step5
for.step5:
  br label %for.cond4
for.end3:
  ret void
}

define void @aila_std_sort__sort_f64_by({ ptr, i64 } %p_buf, ptr %p_cmp) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_cmp_2 = alloca ptr, align 8
  %slot_n_3 = alloca i64, align 8
  %slot_gap_4 = alloca i64, align 8
  %slot_i_5 = alloca i64, align 8
  %slot_tmp_6 = alloca double, align 8
  %slot_j_7 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store ptr %p_cmp, ptr %slot_cmp_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = extractvalue { ptr, i64 } %t1, 1
  store i64 %t2, ptr %slot_n_3
  %t3 = load i64, ptr %slot_n_3
  %t4 = icmp eq i64 2, 0
  br i1 %t4, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 78)
  unreachable
chk.ok1:
  %t5 = udiv i64 %t3, 2
  store i64 %t5, ptr %slot_gap_4
  br label %for.cond4
for.cond4:
  %t6 = load i64, ptr %slot_gap_4
  %t7 = icmp ugt i64 %t6, 0
  br i1 %t7, label %for.body6, label %for.end3
for.body6:
  %t8 = load i64, ptr %slot_gap_4
  store i64 %t8, ptr %slot_i_5
  br label %for.cond8
for.cond8:
  %t9 = load i64, ptr %slot_i_5
  %t10 = load i64, ptr %slot_n_3
  %t11 = icmp ult i64 %t9, %t10
  br i1 %t11, label %for.body10, label %for.end7
for.body10:
  %t12 = load i64, ptr %slot_i_5
  %t13 = load { ptr, i64 }, ptr %slot_buf_1
  %t14 = extractvalue { ptr, i64 } %t13, 0
  %t15 = extractvalue { ptr, i64 } %t13, 1
  %t16 = icmp uge i64 %t12, %t15
  br i1 %t16, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 82)
  unreachable
chk.ok11:
  %t17 = getelementptr inbounds double, ptr %t14, i64 %t12
  %t18 = load double, ptr %t17
  store double %t18, ptr %slot_tmp_6
  %t19 = load i64, ptr %slot_i_5
  store i64 %t19, ptr %slot_j_7
  br label %for.cond14
for.cond14:
  %t20 = load i64, ptr %slot_j_7
  %t21 = load i64, ptr %slot_gap_4
  %t22 = icmp uge i64 %t20, %t21
  br i1 %t22, label %for.body16, label %for.end13
for.body16:
  %t23 = load double, ptr %slot_tmp_6
  %t24 = load i64, ptr %slot_j_7
  %t25 = load i64, ptr %slot_gap_4
  %t26 = sub i64 %t24, %t25
  %t27 = load { ptr, i64 }, ptr %slot_buf_1
  %t28 = extractvalue { ptr, i64 } %t27, 0
  %t29 = extractvalue { ptr, i64 } %t27, 1
  %t30 = icmp uge i64 %t26, %t29
  br i1 %t30, label %chk.bad21, label %chk.ok20
chk.bad21:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 85)
  unreachable
chk.ok20:
  %t31 = getelementptr inbounds double, ptr %t28, i64 %t26
  %t32 = load double, ptr %t31
  %t33 = load ptr, ptr %slot_cmp_2
  %t34 = call i1 %t33(double %t23, double %t32)
  br i1 %t34, label %if.then18, label %if.next19
if.then18:
  %t35 = load i64, ptr %slot_j_7
  %t36 = load { ptr, i64 }, ptr %slot_buf_1
  %t37 = extractvalue { ptr, i64 } %t36, 0
  %t38 = extractvalue { ptr, i64 } %t36, 1
  %t39 = icmp uge i64 %t35, %t38
  br i1 %t39, label %chk.bad23, label %chk.ok22
chk.bad23:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 86)
  unreachable
chk.ok22:
  %t40 = getelementptr inbounds double, ptr %t37, i64 %t35
  %t41 = load i64, ptr %slot_j_7
  %t42 = load i64, ptr %slot_gap_4
  %t43 = sub i64 %t41, %t42
  %t44 = load { ptr, i64 }, ptr %slot_buf_1
  %t45 = extractvalue { ptr, i64 } %t44, 0
  %t46 = extractvalue { ptr, i64 } %t44, 1
  %t47 = icmp uge i64 %t43, %t46
  br i1 %t47, label %chk.bad25, label %chk.ok24
chk.bad25:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 86)
  unreachable
chk.ok24:
  %t48 = getelementptr inbounds double, ptr %t45, i64 %t43
  %t49 = load double, ptr %t48
  store double %t49, ptr %t40
  %t50 = load i64, ptr %slot_j_7
  %t51 = load i64, ptr %slot_gap_4
  %t52 = sub i64 %t50, %t51
  store i64 %t52, ptr %slot_j_7
  br label %if.end17
if.next19:
  br label %for.end13
if.end17:
  br label %for.step15
for.step15:
  br label %for.cond14
for.end13:
  %t53 = load i64, ptr %slot_j_7
  %t54 = load { ptr, i64 }, ptr %slot_buf_1
  %t55 = extractvalue { ptr, i64 } %t54, 0
  %t56 = extractvalue { ptr, i64 } %t54, 1
  %t57 = icmp uge i64 %t53, %t56
  br i1 %t57, label %chk.bad27, label %chk.ok26
chk.bad27:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 92)
  unreachable
chk.ok26:
  %t58 = getelementptr inbounds double, ptr %t55, i64 %t53
  %t59 = load double, ptr %slot_tmp_6
  store double %t59, ptr %t58
  %t60 = load i64, ptr %slot_i_5
  %t61 = add i64 %t60, 1
  store i64 %t61, ptr %slot_i_5
  br label %for.step9
for.step9:
  br label %for.cond8
for.end7:
  %t62 = load i64, ptr %slot_gap_4
  %t63 = icmp eq i64 2, 0
  br i1 %t63, label %chk.bad29, label %chk.ok28
chk.bad29:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 95)
  unreachable
chk.ok28:
  %t64 = udiv i64 %t62, 2
  store i64 %t64, ptr %slot_gap_4
  br label %for.step5
for.step5:
  br label %for.cond4
for.end3:
  ret void
}

define i1 @aila_std_sort__lt_i64(i64 %p_a, i64 %p_b) {
  %slot_a_1 = alloca i64, align 8
  %slot_b_2 = alloca i64, align 8
  store i64 %p_a, ptr %slot_a_1
  store i64 %p_b, ptr %slot_b_2
  %t1 = load i64, ptr %slot_a_1
  %t2 = load i64, ptr %slot_b_2
  %t3 = icmp slt i64 %t1, %t2
  ret i1 %t3
}

define i1 @aila_std_sort__lt_u64(i64 %p_a, i64 %p_b) {
  %slot_a_1 = alloca i64, align 8
  %slot_b_2 = alloca i64, align 8
  store i64 %p_a, ptr %slot_a_1
  store i64 %p_b, ptr %slot_b_2
  %t1 = load i64, ptr %slot_a_1
  %t2 = load i64, ptr %slot_b_2
  %t3 = icmp ult i64 %t1, %t2
  ret i1 %t3
}

define i1 @aila_std_sort__lt_f64(double %p_a, double %p_b) {
  %slot_a_1 = alloca double, align 8
  %slot_b_2 = alloca double, align 8
  store double %p_a, ptr %slot_a_1
  store double %p_b, ptr %slot_b_2
  %t1 = load double, ptr %slot_a_1
  %t2 = load double, ptr %slot_b_2
  %t3 = fcmp olt double %t1, %t2
  ret i1 %t3
}
