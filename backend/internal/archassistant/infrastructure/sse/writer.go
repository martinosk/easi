package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"easi/backend/internal/archassistant/application/orchestrator"
)

type eventType string

type Writer struct {
	mu      sync.Mutex
	w       http.ResponseWriter
	flusher http.Flusher
}

func NewWriter(w http.ResponseWriter) (*Writer, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("response writer does not support flushing")
	}
	return &Writer{w: w, flusher: flusher}, nil
}

type tokenPayload struct {
	Content string `json:"content"`
}

func (s *Writer) WriteToken(content string) error {
	return s.writeEvent("token", tokenPayload{Content: content})
}

type donePayload struct {
	MessageID  string `json:"messageId"`
	TokensUsed int    `json:"tokensUsed"`
}

func (s *Writer) WriteDone(messageID string, tokensUsed int) error {
	return s.writeEvent("done", donePayload{MessageID: messageID, TokensUsed: tokensUsed})
}

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (s *Writer) WriteError(code, message string) error {
	return s.writeEvent("error", errorPayload{Code: code, Message: message})
}

type toolCallStartPayload struct {
	ToolCallID string `json:"toolCallId"`
	Name       string `json:"name"`
	Arguments  string `json:"arguments"`
}

func (s *Writer) WriteToolCallStart(event orchestrator.ToolCallStartEvent) error {
	return s.writeEvent("tool_call_start", toolCallStartPayload{
		ToolCallID: event.ToolCallID,
		Name:       event.Name,
		Arguments:  event.Arguments,
	})
}

type toolCallResultPayload struct {
	ToolCallID    string `json:"toolCallId"`
	Name          string `json:"name"`
	ResultPreview string `json:"resultPreview"`
}

func (s *Writer) WriteToolCallResult(event orchestrator.ToolCallResultEvent) error {
	return s.writeEvent("tool_call_result", toolCallResultPayload{
		ToolCallID:    event.ToolCallID,
		Name:          event.Name,
		ResultPreview: event.ResultPreview,
	})
}

type thinkingPayload struct {
	Message string `json:"message"`
}

func (s *Writer) WriteThinking(event orchestrator.ThinkingEvent) error {
	return s.writeEvent("thinking", thinkingPayload{Message: event.Message})
}

func (s *Writer) WritePing() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := fmt.Fprint(s.w, "event: ping\ndata: {}\n\n")
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

func (s *Writer) writeEvent(evt eventType, payload interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal SSE payload: %w", err)
	}
	_, err = fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", evt, data)
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

func SetSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
}
