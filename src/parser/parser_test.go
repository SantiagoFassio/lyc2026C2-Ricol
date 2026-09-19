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
	return common.Token{TokenType: tokenType, Lexeme: lexeme}
}

func tokens(inputTokens ...common.Token) []common.Token {
	return append(inputTokens, token(common.SEMICOLON, ";"), token(common.EOF, ""))
}

func tokensWithoutSemicolon(inputTokens ...common.Token) []common.Token {
	return append(inputTokens, token(common.EOF, ""))
}

func integerLiteralExpression(lexeme string, value int64) *common.LiteralExpression {
	return &common.LiteralExpression{Token: token(common.INTEGER, lexeme), Value: types.NewInteger(value)}
}

func floatLiteralExpression(lexeme string, value float64) *common.LiteralExpression {
	return &common.LiteralExpression{Token: token(common.FLOAT, lexeme), Value: types.NewFloat(value)}
}

func binaryExpression(leftExpression common.Expression, operator common.Token, rightExpression common.Expression) *common.BinaryExpression {
	return &common.BinaryExpression{
		LeftExpression:  leftExpression,
		Operator:        operator,
		RightExpression: rightExpression,
	}
}

func unaryExpression(operator common.Token, expression common.Expression) *common.UnaryExpression {
	return &common.UnaryExpression{Operator: operator, Expression: expression}
}

func groupingExpression(expression common.Expression) *common.GroupingExpression {
	return &common.GroupingExpression{Expression: expression}
}

func statements(expressions ...common.Expression) []common.Statement {
	expectedStatements := []common.Statement{}
	for _, expression := range expressions {
		expectedStatements = append(expectedStatements, &common.ExpressionStatement{Expression: expression})
	}
	return expectedStatements
}

func assertParse(t *testing.T, tokens []common.Token, expectedStatements []common.Statement) {
	t.Helper()

	parsedStatements, err := parser.NewParser(tokens).Parse()

	if err != nil {
		t.Fatalf("parser.Parse(%q) unexpected error: %v", tokens, err)
	}
	if diff := cmp.Diff(expectedStatements, parsedStatements, cmpopts.EquateComparable(types.Number{})); diff != "" {
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

func TestSimpleMathOperations(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"negation",
			tokens(token(common.MINUS, "-"), token(common.FLOAT, "25.11")),
			statements(unaryExpression(token(common.MINUS, "-"), floatLiteralExpression("25.11", 25.11))),
		},
		{
			"addition",
			tokens(token(common.INTEGER, "3"), token(common.PLUS, "+"), token(common.FLOAT, "8.5")),
			statements(binaryExpression(integerLiteralExpression("3", 3), token(common.PLUS, "+"), floatLiteralExpression("8.5", 8.5))),
		},
		{
			"subtraction",
			tokens(token(common.INTEGER, "3"), token(common.MINUS, "-"), token(common.FLOAT, "8.5")),
			statements(binaryExpression(integerLiteralExpression("3", 3), token(common.MINUS, "-"), floatLiteralExpression("8.5", 8.5))),
		},
		{
			"multiplication",
			tokens(token(common.INTEGER, "3"), token(common.STAR, "*"), token(common.FLOAT, "8.5")),
			statements(binaryExpression(integerLiteralExpression("3", 3), token(common.STAR, "*"), floatLiteralExpression("8.5", 8.5))),
		},
		{
			"division",
			tokens(token(common.INTEGER, "3"), token(common.SLASH, "/"), token(common.FLOAT, "8.5")),
			statements(binaryExpression(integerLiteralExpression("3", 3), token(common.SLASH, "/"), floatLiteralExpression("8.5", 8.5))),
		},
		{
			"integer division",
			tokens(token(common.INTEGER, "3"), token(common.DOUBLE_SLASH, "//"), token(common.FLOAT, "8.5")),
			statements(binaryExpression(integerLiteralExpression("3", 3), token(common.DOUBLE_SLASH, "//"), floatLiteralExpression("8.5", 8.5))),
		},
		{
			"modulo",
			tokens(token(common.INTEGER, "3"), token(common.PERCENTAGE, "%"), token(common.FLOAT, "8.5")),
			statements(binaryExpression(integerLiteralExpression("3", 3), token(common.PERCENTAGE, "%"), floatLiteralExpression("8.5", 8.5))),
		},
		{
			"exponentiation",
			tokens(token(common.INTEGER, "3"), token(common.DOUBLE_STAR, "**"), token(common.FLOAT, "8.5")),
			statements(binaryExpression(integerLiteralExpression("3", 3), token(common.DOUBLE_STAR, "**"), floatLiteralExpression("8.5", 8.5))),
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
			statements(binaryExpression(
				integerLiteralExpression("1", 1),
				token(common.PLUS, "+"),
				binaryExpression(integerLiteralExpression("2", 2), token(common.STAR, "*"), integerLiteralExpression("3", 3)),
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
			statements(binaryExpression(
				binaryExpression(integerLiteralExpression("8", 8), token(common.SLASH, "/"), integerLiteralExpression("4", 4)),
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
			statements(binaryExpression(
				binaryExpression(integerLiteralExpression("7", 7), token(common.PERCENTAGE, "%"), integerLiteralExpression("4", 4)),
				token(common.STAR, "*"),
				integerLiteralExpression("2", 2),
			)),
		},
		{
			"integer division and division share precedence",
			tokens(
				token(common.INTEGER, "8"),
				token(common.DOUBLE_SLASH, "//"),
				token(common.INTEGER, "3"),
				token(common.SLASH, "/"),
				token(common.INTEGER, "2"),
			),
			statements(binaryExpression(
				binaryExpression(integerLiteralExpression("8", 8), token(common.DOUBLE_SLASH, "//"), integerLiteralExpression("3", 3)),
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
			statements(binaryExpression(
				binaryExpression(integerLiteralExpression("2", 2), token(common.DOUBLE_STAR, "**"), integerLiteralExpression("3", 3)),
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
			statements(unaryExpression(
				token(common.MINUS, "-"),
				binaryExpression(integerLiteralExpression("3", 3), token(common.DOUBLE_STAR, "**"), integerLiteralExpression("2", 2)),
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
			statements(binaryExpression(
				unaryExpression(token(common.MINUS, "-"), integerLiteralExpression("3", 3)),
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
			statements(binaryExpression(
				binaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2)),
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
			statements(binaryExpression(
				binaryExpression(integerLiteralExpression("10", 10), token(common.MINUS, "-"), integerLiteralExpression("3", 3)),
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
			statements(binaryExpression(
				binaryExpression(integerLiteralExpression("16", 16), token(common.SLASH, "/"), integerLiteralExpression("4", 4)),
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
			statements(binaryExpression(
				integerLiteralExpression("2", 2),
				token(common.DOUBLE_STAR, "**"),
				binaryExpression(integerLiteralExpression("3", 3), token(common.DOUBLE_STAR, "**"), integerLiteralExpression("2", 2)),
			)),
		},
	})
}

func TestUnaryExpressions(t *testing.T) {
	runParseTestCases(t, []parserTestCase{
		{
			"double negation",
			tokens(token(common.MINUS, "-"), token(common.MINUS, "-"), token(common.INTEGER, "3")),
			statements(unaryExpression(
				token(common.MINUS, "-"),
				unaryExpression(token(common.MINUS, "-"), integerLiteralExpression("3", 3)),
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
			statements(unaryExpression(
				token(common.MINUS, "-"),
				unaryExpression(
					token(common.MINUS, "-"),
					unaryExpression(token(common.MINUS, "-"), floatLiteralExpression("1.5", 1.5)),
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
			statements(unaryExpression(
				token(common.MINUS, "-"),
				groupingExpression(binaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2))),
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
			statements(binaryExpression(
				integerLiteralExpression("3", 3),
				token(common.STAR, "*"),
				unaryExpression(token(common.MINUS, "-"), integerLiteralExpression("2", 2)),
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
			statements(binaryExpression(
				integerLiteralExpression("2", 2),
				token(common.DOUBLE_STAR, "**"),
				unaryExpression(token(common.MINUS, "-"), integerLiteralExpression("3", 3)),
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
			statements(binaryExpression(
				integerLiteralExpression("1", 1),
				token(common.PLUS, "+"),
				unaryExpression(token(common.MINUS, "-"), integerLiteralExpression("2", 2)),
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
			statements(binaryExpression(
				groupingExpression(binaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2))),
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
			statements(binaryExpression(
				groupingExpression(binaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2))),
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
			statements(groupingExpression(unaryExpression(token(common.MINUS, "-"), integerLiteralExpression("3", 3)))),
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
				binaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2)),
				groupingExpression(floatLiteralExpression("3.5", 3.5)),
				unaryExpression(token(common.MINUS, "-"), integerLiteralExpression("4", 4)),
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
		statements(binaryExpression(
			binaryExpression(
				binaryExpression(
					binaryExpression(
						unaryExpression(
							token(common.MINUS, "-"),
							groupingExpression(binaryExpression(integerLiteralExpression("1", 1), token(common.PLUS, "+"), integerLiteralExpression("2", 2))),
						),
						token(common.STAR, "*"),
						binaryExpression(integerLiteralExpression("3", 3), token(common.DOUBLE_STAR, "**"), integerLiteralExpression("2", 2)),
					),
					token(common.DOUBLE_SLASH, "//"),
					integerLiteralExpression("4", 4),
				),
				token(common.PERCENTAGE, "%"),
				integerLiteralExpression("5", 5),
			),
			token(common.MINUS, "-"),
			binaryExpression(integerLiteralExpression("6", 6), token(common.SLASH, "/"), floatLiteralExpression("7.5", 7.5)),
		)),
	)
}

func TestMissingSemicolon(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"no semicolon at all",
			tokensWithoutSemicolon(token(common.INTEGER, "1")),
			"Expected ';' after expression",
		},
		{
			"no semicolon between statements",
			tokensWithoutSemicolon(
				token(common.INTEGER, "1"),
				token(common.INTEGER, "2"),
				token(common.SEMICOLON, ";"),
			),
			"Expected ';' after expression",
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
			"Expected ';' after expression",
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
			"Grouping expression without close",
		},
		{
			"EOF before closing parentheses",
			tokensWithoutSemicolon(token(common.OPEN_PAR, "("), token(common.INTEGER, "1")),
			"Grouping expression without close",
		},
		{
			"only inner grouping closed",
			tokens(
				token(common.OPEN_PAR, "("),
				token(common.OPEN_PAR, "("),
				token(common.INTEGER, "1"),
				token(common.CLOSED_PAR, ")"),
			),
			"Grouping expression without close",
		},
	})
}

func TestInvalidPrimaryExpression(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"only semicolon",
			tokens(),
			"Invalid primary expression",
		},
		{
			"empty grouping",
			tokens(token(common.OPEN_PAR, "("), token(common.CLOSED_PAR, ")")),
			"Invalid primary expression",
		},
		{
			"missing right operand",
			tokens(token(common.INTEGER, "1"), token(common.PLUS, "+")),
			"Invalid primary expression",
		},
		{
			"missing left operand",
			tokens(token(common.STAR, "*"), token(common.INTEGER, "3")),
			"Invalid primary expression",
		},
		{
			"missing exponent",
			tokens(token(common.INTEGER, "2"), token(common.DOUBLE_STAR, "**")),
			"Invalid primary expression",
		},
		{
			"dangling unary minus",
			tokens(token(common.MINUS, "-")),
			"Invalid primary expression",
		},
		{
			"consecutive operators",
			tokens(token(common.INTEGER, "1"), token(common.PLUS, "+"), token(common.STAR, "*"), token(common.INTEGER, "2")),
			"Invalid primary expression",
		},
		{
			"unsupported string literal",
			tokens(token(common.STRING, "\"hello\"")),
			"Invalid primary expression",
		},
		{
			"unsupported dot",
			tokens(token(common.DOT, ".")),
			"Invalid primary expression",
		},
		{
			"error in a later statement discards the previous ones",
			tokensWithoutSemicolon(
				token(common.INTEGER, "1"),
				token(common.SEMICOLON, ";"),
				token(common.PLUS, "+"),
				token(common.SEMICOLON, ";"),
			),
			"Invalid primary expression",
		},
	})
}

func TestInvalidLiteralValue(t *testing.T) {
	runParseErrorTestCases(t, []parserErrorTestCase{
		{
			"integer out of range",
			tokens(token(common.INTEGER, "9223372036854775808")),
			"Invalid integer: 9223372036854775808",
		},
		{
			"negated integer out of range",
			tokens(token(common.MINUS, "-"), token(common.INTEGER, "9223372036854775808")),
			"Invalid integer: 9223372036854775808",
		},
		{
			"integer with non-numeric characters",
			tokens(token(common.INTEGER, "12a")),
			"Invalid integer: 12a",
		},
		{
			"empty integer lexeme",
			tokens(token(common.INTEGER, "")),
			"Invalid integer: ",
		},
		{
			"float with two dots",
			tokens(token(common.FLOAT, "1.2.3")),
			"Invalid float: 1.2.3",
		},
		{
			"float out of range",
			tokens(token(common.FLOAT, "1e400")),
			"Invalid float: 1e400",
		},
	})
}
