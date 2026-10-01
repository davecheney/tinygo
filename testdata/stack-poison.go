// Stack-promoted objects checked under -internal-poison-stackallocs. The
// escape analysis must heap allocate returned; the rest stay on the stack.
package main

type object struct{ x [4]int }

func (p *object) Error() string { return "object" }

//go:noinline
func wrap(p *object) (int, error) {
	return 1, p
}

//go:noinline
func returned() error {
	p := &object{x: [4]int{1, 42, 3, 4}}
	_, result := wrap(p)
	return result
}

//go:noinline
func local() int {
	p := &object{x: [4]int{1, 43, 3, 4}}
	_, result := wrap(p)
	return result.(*object).x[1]
}

//go:noinline
func sum(p *object) int {
	return p.x[0] + p.x[1] + p.x[2] + p.x[3]
}

//go:noinline
func loop(n int) int {
	total := 0
	for i := 0; i < n; i++ {
		p := &object{x: [4]int{i, i, i, i}}
		total += sum(p)
	}
	return total
}

var deferCalls int

//go:noinline
func note() {
	deferCalls++
}

//go:noinline
func deferred() int {
	p := &object{x: [4]int{5, 6, 7, 8}}
	defer note()
	return sum(p)
}

//go:noinline
func recursive(depth int) int {
	p := &object{x: [4]int{depth, depth, depth, depth}}
	if depth > 0 {
		recursive(depth - 1)
	}
	return sum(p)
}

//go:noinline
func panics() {
	p := &object{x: [4]int{9, 9, 9, 9}}
	if sum(p) == 36 {
		panic("expected")
	}
}

//go:noinline
func recovers() (ok bool) {
	defer func() {
		ok = recover() != nil
	}()
	panics()
	return false
}

//go:noinline
func outer() bool {
	p := &object{x: [4]int{7, 7, 7, 7}}
	return recovers() && sum(p) == 28
}

func main() {
	p := returned().(*object)
	if p.x[1] != 42 {
		println("poisoned", p.x[1])
		panic("aggregate return corrupted")
	}
	if local() != 43 {
		panic("nonescaping control corrupted")
	}
	if n := loop(5); n != 40 {
		println("loop", n)
		panic("loop control corrupted")
	}
	if n := deferred(); n != 26 || deferCalls != 1 {
		println("defer", n, deferCalls)
		panic("defer control corrupted")
	}
	if n := recursive(4); n != 16 {
		println("recursive", n)
		panic("recursion control corrupted")
	}
	if !outer() {
		panic("recover control corrupted")
	}
	println("ok")
}
