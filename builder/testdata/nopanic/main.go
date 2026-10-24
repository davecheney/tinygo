package main

var words = []string{"alpha", "beta", "gamma"}

func main() {
	m := map[string]int{}
	for _, w := range words {
		m[w] = len(w)
	}
	delete(m, "alpha")
	n := 0
	for k, v := range m {
		if k < "c" || k == "gamma" {
			n += v
		}
	}
	for _, r := range "héllo" {
		n += int(r)
	}
	println("n:", n, m["beta"])
}
