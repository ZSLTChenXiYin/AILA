; AILA module: std.utf8
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.1 = private unnamed_addr constant [14 x i8] c"std/utf8.aila\00", align 1
@.str.2 = private unnamed_addr constant [25 x i8] c"shift count out of range\00", align 1

declare void @aila_rt_trap(ptr, i64, ptr, i64)

define i64 @aila_std_utf8__encode(i32 %p_cp, { ptr, i64 } %p_buf) {
  %slot_cp_1 = alloca i32, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  store i32 %p_cp, ptr %slot_cp_1
  store { ptr, i64 } %p_buf, ptr %slot_buf_2
  %t1 = load i32, ptr %slot_cp_1
  %t2 = icmp ult i32 %t1, 128
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  %t3 = load { ptr, i64 }, ptr %slot_buf_2
  %t4 = extractvalue { ptr, i64 } %t3, 0
  %t5 = extractvalue { ptr, i64 } %t3, 1
  %t6 = icmp uge i64 0, %t5
  br i1 %t6, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 14)
  unreachable
chk.ok4:
  %t7 = getelementptr inbounds i8, ptr %t4, i64 0
  %t8 = load i32, ptr %slot_cp_1
  %t9 = trunc i32 %t8 to i8
  store i8 %t9, ptr %t7
  ret i64 1
if.next3:
  br label %if.end1
if.end1:
  %t10 = load i32, ptr %slot_cp_1
  %t11 = icmp ult i32 %t10, 2048
  br i1 %t11, label %if.then7, label %if.next8
if.then7:
  %t12 = load { ptr, i64 }, ptr %slot_buf_2
  %t13 = extractvalue { ptr, i64 } %t12, 0
  %t14 = extractvalue { ptr, i64 } %t12, 1
  %t15 = icmp uge i64 0, %t14
  br i1 %t15, label %chk.bad10, label %chk.ok9
chk.bad10:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 18)
  unreachable
chk.ok9:
  %t16 = getelementptr inbounds i8, ptr %t13, i64 0
  %t17 = load i32, ptr %slot_cp_1
  %t18 = icmp uge i32 6, 32
  br i1 %t18, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 18)
  unreachable
chk.ok11:
  %t19 = lshr i32 %t17, 6
  %t20 = add i32 192, %t19
  %t21 = trunc i32 %t20 to i8
  store i8 %t21, ptr %t16
  %t22 = load { ptr, i64 }, ptr %slot_buf_2
  %t23 = extractvalue { ptr, i64 } %t22, 0
  %t24 = extractvalue { ptr, i64 } %t22, 1
  %t25 = icmp uge i64 1, %t24
  br i1 %t25, label %chk.bad14, label %chk.ok13
chk.bad14:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 19)
  unreachable
chk.ok13:
  %t26 = getelementptr inbounds i8, ptr %t23, i64 1
  %t27 = load i32, ptr %slot_cp_1
  %t28 = and i32 %t27, 63
  %t29 = add i32 128, %t28
  %t30 = trunc i32 %t29 to i8
  store i8 %t30, ptr %t26
  ret i64 2
if.next8:
  br label %if.end6
if.end6:
  %t31 = load i32, ptr %slot_cp_1
  %t32 = icmp ult i32 %t31, 65536
  br i1 %t32, label %if.then16, label %if.next17
if.then16:
  %t33 = load { ptr, i64 }, ptr %slot_buf_2
  %t34 = extractvalue { ptr, i64 } %t33, 0
  %t35 = extractvalue { ptr, i64 } %t33, 1
  %t36 = icmp uge i64 0, %t35
  br i1 %t36, label %chk.bad19, label %chk.ok18
chk.bad19:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 23)
  unreachable
chk.ok18:
  %t37 = getelementptr inbounds i8, ptr %t34, i64 0
  %t38 = load i32, ptr %slot_cp_1
  %t39 = icmp uge i32 12, 32
  br i1 %t39, label %chk.bad21, label %chk.ok20
chk.bad21:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 23)
  unreachable
chk.ok20:
  %t40 = lshr i32 %t38, 12
  %t41 = add i32 224, %t40
  %t42 = trunc i32 %t41 to i8
  store i8 %t42, ptr %t37
  %t43 = load { ptr, i64 }, ptr %slot_buf_2
  %t44 = extractvalue { ptr, i64 } %t43, 0
  %t45 = extractvalue { ptr, i64 } %t43, 1
  %t46 = icmp uge i64 1, %t45
  br i1 %t46, label %chk.bad23, label %chk.ok22
chk.bad23:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 24)
  unreachable
chk.ok22:
  %t47 = getelementptr inbounds i8, ptr %t44, i64 1
  %t48 = load i32, ptr %slot_cp_1
  %t49 = icmp uge i32 6, 32
  br i1 %t49, label %chk.bad25, label %chk.ok24
chk.bad25:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 24)
  unreachable
chk.ok24:
  %t50 = lshr i32 %t48, 6
  %t51 = and i32 %t50, 63
  %t52 = add i32 128, %t51
  %t53 = trunc i32 %t52 to i8
  store i8 %t53, ptr %t47
  %t54 = load { ptr, i64 }, ptr %slot_buf_2
  %t55 = extractvalue { ptr, i64 } %t54, 0
  %t56 = extractvalue { ptr, i64 } %t54, 1
  %t57 = icmp uge i64 2, %t56
  br i1 %t57, label %chk.bad27, label %chk.ok26
chk.bad27:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 25)
  unreachable
chk.ok26:
  %t58 = getelementptr inbounds i8, ptr %t55, i64 2
  %t59 = load i32, ptr %slot_cp_1
  %t60 = and i32 %t59, 63
  %t61 = add i32 128, %t60
  %t62 = trunc i32 %t61 to i8
  store i8 %t62, ptr %t58
  ret i64 3
if.next17:
  br label %if.end15
if.end15:
  %t63 = load { ptr, i64 }, ptr %slot_buf_2
  %t64 = extractvalue { ptr, i64 } %t63, 0
  %t65 = extractvalue { ptr, i64 } %t63, 1
  %t66 = icmp uge i64 0, %t65
  br i1 %t66, label %chk.bad29, label %chk.ok28
chk.bad29:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 28)
  unreachable
chk.ok28:
  %t67 = getelementptr inbounds i8, ptr %t64, i64 0
  %t68 = load i32, ptr %slot_cp_1
  %t69 = icmp uge i32 18, 32
  br i1 %t69, label %chk.bad31, label %chk.ok30
chk.bad31:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 28)
  unreachable
chk.ok30:
  %t70 = lshr i32 %t68, 18
  %t71 = add i32 240, %t70
  %t72 = trunc i32 %t71 to i8
  store i8 %t72, ptr %t67
  %t73 = load { ptr, i64 }, ptr %slot_buf_2
  %t74 = extractvalue { ptr, i64 } %t73, 0
  %t75 = extractvalue { ptr, i64 } %t73, 1
  %t76 = icmp uge i64 1, %t75
  br i1 %t76, label %chk.bad33, label %chk.ok32
chk.bad33:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 29)
  unreachable
chk.ok32:
  %t77 = getelementptr inbounds i8, ptr %t74, i64 1
  %t78 = load i32, ptr %slot_cp_1
  %t79 = icmp uge i32 12, 32
  br i1 %t79, label %chk.bad35, label %chk.ok34
chk.bad35:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 29)
  unreachable
chk.ok34:
  %t80 = lshr i32 %t78, 12
  %t81 = and i32 %t80, 63
  %t82 = add i32 128, %t81
  %t83 = trunc i32 %t82 to i8
  store i8 %t83, ptr %t77
  %t84 = load { ptr, i64 }, ptr %slot_buf_2
  %t85 = extractvalue { ptr, i64 } %t84, 0
  %t86 = extractvalue { ptr, i64 } %t84, 1
  %t87 = icmp uge i64 2, %t86
  br i1 %t87, label %chk.bad37, label %chk.ok36
chk.bad37:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 30)
  unreachable
chk.ok36:
  %t88 = getelementptr inbounds i8, ptr %t85, i64 2
  %t89 = load i32, ptr %slot_cp_1
  %t90 = icmp uge i32 6, 32
  br i1 %t90, label %chk.bad39, label %chk.ok38
chk.bad39:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 30)
  unreachable
chk.ok38:
  %t91 = lshr i32 %t89, 6
  %t92 = and i32 %t91, 63
  %t93 = add i32 128, %t92
  %t94 = trunc i32 %t93 to i8
  store i8 %t94, ptr %t88
  %t95 = load { ptr, i64 }, ptr %slot_buf_2
  %t96 = extractvalue { ptr, i64 } %t95, 0
  %t97 = extractvalue { ptr, i64 } %t95, 1
  %t98 = icmp uge i64 3, %t97
  br i1 %t98, label %chk.bad41, label %chk.ok40
chk.bad41:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 31)
  unreachable
chk.ok40:
  %t99 = getelementptr inbounds i8, ptr %t96, i64 3
  %t100 = load i32, ptr %slot_cp_1
  %t101 = and i32 %t100, 63
  %t102 = add i32 128, %t101
  %t103 = trunc i32 %t102 to i8
  store i8 %t103, ptr %t99
  ret i64 4
}

define i64 @aila_std_utf8__rune_len(i8 %p_first) {
  %slot_first_1 = alloca i8, align 8
  store i8 %p_first, ptr %slot_first_1
  %t1 = load i8, ptr %slot_first_1
  %t2 = icmp ult i8 %t1, 128
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  ret i64 1
if.next3:
  br label %if.end1
if.end1:
  %t3 = load i8, ptr %slot_first_1
  %t4 = icmp ult i8 %t3, 192
  br i1 %t4, label %if.then5, label %if.next6
if.then5:
  ret i64 0
if.next6:
  br label %if.end4
if.end4:
  %t5 = load i8, ptr %slot_first_1
  %t6 = icmp ult i8 %t5, 224
  br i1 %t6, label %if.then8, label %if.next9
if.then8:
  ret i64 2
if.next9:
  br label %if.end7
if.end7:
  %t7 = load i8, ptr %slot_first_1
  %t8 = icmp ult i8 %t7, 240
  br i1 %t8, label %if.then11, label %if.next12
if.then11:
  ret i64 3
if.next12:
  br label %if.end10
if.end10:
  %t9 = load i8, ptr %slot_first_1
  %t10 = icmp ult i8 %t9, 248
  br i1 %t10, label %if.then14, label %if.next15
if.then14:
  ret i64 4
if.next15:
  br label %if.end13
if.end13:
  ret i64 0
}

define i32 @aila_std_utf8__cp_at({ ptr, i64 } %p_t, i64 %p_i) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_i_2 = alloca i64, align 8
  %slot_b_3 = alloca { ptr, i64 }, align 8
  %slot_n_4 = alloca i64, align 8
  %slot_c0_5 = alloca i32, align 8
  %slot_step_6 = alloca i64, align 8
  %slot_cp_7 = alloca i32, align 8
  %slot_k_8 = alloca i64, align 8
  %slot_cb_9 = alloca i32, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  store i64 %p_i, ptr %slot_i_2
  %t1 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t1, ptr %slot_b_3
  %t2 = load { ptr, i64 }, ptr %slot_b_3
  %t3 = extractvalue { ptr, i64 } %t2, 1
  store i64 %t3, ptr %slot_n_4
  %t4 = load i64, ptr %slot_i_2
  %t5 = load i64, ptr %slot_n_4
  %t6 = icmp uge i64 %t4, %t5
  br i1 %t6, label %if.then2, label %if.next3
if.then2:
  ret i32 4294967295
if.next3:
  br label %if.end1
if.end1:
  %t7 = load i64, ptr %slot_i_2
  %t8 = load { ptr, i64 }, ptr %slot_b_3
  %t9 = extractvalue { ptr, i64 } %t8, 0
  %t10 = extractvalue { ptr, i64 } %t8, 1
  %t11 = icmp uge i64 %t7, %t10
  br i1 %t11, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 64)
  unreachable
chk.ok4:
  %t12 = getelementptr inbounds i8, ptr %t9, i64 %t7
  %t13 = load i8, ptr %t12
  %t14 = zext i8 %t13 to i32
  store i32 %t14, ptr %slot_c0_5
  %t15 = load i32, ptr %slot_c0_5
  %t16 = icmp ult i32 %t15, 128
  br i1 %t16, label %if.then7, label %if.next8
if.then7:
  %t17 = load i32, ptr %slot_c0_5
  ret i32 %t17
if.next8:
  br label %if.end6
if.end6:
  %t18 = load i64, ptr %slot_i_2
  %t19 = load { ptr, i64 }, ptr %slot_b_3
  %t20 = extractvalue { ptr, i64 } %t19, 0
  %t21 = extractvalue { ptr, i64 } %t19, 1
  %t22 = icmp uge i64 %t18, %t21
  br i1 %t22, label %chk.bad10, label %chk.ok9
chk.bad10:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 68)
  unreachable
chk.ok9:
  %t23 = getelementptr inbounds i8, ptr %t20, i64 %t18
  %t24 = load i8, ptr %t23
  %t25 = call i64 @aila_std_utf8__rune_len(i8 %t24)
  store i64 %t25, ptr %slot_step_6
  %t26 = load i64, ptr %slot_step_6
  %t27 = icmp eq i64 %t26, 0
  br i1 %t27, label %if.then12, label %if.next13
if.then12:
  ret i32 4294967295
if.next13:
  br label %if.end11
if.end11:
  %t28 = load i64, ptr %slot_i_2
  %t29 = load i64, ptr %slot_step_6
  %t30 = add i64 %t28, %t29
  %t31 = load i64, ptr %slot_n_4
  %t32 = icmp ugt i64 %t30, %t31
  br i1 %t32, label %if.then15, label %if.next16
if.then15:
  ret i32 4294967295
if.next16:
  br label %if.end14
if.end14:
  store i32 0, ptr %slot_cp_7
  %t33 = load i64, ptr %slot_step_6
  %t34 = icmp eq i64 %t33, 2
  br i1 %t34, label %if.then18, label %if.next19
if.then18:
  %t35 = load i32, ptr %slot_c0_5
  %t36 = and i32 %t35, 31
  store i32 %t36, ptr %slot_cp_7
  br label %if.end17
if.next19:
  %t37 = load i64, ptr %slot_step_6
  %t38 = icmp eq i64 %t37, 3
  br i1 %t38, label %if.then20, label %if.next21
if.then20:
  %t39 = load i32, ptr %slot_c0_5
  %t40 = and i32 %t39, 15
  store i32 %t40, ptr %slot_cp_7
  br label %if.end17
if.next21:
  %t41 = load i32, ptr %slot_c0_5
  %t42 = and i32 %t41, 7
  store i32 %t42, ptr %slot_cp_7
  br label %if.end17
if.end17:
  store i64 1, ptr %slot_k_8
  br label %for.cond23
for.cond23:
  %t43 = load i64, ptr %slot_k_8
  %t44 = load i64, ptr %slot_step_6
  %t45 = icmp ult i64 %t43, %t44
  br i1 %t45, label %for.body25, label %for.end22
for.body25:
  %t46 = load i64, ptr %slot_i_2
  %t47 = load i64, ptr %slot_k_8
  %t48 = add i64 %t46, %t47
  %t49 = load { ptr, i64 }, ptr %slot_b_3
  %t50 = extractvalue { ptr, i64 } %t49, 0
  %t51 = extractvalue { ptr, i64 } %t49, 1
  %t52 = icmp uge i64 %t48, %t51
  br i1 %t52, label %chk.bad27, label %chk.ok26
chk.bad27:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 85)
  unreachable
chk.ok26:
  %t53 = getelementptr inbounds i8, ptr %t50, i64 %t48
  %t54 = load i8, ptr %t53
  %t55 = zext i8 %t54 to i32
  store i32 %t55, ptr %slot_cb_9
  %t56 = load i32, ptr %slot_cb_9
  %t57 = and i32 %t56, 192
  %t58 = icmp ne i32 %t57, 128
  br i1 %t58, label %if.then29, label %if.next30
if.then29:
  ret i32 4294967295
if.next30:
  br label %if.end28
if.end28:
  %t59 = load i32, ptr %slot_cp_7
  %t60 = icmp uge i32 6, 32
  br i1 %t60, label %chk.bad32, label %chk.ok31
chk.bad32:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 89)
  unreachable
chk.ok31:
  %t61 = shl i32 %t59, 6
  %t62 = load i32, ptr %slot_cb_9
  %t63 = and i32 %t62, 63
  %t64 = or i32 %t61, %t63
  store i32 %t64, ptr %slot_cp_7
  %t65 = load i64, ptr %slot_k_8
  %t66 = add i64 %t65, 1
  store i64 %t66, ptr %slot_k_8
  br label %for.step24
for.step24:
  br label %for.cond23
for.end22:
  %t67 = load i32, ptr %slot_cp_7
  ret i32 %t67
}

define i1 @aila_std_utf8__valid({ ptr, i64 } %p_t) {
  %slot_t_1 = alloca { ptr, i64 }, align 8
  %slot_b_2 = alloca { ptr, i64 }, align 8
  %slot_n_3 = alloca i64, align 8
  %slot_i_4 = alloca i64, align 8
  %slot_step_5 = alloca i64, align 8
  %slot_k_6 = alloca i64, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_1
  %t1 = load { ptr, i64 }, ptr %slot_t_1
  store { ptr, i64 } %t1, ptr %slot_b_2
  %t2 = load { ptr, i64 }, ptr %slot_b_2
  %t3 = extractvalue { ptr, i64 } %t2, 1
  store i64 %t3, ptr %slot_n_3
  store i64 0, ptr %slot_i_4
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
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 102)
  unreachable
chk.ok5:
  %t12 = getelementptr inbounds i8, ptr %t9, i64 %t7
  %t13 = load i8, ptr %t12
  %t14 = call i64 @aila_std_utf8__rune_len(i8 %t13)
  store i64 %t14, ptr %slot_step_5
  %t15 = load i64, ptr %slot_step_5
  %t16 = icmp eq i64 %t15, 0
  br i1 %t16, label %if.then8, label %if.next9
if.then8:
  ret i1 0
if.next9:
  br label %if.end7
if.end7:
  %t17 = load i64, ptr %slot_i_4
  %t18 = load i64, ptr %slot_step_5
  %t19 = add i64 %t17, %t18
  %t20 = load i64, ptr %slot_n_3
  %t21 = icmp ugt i64 %t19, %t20
  br i1 %t21, label %if.then11, label %if.next12
if.then11:
  ret i1 0
if.next12:
  br label %if.end10
if.end10:
  store i64 1, ptr %slot_k_6
  br label %for.cond14
for.cond14:
  %t22 = load i64, ptr %slot_k_6
  %t23 = load i64, ptr %slot_step_5
  %t24 = icmp ult i64 %t22, %t23
  br i1 %t24, label %for.body16, label %for.end13
for.body16:
  %t25 = load i64, ptr %slot_i_4
  %t26 = load i64, ptr %slot_k_6
  %t27 = add i64 %t25, %t26
  %t28 = load { ptr, i64 }, ptr %slot_b_2
  %t29 = extractvalue { ptr, i64 } %t28, 0
  %t30 = extractvalue { ptr, i64 } %t28, 1
  %t31 = icmp uge i64 %t27, %t30
  br i1 %t31, label %chk.bad21, label %chk.ok20
chk.bad21:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 111)
  unreachable
chk.ok20:
  %t32 = getelementptr inbounds i8, ptr %t29, i64 %t27
  %t33 = load i8, ptr %t32
  %t34 = zext i8 %t33 to i32
  %t35 = and i32 %t34, 192
  %t36 = icmp ne i32 %t35, 128
  br i1 %t36, label %if.then18, label %if.next19
if.then18:
  ret i1 0
if.next19:
  br label %if.end17
if.end17:
  %t37 = load i64, ptr %slot_k_6
  %t38 = add i64 %t37, 1
  store i64 %t38, ptr %slot_k_6
  br label %for.step15
for.step15:
  br label %for.cond14
for.end13:
  %t39 = load i64, ptr %slot_i_4
  %t40 = load i64, ptr %slot_step_5
  %t41 = add i64 %t39, %t40
  store i64 %t41, ptr %slot_i_4
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  ret i1 1
}
