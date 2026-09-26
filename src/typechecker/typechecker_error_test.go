package typechecker_test

import (
	"strconv"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/typechecker"
)

type checkErrorTestCase struct {
	name            string
	statements      []common.Statement
	expectedMessage string
}

func token(tokenType common.TokenType, lexeme string) common.Token {
	return common.NewToken(tokenType, lexeme, common.Position{})
}

func operatorAt(tokenType common.TokenType, lexeme string, line int, column int) common.Token {
	return common.NewToken(tokenType, lexeme, common.Position{Line: line, Column: column})
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

func binary(leftExpression common.Expression, operator common.Token, rightExpression common.Expression) *common.BinaryExpression {
	return common.NewBinaryExpression(leftExpression, operator, rightExpression)
}

func unary(operator common.Token, expression common.Expression) *common.UnaryExpression {
	return common.NewUnaryExpression(operator, expression)
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

func assertCheckErrors(t *testing.T, inputStatements []common.Statement, expectedMessages []string) {
	t.Helper()

	errors := typechecker.NewTypeChecker(inputStatements).Check()

	if len(errors) != len(expectedMessages) {
		t.Fatalf("checking %v returned %d errors (%v); want %d",
			inputStatements, len(errors), errors, len(expectedMessages))
	}
	for i, expectedMessage := range expectedMessages {
		if errors[i].Error() != expectedMessage {
			t.Errorf("checking %v error %d = %q; want %q", inputStatements, i, errors[i], expectedMessage)
		}
	}
}

func runCheckErrorTestCases(t *testing.T, testCases []checkErrorTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertCheckErrors(t, testCase.statements, []string{testCase.expectedMessage})
		})
	}
}

func TestValidProgramsAreAccepted(t *testing.T) {
	testCases := []struct {
		name       string
		statements []common.Statement
	}{
		{"nil statements", nil},
		{"empty program", statements()},
		{
			"integer arithmetic",
			statements(binary(
				integerLiteral(1),
				token(common.PLUS, "+"),
				binary(integerLiteral(2), token(common.STAR, "*"), integerLiteral(3)),
			)),
		},
		{
			"mixed arithmetic",
			statements(binary(
				integerLiteral(1),
				token(common.PLUS, "+"),
				binary(
					floatLiteral(2.5),
					token(common.DOUBLE_SLASH, "//"),
					grouping(binary(integerLiteral(3), token(common.MINUS, "-"), integerLiteral(1))),
				),
			)),
		},
		{
			"concatenation",
			statements(binary(stringLiteral("a"), token(common.PLUS, "+"), stringLiteral("b"))),
		},
		{
			"several statements",
			statements(
				binary(integerLiteral(1), token(common.PLUS, "+"), integerLiteral(2)),
				binary(floatLiteral(2.5), token(common.STAR, "*"), integerLiteral(3)),
				binary(stringLiteral("a"), token(common.PLUS, "+"), stringLiteral("b")),
			),
		},
		{
			"division by zero",
			statements(binary(integerLiteral(1), token(common.SLASH, "/"), integerLiteral(0))),
		},
		{
			"negative exponent",
			statements(binary(
				integerLiteral(2),
				token(common.DOUBLE_STAR, "**"),
				unary(token(common.MINUS, "-"), integerLiteral(1)),
			)),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertCheckErrors(t, testCase.statements, nil)
		})
	}
}

func TestUnsupportedOperandTypes(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"subtraction of strings",
			statements(binary(stringLiteral("a"), operatorAt(common.MINUS, "-", 1, 5), stringLiteral("b"))),
			"[line 1, column 5] Unsupported operand types for -: String and String",
		},
		{
			"multiplication of strings",
			statements(binary(stringLiteral("a"), operatorAt(common.STAR, "*", 1, 5), stringLiteral("b"))),
			"[line 1, column 5] Unsupported operand types for *: String and String",
		},
		{
			"division of strings",
			statements(binary(stringLiteral("a"), operatorAt(common.SLASH, "/", 1, 5), stringLiteral("b"))),
			"[line 1, column 5] Unsupported operand types for /: String and String",
		},
		{
			"floor division of strings",
			statements(binary(stringLiteral("a"), operatorAt(common.DOUBLE_SLASH, "//", 1, 5), stringLiteral("b"))),
			"[line 1, column 5] Unsupported operand types for //: String and String",
		},
		{
			"modulo of strings",
			statements(binary(stringLiteral("a"), operatorAt(common.PERCENTAGE, "%", 1, 5), stringLiteral("b"))),
			"[line 1, column 5] Unsupported operand types for %: String and String",
		},
		{
			"power of strings",
			statements(binary(stringLiteral("a"), operatorAt(common.DOUBLE_STAR, "**", 1, 5), stringLiteral("b"))),
			"[line 1, column 5] Unsupported operand types for **: String and String",
		},
		{
			"integer plus string",
			statements(binary(integerLiteral(1), operatorAt(common.PLUS, "+", 1, 3), stringLiteral("a"))),
			"[line 1, column 3] Unsupported operand types for +: Int and String",
		},
		{
			"string plus integer",
			statements(binary(stringLiteral("a"), operatorAt(common.PLUS, "+", 1, 5), integerLiteral(1))),
			"[line 1, column 5] Unsupported operand types for +: String and Int",
		},
		{
			"float plus string",
			statements(binary(floatLiteral(2.5), operatorAt(common.PLUS, "+", 1, 5), stringLiteral("a"))),
			"[line 1, column 5] Unsupported operand types for +: Float and String",
		},
		{
			"grouped integer plus string",
			statements(binary(
				grouping(binary(integerLiteral(1), operatorAt(common.PLUS, "+", 1, 4), integerLiteral(2))),
				operatorAt(common.PLUS, "+", 1, 9),
				stringLiteral("a"),
			)),
			"[line 1, column 9] Unsupported operand types for +: Int and String",
		},
		{
			"concatenation result used in a subtraction",
			statements(binary(
				grouping(binary(stringLiteral("a"), operatorAt(common.PLUS, "+", 1, 6), stringLiteral("b"))),
				operatorAt(common.MINUS, "-", 1, 13),
				integerLiteral(1),
			)),
			"[line 1, column 13] Unsupported operand types for -: String and Int",
		},
		{
			"negation of a string",
			statements(unary(operatorAt(common.MINUS, "-", 1, 1), stringLiteral("a"))),
			"[line 1, column 1] Unsupported operand type for -: String",
		},
		{
			"negation of a concatenation",
			statements(unary(
				operatorAt(common.MINUS, "-", 1, 1),
				grouping(binary(stringLiteral("a"), operatorAt(common.PLUS, "+", 1, 7), stringLiteral("b"))),
			)),
			"[line 1, column 1] Unsupported operand type for -: String",
		},
	})
}

func TestErrorsAreAccumulated(t *testing.T) {
	assertCheckErrors(t, statements(
		binary(stringLiteral("a"), operatorAt(common.MINUS, "-", 1, 5), stringLiteral("b")),
		binary(integerLiteral(1), operatorAt(common.PLUS, "+", 2, 3), stringLiteral("c")),
		unary(operatorAt(common.MINUS, "-", 3, 1), stringLiteral("d")),
	), []string{
		"[line 1, column 5] Unsupported operand types for -: String and String",
		"[line 2, column 3] Unsupported operand types for +: Int and String",
		"[line 3, column 1] Unsupported operand type for -: String",
	})
}

func TestOneErrorDoesNotCascade(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"the enclosing binary expression stays silent",
			statements(binary(
				grouping(binary(stringLiteral("a"), operatorAt(common.MINUS, "-", 1, 6), stringLiteral("b"))),
				operatorAt(common.PLUS, "+", 1, 13),
				integerLiteral(1),
			)),
			"[line 1, column 6] Unsupported operand types for -: String and String",
		},
		{
			"the enclosing negation stays silent",
			statements(unary(
				operatorAt(common.MINUS, "-", 1, 1),
				grouping(binary(stringLiteral("a"), operatorAt(common.MINUS, "-", 1, 7), stringLiteral("b"))),
			)),
			"[line 1, column 7] Unsupported operand types for -: String and String",
		},
		{
			"an invalid operand poisons the whole chain",
			statements(binary(
				binary(
					grouping(unary(operatorAt(common.MINUS, "-", 1, 2), stringLiteral("a"))),
					operatorAt(common.STAR, "*", 1, 8),
					integerLiteral(2),
				),
				operatorAt(common.PLUS, "+", 1, 12),
				integerLiteral(1),
			)),
			"[line 1, column 2] Unsupported operand type for -: String",
		},
	})
}

func TestInvalidOperator(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"binary dot",
			statements(binary(integerLiteral(1), operatorAt(common.DOT, ".", 1, 3), integerLiteral(2))),
			"[line 1, column 3] Invalid binary operator: DOT<.>",
		},
		{
			"binary semicolon",
			statements(binary(integerLiteral(1), operatorAt(common.SEMICOLON, ";", 1, 3), integerLiteral(2))),
			"[line 1, column 3] Invalid binary operator: SEMICOLON<;>",
		},
		{
			"binary open parentheses",
			statements(binary(integerLiteral(1), operatorAt(common.OPEN_PAR, "(", 1, 3), integerLiteral(2))),
			"[line 1, column 3] Invalid binary operator: OPEN_PAR<(>",
		},
		{
			"binary EOF",
			statements(binary(integerLiteral(1), operatorAt(common.EOF, "", 1, 3), integerLiteral(2))),
			"[line 1, column 3] Invalid binary operator: EOF<>",
		},
		{
			"unary plus",
			statements(unary(operatorAt(common.PLUS, "+", 1, 1), integerLiteral(1))),
			"[line 1, column 1] Invalid unary operator: PLUS<+>",
		},
		{
			"unary star",
			statements(unary(operatorAt(common.STAR, "*", 1, 1), integerLiteral(1))),
			"[line 1, column 1] Invalid unary operator: STAR<*>",
		},
	})
}

func TestValidPrintStatementsAreAccepted(t *testing.T) {
	testCases := []struct {
		name       string
		statements []common.Statement
	}{
		{"string literal", []common.Statement{common.NewPrintStatement(stringLiteral("hola"))}},
		{
			"mixed arithmetic",
			[]common.Statement{common.NewPrintStatement(
				binary(integerLiteral(1), token(common.PLUS, "+"), floatLiteral(2.5)),
			)},
		},
		{
			"concatenation",
			[]common.Statement{common.NewPrintStatement(
				binary(stringLiteral("a"), token(common.PLUS, "+"), stringLiteral("b")),
			)},
		},
		{
			"division by zero",
			[]common.Statement{common.NewPrintStatement(
				binary(integerLiteral(1), token(common.SLASH, "/"), integerLiteral(0)),
			)},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertCheckErrors(t, testCase.statements, nil)
		})
	}
}

func TestPrintStatementUnsupportedOperandTypes(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"string minus integer",
			[]common.Statement{common.NewPrintStatement(
				binary(stringLiteral("a"), operatorAt(common.MINUS, "-", 1, 11), integerLiteral(1)),
			)},
			"[line 1, column 11] Unsupported operand types for -: String and Int",
		},
		{
			"negation of a string",
			[]common.Statement{common.NewPrintStatement(
				unary(operatorAt(common.MINUS, "-", 1, 7), stringLiteral("a")),
			)},
			"[line 1, column 7] Unsupported operand type for -: String",
		},
		{
			"error inside a grouping",
			[]common.Statement{common.NewPrintStatement(
				grouping(binary(integerLiteral(1), operatorAt(common.PLUS, "+", 1, 10), stringLiteral("a"))),
			)},
			"[line 1, column 10] Unsupported operand types for +: Int and String",
		},
	})
}

func TestErrorsAreAccumulatedAcrossPrintStatements(t *testing.T) {
	assertCheckErrors(t, []common.Statement{
		common.NewPrintStatement(binary(stringLiteral("a"), operatorAt(common.MINUS, "-", 1, 11), stringLiteral("b"))),
		common.NewExpressionStatement(binary(integerLiteral(1), operatorAt(common.PLUS, "+", 2, 3), stringLiteral("c"))),
		common.NewPrintStatement(stringLiteral("ok")),
		common.NewPrintStatement(unary(operatorAt(common.MINUS, "-", 4, 7), stringLiteral("d"))),
	}, []string{
		"[line 1, column 11] Unsupported operand types for -: String and String",
		"[line 2, column 3] Unsupported operand types for +: Int and String",
		"[line 4, column 7] Unsupported operand type for -: String",
	})
}

func booleanLiteral(value bool) *common.LiteralExpression {
	if value {
		return common.NewLiteralExpression(token(common.TRUE, "True"), types.NewBoolean(true))
	}
	return common.NewLiteralExpression(token(common.FALSE, "False"), types.NewBoolean(false))
}

func TestBooleanAndComparisonUnsupportedOperandTypes(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"equality between an integer and a string",
			statements(binary(integerLiteral(1), operatorAt(common.DOUBLE_EQUAL, "==", 1, 3), stringLiteral("a"))),
			"[line 1, column 3] Unsupported operand types for ==: Int and String",
		},
		{
			"inequality between a boolean and an integer",
			statements(binary(booleanLiteral(true), operatorAt(common.NOT_EQUAL, "!=", 1, 6), integerLiteral(1))),
			"[line 1, column 6] Unsupported operand types for !=: Bool and Int",
		},
		{
			"booleans have no order",
			statements(binary(booleanLiteral(true), operatorAt(common.LESS, "<", 1, 6), booleanLiteral(false))),
			"[line 1, column 6] Unsupported operand types for <: Bool and Bool",
		},
		{
			"order between a string and a number",
			statements(binary(stringLiteral("a"), operatorAt(common.GREATER_EQUAL, ">=", 1, 5), integerLiteral(1))),
			"[line 1, column 5] Unsupported operand types for >=: String and Int",
		},
		{
			"chained comparison compares a boolean against a number",
			statements(binary(
				binary(integerLiteral(1), operatorAt(common.LESS, "<", 1, 3), integerLiteral(2)),
				operatorAt(common.LESS, "<", 1, 7),
				integerLiteral(3),
			)),
			"[line 1, column 7] Unsupported operand types for <: Bool and Int",
		},
		{
			"addition of a boolean",
			statements(binary(booleanLiteral(true), operatorAt(common.PLUS, "+", 1, 6), integerLiteral(1))),
			"[line 1, column 6] Unsupported operand types for +: Bool and Int",
		},
		{
			"negation of a boolean",
			statements(unary(operatorAt(common.MINUS, "-", 1, 1), booleanLiteral(true))),
			"[line 1, column 1] Unsupported operand type for -: Bool",
		},
	})
}
