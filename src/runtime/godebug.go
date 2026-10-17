//go:build !baremetal

package runtime

import "sync/atomic"

var godebugEmpty string
var godebugEnv atomic.Pointer[string]
var godebugUpdate atomic.Pointer[func(string, string)]

//go:linkname godebug_setUpdate internal/godebug.setUpdate
func godebug_setUpdate(update func(string, string)) {
	if !godebugEnvEnabled() {
		return
	}
	p := new(func(string, string))
	*p = update
	godebugUpdate.Store(p)
	godebugNotify()
}

func godebugSetEnv(value string) {
	p := new(string)
	*p = value
	godebugEnv.Store(p)
	godebugNotify()
}

func godebugUnsetEnv() {
	godebugEnv.Store(&godebugEmpty)
	godebugNotify()
}

func godebugNotify() {
	update := godebugUpdate.Load()
	if update == nil {
		return
	}
	env := godebugStartup
	if p := godebugEnv.Load(); p != nil {
		env = *p
	}
	(*update)(godebugDefault, env)
}

//go:linkname godebug_setNewIncNonDefault internal/godebug.setNewIncNonDefault
func godebug_setNewIncNonDefault(newIncNonDefault func(string) func()) {
}
