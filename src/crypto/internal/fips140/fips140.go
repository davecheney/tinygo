// Package fips140 provides compatibility with the Go cryptography packages.
// TinyGo does not support FIPS 140-3 mode.
package fips140

import (
	"errors"
	"strings"
	_ "unsafe"
)

const Enabled = false

const unsupported = "FIPS 140-3 mode is not supported by TinyGo"

func Supported() error {
	return errors.New(unsupported)
}

func Name() string {
	return "Go Cryptographic Module"
}

func Version() string {
	return "latest"
}

func CAST(name string, f func() error) {
	checkName(name)
}

func PCT(name string, f func() error) {
	checkName(name)
}

func checkName(name string) {
	if strings.ContainsAny(name, ",#=:") {
		panic("fips: invalid self-test name: " + name)
	}
}

//go:linkname getIndicator crypto/internal/fips140.getIndicator
func getIndicator() uint8

//go:linkname setIndicator crypto/internal/fips140.setIndicator
func setIndicator(uint8)

const (
	indicatorUnset uint8 = iota
	indicatorFalse
	indicatorTrue
)

func ResetServiceIndicator() {
	setIndicator(indicatorUnset)
}

func ServiceIndicator() bool {
	return getIndicator() == indicatorTrue
}

func RecordApproved() {
	if getIndicator() == indicatorUnset {
		setIndicator(indicatorTrue)
	}
}

func RecordNonApproved() {
	setIndicator(indicatorFalse)
}
