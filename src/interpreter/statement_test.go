package interpreter

import (
	"bytes"
	"io"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
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
