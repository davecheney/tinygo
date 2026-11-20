//go:build esp32c3 || esp32p4

package machine

// Configure and make the SPI peripheral ready to use.
func (spi *SPI) Configure(config SPIConfig) error {
	return spi.configure(config)
}
