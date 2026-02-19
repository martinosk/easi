package adapters

import (
	"context"
	"net"
	"net/url"

	"easi/backend/internal/archassistant/application/orchestrator"
	"easi/backend/internal/archassistant/infrastructure/llm"
)

type LLMClientFactory struct{}

func NewLLMClientFactory() *LLMClientFactory {
	return &LLMClientFactory{}
}

func (f *LLMClientFactory) Create(provider, endpoint, apiKey string) (orchestrator.LLMClient, error) {
	if err := validateEndpointSafety(endpoint); err != nil {
		return nil, err
	}
	client, err := llm.NewClient(provider, endpoint, apiKey)
	if err != nil {
		return nil, err
	}
	return &llmClientAdapter{client: client}, nil
}

type llmClientAdapter struct {
	client llm.Client
}

func (a *llmClientAdapter) StreamChat(ctx context.Context, messages []orchestrator.ChatMessage, opts orchestrator.ChatOptions) (<-chan orchestrator.ChatEvent, error) {
	llmMessages := convertMessages(messages)

	stream, err := a.client.StreamChat(ctx, llmMessages, convertOptions(opts))
	if err != nil {
		return nil, err
	}

	ch := make(chan orchestrator.ChatEvent, 64)
	go func() {
		defer close(ch)
		for event := range stream {
			ch <- convertEvent(event)
		}
	}()
	return ch, nil
}

func convertMessages(messages []orchestrator.ChatMessage) []llm.Message {
	llmMessages := make([]llm.Message, len(messages))
	for i, m := range messages {
		var toolCalls []llm.ToolCall
		for _, tc := range m.ToolCalls {
			toolCalls = append(toolCalls, llm.ToolCall{ID: tc.ID, Name: tc.Name, Arguments: tc.Arguments})
		}
		llmMessages[i] = llm.Message{
			Role:       llm.Role(m.Role),
			Content:    m.Content,
			ToolCalls:  toolCalls,
			ToolCallID: m.ToolCallID,
			Name:       m.Name,
		}
	}
	return llmMessages
}

func convertOptions(opts orchestrator.ChatOptions) llm.Options {
	var tools []llm.ToolDef
	for _, t := range opts.Tools {
		if td, ok := t.(map[string]interface{}); ok {
			tools = append(tools, parseToolDef(td))
		}
	}
	return llm.Options{
		Model:       opts.Model,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		Tools:       tools,
	}
}

func parseToolDef(td map[string]interface{}) llm.ToolDef {
	toolDef := llm.ToolDef{}
	if tp, ok := td["type"].(string); ok {
		toolDef.Type = tp
	}
	if fn, ok := td["function"].(map[string]interface{}); ok {
		toolDef.Function = parseToolFunctionDef(fn)
	}
	return toolDef
}

func parseToolFunctionDef(fn map[string]interface{}) llm.ToolFunctionDef {
	def := llm.ToolFunctionDef{Parameters: fn["parameters"]}
	if name, ok := fn["name"].(string); ok {
		def.Name = name
	}
	if desc, ok := fn["description"].(string); ok {
		def.Description = desc
	}
	return def
}

func convertEvent(e llm.StreamEvent) orchestrator.ChatEvent {
	switch e.Type {
	case llm.EventToken:
		return orchestrator.ChatEvent{Type: orchestrator.ChatEventToken, Content: e.Content}
	case llm.EventDone:
		return orchestrator.ChatEvent{Type: orchestrator.ChatEventDone, TokensUsed: e.TokensUsed}
	case llm.EventError:
		return orchestrator.ChatEvent{Type: orchestrator.ChatEventError, Error: e.Error}
	case llm.EventToolCall:
		chatToolCalls := make([]orchestrator.ChatToolCall, len(e.ToolCalls))
		for i, tc := range e.ToolCalls {
			chatToolCalls[i] = orchestrator.ChatToolCall{ID: tc.ID, Name: tc.Name, Arguments: tc.Arguments}
		}
		return orchestrator.ChatEvent{Type: orchestrator.ChatEventToolCall, ToolCalls: chatToolCalls}
	default:
		return orchestrator.ChatEvent{Type: orchestrator.ChatEventError, Error: e.Error}
	}
}

func validateEndpointSafety(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return &orchestrator.ValidationError{Err: err}
	}
	ip := net.ParseIP(u.Hostname())
	if isNonLoopbackPrivateIP(ip) {
		return &orchestrator.ValidationError{Err: errPrivateEndpoint}
	}
	return nil
}

func isNonLoopbackPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	return ip.IsPrivate() || ip.IsLinkLocalUnicast()
}

var errPrivateEndpoint = net.InvalidAddrError("endpoint must not point to private/internal addresses")
