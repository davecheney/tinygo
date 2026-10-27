//go:build baremetal

package cdc

import (
	"errors"
	"machine"
	"machine/usb"
	_ "unsafe"
)

var ErrBufferEmpty = errors.New("USB-CDC buffer empty")

const cdcLineInfoSize = 7

type cdcLineInfo struct {
	dwDTERate   uint32
	bCharFormat uint8
	bParityType uint8
	bDataBits   uint8
	lineState   uint8
}

var (
	// USB is a USB CDC interface.
	USB *USBCDC

	usbLineInfo = cdcLineInfo{115200, 0x00, 0x00, 0x08, 0x00}
)

func (usbcdc *USBCDC) DTR() bool {
	return (usbLineInfo.lineState & usb_CDC_LINESTATE_DTR) > 0
}

func (usbcdc *USBCDC) RTS() bool {
	return (usbLineInfo.lineState & usb_CDC_LINESTATE_RTS) > 0
}

var cdcSetupBuff [cdcLineInfoSize]byte

func cdcSetup(setup usb.Setup) bool {
	if setup.BmRequestType == usb_REQUEST_DEVICETOHOST_CLASS_INTERFACE {
		if setup.BRequest == usb_CDC_GET_LINE_CODING {
			cdcSetupBuff[0] = byte(usbLineInfo.dwDTERate)
			cdcSetupBuff[1] = byte(usbLineInfo.dwDTERate >> 8)
			cdcSetupBuff[2] = byte(usbLineInfo.dwDTERate >> 16)
			cdcSetupBuff[3] = byte(usbLineInfo.dwDTERate >> 24)
			cdcSetupBuff[4] = byte(usbLineInfo.bCharFormat)
			cdcSetupBuff[5] = byte(usbLineInfo.bParityType)
			cdcSetupBuff[6] = byte(usbLineInfo.bDataBits)

			machine.SendUSBInPacket(0, cdcSetupBuff[:])
			return true
		}
	}

	if setup.BmRequestType == usb_REQUEST_HOSTTODEVICE_CLASS_INTERFACE {
		if setup.BRequest == usb_CDC_SET_LINE_CODING {
			b, err := machine.ReceiveUSBControlPacket()
			if err != nil {
				return false
			}

			usbLineInfo.dwDTERate = uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
			usbLineInfo.bCharFormat = b[4]
			usbLineInfo.bParityType = b[5]
			usbLineInfo.bDataBits = b[6]
		}

		if setup.BRequest == usb_CDC_SET_CONTROL_LINE_STATE {
			usbLineInfo.lineState = setup.WValueL
		}

		if setup.BRequest == usb_CDC_SET_LINE_CODING || setup.BRequest == usb_CDC_SET_CONTROL_LINE_STATE {
			// auto-reset into the bootloader
			if usbLineInfo.dwDTERate == 1200 && usbLineInfo.lineState&usb_CDC_LINESTATE_DTR == 0 {
				machine.EnterBootloader()
			} else {
				// TODO: cancel any reset
			}
			machine.SendZlp()
		}

		if setup.BRequest == usb_CDC_SEND_BREAK {
			// TODO: something with this value?
			// breakValue = ((uint16_t)setup.wValueH << 8) | setup.wValueL;
			// return false;
			machine.SendZlp()
		}
		return true
	}
	return false
}

//go:linkname gosched runtime.Gosched
func gosched()
