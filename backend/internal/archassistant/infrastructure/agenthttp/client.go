package agenthttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

type Response struct {
	StatusCode int
	Body       []byte
}

func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

func (r *Response) ErrorMessage() string {
	var parsed struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(r.Body, &parsed); err == nil && parsed.Message != "" {
		return parsed.Message
	}
	return http.StatusText(r.StatusCode)
}

func (c *Client) Get(ctx context.Context, path string) (*Response, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *Client) Post(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

func (c *Client) Put(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.do(ctx, http.MethodPut, path, body)
}

func (c *Client) Delete(ctx context.Context, path string) (*Response, error) {
	return c.do(ctx, http.MethodDelete, path, nil)
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*Response, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "AgentToken "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
	}, nil
}
