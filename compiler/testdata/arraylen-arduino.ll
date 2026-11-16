; ModuleID = 'arraylen.go'
source_filename = "arraylen.go"
target datalayout = "e-P1-p:16:8-i8:8-i16:8-i32:8-i64:8-f32:8-f64:8-n8:16-a:8"
target triple = "avr"

%runtime._interface = type { ptr, ptr }

@"reflect/types.type:pointer:array:65536:basic:uint8" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:array:65536:basic:uint8" }, align 4
@"reflect/types.type:array:65536:basic:uint8" = linkonce_odr constant { i8, i16, ptr, ptr, i16, ptr, ptr } { i8 -41, i16 0, ptr @"reflect/types.type:pointer:array:65536:basic:uint8", ptr @"reflect/types.type:basic:uint8", i16 -1, ptr @"reflect/types.type:slice:basic:uint8", ptr inttoptr (i16 3 to ptr) }, align 4
@"reflect/types.type:basic:uint8" = linkonce_odr constant { i8, ptr } { i8 -56, ptr @"reflect/types.type:pointer:basic:uint8" }, align 4
@"reflect/types.type:pointer:basic:uint8" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:basic:uint8" }, align 4
@"reflect/types.type:slice:basic:uint8" = linkonce_odr constant { i8, i16, ptr, ptr } { i8 22, i16 0, ptr @"reflect/types.type:pointer:slice:basic:uint8", ptr @"reflect/types.type:basic:uint8" }, align 4
@"reflect/types.type:pointer:slice:basic:uint8" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:slice:basic:uint8" }, align 4

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i8 @main.indexBig(ptr dereferenceable_or_null(65536) %p, i16 %i, ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %1 = icmp eq i16 %i, -1
  br i1 %1, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %gep.next
  %2 = getelementptr inbounds i8, ptr %p, i16 %i
  %3 = load i8, ptr %2, align 1
  ret i8 %3

gep.throw:                                        ; preds = %entry
  call addrspace(1) void @runtime.nilPanic(ptr undef) #0
  unreachable

lookup.throw:                                     ; preds = %gep.next
  call addrspace(1) void @runtime.lookupPanic(ptr undef) #0
  unreachable
}

declare void @runtime.nilPanic(ptr) addrspace(1)

declare void @runtime.lookupPanic(ptr) addrspace(1)

; Function Attrs: nounwind
define hidden { ptr, i16, i16 } @main.sliceBig(ptr dereferenceable_or_null(65536) %p, ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %1 = insertvalue { ptr, i16, i16 } undef, ptr %p, 0
  %2 = insertvalue { ptr, i16, i16 } %1, i16 -1, 1
  %3 = insertvalue { ptr, i16, i16 } %2, i16 -1, 2
  ret { ptr, i16, i16 } %3

slice.throw:                                      ; preds = %entry
  call addrspace(1) void @runtime.nilPanic(ptr undef) #0
  unreachable
}

; Function Attrs: nounwind
define hidden ptr @main.sliceToBig(ptr %s.data, i16 %s.len, i16 %s.cap, ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  %.not = icmp eq i16 %s.len, -1
  br i1 %.not, label %slicetoarray.next, label %slicetoarray.throw

slicetoarray.next:                                ; preds = %entry
  ret ptr %s.data

slicetoarray.throw:                               ; preds = %entry
  call addrspace(1) void @runtime.sliceToArrayPointerPanic(ptr undef) #0
  unreachable
}

declare void @runtime.sliceToArrayPointerPanic(ptr) addrspace(1)

; Function Attrs: nounwind
define hidden %runtime._interface @main.boxBig(ptr dereferenceable_or_null(65536) %p, ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  %0 = insertvalue %runtime._interface { ptr @"reflect/types.type:pointer:array:65536:basic:uint8", ptr undef }, ptr %p, 1
  ret %runtime._interface %0
}

attributes #0 = { nounwind }
