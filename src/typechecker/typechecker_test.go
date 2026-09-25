package typechecker

import (
	"strconv"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type inferredTypeTestCase struct {
	name       string
	expression common.Expression
	expected   types.Type
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

func negation(expression common.Expression) *common.UnaryExpression {
	return common.NewUnaryExpression(token(common.MINUS, "-"), expression)
}

func grouping(expression common.Expression) *common.GroupingExpression {
	return common.NewGroupingExpression(token(common.OPEN_PAR, "("), expression)
}

func assertInferredType(t *testing.T, expression common.Expression, expected types.Type) {
	t.Helper()

	typeChecker := NewTypeChecker(nil)
	result := typeChecker.checkExpression(expression)

	if len(typeChecker.errors) > 0 {
		t.Fatalf("checking %s unexpected errors: %v", expression, typeChecker.errors)
	}
	if result != expected {
		t.Errorf("checking %s = %v; want %v", expression, result, expected)
	}
}

func runInferredTypeTestCases(t *testing.T, testCases []inferredTypeTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertInferredType(t, testCase.expression, testCase.expected)
		})
	}
}

func TestLiteralType(t *testing.T) {
	runInferredTypeTestCases(t, []inferredTypeTestCase{
		{"integer", integerLiteral(3), types.Int},
		{"float", floatLiteral(2.5), types.Float},
		{"string", stringLiteral("hola"), types.Str},
		{"empty string", stringLiteral(""), types.Str},
	})
}

func TestGroupingType(t *testing.T) {
	runInferredTypeTestCases(t, []inferredTypeTestCase{
		{"grouped integer", grouping(integerLiteral(3)), types.Int},
		{"nested groupings", grouping(grouping(grouping(floatLiteral(2.5)))), types.Float},
		{"grouped string", grouping(stringLiteral("hola")), types.Str},
	})
}

func TestUnaryType(t *testing.T) {
	runInferredTypeTestCases(t, []inferredTypeTestCase{
		{"negation preserves integer", negation(integerLiteral(3)), types.Int},
		{"negation preserves float", negation(floatLiteral(2.5)), types.Float},
		{"double negation", negation(negation(integerLiteral(3))), types.Int},
		{
			"negation of a grouping",
			negation(grouping(binary(integerLiteral(1), common.PLUS, "+", floatLiteral(2.5)))),
			types.Float,
		},
	})
}

func TestAdditionType(t *testing.T) {
	runInferredTypeTestCases(t, []inferredTypeTestCase{
		{
			"two integers stay integer",
			binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2)),
			types.Int,
		},
		{
			"an integer and a float widen to float",
			binary(integerLiteral(1), common.PLUS, "+", floatLiteral(2.5)),
			types.Float,
		},
		{
			"a float and an integer widen to float",
			binary(floatLiteral(2.5), common.PLUS, "+", integerLiteral(1)),
			types.Float,
		},
		{
			"two floats stay float",
			binary(floatLiteral(2.5), common.PLUS, "+", floatLiteral(1.5)),
			types.Float,
		},
		{
			"two strings concatenate",
			binary(stringLiteral("a"), common.PLUS, "+", stringLiteral("b")),
			types.Str,
		},
	})
}

func TestSubtractionAndMultiplicationType(t *testing.T) {
	runInferredTypeTestCases(t, []inferredTypeTestCase{
		{
			"subtraction of integers",
			binary(integerLiteral(8), common.MINUS, "-", integerLiteral(5)),
			types.Int,
		},
		{
			"subtraction with a float",
			binary(integerLiteral(8), common.MINUS, "-", floatLiteral(5.5)),
			types.Float,
		},
		{
			"multiplication of integers",
			binary(integerLiteral(3), common.STAR, "*", integerLiteral(5)),
			types.Int,
		},
		{
			"multiplication with a float",
			binary(floatLiteral(3.5), common.STAR, "*", integerLiteral(5)),
			types.Float,
		},
	})
}

func TestDivisionTypeIsAlwaysFloat(t *testing.T) {
	runInferredTypeTestCases(t, []inferredTypeTestCase{
		{
			"two integers",
			binary(integerLiteral(10), common.SLASH, "/", integerLiteral(2)),
			types.Float,
		},
		{
			"an integer and a float",
			binary(integerLiteral(10), common.SLASH, "/", floatLiteral(2.5)),
			types.Float,
		},
		{
			"two floats",
			binary(floatLiteral(10.5), common.SLASH, "/", floatLiteral(2.5)),
			types.Float,
		},
	})
}

func TestFloorDivisionAndModuloType(t *testing.T) {
	runInferredTypeTestCases(t, []inferredTypeTestCase{
		{
			"floor division of integers",
			binary(integerLiteral(10), common.DOUBLE_SLASH, "//", integerLiteral(3)),
			types.Int,
		},
		{
			"floor division with a float",
			binary(floatLiteral(10.9), common.DOUBLE_SLASH, "//", integerLiteral(3)),
			types.Float,
		},
		{
			"modulo of integers",
			binary(integerLiteral(7), common.PERCENTAGE, "%", integerLiteral(4)),
			types.Int,
		},
		{
			"modulo with a float",
			binary(integerLiteral(7), common.PERCENTAGE, "%", floatLiteral(4.5)),
			types.Float,
		},
	})
}

func TestPowerType(t *testing.T) {
	runInferredTypeTestCases(t, []inferredTypeTestCase{
		{
			"two integers stay integer",
			binary(integerLiteral(2), common.DOUBLE_STAR, "**", integerLiteral(3)),
			types.Int,
		},
		{
			"a float base widens to float",
			binary(floatLiteral(2.5), common.DOUBLE_STAR, "**", integerLiteral(3)),
			types.Float,
		},
		{
			"a float exponent widens to float",
			binary(integerLiteral(2), common.DOUBLE_STAR, "**", floatLiteral(3)),
			types.Float,
		},
		{
			"a negative integer exponent stays integer",
			binary(integerLiteral(2), common.DOUBLE_STAR, "**", negation(integerLiteral(1))),
			types.Int,
		},
	})
}

func TestNestedExpressionType(t *testing.T) {
	runInferredTypeTestCases(t, []inferredTypeTestCase{
		{
			"integer arithmetic",
			binary(
				grouping(binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2))),
				common.STAR, "*",
				integerLiteral(3),
			),
			types.Int,
		},
		{
			"a division makes the whole expression float",
			binary(
				grouping(binary(integerLiteral(1), common.SLASH, "/", integerLiteral(2))),
				common.PLUS, "+",
				integerLiteral(1),
			),
			types.Float,
		},
		{
			"a float deep inside propagates",
			binary(
				integerLiteral(1),
				common.PLUS, "+",
				grouping(binary(
					integerLiteral(2),
					common.STAR, "*",
					grouping(binary(integerLiteral(3), common.MINUS, "-", floatLiteral(0.5))),
				)),
			),
			types.Float,
		},
		{
			"chained concatenations",
			binary(
				binary(stringLiteral("a"), common.PLUS, "+", stringLiteral("b")),
				common.PLUS, "+",
				stringLiteral("c"),
			),
			types.Str,
		},
		{
			"negation of a division",
			negation(grouping(binary(integerLiteral(10), common.SLASH, "/", integerLiteral(2)))),
			types.Float,
		},
	})
}
