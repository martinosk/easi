package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"easi/backend/internal/archassistant/application/systemprompt"
	"easi/backend/internal/archassistant/application/tools"
	"easi/backend/internal/archassistant/domain"
	"easi/backend/internal/archassistant/domain/aggregates"
	"easi/backend/internal/archassistant/publishedlanguage"
)

const (
	charsPerToken        = 4
	defaultContextWindow = 128000
	maxIterations        = 50
	maxSameToolCalls     = 500
	maxToolResultChars   = 32000
	previewMaxChars      = 200
)

type SendMessageParams struct {
	ConversationID       string
	UserID               string
	Content              string
	TenantID             string
	UserRole             string
	AllowWriteOperations bool
	SystemPromptOverride *string
	Config               *publishedlanguage.AIConfigInfo
	Permissions          tools.PermissionChecker
	ToolRegistry         *tools.Registry
}

type Orchestrator struct {
	convRepo      domain.ConversationRepository
	clientFactory LLMClientFactory
}

func New(convRepo domain.ConversationRepository, clientFactory LLMClientFactory) *Orchestrator {
	return &Orchestrator{convRepo: convRepo, clientFactory: clientFactory}
}

func (o *Orchestrator) SendMessage(ctx context.Context, writer StreamWriter, params SendMessageParams) error {
	conv, err := o.prepareConversation(ctx, params)
	if err != nil {
		return err
	}

	if !o.hasTools(params) {
		log.Printf("[archassistant] sending WITHOUT tools (no registry or permissions)")
		return o.sendWithoutTools(ctx, writer, params, conv)
	}

	log.Printf("[archassistant] sending WITH agent loop (tools available)")
	return o.sendWithAgentLoop(ctx, writer, params, conv)
}

func (o *Orchestrator) hasTools(params SendMessageParams) bool {
	return params.ToolRegistry != nil && params.Permissions != nil
}

func (o *Orchestrator) sendWithoutTools(ctx context.Context, writer StreamWriter, params SendMessageParams, conv *aggregates.Conversation) error {
	stream, err := o.prepareStream(ctx, params, conv, nil)
	if err != nil {
		return wrapStreamError(err)
	}

	result, err := consumeStream(ctx, writer, stream)
	if err != nil {
		return wrapStreamError(err)
	}

	return o.persistAndComplete(ctx, writer, conv, result)
}

func (o *Orchestrator) sendWithAgentLoop(ctx context.Context, writer StreamWriter, params SendMessageParams, conv *aggregates.Conversation) error {
	messages, err := o.buildInitialMessages(ctx, params, conv)
	if err != nil {
		return wrapStreamError(err)
	}

	toolDefs := convertToolDefs(params.ToolRegistry.FormatForLLM(params.Permissions, params.AllowWriteOperations))
	log.Printf("[archassistant] %d tool definitions sent to LLM", len(toolDefs))
	toolCallCounts := make(map[string]int)

	for iteration := 0; iteration < maxIterations; iteration++ {
		agentResult, err := o.streamIteration(ctx, params, messages, toolDefs)
		if err != nil {
			return wrapStreamError(err)
		}

		applyTextToolCallFallback(&agentResult, params.ToolRegistry.ToolNames())

		if len(agentResult.toolCalls) == 0 {
			flushTokens(writer, agentResult.bufferedTokens)
			return o.persistAndComplete(ctx, writer, conv, agentResult.streamResult)
		}

		log.Printf("[archassistant] tool calls detected (iteration %d): %d calls", iteration, len(agentResult.toolCalls))
		if agentResult.content != "" {
			_ = writer.WriteToken(agentResult.content + "\n\n")
		}

		_ = writer.WriteThinking(ThinkingEvent{Message: "Processing..."})
		tcCtx := toolCallContext{
			ctx:         ctx,
			writer:      writer,
			permissions: params.Permissions,
			registry:    params.ToolRegistry,
			callCounts:  toolCallCounts,
		}
		toolMessages := o.executeToolCalls(tcCtx, agentResult.toolCalls)

		messages = append(messages, ChatMessage{Role: ChatRoleAssistant, ToolCalls: agentResult.toolCalls})
		messages = append(messages, toolMessages...)
	}

	return &IterationLimitError{}
}

func (o *Orchestrator) streamIteration(ctx context.Context, params SendMessageParams, messages []ChatMessage, toolDefs []interface{}) (agentStreamResult, error) {
	client, err := o.clientFactory.Create(params.Config.Provider, params.Config.Endpoint, params.Config.APIKey)
	if err != nil {
		return agentStreamResult{}, fmt.Errorf("failed to create LLM client: %w", err)
	}

	stream, err := client.StreamChat(ctx, messages, ChatOptions{
		Model:       params.Config.Model,
		MaxTokens:   params.Config.MaxTokens,
		Temperature: params.Config.Temperature,
		Tools:       toolDefs,
	})
	if err != nil {
		return agentStreamResult{}, err
	}

	return consumeAgentStreamBuffered(ctx, stream)
}

func (o *Orchestrator) buildInitialMessages(ctx context.Context, params SendMessageParams, conv *aggregates.Conversation) ([]ChatMessage, error) {
	history, err := o.convRepo.GetMessages(ctx, conv.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to load message history: %w", err)
	}

	sysPrompt := systemprompt.Build(systemprompt.BuildParams{
		TenantID:             params.TenantID,
		UserRole:             params.UserRole,
		AllowWriteOperations: params.AllowWriteOperations,
		SystemPromptOverride: params.SystemPromptOverride,
	})

	return buildChatMessages(sysPrompt, history, params.Config.MaxTokens), nil
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

func (o *Orchestrator) prepareStream(ctx context.Context, params SendMessageParams, conv *aggregates.Conversation, toolDefs []interface{}) (<-chan ChatEvent, error) {
	history, err := o.convRepo.GetMessages(ctx, conv.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to load message history: %w", err)
	}

	sysPrompt := systemprompt.Build(systemprompt.BuildParams{
		TenantID:             params.TenantID,
		UserRole:             params.UserRole,
		AllowWriteOperations: params.AllowWriteOperations,
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
		Tools:       toolDefs,
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

type agentStreamResult struct {
	streamResult
	toolCalls      []ChatToolCall
	bufferedTokens []string
}

func consumeAgentStreamBuffered(ctx context.Context, stream <-chan ChatEvent) (agentStreamResult, error) {
	var content strings.Builder
	var toolCalls []ChatToolCall
	var tokens []string
	var tokensUsed int

	for event := range stream {
		if ctx.Err() != nil {
			return agentStreamResult{}, ctx.Err()
		}
		switch event.Type {
		case ChatEventToolCall:
			toolCalls = append(toolCalls, event.ToolCalls...)
		case ChatEventToken:
			content.WriteString(event.Content)
			tokens = append(tokens, event.Content)
		case ChatEventDone:
			tokensUsed = event.TokensUsed
		case ChatEventError:
			return agentStreamResult{}, event.Error
		}
	}

	return agentStreamResult{
		streamResult:   streamResult{content: content.String(), tokensUsed: tokensUsed},
		toolCalls:      toolCalls,
		bufferedTokens: tokens,
	}, nil
}

func consumeAgentStream(ctx context.Context, writer StreamWriter, stream <-chan ChatEvent) (agentStreamResult, error) {
	var acc streamAccumulator
	var toolCalls []ChatToolCall
	for event := range stream {
		switch event.Type {
		case ChatEventToolCall:
			toolCalls = append(toolCalls, event.ToolCalls...)
		default:
			if err := acc.processEvent(ctx, event, writer); err != nil {
				return agentStreamResult{}, err
			}
		}
	}
	return agentStreamResult{
		streamResult: streamResult{content: acc.content.String(), tokensUsed: acc.tokensUsed},
		toolCalls:    toolCalls,
	}, nil
}

func flushTokens(writer StreamWriter, tokens []string) {
	for _, t := range tokens {
		_ = writer.WriteToken(t)
	}
}

type toolCallContext struct {
	ctx         context.Context
	writer      StreamWriter
	permissions tools.PermissionChecker
	registry    *tools.Registry
	callCounts  map[string]int
}

func (o *Orchestrator) executeToolCalls(tcCtx toolCallContext, toolCalls []ChatToolCall) []ChatMessage {
	var messages []ChatMessage
	for _, tc := range toolCalls {
		_ = tcCtx.writer.WriteToolCallStart(ToolCallStartEvent{
			ToolCallID: tc.ID,
			Name:       tc.Name,
			Arguments:  tc.Arguments,
		})

		messages = append(messages, o.executeSingleToolCall(tcCtx, tc))
	}
	return messages
}

func (o *Orchestrator) executeSingleToolCall(tcCtx toolCallContext, tc ChatToolCall) ChatMessage {
	tcCtx.callCounts[tc.Name]++
	if tcCtx.callCounts[tc.Name] > maxSameToolCalls {
		return o.buildToolLimitError(tcCtx, tc)
	}

	args := parseToolArgs(tc.Arguments)
	log.Printf("[archassistant] executing tool %q with args: %v", tc.Name, args)
	result, err := tcCtx.registry.Execute(tcCtx.ctx, tcCtx.permissions, tc.Name, args)
	if err != nil {
		log.Printf("[archassistant] tool %q execution error: %v", tc.Name, err)
		result = tools.ToolResult{Content: fmt.Sprintf("Error: %s", err.Error()), IsError: true}
	} else {
		log.Printf("[archassistant] tool %q returned %d chars (isError=%v)", tc.Name, len(result.Content), result.IsError)
	}

	content := truncateToolResult(result.Content)
	_ = tcCtx.writer.WriteToolCallResult(ToolCallResultEvent{
		ToolCallID:    tc.ID,
		Name:          tc.Name,
		ResultPreview: truncatePreview(result.Content),
	})

	return ChatMessage{
		Role:       ChatRoleTool,
		Content:    content,
		ToolCallID: tc.ID,
		Name:       tc.Name,
	}
}

func (o *Orchestrator) buildToolLimitError(tcCtx toolCallContext, tc ChatToolCall) ChatMessage {
	errContent := fmt.Sprintf("Tool call limit exceeded for %s: maximum %d calls per message", tc.Name, maxSameToolCalls)
	_ = tcCtx.writer.WriteToolCallResult(ToolCallResultEvent{
		ToolCallID:    tc.ID,
		Name:          tc.Name,
		ResultPreview: errContent,
	})
	return ChatMessage{
		Role:       ChatRoleTool,
		Content:    errContent,
		ToolCallID: tc.ID,
		Name:       tc.Name,
	}
}

func convertToolDefs(defs []tools.LLMToolDef) []interface{} {
	result := make([]interface{}, len(defs))
	for i, d := range defs {
		data, err := json.Marshal(d)
		if err != nil {
			continue
		}
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		result[i] = m
	}
	return result
}

func parseToolArgs(argsJSON string) map[string]interface{} {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return map[string]interface{}{}
	}
	return args
}

func truncateToolResult(s string) string {
	if len(s) <= maxToolResultChars {
		return s
	}
	return s[:maxToolResultChars] + "... [truncated]"
}

func truncatePreview(s string) string {
	if len(s) <= previewMaxChars {
		return s
	}
	return s[:previewMaxChars] + "..."
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
