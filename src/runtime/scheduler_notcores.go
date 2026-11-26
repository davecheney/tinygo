//go:build !scheduler.cores

package runtime

// Only core 0 runs without the cores scheduler.
func currentCPU() uint32 {
	return 0
}

// No other core can end a sleep without the cores scheduler.
func sleepWoken(core uint32) bool {
	return false
}

func waitForSecondaryCoresReady() {}
