//go:build esp32p4

package runtime

import (
	"machine"
)

// TIMG0 runs from APB (80MHz): divider 2 gives 40MHz ticks, 25ns each.
// T0CONFIG.USE_XTAL is left clear, so the source is APB, not XTAL.
const (
	timerDivider = 2
	nsPerTick    = 25
)

// This is the function called on startup after the flash (IROM/DROM) is
// initialized and the stack pointer has been set.
//
//export main
func main() {
	// NOTE: Watchdog timers and peripheral clocks (UART0, TIMG0) were already
	// configured in call_start_cpu0 assembly before jumping here. Assembly is
	// used to avoid TinyGo's BSS-global issue where uintptr address literals
	// are stored in BSS globals that the ROM zeroes, causing load faults.

	clearbss()

	// Configure interrupt handler.
	interruptInit()

	// Initialize main system timer used for time.Now.
	initTimer()

	// Initialize the heap, call main.main, etc.
	run()

	// Fallback: if main ever returns, hang the CPU.
	exit(0)
}

func init() {
	// The PSRAM driver allocates, so it runs after the heap is ready.
	startPSRAM()
	machine.InitSerial()
	checkPSRAM()
}

// sleepTicks spins until the given number of ticks have elapsed. It does not
// use the TIMG0 alarm interrupt.
func sleepTicks(d timeUnit) {
	machine.FlushSerial()
	target := ticks() + d

	for ticks() < target {
	}
}
