target datalayout = "e-m:e-p:32:32-i64:64-n32-S128"
target triple = "riscv32-unknown-none"

@str3 = internal unnamed_addr constant [3 x i8] c"\80bc"

declare i1 @runtime.stringEqual(ptr, i32, ptr, i32, ptr)

; No fast unaligned loads: compare byte by byte.
define i1 @main.stringCompareConst3(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @str3, i32 3, ptr undef)
  ret i1 %0
}
