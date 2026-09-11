//go:build integration
// +build integration

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	sharedAPI "easi/backend/internal/shared/api"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type singleTenantDirectory struct{ tenantID string }

func (d *singleTenantDirectory) TenantIDs(_ context.Context) ([]string, error) {
	return []string{d.tenantID}, nil
}

type schemaAttributeDTO struct {
	ID     string                    `json:"id"`
	Name   string                    `json:"name"`
	Type   string                    `json:"type"`
	Active bool                      `json:"active"`
	Links  map[string]sharedAPI.Link `json:"_links"`
}

type schemaDTO struct {
	ID         string                    `json:"id"`
	Version    int                       `json:"version"`
	Attributes []schemaAttributeDTO      `json:"attributes"`
	Links      map[string]sharedAPI.Link `json:"_links"`
}

func (ic *integrationContext) getSchema(t *testing.T) schemaDTO {
	t.Helper()
	rec := ic.do(t, http.MethodGet, "/meta-model/subject-types/application/attributes", nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var dto schemaDTO
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &dto))
	ic.createdIDs = append(ic.createdIDs, dto.ID)
	return dto
}

func findCustomField(dto OnePagerConfigurationDTO, id string) (CustomFieldDTO, bool) {
	for _, field := range dto.CustomFields {
		if field.ID == id {
			return field, true
		}
	}
	return CustomFieldDTO{}, false
}

func TestDefineAttributeInMetaModel_AppearsOnOnePagerConfiguration_Integration(t *testing.T) {
	ic := setupIntegrationWith(t, integrationOptions{metaModel: true})
	config := ic.getConfiguration(t)
	schema := ic.getSchema(t)
	assert.Equal(t, "/api/v1/meta-model/subject-types/application/attributes", config.Links["x-attribute-schema"].Href)

	rec := ic.do(t, http.MethodPost, "/meta-model/subject-types/application/attributes", map[string]any{
		"name": "Hosting Region", "type": "selection", "options": []string{"EU", "US"}, "version": schema.Version,
	})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var defined schemaDTO
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &defined))
	require.Len(t, defined.Attributes, 1)
	attributeID := defined.Attributes[0].ID

	updated := ic.getConfiguration(t)
	field, found := findCustomField(updated, attributeID)
	require.True(t, found, "the attribute defined in MetaModel is available on the one-pager configuration")
	assert.Equal(t, "Hosting Region", field.Name)
	assert.True(t, field.Included)
	assert.Contains(t, field.Links, "x-set-requirement")
	assert.Equal(t, config.Version+1, updated.Version, "inclusion is recorded on the one-pager configuration")

	requireRec := ic.do(t, http.MethodPut, fmt.Sprintf("/one-pagers/configurations/application/custom-fields/%s/requirement", attributeID), map[string]any{
		"required": true, "version": updated.Version,
	})
	require.Equal(t, http.StatusOK, requireRec.Code, requireRec.Body.String())
	assert.Equal(t, defined.Version, ic.getSchema(t).Version, "required-ness leaves MetaModel unchanged")

	retireRec := ic.do(t, http.MethodPost, fmt.Sprintf("/meta-model/subject-types/application/attributes/%s/retire", attributeID), map[string]any{
		"version": defined.Version,
	})
	require.Equal(t, http.StatusOK, retireRec.Code, retireRec.Body.String())

	afterRetire := ic.getConfiguration(t)
	retired, found := findCustomField(afterRetire, attributeID)
	require.True(t, found)
	assert.False(t, retired.Active)
	assert.False(t, retired.Included, "retiring in MetaModel stops the one-pager offering the field")
	assert.NotContains(t, retired.Links, "x-set-requirement")
}

func TestLegacyDefinitionsTransferToMetaModelWithPreservedIdentity_Integration(t *testing.T) {
	fieldID := uuid.New().String()
	optionID := uuid.New().String()
	tenants := &singleTenantDirectory{}
	ic := setupIntegrationWith(t, integrationOptions{
		metaModel: true,
		tenants:   tenants,
		beforeRoutes: func(ic *integrationContext) {
			tenants.tenantID = ic.tenantID
			require.NoError(t, ic.execAsTenant(t,
				`INSERT INTO onepagers.custom_field_definition_cache (tenant_id, field_id, subject_type, definition, pending_transfer)
				VALUES ($1, $2, 'application', $3::jsonb, TRUE)`,
				ic.tenantID, fieldID,
				fmt.Sprintf(`{"id":%q,"name":"Legacy hosting","type":"selection","helpText":"h","active":true,"options":[{"id":%q,"label":"Cloud","active":true}]}`, fieldID, optionID),
			))
		},
	})

	schema := ic.getSchema(t)
	require.Len(t, schema.Attributes, 1)
	assert.Equal(t, fieldID, schema.Attributes[0].ID, "FieldID preserved through the transfer")
	assert.Equal(t, "Legacy hosting", schema.Attributes[0].Name)

	pending := ic.countAsTenant(t, "SELECT COUNT(*) FROM onepagers.custom_field_definition_cache WHERE tenant_id = $1 AND pending_transfer", ic.tenantID)
	assert.Equal(t, 0, pending, "the cache row is confirmed by MetaModel's published event")
}
