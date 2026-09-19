package scanner_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/scanner"
)

type scannerTestCase struct {
	name           string
	code           string
	expectedTokens []common.Token
}

func token(tokenType common.TokenType, lexeme string) common.Token {
	return common.Token{TokenType: tokenType, Lexeme: lexeme}
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
