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

// Configure configures the SPI bus for use.
func (spi *SPI) Configure(config SPIConfig) error {
	bus := spiBusFor(spi.Bus)
	if bus.sckOut == 0 {
		return ErrInvalidSPIBus
	}
	// A zero CS is GPIO0, so treat it as unset.
	if config.CS == 0 {
		config.CS = NoPin
	}
	mode := spiModeFor(config.Mode)
	if mode&spiModeValid == 0 {
		return ErrInvalidSPIMode
	}
	if err := spi.configure(config, bus, mode); err != nil {
		return err
	}
	// Route the pins through the GPIO matrix. NoPin is skipped.
	config.SDI.configure(PinConfig{Mode: PinInput}, bus.sdiIn)
	config.SDO.configure(PinConfig{Mode: PinOutput}, bus.sdoOut)
	config.SCK.configure(PinConfig{Mode: PinOutput}, bus.sckOut)
	config.CS.configure(PinConfig{Mode: PinOutput}, bus.csOut)
	return nil
}

// Transfer writes and reads a single byte.
func (spi *SPI) Transfer(w byte) (byte, error) {
	buf := spi.buffer()
	spi.setBitLen(7)
	buf[0].Set(uint32(w))
	spi.start()
	return byte(buf[0].Get()), nil
}

// Tx writes w and reads into r. The shorter slice is padded with zeros or
// its extra input is discarded.
func (spi *SPI) Tx(w, r []byte) error {
	toTransfer := len(w)
	if len(r) > toTransfer {
		toTransfer = len(r)
	}
	buf := spi.buffer()
	for toTransfer > 0 {
		// Chunk 64 bytes at a time.
		chunkSize := toTransfer
		if chunkSize > 64 {
			chunkSize = 64
		}
		spiTxFillBuffer(buf, w)
		spi.setBitLen(uint32(chunkSize)*8 - 1)
		spi.start()

		// Read rx buffer.
		rxSize := 64
		if rxSize > len(r) {
			rxSize = len(r)
		}
		for i := 0; i < rxSize; i++ {
			r[i] = byte(buf[i/4].Get() >> ((i % 4) * 8))
		}

		// Drop the bytes that were sent and received.
		if len(w) < chunkSize {
			w = nil
		} else {
			w = w[chunkSize:]
		}
		if len(r) < chunkSize {
			r = nil
		} else {
			r = r[chunkSize:]
		}
		toTransfer -= chunkSize
	}
	return nil
}
