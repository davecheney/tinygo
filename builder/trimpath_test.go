package builder

import (
	"debug/dwarf"
	"debug/elf"
	"errors"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/loader"
)

func TestLibraryTrimPathPrefixMapOrder(t *testing.T) {
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
		Options: &compileopts.Options{TrimPath: true},
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
	if err := runJobs(job, nil); err != nil {
		t.Fatal(err)
	}
	_, unit := readELFCompileUnit(t, filepath.Join(dir, "probe.c.o"))
	want := config.CSourcePath("github.com/tinygo-org/tinygo/.tinygo-build/prefixmap/probe.c")
	if got := unit.Val(dwarf.AttrName); got != want {
		t.Errorf("DWARF source path = %q, want %q", got, want)
	}
}

func TestCGoTrimPathDWARF(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "b")
	home := filepath.Join(root, "bc")
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
	t.Setenv("LocalAppData", home)
	t.Setenv("GOWORK", "off")
	t.Setenv("CGO_CPPFLAGS", "")
	t.Setenv("CGO_CFLAGS", "")
	t.Setenv("CGO_CXXFLAGS", "")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/probe\n\ngo 1.26\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "probe.go"), []byte("package probe\n/* int value; */\nimport \"C\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	header := filepath.Join(goenv.Get("GOCACHE"), "include", "probe.h")
	if err := os.MkdirAll(filepath.Dir(header), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(header, []byte("static int cached(void) { return 1; }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(dir, "probe.c")
	code := "#include " + strconv.Quote(filepath.ToSlash(header)) + "\nint probe(void) { return cached(); }\n"
	if err := os.WriteFile(source, []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}
	_, minor, err := goenv.GetGorootVersion()
	if err != nil {
		t.Fatal(err)
	}
	config := &compileopts.Config{
		Options: &compileopts.Options{Directory: dir, Opt: "0", TrimPath: true},
		Target: &compileopts.TargetSpec{
			Triple: "x86_64-unknown-linux-gnu", GOOS: "linux", GOARCH: "amd64",
		},
		GoMinorVersion: minor,
	}
	program, err := loader.Load(config, ".", types.Config{})
	if err != nil {
		t.Fatal(err)
	}
	pkg := program.MainPkg()
	if err := pkg.Parse(); err != nil {
		t.Fatal(err)
	}
	object := filepath.Join(root, "probe.o")
	flags := append([]string{"-c", "-o", object, source}, pkg.CGoHeaderCFlags(dir)...)
	if err := runCCompiler(flags...); err != nil {
		t.Fatal(err)
	}
	debugInfo, unit := readELFCompileUnit(t, object)
	lines, err := debugInfo.LineReader(unit)
	if err != nil {
		t.Fatal(err)
	}
	want := config.CSourcePath("tinygo-cache/include/probe.h")
	files := make(map[string]bool)
	for {
		var line dwarf.LineEntry
		err := lines.Next(&line)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if line.File != nil {
			files[line.File.Name] = true
		}
	}
	if !files[want] {
		t.Errorf("DWARF files = %v, want %q", files, want)
	}
	for file := range files {
		if filepath.Base(file) == "probe.h" && file != want {
			t.Errorf("incorrect DWARF header path %q", file)
		}
	}
}

func TestLibraryTrimPathInvalidatesOldCache(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
	t.Setenv("LocalAppData", home)
	cache := goenv.Get("GOCACHE")
	oldDir := filepath.Join(cache, "prefixmap+trimpath-v1-x86_64-unknown-linux-gnu")
	if err := os.MkdirAll(oldDir, 0o700); err != nil {
		t.Fatal(err)
	}
	oldArchive := filepath.Join(oldDir, "lib.a")
	if err := os.WriteFile(oldArchive, []byte("old cache"), 0o600); err != nil {
		t.Fatal(err)
	}
	lib := &Library{
		name:           "prefixmap",
		cflags:         func(string, string) []string { return nil },
		sourceDir:      func() string { return home },
		librarySources: func(string, bool) ([]string, error) { return nil, nil },
	}
	config := &compileopts.Config{
		Options: &compileopts.Options{TrimPath: true},
		Target: &compileopts.TargetSpec{
			Triple: "x86_64-unknown-linux-gnu", GOOS: "linux", GOARCH: "amd64",
		},
	}
	job, abort, err := lib.load(config, home)
	if err != nil {
		t.Fatal(err)
	}
	defer abort()
	if job.result == oldArchive {
		t.Errorf("reused old archive %q", oldArchive)
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
