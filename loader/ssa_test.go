package loader

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"golang.org/x/tools/go/ssa"
)

func TestLoadSSANestedShift(t *testing.T) {
	for _, entry := range []string{"package", "program"} {
		for _, test := range []struct {
			name  string
			expr  string
			kinds []types.BasicKind
		}{
			{"left", "x >> (1 << k)", []types.BasicKind{types.Int, types.Uint64}},
			{"right", "x << (8 >> k)", []types.BasicKind{types.Int, types.Uint64}},
			{"arithmetic", "x >> ((1 << k) + 1)", []types.BasicKind{types.Int, types.Uint64}},
			{"unary", "x >> -(1 << k)", []types.BasicKind{types.Int, types.Uint64}},
			{"triple", "x >> (1 << (1 << k))", []types.BasicKind{types.Int, types.Int, types.Uint64}},
			{"rune", "x >> ('a' << k)", []types.BasicKind{types.Int32, types.Uint64}},
			{"typed", "x >> (uint64(1) << k)", []types.BasicKind{types.Uint64, types.Uint64}},
			{"converted", "x >> uint(1 << k)", []types.BasicKind{types.Uint, types.Uint64}},
			{"constant", "x >> (1 << 3)", []types.BasicKind{types.Uint64}},
			{"large constant", "x >> (1 << 63)", []types.BasicKind{types.Uint64}},
		} {
			t.Run(entry+"/"+test.name, func(t *testing.T) {
				fset := token.NewFileSet()
				file, err := parser.ParseFile(fset, "shift.go", fmt.Sprintf(`package shift
func shiftNested(x uint64, k int) uint64 {
	return %s
}
`, test.expr), 0)
				if err != nil {
					t.Fatal(err)
				}
				info := types.Info{
					Types: make(map[ast.Expr]types.TypeAndValue),
					Defs:  make(map[*ast.Ident]types.Object),
					Uses:  make(map[*ast.Ident]types.Object),
				}
				pkg, err := new(types.Config).Check("shift", fset, []*ast.File{file}, &info)
				if err != nil {
					t.Fatal(err)
				}
				program := &Program{fset: fset}
				p := &Package{program: program, Pkg: pkg, Files: []*ast.File{file}, info: info}
				program.sorted = []*Package{p}
				var ssaPkg *ssa.Package
				if entry == "package" {
					ssaPkg = p.LoadSSA()
				} else {
					ssaPkg = program.LoadSSA().Package(pkg)
				}
				ssaPkg.Build()
				var kinds []types.BasicKind
				for _, block := range ssaPkg.Func("shiftNested").Blocks {
					for _, instr := range block.Instrs {
						if binop, ok := instr.(*ssa.BinOp); ok && (binop.Op == token.SHL || binop.Op == token.SHR) {
							kinds = append(kinds, binop.Type().(*types.Basic).Kind())
						}
					}
				}
				if !slices.Equal(kinds, test.kinds) {
					t.Errorf("shift result types = %v, want %v", kinds, test.kinds)
				}
			})
		}
	}
}
