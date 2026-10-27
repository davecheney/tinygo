//go:build arduino_leonardo

package machine

import "machine/usb/descriptor"

const NumberOfUSBEndpoints = 4

var (
	udd_ep_control_cache_buffer [128]uint8
	usb_trans_buffer            [128]uint8
)

func physicalUSBEndpoint(ep uint8, isIn bool) uint8 {
	if ep == 2 && isIn {
		return 3
	}
	return ep
}

func prepareUSBDescriptor(desc descriptor.Descriptor) descriptor.Descriptor {
	for i := 0; i+7 <= len(desc.Configuration); {
		length := int(desc.Configuration[i])
		if length < 2 || i+length > len(desc.Configuration) {
			break
		}
		if length >= 7 && desc.Configuration[i+1] == descriptor.TypeEndpoint {
			address := desc.Configuration[i+2]
			attributes := desc.Configuration[i+3]
			if address == 0x82 && attributes&0x03 == 0x02 {
				desc.Configuration[i+2] = 0x83
			}
		}
		i += length
	}
	return desc
}
