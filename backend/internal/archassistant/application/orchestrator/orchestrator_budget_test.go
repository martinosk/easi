package orchestrator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEstimateTokenBudget_FloorProtection(t *testing.T) {
	budget := estimateTokenBudget(127000, "short prompt")
	assert.Equal(t, 127000, budget, "budget should be floored to maxTokens when raw budget is too small")
}

func TestEstimateTokenBudget_NormalCase(t *testing.T) {
	budget := estimateTokenBudget(4096, "hello")
	expected := defaultContextWindow - len("hello")/charsPerToken - 4096
	assert.Equal(t, expected, budget)
}

func TestTruncateToFit_PreservesLastMessage(t *testing.T) {
	messages := []ChatMessage{
		{Role: ChatRoleUser, Content: strings.Repeat("x", 40000)},
	}
	result := truncateToFit(messages, 100)
	assert.Len(t, result, 1, "last remaining message should never be removed")
}

func TestTruncateToFit_RemovesOldestFirst(t *testing.T) {
	messages := []ChatMessage{
		{Role: ChatRoleUser, Content: strings.Repeat("x", 4000)},
		{Role: ChatRoleAssistant, Content: strings.Repeat("y", 4000)},
		{Role: ChatRoleUser, Content: "latest"},
	}
	result := truncateToFit(messages, 500)
	assert.Len(t, result, 1)
	assert.Equal(t, "latest", result[0].Content)
}

func TestTruncateToFit_KeepsAllWhenUnderBudget(t *testing.T) {
	messages := []ChatMessage{
		{Role: ChatRoleUser, Content: "short"},
		{Role: ChatRoleAssistant, Content: "also short"},
	}
	result := truncateToFit(messages, 100000)
	assert.Len(t, result, 2)
}
