//go:build baremetal && atmega32u4

package cdc

import (
	"machine"
	_ "unsafe"
)

// USBCDC accesses the controller FIFOs without software buffering.
type USBCDC struct{}

func (usbcdc *USBCDC) Read(data []byte) (int, error) {
	return machine.ReadUSBOutPacket(cdcEndpointOut, data), nil
}

func (usbcdc *USBCDC) ReadByte() (byte, error) {
	var data [1]byte
	if machine.ReadUSBOutPacket(cdcEndpointOut, data[:]) == 0 {
		return 0, ErrBufferEmpty
	}
	return data[0], nil
}

func (usbcdc *USBCDC) Buffered() int {
	return machine.USBOutBuffered(cdcEndpointOut)
}

func (usbcdc *USBCDC) Configure(config machine.UARTConfig) error {
	return nil
}

func (usbcdc *USBCDC) Flush() {
	machine.FlushUSBInFIFO(cdcEndpointIn)
}

func (usbcdc *USBCDC) Write(data []byte) (n int, err error) {
	timeout := 250
	for len(data) > 0 {
		if !machine.USBDev.InitEndpointComplete || usbLineInfo.lineState == 0 {
			return n, machine.ErrUSBBytesWritten
		}
		written := machine.WriteUSBInFIFO(cdcEndpointIn, data)
		n += written
		data = data[written:]
		if written == 0 {
			timeout--
			if timeout == 0 {
				return n, machine.ErrUSBWriteTimeout
			}
			sleep(1e6)
		}
	}
	return n, nil
}

func (usbcdc *USBCDC) WriteByte(c byte) error {
	data := [1]byte{c}
	_, err := usbcdc.Write(data[:])
	return err
}

func reset() {
	usbLineInfo.lineState = 0
}

func EnableUSBCDC() {
	machine.USBCDC = New()
	machine.EnableCDC(nil, nil, cdcSetup)
}

//go:linkname sleep time.Sleep
func sleep(duration int64)
