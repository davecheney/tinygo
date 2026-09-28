package builder

import (
	"debug/elf"
	"strings"
	"testing"
)

// Test that .psram_bss and .psram_noinit are placed in the RP2350 QMI window
// 1 on a PSRAM target, and that using them on a target without PSRAM fails to
// link.
func TestRP2350PSRAMSections(t *testing.T) {
	t.Parallel()

	result, err := buildBinaryInDir("presto", "examples/rp2350-psram", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f, err := elf.Open(result.Executable)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	const (
		base = 0x11000000
		end  = base + 8*1024*1024
	)
	for _, tc := range []struct {
		name string
		min  uint64
	}{
		{".psram_bss", 2 * 480 * 480 * 2},
		{".psram_noinit", 64 * 1024 * 4},
	} {
		s := f.Section(tc.name)
		if s == nil {
			t.Errorf("section %s not found", tc.name)
			continue
		}
		if s.Type != elf.SHT_NOBITS {
			t.Errorf("section %s has type %v, want SHT_NOBITS", tc.name, s.Type)
		}
		if s.Addr < base || s.Addr+s.Size > end {
			t.Errorf("section %s at %#x size %#x is outside PSRAM", tc.name, s.Addr, s.Size)
		}
		if s.Size < tc.min {
			t.Errorf("section %s size %#x, want at least %#x", tc.name, s.Size, tc.min)
		}
	}
	for _, p := range f.Progs {
		if p.Type == elf.PT_LOAD && p.Paddr >= base && p.Paddr < end && p.Filesz != 0 {
			t.Errorf("PSRAM segment at %#x has %d bytes of file data", p.Paddr, p.Filesz)
		}
	}

	syms, err := f.Symbols()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]uint64{"__psram_start": base, "__psram_end": end}
	for _, sym := range syms {
		if v, ok := want[sym.Name]; ok {
			if sym.Value != v {
				t.Errorf("%s = %#x, want %#x", sym.Name, sym.Value, v)
			}
			delete(want, sym.Name)
		}
	}
	for name := range want {
		t.Errorf("symbol %s not found", name)
	}

	_, err = buildBinaryInDir("pico2", "examples/rp2350-psram", t.TempDir())
	if err == nil {
		t.Fatal("build for pico2 succeeded, want PSRAM region overflow")
	}
	if !strings.Contains(err.Error(), "PSRAM") {
		t.Errorf("build for pico2 failed with unexpected error: %v", err)
	}
}
