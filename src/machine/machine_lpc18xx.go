//go:build lpc18xx

package machine

import (
	"device/nxp"
	"runtime/volatile"
	"unsafe"
)

const deviceName = nxp.Device

// CPUFrequency returns the frequency of the ARM core clock.
// The runtime runs the core from the 12 MHz IRC.
func CPUFrequency() uint32 {
	return 12 * MHz
}

const (
	PinOutput PinMode = iota
	PinInput
	PinInputPullup
	PinInputPulldown
)

// A Pin is a GPIO port and bit, numbered port*32 + bit.
// This matches the GPIO byte pin registers, see UM10430 GPIO chapter.

// SCU pin configuration bits, see UM10430 SFS register description.
const (
	scuEPD  = 1 << 3 // enable pull-down
	scuEPUN = 1 << 4 // disable pull-up
	scuEZI  = 1 << 6 // enable input buffer
)

// pinPad describes the SCU pad and function that route a pad to a GPIO.
type pinPad struct {
	pin   Pin
	group uint8
	index uint8
	fn    uint8
}

func (p Pin) pad() (pinPad, bool) {
	for _, pad := range pinPads {
		if pad.pin == p {
			return pad, true
		}
	}
	return pinPad{}, false
}

func (p Pin) dir() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&nxp.GPIO_PORT.DIR0), uintptr(p/32)*4))
}

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	if p == NoPin {
		return
	}
	pad, ok := p.pad()
	if !ok {
		return
	}
	cfg := uint32(pad.fn)
	switch config.Mode {
	case PinOutput:
		cfg |= scuEPUN
	case PinInput:
		cfg |= scuEPUN | scuEZI
	case PinInputPullup:
		cfg |= scuEZI
	case PinInputPulldown:
		cfg |= scuEPUN | scuEPD | scuEZI
	}
	sfs := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(nxp.SCU), uintptr(pad.group)*0x80+uintptr(pad.index)*4))
	sfs.Set(cfg)

	bit := uint32(1) << (p % 32)
	if config.Mode == PinOutput {
		p.dir().SetBits(bit)
	} else {
		p.dir().ClearBits(bit)
	}
}

// Set the pin to high or low.
func (p Pin) Set(high bool) {
	if high {
		nxp.GPIO_PORT.B[p].Set(1)
	} else {
		nxp.GPIO_PORT.B[p].Set(0)
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	return nxp.GPIO_PORT.B[p].Get() != 0
}
