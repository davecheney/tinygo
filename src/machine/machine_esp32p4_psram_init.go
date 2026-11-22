//go:build esp32p4 && machine.psram

package machine

// initPSRAM brings up the PSRAM for the runtime.
//
//go:linkname initPSRAM runtime.machineInitPSRAM
func initPSRAM() bool {
	return InitPSRAM80MHz()
}
