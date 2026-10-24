target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.arr = global [4 x i32] [i32 10, i32 20, i32 30, i32 40]
@main.g1 = global [4 x i32] [i32 101, i32 102, i32 103, i32 104]
@main.g2 = global [4 x i32] [i32 201, i32 202, i32 203, i32 204]
@main.g3 = global [4 x i32] [i32 301, i32 302, i32 303, i32 304]
@main.g4 = global [4 x i32] [i32 401, i32 402, i32 403, i32 404]
@main.s = global { i32, [4 x i16] } { i32 1, [4 x i16] [i16 2, i16 3, i16 4, i16 5] }
@main.out1 = global i32 0
@main.out2 = global i32 0
@main.out3 = global i16 0
@main.p = global ptr null

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

define internal void @main.init(ptr %context) unnamed_addr {
entry:
  %a = load i32, ptr @main.arr
  %b = load i32, ptr @main.g1
  %c = load i32, ptr @main.g2
  %d = load i32, ptr @main.g3
  %e = load i32, ptr @main.g4

  ; GEP indices are signed.
  ; https://llvm.org/docs/LangRef.html#getelementptr-instruction
  %x1 = load i32, ptr getelementptr ([4 x i32], ptr @main.arr, i32 1, i32 -1)
  store i32 %x1, ptr @main.out1
  %x2 = load i32, ptr getelementptr (i8, ptr getelementptr ([4 x i32], ptr @main.arr, i32 0, i32 3), i16 -8)
  store i32 %x2, ptr @main.out2
  %x3 = load i16, ptr getelementptr ({ i32, [4 x i16] }, ptr @main.s, i32 0, i32 1, i32 2)
  store i16 %x3, ptr @main.out3

  ; A pointer before the start of the object is left for runtime.
  store ptr getelementptr (i8, ptr @main.arr, i64 -1), ptr @main.p
  ret void
}
