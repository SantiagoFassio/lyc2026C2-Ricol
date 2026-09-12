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
		s.addToken(common.STAR, string(currentChar))
	case '/':
		s.addToken(common.SLASH, string(currentChar))
	case '(':
		s.addToken(common.OPEN_PAR, string(currentChar))
	case ')':
		s.addToken(common.CLOSED_PAR, string(currentChar))
	case '.':
		s.addToken(common.DOT, string(currentChar))
	default:
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

func (s *Scanner) isInteger() {
	// TODO
}

func (s *Scanner) isAtAnEndOfLine() bool {
	currentChar := s.sourceCode[s.currentPos]
	return currentChar == '\n' || currentChar == '\r'
}

func (s *Scanner) isAtTheEnd() bool {
	return s.currentPos == len(s.sourceCode)
}
