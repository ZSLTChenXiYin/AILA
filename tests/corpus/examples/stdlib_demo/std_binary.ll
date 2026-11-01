; AILA module: std.binary
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.1 = private unnamed_addr constant [16 x i8] c"std/binary.aila\00", align 1
@.str.2 = private unnamed_addr constant [25 x i8] c"shift count out of range\00", align 1

declare void @aila_rt_trap(ptr, i64, ptr, i64)

define i16 @aila_std_binary__le_read_u16({ ptr, i64 } %p_buf, i64 %p_off) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_v_3 = alloca i32, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  %t1 = load i64, ptr %slot_off_2
  %t2 = load { ptr, i64 }, ptr %slot_buf_1
  %t3 = extractvalue { ptr, i64 } %t2, 0
  %t4 = extractvalue { ptr, i64 } %t2, 1
  %t5 = icmp uge i64 %t1, %t4
  br i1 %t5, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 17)
  unreachable
chk.ok1:
  %t6 = getelementptr inbounds i8, ptr %t3, i64 %t1
  %t7 = load i8, ptr %t6
  %t8 = zext i8 %t7 to i32
  %t9 = load i64, ptr %slot_off_2
  %t10 = add i64 %t9, 1
  %t11 = load { ptr, i64 }, ptr %slot_buf_1
  %t12 = extractvalue { ptr, i64 } %t11, 0
  %t13 = extractvalue { ptr, i64 } %t11, 1
  %t14 = icmp uge i64 %t10, %t13
  br i1 %t14, label %chk.bad4, label %chk.ok3
chk.bad4:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 17)
  unreachable
chk.ok3:
  %t15 = getelementptr inbounds i8, ptr %t12, i64 %t10
  %t16 = load i8, ptr %t15
  %t17 = zext i8 %t16 to i32
  %t18 = icmp uge i32 8, 32
  br i1 %t18, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 17)
  unreachable
chk.ok5:
  %t19 = shl i32 %t17, 8
  %t20 = or i32 %t8, %t19
  store i32 %t20, ptr %slot_v_3
  %t21 = load i32, ptr %slot_v_3
  %t22 = trunc i32 %t21 to i16
  ret i16 %t22
}

define i32 @aila_std_binary__le_read_u32({ ptr, i64 } %p_buf, i64 %p_off) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_lo_3 = alloca i32, align 8
  %slot_hi_4 = alloca i32, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = call i16 @aila_std_binary__le_read_u16({ ptr, i64 } %t1, i64 %t2)
  %t4 = zext i16 %t3 to i32
  store i32 %t4, ptr %slot_lo_3
  %t5 = load { ptr, i64 }, ptr %slot_buf_1
  %t6 = load i64, ptr %slot_off_2
  %t7 = add i64 %t6, 2
  %t8 = call i16 @aila_std_binary__le_read_u16({ ptr, i64 } %t5, i64 %t7)
  %t9 = zext i16 %t8 to i32
  store i32 %t9, ptr %slot_hi_4
  %t10 = load i32, ptr %slot_lo_3
  %t11 = load i32, ptr %slot_hi_4
  %t12 = icmp uge i32 16, 32
  br i1 %t12, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 24)
  unreachable
chk.ok1:
  %t13 = shl i32 %t11, 16
  %t14 = or i32 %t10, %t13
  ret i32 %t14
}

define i64 @aila_std_binary__le_read_u64({ ptr, i64 } %p_buf, i64 %p_off) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_lo_3 = alloca i32, align 8
  %slot_hi_4 = alloca i32, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = call i32 @aila_std_binary__le_read_u32({ ptr, i64 } %t1, i64 %t2)
  store i32 %t3, ptr %slot_lo_3
  %t4 = load { ptr, i64 }, ptr %slot_buf_1
  %t5 = load i64, ptr %slot_off_2
  %t6 = add i64 %t5, 4
  %t7 = call i32 @aila_std_binary__le_read_u32({ ptr, i64 } %t4, i64 %t6)
  store i32 %t7, ptr %slot_hi_4
  %t8 = load i32, ptr %slot_lo_3
  %t9 = zext i32 %t8 to i64
  %t10 = load i32, ptr %slot_hi_4
  %t11 = zext i32 %t10 to i64
  %t12 = icmp uge i64 32, 64
  br i1 %t12, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 30)
  unreachable
chk.ok1:
  %t13 = shl i64 %t11, 32
  %t14 = or i64 %t9, %t13
  ret i64 %t14
}

define i16 @aila_std_binary__le_read_i16({ ptr, i64 } %p_buf, i64 %p_off) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = call i16 @aila_std_binary__le_read_u16({ ptr, i64 } %t1, i64 %t2)
  ret i16 %t3
}

define i32 @aila_std_binary__le_read_i32({ ptr, i64 } %p_buf, i64 %p_off) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = call i32 @aila_std_binary__le_read_u32({ ptr, i64 } %t1, i64 %t2)
  ret i32 %t3
}

define i64 @aila_std_binary__le_read_i64({ ptr, i64 } %p_buf, i64 %p_off) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = call i64 @aila_std_binary__le_read_u64({ ptr, i64 } %t1, i64 %t2)
  ret i64 %t3
}

define void @aila_std_binary__le_write_u16({ ptr, i64 } %p_buf, i64 %p_off, i16 %p_v) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_v_3 = alloca i16, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  store i16 %p_v, ptr %slot_v_3
  %t1 = load i64, ptr %slot_off_2
  %t2 = load { ptr, i64 }, ptr %slot_buf_1
  %t3 = extractvalue { ptr, i64 } %t2, 0
  %t4 = extractvalue { ptr, i64 } %t2, 1
  %t5 = icmp uge i64 %t1, %t4
  br i1 %t5, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 46)
  unreachable
chk.ok1:
  %t6 = getelementptr inbounds i8, ptr %t3, i64 %t1
  %t7 = load i16, ptr %slot_v_3
  %t8 = trunc i16 %t7 to i8
  store i8 %t8, ptr %t6
  %t9 = load i64, ptr %slot_off_2
  %t10 = add i64 %t9, 1
  %t11 = load { ptr, i64 }, ptr %slot_buf_1
  %t12 = extractvalue { ptr, i64 } %t11, 0
  %t13 = extractvalue { ptr, i64 } %t11, 1
  %t14 = icmp uge i64 %t10, %t13
  br i1 %t14, label %chk.bad4, label %chk.ok3
chk.bad4:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 47)
  unreachable
chk.ok3:
  %t15 = getelementptr inbounds i8, ptr %t12, i64 %t10
  %t16 = load i16, ptr %slot_v_3
  %t17 = icmp uge i16 8, 16
  br i1 %t17, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 47)
  unreachable
chk.ok5:
  %t18 = lshr i16 %t16, 8
  %t19 = trunc i16 %t18 to i8
  store i8 %t19, ptr %t15
  ret void
}

define void @aila_std_binary__le_write_u32({ ptr, i64 } %p_buf, i64 %p_off, i32 %p_v) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_v_3 = alloca i32, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  store i32 %p_v, ptr %slot_v_3
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = load i32, ptr %slot_v_3
  %t4 = trunc i32 %t3 to i16
  call void @aila_std_binary__le_write_u16({ ptr, i64 } %t1, i64 %t2, i16 %t4)
  %t5 = load { ptr, i64 }, ptr %slot_buf_1
  %t6 = load i64, ptr %slot_off_2
  %t7 = add i64 %t6, 2
  %t8 = load i32, ptr %slot_v_3
  %t9 = icmp uge i32 16, 32
  br i1 %t9, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 52)
  unreachable
chk.ok1:
  %t10 = lshr i32 %t8, 16
  %t11 = trunc i32 %t10 to i16
  call void @aila_std_binary__le_write_u16({ ptr, i64 } %t5, i64 %t7, i16 %t11)
  ret void
}

define void @aila_std_binary__le_write_u64({ ptr, i64 } %p_buf, i64 %p_off, i64 %p_v) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_v_3 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  store i64 %p_v, ptr %slot_v_3
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = load i64, ptr %slot_v_3
  %t4 = trunc i64 %t3 to i32
  call void @aila_std_binary__le_write_u32({ ptr, i64 } %t1, i64 %t2, i32 %t4)
  %t5 = load { ptr, i64 }, ptr %slot_buf_1
  %t6 = load i64, ptr %slot_off_2
  %t7 = add i64 %t6, 4
  %t8 = load i64, ptr %slot_v_3
  %t9 = icmp uge i64 32, 64
  br i1 %t9, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 57)
  unreachable
chk.ok1:
  %t10 = lshr i64 %t8, 32
  %t11 = trunc i64 %t10 to i32
  call void @aila_std_binary__le_write_u32({ ptr, i64 } %t5, i64 %t7, i32 %t11)
  ret void
}

define void @aila_std_binary__le_write_i16({ ptr, i64 } %p_buf, i64 %p_off, i16 %p_v) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_v_3 = alloca i16, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  store i16 %p_v, ptr %slot_v_3
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = load i16, ptr %slot_v_3
  call void @aila_std_binary__le_write_u16({ ptr, i64 } %t1, i64 %t2, i16 %t3)
  ret void
}

define void @aila_std_binary__le_write_i32({ ptr, i64 } %p_buf, i64 %p_off, i32 %p_v) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_v_3 = alloca i32, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  store i32 %p_v, ptr %slot_v_3
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = load i32, ptr %slot_v_3
  call void @aila_std_binary__le_write_u32({ ptr, i64 } %t1, i64 %t2, i32 %t3)
  ret void
}

define void @aila_std_binary__le_write_i64({ ptr, i64 } %p_buf, i64 %p_off, i64 %p_v) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_v_3 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  store i64 %p_v, ptr %slot_v_3
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = load i64, ptr %slot_v_3
  call void @aila_std_binary__le_write_u64({ ptr, i64 } %t1, i64 %t2, i64 %t3)
  ret void
}

define i16 @aila_std_binary__be_read_u16({ ptr, i64 } %p_buf, i64 %p_off) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_v_3 = alloca i32, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  %t1 = load i64, ptr %slot_off_2
  %t2 = load { ptr, i64 }, ptr %slot_buf_1
  %t3 = extractvalue { ptr, i64 } %t2, 0
  %t4 = extractvalue { ptr, i64 } %t2, 1
  %t5 = icmp uge i64 %t1, %t4
  br i1 %t5, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 75)
  unreachable
chk.ok1:
  %t6 = getelementptr inbounds i8, ptr %t3, i64 %t1
  %t7 = load i8, ptr %t6
  %t8 = zext i8 %t7 to i32
  %t9 = icmp uge i32 8, 32
  br i1 %t9, label %chk.bad4, label %chk.ok3
chk.bad4:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 75)
  unreachable
chk.ok3:
  %t10 = shl i32 %t8, 8
  %t11 = load i64, ptr %slot_off_2
  %t12 = add i64 %t11, 1
  %t13 = load { ptr, i64 }, ptr %slot_buf_1
  %t14 = extractvalue { ptr, i64 } %t13, 0
  %t15 = extractvalue { ptr, i64 } %t13, 1
  %t16 = icmp uge i64 %t12, %t15
  br i1 %t16, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 75)
  unreachable
chk.ok5:
  %t17 = getelementptr inbounds i8, ptr %t14, i64 %t12
  %t18 = load i8, ptr %t17
  %t19 = zext i8 %t18 to i32
  %t20 = or i32 %t10, %t19
  store i32 %t20, ptr %slot_v_3
  %t21 = load i32, ptr %slot_v_3
  %t22 = trunc i32 %t21 to i16
  ret i16 %t22
}

define i32 @aila_std_binary__be_read_u32({ ptr, i64 } %p_buf, i64 %p_off) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_hi_3 = alloca i32, align 8
  %slot_lo_4 = alloca i32, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = call i16 @aila_std_binary__be_read_u16({ ptr, i64 } %t1, i64 %t2)
  %t4 = zext i16 %t3 to i32
  store i32 %t4, ptr %slot_hi_3
  %t5 = load { ptr, i64 }, ptr %slot_buf_1
  %t6 = load i64, ptr %slot_off_2
  %t7 = add i64 %t6, 2
  %t8 = call i16 @aila_std_binary__be_read_u16({ ptr, i64 } %t5, i64 %t7)
  %t9 = zext i16 %t8 to i32
  store i32 %t9, ptr %slot_lo_4
  %t10 = load i32, ptr %slot_hi_3
  %t11 = icmp uge i32 16, 32
  br i1 %t11, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 82)
  unreachable
chk.ok1:
  %t12 = shl i32 %t10, 16
  %t13 = load i32, ptr %slot_lo_4
  %t14 = or i32 %t12, %t13
  ret i32 %t14
}

define i64 @aila_std_binary__be_read_u64({ ptr, i64 } %p_buf, i64 %p_off) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_hi_3 = alloca i32, align 8
  %slot_lo_4 = alloca i32, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = call i32 @aila_std_binary__be_read_u32({ ptr, i64 } %t1, i64 %t2)
  store i32 %t3, ptr %slot_hi_3
  %t4 = load { ptr, i64 }, ptr %slot_buf_1
  %t5 = load i64, ptr %slot_off_2
  %t6 = add i64 %t5, 4
  %t7 = call i32 @aila_std_binary__be_read_u32({ ptr, i64 } %t4, i64 %t6)
  store i32 %t7, ptr %slot_lo_4
  %t8 = load i32, ptr %slot_hi_3
  %t9 = zext i32 %t8 to i64
  %t10 = icmp uge i64 32, 64
  br i1 %t10, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 88)
  unreachable
chk.ok1:
  %t11 = shl i64 %t9, 32
  %t12 = load i32, ptr %slot_lo_4
  %t13 = zext i32 %t12 to i64
  %t14 = or i64 %t11, %t13
  ret i64 %t14
}

define i16 @aila_std_binary__be_read_i16({ ptr, i64 } %p_buf, i64 %p_off) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = call i16 @aila_std_binary__be_read_u16({ ptr, i64 } %t1, i64 %t2)
  ret i16 %t3
}

define i32 @aila_std_binary__be_read_i32({ ptr, i64 } %p_buf, i64 %p_off) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = call i32 @aila_std_binary__be_read_u32({ ptr, i64 } %t1, i64 %t2)
  ret i32 %t3
}

define i64 @aila_std_binary__be_read_i64({ ptr, i64 } %p_buf, i64 %p_off) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = call i64 @aila_std_binary__be_read_u64({ ptr, i64 } %t1, i64 %t2)
  ret i64 %t3
}

define void @aila_std_binary__be_write_u16({ ptr, i64 } %p_buf, i64 %p_off, i16 %p_v) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_v_3 = alloca i16, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  store i16 %p_v, ptr %slot_v_3
  %t1 = load i64, ptr %slot_off_2
  %t2 = load { ptr, i64 }, ptr %slot_buf_1
  %t3 = extractvalue { ptr, i64 } %t2, 0
  %t4 = extractvalue { ptr, i64 } %t2, 1
  %t5 = icmp uge i64 %t1, %t4
  br i1 %t5, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 104)
  unreachable
chk.ok1:
  %t6 = getelementptr inbounds i8, ptr %t3, i64 %t1
  %t7 = load i16, ptr %slot_v_3
  %t8 = icmp uge i16 8, 16
  br i1 %t8, label %chk.bad4, label %chk.ok3
chk.bad4:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 104)
  unreachable
chk.ok3:
  %t9 = lshr i16 %t7, 8
  %t10 = trunc i16 %t9 to i8
  store i8 %t10, ptr %t6
  %t11 = load i64, ptr %slot_off_2
  %t12 = add i64 %t11, 1
  %t13 = load { ptr, i64 }, ptr %slot_buf_1
  %t14 = extractvalue { ptr, i64 } %t13, 0
  %t15 = extractvalue { ptr, i64 } %t13, 1
  %t16 = icmp uge i64 %t12, %t15
  br i1 %t16, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 105)
  unreachable
chk.ok5:
  %t17 = getelementptr inbounds i8, ptr %t14, i64 %t12
  %t18 = load i16, ptr %slot_v_3
  %t19 = trunc i16 %t18 to i8
  store i8 %t19, ptr %t17
  ret void
}

define void @aila_std_binary__be_write_u32({ ptr, i64 } %p_buf, i64 %p_off, i32 %p_v) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_v_3 = alloca i32, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  store i32 %p_v, ptr %slot_v_3
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = load i32, ptr %slot_v_3
  %t4 = icmp uge i32 16, 32
  br i1 %t4, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 109)
  unreachable
chk.ok1:
  %t5 = lshr i32 %t3, 16
  %t6 = trunc i32 %t5 to i16
  call void @aila_std_binary__be_write_u16({ ptr, i64 } %t1, i64 %t2, i16 %t6)
  %t7 = load { ptr, i64 }, ptr %slot_buf_1
  %t8 = load i64, ptr %slot_off_2
  %t9 = add i64 %t8, 2
  %t10 = load i32, ptr %slot_v_3
  %t11 = trunc i32 %t10 to i16
  call void @aila_std_binary__be_write_u16({ ptr, i64 } %t7, i64 %t9, i16 %t11)
  ret void
}

define void @aila_std_binary__be_write_u64({ ptr, i64 } %p_buf, i64 %p_off, i64 %p_v) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_v_3 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  store i64 %p_v, ptr %slot_v_3
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = load i64, ptr %slot_v_3
  %t4 = icmp uge i64 32, 64
  br i1 %t4, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 24, ptr @.str.1, i64 114)
  unreachable
chk.ok1:
  %t5 = lshr i64 %t3, 32
  %t6 = trunc i64 %t5 to i32
  call void @aila_std_binary__be_write_u32({ ptr, i64 } %t1, i64 %t2, i32 %t6)
  %t7 = load { ptr, i64 }, ptr %slot_buf_1
  %t8 = load i64, ptr %slot_off_2
  %t9 = add i64 %t8, 4
  %t10 = load i64, ptr %slot_v_3
  %t11 = trunc i64 %t10 to i32
  call void @aila_std_binary__be_write_u32({ ptr, i64 } %t7, i64 %t9, i32 %t11)
  ret void
}

define void @aila_std_binary__be_write_i16({ ptr, i64 } %p_buf, i64 %p_off, i16 %p_v) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_v_3 = alloca i16, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  store i16 %p_v, ptr %slot_v_3
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = load i16, ptr %slot_v_3
  call void @aila_std_binary__be_write_u16({ ptr, i64 } %t1, i64 %t2, i16 %t3)
  ret void
}

define void @aila_std_binary__be_write_i32({ ptr, i64 } %p_buf, i64 %p_off, i32 %p_v) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_v_3 = alloca i32, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  store i32 %p_v, ptr %slot_v_3
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = load i32, ptr %slot_v_3
  call void @aila_std_binary__be_write_u32({ ptr, i64 } %t1, i64 %t2, i32 %t3)
  ret void
}

define void @aila_std_binary__be_write_i64({ ptr, i64 } %p_buf, i64 %p_off, i64 %p_v) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_off_2 = alloca i64, align 8
  %slot_v_3 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_off, ptr %slot_off_2
  store i64 %p_v, ptr %slot_v_3
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  %t2 = load i64, ptr %slot_off_2
  %t3 = load i64, ptr %slot_v_3
  call void @aila_std_binary__be_write_u64({ ptr, i64 } %t1, i64 %t2, i64 %t3)
  ret void
}
