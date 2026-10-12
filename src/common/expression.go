package common

import (
	"fmt"
	"strings"

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

type CallExpression struct {
	NameToken Token
	Arguments []Expression
}

// FieldInitializer representa un campo inicializado en un literal de struct
type FieldInitializer struct {
	NameToken       Token
	ValueExpression Expression
}

// Expresion que representa la creación de un literal de struct
// Incluye su nombre y los campos inicializados
type StructLiteralExpression struct {
	NameToken Token
	Fields    []FieldInitializer
}

// Expresion que representa el acceso a un campo de un struct
type FieldAccessExpression struct {
	Expression Expression
	NameToken  Token
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

func NewCallExpression(nameToken Token, arguments []Expression) *CallExpression {
	return &CallExpression{
		NameToken: nameToken,
		Arguments: arguments,
	}
}

func NewFieldInitializer(nameToken Token, valueExpression Expression) FieldInitializer {
	return FieldInitializer{
		NameToken:       nameToken,
		ValueExpression: valueExpression,
	}
}

func NewStructLiteralExpression(nameToken Token, fields []FieldInitializer) *StructLiteralExpression {
	return &StructLiteralExpression{
		NameToken: nameToken,
		Fields:    fields,
	}
}

func NewFieldAccessExpression(expression Expression, nameToken Token) *FieldAccessExpression {
	return &FieldAccessExpression{
		Expression: expression,
		NameToken:  nameToken,
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

func (c *CallExpression) String() string {
	arguments := make([]string, len(c.Arguments))
	for i, argument := range c.Arguments {
		arguments[i] = argument.String()
	}
	return fmt.Sprintf("%s(%s)", c.NameToken.Lexeme, strings.Join(arguments, ", "))
}

func (s *StructLiteralExpression) String() string {
	if len(s.Fields) == 0 {
		return s.NameToken.Lexeme + " {}"
	}
	fields := make([]string, len(s.Fields))
	for i, field := range s.Fields {
		fields[i] = fmt.Sprintf("%s : %v", field.NameToken.Lexeme, field.ValueExpression)
	}
	return fmt.Sprintf("%s { %s }", s.NameToken.Lexeme, strings.Join(fields, ", "))
}

func (f *FieldAccessExpression) String() string {
	return fmt.Sprintf("%v.%s", f.Expression, f.NameToken.Lexeme)
}

func (*BinaryExpression) isExpression()        {}
func (*GroupingExpression) isExpression()      {}
func (*LiteralExpression) isExpression()       {}
func (*UnaryExpression) isExpression()         {}
func (*VariableExpression) isExpression()      {}
func (*VarAssignmentExpression) isExpression() {}
func (*CallExpression) isExpression()          {}
func (*StructLiteralExpression) isExpression() {}
func (*FieldAccessExpression) isExpression()   {}
