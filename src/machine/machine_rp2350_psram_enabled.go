//go:build tinygo && rp2350 && rp2350_psram

package machine

import (
	"runtime/interrupt"
	"unsafe"
)

/*
typedef unsigned char uint8_t;
typedef unsigned long uint32_t;
typedef volatile uint32_t io_rw_32;

#define ram_func __attribute__((section(".ramfuncs"),noinline))

// RP2350 datasheet section 12.14.6 (QMI registers) and 4.4.5 (XIP_CTRL).
#define QMI_BASE        0x400d0000
#define QMI_DIRECT_CSR  (*(io_rw_32 *)(QMI_BASE + 0x00))
#define QMI_DIRECT_TX   (*(io_rw_32 *)(QMI_BASE + 0x04))
#define QMI_DIRECT_RX   (*(io_rw_32 *)(QMI_BASE + 0x08))
#define QMI_M1_TIMING   (*(io_rw_32 *)(QMI_BASE + 0x20))
#define QMI_M1_RFMT     (*(io_rw_32 *)(QMI_BASE + 0x24))
#define QMI_M1_RCMD     (*(io_rw_32 *)(QMI_BASE + 0x28))
#define QMI_M1_WFMT     (*(io_rw_32 *)(QMI_BASE + 0x2c))
#define QMI_M1_WCMD     (*(io_rw_32 *)(QMI_BASE + 0x30))
#define XIP_CTRL_CTRL_SET (*(io_rw_32 *)(0x400c8000 + 0x2000))

#define DIRECT_CSR_EN          0x00000001
#define DIRECT_CSR_BUSY        0x00000002
#define DIRECT_CSR_ASSERT_CS1N 0x00000008
#define DIRECT_CSR_TXEMPTY     0x00000800
#define DIRECT_CSR_CLKDIV_LSB  22
#define DIRECT_TX_OE           0x00080000
#define DIRECT_TX_IWIDTH_Q     (2u << 16)
#define XIP_CTRL_WRITABLE_M1   0x00000800

// Quad width prefix, address, suffix, dummy and data. 8 bit prefix.
#define M1_FMT_QUAD   0x000012aa
#define M1_DUMMY_24   (6u << 16)

// APS6404L commands.
#define PSRAM_CMD_QUAD_END    0xf5
#define PSRAM_CMD_QUAD_ENABLE 0x35
#define PSRAM_CMD_READ_ID     0x9f
#define PSRAM_CMD_RSTEN       0x66
#define PSRAM_CMD_RST         0x99
#define PSRAM_CMD_QUAD_READ   0xeb
#define PSRAM_CMD_QUAD_WRITE  0x38
#define PSRAM_CMD_NOOP        0xff
#define PSRAM_KGD             0x5d

// Derived from sfe_psram.c, Copyright (c) 2024 SparkFun Electronics, MIT license.
// https://github.com/sparkfun/sparkfun-pico/blob/main/sparkfun_pico/sfe_psram.c
// Runs from RAM with interrupts disabled so nothing fetches from flash in direct mode.
// Returns 0 if no PSRAM answered, otherwise 0x100 | EID byte.

static ram_func void psram_wait_idle(void) {
	while (QMI_DIRECT_CSR & DIRECT_CSR_BUSY) {
	}
}

static ram_func void psram_delay(void) {
	for (int i = 0; i < 20; i++) {
		__asm__ volatile ("nop");
	}
}

static ram_func void psram_command(uint32_t cmd) {
	QMI_DIRECT_CSR |= DIRECT_CSR_ASSERT_CS1N;
	QMI_DIRECT_TX = cmd;
	psram_wait_idle();
	QMI_DIRECT_CSR &= ~DIRECT_CSR_ASSERT_CS1N;
	psram_delay();
	(void)QMI_DIRECT_RX;
}

ram_func uint32_t psram_setup(uint32_t timing) {
	QMI_DIRECT_CSR = 30u << DIRECT_CSR_CLKDIV_LSB | DIRECT_CSR_EN;
	psram_wait_idle();

	// Leave QPI mode in case the chip was left in it by a previous boot.
	QMI_DIRECT_CSR |= DIRECT_CSR_ASSERT_CS1N;
	QMI_DIRECT_TX = DIRECT_TX_OE | DIRECT_TX_IWIDTH_Q | PSRAM_CMD_QUAD_END;
	psram_wait_idle();
	(void)QMI_DIRECT_RX;
	QMI_DIRECT_CSR &= ~DIRECT_CSR_ASSERT_CS1N;
	psram_delay();

	uint8_t kgd = 0;
	uint8_t eid = 0;
	QMI_DIRECT_CSR |= DIRECT_CSR_ASSERT_CS1N;
	for (int i = 0; i < 7; i++) {
		QMI_DIRECT_TX = i == 0 ? PSRAM_CMD_READ_ID : PSRAM_CMD_NOOP;
		while ((QMI_DIRECT_CSR & DIRECT_CSR_TXEMPTY) == 0) {
		}
		psram_wait_idle();
		uint8_t b = (uint8_t)QMI_DIRECT_RX;
		if (i == 5) {
			kgd = b;
		} else if (i == 6) {
			eid = b;
		}
	}
	QMI_DIRECT_CSR &= ~DIRECT_CSR_ASSERT_CS1N;
	psram_delay();

	if (kgd != PSRAM_KGD) {
		QMI_DIRECT_CSR &= ~DIRECT_CSR_EN;
		return 0;
	}

	psram_command(PSRAM_CMD_RSTEN);
	psram_command(PSRAM_CMD_RST);
	psram_command(PSRAM_CMD_QUAD_ENABLE);

	QMI_DIRECT_CSR &= ~(DIRECT_CSR_ASSERT_CS1N | DIRECT_CSR_EN);

	QMI_M1_TIMING = timing;
	QMI_M1_RFMT = M1_FMT_QUAD | M1_DUMMY_24;
	QMI_M1_RCMD = PSRAM_CMD_QUAD_READ;
	QMI_M1_WFMT = M1_FMT_QUAD;
	QMI_M1_WCMD = PSRAM_CMD_QUAD_WRITE;
	XIP_CTRL_CTRL_SET = XIP_CTRL_WRITABLE_M1;

	return 0x100 | eid;
}
*/
import "C"

//go:extern __psram_size
var psramDeclaredSize [0]byte

//go:extern __psram_bss_start
var psramBSSStart [0]byte

//go:extern __psram_bss_end
var psramBSSEnd [0]byte

//go:extern __psram_noinit_end
var psramStaticEnd [0]byte

func initPSRAM() {
	psramCSPin.setFunc(fnQMI)

	state := interrupt.Disable()
	id := uint32(C.psram_setup(C.uint32_t(psramTiming(cpuFreq))))
	interrupt.Restore(state)

	size := uintptr(0)
	if id != 0 {
		size = psramSizeFromEID(uint8(id))
	}

	if declared := uintptr(unsafe.Pointer(&psramDeclaredSize)); size > declared {
		size = declared
	}
	psramSize = size

	if uintptr(unsafe.Pointer(&psramStaticEnd)) > PSRAMBase+psramSize {
		panic("machine: PSRAM not found but .psram_bss or .psram_noinit is in use")
	}

	// Both ends are 8 byte aligned by the linker script.
	end := unsafe.Pointer(&psramBSSEnd)
	for p := unsafe.Pointer(&psramBSSStart); p != end; p = unsafe.Add(p, 8) {
		*(*uint64)(p) = 0
	}
}
