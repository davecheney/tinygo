package transform

import (
	"strings"
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestPoisonStackAllocs(t *testing.T) {
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	buf, err := llvm.NewMemoryBufferFromFile("testdata/stackpoison.ll")
	if err != nil {
		t.Fatal(err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer mod.Dispose()
	optimizeAllocs(mod, nil, 256, nil, true)
	optimizeAllocs(mod, nil, 256, nil, true)
	if strings.Contains(mod.String(), "llvm.memset") {
		t.Fatal("instrumentation ran before both allocation passes")
	}
	poisonStackAllocs(mod)
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
	local := mod.NamedFunction("local").String()
	if strings.Count(local, "call void @llvm.memset.p0.i64(ptr align 8 %stackalloc, i8 -91, i64 16, i1 true)") != 2 {
		t.Fatalf("expected one poison store at each return:\n%s", local)
	}
	escape := mod.NamedFunction("escape").String()
	if strings.Contains(escape, "llvm.memset") || !strings.Contains(escape, "call ptr @runtime.alloc") {
		t.Fatalf("escaping allocation changed:\n%s", escape)
	}
	po := llvm.NewPassBuilderOptions()
	defer po.Dispose()
	if err := mod.RunPasses("default<O2>", llvm.TargetMachine{}, po); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mod.NamedFunction("local").String(), "i8 -91, i64 16, i1 true)") {
		t.Fatalf("LLVM removed poison stores:\n%s", mod.String())
	}
}

func TestPoisonStackAllocsDisabled(t *testing.T) {
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	buf, err := llvm.NewMemoryBufferFromFile("testdata/stackpoison.ll")
	if err != nil {
		t.Fatal(err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer mod.Dispose()
	OptimizeAllocs(mod, nil, 256, nil)
	poisonStackAllocs(mod)
	if strings.Contains(mod.String(), "tinygo.stackalloc") || strings.Contains(mod.String(), "llvm.memset") {
		t.Fatalf("default allocation pass emitted poison metadata:\n%s", mod.String())
	}
}
