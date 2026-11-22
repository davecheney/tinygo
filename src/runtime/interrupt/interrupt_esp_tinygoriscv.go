//go:build esp.intc || esp.clic

package interrupt

import (
	"device/riscv"
	"errors"
	"internal/esp"
)

//go:extern tinygo_saved_ra
var tinygo_saved_ra uintptr

const defaultPriority = 5

var errInterruptRange = errors.New("interrupt number out of range")

// interruptDepth counts nested handlers. In uses it, because on the CLIC
// mcause must keep the previous level for mret.
var interruptDepth uint32

// In returns whether the system is currently in an interrupt.
func In() bool {
	return interruptDepth != 0
}

// Enable enables CPU interrupt line i.num. The caller must first route the
// peripheral to this line with its interrupt matrix map register, with the
// value i.num + esp.MatrixOffset.
func (i Interrupt) Enable() error {
	if i.num < 1 || i.num > esp.NumInterrupts {
		return errInterruptRange
	}
	mask := riscv.DisableInterrupts()
	esp.Enable(uint32(i.num), defaultPriority)
	riscv.Asm("fence")
	riscv.EnableInterrupts(mask)
	return nil
}

//go:linkname callHandlers runtime/interrupt.callHandlers
func callHandlers(num int)

//go:linkname signalInterrupt runtime.signalInterrupt
func signalInterrupt()

//export handleInterrupt
func handleInterrupt() {
	mcause := riscv.MCAUSE.Get()
	n, ok := esp.Line(mcause)
	if mcause&(1<<31) == 0 || !ok {
		handleException(mcause)
		return
	}
	interruptDepth++

	// A nested interrupt can change these.
	mstatus := riscv.MSTATUS.Get()
	mepc := riscv.MEPC.Get()

	// Mask this line while MIE is set. A level interrupt stays asserted
	// until the handler clears it at the peripheral.
	priority := esp.Priority(n)
	esp.SetPriority(n, 0)
	riscv.Asm("fence")
	esp.Ack(n)

	riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)
	callHandler(int(n))
	signalInterrupt()
	riscv.MSTATUS.ClearBits(riscv.MSTATUS_MIE)

	esp.SetPriority(n, priority)
	riscv.Asm("fence")

	// The CLIC restores the interrupt level from mcause on mret.
	riscv.MCAUSE.Set(mcause)
	riscv.MSTATUS.Set(mstatus)
	riscv.MEPC.Set(mepc)
	interruptDepth--
}

func handleException(mcause uintptr) {
	println("*** Exception:     pc:", riscv.MEPC.Get())
	println("*** Exception:   code:", uint32(mcause&0x1f))
	println("*** Exception: mcause:", mcause)
	println("*** Exception:     ra:", tinygo_saved_ra)
	switch uint32(mcause & 0x1f) {
	case riscv.InstructionAccessFault:
		println("***    virtual address:", riscv.MTVAL.Get())
	case riscv.IllegalInstruction:
		println("***            opcode:", riscv.MTVAL.Get())
	case riscv.LoadAccessFault:
		println("***      read address:", riscv.MTVAL.Get())
	case riscv.StoreOrAMOAccessFault:
		println("***     write address:", riscv.MTVAL.Get())
	}
	for {
		riscv.Asm("wfi")
	}
}

// callHandlers needs a constant argument.
//
//go:inline
func callHandler(n int) {
	switch n {
	case 1:
		callHandlers(1)
	case 2:
		callHandlers(2)
	case 3:
		callHandlers(3)
	case 4:
		callHandlers(4)
	case 5:
		callHandlers(5)
	case 6:
		callHandlers(6)
	case 7:
		callHandlers(7)
	case 8:
		callHandlers(8)
	case 9:
		callHandlers(9)
	case 10:
		callHandlers(10)
	case 11:
		callHandlers(11)
	case 12:
		callHandlers(12)
	case 13:
		callHandlers(13)
	case 14:
		callHandlers(14)
	case 15:
		callHandlers(15)
	case 16:
		callHandlers(16)
	case 17:
		callHandlers(17)
	case 18:
		callHandlers(18)
	case 19:
		callHandlers(19)
	case 20:
		callHandlers(20)
	case 21:
		callHandlers(21)
	case 22:
		callHandlers(22)
	case 23:
		callHandlers(23)
	case 24:
		callHandlers(24)
	case 25:
		callHandlers(25)
	case 26:
		callHandlers(26)
	case 27:
		callHandlers(27)
	case 28:
		callHandlers(28)
	case 29:
		callHandlers(29)
	case 30:
		callHandlers(30)
	case 31:
		callHandlers(31)
	case 32:
		callHandlers(32)
	case 33:
		callHandlers(33)
	case 34:
		callHandlers(34)
	case 35:
		callHandlers(35)
	case 36:
		callHandlers(36)
	case 37:
		callHandlers(37)
	case 38:
		callHandlers(38)
	case 39:
		callHandlers(39)
	case 40:
		callHandlers(40)
	case 41:
		callHandlers(41)
	case 42:
		callHandlers(42)
	case 43:
		callHandlers(43)
	case 44:
		callHandlers(44)
	case 45:
		callHandlers(45)
	case 46:
		callHandlers(46)
	case 47:
		callHandlers(47)
	}
}
