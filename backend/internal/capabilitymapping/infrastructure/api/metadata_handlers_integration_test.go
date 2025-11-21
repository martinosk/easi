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
		"SELECT aggregate_id FROM events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&capabilityID)
	require.NoError(t, err)
	testCtx.trackID(capabilityID)

	time.Sleep(100 * time.Millisecond)

	metadataReqBody := UpdateCapabilityMetadataRequest{
		StrategyPillar: "Transform",
		PillarWeight:   75,
		MaturityLevel:  "Developing",
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
		"SELECT event_data FROM events WHERE aggregate_id = $1 AND event_type = 'CapabilityMetadataUpdated'",
		capabilityID,
	).Scan(&metadataEventData)
	require.NoError(t, err)
	assert.Contains(t, metadataEventData, "Transform")
	assert.Contains(t, metadataEventData, "Developing")
	assert.Contains(t, metadataEventData, "TribeOwned")

	time.Sleep(100 * time.Millisecond)

	var strategyPillar, maturityLevel, ownershipModel, status string
	var pillarWeight int
	err = testCtx.db.QueryRow(
		"SELECT strategy_pillar, pillar_weight, maturity_level, ownership_model, status FROM capabilities WHERE id = $1",
		capabilityID,
	).Scan(&strategyPillar, &pillarWeight, &maturityLevel, &ownershipModel, &status)
	require.NoError(t, err)
	assert.Equal(t, "Transform", strategyPillar)
	assert.Equal(t, 75, pillarWeight)
	assert.Equal(t, "Developing", maturityLevel)
	assert.Equal(t, "TribeOwned", ownershipModel)
	assert.Equal(t, "Active", status)
}

func TestUpdateCapabilityMetadata_InvalidPillarWeight_Integration(t *testing.T) {
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
		"SELECT aggregate_id FROM events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
	).Scan(&capabilityID)
	require.NoError(t, err)
	testCtx.trackID(capabilityID)

	time.Sleep(100 * time.Millisecond)

	metadataReqBody := UpdateCapabilityMetadataRequest{
		PillarWeight:  150,
		MaturityLevel: "Initial",
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
		"SELECT aggregate_id FROM events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
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
		"SELECT aggregate_id FROM events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
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

	assert.Equal(t, http.StatusCreated, expertW.Code)

	testCtx.setTenantContext(t)
	var expertEventData string
	err = testCtx.db.QueryRow(
		"SELECT event_data FROM events WHERE aggregate_id = $1 AND event_type = 'CapabilityExpertAdded'",
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
		Level: "L2",
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
		"SELECT aggregate_id FROM events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
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

	assert.Equal(t, http.StatusCreated, tagW.Code)

	testCtx.setTenantContext(t)
	var tagEventData string
	err = testCtx.db.QueryRow(
		"SELECT event_data FROM events WHERE aggregate_id = $1 AND event_type = 'CapabilityTagAdded'",
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
		"SELECT aggregate_id FROM events WHERE event_type = 'CapabilityCreated' ORDER BY created_at DESC LIMIT 1",
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
		MaturityLevel: "Initial",
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
