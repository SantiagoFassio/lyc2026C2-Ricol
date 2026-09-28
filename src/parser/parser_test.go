package parser_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/parser"
)

type parserTestCase struct {
	name               string
	tokens             []common.Token
	expectedStatements []common.Statement
}

type parserErrorTestCase struct {
	name            string
	tokens          []common.Token
	expectedMessage string
}

func token(tokenType common.TokenType, lexeme string) common.Token {
	return common.NewToken(tokenType, lexeme, common.Position{})
}

func tokenAt(tokenType common.TokenType, lexeme string, line int, column int) common.Token {
	return common.NewToken(tokenType, lexeme, common.Position{Line: line, Column: column})
}

func tokens(inputTokens ...common.Token) []common.Token {
	return append(inputTokens, token(common.SEMICOLON, ";"), token(common.EOF, ""))
}

func tokensWithoutSemicolon(inputTokens ...common.Token) []common.Token {
	return append(inputTokens, token(common.EOF, ""))
}

func integerLiteralExpression(lexeme string, value int64) *common.LiteralExpression {
	return common.NewLiteralExpression(token(common.INTEGER, lexeme), types.NewInteger(value))
}

func floatLiteralExpression(lexeme string, value float64) *common.LiteralExpression {
	return common.NewLiteralExpression(token(common.FLOAT, lexeme), types.NewFloat(value))
}

func stringLiteralExpression(lexeme string, value string) *common.LiteralExpression {
	return common.NewLiteralExpression(token(common.STRING, lexeme), types.NewString(value))
}

func booleanLiteralExpression(tokenType common.TokenType, lexeme string, value bool) *common.LiteralExpression {
	return common.NewLiteralExpression(token(tokenType, lexeme), types.NewBoolean(value))
}

func groupingExpression(expression common.Expression) *common.GroupingExpression {
	return common.NewGroupingExpression(token(common.OPEN_PAR, "("), expression)
}

func statements(expressions ...common.Expression) []common.Statement {
	expectedStatements := []common.Statement{}
	for _, expression := range expressions {
		expectedStatements = append(expectedStatements, common.NewExpressionStatement(expression))
	}
	return expectedStatements
}

func assertParse(t *testing.T, tokens []common.Token, expectedStatements []common.Statement) {
	t.Helper()

	parsedStatements, err := parser.NewParser(tokens).Parse()

	if err != nil {
		t.Fatalf("parser.Parse(%q) unexpected error: %v", tokens, err)
	}
	if diff := cmp.Diff(expectedStatements, parsedStatements, cmpopts.EquateComparable(types.Number{}, types.String{}, types.Boolean{}, types.Int)); diff != "" {
		t.Errorf("parser.Parse(%q) mismatch (-want +got):\n%s", tokens, diff)
	}
}

func assertParseError(t *testing.T, inputTokens []common.Token, expectedMessage string) {
	t.Helper()

	parsedStatements, err := parser.NewParser(inputTokens).Parse()

	if err == nil {
		t.Fatalf("parser.Parse(%q) = nil error; want %q", inputTokens, expectedMessage)
	}
	if err.Error() != expectedMessage {
		t.Errorf("parser.Parse(%q) error = %q; want %q", inputTokens, err, expectedMessage)
	}
	if len(parsedStatements) != 0 {
		t.Errorf("parser.Parse(%q) = %v; want no statements", inputTokens, parsedStatements)
	}
}

func runParseTestCases(t *testing.T, testCases []parserTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertParse(t, testCase.tokens, testCase.expectedStatements)
		})
	}
}

func runParseErrorTestCases(t *testing.T, testCases []parserErrorTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertParseError(t, testCase.tokens, testCase.expectedMessage)
		})
	}
}

func blockStatement(statements ...common.Statement) *common.BlockStatement {
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

func varDeclarationStatement(name string, varType types.Type, valueExpression common.Expression) *common.VarDeclarationStatement {
	return common.NewVarDeclarationStatement(token(common.LET, "let"), token(common.IDENTIFIER, name), varType, valueExpression)
}

func variableExpression(name string) *common.VariableExpression {
	return common.NewVariableExpression(token(common.IDENTIFIER, name))
}

func varAssignmentExpression(name string, valueExpression common.Expression) *common.VarAssignmentExpression {
	return common.NewVarAssignmentExpression(token(common.IDENTIFIER, name), valueExpression)
}

func TestNoTokens(t *testing.T) {
	assertParse(t, []common.Token{}, statements())
}

func TestEOFToken(t *testing.T) {
	assertParse(t, tokensWithoutSemicolon(), statements())
}

func TestLiterals(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"integer",
			tokens(token(common.INTEGER, "3")),
			statements(integerLiteralExpression("3", 3)),
		},
		{
			"float",
			tokens(token(common.FLOAT, "8.3")),
			statements(floatLiteralExpression("8.3", 8.3)),
		},
		{
			"string",
			tokens(token(common.STRING, `"hola"`)),
			statements(stringLiteralExpression(`"hola"`, "hola")),
		},
		{
			"true",
			tokens(token(common.TRUE, "True")),
			statements(booleanLiteralExpression(common.TRUE, "True", true)),
		},
		{
			"false",
			tokens(token(common.FALSE, "False")),
			statements(booleanLiteralExpression(common.FALSE, "False", false)),
		},
	})
}

func TestBooleanLiterals(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"grouped boolean",
			tokens(token(common.OPEN_PAR, "("), token(common.TRUE, "True"), token(common.CLOSED_PAR, ")")),
			statements(groupingExpression(booleanLiteralExpression(common.TRUE, "True", true))),
		},
		{
			"negation of a boolean is left to the type checker",
			tokens(token(common.MINUS, "-"), token(common.TRUE, "True")),
			statements(common.NewUnaryExpression(token(common.MINUS, "-"), booleanLiteralExpression(common.TRUE, "True", true))),
		},
		{
			"addition of booleans is left to the type checker",
			tokens(token(common.TRUE, "True"), token(common.PLUS, "+"), token(common.FALSE, "False")),
			statements(common.NewBinaryExpression(
				booleanLiteralExpression(common.TRUE, "True", true),
				token(common.PLUS, "+"),
				booleanLiteralExpression(common.FALSE, "False", false),
			)),
		},
		{
			"boolean mixed with a number",
			tokens(token(common.TRUE, "True"), token(common.PLUS, "+"), token(common.INTEGER, "1")),
			statements(common.NewBinaryExpression(
				booleanLiteralExpression(common.TRUE, "True", true),
				token(common.PLUS, "+"),
				integerLiteralExpression("1", 1),
			)),
		},
	})
}

func TestEqualityExpressions(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"between integers",
			tokens(token(common.INTEGER, "1"), token(common.DOUBLE_EQUAL, "=="), token(common.INTEGER, "2")),
			statements(common.NewBinaryExpression(
				integerLiteralExpression("1", 1),
				token(common.DOUBLE_EQUAL, "=="),
				integerLiteralExpression("2", 2),
			)),
		},
		{
			"between booleans",
			tokens(token(common.TRUE, "True"), token(common.DOUBLE_EQUAL, "=="), token(common.FALSE, "False")),
			statements(common.NewBinaryExpression(
				booleanLiteralExpression(common.TRUE, "True", true),
				token(common.DOUBLE_EQUAL, "=="),
				booleanLiteralExpression(common.FALSE, "False", false),
			)),
		},
		{
			"addition binds tighter than equality",
			tokens(
				token(common.INTEGER, "1"),
				token(common.PLUS, "+"),
				token(common.INTEGER, "2"),
				token(common.DOUBLE_EQUAL, "=="),
				token(common.INTEGER, "3"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2)),
				token(common.DOUBLE_EQUAL, "=="),
				integerLiteralExpression("3", 3),
			)),
		},
		{
			"equality is left associative",
			tokens(
				token(common.INTEGER, "1"),
				token(common.DOUBLE_EQUAL, "=="),
				token(common.INTEGER, "2"),
				token(common.DOUBLE_EQUAL, "=="),
				token(common.INTEGER, "3"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(
					integerLiteralExpression("1", 1),
					token(common.DOUBLE_EQUAL, "=="),
					integerLiteralExpression("2", 2),
				),
				token(common.DOUBLE_EQUAL, "=="),
				integerLiteralExpression("3", 3),
			)),
		},
		{
			"grouped equality",
			tokens(
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.DOUBLE_EQUAL, "=="),
				token(common.INTEGER, "2"),
				token(common.CLOSED_PAR, ")"),
			),
			statements(groupingExpression(common.NewBinaryExpression(
				integerLiteralExpression("1", 1),
				token(common.DOUBLE_EQUAL, "=="),
				integerLiteralExpression("2", 2),
			))),
		},
	})
}

func TestComparisonExpressions(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"less",
			tokens(token(common.INTEGER, "1"), token(common.LESS, "<"), token(common.INTEGER, "2")),
			statements(common.NewBinaryExpression(
				integerLiteralExpression("1", 1),
				token(common.LESS, "<"),
				integerLiteralExpression("2", 2),
			)),
		},
		{
			"greater equal between strings",
			tokens(token(common.STRING, `"a"`), token(common.GREATER_EQUAL, ">="), token(common.STRING, `"b"`)),
			statements(common.NewBinaryExpression(
				stringLiteralExpression(`"a"`, "a"),
				token(common.GREATER_EQUAL, ">="),
				stringLiteralExpression(`"b"`, "b"),
			)),
		},
		{
			"addition binds tighter than comparison",
			tokens(
				token(common.INTEGER, "1"),
				token(common.PLUS, "+"),
				token(common.INTEGER, "2"),
				token(common.LESS, "<"),
				token(common.INTEGER, "4"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2)),
				token(common.LESS, "<"),
				integerLiteralExpression("4", 4),
			)),
		},
		{
			"comparison binds tighter than equality",
			tokens(
				token(common.INTEGER, "1"),
				token(common.LESS, "<"),
				token(common.INTEGER, "2"),
				token(common.DOUBLE_EQUAL, "=="),
				token(common.TRUE, "True"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.LESS, "<"), integerLiteralExpression("2", 2)),
				token(common.DOUBLE_EQUAL, "=="),
				booleanLiteralExpression(common.TRUE, "True", true),
			)),
		},
		{
			"comparison is left associative",
			tokens(
				token(common.INTEGER, "1"),
				token(common.LESS, "<"),
				token(common.INTEGER, "2"),
				token(common.LESS_EQUAL, "<="),
				token(common.INTEGER, "3"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.LESS, "<"), integerLiteralExpression("2", 2)),
				token(common.LESS_EQUAL, "<="),
				integerLiteralExpression("3", 3),
			)),
		},
		{
			"not equal",
			tokens(token(common.INTEGER, "1"), token(common.NOT_EQUAL, "!="), token(common.INTEGER, "2")),
			statements(common.NewBinaryExpression(
				integerLiteralExpression("1", 1),
				token(common.NOT_EQUAL, "!="),
				integerLiteralExpression("2", 2),
			)),
		},
		{
			"comparison of a negated number",
			tokens(
				token(common.MINUS, "-"),
				token(common.INTEGER, "1"),
				token(common.GREATER, ">"),
				token(common.INTEGER, "0"),
			),
			statements(common.NewBinaryExpression(
				common.NewUnaryExpression(token(common.MINUS, "-"), integerLiteralExpression("1", 1)),
				token(common.GREATER, ">"),
				integerLiteralExpression("0", 0),
			)),
		},
	})
}

func TestLogicalExpressions(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"and",
			tokens(token(common.TRUE, "True"), token(common.AND, "and"), token(common.FALSE, "False")),
			statements(common.NewBinaryExpression(
				booleanLiteralExpression(common.TRUE, "True", true),
				token(common.AND, "and"),
				booleanLiteralExpression(common.FALSE, "False", false),
			)),
		},
		{
			"or",
			tokens(token(common.FALSE, "False"), token(common.OR, "or"), token(common.TRUE, "True")),
			statements(common.NewBinaryExpression(
				booleanLiteralExpression(common.FALSE, "False", false),
				token(common.OR, "or"),
				booleanLiteralExpression(common.TRUE, "True", true),
			)),
		},
		{
			"not",
			tokens(token(common.NOT, "not"), token(common.TRUE, "True")),
			statements(common.NewUnaryExpression(token(common.NOT, "not"), booleanLiteralExpression(common.TRUE, "True", true))),
		},
		{
			"double not",
			tokens(token(common.NOT, "not"), token(common.NOT, "not"), token(common.TRUE, "True")),
			statements(common.NewUnaryExpression(
				token(common.NOT, "not"),
				common.NewUnaryExpression(token(common.NOT, "not"), booleanLiteralExpression(common.TRUE, "True", true)),
			)),
		},
		{
			"and binds tighter than or",
			tokens(
				token(common.TRUE, "True"),
				token(common.OR, "or"),
				token(common.FALSE, "False"),
				token(common.AND, "and"),
				token(common.FALSE, "False"),
			),
			statements(common.NewBinaryExpression(
				booleanLiteralExpression(common.TRUE, "True", true),
				token(common.OR, "or"),
				common.NewBinaryExpression(
					booleanLiteralExpression(common.FALSE, "False", false),
					token(common.AND, "and"),
					booleanLiteralExpression(common.FALSE, "False", false),
				),
			)),
		},
		{
			"and binds tighter than or when it comes first",
			tokens(
				token(common.TRUE, "True"),
				token(common.AND, "and"),
				token(common.FALSE, "False"),
				token(common.OR, "or"),
				token(common.TRUE, "True"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(
					booleanLiteralExpression(common.TRUE, "True", true),
					token(common.AND, "and"),
					booleanLiteralExpression(common.FALSE, "False", false),
				),
				token(common.OR, "or"),
				booleanLiteralExpression(common.TRUE, "True", true),
			)),
		},
		{
			"not binds tighter than and",
			tokens(
				token(common.NOT, "not"),
				token(common.TRUE, "True"),
				token(common.AND, "and"),
				token(common.FALSE, "False"),
			),
			statements(common.NewBinaryExpression(
				common.NewUnaryExpression(token(common.NOT, "not"), booleanLiteralExpression(common.TRUE, "True", true)),
				token(common.AND, "and"),
				booleanLiteralExpression(common.FALSE, "False", false),
			)),
		},
		{
			"equality binds tighter than not",
			tokens(
				token(common.NOT, "not"),
				token(common.INTEGER, "1"),
				token(common.DOUBLE_EQUAL, "=="),
				token(common.INTEGER, "2"),
			),
			statements(common.NewUnaryExpression(
				token(common.NOT, "not"),
				common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.DOUBLE_EQUAL, "=="), integerLiteralExpression("2", 2)),
			)),
		},
		{
			"comparisons as operands of an and",
			tokens(
				token(common.INTEGER, "1"),
				token(common.LESS, "<"),
				token(common.INTEGER, "2"),
				token(common.AND, "and"),
				token(common.INTEGER, "3"),
				token(common.LESS, "<"),
				token(common.INTEGER, "4"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.LESS, "<"), integerLiteralExpression("2", 2)),
				token(common.AND, "and"),
				common.NewBinaryExpression(integerLiteralExpression("3", 3), token(common.LESS, "<"), integerLiteralExpression("4", 4)),
			)),
		},
		{
			"or is left associative",
			tokens(
				token(common.TRUE, "True"),
				token(common.OR, "or"),
				token(common.FALSE, "False"),
				token(common.OR, "or"),
				token(common.TRUE, "True"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(
					booleanLiteralExpression(common.TRUE, "True", true),
					token(common.OR, "or"),
					booleanLiteralExpression(common.FALSE, "False", false),
				),
				token(common.OR, "or"),
				booleanLiteralExpression(common.TRUE, "True", true),
			)),
		},
		{
			"and is left associative",
			tokens(
				token(common.TRUE, "True"),
				token(common.AND, "and"),
				token(common.FALSE, "False"),
				token(common.AND, "and"),
				token(common.TRUE, "True"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(
					booleanLiteralExpression(common.TRUE, "True", true),
					token(common.AND, "and"),
					booleanLiteralExpression(common.FALSE, "False", false),
				),
				token(common.AND, "and"),
				booleanLiteralExpression(common.TRUE, "True", true),
			)),
		},
		{
			"grouping changes the precedence",
			tokens(
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.OR, "or"),
				token(common.FALSE, "False"),
				token(common.CLOSED_PAR, ")"),
				token(common.AND, "and"),
				token(common.FALSE, "False"),
			),
			statements(common.NewBinaryExpression(
				groupingExpression(common.NewBinaryExpression(
					booleanLiteralExpression(common.TRUE, "True", true),
					token(common.OR, "or"),
					booleanLiteralExpression(common.FALSE, "False", false),
				)),
				token(common.AND, "and"),
				booleanLiteralExpression(common.FALSE, "False", false),
			)),
		},
		{
			"not of a grouping",
			tokens(
				token(common.NOT, "not"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.AND, "and"),
				token(common.FALSE, "False"),
				token(common.CLOSED_PAR, ")"),
			),
			statements(common.NewUnaryExpression(
				token(common.NOT, "not"),
				groupingExpression(common.NewBinaryExpression(
					booleanLiteralExpression(common.TRUE, "True", true),
					token(common.AND, "and"),
					booleanLiteralExpression(common.FALSE, "False", false),
				)),
			)),
		},
	})
}

func TestLiteralEdgeCases(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"zero integer",
			tokens(token(common.INTEGER, "0")),
			statements(integerLiteralExpression("0", 0)),
		},
		{
			"max int64",
			tokens(token(common.INTEGER, "9223372036854775807")),
			statements(integerLiteralExpression("9223372036854775807", 9223372036854775807)),
		},
		{
			"integer with leading zeros",
			tokens(token(common.INTEGER, "007")),
			statements(integerLiteralExpression("007", 7)),
		},
		{
			"zero float",
			tokens(token(common.FLOAT, "0.0")),
			statements(floatLiteralExpression("0.0", 0.0)),
		},
		{
			"float without decimals",
			tokens(token(common.FLOAT, "8.0")),
			statements(floatLiteralExpression("8.0", 8.0)),
		},
		{
			"float with many decimals",
			tokens(token(common.FLOAT, "3.14159265358979")),
			statements(floatLiteralExpression("3.14159265358979", 3.14159265358979)),
		},
	})
}

func TestStringLiterals(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"empty string",
			tokens(token(common.STRING, `""`)),
			statements(stringLiteralExpression(`""`, "")),
		},
		{
			"comment character and operators are kept",
			tokens(token(common.STRING, `"1 + 2; @ hola"`)),
			statements(stringLiteralExpression(`"1 + 2; @ hola"`, "1 + 2; @ hola")),
		},
		{
			"unicode characters",
			tokens(token(common.STRING, `"ñandú"`)),
			statements(stringLiteralExpression(`"ñandú"`, "ñandú")),
		},
		{
			"escaped double quote",
			tokens(token(common.STRING, `"dijo \"hola\""`)),
			statements(stringLiteralExpression(`"dijo \"hola\""`, `dijo "hola"`)),
		},
		{
			"escaped backslash",
			tokens(token(common.STRING, `"a\\b"`)),
			statements(stringLiteralExpression(`"a\\b"`, `a\b`)),
		},
		{
			"escaped line feed",
			tokens(token(common.STRING, `"a\nb"`)),
			statements(stringLiteralExpression(`"a\nb"`, "a\nb")),
		},
		{
			"escaped tabulation",
			tokens(token(common.STRING, `"a\tb"`)),
			statements(stringLiteralExpression(`"a\tb"`, "a\tb")),
		},
		{
			"escaped backslash before the closing double quote",
			tokens(token(common.STRING, `"a\\"`)),
			statements(stringLiteralExpression(`"a\\"`, `a\`)),
		},
		{
			"escaped backslash followed by an n is not a line feed",
			tokens(token(common.STRING, `"a\\nb"`)),
			statements(stringLiteralExpression(`"a\\nb"`, `a\nb`)),
		},
	})
}

func TestStringOperations(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"concatenation",
			tokens(token(common.STRING, `"a"`), token(common.PLUS, "+"), token(common.STRING, `"b"`)),
			statements(common.NewBinaryExpression(stringLiteralExpression(`"a"`, "a"), token(common.PLUS, "+"), stringLiteralExpression(`"b"`, "b"))),
		},
		{
			"multiplication binds tighter than concatenation",
			tokens(
				token(common.STRING, `"a"`),
				token(common.PLUS, "+"),
				token(common.STRING, `"b"`),
				token(common.STAR, "*"),
				token(common.INTEGER, "2"),
			),
			statements(common.NewBinaryExpression(
				stringLiteralExpression(`"a"`, "a"),
				token(common.PLUS, "+"),
				common.NewBinaryExpression(stringLiteralExpression(`"b"`, "b"), token(common.STAR, "*"), integerLiteralExpression("2", 2)),
			)),
		},
		{
			"grouped concatenation",
			tokens(
				token(common.OPEN_PAR, "("),
				token(common.STRING, `"a"`),
				token(common.PLUS, "+"),
				token(common.STRING, `"b"`),
				token(common.CLOSED_PAR, ")"),
			),
			statements(groupingExpression(
				common.NewBinaryExpression(stringLiteralExpression(`"a"`, "a"), token(common.PLUS, "+"), stringLiteralExpression(`"b"`, "b")),
			)),
		},
		{
			"negation of a string is only rejected when evaluating",
			tokens(token(common.MINUS, "-"), token(common.STRING, `"a"`)),
			statements(common.NewUnaryExpression(token(common.MINUS, "-"), stringLiteralExpression(`"a"`, "a"))),
		},
	})
}

func TestSimpleMathOperations(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"negation",
			tokens(token(common.MINUS, "-"), token(common.FLOAT, "25.11")),
			statements(common.NewUnaryExpression(token(common.MINUS, "-"), floatLiteralExpression("25.11", 25.11))),
		},
		{
			"addition",
			tokens(token(common.INTEGER, "3"), token(common.PLUS, "+"), token(common.FLOAT, "8.5")),
			statements(common.NewBinaryExpression(integerLiteralExpression("3", 3), token(common.PLUS, "+"), floatLiteralExpression("8.5", 8.5))),
		},
		{
			"subtraction",
			tokens(token(common.INTEGER, "3"), token(common.MINUS, "-"), token(common.FLOAT, "8.5")),
			statements(common.NewBinaryExpression(integerLiteralExpression("3", 3), token(common.MINUS, "-"), floatLiteralExpression("8.5", 8.5))),
		},
		{
			"multiplication",
			tokens(token(common.INTEGER, "3"), token(common.STAR, "*"), token(common.FLOAT, "8.5")),
			statements(common.NewBinaryExpression(integerLiteralExpression("3", 3), token(common.STAR, "*"), floatLiteralExpression("8.5", 8.5))),
		},
		{
			"division",
			tokens(token(common.INTEGER, "3"), token(common.SLASH, "/"), token(common.FLOAT, "8.5")),
			statements(common.NewBinaryExpression(integerLiteralExpression("3", 3), token(common.SLASH, "/"), floatLiteralExpression("8.5", 8.5))),
		},
		{
			"floor division",
			tokens(token(common.INTEGER, "3"), token(common.DOUBLE_SLASH, "//"), token(common.FLOAT, "8.5")),
			statements(common.NewBinaryExpression(integerLiteralExpression("3", 3), token(common.DOUBLE_SLASH, "//"), floatLiteralExpression("8.5", 8.5))),
		},
		{
			"modulo",
			tokens(token(common.INTEGER, "3"), token(common.PERCENTAGE, "%"), token(common.FLOAT, "8.5")),
			statements(common.NewBinaryExpression(integerLiteralExpression("3", 3), token(common.PERCENTAGE, "%"), floatLiteralExpression("8.5", 8.5))),
		},
		{
			"exponentiation",
			tokens(token(common.INTEGER, "3"), token(common.DOUBLE_STAR, "**"), token(common.FLOAT, "8.5")),
			statements(common.NewBinaryExpression(integerLiteralExpression("3", 3), token(common.DOUBLE_STAR, "**"), floatLiteralExpression("8.5", 8.5))),
		},
	})
}

func TestOperatorPrecedence(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"multiplication binds tighter than addition",
			tokens(
				token(common.INTEGER, "1"),
				token(common.PLUS, "+"),
				token(common.INTEGER, "2"),
				token(common.STAR, "*"),
				token(common.INTEGER, "3"),
			),
			statements(common.NewBinaryExpression(
				integerLiteralExpression("1", 1),
				token(common.PLUS, "+"),
				common.NewBinaryExpression(integerLiteralExpression("2", 2), token(common.STAR, "*"), integerLiteralExpression("3", 3)),
			)),
		},
		{
			"division binds tighter than subtraction",
			tokens(
				token(common.INTEGER, "8"),
				token(common.SLASH, "/"),
				token(common.INTEGER, "4"),
				token(common.MINUS, "-"),
				token(common.INTEGER, "1"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(integerLiteralExpression("8", 8), token(common.SLASH, "/"), integerLiteralExpression("4", 4)),
				token(common.MINUS, "-"),
				integerLiteralExpression("1", 1),
			)),
		},
		{
			"modulo and multiplication share precedence",
			tokens(
				token(common.INTEGER, "7"),
				token(common.PERCENTAGE, "%"),
				token(common.INTEGER, "4"),
				token(common.STAR, "*"),
				token(common.INTEGER, "2"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(integerLiteralExpression("7", 7), token(common.PERCENTAGE, "%"), integerLiteralExpression("4", 4)),
				token(common.STAR, "*"),
				integerLiteralExpression("2", 2),
			)),
		},
		{
			"floor division and division share precedence",
			tokens(
				token(common.INTEGER, "8"),
				token(common.DOUBLE_SLASH, "//"),
				token(common.INTEGER, "3"),
				token(common.SLASH, "/"),
				token(common.INTEGER, "2"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(integerLiteralExpression("8", 8), token(common.DOUBLE_SLASH, "//"), integerLiteralExpression("3", 3)),
				token(common.SLASH, "/"),
				integerLiteralExpression("2", 2),
			)),
		},
		{
			"exponentiation binds tighter than multiplication",
			tokens(
				token(common.INTEGER, "2"),
				token(common.DOUBLE_STAR, "**"),
				token(common.INTEGER, "3"),
				token(common.STAR, "*"),
				token(common.INTEGER, "4"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(integerLiteralExpression("2", 2), token(common.DOUBLE_STAR, "**"), integerLiteralExpression("3", 3)),
				token(common.STAR, "*"),
				integerLiteralExpression("4", 4),
			)),
		},
		{
			"exponentiation binds tighter than unary minus",
			tokens(
				token(common.MINUS, "-"),
				token(common.INTEGER, "3"),
				token(common.DOUBLE_STAR, "**"),
				token(common.INTEGER, "2"),
			),
			statements(common.NewUnaryExpression(
				token(common.MINUS, "-"),
				common.NewBinaryExpression(integerLiteralExpression("3", 3), token(common.DOUBLE_STAR, "**"), integerLiteralExpression("2", 2)),
			)),
		},
		{
			"unary minus binds tighter than multiplication",
			tokens(
				token(common.MINUS, "-"),
				token(common.INTEGER, "3"),
				token(common.STAR, "*"),
				token(common.INTEGER, "2"),
			),
			statements(common.NewBinaryExpression(
				common.NewUnaryExpression(token(common.MINUS, "-"), integerLiteralExpression("3", 3)),
				token(common.STAR, "*"),
				integerLiteralExpression("2", 2),
			)),
		},
	})
}

func TestOperatorAssociativity(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"addition is left associative",
			tokens(
				token(common.INTEGER, "1"),
				token(common.PLUS, "+"),
				token(common.INTEGER, "2"),
				token(common.PLUS, "+"),
				token(common.INTEGER, "3"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2)),
				token(common.PLUS, "+"),
				integerLiteralExpression("3", 3),
			)),
		},
		{
			"subtraction is left associative",
			tokens(
				token(common.INTEGER, "10"),
				token(common.MINUS, "-"),
				token(common.INTEGER, "3"),
				token(common.MINUS, "-"),
				token(common.INTEGER, "2"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(integerLiteralExpression("10", 10), token(common.MINUS, "-"), integerLiteralExpression("3", 3)),
				token(common.MINUS, "-"),
				integerLiteralExpression("2", 2),
			)),
		},
		{
			"division is left associative",
			tokens(
				token(common.INTEGER, "16"),
				token(common.SLASH, "/"),
				token(common.INTEGER, "4"),
				token(common.SLASH, "/"),
				token(common.INTEGER, "2"),
			),
			statements(common.NewBinaryExpression(
				common.NewBinaryExpression(integerLiteralExpression("16", 16), token(common.SLASH, "/"), integerLiteralExpression("4", 4)),
				token(common.SLASH, "/"),
				integerLiteralExpression("2", 2),
			)),
		},
		{
			"exponentiation is right associative",
			tokens(
				token(common.INTEGER, "2"),
				token(common.DOUBLE_STAR, "**"),
				token(common.INTEGER, "3"),
				token(common.DOUBLE_STAR, "**"),
				token(common.INTEGER, "2"),
			),
			statements(common.NewBinaryExpression(
				integerLiteralExpression("2", 2),
				token(common.DOUBLE_STAR, "**"),
				common.NewBinaryExpression(integerLiteralExpression("3", 3), token(common.DOUBLE_STAR, "**"), integerLiteralExpression("2", 2)),
			)),
		},
	})
}

func TestUnaryExpressions(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"double negation",
			tokens(token(common.MINUS, "-"), token(common.MINUS, "-"), token(common.INTEGER, "3")),
			statements(common.NewUnaryExpression(
				token(common.MINUS, "-"),
				common.NewUnaryExpression(token(common.MINUS, "-"), integerLiteralExpression("3", 3)),
			)),
		},
		{
			"triple negation",
			tokens(
				token(common.MINUS, "-"),
				token(common.MINUS, "-"),
				token(common.MINUS, "-"),
				token(common.FLOAT, "1.5"),
			),
			statements(common.NewUnaryExpression(
				token(common.MINUS, "-"),
				common.NewUnaryExpression(
					token(common.MINUS, "-"),
					common.NewUnaryExpression(token(common.MINUS, "-"), floatLiteralExpression("1.5", 1.5)),
				),
			)),
		},
		{
			"negation of a grouping",
			tokens(
				token(common.MINUS, "-"),
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.PLUS, "+"),
				token(common.INTEGER, "2"),
				token(common.CLOSED_PAR, ")"),
			),
			statements(common.NewUnaryExpression(
				token(common.MINUS, "-"),
				groupingExpression(common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2))),
			)),
		},
		{
			"negation as right operand of a binary operator",
			tokens(
				token(common.INTEGER, "3"),
				token(common.STAR, "*"),
				token(common.MINUS, "-"),
				token(common.INTEGER, "2"),
			),
			statements(common.NewBinaryExpression(
				integerLiteralExpression("3", 3),
				token(common.STAR, "*"),
				common.NewUnaryExpression(token(common.MINUS, "-"), integerLiteralExpression("2", 2)),
			)),
		},
		{
			"negation as exponent",
			tokens(
				token(common.INTEGER, "2"),
				token(common.DOUBLE_STAR, "**"),
				token(common.MINUS, "-"),
				token(common.INTEGER, "3"),
			),
			statements(common.NewBinaryExpression(
				integerLiteralExpression("2", 2),
				token(common.DOUBLE_STAR, "**"),
				common.NewUnaryExpression(token(common.MINUS, "-"), integerLiteralExpression("3", 3)),
			)),
		},
		{
			"negation after addition",
			tokens(
				token(common.INTEGER, "1"),
				token(common.PLUS, "+"),
				token(common.MINUS, "-"),
				token(common.INTEGER, "2"),
			),
			statements(common.NewBinaryExpression(
				integerLiteralExpression("1", 1),
				token(common.PLUS, "+"),
				common.NewUnaryExpression(token(common.MINUS, "-"), integerLiteralExpression("2", 2)),
			)),
		},
	})
}

func TestGroupingExpressions(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"grouped literal",
			tokens(token(common.OPEN_PAR, "("), token(common.INTEGER, "3"), token(common.CLOSED_PAR, ")")),
			statements(groupingExpression(integerLiteralExpression("3", 3))),
		},
		{
			"nested groupings",
			tokens(
				token(common.OPEN_PAR, "("),
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "3"),
				token(common.CLOSED_PAR, ")"),
				token(common.CLOSED_PAR, ")"),
			),
			statements(groupingExpression(groupingExpression(integerLiteralExpression("3", 3)))),
		},
		{
			"grouping overrides precedence",
			tokens(
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.PLUS, "+"),
				token(common.INTEGER, "2"),
				token(common.CLOSED_PAR, ")"),
				token(common.STAR, "*"),
				token(common.INTEGER, "3"),
			),
			statements(common.NewBinaryExpression(
				groupingExpression(common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2))),
				token(common.STAR, "*"),
				integerLiteralExpression("3", 3),
			)),
		},
		{
			"grouping as exponentiation base",
			tokens(
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.PLUS, "+"),
				token(common.INTEGER, "2"),
				token(common.CLOSED_PAR, ")"),
				token(common.DOUBLE_STAR, "**"),
				token(common.INTEGER, "2"),
			),
			statements(common.NewBinaryExpression(
				groupingExpression(common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2))),
				token(common.DOUBLE_STAR, "**"),
				integerLiteralExpression("2", 2),
			)),
		},
		{
			"grouping containing a unary expression",
			tokens(
				token(common.OPEN_PAR, "("),
				token(common.MINUS, "-"),
				token(common.INTEGER, "3"),
				token(common.CLOSED_PAR, ")"),
			),
			statements(groupingExpression(common.NewUnaryExpression(token(common.MINUS, "-"), integerLiteralExpression("3", 3)))),
		},
	})
}

func TestMultipleStatements(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"two literal statements",
			tokensWithoutSemicolon(
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.INTEGER, "2"),
				token(common.SEMICOLON, ";"),
			),
			statements(integerLiteralExpression("1", 1), integerLiteralExpression("2", 2)),
		},
		{
			"three mixed statements",
			tokensWithoutSemicolon(
				token(common.INTEGER, "1"),
				token(common.PLUS, "+"),
				token(common.INTEGER, "2"),
				token(common.SEMICOLON, ";"),
				token(common.OPEN_PAR, "("),
				token(common.FLOAT, "3.5"),
				token(common.CLOSED_PAR, ")"),
				token(common.SEMICOLON, ";"),
				token(common.MINUS, "-"),
				token(common.INTEGER, "4"),
				token(common.SEMICOLON, ";"),
			),
			statements(
				common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2)),
				groupingExpression(floatLiteralExpression("3.5", 3.5)),
				common.NewUnaryExpression(token(common.MINUS, "-"), integerLiteralExpression("4", 4)),
			),
		},
		{
			"tokens after EOF are ignored",
			tokensWithoutSemicolon(
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.EOF, ""),
				token(common.INTEGER, "2"),
				token(common.SEMICOLON, ";"),
			),
			statements(integerLiteralExpression("1", 1)),
		},
	})
}

func TestComplexExpression(t *testing.T) {
	assertParse(
		t,
		tokens(
			token(common.MINUS, "-"),
			token(common.OPEN_PAR, "("),
			token(common.INTEGER, "1"),
			token(common.PLUS, "+"),
			token(common.INTEGER, "2"),
			token(common.CLOSED_PAR, ")"),
			token(common.STAR, "*"),
			token(common.INTEGER, "3"),
			token(common.DOUBLE_STAR, "**"),
			token(common.INTEGER, "2"),
			token(common.DOUBLE_SLASH, "//"),
			token(common.INTEGER, "4"),
			token(common.PERCENTAGE, "%"),
			token(common.INTEGER, "5"),
			token(common.MINUS, "-"),
			token(common.INTEGER, "6"),
			token(common.SLASH, "/"),
			token(common.FLOAT, "7.5"),
		),
		statements(common.NewBinaryExpression(
			common.NewBinaryExpression(
				common.NewBinaryExpression(
					common.NewBinaryExpression(
						common.NewUnaryExpression(
							token(common.MINUS, "-"),
							groupingExpression(common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2))),
						),
						token(common.STAR, "*"),
						common.NewBinaryExpression(integerLiteralExpression("3", 3), token(common.DOUBLE_STAR, "**"), integerLiteralExpression("2", 2)),
					),
					token(common.DOUBLE_SLASH, "//"),
					integerLiteralExpression("4", 4),
				),
				token(common.PERCENTAGE, "%"),
				integerLiteralExpression("5", 5),
			),
			token(common.MINUS, "-"),
			common.NewBinaryExpression(integerLiteralExpression("6", 6), token(common.SLASH, "/"), floatLiteralExpression("7.5", 7.5)),
		)),
	)
}

func TestMissingSemicolon(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"no semicolon at all",
			tokensWithoutSemicolon(token(common.INTEGER, "1")),
			"[line 0, column 0] Expected ';' after expression",
		},
		{
			"no semicolon between statements",
			tokensWithoutSemicolon(
				token(common.INTEGER, "1"),
				token(common.INTEGER, "2"),
				token(common.SEMICOLON, ";"),
			),
			"[line 0, column 0] Expected ';' after expression",
		},
		{
			"unmatched closing parentheses",
			tokensWithoutSemicolon(
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.CLOSED_PAR, ")"),
				token(common.CLOSED_PAR, ")"),
				token(common.SEMICOLON, ";"),
			),
			"[line 0, column 0] Expected ';' after expression",
		},
	})
}

func TestUnclosedGroupingExpression(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"semicolon before closing parentheses",
			tokensWithoutSemicolon(
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
			),
			"[line 0, column 0] Grouping expression without close",
		},
		{
			"EOF before closing parentheses",
			tokensWithoutSemicolon(token(common.OPEN_PAR, "("), token(common.INTEGER, "1")),
			"[line 0, column 0] Grouping expression without close",
		},
		{
			"only inner grouping closed",
			tokens(
				token(common.OPEN_PAR, "("),
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.CLOSED_PAR, ")"),
			),
			"[line 0, column 0] Grouping expression without close",
		},
	})
}

func TestInvalidPrimaryExpression(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"only semicolon",
			tokens(),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"empty grouping",
			tokens(token(common.OPEN_PAR, "("), token(common.CLOSED_PAR, ")")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing right operand",
			tokens(token(common.INTEGER, "1"), token(common.PLUS, "+")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing right operand of an equality",
			tokens(token(common.INTEGER, "1"), token(common.DOUBLE_EQUAL, "==")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing right operand of a comparison",
			tokens(token(common.INTEGER, "1"), token(common.LESS_EQUAL, "<=")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing left operand of a comparison",
			tokens(token(common.GREATER, ">"), token(common.INTEGER, "1")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing left operand of an equality",
			tokens(token(common.DOUBLE_EQUAL, "=="), token(common.INTEGER, "1")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing left operand",
			tokens(token(common.STAR, "*"), token(common.INTEGER, "3")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing exponent",
			tokens(token(common.INTEGER, "2"), token(common.DOUBLE_STAR, "**")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"dangling unary minus",
			tokens(token(common.MINUS, "-")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"dangling not",
			tokens(token(common.NOT, "not")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing right operand of an and",
			tokens(token(common.TRUE, "True"), token(common.AND, "and")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing right operand of an or",
			tokens(token(common.TRUE, "True"), token(common.OR, "or")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing left operand of an or",
			tokens(token(common.OR, "or"), token(common.TRUE, "True")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"two logical operators in a row",
			tokens(token(common.TRUE, "True"), token(common.AND, "and"), token(common.OR, "or"), token(common.FALSE, "False")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"consecutive operators",
			tokens(token(common.INTEGER, "1"), token(common.PLUS, "+"), token(common.STAR, "*"), token(common.INTEGER, "2")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"unsupported dot",
			tokens(token(common.DOT, ".")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"error in a later statement discards the previous ones",
			tokensWithoutSemicolon(
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.PLUS, "+"),
				token(common.SEMICOLON, ";"),
			),
			"[line 0, column 0] Invalid primary expression",
		},
	})
}

func TestInvalidLiteralValue(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"integer out of range",
			tokens(token(common.INTEGER, "9223372036854775808")),
			"[line 0, column 0] Invalid integer: 9223372036854775808",
		},
		{
			"negated integer out of range",
			tokens(token(common.MINUS, "-"), token(common.INTEGER, "9223372036854775808")),
			"[line 0, column 0] Invalid integer: 9223372036854775808",
		},
		{
			"integer with non-numeric characters",
			tokens(token(common.INTEGER, "12a")),
			"[line 0, column 0] Invalid integer: 12a",
		},
		{
			"empty integer lexeme",
			tokens(token(common.INTEGER, "")),
			"[line 0, column 0] Invalid integer: ",
		},
		{
			"float with two dots",
			tokens(token(common.FLOAT, "1.2.3")),
			"[line 0, column 0] Invalid float: 1.2.3",
		},
		{
			"float out of range",
			tokens(token(common.FLOAT, "1e400")),
			"[line 0, column 0] Invalid float: 1e400",
		},
		{
			"string without double quotes",
			tokens(token(common.STRING, "hola")),
			"[line 0, column 0] Invalid string: hola",
		},
		{
			"string without closing double quote",
			tokens(token(common.STRING, `"hola`)),
			`[line 0, column 0] Invalid string: "hola`,
		},
		{
			"string with a single double quote",
			tokens(token(common.STRING, `"`)),
			`[line 0, column 0] Invalid string: "`,
		},
		{
			"empty string lexeme",
			tokens(token(common.STRING, "")),
			"[line 0, column 0] Invalid string: ",
		},
		{
			"string with an unescaped double quote inside",
			tokens(token(common.STRING, `"a"b"`)),
			`[line 0, column 0] Invalid string: "a"b"`,
		},
		{
			"string with an invalid escape sequence",
			tokens(token(common.STRING, `"a\qb"`)),
			`[line 0, column 0] Invalid string: "a\qb"`,
		},
		{
			"string with an escaped closing double quote",
			tokens(token(common.STRING, `"a\"`)),
			`[line 0, column 0] Invalid string: "a\"`,
		},
	})
}

func TestPrintStatement(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"integer literal",
			tokens(token(common.PRINT, "print"), token(common.INTEGER, "3")),
			[]common.Statement{common.NewPrintStatement(integerLiteralExpression("3", 3))},
		},
		{
			"string literal",
			tokens(token(common.PRINT, "print"), token(common.STRING, `"hola"`)),
			[]common.Statement{common.NewPrintStatement(stringLiteralExpression(`"hola"`, "hola"))},
		},
		{
			"binary expression",
			tokens(token(common.PRINT, "print"), token(common.INTEGER, "1"), token(common.PLUS, "+"), token(common.INTEGER, "2")),
			[]common.Statement{common.NewPrintStatement(
				common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2)),
			)},
		},
		{
			"grouping expression",
			tokens(token(common.PRINT, "print"), token(common.OPEN_PAR, "("), token(common.INTEGER, "1"), token(common.CLOSED_PAR, ")")),
			[]common.Statement{common.NewPrintStatement(groupingExpression(integerLiteralExpression("1", 1)))},
		},
		{
			"negation",
			tokens(token(common.PRINT, "print"), token(common.MINUS, "-"), token(common.FLOAT, "2.5")),
			[]common.Statement{common.NewPrintStatement(
				common.NewUnaryExpression(token(common.MINUS, "-"), floatLiteralExpression("2.5", 2.5)),
			)},
		},
	})
}

func TestPrintStatementWithOtherStatements(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"two print statements",
			tokensWithoutSemicolon(
				token(common.PRINT, "print"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.PRINT, "print"),
				token(common.INTEGER, "2"),
				token(common.SEMICOLON, ";"),
			),
			[]common.Statement{
				common.NewPrintStatement(integerLiteralExpression("1", 1)),
				common.NewPrintStatement(integerLiteralExpression("2", 2)),
			},
		},
		{
			"print statement between expression statements",
			tokensWithoutSemicolon(
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.PRINT, "print"),
				token(common.INTEGER, "2"),
				token(common.SEMICOLON, ";"),
				token(common.INTEGER, "3"),
				token(common.SEMICOLON, ";"),
			),
			[]common.Statement{
				common.NewExpressionStatement(integerLiteralExpression("1", 1)),
				common.NewPrintStatement(integerLiteralExpression("2", 2)),
				common.NewExpressionStatement(integerLiteralExpression("3", 3)),
			},
		},
	})
}

func TestPrintStatementErrors(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"only keyword",
			tokensWithoutSemicolon(token(common.PRINT, "print")),
			"[line 0, column 0] Expected expression after 'print'",
		},
		{
			"keyword followed by a semicolon",
			tokens(token(common.PRINT, "print")),
			"[line 0, column 0] Expected expression after 'print'",
		},
		{
			"keyword without EOF token",
			[]common.Token{token(common.PRINT, "print")},
			"[line 0, column 0] Expected expression after 'print'",
		},
		{
			"missing semicolon",
			tokensWithoutSemicolon(token(common.PRINT, "print"), token(common.INTEGER, "1")),
			"[line 0, column 0] Expected ';' after expression",
		},
		{
			"invalid expression",
			tokens(token(common.PRINT, "print"), token(common.CLOSED_PAR, ")")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"two consecutive keywords",
			tokens(token(common.PRINT, "print"), token(common.PRINT, "print"), token(common.INTEGER, "1")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"keyword inside an expression",
			tokens(token(common.INTEGER, "1"), token(common.PLUS, "+"), token(common.PRINT, "print"), token(common.INTEGER, "2")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"error in a print statement discards the previous ones",
			tokensWithoutSemicolon(
				token(common.PRINT, "print"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.PRINT, "print"),
				token(common.SEMICOLON, ";"),
			),
			"[line 0, column 0] Expected expression after 'print'",
		},
	})
}

func TestPrintStatementErrorPosition(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"points to the semicolon after the keyword",
			[]common.Token{tokenAt(common.PRINT, "print", 1, 1), tokenAt(common.SEMICOLON, ";", 1, 6), tokenAt(common.EOF, "", 1, 7)},
			"[line 1, column 6] Expected expression after 'print'",
		},
		{
			"points to the EOF after the keyword",
			[]common.Token{tokenAt(common.PRINT, "print", 1, 1), tokenAt(common.EOF, "", 1, 6)},
			"[line 1, column 6] Expected expression after 'print'",
		},
	})
}

func TestBlockStatement(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"empty block",
			tokensWithoutSemicolon(token(common.OPEN_BRACE, "{"), token(common.CLOSED_BRACE, "}")),
			[]common.Statement{blockStatement()},
		},
		{
			"block with statements",
			tokensWithoutSemicolon(
				token(common.OPEN_BRACE, "{"),
				token(common.PRINT, "print"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.INTEGER, "2"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{blockStatement(
				common.NewPrintStatement(integerLiteralExpression("1", 1)),
				common.NewExpressionStatement(integerLiteralExpression("2", 2)),
			)},
		},
		{
			"nested blocks",
			tokensWithoutSemicolon(
				token(common.OPEN_BRACE, "{"),
				token(common.OPEN_BRACE, "{"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{blockStatement(
				blockStatement(common.NewExpressionStatement(integerLiteralExpression("1", 1))),
				blockStatement(),
			)},
		},
		{
			"block between other statements",
			tokensWithoutSemicolon(
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.OPEN_BRACE, "{"),
				token(common.INTEGER, "2"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
				token(common.INTEGER, "3"),
				token(common.SEMICOLON, ";"),
			),
			[]common.Statement{
				common.NewExpressionStatement(integerLiteralExpression("1", 1)),
				blockStatement(common.NewExpressionStatement(integerLiteralExpression("2", 2))),
				common.NewExpressionStatement(integerLiteralExpression("3", 3)),
			},
		},
	})
}

func TestBlockStatementErrors(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"only open brace",
			tokensWithoutSemicolon(token(common.OPEN_BRACE, "{")),
			"[line 0, column 0] Expected '}' after block",
		},
		{
			"only open brace without EOF token",
			[]common.Token{token(common.OPEN_BRACE, "{")},
			"[line 0, column 0] Expected '}' after block",
		},
		{
			"unclosed block with statements",
			tokensWithoutSemicolon(
				token(common.OPEN_BRACE, "{"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
			),
			"[line 0, column 0] Expected '}' after block",
		},
		{
			"unclosed nested block",
			tokensWithoutSemicolon(
				token(common.OPEN_BRACE, "{"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Expected '}' after block",
		},
		{
			"only closed brace",
			tokensWithoutSemicolon(token(common.CLOSED_BRACE, "}")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing semicolon inside the block",
			tokensWithoutSemicolon(
				token(common.OPEN_BRACE, "{"),
				token(common.INTEGER, "1"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Expected ';' after expression",
		},
	})
}

func TestBlockStatementErrorPosition(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"points to the EOF of an unclosed block",
			[]common.Token{
				tokenAt(common.OPEN_BRACE, "{", 1, 1),
				tokenAt(common.INTEGER, "1", 1, 3),
				tokenAt(common.SEMICOLON, ";", 1, 4),
				tokenAt(common.EOF, "", 2, 1),
			},
			"[line 2, column 1] Expected '}' after block",
		},
	})
}

func TestIfStatement(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"if without else",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.PRINT, "print"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{ifStatement(
				booleanLiteralExpression(common.TRUE, "True", true),
				blockStatement(common.NewPrintStatement(integerLiteralExpression("1", 1))),
				nil,
			)},
		},
		{
			"if with else",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.FALSE, "False"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.ELSE, "else"),
				token(common.OPEN_BRACE, "{"),
				token(common.INTEGER, "2"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{ifStatement(
				booleanLiteralExpression(common.FALSE, "False", false),
				blockStatement(),
				blockStatement(common.NewExpressionStatement(integerLiteralExpression("2", 2))),
			)},
		},
		{
			"else if chain",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.FALSE, "False"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.ELSE, "else"),
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.ELSE, "else"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{ifStatement(
				booleanLiteralExpression(common.FALSE, "False", false),
				blockStatement(),
				ifStatement(
					booleanLiteralExpression(common.TRUE, "True", true),
					blockStatement(),
					blockStatement(),
				),
			)},
		},
		{
			"else if without final else",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.FALSE, "False"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.ELSE, "else"),
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{ifStatement(
				booleanLiteralExpression(common.FALSE, "False", false),
				blockStatement(),
				ifStatement(booleanLiteralExpression(common.TRUE, "True", true), blockStatement(), nil),
			)},
		},
		{
			"compound condition",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.LESS, "<"),
				token(common.INTEGER, "2"),
				token(common.AND, "and"),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{ifStatement(
				common.NewBinaryExpression(
					common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.LESS, "<"), integerLiteralExpression("2", 2)),
					token(common.AND, "and"),
					booleanLiteralExpression(common.TRUE, "True", true),
				),
				blockStatement(),
				nil,
			)},
		},
		{
			"grouped condition",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{ifStatement(
				groupingExpression(booleanLiteralExpression(common.TRUE, "True", true)),
				blockStatement(),
				nil,
			)},
		},
		{
			"nested if",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.FALSE, "False"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.ELSE, "else"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{ifStatement(
				booleanLiteralExpression(common.TRUE, "True", true),
				blockStatement(ifStatement(
					booleanLiteralExpression(common.FALSE, "False", false),
					blockStatement(),
					blockStatement(),
				)),
				nil,
			)},
		},
		{
			"if followed by another statement",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.PRINT, "print"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
			),
			[]common.Statement{
				ifStatement(booleanLiteralExpression(common.TRUE, "True", true), blockStatement(), nil),
				common.NewPrintStatement(integerLiteralExpression("1", 1)),
			},
		},
		{
			"keeps the position of the if token",
			[]common.Token{
				tokenAt(common.IF, "if", 2, 3),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.EOF, ""),
			},
			[]common.Statement{common.NewIfStatement(
				tokenAt(common.IF, "if", 2, 3),
				booleanLiteralExpression(common.TRUE, "True", true),
				blockStatement(),
				nil,
			)},
		},
	})
}

func TestIfStatementErrors(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"only keyword",
			tokensWithoutSemicolon(token(common.IF, "if")),
			"[line 0, column 0] Expected '(' after 'if'",
		},
		{
			"only keyword without EOF token",
			[]common.Token{token(common.IF, "if")},
			"[line 0, column 0] Expected '(' after 'if'",
		},
		{
			"missing open parentheses",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.TRUE, "True"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Expected '(' after 'if'",
		},
		{
			"open parentheses followed by EOF",
			tokensWithoutSemicolon(token(common.IF, "if"), token(common.OPEN_PAR, "(")),
			"[line 0, column 0] Expected expression after 'if ('",
		},
		{
			"open parentheses followed by a semicolon",
			tokens(token(common.IF, "if"), token(common.OPEN_PAR, "(")),
			"[line 0, column 0] Expected expression after 'if ('",
		},
		{
			"empty condition",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing closed parentheses",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Expected ')' after expression",
		},
		{
			"missing block",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
			),
			"[line 0, column 0] Expected block statement",
		},
		{
			"statement instead of block",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.PRINT, "print"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
			),
			"[line 0, column 0] Expected block statement",
		},
		{
			"unclosed if block",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
			),
			"[line 0, column 0] Expected '}' after block",
		},
		{
			"else followed by EOF",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.ELSE, "else"),
			),
			"[line 0, column 0] Expected '{' or 'if' after 'else'",
		},
		{
			"else followed by a statement",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.ELSE, "else"),
				token(common.PRINT, "print"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
			),
			"[line 0, column 0] Expected '{' or 'if' after 'else'",
		},
		{
			"else followed by a semicolon",
			tokens(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.ELSE, "else"),
			),
			"[line 0, column 0] Expected '{' or 'if' after 'else'",
		},
		{
			"unclosed else block",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.ELSE, "else"),
				token(common.OPEN_BRACE, "{"),
			),
			"[line 0, column 0] Expected '}' after block",
		},
		{
			"error in the else if",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.ELSE, "else"),
				token(common.IF, "if"),
				token(common.TRUE, "True"),
			),
			"[line 0, column 0] Expected '(' after 'if'",
		},
		{
			"else without if",
			tokensWithoutSemicolon(
				token(common.ELSE, "else"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"error in an if discards the previous statements",
			tokensWithoutSemicolon(
				token(common.PRINT, "print"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.IF, "if"),
				token(common.TRUE, "True"),
			),
			"[line 0, column 0] Expected '(' after 'if'",
		},
	})
}

func TestIfStatementErrorPosition(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"points to the token after the keyword",
			[]common.Token{
				tokenAt(common.IF, "if", 1, 1),
				tokenAt(common.TRUE, "True", 1, 4),
				tokenAt(common.EOF, "", 1, 8),
			},
			"[line 1, column 4] Expected '(' after 'if'",
		},
		{
			"points to the token after the closed parentheses",
			[]common.Token{
				tokenAt(common.IF, "if", 1, 1),
				tokenAt(common.OPEN_PAR, "(", 1, 4),
				tokenAt(common.TRUE, "True", 1, 5),
				tokenAt(common.CLOSED_PAR, ")", 1, 9),
				tokenAt(common.PRINT, "print", 1, 11),
				tokenAt(common.INTEGER, "1", 1, 17),
				tokenAt(common.SEMICOLON, ";", 1, 18),
				tokenAt(common.EOF, "", 1, 19),
			},
			"[line 1, column 11] Expected block statement",
		},
		{
			"points to the token after the else",
			[]common.Token{
				tokenAt(common.IF, "if", 1, 1),
				tokenAt(common.OPEN_PAR, "(", 1, 4),
				tokenAt(common.TRUE, "True", 1, 5),
				tokenAt(common.CLOSED_PAR, ")", 1, 9),
				tokenAt(common.OPEN_BRACE, "{", 1, 11),
				tokenAt(common.CLOSED_BRACE, "}", 1, 12),
				tokenAt(common.ELSE, "else", 1, 14),
				tokenAt(common.PRINT, "print", 1, 19),
				tokenAt(common.INTEGER, "1", 1, 25),
				tokenAt(common.SEMICOLON, ";", 1, 26),
				tokenAt(common.EOF, "", 1, 27),
			},
			"[line 1, column 19] Expected '{' or 'if' after 'else'",
		},
	})
}

func TestWhileStatement(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"while with a statement",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.PRINT, "print"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{whileStatement(
				booleanLiteralExpression(common.TRUE, "True", true),
				blockStatement(common.NewPrintStatement(integerLiteralExpression("1", 1))),
			)},
		},
		{
			"while with an empty body",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.FALSE, "False"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{whileStatement(booleanLiteralExpression(common.FALSE, "False", false), blockStatement())},
		},
		{
			"compound condition",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.LESS, "<"),
				token(common.INTEGER, "2"),
				token(common.AND, "and"),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{whileStatement(
				common.NewBinaryExpression(
					common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.LESS, "<"), integerLiteralExpression("2", 2)),
					token(common.AND, "and"),
					booleanLiteralExpression(common.TRUE, "True", true),
				),
				blockStatement(),
			)},
		},
		{
			"grouped condition",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{whileStatement(
				groupingExpression(booleanLiteralExpression(common.TRUE, "True", true)),
				blockStatement(),
			)},
		},
		{
			"break and continue in the body",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CONTINUE, "continue"),
				token(common.SEMICOLON, ";"),
				token(common.BREAK, "break"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{whileStatement(
				booleanLiteralExpression(common.TRUE, "True", true),
				blockStatement(continueStatement(), breakStatement()),
			)},
		},
		{
			"break inside an if in the body",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.FALSE, "False"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.BREAK, "break"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{whileStatement(
				booleanLiteralExpression(common.TRUE, "True", true),
				blockStatement(ifStatement(
					booleanLiteralExpression(common.FALSE, "False", false),
					blockStatement(breakStatement()),
					nil,
				)),
			)},
		},
		{
			"nested while",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.FALSE, "False"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.BREAK, "break"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{whileStatement(
				booleanLiteralExpression(common.TRUE, "True", true),
				blockStatement(
					whileStatement(booleanLiteralExpression(common.FALSE, "False", false), blockStatement()),
					breakStatement(),
				),
			)},
		},
		{
			"while followed by another statement",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.FALSE, "False"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.PRINT, "print"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
			),
			[]common.Statement{
				whileStatement(booleanLiteralExpression(common.FALSE, "False", false), blockStatement()),
				common.NewPrintStatement(integerLiteralExpression("1", 1)),
			},
		},
		{
			"break and continue outside a loop are parsed",
			tokensWithoutSemicolon(
				token(common.BREAK, "break"),
				token(common.SEMICOLON, ";"),
				token(common.CONTINUE, "continue"),
				token(common.SEMICOLON, ";"),
			),
			[]common.Statement{breakStatement(), continueStatement()},
		},
		{
			"keeps the position of the while, break and continue tokens",
			[]common.Token{
				tokenAt(common.WHILE, "while", 2, 3),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				tokenAt(common.BREAK, "break", 3, 5),
				token(common.SEMICOLON, ";"),
				tokenAt(common.CONTINUE, "continue", 4, 5),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
				token(common.EOF, ""),
			},
			[]common.Statement{common.NewWhileStatement(
				tokenAt(common.WHILE, "while", 2, 3),
				booleanLiteralExpression(common.TRUE, "True", true),
				blockStatement(
					common.NewBreakStatement(tokenAt(common.BREAK, "break", 3, 5)),
					common.NewContinueStatement(tokenAt(common.CONTINUE, "continue", 4, 5)),
				),
			)},
		},
	})
}

func TestWhileStatementErrors(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"only keyword",
			tokensWithoutSemicolon(token(common.WHILE, "while")),
			"[line 0, column 0] Expected '(' after 'while'",
		},
		{
			"only keyword without EOF token",
			[]common.Token{token(common.WHILE, "while")},
			"[line 0, column 0] Expected '(' after 'while'",
		},
		{
			"missing open parentheses",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.TRUE, "True"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Expected '(' after 'while'",
		},
		{
			"open parentheses followed by EOF",
			tokensWithoutSemicolon(token(common.WHILE, "while"), token(common.OPEN_PAR, "(")),
			"[line 0, column 0] Expected expression after 'while ('",
		},
		{
			"open parentheses followed by a semicolon",
			tokens(token(common.WHILE, "while"), token(common.OPEN_PAR, "(")),
			"[line 0, column 0] Expected expression after 'while ('",
		},
		{
			"empty condition",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing closed parentheses",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Expected ')' after expression",
		},
		{
			"missing block",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
			),
			"[line 0, column 0] Expected block statement",
		},
		{
			"statement instead of block",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.BREAK, "break"),
				token(common.SEMICOLON, ";"),
			),
			"[line 0, column 0] Expected block statement",
		},
		{
			"unclosed while block",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.BREAK, "break"),
				token(common.SEMICOLON, ";"),
			),
			"[line 0, column 0] Expected '}' after block",
		},
		{
			"else after while",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.ELSE, "else"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"error in the body",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.INTEGER, "1"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Expected ';' after expression",
		},
		{
			"error in a while discards the previous statements",
			tokensWithoutSemicolon(
				token(common.PRINT, "print"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.WHILE, "while"),
				token(common.TRUE, "True"),
			),
			"[line 0, column 0] Expected '(' after 'while'",
		},
	})
}

func TestBreakAndContinueStatementErrors(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{"break without semicolon", tokensWithoutSemicolon(token(common.BREAK, "break")), "[line 0, column 0] Expected ';' after 'break'"},
		{"break without EOF token", []common.Token{token(common.BREAK, "break")}, "[line 0, column 0] Expected ';' after 'break'"},
		{
			"break followed by an expression",
			tokens(token(common.BREAK, "break"), token(common.INTEGER, "1")),
			"[line 0, column 0] Expected ';' after 'break'",
		},
		{
			"continue without semicolon",
			tokensWithoutSemicolon(token(common.CONTINUE, "continue")),
			"[line 0, column 0] Expected ';' after 'continue'",
		},
		{
			"continue without EOF token",
			[]common.Token{token(common.CONTINUE, "continue")},
			"[line 0, column 0] Expected ';' after 'continue'",
		},
		{
			"continue followed by an expression",
			tokens(token(common.CONTINUE, "continue"), token(common.INTEGER, "1")),
			"[line 0, column 0] Expected ';' after 'continue'",
		},
		{
			"break without semicolon inside a while",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.BREAK, "break"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Expected ';' after 'break'",
		},
	})
}

func TestWhileStatementErrorPosition(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"points to the token after the keyword",
			[]common.Token{
				tokenAt(common.WHILE, "while", 1, 1),
				tokenAt(common.TRUE, "True", 1, 7),
				tokenAt(common.EOF, "", 1, 11),
			},
			"[line 1, column 7] Expected '(' after 'while'",
		},
		{
			"points to the token after the closed parentheses",
			[]common.Token{
				tokenAt(common.WHILE, "while", 1, 1),
				tokenAt(common.OPEN_PAR, "(", 1, 7),
				tokenAt(common.TRUE, "True", 1, 8),
				tokenAt(common.CLOSED_PAR, ")", 1, 12),
				tokenAt(common.BREAK, "break", 1, 14),
				tokenAt(common.SEMICOLON, ";", 1, 19),
				tokenAt(common.EOF, "", 1, 20),
			},
			"[line 1, column 14] Expected block statement",
		},
		{
			"points to the token after the break",
			[]common.Token{
				tokenAt(common.WHILE, "while", 1, 1),
				tokenAt(common.OPEN_PAR, "(", 1, 7),
				tokenAt(common.TRUE, "True", 1, 8),
				tokenAt(common.CLOSED_PAR, ")", 1, 12),
				tokenAt(common.OPEN_BRACE, "{", 1, 14),
				tokenAt(common.BREAK, "break", 2, 3),
				tokenAt(common.CLOSED_BRACE, "}", 3, 1),
				tokenAt(common.EOF, "", 3, 2),
			},
			"[line 3, column 1] Expected ';' after 'break'",
		},
		{
			"points to the token after the continue",
			[]common.Token{
				tokenAt(common.CONTINUE, "continue", 1, 1),
				tokenAt(common.EOF, "", 2, 1),
			},
			"[line 2, column 1] Expected ';' after 'continue'",
		},
	})
}

func TestVarDeclarationStatement(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"int variable",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
			),
			[]common.Statement{varDeclarationStatement("x", types.Int, integerLiteralExpression("1", 1))},
		},
		{
			"float variable",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.FLOAT_TYPE, "Float"),
				token(common.EQUAL, "="),
				token(common.FLOAT, "2.5"),
			),
			[]common.Statement{varDeclarationStatement("x", types.Float, floatLiteralExpression("2.5", 2.5))},
		},
		{
			"string variable",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.STRING_TYPE, "String"),
				token(common.EQUAL, "="),
				token(common.STRING, `"a"`),
			),
			[]common.Statement{varDeclarationStatement("x", types.Str, stringLiteralExpression(`"a"`, "a"))},
		},
		{
			"bool variable",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.BOOL_TYPE, "Bool"),
				token(common.EQUAL, "="),
				token(common.TRUE, "True"),
			),
			[]common.Statement{varDeclarationStatement("x", types.Bool, booleanLiteralExpression(common.TRUE, "True", true))},
		},
		{
			"the type is not checked against the value",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.BOOL_TYPE, "Bool"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
			),
			[]common.Statement{varDeclarationStatement("x", types.Bool, integerLiteralExpression("1", 1))},
		},
		{
			"compound value",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
				token(common.PLUS, "+"),
				token(common.INTEGER, "2"),
				token(common.STAR, "*"),
				token(common.INTEGER, "3"),
			),
			[]common.Statement{varDeclarationStatement("x", types.Int, common.NewBinaryExpression(
				integerLiteralExpression("1", 1),
				token(common.PLUS, "+"),
				common.NewBinaryExpression(integerLiteralExpression("2", 2), token(common.STAR, "*"), integerLiteralExpression("3", 3)),
			))},
		},
		{
			"value with another variable",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "y"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
				token(common.IDENTIFIER, "x"),
				token(common.MINUS, "-"),
				token(common.INTEGER, "1"),
			),
			[]common.Statement{varDeclarationStatement("y", types.Int, common.NewBinaryExpression(
				variableExpression("x"),
				token(common.MINUS, "-"),
				integerLiteralExpression("1", 1),
			))},
		},
		{
			"value with an assignment",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "y"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
				token(common.IDENTIFIER, "x"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
			),
			[]common.Statement{varDeclarationStatement("y", types.Int, varAssignmentExpression("x", integerLiteralExpression("1", 1)))},
		},
		{
			"declaration inside a block",
			tokensWithoutSemicolon(
				token(common.OPEN_BRACE, "{"),
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{blockStatement(varDeclarationStatement("x", types.Int, integerLiteralExpression("1", 1)))},
		},
		{
			"declaration followed by another statement",
			tokensWithoutSemicolon(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.PRINT, "print"),
				token(common.IDENTIFIER, "x"),
				token(common.SEMICOLON, ";"),
			),
			[]common.Statement{
				varDeclarationStatement("x", types.Int, integerLiteralExpression("1", 1)),
				common.NewPrintStatement(variableExpression("x")),
			},
		},
		{
			"redeclaration is parsed",
			tokensWithoutSemicolon(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.STRING_TYPE, "String"),
				token(common.EQUAL, "="),
				token(common.STRING, `"a"`),
				token(common.SEMICOLON, ";"),
			),
			[]common.Statement{
				varDeclarationStatement("x", types.Int, integerLiteralExpression("1", 1)),
				varDeclarationStatement("x", types.Str, stringLiteralExpression(`"a"`, "a")),
			},
		},
		{
			"keeps the position of the let and identifier tokens",
			[]common.Token{
				tokenAt(common.LET, "let", 2, 3),
				tokenAt(common.IDENTIFIER, "x", 2, 7),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.EOF, ""),
			},
			[]common.Statement{common.NewVarDeclarationStatement(
				tokenAt(common.LET, "let", 2, 3),
				tokenAt(common.IDENTIFIER, "x", 2, 7),
				types.Int,
				integerLiteralExpression("1", 1),
			)},
		},
	})
}

func TestVarDeclarationStatementErrors(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{"only keyword", tokensWithoutSemicolon(token(common.LET, "let")), "[line 0, column 0] Expected identifier after let"},
		{"only keyword without EOF token", []common.Token{token(common.LET, "let")}, "[line 0, column 0] Expected identifier after let"},
		{
			"keyword instead of identifier",
			tokens(
				token(common.LET, "let"),
				token(common.WHILE, "while"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
			),
			"[line 0, column 0] Expected identifier after let",
		},
		{
			"type instead of identifier",
			tokens(token(common.LET, "let"), token(common.INT_TYPE, "Int"), token(common.EQUAL, "="), token(common.INTEGER, "1")),
			"[line 0, column 0] Expected identifier after let",
		},
		{
			"missing colon",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
			),
			"[line 0, column 0] Expected ':' after let x",
		},
		{
			"missing type",
			tokens(token(common.LET, "let"), token(common.IDENTIFIER, "x"), token(common.EQUAL, "="), token(common.INTEGER, "1")),
			"[line 0, column 0] Expected ':' after let x",
		},
		{
			"missing type after colon",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
			),
			"[line 0, column 0] Expected type after let x :",
		},
		{
			"colon followed by EOF",
			tokensWithoutSemicolon(token(common.LET, "let"), token(common.IDENTIFIER, "x"), token(common.COLON, ":")),
			"[line 0, column 0] Expected type after let x :",
		},
		{
			"identifier instead of type",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.IDENTIFIER, "int"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
			),
			"[line 0, column 0] Expected type after let x :",
		},
		{
			"missing equal",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.INTEGER, "1"),
			),
			"[line 0, column 0] Expected '=' after let x : Int",
		},
		{
			"double equal instead of equal",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.DOUBLE_EQUAL, "=="),
				token(common.INTEGER, "1"),
			),
			"[line 0, column 0] Expected '=' after let x : Int",
		},
		{
			"declaration without value",
			tokens(token(common.LET, "let"), token(common.IDENTIFIER, "x"), token(common.COLON, ":"), token(common.INT_TYPE, "Int")),
			"[line 0, column 0] Expected '=' after let x : Int",
		},
		{
			"missing value",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
			),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing value without EOF token",
			[]common.Token{
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
			},
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing semicolon",
			tokensWithoutSemicolon(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
			),
			"[line 0, column 0] Expected ';' after variable declaration",
		},
		{
			"declaration as the value",
			tokens(
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
				token(common.LET, "let"),
			),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"declaration inside an expression",
			tokens(token(common.PRINT, "print"), token(common.LET, "let"), token(common.IDENTIFIER, "x")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"error in a declaration discards the previous statements",
			tokensWithoutSemicolon(
				token(common.PRINT, "print"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
			),
			"[line 0, column 0] Expected ':' after let x",
		},
	})
}

func TestVarDeclarationStatementErrorPosition(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"points to the token after the keyword",
			[]common.Token{
				tokenAt(common.LET, "let", 1, 1),
				tokenAt(common.INTEGER, "1", 1, 5),
				tokenAt(common.EOF, "", 1, 6),
			},
			"[line 1, column 5] Expected identifier after let",
		},
		{
			"points to the token after the identifier",
			[]common.Token{
				tokenAt(common.LET, "let", 1, 1),
				tokenAt(common.IDENTIFIER, "x", 1, 5),
				tokenAt(common.INT_TYPE, "Int", 1, 7),
				tokenAt(common.EOF, "", 1, 10),
			},
			"[line 1, column 7] Expected ':' after let x",
		},
		{
			"points to the token after the colon",
			[]common.Token{
				tokenAt(common.LET, "let", 1, 1),
				tokenAt(common.IDENTIFIER, "x", 1, 5),
				tokenAt(common.COLON, ":", 1, 6),
				tokenAt(common.EQUAL, "=", 1, 8),
				tokenAt(common.EOF, "", 1, 9),
			},
			"[line 1, column 8] Expected type after let x :",
		},
		{
			"points to the token after the type",
			[]common.Token{
				tokenAt(common.LET, "let", 1, 1),
				tokenAt(common.IDENTIFIER, "x", 1, 5),
				tokenAt(common.COLON, ":", 1, 6),
				tokenAt(common.INT_TYPE, "Int", 1, 8),
				tokenAt(common.INTEGER, "1", 1, 12),
				tokenAt(common.EOF, "", 1, 13),
			},
			"[line 1, column 12] Expected '=' after let x : Int",
		},
		{
			"points to the token after the value",
			[]common.Token{
				tokenAt(common.LET, "let", 1, 1),
				tokenAt(common.IDENTIFIER, "x", 1, 5),
				tokenAt(common.COLON, ":", 1, 6),
				tokenAt(common.INT_TYPE, "Int", 1, 8),
				tokenAt(common.EQUAL, "=", 1, 12),
				tokenAt(common.INTEGER, "1", 1, 14),
				tokenAt(common.EOF, "", 2, 1),
			},
			"[line 2, column 1] Expected ';' after variable declaration",
		},
	})
}

func TestVariableExpression(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{"variable", tokens(token(common.IDENTIFIER, "x")), statements(variableExpression("x"))},
		{
			"variable in a binary expression",
			tokens(token(common.IDENTIFIER, "x"), token(common.PLUS, "+"), token(common.IDENTIFIER, "y")),
			statements(common.NewBinaryExpression(variableExpression("x"), token(common.PLUS, "+"), variableExpression("y"))),
		},
		{
			"variable in a unary expression",
			tokens(token(common.NOT, "not"), token(common.IDENTIFIER, "x")),
			statements(common.NewUnaryExpression(token(common.NOT, "not"), variableExpression("x"))),
		},
		{
			"variable in a power",
			tokens(token(common.IDENTIFIER, "x"), token(common.DOUBLE_STAR, "**"), token(common.INTEGER, "2")),
			statements(common.NewBinaryExpression(variableExpression("x"), token(common.DOUBLE_STAR, "**"), integerLiteralExpression("2", 2))),
		},
		{
			"grouped variable",
			tokens(token(common.OPEN_PAR, "("), token(common.IDENTIFIER, "x"), token(common.CLOSED_PAR, ")")),
			statements(groupingExpression(variableExpression("x"))),
		},
		{
			"variable in a print statement",
			tokensWithoutSemicolon(token(common.PRINT, "print"), token(common.IDENTIFIER, "x"), token(common.SEMICOLON, ";")),
			[]common.Statement{common.NewPrintStatement(variableExpression("x"))},
		},
		{
			"variable as a condition",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{whileStatement(variableExpression("x"), blockStatement())},
		},
		{
			"keeps the position of the identifier token",
			[]common.Token{tokenAt(common.IDENTIFIER, "x", 3, 5), token(common.SEMICOLON, ";"), token(common.EOF, "")},
			statements(common.NewVariableExpression(tokenAt(common.IDENTIFIER, "x", 3, 5))),
		},
	})
}

func TestVariableExpressionErrors(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{"missing semicolon", tokensWithoutSemicolon(token(common.IDENTIFIER, "x")), "[line 0, column 0] Expected ';' after expression"},
		{
			"two variables without operator",
			tokens(token(common.IDENTIFIER, "x"), token(common.IDENTIFIER, "y")),
			"[line 0, column 0] Expected ';' after expression",
		},
		{
			"variable followed by a colon",
			tokens(token(common.IDENTIFIER, "x"), token(common.COLON, ":"), token(common.INT_TYPE, "Int")),
			"[line 0, column 0] Expected ';' after expression",
		},
		{"type as a value", tokens(token(common.INT_TYPE, "Int")), "[line 0, column 0] Invalid primary expression"},
	})
}

func TestVarAssignmentExpression(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"assignment",
			tokens(token(common.IDENTIFIER, "x"), token(common.EQUAL, "="), token(common.INTEGER, "1")),
			statements(varAssignmentExpression("x", integerLiteralExpression("1", 1))),
		},
		{
			"assignment of a variable",
			tokens(token(common.IDENTIFIER, "x"), token(common.EQUAL, "="), token(common.IDENTIFIER, "y")),
			statements(varAssignmentExpression("x", variableExpression("y"))),
		},
		{
			"assignment of the variable itself",
			tokens(
				token(common.IDENTIFIER, "x"),
				token(common.EQUAL, "="),
				token(common.IDENTIFIER, "x"),
				token(common.PLUS, "+"),
				token(common.INTEGER, "1"),
			),
			statements(varAssignmentExpression("x", common.NewBinaryExpression(
				variableExpression("x"),
				token(common.PLUS, "+"),
				integerLiteralExpression("1", 1),
			))),
		},
		{
			"assignment has lower precedence than logical operators",
			tokens(
				token(common.IDENTIFIER, "x"),
				token(common.EQUAL, "="),
				token(common.TRUE, "True"),
				token(common.OR, "or"),
				token(common.FALSE, "False"),
			),
			statements(varAssignmentExpression("x", common.NewBinaryExpression(
				booleanLiteralExpression(common.TRUE, "True", true),
				token(common.OR, "or"),
				booleanLiteralExpression(common.FALSE, "False", false),
			))),
		},
		{
			"assignment of an equality",
			tokens(
				token(common.IDENTIFIER, "x"),
				token(common.EQUAL, "="),
				token(common.IDENTIFIER, "y"),
				token(common.DOUBLE_EQUAL, "=="),
				token(common.INTEGER, "1"),
			),
			statements(varAssignmentExpression("x", common.NewBinaryExpression(
				variableExpression("y"),
				token(common.DOUBLE_EQUAL, "=="),
				integerLiteralExpression("1", 1),
			))),
		},
		{
			"chained assignment is right associative",
			tokens(
				token(common.IDENTIFIER, "a"),
				token(common.EQUAL, "="),
				token(common.IDENTIFIER, "b"),
				token(common.EQUAL, "="),
				token(common.IDENTIFIER, "c"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
			),
			statements(varAssignmentExpression("a", varAssignmentExpression("b", varAssignmentExpression("c", integerLiteralExpression("1", 1))))),
		},
		{
			"grouped assignment",
			tokens(
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
				token(common.CLOSED_PAR, ")"),
			),
			statements(groupingExpression(varAssignmentExpression("x", integerLiteralExpression("1", 1)))),
		},
		{
			"grouped assignment as an operand",
			tokens(
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
				token(common.CLOSED_PAR, ")"),
				token(common.PLUS, "+"),
				token(common.INTEGER, "2"),
			),
			statements(common.NewBinaryExpression(
				groupingExpression(varAssignmentExpression("x", integerLiteralExpression("1", 1))),
				token(common.PLUS, "+"),
				integerLiteralExpression("2", 2),
			)),
		},
		{
			"assignment in a print statement",
			tokensWithoutSemicolon(
				token(common.PRINT, "print"),
				token(common.IDENTIFIER, "x"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
			),
			[]common.Statement{common.NewPrintStatement(varAssignmentExpression("x", integerLiteralExpression("1", 1)))},
		},
		{
			"assignment as a condition",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.EQUAL, "="),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{ifStatement(varAssignmentExpression("x", booleanLiteralExpression(common.TRUE, "True", true)), blockStatement(), nil)},
		},
		{
			"assignment inside a while body",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.LESS, "<"),
				token(common.INTEGER, "3"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.IDENTIFIER, "x"),
				token(common.EQUAL, "="),
				token(common.IDENTIFIER, "x"),
				token(common.PLUS, "+"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{whileStatement(
				common.NewBinaryExpression(variableExpression("x"), token(common.LESS, "<"), integerLiteralExpression("3", 3)),
				blockStatement(common.NewExpressionStatement(varAssignmentExpression("x", common.NewBinaryExpression(
					variableExpression("x"),
					token(common.PLUS, "+"),
					integerLiteralExpression("1", 1),
				)))),
			)},
		},
		{
			"keeps the position of the identifier token",
			[]common.Token{
				tokenAt(common.IDENTIFIER, "x", 2, 3),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.EOF, ""),
			},
			statements(common.NewVarAssignmentExpression(tokenAt(common.IDENTIFIER, "x", 2, 3), integerLiteralExpression("1", 1))),
		},
	})
}

func TestVarAssignmentExpressionErrors(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{"missing value", tokens(token(common.IDENTIFIER, "x"), token(common.EQUAL, "=")), "[line 0, column 0] Invalid primary expression"},
		{
			"missing value without EOF token",
			[]common.Token{token(common.IDENTIFIER, "x"), token(common.EQUAL, "=")},
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing semicolon",
			tokensWithoutSemicolon(token(common.IDENTIFIER, "x"), token(common.EQUAL, "="), token(common.INTEGER, "1")),
			"[line 0, column 0] Expected ';' after expression",
		},
		{"missing variable", tokens(token(common.EQUAL, "="), token(common.INTEGER, "1")), "[line 0, column 0] Invalid primary expression"},
		{
			"assignment to a literal",
			tokens(token(common.INTEGER, "1"), token(common.EQUAL, "="), token(common.INTEGER, "2")),
			"[line 0, column 0] Expected ';' after expression",
		},
		{
			"assignment to a grouped variable",
			tokens(
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.CLOSED_PAR, ")"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
			),
			"[line 0, column 0] Expected ';' after expression",
		},
		{
			"assignment to a binary expression",
			tokens(
				token(common.IDENTIFIER, "x"),
				token(common.PLUS, "+"),
				token(common.IDENTIFIER, "y"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
			),
			"[line 0, column 0] Expected ';' after expression",
		},
		{
			"assignment as an operand without grouping",
			tokens(
				token(common.INTEGER, "1"),
				token(common.PLUS, "+"),
				token(common.IDENTIFIER, "x"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "2"),
			),
			"[line 0, column 0] Expected ';' after expression",
		},
		{
			"double equal followed by equal",
			tokens(token(common.IDENTIFIER, "x"), token(common.DOUBLE_EQUAL, "=="), token(common.EQUAL, "="), token(common.INTEGER, "1")),
			"[line 0, column 0] Invalid primary expression",
		},
	})
}

func TestVarAssignmentExpressionErrorPosition(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"points to the token after the equal",
			[]common.Token{
				tokenAt(common.IDENTIFIER, "x", 1, 1),
				tokenAt(common.EQUAL, "=", 1, 3),
				tokenAt(common.SEMICOLON, ";", 1, 4),
				tokenAt(common.EOF, "", 1, 5),
			},
			"[line 1, column 4] Invalid primary expression",
		},
		{
			"points to the equal after a non assignable expression",
			[]common.Token{
				tokenAt(common.INTEGER, "1", 1, 1),
				tokenAt(common.EQUAL, "=", 1, 3),
				tokenAt(common.INTEGER, "2", 1, 5),
				tokenAt(common.SEMICOLON, ";", 1, 6),
				tokenAt(common.EOF, "", 1, 7),
			},
			"[line 1, column 3] Expected ';' after expression",
		},
	})
}

func parameter(name string, paramType types.Type) common.Parameter {
	return common.NewParameter(token(common.IDENTIFIER, name), paramType)
}

func funcDeclarationStatement(
	name string,
	parameters []common.Parameter,
	returnType types.Type,
	body ...common.Statement,
) *common.FuncDeclarationStatement {
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

func callExpression(name string, arguments ...common.Expression) *common.CallExpression {
	return common.NewCallExpression(token(common.IDENTIFIER, name), append([]common.Expression{}, arguments...))
}

func TestFuncDeclarationStatement(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"without parameters and without return type",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{funcDeclarationStatement("f", nil, types.Void)},
		},
		{
			"with one parameter and return type",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "doble"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.CLOSED_PAR, ")"),
				token(common.ARROW, "->"),
				token(common.INT_TYPE, "Int"),
				token(common.OPEN_BRACE, "{"),
				token(common.RETURN, "return"),
				token(common.IDENTIFIER, "x"),
				token(common.STAR, "*"),
				token(common.INTEGER, "2"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{funcDeclarationStatement("doble", []common.Parameter{parameter("x", types.Int)}, types.Int,
				returnStatement(common.NewBinaryExpression(variableExpression("x"), token(common.STAR, "*"), integerLiteralExpression("2", 2))),
			)},
		},
		{
			"with several parameters of every type",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "a"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.COMMA, ","),
				token(common.IDENTIFIER, "b"),
				token(common.COLON, ":"),
				token(common.FLOAT_TYPE, "Float"),
				token(common.COMMA, ","),
				token(common.IDENTIFIER, "c"),
				token(common.COLON, ":"),
				token(common.STRING_TYPE, "String"),
				token(common.COMMA, ","),
				token(common.IDENTIFIER, "d"),
				token(common.COLON, ":"),
				token(common.BOOL_TYPE, "Bool"),
				token(common.CLOSED_PAR, ")"),
				token(common.ARROW, "->"),
				token(common.BOOL_TYPE, "Bool"),
				token(common.OPEN_BRACE, "{"),
				token(common.RETURN, "return"),
				token(common.IDENTIFIER, "d"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{funcDeclarationStatement("f",
				[]common.Parameter{
					parameter("a", types.Int),
					parameter("b", types.Float),
					parameter("c", types.Str),
					parameter("d", types.Bool),
				},
				types.Bool,
				returnStatement(variableExpression("d")),
			)},
		},
		{
			"body with several statements",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.LET, "let"),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.PRINT, "print"),
				token(common.IDENTIFIER, "x"),
				token(common.SEMICOLON, ";"),
				token(common.RETURN, "return"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{funcDeclarationStatement("f", nil, types.Void,
				varDeclarationStatement("x", types.Int, integerLiteralExpression("1", 1)),
				common.NewPrintStatement(variableExpression("x")),
				returnStatement(nil),
			)},
		},
		{
			"nested function",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "g"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{funcDeclarationStatement("f", nil, types.Void, funcDeclarationStatement("g", nil, types.Void))},
		},
		{
			"function inside a block",
			tokensWithoutSemicolon(
				token(common.OPEN_BRACE, "{"),
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{blockStatement(funcDeclarationStatement("f", nil, types.Void))},
		},
		{
			"function followed by another statement",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.SEMICOLON, ";"),
			),
			[]common.Statement{
				funcDeclarationStatement("f", nil, types.Void),
				common.NewExpressionStatement(callExpression("f")),
			},
		},
		{
			"repeated parameter names are parsed",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.COMMA, ","),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.STRING_TYPE, "String"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{funcDeclarationStatement("f", []common.Parameter{parameter("x", types.Int), parameter("x", types.Str)}, types.Void)},
		},
		{
			"keeps the position of the func, name and parameter tokens",
			[]common.Token{
				tokenAt(common.FUNC, "func", 2, 1),
				tokenAt(common.IDENTIFIER, "f", 2, 6),
				token(common.OPEN_PAR, "("),
				tokenAt(common.IDENTIFIER, "x", 2, 8),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
				token(common.EOF, ""),
			},
			[]common.Statement{common.NewFuncDeclarationStatement(
				tokenAt(common.FUNC, "func", 2, 1),
				tokenAt(common.IDENTIFIER, "f", 2, 6),
				[]common.Parameter{common.NewParameter(tokenAt(common.IDENTIFIER, "x", 2, 8), types.Int)},
				types.Void,
				[]common.Statement{},
			)},
		},
	})
}

func TestFuncDeclarationStatementErrors(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{"only keyword", tokensWithoutSemicolon(token(common.FUNC, "func")), "[line 0, column 0] Expected identifier after func"},
		{"only keyword without EOF token", []common.Token{token(common.FUNC, "func")}, "[line 0, column 0] Expected identifier after func"},
		{
			"keyword instead of name",
			tokensWithoutSemicolon(token(common.FUNC, "func"), token(common.WHILE, "while"), token(common.OPEN_PAR, "(")),
			"[line 0, column 0] Expected identifier after func",
		},
		{
			"missing name",
			tokensWithoutSemicolon(token(common.FUNC, "func"), token(common.OPEN_PAR, "("), token(common.CLOSED_PAR, ")")),
			"[line 0, column 0] Expected identifier after func",
		},
		{
			"missing open parentheses",
			tokensWithoutSemicolon(token(common.FUNC, "func"), token(common.IDENTIFIER, "f"), token(common.OPEN_BRACE, "{")),
			"[line 0, column 0] Expected '(' after func f",
		},
		{
			"name followed by EOF",
			tokensWithoutSemicolon(token(common.FUNC, "func"), token(common.IDENTIFIER, "f")),
			"[line 0, column 0] Expected '(' after func f",
		},
		{
			"literal instead of parameter name",
			tokensWithoutSemicolon(token(common.FUNC, "func"), token(common.IDENTIFIER, "f"), token(common.OPEN_PAR, "("), token(common.INTEGER, "1")),
			"[line 0, column 0] Expected parameter name in func f",
		},
		{
			"type instead of parameter name",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.INT_TYPE, "Int"),
				token(common.CLOSED_PAR, ")"),
			),
			"[line 0, column 0] Expected parameter name in func f",
		},
		{
			"open parentheses followed by EOF",
			tokensWithoutSemicolon(token(common.FUNC, "func"), token(common.IDENTIFIER, "f"), token(common.OPEN_PAR, "(")),
			"[line 0, column 0] Expected parameter name in func f",
		},
		{
			"missing colon",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.INT_TYPE, "Int"),
				token(common.CLOSED_PAR, ")"),
			),
			"[line 0, column 0] Expected ':' after parameter x",
		},
		{
			"missing type",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.CLOSED_PAR, ")"),
			),
			"[line 0, column 0] Expected type after parameter x :",
		},
		{
			"identifier instead of type",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.IDENTIFIER, "int"),
				token(common.CLOSED_PAR, ")"),
			),
			"[line 0, column 0] Expected type after parameter x :",
		},
		{
			"colon followed by EOF",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
			),
			"[line 0, column 0] Expected type after parameter x :",
		},
		{
			"missing comma between parameters",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.IDENTIFIER, "y"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.CLOSED_PAR, ")"),
			),
			"[line 0, column 0] Expected ',' or ')' after parameter",
		},
		{
			"trailing comma",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.COMMA, ","),
				token(common.CLOSED_PAR, ")"),
			),
			"[line 0, column 0] Expected parameter name in func f",
		},
		{
			"missing closed parentheses",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Expected ',' or ')' after parameter",
		},
		{
			"parameter followed by EOF",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.COLON, ":"),
				token(common.INT_TYPE, "Int"),
			),
			"[line 0, column 0] Expected ',' or ')' after parameter",
		},
		{
			"arrow without return type",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.ARROW, "->"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Expected return type after '->'",
		},
		{
			"arrow followed by EOF",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.ARROW, "->"),
			),
			"[line 0, column 0] Expected return type after '->'",
		},
		{
			"identifier as return type",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.ARROW, "->"),
				token(common.IDENTIFIER, "Void"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Expected return type after '->'",
		},
		{
			"return type without arrow",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.INT_TYPE, "Int"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Expected block statement",
		},
		{
			"missing body",
			tokens(token(common.FUNC, "func"), token(common.IDENTIFIER, "f"), token(common.OPEN_PAR, "("), token(common.CLOSED_PAR, ")")),
			"[line 0, column 0] Expected block statement",
		},
		{
			"body not closed",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.PRINT, "print"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
			),
			"[line 0, column 0] Expected '}' after block",
		},
		{
			"error inside the body",
			tokensWithoutSemicolon(
				token(common.FUNC, "func"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.PRINT, "print"),
				token(common.INTEGER, "1"),
				token(common.CLOSED_BRACE, "}"),
			),
			"[line 0, column 0] Expected ';' after expression",
		},
	})
}

func TestFuncDeclarationStatementErrorPosition(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"points to the token after the keyword",
			[]common.Token{
				tokenAt(common.FUNC, "func", 1, 1),
				tokenAt(common.OPEN_PAR, "(", 1, 6),
				tokenAt(common.EOF, "", 1, 7),
			},
			"[line 1, column 6] Expected identifier after func",
		},
		{
			"points to the token after the name",
			[]common.Token{
				tokenAt(common.FUNC, "func", 1, 1),
				tokenAt(common.IDENTIFIER, "f", 1, 6),
				tokenAt(common.OPEN_BRACE, "{", 1, 8),
				tokenAt(common.EOF, "", 1, 9),
			},
			"[line 1, column 8] Expected '(' after func f",
		},
		{
			"points to the token after the parameter",
			[]common.Token{
				tokenAt(common.FUNC, "func", 1, 1),
				tokenAt(common.IDENTIFIER, "f", 1, 6),
				tokenAt(common.OPEN_PAR, "(", 1, 7),
				tokenAt(common.IDENTIFIER, "x", 1, 8),
				tokenAt(common.COLON, ":", 1, 9),
				tokenAt(common.INT_TYPE, "Int", 1, 11),
				tokenAt(common.IDENTIFIER, "y", 1, 15),
				tokenAt(common.EOF, "", 1, 16),
			},
			"[line 1, column 15] Expected ',' or ')' after parameter",
		},
		{
			"points to the token after the arrow",
			[]common.Token{
				tokenAt(common.FUNC, "func", 1, 1),
				tokenAt(common.IDENTIFIER, "f", 1, 6),
				tokenAt(common.OPEN_PAR, "(", 1, 7),
				tokenAt(common.CLOSED_PAR, ")", 1, 8),
				tokenAt(common.ARROW, "->", 1, 10),
				tokenAt(common.OPEN_BRACE, "{", 1, 13),
				tokenAt(common.EOF, "", 1, 14),
			},
			"[line 1, column 13] Expected return type after '->'",
		},
		{
			"points to the end when the body is missing",
			[]common.Token{
				tokenAt(common.FUNC, "func", 1, 1),
				tokenAt(common.IDENTIFIER, "f", 1, 6),
				tokenAt(common.OPEN_PAR, "(", 1, 7),
				tokenAt(common.CLOSED_PAR, ")", 1, 8),
				tokenAt(common.EOF, "", 2, 1),
			},
			"[line 2, column 1] Expected block statement",
		},
	})
}

func TestReturnStatement(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{"without value", tokens(token(common.RETURN, "return")), []common.Statement{returnStatement(nil)}},
		{"with a literal", tokens(token(common.RETURN, "return"), token(common.INTEGER, "1")), []common.Statement{returnStatement(integerLiteralExpression("1", 1))}},
		{
			"with a binary expression",
			tokens(token(common.RETURN, "return"), token(common.IDENTIFIER, "a"), token(common.PLUS, "+"), token(common.IDENTIFIER, "b")),
			[]common.Statement{returnStatement(common.NewBinaryExpression(variableExpression("a"), token(common.PLUS, "+"), variableExpression("b")))},
		},
		{
			"with an assignment",
			tokens(token(common.RETURN, "return"), token(common.IDENTIFIER, "x"), token(common.EQUAL, "="), token(common.INTEGER, "1")),
			[]common.Statement{returnStatement(varAssignmentExpression("x", integerLiteralExpression("1", 1)))},
		},
		{
			"with a call",
			tokens(token(common.RETURN, "return"), token(common.IDENTIFIER, "f"), token(common.OPEN_PAR, "("), token(common.CLOSED_PAR, ")")),
			[]common.Statement{returnStatement(callExpression("f"))},
		},
		{
			"outside a function is parsed",
			tokensWithoutSemicolon(
				token(common.WHILE, "while"),
				token(common.OPEN_PAR, "("),
				token(common.TRUE, "True"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.RETURN, "return"),
				token(common.SEMICOLON, ";"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{whileStatement(booleanLiteralExpression(common.TRUE, "True", true), blockStatement(returnStatement(nil)))},
		},
		{
			"keeps the position of the return token",
			[]common.Token{tokenAt(common.RETURN, "return", 4, 5), token(common.SEMICOLON, ";"), token(common.EOF, "")},
			[]common.Statement{common.NewReturnStatement(tokenAt(common.RETURN, "return", 4, 5), nil)},
		},
	})
}

func TestReturnStatementErrors(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{"only keyword", tokensWithoutSemicolon(token(common.RETURN, "return")), "[line 0, column 0] Expected ';' after 'return'"},
		{"only keyword without EOF token", []common.Token{token(common.RETURN, "return")}, "[line 0, column 0] Expected ';' after 'return'"},
		{
			"missing semicolon after value",
			tokensWithoutSemicolon(token(common.RETURN, "return"), token(common.INTEGER, "1")),
			"[line 0, column 0] Expected ';' after 'return'",
		},
		{
			"two values",
			tokens(token(common.RETURN, "return"), token(common.INTEGER, "1"), token(common.INTEGER, "2")),
			"[line 0, column 0] Expected ';' after 'return'",
		},
		{"invalid value", tokens(token(common.RETURN, "return"), token(common.INT_TYPE, "Int")), "[line 0, column 0] Invalid primary expression"},
		{
			"points to the token after the value",
			[]common.Token{
				tokenAt(common.RETURN, "return", 1, 1),
				tokenAt(common.INTEGER, "1", 1, 8),
				tokenAt(common.EOF, "", 2, 1),
			},
			"[line 2, column 1] Expected ';' after 'return'",
		},
	})
}

func TestCallExpression(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"without arguments",
			tokens(token(common.IDENTIFIER, "f"), token(common.OPEN_PAR, "("), token(common.CLOSED_PAR, ")")),
			statements(callExpression("f")),
		},
		{
			"with one argument",
			tokens(token(common.IDENTIFIER, "f"), token(common.OPEN_PAR, "("), token(common.INTEGER, "1"), token(common.CLOSED_PAR, ")")),
			statements(callExpression("f", integerLiteralExpression("1", 1))),
		},
		{
			"with several arguments",
			tokens(
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.COMMA, ","),
				token(common.STRING, `"a"`),
				token(common.COMMA, ","),
				token(common.IDENTIFIER, "x"),
				token(common.CLOSED_PAR, ")"),
			),
			statements(callExpression("f", integerLiteralExpression("1", 1), stringLiteralExpression(`"a"`, "a"), variableExpression("x"))),
		},
		{
			"with expressions as arguments",
			tokens(
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.PLUS, "+"),
				token(common.INTEGER, "2"),
				token(common.COMMA, ","),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.CLOSED_PAR, ")"),
				token(common.CLOSED_PAR, ")"),
			),
			statements(callExpression("f",
				common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2)),
				groupingExpression(variableExpression("x")),
			)),
		},
		{
			"with an assignment as argument",
			tokens(
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "x"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
				token(common.CLOSED_PAR, ")"),
			),
			statements(callExpression("f", varAssignmentExpression("x", integerLiteralExpression("1", 1)))),
		},
		{
			"nested calls",
			tokens(
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "g"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.COMMA, ","),
				token(common.IDENTIFIER, "h"),
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.CLOSED_PAR, ")"),
				token(common.CLOSED_PAR, ")"),
			),
			statements(callExpression("f", callExpression("g"), callExpression("h", integerLiteralExpression("1", 1)))),
		},
		{
			"call binds tighter than power",
			tokens(
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.DOUBLE_STAR, "**"),
				token(common.INTEGER, "2"),
			),
			statements(common.NewBinaryExpression(callExpression("f"), token(common.DOUBLE_STAR, "**"), integerLiteralExpression("2", 2))),
		},
		{
			"negated call",
			tokens(token(common.MINUS, "-"), token(common.IDENTIFIER, "f"), token(common.OPEN_PAR, "("), token(common.CLOSED_PAR, ")")),
			statements(common.NewUnaryExpression(token(common.MINUS, "-"), callExpression("f"))),
		},
		{
			"call in a binary expression",
			tokens(
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.STAR, "*"),
				token(common.IDENTIFIER, "g"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
			),
			statements(common.NewBinaryExpression(callExpression("f"), token(common.STAR, "*"), callExpression("g"))),
		},
		{
			"call assigned to a variable",
			tokens(
				token(common.IDENTIFIER, "x"),
				token(common.EQUAL, "="),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
			),
			statements(varAssignmentExpression("x", callExpression("f"))),
		},
		{
			"call in a print statement",
			tokensWithoutSemicolon(
				token(common.PRINT, "print"),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.SEMICOLON, ";"),
			),
			[]common.Statement{common.NewPrintStatement(callExpression("f"))},
		},
		{
			"call as a condition",
			tokensWithoutSemicolon(
				token(common.IF, "if"),
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_BRACE, "{"),
				token(common.CLOSED_BRACE, "}"),
			),
			[]common.Statement{ifStatement(callExpression("f"), blockStatement(), nil)},
		},
		{
			"keeps the position of the name token",
			[]common.Token{
				tokenAt(common.IDENTIFIER, "f", 3, 2),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.SEMICOLON, ";"),
				token(common.EOF, ""),
			},
			statements(common.NewCallExpression(tokenAt(common.IDENTIFIER, "f", 3, 2), []common.Expression{})),
		},
	})
}

func TestCallExpressionErrors(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"missing closed parentheses",
			tokens(token(common.IDENTIFIER, "f"), token(common.OPEN_PAR, "("), token(common.INTEGER, "1")),
			"[line 0, column 0] Expected ',' or ')' after argument",
		},
		{
			"open parentheses followed by EOF",
			tokensWithoutSemicolon(token(common.IDENTIFIER, "f"), token(common.OPEN_PAR, "(")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"missing comma between arguments",
			tokens(
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.INTEGER, "2"),
				token(common.CLOSED_PAR, ")"),
			),
			"[line 0, column 0] Expected ',' or ')' after argument",
		},
		{
			"trailing comma",
			tokens(
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.COMMA, ","),
				token(common.CLOSED_PAR, ")"),
			),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"only a comma",
			tokens(token(common.IDENTIFIER, "f"), token(common.OPEN_PAR, "("), token(common.COMMA, ","), token(common.CLOSED_PAR, ")")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"calling the result of a call",
			tokens(
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
			),
			"[line 0, column 0] Expected ';' after expression",
		},
		{
			"calling a grouping expression",
			tokens(
				token(common.OPEN_PAR, "("),
				token(common.IDENTIFIER, "f"),
				token(common.CLOSED_PAR, ")"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
			),
			"[line 0, column 0] Expected ';' after expression",
		},
		{
			"assigning to a call",
			tokens(
				token(common.IDENTIFIER, "f"),
				token(common.OPEN_PAR, "("),
				token(common.CLOSED_PAR, ")"),
				token(common.EQUAL, "="),
				token(common.INTEGER, "1"),
			),
			"[line 0, column 0] Expected ';' after expression",
		},
		{
			"points to the token after the argument",
			[]common.Token{
				tokenAt(common.IDENTIFIER, "f", 1, 1),
				tokenAt(common.OPEN_PAR, "(", 1, 2),
				tokenAt(common.INTEGER, "1", 1, 3),
				tokenAt(common.SEMICOLON, ";", 1, 4),
				tokenAt(common.EOF, "", 1, 5),
			},
			"[line 1, column 4] Expected ',' or ')' after argument",
		},
	})
}
