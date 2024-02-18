package token

import "strings"

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT  = "IDENT"  // add, foobar, x, y, ...
	INT    = "INT"    // 1343456
	STRING = "STRING" // "foobar"

	// Operators
	ASSIGN   = "=" // Assignment
	PLUS     = "+" // Addition
	MINUS    = "-" // Subtraction
	ASTERISK = "*" // Multiplication
	SLASH    = "/" // Division
	HAT      = "^" // Exponentiation

	// Boolean operators
	BANG   = "!"  // Negation
	EQ     = "==" // Equal to
	NOT_EQ = "!=" // Not equal to
	LT     = "<"  // Less than
	GT     = ">"  // Greater than
	LTEQ   = "<=" // Less than or equal to
	GTEQ   = ">=" // Greater than or equal to

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	LBRACKET = "["
	RBRACKET = "]"

	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
)

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
}

func LookupIdent(ident string) TokenType {
	// Check if the identifier is a keyword, case-insensitive
	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}
