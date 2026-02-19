package adapters

import (
	"testing"

	"easi/backend/internal/archassistant/application/orchestrator"
	"easi/backend/internal/archassistant/infrastructure/llm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertEvent_Token(t *testing.T) {
	event := convertEvent(llm.StreamEvent{Type: llm.EventToken, Content: "hello"})
	assert.Equal(t, orchestrator.ChatEventToken, event.Type)
	assert.Equal(t, "hello", event.Content)
}

func TestConvertEvent_Done(t *testing.T) {
	event := convertEvent(llm.StreamEvent{Type: llm.EventDone, TokensUsed: 42})
	assert.Equal(t, orchestrator.ChatEventDone, event.Type)
	assert.Equal(t, 42, event.TokensUsed)
}

func TestConvertEvent_Error(t *testing.T) {
	event := convertEvent(llm.StreamEvent{Type: llm.EventError, Error: assert.AnError})
	assert.Equal(t, orchestrator.ChatEventError, event.Type)
	assert.Equal(t, assert.AnError, event.Error)
}

func TestConvertEvent_ToolCall(t *testing.T) {
	event := convertEvent(llm.StreamEvent{
		Type: llm.EventToolCall,
		ToolCalls: []llm.ToolCall{
			{ID: "call_1", Name: "list_apps", Arguments: `{"name":"test"}`},
			{ID: "call_2", Name: "get_vendor", Arguments: `{"id":2}`},
		},
	})

	assert.Equal(t, orchestrator.ChatEventToolCall, event.Type)
	require.Len(t, event.ToolCalls, 2)
	assert.Equal(t, "call_1", event.ToolCalls[0].ID)
	assert.Equal(t, "list_apps", event.ToolCalls[0].Name)
	assert.Equal(t, `{"name":"test"}`, event.ToolCalls[0].Arguments)
	assert.Equal(t, "call_2", event.ToolCalls[1].ID)
	assert.Equal(t, "get_vendor", event.ToolCalls[1].Name)
	assert.Equal(t, `{"id":2}`, event.ToolCalls[1].Arguments)
}

func TestConvertMessages_WithToolCallFields(t *testing.T) {
	messages := []orchestrator.ChatMessage{
		{Role: orchestrator.ChatRoleUser, Content: "List apps"},
		{
			Role: orchestrator.ChatRoleAssistant,
			ToolCalls: []orchestrator.ChatToolCall{
				{ID: "call_1", Name: "list_apps", Arguments: `{}`},
			},
		},
		{
			Role:       orchestrator.ChatRoleTool,
			Content:    `[{"name":"App1"}]`,
			ToolCallID: "call_1",
			Name:       "list_apps",
		},
	}

	llmMessages := convertMessages(messages)

	require.Len(t, llmMessages, 3)
	assert.Equal(t, llm.RoleUser, llmMessages[0].Role)
	assert.Equal(t, "List apps", llmMessages[0].Content)

	assert.Equal(t, llm.RoleAssistant, llmMessages[1].Role)
	require.Len(t, llmMessages[1].ToolCalls, 1)
	assert.Equal(t, "call_1", llmMessages[1].ToolCalls[0].ID)
	assert.Equal(t, "list_apps", llmMessages[1].ToolCalls[0].Name)

	assert.Equal(t, llm.RoleTool, llmMessages[2].Role)
	assert.Equal(t, `[{"name":"App1"}]`, llmMessages[2].Content)
	assert.Equal(t, "call_1", llmMessages[2].ToolCallID)
	assert.Equal(t, "list_apps", llmMessages[2].Name)
}

func TestConvertOptions_WithTools(t *testing.T) {
	opts := orchestrator.ChatOptions{
		Model:       "gpt-4",
		MaxTokens:   100,
		Temperature: 0.5,
		Tools: []interface{}{
			map[string]interface{}{"type": "function", "function": map[string]interface{}{"name": "test"}},
		},
	}

	llmOpts := convertOptions(opts)

	assert.Equal(t, "gpt-4", llmOpts.Model)
	assert.Equal(t, 100, llmOpts.MaxTokens)
	assert.InDelta(t, 0.5, llmOpts.Temperature, 0.001)
	require.Len(t, llmOpts.Tools, 1)
	assert.Equal(t, "function", llmOpts.Tools[0].Type)
	assert.Equal(t, "test", llmOpts.Tools[0].Function.Name)
}
