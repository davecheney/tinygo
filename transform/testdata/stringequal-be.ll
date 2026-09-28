target datalayout = "E-m:m-p:32:32-i8:8:32-i16:16:32-i64:64-n32-S64"
target triple = "mips-unknown-linux-gnu"

@zeroString = constant [0 x i8] zeroinitializer
@str3 = internal unnamed_addr constant [3 x i8] c"\80bc"
@str4 = internal unnamed_addr constant [4 x i8] c"\80bc\9B"
@str8 = internal unnamed_addr constant [8 x i8] c"\80bc\9Bef\B6h"

declare i1 @runtime.stringEqual(ptr, i32, ptr, i32, ptr)

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

define i1 @main.stringCompareConst8(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @str8, i32 8, ptr undef)
  ret i1 %0
}
