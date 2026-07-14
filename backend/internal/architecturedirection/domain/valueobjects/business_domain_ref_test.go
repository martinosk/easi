package valueobjects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBusinessDomainRef_Valid(t *testing.T) {
	id := uuid.New().String()
	ref, err := NewBusinessDomainRef(id)
	require.NoError(t, err)
	assert.Equal(t, id, ref.Value())
}

func TestNewBusinessDomainRef_Invalid(t *testing.T) {
	_, err := NewBusinessDomainRef("not-a-uuid")
	assert.Error(t, err)
}

func TestBusinessDomainRef_Equals(t *testing.T) {
	id := uuid.New().String()
	a, _ := NewBusinessDomainRef(id)
	b, _ := NewBusinessDomainRef(id)
	c, _ := NewBusinessDomainRef(uuid.New().String())
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
