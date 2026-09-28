package common

import (
	"fmt"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
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

	// Operador de asignación
	EQUAL // =

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

	// Especificación de tipo al declarar variables
	COLON

	// Abrir y cerrar parentesis y llaves
	OPEN_PAR   // (
	CLOSED_PAR // )

	// Llaves
	OPEN_BRACE   // {
	CLOSED_BRACE // }

	// Literales
	INTEGER // 42
	FLOAT   // 3.14
	STRING  // "Hello, World!"

	// Identificador
	IDENTIFIER

	// Palabras clave
	TRUE     // True
	FALSE    // False
	PRINT    // PRINT
	AND      // and
	OR       // or
	NOT      // not
	IF       // if
	ELSE     // else
	WHILE    // while
	CONTINUE // continue
	BREAK    // break
	LET      // let

	// Tipos de datos
	INT_TYPE
	FLOAT_TYPE
	STRING_TYPE
	BOOL_TYPE
)

var TypeTokenTypes = map[TokenType]types.Type{
	INT_TYPE:    types.Int,
	FLOAT_TYPE:  types.Float,
	STRING_TYPE: types.Str,
	BOOL_TYPE:   types.Bool,
}

var tokenNames = map[TokenType]string{
	EOF:           "EOF",
	PLUS:          "PLUS",
	MINUS:         "MINUS",
	STAR:          "STAR",
	DOUBLE_STAR:   "DOUBLE_STAR",
	SLASH:         "SLASH",
	DOUBLE_SLASH:  "DOUBLE_SLASH",
	PERCENTAGE:    "PERCENTAGE",
	EQUAL:         "EQUAL",
	DOUBLE_EQUAL:  "DOUBLE_EQUAL",
	NOT_EQUAL:     "NOT_EQUAL",
	LESS:          "LESS",
	LESS_EQUAL:    "LESS_EQUAL",
	GREATER:       "GREATER",
	GREATER_EQUAL: "GREATER_EQUAL",
	DOT:           "DOT",
	SEMICOLON:     "SEMICOLON",
	COLON:         "COLON",
	OPEN_PAR:      "OPEN_PAR",
	CLOSED_PAR:    "CLOSED_PAR",
	OPEN_BRACE:    "OPEN_BRACE",
	CLOSED_BRACE:  "CLOSED_BRACE",
	INTEGER:       "INTEGER",
	FLOAT:         "FLOAT",
	STRING:        "STRING",
	IDENTIFIER:    "IDENTIFIER",
	TRUE:          "TRUE",
	FALSE:         "FALSE",
	PRINT:         "PRINT",
	AND:           "AND",
	OR:            "OR",
	NOT:           "NOT",
	IF:            "IF",
	ELSE:          "ELSE",
	WHILE:         "WHILE",
	CONTINUE:      "CONTINUE",
	BREAK:         "BREAK",
	LET:           "LET",
	INT_TYPE:      "INT_TYPE",
	FLOAT_TYPE:    "FLOAT_TYPE",
	STRING_TYPE:   "STRING_TYPE",
	BOOL_TYPE:     "BOOL_TYPE",
}

// Palabras clave del lenguaje y el tipo de token que genera cada una.
var Keywords = map[string]TokenType{
	"True":     TRUE,
	"False":    FALSE,
	"PRINT":    PRINT,
	"and":      AND,
	"or":       OR,
	"not":      NOT,
	"if":       IF,
	"else":     ELSE,
	"while":    WHILE,
	"continue": CONTINUE,
	"break":    BREAK,
	"let":      LET,
	"Int":      INT_TYPE,
	"Float":    FLOAT_TYPE,
	"String":   STRING_TYPE,
	"Bool":     BOOL_TYPE,
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
