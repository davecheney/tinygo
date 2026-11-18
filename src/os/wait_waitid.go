//go:build linux && !baremetal && !tinygo.wasm && !nintendoswitch

package os

import (
	"syscall"
	"unsafe"
)

// blockUntilWaitable waits for the child to exit but leaves it for wait4 to reap.
func (p *Process) blockUntilWaitable() error {
	const _P_PID = 1
	var siginfo [16]uint64
	for {
		_, _, errno := syscall.Syscall6(syscall.SYS_WAITID, _P_PID, uintptr(p.Pid), uintptr(unsafe.Pointer(&siginfo[0])), syscall.WEXITED|syscall.WNOWAIT, 0, 0)
		switch errno {
		case 0:
			return nil
		case syscall.EINTR:
			continue
		default:
			return NewSyscallError("waitid", errno)
		}
	}
}
