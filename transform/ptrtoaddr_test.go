package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestOptimizePtrToAddr(t *testing.T) {
	t.Parallel()
	if llvmutil.Version() < 22 {
		t.Skip("ptrtoaddr requires LLVM 22")
	}
	testTransform(t, "testdata/ptrtoaddr", func(mod llvm.Module) {
		transform.OptimizePtrToAddr(mod)
	})
}
