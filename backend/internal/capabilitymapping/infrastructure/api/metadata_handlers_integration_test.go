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

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateCapabilityMetadata_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	createReqBody := CreateCapabilityRequest{
		Name:        "Digital Transformation",
		Description: "Transform business digitally",
		Level:       "L1",
	}
	createBody, _ := json.Marshal(createReqBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = withTestTenant(createReq)
	createW := httptest.NewRecorder()

	handlers.CreateCapability(createW, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	testCtx.setTenantContext(t)
	var capabilityID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM infrastructure.events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&capabilityID)
	require.NoError(t, err)
	testCtx.trackID(capabilityID)

	time.Sleep(100 * time.Millisecond)

	metadataReqBody := UpdateCapabilityMetadataRequest{
		MaturityLevel:  "Custom Build",
		OwnershipModel: "TribeOwned",
		PrimaryOwner:   "Platform Tribe - John Doe",
		EAOwner:        "Jane Smith",
		Status:         "Active",
	}
	metadataBody, _ := json.Marshal(metadataReqBody)

	metadataReq := httptest.NewRequest(http.MethodPut, "/api/v1/capabilities/"+capabilityID+"/metadata", bytes.NewReader(metadataBody))
	metadataReq.Header.Set("Content-Type", "application/json")
	metadataReq = withTestTenant(metadataReq)
	metadataReq = metadataReq.WithContext(context.WithValue(metadataReq.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{capabilityID},
		},
	}))
	metadataW := httptest.NewRecorder()

	handlers.UpdateCapabilityMetadata(metadataW, metadataReq)

	assert.Equal(t, http.StatusOK, metadataW.Code)

	testCtx.setTenantContext(t)
	var metadataEventData string
	err = testCtx.db.QueryRow(
		"SELECT event_data FROM infrastructure.events WHERE aggregate_id = $1 AND event_type = 'CapabilityMetadataUpdated'",
		capabilityID,
	).Scan(&metadataEventData)
	require.NoError(t, err)
	assert.Contains(t, metadataEventData, `"maturityValue": 37`)
	assert.Contains(t, metadataEventData, "TribeOwned")

	time.Sleep(100 * time.Millisecond)

	var ownershipModel, status string
	var maturityValue int
	err = testCtx.db.QueryRow(
		"SELECT maturity_value, ownership_model, status FROM capabilitymapping.capabilities WHERE id = $1",
		capabilityID,
	).Scan(&maturityValue, &ownershipModel, &status)
	require.NoError(t, err)
	assert.Equal(t, 37, maturityValue)
	assert.Equal(t, "TribeOwned", ownershipModel)
	assert.Equal(t, "Active", status)
}

func TestUpdateCapabilityMetadata_InvalidMaturityValue_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	createReqBody := CreateCapabilityRequest{
		Name:  "Test Capability",
		Level: "L1",
	}
	createBody, _ := json.Marshal(createReqBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = withTestTenant(createReq)
	createW := httptest.NewRecorder()

	handlers.CreateCapability(createW, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	testCtx.setTenantContext(t)
	var capabilityID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM infrastructure.events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&capabilityID)
	require.NoError(t, err)
	testCtx.trackID(capabilityID)

	time.Sleep(100 * time.Millisecond)

	invalidValue := 150
	metadataReqBody := UpdateCapabilityMetadataRequest{
		MaturityValue: &invalidValue,
		Status:        "Active",
	}
	metadataBody, _ := json.Marshal(metadataReqBody)

	metadataReq := httptest.NewRequest(http.MethodPut, "/api/v1/capabilities/"+capabilityID+"/metadata", bytes.NewReader(metadataBody))
	metadataReq.Header.Set("Content-Type", "application/json")
	metadataReq = withTestTenant(metadataReq)
	metadataReq = metadataReq.WithContext(context.WithValue(metadataReq.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{capabilityID},
		},
	}))
	metadataW := httptest.NewRecorder()

	handlers.UpdateCapabilityMetadata(metadataW, metadataReq)

	assert.Equal(t, http.StatusBadRequest, metadataW.Code)
}

func TestUpdateCapabilityMetadata_InvalidMaturityLevel_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	createReqBody := CreateCapabilityRequest{
		Name:  "Test Capability",
		Level: "L1",
	}
	createBody, _ := json.Marshal(createReqBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = withTestTenant(createReq)
	createW := httptest.NewRecorder()

	handlers.CreateCapability(createW, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	testCtx.setTenantContext(t)
	var capabilityID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM infrastructure.events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&capabilityID)
	require.NoError(t, err)
	testCtx.trackID(capabilityID)

	time.Sleep(100 * time.Millisecond)

	metadataReqBody := UpdateCapabilityMetadataRequest{
		MaturityLevel: "InvalidLevel",
		Status:        "Active",
	}
	metadataBody, _ := json.Marshal(metadataReqBody)

	metadataReq := httptest.NewRequest(http.MethodPut, "/api/v1/capabilities/"+capabilityID+"/metadata", bytes.NewReader(metadataBody))
	metadataReq.Header.Set("Content-Type", "application/json")
	metadataReq = withTestTenant(metadataReq)
	metadataReq = metadataReq.WithContext(context.WithValue(metadataReq.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{capabilityID},
		},
	}))
	metadataW := httptest.NewRecorder()

	handlers.UpdateCapabilityMetadata(metadataW, metadataReq)

	assert.Equal(t, http.StatusBadRequest, metadataW.Code)
}

func TestAddCapabilityExpert_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	createReqBody := CreateCapabilityRequest{
		Name:  "Data Management",
		Level: "L1",
	}
	createBody, _ := json.Marshal(createReqBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = withTestTenant(createReq)
	createW := httptest.NewRecorder()

	handlers.CreateCapability(createW, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	testCtx.setTenantContext(t)
	var capabilityID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM infrastructure.events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&capabilityID)
	require.NoError(t, err)
	testCtx.trackID(capabilityID)

	time.Sleep(100 * time.Millisecond)

	expertReqBody := AddCapabilityExpertRequest{
		ExpertName:  "Dr. Alice Johnson",
		ExpertRole:  "Data Architect",
		ContactInfo: "alice.johnson@example.com",
	}
	expertBody, _ := json.Marshal(expertReqBody)

	expertReq := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities/"+capabilityID+"/experts", bytes.NewReader(expertBody))
	expertReq.Header.Set("Content-Type", "application/json")
	expertReq = withTestTenant(expertReq)
	expertReq = expertReq.WithContext(context.WithValue(expertReq.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{capabilityID},
		},
	}))
	expertW := httptest.NewRecorder()

	handlers.AddCapabilityExpert(expertW, expertReq)

	assert.Equal(t, http.StatusNoContent, expertW.Code)

	testCtx.setTenantContext(t)
	var expertEventData string
	err = testCtx.db.QueryRow(
		"SELECT event_data FROM infrastructure.events WHERE aggregate_id = $1 AND event_type = 'CapabilityExpertAdded'",
		capabilityID,
	).Scan(&expertEventData)
	require.NoError(t, err)
	assert.Contains(t, expertEventData, "Dr. Alice Johnson")
	assert.Contains(t, expertEventData, "Data Architect")
	assert.Contains(t, expertEventData, "alice.johnson@example.com")
}

func TestAddCapabilityTag_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	createReqBody := CreateCapabilityRequest{
		Name:  "API Management",
		Level: "L1",
	}
	createBody, _ := json.Marshal(createReqBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = withTestTenant(createReq)
	createW := httptest.NewRecorder()

	handlers.CreateCapability(createW, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	testCtx.setTenantContext(t)
	var capabilityID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM infrastructure.events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&capabilityID)
	require.NoError(t, err)
	testCtx.trackID(capabilityID)

	time.Sleep(100 * time.Millisecond)

	tagReqBody := AddCapabilityTagRequest{
		Tag: "Cloud-native",
	}
	tagBody, _ := json.Marshal(tagReqBody)

	tagReq := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities/"+capabilityID+"/tags", bytes.NewReader(tagBody))
	tagReq.Header.Set("Content-Type", "application/json")
	tagReq = withTestTenant(tagReq)
	tagReq = tagReq.WithContext(context.WithValue(tagReq.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{capabilityID},
		},
	}))
	tagW := httptest.NewRecorder()

	handlers.AddCapabilityTag(tagW, tagReq)

	assert.Equal(t, http.StatusNoContent, tagW.Code)

	testCtx.setTenantContext(t)
	var tagEventData string
	err = testCtx.db.QueryRow(
		"SELECT event_data FROM infrastructure.events WHERE aggregate_id = $1 AND event_type = 'CapabilityTagAdded'",
		capabilityID,
	).Scan(&tagEventData)
	require.NoError(t, err)
	assert.Contains(t, tagEventData, "Cloud-native")
}

func TestAddCapabilityTag_EmptyTag_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	createReqBody := CreateCapabilityRequest{
		Name:  "Test Capability",
		Level: "L1",
	}
	createBody, _ := json.Marshal(createReqBody)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq = withTestTenant(createReq)
	createW := httptest.NewRecorder()

	handlers.CreateCapability(createW, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	testCtx.setTenantContext(t)
	var capabilityID string
	err := testCtx.db.QueryRow(
		"SELECT aggregate_id FROM infrastructure.events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&capabilityID)
	require.NoError(t, err)
	testCtx.trackID(capabilityID)

	time.Sleep(100 * time.Millisecond)

	tagReqBody := AddCapabilityTagRequest{
		Tag: "",
	}
	tagBody, _ := json.Marshal(tagReqBody)

	tagReq := httptest.NewRequest(http.MethodPost, "/api/v1/capabilities/"+capabilityID+"/tags", bytes.NewReader(tagBody))
	tagReq.Header.Set("Content-Type", "application/json")
	tagReq = withTestTenant(tagReq)
	tagReq = tagReq.WithContext(context.WithValue(tagReq.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{capabilityID},
		},
	}))
	tagW := httptest.NewRecorder()

	handlers.AddCapabilityTag(tagW, tagReq)

	assert.Equal(t, http.StatusBadRequest, tagW.Code)
}

func TestUpdateCapabilityMetadata_NotFound_Integration(t *testing.T) {
	testCtx, cleanup := setupTestDB(t)
	defer cleanup()

	handlers := setupHandlers(testCtx.db)

	nonExistentID := fmt.Sprintf("non-existent-%d", time.Now().UnixNano())

	metadataReqBody := UpdateCapabilityMetadataRequest{
		MaturityLevel: "Genesis",
		Status:        "Active",
	}
	metadataBody, _ := json.Marshal(metadataReqBody)

	metadataReq := httptest.NewRequest(http.MethodPut, "/api/v1/capabilities/"+nonExistentID+"/metadata", bytes.NewReader(metadataBody))
	metadataReq.Header.Set("Content-Type", "application/json")
	metadataReq = withTestTenant(metadataReq)
	metadataReq = metadataReq.WithContext(context.WithValue(metadataReq.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{nonExistentID},
		},
	}))
	metadataW := httptest.NewRecorder()

	handlers.UpdateCapabilityMetadata(metadataW, metadataReq)

	assert.Equal(t, http.StatusNotFound, metadataW.Code)
}
