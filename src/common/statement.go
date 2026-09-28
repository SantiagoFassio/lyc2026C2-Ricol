package common

import (
	"fmt"
	"strings"
)

type Statement interface {
	isStatement()
	String() string
}

type ExpressionStatement struct {
	Expression Expression
}

type PrintStatement struct {
	Expression Expression
}

type BlockStatement struct {
	Statements []Statement
}

type IfStatement struct {
	IfToken    Token
	Condition  Expression
	IfBranch   Statement
	ElseBranch Statement
}

func NewExpressionStatement(expression Expression) *ExpressionStatement {
	return &ExpressionStatement{
		Expression: expression,
	}
}

func NewPrintStatement(expression Expression) *PrintStatement {
	return &PrintStatement{
		Expression: expression,
	}
}

func NewBlockStatement(statements []Statement) *BlockStatement {
	return &BlockStatement{
		Statements: statements,
	}
}

func NewIfStatement(ifToken Token, condition Expression, ifBranch Statement, elseBranch Statement) *IfStatement {
	return &IfStatement{
		IfToken:    ifToken,
		Condition:  condition,
		IfBranch:   ifBranch,
		ElseBranch: elseBranch,
	}
}

func (e *ExpressionStatement) String() string {
	return fmt.Sprintf("%s;\n", e.Expression.String())
}

func (p *PrintStatement) String() string {
	return fmt.Sprintf("PRINT %v;\n", p.Expression)
}

func (b *BlockStatement) String() string {
	var stringBuilder strings.Builder
	stringBuilder.WriteString("{\n")
	for _, statement := range b.Statements {
		for _, line := range strings.SplitAfter(statement.String(), "\n") {
			if line != "" {
				stringBuilder.WriteString("  " + line)
			}
		}
	}
	stringBuilder.WriteString("}\n")
	return stringBuilder.String()
}

func (i *IfStatement) String() string {
	ifString := fmt.Sprintf("if (%v) %v", i.Condition, i.IfBranch)
	if i.ElseBranch == nil {
		return ifString
	}
	return fmt.Sprintf("%s else %v", strings.TrimSuffix(ifString, "\n"), i.ElseBranch)
}

func (e *ExpressionStatement) isStatement() {}
func (p *PrintStatement) isStatement()      {}
func (b *BlockStatement) isStatement()      {}
func (i *IfStatement) isStatement()         {}
