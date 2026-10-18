target datalayout = "E-m:m-p:32:32-i8:8:32-i16:16:32-i64:64-n32-S64"
target triple = "mips-unknown-linux-gnu"

@main.p = local_unnamed_addr global ptr inttoptr (i32 305419896 to ptr)
@main.q = local_unnamed_addr global i32 4660
@main.r = local_unnamed_addr global ptr inttoptr (i32 4660 to ptr)

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
