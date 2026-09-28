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

type checkValidTestCase struct {
	name       string
	statements []common.Statement
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

func booleanLiteral(value bool) *common.LiteralExpression {
	if value {
		return common.NewLiteralExpression(token(common.TRUE, "True"), types.NewBoolean(true))
	}
	return common.NewLiteralExpression(token(common.FALSE, "False"), types.NewBoolean(false))
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

	_, errors := typechecker.NewTypeChecker(inputStatements).Check()

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

func runCheckValidTestCases(t *testing.T, testCases []checkValidTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertCheckErrors(t, testCase.statements, nil)
		})
	}
}

func block(statements ...common.Statement) *common.BlockStatement {
	return common.NewBlockStatement(append([]common.Statement{}, statements...))
}

func ifStatement(condition common.Expression, ifBranch common.Statement, elseBranch common.Statement) *common.IfStatement {
	return common.NewIfStatement(token(common.IF, "if"), condition, ifBranch, elseBranch)
}

func ifStatementAt(line int, column int, condition common.Expression, ifBranch common.Statement, elseBranch common.Statement) *common.IfStatement {
	return common.NewIfStatement(operatorAt(common.IF, "if", line, column), condition, ifBranch, elseBranch)
}

func whileStatement(condition common.Expression, body common.Statement) *common.WhileStatement {
	return common.NewWhileStatement(token(common.WHILE, "while"), condition, body)
}

func whileStatementAt(line int, column int, condition common.Expression, body common.Statement) *common.WhileStatement {
	return common.NewWhileStatement(operatorAt(common.WHILE, "while", line, column), condition, body)
}

func breakStatement() *common.BreakStatement {
	return common.NewBreakStatement(token(common.BREAK, "break"))
}

func breakStatementAt(line int, column int) *common.BreakStatement {
	return common.NewBreakStatement(operatorAt(common.BREAK, "break", line, column))
}

func continueStatement() *common.ContinueStatement {
	return common.NewContinueStatement(token(common.CONTINUE, "continue"))
}

func continueStatementAt(line int, column int) *common.ContinueStatement {
	return common.NewContinueStatement(operatorAt(common.CONTINUE, "continue", line, column))
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
		{
			"logical operators",
			statements(binary(
				unary(token(common.NOT, "not"), booleanLiteral(false)),
				token(common.AND, "and"),
				binary(
					binary(integerLiteral(1), token(common.LESS, "<"), integerLiteral(2)),
					token(common.OR, "or"),
					booleanLiteral(false),
				),
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

func TestLogicalUnsupportedOperandTypes(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"and with an integer on the left",
			statements(binary(integerLiteral(1), operatorAt(common.AND, "and", 1, 3), booleanLiteral(true))),
			"[line 1, column 3] Unsupported operand types for and: Int and Bool",
		},
		{
			"or with a string on the right",
			statements(binary(booleanLiteral(true), operatorAt(common.OR, "or", 1, 6), stringLiteral("a"))),
			"[line 1, column 6] Unsupported operand types for or: Bool and String",
		},
		{
			"and of numbers",
			statements(binary(integerLiteral(1), operatorAt(common.AND, "and", 1, 3), floatLiteral(2.5))),
			"[line 1, column 3] Unsupported operand types for and: Int and Float",
		},
		{
			"or of a comparison and a number",
			statements(binary(
				binary(integerLiteral(1), operatorAt(common.LESS, "<", 1, 3), integerLiteral(2)),
				operatorAt(common.OR, "or", 1, 7),
				integerLiteral(3),
			)),
			"[line 1, column 7] Unsupported operand types for or: Bool and Int",
		},
		{
			"not of an integer",
			statements(unary(operatorAt(common.NOT, "not", 1, 1), integerLiteral(1))),
			"[line 1, column 1] Unsupported operand type for not: Int",
		},
		{
			"not of a string",
			statements(unary(operatorAt(common.NOT, "not", 1, 1), stringLiteral("a"))),
			"[line 1, column 1] Unsupported operand type for not: String",
		},
		{
			"an error inside a not does not cascade",
			statements(unary(
				operatorAt(common.NOT, "not", 1, 1),
				grouping(binary(integerLiteral(1), operatorAt(common.AND, "and", 1, 8), booleanLiteral(true))),
			)),
			"[line 1, column 8] Unsupported operand types for and: Int and Bool",
		},
	})
}

func TestValidIfAndBlockStatementsAreAccepted(t *testing.T) {
	runCheckValidTestCases(t, []checkValidTestCase{
		{"empty block", []common.Statement{block()}},
		{
			"block with statements",
			[]common.Statement{block(
				common.NewPrintStatement(stringLiteral("a")),
				common.NewExpressionStatement(binary(integerLiteral(1), token(common.PLUS, "+"), floatLiteral(2.5))),
			)},
		},
		{"nested blocks", []common.Statement{block(block(common.NewPrintStatement(integerLiteral(1))))}},
		{"if with a boolean literal", []common.Statement{ifStatement(booleanLiteral(true), block(), nil)}},
		{
			"if with a comparison",
			[]common.Statement{ifStatement(
				binary(integerLiteral(1), token(common.LESS, "<"), floatLiteral(2.5)),
				block(common.NewPrintStatement(integerLiteral(1))),
				nil,
			)},
		},
		{
			"if with logical operators",
			[]common.Statement{ifStatement(
				binary(
					unary(token(common.NOT, "not"), booleanLiteral(false)),
					token(common.AND, "and"),
					binary(stringLiteral("a"), token(common.DOUBLE_EQUAL, "=="), stringLiteral("b")),
				),
				block(),
				nil,
			)},
		},
		{"if with a grouped condition", []common.Statement{ifStatement(grouping(booleanLiteral(true)), block(), nil)}},
		{
			"if with else",
			[]common.Statement{ifStatement(
				booleanLiteral(false),
				block(common.NewPrintStatement(integerLiteral(1))),
				block(common.NewPrintStatement(integerLiteral(2))),
			)},
		},
		{
			"else if chain",
			[]common.Statement{ifStatement(
				booleanLiteral(false),
				block(),
				ifStatement(booleanLiteral(true), block(), block()),
			)},
		},
		{
			"nested if",
			[]common.Statement{ifStatement(
				booleanLiteral(true),
				block(ifStatement(booleanLiteral(false), block(), block())),
				nil,
			)},
		},
	})
}

func TestIfStatementNonBooleanCondition(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"integer",
			[]common.Statement{ifStatementAt(1, 1, integerLiteral(1), block(), nil)},
			"[line 1, column 1] Non boolean expression in if condition: Int",
		},
		{
			"float",
			[]common.Statement{ifStatementAt(1, 1, floatLiteral(2.5), block(), nil)},
			"[line 1, column 1] Non boolean expression in if condition: Float",
		},
		{
			"string",
			[]common.Statement{ifStatementAt(1, 1, stringLiteral("a"), block(), nil)},
			"[line 1, column 1] Non boolean expression in if condition: String",
		},
		{
			"arithmetic expression",
			[]common.Statement{ifStatementAt(1, 1, binary(integerLiteral(1), token(common.PLUS, "+"), integerLiteral(2)), block(), nil)},
			"[line 1, column 1] Non boolean expression in if condition: Int",
		},
		{
			"grouped integer",
			[]common.Statement{ifStatementAt(1, 1, grouping(integerLiteral(1)), block(), nil)},
			"[line 1, column 1] Non boolean expression in if condition: Int",
		},
		{
			"points to the if token",
			[]common.Statement{ifStatementAt(3, 5, integerLiteral(1), block(), nil)},
			"[line 3, column 5] Non boolean expression in if condition: Int",
		},
		{
			"points to the if of the else if",
			[]common.Statement{ifStatementAt(
				1, 1,
				booleanLiteral(false),
				block(),
				ifStatementAt(1, 19, integerLiteral(1), block(), nil),
			)},
			"[line 1, column 19] Non boolean expression in if condition: Int",
		},
		{
			"nested if",
			[]common.Statement{ifStatementAt(
				1, 1,
				booleanLiteral(true),
				block(ifStatementAt(2, 3, stringLiteral("a"), block(), nil)),
				nil,
			)},
			"[line 2, column 3] Non boolean expression in if condition: String",
		},
	})
}

func TestIfStatementInvalidConditionDoesNotCascade(t *testing.T) {
	assertCheckErrors(t, []common.Statement{ifStatementAt(
		1, 1,
		binary(integerLiteral(1), operatorAt(common.PLUS, "+", 1, 7), stringLiteral("a")),
		block(),
		nil,
	)}, []string{
		"[line 1, column 7] Unsupported operand types for +: Int and String",
	})
}

func TestIfAndBlockStatementBranchErrors(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"error inside a block",
			[]common.Statement{block(common.NewPrintStatement(
				binary(stringLiteral("a"), operatorAt(common.MINUS, "-", 2, 13), integerLiteral(1)),
			))},
			"[line 2, column 13] Unsupported operand types for -: String and Int",
		},
		{
			"error inside a nested block",
			[]common.Statement{block(block(common.NewExpressionStatement(
				unary(operatorAt(common.MINUS, "-", 3, 5), stringLiteral("a")),
			)))},
			"[line 3, column 5] Unsupported operand type for -: String",
		},
		{
			"error in the if branch",
			[]common.Statement{ifStatement(
				booleanLiteral(true),
				block(common.NewPrintStatement(binary(integerLiteral(1), operatorAt(common.PLUS, "+", 2, 13), stringLiteral("a")))),
				nil,
			)},
			"[line 2, column 13] Unsupported operand types for +: Int and String",
		},
		{
			"error in the else branch",
			[]common.Statement{ifStatement(
				booleanLiteral(true),
				block(),
				block(common.NewPrintStatement(binary(integerLiteral(1), operatorAt(common.PLUS, "+", 4, 13), stringLiteral("a")))),
			)},
			"[line 4, column 13] Unsupported operand types for +: Int and String",
		},
		{
			"error in a branch that is never executed",
			[]common.Statement{ifStatement(
				booleanLiteral(false),
				block(common.NewPrintStatement(unary(operatorAt(common.MINUS, "-", 2, 11), stringLiteral("a")))),
				nil,
			)},
			"[line 2, column 11] Unsupported operand type for -: String",
		},
	})
}

func TestIfStatementErrorsAreAccumulated(t *testing.T) {
	assertCheckErrors(t, []common.Statement{
		ifStatementAt(
			1, 1,
			integerLiteral(1),
			block(common.NewPrintStatement(binary(stringLiteral("a"), operatorAt(common.MINUS, "-", 2, 13), stringLiteral("b")))),
			ifStatementAt(
				3, 8,
				stringLiteral("c"),
				block(),
				block(common.NewExpressionStatement(unary(operatorAt(common.MINUS, "-", 6, 3), stringLiteral("d")))),
			),
		),
		block(common.NewPrintStatement(binary(integerLiteral(1), operatorAt(common.PLUS, "+", 9, 11), stringLiteral("e")))),
	}, []string{
		"[line 1, column 1] Non boolean expression in if condition: Int",
		"[line 2, column 13] Unsupported operand types for -: String and String",
		"[line 3, column 8] Non boolean expression in if condition: String",
		"[line 6, column 3] Unsupported operand type for -: String",
		"[line 9, column 11] Unsupported operand types for +: Int and String",
	})
}

func TestValidWhileStatementsAreAccepted(t *testing.T) {
	runCheckValidTestCases(t, []checkValidTestCase{
		{"while with a boolean literal", []common.Statement{whileStatement(booleanLiteral(false), block())}},
		{
			"while with a comparison",
			[]common.Statement{whileStatement(
				binary(integerLiteral(1), token(common.LESS, "<"), floatLiteral(2.5)),
				block(common.NewPrintStatement(integerLiteral(1))),
			)},
		},
		{
			"while with logical operators",
			[]common.Statement{whileStatement(
				binary(
					unary(token(common.NOT, "not"), booleanLiteral(false)),
					token(common.OR, "or"),
					binary(stringLiteral("a"), token(common.NOT_EQUAL, "!="), stringLiteral("b")),
				),
				block(),
			)},
		},
		{"while with a grouped condition", []common.Statement{whileStatement(grouping(booleanLiteral(true)), block(breakStatement()))}},
		{"break in the body", []common.Statement{whileStatement(booleanLiteral(true), block(breakStatement()))}},
		{"continue in the body", []common.Statement{whileStatement(booleanLiteral(true), block(continueStatement()))}},
		{
			"break and continue inside an if in the body",
			[]common.Statement{whileStatement(
				booleanLiteral(true),
				block(ifStatement(booleanLiteral(false), block(continueStatement()), block(breakStatement()))),
			)},
		},
		{
			"break inside a nested block in the body",
			[]common.Statement{whileStatement(booleanLiteral(true), block(block(block(breakStatement()))))},
		},
		{
			"break after a nested while",
			[]common.Statement{whileStatement(
				booleanLiteral(true),
				block(whileStatement(booleanLiteral(true), block(continueStatement())), breakStatement()),
			)},
		},
		{
			"while inside an if",
			[]common.Statement{ifStatement(booleanLiteral(true), block(whileStatement(booleanLiteral(true), block(breakStatement()))), nil)},
		},
	})
}

func TestWhileStatementNonBooleanCondition(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"integer",
			[]common.Statement{whileStatementAt(1, 1, integerLiteral(1), block())},
			"[line 1, column 1] Non boolean expression in while condition: Int",
		},
		{
			"float",
			[]common.Statement{whileStatementAt(1, 1, floatLiteral(2.5), block())},
			"[line 1, column 1] Non boolean expression in while condition: Float",
		},
		{
			"string",
			[]common.Statement{whileStatementAt(1, 1, stringLiteral("a"), block())},
			"[line 1, column 1] Non boolean expression in while condition: String",
		},
		{
			"arithmetic expression",
			[]common.Statement{whileStatementAt(1, 1, binary(integerLiteral(1), token(common.PLUS, "+"), integerLiteral(2)), block())},
			"[line 1, column 1] Non boolean expression in while condition: Int",
		},
		{
			"grouped integer",
			[]common.Statement{whileStatementAt(1, 1, grouping(integerLiteral(1)), block())},
			"[line 1, column 1] Non boolean expression in while condition: Int",
		},
		{
			"points to the while token",
			[]common.Statement{whileStatementAt(3, 5, integerLiteral(1), block())},
			"[line 3, column 5] Non boolean expression in while condition: Int",
		},
		{
			"nested while",
			[]common.Statement{whileStatementAt(
				1, 1,
				booleanLiteral(true),
				block(whileStatementAt(2, 3, stringLiteral("a"), block())),
			)},
			"[line 2, column 3] Non boolean expression in while condition: String",
		},
	})
}

func TestWhileStatementInvalidConditionDoesNotCascade(t *testing.T) {
	assertCheckErrors(t, []common.Statement{whileStatementAt(
		1, 1,
		binary(integerLiteral(1), operatorAt(common.PLUS, "+", 1, 10), stringLiteral("a")),
		block(),
	)}, []string{
		"[line 1, column 10] Unsupported operand types for +: Int and String",
	})
}

func TestWhileStatementBodyErrors(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"error in the body",
			[]common.Statement{whileStatement(
				booleanLiteral(true),
				block(common.NewPrintStatement(binary(integerLiteral(1), operatorAt(common.PLUS, "+", 2, 13), stringLiteral("a")))),
			)},
			"[line 2, column 13] Unsupported operand types for +: Int and String",
		},
		{
			"error in a body that is never executed",
			[]common.Statement{whileStatement(
				booleanLiteral(false),
				block(common.NewPrintStatement(unary(operatorAt(common.MINUS, "-", 2, 11), stringLiteral("a")))),
			)},
			"[line 2, column 11] Unsupported operand type for -: String",
		},
		{
			"error in a nested while body",
			[]common.Statement{whileStatement(
				booleanLiteral(true),
				block(whileStatement(
					booleanLiteral(true),
					block(common.NewExpressionStatement(unary(operatorAt(common.MINUS, "-", 3, 5), stringLiteral("a")))),
				)),
			)},
			"[line 3, column 5] Unsupported operand type for -: String",
		},
	})
}

func TestBreakAndContinueOutsideLoop(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{"break at the top level", []common.Statement{breakStatementAt(1, 1)}, "[line 1, column 1] 'break' outside loop"},
		{"continue at the top level", []common.Statement{continueStatementAt(1, 1)}, "[line 1, column 1] 'continue' outside loop"},
		{"break inside a block", []common.Statement{block(breakStatementAt(2, 3))}, "[line 2, column 3] 'break' outside loop"},
		{
			"continue inside an if",
			[]common.Statement{ifStatement(booleanLiteral(true), block(continueStatementAt(2, 5)), nil)},
			"[line 2, column 5] 'continue' outside loop",
		},
		{
			"break inside an else",
			[]common.Statement{ifStatement(booleanLiteral(true), block(), block(breakStatementAt(4, 5)))},
			"[line 4, column 5] 'break' outside loop",
		},
		{
			"break after a while",
			[]common.Statement{
				whileStatement(booleanLiteral(true), block(breakStatement())),
				breakStatementAt(3, 1),
			},
			"[line 3, column 1] 'break' outside loop",
		},
		{
			"continue after a nested while",
			[]common.Statement{block(
				whileStatement(booleanLiteral(true), block(whileStatement(booleanLiteral(true), block(continueStatement())))),
				continueStatementAt(5, 3),
			)},
			"[line 5, column 3] 'continue' outside loop",
		},
	})
}

func TestWhileStatementErrorsAreAccumulated(t *testing.T) {
	assertCheckErrors(t, []common.Statement{
		breakStatementAt(1, 1),
		whileStatementAt(
			2, 1,
			integerLiteral(1),
			block(
				common.NewPrintStatement(binary(stringLiteral("a"), operatorAt(common.MINUS, "-", 3, 13), stringLiteral("b"))),
				whileStatementAt(4, 5, stringLiteral("c"), block(breakStatement())),
				continueStatement(),
			),
		),
		continueStatementAt(7, 1),
	}, []string{
		"[line 1, column 1] 'break' outside loop",
		"[line 2, column 1] Non boolean expression in while condition: Int",
		"[line 3, column 13] Unsupported operand types for -: String and String",
		"[line 4, column 5] Non boolean expression in while condition: String",
		"[line 7, column 1] 'continue' outside loop",
	})
}
