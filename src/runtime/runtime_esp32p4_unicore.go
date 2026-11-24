//go:build esp32p4 && !scheduler.cores

package runtime

//go:inline
func sleepWoken(core uint32) bool {
	return false
}

func enableInterruptOnCore0(num uint32) {
}
