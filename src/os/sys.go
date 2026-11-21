//go:build !tinygo.linux

package os

func Hostname() (name string, err error) {
	return "", ErrNotImplemented
}
