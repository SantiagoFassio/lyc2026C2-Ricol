package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/ricol"
)

func main() {
	scanFlag := flag.Bool("scan", false, "run only the scanning phase and print the tokens")
	parseFlag := flag.Bool("parse", false, "run only up to the parsing phase and print the statements")
	typecheckFlag := flag.Bool("typecheck", false, "run only up to the type checking phase")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: ricol [--scan | --parse | --typecheck] <file>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		return
	}

	mode, err := modeFromFlags(*scanFlag, *parseFlag, *typecheckFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		os.Exit(1)
	}

	file := flag.Arg(0)
	ricol := ricol.NewRicol(file, mode)
	err = ricol.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "An error occurred:", err)
	}
}

func modeFromFlags(scanFlag, parseFlag, typecheckFlag bool) (ricol.Mode, error) {
	selectedModes := 0
	mode := ricol.ModeFull
	if scanFlag {
		selectedModes++
		mode = ricol.ModeScan
	}
	if parseFlag {
		selectedModes++
		mode = ricol.ModeParse
	}
	if typecheckFlag {
		selectedModes++
		mode = ricol.ModeTypeCheck
	}
	if selectedModes > 1 {
		return ricol.ModeFull, fmt.Errorf("flags --scan, --parse and --typecheck are mutually exclusive")
	}
	return mode, nil
}
