target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.x = local_unnamed_addr global i64 0
@main.y = local_unnamed_addr global i64 0
@main.z = local_unnamed_addr global i64 3

declare void @callSetX() local_unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call void @callSetX()
  %x = load i64, ptr @main.x, align 8
  store i64 %x, ptr @main.y, align 8
  %x.next = add i64 %x, 1
  store i64 %x.next, ptr @main.x, align 8
  ret void
}

define void @main() local_unnamed_addr {
entry:
  call void @runtime.initAll()
  %z = load i64, ptr @main.z, align 8
  call void @runtime.printint64(i64 %z)
  ret void
}

declare void @runtime.printint64(i64) local_unnamed_addr

define void @setX() local_unnamed_addr {
entry:
  store i64 5, ptr @main.x, align 8
  ret void
}
