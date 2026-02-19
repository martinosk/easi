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
}

type openAIStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
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
		oaiMessages[i] = openAIMessage{Role: string(m.Role), Content: m.Content}
	}
	return openAIRequest{
		Model:         opts.Model,
		Messages:      oaiMessages,
		MaxTokens:     opts.MaxTokens,
		Temperature:   opts.Temperature,
		Stream:        true,
		StreamOptions: &openAIStreamOptions{IncludeUsage: true},
	}
}

func (c *OpenAIClient) readStream(ctx context.Context, resp *http.Response, ch chan<- StreamEvent) {
	defer close(ch)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	totalTokens := 0

	for scanner.Scan() {
		if ctx.Err() != nil {
			return
		}

		if done := c.handleOpenAILine(scanner.Text(), &totalTokens, ch); done {
			return
		}
	}

	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		ch <- StreamEvent{Type: EventError, Error: fmt.Errorf("stream read error: %w", err)}
	}
}

func (c *OpenAIClient) handleOpenAILine(line string, totalTokens *int, ch chan<- StreamEvent) bool {
	if !strings.HasPrefix(line, "data: ") {
		return false
	}

	data := strings.TrimPrefix(line, "data: ")
	if data == "[DONE]" {
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

	if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
		ch <- StreamEvent{Type: EventToken, Content: chunk.Choices[0].Delta.Content}
	}
	return false
}
