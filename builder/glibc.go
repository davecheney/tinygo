package builder

import (
	"path/filepath"
)

// The system glibc, linked dynamically.
//
// Every other libc TinyGo supports is one it ships and links statically, which
// is what lets a build behave the same on any machine. This one is different:
// it uses the glibc already installed on the host. That is the point of it.
// Linking any system shared library - SDL, GTK, ALSA - drags in the libc that
// library was built against, and on a glibc distribution that is glibc. Two
// libcs cannot share one process: whichever installs the thread pointer first
// owns it, and the other one's TLS accesses - errno, malloc's tcache - then
// read the wrong memory. So to link against a system library at all, TinyGo
// has to give up its own libc and use the one already in the room.
//
// The consequences are worth stating plainly. This libc can only target the
// host, because it reads the host's headers and crt objects; it is not
// reproducible across machines the way the bundled libcs are; and a binary
// built with it is dynamically linked, not the single static file TinyGo
// usually produces. It is therefore opt-in, via -libc=glibc, and musl remains
// the default.

// glibcStartupJobs returns the startup objects to link, in the order they must
// appear. crt1.o defines _start, which calls __libc_start_main with the main
// that TinyGo's runtime exports; crti.o and crtn.o bracket the .init and .fini
// sections, so crtn.o has to come last.
func glibcStartupJobs(libDir string) []*compileJob {
	var jobs []*compileJob
	for _, name := range []string{"crt1.o", "crti.o", "crtn.o"} {
		jobs = append(jobs, dummyCompileJob(filepath.Join(libDir, name)))
	}
	return jobs
}
