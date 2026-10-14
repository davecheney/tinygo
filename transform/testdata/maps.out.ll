target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@answer = constant [6 x i8] c"answer"

declare nonnull ptr @runtime.hashmapMake(i8, i8, i32)

declare void @runtime.hashmapStringSet(ptr nocapture, ptr, i32, ptr nocapture readonly)

declare i1 @runtime.hashmapStringGet(ptr nocapture, ptr, i32, ptr nocapture)

declare nonnull ptr @runtime.hashmapMakeGeneric(i32, i32, i32, ptr, ptr, ptr, ptr, ptr)

declare void @runtime.hashmapGenericSet(ptr nocapture, ptr, ptr nocapture readonly)

define void @testUnused() {
  ret void
}

define i32 @testReadonly() {
  %map = call ptr @runtime.hashmapMake(i8 4, i8 4, i32 0)
  %hashmap.value = alloca i32, align 4
  store i32 42, ptr %hashmap.value, align 4
  call void @runtime.hashmapStringSet(ptr %map, ptr @answer, i32 6, ptr %hashmap.value)
  %hashmap.value2 = alloca i32, align 4
  %commaOk = call i1 @runtime.hashmapStringGet(ptr %map, ptr @answer, i32 6, ptr %hashmap.value2)
  %loadedValue = load i32, ptr %hashmap.value2, align 4
  ret i32 %loadedValue
}

define void @testUnusedGeneric(ptr %key) {
  %map = call ptr @runtime.hashmapMakeGeneric(i32 8, i32 4, i32 8, ptr null, ptr null, ptr null, ptr null, ptr null)
  %hashmap.value = alloca i32, align 4
  store i32 1, ptr %hashmap.value, align 4
  call void @runtime.hashmapGenericSet(ptr %map, ptr %key, ptr %hashmap.value)
  ret void
}

define ptr @testUsed() {
  %1 = call ptr @runtime.hashmapMake(i8 4, i8 4, i32 0)
  ret ptr %1
}
