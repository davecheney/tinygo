package builder

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/compileopts"
)

// knownUnproven lists //go:nopanic functions with checks that the compiler
// does not prove yet on all targets, and why.
var knownUnproven = map[string]string{
	"printstring": "the loop compares i < len(s) as signed",
	"stringEqual": "the loop compares i < len(x) as signed",
	"stringLess":  "the loop compares i < l as signed",
	"decodeUTF8":  "the caller must give an index below len(s)",
	"lenToCap":    "the compiler does not know that len <= cap",
}

// Test that the checks in //go:nopanic functions are proven, so a compiler
// change that adds one back is found.
func TestNoPanic(t *testing.T) {
	type config struct{ target, pkg, opt string }
	var configs []config
	for _, opt := range []string{"z", "2"} {
		for _, target := range []string{"cortex-m-qemu", "riscv-qemu", "pico", "pico2"} {
			configs = append(configs, config{target, "./testdata/nopanic", opt})
		}
		for _, target := range []string{"pico", "pico2"} {
			configs = append(configs, config{target, "examples/serial", opt})
		}
	}

	var mu sync.Mutex
	seen := map[string]bool{}
	var wg sync.WaitGroup
	for _, c := range configs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs, err := buildNoPanic(c.target, c.pkg, c.opt, t.TempDir())
			if err != nil {
				t.Errorf("%s %s -opt=%s: %v", c.target, c.pkg, c.opt, err)
				return
			}
			for _, e := range errs {
				name := enclosingFunc(e.Pos)
				mu.Lock()
				seen[name] = true
				mu.Unlock()
				if _, ok := knownUnproven[name]; !ok {
					t.Errorf("%s %s -opt=%s: %s: %s in %s", c.target, c.pkg, c.opt, e.Pos, e.Msg, name)
				}
			}
		}()
	}
	wg.Wait()

	var stale []string
	for name := range knownUnproven {
		if !seen[name] {
			stale = append(stale, name)
		}
	}
	sort.Strings(stale)
	for _, name := range stale {
		t.Errorf("%s is now proven on all tested targets, remove it from knownUnproven", name)
	}
}

// buildNoPanic builds a program with CheckNoPanic and returns the checks that
// are not proven.
func buildNoPanic(targetString, pkgName, opt, tmpdir string) ([]scanner.Error, error) {
	options := compileopts.Options{
		Target:        targetString,
		Opt:           opt,
		Semaphore:     sema,
		InterpTimeout: 60 * time.Second,
		Debug:         true,
		CheckNoPanic:  true,
	}
	target, err := compileopts.LoadTarget(&options)
	if err != nil {
		return nil, err
	}
	config := &compileopts.Config{Options: &options, Target: target}
	_, err = Build(pkgName, "", tmpdir, config)
	if err == nil {
		return nil, nil
	}
	errs := []error{err}
	var multi *MultiError
	if errors.As(err, &multi) {
		errs = multi.Errs
	}
	var result []scanner.Error
	for _, err := range errs {
		var serr scanner.Error
		if !errors.As(err, &serr) || !strings.HasPrefix(serr.Msg, "//go:nopanic") {
			return nil, err
		}
		result = append(result, serr)
	}
	return result, nil
}

// enclosingFunc returns the name of the function declaration at pos.
func enclosingFunc(pos token.Position) string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, pos.Filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return pos.String()
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fset.Position(fn.Pos()).Line <= pos.Line && pos.Line <= fset.Position(fn.End()).Line {
			return fn.Name.Name
		}
	}
	return pos.String()
}
