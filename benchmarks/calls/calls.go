// Many calls to a short function: measures the cost of each call
package main

import "fmt"

//go:noinline
func step(acc, i int64) int64 {
	return (acc + i*7) % 1000003
}

func main() {
	var acc, i int64 = 0, 0
	for i < 1000000 {
		acc = step(acc, i)
		i = i + 1
	}
	fmt.Println(acc)
}
