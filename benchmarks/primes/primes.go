// Nested loops with conditions: counts the primes below N by trial division
package main

import "fmt"

func isPrime(n int64) bool {
	if n < 2 {
		return false
	}
	var d int64 = 2
	for d*d <= n {
		if n%d == 0 {
			return false
		}
		d = d + 1
	}
	return true
}

func main() {
	var count, n int64 = 0, 0
	for n < 200000 {
		if isPrime(n) {
			count = count + 1
		}
		n = n + 1
	}
	fmt.Println(count)
}
