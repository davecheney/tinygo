target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@x = global i64 0
@pair = local_unnamed_addr global [2 x ptr] [ptr @x, ptr @x]
@main.whole = local_unnamed_addr global ptr @x
@main.wholeCopy = local_unnamed_addr global i64 ptrtoint (ptr @x to i64)
@main.mid = local_unnamed_addr global i64 0
@main.midCopy = local_unnamed_addr global i64 0

define void @runtime.initAll() local_unnamed_addr {
  %mid = load i64, ptr getelementptr inbounds (i8, ptr @pair, i64 4), align 8
  store i64 %mid, ptr @main.mid, align 8
  call void @llvm.memcpy.p0.p0.i64(ptr @main.midCopy, ptr getelementptr inbounds (i8, ptr @pair, i64 4), i64 8, i1 false)
  ret void
}

declare void @llvm.memcpy.p0.p0.i64(ptr noalias nocapture writeonly, ptr noalias nocapture readonly, i64, i1 immarg) #0

attributes #0 = { nocallback nofree nounwind willreturn memory(argmem: readwrite) }
