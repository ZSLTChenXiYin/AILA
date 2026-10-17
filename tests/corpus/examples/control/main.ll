; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [28 x i8] c"match \E6\9C\AA\E8\A6\86\E7\9B\96\E7\9A\84\E5\88\A4\E5\88\AB\E5\80\BC\00", align 1
@.str.1 = private unnamed_addr constant [10 x i8] c"main.aila\00", align 1
@.str.2 = private unnamed_addr constant [4 x i8] c"red\00", align 1
@.str.3 = private unnamed_addr constant [6 x i8] c"green\00", align 1
@.str.4 = private unnamed_addr constant [5 x i8] c"blue\00", align 1
@.str.5 = private unnamed_addr constant [17 x i8] c"integer overflow\00", align 1
@.str.6 = private unnamed_addr constant [8 x i8] c"digit: \00", align 1

declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare void @aila_std_io__print_i32_nl(i32)
declare void @aila_std_io__print_text({ ptr, i64 })
declare void @aila_std_io__print_nl()
declare ptr @aila_rt_static_region()
declare ptr @aila_rt_region_alloc(ptr, i64, i64)

define { ptr, i64 } @aila_main__name_of(i32 %p_c) {
  %slot_c_1 = alloca i32, align 8
  %slot___match_2 = alloca { ptr, i64 }, align 8
  %slot___head_3 = alloca i32, align 8
  store i32 %p_c, ptr %slot_c_1
  %t1 = load i32, ptr %slot_c_1
  store i32 %t1, ptr %slot___head_3
  %t2 = load i32, ptr %slot___head_3
  %t3 = icmp eq i32 %t2, 0
  br i1 %t3, label %m.arm02, label %m.next5
m.next5:
  %t4 = icmp eq i32 %t2, 1
  br i1 %t4, label %m.arm13, label %m.next6
m.next6:
  %t5 = icmp eq i32 %t2, 2
  br i1 %t5, label %m.arm24, label %m.next7
m.next7:
  call void @aila_rt_trap(ptr @.str.0, i64 27, ptr @.str.1, i64 13)
  unreachable
m.arm02:
  %t6 = insertvalue { ptr, i64 } undef, ptr @.str.2, 0
  %t7 = insertvalue { ptr, i64 } %t6, i64 3, 1
  store { ptr, i64 } %t7, ptr %slot___match_2
  br label %m.end1
m.arm13:
  %t8 = insertvalue { ptr, i64 } undef, ptr @.str.3, 0
  %t9 = insertvalue { ptr, i64 } %t8, i64 5, 1
  store { ptr, i64 } %t9, ptr %slot___match_2
  br label %m.end1
m.arm24:
  %t10 = insertvalue { ptr, i64 } undef, ptr @.str.4, 0
  %t11 = insertvalue { ptr, i64 } %t10, i64 4, 1
  store { ptr, i64 } %t11, ptr %slot___match_2
  br label %m.end1
m.end1:
  %t12 = load { ptr, i64 }, ptr %slot___match_2
  ret { ptr, i64 } %t12
}

define void @aila_main__main() {
  %slot_total_1 = alloca i32, align 8
  %slot_i_2 = alloca i32, align 8
  %slot_hits_3 = alloca i32, align 8
  %slot_done_4 = alloca i1, align 8
  %slot_samples_5 = alloca { ptr, i64 }, align 8
  %slot_i_6 = alloca i64, align 8
  %slot_x_7 = alloca i32, align 8
  %slot___idx_8 = alloca i64, align 8
  %slot___iter_9 = alloca { ptr, i64 }, align 8
  %slot___len_10 = alloca i64, align 8
  store i32 0, ptr %slot_total_1
  store i32 1, ptr %slot_i_2
  br label %for.cond2
for.cond2:
  %t1 = load i32, ptr %slot_i_2
  %t2 = icmp sle i32 %t1, 5
  br i1 %t2, label %for.body4, label %for.end1
for.body4:
  %t3 = load i32, ptr %slot_total_1
  %t4 = load i32, ptr %slot_i_2
  %t5 = call { i32, i1 } @llvm.sadd.with.overflow.i32(i32 %t3, i32 %t4)
  %t6 = extractvalue { i32, i1 } %t5, 0
  %t7 = extractvalue { i32, i1 } %t5, 1
  br i1 %t7, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.5, i64 16, ptr @.str.1, i64 24)
  unreachable
chk.ok5:
  store i32 %t6, ptr %slot_total_1
  br label %for.step3
for.step3:
  %t8 = load i32, ptr %slot_i_2
  %t9 = call { i32, i1 } @llvm.sadd.with.overflow.i32(i32 %t8, i32 1)
  %t10 = extractvalue { i32, i1 } %t9, 0
  %t11 = extractvalue { i32, i1 } %t9, 1
  br i1 %t11, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.5, i64 16, ptr @.str.1, i64 23)
  unreachable
chk.ok7:
  store i32 %t10, ptr %slot_i_2
  br label %for.cond2
for.end1:
  %t12 = load i32, ptr %slot_total_1
  call void @aila_std_io__print_i32_nl(i32 %t12)
  store i32 0, ptr %slot_hits_3
  store i1 0, ptr %slot_done_4
  br label %for.cond10
for.cond10:
  %t13 = load i1, ptr %slot_done_4
  %t14 = xor i1 %t13, true
  br i1 %t14, label %for.body12, label %for.end9
for.body12:
  %t15 = load i32, ptr %slot_hits_3
  %t16 = call { i32, i1 } @llvm.sadd.with.overflow.i32(i32 %t15, i32 1)
  %t17 = extractvalue { i32, i1 } %t16, 0
  %t18 = extractvalue { i32, i1 } %t16, 1
  br i1 %t18, label %chk.bad14, label %chk.ok13
chk.bad14:
  call void @aila_rt_trap(ptr @.str.5, i64 16, ptr @.str.1, i64 32)
  unreachable
chk.ok13:
  store i32 %t17, ptr %slot_hits_3
  %t19 = load i32, ptr %slot_hits_3
  %t20 = icmp sge i32 %t19, 3
  br i1 %t20, label %if.then16, label %if.next17
if.then16:
  store i1 1, ptr %slot_done_4
  br label %if.end15
if.next17:
  br label %if.end15
if.end15:
  br label %for.step11
for.step11:
  br label %for.cond10
for.end9:
  %t21 = load i32, ptr %slot_hits_3
  call void @aila_std_io__print_i32_nl(i32 %t21)
  %t22 = call { ptr, i64 } @aila_main__name_of(i32 2)
  call void @aila_std_io__print_text({ ptr, i64 } %t22)
  call void @aila_std_io__print_nl()
  %t23 = call ptr @aila_rt_static_region()
  %t24 = call ptr @aila_rt_region_alloc(ptr %t23, i64 12, i64 4)
  %t25 = getelementptr inbounds i32, ptr %t24, i64 0
  store i32 10, ptr %t25
  %t26 = getelementptr inbounds i32, ptr %t24, i64 1
  store i32 20, ptr %t26
  %t27 = getelementptr inbounds i32, ptr %t24, i64 2
  store i32 7, ptr %t27
  %t28 = insertvalue { ptr, i64 } undef, ptr %t24, 0
  %t29 = insertvalue { ptr, i64 } %t28, i64 3, 1
  store { ptr, i64 } %t29, ptr %slot_samples_5
  store i64 0, ptr %slot___idx_8
  %t30 = load { ptr, i64 }, ptr %slot_samples_5
  store { ptr, i64 } %t30, ptr %slot___iter_9
  %t32 = getelementptr inbounds { ptr, i64 }, ptr %slot___iter_9, i32 0, i32 1
  %t31 = load i64, ptr %t32
  store i64 %t31, ptr %slot___len_10
  %t33 = load ptr, ptr %slot___iter_9
  br label %fi.cond19
fi.cond19:
  %t34 = load i64, ptr %slot___idx_8
  %t35 = load i64, ptr %slot___len_10
  %t36 = icmp ult i64 %t34, %t35
  br i1 %t36, label %fi.body21, label %for.end18
fi.body21:
  %t37 = getelementptr inbounds i32, ptr %t33, i64 %t34
  %t38 = load i32, ptr %t37
  store i32 %t38, ptr %slot_x_7
  store i64 %t34, ptr %slot_i_6
  %t39 = load i64, ptr %slot_i_6
  %t40 = icmp eq i64 %t39, 2
  br i1 %t40, label %if.then23, label %if.next24
if.then23:
  %t41 = insertvalue { ptr, i64 } undef, ptr @.str.6, 0
  %t42 = insertvalue { ptr, i64 } %t41, i64 7, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t42)
  %t43 = load i32, ptr %slot_x_7
  call void @aila_std_io__print_i32_nl(i32 %t43)
  br label %if.end22
if.next24:
  br label %if.end22
if.end22:
  br label %fi.step20
fi.step20:
  %t44 = add i64 %t34, 1
  store i64 %t44, ptr %slot___idx_8
  br label %fi.cond19
for.end18:
  ret void
}

define i32 @main() {
entry:
  call void @aila_main__main()
  ret i32 0
}
