package foundation

import (
	"context"
	"fmt"
)

type RuntimeError struct {
	Cause error
	Stack []byte
}

func (err RuntimeError) Error() string {
	s := "runtime error"

	if cause := err.Cause; cause != nil {
		s = fmt.Sprintf("%s: %s", s, cause.Error())
	}

	return s
}

type CleanupError struct {
	Cause error
	Stack []byte
}

func (err CleanupError) Error() string {
	s := "cleanup error"

	if cause := err.Cause; cause != nil {
		s = fmt.Sprintf("%s: %s", s, cause.Error())
	}

	return s
}

type PanicError struct {
	Cause any
}

func (err PanicError) Error() string {
	s := "caught panic"

	if cause := err.Cause; cause != nil {
		s = fmt.Sprintf("%s: %s", s, cause)
	}

	return s
}

// Error is a placeholder for common error handling patterns
func Error(error) {}

// ErrorWithContext is a placeholder for common error handling patterns with a context.
func ErrorWithContext(context.Context, error) {}
