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

func breakStatement() *common.BreakStatement {
	return common.NewBreakStatement(token(common.BREAK, "break"))
}

func continueStatement() *common.ContinueStatement {
	return common.NewContinueStatement(token(common.CONTINUE, "continue"))
}

func variantDeclaration(name string, fields ...common.FieldDeclaration) common.VariantDeclaration {
	return common.NewVariantDeclaration(token(common.IDENTIFIER, name), append([]common.FieldDeclaration{}, fields...))
}

func enumDeclaration(name string, variants ...common.VariantDeclaration) *common.EnumDeclarationStatement {
	return common.NewEnumDeclarationStatement(token(common.ENUM, "enum"), token(common.IDENTIFIER, name), append([]common.VariantDeclaration{}, variants...))
}

func listDeclaration() *common.EnumDeclarationStatement {
	return enumDeclaration("List",
		variantDeclaration("Empty"),
		variantDeclaration("Node", fieldDeclaration("value", types.Int), fieldDeclaration("rest", named("List"))),
	)
}

func resultDeclaration() *common.EnumDeclarationStatement {
	return enumDeclaration("Result",
		variantDeclaration("Ok", fieldDeclaration("value", types.Int)),
		variantDeclaration("Failure", fieldDeclaration("code", types.Int), fieldDeclaration("message", types.Str)),
		variantDeclaration("Closed"),
	)
}

func (d distanceMap) variant(enumName string, distance int, variant string) *common.FieldAccessExpression {
	return fieldAccess(d.variable(enumName, distance), variant)
}

func (d distanceMap) variantCall(enumName string, distance int, variant string, arguments ...common.Expression) *common.MemberCallExpression {
	return common.NewMemberCallExpression(d.variable(enumName, distance), token(common.IDENTIFIER, variant), append([]common.Expression{}, arguments...))
}

func (d distanceMap) emptyList(distance int) *common.FieldAccessExpression {
	return d.variant("List", distance, "Empty")
}

func (d distanceMap) listNode(distance int, value int64, rest common.Expression) *common.MemberCallExpression {
	return d.variantCall("List", distance, "Node", integerLiteral(value), rest)
}

func (d distanceMap) oneTwoThree(distance int) *common.MemberCallExpression {
	return d.listNode(distance, 1, d.listNode(distance, 2, d.listNode(distance, 3, d.emptyList(distance))))
}

func pattern(name string, bindings ...string) common.MatchPattern {
	bindingTokens := []common.Token{}
	for _, binding := range bindings {
		if binding == "_" {
			bindingTokens = append(bindingTokens, token(common.UNDERSCORE, binding))
		} else {
			bindingTokens = append(bindingTokens, token(common.IDENTIFIER, binding))
		}
	}
	return common.NewMatchPattern(token(common.IDENTIFIER, name), bindingTokens)
}

func wildcard() common.MatchPattern {
	return common.NewMatchPattern(token(common.UNDERSCORE, "_"), []common.Token{})
}

func matchArm(pattern common.MatchPattern, body common.Statement) common.MatchArm {
	return common.NewMatchArm(pattern, body)
}

func matchStatement(expression common.Expression, arms ...common.MatchArm) *common.MatchStatement {
	return common.NewMatchStatement(token(common.MATCH, "match"), expression, append([]common.MatchArm{}, arms...))
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

func (d distanceMap) structLiteral(name string, distance int, fields ...common.FieldInitializer) *common.StructLiteralExpression {
	expression := common.NewStructLiteralExpression(token(common.IDENTIFIER, name), append([]common.FieldInitializer{}, fields...))
	d[expression] = distance
	return expression
}

func (d distanceMap) punto(distance int, x int64, y int64) *common.StructLiteralExpression {
	return d.structLiteral("Punto", distance, fieldInitializer("x", integerLiteral(x)), fieldInitializer("y", integerLiteral(y)))
}

func named(name string) types.Type {
	return types.NamedType{Name: name}
}

func fieldDeclaration(name string, fieldType types.Type) common.FieldDeclaration {
	return common.NewFieldDeclaration(token(common.IDENTIFIER, name), fieldType)
}

func structDeclaration(name string, fields ...common.FieldDeclaration) *common.StructDeclarationStatement {
	return common.NewStructDeclarationStatement(token(common.STRUCT, "struct"), token(common.IDENTIFIER, name), append([]common.FieldDeclaration{}, fields...))
}

func puntoDeclaration() *common.StructDeclarationStatement {
	return structDeclaration("Punto", fieldDeclaration("x", types.Int), fieldDeclaration("y", types.Int))
}

func lineaDeclaration() *common.StructDeclarationStatement {
	return structDeclaration("Linea", fieldDeclaration("desde", named("Punto")), fieldDeclaration("hasta", named("Punto")))
}

func fieldInitializer(name string, valueExpression common.Expression) common.FieldInitializer {
	return common.NewFieldInitializer(token(common.IDENTIFIER, name), valueExpression)
}

func fieldAccess(expression common.Expression, name string) *common.FieldAccessExpression {
	return common.NewFieldAccessExpression(expression, token(common.IDENTIFIER, name))
}

func TestStructPrograms(t *testing.T) {
	runVariableProgramTestCases(t, []variableProgramTestCase{
		{
			"print a struct",
			func(d distanceMap) []common.Statement {
				return []common.Statement{puntoDeclaration(), printStatement(d.punto(0, 1, 2))}
			},
			"Punto { x: 1, y: 2 }",
		},
		{
			"fields are stored in declaration order",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					puntoDeclaration(),
					printStatement(d.structLiteral("Punto", 0, fieldInitializer("y", integerLiteral(2)), fieldInitializer("x", integerLiteral(1)))),
				}
			},
			"Punto { x: 1, y: 2 }",
		},
		{
			"fields of every primitive type",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					structDeclaration("Dato",
						fieldDeclaration("a", types.Int),
						fieldDeclaration("b", types.Float),
						fieldDeclaration("c", types.Str),
						fieldDeclaration("d", types.Bool),
					),
					printStatement(d.structLiteral("Dato", 0,
						fieldInitializer("a", integerLiteral(1)),
						fieldInitializer("b", floatLiteral(2.5)),
						fieldInitializer("c", stringLiteral("hola")),
						fieldInitializer("d", booleanLiteral(true)),
					)),
				}
			},
			`Dato { a: 1, b: 2.5, c: "hola", d: True }`,
		},
		{
			"empty struct",
			func(d distanceMap) []common.Statement {
				return []common.Statement{structDeclaration("Vacio"), printStatement(d.structLiteral("Vacio", 0))}
			},
			"Vacio {}",
		},
		{
			"nested structs",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					puntoDeclaration(),
					lineaDeclaration(),
					printStatement(d.structLiteral("Linea", 0, fieldInitializer("desde", d.punto(0, 0, 0)), fieldInitializer("hasta", d.punto(0, 3, 4)))),
				}
			},
			"Linea { desde: Punto { x: 0, y: 0 }, hasta: Punto { x: 3, y: 4 } }",
		},
		{
			"field access",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					puntoDeclaration(),
					varDeclaration("p", named("Punto"), d.punto(0, 1, 2)),
					printStatement(fieldAccess(d.variable("p", 0), "x")),
					printStatement(fieldAccess(d.variable("p", 0), "y")),
				}
			},
			"12",
		},
		{
			"chained field access",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					puntoDeclaration(),
					lineaDeclaration(),
					varDeclaration("l", named("Linea"),
						d.structLiteral("Linea", 0, fieldInitializer("desde", d.punto(0, 0, 0)), fieldInitializer("hasta", d.punto(0, 3, 4)))),
					printStatement(fieldAccess(fieldAccess(d.variable("l", 0), "hasta"), "y")),
				}
			},
			"4",
		},
		{
			"field of a literal",
			func(d distanceMap) []common.Statement {
				return []common.Statement{puntoDeclaration(), printStatement(fieldAccess(d.punto(0, 7, 8), "y"))}
			},
			"8",
		},
		{
			"string field is printed without quotes",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					structDeclaration("Persona", fieldDeclaration("nombre", types.Str)),
					varDeclaration("persona", named("Persona"), d.structLiteral("Persona", 0, fieldInitializer("nombre", stringLiteral("Ana")))),
					printStatement(fieldAccess(d.variable("persona", 0), "nombre")),
					printStatement(d.variable("persona", 0)),
				}
			},
			`AnaPersona { nombre: "Ana" }`,
		},
		{
			"fields in expressions",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					puntoDeclaration(),
					varDeclaration("p", named("Punto"), d.punto(0, 1, 2)),
					printStatement(binary(fieldAccess(d.variable("p", 0), "x"), common.PLUS, "+",
						binary(fieldAccess(d.variable("p", 0), "y"), common.STAR, "*", integerLiteral(10)))),
				}
			},
			"21",
		},
		{
			"struct as a parameter and as a returned value",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					puntoDeclaration(),
					funcDeclaration("mover", params(parameter("p", named("Punto")), parameter("dx", types.Int)), named("Punto"),
						returnStatement(d.structLiteral("Punto", 1,
							fieldInitializer("x", binary(fieldAccess(d.variable("p", 0), "x"), common.PLUS, "+", d.variable("dx", 0))),
							fieldInitializer("y", fieldAccess(d.variable("p", 0), "y")),
						)),
					),
					printStatement(d.call("mover", 0, d.punto(0, 1, 2), integerLiteral(5))),
				}
			},
			"Punto { x: 6, y: 2 }",
		},
		{
			"assigning a variable does not change other variables",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					puntoDeclaration(),
					varDeclaration("a", named("Punto"), d.punto(0, 1, 2)),
					varDeclaration("b", named("Punto"), d.variable("a", 0)),
					expressionStatement(d.assignment("a", 0, d.punto(0, 3, 4))),
					printStatement(d.variable("a", 0)),
					printStatement(d.variable("b", 0)),
				}
			},
			"Punto { x: 3, y: 4 }Punto { x: 1, y: 2 }",
		},
		{
			"assigning a parameter does not change the argument",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					puntoDeclaration(),
					funcDeclaration("reiniciar", params(parameter("p", named("Punto"))), types.Void,
						expressionStatement(d.assignment("p", 0, d.punto(1, 0, 0))),
					),
					varDeclaration("q", named("Punto"), d.punto(0, 1, 2)),
					expressionStatement(d.call("reiniciar", 0, d.variable("q", 0))),
					printStatement(d.variable("q", 0)),
				}
			},
			"Punto { x: 1, y: 2 }",
		},
		{
			"field values are evaluated from left to right",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					puntoDeclaration(),
					funcDeclaration("uno", nil, types.Int, printStatement(stringLiteral("uno ")), returnStatement(integerLiteral(1))),
					funcDeclaration("dos", nil, types.Int, printStatement(stringLiteral("dos ")), returnStatement(integerLiteral(2))),
					printStatement(d.structLiteral("Punto", 0, fieldInitializer("y", d.call("dos", 0)), fieldInitializer("x", d.call("uno", 0)))),
				}
			},
			"dos uno Punto { x: 1, y: 2 }",
		},
		{
			"equality",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					puntoDeclaration(),
					printStatement(binary(d.punto(0, 1, 2), common.DOUBLE_EQUAL, "==", d.punto(0, 1, 2))),
					printStatement(binary(d.punto(0, 1, 2), common.DOUBLE_EQUAL, "==", d.punto(0, 2, 1))),
					printStatement(binary(d.punto(0, 1, 2), common.NOT_EQUAL, "!=", d.punto(0, 2, 1))),
					printStatement(binary(d.punto(0, 1, 2), common.NOT_EQUAL, "!=", d.punto(0, 1, 2))),
				}
			},
			"TrueFalseTrueFalse",
		},
		{
			"equality of nested and empty structs",
			func(d distanceMap) []common.Statement {
				linea := func(x int64) common.Expression {
					return d.structLiteral("Linea", 0, fieldInitializer("desde", d.punto(0, 0, 0)), fieldInitializer("hasta", d.punto(0, x, 4)))
				}
				return []common.Statement{
					puntoDeclaration(),
					lineaDeclaration(),
					structDeclaration("Vacio"),
					printStatement(binary(linea(3), common.DOUBLE_EQUAL, "==", linea(3))),
					printStatement(binary(linea(3), common.DOUBLE_EQUAL, "==", linea(5))),
					printStatement(binary(d.structLiteral("Vacio", 0), common.DOUBLE_EQUAL, "==", d.structLiteral("Vacio", 0))),
				}
			},
			"TrueFalseTrue",
		},
		{
			"struct declared inside a block",
			func(d distanceMap) []common.Statement {
				return []common.Statement{block(
					structDeclaration("P", fieldDeclaration("x", types.Int)),
					printStatement(d.structLiteral("P", 0, fieldInitializer("x", integerLiteral(1)))),
				)}
			},
			"P { x: 1 }",
		},
		{
			"outer struct used inside a block",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					structDeclaration("P", fieldDeclaration("x", types.Int)),
					block(printStatement(d.structLiteral("P", 1, fieldInitializer("x", integerLiteral(1))))),
				}
			},
			"P { x: 1 }",
		},
		{
			"inner struct shadows an outer struct",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					structDeclaration("P", fieldDeclaration("x", types.Int)),
					printStatement(d.structLiteral("P", 0, fieldInitializer("x", integerLiteral(1)))),
					block(
						structDeclaration("P", fieldDeclaration("y", types.Str)),
						printStatement(d.structLiteral("P", 0, fieldInitializer("y", stringLiteral("a")))),
					),
					printStatement(d.structLiteral("P", 0, fieldInitializer("x", integerLiteral(2)))),
				}
			},
			`P { x: 1 }P { y: "a" }P { x: 2 }`,
		},
		{
			"struct declared inside a function",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					funcDeclaration("f", params(parameter("n", types.Int)), types.Int,
						structDeclaration("P", fieldDeclaration("x", types.Int)),
						varDeclaration("p", named("P"), d.structLiteral("P", 0, fieldInitializer("x", d.variable("n", 0)))),
						returnStatement(fieldAccess(d.variable("p", 0), "x")),
					),
					printStatement(d.call("f", 0, integerLiteral(5))),
					printStatement(d.call("f", 0, integerLiteral(7))),
				}
			},
			"57",
		},
		{
			"field as a condition",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					structDeclaration("Flag", fieldDeclaration("activo", types.Bool)),
					ifStatement(fieldAccess(d.structLiteral("Flag", 0, fieldInitializer("activo", booleanLiteral(true))), "activo"),
						block(printStatement(stringLiteral("si"))),
						block(printStatement(stringLiteral("no"))),
					),
				}
			},
			"si",
		},
		{
			"struct updated in a loop",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					puntoDeclaration(),
					varDeclaration("p", named("Punto"), d.punto(0, 0, 0)),
					whileStatement(binary(fieldAccess(d.variable("p", 0), "x"), common.LESS, "<", integerLiteral(3)), block(
						expressionStatement(d.assignment("p", 1, d.structLiteral("Punto", 1,
							fieldInitializer("x", binary(fieldAccess(d.variable("p", 1), "x"), common.PLUS, "+", integerLiteral(1))),
							fieldInitializer("y", binary(fieldAccess(d.variable("p", 1), "y"), common.PLUS, "+", integerLiteral(2))),
						))),
					)),
					printStatement(d.variable("p", 0)),
				}
			},
			"Punto { x: 3, y: 6 }",
		},
	})
}

func TestStructProgramsOutputBeforeError(t *testing.T) {
	runVariableProgramErrorTestCases(t, []variableProgramErrorTestCase{
		{
			"error in a field value",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					structDeclaration("Medida", fieldDeclaration("valor", types.Float)),
					printStatement(stringLiteral("antes")),
					printStatement(d.structLiteral("Medida", 0, fieldInitializer("valor", divisionByZero()))),
					printStatement(stringLiteral("no se imprime")),
				}
			},
			"antes",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"error in a field after evaluating the previous ones",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					puntoDeclaration(),
					funcDeclaration("uno", nil, types.Int, printStatement(stringLiteral("uno ")), returnStatement(integerLiteral(1))),
					printStatement(d.structLiteral("Punto", 0, fieldInitializer("x", d.call("uno", 0)), fieldInitializer("y", moduloByZero()))),
				}
			},
			"uno ",
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
		{
			"error inside a function that builds a struct",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					puntoDeclaration(),
					funcDeclaration("crear", nil, named("Punto"),
						returnStatement(d.structLiteral("Punto", 1, fieldInitializer("x", moduloByZero()), fieldInitializer("y", integerLiteral(1)))),
					),
					printStatement(stringLiteral("antes")),
					printStatement(fieldAccess(d.call("crear", 0), "x")),
				}
			},
			"antes",
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
		// Sin el typechecker, el intérprete reporta los operadores que los structs no admiten.
		{
			"ordering structs",
			func(d distanceMap) []common.Statement {
				return []common.Statement{puntoDeclaration(), printStatement(binary(d.punto(0, 1, 2), common.LESS, "<", d.punto(0, 1, 2)))}
			},
			"",
			"[line 0, column 0] Unsupported operand types for <: Punto and Punto",
		},
		{
			"adding structs",
			func(d distanceMap) []common.Statement {
				return []common.Statement{puntoDeclaration(), printStatement(binary(d.punto(0, 1, 2), common.PLUS, "+", d.punto(0, 1, 2)))}
			},
			"",
			"[line 0, column 0] Unsupported operand types for +: Punto and Punto",
		},
		{
			"comparing a struct with an int",
			func(d distanceMap) []common.Statement {
				return []common.Statement{puntoDeclaration(), printStatement(binary(d.punto(0, 1, 2), common.DOUBLE_EQUAL, "==", integerLiteral(1)))}
			},
			"",
			"[line 0, column 0] Unsupported operand types for ==: Punto and Int",
		},
	})
}

func TestEnumPrograms(t *testing.T) {
	runVariableProgramTestCases(t, []variableProgramTestCase{
		{
			"print a variant without fields",
			func(d distanceMap) []common.Statement {
				return []common.Statement{enumDeclaration("Color", variantDeclaration("Red")), printStatement(d.variant("Color", 0, "Red"))}
			},
			"Color.Red",
		},
		{
			"print a variant with fields",
			func(d distanceMap) []common.Statement {
				return []common.Statement{resultDeclaration(), printStatement(d.variantCall("Result", 0, "Failure", integerLiteral(404), stringLiteral("not found")))}
			},
			`Result.Failure(404, "not found")`,
		},
		{
			"print a recursive enum",
			func(d distanceMap) []common.Statement {
				return []common.Statement{listDeclaration(), printStatement(d.oneTwoThree(0))}
			},
			"List.Node(1, List.Node(2, List.Node(3, List.Empty)))",
		},
		{
			"variant with a struct field",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					puntoDeclaration(),
					enumDeclaration("Shape", variantDeclaration("Dot", fieldDeclaration("at", named("Punto")))),
					printStatement(d.variantCall("Shape", 0, "Dot", d.punto(0, 1, 2))),
				}
			},
			"Shape.Dot(Punto { x: 1, y: 2 })",
		},
		{
			"arguments are evaluated from left to right",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					resultDeclaration(),
					funcDeclaration("first", nil, types.Int, printStatement(stringLiteral("a ")), returnStatement(integerLiteral(1))),
					funcDeclaration("second", nil, types.Str, printStatement(stringLiteral("b ")), returnStatement(stringLiteral("x"))),
					printStatement(d.variantCall("Result", 0, "Failure", d.call("first", 0), d.call("second", 0))),
				}
			},
			`a b Result.Failure(1, "x")`,
		},
		{
			"enum stored in a variable and reassigned",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					resultDeclaration(),
					varDeclaration("r", named("Result"), d.variant("Result", 0, "Closed")),
					printStatement(d.variable("r", 0)),
					expressionStatement(d.assignment("r", 0, d.variantCall("Result", 0, "Ok", integerLiteral(1)))),
					printStatement(d.variable("r", 0)),
				}
			},
			"Result.ClosedResult.Ok(1)",
		},
		{
			"equality of variants",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					resultDeclaration(),
					printStatement(binary(d.variantCall("Result", 0, "Ok", integerLiteral(1)), common.DOUBLE_EQUAL, "==", d.variantCall("Result", 0, "Ok", integerLiteral(1)))),
					printStatement(binary(d.variant("Result", 0, "Closed"), common.DOUBLE_EQUAL, "==", d.variantCall("Result", 0, "Ok", integerLiteral(1)))),
					printStatement(binary(d.variantCall("Result", 0, "Ok", integerLiteral(1)), common.DOUBLE_EQUAL, "==", d.variantCall("Result", 0, "Ok", integerLiteral(2)))),
					printStatement(binary(d.variant("Result", 0, "Closed"), common.DOUBLE_EQUAL, "==", d.variant("Result", 0, "Closed"))),
				}
			},
			"TrueFalseFalseTrue",
		},
		{
			"inequality of variants",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					resultDeclaration(),
					printStatement(binary(d.variantCall("Result", 0, "Ok", integerLiteral(1)), common.NOT_EQUAL, "!=", d.variantCall("Result", 0, "Ok", integerLiteral(2)))),
					printStatement(binary(d.variant("Result", 0, "Closed"), common.NOT_EQUAL, "!=", d.variant("Result", 0, "Closed"))),
				}
			},
			"TrueFalse",
		},
		{
			"equality of recursive enums",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					listDeclaration(),
					printStatement(binary(d.oneTwoThree(0), common.DOUBLE_EQUAL, "==", d.oneTwoThree(0))),
					printStatement(binary(d.oneTwoThree(0), common.DOUBLE_EQUAL, "==", d.listNode(0, 1, d.listNode(0, 2, d.emptyList(0))))),
				}
			},
			"TrueFalse",
		},
		{
			"equality of structs with enum fields",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					listDeclaration(),
					structDeclaration("Box", fieldDeclaration("items", named("List"))),
					printStatement(binary(
						d.structLiteral("Box", 0, fieldInitializer("items", d.emptyList(0))), common.DOUBLE_EQUAL, "==",
						d.structLiteral("Box", 0, fieldInitializer("items", d.listNode(0, 1, d.emptyList(0)))),
					)),
				}
			},
			"False",
		},
		{
			"enum declared inside a function is defined in every call",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					funcDeclaration("show", params(parameter("n", types.Int)), types.Void,
						enumDeclaration("Local", variantDeclaration("A", fieldDeclaration("n", types.Int))),
						printStatement(d.variantCall("Local", 0, "A", d.variable("n", 0))),
					),
					expressionStatement(d.call("show", 0, integerLiteral(1))),
					expressionStatement(d.call("show", 0, integerLiteral(2))),
				}
			},
			"Local.A(1)Local.A(2)",
		},
		{
			"inner enum that shadows an outer enum",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					listDeclaration(),
					block(
						enumDeclaration("List", variantDeclaration("Single", fieldDeclaration("text", types.Str))),
						printStatement(d.variantCall("List", 0, "Single", stringLiteral("a"))),
					),
					printStatement(d.emptyList(0)),
				}
			},
			`List.Single("a")List.Empty`,
		},
		{
			"variable that shadows an enum is used for field access",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					listDeclaration(),
					puntoDeclaration(),
					block(
						varDeclaration("List", named("Punto"), d.punto(1, 1, 2)),
						printStatement(fieldAccess(d.variable("List", 0), "x")),
					),
				}
			},
			"1",
		},
		{
			"match runs the arm of the variant with its bindings",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					resultDeclaration(),
					matchStatement(d.variantCall("Result", 0, "Failure", integerLiteral(404), stringLiteral("x")),
						matchArm(pattern("Ok", "v"), printStatement(d.variable("v", 0))),
						matchArm(pattern("Failure", "code", "message"), block(printStatement(d.variable("code", 0)), printStatement(d.variable("message", 0)))),
						matchArm(pattern("Closed"), printStatement(stringLiteral("closed"))),
					),
				}
			},
			"404x",
		},
		{
			"function that returns inside the match",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					enumDeclaration("Withdrawal",
						variantDeclaration("Ok", fieldDeclaration("balance", types.Int)),
						variantDeclaration("Insufficient", fieldDeclaration("available", types.Int)),
					),
					varDeclaration("balance", types.Int, integerLiteral(100)),
					funcDeclaration("withdraw", params(parameter("amount", types.Int)), named("Withdrawal"),
						ifStatement(binary(d.variable("amount", 0), common.GREATER, ">", d.variable("balance", 1)), block(
							returnStatement(d.variantCall("Withdrawal", 2, "Insufficient", d.variable("balance", 2))),
						), nil),
						expressionStatement(d.assignment("balance", 1, binary(d.variable("balance", 1), common.MINUS, "-", d.variable("amount", 0)))),
						returnStatement(d.variantCall("Withdrawal", 1, "Ok", d.variable("balance", 1))),
					),
					funcDeclaration("show", params(parameter("amount", types.Int)), types.Bool,
						matchStatement(d.call("withdraw", 1, d.variable("amount", 0)),
							matchArm(pattern("Ok", "b"), block(printStatement(d.variable("b", 0)), returnStatement(booleanLiteral(true)))),
							matchArm(pattern("Insufficient", "a"), block(printStatement(d.variable("a", 0)), returnStatement(booleanLiteral(false)))),
						),
					),
					printStatement(d.call("show", 0, integerLiteral(30))),
					printStatement(d.call("show", 0, integerLiteral(500))),
				}
			},
			"70True70False",
		},
		{
			"recursive sum of a list",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					listDeclaration(),
					funcDeclaration("sum", params(parameter("l", named("List"))), types.Int,
						matchStatement(d.variable("l", 0),
							matchArm(pattern("Node", "v", "rest"), returnStatement(binary(d.variable("v", 0), common.PLUS, "+", d.call("sum", 2, d.variable("rest", 0))))),
							matchArm(wildcard(), returnStatement(integerLiteral(0))),
						),
					),
					printStatement(d.call("sum", 0, d.oneTwoThree(0))),
				}
			},
			"6",
		},
		{
			"length of a list with break and continue",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					listDeclaration(),
					funcDeclaration("length", params(parameter("l", named("List"))), types.Int,
						varDeclaration("n", types.Int, integerLiteral(0)),
						whileStatement(booleanLiteral(true), block(matchStatement(d.variable("l", 1),
							matchArm(pattern("Empty"), breakStatement()),
							matchArm(pattern("Node", "_", "rest"), block(
								expressionStatement(d.assignment("n", 2, binary(d.variable("n", 2), common.PLUS, "+", integerLiteral(1)))),
								expressionStatement(d.assignment("l", 2, d.variable("rest", 0))),
								continueStatement(),
							)),
						))),
						returnStatement(d.variable("n", 0)),
					),
					printStatement(d.call("length", 0, d.oneTwoThree(0))),
				}
			},
			"3",
		},
		{
			"nested match",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					listDeclaration(),
					funcDeclaration("second", params(parameter("l", named("List"))), types.Int,
						matchStatement(d.variable("l", 0),
							matchArm(pattern("Node", "_", "rest"), matchStatement(d.variable("rest", 0),
								matchArm(pattern("Node", "v", "_"), returnStatement(d.variable("v", 0))),
								matchArm(wildcard(), returnStatement(integerLiteral(0))),
							)),
							matchArm(wildcard(), returnStatement(integerLiteral(0))),
						),
					),
					printStatement(d.call("second", 0, d.oneTwoThree(0))),
				}
			},
			"2",
		},
		{
			"function declared in the body of an arm",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					resultDeclaration(),
					matchStatement(d.variant("Result", 0, "Closed"), matchArm(wildcard(), block(
						funcDeclaration("greet", nil, types.Void, printStatement(stringLiteral("hi"))),
						expressionStatement(d.call("greet", 0)),
					))),
				}
			},
			"hi",
		},
		{
			"match on an enum stored in a struct field",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					listDeclaration(),
					structDeclaration("Box", fieldDeclaration("items", named("List"))),
					varDeclaration("b", named("Box"), d.structLiteral("Box", 0, fieldInitializer("items", d.listNode(0, 1, d.emptyList(0))))),
					matchStatement(fieldAccess(d.variable("b", 0), "items"),
						matchArm(pattern("Node", "v", "_"), printStatement(d.variable("v", 0))),
						matchArm(pattern("Empty"), printStatement(stringLiteral("empty"))),
					),
				}
			},
			"1",
		},
	})
}

func TestEnumProgramsOutputBeforeError(t *testing.T) {
	runVariableProgramErrorTestCases(t, []variableProgramErrorTestCase{
		{
			"error in an argument",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					resultDeclaration(),
					printStatement(stringLiteral("before")),
					printStatement(d.variantCall("Result", 0, "Ok", divisionByZero())),
					printStatement(stringLiteral("not printed")),
				}
			},
			"before",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"error in an argument after evaluating the previous ones",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					enumDeclaration("Pair", variantDeclaration("Both", fieldDeclaration("a", types.Int), fieldDeclaration("b", types.Int))),
					funcDeclaration("one", nil, types.Int, printStatement(stringLiteral("one ")), returnStatement(integerLiteral(1))),
					printStatement(d.variantCall("Pair", 0, "Both", d.call("one", 0), moduloByZero())),
				}
			},
			"one ",
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
		{
			"error in the matched expression",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					resultDeclaration(),
					printStatement(stringLiteral("before")),
					matchStatement(d.variantCall("Result", 0, "Ok", divisionByZero()), matchArm(wildcard(), printStatement(stringLiteral("arm")))),
				}
			},
			"before",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"error inside an arm",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					resultDeclaration(),
					matchStatement(d.variant("Result", 0, "Closed"), matchArm(wildcard(), block(
						printStatement(stringLiteral("arm ")),
						expressionStatement(divisionByZero()),
					))),
					printStatement(stringLiteral("not printed")),
				}
			},
			"arm ",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"ordering enums",
			func(d distanceMap) []common.Statement {
				return []common.Statement{resultDeclaration(), printStatement(binary(d.variant("Result", 0, "Closed"), common.LESS, "<", d.variant("Result", 0, "Closed")))}
			},
			"",
			"[line 0, column 0] Unsupported operand types for <: Result and Result",
		},
		{
			"adding enums",
			func(d distanceMap) []common.Statement {
				return []common.Statement{resultDeclaration(), printStatement(binary(d.variant("Result", 0, "Closed"), common.PLUS, "+", d.variant("Result", 0, "Closed")))}
			},
			"",
			"[line 0, column 0] Unsupported operand types for +: Result and Result",
		},
		{
			"comparing an enum with an int",
			func(d distanceMap) []common.Statement {
				return []common.Statement{resultDeclaration(), printStatement(binary(d.variant("Result", 0, "Closed"), common.DOUBLE_EQUAL, "==", integerLiteral(1)))}
			},
			"",
			"[line 0, column 0] Unsupported operand types for ==: Result and Int",
		},
		{
			"comparing an enum with a struct",
			func(d distanceMap) []common.Statement {
				return []common.Statement{
					resultDeclaration(),
					puntoDeclaration(),
					printStatement(binary(d.variant("Result", 0, "Closed"), common.NOT_EQUAL, "!=", d.punto(0, 1, 2))),
				}
			},
			"",
			"[line 0, column 0] Unsupported operand types for !=: Result and Punto",
		},
	})
}
