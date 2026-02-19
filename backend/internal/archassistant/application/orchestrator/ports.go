package orchestrator

import "context"

type StreamWriter interface {
	WriteToken(content string) error
	WriteDone(messageID string, tokensUsed int) error
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
)

type ChatMessage struct {
	Role    ChatRole
	Content string
}

type ChatOptions struct {
	Model       string
	MaxTokens   int
	Temperature float64
}

type ChatEventType int

const (
	ChatEventToken ChatEventType = iota
	ChatEventDone
	ChatEventError
)

type ChatEvent struct {
	Type       ChatEventType
	Content    string
	TokensUsed int
	Error      error
}
