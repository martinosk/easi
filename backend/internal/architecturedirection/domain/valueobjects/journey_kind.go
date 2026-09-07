package valueobjects

import (
	"errors"
	"fmt"

	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrInvalidJourneyKind            = errors.New("journey kind must be one of migration, consolidation, carve-out, move, maturity")
	ErrInvalidSourceApplicationCount = errors.New("source application count does not match the journey kind")
)

const (
	JourneyKindMigration     = "migration"
	JourneyKindConsolidation = "consolidation"
	JourneyKindCarveOut      = "carve-out"
	JourneyKindMove          = "move"
	JourneyKindMaturity      = "maturity"
)

var applicationJourneyKinds = []string{
	JourneyKindMigration, JourneyKindConsolidation, JourneyKindCarveOut, JourneyKindMove,
}

type JourneyKind struct {
	value string
}

func NewJourneyKind(value string) (JourneyKind, error) {
	switch value {
	case JourneyKindMigration, JourneyKindConsolidation, JourneyKindCarveOut, JourneyKindMove, JourneyKindMaturity:
		return JourneyKind{value: value}, nil
	default:
		return JourneyKind{}, ErrInvalidJourneyKind
	}
}

func (k JourneyKind) Value() string { return k.value }

func (k JourneyKind) IsMove() bool { return k.value == JourneyKindMove }

func (k JourneyKind) IsMaturity() bool { return k.value == JourneyKindMaturity }

func (k JourneyKind) TrackKinds() []string {
	if k.IsMaturity() {
		return []string{JourneyKindMaturity}
	}
	return append([]string(nil), applicationJourneyKinds...)
}

const unboundedSourceCount = -1

type sourceCardinality struct {
	min         int
	max         int
	requirement string
}

func (c sourceCardinality) admits(n int) bool {
	return n >= c.min && (c.max == unboundedSourceCount || n <= c.max)
}

var sourceCardinalities = map[string]sourceCardinality{
	JourneyKindMigration:     {min: 1, max: unboundedSourceCount, requirement: "migration requires at least 1 from-application"},
	JourneyKindConsolidation: {min: 1, max: unboundedSourceCount, requirement: "consolidation requires at least 1 from-application"},
	JourneyKindCarveOut:      {min: 1, max: 1, requirement: "carve-out requires exactly 1 from-application"},
	JourneyKindMove:          {min: 0, max: unboundedSourceCount, requirement: "move accepts any number of from-applications"},
	JourneyKindMaturity:      {min: 0, max: 0, requirement: "maturity requires 0 from-applications"},
}

func (k JourneyKind) ValidateSourceCount(n int) error {
	cardinality := sourceCardinalities[k.value]
	if cardinality.admits(n) {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrInvalidSourceApplicationCount, cardinality.requirement)
}

func (k JourneyKind) Equals(other domain.ValueObject) bool {
	if o, ok := other.(JourneyKind); ok {
		return k.value == o.value
	}
	return false
}
