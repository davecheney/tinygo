package ringbuffer_test

import (
	"internal/ringbuffer"
	"runtime"
	"sync"
	"testing"
)

// TestConcurrent runs Put and Get on separate threads while the buffer is
// mostly full. The test is amd64 only because volatile accesses do not order
// memory between cores on weaker memory models such as arm64.
func TestConcurrent(t *testing.T) {
	const total = 1 << 20
	var rb ringbuffer.Buffer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < total; {
			if rb.Put(byte(i)) {
				i++
			} else {
				runtime.Gosched()
			}
		}
	}()
	bad := 0
	for i := 0; i < total; {
		b, ok := rb.Get()
		if !ok {
			runtime.Gosched()
			continue
		}
		if b != byte(i) {
			bad++
		}
		i++
	}
	wg.Wait()
	if bad != 0 {
		t.Errorf("Get returned the wrong byte %d times in %d", bad, total)
	}
}
