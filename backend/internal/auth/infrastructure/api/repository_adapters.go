package api

import (
	"context"

	"easi/backend/internal/auth/infrastructure/repositories"
)

type UserRepositoryAdapter struct {
	repo *repositories.UserRepository
}

func NewUserRepositoryAdapter(repo *repositories.UserRepository) *UserRepositoryAdapter {
	return &UserRepositoryAdapter{repo: repo}
}

func (a *UserRepositoryAdapter) GetByEmail(ctx context.Context, tenantID, email string) (*UserDTO, error) {
	user, err := a.repo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return nil, err
	}
	return &UserDTO{
		ID:     user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Role:   user.Role,
		Status: user.Status,
	}, nil
}

type TenantRepositoryAdapter struct {
	repo *repositories.TenantRepository
}

func NewTenantRepositoryAdapter(repo *repositories.TenantRepository) *TenantRepositoryAdapter {
	return &TenantRepositoryAdapter{repo: repo}
}

func (a *TenantRepositoryAdapter) GetByID(ctx context.Context, tenantID string) (*TenantDTO, error) {
	tenant, err := a.repo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return &TenantDTO{
		ID:   tenant.ID,
		Name: tenant.Name,
	}, nil
}
