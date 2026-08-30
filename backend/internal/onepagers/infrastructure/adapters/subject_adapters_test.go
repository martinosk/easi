package adapters_test

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/infrastructure/adapters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubjectExistenceChecker_ReadsTheSubjectIndex(t *testing.T) {
	subjects := &fakeAttributeStore{exists: map[string]bool{"app-1": true}}
	checker := adapters.NewSubjectExistenceChecker(subjects)

	found, err := checker.SubjectExists(context.Background(), "application", "app-1")

	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, readmodels.SubjectKey{SubjectType: "application", SubjectID: "app-1"}, subjects.gotSubject)
}

func TestSubjectExistenceChecker_UnknownSubjectIsNotFound(t *testing.T) {
	checker := adapters.NewSubjectExistenceChecker(&fakeAttributeStore{exists: map[string]bool{}})

	found, err := checker.SubjectExists(context.Background(), "capability", "missing")

	require.NoError(t, err)
	assert.False(t, found)
}

func TestSubjectExistenceChecker_UnknownSubjectTypeIsRejected(t *testing.T) {
	checker := adapters.NewSubjectExistenceChecker(&fakeAttributeStore{})

	_, err := checker.SubjectExists(context.Background(), "widget", "w-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "widget")
}

func TestSubjectExistenceChecker_WrapsStoreErrors(t *testing.T) {
	wantErr := errors.New("boom")
	checker := adapters.NewSubjectExistenceChecker(&fakeAttributeStore{err: wantErr, exists: map[string]bool{}})

	_, err := checker.SubjectExists(context.Background(), "vendor", "v-1")

	assert.ErrorIs(t, err, wantErr)
}

type fakeMaturityScaleCache struct {
	sections []readmodels.MaturityScaleSection
	err      error
}

func (f fakeMaturityScaleCache) Sections(_ context.Context) ([]readmodels.MaturityScaleSection, error) {
	return f.sections, f.err
}

func TestMaturityScaleSource_MapsEverySection(t *testing.T) {
	cache := fakeMaturityScaleCache{sections: []readmodels.MaturityScaleSection{
		{Name: "Exploring", MinValue: 0, MaxValue: 39},
		{Name: "Optimizing", MinValue: 40, MaxValue: 100},
	}}

	sections, err := adapters.NewMaturityScaleSource(cache).Sections(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []ports.MaturitySection{
		{Name: "Exploring", MinValue: 0, MaxValue: 39},
		{Name: "Optimizing", MinValue: 40, MaxValue: 100},
	}, sections)
}

func TestMaturityScaleSource_UnconfiguredScaleHasNoSections(t *testing.T) {
	sections, err := adapters.NewMaturityScaleSource(fakeMaturityScaleCache{}).Sections(context.Background())

	require.NoError(t, err)
	assert.Nil(t, sections)
}

func TestMaturityScaleSource_WrapsCacheErrors(t *testing.T) {
	wantErr := errors.New("boom")

	_, err := adapters.NewMaturityScaleSource(fakeMaturityScaleCache{err: wantErr}).Sections(context.Background())

	assert.ErrorIs(t, err, wantErr)
}
