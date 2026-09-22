package common_test

import (
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
)

type statementErrorTestCase struct {
	name            string
	statement       common.Statement
	expectedMessage string
}

type statementStringTestCase struct {
	name      string
	statement common.Statement
	expected  string
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
			if err := testCase.statement.Execute(); err != nil {
				t.Errorf("(%s).Execute() unexpected error: %v", testCase.statement, err)
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
			err := testCase.statement.Execute()

			if err == nil {
				t.Fatalf("(%s).Execute() = nil error; want %q", testCase.statement, testCase.expectedMessage)
			}
			if err.Error() != testCase.expectedMessage {
				t.Errorf("(%s).Execute() error = %q; want %q", testCase.statement, err, testCase.expectedMessage)
			}
		})
	}
}

func TestExpressionStatementString(t *testing.T) {
	testCases := []statementStringTestCase{
		{"literal", common.NewExpressionStatement(integerLiteral(3)), "INTEGER<3>;\n"},
		{
			"binary expression",
			common.NewExpressionStatement(binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2))),
			"(INTEGER<1> PLUS<+> INTEGER<2>);\n",
		},
		{
			"grouped negation",
			common.NewExpressionStatement(grouping(negation(floatLiteral(2.5)))),
			"((MINUS<-> FLOAT<2.5>));\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.statement.String()

			if result != testCase.expected {
				t.Errorf("String() = %q; want %q", result, testCase.expected)
			}
		})
	}
}
