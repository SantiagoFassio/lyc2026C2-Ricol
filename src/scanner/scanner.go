package scanner

import (
	"fmt"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
)

type Scanner struct {
	sourceCode          []rune
	tokens              []common.Token
	currentPos          int
	tokenStartPos       int
	currentLineNumber   int
	currentLineStartPos int
}

func NewScanner(sourceCode string) *Scanner {
	return &Scanner{
		sourceCode:        []rune(sourceCode),
		tokens:            []common.Token{},
		currentLineNumber: 1,
	}
}

func (s *Scanner) Scan() ([]common.Token, error) {
	for !s.isAtTheEnd() {
		err := s.scanNextToken()
		if err != nil {
			return []common.Token{}, err
		}
	}
	s.startNextToken()
	s.addToken(common.EOF, "")
	return s.tokens, nil
}

func (s *Scanner) scanNextToken() error {
	s.startNextToken()
	currentChar := s.sourceCode[s.currentPos]
	switch currentChar {
	case ' ', '\r', '\t':
		break
	case '\n':
		s.currentLineNumber++
		s.currentLineStartPos = s.currentPos + 1
	case '@':
		s.skipRemainingLine()
	case '+':
		s.addToken(common.PLUS, string(currentChar))
	case '-':
		s.addToken(common.MINUS, string(currentChar))
	case '*':
		if s.nextCharIs('*') {
			s.addToken(common.DOUBLE_STAR, "**")
			s.currentPos++
		} else {
			s.addToken(common.STAR, string(currentChar))
		}
	case '/':
		if s.nextCharIs('/') {
			s.addToken(common.DOUBLE_SLASH, "//")
			s.currentPos++
		} else {
			s.addToken(common.SLASH, string(currentChar))
		}
	case '%':
		s.addToken(common.PERCENTAGE, string(currentChar))
	case '=':
		if !s.nextCharIs('=') {
			return common.NewRicolError(s.currentTokenPosition(), fmt.Sprintf("Non-recognizable character '%c'", currentChar))
		}
		s.addToken(common.DOUBLE_EQUAL, "==")
		s.currentPos++
	case '!':
		if !s.nextCharIs('=') {
			return common.NewRicolError(s.currentTokenPosition(), fmt.Sprintf("Non-recognizable character '%c'", currentChar))
		}
		s.addToken(common.NOT_EQUAL, "!=")
		s.currentPos++
	case '<':
		if s.nextCharIs('=') {
			s.addToken(common.LESS_EQUAL, "<=")
			s.currentPos++
		} else {
			s.addToken(common.LESS, string(currentChar))
		}
	case '>':
		if s.nextCharIs('=') {
			s.addToken(common.GREATER_EQUAL, ">=")
			s.currentPos++
		} else {
			s.addToken(common.GREATER, string(currentChar))
		}
	case ';':
		s.addToken(common.SEMICOLON, string(currentChar))
	case '(':
		s.addToken(common.OPEN_PAR, string(currentChar))
	case ')':
		s.addToken(common.CLOSED_PAR, string(currentChar))
	case '"':
		err := s.scanString()
		if err != nil {
			return err
		}
	default:
		if s.isInteger(currentChar) {
			err := s.scanNumber()
			if err != nil {
				return err
			}
			break
		}
		if s.isLetter(currentChar) {
			err := s.scanKeyword()
			if err != nil {
				return err
			}
			break
		}
		return common.NewRicolError(s.currentTokenPosition(), fmt.Sprintf("Non-recognizable character '%c'", currentChar))
	}
	s.currentPos++
	return nil
}

func (s *Scanner) startNextToken() {
	s.tokenStartPos = s.currentPos
}

func (s *Scanner) addToken(tokenType common.TokenType, lexeme string) {
	s.tokens = append(s.tokens, common.NewToken(tokenType, lexeme, s.currentTokenPosition()))
}

func (s *Scanner) currentTokenPosition() common.Position {
	return s.positionAt(s.tokenStartPos)
}

func (s *Scanner) positionAt(sourcePos int) common.Position {
	return common.Position{
		Line:   s.currentLineNumber,
		Column: sourcePos - s.currentLineStartPos + 1,
	}
}

func (s *Scanner) skipRemainingLine() {
	for !s.isAtAnEndOfLine() {
		s.currentPos++
	}
	s.currentPos--
}

func (s *Scanner) scanNumber() error {
	numberChars := []rune{}
	dotFound := false
	for !s.isAtTheEnd() && (s.isInteger(s.sourceCode[s.currentPos]) || s.sourceCode[s.currentPos] == '.') {
		numberChars = append(numberChars, s.sourceCode[s.currentPos])
		if s.sourceCode[s.currentPos] == '.' {
			if dotFound {
				return common.NewRicolError(s.currentTokenPosition(),
					fmt.Sprintf("Number with multiple decimal separator: '%s'", string(numberChars)))
			}
			dotFound = true
		}
		s.currentPos++
	}
	s.currentPos--

	var tokenType common.TokenType
	if dotFound {
		tokenType = common.FLOAT
	} else {
		tokenType = common.INTEGER
	}
	s.addToken(tokenType, string(numberChars))
	return nil
}

func (s *Scanner) scanKeyword() error {
	keywordChars := []rune{}
	for !s.isAtTheEnd() && (s.isLetter(s.sourceCode[s.currentPos]) || s.isInteger(s.sourceCode[s.currentPos])) {
		keywordChars = append(keywordChars, s.sourceCode[s.currentPos])
		s.currentPos++
	}
	s.currentPos--

	keyword := string(keywordChars)
	tokenType, ok := common.Keywords[keyword]
	if !ok {
		return common.NewRicolError(s.currentTokenPosition(), fmt.Sprintf("Unknown keyword '%s'", keyword))
	}
	s.addToken(tokenType, keyword)
	return nil
}

func (s *Scanner) scanString() error {
	s.currentPos++
	for !s.isAtAnEndOfLine() && s.sourceCode[s.currentPos] != '"' {
		if s.sourceCode[s.currentPos] == '\\' {
			err := s.skipEscapeSequence()
			if err != nil {
				return err
			}
		}
		s.currentPos++
	}
	if s.isAtAnEndOfLine() {
		return s.unterminatedStringError()
	}
	s.addToken(common.STRING, string(s.sourceCode[s.tokenStartPos:s.currentPos+1]))
	return nil
}

func (s *Scanner) skipEscapeSequence() error {
	backslashPos := s.currentPos
	s.currentPos++
	if s.isAtAnEndOfLine() {
		return s.unterminatedStringError()
	}
	escapedChar := s.sourceCode[s.currentPos]
	if _, ok := common.EscapeSequences[escapedChar]; !ok {
		backslashPosition := s.positionAt(backslashPos)
		return common.NewRicolError(backslashPosition, fmt.Sprintf("Invalid escape sequence '\\%c'", escapedChar))
	}
	return nil
}

func (s *Scanner) unterminatedStringError() error {
	return common.NewRicolError(s.currentTokenPosition(), "Unterminated string")
}

func (s *Scanner) isAtAnEndOfLine() bool {
	if s.isAtTheEnd() {
		return true
	}
	currentChar := s.sourceCode[s.currentPos]
	return currentChar == '\n' || currentChar == '\r'
}

func (s *Scanner) isAtTheEnd() bool {
	return s.currentPos == len(s.sourceCode)
}

func (s *Scanner) isInteger(char rune) bool {
	return char >= '0' && char <= '9'
}

func (s *Scanner) nextCharIs(char rune) bool {
	return s.currentPos+1 < len(s.sourceCode) && s.sourceCode[s.currentPos+1] == char
}

func (s *Scanner) isLetter(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}
