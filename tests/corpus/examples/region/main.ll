; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [17 x i8] c"integer overflow\00", align 1
@.str.1 = private unnamed_addr constant [10 x i8] c"main.aila\00", align 1
@.str.2 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.3 = private unnamed_addr constant [7 x i8] c"sum = \00", align 1
@.str.4 = private unnamed_addr constant [10 x i8] c"buf[0] = \00", align 1
@.str.5 = private unnamed_addr constant [9 x i8] c"  len = \00", align 1
@.str.6 = private unnamed_addr constant [9 x i8] c" (bytes \00", align 1
@.str.7 = private unnamed_addr constant [2 x i8] c")\00", align 1
@.str.8 = private unnamed_addr constant [5 x i8] c"AILA\00", align 1

declare ptr @aila_rt_region_alloc(ptr, i64, i64)
declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare ptr @aila_rt_region_enter(i64)
declare void @aila_std_io__print_text({ ptr, i64 })
declare void @aila_std_io__print_i32_nl(i32)
declare void @aila_std_io__print_i32(i32)
declare void @aila_std_io__print_usize(i64)
declare void @aila_std_io__print_nl()
declare void @aila_rt_region_leave(ptr)

define { ptr, i64 } @aila_main__fill(ptr %r0, i64 %p_n, i8 %p_v) {
  %slot_n_1 = alloca i64, align 8
  %slot_v_2 = alloca i8, align 8
  %slot___fi_3 = alloca i64, align 8
  store i64 %p_n, ptr %slot_n_1
  store i8 %p_v, ptr %slot_v_2
  %t1 = load i64, ptr %slot_n_1
  %t2 = mul i64 %t1, 1
  %t3 = call ptr @aila_rt_region_alloc(ptr %r0, i64 %t2, i64 1)
  %t4 = load i8, ptr %slot_v_2
  store i64 0, ptr %slot___fi_3
  br label %ai.cond1
ai.cond1:
  %t5 = load i64, ptr %slot___fi_3
  %t6 = icmp ult i64 %t5, %t1
  br i1 %t6, label %ai.body2, label %ai.end3
ai.body2:
  %t7 = getelementptr inbounds i8, ptr %t3, i64 %t5
  store i8 %t4, ptr %t7
  %t8 = add i64 %t5, 1
  store i64 %t8, ptr %slot___fi_3
  br label %ai.cond1
ai.end3:
  %t9 = insertvalue { ptr, i64 } undef, ptr %t3, 0
  %t10 = insertvalue { ptr, i64 } %t9, i64 %t1, 1
  ret { ptr, i64 } %t10
}

define i32 @aila_main__sum_of({ ptr, i64 } %p_buf) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_s_2 = alloca i32, align 8
  %slot_x_3 = alloca i8, align 8
  %slot___idx_4 = alloca i64, align 8
  %slot___iter_5 = alloca { ptr, i64 }, align 8
  %slot___len_6 = alloca i64, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i32 0, ptr %slot_s_2
  store i64 0, ptr %slot___idx_4
  %t1 = load { ptr, i64 }, ptr %slot_buf_1
  store { ptr, i64 } %t1, ptr %slot___iter_5
  %t3 = getelementptr inbounds { ptr, i64 }, ptr %slot___iter_5, i32 0, i32 1
  %t2 = load i64, ptr %t3
  store i64 %t2, ptr %slot___len_6
  %t4 = load ptr, ptr %slot___iter_5
  br label %fi.cond2
fi.cond2:
  %t5 = load i64, ptr %slot___idx_4
  %t6 = load i64, ptr %slot___len_6
  %t7 = icmp ult i64 %t5, %t6
  br i1 %t7, label %fi.body4, label %for.end1
fi.body4:
  %t8 = getelementptr inbounds i8, ptr %t4, i64 %t5
  %t9 = load i8, ptr %t8
  store i8 %t9, ptr %slot_x_3
  %t10 = load i32, ptr %slot_s_2
  %t11 = load i8, ptr %slot_x_3
  %t12 = zext i8 %t11 to i32
  %t13 = call { i32, i1 } @llvm.sadd.with.overflow.i32(i32 %t10, i32 %t12)
  %t14 = extractvalue { i32, i1 } %t13, 0
  %t15 = extractvalue { i32, i1 } %t13, 1
  br i1 %t15, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 17)
  unreachable
chk.ok5:
  store i32 %t14, ptr %slot_s_2
  br label %fi.step3
fi.step3:
  %t16 = add i64 %t5, 1
  store i64 %t16, ptr %slot___idx_4
  br label %fi.cond2
for.end1:
  %t17 = load i32, ptr %slot_s_2
  ret i32 %t17
}

define void @aila_main__main() {
  %slot_nums_1 = alloca { ptr, i64 }, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_text_3 = alloca { ptr, i64 }, align 8
  %slot_bytes_4 = alloca { ptr, i64 }, align 8
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 4, i64 1)
  %t3 = getelementptr inbounds i8, ptr %t2, i64 0
  store i8 5, ptr %t3
  %t4 = getelementptr inbounds i8, ptr %t2, i64 1
  store i8 5, ptr %t4
  %t5 = getelementptr inbounds i8, ptr %t2, i64 2
  store i8 5, ptr %t5
  %t6 = getelementptr inbounds i8, ptr %t2, i64 3
  store i8 5, ptr %t6
  %t7 = insertvalue { ptr, i64 } undef, ptr %t2, 0
  %t8 = insertvalue { ptr, i64 } %t7, i64 4, 1
  store { ptr, i64 } %t8, ptr %slot_nums_1
  %t9 = load { ptr, i64 }, ptr %slot_nums_1
  %t10 = extractvalue { ptr, i64 } %t9, 0
  %t11 = extractvalue { ptr, i64 } %t9, 1
  %t12 = icmp uge i64 0, %t11
  br i1 %t12, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 26)
  unreachable
chk.ok1:
  %t13 = getelementptr inbounds i8, ptr %t10, i64 0
  store i8 5, ptr %t13
  %t14 = insertvalue { ptr, i64 } undef, ptr @.str.3, 0
  %t15 = insertvalue { ptr, i64 } %t14, i64 6, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t15)
  %t16 = load { ptr, i64 }, ptr %slot_nums_1
  %t17 = call i32 @aila_main__sum_of({ ptr, i64 } %t16)
  call void @aila_std_io__print_i32_nl(i32 %t17)
  %t18 = call { ptr, i64 } @aila_main__fill(ptr %t1, i64 4, i8 7)
  store { ptr, i64 } %t18, ptr %slot_buf_2
  %t19 = load { ptr, i64 }, ptr %slot_buf_2
  %t20 = extractvalue { ptr, i64 } %t19, 0
  %t21 = extractvalue { ptr, i64 } %t19, 1
  %t22 = icmp uge i64 0, %t21
  br i1 %t22, label %chk.bad4, label %chk.ok3
chk.bad4:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 32)
  unreachable
chk.ok3:
  %t23 = getelementptr inbounds i8, ptr %t20, i64 0
  store i8 7, ptr %t23
  %t24 = insertvalue { ptr, i64 } undef, ptr @.str.4, 0
  %t25 = insertvalue { ptr, i64 } %t24, i64 9, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t25)
  %t26 = load { ptr, i64 }, ptr %slot_buf_2
  %t27 = extractvalue { ptr, i64 } %t26, 0
  %t28 = extractvalue { ptr, i64 } %t26, 1
  %t29 = icmp uge i64 0, %t28
  br i1 %t29, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.1, i64 35)
  unreachable
chk.ok5:
  %t30 = getelementptr inbounds i8, ptr %t27, i64 0
  %t31 = load i8, ptr %t30
  %t32 = zext i8 %t31 to i32
  call void @aila_std_io__print_i32(i32 %t32)
  %t33 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t34 = insertvalue { ptr, i64 } %t33, i64 8, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t34)
  %t35 = load { ptr, i64 }, ptr %slot_buf_2
  %t36 = extractvalue { ptr, i64 } %t35, 1
  call void @aila_std_io__print_usize(i64 %t36)
  %t37 = insertvalue { ptr, i64 } undef, ptr @.str.6, 0
  %t38 = insertvalue { ptr, i64 } %t37, i64 8, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t38)
  %t39 = call { i64, i1 } @llvm.umul.with.overflow.i64(i64 4, i64 4)
  %t40 = extractvalue { i64, i1 } %t39, 0
  %t41 = extractvalue { i64, i1 } %t39, 1
  br i1 %t41, label %chk.bad8, label %chk.ok7
chk.bad8:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 39)
  unreachable
chk.ok7:
  call void @aila_std_io__print_usize(i64 %t40)
  %t42 = insertvalue { ptr, i64 } undef, ptr @.str.7, 0
  %t43 = insertvalue { ptr, i64 } %t42, i64 1, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t43)
  call void @aila_std_io__print_nl()
  %t44 = insertvalue { ptr, i64 } undef, ptr @.str.8, 0
  %t45 = insertvalue { ptr, i64 } %t44, i64 4, 1
  store { ptr, i64 } %t45, ptr %slot_text_3
  %t46 = load { ptr, i64 }, ptr %slot_text_3
  store { ptr, i64 } %t46, ptr %slot_bytes_4
  %t47 = load { ptr, i64 }, ptr %slot_bytes_4
  %t48 = extractvalue { ptr, i64 } %t47, 1
  %t49 = trunc i64 %t48 to i32
  call void @aila_std_io__print_i32_nl(i32 %t49)
  call void @aila_rt_region_leave(ptr %t1)
  ret void
}

define i32 @main() {
entry:
  call void @aila_main__main()
  ret i32 0
}
