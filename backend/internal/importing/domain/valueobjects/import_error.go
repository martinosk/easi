package valueobjects

import (
	"easi/backend/internal/shared/domain"
)

type ImportError struct {
	sourceElement string
	sourceName    string
	errorMessage  string
	action        string
}

func NewImportError(sourceElement, sourceName, errorMessage, action string) ImportError {
	return ImportError{
		sourceElement: sourceElement,
		sourceName:    sourceName,
		errorMessage:  errorMessage,
		action:        action,
	}
}

func (ie ImportError) SourceElement() string {
	return ie.sourceElement
}

func (ie ImportError) SourceName() string {
	return ie.sourceName
}

func (ie ImportError) Error() string {
	return ie.errorMessage
}

func (ie ImportError) Action() string {
	return ie.action
}

func (ie ImportError) Equals(other domain.ValueObject) bool {
	if otherIE, ok := other.(ImportError); ok {
		return ie.sourceElement == otherIE.sourceElement &&
			ie.sourceName == otherIE.sourceName &&
			ie.errorMessage == otherIE.errorMessage &&
			ie.action == otherIE.action
	}
	return false
}

func (ie ImportError) String() string {
	return ie.errorMessage
}
