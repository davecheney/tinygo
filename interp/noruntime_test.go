package interp

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"tinygo.org/x/go-llvm"
)

func TestNoRuntimeInitializerPasses(t *testing.T) {
	for _, resolved := range []bool{false, true} {
		t.Run(map[bool]string{false: "unresolved", true: "resolved-after-link"}[resolved], func(t *testing.T) {
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			dir := t.TempDir()
			path := filepath.Join(dir, "input.ll")
			const source = `
target datalayout = "e-p:64:64"
@result = global i32 0
@count = global i32 0
declare i32 @imported(ptr)
define internal i1 @marked(ptr %context) #0 {
entry:
  %count = load i32, ptr @count
  %next = add i32 %count, 1
  store i32 %next, ptr @count
  %value = call i32 @imported(ptr undef)
  store i32 %value, ptr @result
  ret i1 false
}
define void @main.init(ptr %context) {
entry:
  %unused = call i1 @marked(ptr undef)
  ret void
}
define void @runtime.initAll() {
entry:
  call void @main.init(ptr undef)
  ret void
}
attributes #0 = { "tinygo-noruntime"="main.result" }
`
			if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
				t.Fatal(err)
			}
			buf, err := llvm.NewMemoryBufferFromFile(path)
			if err != nil {
				t.Fatal(err)
			}
			mod, err := ctx.ParseIR(buf)
			if err != nil {
				t.Fatal(err)
			}
			defer mod.Dispose()
			if err := RunFunc(mod.NamedFunction("main.init"), time.Minute, DefaultMaxInterpBlockEntries, false); err != nil {
				t.Fatal(err)
			}
			call := mod.NamedFunction("main.init").EntryBasicBlock().FirstInstruction()
			if call.IsACallInst().IsNil() || call.CalledValue().Name() != "marked" {
				t.Fatalf("partial initializer was not retained:\n%s", mod.String())
			}
			if mod.NamedGlobal("count").Initializer().ZExtValue() != 0 {
				t.Fatal("partial initializer changed global state")
			}

			cachePath := filepath.Join(dir, "package.bc")
			bitcode := llvm.WriteBitcodeToMemoryBuffer(mod)
			err = os.WriteFile(cachePath, bitcode.Bytes(), 0o600)
			bitcode.Dispose()
			if err != nil {
				t.Fatal(err)
			}
			cached, err := ctx.ParseBitcodeFile(cachePath)
			if err != nil {
				t.Fatal(err)
			}
			defer cached.Dispose()
			if resolved {
				fn := cached.NamedFunction("imported")
				builder := ctx.NewBuilder()
				builder.SetInsertPointAtEnd(ctx.AddBasicBlock(fn, "entry"))
				builder.CreateRet(llvm.ConstInt(ctx.Int32Type(), 84, false))
				builder.Dispose()
			}
			if err := Run(cached, time.Minute, DefaultMaxInterpBlockEntries, false); err != nil {
				t.Fatal(err)
			}
			if err := llvm.VerifyModule(cached, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
			inst := cached.NamedFunction("runtime.initAll").EntryBasicBlock().FirstInstruction()
			if resolved {
				if inst.IsAReturnInst().IsNil() {
					t.Fatalf("initializer was not evaluated after linking:\n%s", cached.String())
				}
				if cached.NamedGlobal("result").Initializer().ZExtValue() != 84 || cached.NamedGlobal("count").Initializer().ZExtValue() != 1 {
					t.Fatal("initializer did not run exactly once")
				}
			} else if inst.IsACallInst().IsNil() || inst.CalledValue().Name() != "marked" {
				t.Fatalf("runtime work lost its initializer boundary:\n%s", cached.String())
			}
		})
	}
}
