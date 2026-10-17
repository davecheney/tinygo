target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@x = global i64 0
@pair = global [2 x ptr] [ptr @x, ptr @x]
@main.whole = global ptr null
@main.wholeCopy = global i64 0
@main.mid = global i64 0
@main.midCopy = global i64 0

define void @runtime.initAll() {
  call void @main.init()
  ret void
}

define internal void @main.init() {
  ; A whole pointer next to an equal pointer is known at compile time.
  %whole = load ptr, ptr getelementptr inbounds (i8, ptr @pair, i64 8)
  store ptr %whole, ptr @main.whole
  call void @llvm.memcpy.p0.p0.i64(ptr @main.wholeCopy, ptr getelementptr inbounds (i8, ptr @pair, i64 8), i64 8, i1 false)

  ; These bytes are the end of one pointer and the start of the next pointer.
  ; They are known only at runtime.
  %mid = load i64, ptr getelementptr inbounds (i8, ptr @pair, i64 4)
  store i64 %mid, ptr @main.mid
  call void @llvm.memcpy.p0.p0.i64(ptr @main.midCopy, ptr getelementptr inbounds (i8, ptr @pair, i64 4), i64 8, i1 false)
  ret void
}

declare void @llvm.memcpy.p0.p0.i64(ptr noalias nocapture writeonly, ptr noalias nocapture readonly, i64, i1 immarg)
