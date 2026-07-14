package api

import (
	"context"
	"errors"
	"testing"
	"time"

	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"
	eaReadModels "easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/infrastructure/database"
	metaReadModels "easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/domain/catalog"
	"easi/backend/internal/onepagers/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type widget struct {
	id     string
	fields map[string]ports.BuiltInFieldValue
}

func widgetSnapshot(w *widget) *ports.SubjectSnapshot {
	if w == nil {
		return nil
	}
	return &ports.SubjectSnapshot{Name: w.id, Fields: w.fields}
}

func widgetSourceConfig() builtInSourceConfig[widget] {
	return builtInSourceConfig[widget]{
		subjectType:   "widget",
		getByID:       func(context.Context, string) (*widget, error) { return nil, nil },
		toSnapshot:    widgetSnapshot,
		idOf:          func(w *widget) string { return w.id },
		countSubjects: func(context.Context) (int, error) { return 0, nil },
	}
}

func TestBuiltInFieldSource_FetchSubject_ResolvesOnlyIncludedRelations(t *testing.T) {
	resolved := map[string]int{}
	cfg := widgetSourceConfig()
	cfg.getByID = func(_ context.Context, id string) (*widget, error) {
		return &widget{id: id, fields: map[string]ports.BuiltInFieldValue{"name": ports.TextValue{Text: id}}}, nil
	}
	cfg.relations = []relationBinding[widget]{
		{entryID: "rel-a", resolve: func(_ context.Context, _ *widget) (ports.ReferenceListValue, error) {
			resolved["rel-a"]++
			return ports.ReferenceListValue{References: []ports.Reference{{ID: "x", Label: "X", SubjectType: "application"}}}, nil
		}},
		{entryID: "rel-b", resolve: func(_ context.Context, _ *widget) (ports.ReferenceListValue, error) {
			resolved["rel-b"]++
			return ports.ReferenceListValue{}, nil
		}},
	}
	source := builtInFieldSource(cfg)

	snapshot, err := source.FetchSubject(context.Background(), "w-1", []string{"name", "rel-a"})

	require.NoError(t, err)
	assert.Equal(t, 1, resolved["rel-a"], "included relation is resolved")
	assert.Equal(t, 0, resolved["rel-b"], "excluded relation is never resolved")
	assert.Equal(t, ports.ReferenceListValue{References: []ports.Reference{{ID: "x", Label: "X", SubjectType: "application"}}}, snapshot.Fields["rel-a"])
	_, hasB := snapshot.Fields["rel-b"]
	assert.False(t, hasB)
}

func TestBuiltInFieldSource_FetchSubject_WrapsRelationResolveError(t *testing.T) {
	wantErr := errors.New("boom")
	cfg := widgetSourceConfig()
	cfg.getByID = func(_ context.Context, id string) (*widget, error) {
		return &widget{id: id, fields: map[string]ports.BuiltInFieldValue{}}, nil
	}
	cfg.relations = []relationBinding[widget]{
		{entryID: "rel-a", resolve: func(_ context.Context, _ *widget) (ports.ReferenceListValue, error) {
			return ports.ReferenceListValue{}, wantErr
		}},
	}
	source := builtInFieldSource(cfg)

	_, err := source.FetchSubject(context.Background(), "w-1", []string{"rel-a"})

	assert.ErrorIs(t, err, wantErr)
}

func TestBuiltInFieldSource_FetchSubject_MissingSubjectSkipsRelations(t *testing.T) {
	cfg := widgetSourceConfig()
	cfg.getByID = func(context.Context, string) (*widget, error) { return nil, nil }
	called := false
	cfg.relations = []relationBinding[widget]{
		{entryID: "rel-a", resolve: func(_ context.Context, _ *widget) (ports.ReferenceListValue, error) {
			called = true
			return ports.ReferenceListValue{}, nil
		}},
	}
	source := builtInFieldSource(cfg)

	snapshot, err := source.FetchSubject(context.Background(), "missing", []string{"rel-a"})

	require.NoError(t, err)
	assert.Nil(t, snapshot)
	assert.False(t, called)
}

func TestBuiltInFieldSource_CountSubjects(t *testing.T) {
	t.Run("delegates to the count function", func(t *testing.T) {
		cfg := widgetSourceConfig()
		cfg.countSubjects = func(context.Context) (int, error) { return 42, nil }
		source := builtInFieldSource(cfg)

		count, err := source.CountSubjects(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 42, count)
	})

	t.Run("wraps the count function error", func(t *testing.T) {
		wantErr := errors.New("boom")
		cfg := widgetSourceConfig()
		cfg.countSubjects = func(context.Context) (int, error) { return 0, wantErr }
		source := builtInFieldSource(cfg)

		_, err := source.CountSubjects(context.Background())

		assert.ErrorIs(t, err, wantErr)
	})
}

func TestBuiltInFieldSource_FilledBuiltInFields(t *testing.T) {
	widgets := []widget{
		{id: "w-1", fields: map[string]ports.BuiltInFieldValue{"description": ports.TextValue{Text: "has value"}, "experts": ports.ExpertsValue{Experts: nil}}},
		{id: "w-2", fields: map[string]ports.BuiltInFieldValue{"experts": ports.ExpertsValue{Experts: []ports.Expert{{Name: "Alice"}}}}},
	}
	cfg := widgetSourceConfig()
	var gotIDs []string
	cfg.getByIDs = func(_ context.Context, ids []string) ([]widget, error) {
		gotIDs = ids
		return widgets, nil
	}
	source := builtInFieldSource(cfg)

	result, err := source.FilledBuiltInFields(context.Background(), []string{"w-1", "w-2"}, []string{"description", "experts"})

	require.NoError(t, err)
	assert.Equal(t, []string{"w-1", "w-2"}, gotIDs)
	assert.Equal(t, map[string]map[string]bool{
		"w-1": {"description": true, "experts": false},
		"w-2": {"description": false, "experts": true},
	}, result)
}

func TestBuiltInFieldSource_FilledBuiltInFields_EmptyInputsSkipFetch(t *testing.T) {
	cfg := widgetSourceConfig()
	called := false
	cfg.getByIDs = func(context.Context, []string) ([]widget, error) {
		called = true
		return nil, nil
	}
	source := builtInFieldSource(cfg)

	result, err := source.FilledBuiltInFields(context.Background(), nil, []string{"experts"})

	require.NoError(t, err)
	assert.Empty(t, result)
	assert.False(t, called, "no fetch when there are no subject IDs")
}

func TestBuiltInFieldSource_FilledBuiltInFields_WrapsError(t *testing.T) {
	wantErr := errors.New("boom")
	cfg := widgetSourceConfig()
	cfg.getByIDs = func(context.Context, []string) ([]widget, error) { return nil, wantErr }
	source := builtInFieldSource(cfg)

	_, err := source.FilledBuiltInFields(context.Background(), []string{"w-1"}, []string{"experts"})

	assert.ErrorIs(t, err, wantErr)
}

func TestBuiltInFieldSource_CountSubjectsWithBuiltInValue(t *testing.T) {
	widgets := []widget{
		{id: "w-1", fields: map[string]ports.BuiltInFieldValue{"experts": ports.ExpertsValue{Experts: []ports.Expert{{Name: "Alice"}}}}},
		{id: "w-2", fields: map[string]ports.BuiltInFieldValue{"experts": ports.ExpertsValue{Experts: nil}}},
		{id: "w-3", fields: map[string]ports.BuiltInFieldValue{"experts": ports.ExpertsValue{Experts: []ports.Expert{{Name: "Bob"}}}}},
	}
	cfg := widgetSourceConfig()
	cfg.getAll = func(context.Context) ([]widget, error) { return widgets, nil }
	source := builtInFieldSource(cfg)

	count, err := source.CountSubjectsWithBuiltInValue(context.Background(), "experts")

	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestBuiltInFieldSource_CountSubjectsWithBuiltInValue_WrapsError(t *testing.T) {
	wantErr := errors.New("boom")
	cfg := widgetSourceConfig()
	cfg.getAll = func(context.Context) ([]widget, error) { return nil, wantErr }
	source := builtInFieldSource(cfg)

	_, err := source.CountSubjectsWithBuiltInValue(context.Background(), "experts")

	assert.ErrorIs(t, err, wantErr)
}

func catalogFieldIDs(t *testing.T, subjectType string) []string {
	t.Helper()
	st, err := valueobjects.NewSubjectType(subjectType)
	require.NoError(t, err)

	entries := catalog.DefaultEntriesFor(st)
	ids := make([]string, len(entries))
	for i, entry := range entries {
		ids[i] = entry.ID
	}
	return ids
}

func assertCatalogFieldsPresent(t *testing.T, subjectType string, snapshot *ports.SubjectSnapshot) {
	t.Helper()
	for _, id := range catalogFieldIDs(t, subjectType) {
		_, ok := snapshot.Fields[id]
		assert.Truef(t, ok, "catalog entry %q missing from snapshot fields", id)
	}
}

func assertTextField(t *testing.T, snapshot *ports.SubjectSnapshot, key, want string) {
	t.Helper()
	assert.Equal(t, ports.TextValue{Text: want}, snapshot.Fields[key])
}

func assertNilFields(t *testing.T, snapshot *ports.SubjectSnapshot, keys ...string) {
	t.Helper()
	for _, key := range keys {
		assert.Nil(t, snapshot.Fields[key])
	}
}

func TestCapabilitySnapshot(t *testing.T) {
	t.Run("nil dto returns nil", func(t *testing.T) {
		assert.Nil(t, capabilitySnapshot(nil))
	})

	t.Run("fully populated", func(t *testing.T) {
		dto := &capReadModels.CapabilityDTO{
			Name:          "Order Management",
			Description:   "Handles customer orders",
			MaturityValue: 62,
			Experts: []capReadModels.ExpertDTO{
				{Name: "Alice", Role: "Owner", Contact: "alice@example.com"},
			},
		}

		snapshot := capabilitySnapshot(dto)

		require.NotNil(t, snapshot)
		assert.Equal(t, "Order Management", snapshot.Name)
		assertCatalogFieldsPresent(t, "capability", snapshot)
		assertTextField(t, snapshot, "name", "Order Management")
		assertTextField(t, snapshot, "description", "Handles customer orders")
		assert.Equal(t, ports.MaturityValue{Value: 62}, snapshot.Fields["maturity"])
		assert.Equal(t, ports.ExpertsValue{Experts: []ports.Expert{{Name: "Alice", Role: "Owner", Contact: "alice@example.com"}}}, snapshot.Fields["experts"])
	})

	t.Run("empty optional fields", func(t *testing.T) {
		dto := &capReadModels.CapabilityDTO{Name: "Order Management"}

		snapshot := capabilitySnapshot(dto)

		require.NotNil(t, snapshot)
		assert.Equal(t, "Order Management", snapshot.Name)
		assertCatalogFieldsPresent(t, "capability", snapshot)
		assertNilFields(t, snapshot, "description", "experts")
		assert.Equal(t, ports.MaturityValue{Value: 0}, snapshot.Fields["maturity"])
	})
}

func TestEnterpriseCapabilitySnapshot(t *testing.T) {
	t.Run("nil dto returns nil", func(t *testing.T) {
		assert.Nil(t, enterpriseCapabilitySnapshot(nil))
	})

	t.Run("fully populated", func(t *testing.T) {
		dto := &eaReadModels.EnterpriseCapabilityDTO{
			Name:        "Customer Experience",
			Description: "Cross-domain grouping",
			Category:    "Front Office",
		}

		snapshot := enterpriseCapabilitySnapshot(dto)

		require.NotNil(t, snapshot)
		assert.Equal(t, "Customer Experience", snapshot.Name)
		assertCatalogFieldsPresent(t, "enterprise-capability", snapshot)
		assertTextField(t, snapshot, "name", "Customer Experience")
		assertTextField(t, snapshot, "description", "Cross-domain grouping")
		assertTextField(t, snapshot, "category", "Front Office")
	})

	t.Run("empty optional fields", func(t *testing.T) {
		dto := &eaReadModels.EnterpriseCapabilityDTO{Name: "Customer Experience"}

		snapshot := enterpriseCapabilitySnapshot(dto)

		require.NotNil(t, snapshot)
		assert.Equal(t, "Customer Experience", snapshot.Name)
		assertCatalogFieldsPresent(t, "enterprise-capability", snapshot)
		assertNilFields(t, snapshot, "description", "category")
	})
}

func TestApplicationSnapshot(t *testing.T) {
	t.Run("nil dto returns nil", func(t *testing.T) {
		assert.Nil(t, applicationSnapshot(nil))
	})

	t.Run("fully populated", func(t *testing.T) {
		dto := &archReadModels.ApplicationComponentDTO{
			Name:        "Billing Service",
			Description: "Handles invoicing",
			Experts: []archReadModels.ExpertDTO{
				{Name: "Bob", Role: "Tech Lead", Contact: "bob@example.com"},
			},
		}

		snapshot := applicationSnapshot(dto)

		require.NotNil(t, snapshot)
		assert.Equal(t, "Billing Service", snapshot.Name)
		assertCatalogFieldsPresent(t, "application", snapshot)
		assertTextField(t, snapshot, "name", "Billing Service")
		assertTextField(t, snapshot, "description", "Handles invoicing")
		assert.Equal(t, ports.ExpertsValue{Experts: []ports.Expert{{Name: "Bob", Role: "Tech Lead", Contact: "bob@example.com"}}}, snapshot.Fields["experts"])
	})

	t.Run("empty optional fields", func(t *testing.T) {
		dto := &archReadModels.ApplicationComponentDTO{Name: "Billing Service"}

		snapshot := applicationSnapshot(dto)

		require.NotNil(t, snapshot)
		assert.Equal(t, "Billing Service", snapshot.Name)
		assertCatalogFieldsPresent(t, "application", snapshot)
		assertNilFields(t, snapshot, "description", "experts")
	})
}

func TestAcquiredEntitySnapshot(t *testing.T) {
	t.Run("nil dto returns nil", func(t *testing.T) {
		assert.Nil(t, acquiredEntitySnapshot(nil))
	})

	t.Run("fully populated", func(t *testing.T) {
		acquisitionDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		dto := &archReadModels.AcquiredEntityDTO{
			Name:              "Acme Corp",
			AcquisitionDate:   &acquisitionDate,
			IntegrationStatus: "In Progress",
		}

		snapshot := acquiredEntitySnapshot(dto)

		require.NotNil(t, snapshot)
		assert.Equal(t, "Acme Corp", snapshot.Name)
		assertCatalogFieldsPresent(t, "acquired-entity", snapshot)
		assertTextField(t, snapshot, "name", "Acme Corp")
		assert.Equal(t, ports.DateValue{Date: acquisitionDate}, snapshot.Fields["acquisition-date"])
		assertTextField(t, snapshot, "integration-status", "In Progress")
	})

	t.Run("empty optional fields", func(t *testing.T) {
		dto := &archReadModels.AcquiredEntityDTO{Name: "Acme Corp"}

		snapshot := acquiredEntitySnapshot(dto)

		require.NotNil(t, snapshot)
		assert.Equal(t, "Acme Corp", snapshot.Name)
		assertCatalogFieldsPresent(t, "acquired-entity", snapshot)
		assertNilFields(t, snapshot, "acquisition-date", "integration-status")
	})
}

func TestVendorSnapshot(t *testing.T) {
	t.Run("nil dto returns nil", func(t *testing.T) {
		assert.Nil(t, vendorSnapshot(nil))
	})

	t.Run("fully populated", func(t *testing.T) {
		dto := &archReadModels.VendorDTO{
			Name:                  "Contoso",
			ImplementationPartner: "Contoso Consulting",
			Notes:                 "Preferred vendor",
		}

		snapshot := vendorSnapshot(dto)

		require.NotNil(t, snapshot)
		assert.Equal(t, "Contoso", snapshot.Name)
		assertCatalogFieldsPresent(t, "vendor", snapshot)
		assertTextField(t, snapshot, "name", "Contoso")
		assertTextField(t, snapshot, "implementation-partner", "Contoso Consulting")
		assertTextField(t, snapshot, "notes", "Preferred vendor")
	})

	t.Run("empty optional fields", func(t *testing.T) {
		dto := &archReadModels.VendorDTO{Name: "Contoso"}

		snapshot := vendorSnapshot(dto)

		require.NotNil(t, snapshot)
		assert.Equal(t, "Contoso", snapshot.Name)
		assertCatalogFieldsPresent(t, "vendor", snapshot)
		assertNilFields(t, snapshot, "implementation-partner", "notes")
	})
}

func TestInternalTeamSnapshot(t *testing.T) {
	t.Run("nil dto returns nil", func(t *testing.T) {
		assert.Nil(t, internalTeamSnapshot(nil))
	})

	t.Run("fully populated", func(t *testing.T) {
		dto := &archReadModels.InternalTeamDTO{
			Name:          "Platform Team",
			Department:    "Engineering",
			ContactPerson: "Carol",
		}

		snapshot := internalTeamSnapshot(dto)

		require.NotNil(t, snapshot)
		assert.Equal(t, "Platform Team", snapshot.Name)
		assertCatalogFieldsPresent(t, "internal-team", snapshot)
		assertTextField(t, snapshot, "name", "Platform Team")
		assertTextField(t, snapshot, "department", "Engineering")
		assertTextField(t, snapshot, "contact-person", "Carol")
	})

	t.Run("empty optional fields", func(t *testing.T) {
		dto := &archReadModels.InternalTeamDTO{Name: "Platform Team"}

		snapshot := internalTeamSnapshot(dto)

		require.NotNil(t, snapshot)
		assert.Equal(t, "Platform Team", snapshot.Name)
		assertCatalogFieldsPresent(t, "internal-team", snapshot)
		assertNilFields(t, snapshot, "department", "contact-person")
	})
}

func TestMaturitySections(t *testing.T) {
	t.Run("nil config returns nil", func(t *testing.T) {
		assert.Nil(t, maturitySections(nil))
	})

	t.Run("maps every section", func(t *testing.T) {
		config := &metaReadModels.MetaModelConfigurationDTO{
			Sections: []metaReadModels.MaturitySectionDTO{
				{Order: 1, Name: "Exploring", MinValue: 0, MaxValue: 39},
				{Order: 2, Name: "Optimizing", MinValue: 40, MaxValue: 100},
			},
		}

		sections := maturitySections(config)

		assert.Equal(t, []ports.MaturitySection{
			{Name: "Exploring", MinValue: 0, MaxValue: 39},
			{Name: "Optimizing", MinValue: 40, MaxValue: 100},
		}, sections)
	})
}

func TestNewOnePagerBuiltInFieldSources_KeysMatchSubjectTypes(t *testing.T) {
	sources := newOnePagerBuiltInFieldSources(database.NewTenantAwareDB(nil))

	got := make([]string, 0, len(sources))
	for subjectType := range sources {
		got = append(got, subjectType)
	}

	want := make([]string, 0, len(valueobjects.AllSubjectTypes()))
	for _, st := range valueobjects.AllSubjectTypes() {
		want = append(want, st.Value())
	}

	assert.ElementsMatch(t, want, got)
}
