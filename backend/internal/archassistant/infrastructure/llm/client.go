package llm

import (
	"context"
	"time"
)

const streamingDeadline = 5 * time.Minute

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role
	Content string
}

type Options struct {
	Model       string
	MaxTokens   int
	Temperature float64
}

type EventType int

const (
	EventToken EventType = iota
	EventDone
	EventError
)

type StreamEvent struct {
	Type       EventType
	Content    string
	TokensUsed int
	Error      error
}

type Client interface {
	StreamChat(ctx context.Context, messages []Message, opts Options) (<-chan StreamEvent, error)
}
