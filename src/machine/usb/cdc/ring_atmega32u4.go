//go:build atmega32u4

package cdc

import "runtime/volatile"

// The 128-byte capacity keeps uint8 head-tail subtraction unambiguous across wraparound.
type ring512 struct {
	buf  [ringBufLen]byte
	head uint8
	tail uint8
}

const ringMask = ringBufLen - 1

func (r *ring512) Reset() {
	volatile.StoreUint8(&r.head, 0)
	volatile.StoreUint8(&r.tail, 0)
}

func (r *ring512) Free() uint32 {
	return ringBufLen - r.Used()
}

func (r *ring512) Used() uint32 {
	head := volatile.LoadUint8(&r.head)
	tail := volatile.LoadUint8(&r.tail)
	return uint32(head - tail)
}

func (r *ring512) Peek() (data1, data2 []byte) {
	head := volatile.LoadUint8(&r.head)
	tail := volatile.LoadUint8(&r.tail)
	used := uint32(head - tail)
	if used == 0 {
		return nil, nil
	}
	pos := uint32(tail & ringMask)
	contig := ringBufLen - pos
	if contig >= used {
		return r.buf[pos : pos+used], nil
	}
	return r.buf[pos:], r.buf[:used-contig]
}

func (r *ring512) Discard(numBytes uint32) {
	if numBytes == 0 {
		return
	}
	head := volatile.LoadUint8(&r.head)
	tail := volatile.LoadUint8(&r.tail)
	used := uint32(head - tail)
	if numBytes > used {
		panic("ring: discard exceeds used")
	}
	volatile.StoreUint8(&r.tail, tail+uint8(numBytes))
}

func (r *ring512) Put(data []byte) bool {
	wlen := uint32(len(data))
	if wlen == 0 {
		return true
	}
	head := volatile.LoadUint8(&r.head)
	tail := volatile.LoadUint8(&r.tail)
	used := uint32(head - tail)
	free := ringBufLen - used
	if wlen > free {
		return false
	}
	pos := uint32(head & ringMask)
	n := copy(r.buf[pos:], data)
	if n < len(data) {
		copy(r.buf[:], data[n:])
	}
	volatile.StoreUint8(&r.head, head+uint8(wlen))
	return true
}
