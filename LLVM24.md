# LLVM 24 branch notes

This branch (`llvm24` on github.com/davecheney/tinygo) adds LLVM 24 support
to TinyGo before LLVM 24 is released. It is a long-running branch: keep it
rebased on `origin/dev` until LLVM 24 is released and the changes are
upstream. Delete this file when the branch is merged.

## Branch layout

The branch is `origin/dev` plus these commits, oldest first:

| Commit subject | Upstream PR | Notes |
|---|---|---|
| all: add LLVM 24 support | not sent upstream yet | Build tags, `cgo/libclang_config_llvm24.go`, go.mod replace. |
| builder: ignore -Wunsequenced-volatile in libc | not sent upstream yet | Clang 24 warning in musl and picolibc. Also valid with LLVM 22 and 23. |
| (this file) | never sent upstream | |

These PRs are merged: #6353, #6354, #6355 and #6356. When a PR merges, rebase
on `origin/dev`. Git drops the merged commit. If `origin/dev` changed the same
code after the merge, the rebase stops with a conflict. Then use
`git rebase --skip`. This occurred for #6353, because f996ea30 changed the
code to use `AggregateElement`.

## go-llvm dependency

`tinygo.org/x/go-llvm` does not support LLVM 24 yet. `go.mod` replaces it with
the `llvm24` branch of github.com/davecheney/go-llvm. That branch is
tinygo-org/go-llvm `main` plus one commit, "all: add LLVM 24 support". It
adds `llvm_config_llvm24.go`, the `llvm24` build tag, wrappers for the C
functions that LLVM 24 deprecates, the `PassBuilder` context, and
`Module.NewBuilder`.

tinygo-org/go-llvm#98 ("ir: create basic blocks in the function's context")
is merged. Before #98, `AddBasicBlock` and `InsertBasicBlock` used the global
context, which made invalid IR. LLVM 24 with assertions stops on this IR.

When `origin/dev` needs a newer go-llvm, rebase the fork branch on go-llvm
`main` first. For example, `origin/dev` uses `Value.AggregateElement`
(tinygo-org/go-llvm#103). A fork branch without it does not compile.

To use a new go-llvm commit, push it to the fork, then compute the
pseudo-version and update the replace:

```sh
cd ../go-llvm
ts=$(TZ=UTC git log -1 --format=%cd --date=format-local:%Y%m%d%H%M%S <commit>)
echo v0.0.0-$ts-$(git rev-parse --short=12 <commit>)
cd -
go mod edit -replace tinygo.org/x/go-llvm=github.com/davecheney/go-llvm@<pseudo-version>
GOFLAGS=-mod=mod go list ./compiler/ ./cgo/ ./builder/ >/dev/null   # adds go.sum lines
```

Do not run `go mod tidy` for this. It also removes `golang.org/x/net`, which
is not related to LLVM 24.

When go-llvm releases LLVM 24 support, remove the replace and update the
`require` line.

## The main LLVM 24 change: builders need a module

In LLVM 24, a builder from `LLVMCreateBuilderInContext` (go-llvm
`Context.NewBuilder`) does not know its module until you set an insertion
point. If you use it before that, LLVM stops the process with:

```
Cannot use Builder without known module. Use LLVMCreateBuilderForModule ...
```

The message can start with `LIBCLANG FATAL ERROR`, because libclang installs
the fatal error handler. The error is from LLVM, not clang. It does not come
back to Go as an error, and it stops the complete test binary, so it can hide
other failures.

Upstream commits: llvm/llvm-project ecc201f1d0d5 (IRBuilder takes a Module)
and f740c013f0ad (C API delays the IRBuilder until the insertion point is set).

All 32 calls to `NewBuilder()` in TinyGo were examined in November 2026:

- 21 set the insertion point before any other use. No change.
- 5 read constants with `CreateExtractValue` (`llvmutil`, `noruntime.go`,
  `interrupt.go`). Fixed in #6353 with `Operand(i)`. Later, f996ea30 changed
  `interrupt.go` to `AggregateElement`.
- 5 set the debug location before the insertion point
  (`createFunctionStart`, `createWasmExport`, `getInterfaceInvokeWrapper`).
  Fixed in #6354.
- 1 (`transform/interface-lowering.go`) read 22 constants and made 2 constant
  GEPs. Fixed in #6355 with `constElement` and `llvm.ConstInBoundsGEP`.
  Later, f996ea30 replaced `constElement` with `AggregateElement`.

Rules for new code:

- Set the insertion point before you use a builder for anything, including
  `SetCurrentDebugLocation`.
- Do not use a builder to read a constant. Use `AggregateElement(i)`
  (tinygo-org/go-llvm#103). It works on every kind of constant aggregate.
- `Operand(i)` does not work on a `zeroinitializer` or a `ConstantDataArray`.
  LLVM stores an aggregate with only zero elements as a `zeroinitializer`,
  which has no operands. For example, `Interrupt{0}` and the method set of
  `interface{}`. Use `Operand(i)` only on values that cannot have these forms.

## Build LLVM 24

LLVM 24 is the `main` branch of llvm-project. TinyGo needs clang and lld from
the same build: the `cgo` package links libclang, and TinyGo runs `clang-24`,
`ld.lld` and `wasm-ld`.

```sh
cd ~/devel/llvm-project
cmake -G Ninja -S llvm -B build -DCMAKE_BUILD_TYPE=Release \
  -DLLVM_ENABLE_ASSERTIONS=ON \
  -DLLVM_BUILD_LLVM_DYLIB=ON -DLLVM_LINK_LLVM_DYLIB=ON \
  -DLLVM_ENABLE_PROJECTS="clang;lld" \
  -DLLVM_TARGETS_TO_BUILD="X86;AArch64;ARM;RISCV;WebAssembly;AVR;Mips" \
  -DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD=Xtensa \
  -DLLVM_INCLUDE_TESTS=OFF -DLLVM_INCLUDE_BENCHMARKS=OFF -DLLVM_INCLUDE_EXAMPLES=OFF \
  -DCMAKE_C_COMPILER=clang -DCMAKE_CXX_COMPILER=clang++ -DLLVM_USE_LINKER=lld \
  -DCMAKE_INSTALL_PREFIX=$PWD/build/install
ninja -C build install
```

Keep `LLVM_ENABLE_ASSERTIONS=ON`. The assertions found the go-llvm context
bug. Release packages do not have assertions.

The targets are the same as in `make/llvm.mk`. Without `Mips`, all MIPS tests
in `TestBuild` and `TestTest` fail.

The build makes `libLLVM-24git.so`. go-llvm links `-lLLVM-24`, like the
apt.llvm.org packages. Make a link:

```sh
mkdir -p /tmp/llvm24-link
ln -sf ~/devel/llvm-project/build/install/lib/libLLVM-24git.so /tmp/llvm24-link/libLLVM-24.so
```

## Run the tests

```sh
I=~/devel/llvm-project/build/install
export PATH=$I/bin:$PATH
export CGO_CPPFLAGS="-I$I/include"
export CGO_LDFLAGS="-L/tmp/llvm24-link -L$I/lib -Wl,-rpath,$I/lib"
go test -tags llvm24 \
  -ldflags "-X github.com/tinygo-org/tinygo/goenv.clangResourceDir=$I/lib/clang/24" \
  ./compiler/... ./transform/ ./interp/ ./cgo/ ./compileopts/
```

- The `clangResourceDir` flag is necessary for a local build. On Linux,
  TinyGo looks for the clang headers only in `/usr/lib/clang/24`, where a
  distribution package puts them. Without the flag, the `cgo` tests fail
  with `'stddef.h' file not found`.
- Make sure the test binary uses LLVM 24 only:
  `go test -c -tags llvm24 ./interp/ && ldd interp.test | grep -i llvm`.
  If the system has an older libclang, `-lclang` can find it and put two LLVM
  versions in one process.

Expected result (November 2026, llvm/llvm-project eb17746d68f5): all packages
pass, except `TestCompiler`:

- `slice.go`, `float.go`, `zeromap.go` and `large.go` fail because the golden
  files are written with LLVM 22. The same tests fail in the same way with
  LLVM 23. Golden files must be written with LLVM 22 (`-update` refuses other
  versions).
- `lenrange.go-arduino` fails with `undefined: avr.DEVICE` because
  `src/device/avr` is generated. Run `make gen-device` first. This fails on
  `origin/dev` too.

To make sure that a change does not cause a failure, run the same tests on
`origin/dev` and compare the lists of failures.

## Build and run programs

Use the same environment as for the tests. Build the `tinygo` binary with the
same `-tags` and `-ldflags`. TinyGo compiles the C libraries from source, so
the checkout needs:

- The submodules `lib/musl`, `lib/picolibc`, `lib/wasi-libc` and `lib/bdwgc`.
- `llvm-project` in the checkout root, with the compiler-rt source of the same
  LLVM version. A link is sufficient:
  `ln -s ~/devel/llvm-project llvm-project`.

Clang 24 adds `-Wunsequenced-volatile` (llvm/llvm-project#244073). musl and
picolibc compile with `-Werror` and read volatiles two times in one expression
on purpose. The `builder` commit on this branch disables the warning.

Result (November 2026, llvm/llvm-project eb17746d68f5): hello world builds and
runs on linux/amd64, linux/arm64, cortex-m-qemu, riscv-qemu and wasip1.

## Run the full test suite

`make test` uses the `byollvm llvm22` tags. For LLVM 24, run its `go test`
command with the LLVM 24 environment:

```sh
go test -timeout=1h -buildmode exe -tags "llvm24 osusergo" \
  -ldflags "-X github.com/tinygo-org/tinygo/goenv.clangResourceDir=$I/lib/clang/24" \
  $(go list -e ./... | grep -v -E 'tinygo/(src|tests|lib)/')
```

Initialize all submodules and run `make gen-device` first. Go does not see a
new `libLLVM` build, so use `-count=1` after you rebuild LLVM.

Result (November 2026, llvm/llvm-project eb17746d68f5): `TestBuild` and
`TestTest` pass for all targets, with no assertion failures. These tests fail:

- The 4 `TestCompiler` golden files, as above.
- `TestClangAttributes/esp32s3`, because upstream clang does not know the ESP
  CPUs. The tinygo-org LLVM fork adds them. With the Arch `llvm22` package,
  this test fails in the same way.

## Open work

- Add a CI job for LLVM 24 (apt.llvm.org publishes development packages in
  the unversioned `llvm-toolchain-<distro>` repository).
- Send "all: add LLVM 24 support" upstream after go-llvm supports LLVM 24.
- go-llvm: maybe deprecate the package-level `AddBasicBlock` and
  `InsertBasicBlock` in favour of the `Context` methods.
- go-llvm: maybe change the `IsA*` methods to return `bool`. TinyGo uses the
  result only with `.IsNil()`. This is a breaking change.

## Conventions

- Write commit messages in ASD-STE100 Simplified Technical English: short
  sentences, active voice. Give the cause, not a description of the diff.
- Do not add `Co-Authored-By` or "Generated with" lines to commits or pull
  requests. CI rejects commits that a coding tool authored.
- Code comments explain why, not what.
