package compileopts

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/tinygo-org/tinygo/goenv"
)

func TestCFlagsWithPrefixMaps(t *testing.T) {
	root := goenv.Get("TINYGOROOT")
	parentMap := FilePrefixMap(filepath.Dir(root), "/_/parent")
	moduleMap := FilePrefixMap(filepath.Join(root, "nested", "module"), "/_/module")
	targetFlags := []string{moduleMap, "-DPROBE=1", parentMap}
	extraMaps := []string{moduleMap, parentMap}
	for _, trimPath := range []bool{false, true} {
		config := &Config{
			Options: &Options{TrimPath: trimPath},
			Target: &TargetSpec{
				Triple: "x86_64-unknown-linux-gnu", GOOS: "linux", GOARCH: "amd64",
				CFlags: targetFlags,
			},
		}
		for _, libclang := range []bool{false, true} {
			flags := config.CFlagsWithPrefixMaps(libclang, extraMaps)
			if !slices.Equal(flags[:len(targetFlags)], targetFlags) {
				t.Fatalf("target flags changed: %q", flags)
			}
			generated := flags[len(targetFlags):]
			if trimPath {
				rootMap := FilePrefixMap(root, config.CSourcePath("github.com/tinygo-org/tinygo"))
				parentIndex := slices.Index(generated, parentMap)
				rootIndex := slices.Index(generated, rootMap)
				moduleIndex := slices.Index(generated, moduleMap)
				if parentIndex < 0 || rootIndex <= parentIndex || moduleIndex <= rootIndex {
					t.Fatalf("nested maps are not most-specific-last: %q", generated)
				}
			} else if slices.Contains(generated, parentMap) || slices.Contains(generated, moduleMap) {
				t.Fatalf("untrimmed flags contain extra maps: %q", generated)
			}
		}
	}
	if !slices.Equal(extraMaps, []string{moduleMap, parentMap}) {
		t.Fatalf("extra maps changed: %q", extraMaps)
	}
}

func TestExtraFilesBoehm(t *testing.T) {
	target := &TargetSpec{
		GC: "precise",
		ExtraFiles: []string{
			"src/runtime/asm_tinygowasm.S",
		},
	}
	config := &Config{
		Options: &Options{},
		Target:  target,
	}

	got := config.ExtraFiles()
	want := []string{"src/runtime/asm_tinygowasm.S"}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected precise GC files: got %v, want %v", got, want)
	}

	config.Options.GC = "boehm"
	got = config.ExtraFiles()
	want = []string{
		"src/runtime/asm_tinygowasm.S",
		"src/runtime/gc_boehm.c",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected Boehm GC files: got %v, want %v", got, want)
	}
}

func TestValidateOutputExtension(t *testing.T) {
	tests := []struct {
		name      string
		target    TargetSpec
		extension string
		wantErr   string
	}{
		{
			name: "supported firmware extension",
			target: TargetSpec{
				FirmwareExts: []string{".elf", ".hex", ".uf2"},
			},
			extension: ".uf2",
		},
		{
			name: "unsupported firmware extension",
			target: TargetSpec{
				FirmwareExts: []string{".elf", ".hex", ".uf2"},
			},
			extension: ".bin",
			wantErr:   "target pico does not support firmware output extension .bin; use .elf, .hex, or .uf2",
		},
		{
			name: "object output",
			target: TargetSpec{
				FirmwareExts: []string{".elf", ".hex", ".uf2"},
			},
			extension: ".o",
		},
		{
			name: "bitcode output",
			target: TargetSpec{
				FirmwareExts: []string{".elf", ".hex", ".uf2"},
			},
			extension: ".bc",
		},
		{
			name: "LLVM IR output",
			target: TargetSpec{
				FirmwareExts: []string{".elf", ".hex", ".uf2"},
			},
			extension: ".ll",
		},
		{
			name: "unknown extension",
			target: TargetSpec{
				FirmwareExts: []string{".elf", ".hex", ".uf2"},
			},
			extension: ".unknown",
		},
		{
			name: "extensionless output",
			target: TargetSpec{
				FirmwareExts: []string{".elf", ".hex", ".uf2"},
			},
		},
		{
			name: "target without restrictions",
			target: TargetSpec{
				BinaryFormat: "esp32",
			},
			extension: ".bin",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := &Config{
				Options: &Options{Target: "pico"},
				Target:  &tc.target,
			}
			err := config.ValidateOutputExtension(tc.extension)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateOutputExtension(%q) failed: %v", tc.extension, err)
				}
				return
			}
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("ValidateOutputExtension(%q) error = %v, want %q", tc.extension, err, tc.wantErr)
			}
		})
	}
}

func TestBinaryFormatFirmwareExtensions(t *testing.T) {
	config := &Config{
		Options: &Options{},
		Target: &TargetSpec{
			BinaryFormat: "esp32",
		},
	}
	for ext, want := range map[string]string{
		".bin": "esp32",
		".hex": "hex",
		".uf2": "uf2",
		".elf": "elf",
	} {
		if got := config.BinaryFormat(ext); got != want {
			t.Errorf("BinaryFormat(%q) = %q, want %q", ext, got, want)
		}
	}
	config.Target.BinaryFormat = ""
	if got := config.BinaryFormat(".bin"); got != "bin" {
		t.Errorf("BinaryFormat(.bin) without a target format = %q, want %q", got, "bin")
	}
}
