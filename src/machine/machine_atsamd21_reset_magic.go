//go:build sam && atsamd21 && !arduino_m0

package machine

import (
	"device/arm"
	"unsafe"
)

// EnterBootloader should perform a system reset in preparation
// to switch to the bootloader to flash new firmware.
func EnterBootloader() {
	arm.DisableInterrupts()

	// Perform magic reset into bootloader, as mentioned in
	// https://github.com/arduino/ArduinoCore-samd/issues/197
	*(*uint32)(unsafe.Pointer(uintptr(0x20007FFC))) = resetMagicValue

	arm.SystemReset()
}
