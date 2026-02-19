package llm

import "fmt"

func NewClient(provider, endpoint, apiKey string) (Client, error) {
	switch provider {
	case "openai":
		return NewOpenAIClient(endpoint, apiKey), nil
	case "anthropic":
		return NewAnthropicClient(endpoint, apiKey), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}
