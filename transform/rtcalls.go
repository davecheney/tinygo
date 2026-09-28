package transform

// This file implements several small optimizations of runtime and reflect
// calls.

import (
	"strconv"
	"strings"

	"tinygo.org/x/go-llvm"
)

// OptimizeStringToBytes transforms runtime.stringToBytes(...) calls into const
// []byte slices whenever possible. This optimizes the following pattern:
//
//	w.Write([]byte("foo"))
//
// where Write does not store to the slice.
func OptimizeStringToBytes(mod llvm.Module) {
	stringToBytes := mod.NamedFunction("runtime.stringToBytes")
	if stringToBytes.IsNil() {
		// nothing to optimize
		return
	}

	for _, call := range getUses(stringToBytes) {
		strptr := call.Operand(0)
		strlen := call.Operand(1)

		// strptr is always constant because strings are always constant.

		var pointerUses []llvm.Value
		canConvertPointer := true
		for _, use := range getUses(call) {
			if use.IsAExtractValueInst().IsNil() {
				// Expected an extractvalue, but this is something else.
				canConvertPointer = false
				break
			}
			switch use.Type().TypeKind() {
			case llvm.IntegerTypeKind:
				// A length (len or cap). Propagate the length value.
				// This can always be done because the byte slice is always the
				// same length as the original string.
				use.ReplaceAllUsesWith(strlen)
				use.EraseFromParentAsInstruction()
			case llvm.PointerTypeKind:
				// The string pointer itself.
				if !isReadOnly(use) {
					// There is a store to the byte slice. This means that none
					// of the pointer uses can't be propagated.
					canConvertPointer = false
					break
				}
				// It may be that the pointer value can be propagated, if all of
				// the pointer uses are readonly.
				pointerUses = append(pointerUses, use)
			default:
				// should not happen
				panic("unknown return type of runtime.stringToBytes: " + use.Type().String())
			}
		}
		if canConvertPointer {
			// All pointer uses are readonly, so they can be converted.
			for _, use := range pointerUses {
				use.ReplaceAllUsesWith(strptr)
				use.EraseFromParentAsInstruction()
			}

			// Call to runtime.stringToBytes can be eliminated: both the input
			// and the output is constant.
			call.EraseFromParentAsInstruction()
		}
	}
}

// OptimizeStringEqual optimizes runtime.stringEqual(...) calls where at least
// one side has a constant length:
//
//   - If one of the lengths is zero, the call is replaced with a length
//     comparison. This converts str == "" into len(str) == 0.
//   - If both lengths are constant and differ, the result is false.
//   - If one side is a constant string of length C with 0 < C <= inlineLimit,
//     the call is replaced with an inline length check followed (only when
//     the lengths match) by a few unaligned integer loads of the other string
//     that are compared against integer constants built from the literal.
//
// An inlineLimit of zero disables the inline expansion. If fastUnaligned is
// false, the target cannot do unaligned loads efficiently (the backend would
// split them into byte loads, shifts and ors), so the expansion compares one
// byte at a time instead.
func OptimizeStringEqual(mod llvm.Module, inlineLimit int, fastUnaligned bool) {
	stringEqual := mod.NamedFunction("runtime.stringEqual")
	if stringEqual.IsNil() {
		// nothing to optimize
		return
	}

	ctx := mod.Context()
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	bigEndian := targetData.ByteOrder() == llvm.BigEndian
	// Use the widest native integer type (up to 64 bits) for the loads.
	maxChunk := 1
	for _, width := range []int{16, 32, 64} {
		if nativeIntWidth(mod.DataLayout(), width) {
			maxChunk = width / 8
		}
	}
	if !fastUnaligned {
		maxChunk = 1
	}

	// The helper functions are built with a separate builder, so that they
	// don't get the debug location of a call site.
	helperBuilder := ctx.NewBuilder()
	defer helperBuilder.Dispose()
	expander := &stringEqualExpander{
		mod:           mod,
		builder:       builder,
		helperBuilder: helperBuilder,
		maxChunk:      maxChunk,
		bigEndian:     bigEndian,
		helpers:       map[string]llvm.Value{},
	}
	for _, call := range getUses(stringEqual) {
		str1ptr := call.Operand(0)
		str1len := call.Operand(1)
		str2ptr := call.Operand(2)
		str2len := call.Operand(3)

		zero := llvm.ConstInt(str1len.Type(), 0, false)
		if str1len == zero || str2len == zero {
			builder.SetInsertPointBefore(call)
			icmp := builder.CreateICmp(llvm.IntEQ, str1len, str2len, "")
			call.ReplaceAllUsesWith(icmp)
			call.EraseFromParentAsInstruction()
			continue
		}

		if inlineLimit <= 0 {
			continue
		}

		// Both lengths constant but different: the strings can't be equal.
		if !str1len.IsAConstantInt().IsNil() && !str2len.IsAConstantInt().IsNil() &&
			str1len.ZExtValue() != str2len.ZExtValue() {
			call.ReplaceAllUsesWith(llvm.ConstInt(call.Type(), 0, false))
			call.EraseFromParentAsInstruction()
			continue
		}

		// Find the side that is a constant string. Prefer the second operand,
		// which is where the compiler puts the literal in s == "lit".
		lit, litOK := constStringBytes(str2ptr, str2len, targetData)
		varptr, varlen := str1ptr, str1len
		if !litOK {
			lit, litOK = constStringBytes(str1ptr, str1len, targetData)
			varptr, varlen = str2ptr, str2len
		}
		if !litOK {
			continue
		}

		// If the other side is also a constant string, fold the comparison.
		if other, ok := constStringBytes(varptr, varlen, targetData); ok {
			result := uint64(0)
			if string(other) == string(lit) {
				result = 1
			}
			call.ReplaceAllUsesWith(llvm.ConstInt(call.Type(), result, false))
			call.EraseFromParentAsInstruction()
			continue
		}

		if len(lit) > inlineLimit {
			// Too long: keep the call. (Guarding the call with an inline
			// length check was measured to increase code size at -opt=z,
			// and at other levels LLVM inlines runtime.stringEqual anyway.)
			continue
		}
		expander.expand(call, varptr, varlen, lit)
	}
}

// stringEqualExpander replaces runtime.stringEqual calls against constant
// strings with calls to small alwaysinline helper functions, one per constant
// string:
//
//	define internal i1 @"runtime.stringEqual:abc"(ptr %s.data, i32 %s.len) alwaysinline {
//	entry:
//	  %len = icmp eq i32 %s.len, 3
//	  br i1 %len, label %compare, label %exit
//	compare:
//	  %a = load i16, ptr %s.data, align 1           ; bytes 0-1
//	  %b = load i16, ptr (%s.data + 1), align 1     ; bytes 1-2 (overlapping)
//	  %x = or (xor %a, 0x6261), (xor %b, 0x6362)    ; little endian
//	  %bytes = icmp eq i16 %x, 0
//	  br label %exit
//	exit:
//	  %result = phi i1 [ false, %entry ], [ %bytes, %compare ]
//	  ret i1 %result
//	}
//
// The loads are only executed when the lengths match, so the other string is
// never read out of bounds. The helpers are inlined by the LLVM inliner, which
// (unlike this pass) correctly splits basic blocks with debug records.
type stringEqualExpander struct {
	mod           llvm.Module
	builder       llvm.Builder // for the call sites
	helperBuilder llvm.Builder // for the helper functions (no debug location)
	maxChunk      int
	bigEndian     bool
	helpers       map[string]llvm.Value // key: length type + constant string
}

func (e *stringEqualExpander) expand(call, varptr, varlen llvm.Value, lit []byte) {
	lengthType := varlen.Type()
	key := lengthType.String() + ":" + string(lit)
	helper, ok := e.helpers[key]
	if !ok {
		helper = e.createHelper(lit, varptr.Type(), lengthType)
		e.helpers[key] = helper
	}
	e.builder.SetInsertPointBefore(call) // also sets the debug location
	result := e.builder.CreateCall(helper.GlobalValueType(), helper, []llvm.Value{varptr, varlen}, "")
	call.ReplaceAllUsesWith(result)
	call.EraseFromParentAsInstruction()
}

func (e *stringEqualExpander) createHelper(lit []byte, ptrType, lengthType llvm.Type) llvm.Value {
	ctx := e.mod.Context()
	i1 := ctx.Int1Type()
	i8 := ctx.Int8Type()
	fnType := llvm.FunctionType(i1, []llvm.Type{ptrType, lengthType}, false)
	fn := llvm.AddFunction(e.mod, "runtime.stringEqual:const", fnType)
	fn.SetLinkage(llvm.InternalLinkage)
	fn.SetUnnamedAddr(true)
	for _, attr := range []string{"alwaysinline", "nounwind", "willreturn"} {
		fn.AddFunctionAttr(ctx.CreateEnumAttribute(llvm.AttributeKindID(attr), 0))
	}
	ptr := fn.Param(0)
	ptr.SetName("s.data")
	length := fn.Param(1)
	length.SetName("s.len")

	entry := ctx.AddBasicBlock(fn, "entry")
	compare := ctx.AddBasicBlock(fn, "compare")
	exit := ctx.AddBasicBlock(fn, "exit")
	b := e.helperBuilder

	// Length check.
	b.SetInsertPointAtEnd(entry)
	lenEq := b.CreateICmp(llvm.IntEQ, length, llvm.ConstInt(lengthType, uint64(len(lit)), false), "len")
	b.CreateCondBr(lenEq, compare, exit)

	// Compare the contents with a few (possibly overlapping) loads of the
	// largest power-of-two size that fits.
	size := e.maxChunk
	for size > len(lit) {
		size /= 2
	}
	chunkType := ctx.IntType(size * 8)
	falseValue := llvm.ConstInt(i1, 0, false)
	incomingValues := []llvm.Value{falseValue}
	incomingBlocks := []llvm.BasicBlock{entry}
	var acc llvm.Value
	block := compare
	b.SetInsertPointAtEnd(block)
	for offset := 0; offset < len(lit); offset += size {
		if offset+size > len(lit) {
			// Overlap with the previous chunk to cover the tail.
			offset = len(lit) - size
		}
		chunkPtr := ptr
		if offset != 0 {
			chunkPtr = b.CreateInBoundsGEP(i8, ptr, []llvm.Value{llvm.ConstInt(lengthType, uint64(offset), false)}, "")
		}
		load := b.CreateLoad(chunkType, chunkPtr, "")
		load.SetAlignment(1)
		var bits uint64
		for i := 0; i < size; i++ {
			c := uint64(lit[offset+i])
			if e.bigEndian {
				bits |= c << (8 * (size - 1 - i))
			} else {
				bits |= c << (8 * i)
			}
		}
		if size == 1 && offset+1 < len(lit) {
			// Byte loads: exit early on the first mismatch, like the
			// unrolled loop in runtime.stringEqual. Or-ing all bytes
			// together would need more instructions and registers.
			next := ctx.AddBasicBlock(fn, "compare")
			next.MoveBefore(exit)
			eq := b.CreateICmp(llvm.IntEQ, load, llvm.ConstInt(chunkType, bits, false), "")
			b.CreateCondBr(eq, next, exit)
			incomingValues = append(incomingValues, falseValue)
			incomingBlocks = append(incomingBlocks, block)
			block = next
			b.SetInsertPointAtEnd(block)
			continue
		}
		diff := b.CreateXor(load, llvm.ConstInt(chunkType, bits, false), "")
		if acc.IsNil() {
			acc = diff
		} else {
			acc = b.CreateOr(acc, diff, "")
		}
	}
	bytesEq := b.CreateICmp(llvm.IntEQ, acc, llvm.ConstInt(chunkType, 0, false), "bytes")
	b.CreateBr(exit)
	incomingValues = append(incomingValues, bytesEq)
	incomingBlocks = append(incomingBlocks, block)

	// Merge the results.
	b.SetInsertPointAtEnd(exit)
	phi := b.CreatePHI(i1, "result")
	phi.AddIncoming(incomingValues, incomingBlocks)
	b.CreateRet(phi)
	return fn
}

// constStringBytes returns the contents of the constant string with the given
// pointer and length, if both are constant and the pointer points into a
// constant global with a known initializer.
func constStringBytes(ptr, length llvm.Value, targetData llvm.TargetData) ([]byte, bool) {
	if length.IsAConstantInt().IsNil() {
		return nil, false
	}
	n := length.ZExtValue()
	offset := uint64(0)
	for !ptr.IsAConstantExpr().IsNil() && ptr.Opcode() == llvm.GetElementPtr {
		// Only handle constant GEPs with an integer offset.
		off, ok := constGEPOffset(ptr, targetData)
		if !ok {
			return nil, false
		}
		offset += off
		ptr = ptr.Operand(0)
	}
	global := ptr.IsAGlobalVariable()
	if global.IsNil() || !global.IsGlobalConstant() || global.IsDeclaration() {
		return nil, false
	}
	switch global.Linkage() {
	case llvm.InternalLinkage, llvm.PrivateLinkage, llvm.ExternalLinkage:
	default:
		// The initializer might be replaced at link time.
		return nil, false
	}
	init := global.Initializer()
	initType := init.Type()
	if initType.TypeKind() != llvm.ArrayTypeKind || initType.ElementType() != initType.Context().Int8Type() {
		return nil, false
	}
	var data []byte
	switch {
	case !init.IsAConstantAggregateZero().IsNil():
		data = make([]byte, initType.ArrayLength())
	case !init.IsAConstantArray().IsNil(), !init.IsAUndefValue().IsNil():
		// Not plain bytes (or undef/poison, which also matches IsAUndefValue).
		return nil, false
	default:
		// The only remaining kind of [N x i8] constant is a
		// ConstantDataArray.
		data = []byte(init.ConstGetAsString())
	}
	if offset > uint64(len(data)) || n > uint64(len(data))-offset {
		return nil, false
	}
	return data[offset : offset+n], true
}

// constGEPOffset returns the byte offset of a constant getelementptr
// expression relative to its base pointer, for the common forms emitted for
// string data: getelementptr i8, ptr @g, iN k and
// getelementptr [N x i8], ptr @g, iN 0, iN k.
func constGEPOffset(gep llvm.Value, targetData llvm.TargetData) (uint64, bool) {
	elemType := gep.GEPSourceElementType()
	numIndices := gep.OperandsCount() - 1
	var indices []uint64
	for i := 1; i <= numIndices; i++ {
		index := gep.Operand(i)
		if index.IsAConstantInt().IsNil() {
			return 0, false
		}
		v := index.SExtValue()
		if v < 0 {
			return 0, false
		}
		indices = append(indices, uint64(v))
	}
	offset := indices[0] * targetData.TypeAllocSize(elemType)
	for _, index := range indices[1:] {
		if elemType.TypeKind() != llvm.ArrayTypeKind {
			return 0, false
		}
		elemType = elemType.ElementType()
		offset += index * targetData.TypeAllocSize(elemType)
	}
	return offset, true
}

// nativeIntWidth reports whether the given integer width (in bits) is listed
// as a native integer width ("n32:64" etc) in the data layout string.
func nativeIntWidth(dataLayout string, width int) bool {
	for _, spec := range strings.Split(dataLayout, "-") {
		if !strings.HasPrefix(spec, "n") {
			continue
		}
		for _, w := range strings.Split(spec[1:], ":") {
			if w == strconv.Itoa(width) {
				return true
			}
		}
	}
	return false
}
