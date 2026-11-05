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

type methodSignatureTestCase struct {
	name        string
	declaration common.MethodDeclaration
	expected    types.Signature
}

func methodDeclaration(name string, parameters []common.Parameter, returnType types.Type) common.MethodDeclaration {
	return common.NewMethodDeclaration(token(common.FUNC, "func"), token(common.IDENTIFIER, name), parameters, returnType)
}

func serviceDeclaration(name string, methods ...common.MethodDeclaration) *common.ServiceDeclarationStatement {
	return common.NewServiceDeclarationStatement(
		token(common.SERVICE, "service"),
		token(common.IDENTIFIER, name),
		append([]common.MethodDeclaration{}, methods...),
	)
}

func serverDeclaration(name string, address string, body ...common.Statement) *common.ServerDeclarationStatement {
	return common.NewServerDeclarationStatement(
		token(common.SERVER, "server"),
		token(common.IDENTIFIER, name),
		token(common.STRING, `"`+address+`"`),
		address,
		append([]common.Statement{}, body...),
	)
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
		{"integer literal", common.NewPrintStatement(integerLiteral(3)), "print INTEGER<3>;\n"},
		{"string literal", common.NewPrintStatement(stringLiteral("hola")), `print STRING<"hola">;` + "\n"},
		{
			"binary expression",
			common.NewPrintStatement(binary(integerLiteral(1), common.PLUS, "+", integerLiteral(2))),
			"print (INTEGER<1> PLUS<+> INTEGER<2>);\n",
		},
		{
			"grouped negation",
			common.NewPrintStatement(grouping(negation(floatLiteral(2.5)))),
			"print ((MINUS<-> FLOAT<2.5>));\n",
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
				"  print INTEGER<1>;\n" +
				"  INTEGER<2>;\n" +
				"}\n",
		},
		{
			"nested blocks accumulate the indentation",
			block(printStatement(1), block(printStatement(2), block(printStatement(3)))),
			"{\n" +
				"  print INTEGER<1>;\n" +
				"  {\n" +
				"    print INTEGER<2>;\n" +
				"    {\n" +
				"      print INTEGER<3>;\n" +
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
				"  print INTEGER<1>;\n" +
				"}\n",
		},
		{
			"else on the same line as the closed brace",
			ifStatement(booleanLiteral(false), block(printStatement(1)), block(printStatement(2))),
			"if (FALSE<False>) {\n" +
				"  print INTEGER<1>;\n" +
				"} else {\n" +
				"  print INTEGER<2>;\n" +
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
				"  print INTEGER<1>;\n" +
				"} else if (TRUE<True>) {\n" +
				"  print INTEGER<2>;\n" +
				"} else {\n" +
				"  print INTEGER<3>;\n" +
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
				"  print INTEGER<1>;\n" +
				"  if (FALSE<False>) {\n" +
				"    print INTEGER<2>;\n" +
				"  } else {\n" +
				"    print INTEGER<3>;\n" +
				"  }\n" +
				"}\n",
		},
		{
			"if inside a block",
			block(ifStatement(booleanLiteral(true), block(printStatement(1)), nil)),
			"{\n" +
				"  if (TRUE<True>) {\n" +
				"    print INTEGER<1>;\n" +
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
				"  print INTEGER<1>;\n" +
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
				"  print INTEGER<1>;\n" +
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
				"    print INTEGER<1>;\n" +
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
				"  print x;\n" +
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
		{"print statement with a variable", common.NewPrintStatement(variable("x")), "print x;\n"},
		{
			"print statement with an assignment",
			common.NewPrintStatement(assignment("x", assignment("y", integerLiteral(1)))),
			"print (x = (y = INTEGER<1>));\n",
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
				"  print STRING<\"hola\">;\n" +
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
		{
			"named types",
			funcDeclaration("f", []common.Parameter{parameter("p", types.NamedType{Name: "Punto"})}, types.NamedType{Name: "Linea"}),
			types.Signature{Params: []types.Type{types.NamedType{Name: "Punto"}}, Return: types.NamedType{Name: "Linea"}},
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

func fieldDeclaration(name string, fieldType types.Type) common.FieldDeclaration {
	return common.NewFieldDeclaration(token(common.IDENTIFIER, name), fieldType)
}

func structDeclaration(name string, fields ...common.FieldDeclaration) *common.StructDeclarationStatement {
	return common.NewStructDeclarationStatement(
		token(common.STRUCT, "struct"),
		token(common.IDENTIFIER, name),
		append([]common.FieldDeclaration{}, fields...),
	)
}

func TestStructDeclarationStatementString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{"empty struct", structDeclaration("Vacio"), "struct Vacio {\n}\n"},
		{
			"struct with one field",
			structDeclaration("Edad", fieldDeclaration("valor", types.Int)),
			"struct Edad {\n" +
				"  valor : Int;\n" +
				"}\n",
		},
		{
			"struct with fields of every primitive type",
			structDeclaration("Dato",
				fieldDeclaration("a", types.Int),
				fieldDeclaration("b", types.Float),
				fieldDeclaration("c", types.Str),
				fieldDeclaration("d", types.Bool),
			),
			"struct Dato {\n" +
				"  a : Int;\n" +
				"  b : Float;\n" +
				"  c : String;\n" +
				"  d : Bool;\n" +
				"}\n",
		},
		{
			"struct with a field of a named type",
			structDeclaration("Linea", fieldDeclaration("desde", types.NamedType{Name: "Punto"})),
			"struct Linea {\n" +
				"  desde : Punto;\n" +
				"}\n",
		},
		{
			"struct inside a block",
			block(structDeclaration("Punto", fieldDeclaration("x", types.Int))),
			"{\n" +
				"  struct Punto {\n" +
				"    x : Int;\n" +
				"  }\n" +
				"}\n",
		},
		{
			"struct inside a function",
			funcDeclaration("f", nil, types.Void, structDeclaration("Punto", fieldDeclaration("x", types.Int))),
			"func f() -> Void {\n" +
				"  struct Punto {\n" +
				"    x : Int;\n" +
				"  }\n" +
				"}\n",
		},
	})
}

func TestNamedTypeString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{
			"variable of a named type",
			varDeclaration("p", types.NamedType{Name: "Punto"}, variable("q")),
			"let p : Punto = q;\n",
		},
		{
			"parameter and return of named types",
			funcDeclaration("f", []common.Parameter{parameter("p", types.NamedType{Name: "Punto"})}, types.NamedType{Name: "Linea"}),
			"func f(p : Punto) -> Linea {\n}\n",
		},
	})
}

func variantDeclaration(name string, fields ...common.FieldDeclaration) common.VariantDeclaration {
	return common.NewVariantDeclaration(token(common.IDENTIFIER, name), append([]common.FieldDeclaration{}, fields...))
}

func enumDeclaration(name string, variants ...common.VariantDeclaration) *common.EnumDeclarationStatement {
	return common.NewEnumDeclarationStatement(
		token(common.ENUM, "enum"),
		token(common.IDENTIFIER, name),
		append([]common.VariantDeclaration{}, variants...),
	)
}

func TestEnumDeclarationStatementString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{"enum without variants", enumDeclaration("Vacio"), "enum Vacio {\n}\n"},
		{
			"variant without fields",
			enumDeclaration("Color", variantDeclaration("Rojo")),
			"enum Color {\n" +
				"  Rojo;\n" +
				"}\n",
		},
		{
			"variant with one field",
			enumDeclaration("Resultado", variantDeclaration("Ok", fieldDeclaration("valor", types.Int))),
			"enum Resultado {\n" +
				"  Ok(valor : Int);\n" +
				"}\n",
		},
		{
			"variant with fields of every primitive type",
			enumDeclaration("Dato", variantDeclaration("Todo",
				fieldDeclaration("a", types.Int),
				fieldDeclaration("b", types.Float),
				fieldDeclaration("c", types.Str),
				fieldDeclaration("d", types.Bool),
			)),
			"enum Dato {\n" +
				"  Todo(a : Int, b : Float, c : String, d : Bool);\n" +
				"}\n",
		},
		{
			"variants with and without fields",
			enumDeclaration("ResultadoRetiro",
				variantDeclaration("Ok", fieldDeclaration("mov", types.NamedType{Name: "Movimiento"})),
				variantDeclaration("SaldoInsuficiente", fieldDeclaration("disponible", types.Int)),
				variantDeclaration("Cerrada"),
			),
			"enum ResultadoRetiro {\n" +
				"  Ok(mov : Movimiento);\n" +
				"  SaldoInsuficiente(disponible : Int);\n" +
				"  Cerrada;\n" +
				"}\n",
		},
		{
			"recursive enum",
			enumDeclaration("Lista",
				variantDeclaration("Vacia"),
				variantDeclaration("Nodo", fieldDeclaration("valor", types.Int), fieldDeclaration("resto", types.NamedType{Name: "Lista"})),
			),
			"enum Lista {\n" +
				"  Vacia;\n" +
				"  Nodo(valor : Int, resto : Lista);\n" +
				"}\n",
		},
		{
			"enum inside a block",
			block(enumDeclaration("Color", variantDeclaration("Rojo"))),
			"{\n" +
				"  enum Color {\n" +
				"    Rojo;\n" +
				"  }\n" +
				"}\n",
		},
		{
			"enum inside a function",
			funcDeclaration("f", nil, types.Void, enumDeclaration("Color", variantDeclaration("Rojo"))),
			"func f() -> Void {\n" +
				"  enum Color {\n" +
				"    Rojo;\n" +
				"  }\n" +
				"}\n",
		},
	})
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

func TestMatchStatementString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{"match without arms", matchStatement(variable("r")), "match (r) {\n}\n"},
		{
			"variant without fields",
			matchStatement(variable("c"), matchArm(pattern("Rojo"), printStatement(1))),
			"match (c) {\n" +
				"  Rojo => print INTEGER<1>;\n" +
				"}\n",
		},
		{
			"variant with bindings",
			matchStatement(variable("l"),
				matchArm(pattern("Nodo", "valor", "resto"), common.NewPrintStatement(variable("valor"))),
			),
			"match (l) {\n" +
				"  Nodo(valor, resto) => print valor;\n" +
				"}\n",
		},
		{
			"ignored binding",
			matchStatement(variable("l"), matchArm(pattern("Nodo", "v", "_"), printStatement(1))),
			"match (l) {\n" +
				"  Nodo(v, _) => print INTEGER<1>;\n" +
				"}\n",
		},
		{
			"wildcard arm",
			matchStatement(variable("r"),
				matchArm(pattern("Ok", "mov"), common.NewPrintStatement(fieldAccess(variable("mov"), "saldoFinal"))),
				matchArm(wildcard(), printStatement(0)),
			),
			"match (r) {\n" +
				"  Ok(mov) => print mov.saldoFinal;\n" +
				"  _ => print INTEGER<0>;\n" +
				"}\n",
		},
		{
			"block arm",
			matchStatement(variable("r"),
				matchArm(pattern("Ok", "m"), block(printStatement(1), returnStatement(booleanLiteral(true)))),
				matchArm(wildcard(), block()),
			),
			"match (r) {\n" +
				"  Ok(m) => {\n" +
				"    print INTEGER<1>;\n" +
				"    return TRUE<True>;\n" +
				"  }\n" +
				"  _ => {\n" +
				"  }\n" +
				"}\n",
		},
		{
			"call as the matched expression",
			matchStatement(memberCall(variable("c"), "retirar", variable("monto")), matchArm(wildcard(), breakStatement())),
			"match (c.retirar(monto)) {\n" +
				"  _ => break;\n" +
				"}\n",
		},
		{
			"nested match",
			matchStatement(variable("a"),
				matchArm(pattern("Nodo", "_", "resto"), matchStatement(variable("resto"), matchArm(wildcard(), printStatement(2)))),
			),
			"match (a) {\n" +
				"  Nodo(_, resto) => match (resto) {\n" +
				"    _ => print INTEGER<2>;\n" +
				"  }\n" +
				"}\n",
		},
		{
			"match inside a while",
			whileStatement(booleanLiteral(true), block(matchStatement(variable("r"), matchArm(wildcard(), continueStatement())))),
			"while (TRUE<True>) {\n" +
				"  match (r) {\n" +
				"    _ => continue;\n" +
				"  }\n" +
				"}\n",
		},
	})
}

func importStatement(lexeme string, path string) *common.ImportStatement {
	return common.NewImportStatement(token(common.IMPORT, "import"), token(common.STRING, lexeme), path)
}

func TestImportStatementString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{"import of a file", importStatement(`"cuenta.ric"`, "cuenta.ric"), "import \"cuenta.ric\";\n"},
		{"import of a relative path", importStatement(`"../lib/cuenta.ric"`, "../lib/cuenta.ric"), "import \"../lib/cuenta.ric\";\n"},
		{"import keeps the escape sequences of the path", importStatement(`"a\\b.ric"`, `a\b.ric`), "import \"a\\\\b.ric\";\n"},
	})
}

type bodyStatementsTestCase struct {
	name     string
	arm      common.MatchArm
	expected []common.Statement
}

func TestMatchArmBodyStatements(t *testing.T) {
	single, first, second := printStatement(1), printStatement(2), printStatement(3)
	inner := block(printStatement(4))
	testCases := []bodyStatementsTestCase{
		{"single statement", matchArm(wildcard(), single), []common.Statement{single}},
		{"block", matchArm(wildcard(), block(first, second)), []common.Statement{first, second}},
		{"empty block", matchArm(wildcard(), block()), []common.Statement{}},
		{"only the outer block is unwrapped", matchArm(wildcard(), block(inner)), []common.Statement{inner}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.arm.BodyStatements()

			if len(result) != len(testCase.expected) {
				t.Fatalf("BodyStatements() = %v; want %v", result, testCase.expected)
			}
			for i := range result {
				if result[i] != testCase.expected[i] {
					t.Errorf("BodyStatements()[%d] = %v; want %v", i, result[i], testCase.expected[i])
				}
			}
		})
	}
}

func TestServiceDeclarationStatementString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{"service without methods", serviceDeclaration("Empty"), "service Empty {\n}\n"},
		{
			"method without parameters or return value",
			serviceDeclaration("Account", methodDeclaration("reset", nil, types.Void)),
			"service Account {\n" +
				"  func reset() -> Void;\n" +
				"}\n",
		},
		{
			"several methods",
			serviceDeclaration("Account",
				methodDeclaration("balance", nil, types.Int),
				methodDeclaration("deposit", []common.Parameter{parameter("amount", types.Int)}, types.NamedType{Name: "Movement"}),
				methodDeclaration("transfer",
					[]common.Parameter{parameter("target", types.Str), parameter("amount", types.Float), parameter("urgent", types.Bool)},
					types.Bool),
			),
			"service Account {\n" +
				"  func balance() -> Int;\n" +
				"  func deposit(amount : Int) -> Movement;\n" +
				"  func transfer(target : String, amount : Float, urgent : Bool) -> Bool;\n" +
				"}\n",
		},
		{
			"service inside a block",
			block(serviceDeclaration("S", methodDeclaration("f", nil, types.Void))),
			"{\n" +
				"  service S {\n" +
				"    func f() -> Void;\n" +
				"  }\n" +
				"}\n",
		},
	})
}

func TestMethodDeclarationString(t *testing.T) {
	method := methodDeclaration("withdraw", []common.Parameter{parameter("amount", types.Int)}, types.NamedType{Name: "Result"})

	result := method.String()

	expected := "func withdraw(amount : Int) -> Result"
	if result != expected {
		t.Errorf("String() = %q; want %q", result, expected)
	}
}

func TestMethodDeclarationSignature(t *testing.T) {
	testCases := []methodSignatureTestCase{
		{"no parameters", methodDeclaration("f", nil, types.Void), types.Signature{Params: []types.Type{}, Return: types.Void}},
		{
			"several parameters",
			methodDeclaration("f", []common.Parameter{parameter("a", types.Int), parameter("b", types.Str)}, types.Bool),
			types.Signature{Params: []types.Type{types.Int, types.Str}, Return: types.Bool},
		},
		{
			"named types",
			methodDeclaration("f", []common.Parameter{parameter("p", types.NamedType{Name: "Point"})}, types.NamedType{Name: "Line"}),
			types.Signature{Params: []types.Type{types.NamedType{Name: "Point"}}, Return: types.NamedType{Name: "Line"}},
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

func TestServerDeclarationStatementString(t *testing.T) {
	runStatementStringTestCases(t, []statementStringTestCase{
		{"server with empty body", serverDeclaration("Account", "0.0.0.0:3825"), "server Account at \"0.0.0.0:3825\" {\n}\n"},
		{
			"server with state and functions",
			serverDeclaration("Account", "localhost:3825",
				varDeclaration("current", types.Int, integerLiteral(0)),
				funcDeclaration("balance", nil, types.Int, returnStatement(variable("current"))),
			),
			"server Account at \"localhost:3825\" {\n" +
				"  let current : Int = INTEGER<0>;\n" +
				"  func balance() -> Int {\n" +
				"    return current;\n" +
				"  }\n" +
				"}\n",
		},
		{
			"address with escapes is shown as written",
			common.NewServerDeclarationStatement(
				token(common.SERVER, "server"),
				token(common.IDENTIFIER, "S"),
				token(common.STRING, `"a\"b:1"`),
				`a"b:1`,
				[]common.Statement{},
			),
			"server S at \"a\\\"b:1\" {\n}\n",
		},
		{
			"server inside a block",
			block(serverDeclaration("S", "a:1", printStatement(1))),
			"{\n" +
				"  server S at \"a:1\" {\n" +
				"    print INTEGER<1>;\n" +
				"  }\n" +
				"}\n",
		},
	})
}
