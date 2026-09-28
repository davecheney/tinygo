// This example checks the PSRAM on RP2350 boards that declare it, such as
// the Pimoroni Presto. It prints the PSRAM base and size, then writes and
// verifies a pattern over a static buffer placed in PSRAM.
//
//	tinygo flash -target=presto -monitor examples/rp2350-psram
package main

import (
	"machine"
	"time"
	"unsafe"
)

// Two 480x480 RGB565 frames, placed in PSRAM and zeroed at startup.
//
//go:section .psram_bss
var frames [2][480 * 480]uint16

// Contents are undefined after reset.
//
//go:section .psram_noinit
var scratch [64 * 1024]uint32

func main() {
	time.Sleep(2 * time.Second)

	println("psram base:", hex(machine.PSRAMBase), "size:", machine.PSRAMSize())
	println("frames:", hex(uintptr(unsafe.Pointer(&frames))), "bytes:", unsafe.Sizeof(frames))
	println("scratch:", hex(uintptr(unsafe.Pointer(&scratch))), "bytes:", unsafe.Sizeof(scratch))

	ok := true
	for i := range frames {
		for j, v := range frames[i] {
			if v != 0 {
				println("FAIL frames not zeroed at", i, j, v)
				ok = false
				break
			}
		}
	}

	for pass := uint32(0); pass < 3; pass++ {
		start := time.Now()
		for i := range frames {
			for j := range frames[i] {
				frames[i][j] = uint16(uint32(j)*2654435761>>16) ^ uint16(i) ^ uint16(pass)
			}
		}
		for j := range scratch {
			scratch[j] = uint32(j)*2654435761 ^ pass
		}
		written := time.Since(start)

		errors := 0
		for i := range frames {
			for j, v := range frames[i] {
				if want := uint16(uint32(j)*2654435761>>16) ^ uint16(i) ^ uint16(pass); v != want {
					if errors < 4 {
						println("FAIL frames", i, j, "got", v, "want", want)
					}
					errors++
				}
			}
		}
		for j, v := range scratch {
			if want := uint32(j)*2654435761 ^ pass; v != want {
				if errors < 4 {
					println("FAIL scratch", j, "got", v, "want", want)
				}
				errors++
			}
		}
		println("pass", pass, "errors", errors, "write ms", written.Milliseconds(), "total ms", time.Since(start).Milliseconds())
		if errors != 0 {
			ok = false
		}
	}

	if ok {
		println("PSRAM OK")
	} else {
		println("PSRAM FAIL")
	}
	for {
		time.Sleep(time.Second)
	}
}

func hex(v uintptr) string {
	const digits = "0123456789abcdef"
	var buf [10]byte
	buf[0], buf[1] = '0', 'x'
	for i := 9; i >= 2; i-- {
		buf[i] = digits[v&0xf]
		v >>= 4
	}
	return string(buf[:])
}
