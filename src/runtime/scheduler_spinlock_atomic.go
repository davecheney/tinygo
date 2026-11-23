//go:build (scheduler.cores || (tinygo.riscv && virt)) && !rp2040 && !rp2350

package runtime

import "sync/atomic"

// Locks needed by the runtime. rp2 has its own, built on the SIO spinlocks.
var (
	printLock     spinLock
	schedulerLock schedulerSpinLock
	atomicsLock   spinLock
	futexLock     spinLock
)

// A spinlock built on an atomic compare-and-swap.
type spinLock struct {
	atomic.Uint32
}

func (l *spinLock) Lock() {
	// Try to replace 0 with 1. Once we succeed, the lock has been acquired.
	for !l.Uint32.CompareAndSwap(0, 1) {
		spinLoopWait()
	}
}

func (l *spinLock) Unlock() {
	// Safety check: the spinlock should have been locked.
	if schedulerAsserts && l.Uint32.Load() != 1 {
		runtimeFatal("unlock of unlocked spinlock")
	}

	// Unlock the lock. Simply write 0, because we already know it is locked.
	l.Uint32.Store(0)
}
