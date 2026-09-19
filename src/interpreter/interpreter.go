package interpreter

import "github.com/SantiagoFassio/lyc2026C2-Ricol/common"

type Interpreter struct {
	statements []common.Statement
}

func NewInterpreter(statements []common.Statement) *Interpreter {
	return &Interpreter{
		statements: statements,
	}
}

func (i *Interpreter) Interpret() error {
	for _, statement := range i.statements {
		err := statement.Execute()
		if err != nil {
			return err
		}
	}
	return nil
}
