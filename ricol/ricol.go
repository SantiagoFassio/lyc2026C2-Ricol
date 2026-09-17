package ricol

import (
	"fmt"
	"os"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/parser"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/scanner"
)

type Ricol struct {
	filePath string
}

func NewRicol(filePath string) *Ricol {
	return &Ricol{
		filePath: filePath,
	}
}

func (r *Ricol) Run() error {
	fileContent, err := r.readFile()
	if err != nil {
		return err
	}
	tokens, err := scanner.NewScanner(fileContent).Scan()
	if err != nil {
		return err
	}
	fmt.Println("Scanning result:")
	r.printTokens(tokens)

	statements, err := parser.NewParser(tokens).Parse()
	if err != nil {
		return err
	}
	fmt.Println("---------------")
	fmt.Println("Parsing result:")
	r.printStatements(statements)
	return nil
}

// > 2 + 5
// NUMBER<2.0>
// PLUS
// NUMBER<5.0>
// EOF

func (r *Ricol) readFile() (string, error) {
	fileContent, err := os.ReadFile(r.filePath)
	if err != nil {
		return "", err
	}
	return string(fileContent), nil
}

func (r *Ricol) printTokens(tokens []common.Token) {
	for _, token := range tokens {
		fmt.Printf("%v\n", token)
	}
}

func (r *Ricol) printStatements(statements []common.Statement) {
	for _, statement := range statements {
		fmt.Print(statement)
	}
}
