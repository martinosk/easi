package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

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

func (s *Writer) writeEvent(eventType string, payload interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal SSE payload: %w", err)
	}
	_, err = fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", eventType, data)
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
