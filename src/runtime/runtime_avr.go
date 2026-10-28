//go:build avr && !avrtiny

package runtime

import (
	"device/avr"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const BOARD = "arduino"

// Watchdog timer periods. These can be off by a large margin (hence the jump
// between 64ms and 125ms which is not an exact double), so don't rely on this
// for accurate time keeping.
const (
	WDT_PERIOD_16MS = iota
	WDT_PERIOD_32MS
	WDT_PERIOD_64MS
	WDT_PERIOD_125MS
	WDT_PERIOD_250MS
	WDT_PERIOD_500MS
	WDT_PERIOD_1S
	WDT_PERIOD_2S
)

const timerRecalibrateInterval = 6e7 // 1 minute

var nextTimerRecalibrate timeUnit

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte

//export main
func main() {
	preinit()
	initHardware()
	run()
	exit(0)
}

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint8)(ptr) = 0
		ptr = unsafe.Add(ptr, 1)
	}
}

func initHardware() {
	initUART()
	initMonotonicTimer()
	nextTimerRecalibrate = ticks() + timerRecalibrateInterval

	// Enable interrupts after initialization.
	avr.Asm("sei")
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

// Sleep this number of ticks of nanoseconds.
func sleepTicks(d timeUnit) {
	waitTill := ticks() + d
	for {
		now := ticks()
		if waitTill <= now {
			// done waiting
			return
		}
		if waitTill-now < timeUnit(nanosecondsInTick) {
			// The next tick interrupt may come too late, so busy wait.
			continue
		}
		// wait for interrupt
		avr.Asm("sleep")
		if hasScheduler {
			// The interrupt may have awoken a goroutine, so bail out early.
			return
		}
	}
}

func ticks() timeUnit {
	state := interrupt.Disable()
	// use volatile since ticksCount can be changed when running on multi-core boards.
	t := volatile.LoadUint64((*uint64)(unsafe.Pointer(&ticksCount)))
	count := avr.TCNT0.Get()
	// An overflow that is not yet handled means a full tick has passed.
	if avr.TIFR0.Get()&avr.TIFR0_TOV0 != 0 && count != 0xff {
		t += uint64(nanosecondsInTick)
	}
	interrupt.Restore(state)
	return timeUnit(t + uint64(uint32(count)*nanosecondsInCount))
}

func exit(code int) {
	abort()
}

func abort() {
	// Disable interrupts and go to sleep.
	// This can never be awoken except for reset, and is recognized as termination by simavr.
	avr.Asm("cli")
	for {
		avr.Asm("sleep")
	}
}

var ticksCount int64 // nanoseconds since start

// The initial values match the prescaler and mode set in initMonotonicTimer.
var nanosecondsInTick int64 = 64 * 256 * 1e9 / int64(machine.CPUFrequency()) // nanoseconds per each tick
var nanosecondsInCount = uint32(nanosecondsInTick / 256)                     // nanoseconds per each count of TCNT0, 0 if unknown

func initMonotonicTimer() {
	ticksCount = 0

	interrupt.New(avr.IRQ_TIMER0_OVF, func(i interrupt.Interrupt) {
		// use volatile
		ticks := volatile.LoadUint64((*uint64)(unsafe.Pointer(&ticksCount)))
		ticks += uint64(nanosecondsInTick)
		volatile.StoreUint64((*uint64)(unsafe.Pointer(&ticksCount)), ticks)
	})

	// initial initialization of the Timer0
	// - Mask interrupt
	avr.TIMSK0.ClearBits(avr.TIMSK0_TOIE0 | avr.TIMSK0_OCIE0A | avr.TIMSK0_OCIE0B)

	// - Write new values to TCNT2, OCR2x, and TCCR2x.
	avr.TCNT0.Set(0)
	avr.OCR0A.Set(0xff)
	// - Set mode 3
	avr.TCCR0A.Set(avr.TCCR0A_WGM00 | avr.TCCR0A_WGM01)
	// - Set prescaler 64, the same as the Arduino core, for one interrupt per 1024µs at 16MHz
	avr.TCCR0B.Set(avr.TCCR0B_CS01 | avr.TCCR0B_CS00)

	// - Unmask interrupt
	avr.TIMSK0.SetBits(avr.TIMSK0_TOIE0)
}

//go:linkname adjustMonotonicTimer machine.adjustMonotonicTimer
func adjustMonotonicTimer() {
	mask := interrupt.Disable()
	setTickLength()
	interrupt.Restore(mask)
}

func setTickLength() {
	nanosecondsInTick = currentNanosecondsInTick()
	nanosecondsInCount = 0
	switch avr.TCCR0A.Get() & 0x7 {
	case 0, 3:
		// TCNT0 counts up from 0x00 to 0xff.
		nanosecondsInCount = uint32(nanosecondsInTick / 256)
	}
}

func currentNanosecondsInTick() int64 {
	// this time depends on clk_IO, prescale, mode and OCR0A
	// assuming the clock source is CPU clock
	var prescaler int64
	switch avr.TCCR0B.Get() & 0x7 {
	case 2:
		prescaler = 8
	case 3:
		prescaler = 64
	case 4:
		prescaler = 256
	case 5:
		prescaler = 1024
	default:
		prescaler = 1
	}
	mode := avr.TCCR0A.Get() & 0x7

	/*
	 Mode WGM02 WGM01 WGM00 Timer/Counter       TOP  Update of  TOV Flag
	                        Mode of Operation        OCRx at    Set on
	 0    0     0     0     Normal              0xFF Immediate  MAX
	 1    0     0     1     PWM, Phase Correct  0xFF TOP        BOTTOM
	 2    0     1     0     CTC                 OCRA Immediate  MAX
	 3    0     1     1     Fast PWM            0xFF BOTTOM     MAX
	 5    1     0     1     PWM, Phase Correct  OCRA TOP        BOTTOM
	 7    1     1     1     Fast PWM            OCRA BOTTOM     TOP
	*/
	var counts int64
	switch mode {
	case 0, 3:
		// normal & fast PWM
		// TOV0 Interrupt when moving from MAX (0xff) to 0x00
		counts = 256
	case 1:
		// Phase Correct PWM
		// TOV0 Interrupt when moving from MAX (0xff) to 0x00
		counts = 256 * 2
	case 2, 7:
		// CTC & fast PWM
		// TOV0 Interrupt when moving from MAX (OCRA) to 0x00
		counts = int64(avr.OCR0A.Get())
	case 5:
		// Phase Correct PWM
		// TOV0 Interrupt when moving from MAX (OCRA) to 0x00
		counts = int64(avr.OCR0A.Get()) * 2
	default:
		counts = 1 // for unknown
	}
	return prescaler * counts * 1e9 / int64(machine.CPUFrequency())
}
