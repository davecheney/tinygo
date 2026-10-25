package main

import "embed"

// Missing compiler-rt sources require `make llvm-source`.
// TestCompilerRTSourceReferences keeps these paths in sync with builder/builtins.go.
//
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_cdcmp.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_cfcmp.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_dcmp.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_fcmp.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_idivmod.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_ldivmod.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_memcmp.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_memcpy.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_memmove.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_memset.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_uidivmod.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_uldivmod.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_uread4.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_uread8.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_uwrite4.S
//go:embed llvm-project/compiler-rt/lib/builtins/arm/aeabi_uwrite8.S
//go:embed llvm-project/compiler-rt/lib/builtins/avr/divmodhi4.S
//go:embed llvm-project/compiler-rt/lib/builtins/avr/divmodqi4.S
//go:embed llvm-project/compiler-rt/lib/builtins/avr/mulhi3.S
//go:embed llvm-project/compiler-rt/lib/builtins/avr/mulqi3.S
//go:embed llvm-project/compiler-rt/lib/builtins/avr/udivmodhi4.S
//go:embed llvm-project/compiler-rt/lib/builtins/avr/udivmodqi4.S
//go:embed llvm-project/compiler-rt/lib/builtins/i386/chkstk.S
//go:embed llvm-project/compiler-rt/lib/builtins/i386/chkstk2.S
//go:embed llvm-project/compiler-rt/lib/builtins/x86_64/chkstk.S
//go:embed llvm-project/compiler-rt/lib/builtins/aarch64/chkstk.S
var _ embed.FS
