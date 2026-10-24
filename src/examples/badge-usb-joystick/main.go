// Exposes the TinyGo Conf 2026 badge joystick and buttons as a USB HID joystick.
// Build with -tags=usbotg so the ESP32-S3 USB-OTG controller is used.
package main

import (
	"machine"
	"machine/usb/hid/joystick"
	"time"
)

// Pins from firmware/README.en.md in tinygo-keeb/tinygo-conf-2026-badge.
const (
	pinX   = machine.GPIO7
	pinY   = machine.GPIO6
	pinBtn = machine.GPIO15
	pinSW1 = machine.GPIO13
	pinSW2 = machine.GPIO14
)

const (
	// Raw ADC counts from center to full deflection, measured by the badge's joyraw example.
	travel   = 16000
	deadZone = 800
)

var js = joystick.Port()

func main() {
	machine.InitADC()
	x := machine.ADC{Pin: pinX}
	y := machine.ADC{Pin: pinY}
	x.Configure(machine.ADCConfig{})
	y.Configure(machine.ADCConfig{})

	buttons := []machine.Pin{pinBtn, pinSW1, pinSW2}
	for _, b := range buttons {
		b.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}

	cx, cy := calibrate(x, y)

	ticker := time.NewTicker(10 * time.Millisecond)
	for range ticker.C {
		// X reads lower to the right and Y reads higher upwards, HID wants right and down positive.
		js.SetAxis(0, scale(int(cx)-int(x.Get())))
		js.SetAxis(1, scale(int(cy)-int(y.Get())))
		for i, b := range buttons {
			js.SetButton(i, !b.Get())
		}
		js.SendState()
	}
}

func calibrate(x, y machine.ADC) (cx, cy uint16) {
	const n = 64
	var sx, sy uint32
	for range n {
		sx += uint32(x.Get())
		sy += uint32(y.Get())
		time.Sleep(5 * time.Millisecond)
	}
	return uint16(sx / n), uint16(sy / n)
}

func scale(d int) int {
	if d > -deadZone && d < deadZone {
		return 0
	}
	v := d * 32767 / travel
	return max(-32767, min(32767, v))
}
