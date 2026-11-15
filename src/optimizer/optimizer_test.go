package optimizer_test

import (
	"strings"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/optimizer"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/parser"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/scanner"
)

type optimizationTestCase struct {
	name     string
	source   string
	expected string
}

func optimize(t *testing.T, source string) string {
	t.Helper()

	tokens, err := scanner.NewScanner(source, "test.ric").Scan()
	if err != nil {
		t.Fatalf("Scan() unexpected error: %v", err)
	}
	statements, err := parser.NewParser(tokens).Parse()
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	var output strings.Builder
	for _, statement := range optimizer.Optimize(statements) {
		output.WriteString(statement.String())
	}
	return output.String()
}

func runOptimizationTestCases(t *testing.T, testCases []optimizationTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := optimize(t, testCase.source)
			if got != testCase.expected {
				t.Errorf("Optimize(%q) =\n%s\nwant:\n%s", testCase.source, got, testCase.expected)
			}
		})
	}
}

func TestFoldsConstantExpressions(t *testing.T) {
	runOptimizationTestCases(t, []optimizationTestCase{
		{"integer arithmetic", "print 1 + 2 * 3;", "print INTEGER<7>;\n"},
		{"division always gives a float", "print 7 / 2;", "print FLOAT<3.5>;\n"},
		{"floor division rounds down", "print -7 // 2;", "print INTEGER<-4>;\n"},
		{"modulo takes the sign of the divisor", "print -7 % 3;", "print INTEGER<2>;\n"},
		{"integer power", "print 2 ** 10;", "print INTEGER<1024>;\n"},
		{"integer and float", "print 1 + 0.5;", "print FLOAT<1.5>;\n"},
		{"integer overflow wraps around", "print 9223372036854775807 + 1;", "print INTEGER<-9223372036854775808>;\n"},
		{"negative literal", "print -5;", "print INTEGER<-5>;\n"},
		{"double negation", "print - -5;", "print INTEGER<5>;\n"},
		{"logical not", "print not True;", "print FALSE<False>;\n"},
		{"string concatenation", `print "con" + "cat";`, "print STRING<\"concat\">;\n"},
		{"number comparison", "print 1 < 2;", "print TRUE<True>;\n"},
		{"string equality", `print "a" == "b";`, "print FALSE<False>;\n"},
		{"grouped subexpression", "print (1 + 2) * 3;", "print INTEGER<9>;\n"},
		{"nested grouping of a literal", "print ((4));", "print INTEGER<4>;\n"},
	})
}

func TestFoldsLogicalOperatorsWithLiteralLeftOperand(t *testing.T) {
	runOptimizationTestCases(t, []optimizationTestCase{
		{"false and short-circuits", "print False and flag;", "print FALSE<False>;\n"},
		{"true or short-circuits", "print True or flag;", "print TRUE<True>;\n"},
		{"true and gives the right operand", "print True and flag;", "print flag;\n"},
		{"false or gives the right operand", "print False or flag;", "print flag;\n"},
		{"both operands literal", "print True and False;", "print FALSE<False>;\n"},
		{"variable left operand is kept", "print flag and False;", "print (flag AND<and> FALSE<False>);\n"},
	})
}

func TestKeepsOperationsThatFailAtRuntime(t *testing.T) {
	runOptimizationTestCases(t, []optimizationTestCase{
		{"division by zero", "print 1 / 0;", "print (INTEGER<1> SLASH</> INTEGER<0>);\n"},
		{"floor division by zero", "print 1 // 0;", "print (INTEGER<1> DOUBLE_SLASH<//> INTEGER<0>);\n"},
		{"modulo by zero", "print 1 % 0;", "print (INTEGER<1> PERCENTAGE<%> INTEGER<0>);\n"},
		{"negative integer power", "print 2 ** -1;", "print (INTEGER<2> DOUBLE_STAR<**> INTEGER<-1>);\n"},
		{"divisor folded to zero", "print 1 / (2 - 2);", "print (INTEGER<1> SLASH</> INTEGER<0>);\n"},
		{"short-circuit skips the failing operand", "print False and 1 / 0 == 1;", "print FALSE<False>;\n"},
	})
}

func TestKeepsExpressionsThatAreNotConstant(t *testing.T) {
	runOptimizationTestCases(t, []optimizationTestCase{
		{"constant operand of a variable", "print x + 2 * 3;", "print (x PLUS<+> INTEGER<6>);\n"},
		{"operations are not reassociated", "print x + 1 + 2;", "print ((x PLUS<+> INTEGER<1>) PLUS<+> INTEGER<2>);\n"},
		{"call", "print f();", "print f();\n"},
		{"input", "print input();", "print input();\n"},
	})
}

func TestFoldsInsideEveryStatement(t *testing.T) {
	runOptimizationTestCases(t, []optimizationTestCase{
		{"variable declaration", "let x: Int = 1 + 1;", "let x : Int = INTEGER<2>;\n"},
		{"assignment", "x = 2 * 2;", "(x = INTEGER<4>);\n"},
		{
			"if condition and branches",
			"if (x < 1 + 1) { print 1 + 1; } else { print 2 + 2; }",
			"if ((x LESS<<> INTEGER<2>)) {\n  print INTEGER<2>;\n} else {\n  print INTEGER<4>;\n}\n",
		},
		{
			"while condition and body",
			"while (x < 2 * 5) { x = x + 1 * 1; }",
			"while ((x LESS<<> INTEGER<10>)) {\n  (x = (x PLUS<+> INTEGER<1>));\n}\n",
		},
		{
			"function body and return",
			"func six() -> Int { return 2 * 3; }",
			"func six() -> Int {\n  return INTEGER<6>;\n}\n",
		},
		{"call arguments", "print add(1 + 1, 2 * 2);", "print add(INTEGER<2>, INTEGER<4>);\n"},
		{
			"struct literal fields",
			"print Point { x: 1 + 1, y: 2 * 2 };",
			"print Point { x : INTEGER<2>, y : INTEGER<4> };\n",
		},
		{"member call arguments", "print account.deposit(50 * 2);", "print account.deposit(INTEGER<100>);\n"},
		{
			"match arms",
			"match (shape) { Circle(radius) => print 2 * 3; _ => { print 1 + 1; } }",
			"match (shape) {\n  Circle(radius) => print INTEGER<6>;\n  _ => {\n    print INTEGER<2>;\n  }\n}\n",
		},
		{
			"server functions",
			`server Account at "0.0.0.0:3825" { func balance() -> Int { return 10 * 10; } }`,
			"server Account at \"0.0.0.0:3825\" {\n  func balance() -> Int {\n    return INTEGER<100>;\n  }\n}\n",
		},
		{
			"session timeout, body and rescue",
			"session link timeout 2 * 1000 as account { print 1 + 1; } rescue error { print 2 + 2; }",
			"session link timeout INTEGER<2000> as account {\n  print INTEGER<2>;\n} rescue error {\n  print INTEGER<4>;\n}\n",
		},
	})
}

func TestEliminatesBranchesWithLiteralConditions(t *testing.T) {
	runOptimizationTestCases(t, []optimizationTestCase{
		{"false if without else", "if (False) { print 1; }", ""},
		{"condition folded to false", "if (1 > 2) { print 1; }", ""},
		{"false if keeps the else branch", "if (1 > 2) { print 1; } else { print 2; }", "{\n  print INTEGER<2>;\n}\n"},
		{"true if keeps the if branch", "if (2 * 3 == 6) { print 1; } else { print 2; }", "{\n  print INTEGER<1>;\n}\n"},
		{
			"false else if is removed",
			"if (x) { print 1; } else if (False) { print 2; }",
			"if (x) {\n  print INTEGER<1>;\n}\n",
		},
		{
			"true else if replaces the remaining chain",
			"if (x) { print 1; } else if (True) { print 2; } else { print 3; }",
			"if (x) {\n  print INTEGER<1>;\n} else {\n  print INTEGER<2>;\n}\n",
		},
		{"false while", "while (False) { print 1; }", ""},
		{"while condition folded to false", "while (1 > 2) { print 1; }", ""},
		{"true while is kept", "while (True) { break; }", "while (TRUE<True>) {\n  break;\n}\n"},
		{
			"variable condition is kept",
			"if (x) { print 1; } else { print 2; }",
			"if (x) {\n  print INTEGER<1>;\n} else {\n  print INTEGER<2>;\n}\n",
		},
		{"removed branch inside a while body", "while (x) { if (False) { print 1; } }", "while (x) {\n}\n"},
	})
}

func TestEliminatesStatementsAfterControlFlowJumps(t *testing.T) {
	runOptimizationTestCases(t, []optimizationTestCase{
		{
			"after return",
			"func one() -> Int { return 1; print 2; return 3; }",
			"func one() -> Int {\n  return INTEGER<1>;\n}\n",
		},
		{"after break", "while (x) { break; print 1; }", "while (x) {\n  break;\n}\n"},
		{"after continue", "while (x) { continue; x = False; }", "while (x) {\n  continue;\n}\n"},
		{
			"only in the same block",
			"while (x) { if (y) { break; } print 1; }",
			"while (x) {\n  if (y) {\n    break;\n  }\n  print INTEGER<1>;\n}\n",
		},
		{
			"after a return that comes from a true branch",
			"func one() -> Int { if (True) { return 1; } return 2; }",
			"func one() -> Int {\n  {\n    return INTEGER<1>;\n  }\n  return INTEGER<2>;\n}\n",
		},
	})
}

func TestEliminatesDeadCodeInsideEveryStatement(t *testing.T) {
	runOptimizationTestCases(t, []optimizationTestCase{
		{
			"function body",
			"func one() -> Int { if (False) { return 0; } return 1; }",
			"func one() -> Int {\n  return INTEGER<1>;\n}\n",
		},
		{
			"match arm with a single statement",
			"match (shape) { _ => if (False) { print 1; } }",
			"match (shape) {\n  _ => {\n  }\n}\n",
		},
		{
			"server functions",
			`server Account at "0.0.0.0:3825" { func balance() -> Int { while (False) { print 1; } return 1; } }`,
			"server Account at \"0.0.0.0:3825\" {\n  func balance() -> Int {\n    return INTEGER<1>;\n  }\n}\n",
		},
		{
			"session body and rescue",
			"session link as account { if (False) { print 1; } print 2; } rescue error { return; print 3; }",
			"session link as account {\n  print INTEGER<2>;\n} rescue error {\n  return;\n}\n",
		},
	})
}

func TestRemovesGroupings(t *testing.T) {
	runOptimizationTestCases(t, []optimizationTestCase{
		{"grouping of a variable expression", "print (x + 1);", "print (x PLUS<+> INTEGER<1>);\n"},
		{"nested groupings", "print ((x));", "print x;\n"},
		{"precedence is kept", "print (x + 1) * 2;", "print ((x PLUS<+> INTEGER<1>) STAR<*> INTEGER<2>);\n"},
		{"inside a folded grouping", "print (x * (2 + 3));", "print (x STAR<*> INTEGER<5>);\n"},
		{"unary operand", "print -(x);", "print (MINUS<-> x);\n"},
		{"call argument", "print f((x));", "print f(x);\n"},
		{"field access", "print (p).x;", "print p.x;\n"},
		{
			"condition and assignment",
			"while ((x < 1)) { x = (x + 1); }",
			"while ((x LESS<<> INTEGER<1>)) {\n  (x = (x PLUS<+> INTEGER<1>));\n}\n",
		},
	})
}
