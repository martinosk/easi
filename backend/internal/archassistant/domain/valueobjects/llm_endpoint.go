package valueobjects

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

var (
	ErrEndpointEmpty       = errors.New("LLM endpoint cannot be empty")
	ErrEndpointTooLong     = errors.New("LLM endpoint cannot exceed 500 characters")
	ErrEndpointInvalid     = errors.New("LLM endpoint must be a valid URL with https:// or http://localhost")
	ErrEndpointPrivateAddr = errors.New("LLM endpoint must not point to a private or link-local IP address")
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
	if !isAllowedScheme(parsed) {
		return LLMEndpoint{}, ErrEndpointInvalid
	}
	if isPrivateOrLinkLocalIP(parsed.Hostname()) {
		return LLMEndpoint{}, ErrEndpointPrivateAddr
	}
	return LLMEndpoint{value: trimmed}, nil
}

func isAllowedScheme(parsed *url.URL) bool {
	if parsed.Scheme == "https" {
		return true
	}
	return parsed.Scheme == "http" && strings.EqualFold(parsed.Hostname(), "localhost")
}

func isPrivateOrLinkLocalIP(hostname string) bool {
	ip := net.ParseIP(hostname)
	if ip == nil {
		return false
	}
	return ip.IsPrivate() || ip.IsLinkLocalUnicast()
}

func ReconstructLLMEndpoint(value string) LLMEndpoint {
	return LLMEndpoint{value: value}
}

func (e LLMEndpoint) Value() string { return e.value }
