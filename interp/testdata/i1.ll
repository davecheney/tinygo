target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.sext = global i32 0
@main.slt = global i1 false
@main.sgt = global i1 true
@main.sitofp = global double 0.0
@main.add = global i1 true
@main.br = global i32 0
@main.select = global i32 0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

define internal i1 @id(i1 %x) {
  ret i1 %x
}

define internal void @main.init(ptr %context) unnamed_addr {
entry:
  %t = call i1 @id(i1 true)
  %f = call i1 @id(i1 false)
  %sext = sext i1 %t to i32
  store i32 %sext, ptr @main.sext
  %slt = icmp slt i1 %t, %f
  store i1 %slt, ptr @main.slt
  %sgt = icmp sgt i1 %t, %f
  store i1 %sgt, ptr @main.sgt
  %sitofp = sitofp i1 %t to double
  store double %sitofp, ptr @main.sitofp
  %add = add i1 %t, %t
  store i1 %add, ptr @main.add
  %sel = select i1 %add, i32 1, i32 2
  store i32 %sel, ptr @main.select
  br i1 %add, label %then, label %else

then:
  store i32 1, ptr @main.br
  ret void

else:
  store i32 2, ptr @main.br
  ret void
}
