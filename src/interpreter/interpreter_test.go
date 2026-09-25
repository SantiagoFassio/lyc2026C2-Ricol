package interpreter_test

import (
	"io"
	"strconv"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/interpreter"
)

type interpreterTestCase struct {
	name       string
	statements []common.Statement
}

type interpreterErrorTestCase struct {
	name            string
	statements      []common.Statement
	expectedMessage string
}

func token(tokenType common.TokenType, lexeme string) common.Token {
	return common.NewToken(tokenType, lexeme, common.Position{})
}

func integerLiteral(value int64) *common.LiteralExpression {
	return common.NewLiteralExpression(token(common.INTEGER, strconv.FormatInt(value, 10)), types.NewInteger(value))
}

func binary(leftExpression common.Expression, tokenType common.TokenType, lexeme string, rightExpression common.Expression) *common.BinaryExpression {
	return common.NewBinaryExpression(leftExpression, token(tokenType, lexeme), rightExpression)
}

func grouping(expression common.Expression) *common.GroupingExpression {
	return common.NewGroupingExpression(token(common.OPEN_PAR, "("), expression)
}

func statements(expressions ...common.Expression) []common.Statement {
	inputStatements := []common.Statement{}
	for _, expression := range expressions {
		inputStatements = append(inputStatements, common.NewExpressionStatement(expression))
	}
	return inputStatements
}

func divisionByZero() *common.BinaryExpression {
	return binary(integerLiteral(1), common.SLASH, "/", integerLiteral(0))
}

func moduloByZero() *common.BinaryExpression {
	return binary(integerLiteral(2), common.PERCENTAGE, "%", integerLiteral(0))
}

func invalidOperator() *common.BinaryExpression {
	return binary(integerLiteral(1), common.DOT, ".", integerLiteral(2))
}

func assertInterpret(t *testing.T, inputStatements []common.Statement) {
	t.Helper()

	err := interpreter.NewInterpreter(inputStatements, io.Discard).Interpret()

	if err != nil {
		t.Errorf("interpreter.Interpret(%v) unexpected error: %v", inputStatements, err)
	}
}

func assertInterpretError(t *testing.T, inputStatements []common.Statement, expectedMessage string) {
	t.Helper()

	err := interpreter.NewInterpreter(inputStatements, io.Discard).Interpret()

	if err == nil {
		t.Fatalf("interpreter.Interpret(%v) = nil error; want %q", inputStatements, expectedMessage)
	}
	if err.Error() != expectedMessage {
		t.Errorf("interpreter.Interpret(%v) error = %q; want %q", inputStatements, err, expectedMessage)
	}
}

func runInterpretTestCases(t *testing.T, testCases []interpreterTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertInterpret(t, testCase.statements)
		})
	}
}

func runInterpretErrorTestCases(t *testing.T, testCases []interpreterErrorTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertInterpretError(t, testCase.statements, testCase.expectedMessage)
		})
	}
}

func TestNoStatements(t *testing.T) {
	runInterpretTestCases(t, []interpreterTestCase{
		{"nil statements", nil},
		{"empty statements", []common.Statement{}},
		{"no expressions", statements()},
	})
}

func TestValidStatements(t *testing.T) {
	runInterpretTestCases(t, []interpreterTestCase{
		{"single literal", statements(integerLiteral(3))},
		{
			"single binary expression",
			statements(binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2))),
		},
		{
			"several statements",
			statements(
				integerLiteral(3),
				binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2)),
				common.NewUnaryExpression(token(common.MINUS, "-"), integerLiteral(4)),
				grouping(binary(integerLiteral(10), common.SLASH, "/", integerLiteral(4))),
			),
		},
		{
			"the same statement repeated",
			statements(integerLiteral(1), integerLiteral(1), integerLiteral(1)),
		},
	})
}

func TestStatementError(t *testing.T) {
	runInterpretErrorTestCases(t, []interpreterErrorTestCase{
		{
			"only statement fails",
			statements(divisionByZero()),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"first statement fails",
			statements(divisionByZero(), integerLiteral(3)),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"middle statement fails",
			statements(integerLiteral(3), moduloByZero(), integerLiteral(4)),
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
		{
			"last statement fails",
			statements(integerLiteral(3), integerLiteral(4), invalidOperator()),
			"[line 0, column 0] Invalid binary operator: DOT<.>",
		},
		{
			"the error of the first failing statement is returned",
			statements(integerLiteral(3), moduloByZero(), divisionByZero(), invalidOperator()),
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
		{
			"the error is propagated from a nested expression",
			statements(grouping(
				common.NewUnaryExpression(token(common.MINUS, "-"), divisionByZero()),
			)),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
	})
}
