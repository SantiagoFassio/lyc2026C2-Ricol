package scanner_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/scanner"
)

type scannerTestCase struct {
	name           string
	code           string
	expectedTokens []common.Token
}

type scannerPositionTestCase struct {
	name           string
	code           string
	expectedTokens []common.Token
}

func token(tokenType common.TokenType, lexeme string) common.Token {
	return common.NewToken(tokenType, lexeme, common.Position{})
}

func tokenAt(tokenType common.TokenType, lexeme string, line int, column int) common.Token {
	return common.NewToken(tokenType, lexeme, common.Position{Line: line, Column: column})
}

func tokens(expectedTokens ...common.Token) []common.Token {
	return append(expectedTokens, token(common.EOF, ""))
}

func assertScan(t *testing.T, sourceCode string, expectedTokens []common.Token) {
	t.Helper()

	scannedTokens, err := scanner.NewScanner(sourceCode).Scan()

	if err != nil {
		t.Fatalf("scanner.Scan(%q) unexpected error: %v", sourceCode, err)
	}
	if diff := cmp.Diff(expectedTokens, scannedTokens, cmpopts.IgnoreFields(common.Token{}, "Position")); diff != "" {
		t.Errorf("scanner.Scan(%q) mismatch (-want +got):\n%s", sourceCode, diff)
	}
}

func assertScanPositions(t *testing.T, sourceCode string, expectedTokens []common.Token) {
	t.Helper()

	scannedTokens, err := scanner.NewScanner(sourceCode).Scan()

	if err != nil {
		t.Fatalf("scanner.Scan(%q) unexpected error: %v", sourceCode, err)
	}
	if diff := cmp.Diff(expectedTokens, scannedTokens); diff != "" {
		t.Errorf("scanner.Scan(%q) mismatch (-want +got):\n%s", sourceCode, diff)
	}
}

func runScanTestCases(t *testing.T, testCases []scannerTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertScan(t, testCase.code, testCase.expectedTokens)
		})
	}
}

func runScanPositionTestCases(t *testing.T, testCases []scannerPositionTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertScanPositions(t, testCase.code, testCase.expectedTokens)
		})
	}
}

func TestEmptySourceCode(t *testing.T) {
	assertScan(t, "", tokens())
}

func TestInteger(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"positive", "3", tokens(token(common.INTEGER, "3"))},
		{"zero", "0", tokens(token(common.INTEGER, "0"))},
	})
}

func TestFloat(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"positive without decimals", "3.0", tokens(token(common.FLOAT, "3.0"))},
		{"zero", "0.0", tokens(token(common.FLOAT, "0.0"))},
		{"positive with two decimals", "38.25", tokens(token(common.FLOAT, "38.25"))},
	})
}

func TestSkippedCharacters(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"space", " ", tokens()},
		{"carriage return", "\r", tokens()},
		{"line feed", "\n", tokens()},
		{"tabulation", "\t", tokens()},
	})
}

func TestComments(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"only comment", "@ This is a comment", tokens()},
		{"literal and comment", "3 @ This is a comment", tokens(token(common.INTEGER, "3"))},
	})
}

func TestSpecialCharacters(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"plus", "+", tokens(token(common.PLUS, "+"))},
		{"minus", "-", tokens(token(common.MINUS, "-"))},
		{"star", "*", tokens(token(common.STAR, "*"))},
		{"slash", "/", tokens(token(common.SLASH, "/"))},
		{"percentage", "%", tokens(token(common.PERCENTAGE, "%"))},
		{"semicolon", ";", tokens(token(common.SEMICOLON, ";"))},
		{"open parentheses", "(", tokens(token(common.OPEN_PAR, "("))},
		{"closed parentheses", ")", tokens(token(common.CLOSED_PAR, ")"))},
	})
}

func TestDoubleSpecialCharacters(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"double star", "**", tokens(token(common.DOUBLE_STAR, "**"))},
		{"double slash", "//", tokens(token(common.DOUBLE_SLASH, "//"))},
	})
}

func TestInvalidCharacter(t *testing.T) {
	sourceCode := "?"
	expectedMessage := "Non-recognizable character '?'"

	_, err := scanner.NewScanner(sourceCode).Scan()

	if err == nil {
		t.Fatalf("scanner.Scan(%q) = nil error; want %q", sourceCode, expectedMessage)
	}
	if err.Error() != expectedMessage {
		t.Errorf("scanner.Scan(%q) error = %q; want %q", sourceCode, err, expectedMessage)
	}
}

func TestScanPositions(t *testing.T) {
	runScanPositionTestCases(t, []scannerPositionTestCase{
		{"empty source code", "", []common.Token{
			tokenAt(common.EOF, "", 1, 1),
		}},
		{"single line", "1 + 2;", []common.Token{
			tokenAt(common.INTEGER, "1", 1, 1),
			tokenAt(common.PLUS, "+", 1, 3),
			tokenAt(common.INTEGER, "2", 1, 5),
			tokenAt(common.SEMICOLON, ";", 1, 6),
			tokenAt(common.EOF, "", 1, 7),
		}},
		{"number at the start of its lexeme", "38.25 + 1;", []common.Token{
			tokenAt(common.FLOAT, "38.25", 1, 1),
			tokenAt(common.PLUS, "+", 1, 7),
			tokenAt(common.INTEGER, "1", 1, 9),
			tokenAt(common.SEMICOLON, ";", 1, 10),
			tokenAt(common.EOF, "", 1, 11),
		}},
		{"double special characters", "2 ** 3 // 4;", []common.Token{
			tokenAt(common.INTEGER, "2", 1, 1),
			tokenAt(common.DOUBLE_STAR, "**", 1, 3),
			tokenAt(common.INTEGER, "3", 1, 6),
			tokenAt(common.DOUBLE_SLASH, "//", 1, 8),
			tokenAt(common.INTEGER, "4", 1, 11),
			tokenAt(common.SEMICOLON, ";", 1, 12),
			tokenAt(common.EOF, "", 1, 13),
		}},
		{"third line after a comment", "@ This is a comment\n3 + 4;\n  38.25;", []common.Token{
			tokenAt(common.INTEGER, "3", 2, 1),
			tokenAt(common.PLUS, "+", 2, 3),
			tokenAt(common.INTEGER, "4", 2, 5),
			tokenAt(common.SEMICOLON, ";", 2, 6),
			tokenAt(common.FLOAT, "38.25", 3, 3),
			tokenAt(common.SEMICOLON, ";", 3, 8),
			tokenAt(common.EOF, "", 3, 9),
		}},
		{"comment at the end of a line", "1;@ This is a comment\n2;", []common.Token{
			tokenAt(common.INTEGER, "1", 1, 1),
			tokenAt(common.SEMICOLON, ";", 1, 2),
			tokenAt(common.INTEGER, "2", 2, 1),
			tokenAt(common.SEMICOLON, ";", 2, 2),
			tokenAt(common.EOF, "", 2, 3),
		}},
		{"comment at the end of the source code", "3 @ This is a comment", []common.Token{
			tokenAt(common.INTEGER, "3", 1, 1),
			tokenAt(common.EOF, "", 1, 22),
		}},
		{"carriage return and line feed", "1;\r\n2;", []common.Token{
			tokenAt(common.INTEGER, "1", 1, 1),
			tokenAt(common.SEMICOLON, ";", 1, 2),
			tokenAt(common.INTEGER, "2", 2, 1),
			tokenAt(common.SEMICOLON, ";", 2, 2),
			tokenAt(common.EOF, "", 2, 3),
		}},
		{"source code ending with a line feed", "3;\n", []common.Token{
			tokenAt(common.INTEGER, "3", 1, 1),
			tokenAt(common.SEMICOLON, ";", 1, 2),
			tokenAt(common.EOF, "", 2, 1),
		}},
		{"empty line between statements", "1;\n\n2;", []common.Token{
			tokenAt(common.INTEGER, "1", 1, 1),
			tokenAt(common.SEMICOLON, ";", 1, 2),
			tokenAt(common.INTEGER, "2", 3, 1),
			tokenAt(common.SEMICOLON, ";", 3, 2),
			tokenAt(common.EOF, "", 3, 3),
		}},
	})
}
