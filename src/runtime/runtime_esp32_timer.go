//go:build esp32

package runtime

// sleepTicks spins until the given number of ticks have elapsed.
func sleepTicks(d timeUnit) {
	core := currentCPU()
	target := ticks() + d
	for ticks() < target && !sleepWoken(core) {
	}
}
