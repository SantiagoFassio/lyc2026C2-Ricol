package common

import "fmt"

type Statement interface {
	isStatement()
	Execute() error
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

func (e *ExpressionStatement) Execute() error {
	expressionResult, err := e.Expression.Evaluate()
	if err != nil {
		return err
	}
	fmt.Printf("[TEMPORAL] Expression statement result: %v\n", expressionResult)
	return err
}

func (e *ExpressionStatement) String() string {
	return fmt.Sprintf("%s;\n", e.Expression.String())
}

func (e *ExpressionStatement) isStatement() {}
