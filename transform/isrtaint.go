package transform

import (
	"fmt"
	"go/token"
	"io"
	"sort"
	"strings"

	"tinygo.org/x/go-llvm"
)

// isrAttr marks functions that LowerInterrupts wired to interrupt handlers.
const isrAttr = "tinygo-isr"

// isrSinks are the functions that an interrupt handler must not reach.
var isrSinks = map[string]string{
	"runtime.alloc":       "alloc",
	"internal/task.Pause": "block",
}

type isrEdge struct {
	call     llvm.Value // call site in the caller
	caller   llvm.Value
	indirect bool
}

type isrHit struct {
	root  llvm.Value
	site  llvm.Value // call to the sink
	kind  string
	chain []isrEdge
}

// ReportISRTaint walks the call graph from interrupt handlers and writes each
// reachable call to an isrSinks function to w. With cold set, calls in blocks
// that end in unreachable are also followed.
func ReportISRTaint(mod llvm.Module, w io.Writer, cold bool) {
	roots := isrRoots(mod)
	addrTaken := addressTakenByType(mod)

	var hits []isrHit
	for _, root := range roots {
		hits = append(hits, isrWalk(root, addrTaken, cold)...)
	}

	fmt.Fprintf(w, "isr-taint: %d roots\n", len(roots))
	for _, root := range roots {
		fmt.Fprintf(w, "  root %s\n", root.Name())
	}
	sort.SliceStable(hits, func(i, j int) bool {
		return len(hits[i].chain) < len(hits[j].chain)
	})
	for _, h := range hits {
		maybe := ""
		for _, e := range h.chain {
			if e.indirect {
				maybe = " (via indirect call)"
				break
			}
		}
		fmt.Fprintf(w, "%s: %s reaches %s%s\n", fmtPos(getPosition(h.site)), h.root.Name(), h.kind, maybe)
		for _, e := range h.chain {
			arrow := "->"
			if e.indirect {
				arrow = "~>"
			}
			fmt.Fprintf(w, "    %s %s %s\n", e.caller.Name(), arrow, fmtPos(getPosition(e.call)))
		}
	}
}

func fmtPos(pos token.Position) string {
	if pos.Filename == "" {
		return "?"
	}
	return pos.String()
}

// isrRoots returns functions marked by LowerInterrupts and exported vector
// functions found by name.
func isrRoots(mod llvm.Module) []llvm.Value {
	var roots []llvm.Value
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if fn.IsDeclaration() {
			continue
		}
		name := fn.Name()
		switch {
		case !fn.GetStringAttributeAtIndex(-1, isrAttr).IsNil():
		case name == "Reset_Handler":
			continue
		case strings.HasSuffix(name, "_Handler"), strings.HasSuffix(name, "_IRQHandler"):
		case strings.HasPrefix(name, "__vector_"):
		case name == "handleInterrupt":
		default:
			continue
		}
		roots = append(roots, fn)
	}
	return roots
}

// addressTakenByType maps a function type to the functions of that type that
// are used other than as a direct callee. Goroutine wrappers are left out
// because only the scheduler starts them.
func addressTakenByType(mod llvm.Module) map[llvm.Type][]llvm.Value {
	m := map[llvm.Type][]llvm.Value{}
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if fn.IsDeclaration() || strings.HasSuffix(fn.Name(), "$gowrapper") {
			continue
		}
		for _, use := range getUses(fn) {
			if !use.IsACallInst().IsNil() && use.CalledValue() == fn {
				continue
			}
			typ := fn.GlobalValueType()
			m[typ] = append(m[typ], fn)
			break
		}
	}
	return m
}

func isrWalk(root llvm.Value, addrTaken map[llvm.Type][]llvm.Value, cold bool) []isrHit {
	parent := map[llvm.Value]isrEdge{root: {}}
	queue := []llvm.Value{root}
	var hits []isrHit
	for len(queue) > 0 {
		fn := queue[0]
		queue = queue[1:]
		for _, bb := range isrBlocks(fn) {
			if !cold && !bb.LastInstruction().IsAUnreachableInst().IsNil() {
				continue
			}
			for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
				if inst.IsACallInst().IsNil() && inst.IsAInvokeInst().IsNil() {
					continue
				}
				callee := stripPointerCasts(inst.CalledValue())
				var targets []llvm.Value
				indirect := callee.IsAFunction().IsNil()
				if indirect {
					if !callee.IsAInlineAsm().IsNil() {
						continue
					}
					targets = addrTaken[inst.CalledFunctionType()]
				} else {
					if kind, ok := isrSinks[callee.Name()]; ok {
						hits = append(hits, isrHit{root, inst, kind, isrChain(parent, fn, isrEdge{inst, fn, false})})
						continue
					}
					if strings.HasPrefix(callee.Name(), "llvm.") || callee.IsDeclaration() {
						continue
					}
					targets = []llvm.Value{callee}
				}
				for _, t := range targets {
					if _, seen := parent[t]; seen {
						continue
					}
					parent[t] = isrEdge{inst, fn, indirect}
					queue = append(queue, t)
				}
			}
		}
	}
	return hits
}

// isrBlocks returns the blocks of fn that are reachable in interrupt context,
// where runtime/interrupt.In returns true.
func isrBlocks(fn llvm.Value) []llvm.BasicBlock {
	entry := fn.EntryBasicBlock()
	seen := map[llvm.BasicBlock]bool{entry: true}
	work := []llvm.BasicBlock{entry}
	var blocks []llvm.BasicBlock
	for len(work) > 0 {
		bb := work[len(work)-1]
		work = work[:len(work)-1]
		blocks = append(blocks, bb)
		term := bb.LastInstruction()
		n := term.SuccessorsCount()
		first, last := 0, n-1
		if !term.IsABranchInst().IsNil() && n == 2 {
			if in, ok := isInterruptIn(term.Operand(0)); ok {
				if in {
					last = 0
				} else {
					first = 1
				}
			}
		}
		for i := first; i <= last; i++ {
			succ := term.Successor(i)
			if !seen[succ] {
				seen[succ] = true
				work = append(work, succ)
			}
		}
	}
	return blocks
}

// isInterruptIn reports whether cond is runtime/interrupt.In() or its
// negation. The bool result is false for the negation.
func isInterruptIn(cond llvm.Value) (bool, bool) {
	positive := true
	if !cond.IsAInstruction().IsNil() && cond.InstructionOpcode() == llvm.Xor {
		if c := cond.Operand(1); !c.IsAConstantInt().IsNil() && c.ZExtValue() == 1 {
			cond = cond.Operand(0)
			positive = false
		}
	}
	if cond.IsACallInst().IsNil() {
		return false, false
	}
	callee := cond.CalledValue()
	if callee.IsAFunction().IsNil() || callee.Name() != "runtime/interrupt.In" {
		return false, false
	}
	return positive, true
}

func isrChain(parent map[llvm.Value]isrEdge, fn llvm.Value, last isrEdge) []isrEdge {
	chain := []isrEdge{last}
	for {
		e := parent[fn]
		if e.caller.IsNil() {
			break
		}
		chain = append(chain, e)
		fn = e.caller
	}
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}
