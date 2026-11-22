//go:build !scheduler.cores

package runtime

func waitForSecondaryCoresReady() {}

// Core sleeping in sleepTicks, or 0xff. Used by ports that also build with the
// cores scheduler, which declares it in scheduler_cores.go.
var sleepingCore uint8 = 0xff

// Nothing pauses other cores for the GC without the cores scheduler.
func gcInterruptHandler(core uint32) {}
