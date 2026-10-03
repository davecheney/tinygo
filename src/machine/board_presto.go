//go:build presto

// This contains the pin mappings for the Pimoroni Presto board.
//
// For more information, see: https://shop.pimoroni.com/products/presto
// Pinout source: https://github.com/pimoroni/presto/blob/main/boards/presto/presto.h
package machine

const (
	LED Pin = GPIO25

	// 8 MiB APS6404L PSRAM on QMI chip select 1.
	psramCSPin Pin = GPIO47

	// Onboard crystal oscillator frequency, in MHz.
	xoscFreq = 12 // MHz
)

// I2C pins.
const (
	I2C0_SDA_PIN Pin = GPIO40
	I2C0_SCL_PIN Pin = GPIO41

	I2C1_SDA_PIN Pin = NoPin
	I2C1_SCL_PIN Pin = NoPin
)

// SPI pins.
const (
	SPI0_SCK_PIN Pin = GPIO18
	SPI0_SDO_PIN Pin = GPIO19
	SPI0_SDI_PIN Pin = GPIO16

	SPI1_SCK_PIN Pin = NoPin
	SPI1_SDO_PIN Pin = NoPin
	SPI1_SDI_PIN Pin = NoPin
)

// UART pins.
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

var DefaultUART = UART0

// USB identifiers
const (
	usb_STRING_PRODUCT      = "Presto"
	usb_STRING_MANUFACTURER = "Pimoroni"
)

var (
	usb_VID uint16 = 0x2E8A
	usb_PID uint16 = 0x000F
)
