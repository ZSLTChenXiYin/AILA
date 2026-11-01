; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [14 x i8] c"[coro] tick 1\00", align 1
@.str.1 = private unnamed_addr constant [14 x i8] c"[coro] tick 2\00", align 1
@.str.2 = private unnamed_addr constant [16 x i8] c"[coro] done: 42\00", align 1
@.str.3 = private unnamed_addr constant [28 x i8] c"match \E6\9C\AA\E8\A6\86\E7\9B\96\E7\9A\84\E5\88\A4\E5\88\AB\E5\80\BC\00", align 1
@.str.4 = private unnamed_addr constant [10 x i8] c"main.aila\00", align 1
@.str.5 = private unnamed_addr constant [59 x i8] c"resume \E5\B7\B2\E8\A7\82\E5\AF\9F\E8\BF\87\E5\AE\8C\E6\88\90\E6\80\81\E7\9A\84\E5\8D\8F\E7\A8\8B\EF\BC\88\E7\8A\B6\E6\80\81\E6\9C\BA\E8\BF\9D\E8\A7\84\EF\BC\89\00", align 1
@.str.6 = private unnamed_addr constant [20 x i8] c"[coro] returned %d\0A\00", align 1
@.str.7 = private unnamed_addr constant [44 x i8] c"[chan] producer sent %d, consumer sum = %d\0A\00", align 1
@.str.8 = private unnamed_addr constant [19 x i8] c"\E7\BA\BF\E7\A8\8B\E5\88\9B\E5\BB\BA\E5\A4\B1\E8\B4\A5\00", align 1
@.str.9 = private unnamed_addr constant [38 x i8] c"[atomics] count = %d (expect 200000)\0A\00", align 1
@.str.10 = private unnamed_addr constant [41 x i8] c"[atomics] race settled exactly once: %b\0A\00", align 1
@.str.11 = private unnamed_addr constant [48 x i8] c"[atomics] u8: 10 - 4 = %d (fetch_sub old = %d)\0A\00", align 1
@g_main_block = internal global [8 x i64] zeroinitializer, align 8

declare token @llvm.coro.id(i32, ptr, ptr, ptr)
declare i64 @llvm.coro.size.i64()
declare ptr @llvm.coro.begin(token, ptr)
declare token @llvm.coro.save(ptr)
declare i8 @llvm.coro.suspend(token, i1)
declare void @llvm.coro.end(ptr, i1, token)
declare void @llvm.coro.resume(ptr)
declare ptr @aila_rt_region_alloc(ptr, i64, i64)
declare void @aila_std_io__println({ ptr, i64 })
declare i1 @aila_rt_chan_try_send(ptr, ptr)
declare i1 @aila_rt_chan_try_recv(ptr, ptr)
declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare i32 @aila_rt_atomic_fetch_add32(ptr, i64, i32)
declare i32 @aila_rt_atomic_cas32(ptr, i64, i32, i32)
declare void @aila_rt_atomic_store64(ptr, i64, i64)
declare ptr @aila_rt_region_enter(i64)
declare void @aila_std_io__printf({ ptr, i64 }, { ptr, i64 })
declare void @aila_rt_region_leave(ptr)
declare ptr @aila_rt_chan_create(ptr, i64, i64)
declare ptr @aila_rt_thread_args(i64)
declare void @aila_rt_thread_args_free(ptr)
declare i32 @aila_rt_thread_spawn(i64, ptr, ptr)
declare i32 @aila_rt_atomic_load32(ptr, i64)
declare void @aila_std_time__sleep_ms(i64)
declare i64 @aila_rt_atomic_load64(ptr, i64)
declare void @aila_rt_atomic_store8(ptr, i64, i8)
declare i8 @aila_rt_atomic_fetch_sub8(ptr, i64, i8)
declare i8 @aila_rt_atomic_load8(ptr, i64)

define ptr @aila_main__stepper(ptr %cb, ptr %frameReg) #0 {
  %slot___lit_1 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_2 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_3 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_4 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_5 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_6 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %t1 = call token @llvm.coro.id(i32 0, ptr null, ptr null, ptr null)
  %t2 = call i64 @llvm.coro.size.i64()
  %t3 = call ptr @aila_rt_region_alloc(ptr %frameReg, i64 %t2, i64 16)
  %t4 = call ptr @llvm.coro.begin(token %t1, ptr %t3)
  %t5 = getelementptr inbounds i8, ptr %cb, i64 8
  store ptr %t4, ptr %t5
  %t6 = getelementptr inbounds i8, ptr %cb, i64 0
  store i8 0, ptr %t6
  %t7 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t7, ptr %slot___lit_1
  %t8 = insertvalue { ptr, i64 } undef, ptr @.str.0, 0
  %t9 = insertvalue { ptr, i64 } %t8, i64 13, 1
  %t10 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_1, i32 0, i32 1, i32 0
  store { ptr, i64 } %t9, ptr %t10
  %t11 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_1
  %t12 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_2, i64 0
  store { i8, { { ptr, i64 } } } %t11, ptr %t12
  %t13 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_2, 0
  %t14 = insertvalue { ptr, i64 } %t13, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t14)
  %t15 = call token @llvm.coro.save(ptr null)
  %t16 = call i8 @llvm.coro.suspend(token %t15, i1 false)
  switch i8 %t16, label %coro.susp2 [ i8 0, label %coro.after1 ]
coro.susp2:
  br label %coro.ret3
coro.after1:
  %t17 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t17, ptr %slot___lit_3
  %t18 = insertvalue { ptr, i64 } undef, ptr @.str.1, 0
  %t19 = insertvalue { ptr, i64 } %t18, i64 13, 1
  %t20 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_3, i32 0, i32 1, i32 0
  store { ptr, i64 } %t19, ptr %t20
  %t21 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_3
  %t22 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_4, i64 0
  store { i8, { { ptr, i64 } } } %t21, ptr %t22
  %t23 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_4, 0
  %t24 = insertvalue { ptr, i64 } %t23, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t24)
  %t25 = call token @llvm.coro.save(ptr null)
  %t26 = call i8 @llvm.coro.suspend(token %t25, i1 false)
  switch i8 %t26, label %coro.susp5 [ i8 0, label %coro.after4 ]
coro.susp5:
  br label %coro.ret3
coro.after4:
  %t27 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t27, ptr %slot___lit_5
  %t28 = insertvalue { ptr, i64 } undef, ptr @.str.2, 0
  %t29 = insertvalue { ptr, i64 } %t28, i64 15, 1
  %t30 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_5, i32 0, i32 1, i32 0
  store { ptr, i64 } %t29, ptr %t30
  %t31 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_5
  %t32 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_6, i64 0
  store { i8, { { ptr, i64 } } } %t31, ptr %t32
  %t33 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_6, 0
  %t34 = insertvalue { ptr, i64 } %t33, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t34)
  %t35 = getelementptr inbounds i8, ptr %cb, i64 16
  store i32 42, ptr %t35
  %t36 = getelementptr inbounds i8, ptr %cb, i64 0
  store i8 1, ptr %t36
  br label %coro.ret3
coro.ret3:
  call void @llvm.coro.end(ptr %t4, i1 false, token none)
  ret ptr %cb
}

define ptr @aila_main__producer(ptr %p_ch, ptr %cb, ptr %frameReg) #0 {
  %slot_ch_1 = alloca ptr, align 8
  %slot_i_2 = alloca i32, align 8
  %slot___lit_3 = alloca { i8, { i32 } }, align 8
  %slot___send_4 = alloca { i8, { i32 } }, align 8
  %slot___send_5 = alloca { i8, { i32 } }, align 8
  store ptr %p_ch, ptr %slot_ch_1
  %t1 = call token @llvm.coro.id(i32 0, ptr null, ptr null, ptr null)
  %t2 = call i64 @llvm.coro.size.i64()
  %t3 = call ptr @aila_rt_region_alloc(ptr %frameReg, i64 %t2, i64 16)
  %t4 = call ptr @llvm.coro.begin(token %t1, ptr %t3)
  %t5 = getelementptr inbounds i8, ptr %cb, i64 8
  store ptr %t4, ptr %t5
  %t6 = getelementptr inbounds i8, ptr %cb, i64 0
  store i8 0, ptr %t6
  store i32 0, ptr %slot_i_2
  br label %for.cond2
for.cond2:
  %t7 = load i32, ptr %slot_i_2
  %t8 = icmp slt i32 %t7, 3
  br i1 %t8, label %for.body4, label %for.end1
for.body4:
  %t9 = load ptr, ptr %slot_ch_1
  %t10 = insertvalue { i8, { i32 } } undef, i8 0, 0
  store { i8, { i32 } } %t10, ptr %slot___lit_3
  %t11 = load i32, ptr %slot_i_2
  %t12 = getelementptr inbounds { i8, { i32 } }, ptr %slot___lit_3, i32 0, i32 1, i32 0
  store i32 %t11, ptr %t12
  %t13 = load { i8, { i32 } }, ptr %slot___lit_3
  store { i8, { i32 } } %t13, ptr %slot___send_4
  br label %chan.send5
chan.send5:
  %t14 = call i1 @aila_rt_chan_try_send(ptr %t9, ptr %slot___send_4)
  br i1 %t14, label %chan.sent6, label %chan.wait7
chan.wait7:
  %t15 = call token @llvm.coro.save(ptr null)
  %t16 = call i8 @llvm.coro.suspend(token %t15, i1 false)
  switch i8 %t16, label %coro.susp8 [ i8 0, label %chan.send5 ]
coro.susp8:
  br label %coro.ret9
chan.sent6:
  %t17 = load i32, ptr %slot_i_2
  %t18 = add i32 %t17, 1
  store i32 %t18, ptr %slot_i_2
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t19 = load ptr, ptr %slot_ch_1
  %t20 = insertvalue { i8, { i32 } } undef, i8 1, 0
  store { i8, { i32 } } %t20, ptr %slot___send_5
  br label %chan.send10
chan.send10:
  %t21 = call i1 @aila_rt_chan_try_send(ptr %t19, ptr %slot___send_5)
  br i1 %t21, label %chan.sent11, label %chan.wait12
chan.wait12:
  %t22 = call token @llvm.coro.save(ptr null)
  %t23 = call i8 @llvm.coro.suspend(token %t22, i1 false)
  switch i8 %t23, label %coro.susp13 [ i8 0, label %chan.send10 ]
coro.susp13:
  br label %coro.ret9
chan.sent11:
  %t24 = getelementptr inbounds i8, ptr %cb, i64 16
  store i32 3, ptr %t24
  %t25 = getelementptr inbounds i8, ptr %cb, i64 0
  store i8 1, ptr %t25
  br label %coro.ret9
coro.ret9:
  call void @llvm.coro.end(ptr %t4, i1 false, token none)
  ret ptr %cb
}

define ptr @aila_main__consumer(ptr %p_ch, ptr %cb, ptr %frameReg) #0 {
  %slot_ch_1 = alloca ptr, align 8
  %slot_sum_2 = alloca i32, align 8
  %slot_running_3 = alloca i1, align 8
  %slot_v_4 = alloca i32, align 8
  %slot___head_5 = alloca { i8, { i32 } }, align 8
  %slot___recv_6 = alloca { i8, { i32 } }, align 8
  store ptr %p_ch, ptr %slot_ch_1
  %t1 = call token @llvm.coro.id(i32 0, ptr null, ptr null, ptr null)
  %t2 = call i64 @llvm.coro.size.i64()
  %t3 = call ptr @aila_rt_region_alloc(ptr %frameReg, i64 %t2, i64 16)
  %t4 = call ptr @llvm.coro.begin(token %t1, ptr %t3)
  %t5 = getelementptr inbounds i8, ptr %cb, i64 8
  store ptr %t4, ptr %t5
  %t6 = getelementptr inbounds i8, ptr %cb, i64 0
  store i8 0, ptr %t6
  store i32 0, ptr %slot_sum_2
  store i1 1, ptr %slot_running_3
  br label %for.cond2
for.cond2:
  %t7 = load i1, ptr %slot_running_3
  br i1 %t7, label %for.body4, label %for.end1
for.body4:
  %t8 = load ptr, ptr %slot_ch_1
  br label %chan.recv5
chan.recv5:
  %t9 = call i1 @aila_rt_chan_try_recv(ptr %t8, ptr %slot___recv_6)
  br i1 %t9, label %chan.have6, label %chan.wait7
chan.wait7:
  %t10 = call token @llvm.coro.save(ptr null)
  %t11 = call i8 @llvm.coro.suspend(token %t10, i1 false)
  switch i8 %t11, label %coro.susp8 [ i8 0, label %chan.recv5 ]
coro.susp8:
  br label %coro.ret9
chan.have6:
  %t12 = load { i8, { i32 } }, ptr %slot___recv_6
  store { i8, { i32 } } %t12, ptr %slot___head_5
  %t13 = load { i8, { i32 } }, ptr %slot___head_5
  %t14 = extractvalue { i8, { i32 } } %t13, 0
  %t15 = icmp eq i8 %t14, 0
  br i1 %t15, label %m.arm011, label %m.next13
m.next13:
  %t16 = icmp eq i8 %t14, 1
  br i1 %t16, label %m.arm112, label %m.next14
m.next14:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.4, i64 51)
  unreachable
m.arm011:
  %t17 = getelementptr inbounds { i8, { i32 } }, ptr %slot___head_5, i32 0, i32 1, i32 0
  %t18 = load i32, ptr %t17
  store i32 %t18, ptr %slot_v_4
  %t19 = load i32, ptr %slot_sum_2
  %t20 = load i32, ptr %slot_v_4
  %t21 = add i32 %t19, %t20
  store i32 %t21, ptr %slot_sum_2
  br label %m.end10
m.arm112:
  store i1 0, ptr %slot_running_3
  br label %m.end10
m.end10:
  %t22 = call token @llvm.coro.save(ptr null)
  %t23 = call i8 @llvm.coro.suspend(token %t22, i1 false)
  switch i8 %t23, label %coro.susp16 [ i8 0, label %coro.after15 ]
coro.susp16:
  br label %coro.ret9
coro.after15:
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t24 = load i32, ptr %slot_sum_2
  %t25 = getelementptr inbounds i8, ptr %cb, i64 16
  store i32 %t24, ptr %t25
  %t26 = getelementptr inbounds i8, ptr %cb, i64 0
  store i8 1, ptr %t26
  br label %coro.ret9
coro.ret9:
  call void @llvm.coro.end(ptr %t4, i1 false, token none)
  ret ptr %cb
}

define void @aila_main__counter(i32 %p_times) {
  %slot_times_1 = alloca i32, align 8
  %slot_sh_2 = alloca { ptr, i64 }, align 8
  %slot_i_3 = alloca i32, align 8
  store i32 %p_times, ptr %slot_times_1
  %t1 = insertvalue { ptr, i64 } poison, ptr @g_main_block, 0
  %t2 = insertvalue { ptr, i64 } %t1, i64 64, 1
  store { ptr, i64 } %t2, ptr %slot_sh_2
  store i32 0, ptr %slot_i_3
  br label %for.cond2
for.cond2:
  %t3 = load i32, ptr %slot_i_3
  %t4 = load i32, ptr %slot_times_1
  %t5 = icmp ult i32 %t3, %t4
  br i1 %t5, label %for.body4, label %for.end1
for.body4:
  %t6 = load { ptr, i64 }, ptr %slot_sh_2
  %t7 = extractvalue { ptr, i64 } %t6, 0
  %t8 = call i32 @aila_rt_atomic_fetch_add32(ptr %t7, i64 0, i32 1)
  %t9 = load i32, ptr %slot_i_3
  %t10 = add i32 %t9, 1
  store i32 %t10, ptr %slot_i_3
  br label %for.step3
for.step3:
  br label %for.cond2
for.end1:
  %t11 = load { ptr, i64 }, ptr %slot_sh_2
  %t12 = extractvalue { ptr, i64 } %t11, 0
  %t13 = call i32 @aila_rt_atomic_fetch_add32(ptr %t12, i64 4, i32 1)
  ret void
}

define void @aila_main__racer(i64 %p_id) {
  %slot_id_1 = alloca i64, align 8
  %slot_sh_2 = alloca { ptr, i64 }, align 8
  %slot_old_3 = alloca i32, align 8
  store i64 %p_id, ptr %slot_id_1
  %t1 = insertvalue { ptr, i64 } poison, ptr @g_main_block, 0
  %t2 = insertvalue { ptr, i64 } %t1, i64 64, 1
  store { ptr, i64 } %t2, ptr %slot_sh_2
  %t3 = load { ptr, i64 }, ptr %slot_sh_2
  %t4 = extractvalue { ptr, i64 } %t3, 0
  %t5 = call i32 @aila_rt_atomic_cas32(ptr %t4, i64 8, i32 0, i32 1)
  store i32 %t5, ptr %slot_old_3
  %t6 = load i32, ptr %slot_old_3
  %t7 = icmp eq i32 %t6, 0
  br i1 %t7, label %if.then2, label %if.next3
if.then2:
  %t8 = load { ptr, i64 }, ptr %slot_sh_2
  %t9 = extractvalue { ptr, i64 } %t8, 0
  %t10 = load i64, ptr %slot_id_1
  call void @aila_rt_atomic_store64(ptr %t9, i64 16, i64 %t10)
  br label %if.end1
if.next3:
  %t11 = load { ptr, i64 }, ptr %slot_sh_2
  %t12 = extractvalue { ptr, i64 } %t11, 0
  %t13 = load i64, ptr %slot_id_1
  call void @aila_rt_atomic_store64(ptr %t12, i64 24, i64 %t13)
  br label %if.end1
if.end1:
  %t14 = load { ptr, i64 }, ptr %slot_sh_2
  %t15 = extractvalue { ptr, i64 } %t14, 0
  %t16 = call i32 @aila_rt_atomic_fetch_add32(ptr %t15, i64 4, i32 1)
  ret void
}

  define internal ptr @aila_main__counter__thunk(ptr %arg) {
  %v0 = load i32, ptr %arg
  call void @aila_main__counter(i32 %v0)
  call void @aila_rt_thread_args_free(ptr %arg)
  ret ptr null
  }

  define internal ptr @aila_main__racer__thunk(ptr %arg) {
  %v0 = load i64, ptr %arg
  call void @aila_main__racer(i64 %v0)
  call void @aila_rt_thread_args_free(ptr %arg)
  ret ptr null
  }

define void @aila_main__main() {
  %slot_hs_1 = alloca ptr, align 8
  %slot_r_2 = alloca i32, align 8
  %slot_tx_3 = alloca ptr, align 8
  %slot_rx_4 = alloca ptr, align 8
  %slot_hc_5 = alloca ptr, align 8
  %slot_hp_6 = alloca ptr, align 8
  %slot_p_7 = alloca i32, align 8
  %slot_c_8 = alloca i32, align 8
  %slot_pd_9 = alloca i1, align 8
  %slot_cd_10 = alloca i1, align 8
  %slot_r_11 = alloca i32, align 8
  %slot_r_12 = alloca i32, align 8
  %slot_sh_13 = alloca { ptr, i64 }, align 8
  %slot_w_14 = alloca i64, align 8
  %slot_l_15 = alloca i64, align 8
  %slot_settled_16 = alloca i1, align 8
  %slot_old8_17 = alloca i8, align 8
  %slot___sel_sum_18 = alloca { i8, i32 }, align 8
  %slot___rst_19 = alloca { i8, i32 }, align 8
  %slot___lit_20 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_21 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___destr_22 = alloca { ptr, ptr }, align 8
  %slot___sc_23 = alloca i1, align 8
  %slot___sel_sum_24 = alloca { i8, i32 }, align 8
  %slot___sel_sum_25 = alloca { i8, i32 }, align 8
  %slot___rst_26 = alloca { i8, i32 }, align 8
  %slot___rst_27 = alloca { i8, i32 }, align 8
  %slot___lit_28 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___lit_29 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_30 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_31 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_32 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___sc_33 = alloca i1, align 8
  %slot___sc_34 = alloca i1, align 8
  %slot___sc_35 = alloca i1, align 8
  %slot___lit_36 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_37 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_38 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___lit_39 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_40 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = call ptr @aila_rt_region_alloc(ptr %t1, i64 24, i64 8)
  %t3 = call ptr @aila_main__stepper(ptr %t2, ptr %t1)
  store ptr %t3, ptr %slot_hs_1
  br label %sel.retry2
sel.retry2:
  %t4 = load ptr, ptr %slot_hs_1
  %t5 = load i8, ptr %t4
  %t6 = icmp eq i8 %t5, 3
  br i1 %t6, label %sel.next4, label %sel.coro3
sel.coro3:
  %t7 = load i8, ptr %t4
  %t8 = icmp eq i8 %t7, 0
  br i1 %t8, label %resume.run5, label %resume.donechk6
resume.run5:
  %t9 = getelementptr inbounds i8, ptr %t4, i64 8
  %t10 = load ptr, ptr %t9
  call void @llvm.coro.resume(ptr %t10)
  %t11 = load i8, ptr %t4
  %t12 = insertvalue { i8, i32 } undef, i8 %t11, 0
  %t13 = getelementptr inbounds i8, ptr %t4, i64 16
  %t14 = load i32, ptr %t13
  %t15 = insertvalue { i8, i32 } %t12, i32 %t14, 1
  store { i8, i32 } %t15, ptr %slot___rst_19
  %t16 = icmp ne i8 %t11, 0
  br i1 %t16, label %resume.mark7, label %resume.join8
resume.mark7:
  store i8 3, ptr %t4
  br label %resume.join8
resume.donechk6:
  %t17 = icmp eq i8 %t7, 3
  br i1 %t17, label %resume.trap9, label %resume.done10
resume.trap9:
  call void @aila_rt_trap(ptr @.str.5, i64 58, ptr @.str.4, i64 96)
  unreachable
resume.done10:
  %t18 = insertvalue { i8, i32 } undef, i8 %t7, 0
  %t19 = getelementptr inbounds i8, ptr %t4, i64 16
  %t20 = load i32, ptr %t19
  %t21 = insertvalue { i8, i32 } %t18, i32 %t20, 1
  store { i8, i32 } %t21, ptr %slot___rst_19
  store i8 3, ptr %t4
  br label %resume.join8
resume.join8:
  %t22 = load { i8, i32 }, ptr %slot___rst_19
  store { i8, i32 } %t22, ptr %slot___sel_sum_18
  %t23 = extractvalue { i8, i32 } %t22, 0
  %t24 = icmp ne i8 %t23, 0
  br i1 %t24, label %sel.arm1, label %sel.next4
sel.next4:
  br label %sel.retry2
sel.arm1:
  %t25 = load { i8, i32 }, ptr %slot___sel_sum_18
  %t26 = extractvalue { i8, i32 } %t25, 0
  %t27 = extractvalue { i8, i32 } %t25, 1
  store i32 %t27, ptr %slot_r_2
  %t28 = insertvalue { ptr, i64 } undef, ptr @.str.6, 0
  %t29 = insertvalue { ptr, i64 } %t28, i64 19, 1
  %t30 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t30, ptr %slot___lit_20
  %t31 = load i32, ptr %slot_r_2
  %t32 = sext i32 %t31 to i64
  %t33 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_20, i32 0, i32 1, i32 0
  store i64 %t32, ptr %t33
  %t34 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_20
  %t35 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_21, i64 0
  store { i8, { { ptr, i64 } } } %t34, ptr %t35
  %t36 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_21, 0
  %t37 = insertvalue { ptr, i64 } %t36, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t29, { ptr, i64 } %t37)
  br label %sel.end11
sel.end11:
  call void @aila_rt_region_leave(ptr %t1)
  %t38 = call ptr @aila_rt_region_enter(i64 0)
  %t39 = call ptr @aila_rt_chan_create(ptr %t38, i64 1, i64 8)
  %t40 = insertvalue { ptr, ptr } undef, ptr %t39, 0
  %t41 = insertvalue { ptr, ptr } %t40, ptr %t39, 1
  store { ptr, ptr } %t41, ptr %slot___destr_22
  %t42 = getelementptr inbounds { ptr, ptr }, ptr %slot___destr_22, i32 0, i32 0
  %t43 = load ptr, ptr %t42
  store ptr %t43, ptr %slot_tx_3
  %t44 = getelementptr inbounds { ptr, ptr }, ptr %slot___destr_22, i32 0, i32 1
  %t45 = load ptr, ptr %t44
  store ptr %t45, ptr %slot_rx_4
  %t46 = call ptr @aila_rt_region_alloc(ptr %t38, i64 24, i64 8)
  %t47 = load ptr, ptr %slot_rx_4
  %t48 = call ptr @aila_main__consumer(ptr %t47, ptr %t46, ptr %t38)
  store ptr %t48, ptr %slot_hc_5
  %t49 = call ptr @aila_rt_region_alloc(ptr %t38, i64 24, i64 8)
  %t50 = load ptr, ptr %slot_tx_3
  %t51 = call ptr @aila_main__producer(ptr %t50, ptr %t49, ptr %t38)
  store ptr %t51, ptr %slot_hp_6
  store i32 0, ptr %slot_p_7
  store i32 0, ptr %slot_c_8
  store i1 0, ptr %slot_pd_9
  store i1 0, ptr %slot_cd_10
  br label %for.cond13
for.cond13:
  %t52 = load i1, ptr %slot_pd_9
  %t53 = xor i1 %t52, true
  br i1 %t53, label %sc.l15, label %sc.r16
sc.l15:
  store i1 1, ptr %slot___sc_23
  br label %sc.end17
sc.r16:
  %t54 = load i1, ptr %slot_cd_10
  %t55 = xor i1 %t54, true
  store i1 %t55, ptr %slot___sc_23
  br label %sc.end17
sc.end17:
  %t56 = load i1, ptr %slot___sc_23
  br i1 %t56, label %for.body18, label %for.end12
for.body18:
  br label %sel.retry21
sel.retry21:
  %t57 = load ptr, ptr %slot_hp_6
  %t58 = load i8, ptr %t57
  %t59 = icmp eq i8 %t58, 3
  br i1 %t59, label %sel.next23, label %sel.coro22
sel.coro22:
  %t60 = load i8, ptr %t57
  %t61 = icmp eq i8 %t60, 0
  br i1 %t61, label %resume.run24, label %resume.donechk25
resume.run24:
  %t62 = getelementptr inbounds i8, ptr %t57, i64 8
  %t63 = load ptr, ptr %t62
  call void @llvm.coro.resume(ptr %t63)
  %t64 = load i8, ptr %t57
  %t65 = insertvalue { i8, i32 } undef, i8 %t64, 0
  %t66 = getelementptr inbounds i8, ptr %t57, i64 16
  %t67 = load i32, ptr %t66
  %t68 = insertvalue { i8, i32 } %t65, i32 %t67, 1
  store { i8, i32 } %t68, ptr %slot___rst_26
  %t69 = icmp ne i8 %t64, 0
  br i1 %t69, label %resume.mark26, label %resume.join27
resume.mark26:
  store i8 3, ptr %t57
  br label %resume.join27
resume.donechk25:
  %t70 = icmp eq i8 %t60, 3
  br i1 %t70, label %resume.trap28, label %resume.done29
resume.trap28:
  call void @aila_rt_trap(ptr @.str.5, i64 58, ptr @.str.4, i64 114)
  unreachable
resume.done29:
  %t71 = insertvalue { i8, i32 } undef, i8 %t60, 0
  %t72 = getelementptr inbounds i8, ptr %t57, i64 16
  %t73 = load i32, ptr %t72
  %t74 = insertvalue { i8, i32 } %t71, i32 %t73, 1
  store { i8, i32 } %t74, ptr %slot___rst_26
  store i8 3, ptr %t57
  br label %resume.join27
resume.join27:
  %t75 = load { i8, i32 }, ptr %slot___rst_26
  store { i8, i32 } %t75, ptr %slot___sel_sum_24
  %t76 = extractvalue { i8, i32 } %t75, 0
  %t77 = icmp ne i8 %t76, 0
  br i1 %t77, label %sel.arm19, label %sel.next23
sel.next23:
  %t78 = load ptr, ptr %slot_hc_5
  %t79 = load i8, ptr %t78
  %t80 = icmp eq i8 %t79, 3
  br i1 %t80, label %sel.next31, label %sel.coro30
sel.coro30:
  %t81 = load i8, ptr %t78
  %t82 = icmp eq i8 %t81, 0
  br i1 %t82, label %resume.run32, label %resume.donechk33
resume.run32:
  %t83 = getelementptr inbounds i8, ptr %t78, i64 8
  %t84 = load ptr, ptr %t83
  call void @llvm.coro.resume(ptr %t84)
  %t85 = load i8, ptr %t78
  %t86 = insertvalue { i8, i32 } undef, i8 %t85, 0
  %t87 = getelementptr inbounds i8, ptr %t78, i64 16
  %t88 = load i32, ptr %t87
  %t89 = insertvalue { i8, i32 } %t86, i32 %t88, 1
  store { i8, i32 } %t89, ptr %slot___rst_27
  %t90 = icmp ne i8 %t85, 0
  br i1 %t90, label %resume.mark34, label %resume.join35
resume.mark34:
  store i8 3, ptr %t78
  br label %resume.join35
resume.donechk33:
  %t91 = icmp eq i8 %t81, 3
  br i1 %t91, label %resume.trap36, label %resume.done37
resume.trap36:
  call void @aila_rt_trap(ptr @.str.5, i64 58, ptr @.str.4, i64 115)
  unreachable
resume.done37:
  %t92 = insertvalue { i8, i32 } undef, i8 %t81, 0
  %t93 = getelementptr inbounds i8, ptr %t78, i64 16
  %t94 = load i32, ptr %t93
  %t95 = insertvalue { i8, i32 } %t92, i32 %t94, 1
  store { i8, i32 } %t95, ptr %slot___rst_27
  store i8 3, ptr %t78
  br label %resume.join35
resume.join35:
  %t96 = load { i8, i32 }, ptr %slot___rst_27
  store { i8, i32 } %t96, ptr %slot___sel_sum_25
  %t97 = extractvalue { i8, i32 } %t96, 0
  %t98 = icmp ne i8 %t97, 0
  br i1 %t98, label %sel.arm20, label %sel.next31
sel.next31:
  br label %sel.retry21
sel.arm19:
  %t99 = load { i8, i32 }, ptr %slot___sel_sum_24
  %t100 = extractvalue { i8, i32 } %t99, 0
  %t101 = extractvalue { i8, i32 } %t99, 1
  store i32 %t101, ptr %slot_r_11
  %t102 = load i32, ptr %slot_r_11
  store i32 %t102, ptr %slot_p_7
  store i1 1, ptr %slot_pd_9
  br label %sel.end38
sel.arm20:
  %t103 = load { i8, i32 }, ptr %slot___sel_sum_25
  %t104 = extractvalue { i8, i32 } %t103, 0
  %t105 = extractvalue { i8, i32 } %t103, 1
  store i32 %t105, ptr %slot_r_12
  %t106 = load i32, ptr %slot_r_12
  store i32 %t106, ptr %slot_c_8
  store i1 1, ptr %slot_cd_10
  br label %sel.end38
sel.end38:
  br label %for.step14
for.step14:
  br label %for.cond13
for.end12:
  %t107 = insertvalue { ptr, i64 } undef, ptr @.str.7, 0
  %t108 = insertvalue { ptr, i64 } %t107, i64 43, 1
  %t109 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t109, ptr %slot___lit_28
  %t110 = load i32, ptr %slot_p_7
  %t111 = sext i32 %t110 to i64
  %t112 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_28, i32 0, i32 1, i32 0
  store i64 %t111, ptr %t112
  %t113 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_28
  %t114 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t114, ptr %slot___lit_29
  %t115 = load i32, ptr %slot_c_8
  %t116 = sext i32 %t115 to i64
  %t117 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_29, i32 0, i32 1, i32 0
  store i64 %t116, ptr %t117
  %t118 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_29
  %t119 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_30, i64 0
  store { i8, { { ptr, i64 } } } %t113, ptr %t119
  %t120 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_30, i64 1
  store { i8, { { ptr, i64 } } } %t118, ptr %t120
  %t121 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_30, 0
  %t122 = insertvalue { ptr, i64 } %t121, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t108, { ptr, i64 } %t122)
  call void @aila_rt_region_leave(ptr %t38)
  %t123 = insertvalue { ptr, i64 } poison, ptr @g_main_block, 0
  %t124 = insertvalue { ptr, i64 } %t123, i64 64, 1
  store { ptr, i64 } %t124, ptr %slot_sh_13
  %t125 = call ptr @aila_rt_thread_args(i64 4)
  store i32 100000, ptr %t125
  %t126 = call i32 @aila_rt_thread_spawn(i64 0, ptr @aila_main__counter__thunk, ptr %t125)
  %t127 = icmp ne i32 %t126, 0
  br i1 %t127, label %chk.bad40, label %chk.ok39
chk.bad40:
  call void @aila_rt_trap(ptr @.str.8, i64 18, ptr @.str.4, i64 124)
  unreachable
chk.ok39:
  %t128 = call ptr @aila_rt_thread_args(i64 4)
  store i32 100000, ptr %t128
  %t129 = call i32 @aila_rt_thread_spawn(i64 0, ptr @aila_main__counter__thunk, ptr %t128)
  %t130 = icmp ne i32 %t129, 0
  br i1 %t130, label %chk.bad42, label %chk.ok41
chk.bad42:
  call void @aila_rt_trap(ptr @.str.8, i64 18, ptr @.str.4, i64 125)
  unreachable
chk.ok41:
  %t131 = call ptr @aila_rt_thread_args(i64 8)
  store i64 7, ptr %t131
  %t132 = call i32 @aila_rt_thread_spawn(i64 0, ptr @aila_main__racer__thunk, ptr %t131)
  %t133 = icmp ne i32 %t132, 0
  br i1 %t133, label %chk.bad44, label %chk.ok43
chk.bad44:
  call void @aila_rt_trap(ptr @.str.8, i64 18, ptr @.str.4, i64 126)
  unreachable
chk.ok43:
  %t134 = call ptr @aila_rt_thread_args(i64 8)
  store i64 9, ptr %t134
  %t135 = call i32 @aila_rt_thread_spawn(i64 0, ptr @aila_main__racer__thunk, ptr %t134)
  %t136 = icmp ne i32 %t135, 0
  br i1 %t136, label %chk.bad46, label %chk.ok45
chk.bad46:
  call void @aila_rt_trap(ptr @.str.8, i64 18, ptr @.str.4, i64 127)
  unreachable
chk.ok45:
  br label %for.cond48
for.cond48:
  %t137 = load { ptr, i64 }, ptr %slot_sh_13
  %t138 = extractvalue { ptr, i64 } %t137, 0
  %t139 = call i32 @aila_rt_atomic_load32(ptr %t138, i64 4)
  %t140 = icmp ult i32 %t139, 4
  br i1 %t140, label %for.body50, label %for.end47
for.body50:
  call void @aila_std_time__sleep_ms(i64 1)
  br label %for.step49
for.step49:
  br label %for.cond48
for.end47:
  %t141 = insertvalue { ptr, i64 } undef, ptr @.str.9, 0
  %t142 = insertvalue { ptr, i64 } %t141, i64 37, 1
  %t143 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t143, ptr %slot___lit_31
  %t144 = load { ptr, i64 }, ptr %slot_sh_13
  %t145 = extractvalue { ptr, i64 } %t144, 0
  %t146 = call i32 @aila_rt_atomic_load32(ptr %t145, i64 0)
  %t147 = zext i32 %t146 to i64
  %t148 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_31, i32 0, i32 1, i32 0
  store i64 %t147, ptr %t148
  %t149 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_31
  %t150 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_32, i64 0
  store { i8, { { ptr, i64 } } } %t149, ptr %t150
  %t151 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_32, 0
  %t152 = insertvalue { ptr, i64 } %t151, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t142, { ptr, i64 } %t152)
  %t153 = load { ptr, i64 }, ptr %slot_sh_13
  %t154 = extractvalue { ptr, i64 } %t153, 0
  %t155 = call i64 @aila_rt_atomic_load64(ptr %t154, i64 16)
  store i64 %t155, ptr %slot_w_14
  %t156 = load { ptr, i64 }, ptr %slot_sh_13
  %t157 = extractvalue { ptr, i64 } %t156, 0
  %t158 = call i64 @aila_rt_atomic_load64(ptr %t157, i64 24)
  store i64 %t158, ptr %slot_l_15
  %t159 = load i64, ptr %slot_w_14
  %t160 = icmp eq i64 %t159, 7
  br i1 %t160, label %sc.r55, label %sc.l54
sc.l54:
  store i1 0, ptr %slot___sc_34
  br label %sc.end56
sc.r55:
  %t161 = load i64, ptr %slot_l_15
  %t162 = icmp eq i64 %t161, 9
  store i1 %t162, ptr %slot___sc_34
  br label %sc.end56
sc.end56:
  %t163 = load i1, ptr %slot___sc_34
  br i1 %t163, label %sc.l51, label %sc.r52
sc.l51:
  store i1 1, ptr %slot___sc_33
  br label %sc.end53
sc.r52:
  %t164 = load i64, ptr %slot_w_14
  %t165 = icmp eq i64 %t164, 9
  br i1 %t165, label %sc.r58, label %sc.l57
sc.l57:
  store i1 0, ptr %slot___sc_35
  br label %sc.end59
sc.r58:
  %t166 = load i64, ptr %slot_l_15
  %t167 = icmp eq i64 %t166, 7
  store i1 %t167, ptr %slot___sc_35
  br label %sc.end59
sc.end59:
  %t168 = load i1, ptr %slot___sc_35
  store i1 %t168, ptr %slot___sc_33
  br label %sc.end53
sc.end53:
  %t169 = load i1, ptr %slot___sc_33
  store i1 %t169, ptr %slot_settled_16
  %t170 = insertvalue { ptr, i64 } undef, ptr @.str.10, 0
  %t171 = insertvalue { ptr, i64 } %t170, i64 40, 1
  %t172 = insertvalue { i8, { { ptr, i64 } } } undef, i8 3, 0
  store { i8, { { ptr, i64 } } } %t172, ptr %slot___lit_36
  %t173 = load i1, ptr %slot_settled_16
  %t174 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_36, i32 0, i32 1, i32 0
  store i1 %t173, ptr %t174
  %t175 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_36
  %t176 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_37, i64 0
  store { i8, { { ptr, i64 } } } %t175, ptr %t176
  %t177 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_37, 0
  %t178 = insertvalue { ptr, i64 } %t177, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t171, { ptr, i64 } %t178)
  %t179 = load { ptr, i64 }, ptr %slot_sh_13
  %t180 = extractvalue { ptr, i64 } %t179, 0
  call void @aila_rt_atomic_store8(ptr %t180, i64 32, i8 10)
  %t181 = load { ptr, i64 }, ptr %slot_sh_13
  %t182 = extractvalue { ptr, i64 } %t181, 0
  %t183 = call i8 @aila_rt_atomic_fetch_sub8(ptr %t182, i64 32, i8 4)
  store i8 %t183, ptr %slot_old8_17
  %t184 = insertvalue { ptr, i64 } undef, ptr @.str.11, 0
  %t185 = insertvalue { ptr, i64 } %t184, i64 47, 1
  %t186 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t186, ptr %slot___lit_38
  %t187 = load { ptr, i64 }, ptr %slot_sh_13
  %t188 = extractvalue { ptr, i64 } %t187, 0
  %t189 = call i8 @aila_rt_atomic_load8(ptr %t188, i64 32)
  %t190 = zext i8 %t189 to i64
  %t191 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_38, i32 0, i32 1, i32 0
  store i64 %t190, ptr %t191
  %t192 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_38
  %t193 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t193, ptr %slot___lit_39
  %t194 = load i8, ptr %slot_old8_17
  %t195 = zext i8 %t194 to i64
  %t196 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_39, i32 0, i32 1, i32 0
  store i64 %t195, ptr %t196
  %t197 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_39
  %t198 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_40, i64 0
  store { i8, { { ptr, i64 } } } %t192, ptr %t198
  %t199 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_40, i64 1
  store { i8, { { ptr, i64 } } } %t197, ptr %t199
  %t200 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_40, 0
  %t201 = insertvalue { ptr, i64 } %t200, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t185, { ptr, i64 } %t201)
  ret void
}

define i32 @main(i32 %argc, ptr %argv) {
entry:
  call void @aila_main__main()
  ret i32 0
}

attributes #0 = { presplitcoroutine }
