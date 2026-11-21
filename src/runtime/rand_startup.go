//go:build !tinygo.linux

package runtime

func startupRand() (uint64, bool) {
	return hardwareRand()
}
