//go:build esp32c3 || esp32p4

package machine

import "errors"

var (
	ErrInvalidSPIBus  = errors.New("machine: SPI bus is invalid")
	ErrInvalidSPIMode = errors.New("machine: SPI mode is invalid")
)

// spiBus holds the GPIO matrix signals of one SPI bus. spiBusFor gives the
// zero value for a bus that the chip does not support.
type spiBus struct {
	sckOut, sdoOut, sdiIn, csOut uint32
}

// Configure and make the SPI peripheral ready to use.
func (spi *SPI) Configure(config SPIConfig) error {
	bus := spiBusFor(spi.Bus)
	if bus.sckOut == 0 {
		return ErrInvalidSPIBus
	}
	// A zero CS is GPIO0, so treat it as unset.
	if config.CS == 0 {
		config.CS = NoPin
	}
	return spi.configure(config, bus)
}
