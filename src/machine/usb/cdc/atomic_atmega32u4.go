//go:build atmega32u4

package cdc

import (
	"runtime/interrupt"
	"runtime/volatile"
)

// CDC stores only packet lengths, a zero-length sentinel, and a 0/1 pump flag.
type cdcAtomicUint32 struct {
	value uint8
}

func (a *cdcAtomicUint32) Load() uint32 {
	value := volatile.LoadUint8(&a.value)
	if value == 0xff {
		return zlpInflight
	}
	return uint32(value)
}

func (a *cdcAtomicUint32) Store(value uint32) {
	if value == zlpInflight {
		volatile.StoreUint8(&a.value, 0xff)
		return
	}
	volatile.StoreUint8(&a.value, uint8(value))
}

func (a *cdcAtomicUint32) CompareAndSwap(old, new uint32) bool {
	state := interrupt.Disable()
	value := volatile.LoadUint8(&a.value)
	if uint32(value) == old {
		volatile.StoreUint8(&a.value, uint8(new))
		interrupt.Restore(state)
		return true
	}
	interrupt.Restore(state)
	return false
}
