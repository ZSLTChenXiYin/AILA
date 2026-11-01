; AILA module: std.damage
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.1 = private unnamed_addr constant [16 x i8] c"std/damage.aila\00", align 1
@.str.2 = private unnamed_addr constant [25 x i8] c"shift count out of range\00", align 1
@.str.3 = private unnamed_addr constant [17 x i8] c"division by zero\00", align 1


declare void @aila_rt_trap(ptr, i64, ptr, i64)

define { i16, i16, i16, i16 } @aila_std_damage__rect(i16 %p_x0, i16 %p_y0, i16 %p_x1, i16 %p_y1) {
  %slot_x0_1 = alloca i16, align 8
  %slot_y0_2 = alloca i16, align 8
  %slot_x1_3 = alloca i16, align 8
  %slot_y1_4 = alloca i16, align 8
  %slot___lit_5 = alloca { i16, i16, i16, i16 }, align 8
  store i16 %p_x0, ptr %slot_x0_1
  store i16 %p_y0, ptr %slot_y0_2
  store i16 %p_x1, ptr %slot_x1_3
  store i16 %p_y1, ptr %slot_y1_4
  %t1 = load i16, ptr %slot_x0_1
  %t2 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_5, i32 0, i32 0
  store i16 %t1, ptr %t2
  %t3 = load i16, ptr %slot_y0_2
  %t4 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_5, i32 0, i32 1
  store i16 %t3, ptr %t4
  %t5 = load i16, ptr %slot_x1_3
  %t6 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_5, i32 0, i32 2
  store i16 %t5, ptr %t6
  %t7 = load i16, ptr %slot_y1_4
  %t8 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_5, i32 0, i32 3
  store i16 %t7, ptr %t8
  %t9 = load { i16, i16, i16, i16 }, ptr %slot___lit_5
  ret { i16, i16, i16, i16 } %t9
}

define { i16, i16, i16, i16 } @aila_std_damage__rect_empty() {
  %slot___lit_1 = alloca { i16, i16, i16, i16 }, align 8
  %t1 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_1, i32 0, i32 0
  store i16 0, ptr %t1
  %t2 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_1, i32 0, i32 1
  store i16 0, ptr %t2
  %t3 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_1, i32 0, i32 2
  store i16 0, ptr %t3
  %t4 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_1, i32 0, i32 3
  store i16 0, ptr %t4
  %t5 = load { i16, i16, i16, i16 }, ptr %slot___lit_1
  ret { i16, i16, i16, i16 } %t5
}

define i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 } %p_self) {
  %slot_self_1 = alloca { i16, i16, i16, i16 }, align 8
  store { i16, i16, i16, i16 } %p_self, ptr %slot_self_1
  %t1 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 2
  %t2 = load i16, ptr %t1
  %t3 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 0
  %t4 = load i16, ptr %t3
  %t5 = icmp ule i16 %t2, %t4
  br i1 %t5, label %if.then2, label %if.next3
if.then2:
  ret i1 1
if.next3:
  br label %if.end1
if.end1:
  %t6 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 3
  %t7 = load i16, ptr %t6
  %t8 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 1
  %t9 = load i16, ptr %t8
  %t10 = icmp ule i16 %t7, %t9
  br i1 %t10, label %if.then5, label %if.next6
if.then5:
  ret i1 1
if.next6:
  br label %if.end4
if.end4:
  ret i1 0
}

define i16 @aila_std_damage__Rect__w({ i16, i16, i16, i16 } %p_self) {
  %slot_self_1 = alloca { i16, i16, i16, i16 }, align 8
  store { i16, i16, i16, i16 } %p_self, ptr %slot_self_1
  %t1 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 2
  %t2 = load i16, ptr %t1
  %t3 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 0
  %t4 = load i16, ptr %t3
  %t5 = icmp ule i16 %t2, %t4
  br i1 %t5, label %if.then2, label %if.next3
if.then2:
  ret i16 0
if.next3:
  br label %if.end1
if.end1:
  %t6 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 2
  %t7 = load i16, ptr %t6
  %t8 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 0
  %t9 = load i16, ptr %t8
  %t10 = sub i16 %t7, %t9
  ret i16 %t10
}

define i16 @aila_std_damage__Rect__h({ i16, i16, i16, i16 } %p_self) {
  %slot_self_1 = alloca { i16, i16, i16, i16 }, align 8
  store { i16, i16, i16, i16 } %p_self, ptr %slot_self_1
  %t1 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 3
  %t2 = load i16, ptr %t1
  %t3 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 1
  %t4 = load i16, ptr %t3
  %t5 = icmp ule i16 %t2, %t4
  br i1 %t5, label %if.then2, label %if.next3
if.then2:
  ret i16 0
if.next3:
  br label %if.end1
if.end1:
  %t6 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 3
  %t7 = load i16, ptr %t6
  %t8 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 1
  %t9 = load i16, ptr %t8
  %t10 = sub i16 %t7, %t9
  ret i16 %t10
}

define i1 @aila_std_damage__Rect__touches({ i16, i16, i16, i16 } %p_self, { i16, i16, i16, i16 } %p_o, i16 %p_gap) {
  %slot_self_1 = alloca { i16, i16, i16, i16 }, align 8
  %slot_o_2 = alloca { i16, i16, i16, i16 }, align 8
  %slot_gap_3 = alloca i16, align 8
  store { i16, i16, i16, i16 } %p_self, ptr %slot_self_1
  store { i16, i16, i16, i16 } %p_o, ptr %slot_o_2
  store i16 %p_gap, ptr %slot_gap_3
  %t1 = load { i16, i16, i16, i16 }, ptr %slot_self_1
  %t2 = call i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 } %t1)
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  ret i1 0
if.next3:
  br label %if.end1
if.end1:
  %t3 = load { i16, i16, i16, i16 }, ptr %slot_o_2
  %t4 = call i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 } %t3)
  br i1 %t4, label %if.then5, label %if.next6
if.then5:
  ret i1 0
if.next6:
  br label %if.end4
if.end4:
  %t5 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 0
  %t6 = load i16, ptr %t5
  %t7 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_o_2, i32 0, i32 2
  %t8 = load i16, ptr %t7
  %t9 = load i16, ptr %slot_gap_3
  %t10 = add i16 %t8, %t9
  %t11 = icmp ugt i16 %t6, %t10
  br i1 %t11, label %if.then8, label %if.next9
if.then8:
  ret i1 0
if.next9:
  br label %if.end7
if.end7:
  %t12 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_o_2, i32 0, i32 0
  %t13 = load i16, ptr %t12
  %t14 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 2
  %t15 = load i16, ptr %t14
  %t16 = load i16, ptr %slot_gap_3
  %t17 = add i16 %t15, %t16
  %t18 = icmp ugt i16 %t13, %t17
  br i1 %t18, label %if.then11, label %if.next12
if.then11:
  ret i1 0
if.next12:
  br label %if.end10
if.end10:
  %t19 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 1
  %t20 = load i16, ptr %t19
  %t21 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_o_2, i32 0, i32 3
  %t22 = load i16, ptr %t21
  %t23 = load i16, ptr %slot_gap_3
  %t24 = add i16 %t22, %t23
  %t25 = icmp ugt i16 %t20, %t24
  br i1 %t25, label %if.then14, label %if.next15
if.then14:
  ret i1 0
if.next15:
  br label %if.end13
if.end13:
  %t26 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_o_2, i32 0, i32 1
  %t27 = load i16, ptr %t26
  %t28 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_self_1, i32 0, i32 3
  %t29 = load i16, ptr %t28
  %t30 = load i16, ptr %slot_gap_3
  %t31 = add i16 %t29, %t30
  %t32 = icmp ugt i16 %t27, %t31
  br i1 %t32, label %if.then17, label %if.next18
if.then17:
  ret i1 0
if.next18:
  br label %if.end16
if.end16:
  ret i1 1
}

define { i16, i16, i16, i16 } @aila_std_damage__Rect__union({ i16, i16, i16, i16 } %p_self, { i16, i16, i16, i16 } %p_o) {
  %slot_self_1 = alloca { i16, i16, i16, i16 }, align 8
  %slot_o_2 = alloca { i16, i16, i16, i16 }, align 8
  %slot_r_3 = alloca { i16, i16, i16, i16 }, align 8
  store { i16, i16, i16, i16 } %p_self, ptr %slot_self_1
  store { i16, i16, i16, i16 } %p_o, ptr %slot_o_2
  %t1 = load { i16, i16, i16, i16 }, ptr %slot_self_1
  %t2 = call i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 } %t1)
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  %t3 = load { i16, i16, i16, i16 }, ptr %slot_o_2
  ret { i16, i16, i16, i16 } %t3
if.next3:
  br label %if.end1
if.end1:
  %t4 = load { i16, i16, i16, i16 }, ptr %slot_o_2
  %t5 = call i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 } %t4)
  br i1 %t5, label %if.then5, label %if.next6
if.then5:
  %t6 = load { i16, i16, i16, i16 }, ptr %slot_self_1
  ret { i16, i16, i16, i16 } %t6
if.next6:
  br label %if.end4
if.end4:
  %t7 = load { i16, i16, i16, i16 }, ptr %slot_self_1
  store { i16, i16, i16, i16 } %t7, ptr %slot_r_3
  %t8 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_o_2, i32 0, i32 0
  %t9 = load i16, ptr %t8
  %t10 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_r_3, i32 0, i32 0
  %t11 = load i16, ptr %t10
  %t12 = icmp ult i16 %t9, %t11
  br i1 %t12, label %if.then8, label %if.next9
if.then8:
  %t13 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_r_3, i32 0, i32 0
  %t14 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_o_2, i32 0, i32 0
  %t15 = load i16, ptr %t14
  store i16 %t15, ptr %t13
  br label %if.end7
if.next9:
  br label %if.end7
if.end7:
  %t16 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_o_2, i32 0, i32 1
  %t17 = load i16, ptr %t16
  %t18 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_r_3, i32 0, i32 1
  %t19 = load i16, ptr %t18
  %t20 = icmp ult i16 %t17, %t19
  br i1 %t20, label %if.then11, label %if.next12
if.then11:
  %t21 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_r_3, i32 0, i32 1
  %t22 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_o_2, i32 0, i32 1
  %t23 = load i16, ptr %t22
  store i16 %t23, ptr %t21
  br label %if.end10
if.next12:
  br label %if.end10
if.end10:
  %t24 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_o_2, i32 0, i32 2
  %t25 = load i16, ptr %t24
  %t26 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_r_3, i32 0, i32 2
  %t27 = load i16, ptr %t26
  %t28 = icmp ugt i16 %t25, %t27
  br i1 %t28, label %if.then14, label %if.next15
if.then14:
  %t29 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_r_3, i32 0, i32 2
  %t30 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_o_2, i32 0, i32 2
  %t31 = load i16, ptr %t30
  store i16 %t31, ptr %t29
  br label %if.end13
if.next15:
  br label %if.end13
if.end13:
  %t32 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_o_2, i32 0, i32 3
  %t33 = load i16, ptr %t32
  %t34 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_r_3, i32 0, i32 3
  %t35 = load i16, ptr %t34
  %t36 = icmp ugt i16 %t33, %t35
  br i1 %t36, label %if.then17, label %if.next18
if.then17:
  %t37 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_r_3, i32 0, i32 3
  %t38 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_o_2, i32 0, i32 3
  %t39 = load i16, ptr %t38
  store i16 %t39, ptr %t37
  br label %if.end16
if.next18:
  br label %if.end16
if.end16:
  %t40 = load { i16, i16, i16, i16 }, ptr %slot_r_3
  ret { i16, i16, i16, i16 } %t40
}

define { i16, i16, i16, i16 } @aila_std_damage__Rect__clamp({ i16, i16, i16, i16 } %p_self, i16 %p_w, i16 %p_h) {
  %slot_self_1 = alloca { i16, i16, i16, i16 }, align 8
  %slot_w_2 = alloca i16, align 8
  %slot_h_3 = alloca i16, align 8
  %slot_c_4 = alloca { i16, i16, i16, i16 }, align 8
  store { i16, i16, i16, i16 } %p_self, ptr %slot_self_1
  store i16 %p_w, ptr %slot_w_2
  store i16 %p_h, ptr %slot_h_3
  %t1 = load { i16, i16, i16, i16 }, ptr %slot_self_1
  store { i16, i16, i16, i16 } %t1, ptr %slot_c_4
  %t2 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 0
  %t3 = load i16, ptr %t2
  %t4 = load i16, ptr %slot_w_2
  %t5 = icmp ugt i16 %t3, %t4
  br i1 %t5, label %if.then2, label %if.next3
if.then2:
  %t6 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 0
  %t7 = load i16, ptr %slot_w_2
  store i16 %t7, ptr %t6
  br label %if.end1
if.next3:
  br label %if.end1
if.end1:
  %t8 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 1
  %t9 = load i16, ptr %t8
  %t10 = load i16, ptr %slot_h_3
  %t11 = icmp ugt i16 %t9, %t10
  br i1 %t11, label %if.then5, label %if.next6
if.then5:
  %t12 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 1
  %t13 = load i16, ptr %slot_h_3
  store i16 %t13, ptr %t12
  br label %if.end4
if.next6:
  br label %if.end4
if.end4:
  %t14 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 2
  %t15 = load i16, ptr %t14
  %t16 = load i16, ptr %slot_w_2
  %t17 = icmp ugt i16 %t15, %t16
  br i1 %t17, label %if.then8, label %if.next9
if.then8:
  %t18 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 2
  %t19 = load i16, ptr %slot_w_2
  store i16 %t19, ptr %t18
  br label %if.end7
if.next9:
  br label %if.end7
if.end7:
  %t20 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 3
  %t21 = load i16, ptr %t20
  %t22 = load i16, ptr %slot_h_3
  %t23 = icmp ugt i16 %t21, %t22
  br i1 %t23, label %if.then11, label %if.next12
if.then11:
  %t24 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 3
  %t25 = load i16, ptr %slot_h_3
  store i16 %t25, ptr %t24
  br label %if.end10
if.next12:
  br label %if.end10
if.end10:
  %t26 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 2
  %t27 = load i16, ptr %t26
  %t28 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 0
  %t29 = load i16, ptr %t28
  %t30 = icmp ult i16 %t27, %t29
  br i1 %t30, label %if.then14, label %if.next15
if.then14:
  %t31 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 2
  %t32 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 0
  %t33 = load i16, ptr %t32
  store i16 %t33, ptr %t31
  br label %if.end13
if.next15:
  br label %if.end13
if.end13:
  %t34 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 3
  %t35 = load i16, ptr %t34
  %t36 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 1
  %t37 = load i16, ptr %t36
  %t38 = icmp ult i16 %t35, %t37
  br i1 %t38, label %if.then17, label %if.next18
if.then17:
  %t39 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 3
  %t40 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_c_4, i32 0, i32 1
  %t41 = load i16, ptr %t40
  store i16 %t41, ptr %t39
  br label %if.end16
if.next18:
  br label %if.end16
if.end16:
  %t42 = load { i16, i16, i16, i16 }, ptr %slot_c_4
  ret { i16, i16, i16, i16 } %t42
}

define { { ptr, i64 } } @aila_std_damage__new({ ptr, i64 } %p_raw) {
  %slot_raw_1 = alloca { ptr, i64 }, align 8
  %slot_d_2 = alloca { { ptr, i64 } }, align 8
  %slot___lit_3 = alloca { { ptr, i64 } }, align 8
  store { ptr, i64 } %p_raw, ptr %slot_raw_1
  %t1 = load { ptr, i64 }, ptr %slot_raw_1
  %t2 = getelementptr inbounds { { ptr, i64 } }, ptr %slot___lit_3, i32 0, i32 0
  store { ptr, i64 } %t1, ptr %t2
  %t3 = load { { ptr, i64 } }, ptr %slot___lit_3
  store { { ptr, i64 } } %t3, ptr %slot_d_2
  %t4 = load { { ptr, i64 } }, ptr %slot_d_2
  call void @aila_std_damage__clear({ { ptr, i64 } } %t4)
  %t5 = load { { ptr, i64 } }, ptr %slot_d_2
  ret { { ptr, i64 } } %t5
}

define void @aila_std_damage__clear({ { ptr, i64 } } %p_d) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_i_2 = alloca i64, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  store i64 0, ptr %slot_i_2
  br label %for.cond2
for.cond2:
  %t1 = load i64, ptr %slot_i_2
  %t2 = icmp ult i64 %t1, 21
  br i1 %t2, label %for.body4, label %for.end1
for.body4:
  %t3 = load i64, ptr %slot_i_2
  %t4 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t5 = load { ptr, i64 }, ptr %t4
  %t6 = extractvalue { ptr, i64 } %t5, 1
  %t8 = icmp ult i64 %t3, %t6
  br i1 %t8, label %if.then6, label %if.next7
if.then6:
  %t9 = load i64, ptr %slot_i_2
  %t10 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t11 = load { ptr, i64 }, ptr %t10
  %t12 = extractvalue { ptr, i64 } %t11, 0
  %t13 = extractvalue { ptr, i64 } %t11, 1
  %t14 = icmp uge i64 %t9, %t13
  br i1 %t14, label %chk.bad9, label %chk.ok8
chk.bad9:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 219)
  unreachable
chk.ok8:
  %t15 = getelementptr inbounds i32, ptr %t12, i64 %t9
  store i32 0, ptr %t15
  br label %if.end5
if.next7:
  br label %if.end5
if.end5:
  %t16 = load i64, ptr %slot_i_2
  %t17 = add i64 %t16, 1
  store i64 %t17, ptr %slot_i_2
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  ret void
}

define { i16, i16, i16, i16 } @aila_std_damage__clip_at({ { ptr, i64 } } %p_d, i64 %p_i) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_i_2 = alloca i64, align 8
  %slot_a_3 = alloca i32, align 8
  %slot_b_4 = alloca i32, align 8
  %slot___lit_5 = alloca { i16, i16, i16, i16 }, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  store i64 %p_i, ptr %slot_i_2
  %t1 = load i64, ptr %slot_i_2
  %t2 = mul i64 %t1, 2
  %t3 = add i64 1, %t2
  %t4 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t5 = load { ptr, i64 }, ptr %t4
  %t6 = extractvalue { ptr, i64 } %t5, 0
  %t7 = extractvalue { ptr, i64 } %t5, 1
  %t8 = icmp uge i64 %t3, %t7
  br i1 %t8, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 227)
  unreachable
chk.ok1:
  %t9 = getelementptr inbounds i32, ptr %t6, i64 %t3
  %t10 = load i32, ptr %t9
  store i32 %t10, ptr %slot_a_3
  %t11 = load i64, ptr %slot_i_2
  %t12 = mul i64 %t11, 2
  %t13 = add i64 2, %t12
  %t14 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t15 = load { ptr, i64 }, ptr %t14
  %t16 = extractvalue { ptr, i64 } %t15, 0
  %t17 = extractvalue { ptr, i64 } %t15, 1
  %t18 = icmp uge i64 %t13, %t17
  br i1 %t18, label %chk.bad4, label %chk.ok3
chk.bad4:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 228)
  unreachable
chk.ok3:
  %t19 = getelementptr inbounds i32, ptr %t16, i64 %t13
  %t20 = load i32, ptr %t19
  store i32 %t20, ptr %slot_b_4
  %t21 = load i32, ptr %slot_a_3
  %t22 = and i32 %t21, 65535
  %t23 = trunc i32 %t22 to i16
  %t24 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_5, i32 0, i32 0
  store i16 %t23, ptr %t24
  %t25 = load i32, ptr %slot_b_4
  %t26 = and i32 %t25, 65535
  %t27 = trunc i32 %t26 to i16
  %t28 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_5, i32 0, i32 1
  store i16 %t27, ptr %t28
  %t29 = load i32, ptr %slot_a_3
  %t30 = icmp uge i32 16, 32
  br i1 %t30, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 231)
  unreachable
chk.ok5:
  %t31 = lshr i32 %t29, 16
  %t32 = trunc i32 %t31 to i16
  %t33 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_5, i32 0, i32 2
  store i16 %t32, ptr %t33
  %t34 = load i32, ptr %slot_b_4
  %t35 = icmp uge i32 16, 32
  br i1 %t35, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 233)
  unreachable
chk.ok7:
  %t36 = lshr i32 %t34, 16
  %t37 = trunc i32 %t36 to i16
  %t38 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_5, i32 0, i32 3
  store i16 %t37, ptr %t38
  %t39 = load { i16, i16, i16, i16 }, ptr %slot___lit_5
  ret { i16, i16, i16, i16 } %t39
}

define void @aila_std_damage__set_clip({ { ptr, i64 } } %p_d, i64 %p_i, { i16, i16, i16, i16 } %p_r) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_i_2 = alloca i64, align 8
  %slot_r_3 = alloca { i16, i16, i16, i16 }, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  store i64 %p_i, ptr %slot_i_2
  store { i16, i16, i16, i16 } %p_r, ptr %slot_r_3
  %t1 = load i64, ptr %slot_i_2
  %t2 = mul i64 %t1, 2
  %t3 = add i64 1, %t2
  %t4 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t5 = load { ptr, i64 }, ptr %t4
  %t6 = extractvalue { ptr, i64 } %t5, 0
  %t7 = extractvalue { ptr, i64 } %t5, 1
  %t8 = icmp uge i64 %t3, %t7
  br i1 %t8, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 239)
  unreachable
chk.ok1:
  %t9 = getelementptr inbounds i32, ptr %t6, i64 %t3
  %t10 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_r_3, i32 0, i32 0
  %t11 = load i16, ptr %t10
  %t12 = zext i16 %t11 to i32
  %t13 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_r_3, i32 0, i32 2
  %t14 = load i16, ptr %t13
  %t15 = zext i16 %t14 to i32
  %t16 = icmp uge i32 16, 32
  br i1 %t16, label %chk.bad4, label %chk.ok3
chk.bad4:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 239)
  unreachable
chk.ok3:
  %t17 = shl i32 %t15, 16
  %t18 = or i32 %t12, %t17
  store i32 %t18, ptr %t9
  %t19 = load i64, ptr %slot_i_2
  %t20 = mul i64 %t19, 2
  %t21 = add i64 2, %t20
  %t22 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t23 = load { ptr, i64 }, ptr %t22
  %t24 = extractvalue { ptr, i64 } %t23, 0
  %t25 = extractvalue { ptr, i64 } %t23, 1
  %t26 = icmp uge i64 %t21, %t25
  br i1 %t26, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 240)
  unreachable
chk.ok5:
  %t27 = getelementptr inbounds i32, ptr %t24, i64 %t21
  %t28 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_r_3, i32 0, i32 1
  %t29 = load i16, ptr %t28
  %t30 = zext i16 %t29 to i32
  %t31 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot_r_3, i32 0, i32 3
  %t32 = load i16, ptr %t31
  %t33 = zext i16 %t32 to i32
  %t34 = icmp uge i32 16, 32
  br i1 %t34, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 240)
  unreachable
chk.ok7:
  %t35 = shl i32 %t33, 16
  %t36 = or i32 %t30, %t35
  store i32 %t36, ptr %t27
  ret void
}

define { i16, i16, i16, i16 } @aila_std_damage__clip({ { ptr, i64 } } %p_d, i64 %p_i) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_i_2 = alloca i64, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  store i64 %p_i, ptr %slot_i_2
  %t1 = load i64, ptr %slot_i_2
  %t2 = icmp uge i64 %t1, 8
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  %t3 = call { i16, i16, i16, i16 } @aila_std_damage__rect_empty()
  ret { i16, i16, i16, i16 } %t3
if.next3:
  br label %if.end1
if.end1:
  %t4 = load { { ptr, i64 } }, ptr %slot_d_1
  %t5 = load i64, ptr %slot_i_2
  %t6 = call { i16, i16, i16, i16 } @aila_std_damage__clip_at({ { ptr, i64 } } %t4, i64 %t5)
  ret { i16, i16, i16, i16 } %t6
}

define i64 @aila_std_damage__clip_count({ { ptr, i64 } } %p_d) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_n_2 = alloca i64, align 8
  %slot_i_3 = alloca i64, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  store i64 0, ptr %slot_n_2
  store i64 0, ptr %slot_i_3
  br label %for.cond2
for.cond2:
  %t1 = load i64, ptr %slot_i_3
  %t2 = icmp ult i64 %t1, 8
  br i1 %t2, label %for.body4, label %for.end1
for.body4:
  %t3 = load { { ptr, i64 } }, ptr %slot_d_1
  %t4 = load i64, ptr %slot_i_3
  %t5 = call { i16, i16, i16, i16 } @aila_std_damage__clip_at({ { ptr, i64 } } %t3, i64 %t4)
  %t6 = call i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 } %t5)
  %t7 = xor i1 %t6, true
  br i1 %t7, label %if.then6, label %if.next7
if.then6:
  %t8 = load i64, ptr %slot_n_2
  %t9 = add i64 %t8, 1
  store i64 %t9, ptr %slot_n_2
  br label %if.end5
if.next7:
  br label %if.end5
if.end5:
  %t10 = load i64, ptr %slot_i_3
  %t11 = add i64 %t10, 1
  store i64 %t11, ptr %slot_i_3
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t12 = load i64, ptr %slot_n_2
  ret i64 %t12
}

define i1 @aila_std_damage__is_empty({ { ptr, i64 } } %p_d) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_k_2 = alloca i64, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  %t1 = load { { ptr, i64 } }, ptr %slot_d_1
  %t2 = call i64 @aila_std_damage__clip_count({ { ptr, i64 } } %t1)
  %t3 = icmp ne i64 %t2, 0
  br i1 %t3, label %if.then2, label %if.next3
if.then2:
  ret i1 0
if.next3:
  br label %if.end1
if.end1:
  store i64 0, ptr %slot_k_2
  br label %for.cond5
for.cond5:
  %t4 = load i64, ptr %slot_k_2
  %t5 = icmp ult i64 %t4, 4
  br i1 %t5, label %for.body7, label %for.end4
for.body7:
  %t6 = load i64, ptr %slot_k_2
  %t7 = add i64 17, %t6
  %t8 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t9 = load { ptr, i64 }, ptr %t8
  %t10 = extractvalue { ptr, i64 } %t9, 0
  %t11 = extractvalue { ptr, i64 } %t9, 1
  %t12 = icmp uge i64 %t7, %t11
  br i1 %t12, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 268)
  unreachable
chk.ok11:
  %t13 = getelementptr inbounds i32, ptr %t10, i64 %t7
  %t14 = load i32, ptr %t13
  %t15 = icmp ne i32 %t14, 0
  br i1 %t15, label %if.then9, label %if.next10
if.then9:
  ret i1 0
if.next10:
  br label %if.end8
if.end8:
  %t16 = load i64, ptr %slot_k_2
  %t17 = add i64 %t16, 1
  store i64 %t17, ptr %slot_k_2
  br label %for.step6
for.step6:
  br label %for.cond5
for.end4:
  ret i1 1
}

define void @aila_std_damage__add_rect({ { ptr, i64 } } %p_d, { i16, i16, i16, i16 } %p_r) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_r_2 = alloca { i16, i16, i16, i16 }, align 8
  %slot_i_3 = alloca i64, align 8
  %slot_cur_4 = alloca { i16, i16, i16, i16 }, align 8
  %slot_k_5 = alloca i64, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  store { i16, i16, i16, i16 } %p_r, ptr %slot_r_2
  %t1 = load { i16, i16, i16, i16 }, ptr %slot_r_2
  %t2 = call i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 } %t1)
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  ret void
if.next3:
  br label %if.end1
if.end1:
  store i64 0, ptr %slot_i_3
  br label %for.cond5
for.cond5:
  %t3 = load i64, ptr %slot_i_3
  %t4 = icmp ult i64 %t3, 8
  br i1 %t4, label %for.body7, label %for.end4
for.body7:
  %t5 = load { { ptr, i64 } }, ptr %slot_d_1
  %t6 = load i64, ptr %slot_i_3
  %t7 = call { i16, i16, i16, i16 } @aila_std_damage__clip_at({ { ptr, i64 } } %t5, i64 %t6)
  store { i16, i16, i16, i16 } %t7, ptr %slot_cur_4
  %t8 = load { i16, i16, i16, i16 }, ptr %slot_cur_4
  %t9 = call i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 } %t8)
  %t10 = xor i1 %t9, true
  br i1 %t10, label %if.then9, label %if.next10
if.then9:
  %t11 = load { i16, i16, i16, i16 }, ptr %slot_cur_4
  %t12 = load { i16, i16, i16, i16 }, ptr %slot_r_2
  %t13 = call i1 @aila_std_damage__Rect__touches({ i16, i16, i16, i16 } %t11, { i16, i16, i16, i16 } %t12, i16 2)
  br i1 %t13, label %if.then12, label %if.next13
if.then12:
  %t14 = load { { ptr, i64 } }, ptr %slot_d_1
  %t15 = load i64, ptr %slot_i_3
  %t16 = load { i16, i16, i16, i16 }, ptr %slot_cur_4
  %t17 = load { i16, i16, i16, i16 }, ptr %slot_r_2
  %t18 = call { i16, i16, i16, i16 } @aila_std_damage__Rect__union({ i16, i16, i16, i16 } %t16, { i16, i16, i16, i16 } %t17)
  call void @aila_std_damage__set_clip({ { ptr, i64 } } %t14, i64 %t15, { i16, i16, i16, i16 } %t18)
  ret void
if.next13:
  br label %if.end11
if.end11:
  br label %if.end8
if.next10:
  br label %if.end8
if.end8:
  %t19 = load i64, ptr %slot_i_3
  %t20 = add i64 %t19, 1
  store i64 %t20, ptr %slot_i_3
  br label %for.step6
for.step6:
  br label %for.cond5
for.end4:
  store i64 0, ptr %slot_k_5
  br label %for.cond15
for.cond15:
  %t21 = load i64, ptr %slot_k_5
  %t22 = icmp ult i64 %t21, 8
  br i1 %t22, label %for.body17, label %for.end14
for.body17:
  %t23 = load { { ptr, i64 } }, ptr %slot_d_1
  %t24 = load i64, ptr %slot_k_5
  %t25 = call { i16, i16, i16, i16 } @aila_std_damage__clip_at({ { ptr, i64 } } %t23, i64 %t24)
  %t26 = call i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 } %t25)
  br i1 %t26, label %if.then19, label %if.next20
if.then19:
  %t27 = load { { ptr, i64 } }, ptr %slot_d_1
  %t28 = load i64, ptr %slot_k_5
  %t29 = load { i16, i16, i16, i16 }, ptr %slot_r_2
  call void @aila_std_damage__set_clip({ { ptr, i64 } } %t27, i64 %t28, { i16, i16, i16, i16 } %t29)
  %t30 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t31 = load { ptr, i64 }, ptr %t30
  %t32 = extractvalue { ptr, i64 } %t31, 0
  %t33 = extractvalue { ptr, i64 } %t31, 1
  %t34 = icmp uge i64 0, %t33
  br i1 %t34, label %chk.bad25, label %chk.ok24
chk.bad25:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 297)
  unreachable
chk.ok24:
  %t35 = getelementptr inbounds i32, ptr %t32, i64 0
  %t36 = load i32, ptr %t35
  %t37 = trunc i64 8 to i32
  %t38 = icmp ult i32 %t36, %t37
  br i1 %t38, label %if.then22, label %if.next23
if.then22:
  %t39 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t40 = load { ptr, i64 }, ptr %t39
  %t41 = extractvalue { ptr, i64 } %t40, 0
  %t42 = extractvalue { ptr, i64 } %t40, 1
  %t43 = icmp uge i64 0, %t42
  br i1 %t43, label %chk.bad27, label %chk.ok26
chk.bad27:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 298)
  unreachable
chk.ok26:
  %t44 = getelementptr inbounds i32, ptr %t41, i64 0
  %t45 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t46 = load { ptr, i64 }, ptr %t45
  %t47 = extractvalue { ptr, i64 } %t46, 0
  %t48 = extractvalue { ptr, i64 } %t46, 1
  %t49 = icmp uge i64 0, %t48
  br i1 %t49, label %chk.bad29, label %chk.ok28
chk.bad29:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 298)
  unreachable
chk.ok28:
  %t50 = getelementptr inbounds i32, ptr %t47, i64 0
  %t51 = load i32, ptr %t50
  %t52 = add i32 %t51, 1
  store i32 %t52, ptr %t44
  br label %if.end21
if.next23:
  br label %if.end21
if.end21:
  ret void
if.next20:
  br label %if.end18
if.end18:
  %t53 = load i64, ptr %slot_k_5
  %t54 = add i64 %t53, 1
  store i64 %t54, ptr %slot_k_5
  br label %for.step16
for.step16:
  br label %for.cond15
for.end14:
  %t55 = load { { ptr, i64 } }, ptr %slot_d_1
  %t56 = load { { ptr, i64 } }, ptr %slot_d_1
  %t57 = call { i16, i16, i16, i16 } @aila_std_damage__clip_at({ { ptr, i64 } } %t56, i64 0)
  %t58 = load { i16, i16, i16, i16 }, ptr %slot_r_2
  %t59 = call { i16, i16, i16, i16 } @aila_std_damage__Rect__union({ i16, i16, i16, i16 } %t57, { i16, i16, i16, i16 } %t58)
  call void @aila_std_damage__set_clip({ { ptr, i64 } } %t55, i64 0, { i16, i16, i16, i16 } %t59)
  ret void
}

define void @aila_std_damage__add_cell({ { ptr, i64 } } %p_d, i16 %p_col, i16 %p_row) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_col_2 = alloca i16, align 8
  %slot_row_3 = alloca i16, align 8
  %slot_bit_4 = alloca i64, align 8
  %slot_word_5 = alloca i64, align 8
  %slot_mask_6 = alloca i32, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  store i16 %p_col, ptr %slot_col_2
  store i16 %p_row, ptr %slot_row_3
  %t1 = load i16, ptr %slot_col_2
  %t2 = icmp uge i16 %t1, 16
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  ret void
if.next3:
  br label %if.end1
if.end1:
  %t3 = load i16, ptr %slot_row_3
  %t4 = icmp uge i16 %t3, 8
  br i1 %t4, label %if.then5, label %if.next6
if.then5:
  ret void
if.next6:
  br label %if.end4
if.end4:
  %t5 = load i16, ptr %slot_row_3
  %t6 = zext i16 %t5 to i64
  %t7 = zext i16 16 to i64
  %t8 = mul i64 %t6, %t7
  %t9 = load i16, ptr %slot_col_2
  %t10 = zext i16 %t9 to i64
  %t11 = add i64 %t8, %t10
  store i64 %t11, ptr %slot_bit_4
  %t12 = load i64, ptr %slot_bit_4
  %t13 = icmp eq i64 32, 0
  br i1 %t13, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 316)
  unreachable
chk.ok7:
  %t14 = udiv i64 %t12, 32
  %t15 = add i64 17, %t14
  store i64 %t15, ptr %slot_word_5
  %t16 = trunc i64 1 to i32
  %t17 = load i64, ptr %slot_bit_4
  %t18 = icmp eq i64 32, 0
  br i1 %t18, label %chk.bad10, label %chk.ok9
chk.bad10:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 317)
  unreachable
chk.ok9:
  %t19 = urem i64 %t17, 32
  %t20 = trunc i64 %t19 to i32
  %t21 = icmp uge i32 %t20, 32
  br i1 %t21, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 317)
  unreachable
chk.ok11:
  %t22 = shl i32 %t16, %t20
  store i32 %t22, ptr %slot_mask_6
  %t23 = load i64, ptr %slot_word_5
  %t24 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t25 = load { ptr, i64 }, ptr %t24
  %t26 = extractvalue { ptr, i64 } %t25, 0
  %t27 = extractvalue { ptr, i64 } %t25, 1
  %t28 = icmp uge i64 %t23, %t27
  br i1 %t28, label %chk.bad14, label %chk.ok13
chk.bad14:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 318)
  unreachable
chk.ok13:
  %t29 = getelementptr inbounds i32, ptr %t26, i64 %t23
  %t30 = load i64, ptr %slot_word_5
  %t31 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t32 = load { ptr, i64 }, ptr %t31
  %t33 = extractvalue { ptr, i64 } %t32, 0
  %t34 = extractvalue { ptr, i64 } %t32, 1
  %t35 = icmp uge i64 %t30, %t34
  br i1 %t35, label %chk.bad16, label %chk.ok15
chk.bad16:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 318)
  unreachable
chk.ok15:
  %t36 = getelementptr inbounds i32, ptr %t33, i64 %t30
  %t37 = load i32, ptr %t36
  %t38 = load i32, ptr %slot_mask_6
  %t39 = or i32 %t37, %t38
  store i32 %t39, ptr %t29
  ret void
}

define i1 @aila_std_damage__has_cell({ { ptr, i64 } } %p_d, i16 %p_col, i16 %p_row) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_col_2 = alloca i16, align 8
  %slot_row_3 = alloca i16, align 8
  %slot_bit_4 = alloca i64, align 8
  %slot_word_5 = alloca i64, align 8
  %slot_mask_6 = alloca i32, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  store i16 %p_col, ptr %slot_col_2
  store i16 %p_row, ptr %slot_row_3
  %t1 = load i16, ptr %slot_col_2
  %t2 = icmp uge i16 %t1, 16
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  ret i1 0
if.next3:
  br label %if.end1
if.end1:
  %t3 = load i16, ptr %slot_row_3
  %t4 = icmp uge i16 %t3, 8
  br i1 %t4, label %if.then5, label %if.next6
if.then5:
  ret i1 0
if.next6:
  br label %if.end4
if.end4:
  %t5 = load i16, ptr %slot_row_3
  %t6 = zext i16 %t5 to i64
  %t7 = zext i16 16 to i64
  %t8 = mul i64 %t6, %t7
  %t9 = load i16, ptr %slot_col_2
  %t10 = zext i16 %t9 to i64
  %t11 = add i64 %t8, %t10
  store i64 %t11, ptr %slot_bit_4
  %t12 = load i64, ptr %slot_bit_4
  %t13 = icmp eq i64 32, 0
  br i1 %t13, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 329)
  unreachable
chk.ok7:
  %t14 = udiv i64 %t12, 32
  %t15 = add i64 17, %t14
  store i64 %t15, ptr %slot_word_5
  %t16 = trunc i64 1 to i32
  %t17 = load i64, ptr %slot_bit_4
  %t18 = icmp eq i64 32, 0
  br i1 %t18, label %chk.bad10, label %chk.ok9
chk.bad10:
  call void @aila_rt_trap(ptr @.str.3, i64 16, ptr @.str.1, i64 330)
  unreachable
chk.ok9:
  %t19 = urem i64 %t17, 32
  %t20 = trunc i64 %t19 to i32
  %t21 = icmp uge i32 %t20, 32
  br i1 %t21, label %chk.bad12, label %chk.ok11
chk.bad12:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 330)
  unreachable
chk.ok11:
  %t22 = shl i32 %t16, %t20
  store i32 %t22, ptr %slot_mask_6
  %t23 = load i64, ptr %slot_word_5
  %t24 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t25 = load { ptr, i64 }, ptr %t24
  %t26 = extractvalue { ptr, i64 } %t25, 0
  %t27 = extractvalue { ptr, i64 } %t25, 1
  %t28 = icmp uge i64 %t23, %t27
  br i1 %t28, label %chk.bad14, label %chk.ok13
chk.bad14:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 331)
  unreachable
chk.ok13:
  %t29 = getelementptr inbounds i32, ptr %t26, i64 %t23
  %t30 = load i32, ptr %t29
  %t31 = load i32, ptr %slot_mask_6
  %t32 = and i32 %t30, %t31
  %t33 = icmp ne i32 %t32, 0
  ret i1 %t33
}

define i64 @aila_std_damage__cell_count({ { ptr, i64 } } %p_d) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_total_2 = alloca i64, align 8
  %slot_k_3 = alloca i64, align 8
  %slot_b_4 = alloca i32, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  store i64 0, ptr %slot_total_2
  store i64 0, ptr %slot_k_3
  br label %for.cond2
for.cond2:
  %t1 = load i64, ptr %slot_k_3
  %t2 = icmp ult i64 %t1, 4
  br i1 %t2, label %for.body4, label %for.end1
for.body4:
  %t3 = load i64, ptr %slot_k_3
  %t4 = add i64 17, %t3
  %t5 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t6 = load { ptr, i64 }, ptr %t5
  %t7 = extractvalue { ptr, i64 } %t6, 0
  %t8 = extractvalue { ptr, i64 } %t6, 1
  %t9 = icmp uge i64 %t4, %t8
  br i1 %t9, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 338)
  unreachable
chk.ok5:
  %t10 = getelementptr inbounds i32, ptr %t7, i64 %t4
  %t11 = load i32, ptr %t10
  store i32 %t11, ptr %slot_b_4
  br label %for.cond8
for.cond8:
  %t12 = load i32, ptr %slot_b_4
  %t13 = icmp ne i32 %t12, 0
  br i1 %t13, label %for.body10, label %for.end7
for.body10:
  %t14 = load i64, ptr %slot_total_2
  %t15 = add i64 %t14, 1
  store i64 %t15, ptr %slot_total_2
  %t16 = load i32, ptr %slot_b_4
  %t17 = load i32, ptr %slot_b_4
  %t18 = sub i32 %t17, 1
  %t19 = and i32 %t16, %t18
  store i32 %t19, ptr %slot_b_4
  br label %for.step9
for.step9:
  br label %for.cond8
for.end7:
  %t20 = load i64, ptr %slot_k_3
  %t21 = add i64 %t20, 1
  store i64 %t21, ptr %slot_k_3
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t22 = load i64, ptr %slot_total_2
  ret i64 %t22
}

define void @aila_std_damage__union_into({ { ptr, i64 } } %p_d, { { ptr, i64 } } %p_o) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_o_2 = alloca { { ptr, i64 } }, align 8
  %slot_i_3 = alloca i64, align 8
  %slot_r_4 = alloca { i16, i16, i16, i16 }, align 8
  %slot_k_5 = alloca i64, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  store { { ptr, i64 } } %p_o, ptr %slot_o_2
  store i64 0, ptr %slot_i_3
  br label %for.cond2
for.cond2:
  %t1 = load i64, ptr %slot_i_3
  %t2 = icmp ult i64 %t1, 8
  br i1 %t2, label %for.body4, label %for.end1
for.body4:
  %t3 = load { { ptr, i64 } }, ptr %slot_o_2
  %t4 = load i64, ptr %slot_i_3
  %t5 = call { i16, i16, i16, i16 } @aila_std_damage__clip_at({ { ptr, i64 } } %t3, i64 %t4)
  store { i16, i16, i16, i16 } %t5, ptr %slot_r_4
  %t6 = load { i16, i16, i16, i16 }, ptr %slot_r_4
  %t7 = call i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 } %t6)
  %t8 = xor i1 %t7, true
  br i1 %t8, label %if.then6, label %if.next7
if.then6:
  %t9 = load { { ptr, i64 } }, ptr %slot_d_1
  %t10 = load { i16, i16, i16, i16 }, ptr %slot_r_4
  call void @aila_std_damage__add_rect({ { ptr, i64 } } %t9, { i16, i16, i16, i16 } %t10)
  br label %if.end5
if.next7:
  br label %if.end5
if.end5:
  %t11 = load i64, ptr %slot_i_3
  %t12 = add i64 %t11, 1
  store i64 %t12, ptr %slot_i_3
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  store i64 0, ptr %slot_k_5
  br label %for.cond9
for.cond9:
  %t13 = load i64, ptr %slot_k_5
  %t14 = icmp ult i64 %t13, 4
  br i1 %t14, label %for.body11, label %for.end8
for.body11:
  %t15 = load i64, ptr %slot_k_5
  %t16 = add i64 17, %t15
  %t17 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t18 = load { ptr, i64 }, ptr %t17
  %t19 = extractvalue { ptr, i64 } %t18, 0
  %t20 = extractvalue { ptr, i64 } %t18, 1
  %t21 = icmp uge i64 %t16, %t20
  br i1 %t21, label %chk.bad13, label %chk.ok12
chk.bad13:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 359)
  unreachable
chk.ok12:
  %t22 = getelementptr inbounds i32, ptr %t19, i64 %t16
  %t23 = load i64, ptr %slot_k_5
  %t24 = add i64 17, %t23
  %t25 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_d_1, i32 0, i32 0
  %t26 = load { ptr, i64 }, ptr %t25
  %t27 = extractvalue { ptr, i64 } %t26, 0
  %t28 = extractvalue { ptr, i64 } %t26, 1
  %t29 = icmp uge i64 %t24, %t28
  br i1 %t29, label %chk.bad15, label %chk.ok14
chk.bad15:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 359)
  unreachable
chk.ok14:
  %t30 = getelementptr inbounds i32, ptr %t27, i64 %t24
  %t31 = load i32, ptr %t30
  %t32 = load i64, ptr %slot_k_5
  %t33 = add i64 17, %t32
  %t34 = getelementptr inbounds { { ptr, i64 } }, ptr %slot_o_2, i32 0, i32 0
  %t35 = load { ptr, i64 }, ptr %t34
  %t36 = extractvalue { ptr, i64 } %t35, 0
  %t37 = extractvalue { ptr, i64 } %t35, 1
  %t38 = icmp uge i64 %t33, %t37
  br i1 %t38, label %chk.bad17, label %chk.ok16
chk.bad17:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 359)
  unreachable
chk.ok16:
  %t39 = getelementptr inbounds i32, ptr %t36, i64 %t33
  %t40 = load i32, ptr %t39
  %t41 = or i32 %t31, %t40
  store i32 %t41, ptr %t22
  %t42 = load i64, ptr %slot_k_5
  %t43 = add i64 %t42, 1
  store i64 %t43, ptr %slot_k_5
  br label %for.step10
for.step10:
  br label %for.cond9
for.end8:
  ret void
}

define { i16, i16, i16, i16 } @aila_std_damage__cells_bounds({ { ptr, i64 } } %p_d) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_r_2 = alloca { i16, i16, i16, i16 }, align 8
  %slot_cw_3 = alloca i16, align 8
  %slot_ch_4 = alloca i16, align 8
  %slot_row_5 = alloca i16, align 8
  %slot_col_6 = alloca i16, align 8
  %slot___lit_7 = alloca { i16, i16, i16, i16 }, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  %t1 = call { i16, i16, i16, i16 } @aila_std_damage__rect_empty()
  store { i16, i16, i16, i16 } %t1, ptr %slot_r_2
  %t2 = zext i8 8 to i16
  store i16 %t2, ptr %slot_cw_3
  %t3 = zext i8 16 to i16
  store i16 %t3, ptr %slot_ch_4
  store i16 0, ptr %slot_row_5
  br label %for.cond2
for.cond2:
  %t4 = load i16, ptr %slot_row_5
  %t5 = icmp ult i16 %t4, 8
  br i1 %t5, label %for.body4, label %for.end1
for.body4:
  store i16 0, ptr %slot_col_6
  br label %for.cond6
for.cond6:
  %t6 = load i16, ptr %slot_col_6
  %t7 = icmp ult i16 %t6, 16
  br i1 %t7, label %for.body8, label %for.end5
for.body8:
  %t8 = load { { ptr, i64 } }, ptr %slot_d_1
  %t9 = load i16, ptr %slot_col_6
  %t10 = load i16, ptr %slot_row_5
  %t11 = call i1 @aila_std_damage__has_cell({ { ptr, i64 } } %t8, i16 %t9, i16 %t10)
  br i1 %t11, label %if.then10, label %if.next11
if.then10:
  %t12 = load { i16, i16, i16, i16 }, ptr %slot_r_2
  %t13 = load i16, ptr %slot_col_6
  %t14 = load i16, ptr %slot_cw_3
  %t15 = mul i16 %t13, %t14
  %t16 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_7, i32 0, i32 0
  store i16 %t15, ptr %t16
  %t17 = load i16, ptr %slot_row_5
  %t18 = load i16, ptr %slot_ch_4
  %t19 = mul i16 %t17, %t18
  %t20 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_7, i32 0, i32 1
  store i16 %t19, ptr %t20
  %t21 = load i16, ptr %slot_col_6
  %t22 = load i16, ptr %slot_cw_3
  %t23 = mul i16 %t21, %t22
  %t24 = load i16, ptr %slot_cw_3
  %t25 = add i16 %t23, %t24
  %t26 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_7, i32 0, i32 2
  store i16 %t25, ptr %t26
  %t27 = load i16, ptr %slot_row_5
  %t28 = load i16, ptr %slot_ch_4
  %t29 = mul i16 %t27, %t28
  %t30 = load i16, ptr %slot_ch_4
  %t31 = add i16 %t29, %t30
  %t32 = getelementptr inbounds { i16, i16, i16, i16 }, ptr %slot___lit_7, i32 0, i32 3
  store i16 %t31, ptr %t32
  %t33 = load { i16, i16, i16, i16 }, ptr %slot___lit_7
  %t34 = call { i16, i16, i16, i16 } @aila_std_damage__Rect__union({ i16, i16, i16, i16 } %t12, { i16, i16, i16, i16 } %t33)
  store { i16, i16, i16, i16 } %t34, ptr %slot_r_2
  br label %if.end9
if.next11:
  br label %if.end9
if.end9:
  %t35 = load i16, ptr %slot_col_6
  %t36 = add i16 %t35, 1
  store i16 %t36, ptr %slot_col_6
  br label %for.step7
for.step7:
  br label %for.cond6
for.end5:
  %t37 = load i16, ptr %slot_row_5
  %t38 = add i16 %t37, 1
  store i16 %t38, ptr %slot_row_5
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t39 = load { i16, i16, i16, i16 }, ptr %slot_r_2
  ret { i16, i16, i16, i16 } %t39
}

define { i16, i16, i16, i16 } @aila_std_damage__bounds({ { ptr, i64 } } %p_d) {
  %slot_d_1 = alloca { { ptr, i64 } }, align 8
  %slot_r_2 = alloca { i16, i16, i16, i16 }, align 8
  %slot_i_3 = alloca i64, align 8
  %slot_c_4 = alloca { i16, i16, i16, i16 }, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_1
  %t1 = load { { ptr, i64 } }, ptr %slot_d_1
  %t2 = call { i16, i16, i16, i16 } @aila_std_damage__cells_bounds({ { ptr, i64 } } %t1)
  store { i16, i16, i16, i16 } %t2, ptr %slot_r_2
  store i64 0, ptr %slot_i_3
  br label %for.cond2
for.cond2:
  %t3 = load i64, ptr %slot_i_3
  %t4 = icmp ult i64 %t3, 8
  br i1 %t4, label %for.body4, label %for.end1
for.body4:
  %t5 = load { { ptr, i64 } }, ptr %slot_d_1
  %t6 = load i64, ptr %slot_i_3
  %t7 = call { i16, i16, i16, i16 } @aila_std_damage__clip_at({ { ptr, i64 } } %t5, i64 %t6)
  store { i16, i16, i16, i16 } %t7, ptr %slot_c_4
  %t8 = load { i16, i16, i16, i16 }, ptr %slot_c_4
  %t9 = call i1 @aila_std_damage__Rect__is_empty({ i16, i16, i16, i16 } %t8)
  %t10 = xor i1 %t9, true
  br i1 %t10, label %if.then6, label %if.next7
if.then6:
  %t11 = load { i16, i16, i16, i16 }, ptr %slot_r_2
  %t12 = load { i16, i16, i16, i16 }, ptr %slot_c_4
  %t13 = call { i16, i16, i16, i16 } @aila_std_damage__Rect__union({ i16, i16, i16, i16 } %t11, { i16, i16, i16, i16 } %t12)
  store { i16, i16, i16, i16 } %t13, ptr %slot_r_2
  br label %if.end5
if.next7:
  br label %if.end5
if.end5:
  %t14 = load i64, ptr %slot_i_3
  %t15 = add i64 %t14, 1
  store i64 %t15, ptr %slot_i_3
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t16 = load { i16, i16, i16, i16 }, ptr %slot_r_2
  ret { i16, i16, i16, i16 } %t16
}
