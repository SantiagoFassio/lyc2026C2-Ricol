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

type variableProgramTestCase struct {
	name           string
	statements     func(distances distanceMap) []common.Statement
	expectedOutput string
}

type variableProgramErrorTestCase struct {
	name            string
	statements      func(distances distanceMap) []common.Statement
	expectedOutput  string
	expectedMessage string
}

type distanceMap map[common.Expression]int

func (d distanceMap) variable(name string, distance int) *common.VariableExpression {
	expression := common.NewVariableExpression(token(common.IDENTIFIER, name))
	d[expression] = distance
	return expression
}

func (d distanceMap) assignment(name string, distance int, valueExpression common.Expression) *common.VarAssignmentExpression {
	expression := common.NewVarAssignmentExpression(token(common.IDENTIFIER, name), valueExpression)
	d[expression] = distance
	return expression
}

func varDeclaration(name string, varType types.Type, valueExpression common.Expression) *common.VarDeclarationStatement {
	return common.NewVarDeclarationStatement(token(common.LET, "let"), token(common.IDENTIFIER, name), varType, valueExpression)
}

func block(statements ...common.Statement) *common.BlockStatement {
	return common.NewBlockStatement(append([]common.Statement{}, statements...))
}

func printStatement(expression common.Expression) *common.PrintStatement {
	return common.NewPrintStatement(expression)
}

func expressionStatement(expression common.Expression) *common.ExpressionStatement {
	return common.NewExpressionStatement(expression)
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

	err := interpreter.NewInterpreter(inputStatements, nil, io.Discard).Interpret()

	if err != nil {
		t.Errorf("interpreter.Interpret(%v) unexpected error: %v", inputStatements, err)
	}
}

func assertInterpretError(t *testing.T, inputStatements []common.Statement, expectedMessage string) {
	t.Helper()

	err := interpreter.NewInterpreter(inputStatements, nil, io.Discard).Interpret()

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

	err := interpreter.NewInterpreter(inputStatements, nil, output).Interpret()

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

	err := interpreter.NewInterpreter(inputStatements, nil, output).Interpret()

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

func runVariableProgramTestCases(t *testing.T, testCases []variableProgramTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			distances := distanceMap{}
			inputStatements := testCase.statements(distances)
			output := &bytes.Buffer{}

			err := interpreter.NewInterpreter(inputStatements, distances, output).Interpret()

			if err != nil {
				t.Fatalf("interpreter.Interpret(%v) unexpected error: %v", inputStatements, err)
			}
			if output.String() != testCase.expectedOutput {
				t.Errorf("interpreter.Interpret(%v) output = %q; want %q", inputStatements, output, testCase.expectedOutput)
			}
		})
	}
}

func runVariableProgramErrorTestCases(t *testing.T, testCases []variableProgramErrorTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			distances := distanceMap{}
			inputStatements := testCase.statements(distances)
			output := &bytes.Buffer{}

			err := interpreter.NewInterpreter(inputStatements, distances, output).Interpret()

			if err == nil {
				t.Fatalf("interpreter.Interpret(%v) = nil error; want %q", inputStatements, testCase.expectedMessage)
			}
			if err.Error() != testCase.expectedMessage {
				t.Errorf("interpreter.Interpret(%v) error = %q; want %q", inputStatements, err, testCase.expectedMessage)
			}
			if output.String() != testCase.expectedOutput {
				t.Errorf("interpreter.Interpret(%v) output = %q; want %q", inputStatements, output, testCase.expectedOutput)
			}
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
		{
			"result of a logical expression",
			printStatements(binary(booleanLiteral(true), common.AND, "and", booleanLiteral(false))),
			"False",
		},
		{
			"result of a not",
			printStatements(common.NewUnaryExpression(token(common.NOT, "not"), booleanLiteral(false))),
			"True",
		},
	})
}

func TestVariablePrograms(t *testing.T) {
	runVariableProgramTestCases(t, []variableProgramTestCase{
		{
			"global variables",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					varDeclaration("greeting", types.Str, stringLiteral("Hola")),
					varDeclaration("name", types.Str, stringLiteral("Ricol")),
					printStatement(binary(binary(d.variable("greeting", 0), common.PLUS, "+", stringLiteral(", ")), common.PLUS, "+", d.variable("name", 0))),
				}
			},
			"Hola, Ricol",
		},
		{
			"swap two variables",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					varDeclaration("a", types.Int, integerLiteral(1)),
					varDeclaration("b", types.Int, integerLiteral(2)),
					block(
						varDeclaration("temp", types.Int, d.variable("a", 1)),
						expressionStatement(d.assignment("a", 1, d.variable("b", 1))),
						expressionStatement(d.assignment("b", 1, d.variable("temp", 0))),
					),
					printStatement(d.variable("a", 0)),
					printStatement(d.variable("b", 0)),
				}
			},
			"21",
		},
		{
			"counter loop",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					varDeclaration("i", types.Int, integerLiteral(1)),
					common.NewWhileStatement(
						token(common.WHILE, "while"),
						binary(d.variable("i", 0), common.LESS_EQUAL, "<=", integerLiteral(5)),
						block(
							printStatement(d.variable("i", 1)),
							expressionStatement(d.assignment("i", 1, binary(d.variable("i", 1), common.PLUS, "+", integerLiteral(1)))),
						),
					),
					printStatement(stringLiteral(" end ")),
					printStatement(d.variable("i", 0)),
				}
			},
			"12345 end 6",
		},
		{
			"fibonacci",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					varDeclaration("previous", types.Int, integerLiteral(0)),
					varDeclaration("current", types.Int, integerLiteral(1)),
					varDeclaration("count", types.Int, integerLiteral(0)),
					common.NewWhileStatement(
						token(common.WHILE, "while"),
						binary(d.variable("count", 0), common.LESS, "<", integerLiteral(8)),
						block(
							printStatement(d.variable("previous", 1)),
							printStatement(stringLiteral(" ")),
							varDeclaration("next", types.Int, binary(d.variable("previous", 1), common.PLUS, "+", d.variable("current", 1))),
							expressionStatement(d.assignment("previous", 1, d.variable("current", 1))),
							expressionStatement(d.assignment("current", 1, d.variable("next", 0))),
							expressionStatement(d.assignment("count", 1, binary(d.variable("count", 1), common.PLUS, "+", integerLiteral(1)))),
						),
					),
				}
			},
			"0 1 1 2 3 5 8 13 ",
		},
		{
			"flag variable stops a loop",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					varDeclaration("running", types.Bool, booleanLiteral(true)),
					varDeclaration("total", types.Float, floatLiteral(0.5)),
					common.NewWhileStatement(
						token(common.WHILE, "while"),
						d.variable("running", 0),
						block(
							expressionStatement(d.assignment("total", 1, binary(d.variable("total", 1), common.STAR, "*", integerLiteral(2)))),
							common.NewIfStatement(
								token(common.IF, "if"),
								binary(d.variable("total", 1), common.GREATER, ">", integerLiteral(10)),
								block(expressionStatement(d.assignment("running", 2, booleanLiteral(false)))),
								nil,
							),
						),
					),
					printStatement(d.variable("total", 0)),
				}
			},
			"16",
		},
		{
			"shadowing in nested blocks",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					varDeclaration("x", types.Str, stringLiteral("global")),
					block(
						varDeclaration("x", types.Str, stringLiteral("block")),
						block(
							printStatement(d.variable("x", 1)),
							printStatement(stringLiteral(" ")),
							varDeclaration("x", types.Str, stringLiteral("inner")),
							printStatement(d.variable("x", 0)),
							printStatement(stringLiteral(" ")),
						),
						printStatement(d.variable("x", 0)),
						printStatement(stringLiteral(" ")),
					),
					printStatement(d.variable("x", 0)),
				}
			},
			"block inner block global",
		},
		{
			"chained assignment resets several variables",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					varDeclaration("a", types.Int, integerLiteral(1)),
					varDeclaration("b", types.Int, integerLiteral(2)),
					varDeclaration("c", types.Int, integerLiteral(3)),
					expressionStatement(d.assignment("a", 0, d.assignment("b", 0, d.assignment("c", 0, integerLiteral(0))))),
					printStatement(binary(binary(d.variable("a", 0), common.PLUS, "+", d.variable("b", 0)), common.PLUS, "+", d.variable("c", 0))),
				}
			},
			"0",
		},
	})
}

func TestVariableProgramsOutputBeforeError(t *testing.T) {
	runVariableProgramErrorTestCases(t, []variableProgramErrorTestCase{
		{
			"division by a variable that reaches zero",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					varDeclaration("divisor", types.Int, integerLiteral(2)),
					common.NewWhileStatement(
						token(common.WHILE, "while"),
						booleanLiteral(true),
						block(
							printStatement(binary(integerLiteral(10), common.DOUBLE_SLASH, "//", d.variable("divisor", 1))),
							printStatement(stringLiteral(" ")),
							expressionStatement(d.assignment("divisor", 1, binary(d.variable("divisor", 1), common.MINUS, "-", integerLiteral(1)))),
						),
					),
				}
			},
			"5 10 ",
			"[line 0, column 0] Cannot divide by zero: 10 // 0",
		},
		{
			"error in the value of a declaration",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					varDeclaration("x", types.Int, integerLiteral(1)),
					printStatement(d.variable("x", 0)),
					varDeclaration("y", types.Int, binary(d.variable("x", 0), common.PERCENTAGE, "%", integerLiteral(0))),
					printStatement(d.variable("y", 0)),
				}
			},
			"1",
			"[line 0, column 0] Cannot divide by zero: 1 % 0",
		},
	})
}
