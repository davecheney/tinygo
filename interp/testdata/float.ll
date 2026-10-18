target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.a = global float 0x7FF0000020000000
@main.b = global float 0.0
@main.c = global float 0.0
@main.d = global { float, i32 } zeroinitializer
@main.e = global double 0x7FF0000000000001
@main.f = global double 0.0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

define internal void @main.init(ptr %context) unnamed_addr {
entry:
  %a = load float, ptr @main.a
  store float %a, ptr @main.b
  store i32 2139095041, ptr @main.c
  store { float, i32 } { float 0x7FF0000020000000, i32 1 }, ptr @main.d
  %e = load double, ptr @main.e
  store double %e, ptr @main.f
  ret void
}
