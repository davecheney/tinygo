target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.a = local_unnamed_addr global float 0x7FF0000020000000
@main.b = local_unnamed_addr global float 0x7FF0000020000000
@main.c = local_unnamed_addr global float 0x7FF0000020000000
@main.d = local_unnamed_addr global { float, i32 } { float 0x7FF0000020000000, i32 1 }
@main.e = local_unnamed_addr global double 0x7FF0000000000001
@main.f = local_unnamed_addr global double 0x7FF0000000000001

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
