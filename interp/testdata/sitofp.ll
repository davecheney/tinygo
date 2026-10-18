target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.signed = internal global i64 1152921573326323713
@main.unsigned = internal global i64 -9223371487098961919
@main.f32signed = global float 0.0
@main.f32unsigned = global float 0.0
@main.f64signed = global double 0.0
@main.f64unsigned = global double 0.0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

; The float32 results must be rounded once, not first to float64.
define internal void @main.init(ptr %context) unnamed_addr {
entry:
  %s = load i64, ptr @main.signed
  %u = load i64, ptr @main.unsigned
  %a = sitofp i64 %s to float
  store float %a, ptr @main.f32signed
  %b = uitofp i64 %u to float
  store float %b, ptr @main.f32unsigned
  %c = sitofp i64 %s to double
  store double %c, ptr @main.f64signed
  %d = uitofp i64 %u to double
  store double %d, ptr @main.f64unsigned
  ret void
}
