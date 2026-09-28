package common

import (
	"fmt"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
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
	OpenPar    Token
	Expression Expression
}

type LiteralExpression struct {
	Token Token
	Value types.Value
}

type UnaryExpression struct {
	Operator   Token
	Expression Expression
}

type VariableExpression struct {
	NameToken Token
}

type VarAssignmentExpression struct {
	NameToken       Token
	ValueExpression Expression
}

func NewBinaryExpression(leftExpression Expression, operator Token, rightExpression Expression) *BinaryExpression {
	return &BinaryExpression{
		LeftExpression:  leftExpression,
		Operator:        operator,
		RightExpression: rightExpression,
	}
}

func NewGroupingExpression(openPar Token, expression Expression) *GroupingExpression {
	return &GroupingExpression{
		OpenPar:    openPar,
		Expression: expression,
	}
}

func NewLiteralExpression(token Token, value types.Value) *LiteralExpression {
	return &LiteralExpression{
		Token: token,
		Value: value,
	}
}

func NewUnaryExpression(operator Token, expression Expression) *UnaryExpression {
	return &UnaryExpression{
		Operator:   operator,
		Expression: expression,
	}
}

func NewVariableExpression(nameToken Token) *VariableExpression {
	return &VariableExpression{
		NameToken: nameToken,
	}
}

func NewVarAssignmentExpression(nameToken Token, valueExpression Expression) *VarAssignmentExpression {
	return &VarAssignmentExpression{
		NameToken:       nameToken,
		ValueExpression: valueExpression,
	}
}

func (b *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", b.LeftExpression.String(), b.Operator.String(), b.RightExpression.String())
}

func (g *GroupingExpression) String() string {
	return fmt.Sprintf("(%s)", g.Expression.String())
}

func (l *LiteralExpression) String() string {
	return l.Token.String()
}

func (u *UnaryExpression) String() string {
	return fmt.Sprintf("(%s %s)", u.Operator.String(), u.Expression.String())
}

func (v *VariableExpression) String() string {
	return v.NameToken.Lexeme
}

func (v *VarAssignmentExpression) String() string {
	return fmt.Sprintf("(%s = %v)", v.NameToken.Lexeme, v.ValueExpression)
}

func (*BinaryExpression) isExpression()        {}
func (*GroupingExpression) isExpression()      {}
func (*LiteralExpression) isExpression()       {}
func (*UnaryExpression) isExpression()         {}
func (*VariableExpression) isExpression()      {}
func (*VarAssignmentExpression) isExpression() {}
