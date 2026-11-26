//go:build esp32 || esp32s3

package runtime

import "sync/atomic"

// CPU interrupt lines that core 1 asked core 0 to enable.
var core0EnableRequest atomic.Uint32

// Enable a CPU interrupt line on core 0. Called by interrupt.Enable on core 1,
// which must have interrupts enabled so the GC can still pause it.
func enableInterruptOnCore0(num uint32) {
	bit := uint32(1) << num
	core0EnableRequest.Or(bit)
	signalCore(0, 1)
	for core0EnableRequest.Load()&bit != 0 {
		spinLoopWait()
	}
}
