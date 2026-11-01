; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [20 x i8] c"index out of bounds\00", align 1
@.str.1 = private unnamed_addr constant [10 x i8] c"main.aila\00", align 1
@.str.2 = private unnamed_addr constant [6 x i8] c"  -42\00", align 1
@.str.3 = private unnamed_addr constant [28 x i8] c"match \E6\9C\AA\E8\A6\86\E7\9B\96\E7\9A\84\E5\88\A4\E5\88\AB\E5\80\BC\00", align 1
@.str.4 = private unnamed_addr constant [25 x i8] c"parse_i64(\22  -42\22) = %d\0A\00", align 1
@.str.5 = private unnamed_addr constant [11 x i8] c"unexpected\00", align 1
@.str.6 = private unnamed_addr constant [4 x i8] c"12x\00", align 1
@.str.7 = private unnamed_addr constant [36 x i8] c"parse_i64(\2212x\22) -> Invalid{at=%u}\0A\00", align 1
@.str.8 = private unnamed_addr constant [6 x i8] c"other\00", align 1
@.str.9 = private unnamed_addr constant [20 x i8] c"9223372036854775808\00", align 1
@.str.10 = private unnamed_addr constant [42 x i8] c"parse_i64(\229223372036854775808\22) -> Range\00", align 1
@.str.11 = private unnamed_addr constant [21 x i8] c"18446744073709551615\00", align 1
@.str.12 = private unnamed_addr constant [21 x i8] c"parse_u64(max) = %u\0A\00", align 1
@.str.13 = private unnamed_addr constant [7 x i8] c"2.5e-1\00", align 1
@.str.14 = private unnamed_addr constant [26 x i8] c"parse_f64(\222.5e-1\22) = %f\0A\00", align 1
@.str.15 = private unnamed_addr constant [19 x i8] c"format_i64(-42) = \00", align 1
@.str.16 = private unnamed_addr constant [13 x i8] c" (%u bytes)\0A\00", align 1
@.str.17 = private unnamed_addr constant [19 x i8] c"format_u64(max) = \00", align 1
@.str.18 = private unnamed_addr constant [19 x i8] c"format_f64(0.5) = \00", align 1
@.str.19 = private unnamed_addr constant [1 x i8] c"\00", align 1
@.str.20 = private unnamed_addr constant [35 x i8] c"index(\22hello, AILA\22, \22AILA\22) = %u\0A\00", align 1
@.str.21 = private unnamed_addr constant [12 x i8] c"hello, AILA\00", align 1
@.str.22 = private unnamed_addr constant [5 x i8] c"AILA\00", align 1
@.str.23 = private unnamed_addr constant [37 x i8] c"contains(\22hello, AILA\22, \22zzz\22) = %b\0A\00", align 1
@.str.24 = private unnamed_addr constant [4 x i8] c"zzz\00", align 1
@.str.25 = private unnamed_addr constant [64 x i8] c"has_prefix(\22hello\22, \22he\22) = %b  has_suffix(\22hello\22, \22lo\22) = %b\0A\00", align 1
@.str.26 = private unnamed_addr constant [6 x i8] c"hello\00", align 1
@.str.27 = private unnamed_addr constant [3 x i8] c"he\00", align 1
@.str.28 = private unnamed_addr constant [3 x i8] c"lo\00", align 1
@.str.29 = private unnamed_addr constant [52 x i8] c"len_runes(\22a\E4\B8\ADb\22) = %u  cp_at(1) = %u  valid = %b\0A\00", align 1
@.str.30 = private unnamed_addr constant [6 x i8] c"a\E4\B8\ADb\00", align 1
@.str.31 = private unnamed_addr constant [29 x i8] c"cp_at(\E8\B6\8A\E7\95\8C) = %u (\E5\93\A8\E5\85\B5)\0A\00", align 1
@.str.32 = private unnamed_addr constant [3 x i8] c"ab\00", align 1
@.str.33 = private unnamed_addr constant [7 x i8] c"  hi  \00", align 1
@.str.34 = private unnamed_addr constant [29 x i8] c"bounds(\22  hi  \22) = [%u, %u)\0A\00", align 1
@.str.35 = private unnamed_addr constant [4 x i8] c"abc\00", align 1
@.str.36 = private unnamed_addr constant [3 x i8] c"42\00", align 1
@.str.37 = private unnamed_addr constant [11 x i8] c"concat -> \00", align 1
@.str.38 = private unnamed_addr constant [21 x i8] c" (%u, cap 3 \E6\88\AA\E6\96\AD)\0A\00", align 1
@.str.39 = private unnamed_addr constant [7 x i8] c" (%u)\0A\00", align 1
@.str.40 = private unnamed_addr constant [6 x i8] c"a,b,c\00", align 1
@.str.41 = private unnamed_addr constant [18 x i8] c"split(\22a,b,c\22) = \00", align 1
@.str.42 = private unnamed_addr constant [13 x i8] c" \E6\AE\B5 marks: \00", align 1
@.str.43 = private unnamed_addr constant [2 x i8] c" \00", align 1
@.str.44 = private unnamed_addr constant [2 x i8] c"x\00", align 1
@.str.45 = private unnamed_addr constant [2 x i8] c"y\00", align 1
@.str.46 = private unnamed_addr constant [2 x i8] c"z\00", align 1
@.str.47 = private unnamed_addr constant [9 x i8] c"join -> \00", align 1
@.str.48 = private unnamed_addr constant [6 x i8] c"a?b?c\00", align 1
@.str.49 = private unnamed_addr constant [2 x i8] c"?\00", align 1
@.str.50 = private unnamed_addr constant [2 x i8] c"!\00", align 1
@.str.51 = private unnamed_addr constant [12 x i8] c"replace -> \00", align 1
@.str.52 = private unnamed_addr constant [5 x i8] c"aila\00", align 1
@.str.53 = private unnamed_addr constant [10 x i8] c"upper -> \00", align 1
@.str.54 = private unnamed_addr constant [12 x i8] c"  lower -> \00", align 1
@.str.55 = private unnamed_addr constant [26 x i8] c"le/be roundtrip: %u / %u\0A\00", align 1
@.str.56 = private unnamed_addr constant [23 x i8] c"i16 -2 roundtrip = %d\0A\00", align 1
@.str.57 = private unnamed_addr constant [11 x i8] c"sort asc: \00", align 1
@.str.58 = private unnamed_addr constant [10 x i8] c"   desc: \00", align 1

declare ptr @aila_rt_region_alloc(ptr, i64, i64)
declare void @aila_rt_trap(ptr, i64, ptr, i64)
declare { ptr, { ptr, i64 } } @aila_std_io__stdout()
declare i64 @aila_std_io__Writer__put(ptr, { ptr, i64 })
declare ptr @aila_rt_region_enter(i64)
declare { i1, i64, { i8, { i64 } } } @aila_std_strconv__parse_i64({ ptr, i64 })
declare void @aila_std_io__printf({ ptr, i64 }, { ptr, i64 })
declare void @aila_std_io__println({ ptr, i64 })
declare { i1, i64, { i8, { i64 } } } @aila_std_strconv__parse_u64({ ptr, i64 })
declare { i1, double, { i8, { i64 } } } @aila_std_strconv__parse_f64({ ptr, i64 })
declare i64 @aila_std_strconv__format_i64(i64, { ptr, i64 })
declare void @aila_std_io__print({ ptr, i64 })
declare i64 @aila_std_strconv__format_u64(i64, { ptr, i64 })
declare i64 @aila_std_strconv__format_f64(double, { ptr, i64 })
declare i64 @aila_std_strings__index({ ptr, i64 }, { ptr, i64 })
declare i1 @aila_std_strings__contains({ ptr, i64 }, { ptr, i64 })
declare i1 @aila_std_strings__has_prefix({ ptr, i64 }, { ptr, i64 })
declare i1 @aila_std_strings__has_suffix({ ptr, i64 }, { ptr, i64 })
declare i64 @aila_std_strings__len_runes({ ptr, i64 })
declare i32 @aila_std_utf8__cp_at({ ptr, i64 }, i64)
declare i1 @aila_std_utf8__valid({ ptr, i64 })
declare { i64, i64 } @aila_std_strings__bounds({ ptr, i64 })
declare i64 @aila_std_strings__concat({ ptr, i64 }, { ptr, i64 }, { ptr, i64 })
declare i64 @aila_std_strings__split({ ptr, i64 }, i8, { ptr, i64 })
declare i64 @aila_std_strings__join({ ptr, i64 }, i8, { ptr, i64 })
declare i64 @aila_std_strings__replace({ ptr, i64 }, { ptr, i64 }, { ptr, i64 }, { ptr, i64 })
declare i64 @aila_std_strings__upper_ascii({ ptr, i64 }, { ptr, i64 })
declare i64 @aila_std_strings__lower_ascii({ ptr, i64 }, { ptr, i64 })
declare void @aila_std_binary__le_write_u64({ ptr, i64 }, i64, i64)
declare i64 @aila_std_binary__le_read_u64({ ptr, i64 }, i64)
declare i64 @aila_std_binary__be_read_u64({ ptr, i64 }, i64)
declare void @aila_std_binary__le_write_i16({ ptr, i64 }, i64, i16)
declare i16 @aila_std_binary__le_read_i16({ ptr, i64 }, i64)
declare void @aila_std_sort__sort_i64({ ptr, i64 })
declare void @aila_std_sort__sort_i64_by({ ptr, i64 }, ptr)
declare void @aila_rt_region_leave(ptr)

define i1 @aila_main__desc_i64(i64 %p_a, i64 %p_b) {
  %slot_a_1 = alloca i64, align 8
  %slot_b_2 = alloca i64, align 8
  store i64 %p_a, ptr %slot_a_1
  store i64 %p_b, ptr %slot_b_2
  %t1 = load i64, ptr %slot_a_1
  %t2 = load i64, ptr %slot_b_2
  %t3 = icmp sgt i64 %t1, %t2
  ret i1 %t3
}

define void @aila_main__show_bytes(ptr %r0, { ptr, i64 } %p_buf, i64 %p_n) {
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
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 50)
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
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 50)
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
  %slot_v_1 = alloca i64, align 8
  %slot_e_2 = alloca { i8, { i64 } }, align 8
  %slot_v_3 = alloca i64, align 8
  %slot_e_4 = alloca { i8, { i64 } }, align 8
  %slot_at_5 = alloca i64, align 8
  %slot_v_6 = alloca i64, align 8
  %slot_e_7 = alloca { i8, { i64 } }, align 8
  %slot_v_8 = alloca i64, align 8
  %slot_e_9 = alloca { i8, { i64 } }, align 8
  %slot_v_10 = alloca double, align 8
  %slot_e_11 = alloca { i8, { i64 } }, align 8
  %slot_fbuf_12 = alloca { ptr, i64 }, align 8
  %slot_n_13 = alloca i64, align 8
  %slot_s_14 = alloca i64, align 8
  %slot_e_15 = alloca i64, align 8
  %slot_small_16 = alloca { ptr, i64 }, align 8
  %slot_buf_17 = alloca { ptr, i64 }, align 8
  %slot_marks_18 = alloca { ptr, i64 }, align 8
  %slot_segs_19 = alloca i64, align 8
  %slot_k_20 = alloca i64, align 8
  %slot_bbuf_21 = alloca { ptr, i64 }, align 8
  %slot_a_22 = alloca { ptr, i64 }, align 8
  %slot___head_23 = alloca { i1, i64, { i8, { i64 } } }, align 8
  %slot___lit_24 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_25 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_26 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_27 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_28 = alloca { i1, i64, { i8, { i64 } } }, align 8
  %slot___lit_29 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_30 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_31 = alloca { i8, { i64 } }, align 8
  %slot___lit_32 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_33 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_34 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_35 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_36 = alloca { i1, i64, { i8, { i64 } } }, align 8
  %slot___lit_37 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_38 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_39 = alloca { i8, { i64 } }, align 8
  %slot___lit_40 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_41 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_42 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_43 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_44 = alloca { i1, i64, { i8, { i64 } } }, align 8
  %slot___lit_45 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_46 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_47 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_48 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___head_49 = alloca { i1, double, { i8, { i64 } } }, align 8
  %slot___lit_50 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_51 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_52 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_53 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_54 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_55 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_56 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_57 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_58 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_59 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_60 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_61 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_62 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_63 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_64 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_65 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_66 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_67 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_68 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_69 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_70 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___lit_71 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_72 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_73 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___lit_74 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___lit_75 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_76 = alloca [3 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_77 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_78 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___destr_79 = alloca { i64, i64 }, align 8
  %slot___lit_80 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___lit_81 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_82 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_83 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_84 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_85 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_86 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_87 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_88 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_89 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_90 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_91 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_92 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_93 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_94 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_95 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_96 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_97 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_98 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_99 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_100 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_101 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_102 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot_vargs_103 = alloca [3 x { ptr, i64 }], align 8
  %slot___lit_104 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_105 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_106 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_107 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_108 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_109 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_110 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_111 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_112 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_113 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_114 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_115 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_116 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_117 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_118 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot___lit_119 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_120 = alloca [2 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_121 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_122 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_123 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_124 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_125 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_126 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_127 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_128 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_129 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_130 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_131 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_132 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_133 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_134 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %slot___lit_135 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_136 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %t1 = call ptr @aila_rt_region_enter(i64 0)
  %t2 = insertvalue { ptr, i64 } undef, ptr @.str.2, 0
  %t3 = insertvalue { ptr, i64 } %t2, i64 5, 1
  %t4 = call { i1, i64, { i8, { i64 } } } @aila_std_strconv__parse_i64({ ptr, i64 } %t3)
  store { i1, i64, { i8, { i64 } } } %t4, ptr %slot___head_23
  %t5 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_23
  %t6 = extractvalue { i1, i64, { i8, { i64 } } } %t5, 0
  br i1 %t6, label %m.arm02, label %m.next4
m.next4:
  %t7 = xor i1 %t6, true
  br i1 %t7, label %m.arm13, label %m.next5
m.next5:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 59)
  unreachable
m.arm02:
  %t8 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_23
  %t9 = extractvalue { i1, i64, { i8, { i64 } } } %t8, 1
  store i64 %t9, ptr %slot_v_1
  %t10 = insertvalue { ptr, i64 } undef, ptr @.str.4, 0
  %t11 = insertvalue { ptr, i64 } %t10, i64 24, 1
  %t12 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t12, ptr %slot___lit_24
  %t13 = load i64, ptr %slot_v_1
  %t14 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_24, i32 0, i32 1, i32 0
  store i64 %t13, ptr %t14
  %t15 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_24
  %t16 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_25, i64 0
  store { i8, { { ptr, i64 } } } %t15, ptr %t16
  %t17 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_25, 0
  %t18 = insertvalue { ptr, i64 } %t17, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t11, { ptr, i64 } %t18)
  br label %m.end1
m.arm13:
  %t19 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_23
  %t20 = extractvalue { i1, i64, { i8, { i64 } } } %t19, 2
  store { i8, { i64 } } %t20, ptr %slot_e_2
  %t21 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t21, ptr %slot___lit_26
  %t22 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t23 = insertvalue { ptr, i64 } %t22, i64 10, 1
  %t24 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_26, i32 0, i32 1, i32 0
  store { ptr, i64 } %t23, ptr %t24
  %t25 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_26
  %t26 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_27, i64 0
  store { i8, { { ptr, i64 } } } %t25, ptr %t26
  %t27 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_27, 0
  %t28 = insertvalue { ptr, i64 } %t27, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t28)
  %t29 = load { i8, { i64 } }, ptr %slot_e_2
  br label %m.end1
m.end1:
  %t30 = insertvalue { ptr, i64 } undef, ptr @.str.6, 0
  %t31 = insertvalue { ptr, i64 } %t30, i64 3, 1
  %t32 = call { i1, i64, { i8, { i64 } } } @aila_std_strconv__parse_i64({ ptr, i64 } %t31)
  store { i1, i64, { i8, { i64 } } } %t32, ptr %slot___head_28
  %t33 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_28
  %t34 = extractvalue { i1, i64, { i8, { i64 } } } %t33, 0
  br i1 %t34, label %m.arm07, label %m.next9
m.next9:
  %t35 = xor i1 %t34, true
  br i1 %t35, label %m.arm18, label %m.next10
m.next10:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 63)
  unreachable
m.arm07:
  %t36 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_28
  %t37 = extractvalue { i1, i64, { i8, { i64 } } } %t36, 1
  store i64 %t37, ptr %slot_v_3
  %t38 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t38, ptr %slot___lit_29
  %t39 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t40 = insertvalue { ptr, i64 } %t39, i64 10, 1
  %t41 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_29, i32 0, i32 1, i32 0
  store { ptr, i64 } %t40, ptr %t41
  %t42 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_29
  %t43 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_30, i64 0
  store { i8, { { ptr, i64 } } } %t42, ptr %t43
  %t44 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_30, 0
  %t45 = insertvalue { ptr, i64 } %t44, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t45)
  %t46 = load i64, ptr %slot_v_3
  br label %m.end6
m.arm18:
  %t47 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_28
  %t48 = extractvalue { i1, i64, { i8, { i64 } } } %t47, 2
  store { i8, { i64 } } %t48, ptr %slot_e_4
  %t49 = load { i8, { i64 } }, ptr %slot_e_4
  store { i8, { i64 } } %t49, ptr %slot___head_31
  %t50 = load { i8, { i64 } }, ptr %slot___head_31
  %t51 = extractvalue { i8, { i64 } } %t50, 0
  %t52 = icmp eq i8 %t51, 1
  br i1 %t52, label %m.arm012, label %m.next14
m.next14:
  br label %m.arm113
m.arm012:
  %t53 = getelementptr inbounds { i8, { i64 } }, ptr %slot___head_31, i32 0, i32 1, i32 0
  %t54 = load i64, ptr %t53
  store i64 %t54, ptr %slot_at_5
  %t55 = insertvalue { ptr, i64 } undef, ptr @.str.7, 0
  %t56 = insertvalue { ptr, i64 } %t55, i64 35, 1
  %t57 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t57, ptr %slot___lit_32
  %t58 = load i64, ptr %slot_at_5
  %t60 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_32, i32 0, i32 1, i32 0
  store i64 %t58, ptr %t60
  %t61 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_32
  %t62 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_33, i64 0
  store { i8, { { ptr, i64 } } } %t61, ptr %t62
  %t63 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_33, 0
  %t64 = insertvalue { ptr, i64 } %t63, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t56, { ptr, i64 } %t64)
  br label %m.end11
m.arm113:
  %t65 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t65, ptr %slot___lit_34
  %t66 = insertvalue { ptr, i64 } undef, ptr @.str.8, 0
  %t67 = insertvalue { ptr, i64 } %t66, i64 5, 1
  %t68 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_34, i32 0, i32 1, i32 0
  store { ptr, i64 } %t67, ptr %t68
  %t69 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_34
  %t70 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_35, i64 0
  store { i8, { { ptr, i64 } } } %t69, ptr %t70
  %t71 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_35, 0
  %t72 = insertvalue { ptr, i64 } %t71, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t72)
  br label %m.end11
m.end11:
  br label %m.end6
m.end6:
  %t73 = insertvalue { ptr, i64 } undef, ptr @.str.9, 0
  %t74 = insertvalue { ptr, i64 } %t73, i64 19, 1
  %t75 = call { i1, i64, { i8, { i64 } } } @aila_std_strconv__parse_i64({ ptr, i64 } %t74)
  store { i1, i64, { i8, { i64 } } } %t75, ptr %slot___head_36
  %t76 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_36
  %t77 = extractvalue { i1, i64, { i8, { i64 } } } %t76, 0
  br i1 %t77, label %m.arm016, label %m.next18
m.next18:
  %t78 = xor i1 %t77, true
  br i1 %t78, label %m.arm117, label %m.next19
m.next19:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 75)
  unreachable
m.arm016:
  %t79 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_36
  %t80 = extractvalue { i1, i64, { i8, { i64 } } } %t79, 1
  store i64 %t80, ptr %slot_v_6
  %t81 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t81, ptr %slot___lit_37
  %t82 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t83 = insertvalue { ptr, i64 } %t82, i64 10, 1
  %t84 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_37, i32 0, i32 1, i32 0
  store { ptr, i64 } %t83, ptr %t84
  %t85 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_37
  %t86 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_38, i64 0
  store { i8, { { ptr, i64 } } } %t85, ptr %t86
  %t87 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_38, 0
  %t88 = insertvalue { ptr, i64 } %t87, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t88)
  %t89 = load i64, ptr %slot_v_6
  br label %m.end15
m.arm117:
  %t90 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_36
  %t91 = extractvalue { i1, i64, { i8, { i64 } } } %t90, 2
  store { i8, { i64 } } %t91, ptr %slot_e_7
  %t92 = load { i8, { i64 } }, ptr %slot_e_7
  store { i8, { i64 } } %t92, ptr %slot___head_39
  %t93 = load { i8, { i64 } }, ptr %slot___head_39
  %t94 = extractvalue { i8, { i64 } } %t93, 0
  %t95 = icmp eq i8 %t94, 2
  br i1 %t95, label %m.arm021, label %m.next23
m.next23:
  br label %m.arm122
m.arm021:
  %t96 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t96, ptr %slot___lit_40
  %t97 = insertvalue { ptr, i64 } undef, ptr @.str.10, 0
  %t98 = insertvalue { ptr, i64 } %t97, i64 41, 1
  %t99 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_40, i32 0, i32 1, i32 0
  store { ptr, i64 } %t98, ptr %t99
  %t100 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_40
  %t101 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_41, i64 0
  store { i8, { { ptr, i64 } } } %t100, ptr %t101
  %t102 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_41, 0
  %t103 = insertvalue { ptr, i64 } %t102, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t103)
  br label %m.end20
m.arm122:
  %t104 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t104, ptr %slot___lit_42
  %t105 = insertvalue { ptr, i64 } undef, ptr @.str.8, 0
  %t106 = insertvalue { ptr, i64 } %t105, i64 5, 1
  %t107 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_42, i32 0, i32 1, i32 0
  store { ptr, i64 } %t106, ptr %t107
  %t108 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_42
  %t109 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_43, i64 0
  store { i8, { { ptr, i64 } } } %t108, ptr %t109
  %t110 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_43, 0
  %t111 = insertvalue { ptr, i64 } %t110, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t111)
  br label %m.end20
m.end20:
  br label %m.end15
m.end15:
  %t112 = insertvalue { ptr, i64 } undef, ptr @.str.11, 0
  %t113 = insertvalue { ptr, i64 } %t112, i64 20, 1
  %t114 = call { i1, i64, { i8, { i64 } } } @aila_std_strconv__parse_u64({ ptr, i64 } %t113)
  store { i1, i64, { i8, { i64 } } } %t114, ptr %slot___head_44
  %t115 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_44
  %t116 = extractvalue { i1, i64, { i8, { i64 } } } %t115, 0
  br i1 %t116, label %m.arm025, label %m.next27
m.next27:
  %t117 = xor i1 %t116, true
  br i1 %t117, label %m.arm126, label %m.next28
m.next28:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 86)
  unreachable
m.arm025:
  %t118 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_44
  %t119 = extractvalue { i1, i64, { i8, { i64 } } } %t118, 1
  store i64 %t119, ptr %slot_v_8
  %t120 = insertvalue { ptr, i64 } undef, ptr @.str.12, 0
  %t121 = insertvalue { ptr, i64 } %t120, i64 20, 1
  %t122 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t122, ptr %slot___lit_45
  %t123 = load i64, ptr %slot_v_8
  %t124 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_45, i32 0, i32 1, i32 0
  store i64 %t123, ptr %t124
  %t125 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_45
  %t126 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_46, i64 0
  store { i8, { { ptr, i64 } } } %t125, ptr %t126
  %t127 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_46, 0
  %t128 = insertvalue { ptr, i64 } %t127, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t121, { ptr, i64 } %t128)
  br label %m.end24
m.arm126:
  %t129 = load { i1, i64, { i8, { i64 } } }, ptr %slot___head_44
  %t130 = extractvalue { i1, i64, { i8, { i64 } } } %t129, 2
  store { i8, { i64 } } %t130, ptr %slot_e_9
  %t131 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t131, ptr %slot___lit_47
  %t132 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t133 = insertvalue { ptr, i64 } %t132, i64 10, 1
  %t134 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_47, i32 0, i32 1, i32 0
  store { ptr, i64 } %t133, ptr %t134
  %t135 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_47
  %t136 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_48, i64 0
  store { i8, { { ptr, i64 } } } %t135, ptr %t136
  %t137 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_48, 0
  %t138 = insertvalue { ptr, i64 } %t137, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t138)
  %t139 = load { i8, { i64 } }, ptr %slot_e_9
  br label %m.end24
m.end24:
  %t140 = insertvalue { ptr, i64 } undef, ptr @.str.13, 0
  %t141 = insertvalue { ptr, i64 } %t140, i64 6, 1
  %t142 = call { i1, double, { i8, { i64 } } } @aila_std_strconv__parse_f64({ ptr, i64 } %t141)
  store { i1, double, { i8, { i64 } } } %t142, ptr %slot___head_49
  %t143 = load { i1, double, { i8, { i64 } } }, ptr %slot___head_49
  %t144 = extractvalue { i1, double, { i8, { i64 } } } %t143, 0
  br i1 %t144, label %m.arm030, label %m.next32
m.next32:
  %t145 = xor i1 %t144, true
  br i1 %t145, label %m.arm131, label %m.next33
m.next33:
  call void @aila_rt_trap(ptr @.str.3, i64 27, ptr @.str.1, i64 90)
  unreachable
m.arm030:
  %t146 = load { i1, double, { i8, { i64 } } }, ptr %slot___head_49
  %t147 = extractvalue { i1, double, { i8, { i64 } } } %t146, 1
  store double %t147, ptr %slot_v_10
  %t148 = insertvalue { ptr, i64 } undef, ptr @.str.14, 0
  %t149 = insertvalue { ptr, i64 } %t148, i64 25, 1
  %t150 = insertvalue { i8, { { ptr, i64 } } } undef, i8 2, 0
  store { i8, { { ptr, i64 } } } %t150, ptr %slot___lit_50
  %t151 = load double, ptr %slot_v_10
  %t152 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_50, i32 0, i32 1, i32 0
  store double %t151, ptr %t152
  %t153 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_50
  %t154 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_51, i64 0
  store { i8, { { ptr, i64 } } } %t153, ptr %t154
  %t155 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_51, 0
  %t156 = insertvalue { ptr, i64 } %t155, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t149, { ptr, i64 } %t156)
  br label %m.end29
m.arm131:
  %t157 = load { i1, double, { i8, { i64 } } }, ptr %slot___head_49
  %t158 = extractvalue { i1, double, { i8, { i64 } } } %t157, 2
  store { i8, { i64 } } %t158, ptr %slot_e_11
  %t159 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t159, ptr %slot___lit_52
  %t160 = insertvalue { ptr, i64 } undef, ptr @.str.5, 0
  %t161 = insertvalue { ptr, i64 } %t160, i64 10, 1
  %t162 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_52, i32 0, i32 1, i32 0
  store { ptr, i64 } %t161, ptr %t162
  %t163 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_52
  %t164 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_53, i64 0
  store { i8, { { ptr, i64 } } } %t163, ptr %t164
  %t165 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_53, 0
  %t166 = insertvalue { ptr, i64 } %t165, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t166)
  %t167 = load { i8, { i64 } }, ptr %slot_e_11
  br label %m.end29
m.end29:
  %t168 = call ptr @aila_rt_region_alloc(ptr %t1, i64 40, i64 1)
  %t169 = insertvalue { ptr, i64 } undef, ptr %t168, 0
  %t170 = insertvalue { ptr, i64 } %t169, i64 40, 1
  store { ptr, i64 } %t170, ptr %slot_fbuf_12
  %t171 = sub i64 0, 42
  %t172 = load { ptr, i64 }, ptr %slot_fbuf_12
  %t173 = call i64 @aila_std_strconv__format_i64(i64 %t171, { ptr, i64 } %t172)
  store i64 %t173, ptr %slot_n_13
  %t174 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t174, ptr %slot___lit_54
  %t175 = insertvalue { ptr, i64 } undef, ptr @.str.15, 0
  %t176 = insertvalue { ptr, i64 } %t175, i64 18, 1
  %t177 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_54, i32 0, i32 1, i32 0
  store { ptr, i64 } %t176, ptr %t177
  %t178 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_54
  %t179 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_55, i64 0
  store { i8, { { ptr, i64 } } } %t178, ptr %t179
  %t180 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_55, 0
  %t181 = insertvalue { ptr, i64 } %t180, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t181)
  %t182 = load { ptr, i64 }, ptr %slot_fbuf_12
  %t183 = load i64, ptr %slot_n_13
  call void @aila_main__show_bytes(ptr %t1, { ptr, i64 } %t182, i64 %t183)
  %t184 = insertvalue { ptr, i64 } undef, ptr @.str.16, 0
  %t185 = insertvalue { ptr, i64 } %t184, i64 12, 1
  %t186 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t186, ptr %slot___lit_56
  %t187 = load i64, ptr %slot_n_13
  %t189 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_56, i32 0, i32 1, i32 0
  store i64 %t187, ptr %t189
  %t190 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_56
  %t191 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_57, i64 0
  store { i8, { { ptr, i64 } } } %t190, ptr %t191
  %t192 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_57, 0
  %t193 = insertvalue { ptr, i64 } %t192, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t185, { ptr, i64 } %t193)
  %t194 = load { ptr, i64 }, ptr %slot_fbuf_12
  %t195 = call i64 @aila_std_strconv__format_u64(i64 -1, { ptr, i64 } %t194)
  store i64 %t195, ptr %slot_n_13
  %t196 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t196, ptr %slot___lit_58
  %t197 = insertvalue { ptr, i64 } undef, ptr @.str.17, 0
  %t198 = insertvalue { ptr, i64 } %t197, i64 18, 1
  %t199 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_58, i32 0, i32 1, i32 0
  store { ptr, i64 } %t198, ptr %t199
  %t200 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_58
  %t201 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_59, i64 0
  store { i8, { { ptr, i64 } } } %t200, ptr %t201
  %t202 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_59, 0
  %t203 = insertvalue { ptr, i64 } %t202, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t203)
  %t204 = load { ptr, i64 }, ptr %slot_fbuf_12
  %t205 = load i64, ptr %slot_n_13
  call void @aila_main__show_bytes(ptr %t1, { ptr, i64 } %t204, i64 %t205)
  %t206 = insertvalue { ptr, i64 } undef, ptr @.str.16, 0
  %t207 = insertvalue { ptr, i64 } %t206, i64 12, 1
  %t208 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t208, ptr %slot___lit_60
  %t209 = load i64, ptr %slot_n_13
  %t211 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_60, i32 0, i32 1, i32 0
  store i64 %t209, ptr %t211
  %t212 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_60
  %t213 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_61, i64 0
  store { i8, { { ptr, i64 } } } %t212, ptr %t213
  %t214 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_61, 0
  %t215 = insertvalue { ptr, i64 } %t214, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t207, { ptr, i64 } %t215)
  %t216 = load { ptr, i64 }, ptr %slot_fbuf_12
  %t217 = call i64 @aila_std_strconv__format_f64(double 0.5, { ptr, i64 } %t216)
  store i64 %t217, ptr %slot_n_13
  %t218 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t218, ptr %slot___lit_62
  %t219 = insertvalue { ptr, i64 } undef, ptr @.str.18, 0
  %t220 = insertvalue { ptr, i64 } %t219, i64 18, 1
  %t221 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_62, i32 0, i32 1, i32 0
  store { ptr, i64 } %t220, ptr %t221
  %t222 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_62
  %t223 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_63, i64 0
  store { i8, { { ptr, i64 } } } %t222, ptr %t223
  %t224 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_63, 0
  %t225 = insertvalue { ptr, i64 } %t224, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t225)
  %t226 = load { ptr, i64 }, ptr %slot_fbuf_12
  %t227 = load i64, ptr %slot_n_13
  call void @aila_main__show_bytes(ptr %t1, { ptr, i64 } %t226, i64 %t227)
  %t228 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t228, ptr %slot___lit_64
  %t229 = insertvalue { ptr, i64 } undef, ptr @.str.19, 0
  %t230 = insertvalue { ptr, i64 } %t229, i64 0, 1
  %t231 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_64, i32 0, i32 1, i32 0
  store { ptr, i64 } %t230, ptr %t231
  %t232 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_64
  %t233 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_65, i64 0
  store { i8, { { ptr, i64 } } } %t232, ptr %t233
  %t234 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_65, 0
  %t235 = insertvalue { ptr, i64 } %t234, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t235)
  %t236 = insertvalue { ptr, i64 } undef, ptr @.str.20, 0
  %t237 = insertvalue { ptr, i64 } %t236, i64 34, 1
  %t238 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t238, ptr %slot___lit_66
  %t239 = insertvalue { ptr, i64 } undef, ptr @.str.21, 0
  %t240 = insertvalue { ptr, i64 } %t239, i64 11, 1
  %t241 = insertvalue { ptr, i64 } undef, ptr @.str.22, 0
  %t242 = insertvalue { ptr, i64 } %t241, i64 4, 1
  %t243 = call i64 @aila_std_strings__index({ ptr, i64 } %t240, { ptr, i64 } %t242)
  %t245 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_66, i32 0, i32 1, i32 0
  store i64 %t243, ptr %t245
  %t246 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_66
  %t247 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_67, i64 0
  store { i8, { { ptr, i64 } } } %t246, ptr %t247
  %t248 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_67, 0
  %t249 = insertvalue { ptr, i64 } %t248, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t237, { ptr, i64 } %t249)
  %t250 = insertvalue { ptr, i64 } undef, ptr @.str.23, 0
  %t251 = insertvalue { ptr, i64 } %t250, i64 36, 1
  %t252 = insertvalue { i8, { { ptr, i64 } } } undef, i8 3, 0
  store { i8, { { ptr, i64 } } } %t252, ptr %slot___lit_68
  %t253 = insertvalue { ptr, i64 } undef, ptr @.str.21, 0
  %t254 = insertvalue { ptr, i64 } %t253, i64 11, 1
  %t255 = insertvalue { ptr, i64 } undef, ptr @.str.24, 0
  %t256 = insertvalue { ptr, i64 } %t255, i64 3, 1
  %t257 = call i1 @aila_std_strings__contains({ ptr, i64 } %t254, { ptr, i64 } %t256)
  %t258 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_68, i32 0, i32 1, i32 0
  store i1 %t257, ptr %t258
  %t259 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_68
  %t260 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_69, i64 0
  store { i8, { { ptr, i64 } } } %t259, ptr %t260
  %t261 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_69, 0
  %t262 = insertvalue { ptr, i64 } %t261, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t251, { ptr, i64 } %t262)
  %t263 = insertvalue { ptr, i64 } undef, ptr @.str.25, 0
  %t264 = insertvalue { ptr, i64 } %t263, i64 63, 1
  %t265 = insertvalue { i8, { { ptr, i64 } } } undef, i8 3, 0
  store { i8, { { ptr, i64 } } } %t265, ptr %slot___lit_70
  %t266 = insertvalue { ptr, i64 } undef, ptr @.str.26, 0
  %t267 = insertvalue { ptr, i64 } %t266, i64 5, 1
  %t268 = insertvalue { ptr, i64 } undef, ptr @.str.27, 0
  %t269 = insertvalue { ptr, i64 } %t268, i64 2, 1
  %t270 = call i1 @aila_std_strings__has_prefix({ ptr, i64 } %t267, { ptr, i64 } %t269)
  %t271 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_70, i32 0, i32 1, i32 0
  store i1 %t270, ptr %t271
  %t272 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_70
  %t273 = insertvalue { i8, { { ptr, i64 } } } undef, i8 3, 0
  store { i8, { { ptr, i64 } } } %t273, ptr %slot___lit_71
  %t274 = insertvalue { ptr, i64 } undef, ptr @.str.26, 0
  %t275 = insertvalue { ptr, i64 } %t274, i64 5, 1
  %t276 = insertvalue { ptr, i64 } undef, ptr @.str.28, 0
  %t277 = insertvalue { ptr, i64 } %t276, i64 2, 1
  %t278 = call i1 @aila_std_strings__has_suffix({ ptr, i64 } %t275, { ptr, i64 } %t277)
  %t279 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_71, i32 0, i32 1, i32 0
  store i1 %t278, ptr %t279
  %t280 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_71
  %t281 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_72, i64 0
  store { i8, { { ptr, i64 } } } %t272, ptr %t281
  %t282 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_72, i64 1
  store { i8, { { ptr, i64 } } } %t280, ptr %t282
  %t283 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_72, 0
  %t284 = insertvalue { ptr, i64 } %t283, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t264, { ptr, i64 } %t284)
  %t285 = insertvalue { ptr, i64 } undef, ptr @.str.29, 0
  %t286 = insertvalue { ptr, i64 } %t285, i64 51, 1
  %t287 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t287, ptr %slot___lit_73
  %t288 = insertvalue { ptr, i64 } undef, ptr @.str.30, 0
  %t289 = insertvalue { ptr, i64 } %t288, i64 5, 1
  %t290 = call i64 @aila_std_strings__len_runes({ ptr, i64 } %t289)
  %t292 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_73, i32 0, i32 1, i32 0
  store i64 %t290, ptr %t292
  %t293 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_73
  %t294 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t294, ptr %slot___lit_74
  %t295 = insertvalue { ptr, i64 } undef, ptr @.str.30, 0
  %t296 = insertvalue { ptr, i64 } %t295, i64 5, 1
  %t297 = call i32 @aila_std_utf8__cp_at({ ptr, i64 } %t296, i64 1)
  %t298 = zext i32 %t297 to i64
  %t299 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_74, i32 0, i32 1, i32 0
  store i64 %t298, ptr %t299
  %t300 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_74
  %t301 = insertvalue { i8, { { ptr, i64 } } } undef, i8 3, 0
  store { i8, { { ptr, i64 } } } %t301, ptr %slot___lit_75
  %t302 = insertvalue { ptr, i64 } undef, ptr @.str.30, 0
  %t303 = insertvalue { ptr, i64 } %t302, i64 5, 1
  %t304 = call i1 @aila_std_utf8__valid({ ptr, i64 } %t303)
  %t305 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_75, i32 0, i32 1, i32 0
  store i1 %t304, ptr %t305
  %t306 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_75
  %t307 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_76, i64 0
  store { i8, { { ptr, i64 } } } %t293, ptr %t307
  %t308 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_76, i64 1
  store { i8, { { ptr, i64 } } } %t300, ptr %t308
  %t309 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_76, i64 2
  store { i8, { { ptr, i64 } } } %t306, ptr %t309
  %t310 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_76, 0
  %t311 = insertvalue { ptr, i64 } %t310, i64 3, 1
  call void @aila_std_io__printf({ ptr, i64 } %t286, { ptr, i64 } %t311)
  %t312 = insertvalue { ptr, i64 } undef, ptr @.str.31, 0
  %t313 = insertvalue { ptr, i64 } %t312, i64 28, 1
  %t314 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t314, ptr %slot___lit_77
  %t315 = insertvalue { ptr, i64 } undef, ptr @.str.32, 0
  %t316 = insertvalue { ptr, i64 } %t315, i64 2, 1
  %t317 = call i32 @aila_std_utf8__cp_at({ ptr, i64 } %t316, i64 5)
  %t318 = zext i32 %t317 to i64
  %t319 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_77, i32 0, i32 1, i32 0
  store i64 %t318, ptr %t319
  %t320 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_77
  %t321 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_78, i64 0
  store { i8, { { ptr, i64 } } } %t320, ptr %t321
  %t322 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_78, 0
  %t323 = insertvalue { ptr, i64 } %t322, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t313, { ptr, i64 } %t323)
  store i64 0, ptr %slot_s_14
  store i64 0, ptr %slot_e_15
  %t324 = insertvalue { ptr, i64 } undef, ptr @.str.33, 0
  %t325 = insertvalue { ptr, i64 } %t324, i64 6, 1
  %t326 = call { i64, i64 } @aila_std_strings__bounds({ ptr, i64 } %t325)
  store { i64, i64 } %t326, ptr %slot___destr_79
  %t327 = getelementptr inbounds { i64, i64 }, ptr %slot___destr_79, i32 0, i32 0
  %t328 = load i64, ptr %t327
  store i64 %t328, ptr %slot_s_14
  %t329 = getelementptr inbounds { i64, i64 }, ptr %slot___destr_79, i32 0, i32 1
  %t330 = load i64, ptr %t329
  store i64 %t330, ptr %slot_e_15
  %t331 = insertvalue { ptr, i64 } undef, ptr @.str.34, 0
  %t332 = insertvalue { ptr, i64 } %t331, i64 28, 1
  %t333 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t333, ptr %slot___lit_80
  %t334 = load i64, ptr %slot_s_14
  %t336 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_80, i32 0, i32 1, i32 0
  store i64 %t334, ptr %t336
  %t337 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_80
  %t338 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t338, ptr %slot___lit_81
  %t339 = load i64, ptr %slot_e_15
  %t341 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_81, i32 0, i32 1, i32 0
  store i64 %t339, ptr %t341
  %t342 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_81
  %t343 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_82, i64 0
  store { i8, { { ptr, i64 } } } %t337, ptr %t343
  %t344 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_82, i64 1
  store { i8, { { ptr, i64 } } } %t342, ptr %t344
  %t345 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_82, 0
  %t346 = insertvalue { ptr, i64 } %t345, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t332, { ptr, i64 } %t346)
  %t347 = call ptr @aila_rt_region_alloc(ptr %t1, i64 3, i64 1)
  %t348 = insertvalue { ptr, i64 } undef, ptr %t347, 0
  %t349 = insertvalue { ptr, i64 } %t348, i64 3, 1
  store { ptr, i64 } %t349, ptr %slot_small_16
  %t350 = insertvalue { ptr, i64 } undef, ptr @.str.35, 0
  %t351 = insertvalue { ptr, i64 } %t350, i64 3, 1
  %t352 = insertvalue { ptr, i64 } undef, ptr @.str.36, 0
  %t353 = insertvalue { ptr, i64 } %t352, i64 2, 1
  %t354 = load { ptr, i64 }, ptr %slot_small_16
  %t355 = call i64 @aila_std_strings__concat({ ptr, i64 } %t351, { ptr, i64 } %t353, { ptr, i64 } %t354)
  store i64 %t355, ptr %slot_n_13
  %t356 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t356, ptr %slot___lit_83
  %t357 = insertvalue { ptr, i64 } undef, ptr @.str.37, 0
  %t358 = insertvalue { ptr, i64 } %t357, i64 10, 1
  %t359 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_83, i32 0, i32 1, i32 0
  store { ptr, i64 } %t358, ptr %t359
  %t360 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_83
  %t361 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_84, i64 0
  store { i8, { { ptr, i64 } } } %t360, ptr %t361
  %t362 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_84, 0
  %t363 = insertvalue { ptr, i64 } %t362, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t363)
  %t364 = load { ptr, i64 }, ptr %slot_small_16
  %t365 = load i64, ptr %slot_n_13
  call void @aila_main__show_bytes(ptr %t1, { ptr, i64 } %t364, i64 %t365)
  %t366 = insertvalue { ptr, i64 } undef, ptr @.str.38, 0
  %t367 = insertvalue { ptr, i64 } %t366, i64 20, 1
  %t368 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t368, ptr %slot___lit_85
  %t369 = load i64, ptr %slot_n_13
  %t371 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_85, i32 0, i32 1, i32 0
  store i64 %t369, ptr %t371
  %t372 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_85
  %t373 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_86, i64 0
  store { i8, { { ptr, i64 } } } %t372, ptr %t373
  %t374 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_86, 0
  %t375 = insertvalue { ptr, i64 } %t374, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t367, { ptr, i64 } %t375)
  %t376 = call ptr @aila_rt_region_alloc(ptr %t1, i64 40, i64 1)
  %t377 = insertvalue { ptr, i64 } undef, ptr %t376, 0
  %t378 = insertvalue { ptr, i64 } %t377, i64 40, 1
  store { ptr, i64 } %t378, ptr %slot_buf_17
  %t379 = insertvalue { ptr, i64 } undef, ptr @.str.35, 0
  %t380 = insertvalue { ptr, i64 } %t379, i64 3, 1
  %t381 = insertvalue { ptr, i64 } undef, ptr @.str.36, 0
  %t382 = insertvalue { ptr, i64 } %t381, i64 2, 1
  %t383 = load { ptr, i64 }, ptr %slot_buf_17
  %t384 = call i64 @aila_std_strings__concat({ ptr, i64 } %t380, { ptr, i64 } %t382, { ptr, i64 } %t383)
  store i64 %t384, ptr %slot_n_13
  %t385 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t385, ptr %slot___lit_87
  %t386 = insertvalue { ptr, i64 } undef, ptr @.str.37, 0
  %t387 = insertvalue { ptr, i64 } %t386, i64 10, 1
  %t388 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_87, i32 0, i32 1, i32 0
  store { ptr, i64 } %t387, ptr %t388
  %t389 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_87
  %t390 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_88, i64 0
  store { i8, { { ptr, i64 } } } %t389, ptr %t390
  %t391 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_88, 0
  %t392 = insertvalue { ptr, i64 } %t391, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t392)
  %t393 = load { ptr, i64 }, ptr %slot_buf_17
  %t394 = load i64, ptr %slot_n_13
  call void @aila_main__show_bytes(ptr %t1, { ptr, i64 } %t393, i64 %t394)
  %t395 = insertvalue { ptr, i64 } undef, ptr @.str.39, 0
  %t396 = insertvalue { ptr, i64 } %t395, i64 6, 1
  %t397 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t397, ptr %slot___lit_89
  %t398 = load i64, ptr %slot_n_13
  %t400 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_89, i32 0, i32 1, i32 0
  store i64 %t398, ptr %t400
  %t401 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_89
  %t402 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_90, i64 0
  store { i8, { { ptr, i64 } } } %t401, ptr %t402
  %t403 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_90, 0
  %t404 = insertvalue { ptr, i64 } %t403, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t396, { ptr, i64 } %t404)
  %t405 = call ptr @aila_rt_region_alloc(ptr %t1, i64 64, i64 8)
  %t406 = insertvalue { ptr, i64 } undef, ptr %t405, 0
  %t407 = insertvalue { ptr, i64 } %t406, i64 8, 1
  store { ptr, i64 } %t407, ptr %slot_marks_18
  %t408 = insertvalue { ptr, i64 } undef, ptr @.str.40, 0
  %t409 = insertvalue { ptr, i64 } %t408, i64 5, 1
  %t410 = load { ptr, i64 }, ptr %slot_marks_18
  %t411 = call i64 @aila_std_strings__split({ ptr, i64 } %t409, i8 44, { ptr, i64 } %t410)
  store i64 %t411, ptr %slot_segs_19
  %t412 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t412, ptr %slot___lit_91
  %t413 = insertvalue { ptr, i64 } undef, ptr @.str.41, 0
  %t414 = insertvalue { ptr, i64 } %t413, i64 17, 1
  %t415 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_91, i32 0, i32 1, i32 0
  store { ptr, i64 } %t414, ptr %t415
  %t416 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_91
  %t417 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_92, i64 0
  store { i8, { { ptr, i64 } } } %t416, ptr %t417
  %t418 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_92, 0
  %t419 = insertvalue { ptr, i64 } %t418, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t419)
  %t420 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t420, ptr %slot___lit_93
  %t421 = load i64, ptr %slot_segs_19
  %t423 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_93, i32 0, i32 1, i32 0
  store i64 %t421, ptr %t423
  %t424 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_93
  %t425 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_94, i64 0
  store { i8, { { ptr, i64 } } } %t424, ptr %t425
  %t426 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_94, 0
  %t427 = insertvalue { ptr, i64 } %t426, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t427)
  %t428 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t428, ptr %slot___lit_95
  %t429 = insertvalue { ptr, i64 } undef, ptr @.str.42, 0
  %t430 = insertvalue { ptr, i64 } %t429, i64 12, 1
  %t431 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_95, i32 0, i32 1, i32 0
  store { ptr, i64 } %t430, ptr %t431
  %t432 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_95
  %t433 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_96, i64 0
  store { i8, { { ptr, i64 } } } %t432, ptr %t433
  %t434 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_96, 0
  %t435 = insertvalue { ptr, i64 } %t434, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t435)
  store i64 0, ptr %slot_k_20
  br label %for.cond35
for.cond35:
  %t436 = load i64, ptr %slot_k_20
  %t437 = load i64, ptr %slot_segs_19
  %t438 = icmp ult i64 %t436, %t437
  br i1 %t438, label %for.body37, label %for.end34
for.body37:
  %t439 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t439, ptr %slot___lit_97
  %t440 = load i64, ptr %slot_k_20
  %t441 = load { ptr, i64 }, ptr %slot_marks_18
  %t442 = extractvalue { ptr, i64 } %t441, 0
  %t443 = extractvalue { ptr, i64 } %t441, 1
  %t444 = icmp uge i64 %t440, %t443
  br i1 %t444, label %chk.bad39, label %chk.ok38
chk.bad39:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 154)
  unreachable
chk.ok38:
  %t445 = getelementptr inbounds i64, ptr %t442, i64 %t440
  %t446 = load i64, ptr %t445
  %t448 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_97, i32 0, i32 1, i32 0
  store i64 %t446, ptr %t448
  %t449 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_97
  %t450 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_98, i64 0
  store { i8, { { ptr, i64 } } } %t449, ptr %t450
  %t451 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_98, 0
  %t452 = insertvalue { ptr, i64 } %t451, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t452)
  %t453 = load i64, ptr %slot_k_20
  %t454 = add i64 %t453, 1
  %t455 = load i64, ptr %slot_segs_19
  %t456 = icmp ult i64 %t454, %t455
  br i1 %t456, label %if.then41, label %if.next42
if.then41:
  %t457 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t457, ptr %slot___lit_99
  %t458 = insertvalue { ptr, i64 } undef, ptr @.str.43, 0
  %t459 = insertvalue { ptr, i64 } %t458, i64 1, 1
  %t460 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_99, i32 0, i32 1, i32 0
  store { ptr, i64 } %t459, ptr %t460
  %t461 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_99
  %t462 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_100, i64 0
  store { i8, { { ptr, i64 } } } %t461, ptr %t462
  %t463 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_100, 0
  %t464 = insertvalue { ptr, i64 } %t463, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t464)
  br label %if.end40
if.next42:
  br label %if.end40
if.end40:
  %t465 = load i64, ptr %slot_k_20
  %t466 = add i64 %t465, 1
  store i64 %t466, ptr %slot_k_20
  br label %for.step36
for.step36:
  br label %for.cond35
for.end34:
  %t467 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t467, ptr %slot___lit_101
  %t468 = insertvalue { ptr, i64 } undef, ptr @.str.19, 0
  %t469 = insertvalue { ptr, i64 } %t468, i64 0, 1
  %t470 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_101, i32 0, i32 1, i32 0
  store { ptr, i64 } %t469, ptr %t470
  %t471 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_101
  %t472 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_102, i64 0
  store { i8, { { ptr, i64 } } } %t471, ptr %t472
  %t473 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_102, 0
  %t474 = insertvalue { ptr, i64 } %t473, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t474)
  %t475 = load { ptr, i64 }, ptr %slot_buf_17
  %t476 = insertvalue { ptr, i64 } undef, ptr @.str.44, 0
  %t477 = insertvalue { ptr, i64 } %t476, i64 1, 1
  %t478 = insertvalue { ptr, i64 } undef, ptr @.str.45, 0
  %t479 = insertvalue { ptr, i64 } %t478, i64 1, 1
  %t480 = insertvalue { ptr, i64 } undef, ptr @.str.46, 0
  %t481 = insertvalue { ptr, i64 } %t480, i64 1, 1
  %t482 = getelementptr inbounds { ptr, i64 }, ptr %slot_vargs_103, i64 0
  store { ptr, i64 } %t477, ptr %t482
  %t483 = getelementptr inbounds { ptr, i64 }, ptr %slot_vargs_103, i64 1
  store { ptr, i64 } %t479, ptr %t483
  %t484 = getelementptr inbounds { ptr, i64 }, ptr %slot_vargs_103, i64 2
  store { ptr, i64 } %t481, ptr %t484
  %t485 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_103, 0
  %t486 = insertvalue { ptr, i64 } %t485, i64 3, 1
  %t487 = call i64 @aila_std_strings__join({ ptr, i64 } %t475, i8 45, { ptr, i64 } %t486)
  store i64 %t487, ptr %slot_n_13
  %t488 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t488, ptr %slot___lit_104
  %t489 = insertvalue { ptr, i64 } undef, ptr @.str.47, 0
  %t490 = insertvalue { ptr, i64 } %t489, i64 8, 1
  %t491 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_104, i32 0, i32 1, i32 0
  store { ptr, i64 } %t490, ptr %t491
  %t492 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_104
  %t493 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_105, i64 0
  store { i8, { { ptr, i64 } } } %t492, ptr %t493
  %t494 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_105, 0
  %t495 = insertvalue { ptr, i64 } %t494, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t495)
  %t496 = load { ptr, i64 }, ptr %slot_buf_17
  %t497 = load i64, ptr %slot_n_13
  call void @aila_main__show_bytes(ptr %t1, { ptr, i64 } %t496, i64 %t497)
  %t498 = insertvalue { ptr, i64 } undef, ptr @.str.39, 0
  %t499 = insertvalue { ptr, i64 } %t498, i64 6, 1
  %t500 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t500, ptr %slot___lit_106
  %t501 = load i64, ptr %slot_n_13
  %t503 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_106, i32 0, i32 1, i32 0
  store i64 %t501, ptr %t503
  %t504 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_106
  %t505 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_107, i64 0
  store { i8, { { ptr, i64 } } } %t504, ptr %t505
  %t506 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_107, 0
  %t507 = insertvalue { ptr, i64 } %t506, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t499, { ptr, i64 } %t507)
  %t508 = insertvalue { ptr, i64 } undef, ptr @.str.48, 0
  %t509 = insertvalue { ptr, i64 } %t508, i64 5, 1
  %t510 = insertvalue { ptr, i64 } undef, ptr @.str.49, 0
  %t511 = insertvalue { ptr, i64 } %t510, i64 1, 1
  %t512 = insertvalue { ptr, i64 } undef, ptr @.str.50, 0
  %t513 = insertvalue { ptr, i64 } %t512, i64 1, 1
  %t514 = load { ptr, i64 }, ptr %slot_buf_17
  %t515 = call i64 @aila_std_strings__replace({ ptr, i64 } %t509, { ptr, i64 } %t511, { ptr, i64 } %t513, { ptr, i64 } %t514)
  store i64 %t515, ptr %slot_n_13
  %t516 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t516, ptr %slot___lit_108
  %t517 = insertvalue { ptr, i64 } undef, ptr @.str.51, 0
  %t518 = insertvalue { ptr, i64 } %t517, i64 11, 1
  %t519 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_108, i32 0, i32 1, i32 0
  store { ptr, i64 } %t518, ptr %t519
  %t520 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_108
  %t521 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_109, i64 0
  store { i8, { { ptr, i64 } } } %t520, ptr %t521
  %t522 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_109, 0
  %t523 = insertvalue { ptr, i64 } %t522, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t523)
  %t524 = load { ptr, i64 }, ptr %slot_buf_17
  %t525 = load i64, ptr %slot_n_13
  call void @aila_main__show_bytes(ptr %t1, { ptr, i64 } %t524, i64 %t525)
  %t526 = insertvalue { ptr, i64 } undef, ptr @.str.39, 0
  %t527 = insertvalue { ptr, i64 } %t526, i64 6, 1
  %t528 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t528, ptr %slot___lit_110
  %t529 = load i64, ptr %slot_n_13
  %t531 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_110, i32 0, i32 1, i32 0
  store i64 %t529, ptr %t531
  %t532 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_110
  %t533 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_111, i64 0
  store { i8, { { ptr, i64 } } } %t532, ptr %t533
  %t534 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_111, 0
  %t535 = insertvalue { ptr, i64 } %t534, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t527, { ptr, i64 } %t535)
  %t536 = insertvalue { ptr, i64 } undef, ptr @.str.52, 0
  %t537 = insertvalue { ptr, i64 } %t536, i64 4, 1
  %t538 = load { ptr, i64 }, ptr %slot_buf_17
  %t539 = call i64 @aila_std_strings__upper_ascii({ ptr, i64 } %t537, { ptr, i64 } %t538)
  store i64 %t539, ptr %slot_n_13
  %t540 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t540, ptr %slot___lit_112
  %t541 = insertvalue { ptr, i64 } undef, ptr @.str.53, 0
  %t542 = insertvalue { ptr, i64 } %t541, i64 9, 1
  %t543 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_112, i32 0, i32 1, i32 0
  store { ptr, i64 } %t542, ptr %t543
  %t544 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_112
  %t545 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_113, i64 0
  store { i8, { { ptr, i64 } } } %t544, ptr %t545
  %t546 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_113, 0
  %t547 = insertvalue { ptr, i64 } %t546, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t547)
  %t548 = load { ptr, i64 }, ptr %slot_buf_17
  %t549 = load i64, ptr %slot_n_13
  call void @aila_main__show_bytes(ptr %t1, { ptr, i64 } %t548, i64 %t549)
  %t550 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t550, ptr %slot___lit_114
  %t551 = insertvalue { ptr, i64 } undef, ptr @.str.54, 0
  %t552 = insertvalue { ptr, i64 } %t551, i64 11, 1
  %t553 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_114, i32 0, i32 1, i32 0
  store { ptr, i64 } %t552, ptr %t553
  %t554 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_114
  %t555 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_115, i64 0
  store { i8, { { ptr, i64 } } } %t554, ptr %t555
  %t556 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_115, 0
  %t557 = insertvalue { ptr, i64 } %t556, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t557)
  %t558 = insertvalue { ptr, i64 } undef, ptr @.str.22, 0
  %t559 = insertvalue { ptr, i64 } %t558, i64 4, 1
  %t560 = load { ptr, i64 }, ptr %slot_buf_17
  %t561 = call i64 @aila_std_strings__lower_ascii({ ptr, i64 } %t559, { ptr, i64 } %t560)
  store i64 %t561, ptr %slot_n_13
  %t562 = load { ptr, i64 }, ptr %slot_buf_17
  %t563 = load i64, ptr %slot_n_13
  call void @aila_main__show_bytes(ptr %t1, { ptr, i64 } %t562, i64 %t563)
  %t564 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t564, ptr %slot___lit_116
  %t565 = insertvalue { ptr, i64 } undef, ptr @.str.19, 0
  %t566 = insertvalue { ptr, i64 } %t565, i64 0, 1
  %t567 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_116, i32 0, i32 1, i32 0
  store { ptr, i64 } %t566, ptr %t567
  %t568 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_116
  %t569 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_117, i64 0
  store { i8, { { ptr, i64 } } } %t568, ptr %t569
  %t570 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_117, 0
  %t571 = insertvalue { ptr, i64 } %t570, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t571)
  %t572 = call ptr @aila_rt_region_alloc(ptr %t1, i64 16, i64 1)
  %t573 = insertvalue { ptr, i64 } undef, ptr %t572, 0
  %t574 = insertvalue { ptr, i64 } %t573, i64 16, 1
  store { ptr, i64 } %t574, ptr %slot_bbuf_21
  %t575 = load { ptr, i64 }, ptr %slot_bbuf_21
  call void @aila_std_binary__le_write_u64({ ptr, i64 } %t575, i64 0, i64 1234605616436508552)
  %t576 = insertvalue { ptr, i64 } undef, ptr @.str.55, 0
  %t577 = insertvalue { ptr, i64 } %t576, i64 25, 1
  %t578 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t578, ptr %slot___lit_118
  %t579 = load { ptr, i64 }, ptr %slot_bbuf_21
  %t580 = call i64 @aila_std_binary__le_read_u64({ ptr, i64 } %t579, i64 0)
  %t581 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_118, i32 0, i32 1, i32 0
  store i64 %t580, ptr %t581
  %t582 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_118
  %t583 = insertvalue { i8, { { ptr, i64 } } } undef, i8 1, 0
  store { i8, { { ptr, i64 } } } %t583, ptr %slot___lit_119
  %t584 = load { ptr, i64 }, ptr %slot_bbuf_21
  %t585 = call i64 @aila_std_binary__be_read_u64({ ptr, i64 } %t584, i64 0)
  %t586 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_119, i32 0, i32 1, i32 0
  store i64 %t585, ptr %t586
  %t587 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_119
  %t588 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_120, i64 0
  store { i8, { { ptr, i64 } } } %t582, ptr %t588
  %t589 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_120, i64 1
  store { i8, { { ptr, i64 } } } %t587, ptr %t589
  %t590 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_120, 0
  %t591 = insertvalue { ptr, i64 } %t590, i64 2, 1
  call void @aila_std_io__printf({ ptr, i64 } %t577, { ptr, i64 } %t591)
  %t592 = load { ptr, i64 }, ptr %slot_bbuf_21
  %t593 = sub i64 0, 2
  %t594 = trunc i64 %t593 to i16
  call void @aila_std_binary__le_write_i16({ ptr, i64 } %t592, i64 8, i16 %t594)
  %t595 = insertvalue { ptr, i64 } undef, ptr @.str.56, 0
  %t596 = insertvalue { ptr, i64 } %t595, i64 22, 1
  %t597 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t597, ptr %slot___lit_121
  %t598 = load { ptr, i64 }, ptr %slot_bbuf_21
  %t599 = call i16 @aila_std_binary__le_read_i16({ ptr, i64 } %t598, i64 8)
  %t600 = sext i16 %t599 to i64
  %t601 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_121, i32 0, i32 1, i32 0
  store i64 %t600, ptr %t601
  %t602 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_121
  %t603 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_122, i64 0
  store { i8, { { ptr, i64 } } } %t602, ptr %t603
  %t604 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_122, 0
  %t605 = insertvalue { ptr, i64 } %t604, i64 1, 1
  call void @aila_std_io__printf({ ptr, i64 } %t596, { ptr, i64 } %t605)
  %t606 = call ptr @aila_rt_region_alloc(ptr %t1, i64 32, i64 8)
  %t607 = insertvalue { ptr, i64 } undef, ptr %t606, 0
  %t608 = insertvalue { ptr, i64 } %t607, i64 4, 1
  store { ptr, i64 } %t608, ptr %slot_a_22
  %t609 = load { ptr, i64 }, ptr %slot_a_22
  %t610 = extractvalue { ptr, i64 } %t609, 0
  %t611 = extractvalue { ptr, i64 } %t609, 1
  %t612 = icmp uge i64 0, %t611
  br i1 %t612, label %chk.bad44, label %chk.ok43
chk.bad44:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 192)
  unreachable
chk.ok43:
  %t613 = getelementptr inbounds i64, ptr %t610, i64 0
  store i64 9, ptr %t613
  %t614 = load { ptr, i64 }, ptr %slot_a_22
  %t615 = extractvalue { ptr, i64 } %t614, 0
  %t616 = extractvalue { ptr, i64 } %t614, 1
  %t617 = icmp uge i64 1, %t616
  br i1 %t617, label %chk.bad46, label %chk.ok45
chk.bad46:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 193)
  unreachable
chk.ok45:
  %t618 = getelementptr inbounds i64, ptr %t615, i64 1
  store i64 3, ptr %t618
  %t619 = load { ptr, i64 }, ptr %slot_a_22
  %t620 = extractvalue { ptr, i64 } %t619, 0
  %t621 = extractvalue { ptr, i64 } %t619, 1
  %t622 = icmp uge i64 2, %t621
  br i1 %t622, label %chk.bad48, label %chk.ok47
chk.bad48:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 194)
  unreachable
chk.ok47:
  %t623 = getelementptr inbounds i64, ptr %t620, i64 2
  store i64 12, ptr %t623
  %t624 = load { ptr, i64 }, ptr %slot_a_22
  %t625 = extractvalue { ptr, i64 } %t624, 0
  %t626 = extractvalue { ptr, i64 } %t624, 1
  %t627 = icmp uge i64 3, %t626
  br i1 %t627, label %chk.bad50, label %chk.ok49
chk.bad50:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 195)
  unreachable
chk.ok49:
  %t628 = getelementptr inbounds i64, ptr %t625, i64 3
  store i64 7, ptr %t628
  %t629 = load { ptr, i64 }, ptr %slot_a_22
  call void @aila_std_sort__sort_i64({ ptr, i64 } %t629)
  %t630 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t630, ptr %slot___lit_123
  %t631 = insertvalue { ptr, i64 } undef, ptr @.str.57, 0
  %t632 = insertvalue { ptr, i64 } %t631, i64 10, 1
  %t633 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_123, i32 0, i32 1, i32 0
  store { ptr, i64 } %t632, ptr %t633
  %t634 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_123
  %t635 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_124, i64 0
  store { i8, { { ptr, i64 } } } %t634, ptr %t635
  %t636 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_124, 0
  %t637 = insertvalue { ptr, i64 } %t636, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t637)
  store i64 0, ptr %slot_k_20
  br label %for.cond52
for.cond52:
  %t638 = load i64, ptr %slot_k_20
  %t639 = icmp ult i64 %t638, 4
  br i1 %t639, label %for.body54, label %for.end51
for.body54:
  %t640 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t640, ptr %slot___lit_125
  %t641 = load i64, ptr %slot_k_20
  %t642 = load { ptr, i64 }, ptr %slot_a_22
  %t643 = extractvalue { ptr, i64 } %t642, 0
  %t644 = extractvalue { ptr, i64 } %t642, 1
  %t645 = icmp uge i64 %t641, %t644
  br i1 %t645, label %chk.bad56, label %chk.ok55
chk.bad56:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 200)
  unreachable
chk.ok55:
  %t646 = getelementptr inbounds i64, ptr %t643, i64 %t641
  %t647 = load i64, ptr %t646
  %t648 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_125, i32 0, i32 1, i32 0
  store i64 %t647, ptr %t648
  %t649 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_125
  %t650 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_126, i64 0
  store { i8, { { ptr, i64 } } } %t649, ptr %t650
  %t651 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_126, 0
  %t652 = insertvalue { ptr, i64 } %t651, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t652)
  %t653 = load i64, ptr %slot_k_20
  %t654 = add i64 %t653, 1
  %t655 = icmp ult i64 %t654, 4
  br i1 %t655, label %if.then58, label %if.next59
if.then58:
  %t656 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t656, ptr %slot___lit_127
  %t657 = insertvalue { ptr, i64 } undef, ptr @.str.43, 0
  %t658 = insertvalue { ptr, i64 } %t657, i64 1, 1
  %t659 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_127, i32 0, i32 1, i32 0
  store { ptr, i64 } %t658, ptr %t659
  %t660 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_127
  %t661 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_128, i64 0
  store { i8, { { ptr, i64 } } } %t660, ptr %t661
  %t662 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_128, 0
  %t663 = insertvalue { ptr, i64 } %t662, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t663)
  br label %if.end57
if.next59:
  br label %if.end57
if.end57:
  %t664 = load i64, ptr %slot_k_20
  %t665 = add i64 %t664, 1
  store i64 %t665, ptr %slot_k_20
  br label %for.step53
for.step53:
  br label %for.cond52
for.end51:
  %t666 = load { ptr, i64 }, ptr %slot_a_22
  call void @aila_std_sort__sort_i64_by({ ptr, i64 } %t666, ptr @aila_main__desc_i64)
  %t667 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t667, ptr %slot___lit_129
  %t668 = insertvalue { ptr, i64 } undef, ptr @.str.58, 0
  %t669 = insertvalue { ptr, i64 } %t668, i64 9, 1
  %t670 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_129, i32 0, i32 1, i32 0
  store { ptr, i64 } %t669, ptr %t670
  %t671 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_129
  %t672 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_130, i64 0
  store { i8, { { ptr, i64 } } } %t671, ptr %t672
  %t673 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_130, 0
  %t674 = insertvalue { ptr, i64 } %t673, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t674)
  store i64 0, ptr %slot_k_20
  br label %for.cond61
for.cond61:
  %t675 = load i64, ptr %slot_k_20
  %t676 = icmp ult i64 %t675, 4
  br i1 %t676, label %for.body63, label %for.end60
for.body63:
  %t677 = insertvalue { i8, { { ptr, i64 } } } undef, i8 0, 0
  store { i8, { { ptr, i64 } } } %t677, ptr %slot___lit_131
  %t678 = load i64, ptr %slot_k_20
  %t679 = load { ptr, i64 }, ptr %slot_a_22
  %t680 = extractvalue { ptr, i64 } %t679, 0
  %t681 = extractvalue { ptr, i64 } %t679, 1
  %t682 = icmp uge i64 %t678, %t681
  br i1 %t682, label %chk.bad65, label %chk.ok64
chk.bad65:
  call void @aila_rt_trap(ptr @.str.0, i64 19, ptr @.str.1, i64 210)
  unreachable
chk.ok64:
  %t683 = getelementptr inbounds i64, ptr %t680, i64 %t678
  %t684 = load i64, ptr %t683
  %t685 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_131, i32 0, i32 1, i32 0
  store i64 %t684, ptr %t685
  %t686 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_131
  %t687 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_132, i64 0
  store { i8, { { ptr, i64 } } } %t686, ptr %t687
  %t688 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_132, 0
  %t689 = insertvalue { ptr, i64 } %t688, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t689)
  %t690 = load i64, ptr %slot_k_20
  %t691 = add i64 %t690, 1
  %t692 = icmp ult i64 %t691, 4
  br i1 %t692, label %if.then67, label %if.next68
if.then67:
  %t693 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t693, ptr %slot___lit_133
  %t694 = insertvalue { ptr, i64 } undef, ptr @.str.43, 0
  %t695 = insertvalue { ptr, i64 } %t694, i64 1, 1
  %t696 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_133, i32 0, i32 1, i32 0
  store { ptr, i64 } %t695, ptr %t696
  %t697 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_133
  %t698 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_134, i64 0
  store { i8, { { ptr, i64 } } } %t697, ptr %t698
  %t699 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_134, 0
  %t700 = insertvalue { ptr, i64 } %t699, i64 1, 1
  call void @aila_std_io__print({ ptr, i64 } %t700)
  br label %if.end66
if.next68:
  br label %if.end66
if.end66:
  %t701 = load i64, ptr %slot_k_20
  %t702 = add i64 %t701, 1
  store i64 %t702, ptr %slot_k_20
  br label %for.step62
for.step62:
  br label %for.cond61
for.end60:
  %t703 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t703, ptr %slot___lit_135
  %t704 = insertvalue { ptr, i64 } undef, ptr @.str.19, 0
  %t705 = insertvalue { ptr, i64 } %t704, i64 0, 1
  %t706 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_135, i32 0, i32 1, i32 0
  store { ptr, i64 } %t705, ptr %t706
  %t707 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_135
  %t708 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_136, i64 0
  store { i8, { { ptr, i64 } } } %t707, ptr %t708
  %t709 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_136, 0
  %t710 = insertvalue { ptr, i64 } %t709, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t710)
  call void @aila_rt_region_leave(ptr %t1)
  ret void
}

define i32 @main(i32 %argc, ptr %argv) {
entry:
  call void @aila_main__main()
  ret i32 0
}
