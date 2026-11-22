//go:build esp.plic

package esp

// DR_REG_PLIC_MX_BASE. The INTPRI block at 0x600c5000 is not connected to the CPU.
const intcBase = 0x20001000

// The mie CSR also gates each line. esp-hal sets all bits.
const mieGate = true
