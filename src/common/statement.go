package common

import "fmt"

type Statement interface {
	isStatement()
	String() string
}

type ExpressionStatement struct {
	Expression Expression
}

func NewExpressionStatement(expression Expression) *ExpressionStatement {
	return &ExpressionStatement{
		Expression: expression,
	}
}

func (e *ExpressionStatement) String() string {
	return fmt.Sprintf("%s;\n", e.Expression.String())
}

func (e *ExpressionStatement) isStatement() {}
