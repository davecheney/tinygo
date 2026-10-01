//go:build (gc.conservative || gc.precise) && runtime_clobberfree

package runtime

import (
	"internal/gclayout"
	"unsafe"
)

var gcSweepProbeKeep [1024]unsafe.Pointer

// GCClobberSweepProbe frees one object through sweep and checks that its
// storage is poisoned and later handed out zeroed.
func GCClobberSweepProbe() string {
	const size = 128
	ptr := alloc(size, gclayout.NoPtrs.AsPtr())
	guard := allocManual(size)
	*(*uint32)(guard) = 42
	for offset := uintptr(0); offset < size; offset += 4 {
		*(*uint32)(unsafe.Add(ptr, offset)) = 0x12345678
	}
	gcLock.Lock()
	for block := gcBlock(0); block < endBlock; block++ {
		if block.state() == blockStateHead {
			block.setState(blockStateMark)
		}
	}
	blockFromAddr(uintptr(ptr)).findHead().unmark()
	sweep()
	var failure string
	for offset := unsafe.Sizeof(freeRange{}); offset < size; offset += 4 {
		if *(*uint32)(unsafe.Add(ptr, offset)) != 0xdeadbeef {
			failure = "sweep did not poison reclaimed storage"
		}
	}
	if *(*uint32)(guard) != 42 {
		failure = "sweep poisoned a manual allocation"
	}
	gcLock.Unlock()
	if failure != "" {
		return failure
	}
	reused := false
	for i := range gcSweepProbeKeep {
		p := alloc(size, gclayout.NoPtrs.AsPtr())
		gcSweepProbeKeep[i] = p
		for offset := uintptr(0); offset < size; offset++ {
			if *(*byte)(unsafe.Add(p, offset)) != 0 {
				return "allocation did not zero reclaimed storage"
			}
		}
		if uintptr(p) < uintptr(ptr)+size && uintptr(ptr) < uintptr(p)+size {
			reused = true
			break
		}
	}
	gcSweepProbeKeep = [len(gcSweepProbeKeep)]unsafe.Pointer{}
	free(guard)
	if !reused {
		return "probe did not reuse the reclaimed range"
	}
	return ""
}
