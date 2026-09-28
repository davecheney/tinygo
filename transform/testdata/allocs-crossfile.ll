; An allocation in caller.go that escapes through a store in callee.go. The
; reported escape position must name callee.go: a bare line number would be
; read against caller.go.
target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@sink = global ptr null

declare ptr @runtime.alloc(i32, ptr)

define void @crossFile() !dbg !10 {
  %p = call ptr @runtime.alloc(i32 4, ptr null), !dbg !13
  call void @storeInOtherFile(ptr %p), !dbg !14
  ret void
}

define void @sameFile() !dbg !11 {
  %p = call ptr @runtime.alloc(i32 4, ptr null), !dbg !15
  store ptr %p, ptr @sink, !dbg !16
  ret void
}

define void @storeInOtherFile(ptr %p) !dbg !12 {
  store ptr %p, ptr @sink, !dbg !17
  ret void
}

!llvm.module.flags = !{!0}
!llvm.dbg.cu = !{!3}

!0 = !{i32 2, !"Debug Info Version", i32 3}
!1 = !DIFile(filename: "caller.go", directory: "/src")
!2 = !DIFile(filename: "callee.go", directory: "/src")
!3 = distinct !DICompileUnit(language: DW_LANG_Go, file: !1, producer: "tinygo", isOptimized: true, runtimeVersion: 0, emissionKind: FullDebug)
!4 = !DISubroutineType(types: !5)
!5 = !{null}
!10 = distinct !DISubprogram(name: "crossFile", file: !1, line: 5, type: !4, scopeLine: 5, spFlags: DISPFlagDefinition, unit: !3)
!11 = distinct !DISubprogram(name: "sameFile", file: !1, line: 10, type: !4, scopeLine: 10, spFlags: DISPFlagDefinition, unit: !3)
!12 = distinct !DISubprogram(name: "storeInOtherFile", file: !2, line: 20, type: !4, scopeLine: 20, spFlags: DISPFlagDefinition, unit: !3)
!13 = !DILocation(line: 6, column: 7, scope: !10)
!14 = !DILocation(line: 7, column: 3, scope: !10)
!15 = !DILocation(line: 11, column: 7, scope: !11)
!16 = !DILocation(line: 12, column: 3, scope: !11)
!17 = !DILocation(line: 21, column: 4, scope: !12)
