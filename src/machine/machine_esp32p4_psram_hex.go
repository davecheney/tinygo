//go:build esp32p4 && machine.psram

package machine

// initPSRAM brings up the PSRAM for the runtime and maps the given number of
// 64KB pages.
//
//go:linkname initPSRAM runtime.machineInitPSRAM
func initPSRAM(pages uint32) bool {
	return initPSRAM80MHz(pages)
}
