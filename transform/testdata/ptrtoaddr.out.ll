target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@global = global i32 0
@table = global [16 x i32] zeroinitializer

declare void @use(i32)

declare void @runtime.printptr(i32, ptr)

declare void @runtime.printint32(i32, ptr)

define i1 @alignCheck(ptr %p) {
  %addr = ptrtoaddr ptr %p to i32
  %rem = and i32 %addr, 7
  %cmp = icmp eq i32 %rem, 0
  ret i1 %cmp
}

define i32 @hashIndex(ptr %p, i1 %c) {
entry:
  %addr = ptrtoaddr ptr %p to i32
  br i1 %c, label %a, label %b

a:                                                ; preds = %entry
  %shr = lshr i32 %addr, 4
  br label %b

b:                                                ; preds = %a, %entry
  %h = phi i32 [ %addr, %entry ], [ %shr, %a ]
  %idx = and i32 %h, 15
  %gep = getelementptr [16 x i32], ptr @table, i32 0, i32 %idx
  %v = load i32, ptr %gep, align 4
  ret i32 %v
}

define void @switchOnAddr(ptr %p) {
  %addr = ptrtoaddr ptr %p to i32
  switch i32 %addr, label %default [
    i32 0, label %zero
  ]

zero:                                             ; preds = %0
  ret void

default:                                          ; preds = %0
  ret void
}

define void @printPtr(ptr %p) {
  %addr = ptrtoaddr ptr %p to i32
  call void @runtime.printptr(i32 %addr, ptr undef)
  ret void
}

define ptr @roundTrip(ptr %p) {
  %addr = ptrtoint ptr %p to i32
  %add = add i32 %addr, 4
  %q = inttoptr i32 %add to ptr
  ret ptr %q
}

define void @stored(ptr %p) {
  %addr = ptrtoint ptr %p to i32
  store i32 %addr, ptr @global, align 4
  ret void
}

define void @passed(ptr %p) {
  %addr = ptrtoint ptr %p to i32
  call void @use(i32 %addr)
  ret void
}

define void @printedAsInt(ptr %p) {
  %addr = ptrtoint ptr %p to i32
  call void @runtime.printint32(i32 %addr, ptr undef)
  ret void
}

define i32 @returned(ptr %p) {
  %addr = ptrtoint ptr %p to i32
  %masked = and i32 %addr, -8
  ret i32 %masked
}

define void @selectStored(ptr %p, i1 %c) {
  %addr = ptrtoint ptr %p to i32
  %sel = select i1 %c, i32 %addr, i32 0
  store i32 %sel, ptr @global, align 4
  ret void
}

define i1 @narrow(ptr %p) {
  %addr = ptrtoint ptr %p to i8
  %cmp = icmp eq i8 %addr, 0
  ret i1 %cmp
}

define i1 @compareNull(ptr %p) {
  %cmp = icmp ne ptr %p, null
  ret i1 %cmp
}

define i1 @comparePtrs(ptr %p, ptr %q) {
  %a = ptrtoaddr ptr %p to i32
  %cmp = icmp ult ptr %q, %p
  %mask = and i32 %a, 3
  %aligned = icmp eq i32 %mask, 0
  %res = and i1 %cmp, %aligned
  ret i1 %res
}
