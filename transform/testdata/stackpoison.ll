target datalayout = "e-m:e-i64:64-n32:64-S128"

declare ptr @runtime.alloc(i64, ptr)

declare void @consume(ptr nocapture)

define i32 @local(i1 %cond) {
entry:
  %p = call ptr @runtime.alloc(i64 16, ptr null)
  call void @consume(ptr %p)
  br i1 %cond, label %yes, label %no

yes:
  ret i32 42

no:
  ret i32 43
}

define ptr @escape() {
entry:
  %p = call ptr @runtime.alloc(i64 16, ptr null)
  ret ptr %p
}
