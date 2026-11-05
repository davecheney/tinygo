package compiler

import (
	"fmt"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"github.com/tinygo-org/tinygo/transform"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

func TestLengthRangeAttributes(t *testing.T) {
	targets := []struct {
		name    string
		options *compileopts.Options
		bits    int
	}{
		{"wasm", &compileopts.Options{Target: "wasm"}, 32},
		{"amd64", &compileopts.Options{GOOS: "linux", GOARCH: "amd64"}, 64},
	}
	params := []struct {
		function string
		index    int
		name     string
		length   bool
	}{
		{"main.sliceArgs", 1, "s.data", false},
		{"main.sliceArgs", 2, "s.len", true},
		{"main.sliceArgs", 3, "s.cap", true},
		{"main.sliceArgs", 4, "n", false},
		{"main.sliceArgs", 5, "context", false},
		{"main.stringArg", 1, "s.data", false},
		{"main.stringArg", 2, "s.len", true},
		{"main.namedArgs", 2, "s.len", true},
		{"main.namedArgs", 3, "s.cap", true},
		{"main.namedArgs", 5, "text.len", true},
		{"main.nestedArg", 2, "s.text.len", true},
		{"main.nestedArg", 3, "s.n", false},
		{"main.unexpandedArg", 1, "s", false},
		{"main.ordinaryFields", 1, "s.len", false},
		{"main.ordinaryFields", 2, "s.cap", false},
		{"main.indirectResult", 1, "return", false},
		{"main.indirectResult", 3, "s.len", true},
		{"main.indirectResult", 4, "s.cap", true},
		{"(main.byteSlice).size", 2, "s.len", true},
		{"(main.byteSlice).size", 3, "s.cap", true},
		{"exportedArgs", 2, "s.len", false},
		{"exportedArgs", 3, "s.cap", false},
		{"exportedArgs", 5, "text.len", false},
	}

	for _, target := range targets {
		t.Run(target.name, func(t *testing.T) {
			mod, errs := testCompilePackage(t, target.options, "lenrange.go")
			if len(errs) != 0 {
				t.Fatal(errs)
			}
			defer mod.Dispose()

			kind := llvm.AttributeKindID("range")
			if llvmutil.Version() >= 19 && kind == 0 {
				t.Fatal("LLVM does not recognize the range attribute")
			}
			for _, param := range params {
				fn := mod.NamedFunction(param.function)
				if fn.IsNil() {
					t.Errorf("missing function %s", param.function)
					continue
				}
				if param.index > fn.ParamsCount() {
					t.Errorf("%s has no parameter %d", param.function, param.index)
					continue
				}
				if got := fn.Param(param.index - 1).Name(); got != param.name {
					t.Errorf("%s parameter %d is %q, want %q", param.function, param.index, got, param.name)
					continue
				}
				hasRange := kind != 0 && !fn.GetEnumAttributeAtIndex(param.index, kind).IsNil()
				wantRange := param.length && llvmutil.Version() >= 19
				if hasRange != wantRange {
					t.Errorf("%s parameter %s: range attribute = %t, want %t", param.function, param.name, hasRange, wantRange)
					continue
				}
				if wantRange {
					want := fmt.Sprintf("i%d range(i%d 0, -%d) %%%s", target.bits, target.bits, uint64(1)<<(target.bits-1), param.name)
					if !strings.Contains(fn.String(), want) {
						t.Errorf("%s parameter %s does not have %q:\n%s", param.function, param.name, want, fn.String())
					}
				}
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// On AVR, maxSliceSize permits byte slice lengths that set the sign bit of
// the 16-bit length field, so these lengths must not get a range attribute.
func TestLengthRangeAttributes16Bit(t *testing.T) {
	config := &Config{Triple: "avr", GOOS: "linux", GOARCH: "arm"}
	machine, err := NewTargetMachine(config)
	if err != nil {
		t.Fatal(err)
	}
	defer machine.Dispose()

	c := newCompilerContext("lenrange-avr", machine, config, false)
	defer c.ctx.Dispose()
	defer c.mod.Dispose()
	defer c.targetData.Dispose()
	defer c.dispose()
	if c.uintptrType.IntTypeWidth() != 16 || c.intType.IntTypeWidth() != 32 {
		t.Fatal("expected 16-bit pointers and 32-bit Go ints")
	}
	if got := c.maxSliceSize(c.ctx.Int8Type()); got != 65535 {
		t.Fatalf("maximum byte slice length = %d, want 65535", got)
	}

	c.llvmTypes.Set(types.Typ[types.String], c.ctx.StructType([]llvm.Type{c.dataPtrType, c.uintptrType}, false))
	params := types.NewTuple(
		types.NewParam(token.NoPos, nil, "s", types.NewSlice(types.Typ[types.Byte])),
		types.NewParam(token.NoPos, nil, "text", types.Typ[types.String]),
	)
	sig := types.NewSignatureType(nil, nil, nil, params, nil, false)
	program := ssa.NewProgram(token.NewFileSet(), 0)
	fn := program.NewFunction("sliceAndString", sig, "")
	_, llvmFn := c.getFunction(fn)

	kind := llvm.AttributeKindID("range")
	if llvmutil.Version() >= 19 && kind == 0 {
		t.Fatal("LLVM does not recognize the range attribute")
	}
	for _, index := range []int{2, 3, 5} {
		param := llvmFn.Param(index - 1)
		if got := param.Type().IntTypeWidth(); got != 16 {
			t.Errorf("parameter %d has width %d, want 16", index, got)
		}
		if kind != 0 && !llvmFn.GetEnumAttributeAtIndex(index, kind).IsNil() {
			t.Errorf("parameter %d has a range attribute on a 16-bit length", index)
		}
	}
	if err := llvm.VerifyModule(c.mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
}

func TestLengthRangeWasmExport(t *testing.T) {
	for _, scheduler := range []string{"none", "asyncify"} {
		t.Run(scheduler, func(t *testing.T) {
			mod, errs := testCompilePackage(t, &compileopts.Options{Target: "wasm", Scheduler: scheduler}, "lenrange-wasmexport.go")
			if len(errs) != 0 {
				t.Fatal(errs)
			}
			defer mod.Dispose()

			kind := llvm.AttributeKindID("range")
			for _, name := range []string{
				"main.exportString",
				"github.com/tinygo-org/tinygo/compiler/testdata/lenrangedeps.WasmExportString",
			} {
				fn := mod.NamedFunction(name)
				if fn.IsNil() {
					t.Fatalf("missing function %s", name)
				}
				if kind != 0 && !fn.GetEnumAttributeAtIndex(2, kind).IsNil() {
					t.Errorf("WebAssembly export has a range attribute on its string length:\n%s", fn.String())
				}
			}

			passOptions := llvm.NewPassBuilderOptions()
			defer passOptions.Dispose()
			if err := mod.RunPasses("instcombine", llvm.TargetMachine{}, passOptions); err != nil {
				t.Fatal(err)
			}
			fn := mod.NamedFunction("main.exportString")
			if strings.Contains(fn.String(), "ret i1 false") {
				t.Errorf("WebAssembly export cannot check for a negative string length:\n%s", fn.String())
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLengthRangeOptimization(t *testing.T) {
	for _, withRange := range []bool{false, true} {
		t.Run(fmt.Sprintf("range=%t", withRange), func(t *testing.T) {
			mod, errs := testCompilePackage(t, &compileopts.Options{Target: "wasm"}, "lenrange.go")
			if len(errs) != 0 {
				t.Fatal(errs)
			}
			defer mod.Dispose()
			if !withRange {
				removeRangeAttributes(mod)
			}

			passOptions := llvm.NewPassBuilderOptions()
			defer passOptions.Dispose()
			if err := mod.RunPasses("instcombine", llvm.TargetMachine{}, passOptions); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"main.negativeLength", "main.negativeCapacity", "main.negativeString"} {
				fn := mod.NamedFunction(name)
				if fn.IsNil() {
					t.Fatalf("missing function %s", name)
				}
				isFalse := strings.Contains(fn.String(), "ret i1 false")
				if want := withRange && llvmutil.Version() >= 19; isFalse != want {
					t.Errorf("%s optimized to false = %t, want %t:\n%s", name, isFalse, want, fn.String())
				}
			}
			fn := mod.NamedFunction("main.negativeInt")
			if fn.IsNil() {
				t.Fatal("missing function main.negativeInt")
			}
			if strings.Contains(fn.String(), "ret i1 false") {
				t.Errorf("an ordinary int was treated as non-negative:\n%s", fn.String())
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLengthRangeBoundsChecks(t *testing.T) {
	options := &compileopts.Options{Target: "wasm", Opt: "z"}
	target, err := compileopts.LoadTarget(options)
	if err != nil {
		t.Fatal(err)
	}
	config := &compileopts.Config{Options: options, Target: target}

	for _, withRange := range []bool{false, true} {
		t.Run(fmt.Sprintf("range=%t", withRange), func(t *testing.T) {
			mod, errs := testCompilePackage(t, options, "lenrange.go")
			if len(errs) != 0 {
				t.Fatal(errs)
			}
			defer mod.Dispose()
			if !withRange {
				removeRangeAttributes(mod)
			}

			for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
				if !fn.IsDeclaration() {
					transform.AddStandardAttributes(fn, config)
				}
			}
			passOptions := llvm.NewPassBuilderOptions()
			defer passOptions.Dispose()
			// Match the size pipeline in transform.Optimize.
			level := "Oz"
			if llvmutil.Version() >= 23 {
				level = "O2"
			}
			if err := mod.RunPasses("thinlto-pre-link<"+level+">", llvm.TargetMachine{}, passOptions); err != nil {
				t.Fatal(err)
			}
			for _, test := range []struct {
				name      string
				panic     string
				wantCheck bool
			}{
				{"main.rangeSum", "@runtime.lookupPanic", false},
				{"main.index", "@runtime.lookupPanic", true},
				{"main.tail", "@runtime.slicePanic", true},
			} {
				fn := mod.NamedFunction(test.name)
				if fn.IsNil() {
					t.Fatalf("missing function %s", test.name)
				}
				hasCheck := strings.Contains(fn.String(), test.panic)
				if hasCheck != test.wantCheck {
					t.Errorf("%s bounds check = %t, want %t:\n%s", test.name, hasCheck, test.wantCheck, fn.String())
				}
			}
			fn := mod.NamedFunction("main.countedSum")
			if fn.IsNil() {
				t.Fatal("missing function main.countedSum")
			}
			hasCheck := strings.Contains(fn.String(), "@runtime.lookupPanic")
			t.Logf("counted loop bounds check with range=%t: %t", withRange, hasCheck)
			if withRange && llvmutil.Version() >= 19 && hasCheck {
				t.Errorf("main.countedSum retained a loop bounds check:\n%s", fn.String())
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func removeRangeAttributes(mod llvm.Module) {
	kind := llvm.AttributeKindID("range")
	if kind == 0 {
		return
	}
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		for i := 1; i <= fn.ParamsCount(); i++ {
			fn.RemoveEnumAttributeAtIndex(i, kind)
		}
	}
}

// Runtime helpers receive Go lengths as uintptr or runtime._string values.
func TestLengthRangeRuntimeDeclarations(t *testing.T) {
	targets := []struct {
		name    string
		options *compileopts.Options
	}{
		{"wasm", &compileopts.Options{Target: "wasm"}},
		{"amd64", &compileopts.Options{GOOS: "linux", GOARCH: "amd64"}},
	}
	params := []struct {
		function string
		index    int
		length   bool
	}{
		{"runtime.sliceAppend", 1, false}, // srcBuf
		{"runtime.sliceAppend", 2, false}, // elemsBuf
		{"runtime.sliceAppend", 3, true},  // srcLen
		{"runtime.sliceAppend", 4, true},  // srcCap
		{"runtime.sliceAppend", 5, true},  // elemsLen
		{"runtime.sliceAppend", 6, false}, // elemSize
		{"runtime.sliceGrow", 2, true},    // oldLen
		{"runtime.sliceGrow", 3, true},    // oldCap
		{"runtime.sliceGrow", 4, false},   // newCap
		{"runtime.sliceGrow", 5, false},   // elemSize
		{"runtime.stringConcat", 1, false},
		{"runtime.stringConcat", 2, true},
		{"runtime.stringConcat", 3, false},
		{"runtime.stringConcat", 4, true},
		{"runtime.stringToBytes", 1, false},
		{"runtime.stringToBytes", 2, true},
	}

	for _, target := range targets {
		t.Run(target.name, func(t *testing.T) {
			mod, errs := testCompilePackage(t, target.options, "lenrange-runtime.go")
			if len(errs) != 0 {
				t.Fatal(errs)
			}
			defer mod.Dispose()

			kind := llvm.AttributeKindID("range")
			for _, param := range params {
				fn := mod.NamedFunction(param.function)
				if fn.IsNil() {
					t.Errorf("missing function %s", param.function)
					continue
				}
				hasRange := kind != 0 && !fn.GetEnumAttributeAtIndex(param.index, kind).IsNil()
				wantRange := param.length && llvmutil.Version() >= 19
				if hasRange != wantRange {
					t.Errorf("%s parameter %d: range attribute = %t, want %t:\n%s", param.function, param.index, hasRange, wantRange, fn.String())
				}
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Lengths loaded from memory are Go lengths too, so the loads get !range.
func TestLengthRangeLoads(t *testing.T) {
	targets := []struct {
		name    string
		options *compileopts.Options
		bits    int
	}{
		{"wasm", &compileopts.Options{Target: "wasm"}, 32},
		{"amd64", &compileopts.Options{GOOS: "linux", GOARCH: "amd64"}, 64},
	}
	functions := []struct {
		name   string
		ranged int
	}{
		{"main.fieldLength", 2},
		{"main.globalLength", 3},
		{"main.derefLength", 2},
		{"main.derefString", 1},
		{"main.elementLength", 2},
		{"main.ordinaryFields", 0},
	}

	for _, target := range targets {
		t.Run(target.name, func(t *testing.T) {
			mod, errs := testCompilePackage(t, target.options, "lenrange-load.go")
			if len(errs) != 0 {
				t.Fatal(errs)
			}
			defer mod.Dispose()

			want := fmt.Sprintf("!{i%d 0, i%d -%d}", target.bits, target.bits, uint64(1)<<(target.bits-1))
			for _, test := range functions {
				fn := mod.NamedFunction(test.name)
				if fn.IsNil() {
					t.Fatalf("missing function %s", test.name)
				}
				if got := countRangeLoads(fn); got != test.ranged {
					t.Errorf("%s has %d loads with !range, want %d:\n%s", test.name, got, test.ranged, fn.String())
				}
			}
			if !strings.Contains(mod.String(), want) {
				t.Errorf("module does not have range metadata %s", want)
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// On AVR, lengths are 16 bits wide and can set the sign bit, so loads of them
// must not get !range.
func TestLengthRangeLoads16Bit(t *testing.T) {
	config := &Config{Triple: "avr", GOOS: "linux", GOARCH: "arm"}
	machine, err := NewTargetMachine(config)
	if err != nil {
		t.Fatal(err)
	}
	defer machine.Dispose()

	c := newCompilerContext("lenrange-load-avr", machine, config, false)
	defer c.ctx.Dispose()
	defer c.mod.Dispose()
	defer c.targetData.Dispose()
	defer c.dispose()
	if c.uintptrType.IntTypeWidth() != 16 || c.intType.IntTypeWidth() != 32 {
		t.Fatal("expected 16-bit pointers and 32-bit Go ints")
	}

	c.llvmTypes.Set(types.Typ[types.String], c.ctx.StructType([]llvm.Type{c.dataPtrType, c.uintptrType}, false))
	sliceType := types.NewSlice(types.Typ[types.Byte])
	params := types.NewTuple(
		types.NewParam(token.NoPos, nil, "s", types.NewPointer(sliceType)),
		types.NewParam(token.NoPos, nil, "text", types.NewPointer(types.Typ[types.String])),
	)
	sig := types.NewSignatureType(nil, nil, nil, params, nil, false)
	program := ssa.NewProgram(token.NewFileSet(), 0)
	irbuilder := c.ctx.NewBuilder()
	defer irbuilder.Dispose()
	b := newBuilder(c, irbuilder, program.NewFunction("loadLengths", sig, ""))
	b.SetInsertPointAtEnd(c.ctx.AddBasicBlock(b.llvmFn, "entry"))
	b.loadFromStorage(b.llvmFn.Param(0), sliceType, "")
	b.loadFromStorage(b.llvmFn.Param(1), types.Typ[types.String], "")
	b.CreateRetVoid()

	lengthLoads := 0
	for inst := b.llvmFn.EntryBasicBlock().FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
		if !inst.IsALoadInst().IsNil() && inst.Type() == c.uintptrType {
			lengthLoads++
		}
	}
	if lengthLoads != 3 {
		t.Errorf("found %d 16-bit length loads, want 3:\n%s", lengthLoads, b.llvmFn.String())
	}
	if got := countRangeLoads(b.llvmFn); got != 0 {
		t.Errorf("%d loads have !range on a 16-bit length:\n%s", got, b.llvmFn.String())
	}
	if err := llvm.VerifyModule(c.mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
}

func TestLengthRangeLoadOptimization(t *testing.T) {
	for _, withRange := range []bool{false, true} {
		t.Run(fmt.Sprintf("range=%t", withRange), func(t *testing.T) {
			mod, errs := testCompilePackage(t, &compileopts.Options{Target: "wasm"}, "lenrange-load.go")
			if len(errs) != 0 {
				t.Fatal(errs)
			}
			defer mod.Dispose()
			if !withRange {
				removeRangeMetadata(mod)
			}

			passOptions := llvm.NewPassBuilderOptions()
			defer passOptions.Dispose()
			if err := mod.RunPasses("instcombine", llvm.TargetMachine{}, passOptions); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"main.negativeLoadedLength", "main.negativeLoadedCapacity", "main.negativeLoadedString"} {
				fn := mod.NamedFunction(name)
				if fn.IsNil() {
					t.Fatalf("missing function %s", name)
				}
				if isFalse := strings.Contains(fn.String(), "ret i1 false"); isFalse != withRange {
					t.Errorf("%s optimized to false = %t, want %t:\n%s", name, isFalse, withRange, fn.String())
				}
			}
			fn := mod.NamedFunction("main.negativeLoadedInt")
			if fn.IsNil() {
				t.Fatal("missing function main.negativeLoadedInt")
			}
			if strings.Contains(fn.String(), "ret i1 false") {
				t.Errorf("an ordinary int was treated as non-negative:\n%s", fn.String())
			}
		})
	}
}

func TestLengthRangeLoadBoundsChecks(t *testing.T) {
	options := &compileopts.Options{Target: "wasm", Opt: "z"}
	target, err := compileopts.LoadTarget(options)
	if err != nil {
		t.Fatal(err)
	}
	config := &compileopts.Config{Options: options, Target: target}

	for _, withRange := range []bool{false, true} {
		t.Run(fmt.Sprintf("range=%t", withRange), func(t *testing.T) {
			mod, errs := testCompilePackage(t, options, "lenrange-load.go")
			if len(errs) != 0 {
				t.Fatal(errs)
			}
			defer mod.Dispose()
			if !withRange {
				removeRangeMetadata(mod)
			}

			for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
				if !fn.IsDeclaration() {
					transform.AddStandardAttributes(fn, config)
				}
			}
			passOptions := llvm.NewPassBuilderOptions()
			defer passOptions.Dispose()
			// Match the size pipeline in transform.Optimize.
			level := "Oz"
			if llvmutil.Version() >= 23 {
				level = "O2"
			}
			if err := mod.RunPasses("thinlto-pre-link<"+level+">", llvm.TargetMachine{}, passOptions); err != nil {
				t.Fatal(err)
			}
			fn := mod.NamedFunction("main.globalSum")
			if fn.IsNil() {
				t.Fatal("missing function main.globalSum")
			}
			if hasCheck := strings.Contains(fn.String(), "@runtime.lookupPanic"); hasCheck == withRange {
				t.Errorf("main.globalSum bounds check = %t, want %t:\n%s", hasCheck, !withRange, fn.String())
			}
		})
	}
}

func countRangeLoads(fn llvm.Value) int {
	kind := llvm.MDKindID("range")
	count := 0
	for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
		for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
			if !inst.IsALoadInst().IsNil() && !inst.Metadata(kind).IsNil() {
				count++
			}
		}
	}
	return count
}

func removeRangeMetadata(mod llvm.Module) {
	kind := llvm.MDKindID("range")
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
			for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
				if !inst.IsALoadInst().IsNil() {
					inst.SetMetadata(kind, llvm.Metadata{})
				}
			}
		}
	}
}
