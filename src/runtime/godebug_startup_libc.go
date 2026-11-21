//go:build tinygo.linux || darwin || windows

package runtime

func init() {
	if godebugEnvEnabled() {
		godebugStartup = startupEnv("GODEBUG")
	}
}
