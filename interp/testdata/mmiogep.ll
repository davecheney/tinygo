target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

%main.Periph = type { i32, i32 }

@main.p = global ptr null
@main.q = global ptr getelementptr inbounds (%main.Periph, ptr inttoptr (i64 1073741824 to ptr), i64 1, i32 1)

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init()
  ret void
}

define internal void @main.init() unnamed_addr {
entry:
  ; Take the address of a field at a fixed address, like memory-mapped I/O.
  store ptr getelementptr inbounds (%main.Periph, ptr inttoptr (i64 1073741824 to ptr), i32 0, i32 1), ptr @main.p
  ret void
}
