package main

import "runtime"

var (
	intSink int
	okSink  bool
)

func main() {
	stringMap := map[string]int{"key": 1}
	noAllocs("string map", func() {
		for i := range 100 {
			stringMap["key"] = i
			intSink, okSink = stringMap["key"]
		}
		delete(stringMap, "missing")
	})

	float32Map := map[float32]int{1.25: 2}
	noAllocs("float32 map", func() {
		for range 100 {
			intSink, okSink = float32Map[1.25]
		}
	})

	float64Map := map[float64]int{2.5: 3}
	noAllocs("float64 map", func() {
		for range 100 {
			intSink, okSink = float64Map[2.5]
		}
	})

	interfaceMap := map[interface{}]int{
		int32(4):        4,
		"key":           5,
		float32(1.5):    6,
		int64(7):        7,
		[2]uint16{8, 9}: 8,
	}
	noAllocs("interface map", func() {
		for range 100 {
			intSink, okSink = interfaceMap[int32(4)]
			intSink, okSink = interfaceMap["key"]
			intSink, okSink = interfaceMap[float32(1.5)]
			intSink, okSink = interfaceMap[int64(7)]
			intSink, okSink = interfaceMap[[2]uint16{8, 9}]
		}
	})
}

func noAllocs(name string, f func()) {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	f()
	runtime.ReadMemStats(&after)
	if after.Mallocs != before.Mallocs {
		panic(name + " allocated")
	}
}
