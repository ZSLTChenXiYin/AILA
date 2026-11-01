; AILA module: std.pixbuf
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.1 = private unnamed_addr constant [16 x i8] c"std/pixbuf.aila\00", align 1
@.str.2 = private unnamed_addr constant [17 x i8] c"division by zero\00", align 1


declare { i8, i8, i8, i8 } @aila_std_pixfmt__mono1()
declare i16 @aila_std_pixfmt__stride_for({ i8, i8, i8, i8 }, i16)
declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare i16 @aila_std_pixfmt__byte_of({ i8, i8, i8, i8 }, i16)
declare i8 @aila_std_pixfmt__bit_of({ i8, i8, i8, i8 }, i16)

define { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } @aila_std_pixbuf__new({ ptr, i64 } %p_data, i16 %p_w, i16 %p_h) {
  %slot_data_1 = alloca { ptr, i64 }, align 8
  %slot_w_2 = alloca i16, align 8
  %slot_h_3 = alloca i16, align 8
  store { ptr, i64 } %p_data, ptr %slot_data_1
  store i16 %p_w, ptr %slot_w_2
  store i16 %p_h, ptr %slot_h_3
  %t1 = load { ptr, i64 }, ptr %slot_data_1
  %t2 = load i16, ptr %slot_w_2
  %t3 = load i16, ptr %slot_h_3
  %t4 = call { i8, i8, i8, i8 } @aila_std_pixfmt__mono1()
  %t5 = call { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } @aila_std_pixbuf__new_of({ ptr, i64 } %t1, i16 %t2, i16 %t3, { i8, i8, i8, i8 } %t4)
  ret { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } %t5
}

define { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } @aila_std_pixbuf__new_of({ ptr, i64 } %p_data, i16 %p_w, i16 %p_h, { i8, i8, i8, i8 } %p_fmt) {
  %slot_data_1 = alloca { ptr, i64 }, align 8
  %slot_w_2 = alloca i16, align 8
  %slot_h_3 = alloca i16, align 8
  %slot_fmt_4 = alloca { i8, i8, i8, i8 }, align 8
  %slot_st_5 = alloca i16, align 8
  %slot___lit_6 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  store { ptr, i64 } %p_data, ptr %slot_data_1
  store i16 %p_w, ptr %slot_w_2
  store i16 %p_h, ptr %slot_h_3
  store { i8, i8, i8, i8 } %p_fmt, ptr %slot_fmt_4
  %t1 = load { i8, i8, i8, i8 }, ptr %slot_fmt_4
  %t2 = load i16, ptr %slot_w_2
  %t3 = call i16 @aila_std_pixfmt__stride_for({ i8, i8, i8, i8 } %t1, i16 %t2)
  store i16 %t3, ptr %slot_st_5
  %t4 = load { ptr, i64 }, ptr %slot_data_1
  %t5 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot___lit_6, i32 0, i32 0
  store { ptr, i64 } %t4, ptr %t5
  %t6 = load i16, ptr %slot_w_2
  %t7 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot___lit_6, i32 0, i32 1
  store i16 %t6, ptr %t7
  %t8 = load i16, ptr %slot_h_3
  %t9 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot___lit_6, i32 0, i32 2
  store i16 %t8, ptr %t9
  %t10 = load i16, ptr %slot_st_5
  %t11 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot___lit_6, i32 0, i32 3
  store i16 %t10, ptr %t11
  %t12 = load { i8, i8, i8, i8 }, ptr %slot_fmt_4
  %t13 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot___lit_6, i32 0, i32 4
  store { i8, i8, i8, i8 } %t12, ptr %t13
  %t14 = load { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot___lit_6
  ret { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } %t14
}

define i64 @aila_std_pixbuf__cap_bytes(i16 %p_w, i16 %p_h) {
  %slot_w_1 = alloca i16, align 8
  %slot_h_2 = alloca i16, align 8
  %slot_st_3 = alloca i64, align 8
  store i16 %p_w, ptr %slot_w_1
  store i16 %p_h, ptr %slot_h_2
  %t1 = call { i8, i8, i8, i8 } @aila_std_pixfmt__mono1()
  %t2 = load i16, ptr %slot_w_1
  %t3 = call i16 @aila_std_pixfmt__stride_for({ i8, i8, i8, i8 } %t1, i16 %t2)
  %t4 = zext i16 %t3 to i64
  store i64 %t4, ptr %slot_st_3
  %t5 = load i64, ptr %slot_st_3
  %t6 = load i16, ptr %slot_h_2
  %t7 = zext i16 %t6 to i64
  %t8 = mul i64 %t5, %t7
  ret i64 %t8
}

define i64 @aila_std_pixbuf__row_bytes(ptr %p_pb) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %t1 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 3
  %t2 = load i16, ptr %t1
  %t3 = zext i16 %t2 to i64
  ret i64 %t3
}

define void @aila_std_pixbuf__clear(ptr %p_pb) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  call void @aila_std_pixbuf__fill(ptr %p_pb, i1 0)
  ret void
}

define void @aila_std_pixbuf__fill(ptr %p_pb, i1 %p_on) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_on_2 = alloca i1, align 8
  %slot_v_3 = alloca i8, align 8
  %slot_byte_4 = alloca i8, align 8
  %slot_n_5 = alloca i64, align 8
  %slot_len_6 = alloca i64, align 8
  %slot_i_7 = alloca i64, align 8
  store i1 %p_on, ptr %slot_on_2
  store i8 0, ptr %slot_v_3
  %t1 = load i8, ptr %slot_v_3
  store i8 %t1, ptr %slot_byte_4
  %t2 = load i1, ptr %slot_on_2
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  store i8 255, ptr %slot_byte_4
  br label %if.end1
if.next3:
  br label %if.end1
if.end1:
  %t3 = call i64 @aila_std_pixbuf__row_bytes(ptr %p_pb)
  %t4 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 2
  %t5 = load i16, ptr %t4
  %t6 = zext i16 %t5 to i64
  %t7 = mul i64 %t3, %t6
  store i64 %t7, ptr %slot_n_5
  %t8 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 0
  %t9 = load { ptr, i64 }, ptr %t8
  %t10 = extractvalue { ptr, i64 } %t9, 1
  store i64 %t10, ptr %slot_len_6
  store i64 0, ptr %slot_i_7
  br label %for.cond5
for.cond5:
  %t12 = load i64, ptr %slot_i_7
  %t13 = load i64, ptr %slot_n_5
  %t14 = icmp ult i64 %t12, %t13
  br i1 %t14, label %for.body7, label %for.end4
for.body7:
  %t15 = load i64, ptr %slot_i_7
  %t16 = load i64, ptr %slot_len_6
  %t17 = icmp uge i64 %t15, %t16
  br i1 %t17, label %if.then9, label %if.next10
if.then9:
  br label %for.end4
if.next10:
  br label %if.end8
if.end8:
  %t18 = load i64, ptr %slot_i_7
  %t19 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 0
  %t20 = load { ptr, i64 }, ptr %t19
  %t21 = extractvalue { ptr, i64 } %t20, 0
  %t22 = extractvalue { ptr, i64 } %t20, 1
  %t23 = icmp uge i64 %t18, %t22
  br i1 %t23, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 89)
  unreachable
chk.ok11:
  %t24 = getelementptr inbounds i8, ptr %t21, i64 %t18
  %t25 = load i8, ptr %slot_byte_4
  store i8 %t25, ptr %t24
  %t26 = load i64, ptr %slot_i_7
  %t27 = add i64 %t26, 1
  store i64 %t27, ptr %slot_i_7
  br label %for.step6
for.step6:
  br label %for.cond5
for.end4:
  ret void
}

define i1 @aila_std_pixbuf__in_bounds(ptr %p_pb, i16 %p_x, i16 %p_y) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_x_2 = alloca i16, align 8
  %slot_y_3 = alloca i16, align 8
  store i16 %p_x, ptr %slot_x_2
  store i16 %p_y, ptr %slot_y_3
  %t1 = load i16, ptr %slot_x_2
  %t2 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 1
  %t3 = load i16, ptr %t2
  %t4 = icmp uge i16 %t1, %t3
  br i1 %t4, label %if.then2, label %if.next3
if.then2:
  ret i1 0
if.next3:
  br label %if.end1
if.end1:
  %t5 = load i16, ptr %slot_y_3
  %t6 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 2
  %t7 = load i16, ptr %t6
  %t8 = icmp uge i16 %t5, %t7
  br i1 %t8, label %if.then5, label %if.next6
if.then5:
  ret i1 0
if.next6:
  br label %if.end4
if.end4:
  ret i1 1
}

define i64 @aila_std_pixbuf__byte_index(ptr %p_pb, i16 %p_x, i16 %p_y) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_x_2 = alloca i16, align 8
  %slot_y_3 = alloca i16, align 8
  %slot_row_4 = alloca i64, align 8
  %slot_col_5 = alloca i64, align 8
  store i16 %p_x, ptr %slot_x_2
  store i16 %p_y, ptr %slot_y_3
  %t1 = load i16, ptr %slot_y_3
  %t2 = zext i16 %t1 to i64
  %t3 = call i64 @aila_std_pixbuf__row_bytes(ptr %p_pb)
  %t4 = mul i64 %t2, %t3
  store i64 %t4, ptr %slot_row_4
  %t5 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 4
  %t6 = load { i8, i8, i8, i8 }, ptr %t5
  %t7 = load i16, ptr %slot_x_2
  %t8 = call i16 @aila_std_pixfmt__byte_of({ i8, i8, i8, i8 } %t6, i16 %t7)
  %t9 = zext i16 %t8 to i64
  store i64 %t9, ptr %slot_col_5
  %t10 = load i64, ptr %slot_row_4
  %t11 = load i64, ptr %slot_col_5
  %t12 = add i64 %t10, %t11
  ret i64 %t12
}

define void @aila_std_pixbuf__set_px(ptr %p_pb, i16 %p_x, i16 %p_y, i1 %p_on) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_x_2 = alloca i16, align 8
  %slot_y_3 = alloca i16, align 8
  %slot_on_4 = alloca i1, align 8
  %slot_bi_5 = alloca i64, align 8
  %slot_len_6 = alloca i64, align 8
  %slot_mask_7 = alloca i8, align 8
  store i16 %p_x, ptr %slot_x_2
  store i16 %p_y, ptr %slot_y_3
  store i1 %p_on, ptr %slot_on_4
  %t1 = load i16, ptr %slot_x_2
  %t2 = load i16, ptr %slot_y_3
  %t3 = call i1 @aila_std_pixbuf__in_bounds(ptr %p_pb, i16 %t1, i16 %t2)
  %t4 = xor i1 %t3, true
  br i1 %t4, label %if.then2, label %if.next3
if.then2:
  ret void
if.next3:
  br label %if.end1
if.end1:
  %t5 = load i16, ptr %slot_x_2
  %t6 = load i16, ptr %slot_y_3
  %t7 = call i64 @aila_std_pixbuf__byte_index(ptr %p_pb, i16 %t5, i16 %t6)
  store i64 %t7, ptr %slot_bi_5
  %t8 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 0
  %t9 = load { ptr, i64 }, ptr %t8
  %t10 = extractvalue { ptr, i64 } %t9, 1
  store i64 %t10, ptr %slot_len_6
  %t12 = load i64, ptr %slot_bi_5
  %t13 = load i64, ptr %slot_len_6
  %t14 = icmp uge i64 %t12, %t13
  br i1 %t14, label %if.then5, label %if.next6
if.then5:
  ret void
if.next6:
  br label %if.end4
if.end4:
  %t15 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 4
  %t16 = load { i8, i8, i8, i8 }, ptr %t15
  %t17 = load i16, ptr %slot_x_2
  %t18 = call i8 @aila_std_pixfmt__bit_of({ i8, i8, i8, i8 } %t16, i16 %t17)
  store i8 %t18, ptr %slot_mask_7
  %t19 = load i1, ptr %slot_on_4
  br i1 %t19, label %if.then8, label %if.next9
if.then8:
  %t20 = load i64, ptr %slot_bi_5
  %t21 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 0
  %t22 = load { ptr, i64 }, ptr %t21
  %t23 = extractvalue { ptr, i64 } %t22, 0
  %t24 = extractvalue { ptr, i64 } %t22, 1
  %t25 = icmp uge i64 %t20, %t24
  br i1 %t25, label %chk.bad11, label %chk.ok10
chk.bad11:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 121)
  unreachable
chk.ok10:
  %t26 = getelementptr inbounds i8, ptr %t23, i64 %t20
  %t27 = load i64, ptr %slot_bi_5
  %t28 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 0
  %t29 = load { ptr, i64 }, ptr %t28
  %t30 = extractvalue { ptr, i64 } %t29, 0
  %t31 = extractvalue { ptr, i64 } %t29, 1
  %t32 = icmp uge i64 %t27, %t31
  br i1 %t32, label %chk.bad13, label %chk.ok12
chk.bad13:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 121)
  unreachable
chk.ok12:
  %t33 = getelementptr inbounds i8, ptr %t30, i64 %t27
  %t34 = load i8, ptr %t33
  %t35 = load i8, ptr %slot_mask_7
  %t36 = or i8 %t34, %t35
  store i8 %t36, ptr %t26
  br label %if.end7
if.next9:
  %t37 = load i64, ptr %slot_bi_5
  %t38 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 0
  %t39 = load { ptr, i64 }, ptr %t38
  %t40 = extractvalue { ptr, i64 } %t39, 0
  %t41 = extractvalue { ptr, i64 } %t39, 1
  %t42 = icmp uge i64 %t37, %t41
  br i1 %t42, label %chk.bad15, label %chk.ok14
chk.bad15:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 123)
  unreachable
chk.ok14:
  %t43 = getelementptr inbounds i8, ptr %t40, i64 %t37
  %t44 = load i64, ptr %slot_bi_5
  %t45 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 0
  %t46 = load { ptr, i64 }, ptr %t45
  %t47 = extractvalue { ptr, i64 } %t46, 0
  %t48 = extractvalue { ptr, i64 } %t46, 1
  %t49 = icmp uge i64 %t44, %t48
  br i1 %t49, label %chk.bad17, label %chk.ok16
chk.bad17:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 123)
  unreachable
chk.ok16:
  %t50 = getelementptr inbounds i8, ptr %t47, i64 %t44
  %t51 = load i8, ptr %t50
  %t52 = load i8, ptr %slot_mask_7
  %t53 = xor i8 %t52, -1
  %t54 = and i8 %t51, %t53
  store i8 %t54, ptr %t43
  br label %if.end7
if.end7:
  ret void
}

define i1 @aila_std_pixbuf__get_px(ptr %p_pb, i16 %p_x, i16 %p_y) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_x_2 = alloca i16, align 8
  %slot_y_3 = alloca i16, align 8
  %slot_bi_4 = alloca i64, align 8
  %slot_len_5 = alloca i64, align 8
  %slot_mask_6 = alloca i8, align 8
  store i16 %p_x, ptr %slot_x_2
  store i16 %p_y, ptr %slot_y_3
  %t1 = load i16, ptr %slot_x_2
  %t2 = load i16, ptr %slot_y_3
  %t3 = call i1 @aila_std_pixbuf__in_bounds(ptr %p_pb, i16 %t1, i16 %t2)
  %t4 = xor i1 %t3, true
  br i1 %t4, label %if.then2, label %if.next3
if.then2:
  ret i1 0
if.next3:
  br label %if.end1
if.end1:
  %t5 = load i16, ptr %slot_x_2
  %t6 = load i16, ptr %slot_y_3
  %t7 = call i64 @aila_std_pixbuf__byte_index(ptr %p_pb, i16 %t5, i16 %t6)
  store i64 %t7, ptr %slot_bi_4
  %t8 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 0
  %t9 = load { ptr, i64 }, ptr %t8
  %t10 = extractvalue { ptr, i64 } %t9, 1
  store i64 %t10, ptr %slot_len_5
  %t12 = load i64, ptr %slot_bi_4
  %t13 = load i64, ptr %slot_len_5
  %t14 = icmp uge i64 %t12, %t13
  br i1 %t14, label %if.then5, label %if.next6
if.then5:
  ret i1 0
if.next6:
  br label %if.end4
if.end4:
  %t15 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 4
  %t16 = load { i8, i8, i8, i8 }, ptr %t15
  %t17 = load i16, ptr %slot_x_2
  %t18 = call i8 @aila_std_pixfmt__bit_of({ i8, i8, i8, i8 } %t16, i16 %t17)
  store i8 %t18, ptr %slot_mask_6
  %t19 = load i64, ptr %slot_bi_4
  %t20 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 0
  %t21 = load { ptr, i64 }, ptr %t20
  %t22 = extractvalue { ptr, i64 } %t21, 0
  %t23 = extractvalue { ptr, i64 } %t21, 1
  %t24 = icmp uge i64 %t19, %t23
  br i1 %t24, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 137)
  unreachable
chk.ok7:
  %t25 = getelementptr inbounds i8, ptr %t22, i64 %t19
  %t26 = load i8, ptr %t25
  %t27 = load i8, ptr %slot_mask_6
  %t28 = and i8 %t26, %t27
  %t29 = icmp ne i8 %t28, 0
  ret i1 %t29
}

define i16 @aila_std_pixbuf__cols(ptr %p_pb) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %t1 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 1
  %t2 = load i16, ptr %t1
  %t3 = zext i8 8 to i16
  %t4 = icmp eq i16 %t3, 0
  br i1 %t4, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 16, ptr @.str.1, i64 141)
  unreachable
chk.ok1:
  %t5 = udiv i16 %t2, %t3
  ret i16 %t5
}

define i16 @aila_std_pixbuf__rows(ptr %p_pb) {
  %slot_pb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %t1 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_pb, i32 0, i32 2
  %t2 = load i16, ptr %t1
  %t3 = zext i8 16 to i16
  %t4 = icmp eq i16 %t3, 0
  br i1 %t4, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 16, ptr @.str.1, i64 145)
  unreachable
chk.ok1:
  %t5 = udiv i16 %t2, %t3
  ret i16 %t5
}
