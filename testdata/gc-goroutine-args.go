package main

import "sync"

// On AVR the collector freed the arguments of a goroutine that had not started
// yet, because a saved register held the only pointer. See #6358.

var sink chan int

func main() {
	// Fill the heap so that the goroutine starts below trigger a GC.
	for range 448 {
		sink = make(chan int, 1)
	}

	ch := make(chan int)
	var wg sync.WaitGroup
	sums := make([]int, 3)
	for i := range sums {
		wg.Add(1)
		go func() {
			sums[i] = <-ch + <-ch
			wg.Done()
		}()
	}
	for n := range 6 {
		ch <- n
	}
	wg.Wait()
	println("sum:", sums[0]+sums[1]+sums[2])
}
