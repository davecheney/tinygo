package main

// Each init below changes some bytes of a pointer. Main does the same change at
// runtime and compares the results.

import "unsafe"

var buf [512]byte

var one = uint16(1)

// lowByte returns the offset of the low byte of a pointer.
func lowByte() uintptr {
	if *(*byte)(unsafe.Pointer(&one)) == 1 {
		return 0
	}
	return unsafe.Sizeof(uintptr(0)) - 1
}

// Change the low byte only, so that the pointer stays in buf.
var bytePtr = &buf[256]

func init() {
	*(*byte)(unsafe.Add(unsafe.Pointer(&bytePtr), lowByte())) = 0x5a
}

var byteAddr = uintptr(unsafe.Pointer(&buf[1]))

func init() {
	*(*byte)(unsafe.Pointer(&byteAddr)) = 0x5a
}

// This store starts in the first pointer and ends in the second pointer.
var pair = [2]uintptr{uintptr(unsafe.Pointer(&buf[2])), uintptr(unsafe.Pointer(&buf[3]))}

func init() {
	*(*[2]byte)(unsafe.Add(unsafe.Pointer(&pair), unsafe.Sizeof(uintptr(0))-1)) = [2]byte{0x5a, 0x5a}
}

var copied = uintptr(unsafe.Pointer(&buf[4]))

func init() {
	src := [2]byte{0x5a, 0x5a}
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&copied)), unsafe.Sizeof(copied))[1:], src[:])
}

// write writes b at offset off in the memory of v.
func write(v *[2]uintptr, off uintptr, b ...byte) {
	copy(unsafe.Slice((*byte)(unsafe.Pointer(v)), unsafe.Sizeof(*v))[off:], b)
}

func main() {
	want := [2]uintptr{uintptr(unsafe.Pointer(&buf[256]))}
	write(&want, lowByte(), 0x5a)
	println("byte store in pointer:", uintptr(unsafe.Pointer(bytePtr)) == want[0])

	want = [2]uintptr{uintptr(unsafe.Pointer(&buf[1]))}
	write(&want, 0, 0x5a)
	println("byte store in uintptr:", byteAddr == want[0])

	want = [2]uintptr{uintptr(unsafe.Pointer(&buf[2])), uintptr(unsafe.Pointer(&buf[3]))}
	write(&want, unsafe.Sizeof(uintptr(0))-1, 0x5a, 0x5a)
	println("store over two pointers:", pair == want)

	want = [2]uintptr{uintptr(unsafe.Pointer(&buf[4]))}
	write(&want, 1, 0x5a, 0x5a)
	println("copy into pointer:", copied == want[0])
}
