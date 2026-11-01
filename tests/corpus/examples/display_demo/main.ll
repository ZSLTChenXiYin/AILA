; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [2 x i8] c".\00", align 1
@.str.1 = private unnamed_addr constant [1 x i8] c"\00", align 1
@.str.2 = private unnamed_addr constant [25 x i8] c"shift count out of range\00", align 1
@.str.3 = private unnamed_addr constant [10 x i8] c"main.aila\00", align 1
@.str.4 = private unnamed_addr constant [17 x i8] c"division by zero\00", align 1
@.str.5 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.6 = private unnamed_addr constant [12 x i8] c"AB\E6\B8\A9C\E5\BA\A6D!\00", align 1
@.str.7 = private unnamed_addr constant [6 x i8] c"line1\00", align 1
@.str.8 = private unnamed_addr constant [6 x i8] c"line2\00", align 1
@.str.9 = private unnamed_addr constant [6 x i8] c"line3\00", align 1
@.str.10 = private unnamed_addr constant [13 x i8] c"\E6\B8\A9\E5\BA\A6\E4\BD\A0\E5\A5\BD\00", align 1
@.str.11 = private unnamed_addr constant [21 x i8] c"0123456789ABCDEFGHIJ\00", align 1
@.str.12 = private unnamed_addr constant [55 x i8] c"canvas %ux%u stride=%u bytes=%u cells=%ux%u glyphs=%u\0A\00", align 1
@.str.13 = private unnamed_addr constant [84 x i8] c"[1] widths: A=%u \E6\B8\A9=%u \E4\B8\80=%u | \E6\B7\B7\E6\8E\92\E5\88\B0 cell %u\EF\BC\88A1+B1+\E6\B8\A92+C1+\E5\BA\A62+D1+!1=9\EF\BC\89\0A\00", align 1
@.str.14 = private unnamed_addr constant [27 x i8] c"[2a] wrap: cursor=(%u,%u)\0A\00", align 1
@.str.15 = private unnamed_addr constant [31 x i8] c"[2b] truncate: cursor=(%u,%u)\0A\00", align 1
@.str.16 = private unnamed_addr constant [38 x i8] c"[3a] first render: cells=%u pages=%u\0A\00", align 1
@.str.17 = private unnamed_addr constant [38 x i8] c"[3b] no change   : cells=%u pages=%u\0A\00", align 1
@.str.18 = private unnamed_addr constant [38 x i8] c"[3c] one cell    : cells=%u pages=%u\0A\00", align 1
@.str.19 = private unnamed_addr constant [37 x i8] c"[3d] pre-scroll : cells=%u pages=%u\0A\00", align 1
@.str.20 = private unnamed_addr constant [96 x i8] c"[3e] scroll(1)  : cells=%u pages=%u\EF\BC\88line1\E2\86\92line2 \E5\8F\AA\E5\B7\AE\E6\9C\AB\E4\BD\8D\EF\BC\8C\E6\BB\9A\E5\8A\A8\E4\B8\8D\E6\98\AF\E6\95\B4\E5\B1\8F\E9\87\8D\E5\8F\91\EF\BC\89\0A\00", align 1
@.str.21 = private unnamed_addr constant [37 x i8] c"[3f] pre-scroll2: cells=%u pages=%u\0A\00", align 1
@.str.22 = private unnamed_addr constant [79 x i8] c"[3g] scroll rep : cells=%u pages=%u\EF\BC\88\E5\86\85\E5\AE\B9\E6\A8\A1\E5\9E\8B\EF\BC\9A\E5\8F\AA\E8\84\8F\E7\9C\9F\E5\8F\98\E4\BA\86\E7\9A\84\E8\A1\8C\EF\BC\89\0A\00", align 1
@.str.23 = private unnamed_addr constant [104 x i8] c"[4] commit: page_mask=%u pages=%u bytes_on_wire=%u\EF\BC\88\E6\95\B4\E5\B1\8F 1024 \E5\AD\97\E8\8A\82 I2C \E2\89\8837ms\EF\BC\8C\E5\8D\95\E9\A1\B5 \E2\89\885ms\EF\BC\89\0A\00", align 1
@.str.24 = private unnamed_addr constant [48 x i8] c"[5] final: cells=%u clips=%u pages=%u fnv64=%u\0A\00", align 1


declare i16 @aila_std_textgrid__rows_of({ { ptr, i64 } })
declare i16 @aila_std_textgrid__cols_of({ { ptr, i64 } })
declare i32 @aila_std_textgrid__get({ { ptr, i64 } }, i16, i16)
declare i32 @aila_std_textgrid__cell_w(i32)
declare void @aila_std_io__print({ ptr, i64 })
declare i32 @aila_std_textgrid__cell_cp(i32)
declare void @aila_std_io__println({ ptr, i64 })
declare i1 @aila_std_damage__has_cell({ { ptr, i64 } }, i16, i16)
declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare { i16, i16, i16, i16 } @aila_std_damage__clip({ { ptr, i64 } }, i64)
declare i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 })
declare ptr @aila_rt_region_enter(i64)
declare ptr @aila_rt_region_alloc(ptr, i64, i64)
declare i64 @aila_std_textgrid__words(i16, i16)
declare { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } @aila_std_pixbuf__new({ ptr, i64 }, i16, i16)
declare { ptr, i64 } @aila_font_font16__font_cps()
declare { ptr, i64 } @aila_font_font16__font_bits()
declare { ptr, i64 } @aila_font_font16__font_widths()
declare i64 @aila_font_font16__font_count()
declare i16 @aila_font_font16__font_px_h()
declare { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 } @aila_std_paint__new_glyph_set({ ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16)
declare { { ptr, i64 } } @aila_std_textgrid__new({ ptr, i64 }, i16, i16)
declare { { ptr, i64 } } @aila_std_damage__new({ ptr, i64 })
declare { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 } @aila_std_display__new_commit(ptr, { { ptr, i64 } })
declare i64 @aila_std_pixbuf__row_bytes(ptr)
declare void @aila_std_io__printf({ ptr, i64 }, { ptr, i64 })
declare i16 @aila_std_paint__glyph_width(ptr, i32)
declare void @aila_std_paint__fill(ptr, i1)
declare i16 @aila_std_paint__text_cells(ptr, ptr, { ptr, i64 }, i64, i16, i16)
declare void @aila_std_textgrid__clear({ { ptr, i64 } })
declare void @aila_std_textgrid__put_text({ { ptr, i64 } }, ptr, { ptr, i64 }, i64)
declare { i16, i16 } @aila_std_textgrid__cursor({ { ptr, i64 } })
declare void @aila_std_textgrid__newline({ { ptr, i64 } })
declare void @aila_std_damage__clear({ { ptr, i64 } })
declare void @aila_std_textgrid__render_all({ { ptr, i64 } }, ptr, ptr, { { ptr, i64 } })
declare i64 @aila_std_damage__cell_count({ { ptr, i64 } })
declare void @aila_std_textgrid__render({ { ptr, i64 } }, ptr, ptr, { { ptr, i64 } })
declare i32 @aila_std_textgrid__encode(i32, i32)
declare void @aila_std_textgrid__set({ { ptr, i64 } }, i16, i16, i32)
declare void @aila_std_textgrid__scroll_by({ { ptr, i64 } }, i16)
declare void @aila_std_display__Commit__set_damage(ptr, { { ptr, i64 } })
declare { { ptr, i64 } } @aila_std_display__Commit__damage_of(ptr)
declare void @aila_std_paint__hline(ptr, i16, i16, i16, i1)
declare { i16, i16, i16, i16 } @aila_std_damage__rect(i16, i16, i16, i16)
declare void @aila_std_damage__add_rect({ { ptr, i64 } }, { i16, i16, i16, i16 })
declare i64 @aila_std_damage__clip_count({ { ptr, i64 } })
declare void @aila_std_display_host__sink_host_ascii(ptr)
declare void @aila_rt_region_leave(ptr)

define { i8, { { ptr, i64 } } } @aila_main__u(i64 %p_v) {
  %slot_v_1 = alloca i64, align 8
  %slot___lit_2 = alloca { i8, { { ptr, i64 } } }, align 8
  store i64 %p_v, ptr %slot_v_1
  %t1 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t1, ptr %slot___lit_2
  %t2 = load i64, ptr %slot_v_1
  %t4 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_2, i32 0, i32 1, i32 0
  store i64 %t2, ptr %t4
  %t5 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_2
  ret { i8, { { ptr, i64 } } } %t5
}

define void @aila_main__dump_grid({ { ptr, i64 } } %p_g) {
  %slot_g_1 = alloca { { ptr, i64 } }, align 8
  %slot_row_2 = alloca i16, align 8
  %slot_col_3 = alloca i16, align 8
  %slot_v_4 = alloca i32, align 8
  %slot___lit_5 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_6 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_7 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_8 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_9 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_10 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  store { { ptr, i64 } } %p_g, ptr %slot_g_1
  store i16 0, ptr %slot_row_2
  br label %for.cond2
for.cond2:
  %t1 = load i16, ptr %slot_row_2
  %t2 = load { { ptr, i64 } }, ptr %slot_g_1
  %t3 = call i16 @aila_std_textgrid__rows_of({ { ptr, i64 } } %t2)
  %t4 = icmp ult i16 %t1, %t3
  br i1 %t4, label %for.body4, label %for.end1
for.body4:
  store i16 0, ptr %slot_col_3
  br label %for.cond6
for.cond6:
  %t5 = load i16, ptr %slot_col_3
  %t6 = load { { ptr, i64 } }, ptr %slot_g_1
  %t7 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t6)
  %t8 = icmp ult i16 %t5, %t7
  br i1 %t8, label %for.body8, label %for.end5
for.body8:
  %t9 = load { { ptr, i64 } }, ptr %slot_g_1
  %t10 = load i16, ptr %slot_col_3
  %t11 = load i16, ptr %slot_row_2
  %t12 = call i32 @aila_std_textgrid__get({ { ptr, i64 } } %t9, i16 %t10, i16 %t11)
  store i32 %t12, ptr %slot_v_4
  %t13 = load i32, ptr %slot_v_4
  %t14 = call i32 @aila_std_textgrid__cell_w(i32 %t13)
  %t15 = icmp eq i32 %t14, 0
  br i1 %t15, label %if.then10, label %if.next11
if.then10:
  %t16 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t16, ptr %slot___lit_5
  %t17 = insertvalue { ptr, i64 } undef, ptr @.str.0, 0
  %t18 = insertvalue { ptr, i64 } %t17, i64 1, 1
  %t19 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_5, i32 0, i32 1, i32 0
  store { ptr, i64 } %t18, ptr %t19
  %t20 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_5
  %t21 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_6, i64 0
  store { i8, { { ptr, i64 } } } %t20, ptr %t21
  %t22 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_6, 0
  %t23 = insertvalue { ptr, i64 } %t22, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t23)
  br label %if.end9
if.next11:
  %t24 = insertvalue { i8, { { ptr, i64 } } } undef, i8 4, 0
  store { i8, { { ptr, i64 } } } %t24, ptr %slot___lit_7
  %t25 = load i32, ptr %slot_v_4
  %t26 = call i32 @aila_std_textgrid__cell_cp(i32 %t25)
  %t28 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_7, i32 0, i32 1, i32 0
  store i32 %t26, ptr %t28
  %t29 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_7
  %t30 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_8, i64 0
  store { i8, { { ptr, i64 } } } %t29, ptr %t30
  %t31 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_8, 0
  %t32 = insertvalue { ptr, i64 } %t31, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t32)
  br label %if.end9
if.end9:
  %t33 = load i16, ptr %slot_col_3
  %t34 = add i16 %t33, 1
  store i16 %t34, ptr %slot_col_3
  br label %for.step7
for.step7:
  br label %for.cond6
for.end5:
  %t35 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t35, ptr %slot___lit_9
  %t36 = insertvalue { ptr, i64 } undef, ptr @.str.1, 0
  %t37 = insertvalue { ptr, i64 } %t36, i64 0, 1
  %t38 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_9, i32 0, i32 1, i32 0
  store { ptr, i64 } %t37, ptr %t38
  %t39 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_9
  %t40 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_10, i64 0
  store { i8, { { ptr, i64 } } } %t39, ptr %t40
  %t41 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_10, 0
  %t42 = insertvalue { ptr, i64 } %t41, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t42)
  %t43 = load i16, ptr %slot_row_2
  %t44 = add i16 %t43, 1
  store i16 %t44, ptr %slot_row_2
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  ret void
}

define i8 @aila_main__page_mask({ { ptr, i64 } } %p_d) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_m_2 = alloca i8, align 8
  %slot_row_3 = alloca i16, align 8
  %slot_col_4 = alloca i16, align 8
  %slot_i_5 = alloca i64, align 8
  %slot_r_6 = alloca { i16, i16, i16, i16 }, align 8
  %slot_p_7 = alloca i16, align 8
  %slot_pe_8 = alloca i16, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  store i8 0, ptr %slot_m_2
  store i16 0, ptr %slot_row_3
  br label %for.cond2
for.cond2:
  %t1 = load i16, ptr %slot_row_3
  %t2 = icmp ult i16 %t1, 8
  br i1 %t2, label %for.body4, label %for.end1
for.body4:
  store i16 0, ptr %slot_col_4
  br label %for.cond6
for.cond6:
  %t3 = load i16, ptr %slot_col_4
  %t4 = icmp ult i16 %t3, 16
  br i1 %t4, label %for.body8, label %for.end5
for.body8:
  %t5 = load { { ptr, i64 } }, ptr %slot_d_1
  %t6 = load i16, ptr %slot_col_4
  %t7 = load i16, ptr %slot_row_3
  %t8 = call i1 @aila_std_damage__has_cell({ { ptr, i64 } } %t5, i16 %t6, i16 %t7)
  br i1 %t8, label %if.then10, label %if.next11
if.then10:
  %t9 = load i8, ptr %slot_m_2
  %t10 = trunc i64 3 to i8
  %t11 = load i16, ptr %slot_row_3
  %t12 = mul i16 %t11, 2
  %t13 = trunc i16 %t12 to i8
  %t14 = icmp uge i8 %t13, 8
  br i1 %t14, label %chk.bad13, label %chk.ok12
chk.bad13:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.3, i64 68)
  unreachable
chk.ok12:
  %t15 = shl i8 %t10, %t13
  %t16 = or i8 %t9, %t15
  store i8 %t16, ptr %slot_m_2
  br label %if.end9
if.next11:
  br label %if.end9
if.end9:
  %t17 = load i16, ptr %slot_col_4
  %t18 = add i16 %t17, 1
  store i16 %t18, ptr %slot_col_4
  br label %for.step7
for.step7:
  br label %for.cond6
for.end5:
  %t19 = load i16, ptr %slot_row_3
  %t20 = add i16 %t19, 1
  store i16 %t20, ptr %slot_row_3
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  store i64 0, ptr %slot_i_5
  br label %for.cond15
for.cond15:
  %t21 = load i64, ptr %slot_i_5
  %t22 = icmp ult i64 %t21, 8
  br i1 %t22, label %for.body17, label %for.end14
for.body17:
  %t23 = load { { ptr, i64 } }, ptr %slot_d_1
  %t24 = load i64, ptr %slot_i_5
  %t25 = call { i16, i16, i16, i16 } @aila_std_damage__clip({ { ptr, i64 } } %t23, i64 %t24)
  store { i16, i16, i16, i16 } %t25, ptr %slot_r_6
  %t26 = load { i16, i16, i16, i16 }, ptr %slot_r_6
  %t27 = call i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 } %t26)
  %t28 = xor i1 %t27, true
  br i1 %t28, label %if.then19, label %if.next20
if.then19:
  %t29 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_r_6, i32 0, i32 1
  %t30 = load i16, ptr %t29
  %t31 = icmp eq i16 8, 0
  br i1 %t31, label %chk.bad22, label %chk.ok21
chk.bad22:
  call void @aila_rt_trap(ptr @.str.4, i64 16, ptr @.str.3, i64 78)
  unreachable
chk.ok21:
  %t32 = udiv i16 %t30, 8
  store i16 %t32, ptr %slot_p_7
  %t33 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_r_6, i32 0, i32 3
  %t34 = load i16, ptr %t33
  %t35 = add i16 %t34, 7
  %t36 = icmp eq i16 8, 0
  br i1 %t36, label %chk.bad24, label %chk.ok23
chk.bad24:
  call void @aila_rt_trap(ptr @.str.4, i64 16, ptr @.str.3, i64 79)
  unreachable
chk.ok23:
  %t37 = udiv i16 %t35, 8
  store i16 %t37, ptr %slot_pe_8
  br label %for.cond26
for.cond26:
  %t38 = load i16, ptr %slot_p_7
  %t39 = load i16, ptr %slot_pe_8
  %t40 = icmp ult i16 %t38, %t39
  br i1 %t40, label %for.body28, label %for.end25
for.body28:
  %t41 = load i8, ptr %slot_m_2
  %t42 = trunc i64 1 to i8
  %t43 = load i16, ptr %slot_p_7
  %t44 = trunc i16 %t43 to i8
  %t45 = icmp uge i8 %t44, 8
  br i1 %t45, label %chk.bad30, label %chk.ok29
chk.bad30:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.3, i64 81)
  unreachable
chk.ok29:
  %t46 = shl i8 %t42, %t44
  %t47 = or i8 %t41, %t46
  store i8 %t47, ptr %slot_m_2
  %t48 = load i16, ptr %slot_p_7
  %t49 = add i16 %t48, 1
  store i16 %t49, ptr %slot_p_7
  br label %for.step27
for.step27:
  br label %for.cond26
for.end25:
  br label %if.end18
if.next20:
  br label %if.end18
if.end18:
  %t50 = load i64, ptr %slot_i_5
  %t51 = add i64 %t50, 1
  store i64 %t51, ptr %slot_i_5
  br label %for.step16
for.step16:
  br label %for.cond15
for.end14:
  %t52 = load i8, ptr %slot_m_2
  ret i8 %t52
}

define i64 @aila_main__pages_of({ { ptr, i64 } } %p_d) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_m_2 = alloca i8, align 8
  %slot_n_3 = alloca i64, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  %t1 = load { { ptr, i64 } }, ptr %slot_d_1
  %t2 = call i8 @aila_main__page_mask({ { ptr, i64 } } %t1)
  store i8 %t2, ptr %slot_m_2
  store i64 0, ptr %slot_n_3
  br label %for.cond2
for.cond2:
  %t3 = load i8, ptr %slot_m_2
  %t4 = icmp ne i8 %t3, 0
  br i1 %t4, label %for.body4, label %for.end1
for.body4:
  %t5 = load i64, ptr %slot_n_3
  %t6 = add i64 %t5, 1
  store i64 %t6, ptr %slot_n_3
  %t7 = load i8, ptr %slot_m_2
  %t8 = load i8, ptr %slot_m_2
  %t9 = sub i8 %t8, 1
  %t10 = and i8 %t7, %t9
  store i8 %t10, ptr %slot_m_2
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t11 = load i64, ptr %slot_n_3
  ret i64 %t11
}

define i64 @aila_main__hash_fnv64({ ptr, i64 } %p_bytes, i64 %p_n) {
  %slot_bytes_1 = alloca { ptr, i64 }, align 8
  %slot_n_2 = alloca i64, align 8
  %slot_h_3 = alloca i64, align 8
  %slot_i_4 = alloca i64, align 8
  store { ptr, i64 } %p_bytes, ptr %slot_bytes_1
  store i64 %p_n, ptr %slot_n_2
  store i64 -3750763034362895579, ptr %slot_h_3
  store i64 0, ptr %slot_i_4
  br label %for.cond2
for.cond2:
  %t1 = load i64, ptr %slot_i_4
  %t2 = load i64, ptr %slot_n_2
  %t3 = icmp ult i64 %t1, %t2
  br i1 %t3, label %for.body4, label %for.end1
for.body4:
  %t4 = load i64, ptr %slot_h_3
  %t5 = load i64, ptr %slot_i_4
  %t6 = load { ptr, i64 }, ptr %slot_bytes_1
  %t7 = extractvalue { ptr, i64 } %t6, 0
  %t8 = extractvalue { ptr, i64 } %t6, 1
  %t9 = icmp uge i64 %t5, %t8
  br i1 %t9, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.5, i64 19, ptr @.str.3, i64 105)
  unreachable
chk.ok5:
  %t10 = getelementptr inbounds i8, ptr %t7, i64 %t5
  %t11 = load i8, ptr %t10
  %t12 = zext i8 %t11 to i64
  %t13 = xor i64 %t4, %t12
  store i64 %t13, ptr %slot_h_3
  %t14 = load i64, ptr %slot_h_3
  %t15 = mul i64 %t14, 1099511628211
  store i64 %t15, ptr %slot_h_3
  %t16 = load i64, ptr %slot_i_4
  %t17 = add i64 %t16, 1
  store i64 %t17, ptr %slot_i_4
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t18 = load i64, ptr %slot_h_3
  ret i64 %t18
}

define void @aila_main__main() {
  %slot_s_mix_1 = alloca { ptr, i64 }, align 8
  %slot_s_l1_2 = alloca { ptr, i64 }, align 8
  %slot_s_l2_3 = alloca { ptr, i64 }, align 8
  %slot_s_l3_4 = alloca { ptr, i64 }, align 8
  %slot_s_cjk_5 = alloca { ptr, i64 }, align 8
  %slot_s_long_6 = alloca { ptr, i64 }, align 8
  %slot_b_mix_7 = alloca { ptr, i64 }, align 8
  %slot_b_l1_8 = alloca { ptr, i64 }, align 8
  %slot_b_l2_9 = alloca { ptr, i64 }, align 8
  %slot_b_l3_10 = alloca { ptr, i64 }, align 8
  %slot_b_cjk_11 = alloca { ptr, i64 }, align 8
  %slot_b_long_12 = alloca { ptr, i64 }, align 8
  %slot_fbdata_13 = alloca { ptr, i64 }, align 8
  %slot_grid_buf_14 = alloca { ptr, i64 }, align 8
  %slot_dmg_buf_15 = alloca { ptr, i64 }, align 8
  %slot_pb_16 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_gs_17 = alloca { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, align 8
  %slot_g_18 = alloca { { ptr, i64 } }, align 8
  %slot_d_19 = alloca { { ptr, i64 } }, align 8
  %slot_c_20 = alloca { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, align 8
  %slot_aw_21 = alloca i16, align 8
  %slot_cw_22 = alloca i16, align 8
  %slot_ew_23 = alloca i16, align 8
  %slot_endcol_24 = alloca i16, align 8
  %slot_cu1_25 = alloca i16, align 8
  %slot_cu2_26 = alloca i16, align 8
  %slot_cu3_27 = alloca i16, align 8
  %slot_cu4_28 = alloca i16, align 8
  %slot_pm_29 = alloca i8, align 8
  %slot_vargs_30 = alloca [7 x { i8, { { ptr, i64 } } }], align 8
  %slot_vargs_31 = alloca [4 x { i8, { { ptr, i64 } } }], align 8
  %slot___destr_32 = alloca { i16, i16 }, align 8
  %slot_vargs_33 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot___destr_34 = alloca { i16, i16 }, align 8
  %slot_vargs_35 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot_vargs_36 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot_vargs_37 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot_vargs_38 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot_vargs_39 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot_vargs_40 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot_vargs_41 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot_vargs_42 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot_vargs_43 = alloca [3 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_44 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_45 = alloca [4 x { i8, { { ptr, i64 } } }], align 8
  %t1 = insertvalue { ptr, i64 } undef, ptr @.str.6, 0
  %t2 = insertvalue { ptr, i64 } %t1, i64 11, 1
  store { ptr, i64 } %t2, ptr %slot_s_mix_1
  %t3 = insertvalue { ptr, i64 } undef, ptr @.str.7, 0
  %t4 = insertvalue { ptr, i64 } %t3, i64 5, 1
  store { ptr, i64 } %t4, ptr %slot_s_l1_2
  %t5 = insertvalue { ptr, i64 } undef, ptr @.str.8, 0
  %t6 = insertvalue { ptr, i64 } %t5, i64 5, 1
  store { ptr, i64 } %t6, ptr %slot_s_l2_3
  %t7 = insertvalue { ptr, i64 } undef, ptr @.str.9, 0
  %t8 = insertvalue { ptr, i64 } %t7, i64 5, 1
  store { ptr, i64 } %t8, ptr %slot_s_l3_4
  %t9 = insertvalue { ptr, i64 } undef, ptr @.str.10, 0
  %t10 = insertvalue { ptr, i64 } %t9, i64 12, 1
  store { ptr, i64 } %t10, ptr %slot_s_cjk_5
  %t11 = insertvalue { ptr, i64 } undef, ptr @.str.11, 0
  %t12 = insertvalue { ptr, i64 } %t11, i64 20, 1
  store { ptr, i64 } %t12, ptr %slot_s_long_6
  %t13 = load { ptr, i64 }, ptr %slot_s_mix_1
  store { ptr, i64 } %t13, ptr %slot_b_mix_7
  %t14 = load { ptr, i64 }, ptr %slot_s_l1_2
  store { ptr, i64 } %t14, ptr %slot_b_l1_8
  %t15 = load { ptr, i64 }, ptr %slot_s_l2_3
  store { ptr, i64 } %t15, ptr %slot_b_l2_9
  %t16 = load { ptr, i64 }, ptr %slot_s_l3_4
  store { ptr, i64 } %t16, ptr %slot_b_l3_10
  %t17 = load { ptr, i64 }, ptr %slot_s_cjk_5
  store { ptr, i64 } %t17, ptr %slot_b_cjk_11
  %t18 = load { ptr, i64 }, ptr %slot_s_long_6
  store { ptr, i64 } %t18, ptr %slot_b_long_12
  %t19 = call ptr @aila_rt_region_enter(i64 0)
  %t20 = mul i64 1024, 1
  %t21 = call ptr @aila_rt_region_alloc(ptr %t19, i64 %t20, i64 1)
  %t22 = insertvalue { ptr, i64 } undef, ptr %t21, 0
  %t23 = insertvalue { ptr, i64 } %t22, i64 1024, 1
  store { ptr, i64 } %t23, ptr %slot_fbdata_13
  %t24 = call i64 @aila_std_textgrid__words(i16 16, i16 4)
  %t25 = mul i64 %t24, 4
  %t26 = call ptr @aila_rt_region_alloc(ptr %t19, i64 %t25, i64 4)
  %t27 = insertvalue { ptr, i64 } undef, ptr %t26, 0
  %t28 = insertvalue { ptr, i64 } %t27, i64 %t24, 1
  store { ptr, i64 } %t28, ptr %slot_grid_buf_14
  %t29 = mul i64 21, 4
  %t30 = call ptr @aila_rt_region_alloc(ptr %t19, i64 %t29, i64 4)
  %t31 = insertvalue { ptr, i64 } undef, ptr %t30, 0
  %t32 = insertvalue { ptr, i64 } %t31, i64 21, 1
  store { ptr, i64 } %t32, ptr %slot_dmg_buf_15
  %t33 = load { ptr, i64 }, ptr %slot_fbdata_13
  %t34 = call { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } @aila_std_pixbuf__new({ ptr, i64 } %t33, i16 128, i16 64)
  store { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } %t34, ptr %slot_pb_16
  %t35 = call { ptr, i64 } @aila_font_font16__font_cps()
  %t36 = call { ptr, i64 } @aila_font_font16__font_bits()
  %t37 = call { ptr, i64 } @aila_font_font16__font_widths()
  %t38 = call i64 @aila_font_font16__font_count()
  %t39 = call i16 @aila_font_font16__font_px_h()
  %t40 = call { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 } @aila_std_paint__new_glyph_set({ ptr, i64 } %t35, { ptr, i64 } %t36, { ptr, i64 } %t37, i64 %t38, i16 %t39)
  store { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 } %t40, ptr %slot_gs_17
  %t41 = load { ptr, i64 }, ptr %slot_grid_buf_14
  %t42 = call { { ptr, i64 } } @aila_std_textgrid__new({ ptr, i64 } %t41, i16 16, i16 4)
  store { { ptr, i64 } } %t42, ptr %slot_g_18
  %t43 = load { ptr, i64 }, ptr %slot_dmg_buf_15
  %t44 = call { { ptr, i64 } } @aila_std_damage__new({ ptr, i64 } %t43)
  store { { ptr, i64 } } %t44, ptr %slot_d_19
  %t45 = load { { ptr, i64 } }, ptr %slot_d_19
  %t46 = call { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 } @aila_std_display__new_commit(ptr %slot_pb_16, { { ptr, i64 } } %t45)
  store { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 } %t46, ptr %slot_c_20
  %t47 = insertvalue { ptr, i64 } undef, ptr @.str.12, 0
  %t48 = insertvalue { ptr, i64 } %t47, i64 54, 1
  %t49 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot_pb_16, i32 0, i32 1
  %t50 = load i16, ptr %t49
  %t51 = zext i16 %t50 to i64
  %t52 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t51)
  %t53 = getelementptr inbounds { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %slot_pb_16, i32 0, i32 2
  %t54 = load i16, ptr %t53
  %t55 = zext i16 %t54 to i64
  %t56 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t55)
  %t57 = call i64 @aila_std_pixbuf__row_bytes(ptr %slot_pb_16)
  %t58 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t57)
  %t59 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 1024)
  %t60 = load { { ptr, i64 } }, ptr %slot_g_18
  %t61 = call i16 @aila_std_textgrid__cols_of({ { ptr, i64 } } %t60)
  %t62 = zext i16 %t61 to i64
  %t63 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t62)
  %t64 = load { { ptr, i64 } }, ptr %slot_g_18
  %t65 = call i16 @aila_std_textgrid__rows_of({ { ptr, i64 } } %t64)
  %t66 = zext i16 %t65 to i64
  %t67 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t66)
  %t68 = getelementptr inbounds { { ptr, i64 }, { ptr, i64 }, { ptr, i64 }, i64, i16 }, ptr %slot_gs_17, i32 0, i32 3
  %t69 = load i64, ptr %t68
  %t70 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t69)
  %t71 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_30, i64 0
  store { i8, { { ptr, i64 } } } %t52, ptr %t71
  %t72 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_30, i64 1
  store { i8, { { ptr, i64 } } } %t56, ptr %t72
  %t73 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_30, i64 2
  store { i8, { { ptr, i64 } } } %t58, ptr %t73
  %t74 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_30, i64 3
  store { i8, { { ptr, i64 } } } %t59, ptr %t74
  %t75 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_30, i64 4
  store { i8, { { ptr, i64 } } } %t63, ptr %t75
  %t76 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_30, i64 5
  store { i8, { { ptr, i64 } } } %t67, ptr %t76
  %t77 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_30, i64 6
  store { i8, { { ptr, i64 } } } %t70, ptr %t77
  %t78 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_30, 0
  %t79 = insertvalue { ptr, i64 } %t78, i64 7, 1
  call void @aila_std_io__printf({ ptr, i64 } %t48, { ptr, i64 } %t79)
  %t80 = call i16 @aila_std_paint__glyph_width(ptr %slot_gs_17, i32 65)
  store i16 %t80, ptr %slot_aw_21
  %t81 = call i16 @aila_std_paint__glyph_width(ptr %slot_gs_17, i32 28201)
  store i16 %t81, ptr %slot_cw_22
  %t82 = call i16 @aila_std_paint__glyph_width(ptr %slot_gs_17, i32 19968)
  store i16 %t82, ptr %slot_ew_23
  call void @aila_std_paint__fill(ptr %slot_pb_16, i1 0)
  %t83 = load { ptr, i64 }, ptr %slot_b_mix_7
  %t84 = load { ptr, i64 }, ptr %slot_b_mix_7
  %t85 = extractvalue { ptr, i64 } %t84, 1
  %t86 = call i16 @aila_std_paint__text_cells(ptr %slot_pb_16, ptr %slot_gs_17, { ptr, i64 } %t83, i64 %t85, i16 0, i16 0)
  store i16 %t86, ptr %slot_endcol_24
  %t87 = insertvalue { ptr, i64 } undef, ptr @.str.13, 0
  %t88 = insertvalue { ptr, i64 } %t87, i64 83, 1
  %t89 = load i16, ptr %slot_aw_21
  %t90 = zext i16 %t89 to i64
  %t91 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t90)
  %t92 = load i16, ptr %slot_cw_22
  %t93 = zext i16 %t92 to i64
  %t94 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t93)
  %t95 = load i16, ptr %slot_ew_23
  %t96 = zext i16 %t95 to i64
  %t97 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t96)
  %t98 = load i16, ptr %slot_endcol_24
  %t99 = zext i16 %t98 to i64
  %t100 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t99)
  %t101 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_31, i64 0
  store { i8, { { ptr, i64 } } } %t91, ptr %t101
  %t102 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_31, i64 1
  store { i8, { { ptr, i64 } } } %t94, ptr %t102
  %t103 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_31, i64 2
  store { i8, { { ptr, i64 } } } %t97, ptr %t103
  %t104 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_31, i64 3
  store { i8, { { ptr, i64 } } } %t100, ptr %t104
  %t105 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_31, 0
  %t106 = insertvalue { ptr, i64 } %t105, i64 4, 1
  call void @aila_std_io__printf({ ptr, i64 } %t88, { ptr, i64 } %t106)
  %t107 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__clear({ { ptr, i64 } } %t107)
  %t108 = load { { ptr, i64 } }, ptr %slot_g_18
  %t109 = load { ptr, i64 }, ptr %slot_b_cjk_11
  %t110 = load { ptr, i64 }, ptr %slot_b_cjk_11
  %t111 = extractvalue { ptr, i64 } %t110, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t108, ptr %slot_gs_17, { ptr, i64 } %t109, i64 %t111)
  %t112 = load { { ptr, i64 } }, ptr %slot_g_18
  %t113 = load { ptr, i64 }, ptr %slot_b_long_12
  %t114 = load { ptr, i64 }, ptr %slot_b_long_12
  %t115 = extractvalue { ptr, i64 } %t114, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t112, ptr %slot_gs_17, { ptr, i64 } %t113, i64 %t115)
  %t116 = load { { ptr, i64 } }, ptr %slot_g_18
  %t117 = call { i16, i16 } @aila_std_textgrid__cursor({ { ptr, i64 } } %t116)
  store { i16, i16 } %t117, ptr %slot___destr_32
  %t118 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_32, i32 0, i32 0
  %t119 = load i16, ptr %t118
  store i16 %t119, ptr %slot_cu1_25
  %t120 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_32, i32 0, i32 1
  %t121 = load i16, ptr %t120
  store i16 %t121, ptr %slot_cu2_26
  %t122 = insertvalue { ptr, i64 } undef, ptr @.str.14, 0
  %t123 = insertvalue { ptr, i64 } %t122, i64 26, 1
  %t124 = load i16, ptr %slot_cu1_25
  %t125 = zext i16 %t124 to i64
  %t126 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t125)
  %t127 = load i16, ptr %slot_cu2_26
  %t128 = zext i16 %t127 to i64
  %t129 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t128)
  %t130 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_33, i64 0
  store { i8, { { ptr, i64 } } } %t126, ptr %t130
  %t131 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_33, i64 1
  store { i8, { { ptr, i64 } } } %t129, ptr %t131
  %t132 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_33, 0
  %t133 = insertvalue { ptr, i64 } %t132, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t123, { ptr, i64 } %t133)
  %t134 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_main__dump_grid({ { ptr, i64 } } %t134)
  %t135 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__clear({ { ptr, i64 } } %t135)
  %t136 = load { { ptr, i64 } }, ptr %slot_g_18
  %t137 = load { ptr, i64 }, ptr %slot_b_long_12
  %t138 = load { ptr, i64 }, ptr %slot_b_long_12
  %t139 = extractvalue { ptr, i64 } %t138, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t136, ptr %slot_gs_17, { ptr, i64 } %t137, i64 %t139)
  %t140 = load { { ptr, i64 } }, ptr %slot_g_18
  %t141 = call { i16, i16 } @aila_std_textgrid__cursor({ { ptr, i64 } } %t140)
  store { i16, i16 } %t141, ptr %slot___destr_34
  %t142 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_34, i32 0, i32 0
  %t143 = load i16, ptr %t142
  store i16 %t143, ptr %slot_cu3_27
  %t144 = getelementptr inbounds { i16, i16 }, ptr %slot___destr_34, i32 0, i32 1
  %t145 = load i16, ptr %t144
  store i16 %t145, ptr %slot_cu4_28
  %t146 = insertvalue { ptr, i64 } undef, ptr @.str.15, 0
  %t147 = insertvalue { ptr, i64 } %t146, i64 30, 1
  %t148 = load i16, ptr %slot_cu3_27
  %t149 = zext i16 %t148 to i64
  %t150 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t149)
  %t151 = load i16, ptr %slot_cu4_28
  %t152 = zext i16 %t151 to i64
  %t153 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t152)
  %t154 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_35, i64 0
  store { i8, { { ptr, i64 } } } %t150, ptr %t154
  %t155 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_35, i64 1
  store { i8, { { ptr, i64 } } } %t153, ptr %t155
  %t156 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_35, 0
  %t157 = insertvalue { ptr, i64 } %t156, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t147, { ptr, i64 } %t157)
  %t158 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_main__dump_grid({ { ptr, i64 } } %t158)
  %t159 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__clear({ { ptr, i64 } } %t159)
  %t160 = load { { ptr, i64 } }, ptr %slot_g_18
  %t161 = load { ptr, i64 }, ptr %slot_b_l1_8
  %t162 = load { ptr, i64 }, ptr %slot_b_l1_8
  %t163 = extractvalue { ptr, i64 } %t162, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t160, ptr %slot_gs_17, { ptr, i64 } %t161, i64 %t163)
  %t164 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__newline({ { ptr, i64 } } %t164)
  %t165 = load { { ptr, i64 } }, ptr %slot_g_18
  %t166 = load { ptr, i64 }, ptr %slot_b_l2_9
  %t167 = load { ptr, i64 }, ptr %slot_b_l2_9
  %t168 = extractvalue { ptr, i64 } %t167, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t165, ptr %slot_gs_17, { ptr, i64 } %t166, i64 %t168)
  %t169 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__newline({ { ptr, i64 } } %t169)
  %t170 = load { { ptr, i64 } }, ptr %slot_g_18
  %t171 = load { ptr, i64 }, ptr %slot_b_l3_10
  %t172 = load { ptr, i64 }, ptr %slot_b_l3_10
  %t173 = extractvalue { ptr, i64 } %t172, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t170, ptr %slot_gs_17, { ptr, i64 } %t171, i64 %t173)
  %t174 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_damage__clear({ { ptr, i64 } } %t174)
  %t175 = load { { ptr, i64 } }, ptr %slot_g_18
  %t176 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_textgrid__render_all({ { ptr, i64 } } %t175, ptr %slot_gs_17, ptr %slot_pb_16, { { ptr, i64 } } %t176)
  %t177 = insertvalue { ptr, i64 } undef, ptr @.str.16, 0
  %t178 = insertvalue { ptr, i64 } %t177, i64 37, 1
  %t179 = load { { ptr, i64 } }, ptr %slot_d_19
  %t180 = call i64 @aila_std_damage__cell_count({ { ptr, i64 } } %t179)
  %t181 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t180)
  %t182 = load { { ptr, i64 } }, ptr %slot_d_19
  %t183 = call i64 @aila_main__pages_of({ { ptr, i64 } } %t182)
  %t184 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t183)
  %t185 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_36, i64 0
  store { i8, { { ptr, i64 } } } %t181, ptr %t185
  %t186 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_36, i64 1
  store { i8, { { ptr, i64 } } } %t184, ptr %t186
  %t187 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_36, 0
  %t188 = insertvalue { ptr, i64 } %t187, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t178, { ptr, i64 } %t188)
  %t189 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_damage__clear({ { ptr, i64 } } %t189)
  %t190 = load { { ptr, i64 } }, ptr %slot_g_18
  %t191 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_textgrid__render({ { ptr, i64 } } %t190, ptr %slot_gs_17, ptr %slot_pb_16, { { ptr, i64 } } %t191)
  %t192 = insertvalue { ptr, i64 } undef, ptr @.str.17, 0
  %t193 = insertvalue { ptr, i64 } %t192, i64 37, 1
  %t194 = load { { ptr, i64 } }, ptr %slot_d_19
  %t195 = call i64 @aila_std_damage__cell_count({ { ptr, i64 } } %t194)
  %t196 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t195)
  %t197 = load { { ptr, i64 } }, ptr %slot_d_19
  %t198 = call i64 @aila_main__pages_of({ { ptr, i64 } } %t197)
  %t199 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t198)
  %t200 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_37, i64 0
  store { i8, { { ptr, i64 } } } %t196, ptr %t200
  %t201 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_37, i64 1
  store { i8, { { ptr, i64 } } } %t199, ptr %t201
  %t202 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_37, 0
  %t203 = insertvalue { ptr, i64 } %t202, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t193, { ptr, i64 } %t203)
  %t204 = load { { ptr, i64 } }, ptr %slot_g_18
  %t205 = call i32 @aila_std_textgrid__encode(i32 33, i32 1)
  call void @aila_std_textgrid__set({ { ptr, i64 } } %t204, i16 15, i16 1, i32 %t205)
  %t206 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_damage__clear({ { ptr, i64 } } %t206)
  %t207 = load { { ptr, i64 } }, ptr %slot_g_18
  %t208 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_textgrid__render({ { ptr, i64 } } %t207, ptr %slot_gs_17, ptr %slot_pb_16, { { ptr, i64 } } %t208)
  %t209 = insertvalue { ptr, i64 } undef, ptr @.str.18, 0
  %t210 = insertvalue { ptr, i64 } %t209, i64 37, 1
  %t211 = load { { ptr, i64 } }, ptr %slot_d_19
  %t212 = call i64 @aila_std_damage__cell_count({ { ptr, i64 } } %t211)
  %t213 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t212)
  %t214 = load { { ptr, i64 } }, ptr %slot_d_19
  %t215 = call i64 @aila_main__pages_of({ { ptr, i64 } } %t214)
  %t216 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t215)
  %t217 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_38, i64 0
  store { i8, { { ptr, i64 } } } %t213, ptr %t217
  %t218 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_38, i64 1
  store { i8, { { ptr, i64 } } } %t216, ptr %t218
  %t219 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_38, 0
  %t220 = insertvalue { ptr, i64 } %t219, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t210, { ptr, i64 } %t220)
  %t221 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__set({ { ptr, i64 } } %t221, i16 15, i16 1, i32 0)
  %t222 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__clear({ { ptr, i64 } } %t222)
  %t223 = load { { ptr, i64 } }, ptr %slot_g_18
  %t224 = load { ptr, i64 }, ptr %slot_b_l1_8
  %t225 = load { ptr, i64 }, ptr %slot_b_l1_8
  %t226 = extractvalue { ptr, i64 } %t225, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t223, ptr %slot_gs_17, { ptr, i64 } %t224, i64 %t226)
  %t227 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__newline({ { ptr, i64 } } %t227)
  %t228 = load { { ptr, i64 } }, ptr %slot_g_18
  %t229 = load { ptr, i64 }, ptr %slot_b_l2_9
  %t230 = load { ptr, i64 }, ptr %slot_b_l2_9
  %t231 = extractvalue { ptr, i64 } %t230, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t228, ptr %slot_gs_17, { ptr, i64 } %t229, i64 %t231)
  %t232 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__newline({ { ptr, i64 } } %t232)
  %t233 = load { { ptr, i64 } }, ptr %slot_g_18
  %t234 = load { ptr, i64 }, ptr %slot_b_l3_10
  %t235 = load { ptr, i64 }, ptr %slot_b_l3_10
  %t236 = extractvalue { ptr, i64 } %t235, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t233, ptr %slot_gs_17, { ptr, i64 } %t234, i64 %t236)
  %t237 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_damage__clear({ { ptr, i64 } } %t237)
  %t238 = load { { ptr, i64 } }, ptr %slot_g_18
  %t239 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_textgrid__render_all({ { ptr, i64 } } %t238, ptr %slot_gs_17, ptr %slot_pb_16, { { ptr, i64 } } %t239)
  %t240 = insertvalue { ptr, i64 } undef, ptr @.str.19, 0
  %t241 = insertvalue { ptr, i64 } %t240, i64 36, 1
  %t242 = load { { ptr, i64 } }, ptr %slot_d_19
  %t243 = call i64 @aila_std_damage__cell_count({ { ptr, i64 } } %t242)
  %t244 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t243)
  %t245 = load { { ptr, i64 } }, ptr %slot_d_19
  %t246 = call i64 @aila_main__pages_of({ { ptr, i64 } } %t245)
  %t247 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t246)
  %t248 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_39, i64 0
  store { i8, { { ptr, i64 } } } %t244, ptr %t248
  %t249 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_39, i64 1
  store { i8, { { ptr, i64 } } } %t247, ptr %t249
  %t250 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_39, 0
  %t251 = insertvalue { ptr, i64 } %t250, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t241, { ptr, i64 } %t251)
  %t252 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__scroll_by({ { ptr, i64 } } %t252, i16 1)
  %t253 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_damage__clear({ { ptr, i64 } } %t253)
  %t254 = load { { ptr, i64 } }, ptr %slot_g_18
  %t255 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_textgrid__render({ { ptr, i64 } } %t254, ptr %slot_gs_17, ptr %slot_pb_16, { { ptr, i64 } } %t255)
  %t256 = insertvalue { ptr, i64 } undef, ptr @.str.20, 0
  %t257 = insertvalue { ptr, i64 } %t256, i64 95, 1
  %t258 = load { { ptr, i64 } }, ptr %slot_d_19
  %t259 = call i64 @aila_std_damage__cell_count({ { ptr, i64 } } %t258)
  %t260 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t259)
  %t261 = load { { ptr, i64 } }, ptr %slot_d_19
  %t262 = call i64 @aila_main__pages_of({ { ptr, i64 } } %t261)
  %t263 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t262)
  %t264 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_40, i64 0
  store { i8, { { ptr, i64 } } } %t260, ptr %t264
  %t265 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_40, i64 1
  store { i8, { { ptr, i64 } } } %t263, ptr %t265
  %t266 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_40, 0
  %t267 = insertvalue { ptr, i64 } %t266, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t257, { ptr, i64 } %t267)
  %t268 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__clear({ { ptr, i64 } } %t268)
  %t269 = load { { ptr, i64 } }, ptr %slot_g_18
  %t270 = load { ptr, i64 }, ptr %slot_b_l1_8
  %t271 = load { ptr, i64 }, ptr %slot_b_l1_8
  %t272 = extractvalue { ptr, i64 } %t271, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t269, ptr %slot_gs_17, { ptr, i64 } %t270, i64 %t272)
  %t273 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__newline({ { ptr, i64 } } %t273)
  %t274 = load { { ptr, i64 } }, ptr %slot_g_18
  %t275 = load { ptr, i64 }, ptr %slot_b_l2_9
  %t276 = load { ptr, i64 }, ptr %slot_b_l2_9
  %t277 = extractvalue { ptr, i64 } %t276, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t274, ptr %slot_gs_17, { ptr, i64 } %t275, i64 %t277)
  %t278 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__newline({ { ptr, i64 } } %t278)
  %t279 = load { { ptr, i64 } }, ptr %slot_g_18
  %t280 = load { ptr, i64 }, ptr %slot_b_l3_10
  %t281 = load { ptr, i64 }, ptr %slot_b_l3_10
  %t282 = extractvalue { ptr, i64 } %t281, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t279, ptr %slot_gs_17, { ptr, i64 } %t280, i64 %t282)
  %t283 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__newline({ { ptr, i64 } } %t283)
  %t284 = load { { ptr, i64 } }, ptr %slot_g_18
  %t285 = load { ptr, i64 }, ptr %slot_b_l3_10
  %t286 = load { ptr, i64 }, ptr %slot_b_l3_10
  %t287 = extractvalue { ptr, i64 } %t286, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t284, ptr %slot_gs_17, { ptr, i64 } %t285, i64 %t287)
  %t288 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_damage__clear({ { ptr, i64 } } %t288)
  %t289 = load { { ptr, i64 } }, ptr %slot_g_18
  %t290 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_textgrid__render_all({ { ptr, i64 } } %t289, ptr %slot_gs_17, ptr %slot_pb_16, { { ptr, i64 } } %t290)
  %t291 = insertvalue { ptr, i64 } undef, ptr @.str.21, 0
  %t292 = insertvalue { ptr, i64 } %t291, i64 36, 1
  %t293 = load { { ptr, i64 } }, ptr %slot_d_19
  %t294 = call i64 @aila_std_damage__cell_count({ { ptr, i64 } } %t293)
  %t295 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t294)
  %t296 = load { { ptr, i64 } }, ptr %slot_d_19
  %t297 = call i64 @aila_main__pages_of({ { ptr, i64 } } %t296)
  %t298 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t297)
  %t299 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_41, i64 0
  store { i8, { { ptr, i64 } } } %t295, ptr %t299
  %t300 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_41, i64 1
  store { i8, { { ptr, i64 } } } %t298, ptr %t300
  %t301 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_41, 0
  %t302 = insertvalue { ptr, i64 } %t301, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t292, { ptr, i64 } %t302)
  %t303 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__scroll_by({ { ptr, i64 } } %t303, i16 1)
  %t304 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_damage__clear({ { ptr, i64 } } %t304)
  %t305 = load { { ptr, i64 } }, ptr %slot_g_18
  %t306 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_textgrid__render({ { ptr, i64 } } %t305, ptr %slot_gs_17, ptr %slot_pb_16, { { ptr, i64 } } %t306)
  %t307 = insertvalue { ptr, i64 } undef, ptr @.str.22, 0
  %t308 = insertvalue { ptr, i64 } %t307, i64 78, 1
  %t309 = load { { ptr, i64 } }, ptr %slot_d_19
  %t310 = call i64 @aila_std_damage__cell_count({ { ptr, i64 } } %t309)
  %t311 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t310)
  %t312 = load { { ptr, i64 } }, ptr %slot_d_19
  %t313 = call i64 @aila_main__pages_of({ { ptr, i64 } } %t312)
  %t314 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t313)
  %t315 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_42, i64 0
  store { i8, { { ptr, i64 } } } %t311, ptr %t315
  %t316 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_42, i64 1
  store { i8, { { ptr, i64 } } } %t314, ptr %t316
  %t317 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_42, 0
  %t318 = insertvalue { ptr, i64 } %t317, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t308, { ptr, i64 } %t318)
  %t319 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_display__Commit__set_damage(ptr %slot_c_20, { { ptr, i64 } } %t319)
  %t320 = call { { ptr, i64 } } @aila_std_display__Commit__damage_of(ptr %slot_c_20)
  %t321 = call i8 @aila_main__page_mask({ { ptr, i64 } } %t320)
  store i8 %t321, ptr %slot_pm_29
  %t322 = insertvalue { ptr, i64 } undef, ptr @.str.23, 0
  %t323 = insertvalue { ptr, i64 } %t322, i64 103, 1
  %t324 = load i8, ptr %slot_pm_29
  %t325 = zext i8 %t324 to i64
  %t326 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t325)
  %t327 = call { { ptr, i64 } } @aila_std_display__Commit__damage_of(ptr %slot_c_20)
  %t328 = call i64 @aila_main__pages_of({ { ptr, i64 } } %t327)
  %t329 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t328)
  %t330 = call { { ptr, i64 } } @aila_std_display__Commit__damage_of(ptr %slot_c_20)
  %t331 = call i64 @aila_main__pages_of({ { ptr, i64 } } %t330)
  %t332 = mul i64 %t331, 128
  %t333 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t332)
  %t334 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_43, i64 0
  store { i8, { { ptr, i64 } } } %t326, ptr %t334
  %t335 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_43, i64 1
  store { i8, { { ptr, i64 } } } %t329, ptr %t335
  %t336 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_43, i64 2
  store { i8, { { ptr, i64 } } } %t333, ptr %t336
  %t337 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_43, 0
  %t338 = insertvalue { ptr, i64 } %t337, i64 3, 1
  call void @aila_std_io__printf({ ptr, i64 } %t323, { ptr, i64 } %t338)
  %t339 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__clear({ { ptr, i64 } } %t339)
  call void @aila_std_paint__fill(ptr %slot_pb_16, i1 0)
  %t340 = load { { ptr, i64 } }, ptr %slot_g_18
  %t341 = load { ptr, i64 }, ptr %slot_b_mix_7
  %t342 = load { ptr, i64 }, ptr %slot_b_mix_7
  %t343 = extractvalue { ptr, i64 } %t342, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t340, ptr %slot_gs_17, { ptr, i64 } %t341, i64 %t343)
  %t344 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__newline({ { ptr, i64 } } %t344)
  %t345 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_std_textgrid__newline({ { ptr, i64 } } %t345)
  %t346 = load { { ptr, i64 } }, ptr %slot_g_18
  %t347 = load { ptr, i64 }, ptr %slot_b_cjk_11
  %t348 = load { ptr, i64 }, ptr %slot_b_cjk_11
  %t349 = extractvalue { ptr, i64 } %t348, 1
  call void @aila_std_textgrid__put_text({ { ptr, i64 } } %t346, ptr %slot_gs_17, { ptr, i64 } %t347, i64 %t349)
  %t350 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_damage__clear({ { ptr, i64 } } %t350)
  %t351 = load { { ptr, i64 } }, ptr %slot_g_18
  %t352 = load { { ptr, i64 } }, ptr %slot_d_19
  call void @aila_std_textgrid__render({ { ptr, i64 } } %t351, ptr %slot_gs_17, ptr %slot_pb_16, { { ptr, i64 } } %t352)
  call void @aila_std_paint__hline(ptr %slot_pb_16, i16 0, i16 127, i16 62, i1 1)
  %t353 = load { { ptr, i64 } }, ptr %slot_d_19
  %t354 = call { i16, i16, i16, i16 } @aila_std_damage__rect(i16 0, i16 62, i16 128, i16 63)
  call void @aila_std_damage__add_rect({ { ptr, i64 } } %t353, { i16, i16, i16, i16 } %t354)
  %t355 = insertvalue { ptr, i64 } undef, ptr @.str.24, 0
  %t356 = insertvalue { ptr, i64 } %t355, i64 47, 1
  %t357 = load { { ptr, i64 } }, ptr %slot_d_19
  %t358 = call i64 @aila_std_damage__cell_count({ { ptr, i64 } } %t357)
  %t359 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t358)
  %t360 = load { { ptr, i64 } }, ptr %slot_d_19
  %t361 = call i64 @aila_std_damage__clip_count({ { ptr, i64 } } %t360)
  %t362 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t361)
  %t363 = load { { ptr, i64 } }, ptr %slot_d_19
  %t364 = call i64 @aila_main__pages_of({ { ptr, i64 } } %t363)
  %t365 = call { i8, { { ptr, i64 } } } @aila_main__u(i64 %t364)
  %t366 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t366, ptr %slot___lit_44
  %t367 = load { ptr, i64 }, ptr %slot_fbdata_13
  %t368 = call i64 @aila_main__hash_fnv64({ ptr, i64 } %t367, i64 1024)
  %t369 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_44, i32 0, i32 1, i32 0
  store i64 %t368, ptr %t369
  %t370 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_44
  %t371 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_45, i64 0
  store { i8, { { ptr, i64 } } } %t359, ptr %t371
  %t372 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_45, i64 1
  store { i8, { { ptr, i64 } } } %t362, ptr %t372
  %t373 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_45, i64 2
  store { i8, { { ptr, i64 } } } %t365, ptr %t373
  %t374 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_45, i64 3
  store { i8, { { ptr, i64 } } } %t370, ptr %t374
  %t375 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_45, 0
  %t376 = insertvalue { ptr, i64 } %t375, i64 4, 1
  call void @aila_std_io__printf({ ptr, i64 } %t356, { ptr, i64 } %t376)
  %t377 = load { { ptr, i64 } }, ptr %slot_g_18
  call void @aila_main__dump_grid({ { ptr, i64 } } %t377)
  call void @aila_std_display_host__sink_host_ascii(ptr %slot_c_20)
  call void @aila_rt_region_leave(ptr %t19)
  ret void
}

define i32 @main(i32 %argc, ptr %argv) {
entry:
  call void @aila_main__main()
  ret i32 0
}
