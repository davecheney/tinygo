//go:build rp2350

package machine

// PSRAM on QMI chip select 1 (XIP window 1).
//
// Boards with PSRAM opt in with the rp2350_psram build tag, a psramCSPin
// constant in the board file, and --defsym=__psram_size=<bytes> in the target.
// Startup then configures the PSRAM and zeroes the .psram_bss section.
//
// Place pointer free, zero initialized package level variables in PSRAM with
// //go:section .psram_bss (zeroed at startup) or //go:section .psram_noinit
// (left as is). The GC does not scan these sections and the heap stays in SRAM.
//
// See RP2350 datasheet section 4.4 (XIP) and section 12.14 (QMI).
// https://datasheets.raspberrypi.com/rp2350/rp2350-datasheet.pdf

const (
	// PSRAMBase is the cached address of QMI window 1. Accesses through it go
	// through the shared XIP cache, so they are coherent for both cores and DMA.
	PSRAMBase = 0x11000000

	// PSRAMUncachedBase is the uncached alias of QMI window 1. It bypasses the
	// XIP cache and can observe stale data while cached writes are pending.
	PSRAMUncachedBase = 0x15000000
)

// psramSize is the usable PSRAM size in bytes, set during startup.
var psramSize uintptr

// PSRAMSize returns the size in bytes of the PSRAM found at startup. It is 0
// when the target does not declare PSRAM or the chip did not respond.
func PSRAMSize() int {
	return int(psramSize)
}

// psramTiming returns the QMI M1_TIMING value for an APS6404L style QSPI PSRAM
// at the given system clock. The limits are 109 MHz SCK at 3.3 V, 8 us maximum
// CS low time and 50 ns minimum CS high time from the APS6404L datasheet.
// Derived from sfe_psram.c, Copyright (c) 2024 SparkFun Electronics, MIT
// license. https://github.com/sparkfun/sparkfun-pico
func psramTiming(sysHz uint32) uint32 {
	const (
		maxSCKHz      = 109_000_000
		fsPerSecond   = 1_000_000_000_000_000
		maxSelectFS64 = 125_000_000 // 8 us in units of 64 clock cycles
		minDeselectFS = 50_000_000  // 50 ns

		cooldownPos    = 30
		pagebreakPos   = 28
		pagebreak1024  = 2
		selectHoldPos  = 23
		maxSelectPos   = 17
		maxSelectMax   = 0x3f
		minDeselectPos = 12
		minDeselectMax = 0x1f
		rxDelayPos     = 8
		clkDivMax      = 0xff
	)
	clkDiv := (sysHz + maxSCKHz - 1) / maxSCKHz
	if clkDiv > clkDivMax {
		clkDiv = clkDivMax
	}
	fsPerCycle := uint32(fsPerSecond / uint64(sysHz))
	maxSelect := maxSelectFS64 / fsPerCycle
	if maxSelect > maxSelectMax {
		maxSelect = maxSelectMax
	}
	minDeselect := (minDeselectFS + fsPerCycle - 1) / fsPerCycle
	if minDeselect > minDeselectMax {
		minDeselect = minDeselectMax
	}
	return 1<<cooldownPos |
		pagebreak1024<<pagebreakPos |
		3<<selectHoldPos |
		maxSelect<<maxSelectPos |
		minDeselect<<minDeselectPos |
		1<<rxDelayPos |
		clkDiv
}

// psramSizeFromEID decodes the APS6404L read ID EID byte into a size in bytes.
// Derived from sfe_psram.c, Copyright (c) 2024 SparkFun Electronics, MIT
// license. https://github.com/sparkfun/sparkfun-pico
func psramSizeFromEID(eid uint8) uintptr {
	const mib = 1024 * 1024
	switch {
	case eid == 0x26 || eid>>5 == 2:
		return 8 * mib
	case eid>>5 == 0:
		return 2 * mib
	case eid>>5 == 1:
		return 4 * mib
	default:
		return mib
	}
}
