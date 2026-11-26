//go:build esp32p4 && machine.psram

// ESP32-P4 AP-Hex PSRAM bring-up, after ESP-IDF esp_psram_impl_ap_hex.c.
//
// Register layout notes:
//   - The PSRAM "system" controller is MSPI id 2 (SPIMEM2) @ 0x5008E000, and the
//     "peripheral" controller is MSPI id 3 (SPIMEM3) @ 0x5008F000.  These are the
//     same SPI_MEM IP as the flash SPI0 block, so SPI0_Type (which carries the
//     shared SPI_MEM_* / SPI_SMEM_* registers and their field accessors) is
//     reused here cast to the PSRAM base.
//   - The per-region MMU (SPI_MEM_S_MMU_ITEM_{CONTENT,INDEX} @ +0x37C/0x380)
//     maps PSRAM virtual pages (64KB) into CPU address space at 0x48000000.
package runtime

import (
	"device/esp"
	"machine"
	"runtime/volatile"
	"unsafe"
)

// ROM helper wrappers for the ESP32-P4 PSRAM user-command path.  These call
// the same esp_rom_spi_* functions ESP-IDF's AP-Hex PSRAM driver uses for its
// mode-register and probe transactions.  The RISC-V trampolines live in
// src/device/esp/esp32p4_psram.S, which is compiled as a target extra-file
// (TinyGo does not assemble .S files discovered inside Go packages).
//
// Addresses from components/esp_rom/esp32p4/ld/esp32p4.rom.ld (IDF 5.5.5).

//go:export psram_rom_set_op_mode
func psramROMSetOpMode(spiNum uint32, mode uint32)

//go:export psram_rom_cmd_config
func psramROMCmdConfig(spiNum uint32, cmd *psramROMCommand)

//go:export psram_rom_cmd_start
func psramROMCmdStart(spiNum uint32, rxBuf *uint32, rxLen uint32, csMask uint32, writeErase uint32)

// psramROMCommand mirrors esp_rom_spi_cmd_t from
// components/esp_rom/include/esp_rom_spiflash.h.
type psramROMCommand struct {
	cmd          uint16
	cmdBitLen    uint16
	addr         *uint32
	addrBitLen   uint32
	txData       *uint32
	txDataBitLen uint32
	rxData       *uint32
	rxDataBitLen uint32
	dummyBitLen  uint32
}

// ESP_ROM_SPIFLASH_OPI_DTR_MODE from esp_rom_spiflash.h.
const psramROMOPIDTRMode = 7

// psramROMSpiID maps an MSPI register base to the controller index expected by
// the ROM helpers: MSPI2 (system) = 2, MSPI3 (peripheral) = 3.
func psramROMSpiID(base uintptr) uint32 {
	if base == psramMspiPeriBase {
		return 3
	}
	return 2
}

// Base addresses for the dual PSRAM MSPI controllers (same SPI_MEM IP as SPI0).
const (
	psramMspiSystemBase = 0x5008E000 // MSPI id 2: SPIMEM2 controller (system/cache path)
	psramMspiPeriBase   = 0x5008F000 // MSPI id 3: SPIMEM3 (peripheral path)
)

// MPLL clock-source select (reg_psram_clk_src_sel).
const psramClkSrcMPLL = 1

// PSRAM MMU configuration.  The P4 per-target MMU (SPI_MEM_S, MSPI id 2) maps
// 64KB pages from PSRAM0 into CPU virtual address space starting at 0x48000000
// (SOC_MMU_PSRAM_VADDR_BASE).  The physical page number is 0-based within the
// PSRAM0 region; the PSRAM0 physical AXI base is 0x88000000 (direct/uncached).
const (
	psramVAddrBase    = 0x48000000
	psramPageSize     = 0x10000 // 64KB
	psramMMUEntryMask = psramPageSize*1024 - 1

	// MMU ITEM_CONTENT flag bits within the PSRAM (SPI_MEM_S) MMU:
	socMMUPsramValid   = 1 << 11 // SOC_MMU_PSRAM_VALID
	socMMUPAccessPsram = 1 << 10 // SOC_MMU_ACCESS_PSRAM
)

// psramMSPI returns the system PSRAM MSPI controller register block, viewed
// through the shared SPI_MEM register layout of SPI0_Type.
func psramMSPI() *esp.SPI0_Type {
	return (*esp.SPI0_Type)(unsafe.Pointer(uintptr(psramMspiSystemBase)))
}

// enablePSRAMModuleClock gates on the PSRAM MSPI functional clocks and releases
// the dual-MSPI reset, then selects the PSRAM clock source.  This mirrors
// ESP-IDF psram_ctrlr_ll_enable_module_clock / _reset_module_clock /
// _select_clk_source.
func enablePSRAMModuleClock() {
	// Gate the PSRAM system APB clock and the PLL/core functional clocks.
	esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL0_REG_PSRAM_SYS_CLK_EN(1)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL00_REG_PSRAM_PLL_CLK_EN(1)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL00_REG_PSRAM_CORE_CLK_EN(1)

	// Pulse the dual-MSPI AXI/APB reset (assert, then release).  Order matches
	// ESP-IDF _psram_ctrlr_ll_reset_module_clock.
	esp.HP_SYS_CLKRST.SetHP_RST_EN0_REG_RST_EN_DUAL_MSPI_AXI(1)
	esp.HP_SYS_CLKRST.SetHP_RST_EN0_REG_RST_EN_DUAL_MSPI_APB(1)
	esp.HP_SYS_CLKRST.SetHP_RST_EN0_REG_RST_EN_DUAL_MSPI_APB(0)
	esp.HP_SYS_CLKRST.SetHP_RST_EN0_REG_RST_EN_DUAL_MSPI_AXI(0)

	// Select the PSRAM clock source.  The AP-Hex driver uses MPLL (320MHz at
	// 80MHz bus speed); ESP-IDF's esp_psram_impl_enable calls
	// psram_ctrlr_ll_select_clk_source(..., PSRAM_CLK_SRC_MPLL) after the
	// reset pulse.
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL00_REG_PSRAM_CLK_SRC_SEL(psramClkSrcMPLL)
}

// mapPSRAM maps `pages` (64KB each) of physical PSRAM0 into CPU virtual address
// space starting at the given vaddr (which must be 64KB aligned).  For the
// default region this is vaddr 0x48000000 with a 0-based physical page start.
// It mirrors ESP-IDF mmu_hal_map_region(mmu_id=1, MMU_TARGET_PSRAM0, ...).
func mapPSRAM(vaddr, physPageStart uintptr, pages uint32) {
	mmu := psramMSPI()
	paddrPage := uint32(physPageStart >> 16)
	for i := uint32(0); i < pages; i++ {
		entryID := uint32((vaddr+uintptr(i)*psramPageSize)&psramMMUEntryMask) >> 16
		mmuVal := (paddrPage + i) | socMMUPAccessPsram | socMMUPsramValid
		mmu.SetSPI_MEM_MMU_ITEM_INDEX(entryID)
		mmu.SetSPI_MEM_MMU_ITEM_CONTENT(mmuVal)
	}
}

// The PSRAM clock source is the MPLL at 320MHz from the 40MHz XTAL, so the 80MHz
// PSRAM bus clock is MPLL/4.  The MPLL divider is set through regi2c.

// MPLL regi2c slave id and register addresses (soc/regi2c_mpll.h).
const (
	r2cMpllSlave     = r2cBlockMSPI
	r2cMpllIRCalRstb = 1 // bit 5
	r2cMpllDivReg    = 2 // div bits[7:3], ref_div bits[2:0]
	r2cMpllDHref     = 3 // bits[5:4]
)

// enablePSRAMMPLLFreq powers up and programmes the MPLL for the requested
// frequency from the 40MHz XTAL.  ESP-IDF's AP-Hex driver uses 320MHz for an
// 80MHz PSRAM bus and 400MHz for a 200MHz bus.
//
// Like ESP-IDF's esp_clk_tree_mpll_acquire, the PSRAM+MPLL analog domain must be
// powered first: on the P4 this is the internal LDO channel 2 (VO2) at 1.8V
// (CONFIG_ESP_LDO_CHAN_PSRAM_DOMAIN=2, VOLTAGE=1800mV), plus the 3.3V peripheral
// rail (LDO4) that feeds the PSRAM I/O on the CrowPanel.
func enablePSRAMMPLLFreq(mhz uint32) bool {
	if mhz < 20 {
		return false
	}
	// Power the PSRAM/MPLL + peripheral domains.  If the LDOs fail, abort.
	if lerr := machine.LDOConfigure(machine.LDO2, 1800); lerr != nil {
		return false
	}
	if lerr := machine.LDOConfigure(machine.LDO4, 3300); lerr != nil {
		return false
	}

	regi2cEnableGate()

	// Power up the MPLL analog circuit and gate its 500MHz clock.
	esp.PMU.SetRF_PWC_MSPI_PHY_XPD(1)
	esp.LP_AON_CLKRST.SetLP_AONCLKRST_HP_CLK_CTRL_LP_AONCLKRST_HP_MPLL_500M_CLK_EN(1)

	// Program the MPLL divider over the analog I2C master (clk_ll_mpll_set_config).
	dhref := regi2cRead(r2cMpllSlave, r2cMpllDHref)
	regi2cWrite(r2cMpllSlave, r2cMpllDHref, dhref|(3<<4))
	rstb := regi2cRead(r2cMpllSlave, r2cMpllIRCalRstb)
	regi2cWrite(r2cMpllSlave, r2cMpllIRCalRstb, rstb&0xdf)
	regi2cWrite(r2cMpllSlave, r2cMpllIRCalRstb, rstb|(1<<5))
	// MPLL = XTAL*(div+1)/(ref_div+1); ref_div=1, so div = mhz/20 - 1.
	div := mhz/20 - 1
	regi2cWrite(r2cMpllSlave, r2cMpllDivReg, (uint8(div)<<3)|1)

	// Self-calibration: start, wait done (bounded), stop.
	esp.HP_SYS_CLKRST.SetANA_PLL_CTRL0_REG_MSPI_CAL_STOP(0)
	for i := 0; esp.HP_SYS_CLKRST.GetANA_PLL_CTRL0_REG_MSPI_CAL_END() == 0 && i < regi2cMaxBusy; i++ {
	}
	done := esp.HP_SYS_CLKRST.GetANA_PLL_CTRL0_REG_MSPI_CAL_END() != 0
	esp.HP_SYS_CLKRST.SetANA_PLL_CTRL0_REG_MSPI_CAL_STOP(1)
	return done
}

// psramSetBusClock writes the PSRAM bus-clock divider.  MSPI2 uses
// SPI_MEM_S_SRAM_CLK (+0x50); MSPI3 uses SPI1_MEM_S_CLOCK (+0x14).  Both encode
// the same SCLKCNT_N/H/L fields.
func psramSetBusClock(base uintptr, freqdiv uint32) {
	if freqdiv == 0 {
		freqdiv = 1
	}
	var freqbits uint32
	if freqdiv == 1 {
		freqbits = 1 << 31 // SCLK_EQU_SYSCLK
	} else {
		freqbits = ((freqdiv - 1) << 16) | ((freqdiv/2 - 1) << 8) | (freqdiv - 1)
	}
	if base == psramMspiPeriBase {
		(*esp.SPI1_Type)(unsafe.Pointer(base)).SPI_MEM_CLOCK.Set(freqbits)
		return
	}
	psramMSPI().SPI_MEM_SRAM_CLK.Set(freqbits)
}

// AP-Hex PSRAM command codes and bus constants for the CrowPanel's 80 MHz
// hex-mode configuration (V1.2 sdkconfig: CONFIG_SPIRAM_MODE_HEX=y,
// CONFIG_SPIRAM_SPEED_80M=y).  ESP-IDF's AP-Hex driver selects these values
// from the "else" branch of esp_psram_impl_ap_hex.c (any speed below 200M).
const (
	apHexSyncWrite        = 0x8080 // PSDRAM sync write command
	apHexSyncRead         = 0x0000 // PSRAM sync read command
	apHexWrCmdBitlen      = 16
	apHexRdCmdBitlen      = 16
	apHexAddrBitlen       = 32
	apHexRdDummy          = 2 * (10 - 1) // 80MHz: 18
	apHexWrDummy          = 2 * (5 - 1)  // 80MHz: 8
	apHexRegRdDummy       = 2 * (5 - 1)  // 80MHz register read: 8
	apHexRdLatency        = 2            // MR0 read latency at 80MHz
	apHexWrLatency        = 2            // MR4 write latency at 80MHz
	apHexCSSetup          = 4
	apHexCSHold           = 4
	apHexCSHoldDelay      = 3
	psramPageSizeCode2048 = 3 // SPI_SMEM_ECC_CTRL page-size code for 2048B

	// MPLL is programmed to 320MHz and the PSRAM bus runs at MPLL/4 = 80MHz.
	psram80MPLLMHz = 320
	psram80BusMHz  = 80
	psram80Div     = psram80MPLLMHz / psram80BusMHz
)

// configurePSRAMController programmes the PSRAM system MSPI (SPIMEM2) for the
// AP-Hex octal-DDR device: sync read/write command codes, 32-bit address depth,
// read/write dummy cycles (80MHz: rd 18, wr 8), DDR mode, octal line mode, and
// AXI access with read/write splicing.  Mirrors ESP-IDF s_config_mspi_for_psram.
func configurePSRAMController() {
	m := psramMSPI()

	// Write cmd: 0x8080, 16-bit (value set is an RMW on bits [15:0]).
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_CACHE_SRAM_USR_WCMD(1)
	m.SetSPI_MEM_SRAM_DWR_CMD_SPI_MEM_CACHE_SRAM_USR_WR_CMD_BITLEN(apHexWrCmdBitlen - 1)
	m.SetSPI_MEM_SRAM_DWR_CMD_SPI_MEM_CACHE_SRAM_USR_WR_CMD_VALUE(apHexSyncWrite)
	// Read cmd: 0x0000, 16-bit.
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_CACHE_SRAM_USR_RCMD(1)
	m.SetSPI_MEM_SRAM_DRD_CMD_SPI_MEM_CACHE_SRAM_USR_RD_CMD_BITLEN(apHexRdCmdBitlen - 1)
	m.SetSPI_MEM_SRAM_DRD_CMD_SPI_MEM_CACHE_SRAM_USR_RD_CMD_VALUE(apHexSyncRead)

	// Address: 32-bit, 4-byte mode.
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_SRAM_ADDR_BITLEN(apHexAddrBitlen - 1)
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_CACHE_USR_SADDR_4BYTE(1)

	// Dummy: write 8, read 18 (80MHz), variable dummy + write dummy level ctrl.
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_USR_WR_SRAM_DUMMY(1)
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_SRAM_WDUMMY_CYCLELEN(apHexWrDummy - 1)
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_USR_RD_SRAM_DUMMY(1)
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_SRAM_RDUMMY_CYCLELEN(apHexRdDummy - 1)
	m.SetSPI_SMEM_DDR_SPI_SMEM_VAR_DUMMY(1)
	m.SetSPI_MEM_SRAM_CMD_SPI_MEM_SDUMMY_WOUT(1)

	// DDR with no data swap.
	m.SetSPI_SMEM_DDR_WDAT_SWP(0)
	m.SetSPI_SMEM_DDR_RDAT_SWP(0)
	m.SetSPI_SMEM_DDR_EN(1)

	// Octal command/address lines and x16 ("hex") data lines.  The AP-Hex
	// device transfers 16 data bits per clock, so SDIN_HEX/SDOUT_HEX must be
	// set (IDF psram_ctrlr_ll_enable_hex_data_line_mode(true)); leaving them
	// zero configures x8 data lines and makes every read/write corrupt.
	m.SetSPI_MEM_CACHE_SCTRL_SPI_MEM_SRAM_OCT(1)
	m.SetSPI_MEM_SRAM_CMD_SPI_MEM_SCMD_OCT(1)
	m.SetSPI_MEM_SRAM_CMD_SPI_MEM_SADDR_OCT(1)
	m.SetSPI_MEM_SRAM_CMD_SPI_MEM_SDOUT_OCT(1)
	m.SetSPI_MEM_SRAM_CMD_SPI_MEM_SDIN_OCT(1)
	m.SetSPI_MEM_SRAM_CMD_SPI_MEM_SDIN_HEX(1)
	m.SetSPI_MEM_SRAM_CMD_SPI_MEM_SDOUT_HEX(1)

	// AXI access + read/write splice.
	m.SetSPI_MEM_CACHE_FCTRL_SPI_MEM_AXI_REQ_EN(1)
	m.SetSPI_MEM_CACHE_FCTRL_SPI_CLOSE_AXI_INF_EN(0)
	m.SetSPI_MEM_CTRL1_SPI_MEM_AW_SPLICE_EN(1)
	m.SetSPI_MEM_CTRL1_SPI_MEM_AR_SPLICE_EN(1)
}

// configurePSRAMTiming applies the pre-chip CS timing, split transaction, page
// size, bus clock (MPLL/4 = 80MHz) and DLL enable on both MSPI controllers,
// mirroring the lead-in portion of esp_psram_impl_enable.
func configurePSRAMTiming() {
	m := psramMSPI()
	// CS timing (s_set_psram_cs_timing).
	m.SetSPI_SMEM_AC_SPI_SMEM_CS_SETUP(1)
	m.SetSPI_SMEM_AC_SPI_SMEM_CS_SETUP_TIME(apHexCSSetup - 1)
	m.SetSPI_SMEM_AC_SPI_SMEM_CS_HOLD(1)
	m.SetSPI_SMEM_AC_SPI_SMEM_CS_HOLD_TIME(apHexCSHold - 1)
	m.SetSPI_SMEM_AC_SPI_SMEM_CS_HOLD_DELAY(apHexCSHoldDelay - 1)
	// Split transaction + 2048B page.
	m.SetSPI_SMEM_AC_SPI_SMEM_SPLIT_TRANS_EN(1)
	m.SetSPI_SMEM_ECC_CTRL_SPI_SMEM_PAGE_SIZE(psramPageSizeCode2048)
	// Bus clock on both controllers (MPLL 320MHz / 80MHz = freqdiv 4).
	psramSetBusClock(psramMspiSystemBase, psram80Div)
	psramSetBusClock(psramMspiPeriBase, psram80Div)
	// DLL enable on both controllers (SPI_MEM_S / SPI1_MEM_S timing-cali).
	m.SetSPI_SMEM_TIMING_CALI_SPI_SMEM_DLL_TIMING_CALI(1)
	psramPeriTimingCali().SetBits(psramDLLTimingCaliBit)
}

// psramPeriTimingCali returns the MSPI3 timing-cali register.
func psramPeriTimingCali() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer(uintptr(psramMspiPeriBase + 0x180)))
}

const psramDLLTimingCaliBit = 1 << 5

// AP-Hex register read/write command codes and constants.
const (
	apHexRegRead   = 0x4040
	apHexRegWrite  = 0xC0C0
	apHexCmdBitlen = 16
	apHexRefData   = 0x5a6b7c8d
)

// psramUserCommand drives a single SPI_MEM user transaction on the given
// controller (MSPI2 @0x5008E000 or MSPI3 @0x5008F000) through the ESP32-P4
// mask-ROM helpers used by ESP-IDF's AP-Hex PSRAM driver.  Using the ROM path
// gives the exact OPI-DTR line-mode setup and bit-length encodings (the ROM
// subtracts one from the address/dummy/data lengths internally).
//
// The write data word must be placed in SPI_MEM_W0 by the caller (psramDataWord);
// the read result is returned when misoBits > 0.
func psramUserCommand(base uintptr, cmd uint32, addr uint32, addrBitlen, dummy, mosiBits, misoBits uint32) uint32 {
	u := (*esp.SPI1_Type)(unsafe.Pointer(base))
	spi := psramROMSpiID(base)

	psramCmdAddr = addr
	psramCmdTx = u.SPI_MEM_W0.Get()
	psramCmdRx = 0
	c := psramROMCommand{
		cmd:          uint16(cmd),
		cmdBitLen:    apHexCmdBitlen,
		addr:         &psramCmdAddr,
		addrBitLen:   addrBitlen,
		txData:       &psramCmdTx,
		txDataBitLen: mosiBits,
		rxData:       &psramCmdRx,
		rxDataBitLen: misoBits,
		dummyBitLen:  dummy,
	}

	psramROMSetOpMode(spi, psramROMOPIDTRMode)
	psramROMCmdConfig(spi, &c)
	if misoBits > 0 {
		psramROMCmdStart(spi, &psramCmdRx, misoBits/8, 1<<1, 0)
	} else {
		psramROMCmdStart(spi, nil, 0, 1<<1, 0)
	}
	return psramCmdRx
}

// Data words for psramUserCommand. A pointer to a local in psramROMCommand
// causes a heap allocation.
var psramCmdAddr, psramCmdTx, psramCmdRx uint32

// initAPHexPSRAM performs the AP-Hex PSRAM chip init: programs MR0/MR4/MR8 mode
// registers (fixed read/write latency, drive strength, burst length, x16 mode)
// over the peripheral controller (MSPI3) and confirms the part is connected by
// a write/read-back of a reference pattern.  Mirrors ESP-IDF
// s_init_psram_mode_reg + s_check_psram_connected.
func initAPHexPSRAM(mspi3Base uintptr) bool {
	// MR0: lt=1 (fixed latency), read_latency=2, drive_str=0 @80MHz.
	// mr0 bits: drive_str[1:0], read_latency[4:2], lt[5], rsvd6[6], tso[7].
	// Read MR0+MR1 (addr 0x0): low byte = MR0, high byte = MR1.
	v01 := psramUserCommand(mspi3Base, apHexRegRead, 0x0, 32, apHexRegRdDummy, 0, 16)
	mr0 := byte(v01 & 0xff)
	mr1 := byte(v01 >> 8)
	mr0 = (mr0 & 0xc2) | (1 << 5) | uint8(apHexRdLatency<<2)
	psramUserCommand16(mspi3Base, apHexRegWrite, 0x0, uint16(mr0)|uint16(mr1)<<8)

	// MR4 (addr 0x4): low byte = MR4, high byte = MR8; set wr_latency[7:5]=2.
	v48 := psramUserCommand(mspi3Base, apHexRegRead, 0x4, 32, apHexRegRdDummy, 0, 16)
	mr4 := byte(v48 & 0xff)
	mr8Old := byte(v48 >> 8)
	mr4 = (mr4 & 0x1f) | uint8(apHexWrLatency<<5)
	psramUserCommand16(mspi3Base, apHexRegWrite, 0x4, uint16(mr4)|uint16(mr8Old)<<8)

	// MR8 (addr 0x8, 8-bit read): set bl=3, bt=0, rbx=1, x16=1.
	mr8 := byte(psramUserCommand(mspi3Base, apHexRegRead, 0x8, 32, apHexRegRdDummy, 0, 8) & 0xff)
	// mr8 bits: bl[1:0], bt[2], rbx[3], x16[6].
	mr8 = (mr8 & 0x68) | 3 | (1 << 3) | (1 << 6)
	psramUserCommand16(mspi3Base, apHexRegWrite, 0x8, uint16(mr8))

	// Connected check: write a reference pattern, read it back (round trip).
	psramDataWord(mspi3Base, apHexRefData)
	psramUserCommand(mspi3Base, apHexSyncWrite, 0x0, 32, apHexWrDummy, 32, 0)
	rd := psramUserCommand(mspi3Base, apHexSyncRead, 0x0, 32, apHexRdDummy, 0, 32)
	return rd == apHexRefData
}

// psramUserCommand16 is a convenience wrapper for a 16-bit data-phase write.
func psramUserCommand16(base uintptr, cmd, addr uint32, data uint16) {
	psramDataWord(base, uint32(data))
	psramUserCommand(base, cmd, addr, 32, 0, 16, 0)
}

// psramDataWord writes a 32-bit data word into the SPI_MEM W0 FIFO of the block.
func psramDataWord(base uintptr, w uint32) {
	u := (*esp.SPI1_Type)(unsafe.Pointer(base))
	u.SPI_MEM_W0.Set(w)
}

// initPSRAM performs the AP-Hex PSRAM bring-up: MPLL=320MHz, PSRAM bus=80MHz,
// x16 hex mode, 2048-byte pages, DLL enabled, then chip mode-register init over
// MSPI3. It maps the __psram_size bytes the target declares at psramVAddrBase.
func initPSRAM() string {
	if !enablePSRAMMPLLFreq(psram80MPLLMHz) {
		return "MPLL init failed"
	}
	enablePSRAMModuleClock()
	configPSRAMPinIO(2)
	configurePSRAMTiming()

	// Configure the chip itself through the peripheral controller (MSPI3).
	if !initAPHexPSRAM(psramMspiPeriBase) {
		return "sync read/write mismatch, chip not detected"
	}

	// Now configure the system controller (MSPI2) for cache/AXI access and map
	// the physical pages into the CPU data address space.
	configurePSRAMController()
	mapPSRAM(psramVAddrBase, 0, uint32(psramDeclaredSize()/psramPageSize))
	return ""
}

// The PSRAM octal-DDR bus uses dedicated IOMUX MSPI pins.  The pad drive
// strength and the DQS pads are set in the IOMUX_MSPI_PIN block.

// IOMUX MSPI PSRAM pin register block base and field layout.
const (
	iomuxMspiPinBase = 0x500E1200 // DR_REG_HPPERIPH1_BASE(0x500C0000)+0x21200
	psramDrvSh       = 12         // PSRAM_*_DRV[2] bit offset in the data/control pins
	psramDqsDrvSh    = 15         // PSRAM_DQS_*_DRV[2] bit offset
	psramDqsXPD      = 1 << 0     // PSRAM_DQS_*_XPD (power up DQS pad)
)

// PSRAM data/control pin register offsets within the IOMUX MSPI pin block.
var psramPinOffsets = [20]uint32{
	0x1c, 0x20, 0x24, 0x28, 0x2c, 0x30, 0x34, 0x38, // D,Q,WP,HD,DQ4..DQ7
	0x3c,       // DQS_0
	0x40, 0x44, // CK, CS
	0x48, 0x4c, 0x50, 0x54, 0x58, 0x5c, 0x60, 0x64, // DQ8..DQ15
	0x68, // DQS_1
}

func psramPinReg(off uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer(uintptr(iomuxMspiPinBase + off)))
}

// configPSRAMPinIO sets every PSRAM IOMUX pad drive strength (drv, matches
// ESP-IDF mspi_timing_ll_pin_drv_set(drv=2)) and powers up the DQS pads
// (mspi_timing_ll_enable_dqs(true)).
func configPSRAMPinIO(drv uint8) {
	for _, off := range psramPinOffsets {
		reg := psramPinReg(off)
		if off == 0x3c || off == 0x68 { // DQS_0 / DQS_1: drive at bit 15 + XPD
			reg.ReplaceBits(uint32(drv), 0x3<<psramDqsDrvSh, psramDqsDrvSh)
			reg.SetBits(psramDqsXPD)
		} else {
			reg.ReplaceBits(uint32(drv), 0x3<<psramDrvSh, psramDrvSh)
		}
	}
}
