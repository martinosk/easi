package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ResponsesAPIClient implements Client for the OpenAI Responses API
// (POST /responses), used by Azure AI Foundry and modern OpenAI endpoints.
type ResponsesAPIClient struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

// NewResponsesAPIClient creates a client that targets the Responses API.
// endpoint must be the full URL including path and any query parameters,
// e.g. "https://cog-xxx.cognitiveservices.azure.com/openai/responses?api-version=2025-04-01-preview".
func NewResponsesAPIClient(endpoint, apiKey string) *ResponsesAPIClient {
	return &ResponsesAPIClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// IsResponsesAPIEndpoint reports whether the raw URL targets the Responses API
// (i.e. its path contains "/responses").
func IsResponsesAPIEndpoint(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.Contains(u.Path, "/responses")
}

// ---- request types ----

type responsesAPIRequest struct {
	Model           string                `json:"model"`
	Input           []json.RawMessage     `json:"input"`
	Instructions    string                `json:"instructions,omitempty"`
	MaxOutputTokens int                   `json:"max_output_tokens"`
	Temperature     float64               `json:"temperature"`
	Stream          bool                  `json:"stream"`
	Tools           []responsesAPIToolDef `json:"tools,omitempty"`
	ToolChoice      string                `json:"tool_choice,omitempty"`
}

type responsesAPIToolDef struct {
	Type        string      `json:"type"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// ---- input item helpers ----

type responsesInputMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responsesFunctionCallItem struct {
	Type      string `json:"type"` // "function_call"
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type responsesFunctionCallOutputItem struct {
	Type   string `json:"type"` // "function_call_output"
	CallID string `json:"call_id"`
	Output string `json:"output"`
}

type ResponsesToolCallAccumulator struct {
	callIDs map[string]string // item ID → call_id
	names   map[string]string // item ID → function name
	calls   []ToolCall
}

func NewResponsesToolCallAccumulator() *ResponsesToolCallAccumulator {
	return &ResponsesToolCallAccumulator{
		callIDs: make(map[string]string),
		names:   make(map[string]string),
	}
}

func (a *ResponsesToolCallAccumulator) RegisterItem(itemID, callID, name string) {
	if callID != "" {
		a.callIDs[itemID] = callID
	}
	if name != "" {
		a.names[itemID] = name
	}
}

func (a *ResponsesToolCallAccumulator) Finalize(itemID, name, arguments string) {
	callID := itemID
	if mapped, ok := a.callIDs[itemID]; ok {
		callID = mapped
	}
	if name == "" {
		name = a.names[itemID]
	}
	a.calls = append(a.calls, ToolCall{ID: callID, Name: name, Arguments: arguments})
}

func (a *ResponsesToolCallAccumulator) hasToolCalls() bool {
	return len(a.calls) > 0
}

func (a *ResponsesToolCallAccumulator) Emit(ch chan<- StreamEvent) {
	if a.hasToolCalls() {
		ch <- StreamEvent{Type: EventToolCall, ToolCalls: a.calls}
		a.calls = nil
	}
}

// ---- streaming event types ----

type responsesStreamEvent struct {
	Type      string          `json:"type"`
	Delta     string          `json:"delta"`     // response.output_text.delta
	ItemID    string          `json:"item_id"`   // response.function_call_arguments.done
	Name      string          `json:"name"`      // response.function_call_arguments.done
	Arguments string          `json:"arguments"` // response.function_call_arguments.done
	Item      json.RawMessage `json:"item"`      // response.output_item.added
	Response  *struct {
		Usage *struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	} `json:"response"` // response.completed
	Message string `json:"message"` // error
}

type responsesOutputItemPayload struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	CallID string `json:"call_id"`
	Name   string `json:"name"`
}

// ---- implementation ----

func (c *ResponsesAPIClient) buildRequest(messages []Message, opts Options) responsesAPIRequest {
	var instructions string
	var input []json.RawMessage

	for _, m := range messages {
		switch m.Role {
		case RoleSystem:
			instructions = m.Content
		case RoleUser:
			b, _ := json.Marshal(responsesInputMessage{Role: "user", Content: m.Content})
			input = append(input, b)
		case RoleAssistant:
			if len(m.ToolCalls) > 0 {
				for _, tc := range m.ToolCalls {
					b, _ := json.Marshal(responsesFunctionCallItem{
						Type:      "function_call",
						CallID:    tc.ID,
						Name:      tc.Name,
						Arguments: tc.Arguments,
					})
					input = append(input, b)
				}
			} else {
				b, _ := json.Marshal(responsesInputMessage{Role: "assistant", Content: m.Content})
				input = append(input, b)
			}
		case RoleTool:
			b, _ := json.Marshal(responsesFunctionCallOutputItem{
				Type:   "function_call_output",
				CallID: m.ToolCallID,
				Output: m.Content,
			})
			input = append(input, b)
		}
	}

	req := responsesAPIRequest{
		Model:           opts.Model,
		Input:           input,
		Instructions:    instructions,
		MaxOutputTokens: opts.MaxTokens,
		Temperature:     opts.Temperature,
		Stream:          true,
	}

	if len(opts.Tools) > 0 {
		for _, t := range opts.Tools {
			req.Tools = append(req.Tools, responsesAPIToolDef{
				Type:        "function",
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			})
		}
		req.ToolChoice = "auto"
	}

	return req
}

func (c *ResponsesAPIClient) StreamChat(ctx context.Context, messages []Message, opts Options) (<-chan StreamEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, streamingDeadline)

	reqBody := c.buildRequest(messages, opts)
	body, err := json.Marshal(reqBody)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(body))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
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

func (c *ResponsesAPIClient) readStream(ctx context.Context, resp *http.Response, ch chan<- StreamEvent) {
	defer close(ch)
	defer func() { _ = resp.Body.Close() }()

	scanner := bufio.NewScanner(resp.Body)
	// Increase buffer for large tool-call argument payloads.
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	totalTokens := 0
	acc := NewResponsesToolCallAccumulator()

	for scanner.Scan() {
		if ctx.Err() != nil {
			return
		}

		line := scanner.Text()
		// The Responses API SSE stream emits both "event:" and "data:" lines;
		// skip anything that is not a "data:" line.
		if strings.HasPrefix(line, "event:") || line == "" {
			continue
		}

		data, ok := parseSSEData(line)
		if !ok {
			continue
		}

		var event responsesStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		switch event.Type {
		case "response.output_text.delta":
			if event.Delta != "" {
				ch <- StreamEvent{Type: EventToken, Content: event.Delta}
			}

		case "response.output_item.added":
			var item responsesOutputItemPayload
			if err := json.Unmarshal(event.Item, &item); err == nil {
				if item.Type == "function_call" {
					acc.RegisterItem(item.ID, item.CallID, item.Name)
				}
			}

		case "response.function_call_arguments.done":
			acc.Finalize(event.ItemID, event.Name, event.Arguments)

		case "response.completed":
			if event.Response != nil && event.Response.Usage != nil {
				totalTokens = event.Response.Usage.TotalTokens
			}
			acc.Emit(ch)
			ch <- StreamEvent{Type: EventDone, TokensUsed: totalTokens}
			return

		case "response.failed", "error":
			errMsg := event.Message
			if errMsg == "" {
				errMsg = "LLM returned an error"
			}
			ch <- StreamEvent{Type: EventError, Error: fmt.Errorf("%s", errMsg)}
			return
		}
	}

	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		ch <- StreamEvent{Type: EventError, Error: fmt.Errorf("stream read error: %w", err)}
	}
}
