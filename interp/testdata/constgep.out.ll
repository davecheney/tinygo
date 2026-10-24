target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.arr = global [4 x i32] [i32 10, i32 20, i32 30, i32 40]
@main.g1 = local_unnamed_addr global [4 x i32] [i32 101, i32 102, i32 103, i32 104]
@main.g2 = local_unnamed_addr global [4 x i32] [i32 201, i32 202, i32 203, i32 204]
@main.g3 = local_unnamed_addr global [4 x i32] [i32 301, i32 302, i32 303, i32 304]
@main.g4 = local_unnamed_addr global [4 x i32] [i32 401, i32 402, i32 403, i32 404]
@main.s = local_unnamed_addr global { i32, [4 x i16] } { i32 1, [4 x i16] [i16 2, i16 3, i16 4, i16 5] }
@main.out1 = local_unnamed_addr global i32 40
@main.out2 = local_unnamed_addr global i32 20
@main.out3 = local_unnamed_addr global i16 4
@main.p = local_unnamed_addr global ptr null

define void @runtime.initAll() unnamed_addr {
entry:
  store ptr getelementptr (i8, ptr @main.arr, i64 -1), ptr @main.p, align 8
  ret void
}
