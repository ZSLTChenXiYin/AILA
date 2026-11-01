; AILA module: std.paint
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [17 x i8] c"division by zero\00", align 1
@.str.1 = private unnamed_addr constant [15 x i8] c"std/paint.aila\00", align 1
@.str.2 = private unnamed_addr constant [25 x i8] c"shift count out of range\00", align 1
@.str.3 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1

declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare void @aila_std_pixbuf__fill(ptr, i1)
declare { i16, i16, i16, i16 } @aila_std_damage__Rect__clamp({ i16, i16, i16, i16 }, i16, i16)
declare i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 })
declare void @aila_std_pixbuf__set_px(ptr, i16, i16, i1)
declare i16 @aila_std_pixbuf__cols(ptr)

define void @aila_std_paint__transpose_page(ptr %p_pb, i16 %p_page, i16 %p_c0, i16 %p_c1, { ptr, i64 } %p_dst) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_page_2 = alloca i16, align 8
  %slot_c0_3 = alloca i16, align 8
  %slot_c1_4 = alloca i16, align 8
  %slot_dst_5 = alloca { ptr, i64 }, align 8
  %slot_rs_6 = alloca i64, align 8
  %slot_cap_7 = alloca i64, align 8
  %slot_fbcap_8 = alloca i64, align 8
  %slot_y0_9 = alloca i16, align 8
  %slot_i_10 = alloca i64, align 8
  %slot_x_11 = alloca i16, align 8
  %slot_byte_off_12 = alloca i64, align 8
  %slot_mask_13 = alloca i8, align 8
  %slot_acc_14 = alloca i8, align 8
  %slot_k_15 = alloca i16, align 8
  %slot_y_16 = alloca i16, align 8
  %slot_at_17 = alloca i64, align 8
  store i16 %p_page, ptr %slot_page_2
  store i16 %p_c0, ptr %slot_c0_3
  store i16 %p_c1, ptr %slot_c1_4
  store { ptr, i64 } %p_dst, ptr %slot_dst_5
  %t1 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 3
  %t2 = load i16, ptr %t1
  %t3 = zext i16 %t2 to i64
  store i64 %t3, ptr %slot_rs_6
  %t4 = load { ptr, i64 }, ptr %slot_dst_5
  %t5 = extractvalue { ptr, i64 } %t4, 1
  store i64 %t5, ptr %slot_cap_7
  %t7 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 0
  %t8 = load { ptr, i64 }, ptr %t7
  %t9 = extractvalue { ptr, i64 } %t8, 1
  store i64 %t9, ptr %slot_fbcap_8
  %t11 = load i16, ptr %slot_page_2
  %t12 = icmp eq i8 2, 0
  br i1 %t12, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 57)
  unreachable
chk.ok1:
  %t13 = udiv i8 16, 2
  %t14 = zext i8 %t13 to i16
  %t15 = mul i16 %t11, %t14
  store i16 %t15, ptr %slot_y0_9
  store i64 0, ptr %slot_i_10
  %t16 = load i16, ptr %slot_c0_3
  store i16 %t16, ptr %slot_x_11
  br label %for.cond4
for.cond4:
  %t17 = load i16, ptr %slot_x_11
  %t18 = load i16, ptr %slot_c1_4
  %t19 = icmp ult i16 %t17, %t18
  br i1 %t19, label %for.body6, label %for.end3
for.body6:
  %t20 = load i64, ptr %slot_i_10
  %t21 = load i64, ptr %slot_cap_7
  %t22 = icmp uge i64 %t20, %t21
  br i1 %t22, label %if.then8, label %if.next9
if.then8:
  br label %for.end3
if.next9:
  br label %if.end7
if.end7:
  %t23 = load i16, ptr %slot_x_11
  %t24 = icmp eq i16 8, 0
  br i1 %t24, label %chk.bad11, label %chk.ok10
chk.bad11:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 64)
  unreachable
chk.ok10:
  %t25 = udiv i16 %t23, 8
  %t26 = zext i16 %t25 to i64
  store i64 %t26, ptr %slot_byte_off_12
  %t27 = trunc i64 128 to i8
  %t28 = load i16, ptr %slot_x_11
  %t29 = icmp eq i16 8, 0
  br i1 %t29, label %chk.bad13, label %chk.ok12
chk.bad13:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 65)
  unreachable
chk.ok12:
  %t30 = urem i16 %t28, 8
  %t31 = trunc i16 %t30 to i8
  %t32 = icmp uge i8 %t31, 8
  br i1 %t32, label %chk.bad15, label %chk.ok14
chk.bad15:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 65)
  unreachable
chk.ok14:
  %t33 = lshr i8 %t27, %t31
  store i8 %t33, ptr %slot_mask_13
  store i8 0, ptr %slot_acc_14
  store i16 0, ptr %slot_k_15
  br label %for.cond17
for.cond17:
  %t34 = load i16, ptr %slot_k_15
  %t35 = icmp ult i16 %t34, 8
  br i1 %t35, label %for.body19, label %for.end16
for.body19:
  %t36 = load i16, ptr %slot_y0_9
  %t37 = load i16, ptr %slot_k_15
  %t38 = add i16 %t36, %t37
  store i16 %t38, ptr %slot_y_16
  %t39 = load i16, ptr %slot_y_16
  %t40 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 2
  %t41 = load i16, ptr %t40
  %t42 = icmp ult i16 %t39, %t41
  br i1 %t42, label %if.then21, label %if.next22
if.then21:
  %t43 = load i16, ptr %slot_y_16
  %t44 = zext i16 %t43 to i64
  %t45 = load i64, ptr %slot_rs_6
  %t46 = mul i64 %t44, %t45
  %t47 = load i64, ptr %slot_byte_off_12
  %t48 = add i64 %t46, %t47
  store i64 %t48, ptr %slot_at_17
  %t49 = load i64, ptr %slot_at_17
  %t50 = load i64, ptr %slot_fbcap_8
  %t51 = icmp ult i64 %t49, %t50
  br i1 %t51, label %if.then24, label %if.next25
if.then24:
  %t52 = load i64, ptr %slot_at_17
  %t53 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 0
  %t54 = load { ptr, i64 }, ptr %t53
  %t55 = extractvalue { ptr, i64 } %t54, 0
  %t56 = extractvalue { ptr, i64 } %t54, 1
  %t57 = icmp uge i64 %t52, %t56
  br i1 %t57, label %chk.bad30, label %chk.ok29
chk.bad30:
  call void @aila_rt_trap(ptr @.str.3, i64 19, ptr @.str.1, i64 73)
  unreachable
chk.ok29:
  %t58 = getelementptr inbounds i8, ptr %t55, i64 %t52
  %t59 = load i8, ptr %t58
  %t60 = load i8, ptr %slot_mask_13
  %t61 = and i8 %t59, %t60
  %t62 = icmp ne i8 %t61, 0
  br i1 %t62, label %if.then27, label %if.next28
if.then27:
  %t63 = load i8, ptr %slot_acc_14
  %t64 = trunc i64 1 to i8
  %t65 = load i16, ptr %slot_k_15
  %t66 = trunc i16 %t65 to i8
  %t67 = icmp uge i8 %t66, 8
  br i1 %t67, label %chk.bad32, label %chk.ok31
chk.bad32:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 74)
  unreachable
chk.ok31:
  %t68 = shl i8 %t64, %t66
  %t69 = or i8 %t63, %t68
  store i8 %t69, ptr %slot_acc_14
  br label %if.end26
if.next28:
  br label %if.end26
if.end26:
  br label %if.end23
if.next25:
  br label %if.end23
if.end23:
  br label %if.end20
if.next22:
  br label %if.end20
if.end20:
  %t70 = load i16, ptr %slot_k_15
  %t71 = add i16 %t70, 1
  store i16 %t71, ptr %slot_k_15
  br label %for.step18
for.step18:
  br label %for.cond17
for.end16:
  %t72 = load i64, ptr %slot_i_10
  %t73 = load { ptr, i64 }, ptr %slot_dst_5
  %t74 = extractvalue { ptr, i64 } %t73, 0
  %t75 = extractvalue { ptr, i64 } %t73, 1
  %t76 = icmp uge i64 %t72, %t75
  br i1 %t76, label %chk.bad34, label %chk.ok33
chk.bad34:
  call void @aila_rt_trap(ptr @.str.3, i64 19, ptr @.str.1, i64 80)
  unreachable
chk.ok33:
  %t77 = getelementptr inbounds i8, ptr %t74, i64 %t72
  %t78 = load i8, ptr %slot_acc_14
  store i8 %t78, ptr %t77
  %t79 = load i64, ptr %slot_i_10
  %t80 = add i64 %t79, 1
  store i64 %t80, ptr %slot_i_10
  %t81 = load i16, ptr %slot_x_11
  %t82 = add i16 %t81, 1
  store i16 %t82, ptr %slot_x_11
  br label %for.step5
for.step5:
  br label %for.cond4
for.end3:
  ret void
}

define { i16, i16, i16, i16 } @aila_std_paint__cell_rect(i16 %p_col, i16 %p_row) {
  %slot_col_1 = alloca i16, align 8
  %slot_row_2 = alloca i16, align 8
  %slot_cw_3 = alloca i16, align 8
  %slot_ch_4 = alloca i16, align 8
  %slot___lit_5 = alloca { i16, i16, i16, i16 }, align 8
  store i16 %p_col, ptr %slot_col_1
  store i16 %p_row, ptr %slot_row_2
  %t1 = zext i8 8 to i16
  store i16 %t1, ptr %slot_cw_3
  %t2 = zext i8 16 to i16
  store i16 %t2, ptr %slot_ch_4
  %t3 = load i16, ptr %slot_col_1
  %t4 = load i16, ptr %slot_cw_3
  %t5 = mul i16 %t3, %t4
  %t6 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_5, i32 0, i32 0
  store i16 %t5, ptr %t6
  %t7 = load i16, ptr %slot_row_2
  %t8 = load i16, ptr %slot_ch_4
  %t9 = mul i16 %t7, %t8
  %t10 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_5, i32 0, i32 1
  store i16 %t9, ptr %t10
  %t11 = load i16, ptr %slot_col_1
  %t12 = load i16, ptr %slot_cw_3
  %t13 = mul i16 %t11, %t12
  %t14 = load i16, ptr %slot_cw_3
  %t15 = add i16 %t13, %t14
  %t16 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_5, i32 0, i32 2
  store i16 %t15, ptr %t16
  %t17 = load i16, ptr %slot_row_2
  %t18 = load i16, ptr %slot_ch_4
  %t19 = mul i16 %t17, %t18
  %t20 = load i16, ptr %slot_ch_4
  %t21 = add i16 %t19, %t20
  %t22 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_5, i32 0, i32 3
  store i16 %t21, ptr %t22
  %t23 = load { i16, i16, i16, i16 }, ptr %slot___lit_5
  ret { i16, i16, i16, i16 } %t23
}

define { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 } @aila_std_paint__new_glyph_set({ ptr, i64 } %p_cps, { ptr, i64 } %p_bits, { ptr, i64 } %p_widths, i64 %p_count, i16 %p_px_h) {
  %slot_cps_1 = alloca { ptr, i64 }, align 8
  %slot_bits_2 = alloca { ptr, i64 }, align 8
  %slot_widths_3 = alloca { ptr, i64 }, align 8
  %slot_count_4 = alloca i64, align 8
  %slot_px_h_5 = alloca i16, align 8
  %slot___lit_6 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  store { ptr, i64 } %p_cps, ptr %slot_cps_1
  store { ptr, i64 } %p_bits, ptr %slot_bits_2
  store { ptr, i64 } %p_widths, ptr %slot_widths_3
  store i64 %p_count, ptr %slot_count_4
  store i16 %p_px_h, ptr %slot_px_h_5
  %t1 = load { ptr, i64 }, ptr %slot_cps_1
  %t2 = getelementptr inbounds { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, ptr %slot___lit_6, i32 0, i32 0
  store { ptr, i64 } %t1, ptr %t2
  %t3 = load { ptr, i64 }, ptr %slot_bits_2
  %t4 = getelementptr inbounds { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, ptr %slot___lit_6, i32 0, i32 1
  store { ptr, i64 } %t3, ptr %t4
  %t5 = load { ptr, i64 }, ptr %slot_widths_3
  %t6 = getelementptr inbounds { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, ptr %slot___lit_6, i32 0, i32 2
  store { ptr, i64 } %t5, ptr %t6
  %t7 = load i64, ptr %slot_count_4
  %t8 = getelementptr inbounds { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, ptr %slot___lit_6, i32 0, i32 3
  store i64 %t7, ptr %t8
  %t9 = load i16, ptr %slot_px_h_5
  %t10 = getelementptr inbounds { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, ptr %slot___lit_6, i32 0, i32 4
  store i16 %t9, ptr %t10
  %t11 = load { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, ptr %slot___lit_6
  ret { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 } %t11
}

define i64 @aila_std_paint__glyph_index(ptr %p_gs, i32 %p_cp) {
  %slot_gs_1 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  %slot_cp_2 = alloca i32, align 8
  %slot_i_3 = alloca i64, align 8
  store i32 %p_cp, ptr %slot_cp_2
  store i64 0, ptr %slot_i_3
  br label %for.cond2
for.cond2:
  %t1 = load i64, ptr %slot_i_3
  %t2 = getelementptr inbounds { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, ptr %p_gs, i32 0, i32 3
  %t3 = load i64, ptr %t2
  %t4 = icmp ult i64 %t1, %t3
  br i1 %t4, label %for.body4, label %for.end1
for.body4:
  %t5 = load i64, ptr %slot_i_3
  %t6 = getelementptr inbounds { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, ptr %p_gs, i32 0, i32 0
  %t7 = load { ptr, i64 }, ptr %t6
  %t8 = extractvalue { ptr, i64 } %t7, 0
  %t9 = extractvalue { ptr, i64 } %t7, 1
  %t10 = icmp uge i64 %t5, %t9
  br i1 %t10, label %chk.bad9, label %chk.ok8
chk.bad9:
  call void @aila_rt_trap(ptr @.str.3, i64 19, ptr @.str.1, i64 104)
  unreachable
chk.ok8:
  %t11 = getelementptr inbounds i32, ptr %t8, i64 %t5
  %t12 = load i32, ptr %t11
  %t13 = load i32, ptr %slot_cp_2
  %t14 = icmp eq i32 %t12, %t13
  br i1 %t14, label %if.then6, label %if.next7
if.then6:
  %t15 = load i64, ptr %slot_i_3
  ret i64 %t15
if.next7:
  br label %if.end5
if.end5:
  %t17 = load i64, ptr %slot_i_3
  %t18 = add i64 %t17, 1
  store i64 %t18, ptr %slot_i_3
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t19 = sub i64 0, 1
  ret i64 %t19
}

define i16 @aila_std_paint__glyph_width(ptr %p_gs, i32 %p_cp) {
  %slot_gs_1 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  %slot_cp_2 = alloca i32, align 8
  %slot_gi_3 = alloca i64, align 8
  store i32 %p_cp, ptr %slot_cp_2
  %t1 = load i32, ptr %slot_cp_2
  %t2 = call i64 @aila_std_paint__glyph_index(ptr %p_gs, i32 %t1)
  store i64 %t2, ptr %slot_gi_3
  %t3 = load i64, ptr %slot_gi_3
  %t4 = icmp slt i64 %t3, 0
  br i1 %t4, label %if.then2, label %if.next3
if.then2:
  ret i16 2
if.next3:
  br label %if.end1
if.end1:
  %t5 = load i64, ptr %slot_gi_3
  %t7 = getelementptr inbounds { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, ptr %p_gs, i32 0, i32 2
  %t8 = load { ptr, i64 }, ptr %t7
  %t9 = extractvalue { ptr, i64 } %t8, 0
  %t10 = extractvalue { ptr, i64 } %t8, 1
  %t11 = icmp uge i64 %t5, %t10
  br i1 %t11, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.3, i64 19, ptr @.str.1, i64 117)
  unreachable
chk.ok4:
  %t12 = getelementptr inbounds i8, ptr %t9, i64 %t5
  %t13 = load i8, ptr %t12
  %t14 = zext i8 %t13 to i16
  ret i16 %t14
}

define void @aila_std_paint__fill(ptr %p_pb, i1 %p_on) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_on_2 = alloca i1, align 8
  store i1 %p_on, ptr %slot_on_2
  %t1 = load i1, ptr %slot_on_2
  call void @aila_std_pixbuf__fill(ptr %p_pb, i1 %t1)
  ret void
}

define void @aila_std_paint__fill_rect(ptr %p_pb, { i16, i16, i16, i16 } %p_r, i1 %p_on) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_r_2 = alloca { i16, i16, i16, i16 }, align 8
  %slot_on_3 = alloca i1, align 8
  %slot_cr_4 = alloca { i16, i16, i16, i16 }, align 8
  %slot_y_5 = alloca i16, align 8
  %slot_x_6 = alloca i16, align 8
  store { i16, i16, i16, i16 } %p_r, ptr %slot_r_2
  store i1 %p_on, ptr %slot_on_3
  %t1 = load { i16, i16, i16, i16 }, ptr %slot_r_2
  %t2 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 1
  %t3 = load i16, ptr %t2
  %t4 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 2
  %t5 = load i16, ptr %t4
  %t6 = call { i16, i16, i16, i16 } @aila_std_damage__Rect__clamp({ i16, i16, i16, i16 } %t1, i16 %t3, i16 %t5)
  store { i16, i16, i16, i16 } %t6, ptr %slot_cr_4
  %t7 = load { i16, i16, i16, i16 }, ptr %slot_cr_4
  %t8 = call i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 } %t7)
  br i1 %t8, label %if.then2, label %if.next3
if.then2:
  ret void
if.next3:
  br label %if.end1
if.end1:
  %t9 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_cr_4, i32 0, i32 1
  %t10 = load i16, ptr %t9
  store i16 %t10, ptr %slot_y_5
  br label %for.cond5
for.cond5:
  %t11 = load i16, ptr %slot_y_5
  %t12 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_cr_4, i32 0, i32 3
  %t13 = load i16, ptr %t12
  %t14 = icmp ult i16 %t11, %t13
  br i1 %t14, label %for.body7, label %for.end4
for.body7:
  %t15 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_cr_4, i32 0, i32 0
  %t16 = load i16, ptr %t15
  store i16 %t16, ptr %slot_x_6
  br label %for.cond9
for.cond9:
  %t17 = load i16, ptr %slot_x_6
  %t18 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_cr_4, i32 0, i32 2
  %t19 = load i16, ptr %t18
  %t20 = icmp ult i16 %t17, %t19
  br i1 %t20, label %for.body11, label %for.end8
for.body11:
  %t21 = load i16, ptr %slot_x_6
  %t22 = load i16, ptr %slot_y_5
  %t23 = load i1, ptr %slot_on_3
  call void @aila_std_pixbuf__set_px(ptr %p_pb, i16 %t21, i16 %t22, i1 %t23)
  %t24 = load i16, ptr %slot_x_6
  %t25 = add i16 %t24, 1
  store i16 %t25, ptr %slot_x_6
  br label %for.step10
for.step10:
  br label %for.cond9
for.end8:
  %t26 = load i16, ptr %slot_y_5
  %t27 = add i16 %t26, 1
  store i16 %t27, ptr %slot_y_5
  br label %for.step6
for.step6:
  br label %for.cond5
for.end4:
  ret void
}

define void @aila_std_paint__hline(ptr %p_pb, i16 %p_x0, i16 %p_x1, i16 %p_y, i1 %p_on) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_x0_2 = alloca i16, align 8
  %slot_x1_3 = alloca i16, align 8
  %slot_y_4 = alloca i16, align 8
  %slot_on_5 = alloca i1, align 8
  %slot___lit_6 = alloca { i16, i16, i16, i16 }, align 8
  store i16 %p_x0, ptr %slot_x0_2
  store i16 %p_x1, ptr %slot_x1_3
  store i16 %p_y, ptr %slot_y_4
  store i1 %p_on, ptr %slot_on_5
  %t1 = load i16, ptr %slot_x0_2
  %t2 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_6, i32 0, i32 0
  store i16 %t1, ptr %t2
  %t3 = load i16, ptr %slot_y_4
  %t4 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_6, i32 0, i32 1
  store i16 %t3, ptr %t4
  %t5 = load i16, ptr %slot_x1_3
  %t6 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_6, i32 0, i32 2
  store i16 %t5, ptr %t6
  %t7 = load i16, ptr %slot_y_4
  %t8 = add i16 %t7, 1
  %t9 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_6, i32 0, i32 3
  store i16 %t8, ptr %t9
  %t10 = load { i16, i16, i16, i16 }, ptr %slot___lit_6
  %t11 = load i1, ptr %slot_on_5
  call void @aila_std_paint__fill_rect(ptr %p_pb, { i16, i16, i16, i16 } %t10, i1 %t11)
  ret void
}

define void @aila_std_paint__vline(ptr %p_pb, i16 %p_x, i16 %p_y0, i16 %p_y1, i1 %p_on) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_x_2 = alloca i16, align 8
  %slot_y0_3 = alloca i16, align 8
  %slot_y1_4 = alloca i16, align 8
  %slot_on_5 = alloca i1, align 8
  %slot___lit_6 = alloca { i16, i16, i16, i16 }, align 8
  store i16 %p_x, ptr %slot_x_2
  store i16 %p_y0, ptr %slot_y0_3
  store i16 %p_y1, ptr %slot_y1_4
  store i1 %p_on, ptr %slot_on_5
  %t1 = load i16, ptr %slot_x_2
  %t2 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_6, i32 0, i32 0
  store i16 %t1, ptr %t2
  %t3 = load i16, ptr %slot_y0_3
  %t4 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_6, i32 0, i32 1
  store i16 %t3, ptr %t4
  %t5 = load i16, ptr %slot_x_2
  %t6 = add i16 %t5, 1
  %t7 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_6, i32 0, i32 2
  store i16 %t6, ptr %t7
  %t8 = load i16, ptr %slot_y1_4
  %t9 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_6, i32 0, i32 3
  store i16 %t8, ptr %t9
  %t10 = load { i16, i16, i16, i16 }, ptr %slot___lit_6
  %t11 = load i1, ptr %slot_on_5
  call void @aila_std_paint__fill_rect(ptr %p_pb, { i16, i16, i16, i16 } %t10, i1 %t11)
  ret void
}

define i16 @aila_std_paint__blit_glyph(ptr %p_pb, ptr %p_gs, i32 %p_cp, i16 %p_col, i16 %p_row) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_gs_2 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  %slot_cp_3 = alloca i32, align 8
  %slot_col_4 = alloca i16, align 8
  %slot_row_5 = alloca i16, align 8
  %slot_gi_6 = alloca i64, align 8
  %slot_wd_7 = alloca i16, align 8
  %slot_x0_8 = alloca i16, align 8
  %slot_y0_9 = alloca i16, align 8
  %slot_yy_10 = alloca i64, align 8
  %slot_hi_11 = alloca i8, align 8
  %slot_lo_12 = alloca i8, align 8
  %slot_base_13 = alloca i64, align 8
  %slot_c_14 = alloca i16, align 8
  %slot_bitc_15 = alloca i8, align 8
  store i32 %p_cp, ptr %slot_cp_3
  store i16 %p_col, ptr %slot_col_4
  store i16 %p_row, ptr %slot_row_5
  %t1 = load i32, ptr %slot_cp_3
  %t2 = call i64 @aila_std_paint__glyph_index(ptr %p_gs, i32 %t1)
  store i64 %t2, ptr %slot_gi_6
  store i16 2, ptr %slot_wd_7
  %t3 = load i64, ptr %slot_gi_6
  %t4 = icmp sge i64 %t3, 0
  br i1 %t4, label %if.then2, label %if.next3
if.then2:
  %t5 = load i64, ptr %slot_gi_6
  %t7 = getelementptr inbounds { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, ptr %p_gs, i32 0, i32 2
  %t8 = load { ptr, i64 }, ptr %t7
  %t9 = extractvalue { ptr, i64 } %t8, 0
  %t10 = extractvalue { ptr, i64 } %t8, 1
  %t11 = icmp uge i64 %t5, %t10
  br i1 %t11, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.3, i64 19, ptr @.str.1, i64 152)
  unreachable
chk.ok4:
  %t12 = getelementptr inbounds i8, ptr %t9, i64 %t5
  %t13 = load i8, ptr %t12
  %t14 = zext i8 %t13 to i16
  store i16 %t14, ptr %slot_wd_7
  br label %if.end1
if.next3:
  br label %if.end1
if.end1:
  %t15 = load i16, ptr %slot_wd_7
  %t16 = icmp eq i16 %t15, 0
  br i1 %t16, label %if.then7, label %if.next8
if.then7:
  store i16 2, ptr %slot_wd_7
  br label %if.end6
if.next8:
  br label %if.end6
if.end6:
  %t17 = load i16, ptr %slot_col_4
  %t18 = zext i8 8 to i16
  %t19 = mul i16 %t17, %t18
  store i16 %t19, ptr %slot_x0_8
  %t20 = load i16, ptr %slot_row_5
  %t21 = zext i8 16 to i16
  %t22 = mul i16 %t20, %t21
  store i16 %t22, ptr %slot_y0_9
  store i64 0, ptr %slot_yy_10
  br label %for.cond10
for.cond10:
  %t23 = load i64, ptr %slot_yy_10
  %t24 = zext i8 16 to i64
  %t25 = icmp ult i64 %t23, %t24
  br i1 %t25, label %for.body12, label %for.end9
for.body12:
  store i8 255, ptr %slot_hi_11
  store i8 255, ptr %slot_lo_12
  %t26 = load i64, ptr %slot_gi_6
  %t27 = icmp sge i64 %t26, 0
  br i1 %t27, label %if.then14, label %if.next15
if.then14:
  %t28 = load i64, ptr %slot_gi_6
  %t30 = mul i64 %t28, 32
  %t31 = load i64, ptr %slot_yy_10
  %t32 = mul i64 %t31, 2
  %t33 = add i64 %t30, %t32
  store i64 %t33, ptr %slot_base_13
  %t34 = load i64, ptr %slot_base_13
  %t35 = getelementptr inbounds { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, ptr %p_gs, i32 0, i32 1
  %t36 = load { ptr, i64 }, ptr %t35
  %t37 = extractvalue { ptr, i64 } %t36, 0
  %t38 = extractvalue { ptr, i64 } %t36, 1
  %t39 = icmp uge i64 %t34, %t38
  br i1 %t39, label %chk.bad17, label %chk.ok16
chk.bad17:
  call void @aila_rt_trap(ptr @.str.3, i64 19, ptr @.str.1, i64 165)
  unreachable
chk.ok16:
  %t40 = getelementptr inbounds i8, ptr %t37, i64 %t34
  %t41 = load i8, ptr %t40
  store i8 %t41, ptr %slot_hi_11
  %t42 = load i64, ptr %slot_base_13
  %t43 = add i64 %t42, 1
  %t44 = getelementptr inbounds { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, ptr %p_gs, i32 0, i32 1
  %t45 = load { ptr, i64 }, ptr %t44
  %t46 = extractvalue { ptr, i64 } %t45, 0
  %t47 = extractvalue { ptr, i64 } %t45, 1
  %t48 = icmp uge i64 %t43, %t47
  br i1 %t48, label %chk.bad19, label %chk.ok18
chk.bad19:
  call void @aila_rt_trap(ptr @.str.3, i64 19, ptr @.str.1, i64 166)
  unreachable
chk.ok18:
  %t49 = getelementptr inbounds i8, ptr %t46, i64 %t43
  %t50 = load i8, ptr %t49
  store i8 %t50, ptr %slot_lo_12
  br label %if.end13
if.next15:
  br label %if.end13
if.end13:
  store i16 0, ptr %slot_c_14
  br label %for.cond21
for.cond21:
  %t51 = load i16, ptr %slot_c_14
  %t52 = icmp ult i16 %t51, 8
  br i1 %t52, label %for.body23, label %for.end20
for.body23:
  %t53 = trunc i64 128 to i8
  %t54 = load i16, ptr %slot_c_14
  %t55 = trunc i16 %t54 to i8
  %t56 = icmp uge i8 %t55, 8
  br i1 %t56, label %chk.bad25, label %chk.ok24
chk.bad25:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 170)
  unreachable
chk.ok24:
  %t57 = lshr i8 %t53, %t55
  store i8 %t57, ptr %slot_bitc_15
  %t58 = load i8, ptr %slot_hi_11
  %t59 = load i8, ptr %slot_bitc_15
  %t60 = and i8 %t58, %t59
  %t61 = icmp ne i8 %t60, 0
  br i1 %t61, label %if.then27, label %if.next28
if.then27:
  %t62 = load i16, ptr %slot_x0_8
  %t63 = load i16, ptr %slot_c_14
  %t64 = add i16 %t62, %t63
  %t65 = load i16, ptr %slot_y0_9
  %t66 = load i64, ptr %slot_yy_10
  %t67 = trunc i64 %t66 to i16
  %t68 = add i16 %t65, %t67
  call void @aila_std_pixbuf__set_px(ptr %p_pb, i16 %t64, i16 %t68, i1 1)
  br label %if.end26
if.next28:
  br label %if.end26
if.end26:
  %t69 = load i16, ptr %slot_wd_7
  %t70 = icmp ugt i16 %t69, 1
  br i1 %t70, label %if.then30, label %if.next31
if.then30:
  %t71 = load i8, ptr %slot_lo_12
  %t72 = load i8, ptr %slot_bitc_15
  %t73 = and i8 %t71, %t72
  %t74 = icmp ne i8 %t73, 0
  br i1 %t74, label %if.then33, label %if.next34
if.then33:
  %t75 = load i16, ptr %slot_x0_8
  %t76 = add i16 %t75, 8
  %t77 = load i16, ptr %slot_c_14
  %t78 = add i16 %t76, %t77
  %t79 = load i16, ptr %slot_y0_9
  %t80 = load i64, ptr %slot_yy_10
  %t81 = trunc i64 %t80 to i16
  %t82 = add i16 %t79, %t81
  call void @aila_std_pixbuf__set_px(ptr %p_pb, i16 %t78, i16 %t82, i1 1)
  br label %if.end32
if.next34:
  br label %if.end32
if.end32:
  br label %if.end29
if.next31:
  br label %if.end29
if.end29:
  %t83 = load i16, ptr %slot_c_14
  %t84 = add i16 %t83, 1
  store i16 %t84, ptr %slot_c_14
  br label %for.step22
for.step22:
  br label %for.cond21
for.end20:
  %t85 = load i64, ptr %slot_yy_10
  %t86 = add i64 %t85, 1
  store i64 %t86, ptr %slot_yy_10
  br label %for.step11
for.step11:
  br label %for.cond10
for.end9:
  %t87 = load i16, ptr %slot_wd_7
  ret i16 %t87
}

define i16 @aila_std_paint__text_cells(ptr %p_pb, ptr %p_gs, { ptr, i64 } %p_bytes, i64 %p_n, i16 %p_col, i16 %p_row) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_gs_2 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  %slot_bytes_3 = alloca { ptr, i64 }, align 8
  %slot_n_4 = alloca i64, align 8
  %slot_col_5 = alloca i16, align 8
  %slot_row_6 = alloca i16, align 8
  %slot_cols_7 = alloca i16, align 8
  %slot_i_8 = alloca i64, align 8
  %slot_x_9 = alloca i16, align 8
  %slot_cp_10 = alloca i32, align 8
  %slot_step_11 = alloca i64, align 8
  %slot_wd_12 = alloca i16, align 8
  %slot___destr_13 = alloca { i32, i64 }, align 8
  store { ptr, i64 } %p_bytes, ptr %slot_bytes_3
  store i64 %p_n, ptr %slot_n_4
  store i16 %p_col, ptr %slot_col_5
  store i16 %p_row, ptr %slot_row_6
  %t1 = call i16 @aila_std_pixbuf__cols(ptr %p_pb)
  store i16 %t1, ptr %slot_cols_7
  store i64 0, ptr %slot_i_8
  %t2 = load i16, ptr %slot_col_5
  store i16 %t2, ptr %slot_x_9
  br label %for.cond2
for.cond2:
  %t3 = load i64, ptr %slot_i_8
  %t4 = load i64, ptr %slot_n_4
  %t5 = icmp ult i64 %t3, %t4
  br i1 %t5, label %for.body4, label %for.end1
for.body4:
  %t6 = load i16, ptr %slot_x_9
  %t7 = load i16, ptr %slot_cols_7
  %t8 = icmp uge i16 %t6, %t7
  br i1 %t8, label %if.then6, label %if.next7
if.then6:
  br label %for.end1
if.next7:
  br label %if.end5
if.end5:
  %t9 = load { ptr, i64 }, ptr %slot_bytes_3
  %t10 = load i64, ptr %slot_n_4
  %t11 = load i64, ptr %slot_i_8
  %t12 = call { i32, i64 } @aila_std_paint__decode({ ptr, i64 } %t9, i64 %t10, i64 %t11)
  store { i32, i64 } %t12, ptr %slot___destr_13
  %t13 = getelementptr inbounds { i32, i64 }, ptr %slot___destr_13, i32 0, i32 0
  %t14 = load i32, ptr %t13
  store i32 %t14, ptr %slot_cp_10
  %t15 = getelementptr inbounds { i32, i64 }, ptr %slot___destr_13, i32 0, i32 1
  %t16 = load i64, ptr %t15
  store i64 %t16, ptr %slot_step_11
  %t17 = load i64, ptr %slot_step_11
  %t18 = icmp eq i64 %t17, 0
  br i1 %t18, label %if.then9, label %if.next10
if.then9:
  br label %for.end1
if.next10:
  br label %if.end8
if.end8:
  %t19 = load i32, ptr %slot_cp_10
  %t20 = call i16 @aila_std_paint__glyph_width(ptr %p_gs, i32 %t19)
  store i16 %t20, ptr %slot_wd_12
  %t21 = load i16, ptr %slot_x_9
  %t22 = load i16, ptr %slot_wd_12
  %t23 = add i16 %t21, %t22
  %t24 = load i16, ptr %slot_cols_7
  %t25 = icmp ugt i16 %t23, %t24
  br i1 %t25, label %if.then12, label %if.next13
if.then12:
  br label %for.end1
if.next13:
  br label %if.end11
if.end11:
  %t26 = load i32, ptr %slot_cp_10
  %t27 = load i16, ptr %slot_x_9
  %t28 = load i16, ptr %slot_row_6
  %t29 = call i16 @aila_std_paint__blit_glyph(ptr %p_pb, ptr %p_gs, i32 %t26, i16 %t27, i16 %t28)
  %t30 = load i16, ptr %slot_x_9
  %t31 = load i16, ptr %slot_wd_12
  %t32 = add i16 %t30, %t31
  store i16 %t32, ptr %slot_x_9
  %t33 = load i64, ptr %slot_i_8
  %t34 = load i64, ptr %slot_step_11
  %t35 = add i64 %t33, %t34
  store i64 %t35, ptr %slot_i_8
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t36 = load i16, ptr %slot_x_9
  ret i16 %t36
}

define { i32, i64 } @aila_std_paint__decode({ ptr, i64 } %p_bytes, i64 %p_n, i64 %p_i) {
  %slot_bytes_1 = alloca { ptr, i64 }, align 8
  %slot_n_2 = alloca i64, align 8
  %slot_i_3 = alloca i64, align 8
  %slot_b0_4 = alloca i32, align 8
  %slot_need_5 = alloca i64, align 8
  %slot_cp_6 = alloca i32, align 8
  %slot_k_7 = alloca i64, align 8
  %slot_cb_8 = alloca i32, align 8
  store { ptr, i64 } %p_bytes, ptr %slot_bytes_1
  store i64 %p_n, ptr %slot_n_2
  store i64 %p_i, ptr %slot_i_3
  %t1 = load i64, ptr %slot_i_3
  %t2 = load i64, ptr %slot_n_2
  %t3 = icmp uge i64 %t1, %t2
  br i1 %t3, label %if.then2, label %if.next3
if.then2:
  %t4 = trunc i64 0 to i32
  %t5 = insertvalue { i32, i64 } undef, i32 %t4, 0
  %t6 = insertvalue { i32, i64 } %t5, i64 0, 1
  ret { i32, i64 } %t6
if.next3:
  br label %if.end1
if.end1:
  %t7 = load i64, ptr %slot_i_3
  %t8 = load { ptr, i64 }, ptr %slot_bytes_1
  %t9 = extractvalue { ptr, i64 } %t8, 0
  %t10 = extractvalue { ptr, i64 } %t8, 1
  %t11 = icmp uge i64 %t7, %t10
  br i1 %t11, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.3, i64 19, ptr @.str.1, i64 214)
  unreachable
chk.ok4:
  %t12 = getelementptr inbounds i8, ptr %t9, i64 %t7
  %t13 = load i8, ptr %t12
  %t14 = zext i8 %t13 to i32
  store i32 %t14, ptr %slot_b0_4
  %t15 = load i32, ptr %slot_b0_4
  %t16 = icmp ult i32 %t15, 128
  br i1 %t16, label %if.then7, label %if.next8
if.then7:
  %t17 = load i32, ptr %slot_b0_4
  %t18 = insertvalue { i32, i64 } undef, i32 %t17, 0
  %t19 = insertvalue { i32, i64 } %t18, i64 1, 1
  ret { i32, i64 } %t19
if.next8:
  br label %if.end6
if.end6:
  store i64 0, ptr %slot_need_5
  store i32 0, ptr %slot_cp_6
  %t20 = load i32, ptr %slot_b0_4
  %t21 = icmp uge i32 %t20, 194
  br i1 %t21, label %if.then10, label %if.next11
if.then10:
  %t22 = load i32, ptr %slot_b0_4
  %t23 = icmp ult i32 %t22, 224
  br i1 %t23, label %if.then13, label %if.next14
if.then13:
  store i64 1, ptr %slot_need_5
  %t24 = load i32, ptr %slot_b0_4
  %t25 = and i32 %t24, 31
  store i32 %t25, ptr %slot_cp_6
  br label %if.end12
if.next14:
  %t26 = load i32, ptr %slot_b0_4
  %t27 = icmp ult i32 %t26, 240
  br i1 %t27, label %if.then15, label %if.next16
if.then15:
  store i64 2, ptr %slot_need_5
  %t28 = load i32, ptr %slot_b0_4
  %t29 = and i32 %t28, 15
  store i32 %t29, ptr %slot_cp_6
  br label %if.end12
if.next16:
  %t30 = load i32, ptr %slot_b0_4
  %t31 = icmp ult i32 %t30, 245
  br i1 %t31, label %if.then17, label %if.next18
if.then17:
  store i64 3, ptr %slot_need_5
  %t32 = load i32, ptr %slot_b0_4
  %t33 = and i32 %t32, 7
  store i32 %t33, ptr %slot_cp_6
  br label %if.end12
if.next18:
  br label %if.end12
if.end12:
  br label %if.end9
if.next11:
  br label %if.end9
if.end9:
  %t34 = load i64, ptr %slot_need_5
  %t35 = icmp eq i64 %t34, 0
  br i1 %t35, label %if.then20, label %if.next21
if.then20:
  %t36 = trunc i64 0 to i32
  %t37 = insertvalue { i32, i64 } undef, i32 %t36, 0
  %t38 = insertvalue { i32, i64 } %t37, i64 0, 1
  ret { i32, i64 } %t38
if.next21:
  br label %if.end19
if.end19:
  %t39 = load i64, ptr %slot_i_3
  %t40 = load i64, ptr %slot_need_5
  %t41 = add i64 %t39, %t40
  %t42 = load i64, ptr %slot_n_2
  %t43 = icmp uge i64 %t41, %t42
  br i1 %t43, label %if.then23, label %if.next24
if.then23:
  %t44 = trunc i64 0 to i32
  %t45 = insertvalue { i32, i64 } undef, i32 %t44, 0
  %t46 = insertvalue { i32, i64 } %t45, i64 0, 1
  ret { i32, i64 } %t46
if.next24:
  br label %if.end22
if.end22:
  store i64 1, ptr %slot_k_7
  br label %for.cond26
for.cond26:
  %t47 = load i64, ptr %slot_k_7
  %t48 = load i64, ptr %slot_need_5
  %t49 = icmp ule i64 %t47, %t48
  br i1 %t49, label %for.body28, label %for.end25
for.body28:
  %t50 = load i64, ptr %slot_i_3
  %t51 = load i64, ptr %slot_k_7
  %t52 = add i64 %t50, %t51
  %t53 = load { ptr, i64 }, ptr %slot_bytes_1
  %t54 = extractvalue { ptr, i64 } %t53, 0
  %t55 = extractvalue { ptr, i64 } %t53, 1
  %t56 = icmp uge i64 %t52, %t55
  br i1 %t56, label %chk.bad30, label %chk.ok29
chk.bad30:
  call void @aila_rt_trap(ptr @.str.3, i64 19, ptr @.str.1, i64 240)
  unreachable
chk.ok29:
  %t57 = getelementptr inbounds i8, ptr %t54, i64 %t52
  %t58 = load i8, ptr %t57
  %t59 = zext i8 %t58 to i32
  store i32 %t59, ptr %slot_cb_8
  %t60 = load i32, ptr %slot_cb_8
  %t61 = and i32 %t60, 192
  %t62 = icmp ne i32 %t61, 128
  br i1 %t62, label %if.then32, label %if.next33
if.then32:
  %t63 = trunc i64 0 to i32
  %t64 = insertvalue { i32, i64 } undef, i32 %t63, 0
  %t65 = insertvalue { i32, i64 } %t64, i64 0, 1
  ret { i32, i64 } %t65
if.next33:
  br label %if.end31
if.end31:
  %t66 = load i32, ptr %slot_cp_6
  %t67 = icmp uge i32 6, 32
  br i1 %t67, label %chk.bad35, label %chk.ok34
chk.bad35:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 244)
  unreachable
chk.ok34:
  %t68 = shl i32 %t66, 6
  %t69 = load i32, ptr %slot_cb_8
  %t70 = and i32 %t69, 63
  %t71 = or i32 %t68, %t70
  store i32 %t71, ptr %slot_cp_6
  %t72 = load i64, ptr %slot_k_7
  %t73 = add i64 %t72, 1
  store i64 %t73, ptr %slot_k_7
  br label %for.step27
for.step27:
  br label %for.cond26
for.end25:
  %t74 = load i32, ptr %slot_cp_6
  %t75 = load i64, ptr %slot_need_5
  %t76 = add i64 %t75, 1
  %t77 = insertvalue { i32, i64 } undef, i32 %t74, 0
  %t78 = insertvalue { i32, i64 } %t77, i64 %t76, 1
  ret { i32, i64 } %t78
}
