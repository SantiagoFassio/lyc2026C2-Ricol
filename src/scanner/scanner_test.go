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

type scannerIdentifierTestCase struct {
	name   string
	code   string
	lexeme string
	line   int
	column int
}

type scannerErrorTestCase struct {
	name            string
	code            string
	expectedMessage string
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

func assertScanError(t *testing.T, sourceCode string, expectedMessage string) {
	t.Helper()

	_, err := scanner.NewScanner(sourceCode).Scan()

	if err == nil {
		t.Fatalf("scanner.Scan(%q) = nil error; want %q", sourceCode, expectedMessage)
	}
	if err.Error() != expectedMessage {
		t.Errorf("scanner.Scan(%q) error = %q; want %q", sourceCode, err, expectedMessage)
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

func assertScanIdentifier(t *testing.T, sourceCode string, expectedToken common.Token) {
	t.Helper()

	scannedTokens, err := scanner.NewScanner(sourceCode).Scan()

	if err != nil {
		t.Fatalf("scanner.Scan(%q) unexpected error: %v", sourceCode, err)
	}
	for _, scannedToken := range scannedTokens {
		if scannedToken == expectedToken {
			return
		}
	}
	t.Errorf("scanner.Scan(%q) = %v; want it to contain %v at %v", sourceCode, scannedTokens, expectedToken, expectedToken.Position)
}

func runScanIdentifierTestCases(t *testing.T, testCases []scannerIdentifierTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			expectedToken := tokenAt(common.IDENTIFIER, testCase.lexeme, testCase.line, testCase.column)
			assertScanIdentifier(t, testCase.code, expectedToken)
		})
	}
}

func runScanErrorTestCases(t *testing.T, testCases []scannerErrorTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertScanError(t, testCase.code, testCase.expectedMessage)
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

func TestString(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"simple", `"hello"`, tokens(token(common.STRING, `"hello"`))},
		{"empty", `""`, tokens(token(common.STRING, `""`))},
		{"with comment character", `"a @ b"`, tokens(token(common.STRING, `"a @ b"`))},
		{"with operators and semicolon", `"1 + 2;"`, tokens(token(common.STRING, `"1 + 2;"`))},
		{"with unicode characters", `"ñandú"`, tokens(token(common.STRING, `"ñandú"`))},
		{"escaped double quote", `"dijo \"hola\""`, tokens(token(common.STRING, `"dijo \"hola\""`))},
		{"escaped backslash", `"a\\b"`, tokens(token(common.STRING, `"a\\b"`))},
		{"escaped line feed", `"a\nb"`, tokens(token(common.STRING, `"a\nb"`))},
		{"escaped tabulation", `"a\tb"`, tokens(token(common.STRING, `"a\tb"`))},
		{"escaped backslash before closing double quote", `"a\\"`, tokens(token(common.STRING, `"a\\"`))},
		{"two strings", `"a" "b"`, tokens(
			token(common.STRING, `"a"`),
			token(common.STRING, `"b"`),
		)},
		{"followed by other tokens", `"1 + 2" + 3;`, tokens(
			token(common.STRING, `"1 + 2"`),
			token(common.PLUS, "+"),
			token(common.INTEGER, "3"),
			token(common.SEMICOLON, ";"),
		)},
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

func TestBoolean(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"true", "True", tokens(token(common.TRUE, "True"))},
		{"false", "False", tokens(token(common.FALSE, "False"))},
		{"followed by a semicolon", "True;", tokens(token(common.TRUE, "True"), token(common.SEMICOLON, ";"))},
		{"inside a grouping", "(False)", tokens(
			token(common.OPEN_PAR, "("),
			token(common.FALSE, "False"),
			token(common.CLOSED_PAR, ")"),
		)},
		{"two booleans separated by an operator", "True+False", tokens(
			token(common.TRUE, "True"),
			token(common.PLUS, "+"),
			token(common.FALSE, "False"),
		)},
		{"inside a comment", "@ True", tokens()},
		{"inside a string", `"True"`, tokens(token(common.STRING, `"True"`))},
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
		{"open brace", "{", tokens(token(common.OPEN_BRACE, "{"))},
		{"closed brace", "}", tokens(token(common.CLOSED_BRACE, "}"))},
	})
}

func TestDoubleSpecialCharacters(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"double star", "**", tokens(token(common.DOUBLE_STAR, "**"))},
		{"double slash", "//", tokens(token(common.DOUBLE_SLASH, "//"))},
		{"double equal", "==", tokens(token(common.DOUBLE_EQUAL, "=="))},
		{"not equal", "!=", tokens(token(common.NOT_EQUAL, "!="))},
		{"less equal", "<=", tokens(token(common.LESS_EQUAL, "<="))},
		{"greater equal", ">=", tokens(token(common.GREATER_EQUAL, ">="))},
	})
}

func TestComparisonOperators(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"less", "<", tokens(token(common.LESS, "<"))},
		{"greater", ">", tokens(token(common.GREATER, ">"))},
		{"less without spaces", "1<2", tokens(
			token(common.INTEGER, "1"),
			token(common.LESS, "<"),
			token(common.INTEGER, "2"),
		)},
		{"greater equal without spaces", "1>=2", tokens(
			token(common.INTEGER, "1"),
			token(common.GREATER_EQUAL, ">="),
			token(common.INTEGER, "2"),
		)},
		{"less followed by a negation", "1 < -2", tokens(
			token(common.INTEGER, "1"),
			token(common.LESS, "<"),
			token(common.MINUS, "-"),
			token(common.INTEGER, "2"),
		)},
		{"two comparisons", "1 < 2 <= 3", tokens(
			token(common.INTEGER, "1"),
			token(common.LESS, "<"),
			token(common.INTEGER, "2"),
			token(common.LESS_EQUAL, "<="),
			token(common.INTEGER, "3"),
		)},
		{"greater than an equality", "1 > 2 == False", tokens(
			token(common.INTEGER, "1"),
			token(common.GREATER, ">"),
			token(common.INTEGER, "2"),
			token(common.DOUBLE_EQUAL, "=="),
			token(common.FALSE, "False"),
		)},
		{"inside a string", `"1 <= 2 != 3"`, tokens(token(common.STRING, `"1 <= 2 != 3"`))},
		{"inside a comment", "@ 1 != 2", tokens()},
	})
}

func TestSingleExclamation(t *testing.T) {
	runScanErrorTestCases(t, []scannerErrorTestCase{
		{"alone", "!", "[line 1, column 1] Non-recognizable character '!'"},
		{"before a number", "!1;", "[line 1, column 1] Non-recognizable character '!'"},
		{"separated from the equal", "1 ! = 2;", "[line 1, column 3] Non-recognizable character '!'"},
		{"before a boolean", "!True;", "[line 1, column 1] Non-recognizable character '!'"},
	})
}

func TestEquality(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"between integers without spaces", "1==2", tokens(
			token(common.INTEGER, "1"),
			token(common.DOUBLE_EQUAL, "=="),
			token(common.INTEGER, "2"),
		)},
		{"between booleans", "True == False", tokens(
			token(common.TRUE, "True"),
			token(common.DOUBLE_EQUAL, "=="),
			token(common.FALSE, "False"),
		)},
		{"inside a string", `"=="`, tokens(token(common.STRING, `"=="`))},
		{"inside a comment", "@ 1 == 2", tokens()},
	})
}

func TestSingleEqual(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"alone", "=", tokens(token(common.EQUAL, "="))},
		{"between numbers", "1 = 2;", tokens(
			token(common.INTEGER, "1"),
			token(common.EQUAL, "="),
			token(common.INTEGER, "2"),
			token(common.SEMICOLON, ";"),
		)},
		{"separated equals", "1 = = 2;", tokens(
			token(common.INTEGER, "1"),
			token(common.EQUAL, "="),
			token(common.EQUAL, "="),
			token(common.INTEGER, "2"),
			token(common.SEMICOLON, ";"),
		)},
		{"three equals", "1 === 2;", tokens(
			token(common.INTEGER, "1"),
			token(common.DOUBLE_EQUAL, "=="),
			token(common.EQUAL, "="),
			token(common.INTEGER, "2"),
			token(common.SEMICOLON, ";"),
		)},
	})
}

func TestInvalidCharacter(t *testing.T) {
	sourceCode := "?"
	expectedMessage := "[line 1, column 1] Non-recognizable character '?'"

	_, err := scanner.NewScanner(sourceCode).Scan()

	if err == nil {
		t.Fatalf("scanner.Scan(%q) = nil error; want %q", sourceCode, expectedMessage)
	}
	if err.Error() != expectedMessage {
		t.Errorf("scanner.Scan(%q) error = %q; want %q", sourceCode, err, expectedMessage)
	}
}

func TestKeywordLookalikes(t *testing.T) {
	runScanIdentifierTestCases(t, []scannerIdentifierTestCase{
		{"unknown word", "verdadero;", "verdadero", 1, 1},
		{"lowercase true", "true;", "true", 1, 1},
		{"keyword with a digit", "True1;", "True1", 1, 1},
		{"two keywords without separation", "TrueFalse;", "TrueFalse", 1, 1},
		{"after other tokens", "1 + verdadero;", "verdadero", 1, 5},
		{"on the second line", "True;\n  verdadero;", "verdadero", 2, 3},
	})
}

func TestStringErrors(t *testing.T) {
	runScanErrorTestCases(t, []scannerErrorTestCase{
		{"unterminated at the end of the source code", `"hola`, "[line 1, column 1] Unterminated string"},
		{"unterminated at the end of the line", "\"hola\n2;", "[line 1, column 1] Unterminated string"},
		{"unterminated before a carriage return", "\"hola\r\n2;", "[line 1, column 1] Unterminated string"},
		{"unterminated after other tokens", `1 + "hola`, "[line 1, column 5] Unterminated string"},
		{"unterminated on the second line", "1;\n  \"hola", "[line 2, column 3] Unterminated string"},
		{"escaped closing double quote", `"hola\"`, "[line 1, column 1] Unterminated string"},
		{"backslash at the end of the source code", `"hola\`, "[line 1, column 1] Unterminated string"},
		{"backslash at the end of the line", "\"hola\\\n2;", "[line 1, column 1] Unterminated string"},
		{"invalid escape sequence", `"a\qb";`, `[line 1, column 3] Invalid escape sequence '\q'`},
		{"invalid escape sequence after unicode characters", `"ñ\q";`, `[line 1, column 3] Invalid escape sequence '\q'`},
		{"invalid escape sequence on the second line", "1;\n\"\\x\";", `[line 2, column 2] Invalid escape sequence '\x'`},
	})
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
		{"string followed by other tokens", `"a b" + 1;`, []common.Token{
			tokenAt(common.STRING, `"a b"`, 1, 1),
			tokenAt(common.PLUS, "+", 1, 7),
			tokenAt(common.INTEGER, "1", 1, 9),
			tokenAt(common.SEMICOLON, ";", 1, 10),
			tokenAt(common.EOF, "", 1, 11),
		}},
		{"string with escape sequences", `"\"\\" + 1;`, []common.Token{
			tokenAt(common.STRING, `"\"\\"`, 1, 1),
			tokenAt(common.PLUS, "+", 1, 8),
			tokenAt(common.INTEGER, "1", 1, 10),
			tokenAt(common.SEMICOLON, ";", 1, 11),
			tokenAt(common.EOF, "", 1, 12),
		}},
		{"string with unicode characters", `"ñandú" + 1;`, []common.Token{
			tokenAt(common.STRING, `"ñandú"`, 1, 1),
			tokenAt(common.PLUS, "+", 1, 9),
			tokenAt(common.INTEGER, "1", 1, 11),
			tokenAt(common.SEMICOLON, ";", 1, 12),
			tokenAt(common.EOF, "", 1, 13),
		}},
		{"double equal takes two columns", "1 == 2;", []common.Token{
			tokenAt(common.INTEGER, "1", 1, 1),
			tokenAt(common.DOUBLE_EQUAL, "==", 1, 3),
			tokenAt(common.INTEGER, "2", 1, 6),
			tokenAt(common.SEMICOLON, ";", 1, 7),
			tokenAt(common.EOF, "", 1, 8),
		}},
		{"comparison operators of one and two characters", "1 < 2 >= 3;", []common.Token{
			tokenAt(common.INTEGER, "1", 1, 1),
			tokenAt(common.LESS, "<", 1, 3),
			tokenAt(common.INTEGER, "2", 1, 5),
			tokenAt(common.GREATER_EQUAL, ">=", 1, 7),
			tokenAt(common.INTEGER, "3", 1, 10),
			tokenAt(common.SEMICOLON, ";", 1, 11),
			tokenAt(common.EOF, "", 1, 12),
		}},
		{"boolean followed by other tokens", "True + 1;", []common.Token{
			tokenAt(common.TRUE, "True", 1, 1),
			tokenAt(common.PLUS, "+", 1, 6),
			tokenAt(common.INTEGER, "1", 1, 8),
			tokenAt(common.SEMICOLON, ";", 1, 9),
			tokenAt(common.EOF, "", 1, 10),
		}},
		{"boolean on the second line", "True;\n  False;", []common.Token{
			tokenAt(common.TRUE, "True", 1, 1),
			tokenAt(common.SEMICOLON, ";", 1, 5),
			tokenAt(common.FALSE, "False", 2, 3),
			tokenAt(common.SEMICOLON, ";", 2, 8),
			tokenAt(common.EOF, "", 2, 9),
		}},
		{"string on the second line", "1;\n\"a\";", []common.Token{
			tokenAt(common.INTEGER, "1", 1, 1),
			tokenAt(common.SEMICOLON, ";", 1, 2),
			tokenAt(common.STRING, `"a"`, 2, 1),
			tokenAt(common.SEMICOLON, ";", 2, 4),
			tokenAt(common.EOF, "", 2, 5),
		}},
	})
}

func TestPrintKeyword(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"only keyword", "PRINT", tokens(token(common.PRINT, "PRINT"))},
		{"print statement", "PRINT 1;", tokens(
			token(common.PRINT, "PRINT"),
			token(common.INTEGER, "1"),
			token(common.SEMICOLON, ";"),
		)},
		{"followed by open parentheses", "PRINT(1);", tokens(
			token(common.PRINT, "PRINT"),
			token(common.OPEN_PAR, "("),
			token(common.INTEGER, "1"),
			token(common.CLOSED_PAR, ")"),
			token(common.SEMICOLON, ";"),
		)},
		{"followed by a string", `PRINT"a";`, tokens(
			token(common.PRINT, "PRINT"),
			token(common.STRING, `"a"`),
			token(common.SEMICOLON, ";"),
		)},
		{"followed by a semicolon", "PRINT;", tokens(
			token(common.PRINT, "PRINT"),
			token(common.SEMICOLON, ";"),
		)},
		{"followed by minus", "PRINT-1;", tokens(
			token(common.PRINT, "PRINT"),
			token(common.MINUS, "-"),
			token(common.INTEGER, "1"),
			token(common.SEMICOLON, ";"),
		)},
		{"followed by a tabulation", "PRINT\t1;", tokens(
			token(common.PRINT, "PRINT"),
			token(common.INTEGER, "1"),
			token(common.SEMICOLON, ";"),
		)},
		{"followed by a comment", "PRINT@ This is a comment", tokens(token(common.PRINT, "PRINT"))},
		{"two print statements", "PRINT 1;PRINT 2;", tokens(
			token(common.PRINT, "PRINT"),
			token(common.INTEGER, "1"),
			token(common.SEMICOLON, ";"),
			token(common.PRINT, "PRINT"),
			token(common.INTEGER, "2"),
			token(common.SEMICOLON, ";"),
		)},
		{"after an expression statement", "1;PRINT 2;", tokens(
			token(common.INTEGER, "1"),
			token(common.SEMICOLON, ";"),
			token(common.PRINT, "PRINT"),
			token(common.INTEGER, "2"),
			token(common.SEMICOLON, ";"),
		)},
		{"keyword inside a string", `"PRINT";`, tokens(
			token(common.STRING, `"PRINT"`),
			token(common.SEMICOLON, ";"),
		)},
		{"keyword inside a comment", "@ PRINT 1;", tokens()},
	})
}

func TestPrintKeywordPositions(t *testing.T) {
	runScanPositionTestCases(t, []scannerPositionTestCase{
		{"print statement", "PRINT 1;", []common.Token{
			tokenAt(common.PRINT, "PRINT", 1, 1),
			tokenAt(common.INTEGER, "1", 1, 7),
			tokenAt(common.SEMICOLON, ";", 1, 8),
			tokenAt(common.EOF, "", 1, 9),
		}},
		{"token right after the keyword", "PRINT(1);", []common.Token{
			tokenAt(common.PRINT, "PRINT", 1, 1),
			tokenAt(common.OPEN_PAR, "(", 1, 6),
			tokenAt(common.INTEGER, "1", 1, 7),
			tokenAt(common.CLOSED_PAR, ")", 1, 8),
			tokenAt(common.SEMICOLON, ";", 1, 9),
			tokenAt(common.EOF, "", 1, 10),
		}},
		{"only keyword", "PRINT", []common.Token{
			tokenAt(common.PRINT, "PRINT", 1, 1),
			tokenAt(common.EOF, "", 1, 6),
		}},
		{"line feed right after the keyword", "PRINT\n1;\n2;", []common.Token{
			tokenAt(common.PRINT, "PRINT", 1, 1),
			tokenAt(common.INTEGER, "1", 2, 1),
			tokenAt(common.SEMICOLON, ";", 2, 2),
			tokenAt(common.INTEGER, "2", 3, 1),
			tokenAt(common.SEMICOLON, ";", 3, 2),
			tokenAt(common.EOF, "", 3, 3),
		}},
		{"keyword on the second line", "1;\n  PRINT 2;", []common.Token{
			tokenAt(common.INTEGER, "1", 1, 1),
			tokenAt(common.SEMICOLON, ";", 1, 2),
			tokenAt(common.PRINT, "PRINT", 2, 3),
			tokenAt(common.INTEGER, "2", 2, 9),
			tokenAt(common.SEMICOLON, ";", 2, 10),
			tokenAt(common.EOF, "", 2, 11),
		}},
	})
}

func TestPrintKeywordLookalikes(t *testing.T) {
	runScanIdentifierTestCases(t, []scannerIdentifierTestCase{
		{"followed by a letter", "PRINTX 1;", "PRINTX", 1, 1},
		{"followed by a digit", "PRINT1;", "PRINT1", 1, 1},
		{"followed by an underscore", "PRINT_ 1;", "PRINT_", 1, 1},
		{"followed by a unicode letter", "PRINTá 1;", "PRINTá", 1, 1},
		{"lowercase", "print 1;", "print", 1, 1},
		{"mixed case", "Print 1;", "Print", 1, 1},
		{"incomplete keyword", "PRIN 1;", "PRIN", 1, 1},
		{"incomplete keyword at the end of the source code", "PRIN", "PRIN", 1, 1},
		{"invalid keyword on the second line", "1;\n  PRINTX;", "PRINTX", 2, 3},
	})
}

func TestLogicalKeywords(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"and", "and", tokens(token(common.AND, "and"))},
		{"or", "or", tokens(token(common.OR, "or"))},
		{"not", "not", tokens(token(common.NOT, "not"))},
		{"and between booleans", "True and False;", tokens(
			token(common.TRUE, "True"),
			token(common.AND, "and"),
			token(common.FALSE, "False"),
			token(common.SEMICOLON, ";"),
		)},
		{"or between comparisons", "1 < 2 or False;", tokens(
			token(common.INTEGER, "1"),
			token(common.LESS, "<"),
			token(common.INTEGER, "2"),
			token(common.OR, "or"),
			token(common.FALSE, "False"),
			token(common.SEMICOLON, ";"),
		)},
		{"not followed by open parentheses", "not(True);", tokens(
			token(common.NOT, "not"),
			token(common.OPEN_PAR, "("),
			token(common.TRUE, "True"),
			token(common.CLOSED_PAR, ")"),
			token(common.SEMICOLON, ";"),
		)},
		{"double not", "not not True;", tokens(
			token(common.NOT, "not"),
			token(common.NOT, "not"),
			token(common.TRUE, "True"),
			token(common.SEMICOLON, ";"),
		)},
		{"keyword inside a string", `"and";`, tokens(
			token(common.STRING, `"and"`),
			token(common.SEMICOLON, ";"),
		)},
		{"keyword inside a comment", "@ True and False;", tokens()},
	})
}

func TestLogicalKeywordPositions(t *testing.T) {
	runScanPositionTestCases(t, []scannerPositionTestCase{
		{"and between booleans", "True and False;", []common.Token{
			tokenAt(common.TRUE, "True", 1, 1),
			tokenAt(common.AND, "and", 1, 6),
			tokenAt(common.FALSE, "False", 1, 10),
			tokenAt(common.SEMICOLON, ";", 1, 15),
			tokenAt(common.EOF, "", 1, 16),
		}},
		{"not and or in the same statement", "not True or False;", []common.Token{
			tokenAt(common.NOT, "not", 1, 1),
			tokenAt(common.TRUE, "True", 1, 5),
			tokenAt(common.OR, "or", 1, 10),
			tokenAt(common.FALSE, "False", 1, 13),
			tokenAt(common.SEMICOLON, ";", 1, 18),
			tokenAt(common.EOF, "", 1, 19),
		}},
		{"keyword on the second line", "True;\n  not False;", []common.Token{
			tokenAt(common.TRUE, "True", 1, 1),
			tokenAt(common.SEMICOLON, ";", 1, 5),
			tokenAt(common.NOT, "not", 2, 3),
			tokenAt(common.FALSE, "False", 2, 7),
			tokenAt(common.SEMICOLON, ";", 2, 12),
			tokenAt(common.EOF, "", 2, 13),
		}},
	})
}

func TestLogicalKeywordLookalikes(t *testing.T) {
	runScanIdentifierTestCases(t, []scannerIdentifierTestCase{
		{"uppercase and", "True AND False;", "AND", 1, 6},
		{"uppercase or", "True OR False;", "OR", 1, 6},
		{"uppercase not", "NOT True;", "NOT", 1, 1},
		{"capitalized not", "Not True;", "Not", 1, 1},
		{"and followed by a letter", "True andy False;", "andy", 1, 6},
		{"or followed by a digit", "True or1 False;", "or1", 1, 6},
		{"incomplete keyword", "True an False;", "an", 1, 6},
		{"keyword glued to the next word", "not(True)andFalse;", "andFalse", 1, 10},
	})
}

func TestBlockBraces(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"empty block", "{}", tokens(
			token(common.OPEN_BRACE, "{"),
			token(common.CLOSED_BRACE, "}"),
		)},
		{"block with a statement", "{PRINT 1;}", tokens(
			token(common.OPEN_BRACE, "{"),
			token(common.PRINT, "PRINT"),
			token(common.INTEGER, "1"),
			token(common.SEMICOLON, ";"),
			token(common.CLOSED_BRACE, "}"),
		)},
		{"nested blocks", "{{}}", tokens(
			token(common.OPEN_BRACE, "{"),
			token(common.OPEN_BRACE, "{"),
			token(common.CLOSED_BRACE, "}"),
			token(common.CLOSED_BRACE, "}"),
		)},
		{"braces inside a string", `"{}";`, tokens(
			token(common.STRING, `"{}"`),
			token(common.SEMICOLON, ";"),
		)},
		{"braces inside a comment", "@ { PRINT 1; }", tokens()},
	})
}

func TestIfKeywords(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"if", "if", tokens(token(common.IF, "if"))},
		{"else", "else", tokens(token(common.ELSE, "else"))},
		{"if statement", "if (True) { PRINT 1; }", tokens(
			token(common.IF, "if"),
			token(common.OPEN_PAR, "("),
			token(common.TRUE, "True"),
			token(common.CLOSED_PAR, ")"),
			token(common.OPEN_BRACE, "{"),
			token(common.PRINT, "PRINT"),
			token(common.INTEGER, "1"),
			token(common.SEMICOLON, ";"),
			token(common.CLOSED_BRACE, "}"),
		)},
		{"if else statement", "if (False) {} else {}", tokens(
			token(common.IF, "if"),
			token(common.OPEN_PAR, "("),
			token(common.FALSE, "False"),
			token(common.CLOSED_PAR, ")"),
			token(common.OPEN_BRACE, "{"),
			token(common.CLOSED_BRACE, "}"),
			token(common.ELSE, "else"),
			token(common.OPEN_BRACE, "{"),
			token(common.CLOSED_BRACE, "}"),
		)},
		{"else if", "else if", tokens(
			token(common.ELSE, "else"),
			token(common.IF, "if"),
		)},
		{"keywords without spaces around them", "if(True){}else{}", tokens(
			token(common.IF, "if"),
			token(common.OPEN_PAR, "("),
			token(common.TRUE, "True"),
			token(common.CLOSED_PAR, ")"),
			token(common.OPEN_BRACE, "{"),
			token(common.CLOSED_BRACE, "}"),
			token(common.ELSE, "else"),
			token(common.OPEN_BRACE, "{"),
			token(common.CLOSED_BRACE, "}"),
		)},
		{"keyword inside a string", `"if";`, tokens(
			token(common.STRING, `"if"`),
			token(common.SEMICOLON, ";"),
		)},
		{"keyword inside a comment", "@ if (True) {} else {}", tokens()},
	})
}

func TestIfKeywordPositions(t *testing.T) {
	runScanPositionTestCases(t, []scannerPositionTestCase{
		{"if else statement", "if (True) {} else {}", []common.Token{
			tokenAt(common.IF, "if", 1, 1),
			tokenAt(common.OPEN_PAR, "(", 1, 4),
			tokenAt(common.TRUE, "True", 1, 5),
			tokenAt(common.CLOSED_PAR, ")", 1, 9),
			tokenAt(common.OPEN_BRACE, "{", 1, 11),
			tokenAt(common.CLOSED_BRACE, "}", 1, 12),
			tokenAt(common.ELSE, "else", 1, 14),
			tokenAt(common.OPEN_BRACE, "{", 1, 19),
			tokenAt(common.CLOSED_BRACE, "}", 1, 20),
			tokenAt(common.EOF, "", 1, 21),
		}},
		{"if statement in several lines", "if (True) {\n  PRINT 1;\n}", []common.Token{
			tokenAt(common.IF, "if", 1, 1),
			tokenAt(common.OPEN_PAR, "(", 1, 4),
			tokenAt(common.TRUE, "True", 1, 5),
			tokenAt(common.CLOSED_PAR, ")", 1, 9),
			tokenAt(common.OPEN_BRACE, "{", 1, 11),
			tokenAt(common.PRINT, "PRINT", 2, 3),
			tokenAt(common.INTEGER, "1", 2, 9),
			tokenAt(common.SEMICOLON, ";", 2, 10),
			tokenAt(common.CLOSED_BRACE, "}", 3, 1),
			tokenAt(common.EOF, "", 3, 2),
		}},
	})
}

func TestIfKeywordLookalikes(t *testing.T) {
	runScanIdentifierTestCases(t, []scannerIdentifierTestCase{
		{"uppercase if", "IF (True) {}", "IF", 1, 1},
		{"capitalized if", "If (True) {}", "If", 1, 1},
		{"uppercase else", "if (True) {} ELSE {}", "ELSE", 1, 14},
		{"if followed by a letter", "iff (True) {}", "iff", 1, 1},
		{"else glued to if", "if (True) {} elseif (False) {}", "elseif", 1, 14},
		{"incomplete else", "if (True) {} els {}", "els", 1, 14},
	})
}

func TestWhileKeywords(t *testing.T) {
	runScanTestCases(t, []scannerTestCase{
		{"while", "while", tokens(token(common.WHILE, "while"))},
		{"break", "break", tokens(token(common.BREAK, "break"))},
		{"continue", "continue", tokens(token(common.CONTINUE, "continue"))},
		{"while statement", "while (True) { PRINT 1; }", tokens(
			token(common.WHILE, "while"),
			token(common.OPEN_PAR, "("),
			token(common.TRUE, "True"),
			token(common.CLOSED_PAR, ")"),
			token(common.OPEN_BRACE, "{"),
			token(common.PRINT, "PRINT"),
			token(common.INTEGER, "1"),
			token(common.SEMICOLON, ";"),
			token(common.CLOSED_BRACE, "}"),
		)},
		{"break and continue statements", "break; continue;", tokens(
			token(common.BREAK, "break"),
			token(common.SEMICOLON, ";"),
			token(common.CONTINUE, "continue"),
			token(common.SEMICOLON, ";"),
		)},
		{"keywords without spaces around them", "while(True){break;continue;}", tokens(
			token(common.WHILE, "while"),
			token(common.OPEN_PAR, "("),
			token(common.TRUE, "True"),
			token(common.CLOSED_PAR, ")"),
			token(common.OPEN_BRACE, "{"),
			token(common.BREAK, "break"),
			token(common.SEMICOLON, ";"),
			token(common.CONTINUE, "continue"),
			token(common.SEMICOLON, ";"),
			token(common.CLOSED_BRACE, "}"),
		)},
		{"keywords inside a string", `"while break continue";`, tokens(
			token(common.STRING, `"while break continue"`),
			token(common.SEMICOLON, ";"),
		)},
		{"keywords inside a comment", "@ while (True) { break; continue; }", tokens()},
	})
}

func TestWhileKeywordPositions(t *testing.T) {
	runScanPositionTestCases(t, []scannerPositionTestCase{
		{"while statement with break", "while (True) { break; }", []common.Token{
			tokenAt(common.WHILE, "while", 1, 1),
			tokenAt(common.OPEN_PAR, "(", 1, 7),
			tokenAt(common.TRUE, "True", 1, 8),
			tokenAt(common.CLOSED_PAR, ")", 1, 12),
			tokenAt(common.OPEN_BRACE, "{", 1, 14),
			tokenAt(common.BREAK, "break", 1, 16),
			tokenAt(common.SEMICOLON, ";", 1, 21),
			tokenAt(common.CLOSED_BRACE, "}", 1, 23),
			tokenAt(common.EOF, "", 1, 24),
		}},
		{"while statement with continue in several lines", "while (False) {\n  continue;\n}", []common.Token{
			tokenAt(common.WHILE, "while", 1, 1),
			tokenAt(common.OPEN_PAR, "(", 1, 7),
			tokenAt(common.FALSE, "False", 1, 8),
			tokenAt(common.CLOSED_PAR, ")", 1, 13),
			tokenAt(common.OPEN_BRACE, "{", 1, 15),
			tokenAt(common.CONTINUE, "continue", 2, 3),
			tokenAt(common.SEMICOLON, ";", 2, 11),
			tokenAt(common.CLOSED_BRACE, "}", 3, 1),
			tokenAt(common.EOF, "", 3, 2),
		}},
	})
}

func TestWhileKeywordLookalikes(t *testing.T) {
	runScanIdentifierTestCases(t, []scannerIdentifierTestCase{
		{"uppercase while", "WHILE (True) {}", "WHILE", 1, 1},
		{"capitalized while", "While (True) {}", "While", 1, 1},
		{"while followed by a letter", "whilee (True) {}", "whilee", 1, 1},
		{"incomplete while", "whil (True) {}", "whil", 1, 1},
		{"uppercase break", "while (True) { BREAK; }", "BREAK", 1, 16},
		{"break followed by a letter", "while (True) { breaks; }", "breaks", 1, 16},
		{"uppercase continue", "while (True) { CONTINUE; }", "CONTINUE", 1, 16},
		{"incomplete continue", "while (True) { cont; }", "cont", 1, 16},
		{"while glued to break", "whilebreak;", "whilebreak", 1, 1},
	})
}
