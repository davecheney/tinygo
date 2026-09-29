//go:build gc.conservative || gc.precise

package runtime

import (
	"internal/gclayout"
	"unsafe"
)

// GCStackScanProbe checks that markCurrentGoroutineStack scans a stack object
// through its last body word before finishMark runs.
func GCStackScanProbe() string {
	const size = 64
	guard := allocManual(16)
	stack := alloc(size, gclayout.Conservative.AsPtr())
	target := alloc(16, gclayout.NoPtrs.AsPtr())
	blocks := (size + unsafe.Sizeof(objHeader{}) + bytesPerBlock - 1) / bytesPerBlock
	lastWord := unsafe.Add(stack, blocks*bytesPerBlock-unsafe.Sizeof(objHeader{})-unsafe.Sizeof(uintptr(0)))
	*(*unsafe.Pointer)(lastWord) = target

	gcLock.Lock()
	markCurrentGoroutineStack(uintptr(stack))
	marked := blockFromAddr(uintptr(target)).findHead().state() == blockStateMark
	finishMark()
	for block := gcBlock(0); block < endBlock; block++ {
		if block.state() == blockStateMark {
			header := (*objHeader)(unsafe.Add(block.pointer(), bytesPerBlock-unsafe.Sizeof(objHeader{})))
			if header.next != 1 {
				block.unmark()
			}
		}
	}
	manualMarked := blockFromAddr(uintptr(guard)).findHead().state() == blockStateMark
	gcLock.Unlock()

	*(*unsafe.Pointer)(lastWord) = nil
	free(guard)
	if !marked {
		return "stack scan missed a root in the last body word"
	}
	if !manualMarked {
		return "probe cleanup unmarked a manual allocation"
	}
	return ""
}
