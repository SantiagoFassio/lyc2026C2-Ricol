// Deeply nested linear recursion, close to the Ricol limit of 10000 nested calls
package main

import "fmt"

func sum(n int64) int64 {
	if n == 0 {
		return 0
	}
	return n + sum(n-1)
}

func main() {
	var total, k int64 = 0, 0
	for k < 300 {
		total = total + sum(9000+k%900)
		k = k + 1
	}
	fmt.Println(total)
}
