package repository

import (
	"fmt"
)

type DeserializationError struct {
	AggregateID    string
	EventType      string
	SequenceNumber int
	FieldName      string
	Cause          error
}

func (e *DeserializationError) Error() string {
	if e.FieldName != "" {
		return fmt.Sprintf("deserialization failed for aggregate %s, event %s at position %d, field %s: %v",
			e.AggregateID, e.EventType, e.SequenceNumber, e.FieldName, e.Cause)
	}
	return fmt.Sprintf("deserialization failed for aggregate %s, event %s at position %d: %v",
		e.AggregateID, e.EventType, e.SequenceNumber, e.Cause)
}

func (e *DeserializationError) Unwrap() error {
	return e.Cause
}

func NewDeserializationError(aggregateID, eventType string, sequenceNumber int, cause error) *DeserializationError {
	return &DeserializationError{
		AggregateID:    aggregateID,
		EventType:      eventType,
		SequenceNumber: sequenceNumber,
		Cause:          cause,
	}
}

func NewFieldDeserializationError(aggregateID, eventType string, sequenceNumber int, fieldName string, cause error) *DeserializationError {
	return &DeserializationError{
		AggregateID:    aggregateID,
		EventType:      eventType,
		SequenceNumber: sequenceNumber,
		FieldName:      fieldName,
		Cause:          cause,
	}
}

type UnknownEventTypeError struct {
	AggregateID    string
	EventType      string
	SequenceNumber int
}

func (e *UnknownEventTypeError) Error() string {
	return fmt.Sprintf(
		"unknown event type %s for aggregate %s at position %d: no deserializer is registered, "+
			"so the aggregate cannot be rebuilt from its history",
		e.EventType, e.AggregateID, e.SequenceNumber)
}

func NewUnknownEventTypeError(aggregateID, eventType string, sequenceNumber int) *UnknownEventTypeError {
	return &UnknownEventTypeError{
		AggregateID:    aggregateID,
		EventType:      eventType,
		SequenceNumber: sequenceNumber,
	}
}

type FieldError struct {
	FieldName    string
	ExpectedType string
	ActualType   string
	Message      string
}

func (e *FieldError) Error() string {
	if e.ExpectedType != "" && e.ActualType != "" {
		return fmt.Sprintf("field %s: expected type %s, got %s", e.FieldName, e.ExpectedType, e.ActualType)
	}
	return fmt.Sprintf("field %s: %s", e.FieldName, e.Message)
}

func NewMissingFieldError(fieldName string) *FieldError {
	return &FieldError{
		FieldName: fieldName,
		Message:   "required field is missing",
	}
}

func NewTypeError(fieldName, expectedType, actualType string) *FieldError {
	return &FieldError{
		FieldName:    fieldName,
		ExpectedType: expectedType,
		ActualType:   actualType,
	}
}
