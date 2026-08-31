package projectors_test

import (
	"context"
	"encoding/json"
	"testing"

	metaPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/onepagers/application/projectors"
	"easi/backend/internal/onepagers/application/readmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMaturityScaleStore struct {
	saved [][]readmodels.MaturityScaleSection
}

func (f *fakeMaturityScaleStore) Save(_ context.Context, sections []readmodels.MaturityScaleSection) error {
	f.saved = append(f.saved, sections)
	return nil
}

func projectMaturityScale(t *testing.T, store *fakeMaturityScaleStore, eventType string, payload map[string]any) {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	require.NoError(t, projectors.NewMaturityScaleProjector(store).ProjectEvent(context.Background(), eventType, data))
}

func TestMaturityScaleProjector_CachesSectionsFromEveryScaleEvent(t *testing.T) {
	sections := []map[string]any{
		{"order": 1, "name": "Exploring", "minValue": 0, "maxValue": 39},
		{"order": 2, "name": "Optimizing", "minValue": 40, "maxValue": 100},
	}
	want := []readmodels.MaturityScaleSection{
		{Name: "Exploring", MinValue: 0, MaxValue: 39},
		{Name: "Optimizing", MinValue: 40, MaxValue: 100},
	}

	cases := map[string]map[string]any{
		metaPL.MetaModelConfigurationCreated: {"id": "cfg-1", "sections": sections},
		metaPL.MaturityScaleConfigUpdated:    {"id": "cfg-1", "newSections": sections},
		metaPL.MaturityScaleConfigReset:      {"id": "cfg-1", "sections": sections},
	}

	for eventType, payload := range cases {
		t.Run(eventType, func(t *testing.T) {
			store := &fakeMaturityScaleStore{}

			projectMaturityScale(t, store, eventType, payload)

			require.Len(t, store.saved, 1)
			assert.Equal(t, want, store.saved[0])
		})
	}
}

func TestMaturityScaleProjector_UnknownEvent_NoOp(t *testing.T) {
	store := &fakeMaturityScaleStore{}

	projectMaturityScale(t, store, "SomethingElse", map[string]any{"sections": []map[string]any{{"name": "X"}}})

	assert.Empty(t, store.saved)
}

func TestMaturityScaleEventTypes_CoverEveryScaleEvent(t *testing.T) {
	assert.ElementsMatch(t, []string{
		metaPL.MetaModelConfigurationCreated,
		metaPL.MaturityScaleConfigUpdated,
		metaPL.MaturityScaleConfigReset,
	}, projectors.MaturityScaleEventTypes())
}
