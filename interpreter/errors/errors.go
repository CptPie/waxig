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
