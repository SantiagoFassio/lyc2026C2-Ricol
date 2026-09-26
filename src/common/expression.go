package common

import (
	"fmt"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type Expression interface {
	isExpression()
	Evaluate() (types.Value, error)
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

func (b *BinaryExpression) Evaluate() (types.Value, error) {
	resultLeft, err := b.LeftExpression.Evaluate()
	if err != nil {
		return resultLeft, err
	}
	resultRight, err := b.RightExpression.Evaluate()
	if err != nil {
		return resultRight, err
	}
	switch left := resultLeft.(type) {
	case types.Number:
		right, ok := resultRight.(types.Number)
		if ok {
			return b.evaluateNumbers(left, right)
		}
	case types.String:
		right, ok := resultRight.(types.String)
		if ok {
			return b.evaluateStrings(left, right)
		}
	case types.Boolean:
		right, ok := resultRight.(types.Boolean)
		if ok {
			return b.evaluateBooleans(left, right)
		}
	}
	return nil, b.unsupportedOperandsError(resultLeft, resultRight)
}

func (b *BinaryExpression) evaluateNumbers(left types.Number, right types.Number) (types.Value, error) {
	switch b.Operator.TokenType {
	case PLUS:
		return left.Add(right), nil
	case MINUS:
		return left.Substract(right), nil
	case STAR:
		return left.Multiply(right), nil
	case SLASH:
		return b.handleWithOperatorPosition(left.Divide(right))
	case DOUBLE_SLASH:
		return b.handleWithOperatorPosition(left.FloorDivide(right))
	case PERCENTAGE:
		return b.handleWithOperatorPosition(left.Modulo(right))
	case DOUBLE_STAR:
		return left.Power(right), nil
	case DOUBLE_EQUAL:
		return types.NewBoolean(left.Equals(right)), nil
	default:
		return nil, NewRicolError(b.Operator.Position, fmt.Sprintf("Invalid binary operator: %v", b.Operator))
	}
}

func (b *BinaryExpression) handleWithOperatorPosition(result types.Number, err error) (types.Value, error) {
	if err != nil {
		return nil, NewRicolError(b.Operator.Position, err.Error())
	}
	return result, nil
}

func (b *BinaryExpression) evaluateStrings(left types.String, right types.String) (types.Value, error) {
	switch b.Operator.TokenType {
	case PLUS:
		return left.Concatenate(right), nil
	case DOUBLE_EQUAL:
		return types.NewBoolean(left.Equals(right)), nil
	default:
		return nil, b.unsupportedOperandsError(left, right)
	}
}

func (b *BinaryExpression) evaluateBooleans(left types.Boolean, right types.Boolean) (types.Value, error) {
	if b.Operator.TokenType != DOUBLE_EQUAL {
		return nil, b.unsupportedOperandsError(left, right)
	}
	return types.NewBoolean(left.Equals(right)), nil
}

func (b *BinaryExpression) unsupportedOperandsError(left types.Value, right types.Value) error {
	return NewRicolError(b.Operator.Position,
		fmt.Sprintf("Unsupported operand types for %s: %s and %s", b.Operator.Lexeme, left.TypeName(), right.TypeName()))
}

func (g *GroupingExpression) Evaluate() (types.Value, error) {
	return g.Expression.Evaluate()
}

func (l *LiteralExpression) Evaluate() (types.Value, error) {
	return l.Value, nil
}

func (u *UnaryExpression) Evaluate() (types.Value, error) {
	expressionResult, err := u.Expression.Evaluate()
	if err != nil {
		return expressionResult, err
	}
	number, ok := expressionResult.(types.Number)
	if !ok {
		return nil, NewRicolError(u.Operator.Position,
			fmt.Sprintf("Unsupported operand type for %s: %s", u.Operator.Lexeme, expressionResult.TypeName()))
	}
	return number.Negate(), nil
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

func (*BinaryExpression) isExpression()   {}
func (*GroupingExpression) isExpression() {}
func (*LiteralExpression) isExpression()  {}
func (*UnaryExpression) isExpression()    {}
