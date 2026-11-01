; AILA module: std.os
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"



declare i64 @aila_rt_file_open(ptr, i64, i32)
declare i64 @aila_rt_file_read(i64, ptr, i64)
declare i64 @aila_rt_file_write(i64, ptr, i64)
declare void @aila_rt_file_close(i64)
declare void @aila_rt_file_flush(i64)
declare void @aila_rt_exit(i32)
declare i64 @aila_rt_env_get(ptr, i64, ptr, i64)
declare i64 @aila_rt_env_set(ptr, i64, ptr, i64)
declare i64 @aila_rt_args_count()
declare i64 @aila_rt_args_get(i64, ptr, i64)

define { i1, { i64 }, { i8, { i64 } } } @aila_std_os__open({ ptr, i64 } %p_path, i32 %p_mode) {
  %slot_path_1 = alloca { ptr, i64 }, align 8
  %slot_mode_2 = alloca i32, align 8
  %slot_pb_3 = alloca { ptr, i64 }, align 8
  %slot_h_4 = alloca i64, align 8
  %slot___lit_5 = alloca { i8, { i64 } }, align 8
  %slot___lit_6 = alloca { i64 }, align 8
  store { ptr, i64 } %p_path, ptr %slot_path_1
  store i32 %p_mode, ptr %slot_mode_2
  %t1 = load { ptr, i64 }, ptr %slot_path_1
  store { ptr, i64 } %t1, ptr %slot_pb_3
  %t2 = load { ptr, i64 }, ptr %slot_pb_3
  %t3 = extractvalue { ptr, i64 } %t2, 0
  %t4 = load { ptr, i64 }, ptr %slot_pb_3
  %t5 = extractvalue { ptr, i64 } %t4, 1
  %t6 = load i32, ptr %slot_mode_2
  %t7 = call i64 @aila_rt_file_open(ptr %t3, i64 %t5, i32 %t6)
  store i64 %t7, ptr %slot_h_4
  %t8 = load i64, ptr %slot_h_4
  %t9 = icmp eq i64 %t8, 0
  br i1 %t9, label %if.then2, label %if.next3
if.then2:
  %t10 = insertvalue { i8, { i64 } } undef, i8 3, 0
  %t11 = insertvalue { i1, { i64 }, { i8, { i64 } } } undef, i1 0, 0
  %t12 = insertvalue { i1, { i64 }, { i8, { i64 } } } %t11, { i8, { i64 } } %t10, 2
  ret { i1, { i64 }, { i8, { i64 } } } %t12
if.next3:
  br label %if.end1
if.end1:
  %t13 = load i64, ptr %slot_h_4
  %t14 = getelementptr inbounds { i64 }, ptr %slot___lit_6, i32 0, i32 0
  store i64 %t13, ptr %t14
  %t15 = load { i64 }, ptr %slot___lit_6
  %t16 = insertvalue { i1, { i64 }, { i8, { i64 } } } undef, i1 1, 0
  %t17 = insertvalue { i1, { i64 }, { i8, { i64 } } } %t16, { i64 } %t15, 1
  ret { i1, { i64 }, { i8, { i64 } } } %t17
}

define { i1, i64, { i8, { i64 } } } @aila_std_os__read({ i64 } %p_f, { ptr, i64 } %p_buf) {
  %slot_f_1 = alloca { i64 }, align 8
  %slot_buf_2 = alloca { ptr, i64 }, align 8
  %slot_n_3 = alloca i64, align 8
  %slot___lit_4 = alloca { i8, { i64 } }, align 8
  store { i64 } %p_f, ptr %slot_f_1
  store { ptr, i64 } %p_buf, ptr %slot_buf_2
  %t1 = getelementptr inbounds { i64 }, ptr %slot_f_1, i32 0, i32 0
  %t2 = load i64, ptr %t1
  %t3 = load { ptr, i64 }, ptr %slot_buf_2
  %t4 = extractvalue { ptr, i64 } %t3, 0
  %t5 = load { ptr, i64 }, ptr %slot_buf_2
  %t6 = extractvalue { ptr, i64 } %t5, 1
  %t7 = call i64 @aila_rt_file_read(i64 %t2, ptr %t4, i64 %t6)
  store i64 %t7, ptr %slot_n_3
  %t8 = load i64, ptr %slot_n_3
  %t9 = icmp slt i64 %t8, 0
  br i1 %t9, label %if.then2, label %if.next3
if.then2:
  %t10 = insertvalue { i8, { i64 } } undef, i8 1, 0
  store { i8, { i64 } } %t10, ptr %slot___lit_4
  %t11 = load { ptr, i64 }, ptr %slot_buf_2
  %t12 = extractvalue { ptr, i64 } %t11, 1
  %t13 = getelementptr inbounds { i8, { i64 } }, ptr %slot___lit_4, i32 0, i32 1, i32 0
  store i64 %t12, ptr %t13
  %t14 = load { i8, { i64 } }, ptr %slot___lit_4
  %t15 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t16 = insertvalue { i1, i64, { i8, { i64 } } } %t15, { i8, { i64 } } %t14, 2
  ret { i1, i64, { i8, { i64 } } } %t16
if.next3:
  br label %if.end1
if.end1:
  %t17 = load i64, ptr %slot_n_3
  %t19 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 1, 0
  %t20 = insertvalue { i1, i64, { i8, { i64 } } } %t19, i64 %t17, 1
  ret { i1, i64, { i8, { i64 } } } %t20
}

define { i1, i64, { i8, { i64 } } } @aila_std_os__write({ i64 } %p_f, { ptr, i64 } %p_bytes) {
  %slot_f_1 = alloca { i64 }, align 8
  %slot_bytes_2 = alloca { ptr, i64 }, align 8
  %slot_n_3 = alloca i64, align 8
  %slot___sc_4 = alloca i1, align 8
  %slot___lit_5 = alloca { i8, { i64 } }, align 8
  store { i64 } %p_f, ptr %slot_f_1
  store { ptr, i64 } %p_bytes, ptr %slot_bytes_2
  %t1 = getelementptr inbounds { i64 }, ptr %slot_f_1, i32 0, i32 0
  %t2 = load i64, ptr %t1
  %t3 = load { ptr, i64 }, ptr %slot_bytes_2
  %t4 = extractvalue { ptr, i64 } %t3, 0
  %t5 = load { ptr, i64 }, ptr %slot_bytes_2
  %t6 = extractvalue { ptr, i64 } %t5, 1
  %t7 = call i64 @aila_rt_file_write(i64 %t2, ptr %t4, i64 %t6)
  store i64 %t7, ptr %slot_n_3
  %t8 = load i64, ptr %slot_n_3
  %t9 = icmp slt i64 %t8, 0
  br i1 %t9, label %sc.l4, label %sc.r5
sc.l4:
  store i1 1, ptr %slot___sc_4
  br label %sc.end6
sc.r5:
  %t10 = load i64, ptr %slot_n_3
  %t11 = load { ptr, i64 }, ptr %slot_bytes_2
  %t12 = extractvalue { ptr, i64 } %t11, 1
  %t14 = icmp ne i64 %t10, %t12
  store i1 %t14, ptr %slot___sc_4
  br label %sc.end6
sc.end6:
  %t15 = load i1, ptr %slot___sc_4
  br i1 %t15, label %if.then2, label %if.next3
if.then2:
  %t16 = insertvalue { i8, { i64 } } undef, i8 1, 0
  store { i8, { i64 } } %t16, ptr %slot___lit_5
  %t17 = load { ptr, i64 }, ptr %slot_bytes_2
  %t18 = extractvalue { ptr, i64 } %t17, 1
  %t19 = getelementptr inbounds { i8, { i64 } }, ptr %slot___lit_5, i32 0, i32 1, i32 0
  store i64 %t18, ptr %t19
  %t20 = load { i8, { i64 } }, ptr %slot___lit_5
  %t21 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t22 = insertvalue { i1, i64, { i8, { i64 } } } %t21, { i8, { i64 } } %t20, 2
  ret { i1, i64, { i8, { i64 } } } %t22
if.next3:
  br label %if.end1
if.end1:
  %t23 = load i64, ptr %slot_n_3
  %t25 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 1, 0
  %t26 = insertvalue { i1, i64, { i8, { i64 } } } %t25, i64 %t23, 1
  ret { i1, i64, { i8, { i64 } } } %t26
}

define void @aila_std_os__close({ i64 } %p_f) {
  %slot_f_1 = alloca { i64 }, align 8
  store { i64 } %p_f, ptr %slot_f_1
  %t1 = getelementptr inbounds { i64 }, ptr %slot_f_1, i32 0, i32 0
  %t2 = load i64, ptr %t1
  call void @aila_rt_file_close(i64 %t2)
  ret void
}

define void @aila_std_os__flush({ i64 } %p_f) {
  %slot_f_1 = alloca { i64 }, align 8
  store { i64 } %p_f, ptr %slot_f_1
  %t1 = getelementptr inbounds { i64 }, ptr %slot_f_1, i32 0, i32 0
  %t2 = load i64, ptr %t1
  call void @aila_rt_file_flush(i64 %t2)
  ret void
}

define void @aila_std_os__exit(i32 %p_code) {
  %slot_code_1 = alloca i32, align 8
  store i32 %p_code, ptr %slot_code_1
  %t1 = load i32, ptr %slot_code_1
  call void @aila_rt_exit(i32 %t1)
  ret void
}

define { i1, i64, { i8, { i64 } } } @aila_std_os__env({ ptr, i64 } %p_name, { ptr, i64 } %p_dst) {
  %slot_name_1 = alloca { ptr, i64 }, align 8
  %slot_dst_2 = alloca { ptr, i64 }, align 8
  %slot_nb_3 = alloca { ptr, i64 }, align 8
  %slot_n_4 = alloca i64, align 8
  %slot_full_5 = alloca i64, align 8
  %slot___lit_6 = alloca { i8, { i64 } }, align 8
  %slot___lit_7 = alloca { i8, { i64 } }, align 8
  store { ptr, i64 } %p_name, ptr %slot_name_1
  store { ptr, i64 } %p_dst, ptr %slot_dst_2
  %t1 = load { ptr, i64 }, ptr %slot_name_1
  store { ptr, i64 } %t1, ptr %slot_nb_3
  %t2 = load { ptr, i64 }, ptr %slot_nb_3
  %t3 = extractvalue { ptr, i64 } %t2, 0
  %t4 = load { ptr, i64 }, ptr %slot_nb_3
  %t5 = extractvalue { ptr, i64 } %t4, 1
  %t6 = load { ptr, i64 }, ptr %slot_dst_2
  %t7 = extractvalue { ptr, i64 } %t6, 0
  %t8 = load { ptr, i64 }, ptr %slot_dst_2
  %t9 = extractvalue { ptr, i64 } %t8, 1
  %t10 = call i64 @aila_rt_env_get(ptr %t3, i64 %t5, ptr %t7, i64 %t9)
  store i64 %t10, ptr %slot_n_4
  %t11 = load i64, ptr %slot_n_4
  %t12 = icmp slt i64 %t11, 0
  br i1 %t12, label %if.then2, label %if.next3
if.then2:
  %t13 = insertvalue { i8, { i64 } } undef, i8 3, 0
  %t14 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t15 = insertvalue { i1, i64, { i8, { i64 } } } %t14, { i8, { i64 } } %t13, 2
  ret { i1, i64, { i8, { i64 } } } %t15
if.next3:
  br label %if.end1
if.end1:
  %t16 = load i64, ptr %slot_n_4
  store i64 %t16, ptr %slot_full_5
  %t18 = load i64, ptr %slot_full_5
  %t19 = load { ptr, i64 }, ptr %slot_dst_2
  %t20 = extractvalue { ptr, i64 } %t19, 1
  %t21 = icmp ugt i64 %t18, %t20
  br i1 %t21, label %if.then5, label %if.next6
if.then5:
  %t22 = insertvalue { i8, { i64 } } undef, i8 1, 0
  store { i8, { i64 } } %t22, ptr %slot___lit_7
  %t23 = load i64, ptr %slot_full_5
  %t24 = getelementptr inbounds { i8, { i64 } }, ptr %slot___lit_7, i32 0, i32 1, i32 0
  store i64 %t23, ptr %t24
  %t25 = load { i8, { i64 } }, ptr %slot___lit_7
  %t26 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t27 = insertvalue { i1, i64, { i8, { i64 } } } %t26, { i8, { i64 } } %t25, 2
  ret { i1, i64, { i8, { i64 } } } %t27
if.next6:
  br label %if.end4
if.end4:
  %t28 = load i64, ptr %slot_full_5
  %t29 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 1, 0
  %t30 = insertvalue { i1, i64, { i8, { i64 } } } %t29, i64 %t28, 1
  ret { i1, i64, { i8, { i64 } } } %t30
}

define { i1, i64, { i8, { i64 } } } @aila_std_os__env_set({ ptr, i64 } %p_name, { ptr, i64 } %p_val) {
  %slot_name_1 = alloca { ptr, i64 }, align 8
  %slot_val_2 = alloca { ptr, i64 }, align 8
  %slot_nb_3 = alloca { ptr, i64 }, align 8
  %slot_vb_4 = alloca { ptr, i64 }, align 8
  %slot_r_5 = alloca i64, align 8
  %slot___lit_6 = alloca { i8, { i64 } }, align 8
  store { ptr, i64 } %p_name, ptr %slot_name_1
  store { ptr, i64 } %p_val, ptr %slot_val_2
  %t1 = load { ptr, i64 }, ptr %slot_name_1
  store { ptr, i64 } %t1, ptr %slot_nb_3
  %t2 = load { ptr, i64 }, ptr %slot_val_2
  store { ptr, i64 } %t2, ptr %slot_vb_4
  %t3 = load { ptr, i64 }, ptr %slot_nb_3
  %t4 = extractvalue { ptr, i64 } %t3, 0
  %t5 = load { ptr, i64 }, ptr %slot_nb_3
  %t6 = extractvalue { ptr, i64 } %t5, 1
  %t7 = load { ptr, i64 }, ptr %slot_vb_4
  %t8 = extractvalue { ptr, i64 } %t7, 0
  %t9 = load { ptr, i64 }, ptr %slot_vb_4
  %t10 = extractvalue { ptr, i64 } %t9, 1
  %t11 = call i64 @aila_rt_env_set(ptr %t4, i64 %t6, ptr %t8, i64 %t10)
  store i64 %t11, ptr %slot_r_5
  %t12 = load i64, ptr %slot_r_5
  %t13 = icmp ne i64 %t12, 0
  br i1 %t13, label %if.then2, label %if.next3
if.then2:
  %t14 = insertvalue { i8, { i64 } } undef, i8 1, 0
  store { i8, { i64 } } %t14, ptr %slot___lit_6
  %t15 = getelementptr inbounds { i8, { i64 } }, ptr %slot___lit_6, i32 0, i32 1, i32 0
  store i64 0, ptr %t15
  %t16 = load { i8, { i64 } }, ptr %slot___lit_6
  %t17 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t18 = insertvalue { i1, i64, { i8, { i64 } } } %t17, { i8, { i64 } } %t16, 2
  ret { i1, i64, { i8, { i64 } } } %t18
if.next3:
  br label %if.end1
if.end1:
  %t19 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 1, 0
  %t20 = insertvalue { i1, i64, { i8, { i64 } } } %t19, i64 0, 1
  ret { i1, i64, { i8, { i64 } } } %t20
}

define i64 @aila_std_os__args_count() {
  %slot_n_1 = alloca i64, align 8
  %t1 = call i64 @aila_rt_args_count()
  store i64 %t1, ptr %slot_n_1
  %t2 = load i64, ptr %slot_n_1
  %t3 = icmp slt i64 %t2, 0
  br i1 %t3, label %if.then2, label %if.next3
if.then2:
  ret i64 0
if.next3:
  br label %if.end1
if.end1:
  %t4 = load i64, ptr %slot_n_1
  ret i64 %t4
}

define { i1, i64, { i8, { i64 } } } @aila_std_os__args_get(i64 %p_i, { ptr, i64 } %p_dst) {
  %slot_i_1 = alloca i64, align 8
  %slot_dst_2 = alloca { ptr, i64 }, align 8
  %slot_n_3 = alloca i64, align 8
  %slot_full_4 = alloca i64, align 8
  %slot___lit_5 = alloca { i8, { i64 } }, align 8
  %slot___lit_6 = alloca { i8, { i64 } }, align 8
  store i64 %p_i, ptr %slot_i_1
  store { ptr, i64 } %p_dst, ptr %slot_dst_2
  %t1 = load i64, ptr %slot_i_1
  %t2 = load { ptr, i64 }, ptr %slot_dst_2
  %t3 = extractvalue { ptr, i64 } %t2, 0
  %t4 = load { ptr, i64 }, ptr %slot_dst_2
  %t5 = extractvalue { ptr, i64 } %t4, 1
  %t6 = call i64 @aila_rt_args_get(i64 %t1, ptr %t3, i64 %t5)
  store i64 %t6, ptr %slot_n_3
  %t7 = load i64, ptr %slot_n_3
  %t8 = icmp slt i64 %t7, 0
  br i1 %t8, label %if.then2, label %if.next3
if.then2:
  %t9 = insertvalue { i8, { i64 } } undef, i8 3, 0
  %t10 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t11 = insertvalue { i1, i64, { i8, { i64 } } } %t10, { i8, { i64 } } %t9, 2
  ret { i1, i64, { i8, { i64 } } } %t11
if.next3:
  br label %if.end1
if.end1:
  %t12 = load i64, ptr %slot_n_3
  store i64 %t12, ptr %slot_full_4
  %t14 = load i64, ptr %slot_full_4
  %t15 = load { ptr, i64 }, ptr %slot_dst_2
  %t16 = extractvalue { ptr, i64 } %t15, 1
  %t17 = icmp ugt i64 %t14, %t16
  br i1 %t17, label %if.then5, label %if.next6
if.then5:
  %t18 = insertvalue { i8, { i64 } } undef, i8 1, 0
  store { i8, { i64 } } %t18, ptr %slot___lit_6
  %t19 = load i64, ptr %slot_full_4
  %t20 = getelementptr inbounds { i8, { i64 } }, ptr %slot___lit_6, i32 0, i32 1, i32 0
  store i64 %t19, ptr %t20
  %t21 = load { i8, { i64 } }, ptr %slot___lit_6
  %t22 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 0, 0
  %t23 = insertvalue { i1, i64, { i8, { i64 } } } %t22, { i8, { i64 } } %t21, 2
  ret { i1, i64, { i8, { i64 } } } %t23
if.next6:
  br label %if.end4
if.end4:
  %t24 = load i64, ptr %slot_full_4
  %t25 = insertvalue { i1, i64, { i8, { i64 } } } undef, i1 1, 0
  %t26 = insertvalue { i1, i64, { i8, { i64 } } } %t25, i64 %t24, 1
  ret { i1, i64, { i8, { i64 } } } %t26
}
