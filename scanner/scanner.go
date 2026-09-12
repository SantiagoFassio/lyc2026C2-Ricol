package scanner

import (
	"fmt"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
)

type Scanner struct {
	tokens []common.Token
}

func NewScanner() *Scanner {
	return &Scanner{
		tokens: []common.Token{},
	}
}

func (s *Scanner) AddToken(tokenType common.TokenType, lexeme string) {
	token := common.Token{
		TokenType: tokenType,
		Lexeme:    lexeme,
	}
	s.tokens = append(s.tokens, token)
}

// 2 + 3.5
// Input: un string de go
// Output: Lista de tokens
func (s *Scanner) Scan(code string) ([]common.Token, error) {
	codeRunes := []rune(code)
	for i := 0; i < len(codeRunes); i++ {
		char := codeRunes[i]
		switch char {
		case common.PLUS_SYMBOL:
			s.AddToken(common.PLUS, string(char))
		case common.MINUS_SYMBOL:
			s.AddToken(common.MINUS, string(char))
		case common.STAR_SYMBOL:
			s.AddToken(common.STAR, string(char))
		case common.SLASH_SYMBOL:
			s.AddToken(common.SLASH, string(char))
		case common.OPEN_PARENTHESES_SYMBOL:
			s.AddToken(common.OPEN_PAR, string(char))
		case common.CLOSED_PARENTHESES_SYMBOL:
			s.AddToken(common.CLOSED_PAR, string(char))
		case common.DOT_SYMBOL:
			s.AddToken(common.DOT, string(char))
		default:
			return []common.Token{}, fmt.Errorf("Non-recognizable character '%c'", char)
		}
	}
	return s.tokens, nil
}

func (s *Scanner) isInteger() {
	// TODO
}
