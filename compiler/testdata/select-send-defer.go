package main

func selectSendDefer(ch, other chan int) {
	defer func() {
		recover()
	}()
	select {
	case ch <- 1:
	case <-other:
	}
}
