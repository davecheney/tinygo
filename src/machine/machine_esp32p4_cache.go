//go:build esp32p4

package machine

import (
	"runtime/volatile"
	"unsafe"
)

// CACHE_SYNC_* registers describe the two-level cache write-back/invalidate
// operation used to make CPU-written memory visible to the DW-GDMA (which reads
// via AXI and bypasses the caches).
const (
	cacheBase              = 0x3ff10000
	cacheSyncCtrlOff       = 0x98
	cacheSyncMapOff        = 0x9c
	cacheSyncAddrOff       = 0xa0
	cacheSyncSizeOff       = 0xa4
	cacheSyncInvalidateEna = 1 << 0 // CACHE_INVALIDATE_ENA
	cacheSyncWriteBackEna  = 1 << 2 // CACHE_WRITEBACK_ENA
	cacheSyncDone          = 1 << 4 // CACHE_SYNC_DONE (set when operation finished)
	// CACHE_SYNC_MAP bits: [4]=L1-DCache, [5]=L2-Cache.
	cacheSyncMapL1DCache = 1 << 4
	cacheSyncMapL2Cache  = 1 << 5
	// L1 DCache line size on ESP32-P4 (used for cache-line alignment).
	cacheLineSize = 64
)

func cacheSyncReg(off uintptr) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer(cacheBase + off))
}

// cacheSyncOp performs a cache sync operation on the caches in cacheMap,
// matching ESP-IDF's Cache_WriteBack_Addr / Cache_Invalidate_Addr pattern:
//   - Aligns addr DOWN and size UP to cache-line boundaries
//   - Triggers the operation TWICE (hardware requirement on ESP32-P4)
func cacheSyncOp(addr uintptr, size uintptr, cacheMap, ena uint32) {
	if size == 0 {
		return
	}
	plus := addr & (cacheLineSize - 1)
	addr -= plus
	size += plus
	size = (size + cacheLineSize - 1) &^ (cacheLineSize - 1)

	ctrl := cacheSyncReg(cacheSyncCtrlOff)
	cacheSyncReg(cacheSyncMapOff).Set(cacheMap)
	cacheSyncReg(cacheSyncAddrOff).Set(uint32(addr))
	cacheSyncReg(cacheSyncSizeOff).Set(uint32(size))
	ctrl.Set(ena)
	for ctrl.Get()&cacheSyncDone == 0 {
	}
	ctrl.Set(ena)
	for ctrl.Get()&cacheSyncDone == 0 {
	}
}

// CacheInvalidateL1 discards L1-DCACHE-only lines for [addr, addr+size).
// Use for SRAM regions — SRAM is not cached by L2.  This prevents stale
// dirty .bss zero lines from evicting back over NC-alias-written data.
func CacheInvalidateL1(addr uintptr, size uintptr) {
	cacheSyncOp(addr, size, cacheSyncMapL1DCache, cacheSyncInvalidateEna)
}

// CacheWriteBack flushes the dirty L1-DCACHE and L2-CACHE lines covering the
// virtual address range [addr, addr+size) out to memory, so the DW-GDMA (which
// reads via AXI, bypassing the caches) observes the freshly written data.
// Mirrors ESP-IDF esp_cache_msync(..., DIR_C2M) for the P4 cache-sync block.
func CacheWriteBack(addr uintptr, size uintptr) {
	cacheSyncOp(addr, size, cacheSyncMapL1DCache|cacheSyncMapL2Cache, cacheSyncWriteBackEna)
}
