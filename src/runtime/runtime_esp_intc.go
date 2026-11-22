//go:build esp.intc || esp.plic

package runtime

import (
	"device/esp"
	intc "internal/esp"
	"machine"
	"runtime/interrupt"
)

func init() {
	machine.InitSerial()
}

const timerAlarmCPUInterrupt = 9

func initTimerInterrupt() {
	timerAlarmMap().Set(timerAlarmCPUInterrupt + intc.MatrixOffset)
	esp.TIMG0.INT_ENA_TIMERS.SetBits(1)
	interrupt.New(timerAlarmCPUInterrupt, func(interrupt.Interrupt) {
		esp.TIMG0.INT_CLR_TIMERS.Set(1)
	}).Enable()

	// Above the default priority of peripherals.
	intc.SetPriority(timerAlarmCPUInterrupt, 10)
}

// sleepTicks uses the TIMG0 alarm so that it does not busy-wait for the full time.
func sleepTicks(d timeUnit) {
	machine.FlushSerial()
	target := ticks() + d
	for ticks() < target {
		interruptPending.Set(0)

		esp.TIMG0.T0ALARMLO.Set(uint32(target))
		esp.TIMG0.T0ALARMHI.Set(uint32(target >> 32))

		// The alarm enable clears when the alarm fires.
		esp.TIMG0.T0CONFIG.SetBits(esp.TIMG_T0CONFIG_ALARM_EN)

		// Wake on any interrupt, or on the timeout.
		for interruptPending.Get() == 0 {
			if ticks() >= target {
				return
			}
		}
	}
}
