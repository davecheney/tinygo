//go:build cortexm && qemu

package runtime

// This file implements the Arm MPS2 AN385 Cortex-M3 board as implemented by
// QEMU.

import (
	"device/arm"
	"runtime/volatile"
	"unsafe"
)

var timestamp timeUnit

//export Reset_Handler
func main() {
	preinit()
	initUART()
	run()

	// Signal successful exit.
	exit(0)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

func sleepTicks(d timeUnit) {
	// TODO: actually sleep here for the given time.
	timestamp += d
}

func ticks() timeUnit {
	return timestamp
}

// CMSDK APB UART0, see Arm DDI 0479C section 4.3 and QEMU hw/arm/mps2.c.
// QEMU hw/char/cmsdk-apb-uart.c loses bytes written while TX is off or full.
type cmsdkUART struct {
	data    volatile.Register32
	state   volatile.Register32
	ctrl    volatile.Register32
	intr    volatile.Register32
	bauddiv volatile.Register32
}

var uart0 = (*cmsdkUART)(unsafe.Pointer(uintptr(0x40004000)))

func initUART() {
	uart0.bauddiv.Set(16)
	uart0.ctrl.Set(1) // TX enable
}

//go:noinline
func putchar(c byte) {
	for uart0.state.Get()&1 != 0 { // TX buffer full
	}
	uart0.data.Set(uint32(c))
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}

func waitForEvents() {
	arm.Asm("wfe")
}

func abort() {
	exit(1)
}

func exit(code int) {
	// Exit QEMU.
	if code == 0 {
		arm.SemihostingCall(arm.SemihostingReportException, arm.SemihostingApplicationExit)
	} else {
		arm.SemihostingCall(arm.SemihostingReportException, arm.SemihostingRunTimeErrorUnknown)
	}

	// Lock up forever (should be unreachable).
	for {
		arm.Asm("wfi")
	}
}
