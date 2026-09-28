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
  %0 = icmp eq i32 %s1.len, 0
  ret i1 %0
}

define i1 @main.stringCompareUnequalConstantZero(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = icmp eq i32 %s1.len, 0
  %1 = xor i1 %0, true
  ret i1 %1
}

define i1 @main.stringCompareConst1(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 1
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.9.exit"

compare.i:                                        ; preds = %entry
  %0 = load i8, ptr %s1.data, align 1
  %1 = xor i8 %0, -128
  %bytes.i = icmp eq i8 %1, 0
  br label %"runtime.stringEqual:const.9.exit"

"runtime.stringEqual:const.9.exit":               ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConst2(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 2
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.8.exit"

compare.i:                                        ; preds = %entry
  %0 = load i16, ptr %s1.data, align 1
  %1 = xor i16 %0, 25216
  %bytes.i = icmp eq i16 %1, 0
  br label %"runtime.stringEqual:const.8.exit"

"runtime.stringEqual:const.8.exit":               ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConst3(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 3
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.7.exit"

compare.i:                                        ; preds = %entry
  %0 = load i16, ptr %s1.data, align 1
  %1 = xor i16 %0, 25216
  %2 = getelementptr inbounds i8, ptr %s1.data, i32 1
  %3 = load i16, ptr %2, align 1
  %4 = xor i16 %3, 25442
  %5 = or i16 %1, %4
  %bytes.i = icmp eq i16 %5, 0
  br label %"runtime.stringEqual:const.7.exit"

"runtime.stringEqual:const.7.exit":               ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConst4(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 4
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.exit"

compare.i:                                        ; preds = %entry
  %0 = load i32, ptr %s1.data, align 1
  %1 = xor i32 %0, -1687985536
  %bytes.i = icmp eq i32 %1, 0
  br label %"runtime.stringEqual:const.exit"

"runtime.stringEqual:const.exit":                 ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConst7(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 7
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.6.exit"

compare.i:                                        ; preds = %entry
  %0 = load i32, ptr %s1.data, align 1
  %1 = xor i32 %0, -1687985536
  %2 = getelementptr inbounds i8, ptr %s1.data, i32 3
  %3 = load i32, ptr %2, align 1
  %4 = xor i32 %3, -1234803301
  %5 = or i32 %1, %4
  %bytes.i = icmp eq i32 %5, 0
  br label %"runtime.stringEqual:const.6.exit"

"runtime.stringEqual:const.6.exit":               ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConst8(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 8
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.5.exit"

compare.i:                                        ; preds = %entry
  %0 = load i32, ptr %s1.data, align 1
  %1 = xor i32 %0, -1687985536
  %2 = getelementptr inbounds i8, ptr %s1.data, i32 4
  %3 = load i32, ptr %2, align 1
  %4 = xor i32 %3, 1756784229
  %5 = or i32 %1, %4
  %bytes.i = icmp eq i32 %5, 0
  br label %"runtime.stringEqual:const.5.exit"

"runtime.stringEqual:const.5.exit":               ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConst15(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 15
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.4.exit"

compare.i:                                        ; preds = %entry
  %0 = load i32, ptr %s1.data, align 1
  %1 = xor i32 %0, -1687985536
  %2 = getelementptr inbounds i8, ptr %s1.data, i32 4
  %3 = load i32, ptr %2, align 1
  %4 = xor i32 %3, 1756784229
  %5 = or i32 %1, %4
  %6 = getelementptr inbounds i8, ptr %s1.data, i32 8
  %7 = load i32, ptr %6, align 1
  %8 = xor i32 %7, 1819005289
  %9 = or i32 %5, %8
  %10 = getelementptr inbounds i8, ptr %s1.data, i32 11
  %11 = load i32, ptr %10, align 1
  %12 = xor i32 %11, 1869540460
  %13 = or i32 %9, %12
  %bytes.i = icmp eq i32 %13, 0
  br label %"runtime.stringEqual:const.4.exit"

"runtime.stringEqual:const.4.exit":               ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConst16(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 16
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.3.exit"

compare.i:                                        ; preds = %entry
  %0 = load i32, ptr %s1.data, align 1
  %1 = xor i32 %0, -1687985536
  %2 = getelementptr inbounds i8, ptr %s1.data, i32 4
  %3 = load i32, ptr %2, align 1
  %4 = xor i32 %3, 1756784229
  %5 = or i32 %1, %4
  %6 = getelementptr inbounds i8, ptr %s1.data, i32 8
  %7 = load i32, ptr %6, align 1
  %8 = xor i32 %7, 1819005289
  %9 = or i32 %5, %8
  %10 = getelementptr inbounds i8, ptr %s1.data, i32 12
  %11 = load i32, ptr %10, align 1
  %12 = xor i32 %11, 124743404
  %13 = or i32 %9, %12
  %bytes.i = icmp eq i32 %13, 0
  br label %"runtime.stringEqual:const.3.exit"

"runtime.stringEqual:const.3.exit":               ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConst17(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @str17, i32 17, ptr undef)
  ret i1 %0
}

define i1 @main.stringCompareConstLeft(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 4
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.exit"

compare.i:                                        ; preds = %entry
  %0 = load i32, ptr %s1.data, align 1
  %1 = xor i32 %0, -1687985536
  %bytes.i = icmp eq i32 %1, 0
  br label %"runtime.stringEqual:const.exit"

"runtime.stringEqual:const.exit":                 ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConstGEP(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 5
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.2.exit"

compare.i:                                        ; preds = %entry
  %0 = load i32, ptr %s1.data, align 1
  %1 = xor i32 %0, 1819438967
  %2 = getelementptr inbounds i8, ptr %s1.data, i32 1
  %3 = load i32, ptr %2, align 1
  %4 = xor i32 %3, 1684828783
  %5 = or i32 %1, %4
  %bytes.i = icmp eq i32 %5, 0
  br label %"runtime.stringEqual:const.2.exit"

"runtime.stringEqual:const.2.exit":               ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConstZeros(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 4
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.1.exit"

compare.i:                                        ; preds = %entry
  %0 = load i32, ptr %s1.data, align 1
  %bytes.i = icmp eq i32 %0, 0
  br label %"runtime.stringEqual:const.1.exit"

"runtime.stringEqual:const.1.exit":               ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i32 @main.stringCompareBranch(ptr %s1.data, i32 %s1.len, i1 %c, ptr %context) {
entry:
  br i1 %c, label %cmp, label %done

cmp:                                              ; preds = %entry
  %len.i = icmp eq i32 %s1.len, 4
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.exit"

compare.i:                                        ; preds = %cmp
  %0 = load i32, ptr %s1.data, align 1
  %1 = xor i32 %0, -1687985536
  %bytes.i = icmp eq i32 %1, 0
  br label %"runtime.stringEqual:const.exit"

"runtime.stringEqual:const.exit":                 ; preds = %cmp, %compare.i
  %result.i = phi i1 [ false, %cmp ], [ %bytes.i, %compare.i ]
  br i1 %result.i, label %done, label %other

other:                                            ; preds = %"runtime.stringEqual:const.exit"
  br label %done

done:                                             ; preds = %other, %"runtime.stringEqual:const.exit", %entry
  %result = phi i32 [ 0, %entry ], [ 1, %"runtime.stringEqual:const.exit" ], [ 2, %other ]
  ret i32 %result
}

define i1 @main.stringCompareBothConst(ptr %context) {
entry:
  %0 = and i1 false, true
  ret i1 %0
}

define i1 @main.stringCompareDifferentLengths(ptr %s1.data, ptr %context) {
entry:
  ret i1 false
}

define i1 @main.stringCompareVarLength(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr undef)
  ret i1 %0
}

define i1 @main.stringCompareVarData(ptr %s1.data, i32 %s1.len, ptr %s2.data, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 4, ptr undef)
  %1 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @mutable, i32 4, ptr undef)
  %2 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @external, i32 4, ptr undef)
  %3 = and i1 %0, %1
  %4 = and i1 %3, %2
  ret i1 %4
}
