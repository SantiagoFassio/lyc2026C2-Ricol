package common

type TokenType int

const (
	EOF TokenType = iota // End of file

	// Operadores matematicos
	PLUS  // +
	MINUS // -
	STAR  // *
	SLASH // /

	DOT // .

	// Comentarios
	AT // @

	// Abrir y cerrar parentesis y llaves
	OPEN_PAR   // (
	CLOSED_PAR // )

	// Literales
	NUMBER // 42, 3.14
	STRING // "Hello, World!"
)

// > 2 + 5
// NUMBER<2.0>
// PLUS
// NUMBER<5.0>
// EOF

const (
	PLUS_SYMBOL               = '+'
	MINUS_SYMBOL              = '-'
	STAR_SYMBOL               = '*'
	SLASH_SYMBOL              = '/'
	DOT_SYMBOL                = '.'
	AT_SYMBOL                 = '@'
	OPEN_PARENTHESES_SYMBOL   = '('
	CLOSED_PARENTHESES_SYMBOL = ')'
)

// switch char
// case PLUS:
// case MINUS:

var tokenNames = map[TokenType]string{
	PLUS:       "PLUS",
	MINUS:      "MINUS",
	STAR:       "STAR",
	SLASH:      "SLASH",
	DOT:        "DOT",
	OPEN_PAR:   "OPEN_PAR",
	CLOSED_PAR: "CLOSED_PAR",
}

type Token struct {
	TokenType TokenType
	Lexeme    string
}
