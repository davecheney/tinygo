package main

import (
	"runtime"
	"unsafe"
)

//go:linkname alloc runtime.alloc
func alloc(uintptr, unsafe.Pointer) unsafe.Pointer

//go:linkname allocManual runtime.allocManual
func allocManual(uintptr) unsafe.Pointer

//go:linkname free runtime.free
func free(unsafe.Pointer)

var roots [8]*[16]uint32
var zeroRoot unsafe.Pointer

func main() {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	zeroRoot = alloc(0, unsafe.Pointer(uintptr(3)))
	runtime.ReadMemStats(&after)
	if after.NumGC != before.NumGC {
		panic("zero allocation collected")
	}
	for i := range roots {
		runtime.ReadMemStats(&before)
		roots[i] = (*[16]uint32)(alloc(64, unsafe.Pointer(uintptr(3))))
		runtime.ReadMemStats(&after)
		if after.NumGC-before.NumGC != 1 {
			println("collections", after.NumGC-before.NumGC)
			panic("expected one collection per allocation")
		}
		for _, value := range roots[i] {
			if value != 0 {
				panic("allocation not zeroed")
			}
		}
		roots[i][0] = uint32(i + 1)
	}
	runtime.ReadMemStats(&before)
	p := allocManual(64)
	runtime.ReadMemStats(&after)
	if after.NumGC-before.NumGC != 1 {
		panic("manual allocation must collect once")
	}
	*(*uint32)(p) = 42
	runtime.GC()
	if *(*uint32)(p) != 42 {
		panic("manual allocation swept")
	}
	free(p)
	for i, p := range roots {
		if p[0] != uint32(i+1) {
			panic("live allocation swept")
		}
	}
	println("ok")
}
