target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.x = global i64 0
@main.s = local_unnamed_addr global { [7 x i8], ptr } { [7 x i8] zeroinitializer, ptr @main.x }
@main.n = local_unnamed_addr global i16 0

define void @runtime.initAll() unnamed_addr {
entry:
  %n = load i16, ptr getelementptr inbounds (i8, ptr @main.s, i32 7), align 2
  store i16 %n, ptr @main.n, align 2
  ret void
}
