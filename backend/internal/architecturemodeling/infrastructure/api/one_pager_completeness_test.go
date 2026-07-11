package api

import (
	"context"
	"encoding/json"
	"testing"

	"easi/backend/internal/architecturemodeling/application/readmodels"

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

func TestDecorateComponentsOnePagerCompleteness_SetsIndicatorPerRow(t *testing.T) {
	source := &fakeOnePagerCompletenessSource{indicators: map[string]bool{"comp-1": true}, present: true}
	rows := []readmodels.ApplicationComponentDTO{{ID: "comp-1"}, {ID: "comp-2"}}

	err := decorateComponentsOnePagerCompleteness(context.Background(), source, rows)

	require.NoError(t, err)
	assert.Equal(t, []string{"comp-1", "comp-2"}, source.gotIDs)
	require.NotNil(t, rows[0].OnePagerComplete)
	assert.True(t, *rows[0].OnePagerComplete)
	require.NotNil(t, rows[1].OnePagerComplete)
	assert.False(t, *rows[1].OnePagerComplete)
}

func TestDecorateComponentsOnePagerCompleteness_NilSource_LeavesRowsUntouched(t *testing.T) {
	rows := []readmodels.ApplicationComponentDTO{{ID: "comp-1"}}

	err := decorateComponentsOnePagerCompleteness(context.Background(), nil, rows)

	require.NoError(t, err)
	assert.Nil(t, rows[0].OnePagerComplete)
}

func TestDecorateComponentsOnePagerCompleteness_IndicatorAbsent_LeavesRowsUntouched(t *testing.T) {
	source := &fakeOnePagerCompletenessSource{present: false}
	rows := []readmodels.ApplicationComponentDTO{{ID: "comp-1"}}

	err := decorateComponentsOnePagerCompleteness(context.Background(), source, rows)

	require.NoError(t, err)
	assert.Equal(t, 1, source.calls)
	assert.Nil(t, rows[0].OnePagerComplete)
}

func TestDecorateComponentsOnePagerCompleteness_EmptyRows_SkipsSource(t *testing.T) {
	source := &fakeOnePagerCompletenessSource{present: true}

	err := decorateComponentsOnePagerCompleteness(context.Background(), source, nil)

	require.NoError(t, err)
	assert.Equal(t, 0, source.calls)
}

func TestDecorateComponentsOnePagerCompleteness_SourceError_IsPropagated(t *testing.T) {
	source := &fakeOnePagerCompletenessSource{err: assert.AnError}
	rows := []readmodels.ApplicationComponentDTO{{ID: "comp-1"}}

	err := decorateComponentsOnePagerCompleteness(context.Background(), source, rows)

	require.Error(t, err)
	assert.Nil(t, rows[0].OnePagerComplete)
}

func TestDecorateOriginEntitiesOnePagerCompleteness_SetsIndicatorPerRow(t *testing.T) {
	source := &fakeOnePagerCompletenessSource{indicators: map[string]bool{"row-1": true}, present: true}

	acquired := []readmodels.AcquiredEntityDTO{{ID: "row-1"}, {ID: "row-2"}}
	require.NoError(t, decorateAcquiredEntitiesOnePagerCompleteness(context.Background(), source, acquired))
	require.NotNil(t, acquired[0].OnePagerComplete)
	assert.True(t, *acquired[0].OnePagerComplete)
	require.NotNil(t, acquired[1].OnePagerComplete)
	assert.False(t, *acquired[1].OnePagerComplete)

	vendors := []readmodels.VendorDTO{{ID: "row-1"}, {ID: "row-2"}}
	require.NoError(t, decorateVendorsOnePagerCompleteness(context.Background(), source, vendors))
	require.NotNil(t, vendors[0].OnePagerComplete)
	assert.True(t, *vendors[0].OnePagerComplete)
	require.NotNil(t, vendors[1].OnePagerComplete)
	assert.False(t, *vendors[1].OnePagerComplete)

	teams := []readmodels.InternalTeamDTO{{ID: "row-1"}, {ID: "row-2"}}
	require.NoError(t, decorateInternalTeamsOnePagerCompleteness(context.Background(), source, teams))
	require.NotNil(t, teams[0].OnePagerComplete)
	assert.True(t, *teams[0].OnePagerComplete)
	require.NotNil(t, teams[1].OnePagerComplete)
	assert.False(t, *teams[1].OnePagerComplete)
}

func TestComponentDTO_OnePagerCompleteJSONContract(t *testing.T) {
	withoutIndicator, err := json.Marshal(readmodels.ApplicationComponentDTO{ID: "comp-1", Name: "Billing"})
	require.NoError(t, err)
	assert.NotContains(t, string(withoutIndicator), "onePagerComplete")

	complete := false
	withIndicator, err := json.Marshal(readmodels.ApplicationComponentDTO{ID: "comp-1", Name: "Billing", OnePagerComplete: &complete})
	require.NoError(t, err)
	assert.Contains(t, string(withIndicator), `"onePagerComplete":false`)
}
