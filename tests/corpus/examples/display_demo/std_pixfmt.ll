; AILA module: std.pixfmt
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [17 x i8] c"division by zero\00", align 1
@.str.1 = private unnamed_addr constant [16 x i8] c"std/pixfmt.aila\00", align 1
@.str.2 = private unnamed_addr constant [25 x i8] c"shift count out of range\00", align 1


declare void @aila_rt_trap(ptr, i64, ptr, i64)

define { i8, i8, i8, i8 } @aila_std_pixfmt__mono1() {
  %slot___lit_1 = alloca { i8, i8, i8, i8 }, align 8
  %t1 = getelementptr inbounds { i8, i8, i8, i8 }, ptr %slot___lit_1, i32 0, i32 0
  store i8 1, ptr %t1
  %t2 = getelementptr inbounds { i8, i8, i8, i8 }, ptr %slot___lit_1, i32 0, i32 1
  store i8 1, ptr %t2
  %t3 = getelementptr inbounds { i8, i8, i8, i8 }, ptr %slot___lit_1, i32 0, i32 2
  store i8 4, ptr %t3
  %t4 = getelementptr inbounds { i8, i8, i8, i8 }, ptr %slot___lit_1, i32 0, i32 3
  store i8 0, ptr %t4
  %t5 = load { i8, i8, i8, i8 }, ptr %slot___lit_1
  ret { i8, i8, i8, i8 } %t5
}

define { i8, i8, i8, i8 } @aila_std_pixfmt__of(i8 %p_code) {
  %slot_code_1 = alloca i8, align 8
  store i8 %p_code, ptr %slot_code_1
  %t1 = load i8, ptr %slot_code_1
  %t2 = icmp eq i8 %t1, 1
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  %t3 = call { i8, i8, i8, i8 } @aila_std_pixfmt__mono1()
  ret { i8, i8, i8, i8 } %t3
if.next3:
  br label %if.end1
if.end1:
  %t4 = call { i8, i8, i8, i8 } @aila_std_pixfmt__mono1()
  ret { i8, i8, i8, i8 } %t4
}

define i16 @aila_std_pixfmt__row_bytes({ i8, i8, i8, i8 } %p_fmt, i16 %p_w) {
  %slot_fmt_1 = alloca { i8, i8, i8, i8 }, align 8
  %slot_w_2 = alloca i16, align 8
  %slot_bits_3 = alloca i64, align 8
  store { i8, i8, i8, i8 } %p_fmt, ptr %slot_fmt_1
  store i16 %p_w, ptr %slot_w_2
  %t1 = getelementptr inbounds { i8, i8, i8, i8 }, ptr %slot_fmt_1, i32 0, i32 1
  %t2 = load i8, ptr %t1
  %t3 = icmp eq i8 %t2, 0
  br i1 %t3, label %if.then2, label %if.next3
if.then2:
  ret i16 0
if.next3:
  br label %if.end1
if.end1:
  %t4 = load i16, ptr %slot_w_2
  %t5 = zext i16 %t4 to i64
  %t6 = getelementptr inbounds { i8, i8, i8, i8 }, ptr %slot_fmt_1, i32 0, i32 1
  %t7 = load i8, ptr %t6
  %t8 = zext i8 %t7 to i64
  %t9 = mul i64 %t5, %t8
  store i64 %t9, ptr %slot_bits_3
  %t10 = load i64, ptr %slot_bits_3
  %t11 = add i64 %t10, 7
  %t12 = icmp eq i64 8, 0
  br i1 %t12, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 62)
  unreachable
chk.ok4:
  %t13 = udiv i64 %t11, 8
  %t14 = trunc i64 %t13 to i16
  ret i16 %t14
}

define i16 @aila_std_pixfmt__stride_for({ i8, i8, i8, i8 } %p_fmt, i16 %p_w) {
  %slot_fmt_1 = alloca { i8, i8, i8, i8 }, align 8
  %slot_w_2 = alloca i16, align 8
  %slot_rb_3 = alloca i64, align 8
  %slot_al_4 = alloca i64, align 8
  %slot_st_5 = alloca i64, align 8
  store { i8, i8, i8, i8 } %p_fmt, ptr %slot_fmt_1
  store i16 %p_w, ptr %slot_w_2
  %t1 = load { i8, i8, i8, i8 }, ptr %slot_fmt_1
  %t2 = load i16, ptr %slot_w_2
  %t3 = call i16 @aila_std_pixfmt__row_bytes({ i8, i8, i8, i8 } %t1, i16 %t2)
  %t4 = zext i16 %t3 to i64
  store i64 %t4, ptr %slot_rb_3
  %t5 = getelementptr inbounds { i8, i8, i8, i8 }, ptr %slot_fmt_1, i32 0, i32 2
  %t6 = load i8, ptr %t5
  %t7 = zext i8 %t6 to i64
  store i64 %t7, ptr %slot_al_4
  %t8 = load i64, ptr %slot_al_4
  %t9 = icmp eq i64 %t8, 0
  br i1 %t9, label %if.then2, label %if.next3
if.then2:
  %t10 = load i64, ptr %slot_rb_3
  %t11 = trunc i64 %t10 to i16
  ret i16 %t11
if.next3:
  br label %if.end1
if.end1:
  %t12 = load i64, ptr %slot_rb_3
  %t13 = load i64, ptr %slot_al_4
  %t14 = icmp eq i64 %t13, 0
  br i1 %t14, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 71)
  unreachable
chk.ok7:
  %t15 = urem i64 %t12, %t13
  %t16 = icmp eq i64 %t15, 0
  br i1 %t16, label %if.then5, label %if.next6
if.then5:
  %t17 = load i64, ptr %slot_rb_3
  %t18 = trunc i64 %t17 to i16
  ret i16 %t18
if.next6:
  br label %if.end4
if.end4:
  %t19 = load i64, ptr %slot_rb_3
  %t20 = load i64, ptr %slot_al_4
  %t21 = icmp eq i64 %t20, 0
  br i1 %t21, label %chk.bad10, label %chk.ok9
chk.bad10:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 74)
  unreachable
chk.ok9:
  %t22 = udiv i64 %t19, %t20
  %t23 = load i64, ptr %slot_al_4
  %t24 = mul i64 %t22, %t23
  %t25 = load i64, ptr %slot_al_4
  %t26 = add i64 %t24, %t25
  store i64 %t26, ptr %slot_st_5
  %t27 = load i64, ptr %slot_st_5
  %t28 = trunc i64 %t27 to i16
  ret i16 %t28
}

define i16 @aila_std_pixfmt__byte_of({ i8, i8, i8, i8 } %p_fmt, i16 %p_x) {
  %slot_fmt_1 = alloca { i8, i8, i8, i8 }, align 8
  %slot_x_2 = alloca i16, align 8
  %slot_per_byte_3 = alloca i64, align 8
  store { i8, i8, i8, i8 } %p_fmt, ptr %slot_fmt_1
  store i16 %p_x, ptr %slot_x_2
  %t1 = getelementptr inbounds { i8, i8, i8, i8 }, ptr %slot_fmt_1, i32 0, i32 1
  %t2 = load i8, ptr %t1
  %t3 = icmp eq i8 %t2, 0
  br i1 %t3, label %if.then2, label %if.next3
if.then2:
  ret i16 0
if.next3:
  br label %if.end1
if.end1:
  %t4 = getelementptr inbounds { i8, i8, i8, i8 }, ptr %slot_fmt_1, i32 0, i32 1
  %t5 = load i8, ptr %t4
  %t6 = zext i8 %t5 to i64
  %t7 = icmp eq i64 %t6, 0
  br i1 %t7, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 82)
  unreachable
chk.ok4:
  %t8 = udiv i64 8, %t6
  store i64 %t8, ptr %slot_per_byte_3
  %t9 = load i64, ptr %slot_per_byte_3
  %t10 = icmp eq i64 %t9, 0
  br i1 %t10, label %if.then7, label %if.next8
if.then7:
  ret i16 0
if.next8:
  br label %if.end6
if.end6:
  %t11 = load i16, ptr %slot_x_2
  %t12 = zext i16 %t11 to i64
  %t13 = load i64, ptr %slot_per_byte_3
  %t14 = icmp eq i64 %t13, 0
  br i1 %t14, label %chk.bad10, label %chk.ok9
chk.bad10:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 86)
  unreachable
chk.ok9:
  %t15 = udiv i64 %t12, %t13
  %t16 = trunc i64 %t15 to i16
  ret i16 %t16
}

define i8 @aila_std_pixfmt__bit_of({ i8, i8, i8, i8 } %p_fmt, i16 %p_x) {
  %slot_fmt_1 = alloca { i8, i8, i8, i8 }, align 8
  %slot_x_2 = alloca i16, align 8
  %slot_per_byte_3 = alloca i64, align 8
  %slot_k_4 = alloca i8, align 8
  store { i8, i8, i8, i8 } %p_fmt, ptr %slot_fmt_1
  store i16 %p_x, ptr %slot_x_2
  %t1 = getelementptr inbounds { i8, i8, i8, i8 }, ptr %slot_fmt_1, i32 0, i32 1
  %t2 = load i8, ptr %t1
  %t3 = icmp eq i8 %t2, 0
  br i1 %t3, label %if.then2, label %if.next3
if.then2:
  ret i8 0
if.next3:
  br label %if.end1
if.end1:
  %t4 = getelementptr inbounds { i8, i8, i8, i8 }, ptr %slot_fmt_1, i32 0, i32 1
  %t5 = load i8, ptr %t4
  %t6 = zext i8 %t5 to i64
  %t7 = icmp eq i64 %t6, 0
  br i1 %t7, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 93)
  unreachable
chk.ok4:
  %t8 = udiv i64 8, %t6
  store i64 %t8, ptr %slot_per_byte_3
  %t9 = load i64, ptr %slot_per_byte_3
  %t10 = icmp eq i64 %t9, 0
  br i1 %t10, label %if.then7, label %if.next8
if.then7:
  ret i8 0
if.next8:
  br label %if.end6
if.end6:
  %t11 = load i16, ptr %slot_x_2
  %t12 = zext i16 %t11 to i64
  %t13 = load i64, ptr %slot_per_byte_3
  %t14 = icmp eq i64 %t13, 0
  br i1 %t14, label %chk.bad10, label %chk.ok9
chk.bad10:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 97)
  unreachable
chk.ok9:
  %t15 = urem i64 %t12, %t13
  %t16 = trunc i64 %t15 to i8
  store i8 %t16, ptr %slot_k_4
  %t17 = getelementptr inbounds { i8, i8, i8, i8 }, ptr %slot_fmt_1, i32 0, i32 3
  %t18 = load i8, ptr %t17
  %t19 = icmp eq i8 %t18, 1
  br i1 %t19, label %if.then12, label %if.next13
if.then12:
  %t20 = trunc i64 1 to i8
  %t21 = load i8, ptr %slot_k_4
  %t22 = icmp uge i8 %t21, 8
  br i1 %t22, label %chk.bad15, label %chk.ok14
chk.bad15:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 99)
  unreachable
chk.ok14:
  %t23 = shl i8 %t20, %t21
  ret i8 %t23
if.next13:
  br label %if.end11
if.end11:
  %t24 = trunc i64 128 to i8
  %t25 = load i8, ptr %slot_k_4
  %t26 = icmp uge i8 %t25, 8
  br i1 %t26, label %chk.bad17, label %chk.ok16
chk.bad17:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 101)
  unreachable
chk.ok16:
  %t27 = lshr i8 %t24, %t25
  ret i8 %t27
}
