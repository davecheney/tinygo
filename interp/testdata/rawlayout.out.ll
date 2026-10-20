target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.x = global i64 0
@main.pointerFree = local_unnamed_addr global ptr @"main$alloc"
@main.misaligned = local_unnamed_addr global ptr @"main$alloc.1"
@main.self = local_unnamed_addr global ptr @"main$alloc.2"
@main.straddle = local_unnamed_addr global ptr @"main$alloc.3"
@main.aligned = local_unnamed_addr global ptr @"main$alloc.4"
@"main$alloc" = internal global [16 x i8] zeroinitializer, align 8
@"main$alloc.1" = internal global [16 x i8] zeroinitializer, align 8
@"main$alloc.2" = internal global [16 x i8] zeroinitializer, align 8
@"main$alloc.3" = internal global { ptr, i64, i64 } { ptr @main.x, i64 0, i64 0 }, align 8
@"main$alloc.4" = internal global { ptr, i64, i64 } { ptr null, i64 ptrtoint (ptr @main.x to i64), i64 0 }, align 8

define void @runtime.initAll() unnamed_addr {
entry:
  store i64 ptrtoint (ptr @main.x to i64), ptr getelementptr inbounds (i8, ptr @"main$alloc", i32 8), align 8
  store i64 ptrtoint (ptr @main.x to i64), ptr getelementptr inbounds (i8, ptr @"main$alloc.1", i32 4), align 4
  store i64 ptrtoint (ptr @"main$alloc.2" to i64), ptr getelementptr inbounds (i8, ptr @"main$alloc.2", i32 8), align 8
  store i64 ptrtoint (ptr @main.x to i64), ptr getelementptr inbounds (i8, ptr @"main$alloc.3", i32 12), align 4
  ret void
}
