target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@byte.target = global [16 x i8] zeroinitializer
@byte.ptr = local_unnamed_addr global ptr getelementptr inbounds nuw (i8, ptr @byte.target, i64 4)
@byte.copy = local_unnamed_addr global ptr null
@start.target = global i8 0
@start.value = local_unnamed_addr global { ptr, [8 x i8] } { ptr @start.target, [8 x i8] zeroinitializer }
@end.target = global i8 0
@end.value = local_unnamed_addr global { [8 x i8], ptr } { [8 x i8] zeroinitializer, ptr @end.target }
@copy.target = global i8 0
@copy.ptr = local_unnamed_addr global ptr @copy.target
@copy.src = local_unnamed_addr global [2 x i8] c"\01\02"
@whole.a = global i8 0
@whole.b = global i8 0
@whole.pair = local_unnamed_addr global [2 x ptr] [ptr @whole.a, ptr @whole.b]
@whole.int = local_unnamed_addr global [2 x ptr] [ptr null, ptr @whole.a]
@whole.near = local_unnamed_addr global { ptr, [8 x i8] } { ptr @whole.a, [8 x i8] c"\01\00\00\00\00\00\00\00" }

define void @runtime.initAll() unnamed_addr {
entry:
  store i8 1, ptr @byte.ptr, align 1
  %val = load ptr, ptr @byte.ptr, align 8
  store ptr %val, ptr @byte.copy, align 8
  store [4 x i8] c"\01\02\03\04", ptr getelementptr inbounds (i8, ptr @start.value, i32 6), align 1
  store [4 x i8] c"\01\02\03\04", ptr getelementptr inbounds (i8, ptr @end.value, i32 6), align 1
  call void @llvm.memcpy.p0.p0.i64(ptr getelementptr inbounds (i8, ptr @copy.ptr, i32 3), ptr @copy.src, i64 2, i1 false)
  ret void
}

declare void @llvm.memcpy.p0.p0.i64(ptr noalias nocapture writeonly, ptr noalias nocapture readonly, i64, i1 immarg) #0

attributes #0 = { nocallback nofree nounwind willreturn memory(argmem: readwrite) }
