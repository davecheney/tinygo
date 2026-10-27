//go:build avr && (atmega || atmega32u4) && !arduino_leonardo

package runtime

import "machine"

func initUART() {
	machine.InitSerial()
}
