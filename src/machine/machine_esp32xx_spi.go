//go:build esp32c3 || esp32p4

package machine

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
	return spi.configure(config, bus)
}
