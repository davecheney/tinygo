
/* The Tessel bootloader owns the first 64K of SPI flash. */
MEMORY
{
    FLASH_TEXT (rx) : ORIGIN = 0x14010000, LENGTH = 2M - 64K
    RAM (xrw)       : ORIGIN = 0x10000000, LENGTH = 96K
}

/* The bootloader only starts firmware with this word in vector slot 7.
 * See boot/main.c in https://github.com/tessel/t1-firmware */
__isr_vector_signature = 0x5A5A5A5A;

_stack_size = 4K;

INCLUDE "targets/arm.ld"
