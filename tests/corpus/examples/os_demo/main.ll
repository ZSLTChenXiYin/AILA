; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.1 = private unnamed_addr constant [10 x i8] c"main.aila\00", align 1
@.str.2 = private unnamed_addr constant [17 x i8] c"args_count = %u\0A\00", align 1
@.str.3 = private unnamed_addr constant [28 x i8] c"match \E6\9C\AA\E8\A6\86\E7\9B\96\E7\9A\84\E5\88\A4\E5\88\AB\E5\80\BC\00", align 1
@.str.4 = private unnamed_addr constant [16 x i8] c"unexpected: %u\0A\00", align 1
@.str.5 = private unnamed_addr constant [25 x i8] c"args_get(99) -> NotFound\00", align 1
@.str.6 = private unnamed_addr constant [6 x i8] c"other\00", align 1
@.str.7 = private unnamed_addr constant [10 x i8] c"AILA_DEMO\00", align 1
@.str.8 = private unnamed_addr constant [3 x i8] c"42\00", align 1
@.str.9 = private unnamed_addr constant [13 x i8] c"AILA_DEMO = \00", align 1
@.str.10 = private unnamed_addr constant [1 x i8] c"\00", align 1
@.str.11 = private unnamed_addr constant [16 x i8] c"env read failed\00", align 1
@.str.12 = private unnamed_addr constant [15 x i8] c"env_set failed\00", align 1
@.str.13 = private unnamed_addr constant [14 x i8] c"AILA_NOPE_XYZ\00", align 1
@.str.14 = private unnamed_addr constant [26 x i8] c"AILA_NOPE_XYZ -> NotFound\00", align 1
@.str.15 = private unnamed_addr constant [17 x i8] c"hello from aila\0A\00", align 1
@.str.16 = private unnamed_addr constant [16 x i8] c"os_demo_tmp.txt\00", align 1
@.str.17 = private unnamed_addr constant [21 x i8] c"file wrote %u bytes\0A\00", align 1
@.str.18 = private unnamed_addr constant [13 x i8] c"write failed\00", align 1
@.str.19 = private unnamed_addr constant [14 x i8] c"open w failed\00", align 1
@.str.20 = private unnamed_addr constant [17 x i8] c"file read back: \00", align 1
@.str.21 = private unnamed_addr constant [12 x i8] c"read failed\00", align 1
@.str.22 = private unnamed_addr constant [23 x i8] c"file read -> %u (EOF)\0A\00", align 1
@.str.23 = private unnamed_addr constant [13 x i8] c"read2 failed\00", align 1
@.str.24 = private unnamed_addr constant [14 x i8] c"open r failed\00", align 1
@.str.25 = private unnamed_addr constant [27 x i8] c"sleep(40) measured ok: %b\0A\00", align 1
@.str.26 = private unnamed_addr constant [23 x i8] c"replay consistent: %b\0A\00", align 1
@.str.27 = private unnamed_addr constant [24 x i8] c"below(10) in range: %b\0A\00", align 1
@.str.28 = private unnamed_addr constant [10 x i8] c"a/b/c.txt\00", align 1
@.str.29 = private unnamed_addr constant [30 x i8] c"base(\22a/b/c.txt\22) = [%u, %u)\0A\00", align 1
@.str.30 = private unnamed_addr constant [29 x i8] c"dir(\22a/b/c.txt\22) = [%u, %u)\0A\00", align 1
@.str.31 = private unnamed_addr constant [29 x i8] c"ext(\22a/b/c.txt\22) = [%u, %u)\0A\00", align 1
@.str.32 = private unnamed_addr constant [6 x i8] c"a/b/c\00", align 1
@.str.33 = private unnamed_addr constant [25 x i8] c"ext(\22a/b/c\22) = [%u, %u)\0A\00", align 1
@.str.34 = private unnamed_addr constant [2 x i8] c"a\00", align 1
@.str.35 = private unnamed_addr constant [2 x i8] c"b\00", align 1
@.str.36 = private unnamed_addr constant [18 x i8] c"join(\22a\22, \22b\22) = \00", align 1
@.str.37 = private unnamed_addr constant [18 x i8] c"\0Ajoin(\22\22, \22b\22) = \00", align 1
@.str.38 = private unnamed_addr constant [5 x i8] c"/abs\00", align 1
@.str.39 = private unnamed_addr constant [22 x i8] c"\0Ajoin(\22a\22, \22/abs\22) = \00", align 1

declare ptr @aila_rt_region_alloc(ptr, i64, i64)
declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare { ptr, { ptr, i64 } } @aila_std_io__stdout()
declare i64 @aila_std_io__Writer__put(ptr, { ptr, i64 })
declare ptr @aila_rt_region_enter(i64)
declare i64 @aila_std_os__args_count()
declare void @aila_std_io__printf({ ptr, i64 }, { ptr, i64 })
declare { i1, i64, { i8, { i64 } } } @aila_std_os__args_get(i64, { ptr, i64 })
declare void @aila_std_io__println({ ptr, i64 })
declare { i1, i64, { i8, { i64 } } } @aila_std_os__env_set({ ptr, i64 }, { ptr, i64 })
declare { i1, i64, { i8, { i64 } } } @aila_std_os__env({ ptr, i64 }, { ptr, i64 })
declare void @aila_std_io__print({ ptr, i64 })
declare { i1, { i64 }, { i8, { i64 } } } @aila_std_os__open({ ptr, i64 }, i32)
declare { i1, i64, { i8, { i64 } } } @aila_std_os__write({ i64 }, { ptr, i64 })
declare void @aila_std_os__close({ i64 })
declare { i1, i64, { i8, { i64 } } } @aila_std_os__read({ i64 }, { ptr, i64 })
declare i64 @aila_std_time__mono_ms()
declare void @aila_std_time__sleep_ms(i64)
declare i64 @aila_std_rand__Rand__next_u64(ptr)
declare i64 @aila_std_rand__Rand__below(ptr, i64)
declare { i64, i64 } @aila_std_path__base({ ptr, i64 }, i8)
declare { i64, i64 } @aila_std_path__dir({ ptr, i64 }, i8)
declare { i64, i64 } @aila_std_path__ext({ ptr, i64 })
declare i64 @aila_std_path__join({ ptr, i64 }, { ptr, i64 }, i8, { ptr, i64 })
declare void @aila_rt_region_leave(ptr)
declare void @aila_rt_args_save(i32, ptr)

define void @aila_main__show(ptr %r0, { ptr, i64 } %p_buf, i64 %p_n) {
  %slot_buf_1 = alloca { ptr, i64 }, align 8
  %slot_n_2 = alloca i64, align 8
  %slot_tmp_3 = alloca { ptr, i64 }, align 8
  %slot_i_4 = alloca i64, align 8
  %slot_recv_5 = alloca { ptr, { ptr, i64 } }, align 8
  store { ptr, i64 } %p_buf, ptr %slot_buf_1
  store i64 %p_n, ptr %slot_n_2
  %t1 = load i64, ptr %slot_n_2
  %t2 = icmp eq i64 %t1, 0
  br i1 %t2, label %if.then2, label %if.next3
if.then2:
  ret void
if.next3:
  br label %if.end1
if.end1:
  %t3 = load i64, ptr %slot_n_2
  %t4 = mul i64 %t3, 1
  %t5 = call ptr @aila_rt_region_alloc(ptr %r0, i64 %t4, i64 1)
  %t6 = insertvalue { ptr, i64 } undef, ptr %t5, 0
  %t7 = insertvalue { ptr, i64 } %t6, i64 %t3, 1
  store { ptr, i64 } %t7, ptr %slot_tmp_3
  store i64 0, ptr %slot_i_4
  br label %for.cond5
for.cond5:
  %t8 = load i64, ptr %slot_i_4
  %t9 = load i64, ptr %slot_n_2
  %t10 = icmp ult i64 %t8, %t9
  br i1 %t10, label %for.body7, label %for.end4
for.body7:
  %t11 = load i64, ptr %slot_i_4
  %t12 = load { ptr, i64 }, ptr %slot_tmp_3
  %t13 = extractvalue { ptr, i64 } %t12, 0
  %t14 = extractvalue { ptr, i64 } %t12, 1
  %t15 = icmp uge i64 %t11, %t14
  br i1 %t15, label %chk.bad9, label %chk.ok8
chk.bad9:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 40)
  unreachable
chk.ok8:
  %t16 = getelementptr inbounds i8, ptr %t13, i64 %t11
  %t17 = load i64, ptr %slot_i_4
  %t18 = load { ptr, i64 }, ptr %slot_buf_1
  %t19 = extractvalue { ptr, i64 } %t18, 0
  %t20 = extractvalue { ptr, i64 } %t18, 1
  %t21 = icmp uge i64 %t17, %t20
  br i1 %t21, label %chk.bad11, label %chk.ok10
chk.bad11:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 40)
  unreachable
chk.ok10:
  %t22 = getelementptr inbounds i8, ptr %t19, i64 %t17
  %t23 = load i8, ptr %t22
  store i8 %t23, ptr %t16
  %t24 = load i64, ptr %slot_i_4
  %t25 = add i64 %t24, 1
  store i64 %t25, ptr %slot_i_4
  br label %for.step6
for.step6:
  br label %for.cond5
for.end4:
  %t26 = call { ptr, { ptr, i64 } } @aila_std_io__stdout()
  store { ptr, { ptr, i64 } } %t26, ptr %slot_recv_5
  %t27 = load { ptr, i64 }, ptr %slot_tmp_3
  %t28 = call i64 @aila_std_io__Writer__put(ptr %slot_recv_5, { ptr, i64 } %t27)
  ret void
}

define void @aila_main__main() {
  %slot_ab_1 = alloca { ptr, i64 }, align 8
  %slot_n_2 = alloca i64, align 8
  %slot_e_3 = alloca { i8, { i64 } }, align 8
  %slot_n_4 = alloca i64, align 8
  %slot_eb_5 = alloca { ptr, i64 }, align 8
  %slot_n_6 = alloca i64, align 8
  %slot_e_7 = alloca { i8, { i64 } }, align 8
  %slot_e_8 = alloca { i8, { i64 } }, align 8
  %slot_n_9 = alloca i64, align 8
  %slot_e_10 = alloca { i8, { i64 } }, align 8
  %slot_msg_11 = alloca { ptr, i64 }, align 8
  %slot_f_12 = alloca { i64 }, align 8
  %slot_n_13 = alloca i64, align 8
  %slot_e_14 = alloca { i8, { i64 } }, align 8
  %slot_e_15 = alloca { i8, { i64 } }, align 8
  %slot_f_16 = alloca { i64 }, align 8
  %slot_fbuf_17 = alloca { ptr, i64 }, align 8
  %slot_n_18 = alloca i64, align 8
  %slot_e_19 = alloca { i8, { i64 } }, align 8
  %slot_n_20 = alloca i64, align 8
  %slot_e_21 = alloca { i8, { i64 } }, align 8
  %slot_e_22 = alloca { i8, { i64 } }, align 8
  %slot_t0_23 = alloca i64, align 8
  %slot_dt_24 = alloca i64, align 8
  %slot_r1_25 = alloca { i64 }, align 8
  %slot_a1_26 = alloca i64, align 8
  %slot_a2_27 = alloca i64, align 8
  %slot_r2_28 = alloca { i64 }, align 8
  %slot_b1_29 = alloca i64, align 8
  %slot_b2_30 = alloca i64, align 8
  %slot_inrange_31 = alloca i1, align 8
  %slot_i_32 = alloca i64, align 8
  %slot_s_33 = alloca i64, align 8
  %slot_e_34 = alloca i64, align 8
  %slot_pbuf_35 = alloca { ptr, i64 }, align 8
  %slot_w_36 = alloca i64, align 8
  %slot___lit_37 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_38 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_39 = alloca { i1, i64, { i8, { i64 } } }, align 8
  %slot___lit_40 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_41 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_42 = alloca { i8, { i64 } }, align 8
  %slot___lit_43 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_44 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_45 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_46 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_47 = alloca { i1, i64, { i8, { i64 } } }, align 8
  %slot___head_48 = alloca { i1, i64, { i8, { i64 } } }, align 8
  %slot___lit_49 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_50 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_51 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_52 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_53 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_54 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_55 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_56 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_57 = alloca { i1, i64, { i8, { i64 } } }, align 8
  %slot___lit_58 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_59 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_60 = alloca { i8, { i64 } }, align 8
  %slot___lit_61 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_62 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_63 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_64 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_65 = alloca { i1, { i64 }, { i8, { i64 } } }, align 8
  %slot___head_66 = alloca { i1, i64, { i8, { i64 } } }, align 8
  %slot___lit_67 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_68 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_69 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_70 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_71 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_72 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_73 = alloca { i1, { i64 }, { i8, { i64 } } }, align 8
  %slot___head_74 = alloca { i1, i64, { i8, { i64 } } }, align 8
  %slot___lit_75 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_76 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_77 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_78 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_79 = alloca { i1, i64, { i8, { i64 } } }, align 8
  %slot___lit_80 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_81 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_82 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_83 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_84 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_85 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_86 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___sc_87 = alloca i1, align 8
  %slot_vargs_88 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_89 = alloca { i64 }, align 8
  %slot___lit_90 = alloca { i64 }, align 8
  %slot___lit_91 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___sc_92 = alloca i1, align 8
  %slot___sc_93 = alloca i1, align 8
  %slot_vargs_94 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_95 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_96 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___destr_97 = alloca { i64, i64 }, align 8
  %slot___lit_98 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___lit_99 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_100 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot___destr_101 = alloca { i64, i64 }, align 8
  %slot___lit_102 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___lit_103 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_104 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot___destr_105 = alloca { i64, i64 }, align 8
  %slot___lit_106 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___lit_107 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_108 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot___destr_109 = alloca { i64, i64 }, align 8
  %slot___lit_110 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___lit_111 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_112 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_113 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_114 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_115 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_116 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_117 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_118 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_119 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_120 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = insertvalue { ptr, i64 } undef, ptr @.str.2, 0
  %t3 = insertvalue { ptr, i64 } %t2, i64 16, 1
  %t4 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t4, ptr %slot___lit_37
  %t5 = call i64 @aila_std_os__args_count()
  %t7 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_37, i32 0, i32 1, i32 0
  store i64 %t5, ptr %t7
  %t8 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_37
  %t9 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_38, i64 0
  store { i8, { { ptr, i64 } } } %t8, ptr %t9
  %t10 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_38, 0
  %t11 = insertvalue { ptr, i64 } %t10, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t3, { ptr, i64 } %t11)
  %t12 = call ptr @aila_rt_region_alloc(ptr %t1, i64 64, i64 1)
  %t13 = insertvalue { ptr, i64 } undef, ptr %t12, 0
  %t14 = insertvalue { ptr, i64 } %t13, i64 64, 1
  store { ptr, i64 } %t14, ptr %slot_ab_1
  %t15 = load { ptr, i64 }, ptr %slot_ab_1
  %t16 = call { i1, i64, { i8, { i64 } } } @aila_std_os__args_get(i64 99, { ptr, i64 } %t15)
  store { i1, i64, { i8, { i64 } } } %t16, ptr %slot___head_39
  %t17 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_39
  %t18 = extractvalue { i1, i64, { i8, { i64 } } } %t17, 0
  br i1 %t18, label %m.arm02, label %m.next4
m.next4:
  %t19 = xor i1 %t18, true
  br i1 %t19, label %m.arm13, label %m.next5
m.next5:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 51)
  unreachable
m.arm02:
  %t20 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_39
  %t21 = extractvalue { i1, i64, { i8, { i64 } } } %t20, 1
  store i64 %t21, ptr %slot_n_2
  %t22 = insertvalue { ptr, i64 } undef, ptr @.str.4, 0
  %t23 = insertvalue { ptr, i64 } %t22, i64 15, 1
  %t24 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t24, ptr %slot___lit_40
  %t25 = load i64, ptr %slot_n_2
  %t27 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_40, i32 0, i32 1, i32 0
  store i64 %t25, ptr %t27
  %t28 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_40
  %t29 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_41, i64 0
  store { i8, { { ptr, i64 } } } %t28, ptr %t29
  %t30 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_41, 0
  %t31 = insertvalue { ptr, i64 } %t30, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t23, { ptr, i64 } %t31)
  br label %m.end1
m.arm13:
  %t32 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_39
  %t33 = extractvalue { i1, i64, { i8, { i64 } } } %t32, 2
  store { i8, { i64 } } %t33, ptr %slot_e_3
  %t34 = load { i8, { i64 } }, ptr %slot_e_3
  store { i8, { i64 } } %t34, ptr %slot___head_42
  %t35 = load { i8, { i64 } }, ptr %slot___head_42
  %t36 = extractvalue { i8, { i64 } } %t35, 0
  %t37 = icmp eq i8 %t36, 3
  br i1 %t37, label %m.arm07, label %m.next9
m.next9:
  br label %m.arm18
m.arm07:
  %t38 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t38, ptr %slot___lit_43
  %t39 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t40 = insertvalue { ptr, i64 } %t39, i64 24, 1
  %t41 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_43, i32 0, i32 1, i32 0
  store { ptr, i64 } %t40, ptr %t41
  %t42 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_43
  %t43 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_44, i64 0
  store { i8, { { ptr, i64 } } } %t42, ptr %t43
  %t44 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_44, 0
  %t45 = insertvalue { ptr, i64 } %t44, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t45)
  br label %m.end6
m.arm18:
  %t46 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t46, ptr %slot___lit_45
  %t47 = insertvalue { ptr, i64 } undef, ptr @.str.6, 0
  %t48 = insertvalue { ptr, i64 } %t47, i64 5, 1
  %t49 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_45, i32 0, i32 1, i32 0
  store { ptr, i64 } %t48, ptr %t49
  %t50 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_45
  %t51 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_46, i64 0
  store { i8, { { ptr, i64 } } } %t50, ptr %t51
  %t52 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_46, 0
  %t53 = insertvalue { ptr, i64 } %t52, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t53)
  br label %m.end6
m.end6:
  br label %m.end1
m.end1:
  %t54 = insertvalue { ptr, i64 } undef, ptr @.str.7, 0
  %t55 = insertvalue { ptr, i64 } %t54, i64 9, 1
  %t56 = insertvalue { ptr, i64 } undef, ptr @.str.8, 0
  %t57 = insertvalue { ptr, i64 } %t56, i64 2, 1
  %t58 = call { i1, i64, { i8, { i64 } } } @aila_std_os__env_set({ ptr, i64 } %t55, { ptr, i64 } %t57)
  store { i1, i64, { i8, { i64 } } } %t58, ptr %slot___head_47
  %t59 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_47
  %t60 = extractvalue { i1, i64, { i8, { i64 } } } %t59, 0
  br i1 %t60, label %m.arm011, label %m.next13
m.next13:
  %t61 = xor i1 %t60, true
  br i1 %t61, label %m.arm112, label %m.next14
m.next14:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 62)
  unreachable
m.arm011:
  %t62 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_47
  %t63 = extractvalue { i1, i64, { i8, { i64 } } } %t62, 1
  store i64 %t63, ptr %slot_n_4
  %t64 = call ptr @aila_rt_region_alloc(ptr %t1, i64 32, i64 1)
  %t65 = insertvalue { ptr, i64 } undef, ptr %t64, 0
  %t66 = insertvalue { ptr, i64 } %t65, i64 32, 1
  store { ptr, i64 } %t66, ptr %slot_eb_5
  %t67 = insertvalue { ptr, i64 } undef, ptr @.str.7, 0
  %t68 = insertvalue { ptr, i64 } %t67, i64 9, 1
  %t69 = load { ptr, i64 }, ptr %slot_eb_5
  %t70 = call { i1, i64, { i8, { i64 } } } @aila_std_os__env({ ptr, i64 } %t68, { ptr, i64 } %t69)
  store { i1, i64, { i8, { i64 } } } %t70, ptr %slot___head_48
  %t71 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_48
  %t72 = extractvalue { i1, i64, { i8, { i64 } } } %t71, 0
  br i1 %t72, label %m.arm016, label %m.next18
m.next18:
  %t73 = xor i1 %t72, true
  br i1 %t73, label %m.arm117, label %m.next19
m.next19:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 65)
  unreachable
m.arm016:
  %t74 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_48
  %t75 = extractvalue { i1, i64, { i8, { i64 } } } %t74, 1
  store i64 %t75, ptr %slot_n_6
  %t76 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t76, ptr %slot___lit_49
  %t77 = insertvalue { ptr, i64 } undef, ptr @.str.9, 0
  %t78 = insertvalue { ptr, i64 } %t77, i64 12, 1
  %t79 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_49, i32 0, i32 1, i32 0
  store { ptr, i64 } %t78, ptr %t79
  %t80 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_49
  %t81 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_50, i64 0
  store { i8, { { ptr, i64 } } } %t80, ptr %t81
  %t82 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_50, 0
  %t83 = insertvalue { ptr, i64 } %t82, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t83)
  %t84 = load { ptr, i64 }, ptr %slot_eb_5
  %t85 = load i64, ptr %slot_n_6
  call void @aila_main__show(ptr %t1, { ptr, i64 } %t84, i64 %t85)
  %t86 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t86, ptr %slot___lit_51
  %t87 = insertvalue { ptr, i64 } undef, ptr @.str.10, 0
  %t88 = insertvalue { ptr, i64 } %t87, i64 0, 1
  %t89 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_51, i32 0, i32 1, i32 0
  store { ptr, i64 } %t88, ptr %t89
  %t90 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_51
  %t91 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_52, i64 0
  store { i8, { { ptr, i64 } } } %t90, ptr %t91
  %t92 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_52, 0
  %t93 = insertvalue { ptr, i64 } %t92, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t93)
  br label %m.end15
m.arm117:
  %t94 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_48
  %t95 = extractvalue { i1, i64, { i8, { i64 } } } %t94, 2
  store { i8, { i64 } } %t95, ptr %slot_e_7
  %t96 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t96, ptr %slot___lit_53
  %t97 = insertvalue { ptr, i64 } undef, ptr @.str.11, 0
  %t98 = insertvalue { ptr, i64 } %t97, i64 15, 1
  %t99 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_53, i32 0, i32 1, i32 0
  store { ptr, i64 } %t98, ptr %t99
  %t100 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_53
  %t101 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_54, i64 0
  store { i8, { { ptr, i64 } } } %t100, ptr %t101
  %t102 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_54, 0
  %t103 = insertvalue { ptr, i64 } %t102, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t103)
  %t104 = load { i8, { i64 } }, ptr %slot_e_7
  br label %m.end15
m.end15:
  br label %m.end10
m.arm112:
  %t105 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_47
  %t106 = extractvalue { i1, i64, { i8, { i64 } } } %t105, 2
  store { i8, { i64 } } %t106, ptr %slot_e_8
  %t107 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t107, ptr %slot___lit_55
  %t108 = insertvalue { ptr, i64 } undef, ptr @.str.12, 0
  %t109 = insertvalue { ptr, i64 } %t108, i64 14, 1
  %t110 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_55, i32 0, i32 1, i32 0
  store { ptr, i64 } %t109, ptr %t110
  %t111 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_55
  %t112 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_56, i64 0
  store { i8, { { ptr, i64 } } } %t111, ptr %t112
  %t113 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_56, 0
  %t114 = insertvalue { ptr, i64 } %t113, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t114)
  %t115 = load { i8, { i64 } }, ptr %slot_e_8
  br label %m.end10
m.end10:
  %t116 = insertvalue { ptr, i64 } undef, ptr @.str.13, 0
  %t117 = insertvalue { ptr, i64 } %t116, i64 13, 1
  %t118 = load { ptr, i64 }, ptr %slot_ab_1
  %t119 = call { i1, i64, { i8, { i64 } } } @aila_std_os__env({ ptr, i64 } %t117, { ptr, i64 } %t118)
  store { i1, i64, { i8, { i64 } } } %t119, ptr %slot___head_57
  %t120 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_57
  %t121 = extractvalue { i1, i64, { i8, { i64 } } } %t120, 0
  br i1 %t121, label %m.arm021, label %m.next23
m.next23:
  %t122 = xor i1 %t121, true
  br i1 %t122, label %m.arm122, label %m.next24
m.next24:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 76)
  unreachable
m.arm021:
  %t123 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_57
  %t124 = extractvalue { i1, i64, { i8, { i64 } } } %t123, 1
  store i64 %t124, ptr %slot_n_9
  %t125 = insertvalue { ptr, i64 } undef, ptr @.str.4, 0
  %t126 = insertvalue { ptr, i64 } %t125, i64 15, 1
  %t127 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t127, ptr %slot___lit_58
  %t128 = load i64, ptr %slot_n_9
  %t130 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_58, i32 0, i32 1, i32 0
  store i64 %t128, ptr %t130
  %t131 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_58
  %t132 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_59, i64 0
  store { i8, { { ptr, i64 } } } %t131, ptr %t132
  %t133 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_59, 0
  %t134 = insertvalue { ptr, i64 } %t133, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t126, { ptr, i64 } %t134)
  br label %m.end20
m.arm122:
  %t135 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_57
  %t136 = extractvalue { i1, i64, { i8, { i64 } } } %t135, 2
  store { i8, { i64 } } %t136, ptr %slot_e_10
  %t137 = load { i8, { i64 } }, ptr %slot_e_10
  store { i8, { i64 } } %t137, ptr %slot___head_60
  %t138 = load { i8, { i64 } }, ptr %slot___head_60
  %t139 = extractvalue { i8, { i64 } } %t138, 0
  %t140 = icmp eq i8 %t139, 3
  br i1 %t140, label %m.arm026, label %m.next28
m.next28:
  br label %m.arm127
m.arm026:
  %t141 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t141, ptr %slot___lit_61
  %t142 = insertvalue { ptr, i64 } undef, ptr @.str.14, 0
  %t143 = insertvalue { ptr, i64 } %t142, i64 25, 1
  %t144 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_61, i32 0, i32 1, i32 0
  store { ptr, i64 } %t143, ptr %t144
  %t145 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_61
  %t146 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_62, i64 0
  store { i8, { { ptr, i64 } } } %t145, ptr %t146
  %t147 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_62, 0
  %t148 = insertvalue { ptr, i64 } %t147, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t148)
  br label %m.end25
m.arm127:
  %t149 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t149, ptr %slot___lit_63
  %t150 = insertvalue { ptr, i64 } undef, ptr @.str.6, 0
  %t151 = insertvalue { ptr, i64 } %t150, i64 5, 1
  %t152 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_63, i32 0, i32 1, i32 0
  store { ptr, i64 } %t151, ptr %t152
  %t153 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_63
  %t154 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_64, i64 0
  store { i8, { { ptr, i64 } } } %t153, ptr %t154
  %t155 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_64, 0
  %t156 = insertvalue { ptr, i64 } %t155, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t156)
  br label %m.end25
m.end25:
  br label %m.end20
m.end20:
  %t157 = insertvalue { ptr, i64 } undef, ptr @.str.15, 0
  %t158 = insertvalue { ptr, i64 } %t157, i64 16, 1
  store { ptr, i64 } %t158, ptr %slot_msg_11
  %t159 = insertvalue { ptr, i64 } undef, ptr @.str.16, 0
  %t160 = insertvalue { ptr, i64 } %t159, i64 15, 1
  %t161 = call { i1, { i64 }, { i8, { i64 } } } @aila_std_os__open({ ptr, i64 } %t160, i32 1)
  store { i1, { i64 }, { i8, { i64 } } } %t161, ptr %slot___head_65
  %t162 = load { i1, { i64 }, { i8, { i64 } } }, ptr %slot___head_65
  %t163 = extractvalue { i1, { i64 }, { i8, { i64 } } } %t162, 0
  br i1 %t163, label %m.arm030, label %m.next32
m.next32:
  %t164 = xor i1 %t163, true
  br i1 %t164, label %m.arm131, label %m.next33
m.next33:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 88)
  unreachable
m.arm030:
  %t165 = load { i1, { i64 }, { i8, { i64 } } }, ptr %slot___head_65
  %t166 = extractvalue { i1, { i64 }, { i8, { i64 } } } %t165, 1
  store { i64 } %t166, ptr %slot_f_12
  %t167 = load { i64 }, ptr %slot_f_12
  %t168 = load { ptr, i64 }, ptr %slot_msg_11
  %t169 = call { i1, i64, { i8, { i64 } } } @aila_std_os__write({ i64 } %t167, { ptr, i64 } %t168)
  store { i1, i64, { i8, { i64 } } } %t169, ptr %slot___head_66
  %t170 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_66
  %t171 = extractvalue { i1, i64, { i8, { i64 } } } %t170, 0
  br i1 %t171, label %m.arm035, label %m.next37
m.next37:
  %t172 = xor i1 %t171, true
  br i1 %t172, label %m.arm136, label %m.next38
m.next38:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 90)
  unreachable
m.arm035:
  %t173 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_66
  %t174 = extractvalue { i1, i64, { i8, { i64 } } } %t173, 1
  store i64 %t174, ptr %slot_n_13
  %t175 = insertvalue { ptr, i64 } undef, ptr @.str.17, 0
  %t176 = insertvalue { ptr, i64 } %t175, i64 20, 1
  %t177 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t177, ptr %slot___lit_67
  %t178 = load i64, ptr %slot_n_13
  %t180 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_67, i32 0, i32 1, i32 0
  store i64 %t178, ptr %t180
  %t181 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_67
  %t182 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_68, i64 0
  store { i8, { { ptr, i64 } } } %t181, ptr %t182
  %t183 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_68, 0
  %t184 = insertvalue { ptr, i64 } %t183, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t176, { ptr, i64 } %t184)
  br label %m.end34
m.arm136:
  %t185 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_66
  %t186 = extractvalue { i1, i64, { i8, { i64 } } } %t185, 2
  store { i8, { i64 } } %t186, ptr %slot_e_14
  %t187 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t187, ptr %slot___lit_69
  %t188 = insertvalue { ptr, i64 } undef, ptr @.str.18, 0
  %t189 = insertvalue { ptr, i64 } %t188, i64 12, 1
  %t190 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_69, i32 0, i32 1, i32 0
  store { ptr, i64 } %t189, ptr %t190
  %t191 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_69
  %t192 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_70, i64 0
  store { i8, { { ptr, i64 } } } %t191, ptr %t192
  %t193 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_70, 0
  %t194 = insertvalue { ptr, i64 } %t193, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t194)
  %t195 = load { i8, { i64 } }, ptr %slot_e_14
  br label %m.end34
m.end34:
  %t196 = load { i64 }, ptr %slot_f_12
  call void @aila_std_os__close({ i64 } %t196)
  br label %m.end29
m.arm131:
  %t197 = load { i1, { i64 }, { i8, { i64 } } }, ptr %slot___head_65
  %t198 = extractvalue { i1, { i64 }, { i8, { i64 } } } %t197, 2
  store { i8, { i64 } } %t198, ptr %slot_e_15
  %t199 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t199, ptr %slot___lit_71
  %t200 = insertvalue { ptr, i64 } undef, ptr @.str.19, 0
  %t201 = insertvalue { ptr, i64 } %t200, i64 13, 1
  %t202 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_71, i32 0, i32 1, i32 0
  store { ptr, i64 } %t201, ptr %t202
  %t203 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_71
  %t204 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_72, i64 0
  store { i8, { { ptr, i64 } } } %t203, ptr %t204
  %t205 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_72, 0
  %t206 = insertvalue { ptr, i64 } %t205, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t206)
  %t207 = load { i8, { i64 } }, ptr %slot_e_15
  br label %m.end29
m.end29:
  %t208 = insertvalue { ptr, i64 } undef, ptr @.str.16, 0
  %t209 = insertvalue { ptr, i64 } %t208, i64 15, 1
  %t210 = call { i1, { i64 }, { i8, { i64 } } } @aila_std_os__open({ ptr, i64 } %t209, i32 0)
  store { i1, { i64 }, { i8, { i64 } } } %t210, ptr %slot___head_73
  %t211 = load { i1, { i64 }, { i8, { i64 } } }, ptr %slot___head_73
  %t212 = extractvalue { i1, { i64 }, { i8, { i64 } } } %t211, 0
  br i1 %t212, label %m.arm040, label %m.next42
m.next42:
  %t213 = xor i1 %t212, true
  br i1 %t213, label %m.arm141, label %m.next43
m.next43:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 98)
  unreachable
m.arm040:
  %t214 = load { i1, { i64 }, { i8, { i64 } } }, ptr %slot___head_73
  %t215 = extractvalue { i1, { i64 }, { i8, { i64 } } } %t214, 1
  store { i64 } %t215, ptr %slot_f_16
  %t216 = call ptr @aila_rt_region_alloc(ptr %t1, i64 32, i64 1)
  %t217 = insertvalue { ptr, i64 } undef, ptr %t216, 0
  %t218 = insertvalue { ptr, i64 } %t217, i64 32, 1
  store { ptr, i64 } %t218, ptr %slot_fbuf_17
  %t219 = load { i64 }, ptr %slot_f_16
  %t220 = load { ptr, i64 }, ptr %slot_fbuf_17
  %t221 = call { i1, i64, { i8, { i64 } } } @aila_std_os__read({ i64 } %t219, { ptr, i64 } %t220)
  store { i1, i64, { i8, { i64 } } } %t221, ptr %slot___head_74
  %t222 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_74
  %t223 = extractvalue { i1, i64, { i8, { i64 } } } %t222, 0
  br i1 %t223, label %m.arm045, label %m.next47
m.next47:
  %t224 = xor i1 %t223, true
  br i1 %t224, label %m.arm146, label %m.next48
m.next48:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 101)
  unreachable
m.arm045:
  %t225 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_74
  %t226 = extractvalue { i1, i64, { i8, { i64 } } } %t225, 1
  store i64 %t226, ptr %slot_n_18
  %t227 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t227, ptr %slot___lit_75
  %t228 = insertvalue { ptr, i64 } undef, ptr @.str.20, 0
  %t229 = insertvalue { ptr, i64 } %t228, i64 16, 1
  %t230 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_75, i32 0, i32 1, i32 0
  store { ptr, i64 } %t229, ptr %t230
  %t231 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_75
  %t232 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_76, i64 0
  store { i8, { { ptr, i64 } } } %t231, ptr %t232
  %t233 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_76, 0
  %t234 = insertvalue { ptr, i64 } %t233, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t234)
  %t235 = load { ptr, i64 }, ptr %slot_fbuf_17
  %t236 = load i64, ptr %slot_n_18
  call void @aila_main__show(ptr %t1, { ptr, i64 } %t235, i64 %t236)
  br label %m.end44
m.arm146:
  %t237 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_74
  %t238 = extractvalue { i1, i64, { i8, { i64 } } } %t237, 2
  store { i8, { i64 } } %t238, ptr %slot_e_19
  %t239 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t239, ptr %slot___lit_77
  %t240 = insertvalue { ptr, i64 } undef, ptr @.str.21, 0
  %t241 = insertvalue { ptr, i64 } %t240, i64 11, 1
  %t242 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_77, i32 0, i32 1, i32 0
  store { ptr, i64 } %t241, ptr %t242
  %t243 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_77
  %t244 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_78, i64 0
  store { i8, { { ptr, i64 } } } %t243, ptr %t244
  %t245 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_78, 0
  %t246 = insertvalue { ptr, i64 } %t245, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t246)
  %t247 = load { i8, { i64 } }, ptr %slot_e_19
  br label %m.end44
m.end44:
  %t248 = load { i64 }, ptr %slot_f_16
  %t249 = load { ptr, i64 }, ptr %slot_fbuf_17
  %t250 = call { i1, i64, { i8, { i64 } } } @aila_std_os__read({ i64 } %t248, { ptr, i64 } %t249)
  store { i1, i64, { i8, { i64 } } } %t250, ptr %slot___head_79
  %t251 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_79
  %t252 = extractvalue { i1, i64, { i8, { i64 } } } %t251, 0
  br i1 %t252, label %m.arm050, label %m.next52
m.next52:
  %t253 = xor i1 %t252, true
  br i1 %t253, label %m.arm151, label %m.next53
m.next53:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 108)
  unreachable
m.arm050:
  %t254 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_79
  %t255 = extractvalue { i1, i64, { i8, { i64 } } } %t254, 1
  store i64 %t255, ptr %slot_n_20
  %t256 = insertvalue { ptr, i64 } undef, ptr @.str.22, 0
  %t257 = insertvalue { ptr, i64 } %t256, i64 22, 1
  %t258 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t258, ptr %slot___lit_80
  %t259 = load i64, ptr %slot_n_20
  %t261 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_80, i32 0, i32 1, i32 0
  store i64 %t259, ptr %t261
  %t262 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_80
  %t263 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_81, i64 0
  store { i8, { { ptr, i64 } } } %t262, ptr %t263
  %t264 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_81, 0
  %t265 = insertvalue { ptr, i64 } %t264, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t257, { ptr, i64 } %t265)
  br label %m.end49
m.arm151:
  %t266 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_79
  %t267 = extractvalue { i1, i64, { i8, { i64 } } } %t266, 2
  store { i8, { i64 } } %t267, ptr %slot_e_21
  %t268 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t268, ptr %slot___lit_82
  %t269 = insertvalue { ptr, i64 } undef, ptr @.str.23, 0
  %t270 = insertvalue { ptr, i64 } %t269, i64 12, 1
  %t271 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_82, i32 0, i32 1, i32 0
  store { ptr, i64 } %t270, ptr %t271
  %t272 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_82
  %t273 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_83, i64 0
  store { i8, { { ptr, i64 } } } %t272, ptr %t273
  %t274 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_83, 0
  %t275 = insertvalue { ptr, i64 } %t274, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t275)
  %t276 = load { i8, { i64 } }, ptr %slot_e_21
  br label %m.end49
m.end49:
  %t277 = load { i64 }, ptr %slot_f_16
  call void @aila_std_os__close({ i64 } %t277)
  br label %m.end39
m.arm141:
  %t278 = load { i1, { i64 }, { i8, { i64 } } }, ptr %slot___head_73
  %t279 = extractvalue { i1, { i64 }, { i8, { i64 } } } %t278, 2
  store { i8, { i64 } } %t279, ptr %slot_e_22
  %t280 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t280, ptr %slot___lit_84
  %t281 = insertvalue { ptr, i64 } undef, ptr @.str.24, 0
  %t282 = insertvalue { ptr, i64 } %t281, i64 13, 1
  %t283 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_84, i32 0, i32 1, i32 0
  store { ptr, i64 } %t282, ptr %t283
  %t284 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_84
  %t285 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_85, i64 0
  store { i8, { { ptr, i64 } } } %t284, ptr %t285
  %t286 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_85, 0
  %t287 = insertvalue { ptr, i64 } %t286, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t287)
  %t288 = load { i8, { i64 } }, ptr %slot_e_22
  br label %m.end39
m.end39:
  %t289 = call i64 @aila_std_time__mono_ms()
  store i64 %t289, ptr %slot_t0_23
  call void @aila_std_time__sleep_ms(i64 40)
  %t290 = call i64 @aila_std_time__mono_ms()
  %t291 = load i64, ptr %slot_t0_23
  %t292 = sub i64 %t290, %t291
  store i64 %t292, ptr %slot_dt_24
  %t293 = insertvalue { ptr, i64 } undef, ptr @.str.25, 0
  %t294 = insertvalue { ptr, i64 } %t293, i64 26, 1
  %t295 = insertvalue { i8, { { ptr, i64 } } } undef, i8 3, 0
  store { i8, { { ptr, i64 } } } %t295, ptr %slot___lit_86
  %t296 = load i64, ptr %slot_dt_24
  %t297 = icmp uge i64 %t296, 40
  br i1 %t297, label %sc.r55, label %sc.l54
sc.l54:
  store i1 0, ptr %slot___sc_87
  br label %sc.end56
sc.r55:
  %t298 = load i64, ptr %slot_dt_24
  %t299 = icmp ult i64 %t298, 5000
  store i1 %t299, ptr %slot___sc_87
  br label %sc.end56
sc.end56:
  %t300 = load i1, ptr %slot___sc_87
  %t301 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_86, i32 0, i32 1, i32 0
  store i1 %t300, ptr %t301
  %t302 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_86
  %t303 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_88, i64 0
  store { i8, { { ptr, i64 } } } %t302, ptr %t303
  %t304 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_88, 0
  %t305 = insertvalue { ptr, i64 } %t304, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t294, { ptr, i64 } %t305)
  %t306 = getelementptr inbounds { i64 }, ptr %slot___lit_89, i32 0, i32 0
  store i64 42, ptr %t306
  %t307 = load { i64 }, ptr %slot___lit_89
  store { i64 } %t307, ptr %slot_r1_25
  %t308 = call i64 @aila_std_rand__Rand__next_u64(ptr %slot_r1_25)
  store i64 %t308, ptr %slot_a1_26
  %t309 = call i64 @aila_std_rand__Rand__next_u64(ptr %slot_r1_25)
  store i64 %t309, ptr %slot_a2_27
  %t310 = getelementptr inbounds { i64 }, ptr %slot___lit_90, i32 0, i32 0
  store i64 42, ptr %t310
  %t311 = load { i64 }, ptr %slot___lit_90
  store { i64 } %t311, ptr %slot_r2_28
  %t312 = call i64 @aila_std_rand__Rand__next_u64(ptr %slot_r2_28)
  store i64 %t312, ptr %slot_b1_29
  %t313 = call i64 @aila_std_rand__Rand__next_u64(ptr %slot_r2_28)
  store i64 %t313, ptr %slot_b2_30
  %t314 = insertvalue { ptr, i64 } undef, ptr @.str.26, 0
  %t315 = insertvalue { ptr, i64 } %t314, i64 22, 1
  %t316 = insertvalue { i8, { { ptr, i64 } } } undef, i8 3, 0
  store { i8, { { ptr, i64 } } } %t316, ptr %slot___lit_91
  %t317 = load i64, ptr %slot_a1_26
  %t318 = load i64, ptr %slot_b1_29
  %t319 = icmp eq i64 %t317, %t318
  br i1 %t319, label %sc.r61, label %sc.l60
sc.l60:
  store i1 0, ptr %slot___sc_93
  br label %sc.end62
sc.r61:
  %t320 = load i64, ptr %slot_a2_27
  %t321 = load i64, ptr %slot_b2_30
  %t322 = icmp eq i64 %t320, %t321
  store i1 %t322, ptr %slot___sc_93
  br label %sc.end62
sc.end62:
  %t323 = load i1, ptr %slot___sc_93
  br i1 %t323, label %sc.r58, label %sc.l57
sc.l57:
  store i1 0, ptr %slot___sc_92
  br label %sc.end59
sc.r58:
  %t324 = load i64, ptr %slot_a1_26
  %t325 = load i64, ptr %slot_a2_27
  %t326 = icmp ne i64 %t324, %t325
  store i1 %t326, ptr %slot___sc_92
  br label %sc.end59
sc.end59:
  %t327 = load i1, ptr %slot___sc_92
  %t328 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_91, i32 0, i32 1, i32 0
  store i1 %t327, ptr %t328
  %t329 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_91
  %t330 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_94, i64 0
  store { i8, { { ptr, i64 } } } %t329, ptr %t330
  %t331 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_94, 0
  %t332 = insertvalue { ptr, i64 } %t331, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t315, { ptr, i64 } %t332)
  store i1 1, ptr %slot_inrange_31
  store i64 0, ptr %slot_i_32
  br label %for.cond64
for.cond64:
  %t333 = load i64, ptr %slot_i_32
  %t334 = icmp ult i64 %t333, 8
  br i1 %t334, label %for.body66, label %for.end63
for.body66:
  %t335 = call i64 @aila_std_rand__Rand__below(ptr %slot_r1_25, i64 10)
  %t336 = icmp uge i64 %t335, 10
  br i1 %t336, label %if.then68, label %if.next69
if.then68:
  store i1 0, ptr %slot_inrange_31
  br label %if.end67
if.next69:
  br label %if.end67
if.end67:
  %t337 = load i64, ptr %slot_i_32
  %t338 = add i64 %t337, 1
  store i64 %t338, ptr %slot_i_32
  br label %for.step65
for.step65:
  br label %for.cond64
for.end63:
  %t339 = insertvalue { ptr, i64 } undef, ptr @.str.27, 0
  %t340 = insertvalue { ptr, i64 } %t339, i64 23, 1
  %t341 = insertvalue { i8, { { ptr, i64 } } } undef, i8 3, 0
  store { i8, { { ptr, i64 } } } %t341, ptr %slot___lit_95
  %t342 = load i1, ptr %slot_inrange_31
  %t343 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_95, i32 0, i32 1, i32 0
  store i1 %t342, ptr %t343
  %t344 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_95
  %t345 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_96, i64 0
  store { i8, { { ptr, i64 } } } %t344, ptr %t345
  %t346 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_96, 0
  %t347 = insertvalue { ptr, i64 } %t346, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t340, { ptr, i64 } %t347)
  store i64 0, ptr %slot_s_33
  store i64 0, ptr %slot_e_34
  %t348 = insertvalue { ptr, i64 } undef, ptr @.str.28, 0
  %t349 = insertvalue { ptr, i64 } %t348, i64 9, 1
  %t350 = call { i64, i64 } @aila_std_path__base({ ptr, i64 } %t349, i8 47)
  store { i64, i64 } %t350, ptr %slot___destr_97
  %t351 = getelementptr inbounds { i64, i64 }, ptr %slot___destr_97, i32 0, i32 0
  %t352 = load i64, ptr %t351
  store i64 %t352, ptr %slot_s_33
  %t353 = getelementptr inbounds { i64, i64 }, ptr %slot___destr_97, i32 0, i32 1
  %t354 = load i64, ptr %t353
  store i64 %t354, ptr %slot_e_34
  %t355 = insertvalue { ptr, i64 } undef, ptr @.str.29, 0
  %t356 = insertvalue { ptr, i64 } %t355, i64 29, 1
  %t357 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t357, ptr %slot___lit_98
  %t358 = load i64, ptr %slot_s_33
  %t360 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_98, i32 0, i32 1, i32 0
  store i64 %t358, ptr %t360
  %t361 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_98
  %t362 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t362, ptr %slot___lit_99
  %t363 = load i64, ptr %slot_e_34
  %t365 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_99, i32 0, i32 1, i32 0
  store i64 %t363, ptr %t365
  %t366 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_99
  %t367 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_100, i64 0
  store { i8, { { ptr, i64 } } } %t361, ptr %t367
  %t368 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_100, i64 1
  store { i8, { { ptr, i64 } } } %t366, ptr %t368
  %t369 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_100, 0
  %t370 = insertvalue { ptr, i64 } %t369, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t356, { ptr, i64 } %t370)
  %t371 = insertvalue { ptr, i64 } undef, ptr @.str.28, 0
  %t372 = insertvalue { ptr, i64 } %t371, i64 9, 1
  %t373 = call { i64, i64 } @aila_std_path__dir({ ptr, i64 } %t372, i8 47)
  store { i64, i64 } %t373, ptr %slot___destr_101
  %t374 = getelementptr inbounds { i64, i64 }, ptr %slot___destr_101, i32 0, i32 0
  %t375 = load i64, ptr %t374
  store i64 %t375, ptr %slot_s_33
  %t376 = getelementptr inbounds { i64, i64 }, ptr %slot___destr_101, i32 0, i32 1
  %t377 = load i64, ptr %t376
  store i64 %t377, ptr %slot_e_34
  %t378 = insertvalue { ptr, i64 } undef, ptr @.str.30, 0
  %t379 = insertvalue { ptr, i64 } %t378, i64 28, 1
  %t380 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t380, ptr %slot___lit_102
  %t381 = load i64, ptr %slot_s_33
  %t383 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_102, i32 0, i32 1, i32 0
  store i64 %t381, ptr %t383
  %t384 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_102
  %t385 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t385, ptr %slot___lit_103
  %t386 = load i64, ptr %slot_e_34
  %t388 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_103, i32 0, i32 1, i32 0
  store i64 %t386, ptr %t388
  %t389 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_103
  %t390 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_104, i64 0
  store { i8, { { ptr, i64 } } } %t384, ptr %t390
  %t391 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_104, i64 1
  store { i8, { { ptr, i64 } } } %t389, ptr %t391
  %t392 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_104, 0
  %t393 = insertvalue { ptr, i64 } %t392, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t379, { ptr, i64 } %t393)
  %t394 = insertvalue { ptr, i64 } undef, ptr @.str.28, 0
  %t395 = insertvalue { ptr, i64 } %t394, i64 9, 1
  %t396 = call { i64, i64 } @aila_std_path__ext({ ptr, i64 } %t395)
  store { i64, i64 } %t396, ptr %slot___destr_105
  %t397 = getelementptr inbounds { i64, i64 }, ptr %slot___destr_105, i32 0, i32 0
  %t398 = load i64, ptr %t397
  store i64 %t398, ptr %slot_s_33
  %t399 = getelementptr inbounds { i64, i64 }, ptr %slot___destr_105, i32 0, i32 1
  %t400 = load i64, ptr %t399
  store i64 %t400, ptr %slot_e_34
  %t401 = insertvalue { ptr, i64 } undef, ptr @.str.31, 0
  %t402 = insertvalue { ptr, i64 } %t401, i64 28, 1
  %t403 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t403, ptr %slot___lit_106
  %t404 = load i64, ptr %slot_s_33
  %t406 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_106, i32 0, i32 1, i32 0
  store i64 %t404, ptr %t406
  %t407 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_106
  %t408 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t408, ptr %slot___lit_107
  %t409 = load i64, ptr %slot_e_34
  %t411 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_107, i32 0, i32 1, i32 0
  store i64 %t409, ptr %t411
  %t412 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_107
  %t413 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_108, i64 0
  store { i8, { { ptr, i64 } } } %t407, ptr %t413
  %t414 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_108, i64 1
  store { i8, { { ptr, i64 } } } %t412, ptr %t414
  %t415 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_108, 0
  %t416 = insertvalue { ptr, i64 } %t415, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t402, { ptr, i64 } %t416)
  %t417 = insertvalue { ptr, i64 } undef, ptr @.str.32, 0
  %t418 = insertvalue { ptr, i64 } %t417, i64 5, 1
  %t419 = call { i64, i64 } @aila_std_path__ext({ ptr, i64 } %t418)
  store { i64, i64 } %t419, ptr %slot___destr_109
  %t420 = getelementptr inbounds { i64, i64 }, ptr %slot___destr_109, i32 0, i32 0
  %t421 = load i64, ptr %t420
  store i64 %t421, ptr %slot_s_33
  %t422 = getelementptr inbounds { i64, i64 }, ptr %slot___destr_109, i32 0, i32 1
  %t423 = load i64, ptr %t422
  store i64 %t423, ptr %slot_e_34
  %t424 = insertvalue { ptr, i64 } undef, ptr @.str.33, 0
  %t425 = insertvalue { ptr, i64 } %t424, i64 24, 1
  %t426 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t426, ptr %slot___lit_110
  %t427 = load i64, ptr %slot_s_33
  %t429 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_110, i32 0, i32 1, i32 0
  store i64 %t427, ptr %t429
  %t430 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_110
  %t431 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t431, ptr %slot___lit_111
  %t432 = load i64, ptr %slot_e_34
  %t434 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_111, i32 0, i32 1, i32 0
  store i64 %t432, ptr %t434
  %t435 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_111
  %t436 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_112, i64 0
  store { i8, { { ptr, i64 } } } %t430, ptr %t436
  %t437 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_112, i64 1
  store { i8, { { ptr, i64 } } } %t435, ptr %t437
  %t438 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_112, 0
  %t439 = insertvalue { ptr, i64 } %t438, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t425, { ptr, i64 } %t439)
  %t440 = call ptr @aila_rt_region_alloc(ptr %t1, i64 32, i64 1)
  %t441 = insertvalue { ptr, i64 } undef, ptr %t440, 0
  %t442 = insertvalue { ptr, i64 } %t441, i64 32, 1
  store { ptr, i64 } %t442, ptr %slot_pbuf_35
  %t443 = insertvalue { ptr, i64 } undef, ptr @.str.34, 0
  %t444 = insertvalue { ptr, i64 } %t443, i64 1, 1
  %t445 = insertvalue { ptr, i64 } undef, ptr @.str.35, 0
  %t446 = insertvalue { ptr, i64 } %t445, i64 1, 1
  %t447 = load { ptr, i64 }, ptr %slot_pbuf_35
  %t448 = call i64 @aila_std_path__join({ ptr, i64 } %t444, { ptr, i64 } %t446, i8 47, { ptr, i64 } %t447)
  store i64 %t448, ptr %slot_w_36
  %t449 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t449, ptr %slot___lit_113
  %t450 = insertvalue { ptr, i64 } undef, ptr @.str.36, 0
  %t451 = insertvalue { ptr, i64 } %t450, i64 17, 1
  %t452 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_113, i32 0, i32 1, i32 0
  store { ptr, i64 } %t451, ptr %t452
  %t453 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_113
  %t454 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_114, i64 0
  store { i8, { { ptr, i64 } } } %t453, ptr %t454
  %t455 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_114, 0
  %t456 = insertvalue { ptr, i64 } %t455, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t456)
  %t457 = load { ptr, i64 }, ptr %slot_pbuf_35
  %t458 = load i64, ptr %slot_w_36
  call void @aila_main__show(ptr %t1, { ptr, i64 } %t457, i64 %t458)
  %t459 = insertvalue { ptr, i64 } undef, ptr @.str.10, 0
  %t460 = insertvalue { ptr, i64 } %t459, i64 0, 1
  %t461 = insertvalue { ptr, i64 } undef, ptr @.str.35, 0
  %t462 = insertvalue { ptr, i64 } %t461, i64 1, 1
  %t463 = load { ptr, i64 }, ptr %slot_pbuf_35
  %t464 = call i64 @aila_std_path__join({ ptr, i64 } %t460, { ptr, i64 } %t462, i8 47, { ptr, i64 } %t463)
  store i64 %t464, ptr %slot_w_36
  %t465 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t465, ptr %slot___lit_115
  %t466 = insertvalue { ptr, i64 } undef, ptr @.str.37, 0
  %t467 = insertvalue { ptr, i64 } %t466, i64 17, 1
  %t468 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_115, i32 0, i32 1, i32 0
  store { ptr, i64 } %t467, ptr %t468
  %t469 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_115
  %t470 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_116, i64 0
  store { i8, { { ptr, i64 } } } %t469, ptr %t470
  %t471 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_116, 0
  %t472 = insertvalue { ptr, i64 } %t471, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t472)
  %t473 = load { ptr, i64 }, ptr %slot_pbuf_35
  %t474 = load i64, ptr %slot_w_36
  call void @aila_main__show(ptr %t1, { ptr, i64 } %t473, i64 %t474)
  %t475 = insertvalue { ptr, i64 } undef, ptr @.str.34, 0
  %t476 = insertvalue { ptr, i64 } %t475, i64 1, 1
  %t477 = insertvalue { ptr, i64 } undef, ptr @.str.38, 0
  %t478 = insertvalue { ptr, i64 } %t477, i64 4, 1
  %t479 = load { ptr, i64 }, ptr %slot_pbuf_35
  %t480 = call i64 @aila_std_path__join({ ptr, i64 } %t476, { ptr, i64 } %t478, i8 47, { ptr, i64 } %t479)
  store i64 %t480, ptr %slot_w_36
  %t481 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t481, ptr %slot___lit_117
  %t482 = insertvalue { ptr, i64 } undef, ptr @.str.39, 0
  %t483 = insertvalue { ptr, i64 } %t482, i64 21, 1
  %t484 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_117, i32 0, i32 1, i32 0
  store { ptr, i64 } %t483, ptr %t484
  %t485 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_117
  %t486 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_118, i64 0
  store { i8, { { ptr, i64 } } } %t485, ptr %t486
  %t487 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_118, 0
  %t488 = insertvalue { ptr, i64 } %t487, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t488)
  %t489 = load { ptr, i64 }, ptr %slot_pbuf_35
  %t490 = load i64, ptr %slot_w_36
  call void @aila_main__show(ptr %t1, { ptr, i64 } %t489, i64 %t490)
  %t491 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t491, ptr %slot___lit_119
  %t492 = insertvalue { ptr, i64 } undef, ptr @.str.10, 0
  %t493 = insertvalue { ptr, i64 } %t492, i64 0, 1
  %t494 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_119, i32 0, i32 1, i32 0
  store { ptr, i64 } %t493, ptr %t494
  %t495 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_119
  %t496 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_120, i64 0
  store { i8, { { ptr, i64 } } } %t495, ptr %t496
  %t497 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_120, 0
  %t498 = insertvalue { ptr, i64 } %t497, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t498)
  call void @aila_rt_region_leave(ptr %t1)
  ret void
}

define i32 @main(i32 %argc, ptr %argv) {
entry:
  call void @aila_rt_args_save(i32 %argc, ptr %argv)
  call void @aila_main__main()
  ret i32 0
}
