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

type AnthropicClient struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

func NewAnthropicClient(endpoint, apiKey string) *AnthropicClient {
	return &AnthropicClient{
		endpoint: strings.TrimRight(endpoint, "/"),
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []json.RawMessage  `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature"`
	System      string             `json:"system,omitempty"`
	Stream      bool               `json:"stream"`
	Tools       []anthropicToolDef `json:"tools,omitempty"`
}

type anthropicToolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema"`
}

type anthropicStreamEvent struct {
	Type         string `json:"type"`
	Index        int    `json:"index"`
	ContentBlock *struct {
		Type  string `json:"type"`
		ID    string `json:"id,omitempty"`
		Name  string `json:"name,omitempty"`
		Text  string `json:"text,omitempty"`
	} `json:"content_block,omitempty"`
	Delta *struct {
		Type        string `json:"type"`
		Text        string `json:"text"`
		PartialJSON string `json:"partial_json"`
	} `json:"delta,omitempty"`
	Message *struct {
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message,omitempty"`
	Usage *struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage,omitempty"`
}

type anthropicToolCallAccumulator struct {
	blocks map[int]*ToolCall
	args   map[int]*strings.Builder
	order  []int
}

func newAnthropicToolCallAccumulator() *anthropicToolCallAccumulator {
	return &anthropicToolCallAccumulator{
		blocks: make(map[int]*ToolCall),
		args:   make(map[int]*strings.Builder),
	}
}

func (a *anthropicToolCallAccumulator) startBlock(index int, id, name string) {
	a.blocks[index] = &ToolCall{ID: id, Name: name}
	a.args[index] = &strings.Builder{}
	a.order = append(a.order, index)
}

func (a *anthropicToolCallAccumulator) appendJSON(index int, partial string) {
	if b, ok := a.args[index]; ok {
		b.WriteString(partial)
	}
}

func (a *anthropicToolCallAccumulator) finishBlock(index int) {
	if tc, ok := a.blocks[index]; ok {
		if b, ok := a.args[index]; ok {
			tc.Arguments = b.String()
		}
	}
}

func (a *anthropicToolCallAccumulator) hasToolCalls() bool {
	return len(a.blocks) > 0
}

func (a *anthropicToolCallAccumulator) emit(ch chan<- StreamEvent) {
	if !a.hasToolCalls() {
		return
	}
	calls := make([]ToolCall, 0, len(a.order))
	for _, idx := range a.order {
		if tc, ok := a.blocks[idx]; ok {
			calls = append(calls, *tc)
		}
	}
	ch <- StreamEvent{Type: EventToolCall, ToolCalls: calls}
	a.blocks = make(map[int]*ToolCall)
	a.args = make(map[int]*strings.Builder)
	a.order = nil
}

func (c *AnthropicClient) StreamChat(ctx context.Context, messages []Message, opts Options) (<-chan StreamEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, streamingDeadline)

	reqBody := c.buildRequest(messages, opts)
	body, err := json.Marshal(reqBody)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

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

func (c *AnthropicClient) buildRequest(messages []Message, opts Options) anthropicRequest {
	var systemPrompt string
	var anthropicMsgs []json.RawMessage

	for _, m := range messages {
		if m.Role == RoleSystem {
			systemPrompt = m.Content
			continue
		}
		anthropicMsgs = append(anthropicMsgs, buildAnthropicMessage(m))
	}

	req := anthropicRequest{
		Model:       opts.Model,
		Messages:    anthropicMsgs,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		System:      systemPrompt,
		Stream:      true,
	}

	for _, tool := range opts.Tools {
		req.Tools = append(req.Tools, anthropicToolDef{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			InputSchema: tool.Function.Parameters,
		})
	}

	return req
}

func buildAnthropicMessage(m Message) json.RawMessage {
	if len(m.ToolCalls) > 0 {
		return buildAnthropicAssistantToolCallMessage(m)
	}
	if m.Role == RoleTool {
		return buildAnthropicToolResultMessage(m)
	}
	msg := struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{Role: string(m.Role), Content: m.Content}
	b, _ := json.Marshal(msg)
	return b
}

func buildAnthropicAssistantToolCallMessage(m Message) json.RawMessage {
	type toolUseBlock struct {
		Type  string          `json:"type"`
		ID    string          `json:"id"`
		Name  string          `json:"name"`
		Input json.RawMessage `json:"input"`
	}
	var blocks []toolUseBlock
	for _, tc := range m.ToolCalls {
		input := json.RawMessage(tc.Arguments)
		if !json.Valid(input) {
			input = json.RawMessage("{}")
		}
		blocks = append(blocks, toolUseBlock{Type: "tool_use", ID: tc.ID, Name: tc.Name, Input: input})
	}
	msg := struct {
		Role    string         `json:"role"`
		Content []toolUseBlock `json:"content"`
	}{Role: "assistant", Content: blocks}
	b, _ := json.Marshal(msg)
	return b
}

func buildAnthropicToolResultMessage(m Message) json.RawMessage {
	type toolResultBlock struct {
		Type      string `json:"type"`
		ToolUseID string `json:"tool_use_id"`
		Content   string `json:"content"`
	}
	msg := struct {
		Role    string            `json:"role"`
		Content []toolResultBlock `json:"content"`
	}{
		Role: "user",
		Content: []toolResultBlock{{
			Type:      "tool_result",
			ToolUseID: m.ToolCallID,
			Content:   m.Content,
		}},
	}
	b, _ := json.Marshal(msg)
	return b
}

func (c *AnthropicClient) readStream(ctx context.Context, resp *http.Response, ch chan<- StreamEvent) {
	defer close(ch)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	totalTokens := 0
	acc := newAnthropicToolCallAccumulator()

	for scanner.Scan() {
		if ctx.Err() != nil {
			return
		}

		event, ok := parseAnthropicLine(scanner.Text())
		if !ok {
			continue
		}

		if done := c.handleAnthropicEvent(event, &totalTokens, acc, ch); done {
			return
		}
	}

	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		ch <- StreamEvent{Type: EventError, Error: fmt.Errorf("stream read error: %w", err)}
	}
}

func parseAnthropicLine(line string) (*anthropicStreamEvent, bool) {
	if !strings.HasPrefix(line, "data: ") {
		return nil, false
	}
	var event anthropicStreamEvent
	if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event); err != nil {
		return nil, false
	}
	return &event, true
}

func (c *AnthropicClient) handleAnthropicEvent(event *anthropicStreamEvent, totalTokens *int, acc *anthropicToolCallAccumulator, ch chan<- StreamEvent) bool {
	switch event.Type {
	case "content_block_start":
		handleContentBlockStart(event, acc)
	case "content_block_delta":
		handleContentBlockDelta(event, acc, ch)
	case "content_block_stop":
		acc.finishBlock(event.Index)
	case "message_delta":
		if event.Usage != nil {
			*totalTokens = event.Usage.InputTokens + event.Usage.OutputTokens
		}
	case "message_stop":
		acc.emit(ch)
		ch <- StreamEvent{Type: EventDone, TokensUsed: *totalTokens}
		return true
	}
	return false
}

func handleContentBlockStart(event *anthropicStreamEvent, acc *anthropicToolCallAccumulator) {
	if event.ContentBlock != nil && event.ContentBlock.Type == "tool_use" {
		acc.startBlock(event.Index, event.ContentBlock.ID, event.ContentBlock.Name)
	}
}

func handleContentBlockDelta(event *anthropicStreamEvent, acc *anthropicToolCallAccumulator, ch chan<- StreamEvent) {
	if event.Delta == nil {
		return
	}
	switch event.Delta.Type {
	case "text_delta":
		if event.Delta.Text != "" {
			ch <- StreamEvent{Type: EventToken, Content: event.Delta.Text}
		}
	case "input_json_delta":
		acc.appendJSON(event.Index, event.Delta.PartialJSON)
	}
}
