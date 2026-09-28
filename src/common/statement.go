package common

import (
	"fmt"
	"strings"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
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

type WhileStatement struct {
	WhileToken Token
	Condition  Expression
	Body       Statement
}

type ContinueStatement struct {
	ContinueToken Token
}

type BreakStatement struct {
	BreakToken Token
}

type VarDeclarationStatement struct {
	LetToken        Token
	NameToken       Token
	VarType         types.Type
	ValueExpression Expression
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

func NewWhileStatement(whileToken Token, condition Expression, body Statement) *WhileStatement {
	return &WhileStatement{
		WhileToken: whileToken,
		Condition:  condition,
		Body:       body,
	}
}

func NewContinueStatement(continueToken Token) *ContinueStatement {
	return &ContinueStatement{
		ContinueToken: continueToken,
	}
}

func NewBreakStatement(breakToken Token) *BreakStatement {
	return &BreakStatement{
		BreakToken: breakToken,
	}
}

func NewVarDeclarationStatement(
	letToken Token,
	nameToken Token,
	varType types.Type,
	valueExpression Expression,
) *VarDeclarationStatement {
	return &VarDeclarationStatement{
		LetToken:        letToken,
		NameToken:       nameToken,
		VarType:         varType,
		ValueExpression: valueExpression,
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

func (w *WhileStatement) String() string {
	return fmt.Sprintf("while (%v) %v", w.Condition, w.Body)
}

func (c *ContinueStatement) String() string {
	return "continue;\n"
}

func (b *BreakStatement) String() string {
	return "break;\n"
}

func (v *VarDeclarationStatement) String() string {
	return fmt.Sprintf("let %s : %v = %v;\n", v.NameToken.Lexeme, v.VarType, v.ValueExpression)
}

func (e *ExpressionStatement) isStatement()     {}
func (p *PrintStatement) isStatement()          {}
func (b *BlockStatement) isStatement()          {}
func (i *IfStatement) isStatement()             {}
func (w *WhileStatement) isStatement()          {}
func (c *ContinueStatement) isStatement()       {}
func (b *BreakStatement) isStatement()          {}
func (v *VarDeclarationStatement) isStatement() {}
