package main

import "unsafe"

//go:align 256
var byteA byte

//go:align 256
var byteB byte

var (
	addrA = uintptr(unsafe.Pointer(&byteA))
	addrB = uintptr(unsafe.Pointer(&byteB))
)

// Interp runs this call at compile time. The value of a byte of an address is
// not known until the program is linked, so interp must not compute it.
var initEqual = compareAddressBytes()

//go:noinline
func compareAddressBytes() bool {
	return *(*byte)(unsafe.Pointer(&addrA)) == *(*byte)(unsafe.Pointer(&addrB))
}

func main() {
	// The result of the comparison depends on the target.
	println("address bytes:", initEqual == compareAddressBytes())
}
