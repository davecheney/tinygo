//go:build arduino_leonardo

package runtime

import (
	"machine"
	_ "machine/usb/cdc"
)

func initUART() {}

func init() {
	// USB initialization allocates and must run after initHeap.
	machine.InitSerial()
}
