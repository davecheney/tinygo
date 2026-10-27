//go:build arduino_leonardo

package machine

import (
	"device"
	"device/avr"
	"machine/usb"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const (
	usbEndpointSize = 64

	usbEndpointReleaseRX = 0x6B
	usbEndpointReleaseTX = 0x3A
)

var (
	usbEP0Transfer struct {
		data    []byte
		offset  uint16
		pending bool
	}

	usbINFlight          [NumberOfUSBEndpoints]bool
	usbSetAddressPending bool
	usbPendingAddress    uint8
)

func (dev *USBDevice) Configure(config UARTConfig) {
	if dev.initcomplete {
		return
	}

	avr.UDCON.Set(avr.UDCON_DETACH)
	avr.UHWCON.SetBits(avr.UHWCON_UVREGE)
	avr.USBCON.Set(avr.USBCON_USBE | avr.USBCON_FRZCLK)
	avr.PLLCSR.SetBits(avr.PLLCSR_PINDIV | avr.PLLCSR_PLLE)
	for !avr.PLLCSR.HasBits(avr.PLLCSR_PLOCK) {
	}
	avr.USBCON.ClearBits(avr.USBCON_FRZCLK)

	avr.UDIEN.Set(avr.UDIEN_EORSTE)
	interrupt.New(avr.IRQ_USB_GEN, handleUSBGenIRQ)
	interrupt.New(avr.IRQ_USB_COM, handleUSBEndpointIRQ)

	dev.Attach()
	dev.initcomplete = true
}

func (dev *USBDevice) Attach() {
	avr.UDCON.ClearBits(avr.UDCON_DETACH)
}

func (dev *USBDevice) Detach() {
	avr.UDCON.SetBits(avr.UDCON_DETACH)
}

func handleUSBGenIRQ(interrupt.Interrupt) {
	flags := avr.UDINT.Get()
	avr.UDINT.Set(flags &^ avr.UDINT_EORSTI)
	if flags&avr.UDINT_EORSTI == 0 {
		return
	}

	usbEP0Transfer.data = nil
	usbEP0Transfer.offset = 0
	usbEP0Transfer.pending = false
	usbSetAddressPending = false
	usbConfiguration = 0
	USBDev.InitEndpointComplete = false

	initEndpoint(usb.CONTROL_ENDPOINT, usb.ENDPOINT_TYPE_CONTROL)
	avr.UDADDR.Set(0)
}

func handleUSBEndpointIRQ(interrupt.Interrupt) {
	for ep := uint32(0); ep < NumberOfUSBEndpoints; ep++ {
		avr.UENUM.Set(uint8(ep))
		flags := avr.UEINTX.Get()

		if ep == usb.CONTROL_ENDPOINT && flags&avr.UEINTX_RXSTPI != 0 {
			handleUSBSetup()
			continue
		}

		if flags&avr.UEINTX_RXOUTI != 0 {
			handleEndpointOUT(ep)
		}

		if flags&avr.UEINTX_TXINI != 0 {
			handleEndpointIN(ep)
		}
	}
}

func handleUSBSetup() {
	var data [8]byte
	count := int(avr.UEBCLX.Get())
	for i := range data {
		if i < count {
			data[i] = avr.UEDATX.Get()
		}
	}
	for i := len(data); i < count; i++ {
		avr.UEDATX.Get()
	}

	avr.UEINTX.Set(^uint8(avr.UEINTX_RXSTPI | avr.UEINTX_RXOUTI | avr.UEINTX_TXINI))
	usbEP0Transfer.data = nil
	usbEP0Transfer.offset = 0
	usbEP0Transfer.pending = false
	usbSetAddressPending = false
	avr.UEIENX.ClearBits(avr.UEIENX_TXINE)

	if count != len(data) {
		stallEndpoint()
		return
	}

	setup := usb.NewSetup(data[:])
	if setup.BmRequestType&usb.REQUEST_DIRECTION == 0 {
		clearEndpointIN()
	}

	ok := false
	if setup.BmRequestType&usb.REQUEST_TYPE == usb.REQUEST_STANDARD {
		ok = handleStandardSetup(setup)
	} else if setup.WIndex < uint16(len(usbSetupHandler)) && usbSetupHandler[setup.WIndex] != nil {
		ok = usbSetupHandler[setup.WIndex](setup)
	}

	if !ok {
		stallEndpoint()
		return
	}
	if usbSetAddressPending {
		clearEndpointIN()
		avr.UEIENX.SetBits(avr.UEIENX_TXINE)
		return
	}
	clearEndpointIN()
}

func handleEndpointOUT(ep uint32) {
	count := int(avr.UEBCLX.Get())
	if ep == usb.CONTROL_ENDPOINT {
		for i := 0; i < count; i++ {
			avr.UEDATX.Get()
		}
		AckUsbOutTransfer(ep)
		return
	}
	if count > usbEndpointSize {
		for i := 0; i < count; i++ {
			avr.UEDATX.Get()
		}
		AckUsbOutTransfer(ep)
		return
	}

	buffer := udd_ep_out_cache_buffer[ep][:count]
	for i := range buffer {
		buffer[i] = avr.UEDATX.Get()
	}
	if usbRxHandler[ep] == nil || usbRxHandler[ep](buffer) {
		AckUsbOutTransfer(ep)
	}
}

func handleEndpointIN(ep uint32) {
	switch {
	case ep == usb.CONTROL_ENDPOINT && usbSetAddressPending:
		avr.UDADDR.Set(usbPendingAddress | avr.UDADDR_ADDEN)
		usbSetAddressPending = false
		avr.UEIENX.ClearBits(avr.UEIENX_TXINE)

	case ep == usb.CONTROL_ENDPOINT && usbEP0Transfer.pending:
		sendNextEP0Packet()

	case ep != usb.CONTROL_ENDPOINT && usbINFlight[ep]:
		usbINFlight[ep] = false
		avr.UEIENX.ClearBits(avr.UEIENX_TXINE)
		if usbTxHandler[ep] != nil {
			usbTxHandler[ep]()
		}
	}
}

func initEndpoint(ep, config uint32) {
	if ep >= NumberOfUSBEndpoints {
		return
	}

	avr.UENUM.Set(uint8(ep))
	avr.UECONX.Set(0)
	if config == usb.ENDPOINT_TYPE_DISABLE {
		return
	}

	direction := uint8(0)
	if config&usb.EndpointIn != 0 {
		direction = avr.UECFG0X_EPDIR
	}
	endpointType := uint8(config&0x03) << 6

	avr.UECONX.Set(avr.UECONX_EPEN)
	avr.UECFG0X.Set(endpointType | direction)
	avr.UECFG1X.Set(3<<4 | avr.UECFG1X_ALLOC)
	if !avr.UESTA0X.HasBits(avr.UESTA0X_CFGOK) {
		panic("machine: ATmega32U4 USB endpoint allocation failed")
	}
	if ep != usb.CONTROL_ENDPOINT {
		mask := uint8(1 << ep)
		avr.UERST.SetBits(mask)
		avr.UERST.ClearBits(mask)
	}

	switch {
	case ep == usb.CONTROL_ENDPOINT:
		avr.UEIENX.Set(avr.UEIENX_RXSTPE | avr.UEIENX_RXOUTE)
	case direction == avr.UECFG0X_EPDIR && config&0x03 == usb.ENDPOINT_TYPE_BULK:
		avr.UEIENX.Set(0)
	case direction == 0 && config&0x03 == usb.ENDPOINT_TYPE_BULK:
		avr.UEIENX.Set(avr.UEIENX_RXOUTE)
	default:
		avr.UEIENX.Set(0)
	}
}

func sendUSBPacket(ep uint32, data []byte) {
	if ep != usb.CONTROL_ENDPOINT {
		SendUSBInPacket(ep, data)
		return
	}

	usbEP0Transfer.data = data
	usbEP0Transfer.offset = 0
	usbEP0Transfer.pending = len(data) > usbEndpointSize
	sendNextEP0Packet()
	if usbEP0Transfer.pending {
		avr.UEIENX.SetBits(avr.UEIENX_TXINE)
	}
}

func sendNextEP0Packet() {
	remaining := len(usbEP0Transfer.data) - int(usbEP0Transfer.offset)
	count := min(remaining, usbEndpointSize)
	for i := 0; i < count; i++ {
		avr.UEDATX.Set(usbEP0Transfer.data[int(usbEP0Transfer.offset)+i])
	}
	usbEP0Transfer.offset += uint16(count)
	usbEP0Transfer.pending = int(usbEP0Transfer.offset) < len(usbEP0Transfer.data)
	if !usbEP0Transfer.pending {
		usbEP0Transfer.data = nil
		avr.UEIENX.ClearBits(avr.UEIENX_TXINE)
	}
}

func SendUSBInPacket(ep uint32, data []byte) bool {
	if ep == usb.CONTROL_ENDPOINT {
		sendUSBPacket(ep, data)
		return true
	}
	if ep >= NumberOfUSBEndpoints || len(data) > usbEndpointSize {
		return false
	}

	state := interrupt.Disable()
	avr.UENUM.Set(uint8(ep))
	if !avr.UEINTX.HasBits(avr.UEINTX_TXINI) {
		interrupt.Restore(state)
		return false
	}

	for _, b := range data {
		avr.UEDATX.Set(b)
	}
	usbINFlight[ep] = true
	avr.UEIENX.SetBits(avr.UEIENX_TXINE)
	avr.UEINTX.Set(usbEndpointReleaseTX)
	interrupt.Restore(state)
	return true
}

func ReceiveUSBControlPacket() ([cdcLineInfoSize]byte, error) {
	var data [cdcLineInfoSize]byte
	avr.UENUM.Set(usb.CONTROL_ENDPOINT)

	timeout := uint16(0xffff)
	for !avr.UEINTX.HasBits(avr.UEINTX_RXOUTI) && timeout > 0 {
		timeout--
	}
	if !avr.UEINTX.HasBits(avr.UEINTX_RXOUTI) {
		return data, ErrUSBReadTimeout
	}

	count := int(avr.UEBCLX.Get())
	if count != len(data) {
		for i := 0; i < count; i++ {
			avr.UEDATX.Get()
		}
		AckUsbOutTransfer(usb.CONTROL_ENDPOINT)
		return data, ErrUSBBytesRead
	}
	for i := range data {
		data[i] = avr.UEDATX.Get()
	}
	AckUsbOutTransfer(usb.CONTROL_ENDPOINT)
	return data, nil
}

func handleEndpointRx(ep uint32) []byte {
	count := int(avr.UEBCLX.Get())
	if count > usbEndpointSize {
		count = usbEndpointSize
	}
	buffer := udd_ep_out_cache_buffer[ep][:count]
	for i := range buffer {
		buffer[i] = avr.UEDATX.Get()
	}
	return buffer
}

func AckUsbOutTransfer(ep uint32) {
	avr.UENUM.Set(uint8(ep))
	avr.UEINTX.Set(usbEndpointReleaseRX)
}

func SendZlp() {
	sendUSBPacket(usb.CONTROL_ENDPOINT, nil)
}

func handleUSBSetAddress(setup usb.Setup) bool {
	if setup.WValueH != 0 || setup.WValueL > 127 || setup.WIndex != 0 || setup.WLength != 0 {
		return false
	}
	usbPendingAddress = setup.WValueL
	usbSetAddressPending = true
	return true
}

func clearEndpointIN() {
	avr.UEINTX.Set(^uint8(avr.UEINTX_TXINI))
}

func stallEndpoint() {
	avr.UECONX.Set(avr.UECONX_EPEN | avr.UECONX_STALLRQ)
}

func (dev *USBDevice) SetStallEPIn(ep uint32) {
	avr.UENUM.Set(uint8(ep & 0x7f))
	avr.UECONX.Set(avr.UECONX_EPEN | avr.UECONX_STALLRQ)
}

func (dev *USBDevice) SetStallEPOut(ep uint32) {
	avr.UENUM.Set(uint8(ep & 0x7f))
	avr.UECONX.Set(avr.UECONX_EPEN | avr.UECONX_STALLRQ)
}

func (dev *USBDevice) ClearStallEPIn(ep uint32) {
	avr.UENUM.Set(uint8(ep & 0x7f))
	avr.UECONX.Set(avr.UECONX_EPEN | avr.UECONX_STALLRQC | avr.UECONX_RSTDT)
}

func (dev *USBDevice) ClearStallEPOut(ep uint32) {
	avr.UENUM.Set(uint8(ep & 0x7f))
	avr.UECONX.Set(avr.UECONX_EPEN | avr.UECONX_STALLRQC | avr.UECONX_RSTDT)
}

func EnterBootloader() {
	state := interrupt.Disable()
	(*volatile.Register16)(unsafe.Pointer(uintptr(0x0800))).Set(0x7777)
	device.AsmFull("sts 0x60, {change}\n sts 0x60, {reset}", map[string]interface{}{
		"change": uint8(avr.WDTCSR_WDCE | avr.WDTCSR_WDE),
		"reset":  uint8(avr.WDTCSR_WDE | avr.WDTCSR_WDP1 | avr.WDTCSR_WDP0),
	})
	interrupt.Restore(state)
}
