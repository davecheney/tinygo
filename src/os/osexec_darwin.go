//go:build darwin

package os

import "unsafe"

// spawnFileActions has the layout of posix_spawn_file_actions_t in the macOS SDK, which is void *.
type spawnFileActions struct {
	actions unsafe.Pointer
}

// spawnKeepFd keeps fd open in the child. On macOS adddup2 with the same fd keeps FD_CLOEXEC,
// so use addinherit_np, see posix_spawn_file_actions_addclose(3).
func spawnKeepFd(fa *spawnFileActions, fd int32) int32 {
	return libc_posix_spawn_file_actions_addinherit_np(fa, fd)
}

//export posix_spawn_file_actions_addinherit_np
func libc_posix_spawn_file_actions_addinherit_np(fa *spawnFileActions, fd int32) int32
