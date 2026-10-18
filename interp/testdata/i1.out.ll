target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.sext = local_unnamed_addr global i32 -1
@main.slt = local_unnamed_addr global i1 true
@main.sgt = local_unnamed_addr global i1 false
@main.sitofp = local_unnamed_addr global double -1.000000e+00
@main.add = local_unnamed_addr global i1 false
@main.br = local_unnamed_addr global i32 2
@main.select = local_unnamed_addr global i32 2

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
