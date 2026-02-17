package valueobjects

import (
	"errors"
	"strings"
)

var (
	ErrModelNameEmpty   = errors.New("model name cannot be empty")
	ErrModelNameTooLong = errors.New("model name cannot exceed 100 characters")
)

type ModelName struct {
	value string
}

func NewModelName(value string) (ModelName, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ModelName{}, ErrModelNameEmpty
	}
	if len(trimmed) > 100 {
		return ModelName{}, ErrModelNameTooLong
	}
	return ModelName{value: trimmed}, nil
}

func ReconstructModelName(value string) ModelName {
	return ModelName{value: value}
}

func (m ModelName) Value() string { return m.value }
