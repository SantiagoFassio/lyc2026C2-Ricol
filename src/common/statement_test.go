package common_test

import (
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
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

func varDeclaration(name string, varType types.Type, valueExpression common.Expression) *common.VarDeclarationStatement {
	return common.NewVarDeclarationStatement(token(common.LET, "let"), token(common.IDENTIFIER, name), varType, valueExpression)
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

func TestVarDeclarationStatementString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{"int variable", varDeclaration("x", types.Int, integerLiteral(1)), "let x : Int = INTEGER<1>;\n"},
		{"float variable", varDeclaration("x", types.Float, floatLiteral(2.5)), "let x : Float = FLOAT<2.5>;\n"},
		{"string variable", varDeclaration("x", types.Str, stringLiteral("a")), "let x : String = STRING<\"a\">;\n"},
		{"bool variable", varDeclaration("x", types.Bool, booleanLiteral(true)), "let x : Bool = TRUE<True>;\n"},
		{
			"value with a variable",
			varDeclaration("y", types.Int, binary(variable("x"), common.MINUS, "-", integerLiteral(1))),
			"let y : Int = (x MINUS<-> INTEGER<1>);\n",
		},
		{
			"value with an assignment",
			varDeclaration("y", types.Int, assignment("x", integerLiteral(1))),
			"let y : Int = (x = INTEGER<1>);\n",
		},
		{
			"declaration inside a block",
			block(varDeclaration("x", types.Int, integerLiteral(1)), common.NewPrintStatement(variable("x"))),
			"{\n" +
				"  let x : Int = INTEGER<1>;\n" +
				"  PRINT x;\n" +
				"}\n",
		},
		{
			"declaration inside a while body",
			whileStatement(variable("running"), block(
				varDeclaration("x", types.Str, stringLiteral("a")),
				common.NewExpressionStatement(assignment("running", booleanLiteral(false))),
			)),
			"while (running) {\n" +
				"  let x : String = STRING<\"a\">;\n" +
				"  (running = FALSE<False>);\n" +
				"}\n",
		},
	})
}

func TestVariableStatementString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{"expression statement with a variable", common.NewExpressionStatement(variable("x")), "x;\n"},
		{"expression statement with an assignment", common.NewExpressionStatement(assignment("x", integerLiteral(2))), "(x = INTEGER<2>);\n"},
		{"print statement with a variable", common.NewPrintStatement(variable("x")), "PRINT x;\n"},
		{
			"print statement with an assignment",
			common.NewPrintStatement(assignment("x", assignment("y", integerLiteral(1)))),
			"PRINT (x = (y = INTEGER<1>));\n",
		},
		{
			"if with a variable condition",
			ifStatement(variable("x"), block(common.NewExpressionStatement(assignment("x", booleanLiteral(false)))), nil),
			"if (x) {\n" +
				"  (x = FALSE<False>);\n" +
				"}\n",
		},
	})
}

func parameter(name string, paramType types.Type) common.Parameter {
	return common.NewParameter(token(common.IDENTIFIER, name), paramType)
}

func funcDeclaration(
	name string,
	parameters []common.Parameter,
	returnType types.Type,
	body ...common.Statement,
) *common.FuncDeclarationStatement {
	return common.NewFuncDeclarationStatement(
		token(common.FUNC, "func"),
		token(common.IDENTIFIER, name),
		parameters,
		returnType,
		append([]common.Statement{}, body...),
	)
}

func returnStatement(valueExpression common.Expression) *common.ReturnStatement {
	return common.NewReturnStatement(token(common.RETURN, "return"), valueExpression)
}

func TestFuncDeclarationStatementString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{"empty function", funcDeclaration("f", nil, types.Void), "func f() -> Void {\n}\n"},
		{
			"function with one parameter",
			funcDeclaration("doble", []common.Parameter{parameter("x", types.Int)}, types.Int,
				returnStatement(binary(variable("x"), common.STAR, "*", integerLiteral(2)))),
			"func doble(x : Int) -> Int {\n" +
				"  return (x STAR<*> INTEGER<2>);\n" +
				"}\n",
		},
		{
			"function with several parameters",
			funcDeclaration("f", []common.Parameter{parameter("a", types.Str), parameter("b", types.Float), parameter("c", types.Bool)},
				types.Bool, returnStatement(variable("c"))),
			"func f(a : String, b : Float, c : Bool) -> Bool {\n" +
				"  return c;\n" +
				"}\n",
		},
		{
			"function without return value",
			funcDeclaration("saludar", nil, types.Void, common.NewPrintStatement(stringLiteral("hola")), returnStatement(nil)),
			"func saludar() -> Void {\n" +
				"  PRINT STRING<\"hola\">;\n" +
				"  return;\n" +
				"}\n",
		},
		{
			"function with nested blocks",
			funcDeclaration("f", []common.Parameter{parameter("x", types.Bool)}, types.Int,
				ifStatement(variable("x"), block(returnStatement(integerLiteral(1))), nil),
				returnStatement(integerLiteral(0))),
			"func f(x : Bool) -> Int {\n" +
				"  if (x) {\n" +
				"    return INTEGER<1>;\n" +
				"  }\n" +
				"  return INTEGER<0>;\n" +
				"}\n",
		},
		{
			"function inside a block",
			block(funcDeclaration("f", nil, types.Void)),
			"{\n" +
				"  func f() -> Void {\n" +
				"  }\n" +
				"}\n",
		},
	})
}

func TestReturnStatementString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{"return without value", returnStatement(nil), "return;\n"},
		{"return with a literal", returnStatement(integerLiteral(1)), "return INTEGER<1>;\n"},
		{"return with a call", returnStatement(call("f", variable("x"))), "return f(x);\n"},
		{"call as an expression statement", common.NewExpressionStatement(call("f")), "f();\n"},
	})
}

type signatureTestCase struct {
	name        string
	declaration *common.FuncDeclarationStatement
	expected    types.Signature
}

func TestFuncDeclarationSignature(t *testing.T) {
	testCases := []signatureTestCase{
		{"no parameters", funcDeclaration("f", nil, types.Void), types.Signature{Params: []types.Type{}, Return: types.Void}},
		{
			"several parameters",
			funcDeclaration("f", []common.Parameter{parameter("a", types.Int), parameter("b", types.Str)}, types.Bool),
			types.Signature{Params: []types.Type{types.Int, types.Str}, Return: types.Bool},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.declaration.Signature()

			if result.String() != testCase.expected.String() || len(result.Params) != len(testCase.expected.Params) {
				t.Errorf("Signature() = %v; want %v", result, testCase.expected)
			}
		})
	}
}
