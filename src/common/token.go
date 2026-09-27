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

	// Operadores de comparacion
	DOUBLE_EQUAL  // ==
	NOT_EQUAL     // !=
	LESS          // <
	LESS_EQUAL    // <=
	GREATER       // >
	GREATER_EQUAL // >=

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

	// Palabras clave
	TRUE  // True
	FALSE // False
	PRINT // PRINT
	AND   // and
	OR    // or
	NOT   // not
)

// > 2 + 5
// NUMBER<2.0>
// PLUS
// NUMBER<5.0>
// EOF

var tokenNames = map[TokenType]string{
	EOF:           "EOF",
	PLUS:          "PLUS",
	MINUS:         "MINUS",
	STAR:          "STAR",
	DOUBLE_STAR:   "DOUBLE_STAR",
	SLASH:         "SLASH",
	DOUBLE_SLASH:  "DOUBLE_SLASH",
	PERCENTAGE:    "PERCENTAGE",
	DOUBLE_EQUAL:  "DOUBLE_EQUAL",
	NOT_EQUAL:     "NOT_EQUAL",
	LESS:          "LESS",
	LESS_EQUAL:    "LESS_EQUAL",
	GREATER:       "GREATER",
	GREATER_EQUAL: "GREATER_EQUAL",
	DOT:           "DOT",
	SEMICOLON:     "SEMICOLON",
	OPEN_PAR:      "OPEN_PAR",
	CLOSED_PAR:    "CLOSED_PAR",
	INTEGER:       "INTEGER",
	FLOAT:         "FLOAT",
	STRING:        "STRING",
	TRUE:          "TRUE",
	FALSE:         "FALSE",
	PRINT:         "PRINT",
	AND:           "AND",
	OR:            "OR",
	NOT:           "NOT",
}

// Palabras clave del lenguaje y el tipo de token que genera cada una.
var Keywords = map[string]TokenType{
	"True":  TRUE,
	"False": FALSE,
	"PRINT": PRINT,
	"and":   AND,
	"or":    OR,
	"not":   NOT,
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
