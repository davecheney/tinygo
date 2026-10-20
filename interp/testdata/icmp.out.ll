target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.a = global [256 x i8] zeroinitializer, align 256
@main.b = global [256 x i8] zeroinitializer, align 256
@main.ua = local_unnamed_addr global i64 ptrtoint (ptr @main.a to i64)
@main.ub = local_unnamed_addr global i64 ptrtoint (ptr @main.b to i64)
@main.ua1 = local_unnamed_addr global i64 ptrtoint (ptr @main.a to i64)
@main.ub1 = local_unnamed_addr global i64 ptrtoint (ptr @main.b to i64)
@main.ua2 = local_unnamed_addr global i64 ptrtoint (ptr @main.a to i64)
@main.ua3 = local_unnamed_addr global i64 ptrtoint (ptr @main.a to i64)
@main.addressesEqual = local_unnamed_addr global i1 false
@main.bytesEqual = local_unnamed_addr global i1 true
@main.byteIsZero = local_unnamed_addr global i1 true
@main.bytesDiffer = local_unnamed_addr global i1 true

define void @runtime.initAll() unnamed_addr {
entry:
  call fastcc void @main.compareBytes(ptr undef)
  call fastcc void @main.compareByteWithZero(ptr undef)
  call fastcc void @main.compareBytesOfSameAddress(ptr undef)
  ret void
}

define internal fastcc void @main.compareBytes(ptr %context) unnamed_addr {
entry:
  %av = load i8, ptr @main.ua1, align 1
  %bv = load i8, ptr @main.ub1, align 1
  %equal = icmp eq i8 %av, %bv
  store i1 %equal, ptr @main.bytesEqual, align 1
  ret void
}

define internal fastcc void @main.compareByteWithZero(ptr %context) unnamed_addr {
entry:
  %av = load i8, ptr @main.ua2, align 1
  %isZero = icmp eq i8 %av, 0
  store i1 %isZero, ptr @main.byteIsZero, align 1
  ret void
}

define internal fastcc void @main.compareBytesOfSameAddress(ptr %context) unnamed_addr {
entry:
  %b0 = load i8, ptr @main.ua3, align 1
  %b1 = load i8, ptr getelementptr inbounds (i8, ptr @main.ua3, i64 1), align 1
  %notEqual = icmp ne i8 %b0, %b1
  store i1 %notEqual, ptr @main.bytesDiffer, align 1
  ret void
}
