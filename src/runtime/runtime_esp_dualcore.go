//go:build esp32 || esp32s3 || esp32p4

package runtime

import "device"

func spinLoopWait() {
	device.Asm("nop")
}
