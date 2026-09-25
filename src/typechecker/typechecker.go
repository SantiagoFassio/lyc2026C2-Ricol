package typechecker

import (
	"fmt"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type TypeChecker struct {
	statements []common.Statement
	errors     common.RicolErrorList
}

func NewTypeChecker(statements []common.Statement) *TypeChecker {
	return &TypeChecker{
		statements: statements,
	}
}

func (t *TypeChecker) Check() common.RicolErrorList {
	for _, statement := range t.statements {
		t.checkStatement(statement)
	}
	return t.errors
}

func (t *TypeChecker) checkStatement(statement common.Statement) {
	switch typedStatement := statement.(type) {
	case *common.ExpressionStatement:
		t.checkExpression(typedStatement.Expression)
	default:
		panic(fmt.Sprintf("Unknown statement node: %T", statement))
	}
}

func (t *TypeChecker) checkExpression(expression common.Expression) types.Type {
	switch typedExpression := expression.(type) {
	case *common.BinaryExpression:
		return t.checkBinaryExpression(typedExpression)
	case *common.GroupingExpression:
		return t.checkExpression(typedExpression.Expression)
	case *common.LiteralExpression:
		return t.checkLiteralExpression(typedExpression)
	case *common.UnaryExpression:
		return t.checkUnaryExpression(typedExpression)
	default:
		panic(fmt.Sprintf("Unknown expression node: %T", expression))
	}
}

func (t *TypeChecker) checkLiteralExpression(expression *common.LiteralExpression) types.Type {
	switch expression.Token.TokenType {
	case common.INTEGER:
		return types.Int
	case common.FLOAT:
		return types.Float
	case common.STRING:
		return types.Str
	default:
		panic(fmt.Sprintf("Unknown literal token: %v", expression.Token))
	}
}

func (t *TypeChecker) checkUnaryExpression(expression *common.UnaryExpression) types.Type {
	operandType := t.checkExpression(expression.Expression)
	if operandType == types.Invalid {
		return types.Invalid
	}
	if expression.Operator.TokenType != common.MINUS {
		return t.reportError(expression.Operator.Position,
			fmt.Sprintf("Invalid unary operator: %v", expression.Operator))
	}
	if !isNumeric(operandType) {
		return t.reportError(expression.Operator.Position,
			fmt.Sprintf("Unsupported operand type for %s: %s", expression.Operator.Lexeme, operandType))
	}
	return operandType
}

func (t *TypeChecker) checkBinaryExpression(expression *common.BinaryExpression) types.Type {
	leftType := t.checkExpression(expression.LeftExpression)
	rightType := t.checkExpression(expression.RightExpression)
	if leftType == types.Invalid || rightType == types.Invalid {
		return types.Invalid
	}
	if leftType == types.Str && rightType == types.Str && expression.Operator.TokenType == common.PLUS {
		return types.Str
	}
	if isNumeric(leftType) && isNumeric(rightType) {
		return t.checkNumericOperation(expression.Operator, leftType, rightType)
	}
	return t.reportError(expression.Operator.Position,
		fmt.Sprintf("Unsupported operand types for %s: %s and %s", expression.Operator.Lexeme, leftType, rightType))
}

func (t *TypeChecker) checkNumericOperation(operator common.Token, leftType types.Type, rightType types.Type) types.Type {
	switch operator.TokenType {
	case common.PLUS, common.MINUS, common.STAR, common.DOUBLE_SLASH, common.PERCENTAGE, common.DOUBLE_STAR:
		return numericResult(leftType, rightType)
	case common.SLASH:
		return types.Float
	default:
		return t.reportError(operator.Position, fmt.Sprintf("Invalid binary operator: %v", operator))
	}
}

func (t *TypeChecker) reportError(position common.Position, message string) types.Type {
	t.errors = append(t.errors, common.NewRicolError(position, message))
	return types.Invalid
}

func numericResult(leftType types.Type, rightType types.Type) types.Type {
	if leftType == types.Int && rightType == types.Int {
		return types.Int
	}
	return types.Float
}

func isNumeric(expressionType types.Type) bool {
	return expressionType == types.Int || expressionType == types.Float
}
