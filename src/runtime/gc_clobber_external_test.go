//go:build (gc.conservative || gc.precise) && runtime_clobberfree

package runtime_test

import (
	"runtime"
	"testing"
)

func TestGCSweepPoison(t *testing.T) {
	if failure := runtime.GCClobberSweepProbe(); failure != "" {
		t.Fatal(failure)
	}
}
