package os

import "syscall"

// darwin has no pipe2, so set close-on-exec under ForkLock as Go does in os/pipe_unix.go.
func pipe(p []int) error {
	syscall.ForkLock.RLock()
	defer syscall.ForkLock.RUnlock()
	if err := syscall.Pipe(p); err != nil {
		return err
	}
	syscall.CloseOnExec(p[0])
	syscall.CloseOnExec(p[1])
	return nil
}
