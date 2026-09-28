// Loop with many iterations and arithmetic the compiler cannot evaluate at compile time
package main

import "fmt"

func main() {
	var i, acc int64 = 0, 0
	for i < 5000000 {
		acc = (acc*31 + i) % 1000003
		i = i + 1
	}
	fmt.Println(acc)
}
