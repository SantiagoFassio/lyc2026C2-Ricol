package common

import (
	"fmt"
	"strings"
)

type RicolError struct {
	Position Position
	Message  string
}

type RicolErrorList []RicolError

func NewRicolError(position Position, message string) RicolError {
	return RicolError{Position: position, Message: message}
}

func (r RicolError) Error() string {
	return fmt.Sprintf("[line %d, column %d] %s", r.Position.Line, r.Position.Column, r.Message)
}

func (l RicolErrorList) Error() string {
	messages := make([]string, len(l))
	for i, ricolError := range l {
		messages[i] = ricolError.Error()
	}
	return strings.Join(messages, "\n")
}
