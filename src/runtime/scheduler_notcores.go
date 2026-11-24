//go:build !scheduler.cores

package runtime

const numCPU = 1

func currentCPU() uint32 {
	return 0
}

func waitForSecondaryCoresReady() {}
