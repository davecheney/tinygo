package main

import "unsafe"

//go:linkname sliceGrow runtime.sliceGrow
func sliceGrow(buf unsafe.Pointer, oldLen, oldCap, newCap, elemSize uintptr, layout unsafe.Pointer) (unsafe.Pointer, uintptr, uintptr)

func appendByte(s []byte, b byte) []byte {
	return append(s, b)
}

func concat(a, b string) string {
	return a + b
}

func toBytes(s string) []byte {
	return []byte(s)
}

func grow(buf unsafe.Pointer, n uintptr) {
	sliceGrow(buf, n, n, n+1, 1, nil)
}
