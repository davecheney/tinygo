package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestOptimizeStringToBytes(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/stringtobytes", func(mod llvm.Module) {
		// Run optimization pass.
		transform.OptimizeStringToBytes(mod)
	})
}

func TestOptimizeStringEqual(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name          string
		fastUnaligned bool
	}{
		{"stringequal", true},        // armv7m, 32-bit loads
		{"stringequal-be", true},     // mips, big endian
		{"stringequal-64", true},     // wasm32, 64-bit loads
		{"stringequal-bytes", false}, // riscv32, byte loads
	} {
		t.Run(tc.name, func(t *testing.T) {
			testTransform(t, "testdata/"+tc.name, func(mod llvm.Module) {
				// Run optimization pass.
				transform.OptimizeStringEqual(mod, 16, tc.fastUnaligned)

				// Inline the helper functions, to make the output easier
				// to read.
				po := llvm.NewPassBuilderOptions()
				defer po.Dispose()
				if err := mod.RunPasses("always-inline", llvm.TargetMachine{}, po); err != nil {
					t.Fatal(err)
				}
			})
		})
	}
}
