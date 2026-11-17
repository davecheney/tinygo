# go1.28 tracking branch

Long-lived branch on `davecheney/tinygo` that tracks the Go toolchain at
tip (which will become go1.28) so TinyGo breakage is found before the release.
This file is a handoff note for whoever (human or agent) picks it up next.
Delete it before any of this is upstreamed.

## Setup

- Go at tip: `~/devel/go` (`cd src && ./make.bash`). First checked at
  `go1.28-devel_3c6d092895` (2026-11-15).
- Branch base: tinygo `origin/dev` @ `7db05af8` (2026-11-17, was `c35ae925`). Rebase onto `origin/dev`
  periodically; keep go1.28-only fixes as small, separate commits.
- Build: `PATH=~/devel/go/bin:$PATH go build -tags llvm23 -o build/tinygo .`
  (system LLVM was 23.1.5). Submodules are needed for running tinygo
  (`git submodule update --init`). Don't symlink `src/net`: the GOROOT
  overlay treats a symlinked dir as a file and fails with "file exists".

## Status (2026-11-17)

Worktree `.claude/worktrees/go128-dev`, branch `worktree-go128-dev` on `davecheney`, based on `origin/dev` @ `7db05af8`.
The old `~/devel/tinygo-go1.28` checkout was removed. Commit 2 was lost with it and has been redone here.
`~/devel/go` is still at `3c6d092895`. Go master was at `98ead39515` on 2026-11-17, so update and rerun.

| Check | Result |
|---|---|
| `go build -tags llvm23` with go1.28-devel_3c6d092895 | OK |
| `tinygo run hello.go` (host) | OK |
| `tinygo run -target=riscv-qemu` | OK |
| `tinygo build -target=wasip1` | OK |
| `tinygo run -target=cortex-m-qemu` | OK with commit 2, failed with `-marm` before it (#6357) |
| Same smoke tests with go1.27.4 and commit 2 | OK |
| `go test ./loader`, `TestGetListOfPackages` | OK |

## Status (2026-11-16)

| Check | Result |
|---|---|
| `go build` of tinygo with go1.28 | OK, no errors |
| `go test ./builder ./transform ./interp ./compileopts ./cgo ./loader ./goenv` | OK |
| `tinygo run hello.go` (host linux/amd64) | OK once the version cap is raised |
| `tinygo build -target=cortex-m-qemu` | Failed (see below); OK with commit 2 |
| `go test ./compiler` | 3 extra failures vs go1.27.4, all cortex-m-qemu (same bug); back to baseline with commit 2 |
| `tinygo run -target=cortex-m-qemu` / `riscv-qemu` (QEMU) | OK with commit 2: both print the expected output |

Not yet run: the full `go test .` (main_test.go testdata matrix),
`make tinygo-test` (stdlib package tests), wasm/wasip1/wasip2, other
baremetal targets, or macOS/Windows.

Baseline noise: with go1.27.4 + LLVM 23, `go test ./compiler` already fails
`float.go`, `large.go`, `slice.go`, `zeromap.go` (IR attribute renumbering)
and `lenrange.go-arduino` (needs `make gen-device`). These are not go1.28
issues.

## Commits on this branch

1. `compileopts: allow go1.28` sets `MaximumGoMinorVersion = 28`.
   Without it: `requires go version 1.26 through 1.27, got go1.28`.
2. `loader: list only the PackageJSON fields` adds `loader.PackageJSONFlag()`
   (`-json=<fields of PackageJSON>`, built by reflection) and uses it in
   `loader.Load` and `getListOfPackages` in main.go. This is fix option 1
   below. It doesn't depend on go1.28, so it's a good candidate to upstream
   to tinygo-org on its own.
   Verified with go1.28 and go1.27.4: `tinygo build -target=cortex-m-qemu`
   OK, and `go test ./compiler ./loader` shows only the baseline failures.

CI workflows (`.github/workflows/*.yml`) still pin `1.27.1`. setup-go can't
install an unreleased Go, so they're unchanged for now. Switch them once a
go1.28rc exists.

## Fixed here: `go list -json` now runs cgo (broke every baremetal ARM target)

Upstream TinyGo issue: tinygo-org/tinygo#6357 (milestone 0.46, assigned to
davecheney).

The `-json=<fields>` fix is a short-term fix to unblock this branch. The
real fix is more subtle: TinyGo uses the linux/arm placeholder for stdlib
file selection, and that needs more work to remove.

Rejected test: do not set GOOS/GOARCH for `go list -deps` (2026-11-16).
- Without the `-json` fix: `go list` still runs cgo, now for the host
  GOARCH (amd64). It then assembles `device/arm/cortexm.S` with the x86
  assembler, and that fails (`unknown pseudo-op: .syntax`).
- With the `-json` fix: cortex-m-qemu and riscv-qemu fail in
  `internal/cpu` (`CacheLinePadSize redeclared`). The TinyGo tags include
  `arm`, and the host GOARCH is amd64, so Go selects both `cpu_arm.go` and
  `cpu_x86.go`. GOARCH controls stdlib file selection, so TinyGo must set it.

Symptom:

```
$ tinygo build -target=cortex-m-qemu hello.go
# device/arm
gcc: error: unrecognized command-line option '-marm'; did you mean '-mabm'?
```

Mechanism:
- `loader/list.go` runs `go list -json -deps -e` with `GOARCH=arm CGO_ENABLED=1`
  and TinyGo's merged GOROOT. CC is left as the host `gcc`.
- Plain `-json` asks for all fields, including `Stale`, so cmd/go builds an
  action graph (`needStale` in `cmd/go/internal/list/list.go`, with
  `b.IsCmdList = true`).
- go1.27: under `IsCmdList` the cgo package only got cache keys computed;
  cgo never ran.
  go1.28: `go list -x -json` shows it **executing `pkg/tool/.../cgo`** on
  `device/arm` (CgoFiles `arm.go`, plus `interrupts.c`, `cortexm.S`). cgo
  calls host gcc with `-marm`, which fails, so `device/arm` comes back
  `Incomplete` with `Error` set and tinygo's loader aborts.
- Upstream cause, **confirmed**: `89b06d0c43` "cmd/go/internal/work: cache
  run cgo and cgo compile actions separately" (Matloob, 2026-07-11, CL 799960,
  #15681). Built at `89b06d0c43^`, the repro doesn't run cgo. Built at
  `89b06d0c43`, it does.

Why no host toolchain can satisfy this: for targets Go doesn't support
(AVR, RISC-V, Xtensa, wasm-unknown) TinyGo sets `GOOS=linux GOARCH=arm` only
so that `go list` will resolve packages. It is a placeholder, not a
description of the target (AVR has 16-bit pointers, for example).
Target-specific files are picked by TinyGo build tags, type sizes come from
LLVM (`compiler.Sizes(machine)`), and codegen is TinyGo's own. TinyGo uses
`go list` purely to resolve files and imports. Once `go list` runs cgo, the
host CC gets flags for that fictional GOARCH (`-marm` from
`cmd/cgo/gcc.go` `gccMachine`), which can't work on an x86_64 or arm64 host.
Dropping the GOARCH override isn't a fix: file selection would be wrong, and
`go list` would still run cgo.

Packages in TinyGo's own sources that use `import "C"`: `device/arm`,
`machine`, `internal/task`. So all `GOARCH=arm` targets are likely affected,
not just cortex-m. A per-target matrix (cortex-m, riscv-qemu, hifive1b,
esp32, avr, wasm-unknown, wasip1, linux/arm64) has not been run yet.

Minimal repro without tinygo (`$R` = a tinygo cached GOROOT,
`~/.cache/tinygo/goroot-*`):

```
GOROOT=$R GOOS=linux GOARCH=arm CGO_ENABLED=1 \
  go list -x -json -e -tags "cortexm baremetal tinygo" device/arm
```

go1.27.4 has no cgo exec and no `Error`. go1.28 runs cgo and sets
`Error`/`Incomplete`.

Fix options (1 is done on this branch, 2 is still open):
1. TinyGo side (works on any Go): pass `-json=<fields>` listing only the
   fields the loader decodes (no `Stale`/`StaleReason`). Then `needStale` is
   false and no build actions run. Probably the right fix regardless, and
   faster too.
2. Upstream: file a cmd/go issue saying `go list -json` without
   `-compiled`/`-export` shouldn't execute cgo. That's a regression for every
   tool that cross-lists cgo packages without a working cross CC.
