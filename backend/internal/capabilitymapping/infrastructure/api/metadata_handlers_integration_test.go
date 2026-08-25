//go:build integration
// +build integration

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/infrastructure/database"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withChiID(req *http.Request, id string) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{id},
		},
	}))
}

type jsonRequest struct {
	method string
	target string
	id     string
	body   interface{}
}

func newJSONRequest(t *testing.T, spec jsonRequest) *http.Request {
	encoded, err := json.Marshal(spec.body)
	require.NoError(t, err)

	req := httptest.NewRequest(spec.method, spec.target, bytes.NewReader(encoded))
	req.Header.Set("Content-Type", "application/json")
	req = withTestTenant(req)
	if spec.id != "" {
		req = withChiID(req, spec.id)
	}
	return req
}

func invokeHandler(handler http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func createCapabilityForTest(t *testing.T, testCtx *testContext, handlers *CapabilityHandlers, body CreateCapabilityRequest) string {
	createReq := newJSONRequest(t, jsonRequest{
		method: http.MethodPost,
		target: "/api/v1/capabilities",
		body:   body,
	})
	createW := invokeHandler(handlers.CreateCapability, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	testCtx.setTenantContext(t)
	var capabilityID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM infrastructure.events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&capabilityID)
	require.NoError(t, err)
	testCtx.trackID(capabilityID)

	time.Sleep(100 * time.Millisecond)
	return capabilityID
}

func (ctx *testContext) requireEventDataContains(t *testing.T, aggregateID, eventType string, substrings ...string) {
	ctx.setTenantContext(t)
	var eventData string
	err := ctx.db.QueryRow(
		"SELECT event_data FROM infrastructure.events WHERE aggregate_id = $1 AND event_type = $2",
		aggregateID, eventType,
	).Scan(&eventData)
	require.NoError(t, err)
	for _, s := range substrings {
		assert.Contains(t, eventData, s)
	}
}

func seedCachedUser(t *testing.T, testCtx *testContext, name, email string) string {
	t.Helper()
	userID := uuid.New().String()
	cache := readmodels.NewUserNameCacheReadModel(database.NewTenantAwareDB(testCtx.db))
	require.NoError(t, cache.Upsert(tenantContext(), userID, name, email))
	testCtx.createdUserIDs = append(testCtx.createdUserIDs, userID)
	return userID
}

func decodeCapability(t *testing.T, w *httptest.ResponseRecorder) readmodels.CapabilityDTO {
	t.Helper()
	var dto readmodels.CapabilityDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &dto))
	return dto
}

func TestUpdateCapabilityMetadata_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)
	janeID := seedCachedUser(t, testCtx, "Jane Smith", "jane.smith@example.com")

	capabilityID := createCapabilityForTest(t, testCtx, handlers, CreateCapabilityRequest{
		Name:        "Digital Transformation",
		Description: "Transform business digitally",
		Level:       "L1",
	})

	metadataReq := newJSONRequest(t, jsonRequest{
		method: http.MethodPut,
		target: "/api/v1/capabilities/" + capabilityID + "/metadata",
		id:     capabilityID,
		body: UpdateCapabilityMetadataRequest{
			MaturityLevel:  "Custom Build",
			OwnershipModel: "TribeOwned",
			PrimaryOwner:   "Platform Tribe - John Doe",
			EAOwner:        "Jane Smith",
			Status:         "Active",
		},
	})
	metadataW := invokeHandler(handlers.UpdateCapabilityMetadata, metadataReq)

	require.Equal(t, http.StatusOK, metadataW.Code)
	dto := decodeCapability(t, metadataW)
	assert.Equal(t, janeID, dto.EAOwner)
	assert.Equal(t, "Jane Smith", dto.EAOwnerName)

	testCtx.requireEventDataContains(t, capabilityID, "CapabilityMetadataUpdated", `"maturityValue": 37`, "TribeOwned", janeID)

	time.Sleep(100 * time.Millisecond)

	testCtx.setTenantContext(t)
	var ownershipModel, status, eaOwner string
	var maturityValue int
	err := testCtx.db.QueryRow(
		"SELECT maturity_value, ownership_model, status, ea_owner FROM capabilitymapping.capabilities WHERE id = $1",
		capabilityID,
	).Scan(&maturityValue, &ownershipModel, &status, &eaOwner)
	require.NoError(t, err)
	assert.Equal(t, 37, maturityValue)
	assert.Equal(t, "TribeOwned", ownershipModel)
	assert.Equal(t, "Active", status)
	assert.Equal(t, janeID, eaOwner)
}

func TestUpdateCapabilityMetadata_EAOwnerUserID_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)
	aliceID := seedCachedUser(t, testCtx, "Alice Smith", "alice.smith@example.com")

	capabilityID := createCapabilityForTest(t, testCtx, handlers, CreateCapabilityRequest{
		Name:  "Payments",
		Level: "L1",
	})

	metadataReq := newJSONRequest(t, jsonRequest{
		method: http.MethodPut,
		target: "/api/v1/capabilities/" + capabilityID + "/metadata",
		id:     capabilityID,
		body: UpdateCapabilityMetadataRequest{
			EAOwner: aliceID,
			Status:  "Active",
		},
	})
	metadataW := invokeHandler(handlers.UpdateCapabilityMetadata, metadataReq)

	require.Equal(t, http.StatusOK, metadataW.Code)
	dto := decodeCapability(t, metadataW)
	assert.Equal(t, aliceID, dto.EAOwner)
	assert.Equal(t, "Alice Smith", dto.EAOwnerName)
}

func TestUpdateCapabilityMetadata_ValidationErrors_Integration(t *testing.T) {
	invalidValue := 150

	cases := []struct {
		name  string
		body  UpdateCapabilityMetadataRequest
		setup func(t *testing.T, testCtx *testContext)
	}{
		{
			name: "maturity value out of range",
			body: UpdateCapabilityMetadataRequest{MaturityValue: &invalidValue, Status: "Active"},
		},
		{
			name: "unknown maturity level",
			body: UpdateCapabilityMetadataRequest{MaturityLevel: "InvalidLevel", Status: "Active"},
		},
		{
			name: "unresolvable EA owner",
			body: UpdateCapabilityMetadataRequest{EAOwner: "Nobody Anyone Knows", Status: "Active"},
		},
		{
			name: "unknown EA owner user id",
			body: UpdateCapabilityMetadataRequest{EAOwner: uuid.New().String(), Status: "Active"},
		},
		{
			name: "ambiguous EA owner name",
			body: UpdateCapabilityMetadataRequest{EAOwner: "Alex Kim", Status: "Active"},
			setup: func(t *testing.T, testCtx *testContext) {
				seedCachedUser(t, testCtx, "Alex Kim", "alex.kim.1@example.com")
				seedCachedUser(t, testCtx, "Alex Kim", "alex.kim.2@example.com")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testCtx, cleanup := setupTestDB(t)
			defer cleanup()

			if tc.setup != nil {
				tc.setup(t, testCtx)
			}
			handlers := setupHandlers(testCtx.db)
			capabilityID := createCapabilityForTest(t, testCtx, handlers, CreateCapabilityRequest{
				Name:  "Test Capability",
				Level: "L1",
			})

			req := newJSONRequest(t, jsonRequest{
				method: http.MethodPut,
				target: "/api/v1/capabilities/" + capabilityID + "/metadata",
				id:     capabilityID,
				body:   tc.body,
			})
			w := invokeHandler(handlers.UpdateCapabilityMetadata, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestAddCapabilityExpert_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	capabilityID := createCapabilityForTest(t, testCtx, handlers, CreateCapabilityRequest{
		Name:  "Data Management",
		Level: "L1",
	})

	expertReq := newJSONRequest(t, jsonRequest{
		method: http.MethodPost,
		target: "/api/v1/capabilities/" + capabilityID + "/experts",
		id:     capabilityID,
		body: AddCapabilityExpertRequest{
			ExpertName:  "Dr. Alice Johnson",
			ExpertRole:  "Data Architect",
			ContactInfo: "alice.johnson@example.com",
		},
	})
	expertW := invokeHandler(handlers.AddCapabilityExpert, expertReq)

	assert.Equal(t, http.StatusNoContent, expertW.Code)

	testCtx.requireEventDataContains(t, capabilityID, "CapabilityExpertAdded",
		"Dr. Alice Johnson", "Data Architect", "alice.johnson@example.com")
}

func TestAddCapabilityTag_Integration(t *testing.T) {
	cases := []struct {
		name             string
		tag              string
		expectedStatus   int
		expectedEventTag string
	}{
		{
			name:             "valid tag is added",
			tag:              "Cloud-native",
			expectedStatus:   http.StatusNoContent,
			expectedEventTag: "Cloud-native",
		},
		{
			name:           "empty tag is rejected",
			tag:            "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testCtx, cleanup := setupTestDB(t)
			defer cleanup()

			handlers := setupHandlers(testCtx.db)
			capabilityID := createCapabilityForTest(t, testCtx, handlers, CreateCapabilityRequest{
				Name:  "Test Capability",
				Level: "L1",
			})

			tagReq := newJSONRequest(t, jsonRequest{
				method: http.MethodPost,
				target: "/api/v1/capabilities/" + capabilityID + "/tags",
				id:     capabilityID,
				body:   AddCapabilityTagRequest{Tag: tc.tag},
			})
			tagW := invokeHandler(handlers.AddCapabilityTag, tagReq)

			assert.Equal(t, tc.expectedStatus, tagW.Code)

			if tc.expectedEventTag != "" {
				testCtx.requireEventDataContains(t, capabilityID, "CapabilityTagAdded", tc.expectedEventTag)
			}
		})
	}
}

func TestUpdateCapabilityMetadata_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())

	metadataReq := newJSONRequest(t, jsonRequest{
		method: http.MethodPut,
		target: "/api/v1/capabilities/" + nonExistentID + "/metadata",
		id:     nonExistentID,
		body: UpdateCapabilityMetadataRequest{
			MaturityLevel: "Genesis",
			Status:        "Active",
		},
	})
	metadataW := invokeHandler(handlers.UpdateCapabilityMetadata, metadataReq)

	assert.Equal(t, http.StatusNotFound, metadataW.Code)
}
