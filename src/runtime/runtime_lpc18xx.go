//go:build lpc18xx

package runtime

import (
	"device/arm"
	"device/nxp"
	"machine"
	"runtime/volatile"
)

var tickCounter volatile.Register64

//export Reset_Handler
func main() {
	preinit()

	// The Tessel bootloader jumps here with interrupts disabled.
	arm.Asm("cpsie i")

	configureClock()
	arm.SetupSystemTimer(machine.CPUFrequency() / 1000)
	machine.InitSerial()

	run()
	exit(0)
}

// configureClock runs the core from the IRC, whatever the bootloader left behind.
func configureClock() {
	nxp.CGU.BASE_M3_CLK.Set(nxp.CGU_BASE_M3_CLK_CLK_SEL_IRC_DEFAULT<<nxp.CGU_BASE_M3_CLK_CLK_SEL_Pos | nxp.CGU_BASE_M3_CLK_AUTOBLOCK)
}

func ticksToNanoseconds(value timeUnit) int64 {
	return int64(value * 1000_000)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000_000)
}

//go:linkname ticks runtime.ticks
func ticks() timeUnit {
	return timeUnit(tickCounter.Get())
}

func sleepTicks(d timeUnit) {
	if d <= 0 {
		return
	}
	stop := ticks() + d
	for ticks() < stop {
		waitForEvents()
	}
}

func waitForEvents() {
	arm.Asm("wfe")
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

func getchar() byte {
	for machine.Serial.Buffered() == 0 {
		Gosched()
	}
	v, _ := machine.Serial.ReadByte()
	return v
}

//export SysTick_Handler
func handleSysTick() {
	tickCounter.Set(tickCounter.Get() + 1)
}
