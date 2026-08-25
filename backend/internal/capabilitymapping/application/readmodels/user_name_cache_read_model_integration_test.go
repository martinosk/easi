//go:build integration

package readmodels

import (
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/infrastructure/database"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type userNameCacheFixture struct {
	t          *testing.T
	cache      *UserNameCacheReadModel
	capability *CapabilityReadModel
	tenantDB   *database.TenantAwareDB
}

func newUserNameCacheFixture(t *testing.T) *userNameCacheFixture {
	db, cleanup := setupTestDB(t)
	t.Cleanup(cleanup)
	tenantDB := database.NewTenantAwareDB(db)
	return &userNameCacheFixture{
		t:          t,
		cache:      NewUserNameCacheReadModel(tenantDB),
		capability: NewCapabilityReadModel(tenantDB),
		tenantDB:   tenantDB,
	}
}

func (f *userNameCacheFixture) seedUser(name, email string) string {
	f.t.Helper()
	userID := uuid.New().String()
	require.NoError(f.t, f.cache.Upsert(tenantContext(), userID, name, email))
	f.t.Cleanup(func() {
		_, _ = f.tenantDB.ExecContext(tenantContext(), "DELETE FROM capabilitymapping.user_names WHERE tenant_id = 'default' AND user_id = $1", userID)
	})
	return userID
}

func (f *userNameCacheFixture) seedCapability(name, eaOwner string) string {
	f.t.Helper()
	id := uuid.New().String()
	ctx := tenantContext()
	require.NoError(f.t, f.capability.Insert(ctx, CapabilityDTO{ID: id, Name: name, Level: "L1", CreatedAt: time.Now()}))
	require.NoError(f.t, f.capability.UpdateMetadata(ctx, id, CapabilityMetadataUpdate{EAOwner: eaOwner, Status: "Active"}))
	f.t.Cleanup(func() { _ = f.capability.Delete(tenantContext(), id) })
	return id
}

func TestUserNameCacheReadModel_ResolveEAOwner_Integration(t *testing.T) {
	f := newUserNameCacheFixture(t)
	aliceID := f.seedUser("Alice Resolve", "alice.resolve@example.com")

	cases := []struct {
		name  string
		input string
	}{
		{name: "by user id", input: aliceID},
		{name: "by name", input: "Alice Resolve"},
		{name: "by name case-insensitively", input: "alice RESOLVE"},
		{name: "by email", input: "alice.resolve@example.com"},
		{name: "by email case-insensitively", input: "Alice.Resolve@Example.com"},
		{name: "with surrounding whitespace", input: "  Alice Resolve  "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resolved, err := f.cache.ResolveEAOwner(tenantContext(), tc.input)
			require.NoError(t, err)
			assert.Equal(t, aliceID, resolved)
		})
	}
}

func TestUserNameCacheReadModel_ResolveEAOwner_RejectsUnknownUserID_Integration(t *testing.T) {
	f := newUserNameCacheFixture(t)

	_, err := f.cache.ResolveEAOwner(tenantContext(), uuid.New().String())

	assert.ErrorIs(t, err, valueobjects.ErrEAOwnerNotUser)
}

func TestUserNameCacheReadModel_ResolveEAOwner_RejectsAmbiguousName_Integration(t *testing.T) {
	f := newUserNameCacheFixture(t)
	f.seedUser("Alex Kim", "alex.kim.1@example.com")
	f.seedUser("Alex Kim", "alex.kim.2@example.com")

	_, err := f.cache.ResolveEAOwner(tenantContext(), "Alex Kim")

	assert.ErrorIs(t, err, valueobjects.ErrEAOwnerAmbiguous)
}

func TestUserNameCacheReadModel_UpsertReplacesExistingEntry_Integration(t *testing.T) {
	f := newUserNameCacheFixture(t)
	userID := f.seedUser("Before Rename", "before@example.com")

	require.NoError(t, f.cache.Upsert(tenantContext(), userID, "After Rename", "after@example.com"))

	resolved, err := f.cache.ResolveEAOwner(tenantContext(), "After Rename")
	require.NoError(t, err)
	assert.Equal(t, userID, resolved)
	_, err = f.cache.ResolveEAOwner(tenantContext(), "Before Rename")
	assert.ErrorIs(t, err, valueobjects.ErrEAOwnerNotUser)
}

func TestCapabilityReadModel_EAOwnerName_Integration(t *testing.T) {
	f := newUserNameCacheFixture(t)
	aliceID := f.seedUser("Alice Display", "alice.display@example.com")
	resolvedCapID := f.seedCapability("Resolved Owner Capability", aliceID)
	legacyCapID := f.seedCapability("Legacy Owner Capability", "Legacy Free Text")
	unownedCapID := f.seedCapability("Unowned Capability", "")

	t.Run("GetByID resolves cached user id to name", func(t *testing.T) {
		dto, err := f.capability.GetByID(tenantContext(), resolvedCapID)
		require.NoError(t, err)
		assert.Equal(t, aliceID, dto.EAOwner)
		assert.Equal(t, "Alice Display", dto.EAOwnerName)
	})

	t.Run("GetByID falls back to the stored value for legacy text", func(t *testing.T) {
		dto, err := f.capability.GetByID(tenantContext(), legacyCapID)
		require.NoError(t, err)
		assert.Equal(t, "Legacy Free Text", dto.EAOwner)
		assert.Equal(t, "Legacy Free Text", dto.EAOwnerName)
	})

	t.Run("GetByID leaves the name empty when no owner is set", func(t *testing.T) {
		dto, err := f.capability.GetByID(tenantContext(), unownedCapID)
		require.NoError(t, err)
		assert.Empty(t, dto.EAOwnerName)
	})

	t.Run("GetAll resolves names for every listed capability", func(t *testing.T) {
		all, err := f.capability.GetAll(tenantContext())
		require.NoError(t, err)
		names := map[string]string{}
		for _, dto := range all {
			names[dto.ID] = dto.EAOwnerName
		}
		assert.Equal(t, "Alice Display", names[resolvedCapID])
		assert.Equal(t, "Legacy Free Text", names[legacyCapID])
	})
}
