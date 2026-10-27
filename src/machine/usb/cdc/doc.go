// Package cdc implements USB Communication Device Class serial devices.
//
// The Arduino Leonardo target supports USB CDC on the ATmega32U4 and selects
// it by default. Use -serial=none to disable the USB serial implementation.
// The target reserves 4 KiB for the standard Leonardo bootloader; verify
// bootloader compatibility before flashing third-party ATmega32U4 boards.
package cdc
