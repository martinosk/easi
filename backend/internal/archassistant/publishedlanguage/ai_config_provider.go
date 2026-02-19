package publishedlanguage

import (
	"context"
	"errors"
)

var ErrNotConfigured = errors.New("AI assistant is not configured")

type AIConfigInfo struct {
	Provider    string
	Endpoint    string
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
}

type AIConfigProvider interface {
	GetDecryptedConfig(ctx context.Context) (*AIConfigInfo, error)
	IsConfigured(ctx context.Context) (bool, error)
}
