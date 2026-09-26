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
	expected   types.Value
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

func booleanLiteral(value bool) *common.LiteralExpression {
	if value {
		return common.NewLiteralExpression(token(common.TRUE, "True"), types.NewBoolean(true))
	}
	return common.NewLiteralExpression(token(common.FALSE, "False"), types.NewBoolean(false))
}

func binary(leftExpression common.Expression, tokenType common.TokenType, lexeme string, rightExpression common.Expression) *common.BinaryExpression {
	return common.NewBinaryExpression(leftExpression, token(tokenType, lexeme), rightExpression)
}

func negation(expression common.Expression) *common.UnaryExpression {
	return common.NewUnaryExpression(token(common.MINUS, "-"), expression)
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

func TestUnsupportedOperandTypes(t *testing.T) {
	runEvaluateErrorTestCases(t, []evaluateErrorTestCase{
		{
			"string plus integer",
			binary(stringLiteral("a"), common.PLUS, "+", integerLiteral(1)),
			"[line 0, column 0] Unsupported operand types for +: string and integer",
		},
		{
			"float plus string",
			binary(floatLiteral(2.5), common.PLUS, "+", stringLiteral("a")),
			"[line 0, column 0] Unsupported operand types for +: float and string",
		},
		{
			"subtraction of strings",
			binary(stringLiteral("a"), common.MINUS, "-", stringLiteral("b")),
			"[line 0, column 0] Unsupported operand types for -: string and string",
		},
		{
			"multiplication of a string by an integer",
			binary(stringLiteral("a"), common.STAR, "*", integerLiteral(3)),
			"[line 0, column 0] Unsupported operand types for *: string and integer",
		},
		{
			"boolean plus integer",
			binary(booleanLiteral(true), common.PLUS, "+", integerLiteral(1)),
			"[line 0, column 0] Unsupported operand types for +: boolean and integer",
		},
		{
			"string plus boolean",
			binary(stringLiteral("a"), common.PLUS, "+", booleanLiteral(false)),
			"[line 0, column 0] Unsupported operand types for +: string and boolean",
		},
		{
			"addition of booleans",
			binary(booleanLiteral(true), common.PLUS, "+", booleanLiteral(false)),
			"[line 0, column 0] Unsupported operand types for +: boolean and boolean",
		},
		{
			"negation of a boolean",
			negation(booleanLiteral(true)),
			"[line 0, column 0] Unsupported operand type for -: boolean",
		},
		{
			"equality between an integer and a string",
			binary(integerLiteral(1), common.DOUBLE_EQUAL, "==", stringLiteral("a")),
			"[line 0, column 0] Unsupported operand types for ==: integer and string",
		},
		{
			"equality between a boolean and an integer",
			binary(booleanLiteral(true), common.DOUBLE_EQUAL, "==", integerLiteral(1)),
			"[line 0, column 0] Unsupported operand types for ==: boolean and integer",
		},
		{
			"equality between a string and a boolean",
			binary(stringLiteral("a"), common.DOUBLE_EQUAL, "==", booleanLiteral(false)),
			"[line 0, column 0] Unsupported operand types for ==: string and boolean",
		},
		{
			"chained equality of numbers compares a boolean against a number",
			binary(
				binary(integerLiteral(1), common.DOUBLE_EQUAL, "==", integerLiteral(1)),
				common.DOUBLE_EQUAL, "==",
				integerLiteral(1),
			),
			"[line 0, column 0] Unsupported operand types for ==: boolean and integer",
		},
		{
			"subtraction of booleans",
			binary(booleanLiteral(true), common.MINUS, "-", booleanLiteral(false)),
			"[line 0, column 0] Unsupported operand types for -: boolean and boolean",
		},
		{
			"booleans have no order",
			binary(booleanLiteral(true), common.LESS, "<", booleanLiteral(false)),
			"[line 0, column 0] Unsupported operand types for <: boolean and boolean",
		},
		{
			"greater equal between booleans",
			binary(booleanLiteral(true), common.GREATER_EQUAL, ">=", booleanLiteral(true)),
			"[line 0, column 0] Unsupported operand types for >=: boolean and boolean",
		},
		{
			"comparison between a number and a string",
			binary(integerLiteral(1), common.LESS, "<", stringLiteral("a")),
			"[line 0, column 0] Unsupported operand types for <: integer and string",
		},
		{
			"not equal between a string and a boolean",
			binary(stringLiteral("a"), common.NOT_EQUAL, "!=", booleanLiteral(true)),
			"[line 0, column 0] Unsupported operand types for !=: string and boolean",
		},
		{
			"chained comparison compares a boolean against a number",
			binary(
				binary(integerLiteral(1), common.LESS, "<", integerLiteral(2)),
				common.LESS, "<",
				integerLiteral(3),
			),
			"[line 0, column 0] Unsupported operand types for <: boolean and integer",
		},
		{
			"power of strings",
			binary(stringLiteral("a"), common.DOUBLE_STAR, "**", stringLiteral("b")),
			"[line 0, column 0] Unsupported operand types for **: string and string",
		},
		{
			"negation of a string",
			negation(stringLiteral("a")),
			"[line 0, column 0] Unsupported operand type for -: string",
		},
		{
			"concatenation result used in a subtraction",
			binary(
				grouping(binary(stringLiteral("a"), common.PLUS, "+", stringLiteral("b"))),
				common.MINUS, "-",
				integerLiteral(1),
			),
			"[line 0, column 0] Unsupported operand types for -: string and integer",
		},
	})
}

func TestBinaryExpressionInvalidOperator(t *testing.T) {
	runEvaluateErrorTestCases(t, []evaluateErrorTestCase{
		{
			"dot",
			binary(integerLiteral(1), common.DOT, ".", integerLiteral(2)),
			"[line 0, column 0] Invalid binary operator: DOT<.>",
		},
		{
			"semicolon",
			binary(integerLiteral(1), common.SEMICOLON, ";", integerLiteral(2)),
			"[line 0, column 0] Invalid binary operator: SEMICOLON<;>",
		},
		{
			"open parentheses",
			binary(integerLiteral(1), common.OPEN_PAR, "(", integerLiteral(2)),
			"[line 0, column 0] Invalid binary operator: OPEN_PAR<(>",
		},
		{
			"EOF",
			binary(integerLiteral(1), common.EOF, "", integerLiteral(2)),
			"[line 0, column 0] Invalid binary operator: EOF<>",
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

func TestExpressionString(t *testing.T) {
	runExpressionStringTestCases(t, []expressionStringTestCase{
		{"integer literal", integerLiteral(3), "INTEGER<3>"},
		{"float literal", floatLiteral(2.5), "FLOAT<2.5>"},
		{"string literal", stringLiteral("hola"), `STRING<"hola">`},
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
