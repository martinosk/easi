package api

import (
	"context"
	"encoding/json"
	"testing"

	"easi/backend/internal/capabilitymapping/application/readmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOnePagerCompletenessSource struct {
	calls      int
	gotIDs     []string
	indicators map[string]bool
	present    bool
	err        error
}

func (f *fakeOnePagerCompletenessSource) CompletenessFor(_ context.Context, subjectIDs []string) (map[string]bool, bool, error) {
	f.calls++
	f.gotIDs = subjectIDs
	return f.indicators, f.present, f.err
}

func TestDecorateCapabilitiesOnePagerCompleteness_SetsIndicatorPerRow(t *testing.T) {
	source := &fakeOnePagerCompletenessSource{indicators: map[string]bool{"cap-1": true}, present: true}
	rows := []readmodels.CapabilityDTO{{ID: "cap-1"}, {ID: "cap-2"}}

	err := decorateCapabilitiesOnePagerCompleteness(context.Background(), source, rows)

	require.NoError(t, err)
	assert.Equal(t, []string{"cap-1", "cap-2"}, source.gotIDs)
	require.NotNil(t, rows[0].OnePagerComplete)
	assert.True(t, *rows[0].OnePagerComplete)
	require.NotNil(t, rows[1].OnePagerComplete)
	assert.False(t, *rows[1].OnePagerComplete)
}

func TestDecorateCapabilitiesOnePagerCompleteness_NilSource_LeavesRowsUntouched(t *testing.T) {
	rows := []readmodels.CapabilityDTO{{ID: "cap-1"}}

	err := decorateCapabilitiesOnePagerCompleteness(context.Background(), nil, rows)

	require.NoError(t, err)
	assert.Nil(t, rows[0].OnePagerComplete)
}

func TestDecorateCapabilitiesOnePagerCompleteness_IndicatorAbsent_LeavesRowsUntouched(t *testing.T) {
	source := &fakeOnePagerCompletenessSource{present: false}
	rows := []readmodels.CapabilityDTO{{ID: "cap-1"}}

	err := decorateCapabilitiesOnePagerCompleteness(context.Background(), source, rows)

	require.NoError(t, err)
	assert.Equal(t, 1, source.calls)
	assert.Nil(t, rows[0].OnePagerComplete)
}

func TestDecorateCapabilitiesOnePagerCompleteness_EmptyRows_SkipsSource(t *testing.T) {
	source := &fakeOnePagerCompletenessSource{present: true}

	err := decorateCapabilitiesOnePagerCompleteness(context.Background(), source, nil)

	require.NoError(t, err)
	assert.Equal(t, 0, source.calls)
}

func TestDecorateCapabilitiesOnePagerCompleteness_SourceError_IsPropagated(t *testing.T) {
	source := &fakeOnePagerCompletenessSource{err: assert.AnError}
	rows := []readmodels.CapabilityDTO{{ID: "cap-1"}}

	err := decorateCapabilitiesOnePagerCompleteness(context.Background(), source, rows)

	require.Error(t, err)
	assert.Nil(t, rows[0].OnePagerComplete)
}

func TestCapabilityDTO_OnePagerCompleteJSONContract(t *testing.T) {
	withoutIndicator, err := json.Marshal(readmodels.CapabilityDTO{ID: "cap-1", Name: "Order Management"})
	require.NoError(t, err)
	assert.NotContains(t, string(withoutIndicator), "onePagerComplete")

	complete := true
	withIndicator, err := json.Marshal(readmodels.CapabilityDTO{ID: "cap-1", Name: "Order Management", OnePagerComplete: &complete})
	require.NoError(t, err)
	assert.Contains(t, string(withIndicator), `"onePagerComplete":true`)
}
