package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeletionScope_Contains_RootID(t *testing.T) {
	rootID := mustCapabilityID(t, "550e8400-e29b-41d4-a716-446655440001")
	scope := NewDeletionScope(rootID, nil)

	assert.True(t, scope.Contains(rootID.Value()))
}

func TestDeletionScope_Contains_DescendantID(t *testing.T) {
	rootID := mustCapabilityID(t, "550e8400-e29b-41d4-a716-446655440001")
	childID := mustCapabilityID(t, "550e8400-e29b-41d4-a716-446655440002")
	scope := NewDeletionScope(rootID, []CapabilityID{childID})

	assert.True(t, scope.Contains(childID.Value()))
}

func TestDeletionScope_Contains_ReturnsFalseForUnknownID(t *testing.T) {
	rootID := mustCapabilityID(t, "550e8400-e29b-41d4-a716-446655440001")
	scope := NewDeletionScope(rootID, nil)

	assert.False(t, scope.Contains("550e8400-e29b-41d4-a716-999999999999"))
}

func TestDeletionScope_BottomUp_ReturnsRootLastWhenNoDescendants(t *testing.T) {
	rootID := mustCapabilityID(t, "550e8400-e29b-41d4-a716-446655440001")
	scope := NewDeletionScope(rootID, nil)

	result := scope.BottomUp()
	require.Len(t, result, 1)
	assert.Equal(t, rootID.Value(), result[0].Value())
}

func TestDeletionScope_BottomUp_ReturnsDescendantsBeforeRoot(t *testing.T) {
	rootID := mustCapabilityID(t, "550e8400-e29b-41d4-a716-446655440001")
	childID := mustCapabilityID(t, "550e8400-e29b-41d4-a716-446655440002")
	grandchildID := mustCapabilityID(t, "550e8400-e29b-41d4-a716-446655440003")

	scope := NewDeletionScope(rootID, []CapabilityID{childID, grandchildID})

	result := scope.BottomUp()
	require.Len(t, result, 3)

	rootIdx := indexOf(result, rootID.Value())
	childIdx := indexOf(result, childID.Value())
	grandchildIdx := indexOf(result, grandchildID.Value())

	assert.Greater(t, rootIdx, childIdx)
	assert.Greater(t, rootIdx, grandchildIdx)
}

func TestDeletionScope_HasDescendants_FalseWhenEmpty(t *testing.T) {
	rootID := mustCapabilityID(t, "550e8400-e29b-41d4-a716-446655440001")
	scope := NewDeletionScope(rootID, nil)

	assert.False(t, scope.HasDescendants())
}

func TestDeletionScope_HasDescendants_TrueWhenChildren(t *testing.T) {
	rootID := mustCapabilityID(t, "550e8400-e29b-41d4-a716-446655440001")
	childID := mustCapabilityID(t, "550e8400-e29b-41d4-a716-446655440002")
	scope := NewDeletionScope(rootID, []CapabilityID{childID})

	assert.True(t, scope.HasDescendants())
}

func TestDeletionScope_AllIDs_IncludesRootAndDescendants(t *testing.T) {
	rootID := mustCapabilityID(t, "550e8400-e29b-41d4-a716-446655440001")
	childID := mustCapabilityID(t, "550e8400-e29b-41d4-a716-446655440002")
	scope := NewDeletionScope(rootID, []CapabilityID{childID})

	allIDs := scope.AllIDs()
	require.Len(t, allIDs, 2)
	assert.Contains(t, idsToStrings(allIDs), rootID.Value())
	assert.Contains(t, idsToStrings(allIDs), childID.Value())
}

func mustCapabilityID(t *testing.T, value string) CapabilityID {
	t.Helper()
	id, err := NewCapabilityIDFromString(value)
	require.NoError(t, err)
	return id
}

func indexOf(ids []CapabilityID, value string) int {
	for i, id := range ids {
		if id.Value() == value {
			return i
		}
	}
	return -1
}

func idsToStrings(ids []CapabilityID) []string {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = id.Value()
	}
	return strs
}
