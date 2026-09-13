package common

type Expression interface {
	isExpression()
}

type BinaryExpression struct {
	LeftExpression  Expression
	Operator        Token
	RightExpression Expression
}

type GroupingExpression struct {
	Expression Expression
}

type LiteralExpression struct {
	Value Token
}

type UnaryExpression struct {
	Operator   Token
	Expression Expression
}

func (BinaryExpression) isExpression()   {}
func (GroupingExpression) isExpression() {}
func (LiteralExpression) isExpression()  {}
func (UnaryExpression) isExpression()    {}
