package compileopts

import (
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
)

// CFlags maps a C compiler option to its value, so that a target can override
// an option it inherits. For "-march=rv32imc" the option is "-march" and the
// value is "rv32imc". For "-Werror" or "-I/path" the value is empty.
type CFlags map[string]string

// Set adds a flag, replacing any earlier value for the same option.
func (f *CFlags) Set(flag string) {
	if *f == nil {
		*f = CFlags{}
	}
	option, value, _ := strings.Cut(flag, "=")
	(*f)[option] = value
}

// Args returns the flags as compiler arguments, sorted by option.
func (f CFlags) Args() []string {
	var args []string
	for _, option := range slices.Sorted(maps.Keys(f)) {
		if value := f[option]; value != "" {
			args = append(args, option+"="+value)
		} else {
			args = append(args, option)
		}
	}
	return args
}

// UnmarshalJSON reads the flags from a JSON array of strings.
func (f *CFlags) UnmarshalJSON(data []byte) error {
	var flags []string
	if err := json.Unmarshal(data, &flags); err != nil {
		return err
	}
	*f = make(CFlags, len(flags))
	for _, flag := range flags {
		if !strings.HasPrefix(flag, "-") {
			return fmt.Errorf("cflag %q is not an option, use the -option=value form", flag)
		}
		option, _, _ := strings.Cut(flag, "=")
		if _, ok := (*f)[option]; ok {
			return fmt.Errorf("duplicate cflag %q", option)
		}
		f.Set(flag)
	}
	return nil
}

// MarshalJSON writes the flags as a JSON array of strings.
func (f CFlags) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Args())
}

// FilePrefixMap returns a Clang file-prefix map for two directory paths.
func FilePrefixMap(from, to string) string {
	return directoryPrefixMap("-ffile-prefix-map=", from, to)
}

// DebugPrefixMap returns a Clang debug-prefix map for two directory paths.
func DebugPrefixMap(from, to string) string {
	return directoryPrefixMap("-fdebug-prefix-map=", from, to)
}

func directoryPrefixMap(flag, from, to string) string {
	if !strings.HasSuffix(from, string(filepath.Separator)) {
		from += string(filepath.Separator)
	}
	if !strings.HasSuffix(to, "/") {
		to += "/"
	}
	return flag + from + "=" + to
}

// RewriteCFlags rewrites path operands without changing other compiler options.
func RewriteCFlags(flags []string, rewrite func(string) string) []string {
	flags = append([]string(nil), flags...)
	nextIsPath := false
	for i, flag := range flags {
		if nextIsPath {
			flags[i] = rewrite(flag)
			nextIsPath = false
			continue
		}
		switch flag {
		case "-I", "-L", "-F", "-isystem", "-isystem-after", "-iquote", "-idirafter", "-include", "--include", "-imacros", "--imacros", "-include-pch", "-isysroot", "--sysroot", "-resource-dir", "-iframework", "-iframeworkwithsysroot":
			nextIsPath = true
			continue
		}
		for _, prefix := range []string{
			"-I", "-L", "-F",
			"-isystem-after", "-isystem", "-iquote", "-idirafter",
			"-iframeworkwithsysroot", "-iframework",
			"-include", "--include", "-imacros", "--imacros",
			"--sysroot=", "-isysroot=", "-isysroot", "-resource-dir=",
			"-fmodule-map-file=", "-fmodules-cache-path=",
		} {
			if after, ok := strings.CutPrefix(flag, prefix); ok {
				flags[i] = prefix + rewrite(after)
				break
			}
		}
		for _, prefix := range []string{"-fdebug-prefix-map=", "-ffile-prefix-map=", "-fmacro-prefix-map="} {
			if mapping, ok := strings.CutPrefix(flag, prefix); ok {
				if sep := strings.LastIndexByte(mapping, '='); sep >= 0 {
					flags[i] = prefix + rewrite(mapping[:sep]) + mapping[sep:]
				}
				break
			}
		}
	}
	return flags
}
