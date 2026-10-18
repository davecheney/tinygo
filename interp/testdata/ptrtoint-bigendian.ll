target datalayout = "E-m:m-p:32:32-i8:8:32-i16:16:32-i64:64-n32-S64"
target triple = "mips-unknown-linux-gnu"

@main.p = global ptr null
@main.q = global i32 0
@main.r = global ptr null

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

define internal void @main.init(ptr %context) unnamed_addr {
entry:
  store ptr inttoptr (i64 305419896 to ptr), ptr @main.p
  store i32 ptrtoint (ptr inttoptr (i16 4660 to ptr) to i32), ptr @main.q
  store ptr inttoptr (i16 4660 to ptr), ptr @main.r
  ret void
}
