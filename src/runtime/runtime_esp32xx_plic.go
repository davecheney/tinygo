//go:build esp32c6 || esp32h2

package runtime

import (
	"device/esp"
	"device/riscv"
	plic "internal/esp"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

func init() {
	machine.InitSerial()
}

// interruptInit initializes the interrupt controller.
func interruptInit() {
	riscv.DisableInterrupts()

	plic.Init()

	// Set the interrupt address.
	// Set MODE field to 1 - a vector base address.
	// Note that this address must be aligned to 256 bytes.
	riscv.MTVEC.Set((uintptr(unsafe.Pointer(&_vector_table))) | 1)

	// The ROM can leave MSTATUS.MIE cleared, so set it here.
	// Without it, no interrupt gets to the CPU.
	riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)
}

// CPU interrupt number used for the TIMG0 timer alarm.
const timerAlarmCPUInterrupt = 9

var interruptPending volatile.Register8

func signalInterrupt() {
	interruptPending.Set(1)
}

var timerAlarmInterrupt interrupt.Interrupt

// timerAlarmHandler clears the timer interrupt at the peripheral level.
func timerAlarmHandler(interrupt.Interrupt) {
	esp.TIMG0.INT_CLR_TIMERS.Set(1)
}

// initTimerInterrupt routes the TIMG0 timer 0 alarm interrupt to a CPU
// interrupt and registers a handler.
func initTimerInterrupt() {
	// Map the TIMG0 T0 peripheral interrupt to a CPU interrupt line.
	esp.INTERRUPT_CORE0.TG0_T0_INTR_MAP.Set(timerAlarmCPUInterrupt)

	// Enable T0 interrupt at the timer group level.
	esp.TIMG0.INT_ENA_TIMERS.SetBits(1)

	// Register the interrupt handler.
	timerAlarmInterrupt = interrupt.New(timerAlarmCPUInterrupt, timerAlarmHandler)
	timerAlarmInterrupt.Enable()

	// Raise the priority above the default that Enable sets for peripherals.
	plic.SetPriority(timerAlarmCPUInterrupt, 10)
}

// sleepTicks spins until the given number of ticks have elapsed, using the
// TIMG0 alarm interrupt to avoid busy-waiting for the entire duration.
func sleepTicks(d timeUnit) {
	machine.FlushSerial()
	target := ticks() + d
	for ticks() < target {
		interruptPending.Set(0)

		esp.TIMG0.T0ALARMLO.Set(uint32(target))
		esp.TIMG0.T0ALARMHI.Set(uint32(target >> 32))

		// Enable the alarm (auto-clears when alarm fires).
		esp.TIMG0.T0CONFIG.SetBits(esp.TIMG_T0CONFIG_ALARM_EN)

		for interruptPending.Get() == 0 {
			if ticks() >= target {
				return
			}
		}
	}
}

//go:extern _vector_table
var _vector_table [0]uintptr
