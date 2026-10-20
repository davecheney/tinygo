target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.x = global i64 0
@main.y = global i64 0
@main.pa1 = local_unnamed_addr global ptr @main.x
@main.pa2 = local_unnamed_addr global ptr @main.x
@main.pa3 = local_unnamed_addr global ptr @main.x
@main.pa4 = local_unnamed_addr global ptr @main.x
@main.pa5 = local_unnamed_addr global ptr @main.x
@main.pair = local_unnamed_addr global [2 x ptr] [ptr @main.x, ptr @main.y]
@main.wholePointer = local_unnamed_addr global ptr @main.x
@main.wholeCopy = local_unnamed_addr global ptr @main.x
@main.wholeStruct = local_unnamed_addr global { ptr, i64 } { ptr @main.x, i64 3 }
@main.firstByte = local_unnamed_addr global i8 0
@main.field = local_unnamed_addr global { i8, i8 } zeroinitializer
@main.escaped = local_unnamed_addr global ptr @"main$alloc"
@main.mixed = local_unnamed_addr global i64 0
@main.array = local_unnamed_addr global [2 x i8] zeroinitializer
@"main$alloc" = internal global [2 x i8] zeroinitializer, align 8

define void @runtime.initAll() unnamed_addr {
entry:
  %b = load i8, ptr @main.pa2, align 1
  store i8 %b, ptr @main.firstByte, align 1
  %b1 = load i8, ptr @main.pa3, align 1
  %f.insertvalue0 = insertvalue { i8, i8 } { i8 0, i8 5 }, i8 %b1, 0
  store { i8, i8 } %f.insertvalue0, ptr @main.field, align 1
  %b2 = load i8, ptr @main.pa4, align 1
  store i8 %b2, ptr @"main$alloc", align 1
  %m = load i64, ptr getelementptr inbounds (i8, ptr @main.pair, i64 4), align 8
  store i64 %m, ptr @main.mixed, align 8
  call void @llvm.memcpy.p0.p0.i64(ptr @main.array, ptr @main.pa5, i64 2, i1 false)
  ret void
}

; Function Attrs: nocallback nofree nounwind willreturn memory(argmem: readwrite)
declare void @llvm.memcpy.p0.p0.i64(ptr noalias nocapture writeonly, ptr noalias nocapture readonly, i64, i1 immarg) #0

attributes #0 = { nocallback nofree nounwind willreturn memory(argmem: readwrite) }
