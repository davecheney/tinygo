//go:build tessel1

package machine

// Pin mapping for board revision TM-00-04, from g_APinDescription_boardV3 in
// src/variants/lpc18xx/variant.c of https://github.com/tessel/t1-firmware
const (
	LED1 Pin = 5*32 + 22 // PB_2, GPIO5[22]
	LED2 Pin = 3*32 + 11 // P7_3, GPIO3[11]

	LED_WIFI_CONN Pin = 2*32 + 3 // P4_3, GPIO2[3]
	LED_WIFI_ERR  Pin = 2*32 + 4 // P4_4, GPIO2[4]

	LED = LED1

	BUTTON Pin = 5*32 + 13 // P4_9, GPIO5[13], the Config button

	A_G1 Pin = 4*32 + 15 // P9_3, U3_TXD
	A_G2 Pin = 5*32 + 17 // P9_4, U3_RXD
	A_G3 Pin = 4*32 + 12 // P9_0
	B_G1 Pin = 0*32 + 2  // P1_15, U2_TXD
	B_G2 Pin = 0*32 + 3  // P1_16, U2_RXD
	B_G3 Pin = 0*32 + 11 // P1_4
	C_G1 Pin = 5*32 + 21 // PB_1
	C_G2 Pin = 3*32 + 10 // P7_2
	C_G3 Pin = 5*32 + 25 // PB_5
	D_G1 Pin = 5*32 + 18 // P9_5, U0_TXD
	D_G2 Pin = 4*32 + 11 // P9_6, U0_RXD
	D_G3 Pin = 3*32 + 8  // P7_0

	// GPIO bank
	G1 Pin = 4*32 + 3  // P8_3
	G2 Pin = 5*32 + 14 // P4_10
	G3 Pin = 5*32 + 12 // P4_8
	G4 Pin = 5*32 + 23 // PB_3, SCT output 8
	G5 Pin = 2*32 + 5  // P4_5, SCT output 5
	G6 Pin = 5*32 + 20 // PB_0, SCT output 10
	A2 Pin = 3*32 + 13 // P7_5
	A3 Pin = 3*32 + 12 // P7_4
	A4 Pin = 5*32 + 26 // PB_6
)

var pinPads = [...]pinPad{
	{LED1, 0xB, 2, 4},
	{LED2, 7, 3, 0},
	{LED_WIFI_CONN, 4, 3, 0},
	{LED_WIFI_ERR, 4, 4, 0},
	{BUTTON, 4, 9, 4},
	{A_G1, 9, 3, 0},
	{A_G2, 9, 4, 4},
	{A_G3, 9, 0, 0},
	{B_G1, 1, 15, 0},
	{B_G2, 1, 16, 0},
	{B_G3, 1, 4, 0},
	{C_G1, 0xB, 1, 4},
	{C_G2, 7, 2, 0},
	{C_G3, 0xB, 5, 4},
	{D_G1, 9, 5, 4},
	{D_G2, 9, 6, 0},
	{D_G3, 7, 0, 0},
	{G1, 8, 3, 0},
	{G2, 4, 10, 4},
	{G3, 4, 8, 4},
	{G4, 0xB, 3, 4},
	{G5, 4, 5, 0},
	{G6, 0xB, 0, 4},
	{A2, 7, 5, 0},
	{A3, 7, 4, 0},
	{A4, 0xB, 6, 4},
}
