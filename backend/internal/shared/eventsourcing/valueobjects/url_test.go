package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewURL_AcceptsAbsoluteHTTPAndHTTPS(t *testing.T) {
	cases := []string{
		"https://contracts.example.com",
		"http://intranet.example/path?query=1",
		"  https://example.com/trimmed  ",
	}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			url, err := NewURL(raw)
			require.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(raw), url.Value())
		})
	}
}

func TestNewURL_RejectsEmpty(t *testing.T) {
	_, err := NewURL("   ")
	assert.ErrorIs(t, err, ErrURLEmpty)
}

func TestNewURL_RejectsNonAbsoluteOrNonHTTP(t *testing.T) {
	cases := []string{
		"ftp://x",
		"/relative/path",
		"example.com/no-scheme",
		"https://",
		"not a url at all",
		"mailto:someone@example.com",
	}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			_, err := NewURL(raw)
			assert.ErrorIs(t, err, ErrURLInvalid)
		})
	}
}

func TestNewURL_RejectsTooLong(t *testing.T) {
	raw := "https://example.com/" + strings.Repeat("a", MaxURLLength)
	_, err := NewURL(raw)
	assert.ErrorIs(t, err, ErrURLTooLong)
}

func TestURL_Equals(t *testing.T) {
	a, err := NewURL("https://example.com")
	require.NoError(t, err)
	b, err := NewURL("https://example.com")
	require.NoError(t, err)
	c, err := NewURL("https://other.example.com")
	require.NoError(t, err)

	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
	assert.Equal(t, "https://example.com", a.String())
}
