package builder

import (
	"go/parser"
	"go/token"
	"path"
	"strings"
	"testing"
)

func TestCompilerRTSourceReferences(t *testing.T) {
	expected := make(map[string]bool)
	for _, sources := range [][]string{
		genericBuiltins,
		genericBuiltins128,
		aeabiBuiltins,
		avrBuiltins,
		windowsI386Builtins,
		windowsAMD64Builtins,
		windowsARM64Builtins,
	} {
		for _, source := range sources {
			if path.Ext(source) == ".S" {
				expected[path.Join("llvm-project/compiler-rt/lib/builtins", source)] = true
			}
		}
	}

	file, err := parser.ParseFile(token.NewFileSet(), "../compiler_rt_sources.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range file.Comments {
		for _, comment := range group.List {
			patterns, ok := strings.CutPrefix(comment.Text, "//go:embed ")
			if !ok {
				continue
			}
			for _, source := range strings.Fields(patterns) {
				if !expected[source] {
					t.Errorf("unexpected or duplicate compiler-rt source reference: %s", source)
				}
				delete(expected, source)
			}
		}
	}
	for source := range expected {
		t.Errorf("missing compiler-rt source reference: %s", source)
	}
}
