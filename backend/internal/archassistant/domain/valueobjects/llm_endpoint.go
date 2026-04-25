package valueobjects

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrEndpointEmpty          = errors.New("LLM endpoint cannot be empty")
	ErrEndpointTooLong        = errors.New("LLM endpoint cannot exceed 500 characters")
	ErrEndpointInvalid        = errors.New("LLM endpoint must be a valid https:// URL")
	ErrEndpointHostNotAllowed = errors.New("LLM endpoint host is not in the allowed providers list")
)

var allowedExactHosts = map[string]struct{}{
	"api.openai.com":    {},
	"api.anthropic.com": {},
}

var allowedHostSuffixes = []string{
	".openai.azure.com",
	".cognitiveservices.azure.com",
}

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
	if err != nil || parsed.Host == "" || parsed.Scheme != "https" {
		return LLMEndpoint{}, ErrEndpointInvalid
	}
	if !isAllowedHost(parsed.Hostname()) {
		return LLMEndpoint{}, ErrEndpointHostNotAllowed
	}
	return LLMEndpoint{value: trimmed}, nil
}

func isAllowedHost(hostname string) bool {
	host := strings.ToLower(hostname)
	if _, ok := allowedExactHosts[host]; ok {
		return true
	}
	for _, suffix := range allowedHostSuffixes {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}
	return false
}

func ReconstructLLMEndpoint(value string) LLMEndpoint {
	return LLMEndpoint{value: value}
}

func (e LLMEndpoint) Value() string { return e.value }
