target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.lessResult = local_unnamed_addr global [16 x i8] c"\00\00\00\00\01\01\01\01\00\00\00\01\01\01\00\01"
@main.equalResult = local_unnamed_addr global [16 x i8] c"\00\01\00\01\00\01\00\01\01\00\01\00\01\00\00\01"
@main.greaterResult = local_unnamed_addr global [16 x i8] c"\00\00\01\01\00\00\01\01\00\01\01\00\00\01\00\01"
@main.nanResult = local_unnamed_addr global [16 x i8] c"\00\00\00\00\00\00\00\00\01\01\01\01\01\01\01\01"

define void @runtime.initAll() local_unnamed_addr {
  ret void
}
