//go:build gc.conservative || gc.precise

package runtime_test

import (
	"runtime"
	"testing"
)

func TestGCStackScanEnd(t *testing.T) {
	if failure := runtime.GCStackScanProbe(); failure != "" {
		t.Fatal(failure)
	}
}

func TestGCStackDeadFrame(t *testing.T) {
	if failure := runtime.GCStackDeadFrameProbe(); failure != "" {
		t.Fatal(failure)
	}
}
