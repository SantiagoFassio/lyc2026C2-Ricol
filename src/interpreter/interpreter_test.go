package interpreter_test

import (
	"bytes"
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

type interpreterOutputTestCase struct {
	name           string
	statements     []common.Statement
	expectedOutput string
}

type interpreterOutputErrorTestCase struct {
	name            string
	statements      []common.Statement
	expectedOutput  string
	expectedMessage string
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

func stringLiteral(value string) *common.LiteralExpression {
	return common.NewLiteralExpression(token(common.STRING, strconv.Quote(value)), types.NewString(value))
}

func booleanLiteral(value bool) *common.LiteralExpression {
	if value {
		return common.NewLiteralExpression(token(common.TRUE, "True"), types.NewBoolean(true))
	}
	return common.NewLiteralExpression(token(common.FALSE, "False"), types.NewBoolean(false))
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

func printStatements(expressions ...common.Expression) []common.Statement {
	inputStatements := []common.Statement{}
	for _, expression := range expressions {
		inputStatements = append(inputStatements, common.NewPrintStatement(expression))
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

func assertInterpretOutput(t *testing.T, inputStatements []common.Statement, expectedOutput string) {
	t.Helper()

	output := &bytes.Buffer{}

	err := interpreter.NewInterpreter(inputStatements, output).Interpret()

	if err != nil {
		t.Fatalf("interpreter.Interpret(%v) unexpected error: %v", inputStatements, err)
	}
	if output.String() != expectedOutput {
		t.Errorf("interpreter.Interpret(%v) output = %q; want %q", inputStatements, output, expectedOutput)
	}
}

func assertInterpretOutputBeforeError(t *testing.T, inputStatements []common.Statement, expectedOutput string, expectedMessage string) {
	t.Helper()

	output := &bytes.Buffer{}

	err := interpreter.NewInterpreter(inputStatements, output).Interpret()

	if err == nil {
		t.Fatalf("interpreter.Interpret(%v) = nil error; want %q", inputStatements, expectedMessage)
	}
	if err.Error() != expectedMessage {
		t.Errorf("interpreter.Interpret(%v) error = %q; want %q", inputStatements, err, expectedMessage)
	}
	if output.String() != expectedOutput {
		t.Errorf("interpreter.Interpret(%v) output = %q; want %q", inputStatements, output, expectedOutput)
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

func runInterpretOutputTestCases(t *testing.T, testCases []interpreterOutputTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertInterpretOutput(t, testCase.statements, testCase.expectedOutput)
		})
	}
}

func runInterpretOutputErrorTestCases(t *testing.T, testCases []interpreterOutputErrorTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertInterpretOutputBeforeError(t, testCase.statements, testCase.expectedOutput, testCase.expectedMessage)
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

func TestPrintOutput(t *testing.T) {
	runInterpretOutputTestCases(t, []interpreterOutputTestCase{
		{"single print statement", printStatements(stringLiteral("hola")), "hola"},
		{"no line feed between print statements", printStatements(stringLiteral("a"), stringLiteral("b")), "ab"},
		{
			"line feed printed explicitly",
			printStatements(stringLiteral("a"), stringLiteral("\n"), stringLiteral("b")),
			"a\nb",
		},
		{
			"print statements are executed in order",
			printStatements(integerLiteral(1), binary(integerLiteral(1), common.PLUS, "+", integerLiteral(1)), integerLiteral(3)),
			"123",
		},
		{
			"numbers and strings",
			printStatements(
				stringLiteral("x = "),
				integerLiteral(3),
				stringLiteral(", y = "),
				binary(integerLiteral(1), common.SLASH, "/", integerLiteral(4)),
			),
			"x = 3, y = 0.25",
		},
	})
}

func TestPrintOutputBeforeError(t *testing.T) {
	runInterpretOutputErrorTestCases(t, []interpreterOutputErrorTestCase{
		{
			"first print statement fails",
			printStatements(divisionByZero(), stringLiteral("a")),
			"",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"output before the failing statement is kept",
			printStatements(stringLiteral("a"), stringLiteral("b"), moduloByZero(), stringLiteral("c")),
			"ab",
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
		{
			"failing expression statement after a print statement",
			append(printStatements(stringLiteral("a")), statements(invalidOperator())...),
			"a",
			"[line 0, column 0] Invalid binary operator: DOT<.>",
		},
	})
}

func TestPrintBooleans(t *testing.T) {
	runInterpretOutputTestCases(t, []interpreterOutputTestCase{
		{"boolean literals", printStatements(booleanLiteral(true), booleanLiteral(false)), "TrueFalse"},
		{
			"result of a comparison",
			printStatements(binary(integerLiteral(1), common.LESS, "<", integerLiteral(2))),
			"True",
		},
		{
			"result of an equality between strings",
			printStatements(binary(stringLiteral("hola"), common.DOUBLE_EQUAL, "==", stringLiteral("Hola"))),
			"False",
		},
	})
}
