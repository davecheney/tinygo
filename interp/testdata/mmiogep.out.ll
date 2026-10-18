target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.p = local_unnamed_addr global ptr inttoptr (i64 1073741828 to ptr)
@main.q = local_unnamed_addr global ptr inttoptr (i64 1073741836 to ptr)

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
