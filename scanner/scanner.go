package scanner

import (
	"fmt"
	"strconv"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
)


type Scanner struct {
	tokens []Token
}



func NewScanner() *Scanner {
	return &Scanner{
		tokens: []Token{},
	}
}


func (s *Scanner) AddToken(tokenType common.TokenType, lexeme string) {
	token := Token{
		Type:    tokenType,
		lexeme:  lexeme,
	}
	s.tokens = append(s.tokens, token)
}


// 2 + 3.5

// Input: un string de go
// Output: Lista de tokens
func (s *Scanner) Scan(code string) ([]Token, error) {
	code_runes := []rune(s)
	for i := 0; i < len(code_runes); i++ {
		char := code[i]
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
			// case isANumber:
			//   append(s.pendingChars, string(char))

		  default:
				return []Token{}, fmt.Errorf("Non-recognizable character '%c'", char)
		}
	}
	return s.tokens, nil
}

func (s *Scanner) isInteger ()
