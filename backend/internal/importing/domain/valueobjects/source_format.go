package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
)

var ErrInvalidSourceFormat = errors.New("invalid source format: must be 'archimate-openexchange'")

const (
	SourceFormatArchiMateOpenExchange = "archimate-openexchange"
)

type SourceFormat struct {
	value string
}

func NewSourceFormat(value string) (SourceFormat, error) {
	if value != SourceFormatArchiMateOpenExchange {
		return SourceFormat{}, ErrInvalidSourceFormat
	}
	return SourceFormat{value: value}, nil
}

func (sf SourceFormat) Value() string {
	return sf.value
}

func (sf SourceFormat) IsArchiMateOpenExchange() bool {
	return sf.value == SourceFormatArchiMateOpenExchange
}

func (sf SourceFormat) Equals(other domain.ValueObject) bool {
	if otherSF, ok := other.(SourceFormat); ok {
		return sf.value == otherSF.value
	}
	return false
}

func (sf SourceFormat) String() string {
	return sf.value
}
