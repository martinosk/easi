package commands

import "easi/backend/internal/onepagers/domain/valueobjects"

type FactsSubjectField struct {
	TenantID    string
	SubjectType string
	SubjectID   string
	FieldID     string
	ModifiedBy  string
}

type RecordFieldValue struct {
	FactsSubjectField
	Value valueobjects.ValueEnvelope
}

func (c RecordFieldValue) CommandName() string { return "RecordFieldValue" }

type ClearFieldValue struct {
	FactsSubjectField
}

func (c ClearFieldValue) CommandName() string { return "ClearFieldValue" }

type ArchiveOnePagerFacts struct {
	FactsID string
	Reason  string
}

func (c ArchiveOnePagerFacts) CommandName() string { return "ArchiveOnePagerFacts" }
