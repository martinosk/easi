package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurationStatusFromString_ValidStatuses(t *testing.T) {
	tests := []struct {
		input    string
		expected ConfigurationStatus
	}{
		{"not_configured", StatusNotConfigured},
		{"configured", StatusConfigured},
		{"error", StatusError},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			status, err := ConfigurationStatusFromString(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, status)
			assert.Equal(t, tc.input, status.Value())
		})
	}
}

func TestConfigurationStatusFromString_InvalidStatus(t *testing.T) {
	_, err := ConfigurationStatusFromString("invalid")
	assert.ErrorIs(t, err, ErrInvalidConfigurationStatus)
}

func TestConfigurationStatusFromString_EmptyString(t *testing.T) {
	_, err := ConfigurationStatusFromString("")
	assert.ErrorIs(t, err, ErrInvalidConfigurationStatus)
}

func TestConfigurationStatus_IsConfigured(t *testing.T) {
	assert.True(t, StatusConfigured.IsConfigured())
	assert.False(t, StatusNotConfigured.IsConfigured())
	assert.False(t, StatusError.IsConfigured())
}

func TestEncryptedAPIKey_NewAndValue(t *testing.T) {
	key := NewEncryptedAPIKey("encrypted-value")
	assert.Equal(t, "encrypted-value", key.Value())
	assert.False(t, key.IsEmpty())
}

func TestEncryptedAPIKey_Empty(t *testing.T) {
	key := NewEncryptedAPIKey("")
	assert.True(t, key.IsEmpty())
	assert.Equal(t, "", key.Value())
}

func TestNewLLMEndpoint_ValidURLs(t *testing.T) {
	valid := []struct {
		name  string
		input string
	}{
		{"simple https", "https://api.openai.com"},
		{"https with path", "https://api.openai.com/v1"},
		{"https with port", "https://api.example.com:8443"},
		{"localhost with port", "http://localhost:8080"},
		{"localhost no port", "http://localhost"},
		{"localhost with path", "http://localhost:11434/api"},
	}
	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			ep, err := NewLLMEndpoint(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.input, ep.Value())
		})
	}
}

func TestNewLLMEndpoint_Empty(t *testing.T) {
	_, err := NewLLMEndpoint("")
	assert.ErrorIs(t, err, ErrEndpointEmpty)
}

func TestNewLLMEndpoint_WhitespaceOnly(t *testing.T) {
	_, err := NewLLMEndpoint("   ")
	assert.ErrorIs(t, err, ErrEndpointEmpty)
}

func TestNewLLMEndpoint_TooLong(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("a", 500)
	_, err := NewLLMEndpoint(long)
	assert.ErrorIs(t, err, ErrEndpointTooLong)
}

func TestNewLLMEndpoint_InvalidScheme(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"plain http non-localhost", "http://api.openai.com"},
		{"localhost prefix domain", "http://localhost.evil.com"},
		{"localhost prefix domain with port", "http://localhost.evil.com:8080"},
		{"ftp", "ftp://api.example.com"},
		{"no scheme", "api.openai.com"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewLLMEndpoint(tc.input)
			assert.ErrorIs(t, err, ErrEndpointInvalid)
		})
	}
}

func TestNewLLMEndpoint_Trimming(t *testing.T) {
	ep, err := NewLLMEndpoint("  https://api.openai.com  ")
	require.NoError(t, err)
	assert.Equal(t, "https://api.openai.com", ep.Value())
}

func TestReconstructLLMEndpoint(t *testing.T) {
	ep := ReconstructLLMEndpoint("any-value")
	assert.Equal(t, "any-value", ep.Value())
}

func TestNewModelName_Valid(t *testing.T) {
	m, err := NewModelName("gpt-4o")
	require.NoError(t, err)
	assert.Equal(t, "gpt-4o", m.Value())
}

func TestNewModelName_Trimming(t *testing.T) {
	m, err := NewModelName("  claude-3  ")
	require.NoError(t, err)
	assert.Equal(t, "claude-3", m.Value())
}

func TestNewModelName_Empty(t *testing.T) {
	_, err := NewModelName("")
	assert.ErrorIs(t, err, ErrModelNameEmpty)
}

func TestNewModelName_WhitespaceOnly(t *testing.T) {
	_, err := NewModelName("   ")
	assert.ErrorIs(t, err, ErrModelNameEmpty)
}

func TestNewModelName_TooLong(t *testing.T) {
	_, err := NewModelName(strings.Repeat("a", 101))
	assert.ErrorIs(t, err, ErrModelNameTooLong)
}

func TestNewModelName_ExactlyMaxLength(t *testing.T) {
	m, err := NewModelName(strings.Repeat("a", 100))
	require.NoError(t, err)
	assert.Len(t, m.Value(), 100)
}

func TestReconstructModelName(t *testing.T) {
	m := ReconstructModelName("any-model")
	assert.Equal(t, "any-model", m.Value())
}

func TestNewMaxTokens(t *testing.T) {
	t.Run("accepts minimum", func(t *testing.T) {
		mt, err := NewMaxTokens(MinMaxTokens)
		require.NoError(t, err)
		assert.Equal(t, MinMaxTokens, mt.Value())
	})
	t.Run("accepts maximum", func(t *testing.T) {
		mt, err := NewMaxTokens(MaxMaxTokens)
		require.NoError(t, err)
		assert.Equal(t, MaxMaxTokens, mt.Value())
	})
	t.Run("accepts mid range", func(t *testing.T) {
		mt, err := NewMaxTokens(4096)
		require.NoError(t, err)
		assert.Equal(t, 4096, mt.Value())
	})
	t.Run("rejects below minimum", func(t *testing.T) {
		_, err := NewMaxTokens(255)
		assert.ErrorIs(t, err, ErrMaxTokensOutOfRange)
	})
	t.Run("rejects above maximum", func(t *testing.T) {
		_, err := NewMaxTokens(32769)
		assert.ErrorIs(t, err, ErrMaxTokensOutOfRange)
	})
}

func TestDefaultMaxTokensValue(t *testing.T) {
	mt := DefaultMaxTokensValue()
	assert.Equal(t, DefaultMaxTokens, mt.Value())
}

func TestReconstructMaxTokens(t *testing.T) {
	mt := ReconstructMaxTokens(999)
	assert.Equal(t, 999, mt.Value())
}

func TestNewTemperature(t *testing.T) {
	t.Run("accepts zero", func(t *testing.T) {
		temp, err := NewTemperature(0.0)
		require.NoError(t, err)
		assert.Equal(t, 0.0, temp.Value())
	})
	t.Run("accepts maximum", func(t *testing.T) {
		temp, err := NewTemperature(2.0)
		require.NoError(t, err)
		assert.Equal(t, 2.0, temp.Value())
	})
	t.Run("accepts default", func(t *testing.T) {
		temp, err := NewTemperature(0.3)
		require.NoError(t, err)
		assert.Equal(t, 0.3, temp.Value())
	})
	t.Run("rejects below minimum", func(t *testing.T) {
		_, err := NewTemperature(-0.1)
		assert.ErrorIs(t, err, ErrTemperatureOutOfRange)
	})
	t.Run("rejects above maximum", func(t *testing.T) {
		_, err := NewTemperature(2.1)
		assert.ErrorIs(t, err, ErrTemperatureOutOfRange)
	})
}

func TestDefaultTemperatureValue(t *testing.T) {
	temp := DefaultTemperatureValue()
	assert.Equal(t, DefaultTemperature, temp.Value())
}

func TestReconstructTemperature(t *testing.T) {
	temp := ReconstructTemperature(1.5)
	assert.Equal(t, 1.5, temp.Value())
}

func TestNewLLMProvider_ValidProviders(t *testing.T) {
	tests := []struct {
		input    string
		expected LLMProvider
	}{
		{"openai", ProviderOpenAI},
		{"anthropic", ProviderAnthropic},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			p, err := NewLLMProvider(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, p)
			assert.Equal(t, tc.input, p.Value())
		})
	}
}

func TestNewLLMProvider_InvalidProvider(t *testing.T) {
	_, err := NewLLMProvider("invalid")
	assert.ErrorIs(t, err, ErrInvalidLLMProvider)
}

func TestNewLLMProvider_EmptyString(t *testing.T) {
	_, err := NewLLMProvider("")
	assert.ErrorIs(t, err, ErrInvalidLLMProvider)
}

func TestLLMProvider_DefaultEndpoint(t *testing.T) {
	assert.Equal(t, "https://api.openai.com", ProviderOpenAI.DefaultEndpoint())
	assert.Equal(t, "https://api.anthropic.com", ProviderAnthropic.DefaultEndpoint())
}

func TestLLMProvider_IsOpenAI(t *testing.T) {
	assert.True(t, ProviderOpenAI.IsOpenAI())
	assert.False(t, ProviderAnthropic.IsOpenAI())
}

func TestLLMProvider_IsAnthropic(t *testing.T) {
	assert.True(t, ProviderAnthropic.IsAnthropic())
	assert.False(t, ProviderOpenAI.IsAnthropic())
}

func TestReconstructLLMProvider(t *testing.T) {
	p := ReconstructLLMProvider("openai")
	assert.Equal(t, "openai", p.Value())
}
