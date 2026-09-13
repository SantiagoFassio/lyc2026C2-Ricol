package common

type Statement interface {
	isStatement()
}

type ExpressionStatement struct {
	Expression Expression
}

func (e ExpressionStatement) isStatement() {}
