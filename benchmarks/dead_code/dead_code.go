// Loop with branches that become dead after folding their literal conditions
package main

import "fmt"

func main() {
	var i, acc int64 = 0, 0
	for i < 2000000 {
		if 1 > 2 {
			acc = acc * 2
		}
		if 2*3 == 6 {
			acc = (acc + i) % 1000003
		} else {
			acc = 0
		}
		for false {
			acc = acc + 1
		}
		i = i + 1
	}
	fmt.Println(acc)
}
