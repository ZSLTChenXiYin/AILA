; AILA module: std.textgrid
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [25 x i8] c"shift count out of range\00", align 1
@.str.1 = private unnamed_addr constant [18 x i8] c"std/textgrid.aila\00", align 1
@.str.2 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1


declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare { i32, i64 } @aila_std_paint__decode({ ptr, i64 }, i64, i64)
declare i16 @aila_std_paint__glyph_width(ptr, i32)
declare void @aila_std_damage__add_cell(ptr, i16, i16)
declare { i16, i16, i16, i16 } @aila_std_paint__cell_rect(i16, i16)
declare { i16, i16, i16, i16 } @aila_std_damage__rect(i16, i16, i16, i16)
declare void @aila_std_paint__fill_rect(ptr, { i16, i16, i16, i16 }, i1)
declare i16 @aila_std_paint__blit_glyph(ptr, ptr, i32, i16, i16)

define i32 @aila_std_textgrid__encode(i32 %p_cp, i32 %p_w) {
  %slot_cp_1 = alloca i32, align 8
  %slot_w_2 = alloca i32, align 8
  store i32 %p_cp, ptr %slot_cp_1
  store i32 %p_w, ptr %slot_w_2
  %t1 = load i32, ptr %slot_w_2
  %t2 = icmp uge i32 30, 32
  br i1 %t2, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 24, ptr @.str.1, i64 82)
  unreachable
chk.ok1:
  %t3 = shl i32 %t1, 30
  %t4 = load i32, ptr %slot_cp_1
  %t5 = and i32 %t4, 1073741823
  %t6 = or i32 %t3, %t5
  ret i32 %t6
}

define i32 @aila_std_textgrid__cell_cp(i32 %p_v) {
  %slot_v_1 = alloca i32, align 8
  store i32 %p_v, ptr %slot_v_1
  %t1 = load i32, ptr %slot_v_1
  %t2 = and i32 %t1, 1073741823
  ret i32 %t2
}

define i32 @aila_std_textgrid__cell_w(i32 %p_v) {
  %slot_v_1 = alloca i32, align 8
  store i32 %p_v, ptr %slot_v_1
  %t1 = load i32, ptr %slot_v_1
  %t2 = icmp uge i32 30, 32
  br i1 %t2, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 24, ptr @.str.1, i64 90)
  unreachable
chk.ok1:
  %t3 = lshr i32 %t1, 30
  ret i32 %t3
}

define i64 @aila_std_textgrid__words(i16 %p_cols, i16 %p_rows) {
  %slot_cols_1 = alloca i16, align 8
  %slot_rows_2 = alloca i16, align 8
  %slot_n_3 = alloca i64, align 8
  store i16 %p_cols, ptr %slot_cols_1
  store i16 %p_rows, ptr %slot_rows_2
  %t1 = load i16, ptr %slot_cols_1
  %t2 = zext i16 %t1 to i64
  %t3 = load i16, ptr %slot_rows_2
  %t4 = zext i16 %t3 to i64
  %t5 = mul i64 %t2, %t4
  store i64 %t5, ptr %slot_n_3
  %t6 = load i64, ptr %slot_n_3
  %t7 = mul i64 %t6, 2
  %t8 = add i64 4, %t7
  ret i64 %t8
}

define i64 @aila_std_textgrid__cells_at() {
  ret i64 4
}

define i64 @aila_std_textgrid__prev_at({ { ptr, i64 } } %p_g) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  store { { ptr, i64 } } %p_g, ptr %slot_g_1
  %t1 = load { { ptr, i64 } }, ptr %slot_g_1
  %t2 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t1)
  %t3 = zext i16 %t2 to i64
  %t4 = load { { ptr, i64 } }, ptr %slot_g_1
  %t5 = call i16 @aila_std_textgrid__rows_of({ { ptr, i64 } } %t4)
  %t6 = zext i16 %t5 to i64
  %t7 = mul i64 %t3, %t6
  %t8 = add i64 4, %t7
  ret i64 %t8
}

define i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %p_g) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  store { { ptr, i64 } } %p_g, ptr %slot_g_1
  %t1 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_g_1, i32 0, i32 0
  %t2 = load { ptr, i64 }, ptr %t1
  %t3 = extractvalue { ptr, i64 } %t2, 0
  %t4 = extractvalue { ptr, i64 } %t2, 1
  %t5 = icmp uge i64 0, %t4
  br i1 %t5, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 107)
  unreachable
chk.ok1:
  %t6 = getelementptr inbounds i32, ptr %t3, i64 0
  %t7 = load i32, ptr %t6
  %t8 = trunc i32 %t7 to i16
  ret i16 %t8
}

define i16 @aila_std_textgrid__rows_of({ { ptr, i64 } } %p_g) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  store { { ptr, i64 } } %p_g, ptr %slot_g_1
  %t1 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_g_1, i32 0, i32 0
  %t2 = load { ptr, i64 }, ptr %t1
  %t3 = extractvalue { ptr, i64 } %t2, 0
  %t4 = extractvalue { ptr, i64 } %t2, 1
  %t5 = icmp uge i64 1, %t4
  br i1 %t5, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 111)
  unreachable
chk.ok1:
  %t6 = getelementptr inbounds i32, ptr %t3, i64 1
  %t7 = load i32, ptr %t6
  %t8 = trunc i32 %t7 to i16
  ret i16 %t8
}

define { i16, i16 } @aila_std_textgrid__cursor({ { ptr, i64 } } %p_g) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  store { { ptr, i64 } } %p_g, ptr %slot_g_1
  %t1 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_g_1, i32 0, i32 0
  %t2 = load { ptr, i64 }, ptr %t1
  %t3 = extractvalue { ptr, i64 } %t2, 0
  %t4 = extractvalue { ptr, i64 } %t2, 1
  %t5 = icmp uge i64 2, %t4
  br i1 %t5, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 115)
  unreachable
chk.ok1:
  %t6 = getelementptr inbounds i32, ptr %t3, i64 2
  %t7 = load i32, ptr %t6
  %t8 = trunc i32 %t7 to i16
  %t9 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_g_1, i32 0, i32 0
  %t10 = load { ptr, i64 }, ptr %t9
  %t11 = extractvalue { ptr, i64 } %t10, 0
  %t12 = extractvalue { ptr, i64 } %t10, 1
  %t13 = icmp uge i64 3, %t12
  br i1 %t13, label %chk.bad4, label %chk.ok3
chk.bad4:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 115)
  unreachable
chk.ok3:
  %t14 = getelementptr inbounds i32, ptr %t11, i64 3
  %t15 = load i32, ptr %t14
  %t16 = trunc i32 %t15 to i16
  %t17 = insertvalue { i16, i16 } undef, i16 %t8, 0
  %t18 = insertvalue { i16, i16 } %t17, i16 %t16, 1
  ret { i16, i16 } %t18
}

define void @aila_std_textgrid__set_cursor(ptr %p_g, i16 %p_col, i16 %p_row) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_col_2 = alloca i16, align 8
  %slot_row_3 = alloca i16, align 8
  store i16 %p_col, ptr %slot_col_2
  store i16 %p_row, ptr %slot_row_3
  %t1 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t2 = load { ptr, i64 }, ptr %t1
  %t3 = extractvalue { ptr, i64 } %t2, 0
  %t4 = extractvalue { ptr, i64 } %t2, 1
  %t5 = icmp uge i64 2, %t4
  br i1 %t5, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 119)
  unreachable
chk.ok1:
  %t6 = getelementptr inbounds i32, ptr %t3, i64 2
  %t7 = load i16, ptr %slot_col_2
  %t8 = zext i16 %t7 to i32
  store i32 %t8, ptr %t6
  %t9 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t10 = load { ptr, i64 }, ptr %t9
  %t11 = extractvalue { ptr, i64 } %t10, 0
  %t12 = extractvalue { ptr, i64 } %t10, 1
  %t13 = icmp uge i64 3, %t12
  br i1 %t13, label %chk.bad4, label %chk.ok3
chk.bad4:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 120)
  unreachable
chk.ok3:
  %t14 = getelementptr inbounds i32, ptr %t11, i64 3
  %t15 = load i16, ptr %slot_row_3
  %t16 = zext i16 %t15 to i32
  store i32 %t16, ptr %t14
  ret void
}

define { { ptr, i64 } } @aila_std_textgrid__new({ ptr, i64 } %p_raw, i16 %p_cols, i16 %p_rows) {
  %slot_raw_1 = alloca { ptr, i64 }, align 8
  %slot_cols_2 = alloca i16, align 8
  %slot_rows_3 = alloca i16, align 8
  %slot_g_4 = alloca { { ptr, i64 } }, align 8
  %slot___lit_5 = alloca { { ptr, i64 } }, align 8
  store { ptr, i64 } %p_raw, ptr %slot_raw_1
  store i16 %p_cols, ptr %slot_cols_2
  store i16 %p_rows, ptr %slot_rows_3
  %t1 = load { ptr, i64 }, ptr %slot_raw_1
  %t2 = getelementptr inbounds { { ptr, i64 } }, ptr %slot___lit_5, i32 0, i32 0
  store { ptr, i64 } %t1, ptr %t2
  %t3 = load { { ptr, i64 } }, ptr %slot___lit_5
  store { { ptr, i64 } } %t3, ptr %slot_g_4
  %t4 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_g_4, i32 0, i32 0
  %t5 = load { ptr, i64 }, ptr %t4
  %t6 = extractvalue { ptr, i64 } %t5, 0
  %t7 = extractvalue { ptr, i64 } %t5, 1
  %t8 = icmp uge i64 0, %t7
  br i1 %t8, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 125)
  unreachable
chk.ok1:
  %t9 = getelementptr inbounds i32, ptr %t6, i64 0
  %t10 = load i16, ptr %slot_cols_2
  %t11 = zext i16 %t10 to i32
  store i32 %t11, ptr %t9
  %t12 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_g_4, i32 0, i32 0
  %t13 = load { ptr, i64 }, ptr %t12
  %t14 = extractvalue { ptr, i64 } %t13, 0
  %t15 = extractvalue { ptr, i64 } %t13, 1
  %t16 = icmp uge i64 1, %t15
  br i1 %t16, label %chk.bad4, label %chk.ok3
chk.bad4:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 126)
  unreachable
chk.ok3:
  %t17 = getelementptr inbounds i32, ptr %t14, i64 1
  %t18 = load i16, ptr %slot_rows_3
  %t19 = zext i16 %t18 to i32
  store i32 %t19, ptr %t17
  %t20 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_g_4, i32 0, i32 0
  %t21 = load { ptr, i64 }, ptr %t20
  %t22 = extractvalue { ptr, i64 } %t21, 0
  %t23 = extractvalue { ptr, i64 } %t21, 1
  %t24 = icmp uge i64 2, %t23
  br i1 %t24, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 127)
  unreachable
chk.ok5:
  %t25 = getelementptr inbounds i32, ptr %t22, i64 2
  store i32 0, ptr %t25
  %t26 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_g_4, i32 0, i32 0
  %t27 = load { ptr, i64 }, ptr %t26
  %t28 = extractvalue { ptr, i64 } %t27, 0
  %t29 = extractvalue { ptr, i64 } %t27, 1
  %t30 = icmp uge i64 3, %t29
  br i1 %t30, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 128)
  unreachable
chk.ok7:
  %t31 = getelementptr inbounds i32, ptr %t28, i64 3
  store i32 0, ptr %t31
  call void @aila_std_textgrid__clear(ptr %slot_g_4)
  %t32 = load { { ptr, i64 } }, ptr %slot_g_4
  ret { { ptr, i64 } } %t32
}

define void @aila_std_textgrid__clear(ptr %p_g) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_total_2 = alloca i64, align 8
  %slot_i_3 = alloca i64, align 8
  %t1 = load { { ptr, i64 } }, ptr %p_g
  %t2 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t1)
  %t3 = load { { ptr, i64 } }, ptr %p_g
  %t4 = call i16 @aila_std_textgrid__rows_of({ { ptr, i64 } } %t3)
  %t5 = call i64 @aila_std_textgrid__words(i16 %t2, i16 %t4)
  store i64 %t5, ptr %slot_total_2
  store i64 4, ptr %slot_i_3
  br label %for.cond2
for.cond2:
  %t6 = load i64, ptr %slot_i_3
  %t7 = load i64, ptr %slot_total_2
  %t8 = icmp ult i64 %t6, %t7
  br i1 %t8, label %for.body4, label %for.end1
for.body4:
  %t9 = load i64, ptr %slot_i_3
  %t10 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t11 = load { ptr, i64 }, ptr %t10
  %t12 = extractvalue { ptr, i64 } %t11, 1
  %t14 = icmp ult i64 %t9, %t12
  br i1 %t14, label %if.then6, label %if.next7
if.then6:
  %t15 = load i64, ptr %slot_i_3
  %t16 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t17 = load { ptr, i64 }, ptr %t16
  %t18 = extractvalue { ptr, i64 } %t17, 0
  %t19 = extractvalue { ptr, i64 } %t17, 1
  %t20 = icmp uge i64 %t15, %t19
  br i1 %t20, label %chk.bad9, label %chk.ok8
chk.bad9:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 138)
  unreachable
chk.ok8:
  %t21 = getelementptr inbounds i32, ptr %t18, i64 %t15
  store i32 0, ptr %t21
  br label %if.end5
if.next7:
  br label %if.end5
if.end5:
  %t22 = load i64, ptr %slot_i_3
  %t23 = add i64 %t22, 1
  store i64 %t23, ptr %slot_i_3
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  call void @aila_std_textgrid__set_cursor(ptr %p_g, i16 0, i16 0)
  ret void
}

define i64 @aila_std_textgrid__idx({ { ptr, i64 } } %p_g, i16 %p_col, i16 %p_row) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_col_2 = alloca i16, align 8
  %slot_row_3 = alloca i16, align 8
  store { { ptr, i64 } } %p_g, ptr %slot_g_1
  store i16 %p_col, ptr %slot_col_2
  store i16 %p_row, ptr %slot_row_3
  %t1 = load i16, ptr %slot_row_3
  %t2 = zext i16 %t1 to i64
  %t3 = load { { ptr, i64 } }, ptr %slot_g_1
  %t4 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t3)
  %t5 = zext i16 %t4 to i64
  %t6 = mul i64 %t2, %t5
  %t7 = load i16, ptr %slot_col_2
  %t8 = zext i16 %t7 to i64
  %t9 = add i64 %t6, %t8
  ret i64 %t9
}

define i32 @aila_std_textgrid__get({ { ptr, i64 } } %p_g, i16 %p_col, i16 %p_row) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_col_2 = alloca i16, align 8
  %slot_row_3 = alloca i16, align 8
  store { { ptr, i64 } } %p_g, ptr %slot_g_1
  store i16 %p_col, ptr %slot_col_2
  store i16 %p_row, ptr %slot_row_3
  %t1 = load i16, ptr %slot_col_2
  %t2 = load { { ptr, i64 } }, ptr %slot_g_1
  %t3 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t2)
  %t4 = icmp uge i16 %t1, %t3
  br i1 %t4, label %if.then2, label %if.next3
if.then2:
  ret i32 0
if.next3:
  br label %if.end1
if.end1:
  %t5 = load i16, ptr %slot_row_3
  %t6 = load { { ptr, i64 } }, ptr %slot_g_1
  %t7 = call i16 @aila_std_textgrid__rows_of({ { ptr, i64 } } %t6)
  %t8 = icmp uge i16 %t5, %t7
  br i1 %t8, label %if.then5, label %if.next6
if.then5:
  ret i32 0
if.next6:
  br label %if.end4
if.end4:
  %t9 = call i64 @aila_std_textgrid__cells_at()
  %t10 = load { { ptr, i64 } }, ptr %slot_g_1
  %t11 = load i16, ptr %slot_col_2
  %t12 = load i16, ptr %slot_row_3
  %t13 = call i64 @aila_std_textgrid__idx({ { ptr, i64 } } %t10, i16 %t11, i16 %t12)
  %t14 = add i64 %t9, %t13
  %t15 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_g_1, i32 0, i32 0
  %t16 = load { ptr, i64 }, ptr %t15
  %t17 = extractvalue { ptr, i64 } %t16, 0
  %t18 = extractvalue { ptr, i64 } %t16, 1
  %t19 = icmp uge i64 %t14, %t18
  br i1 %t19, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 157)
  unreachable
chk.ok7:
  %t20 = getelementptr inbounds i32, ptr %t17, i64 %t14
  %t21 = load i32, ptr %t20
  ret i32 %t21
}

define void @aila_std_textgrid__set(ptr %p_g, i16 %p_col, i16 %p_row, i32 %p_v) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_col_2 = alloca i16, align 8
  %slot_row_3 = alloca i16, align 8
  %slot_v_4 = alloca i32, align 8
  store i16 %p_col, ptr %slot_col_2
  store i16 %p_row, ptr %slot_row_3
  store i32 %p_v, ptr %slot_v_4
  %t1 = load i16, ptr %slot_col_2
  %t2 = load { { ptr, i64 } }, ptr %p_g
  %t3 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t2)
  %t4 = icmp uge i16 %t1, %t3
  br i1 %t4, label %if.then2, label %if.next3
if.then2:
  ret void
if.next3:
  br label %if.end1
if.end1:
  %t5 = load i16, ptr %slot_row_3
  %t6 = load { { ptr, i64 } }, ptr %p_g
  %t7 = call i16 @aila_std_textgrid__rows_of({ { ptr, i64 } } %t6)
  %t8 = icmp uge i16 %t5, %t7
  br i1 %t8, label %if.then5, label %if.next6
if.then5:
  ret void
if.next6:
  br label %if.end4
if.end4:
  %t9 = call i64 @aila_std_textgrid__cells_at()
  %t10 = load { { ptr, i64 } }, ptr %p_g
  %t11 = load i16, ptr %slot_col_2
  %t12 = load i16, ptr %slot_row_3
  %t13 = call i64 @aila_std_textgrid__idx({ { ptr, i64 } } %t10, i16 %t11, i16 %t12)
  %t14 = add i64 %t9, %t13
  %t15 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t16 = load { ptr, i64 }, ptr %t15
  %t17 = extractvalue { ptr, i64 } %t16, 0
  %t18 = extractvalue { ptr, i64 } %t16, 1
  %t19 = icmp uge i64 %t14, %t18
  br i1 %t19, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 167)
  unreachable
chk.ok7:
  %t20 = getelementptr inbounds i32, ptr %t17, i64 %t14
  %t21 = load i32, ptr %slot_v_4
  store i32 %t21, ptr %t20
  ret void
}

define void @aila_std_textgrid__scroll_by(ptr %p_g, i16 %p_n) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_n_2 = alloca i16, align 8
  %slot_c_3 = alloca i16, align 8
  %slot_r_4 = alloca i16, align 8
  %slot_nr_5 = alloca i16, align 8
  %slot___destr_6 = alloca { i16, i16 }, align 8
  store i16 %p_n, ptr %slot_n_2
  %t1 = load i16, ptr %slot_n_2
  %t2 = icmp eq i16 %t1, 0
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  ret void
if.next3:
  br label %if.end1
if.end1:
  %t3 = load { { ptr, i64 } }, ptr %p_g
  %t4 = call { i16, i16 } @aila_std_textgrid__cursor({ { ptr, i64 } } %t3)
  store { i16, i16 } %t4, ptr %slot___destr_6
  %t5 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_6, i32 0, i32 0
  %t6 = load i16, ptr %t5
  store i16 %t6, ptr %slot_c_3
  %t7 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_6, i32 0, i32 1
  %t8 = load i16, ptr %t7
  store i16 %t8, ptr %slot_r_4
  %t9 = load i16, ptr %slot_n_2
  %t10 = load i16, ptr %slot_r_4
  %t11 = call i16 @aila_std_textgrid__scroll_cells(ptr %p_g, i16 %t9, i16 %t10)
  store i16 %t11, ptr %slot_nr_5
  %t12 = load i16, ptr %slot_c_3
  %t13 = load i16, ptr %slot_nr_5
  call void @aila_std_textgrid__set_cursor(ptr %p_g, i16 0, i16 %t13)
  ret void
}

define void @aila_std_textgrid__newline(ptr %p_g) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_c_2 = alloca i16, align 8
  %slot_r_3 = alloca i16, align 8
  %slot_c2_4 = alloca i16, align 8
  %slot_r2_5 = alloca i16, align 8
  %slot___destr_6 = alloca { i16, i16 }, align 8
  %slot___destr_7 = alloca { i16, i16 }, align 8
  %t1 = load { { ptr, i64 } }, ptr %p_g
  %t2 = call { i16, i16 } @aila_std_textgrid__cursor({ { ptr, i64 } } %t1)
  store { i16, i16 } %t2, ptr %slot___destr_6
  %t3 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_6, i32 0, i32 0
  %t4 = load i16, ptr %t3
  store i16 %t4, ptr %slot_c_2
  %t5 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_6, i32 0, i32 1
  %t6 = load i16, ptr %t5
  store i16 %t6, ptr %slot_r_3
  %t7 = load i16, ptr %slot_c_2
  %t8 = load i16, ptr %slot_r_3
  %t9 = call { i16, i16 } @aila_std_textgrid__newline_at(ptr %p_g, i16 %t7, i16 %t8)
  store { i16, i16 } %t9, ptr %slot___destr_7
  %t10 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_7, i32 0, i32 0
  %t11 = load i16, ptr %t10
  store i16 %t11, ptr %slot_c2_4
  %t12 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_7, i32 0, i32 1
  %t13 = load i16, ptr %t12
  store i16 %t13, ptr %slot_r2_5
  %t14 = load i16, ptr %slot_c2_4
  %t15 = load i16, ptr %slot_r2_5
  call void @aila_std_textgrid__set_cursor(ptr %p_g, i16 %t14, i16 %t15)
  ret void
}

define void @aila_std_textgrid__put_cp(ptr %p_g, ptr %p_gs, i32 %p_cp) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_gs_2 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  %slot_cp_3 = alloca i32, align 8
  %slot_c_4 = alloca i16, align 8
  %slot_r_5 = alloca i16, align 8
  %slot_c2_6 = alloca i16, align 8
  %slot_r2_7 = alloca i16, align 8
  %slot___destr_8 = alloca { i16, i16 }, align 8
  %slot___destr_9 = alloca { i16, i16 }, align 8
  store i32 %p_cp, ptr %slot_cp_3
  %t1 = load { { ptr, i64 } }, ptr %p_g
  %t2 = call { i16, i16 } @aila_std_textgrid__cursor({ { ptr, i64 } } %t1)
  store { i16, i16 } %t2, ptr %slot___destr_8
  %t3 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_8, i32 0, i32 0
  %t4 = load i16, ptr %t3
  store i16 %t4, ptr %slot_c_4
  %t5 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_8, i32 0, i32 1
  %t6 = load i16, ptr %t5
  store i16 %t6, ptr %slot_r_5
  %t7 = load i32, ptr %slot_cp_3
  %t8 = load i16, ptr %slot_c_4
  %t9 = load i16, ptr %slot_r_5
  %t10 = call { i16, i16 } @aila_std_textgrid__put_cp_at(ptr %p_g, ptr %p_gs, i32 %t7, i16 %t8, i16 %t9)
  store { i16, i16 } %t10, ptr %slot___destr_9
  %t11 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_9, i32 0, i32 0
  %t12 = load i16, ptr %t11
  store i16 %t12, ptr %slot_c2_6
  %t13 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_9, i32 0, i32 1
  %t14 = load i16, ptr %t13
  store i16 %t14, ptr %slot_r2_7
  %t15 = load i16, ptr %slot_c2_6
  %t16 = load i16, ptr %slot_r2_7
  call void @aila_std_textgrid__set_cursor(ptr %p_g, i16 %t15, i16 %t16)
  ret void
}

define void @aila_std_textgrid__put_text(ptr %p_g, ptr %p_gs, { ptr, i64 } %p_bytes, i64 %p_n) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_gs_2 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  %slot_bytes_3 = alloca { ptr, i64 }, align 8
  %slot_n_4 = alloca i64, align 8
  %slot_c0_5 = alloca i16, align 8
  %slot_r0_6 = alloca i16, align 8
  %slot_c_7 = alloca i16, align 8
  %slot_r_8 = alloca i16, align 8
  %slot_i_9 = alloca i64, align 8
  %slot_cp_10 = alloca i32, align 8
  %slot_step_11 = alloca i64, align 8
  %slot_c2_12 = alloca i16, align 8
  %slot_r2_13 = alloca i16, align 8
  %slot___destr_14 = alloca { i16, i16 }, align 8
  %slot___destr_15 = alloca { i32, i64 }, align 8
  %slot___destr_16 = alloca { i16, i16 }, align 8
  store { ptr, i64 } %p_bytes, ptr %slot_bytes_3
  store i64 %p_n, ptr %slot_n_4
  %t1 = load { { ptr, i64 } }, ptr %p_g
  %t2 = call { i16, i16 } @aila_std_textgrid__cursor({ { ptr, i64 } } %t1)
  store { i16, i16 } %t2, ptr %slot___destr_14
  %t3 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_14, i32 0, i32 0
  %t4 = load i16, ptr %t3
  store i16 %t4, ptr %slot_c0_5
  %t5 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_14, i32 0, i32 1
  %t6 = load i16, ptr %t5
  store i16 %t6, ptr %slot_r0_6
  %t7 = load i16, ptr %slot_c0_5
  store i16 %t7, ptr %slot_c_7
  %t8 = load i16, ptr %slot_r0_6
  store i16 %t8, ptr %slot_r_8
  store i64 0, ptr %slot_i_9
  br label %for.cond2
for.cond2:
  %t9 = load i64, ptr %slot_i_9
  %t10 = load i64, ptr %slot_n_4
  %t11 = icmp ult i64 %t9, %t10
  br i1 %t11, label %for.body4, label %for.end1
for.body4:
  %t12 = load { ptr, i64 }, ptr %slot_bytes_3
  %t13 = load i64, ptr %slot_n_4
  %t14 = load i64, ptr %slot_i_9
  %t15 = call { i32, i64 } @aila_std_paint__decode({ ptr, i64 } %t12, i64 %t13, i64 %t14)
  store { i32, i64 } %t15, ptr %slot___destr_15
  %t16 = getelementptr inbounds { i32, i64 }, ptr %slot___destr_15, i32 0, i32 0
  %t17 = load i32, ptr %t16
  store i32 %t17, ptr %slot_cp_10
  %t18 = getelementptr inbounds { i32, i64 }, ptr %slot___destr_15, i32 0, i32 1
  %t19 = load i64, ptr %t18
  store i64 %t19, ptr %slot_step_11
  %t20 = load i64, ptr %slot_step_11
  %t21 = icmp eq i64 %t20, 0
  br i1 %t21, label %if.then6, label %if.next7
if.then6:
  br label %for.end1
if.next7:
  br label %if.end5
if.end5:
  %t22 = load i32, ptr %slot_cp_10
  %t23 = load i16, ptr %slot_c_7
  %t24 = load i16, ptr %slot_r_8
  %t25 = call { i16, i16 } @aila_std_textgrid__put_cp_at(ptr %p_g, ptr %p_gs, i32 %t22, i16 %t23, i16 %t24)
  store { i16, i16 } %t25, ptr %slot___destr_16
  %t26 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_16, i32 0, i32 0
  %t27 = load i16, ptr %t26
  store i16 %t27, ptr %slot_c2_12
  %t28 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_16, i32 0, i32 1
  %t29 = load i16, ptr %t28
  store i16 %t29, ptr %slot_r2_13
  %t30 = load i16, ptr %slot_c2_12
  store i16 %t30, ptr %slot_c_7
  %t31 = load i16, ptr %slot_r2_13
  store i16 %t31, ptr %slot_r_8
  %t32 = load i64, ptr %slot_i_9
  %t33 = load i64, ptr %slot_step_11
  %t34 = add i64 %t32, %t33
  store i64 %t34, ptr %slot_i_9
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t35 = load i16, ptr %slot_c_7
  %t36 = load i16, ptr %slot_r_8
  call void @aila_std_textgrid__set_cursor(ptr %p_g, i16 %t35, i16 %t36)
  ret void
}

define void @aila_std_textgrid__put_cps(ptr %p_g, ptr %p_gs, { ptr, i64 } %p_cps, i64 %p_n) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_gs_2 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  %slot_cps_3 = alloca { ptr, i64 }, align 8
  %slot_n_4 = alloca i64, align 8
  %slot_c0_5 = alloca i16, align 8
  %slot_r0_6 = alloca i16, align 8
  %slot_c_7 = alloca i16, align 8
  %slot_r_8 = alloca i16, align 8
  %slot_i_9 = alloca i64, align 8
  %slot_c2_10 = alloca i16, align 8
  %slot_r2_11 = alloca i16, align 8
  %slot___destr_12 = alloca { i16, i16 }, align 8
  %slot___destr_13 = alloca { i16, i16 }, align 8
  store { ptr, i64 } %p_cps, ptr %slot_cps_3
  store i64 %p_n, ptr %slot_n_4
  %t1 = load { { ptr, i64 } }, ptr %p_g
  %t2 = call { i16, i16 } @aila_std_textgrid__cursor({ { ptr, i64 } } %t1)
  store { i16, i16 } %t2, ptr %slot___destr_12
  %t3 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_12, i32 0, i32 0
  %t4 = load i16, ptr %t3
  store i16 %t4, ptr %slot_c0_5
  %t5 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_12, i32 0, i32 1
  %t6 = load i16, ptr %t5
  store i16 %t6, ptr %slot_r0_6
  %t7 = load i16, ptr %slot_c0_5
  store i16 %t7, ptr %slot_c_7
  %t8 = load i16, ptr %slot_r0_6
  store i16 %t8, ptr %slot_r_8
  store i64 0, ptr %slot_i_9
  br label %for.cond2
for.cond2:
  %t9 = load i64, ptr %slot_i_9
  %t10 = load i64, ptr %slot_n_4
  %t11 = icmp ult i64 %t9, %t10
  br i1 %t11, label %for.body4, label %for.end1
for.body4:
  %t12 = load i64, ptr %slot_i_9
  %t13 = load { ptr, i64 }, ptr %slot_cps_3
  %t14 = extractvalue { ptr, i64 } %t13, 0
  %t15 = extractvalue { ptr, i64 } %t13, 1
  %t16 = icmp uge i64 %t12, %t15
  br i1 %t16, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 216)
  unreachable
chk.ok5:
  %t17 = getelementptr inbounds i32, ptr %t14, i64 %t12
  %t18 = load i32, ptr %t17
  %t19 = load i16, ptr %slot_c_7
  %t20 = load i16, ptr %slot_r_8
  %t21 = call { i16, i16 } @aila_std_textgrid__put_cp_at(ptr %p_g, ptr %p_gs, i32 %t18, i16 %t19, i16 %t20)
  store { i16, i16 } %t21, ptr %slot___destr_13
  %t22 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_13, i32 0, i32 0
  %t23 = load i16, ptr %t22
  store i16 %t23, ptr %slot_c2_10
  %t24 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_13, i32 0, i32 1
  %t25 = load i16, ptr %t24
  store i16 %t25, ptr %slot_r2_11
  %t26 = load i16, ptr %slot_c2_10
  store i16 %t26, ptr %slot_c_7
  %t27 = load i16, ptr %slot_r2_11
  store i16 %t27, ptr %slot_r_8
  %t28 = load i64, ptr %slot_i_9
  %t29 = add i64 %t28, 1
  store i64 %t29, ptr %slot_i_9
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t30 = load i16, ptr %slot_c_7
  %t31 = load i16, ptr %slot_r_8
  call void @aila_std_textgrid__set_cursor(ptr %p_g, i16 %t30, i16 %t31)
  ret void
}

define void @aila_std_textgrid__render(ptr %p_g, ptr %p_gs, ptr %p_pb, ptr %p_d) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_gs_2 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  %slot_pb_3 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_d_4 = alloca { { ptr, i64 } }, align 8
  call void @aila_std_textgrid__render_impl(ptr %p_g, ptr %p_gs, ptr %p_pb, ptr %p_d, i1 0)
  ret void
}

define void @aila_std_textgrid__render_all(ptr %p_g, ptr %p_gs, ptr %p_pb, ptr %p_d) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_gs_2 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  %slot_pb_3 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_d_4 = alloca { { ptr, i64 } }, align 8
  call void @aila_std_textgrid__render_impl(ptr %p_g, ptr %p_gs, ptr %p_pb, ptr %p_d, i1 1)
  ret void
}

define i16 @aila_std_textgrid__scroll_cells(ptr %p_g, i16 %p_n, i16 %p_cur_row) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_n_2 = alloca i16, align 8
  %slot_cur_row_3 = alloca i16, align 8
  %slot_cw_4 = alloca i64, align 8
  %slot_rw_5 = alloca i64, align 8
  %slot_total_6 = alloca i64, align 8
  %slot_base_7 = alloca i64, align 8
  %slot_t_8 = alloca i64, align 8
  %slot_shift_9 = alloca i64, align 8
  %slot_i_10 = alloca i64, align 8
  store i16 %p_n, ptr %slot_n_2
  store i16 %p_cur_row, ptr %slot_cur_row_3
  %t1 = load { { ptr, i64 } }, ptr %p_g
  %t2 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t1)
  %t3 = zext i16 %t2 to i64
  store i64 %t3, ptr %slot_cw_4
  %t4 = load { { ptr, i64 } }, ptr %p_g
  %t5 = call i16 @aila_std_textgrid__rows_of({ { ptr, i64 } } %t4)
  %t6 = zext i16 %t5 to i64
  store i64 %t6, ptr %slot_rw_5
  %t7 = load i64, ptr %slot_cw_4
  %t8 = load i64, ptr %slot_rw_5
  %t9 = mul i64 %t7, %t8
  store i64 %t9, ptr %slot_total_6
  %t10 = call i64 @aila_std_textgrid__cells_at()
  store i64 %t10, ptr %slot_base_7
  %t11 = load i16, ptr %slot_n_2
  %t12 = load { { ptr, i64 } }, ptr %p_g
  %t13 = call i16 @aila_std_textgrid__rows_of({ { ptr, i64 } } %t12)
  %t14 = icmp uge i16 %t11, %t13
  br i1 %t14, label %if.then2, label %if.next3
if.then2:
  store i64 0, ptr %slot_t_8
  br label %for.cond5
for.cond5:
  %t15 = load i64, ptr %slot_t_8
  %t16 = load i64, ptr %slot_total_6
  %t17 = icmp ult i64 %t15, %t16
  br i1 %t17, label %for.body7, label %for.end4
for.body7:
  %t18 = load i64, ptr %slot_base_7
  %t19 = load i64, ptr %slot_t_8
  %t20 = add i64 %t18, %t19
  %t21 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t22 = load { ptr, i64 }, ptr %t21
  %t23 = extractvalue { ptr, i64 } %t22, 0
  %t24 = extractvalue { ptr, i64 } %t22, 1
  %t25 = icmp uge i64 %t20, %t24
  br i1 %t25, label %chk.bad9, label %chk.ok8
chk.bad9:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 243)
  unreachable
chk.ok8:
  %t26 = getelementptr inbounds i32, ptr %t23, i64 %t20
  store i32 0, ptr %t26
  %t27 = load i64, ptr %slot_t_8
  %t28 = add i64 %t27, 1
  store i64 %t28, ptr %slot_t_8
  br label %for.step6
for.step6:
  br label %for.cond5
for.end4:
  ret i16 0
if.next3:
  br label %if.end1
if.end1:
  %t29 = load i16, ptr %slot_n_2
  %t30 = zext i16 %t29 to i64
  %t31 = load i64, ptr %slot_cw_4
  %t32 = mul i64 %t30, %t31
  store i64 %t32, ptr %slot_shift_9
  store i64 0, ptr %slot_i_10
  br label %for.cond11
for.cond11:
  %t33 = load i64, ptr %slot_i_10
  %t34 = load i64, ptr %slot_total_6
  %t35 = icmp ult i64 %t33, %t34
  br i1 %t35, label %for.body13, label %for.end10
for.body13:
  %t36 = load i64, ptr %slot_i_10
  %t37 = load i64, ptr %slot_shift_9
  %t38 = add i64 %t36, %t37
  %t39 = load i64, ptr %slot_total_6
  %t40 = icmp ult i64 %t38, %t39
  br i1 %t40, label %if.then15, label %if.next16
if.then15:
  %t41 = load i64, ptr %slot_base_7
  %t42 = load i64, ptr %slot_i_10
  %t43 = add i64 %t41, %t42
  %t44 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t45 = load { ptr, i64 }, ptr %t44
  %t46 = extractvalue { ptr, i64 } %t45, 0
  %t47 = extractvalue { ptr, i64 } %t45, 1
  %t48 = icmp uge i64 %t43, %t47
  br i1 %t48, label %chk.bad18, label %chk.ok17
chk.bad18:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 252)
  unreachable
chk.ok17:
  %t49 = getelementptr inbounds i32, ptr %t46, i64 %t43
  %t50 = load i64, ptr %slot_base_7
  %t51 = load i64, ptr %slot_i_10
  %t52 = add i64 %t50, %t51
  %t53 = load i64, ptr %slot_shift_9
  %t54 = add i64 %t52, %t53
  %t55 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t56 = load { ptr, i64 }, ptr %t55
  %t57 = extractvalue { ptr, i64 } %t56, 0
  %t58 = extractvalue { ptr, i64 } %t56, 1
  %t59 = icmp uge i64 %t54, %t58
  br i1 %t59, label %chk.bad20, label %chk.ok19
chk.bad20:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 252)
  unreachable
chk.ok19:
  %t60 = getelementptr inbounds i32, ptr %t57, i64 %t54
  %t61 = load i32, ptr %t60
  store i32 %t61, ptr %t49
  br label %if.end14
if.next16:
  %t62 = load i64, ptr %slot_base_7
  %t63 = load i64, ptr %slot_i_10
  %t64 = add i64 %t62, %t63
  %t65 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t66 = load { ptr, i64 }, ptr %t65
  %t67 = extractvalue { ptr, i64 } %t66, 0
  %t68 = extractvalue { ptr, i64 } %t66, 1
  %t69 = icmp uge i64 %t64, %t68
  br i1 %t69, label %chk.bad22, label %chk.ok21
chk.bad22:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 254)
  unreachable
chk.ok21:
  %t70 = getelementptr inbounds i32, ptr %t67, i64 %t64
  store i32 0, ptr %t70
  br label %if.end14
if.end14:
  %t71 = load i64, ptr %slot_i_10
  %t72 = add i64 %t71, 1
  store i64 %t72, ptr %slot_i_10
  br label %for.step12
for.step12:
  br label %for.cond11
for.end10:
  %t73 = load i16, ptr %slot_cur_row_3
  %t74 = load i16, ptr %slot_n_2
  %t75 = icmp uge i16 %t73, %t74
  br i1 %t75, label %if.then24, label %if.next25
if.then24:
  %t76 = load i16, ptr %slot_cur_row_3
  %t77 = load i16, ptr %slot_n_2
  %t78 = sub i16 %t76, %t77
  ret i16 %t78
if.next25:
  br label %if.end23
if.end23:
  ret i16 0
}

define { i16, i16 } @aila_std_textgrid__newline_at(ptr %p_g, i16 %p_col, i16 %p_row) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_col_2 = alloca i16, align 8
  %slot_row_3 = alloca i16, align 8
  %slot_rows_4 = alloca i16, align 8
  %slot_r_5 = alloca i16, align 8
  store i16 %p_col, ptr %slot_col_2
  store i16 %p_row, ptr %slot_row_3
  %t1 = load i16, ptr %slot_col_2
  %t2 = load { { ptr, i64 } }, ptr %p_g
  %t3 = call i16 @aila_std_textgrid__rows_of({ { ptr, i64 } } %t2)
  store i16 %t3, ptr %slot_rows_4
  %t4 = load i16, ptr %slot_rows_4
  %t5 = icmp eq i16 %t4, 0
  br i1 %t5, label %if.then2, label %if.next3
if.then2:
  %t6 = trunc i64 0 to i16
  %t7 = trunc i64 0 to i16
  %t8 = insertvalue { i16, i16 } undef, i16 %t6, 0
  %t9 = insertvalue { i16, i16 } %t8, i16 %t7, 1
  ret { i16, i16 } %t9
if.next3:
  br label %if.end1
if.end1:
  %t10 = load i16, ptr %slot_row_3
  %t11 = add i16 %t10, 1
  %t12 = load i16, ptr %slot_rows_4
  %t13 = icmp uge i16 %t11, %t12
  br i1 %t13, label %if.then5, label %if.next6
if.then5:
  %t14 = load i16, ptr %slot_row_3
  %t15 = call i16 @aila_std_textgrid__scroll_cells(ptr %p_g, i16 1, i16 %t14)
  store i16 %t15, ptr %slot_r_5
  %t16 = trunc i64 0 to i16
  %t17 = load i16, ptr %slot_r_5
  %t18 = insertvalue { i16, i16 } undef, i16 %t16, 0
  %t19 = insertvalue { i16, i16 } %t18, i16 %t17, 1
  ret { i16, i16 } %t19
if.next6:
  br label %if.end4
if.end4:
  %t20 = trunc i64 0 to i16
  %t21 = load i16, ptr %slot_row_3
  %t22 = add i16 %t21, 1
  %t23 = insertvalue { i16, i16 } undef, i16 %t20, 0
  %t24 = insertvalue { i16, i16 } %t23, i16 %t22, 1
  ret { i16, i16 } %t24
}

define { i16, i16 } @aila_std_textgrid__put_cp_at(ptr %p_g, ptr %p_gs, i32 %p_cp, i16 %p_col, i16 %p_row) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_gs_2 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  %slot_cp_3 = alloca i32, align 8
  %slot_col_4 = alloca i16, align 8
  %slot_row_5 = alloca i16, align 8
  %slot_wd_6 = alloca i32, align 8
  %slot_c_7 = alloca i16, align 8
  %slot_r_8 = alloca i16, align 8
  %slot_c2_9 = alloca i16, align 8
  %slot_r2_10 = alloca i16, align 8
  %slot_at_11 = alloca i64, align 8
  %slot___destr_12 = alloca { i16, i16 }, align 8
  %slot___sc_13 = alloca i1, align 8
  store i32 %p_cp, ptr %slot_cp_3
  store i16 %p_col, ptr %slot_col_4
  store i16 %p_row, ptr %slot_row_5
  %t1 = load i32, ptr %slot_cp_3
  %t2 = call i16 @aila_std_paint__glyph_width(ptr %p_gs, i32 %t1)
  %t3 = zext i16 %t2 to i32
  store i32 %t3, ptr %slot_wd_6
  %t4 = load i32, ptr %slot_wd_6
  %t5 = icmp eq i32 %t4, 0
  br i1 %t5, label %if.then2, label %if.next3
if.then2:
  store i32 1, ptr %slot_wd_6
  br label %if.end1
if.next3:
  br label %if.end1
if.end1:
  %t6 = load i32, ptr %slot_wd_6
  %t7 = icmp ugt i32 %t6, 2
  br i1 %t7, label %if.then5, label %if.next6
if.then5:
  store i32 2, ptr %slot_wd_6
  br label %if.end4
if.next6:
  br label %if.end4
if.end4:
  %t8 = load i16, ptr %slot_col_4
  store i16 %t8, ptr %slot_c_7
  %t9 = load i16, ptr %slot_row_5
  store i16 %t9, ptr %slot_r_8
  %t10 = load i16, ptr %slot_c_7
  %t11 = load i32, ptr %slot_wd_6
  %t12 = trunc i32 %t11 to i16
  %t13 = add i16 %t10, %t12
  %t14 = load { { ptr, i64 } }, ptr %p_g
  %t15 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t14)
  %t16 = icmp ugt i16 %t13, %t15
  br i1 %t16, label %if.then8, label %if.next9
if.then8:
  %t17 = load i16, ptr %slot_c_7
  %t18 = load i16, ptr %slot_r_8
  %t19 = call { i16, i16 } @aila_std_textgrid__newline_at(ptr %p_g, i16 %t17, i16 %t18)
  store { i16, i16 } %t19, ptr %slot___destr_12
  %t20 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_12, i32 0, i32 0
  %t21 = load i16, ptr %t20
  store i16 %t21, ptr %slot_c2_9
  %t22 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_12, i32 0, i32 1
  %t23 = load i16, ptr %t22
  store i16 %t23, ptr %slot_r2_10
  %t24 = load i16, ptr %slot_c2_9
  store i16 %t24, ptr %slot_c_7
  %t25 = load i16, ptr %slot_r2_10
  store i16 %t25, ptr %slot_r_8
  br label %if.end7
if.next9:
  br label %if.end7
if.end7:
  %t26 = load { { ptr, i64 } }, ptr %p_g
  %t27 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t26)
  %t28 = icmp eq i16 %t27, 0
  br i1 %t28, label %sc.l13, label %sc.r14
sc.l13:
  store i1 1, ptr %slot___sc_13
  br label %sc.end15
sc.r14:
  %t29 = load { { ptr, i64 } }, ptr %p_g
  %t30 = call i16 @aila_std_textgrid__rows_of({ { ptr, i64 } } %t29)
  %t31 = icmp eq i16 %t30, 0
  store i1 %t31, ptr %slot___sc_13
  br label %sc.end15
sc.end15:
  %t32 = load i1, ptr %slot___sc_13
  br i1 %t32, label %if.then11, label %if.next12
if.then11:
  %t33 = load i16, ptr %slot_c_7
  %t34 = load i16, ptr %slot_r_8
  %t35 = insertvalue { i16, i16 } undef, i16 %t33, 0
  %t36 = insertvalue { i16, i16 } %t35, i16 %t34, 1
  ret { i16, i16 } %t36
if.next12:
  br label %if.end10
if.end10:
  %t37 = call i64 @aila_std_textgrid__cells_at()
  %t38 = load { { ptr, i64 } }, ptr %p_g
  %t39 = load i16, ptr %slot_c_7
  %t40 = load i16, ptr %slot_r_8
  %t41 = call i64 @aila_std_textgrid__idx({ { ptr, i64 } } %t38, i16 %t39, i16 %t40)
  %t42 = add i64 %t37, %t41
  store i64 %t42, ptr %slot_at_11
  %t43 = load i64, ptr %slot_at_11
  %t44 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t45 = load { ptr, i64 }, ptr %t44
  %t46 = extractvalue { ptr, i64 } %t45, 1
  %t48 = icmp uge i64 %t43, %t46
  br i1 %t48, label %if.then17, label %if.next18
if.then17:
  %t49 = load i16, ptr %slot_c_7
  %t50 = load i16, ptr %slot_r_8
  %t51 = insertvalue { i16, i16 } undef, i16 %t49, 0
  %t52 = insertvalue { i16, i16 } %t51, i16 %t50, 1
  ret { i16, i16 } %t52
if.next18:
  br label %if.end16
if.end16:
  %t53 = load i64, ptr %slot_at_11
  %t54 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t55 = load { ptr, i64 }, ptr %t54
  %t56 = extractvalue { ptr, i64 } %t55, 0
  %t57 = extractvalue { ptr, i64 } %t55, 1
  %t58 = icmp uge i64 %t53, %t57
  br i1 %t58, label %chk.bad20, label %chk.ok19
chk.bad20:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 301)
  unreachable
chk.ok19:
  %t59 = getelementptr inbounds i32, ptr %t56, i64 %t53
  %t60 = load i32, ptr %slot_cp_3
  %t61 = load i32, ptr %slot_wd_6
  %t62 = call i32 @aila_std_textgrid__encode(i32 %t60, i32 %t61)
  store i32 %t62, ptr %t59
  %t63 = load i32, ptr %slot_wd_6
  %t64 = icmp eq i32 %t63, 2
  br i1 %t64, label %if.then22, label %if.next23
if.then22:
  %t65 = load i16, ptr %slot_c_7
  %t66 = add i16 %t65, 1
  %t67 = load { { ptr, i64 } }, ptr %p_g
  %t68 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t67)
  %t69 = icmp ult i16 %t66, %t68
  br i1 %t69, label %if.then25, label %if.next26
if.then25:
  %t70 = load i64, ptr %slot_at_11
  %t71 = add i64 %t70, 1
  %t72 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t73 = load { ptr, i64 }, ptr %t72
  %t74 = extractvalue { ptr, i64 } %t73, 0
  %t75 = extractvalue { ptr, i64 } %t73, 1
  %t76 = icmp uge i64 %t71, %t75
  br i1 %t76, label %chk.bad28, label %chk.ok27
chk.bad28:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 304)
  unreachable
chk.ok27:
  %t77 = getelementptr inbounds i32, ptr %t74, i64 %t71
  store i32 0, ptr %t77
  br label %if.end24
if.next26:
  br label %if.end24
if.end24:
  br label %if.end21
if.next23:
  br label %if.end21
if.end21:
  %t78 = load i16, ptr %slot_c_7
  %t79 = load i32, ptr %slot_wd_6
  %t80 = trunc i32 %t79 to i16
  %t81 = add i16 %t78, %t80
  %t82 = load i16, ptr %slot_r_8
  %t83 = insertvalue { i16, i16 } undef, i16 %t81, 0
  %t84 = insertvalue { i16, i16 } %t83, i16 %t82, 1
  ret { i16, i16 } %t84
}

define void @aila_std_textgrid__render_impl(ptr %p_g, ptr %p_gs, ptr %p_pb, ptr %p_d, i1 %p_force) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_gs_2 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  %slot_pb_3 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_d_4 = alloca { { ptr, i64 } }, align 8
  %slot_force_5 = alloca i1, align 8
  %slot_row_6 = alloca i16, align 8
  %slot_mask_7 = alloca i64, align 8
  %slot_col_8 = alloca i16, align 8
  store i1 %p_force, ptr %slot_force_5
  store i16 0, ptr %slot_row_6
  br label %for.cond2
for.cond2:
  %t1 = load i16, ptr %slot_row_6
  %t2 = load { { ptr, i64 } }, ptr %p_g
  %t3 = call i16 @aila_std_textgrid__rows_of({ { ptr, i64 } } %t2)
  %t4 = icmp ult i16 %t1, %t3
  br i1 %t4, label %for.body4, label %for.end1
for.body4:
  %t5 = load i16, ptr %slot_row_6
  %t6 = load i1, ptr %slot_force_5
  %t7 = call i64 @aila_std_textgrid__render_row(ptr %p_g, ptr %p_gs, ptr %p_pb, i16 %t5, i1 %t6)
  store i64 %t7, ptr %slot_mask_7
  store i16 0, ptr %slot_col_8
  br label %for.cond6
for.cond6:
  %t8 = load i16, ptr %slot_col_8
  %t9 = load { { ptr, i64 } }, ptr %p_g
  %t10 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t9)
  %t11 = icmp ult i16 %t8, %t10
  br i1 %t11, label %for.body8, label %for.end5
for.body8:
  %t12 = load i64, ptr %slot_mask_7
  %t14 = load i16, ptr %slot_col_8
  %t15 = zext i16 %t14 to i64
  %t16 = icmp uge i64 %t15, 64
  br i1 %t16, label %chk.bad13, label %chk.ok12
chk.bad13:
  call void @aila_rt_trap(ptr @.str.0, i64 24, ptr @.str.1, i64 317)
  unreachable
chk.ok12:
  %t17 = shl i64 1, %t15
  %t18 = and i64 %t12, %t17
  %t19 = icmp ne i64 %t18, 0
  br i1 %t19, label %if.then10, label %if.next11
if.then10:
  %t20 = load i16, ptr %slot_col_8
  %t21 = load i16, ptr %slot_row_6
  call void @aila_std_damage__add_cell(ptr %p_d, i16 %t20, i16 %t21)
  br label %if.end9
if.next11:
  br label %if.end9
if.end9:
  %t22 = load i16, ptr %slot_col_8
  %t23 = add i16 %t22, 1
  store i16 %t23, ptr %slot_col_8
  br label %for.step7
for.step7:
  br label %for.cond6
for.end5:
  %t24 = load i16, ptr %slot_row_6
  %t25 = add i16 %t24, 1
  store i16 %t25, ptr %slot_row_6
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  ret void
}

define i64 @aila_std_textgrid__render_row(ptr %p_g, ptr %p_gs, ptr %p_pb, i16 %p_row, i1 %p_force) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_gs_2 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  %slot_pb_3 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_row_4 = alloca i16, align 8
  %slot_force_5 = alloca i1, align 8
  %slot_mask_6 = alloca i64, align 8
  %slot_n_7 = alloca i64, align 8
  %slot_pv_8 = alloca i64, align 8
  %slot_col_9 = alloca i16, align 8
  %slot_i_10 = alloca i64, align 8
  %slot_cur_11 = alloca i32, align 8
  %slot_old_12 = alloca i32, align 8
  %slot_span_13 = alloca i16, align 8
  %slot___sc_14 = alloca i1, align 8
  store i16 %p_row, ptr %slot_row_4
  store i1 %p_force, ptr %slot_force_5
  store i64 0, ptr %slot_mask_6
  %t1 = load { { ptr, i64 } }, ptr %p_g
  %t2 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t1)
  %t3 = zext i16 %t2 to i64
  %t4 = load { { ptr, i64 } }, ptr %p_g
  %t5 = call i16 @aila_std_textgrid__rows_of({ { ptr, i64 } } %t4)
  %t6 = zext i16 %t5 to i64
  %t7 = mul i64 %t3, %t6
  store i64 %t7, ptr %slot_n_7
  %t8 = load { { ptr, i64 } }, ptr %p_g
  %t9 = call i64 @aila_std_textgrid__prev_at({ { ptr, i64 } } %t8)
  store i64 %t9, ptr %slot_pv_8
  store i16 0, ptr %slot_col_9
  br label %for.cond2
for.cond2:
  %t10 = load i16, ptr %slot_col_9
  %t11 = load { { ptr, i64 } }, ptr %p_g
  %t12 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t11)
  %t13 = icmp ult i16 %t10, %t12
  br i1 %t13, label %for.body4, label %for.end1
for.body4:
  %t14 = load { { ptr, i64 } }, ptr %p_g
  %t15 = load i16, ptr %slot_col_9
  %t16 = load i16, ptr %slot_row_4
  %t17 = call i64 @aila_std_textgrid__idx({ { ptr, i64 } } %t14, i16 %t15, i16 %t16)
  store i64 %t17, ptr %slot_i_10
  %t18 = call i64 @aila_std_textgrid__cells_at()
  %t19 = load i64, ptr %slot_i_10
  %t20 = add i64 %t18, %t19
  %t21 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t22 = load { ptr, i64 }, ptr %t21
  %t23 = extractvalue { ptr, i64 } %t22, 0
  %t24 = extractvalue { ptr, i64 } %t22, 1
  %t25 = icmp uge i64 %t20, %t24
  br i1 %t25, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 338)
  unreachable
chk.ok5:
  %t26 = getelementptr inbounds i32, ptr %t23, i64 %t20
  %t27 = load i32, ptr %t26
  store i32 %t27, ptr %slot_cur_11
  %t28 = load i64, ptr %slot_pv_8
  %t29 = load i64, ptr %slot_i_10
  %t30 = add i64 %t28, %t29
  %t31 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t32 = load { ptr, i64 }, ptr %t31
  %t33 = extractvalue { ptr, i64 } %t32, 0
  %t34 = extractvalue { ptr, i64 } %t32, 1
  %t35 = icmp uge i64 %t30, %t34
  br i1 %t35, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 339)
  unreachable
chk.ok7:
  %t36 = getelementptr inbounds i32, ptr %t33, i64 %t30
  %t37 = load i32, ptr %t36
  store i32 %t37, ptr %slot_old_12
  %t38 = load i16, ptr %slot_col_9
  %t39 = icmp ugt i16 %t38, 0
  br i1 %t39, label %if.then10, label %if.next11
if.then10:
  %t40 = call i64 @aila_std_textgrid__cells_at()
  %t41 = load i64, ptr %slot_i_10
  %t42 = add i64 %t40, %t41
  %t43 = sub i64 %t42, 1
  %t44 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t45 = load { ptr, i64 }, ptr %t44
  %t46 = extractvalue { ptr, i64 } %t45, 0
  %t47 = extractvalue { ptr, i64 } %t45, 1
  %t48 = icmp uge i64 %t43, %t47
  br i1 %t48, label %chk.bad16, label %chk.ok15
chk.bad16:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 342)
  unreachable
chk.ok15:
  %t49 = getelementptr inbounds i32, ptr %t46, i64 %t43
  %t50 = load i32, ptr %t49
  %t51 = call i32 @aila_std_textgrid__cell_w(i32 %t50)
  %t52 = icmp eq i32 %t51, 2
  br i1 %t52, label %if.then13, label %if.next14
if.then13:
  %t53 = load i16, ptr %slot_col_9
  %t54 = add i16 %t53, 1
  store i16 %t54, ptr %slot_col_9
  br label %for.step3
if.next14:
  br label %if.end12
if.end12:
  br label %if.end9
if.next11:
  br label %if.end9
if.end9:
  %t55 = load i1, ptr %slot_force_5
  %t56 = xor i1 %t55, true
  br i1 %t56, label %sc.r21, label %sc.l20
sc.l20:
  store i1 0, ptr %slot___sc_14
  br label %sc.end22
sc.r21:
  %t57 = load i32, ptr %slot_cur_11
  %t58 = load i32, ptr %slot_old_12
  %t59 = icmp eq i32 %t57, %t58
  store i1 %t59, ptr %slot___sc_14
  br label %sc.end22
sc.end22:
  %t60 = load i1, ptr %slot___sc_14
  br i1 %t60, label %if.then18, label %if.next19
if.then18:
  %t61 = load i16, ptr %slot_col_9
  %t62 = add i16 %t61, 1
  store i16 %t62, ptr %slot_col_9
  br label %for.step3
if.next19:
  br label %if.end17
if.end17:
  %t63 = load { { ptr, i64 } }, ptr %p_g
  %t64 = load i32, ptr %slot_cur_11
  %t65 = load i32, ptr %slot_old_12
  %t66 = load i16, ptr %slot_col_9
  %t67 = call i16 @aila_std_textgrid__unit_span({ { ptr, i64 } } %t63, i32 %t64, i32 %t65, i16 %t66)
  store i16 %t67, ptr %slot_span_13
  %t68 = load i32, ptr %slot_cur_11
  %t69 = load i16, ptr %slot_col_9
  %t70 = load i16, ptr %slot_row_4
  %t71 = load i16, ptr %slot_span_13
  call void @aila_std_textgrid__repaint_unit(ptr %p_g, ptr %p_gs, ptr %p_pb, i32 %t68, i16 %t69, i16 %t70, i16 %t71)
  %t72 = load i64, ptr %slot_mask_6
  %t74 = load i16, ptr %slot_col_9
  %t75 = zext i16 %t74 to i64
  %t76 = icmp uge i64 %t75, 64
  br i1 %t76, label %chk.bad24, label %chk.ok23
chk.bad24:
  call void @aila_rt_trap(ptr @.str.0, i64 24, ptr @.str.1, i64 353)
  unreachable
chk.ok23:
  %t77 = shl i64 1, %t75
  %t78 = or i64 %t72, %t77
  store i64 %t78, ptr %slot_mask_6
  %t79 = load i16, ptr %slot_span_13
  %t80 = icmp eq i16 %t79, 2
  br i1 %t80, label %if.then26, label %if.next27
if.then26:
  %t81 = load i64, ptr %slot_mask_6
  %t83 = load i16, ptr %slot_col_9
  %t84 = add i16 %t83, 1
  %t85 = zext i16 %t84 to i64
  %t86 = icmp uge i64 %t85, 64
  br i1 %t86, label %chk.bad29, label %chk.ok28
chk.bad29:
  call void @aila_rt_trap(ptr @.str.0, i64 24, ptr @.str.1, i64 355)
  unreachable
chk.ok28:
  %t87 = shl i64 1, %t85
  %t88 = or i64 %t81, %t87
  store i64 %t88, ptr %slot_mask_6
  %t89 = load i64, ptr %slot_pv_8
  %t90 = load i64, ptr %slot_i_10
  %t91 = add i64 %t89, %t90
  %t92 = add i64 %t91, 1
  %t93 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t94 = load { ptr, i64 }, ptr %t93
  %t95 = extractvalue { ptr, i64 } %t94, 0
  %t96 = extractvalue { ptr, i64 } %t94, 1
  %t97 = icmp uge i64 %t92, %t96
  br i1 %t97, label %chk.bad31, label %chk.ok30
chk.bad31:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 356)
  unreachable
chk.ok30:
  %t98 = getelementptr inbounds i32, ptr %t95, i64 %t92
  %t99 = call i64 @aila_std_textgrid__cells_at()
  %t100 = load i64, ptr %slot_i_10
  %t101 = add i64 %t99, %t100
  %t102 = add i64 %t101, 1
  %t103 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t104 = load { ptr, i64 }, ptr %t103
  %t105 = extractvalue { ptr, i64 } %t104, 0
  %t106 = extractvalue { ptr, i64 } %t104, 1
  %t107 = icmp uge i64 %t102, %t106
  br i1 %t107, label %chk.bad33, label %chk.ok32
chk.bad33:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 356)
  unreachable
chk.ok32:
  %t108 = getelementptr inbounds i32, ptr %t105, i64 %t102
  %t109 = load i32, ptr %t108
  store i32 %t109, ptr %t98
  br label %if.end25
if.next27:
  br label %if.end25
if.end25:
  %t110 = load i64, ptr %slot_pv_8
  %t111 = load i64, ptr %slot_i_10
  %t112 = add i64 %t110, %t111
  %t113 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t114 = load { ptr, i64 }, ptr %t113
  %t115 = extractvalue { ptr, i64 } %t114, 0
  %t116 = extractvalue { ptr, i64 } %t114, 1
  %t117 = icmp uge i64 %t112, %t116
  br i1 %t117, label %chk.bad35, label %chk.ok34
chk.bad35:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 358)
  unreachable
chk.ok34:
  %t118 = getelementptr inbounds i32, ptr %t115, i64 %t112
  %t119 = load i32, ptr %slot_cur_11
  store i32 %t119, ptr %t118
  %t120 = load i16, ptr %slot_col_9
  %t121 = load i16, ptr %slot_span_13
  %t122 = add i16 %t120, %t121
  store i16 %t122, ptr %slot_col_9
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t123 = load i64, ptr %slot_n_7
  %t124 = load i64, ptr %slot_mask_6
  ret i64 %t124
}

define i16 @aila_std_textgrid__unit_span({ { ptr, i64 } } %p_g, i32 %p_cur, i32 %p_old, i16 %p_col) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_cur_2 = alloca i32, align 8
  %slot_old_3 = alloca i32, align 8
  %slot_col_4 = alloca i16, align 8
  %slot_span_5 = alloca i16, align 8
  %slot___sc_6 = alloca i1, align 8
  store { { ptr, i64 } } %p_g, ptr %slot_g_1
  store i32 %p_cur, ptr %slot_cur_2
  store i32 %p_old, ptr %slot_old_3
  store i16 %p_col, ptr %slot_col_4
  store i16 1, ptr %slot_span_5
  %t1 = load i32, ptr %slot_cur_2
  %t2 = call i32 @aila_std_textgrid__cell_w(i32 %t1)
  %t3 = icmp eq i32 %t2, 2
  br i1 %t3, label %sc.l4, label %sc.r5
sc.l4:
  store i1 1, ptr %slot___sc_6
  br label %sc.end6
sc.r5:
  %t4 = load i32, ptr %slot_old_3
  %t5 = call i32 @aila_std_textgrid__cell_w(i32 %t4)
  %t6 = icmp eq i32 %t5, 2
  store i1 %t6, ptr %slot___sc_6
  br label %sc.end6
sc.end6:
  %t7 = load i1, ptr %slot___sc_6
  br i1 %t7, label %if.then2, label %if.next3
if.then2:
  store i16 2, ptr %slot_span_5
  br label %if.end1
if.next3:
  br label %if.end1
if.end1:
  %t8 = load i16, ptr %slot_col_4
  %t9 = load i16, ptr %slot_span_5
  %t10 = add i16 %t8, %t9
  %t11 = load { { ptr, i64 } }, ptr %slot_g_1
  %t12 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t11)
  %t13 = icmp ugt i16 %t10, %t12
  br i1 %t13, label %if.then8, label %if.next9
if.then8:
  store i16 1, ptr %slot_span_5
  br label %if.end7
if.next9:
  br label %if.end7
if.end7:
  %t14 = load i16, ptr %slot_span_5
  ret i16 %t14
}

define void @aila_std_textgrid__repaint_unit(ptr %p_g, ptr %p_gs, ptr %p_pb, i32 %p_cur, i16 %p_col, i16 %p_row, i16 %p_span) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_gs_2 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  %slot_pb_3 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_cur_4 = alloca i32, align 8
  %slot_col_5 = alloca i16, align 8
  %slot_row_6 = alloca i16, align 8
  %slot_span_7 = alloca i16, align 8
  %slot_cw_8 = alloca i16, align 8
  %slot_ch_9 = alloca i16, align 8
  %slot_clear_r_10 = alloca { i16, i16, i16, i16 }, align 8
  %slot_nxt_11 = alloca i32, align 8
  store i32 %p_cur, ptr %slot_cur_4
  store i16 %p_col, ptr %slot_col_5
  store i16 %p_row, ptr %slot_row_6
  store i16 %p_span, ptr %slot_span_7
  %t1 = zext i8 8 to i16
  store i16 %t1, ptr %slot_cw_8
  %t2 = zext i8 16 to i16
  store i16 %t2, ptr %slot_ch_9
  %t3 = load i16, ptr %slot_col_5
  %t4 = load i16, ptr %slot_row_6
  %t5 = call { i16, i16, i16, i16 } @aila_std_paint__cell_rect(i16 %t3, i16 %t4)
  store { i16, i16, i16, i16 } %t5, ptr %slot_clear_r_10
  %t6 = load i16, ptr %slot_span_7
  %t7 = icmp eq i16 %t6, 2
  br i1 %t7, label %if.then2, label %if.next3
if.then2:
  %t8 = load i16, ptr %slot_col_5
  %t9 = load i16, ptr %slot_cw_8
  %t10 = mul i16 %t8, %t9
  %t11 = load i16, ptr %slot_row_6
  %t12 = load i16, ptr %slot_ch_9
  %t13 = mul i16 %t11, %t12
  %t14 = load i16, ptr %slot_col_5
  %t15 = add i16 %t14, 2
  %t16 = load i16, ptr %slot_cw_8
  %t17 = mul i16 %t15, %t16
  %t18 = load i16, ptr %slot_row_6
  %t19 = add i16 %t18, 1
  %t20 = load i16, ptr %slot_ch_9
  %t21 = mul i16 %t19, %t20
  %t22 = call { i16, i16, i16, i16 } @aila_std_damage__rect(i16 %t10, i16 %t13, i16 %t17, i16 %t21)
  store { i16, i16, i16, i16 } %t22, ptr %slot_clear_r_10
  br label %if.end1
if.next3:
  br label %if.end1
if.end1:
  %t23 = load { i16, i16, i16, i16 }, ptr %slot_clear_r_10
  call void @aila_std_paint__fill_rect(ptr %p_pb, { i16, i16, i16, i16 } %t23, i1 0)
  %t24 = load i32, ptr %slot_cur_4
  %t25 = call i32 @aila_std_textgrid__cell_w(i32 %t24)
  %t26 = icmp ne i32 %t25, 0
  br i1 %t26, label %if.then5, label %if.next6
if.then5:
  %t27 = load i32, ptr %slot_cur_4
  %t28 = call i32 @aila_std_textgrid__cell_cp(i32 %t27)
  %t29 = load i16, ptr %slot_col_5
  %t30 = load i16, ptr %slot_row_6
  %t31 = call i16 @aila_std_paint__blit_glyph(ptr %p_pb, ptr %p_gs, i32 %t28, i16 %t29, i16 %t30)
  br label %if.end4
if.next6:
  br label %if.end4
if.end4:
  %t32 = load i16, ptr %slot_span_7
  %t33 = icmp eq i16 %t32, 2
  br i1 %t33, label %if.then8, label %if.next9
if.then8:
  %t34 = call i64 @aila_std_textgrid__cells_at()
  %t35 = load { { ptr, i64 } }, ptr %p_g
  %t36 = load i16, ptr %slot_col_5
  %t37 = add i16 %t36, 1
  %t38 = load i16, ptr %slot_row_6
  %t39 = call i64 @aila_std_textgrid__idx({ { ptr, i64 } } %t35, i16 %t37, i16 %t38)
  %t40 = add i64 %t34, %t39
  %t41 = getelementptr inbounds { { ptr, i64 } }, ptr %p_g, i32 0, i32 0
  %t42 = load { ptr, i64 }, ptr %t41
  %t43 = extractvalue { ptr, i64 } %t42, 0
  %t44 = extractvalue { ptr, i64 } %t42, 1
  %t45 = icmp uge i64 %t40, %t44
  br i1 %t45, label %chk.bad11, label %chk.ok10
chk.bad11:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 390)
  unreachable
chk.ok10:
  %t46 = getelementptr inbounds i32, ptr %t43, i64 %t40
  %t47 = load i32, ptr %t46
  store i32 %t47, ptr %slot_nxt_11
  %t48 = load i32, ptr %slot_nxt_11
  %t49 = call i32 @aila_std_textgrid__cell_w(i32 %t48)
  %t50 = icmp ne i32 %t49, 0
  br i1 %t50, label %if.then13, label %if.next14
if.then13:
  %t51 = load i32, ptr %slot_nxt_11
  %t52 = call i32 @aila_std_textgrid__cell_cp(i32 %t51)
  %t53 = load i16, ptr %slot_col_5
  %t54 = add i16 %t53, 1
  %t55 = load i16, ptr %slot_row_6
  %t56 = call i16 @aila_std_paint__blit_glyph(ptr %p_pb, ptr %p_gs, i32 %t52, i16 %t54, i16 %t55)
  br label %if.end12
if.next14:
  br label %if.end12
if.end12:
  br label %if.end7
if.next9:
  br label %if.end7
if.end7:
  ret void
}
