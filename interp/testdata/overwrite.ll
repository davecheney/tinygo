target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@byte.target = global [16 x i8] zeroinitializer
@byte.ptr = global ptr getelementptr (i8, ptr @byte.target, i64 4)
@byte.copy = global ptr null
@start.target = global i8 0
@start.value = global { ptr, [8 x i8] } { ptr @start.target, [8 x i8] zeroinitializer }
@end.target = global i8 0
@end.value = global { [8 x i8], ptr } { [8 x i8] zeroinitializer, ptr @end.target }
@copy.target = global i8 0
@copy.ptr = global ptr @copy.target
@copy.src = global [2 x i8] c"\01\02"
@whole.a = global i8 0
@whole.b = global i8 0
@whole.pair = global [2 x ptr] [ptr @whole.a, ptr @whole.a]
@whole.int = global [2 x ptr] [ptr @whole.a, ptr @whole.a]
@whole.near = global { ptr, [8 x i8] } { ptr @whole.a, [8 x i8] zeroinitializer }

define void @runtime.initAll() unnamed_addr {
entry:
  call void @byte.init(ptr undef)
  call void @start.init(ptr undef)
  call void @end.init(ptr undef)
  call void @copy.init(ptr undef)
  call void @whole.init(ptr undef)
  ret void
}

; Store one byte into a pointer. The store and the load after it must run at
; runtime.
define internal void @byte.init(ptr %context) unnamed_addr {
entry:
  store i8 1, ptr @byte.ptr
  %val = load ptr, ptr @byte.ptr
  store ptr %val, ptr @byte.copy
  ret void
}

; Store a value that starts inside a pointer.
define internal void @start.init(ptr %context) unnamed_addr {
entry:
  %dst = getelementptr i8, ptr @start.value, i64 6
  store [4 x i8] c"\01\02\03\04", ptr %dst
  ret void
}

; Store a value that ends inside a pointer.
define internal void @end.init(ptr %context) unnamed_addr {
entry:
  %dst = getelementptr i8, ptr @end.value, i64 6
  store [4 x i8] c"\01\02\03\04", ptr %dst
  ret void
}

; Copy bytes into the middle of a pointer.
define internal void @copy.init(ptr %context) unnamed_addr {
entry:
  %dst = getelementptr i8, ptr @copy.ptr, i64 3
  call void @llvm.memcpy.p0.p0.i64(ptr %dst, ptr @copy.src, i64 2, i1 false)
  ret void
}

; Replace whole pointers, or write next to a pointer. These must still run at
; compile time, also when two equal pointers are next to each other.
define internal void @whole.init(ptr %context) unnamed_addr {
entry:
  %pair = getelementptr i8, ptr @whole.pair, i64 8
  store ptr @whole.b, ptr %pair
  store i64 0, ptr @whole.int
  %near = getelementptr i8, ptr @whole.near, i64 8
  store i8 1, ptr %near
  ret void
}

declare void @llvm.memcpy.p0.p0.i64(ptr, ptr, i64, i1)
