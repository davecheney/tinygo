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
			block.unmark()
		}
	}
	gcLock.Unlock()

	*(*unsafe.Pointer)(lastWord) = nil
	if !marked {
		return "stack scan missed a root in the last body word"
	}
	return ""
}
