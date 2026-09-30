//go:build gc.conservative || gc.precise

package runtime

import (
	"internal/gclayout"
	"unsafe"
)

// GCStackDeadFrameProbe checks that markCurrentGoroutineStack ignores the part
// of a stack object below sp, which holds frames that have already returned.
func GCStackDeadFrameProbe() string {
	const size = 256
	stack := alloc(size, gclayout.Conservative.AsPtr())
	live := alloc(16, gclayout.NoPtrs.AsPtr())
	dead := alloc(16, gclayout.NoPtrs.AsPtr())

	// Put sp halfway up, so one word sits in a returned frame below it and
	// one word sits in the frame it is about to scan.
	deadWord := stack
	sp := unsafe.Add(stack, size/2)
	liveWord := unsafe.Add(stack, size/2+unsafe.Sizeof(uintptr(0)))
	*(*unsafe.Pointer)(deadWord) = dead
	*(*unsafe.Pointer)(liveWord) = live

	gcLock.Lock()
	markCurrentGoroutineStack(uintptr(sp))
	finishMark()
	liveMarked := blockFromAddr(uintptr(live)).findHead().state() == blockStateMark
	deadMarked := blockFromAddr(uintptr(dead)).findHead().state() == blockStateMark
	for block := gcBlock(0); block < endBlock; block++ {
		if block.state() == blockStateMark {
			header := (*objHeader)(unsafe.Add(block.pointer(), bytesPerBlock-unsafe.Sizeof(objHeader{})))
			if header.next != 1 {
				block.unmark()
			}
		}
	}
	gcLock.Unlock()

	*(*unsafe.Pointer)(deadWord) = nil
	*(*unsafe.Pointer)(liveWord) = nil

	if !liveMarked {
		return "stack scan missed a root above sp"
	}
	if deadMarked {
		return "stack scan retained a root below sp"
	}
	return ""
}
