package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOwnershipModel_Valid(t *testing.T) {
	model, err := NewOwnershipModel("TeamOwned")
	assert.NoError(t, err)
	assert.Equal(t, OwnershipTeamOwned, model)
}

func TestNewOwnershipModel_Empty(t *testing.T) {
	model, err := NewOwnershipModel("")
	assert.NoError(t, err)
	assert.Equal(t, OwnershipModel(""), model)
	assert.True(t, model.IsEmpty())
}

func TestNewOwnershipModel_InvalidValue(t *testing.T) {
	_, err := NewOwnershipModel("InvalidModel")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidOwnershipModel, err)
}

func TestOwnershipModel_Equals(t *testing.T) {
	model1 := OwnershipTribeOwned
	model2 := OwnershipTribeOwned
	model3 := OwnershipTeamOwned

	assert.True(t, model1.Equals(model2))
	assert.False(t, model1.Equals(model3))
}

func TestOwnershipModel_IsEmpty(t *testing.T) {
	emptyModel := OwnershipModel("")
	nonEmptyModel := OwnershipShared

	assert.True(t, emptyModel.IsEmpty())
	assert.False(t, nonEmptyModel.IsEmpty())
}
