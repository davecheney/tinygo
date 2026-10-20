target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.a = global [256 x i8] zeroinitializer, align 256
@main.b = global [256 x i8] zeroinitializer, align 256
@main.ua = global i64 ptrtoint (ptr @main.a to i64)
@main.ub = global i64 ptrtoint (ptr @main.b to i64)
@main.ua1 = global i64 ptrtoint (ptr @main.a to i64)
@main.ub1 = global i64 ptrtoint (ptr @main.b to i64)
@main.ua2 = global i64 ptrtoint (ptr @main.a to i64)
@main.ua3 = global i64 ptrtoint (ptr @main.a to i64)
@main.addressesEqual = global i1 true
@main.bytesEqual = global i1 true
@main.byteIsZero = global i1 true
@main.bytesDiffer = global i1 true

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

define internal void @main.init(ptr %context) unnamed_addr {
entry:
  ; Compare whole addresses at compile time.
  %ua = load i64, ptr @main.ua
  %ub = load i64, ptr @main.ub
  %addressesEqual = icmp eq i64 %ua, %ub
  store i1 %addressesEqual, ptr @main.addressesEqual

  ; Compare bytes of addresses only at runtime.
  call void @main.compareBytes(ptr undef)
  call void @main.compareByteWithZero(ptr undef)
  call void @main.compareBytesOfSameAddress(ptr undef)
  ret void
}

define internal void @main.compareBytes(ptr %context) unnamed_addr {
entry:
  %av = load i8, ptr @main.ua1
  %bv = load i8, ptr @main.ub1
  %equal = icmp eq i8 %av, %bv
  store i1 %equal, ptr @main.bytesEqual
  ret void
}

define internal void @main.compareByteWithZero(ptr %context) unnamed_addr {
entry:
  %av = load i8, ptr @main.ua2
  %isZero = icmp eq i8 %av, 0
  store i1 %isZero, ptr @main.byteIsZero
  ret void
}

define internal void @main.compareBytesOfSameAddress(ptr %context) unnamed_addr {
entry:
  %b0 = load i8, ptr @main.ua3
  %b1 = load i8, ptr getelementptr inbounds (i8, ptr @main.ua3, i64 1)
  %notEqual = icmp ne i8 %b0, %b1
  store i1 %notEqual, ptr @main.bytesDiffer
  ret void
}
