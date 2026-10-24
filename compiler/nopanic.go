package compiler

import (
	"fmt"
	"strings"
)

// panicKinds is a set of check kinds that a //go:nopanic function asserts the
// compiler can prove.
type panicKinds uint8

const (
	panicIndex panicKinds = 1 << iota
	panicSlice
	panicUnsafe
	panicChan
	panicNil
	panicShift
	panicDivide

	panicAll = panicIndex | panicSlice | panicUnsafe | panicChan | panicNil | panicShift | panicDivide
)

var panicKindNames = map[string]panicKinds{
	"index":  panicIndex,
	"slice":  panicSlice,
	"unsafe": panicUnsafe,
	"chan":   panicChan,
	"nil":    panicNil,
	"shift":  panicShift,
	"divide": panicDivide,
}

// panicKindOf maps the runtime function of a check to its kind.
var panicKindOf = map[string]panicKinds{
	"lookupPanic":              panicIndex,
	"slicePanic":               panicSlice,
	"sliceToArrayPointerPanic": panicSlice,
	"unsafeSlicePanic":         panicUnsafe,
	"chanMakePanic":            panicChan,
	"nilPanic":                 panicNil,
	"negativeShiftPanic":       panicShift,
	"divideByZeroPanic":        panicDivide,
}

// parsePanicKinds parses the parameters of //go:nopanic. No parameters means
// all kinds.
func parsePanicKinds(params []string) (panicKinds, error) {
	if len(params) == 0 {
		return panicAll, nil
	}
	if len(params) > 1 {
		return 0, fmt.Errorf("expected at most one parameter to //go:nopanic, not %d", len(params))
	}
	var kinds panicKinds
	for _, name := range strings.Split(params[0], ",") {
		kind, ok := panicKindNames[name]
		if !ok {
			return 0, fmt.Errorf("unknown //go:nopanic kind %q", name)
		}
		kinds |= kind
	}
	return kinds, nil
}
