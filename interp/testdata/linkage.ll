target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@weak.cfg = weak global i32 1
@weak.x = global i32 0
@weak.y = global i32 0
@common.cfg = common global i32 0
@common.x = global i32 0
@linkonce.cfg = linkonce global i32 3
@linkonce.x = global i32 0
@odr.const = linkonce_odr constant i32 4
@odr.cfg = weak_odr global i32 5
@odr.x = global i32 0
@odr.y = global i32 0
@odr.z = global i32 0
@odr.call = global i32 0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @weak.init(ptr undef)
  call void @common.init(ptr undef)
  call void @linkonce.init(ptr undef)
  call void @odr.init(ptr undef)
  ret void
}

define internal void @weak.init(ptr %context) unnamed_addr {
entry:
  %cfg = load i32, ptr @weak.cfg
  store i32 %cfg, ptr @weak.x
  store i32 7, ptr @weak.cfg
  %y = call i32 @weak.getcfg()
  store i32 %y, ptr @weak.y
  ret void
}

define weak i32 @weak.getcfg() {
entry:
  ret i32 1
}

define internal void @common.init(ptr %context) unnamed_addr {
entry:
  %cfg = load i32, ptr @common.cfg
  store i32 %cfg, ptr @common.x
  ret void
}

define internal void @linkonce.init(ptr %context) unnamed_addr {
entry:
  %cfg = load i32, ptr @linkonce.cfg
  store i32 %cfg, ptr @linkonce.x
  ret void
}

define internal void @odr.init(ptr %context) unnamed_addr {
entry:
  %c = load i32, ptr @odr.const
  store i32 %c, ptr @odr.x
  %cfg = load i32, ptr @odr.cfg
  store i32 %cfg, ptr @odr.y
  store i32 8, ptr @odr.cfg
  %cfg2 = load i32, ptr @odr.cfg
  store i32 %cfg2, ptr @odr.z
  %v = call i32 @odr.get()
  store i32 %v, ptr @odr.call
  ret void
}

define linkonce_odr i32 @odr.get() {
entry:
  ret i32 6
}
