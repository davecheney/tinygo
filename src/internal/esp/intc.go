//go:build esp.intc

// Package esp drives the CPU interrupt controller of the Espressif RISC-V
// chips. The interrupt matrix routes a peripheral source to a CPU interrupt
// line, and this package controls the lines.
package esp

import (
	"device/riscv"
	"runtime/volatile"
	"unsafe"
)

// NumInterrupts is the highest CPU interrupt line. Line 0 is not used.
const NumInterrupts = 31

// MatrixOffset is added to a CPU interrupt line to get the value for an
// interrupt matrix map register.
const MatrixOffset = 0

// The C3 INTERRUPT_CORE0 CPU_INT block and the C6/H2 PLIC_MX block have the
// same layout. See ESP-IDF soc/esp32c3/register/soc/interrupt_core0_reg.h and
// soc/esp32c6/register/soc/plic_reg.h.
type intcRegs struct {
	ENABLE     volatile.Register32
	TYPE       volatile.Register32 // 1 is edge, 0 is level
	CLEAR      volatile.Register32
	EIP_STATUS volatile.Register32
	PRI        [32]volatile.Register32
	THRESH     volatile.Register32
}

var intc = (*intcRegs)(unsafe.Pointer(uintptr(intcBase)))

// Init masks all lines and sets the threshold.
func Init() {
	for i := 1; i <= NumInterrupts; i++ {
		intc.PRI[i].Set(0)
	}
	intc.THRESH.Set(5)
	if mieGate {
		riscv.MIE.Set(0xffffffff)
	}
}

// Enable sets line n to edge trigger with the given priority and enables it.
func Enable(n, priority uint32) {
	intc.ENABLE.SetBits(1 << n)
	intc.TYPE.SetBits(1 << n)
	intc.PRI[n].Set(priority)
	intc.CLEAR.SetBits(1 << n)
	intc.CLEAR.ClearBits(1 << n)
}

// Priority returns the priority of line n.
func Priority(n uint32) uint32 {
	return intc.PRI[n].Get()
}

// SetPriority sets the priority of line n. Priority 0 masks the line.
func SetPriority(n, priority uint32) {
	intc.PRI[n].Set(priority)
}

// Ack clears the pending state of line n.
func Ack(n uint32) {
	bit := uint32(1) << n
	if intc.TYPE.Get()&bit != 0 {
		intc.CLEAR.SetBits(bit)
	}
	intc.CLEAR.ClearBits(bit)
}

// Line returns the CPU interrupt line in mcause.
func Line(mcause uintptr) (uint32, bool) {
	n := uint32(mcause & 0x1f)
	return n, n != 0
}
