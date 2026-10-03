package builder

import (
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"tinygo.org/x/go-llvm"
)

// Commands lists command alternatives for various operating systems. These
// commands may have a slightly different name across operating systems and
// distributions or may not even exist in $PATH, in which case absolute paths
// may be used.
var commands = commandCandidates(runtime.GOOS, runtime.GOARCH, strings.Split(llvm.Version, ".")[0])

func commandCandidates(goos, goarch, llvmMajor string) map[string][]string {
	commands := map[string][]string{}
	commands["clang"] = []string{"clang-" + llvmMajor}
	commands["ld.lld"] = []string{"ld.lld-" + llvmMajor, "ld.lld"}
	commands["wasm-ld"] = []string{"wasm-ld-" + llvmMajor, "wasm-ld"}
	commands["lldb"] = []string{"lldb-" + llvmMajor, "lldb"}
	// Add the path to a Homebrew-installed LLVM for ease of use (no need to
	// manually set $PATH).
	if goos == "darwin" {
		// The newest LLVM release may still be under Homebrew's unversioned
		// "llvm" formula, so that path is tried too.
		var homebrew string
		switch goarch {
		case "amd64":
			homebrew = "/usr/local/opt/"
		case "arm64":
			homebrew = "/opt/homebrew/opt/"
		default:
			// unknown GOARCH
			panic(fmt.Sprintf("unknown GOARCH: %s on darwin", goarch))
		}
		prefix := homebrew + "llvm@" + llvmMajor + "/bin/"
		unversionedPrefix := homebrew + "llvm/bin/"
		commands["clang"] = append(commands["clang"], prefix+"clang-"+llvmMajor, unversionedPrefix+"clang-"+llvmMajor)
		for _, name := range []string{"ld.lld", "wasm-ld"} {
			commands[name] = append(commands[name],
				homebrew+"lld@"+llvmMajor+"/bin/"+name, homebrew+"lld/bin/"+name,
				prefix+name, unversionedPrefix+name)
		}
		commands["lldb"] = append(commands["lldb"], prefix+"lldb", unversionedPrefix+"lldb")
	}
	// Add the path for when LLVM was installed with the installer from
	// llvm.org, which by default doesn't add LLVM to the $PATH environment
	// variable.
	if goos == "windows" {
		commands["clang"] = append(commands["clang"], "clang", "C:\\Program Files\\LLVM\\bin\\clang.exe")
		commands["ld.lld"] = append(commands["ld.lld"], "lld", "C:\\Program Files\\LLVM\\bin\\lld.exe")
		commands["wasm-ld"] = append(commands["wasm-ld"], "C:\\Program Files\\LLVM\\bin\\wasm-ld.exe")
		commands["lldb"] = append(commands["lldb"], "C:\\Program Files\\LLVM\\bin\\lldb.exe")
	}
	// Add the path to LLVM installed from ports.
	if goos == "freebsd" {
		prefix := "/usr/local/llvm" + llvmMajor + "/bin/"
		commands["clang"] = append(commands["clang"], prefix+"clang-"+llvmMajor)
		commands["ld.lld"] = append(commands["ld.lld"], prefix+"ld.lld")
		commands["wasm-ld"] = append(commands["wasm-ld"], prefix+"wasm-ld")
		commands["lldb"] = append(commands["lldb"], prefix+"lldb")
	}
	return commands
}

// LookupCommand finds an LLVM tool. Linkers must match the LLVM major version.
// It returns the command to invoke or an error if no suitable tool was found.
func LookupCommand(name string) (string, error) {
	llvmMajor, _, _ := strings.Cut(llvm.Version, ".")
	return lookupCommand(name, commands[name], llvmMajor)
}

func lookupCommand(name string, candidates []string, llvmMajor string) (string, error) {
	isLinker := name == "ld.lld" || name == "wasm-ld"
	var mismatches []error
	for _, cmdName := range candidates {
		path, err := exec.LookPath(cmdName)
		if err != nil {
			// A missing bare command wraps exec.ErrNotFound. A missing
			// absolute path surfaces fs.ErrNotExist instead. Skip either.
			if errors.Unwrap(err) == exec.ErrNotFound || errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return cmdName, err
		}
		if isLinker {
			major, err := linkerMajorVersion(name, path)
			if err != nil {
				return cmdName, err
			}
			if major != llvmMajor {
				mismatches = append(mismatches, fmt.Errorf("%s reports LLD %s, need LLVM %s", path, major, llvmMajor))
				continue
			}
		}
		return cmdName, nil
	}
	if isLinker {
		err := fmt.Errorf("no %s matching LLVM %s found (tried %s)", name, llvmMajor, strings.Join(candidates, " "))
		return "", errors.Join(append([]error{err}, mismatches...)...)
	}
	return "", errors.New("none of these commands were found in your $PATH: " + strings.Join(candidates, " "))
}

var lldVersionPattern = regexp.MustCompile(`(?:^|\s)LLD ([0-9]+)\.`)

func linkerMajorVersion(name, path string) (string, error) {
	args := []string{"--version"}
	if name == "ld.lld" {
		args = []string{"-flavor", "gnu", "--version"}
	}
	output, err := exec.Command(path, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("could not check LLD version of %s: %w: %s", path, err, strings.TrimSpace(string(output)))
	}
	matches := lldVersionPattern.FindSubmatch(output)
	if matches == nil {
		return "", fmt.Errorf("could not parse LLD version of %s: %q", path, strings.TrimSpace(string(output)))
	}
	return string(matches[1]), nil
}
