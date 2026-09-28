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
