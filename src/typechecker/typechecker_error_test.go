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

func varDeclaration(name string, varType types.Type, valueExpression common.Expression) *common.VarDeclarationStatement {
	return common.NewVarDeclarationStatement(token(common.LET, "let"), token(common.IDENTIFIER, name), varType, valueExpression)
}

func varDeclarationAt(
	line int,
	column int,
	name string,
	varType types.Type,
	valueExpression common.Expression,
) *common.VarDeclarationStatement {
	return common.NewVarDeclarationStatement(
		operatorAt(common.LET, "let", line, column),
		operatorAt(common.IDENTIFIER, name, line, column+4),
		varType,
		valueExpression,
	)
}

func variable(name string) *common.VariableExpression {
	return common.NewVariableExpression(token(common.IDENTIFIER, name))
}

func variableAt(line int, column int, name string) *common.VariableExpression {
	return common.NewVariableExpression(operatorAt(common.IDENTIFIER, name, line, column))
}

func assignment(name string, valueExpression common.Expression) *common.VarAssignmentExpression {
	return common.NewVarAssignmentExpression(token(common.IDENTIFIER, name), valueExpression)
}

func assignmentAt(line int, column int, name string, valueExpression common.Expression) *common.VarAssignmentExpression {
	return common.NewVarAssignmentExpression(operatorAt(common.IDENTIFIER, name, line, column), valueExpression)
}

func assertDistances(t *testing.T, inputStatements []common.Statement, expectedDistances map[common.Expression]int) {
	t.Helper()

	distances, errors := typechecker.NewTypeChecker(inputStatements).Check()

	if len(errors) > 0 {
		t.Fatalf("checking %v unexpected errors: %v", inputStatements, errors)
	}
	if len(distances) != len(expectedDistances) {
		t.Errorf("checking %v returned %d distances (%v); want %d", inputStatements, len(distances), distances, len(expectedDistances))
	}
	for expression, expectedDistance := range expectedDistances {
		distance, ok := distances[expression]
		if !ok {
			t.Errorf("checking %v returned no distance for %v; want %d", inputStatements, expression, expectedDistance)
		} else if distance != expectedDistance {
			t.Errorf("checking %v distance for %v = %d; want %d", inputStatements, expression, distance, expectedDistance)
		}
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

func TestValidVarDeclarationsAreAccepted(t *testing.T) {
	runCheckValidTestCases(t, []checkValidTestCase{
		{"int variable", []common.Statement{varDeclaration("x", types.Int, integerLiteral(1))}},
		{"float variable", []common.Statement{varDeclaration("x", types.Float, floatLiteral(2.5))}},
		{"string variable", []common.Statement{varDeclaration("x", types.Str, stringLiteral("a"))}},
		{"bool variable", []common.Statement{varDeclaration("x", types.Bool, booleanLiteral(true))}},
		{
			"float variable with a mixed arithmetic value",
			[]common.Statement{varDeclaration("x", types.Float, binary(integerLiteral(1), token(common.PLUS, "+"), floatLiteral(2.5)))},
		},
		{
			"bool variable with a comparison",
			[]common.Statement{varDeclaration("x", types.Bool, binary(integerLiteral(1), token(common.LESS, "<"), integerLiteral(2)))},
		},
		{
			"value with a previous variable",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				varDeclaration("y", types.Int, binary(variable("x"), token(common.STAR, "*"), integerLiteral(2))),
			},
		},
		{
			"value with an assignment",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				varDeclaration("y", types.Int, assignment("x", integerLiteral(2))),
			},
		},
		{
			"different variables with the same value",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				varDeclaration("y", types.Int, integerLiteral(1)),
			},
		},
		{
			"declaration inside a block",
			[]common.Statement{block(varDeclaration("x", types.Int, integerLiteral(1)))},
		},
		{
			"same name in sibling blocks",
			[]common.Statement{
				block(varDeclaration("x", types.Int, integerLiteral(1))),
				block(varDeclaration("x", types.Str, stringLiteral("a"))),
			},
		},
		{
			"declaration inside a while body",
			[]common.Statement{whileStatement(booleanLiteral(true), block(varDeclaration("x", types.Int, integerLiteral(1)), breakStatement()))},
		},
		{
			"same name in the if and else branches",
			[]common.Statement{ifStatement(
				booleanLiteral(true),
				block(varDeclaration("x", types.Int, integerLiteral(1))),
				block(varDeclaration("x", types.Bool, booleanLiteral(false))),
			)},
		},
	})
}

func TestVarDeclarationTypeMismatch(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"string to int",
			[]common.Statement{varDeclarationAt(1, 1, "x", types.Int, stringLiteral("a"))},
			"[line 1, column 1] Cannot assign String to variable of type Int",
		},
		{
			"float to int",
			[]common.Statement{varDeclarationAt(1, 1, "x", types.Int, floatLiteral(2.5))},
			"[line 1, column 1] Cannot assign Float to variable of type Int",
		},
		{
			"int is not promoted to float",
			[]common.Statement{varDeclarationAt(1, 1, "x", types.Float, integerLiteral(1))},
			"[line 1, column 1] Cannot assign Int to variable of type Float",
		},
		{
			"int to string",
			[]common.Statement{varDeclarationAt(1, 1, "x", types.Str, integerLiteral(1))},
			"[line 1, column 1] Cannot assign Int to variable of type String",
		},
		{
			"int to bool",
			[]common.Statement{varDeclarationAt(1, 1, "x", types.Bool, integerLiteral(1))},
			"[line 1, column 1] Cannot assign Int to variable of type Bool",
		},
		{
			"comparison to int",
			[]common.Statement{varDeclarationAt(1, 1, "x", types.Int, binary(integerLiteral(1), token(common.LESS, "<"), integerLiteral(2)))},
			"[line 1, column 1] Cannot assign Bool to variable of type Int",
		},
		{
			"mixed arithmetic to int",
			[]common.Statement{varDeclarationAt(1, 1, "x", types.Int, binary(integerLiteral(1), token(common.STAR, "*"), floatLiteral(2.5)))},
			"[line 1, column 1] Cannot assign Float to variable of type Int",
		},
		{
			"variable of another type",
			[]common.Statement{
				varDeclaration("x", types.Str, stringLiteral("a")),
				varDeclarationAt(2, 1, "y", types.Int, variable("x")),
			},
			"[line 2, column 1] Cannot assign String to variable of type Int",
		},
		{
			"points to the let token",
			[]common.Statement{block(varDeclarationAt(3, 5, "x", types.Bool, stringLiteral("a")))},
			"[line 3, column 5] Cannot assign String to variable of type Bool",
		},
	})
}

func TestVarRedeclaration(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"same type at the top level",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				varDeclarationAt(2, 1, "x", types.Int, integerLiteral(2)),
			},
			"[line 2, column 5] Variable 'x' already declared in this scope",
		},
		{
			"another type at the top level",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				varDeclarationAt(2, 1, "x", types.Str, stringLiteral("a")),
			},
			"[line 2, column 5] Variable 'x' already declared in this scope",
		},
		{
			"inside a block",
			[]common.Statement{block(
				varDeclaration("x", types.Int, integerLiteral(1)),
				varDeclarationAt(3, 3, "x", types.Int, integerLiteral(2)),
			)},
			"[line 3, column 7] Variable 'x' already declared in this scope",
		},
		{
			"after shadowing it in an inner block",
			[]common.Statement{block(
				varDeclaration("x", types.Int, integerLiteral(1)),
				block(varDeclaration("x", types.Int, integerLiteral(2))),
				varDeclarationAt(4, 3, "x", types.Int, integerLiteral(3)),
			)},
			"[line 4, column 7] Variable 'x' already declared in this scope",
		},
		{
			"inside a while body",
			[]common.Statement{whileStatement(booleanLiteral(true), block(
				varDeclaration("x", types.Int, integerLiteral(1)),
				varDeclarationAt(3, 3, "x", types.Int, integerLiteral(2)),
				breakStatement(),
			))},
			"[line 3, column 7] Variable 'x' already declared in this scope",
		},
		{
			"keeps the original type",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				varDeclarationAt(2, 1, "x", types.Str, stringLiteral("a")),
				common.NewExpressionStatement(binary(variable("x"), token(common.PLUS, "+"), integerLiteral(1))),
			},
			"[line 2, column 5] Variable 'x' already declared in this scope",
		},
	})
}

func TestValidVariablesAreAccepted(t *testing.T) {
	runCheckValidTestCases(t, []checkValidTestCase{
		{
			"variable as an operand",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				common.NewExpressionStatement(binary(variable("x"), token(common.PLUS, "+"), integerLiteral(2))),
			},
		},
		{
			"float variable in mixed arithmetic",
			[]common.Statement{
				varDeclaration("x", types.Float, floatLiteral(1.5)),
				common.NewExpressionStatement(binary(integerLiteral(2), token(common.DOUBLE_STAR, "**"), variable("x"))),
			},
		},
		{
			"string variable in a concatenation",
			[]common.Statement{
				varDeclaration("x", types.Str, stringLiteral("a")),
				common.NewExpressionStatement(binary(variable("x"), token(common.PLUS, "+"), variable("x"))),
			},
		},
		{
			"negated variable",
			[]common.Statement{
				varDeclaration("x", types.Bool, booleanLiteral(true)),
				common.NewExpressionStatement(unary(token(common.NOT, "not"), grouping(variable("x")))),
			},
		},
		{
			"variable in a print statement",
			[]common.Statement{varDeclaration("x", types.Str, stringLiteral("a")), common.NewPrintStatement(variable("x"))},
		},
		{
			"bool variable as an if condition",
			[]common.Statement{varDeclaration("x", types.Bool, booleanLiteral(true)), ifStatement(variable("x"), block(), nil)},
		},
		{
			"bool variable as a while condition",
			[]common.Statement{
				varDeclaration("x", types.Bool, booleanLiteral(true)),
				whileStatement(variable("x"), block(common.NewExpressionStatement(assignment("x", booleanLiteral(false))))),
			},
		},
		{
			"outer variable inside a block",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				block(block(common.NewPrintStatement(variable("x")))),
			},
		},
		{
			"shadowed variable has the inner type inside the block",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				block(
					varDeclaration("x", types.Str, stringLiteral("a")),
					common.NewExpressionStatement(binary(variable("x"), token(common.PLUS, "+"), stringLiteral("b"))),
				),
			},
		},
		{
			"shadowed variable has the outer type before the inner declaration",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				block(
					common.NewExpressionStatement(binary(variable("x"), token(common.PLUS, "+"), integerLiteral(1))),
					varDeclaration("x", types.Str, stringLiteral("a")),
				),
			},
		},
		{
			"shadowed variable has the outer type after the block",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				block(varDeclaration("x", types.Str, stringLiteral("a"))),
				common.NewExpressionStatement(binary(variable("x"), token(common.PLUS, "+"), integerLiteral(1))),
			},
		},
		{
			"inner variable initialized with the outer one",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				block(varDeclaration("x", types.Int, binary(variable("x"), token(common.PLUS, "+"), integerLiteral(1)))),
			},
		},
	})
}

func TestUndefinedVariable(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{"at the top level", statements(variableAt(1, 1, "x")), "[line 1, column 1] Undefined variable 'x'"},
		{
			"as an operand",
			statements(binary(integerLiteral(1), token(common.PLUS, "+"), variableAt(1, 5, "x"))),
			"[line 1, column 5] Undefined variable 'x'",
		},
		{"in a print statement", []common.Statement{common.NewPrintStatement(variableAt(1, 7, "x"))}, "[line 1, column 7] Undefined variable 'x'"},
		{"as an if condition", []common.Statement{ifStatement(variableAt(1, 5, "x"), block(), nil)}, "[line 1, column 5] Undefined variable 'x'"},
		{
			"as a while condition",
			[]common.Statement{whileStatement(variableAt(1, 8, "x"), block())},
			"[line 1, column 8] Undefined variable 'x'",
		},
		{
			"in the value of a declaration",
			[]common.Statement{varDeclaration("x", types.Int, variableAt(1, 14, "y"))},
			"[line 1, column 14] Undefined variable 'y'",
		},
		{
			"in its own declaration",
			[]common.Statement{varDeclaration("x", types.Int, binary(variableAt(1, 14, "x"), token(common.PLUS, "+"), integerLiteral(1)))},
			"[line 1, column 14] Undefined variable 'x'",
		},
		{
			"before its declaration",
			[]common.Statement{common.NewPrintStatement(variableAt(1, 7, "x")), varDeclaration("x", types.Int, integerLiteral(1))},
			"[line 1, column 7] Undefined variable 'x'",
		},
		{
			"another name",
			[]common.Statement{varDeclaration("x", types.Int, integerLiteral(1)), common.NewPrintStatement(variableAt(2, 7, "X"))},
			"[line 2, column 7] Undefined variable 'X'",
		},
		{
			"declared in a block and used after it",
			[]common.Statement{
				block(varDeclaration("x", types.Int, integerLiteral(1))),
				common.NewPrintStatement(variableAt(4, 7, "x")),
			},
			"[line 4, column 7] Undefined variable 'x'",
		},
		{
			"declared in a sibling block",
			[]common.Statement{
				block(varDeclaration("x", types.Int, integerLiteral(1))),
				block(common.NewPrintStatement(variableAt(5, 9, "x"))),
			},
			"[line 5, column 9] Undefined variable 'x'",
		},
		{
			"declared in the if branch and used in the else branch",
			[]common.Statement{ifStatement(
				booleanLiteral(true),
				block(varDeclaration("x", types.Int, integerLiteral(1))),
				block(common.NewPrintStatement(variableAt(4, 9, "x"))),
			)},
			"[line 4, column 9] Undefined variable 'x'",
		},
		{
			"declared in a while body and used after the loop",
			[]common.Statement{
				whileStatement(booleanLiteral(false), block(varDeclaration("x", types.Int, integerLiteral(1)))),
				common.NewPrintStatement(variableAt(4, 7, "x")),
			},
			"[line 4, column 7] Undefined variable 'x'",
		},
	})
}

func TestUndefinedVariableDoesNotCascade(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"the enclosing binary expression stays silent",
			statements(binary(variableAt(1, 1, "x"), token(common.PLUS, "+"), stringLiteral("a"))),
			"[line 1, column 1] Undefined variable 'x'",
		},
		{
			"the enclosing negation stays silent",
			statements(unary(token(common.MINUS, "-"), variableAt(1, 2, "x"))),
			"[line 1, column 2] Undefined variable 'x'",
		},
		{
			"the if condition stays silent",
			[]common.Statement{ifStatementAt(1, 1, variableAt(1, 5, "x"), block(), nil)},
			"[line 1, column 5] Undefined variable 'x'",
		},
		{
			"the declaration stays silent",
			[]common.Statement{varDeclarationAt(1, 1, "y", types.Int, variableAt(1, 14, "x"))},
			"[line 1, column 14] Undefined variable 'x'",
		},
	})
}

func TestInvalidVarDeclarationDoesNotCascade(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"invalid value does not report a type mismatch",
			[]common.Statement{varDeclarationAt(
				1, 1,
				"x", types.Int,
				binary(stringLiteral("a"), operatorAt(common.MINUS, "-", 1, 18), integerLiteral(1)),
			)},
			"[line 1, column 18] Unsupported operand types for -: String and Int",
		},
		{
			"variable with an invalid value is still declared",
			[]common.Statement{
				varDeclaration("x", types.Int, unary(operatorAt(common.MINUS, "-", 1, 14), stringLiteral("a"))),
				common.NewPrintStatement(binary(variable("x"), token(common.PLUS, "+"), integerLiteral(1))),
			},
			"[line 1, column 14] Unsupported operand type for -: String",
		},
		{
			"variable with a mismatched value keeps the declared type",
			[]common.Statement{
				varDeclarationAt(1, 1, "x", types.Int, stringLiteral("a")),
				common.NewPrintStatement(binary(variable("x"), token(common.PLUS, "+"), integerLiteral(1))),
			},
			"[line 1, column 1] Cannot assign String to variable of type Int",
		},
	})
}

func TestValidVarAssignmentsAreAccepted(t *testing.T) {
	runCheckValidTestCases(t, []checkValidTestCase{
		{
			"assignment of the same type",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				common.NewExpressionStatement(assignment("x", integerLiteral(2))),
			},
		},
		{
			"assignment of an expression with the variable",
			[]common.Statement{
				varDeclaration("x", types.Float, floatLiteral(1.5)),
				common.NewExpressionStatement(assignment("x", binary(variable("x"), token(common.STAR, "*"), integerLiteral(2)))),
			},
		},
		{
			"chained assignment",
			[]common.Statement{
				varDeclaration("a", types.Str, stringLiteral("a")),
				varDeclaration("b", types.Str, stringLiteral("b")),
				common.NewExpressionStatement(assignment("a", assignment("b", stringLiteral("c")))),
			},
		},
		{
			"assignment has the type of the variable",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				common.NewExpressionStatement(binary(grouping(assignment("x", integerLiteral(2))), token(common.PLUS, "+"), integerLiteral(3))),
			},
		},
		{
			"assignment as an if condition",
			[]common.Statement{
				varDeclaration("x", types.Bool, booleanLiteral(false)),
				ifStatement(assignment("x", booleanLiteral(true)), block(), nil),
			},
		},
		{
			"assignment in a print statement",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				common.NewPrintStatement(assignment("x", integerLiteral(2))),
			},
		},
		{
			"assignment to an outer variable inside a block",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				block(block(common.NewExpressionStatement(assignment("x", integerLiteral(2))))),
			},
		},
		{
			"assignment to a shadowed variable uses the inner type",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				block(
					varDeclaration("x", types.Str, stringLiteral("a")),
					common.NewExpressionStatement(assignment("x", stringLiteral("b"))),
				),
			},
		},
	})
}

func TestVarAssignmentErrors(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"undefined variable",
			statements(assignmentAt(1, 1, "x", integerLiteral(1))),
			"[line 1, column 1] Undefined variable 'x'",
		},
		{
			"string to int",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				common.NewExpressionStatement(assignmentAt(2, 1, "x", stringLiteral("a"))),
			},
			"[line 2, column 1] Cannot assign String to variable of type Int",
		},
		{
			"int is not promoted to float",
			[]common.Statement{
				varDeclaration("x", types.Float, floatLiteral(1.5)),
				common.NewExpressionStatement(assignmentAt(2, 1, "x", integerLiteral(1))),
			},
			"[line 2, column 1] Cannot assign Int to variable of type Float",
		},
		{
			"variable of another type",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				varDeclaration("y", types.Bool, booleanLiteral(true)),
				common.NewExpressionStatement(assignmentAt(3, 1, "x", variable("y"))),
			},
			"[line 3, column 1] Cannot assign Bool to variable of type Int",
		},
		{
			"chained assignment with different types",
			[]common.Statement{
				varDeclaration("a", types.Int, integerLiteral(1)),
				varDeclaration("b", types.Str, stringLiteral("b")),
				common.NewExpressionStatement(assignmentAt(3, 1, "a", assignmentAt(3, 5, "b", stringLiteral("c")))),
			},
			"[line 3, column 1] Cannot assign String to variable of type Int",
		},
		{
			"assignment has the type of the variable",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				common.NewExpressionStatement(binary(
					grouping(assignment("x", integerLiteral(2))),
					operatorAt(common.PLUS, "+", 2, 8),
					stringLiteral("a"),
				)),
			},
			"[line 2, column 8] Unsupported operand types for +: Int and String",
		},
		{
			"int assignment as an if condition",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				ifStatementAt(2, 1, assignment("x", integerLiteral(2)), block(), nil),
			},
			"[line 2, column 1] Non boolean expression in if condition: Int",
		},
		{
			"variable declared in a block and assigned after it",
			[]common.Statement{
				block(varDeclaration("x", types.Int, integerLiteral(1))),
				common.NewExpressionStatement(assignmentAt(4, 1, "x", integerLiteral(2))),
			},
			"[line 4, column 1] Undefined variable 'x'",
		},
		{
			"assignment with the outer type to a shadowed variable",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				block(
					varDeclaration("x", types.Str, stringLiteral("a")),
					common.NewExpressionStatement(assignmentAt(4, 3, "x", integerLiteral(2))),
				),
			},
			"[line 4, column 3] Cannot assign Int to variable of type String",
		},
		{
			"invalid value does not report a type mismatch",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				common.NewExpressionStatement(assignment("x", unary(operatorAt(common.MINUS, "-", 2, 5), stringLiteral("a")))),
			},
			"[line 2, column 5] Unsupported operand type for -: String",
		},
	})
}

func TestVariableErrorsAreAccumulated(t *testing.T) {
	assertCheckErrors(t, []common.Statement{
		varDeclarationAt(1, 1, "x", types.Int, stringLiteral("a")),
		varDeclarationAt(2, 1, "x", types.Int, integerLiteral(1)),
		common.NewPrintStatement(variableAt(3, 7, "y")),
		block(
			varDeclaration("z", types.Bool, booleanLiteral(true)),
			common.NewExpressionStatement(assignmentAt(5, 3, "z", integerLiteral(1))),
		),
		common.NewExpressionStatement(assignmentAt(7, 1, "z", assignmentAt(7, 5, "x", unary(operatorAt(common.MINUS, "-", 7, 9), stringLiteral("b"))))),
	}, []string{
		"[line 1, column 1] Cannot assign String to variable of type Int",
		"[line 2, column 5] Variable 'x' already declared in this scope",
		"[line 3, column 7] Undefined variable 'y'",
		"[line 5, column 3] Cannot assign Int to variable of type Bool",
		"[line 7, column 9] Unsupported operand type for -: String",
		"[line 7, column 1] Undefined variable 'z'",
	})
}

func TestVariableDistances(t *testing.T) {
	t.Run("no variables", func(t *testing.T) {
		assertDistances(t, statements(integerLiteral(1)), map[common.Expression]int{})
	})

	t.Run("variable in the same scope", func(t *testing.T) {
		use := variable("x")
		assertDistances(t, []common.Statement{
			varDeclaration("x", types.Int, integerLiteral(1)),
			common.NewPrintStatement(use),
		}, map[common.Expression]int{use: 0})
	})

	t.Run("assignment in the same scope", func(t *testing.T) {
		use := assignment("x", integerLiteral(2))
		assertDistances(t, []common.Statement{
			varDeclaration("x", types.Int, integerLiteral(1)),
			common.NewExpressionStatement(use),
		}, map[common.Expression]int{use: 0})
	})

	t.Run("declaration inside a block", func(t *testing.T) {
		use := variable("x")
		assertDistances(t, []common.Statement{block(
			varDeclaration("x", types.Int, integerLiteral(1)),
			common.NewPrintStatement(use),
		)}, map[common.Expression]int{use: 0})
	})

	t.Run("outer variable in nested blocks", func(t *testing.T) {
		inBlock := variable("x")
		inNestedBlock := assignment("x", integerLiteral(2))
		assertDistances(t, []common.Statement{
			varDeclaration("x", types.Int, integerLiteral(1)),
			block(
				common.NewPrintStatement(inBlock),
				block(common.NewExpressionStatement(inNestedBlock)),
			),
		}, map[common.Expression]int{inBlock: 1, inNestedBlock: 2})
	})

	t.Run("each use of the same variable has its own distance", func(t *testing.T) {
		first := variable("x")
		second := variable("x")
		assertDistances(t, []common.Statement{
			varDeclaration("x", types.Int, integerLiteral(1)),
			common.NewPrintStatement(first),
			block(common.NewPrintStatement(second)),
		}, map[common.Expression]int{first: 0, second: 1})
	})

	t.Run("shadowed variable", func(t *testing.T) {
		beforeShadowing := variable("x")
		afterShadowing := variable("x")
		afterBlock := variable("x")
		assertDistances(t, []common.Statement{
			varDeclaration("x", types.Int, integerLiteral(1)),
			block(
				common.NewPrintStatement(beforeShadowing),
				varDeclaration("x", types.Int, integerLiteral(2)),
				common.NewPrintStatement(afterShadowing),
			),
			common.NewPrintStatement(afterBlock),
		}, map[common.Expression]int{beforeShadowing: 1, afterShadowing: 0, afterBlock: 0})
	})

	t.Run("inner variable initialized with the outer one", func(t *testing.T) {
		outer := variable("x")
		assertDistances(t, []common.Statement{
			varDeclaration("x", types.Int, integerLiteral(1)),
			block(varDeclaration("x", types.Int, outer)),
		}, map[common.Expression]int{outer: 1})
	})

	t.Run("variables from different scopes", func(t *testing.T) {
		useX := variable("x")
		useY := variable("y")
		useZ := variable("z")
		assertDistances(t, []common.Statement{
			varDeclaration("x", types.Int, integerLiteral(1)),
			block(
				varDeclaration("y", types.Int, integerLiteral(2)),
				block(
					varDeclaration("z", types.Int, integerLiteral(3)),
					common.NewPrintStatement(binary(useX, token(common.PLUS, "+"), binary(useY, token(common.PLUS, "+"), useZ))),
				),
			),
		}, map[common.Expression]int{useX: 2, useY: 1, useZ: 0})
	})

	t.Run("chained assignment", func(t *testing.T) {
		inner := assignment("b", integerLiteral(1))
		outer := assignment("a", inner)
		assertDistances(t, []common.Statement{
			varDeclaration("a", types.Int, integerLiteral(1)),
			block(
				varDeclaration("b", types.Int, integerLiteral(2)),
				common.NewExpressionStatement(outer),
			),
		}, map[common.Expression]int{outer: 1, inner: 0})
	})

	t.Run("if and while conditions and bodies", func(t *testing.T) {
		ifCondition := variable("x")
		ifBranch := variable("x")
		elseBranch := variable("x")
		whileCondition := variable("x")
		whileBody := assignment("x", booleanLiteral(false))
		assertDistances(t, []common.Statement{
			varDeclaration("x", types.Bool, booleanLiteral(true)),
			ifStatement(ifCondition, block(common.NewPrintStatement(ifBranch)), block(block(common.NewPrintStatement(elseBranch)))),
			whileStatement(whileCondition, block(common.NewExpressionStatement(whileBody))),
		}, map[common.Expression]int{ifCondition: 0, ifBranch: 1, elseBranch: 2, whileCondition: 0, whileBody: 1})
	})
}

func parameter(name string, paramType types.Type) common.Parameter {
	return common.NewParameter(token(common.IDENTIFIER, name), paramType)
}

func parameterAt(line int, column int, name string, paramType types.Type) common.Parameter {
	return common.NewParameter(operatorAt(common.IDENTIFIER, name, line, column), paramType)
}

func funcDeclaration(
	name string,
	parameters []common.Parameter,
	returnType types.Type,
	body ...common.Statement,
) *common.FuncDeclarationStatement {
	return funcDeclarationAt(0, 0, name, parameters, returnType, body...)
}

func funcDeclarationAt(
	line int,
	column int,
	name string,
	parameters []common.Parameter,
	returnType types.Type,
	body ...common.Statement,
) *common.FuncDeclarationStatement {
	nameColumn := column
	if line != 0 {
		nameColumn = column + 5
	}
	return common.NewFuncDeclarationStatement(
		operatorAt(common.FUNC, "func", line, column),
		operatorAt(common.IDENTIFIER, name, line, nameColumn),
		append([]common.Parameter{}, parameters...),
		returnType,
		append([]common.Statement{}, body...),
	)
}

func returnStatement(valueExpression common.Expression) *common.ReturnStatement {
	return common.NewReturnStatement(token(common.RETURN, "return"), valueExpression)
}

func returnStatementAt(line int, column int, valueExpression common.Expression) *common.ReturnStatement {
	return common.NewReturnStatement(operatorAt(common.RETURN, "return", line, column), valueExpression)
}

func call(name string, arguments ...common.Expression) *common.CallExpression {
	return callAt(0, 0, name, arguments...)
}

func callAt(line int, column int, name string, arguments ...common.Expression) *common.CallExpression {
	return common.NewCallExpression(operatorAt(common.IDENTIFIER, name, line, column), append([]common.Expression{}, arguments...))
}

func params(parameters ...common.Parameter) []common.Parameter {
	return parameters
}

func TestValidFunctionsAreAccepted(t *testing.T) {
	runCheckValidTestCases(t, []checkValidTestCase{
		{"empty function", []common.Statement{funcDeclaration("f", nil, types.Void)}},
		{
			"call to a function without return type",
			[]common.Statement{
				funcDeclaration("f", nil, types.Void, common.NewPrintStatement(integerLiteral(1))),
				common.NewExpressionStatement(call("f")),
			},
		},
		{
			"return value used in an expression",
			[]common.Statement{
				funcDeclaration("doble", params(parameter("x", types.Int)), types.Int,
					returnStatement(binary(variable("x"), token(common.STAR, "*"), integerLiteral(2)))),
				common.NewPrintStatement(binary(call("doble", integerLiteral(3)), token(common.PLUS, "+"), integerLiteral(1))),
			},
		},
		{
			"return value assigned to a variable",
			[]common.Statement{
				funcDeclaration("f", nil, types.Str, returnStatement(stringLiteral("a"))),
				varDeclaration("x", types.Str, call("f")),
				common.NewExpressionStatement(assignment("x", call("f"))),
			},
		},
		{
			"parameters of every type",
			[]common.Statement{
				funcDeclaration("f",
					params(parameter("a", types.Int), parameter("b", types.Float), parameter("c", types.Str), parameter("d", types.Bool)),
					types.Bool, returnStatement(variable("d"))),
				common.NewExpressionStatement(call("f", integerLiteral(1), floatLiteral(2.5), stringLiteral("a"), booleanLiteral(true))),
			},
		},
		{
			"call as an argument",
			[]common.Statement{
				funcDeclaration("f", params(parameter("x", types.Int)), types.Int, returnStatement(variable("x"))),
				common.NewExpressionStatement(call("f", call("f", integerLiteral(1)))),
			},
		},
		{
			"call as a condition",
			[]common.Statement{
				funcDeclaration("f", nil, types.Bool, returnStatement(booleanLiteral(true))),
				ifStatement(call("f"), block(), nil),
				whileStatement(call("f"), block(breakStatement())),
			},
		},
		{
			"return without value in a function without return type",
			[]common.Statement{funcDeclaration("f", nil, types.Void, returnStatement(nil))},
		},
		{
			"Void call grouped as an expression statement",
			[]common.Statement{funcDeclaration("f", nil, types.Void), common.NewExpressionStatement(grouping(call("f")))},
		},
		{
			"recursion",
			[]common.Statement{funcDeclaration("fact", params(parameter("n", types.Int)), types.Int,
				ifStatement(binary(variable("n"), token(common.LESS_EQUAL, "<="), integerLiteral(1)), block(returnStatement(integerLiteral(1))), nil),
				returnStatement(binary(variable("n"), token(common.STAR, "*"),
					call("fact", binary(variable("n"), token(common.MINUS, "-"), integerLiteral(1))))),
			)},
		},
		{
			"body uses a variable declared before the function",
			[]common.Statement{
				varDeclaration("x", types.Int, integerLiteral(1)),
				funcDeclaration("f", nil, types.Int, common.NewExpressionStatement(assignment("x", integerLiteral(2))), returnStatement(variable("x"))),
			},
		},
		{
			"parameter shadows an outer variable",
			[]common.Statement{
				varDeclaration("x", types.Str, stringLiteral("a")),
				funcDeclaration("f", params(parameter("x", types.Int)), types.Int, returnStatement(variable("x"))),
			},
		},
		{
			"local variable shadows an outer variable",
			[]common.Statement{
				varDeclaration("x", types.Str, stringLiteral("a")),
				funcDeclaration("f", nil, types.Int, varDeclaration("x", types.Int, integerLiteral(1)), returnStatement(variable("x"))),
			},
		},
		{
			"parameter shadowed in an inner block",
			[]common.Statement{funcDeclaration("f", params(parameter("x", types.Int)), types.Str,
				block(varDeclaration("x", types.Str, stringLiteral("a")), returnStatement(variable("x"))),
			)},
		},
		{
			"nested function uses the parameters of the outer one",
			[]common.Statement{funcDeclaration("f", params(parameter("x", types.Int)), types.Int,
				funcDeclaration("g", nil, types.Int, returnStatement(variable("x"))),
				returnStatement(call("g")),
			)},
		},
		{
			"function inside a block",
			[]common.Statement{block(
				funcDeclaration("f", nil, types.Void),
				common.NewExpressionStatement(call("f")),
			)},
		},
		{
			"same function name in different scopes",
			[]common.Statement{
				funcDeclaration("f", nil, types.Int, returnStatement(integerLiteral(1))),
				block(
					funcDeclaration("f", nil, types.Str, returnStatement(stringLiteral("a"))),
					varDeclaration("x", types.Str, call("f")),
				),
				varDeclaration("y", types.Int, call("f")),
			},
		},
		{
			"variable shadows a function in an inner block",
			[]common.Statement{
				funcDeclaration("f", nil, types.Void),
				block(varDeclaration("f", types.Int, integerLiteral(1)), common.NewPrintStatement(variable("f"))),
				common.NewExpressionStatement(call("f")),
			},
		},
	})
}

func TestFunctionDeclarationOrder(t *testing.T) {
	isEven := funcDeclaration("esPar", params(parameter("n", types.Int)), types.Bool,
		returnStatement(call("esImpar", variable("n"))))
	isOdd := funcDeclaration("esImpar", params(parameter("n", types.Int)), types.Bool,
		returnStatement(call("esPar", variable("n"))))

	runCheckValidTestCases(t, []checkValidTestCase{
		{"mutual recursion between consecutive functions", []common.Statement{isEven, isOdd}},
		{
			"three consecutive functions",
			[]common.Statement{
				funcDeclaration("a", nil, types.Void, common.NewExpressionStatement(call("c"))),
				funcDeclaration("b", nil, types.Void, common.NewExpressionStatement(call("a"))),
				funcDeclaration("c", nil, types.Void, common.NewExpressionStatement(call("b"))),
				common.NewExpressionStatement(call("a")),
			},
		},
		{
			"mutual recursion inside a block",
			[]common.Statement{block(
				funcDeclaration("a", nil, types.Void, common.NewExpressionStatement(call("b"))),
				funcDeclaration("b", nil, types.Void, common.NewExpressionStatement(call("a"))),
			)},
		},
		{
			"mutual recursion inside a function body",
			[]common.Statement{funcDeclaration("f", nil, types.Void,
				funcDeclaration("a", nil, types.Void, common.NewExpressionStatement(call("b"))),
				funcDeclaration("b", nil, types.Void, common.NewExpressionStatement(call("a"))),
			)},
		},
		{
			"later function calls an earlier group",
			[]common.Statement{
				funcDeclaration("a", nil, types.Void),
				varDeclaration("x", types.Int, integerLiteral(1)),
				funcDeclaration("b", nil, types.Void, common.NewExpressionStatement(call("a"))),
			},
		},
	})

	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"call before the declaration",
			[]common.Statement{
				common.NewExpressionStatement(callAt(1, 1, "f")),
				funcDeclaration("f", nil, types.Void),
			},
			"[line 1, column 1] Undefined function 'f'",
		},
		{
			"group interrupted by another statement",
			[]common.Statement{
				funcDeclaration("a", nil, types.Void, common.NewExpressionStatement(callAt(1, 20, "b"))),
				varDeclaration("x", types.Int, integerLiteral(1)),
				funcDeclaration("b", nil, types.Void, common.NewExpressionStatement(call("a"))),
			},
			"[line 1, column 20] Undefined function 'b'",
		},
		{
			"body uses a variable declared after the function",
			[]common.Statement{
				funcDeclaration("f", nil, types.Int, returnStatement(variableAt(1, 26, "x"))),
				varDeclaration("x", types.Int, integerLiteral(1)),
			},
			"[line 1, column 26] Undefined variable 'x'",
		},
		{
			"function declared in a block is not visible outside",
			[]common.Statement{
				block(funcDeclaration("f", nil, types.Void)),
				common.NewExpressionStatement(callAt(2, 1, "f")),
			},
			"[line 2, column 1] Undefined function 'f'",
		},
		{
			"nested function is not visible outside its function",
			[]common.Statement{
				funcDeclaration("f", nil, types.Void, funcDeclaration("g", nil, types.Void)),
				common.NewExpressionStatement(callAt(2, 1, "g")),
			},
			"[line 2, column 1] Undefined function 'g'",
		},
		{
			"parameter is not visible outside its function",
			[]common.Statement{
				funcDeclaration("f", params(parameter("x", types.Int)), types.Void),
				common.NewPrintStatement(variableAt(2, 7, "x")),
			},
			"[line 2, column 7] Undefined variable 'x'",
		},
	})
}

func TestFunctionRedeclaration(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"same function twice",
			[]common.Statement{funcDeclaration("f", nil, types.Void), funcDeclarationAt(2, 1, "f", nil, types.Int, returnStatement(integerLiteral(1)))},
			"[line 2, column 6] Function 'f' already declared in this scope",
		},
		{
			"same function twice in separate groups",
			[]common.Statement{
				funcDeclaration("f", nil, types.Void),
				common.NewPrintStatement(integerLiteral(1)),
				funcDeclarationAt(3, 1, "f", nil, types.Void),
			},
			"[line 3, column 6] Function 'f' already declared in this scope",
		},
		{
			"function with the name of a variable",
			[]common.Statement{varDeclaration("f", types.Int, integerLiteral(1)), funcDeclarationAt(2, 1, "f", nil, types.Void)},
			"[line 2, column 6] Function 'f' already declared in this scope",
		},
		{
			"variable with the name of a function",
			[]common.Statement{funcDeclaration("f", nil, types.Void), varDeclarationAt(2, 1, "f", types.Int, integerLiteral(1))},
			"[line 2, column 5] Variable 'f' already declared in this scope",
		},
		{
			"inside a block",
			[]common.Statement{block(funcDeclaration("f", nil, types.Void), funcDeclarationAt(3, 3, "f", nil, types.Void))},
			"[line 3, column 8] Function 'f' already declared in this scope",
		},
		{
			"duplicate parameter",
			[]common.Statement{funcDeclaration("f", params(parameter("x", types.Int), parameterAt(1, 16, "x", types.Str)), types.Void)},
			"[line 1, column 16] Duplicate parameter 'x' in function 'f'",
		},
		{
			"local variable with the name of a parameter",
			[]common.Statement{funcDeclaration("f", params(parameter("x", types.Int)), types.Void,
				varDeclarationAt(2, 3, "x", types.Int, integerLiteral(1)))},
			"[line 2, column 7] Variable 'x' already declared in this scope",
		},
		{
			"nested function with the name of a parameter",
			[]common.Statement{funcDeclaration("f", params(parameter("g", types.Int)), types.Void,
				funcDeclarationAt(2, 3, "g", nil, types.Void))},
			"[line 2, column 8] Function 'g' already declared in this scope",
		},
	})
}

func TestReturnStatementErrors(t *testing.T) {
	runCheckErrorTestCases(t, []checkErrorTestCase{
		{"return outside function", []common.Statement{returnStatementAt(1, 1, nil)}, "[line 1, column 1] 'return' outside function"},
		{
			"return with value outside function",
			[]common.Statement{returnStatementAt(1, 1, integerLiteral(1))},
			"[line 1, column 1] 'return' outside function",
		},
		{
			"return inside a block outside function",
			[]common.Statement{block(returnStatementAt(2, 3, nil))},
			"[line 2, column 3] 'return' outside function",
		},
		{
			"return inside a while outside function",
			[]common.Statement{whileStatement(booleanLiteral(true), block(returnStatementAt(2, 3, nil)))},
			"[line 2, column 3] 'return' outside function",
		},
		{
			"return after a function declaration",
			[]common.Statement{funcDeclaration("f", nil, types.Void), returnStatementAt(2, 1, nil)},
			"[line 2, column 1] 'return' outside function",
		},
		{
			"value in a function without return type",
			[]common.Statement{funcDeclaration("f", nil, types.Void, returnStatementAt(1, 12, integerLiteral(1)))},
			"[line 1, column 12] Function 'f' cannot return a value",
		},
		{
			"Void call in a function without return type",
			[]common.Statement{
				funcDeclaration("g", nil, types.Void),
				funcDeclaration("f", nil, types.Void, returnStatementAt(2, 12, call("g"))),
			},
			"[line 2, column 12] Function 'f' cannot return a value",
		},
		{
			"missing value",
			[]common.Statement{funcDeclaration("f", nil, types.Int, returnStatementAt(1, 19, nil))},
			"[line 1, column 19] Function 'f' must return a value of type Int",
		},
		{
			"wrong type",
			[]common.Statement{funcDeclaration("f", nil, types.Int, returnStatementAt(1, 19, stringLiteral("a")))},
			"[line 1, column 19] Cannot return String from function 'f' of type Int",
		},
		{
			"Int in a Float function",
			[]common.Statement{funcDeclaration("f", nil, types.Float, returnStatementAt(1, 21, integerLiteral(1)))},
			"[line 1, column 21] Cannot return Int from function 'f' of type Float",
		},
		{
			"Void call in a function with return type",
			[]common.Statement{
				funcDeclaration("g", nil, types.Void),
				funcDeclaration("f", nil, types.Int, returnStatementAt(2, 19, call("g"))),
			},
			"[line 2, column 19] Cannot return Void from function 'f' of type Int",
		},
		{
			"return of the nested function is checked against its own type",
			[]common.Statement{funcDeclaration("f", nil, types.Int,
				funcDeclaration("g", nil, types.Str, returnStatementAt(2, 22, integerLiteral(1))),
				returnStatement(integerLiteral(1)),
			)},
			"[line 2, column 22] Cannot return Int from function 'g' of type String",
		},
		{
			"return after the nested function is checked against the outer type",
			[]common.Statement{funcDeclaration("f", nil, types.Int,
				funcDeclaration("g", nil, types.Str, returnStatement(stringLiteral("a"))),
				returnStatementAt(3, 3, stringLiteral("a")),
			)},
			"[line 3, column 3] Cannot return String from function 'f' of type Int",
		},
		{
			"invalid value does not cascade",
			[]common.Statement{funcDeclaration("f", nil, types.Int,
				returnStatement(unary(operatorAt(common.MINUS, "-", 1, 26), stringLiteral("a"))))},
			"[line 1, column 26] Unsupported operand type for -: String",
		},
	})
}

func TestMissingReturn(t *testing.T) {
	ifReturns := func(elseBranch common.Statement) *common.IfStatement {
		return ifStatement(booleanLiteral(true), block(returnStatement(integerLiteral(1))), elseBranch)
	}

	runCheckValidTestCases(t, []checkValidTestCase{
		{"return at the end", []common.Statement{funcDeclaration("f", nil, types.Int, returnStatement(integerLiteral(1)))}},
		{
			"code after the return",
			[]common.Statement{funcDeclaration("f", nil, types.Int, returnStatement(integerLiteral(1)), common.NewPrintStatement(integerLiteral(2)))},
		},
		{
			"return in a nested block",
			[]common.Statement{funcDeclaration("f", nil, types.Int, block(block(returnStatement(integerLiteral(1)))))},
		},
		{"return in both branches of an if", []common.Statement{funcDeclaration("f", nil, types.Int, ifReturns(block(returnStatement(integerLiteral(2)))))}},
		{
			"return in every branch of an else if chain",
			[]common.Statement{funcDeclaration("f", nil, types.Int, ifReturns(ifReturns(block(returnStatement(integerLiteral(2))))))},
		},
		{"if without else followed by a return", []common.Statement{funcDeclaration("f", nil, types.Int, ifReturns(nil), returnStatement(integerLiteral(2)))}},
		{
			"while followed by a return",
			[]common.Statement{funcDeclaration("f", nil, types.Int,
				whileStatement(booleanLiteral(true), block(returnStatement(integerLiteral(1)))),
				returnStatement(integerLiteral(2)),
			)},
		},
		{
			"function without return type does not need a return",
			[]common.Statement{funcDeclaration("f", nil, types.Void, ifStatement(booleanLiteral(true), block(returnStatement(nil)), nil))},
		},
	})

	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"empty body",
			[]common.Statement{funcDeclarationAt(1, 1, "f", nil, types.Int)},
			"[line 1, column 6] Function 'f' does not return a value on every path",
		},
		{
			"body without return",
			[]common.Statement{funcDeclarationAt(1, 1, "f", nil, types.Int, common.NewPrintStatement(integerLiteral(1)))},
			"[line 1, column 6] Function 'f' does not return a value on every path",
		},
		{
			"if without else",
			[]common.Statement{funcDeclarationAt(1, 1, "f", nil, types.Int, ifReturns(nil))},
			"[line 1, column 6] Function 'f' does not return a value on every path",
		},
		{
			"else without return",
			[]common.Statement{funcDeclarationAt(1, 1, "f", nil, types.Int, ifReturns(block(common.NewPrintStatement(integerLiteral(1)))))},
			"[line 1, column 6] Function 'f' does not return a value on every path",
		},
		{
			"else if without final else",
			[]common.Statement{funcDeclarationAt(1, 1, "f", nil, types.Int, ifReturns(ifReturns(nil)))},
			"[line 1, column 6] Function 'f' does not return a value on every path",
		},
		{
			"return only inside a while",
			[]common.Statement{funcDeclarationAt(1, 1, "f", nil, types.Int,
				whileStatement(booleanLiteral(true), block(returnStatement(integerLiteral(1)))))},
			"[line 1, column 6] Function 'f' does not return a value on every path",
		},
		{
			"return only inside a nested function",
			[]common.Statement{funcDeclarationAt(1, 1, "f", nil, types.Int,
				funcDeclaration("g", nil, types.Int, returnStatement(integerLiteral(1))))},
			"[line 1, column 6] Function 'f' does not return a value on every path",
		},
	})
}

func TestCallErrors(t *testing.T) {
	f := funcDeclaration("f", params(parameter("a", types.Int), parameter("b", types.Str)), types.Int, returnStatement(variable("a")))
	g := funcDeclaration("g", params(parameter("a", types.Float)), types.Void)

	runCheckErrorTestCases(t, []checkErrorTestCase{
		{"undefined function", statements(callAt(1, 1, "f")), "[line 1, column 1] Undefined function 'f'"},
		{
			"calling a variable",
			[]common.Statement{varDeclaration("x", types.Int, integerLiteral(1)), common.NewExpressionStatement(callAt(2, 1, "x"))},
			"[line 2, column 1] 'x' is not a function",
		},
		{
			"calling a parameter",
			[]common.Statement{funcDeclaration("h", params(parameter("x", types.Int)), types.Void, common.NewExpressionStatement(callAt(1, 20, "x")))},
			"[line 1, column 20] 'x' is not a function",
		},
		{
			"calling a variable that shadows a function",
			[]common.Statement{
				funcDeclaration("x", nil, types.Void),
				block(varDeclaration("x", types.Int, integerLiteral(1)), common.NewExpressionStatement(callAt(2, 3, "x"))),
			},
			"[line 2, column 3] 'x' is not a function",
		},
		{"too few arguments", []common.Statement{f, common.NewExpressionStatement(callAt(2, 1, "f", integerLiteral(1)))}, "[line 2, column 1] Function 'f' expects 2 arguments, got 1"},
		{
			"too many arguments",
			[]common.Statement{f, common.NewExpressionStatement(callAt(2, 1, "f", integerLiteral(1), stringLiteral("a"), integerLiteral(2)))},
			"[line 2, column 1] Function 'f' expects 2 arguments, got 3",
		},
		{"arguments to a function without parameters", []common.Statement{funcDeclaration("h", nil, types.Void), common.NewExpressionStatement(callAt(2, 1, "h", integerLiteral(1)))}, "[line 2, column 1] Function 'h' expects 0 arguments, got 1"},
		{"no arguments to a function with one parameter", []common.Statement{g, common.NewExpressionStatement(callAt(2, 1, "g"))}, "[line 2, column 1] Function 'g' expects 1 argument, got 0"},
		{
			"wrong argument type",
			[]common.Statement{f, common.NewExpressionStatement(callAt(2, 1, "f", integerLiteral(1), integerLiteral(2)))},
			"[line 2, column 1] Argument 2 of function 'f' must be String, got Int",
		},
		{"Int argument for a Float parameter", []common.Statement{g, common.NewExpressionStatement(callAt(2, 1, "g", integerLiteral(1)))}, "[line 2, column 1] Argument 1 of function 'g' must be Float, got Int"},
		{
			"Void argument",
			[]common.Statement{g, funcDeclaration("h", nil, types.Void), common.NewExpressionStatement(callAt(3, 1, "g", call("h")))},
			"[line 3, column 1] Argument 1 of function 'g' must be Float, got Void",
		},
		{
			"using a function as a value",
			[]common.Statement{f, common.NewPrintStatement(variableAt(2, 7, "f"))},
			"[line 2, column 7] Cannot use function 'f' as a value",
		},
		{
			"passing a function as an argument",
			[]common.Statement{g, common.NewExpressionStatement(call("g", variableAt(2, 3, "g")))},
			"[line 2, column 3] Cannot use function 'g' as a value",
		},
		{
			"assigning to a function",
			[]common.Statement{f, common.NewExpressionStatement(assignmentAt(2, 1, "f", integerLiteral(1)))},
			"[line 2, column 1] Cannot assign to function 'f'",
		},
		{
			"return type is used in the enclosing expression",
			[]common.Statement{f, common.NewExpressionStatement(binary(call("f", integerLiteral(1), stringLiteral("a")), operatorAt(common.PLUS, "+", 2, 11), stringLiteral("b")))},
			"[line 2, column 11] Unsupported operand types for +: Int and String",
		},
	})

	t.Run("every wrong argument is reported", func(t *testing.T) {
		assertCheckErrors(t, []common.Statement{f, common.NewExpressionStatement(callAt(2, 1, "f", stringLiteral("a"), booleanLiteral(true)))}, []string{
			"[line 2, column 1] Argument 1 of function 'f' must be Int, got String",
			"[line 2, column 1] Argument 2 of function 'f' must be String, got Bool",
		})
	})
}

func TestCallErrorsDoNotCascade(t *testing.T) {
	f := funcDeclaration("f", params(parameter("a", types.Int)), types.Int, returnStatement(variable("a")))

	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"invalid argument keeps the return type",
			[]common.Statement{f, varDeclaration("x", types.Int, call("f", unary(operatorAt(common.MINUS, "-", 2, 17), stringLiteral("a"))))},
			"[line 2, column 17] Unsupported operand type for -: String",
		},
		{
			"undefined argument keeps the return type",
			[]common.Statement{f, common.NewPrintStatement(binary(call("f", variableAt(2, 9, "y")), token(common.PLUS, "+"), integerLiteral(1)))},
			"[line 2, column 9] Undefined variable 'y'",
		},
		{
			"wrong arity keeps the return type",
			[]common.Statement{f, varDeclaration("x", types.Int, callAt(2, 14, "f"))},
			"[line 2, column 14] Function 'f' expects 1 argument, got 0",
		},
		{
			"undefined function poisons the enclosing expression",
			[]common.Statement{varDeclaration("x", types.Int, binary(callAt(1, 14, "g"), token(common.PLUS, "+"), stringLiteral("a")))},
			"[line 1, column 14] Undefined function 'g'",
		},
	})

	t.Run("arguments of an undefined function report their own errors first", func(t *testing.T) {
		assertCheckErrors(t, statements(callAt(1, 1, "g", variableAt(1, 3, "y"))), []string{
			"[line 1, column 3] Undefined variable 'y'",
			"[line 1, column 1] Undefined function 'g'",
		})
	})
}

func TestVoidValueErrors(t *testing.T) {
	p := funcDeclaration("p", nil, types.Void)

	runCheckErrorTestCases(t, []checkErrorTestCase{
		{"printing a Void call", []common.Statement{p, common.NewPrintStatement(callAt(2, 7, "p"))}, "[line 2, column 7] Cannot print a Void value"},
		{"printing a grouped Void call", []common.Statement{p, common.NewPrintStatement(grouping(callAt(2, 8, "p")))}, "[line 2, column 8] Cannot print a Void value"},
		{
			"assigning a Void call in a declaration",
			[]common.Statement{p, varDeclarationAt(2, 1, "x", types.Int, call("p"))},
			"[line 2, column 1] Cannot assign Void to variable of type Int",
		},
		{
			"assigning a Void call to a variable",
			[]common.Statement{p, varDeclaration("x", types.Bool, booleanLiteral(true)), common.NewExpressionStatement(assignmentAt(3, 1, "x", call("p")))},
			"[line 3, column 1] Cannot assign Void to variable of type Bool",
		},
		{
			"comparing Void calls",
			[]common.Statement{p, common.NewExpressionStatement(binary(call("p"), operatorAt(common.DOUBLE_EQUAL, "==", 2, 5), call("p")))},
			"[line 2, column 5] Unsupported operand types for ==: Void and Void",
		},
		{
			"Void call as an operand",
			[]common.Statement{p, common.NewExpressionStatement(binary(call("p"), operatorAt(common.PLUS, "+", 2, 5), integerLiteral(1)))},
			"[line 2, column 5] Unsupported operand types for +: Void and Int",
		},
		{
			"negated Void call",
			[]common.Statement{p, common.NewExpressionStatement(unary(operatorAt(common.NOT, "not", 2, 1), call("p")))},
			"[line 2, column 1] Unsupported operand type for not: Void",
		},
		{
			"Void call as a condition",
			[]common.Statement{p, ifStatementAt(2, 1, call("p"), block(), nil)},
			"[line 2, column 1] Non boolean expression in if condition: Void",
		},
	})
}

func TestLoopStateInsideFunctions(t *testing.T) {
	runCheckValidTestCases(t, []checkValidTestCase{
		{
			"loop inside a function",
			[]common.Statement{funcDeclaration("f", nil, types.Void, whileStatement(booleanLiteral(true), block(breakStatement(), continueStatement())))},
		},
		{
			"break after a function declared inside a loop",
			[]common.Statement{whileStatement(booleanLiteral(true), block(funcDeclaration("f", nil, types.Void), breakStatement()))},
		},
		{
			"return inside a loop inside a function",
			[]common.Statement{funcDeclaration("f", nil, types.Int,
				whileStatement(booleanLiteral(true), block(returnStatement(integerLiteral(1)))),
				returnStatement(integerLiteral(2)),
			)},
		},
	})

	runCheckErrorTestCases(t, []checkErrorTestCase{
		{
			"break in a function declared inside a loop",
			[]common.Statement{whileStatement(booleanLiteral(true), block(funcDeclaration("f", nil, types.Void, breakStatementAt(2, 14))))},
			"[line 2, column 14] 'break' outside loop",
		},
		{
			"continue in a function declared inside a loop",
			[]common.Statement{whileStatement(booleanLiteral(true), block(funcDeclaration("f", nil, types.Void, continueStatementAt(2, 14))))},
			"[line 2, column 14] 'continue' outside loop",
		},
		{
			"break in a function without loops",
			[]common.Statement{funcDeclaration("f", nil, types.Void, breakStatementAt(1, 12))},
			"[line 1, column 12] 'break' outside loop",
		},
		{
			"return after a function declared inside another function",
			[]common.Statement{whileStatement(booleanLiteral(true), block(
				funcDeclaration("f", nil, types.Void),
				returnStatementAt(3, 3, nil),
			))},
			"[line 3, column 3] 'return' outside function",
		},
	})
}

func TestFunctionErrorsAreAccumulated(t *testing.T) {
	assertCheckErrors(t, []common.Statement{
		funcDeclarationAt(1, 1, "f", params(parameter("x", types.Int), parameterAt(1, 16, "x", types.Int)), types.Int,
			returnStatementAt(2, 3, stringLiteral("a")),
		),
		funcDeclarationAt(4, 1, "f", nil, types.Int),
		common.NewExpressionStatement(callAt(5, 1, "f")),
	}, []string{
		"[line 4, column 6] Function 'f' already declared in this scope",
		"[line 1, column 16] Duplicate parameter 'x' in function 'f'",
		"[line 2, column 3] Cannot return String from function 'f' of type Int",
		"[line 4, column 6] Function 'f' does not return a value on every path",
		"[line 5, column 1] Function 'f' expects 2 arguments, got 0",
	})
}

func TestFunctionDistances(t *testing.T) {
	t.Run("call in the same scope", func(t *testing.T) {
		use := call("f")
		assertDistances(t, []common.Statement{
			funcDeclaration("f", nil, types.Void),
			common.NewExpressionStatement(use),
		}, map[common.Expression]int{use: 0})
	})

	t.Run("call from a nested block", func(t *testing.T) {
		use := call("f")
		assertDistances(t, []common.Statement{
			funcDeclaration("f", nil, types.Void),
			block(block(common.NewExpressionStatement(use))),
		}, map[common.Expression]int{use: 2})
	})

	t.Run("parameter and local variable are in the body scope", func(t *testing.T) {
		useParameter := variable("x")
		useLocal := variable("y")
		assertDistances(t, []common.Statement{funcDeclaration("f", params(parameter("x", types.Int)), types.Int,
			varDeclaration("y", types.Int, integerLiteral(1)),
			returnStatement(binary(useParameter, token(common.PLUS, "+"), useLocal)),
		)}, map[common.Expression]int{useParameter: 0, useLocal: 0})
	})

	t.Run("parameter from a block inside the body", func(t *testing.T) {
		use := variable("x")
		assertDistances(t, []common.Statement{funcDeclaration("f", params(parameter("x", types.Int)), types.Void,
			block(common.NewPrintStatement(use)),
		)}, map[common.Expression]int{use: 1})
	})

	t.Run("outer variable from the body", func(t *testing.T) {
		use := assignment("x", integerLiteral(2))
		assertDistances(t, []common.Statement{
			varDeclaration("x", types.Int, integerLiteral(1)),
			funcDeclaration("f", nil, types.Void, common.NewExpressionStatement(use)),
		}, map[common.Expression]int{use: 1})
	})

	t.Run("recursive call", func(t *testing.T) {
		use := call("f")
		assertDistances(t, []common.Statement{
			funcDeclaration("f", nil, types.Void, common.NewExpressionStatement(use)),
		}, map[common.Expression]int{use: 1})
	})

	t.Run("nested function uses the outer parameter and calls the outer function", func(t *testing.T) {
		useParameter := variable("x")
		useOuter := call("f", integerLiteral(1))
		useInner := call("g")
		assertDistances(t, []common.Statement{funcDeclaration("f", params(parameter("x", types.Int)), types.Void,
			funcDeclaration("g", nil, types.Void,
				common.NewPrintStatement(useParameter),
				common.NewExpressionStatement(useOuter),
			),
			common.NewExpressionStatement(useInner),
		)}, map[common.Expression]int{useParameter: 1, useOuter: 2, useInner: 0})
	})

	t.Run("mutual recursion", func(t *testing.T) {
		callB := call("b")
		callA := call("a")
		assertDistances(t, []common.Statement{block(
			funcDeclaration("a", nil, types.Void, common.NewExpressionStatement(callB)),
			funcDeclaration("b", nil, types.Void, common.NewExpressionStatement(callA)),
		)}, map[common.Expression]int{callB: 1, callA: 1})
	})
}
