//go:build linux && !baremetal && !tinygo.wasm && !nintendoswitch

package os

import "unsafe"

// spawnFileActions has the layout of posix_spawn_file_actions_t in musl and glibc.
type spawnFileActions struct {
	_       [2]int32
	actions unsafe.Pointer
	_       [16]int32
}

// spawnKeepFd keeps fd open in the child. adddup2 with the same fd clears FD_CLOEXEC, see
// https://pubs.opengroup.org/onlinepubs/9799919799/functions/posix_spawn_file_actions_adddup2.html
func spawnKeepFd(fa *spawnFileActions, fd int32) int32 {
	return libc_posix_spawn_file_actions_adddup2(fa, fd, fd)
}
