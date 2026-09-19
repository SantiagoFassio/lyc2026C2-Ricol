package common

import (
	"fmt"
	"strconv"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type Expression interface {
	isExpression()
	Evaluate() (types.Number, error)
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

func (b BinaryExpression) Evaluate() (types.Number, error) {
	resultLeft, err := b.LeftExpression.Evaluate()
	if err != nil {
		return resultLeft, err
	}
	resultRight, err := b.RightExpression.Evaluate()
	if err != nil {
		return resultRight, err
	}
	switch b.Operator.TokenType {
	case PLUS:
		return resultLeft.Add(resultRight), nil
	case MINUS:
		return resultLeft.Substract(resultRight), nil
	case STAR:
		return resultLeft.Multiply(resultRight), nil
	case SLASH:
		return resultLeft.Divide(resultRight)
	case DOUBLE_SLASH:
		return resultLeft.DivideInteger(resultRight)
	case PERCENTAGE:
		return resultLeft.Modulo(resultRight)
	case DOUBLE_STAR:
		return resultLeft.Power(resultRight), nil
	default:
		return types.Integer(0), fmt.Errorf("Invalid binary operator: %v", b.Operator)
	}
}

func (g GroupingExpression) Evaluate() (types.Number, error) {
	return g.Expression.Evaluate()
}

func (l LiteralExpression) Evaluate() (types.Number, error) {
	// TODO: Move casting logic to parser
	switch l.Value.TokenType {
	case INTEGER:
		value, err := strconv.ParseInt(l.Value.Lexeme, 10, 64)
		if err != nil {
			return types.Integer(0), fmt.Errorf("Invalid integer: %s", l.Value.Lexeme)
		}
		return types.NewInteger(value), nil
	case FLOAT:
		value, err := strconv.ParseFloat(l.Value.Lexeme, 64)
		if err != nil {
			return types.Integer(0), fmt.Errorf("Invalid float: %s", l.Value.Lexeme)
		}
		return types.NewFloat(value), nil
	default:
		return types.Integer(0), fmt.Errorf("Invalid literal: %s", l.Value.Lexeme)
	}
}

func (u UnaryExpression) Evaluate() (types.Number, error) {
	expressionResult, err := u.Expression.Evaluate()
	if err != nil {
		return expressionResult, err
	}
	return expressionResult.Negate(), nil
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
