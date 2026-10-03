package builder

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestCommandCandidates(t *testing.T) {
	for _, major := range []string{"22", "23", "24"} {
		for _, arch := range []string{"amd64", "arm64"} {
			t.Run("darwin/"+arch+"/"+major, func(t *testing.T) {
				prefix := "/usr/local/opt/"
				if arch == "arm64" {
					prefix = "/opt/homebrew/opt/"
				}
				got := commandCandidates("darwin", arch, major)
				want := map[string][]string{
					"clang": {"clang-" + major, prefix + "llvm@" + major + "/bin/clang-" + major, prefix + "llvm/bin/clang-" + major},
					"lldb":  {"lldb-" + major, "lldb", prefix + "llvm@" + major + "/bin/lldb", prefix + "llvm/bin/lldb"},
				}
				for _, name := range []string{"ld.lld", "wasm-ld"} {
					want[name] = []string{
						name + "-" + major, name,
						prefix + "lld@" + major + "/bin/" + name, prefix + "lld/bin/" + name,
						prefix + "llvm@" + major + "/bin/" + name, prefix + "llvm/bin/" + name,
					}
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("got %v, want %v", got, want)
				}
			})
		}
	}
	for _, goos := range []string{"linux", "windows", "freebsd"} {
		t.Run(goos, func(t *testing.T) {
			want := map[string][]string{
				"clang":   {"clang-23"},
				"ld.lld":  {"ld.lld-23", "ld.lld"},
				"wasm-ld": {"wasm-ld-23", "wasm-ld"},
				"lldb":    {"lldb-23", "lldb"},
			}
			switch goos {
			case "windows":
				want["clang"] = append(want["clang"], "clang", `C:\Program Files\LLVM\bin\clang.exe`)
				want["ld.lld"] = append(want["ld.lld"], "lld", `C:\Program Files\LLVM\bin\lld.exe`)
				want["wasm-ld"] = append(want["wasm-ld"], `C:\Program Files\LLVM\bin\wasm-ld.exe`)
				want["lldb"] = append(want["lldb"], `C:\Program Files\LLVM\bin\lldb.exe`)
			case "freebsd":
				for name := range want {
					cmd := name
					if name == "clang" {
						cmd += "-23"
					}
					want[name] = append(want[name], "/usr/local/llvm23/bin/"+cmd)
				}
			}
			if got := commandCandidates(goos, "amd64", "23"); !reflect.DeepEqual(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}

func TestLookupLinker(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake linkers require /bin/sh")
	}
	for _, name := range []string{"ld.lld", "wasm-ld"} {
		for _, major := range []string{"22", "23"} {
			t.Run(name+"/"+major, func(t *testing.T) {
				for _, scenario := range []string{
					"mismatched-path-versioned-formula",
					"mismatched-path-unversioned-formula",
					"matching-path",
					"matching-versioned-path",
					"mismatched-versioned-path",
					"legacy-llvm-formula",
					"absent",
					"mismatched",
					"version-failure",
					"invalid-version",
					"not-executable",
				} {
					t.Run(scenario, func(t *testing.T) {
						root := t.TempDir()
						pathDir := filepath.Join(root, "path")
						if err := os.Mkdir(pathDir, 0o755); err != nil {
							t.Fatal(err)
						}
						t.Setenv("PATH", pathDir)
						versioned := name + "-" + major
						formula := filepath.Join(root, "lld@"+major, "bin", name)
						unversioned := filepath.Join(root, "lld", "bin", name)
						legacy := filepath.Join(root, "llvm@"+major, "bin", name)
						missing := filepath.Join(root, "absent", name)
						candidates := []string{versioned, name, missing, formula, unversioned, legacy}
						matching := "printf 'Homebrew LLD " + major + ".1.2 (compatible with GNU linkers)\\n'"
						otherMajor := "22"
						if major == "22" {
							otherMajor = "23"
						}
						incompatible := "printf 'LLD " + otherMajor + ".1.0\\n'"
						want := ""
						var wantErrors []string
						switch scenario {
						case "mismatched-path-versioned-formula":
							writeFakeCommand(t, filepath.Join(pathDir, name), incompatible)
							writeFakeCommand(t, formula, matching)
							writeFakeCommand(t, unversioned, incompatible)
							want = formula
						case "mismatched-path-unversioned-formula":
							writeFakeCommand(t, filepath.Join(pathDir, name), incompatible)
							writeFakeCommand(t, unversioned, matching)
							want = unversioned
						case "matching-path":
							writeFakeCommand(t, filepath.Join(pathDir, name), matching)
							writeFakeCommand(t, formula, matching)
							want = name
						case "matching-versioned-path":
							writeFakeCommand(t, filepath.Join(pathDir, versioned), matching)
							writeFakeCommand(t, filepath.Join(pathDir, name), matching)
							want = versioned
						case "mismatched-versioned-path":
							writeFakeCommand(t, filepath.Join(pathDir, versioned), incompatible)
							writeFakeCommand(t, filepath.Join(pathDir, name), matching)
							want = name
						case "legacy-llvm-formula":
							writeFakeCommand(t, legacy, matching)
							want = legacy
						case "absent":
							wantErrors = []string{"no " + name + " matching LLVM " + major + " found", missing}
						case "mismatched":
							writeFakeCommand(t, filepath.Join(pathDir, name), incompatible)
							writeFakeCommand(t, formula, incompatible)
							wantErrors = []string{"no " + name + " matching LLVM " + major + " found", "reports LLD " + otherMajor + ", need LLVM " + major, formula}
						case "version-failure":
							writeFakeCommand(t, filepath.Join(pathDir, name), "printf 'version failed\\n' >&2; exit 1")
							writeFakeCommand(t, formula, matching)
							wantErrors = []string{"could not check LLD version", "exit status 1", "version failed"}
						case "invalid-version":
							writeFakeCommand(t, filepath.Join(pathDir, name), "printf 'not an LLD version\\n'")
							writeFakeCommand(t, formula, matching)
							wantErrors = []string{"could not parse LLD version", "not an LLD version"}
						case "not-executable":
							if err := os.MkdirAll(filepath.Dir(formula), 0o755); err != nil {
								t.Fatal(err)
							}
							if err := os.WriteFile(formula, nil, 0o644); err != nil {
								t.Fatal(err)
							}
							wantErrors = []string{"permission denied"}
						}
						got, err := lookupCommand(name, candidates, major)
						if len(wantErrors) != 0 {
							if err == nil {
								t.Fatalf("got %q without an error", got)
							}
							for _, text := range wantErrors {
								if !strings.Contains(err.Error(), text) {
									t.Errorf("error %q does not contain %q", err, text)
								}
							}
						} else if err != nil || got != want {
							t.Errorf("got %q, %v; want %q, nil", got, err, want)
						}
					})
				}
			})
		}
	}
}

func TestLookupNonLinker(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake commands require /bin/sh")
	}
	for _, name := range []string{"clang", "lldb"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), name)
			writeFakeCommand(t, path, "exit 1")
			t.Setenv("PATH", filepath.Dir(path))
			got, err := lookupCommand(name, []string{name}, "23")
			if err != nil || got != name {
				t.Errorf("got %q, %v; want %q, nil", got, err, name)
			}
		})
	}
}

func TestLinkerMajorVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake linkers require /bin/sh")
	}
	for _, name := range []string{"ld.lld", "wasm-ld"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "lld")
			args := "--version"
			if name == "ld.lld" {
				args = "-flavor gnu --version"
			}
			writeFakeCommand(t, path, `[ "$*" = "`+args+`" ] || exit 1
printf 'LLD 23.0.0git\n' >&2`)
			got, err := linkerMajorVersion(name, path)
			if err != nil || got != "23" {
				t.Errorf("got %q, %v; want 23, nil", got, err)
			}
		})
	}
}

func writeFakeCommand(t *testing.T, path, script string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}
