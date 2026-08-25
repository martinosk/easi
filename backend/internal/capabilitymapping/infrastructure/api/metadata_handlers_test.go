package api

import (
	"testing"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"

	"github.com/stretchr/testify/assert"
)

func TestIsValidationError_EAOwnerErrors(t *testing.T) {
	assert.True(t, isValidationError(valueobjects.ErrEAOwnerNotUser))
	assert.True(t, isValidationError(valueobjects.ErrEAOwnerAmbiguous))
}
