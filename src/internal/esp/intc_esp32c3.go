//go:build esp.intc && esp32c3

package esp

// INTERRUPT_CORE0_CPU_INT_ENABLE_REG.
const intcBase = 0x600c2104

const mieGate = false
