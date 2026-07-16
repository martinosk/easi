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
