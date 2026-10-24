package transform

import (
	"fmt"
	"go/scanner"
	"sort"

	"tinygo.org/x/go-llvm"
)

// CheckNoPanic reports each check in a //go:nopanic function that the
// optimizer did not prove. Call it after all optimization passes. Inlining can
// copy a check into many functions, so each source position is reported once.
func CheckNoPanic(mod llvm.Module) []error {
	trap := mod.NamedFunction("runtime.nopanicTrap")
	if trap.IsNil() {
		return nil
	}
	type site struct {
		err     scanner.Error
		callers []string
	}
	sites := map[string]*site{}
	for _, use := range getUses(trap) {
		if use.IsACallInst().IsNil() {
			continue
		}
		err := errorAt(use, "")
		key := err.Pos.String()
		if sites[key] == nil {
			sites[key] = &site{err: err}
		}
		sites[key].callers = append(sites[key].callers, use.InstructionParent().Parent().Name())
	}
	keys := make([]string, 0, len(sites))
	for key := range sites {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var errs []error
	for _, key := range keys {
		s := sites[key]
		s.err.Msg = "//go:nopanic check is not proven"
		if len(s.callers) == 1 {
			s.err.Msg += " (in " + s.callers[0] + ")"
		} else {
			s.err.Msg += fmt.Sprintf(" (%d copies, for example in %s)", len(s.callers), s.callers[0])
		}
		errs = append(errs, s.err)
	}
	return errs
}
