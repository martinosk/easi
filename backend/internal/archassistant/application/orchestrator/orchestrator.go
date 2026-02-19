package orchestrator

import (
	"context"
	"fmt"
	"strings"

	"easi/backend/internal/archassistant/application/systemprompt"
	"easi/backend/internal/archassistant/domain"
	"easi/backend/internal/archassistant/domain/aggregates"
	"easi/backend/internal/archassistant/publishedlanguage"
)

const (
	charsPerToken        = 4
	defaultContextWindow = 128000
)

type SendMessageParams struct {
	ConversationID       string
	UserID               string
	Content              string
	TenantID             string
	UserRole             string
	SystemPromptOverride *string
	Config               *publishedlanguage.AIConfigInfo
}

type Orchestrator struct {
	convRepo      domain.ConversationRepository
	clientFactory LLMClientFactory
}

func New(convRepo domain.ConversationRepository, clientFactory LLMClientFactory) *Orchestrator {
	return &Orchestrator{convRepo: convRepo, clientFactory: clientFactory}
}

func (o *Orchestrator) CreateConversation(ctx context.Context, conv *aggregates.Conversation) error {
	return o.convRepo.Create(ctx, conv)
}

func (o *Orchestrator) SendMessage(ctx context.Context, writer StreamWriter, params SendMessageParams) error {
	conv, err := o.prepareConversation(ctx, params)
	if err != nil {
		return err
	}

	stream, err := o.prepareStream(ctx, params, conv)
	if err != nil {
		return wrapStreamError(err)
	}

	result, err := consumeStream(ctx, writer, stream)
	if err != nil {
		return wrapStreamError(err)
	}

	return o.persistAndComplete(ctx, writer, conv, result)
}

func (o *Orchestrator) prepareConversation(ctx context.Context, params SendMessageParams) (*aggregates.Conversation, error) {
	conv, err := o.convRepo.GetByIDAndUser(ctx, params.ConversationID, params.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to load conversation: %w", err)
	}
	if conv == nil {
		return nil, ErrConversationNotFound
	}

	userMsg, err := conv.AddUserMessage(params.Content)
	if err != nil {
		return nil, &ValidationError{Err: err}
	}

	if err := o.convRepo.SaveMessage(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	return conv, nil
}

func (o *Orchestrator) prepareStream(ctx context.Context, params SendMessageParams, conv *aggregates.Conversation) (<-chan ChatEvent, error) {
	history, err := o.convRepo.GetMessages(ctx, conv.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to load message history: %w", err)
	}

	sysPrompt := systemprompt.Build(systemprompt.BuildParams{
		TenantID:             params.TenantID,
		UserRole:             params.UserRole,
		SystemPromptOverride: params.SystemPromptOverride,
	})

	messages := buildChatMessages(sysPrompt, history, params.Config.MaxTokens)

	client, err := o.clientFactory.Create(params.Config.Provider, params.Config.Endpoint, params.Config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	return client.StreamChat(ctx, messages, ChatOptions{
		Model:       params.Config.Model,
		MaxTokens:   params.Config.MaxTokens,
		Temperature: params.Config.Temperature,
	})
}

type streamAccumulator struct {
	content    strings.Builder
	tokensUsed int
}

type streamResult struct {
	content    string
	tokensUsed int
}

func (a *streamAccumulator) processEvent(ctx context.Context, event ChatEvent, writer StreamWriter) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	switch event.Type {
	case ChatEventToken:
		a.content.WriteString(event.Content)
		return writer.WriteToken(event.Content)
	case ChatEventDone:
		a.tokensUsed = event.TokensUsed
	case ChatEventError:
		return event.Error
	}
	return nil
}

func consumeStream(ctx context.Context, writer StreamWriter, stream <-chan ChatEvent) (streamResult, error) {
	var acc streamAccumulator
	for event := range stream {
		if err := acc.processEvent(ctx, event, writer); err != nil {
			return streamResult{}, err
		}
	}
	return streamResult{content: acc.content.String(), tokensUsed: acc.tokensUsed}, nil
}

func (o *Orchestrator) persistAndComplete(ctx context.Context, writer StreamWriter, conv *aggregates.Conversation, result streamResult) error {
	assistantMsg := conv.AddAssistantMessage(result.content, result.tokensUsed)

	if err := o.convRepo.SaveMessage(ctx, assistantMsg); err != nil {
		return fmt.Errorf("failed to save assistant message: %w", err)
	}

	if err := o.convRepo.UpdateConversation(ctx, conv); err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return writer.WriteDone(assistantMsg.ID(), result.tokensUsed)
}

func buildChatMessages(sysPrompt string, history []*aggregates.Message, maxTokens int) []ChatMessage {
	messages := []ChatMessage{
		{Role: ChatRoleSystem, Content: sysPrompt},
	}

	budget := estimateTokenBudget(maxTokens, sysPrompt)

	historyMessages := convertHistory(history)
	historyMessages = truncateToFit(historyMessages, budget)
	messages = append(messages, historyMessages...)

	return messages
}

func estimateTokenBudget(maxTokens int, sysPrompt string) int {
	sysTokens := len(sysPrompt) / charsPerToken
	budget := defaultContextWindow - sysTokens - maxTokens
	if budget < maxTokens {
		budget = maxTokens
	}
	return budget
}

func convertHistory(messages []*aggregates.Message) []ChatMessage {
	result := make([]ChatMessage, 0, len(messages))
	for _, m := range messages {
		result = append(result, ChatMessage{
			Role:    ChatRole(m.Role()),
			Content: m.Content(),
		})
	}
	return result
}

func truncateToFit(messages []ChatMessage, budget int) []ChatMessage {
	totalTokens := 0
	for _, m := range messages {
		totalTokens += len(m.Content) / charsPerToken
	}

	for totalTokens > budget && len(messages) > 1 {
		totalTokens -= len(messages[0].Content) / charsPerToken
		messages = messages[1:]
	}

	return messages
}

func wrapStreamError(err error) error {
	if err == context.Canceled || err == context.DeadlineExceeded {
		return &TimeoutError{Err: err}
	}
	msg := err.Error()
	if strings.Contains(msg, "LLM") || strings.Contains(msg, "status") {
		return &LLMError{Message: msg}
	}
	return &LLMError{Message: msg}
}
