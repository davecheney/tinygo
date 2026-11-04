package builder

import (
	"debug/dwarf"
	"debug/elf"
	"os"
	"path/filepath"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

func TestTrimPathCFlagsMapTinyGoSource(t *testing.T) {
	for _, tc := range []struct {
		name    string
		overlap bool
	}{
		{name: "separate"},
		{name: "cache-prefix", overlap: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("XDG_CACHE_HOME", home)
			t.Setenv("LocalAppData", home)
			cache := goenv.Get("GOCACHE")
			tinygoRoot := filepath.Join(filepath.Dir(cache), "source")
			if tc.overlap {
				tinygoRoot = cache + "-source"
			}
			for _, name := range []string{
				"src/runtime/internal/sys/zversion.go",
				"src/device/arm/arm.go",
			} {
				filename := filepath.Join(tinygoRoot, filepath.FromSlash(name))
				if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filename, nil, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("TINYGOROOT", tinygoRoot)

			source := filepath.Join(tinygoRoot, "probe.c")
			if err := os.WriteFile(source, []byte("int probe(void) { return 1; }\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			config := &compileopts.Config{
				Options: &compileopts.Options{TrimPath: true, Opt: "0"},
				Target: &compileopts.TargetSpec{
					Triple: "x86_64-unknown-linux-gnu", GOOS: "linux", GOARCH: "amd64",
				},
			}
			objectPath := filepath.Join(home, "probe.o")
			flags := append([]string{"-c", "-o", objectPath, source}, config.CFlags(false)...)
			if err := runCCompiler(flags...); err != nil {
				t.Fatal(err)
			}
			_, unit := readELFCompileUnit(t, objectPath)
			want := config.CSourcePath("github.com/tinygo-org/tinygo/probe.c")
			if got := unit.Val(dwarf.AttrName); got != want {
				t.Errorf("DWARF source path = %q, want %q", got, want)
			}
		})
	}
}

func TestLibraryPrefixMap(t *testing.T) {
	for _, tc := range []struct {
		name     string
		trimPath bool
	}{
		{name: "untrimmed"},
		{name: "trimmed", trimPath: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("XDG_CACHE_HOME", home)
			t.Setenv("LocalAppData", home)
			cache := goenv.Get("GOCACHE")
			tmpdir := filepath.Join(cache, "build")
			if err := os.MkdirAll(tmpdir, 0o700); err != nil {
				t.Fatal(err)
			}
			dir := filepath.Join(tmpdir, "build-lib-prefixmap")
			lib := &Library{
				name:           "prefixmap",
				cflags:         func(string, string) []string { return nil },
				sourceDir:      func() string { return dir },
				librarySources: func(string, bool) ([]string, error) { return []string{"probe.c"}, nil },
			}
			config := &compileopts.Config{
				Options: &compileopts.Options{TrimPath: tc.trimPath},
				Target: &compileopts.TargetSpec{
					Triple: "x86_64-unknown-linux-gnu", GOOS: "linux", GOARCH: "amd64",
				},
			}
			job, abort, err := lib.load(config, tmpdir)
			if err != nil {
				t.Fatal(err)
			}
			defer abort()
			if err := os.WriteFile(filepath.Join(dir, "probe.c"), []byte("int probe(void) { return 1; }\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := job.dependencies[0].run(job.dependencies[0]); err != nil {
				t.Fatal(err)
			}
			_, unit := readELFCompileUnit(t, filepath.Join(dir, "probe.c.o"))
			want := filepath.Join(os.TempDir(), "tinygo-prefixmap", "probe.c")
			if tc.trimPath {
				want = config.CSourcePath("github.com/tinygo-org/tinygo/.tinygo-build/prefixmap/probe.c")
			}
			got, ok := unit.Val(dwarf.AttrName).(string)
			if !ok || filepath.ToSlash(got) != filepath.ToSlash(want) {
				t.Errorf("DWARF source path = %q, want %q", got, want)
			}
		})
	}
}

func readELFCompileUnit(t *testing.T, path string) (*dwarf.Data, *dwarf.Entry) {
	t.Helper()
	object, err := elf.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer object.Close()
	debugInfo, err := object.DWARF()
	if err != nil {
		t.Fatal(err)
	}
	unit, err := debugInfo.Reader().Next()
	if err != nil {
		t.Fatal(err)
	}
	if unit == nil || unit.Tag != dwarf.TagCompileUnit {
		t.Fatalf("expected a DWARF compile unit, got %v", unit)
	}
	return debugInfo, unit
}
