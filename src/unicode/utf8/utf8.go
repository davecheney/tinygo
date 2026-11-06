// Copyright 2026 The TinyGo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utf8

//go:generate go run ../../../tools/gen-stdlib unicode/utf8

// ptrSize also handles the 16-bit uintptr of AVR, unlike upstream's.
const ptrSize = 1 << (^uintptr(0)>>15&1 + ^uintptr(0)>>31&1 + ^uintptr(0)>>63&1)

func word[T string | []byte](s T) uintptr {
	switch ptrSize {
	case 2:
		return uintptr(s[0]) | uintptr(s[1])<<8
	case 4:
		return uintptr(s[0]) | uintptr(s[1])<<8 | uintptr(s[2])<<16 | uintptr(s[3])<<24
	}
	return uintptr(uint64(s[0]) | uint64(s[1])<<8 | uint64(s[2])<<16 | uint64(s[3])<<24 | uint64(s[4])<<32 | uint64(s[5])<<40 | uint64(s[6])<<48 | uint64(s[7])<<56)
}
