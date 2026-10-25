Please take a look at our [Contributing](https://tinygo.org/docs/guides/contributing/) page on our web site for details. Thank you.

When building from source, run `make llvm-source` after updating TinyGo to get
the required compiler-rt sources. This is also needed when using an installed LLVM
with `go install -tags=llvm22`. A build error reporting a missing `go:embed` path
under `llvm-project/compiler-rt` means these sources need to be downloaded or updated.
