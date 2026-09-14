//go:build pimoroni_presto

// This contains the pin mappings for the Pimoroni Presto board.
//
// For more information, see: https://shop.pimoroni.com/products/presto
// Pin data comes from Pimoroni's firmware source:
// https://github.com/pimoroni/presto/blob/main/boards/presto/presto.h
// https://github.com/pimoroni/presto/blob/main/modules/c/presto/presto.h
// https://github.com/pimoroni/presto/blob/main/drivers/st7701/st7701.hpp
// https://github.com/pimoroni/presto/blob/main/modules/py_frozen/touch.py
// https://github.com/pimoroni/presto/blob/main/examples/buzzer.py
// https://github.com/pimoroni/presto/blob/main/examples/sd_basic.py
package machine

// ST7701S 480x480 parallel LCD control pins.
const (
	LCD_CS        Pin = GPIO28
	LCD_SCK       Pin = GPIO26
	LCD_MOSI      Pin = GPIO27
	LCD_BACKLIGHT Pin = GPIO45

	// 18-bit parallel data bus, GPIO1..GPIO18.
	LCD_D0  Pin = GPIO1
	LCD_D1  Pin = GPIO2
	LCD_D2  Pin = GPIO3
	LCD_D3  Pin = GPIO4
	LCD_D4  Pin = GPIO5
	LCD_D5  Pin = GPIO6
	LCD_D6  Pin = GPIO7
	LCD_D7  Pin = GPIO8
	LCD_D8  Pin = GPIO9
	LCD_D9  Pin = GPIO10
	LCD_D10 Pin = GPIO11
	LCD_D11 Pin = GPIO12
	LCD_D12 Pin = GPIO13
	LCD_D13 Pin = GPIO14
	LCD_D14 Pin = GPIO15
	LCD_D15 Pin = GPIO16
	LCD_D16 Pin = GPIO17
	LCD_D17 Pin = GPIO18

	LCD_HSYNC   Pin = GPIO19
	LCD_VSYNC   Pin = GPIO20
	LCD_DE      Pin = GPIO21
	LCD_DOT_CLK Pin = GPIO22
)

// WS2812/SK6812 ambient lighting, 7 mini RGB LEDs.
const (
	WS2812 Pin = GPIO33
)

// CYW43439 wireless chip control pins (RM2 module).
const (
	WL_REG_ON Pin = GPIO23
	WL_DATA   Pin = GPIO24
	WL_CLOCK  Pin = GPIO29
	WL_CS     Pin = GPIO25
)

// Capacitive touch controller (FT6236), on I2C1.
const (
	TOUCH_SDA Pin = GPIO30
	TOUCH_SCL Pin = GPIO31
	TOUCH_INT Pin = GPIO32
)

// Piezo buzzer.
const (
	BUZZER Pin = GPIO43
)

// MicroSD card slot, on a dedicated SPI0 bus.
const (
	SD_SCK  Pin = GPIO34
	SD_MOSI Pin = GPIO35
	SD_MISO Pin = GPIO36
	SD_CS   Pin = GPIO39
)

// Onboard 8 MB PSRAM, QMI chip select. TinyGo has no automatic PSRAM
// bring-up yet; this constant only documents the pin.
const (
	PSRAM_CS Pin = GPIO47
)

// I2C pins. I2C0 is the Qw/ST (Qwiic/STEMMA-QT) connector for external
// breakouts. I2C1 is the onboard touch controller, see TOUCH_SDA/TOUCH_SCL.
const (
	I2C0_SDA_PIN Pin = GPIO40
	I2C0_SCL_PIN Pin = GPIO41

	I2C1_SDA_PIN Pin = TOUCH_SDA
	I2C1_SCL_PIN Pin = TOUCH_SCL
)

// SPI pins. SPI0 is wired to the microSD card slot. SPI1 drives the LCD
// control lines through the display driver directly and is not exposed as
// a general-purpose bus.
const (
	SPI0_SCK_PIN Pin = SD_SCK
	SPI0_SDO_PIN Pin = SD_MOSI
	SPI0_SDI_PIN Pin = SD_MISO

	SPI1_SCK_PIN Pin = NoPin
	SPI1_SDO_PIN Pin = NoPin
	SPI1_SDI_PIN Pin = NoPin
)

// Onboard crystal oscillator frequency, in MHz.
const (
	xoscFreq = 12 // MHz
)

// USB CDC identifiers.
const (
	usb_STRING_PRODUCT      = "Presto"
	usb_STRING_MANUFACTURER = "Pimoroni"
)

var (
	usb_VID uint16 = 0x2e8a
	usb_PID uint16 = 0x000f
)

// UART pins.
// Note: these are the default header values carried over from Pimoroni's
// firmware, but Presto is a fully enclosed device with no external GPIO
// header, so UART0 is not physically broken out. GPIO1 is also LCD_D0; do
// not use UART0 while the display is active.
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

var DefaultUART = UART0
