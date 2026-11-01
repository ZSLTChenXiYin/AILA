; AILA module: std.rand
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [25 x i8] c"shift count out of range\00", align 1
@.str.1 = private unnamed_addr constant [14 x i8] c"std/rand.aila\00", align 1
@.str.2 = private unnamed_addr constant [17 x i8] c"division by zero\00", align 1


declare i64 @aila_rt_seed()
declare void @aila_rt_trap(ptr, i64, ptr, i64)

define { i64 } @aila_std_rand__new() {
  %slot_s_1 = alloca i64, align 8
  %slot___lit_2 = alloca { i64 }, align 8
  %slot___lit_3 = alloca { i64 }, align 8
  %t1 = call i64 @aila_rt_seed()
  store i64 %t1, ptr %slot_s_1
  %t2 = load i64, ptr %slot_s_1
  %t3 = icmp eq i64 %t2, 0
  br i1 %t3, label %if.then2, label %if.next3
if.then2:
  %t4 = getelementptr inbounds { i64 }, ptr %slot___lit_2, i32 0, i32 0
  store i64 -7046029254386353131, ptr %t4
  %t5 = load { i64 }, ptr %slot___lit_2
  ret { i64 } %t5
if.next3:
  br label %if.end1
if.end1:
  %t6 = load i64, ptr %slot_s_1
  %t7 = getelementptr inbounds { i64 }, ptr %slot___lit_3, i32 0, i32 0
  store i64 %t6, ptr %t7
  %t8 = load { i64 }, ptr %slot___lit_3
  ret { i64 } %t8
}

define void @aila_std_rand__Rand__seed(ptr %p_self, i64 %p_s) {
  %slot_self_1 = alloca { i64 }, align 8
  %slot_s_2 = alloca i64, align 8
  store i64 %p_s, ptr %slot_s_2
  %t1 = load i64, ptr %slot_s_2
  %t2 = icmp eq i64 %t1, 0
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  %t3 = getelementptr inbounds { i64 }, ptr %p_self, i32 0, i32 0
  store i64 -7046029254386353131, ptr %t3
  ret void
if.next3:
  br label %if.end1
if.end1:
  %t4 = getelementptr inbounds { i64 }, ptr %p_self, i32 0, i32 0
  %t5 = load i64, ptr %slot_s_2
  store i64 %t5, ptr %t4
  ret void
}

define i64 @aila_std_rand__Rand__next_u64(ptr %p_self) {
  %slot_self_1 = alloca { i64 }, align 8
  %slot_x_2 = alloca i64, align 8
  %t1 = getelementptr inbounds { i64 }, ptr %p_self, i32 0, i32 0
  %t2 = load i64, ptr %t1
  store i64 %t2, ptr %slot_x_2
  %t3 = load i64, ptr %slot_x_2
  %t4 = load i64, ptr %slot_x_2
  %t5 = icmp uge i64 13, 64
  br i1 %t5, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 24, ptr @.str.1, i64 36)
  unreachable
chk.ok1:
  %t6 = shl i64 %t4, 13
  %t7 = xor i64 %t3, %t6
  store i64 %t7, ptr %slot_x_2
  %t8 = load i64, ptr %slot_x_2
  %t9 = load i64, ptr %slot_x_2
  %t10 = icmp uge i64 7, 64
  br i1 %t10, label %chk.bad4, label %chk.ok3
chk.bad4:
  call void @aila_rt_trap(ptr @.str.0, i64 24, ptr @.str.1, i64 37)
  unreachable
chk.ok3:
  %t11 = lshr i64 %t9, 7
  %t12 = xor i64 %t8, %t11
  store i64 %t12, ptr %slot_x_2
  %t13 = load i64, ptr %slot_x_2
  %t14 = load i64, ptr %slot_x_2
  %t15 = icmp uge i64 17, 64
  br i1 %t15, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 24, ptr @.str.1, i64 38)
  unreachable
chk.ok5:
  %t16 = shl i64 %t14, 17
  %t17 = xor i64 %t13, %t16
  store i64 %t17, ptr %slot_x_2
  %t18 = getelementptr inbounds { i64 }, ptr %p_self, i32 0, i32 0
  %t19 = load i64, ptr %slot_x_2
  store i64 %t19, ptr %t18
  %t20 = load i64, ptr %slot_x_2
  ret i64 %t20
}

define i64 @aila_std_rand__Rand__below(ptr %p_self, i64 %p_bound) {
  %slot_self_1 = alloca { i64 }, align 8
  %slot_bound_2 = alloca i64, align 8
  %slot_zone_3 = alloca i64, align 8
  %slot_x_4 = alloca i64, align 8
  %slot_ok_5 = alloca i1, align 8
  store i64 %p_bound, ptr %slot_bound_2
  %t1 = load i64, ptr %slot_bound_2
  %t2 = icmp ult i64 %t1, 2
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  ret i64 0
if.next3:
  br label %if.end1
if.end1:
  %t3 = load i64, ptr %slot_bound_2
  %t4 = icmp eq i64 %t3, 0
  br i1 %t4, label %chk.bad5, label %chk.ok4
chk.bad5:
  call void @aila_rt_trap(ptr @.str.2, i64 16, ptr @.str.1, i64 50)
  unreachable
chk.ok4:
  %t5 = urem i64 -1, %t3
  %t6 = add i64 %t5, 1
  store i64 %t6, ptr %slot_zone_3
  store i64 0, ptr %slot_x_4
  store i1 0, ptr %slot_ok_5
  br label %for.cond7
for.cond7:
  %t7 = load i1, ptr %slot_ok_5
  %t8 = xor i1 %t7, true
  br i1 %t8, label %for.body9, label %for.end6
for.body9:
  %t9 = call i64 @aila_std_rand__Rand__next_u64(ptr %p_self)
  store i64 %t9, ptr %slot_x_4
  %t10 = load i64, ptr %slot_x_4
  %t11 = load i64, ptr %slot_zone_3
  %t12 = icmp uge i64 %t10, %t11
  br i1 %t12, label %if.then11, label %if.next12
if.then11:
  store i1 1, ptr %slot_ok_5
  br label %if.end10
if.next12:
  br label %if.end10
if.end10:
  br label %for.step8
for.step8:
  br label %for.cond7
for.end6:
  %t13 = load i64, ptr %slot_x_4
  %t14 = load i64, ptr %slot_bound_2
  %t15 = icmp eq i64 %t14, 0
  br i1 %t15, label %chk.bad14, label %chk.ok13
chk.bad14:
  call void @aila_rt_trap(ptr @.str.2, i64 16, ptr @.str.1, i64 59)
  unreachable
chk.ok13:
  %t16 = urem i64 %t13, %t14
  ret i64 %t16
}
