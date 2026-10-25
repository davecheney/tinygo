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
	refs := x.Referrers()
	if refs == nil {
		return false
	}
	for _, ref := range *refs {
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

// isProvenIndex returns whether index is known to be in bounds for coll (a
// slice or string) at the instruction at: index is not negative, and a
// comparison index < len(coll) on the same value of coll is true on every path
// to at.
func (b *builder) isProvenIndex(coll, index ssa.Value, at ssa.Instruction) bool {
	return b.isProvenBelowLen(coll, index, at, false)
}

// isProvenSliceLow returns whether low in coll[low:] is known to be in bounds:
// low is not negative and low <= len(coll) on every path to at.
func (b *builder) isProvenSliceLow(coll, low ssa.Value, at ssa.Instruction) bool {
	if b.isProvenBelowLen(coll, low, at, true) {
		return true
	}
	return b.nonNegative(low, map[ssa.Value]bool{}) && atMostLen(coll, low, map[ssa.Value]bool{})
}

// atMostLen returns whether v <= len(coll) always holds, for example for an
// index that a loop increments while it is smaller than len(coll). A phi that
// is already being checked is assumed to be at most len(coll); this is sound
// because the result is true only if every input of every phi is.
func atMostLen(coll, v ssa.Value, visiting map[ssa.Value]bool) bool {
	switch v := v.(type) {
	case *ssa.Const:
		return v.Value == nil || v.Int64() == 0
	case *ssa.Call:
		return isLenOf(v, coll)
	case *ssa.BinOp:
		if v.Op != token.ADD {
			return false
		}
		x := v.X
		if !isConstOne(v.Y) {
			if !isConstOne(v.X) {
				return false
			}
			x = v.Y
		}
		// x < len(coll) here, so x + 1 <= len(coll).
		return dominatingLess(coll, x, v, false)
	case *ssa.Phi:
		if visiting[v] {
			return true
		}
		if !definedBefore(coll, v.Block()) {
			// coll can change from one iteration to the next.
			return false
		}
		visiting[v] = true
		for _, edge := range v.Edges {
			if !atMostLen(coll, edge, visiting) {
				return false
			}
		}
		return true
	}
	return false
}

// definedBefore returns whether v is defined before block and so does not
// change while block runs again in a loop.
func definedBefore(v ssa.Value, block *ssa.BasicBlock) bool {
	instr, ok := v.(ssa.Instruction)
	if !ok || instr.Block() == nil {
		// Parameters, free variables, globals and constants.
		return true
	}
	return instr.Block() != block && instr.Block().Dominates(block)
}

func (b *builder) isProvenBelowLen(coll, v ssa.Value, at ssa.Instruction, orEqual bool) bool {
	return b.nonNegative(v, map[ssa.Value]bool{}) && dominatingLess(coll, v, at, orEqual)
}

// dominatingLess returns whether v < len(coll) (or v <= len(coll) if orEqual)
// is true on every path to at, because a comparison on the same values of v
// and coll was true.
func dominatingLess(coll, v ssa.Value, at ssa.Instruction, orEqual bool) bool {
	refs := v.Referrers()
	if refs == nil {
		return false
	}
	for _, ref := range *refs {
		cmp, ok := ref.(*ssa.BinOp)
		if !ok {
			continue
		}
		less := (cmp.Op == token.LSS && cmp.X == v && isLenOf(cmp.Y, coll)) ||
			(cmp.Op == token.GTR && cmp.Y == v && isLenOf(cmp.X, coll))
		lessOrEqual := (cmp.Op == token.LEQ && cmp.X == v && isLenOf(cmp.Y, coll)) ||
			(cmp.Op == token.GEQ && cmp.Y == v && isLenOf(cmp.X, coll))
		if !less && !(orEqual && lessOrEqual) {
			continue
		}
		for _, use := range *cmp.Referrers() {
			branch, ok := use.(*ssa.If)
			if !ok {
				continue
			}
			then := branch.Block().Succs[0]
			if len(then.Preds) == 1 && then.Dominates(at.Block()) {
				return true
			}
		}
	}
	return false
}

func isLenOf(v, coll ssa.Value) bool {
	call, ok := v.(*ssa.Call)
	if !ok {
		return false
	}
	builtin, ok := call.Call.Value.(*ssa.Builtin)
	return ok && builtin.Name() == "len" && len(call.Call.Args) == 1 && call.Call.Args[0] == coll
}

// nonNegative returns whether v is never negative. A phi that is already being
// checked is assumed to be non-negative; this is sound because the result is
// true only if every input of every phi is non-negative.
func (b *builder) nonNegative(v ssa.Value, visiting map[ssa.Value]bool) bool {
	switch v := v.(type) {
	case *ssa.Const:
		if v.Value == nil {
			return true
		}
		basic, ok := v.Type().Underlying().(*types.Basic)
		return ok && basic.Info()&types.IsInteger != 0 && v.Int64() >= 0
	case *ssa.Call:
		builtin, ok := v.Call.Value.(*ssa.Builtin)
		return ok && (builtin.Name() == "len" || builtin.Name() == "cap")
	case *ssa.Convert:
		// A small unsigned value converted to a wider signed type.
		from, ok1 := v.X.Type().Underlying().(*types.Basic)
		to, ok2 := v.Type().Underlying().(*types.Basic)
		if !ok1 || !ok2 {
			return false
		}
		switch from.Kind() {
		case types.Uint8:
			return to.Kind() == types.Int16 || to.Kind() == types.Int32 || to.Kind() == types.Int64 || to.Kind() == types.Int
		case types.Uint16:
			return to.Kind() == types.Int32 || to.Kind() == types.Int64 || to.Kind() == types.Int
		}
		return false
	case *ssa.BinOp:
		if v.Op != token.ADD || !b.isNoWrapIncrement(v) {
			return false
		}
		x := v.X
		if isConstOne(x) {
			x = v.Y
		}
		return b.nonNegative(x, visiting)
	case *ssa.Phi:
		if visiting[v] {
			return true
		}
		visiting[v] = true
		for _, edge := range v.Edges {
			if !b.nonNegative(edge, visiting) {
				return false
			}
		}
		return true
	}
	return false
}

// isProvenSliceLowOnly returns whether expr is coll[low:] with low proven to
// be in bounds. Then low <= len(coll) is the only check, so it can be skipped.
func (b *builder) isProvenSliceLowOnly(expr *ssa.Slice) bool {
	if b.info.nobounds || expr.Low == nil || expr.High != nil || expr.Max != nil {
		return false
	}
	return b.isProvenSliceLow(expr.X, expr.Low, expr)
}
