; AILA module: std.display_host
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [17 x i8] c"division by zero\00", align 1
@.str.1 = private unnamed_addr constant [22 x i8] c"std/display_host.aila\00", align 1
@.str.2 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.3 = private unnamed_addr constant [4 x i8] c"P4\0A\00", align 1
@.str.4 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.5 = private unnamed_addr constant [2 x i8] c"\0A\00", align 1

declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare ptr @aila_rt_region_alloc(ptr, i64, i64)
declare i64 @aila_std_io__Writer__put(ptr, { ptr, i64 })
declare { ptr, { ptr, i64 } } @aila_std_io__stdout()
declare { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } @aila_std_display__Commit__buf(ptr)
declare i16 @aila_std_pixfmt__row_bytes({ i8, i8, i8, i8 }, i16)
declare ptr @aila_rt_region_enter(i64)
declare void @aila_rt_region_leave(ptr)
declare i1 @aila_std_pixbuf__get_px(ptr, i16, i16)

define i64 @aila_std_display_host__digits_u64(i64 %p_v) {
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
  %t3 = icmp uge i64 %t2, 10
  br i1 %t3, label %for.body4, label %for.end1
for.body4:
  %t4 = load i64, ptr %slot_x_2
  %t5 = icmp eq i64 10, 0
  br i1 %t5, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 29)
  unreachable
chk.ok5:
  %t6 = udiv i64 %t4, 10
  store i64 %t6, ptr %slot_x_2
  %t7 = load i64, ptr %slot_n_3
  %t8 = add i64 %t7, 1
  store i64 %t8, ptr %slot_n_3
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t9 = load i64, ptr %slot_n_3
  ret i64 %t9
}

define void @aila_std_display_host__fill_dec(i64 %p_v, { ptr, i64 } %p_buf, i64 %p_off, i64 %p_nd) {
  %slot_v_1 = alloca i64, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_off_3 = alloca i64, align 8
  %slot_nd_4 = alloca i64, align 8
  %slot_y_5 = alloca i64, align 8
  %slot_i_6 = alloca i64, align 8
  %slot_dg_7 = alloca i64, align 8
  store i64 %p_v, ptr %slot_v_1
  store { ptr, i64 } %p_buf, ptr %slot_buf_2
  store i64 %p_off, ptr %slot_off_3
  store i64 %p_nd, ptr %slot_nd_4
  %t1 = load i64, ptr %slot_v_1
  store i64 %t1, ptr %slot_y_5
  %t2 = load i64, ptr %slot_off_3
  %t3 = load i64, ptr %slot_nd_4
  %t4 = add i64 %t2, %t3
  store i64 %t4, ptr %slot_i_6
  br label %for.cond2
for.cond2:
  %t5 = load i64, ptr %slot_i_6
  %t6 = load i64, ptr %slot_off_3
  %t7 = icmp ugt i64 %t5, %t6
  br i1 %t7, label %for.body4, label %for.end1
for.body4:
  %t8 = load i64, ptr %slot_i_6
  %t9 = sub i64 %t8, 1
  store i64 %t9, ptr %slot_i_6
  %t10 = load i64, ptr %slot_y_5
  %t11 = icmp eq i64 10, 0
  br i1 %t11, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 41)
  unreachable
chk.ok5:
  %t12 = urem i64 %t10, 10
  store i64 %t12, ptr %slot_dg_7
  %t13 = load i64, ptr %slot_i_6
  %t14 = load { ptr, i64 }, ptr %slot_buf_2
  %t15 = extractvalue { ptr, i64 } %t14, 0
  %t16 = extractvalue { ptr, i64 } %t14, 1
  %t17 = icmp uge i64 %t13, %t16
  br i1 %t17, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 42)
  unreachable
chk.ok7:
  %t18 = getelementptr inbounds i8, ptr %t15, i64 %t13
  %t19 = load i64, ptr %slot_dg_7
  %t20 = trunc i64 %t19 to i32
  %t21 = add i32 %t20, 48
  %t22 = trunc i32 %t21 to i8
  store i8 %t22, ptr %t18
  %t23 = load i64, ptr %slot_y_5
  %t24 = icmp eq i64 10, 0
  br i1 %t24, label %chk.bad10, label %chk.ok9
chk.bad10:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 43)
  unreachable
chk.ok9:
  %t25 = udiv i64 %t23, 10
  store i64 %t25, ptr %slot_y_5
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  ret void
}

define void @aila_std_display_host__put_dec(ptr %r0, ptr %p_w, i64 %p_v, i64 %p_nd) {
  %slot_w_1 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_v_2 = alloca i64, align 8
  %slot_nd_3 = alloca i64, align 8
  %slot_buf_4 = alloca { ptr, i64 }, align 8
  store i64 %p_v, ptr %slot_v_2
  store i64 %p_nd, ptr %slot_nd_3
  %t1 = load i64, ptr %slot_nd_3
  %t2 = mul i64 %t1, 1
  %t3 = call ptr @aila_rt_region_alloc(ptr %r0, i64 %t2, i64 1)
  %t4 = insertvalue { ptr, i64 } undef, ptr %t3, 0
  %t5 = insertvalue { ptr, i64 } %t4, i64 %t1, 1
  store { ptr, i64 } %t5, ptr %slot_buf_4
  %t6 = load i64, ptr %slot_v_2
  %t7 = load { ptr, i64 }, ptr %slot_buf_4
  %t8 = load i64, ptr %slot_nd_3
  call void @aila_std_display_host__fill_dec(i64 %t6, { ptr, i64 } %t7, i64 0, i64 %t8)
  %t9 = load { ptr, i64 }, ptr %slot_buf_4
  %t10 = call i64 @aila_std_io__Writer__put(ptr %p_w, { ptr, i64 } %t9)
  ret void
}

define void @aila_std_display_host__sink_host_pbm(ptr %p_c) {
  %slot_c_1 = alloca { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, align 8
  %slot_w_2 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_fb_3 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_rb_4 = alloca i64, align 8
  %slot_dw_5 = alloca i64, align 8
  %slot_dh_6 = alloca i64, align 8
  %slot___sc_7 = alloca i1, align 8
  %t1 = call { ptr, { ptr, i64 } } @aila_std_io__stdout()
  store { ptr, { ptr, i64 } } %t1, ptr %slot_w_2
  %t2 = call { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } @aila_std_display__Commit__buf(ptr %p_c)
  store { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } %t2, ptr %slot_fb_3
  %t3 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot_fb_3, i32 0, i32 1
  %t4 = load i16, ptr %t3
  %t5 = icmp eq i16 %t4, 0
  br i1 %t5, label %sc.l4, label %sc.r5
sc.l4:
  store i1 1, ptr %slot___sc_7
  br label %sc.end6
sc.r5:
  %t6 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot_fb_3, i32 0, i32 2
  %t7 = load i16, ptr %t6
  %t8 = icmp eq i16 %t7, 0
  store i1 %t8, ptr %slot___sc_7
  br label %sc.end6
sc.end6:
  %t9 = load i1, ptr %slot___sc_7
  br i1 %t9, label %if.then2, label %if.next3
if.then2:
  ret void
if.next3:
  br label %if.end1
if.end1:
  %t10 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot_fb_3, i32 0, i32 4
  %t11 = load { i8, i8, i8, i8 }, ptr %t10
  %t12 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot_fb_3, i32 0, i32 1
  %t13 = load i16, ptr %t12
  %t14 = call i16 @aila_std_pixfmt__row_bytes({ i8, i8, i8, i8 } %t11, i16 %t13)
  %t15 = zext i16 %t14 to i64
  store i64 %t15, ptr %slot_rb_4
  %t16 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot_fb_3, i32 0, i32 1
  %t17 = load i16, ptr %t16
  %t18 = zext i16 %t17 to i64
  %t19 = call i64 @aila_std_display_host__digits_u64(i64 %t18)
  store i64 %t19, ptr %slot_dw_5
  %t20 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot_fb_3, i32 0, i32 2
  %t21 = load i16, ptr %t20
  %t22 = zext i16 %t21 to i64
  %t23 = call i64 @aila_std_display_host__digits_u64(i64 %t22)
  store i64 %t23, ptr %slot_dh_6
  %t24 = call ptr @aila_rt_region_enter(i64 0)
  %t25 = insertvalue { ptr, i64 } undef, ptr @.str.3, 0
  %t26 = insertvalue { ptr, i64 } %t25, i64 3, 1
  call void @aila_std_display_host__put_bytes(ptr %slot_w_2, { ptr, i64 } %t26)
  %t27 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot_fb_3, i32 0, i32 1
  %t28 = load i16, ptr %t27
  %t29 = zext i16 %t28 to i64
  %t30 = load i64, ptr %slot_dw_5
  call void @aila_std_display_host__put_dec(ptr %t24, ptr %slot_w_2, i64 %t29, i64 %t30)
  %t31 = insertvalue { ptr, i64 } undef, ptr @.str.4, 0
  %t32 = insertvalue { ptr, i64 } %t31, i64 1, 1
  call void @aila_std_display_host__put_bytes(ptr %slot_w_2, { ptr, i64 } %t32)
  %t33 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot_fb_3, i32 0, i32 2
  %t34 = load i16, ptr %t33
  %t35 = zext i16 %t34 to i64
  %t36 = load i64, ptr %slot_dh_6
  call void @aila_std_display_host__put_dec(ptr %t24, ptr %slot_w_2, i64 %t35, i64 %t36)
  %t37 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t38 = insertvalue { ptr, i64 } %t37, i64 1, 1
  call void @aila_std_display_host__put_bytes(ptr %slot_w_2, { ptr, i64 } %t38)
  %t39 = load i64, ptr %slot_rb_4
  call void @aila_std_display_host__sink_rows(ptr %t24, ptr %slot_w_2, ptr %slot_fb_3, i64 %t39)
  call void @aila_rt_region_leave(ptr %t24)
  ret void
}

define void @aila_std_display_host__put_bytes(ptr %p_w, { ptr, i64 } %p_t) {
  %slot_w_1 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_t_2 = alloca { ptr, i64 }, align 8
  %slot_b_3 = alloca { ptr, i64 }, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_2
  %t1 = load { ptr, i64 }, ptr %slot_t_2
  store { ptr, i64 } %t1, ptr %slot_b_3
  %t2 = load { ptr, i64 }, ptr %slot_b_3
  %t3 = call i64 @aila_std_io__Writer__put(ptr %p_w, { ptr, i64 } %t2)
  ret void
}

define void @aila_std_display_host__sink_rows(ptr %r0, ptr %p_w, ptr %p_fb, i64 %p_rb) {
  %slot_w_1 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_fb_2 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_rb_3 = alloca i64, align 8
  %slot_rs_4 = alloca i64, align 8
  %slot_row_5 = alloca { ptr, i64 }, align 8
  %slot_y_6 = alloca i64, align 8
  %slot_i_7 = alloca i64, align 8
  store i64 %p_rb, ptr %slot_rb_3
  %t1 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_fb, i32 0, i32 3
  %t2 = load i16, ptr %t1
  %t3 = zext i16 %t2 to i64
  store i64 %t3, ptr %slot_rs_4
  %t4 = load i64, ptr %slot_rb_3
  %t5 = mul i64 %t4, 1
  %t6 = call ptr @aila_rt_region_alloc(ptr %r0, i64 %t5, i64 1)
  %t7 = insertvalue { ptr, i64 } undef, ptr %t6, 0
  %t8 = insertvalue { ptr, i64 } %t7, i64 %t4, 1
  store { ptr, i64 } %t8, ptr %slot_row_5
  store i64 0, ptr %slot_y_6
  br label %for.cond2
for.cond2:
  %t9 = load i64, ptr %slot_y_6
  %t10 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_fb, i32 0, i32 2
  %t11 = load i16, ptr %t10
  %t12 = zext i16 %t11 to i64
  %t13 = icmp ult i64 %t9, %t12
  br i1 %t13, label %for.body4, label %for.end1
for.body4:
  store i64 0, ptr %slot_i_7
  br label %for.cond6
for.cond6:
  %t14 = load i64, ptr %slot_i_7
  %t15 = load i64, ptr %slot_rb_3
  %t16 = icmp ult i64 %t14, %t15
  br i1 %t16, label %for.body8, label %for.end5
for.body8:
  %t17 = load i64, ptr %slot_i_7
  %t18 = load { ptr, i64 }, ptr %slot_row_5
  %t19 = extractvalue { ptr, i64 } %t18, 0
  %t20 = extractvalue { ptr, i64 } %t18, 1
  %t21 = icmp uge i64 %t17, %t20
  br i1 %t21, label %chk.bad10, label %chk.ok9
chk.bad10:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 87)
  unreachable
chk.ok9:
  %t22 = getelementptr inbounds i8, ptr %t19, i64 %t17
  %t23 = load i64, ptr %slot_y_6
  %t24 = load i64, ptr %slot_rs_4
  %t25 = mul i64 %t23, %t24
  %t26 = load i64, ptr %slot_i_7
  %t27 = add i64 %t25, %t26
  %t28 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_fb, i32 0, i32 0
  %t29 = load { ptr, i64 }, ptr %t28
  %t30 = extractvalue { ptr, i64 } %t29, 0
  %t31 = extractvalue { ptr, i64 } %t29, 1
  %t32 = icmp uge i64 %t27, %t31
  br i1 %t32, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 87)
  unreachable
chk.ok11:
  %t33 = getelementptr inbounds i8, ptr %t30, i64 %t27
  %t34 = load i8, ptr %t33
  store i8 %t34, ptr %t22
  %t35 = load i64, ptr %slot_i_7
  %t36 = add i64 %t35, 1
  store i64 %t36, ptr %slot_i_7
  br label %for.step7
for.step7:
  br label %for.cond6
for.end5:
  %t37 = load { ptr, i64 }, ptr %slot_row_5
  %t38 = call i64 @aila_std_io__Writer__put(ptr %p_w, { ptr, i64 } %t37)
  %t39 = load i64, ptr %slot_y_6
  %t40 = add i64 %t39, 1
  store i64 %t40, ptr %slot_y_6
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  ret void
}

define void @aila_std_display_host__sink_host_ascii(ptr %p_c) {
  %slot_c_1 = alloca { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, align 8
  %slot_w_2 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_fb_3 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot___sc_4 = alloca i1, align 8
  %t1 = call { ptr, { ptr, i64 } } @aila_std_io__stdout()
  store { ptr, { ptr, i64 } } %t1, ptr %slot_w_2
  %t2 = call { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } @aila_std_display__Commit__buf(ptr %p_c)
  store { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } %t2, ptr %slot_fb_3
  %t3 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot_fb_3, i32 0, i32 1
  %t4 = load i16, ptr %t3
  %t5 = icmp eq i16 %t4, 0
  br i1 %t5, label %sc.l4, label %sc.r5
sc.l4:
  store i1 1, ptr %slot___sc_4
  br label %sc.end6
sc.r5:
  %t6 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot_fb_3, i32 0, i32 2
  %t7 = load i16, ptr %t6
  %t8 = icmp eq i16 %t7, 0
  store i1 %t8, ptr %slot___sc_4
  br label %sc.end6
sc.end6:
  %t9 = load i1, ptr %slot___sc_4
  br i1 %t9, label %if.then2, label %if.next3
if.then2:
  ret void
if.next3:
  br label %if.end1
if.end1:
  %t10 = call ptr @aila_rt_region_enter(i64 0)
  call void @aila_std_display_host__ascii_rows(ptr %t10, ptr %slot_w_2, ptr %slot_fb_3)
  call void @aila_rt_region_leave(ptr %t10)
  ret void
}

define void @aila_std_display_host__ascii_rows(ptr %r0, ptr %p_w, ptr %p_fb) {
  %slot_w_1 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_fb_2 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_lw_3 = alloca i64, align 8
  %slot_line_4 = alloca { ptr, i64 }, align 8
  %slot_y_5 = alloca i64, align 8
  %slot_x_6 = alloca i64, align 8
  %slot_ch_7 = alloca i8, align 8
  %t1 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_fb, i32 0, i32 1
  %t2 = load i16, ptr %t1
  %t3 = zext i16 %t2 to i64
  %t4 = add i64 %t3, 1
  store i64 %t4, ptr %slot_lw_3
  %t5 = load i64, ptr %slot_lw_3
  %t6 = mul i64 %t5, 1
  %t7 = call ptr @aila_rt_region_alloc(ptr %r0, i64 %t6, i64 1)
  %t8 = insertvalue { ptr, i64 } undef, ptr %t7, 0
  %t9 = insertvalue { ptr, i64 } %t8, i64 %t5, 1
  store { ptr, i64 } %t9, ptr %slot_line_4
  store i64 0, ptr %slot_y_5
  br label %for.cond2
for.cond2:
  %t10 = load i64, ptr %slot_y_5
  %t11 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_fb, i32 0, i32 2
  %t12 = load i16, ptr %t11
  %t13 = zext i16 %t12 to i64
  %t14 = icmp ult i64 %t10, %t13
  br i1 %t14, label %for.body4, label %for.end1
for.body4:
  store i64 0, ptr %slot_x_6
  br label %for.cond6
for.cond6:
  %t15 = load i64, ptr %slot_x_6
  %t16 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_fb, i32 0, i32 1
  %t17 = load i16, ptr %t16
  %t18 = zext i16 %t17 to i64
  %t19 = icmp ult i64 %t15, %t18
  br i1 %t19, label %for.body8, label %for.end5
for.body8:
  store i8 46, ptr %slot_ch_7
  %t20 = load i64, ptr %slot_x_6
  %t21 = trunc i64 %t20 to i16
  %t22 = load i64, ptr %slot_y_5
  %t23 = trunc i64 %t22 to i16
  %t24 = call i1 @aila_std_pixbuf__get_px(ptr %p_fb, i16 %t21, i16 %t23)
  br i1 %t24, label %if.then10, label %if.next11
if.then10:
  store i8 35, ptr %slot_ch_7
  br label %if.end9
if.next11:
  br label %if.end9
if.end9:
  %t25 = load i64, ptr %slot_x_6
  %t26 = load { ptr, i64 }, ptr %slot_line_4
  %t27 = extractvalue { ptr, i64 } %t26, 0
  %t28 = extractvalue { ptr, i64 } %t26, 1
  %t29 = icmp uge i64 %t25, %t28
  br i1 %t29, label %chk.bad13, label %chk.ok12
chk.bad13:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 117)
  unreachable
chk.ok12:
  %t30 = getelementptr inbounds i8, ptr %t27, i64 %t25
  %t31 = load i8, ptr %slot_ch_7
  store i8 %t31, ptr %t30
  %t32 = load i64, ptr %slot_x_6
  %t33 = add i64 %t32, 1
  store i64 %t33, ptr %slot_x_6
  br label %for.step7
for.step7:
  br label %for.cond6
for.end5:
  %t34 = load i64, ptr %slot_lw_3
  %t35 = sub i64 %t34, 1
  %t36 = load { ptr, i64 }, ptr %slot_line_4
  %t37 = extractvalue { ptr, i64 } %t36, 0
  %t38 = extractvalue { ptr, i64 } %t36, 1
  %t39 = icmp uge i64 %t35, %t38
  br i1 %t39, label %chk.bad15, label %chk.ok14
chk.bad15:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 120)
  unreachable
chk.ok14:
  %t40 = getelementptr inbounds i8, ptr %t37, i64 %t35
  store i8 10, ptr %t40
  %t41 = load { ptr, i64 }, ptr %slot_line_4
  %t42 = call i64 @aila_std_io__Writer__put(ptr %p_w, { ptr, i64 } %t41)
  %t43 = load i64, ptr %slot_y_5
  %t44 = add i64 %t43, 1
  store i64 %t44, ptr %slot_y_5
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  ret void
}
