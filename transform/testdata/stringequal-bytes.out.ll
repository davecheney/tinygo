target datalayout = "e-m:e-p:32:32-i64:64-n32-S128"
target triple = "riscv32-unknown-none"

@str3 = internal unnamed_addr constant [3 x i8] c"\80bc"

declare i1 @runtime.stringEqual(ptr, i32, ptr, i32, ptr)

define i1 @main.stringCompareConst3(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %len.i = icmp eq i32 %s1.len, 3
  br i1 %len.i, label %compare.i, label %"runtime.stringEqual:const.exit"

compare.i:                                        ; preds = %entry
  %0 = load i8, ptr %s1.data, align 1
  %1 = icmp eq i8 %0, -128
  br i1 %1, label %compare1.i, label %"runtime.stringEqual:const.exit"

compare1.i:                                       ; preds = %compare.i
  %2 = getelementptr inbounds i8, ptr %s1.data, i32 1
  %3 = load i8, ptr %2, align 1
  %4 = icmp eq i8 %3, 98
  br i1 %4, label %compare2.i, label %"runtime.stringEqual:const.exit"

compare2.i:                                       ; preds = %compare1.i
  %5 = getelementptr inbounds i8, ptr %s1.data, i32 2
  %6 = load i8, ptr %5, align 1
  %7 = xor i8 %6, 99
  %bytes.i = icmp eq i8 %7, 0
  br label %"runtime.stringEqual:const.exit"

"runtime.stringEqual:const.exit":                 ; preds = %entry, %compare.i, %compare1.i, %compare2.i
  %result.i = phi i1 [ false, %entry ], [ false, %compare.i ], [ false, %compare1.i ], [ %bytes.i, %compare2.i ]
  ret i1 %result.i
}
