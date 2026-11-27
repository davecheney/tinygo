//go:build esp32c6 || esp32h2

package runtime

import (
	"device/riscv"
	"machine"
	"runtime/volatile"
	"unsafe"
)

func init() {
	machine.InitSerial()
}

// interruptInit initializes the interrupt controller.
func interruptInit() {
	riscv.DisableInterrupts()

	// Reset all interrupt priorities to zero in the PLIC.
	for i := 1; i < 32; i++ {
		plic.MXINT_PRI[i].Set(0)
	}

	// Default threshold for interrupts is 5.
	plic.MXINT_THRESH.Set(5)

	// Set the interrupt address.
	// Set MODE field to 1 - a vector base address.
	// Note that this address must be aligned to 256 bytes.
	riscv.MTVEC.Set((uintptr(unsafe.Pointer(&_vector_table))) | 1)

	// The mie CSR gates each PLIC interrupt line, so enable all lines.
	// esp-hal does the same with `csrw mie, 0xffffffff`.
	riscv.MIE.Set(0xffffffff)

	// The ROM can leave MSTATUS.MIE cleared, so set it here.
	// Without it, no interrupt gets to the CPU.
	riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)
}

// sleepTicks spins until the given number of ticks have elapsed.
func sleepTicks(d timeUnit) {
	machine.FlushSerial()
	target := ticks() + d
	for ticks() < target {
	}
}

//go:extern _vector_table
var _vector_table [0]uintptr

// plicType maps the machine-mode registers of the PLIC.
// See ESP-IDF components/soc/esp32c6/register/soc/plic_reg.h.
type plicType struct {
	MXINT_ENABLE     volatile.Register32     // 0x00 bit N enables CPU interrupt line N
	MXINT_TYPE       volatile.Register32     // 0x04 bit N: 1=edge, 0=level
	MXINT_CLEAR      volatile.Register32     // 0x08 edge acknowledge
	MXINT_EIP_STATUS volatile.Register32     // 0x0C pending status (read-only)
	MXINT_PRI        [32]volatile.Register32 // 0x10..0x8C per-line priority (4 bits)
	MXINT_THRESH     volatile.Register32     // 0x90 priority threshold (8 bits)
}

// plic points at the PLIC machine-mode register block (DR_REG_PLIC_MX_BASE).
var plic = (*plicType)(unsafe.Pointer(uintptr(0x20001000)))
