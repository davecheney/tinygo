package transform

// This file implements a pass that replaces ptrtoint instructions with
// ptrtoaddr (LLVM 22+) when the resulting integer is only ever inspected
// locally, for example in an alignment check or a pointer comparison.
//
// A ptrtoint exposes the provenance of the pointer: any later inttoptr in the
// program may pick it up, so LLVM must assume the pointed-to memory can be
// accessed by unknown code. A ptrtoaddr only captures the address, which lets
// alias analysis, function-attrs, GVN and DSE treat the object as not escaped.
//
// This is only valid when no inttoptr can (legally) observe the integer. Go
// code frequently round-trips through uintptr (unsafe.Pointer(uintptr(p)+n)),
// passes uintptrs to syscalls, and the runtime stores uintptrs that it later
// turns back into pointers. Therefore a ptrtoint is only converted if every
// transitive use of the integer (through integer arithmetic, casts, phi,
// select and freeze) ends in a comparison, a branch/switch condition, or a
// getelementptr index. A store, call argument, return, inttoptr or any other
// unknown use keeps the ptrtoint, as does anything that leaves the function.

import (
	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"tinygo.org/x/go-llvm"
)

// OptimizePtrToAddr replaces ptrtoint instructions whose results are only
// used locally as an address with ptrtoaddr. It is a no-op before LLVM 22.
// It returns the number of replaced instructions.
func OptimizePtrToAddr(mod llvm.Module) int {
	if llvmutil.Version() < 22 {
		return 0
	}
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	builder := mod.Context().NewBuilder()
	defer builder.Dispose()

	var worklist []llvm.Value
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
			for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
				if inst.InstructionOpcode() != llvm.PtrToInt {
					continue
				}
				if !ptrToIntIsAddressOnly(inst, targetData) {
					continue
				}
				worklist = append(worklist, inst)
			}
		}
	}
	for _, inst := range worklist {
		// LLVM 22 InstCombine folds icmp(ptrtoint p, ptrtoint q) and
		// icmp(ptrtoint p, 0) into pointer comparisons, but doesn't do so
		// for ptrtoaddr yet. Do that fold here so it isn't lost.
		foldPointerCompares(builder, inst)
		if len(getUses(inst)) == 0 {
			inst.EraseFromParentAsInstruction()
			continue
		}
		name := inst.Name()
		inst.SetName("")
		builder.SetInsertPointBefore(inst)
		addr := builder.CreateCast(inst.Operand(0), llvmutil.PtrToAddr, inst.Type(), name)
		inst.ReplaceAllUsesWith(addr)
		inst.EraseFromParentAsInstruction()
	}
	return len(worklist)
}

// foldPointerCompares replaces integer comparisons of a ptrtoint result with
// zero or with another ptrtoint/ptrtoaddr of the same pointer type by a direct
// pointer comparison (icmp on pointers compares their addresses).
func foldPointerCompares(builder llvm.Builder, inst llvm.Value) {
	ptr := inst.Operand(0)
	for _, use := range getUses(inst) {
		if use.IsAInstruction().IsNil() || use.InstructionOpcode() != llvm.ICmp {
			continue
		}
		lhs, rhs := use.Operand(0), use.Operand(1)
		other := rhs
		if other == inst {
			other = lhs
		}
		var otherPtr llvm.Value
		switch {
		case other == inst:
			otherPtr = ptr
		case !other.IsAConstantInt().IsNil() && other.IsNull():
			otherPtr = llvm.ConstNull(ptr.Type())
		case !other.IsAInstruction().IsNil() &&
			(other.InstructionOpcode() == llvm.PtrToInt || other.InstructionOpcode() == llvmutil.PtrToAddr) &&
			other.Type() == inst.Type() && other.Operand(0).Type() == ptr.Type():
			otherPtr = other.Operand(0)
		default:
			continue
		}
		lhsPtr, rhsPtr := ptr, otherPtr
		if lhs != inst {
			lhsPtr, rhsPtr = otherPtr, ptr
		}
		builder.SetInsertPointBefore(use)
		cmp := builder.CreateICmp(use.IntPredicate(), lhsPtr, rhsPtr, "")
		name := use.Name()
		use.ReplaceAllUsesWith(cmp)
		use.EraseFromParentAsInstruction()
		cmp.SetName(name)
	}
}

// ptrToIntIsAddressOnly reports whether the given ptrtoint instruction can be
// replaced by a ptrtoaddr: the integer has the index width of the pointer and
// never flows to anything that could turn it back into a usable pointer.
func ptrToIntIsAddressOnly(inst llvm.Value, targetData llvm.TargetData) bool {
	ptrType := inst.Operand(0).Type()
	if ptrType.TypeKind() != llvm.PointerTypeKind || ptrType.PointerAddressSpace() != 0 {
		return false
	}
	// ptrtoaddr requires the result type to be exactly the index width. All
	// TinyGo targets have index width == pointer size in address space 0.
	if inst.Type().TypeKind() != llvm.IntegerTypeKind ||
		uint64(inst.Type().IntTypeWidth()) != targetData.TypeSizeInBits(ptrType) {
		return false
	}
	return intIsAddressOnly(inst, map[llvm.Value]struct{}{})
}

// addressOnlyCallees are runtime functions that only print the address passed
// in their first parameter (emitted for println(ptr) and println(slice)).
var addressOnlyCallees = map[string]bool{
	"runtime.printptr":   true,
	"runtime.printslice": true,
}

// intIsAddressOnly reports whether all transitive uses of the integer value
// only inspect it locally.
func intIsAddressOnly(value llvm.Value, visited map[llvm.Value]struct{}) bool {
	if _, ok := visited[value]; ok {
		return true
	}
	visited[value] = struct{}{}
	for _, use := range getUses(value) {
		if use.IsAInstruction().IsNil() {
			return false
		}
		switch use.InstructionOpcode() {
		case llvm.Add, llvm.Sub, llvm.Mul, llvm.And, llvm.Or, llvm.Xor,
			llvm.Shl, llvm.LShr, llvm.AShr, llvm.UDiv, llvm.SDiv, llvm.URem, llvm.SRem,
			llvm.Trunc, llvm.ZExt, llvm.SExt, llvm.PHI, llvmutil.Freeze:
			if !intIsAddressOnly(use, visited) {
				return false
			}
		case llvm.Select:
			if use.Operand(0) != value || use.Operand(1) == value || use.Operand(2) == value {
				// Used as one of the selected values.
				if !intIsAddressOnly(use, visited) {
					return false
				}
			}
		case llvm.ICmp, llvm.Switch, llvm.Br:
			// Only inspects the address.
		case llvm.Call:
			// The print builtins only format the address as a number.
			// Anything else may turn the integer back into a pointer.
			callee := use.CalledValue()
			if callee.IsAFunction().IsNil() || !addressOnlyCallees[callee.Name()] {
				return false
			}
			if use.Operand(0) != value {
				// Only the first (address) parameter is allowed.
				return false
			}
			for i := 1; i < use.OperandsCount()-1; i++ {
				if use.Operand(i) == value {
					return false
				}
			}
		case llvm.GetElementPtr:
			// Used as an index: the resulting pointer is based on the GEP
			// base pointer, not on this integer.
			if use.Operand(0) == value {
				return false
			}
		default:
			// Store, call, ret, inttoptr, and anything unknown: the integer
			// may be turned back into a pointer somewhere.
			return false
		}
	}
	return true
}
