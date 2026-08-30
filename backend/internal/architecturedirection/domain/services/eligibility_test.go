package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateSourceEligibility_EligibleWhenNoOtherDirectionSourcesIt(t *testing.T) {
	conflict := EvaluateSourceEligibility("cap-1", "ec-1", []ActiveDirectionSources{
		direction("ec-1", "Customer Identity", "cap-1"),
	})

	assert.Nil(t, conflict, "the target EC's own direction does not conflict")
}

func TestEvaluateSourceEligibility_ConflictWhenAnotherDirectionSourcesSameNode(t *testing.T) {
	conflict := EvaluateSourceEligibility("cap-1", "ec-2", []ActiveDirectionSources{
		direction("ec-1", "Customer Identity", "cap-1"),
	})

	require.NotNil(t, conflict)
	assert.Equal(t, "ec-1", conflict.EnterpriseCapabilityID)
	assert.Equal(t, "Customer Identity", conflict.EnterpriseCapabilityName)
}

func TestEvaluateSourceEligibility_DescendantOfForeignSourceIsEligible(t *testing.T) {
	conflict := EvaluateSourceEligibility("cap-child", "ec-2", []ActiveDirectionSources{
		direction("ec-1", "Customer Identity", "cap-parent"),
	})

	assert.Nil(t, conflict, "only sourcing the exact same node conflicts (R1); subtree overlap carves out (R2)")
}
