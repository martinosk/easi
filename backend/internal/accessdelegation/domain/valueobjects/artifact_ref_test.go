package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewArtifactType_ValidTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected ArtifactType
	}{
		{"capability", ArtifactTypeCapability},
		{"component", ArtifactTypeComponent},
		{"view", ArtifactTypeView},
		{"domain", ArtifactTypeDomain},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			at, err := NewArtifactType(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, at)
		})
	}
}

func TestNewArtifactType_InvalidType(t *testing.T) {
	invalidTypes := []string{"", "invalid", "Capability", "COMPONENT", "unknown", " "}

	for _, input := range invalidTypes {
		t.Run(input, func(t *testing.T) {
			_, err := NewArtifactType(input)
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidArtifactType, err)
		})
	}
}

func TestArtifactType_String(t *testing.T) {
	assert.Equal(t, "capability", ArtifactTypeCapability.String())
	assert.Equal(t, "component", ArtifactTypeComponent.String())
	assert.Equal(t, "view", ArtifactTypeView.String())
	assert.Equal(t, "domain", ArtifactTypeDomain.String())
}

func TestNewArtifactRef_Valid(t *testing.T) {
	ref, err := NewArtifactRef(ArtifactTypeCapability, "550e8400-e29b-41d4-a716-446655440000")
	require.NoError(t, err)
	assert.Equal(t, ArtifactTypeCapability, ref.Type())
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", ref.ID())
}

func TestNewArtifactRef_EmptyID(t *testing.T) {
	_, err := NewArtifactRef(ArtifactTypeCapability, "")
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyArtifactID, err)
}

func TestNewArtifactRef_InvalidUUID(t *testing.T) {
	_, err := NewArtifactRef(ArtifactTypeCapability, "not-a-uuid")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidArtifactID, err)
}

func TestArtifactRef_Equals_SameValues(t *testing.T) {
	ref1, _ := NewArtifactRef(ArtifactTypeCapability, "550e8400-e29b-41d4-a716-446655440000")
	ref2, _ := NewArtifactRef(ArtifactTypeCapability, "550e8400-e29b-41d4-a716-446655440000")

	assert.True(t, ref1.Equals(ref2))
}

func TestArtifactRef_Equals_DifferentType(t *testing.T) {
	ref1, _ := NewArtifactRef(ArtifactTypeCapability, "550e8400-e29b-41d4-a716-446655440000")
	ref2, _ := NewArtifactRef(ArtifactTypeComponent, "550e8400-e29b-41d4-a716-446655440000")

	assert.False(t, ref1.Equals(ref2))
}

func TestArtifactRef_Equals_DifferentID(t *testing.T) {
	ref1, _ := NewArtifactRef(ArtifactTypeCapability, "550e8400-e29b-41d4-a716-446655440001")
	ref2, _ := NewArtifactRef(ArtifactTypeCapability, "550e8400-e29b-41d4-a716-446655440002")

	assert.False(t, ref1.Equals(ref2))
}

func TestArtifactRef_Equals_DifferentValueObjectType(t *testing.T) {
	ref, _ := NewArtifactRef(ArtifactTypeCapability, "550e8400-e29b-41d4-a716-446655440000")

	assert.False(t, ref.Equals(GrantScopeWrite))
}
