//go:build baremetal

// Package godebug reads build defaults and the startup GODEBUG on bare metal.
// Settings are read-only, including after changes to the process environment.
package godebug

import (
	"internal/bisect"
	"internal/godebugs"
	"unsafe"
)

// These immutable strings are supplied at final link time by the builder.
// Separate globals let the initializer interpreter avoid mutable runtime state.
var defaults, environment string

type Setting struct {
	name string
}

func New(name string) *Setting {
	return &Setting{name: name}
}

func (s *Setting) Name() string {
	if s.Undocumented() {
		return s.name[1:]
	}
	return s.name
}

func (s *Setting) Undocumented() bool {
	return len(s.name) != 0 && s.name[0] == '#'
}

func (s *Setting) String() string {
	return s.Name() + "=" + s.Value()
}

func (s *Setting) Value() string {
	if !s.Undocumented() && godebugs.Lookup(s.Name()) == nil {
		panic("godebug: Value of name not listed in godebugs.All: " + s.name)
	}
	text, ok := lookup(environment, s.Name())
	if !ok {
		text, _ = lookup(defaults, s.Name())
	}
	for i := 0; i < len(text); i++ {
		if text[i] == '#' {
			matcher, _ := bisect.New(text[i+1:])
			if matcher != nil && !matcher.Stack(&stderr) {
				return ""
			}
			return text[:i]
		}
	}
	return text
}

// IncNonDefault validates the setting. Bare-metal targets do not collect metrics.
func (s *Setting) IncNonDefault() {
	info := godebugs.Lookup(s.Name())
	if info == nil || info.Opaque {
		panic("godebug: unexpected IncNonDefault of " + s.name)
	}
}

// Match upstream parse precedence, including values containing '='.
// https://go.dev/src/internal/godebug/godebug.go
func lookup(settings, name string) (string, bool) {
	end, eq := len(settings), -1
	for i := end - 1; i >= -1; i-- {
		if i == -1 || settings[i] == ',' {
			if eq >= 0 && settings[i+1:eq] == name {
				return settings[eq+1 : end], true
			}
			end, eq = i, -1
		} else if settings[i] == '=' {
			eq = i
		}
	}
	return "", false
}

type runtimeStderr struct{}

var stderr runtimeStderr

func (*runtimeStderr) Write(b []byte) (int, error) {
	if len(b) != 0 {
		write(2, unsafe.Pointer(&b[0]), int32(len(b)))
	}
	return len(b), nil
}

//go:linkname write runtime.write
func write(fd uintptr, p unsafe.Pointer, n int32) int32
