package orchestrator

import "context"

type ToolCallStartEvent struct {
	ToolCallID string
	Name       string
	Arguments  string
}

type ToolCallResultEvent struct {
	ToolCallID    string
	Name          string
	ResultPreview string
}

type ThinkingEvent struct {
	Message string
}

type StreamWriter interface {
	WriteToken(content string) error
	WriteDone(messageID string, tokensUsed int) error
	WriteToolCallStart(event ToolCallStartEvent) error
	WriteToolCallResult(event ToolCallResultEvent) error
	WriteThinking(event ThinkingEvent) error
}

type LLMClientFactory interface {
	Create(provider, endpoint, apiKey string) (LLMClient, error)
}

type LLMClient interface {
	StreamChat(ctx context.Context, messages []ChatMessage, opts ChatOptions) (<-chan ChatEvent, error)
}

type ChatRole string

const (
	ChatRoleSystem    ChatRole = "system"
	ChatRoleUser      ChatRole = "user"
	ChatRoleAssistant ChatRole = "assistant"
	ChatRoleTool      ChatRole = "tool"
)

type ChatToolCall struct {
	ID        string
	Name      string
	Arguments string
}

type ChatMessage struct {
	Role       ChatRole
	Content    string
	ToolCalls  []ChatToolCall
	ToolCallID string
	Name       string
}

type ChatOptions struct {
	Model       string
	MaxTokens   int
	Temperature float64
	Tools       []interface{}
}

type ChatEventType int

const (
	ChatEventToken ChatEventType = iota
	ChatEventDone
	ChatEventError
	ChatEventToolCall
)

type ChatEvent struct {
	Type       ChatEventType
	Content    string
	TokensUsed int
	Error      error
	ToolCalls  []ChatToolCall
}
