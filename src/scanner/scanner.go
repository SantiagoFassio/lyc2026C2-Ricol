package scanner

import (
	"fmt"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
)

type Scanner struct {
	sourceCode []rune
	tokens     []common.Token
	currentPos int
}

func NewScanner(sourceCode string) *Scanner {
	return &Scanner{
		sourceCode: []rune(sourceCode),
		tokens:     []common.Token{},
	}
}

func (s *Scanner) Scan() ([]common.Token, error) {
	for !s.isAtTheEnd() {
		err := s.scanNextToken()
		if err != nil {
			return []common.Token{}, err
		}
	}
	s.addToken(common.EOF, "")
	return s.tokens, nil
}

func (s *Scanner) scanNextToken() error {
	currentChar := s.sourceCode[s.currentPos]
	switch currentChar {
	case ' ', '\r', '\t', '\n':
		break
	case '@':
		s.skipRemainingLine()
	case '+':
		s.addToken(common.PLUS, string(currentChar))
	case '-':
		s.addToken(common.MINUS, string(currentChar))
	case '*':
		if s.currentPos+1 < len(s.sourceCode) && s.sourceCode[s.currentPos+1] == '*' {
			s.addToken(common.DOUBLE_STAR, "**")
			s.currentPos++
		} else {
			s.addToken(common.STAR, string(currentChar))
		}
	case '/':
		if s.currentPos+1 < len(s.sourceCode) && s.sourceCode[s.currentPos+1] == '/' {
			s.addToken(common.DOUBLE_SLASH, "//")
			s.currentPos++
		} else {
			s.addToken(common.SLASH, string(currentChar))
		}
	case '%':
		s.addToken(common.PERCENTAGE, string(currentChar))
	case ';':
		s.addToken(common.SEMICOLON, string(currentChar))
	case '(':
		s.addToken(common.OPEN_PAR, string(currentChar))
	case ')':
		s.addToken(common.CLOSED_PAR, string(currentChar))
	default:
		if s.isInteger(currentChar) {
			err := s.scanNumber()
			if err != nil {
				return err
			}
			break
		}
		return fmt.Errorf("Non-recognizable character '%c'", currentChar)
	}
	s.currentPos++
	return nil
}

func (s *Scanner) addToken(tokenType common.TokenType, lexeme string) {
	token := common.Token{
		TokenType: tokenType,
		Lexeme:    lexeme,
	}
	s.tokens = append(s.tokens, token)
}

func (s *Scanner) skipRemainingLine() {
	for !s.isAtAnEndOfLine() {
		s.currentPos++
	}
}

func (s *Scanner) scanNumber() error {
	numberChars := []rune{}
	dotFound := false
	for !s.isAtTheEnd() && (s.isInteger(s.sourceCode[s.currentPos]) || s.sourceCode[s.currentPos] == '.') {
		numberChars = append(numberChars, s.sourceCode[s.currentPos])
		if s.sourceCode[s.currentPos] == '.' {
			if dotFound {
				return fmt.Errorf("Number with multiple decimal separator: '%s'", string(numberChars))
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

func (s *Scanner) isAtAnEndOfLine() bool {
	currentChar := s.sourceCode[s.currentPos]
	return currentChar == '\n' || currentChar == '\r'
}

func (s *Scanner) isAtTheEnd() bool {
	return s.currentPos == len(s.sourceCode)
}

func (s *Scanner) isInteger(char rune) bool {
	return char >= '0' && char <= '9'
}
