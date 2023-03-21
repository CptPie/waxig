package errors

import (
	"fmt"
	"waixg/interpreter/token"
)

type PeekTypeMismatch struct {
	Expected token.TokenType
	Actual   token.TokenType
}

func (e *PeekTypeMismatch) Error() string {
	return fmt.Sprintf("[PeekTypeMismatch] Expected next token to be %s, got %s instead", e.Expected, e.Actual)
}

type NoPrefixParseFnError struct {
	TokenType token.TokenType
}

func (e *NoPrefixParseFnError) Error() string {
	return fmt.Sprintf("[NoPrefixParseFnError] No prefix parse function for %s found", e.TokenType)
}

type InvalidIntegerLiteral struct {
	Literal string
}

func (e *InvalidIntegerLiteral) Error() string {
	return fmt.Sprintf("[InvalidIntegerLiteral] Could not parse %q as integer", e.Literal)
}
