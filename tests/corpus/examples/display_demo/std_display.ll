; AILA module: std.display
; target: x86_64-w64-windows-gnu
target triple = "x86_64-w64-windows-gnu"


define { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 } @aila_std_display__new_commit(ptr %p_fb, { { ptr, i64 } } %p_dmg) {
  %slot_fb_1 = alloca { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, align 8
  %slot_dmg_2 = alloca { { ptr, i64 } }, align 8
  %slot___lit_3 = alloca { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, align 8
  store { { ptr, i64 } } %p_dmg, ptr %slot_dmg_2
  %t1 = load { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %p_fb
  %t2 = getelementptr inbounds { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, ptr %slot___lit_3, i32 0, i32 0
  store { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } %t1, ptr %t2
  %t3 = load { { ptr, i64 } }, ptr %slot_dmg_2
  %t4 = getelementptr inbounds { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, ptr %slot___lit_3, i32 0, i32 1
  store { { ptr, i64 } } %t3, ptr %t4
  %t5 = getelementptr inbounds { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, ptr %slot___lit_3, i32 0, i32 2
  store i1 1, ptr %t5
  %t6 = load { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, ptr %slot___lit_3
  ret { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 } %t6
}

define void @aila_std_display__Commit__set_damage(ptr %p_self, { { ptr, i64 } } %p_d) {
  %slot_self_1 = alloca { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, align 8
  %slot_d_2 = alloca { { ptr, i64 } }, align 8
  store { { ptr, i64 } } %p_d, ptr %slot_d_2
  %t1 = getelementptr inbounds { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, ptr %p_self, i32 0, i32 1
  %t2 = load { { ptr, i64 } }, ptr %slot_d_2
  store { { ptr, i64 } } %t2, ptr %t1
  ret void
}

define { { ptr, i64 } } @aila_std_display__Commit__damage_of(ptr %p_self) {
  %slot_self_1 = alloca { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, align 8
  %t1 = getelementptr inbounds { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, ptr %p_self, i32 0, i32 1
  %t2 = load { { ptr, i64 } }, ptr %t1
  ret { { ptr, i64 } } %t2
}

define void @aila_std_display__Commit__set_enable(ptr %p_self, i1 %p_on) {
  %slot_self_1 = alloca { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, align 8
  %slot_on_2 = alloca i1, align 8
  store i1 %p_on, ptr %slot_on_2
  %t1 = getelementptr inbounds { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, ptr %p_self, i32 0, i32 2
  %t2 = load i1, ptr %slot_on_2
  store i1 %t2, ptr %t1
  ret void
}

define i1 @aila_std_display__Commit__enabled(ptr %p_self) {
  %slot_self_1 = alloca { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, align 8
  %t1 = getelementptr inbounds { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, ptr %p_self, i32 0, i32 2
  %t2 = load i1, ptr %t1
  ret i1 %t2
}

define { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } @aila_std_display__Commit__buf(ptr %p_self) {
  %slot_self_1 = alloca { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, align 8
  %t1 = getelementptr inbounds { { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, { { ptr, i64 } }, i1 }, ptr %p_self, i32 0, i32 0
  %t2 = load { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } }, ptr %t1
  ret { { ptr, i64 }, i16, i16, i16, { i8, i8, i8, i8 } } %t2
}
