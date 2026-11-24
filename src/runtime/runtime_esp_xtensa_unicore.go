//go:build (esp32 || esp32s3) && !scheduler.cores

package runtime

import "runtime/interrupt"

//go:inline
func lockTicks() interrupt.State {
	return 0
}

//go:inline
func unlockTicks(interrupt.State) {
}

//go:inline
func sleepWoken(core uint32) bool {
	return false
}

func enableInterruptOnCore0(num uint32) {
}
