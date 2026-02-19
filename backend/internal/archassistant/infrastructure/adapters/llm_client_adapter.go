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
	llmMessages := make([]llm.Message, len(messages))
	for i, m := range messages {
		llmMessages[i] = llm.Message{Role: llm.Role(m.Role), Content: m.Content}
	}

	stream, err := a.client.StreamChat(ctx, llmMessages, llm.Options{
		Model:       opts.Model,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
	})
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

func convertEvent(e llm.StreamEvent) orchestrator.ChatEvent {
	switch e.Type {
	case llm.EventToken:
		return orchestrator.ChatEvent{Type: orchestrator.ChatEventToken, Content: e.Content}
	case llm.EventDone:
		return orchestrator.ChatEvent{Type: orchestrator.ChatEventDone, TokensUsed: e.TokensUsed}
	case llm.EventError:
		return orchestrator.ChatEvent{Type: orchestrator.ChatEventError, Error: e.Error}
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
