//go:build !atmega32u4

package cdc

import "sync/atomic"

type cdcAtomicUint32 struct {
	value atomic.Uint32
}

func (a *cdcAtomicUint32) Load() uint32 {
	return a.value.Load()
}

func (a *cdcAtomicUint32) Store(value uint32) {
	a.value.Store(value)
}

func (a *cdcAtomicUint32) CompareAndSwap(old, new uint32) bool {
	return a.value.CompareAndSwap(old, new)
}
