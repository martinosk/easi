package commands

import (
	"easi/backend/internal/importing/application/parsers"
	"easi/backend/internal/importing/domain/valueobjects"
)

type CreateImportSession struct {
	SourceFormat     string
	BusinessDomainID string
	ParseResult      *parsers.ParseResult
}

func (c CreateImportSession) CommandName() string {
	return "CreateImportSession"
}

func (c *CreateImportSession) Preview() valueobjects.ImportPreview {
	if c.ParseResult == nil {
		return valueobjects.NewImportPreview(valueobjects.SupportedCounts{}, valueobjects.UnsupportedCounts{})
	}
	return c.ParseResult.GetPreview()
}
