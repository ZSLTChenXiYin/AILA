; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [17 x i8] c"integer overflow\00", align 1
@.str.1 = private unnamed_addr constant [10 x i8] c"main.aila\00", align 1
@.str.2 = private unnamed_addr constant [17 x i8] c"Detection{label=\00", align 1
@.str.3 = private unnamed_addr constant [8 x i8] c" score=\00", align 1
@.str.4 = private unnamed_addr constant [8 x i8] c" bbox=[\00", align 1
@.str.5 = private unnamed_addr constant [3 x i8] c", \00", align 1
@.str.6 = private unnamed_addr constant [2 x i8] c"]\00", align 1
@.str.7 = private unnamed_addr constant [15 x i8] c"norm2(3, 4) = \00", align 1
@.str.8 = private unnamed_addr constant [20 x i8] c"#size(Detection) = \00", align 1
@.str.9 = private unnamed_addr constant [12 x i8] c"  #align = \00", align 1

declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare void @aila_std_io__print_text({ ptr, i64 })
declare void @aila_std_io__print_i32(i32)
declare void @aila_std_io__print_f32(float)
declare void @aila_std_io__print_nl()
declare void @aila_std_io__print_i32_nl(i32)
declare void @aila_std_io__print_usize(i64)

define i32 @aila_main__Detection__norm2({ i32, float, [4 x float] } %p_self) {
  %slot_self_1 = alloca { i32, float, [4 x float] }, align 8
  store { i32, float, [4 x float] } %p_self, ptr %slot_self_1
  %t1 = getelementptr inbounds { i32, float, [4 x float] }, ptr %slot_self_1, i32 0, i32 0
  %t2 = load i32, ptr %t1
  %t3 = getelementptr inbounds { i32, float, [4 x float] }, ptr %slot_self_1, i32 0, i32 0
  %t4 = load i32, ptr %t3
  %t5 = call { i32, i1 } @llvm.smul.with.overflow.i32(i32 %t2, i32 %t4)
  %t6 = extractvalue { i32, i1 } %t5, 0
  %t7 = extractvalue { i32, i1 } %t5, 1
  br i1 %t7, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 16)
  unreachable
chk.ok1:
  %t8 = getelementptr inbounds { i32, float, [4 x float] }, ptr %slot_self_1, i32 0, i32 2
  %t9 = getelementptr inbounds [4 x float], ptr %t8, i64 0, i64 0
  %t10 = load float, ptr %t9
  %t11 = fptosi float %t10 to i32
  %t12 = getelementptr inbounds { i32, float, [4 x float] }, ptr %slot_self_1, i32 0, i32 2
  %t13 = getelementptr inbounds [4 x float], ptr %t12, i64 0, i64 0
  %t14 = load float, ptr %t13
  %t15 = fptosi float %t14 to i32
  %t16 = call { i32, i1 } @llvm.smul.with.overflow.i32(i32 %t11, i32 %t15)
  %t17 = extractvalue { i32, i1 } %t16, 0
  %t18 = extractvalue { i32, i1 } %t16, 1
  br i1 %t18, label %chk.bad4, label %chk.ok3
chk.bad4:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 16)
  unreachable
chk.ok3:
  %t19 = call { i32, i1 } @llvm.sadd.with.overflow.i32(i32 %t6, i32 %t17)
  %t20 = extractvalue { i32, i1 } %t19, 0
  %t21 = extractvalue { i32, i1 } %t19, 1
  br i1 %t21, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.0, i64 16, ptr @.str.1, i64 16)
  unreachable
chk.ok5:
  ret i32 %t20
}

define void @aila_main__Detection__describe({ i32, float, [4 x float] } %p_self) {
  %slot_self_1 = alloca { i32, float, [4 x float] }, align 8
  %slot_i_2 = alloca i64, align 8
  %slot_v_3 = alloca float, align 8
  %slot___idx_4 = alloca i64, align 8
  store { i32, float, [4 x float] } %p_self, ptr %slot_self_1
  %t1 = insertvalue { ptr, i64 } undef, ptr @.str.2, 0
  %t2 = insertvalue { ptr, i64 } %t1, i64 16, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t2)
  %t3 = getelementptr inbounds { i32, float, [4 x float] }, ptr %slot_self_1, i32 0, i32 0
  %t4 = load i32, ptr %t3
  call void @aila_std_io__print_i32(i32 %t4)
  %t5 = insertvalue { ptr, i64 } undef, ptr @.str.3, 0
  %t6 = insertvalue { ptr, i64 } %t5, i64 7, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t6)
  %t7 = getelementptr inbounds { i32, float, [4 x float] }, ptr %slot_self_1, i32 0, i32 1
  %t8 = load float, ptr %t7
  call void @aila_std_io__print_f32(float %t8)
  %t9 = insertvalue { ptr, i64 } undef, ptr @.str.4, 0
  %t10 = insertvalue { ptr, i64 } %t9, i64 7, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t10)
  store i64 0, ptr %slot___idx_4
  %t11 = getelementptr inbounds { i32, float, [4 x float] }, ptr %slot_self_1, i32 0, i32 2
  br label %fi.cond2
fi.cond2:
  %t12 = load i64, ptr %slot___idx_4
  %t13 = icmp ult i64 %t12, 4
  br i1 %t13, label %fi.body4, label %for.end1
fi.body4:
  %t14 = getelementptr inbounds [4 x float], ptr %t11, i64 0, i64 %t12
  %t15 = load float, ptr %t14
  store float %t15, ptr %slot_v_3
  store i64 %t12, ptr %slot_i_2
  %t16 = load i64, ptr %slot_i_2
  %t17 = icmp ugt i64 %t16, 0
  br i1 %t17, label %if.then6, label %if.next7
if.then6:
  %t18 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t19 = insertvalue { ptr, i64 } %t18, i64 2, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t19)
  br label %if.end5
if.next7:
  br label %if.end5
if.end5:
  %t20 = load float, ptr %slot_v_3
  %t21 = fptosi float %t20 to i32
  call void @aila_std_io__print_i32(i32 %t21)
  br label %fi.step3
fi.step3:
  %t22 = add i64 %t12, 1
  store i64 %t22, ptr %slot___idx_4
  br label %fi.cond2
for.end1:
  %t23 = insertvalue { ptr, i64 } undef, ptr @.str.6, 0
  %t24 = insertvalue { ptr, i64 } %t23, i64 1, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t24)
  call void @aila_std_io__print_nl()
  ret void
}

define void @aila_main__main() {
  %slot_d_1 = alloca { i32, float, [4 x float] }, align 8
  %slot___lit_2 = alloca { i32, float, [4 x float] }, align 8
  %slot___arr_3 = alloca [4 x float], align 8
  %t1 = getelementptr inbounds { i32, float, [4 x float] }, ptr %slot___lit_2, i32 0, i32 0
  store i32 3, ptr %t1
  %t2 = getelementptr inbounds { i32, float, [4 x float] }, ptr %slot___lit_2, i32 0, i32 1
  store float 0.75, ptr %t2
  %t3 = getelementptr inbounds [4 x float], ptr %slot___arr_3, i64 0, i64 0
  store float 1.0, ptr %t3
  %t4 = getelementptr inbounds [4 x float], ptr %slot___arr_3, i64 0, i64 1
  store float 2.0, ptr %t4
  %t5 = getelementptr inbounds [4 x float], ptr %slot___arr_3, i64 0, i64 2
  store float 3.0, ptr %t5
  %t6 = getelementptr inbounds [4 x float], ptr %slot___arr_3, i64 0, i64 3
  store float 4.0, ptr %t6
  %t7 = load [4 x float], ptr %slot___arr_3
  %t8 = getelementptr inbounds { i32, float, [4 x float] }, ptr %slot___lit_2, i32 0, i32 2
  store [4 x float] %t7, ptr %t8
  %t9 = load { i32, float, [4 x float] }, ptr %slot___lit_2
  store { i32, float, [4 x float] } %t9, ptr %slot_d_1
  %t10 = insertvalue { ptr, i64 } undef, ptr @.str.7, 0
  %t11 = insertvalue { ptr, i64 } %t10, i64 14, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t11)
  %t12 = load { i32, float, [4 x float] }, ptr %slot_d_1
  %t13 = call i32 @aila_main__Detection__norm2({ i32, float, [4 x float] } %t12)
  call void @aila_std_io__print_i32_nl(i32 %t13)
  %t14 = load { i32, float, [4 x float] }, ptr %slot_d_1
  call void @aila_main__Detection__describe({ i32, float, [4 x float] } %t14)
  %t15 = insertvalue { ptr, i64 } undef, ptr @.str.8, 0
  %t16 = insertvalue { ptr, i64 } %t15, i64 19, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t16)
  call void @aila_std_io__print_usize(i64 24)
  %t17 = insertvalue { ptr, i64 } undef, ptr @.str.9, 0
  %t18 = insertvalue { ptr, i64 } %t17, i64 11, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t18)
  call void @aila_std_io__print_usize(i64 4)
  call void @aila_std_io__print_nl()
  ret void
}

define i32 @main() {
entry:
  call void @aila_main__main()
  ret i32 0
}
