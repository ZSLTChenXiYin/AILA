; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [28 x i8] c"match \E6\9C\AA\E8\A6\86\E7\9B\96\E7\9A\84\E5\88\A4\E5\88\AB\E5\80\BC\00", align 1
@.str.1 = private unnamed_addr constant [10 x i8] c"main.aila\00", align 1
@.str.2 = private unnamed_addr constant [4 x i8] c"red\00", align 1
@.str.3 = private unnamed_addr constant [6 x i8] c"green\00", align 1
@.str.4 = private unnamed_addr constant [5 x i8] c"blue\00", align 1
@.str.5 = private unnamed_addr constant [11 x i8] c"digit: %d\0A\00", align 1

declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare void @aila_std_io__println({ ptr, i64 })
declare ptr @aila_rt_static_region()
declare ptr @aila_rt_region_alloc(ptr, i64, i64)
declare void @aila_std_io__printf({ ptr, i64 }, { ptr, i64 })

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
  %slot___lit_8 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_9 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_10 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_11 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_12 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_13 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___idx_14 = alloca i64, align 8
  %slot___iter_15 = alloca { ptr, i64 }, align 8
  %slot___len_16 = alloca i64, align 8
  %slot___lit_17 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_18 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
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
  %t5 = add i32 %t3, %t4
  store i32 %t5, ptr %slot_total_1
  br label %for.step3
for.step3:
  %t6 = load i32, ptr %slot_i_2
  %t7 = add i32 %t6, 1
  store i32 %t7, ptr %slot_i_2
  br label %for.cond2
for.end1:
  %t8 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t8, ptr %slot___lit_8
  %t9 = load i32, ptr %slot_total_1
  %t10 = sext i32 %t9 to i64
  %t11 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_8, i32 0, i32 1, i32 0
  store i64 %t10, ptr %t11
  %t12 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_8
  %t13 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_9, i64 0
  store { i8, { { ptr, i64 } } } %t12, ptr %t13
  %t14 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_9, 0
  %t15 = insertvalue { ptr, i64 } %t14, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t15)
  store i32 0, ptr %slot_hits_3
  store i1 0, ptr %slot_done_4
  br label %for.cond6
for.cond6:
  %t16 = load i1, ptr %slot_done_4
  %t17 = xor i1 %t16, true
  br i1 %t17, label %for.body8, label %for.end5
for.body8:
  %t18 = load i32, ptr %slot_hits_3
  %t19 = add i32 %t18, 1
  store i32 %t19, ptr %slot_hits_3
  %t20 = load i32, ptr %slot_hits_3
  %t21 = icmp sge i32 %t20, 3
  br i1 %t21, label %if.then10, label %if.next11
if.then10:
  store i1 1, ptr %slot_done_4
  br label %if.end9
if.next11:
  br label %if.end9
if.end9:
  br label %for.step7
for.step7:
  br label %for.cond6
for.end5:
  %t22 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t22, ptr %slot___lit_10
  %t23 = load i32, ptr %slot_hits_3
  %t24 = sext i32 %t23 to i64
  %t25 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_10, i32 0, i32 1, i32 0
  store i64 %t24, ptr %t25
  %t26 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_10
  %t27 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_11, i64 0
  store { i8, { { ptr, i64 } } } %t26, ptr %t27
  %t28 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_11, 0
  %t29 = insertvalue { ptr, i64 } %t28, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t29)
  %t30 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t30, ptr %slot___lit_12
  %t31 = call { ptr, i64 } @aila_main__name_of(i32 2)
  %t32 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_12, i32 0, i32 1, i32 0
  store { ptr, i64 } %t31, ptr %t32
  %t33 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_12
  %t34 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_13, i64 0
  store { i8, { { ptr, i64 } } } %t33, ptr %t34
  %t35 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_13, 0
  %t36 = insertvalue { ptr, i64 } %t35, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t36)
  %t37 = call ptr @aila_rt_static_region()
  %t38 = call ptr @aila_rt_region_alloc(ptr %t37, i64 12, i64 4)
  %t39 = getelementptr inbounds i32, ptr %t38, i64 0
  store i32 10, ptr %t39
  %t40 = getelementptr inbounds i32, ptr %t38, i64 1
  store i32 20, ptr %t40
  %t41 = getelementptr inbounds i32, ptr %t38, i64 2
  store i32 7, ptr %t41
  %t42 = insertvalue { ptr, i64 } undef, ptr %t38, 0
  %t43 = insertvalue { ptr, i64 } %t42, i64 3, 1
  store { ptr, i64 } %t43, ptr %slot_samples_5
  store i64 0, ptr %slot___idx_14
  %t44 = load { ptr, i64 }, ptr %slot_samples_5
  store { ptr, i64 } %t44, ptr %slot___iter_15
  %t46 = getelementptr inbounds { ptr, i64 }, ptr %slot___iter_15, i32 0, i32 1
  %t45 = load i64, ptr %t46
  store i64 %t45, ptr %slot___len_16
  %t47 = load ptr, ptr %slot___iter_15
  br label %fi.cond13
fi.cond13:
  %t48 = load i64, ptr %slot___idx_14
  %t49 = load i64, ptr %slot___len_16
  %t50 = icmp ult i64 %t48, %t49
  br i1 %t50, label %fi.body15, label %for.end12
fi.body15:
  %t51 = getelementptr inbounds i32, ptr %t47, i64 %t48
  %t52 = load i32, ptr %t51
  store i32 %t52, ptr %slot_x_7
  store i64 %t48, ptr %slot_i_6
  %t53 = load i64, ptr %slot_i_6
  %t54 = icmp eq i64 %t53, 2
  br i1 %t54, label %if.then17, label %if.next18
if.then17:
  %t55 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t56 = insertvalue { ptr, i64 } %t55, i64 10, 1
  %t57 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t57, ptr %slot___lit_17
  %t58 = load i32, ptr %slot_x_7
  %t59 = sext i32 %t58 to i64
  %t60 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_17, i32 0, i32 1, i32 0
  store i64 %t59, ptr %t60
  %t61 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_17
  %t62 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_18, i64 0
  store { i8, { { ptr, i64 } } } %t61, ptr %t62
  %t63 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_18, 0
  %t64 = insertvalue { ptr, i64 } %t63, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t56, { ptr, i64 } %t64)
  br label %if.end16
if.next18:
  br label %if.end16
if.end16:
  br label %fi.step14
fi.step14:
  %t65 = add i64 %t48, 1
  store i64 %t65, ptr %slot___idx_14
  br label %fi.cond13
for.end12:
  ret void
}

define i32 @main(i32 %argc, ptr %argv) {
entry:
  call void @aila_main__main()
  ret i32 0
}
