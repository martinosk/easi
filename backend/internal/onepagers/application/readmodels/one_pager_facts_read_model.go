package readmodels

import (
	"time"

	"easi/backend/internal/onepagers/domain/valueobjects"
)

type FactRecord struct {
	FactsID     string
	TenantID    string
	SubjectType string
	SubjectID   string
	FieldID     string
	Value       *valueobjects.ValueEnvelope
	ValueType   string
	DisplayText string
	ModifiedAt  time.Time
	ModifiedBy  string
}

type ClearFactParams struct {
	SubjectType string
	SubjectID   string
	FieldID     string
	ModifiedAt  time.Time
	ModifiedBy  string
}

type SubjectKey struct {
	SubjectType string
	SubjectID   string
}
