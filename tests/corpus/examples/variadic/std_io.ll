; AILA module: std.io
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [5 x i8] c"true\00", align 1
@.str.1 = private unnamed_addr constant [6 x i8] c"false\00", align 1
@.str.2 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.3 = private unnamed_addr constant [12 x i8] c"std/io.aila\00", align 1
@.str.4 = private unnamed_addr constant [28 x i8] c"match \E6\9C\AA\E8\A6\86\E7\9B\96\E7\9A\84\E5\88\A4\E5\88\AB\E5\80\BC\00", align 1
@.str.5 = private unnamed_addr constant [44 x i8] c"printf: %d \E9\9C\80\E8\A6\81\E6\95\B4\E6\95\B0\E5\AE\9E\E5\8F\82\EF\BC\88FmtVal.I\EF\BC\89\00", align 1
@.str.6 = private unnamed_addr constant [47 x i8] c"printf: %u \E9\9C\80\E8\A6\81\E6\97\A0\E7\AC\A6\E5\8F\B7\E5\AE\9E\E5\8F\82\EF\BC\88FmtVal.U\EF\BC\89\00", align 1
@.str.7 = private unnamed_addr constant [44 x i8] c"printf: %f \E9\9C\80\E8\A6\81\E6\B5\AE\E7\82\B9\E5\AE\9E\E5\8F\82\EF\BC\88FmtVal.F\EF\BC\89\00", align 1
@.str.8 = private unnamed_addr constant [44 x i8] c"printf: %b \E9\9C\80\E8\A6\81\E5\B8\83\E5\B0\94\E5\AE\9E\E5\8F\82\EF\BC\88FmtVal.B\EF\BC\89\00", align 1
@.str.9 = private unnamed_addr constant [44 x i8] c"printf: %c \E9\9C\80\E8\A6\81\E5\AD\97\E7\AC\A6\E5\AE\9E\E5\8F\82\EF\BC\88FmtVal.C\EF\BC\89\00", align 1
@.str.10 = private unnamed_addr constant [44 x i8] c"printf: %s \E9\9C\80\E8\A6\81\E6\96\87\E6\9C\AC\E5\AE\9E\E5\8F\82\EF\BC\88FmtVal.T\EF\BC\89\00", align 1
@.str.11 = private unnamed_addr constant [57 x i8] c"printf: \E6\9C\AA\E7\9F\A5\E6\A0\BC\E5\BC\8F\E7\AC\A6\EF\BC\88\E6\94\AF\E6\8C\81 %d %u %f %b %c %s %%\EF\BC\89\00", align 1

declare void @aila_rt_write(ptr, i64)
declare ptr @aila_rt_region_enter(i64)
declare ptr @aila_rt_region_alloc(ptr, i64, i64)
declare i64 @aila_std_core__fill_u64(i64, { ptr, i64 })
declare i64 @aila_std_core__fill_i64(i64, { ptr, i64 })
declare i64 @aila_std_core__fill_f64(double, { ptr, i64 })
declare i64 @aila_std_utf8__encode(i32, { ptr, i64 })
declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare void @aila_rt_flush()
declare void @aila_rt_region_leave(ptr)

define i64 @aila_std_io__stdout_write({ ptr, i64 } %p_ctx, { ptr, i64 } %p_bytes, i64 %p_n) {
  %slot_ctx_1 = alloca { ptr, i64 }, align 8
  %slot_bytes_2 = alloca { ptr, i64 }, align 8
  %slot_n_3 = alloca i64, align 8
  store { ptr, i64 } %p_ctx, ptr %slot_ctx_1
  store { ptr, i64 } %p_bytes, ptr %slot_bytes_2
  store i64 %p_n, ptr %slot_n_3
  %t1 = load { ptr, i64 }, ptr %slot_bytes_2
  %t2 = extractvalue { ptr, i64 } %t1, 0
  %t3 = load i64, ptr %slot_n_3
  call void @aila_rt_write(ptr %t2, i64 %t3)
  %t4 = load i64, ptr %slot_n_3
  ret i64 %t4
}

define { ptr, { ptr, i64 } } @aila_std_io__stdout() {
  %slot___lit_1 = alloca { ptr, { ptr, i64 } }, align 8
  %t1 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %slot___lit_1, i32 0, i32 0
  store ptr @aila_std_io__stdout_write, ptr %t1
  %t2 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %slot___lit_1, i32 0, i32 1
  store { ptr, i64 } zeroinitializer, ptr %t2
  %t3 = load { ptr, { ptr, i64 } }, ptr %slot___lit_1
  ret { ptr, { ptr, i64 } } %t3
}

define i64 @aila_std_io__Writer__put(ptr %p_self, { ptr, i64 } %p_bytes) {
  %slot_self_1 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_bytes_2 = alloca { ptr, i64 }, align 8
  store { ptr, i64 } %p_bytes, ptr %slot_bytes_2
  %t1 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %p_self, i32 0, i32 1
  %t2 = load { ptr, i64 }, ptr %t1
  %t3 = load { ptr, i64 }, ptr %slot_bytes_2
  %t4 = load { ptr, i64 }, ptr %slot_bytes_2
  %t5 = extractvalue { ptr, i64 } %t4, 1
  %t6 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %p_self, i32 0, i32 0
  %t7 = load ptr, ptr %t6
  %t8 = call i64 %t7({ ptr, i64 } %t2, { ptr, i64 } %t3, i64 %t5)
  ret i64 %t8
}

define i64 @aila_std_io__Writer__put_text(ptr %p_self, { ptr, i64 } %p_t) {
  %slot_self_1 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_t_2 = alloca { ptr, i64 }, align 8
  %slot_bytes_3 = alloca { ptr, i64 }, align 8
  store { ptr, i64 } %p_t, ptr %slot_t_2
  %t1 = load { ptr, i64 }, ptr %slot_t_2
  store { ptr, i64 } %t1, ptr %slot_bytes_3
  %t2 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %p_self, i32 0, i32 1
  %t3 = load { ptr, i64 }, ptr %t2
  %t4 = load { ptr, i64 }, ptr %slot_bytes_3
  %t5 = load { ptr, i64 }, ptr %slot_bytes_3
  %t6 = extractvalue { ptr, i64 } %t5, 1
  %t7 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %p_self, i32 0, i32 0
  %t8 = load ptr, ptr %t7
  %t9 = call i64 %t8({ ptr, i64 } %t3, { ptr, i64 } %t4, i64 %t6)
  ret i64 %t9
}

define i64 @aila_std_io__Writer__put_u64(ptr %p_self, i64 %p_v) {
  %slot_self_1 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_v_2 = alloca i64, align 8
  %slot_buf_3 = alloca { ptr, i64 }, align 8
  %slot_n_4 = alloca i64, align 8
  store i64 %p_v, ptr %slot_v_2
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 24, i64 1)
  %t3 = insertvalue { ptr, i64 } undef, ptr %t2, 0
  %t4 = insertvalue { ptr, i64 } %t3, i64 24, 1
  store { ptr, i64 } %t4, ptr %slot_buf_3
  %t5 = load i64, ptr %slot_v_2
  %t6 = load { ptr, i64 }, ptr %slot_buf_3
  %t7 = call i64 @aila_std_core__fill_u64(i64 %t5, { ptr, i64 } %t6)
  store i64 %t7, ptr %slot_n_4
  %t8 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %p_self, i32 0, i32 1
  %t9 = load { ptr, i64 }, ptr %t8
  %t10 = load { ptr, i64 }, ptr %slot_buf_3
  %t11 = load i64, ptr %slot_n_4
  %t12 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %p_self, i32 0, i32 0
  %t13 = load ptr, ptr %t12
  %t14 = call i64 %t13({ ptr, i64 } %t9, { ptr, i64 } %t10, i64 %t11)
  ret i64 %t14
}

define i64 @aila_std_io__Writer__put_i64(ptr %p_self, i64 %p_v) {
  %slot_self_1 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_v_2 = alloca i64, align 8
  %slot_buf_3 = alloca { ptr, i64 }, align 8
  %slot_n_4 = alloca i64, align 8
  store i64 %p_v, ptr %slot_v_2
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 24, i64 1)
  %t3 = insertvalue { ptr, i64 } undef, ptr %t2, 0
  %t4 = insertvalue { ptr, i64 } %t3, i64 24, 1
  store { ptr, i64 } %t4, ptr %slot_buf_3
  %t5 = load i64, ptr %slot_v_2
  %t6 = load { ptr, i64 }, ptr %slot_buf_3
  %t7 = call i64 @aila_std_core__fill_i64(i64 %t5, { ptr, i64 } %t6)
  store i64 %t7, ptr %slot_n_4
  %t8 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %p_self, i32 0, i32 1
  %t9 = load { ptr, i64 }, ptr %t8
  %t10 = load { ptr, i64 }, ptr %slot_buf_3
  %t11 = load i64, ptr %slot_n_4
  %t12 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %p_self, i32 0, i32 0
  %t13 = load ptr, ptr %t12
  %t14 = call i64 %t13({ ptr, i64 } %t9, { ptr, i64 } %t10, i64 %t11)
  ret i64 %t14
}

define i64 @aila_std_io__Writer__put_f64(ptr %p_self, double %p_v) {
  %slot_self_1 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_v_2 = alloca double, align 8
  %slot_buf_3 = alloca { ptr, i64 }, align 8
  %slot_n_4 = alloca i64, align 8
  store double %p_v, ptr %slot_v_2
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 40, i64 1)
  %t3 = insertvalue { ptr, i64 } undef, ptr %t2, 0
  %t4 = insertvalue { ptr, i64 } %t3, i64 40, 1
  store { ptr, i64 } %t4, ptr %slot_buf_3
  %t5 = load double, ptr %slot_v_2
  %t6 = load { ptr, i64 }, ptr %slot_buf_3
  %t7 = call i64 @aila_std_core__fill_f64(double %t5, { ptr, i64 } %t6)
  store i64 %t7, ptr %slot_n_4
  %t8 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %p_self, i32 0, i32 1
  %t9 = load { ptr, i64 }, ptr %t8
  %t10 = load { ptr, i64 }, ptr %slot_buf_3
  %t11 = load i64, ptr %slot_n_4
  %t12 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %p_self, i32 0, i32 0
  %t13 = load ptr, ptr %t12
  %t14 = call i64 %t13({ ptr, i64 } %t9, { ptr, i64 } %t10, i64 %t11)
  ret i64 %t14
}

define i64 @aila_std_io__Writer__put_bool(ptr %p_self, i1 %p_v) {
  %slot_self_1 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_v_2 = alloca i1, align 8
  store i1 %p_v, ptr %slot_v_2
  %t1 = load i1, ptr %slot_v_2
  br i1 %t1, label %if.then2, label %if.next3
if.then2:
  %t2 = insertvalue { ptr, i64 } undef, ptr @.str.0, 0
  %t3 = insertvalue { ptr, i64 } %t2, i64 4, 1
  %t4 = call i64 @aila_std_io__Writer__put_text(ptr %p_self, { ptr, i64 } %t3)
  ret i64 %t4
if.next3:
  br label %if.end1
if.end1:
  %t5 = insertvalue { ptr, i64 } undef, ptr @.str.1, 0
  %t6 = insertvalue { ptr, i64 } %t5, i64 5, 1
  %t7 = call i64 @aila_std_io__Writer__put_text(ptr %p_self, { ptr, i64 } %t6)
  ret i64 %t7
}

define i64 @aila_std_io__Writer__put_char(ptr %p_self, i32 %p_c) {
  %slot_self_1 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_c_2 = alloca i32, align 8
  %slot_buf_3 = alloca { ptr, i64 }, align 8
  %slot_n_4 = alloca i64, align 8
  store i32 %p_c, ptr %slot_c_2
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 4, i64 1)
  %t3 = insertvalue { ptr, i64 } undef, ptr %t2, 0
  %t4 = insertvalue { ptr, i64 } %t3, i64 4, 1
  store { ptr, i64 } %t4, ptr %slot_buf_3
  %t5 = load i32, ptr %slot_c_2
  %t7 = load { ptr, i64 }, ptr %slot_buf_3
  %t8 = call i64 @aila_std_utf8__encode(i32 %t5, { ptr, i64 } %t7)
  store i64 %t8, ptr %slot_n_4
  %t9 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %p_self, i32 0, i32 1
  %t10 = load { ptr, i64 }, ptr %t9
  %t11 = load { ptr, i64 }, ptr %slot_buf_3
  %t12 = load i64, ptr %slot_n_4
  %t13 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %p_self, i32 0, i32 0
  %t14 = load ptr, ptr %t13
  %t15 = call i64 %t14({ ptr, i64 } %t10, { ptr, i64 } %t11, i64 %t12)
  ret i64 %t15
}

define i64 @aila_std_io__Writer__put_nl(ptr %p_self) {
  %slot_self_1 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_n_3 = alloca i64, align 8
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 1, i64 1)
  %t3 = insertvalue { ptr, i64 } undef, ptr %t2, 0
  %t4 = insertvalue { ptr, i64 } %t3, i64 1, 1
  store { ptr, i64 } %t4, ptr %slot_buf_2
  %t5 = load { ptr, i64 }, ptr %slot_buf_2
  %t6 = extractvalue { ptr, i64 } %t5, 0
  %t7 = extractvalue { ptr, i64 } %t5, 1
  %t8 = icmp uge i64 0, %t7
  br i1 %t8, label %chk.bad2, label %chk.ok1
chk.bad2:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.3, i64 140)
  unreachable
chk.ok1:
  %t9 = getelementptr inbounds i8, ptr %t6, i64 0
  store i8 10, ptr %t9
  %t10 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %p_self, i32 0, i32 1
  %t11 = load { ptr, i64 }, ptr %t10
  %t12 = load { ptr, i64 }, ptr %slot_buf_2
  %t13 = getelementptr inbounds { ptr, { ptr, i64 } }, ptr %p_self, i32 0, i32 0
  %t14 = load ptr, ptr %t13
  %t15 = call i64 %t14({ ptr, i64 } %t11, { ptr, i64 } %t12, i64 1)
  store i64 %t15, ptr %slot_n_3
  call void @aila_rt_flush()
  %t16 = load i64, ptr %slot_n_3
  ret i64 %t16
}

define void @aila_std_io__put_val(ptr %p_w, ptr %p_v) {
  %slot_w_1 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_v_2 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_v_3 = alloca i64, align 8
  %slot_v_4 = alloca i64, align 8
  %slot_v_5 = alloca double, align 8
  %slot_v_6 = alloca i1, align 8
  %slot_v_7 = alloca i32, align 8
  %slot_v_8 = alloca { ptr, i64 }, align 8
  %slot___head_9 = alloca { i8, { { ptr, i64 } } }, align 8
  %t1 = load { i8, { { ptr, i64 } } }, ptr %p_v
  store { i8, { { ptr, i64 } } } %t1, ptr %slot___head_9
  %t2 = load { i8, { { ptr, i64 } } }, ptr %slot___head_9
  %t3 = extractvalue { i8, { { ptr, i64 } } } %t2, 0
  %t4 = icmp eq i8 %t3, 0
  br i1 %t4, label %m.arm02, label %m.next8
m.next8:
  %t5 = icmp eq i8 %t3, 1
  br i1 %t5, label %m.arm13, label %m.next9
m.next9:
  %t6 = icmp eq i8 %t3, 2
  br i1 %t6, label %m.arm24, label %m.next10
m.next10:
  %t7 = icmp eq i8 %t3, 3
  br i1 %t7, label %m.arm35, label %m.next11
m.next11:
  %t8 = icmp eq i8 %t3, 4
  br i1 %t8, label %m.arm46, label %m.next12
m.next12:
  %t9 = icmp eq i8 %t3, 5
  br i1 %t9, label %m.arm57, label %m.next13
m.next13:
  call void @aila_rt_trap(ptr @.str.4, i64 27, ptr @.str.3, i64 151)
  unreachable
m.arm02:
  %t10 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___head_9, i32 0, i32 1, i32 0
  %t11 = load i64, ptr %t10
  store i64 %t11, ptr %slot_v_3
  %t12 = load i64, ptr %slot_v_3
  %t13 = call i64 @aila_std_io__Writer__put_i64(ptr %p_w, i64 %t12)
  br label %m.end1
m.arm13:
  %t14 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___head_9, i32 0, i32 1, i32 0
  %t15 = load i64, ptr %t14
  store i64 %t15, ptr %slot_v_4
  %t16 = load i64, ptr %slot_v_4
  %t17 = call i64 @aila_std_io__Writer__put_u64(ptr %p_w, i64 %t16)
  br label %m.end1
m.arm24:
  %t18 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___head_9, i32 0, i32 1, i32 0
  %t19 = load double, ptr %t18
  store double %t19, ptr %slot_v_5
  %t20 = load double, ptr %slot_v_5
  %t21 = call i64 @aila_std_io__Writer__put_f64(ptr %p_w, double %t20)
  br label %m.end1
m.arm35:
  %t22 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___head_9, i32 0, i32 1, i32 0
  %t23 = load i1, ptr %t22
  store i1 %t23, ptr %slot_v_6
  %t24 = load i1, ptr %slot_v_6
  %t25 = call i64 @aila_std_io__Writer__put_bool(ptr %p_w, i1 %t24)
  br label %m.end1
m.arm46:
  %t26 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___head_9, i32 0, i32 1, i32 0
  %t27 = load i32, ptr %t26
  store i32 %t27, ptr %slot_v_7
  %t28 = load i32, ptr %slot_v_7
  %t29 = call i64 @aila_std_io__Writer__put_char(ptr %p_w, i32 %t28)
  br label %m.end1
m.arm57:
  %t30 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___head_9, i32 0, i32 1, i32 0
  %t31 = load { ptr, i64 }, ptr %t30
  store { ptr, i64 } %t31, ptr %slot_v_8
  %t32 = load { ptr, i64 }, ptr %slot_v_8
  %t33 = call i64 @aila_std_io__Writer__put_text(ptr %p_w, { ptr, i64 } %t32)
  br label %m.end1
m.end1:
  ret void
}

define void @aila_std_io__print({ ptr, i64 } %p_vals) {
  %slot_vals_1 = alloca { ptr, i64 }, align 8
  %slot_w_2 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_v_3 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___idx_4 = alloca i64, align 8
  %slot___iter_5 = alloca { ptr, i64 }, align 8
  %slot___len_6 = alloca i64, align 8
  store { ptr, i64 } %p_vals, ptr %slot_vals_1
  %t1 = call { ptr, { ptr, i64 } } @aila_std_io__stdout()
  store { ptr, { ptr, i64 } } %t1, ptr %slot_w_2
  store i64 0, ptr %slot___idx_4
  %t2 = load { ptr, i64 }, ptr %slot_vals_1
  store { ptr, i64 } %t2, ptr %slot___iter_5
  %t4 = getelementptr inbounds { ptr, i64 }, ptr %slot___iter_5, i32 0, i32 1
  %t3 = load i64, ptr %t4
  store i64 %t3, ptr %slot___len_6
  %t5 = load ptr, ptr %slot___iter_5
  br label %fi.cond2
fi.cond2:
  %t6 = load i64, ptr %slot___idx_4
  %t7 = load i64, ptr %slot___len_6
  %t8 = icmp ult i64 %t6, %t7
  br i1 %t8, label %fi.body4, label %for.end1
fi.body4:
  %t9 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %t5, i64 %t6
  %t10 = load { i8, { { ptr, i64 } } }, ptr %t9
  store { i8, { { ptr, i64 } } } %t10, ptr %slot_v_3
  call void @aila_std_io__put_val(ptr %slot_w_2, ptr %slot_v_3)
  br label %fi.step3
fi.step3:
  %t11 = add i64 %t6, 1
  store i64 %t11, ptr %slot___idx_4
  br label %fi.cond2
for.end1:
  ret void
}

define void @aila_std_io__println({ ptr, i64 } %p_vals) {
  %slot_vals_1 = alloca { ptr, i64 }, align 8
  %slot_w_2 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_v_3 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___idx_4 = alloca i64, align 8
  %slot___iter_5 = alloca { ptr, i64 }, align 8
  %slot___len_6 = alloca i64, align 8
  store { ptr, i64 } %p_vals, ptr %slot_vals_1
  %t1 = call { ptr, { ptr, i64 } } @aila_std_io__stdout()
  store { ptr, { ptr, i64 } } %t1, ptr %slot_w_2
  store i64 0, ptr %slot___idx_4
  %t2 = load { ptr, i64 }, ptr %slot_vals_1
  store { ptr, i64 } %t2, ptr %slot___iter_5
  %t4 = getelementptr inbounds { ptr, i64 }, ptr %slot___iter_5, i32 0, i32 1
  %t3 = load i64, ptr %t4
  store i64 %t3, ptr %slot___len_6
  %t5 = load ptr, ptr %slot___iter_5
  br label %fi.cond2
fi.cond2:
  %t6 = load i64, ptr %slot___idx_4
  %t7 = load i64, ptr %slot___len_6
  %t8 = icmp ult i64 %t6, %t7
  br i1 %t8, label %fi.body4, label %for.end1
fi.body4:
  %t9 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %t5, i64 %t6
  %t10 = load { i8, { { ptr, i64 } } }, ptr %t9
  store { i8, { { ptr, i64 } } } %t10, ptr %slot_v_3
  call void @aila_std_io__put_val(ptr %slot_w_2, ptr %slot_v_3)
  br label %fi.step3
fi.step3:
  %t11 = add i64 %t6, 1
  store i64 %t11, ptr %slot___idx_4
  br label %fi.cond2
for.end1:
  %t12 = call i64 @aila_std_io__Writer__put_nl(ptr %slot_w_2)
  ret void
}

define void @aila_std_io__printf({ ptr, i64 } %p_fmt, { ptr, i64 } %p_args) {
  %slot_fmt_1 = alloca { ptr, i64 }, align 8
  %slot_args_2 = alloca { ptr, i64 }, align 8
  %slot_w_3 = alloca { ptr, { ptr, i64 } }, align 8
  %slot_bytes_4 = alloca { ptr, i64 }, align 8
  %slot_n_5 = alloca i64, align 8
  %slot_b_6 = alloca { ptr, i64 }, align 8
  %slot_i_7 = alloca i64, align 8
  %slot_k_8 = alloca i64, align 8
  %slot_c_9 = alloca i8, align 8
  %slot_f_10 = alloca i8, align 8
  %slot_v_11 = alloca i64, align 8
  %slot_v_12 = alloca i64, align 8
  %slot_v_13 = alloca double, align 8
  %slot_v_14 = alloca i1, align 8
  %slot_v_15 = alloca i32, align 8
  %slot_v_16 = alloca { ptr, i64 }, align 8
  %slot___head_17 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___head_18 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___head_19 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___head_20 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___head_21 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___head_22 = alloca { i8, { { ptr, i64 } } }, align 8
  store { ptr, i64 } %p_fmt, ptr %slot_fmt_1
  store { ptr, i64 } %p_args, ptr %slot_args_2
  %t1 = call { ptr, { ptr, i64 } } @aila_std_io__stdout()
  store { ptr, { ptr, i64 } } %t1, ptr %slot_w_3
  %t2 = load { ptr, i64 }, ptr %slot_fmt_1
  store { ptr, i64 } %t2, ptr %slot_bytes_4
  %t3 = load { ptr, i64 }, ptr %slot_bytes_4
  %t4 = extractvalue { ptr, i64 } %t3, 1
  store i64 %t4, ptr %slot_n_5
  %t5 = call ptr @aila_rt_region_enter(i64 0)
  %t6 = call ptr @aila_rt_region_alloc(ptr %t5, i64 1, i64 1)
  %t7 = insertvalue { ptr, i64 } undef, ptr %t6, 0
  %t8 = insertvalue { ptr, i64 } %t7, i64 1, 1
  store { ptr, i64 } %t8, ptr %slot_b_6
  store i64 0, ptr %slot_i_7
  store i64 0, ptr %slot_k_8
  br label %for.cond2
for.cond2:
  %t9 = load i64, ptr %slot_i_7
  %t10 = load i64, ptr %slot_n_5
  %t11 = icmp ult i64 %t9, %t10
  br i1 %t11, label %for.body4, label %for.end1
for.body4:
  %t12 = load i64, ptr %slot_i_7
  %t13 = load { ptr, i64 }, ptr %slot_bytes_4
  %t14 = extractvalue { ptr, i64 } %t13, 0
  %t15 = extractvalue { ptr, i64 } %t13, 1
  %t16 = icmp uge i64 %t12, %t15
  br i1 %t16, label %chk.bad6, label %chk.ok5
chk.bad6:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.3, i64 185)
  unreachable
chk.ok5:
  %t17 = getelementptr inbounds i8, ptr %t14, i64 %t12
  %t18 = load i8, ptr %t17
  store i8 %t18, ptr %slot_c_9
  %t19 = load i8, ptr %slot_c_9
  %t20 = icmp eq i8 %t19, 37
  br i1 %t20, label %if.then8, label %if.next9
if.then8:
  %t21 = load i64, ptr %slot_i_7
  %t22 = add i64 %t21, 1
  store i64 %t22, ptr %slot_i_7
  %t23 = load i64, ptr %slot_i_7
  %t24 = load { ptr, i64 }, ptr %slot_bytes_4
  %t25 = extractvalue { ptr, i64 } %t24, 0
  %t26 = extractvalue { ptr, i64 } %t24, 1
  %t27 = icmp uge i64 %t23, %t26
  br i1 %t27, label %chk.bad11, label %chk.ok10
chk.bad11:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.3, i64 189)
  unreachable
chk.ok10:
  %t28 = getelementptr inbounds i8, ptr %t25, i64 %t23
  %t29 = load i8, ptr %t28
  store i8 %t29, ptr %slot_f_10
  %t30 = load i8, ptr %slot_f_10
  %t31 = icmp eq i8 %t30, 100
  br i1 %t31, label %if.then13, label %if.next14
if.then13:
  %t32 = load i64, ptr %slot_k_8
  %t33 = load { ptr, i64 }, ptr %slot_args_2
  %t34 = extractvalue { ptr, i64 } %t33, 0
  %t35 = extractvalue { ptr, i64 } %t33, 1
  %t36 = icmp uge i64 %t32, %t35
  br i1 %t36, label %chk.bad16, label %chk.ok15
chk.bad16:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.3, i64 191)
  unreachable
chk.ok15:
  %t37 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %t34, i64 %t32
  %t38 = load { i8, { { ptr, i64 } } }, ptr %t37
  store { i8, { { ptr, i64 } } } %t38, ptr %slot___head_17
  %t39 = load { i8, { { ptr, i64 } } }, ptr %slot___head_17
  %t40 = extractvalue { i8, { { ptr, i64 } } } %t39, 0
  %t41 = icmp eq i8 %t40, 0
  br i1 %t41, label %m.arm018, label %m.next20
m.next20:
  br label %m.arm119
m.arm018:
  %t42 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___head_17, i32 0, i32 1, i32 0
  %t43 = load i64, ptr %t42
  store i64 %t43, ptr %slot_v_11
  %t44 = load i64, ptr %slot_v_11
  %t45 = call i64 @aila_std_io__Writer__put_i64(ptr %slot_w_3, i64 %t44)
  br label %m.end17
m.arm119:
  call void @aila_rt_trap(ptr @.str.5, i64 43, ptr @.str.3, i64 193)
  unreachable
m.end17:
  %t46 = load i64, ptr %slot_k_8
  %t47 = add i64 %t46, 1
  store i64 %t47, ptr %slot_k_8
  br label %if.end12
if.next14:
  %t48 = load i8, ptr %slot_f_10
  %t49 = icmp eq i8 %t48, 117
  br i1 %t49, label %if.then21, label %if.next22
if.then21:
  %t50 = load i64, ptr %slot_k_8
  %t51 = load { ptr, i64 }, ptr %slot_args_2
  %t52 = extractvalue { ptr, i64 } %t51, 0
  %t53 = extractvalue { ptr, i64 } %t51, 1
  %t54 = icmp uge i64 %t50, %t53
  br i1 %t54, label %chk.bad24, label %chk.ok23
chk.bad24:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.3, i64 197)
  unreachable
chk.ok23:
  %t55 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %t52, i64 %t50
  %t56 = load { i8, { { ptr, i64 } } }, ptr %t55
  store { i8, { { ptr, i64 } } } %t56, ptr %slot___head_18
  %t57 = load { i8, { { ptr, i64 } } }, ptr %slot___head_18
  %t58 = extractvalue { i8, { { ptr, i64 } } } %t57, 0
  %t59 = icmp eq i8 %t58, 1
  br i1 %t59, label %m.arm026, label %m.next28
m.next28:
  br label %m.arm127
m.arm026:
  %t60 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___head_18, i32 0, i32 1, i32 0
  %t61 = load i64, ptr %t60
  store i64 %t61, ptr %slot_v_12
  %t62 = load i64, ptr %slot_v_12
  %t63 = call i64 @aila_std_io__Writer__put_u64(ptr %slot_w_3, i64 %t62)
  br label %m.end25
m.arm127:
  call void @aila_rt_trap(ptr @.str.6, i64 46, ptr @.str.3, i64 199)
  unreachable
m.end25:
  %t64 = load i64, ptr %slot_k_8
  %t65 = add i64 %t64, 1
  store i64 %t65, ptr %slot_k_8
  br label %if.end12
if.next22:
  %t66 = load i8, ptr %slot_f_10
  %t67 = icmp eq i8 %t66, 102
  br i1 %t67, label %if.then29, label %if.next30
if.then29:
  %t68 = load i64, ptr %slot_k_8
  %t69 = load { ptr, i64 }, ptr %slot_args_2
  %t70 = extractvalue { ptr, i64 } %t69, 0
  %t71 = extractvalue { ptr, i64 } %t69, 1
  %t72 = icmp uge i64 %t68, %t71
  br i1 %t72, label %chk.bad32, label %chk.ok31
chk.bad32:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.3, i64 203)
  unreachable
chk.ok31:
  %t73 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %t70, i64 %t68
  %t74 = load { i8, { { ptr, i64 } } }, ptr %t73
  store { i8, { { ptr, i64 } } } %t74, ptr %slot___head_19
  %t75 = load { i8, { { ptr, i64 } } }, ptr %slot___head_19
  %t76 = extractvalue { i8, { { ptr, i64 } } } %t75, 0
  %t77 = icmp eq i8 %t76, 2
  br i1 %t77, label %m.arm034, label %m.next36
m.next36:
  br label %m.arm135
m.arm034:
  %t78 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___head_19, i32 0, i32 1, i32 0
  %t79 = load double, ptr %t78
  store double %t79, ptr %slot_v_13
  %t80 = load double, ptr %slot_v_13
  %t81 = call i64 @aila_std_io__Writer__put_f64(ptr %slot_w_3, double %t80)
  br label %m.end33
m.arm135:
  call void @aila_rt_trap(ptr @.str.7, i64 43, ptr @.str.3, i64 205)
  unreachable
m.end33:
  %t82 = load i64, ptr %slot_k_8
  %t83 = add i64 %t82, 1
  store i64 %t83, ptr %slot_k_8
  br label %if.end12
if.next30:
  %t84 = load i8, ptr %slot_f_10
  %t85 = icmp eq i8 %t84, 98
  br i1 %t85, label %if.then37, label %if.next38
if.then37:
  %t86 = load i64, ptr %slot_k_8
  %t87 = load { ptr, i64 }, ptr %slot_args_2
  %t88 = extractvalue { ptr, i64 } %t87, 0
  %t89 = extractvalue { ptr, i64 } %t87, 1
  %t90 = icmp uge i64 %t86, %t89
  br i1 %t90, label %chk.bad40, label %chk.ok39
chk.bad40:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.3, i64 209)
  unreachable
chk.ok39:
  %t91 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %t88, i64 %t86
  %t92 = load { i8, { { ptr, i64 } } }, ptr %t91
  store { i8, { { ptr, i64 } } } %t92, ptr %slot___head_20
  %t93 = load { i8, { { ptr, i64 } } }, ptr %slot___head_20
  %t94 = extractvalue { i8, { { ptr, i64 } } } %t93, 0
  %t95 = icmp eq i8 %t94, 3
  br i1 %t95, label %m.arm042, label %m.next44
m.next44:
  br label %m.arm143
m.arm042:
  %t96 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___head_20, i32 0, i32 1, i32 0
  %t97 = load i1, ptr %t96
  store i1 %t97, ptr %slot_v_14
  %t98 = load i1, ptr %slot_v_14
  %t99 = call i64 @aila_std_io__Writer__put_bool(ptr %slot_w_3, i1 %t98)
  br label %m.end41
m.arm143:
  call void @aila_rt_trap(ptr @.str.8, i64 43, ptr @.str.3, i64 211)
  unreachable
m.end41:
  %t100 = load i64, ptr %slot_k_8
  %t101 = add i64 %t100, 1
  store i64 %t101, ptr %slot_k_8
  br label %if.end12
if.next38:
  %t102 = load i8, ptr %slot_f_10
  %t103 = icmp eq i8 %t102, 99
  br i1 %t103, label %if.then45, label %if.next46
if.then45:
  %t104 = load i64, ptr %slot_k_8
  %t105 = load { ptr, i64 }, ptr %slot_args_2
  %t106 = extractvalue { ptr, i64 } %t105, 0
  %t107 = extractvalue { ptr, i64 } %t105, 1
  %t108 = icmp uge i64 %t104, %t107
  br i1 %t108, label %chk.bad48, label %chk.ok47
chk.bad48:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.3, i64 215)
  unreachable
chk.ok47:
  %t109 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %t106, i64 %t104
  %t110 = load { i8, { { ptr, i64 } } }, ptr %t109
  store { i8, { { ptr, i64 } } } %t110, ptr %slot___head_21
  %t111 = load { i8, { { ptr, i64 } } }, ptr %slot___head_21
  %t112 = extractvalue { i8, { { ptr, i64 } } } %t111, 0
  %t113 = icmp eq i8 %t112, 4
  br i1 %t113, label %m.arm050, label %m.next52
m.next52:
  br label %m.arm151
m.arm050:
  %t114 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___head_21, i32 0, i32 1, i32 0
  %t115 = load i32, ptr %t114
  store i32 %t115, ptr %slot_v_15
  %t116 = load i32, ptr %slot_v_15
  %t117 = call i64 @aila_std_io__Writer__put_char(ptr %slot_w_3, i32 %t116)
  br label %m.end49
m.arm151:
  call void @aila_rt_trap(ptr @.str.9, i64 43, ptr @.str.3, i64 217)
  unreachable
m.end49:
  %t118 = load i64, ptr %slot_k_8
  %t119 = add i64 %t118, 1
  store i64 %t119, ptr %slot_k_8
  br label %if.end12
if.next46:
  %t120 = load i8, ptr %slot_f_10
  %t121 = icmp eq i8 %t120, 115
  br i1 %t121, label %if.then53, label %if.next54
if.then53:
  %t122 = load i64, ptr %slot_k_8
  %t123 = load { ptr, i64 }, ptr %slot_args_2
  %t124 = extractvalue { ptr, i64 } %t123, 0
  %t125 = extractvalue { ptr, i64 } %t123, 1
  %t126 = icmp uge i64 %t122, %t125
  br i1 %t126, label %chk.bad56, label %chk.ok55
chk.bad56:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.3, i64 221)
  unreachable
chk.ok55:
  %t127 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %t124, i64 %t122
  %t128 = load { i8, { { ptr, i64 } } }, ptr %t127
  store { i8, { { ptr, i64 } } } %t128, ptr %slot___head_22
  %t129 = load { i8, { { ptr, i64 } } }, ptr %slot___head_22
  %t130 = extractvalue { i8, { { ptr, i64 } } } %t129, 0
  %t131 = icmp eq i8 %t130, 5
  br i1 %t131, label %m.arm058, label %m.next60
m.next60:
  br label %m.arm159
m.arm058:
  %t132 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___head_22, i32 0, i32 1, i32 0
  %t133 = load { ptr, i64 }, ptr %t132
  store { ptr, i64 } %t133, ptr %slot_v_16
  %t134 = load { ptr, i64 }, ptr %slot_v_16
  %t135 = call i64 @aila_std_io__Writer__put_text(ptr %slot_w_3, { ptr, i64 } %t134)
  br label %m.end57
m.arm159:
  call void @aila_rt_trap(ptr @.str.10, i64 43, ptr @.str.3, i64 223)
  unreachable
m.end57:
  %t136 = load i64, ptr %slot_k_8
  %t137 = add i64 %t136, 1
  store i64 %t137, ptr %slot_k_8
  br label %if.end12
if.next54:
  %t138 = load i8, ptr %slot_f_10
  %t139 = icmp eq i8 %t138, 37
  br i1 %t139, label %if.then61, label %if.next62
if.then61:
  %t140 = load { ptr, i64 }, ptr %slot_b_6
  %t141 = extractvalue { ptr, i64 } %t140, 0
  %t142 = extractvalue { ptr, i64 } %t140, 1
  %t143 = icmp uge i64 0, %t142
  br i1 %t143, label %chk.bad64, label %chk.ok63
chk.bad64:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.3, i64 227)
  unreachable
chk.ok63:
  %t144 = getelementptr inbounds i8, ptr %t141, i64 0
  store i8 37, ptr %t144
  %t145 = load { ptr, i64 }, ptr %slot_b_6
  %t146 = call i64 @aila_std_io__Writer__put(ptr %slot_w_3, { ptr, i64 } %t145)
  br label %if.end12
if.next62:
  call void @aila_rt_trap(ptr @.str.11, i64 56, ptr @.str.3, i64 230)
  unreachable
if.end12:
  %t147 = load i64, ptr %slot_i_7
  %t148 = add i64 %t147, 1
  store i64 %t148, ptr %slot_i_7
  br label %if.end7
if.next9:
  %t149 = load { ptr, i64 }, ptr %slot_b_6
  %t150 = extractvalue { ptr, i64 } %t149, 0
  %t151 = extractvalue { ptr, i64 } %t149, 1
  %t152 = icmp uge i64 0, %t151
  br i1 %t152, label %chk.bad66, label %chk.ok65
chk.bad66:
  call void @aila_rt_trap(ptr @.str.2, i64 19, ptr @.str.3, i64 234)
  unreachable
chk.ok65:
  %t153 = getelementptr inbounds i8, ptr %t150, i64 0
  %t154 = load i8, ptr %slot_c_9
  store i8 %t154, ptr %t153
  %t155 = load { ptr, i64 }, ptr %slot_b_6
  %t156 = call i64 @aila_std_io__Writer__put(ptr %slot_w_3, { ptr, i64 } %t155)
  %t157 = load i64, ptr %slot_i_7
  %t158 = add i64 %t157, 1
  store i64 %t158, ptr %slot_i_7
  br label %if.end7
if.end7:
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  call void @aila_rt_region_leave(ptr %t5)
  ret void
}
