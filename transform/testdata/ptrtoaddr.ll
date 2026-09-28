target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@global = global i32 0
@table = global [16 x i32] zeroinitializer

declare void @use(i32)
declare void @runtime.printptr(i32, ptr)
declare void @runtime.printint32(i32, ptr)

; Alignment check: only compared, so it can be a ptrtoaddr.
define i1 @alignCheck(ptr %p) {
  %addr = ptrtoint ptr %p to i32
  %rem = and i32 %addr, 7
  %cmp = icmp eq i32 %rem, 0
  ret i1 %cmp
}

; Pointer hash used as a table index (through a phi).
define i32 @hashIndex(ptr %p, i1 %c) {
entry:
  %addr = ptrtoint ptr %p to i32
  br i1 %c, label %a, label %b
a:
  %shr = lshr i32 %addr, 4
  br label %b
b:
  %h = phi i32 [ %addr, %entry ], [ %shr, %a ]
  %idx = and i32 %h, 15
  %gep = getelementptr [16 x i32], ptr @table, i32 0, i32 %idx
  %v = load i32, ptr %gep
  ret i32 %v
}

; Switch on the address.
define void @switchOnAddr(ptr %p) {
  %addr = ptrtoint ptr %p to i32
  switch i32 %addr, label %default [ i32 0, label %zero ]
zero:
  ret void
default:
  ret void
}

; Printing a pointer only formats the address.
define void @printPtr(ptr %p) {
  %addr = ptrtoint ptr %p to i32
  call void @runtime.printptr(i32 %addr, ptr undef)
  ret void
}

; Round trip through inttoptr: must stay a ptrtoint.
define ptr @roundTrip(ptr %p) {
  %addr = ptrtoint ptr %p to i32
  %add = add i32 %addr, 4
  %q = inttoptr i32 %add to ptr
  ret ptr %q
}

; Stored to memory: must stay a ptrtoint.
define void @stored(ptr %p) {
  %addr = ptrtoint ptr %p to i32
  store i32 %addr, ptr @global
  ret void
}

; Passed to an unknown call: must stay a ptrtoint.
define void @passed(ptr %p) {
  %addr = ptrtoint ptr %p to i32
  call void @use(i32 %addr)
  ret void
}

; Passed to a print function that isn't printptr: must stay a ptrtoint.
define void @printedAsInt(ptr %p) {
  %addr = ptrtoint ptr %p to i32
  call void @runtime.printint32(i32 %addr, ptr undef)
  ret void
}

; Returned: must stay a ptrtoint.
define i32 @returned(ptr %p) {
  %addr = ptrtoint ptr %p to i32
  %masked = and i32 %addr, -8
  ret i32 %masked
}

; Selected and then stored: must stay a ptrtoint.
define void @selectStored(ptr %p, i1 %c) {
  %addr = ptrtoint ptr %p to i32
  %sel = select i1 %c, i32 %addr, i32 0
  store i32 %sel, ptr @global
  ret void
}

; Not the pointer width: ptrtoaddr requires the index width.
define i1 @narrow(ptr %p) {
  %addr = ptrtoint ptr %p to i8
  %cmp = icmp eq i8 %addr, 0
  ret i1 %cmp
}

; Comparisons with zero or with another ptrtoint are turned into pointer
; comparisons, which InstCombine would otherwise have done.
define i1 @compareNull(ptr %p) {
  %addr = ptrtoint ptr %p to i32
  %cmp = icmp ne i32 %addr, 0
  ret i1 %cmp
}

define i1 @comparePtrs(ptr %p, ptr %q) {
  %a = ptrtoint ptr %p to i32
  %b = ptrtoint ptr %q to i32
  %cmp = icmp ult i32 %b, %a
  %mask = and i32 %a, 3
  %aligned = icmp eq i32 %mask, 0
  %res = and i1 %cmp, %aligned
  ret i1 %res
}
