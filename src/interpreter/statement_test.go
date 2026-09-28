package interpreter

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type statementErrorTestCase struct {
	name            string
	statement       common.Statement
	expectedMessage string
}

type statementOutputTestCase struct {
	name           string
	statement      common.Statement
	expectedOutput string
}

type statementOutputErrorTestCase struct {
	name            string
	statement       common.Statement
	expectedOutput  string
	expectedMessage string
}

type variableOutputTestCase struct {
	name           string
	statement      func(distances distanceMap) common.Statement
	expectedOutput string
}

type variableOutputErrorTestCase struct {
	name            string
	statement       func(distances distanceMap) common.Statement
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

func continueStatement() *common.ContinueStatement {
	return common.NewContinueStatement(token(common.CONTINUE, "continue"))
}

func block(statements ...common.Statement) *common.BlockStatement {
	return common.NewBlockStatement(append([]common.Statement{}, statements...))
}

func printBlock(expressions ...common.Expression) *common.BlockStatement {
	statements := []common.Statement{}
	for _, expression := range expressions {
		statements = append(statements, common.NewPrintStatement(expression))
	}
	return block(statements...)
}

func ifStatement(condition common.Expression, ifBranch common.Statement, elseBranch common.Statement) *common.IfStatement {
	return common.NewIfStatement(token(common.IF, "if"), condition, ifBranch, elseBranch)
}

func whileStatement(condition common.Expression, body common.Statement) *common.WhileStatement {
	return common.NewWhileStatement(token(common.WHILE, "while"), condition, body)
}

func breakStatement() *common.BreakStatement {
	return common.NewBreakStatement(token(common.BREAK, "break"))
}

func printStatement(expression common.Expression) *common.PrintStatement {
	return common.NewPrintStatement(expression)
}

func runStatementOutputTestCases(t *testing.T, testCases []statementOutputTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			output := &bytes.Buffer{}

			err := NewInterpreter(nil, nil, output).execute(testCase.statement)

			if err != nil {
				t.Fatalf("execute(%s) unexpected error: %v", testCase.statement, err)
			}
			if output.String() != testCase.expectedOutput {
				t.Errorf("execute(%s) output = %q; want %q", testCase.statement, output, testCase.expectedOutput)
			}
		})
	}
}

func runStatementOutputErrorTestCases(t *testing.T, testCases []statementOutputErrorTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			output := &bytes.Buffer{}

			err := NewInterpreter(nil, nil, output).execute(testCase.statement)

			if err == nil {
				t.Fatalf("execute(%s) = nil error; want %q", testCase.statement, testCase.expectedMessage)
			}
			if err.Error() != testCase.expectedMessage {
				t.Errorf("execute(%s) error = %q; want %q", testCase.statement, err, testCase.expectedMessage)
			}
			if output.String() != testCase.expectedOutput {
				t.Errorf("execute(%s) output = %q; want %q", testCase.statement, output, testCase.expectedOutput)
			}
		})
	}
}

func runVariableOutputTestCases(t *testing.T, testCases []variableOutputTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			distances := distanceMap{}
			statement := testCase.statement(distances)
			output := &bytes.Buffer{}

			err := NewInterpreter(nil, distances, output).execute(statement)

			if err != nil {
				t.Fatalf("execute(%s) unexpected error: %v", statement, err)
			}
			if output.String() != testCase.expectedOutput {
				t.Errorf("execute(%s) output = %q; want %q", statement, output, testCase.expectedOutput)
			}
		})
	}
}

func runVariableOutputErrorTestCases(t *testing.T, testCases []variableOutputErrorTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			distances := distanceMap{}
			statement := testCase.statement(distances)
			output := &bytes.Buffer{}

			err := NewInterpreter(nil, distances, output).execute(statement)

			if err == nil {
				t.Fatalf("execute(%s) = nil error; want %q", statement, testCase.expectedMessage)
			}
			if err.Error() != testCase.expectedMessage {
				t.Errorf("execute(%s) error = %q; want %q", statement, err, testCase.expectedMessage)
			}
			if output.String() != testCase.expectedOutput {
				t.Errorf("execute(%s) output = %q; want %q", statement, output, testCase.expectedOutput)
			}
		})
	}
}

func TestExpressionStatementExecute(t *testing.T) {
	testCases := []struct {
		name      string
		statement common.Statement
	}{
		{"literal", common.NewExpressionStatement(integerLiteral(3))},
		{
			"binary expression",
			common.NewExpressionStatement(binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2))),
		},
		{"negation", common.NewExpressionStatement(negation(floatLiteral(2.5)))},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := NewInterpreter(nil, nil, io.Discard).execute(testCase.statement); err != nil {
				t.Errorf("execute(%s) unexpected error: %v", testCase.statement, err)
			}
		})
	}
}

func TestExpressionStatementExecuteError(t *testing.T) {
	testCases := []statementErrorTestCase{
		{
			"division by zero",
			common.NewExpressionStatement(divisionByZero()),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"invalid operator",
			common.NewExpressionStatement(binary(integerLiteral(1), common.DOT, ".", integerLiteral(2))),
			"[line 0, column 0] Invalid binary operator: DOT<.>",
		},
		{
			"error inside a grouping",
			common.NewExpressionStatement(grouping(moduloByZero())),
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := NewInterpreter(nil, nil, io.Discard).execute(testCase.statement)

			if err == nil {
				t.Fatalf("execute(%s) = nil error; want %q", testCase.statement, testCase.expectedMessage)
			}
			if err.Error() != testCase.expectedMessage {
				t.Errorf("execute(%s) error = %q; want %q", testCase.statement, err, testCase.expectedMessage)
			}
		})
	}
}

func TestPrintStatementExecute(t *testing.T) {
	testCases := []struct {
		name           string
		statement      common.Statement
		expectedOutput string
	}{
		{"integer", common.NewPrintStatement(integerLiteral(3)), "3"},
		{"negative integer", common.NewPrintStatement(negation(integerLiteral(3))), "-3"},
		{"float with decimals", common.NewPrintStatement(floatLiteral(2.5)), "2.5"},
		{"float without decimals", common.NewPrintStatement(floatLiteral(8)), "8"},
		{"string without double quotes", common.NewPrintStatement(stringLiteral("hola")), "hola"},
		{"empty string", common.NewPrintStatement(stringLiteral("")), ""},
		{"string with double quotes", common.NewPrintStatement(stringLiteral(`dijo "hola"`)), `dijo "hola"`},
		{"string with line feed", common.NewPrintStatement(stringLiteral("a\nb")), "a\nb"},
		{"string with unicode characters", common.NewPrintStatement(stringLiteral("ñandú")), "ñandú"},
		{
			"binary expression result",
			common.NewPrintStatement(binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2))),
			"3",
		},
		{
			"division result",
			common.NewPrintStatement(binary(integerLiteral(1), common.SLASH, "/", integerLiteral(2))),
			"0.5",
		},
		{
			"concatenation result",
			common.NewPrintStatement(binary(stringLiteral("a"), common.PLUS, "+", stringLiteral("b"))),
			"ab",
		},
		{
			"grouped negation result",
			common.NewPrintStatement(grouping(negation(floatLiteral(2.5)))),
			"-2.5",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			output := &bytes.Buffer{}

			err := NewInterpreter(nil, nil, output).execute(testCase.statement)

			if err != nil {
				t.Fatalf("execute(%s) unexpected error: %v", testCase.statement, err)
			}
			if output.String() != testCase.expectedOutput {
				t.Errorf("execute(%s) output = %q; want %q", testCase.statement, output, testCase.expectedOutput)
			}
		})
	}
}

func TestPrintStatementExecuteError(t *testing.T) {
	testCases := []statementErrorTestCase{
		{
			"division by zero",
			common.NewPrintStatement(divisionByZero()),
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"negation of a string",
			common.NewPrintStatement(negation(stringLiteral("a"))),
			"[line 0, column 0] Unsupported operand type for -: String",
		},
		{
			"error inside a grouping",
			common.NewPrintStatement(grouping(moduloByZero())),
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			output := &bytes.Buffer{}

			err := NewInterpreter(nil, nil, output).execute(testCase.statement)

			if err == nil {
				t.Fatalf("execute(%s) = nil error; want %q", testCase.statement, testCase.expectedMessage)
			}
			if err.Error() != testCase.expectedMessage {
				t.Errorf("execute(%s) error = %q; want %q", testCase.statement, err, testCase.expectedMessage)
			}
			if output.Len() != 0 {
				t.Errorf("execute(%s) output = %q; want no output", testCase.statement, output)
			}
		})
	}
}

func TestBlockStatementExecute(t *testing.T) {
	runStatementOutputTestCases(t, []statementOutputTestCase{
		{"empty block", block(), ""},
		{"statements in order", printBlock(integerLiteral(1), stringLiteral("a"), integerLiteral(2)), "1a2"},
		{
			"expression statements do not print",
			block(
				common.NewExpressionStatement(integerLiteral(1)),
				common.NewPrintStatement(integerLiteral(2)),
			),
			"2",
		},
		{
			"nested blocks",
			block(
				common.NewPrintStatement(integerLiteral(1)),
				block(printBlock(integerLiteral(2)), common.NewPrintStatement(integerLiteral(3))),
				common.NewPrintStatement(integerLiteral(4)),
			),
			"1234",
		},
	})
}

func TestBlockStatementExecuteError(t *testing.T) {
	runStatementOutputErrorTestCases(t, []statementOutputErrorTestCase{
		{
			"error stops the remaining statements",
			block(
				common.NewPrintStatement(integerLiteral(1)),
				common.NewExpressionStatement(divisionByZero()),
				common.NewPrintStatement(integerLiteral(2)),
			),
			"1",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"error in a nested block stops the enclosing block",
			block(
				block(common.NewPrintStatement(integerLiteral(1)), common.NewPrintStatement(moduloByZero())),
				common.NewPrintStatement(integerLiteral(2)),
			),
			"1",
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
	})
}

func TestIfStatementExecute(t *testing.T) {
	runStatementOutputTestCases(t, []statementOutputTestCase{
		{"true condition without else", ifStatement(booleanLiteral(true), printBlock(stringLiteral("if")), nil), "if"},
		{"false condition without else", ifStatement(booleanLiteral(false), printBlock(stringLiteral("if")), nil), ""},
		{
			"true condition with else",
			ifStatement(booleanLiteral(true), printBlock(stringLiteral("if")), printBlock(stringLiteral("else"))),
			"if",
		},
		{
			"false condition with else",
			ifStatement(booleanLiteral(false), printBlock(stringLiteral("if")), printBlock(stringLiteral("else"))),
			"else",
		},
		{
			"comparison condition",
			ifStatement(
				binary(integerLiteral(1), common.LESS, "<", floatLiteral(2.5)),
				printBlock(stringLiteral("less")),
				printBlock(stringLiteral("not less")),
			),
			"less",
		},
		{
			"logical condition",
			ifStatement(
				binary(booleanLiteral(true), common.AND, "and", logicalNot(booleanLiteral(true))),
				printBlock(stringLiteral("if")),
				printBlock(stringLiteral("else")),
			),
			"else",
		},
		{
			"grouped condition",
			ifStatement(grouping(booleanLiteral(true)), printBlock(stringLiteral("if")), nil),
			"if",
		},
		{
			"else if runs the first true branch",
			ifStatement(
				booleanLiteral(false),
				printBlock(stringLiteral("first")),
				ifStatement(
					booleanLiteral(true),
					printBlock(stringLiteral("second")),
					ifStatement(booleanLiteral(true), printBlock(stringLiteral("third")), printBlock(stringLiteral("else"))),
				),
			),
			"second",
		},
		{
			"else if runs the final else",
			ifStatement(
				booleanLiteral(false),
				printBlock(stringLiteral("first")),
				ifStatement(booleanLiteral(false), printBlock(stringLiteral("second")), printBlock(stringLiteral("else"))),
			),
			"else",
		},
		{
			"else if without final else",
			ifStatement(
				booleanLiteral(false),
				printBlock(stringLiteral("first")),
				ifStatement(booleanLiteral(false), printBlock(stringLiteral("second")), nil),
			),
			"",
		},
		{
			"nested if",
			ifStatement(
				booleanLiteral(true),
				block(
					common.NewPrintStatement(stringLiteral("outer ")),
					ifStatement(booleanLiteral(false), printBlock(stringLiteral("if")), printBlock(stringLiteral("inner else"))),
				),
				nil,
			),
			"outer inner else",
		},
		{
			"branch that is not taken is not executed",
			ifStatement(
				booleanLiteral(true),
				printBlock(stringLiteral("if")),
				block(common.NewExpressionStatement(divisionByZero())),
			),
			"if",
		},
		{
			"else if condition is not evaluated when the if branch is taken",
			ifStatement(
				booleanLiteral(true),
				printBlock(stringLiteral("if")),
				ifStatement(binary(divisionByZero(), common.LESS, "<", integerLiteral(1)), block(), nil),
			),
			"if",
		},
	})
}

func TestIfStatementExecuteError(t *testing.T) {
	runStatementOutputErrorTestCases(t, []statementOutputErrorTestCase{
		{
			"error in the condition runs no branch",
			ifStatement(
				binary(divisionByZero(), common.LESS, "<", integerLiteral(1)),
				printBlock(stringLiteral("if")),
				printBlock(stringLiteral("else")),
			),
			"",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"error in the if branch",
			ifStatement(booleanLiteral(true), printBlock(stringLiteral("if"), moduloByZero()), nil),
			"if",
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
		{
			"error in the else branch",
			ifStatement(booleanLiteral(false), block(), printBlock(stringLiteral("else"), divisionByZero())),
			"else",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"error in the else if condition",
			ifStatement(
				booleanLiteral(false),
				printBlock(stringLiteral("if")),
				ifStatement(binary(moduloByZero(), common.DOUBLE_EQUAL, "==", integerLiteral(0)), block(), nil),
			),
			"",
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
	})
}

func TestWhileStatementExecute(t *testing.T) {
	runStatementOutputTestCases(t, []statementOutputTestCase{
		{"false condition does not run the body", whileStatement(booleanLiteral(false), printBlock(stringLiteral("body"))), ""},
		{
			"false comparison does not run the body",
			whileStatement(binary(integerLiteral(2), common.LESS, "<", integerLiteral(1)), printBlock(stringLiteral("body"))),
			"",
		},
		{
			"body is not executed when the condition is false",
			whileStatement(booleanLiteral(false), block(common.NewExpressionStatement(divisionByZero()))),
			"",
		},
		{
			"break stops the loop",
			whileStatement(booleanLiteral(true), block(printStatement(stringLiteral("a")), breakStatement())),
			"a",
		},
		{
			"break skips the rest of the body",
			whileStatement(booleanLiteral(true), block(
				printStatement(stringLiteral("a")),
				breakStatement(),
				printStatement(stringLiteral("b")),
			)),
			"a",
		},
		{
			"break as the first statement",
			whileStatement(booleanLiteral(true), block(breakStatement(), printStatement(stringLiteral("a")))),
			"",
		},
		{
			"break inside an if",
			whileStatement(grouping(booleanLiteral(true)), block(
				printStatement(stringLiteral("a")),
				ifStatement(booleanLiteral(true), block(breakStatement()), nil),
				printStatement(stringLiteral("b")),
			)),
			"a",
		},
		{
			"break inside an else",
			whileStatement(booleanLiteral(true), block(
				ifStatement(booleanLiteral(false), printBlock(stringLiteral("if")), block(printStatement(stringLiteral("else")), breakStatement())),
				printStatement(stringLiteral("after if")),
			)),
			"else",
		},
		{
			"break inside a nested block",
			whileStatement(booleanLiteral(true), block(
				block(block(printStatement(stringLiteral("a")), breakStatement()), printStatement(stringLiteral("b"))),
				printStatement(stringLiteral("c")),
			)),
			"a",
		},
		{
			"break in a nested while only stops the inner loop",
			whileStatement(booleanLiteral(true), block(
				whileStatement(booleanLiteral(true), block(printStatement(stringLiteral("inner ")), breakStatement())),
				printStatement(stringLiteral("outer")),
				breakStatement(),
			)),
			"inner outer",
		},
		{
			"execution continues after a loop stopped by a break",
			block(
				whileStatement(booleanLiteral(true), block(printStatement(stringLiteral("loop ")), breakStatement())),
				printStatement(stringLiteral("after")),
			),
			"loop after",
		},
	})
}

func TestWhileStatementExecuteError(t *testing.T) {
	runStatementOutputErrorTestCases(t, []statementOutputErrorTestCase{
		{
			"error in the condition does not run the body",
			whileStatement(binary(divisionByZero(), common.LESS, "<", integerLiteral(1)), printBlock(stringLiteral("body"))),
			"",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"error in the body stops the loop",
			whileStatement(booleanLiteral(true), printBlock(stringLiteral("a"), moduloByZero(), stringLiteral("b"))),
			"a",
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
		{
			"error before a break",
			whileStatement(booleanLiteral(true), block(common.NewExpressionStatement(divisionByZero()), breakStatement())),
			"",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"error in a nested while stops the outer loop",
			block(
				whileStatement(booleanLiteral(true), block(
					whileStatement(booleanLiteral(true), printBlock(stringLiteral("inner"), divisionByZero())),
					printStatement(stringLiteral("outer")),
				)),
				printStatement(stringLiteral("after")),
			),
			"inner",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
	})
}

func TestBreakOutsideLoopPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("Interpret() with a break outside a loop did not panic")
		}
	}()

	NewInterpreter([]common.Statement{breakStatement()}, nil, io.Discard).Interpret()
}

func TestVarDeclarationStatementExecute(t *testing.T) {
	runVariableOutputTestCases(t, []variableOutputTestCase{
		{
			"declaration does not print anything",
			func(d distanceMap) common.Statement { return varDeclaration("x", types.Int, integerLiteral(1)) },
			"",
		},
		{
			"int variable",
			func(d distanceMap) common.Statement {
				return block(varDeclaration("x", types.Int, integerLiteral(1)), printStatement(d.variable("x", 0)))
			},
			"1",
		},
		{
			"float variable",
			func(d distanceMap) common.Statement {
				return block(varDeclaration("x", types.Float, floatLiteral(2.5)), printStatement(d.variable("x", 0)))
			},
			"2.5",
		},
		{
			"string variable",
			func(d distanceMap) common.Statement {
				return block(varDeclaration("x", types.Str, stringLiteral("a")), printStatement(d.variable("x", 0)))
			},
			"a",
		},
		{
			"bool variable",
			func(d distanceMap) common.Statement {
				return block(varDeclaration("x", types.Bool, booleanLiteral(false)), printStatement(d.variable("x", 0)))
			},
			"False",
		},
		{
			"the value is evaluated",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Int, binary(integerLiteral(1), common.PLUS, "+", binary(integerLiteral(2), common.STAR, "*", integerLiteral(3)))),
					printStatement(d.variable("x", 0)),
				)
			},
			"7",
		},
		{
			"value with a previous variable",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Int, integerLiteral(2)),
					varDeclaration("y", types.Int, binary(d.variable("x", 0), common.STAR, "*", integerLiteral(10))),
					printStatement(d.variable("y", 0)),
				)
			},
			"20",
		},
		{
			"the value is copied",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Int, integerLiteral(1)),
					varDeclaration("y", types.Int, d.variable("x", 0)),
					common.NewExpressionStatement(d.assignment("x", 0, integerLiteral(2))),
					printStatement(d.variable("y", 0)),
				)
			},
			"1",
		},
		{
			"inner variable is not visible from the outer scope",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Str, stringLiteral("outer")),
					block(varDeclaration("x", types.Str, stringLiteral("inner"))),
					printStatement(d.variable("x", 0)),
				)
			},
			"outer",
		},
		{
			"shadowed variable",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Str, stringLiteral("outer ")),
					block(
						printStatement(d.variable("x", 1)),
						varDeclaration("x", types.Str, stringLiteral("inner ")),
						printStatement(d.variable("x", 0)),
						printStatement(d.variable("x", 1)),
					),
					printStatement(d.variable("x", 0)),
				)
			},
			"outer inner outer outer ",
		},
		{
			"inner variable initialized with the outer one",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Int, integerLiteral(1)),
					block(
						varDeclaration("x", types.Int, binary(d.variable("x", 1), common.PLUS, "+", integerLiteral(1))),
						printStatement(d.variable("x", 0)),
					),
				)
			},
			"2",
		},
		{
			"declaration in a while body is created again in each iteration",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("i", types.Int, integerLiteral(0)),
					whileStatement(binary(d.variable("i", 0), common.LESS, "<", integerLiteral(3)), block(
						varDeclaration("x", types.Int, binary(d.variable("i", 1), common.STAR, "*", integerLiteral(10))),
						printStatement(d.variable("x", 0)),
						printStatement(stringLiteral(" ")),
						common.NewExpressionStatement(d.assignment("i", 1, binary(d.variable("i", 1), common.PLUS, "+", integerLiteral(1)))),
					)),
				)
			},
			"0 10 20 ",
		},
	})
}

func TestVarDeclarationStatementExecuteError(t *testing.T) {
	runVariableOutputErrorTestCases(t, []variableOutputErrorTestCase{
		{
			"error in the value",
			func(d distanceMap) common.Statement { return varDeclaration("x", types.Int, divisionByZero()) },
			"",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"error in the value stops the block",
			func(d distanceMap) common.Statement {
				return block(
					printStatement(stringLiteral("before")),
					varDeclaration("x", types.Int, moduloByZero()),
					printStatement(stringLiteral("after")),
				)
			},
			"before",
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
	})
}

func TestVariableExpressionExecute(t *testing.T) {
	runVariableOutputTestCases(t, []variableOutputTestCase{
		{
			"variable as an operand",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Int, integerLiteral(3)),
					printStatement(binary(d.variable("x", 0), common.DOUBLE_STAR, "**", integerLiteral(2))),
				)
			},
			"9",
		},
		{
			"same variable in both operands",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Str, stringLiteral("ab")),
					printStatement(binary(d.variable("x", 0), common.PLUS, "+", d.variable("x", 0))),
				)
			},
			"abab",
		},
		{
			"negated variable",
			func(d distanceMap) common.Statement {
				return block(varDeclaration("x", types.Float, floatLiteral(2.5)), printStatement(negation(d.variable("x", 0))))
			},
			"-2.5",
		},
		{
			"grouped variable",
			func(d distanceMap) common.Statement {
				return block(varDeclaration("x", types.Bool, booleanLiteral(true)), printStatement(logicalNot(grouping(d.variable("x", 0)))))
			},
			"False",
		},
		{
			"variable from an enclosing scope",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Int, integerLiteral(1)),
					block(block(printStatement(d.variable("x", 2)))),
				)
			},
			"1",
		},
		{
			"variables from different scopes",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Int, integerLiteral(1)),
					block(
						varDeclaration("y", types.Int, integerLiteral(20)),
						block(
							varDeclaration("z", types.Int, integerLiteral(300)),
							printStatement(binary(d.variable("x", 2), common.PLUS, "+", binary(d.variable("y", 1), common.PLUS, "+", d.variable("z", 0)))),
						),
					),
				)
			},
			"321",
		},
		{
			"bool variable as an if condition",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Bool, booleanLiteral(false)),
					ifStatement(d.variable("x", 0), printBlock(stringLiteral("if")), printBlock(stringLiteral("else"))),
				)
			},
			"else",
		},
		{
			"variable in a short circuit is not needed",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Bool, booleanLiteral(true)),
					printStatement(binary(d.variable("x", 0), common.OR, "or", grouping(binary(divisionByZero(), common.LESS, "<", integerLiteral(1))))),
				)
			},
			"True",
		},
	})
}

func TestVarAssignmentExpressionExecute(t *testing.T) {
	runVariableOutputTestCases(t, []variableOutputTestCase{
		{
			"assignment changes the value",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Int, integerLiteral(1)),
					common.NewExpressionStatement(d.assignment("x", 0, integerLiteral(2))),
					printStatement(d.variable("x", 0)),
				)
			},
			"2",
		},
		{
			"assignment returns the assigned value",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Int, integerLiteral(1)),
					printStatement(d.assignment("x", 0, integerLiteral(5))),
				)
			},
			"5",
		},
		{
			"grouped assignment as an operand",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Int, integerLiteral(1)),
					printStatement(binary(grouping(d.assignment("x", 0, integerLiteral(2))), common.STAR, "*", d.variable("x", 0))),
				)
			},
			"4",
		},
		{
			"assignment of an expression with the variable",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Str, stringLiteral("a")),
					common.NewExpressionStatement(d.assignment("x", 0, binary(d.variable("x", 0), common.PLUS, "+", stringLiteral("b")))),
					common.NewExpressionStatement(d.assignment("x", 0, binary(d.variable("x", 0), common.PLUS, "+", stringLiteral("c")))),
					printStatement(d.variable("x", 0)),
				)
			},
			"abc",
		},
		{
			"chained assignment",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("a", types.Int, integerLiteral(1)),
					varDeclaration("b", types.Int, integerLiteral(2)),
					common.NewExpressionStatement(d.assignment("a", 0, d.assignment("b", 0, integerLiteral(10)))),
					printStatement(d.variable("a", 0)),
					printStatement(stringLiteral(" ")),
					printStatement(d.variable("b", 0)),
				)
			},
			"10 10",
		},
		{
			"chained assignment evaluates the value once",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("a", types.Int, integerLiteral(1)),
					varDeclaration("b", types.Int, integerLiteral(2)),
					common.NewExpressionStatement(d.assignment("a", 0, d.assignment("b", 0, binary(d.variable("b", 0), common.PLUS, "+", integerLiteral(1))))),
					printStatement(d.variable("a", 0)),
					printStatement(stringLiteral(" ")),
					printStatement(d.variable("b", 0)),
				)
			},
			"3 3",
		},
		{
			"assignment to an enclosing scope",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Int, integerLiteral(1)),
					block(block(common.NewExpressionStatement(d.assignment("x", 2, integerLiteral(2))))),
					printStatement(d.variable("x", 0)),
				)
			},
			"2",
		},
		{
			"assignment to a shadowed variable does not change the outer one",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Str, stringLiteral("outer")),
					block(
						varDeclaration("x", types.Str, stringLiteral("inner")),
						common.NewExpressionStatement(d.assignment("x", 0, stringLiteral("new inner "))),
						printStatement(d.variable("x", 0)),
					),
					printStatement(d.variable("x", 0)),
				)
			},
			"new inner outer",
		},
		{
			"assignment as an if condition",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Bool, booleanLiteral(false)),
					ifStatement(d.assignment("x", 0, booleanLiteral(true)), printBlock(stringLiteral("if ")), nil),
					printStatement(d.variable("x", 0)),
				)
			},
			"if True",
		},
		{
			"assignment in a while body",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("i", types.Int, integerLiteral(0)),
					varDeclaration("sum", types.Int, integerLiteral(0)),
					whileStatement(binary(d.variable("i", 0), common.LESS, "<", integerLiteral(4)), block(
						common.NewExpressionStatement(d.assignment("i", 1, binary(d.variable("i", 1), common.PLUS, "+", integerLiteral(1)))),
						common.NewExpressionStatement(d.assignment("sum", 1, binary(d.variable("sum", 1), common.PLUS, "+", d.variable("i", 1)))),
					)),
					printStatement(d.variable("sum", 0)),
				)
			},
			"10",
		},
		{
			"assignment with continue and break",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("i", types.Int, integerLiteral(0)),
					whileStatement(booleanLiteral(true), block(
						common.NewExpressionStatement(d.assignment("i", 1, binary(d.variable("i", 1), common.PLUS, "+", integerLiteral(1)))),
						ifStatement(
							binary(binary(d.variable("i", 1), common.PERCENTAGE, "%", integerLiteral(2)), common.DOUBLE_EQUAL, "==", integerLiteral(0)),
							block(continueStatement()),
							nil,
						),
						ifStatement(binary(d.variable("i", 1), common.GREATER, ">", integerLiteral(5)), block(breakStatement()), nil),
						printStatement(d.variable("i", 1)),
					)),
				)
			},
			"135",
		},
	})
}

func TestVarAssignmentExpressionExecuteError(t *testing.T) {
	runVariableOutputErrorTestCases(t, []variableOutputErrorTestCase{
		{
			"error in the value does not change the variable",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("x", types.Int, integerLiteral(1)),
					printStatement(d.variable("x", 0)),
					common.NewExpressionStatement(d.assignment("x", 0, divisionByZero())),
					printStatement(d.variable("x", 0)),
				)
			},
			"1",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
		{
			"error in a chained assignment",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("a", types.Int, integerLiteral(1)),
					varDeclaration("b", types.Int, integerLiteral(2)),
					common.NewExpressionStatement(d.assignment("a", 0, d.assignment("b", 0, moduloByZero()))),
				)
			},
			"",
			"[line 0, column 0] Cannot divide by zero: 2 % 0",
		},
		{
			"error in a while body after an assignment",
			func(d distanceMap) common.Statement {
				return block(
					varDeclaration("i", types.Int, integerLiteral(0)),
					whileStatement(booleanLiteral(true), block(
						common.NewExpressionStatement(d.assignment("i", 1, binary(d.variable("i", 1), common.PLUS, "+", integerLiteral(1)))),
						printStatement(d.variable("i", 1)),
						ifStatement(binary(d.variable("i", 1), common.DOUBLE_EQUAL, "==", integerLiteral(3)), block(common.NewExpressionStatement(divisionByZero())), nil),
					)),
				)
			},
			"123",
			"[line 0, column 0] Cannot divide by zero: 1 / 0",
		},
	})
}

func TestBlockStatementRestoresTheEnvironment(t *testing.T) {
	testCases := []struct {
		name      string
		statement common.Statement
	}{
		{"after the block ends", block(varDeclaration("x", types.Int, integerLiteral(1)))},
		{"after nested blocks end", block(block(varDeclaration("x", types.Int, integerLiteral(1))))},
		{"after a break", whileStatement(booleanLiteral(true), block(varDeclaration("x", types.Int, integerLiteral(1)), block(breakStatement())))},
		{"after a continue", block(continueStatement())},
		{"after an error", block(varDeclaration("x", types.Int, integerLiteral(1)), block(common.NewExpressionStatement(divisionByZero())))},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			interpreter := NewInterpreter(nil, nil, io.Discard)
			globalEnvironment := interpreter.currentEnvironment

			interpreter.execute(testCase.statement)

			if interpreter.currentEnvironment != globalEnvironment {
				t.Errorf("execute(%s) did not restore the global environment", testCase.statement)
			}
		})
	}
}

func TestUnresolvedVariablePanics(t *testing.T) {
	testCases := []struct {
		name      string
		statement common.Statement
	}{
		{
			"variable",
			block(varDeclaration("x", types.Int, integerLiteral(1)), printStatement(common.NewVariableExpression(token(common.IDENTIFIER, "x")))),
		},
		{
			"assignment",
			block(
				varDeclaration("x", types.Int, integerLiteral(1)),
				common.NewExpressionStatement(common.NewVarAssignmentExpression(token(common.IDENTIFIER, "x"), integerLiteral(2))),
			),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("execute(%s) with a variable without distance did not panic", testCase.statement)
				}
			}()

			NewInterpreter(nil, distanceMap{}, io.Discard).execute(testCase.statement)
		})
	}
}

func (d distanceMap) call(name string, distance int, arguments ...common.Expression) *common.CallExpression {
	expression := common.NewCallExpression(token(common.IDENTIFIER, name), append([]common.Expression{}, arguments...))
	d[expression] = distance
	return expression
}

func parameter(name string, paramType types.Type) common.Parameter {
	return common.NewParameter(token(common.IDENTIFIER, name), paramType)
}

func funcStatement(name string, parameters []common.Parameter, returnType types.Type, body ...common.Statement) *common.FuncDeclarationStatement {
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

func countdownFunction(d distanceMap) *common.FuncDeclarationStatement {
	return funcStatement("contar", []common.Parameter{parameter("n", types.Int)}, types.Int,
		ifStatement(binary(d.variable("n", 0), common.DOUBLE_EQUAL, "==", integerLiteral(0)), block(returnStatement(integerLiteral(0))), nil),
		returnStatement(binary(integerLiteral(1), common.PLUS, "+",
			d.call("contar", 1, binary(d.variable("n", 0), common.MINUS, "-", integerLiteral(1))))),
	)
}

func TestFuncDeclarationDefinesTheFunction(t *testing.T) {
	interpreter := NewInterpreter(nil, nil, io.Discard)
	declaration := funcStatement("f", nil, types.Void)

	if err := interpreter.execute(declaration); err != nil {
		t.Fatalf("execute(%s) unexpected error: %v", declaration, err)
	}

	if defined, _ := interpreter.currentEnvironment.getFunction("f", 0); defined != declaration {
		t.Errorf("execute(%s) defined %v; want the declaration", declaration, defined)
	}
}

func TestReturnStatementExecute(t *testing.T) {
	testCases := []struct {
		name      string
		statement common.Statement
		expected  types.Value
	}{
		{"without value", returnStatement(nil), nil},
		{"with a literal", returnStatement(integerLiteral(1)), types.NewInteger(1)},
		{"with an expression", returnStatement(binary(stringLiteral("a"), common.PLUS, "+", stringLiteral("b"))), types.NewString("ab")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := NewInterpreter(nil, nil, io.Discard).execute(testCase.statement)

			signal, ok := err.(*returnSignal)
			if !ok {
				t.Fatalf("execute(%s) = %v; want a return signal", testCase.statement, err)
			}
			if signal.value != testCase.expected {
				t.Errorf("execute(%s) returned %#v; want %#v", testCase.statement, signal.value, testCase.expected)
			}
		})
	}
}

func TestReturnStatementExecuteError(t *testing.T) {
	statement := returnStatement(divisionByZero())

	err := NewInterpreter(nil, nil, io.Discard).execute(statement)

	if _, ok := err.(*returnSignal); ok || err == nil {
		t.Fatalf("execute(%s) = %v; want the division error", statement, err)
	}
}

func TestReturnSignalPropagatesThroughStatements(t *testing.T) {
	testCases := []struct {
		name      string
		statement common.Statement
	}{
		{"block", block(returnStatement(integerLiteral(1)), printStatement(integerLiteral(2)))},
		{"nested blocks", block(block(returnStatement(integerLiteral(1))), printStatement(integerLiteral(2)))},
		{"if branch", ifStatement(booleanLiteral(true), block(returnStatement(integerLiteral(1))), nil)},
		{"else branch", ifStatement(booleanLiteral(false), block(), block(returnStatement(integerLiteral(1))))},
		{"while body", whileStatement(booleanLiteral(true), block(returnStatement(integerLiteral(1)), printStatement(integerLiteral(2))))},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			interpreter := NewInterpreter(nil, nil, output)
			globalEnvironment := interpreter.currentEnvironment

			err := interpreter.execute(testCase.statement)

			signal, ok := err.(*returnSignal)
			if !ok || signal.value != types.NewInteger(1) {
				t.Fatalf("execute(%s) = %v; want a return signal with 1", testCase.statement, err)
			}
			if output.String() != "" {
				t.Errorf("execute(%s) output = %q; want nothing after the return", testCase.statement, output)
			}
			if interpreter.currentEnvironment != globalEnvironment {
				t.Errorf("execute(%s) did not restore the global environment", testCase.statement)
			}
		})
	}
}

func TestCallRestoresTheEnvironmentAndDepth(t *testing.T) {
	testCases := []struct {
		name      string
		statement func(d distanceMap) common.Statement
	}{
		{
			"after a return",
			func(d distanceMap) common.Statement {
				return block(
					funcStatement("f", nil, types.Int, block(returnStatement(integerLiteral(1)))),
					common.NewExpressionStatement(d.call("f", 0)),
				)
			},
		},
		{
			"after the end of the body",
			func(d distanceMap) common.Statement {
				return block(funcStatement("f", nil, types.Void, block()), common.NewExpressionStatement(d.call("f", 0)))
			},
		},
		{
			"after an error",
			func(d distanceMap) common.Statement {
				return block(
					funcStatement("f", nil, types.Void, block(common.NewExpressionStatement(divisionByZero()))),
					common.NewExpressionStatement(d.call("f", 0)),
				)
			},
		},
		{
			"after exceeding the call depth",
			func(d distanceMap) common.Statement {
				return block(
					funcStatement("f", nil, types.Void, common.NewExpressionStatement(d.call("f", 1))),
					common.NewExpressionStatement(d.call("f", 0)),
				)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			distances := distanceMap{}
			statement := testCase.statement(distances)
			interpreter := NewInterpreter(nil, distances, io.Discard)
			globalEnvironment := interpreter.currentEnvironment

			interpreter.execute(statement)

			if interpreter.currentEnvironment != globalEnvironment {
				t.Errorf("execute(%s) did not restore the global environment", statement)
			}
			if interpreter.callDepth != 0 {
				t.Errorf("execute(%s) left the call depth at %d; want 0", statement, interpreter.callDepth)
			}
		})
	}
}

func TestCallDepthLimit(t *testing.T) {
	testCases := []struct {
		name            string
		argument        int64
		expectedOutput  string
		expectedMessage string
	}{
		{"exactly the maximum depth", maxCallDepth - 1, strconv.Itoa(maxCallDepth - 1), ""},
		{"one call over the maximum depth", maxCallDepth, "", fmt.Sprintf("[line 0, column 0] Maximum call depth of %d exceeded calling 'contar'", maxCallDepth)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			distances := distanceMap{}
			statements := []common.Statement{
				countdownFunction(distances),
				printStatement(distances.call("contar", 0, integerLiteral(testCase.argument))),
			}
			output := &bytes.Buffer{}

			err := NewInterpreter(statements, distances, output).Interpret()

			switch {
			case testCase.expectedMessage == "" && err != nil:
				t.Fatalf("Interpret() unexpected error: %v", err)
			case testCase.expectedMessage != "" && (err == nil || err.Error() != testCase.expectedMessage):
				t.Fatalf("Interpret() error = %v; want %q", err, testCase.expectedMessage)
			}
			if output.String() != testCase.expectedOutput {
				t.Errorf("Interpret() output = %q; want %q", output, testCase.expectedOutput)
			}
		})
	}
}

func TestReturnOutsideFunctionPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("Interpret() with a return outside a function did not panic")
		}
	}()

	NewInterpreter([]common.Statement{block(returnStatement(nil))}, nil, io.Discard).Interpret()
}

func TestFunctionEndingWithoutReturnPanics(t *testing.T) {
	distances := distanceMap{}
	statements := []common.Statement{
		funcStatement("f", nil, types.Int, printStatement(integerLiteral(1))),
		common.NewExpressionStatement(distances.call("f", 0)),
	}

	defer func() {
		if recover() == nil {
			t.Errorf("Interpret() with a function ending without return did not panic")
		}
	}()

	NewInterpreter(statements, distances, io.Discard).Interpret()
}

func TestUnresolvedCallPanics(t *testing.T) {
	statement := block(
		funcStatement("f", nil, types.Void),
		common.NewExpressionStatement(common.NewCallExpression(token(common.IDENTIFIER, "f"), []common.Expression{})),
	)

	defer func() {
		if recover() == nil {
			t.Errorf("execute(%s) with a call without distance did not panic", statement)
		}
	}()

	NewInterpreter(nil, distanceMap{}, io.Discard).execute(statement)
}
