; AILA module: _.main
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"

@.str.0 = private unnamed_addr constant [13 x i8] c"hello, AILA!\00", align 1

declare void @aila_std_io__print_text({ ptr, i64 })
declare void @aila_std_io__print_nl()

define void @aila_main__main() {
  %t1 = insertvalue { ptr, i64 } undef, ptr @.str.0, 0
  %t2 = insertvalue { ptr, i64 } %t1, i64 12, 1
  call void @aila_std_io__print_text({ ptr, i64 } %t2)
  call void @aila_std_io__print_nl()
  ret void
}

define i32 @main() {
entry:
  call void @aila_main__main()
  ret i32 0
}
