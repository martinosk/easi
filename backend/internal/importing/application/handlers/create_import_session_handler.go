package handlers

import (
	"context"

	"easi/backend/internal/importing/application/commands"
	"easi/backend/internal/importing/application/parsers"
	"easi/backend/internal/importing/domain/aggregates"
	"easi/backend/internal/importing/domain/valueobjects"
	"easi/backend/internal/importing/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"
)

type CreateImportSessionHandler struct {
	repository *repositories.ImportSessionRepository
}

func NewCreateImportSessionHandler(repository *repositories.ImportSessionRepository) *CreateImportSessionHandler {
	return &CreateImportSessionHandler{repository: repository}
}

func (h *CreateImportSessionHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.CreateImportSession)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	sourceFormat, err := valueobjects.NewSourceFormat(command.SourceFormat)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	session, err := aggregates.NewImportSession(aggregates.ImportSessionConfig{
		SourceFormat:      sourceFormat,
		BusinessDomainID:  command.BusinessDomainID,
		CapabilityEAOwner: command.CapabilityEAOwner,
		Preview:           command.Preview(),
		ParsedData:        toParsedData(command.ParseResult),
	})
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, session); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(session.ID()), nil
}

func toParsedData(result *parsers.ParseResult) aggregates.ParsedData {
	if result == nil {
		return aggregates.ParsedData{}
	}
	return aggregates.ParsedData{
		Capabilities:  toElements(result.Capabilities),
		Components:    toElements(result.Components),
		ValueStreams:  toElements(result.ValueStreams),
		Relationships: toRelationships(result.Relationships),
	}
}

func toElements(src []parsers.ParsedElement) []aggregates.ParsedElement {
	elements := make([]aggregates.ParsedElement, len(src))
	for i, e := range src {
		elements[i] = aggregates.ParsedElement{
			SourceID:    e.SourceID,
			Name:        e.Name,
			Description: e.Description,
			ParentID:    e.ParentID,
		}
	}
	return elements
}

func toRelationships(src []parsers.ParsedRelationship) []aggregates.ParsedRelationship {
	rels := make([]aggregates.ParsedRelationship, len(src))
	for i, r := range src {
		rels[i] = aggregates.ParsedRelationship{
			SourceID:      r.SourceID,
			Type:          r.Type,
			SourceRef:     r.SourceRef,
			TargetRef:     r.TargetRef,
			Name:          r.Name,
			Documentation: r.Documentation,
		}
	}
	return rels
}
