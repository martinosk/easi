package queries

import (
	"context"
	"fmt"

	"easi/backend/internal/onepagers/application/ports"
)

func (q *OnePagerQuery) applyMaturitySections(ctx context.Context, fields []Field) error {
	if !anyMaturityValue(fields) {
		return nil
	}
	sections, err := q.deps.MaturityScale.Sections(ctx)
	if err != nil {
		return fmt.Errorf("get maturity scale sections: %w", err)
	}
	for i := range fields {
		setMaturitySection(fields[i].BuiltIn, sections)
	}
	return nil
}

func anyMaturityValue(fields []Field) bool {
	for _, field := range fields {
		if maturityValue(field.BuiltIn) != nil {
			return true
		}
	}
	return false
}

func setMaturitySection(builtIn *BuiltInField, sections []ports.MaturitySection) {
	maturity := maturityValue(builtIn)
	if maturity == nil {
		return
	}
	builtIn.MaturitySection = sectionNameFor(sections, maturity.Value)
}

func maturityValue(builtIn *BuiltInField) *ports.MaturityValue {
	if builtIn == nil {
		return nil
	}
	maturity, ok := builtIn.Value.(ports.MaturityValue)
	if !ok {
		return nil
	}
	return &maturity
}

func sectionNameFor(sections []ports.MaturitySection, value int) string {
	for _, section := range sections {
		if section.MinValue <= value && value <= section.MaxValue {
			return section.Name
		}
	}
	return ""
}
