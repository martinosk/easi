package valueobjects

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrEndpointEmpty   = errors.New("LLM endpoint cannot be empty")
	ErrEndpointTooLong = errors.New("LLM endpoint cannot exceed 500 characters")
	ErrEndpointInvalid = errors.New("LLM endpoint must be a valid URL with https:// or http://localhost")
)

type LLMEndpoint struct {
	value string
}

func NewLLMEndpoint(value string) (LLMEndpoint, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return LLMEndpoint{}, ErrEndpointEmpty
	}
	if len(trimmed) > 500 {
		return LLMEndpoint{}, ErrEndpointTooLong
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" {
		return LLMEndpoint{}, ErrEndpointInvalid
	}
	if parsed.Scheme == "https" {
		return LLMEndpoint{value: trimmed}, nil
	}
	if parsed.Scheme == "http" && strings.EqualFold(parsed.Hostname(), "localhost") {
		return LLMEndpoint{value: trimmed}, nil
	}
	return LLMEndpoint{}, ErrEndpointInvalid
}

func ReconstructLLMEndpoint(value string) LLMEndpoint {
	return LLMEndpoint{value: value}
}

func (e LLMEndpoint) Value() string { return e.value }
