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
	RoleTool      Role = "tool"
)

type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

type Message struct {
	Role       Role
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
	Name       string
}

type ToolDef struct {
	Type     string          `json:"type"`
	Function ToolFunctionDef `json:"function"`
}

type ToolFunctionDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type Options struct {
	Model       string
	MaxTokens   int
	Temperature float64
	Tools       []ToolDef
}

type EventType int

const (
	EventToken EventType = iota
	EventDone
	EventError
	EventToolCall
)

type StreamEvent struct {
	Type       EventType
	Content    string
	TokensUsed int
	Error      error
	ToolCalls  []ToolCall
}

type Client interface {
	StreamChat(ctx context.Context, messages []Message, opts Options) (<-chan StreamEvent, error)
}
