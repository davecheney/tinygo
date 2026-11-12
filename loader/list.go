package loader

import (
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

// List returns a ready-to-run *exec.Cmd for running the `go list` command with
// the configuration used for TinyGo.
func List(config *compileopts.Config, extraArgs, pkgs []string) (*exec.Cmd, error) {
	goroot, err := GetCachedGoroot(config)
	if err != nil {
		return nil, err
	}
	args := append([]string{"list"}, extraArgs...)
	if len(config.BuildTags()) != 0 {
		args = append(args, "-tags", strings.Join(config.BuildTags(), " "))
	}
	args = append(args, pkgs...)
	cmd := exec.Command(filepath.Join(goenv.Get("GOROOT"), "bin", "go"), args...)
	cmd.Env = slices.Concat(config.Options.Env, []string{"GOROOT=" + goroot, "GOOS=" + config.GOOS(), "GOARCH=" + config.GOARCH(), "CGO_ENABLED=1"})
	if config.Options.Directory != "" {
		cmd.Dir = config.Options.Directory
		cmd.Env = appendPWD(cmd.Env, cmd.Dir)
	}
	return cmd, nil
}

// appendPWD sets PWD to dir, as exec.Cmd does when Env is nil.
func appendPWD(env []string, dir string) []string {
	if abs, err := filepath.Abs(dir); err == nil {
		return append(env, "PWD="+abs)
	}
	return env
}
