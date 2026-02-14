package handlers

import (
	"context"

	"easi/backend/internal/importing/application/commands"
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

	preview := command.Preview()

	parsedData := aggregates.ParsedData{}
	if command.ParseResult != nil {
		for _, cap := range command.ParseResult.Capabilities {
			parsedData.Capabilities = append(parsedData.Capabilities, aggregates.ParsedElement{
				SourceID:    cap.SourceID,
				Name:        cap.Name,
				Description: cap.Description,
				ParentID:    cap.ParentID,
			})
		}
		for _, comp := range command.ParseResult.Components {
			parsedData.Components = append(parsedData.Components, aggregates.ParsedElement{
				SourceID:    comp.SourceID,
				Name:        comp.Name,
				Description: comp.Description,
			})
		}
		for _, rel := range command.ParseResult.Relationships {
			parsedData.Relationships = append(parsedData.Relationships, aggregates.ParsedRelationship{
				SourceID:      rel.SourceID,
				Type:          rel.Type,
				SourceRef:     rel.SourceRef,
				TargetRef:     rel.TargetRef,
				Name:          rel.Name,
				Documentation: rel.Documentation,
			})
		}
	}

	session, err := aggregates.NewImportSession(aggregates.ImportSessionConfig{
		SourceFormat:       sourceFormat,
		BusinessDomainID:   command.BusinessDomainID,
		CapabilityEAOwner:  command.CapabilityEAOwner,
		Preview:            preview,
		ParsedData:         parsedData,
	})
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, session); err != nil {
		return cqrs.EmptyResult(), err
	}

	return cqrs.NewResult(session.ID()), nil
}
