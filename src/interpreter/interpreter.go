package interpreter

import (
	"fmt"
	"io"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type Interpreter struct {
	statements []common.Statement
	output     io.Writer
}

func NewInterpreter(statements []common.Statement, output io.Writer) *Interpreter {
	return &Interpreter{
		statements: statements,
		output:     output,
	}
}

func (i *Interpreter) Interpret() error {
	for _, statement := range i.statements {
		err := i.execute(statement)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) execute(statement common.Statement) error {
	switch typedStatement := statement.(type) {
	case *common.PrintStatement:
		return i.executePrintStatement(typedStatement)
	case *common.BlockStatement:
		return i.executeBlockStatement(typedStatement)
	case *common.IfStatement:
		return i.executeIfStatement(typedStatement)
	case *common.ExpressionStatement:
		return i.executeExpressionStatement(typedStatement)
	default:
		panic(fmt.Sprintf("Unknown statement node: %T", statement))
	}
}

func (i *Interpreter) executePrintStatement(statement *common.PrintStatement) error {
	expressionResult, err := i.evaluate(statement.Expression)
	if err != nil {
		return err
	}
	fmt.Fprint(i.output, expressionResult.Display())
	return nil
}

func (i *Interpreter) executeBlockStatement(statement *common.BlockStatement) error {
	for _, statement := range statement.Statements {
		err := i.execute(statement)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) executeIfStatement(statement *common.IfStatement) error {
	conditionResult, err := i.evaluate(statement.Condition)
	if err != nil {
		return err
	}
	condition, ok := conditionResult.(types.Boolean)
	if !ok {
		panic(fmt.Sprintf("Non boolean expression used as if condition: %v", statement.Condition))
	}
	if condition.IsTrue() {
		return i.execute(statement.IfBranch)
	}
	if statement.ElseBranch != nil {
		return i.execute(statement.ElseBranch)
	}
	return nil
}

func (i *Interpreter) executeExpressionStatement(statement *common.ExpressionStatement) error {
	_, err := i.evaluate(statement.Expression)
	return err
}

func (i *Interpreter) evaluate(expression common.Expression) (types.Value, error) {
	switch typedExpression := expression.(type) {
	case *common.BinaryExpression:
		return i.evaluateBinaryExpression(typedExpression)
	case *common.GroupingExpression:
		return i.evaluate(typedExpression.Expression)
	case *common.LiteralExpression:
		return typedExpression.Value, nil
	case *common.UnaryExpression:
		return i.evaluateUnaryExpression(typedExpression)
	default:
		panic(fmt.Sprintf("Unknown expression node: %T", expression))
	}
}

func (i *Interpreter) evaluateBinaryExpression(expression *common.BinaryExpression) (types.Value, error) {
	if isLogicalOperator(expression.Operator) {
		return i.evaluateLogicalExpression(expression)
	}
	resultLeft, err := i.evaluate(expression.LeftExpression)
	if err != nil {
		return resultLeft, err
	}
	resultRight, err := i.evaluate(expression.RightExpression)
	if err != nil {
		return resultRight, err
	}
	switch left := resultLeft.(type) {
	case types.Number:
		right, ok := resultRight.(types.Number)
		if ok {
			return i.evaluateNumbers(expression.Operator, left, right)
		}
	case types.String:
		right, ok := resultRight.(types.String)
		if ok {
			return i.evaluateStrings(expression.Operator, left, right)
		}
	case types.Boolean:
		right, ok := resultRight.(types.Boolean)
		if ok {
			return i.evaluateBooleans(expression.Operator, left, right)
		}
	}
	return nil, i.unsupportedOperandsError(expression.Operator, resultLeft, resultRight)
}

func (i *Interpreter) evaluateLogicalExpression(expression *common.BinaryExpression) (types.Value, error) {
	resultLeft, err := i.evaluate(expression.LeftExpression)
	if err != nil {
		return resultLeft, err
	}
	left, ok := resultLeft.(types.Boolean)
	if !ok {
		return nil, i.unsupportedOperandError(expression.Operator, resultLeft)
	}
	if isShortCircuit(expression.Operator, left) {
		return left, nil
	}
	resultRight, err := i.evaluate(expression.RightExpression)
	if err != nil {
		return resultRight, err
	}
	right, ok := resultRight.(types.Boolean)
	if !ok {
		return nil, i.unsupportedOperandsError(expression.Operator, left, resultRight)
	}
	return right, nil
}

func (i *Interpreter) evaluateUnaryExpression(expression *common.UnaryExpression) (types.Value, error) {
	expressionResult, err := i.evaluate(expression.Expression)
	if err != nil {
		return expressionResult, err
	}
	switch expression.Operator.TokenType {
	case common.MINUS:
		number, ok := expressionResult.(types.Number)
		if !ok {
			return nil, i.unsupportedOperandError(expression.Operator, expressionResult)
		}
		return number.Negate(), nil
	case common.NOT:
		boolean, ok := expressionResult.(types.Boolean)
		if !ok {
			return nil, i.unsupportedOperandError(expression.Operator, expressionResult)
		}
		return boolean.Not(), nil
	default:
		return nil, common.NewRicolError(expression.Operator.Position,
			fmt.Sprintf("Invalid unary operator: %v", expression.Operator))
	}
}

func (i *Interpreter) evaluateNumbers(operator common.Token, left types.Number, right types.Number) (types.Value, error) {
	switch operator.TokenType {
	case common.PLUS:
		return left.Add(right), nil
	case common.MINUS:
		return left.Substract(right), nil
	case common.STAR:
		return left.Multiply(right), nil
	case common.SLASH:
		result, err := left.Divide(right)
		return i.handleWithOperatorPosition(operator, result, err)
	case common.DOUBLE_SLASH:
		result, err := left.FloorDivide(right)
		return i.handleWithOperatorPosition(operator, result, err)
	case common.PERCENTAGE:
		result, err := left.Modulo(right)
		return i.handleWithOperatorPosition(operator, result, err)
	case common.DOUBLE_STAR:
		result, err := left.Power(right)
		return i.handleWithOperatorPosition(operator, result, err)
	case common.DOUBLE_EQUAL:
		return left.Equals(right), nil
	case common.NOT_EQUAL:
		return left.Equals(right).Not(), nil
	case common.LESS:
		return left.LessThan(right), nil
	case common.LESS_EQUAL:
		return left.LessOrEqualThan(right), nil
	case common.GREATER:
		return left.GreaterThan(right), nil
	case common.GREATER_EQUAL:
		return left.GreaterOrEqualThan(right), nil
	default:
		return nil, common.NewRicolError(operator.Position, fmt.Sprintf("Invalid binary operator: %v", operator))
	}
}

func (i *Interpreter) handleWithOperatorPosition(operator common.Token, result types.Number, err error) (types.Value, error) {
	if err != nil {
		return nil, common.NewRicolError(operator.Position, err.Error())
	}
	return result, nil
}

func (i *Interpreter) evaluateStrings(operator common.Token, left types.String, right types.String) (types.Value, error) {
	switch operator.TokenType {
	case common.PLUS:
		return left.Concatenate(right), nil
	case common.DOUBLE_EQUAL:
		return left.Equals(right), nil
	case common.NOT_EQUAL:
		return left.Equals(right).Not(), nil
	case common.LESS:
		return left.LessThan(right), nil
	case common.LESS_EQUAL:
		return left.LessOrEqualThan(right), nil
	case common.GREATER:
		return left.GreaterThan(right), nil
	case common.GREATER_EQUAL:
		return left.GreaterOrEqualThan(right), nil
	default:
		return nil, i.unsupportedOperandsError(operator, left, right)
	}
}

// Los booleanos se pueden comparar por igualdad, pero no tienen orden.
func (i *Interpreter) evaluateBooleans(operator common.Token, left types.Boolean, right types.Boolean) (types.Value, error) {
	switch operator.TokenType {
	case common.DOUBLE_EQUAL:
		return left.Equals(right), nil
	case common.NOT_EQUAL:
		return left.Equals(right).Not(), nil
	default:
		return nil, i.unsupportedOperandsError(operator, left, right)
	}
}

func (i *Interpreter) unsupportedOperandsError(operator common.Token, left types.Value, right types.Value) error {
	return common.NewRicolError(operator.Position,
		fmt.Sprintf("Unsupported operand types for %s: %s and %s", operator.Lexeme, left.TypeName(), right.TypeName()))
}

func (i *Interpreter) unsupportedOperandError(operator common.Token, operand types.Value) error {
	return common.NewRicolError(operator.Position,
		fmt.Sprintf("Unsupported operand type for %s: %s", operator.Lexeme, operand.TypeName()))
}

func isLogicalOperator(operator common.Token) bool {
	return operator.TokenType == common.AND || operator.TokenType == common.OR
}

func isShortCircuit(operator common.Token, left types.Boolean) bool {
	if operator.TokenType == common.AND {
		return !left.IsTrue()
	}
	return left.IsTrue()
}
