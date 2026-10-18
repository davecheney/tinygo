target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.signed = internal global i64 1152921573326323713
@main.unsigned = internal global i64 -9223371487098961919
@main.f32signed = global i32 0
@main.f32unsigned = global i32 0
@main.f64signed = global i64 0
@main.f64unsigned = global i64 0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

; The float32 results must be rounded once, not first to float64.
; The results are stored as bits because LLVM versions print float constants differently.
define internal void @main.init(ptr %context) unnamed_addr {
entry:
  %s = load i64, ptr @main.signed
  %u = load i64, ptr @main.unsigned
  %a = sitofp i64 %s to float
  %a.bits = bitcast float %a to i32
  store i32 %a.bits, ptr @main.f32signed
  %b = uitofp i64 %u to float
  %b.bits = bitcast float %b to i32
  store i32 %b.bits, ptr @main.f32unsigned
  %c = sitofp i64 %s to double
  %c.bits = bitcast double %c to i64
  store i64 %c.bits, ptr @main.f64signed
  %d = uitofp i64 %u to double
  %d.bits = bitcast double %d to i64
  store i64 %d.bits, ptr @main.f64unsigned
  ret void
}
