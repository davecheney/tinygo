target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@weak.cfg = weak local_unnamed_addr global i32 1
@weak.x = local_unnamed_addr global i32 0
@weak.y = local_unnamed_addr global i32 0
@common.cfg = common local_unnamed_addr global i32 0
@common.x = local_unnamed_addr global i32 0
@linkonce.cfg = linkonce local_unnamed_addr global i32 3
@linkonce.x = local_unnamed_addr global i32 0
@odr.cfg = weak_odr local_unnamed_addr global i32 5
@odr.x = local_unnamed_addr global i32 4
@odr.y = local_unnamed_addr global i32 5
@odr.z = local_unnamed_addr global i32 0
@odr.call = local_unnamed_addr global i32 6

define void @runtime.initAll() unnamed_addr {
entry:
  %cfg = load i32, ptr @weak.cfg, align 4
  store i32 %cfg, ptr @weak.x, align 4
  store i32 7, ptr @weak.cfg, align 4
  %y = call i32 @weak.getcfg()
  store i32 %y, ptr @weak.y, align 4
  %cfg1 = load i32, ptr @common.cfg, align 4
  store i32 %cfg1, ptr @common.x, align 4
  %cfg2 = load i32, ptr @linkonce.cfg, align 4
  store i32 %cfg2, ptr @linkonce.x, align 4
  store i32 8, ptr @odr.cfg, align 4
  %cfg23 = load i32, ptr @odr.cfg, align 4
  store i32 %cfg23, ptr @odr.z, align 4
  ret void
}

define weak i32 @weak.getcfg() local_unnamed_addr {
entry:
  ret i32 1
}
