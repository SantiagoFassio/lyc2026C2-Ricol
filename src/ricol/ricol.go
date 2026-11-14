package ricol

import (
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/interpreter"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/loader"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/parser"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/scanner"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/typechecker"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/wire"
)

type Mode int

const (
	ModeFull Mode = iota
	ModeScan
	ModeParse
	ModeTypeCheck
)

type Ricol struct {
	filePath string
	mode     Mode
	output   io.Writer
}

func NewRicol(filePath string, mode Mode, output io.Writer) *Ricol {
	return &Ricol{
		filePath: filePath,
		mode:     mode,
		output:   output,
	}
}

func (r *Ricol) Run() error {
	fileContent, err := r.readFile()
	if err != nil {
		return err
	}
	tokens, err := scanner.NewScanner(fileContent, r.filePath).Scan()
	if err != nil {
		return err
	}
	if r.mode == ModeScan {
		r.printTokens(tokens)
		return nil
	}

	statements, err := parser.NewParser(tokens).Parse()
	if err != nil {
		return err
	}
	if r.mode == ModeParse {
		r.printStatements(statements)
		return nil
	}

	modules, err := loader.NewLoader(r.filePath, statements, os.ReadFile).Load()
	if err != nil {
		return err
	}

	checker := typechecker.NewTypeChecker(modules)
	distances, checkErrors := checker.Check()
	if len(checkErrors) > 0 {
		return checkErrors
	}
	if r.mode == ModeTypeCheck {
		fmt.Fprintln(r.output, "Type checking OK")
		return nil
	}

	program := interpreter.NewInterpreter(slices.Concat(modules...), distances, r.output)
	program.EnableNetwork(interfacesOf(checker.Services()))
	if err := program.Interpret(); err != nil {
		return err
	}
	return program.Serve()
}

func interfacesOf(services []*types.ServiceType) map[string]*wire.Interface {
	interfaces := make(map[string]*wire.Interface, len(services))
	for _, service := range services {
		interfaces[service.Name] = wire.NewInterface(service)
	}
	return interfaces
}

func (r *Ricol) readFile() (string, error) {
	fileContent, err := os.ReadFile(r.filePath)
	if err != nil {
		return "", err
	}
	return string(fileContent), nil
}

func (r *Ricol) printTokens(tokens []common.Token) {
	for _, token := range tokens {
		fmt.Fprintf(r.output, "%v\n", token)
	}
}

func (r *Ricol) printStatements(statements []common.Statement) {
	for _, statement := range statements {
		fmt.Fprint(r.output, statement)
	}
}
