// Package testenv provides the environment for builds in TinyGo tests.
package testenv

import (
	"os"
	"strings"
)

// Environ returns the process environment without the CGO_* flags that make
// test sets to build TinyGo against LLVM.
func Environ() []string {
	var env []string
	for _, kv := range os.Environ() {
		switch key, _, _ := strings.Cut(kv, "="); key {
		case "CGO_CPPFLAGS", "CGO_CFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS":
		default:
			env = append(env, kv)
		}
	}
	return env
}
