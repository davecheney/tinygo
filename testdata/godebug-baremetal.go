//go:debug tarinsecurepath=0

package main

import (
	_ "crypto/sha3"
	"syscall"
	_ "unsafe"
)

type setting struct {
	name string
}

//go:linkname settingValue (*internal/godebug.Setting).Value
func settingValue(*setting) string

//go:extern crypto/internal/fips140.Enabled
var fipsEnabled bool

//go:linkname fipsSupported crypto/internal/fips140.Supported
func fipsSupported() error

var expected string
var initial = settingValue(&setting{"#probe"}) == expected
var documented = settingValue(&setting{"tarinsecurepath"}) == "0"
var afterSetenv, afterUnsetenv, afterClearenv bool

func init() {
	if err := syscall.Setenv("GODEBUG", "probe=changed"); err != nil {
		panic(err)
	}
	afterSetenv = settingValue(&setting{"#probe"}) == expected
	if err := syscall.Unsetenv("GODEBUG"); err != nil {
		panic(err)
	}
	afterUnsetenv = settingValue(&setting{"#probe"}) == expected
	syscall.Clearenv()
	afterClearenv = settingValue(&setting{"#probe"}) == expected
}

func main() {
	if fipsEnabled {
		if err := fipsSupported(); err != nil {
			panic(err)
		}
	}
	println(initial, documented, afterSetenv, afterUnsetenv, afterClearenv, fipsEnabled)
}
