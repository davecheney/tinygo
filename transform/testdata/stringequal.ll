target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@zeroString = constant [0 x i8] zeroinitializer
@str1 = internal unnamed_addr constant [1 x i8] c"\80"
@str2 = internal unnamed_addr constant [2 x i8] c"\80b"
@str3 = internal unnamed_addr constant [3 x i8] c"\80bc"
@str4 = internal unnamed_addr constant [4 x i8] c"\80bc\9B"
@str7 = internal unnamed_addr constant [7 x i8] c"\80bc\9Bef\B6"
@str8 = internal unnamed_addr constant [8 x i8] c"\80bc\9Bef\B6h"
@str15 = internal unnamed_addr constant [15 x i8] c"\80bc\9Bef\B6hi\D1kl\ECno"
@str16 = internal unnamed_addr constant [16 x i8] c"\80bc\9Bef\B6hi\D1kl\ECno\07"
@str17 = internal unnamed_addr constant [17 x i8] c"\80bc\9Bef\B6hi\D1kl\ECno\07q"
@concat = internal unnamed_addr constant [11 x i8] c"hello world"
@zeros = internal unnamed_addr constant [4 x i8] zeroinitializer
@mutable = internal global [4 x i8] c"abcd"
@external = external constant [4 x i8]

declare i1 @runtime.stringEqual(ptr, i32, ptr, i32, ptr)

define i1 @main.stringCompareEqualConstantZero(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @zeroString, i32 0, ptr undef)
  ret i1 %0
}

define i1 @main.stringCompareUnequalConstantZero(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @zeroString, i32 0, ptr undef)
  %1 = xor i1 %0, true
  ret i1 %1
}

define i1 @main.stringCompareConst1(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @str1, i32 1, ptr undef)
  ret i1 %0
}

define i1 @main.stringCompareConst2(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @str2, i32 2, ptr undef)
  ret i1 %0
}

define i1 @main.stringCompareConst3(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @str3, i32 3, ptr undef)
  ret i1 %0
}

define i1 @main.stringCompareConst4(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @str4, i32 4, ptr undef)
  ret i1 %0
}

define i1 @main.stringCompareConst7(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @str7, i32 7, ptr undef)
  ret i1 %0
}

define i1 @main.stringCompareConst8(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @str8, i32 8, ptr undef)
  ret i1 %0
}

define i1 @main.stringCompareConst15(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @str15, i32 15, ptr undef)
  ret i1 %0
}

define i1 @main.stringCompareConst16(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @str16, i32 16, ptr undef)
  ret i1 %0
}

define i1 @main.stringCompareConst17(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @str17, i32 17, ptr undef)
  ret i1 %0
}

; Literal as the first operand.
define i1 @main.stringCompareConstLeft(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr @str4, i32 4, ptr %s1.data, i32 %s1.len, ptr undef)
  ret i1 %0
}

; Literal at a constant offset into a larger string ("world").
define i1 @main.stringCompareConstGEP(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr getelementptr inbounds (i8, ptr @concat, i32 6), i32 5, ptr undef)
  ret i1 %0
}

; Zeroinitializer string data.
define i1 @main.stringCompareConstZeros(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @zeros, i32 4, ptr undef)
  ret i1 %0
}

; The result is used in a branch and a phi in a successor block, which must
; be updated to refer to the new block.
define i32 @main.stringCompareBranch(ptr %s1.data, i32 %s1.len, i1 %c, ptr %context) {
entry:
  br i1 %c, label %cmp, label %done

cmp:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @str4, i32 4, ptr undef)
  br i1 %0, label %done, label %other

other:
  br label %done

done:
  %result = phi i32 [ 0, %entry ], [ 1, %cmp ], [ 2, %other ]
  ret i32 %result
}

; Both sides constant: folded.
define i1 @main.stringCompareBothConst(ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr @str4, i32 4, ptr getelementptr inbounds (i8, ptr @concat, i32 0), i32 4, ptr undef)
  %1 = call i1 @runtime.stringEqual(ptr @concat, i32 3, ptr getelementptr inbounds ([11 x i8], ptr @concat, i32 0, i32 0), i32 3, ptr undef)
  %2 = and i1 %0, %1
  ret i1 %2
}

; Both lengths constant but different: false.
define i1 @main.stringCompareDifferentLengths(ptr %s1.data, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 3, ptr @str4, i32 4, ptr undef)
  ret i1 %0
}

; Not optimized: non-constant lengths.
define i1 @main.stringCompareVarLength(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr undef)
  ret i1 %0
}

; Not optimized: constant length, but the data is not a constant global.
define i1 @main.stringCompareVarData(ptr %s1.data, i32 %s1.len, ptr %s2.data, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 4, ptr undef)
  %1 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @mutable, i32 4, ptr undef)
  %2 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @external, i32 4, ptr undef)
  %3 = and i1 %0, %1
  %4 = and i1 %3, %2
  ret i1 %4
}
