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
}

type Token struct {
	TokenType TokenType
	Lexeme    string
}

func (t Token) String() string {
	return fmt.Sprintf("%s<%s>", tokenNames[t.TokenType], t.Lexeme)
}
