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
	case *common.PrintStatement:
		t.checkExpression(typedStatement.Expression)
	case *common.BlockStatement:
		for _, statement := range typedStatement.Statements {
			t.checkStatement(statement)
		}
	case *common.IfStatement:
		t.checkIfStatement(typedStatement)
	case *common.ExpressionStatement:
		t.checkExpression(typedStatement.Expression)
	default:
		panic(fmt.Sprintf("Unknown statement node: %T", statement))
	}
}

func (t *TypeChecker) checkIfStatement(ifStatement *common.IfStatement) {
	conditionType := t.checkExpression(ifStatement.Condition)
	if conditionType != types.Invalid && !isBool(conditionType) {
		t.reportError(ifStatement.IfToken.Position,
			fmt.Sprintf("Non boolean expression in if condition: %s", conditionType))
	}
	t.checkStatement(ifStatement.IfBranch)
	if ifStatement.ElseBranch != nil {
		t.checkStatement(ifStatement.ElseBranch)
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
	case common.TRUE, common.FALSE:
		return types.Bool
	default:
		panic(fmt.Sprintf("Unknown literal token: %v", expression.Token))
	}
}

func (t *TypeChecker) checkUnaryExpression(expression *common.UnaryExpression) types.Type {
	operandType := t.checkExpression(expression.Expression)
	if operandType == types.Invalid {
		return types.Invalid
	}
	switch expression.Operator.TokenType {
	case common.MINUS:
		if !isNumeric(operandType) {
			return t.unsupportedOperandError(expression.Operator, operandType)
		}
		return operandType
	case common.NOT:
		if !isBool(operandType) {
			return t.unsupportedOperandError(expression.Operator, operandType)
		}
		return types.Bool
	default:
		return t.reportError(expression.Operator.Position,
			fmt.Sprintf("Invalid unary operator: %v", expression.Operator))
	}
}

func (t *TypeChecker) unsupportedOperandError(operator common.Token, operandType types.Type) types.Type {
	return t.reportError(operator.Position,
		fmt.Sprintf("Unsupported operand type for %s: %s", operator.Lexeme, operandType))
}

func (t *TypeChecker) checkBinaryExpression(expression *common.BinaryExpression) types.Type {
	leftType := t.checkExpression(expression.LeftExpression)
	rightType := t.checkExpression(expression.RightExpression)
	if leftType == types.Invalid || rightType == types.Invalid {
		return types.Invalid
	}
	if isLogicalOperator(expression.Operator) {
		return t.checkLogicalOperation(expression.Operator, leftType, rightType)
	}
	if isComparisonOperator(expression.Operator) {
		return t.checkComparison(expression.Operator, leftType, rightType)
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

// Se comparan numeros con numeros, strings con strings y booleanos con booleanos.
// Los booleanos solo se comparan por igualdad, porque no tienen orden.
func (t *TypeChecker) checkComparison(operator common.Token, leftType types.Type, rightType types.Type) types.Type {
	sameKind := (isNumeric(leftType) && isNumeric(rightType)) || leftType == rightType
	isOrdering := operator.TokenType != common.DOUBLE_EQUAL && operator.TokenType != common.NOT_EQUAL
	if !sameKind || (isOrdering && isBool(leftType)) {
		return t.reportError(operator.Position,
			fmt.Sprintf("Unsupported operand types for %s: %s and %s", operator.Lexeme, leftType, rightType))
	}
	return types.Bool
}

func (t *TypeChecker) checkLogicalOperation(operator common.Token, leftType types.Type, rightType types.Type) types.Type {
	if !isBool(leftType) || !isBool(rightType) {
		return t.reportError(operator.Position,
			fmt.Sprintf("Unsupported operand types for %s: %s and %s", operator.Lexeme, leftType, rightType))
	}
	return types.Bool
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

func isLogicalOperator(operator common.Token) bool {
	return operator.TokenType == common.AND || operator.TokenType == common.OR
}

func isComparisonOperator(operator common.Token) bool {
	switch operator.TokenType {
	case common.DOUBLE_EQUAL, common.NOT_EQUAL, common.LESS, common.LESS_EQUAL, common.GREATER, common.GREATER_EQUAL:
		return true
	default:
		return false
	}
}

func isNumeric(expressionType types.Type) bool {
	return expressionType == types.Int || expressionType == types.Float
}

func isBool(expressionType types.Type) bool {
	return expressionType == types.Bool
}
