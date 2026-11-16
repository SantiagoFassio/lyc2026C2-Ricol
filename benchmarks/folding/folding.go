// Loop whose arithmetic mixes variables with subexpressions made only of literals
package main

import "fmt"

func main() {
	var i, acc int64 = 0, 0
	for i < 2000*1000 {
		acc = (acc + i*(60*60*24) + (1000 - 7*3) - -5) % 1000003
		i = i + 1
	}
	fmt.Println(acc)
}
