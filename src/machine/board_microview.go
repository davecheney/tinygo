//go:build microview

// SparkFun MicroView, an ATmega328P module with a 64x48 OLED.
// Pinout: https://learn.sparkfun.com/tutorials/microview-hookup-guide/all
package machine

// Digital pins on the header. D4, D7, D8, D10, D11 and D13 are used by the OLED.
const (
	D0 = PD0 // RX, pin 9
	D1 = PD1 // TX, pin 10
	D2 = PD2 // pin 11
	D3 = PD3 // PWM, pin 12
	D5 = PD5 // PWM, pin 13
	D6 = PD6 // PWM, pin 14
)

// ADC pins on the header.
const (
	ADC0 Pin = PC0 // pin 7
	ADC1 Pin = PC1 // pin 6
	ADC2 Pin = PC2 // pin 5
	ADC3 Pin = PC3 // pin 4
	ADC4 Pin = PC4 // SDA, pin 3
	ADC5 Pin = PC5 // SCL, pin 2
)

// OLED pins from MicroView.h in https://github.com/geekammo/MicroView-Arduino-Library
const (
	OLED_PWR Pin = PD4 // enables the OLED 3.3V regulator
	OLED_RST Pin = PD7
	OLED_DC  Pin = PB0
	OLED_CS  Pin = PB2
	OLED_SDO Pin = PB3
	OLED_SCK Pin = PB5
)

// UART pins
const (
	UART_TX_PIN Pin = PD1
	UART_RX_PIN Pin = PD0
)
