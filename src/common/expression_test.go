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

func variable(name string) *common.VariableExpression {
	return common.NewVariableExpression(token(common.IDENTIFIER, name))
}

func assignment(name string, valueExpression common.Expression) *common.VarAssignmentExpression {
	return common.NewVarAssignmentExpression(token(common.IDENTIFIER, name), valueExpression)
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
		{"boolean literal", booleanLiteral(true), "TRUE<True>"},
		{"unary expression", negation(integerLiteral(3)), "(MINUS<-> INTEGER<3>)"},
		{
			"binary expression",
			binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2)),
			"(INTEGER<1> PLUS<+> INTEGER<2>)",
		},
		{"grouping expression", grouping(integerLiteral(3)), "(INTEGER<3>)"},
		{
			"comparison expression",
			binary(integerLiteral(1), common.LESS_EQUAL, "<=", integerLiteral(2)),
			"(INTEGER<1> LESS_EQUAL<<=> INTEGER<2>)",
		},
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

func TestVariableExpressionString(t *testing.T) {
	runExpressionStringTestCases(t, []expressionStringTestCase{
		{"variable", variable("x"), "x"},
		{"variable with several characters", variable("my_var1"), "my_var1"},
		{"variable in a binary expression", binary(variable("x"), common.PLUS, "+", integerLiteral(1)), "(x PLUS<+> INTEGER<1>)"},
		{"negated variable", negation(variable("x")), "(MINUS<-> x)"},
		{"grouped variable", grouping(variable("x")), "(x)"},
	})
}

func TestVarAssignmentExpressionString(t *testing.T) {
	runExpressionStringTestCases(t, []expressionStringTestCase{
		{"assignment", assignment("x", integerLiteral(1)), "(x = INTEGER<1>)"},
		{"assignment of a variable", assignment("x", variable("y")), "(x = y)"},
		{
			"assignment of a binary expression",
			assignment("x", binary(variable("x"), common.STAR, "*", floatLiteral(2.5))),
			"(x = (x STAR<*> FLOAT<2.5>))",
		},
		{
			"chained assignment",
			assignment("a", assignment("b", assignment("c", stringLiteral("a")))),
			`(a = (b = (c = STRING<"a">)))`,
		},
		{
			"grouped assignment as an operand",
			binary(grouping(assignment("x", integerLiteral(1))), common.PLUS, "+", integerLiteral(2)),
			"(((x = INTEGER<1>)) PLUS<+> INTEGER<2>)",
		},
	})
}
