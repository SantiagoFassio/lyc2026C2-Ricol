package common

import (
	"fmt"
)

type TokenType int

const (
	EOF TokenType = iota // End of file

	// Operadores matematicos
	PLUS         // +
	MINUS        // -
	STAR         // *
	DOUBLE_STAR  // **
	SLASH        // /
	DOUBLE_SLASH // //
	PERCENTAGE   // %

	DOT // .

	// Comentarios
	AT // @

	// Fin de expresión
	SEMICOLON

	// Abrir y cerrar parentesis y llaves
	OPEN_PAR   // (
	CLOSED_PAR // )

	// Literales
	INTEGER // 42
	FLOAT   // 3.14
	STRING  // "Hello, World!"
)

// > 2 + 5
// NUMBER<2.0>
// PLUS
// NUMBER<5.0>
// EOF

var tokenNames = map[TokenType]string{
	EOF:          "EOF",
	PLUS:         "PLUS",
	MINUS:        "MINUS",
	STAR:         "STAR",
	DOUBLE_STAR:  "DOUBLE_STAR",
	SLASH:        "SLASH",
	DOUBLE_SLASH: "DOUBLE_SLASH",
	PERCENTAGE:   "PERCENTAGE",
	DOT:          "DOT",
	SEMICOLON:    "SEMICOLON",
	OPEN_PAR:     "OPEN_PAR",
	CLOSED_PAR:   "CLOSED_PAR",
	INTEGER:      "INTEGER",
	FLOAT:        "FLOAT",
	STRING:       "STRING",
}

// Secuencias de escape válidas en un STRING y el carácter que representa cada una.
var EscapeSequences = map[rune]rune{
	'"':  '"',
	'\\': '\\',
	'n':  '\n',
	't':  '\t',
}

type Position struct {
	Line   int
	Column int
}

type Token struct {
	TokenType TokenType
	Lexeme    string
	Position  Position
}

func NewToken(tokenType TokenType, lexeme string, position Position) Token {
	return Token{
		TokenType: tokenType,
		Lexeme:    lexeme,
		Position:  position,
	}
}

func (t Token) String() string {
	return fmt.Sprintf("%s<%s>", tokenNames[t.TokenType], t.Lexeme)
}
