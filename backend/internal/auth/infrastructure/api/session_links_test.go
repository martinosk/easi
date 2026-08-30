package api

import (
	"context"
	"testing"

	"easi/backend/internal/auth/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBuildSessionLinks_OnePagerQualityEntryPoint(t *testing.T) {
	noReadRole, _ := valueobjects.RoleFromString("nobody")

	cases := []struct {
		name    string
		role    valueobjects.Role
		present bool
	}{
		{"admin can read subjects", valueobjects.RoleAdmin, true},
		{"stakeholder can read subjects", valueobjects.RoleStakeholder, true},
		{"role without read permissions has no entry point", noReadRole, false},
	}

	handlers := &SessionHandlers{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			links := handlers.buildSessionLinks(context.Background(), uuid.New(), tc.role)
			link, ok := links["x-one-pager-quality"]
			assert.Equal(t, tc.present, ok)
			if tc.present {
				assert.Equal(t, "/api/v1/one-pager-quality", link)
			}
		})
	}
}

func TestBuildSessionLinks_AssistantEntryPoints(t *testing.T) {
	cases := []struct {
		name         string
		role         valueobjects.Role
		assistant    bool
		writeVariant bool
	}{
		{"admin gets assistant and write link", valueobjects.RoleAdmin, true, true},
		{"architect gets assistant and write link", valueobjects.RoleArchitect, true, true},
		{"stakeholder gets assistant without write link", valueobjects.RoleStakeholder, true, false},
	}

	handlers := &SessionHandlers{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			links := handlers.buildSessionLinks(context.Background(), uuid.New(), tc.role)
			link, ok := links["x-assistant"]
			assert.Equal(t, tc.assistant, ok)
			if tc.assistant {
				assert.Equal(t, "/api/v1/assistant/conversations", link)
			}
			writeLink, ok := links["x-assistant-write"]
			assert.Equal(t, tc.writeVariant, ok)
			if tc.writeVariant {
				assert.Equal(t, "/api/v1/assistant/conversations", writeLink)
			}
		})
	}
}

func TestBuildSessionLinks_NoAssistantPermission_NoAssistantLinks(t *testing.T) {
	noPermissions, _ := valueobjects.RoleFromString("nobody")
	handlers := &SessionHandlers{}
	links := handlers.buildSessionLinks(context.Background(), uuid.New(), noPermissions)
	assert.NotContains(t, links, "x-assistant")
	assert.NotContains(t, links, "x-assistant-write")
}
