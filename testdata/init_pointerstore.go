package main

import "unsafe"

var x int

//go:noinline
func byteOf(p *uintptr) byte {
	return *(*byte)(unsafe.Pointer(p))
}

// Each case has its own address and init function. Thus interp can run the
// other cases at compile time when one case runs at runtime.

// Store a byte of an address in a global.
var p1 = uintptr(unsafe.Pointer(&x))
var firstByte byte

func init() {
	firstByte = byteOf(&p1)
}

// Store a byte of an address in a struct field.
type pair struct {
	b byte
	n int8
}

var p2 = uintptr(unsafe.Pointer(&x))
var field pair

func init() {
	field = pair{b: byteOf(&p2), n: 5}
}

// Copy a byte of an address into an array.
var p3 = uintptr(unsafe.Pointer(&x))
var array [2]byte

func init() {
	copy(array[:], (*[1]byte)(unsafe.Pointer(&p3))[:])
	array[1] = 6
}

// Store a byte of an address in a heap object.
var p4 = uintptr(unsafe.Pointer(&x))
var heap *pair

func init() {
	heap = &pair{b: byteOf(&p4), n: 7}
}

// Store a whole address.
var p5 = uintptr(unsafe.Pointer(&x))
var whole uintptr

func init() {
	whole = *(*uintptr)(unsafe.Pointer(&p5))
}

func main() {
	println("first byte:", firstByte == byteOf(&p1))
	println("struct field:", field.b == byteOf(&p2), field.n)
	println("array:", array[0] == byteOf(&p3), array[1])
	println("heap object:", heap.b == byteOf(&p4), heap.n)
	println("whole pointer:", whole == uintptr(unsafe.Pointer(&x)))
}
