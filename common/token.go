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
	PLUS:       "PLUS",
	MINUS:      "MINUS",
	STAR:       "STAR",
	SLASH:      "SLASH",
	DOT:        "DOT",
	SEMICOLON:  "SEMICOLON",
	OPEN_PAR:   "OPEN_PAR",
	CLOSED_PAR: "CLOSED_PAR",
}

type Token struct {
	TokenType TokenType
	Lexeme    string
}
