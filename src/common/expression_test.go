package common_test

import (
	"strconv"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

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

func binary(leftExpression common.Expression, tokenType common.TokenType, lexeme string, rightExpression common.Expression) *common.BinaryExpression {
	return common.NewBinaryExpression(leftExpression, token(tokenType, lexeme), rightExpression)
}

func negation(expression common.Expression) *common.UnaryExpression {
	return common.NewUnaryExpression(token(common.MINUS, "-"), expression)
}

func grouping(expression common.Expression) *common.GroupingExpression {
	return common.NewGroupingExpression(token(common.OPEN_PAR, "("), expression)
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
