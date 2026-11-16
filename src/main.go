package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/ricol"
)

func main() {
	scanFlag := flag.Bool("scan", false, "run only the scanning phase and print the tokens")
	parseFlag := flag.Bool("parse", false, "run only up to the parsing phase and print the statements")
	typecheckFlag := flag.Bool("typecheck", false, "run only up to the type checking phase")
	optimizeFlag := flag.Bool("optimize", false, "run only up to the optimization phase and print the optimized statements")
	cpuProfileFlag := flag.String("cpuprofile", "", "write a CPU profile of the run to the given file")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: ricol [--scan | --parse | --typecheck | --optimize] [--cpuprofile <file>] <file>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	mode, err := modeFromFlags(*scanFlag, *parseFlag, *typecheckFlag, *optimizeFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		os.Exit(1)
	}

	stopProfile := func() {}
	if *cpuProfileFlag != "" {
		stopProfile, err = startCPUProfile(*cpuProfileFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, "An error occurred:", err)
			os.Exit(1)
		}
	}
	defer stopProfile()

	file := flag.Arg(0)
	ricol := ricol.NewRicol(file, mode, os.Stdin, os.Stdout)
	err = ricol.Run()
	if err != nil {
		stopProfile()
		fmt.Fprintln(os.Stderr, "An error occurred:", err)
		os.Exit(1)
	}
}

func modeFromFlags(scanFlag, parseFlag, typecheckFlag, optimizeFlag bool) (ricol.Mode, error) {
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
	if optimizeFlag {
		selectedModes++
		mode = ricol.ModeOptimize
	}
	if selectedModes > 1 {
		return ricol.ModeFull, fmt.Errorf("flags --scan, --parse, --typecheck and --optimize are mutually exclusive")
	}
	return mode, nil
}

func startCPUProfile(path string) (func(), error) {
	profileFile, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if err := pprof.StartCPUProfile(profileFile); err != nil {
		profileFile.Close()
		return nil, err
	}
	return func() {
		pprof.StopCPUProfile()
		profileFile.Close()
	}, nil
}
