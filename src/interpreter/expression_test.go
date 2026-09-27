package interpreter

import (
	"io"
	"strconv"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type evaluateTestCase struct {
	name       string
	expression common.Expression
	expected   types.Value
}

type evaluateErrorTestCase struct {
	name            string
	expression      common.Expression
	expectedMessage string
}

func token(tokenType common.TokenType, lexeme string) common.Token {
	return common.NewToken(tokenType, lexeme, common.Position{})
}

func integerLiteral(value int64) *common.LiteralExpression {
	return common.NewLiteralExpression(token(common.INTEGER, strconv.FormatInt(value, 10)), types.NewInteger(value))
}

func floatLiteral(value float64) *common.LiteralExpression {
	return common.NewLiteralExpression(token(common.FLOAT, strconv.FormatFloat(value, 'f', -1, 64)), types.NewFloat(value))
}

func stringLiteral(value string) *common.LiteralExpression {
	return common.NewLiteralExpression(token(common.STRING, strconv.Quote(value)), types.NewString(value))
}

func binary(leftExpression common.Expression, tokenType common.TokenType, lexeme string, rightExpression common.Expression) *common.BinaryExpression {
	return common.NewBinaryExpression(leftExpression, token(tokenType, lexeme), rightExpression)
}

func booleanLiteral(value bool) *common.LiteralExpression {
	if value {
		return common.NewLiteralExpression(token(common.TRUE, "True"), types.NewBoolean(true))
	}
	return common.NewLiteralExpression(token(common.FALSE, "False"), types.NewBoolean(false))
}

func negation(expression common.Expression) *common.UnaryExpression {
	return common.NewUnaryExpression(token(common.MINUS, "-"), expression)
}

func logicalNot(expression common.Expression) *common.UnaryExpression {
	return common.NewUnaryExpression(token(common.NOT, "not"), expression)
}

func grouping(expression common.Expression) *common.GroupingExpression {
	return common.NewGroupingExpression(token(common.OPEN_PAR, "("), expression)
}

func divisionByZero() *common.BinaryExpression {
	return binary(integerLiteral(1), common.SLASH, "/", integerLiteral(0))
}

func moduloByZero() *common.BinaryExpression {
	return binary(integerLiteral(2), common.PERCENTAGE, "%", integerLiteral(0))
}

func assertEvaluate(t *testing.T, expression common.Expression, expected types.Value) {
	t.Helper()

	result, err := NewInterpreter(nil, io.Discard).evaluate(expression)

	if err != nil {
		t.Fatalf("evaluate(%s) unexpected error: %v", expression, err)
	}
	if result != expected {
		t.Errorf("evaluate(%s) = %#v; want %#v", expression, result, expected)
	}
}

func assertEvaluateError(t *testing.T, expression common.Expression, expectedMessage string) {
	t.Helper()

	_, err := NewInterpreter(nil, io.Discard).evaluate(expression)

	if err == nil {
		t.Fatalf("evaluate(%s) = nil error; want %q", expression, expectedMessage)
	}
	if err.Error() != expectedMessage {
		t.Errorf("evaluate(%s) error = %q; want %q", expression, err, expectedMessage)
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

func TestLiteralExpressionEvaluate(t *testing.T) {
	runEvaluateTestCases(t, []evaluateTestCase{
		{"integer", integerLiteral(3), types.NewInteger(3)},
		{"zero integer", integerLiteral(0), types.NewInteger(0)},
		{"float", floatLiteral(2.5), types.NewFloat(2.5)},
		{"zero float", floatLiteral(0), types.NewFloat(0)},
		{"string", stringLiteral("hola"), types.NewString("hola")},
		{"empty string", stringLiteral(""), types.NewString("")},
		{"true", booleanLiteral(true), types.NewBoolean(true)},
		{"false", booleanLiteral(false), types.NewBoolean(false)},
		{"grouped boolean", grouping(booleanLiteral(true)), types.NewBoolean(true)},
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
		{
			"concatenation of strings",
			binary(stringLiteral("Hola, "), common.PLUS, "+", stringLiteral("mundo")),
			types.NewString("Hola, mundo"),
		},
		{
			"concatenation with an empty string",
			binary(stringLiteral("hola"), common.PLUS, "+", stringLiteral("")),
			types.NewString("hola"),
		},
		{
			"chained concatenations",
			binary(
				binary(stringLiteral("a"), common.PLUS, "+", stringLiteral("b")),
				common.PLUS, "+",
				grouping(stringLiteral("c")),
			),
			types.NewString("abc"),
		},
	})
}

func TestEvaluationErrorPropagation(t *testing.T) {
	runEvaluateErrorTestCases(t, []evaluateErrorTestCase{
		{
			"division by zero",
			divisionByZero(),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"modulo by zero",
			moduloByZero(),
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
		{
			"inside a grouping",
			grouping(divisionByZero()),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"inside a negation",
			negation(divisionByZero()),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"in the left operand",
			binary(divisionByZero(), common.PLUS, "+", integerLiteral(2)),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"in the right operand",
			binary(integerLiteral(2), common.PLUS, "+", divisionByZero()),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"the left operand is evaluated first",
			binary(divisionByZero(), common.PLUS, "+", moduloByZero()),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"the operands are evaluated before validating the operator",
			binary(divisionByZero(), common.DOT, ".", integerLiteral(2)),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"negative exponent",
			binary(integerLiteral(2), common.DOUBLE_STAR, "**", integerLiteral(-1)),
			"[line 0, column 0] Cannot raise an integer to a negative power: 2 ** -1",
		},
		{
			"deeply nested",
			negation(grouping(binary(
				integerLiteral(2),
				common.STAR, "*",
				grouping(negation(moduloByZero())),
			))),
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
	})
}

func TestEqualityEvaluate(t *testing.T) {
	runEvaluateTestCases(t, []evaluateTestCase{
		{
			"equal integers",
			binary(integerLiteral(2), common.DOUBLE_EQUAL, "==", integerLiteral(2)),
			types.NewBoolean(true),
		},
		{
			"different integers",
			binary(integerLiteral(1), common.DOUBLE_EQUAL, "==", integerLiteral(2)),
			types.NewBoolean(false),
		},
		{
			"an integer and a float with the same value",
			binary(integerLiteral(1), common.DOUBLE_EQUAL, "==", floatLiteral(1)),
			types.NewBoolean(true),
		},
		{
			"a division against an integer",
			binary(
				binary(integerLiteral(10), common.SLASH, "/", integerLiteral(2)),
				common.DOUBLE_EQUAL, "==",
				integerLiteral(5),
			),
			types.NewBoolean(true),
		},
		{
			"the addition is evaluated before the equality",
			binary(
				binary(integerLiteral(5), common.PLUS, "+", integerLiteral(7)),
				common.DOUBLE_EQUAL, "==",
				integerLiteral(12),
			),
			types.NewBoolean(true),
		},
		{
			"equal strings",
			binary(stringLiteral("Hola"), common.DOUBLE_EQUAL, "==", stringLiteral("Hola")),
			types.NewBoolean(true),
		},
		{
			"strings with different case",
			binary(stringLiteral("hola"), common.DOUBLE_EQUAL, "==", stringLiteral("Hola")),
			types.NewBoolean(false),
		},
		{
			"a concatenation against a string",
			binary(
				binary(stringLiteral("Ho"), common.PLUS, "+", stringLiteral("la")),
				common.DOUBLE_EQUAL, "==",
				stringLiteral("Hola"),
			),
			types.NewBoolean(true),
		},
		{
			"equal booleans",
			binary(booleanLiteral(false), common.DOUBLE_EQUAL, "==", booleanLiteral(false)),
			types.NewBoolean(true),
		},
		{
			"different booleans",
			binary(booleanLiteral(false), common.DOUBLE_EQUAL, "==", booleanLiteral(true)),
			types.NewBoolean(false),
		},
		{
			"the result of an equality compared against a boolean",
			binary(
				grouping(binary(integerLiteral(1), common.DOUBLE_EQUAL, "==", integerLiteral(1))),
				common.DOUBLE_EQUAL, "==",
				booleanLiteral(true),
			),
			types.NewBoolean(true),
		},
	})
}

func TestComparisonEvaluate(t *testing.T) {
	runEvaluateTestCases(t, []evaluateTestCase{
		{
			"different integers are not equal",
			binary(integerLiteral(1), common.NOT_EQUAL, "!=", integerLiteral(2)),
			types.NewBoolean(true),
		},
		{
			"equal integers are not different",
			binary(integerLiteral(2), common.NOT_EQUAL, "!=", integerLiteral(2)),
			types.NewBoolean(false),
		},
		{
			"different booleans",
			binary(booleanLiteral(true), common.NOT_EQUAL, "!=", booleanLiteral(false)),
			types.NewBoolean(true),
		},
		{
			"different strings",
			binary(stringLiteral("a"), common.NOT_EQUAL, "!=", stringLiteral("b")),
			types.NewBoolean(true),
		},
		{
			"less between integers",
			binary(integerLiteral(1), common.LESS, "<", integerLiteral(2)),
			types.NewBoolean(true),
		},
		{
			"less is false for equal integers",
			binary(integerLiteral(2), common.LESS, "<", integerLiteral(2)),
			types.NewBoolean(false),
		},
		{
			"less equal is true for equal integers",
			binary(integerLiteral(2), common.LESS_EQUAL, "<=", integerLiteral(2)),
			types.NewBoolean(true),
		},
		{
			"greater between an integer and a float",
			binary(integerLiteral(3), common.GREATER, ">", floatLiteral(2.5)),
			types.NewBoolean(true),
		},
		{
			"greater equal between a float and an integer with the same value",
			binary(floatLiteral(2), common.GREATER_EQUAL, ">=", integerLiteral(2)),
			types.NewBoolean(true),
		},
		{
			"a negated number is smaller than zero",
			binary(negation(integerLiteral(1)), common.LESS, "<", integerLiteral(0)),
			types.NewBoolean(true),
		},
		{
			"the arithmetic is evaluated before the comparison",
			binary(
				binary(integerLiteral(2), common.STAR, "*", integerLiteral(3)),
				common.GREATER, ">",
				integerLiteral(5),
			),
			types.NewBoolean(true),
		},
		{
			"strings in alphabetical order",
			binary(stringLiteral("a"), common.LESS, "<", stringLiteral("b")),
			types.NewBoolean(true),
		},
		{
			"uppercase strings come first",
			binary(stringLiteral("Z"), common.LESS, "<", stringLiteral("a")),
			types.NewBoolean(true),
		},
		{
			"a prefix is smaller than the whole string",
			binary(stringLiteral("hol"), common.LESS_EQUAL, "<=", stringLiteral("hola")),
			types.NewBoolean(true),
		},
		{
			"the comparison is evaluated before the equality",
			binary(
				binary(integerLiteral(1), common.LESS, "<", integerLiteral(2)),
				common.DOUBLE_EQUAL, "==",
				booleanLiteral(true),
			),
			types.NewBoolean(true),
		},
	})
}

func TestUnsupportedOperandTypesEvaluate(t *testing.T) {
	runEvaluateErrorTestCases(t, []evaluateErrorTestCase{
		{
			"string plus integer",
			binary(stringLiteral("a"), common.PLUS, "+", integerLiteral(1)),
			"[line 0, column 0] Unsupported operand types for +: String and Int",
		},
		{
			"float plus string",
			binary(floatLiteral(2.5), common.PLUS, "+", stringLiteral("a")),
			"[line 0, column 0] Unsupported operand types for +: Float and String",
		},
		{
			"subtraction of strings",
			binary(stringLiteral("a"), common.MINUS, "-", stringLiteral("b")),
			"[line 0, column 0] Unsupported operand types for -: String and String",
		},
		{
			"multiplication of a string by an integer",
			binary(stringLiteral("a"), common.STAR, "*", integerLiteral(3)),
			"[line 0, column 0] Unsupported operand types for *: String and Int",
		},
		{
			"boolean plus integer",
			binary(booleanLiteral(true), common.PLUS, "+", integerLiteral(1)),
			"[line 0, column 0] Unsupported operand types for +: Bool and Int",
		},
		{
			"string plus boolean",
			binary(stringLiteral("a"), common.PLUS, "+", booleanLiteral(false)),
			"[line 0, column 0] Unsupported operand types for +: String and Bool",
		},
		{
			"addition of booleans",
			binary(booleanLiteral(true), common.PLUS, "+", booleanLiteral(false)),
			"[line 0, column 0] Unsupported operand types for +: Bool and Bool",
		},
		{
			"negation of a boolean",
			negation(booleanLiteral(true)),
			"[line 0, column 0] Unsupported operand type for -: Bool",
		},
		{
			"equality between an integer and a string",
			binary(integerLiteral(1), common.DOUBLE_EQUAL, "==", stringLiteral("a")),
			"[line 0, column 0] Unsupported operand types for ==: Int and String",
		},
		{
			"equality between a boolean and an integer",
			binary(booleanLiteral(true), common.DOUBLE_EQUAL, "==", integerLiteral(1)),
			"[line 0, column 0] Unsupported operand types for ==: Bool and Int",
		},
		{
			"equality between a string and a boolean",
			binary(stringLiteral("a"), common.DOUBLE_EQUAL, "==", booleanLiteral(false)),
			"[line 0, column 0] Unsupported operand types for ==: String and Bool",
		},
		{
			"chained equality of numbers compares a boolean against a number",
			binary(
				binary(integerLiteral(1), common.DOUBLE_EQUAL, "==", integerLiteral(1)),
				common.DOUBLE_EQUAL, "==",
				integerLiteral(1),
			),
			"[line 0, column 0] Unsupported operand types for ==: Bool and Int",
		},
		{
			"subtraction of booleans",
			binary(booleanLiteral(true), common.MINUS, "-", booleanLiteral(false)),
			"[line 0, column 0] Unsupported operand types for -: Bool and Bool",
		},
		{
			"booleans have no order",
			binary(booleanLiteral(true), common.LESS, "<", booleanLiteral(false)),
			"[line 0, column 0] Unsupported operand types for <: Bool and Bool",
		},
		{
			"greater equal between booleans",
			binary(booleanLiteral(true), common.GREATER_EQUAL, ">=", booleanLiteral(true)),
			"[line 0, column 0] Unsupported operand types for >=: Bool and Bool",
		},
		{
			"comparison between a number and a string",
			binary(integerLiteral(1), common.LESS, "<", stringLiteral("a")),
			"[line 0, column 0] Unsupported operand types for <: Int and String",
		},
		{
			"not equal between a string and a boolean",
			binary(stringLiteral("a"), common.NOT_EQUAL, "!=", booleanLiteral(true)),
			"[line 0, column 0] Unsupported operand types for !=: String and Bool",
		},
		{
			"chained comparison compares a boolean against a number",
			binary(
				binary(integerLiteral(1), common.LESS, "<", integerLiteral(2)),
				common.LESS, "<",
				integerLiteral(3),
			),
			"[line 0, column 0] Unsupported operand types for <: Bool and Int",
		},
		{
			"power of strings",
			binary(stringLiteral("a"), common.DOUBLE_STAR, "**", stringLiteral("b")),
			"[line 0, column 0] Unsupported operand types for **: String and String",
		},
		{
			"negation of a string",
			negation(stringLiteral("a")),
			"[line 0, column 0] Unsupported operand type for -: String",
		},
		{
			"concatenation result used in a subtraction",
			binary(
				grouping(binary(stringLiteral("a"), common.PLUS, "+", stringLiteral("b"))),
				common.MINUS, "-",
				integerLiteral(1),
			),
			"[line 0, column 0] Unsupported operand types for -: String and Int",
		},
	})
}

func TestLogicalEvaluate(t *testing.T) {
	runEvaluateTestCases(t, []evaluateTestCase{
		{"true and true", binary(booleanLiteral(true), common.AND, "and", booleanLiteral(true)), types.NewBoolean(true)},
		{"true and false", binary(booleanLiteral(true), common.AND, "and", booleanLiteral(false)), types.NewBoolean(false)},
		{"false and true", binary(booleanLiteral(false), common.AND, "and", booleanLiteral(true)), types.NewBoolean(false)},
		{"false and false", binary(booleanLiteral(false), common.AND, "and", booleanLiteral(false)), types.NewBoolean(false)},
		{"true or true", binary(booleanLiteral(true), common.OR, "or", booleanLiteral(true)), types.NewBoolean(true)},
		{"true or false", binary(booleanLiteral(true), common.OR, "or", booleanLiteral(false)), types.NewBoolean(true)},
		{"false or true", binary(booleanLiteral(false), common.OR, "or", booleanLiteral(true)), types.NewBoolean(true)},
		{"false or false", binary(booleanLiteral(false), common.OR, "or", booleanLiteral(false)), types.NewBoolean(false)},
		{"not true", logicalNot(booleanLiteral(true)), types.NewBoolean(false)},
		{"not false", logicalNot(booleanLiteral(false)), types.NewBoolean(true)},
		{"double not", logicalNot(logicalNot(booleanLiteral(true))), types.NewBoolean(true)},
		{
			"and is evaluated before or",
			binary(
				booleanLiteral(true),
				common.OR, "or",
				binary(booleanLiteral(false), common.AND, "and", booleanLiteral(false)),
			),
			types.NewBoolean(true),
		},
		{
			"not of a comparison",
			logicalNot(binary(integerLiteral(1), common.DOUBLE_EQUAL, "==", integerLiteral(2))),
			types.NewBoolean(true),
		},
		{
			"and of comparisons",
			binary(
				binary(integerLiteral(1), common.LESS, "<", integerLiteral(2)),
				common.AND, "and",
				binary(integerLiteral(3), common.LESS, "<", integerLiteral(4)),
			),
			types.NewBoolean(true),
		},
		{
			"not of a grouped and",
			logicalNot(grouping(binary(booleanLiteral(true), common.AND, "and", booleanLiteral(false)))),
			types.NewBoolean(true),
		},
	})
}

func TestShortCircuitEvaluate(t *testing.T) {
	failingComparison := binary(divisionByZero(), common.DOUBLE_EQUAL, "==", integerLiteral(1))
	runEvaluateTestCases(t, []evaluateTestCase{
		{
			"and does not evaluate the right side when the left side is false",
			binary(booleanLiteral(false), common.AND, "and", failingComparison),
			types.NewBoolean(false),
		},
		{
			"or does not evaluate the right side when the left side is true",
			binary(booleanLiteral(true), common.OR, "or", failingComparison),
			types.NewBoolean(true),
		},
		{
			"a short circuit skips a whole chain",
			binary(
				binary(booleanLiteral(false), common.AND, "and", failingComparison),
				common.AND, "and",
				failingComparison,
			),
			types.NewBoolean(false),
		},
		{
			"a short circuit inside an or",
			binary(
				binary(booleanLiteral(false), common.AND, "and", failingComparison),
				common.OR, "or",
				booleanLiteral(true),
			),
			types.NewBoolean(true),
		},
	})
}

func TestLogicalErrorPropagation(t *testing.T) {
	failingComparison := binary(divisionByZero(), common.DOUBLE_EQUAL, "==", integerLiteral(1))
	runEvaluateErrorTestCases(t, []evaluateErrorTestCase{
		{
			"and evaluates the right side when the left side is true",
			binary(booleanLiteral(true), common.AND, "and", failingComparison),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"or evaluates the right side when the left side is false",
			binary(booleanLiteral(false), common.OR, "or", failingComparison),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"in the left operand",
			binary(failingComparison, common.OR, "or", booleanLiteral(true)),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"inside a not",
			logicalNot(failingComparison),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
	})
}

func TestLogicalUnsupportedOperandTypesEvaluate(t *testing.T) {
	runEvaluateErrorTestCases(t, []evaluateErrorTestCase{
		{
			"and with an integer on the left",
			binary(integerLiteral(1), common.AND, "and", booleanLiteral(true)),
			"[line 0, column 0] Unsupported operand type for and: Int",
		},
		{
			"or with a string on the right",
			binary(booleanLiteral(false), common.OR, "or", stringLiteral("a")),
			"[line 0, column 0] Unsupported operand types for or: Bool and String",
		},
		{
			"not of an integer",
			logicalNot(integerLiteral(1)),
			"[line 0, column 0] Unsupported operand type for not: Int",
		},
		{
			"negation of a boolean",
			negation(booleanLiteral(true)),
			"[line 0, column 0] Unsupported operand type for -: Bool",
		},
	})
}
