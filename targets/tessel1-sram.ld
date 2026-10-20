
/* The Tessel bootloader loads SRAM images to 0x10000000 and runs them. */
MEMORY
{
    FLASH_TEXT (rx) : ORIGIN = 0x10000000, LENGTH = 96K
    RAM (xrw)       : ORIGIN = 0x20000000, LENGTH = 64K
}

_stack_size = 4K;

INCLUDE "targets/arm.ld"
