//go:build esp.intc || esp.clic

package runtime

import (
	"device/riscv"
	"internal/esp"
	"runtime/volatile"
	"unsafe"
)

func interruptInit() {
	riscv.DisableInterrupts()
	esp.Init()

	// MODE 1 is vectored. The table must be aligned to 256 bytes.
	riscv.MTVEC.Set(uintptr(unsafe.Pointer(&_vector_table)) | 1)

	// The ROM can leave MSTATUS.MIE clear.
	riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)
}

var interruptPending volatile.Register8

func signalInterrupt() {
	interruptPending.Set(1)
}

//go:extern _vector_table
var _vector_table [0]uintptr
