package common_test

import (
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
)

type statementStringTestCase struct {
	name      string
	statement common.Statement
	expected  string
}

func block(statements ...common.Statement) *common.BlockStatement {
	return common.NewBlockStatement(append([]common.Statement{}, statements...))
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

func continueStatement() *common.ContinueStatement {
	return common.NewContinueStatement(token(common.CONTINUE, "continue"))
}

func printStatement(value int64) *common.PrintStatement {
	return common.NewPrintStatement(integerLiteral(value))
}

func runStatementStringTestCases(t *testing.T, testCases []statementStringTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.statement.String()

			if result != testCase.expected {
				t.Errorf("String() = %q; want %q", result, testCase.expected)
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

func TestPrintStatementString(t *testing.T) {
	testCases := []statementStringTestCase{
		{"integer literal", common.NewPrintStatement(integerLiteral(3)), "PRINT INTEGER<3>;\n"},
		{"string literal", common.NewPrintStatement(stringLiteral("hola")), `PRINT STRING<"hola">;` + "\n"},
		{
			"binary expression",
			common.NewPrintStatement(binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2))),
			"PRINT (INTEGER<1> PLUS<+> INTEGER<2>);\n",
		},
		{
			"grouped negation",
			common.NewPrintStatement(grouping(negation(floatLiteral(2.5)))),
			"PRINT ((MINUS<-> FLOAT<2.5>));\n",
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

func TestBlockStatementString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{"empty block", block(), "{\n}\n"},
		{
			"block with statements",
			block(printStatement(1), common.NewExpressionStatement(integerLiteral(2))),
			"{\n" +
				"  PRINT INTEGER<1>;\n" +
				"  INTEGER<2>;\n" +
				"}\n",
		},
		{
			"nested blocks accumulate the indentation",
			block(printStatement(1), block(printStatement(2), block(printStatement(3)))),
			"{\n" +
				"  PRINT INTEGER<1>;\n" +
				"  {\n" +
				"    PRINT INTEGER<2>;\n" +
				"    {\n" +
				"      PRINT INTEGER<3>;\n" +
				"    }\n" +
				"  }\n" +
				"}\n",
		},
		{
			"nested empty block",
			block(block()),
			"{\n" +
				"  {\n" +
				"  }\n" +
				"}\n",
		},
	})
}

func TestIfStatementString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{
			"if without else",
			ifStatement(booleanLiteral(true), block(printStatement(1)), nil),
			"if (TRUE<True>) {\n" +
				"  PRINT INTEGER<1>;\n" +
				"}\n",
		},
		{
			"else on the same line as the closed brace",
			ifStatement(booleanLiteral(false), block(printStatement(1)), block(printStatement(2))),
			"if (FALSE<False>) {\n" +
				"  PRINT INTEGER<1>;\n" +
				"} else {\n" +
				"  PRINT INTEGER<2>;\n" +
				"}\n",
		},
		{
			"binary condition",
			ifStatement(binary(integerLiteral(1), common.LESS, "<", integerLiteral(2)), block(), nil),
			"if ((INTEGER<1> LESS<<> INTEGER<2>)) {\n" +
				"}\n",
		},
		{
			"else if chain",
			ifStatement(
				booleanLiteral(false),
				block(printStatement(1)),
				ifStatement(booleanLiteral(true), block(printStatement(2)), block(printStatement(3))),
			),
			"if (FALSE<False>) {\n" +
				"  PRINT INTEGER<1>;\n" +
				"} else if (TRUE<True>) {\n" +
				"  PRINT INTEGER<2>;\n" +
				"} else {\n" +
				"  PRINT INTEGER<3>;\n" +
				"}\n",
		},
		{
			"nested if is indented",
			ifStatement(
				booleanLiteral(true),
				block(printStatement(1), ifStatement(booleanLiteral(false), block(printStatement(2)), block(printStatement(3)))),
				nil,
			),
			"if (TRUE<True>) {\n" +
				"  PRINT INTEGER<1>;\n" +
				"  if (FALSE<False>) {\n" +
				"    PRINT INTEGER<2>;\n" +
				"  } else {\n" +
				"    PRINT INTEGER<3>;\n" +
				"  }\n" +
				"}\n",
		},
		{
			"if inside a block",
			block(ifStatement(booleanLiteral(true), block(printStatement(1)), nil)),
			"{\n" +
				"  if (TRUE<True>) {\n" +
				"    PRINT INTEGER<1>;\n" +
				"  }\n" +
				"}\n",
		},
	})
}

func TestWhileStatementString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{
			"while with a statement",
			whileStatement(booleanLiteral(true), block(printStatement(1))),
			"while (TRUE<True>) {\n" +
				"  PRINT INTEGER<1>;\n" +
				"}\n",
		},
		{
			"while with an empty body",
			whileStatement(booleanLiteral(false), block()),
			"while (FALSE<False>) {\n" +
				"}\n",
		},
		{
			"binary condition",
			whileStatement(binary(integerLiteral(1), common.LESS, "<", integerLiteral(2)), block()),
			"while ((INTEGER<1> LESS<<> INTEGER<2>)) {\n" +
				"}\n",
		},
		{
			"break and continue in the body",
			whileStatement(booleanLiteral(true), block(continueStatement(), breakStatement())),
			"while (TRUE<True>) {\n" +
				"  continue;\n" +
				"  break;\n" +
				"}\n",
		},
		{
			"nested while is indented",
			whileStatement(
				booleanLiteral(true),
				block(printStatement(1), whileStatement(booleanLiteral(false), block(breakStatement()))),
			),
			"while (TRUE<True>) {\n" +
				"  PRINT INTEGER<1>;\n" +
				"  while (FALSE<False>) {\n" +
				"    break;\n" +
				"  }\n" +
				"}\n",
		},
		{
			"if with a break inside a while",
			whileStatement(booleanLiteral(true), block(ifStatement(booleanLiteral(false), block(breakStatement()), block(continueStatement())))),
			"while (TRUE<True>) {\n" +
				"  if (FALSE<False>) {\n" +
				"    break;\n" +
				"  } else {\n" +
				"    continue;\n" +
				"  }\n" +
				"}\n",
		},
		{
			"while inside a block",
			block(whileStatement(booleanLiteral(true), block(printStatement(1)))),
			"{\n" +
				"  while (TRUE<True>) {\n" +
				"    PRINT INTEGER<1>;\n" +
				"  }\n" +
				"}\n",
		},
	})
}

func TestBreakAndContinueStatementString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{"break", breakStatement(), "break;\n"},
		{"continue", continueStatement(), "continue;\n"},
		{
			"break and continue inside a block",
			block(breakStatement(), continueStatement()),
			"{\n" +
				"  break;\n" +
				"  continue;\n" +
				"}\n",
		},
	})
}
