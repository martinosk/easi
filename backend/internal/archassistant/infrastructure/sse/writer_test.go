package sse_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/archassistant/application/orchestrator"
	"easi/backend/internal/archassistant/infrastructure/sse"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushCount int
}

func newFlushRecorder() *flushRecorder {
	return &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
}

func (f *flushRecorder) Flush() {
	f.flushCount++
	f.ResponseRecorder.Flush()
}

type sseTestCase struct {
	name       string
	write      func(*sse.Writer) error
	expected   []string
	flushCount int
}

func runSSETests(t *testing.T, tests []sseTestCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newFlushRecorder()
			writer, err := sse.NewWriter(rec)
			require.NoError(t, err)

			require.NoError(t, tt.write(writer))

			body := rec.Body.String()
			for _, exp := range tt.expected {
				assert.Contains(t, body, exp)
			}
			assert.Equal(t, tt.flushCount, rec.flushCount)
		})
	}
}

func TestNewWriter_RequiresFlusher(t *testing.T) {
	w := httptest.NewRecorder()
	_, err := sse.NewWriter(struct{ http.ResponseWriter }{w})
	assert.Error(t, err)
}

func TestSSEWriter_WriteEvents(t *testing.T) {
	runSSETests(t, []sseTestCase{
		{
			name:       "token",
			write:      func(w *sse.Writer) error { return w.WriteToken("Hello") },
			expected:   []string{"event: token\n", `"content":"Hello"`},
			flushCount: 1,
		},
		{
			name:       "done",
			write:      func(w *sse.Writer) error { return w.WriteDone("msg-123", 42) },
			expected:   []string{"event: done\n", `"messageId":"msg-123"`, `"tokensUsed":42`},
			flushCount: 1,
		},
		{
			name:       "error",
			write:      func(w *sse.Writer) error { return w.WriteError("llm_error", "check config") },
			expected:   []string{"event: error\n", `"code":"llm_error"`, `"message":"check config"`},
			flushCount: 1,
		},
		{
			name:       "ping",
			write:      func(w *sse.Writer) error { return w.WritePing() },
			expected:   []string{"event: ping\n", "data: {}\n"},
			flushCount: 1,
		},
	})
}

func TestSSEWriter_WriteToolAndThinkingEvents(t *testing.T) {
	runSSETests(t, []sseTestCase{
		{
			name: "tool_call_start",
			write: func(w *sse.Writer) error {
				return w.WriteToolCallStart(orchestrator.ToolCallStartEvent{
					ToolCallID: "call-1",
					Name:       "list_applications",
					Arguments:  `{"limit":10}`,
				})
			},
			expected: []string{
				"event: tool_call_start\n",
				`"toolCallId":"call-1"`,
				`"name":"list_applications"`,
				`"arguments":"{\"limit\":10}"`,
			},
			flushCount: 1,
		},
		{
			name: "tool_call_result",
			write: func(w *sse.Writer) error {
				return w.WriteToolCallResult(orchestrator.ToolCallResultEvent{
					ToolCallID:    "call-1",
					Name:          "list_applications",
					ResultPreview: "Found 3 applications",
				})
			},
			expected: []string{
				"event: tool_call_result\n",
				`"toolCallId":"call-1"`,
				`"name":"list_applications"`,
				`"resultPreview":"Found 3 applications"`,
			},
			flushCount: 1,
		},
		{
			name: "thinking",
			write: func(w *sse.Writer) error {
				return w.WriteThinking(orchestrator.ThinkingEvent{
					Message: "Analyzing the architecture...",
				})
			},
			expected: []string{
				"event: thinking\n",
				`"message":"Analyzing the architecture..."`,
			},
			flushCount: 1,
		},
	})
}

func TestSetSSEHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	sse.SetSSEHeaders(rec)

	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", rec.Header().Get("Connection"))
}
