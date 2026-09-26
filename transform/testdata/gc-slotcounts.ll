target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

@runtime.stackChainStart = external global ptr

declare void @runtime.trackPointer(ptr nocapture readonly)

declare noalias nonnull ptr @runtime.alloc(i32, ptr)

; %elem derives from %arr, which is already tracked and already has a slot, so
; the main loop gives it none. Reaching it as a phi input must not give it one
; either: an interior pointer is kept alive by its base, not on its own.
define ptr @loopPhiWithGEPInput(i1 %repeat) {
entry:
  %arr = call ptr @runtime.alloc(i32 32, ptr inttoptr (i32 3 to ptr))
  call void @runtime.trackPointer(ptr %arr)
  %elem = getelementptr [32 x i8], ptr %arr, i32 0, i32 4
  br label %loop

loop:
  %merged = phi ptr [ %elem, %entry ], [ %next, %loop ]
  call void @runtime.trackPointer(ptr %merged)
  %next = call ptr @runtime.alloc(i32 4, ptr inttoptr (i32 3 to ptr))
  br i1 %repeat, label %loop, label %end

end:
  ret ptr %merged
}
