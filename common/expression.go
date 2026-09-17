package common

import (
	"fmt"
)

type Expression interface {
	isExpression()
	String() string
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

func (b BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", b.LeftExpression.String(), b.Operator.String(), b.RightExpression.String())
}

func (g GroupingExpression) String() string {
	return fmt.Sprintf("(%s)", g.Expression.String())
}

func (l LiteralExpression) String() string {
	return l.Value.String()
}

func (u UnaryExpression) String() string {
	return fmt.Sprintf("(%s %s)", u.Operator.String(), u.Expression.String())
}

func (BinaryExpression) isExpression()   {}
func (GroupingExpression) isExpression() {}
func (LiteralExpression) isExpression()  {}
func (UnaryExpression) isExpression()    {}
