package transform

import (
	"strconv"

	"tinygo.org/x/go-llvm"
)

func poisonStackAllocs(mod llvm.Module) {
	ctx := mod.Context()
	builder := ctx.NewBuilder()
	defer builder.Dispose()
	kind := ctx.MDKindID("tinygo.stackalloc")
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	uintptrType := ctx.IntType(targetData.PointerSize() * 8)
	memsetName := "llvm.memset.p0.i" + strconv.Itoa(uintptrType.IntTypeWidth())
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		var allocas, returns []llvm.Value
		for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
			for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
				if !inst.IsAAllocaInst().IsNil() && !inst.Metadata(kind).IsNil() {
					allocas = append(allocas, inst)
				}
				if !inst.IsAReturnInst().IsNil() {
					returns = append(returns, inst)
				}
			}
		}
		for _, alloca := range allocas {
			memset := mod.NamedFunction(memsetName)
			if memset.IsNil() {
				fnType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{llvm.PointerType(ctx.Int8Type(), 0), ctx.Int8Type(), uintptrType, ctx.Int1Type()}, false)
				memset = llvm.AddFunction(mod, memsetName, fnType)
			}
			size := llvm.ConstInt(uintptrType, uint64(alloca.AllocatedType().ArrayLength()), false)
			align := ctx.CreateEnumAttribute(llvm.AttributeKindID("align"), uint64(alloca.Alignment()))
			for _, ret := range returns {
				builder.SetInsertPointBefore(ret)
				args := []llvm.Value{alloca, llvm.ConstInt(ctx.Int8Type(), 0xa5, false), size, llvm.ConstInt(ctx.Int1Type(), 1, false)}
				call := builder.CreateCall(memset.GlobalValueType(), memset, args, "")
				call.AddCallSiteAttribute(1, align)
			}
		}
	}
}
