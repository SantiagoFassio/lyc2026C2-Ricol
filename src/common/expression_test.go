package common_test

import (
	"strconv"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type evaluateTestCase struct {
	name       string
	expression common.Expression
	expected   types.Number
}

type evaluateErrorTestCase struct {
	name            string
	expression      common.Expression
	expectedMessage string
}

type expressionStringTestCase struct {
	name       string
	expression common.Expression
	expected   string
}

func token(tokenType common.TokenType, lexeme string) common.Token {
	return common.NewToken(tokenType, lexeme)
}

func integerLiteral(value int64) *common.LiteralExpression {
	return common.NewLiteralExpression(token(common.INTEGER, strconv.FormatInt(value, 10)), types.NewInteger(value))
}

func floatLiteral(value float64) *common.LiteralExpression {
	return common.NewLiteralExpression(token(common.FLOAT, strconv.FormatFloat(value, 'f', -1, 64)), types.NewFloat(value))
}

func binary(leftExpression common.Expression, tokenType common.TokenType, lexeme string, rightExpression common.Expression) *common.BinaryExpression {
	return common.NewBinaryExpression(leftExpression, token(tokenType, lexeme), rightExpression)
}

func negation(expression common.Expression) *common.UnaryExpression {
	return common.NewUnaryExpression(token(common.MINUS, "-"), expression)
}

func grouping(expression common.Expression) *common.GroupingExpression {
	return common.NewGroupingExpression(expression)
}

func divisionByZero() *common.BinaryExpression {
	return binary(integerLiteral(1), common.SLASH, "/", integerLiteral(0))
}

func moduloByZero() *common.BinaryExpression {
	return binary(integerLiteral(2), common.PERCENTAGE, "%", integerLiteral(0))
}

func assertEvaluate(t *testing.T, expression common.Expression, expected types.Number) {
	t.Helper()

	result, err := expression.Evaluate()

	if err != nil {
		t.Fatalf("(%s).Evaluate() unexpected error: %v", expression, err)
	}
	if result != expected {
		t.Errorf("(%s).Evaluate() = %#v; want %#v", expression, result, expected)
	}
}

func assertEvaluateError(t *testing.T, expression common.Expression, expectedMessage string) {
	t.Helper()

	_, err := expression.Evaluate()

	if err == nil {
		t.Fatalf("(%s).Evaluate() = nil error; want %q", expression, expectedMessage)
	}
	if err.Error() != expectedMessage {
		t.Errorf("(%s).Evaluate() error = %q; want %q", expression, err, expectedMessage)
	}
}

func runEvaluateTestCases(t *testing.T, testCases []evaluateTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertEvaluate(t, testCase.expression, testCase.expected)
		})
	}
}

func runEvaluateErrorTestCases(t *testing.T, testCases []evaluateErrorTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertEvaluateError(t, testCase.expression, testCase.expectedMessage)
		})
	}
}

func runExpressionStringTestCases(t *testing.T, testCases []expressionStringTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.expression.String()

			if result != testCase.expected {
				t.Errorf("String() = %q; want %q", result, testCase.expected)
			}
		})
	}
}

func TestLiteralExpressionEvaluate(t *testing.T) {
	runEvaluateTestCases(t, []evaluateTestCase{
		{"integer", integerLiteral(3), types.NewInteger(3)},
		{"zero integer", integerLiteral(0), types.NewInteger(0)},
		{"float", floatLiteral(2.5), types.NewFloat(2.5)},
		{"zero float", floatLiteral(0), types.NewFloat(0)},
	})
}

func TestGroupingExpressionEvaluate(t *testing.T) {
	runEvaluateTestCases(t, []evaluateTestCase{
		{"grouped literal", grouping(integerLiteral(3)), types.NewInteger(3)},
		{"nested groupings", grouping(grouping(integerLiteral(3))), types.NewInteger(3)},
		{
			"grouped binary expression",
			grouping(binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2))),
			types.NewInteger(3),
		},
	})
}

func TestUnaryExpressionEvaluate(t *testing.T) {
	runEvaluateTestCases(t, []evaluateTestCase{
		{"negation of an integer", negation(integerLiteral(3)), types.NewInteger(-3)},
		{"negation of a float", negation(floatLiteral(2.5)), types.NewFloat(-2.5)},
		{"double negation", negation(negation(integerLiteral(3))), types.NewInteger(3)},
		{"triple negation", negation(negation(negation(integerLiteral(3)))), types.NewInteger(-3)},
		{
			"negation of a grouping",
			negation(grouping(binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2)))),
			types.NewInteger(-3),
		},
	})
}

func TestBinaryExpressionEvaluate(t *testing.T) {
	runEvaluateTestCases(t, []evaluateTestCase{
		{
			"addition of integers",
			binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2)),
			types.NewInteger(3),
		},
		{
			"addition of an integer and a float",
			binary(integerLiteral(1), common.PLUS, "+", floatLiteral(2.5)),
			types.NewFloat(3.5),
		},
		{
			"subtraction",
			binary(integerLiteral(8), common.MINUS, "-", integerLiteral(5)),
			types.NewInteger(3),
		},
		{
			"multiplication",
			binary(integerLiteral(3), common.STAR, "*", integerLiteral(5)),
			types.NewInteger(15),
		},
		{
			"division always returns a float",
			binary(integerLiteral(10), common.SLASH, "/", integerLiteral(2)),
			types.NewFloat(5),
		},
		{
			"floor division of integers returns an integer",
			binary(integerLiteral(10), common.DOUBLE_SLASH, "//", integerLiteral(3)),
			types.NewInteger(3),
		},
		{
			"floor division with float operands returns a float",
			binary(floatLiteral(10.9), common.DOUBLE_SLASH, "//", floatLiteral(3.9)),
			types.NewFloat(2),
		},
		{
			"modulo",
			binary(integerLiteral(7), common.PERCENTAGE, "%", integerLiteral(4)),
			types.NewInteger(3),
		},
		{
			"exponentiation of integers returns an integer",
			binary(integerLiteral(2), common.DOUBLE_STAR, "**", integerLiteral(3)),
			types.NewInteger(8),
		},
		{
			"exponentiation with a float operand returns a float",
			binary(floatLiteral(2.5), common.DOUBLE_STAR, "**", integerLiteral(2)),
			types.NewFloat(6.25),
		},
		{
			"nested expressions",
			binary(
				grouping(binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2))),
				common.STAR, "*",
				integerLiteral(3),
			),
			types.NewInteger(9),
		},
		{
			"nested expression with a negation",
			binary(
				negation(grouping(binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2)))),
				common.STAR, "*",
				floatLiteral(2.5),
			),
			types.NewFloat(-7.5),
		},
	})
}

func TestBinaryExpressionInvalidOperator(t *testing.T) {
	runEvaluateErrorTestCases(t, []evaluateErrorTestCase{
		{
			"dot",
			binary(integerLiteral(1), common.DOT, ".", integerLiteral(2)),
			"Invalid binary operator: DOT<.>",
		},
		{
			"semicolon",
			binary(integerLiteral(1), common.SEMICOLON, ";", integerLiteral(2)),
			"Invalid binary operator: SEMICOLON<;>",
		},
		{
			"open parentheses",
			binary(integerLiteral(1), common.OPEN_PAR, "(", integerLiteral(2)),
			"Invalid binary operator: OPEN_PAR<(>",
		},
		{
			"EOF",
			binary(integerLiteral(1), common.EOF, "", integerLiteral(2)),
			"Invalid binary operator: EOF<>",
		},
	})
}

func TestEvaluationErrorPropagation(t *testing.T) {
	runEvaluateErrorTestCases(t, []evaluateErrorTestCase{
		{
			"division by zero",
			divisionByZero(),
			"Cannot divide by zero: 1 / 0",
		},
		{
			"modulo by zero",
			moduloByZero(),
			"Cannot divide by zero: 2 % 0",
		},
		{
			"inside a grouping",
			grouping(divisionByZero()),
			"Cannot divide by zero: 1 / 0",
		},
		{
			"inside a negation",
			negation(divisionByZero()),
			"Cannot divide by zero: 1 / 0",
		},
		{
			"in the left operand",
			binary(divisionByZero(), common.PLUS, "+", integerLiteral(2)),
			"Cannot divide by zero: 1 / 0",
		},
		{
			"in the right operand",
			binary(integerLiteral(2), common.PLUS, "+", divisionByZero()),
			"Cannot divide by zero: 1 / 0",
		},
		{
			"the left operand is evaluated first",
			binary(divisionByZero(), common.PLUS, "+", moduloByZero()),
			"Cannot divide by zero: 1 / 0",
		},
		{
			"the operands are evaluated before validating the operator",
			binary(divisionByZero(), common.DOT, ".", integerLiteral(2)),
			"Cannot divide by zero: 1 / 0",
		},
		{
			"deeply nested",
			negation(grouping(binary(
				integerLiteral(2),
				common.STAR, "*",
				grouping(negation(moduloByZero())),
			))),
			"Cannot divide by zero: 2 % 0",
		},
	})
}

func TestExpressionString(t *testing.T) {
	runExpressionStringTestCases(t, []expressionStringTestCase{
		{"integer literal", integerLiteral(3), "INTEGER<3>"},
		{"float literal", floatLiteral(2.5), "FLOAT<2.5>"},
		{"unary expression", negation(integerLiteral(3)), "(MINUS<-> INTEGER<3>)"},
		{
			"binary expression",
			binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2)),
			"(INTEGER<1> PLUS<+> INTEGER<2>)",
		},
		{"grouping expression", grouping(integerLiteral(3)), "(INTEGER<3>)"},
		{
			"nested expressions",
			binary(
				negation(grouping(binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2)))),
				common.STAR, "*",
				integerLiteral(3),
			),
			"((MINUS<-> ((INTEGER<1> PLUS<+> INTEGER<2>))) STAR<*> INTEGER<3>)",
		},
	})
}
