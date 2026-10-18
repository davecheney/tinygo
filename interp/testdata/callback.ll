target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.x = global i64 0
@main.y = global i64 0
@main.z = global i64 0

declare void @callSetX()

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

; Program entry point. It runs runtime.initAll, so it cannot be called back
; during init and @main.z can still be set at compile time.
define void @main() {
entry:
  call void @runtime.initAll()
  %z = load i64, ptr @main.z
  call void @runtime.printint64(i64 %z)
  ret void
}

declare void @runtime.printint64(i64)

; Exported function that @callSetX can call back.
define void @setX() {
entry:
  store i64 5, ptr @main.x
  ret void
}

define internal void @main.init(ptr %context) unnamed_addr {
entry:
  call void @callSetX()
  %x = load i64, ptr @main.x
  store i64 %x, ptr @main.y
  %x.next = add i64 %x, 1
  store i64 %x.next, ptr @main.x
  store i64 3, ptr @main.z
  ret void
}
