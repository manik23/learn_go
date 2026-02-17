package main

import (
	"fmt"
	"runtime"
)

type Node struct {
	next *Node
	data [900]byte
}

var x Node

func stackAlloc() {
	var n Node
	x = n
	fmt.Println(x)
}

func heapAlloc() *Node {
	n := Node{}
	fmt.Println(n)
	return &n
}

// PrintMemUsage outputs the current memory stats to stdout.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Memory Usage:\n")
	fmt.Printf("\tHeapAlloc = %v MiB", m.HeapAlloc/1024/1024)
	fmt.Printf("\tHeapSys = %v MiB", m.HeapSys/1024/1024)
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func main() {
	// // --- Exercise 1 — Stack vs Heap ---- //
	// stackAlloc()
	// heapAlloc()
	// // --- Exercise 1 — Stack vs Heap ---- //

	// // --- Exercise 2 — Allocation Pattern ---- //
	// nodes := make([]*Node, 0)

	// var prev uintptr
	// for i := 0; i < 10; i++ {
	// 	n := &Node{}
	// 	addr := uintptr(unsafe.Pointer(n))
	// 	nodes = append(nodes, n)
	// 	fmt.Printf("%p\n", n)
	// 	if i > 0 {
	// 		fmt.Println("diff:", addr-prev)
	// 	}
	// 	prev = addr
	// }
	// // // --- Exercise 2 — Allocation Pattern ---- //

	// --- Exercise 3 — GC Trigger Behavior ---- //
	for range 100 {
		PrintMemUsage()
		_ = &Node{}
	}
	// --- Exercise 3 — GC Trigger Behavior ---- //

}
