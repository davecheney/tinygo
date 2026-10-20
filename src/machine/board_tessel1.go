//go:build tessel1

package machine

// Pin mapping from the Tessel 1 firmware, src/variants/lpc18xx/variant.c in
// https://github.com/tessel/t1-firmware
const (
	LED1 Pin = 5*32 + 22 // PB_2, GPIO5[22]
	LED2 Pin = 3*32 + 11 // P7_3, GPIO3[11]

	LED_WIFI_CONN Pin = 2*32 + 3 // P4_3, GPIO2[3]
	LED_WIFI_ERR  Pin = 2*32 + 4 // P4_4, GPIO2[4]

	LED = LED1

	BUTTON Pin = 5*32 + 13 // P4_9, GPIO5[13], the Config button

	A_G1 Pin = 2*32 + 6  // P4_6
	A_G2 Pin = 4*32 + 7  // P8_7
	A_G3 Pin = 4*32 + 12 // P9_0
	B_G1 Pin = 0*32 + 11 // P1_4
	B_G2 Pin = 1*32 + 8  // P1_5
	B_G3 Pin = 4*32 + 15 // P9_3
	C_G1 Pin = 5*32 + 24 // PB_4
	C_G2 Pin = 5*32 + 23 // PB_3
	C_G3 Pin = 5*32 + 25 // PB_5
	D_G1 Pin = 4*32 + 9  // PA_2
	D_G2 Pin = 5*32 + 20 // PB_0
	D_G3 Pin = 3*32 + 8  // P7_0

	GPIO1 Pin = 3*32 + 9  // P7_1
	GPIO2 Pin = 3*32 + 10 // P7_2
	GPIO3 Pin = 5*32 + 21 // PB_1
	GPIO4 Pin = 5*32 + 12 // P4_8
	GPIO5 Pin = 2*32 + 0  // P4_0
	GPIO6 Pin = 5*32 + 18 // P9_5
)

var pinPads = [...]pinPad{
	{LED1, 0xB, 2, 4},
	{LED2, 7, 3, 0},
	{LED_WIFI_CONN, 4, 3, 0},
	{LED_WIFI_ERR, 4, 4, 0},
	{BUTTON, 4, 9, 4},
	{A_G1, 4, 6, 0},
	{A_G2, 8, 7, 0},
	{A_G3, 9, 0, 0},
	{B_G1, 1, 4, 0},
	{B_G2, 1, 5, 0},
	{B_G3, 9, 3, 0},
	{C_G1, 0xB, 4, 4},
	{C_G2, 0xB, 3, 4},
	{C_G3, 0xB, 5, 4},
	{D_G1, 0xA, 2, 0},
	{D_G2, 0xB, 0, 4},
	{D_G3, 7, 0, 0},
	{GPIO1, 7, 1, 0},
	{GPIO2, 7, 2, 0},
	{GPIO3, 0xB, 1, 4},
	{GPIO4, 4, 8, 4},
	{GPIO5, 4, 0, 0},
	{GPIO6, 9, 5, 4},
}
