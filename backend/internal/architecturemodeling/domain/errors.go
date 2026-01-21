package domain

import (
	"errors"
	"fmt"
)

var (
	ErrRelationshipAlreadyExists = errors.New("origin relationship already exists for this component")
)

type RelationshipExistsError struct {
	ExistingRelationshipID string
	ComponentID            string
	OriginEntityID         string
	OriginEntityName       string
	RelationshipType       string
}

func NewRelationshipExistsError(
	existingRelationshipID string,
	componentID string,
	originEntityID string,
	originEntityName string,
	relationshipType string,
) *RelationshipExistsError {
	return &RelationshipExistsError{
		ExistingRelationshipID: existingRelationshipID,
		ComponentID:            componentID,
		OriginEntityID:         originEntityID,
		OriginEntityName:       originEntityName,
		RelationshipType:       relationshipType,
	}
}

func (e *RelationshipExistsError) Error() string {
	return fmt.Sprintf("a %s relationship already exists for component %s (linked to %s)", e.RelationshipType, e.ComponentID, e.OriginEntityName)
}

func (e *RelationshipExistsError) Unwrap() error {
	return ErrRelationshipAlreadyExists
}
