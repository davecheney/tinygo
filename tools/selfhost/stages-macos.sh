#!/bin/bash
# Build tinygo with tinygo on macOS. Stage 1 is built by go, stage 2 by stage 1
# and stage 3 by stage 2. See tools/selfhost/README.md.
set -euo pipefail

LLVM=${LLVM:-23}
OPT=${OPT:-s}

cd "$(git rev-parse --show-toplevel)"
export TINYGOROOT=$PWD
tags="llvm$LLVM osusergo"
llvm=$(brew --prefix "llvm@$LLVM")
sdk=$(xcrun --show-sdk-path)

# Go does not see changes to the Homebrew LLVM headers, so a cached go-llvm
# can report an old LLVM version after an upgrade. -a rebuilds it.
mkdir -p build/stage1 build/stage2 build/stage3
go build -a -tags "$tags" -o build/stage1/tinygo .

# Copies of two modules, used through a separate go.mod:
# - go.bug.st/serial uses IOKit through cgo on darwin, and the minimal SDK of
#   tinygo has no IOKit headers. The copy uses the cgo-free stub of freebsd.
# - go-llvm passes -Wl, flags to the linker, but tinygo calls ld.lld directly.
work=$PWD/build/selfhost-macos
rm -rf "$work"
mkdir -p "$work"
serial=$(go list -m -f '{{.Dir}}' go.bug.st/serial)
gollvm=$(go list -m -f '{{.Dir}}' tinygo.org/x/go-llvm)
cp -R "$serial" "$work/serial"
cp -R "$gollvm" "$work/go-llvm"
chmod -R u+w "$work"
cp "$work/serial/enumerator/usb_freebsd.go" "$work/serial/enumerator/usb_darwin.go"
sed -i '' -e 's/ -Wl,-search_paths_first -Wl,-headerpad_max_install_names//' "$work"/go-llvm/llvm_config_llvm*.go
cp go.mod go.sum "$work/"
printf 'replace go.bug.st/serial => %s/serial\nreplace tinygo.org/x/go-llvm => %s/go-llvm\n' "$work" "$work" >> "$work/go.mod"
export GOFLAGS=-modfile=$work/go.mod

# The minimal SDK of tinygo has no C++ headers and no libz, libm or libc++.
# -I puts libc++ before the minimal SDK, -idirafter fills in the rest.
export CGO_CXXFLAGS="-I$llvm/include/c++/v1 -idirafter $sdk/usr/include"
export CGO_LDFLAGS="-L$sdk/usr/lib -L$llvm/lib/c++ -lc++ -lc++abi"
for stage in 2 3; do
	prev=$((stage - 1))
	time "build/stage$prev/tinygo" build -opt="$OPT" -tags "$tags" -o "build/stage$stage/tinygo" .
done
unset GOFLAGS CGO_CXXFLAGS CGO_LDFLAGS

for stage in 1 2 3; do
	"build/stage$stage/tinygo" version
done
if cmp -s build/stage2/tinygo build/stage3/tinygo; then
	echo "stage 2 and stage 3 are identical"
else
	echo "stage 2 and stage 3 differ"
fi

hello=$(mktemp -d)
trap 'rm -rf "$hello"' EXIT
printf 'package main\n\nfunc main() { println("hello from stage 3") }\n' > "$hello/main.go"
(cd "$hello" && "$TINYGOROOT/build/stage3/tinygo" run main.go)
