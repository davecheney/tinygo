target datalayout = "E-m:m-p:32:32-i8:8:32-i16:16:32-i64:64-n32-S64"
target triple = "mips-unknown-linux-gnu"

@zeroString = constant [0 x i8] zeroinitializer
@str3 = internal unnamed_addr constant [3 x i8] c"\80bc"
@str4 = internal unnamed_addr constant [4 x i8] c"\80bc\9B"
@str8 = internal unnamed_addr constant [8 x i8] c"\80bc\9Bef\B6h"

declare i1 @runtime.stringEqual(ptr, i32, ptr, i32, ptr)

define i1 @main.stringCompareConst3(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 3
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.2.exit"

compare.i:                                        ; preds = %entry
  %0 = load i16, ptr %s1.data, align 1
  %1 = xor i16 %0, -32670
  %2 = getelementptr inbounds i8, ptr %s1.data, i32 1
  %3 = load i16, ptr %2, align 1
  %4 = xor i16 %3, 25187
  %5 = or i16 %1, %4
  %bytes.i = icmp eq i16 %5, 0
  br label %"runtime.stringEqual:const.2.exit"

"runtime.stringEqual:const.2.exit":               ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConst4(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 4
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.1.exit"

compare.i:                                        ; preds = %entry
  %0 = load i32, ptr %s1.data, align 1
  %1 = xor i32 %0, -2141035621
  %bytes.i = icmp eq i32 %1, 0
  br label %"runtime.stringEqual:const.1.exit"

"runtime.stringEqual:const.1.exit":               ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}

define i1 @main.stringCompareConst8(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 8
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.exit"

compare.i:                                        ; preds = %entry
  %0 = load i32, ptr %s1.data, align 1
  %1 = xor i32 %0, -2141035621
  %2 = getelementptr inbounds i8, ptr %s1.data, i32 4
  %3 = load i32, ptr %2, align 1
  %4 = xor i32 %3, 1701230184
  %5 = or i32 %1, %4
  %bytes.i = icmp eq i32 %5, 0
  br label %"runtime.stringEqual:const.exit"

"runtime.stringEqual:const.exit":                 ; preds = %entry, %compare.i
  %result.i = phi i1 [ false, %entry ], [ %bytes.i, %compare.i ]
  ret i1 %result.i
}
