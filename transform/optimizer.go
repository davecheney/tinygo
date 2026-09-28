package transform

import (
	"errors"
	"fmt"
	"go/token"
	"os"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/compiler/ircheck"
	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"tinygo.org/x/go-llvm"
)

// OptimizePackage runs optimization passes over the LLVM module for the given
// Go package.
func OptimizePackage(mod llvm.Module, config *compileopts.Config) {
	_, speedLevel, _ := config.OptLevel()

	// Run TinyGo-specific optimization passes.
	if speedLevel > 0 {
		OptimizeMaps(mod)
	}
}

// stringEqualInlineLimit returns the maximum length of a constant string for
// which a comparison is expanded inline instead of calling
// runtime.stringEqual. The limits were chosen by measuring code size on a few
// programs on Cortex-M, RISC-V, WebAssembly and x86-64:
//
//   - At -opt=z, a call to runtime.stringEqual is only a few instructions, so
//     only expand comparisons that need a single native word load.
//   - At -opt=s, expanding up to 8 bytes (two loads on 32-bit targets) is
//     smaller than the inlined and unrolled byte loop LLVM otherwise produces.
//   - When optimizing for speed, expand up to 16 bytes.
//   - Targets without fast unaligned loads compare byte by byte, which costs
//     about as much as a call already at 2-3 bytes.
func stringEqualInlineLimit(mod llvm.Module, config *compileopts.Config, fastUnaligned bool) int {
	_, _, sizeLevel := config.OptLevel()
	if sizeLevel == 0 {
		return 16
	}
	if !fastUnaligned {
		return 2
	}
	if sizeLevel == 1 {
		return 8
	}
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	return targetData.PointerSize()
}

// hasFastUnalignedAccess returns whether the target can load a misaligned
// 16/32/64-bit integer with a single instruction (or at most two, like
// lwl/lwr on MIPS), instead of the backend splitting the load into bytes.
func hasFastUnalignedAccess(config *compileopts.Config) bool {
	features := config.Features()
	if strings.Contains(features, "+strict-align") {
		return false
	}
	arch, _, _ := strings.Cut(config.Triple(), "-")
	switch {
	case arch == "x86_64", arch == "i386", arch == "i686", arch == "aarch64", arch == "arm64",
		arch == "wasm32", arch == "wasm64":
		return true
	case strings.HasPrefix(arch, "thumbv7"), strings.HasPrefix(arch, "thumbv8"),
		strings.HasPrefix(arch, "armv6"), strings.HasPrefix(arch, "armv7"), strings.HasPrefix(arch, "armv8"):
		// ARMv6 and later (except ARMv6-M, which is thumbv6m) support
		// unaligned word and halfword loads.
		return true
	case strings.HasPrefix(arch, "riscv"):
		return strings.Contains(features, "+unaligned-scalar-mem") || strings.Contains(features, "+fast-unaligned-access")
	case strings.HasPrefix(arch, "mips"):
		// lwl/lwr.
		return true
	}
	return false
}

// Optimize runs a number of optimization and transformation passes over the
// given module. Some passes are specific to TinyGo, others are generic LLVM
// passes.
//
// Please note that some optimizations are not optional, thus Optimize must
// always be run before emitting machine code.
func Optimize(mod llvm.Module, config *compileopts.Config) []error {
	optLevel, speedLevel, _ := config.OptLevel()

	// Make sure these functions are kept in tact during TinyGo transformation passes.
	for _, name := range functionsUsedInTransforms {
		fn := mod.NamedFunction(name)
		if fn.IsNil() {
			panic(fmt.Errorf("missing core function %q", name))
		}
		fn.SetLinkage(llvm.ExternalLinkage)
	}

	// run a check of all of our code
	if config.VerifyIR() {
		errs := ircheck.Module(mod)
		if errs != nil {
			return errs
		}
	}

	if speedLevel > 0 {
		// Run some preparatory passes for the Go optimizer.
		po := llvm.NewPassBuilderOptions()
		defer po.Dispose()
		optPasses := "globaldce,globalopt,ipsccp,instcombine<no-verify-fixpoint>,adce,function-attrs"
		if llvmutil.Version() < 18 {
			// LLVM 17 doesn't have the no-verify-fixpoint flag.
			optPasses = "globaldce,globalopt,ipsccp,instcombine,adce,function-attrs"
		}
		blockGlobalAllocPromotion(mod)
		err := mod.RunPasses(optPasses, llvm.TargetMachine{}, po)
		removeGlobalAllocPromotionMarker(mod)
		if err != nil {
			return []error{fmt.Errorf("could not build pass pipeline: %w", err)}
		}

		// Run TinyGo-specific optimization passes.
		OptimizeStringToBytes(mod)
		maxStackSize := config.MaxStackAlloc()
		OptimizeAllocs(mod, nil, maxStackSize, nil)
		err = LowerInterfaces(mod, config)
		if err != nil {
			return []error{err}
		}

		errs := LowerInterrupts(mod)
		if len(errs) > 0 {
			return errs
		}

		// After interfaces are lowered, there are many more opportunities for
		// interprocedural optimizations. To get them to work, function
		// attributes have to be updated first.
		blockGlobalAllocPromotion(mod)
		err = mod.RunPasses(optPasses, llvm.TargetMachine{}, po)
		removeGlobalAllocPromotionMarker(mod)
		if err != nil {
			return []error{fmt.Errorf("could not build pass pipeline: %w", err)}
		}

		// Run TinyGo-specific interprocedural optimizations.
		if config.Options.PrintAllocs != nil && config.Options.PrintAllocsCover {
			// The go coverage tool expects this header before any blocks.
			fmt.Fprintln(os.Stderr, "mode: set")
		}
		OptimizeAllocs(mod, config.Options.PrintAllocs, maxStackSize,
			func(pos token.Position, reason string) {
				var line string
				if config.Options.PrintAllocsCover {
					line = FormatAllocCover(pos)
				} else {
					line = FormatAllocReason(pos, reason)
				}
				if line != "" {
					fmt.Fprintln(os.Stderr, line)
				}
			},
		)
		OptimizeStringToBytes(mod)
		fastUnaligned := hasFastUnalignedAccess(config)
		OptimizeStringEqual(mod, stringEqualInlineLimit(mod, config, fastUnaligned), fastUnaligned)

	} else {
		// Must be run at any optimization level.
		err := LowerInterfaces(mod, config)
		if err != nil {
			return []error{err}
		}
		errs := LowerInterrupts(mod)
		if len(errs) > 0 {
			return errs
		}

		// Clean up some leftover symbols of the previous transformations.
		po := llvm.NewPassBuilderOptions()
		defer po.Dispose()
		err = mod.RunPasses("globaldce", llvm.TargetMachine{}, po)
		if err != nil {
			return []error{fmt.Errorf("could not build pass pipeline: %w", err)}
		}
	}

	if speedLevel > 0 && config.PanicUnwind() == "asyncify" {
		AddUnwindAssumptions(mod)
	}

	if config.VerifyIR() {
		if errs := ircheck.Module(mod); errs != nil {
			return errs
		}
	}
	if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
		return []error{errors.New("optimizations caused a verification failure")}
	}

	// After TinyGo-specific transforms have finished, undo exporting these functions.
	for _, name := range functionsUsedInTransforms {
		fn := mod.NamedFunction(name)
		if fn.IsNil() || fn.IsDeclaration() {
			continue
		}
		fn.SetLinkage(llvm.InternalLinkage)
	}

	// Run the ThinLTO pre-link passes, meant to be run on each individual
	// module. This saves compilation time compared to "default<#>" and is meant
	// to better match the optimization passes that are happening during
	// ThinLTO.
	po := llvm.NewPassBuilderOptions()
	defer po.Dispose()
	passes := fmt.Sprintf("thinlto-pre-link<%s>", optLevel)
	blockGlobalAllocPromotion(mod)
	err := mod.RunPasses(passes, llvm.TargetMachine{}, po)
	removeGlobalAllocPromotionMarker(mod)
	if err != nil {
		return []error{fmt.Errorf("could not build pass pipeline: %w", err)}
	}

	// Keep these temporary roots through the ThinLTO pre-link pipeline, which
	// can run argument promotion and reconstruct the oversized signatures.
	if !mod.NamedGlobal("tinygo.indirect-abi").IsNil() {
		llvmutil.RemoveGlobalReferences(mod, "llvm.used", "tinygo.indirect-abi")
		cleanupOptions := llvm.NewPassBuilderOptions()
		defer cleanupOptions.Dispose()
		if err := mod.RunPasses("globaldce", llvm.TargetMachine{}, cleanupOptions); err != nil {
			return []error{fmt.Errorf("could not run final globaldce pass: %w", err)}
		}
	}

	if config.Scheduler() == "none" {
		// Check only after temporary ABI roots have been removed and dead code
		// eliminated. Otherwise, a dead function kept alive solely to prevent
		// argument promotion can produce a spurious scheduler error.
		if start := mod.NamedFunction("internal/task.start"); !start.IsNil() && len(getUses(start)) > 0 {
			errs := []error{}
			for _, call := range getUses(start) {
				errs = append(errs, errorAt(call, "attempted to start a goroutine without a scheduler"))
			}
			return errs
		}
	}

	hasGCPass := MakeGCStackSlots(mod)
	if hasGCPass {
		if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
			return []error{errors.New("GC pass caused a verification failure")}
		}
	}

	return nil
}

func blockGlobalAllocPromotion(mod llvm.Module) {
	ctx := mod.Context()
	ptrType := llvm.PointerType(ctx.Int8Type(), 0)
	marker := llvm.AddFunction(mod, "tinygo.gc.alloc.marker", llvm.FunctionType(ctx.VoidType(), []llvm.Type{ptrType}, false))

	builder := ctx.NewBuilder()
	defer builder.Dispose()
	var marked bool
	for _, name := range []string{"runtime.alloc", "runtime.alloc_noheap"} {
		alloc := mod.NamedFunction(name)
		if alloc.IsNil() {
			continue
		}
		for _, call := range getUses(alloc) {
			if call.IsACallInst().IsNil() || call.CalledValue() != alloc {
				continue
			}
			if isPointerFreeAllocation(call) {
				continue
			}

			// GlobalOpt may otherwise turn this allocation into an untyped
			// global, hiding its pointer fields from makeGCGlobalRoots.
			next := llvm.NextInstruction(call)
			if next.IsNil() {
				continue
			}
			builder.SetInsertPointBefore(next)
			builder.CreateCall(marker.GlobalValueType(), marker, []llvm.Value{call}, "")
			marked = true
		}
	}
	if !marked {
		marker.EraseFromParentAsFunction()
	}
}

func isPointerFreeAllocation(call llvm.Value) bool {
	const noPointerLayout = 3

	layout := call.Operand(1)
	return !layout.IsAConstantExpr().IsNil() &&
		layout.Opcode() == llvm.IntToPtr &&
		!layout.Operand(0).IsAConstantInt().IsNil() &&
		layout.Operand(0).ZExtValue() == noPointerLayout
}

func removeGlobalAllocPromotionMarker(mod llvm.Module) {
	marker := mod.NamedFunction("tinygo.gc.alloc.marker")
	if marker.IsNil() {
		return
	}
	for _, call := range getUses(marker) {
		call.EraseFromParentAsInstruction()
	}
	marker.EraseFromParentAsFunction()
}

// functionsUsedInTransform is a list of function symbols that may be used
// during TinyGo optimization passes so they have to be marked as external
// linkage until all TinyGo passes have finished.
var functionsUsedInTransforms = []string{
	"runtime.alloc",
	"runtime.free",
	"runtime.nilPanic",
}
