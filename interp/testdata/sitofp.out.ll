target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.f32signed = local_unnamed_addr global i32 1568669697
@main.f32unsigned = local_unnamed_addr global i32 1593835521
@main.f64signed = local_unnamed_addr global i64 4877398396710682624
@main.f64unsigned = local_unnamed_addr global i64 4890909195592794112

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
