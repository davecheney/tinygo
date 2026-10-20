//go:build sam && atsamd21 && arduino_m0

package machine

import (
	"device/arm"
	"device/sam"
	"unsafe"
)

//go:extern __isr_vector
var appVectors [0]uint32

// EnterBootloader erases the first row of the application and resets. The
// bootloader stays resident when it finds no application, see
// https://github.com/arduino/ArduinoCore-samd/blob/master/cores/arduino/Reset.cpp
func EnterBootloader() {
	arm.DisableInterrupts()
	eraseRowAndReset(uint32(uintptr(unsafe.Pointer(&appVectors))))
}

// eraseRowAndReset runs from RAM because the row it erases also holds code.
//
//go:section .ramfuncs
func eraseRowAndReset(addr uint32) {
	for !sam.NVMCTRL.INTFLAG.HasBits(sam.NVMCTRL_INTFLAG_READY) {
	}
	sam.NVMCTRL.STATUS.Set(sam.NVMCTRL_STATUS_LOAD | sam.NVMCTRL_STATUS_PROGE | sam.NVMCTRL_STATUS_LOCKE | sam.NVMCTRL_STATUS_NVME)
	sam.NVMCTRL.ADDR.Set(addr >> 1)
	sam.NVMCTRL.CTRLA.Set(sam.NVMCTRL_CTRLA_CMD_ER | sam.NVMCTRL_CTRLA_CMDEX_KEY<<sam.NVMCTRL_CTRLA_CMDEX_Pos)
	for !sam.NVMCTRL.INTFLAG.HasBits(sam.NVMCTRL_INTFLAG_READY) {
	}
	arm.SCB.AIRCR.Set(0x5FA<<arm.SCB_AIRCR_VECTKEY_Pos | arm.SCB_AIRCR_SYSRESETREQ_Msk)
	for {
	}
}
