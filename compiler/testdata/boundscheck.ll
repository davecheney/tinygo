; ModuleID = 'boundscheck.go'
source_filename = "boundscheck.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime._string = type { ptr, i32 }

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i32 @main.countedLoop(ptr %s.data, i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %for.loop

for.loop:                                         ; preds = %for.body, %entry
  %0 = phi i32 [ 0, %entry ], [ %6, %for.body ]
  %1 = phi i32 [ 0, %entry ], [ %7, %for.body ]
  %2 = icmp slt i32 %1, %s.len
  br i1 %2, label %for.body, label %for.done

for.body:                                         ; preds = %for.loop
  %3 = getelementptr inbounds i8, ptr %s.data, i32 %1
  %4 = load i8, ptr %3, align 1
  %5 = zext i8 %4 to i32
  %6 = add i32 %0, %5
  %7 = add i32 %1, 1
  br label %for.loop

for.done:                                         ; preds = %for.loop
  ret i32 %0
}

; Function Attrs: nounwind
define hidden i32 @main.whileLoop(ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  br label %for.loop

for.loop:                                         ; preds = %for.body, %entry
  %0 = phi i32 [ 0, %entry ], [ %5, %for.body ]
  %1 = icmp slt i32 %0, %s.len
  br i1 %1, label %cond.true, label %for.done

cond.true:                                        ; preds = %for.loop
  %2 = getelementptr inbounds i8, ptr %s.data, i32 %0
  %3 = load i8, ptr %2, align 1
  %4 = icmp eq i8 %3, 32
  br i1 %4, label %for.body, label %for.done

for.body:                                         ; preds = %cond.true
  %5 = add i32 %0, 1
  br label %for.loop

for.done:                                         ; preds = %cond.true, %for.loop
  ret i32 %0
}

; Function Attrs: nounwind
define hidden i32 @main.twoLoops(ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  br label %for.loop

for.loop:                                         ; preds = %for.body, %entry
  %0 = phi i32 [ 0, %entry ], [ %5, %for.body ]
  %1 = icmp slt i32 %0, %s.len
  br i1 %1, label %cond.true, label %for.done

cond.true:                                        ; preds = %for.loop
  %2 = getelementptr inbounds i8, ptr %s.data, i32 %0
  %3 = load i8, ptr %2, align 1
  %4 = icmp eq i8 %3, 32
  br i1 %4, label %for.body, label %for.done

for.body:                                         ; preds = %cond.true
  %5 = add i32 %0, 1
  br label %for.loop

for.done:                                         ; preds = %cond.true, %for.loop
  br label %for.loop1

for.loop1:                                        ; preds = %if.done, %for.done
  %6 = phi i32 [ %0, %for.done ], [ %13, %if.done ]
  %7 = phi i32 [ 0, %for.done ], [ %12, %if.done ]
  %8 = icmp slt i32 %6, %s.len
  br i1 %8, label %for.body2, label %for.done3

for.body2:                                        ; preds = %for.loop1
  %9 = getelementptr inbounds i8, ptr %s.data, i32 %6
  %10 = load i8, ptr %9, align 1
  %.not = icmp eq i8 %10, 32
  br i1 %.not, label %if.done, label %if.then

if.then:                                          ; preds = %for.body2
  %11 = add i32 %7, 1
  br label %if.done

if.done:                                          ; preds = %if.then, %for.body2
  %12 = phi i32 [ %7, %for.body2 ], [ %11, %if.then ]
  %13 = add i32 %6, 1
  br label %for.loop1

for.done3:                                        ; preds = %for.loop1
  ret i32 %7
}

; Function Attrs: nounwind
define hidden i32 @main.negativeStart(ptr %s.data, i32 %s.len, i32 %s.cap, i32 %start, ptr %context) unnamed_addr #1 {
entry:
  br label %for.loop

for.loop:                                         ; preds = %lookup.next, %entry
  %0 = phi i32 [ 0, %entry ], [ %6, %lookup.next ]
  %1 = phi i32 [ %start, %entry ], [ %7, %lookup.next ]
  %2 = icmp slt i32 %1, %s.len
  br i1 %2, label %for.body, label %for.done

for.body:                                         ; preds = %for.loop
  %.not = icmp ult i32 %1, %s.len
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %for.body
  %3 = getelementptr inbounds i8, ptr %s.data, i32 %1
  %4 = load i8, ptr %3, align 1
  %5 = zext i8 %4 to i32
  %6 = add i32 %0, %5
  %7 = add i32 %1, 1
  br label %for.loop

for.done:                                         ; preds = %for.loop
  ret i32 %0

lookup.throw:                                     ; preds = %for.body
  call void @runtime.lookupPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i32 undef
}

declare void @runtime.lookupPanic(ptr) #0

; Function Attrs: nounwind
define hidden %runtime._string @main.stringTail(ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  br label %for.loop

for.loop:                                         ; preds = %for.body, %entry
  %0 = phi i32 [ 0, %entry ], [ %5, %for.body ]
  %1 = icmp slt i32 %0, %s.len
  br i1 %1, label %cond.true, label %for.done

cond.true:                                        ; preds = %for.loop
  %2 = getelementptr inbounds i8, ptr %s.data, i32 %0
  %3 = load i8, ptr %2, align 1
  %4 = icmp eq i8 %3, 32
  br i1 %4, label %for.body, label %for.done

for.body:                                         ; preds = %cond.true
  %5 = add i32 %0, 1
  br label %for.loop

for.done:                                         ; preds = %cond.true, %for.loop
  %6 = getelementptr inbounds i8, ptr %s.data, i32 %0
  %7 = sub i32 %s.len, %0
  %8 = insertvalue %runtime._string undef, ptr %6, 0
  %9 = insertvalue %runtime._string %8, i32 %7, 1
  ret %runtime._string %9
}

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.sliceFrom(ptr %s.data, i32 %s.len, i32 %s.cap, i32 %i, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp slt i32 %i, %s.len
  br i1 %0, label %if.then, label %if.done

if.then:                                          ; preds = %entry
  %slice.lowhigh = icmp ugt i32 %i, %s.len
  br i1 %slice.lowhigh, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %if.then
  %1 = getelementptr inbounds i8, ptr %s.data, i32 %i
  %2 = sub nuw i32 %s.len, %i
  %3 = sub i32 %s.cap, %i
  %4 = insertvalue { ptr, i32, i32 } undef, ptr %1, 0
  %5 = insertvalue { ptr, i32, i32 } %4, i32 %2, 1
  %6 = insertvalue { ptr, i32, i32 } %5, i32 %3, 2
  ret { ptr, i32, i32 } %6

if.done:                                          ; preds = %entry
  ret { ptr, i32, i32 } zeroinitializer

slice.throw:                                      ; preds = %if.then
  call void @runtime.slicePanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret { ptr, i32, i32 } undef
}

declare void @runtime.slicePanic(ptr) #0

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nounwind }
