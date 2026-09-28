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

func (d distanceMap) call(name string, distance int, arguments ...common.Expression) *common.CallExpression {
	expression := common.NewCallExpression(token(common.IDENTIFIER, name), append([]common.Expression{}, arguments...))
	d[expression] = distance
	return expression
}

func parameter(name string, paramType types.Type) common.Parameter {
	return common.NewParameter(token(common.IDENTIFIER, name), paramType)
}

func params(parameters ...common.Parameter) []common.Parameter {
	return parameters
}

func funcDeclaration(name string, parameters []common.Parameter, returnType types.Type, body ...common.Statement) *common.FuncDeclarationStatement {
	return common.NewFuncDeclarationStatement(
		token(common.FUNC, "func"),
		token(common.IDENTIFIER, name),
		append([]common.Parameter{}, parameters...),
		returnType,
		append([]common.Statement{}, body...),
	)
}

func returnStatement(valueExpression common.Expression) *common.ReturnStatement {
	return common.NewReturnStatement(token(common.RETURN, "return"), valueExpression)
}

func ifStatement(condition common.Expression, ifBranch common.Statement, elseBranch common.Statement) *common.IfStatement {
	return common.NewIfStatement(token(common.IF, "if"), condition, ifBranch, elseBranch)
}

func whileStatement(condition common.Expression, body common.Statement) *common.WhileStatement {
	return common.NewWhileStatement(token(common.WHILE, "while"), condition, body)
}

func TestFunctionPrograms(t *testing.T) {
	runVariableProgramTestCases(t, []variableProgramTestCase{
		{
			"factorial",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					funcDeclaration("fact", params(parameter("n", types.Int)), types.Int,
						ifStatement(binary(d.variable("n", 0), common.LESS_EQUAL, "<=", integerLiteral(1)), block(returnStatement(integerLiteral(1))), nil),
						returnStatement(binary(d.variable("n", 0), common.STAR, "*",
							d.call("fact", 1, binary(d.variable("n", 0), common.MINUS, "-", integerLiteral(1))))),
					),
					printStatement(d.call("fact", 0, integerLiteral(10))),
				}
			},
			"3628800",
		},
		{
			"each call has its own parameters",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					funcDeclaration("fib", params(parameter("n", types.Int)), types.Int,
						ifStatement(binary(d.variable("n", 0), common.LESS, "<", integerLiteral(2)), block(returnStatement(d.variable("n", 1))), nil),
						returnStatement(binary(
							d.call("fib", 1, binary(d.variable("n", 0), common.MINUS, "-", integerLiteral(1))),
							common.PLUS, "+",
							d.call("fib", 1, binary(d.variable("n", 0), common.MINUS, "-", integerLiteral(2))),
						)),
					),
					printStatement(d.call("fib", 0, integerLiteral(15))),
				}
			},
			"610",
		},
		{
			"mutual recursion",
			func(d distanceMap) []common.Statement {
				isZero := func() common.Expression {
					return binary(d.variable("n", 0), common.DOUBLE_EQUAL, "==", integerLiteral(0))
				}
				nMinusOne := func() common.Expression {
					return binary(d.variable("n", 1), common.MINUS, "-", integerLiteral(1))
				}
				return []common.Statement{
					funcDeclaration("esPar", params(parameter("n", types.Int)), types.Bool,
						ifStatement(isZero(), block(returnStatement(booleanLiteral(true))), block(returnStatement(d.call("esImpar", 2, nMinusOne())))),
					),
					funcDeclaration("esImpar", params(parameter("n", types.Int)), types.Bool,
						ifStatement(isZero(), block(returnStatement(booleanLiteral(false))), block(returnStatement(d.call("esPar", 2, nMinusOne())))),
					),
					printStatement(d.call("esPar", 0, integerLiteral(7))),
					printStatement(d.call("esImpar", 0, integerLiteral(7))),
				}
			},
			"FalseTrue",
		},
		{
			"function without return type with an early return",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					funcDeclaration("saludar", params(parameter("nombre", types.Str)), types.Void,
						printStatement(binary(stringLiteral("hola "), common.PLUS, "+", d.variable("nombre", 0))),
						returnStatement(nil),
						printStatement(stringLiteral("no se imprime")),
					),
					expressionStatement(d.call("saludar", 0, stringLiteral("Ricol"))),
					expressionStatement(grouping(d.call("saludar", 0, stringLiteral("!")))),
				}
			},
			"hola Ricolhola !",
		},
		{
			"return inside a while",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					funcDeclaration("raiz", params(parameter("n", types.Int)), types.Int,
						varDeclaration("i", types.Int, integerLiteral(0)),
						whileStatement(booleanLiteral(true), block(
							ifStatement(
								binary(binary(d.variable("i", 1), common.STAR, "*", d.variable("i", 1)), common.GREATER_EQUAL, ">=", d.variable("n", 1)),
								block(returnStatement(d.variable("i", 2))),
								nil,
							),
							expressionStatement(d.assignment("i", 1, binary(d.variable("i", 1), common.PLUS, "+", integerLiteral(1)))),
						)),
						returnStatement(integerLiteral(0)),
					),
					printStatement(d.call("raiz", 0, integerLiteral(50))),
				}
			},
			"8",
		},
		{
			"body uses the variables of the scope where the function was declared",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					varDeclaration("a", types.Str, stringLiteral("global")),
					block(
						funcDeclaration("retA", nil, types.Str, returnStatement(d.variable("a", 2))),
						printStatement(d.call("retA", 0)),
						varDeclaration("a", types.Str, stringLiteral("block")),
						printStatement(d.call("retA", 0)),
						printStatement(d.variable("a", 0)),
					),
				}
			},
			"globalglobalblock",
		},
		{
			"function modifies a global variable",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					varDeclaration("total", types.Int, integerLiteral(0)),
					funcDeclaration("sumar", params(parameter("x", types.Int)), types.Void,
						expressionStatement(d.assignment("total", 1, binary(d.variable("total", 1), common.PLUS, "+", d.variable("x", 0)))),
					),
					expressionStatement(d.call("sumar", 0, integerLiteral(2))),
					expressionStatement(d.call("sumar", 0, integerLiteral(3))),
					printStatement(d.variable("total", 0)),
				}
			},
			"5",
		},
		{
			"nested function modifies a local variable of the outer one",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					funcDeclaration("sumarHasta", params(parameter("n", types.Int)), types.Int,
						varDeclaration("total", types.Int, integerLiteral(0)),
						funcDeclaration("sumar", params(parameter("x", types.Int)), types.Void,
							expressionStatement(d.assignment("total", 1, binary(d.variable("total", 1), common.PLUS, "+", d.variable("x", 0)))),
						),
						varDeclaration("i", types.Int, integerLiteral(1)),
						whileStatement(binary(d.variable("i", 0), common.LESS_EQUAL, "<=", d.variable("n", 0)), block(
							expressionStatement(d.call("sumar", 1, d.variable("i", 1))),
							expressionStatement(d.assignment("i", 1, binary(d.variable("i", 1), common.PLUS, "+", integerLiteral(1)))),
						)),
						returnStatement(d.variable("total", 0)),
					),
					printStatement(d.call("sumarHasta", 0, integerLiteral(100))),
					printStatement(stringLiteral(" ")),
					printStatement(d.call("sumarHasta", 0, integerLiteral(3))),
				}
			},
			"5050 6",
		},
		{
			"local variables are new in each call",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					funcDeclaration("f", nil, types.Int,
						varDeclaration("c", types.Int, integerLiteral(0)),
						expressionStatement(d.assignment("c", 0, binary(d.variable("c", 0), common.PLUS, "+", integerLiteral(1)))),
						returnStatement(d.variable("c", 0)),
					),
					printStatement(d.call("f", 0)),
					printStatement(d.call("f", 0)),
				}
			},
			"11",
		},
		{
			"arguments are evaluated in the scope of the caller",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					varDeclaration("x", types.Int, integerLiteral(10)),
					funcDeclaration("f", params(parameter("x", types.Int)), types.Int,
						returnStatement(binary(d.variable("x", 0), common.PLUS, "+", integerLiteral(1))),
					),
					printStatement(d.call("f", 0, binary(d.variable("x", 0), common.STAR, "*", integerLiteral(2)))),
					printStatement(stringLiteral(" ")),
					printStatement(d.variable("x", 0)),
				}
			},
			"21 10",
		},
		{
			"arguments are evaluated from left to right",
			func(d distanceMap) []common.Statement {
				increment := func() common.Expression {
					return d.assignment("i", 0, binary(d.variable("i", 0), common.PLUS, "+", integerLiteral(1)))
				}
				return []common.Statement{
					funcDeclaration("par", params(parameter("a", types.Int), parameter("b", types.Int)), types.Int,
						returnStatement(binary(binary(d.variable("a", 0), common.STAR, "*", integerLiteral(10)), common.PLUS, "+", d.variable("b", 0))),
					),
					varDeclaration("i", types.Int, integerLiteral(0)),
					printStatement(d.call("par", 0, increment(), increment())),
				}
			},
			"12",
		},
		{
			"function declared in a block shadows the outer one",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					funcDeclaration("f", nil, types.Str, returnStatement(stringLiteral("global "))),
					block(
						funcDeclaration("f", nil, types.Str, returnStatement(stringLiteral("block "))),
						printStatement(d.call("f", 0)),
						block(printStatement(d.call("f", 1))),
					),
					printStatement(d.call("f", 0)),
				}
			},
			"block block global ",
		},
		{
			"calls as operands and arguments",
			func(d distanceMap) []common.Statement {
				identity := func(argument common.Expression) common.Expression { return d.call("id", 0, argument) }
				return []common.Statement{
					funcDeclaration("id", params(parameter("n", types.Int)), types.Int, returnStatement(d.variable("n", 0))),
					printStatement(binary(identity(integerLiteral(1)), common.PLUS, "+", binary(identity(integerLiteral(2)), common.STAR, "*", identity(integerLiteral(3))))),
					printStatement(stringLiteral(" ")),
					printStatement(identity(identity(identity(integerLiteral(4))))),
				}
			},
			"7 4",
		},
	})
}

func TestFunctionProgramsOutputBeforeError(t *testing.T) {
	runVariableProgramErrorTestCases(t, []variableProgramErrorTestCase{
		{
			"error inside the body",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					funcDeclaration("dividir", params(parameter("a", types.Int), parameter("b", types.Int)), types.Float,
						returnStatement(binary(d.variable("a", 0), common.SLASH, "/", d.variable("b", 0))),
					),
					printStatement(d.call("dividir", 0, integerLiteral(1), integerLiteral(2))),
					printStatement(d.call("dividir", 0, integerLiteral(1), integerLiteral(0))),
					printStatement(stringLiteral("no se imprime")),
				}
			},
			"0.5",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"error in an argument does not execute the body",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					funcDeclaration("f", params(parameter("x", types.Float)), types.Void, printStatement(stringLiteral("no se imprime"))),
					printStatement(stringLiteral("antes")),
					expressionStatement(d.call("f", 0, divisionByZero())),
				}
			},
			"antes",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"error in a nested call",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					funcDeclaration("g", nil, types.Int, returnStatement(moduloByZero())),
					funcDeclaration("f", nil, types.Int,
						printStatement(stringLiteral("f ")),
						returnStatement(d.call("g", 1)),
					),
					printStatement(d.call("f", 0)),
				}
			},
			"f ",
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
		{
			"infinite recursion",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					funcDeclaration("infinita", nil, types.Void, expressionStatement(d.call("infinita", 1))),
					printStatement(stringLiteral("antes")),
					expressionStatement(d.call("infinita", 0)),
				}
			},
			"antes",
			"[line 0, column 0] Maximum call depth of 10000 exceeded calling 'infinita'",
		},
	})
}
