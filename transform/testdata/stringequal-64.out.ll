target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

@zeroString = constant [0 x i8] zeroinitializer
@str7 = internal unnamed_addr constant [7 x i8] c"\80bc\9Bef\B6"
@str8 = internal unnamed_addr constant [8 x i8] c"\80bc\9Bef\B6h"
@str15 = internal unnamed_addr constant [15 x i8] c"\80bc\9Bef\B6hi\D1kl\ECno"
@str16 = internal unnamed_addr constant [16 x i8] c"\80bc\9Bef\B6hi\D1kl\ECno\07"
@str17 = internal unnamed_addr constant [17 x i8] c"\80bc\9Bef\B6hi\D1kl\ECno\07q"

declare i1 @runtime.stringEqual(ptr, i32, ptr, i32, ptr)

define i1 @main.stringCompareConst7(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 7
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.3.exit"

compare.i:                                        ; preds = %entry
  %0 = load i32, ptr %s1.data, align 1
  %1 = xor i32 %0, -1687985536
  %2 = getelementptr inbounds i8, ptr %s1.data, i32 3
  %3 = load i32, ptr %2, align 1
  %4 = xor i32 %3, -1234803301
  %5 = or i32 %1, %4
  %bytes.i = icmp eq i32 %5, 0
  br label %"runtime.stringEqual:const.3.exit"

"runtime.stringEqual:const.3.exit":               ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConst8(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 8
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.2.exit"

compare.i:                                        ; preds = %entry
  %0 = load i64, ptr %s1.data, align 1
  %1 = xor i64 %0, 7545330812290556544
  %bytes.i = icmp eq i64 %1, 0
  br label %"runtime.stringEqual:const.2.exit"

"runtime.stringEqual:const.2.exit":               ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConst15(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 15
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.1.exit"

compare.i:                                        ; preds = %entry
  %0 = load i64, ptr %s1.data, align 1
  %1 = xor i64 %0, 7545330812290556544
  %2 = getelementptr inbounds i8, ptr %s1.data, i32 7
  %3 = load i64, ptr %2, align 1
  %4 = xor i64 %3, 8029615136057682280
  %5 = or i64 %1, %4
  %bytes.i = icmp eq i64 %5, 0
  br label %"runtime.stringEqual:const.1.exit"

"runtime.stringEqual:const.1.exit":               ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConst16(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 16
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.exit"

compare.i:                                        ; preds = %entry
  %0 = load i64, ptr %s1.data, align 1
  %1 = xor i64 %0, 7545330812290556544
  %2 = getelementptr inbounds i8, ptr %s1.data, i32 8
  %3 = load i64, ptr %2, align 1
  %4 = xor i64 %3, 535768842390720873
  %5 = or i64 %1, %4
  %bytes.i = icmp eq i64 %5, 0
  br label %"runtime.stringEqual:const.exit"

"runtime.stringEqual:const.exit":                 ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConst17(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @str17, i32 17, ptr undef)
  ret i1 %0
}
