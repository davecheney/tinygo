/* See Rust for a more complete reference: https://github.com/rp-rs/rp-hal/blob/main/rp235x-hal-examples/memory.x */
MEMORY
{
   /* 2MiB safe default.  */
    FLASH : ORIGIN = 0x10000000, LENGTH = __flash_size
    /* RAM consists of 8 banks, SRAM0..SRAM7 with striped mapping. */
    SRAM : ORIGIN = 0x20000000, LENGTH = 512k
    /* Banks 8 and 9 use direct mapping which can be
       specialized for applications where predictable access time is beneficial.
       i.e: Separate stacks for core0 and core1. */
    SRAM4 : ORIGIN = 0x20080000, LENGTH = 4k
    SRAM5 : ORIGIN = 0x20081000, LENGTH = 4k
    FLASH_TEXT (rx) : ORIGIN = 0x10000000, LENGTH = __flash_size
    RAM (rwx)       : ORIGIN = 0x20000000, LENGTH = 512k
    /* QMI window 1 (CS1), cached alias. Empty unless the target sets __psram_size. */
    PSRAM           : ORIGIN = 0x11000000, LENGTH = __psram_size
}

_stack_size = 2K;

SECTIONS
{
    /* Zeroed during startup, after PSRAM is initialized. */
    .psram_bss (NOLOAD) :
    {
        . = ALIGN(8);
        __psram_bss_start = .;
        *(.psram_bss .psram_bss.*)
        . = ALIGN(8);
        __psram_bss_end = .;
    } > PSRAM

    /* Not initialized at all. Contents are undefined after reset. */
    .psram_noinit (NOLOAD) :
    {
        . = ALIGN(8);
        __psram_noinit_start = .;
        *(.psram_noinit .psram_noinit.*)
        . = ALIGN(8);
        __psram_noinit_end = .;
    } > PSRAM
}

__psram_start = ORIGIN(PSRAM);
__psram_end = ORIGIN(PSRAM) + LENGTH(PSRAM);

INCLUDE "targets/arm.ld"
