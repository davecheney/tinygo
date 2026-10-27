//go:build sam || nrf52840 || rp2040 || rp2350 || stm32f4 || stm32f7 || stm32h7 || mimxrt1062 || (esp32s3 && usbotg)

package machine

import "machine/usb/descriptor"

var (
	usbEndpointDescriptors [NumberOfUSBEndpoints]descriptor.Device

	//go:align 4
	udd_ep_control_cache_buffer [256]uint8

	//go:align 4
	udd_ep_in_cache_buffer [NumberOfUSBEndpoints][64]uint8

	//go:align 4
	udd_ep_out_cache_buffer [NumberOfUSBEndpoints][64]uint8

	usb_trans_buffer [255]uint8
)

func physicalUSBEndpoint(ep uint8, isIn bool) uint8 {
	return ep
}

func prepareUSBDescriptor(desc descriptor.Descriptor) descriptor.Descriptor {
	return desc
}
