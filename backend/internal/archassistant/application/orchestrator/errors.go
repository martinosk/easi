package orchestrator

import "errors"

var ErrConversationNotFound = errors.New("conversation not found")

type ValidationError struct {
	Err error
}

func (e *ValidationError) Error() string { return e.Err.Error() }
func (e *ValidationError) Unwrap() error { return e.Err }

type LLMError struct {
	Message string
}

func (e *LLMError) Error() string { return e.Message }

type TimeoutError struct {
	Err error
}

func (e *TimeoutError) Error() string { return e.Err.Error() }
func (e *TimeoutError) Unwrap() error { return e.Err }

type IterationLimitError struct{}

func (e *IterationLimitError) Error() string {
	return "max tool iterations exceeded"
}
