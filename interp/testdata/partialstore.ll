target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.x = global i64 0
@main.y = global i64 0
@main.pa1 = global ptr @main.x
@main.pa2 = global ptr @main.x
@main.pa3 = global ptr @main.x
@main.pa4 = global ptr @main.x
@main.pa5 = global ptr @main.x
@main.pair = global [2 x ptr] [ptr @main.x, ptr @main.y]
@main.wholePointer = global ptr null
@main.wholeCopy = global ptr null
@main.wholeStruct = global { ptr, i64 } zeroinitializer
@main.firstByte = global i8 0
@main.field = global { i8, i8 } zeroinitializer
@main.escaped = global ptr null
@main.mixed = global i64 0
@main.array = global [2 x i8] zeroinitializer

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

define internal void @main.init(ptr %context) unnamed_addr {
entry:
  ; Store and copy whole pointers at compile time.
  %p = load ptr, ptr @main.pa1
  store ptr %p, ptr @main.wholePointer
  call void @llvm.memcpy.p0.p0.i64(ptr @main.wholeCopy, ptr @main.pa1, i64 8, i1 false)
  %s = insertvalue { ptr, i64 } { ptr null, i64 3 }, ptr %p, 0
  store { ptr, i64 } %s, ptr @main.wholeStruct

  ; Store part of a pointer only at runtime.
  call void @main.storeByte(ptr undef)
  call void @main.storeField(ptr undef)
  call void @main.storeEscaped(ptr undef)
  call void @main.storeMixed(ptr undef)

  ; Copy part of a pointer at runtime.
  call void @llvm.memcpy.p0.p0.i64(ptr @main.array, ptr @main.pa5, i64 2, i1 false)
  ret void
}

define internal void @main.storeByte(ptr %context) unnamed_addr {
entry:
  %b = load i8, ptr @main.pa2
  store i8 %b, ptr @main.firstByte
  ret void
}

define internal void @main.storeField(ptr %context) unnamed_addr {
entry:
  %b = load i8, ptr @main.pa3
  %f = insertvalue { i8, i8 } { i8 0, i8 5 }, i8 %b, 0
  store { i8, i8 } %f, ptr @main.field
  ret void
}

define internal void @main.storeEscaped(ptr %context) unnamed_addr {
entry:
  %buf = call ptr @runtime.alloc(i64 2, ptr inttoptr (i64 3 to ptr))
  %b = load i8, ptr @main.pa4
  store i8 %b, ptr %buf
  store ptr %buf, ptr @main.escaped
  ret void
}

define internal void @main.storeMixed(ptr %context) unnamed_addr {
entry:
  ; Load the high half of one pointer and the low half of the next pointer.
  %m = load i64, ptr getelementptr inbounds (i8, ptr @main.pair, i64 4)
  store i64 %m, ptr @main.mixed
  ret void
}

declare ptr @runtime.alloc(i64, ptr) unnamed_addr

declare void @llvm.memcpy.p0.p0.i64(ptr, ptr, i64, i1)
