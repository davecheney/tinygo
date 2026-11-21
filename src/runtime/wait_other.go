//go:build !tinygo.riscv && !cortexm && !tinygo.linux && !darwin && !wasip1 && !uefi

package runtime

func waitForEvents() {
	runtimeFatal("deadlocked: no event source")
}
