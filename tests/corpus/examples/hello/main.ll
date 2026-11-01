; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [13 x i8] c"hello, AILA!\00", align 1

declare void @aila_std_io__println({ ptr, i64 })

define void @aila_main__main() {
  %slot___lit_1 = alloca { i8, { { ptr, i64 } } }, align 8
  %slot_vargs_2 = alloca [1 x { i8, { { ptr, i64 } } }], align 8
  %t1 = insertvalue { i8, { { ptr, i64 } } } undef, i8 5, 0
  store { i8, { { ptr, i64 } } } %t1, ptr %slot___lit_1
  %t2 = insertvalue { ptr, i64 } undef, ptr @.str.0, 0
  %t3 = insertvalue { ptr, i64 } %t2, i64 12, 1
  %t4 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot___lit_1, i32 0, i32 1, i32 0
  store { ptr, i64 } %t3, ptr %t4
  %t5 = load { i8, { { ptr, i64 } } }, ptr %slot___lit_1
  %t6 = getelementptr inbounds { i8, { { ptr, i64 } } }, ptr %slot_vargs_2, i64 0
  store { i8, { { ptr, i64 } } } %t5, ptr %t6
  %t7 = insertvalue { ptr, i64 } undef, ptr %slot_vargs_2, 0
  %t8 = insertvalue { ptr, i64 } %t7, i64 1, 1
  call void @aila_std_io__println({ ptr, i64 } %t8)
  ret void
}

define i32 @main(i32 %argc, ptr %argv) {
entry:
  call void @aila_main__main()
  ret i32 0
}
