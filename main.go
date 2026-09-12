package main

import (
	"fmt"
	"os"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/ricol"
)

func main() {
	ricolArgs := os.Args[1:]
	if len(ricolArgs) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: ricol <file>")
	}
	file := ricolArgs[1]
	ricol := ricol.NewRicol(file)
	err := ricol.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "An error occurred:", err)
	}
}
