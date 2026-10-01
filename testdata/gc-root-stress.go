// This is gc-liveness-repro.go without the allocation churn that makes a
// freed dumper visible. Fields read after the collection sit past the free
// list header, so only runtime_clobberfree makes a lost root visible.
package main

import (
	"runtime"
	"runtime/volatile"
)

const magic = 0x5a5a5a5a

type writer interface {
	Write([]byte) (int, error)
}

type dumper struct {
	pad   [4]uintptr
	used  int
	buf   [4]byte
	magic uint32
	w     writer
}

//go:noinline
func newDumper(w writer) writer {
	d := makeDumper(w)
	// Clear stale copies of d left by alloc in frames Write reuses.
	scrubStack()
	return d
}

//go:noinline
func makeDumper(w writer) *dumper {
	return &dumper{w: w, magic: magic}
}

//go:noinline
func (d *dumper) Write(p []byte) (n int, err error) {
	runtime.GC()
	if d.magic != magic {
		println("magic", d.magic, "want", uint32(magic))
		panic("live object was collected")
	}
	for _, b := range p {
		d.buf[0] = b
		_, err = d.w.Write(d.buf[:1])
		if err != nil {
			return
		}
		d.used++
		n++
	}
	return
}

type collector struct {
	used int
}

func (c *collector) Write(p []byte) (int, error) {
	c.used += len(p)
	return len(p), nil
}

func main() {
	var input [40]byte
	c := &collector{}
	d := newDumper(c)
	for i := range input {
		d.Write(input[i : i+1])
	}
	if d.(*dumper).used != len(input) || c.used != len(input) {
		panic("wrong output")
	}
	println("ok")
}

//go:noinline
func scrubStack() {
	var a [256]uint32
	for i := range a {
		volatile.StoreUint32(&a[i], 0)
	}
}
