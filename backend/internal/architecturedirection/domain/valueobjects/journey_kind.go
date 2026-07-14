package valueobjects

import (
	"errors"
	"fmt"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvalidJourneyKind            = errors.New("journey kind must be one of migration, consolidation, carve-out, move")
	ErrInvalidSourceApplicationCount = errors.New("source application count does not match the journey kind")
)

const (
	JourneyKindMigration     = "migration"
	JourneyKindConsolidation = "consolidation"
	JourneyKindCarveOut      = "carve-out"
	JourneyKindMove          = "move"
)

type JourneyKind struct {
	value string
}

func NewJourneyKind(value string) (JourneyKind, error) {
	switch value {
	case JourneyKindMigration, JourneyKindConsolidation, JourneyKindCarveOut, JourneyKindMove:
		return JourneyKind{value: value}, nil
	default:
		return JourneyKind{}, ErrInvalidJourneyKind
	}
}

func (k JourneyKind) Value() string { return k.value }

func (k JourneyKind) IsMove() bool { return k.value == JourneyKindMove }

func (k JourneyKind) ValidateSourceCount(n int) error {
	switch k.value {
	case JourneyKindMigration:
		if n < 1 {
			return fmt.Errorf("%w: migration requires at least 1 from-application", ErrInvalidSourceApplicationCount)
		}
	case JourneyKindConsolidation:
		if n < 2 {
			return fmt.Errorf("%w: consolidation requires at least 2 from-applications", ErrInvalidSourceApplicationCount)
		}
	case JourneyKindCarveOut:
		if n != 1 {
			return fmt.Errorf("%w: carve-out requires exactly 1 from-application", ErrInvalidSourceApplicationCount)
		}
	}
	return nil
}

func (k JourneyKind) Equals(other domain.ValueObject) bool {
	if o, ok := other.(JourneyKind); ok {
		return k.value == o.value
	}
	return false
}
