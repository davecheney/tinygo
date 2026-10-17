package main

import (
	"bytes"
	"debug/elf"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/loader"
)

func TestGODEBUGOverlay(t *testing.T) {
	for _, target := range []string{"", "pico"} {
		t.Run(target, func(t *testing.T) {
			options := optionsFromTarget(target, sema)
			config, err := builder.NewConfig(&options)
			if err != nil {
				t.Fatal(err)
			}
			root, err := loader.GetCachedGoroot(config)
			if err != nil {
				t.Fatal(err)
			}
			source, err := filepath.EvalSymlinks(filepath.Join(root, "src/internal/godebug/godebug.go"))
			if err != nil {
				t.Fatal(err)
			}
			expectedRoot := goenv.Get("GOROOT")
			if target == "pico" {
				expectedRoot = goenv.Get("TINYGOROOT")
			}
			want, err := filepath.EvalSymlinks(filepath.Join(expectedRoot, "src/internal/godebug/godebug.go"))
			if err != nil {
				t.Fatal(err)
			}
			if source != want {
				t.Fatalf("using %s, want %s", source, want)
			}
		})
	}
}

func TestBaremetalGODEBUGInterpreter(t *testing.T) {
	for _, tc := range []struct {
		name, defaults, env, want string
	}{
		{"absent", "", "", ""},
		{"build default", "probe=one", "", "one"},
		{"startup override", "probe=one", "GODEBUG=probe=two", "two"},
		{"empty startup", "probe=one", "GODEBUG=", "one"},
		{"explicit empty", "probe=one", "GODEBUG=probe=", ""},
		{"last default", "probe=one,probe=two", "", "two"},
		{"last startup", "", "GODEBUG=probe=one,probe=two", "two"},
		{"malformed entry", "", "GODEBUG=probe=one,probe", "one"},
		{"equals in value", "", "GODEBUG=probe=a=b", "a=b"},
		{"first environment entry", "", "GODEBUG=probe=one\x00GODEBUG=probe=two", "one"},
		{"other environment entries", "", "OTHER=x\x00GODEBUG=probe=two\x00LAST=x", "two"},
		{"invalid bisect", "", "GODEBUG=probe=one#invalid", "one"},
		{"disabled bisect", "", "GODEBUG=probe=one#n", ""},
		{"enabled bisect without stack", "", "GODEBUG=probe=one#y", ""},
		{"FIPS default", "fips140=on", "", ""},
		{"FIPS startup", "", "GODEBUG=fips140=on", ""},
		{"FIPS only", "", "GODEBUG=fips140=only", ""},
		{"FIPS disabled", "fips140=on", "GODEBUG=fips140=off", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options := optionsFromTarget("pico", sema)
			options.Opt = "0"
			options.GlobalValues = map[string]map[string]string{
				"runtime": {
					"osEnv": tc.env,
				},
				"main": {"expected": tc.want},
			}
			if tc.defaults != "" {
				options.GlobalValues["runtime"]["godebugDefault"] = "tarinsecurepath=0," + tc.defaults
			}
			config, err := builder.NewConfig(&options)
			if err != nil {
				t.Fatal(err)
			}
			tmpdir := t.TempDir()
			output := filepath.Join(tmpdir, "program.ll")
			_, err = builder.Build("testdata/godebug-baremetal.go", output, tmpdir, config)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"initial", "documented", "afterSetenv", "afterUnsetenv", "afterClearenv"} {
				pattern := `(?m)^@main\.` + name + ` = internal global i1 true\b`
				if !regexp.MustCompile(pattern).Match(data) {
					t.Errorf("%s was not evaluated to true by the initializer interpreter", name)
				}
			}
			wantEnabled := tc.name == "FIPS default" || tc.name == "FIPS startup" || tc.name == "FIPS only"
			pattern := fmt.Sprintf(`(?m)^@"crypto/internal/fips140\.Enabled" = internal global i1 %t\b`, wantEnabled)
			if !regexp.MustCompile(pattern).Match(data) {
				t.Errorf("FIPS enabled was not evaluated to %t", wantEnabled)
			}
			if wantEnabled && !bytes.Contains(data, []byte("FIPS 140-3 mode is incompatible with the purego build tag")) {
				t.Error("explicit FIPS request lost its unsupported-platform diagnostic")
			}
		})
	}
}

func TestBaremetalGODEBUGInvalid(t *testing.T) {
	for _, globals := range []map[string]string{
		{"godebugDefault": "fips140=invalid"},
		{"osEnv": "GODEBUG=fips140=invalid"},
	} {
		options := optionsFromTarget("pico", sema)
		options.Opt = "0"
		options.GlobalValues = map[string]map[string]string{"runtime": globals}
		config, err := builder.NewConfig(&options)
		if err != nil {
			t.Fatal(err)
		}
		tmpdir := t.TempDir()
		output := filepath.Join(tmpdir, "program.ll")
		if _, err := builder.Build("testdata/godebug-baremetal.go", output, tmpdir, config); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(data, []byte("fips140: unknown GODEBUG setting fips140=")) ||
			!bytes.Contains(data, []byte(`c"fips140=invalid"`)) ||
			!bytes.Contains(data, []byte(`call void @"crypto/internal/fips140.init#1"`)) {
			t.Error("invalid FIPS setting lost its diagnostic")
		}
	}
}

func TestBaremetalGODEBUGDeadCode(t *testing.T) {
	for _, target := range []struct{ name, scheduler string }{
		{"pico", "tasks"},
		{"pico2", "tasks"},
		{"pico2", "cores"},
	} {
		for _, opt := range []string{"s", "1"} {
			t.Run(target.name+"/"+target.scheduler+"/"+opt, func(t *testing.T) {
				var baseline [3]uint64
				for _, program := range []string{"plain", "sha3-unused", "sha3"} {
					options := optionsFromTarget(target.name, sema)
					options.Opt = opt
					options.Scheduler = target.scheduler
					config, err := builder.NewConfig(&options)
					if err != nil {
						t.Fatal(err)
					}
					output := filepath.Join(t.TempDir(), "program.elf")
					if err := Build("testdata/godebug-"+program+".go", output, config); err != nil {
						t.Fatal(err)
					}
					f, err := elf.Open(output)
					if err != nil {
						t.Fatal(err)
					}
					defer f.Close()
					var sizes [3]uint64
					for i, name := range []string{".text", ".data", ".bss"} {
						if section := f.Section(name); section != nil {
							sizes[i] = section.Size
						}
					}
					t.Logf("%s text/data/bss %v", program, sizes)
					if program == "plain" {
						baseline = sizes
					} else if program == "sha3-unused" && sizes != baseline {
						t.Errorf("unused SHA3 retained code or data: %v, plain %v", sizes, baseline)
					}
					symbols, err := f.Symbols()
					if err != nil {
						t.Fatal(err)
					}
					hasKeccak := false
					for _, symbol := range symbols {
						isFIPS := strings.Contains(symbol.Name, "crypto/internal/fips140")
						isAlgorithm := strings.Contains(symbol.Name, "crypto/internal/fips140/sha3.") ||
							strings.Contains(symbol.Name, "crypto/internal/fips140/sha3$") ||
							strings.Contains(symbol.Name, "crypto/internal/fips140/subtle$")
						if strings.Contains(symbol.Name, "godebug") || (isFIPS && (program != "sha3" || !isAlgorithm)) {
							t.Errorf("%s retained %s", program, symbol.Name)
						}
						hasKeccak = hasKeccak || strings.Contains(symbol.Name, "keccakF1600")
					}
					if hasKeccak != (program == "sha3") {
						t.Errorf("%s retained Keccak: %t", program, hasKeccak)
					}
				}
			})
		}
	}
}

func TestGODEBUGSHA3(t *testing.T) {
	options := optionsFromTarget("", sema)
	config, err := builder.NewConfig(&options)
	if err != nil {
		t.Fatal(err)
	}
	output := &bytes.Buffer{}
	_, err = buildAndRun("testdata/godebug-sha3.go", config, output, nil, nil, time.Minute, func(cmd *exec.Cmd, _ builder.BuildResult) error {
		return cmd.Run()
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "sha3 ok\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}
