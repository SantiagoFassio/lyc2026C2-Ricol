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
	if diff := cmp.Diff(expectedStatements, parsedStatements, cmpopts.EquateComparable(types.Number{}, types.String{}, types.Boolean{})); diff != "" {
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
			tokens(token(common.PRINT, "PRINT"), token(common.INTEGER, "3")),
			[]common.Statement{common.NewPrintStatement(integerLiteralExpression("3", 3))},
		},
		{
			"string literal",
			tokens(token(common.PRINT, "PRINT"), token(common.STRING, `"hola"`)),
			[]common.Statement{common.NewPrintStatement(stringLiteralExpression(`"hola"`, "hola"))},
		},
		{
			"binary expression",
			tokens(token(common.PRINT, "PRINT"), token(common.INTEGER, "1"), token(common.PLUS, "+"), token(common.INTEGER, "2")),
			[]common.Statement{common.NewPrintStatement(
				common.NewBinaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2)),
			)},
		},
		{
			"grouping expression",
			tokens(token(common.PRINT, "PRINT"), token(common.OPEN_PAR, "("), token(common.INTEGER, "1"), token(common.CLOSED_PAR, ")")),
			[]common.Statement{common.NewPrintStatement(groupingExpression(integerLiteralExpression("1", 1)))},
		},
		{
			"negation",
			tokens(token(common.PRINT, "PRINT"), token(common.MINUS, "-"), token(common.FLOAT, "2.5")),
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
				token(common.PRINT, "PRINT"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.PRINT, "PRINT"),
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
				token(common.PRINT, "PRINT"),
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
			tokensWithoutSemicolon(token(common.PRINT, "PRINT")),
			"[line 0, column 0] Expected expression after 'PRINT'",
		},
		{
			"keyword followed by a semicolon",
			tokens(token(common.PRINT, "PRINT")),
			"[line 0, column 0] Expected expression after 'PRINT'",
		},
		{
			"keyword without EOF token",
			[]common.Token{token(common.PRINT, "PRINT")},
			"[line 0, column 0] Expected expression after 'PRINT'",
		},
		{
			"missing semicolon",
			tokensWithoutSemicolon(token(common.PRINT, "PRINT"), token(common.INTEGER, "1")),
			"[line 0, column 0] Expected ';' after expression",
		},
		{
			"invalid expression",
			tokens(token(common.PRINT, "PRINT"), token(common.CLOSED_PAR, ")")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"two consecutive keywords",
			tokens(token(common.PRINT, "PRINT"), token(common.PRINT, "PRINT"), token(common.INTEGER, "1")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"keyword inside an expression",
			tokens(token(common.INTEGER, "1"), token(common.PLUS, "+"), token(common.PRINT, "PRINT"), token(common.INTEGER, "2")),
			"[line 0, column 0] Invalid primary expression",
		},
		{
			"error in a print statement discards the previous ones",
			tokensWithoutSemicolon(
				token(common.PRINT, "PRINT"),
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.PRINT, "PRINT"),
				token(common.SEMICOLON, ";"),
			),
			"[line 0, column 0] Expected expression after 'PRINT'",
		},
	})
}

func TestPrintStatementErrorPosition(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"points to the semicolon after the keyword",
			[]common.Token{tokenAt(common.PRINT, "PRINT", 1, 1), tokenAt(common.SEMICOLON, ";", 1, 6), tokenAt(common.EOF, "", 1, 7)},
			"[line 1, column 6] Expected expression after 'PRINT'",
		},
		{
			"points to the EOF after the keyword",
			[]common.Token{tokenAt(common.PRINT, "PRINT", 1, 1), tokenAt(common.EOF, "", 1, 6)},
			"[line 1, column 6] Expected expression after 'PRINT'",
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
				token(common.PRINT, "PRINT"),
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
				token(common.PRINT, "PRINT"),
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
				token(common.PRINT, "PRINT"),
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
				token(common.PRINT, "PRINT"),
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
				token(common.PRINT, "PRINT"),
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
				token(common.PRINT, "PRINT"),
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
				tokenAt(common.PRINT, "PRINT", 1, 11),
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
				tokenAt(common.PRINT, "PRINT", 1, 19),
				tokenAt(common.INTEGER, "1", 1, 25),
				tokenAt(common.SEMICOLON, ";", 1, 26),
				tokenAt(common.EOF, "", 1, 27),
			},
			"[line 1, column 19] Expected '{' or 'if' after 'else'",
		},
	})
}
