package builder

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/compileopts"
	"tinygo.org/x/go-llvm"
)

func TestFIPSDisabled(t *testing.T) {
	for _, opt := range []string{"s", "1"} {
		for _, used := range []bool{false, true} {
			name := opt + "/unused"
			if used {
				name = opt + "/used"
			}
			t.Run(name, func(t *testing.T) {
				options := &compileopts.Options{
					Target:        "wasip1",
					Opt:           opt,
					Semaphore:     sema,
					InterpTimeout: time.Minute,
					VerifyIR:      true,
				}
				if used {
					options.Tags = []string{"sha3used"}
				}
				config, err := NewConfig(options)
				if err != nil {
					t.Fatal(err)
				}
				tmpdir := t.TempDir()
				outpath := filepath.Join(tmpdir, "main.ll")
				if _, err := Build("./testdata/fips", outpath, tmpdir, config); err != nil {
					t.Fatal(err)
				}
				ctx := llvm.NewContext()
				defer ctx.Dispose()
				buf, err := llvm.NewMemoryBufferFromFile(outpath)
				if err != nil {
					t.Fatal(err)
				}
				mod, err := ctx.ParseIR(buf)
				if err != nil {
					t.Fatal(err)
				}
				defer mod.Dispose()
				fn := mod.NamedFunction("crypto/internal/fips140/sha3.keccakF1600Generic")
				retained := !fn.IsNil() && !fn.IsDeclaration()
				if retained != used {
					t.Errorf("Keccak retained = %v, want %v", retained, used)
				}
			})
		}
	}
}
