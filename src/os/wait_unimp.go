//go:build darwin

package os

// blockUntilWaitable returns at once, because darwin has no waitid. Go does the same in os/wait_unimp.go.
func (p *Process) blockUntilWaitable() error {
	return nil
}
