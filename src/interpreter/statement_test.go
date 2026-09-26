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
			if err := NewInterpreter(nil, io.Discard).execute(testCase.statement); err != nil {
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
			err := NewInterpreter(nil, io.Discard).execute(testCase.statement)

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

			err := NewInterpreter(nil, output).execute(testCase.statement)

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

			err := NewInterpreter(nil, output).execute(testCase.statement)

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
