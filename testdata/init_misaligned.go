package main

// Interp must not read a pointer from bytes that start inside a pointer.

import "unsafe"

const ptrSize = unsafe.Sizeof(uintptr(0))

type word [ptrSize]byte

var x int
var pair = [2]*int{&x, &x}

var (
	loaded = loadMid()
	copied = copyMid()
)

// loadMid returns the bytes from the middle of the first pointer to the middle
// of the second pointer. A byte array load does not tell LLVM it is aligned.
//
//go:noinline
func loadMid() (v uintptr) {
	*(*word)(unsafe.Pointer(&v)) = *(*word)(unsafe.Add(unsafe.Pointer(&pair), ptrSize/2))
	return
}

// copyMid is the same as loadMid, but uses copy.
//
//go:noinline
func copyMid() (v uintptr) {
	src := (*[2 * ptrSize]byte)(unsafe.Pointer(&pair))
	copy((*word)(unsafe.Pointer(&v))[:], src[ptrSize/2:])
	return
}

func main() {
	println("load:", loaded == loadMid())
	println("copy:", copied == copyMid())
}
