package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type OpenAIClient struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

func NewOpenAIClient(endpoint, apiKey string) *OpenAIClient {
	return &OpenAIClient{
		endpoint: strings.TrimRight(endpoint, "/"),
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type openAIRequest struct {
	Model         string              `json:"model"`
	Messages      []openAIMessage     `json:"messages"`
	MaxTokens     int                 `json:"max_tokens"`
	Temperature   float64             `json:"temperature"`
	Stream        bool                `json:"stream"`
	StreamOptions *openAIStreamOptions `json:"stream_options,omitempty"`
	Tools         []ToolDef           `json:"tools,omitempty"`
	ToolChoice    string              `json:"tool_choice,omitempty"`
}

type openAIStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type openAIMessage struct {
	Role       string              `json:"role"`
	Content    string              `json:"content,omitempty"`
	ToolCalls  []openAIToolCallMsg `json:"tool_calls,omitempty"`
	ToolCallID string              `json:"tool_call_id,omitempty"`
}

type openAIToolCallMsg struct {
	ID       string                    `json:"id"`
	Type     string                    `json:"type"`
	Function openAIToolCallFunctionMsg `json:"function"`
}

type openAIToolCallFunctionMsg struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIStreamChunk struct {
	Choices []openAIStreamChoice `json:"choices"`
	Usage   *struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIStreamChoice struct {
	Delta        openAIStreamDelta `json:"delta"`
	FinishReason *string           `json:"finish_reason"`
}

type openAIStreamDelta struct {
	Content   string                    `json:"content"`
	ToolCalls []openAIStreamToolCallDelta `json:"tool_calls"`
}

type openAIStreamToolCallDelta struct {
	Index    int     `json:"index"`
	ID       string  `json:"id,omitempty"`
	Type     string  `json:"type,omitempty"`
	Function *struct {
		Name      string `json:"name,omitempty"`
		Arguments string `json:"arguments,omitempty"`
	} `json:"function,omitempty"`
}

type openAIToolCallAccumulator struct {
	calls []ToolCall
}

func (a *openAIToolCallAccumulator) accumulate(deltas []openAIStreamToolCallDelta) {
	for _, d := range deltas {
		a.ensureSlot(d.Index)
		a.applyDelta(d)
	}
}

func (a *openAIToolCallAccumulator) ensureSlot(index int) {
	for index >= len(a.calls) {
		a.calls = append(a.calls, ToolCall{})
	}
}

func (a *openAIToolCallAccumulator) applyDelta(d openAIStreamToolCallDelta) {
	if d.ID != "" {
		a.calls[d.Index].ID = d.ID
	}
	if d.Function == nil {
		return
	}
	if d.Function.Name != "" {
		a.calls[d.Index].Name = d.Function.Name
	}
	a.calls[d.Index].Arguments += d.Function.Arguments
}

func (a *openAIToolCallAccumulator) hasToolCalls() bool {
	return len(a.calls) > 0
}

func (a *openAIToolCallAccumulator) emit(ch chan<- StreamEvent) {
	if a.hasToolCalls() {
		ch <- StreamEvent{Type: EventToolCall, ToolCalls: a.calls}
		a.calls = nil
	}
}

func (c *OpenAIClient) StreamChat(ctx context.Context, messages []Message, opts Options) (<-chan StreamEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, streamingDeadline)

	reqBody := c.buildRequest(messages, opts)
	body, err := json.Marshal(reqBody)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		cancel()
		return nil, fmt.Errorf("LLM returned status %d", resp.StatusCode)
	}

	ch := make(chan StreamEvent, 64)
	go func() {
		defer cancel()
		c.readStream(ctx, resp, ch)
	}()
	return ch, nil
}

func (c *OpenAIClient) buildRequest(messages []Message, opts Options) openAIRequest {
	oaiMessages := make([]openAIMessage, len(messages))
	for i, m := range messages {
		msg := openAIMessage{Role: string(m.Role), Content: m.Content}
		for _, tc := range m.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, openAIToolCallMsg{
				ID:   tc.ID,
				Type: "function",
				Function: openAIToolCallFunctionMsg{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				},
			})
		}
		if m.ToolCallID != "" {
			msg.ToolCallID = m.ToolCallID
		}
		oaiMessages[i] = msg
	}

	req := openAIRequest{
		Model:         opts.Model,
		Messages:      oaiMessages,
		MaxTokens:     opts.MaxTokens,
		Temperature:   opts.Temperature,
		Stream:        true,
		StreamOptions: &openAIStreamOptions{IncludeUsage: true},
	}

	if len(opts.Tools) > 0 {
		req.Tools = opts.Tools
		req.ToolChoice = "auto"
	}

	return req
}

func (c *OpenAIClient) readStream(ctx context.Context, resp *http.Response, ch chan<- StreamEvent) {
	defer close(ch)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	totalTokens := 0
	acc := &openAIToolCallAccumulator{}

	for scanner.Scan() {
		if ctx.Err() != nil {
			return
		}

		if done := c.handleOpenAILine(scanner.Text(), &totalTokens, acc, ch); done {
			return
		}
	}

	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		ch <- StreamEvent{Type: EventError, Error: fmt.Errorf("stream read error: %w", err)}
	}
}

func (c *OpenAIClient) handleOpenAILine(line string, totalTokens *int, acc *openAIToolCallAccumulator, ch chan<- StreamEvent) bool {
	data, ok := parseSSEData(line)
	if !ok {
		return false
	}

	if data == "[DONE]" {
		acc.emit(ch)
		ch <- StreamEvent{Type: EventDone, TokensUsed: *totalTokens}
		return true
	}

	var chunk openAIStreamChunk
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return false
	}

	if chunk.Usage != nil {
		*totalTokens = chunk.Usage.TotalTokens
	}

	if len(chunk.Choices) > 0 {
		processOpenAIChoice(chunk.Choices[0], acc, ch)
	}

	return false
}

func parseSSEData(line string) (string, bool) {
	if !strings.HasPrefix(line, "data: ") {
		return "", false
	}
	return strings.TrimPrefix(line, "data: "), true
}

func processOpenAIChoice(choice openAIStreamChoice, acc *openAIToolCallAccumulator, ch chan<- StreamEvent) {
	if len(choice.Delta.ToolCalls) > 0 {
		acc.accumulate(choice.Delta.ToolCalls)
	}

	if choice.Delta.Content != "" {
		ch <- StreamEvent{Type: EventToken, Content: choice.Delta.Content}
	}

	if choice.FinishReason != nil && *choice.FinishReason == "tool_calls" {
		acc.emit(ch)
	}
}
