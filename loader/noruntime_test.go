package loader

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/ssa"
)

func TestNoRuntimeInitializerSSA(t *testing.T) {
	const source = `package main
//go:noruntime
var result = other + 1
var other = 83
//go:noruntime
var override = "default"
func main() { println(result, override) }
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "main.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	pkg := &Package{
		program: &Program{fset: fset},
		Files:   []*ast.File{file},
		info: types.Info{
			Types:        make(map[ast.Expr]types.TypeAndValue),
			Defs:         make(map[*ast.Ident]types.Object),
			Uses:         make(map[*ast.Ident]types.Object),
			Implicits:    make(map[ast.Node]types.Object),
			Scopes:       make(map[ast.Node]*types.Scope),
			Selections:   make(map[*ast.SelectorExpr]*types.Selection),
			FileVersions: make(map[*ast.File]string),
		},
	}
	config := types.Config{GoVersion: "go1.25"}
	pkg.Pkg, err = config.Check("main", fset, pkg.Files, &pkg.info)
	if err != nil {
		t.Fatal(err)
	}
	if errs := pkg.checkNoRuntimeInitializers(); len(errs) != 0 {
		t.Fatal(errs)
	}
	pkg.StripVarInitializer("override")
	ssaPkg := pkg.LoadSSA()
	ssaPkg.Build()
	if len(pkg.NoRuntimeInitializers) != 1 {
		t.Fatalf("expected one initializer after stripping override, got %d", len(pkg.NoRuntimeInitializers))
	}
	for _, name := range pkg.NoRuntimeInitializers {
		if name != "main.result" {
			t.Fatalf("unexpected initializer: %s", name)
		}
	}
	if pkg.info.InitOrder[0].Lhs[0].Name() != "other" {
		t.Fatal("dependency initialization order changed")
	}
	count := len(pkg.NoRuntimeInitializers)
	pkg.lowerNoRuntimeInitializers()
	if len(pkg.NoRuntimeInitializers) != count {
		t.Fatal("lowering was not idempotent")
	}
	init := ssaPkg.Func("init")
	if len(init.AnonFuncs) != 1 {
		t.Fatalf("expected one initializer function, got %d", len(init.AnonFuncs))
	}
	fn := init.AnonFuncs[0]
	if len(fn.FreeVars) != 0 {
		t.Fatal("initializer captures a local variable")
	}
	found := false
	for _, block := range fn.Blocks {
		for _, inst := range block.Instrs {
			if store, ok := inst.(*ssa.Store); ok && store.Addr == ssaPkg.Var("result") {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("initializer does not contain its destination store")
	}
}
