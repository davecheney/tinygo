package compiler

import (
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ssa"
)

// isNoWrapIncrement returns whether expr is x + 1 for a signed integer x that
// is known to be smaller than some other value at this point, so the addition
// cannot overflow.
func (b *builder) isNoWrapIncrement(expr *ssa.BinOp) bool {
	if expr.Op != token.ADD {
		return false
	}
	basic, ok := expr.Type().Underlying().(*types.Basic)
	if !ok || basic.Info()&types.IsInteger == 0 || basic.Info()&types.IsUnsigned != 0 {
		return false
	}
	x := expr.X
	if !isConstOne(expr.Y) {
		if !isConstOne(expr.X) {
			return false
		}
		x = expr.Y
	}
	if _, ok := x.(*ssa.Const); ok {
		return false
	}
	for _, ref := range *x.Referrers() {
		if b.boundsCheckedBefore(ref, x, expr) || lessThanBefore(ref, x, expr) {
			return true
		}
	}
	return false
}

func isConstOne(v ssa.Value) bool {
	c, ok := v.(*ssa.Const)
	return ok && c.Value != nil && c.Int64() == 1
}

// boundsCheckedBefore returns whether ref is a bounds-checked use of x as an
// index that always runs before expr. After it, 0 <= x < len, and len fits in
// an int, so an int x is smaller than the maximum int.
func (b *builder) boundsCheckedBefore(ref ssa.Instruction, x ssa.Value, expr *ssa.BinOp) bool {
	if b.info.nobounds {
		return false
	}
	if basic, ok := x.Type().Underlying().(*types.Basic); !ok || basic.Kind() != types.Int {
		return false
	}
	switch ref := ref.(type) {
	case *ssa.IndexAddr:
		if ref.Index != x {
			return false
		}
	case *ssa.Index:
		if ref.Index != x {
			return false
		}
	case *ssa.Lookup:
		if ref.Index != x {
			return false
		}
		if _, ok := ref.X.Type().Underlying().(*types.Basic); !ok {
			// Map lookups have no bounds check.
			return false
		}
	default:
		return false
	}
	return runsBefore(ref, expr)
}

// lessThanBefore returns whether ref is a comparison x < y (or y > x) whose
// true branch is the only way to reach expr.
func lessThanBefore(ref ssa.Instruction, x ssa.Value, expr *ssa.BinOp) bool {
	cmp, ok := ref.(*ssa.BinOp)
	if !ok || !((cmp.Op == token.LSS && cmp.X == x) || (cmp.Op == token.GTR && cmp.Y == x)) {
		return false
	}
	for _, use := range *cmp.Referrers() {
		branch, ok := use.(*ssa.If)
		if !ok {
			continue
		}
		then := branch.Block().Succs[0]
		if len(then.Preds) == 1 && then.Dominates(expr.Block()) {
			return true
		}
	}
	return false
}

// runsBefore returns whether instruction a always runs before b.
func runsBefore(a ssa.Instruction, b ssa.Instruction) bool {
	if a.Block() != b.Block() {
		return a.Block().Dominates(b.Block())
	}
	for _, instr := range a.Block().Instrs {
		if instr == a {
			return true
		}
		if instr == b {
			return false
		}
	}
	return false
}
