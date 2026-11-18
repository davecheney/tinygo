#!/bin/bash
# Build tinygo with tinygo. Stage 1 is built by go, stage 2 by stage 1 and
# stage 3 by stage 2. See tools/selfhost/README.md.
set -euo pipefail

LLVM=${LLVM:-23}
OPT=${OPT:-s}

cd "$(git rev-parse --show-toplevel)"
export TINYGOROOT=$PWD
tags="llvm$LLVM osusergo"

# Go does not see changes to system LLVM headers, so a cached go-llvm can
# report an old LLVM version after an upgrade. -a rebuilds it.
mkdir -p build/stage1 build/stage2 build/stage3
go build -a -tags "$tags" -o build/stage1/tinygo .

# libLLVM.so links to glibc. tinygo does not add the C++ runtime for cgo
# packages, so CGO_LDFLAGS adds it.
export CGO_LDFLAGS="-lstdc++ -l:libgcc_s.so.1"
for stage in 2 3; do
	prev=$((stage - 1))
	time "build/stage$prev/tinygo" build -opt="$OPT" -libc=glibc -tags "$tags" -o "build/stage$stage/tinygo" .
done

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
