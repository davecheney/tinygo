package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

var testRuntimeStress = flag.Bool("runtime-stress", false, "run the opt-in runtime stress matrix")

func runFixtureProcess(template *exec.Cmd, executable string, combined bool, stdout, stderr *bytes.Buffer, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	process := exec.CommandContext(ctx, template.Path, template.Args[1:]...)
	process.Dir, process.Env, process.Stdin = template.Dir, template.Env, template.Stdin
	process.Stdout = newOutputWriter(stdout, executable)
	process.Stderr = stderr
	if combined {
		process.Stdout, process.Stderr = stdout, stdout
	}
	err := process.Run()
	if ctx.Err() != nil {
		return fmt.Errorf("fixture exceeded %s: %w", timeout, ctx.Err())
	}
	return err
}

func TestFixtureProcess(t *testing.T) {
	for _, combined := range []bool{false, true} {
		t.Run(fmt.Sprint(combined), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestFixtureProcessHelper$")
			cmd.Env = append(os.Environ(), "TINYGO_FIXTURE_HELPER=output")
			var stdout, stderr bytes.Buffer
			for range 2 {
				stdout.Reset()
				stderr.Reset()
				if err := runFixtureProcess(cmd, "", combined, &stdout, &stderr, time.Minute); err != nil {
					t.Fatal(err)
				}
				if combined {
					if !strings.Contains(stdout.String(), "stdout") || !strings.Contains(stdout.String(), "stderr") || stderr.Len() != 0 {
						t.Fatalf("combined stdout=%q stderr=%q", &stdout, &stderr)
					}
				} else if stdout.String() != "stdout" || stderr.String() != "stderr\n" {
					t.Fatalf("split stdout=%q stderr=%q", &stdout, &stderr)
				}
			}
		})
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestFixtureProcessHelper$")
	cmd.Env = append(os.Environ(), "TINYGO_FIXTURE_HELPER=sleep")
	var stdout, stderr bytes.Buffer
	if err := runFixtureProcess(cmd, "", false, &stdout, &stderr, 50*time.Millisecond); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline, got %v", err)
	}
	cmd.Env = append(os.Environ(), "TINYGO_FIXTURE_HELPER=fail")
	err := runFixtureProcess(cmd, "", false, &stdout, &stderr, time.Minute)
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 2 || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected exit 2, got %v", err)
	}
}

func TestFixtureProcessHelper(t *testing.T) {
	switch os.Getenv("TINYGO_FIXTURE_HELPER") {
	case "output":
		fmt.Fprint(os.Stdout, "stdout")
		fmt.Fprintln(os.Stderr, "stderr")
		os.Exit(0)
	case "sleep":
		time.Sleep(10 * time.Second)
		os.Exit(0)
	case "fail":
		os.Exit(2)
	}
}

func TestRuntimeStress(t *testing.T) {
	if !*testRuntimeStress {
		t.Skip("enable with -runtime-stress")
	}
	for _, gc := range []string{"conservative", "precise", "boehm"} {
		t.Run(gc, func(t *testing.T) {
			for _, opt := range []string{"0", "1", "2", "s", "z"} {
				t.Run(opt, func(t *testing.T) {
					for _, mode := range []string{"collect", "clobber", "both"} {
						if gc == "boehm" && mode != "collect" {
							continue
						}
						t.Run(mode, func(t *testing.T) {
							options := optionsFromTarget(*testTarget, sema)
							options.GC, options.Opt = gc, opt
							options.Tags = []string{"runtime_asserts"}
							if mode != "clobber" {
								options.Tags = append(options.Tags, "runtime_gcstress")
							}
							if mode != "collect" {
								options.Tags = append(options.Tags, "runtime_clobberfree")
							}
							emuCheck(t, options)
							fixtures := []string{"gc.go", "gc-register-root.go", "gc-root-stress.go", "zeroalloc.go"}
							if mode != "clobber" {
								fixtures = append(fixtures, "gc-stress.go")
							}
							if gc != "boehm" {
								fixtures = append(fixtures, "finalizerinvariants.go")
							}
							for _, name := range fixtures {
								t.Run(name, func(t *testing.T) {
									runTest(name, options, t, nil, nil)
								})
							}
						})
					}
				})
			}
		})
	}
}
