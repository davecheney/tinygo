package main

import "math"

// PLL table functions from ce5db77d, src/machine/machine_rp2_pll.go.
func abs(a int64) int64 {
	if a == math.MinInt64 {
		return math.MaxInt64
	} else if a < 0 {
		return -a
	}
	return a
}

var pdTable = [50]struct {
	hivco [2]uint8
	lovco [2]uint8
}{}

//go:compiletime
func genTable() {
	if pdTable[1].hivco[1] != 0 {
		return
	}
	for product := 1; product < len(pdTable); product++ {
		genTableEntry(product)
	}
}

//go:compiletime
func genTableEntry(product int) {
	bestProdhi := 255
	bestProdlo := 255
	for pd1 := 7; pd1 > 0; pd1-- {
		for pd2 := pd1; pd2 > 0; pd2-- {
			gotprod := pd1 * pd2
			if abs(int64(gotprod-product)) < abs(int64(bestProdlo-product)) {
				bestProdlo = gotprod
				pdTable[product].lovco[0] = uint8(pd1)
				pdTable[product].lovco[1] = uint8(pd2)
			}
		}
	}
	for pd1 := 1; pd1 < 8; pd1++ {
		for pd2 := 1; pd2 <= pd1; pd2++ {
			gotprod := pd1 * pd2
			if abs(int64(gotprod-product)) < abs(int64(bestProdhi-product)) {
				bestProdhi = gotprod
				pdTable[product].hivco[0] = uint8(pd1)
				pdTable[product].hivco[1] = uint8(pd2)
			}
		}
	}
}

var checksum uint32

func init() {
	genTable()
	for _, entry := range pdTable {
		checksum += uint32(entry.hivco[0]) + uint32(entry.hivco[1])
		checksum += uint32(entry.lovco[0]) + uint32(entry.lovco[1])
	}
}

func main() {
	println(checksum)
}
