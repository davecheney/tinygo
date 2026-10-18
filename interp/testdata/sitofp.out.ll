target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.f32signed = local_unnamed_addr global float 0x43B0000020000000
@main.f32unsigned = local_unnamed_addr global float 0x43E0000020000000
@main.f64signed = local_unnamed_addr global double 0x43B0000010000000
@main.f64unsigned = local_unnamed_addr global double 0x43E0000010000000

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
