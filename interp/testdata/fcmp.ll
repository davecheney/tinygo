target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.less = internal global double 1.5
@main.more = internal global double 2.5
@main.nan = internal global double 0x7FF8000000000000
@main.lessResult = global [16 x i8] zeroinitializer
@main.equalResult = global [16 x i8] zeroinitializer
@main.greaterResult = global [16 x i8] zeroinitializer
@main.nanResult = global [16 x i8] zeroinitializer

define void @runtime.initAll() {
  call void @main.init()
  ret void
}

; Each result array holds the fcmp predicates in LLVM order, from false to true.
define internal void @main.init() {
  %less = load double, ptr @main.less
  %more = load double, ptr @main.more
  %nan = load double, ptr @main.nan
  call void @main.fcmp(double %less, double %more, ptr @main.lessResult)
  call void @main.fcmp(double %more, double %more, ptr @main.equalResult)
  call void @main.fcmp(double %more, double %less, ptr @main.greaterResult)
  call void @main.fcmp(double %nan, double %less, ptr @main.nanResult)
  ret void
}

define internal void @main.fcmp(double %a, double %b, ptr %out) {
  %false = fcmp false double %a, %b
  %false.zext = zext i1 %false to i8
  %false.ptr = getelementptr inbounds i8, ptr %out, i32 0
  store i8 %false.zext, ptr %false.ptr
  %oeq = fcmp oeq double %a, %b
  %oeq.zext = zext i1 %oeq to i8
  %oeq.ptr = getelementptr inbounds i8, ptr %out, i32 1
  store i8 %oeq.zext, ptr %oeq.ptr
  %ogt = fcmp ogt double %a, %b
  %ogt.zext = zext i1 %ogt to i8
  %ogt.ptr = getelementptr inbounds i8, ptr %out, i32 2
  store i8 %ogt.zext, ptr %ogt.ptr
  %oge = fcmp oge double %a, %b
  %oge.zext = zext i1 %oge to i8
  %oge.ptr = getelementptr inbounds i8, ptr %out, i32 3
  store i8 %oge.zext, ptr %oge.ptr
  %olt = fcmp olt double %a, %b
  %olt.zext = zext i1 %olt to i8
  %olt.ptr = getelementptr inbounds i8, ptr %out, i32 4
  store i8 %olt.zext, ptr %olt.ptr
  %ole = fcmp ole double %a, %b
  %ole.zext = zext i1 %ole to i8
  %ole.ptr = getelementptr inbounds i8, ptr %out, i32 5
  store i8 %ole.zext, ptr %ole.ptr
  %one = fcmp one double %a, %b
  %one.zext = zext i1 %one to i8
  %one.ptr = getelementptr inbounds i8, ptr %out, i32 6
  store i8 %one.zext, ptr %one.ptr
  %ord = fcmp ord double %a, %b
  %ord.zext = zext i1 %ord to i8
  %ord.ptr = getelementptr inbounds i8, ptr %out, i32 7
  store i8 %ord.zext, ptr %ord.ptr
  %ueq = fcmp ueq double %a, %b
  %ueq.zext = zext i1 %ueq to i8
  %ueq.ptr = getelementptr inbounds i8, ptr %out, i32 8
  store i8 %ueq.zext, ptr %ueq.ptr
  %ugt = fcmp ugt double %a, %b
  %ugt.zext = zext i1 %ugt to i8
  %ugt.ptr = getelementptr inbounds i8, ptr %out, i32 9
  store i8 %ugt.zext, ptr %ugt.ptr
  %uge = fcmp uge double %a, %b
  %uge.zext = zext i1 %uge to i8
  %uge.ptr = getelementptr inbounds i8, ptr %out, i32 10
  store i8 %uge.zext, ptr %uge.ptr
  %ult = fcmp ult double %a, %b
  %ult.zext = zext i1 %ult to i8
  %ult.ptr = getelementptr inbounds i8, ptr %out, i32 11
  store i8 %ult.zext, ptr %ult.ptr
  %ule = fcmp ule double %a, %b
  %ule.zext = zext i1 %ule to i8
  %ule.ptr = getelementptr inbounds i8, ptr %out, i32 12
  store i8 %ule.zext, ptr %ule.ptr
  %une = fcmp une double %a, %b
  %une.zext = zext i1 %une to i8
  %une.ptr = getelementptr inbounds i8, ptr %out, i32 13
  store i8 %une.zext, ptr %une.ptr
  %uno = fcmp uno double %a, %b
  %uno.zext = zext i1 %uno to i8
  %uno.ptr = getelementptr inbounds i8, ptr %out, i32 14
  store i8 %uno.zext, ptr %uno.ptr
  %true = fcmp true double %a, %b
  %true.zext = zext i1 %true to i8
  %true.ptr = getelementptr inbounds i8, ptr %out, i32 15
  store i8 %true.zext, ptr %true.ptr
  ret void
}
