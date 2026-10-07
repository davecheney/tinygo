package main

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/interp"
)

func TestCompileTime(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
		fail   bool
	}{
		{"evaluated", `var result = calc(12)
//go:compiletime
func calc(n int) int { return n * 7 }
func main() { println(result) }`, false},
		{"loop-limit", `var result = calc(1001)
//go:compiletime
func calc(n int) int { sum := 0; for i := 0; i < n; i++ { sum += i }; return sum }
func main() { println(result) }`, true},
		{"loop-below-limit", `var result = calc(999)
//go:compiletime
func calc(n int) int { sum := 0; for i := 0; i < n; i++ { sum += i }; return sum }
func main() { println(result) }`, false},
		{"runtime", `//go:compiletime
func calc(n int) int { return n * 7 }
func main() { println(calc(12)) }`, true},
		{"inline-hint", `//go:inline
//go:compiletime
func calc(n int) int { return n * 7 }
func main() { println(calc(12)) }`, true},
		{"unused", `//go:compiletime
func calc(n int) int { return n * 7 }
func main() {}`, false},
		{"dead-caller", `//go:compiletime
func calc(n int) int { return n * 7 }
func dead() int { return calc(12) }
func main() {}`, false},
		{"function-value", `var callback = calc
//go:compiletime
func calc(n int) int { return n * 7 }
func main() { println(callback(12)) }`, true},
		{"evaluated-function-value", `var result = func() int { f := calc; return f(12) }()
//go:compiletime
func calc(n int) int { return n * 7 }
func main() { println(result) }`, false},
		{"unused-function-value", `var callback = calc
//go:compiletime
func calc(n int) int { return n * 7 }
func main() {}`, false},
		{"unsupported", `import "runtime/volatile"
var reg volatile.Register32
var result = calc(0)
//go:compiletime
func calc(n int) uint32 { return reg.Get() }
func main() { println(result) }`, true},
		{"dead-aggregate", `type large [100]int
var result = calc(large{12})
//go:compiletime
func calc(n large) int { return n[0] * 7 }
func main() { println(result) }`, false},
		{"generic", `var result = calc(12)
//go:compiletime
func calc[T ~int](n T) T { return n * 7 }
func main() { println(result) }`, false},
		{"generic-runtime", `//go:compiletime
func calc[T ~int](n T) T { return n * 7 }
func main() { println(calc(12)) }`, true},
		{"method-value", `type number int
var callback = number(12).calc
//go:compiletime
func (n number) calc() int { return int(n) * 7 }
func main() { println(callback()) }`, true},
		{"interface-call", `type number int
type calculator interface { calc() int }
var callback calculator = number(12)
//go:compiletime
func (n number) calc() int { return int(n) * 7 }
func main() { println(callback.calc()) }`, true},
		{"exported", `//go:export calc
//go:compiletime
func calc(n int) int { return n * 7 }
func main() {}`, true},
		{"retained-method-table", `type number int
type observer interface { other() int }
var callback observer = number(12)
//go:compiletime
func (n number) calc() int { return int(n) * 7 }
func (n number) other() int { return int(n) }
func main() { println(callback.other()) }`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, opt := range []string{"0", "2"} {
				t.Run(opt, func(t *testing.T) {
					tmpdir := t.TempDir()
					path := filepath.Join(tmpdir, "main.go")
					if err := os.WriteFile(path, []byte("package main\n"+tc.source+"\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					options := optionsFromTarget("wasip1", sema)
					options.Opt = opt
					options.Scheduler = "none"
					options.InterpMaxLoopIterations = interp.DefaultMaxInterpBlockEntries
					config, err := builder.NewConfig(&options)
					if err != nil {
						t.Fatal(err)
					}
					out := filepath.Join(tmpdir, "main.ll")
					_, err = builder.Build(path, out, tmpdir, config)
					if tc.fail {
						if err == nil || !strings.Contains(err.Error(), "//go:compiletime function ") || !strings.Contains(err.Error(), " remains reachable at runtime") {
							t.Fatalf("expected compiletime error, got %v", err)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					data, err := os.ReadFile(out)
					if err != nil {
						t.Fatal(err)
					}
					if strings.Contains(string(data), "tinygo-compiletime") {
						t.Fatal("compiletime function remains in output")
					}
				})
			}
		})
	}
}

func TestCompileTimePLL(t *testing.T) {
	for _, tc := range []struct {
		name     string
		combined bool
		limit    int
		fail     bool
	}{
		{"split", false, interp.DefaultMaxInterpBlockEntries, false},
		{"combined", true, interp.DefaultMaxInterpBlockEntries, true},
		{"combined-no-limit", true, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "testdata/compiletime-pll.go", nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			var table, entry *ast.FuncDecl
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					switch fn.Name.Name {
					case "genTable":
						table = fn
					case "genTableEntry":
						entry = fn
					}
				}
			}
			if table == nil || entry == nil {
				t.Fatal("missing PLL table functions")
			}
			if tc.combined {
				table.Body.List[1].(*ast.ForStmt).Body.List = entry.Body.List
			}
			var source bytes.Buffer
			if err := printer.Fprint(&source, fset, file); err != nil {
				t.Fatal(err)
			}
			code := strings.ReplaceAll(source.String(), "func genTable", "//go:compiletime\nfunc genTable")
			tmpdir := t.TempDir()
			path := filepath.Join(tmpdir, "main.go")
			if err := os.WriteFile(path, []byte(code), 0o600); err != nil {
				t.Fatal(err)
			}
			options := optionsFromTarget("wasip1", sema)
			options.Opt = "2"
			options.Scheduler = "none"
			options.InterpMaxLoopIterations = tc.limit
			config, err := builder.NewConfig(&options)
			if err != nil {
				t.Fatal(err)
			}
			result, err := builder.Build(path, filepath.Join(tmpdir, "main.wasm"), tmpdir, config)
			if tc.fail {
				if err == nil || !strings.Contains(err.Error(), "//go:compiletime function main.genTable remains reachable at runtime") {
					t.Fatalf("expected compiletime error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(result.Binary)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfigInterpreter())
			defer r.Close(ctx)
			if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			if _, err := r.InstantiateWithConfig(ctx, data, wazero.NewModuleConfig().WithStdout(&output).WithStderr(&output)); err != nil {
				t.Fatalf("PLL program failed: %v\n%s", err, &output)
			}
			if output.String() != "961\n" {
				t.Fatalf("unexpected PLL checksum: %q", &output)
			}
		})
	}
}
