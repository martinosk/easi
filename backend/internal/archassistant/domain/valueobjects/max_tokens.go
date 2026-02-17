package valueobjects

import "errors"

var ErrMaxTokensOutOfRange = errors.New("max tokens must be between 256 and 32768")

const (
	MinMaxTokens     = 256
	MaxMaxTokens     = 32768
	DefaultMaxTokens = 4096
)

type MaxTokens struct {
	value int
}

func NewMaxTokens(value int) (MaxTokens, error) {
	if value < MinMaxTokens || value > MaxMaxTokens {
		return MaxTokens{}, ErrMaxTokensOutOfRange
	}
	return MaxTokens{value: value}, nil
}

func DefaultMaxTokensValue() MaxTokens {
	return MaxTokens{value: DefaultMaxTokens}
}

func ReconstructMaxTokens(value int) MaxTokens {
	return MaxTokens{value: value}
}

func (m MaxTokens) Value() int { return m.value }
