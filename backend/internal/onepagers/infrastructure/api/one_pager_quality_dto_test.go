package api

import (
	"testing"
	"time"

	"easi/backend/internal/onepagers/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func actorWithPermissions(perms ...string) sharedctx.Actor {
	permissions := make(map[string]bool, len(perms))
	for _, p := range perms {
		permissions[p] = true
	}
	return sharedctx.Actor{Permissions: permissions}
}

func TestToQualityRow(t *testing.T) {
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	updated := time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC)
	links := NewOnePagerLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	row := toQualityRow(readmodels.SubjectIndexRecord{
		SubjectType: "application", SubjectID: "app-1", Name: "Billing",
		CreatorActorID: "user-1", CreatorEmail: "a@dfds.com",
		CreatedAt: created, LastUpdatedAt: updated, RequiredCount: 3, FilledCount: 1,
	}, sharedctx.Actor{}, links)

	assert.Equal(t, "application", row.SubjectType)
	assert.Equal(t, "app-1", row.SubjectID)
	assert.Equal(t, "Billing", row.Name)
	assert.Equal(t, "incomplete", row.Completeness)
	assert.Equal(t, 3, row.RequiredCount)
	assert.Equal(t, 1, row.FilledCount)
	assert.Equal(t, 2, row.MissingCount)
	assert.Equal(t, "user-1", row.CreatorID)
	assert.Equal(t, "a@dfds.com", row.CreatorEmail)
	assert.Equal(t, created, row.CreatedAt)
	assert.Equal(t, updated, row.LastUpdatedAt)
	assert.Empty(t, row.Links)
}

func TestQualitySubjectGrantPermission(t *testing.T) {
	want := map[string]sharedctx.ResourceName{
		"capability":      "capabilities",
		"application":     "components",
		"acquired-entity": "components",
		"vendor":          "components",
		"internal-team":   "components",
	}
	for subjectType, permission := range want {
		got, ok := qualitySubjectGrantPermission[subjectType]
		assert.True(t, ok, subjectType)
		assert.Equal(t, permission, got, subjectType)
	}

	_, ok := qualitySubjectGrantPermission["enterprise-capability"]
	assert.False(t, ok, "enterprise-capability must not be a supported edit-grant subject type")
}

func TestToQualityRow_EditGrantsLink(t *testing.T) {
	links := NewOnePagerLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	cases := []struct {
		name        string
		subjectType string
		permissions []string
		wantLink    bool
	}{
		{"capability grantor can invite", "capability", []string{"capabilities:write"}, true},
		{"capability non-grantor cannot invite", "capability", []string{"capabilities:read"}, false},
		{"application grantor can invite", "application", []string{"components:write"}, true},
		{"application non-grantor cannot invite", "application", []string{"components:read"}, false},
		{"acquired-entity grantor can invite", "acquired-entity", []string{"components:write"}, true},
		{"vendor grantor can invite", "vendor", []string{"components:write"}, true},
		{"internal-team grantor can invite", "internal-team", []string{"components:write"}, true},
		{"edit-grants:manage grants access regardless of write permission", "capability", []string{"edit-grants:manage"}, true},
		{"enterprise-capability is never invitable, even with edit-grants:manage", "enterprise-capability", []string{"edit-grants:manage"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actor := actorWithPermissions(tc.permissions...)
			row := toQualityRow(readmodels.SubjectIndexRecord{SubjectType: tc.subjectType, SubjectID: "id-1"}, actor, links)

			link, ok := row.Links["x-edit-grants"]
			assert.Equal(t, tc.wantLink, ok, tc.name)
			if tc.wantLink {
				assert.Equal(t, "/api/v1/edit-grants", link.Href)
				assert.Equal(t, "POST", link.Method)
			}
		})
	}
}

func TestQualityCursorRoundtrip(t *testing.T) {
	record := readmodels.SubjectIndexRecord{
		SubjectType: "vendor", SubjectID: "ven-9", Name: "Acme",
		CreatorEmail:  "z@dfds.com",
		CreatedAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		LastUpdatedAt: time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
		RequiredCount: 4, FilledCount: 2,
	}

	token := encodeQualityCursor(record)
	require.NotEmpty(t, token)

	decoded, err := decodeQualityCursor(token)
	require.NoError(t, err)
	require.NotNil(t, decoded)
	assert.Equal(t, record.SubjectType, decoded.SubjectType)
	assert.Equal(t, record.SubjectID, decoded.SubjectID)
	assert.Equal(t, record.Name, decoded.Name)
	assert.Equal(t, record.CreatorEmail, decoded.CreatorEmail)
	assert.True(t, record.CreatedAt.Equal(decoded.CreatedAt))
	assert.True(t, record.LastUpdatedAt.Equal(decoded.LastUpdatedAt))
	assert.Equal(t, record.RequiredCount, decoded.RequiredCount)
	assert.Equal(t, record.FilledCount, decoded.FilledCount)
}

func TestDecodeQualityCursor_Empty(t *testing.T) {
	decoded, err := decodeQualityCursor("")
	require.NoError(t, err)
	assert.Nil(t, decoded)
}

func TestDecodeQualityCursor_Invalid(t *testing.T) {
	_, err := decodeQualityCursor("!!!not-base64!!!")
	require.Error(t, err)
}

func TestParseQualitySort(t *testing.T) {
	cases := map[string]struct {
		want string
		ok   bool
	}{
		"":             {readmodels.SortCompleteness, true},
		"completeness": {readmodels.SortCompleteness, true},
		"name":         {readmodels.SortName, true},
		"creator":      {readmodels.SortCreator, true},
		"created":      {readmodels.SortCreated, true},
		"updated":      {readmodels.SortUpdated, true},
		"bogus":        {"", false},
	}
	for input, expected := range cases {
		got, ok := parseQualitySort(input)
		assert.Equal(t, expected.ok, ok, "input=%q", input)
		if expected.ok {
			assert.Equal(t, expected.want, got, "input=%q", input)
		}
	}
}

func TestParseQualityOrder(t *testing.T) {
	cases := map[string]struct {
		want string
		ok   bool
	}{
		"":     {readmodels.OrderAsc, true},
		"asc":  {readmodels.OrderAsc, true},
		"desc": {readmodels.OrderDesc, true},
		"up":   {"", false},
	}
	for input, expected := range cases {
		got, ok := parseQualityOrder(input)
		assert.Equal(t, expected.ok, ok, "input=%q", input)
		if expected.ok {
			assert.Equal(t, expected.want, got, "input=%q", input)
		}
	}
}

func TestReadableSubjectTypes(t *testing.T) {
	cases := []struct {
		name        string
		permissions map[string]bool
		want        []string
	}{
		{"capabilities only", map[string]bool{"capabilities:read": true}, []string{"capability"}},
		{"enterprise only", map[string]bool{"enterprise-arch:read": true}, []string{"enterprise-capability"}},
		{"components covers four types", map[string]bool{"components:read": true}, []string{"application", "acquired-entity", "vendor", "internal-team"}},
		{"none", map[string]bool{}, nil},
		{
			"all three",
			map[string]bool{"capabilities:read": true, "enterprise-arch:read": true, "components:read": true},
			[]string{"capability", "enterprise-capability", "application", "acquired-entity", "vendor", "internal-team"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actor := sharedctx.Actor{Permissions: tc.permissions}
			assert.Equal(t, tc.want, readableSubjectTypes(actor))
		})
	}
}
