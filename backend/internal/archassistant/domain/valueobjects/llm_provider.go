package valueobjects

import "errors"

var ErrInvalidLLMProvider = errors.New("invalid LLM provider: must be openai or anthropic")

type LLMProvider struct {
	value string
}

var (
	ProviderOpenAI    = LLMProvider{value: "openai"}
	ProviderAnthropic = LLMProvider{value: "anthropic"}
)

func NewLLMProvider(s string) (LLMProvider, error) {
	switch s {
	case "openai":
		return ProviderOpenAI, nil
	case "anthropic":
		return ProviderAnthropic, nil
	default:
		return LLMProvider{}, ErrInvalidLLMProvider
	}
}

func ReconstructLLMProvider(s string) LLMProvider {
	return LLMProvider{value: s}
}

func (p LLMProvider) Value() string { return p.value }

func (p LLMProvider) IsOpenAI() bool { return p.value == "openai" }

func (p LLMProvider) IsAnthropic() bool { return p.value == "anthropic" }

func (p LLMProvider) DefaultEndpoint() string {
	switch p.value {
	case "openai":
		return "https://api.openai.com"
	case "anthropic":
		return "https://api.anthropic.com"
	default:
		return ""
	}
}
