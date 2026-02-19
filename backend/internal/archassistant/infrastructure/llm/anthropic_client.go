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
	Messages    []anthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature"`
	System      string             `json:"system,omitempty"`
	Stream      bool               `json:"stream"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicStreamEvent struct {
	Type  string `json:"type"`
	Delta *struct {
		Type string `json:"type"`
		Text string `json:"text"`
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
	var anthropicMsgs []anthropicMessage

	for _, m := range messages {
		if m.Role == RoleSystem {
			systemPrompt = m.Content
			continue
		}
		anthropicMsgs = append(anthropicMsgs, anthropicMessage{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}

	return anthropicRequest{
		Model:       opts.Model,
		Messages:    anthropicMsgs,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		System:      systemPrompt,
		Stream:      true,
	}
}

func (c *AnthropicClient) readStream(ctx context.Context, resp *http.Response, ch chan<- StreamEvent) {
	defer close(ch)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	totalTokens := 0

	for scanner.Scan() {
		if ctx.Err() != nil {
			return
		}

		event, ok := parseAnthropicLine(scanner.Text())
		if !ok {
			continue
		}

		if done := c.handleAnthropicEvent(event, &totalTokens, ch); done {
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

func (c *AnthropicClient) handleAnthropicEvent(event *anthropicStreamEvent, totalTokens *int, ch chan<- StreamEvent) bool {
	switch event.Type {
	case "content_block_delta":
		if event.Delta != nil && event.Delta.Text != "" {
			ch <- StreamEvent{Type: EventToken, Content: event.Delta.Text}
		}
	case "message_delta":
		if event.Usage != nil {
			*totalTokens = event.Usage.InputTokens + event.Usage.OutputTokens
		}
	case "message_stop":
		ch <- StreamEvent{Type: EventDone, TokensUsed: *totalTokens}
		return true
	}
	return false
}
