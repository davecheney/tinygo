target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

@main.arr = global [4 x i32] [i32 10, i32 20, i32 30, i32 40]
@main.g1 = local_unnamed_addr global [4 x i32] [i32 101, i32 102, i32 103, i32 104]
@main.g2 = local_unnamed_addr global [4 x i32] [i32 201, i32 202, i32 203, i32 204]
@main.g3 = local_unnamed_addr global [4 x i32] [i32 301, i32 302, i32 303, i32 304]
@main.g4 = local_unnamed_addr global [4 x i32] [i32 401, i32 402, i32 403, i32 404]
@main.out1 = local_unnamed_addr global i32 40
@main.out2 = local_unnamed_addr global i32 30
@main.p = local_unnamed_addr global ptr null

define void @runtime.initAll() unnamed_addr {
entry:
  store ptr getelementptr (i8, ptr @main.arr, i32 -1), ptr @main.p, align 4
  ret void
}
