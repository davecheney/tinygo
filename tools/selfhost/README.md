# Self-hosted tinygo

The `self-host` branch on github.com/davecheney/tinygo holds the changes that
let tinygo build tinygo. It is a long running branch. Rebase it onto
`origin/dev` and drop each commit once its change lands upstream.

## Build the stages

Run on linux amd64 with the system LLVM:

    tools/selfhost/stages-linux.sh

Run on macOS arm64 with Homebrew LLVM:

    tools/selfhost/stages-macos.sh

| Stage | Built by | Output |
|-------|----------|--------|
| 1 | `go build` | `build/stage1/tinygo` |
| 2 | stage 1 | `build/stage2/tinygo` |
| 3 | stage 2 | `build/stage3/tinygo` |

`LLVM` selects the LLVM version (default 23) and `OPT` the optimization level
for stages 2 and 3 (default `s`). Each script prints the version of each stage,
checks whether stage 2 and stage 3 are identical, and runs a hello program
with stage 3.

The macOS script builds stages 2 and 3 with copies of `go.bug.st/serial` and
`tinygo.org/x/go-llvm` in `build/selfhost-macos`. The serial copy does not use
IOKit, and the go-llvm copy does not pass `-Wl,` flags to ld.lld.

## Results

On odessa (Arch Linux, x86_64, system LLVM 23.1.5, go 1.27.4) at `cf0b94cb`
with `-opt=s`:

| Stage | Build time | Size |
|-------|------------|------|
| 1 | not timed | 22,969,488 bytes |
| 2 | 38 s | 43,037,424 bytes |
| 3 | 1 min 24 s | 43,037,432 bytes |

All three stages report `0.45.0-dev-cf0b94cb`, and stages 2 and 3 end with
`built with tinygo 0.45.0-dev-cf0b94cb`. Stage 3 builds and runs hello world
for linux and for wasip1.

## Build setup

- Stages 2 and 3 use `-libc=glibc`. The system `libLLVM.so` and `libclang.so`
  link to glibc, so tinygo cannot link them into a musl binary. With glibc,
  tinygo also uses the system headers in `/usr/include`, where Arch puts the
  LLVM headers.
- tinygo does not link the C++ runtime for cgo packages that contain C++, as
  `go build` does. `CGO_LDFLAGS="-lstdc++ -l:libgcc_s.so.1"` adds it. This
  needs `CGO_LDFLAGS` support (#6279). Arch has no `libgcc_s.so` link, only
  `libgcc_s.so.1`.
- Stage 1 uses `go build -a`. Go does not see changes to the system LLVM
  headers, so after an LLVM upgrade a cached go-llvm reports the old version
  and tinygo panics at startup with `incorrect build`.
- The Arch `llvmNN` packages for older LLVM need the paths from the
  `arch-llvm-paths` branch.

## Stage 2 and stage 3 differ

The two binaries are 8 bytes apart in `.data` and also differ in
`.debug_str` and `.debug_pubnames`. Two causes are known:

- Each build puts the cgo snippet files in a new temporary directory, such as
  `/tmp/tinygo1513398767`, and that path ends up in the binary.
- ThinLTO gives local symbols a `.llvm.NNN` suffix, such as
  `(*internal/task.Mutex).Unlock.llvm.6642387286999133842`, and the number
  differs between the two builds.

Stage 2 and stage 3 will only be identical once both are fixed. If
`go build` runs without VCS stamping, for example with
`GOFLAGS=-buildvcs=false`, stage 1 reports `0.45.0-dev` without the commit,
and stage 2 then differs from stage 3 in its version string too.

## Commits on the branch

| Commit | Upstream |
|--------|----------|
| builder: skip the toolchain version check in a tinygo-built tinygo | no PR |
| main: report the tinygo that built tinygo in tinygo version | no PR |
| compileopts: add a GNU build ID to linux binaries | no PR |
| cgo: pass large libclang structs through pointers | no PR, works around #4489 |
| hash/maphash: add a TinyGo version of the package | #6336 |
| runtime: make comparablehash use the seed for a value with no bytes | no PR |
| compileopts: use an 8 MB goroutine stack on linux | no PR |
| loader: support CGO_LDFLAGS | #6279 |
| tools/selfhost: add the stage build script and notes | branch only |

Why each change is needed:

- The goroutine stack is 64 kB by default. interp recurses deeper than that
  and overflows.
- TinyGo does not lower structs larger than 16 bytes to the C ABI (#4489), so
  libclang calls that pass such structs by value crash.
- Go 1.27 removed `maphash_purego.go`, so `hash/maphash` does not build with
  TinyGo without an override.
