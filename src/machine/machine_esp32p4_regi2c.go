//go:build esp32p4

package machine

import "device/esp"

// regi2c (analog I2C master) register field layout of I2C0_CTRL (LP_I2C_ANA_MST).
const (
	r2cSlaveIDSh = 0  // bits [7:0] slave id (block)
	r2cAddrSh    = 8  // bits [15:8] register address
	r2cDataSh    = 16 // bits [23:16] data
	r2cWriteBit  = 1 << 24
	r2cBusy      = 1 << 25

	// ANA_CONF2 block select bits (from regi2c_impl.c).
	// Each analog I2C slave has a dedicated routing bit in ANA_CONF2.
	r2cBlockMSPI = 0x63 // MSPI/MPLL
	r2cBlockSPLL = 0x66 // System PLL
	r2cBlockCPLL = 0x67 // CPU PLL

	r2cSelMSPI = 1 << 9  // REGI2C_MSPI_XTAL_MST_SEL
	r2cSelSPLL = 1 << 5  // REGI2C_PLL_SYS_MST_SEL
	r2cSelCPLL = 1 << 11 // REGI2C_PLL_CPU_MST_SEL
)

// regi2cTimeout bounds the busy-wait on the analog I2C master controller so a
// stuck bus returns instead of hanging the bring-up.
const regi2cMaxBusy = 100000

// regi2cEnableGate enables the LP analog I2C master clock (regi2c) used to
// reach the MPLL analog registers on the MSPI block.
func regi2cEnableGate() {
	esp.LP_PERI.CLK_EN.SetBits(esp.LPPERI_CLK_EN_CK_EN_LP_I2CMST)
}

// regi2cSelectBlock routes the analog-I2C master to the specified block.
// The block parameter is the I2C slave ID (e.g. r2cBlockMSPI, r2cBlockCPLL).
func regi2cSelectBlock(block uint8) {
	esp.LP_I2C_ANA_MST.ANA_CONF1.Set(0)
	esp.LP_I2C_ANA_MST.ANA_CONF2.Set(0)
	var sel uint32
	switch block {
	case r2cBlockMSPI:
		sel = r2cSelMSPI
	case r2cBlockSPLL:
		sel = r2cSelSPLL
	case r2cBlockCPLL:
		sel = r2cSelCPLL
	default:
		sel = r2cSelMSPI // fallback
	}
	esp.LP_I2C_ANA_MST.ANA_CONF2.Set(sel)
}

// regi2cWrite performs a single analog-I2C-master register write to `block`
// (slave id), `reg` (address) with `data`, mirroring ESP-IDF _regi2c_impl_write
// on the P4 LP_I2C_ANA_MST I2C0 controller.
func regi2cWrite(block, reg, data uint8) {
	regi2cSelectBlock(block)
	ctrl := &esp.LP_I2C_ANA_MST.I2C0_CTRL
	for i := 0; ctrl.Get()&r2cBusy != 0 && i < regi2cMaxBusy; i++ {
	}
	ctrl.Set(uint32(block)<<r2cSlaveIDSh |
		uint32(reg)<<r2cAddrSh |
		r2cWriteBit |
		uint32(data)<<r2cDataSh)
	for i := 0; ctrl.Get()&r2cBusy != 0 && i < regi2cMaxBusy; i++ {
	}
}
